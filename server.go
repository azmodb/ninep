package ninep

import (
	"errors"
	"io"
	"net"

	"github.com/azmodb/ninep/proto"
)

type serverConn struct {
	// The network connection itself. We expose it in the struct so that
	// it is available for transport-based auth and any timeouts we need
	// to implement.
	rwc io.ReadWriteCloser

	dec *proto.Decoder
	enc *proto.Encoder

	msize int64
}

func newServerConn(rwc io.ReadWriteCloser, msize int64) *serverConn {
	return &serverConn{
		enc:   proto.NewEncoder(rwc),
		dec:   proto.NewDecoder(rwc),
		rwc:   rwc,
		msize: msize,
	}
}

func (c *serverConn) Close() error { return c.rwc.Close() }

func (c *serverConn) Handshake() (err error) {
	// 128 bytes should is enough for a tversion message
	c.dec.SetMaxMessageSize(128)

	typ, tag, buf, err := c.dec.Next()
	if err != nil {
		return err
	}
	if typ != proto.Tversion {
		err = errors.New("malformed handshake request")
		c.enc.Rerror(tag, err.Error())
		return err
	}

	msize, version := buf.Tversion()
	if err = buf.Err(); err != nil {
		c.enc.Rerror(tag, err.Error())
		return err
	}

	if msize < proto.MinMessageSize {
		c.enc.Rerror(tag, "buffer too small")
		return errors.New("buffer too small")
	}
	if msize < c.msize {
		c.msize = msize
	}
	c.enc.SetMaxMessageSize(uint32(c.msize))
	c.dec.SetMaxMessageSize(uint32(c.msize))

	if version != proto.Version {
		version = "unknown"
	}
	return c.enc.Rversion(tag, c.msize, version)
}

func (c *serverConn) Serve() (err error) {
	for {
		typ, tag, buf, decErr := c.dec.Next()
		if err = decErr; err != nil {
			break // TODO: retry strategy
		}

		go c.handle(typ, tag, buf)
	}
	return err
}

func (c *serverConn) handle(typ proto.FcallType, tag uint16, buf *proto.FcallBuffer) {
	switch typ {
	case proto.Tauth:
		c.tauth(tag, buf)
	case proto.Tattach:
		c.tattach(tag, buf)
	case proto.Tflush:
		c.tflush(tag, buf)

	case proto.Twalk:
	case proto.Topen:
	case proto.Tcreate:
	case proto.Tread:
	case proto.Twrite:
	case proto.Tclunk:
	case proto.Tremove:
	case proto.Tstat:
	case proto.Twstat:
	}
}

func (c *serverConn) tauth(tag uint16, buf *proto.FcallBuffer)   {}
func (c *serverConn) tattach(tag uint16, buf *proto.FcallBuffer) {}
func (c *serverConn) tflush(tag uint16, buf *proto.FcallBuffer)  {}

type Server struct {
	msize int64
}

type ServerOption func(*Server) error

func NewServer(opts ...ServerOption) (*Server, error) {
	return newServer(), nil
}

func newServer() *Server {
	return &Server{msize: proto.DefaultMaxMessageSize}
}

func (s *Server) Listen(listener net.Listener) (err error) {
	for {
		conn, netErr := listener.Accept()
		if err = netErr; err != nil {
			break
		}

		go func(conn net.Conn) (err error) {
			c := newServerConn(conn, s.msize)
			defer c.Close()

			if err = c.Handshake(); err != nil {
				return err
			}

			return c.Serve()
		}(conn)
	}

	return err
}
