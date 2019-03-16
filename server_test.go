package ninep

import (
	"context"
	"net"
	"testing"
)

type testServer struct {
	listener net.Listener
	s        *Server

	done chan struct{}
}

func startTestServer(t *testing.T, network, addr string) *testServer {
	t.Helper()

	listener, err := net.Listen(network, addr)
	if err != nil {
		t.Fatalf("cannot start %s://%s: %v", network, addr, err)
	}

	s := &testServer{
		done:     make(chan struct{}),
		listener: listener,
		s:        newServer(),
	}

	go func() {
		if err := s.s.Listen(listener); err != nil {
			//if err != io.EOF && err != io.ErrUnexpectedEOF {
			//	t.Errorf("%s://%s: %v", network, addr, err)
			//}
		}
		close(s.done)
	}()

	return s
}

func (s *testServer) Close() error { return s.listener.Close() }

func (s *testServer) Done() <-chan struct{} { return s.done }

func testProtocolHandshake(t *testing.T, addr net.Addr) {
	c, err := Dial(context.Background(), addr.Network(), addr.String())
	if err != nil {
		t.Fatalf("dial %s://%s: %v", addr.Network(), addr.String(), err)
	}
	if err = c.Close(); err != nil {
		t.Fatalf("close %s://%s: %v", addr.Network(), addr.String(), err)
	}
}

func TestServerClientCodec(t *testing.T) {
	s := startTestServer(t, "tcp", "localhost:0")
	defer func() {
		s.listener.Close()
		<-s.Done()
	}()

	addr := s.listener.Addr()

	testProtocolHandshake(t, addr)
}
