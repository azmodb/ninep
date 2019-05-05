package proto

import (
	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// EncodeAttr encodes information about a file system object. The byte
// sequencei follow pretty closely the fields returned by the Linux
// stat(2) system call.
//
// Valid is a bitmask indicating which fields are valid in the response.
func EncodeAttr(buf *binary.Buffer, valid uint64, s *unix.Stat_t) error {
	buf.PutUint64(valid)

	buf.PutUint8(UnixFileTypeToQidType(s.Mode)) // marshal Qid
	buf.PutUint32(uint32(s.Mtim.Nano() ^ s.Size<<8))
	buf.PutUint64(s.Ino)

	buf.PutUint32(s.Mode)
	buf.PutUint32(s.Uid)
	buf.PutUint32(s.Gid)
	buf.PutUint64(s.Nlink)
	buf.PutUint64(s.Rdev)
	buf.PutUint64(uint64(s.Size))
	buf.PutUint64(uint64(s.Blksize))
	buf.PutUint64(uint64(s.Blocks))
	encodeTimespec(buf, s.Atim)
	encodeTimespec(buf, s.Mtim)
	encodeTimespec(buf, s.Ctim)
	encodeTimespec(buf, btime)
	buf.PutUint64(0)
	buf.PutUint64(0)

	return buf.Err()
}

var btime = unix.NsecToTimespec(0)

func EncodeStatfs(buf *binary.Buffer, s *unix.Statfs_t) error {
	return buf.Err()
}
