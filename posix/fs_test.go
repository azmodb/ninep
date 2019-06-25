package posix

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/azmodb/pkg/log"
)

func newTestUnixFS(t *testing.T) *unixFS {
	t.Helper()

	root, err := ioutil.TempDir("", "ninep-posix-test")
	if err != nil {
		log.Panicf("cannot create posix package test directory")
	}
	return &unixFS{root: root}
}

func TestCreateOpenRemove(t *testing.T) {
	fs := newTestUnixFS(t)
	defer os.RemoveAll(fs.root)

	for num, test := range []struct {
		name     string
		flags    int
		perm     os.FileMode
		uid, gid uint32
	}{
		{"f1", os.O_RDONLY, 0644, 0, 0},
	} {
		f1, err := fs.Create(test.name, test.flags, test.perm, test.uid, test.gid)
		if err != nil {
			t.Fatalf("unix(#%d): unexpected create error: %v", num, err)
		}

		//name := f1.(*unixFile).f.Name()
		_, err = fs.Stat(test.name)
		if err != nil {
			t.Fatalf("unix(#%d): unexpected stat error: %v", num, err)
		}

		f2, err := fs.Open(test.name, test.flags, test.uid, test.gid)
		if err != nil {
			t.Fatalf("unix(#%d): unexpected open error: %v", num, err)
		}

		if err = f1.Close(); err != nil {
			t.Fatalf("unix(#%d): unexpected close error: %v", num, err)
		}
		if err = f2.Close(); err != nil {
			t.Fatalf("unix(#%d): unexpected close error: %v", num, err)
		}

		if err = fs.Remove(test.name, test.uid, test.gid); err != nil {
			t.Fatalf("unix(#%d): unexpected remove error: %v", num, err)
		}
		_, err = fs.Stat(test.name)
		if err == nil {
			t.Fatalf("unix(#%d): expected stat error: %v", num, err)
		}
	}
}
