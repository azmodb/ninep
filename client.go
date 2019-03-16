package ninep

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/log"
)

// pool caches allocated but unused values for later reuse. It is used for
// both tags and fids.
//
// A pool is safe for use by multiple goroutines simultaneously.
type pool struct {
	mu    sync.Mutex
	cache []uint32
	cur   uint32
	limit uint32
}

func newPool(start, limit uint32) *pool {
	return &pool{cur: start, limit: limit}
}

// Next selects an arbitrary value from the pool, removes it from the
// pool, and returns it to the caller.
func (p *pool) Next() (uint32, bool) {
	p.mu.Lock()
	if len(p.cache) > 0 {
		v := p.cache[len(p.cache)-1]
		p.cache = p.cache[:len(p.cache)-1]
		p.mu.Unlock()
		return v, true
	}

	if p.cur == p.limit {
		p.mu.Unlock()
		return 0, false
	}

	v := p.cur
	p.cur++
	p.mu.Unlock()
	return v, true
}

// Put returns a value to the pool.
func (p *pool) Put(v uint32) {
	p.mu.Lock()
	p.cache = append(p.cache, v)
	p.mu.Unlock()
}

type fcall struct {
	buf *proto.Buffer
	fid *Fid

	err error
	ch  chan *fcall
}

func newFcall() *fcall {
	return &fcall{ch: make(chan *fcall, 1)}
}

func (f *fcall) reset() { *f = fcall{} }

func (f *fcall) done(buf *proto.Buffer, err error) {
	if f == nil {
		return
	}

	if f.err == nil && err != nil {
		f.err = err
	}
	f.buf = buf

	select {
	case f.ch <- f:
	default: // do not block here
	}
}

func (f *fcall) wait(ctx context.Context) error {
	select {
	case f = <-f.ch:
		return f.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Client represents a 9P2000 client. There may be multiple outstanding
// calls associated with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	// The network connection itself. We expose it in the struct so that
	// it is available for transport-based auth and any timeouts we need
	// to implement.
	rwc io.ReadWriteCloser

	dec *proto.Decoder
	enc *encoder

	fidpool *pool
	tagpool *pool

	mu       sync.Mutex // protects following
	pending  map[uint16]*fcall
	shutdown bool
	closing  bool
}

type ClientOption func(*Client) error

// Dial connects to an 9P2000 server at the specified network address. For
// TCP and UDP networks, the address has the form "host:port". For unix
// networks, the address must be a file system path.
func Dial(ctx context.Context, network, addr string, opts ...ClientOption) (*Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return NewClient(conn, opts...)
}

// NewClient returns a new Client to handle requests to the set of
// services at the other end of the connection.
func NewClient(rwc io.ReadWriteCloser, opts ...ClientOption) (*Client, error) {
	c := &Client{
		enc: &encoder{e: proto.NewEncoder(rwc)},
		dec: proto.NewDecoder(rwc),
		rwc: rwc,

		pending: make(map[uint16]*fcall),
		tagpool: newPool(1, proto.NoTag),
		fidpool: newPool(1, proto.NoFid),
	}
	go c.recv()

	if err := c.version(proto.DefaultMaxMessageSize, proto.Version); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

var (
	errClientShutdown = errors.New("client is shut down")
	errOutOfTags      = errors.New("out of tags")
	errOutOfFids      = errors.New("out of fids")
)

// Close calls the underlying connection Close method. If the connection
// is already shutting down, an error is returned.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return errClientShutdown
	}
	c.closing = true
	c.mu.Unlock()
	return c.rwc.Close()
}

func (c *Client) register(needfid bool) (uint16, *fcall, error) {
	vtag, ok := c.tagpool.Next()
	if !ok {
		return 0, nil, errOutOfTags
	}
	tag := uint16(vtag)

	fcall := newFcall()
	if needfid {
		num, ok := c.fidpool.Next()
		if !ok {
			c.tagpool.Put(vtag)
			return 0, nil, errOutOfFids
		}
		fcall.fid = newFid(c, num)
	}

	c.mu.Lock()
	if c.shutdown || c.closing {
		return 0, nil, errClientShutdown
		c.mu.Unlock()
	}
	c.pending[tag] = fcall
	c.mu.Unlock()

	return tag, fcall, nil
}

func (c *Client) deregister(tag uint16) (*fcall, bool) {
	c.mu.Lock()
	fcall, found := c.pending[tag]
	delete(c.pending, tag)
	c.mu.Unlock()

	if tag != 0 && tag != proto.NoTag {
		c.tagpool.Put(uint32(tag))
	}

	return fcall, found
}

// version request negotiates the protocol version and message size to
// be used on the connection and initializes the connection for I/O.
func (c *Client) version(msize int64, version string) (err error) {
	fcall := newFcall()
	defer fcall.reset()

	c.mu.Lock()
	c.pending[proto.NoTag] = fcall
	c.mu.Unlock()

	log.Debugf("<- Tversion tag:%d msize:%d version:%q", proto.NoTag, msize, version)
	if err = c.enc.Tversion(proto.NoTag, msize, version); err != nil {
		c.deregister(proto.NoTag)
		return err
	}
	if err = fcall.wait(context.TODO()); err != nil {
		return err
	}

	rsize, rversion := fcall.buf.Rversion()
	if err = fcall.buf.Err(); err != nil {
		return err
	}
	if rversion != version {
		return errors.New("server supports " + rversion + " only")
	}
	log.Debugf("-> Rversion tag:%d msize:%d version:%q", proto.NoTag, rsize, rversion)
	return nil
}

// Auth is used to authenticate users on a connection. If the server does
// require authentication, it returns a fid defining a file of type
// proto.QTAUTH that may be read and written to execute an authentication
// protocol. That protocol's definition is not part of 9P2000 itself.
func (c *Client) Auth(ctx context.Context, user, root string) (*Fid, error) {
	panic("not implemented")
}

// Access identifies the user (uname) and may select the file tree to
// access (aname). The afid argument specifies a fid previously
// established by an auth message. Access returns the root directory of
// the desired file tree.
func (c *Client) Access(ctx context.Context, fid *Fid, user, root string) (*Fid, error) {
	tag, fcall, err := c.register(true)
	if err != nil {
		return nil, err
	}
	defer fcall.reset()

	afid := uint32(proto.NoFid)
	if fid != nil {
		afid = fid.num
	}

	f := fcall.fid
	log.Debugf("<- Tattach tag:%d fid:%d afid:%d uname:%q aname:%q", tag, f.num, afid, user, root)
	if err = c.enc.Tattach(tag, f.num, afid, user, root); err != nil {
		c.deregister(tag)
		return nil, err
	}
	if err = fcall.wait(ctx); err != nil {
		return nil, err
	}

	typ, version, path := fcall.buf.Rattach()
	if err = fcall.buf.Err(); err != nil {
		return nil, err
	}
	f.qid = proto.Qid{Type: typ, Version: version, Path: path}
	log.Debugf("-> Rattach tag:%d qid:(%s)", tag, f.qid)

	return f, nil
}

func (c *Client) recv() (err error) {
	for err == nil {
		typ, tag, buf, decErr := c.dec.Next()
		if err = decErr; err != nil {
			break
		}

		fcall, found := c.deregister(tag)
		if !found {
			err = fmt.Errorf("pending fcall (tag:%d) not found", tag)
			break
		}

		if typ < proto.Rversion || typ > proto.Rwstat || typ%2 != 1 {
			err = fmt.Errorf("received invalid fcall type (%d)", typ)
			fcall.done(nil, err)
			break
		}

		fcall.done(buf, nil)
	}

	c.enc.writer.Lock()
	c.mu.Lock()
	c.shutdown = true
	if err == io.EOF {
		if !c.closing {
			err = io.ErrUnexpectedEOF
		} else {
			err = errClientShutdown
		}
	}
	for _, fcall := range c.pending {
		fcall.done(nil, err)
	}
	c.mu.Unlock()
	c.enc.writer.Unlock()
	return err
}
