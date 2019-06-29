package posix

import (
	"os"

	"golang.org/x/sys/unix"
)

var (
	_ (FileSystem) = (*posixFS)(nil)   // posixFS implements a chrooted FileSystem
	_ (File)       = (*posixFile)(nil) // unixfile implements File
)

// FileSystem is the filesystem interface. Any simulated or real system
// should implement this interface.
type FileSystem interface {
	Create(path string, flags int, perm os.FileMode, uid, gid uint32) (File, error)
	Open(path string, flags int, uid, gid uint32) (File, error)
	Remove(path string, uid, gid uint32) error
	Mknod(path string, perm os.FileMode, major, minor, uid, gid uint32) error

	Stat(path string) (*Stat, error)

	Close() error
}

// File represents a file in the filesystem and is a set of operations
// corresponding to a single node.
type File interface {
	Close() error
}

type Stat = unix.Stat_t

type posixFS struct {
	root string
	uid  uint32
	gid  uint32
}

type posixFile struct {
	f *os.File
}

func (fs *posixFS) Close() error { return nil }

func (fs *posixFS) Create(path string, flags int, perm os.FileMode, uid, gid uint32) (File, error) {
	if flags&os.O_CREATE == 0 {
		flags |= os.O_CREATE
	}
	if flags&os.O_EXCL == 0 {
		flags |= os.O_EXCL
	}
	if flags&os.O_TRUNC == 0 {
		flags |= os.O_TRUNC
	}

	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	//	if err := fs.setfsid(uid, gid); err != nil {
	//		return nil, err
	//	}
	//	defer fs.resetfsid()

	file, err := os.OpenFile(path, flags, perm)
	if err != nil {
		return nil, err
	}
	return &posixFile{f: file}, err
}

func (fs *posixFS) Open(path string, flags int, uid, gid uint32) (File, error) {
	if flags&os.O_CREATE != 0 {
		flags &= ^os.O_CREATE
	}

	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	//	if err := fs.setfsid(uid, gid); err != nil {
	//		return nil, err
	//	}
	//	defer fs.resetfsid()

	file, err := os.OpenFile(path, flags, 0)
	if err != nil {
		return nil, err
	}
	return &posixFile{f: file}, err
}

func (fs *posixFS) Remove(path string, uid, gid uint32) (err error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return unix.EPERM
	}

	//	if err = fs.setfsid(uid, gid); err != nil {
	//		return err
	//	}
	err = os.Remove(path)
	//	fs.resetfsid()
	return err
}

func mkdev(major uint32, minor uint32) int {
	return int(unix.Mkdev(major, minor))
}

func (fs *posixFS) Mknod(path string, perm os.FileMode, major, minor, uid, gid uint32) (err error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return unix.EPERM
	}

	//	if err = fs.setfsid(uid, gid); err != nil {
	//		return err
	//	}

	err = unix.Mknod(path, uint32(perm), mkdev(major, minor))
	//	fs.resetfsid()
	return err
}

func (fs *posixFS) Stat(path string) (*Stat, error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	stat := &unix.Stat_t{}
	if err := unix.Stat(path, stat); err != nil {
		return nil, &os.PathError{Op: "stat", Path: path, Err: err}
	}
	return stat, nil
}

func (f *posixFile) Close() error { return f.f.Close() }
