package ninep

import (
	"net"
	"testing"
)

func newTestClient(t *testing.T, conn net.Conn) *Client {
	c, err := newClient(conn)
	if err != nil {
		t.Fatalf("client: cannot initialize connection: %v", err)
	}
	return c
}
