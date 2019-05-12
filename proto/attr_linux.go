package proto

import (
	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// EncodeStat encodes information about a file system object. The byte
// sequence follow pretty closely the fields returned by the Linux
// stat(2) system call.
//
// Valid is a bitmask indicating which fields are valid in the response.
func EncodeStat(buf *binary.Buffer, valid uint64, stat *Stat) error {
	buf.PutUint64(valid)

	buf.PutUint8(UnixFileTypeToQidType(stat.Mode)) // marshal Qid
	buf.PutUint32(uint32(stat.Mtim.Nano() ^ stat.Size<<8))
	buf.PutUint64(stat.Ino)

	buf.PutUint32(stat.Mode)
	buf.PutUint32(stat.Uid)
	buf.PutUint32(stat.Gid)
	buf.PutUint64(stat.Nlink)
	buf.PutUint64(stat.Rdev)
	buf.PutUint64(uint64(stat.Size))
	buf.PutUint64(uint64(stat.Blksize))
	buf.PutUint64(uint64(stat.Blocks))
	encodeTimespec(buf, stat.Atim)
	encodeTimespec(buf, stat.Mtim)
	encodeTimespec(buf, stat.Ctim)

	encodeTimespec(buf, btime)
	buf.PutUint64(0)
	buf.PutUint64(0)

	return buf.Err()
}

var btime = unix.NsecToTimespec(0)

// UnixStatToRgetattr converts a unix.Stat_t returned by the stat(2)
// system call.
func UnixStatToRgetattr(stat *Stat) *Rgetattr {
	return &Rgetattr{
		Qid: Qid{
			UnixFileTypeToQidType(stat.Mode),
			uint32(stat.Mtim.Nano() ^ stat.Size<<8),
			stat.Ino,
		},

		Mode:      stat.Mode,
		Uid:       stat.Uid,
		Gid:       stat.Gid,
		Nlink:     stat.Nlink,
		Rdev:      stat.Rdev,
		Size:      uint64(stat.Size),
		BlockSize: uint64(stat.Blksize),
		Blocks:    uint64(stat.Blocks),
		Atime:     stat.Atim,
		Mtime:     stat.Mtim,
		Ctime:     stat.Ctim,

		Btime:       btime,
		Gen:         0,
		DataVersion: 0,
	}
}

// EncodeStatfs encodes information about a file system. The byte
// sequence follow pretty closely the fields returned by the Linux
// statfs(2) system call.
func EncodeStatfs(buf *binary.Buffer, stat *StatFS) error {
	buf.PutUint32(uint32(stat.Type))
	buf.PutUint32(uint32(stat.Bsize))
	buf.PutUint64(stat.Blocks)
	buf.PutUint64(stat.Bfree)
	buf.PutUint64(stat.Bavail)
	buf.PutUint64(stat.Files)
	buf.PutUint64(stat.Ffree)
	buf.PutUint64(uint64(stat.Fsid.Val[0] | stat.Fsid.Val[1]<<32))
	buf.PutUint32(uint32(stat.Namelen))

	return buf.Err()
}
