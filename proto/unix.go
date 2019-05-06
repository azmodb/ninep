package proto

import (
	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

func encodeTimespec(buf *binary.Buffer, t unix.Timespec) {
	sec, nsec := t.Unix()
	buf.PutUint64(uint64(sec))
	buf.PutUint64(uint64(nsec))
}

func decodeTimespec(buf *binary.Buffer) unix.Timespec {
	return unix.Timespec{
		Sec:  int64(buf.Uint64()),
		Nsec: int64(buf.Uint64()),
	}
}

// UnixDirTypeToQidType converts an unix directory type.
func UnixDirTypeToQidType(typ uint8) uint8 {
	switch typ {
	//case unix.DT_BLK: ???
	case unix.DT_CHR, unix.DT_FIFO, unix.DT_SOCK:
		return TypeAppendOnly
	case unix.DT_DIR:
		return TypeDirectory
	case unix.DT_LNK:
		return TypeSymlink
	}
	return TypeRegular
}

// UnixFileTypeToQidType converts an unix file type.
func UnixFileTypeToQidType(mode uint32) uint8 {
	switch mode & unix.S_IFMT {
	//case unix.S_IFBLK: ???
	case unix.S_IFCHR, unix.S_IFIFO, unix.S_IFSOCK:
		return TypeAppendOnly
	case unix.S_IFDIR:
		return TypeDirectory
	case unix.S_IFLNK:
		return TypeSymlink
	}
	return TypeRegular
}
