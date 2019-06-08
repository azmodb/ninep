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
	*proto.Rgetattr
	path   string
	iounit uint32
}

func (fi fileInfo) Mode() os.FileMode { return fi.Rgetattr.Mode.FileMode() }
func (fi fileInfo) Name() string      { return path.Base(fi.path) }
func (fi fileInfo) Size() int64       { return int64(fi.Rgetattr.Size) }
func (fi fileInfo) Sys() interface{}  { return fi.Rgetattr }
func (fi fileInfo) Iounit() uint32    { return fi.iounit }

func (fi fileInfo) ModTime() time.Time {
	return time.Unix(fi.Mtime.Sec, fi.Mtime.Nsec)
}

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

	mu      sync.Mutex // protects following
	fi      *fileInfo
	opened  bool
	closing bool
}

// Num returns the numerical fid identifier.
func (f *Fid) Num() uint32 { return f.num }

// Close informs the file server that the current file represented by
// fid is no longer needed by the client.
func (f *Fid) Close() error {
	f.mu.Lock()
	err := f.clunk()
	f.mu.Unlock()
	return err
}

func (f *Fid) clunk() error {
	if f.closing {
		return errFidNotOpened
	}

	f.opened = false
	f.closing = true

	fcall := f.c.mustAlloc(proto.MessageTclunk)
	tx := fcall.Tx.(*proto.Tclunk)
	tx.Fid = f.num
	err := f.c.rpc(fcall)
	proto.Release(fcall)
	return err
}

// Stat returns information about a file represented by fid. Execute
// (search) permission is required on all of the directories in path
// that lead to the file.
func (f *Fid) Stat() (os.FileInfo, error) {
	f.mu.Lock()
	attr, err := f.c.stat(f.num, proto.GetAttrBasic)
	if err != nil {
		f.mu.Unlock()
		return nil, err
	}

	fi := fileInfo{Rgetattr: attr, path: f.fi.path, iounit: f.fi.iounit}
	f.mu.Unlock()
	return fi, nil
}

func (f *Fid) walk(names ...string) (uint32, *fileInfo, error) {
	if len(names) > proto.MaxNames {
		return 0, nil, unix.EINVAL
	}

	fidnum, ok := f.c.fid.Get()
	if !ok {
		return 0, nil, errFidOverflow
	}

	fcall := f.c.mustAlloc(proto.MessageTwalk)
	defer proto.Release(fcall)

	tx := fcall.Tx.(*proto.Twalk)
	tx.Fid = f.num
	tx.NewFid = uint32(fidnum)
	tx.Names = names
	if err := f.c.rpc(fcall); err != nil {
		return 0, nil, err
	}
	rx := fcall.Rx.(*proto.Rwalk)
	if len(*rx) != len(names) {
		return 0, nil, unix.ENOENT
	}

	attr, err := f.c.stat(tx.NewFid, proto.GetAttrBasic)
	if err != nil {
		return 0, nil, err
	}

	path := path.Join(f.fi.path, path.Join(names...))
	return tx.NewFid, &fileInfo{Rgetattr: attr, path: path}, nil
}

func (f *Fid) Walk(names ...string) (*Fid, error) {
	f.mu.Lock()
	fidnum, fi, err := f.walk(names...)
	if err != nil {
		f.mu.Unlock()
		return nil, err
	}

	fid := &Fid{c: f.c, num: fidnum, fi: fi}
	f.mu.Unlock()
	return fid, nil
}

func (f *Fid) clone() (*Fid, error) {
	fidnum, fi, err := f.walk()
	if err != nil {
		return nil, err
	}
	return &Fid{c: f.c, num: fidnum, fi: fi}, nil
}

/*
func (f *Fid) parse(data []byte) ([]proto.Dirent, uint64, error) {
	var ents []proto.Dirent
	var offset uint64
	var err error

	for len(data) > 0 {
		dirent := proto.Dirent{}
		if data, err = dirent.Unmarshal(data); err != nil {
			return ents, offset, err
		}
		if isReserved(dirent.Name) {
			continue
		}

		ents = append(ents, dirent)
		offset = dirent.Offset
	}
	return ents, offset, err
}

func (f *Fid) readDirent(n int) (ents []proto.Dirent, err error) {
	tx := &proto.Treaddir{Fid: f.num, Offset: 0, Count: f.c.maxDataSize}
	rx := &proto.Rreaddir{}

	for {
		rx.Reset()
		if err = f.c.rpc(tx, rx); err != nil {
			if err == unix.EIO {
				break
			}
			return nil, err
		}
		if len(rx.Data) < proto.FixedDirentSize {
			break
		}

		entries, offset, err := f.parse(rx.Data)
		if err != nil {
			return nil, err
		}
		tx.Offset = offset
		ents = append(ents, entries...)
	}
	return ents, nil
}

func (f *Fid) readDir(n int) (info []os.FileInfo, err error) {
	clone, err := f.clone()
	if err != nil {
		return nil, err
	}
	defer clone.clunk()

	if err = clone.open(os.O_RDONLY); err != nil {
		return nil, err
	}

	ents, err := clone.readDirent(n)
	if err != nil {
		return nil, err
	}

	for _, dirent := range ents {
		_, fi, err := f.walk(dirent.Name)
		if err != nil {
			return nil, err
		}
		info = append(info, fi)
	}
	return info, err
}

func (f *Fid) ReadDir(n int) ([]os.FileInfo, error) {
	f.mu.Lock()
	info, err := f.readDir(n)
	f.mu.Unlock()
	return info, err
}
*/

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
	err := f.create(name, flag, perm)
	f.mu.Unlock()
	return err
}

func (f *Fid) create(name string, flag int, perm os.FileMode) error {
	fcall := f.c.mustAlloc(proto.MessageTlcreate)
	defer proto.Release(fcall)

	tx := fcall.Tx.(*proto.Tlcreate)
	tx.Fid = f.num
	tx.Name = name
	tx.Flags = proto.NewFlag(flag)
	tx.Perm = proto.NewMode(perm)
	tx.Gid = f.fi.Gid
	if err := f.c.rpc(fcall); err != nil {
		return err
	}

	rx := fcall.Rx.(*proto.Rlcreate)
	f.fi.iounit = rx.Iounit
	f.opened = true
	return nil
}

// Mkdir asks the file server to create a new directory with the name
// supplied, in the directory represented by fid, and requires write
// permission in the directory.
func (f *Fid) Mkdir(name string, perm os.FileMode) error {
	f.mu.Lock()
	err := f.mkdir(name, perm)
	f.mu.Unlock()
	return err
}

func (f *Fid) mkdir(name string, perm os.FileMode) error {
	if isReserved(name) {
		return errInvalildName
	}

	fcall := f.c.mustAlloc(proto.MessageTmkdir)
	tx := fcall.Tx.(*proto.Tmkdir)
	tx.DirectoryFid = f.num
	tx.Name = name
	tx.Perm = proto.NewMode(perm)
	tx.Gid = f.fi.Gid
	err := f.c.rpc(fcall)
	proto.Release(fcall)
	return err
}

// Open opens the file represented by fid with specified flag
// (os.O_RDONLY etc.). If successful, it can be used for I/O.
func (f *Fid) Open(flag int) error {
	f.mu.Lock()
	err := f.open(flag)
	f.mu.Unlock()
	return err
}

func (f *Fid) open(flag int) error {
	if f.opened {
		return errFidOpened
	}

	fcall := f.c.mustAlloc(proto.MessageTlopen)
	defer proto.Release(fcall)

	tx := fcall.Tx.(*proto.Tlopen)
	tx.Fid = f.num
	tx.Flags = proto.NewFlag(flag)
	if err := f.c.rpc(fcall); err != nil {
		return err
	}

	rx := fcall.Rx.(*proto.Rlopen)
	f.fi.iounit = rx.Iounit
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
	err := f.remove()
	f.mu.Unlock()
	return err
}

func (f *Fid) remove() error {
	if !f.opened {
		return errFidNotOpened
	}

	f.opened = false
	f.closing = true

	fcall := f.c.mustAlloc(proto.MessageTremove)
	tx := fcall.Tx.(*proto.Tremove)
	tx.Fid = f.num
	err := f.c.rpc(fcall)
	proto.Release(fcall)
	return err
}

func (f *Fid) ReadAt(p []byte, offset int64) (int, error) {
	return 0, unix.ENOTSUP
}

func (f *Fid) WriteAt(p []byte, offset int64) (int, error) {
	return 0, unix.ENOTSUP
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
