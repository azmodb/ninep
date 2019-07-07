package posix

import (
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"testing"

	"github.com/azmodb/pkg/log"
)

func newTestPosixFS(t *testing.T) *posixFS {
	t.Helper()

	root, err := ioutil.TempDir("", "ninep-posix-test")
	if err != nil {
		log.Panicf("cannot create posix package test directory: %v", err)
	}

	fs, err := newPosixFS(root, -1, -1)
	if err != nil {
		log.Panicf("cannot init posix package filesystem: %v", err)
	}
	return fs
}

func getTestUser(t *testing.T) (uid, gid int) {
	user, err := user.Current()
	if err != nil {
		log.Panicf("cannot get current user: %v", err)
	}

	euid, err := strconv.ParseInt(user.Uid, 10, 64)
	if err != nil {
		log.Panicf("cannot parse test user id: %v", err)
	}
	egid, err := strconv.ParseInt(user.Gid, 10, 64)
	if err != nil {
		log.Panicf("cannot parse test group id: %v", err)
	}
	return int(euid), int(egid)
}

func TestCreateOpenRemove(t *testing.T) {
	euid, egid := getTestUser(t)
	fs := newTestPosixFS(t)
	defer os.RemoveAll(fs.root)

	for num, test := range []struct {
		name     string
		flags    int
		perm     os.FileMode
		uid, gid int
	}{
		{"f1", os.O_RDONLY, 0644, euid, egid},
	} {
		f1, err := fs.Create(test.name, test.flags, test.perm, test.uid, test.gid)
		if err != nil {
			t.Fatalf("unix(#%d): unexpected create error: %v", num, err)
		}

		//name := f1.(*posixFile).f.Name()
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
