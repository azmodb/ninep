package ninep

import (
	"context"
	"io"
	"math"
	"net"
	"path"
	"sync"

	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/log"
	"github.com/azmodb/pkg/pool"
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

	tag *pool.Generator
	fid *pool.Generator

	maxMessageSize uint32
	maxDataSize    uint32

	mu       sync.Mutex // protects following
	pending  map[uint16]*Fcall
	closing  bool
	shutdown bool
}

// Option sets Server or Client options such as logging, max message
// size etc.
type Option func(interface{}) error

// Dial connects to an 9P2000.L server at the specified network address.
//
// The provided Context must be non-nil. If the context expires before
// the connection is complete, an error is returned. Once successfully
// connected, any expiration of the context will not affect the
// connection.
func Dial(ctx context.Context, network, address string, opts ...Option) (*Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
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

		pending: make(map[uint16]*Fcall),
		tag:     pool.NewGenerator(1, math.MaxUint16),
		fid:     pool.NewGenerator(1, math.MaxUint32),
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

// Fcall represents an active 9P RPC.
type Fcall struct {
	Type proto.MessageType // The type of the service and method to call.
	Tx   proto.Message     // The argument to the function.
	Rx   proto.Message     // The reply from the function.

	err error
	ch  chan<- *Fcall
}

func (f *Fcall) done(err error) {
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

// Err returns the first error that was encountered by the Fcall.
func (f *Fcall) Err() error { return f.err }

const (
	errConnectionShutdown = proto.Error("connection is shut down")
	errTagOverflow        = proto.Error("out of tags")
	errFidOverflow        = proto.Error("out of fids")

	errVersionNotSupported = proto.Error("protocol not supported")
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

func (c *Client) send(f *Fcall) {
	c.mu.Lock()
	if c.shutdown || c.closing {
		c.mu.Unlock()
		f.done(errConnectionShutdown)
		return
	}
	v, ok := c.tag.Get()
	if !ok {
		c.mu.Unlock()
		f.done(errTagOverflow)
	}
	tag := uint16(v)
	c.pending[tag] = f
	c.mu.Unlock()

	log.Debugf("<- %s tag:%d %s", f.Type, tag, f.Tx)
	c.writer.Lock()
	c.header.Type, c.header.Tag = f.Type, tag
	if err := c.enc.Encode(&c.header, f.Tx); err != nil {
		c.mu.Lock()
		delete(c.pending, tag)
		c.tag.Put(int64(tag))
		c.mu.Unlock()
		f.done(err)
	}
	c.writer.Unlock()
}

func (c *Client) rpc(t proto.MessageType, tx, rx proto.Message) error {
	ch := make(chan *Fcall, 1)
	f := &Fcall{Type: t, Tx: tx, Rx: rx, ch: ch}
	go c.send(f)
	f = <-ch
	return f.err
}

func (c *Client) Do(ctx context.Context, fcall *Fcall, ch chan<- *Fcall) {
	fcall.ch = ch
	c.send(fcall)
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
		c.tag.Put(int64(header.Tag))
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
				f.done(err)
			} else {
				f.done(unix.Errno(rlerror.Errno))
			}
			log.Debugf("-> %s %s", header, rlerror)
		default:
			err = c.dec.Decode(f.Rx)
			f.done(err)
			log.Debugf("-> %s %s", header, f.Rx)
		}
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
		c.tag.Put(int64(tag))
		f.done(err)
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
		return errVersionNotSupported
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
		return nil, errInvalidUid
	}

	return nil, unix.ENOTSUP
}

// Attach introduces a new user to the server, and establishes Fid as the
// root for that user on the file tree selected by export.
func (c *Client) Attach(auth *Fid, export, username string, uid int) (*Fid, error) {
	if export = path.Clean(export); !path.IsAbs(export) || isReserved(export) {
		return nil, errInvalildName
	}
	if uid < 0 || uid > math.MaxUint32 {
		return nil, errInvalidUid
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
		Fid:      uint32(fidnum),
		Path:     export,
		UserName: username,
		Uid:      uint32(uid),
	}
	rx := &proto.Rlattach{}
	if err := c.rpc(proto.MessageTlattach, tx, rx); err != nil {
		return nil, err
	}

	attr, err := stat(c, tx.Fid, proto.GetAttrBasic)
	if err != nil {
		return nil, err
	}

	fid := &Fid{c: c, num: tx.Fid, fi: &fileInfo{path: export}}
	fid.fi.Rgetattr = attr
	fid.fi.iounit = 0
	return fid, nil
}
