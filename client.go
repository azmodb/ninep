package ninep

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"

	"github.com/azmodb/ninep/proto"
)

type ConnOption func(*Conn) error

type Conn struct {
	enc *encoder // exclusive 9P2000 stream encoder
	dec proto.Decoder
	c   io.Closer

	mu       sync.Mutex // protects following
	pending  map[uint16]chan<- interface{}
	freetag  map[uint16]struct{}
	freefid  map[uint32]struct{}
	tag      uint16
	fid      uint32
	closing  bool
	shutdown bool
}

func Dial(ctx context.Context, network, address string, opts ...ConnOption) (*Conn, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewConn(conn, opts...)
}

func NewConn(rwc io.ReadWriteCloser, opts ...ConnOption) (*Conn, error) {
	c := &Conn{
		pending: make(map[uint16]chan<- interface{}),
		freetag: make(map[uint16]struct{}),
		freefid: make(map[uint32]struct{}),

		enc: newEncoder(rwc),
		dec: newDecoder(rwc),
		c:   rwc,
	}
	go c.recv()

	if err := c.version(24+4*8192, "9P2000"); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

func (c *Conn) Close() error {
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return errors.New("connection is shut down")
	}
	c.closing = true
	c.mu.Unlock()
	return c.c.Close()
}

func send(ch chan<- interface{}, m interface{}) {
	select {
	case ch <- m:
	default:
	}
}

func wait(ch <-chan interface{}, rx proto.Message) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("received invalid response type")
		}
	}()

	m := <-ch
	switch rx := m.(type) {
	case proto.Rerror:
		return proto.Error(rx.Ename())
	case error:
		return rx
	}

	switch rx := rx.(type) {
	case *proto.Rversion:
		*rx = m.(proto.Rversion)
	case *proto.Rauth:
		*rx = m.(proto.Rauth)
	case *proto.Rattach:
		*rx = m.(proto.Rattach)
	case *proto.Rflush:
		*rx = m.(proto.Rflush)
	case *proto.Rwalk:
		*rx = m.(proto.Rwalk)
	case *proto.Ropen:
		*rx = m.(proto.Ropen)
	case *proto.Rcreate:
		*rx = m.(proto.Rcreate)
	case *proto.Rread:
		*rx = m.(proto.Rread)
	case *proto.Rwrite:
		*rx = m.(proto.Rwrite)
	case *proto.Rclunk:
		*rx = m.(proto.Rclunk)
	case *proto.Rremove:
		*rx = m.(proto.Rremove)
	case *proto.Rstat:
		*rx = m.(proto.Rstat)
	case *proto.Rwstat:
		*rx = m.(proto.Rwstat)
	default:
		err = errors.New("received invalid response type")
	}
	return err
}

// version request negotiates the protocol version and message size to
// be used on the connection and initializes the connection for I/O.
//
// Not part of the API and NOT syncronized.
func (c *Conn) version(msize uint32, version string) (err error) {
	ch := make(chan interface{}, 1)
	c.pending[proto.NOTAG] = ch

	log.Printf("<- Tversion tag:%d msize:%d version:%q", proto.NOTAG, msize, version)
	if err = c.enc.Tversion(proto.NOTAG, msize, version); err != nil {
		delete(c.pending, proto.NOTAG)
		return err
	}

	var rx proto.Rversion
	if err = wait(ch, &rx); err != nil {
		return err
	}

	if rx.Msize() > c.dec.MaxMessageSize() {
		return errors.New("invalid msize in version response")
	}
	if rx.Msize() != c.dec.MaxMessageSize() {
		c.enc.SetMaxMessageSize(rx.Msize())
		c.dec.SetMaxMessageSize(rx.Msize())
	}
	log.Printf("-> Rversion tag:%d msize:%d version:%q", rx.Tag(), rx.Msize(), rx.Version())
	return nil
}

func (c *Conn) Auth(ctx context.Context, user, root string) (*Fid, error) {
	panic("not implemented")
}

func (c *Conn) Access(ctx context.Context, fid *Fid, user, root string) (*Fid, error) {
	tag, num, ch, err := c.register()
	if err != nil {
		return nil, err
	}

	afid := proto.NOFID
	if fid != nil {
		afid = fid.num
	}

	log.Printf("<- Tattach tag:%d fid:%d afid:%d uname:%q aname:%q", tag, num, afid, user, root)
	if err = c.enc.Tattach(tag, num, afid, user, root); err != nil {
		c.deregister(tag)
		return nil, err
	}

	var rx proto.Rattach
	if err = wait(ch, &rx); err != nil {
		return nil, err
	}
	log.Printf("<- Rattach tag:%d qid:%s", tag, rx.Qid())
	return &Fid{num: num, qid: rx.Qid(), c: c}, err
}

func (c *Conn) nextTag() (uint16, error) {
	var tag uint16
	for tag, _ = range c.freetag {
		delete(c.freetag, tag)
	}
	if tag == 0 {
		c.tag++
		if c.tag == proto.NOTAG {
			return 0, errors.New("out of tags")
		}
		tag = c.tag
	}
	return tag, nil
}

func (c *Conn) nextFid() (uint32, error) {
	var fid uint32
	for fid, _ = range c.freefid {
		delete(c.freefid, fid)
	}
	if fid == 0 {
		c.fid++
		if c.fid == proto.NOFID {
			return 0, errors.New("out of fids")
		}
		fid = c.fid
	}
	return fid, nil
}

func (c *Conn) register() (uint16, uint32, <-chan interface{}, error) {

	c.mu.Lock()
	tag, err := c.nextTag()
	if err != nil {
		c.mu.Unlock()
		return 0, 0, nil, err
	}
	fid, err := c.nextFid()
	if err != nil {
		c.mu.Unlock()
		return 0, 0, nil, err
	}

	ch := make(chan interface{}, 1)
	c.pending[tag] = ch
	c.mu.Unlock()

	return tag, fid, ch, nil
}

func (c *Conn) deregister(tag uint16) chan<- interface{} {
	c.mu.Lock()
	ch := c.pending[tag]
	delete(c.pending, tag)
	c.mu.Unlock()
	return ch
}

func (c *Conn) recv() {
	var err error
	for err == nil {
		var m proto.Message
		if m, err = c.dec.Decode(); err != nil {
			break
		}

		ch := c.deregister(m.Tag())
		if ch == nil {
			// We've got no pending channel. That usually means that
			// encode() partially failed, and the channel was already
			// removed.
			continue
		}

		send(ch, m)
	}

	c.enc.mu.Lock()
	c.mu.Lock()
	c.shutdown = true
	if err == io.EOF {
		if c.closing {
			err = errors.New("connection is shut down")
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, ch := range c.pending {
		send(ch, err)
	}
	c.mu.Unlock()
	c.enc.mu.Unlock()
}

type Fid struct {
	qid proto.Qid
	num uint32
	c   *Conn
}

func (f *Fid) Walk(ctx context.Context, names ...string) (*Fid, error) {
	return nil, nil
}

func (f *Fid) Create(ctx context.Context, name string, perm uint32, mode uint8) error {
	return nil
}

func (f *Fid) Open(ctx context.Context, mode uint8) error { return nil }

func (f *Fid) Remove(ctx context.Context) error { return nil }

func (f *Fid) Write(ctx context.Context, data []byte, offset int64) (int, error) {
	return 0, nil
}

func (f *Fid) Read(ctx context.Context, data []byte, offset int64) (int, error) {
	return 0, nil
}

func (f *Fid) Close() error { return nil }
