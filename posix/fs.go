package posix

import (
	"math"
	"os"
	"os/user"
	"runtime"
	"strconv"

	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

var (
	_ (FileSystem) = (*posixFS)(nil)   // posixFS implements a chrooted FileSystem
	_ (File)       = (*posixFile)(nil) // unixfile implements File
)

// FileSystem is the filesystem interface. Any simulated or real system
// should implement this interface.
type FileSystem interface {
	Mknod(path string, perm os.FileMode, major, minor uint32, uid, gid int) error

	Create(path string, flags int, perm os.FileMode, uid, gid int) (File, error)
	Open(path string, flags int, uid, gid int) (File, error)
	Remove(path string, uid, gid int) error

	StatFS(path string) (*StatFS, error)
	Stat(path string) (*Stat, error)

	Lookup(username string, uid int) (gid int, err error)

	Close() error
}

// File represents a file in the filesystem and is a set of operations
// corresponding to a single node.
type File interface {
	WriteAt(p []byte, offset int64) (int, error)
	ReadAt(p []byte, offset int64) (int, error)
	ReadDir() ([]Record, error)
	Close() error
}

// Stat describes a file system object.
type Stat = unix.Stat_t

// StatFS describes a file system.
type StatFS = unix.Statfs_t

type posixFS struct {
	root string
	euid int
	egid int
}

// Open opens a FileSystem implementation backed by the underlying
// operating system's file system.
func Open(root string, uid, gid int) (FileSystem, error) {
	return newPosixFS(root, uid, gid)
}

func newPosixFS(root string, euid, egid int) (*posixFS, error) {
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, unix.ENOENT
	}

	if euid > math.MaxUint32 || egid > math.MaxUint32 {
		return nil, unix.EINVAL
	}
	if euid < 0 {
		euid = unix.Geteuid()
	}
	if egid < 0 {
		egid = unix.Getegid()
	}
	if err := unix.Setreuid(-1, euid); err != nil {
		return nil, err
	}
	if err := unix.Setregid(-1, egid); err != nil {
		return nil, err
	}

	return &posixFS{root: root, euid: euid, egid: egid}, nil
}

func (fs *posixFS) Close() error { return nil }

func (fs *posixFS) Lookup(username string, uid int) (int, error) {
	var u *user.User
	var err error
	if uid < 0 || uid >= math.MaxUint32 {
		u, err = user.Lookup(username)
	} else {
		u, err = user.LookupId(strconv.FormatInt(int64(uid), 10))
	}
	if err != nil {
		return math.MaxUint32, err
	}

	gid, err := strconv.ParseInt(u.Gid, 10, 64)
	if err != nil {
		return math.MaxUint32, err
	}
	return int(gid), err
}

func (fs *posixFS) setid(euid, egid int) (err error) {
	runtime.LockOSThread()
	if euid != fs.euid {
		if err = unix.Setreuid(-1, euid); err != nil {
			runtime.UnlockOSThread()
			return err
		}
	}
	if egid != fs.egid {
		if err = unix.Setregid(-1, egid); err != nil {
			runtime.UnlockOSThread()
		}
	}
	return err
}

func (fs *posixFS) resetid(euid, egid int) {
	if euid != fs.euid {
		if err := unix.Setreuid(-1, fs.euid); err != nil {
			runtime.UnlockOSThread()
			log.Panicf("posixfs: reseteuid error: %v", err)
		}
	}
	if egid != fs.egid {
		if err := unix.Setregid(-1, fs.egid); err != nil {
			runtime.UnlockOSThread()
			log.Panicf("posixfs: resetegid error: %v", err)
		}
	}
	runtime.UnlockOSThread()
}

type posixFile struct {
	f *os.File
}

func (fs *posixFS) Create(path string, flags int, perm os.FileMode, uid, gid int) (File, error) {
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

	if err := fs.setid(uid, gid); err != nil {
		return nil, err
	}
	defer fs.resetid(uid, gid)

	file, err := os.OpenFile(path, flags, perm)
	if err != nil {
		return nil, err
	}
	return &posixFile{f: file}, err
}

func (fs *posixFS) Open(path string, flags int, uid, gid int) (File, error) {
	if flags&os.O_CREATE != 0 {
		flags &= ^os.O_CREATE
	}

	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	if err := fs.setid(uid, gid); err != nil {
		return nil, err
	}
	defer fs.resetid(uid, gid)

	file, err := os.OpenFile(path, flags, 0)
	if err != nil {
		return nil, err
	}
	return &posixFile{f: file}, err
}

func (fs *posixFS) Remove(path string, uid, gid int) (err error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return unix.EPERM
	}

	if err = fs.setid(uid, gid); err != nil {
		return err
	}
	defer fs.resetid(uid, gid)

	return os.Remove(path)
}

func mkdev(major uint32, minor uint32) int {
	return int(unix.Mkdev(major, minor))
}

func (fs *posixFS) Mknod(path string, perm os.FileMode, major, minor uint32, uid, gid int) (err error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return unix.EPERM
	}

	if err = fs.setid(uid, gid); err != nil {
		return err
	}
	defer fs.resetid(uid, gid)

	return unix.Mknod(path, uint32(perm), mkdev(major, minor))
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

func (fs *posixFS) StatFS(path string) (*StatFS, error) {
	path, ok := chroot(fs.root, path)
	if !ok {
		return nil, unix.EPERM
	}

	statfs := &unix.Statfs_t{}
	if err := unix.Statfs(path, statfs); err != nil {
		return nil, &os.PathError{Op: "statfs", Path: path, Err: err}
	}
	return statfs, nil
}

func (f *posixFile) WriteAt(p []byte, offset int64) (int, error) {
	if f == nil || f.f == nil {
		return 0, unix.EBADF
	}
	return f.f.WriteAt(p, offset)
}

func (f *posixFile) ReadAt(p []byte, offset int64) (int, error) {
	if f == nil || f.f == nil {
		return 0, unix.EBADF
	}
	return f.f.ReadAt(p, offset)
}

func (f *posixFile) ReadDir() ([]Record, error) {
	if f == nil || f.f == nil {
		return nil, unix.EBADF
	}
	return readDir(int(f.f.Fd()))
}

func (f *posixFile) Close() error { return f.f.Close() }
