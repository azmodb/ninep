package proto

import (
	"fmt"
	"io"
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

func (s Stat) String() string {
	return fmt.Sprintf("(%q %q %q %s q %v m %#o at %d mt %d l %d t %d d %d)",
		s.Name, s.UID, s.GID, s.MUID, s.Qid, s.Mode,
		s.Atime, s.Mtime, s.Length, s.Type, s.Dev)
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
	/*
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
	*/
	return fmt.Sprintf("(%.16x %d %d)", q.Path, q.Version, q.Type)
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
