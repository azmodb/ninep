package ninep

import (
	"bytes"
	"testing"
)

/*
func startTestServer(network, addr string) (io.Closer, *Server, error) {
	logger := log.New(os.Stderr, "", log.LstdFlags)
	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, nil, err
	}

	s, err := NewServer(logger)
	if err != nil {
		listener.Close()
		return nil, nil, err
	}

	go func() { s.Serve(listener) }()

	return listener, s, nil
}

func dial(network, addr string) (io.Closer, error) {
	c, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return c, nil
}
*/

type rwc struct {
	*bytes.Buffer
}

func (r rwc) Close() error { return nil }

func TestConnectionHandling(t *testing.T) {
	s, _ := NewServer(nil, nil)
	clients := 6

	check := func(npending, nfreeid int) {
		if len(s.pending) != npending {
			t.Fatalf("expected %d pending conn, have %d",
				npending, len(s.pending))
		}
		if len(s.freeid) != nfreeid {
			t.Fatalf("expected %d free ids, have %d",
				nfreeid, len(s.freeid))
		}
	}

	ids := make([]int64, 0, clients)
	for i := 0; i < clients; i++ {
		id, err := s.addConn(newConn(s, rwc{}))
		if err != nil {
			t.Fatalf("add test conn: %v", err)
		}
		ids = append(ids, id)
	}

	check(clients, 0)

	for i := 0; i < len(ids)/2; i++ {
		s.deleteConn(ids[i])
	}

	check(clients/2, clients/2)

	s.destroyConns()
	check(0, clients)
}
