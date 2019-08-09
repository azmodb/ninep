package posix

import (
	"os"

	"golang.org/x/sys/unix"
)

// Attach introduces a new user to the file system, and establishes a fid
// as the root for that user on the file tree selected by export. If uid
// is not set to proto.NoUid (~0), is the uid of the user and is used in
// preference to username.
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

// Fid represents a file on the file system.
type Fid struct {
	fs   FileSystem
	file File
	path string
	uid  int
	gid  int
}

// Option sets Fid options.
type Option func(*Fid) error

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

//func (f *Fid) Mknod(name string, perm os.FileMode, major, minor uint32, gid int) error {
//	return nil
//}

func (f *Fid) Walk(names []string, walkfn func(*Stat) error) (*Fid, error) {
	path := f.path
	for _, name := range names {
		path = join(path, name)
		stat, err := f.fs.Stat(path)
		if err != nil {
			return nil, err
		}
		if err = walkfn(stat); err != nil {
			return nil, err
		}
	}
	return newFid(f.fs, path, f.uid, f.gid)
}

// Create creates a regular file name in directory represented by fid
// and prepares it for I/O. After the call fid represents the new file.
//
// Flag contains Linux open(2) flags bits, e.g. O_RDONLY, O_WRONLY, and
// perm contains Linux creat(2) mode bits. Gid is the effective gid of
// the caller.
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

// Open prepares fid for file I/O. Flag contains Linux open(2) flags
// bits, e.g. O_RDONLY, O_WRONLY, O_RDWR.
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

// Remove removes the file system object represented by fid.
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

// Stat returns a Stat describing the named file.
func (f *Fid) Stat() (*Stat, error) { return f.fs.Stat(f.path) }

// StatFS returns file system statistics.
func (f *Fid) StatFS() (*StatFS, error) { return f.fs.StatFS(f.path) }

// Close closes the fid, rendering it unusable for I/O.
func (f *Fid) Close() error {
	if !f.isOpened() {
		return unix.EBADF
	}

	err := f.file.Close()
	f.file = nil
	return err
}

func (f *Fid) ReadAt(p []byte, offset int64) (int, error) {
	if !f.isOpened() {
		return 0, unix.EBADF
	}
	return f.file.ReadAt(p, offset)
}

func (f *Fid) WriteAt(p []byte, offset int64) (int, error) {
	if !f.isOpened() {
		return 0, unix.EBADF
	}
	return f.file.WriteAt(p, offset)
}

// ReadDir returns directory entries from the directory represented by
// fid. Offset is the offset returned in the last directory entry of
// the previous call.
func (f *Fid) ReadDir(p []byte, offset int64) (n int, err error) {
	if !f.isOpened() {
		return 0, unix.EBADF
	}

	records, err := f.file.ReadDir() /// TODO(mason): cache
	if err != nil {
		return 0, err
	}
	if offset >= int64(len(records)) {
		return 0, unix.EIO
	}

	for _, rec := range records[offset:] {
		if len(p[n:]) < rec.Len() {
			break
		}

		m, err := rec.MarshalTo(p[n:])
		n += m
		if err != nil {
			break
		}
	}
	return n, err
}
