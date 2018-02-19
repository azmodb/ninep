package ninep

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/azmodb/ninep/proto"
)

type FileServer interface {
	Tattach(ctx context.Context, m *Tattach) error
}

type reqcontext struct {
	parent context.Context
	uid    string
}

func (m *reqcontext) Uid() string { return m.uid }

type Tattach struct {
	m proto.Tattach
	*reqcontext
	c *conn
}

func (m *Tattach) Fid() uint32   { return m.m.Fid() }
func (m *Tattach) Afid() uint32  { return m.m.Afid() }
func (m *Tattach) Uname() string { return m.m.Uname() }
func (m *Tattach) Aname() string { return m.m.Aname() }

func (m *Tattach) Rattach(qid proto.Qid) {
	m.c.Rattach(m.m.Tag(), qid)
}

type conn struct {
	writer sync.Mutex // exclusive connection writer
	enc    proto.Encoder

	dec proto.Decoder

	parent context.Context
	fs     FileServer

	mu       sync.Mutex
	shutdown bool
	c        io.Closer

	reqmu   sync.Mutex // protects request muxer
	reqs    uint16
	request map[uint16]*reqcontext
}

func newConn(rwc io.ReadWriteCloser, msize uint32, log proto.Logger) *conn {
	return &conn{
		enc: proto.NewEncoder(
			rwc,
			proto.WithMaxMessageSize(msize),
			proto.WithLogger(log),
		),
		dec: proto.NewDecoder(
			rwc,
			proto.WithMaxMessageSize(msize),
			proto.WithLogger(log),
		),
		c: rwc,
	}
}

// Close calls the underlying connection Close method. If the connection
// is already shutting down, Close is a noop.
func (c *conn) Close() error {
	c.mu.Lock()
	if c.shutdown {
		c.mu.Unlock()
		return nil
	}

	err := c.c.Close()
	c.shutdown = true
	c.mu.Unlock()
	return err
}

func (c *conn) Rerrorf(tag uint16, fmt string, args ...interface{}) {
	c.writer.Lock()
	if c.enc.Err() != nil {
		c.writer.Unlock()
		return
	}
	c.enc.Rerrorf(tag, fmt, args...)
	c.enc.Flush()
	c.writer.Unlock()
}

func (c *conn) Rerror(tag uint16, err error) {
	c.writer.Lock()
	if c.enc.Err() != nil {
		c.writer.Unlock()
		return
	}
	c.enc.Rerror(tag, err)
	c.enc.Flush()
	c.writer.Unlock()
}

func (c *conn) Rattach(tag uint16, qid proto.Qid) {
	c.writer.Lock()
	if c.enc.Err() != nil {
		c.writer.Unlock()
		return
	}
	c.enc.Rattach(tag, qid)
	c.enc.Flush()
	c.writer.Unlock()
}

const protoVersion = "9P2000"

func (c *conn) handleTversion(m proto.Tversion) error {
	version := m.Version()
	if version != protoVersion {
		version = "unknown"
	}

	msize := m.Msize()
	if msize > c.dec.MaxMessageSize() {
		msize = c.dec.MaxMessageSize()
	}

	c.writer.Lock()
	c.enc.SetMaxMessageSize(msize)
	c.dec.SetMaxMessageSize(msize)
	c.enc.Rversion(m.Tag(), msize, version)
	err := c.enc.Flush()
	c.writer.Unlock()
	return err
}

func (c *conn) handleTauth(m proto.Tauth) error {
	return errors.New("authentification not required")
}

func (c *conn) handleTattach(m proto.Tattach) error {
	ctx := &reqcontext{
		uid:    m.Uname(),
		parent: c.parent,
	}
	c.reqmu.Lock()
	c.reqs++
	c.request[c.reqs] = ctx
	c.reqmu.Unlock()

	tx := &Tattach{
		reqcontext: ctx,
		c:          c,
		m:          m,
	}
	return c.fs.Tattach(ctx.parent, tx)
}

func (c *conn) Serve() error { return nil }
