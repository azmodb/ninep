package posix

import (
	"io"
	"os"
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

func openWalkRoot(t *testing.T, dir string) *Fid {
	t.Helper()

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
	dir := filepath.Join("testdata", "walk")
	root := openWalkRoot(t, dir)
	defer root.Close()

	testMalformedWalkPath(t, root, []string{"a", "b", "c", "X"})
	testMalformedWalkPath(t, root, []string{"a", "b", "X"})
	testMalformedWalkPath(t, root, []string{"a", "X"})
	testMalformedWalkPath(t, root, []string{"X"})

	// []string{".."} // "redirects" to /
	// []string{"../.."} // "redirects" to /
}

func TestWalk(t *testing.T) {
	dir := filepath.Join("testdata", "walk")
	root := openWalkRoot(t, dir)
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

func TestFidReadDir(t *testing.T) {
	want := map[string]bool{
		".":     true,
		"..":    true,
		"dir1":  true,
		"file1": true,
		"file2": true,
		"file3": true,
		"file4": true,
	}
	dir := filepath.Join("testdata", "dirent")
	root := openWalkRoot(t, dir)
	defer root.Close()

	root.Open(os.O_RDONLY)

	data := make([]byte, 64)
	var records []Record
	var offset int64
	for {
		n, err := root.ReadDir(data, offset)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("readdir: read failed: %v", err)
		}

		r, err := UnmarshalRecords(data[:n])
		if err != nil {
			log.Fatalf("readdir: unmarshal failed: %v", err)
		}
		offset = int64(r[len(r)-1].Offset)
		records = append(records, r...)
	}
	if len(records) != len(want) {
		log.Fatalf("readir: exepcted entries %v, got %v", len(want), len(records))
	}
	for _, rec := range records {
		if _, found := want[rec.Name]; !found {
			log.Fatalf("readdir: found unexpected name %q", rec.Name)
		}
	}
}
