package proto

import (
	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// EncodeStat encodes information about a file system object. The byte
// sequencei follow pretty closely the fields returned by the Linux
// stat(2) system call.
//
// Valid is a bitmask indicating which fields are valid in the response.
func EncodeStat(buf *binary.Buffer, valid uint64, st *unix.Stat_t) error {
	buf.PutUint64(valid)

	buf.PutUint8(FileTypeToQidType(uint32(st.Mode))) // marshal Qid
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

func StatToRgetattr(st *unix.Stat_t) *Rgetattr {
	return &Rgetattr{
		Qid: Qid{
			FileTypeToQidType(uint32(st.Mode)),
			uint32(st.Mtimespec.Nano() ^ st.Size<<8),
			st.Ino,
		},
		Mode:        uint32(st.Mode),
		Uid:         st.Uid,
		Gid:         st.Gid,
		Nlink:       uint64(st.Nlink),
		Rdev:        uint64(st.Rdev),
		Size:        uint64(st.Size),
		BlockSize:   uint64(st.Blksize),
		Blocks:      uint64(st.Blocks),
		Atime:       st.Atimespec,
		Mtime:       st.Mtimespec,
		Ctime:       st.Ctimespec,
		Btime:       st.Birthtimespec,
		Gen:         uint64(st.Gen),
		DataVersion: 0,
	}
}

func EncodeStatfs(buf *binary.Buffer, st *unix.Statfs_t) error {
	return buf.Err()
}
