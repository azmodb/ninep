package proto

import (
	"fmt"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// EncodeRgetattr encodes information about a file system object. Valid
// is a bitmask indicating which fields are valid in the response.
func EncodeRgetattr(buf *binary.Buffer, valid uint64, st *unix.Stat_t) error {
	buf.PutUint64(valid)

	buf.PutUint8(UnixFileTypeToQidType(uint32(st.Mode))) // marshal Qid
	buf.PutUint32(uint32(st.Mtim.Nano() ^ st.Size<<8))
	buf.PutUint64(st.Ino)

	mode := NewModeUnix(uint32(st.Mode))
	buf.PutUint32(uint32(mode))
	buf.PutUint32(st.Uid)
	buf.PutUint32(st.Gid)
	buf.PutUint64(uint64(st.Nlink))
	buf.PutUint64(uint64(st.Rdev))
	buf.PutUint64(uint64(st.Size))
	buf.PutUint64(uint64(st.Blksize))
	buf.PutUint64(uint64(st.Blocks))
	encodeTimespec(buf, st.Atim)
	encodeTimespec(buf, st.Mtim)
	encodeTimespec(buf, st.Ctim)

	encodeTimespec(buf, st.Btim)
	buf.PutUint64(uint64(st.Gen))
	buf.PutUint64(0)

	return buf.Err()
}

// DecodeRgetattr decodes information about a file system object. Valid
// is a bitmask indicating which fields are valid in the response.
func DecodeRgetattr(buf *binary.Buffer, valid *uint64, st *unix.Stat_t) error {
	*valid = buf.Uint64()

	_ = buf.Uint8()  // qid directory type
	_ = buf.Uint32() // qid version
	st.Ino = buf.Uint64()

	mode := NewUnixMode(Mode(buf.Uint32()))
	st.Mode = uint16(mode)
	st.Uid = buf.Uint32()
	st.Gid = buf.Uint32()
	st.Nlink = uint16(buf.Uint64())
	st.Rdev = int32(buf.Uint64())
	st.Size = int64(buf.Uint64())
	st.Blksize = int32(buf.Uint64())
	st.Blocks = int64(buf.Uint64())
	st.Atim = decodeTimespec(buf)
	st.Mtim = decodeTimespec(buf)
	st.Ctim = decodeTimespec(buf)

	st.Btim = decodeTimespec(buf)
	st.Gen = uint32(buf.Uint64())
	_ = buf.Uint64() // DataVersion

	return buf.Err()
}

func (m Rgetattr) String() string {
	return fmt.Sprintf("valid:%d path:%d mode:%d uid:%d gid:%d nlink:%d rdev:%d size:%d block_size:%d blocks:%d atime:%d mtime:%d ctime:%d btime:%d gen:%d data_version:%d",
		m.Valid,
		m.Ino,
		m.Mode,
		m.Uid,
		m.Gid,
		m.Nlink,
		m.Rdev,
		m.Size,
		m.Blksize,
		m.Blocks,
		m.Atim.Nano(),
		m.Mtim.Nano(),
		m.Ctim.Nano(),

		m.Btim.Nano(),
		m.Gen,
		0,
	)
}

// EncodeRstatfs encodes information about a file system. The byte
// sequence follow pretty closely the fields returned by the Linux
// statfs(2) system call.
func EncodeRstatfs(buf *binary.Buffer, st *unix.Statfs_t) error {
	buf.PutUint32(st.Type)
	buf.PutUint32(st.Bsize)
	buf.PutUint64(st.Blocks)
	buf.PutUint64(st.Bfree)
	buf.PutUint64(st.Bavail)
	buf.PutUint64(st.Files)
	buf.PutUint64(st.Ffree)
	buf.PutUint64(uint64(st.Fsid.Val[0]) | uint64(st.Fsid.Val[1])<<32)
	buf.PutUint32(255)

	return buf.Err()
}

// DecodeRstatfs decodes information about a file system.
func DecodeRstatfs(buf *binary.Buffer, st *unix.Statfs_t) error {
	st.Type = buf.Uint32()
	st.Bsize = buf.Uint32()
	st.Blocks = buf.Uint64()
	st.Bfree = buf.Uint64()
	st.Bavail = buf.Uint64()
	st.Files = buf.Uint64()
	st.Ffree = buf.Uint64()
	v := buf.Uint64()
	st.Fsid.Val[0] = int32(v)
	st.Fsid.Val[1] = int32(v >> 32)
	_ = buf.Uint32()

	return buf.Err()
}

func (m *Rstatfs) String() string {
	return fmt.Sprintf("type:%d bsize:%d blocks:%d bfree:%d bavail:%d files:%d ffree:%d fsid:%.16x namelen:%d",
		m.Statfs_t.Type,
		m.Statfs_t.Bsize,
		m.Statfs_t.Blocks,
		m.Statfs_t.Bfree,
		m.Statfs_t.Bavail,
		m.Statfs_t.Files,
		m.Statfs_t.Ffree,
		uint64(m.Statfs_t.Fsid.Val[0])|uint64(m.Statfs_t.Fsid.Val[1])<<32,
		255,
	)
}

// StatToQid converts an unix.Stat_t.
func StatToQid(st *unix.Stat_t) Qid {
	return Qid{
		Type:    UnixFileTypeToQidType(uint32(st.Mode)),
		Version: uint32(st.Mtim.Nano() ^ st.Size<<8),
		Path:    st.Ino,
	}
}
