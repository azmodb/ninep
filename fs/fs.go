// Package fs provides the definitions and functions for interacting with
// the filesystem, as an abstraction layer.
package fs

import (
	"os"

	"github.com/azmodb/ninep/proto"
	"golang.org/x/sys/unix"
)

// FileSystem is the filesystem interface. Any simulated or real system
// should implement this interface.
type FileSystem interface {
	Create(path string, flags int, mode os.FileMode, uid, gid uint32) (File, error)
	Open(path string, flags int, uid, gid uint32) (File, error)
	MakeDir(path string, mode os.FileMode, uid, git uint32) error
	Remove(path string, uid, gid uint32) error

	Stat(path string) (uint64, *Stat, error)

	Close() error
}

// File represents a file in the filesystem and is a set of operations
// corresponding to a single node.
type File interface {
	ReadDir(p []byte, offset int64) (int, error)

	WriteAt(p []byte, offset int64) (int, error)
	ReadAt(p []byte, offset int64) (int, error)

	Close() error
}

// Stat describes a file system object.
type Stat = unix.Stat_t

// StatFS describes a file system.
type StatFS = unix.Statfs_t

var (
	_ (FileSystem) = (*osFS)(nil)   // osFS implements FileSystem
	_ (File)       = (*osFile)(nil) // osFile implements File
)

type osFS struct {
	root string
	uid  uint32
	gid  uint32
}

type osFile struct {
	ents []proto.Dirent // directory entries cache TODO
	*os.File
}

func (fs *osFS) Close() error { return nil }

func (fs *osFS) Create(path string, flags int, mode os.FileMode, uid, gid uint32) (File, error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	file, err := os.OpenFile(path, flags, mode)
	if err != nil {
		return nil, err
	}
	return &osFile{File: file}, nil
}

func (fs *osFS) Open(path string, flags int, uid, gid uint32) (File, error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	file, err := os.OpenFile(path, flags, 0)
	if err != nil {
		return nil, err
	}
	return &osFile{File: file}, nil
}

func (fs *osFS) Remove(path string, uid, gid uint32) error {
	path, ok := chroot(fs.root, path)
	if !ok {
		return unix.EPERM
	}

	return os.Remove(path)
}

func (fs *osFS) MakeDir(path string, mode os.FileMode, uid, git uint32) error {
	path, ok := chroot(fs.root, path)
	if !ok {
		return unix.EPERM
	}

	return os.Mkdir(path, mode)
}

// stat returns an *Stat describing the named file. If there is an error,
// it will be of type *os.PathError.
func (fs *osFS) Stat(path string) (uint64, *Stat, error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return 0, nil, unix.EPERM
	}

	st := &unix.Stat_t{}
	if err := unix.Stat(path, st); err != nil {
		return 0, nil, &os.PathError{Op: "stat", Path: path, Err: err}
	}
	return proto.GetAttrBasic, st, nil
}

func (f *osFile) ReadDir(p []byte, offset int64) (n int, err error) {
	if offset < 0 {
		return 0, unix.EINVAL
	}

	if len(f.ents) == 0 { // initialize directory info cache
		f.ents, err = readDirent(int(f.Fd()))
		if err != nil {
			return 0, err
		}
	}

	if offset >= int64(len(f.ents)) {
		return 0, unix.EIO
	}

	for _, ent := range f.ents[offset:] {
		if len(p[n:]) < ent.Len() {
			break
		}
		_, m, err := ent.Marshal(p[n:])
		n += m
		if err != nil {
			break
		}
	}
	return n, err
}
