package ninep

import (
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
}

func newServerConn(rwc io.ReadWriteCloser) *serverConn {
	return &serverConn{
		enc: proto.NewEncoder(rwc),
		dec: proto.NewDecoder(rwc),
		rwc: rwc,
	}
}

func (c *serverConn) Close() error { return c.rwc.Close() }

// TODO: Tversion -> Rversion
func (c *serverConn) Handshake() error { return nil }

func (c *serverConn) Serve() (err error) {
	for {
		typ, buf, decErr := c.dec.Next()
		if err = decErr; err != nil {
			break // TODO: retry strategy
		}

		go c.handle(typ, buf)
	}
	return err
}

func (c *serverConn) handle(typ proto.FcallType, buf *proto.Buffer) {
	switch typ {
	case proto.Tauth:
		c.tauth(buf)
	case proto.Tattach:
		c.tattach(buf)
	case proto.Tflush:
		c.tflush(buf)

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

func (c *serverConn) tauth(buf *proto.Buffer)   {}
func (c *serverConn) tattach(buf *proto.Buffer) {}
func (c *serverConn) tflush(buf *proto.Buffer)  {}

type Server struct{}

func (s *Server) Listen(listener net.Listener) (err error) {
	for {
		conn, netErr := listener.Accept()
		if err = netErr; err != nil {
			break
		}

		go func(conn net.Conn) {
			c := newServerConn(conn)
			_ = c.Handshake()
			_ = c.Serve()
			c.Close()
		}(conn)
	}

	return err
}
