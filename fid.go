package ninep

import (
	"errors"
	"os"
	"path"
	"sync"
	"time"

	"github.com/azmodb/ninep/proto"
	"golang.org/x/sys/unix"
)

// fileInfo describes a file, implements os.FileInfo and is returned by
// Stat (Tgetattr).
type fileInfo struct {
	rx   *proto.Rgetattr
	name string
}

func (fi fileInfo) Name() string { return fi.name }
func (fi fileInfo) Size() int64  { return int64(fi.rx.Size) }

func (fi fileInfo) Mode() os.FileMode {
	return fi.rx.Mode.FileMode()
}

func (fi fileInfo) ModTime() time.Time {
	return time.Unix(fi.rx.Mtime.Sec, fi.rx.Mtime.Nsec)
}

func (fi fileInfo) IsDir() bool      { return fi.rx.IsDir() }
func (fi fileInfo) Sys() interface{} { return fi.rx }

var (
	errInvalildName = errors.New("invalid file or directory name")
	errInvalidUid   = errors.New("invalid user identifier")

	errFidNotOpened = errors.New("fid not opened for I/O")
	errFidOpened    = errors.New("fid already opened")
)

const separator = "/"

// Fid structures are intended to represent a remote file descriptor in
// a 9P connection
type Fid struct {
	c   *Client
	num uint32
	uid uint32
	gid uint32

	mu      sync.Mutex // protects following
	iounit  uint32
	path    string
	opened  bool
	closing bool
}

// Num returns the numerical fid identifier.
func (f *Fid) Num() uint32 { return f.num }

// Close informs the file server that the current file represented by
// fid is no longer needed by the client.
func (f *Fid) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closing {
		return errFidNotOpened
	}

	f.opened = false
	f.closing = true
	tx, rx := &proto.Tclunk{Fid: f.num}, &proto.Rclunk{}
	return f.c.rpc(proto.MessageTclunk, tx, rx)
}

func (f *Fid) stat(mask uint64) (*proto.Rgetattr, error) {
	tx := &proto.Tgetattr{Fid: f.num, RequestMask: mask}
	rx := &proto.Rgetattr{}
	if err := f.c.rpc(proto.MessageTgetattr, tx, rx); err != nil {
		return nil, err
	}
	return rx, nil
}

// Stat returns information about a file represented by fid. Execute
// (search) permission is required on all of the directories in path
// that lead to the file.
func (f *Fid) Stat() (os.FileInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// if !f.opened {
	// 	return nil, errFidNotOpened
	// }

	rx, err := f.stat(proto.GetAttrBasic)
	if err != nil {
		return nil, err
	}
	return fileInfo{name: path.Base(f.path), rx: rx}, nil
}

// Create asks the file server to create a new file with the name
// supplied, in the directory represented by fid, and requires write
// permission in the directory.
//
// Finally, the newly created file is opened according to perm, and fid
// will represent the newly opened file.
func (f *Fid) Create(name string, flag int, perm os.FileMode) error {
	if isReserved(name) {
		return errInvalildName
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	tx := &proto.Tlcreate{
		Fid:   f.num,
		Name:  name,
		Flags: proto.NewFlag(flag),
		Perm:  proto.NewMode(perm),
		Gid:   f.gid,
	}
	rx := &proto.Rlcreate{}
	if err := f.c.rpc(proto.MessageTlcreate, tx, rx); err != nil {
		return err
	}

	f.iounit = rx.Iounit
	f.opened = true
	return nil
}

// Mkdir asks the file server to create a new directory with the name
// supplied, in the directory represented by fid, and requires write
// permission in the directory.
func (f *Fid) Mkdir(name string, perm os.FileMode) error {
	if isReserved(name) {
		return errInvalildName
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	tx := &proto.Tmkdir{
		DirectoryFid: f.num,
		Name:         name,
		Perm:         proto.NewMode(perm),
		Gid:          f.gid,
	}
	rx := &proto.Rmkdir{}
	return f.c.rpc(proto.MessageTmkdir, tx, rx)
}

// Open opens the file represented by fid with specified flag
// (os.O_RDONLY etc.). If successful, it can be used for I/O.
func (f *Fid) Open(flag int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.opened {
		return errFidOpened
	}

	tx := &proto.Tlopen{Fid: f.num, Flags: proto.NewFlag(flag)}
	rx := &proto.Rlopen{}
	if err := f.c.rpc(proto.MessageTlopen, tx, rx); err != nil {
		return err
	}

	f.iounit = rx.Iounit
	f.opened = true
	return nil
}

// Remove asks the file server both to remove the file represented by
// fid and to close the fid, even if the remove fails. This request will
// fail if the client does not have write permission in the parent
// directory.
//
// It is correct to consider remove to be a clunk with the side effect
// of removing the file if permissions allow.
func (f *Fid) Remove() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.opened {
		return errFidNotOpened
	}

	f.opened = false
	f.closing = true
	tx, rx := &proto.Tremove{Fid: f.num}, &proto.Rremove{}
	return f.c.rpc(proto.MessageTremove, tx, rx)
}

func (f *Fid) Walk(names ...string) (*Fid, error) {
	f.mu.Lock()
	fid, err := f.walk(names)
	f.mu.Unlock()
	return fid, err
}

func (f *Fid) walk(names []string) (*Fid, error) {
	newFid, ok := f.c.fid.Get()
	if !ok {
		return nil, errFidOverflow
	}

	tx := &proto.Twalk{Fid: f.num, NewFid: newFid, Names: names}
	rx := &proto.Rwalk{}
	if err := f.c.rpc(proto.MessageTwalk, tx, rx); err != nil {
		return nil, err
	}

	return nil, nil
}

func (f *Fid) ReadAt(p []byte, offset int64) (int, error) {
	return 0, unix.ENOTSUP
}

func (f *Fid) WriteAt(p []byte, offset int64) (int, error) {
	return 0, unix.ENOTSUP
}

func (f *Fid) ReadDir(n int) ([]os.FileInfo, error) {
	return nil, unix.ENOTSUP
}

func (f *Fid) Link(oldname, newname string) error {
	return unix.ENOTSUP
}

func (f *Fid) Mknod(name string, perm os.FileMode, major, minor uint32) error {
	return unix.ENOTSUP
}

func (f *Fid) Rename(oldpath, newpath string) error {
	return unix.ENOTSUP
}

func (f *Fid) ReadLink(name string) (string, error) {
	return "", unix.ENOTSUP
}

func (f *Fid) Sync() error { return unix.ENOTSUP }

// isReserved returns whetever name is a reserved filesystem name.
func isReserved(name string) bool {
	return name == "" || name == "." || name == ".."
}
