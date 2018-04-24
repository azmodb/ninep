package ninep

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/azmodb/ninep/proto"
)

var (
	errInvalidVersionRequest = errors.New("invalid version request")
	errMessageSizeTooSmall   = errors.New("message size too small")
	errInvalidRequest        = errors.New("invalid request")
)

type ServerOption func(*Server) error

type Server struct {
	fs    FileSystem
	msize uint32
}

func NewServer(fs FileSystem, opts ...ServerOption) (*Server, error) {
	//if msize < proto.MinMessageLen {
	//	msize = proto.DefaultMaxMessageLen
	//}
	return &Server{fs: fs, msize: proto.DefaultMaxMessageLen}, nil
}

func (s *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(conn net.Conn) {
			c := newServerConn(conn, s)
			c.handle(conn)
			conn.Close() // TODO
		}(conn)
	}
}

type serverConn struct {
	dec   proto.Decoder
	enc   *encoder
	msize uint32
	fs    FileSystem
}

func newServerConn(rw io.ReadWriter, s *Server) *serverConn {
	c := &serverConn{
		enc:   newEncoder(rw),
		dec:   newDecoder(rw),
		msize: s.msize,
		fs:    s.fs,
	}
	c.enc.SetMaxMessageSize(s.msize)
	c.dec.SetMaxMessageSize(s.msize)

	return c
}

func (c *serverConn) handle(r io.Reader) {
	if err := c.handleTversion(); err != nil {
		return // TODO
	}

	for {
		m, err := c.dec.Decode()
		if err != nil {
			return // TODO
		}

		switch m.(type) {
		case proto.Tflush:
		case proto.Tauth:
			panic("not implemented")
		case proto.Tattach:
		}

		if tx, ok := m.(ioRequest); ok {
			err = c.handleIO(tx)
		} else {
			err = errInvalidRequest
		}
		if err != nil {
			// TODO
		}
	}
}

type ioRequest interface {
	Fid() uint32
}

func (c *serverConn) handleIO(m ioRequest) error {
	switch m.(type) {
	case proto.Twalk:
	case proto.Topen:
	case proto.Tcreate:
	case proto.Tread:
	case proto.Twrite:
	case proto.Tclunk:
	case proto.Tremove:
	case proto.Tstat:
	case proto.Twstat:
	default:
		return errInvalidRequest
	}
	return nil
}

func (c *serverConn) handleTversion() error {
	m, err := c.dec.Decode()
	if err != nil {
		return err
	}
	tag := m.Tag()

	rx, ok := m.(proto.Tversion)
	if !ok {
		c.enc.Rerror(tag, errInvalidVersionRequest)
		return errInvalidVersionRequest
	}

	msize, version := rx.Msize(), rx.Version()
	log.Printf("-> Tversion tag:%d msize:%d version:%q", tag, msize, version)
	if msize < proto.MinMessageLen {
		c.enc.Rerror(tag, errMessageSizeTooSmall)
		return errMessageSizeTooSmall
	}
	if msize < c.msize {
		c.msize = msize
		c.enc.SetMaxMessageSize(c.msize)
		c.dec.SetMaxMessageSize(c.msize)
	}
	if version != "9P2000" {
		version = "unknown"
	}

	log.Printf("<- Rversion tag:%d msize:%d version:%q", tag, msize, version)
	return c.enc.Rversion(rx.Tag(), msize, version)
}
