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

func DirectoryTypeToQidType(typ uint8) uint8 {
	switch typ {
	//case unix.DT_BLK: ???
	case unix.DT_CHR, unix.DT_FIFO, unix.DT_SOCK:
		return typeAppendOnly
	case unix.DT_DIR:
		return typeDirectory
	case unix.DT_LNK:
		return typeSymlink
	}
	return typeRegular
}

func FileTypeToQidType(mode uint32) uint8 {
	switch mode & unix.S_IFMT {
	//case unix.S_IFBLK: ???
	case unix.S_IFCHR, unix.S_IFIFO, unix.S_IFSOCK:
		return typeAppendOnly
	case unix.S_IFDIR:
		return typeDirectory
	case unix.S_IFLNK:
		return typeSymlink
	}
	return typeRegular
}
