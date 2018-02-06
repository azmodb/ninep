package proto

import (
	"fmt"
	"io"
)

// Defines some limits on how big some arbitrarily-long 9P2000 fields can
// be.
const (
	// maxVersionLen is the length (in bytes) of a version identiefier.
	maxVersionLen = 64

	// maxUnameLen is the maximum length (in bytes) of a username or
	// group identifier.
	maxUnameLen = 255

	// maxNameLen is the maximum length of a file name in bytes.
	maxNameLen = 255

	// maxName is the maximum allowed number of path elements in a
	// Twalk request.
	maxName = 16

	maxPathLen = maxName + maxName*maxNameLen // 4096 bytes

	// size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	// size[4] Tread  tag[2] fid[4] offset[8] count[4]
	fixedReadWriteLen = 4 + 1 + 2 + 4 + 8 + 4 // 23

	maxMessageLen = (fixedReadWriteLen + 1) + maxDataLen
	maxDataLen    = (1<<31 - 1) - (fixedReadWriteLen + 1) // ~ 2GB

	headerLen = 7 // size[4] type[1] tag[2]

	// To make the contents of a directory, such as returned by read(5),
	// easy to parse, each directory entry begins with a size field.
	// For consistency, the entries in Twstat and Rstat messages also
	// contain their size, which means the size appears twice.
	//
	// 	http://9p.io/magic/man2html/5/stat
	//
	fixedStatLen = 2 + 4 + 13 + 4 + 4 + 4 + 8 + 4*2 // 47
)

// NOFID is a reserved fid used in a Tattach request for the afid field,
// that indicates that the client does not wish to authenticate his
// session.
const NOFID = ^uint32(0)

// NOTAG is the tag for Tversion and Rversion requests.
const NOTAG = ^uint16(0)

// Flags for the mode field in Topen and Tcreate messages.
const (
	OREAD   = 0  // open read-only
	OWRITE  = 1  // open write-only
	ORDWR   = 2  // open read-write
	OEXEC   = 3  // read but check execute permission
	OTRUNC  = 16 // or'ed in (except for exec), truncate file first
	OCEXEC  = 32 // or'ed in, close on exec
	ORCLOSE = 64 // or'ed in, remove on close
)

// File and directory modes.
const (
	DMDIR    = 0x80000000 // mode bit for directories
	DMAPPEND = 0x40000000 // mode bit for append only files
	DMEXCL   = 0x20000000 // mode bit for exclusive use files
	DMMOUNT  = 0x10000000 // mode bit for mounted channel
	DMAUTH   = 0x08000000 // mode bit for authentication file
	DMTMP    = 0x04000000 // mode bit for non-backed-up file
	DMREAD   = 0x4        // mode bit for read permission
	DMWRITE  = 0x2        // mode bit for write permission
	DMEXEC   = 0x1        // mode bit for execute permission
)

// Qid type field represents the type of a file (directory, etc.),
// represented as a bit vector corresponding to the high 8 bits of the
// file mode word.
const (
	QTDIR    = 0x80 // directories
	QTAPPEND = 0x40 // append only files
	QTEXCL   = 0x20 // exclusive use files
	QTMOUNT  = 0x10 // mounted channel
	QTAUTH   = 0x08 // authentication file (afid)
	QTTMP    = 0x04 // non-backed-up file
	QTFILE   = 0x00
)

// Message represents a 9P2000 message and is used to access fields common
// to all 9P2000 messages.
type Message interface {
	// Len returns the four byte size field specifying the length in bytes
	// of the complete message including the four bytes of the size field
	// itself.
	Len() int64

	// Tag returns the field choosen by the client to identify the
	// message.
	Tag() uint16

	Reset()

	String() string
}

// Qid represents the server's unique identification for the file being
// accessed. Two files on the same server hierarchy are the same if and
// only if their qids are the same.
type Qid struct {
	Type uint8 // type of a file (directory, etc)

	// Version is a version number for a file. It is incremented every
	// time a file is modified. By convention, synthetic files usually
	// have a verison number of 0. Traditional files have a version
	// number that is a hash of their modification time.
	Version uint32

	// Path is an integer unique among all files in the hierarchy. If
	// a file is deleted and recreated with the same name in the same
	// directory, the old and new path components of the qids should
	// be different.
	Path uint64
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (q Qid) MarshalBinary() ([]byte, error) {
	data := make([]byte, 13)
	data[0] = q.Type
	puint32(data[1:5], q.Version)
	puint64(data[5:13], q.Path)
	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (q *Qid) UnmarshalBinary(data []byte) error {
	if len(data) < 13 {
		return io.ErrUnexpectedEOF
	}

	q.Type = data[0]
	q.Version = guint32(data[1:5])
	q.Path = guint64(data[5:13])
	return nil
}

func (q Qid) String() string {
	return fmt.Sprintf("type:%d version:%.8x path:%.16x", q.Type, q.Version, q.Path)
}

// Stat describes a directory entry. It is contained in Rstat and Twstat
// messages. Tread requests on directories return a Stat structure for
// each directory entry.
type Stat struct {
	Type   uint16
	Dev    uint32
	Qid    Qid
	Mode   uint32
	Atime  uint32
	Mtime  uint32
	Length uint64
	Name   string
	Uid    string
	Gid    string
	Muid   string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s Stat) MarshalBinary() ([]byte, error) {
	strLen := len(s.Name) + len(s.Uid) + len(s.Gid) + len(s.Muid)
	statLen := uint16(fixedStatLen + strLen)
	data := make([]byte, 2+statLen)
	puint16(data[0:2], statLen)
	puint16(data[2:4], s.Type)
	puint32(data[4:8], s.Dev)
	data[8] = s.Qid.Type
	puint32(data[9:13], s.Qid.Version)
	puint64(data[13:21], s.Qid.Path)
	puint32(data[21:25], s.Mode)
	puint32(data[25:29], s.Atime)
	puint32(data[29:33], s.Mtime)
	puint64(data[33:41], s.Length)

	n := 41 + pstring(data[41:], s.Name)
	n += pstring(data[n:], s.Uid)
	n += pstring(data[n:], s.Gid)
	n += pstring(data[n:], s.Muid)
	return data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *Stat) UnmarshalBinary(data []byte) error {
	if len(data) < 2+fixedStatLen {
		return io.ErrUnexpectedEOF
	}

	s.Type = guint16(data[2:4])
	s.Dev = guint32(data[4:8])
	s.Qid.UnmarshalBinary(data[8:21])
	s.Mode = guint32(data[21:25])
	s.Atime = guint32(data[25:29])
	s.Mtime = guint32(data[29:33])
	s.Length = guint64(data[33:41])

	n := 41 + gstring(data[41:], &s.Name)
	n += gstring(data[n:], &s.Uid)
	n += gstring(data[n:], &s.Gid)
	n += gstring(data[n:], &s.Muid)
	return nil
}

func (s Stat) String() string {
	return fmt.Sprintf("type:0x%x dev:0x%x qid:%q mode:%d atime:%d mtime:%d length:%d name:%q uid:%q gid:%q muid:%q",
		s.Type, s.Dev, s.Qid, s.Mode, s.Atime, s.Mtime, s.Length, s.Name, s.Uid, s.Gid, s.Muid)
}

// Error represents a 9P2000 protocol error.
type Error string

func (e Error) Error() string { return string(e) }
