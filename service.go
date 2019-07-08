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
}

func newService(fs posix.FileSystem) *service {
	return &service{fs: fs, fidmap: newFidmap()}
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
	return unix.ENOTSUP
}

func (s *service) flush(ctx context.Context, tx *proto.Tflush, rx *proto.Rflush) unix.Errno {
	return unix.ENOTSUP
}
