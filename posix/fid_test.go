package posix

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestFidAttach(t *testing.T) {
	uid, gid := getTestUser(t)
	f, err := Attach(newTestPosixFS(t), nil, "/", "", uid)
	if err != nil {
		t.Fatalf("fid: unexpected attach error: %v", err)
	}

	if f.uid != uid {
		t.Fatalf("fid: expected uid %d, got %d", uid, f.uid)
	}
	if f.gid != gid {
		t.Fatalf("fid: expected gid %d, got %d", gid, f.gid)
	}

	if err = f.Close(); err != unix.EBADF {
		t.Fatalf("fid: expected %v, got %v", unix.EBADF, err)
	}
}
