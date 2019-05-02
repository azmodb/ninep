package proto

import (
	"fmt"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// Rgetattr describes a file system object. The fields follow pretty
// closely the fields returned by the Linux stat(2) system call.
type Rgetattr struct {
	// Valid is a bitmask indicating which fields are valid in the
	// response.
	Valid uint64

	Qid // 13 byte value representing a unique file system object

	Mode      uint32 // inode protection mode
	Uid       uint32 // user-id of owner
	Gid       uint32 // group-id of owner
	Nlink     uint64 // number of hard links to the file
	Rdev      uint64 // device type, for special file inode
	Size      uint64 // file size, in bytes
	BlockSize uint64 // optimal file sys I/O ops blocksize
	Blocks    uint64

	Atime unix.Timespec // time of last access
	Mtime unix.Timespec // time of last data modification
	Ctime unix.Timespec // time of last file status change

	// Btime, Gen and DataVersion fields are reserved for future use.
	Btime       unix.Timespec
	Gen         uint64
	DataVersion uint64
}

var _ Message = (*Rgetattr)(nil) // Rgetattr implements Message interface

// Len returns the length of the message in bytes.
func (m Rgetattr) Len() int {
	return 8 + 13 + 4 + 4 + 4 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8
}

// Reset resets all state.
func (m *Rgetattr) Reset() { *m = Rgetattr{} }

// IsDir reports whether m describes a directory.
func (m Rgetattr) IsDir() bool {
	return m.Mode&unix.S_IFMT == unix.S_IFDIR
}

// Encode encodes to the given binary.Buffer.
func (m Rgetattr) Encode(buf *binary.Buffer) {
	buf.PutUint64(m.Valid)
	m.Qid.Encode(buf)
	buf.PutUint32(m.Mode)
	buf.PutUint32(m.Uid)
	buf.PutUint32(m.Gid)
	buf.PutUint64(m.Nlink)
	buf.PutUint64(m.Rdev)
	buf.PutUint64(m.Size)
	buf.PutUint64(m.BlockSize)
	buf.PutUint64(m.Blocks)
	encodeTimespec(buf, m.Atime)
	encodeTimespec(buf, m.Mtime)
	encodeTimespec(buf, m.Ctime)
	encodeTimespec(buf, m.Btime)
	buf.PutUint64(m.Gen)
	buf.PutUint64(m.DataVersion)
}

// Decode decodes from the given decoder.
func (m *Rgetattr) Decode(buf *binary.Buffer) {
	m.Valid = buf.Uint64()
	m.Qid.Decode(buf)
	m.Mode = buf.Uint32()
	m.Uid = buf.Uint32()
	m.Gid = buf.Uint32()
	m.Nlink = buf.Uint64()
	m.Rdev = buf.Uint64()
	m.Size = buf.Uint64()
	m.BlockSize = buf.Uint64()
	m.Blocks = buf.Uint64()
	m.Atime = decodeTimespec(buf)
	m.Mtime = decodeTimespec(buf)
	m.Ctime = decodeTimespec(buf)
	m.Btime = decodeTimespec(buf)
	m.Gen = buf.Uint64()
	m.DataVersion = buf.Uint64()
}

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

func (m Rgetattr) String() string {
	return fmt.Sprintf("%s mode:%d uid:%d gid:%d nlink:%d rdev:%d size:%d block_size:%d blocks:%d atime:%d mtime:%d ctime:%d",
		m.Qid, m.Mode, m.Uid, m.Gid, m.Nlink, m.Rdev, m.Size, m.BlockSize, m.Blocks,
		m.Atime.Nano(), m.Mtime.Nano(), m.Ctime.Nano(),
	)
}
