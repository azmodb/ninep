package proto

import (
	"fmt"
	"strings"

	"github.com/azmodb/ninep/binary"
)

// Header represents the protocol header. Each 9P message begins with a
// four–byte size field specifying the length in bytes of the complete
// message including the four bytes of the size field itself.
//
// The next byte is the message type and two bytes representing an
// identifying tag.
type Header struct {
	Type MessageType // Type byte is the 9P message type.
	Tag  uint16
}

// Tversion request negotiates the protocol version and message size to be
// used on the connection and initializes the connection for I/O. Tversion
// must be the first message sent on the 9P connection, and the client
// cannot issue any further requests until it has received the Rversion
// reply.
type Tversion struct {
	// MessageSize returns the maximum length, in bytes, that the client
	// will ever generate or expect to receive in a single 9P message.
	// This count includes all 9P protocol data, starting from the size
	// field and extending through the message, but excludes enveloping
	// transport protocols.
	MessageSize uint32

	// Version identifies the level of the protocol that the client
	// supports. The string must always begin with the two characters
	// "9P".
	Version string
}

// Rversion reply is sent in response to a Tversion request. It contains
// the version of the protocol that the server has chosen, and the
// maximum size of all successive messages.
type Rversion struct {
	// MessageSize returns the maximum size (in bytes) of any 9P message
	// that it will send or accept, and must be equal to or less than
	// the maximum suggested in the preceding Tversion message. After the
	// Rversion message is received, both sides of the connection must
	// honor this limit.
	MessageSize uint32

	// Version identifies the level of the protocol that the server
	// supports. If a server does not understand the protocol version
	// sent in a Tversion message, Version will return the string
	// "unknown". A server may choose to specify a version that is less
	// than or equal to that supported by the client.
	Version string
}

// Tflush request is sent to the server to purge the pending response.
type Tflush struct {
	OldTag uint16
}

// Rflush echoes the tag (not oldtag) of the Tflush message.
type Rflush struct{}

// Twalk message is used to descend a directory hierarchy.
type Twalk struct {
	// Fid must have been established by a previous transaction, such as
	// an Tattach.
	Fid uint32

	// NewFid contains the proposed fid that the client wishes to
	// associate with the result of traversing the directory hierarchy.
	NewFid uint32

	// Names contains an ordered list of path name elements that the
	// client wishes to descend into in succession.
	//
	// To simplify the implementation, a maximum of sixteen name elements
	// may be packed in a single message
	Names []string
}

// String implements fmt.Stringer.
func (m Twalk) String() string {
	return fmt.Sprintf("fid:%d new_fid:%d names:%v", m.Fid, m.NewFid, m.Names)
}

// Reset resets all state.
func (m *Twalk) Reset() { *m = Twalk{} }

// Len returns the length of the message in bytes.
func (m Twalk) Len() int { return 4 + 4 + nlen(m.Names) }

// Encode encodes to the given binary.Buffer. To simplify the
// implementation, a maximum of sixteen name elements may be packed in a
// single message.
func (m Twalk) Encode(buf *binary.Buffer) {
	if len(m.Names) > 16 {
		m.Names = m.Names[:16]
	}
	buf.PutUint32(m.Fid)
	buf.PutUint32(m.NewFid)
	buf.PutUint16(uint16(len(m.Names)))
	for _, name := range m.Names {
		buf.PutString(name)
	}
}

// Decode decodes from the given binary.Buffer. To simplify the
// implementation, a maximum of sixteen name elements may be packed in a
// single message.
func (m *Twalk) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.NewFid = buf.Uint32()
	size := buf.Uint16()
	if size > 16 {
		size = 16
	}
	if size == 0 {
		return
	}
	m.Names = make([]string, size)
	for i := range m.Names {
		m.Names[i] = buf.String()
	}
}

// Rwalk message contains a server's reply to a successful Twalk request.
// If the first path in the corresponding Twalk request cannot be walked,
// an Rerror message is returned instead.
type Rwalk []Qid

// String implements fmt.Stringer.
func (m Rwalk) String() string {
	if len(m) == 0 {
		return "qids:[]"
	}

	b := &strings.Builder{}
	b.WriteString("qids:[")
	for i, qid := range m {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(qid.String())
	}
	b.WriteString("]")
	return b.String()
}

// Len returns the length of the message in bytes.
func (m Rwalk) Len() int { return 13 * len(m) }

// Reset resets all state.
func (m *Rwalk) Reset() { *m = Rwalk{} }

// Encode encodes to the given binary.Buffer. To simplify the
// implementation, a maximum of sixteen qid elements may be packed in a
// single message.
func (m *Rwalk) Encode(buf *binary.Buffer) {
	if len(*m) > 16 {
		*m = (*m)[:16]
	}
	buf.PutUint16(uint16(len(*m)))
	for _, qid := range *m {
		qid.Encode(buf)
	}
}

// Decode decodes from the given binary.Buffer. To simplify the
// implementation, a maximum of sixteen name elements may be packed in a
// single message.
func (m *Rwalk) Decode(buf *binary.Buffer) {
	size := buf.Uint16()
	if size > 16 {
		size = 16
	}
	if size == 0 {
		return
	}
	*m = make([]Qid, size)
	for i := range *m {
		(*m)[i].Decode(buf)
	}
}

// Tread request asks for count bytes of data from the file, which must
// be opened for reading, starting offset bytes after the beginning of
// the file.
type Tread struct {
	Fid    uint32
	Offset uint64
	Count  uint32
}

// Rread message returns the bytes requested by a Tread message.
type Rread struct {
	Data []byte
}

// String implements fmt.Stringer.
func (m Rread) String() string {
	return fmt.Sprintf("data_len:%d", len(m.Data))
}

// Len returns the length of the message in bytes.
func (m Rread) Len() int { return 4 + len(m.Data) }

// Reset resets all state.
func (m *Rread) Reset() { *m = Rread{} }

// Encode encodes to the given binary.Buffer.
func (m Rread) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rread) Decode(buf *binary.Buffer) {}

// Payload returns the payload for sending.
func (m Rread) Payload() []byte { return m.Data }

// PutPayload sets the decoded payload.
func (m *Rread) PutPayload(b []byte) { m.Data = b }

// FixedLen returns the fixed message size in bytes.
func (m Rread) FixedLen() int { return 0 }

// Twrite message is sent by a client to write data to a file.
type Twrite struct {
	Fid    uint32
	Offset uint64
	Data   []byte
}

// String implements fmt.Stringer.
func (m Twrite) String() string {
	return fmt.Sprintf("fid:%d offset:%d data_len:%d", m.Fid, m.Offset, len(m.Data))
}

// Len returns the length of the message in bytes.
func (m Twrite) Len() int { return 4 + len(m.Data) }

// Reset resets all state.
func (m *Twrite) Reset() { *m = Twrite{} }

// Encode encodes to the given binary.Buffer.
func (m Twrite) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint64(m.Offset)
}

// Decode decodes from the given binary.Buffer.
func (m *Twrite) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.Offset = buf.Uint64()
}

// Payload returns the payload for sending.
func (m Twrite) Payload() []byte { return m.Data }

// PutPayload sets the decoded payload.
func (m *Twrite) PutPayload(b []byte) { m.Data = b }

// FixedLen returns the fixed message size in bytes.
func (m Twrite) FixedLen() int { return 4 + 8 }

// Rwrite message returns the bytes requested by a Twrite message.
type Rwrite struct {
	Count uint32
}

// Tclunk request informs the file server that the current file
// represented by fid is no longer needed by the client. The actual file
// is not removed on the server unless the fid had been opened with
// ORCLOSE.
type Tclunk struct {
	Fid uint32
}

// Rclunk message contains a servers response to a Tclunk request.
type Rclunk struct{}

// Tremove request asks the file server both to remove a file and to
// clunk it, even if the remove fails. This request will fail if the
// client does not have write permission in the parent directory.
type Tremove struct {
	Fid uint32
}

// Rremove message contains a servers response to a Tremove request. An
// Rremove message is only sent if the server determined that the
// requesting user had the proper permissions required for the Tremove to
// succeed, otherwise Rerror is returned.
type Rremove struct{}

func nlen(arr []string) (size int) {
	size = 2
	for _, n := range arr {
		size += 2 + len(n)
	}
	return size
}
