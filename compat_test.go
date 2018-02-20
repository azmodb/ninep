package ninep

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"9fans.net/go/plan9/client"
	"github.com/azmodb/ninep/proto"
)

type compatFileServer struct{}

func (fs compatFileServer) Tattach(ctx context.Context, m *Tattach) {
	m.Rattach(proto.Qid{Type: proto.QTDIR})
}

const compatAddr = "localhost:5640"

func startCompatTestServer(t *testing.T) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", compatAddr)
	if err != nil {
		t.Fatalf("cannot create listener: %v", err)
	}

	//fs := testFileServer{}
	//opts := []Option{
	//	WithLogger(log.New(os.Stderr, "", log.LstdFlags)),
	//}
	//s, err := NewServer(fs, opts...)
	logger := log.New(os.Stderr, "", log.LstdFlags)
	s, err := NewServer(compatFileServer{}, logger)
	if err != nil {
		t.Fatalf("cannot create server: %v", err)
	}
	go func() {
		if err := s.Serve(listener); err != nil {
			//t.Fatalf("serve test server: %v", err)
		}
	}()

	return listener
}

func TestCompatDialAndMount(t *testing.T) {
	listener := startCompatTestServer(t)
	defer listener.Close()

	conn, err := client.Dial("tcp", compatAddr)
	if err != nil {
		t.Fatalf("dialing test server: %v", err)
	}
	if err = conn.Close(); err != nil {
		t.Fatalf("closing test server conn: %v", err)
	}

	_, err = client.Mount("tcp", compatAddr)
	if err != nil {
		t.Fatalf("mounting test server: %v", err)
	}
}
