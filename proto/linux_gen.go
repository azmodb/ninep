package proto

import (
	"fmt"

	"github.com/azmodb/ninep/binary"
)

// String implements fmt.Stringer.
func (m Tlauth) String() string {
	return fmt.Sprintf("auth_fid:%d user_name:%q path:%q uid:%d", m.AuthFid, m.UserName, m.Path, m.Uid)
}

// Len returns the length of the message in bytes.
func (m Tlauth) Len() int { return 4 + 2 + len(m.UserName) + 2 + len(m.Path) + 4 }

// Reset resets all state.
func (m *Tlauth) Reset() { *m = Tlauth{} }

// Encode encodes to the given binary.Buffer.
func (m Tlauth) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.AuthFid)
	buf.PutString(m.UserName)
	buf.PutString(m.Path)
	buf.PutUint32(m.Uid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tlauth) Decode(buf *binary.Buffer) {
	m.AuthFid = buf.Uint32()
	m.UserName = buf.String()
	m.Path = buf.String()
	m.Uid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tlattach) String() string {
	return fmt.Sprintf("fid:%d auth_fid:%d user_name:%q path:%q uid:%d", m.Fid, m.AuthFid, m.UserName, m.Path, m.Uid)
}

// Len returns the length of the message in bytes.
func (m Tlattach) Len() int { return 4 + 4 + 2 + len(m.UserName) + 2 + len(m.Path) + 4 }

// Reset resets all state.
func (m *Tlattach) Reset() { *m = Tlattach{} }

// Encode encodes to the given binary.Buffer.
func (m Tlattach) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint32(m.AuthFid)
	buf.PutString(m.UserName)
	buf.PutString(m.Path)
	buf.PutUint32(m.Uid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tlattach) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.AuthFid = buf.Uint32()
	m.UserName = buf.String()
	m.Path = buf.String()
	m.Uid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rlerror) String() string { return fmt.Sprintf("errno:%d", m.Errno) }

// Len returns the length of the message in bytes.
func (m Rlerror) Len() int { return 4 }

// Reset resets all state.
func (m *Rlerror) Reset() { *m = Rlerror{} }

// Encode encodes to the given binary.Buffer.
func (m Rlerror) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Errno)
}

// Decode decodes from the given binary.Buffer.
func (m *Rlerror) Decode(buf *binary.Buffer) {
	m.Errno = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tstatfs) String() string { return fmt.Sprintf("fid:%d", m.Fid) }

// Len returns the length of the message in bytes.
func (m Tstatfs) Len() int { return 4 }

// Reset resets all state.
func (m *Tstatfs) Reset() { *m = Tstatfs{} }

// Encode encodes to the given binary.Buffer.
func (m Tstatfs) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tstatfs) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rstatfs) String() string {
	return fmt.Sprintf("type:%d block_size:%d blocks:%d blocks_free:%d blocks_available:%d files:%d files_free:%d fs_id:%d name_length:%d", m.Type, m.BlockSize, m.Blocks, m.BlocksFree, m.BlocksAvailable, m.Files, m.FilesFree, m.FsID, m.NameLength)
}

// Len returns the length of the message in bytes.
func (m Rstatfs) Len() int { return 4 + 4 + 8 + 8 + 8 + 8 + 8 + 8 + 4 }

// Reset resets all state.
func (m *Rstatfs) Reset() { *m = Rstatfs{} }

// Encode encodes to the given binary.Buffer.
func (m Rstatfs) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Type)
	buf.PutUint32(m.BlockSize)
	buf.PutUint64(m.Blocks)
	buf.PutUint64(m.BlocksFree)
	buf.PutUint64(m.BlocksAvailable)
	buf.PutUint64(m.Files)
	buf.PutUint64(m.FilesFree)
	buf.PutUint64(m.FsID)
	buf.PutUint32(m.NameLength)
}

// Decode decodes from the given binary.Buffer.
func (m *Rstatfs) Decode(buf *binary.Buffer) {
	m.Type = buf.Uint32()
	m.BlockSize = buf.Uint32()
	m.Blocks = buf.Uint64()
	m.BlocksFree = buf.Uint64()
	m.BlocksAvailable = buf.Uint64()
	m.Files = buf.Uint64()
	m.FilesFree = buf.Uint64()
	m.FsID = buf.Uint64()
	m.NameLength = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tlopen) String() string { return fmt.Sprintf("fid:%d flags:%d", m.Fid, m.Flags) }

// Len returns the length of the message in bytes.
func (m Tlopen) Len() int { return 4 + 4 }

// Reset resets all state.
func (m *Tlopen) Reset() { *m = Tlopen{} }

// Encode encodes to the given binary.Buffer.
func (m Tlopen) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint32(m.Flags)
}

// Decode decodes from the given binary.Buffer.
func (m *Tlopen) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.Flags = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rlopen) String() string { return fmt.Sprintf("%s iounit:%d", m.Qid, m.Iounit) }

// Len returns the length of the message in bytes.
func (m Rlopen) Len() int { return 13 + 4 }

// Reset resets all state.
func (m *Rlopen) Reset() { *m = Rlopen{} }

// Encode encodes to the given binary.Buffer.
func (m Rlopen) Encode(buf *binary.Buffer) {
	m.Qid.Encode(buf)
	buf.PutUint32(m.Iounit)
}

// Decode decodes from the given binary.Buffer.
func (m *Rlopen) Decode(buf *binary.Buffer) {
	m.Qid.Decode(buf)
	m.Iounit = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tlcreate) String() string {
	return fmt.Sprintf("fid:%d name:%q flags:%d permission:%d gid:%d", m.Fid, m.Name, m.Flags, m.Permission, m.Gid)
}

// Len returns the length of the message in bytes.
func (m Tlcreate) Len() int { return 4 + 2 + len(m.Name) + 4 + 4 + 4 }

// Reset resets all state.
func (m *Tlcreate) Reset() { *m = Tlcreate{} }

// Encode encodes to the given binary.Buffer.
func (m Tlcreate) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutString(m.Name)
	buf.PutUint32(m.Flags)
	buf.PutUint32(m.Permission)
	buf.PutUint32(m.Gid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tlcreate) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.Name = buf.String()
	m.Flags = buf.Uint32()
	m.Permission = buf.Uint32()
	m.Gid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rlcreate) String() string { return fmt.Sprintf("%s iounit:%d", m.Qid, m.Iounit) }

// Len returns the length of the message in bytes.
func (m Rlcreate) Len() int { return 13 + 4 }

// Reset resets all state.
func (m *Rlcreate) Reset() { *m = Rlcreate{} }

// Encode encodes to the given binary.Buffer.
func (m Rlcreate) Encode(buf *binary.Buffer) {
	m.Qid.Encode(buf)
	buf.PutUint32(m.Iounit)
}

// Decode decodes from the given binary.Buffer.
func (m *Rlcreate) Decode(buf *binary.Buffer) {
	m.Qid.Decode(buf)
	m.Iounit = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tsymlink) String() string {
	return fmt.Sprintf("directory_fid:%d name:%q target:%q gid:%d", m.DirectoryFid, m.Name, m.Target, m.Gid)
}

// Len returns the length of the message in bytes.
func (m Tsymlink) Len() int { return 4 + 2 + len(m.Name) + 2 + len(m.Target) + 4 }

// Reset resets all state.
func (m *Tsymlink) Reset() { *m = Tsymlink{} }

// Encode encodes to the given binary.Buffer.
func (m Tsymlink) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.DirectoryFid)
	buf.PutString(m.Name)
	buf.PutString(m.Target)
	buf.PutUint32(m.Gid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tsymlink) Decode(buf *binary.Buffer) {
	m.DirectoryFid = buf.Uint32()
	m.Name = buf.String()
	m.Target = buf.String()
	m.Gid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tmknod) String() string {
	return fmt.Sprintf("directory_fid:%d name:%q permission:%d major:%d minor:%d gid:%d", m.DirectoryFid, m.Name, m.Permission, m.Major, m.Minor, m.Gid)
}

// Len returns the length of the message in bytes.
func (m Tmknod) Len() int { return 4 + 2 + len(m.Name) + 4 + 4 + 4 + 4 }

// Reset resets all state.
func (m *Tmknod) Reset() { *m = Tmknod{} }

// Encode encodes to the given binary.Buffer.
func (m Tmknod) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.DirectoryFid)
	buf.PutString(m.Name)
	buf.PutUint32(m.Permission)
	buf.PutUint32(m.Major)
	buf.PutUint32(m.Minor)
	buf.PutUint32(m.Gid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tmknod) Decode(buf *binary.Buffer) {
	m.DirectoryFid = buf.Uint32()
	m.Name = buf.String()
	m.Permission = buf.Uint32()
	m.Major = buf.Uint32()
	m.Minor = buf.Uint32()
	m.Gid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Trename) String() string {
	return fmt.Sprintf("fid:%d directory_fid:%d name:%q", m.Fid, m.DirectoryFid, m.Name)
}

// Len returns the length of the message in bytes.
func (m Trename) Len() int { return 4 + 4 + 2 + len(m.Name) }

// Reset resets all state.
func (m *Trename) Reset() { *m = Trename{} }

// Encode encodes to the given binary.Buffer.
func (m Trename) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint32(m.DirectoryFid)
	buf.PutString(m.Name)
}

// Decode decodes from the given binary.Buffer.
func (m *Trename) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.DirectoryFid = buf.Uint32()
	m.Name = buf.String()
}

// String implements fmt.Stringer.
func (m Rrename) String() string { return "" }

// Len returns the length of the message in bytes.
func (m Rrename) Len() int { return 0 }

// Reset resets all state.
func (m *Rrename) Reset() { *m = Rrename{} }

// Encode encodes to the given binary.Buffer.
func (m Rrename) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rrename) Decode(buf *binary.Buffer) {}

// String implements fmt.Stringer.
func (m Treadlink) String() string { return fmt.Sprintf("fid:%d", m.Fid) }

// Len returns the length of the message in bytes.
func (m Treadlink) Len() int { return 4 }

// Reset resets all state.
func (m *Treadlink) Reset() { *m = Treadlink{} }

// Encode encodes to the given binary.Buffer.
func (m Treadlink) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
}

// Decode decodes from the given binary.Buffer.
func (m *Treadlink) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rreadlink) String() string { return fmt.Sprintf("target:%q", m.Target) }

// Len returns the length of the message in bytes.
func (m Rreadlink) Len() int { return 2 + len(m.Target) }

// Reset resets all state.
func (m *Rreadlink) Reset() { *m = Rreadlink{} }

// Encode encodes to the given binary.Buffer.
func (m Rreadlink) Encode(buf *binary.Buffer) {
	buf.PutString(m.Target)
}

// Decode decodes from the given binary.Buffer.
func (m *Rreadlink) Decode(buf *binary.Buffer) {
	m.Target = buf.String()
}

// String implements fmt.Stringer.
func (m Tgetattr) String() string { return fmt.Sprintf("fid:%d attr_mask:%d", m.Fid, m.AttrMask) }

// Len returns the length of the message in bytes.
func (m Tgetattr) Len() int { return 4 + 8 }

// Reset resets all state.
func (m *Tgetattr) Reset() { *m = Tgetattr{} }

// Encode encodes to the given binary.Buffer.
func (m Tgetattr) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint64(m.AttrMask)
}

// Decode decodes from the given binary.Buffer.
func (m *Tgetattr) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.AttrMask = buf.Uint64()
}

// String implements fmt.Stringer.
func (m Rsetattr) String() string { return "" }

// Len returns the length of the message in bytes.
func (m Rsetattr) Len() int { return 0 }

// Reset resets all state.
func (m *Rsetattr) Reset() { *m = Rsetattr{} }

// Encode encodes to the given binary.Buffer.
func (m Rsetattr) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rsetattr) Decode(buf *binary.Buffer) {}

// String implements fmt.Stringer.
func (m Txattrwalk) String() string {
	return fmt.Sprintf("fid:%d new_fid:%d name:%q", m.Fid, m.NewFid, m.Name)
}

// Len returns the length of the message in bytes.
func (m Txattrwalk) Len() int { return 4 + 4 + 2 + len(m.Name) }

// Reset resets all state.
func (m *Txattrwalk) Reset() { *m = Txattrwalk{} }

// Encode encodes to the given binary.Buffer.
func (m Txattrwalk) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint32(m.NewFid)
	buf.PutString(m.Name)
}

// Decode decodes from the given binary.Buffer.
func (m *Txattrwalk) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.NewFid = buf.Uint32()
	m.Name = buf.String()
}

// String implements fmt.Stringer.
func (m Rxattrwalk) String() string { return fmt.Sprintf("size:%d", m.Size) }

// Len returns the length of the message in bytes.
func (m Rxattrwalk) Len() int { return 8 }

// Reset resets all state.
func (m *Rxattrwalk) Reset() { *m = Rxattrwalk{} }

// Encode encodes to the given binary.Buffer.
func (m Rxattrwalk) Encode(buf *binary.Buffer) {
	buf.PutUint64(m.Size)
}

// Decode decodes from the given binary.Buffer.
func (m *Rxattrwalk) Decode(buf *binary.Buffer) {
	m.Size = buf.Uint64()
}

// String implements fmt.Stringer.
func (m Txattrcreate) String() string {
	return fmt.Sprintf("fid:%d name:%q attr_size:%d flag:%d", m.Fid, m.Name, m.AttrSize, m.Flag)
}

// Len returns the length of the message in bytes.
func (m Txattrcreate) Len() int { return 4 + 2 + len(m.Name) + 8 + 4 }

// Reset resets all state.
func (m *Txattrcreate) Reset() { *m = Txattrcreate{} }

// Encode encodes to the given binary.Buffer.
func (m Txattrcreate) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutString(m.Name)
	buf.PutUint64(m.AttrSize)
	buf.PutUint32(m.Flag)
}

// Decode decodes from the given binary.Buffer.
func (m *Txattrcreate) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.Name = buf.String()
	m.AttrSize = buf.Uint64()
	m.Flag = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rxattrcreate) String() string { return "" }

// Len returns the length of the message in bytes.
func (m Rxattrcreate) Len() int { return 0 }

// Reset resets all state.
func (m *Rxattrcreate) Reset() { *m = Rxattrcreate{} }

// Encode encodes to the given binary.Buffer.
func (m Rxattrcreate) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rxattrcreate) Decode(buf *binary.Buffer) {}

// String implements fmt.Stringer.
func (m Treaddir) String() string {
	return fmt.Sprintf("fid:%d offset:%d count:%d", m.Fid, m.Offset, m.Count)
}

// Len returns the length of the message in bytes.
func (m Treaddir) Len() int { return 4 + 8 + 4 }

// Reset resets all state.
func (m *Treaddir) Reset() { *m = Treaddir{} }

// Encode encodes to the given binary.Buffer.
func (m Treaddir) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint64(m.Offset)
	buf.PutUint32(m.Count)
}

// Decode decodes from the given binary.Buffer.
func (m *Treaddir) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.Offset = buf.Uint64()
	m.Count = buf.Uint32()
}
