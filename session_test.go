package ninep

import (
	"errors"
	"io"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/azmodb/ninep/proto"
	"golang.org/x/sys/unix"
)

//func init() { log.SetLevel(log.DebugLevel) } // verbose debug messages

func testSessionHandshake(t *testing.T, num int, msize, want uint32) {
	server, client := net.Pipe()

	s := newSession(server, want, calcMaxDataSize(want))
	go s.serve()

	c, err := newClient(client)
	if err != nil {
		t.Fatalf("client: cannot initialize connection: %v", err)
	}
	c.maxMessageSize = msize
	c.maxDataSize = calcMaxDataSize(msize)

	if err := c.handshake(); err != nil {
		t.Errorf("client: handshake failed: %v", err)
	}

	c.Close()
	s.Close()

	if s.dec.MaxMessageSize != want {
		t.Errorf("decoder: expected max message size %d, have %d",
			want, s.dec.MaxMessageSize)
	}
	if s.enc.MaxMessageSize != want {
		t.Errorf("encoder: expected max message size %d, have %d",
			want, s.enc.MaxMessageSize)
	}
}

func TestSessionHandshake(t *testing.T) {
	for num, test := range []struct {
		msize uint32
		want  uint32
	}{
		{proto.MaxMessageSize, proto.MaxMessageSize},
		{8192, 8192},

		{proto.MaxMessageSize + 1, proto.MaxMessageSize},
		{proto.MaxMessageSize + 2, proto.MaxMessageSize},
	} {
		testSessionHandshake(t, num, test.msize, test.want)
	}
}

func calcMaxDataSize(msize uint32) uint32 {
	return msize - (proto.FixedReadWriteSize + 1)
}

func TestNewErrno(t *testing.T) {
	osSyscallError := os.SyscallError{Err: unix.EINVAL}
	osPathError := os.PathError{Err: unix.EINVAL}

	for num, test := range []struct {
		in  error
		out unix.Errno
	}{
		{syscall.EINVAL, unix.EINVAL},
		{unix.EINVAL, unix.EINVAL},

		{&osSyscallError, unix.EINVAL},
		{&osPathError, unix.EINVAL},

		{io.ErrUnexpectedEOF, unix.EIO},
		{unix.EINVAL, unix.EINVAL},
		{errors.New("x"), unix.EIO},

		{nil, 0},
	} {
		if err := newErrno(test.in); err != test.out {
			t.Errorf("errno(%d): expected errno %+v, got %+v", num, test.out, err)
		}
	}
}
