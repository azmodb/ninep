package proto

import (
	"fmt"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// Tlauth message is used to authenticate users on a connection. If the
// server does require authentication, it returns aqid defining a file
// of type QTAUTH that may be read and written to execute an
// authentication protocol. That protocol's definition is not part of
// 9P itself.
//
// Once the protocol is complete, the same afid is presented in the
// attach message for the user, granting entry. The same validated afid
// may be used for multiple attach messages with the same uname and
// aname.
type Tlauth struct {
	// AuthFid serves to authenticate a user, and must have been
	// established in a previous Tlauth request.
	AuthFid uint32

	// UserName is the user name of the attaching user.
	UserName string

	// Path is the name of the file tree that the client wants to
	// access. It may be empty.
	Path string

	// Uid is the user numercial identifiery of the attaching user.
	Uid uint32
}

// Rlauth reply is sent in response to a Tlauth request. If a server does
// not require authentication, it can reply to a Tlauth message with an
// Rerror message.
type Rlauth = Qid

// Tlattach message serves as a fresh introduction from a user on the
// client machine to the server.
type Tlattach struct {
	// Fid establishes a fid to be used as the root of the file tree,
	// should the client's Tlattach request be accepted.
	Fid uint32

	// AuthFid serves to authenticate a user, and must have been
	// established in a previous Tlauth request.
	AuthFid uint32

	// UserName is the user name of the attaching user.
	UserName string

	// Path is the name of the file tree that the client wants to
	// access. It may be empty.
	Path string

	// Uid is the user numercial identifiery of the attaching user.
	Uid uint32
}

// Rlattach message contains a server's reply to a Tlattach request. As a
// result of the attach transaction, the client will have a connection to
// the root directory of the desired file tree, represented by the
// returned qid.
type Rlattach = Qid

// Rlerror message (there is no Tlerror) is used to return an errno
// number describing the failure of a transaction.
type Rlerror struct {
	Errno uint32
}

// Tstatfs is used to request file system information of the file system
// containing fid.
type Tstatfs struct {
	Fid uint32
}

// Rstatfs response corresponds to the fields returned by the statfs(2)
// system call.
type Rstatfs struct {
	Type            uint32
	BlockSize       uint32
	Blocks          uint64
	BlocksFree      uint64
	BlocksAvailable uint64
	Files           uint64
	FilesFree       uint64
	FsID            uint64
	NameLength      uint32
}

// Tlopen prepares fid for file I/O. Flags contains Linux open(2) flags
// bits, e.g. O_RDONLY, O_RDWR, O_WRONLY.
type Tlopen struct {
	Fid   uint32
	Flags uint32
}

// Rlopen message contains a server's reply to a Tlopen request.
type Rlopen struct {
	Qid
	Iounit uint32
}

// Tlcreate creates a regular file name in directory fid and prepares it
// for I/O. Fid initially represents the parent directory of the new
// file. After the call it represents the new file. Mode contains Linux
// creat(2) mode bits. Flags contains Linux open(2) flags bits, e.g.
// O_RDONLY, O_RDWR, O_WRONLY.
type Tlcreate struct {
	Fid        uint32
	Name       string
	Flags      uint32
	Permission uint32
	Gid        uint32
}

// Rlcreate message contains a server's reply to a Tlcreate request. The
// qid for the new file is returned in the reply.
type Rlcreate struct {
	Qid
	Iounit uint32
}

// Tsymlink creates a symbolic link name in directory fid. The link will
// point to target.
type Tsymlink struct {
	DirectoryFid uint32
	Name         string
	Target       string
	Gid          uint32
}

// Rsymlink message contains a server's reply to a Tsymlink request. The
// qid for the new symbolic link is returned in the reply.
type Rsymlink = Qid

// Tmknod creates a device node name in directory fid with major and
// minor numbers. Mode contains Linux mknod(2) mode bits.
type Tmknod struct {
	DirectoryFid uint32
	Name         string
	Permission   uint32
	Major        uint32
	Minor        uint32
	Gid          uint32
}

// Rmknod message contains a server's reply to a Tmknod request. The qid
// for the new device node is returned in the reply.
type Rmknod = Qid

// Trename renames a file system object referenced by fid, to name in
// the directory referenced by directory fid.
type Trename struct {
	Fid          uint32
	DirectoryFid uint32
	Name         string
}

// Rrename message contains a server's reply to a Trename request.
type Rrename struct{}

// Treadlink returns the contents of the symbolic link referenced by
// fid.
type Treadlink struct {
	Fid uint32
}

// Rreadlink message contains a server's reply to a Treadlink request.
type Rreadlink struct {
	Target string
}

// Tgetattr gets attributes of a file system object referenced by fid.
type Tgetattr struct {
	Fid      uint32
	AttrMask uint64
}

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

// Tsetattr sets attributes of a file system object referenced by fid.
type Tsetattr struct {
	Fid uint32

	// Valid is a bitmask selecting which fields to set.
	Valid uint32

	Mode uint32 // inode protection mode
	Uid  uint32 // user-id of owner
	Gid  uint32 // group-id of owner
	Size uint64 // file size as handled by Linux truncate(2)

	Atime unix.Timespec // time of last access
	Mtime unix.Timespec // time of last data modification
}

// Rsetattr message contains a server's reply to a Tsetattr request.
type Rsetattr struct{}

// Txattrwalk gets a new fid pointing to xattr name. This fid can later
// be used to read the xattr value. If name is NULL new fid can be used
// to get the list of extended attributes associated with the file
// system object.
type Txattrwalk struct {
	Fid    uint32
	NewFid uint32
	Name   string
}

// Rxattrwalk message contains a server's reply to a Txattrwalk request.
type Rxattrwalk struct {
	Size uint64
}

// Txattrcreate gets a fid pointing to the xattr name. This fid can
// later be used to set the xattr value. Flag is derived from set Linux
// setxattr(2). The manpage says
//
// 		The flags parameter can be used to refine the semantics of the
// 		operation. XATTR_CREATE specifies a pure create, which fails if
// 		the named attribute exists already. XATTR_REPLACE specifies a
// 		pure replace operation, which fails if the named attribute does
// 		not already exist. By default (no flags), the extended attribute
// 		will be created if need be, or will simply replace the value if
// 		the attribute exists.
//
// The actual setxattr operation happens when the fid is clunked. At
// that point the written byte count and the attrsize specified in
// Txattrcreate should be same otherwise an error will be returned.
type Txattrcreate struct {
	Fid      uint32
	Name     string
	AttrSize uint64
	Flag     uint32
}

// Rxattrcreate message contains a server's reply to a Txattrcreate
// request.
type Rxattrcreate struct{}

// Treaddir requests that the server return directory entries from the
// directory represented by fid, previously opened with Tlopen. Offset
// is zero on the first call.
//
// Directory entries are represented as variable-length records:
//
// 		qid[13] offset[8] type[1] name[s]
//
// At most count bytes will be returned in data. If count is not zero in
// the response, more data is available. On subsequent calls, offset is
// the offset returned in the last directory entry of the previous call.
type Treaddir struct {
	Fid    uint32
	Offset uint64
	Count  uint32
}

// Rreaddir message contains a server's reply to a Treaddir message.
type Rreaddir struct {
	Data []byte
}

// String implements fmt.Stringer.
func (m Rreaddir) String() string {
	return fmt.Sprintf("data_len:%d", len(m.Data))
}

// Len returns the length of the message in bytes.
func (m Rreaddir) Len() int { return 4 + len(m.Data) }

// Reset resets all state.
func (m *Rreaddir) Reset() { *m = Rreaddir{} }

// Encode encodes to the given binary.Buffer.
func (m Rreaddir) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rreaddir) Decode(buf *binary.Buffer) {}

// Payload returns the payload for sending.
func (m Rreaddir) Payload() []byte { return m.Data }

// PutPayload sets the decoded payload.
func (m *Rreaddir) PutPayload(b []byte) { m.Data = b }

// FixedLen returns the fixed message size in bytes.
func (m Rreaddir) FixedLen() int { return 0 }
