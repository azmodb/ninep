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

	// Uid is the user numerical identifiery of the attaching user.
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

	// Uid is the user numerical identifiery of the attaching user.
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
	Flags Flag
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
	Fid   uint32
	Name  string
	Flags Flag
	Perm  Mode
	Gid   uint32
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
	Perm         Mode
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
	Fid uint32

	// RequestMask is a bitmask indicating which fields are requested.
	RequestMask uint64
}

// Represents getattr mask values.
const (
	GetAttrMode  = 0x00000001 // Linux chmod(2) mode bits
	GetAttrNlink = 0x00000002

	// New owner, group of the file as described in Linux chown(2).
	GetAttrUid = 0x00000004
	GetAttrGid = 0x00000008

	GetAttrRdev  = 0x00000010
	GetAttrAtime = 0x00000020 // Time of last file access.
	GetAttrMtime = 0x00000040 // Time of last file modification.
	GetAttrCtime = 0x00000080
	GetAttrIno   = 0x00000100

	// New file size as handled by Linux truncate(2).
	GetAttrSize        = 0x00000200
	GetAttrBlocks      = 0x00000400
	GetAttrBtime       = 0x00000800
	GetAttrGen         = 0x00001000
	GetAttrDataVersion = 0x00002000

	GetAttrBasic = 0x000007ff
	GetAttrAll   = 0x00003fff
)

// Rgetattr describes a file system object. The fields follow pretty
// closely the fields returned by the Linux stat(2) system call.
type Rgetattr struct {
	// Valid is a bitmask indicating which fields are valid in the
	// response.
	Valid uint64

	Qid // 13 byte value representing a unique file system object

	Mode      Mode   // inode protection mode
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

	Mode Mode   // inode protection mode
	Uid  uint32 // user-id of owner
	Gid  uint32 // group-id of owner
	Size uint64 // file size as handled by Linux truncate(2)

	Atime unix.Timespec // time of last access
	Mtime unix.Timespec // time of last data modification
}

// Represents setattr mask values.
const (
	SetAttrMode     = 0x00000001
	SetAttrUid      = 0x00000002
	SetAttrGid      = 0x00000004
	SetAttrSize     = 0x00000008
	SetAttrAtime    = 0x00000010
	SetAttrMtime    = 0x00000020
	SetAttrCtime    = 0x00000040
	SetAttrAtimeSet = 0x00000080
	SetAttrMtimeSet = 0x00000100
)

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

// Tfsync flushes any cached data to disk.
type Tfsync struct {
	Fid uint32
}

// Rfsync message contains a server's reply to a Tfsync message.
type Rfsync struct{}

// Tlock  acquires or releases a POSIX record lock and has semantics
// similar to Linux fcntl(F_SETLK).
type Tlock struct {
	Fid      uint32
	Type     uint8
	Flags    uint32
	Start    uint64
	Length   uint64
	ProcID   uint32
	ClientID string
}

// Rlock message contains a server's reply to a Tlock message.
type Rlock struct {
	Status uint8
}

// Tgetlock tests for the existence of a POSIX record lock and has
// semantics similar to Linux fcntl(F_GETLK).
type Tgetlock struct {
	Fid      uint32
	Type     uint8
	Start    uint64
	Length   uint64
	ProcID   uint32
	ClientID string
}

// Rgetlock message contains a server's reply to a Tgetlock message.
type Rgetlock struct {
	Type     uint8
	Start    uint64
	Length   uint64
	ProcID   uint32
	ClientID string
}

// Tlink creates a hard link name in directory dfid. The link target is
// referenced by fid.
type Tlink struct {
	DirectoryFid uint32
	Target       uint32
	Name         string
}

// Rlink message contains a server's reply to a Tlink message.
type Rlink struct{}

// Tmkdir creates a new directory name in parent directory fid. Mode
// contains Linux mkdir(2) mode bits. Gid is the effective group ID of
// the caller.
type Tmkdir struct {
	DirectoryFid uint32
	Name         string
	Perm         Mode
	Gid          uint32
}

// Rmkdir message contains a server's reply to a Tmkdir message.
type Rmkdir = Qid

// Trenameat changes the name of a file from oldname to newname, possible
// moving it from old directory represented by fid to new directory
// represented by new fid.
//
// If the server returns unix.ENOTSUPP, the client should fall back to
// the rename operation.
type Trenameat struct {
	OldDirectoryFid uint32
	OldName         string
	NewDirectoryFid uint32
	NewName         string
}

// Rrenameat message contains a server's reply to a Trenameat message.
type Rrenameat struct{}

// Tunlinkat unlinks name from directory represented by directory fid. If
// the file is represented by a fid, that fid is not clunked.
//
// If the server returns unix.ENOTSUPP, the client should fall back to
// the remove operation.
type Tunlinkat struct {
	DirectoryFid uint32
	Name         string
	Flags        Flag
}

// Runlinkat message contains a server's reply to a Tunlinkat message.
type Runlinkat struct{}
