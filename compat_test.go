// +build compat

package ninep

import (
	"context"
	"math"
	"os"
	"testing"
)

// func init() { log.SetLevel(log.InfoLevel) }

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

func newCompatTestClient(t *testing.T) *Client {
	t.Helper()

	c, err := Dial(context.Background(), "tcp", diodTestServerAddr)
	if err != nil {
		t.Fatalf("cannot dial diod test-server: %v", err)
	}
	return c
}

func TestCompatAttach(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
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

func TestCompatCreateRemove(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
	defer c.Close()

	f, err := c.Attach(nil, "/tmp", "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("compat attach: %v", err)
	}
	defer f.Close()

	if err = f.Create("create_remove_test", os.O_RDONLY, 0644); err != nil {
		t.Fatalf("compat create: %v", err)
	}

	if err = f.Remove(); err != nil {
		t.Fatalf("compat remove: %v", err)
	}
}

func attach(t *testing.T, c *Client, path string) *Fid {
	f, err := c.Attach(nil, path, "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	return f
}

func TestBasicWalk(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
	defer c.Close()

	names := []string{"dir1", "sub1", "subsub1", "file1"}
	f := attach(t, c, "/tmp/dir-test")
	defer f.Close()

	n, err := f.Walk(names...)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	fi, err := n.Stat()
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if fi.Name() != names[len(names)-1] {
		t.Fatalf("walk: expected file name %q, got %q", fi.Name(), names[len(names)-1])
	}
}

// func TestCompatMkdirRemove(t *testing.T) {
// 	t.Parallel()

// 	c := newCompatTestClient(t)
// 	defer c.Close()

// 	f, err := c.Attach(nil, "/tmp", "root", math.MaxUint32)
// 	if err != nil {
// 		t.Fatalf("compat attach: %v", err)
// 	}
// 	defer f.Close()

// 	if err = f.Mkdir("mkdir_remove_test", 0755); err != nil {
// 		t.Fatalf("compat mkdir: %v", err)
// 	}

// 	// if err = f.Remove(); err != nil {
// 	// 	t.Fatalf("compat remove: %v", err)
// 	// }
// }
