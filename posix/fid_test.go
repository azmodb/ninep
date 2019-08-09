package posix

import (
	"path/filepath"
	"testing"

	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

func testMalformedWalkPath(t *testing.T, f *Fid, names []string) {
	t.Helper()

	if _, err := f.Walk(names, func(st *Stat) error {
		return nil
	}); err == nil {
		log.Fatalf("walk: unexpected <nil> error, %v", err)
	}
}

func openWalkRoot(t *testing.T) *Fid {
	t.Helper()

	dir := filepath.Join("testdata", "walk")
	fs, err := newPosixFS(dir, -1, -1)
	if err != nil {
		log.Panicf("walk: cannot init posix filesystem: %v", err)
	}
	defer fs.Close()

	f, err := newFid(fs, "/", -1, -1)
	if err != nil {
		log.Fatalf("walk: cannot init new fid: %v", err)
	}
	return f
}

func TestMalformedWalkPath(t *testing.T) {
	root := openWalkRoot(t)
	defer root.Close()

	testMalformedWalkPath(t, root, []string{"a", "b", "c", "X"})
	testMalformedWalkPath(t, root, []string{"a", "b", "X"})
	testMalformedWalkPath(t, root, []string{"a", "X"})
	testMalformedWalkPath(t, root, []string{"X"})

	// []string{".."} // "redirects" to /
	// []string{"../.."} // "redirects" to /
}

func TestWalk(t *testing.T) {
	root := openWalkRoot(t)
	defer root.Close()

	names := []string{"a", "b", "c", "file"}
	fid, err := root.Walk(names, func(st *Stat) error {
		return nil
	})
	if err != nil {
		log.Fatalf("walk: cannot walk %v: %v", names, err)
	}
	defer fid.Close()

	st, err := fid.Stat()
	if err != nil {
		log.Fatalf("walk: cannot stat: %v", err)
	}
	if st.Mode&unix.S_IFDIR != 0 {
		log.Fatalf("walk: expected file, got %d", st.Mode)
	}
}

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
