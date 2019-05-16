// +build compat

package ninep

import (
	"context"
	"fmt"
	"math"
	"testing"
)

// func init() { log.SetLevel(log.InfoLevel) }

const diodTestServerAddr = "127.0.0.1:5640"

const diodMaxMessageSize = 64 * 1024

func newCompatTestClient(t *testing.T) *Client {
	t.Helper()

	c, err := Dial(context.Background(), "tcp", diodTestServerAddr)
	if err != nil {
		t.Fatalf("cannot dial diod test-server: %v", err)
	}
	return c
}

func attach(t *testing.T, c *Client, path string) *Fid {
	f, err := c.Attach(nil, path, "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	return f
}

/*
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

func TestAttach(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
	defer c.Close()

	f, err := c.Attach(nil, "/tmp", "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("clunk: %v", err)
		}
	}()
}

func TestCreateRemove(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
	defer c.Close()

	f, err := c.Attach(nil, "/tmp", "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	defer f.Close()

	if err = f.Create("create_remove_test", os.O_RDONLY, 0644); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err = f.Remove(); err != nil {
		t.Fatalf("remove: %v", err)
	}
}
*/

func TestWalk(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
	defer c.Close()

	//names := []string{"dir1", "sub1", "subsub1", "file1"}
	names := []string{"dir1"}
	f := attach(t, c, "/tmp/dir-test")
	defer f.Close()

	dir1, err := f.Walk(names...)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	defer dir1.Close()

	sub1, err := dir1.Walk("sub1")
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	defer sub1.Close()

	//	if err = dir1.Open(os.O_RDONLY); err != nil {
	//		t.Fatalf("open: %v", err)
	//	}

	entries, err := dir1.ReadDir(-1)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	for _, ent := range entries {
		fmt.Println("\t", ent.Name())
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
