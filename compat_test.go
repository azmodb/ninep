package ninep

import (
	"context"
	"math"
	"testing"
)

const diodTestServerAddr = "127.0.0.1:5640"

const diodMaxMessageSize = 64 * 1024

func TestCompatHandshake(t *testing.T) {
	c, err := Dial(context.Background(), "tcp", diodTestServerAddr)
	if err != nil {
		t.Fatalf("cannot dial diod test-server: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Fatalf("closing connection: %v", err)
		}
	}()

	if c.maxMessageSize != diodMaxMessageSize {
		t.Fatalf("compat: expected maxMessageSize = %d, got %d",
			diodMaxMessageSize, c.maxMessageSize)
	}
	if c.maxDataSize != diodMaxMessageSize-24 {
		t.Fatalf("compat: expected maxDataSize = %d, got %d",
			diodMaxMessageSize-24, c.maxDataSize)
	}
}

func newTestClient(t *testing.T) *Client {
	c, err := Dial(context.Background(), "tcp", diodTestServerAddr)
	if err != nil {
		t.Fatalf("cannot dial diod test-server: %v", err)
	}
	return c
}

func TestCompatAttach(t *testing.T) {
	c := newTestClient(t)
	defer c.Close()

	f, err := c.Attach(nil, "/tmp", "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("compat attach: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("compat clunk: %v", err)
		}
	}()
}
