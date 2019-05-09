package proto

import (
	"fmt"
	"io"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// Dirent structure defines the format of 9P2000.L directory entries.
type Dirent struct {
	Qid
	Offset uint64
	Type   uint8
	Name   string
}

// String implements fmt.Stringer.
func (m Dirent) String() string {
	return fmt.Sprintf("%s offset:%d type:%d name:%q",
		m.Qid, m.Offset, m.Type, m.Name)
}

// Len returns the length of the message in bytes.
func (m Dirent) Len() int { return 13 + 8 + 1 + 2 + len(m.Name) }

// Reset resets all state.
func (m *Dirent) Reset() { *m = Dirent{} }

// Encode encodes to the given binary.Buffer.
func (m Dirent) Encode(buf *binary.Buffer) {
	m.Qid.Encode(buf)
	buf.PutUint64(m.Offset)
	buf.PutUint8(m.Type)
	buf.PutString(m.Name)
}

// Decode decodes from the given binary.Buffer.
func (m *Dirent) Decode(buf *binary.Buffer) {
	m.Qid.Decode(buf)
	m.Offset = buf.Uint64()
	m.Type = buf.Uint8()
	m.Name = buf.String()
}

// Marshal implements binary.Marshaler.
func (m Dirent) Marshal(data []byte) ([]byte, int, error) {
	data = binary.PutUint8(data, m.Qid.Type)
	data = binary.PutUint32(data, m.Qid.Version)
	data = binary.PutUint64(data, m.Qid.Path)
	data = binary.PutUint64(data, m.Offset)
	data = binary.PutUint8(data, m.Type)
	data = binary.PutString(data, m.Name)
	return data, m.Len(), nil
}

// Unmarshal implements binary.Unmarshaler.
func (m *Dirent) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 24 {
		return data, io.ErrUnexpectedEOF
	}
	m.Qid = Qid{
		Type:    binary.Uint8(data[:1]),
		Version: binary.Uint32(data[1:5]),
		Path:    binary.Uint64(data[5:13]),
	}
	m.Offset = binary.Uint64(data[13:21])
	m.Type = binary.Uint8(data[21:22])
	m.Name = binary.String(data[22:])
	return data, nil
}

// Stat describes a file system object.
type Stat = unix.Stat_t

// StatFS describes a file system.
type StatFS = unix.Statfs_t

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
