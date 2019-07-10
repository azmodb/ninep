package ninep

import (
	"context"
	"fmt"
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

	dec *proto.Decoder
	c   io.Closer

	tag *pool.Generator
	fid *pool.Generator

	maxMessageSize uint32
	maxDataSize    uint32

	mu       sync.Mutex // protects following
	pending  map[uint16]*proto.Fcall
	closing  bool
	shutdown bool
}

// Option sets Server or Client options such as logging, max message
// size etc.
type Option func(interface{}) error

func WithMaxMessageSize(size uint32) Option {
	return func(v interface{}) error {
		switch v.(type) {
		case *Client:
			v.(*Client).maxMessageSize = size
			v.(*Client).maxDataSize = size - (proto.FixedReadWriteSize + 1)
			return nil
		case *Server:
			v.(*Server).maxMessageSize = size
			v.(*Server).maxDataSize = size - (proto.FixedReadWriteSize + 1)
			return nil
		}
		return fmt.Errorf("unknown ninep option type: %T", v)
	}
}

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
	c, err := newClient(rwc, opts...)
	if err != nil {
		return nil, err
	}

	if err := c.handshake(proto.Version); err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}

func newClient(rwc io.ReadWriteCloser, opts ...Option) (*Client, error) {
	c := &Client{
		maxMessageSize: proto.DefaultMaxMessageSize,
		maxDataSize:    proto.DefaultMaxDataSize,
		c:              rwc,

		pending: make(map[uint16]*proto.Fcall),
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
	return c, nil
}

func done(f *proto.Fcall, err error) {
	if f == nil || f.C == nil {
		return
	}

	if f.Err == nil && err != nil {
		f.Err = err
	}

	select {
	case f.C <- f:
	default: // do not block here
	}
}

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

func (c *Client) send(f *proto.Fcall) {
	c.mu.Lock()
	if c.shutdown || c.closing {
		c.mu.Unlock()
		done(f, errConnectionShutdown)
		return
	}
	v, ok := c.tag.Get()
	if !ok {
		c.mu.Unlock()
		done(f, errTagOverflow)
	}
	tag := uint16(v)
	c.pending[tag] = f
	c.mu.Unlock()

	log.Debugf("<- %s tag:%d %s", f.Tx.MessageType(), tag, f.Tx)
	c.writer.Lock()
	if err := c.enc.Encode(tag, f.Tx); err != nil {
		c.mu.Lock()
		delete(c.pending, tag)
		c.tag.Put(int64(tag))
		c.mu.Unlock()
		done(f, err)
	}
	c.writer.Unlock()
}

func (c *Client) rpc(f *proto.Fcall) error {
	f.C = make(chan *proto.Fcall, 1)
	go c.send(f)
	f = <-f.C
	return f.Err
}

func (c *Client) recv() (err error) {
	rlerror := proto.Rlerror{}

	for err == nil {
		mtype, tag, decErr := c.dec.DecodeHeader()
		if decErr != nil {
			err = decErr
			break
		}

		c.mu.Lock()
		f := c.pending[tag]
		delete(c.pending, tag)
		c.tag.Put(int64(tag))
		c.mu.Unlock()

		switch {
		case f == nil:
			// We've got no pending fcall. That usually means that
			// enc.Encode partially failed, and fcall was already
			// removed.
			err = c.dec.Decode(nil)
		case mtype == proto.MessageRlerror:
			rlerror = proto.Rlerror{}
			if err = c.dec.Decode(&rlerror); err != nil {
				done(f, err)
			} else {
				done(f, unix.Errno(rlerror.Errno))
			}
			log.Debugf("-> %s tag:%d %s", mtype, tag, rlerror)
		default:
			err = c.dec.Decode(f.Rx)
			done(f, err)
			if f.Rx == nil {
				log.Debugf("-> %s tag:%d", mtype, tag)
			} else {
				log.Debugf("-> %s tag:%d %s", mtype, tag, f.Rx)
			}
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
		done(f, err)
	}
	c.mu.Unlock()
	c.writer.Unlock()
	return err
}

func mustAlloc(mtype proto.MessageType) *proto.Fcall {
	f, ok := proto.Alloc(mtype)
	if !ok {
		log.Panicf("cannot alloc %q transaction", mtype)
	}
	return f
}

func (c *Client) handshake(version string) error {
	f := mustAlloc(proto.MessageTversion)
	defer proto.Release(f)

	tx := f.Tx.(*proto.Tversion)
	tx.MessageSize = c.maxMessageSize
	tx.Version = version
	if err := c.rpc(f); err != nil {
		return err
	}

	rx := f.Rx.(*proto.Rversion)
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

	f := mustAlloc(proto.MessageTattach)
	defer proto.Release(f)

	tx := f.Tx.(*proto.Tlattach)
	tx.AuthFid = authnum
	tx.Fid = uint32(fidnum)
	tx.Path = export
	tx.UserName = username
	tx.Uid = uint32(uid)
	if err := c.rpc(f); err != nil {
		return nil, err
	}

	attr, err := c.stat(tx.Fid, proto.GetAttrBasic)
	if err != nil {
		return nil, err
	}

	fid := &Fid{c: c, num: tx.Fid, fi: &fileInfo{path: export}}
	fid.fi.Rgetattr = attr
	fid.fi.iounit = 0
	return fid, nil
}

func (c *Client) stat(num uint32, mask uint64) (*proto.Rgetattr, error) {
	f := mustAlloc(proto.MessageTgetattr)
	defer proto.Release(f)

	tx := f.Tx.(*proto.Tgetattr)
	tx.Fid = num
	tx.RequestMask = mask
	if err := c.rpc(f); err != nil {
		return nil, err
	}

	attr := *f.Rx.(*proto.Rgetattr)
	return &attr, nil
}
