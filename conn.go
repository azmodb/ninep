package ninep

import (
	"fmt"
	"io"
	"sync"

	"github.com/azmodb/ninep/proto"
)

type ErrorResponder interface {
	Rerrorf(format string, args ...interface{})
	Rerror(err error)
}

type errResponder struct {
	tag uint16
	c   *conn
}

func (r errResponder) Rerrorf(format string, args ...interface{}) {
	r.c.rerrorf(r.tag, format, args...)
}

func (r errResponder) Rerror(err error) { r.c.rerror(r.tag, err) }

type session struct {
	//parent context.Context
	uid string
	c   *conn
}

func (r *session) Uid() string { return r.uid }

type Tattach struct {
	ErrorResponder
	*session
	tag uint16

	Fid      uint32
	Afid     uint32
	Username string
	Root     string
}

func (r *Tattach) Rattach(qid proto.Qid) {
	r.c.rattach(r.tag, qid)
}

type conn struct {
	writer sync.Mutex // exclusive connection writer
	enc    proto.Encoder

	dec proto.Decoder

	//	parent context.Context
	//	fs     FileServer
	s *Server

	mu       sync.Mutex // protects conn state
	shutdown bool
	c        io.Closer

	muxer    sync.Mutex // protects request muxer
	sessions uint16
	session  map[uint16]*session
}

func newConn(s *Server, rwc io.ReadWriteCloser) *conn {
	return &conn{
		enc: proto.NewEncoder(
			rwc,
			proto.WithMaxMessageSize(s.msize),
			proto.WithLogger(s.logger),
		),
		dec: proto.NewDecoder(
			rwc,
			proto.WithMaxMessageSize(s.msize),
			proto.WithLogger(s.logger),
		),
		s:       s,
		c:       rwc,
		session: make(map[uint16]*session),
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

func (c *conn) rerrorf(tag uint16, fmt string, args ...interface{}) {
	c.writer.Lock()
	if c.enc.Err() != nil {
		c.writer.Unlock()
		return
	}
	c.enc.Rerrorf(tag, fmt, args...)
	c.enc.Flush()
	c.writer.Unlock()
}

func (c *conn) rerror(tag uint16, err error) {
	c.writer.Lock()
	if c.enc.Err() != nil {
		c.writer.Unlock()
		return
	}
	c.enc.Rerror(tag, err)
	c.enc.Flush()
	c.writer.Unlock()
}

func (c *conn) rattach(tag uint16, qid proto.Qid) {
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

func (c *conn) handleTauth(m proto.Tauth) {
	//return errors.New("authentification not required")
}

func (c *conn) handleTattach(m proto.Tattach) {
	tx := &Tattach{
		ErrorResponder: errResponder{tag: m.Tag(), c: c},
		session: &session{
			uid: m.Uname(),
			c:   c,
			//parent: c.parent,
		},
		tag:      m.Tag(),
		Fid:      m.Fid(),
		Afid:     m.Afid(),
		Username: m.Uname(),
		Root:     m.Aname(),
	}

	if ok := c.s.fs.Tattach(nil, tx); !ok {
		return
	}

	c.muxer.Lock()
	c.sessions++
	c.session[c.sessions] = tx.session
	c.muxer.Unlock()
}

func (c *conn) Serve() error {
	m, err := c.dec.Decode()
	if err != nil {
		return err
	}
	tx, ok := m.(proto.Tversion)
	if !ok {
		c.rerrorf(m.Tag(), "unexpected message (%T)", m)
		return fmt.Errorf("unexpected message (%T)", m)
	}
	if err = c.handleTversion(tx); err != nil {
		return err
	}

	for err == nil {
		m, err := c.dec.Decode()
		if err != nil {
			return err
		}

		switch m := m.(type) {
		case proto.Tauth:
			//err = c.handleTauth(m)
		case proto.Tattach:
			//err = c.handleTattach(m)
			c.handleTattach(m)
		default:
			c.rerrorf(m.Tag(), "unexpected message (%T)", m)
			err = fmt.Errorf("unexpected message (%T)", m)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
