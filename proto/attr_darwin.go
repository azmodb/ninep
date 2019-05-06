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
func EncodeStat(buf *binary.Buffer, valid uint64, st *unix.Stat_t) error {
	buf.PutUint64(valid)

	buf.PutUint8(UnixFileTypeToQidType(uint32(st.Mode))) // marshal Qid
	buf.PutUint32(uint32(st.Mtimespec.Nano() ^ st.Size<<8))
	buf.PutUint64(st.Ino)

	buf.PutUint32(uint32(st.Mode))
	buf.PutUint32(st.Uid)
	buf.PutUint32(st.Gid)
	buf.PutUint64(uint64(st.Nlink))
	buf.PutUint64(uint64(st.Rdev))
	buf.PutUint64(uint64(st.Size))
	buf.PutUint64(uint64(st.Blksize))
	buf.PutUint64(uint64(st.Blocks))
	encodeTimespec(buf, st.Atimespec)
	encodeTimespec(buf, st.Mtimespec)
	encodeTimespec(buf, st.Ctimespec)

	encodeTimespec(buf, st.Birthtimespec)
	buf.PutUint64(uint64(st.Gen))
	buf.PutUint64(0)

	return buf.Err()
}

// UnixStatToRgetattr converts a unix.Stat_t returned by the stat(2)
// system call.
func UnixStatToRgetattr(st *unix.Stat_t) *Rgetattr {
	return &Rgetattr{
		Qid: Qid{
			UnixFileTypeToQidType(uint32(st.Mode)),
			uint32(st.Mtimespec.Nano() ^ st.Size<<8),
			st.Ino,
		},

		Mode:      uint32(st.Mode),
		Uid:       st.Uid,
		Gid:       st.Gid,
		Nlink:     uint64(st.Nlink),
		Rdev:      uint64(st.Rdev),
		Size:      uint64(st.Size),
		BlockSize: uint64(st.Blksize),
		Blocks:    uint64(st.Blocks),
		Atime:     st.Atimespec,
		Mtime:     st.Mtimespec,
		Ctime:     st.Ctimespec,

		Btime:       st.Birthtimespec,
		Gen:         uint64(st.Gen),
		DataVersion: 0,
	}
}

// EncodeStatfs encodes information about a file system. The byte
// sequence follow pretty closely the fields returned by the Linux
// statfs(2) system call.
func EncodeStatfs(buf *binary.Buffer, st *unix.Statfs_t) error {
	buf.PutUint32(st.Type)
	buf.PutUint32(st.Bsize)
	buf.PutUint64(st.Blocks)
	buf.PutUint64(st.Bfree)
	buf.PutUint64(st.Bavail)
	buf.PutUint64(st.Files)
	buf.PutUint64(st.Ffree)
	buf.PutUint64(uint64(st.Fsid.Val[0] | st.Fsid.Val[1]<<32))
	buf.PutUint32(255)

	return buf.Err()
}
