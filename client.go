package ninep

import (
	"context"
	"errors"
	"io"
	"math"
	"net"
	"path"
	"sync"

	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

// Client represents a 9P2000.L client. There may be multiple outstanding
// calls associated with a single client, and a client may be used by
// multiple goroutines simultaneously.
type Client struct {
	writer sync.Mutex // exclusive stream encoder
	enc    *proto.Encoder
	header proto.Header

	dec *proto.Decoder
	c   io.Closer

	tag *pool
	fid *pool

	maxMessageSize uint32
	maxDataSize    uint32

	mu       sync.Mutex // protects following
	pending  map[uint16]*fcall
	closing  bool
	shutdown bool
}

// Dial connects to an 9P2000.L server at the specified network address.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
func Dial(ctx context.Context, network, addr string, opts ...Option) (*Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, opts...)
}

// NewClient returns a new client to handle requests to the set of
// services at the other end of the connection.
func NewClient(rwc io.ReadWriteCloser, opts ...Option) (*Client, error) {
	c := &Client{
		maxMessageSize: proto.DefaultMaxMessageSize,
		maxDataSize:    proto.DefaultMaxDataSize,
		c:              rwc,

		pending: make(map[uint16]*fcall),
		tag:     newPool(math.MaxUint16),
		fid:     newPool(math.MaxUint32),
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	c.enc = proto.NewEncoder(rwc, c.maxMessageSize)
	c.dec = proto.NewDecoder(rwc, c.maxMessageSize)

	go c.recv()

	if err := c.handshake(); err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}

type fcall struct {
	typ proto.MessageType
	tx  proto.Message

	rx  proto.Message
	err error

	ch chan<- *fcall
}

func (f *fcall) Done(err error) {
	if f == nil {
		return
	}

	if f.err == nil && err != nil {
		f.err = err
	}

	select {
	case f.ch <- f:
	default: // do not block here
	}
}

func (f *fcall) Err() error { return f.err }

var (
	errConnectionShutdown = errors.New("connection is shut down")
	errTagOverflow        = errors.New("out of tags")
	errFidOverflow        = errors.New("out of fids")
)

// Close calls the underlying connection's Close method. If the
// connection is already shutting down, an error is returned.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return errConnectionShutdown
	}
	c.closing = true
	c.mu.Unlock()
	return c.c.Close()
}

func (c *Client) send(f *fcall) {
	c.mu.Lock()
	if c.shutdown || c.closing {
		c.mu.Unlock()
		f.Done(errConnectionShutdown)
		return
	}
	v, ok := c.tag.Get()
	if !ok {
		c.mu.Unlock()
		f.Done(errTagOverflow)
	}
	tag := uint16(v)
	c.pending[tag] = f
	c.mu.Unlock()

	c.writer.Lock()
	c.header.Type, c.header.Tag = f.typ, tag
	log.Debugf("<- %s %s", c.header, f.tx)
	if err := c.enc.Encode(&c.header, f.tx); err != nil {
		c.mu.Lock()
		delete(c.pending, tag)
		c.tag.Put(uint32(tag))
		c.mu.Unlock()
		f.Done(err)
	}
	c.writer.Unlock()
}

func (c *Client) rpc(typ proto.MessageType, tx, rx proto.Message) error {
	ch := make(chan *fcall, 1)
	f := &fcall{typ: typ, tx: tx, rx: rx, ch: ch}
	go c.send(f)
	f = <-ch
	return f.err
}

func (c *Client) recv() (err error) {
	header, rlerror := proto.Header{}, proto.Rlerror{}

	for err == nil {
		if decErr := c.dec.DecodeHeader(&header); decErr != nil {
			err = decErr
			break
		}

		c.mu.Lock()
		f := c.pending[header.Tag]
		delete(c.pending, header.Tag)
		c.tag.Put(uint32(header.Tag))
		c.mu.Unlock()

		switch {
		case f == nil:
			// We've got no pending fcall. That usually means that
			// enc.Encode partially failed, and fcall was already
			// removed.
			err = c.dec.Decode(nil)
		case header.Type == proto.MessageRlerror:
			rlerror = proto.Rlerror{}
			if err = c.dec.Decode(&rlerror); err != nil {
				f.Done(err)
			} else {
				f.Done(unix.Errno(rlerror.Errno))
			}
		default:
			err = c.dec.Decode(f.rx)
			f.Done(err)
		}
		log.Debugf("-> %s %s", header, f.rx)
	}

	c.writer.Lock()
	c.mu.Lock()
	c.shutdown = true
	if err == io.EOF {
		if c.closing {
			err = errConnectionShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for tag, f := range c.pending {
		delete(c.pending, tag)
		c.tag.Put(uint32(tag))
		f.Done(err)
	}
	c.mu.Unlock()
	c.writer.Unlock()
	return err
}

func (c *Client) handshake() error {
	tx := &proto.Tversion{MessageSize: c.maxMessageSize, Version: proto.Version}
	rx := &proto.Rversion{}
	if err := c.rpc(proto.MessageTversion, tx, rx); err != nil {
		return err
	}
	if rx.Version != tx.Version {
		return errors.New("server does not support " + proto.Version)
	}
	if rx.MessageSize != c.maxMessageSize {
		c.maxMessageSize = rx.MessageSize
		c.maxDataSize = rx.MessageSize - (proto.FixedReadWriteSize + 1)
	}
	return nil
}

// Auth initiates an authentication handshake for the given user and root
// path. An error is returned if authentication is not required. If
// successful, the returned afid is used to read/write the authentication
// handshake (protocol does not specify what is read/written), and afid
// is presented in the attach.
func (c *Client) Auth(export, username string, uid int) (*Fid, error) {
	if uid < 0 || uid > math.MaxUint32 {
		return nil, unix.EINVAL
	}

	return nil, errors.New("auth not implemented")
}

// Attach introduces a new user to the server, and establishes Fid as the
// root for that user on the file tree selected by export.
func (c *Client) Attach(auth *Fid, export, username string, uid int) (*Fid, error) {
	if export = path.Clean(export); !path.IsAbs(export) {
		return nil, unix.EINVAL
	}
	if isReserved(export) || uid < 0 || uid > math.MaxUint32 {
		return nil, unix.EINVAL
	}

	fidnum, ok := c.fid.Get()
	if !ok {
		return nil, errFidOverflow
	}

	authnum := uint32(proto.NoFid)
	if auth != nil {
		authnum = auth.Num()
	}

	tx := &proto.Tlattach{
		AuthFid:  authnum,
		Fid:      fidnum,
		Path:     export,
		UserName: username,
		Uid:      uint32(uid),
	}
	rx := &proto.Rlattach{}
	if err := c.rpc(proto.MessageTlattach, tx, rx); err != nil {
		return nil, err
	}

	f := &Fid{c: c, path: export, num: fidnum}
	stat, err := f.stat(proto.GetAttrBasic)
	if err != nil {
		return nil, err
	}
	f.uid, f.gid = stat.Uid, stat.Gid

	return f, nil
}

type pool struct {
	mu  sync.Mutex
	m   []uint32
	cur uint32
	max uint32
}

func newPool(max uint32) *pool {
	return &pool{cur: 1, max: max}
}

func (p *pool) Get() (uint32, bool) {
	p.mu.Lock()
	if len(p.m) > 0 {
		v := p.m[len(p.m)-1]
		p.m = p.m[:len(p.m)-1]
		p.mu.Unlock()
		return v, true
	}
	if p.cur == p.max {
		p.mu.Unlock()
		return 0, false
	}
	v := p.cur
	p.cur++
	p.mu.Unlock()
	return v, true
}

func (p *pool) Put(v uint32) {
	p.mu.Lock()
	p.m = append(p.m, v)
	p.mu.Unlock()
}
