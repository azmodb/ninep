package ninep

import (
	"io"
	"net"
	"sync"
	"testing"
)

type mockListener struct {
	conn  chan net.Conn
	state chan struct{}
	once  sync.Once
}

func newMockListener() *mockListener {
	return &mockListener{
		conn:  make(chan net.Conn),
		state: make(chan struct{}),
	}
}

func (l *mockListener) Dial() (net.Conn, error) {
	select {
	case <-l.state:
		return nil, io.ErrUnexpectedEOF
	default:
	}

	server, client := net.Pipe()
	l.conn <- server
	return client, nil
}

func (l *mockListener) Accept() (net.Conn, error) {
	select {
	case server := <-l.conn:
		return server, nil
	case <-l.state:
		return nil, io.ErrUnexpectedEOF
	}
}

func (l *mockListener) Close() error {
	l.once.Do(func() { close(l.state) })
	return nil
}

func (l *mockListener) Addr() net.Addr { return mockAddr{} }

type mockAddr struct{}

func (mockAddr) Network() string { return "mock" }
func (mockAddr) String() string  { return "mock" }

func isClosedChan(c <-chan struct{}) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

func TestSessionHandling(t *testing.T) {
	num := 8

	l := newMockListener()
	s := NewServer(nil)

	done := make(chan error, 1)
	go func() { done <- s.Listen(l) }()

	var clients []*Client

	for i := 0; i < num; i++ {
		conn, _ := l.Dial()
		c, err := newClient(conn)
		if err != nil {
			t.Fatalf("client: cannot initialize connection: %v", err)
		}

		c.handshake("9P2000.L")
		clients = append(clients, c)
	}

	if len(s.sessions) != num {
		t.Fatalf("server: expected %d open sessions, got %d", num, len(s.sessions))
	}

	l.Close()
	<-done

	if len(s.sessions) != 0 {
		t.Fatalf("server: expected all sessions closed, got %d", len(s.sessions))
	}
}
