package proto

import (
	"fmt"

	"github.com/azmodb/ninep/binary"
)

// String implements fmt.Stringer.
func (m Header) String() string { return fmt.Sprintf("type:%s tag:%d", m.Type, m.Tag) }

// Len returns the length of the message in bytes.
func (m Header) Len() int { return 1 + 2 }

// Reset resets all state.
func (m *Header) Reset() { *m = Header{} }

// Encode encodes to the given binary.Buffer.
func (m Header) Encode(buf *binary.Buffer) {
	buf.PutUint8(uint8(m.Type))
	buf.PutUint16(m.Tag)
}

// Decode decodes from the given binary.Buffer.
func (m *Header) Decode(buf *binary.Buffer) {
	m.Type = MessageType(buf.Uint8())
	m.Tag = buf.Uint16()
}

// String implements fmt.Stringer.
func (m Tversion) String() string {
	return fmt.Sprintf("message_size:%d version:%q", m.MessageSize, m.Version)
}

// Len returns the length of the message in bytes.
func (m Tversion) Len() int { return 4 + 2 + len(m.Version) }

// Reset resets all state.
func (m *Tversion) Reset() { *m = Tversion{} }

// Encode encodes to the given binary.Buffer.
func (m Tversion) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.MessageSize)
	buf.PutString(m.Version)
}

// Decode decodes from the given binary.Buffer.
func (m *Tversion) Decode(buf *binary.Buffer) {
	m.MessageSize = buf.Uint32()
	m.Version = buf.String()
}

// String implements fmt.Stringer.
func (m Rversion) String() string {
	return fmt.Sprintf("message_size:%d version:%q", m.MessageSize, m.Version)
}

// Len returns the length of the message in bytes.
func (m Rversion) Len() int { return 4 + 2 + len(m.Version) }

// Reset resets all state.
func (m *Rversion) Reset() { *m = Rversion{} }

// Encode encodes to the given binary.Buffer.
func (m Rversion) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.MessageSize)
	buf.PutString(m.Version)
}

// Decode decodes from the given binary.Buffer.
func (m *Rversion) Decode(buf *binary.Buffer) {
	m.MessageSize = buf.Uint32()
	m.Version = buf.String()
}

// String implements fmt.Stringer.
func (m Tflush) String() string { return fmt.Sprintf("old_tag:%d", m.OldTag) }

// Len returns the length of the message in bytes.
func (m Tflush) Len() int { return 2 }

// Reset resets all state.
func (m *Tflush) Reset() { *m = Tflush{} }

// Encode encodes to the given binary.Buffer.
func (m Tflush) Encode(buf *binary.Buffer) {
	buf.PutUint16(m.OldTag)
}

// Decode decodes from the given binary.Buffer.
func (m *Tflush) Decode(buf *binary.Buffer) {
	m.OldTag = buf.Uint16()
}

// String implements fmt.Stringer.
func (m Rflush) String() string { return "" }

// Len returns the length of the message in bytes.
func (m Rflush) Len() int { return 0 }

// Reset resets all state.
func (m *Rflush) Reset() { *m = Rflush{} }

// Encode encodes to the given binary.Buffer.
func (m Rflush) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rflush) Decode(buf *binary.Buffer) {}

// String implements fmt.Stringer.
func (m Tread) String() string {
	return fmt.Sprintf("fid:%d offset:%d count:%d", m.Fid, m.Offset, m.Count)
}

// Len returns the length of the message in bytes.
func (m Tread) Len() int { return 4 + 8 + 4 }

// Reset resets all state.
func (m *Tread) Reset() { *m = Tread{} }

// Encode encodes to the given binary.Buffer.
func (m Tread) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
	buf.PutUint64(m.Offset)
	buf.PutUint32(m.Count)
}

// Decode decodes from the given binary.Buffer.
func (m *Tread) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
	m.Offset = buf.Uint64()
	m.Count = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rwrite) String() string { return fmt.Sprintf("count:%d", m.Count) }

// Len returns the length of the message in bytes.
func (m Rwrite) Len() int { return 4 }

// Reset resets all state.
func (m *Rwrite) Reset() { *m = Rwrite{} }

// Encode encodes to the given binary.Buffer.
func (m Rwrite) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Count)
}

// Decode decodes from the given binary.Buffer.
func (m *Rwrite) Decode(buf *binary.Buffer) {
	m.Count = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Tclunk) String() string { return fmt.Sprintf("fid:%d", m.Fid) }

// Len returns the length of the message in bytes.
func (m Tclunk) Len() int { return 4 }

// Reset resets all state.
func (m *Tclunk) Reset() { *m = Tclunk{} }

// Encode encodes to the given binary.Buffer.
func (m Tclunk) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tclunk) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rclunk) String() string { return "" }

// Len returns the length of the message in bytes.
func (m Rclunk) Len() int { return 0 }

// Reset resets all state.
func (m *Rclunk) Reset() { *m = Rclunk{} }

// Encode encodes to the given binary.Buffer.
func (m Rclunk) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rclunk) Decode(buf *binary.Buffer) {}

// String implements fmt.Stringer.
func (m Tremove) String() string { return fmt.Sprintf("fid:%d", m.Fid) }

// Len returns the length of the message in bytes.
func (m Tremove) Len() int { return 4 }

// Reset resets all state.
func (m *Tremove) Reset() { *m = Tremove{} }

// Encode encodes to the given binary.Buffer.
func (m Tremove) Encode(buf *binary.Buffer) {
	buf.PutUint32(m.Fid)
}

// Decode decodes from the given binary.Buffer.
func (m *Tremove) Decode(buf *binary.Buffer) {
	m.Fid = buf.Uint32()
}

// String implements fmt.Stringer.
func (m Rremove) String() string { return "" }

// Len returns the length of the message in bytes.
func (m Rremove) Len() int { return 0 }

// Reset resets all state.
func (m *Rremove) Reset() { *m = Rremove{} }

// Encode encodes to the given binary.Buffer.
func (m Rremove) Encode(buf *binary.Buffer) {}

// Decode decodes from the given binary.Buffer.
func (m *Rremove) Decode(buf *binary.Buffer) {}