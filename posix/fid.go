package posix

import (
	"os"

	"golang.org/x/sys/unix"
)

type Option func(*Fid) error

type Fid struct {
	fs   FileSystem
	file File
	path string
	uid  int
	gid  int
}

func Attach(fs FileSystem, file File, path, username string, uid int, opts ...Option) (*Fid, error) {
	if _, err := fs.Stat(path); err != nil {
		return nil, err
	}
	//	if !fi.IsDir() {
	//		return nil, unix.ENOTDIR
	//	}

	gid, err := fs.Lookup(username, uid)
	if err != nil {
		return nil, err
	}

	return newFid(fs, path, uid, gid, opts...)
}

func newFid(fs FileSystem, path string, uid, gid int, opts ...Option) (*Fid, error) {
	f := &Fid{fs: fs, path: path, uid: uid, gid: gid}
	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, err
		}
	}
	return f, nil
}

func (f *Fid) isOpened() bool { return f.file != nil }

func (f *Fid) Mknod(name string, perm os.FileMode, major, minor uint32, gid int) error {
	return nil
}

func (f *Fid) Create(name string, flags int, perm os.FileMode, gid int) error {
	if !isValidName(name) {
		return unix.EINVAL
	}

	path := join(f.path, name)
	file, err := f.fs.Create(path, flags, perm, f.uid, gid)
	if err != nil {
		return err
	}
	f.file.Close()
	f.file = file
	f.path = path
	return nil
}

func (f *Fid) Open(flags int) error {
	if f.isOpened() {
		return unix.EBADF
	}

	file, err := f.fs.Open(f.path, flags, f.uid, f.gid)
	if err != nil {
		return err
	}
	f.file = file
	return nil
}

func (f *Fid) Remove() error {
	if !f.isOpened() {
		return unix.EBADF
	}

	f.file.Close()
	err := f.fs.Remove(f.path, f.uid, f.gid)
	if err != nil {
		return err
	}
	f.file = nil
	return err
}

func (f *Fid) Stat() (*Stat, error) { return f.fs.Stat(f.path) }

func (f *Fid) Close() error {
	if !f.isOpened() {
		return unix.EBADF
	}

	err := f.file.Close()
	f.file = nil
	return err
}
