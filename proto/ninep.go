package proto

import (
	"fmt"
	"io"
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

// Error represents a 9P2000 protocol error.
type Error string

func (e Error) Error() string { return string(e) }

const (
	errElemInvalidUTF8 = Error("path element name is not valid utf8")
	errMaxWalkElem     = Error("maximum walk elements exceeded")
	errElemTooLarge    = Error("path element name too large")
	errPathTooLarge    = Error("file tree name too large")
	errPathName        = Error("separator in path element")

	errVersionInvalidUTF8 = Error("version is not valid utf8")
	errVersionTooLarge    = Error("version is too large")
	errUnameInvalidUTF8   = Error("username is not valid utf8")
	errUnameTooLarge      = Error("username is too large")

	errDataTooLarge = Error("maximum data bytes exeeded")

	errInvalidMessageType = Error("invalid message type")
	errMessageTooLarge    = Error("message too large")
	errMessageTooSmall    = Error("message too small")
)

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
	UID    string
	GID    string
	MUID   string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s Stat) MarshalBinary() ([]byte, error) {
	size := s.len()
	buf := newBuffer(size)
	enc := newEncoder(&buf)
	enc.pstat(s, uint16(size))
	if enc.err != nil {
		return nil, enc.err
	}
	return buf, nil
}

func (s Stat) len() int {
	strLen := len(s.Name) + len(s.UID) + len(s.GID) + len(s.MUID)
	return fixedStatLen + strLen
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *Stat) UnmarshalBinary(data []byte) error {
	if len(data) < 2+fixedStatLen {
		return io.ErrUnexpectedEOF
	}

	*s = gstat(data)
	return nil
}

// Qid represents the server's unique identification for the file being
// accessed. Two files on the same server hierarchy are the same if and
// only if their qids are the same.
type Qid struct {
	Type uint8 // type of a file (directory, etc)

	// Version is a version:number for a file. It is incremented every
	// time a file is modified. By convention, synthetic files usually
	// have a verison number of 0. Traditional files have a version
	// numb:r that is a hash of their modification time.
	Version uint32

	// Path is an integer unique among all files in the hierarchy. If
	// a file is deleted and recreated with the same name in the same
	// directory, the old and new path components of the qids should
	// be different.
	Path uint64
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (q Qid) MarshalBinary() ([]byte, error) {
	buf := newBuffer(13)
	enc := newEncoder(&buf)
	enc.pqid(q)
	if enc.err != nil {
		return nil, enc.err
	}
	return buf, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (q *Qid) UnmarshalBinary(data []byte) error {
	if len(data) < 13 {
		return io.ErrUnexpectedEOF
	}

	*q = gqid(data)
	return nil
}

func (q Qid) String() string {
	t := ""
	if q.Type&QTDIR != 0 {
		t += "d"
	}
	if q.Type&QTAPPEND != 0 {
		t += "a"
	}
	if q.Type&QTEXCL != 0 {
		t += "l"
	}
	if q.Type&QTMOUNT != 0 {
		t += "m"
	}
	if q.Type&QTAUTH != 0 {
		t += "A"
	}
	if q.Type&QTTMP != 0 {
		t += "t"
	}
	return fmt.Sprintf("(%.16x %d %q)", q.Path, q.Version, t)
}

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
}
