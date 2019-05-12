package ninep

import (
	"os"
	"path"
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
	return proto.FileMode(fi.rx.Mode).FileMode()
}

func (fi fileInfo) ModTime() time.Time {
	return time.Unix(fi.rx.Mtime.Sec, fi.rx.Mtime.Nsec)
}

func (fi fileInfo) IsDir() bool {
	return fi.rx.Mode&unix.S_IFMT == unix.S_IFDIR
}

func (fi fileInfo) Sys() interface{} { return fi.rx }

type Fid struct {
	c    *Client
	path string
	num  uint32
	uid  uint32
	gid  uint32
}

func (f *Fid) Num() uint32 { return f.num }

func (f *Fid) Close() error {
	tx, rx := &proto.Tclunk{Fid: f.num}, &proto.Rclunk{}
	if err := f.c.rpc(proto.MessageTclunk, tx, rx); err != nil {
		return err
	}
	return nil
}

func (f *Fid) stat(mask uint64) (*proto.Rgetattr, error) {
	tx := &proto.Tgetattr{Fid: f.num, RequestMask: mask}
	rx := &proto.Rgetattr{}
	if err := f.c.rpc(proto.MessageTgetattr, tx, rx); err != nil {
		return nil, err
	}
	return rx, nil
}

func (f *Fid) Stat() (os.FileInfo, error) {
	rx, err := f.stat(proto.GetAttrBasic)
	if err != nil {
		return nil, err
	}
	return fileInfo{name: path.Base(f.path), rx: rx}, nil
}

/*
func (f *Fid) Create(name string, flag int, perm os.FileMode) error {
	_ = &proto.Tlcreate{
		Fid:        f.num,
		Name:       name,
		Flags:      uint32(proto.NewFlag(flag)),
		Permission: uint32(proto.NewFileMode(perm)),
		Gid:        0, // TODO
	}
	return nil
}
*/

// isReserved returns whetever name is a reserved filesystem name.
func isReserved(name string) bool {
	return name == "" || name == "." || name == ".."
}
