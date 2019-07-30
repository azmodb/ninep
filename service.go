package ninep

import (
	"context"

	"github.com/azmodb/ninep/posix"
	"github.com/azmodb/ninep/proto"
	"golang.org/x/sys/unix"
)

// service defines a 9P2000.L service. In case of an error the handle
// function must return a unix.Errno which will be sent back to the
// client.
type service struct {
	fs     posix.FileSystem
	fidmap *fidmap
	valid  uint64
}

func newService(fs posix.FileSystem) *service {
	return &service{fs: fs, fidmap: newFidmap(), valid: proto.GetAttrAll /* TODO */}
}

func (s *service) attach(ctx context.Context, tx *proto.Tlattach, rx *proto.Rlattach) unix.Errno {
	if tx.AuthFid != proto.NoFid { // TODO: implement authentication
		return unix.EINVAL
	}

	f, err := posix.Attach(s.fs, nil, tx.Path, tx.UserName, int(tx.Uid))
	if err != nil {
		return newErrno(err)
	}

	stat, err := f.Stat()
	if err != nil {
		return newErrno(err)
	}

	// An error is returned if fid is already in use. See:
	//      http://9p.io/magic/man2html/5/attach
	if !s.fidmap.Attach(tx.Fid, newFid(f)) {
		return unix.EBUSY
	}

	rx.Qid = proto.StatToQid(stat)
	return 0
}

func (s *service) auth(ctx context.Context, tx *proto.Tlauth, rx *proto.Rlauth) unix.Errno {
	return unix.ENOSYS
}

func (s *service) flush(ctx context.Context, tx *proto.Tflush, rx *proto.Rflush) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) getattr(ctx context.Context, tx *proto.Tgetattr, rx *proto.Rgetattr) unix.Errno {
	f, found := s.fidmap.Load(tx.Fid)
	if !found {
		return unix.EBADF
	}
	defer f.DecRef()

	stat, err := f.Stat()
	if err != nil {
		return newErrno(err)
	}
	rx.Valid = s.valid
	rx.Stat_t = stat
	return 0
}

func (s *service) setattr(ctx context.Context, tx *proto.Tsetattr, rx *proto.Rsetattr) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) walk(ctx context.Context, tx *proto.Twalk, rx *proto.Rwalk) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) read(ctx context.Context, tx *proto.Tread, rx *proto.Rread) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) write(ctx context.Context, tx *proto.Twrite, rx *proto.Rwrite) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) clunk(ctx context.Context, tx *proto.Tclunk, rx *proto.Rclunk) unix.Errno {
	return 0
}

func (s *service) statfs(ctx context.Context, tx *proto.Tstatfs, rx *proto.Rstatfs) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) create(ctx context.Context, tx *proto.Tlcreate, rx *proto.Rlcreate) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) open(ctx context.Context, tx *proto.Tlopen, rx *proto.Rlopen) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) remove(ctx context.Context, tx *proto.Tremove, rx *proto.Rremove) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) symlink(ctx context.Context, tx *proto.Tsymlink, rx *proto.Rsymlink) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) mknod(ctx context.Context, tx *proto.Tmknod, rx *proto.Rmknod) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) rename(ctx context.Context, tx *proto.Trename, rx *proto.Rrename) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) readlink(ctx context.Context, tx *proto.Treadlink, rx *proto.Rreadlink) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) xattrwalk(ctx context.Context, tx *proto.Txattrwalk, rx *proto.Rxattrwalk) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) xattrcreate(ctx context.Context, tx *proto.Txattrcreate, rx *proto.Rxattrcreate) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) readdir(ctx context.Context, tx *proto.Treaddir, rx *proto.Rreaddir) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) fsync(ctx context.Context, tx *proto.Tfsync, rx *proto.Rfsync) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) lock(ctx context.Context, tx *proto.Tlock, rx *proto.Rlock) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) getlock(ctx context.Context, tx *proto.Tgetlock, rx *proto.Rgetlock) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) link(ctx context.Context, tx *proto.Tlink, rx *proto.Rlink) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) mkdir(ctx context.Context, tx *proto.Tmkdir, rx *proto.Rmkdir) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) renameat(ctx context.Context, tx *proto.Trenameat, rx *proto.Rrenameat) unix.Errno {
	return unix.ENOTSUP
}

func (s *service) unlinkat(ctx context.Context, tx *proto.Tunlinkat, rx *proto.Runlinkat) unix.Errno {
	return unix.ENOTSUP
}
