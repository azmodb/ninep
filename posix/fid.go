package posix

import "os"

func Attach(fs FileSystem, export string, username string, uid, gid uint32) (*Fid, error) {
	return nil, nil
}

type Fid struct{}

func (f *Fid) Create(name string, flags int, perm os.FileMode, gid uint32) error {
	return nil
}

func (f *Fid) Mknod(name string, perm os.FileMode, major, minor, gid uint32) error {
	return nil
}
