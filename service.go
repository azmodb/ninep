package ninep

import (
	"github.com/azmodb/ninep/proto"
	"golang.org/x/sys/unix"
)

// Service defines a 9P2000.L service interface. In case of an error the
// handle function must return a unix.Errno which will be sent back
// to the client.
type Service interface {
	Lauth(tx *proto.Tlauth, rx *proto.Rlauth) unix.Errno
	Lattach(tx *proto.Tlattach, rx *proto.Rlattach) unix.Errno
	Flush(tx *proto.Tflush, rx *proto.Rflush) unix.Errno

	Walk(tx *proto.Twalk, rx *proto.Rwalk) unix.Errno
	Read(tx *proto.Tread, rx *proto.Rread) unix.Errno
	Write(tx *proto.Twrite, rx *proto.Rwrite) unix.Errno
	Clunk(tx *proto.Tclunk, rx *proto.Rclunk) unix.Errno
	Remove(tx *proto.Tremove, rx *proto.Rremove) unix.Errno

	Statfs(tx *proto.Tstatfs, rx *proto.Rstatfs) unix.Errno
	Lopen(tx *proto.Tlopen, rx *proto.Rlopen) unix.Errno
	Lcreate(tx *proto.Tlcreate, rx *proto.Rlcreate) unix.Errno
	Symlink(tx *proto.Tsymlink, rx *proto.Rsymlink) unix.Errno
	Mknod(tx *proto.Tmknod, rx *proto.Rmknod) unix.Errno
	Rename(tx *proto.Trename, rx *proto.Rrename) unix.Errno
	Readlink(tx *proto.Treadlink, rx *proto.Rreadlink) unix.Errno
	Getattr(tx *proto.Tgetattr, rx *proto.Rgetattr) unix.Errno
	Setattr(tx *proto.Tsetattr, rx *proto.Rsetattr) unix.Errno
	Xattrwalk(tx *proto.Txattrwalk, rx *proto.Rxattrwalk) unix.Errno
	Xattrcreate(tx *proto.Txattrcreate, rx *proto.Rxattrcreate) unix.Errno
	Readdir(tx *proto.Treaddir, rx *proto.Rreaddir) unix.Errno
	Fsync(tx *proto.Tfsync, rx *proto.Rfsync) unix.Errno
	Lock(tx *proto.Tlock, rx *proto.Rlock) unix.Errno
	Getlock(tx *proto.Tgetlock, rx *proto.Rgetlock) unix.Errno
	Link(tx *proto.Tlink, rx *proto.Rlink) unix.Errno
	Mkdir(tx *proto.Tmkdir, rx *proto.Rmkdir) unix.Errno
	Renameat(tx *proto.Trenameat, rx *proto.Rrenameat) unix.Errno
	Unlinkat(tx *proto.Tunlinkat, rx *proto.Runlinkat) unix.Errno
}
