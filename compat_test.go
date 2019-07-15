package ninep

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"testing"

	"github.com/azmodb/ninep/posix"
	"github.com/azmodb/pkg/log"
)

const (
	diodTestServerAddr = "127.0.0.1:5640"
	testServerAddr     = "127.0.0.1:5641"

	testMaxMessageSize = 64 * 1024
)

func newTestClient(t *testing.T, addr string, opts ...Option) *Client {
	t.Helper()

	c, err := Dial(context.Background(), "tcp", addr, opts...)
	if err != nil {
		t.Fatalf("cannot dial test-server: %v", err)
	}
	return c
}

// TODO(mason): copied from posix/fs_test.go
func newTestPosixFS(t *testing.T) posix.FileSystem {
	t.Helper()

	root, err := ioutil.TempDir("", "ninep-posix-test")
	if err != nil {
		log.Panicf("cannot create posix package test directory: %v", err)
	}

	fs, err := posix.Open(root, -1, -1)
	if err != nil {
		log.Panicf("cannot init posix package filesystem: %v", err)
	}
	return fs
}

func newTestServer(t *testing.T, addr string) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("listener: %v", err)
	}
	s := NewServer(newTestPosixFS(t))
	go func() {
		if err := s.Listen(listener); err != nil {
			//t.Fatalf("fileserver: %v", err) // TODO(mason)
		}
	}()
	return listener
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
		t.Fatalf("fid: expected directory fid, got %q", f.fi)
	}
}

func checkFidIsFile(t *testing.T, f *Fid) {
	t.Helper()

	if f.fi.IsDir() {
		t.Fatalf("fid: expected file fid, got %+v", f.fi)
	}
}

func testHandshake(t *testing.T, num int, addr string) {
	t.Helper()

	c, err := Dial(context.Background(), "tcp", addr,
		WithMaxMessageSize(testMaxMessageSize),
	)
	if err != nil {
		t.Fatalf("handshake(%d): cannot dial test-server: %v", num, err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			t.Fatalf("handshake(%d): closing connection: %v", num, err)
		}
	}()

	if c.maxMessageSize != testMaxMessageSize {
		t.Fatalf("handshake(%d): expected maxMessageSize = %d, got %d",
			num, testMaxMessageSize, c.maxMessageSize)
	}
	if c.maxDataSize != testMaxMessageSize-24 {
		t.Fatalf("handshake(%d): expected maxDataSize = %d, got %d",
			num, testMaxMessageSize-24, c.maxDataSize)
	}
}

/*
func testAttach(t *testing.T, num int, addr string) {
	t.Parallel()

	c := newTestClient(t, addr)
	defer c.Close()

	f, err := c.Attach(nil, "/", "root", math.MaxUint32)
	if err != nil {
		t.Fatalf("attach(%d): %v", num, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Fatalf("clunk(%d): %v", num, err)
		}
	}()

	checkFidIsDir(t, f)
}
*/

func TestCompat(t *testing.T) {
	s := newTestServer(t, testServerAddr)
	defer s.Close()

	for num, addr := range []string{testServerAddr /*, diodTestServer*/} {
		testHandshake(t, num, addr)
		// testAttach(t, num, addr)
	}
}

/*

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
