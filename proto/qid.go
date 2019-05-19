package proto

import (
	"fmt"
	"io"

	"github.com/azmodb/ninep/binary"
)

// Qid represents the server's unique identification for the file being
// accessed. Two files on the same server hierarchy are the same if and
// only if their qids are the same.
type Qid struct {
	Type uint8 // type of a file (directory, etc)

	// Version is a version:number for a file. It is incremented every
	// time a file is modified. By convention, synthetic files usually
	// have a version number of 0. Traditional files have a version
	// numb:r that is a hash of their modification time.
	Version uint32

	// Path is an integer unique among all files in the hierarchy. If
	// a file is deleted and recreated with the same name in the same
	// directory, the old and new path components of the qids should
	// be different.
	Path uint64
}

// Len returns the length of the message in bytes.
func (q Qid) Len() int { return 13 }

// Reset resets all state.
func (q *Qid) Reset() { *q = Qid{} }

// IsDir reports whether q describes a directory.
func (q Qid) IsDir() bool { return q.Type&TypeDirectory != 0 }

// Encode encodes to the given binary.Buffer.
func (q Qid) Encode(buf *binary.Buffer) {
	buf.PutUint8(q.Type)
	buf.PutUint32(q.Version)
	buf.PutUint64(q.Path)
}

// Decode decodes from the given binary.Buffer.
func (q *Qid) Decode(buf *binary.Buffer) {
	q.Type = buf.Uint8()
	q.Version = buf.Uint32()
	q.Path = buf.Uint64()
}

// Marshal implements binary.Marshaler.
func (q Qid) Marshal(data []byte) ([]byte, int, error) {
	data = binary.PutUint8(data, q.Type)
	data = binary.PutUint32(data, q.Version)
	data = binary.PutUint64(data, q.Path)
	return data, 13, nil
}

// Unmarshal implements binary.Unmarshaler.
func (q *Qid) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 13 {
		return data, io.ErrUnexpectedEOF
	}
	q.Type = binary.Uint8(data[:1])
	q.Version = binary.Uint32(data[1:5])
	q.Path = binary.Uint64(data[5:13])
	return data[13:], nil
}

// Qid type field represents the type of a file (directory, etc.),
// represented as a bit vector corresponding to the high 8 bits of the
// file mode word.
const (
	TypeDirectory  = 0x80
	TypeAppendOnly = 0x40
	TypeExclusive  = 0x20
	TypeMount      = 0x10
	TypeAuth       = 0x08
	TypeTemporary  = 0x04
	TypeSymlink    = 0x02
	TypeLink       = 0x01
	TypeRegular    = 0x00
)

func (q Qid) String() string {
	t := ""
	if q.Type&TypeDirectory != 0 {
		t += "d"
	}
	if q.Type&TypeAppendOnly != 0 {
		t += "a"
	}
	if q.Type&TypeExclusive != 0 {
		t += "l"
	}
	if q.Type&TypeMount != 0 {
		t += "m"
	}
	if q.Type&TypeAuth != 0 {
		t += "A"
	}
	if q.Type&TypeTemporary != 0 {
		t += "t"
	}
	if q.Type&TypeSymlink != 0 {
		t += "S"
	}
	if q.Type&TypeLink != 0 {
		t += "L"
	}
	if len(t) == 0 {
		t = "f"
	}
	return fmt.Sprintf("type:%q version:%d path:%.16x", t, q.Version, q.Path)
}
