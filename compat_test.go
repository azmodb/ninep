// +build compat

package ninep

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"os"
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
	t.Helper()

	f, err := c.Attach(nil, path, "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("attach: %v", err)
	}
	return f
}

func random() string {
	data := make([]byte, 16)
	if _, err := rand.Reader.Read(data); err != nil {
		panic("out of entropy")
	}
	return fmt.Sprintf("%x", data)
}

func checkFidIsDir(t *testing.T, f *Fid) {
	t.Helper()

	if !f.fi.IsDir() {
		t.Fatalf("fid: expected directory fid, got %q", f.fi.Qid)
	}
}

func checkFidIsFile(t *testing.T, f *Fid) {
	t.Helper()

	if f.fi.IsDir() {
		t.Fatalf("fid: expected file fid, got %q", f.fi.Qid)
	}
}

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

	checkFidIsDir(t, f)
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

	if err = f.Create(random(), os.O_RDONLY, 0644); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err = f.Remove(); err != nil {
		t.Fatalf("remove: %v", err)
	}
}

/*
func TestWalk(t *testing.T) {
	t.Parallel()

	c := newCompatTestClient(t)
	defer c.Close()

	f := attach(t, c, "/tmp/dir-test")
	defer f.Close()

	dir1, err := f.Walk("dir1")
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	defer dir1.Close()
	checkFidIsDir(t, dir1)

	sub1, err := dir1.Walk("sub1")
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	defer sub1.Close()
	checkFidIsDir(t, sub1)

	file1, err := sub1.Walk("subsub1", "file1")
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	defer file1.Close()
	checkFidIsFile(t, file1)
}
*/
