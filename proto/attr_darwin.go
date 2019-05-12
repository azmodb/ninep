package proto

import "github.com/azmodb/ninep/binary"

// EncodeStat encodes information about a file system object. The byte
// sequence follow pretty closely the fields returned by the Linux
// stat(2) system call.
//
// Valid is a bitmask indicating which fields are valid in the response.
func EncodeStat(buf *binary.Buffer, valid uint64, stat *Stat) error {
	buf.PutUint64(valid)

	buf.PutUint8(UnixFileTypeToQidType(uint32(stat.Mode))) // marshal Qid
	buf.PutUint32(uint32(stat.Mtimespec.Nano() ^ stat.Size<<8))
	buf.PutUint64(stat.Ino)

	buf.PutUint32(uint32(stat.Mode))
	buf.PutUint32(stat.Uid)
	buf.PutUint32(stat.Gid)
	buf.PutUint64(uint64(stat.Nlink))
	buf.PutUint64(uint64(stat.Rdev))
	buf.PutUint64(uint64(stat.Size))
	buf.PutUint64(uint64(stat.Blksize))
	buf.PutUint64(uint64(stat.Blocks))
	encodeTimespec(buf, stat.Atimespec)
	encodeTimespec(buf, stat.Mtimespec)
	encodeTimespec(buf, stat.Ctimespec)

	encodeTimespec(buf, stat.Birthtimespec)
	buf.PutUint64(uint64(stat.Gen))
	buf.PutUint64(0)

	return buf.Err()
}

// UnixStatToRgetattr converts a unix.Stat_t returned by the stat(2)
// system call.
func UnixStatToRgetattr(stat *Stat) *Rgetattr {
	return &Rgetattr{
		Qid: Qid{
			UnixFileTypeToQidType(uint32(stat.Mode)),
			uint32(stat.Mtimespec.Nano() ^ stat.Size<<8),
			stat.Ino,
		},

		Mode:      uint32(stat.Mode),
		Uid:       stat.Uid,
		Gid:       stat.Gid,
		Nlink:     uint64(stat.Nlink),
		Rdev:      uint64(stat.Rdev),
		Size:      uint64(stat.Size),
		BlockSize: uint64(stat.Blksize),
		Blocks:    uint64(stat.Blocks),
		Atime:     stat.Atimespec,
		Mtime:     stat.Mtimespec,
		Ctime:     stat.Ctimespec,

		Btime:       stat.Birthtimespec,
		Gen:         uint64(stat.Gen),
		DataVersion: 0,
	}
}

// EncodeStatFS encodes information about a file system. The byte
// sequence follow pretty closely the fields returned by the Linux
// statfs(2) system call.
func EncodeStatFS(buf *binary.Buffer, stat *StatFS) error {
	buf.PutUint32(stat.Type)
	buf.PutUint32(stat.Bsize)
	buf.PutUint64(stat.Blocks)
	buf.PutUint64(stat.Bfree)
	buf.PutUint64(stat.Bavail)
	buf.PutUint64(stat.Files)
	buf.PutUint64(stat.Ffree)
	buf.PutUint64(uint64(stat.Fsid.Val[0] | stat.Fsid.Val[1]<<32))
	buf.PutUint32(255)

	return buf.Err()
}
