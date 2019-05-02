// Package binary implements little-endian translation between 9P types
// and byte sequences.
package binary

import (
	"encoding/binary"
	"io"
	"math"
)

const (
	bootstrapSize = 128 // 128 bytes should be enough for most messages
	maxInt        = int64(^uint(0) >> 1)
)

type allocator struct {
	bootstrap int
	maxdouble int
}

var defAllocator = allocator{bootstrapSize, 2 * 1024 * 1024}

func (a allocator) Grow(data []byte, size int) ([]byte, int) {
	if data == nil {
		if a.bootstrap > 0 && size < a.bootstrap {
			data = make([]byte, size, a.bootstrap)
		} else {
			data = make([]byte, size)
		}
		return data[:size], 0
	}

	off := len(data)
	if off+size > cap(data) {
		capacity := size + cap(data)*2
		if a.maxdouble > 0 && capacity > a.maxdouble {
			capacity = size + cap(data)
		}
		if int64(capacity) >= maxInt {
			panic("buffer too large")
		}

		buf := make([]byte, off, capacity)
		copy(buf[:off], data)
		data = buf
	}

	data = data[:off+size]
	return data, off
}

// grow grows the data capacity, if necessary, to guarantee space for
// another size bytes. After Grow(), at least size bytes can be written
// to data buffer without another allocation.
func grow(data []byte, size int) ([]byte, int) {
	return defAllocator.Grow(data, size)
}

// consume consumes size bytes from the buffer. It returns the unread
// portion of data, a slice with len(size) and and a success indicator.
func consume(data []byte, size int) ([]byte, []byte, bool) {
	if len(data) < size {
		return data, nil, false
	}

	buf := data[:size]
	data = data[size:]
	return data, buf, true
}

// PutString encodes a 16-bit count-delimited string to data. The
// returned slice may be a sub-slice of data if data was large enough to
// hold the entire encoded block. Otherwise, a newly allocated slice will
// be returned. It is valid to pass a nil data.
func PutString(data []byte, v string) []byte {
	size := len(v)
	if size > math.MaxUint16 {
		size = math.MaxUint16
	}

	data, n := grow(data, 2+size)
	binary.LittleEndian.PutUint16(data[n:], uint16(size))
	copy(data[n+2:], v[:size])
	return data[:n+2+size]
}

// PutUint64 encodes a unsigned 64-bit integer to data. The returned
// slice may be a sub-slice of data if data was large enough to hold the
// entire encoded block. Otherwise, a newly allocated slice will be
// returned. It is valid to pass a nil data.
func PutUint64(data []byte, v uint64) []byte {
	data, n := grow(data, 8)
	binary.LittleEndian.PutUint64(data[n:], v)
	return data[:n+8]
}

// PutUint32 encodes a unsigned 32-bit integer to data. The returned
// slice may be a sub-slice of data if data was large enough to hold the
// entire encoded block. Otherwise, a newly allocated slice will be
// returned. It is valid to pass a nil data.
func PutUint32(data []byte, v uint32) []byte {
	data, n := grow(data, 4)
	binary.LittleEndian.PutUint32(data[n:], v)
	return data[:n+4]
}

// PutUint16 encodes a unsigned 16-bit integer to data. The returned
// slice may be a sub-slice of data if data was large enough to hold the
// entire encoded block. Otherwise, a newly allocated slice will be
// returned. It is valid to pass a nil data.
func PutUint16(data []byte, v uint16) []byte {
	data, n := grow(data, 2)
	binary.LittleEndian.PutUint16(data[n:], v)
	return data[:n+2]
}

// PutUint8 encodes a unsigned 8-bit integer to data. The returned slice
// may be a sub-slice of data if data was large enough to hold the
// entire encoded block. Otherwise, a newly allocated slice will be
// returned. It is valid to pass a nil data.
func PutUint8(data []byte, v uint8) []byte {
	data, n := grow(data, 1)
	data[n] = v
	return data[:n+1]
}

// Uint8 decodes a unsigned 8-bit integer from data.
func Uint8(data []byte) uint8 { return data[0] }

// Uint16 decodes a unsigned 16-bit integer from data.
func Uint16(data []byte) uint16 {
	return binary.LittleEndian.Uint16(data)
}

// Uint32 decodes a unsigned 32-bit integer from data.
func Uint32(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data)
}

// Uint64 decodes a unsigned 64-bit integer from data.
func Uint64(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data)
}

// String decodes a 16-bit count-delimited string from data and returns
// the unread portion of data and a success indicator.
func String(data []byte) string {
	_ = data[1] // bounds check hint to compile

	var v string
	gstring(data, &v)
	return v
}

// gstring decodes a 16-bit count-delimited string from data and returns
// the unread portion of data and a success indicator.
func gstring(data []byte, v *string) ([]byte, bool) {
	data, buf, ok := consume(data, 2)
	if !ok {
		return data, false
	}

	size := int(binary.LittleEndian.Uint16(buf))
	if len(data) < size {
		return data, false
	}
	*v = string(data[:size])
	return data[size:], true
}

// guint64 decodes a unsigned 64-bit integer from data and returns the
// unread portion of data and a success indicator.
func guint64(data []byte, v *uint64) ([]byte, bool) {
	data, buf, ok := consume(data, 8)
	if !ok {
		return data, false
	}
	*v = binary.LittleEndian.Uint64(buf)
	return data, true
}

// guint32 decodes a unsigned 32-bit integer from data and returns the
// unread portion of data and a success indicator.
func guint32(data []byte, v *uint32) ([]byte, bool) {
	data, buf, ok := consume(data, 4)
	if !ok {
		return data, false
	}
	*v = binary.LittleEndian.Uint32(buf)
	return data, true
}

// guint16 decodes a unsigned 16-bit integer from data and returns the
// unread portion of data and a success indicator.
func guint16(data []byte, v *uint16) ([]byte, bool) {
	data, buf, ok := consume(data, 2)
	if !ok {
		return data, false
	}
	*v = binary.LittleEndian.Uint16(buf)
	return data, true
}

// guint8 decodes a unsigned 8-bit integer from data and returns the
// unread portion of data and a success indicator.
func guint8(data []byte, v *uint8) ([]byte, bool) {
	data, buf, ok := consume(data, 1)
	if !ok {
		return data, false
	}
	*v = buf[0]
	return data, true
}

// Marshaler is the interface implemented by an object that can marshal
// itself into a binary form. Marshal encodes the receiver into a binary
// form and returns the result.
type Marshaler interface {
	Marshal([]byte) ([]byte, int, error)
}

// Unmarshaler is the interface implemented by an object that can
// unmarshal a binary representation of itself. Unmarshal must be able
// to decode the form generated by Marshal. Unmarshal must copy the data
// if it wishes to retain the data after returning.
type Unmarshaler interface {
	Unmarshal([]byte) ([]byte, error)
}

// Marshal encodes the binary representation of v to data.
func Marshal(data []byte, v interface{}) ([]byte, int, error) {
	n := len(data)
	switch v := v.(type) {
	case *uint8:
		data = PutUint8(data, *v)
	case uint8:
		data = PutUint8(data, v)
	case *uint16:
		data = PutUint16(data, *v)
	case uint16:
		data = PutUint16(data, v)
	case *uint32:
		data = PutUint32(data, *v)
	case uint32:
		data = PutUint32(data, v)
	case *uint64:
		data = PutUint64(data, *v)
	case uint64:
		data = PutUint64(data, v)
	case *string:
		data = PutString(data, *v)
	case string:
		data = PutString(data, v)

	case Marshaler:
		return v.Marshal(data)
	}
	return data, len(data) - n, nil
}

// Unmarshal decodes the binary representation of v from data.
func Unmarshal(data []byte, v interface{}) ([]byte, error) {
	success := false
	switch v := v.(type) {
	case *uint8:
		data, success = guint8(data, v)
	case *uint16:
		data, success = guint16(data, v)
	case *uint32:
		data, success = guint32(data, v)
	case *uint64:
		data, success = guint64(data, v)
	case *string:
		data, success = gstring(data, v)

	case Unmarshaler:
		return v.Unmarshal(data)
	}

	if !success {
		return data, io.ErrUnexpectedEOF
	}
	return data, nil
}

// Buffer is a variable-sized buffer of bytes used to encode and decode
// 9P primitives and messages.
type Buffer struct {
	data []byte
	err  error
}

// NewBuffer creates and initializes a new Buffer using data as its
// initial contents. The new Buffer takes ownership of data, and the
// caller should not use data after this call. NewBuffer is intended to
// prepare a Buffer to read existing data. It can also be used to size
// the internal buffer for writing. To do that, data should have the
// desired capacity but a length of zero.
func NewBuffer(data []byte) *Buffer {
	return &Buffer{data: data}
}

// Reset resets the buffer to be empty, but it retains the underlying
// storage for use by future.
func (b *Buffer) Reset() {
	if b.data != nil {
		b.data = b.data[:0]
	}
	b.err = nil
}

// Read reads exactly size bytes from r into buffer.  The error is
// io.EOF only if no bytes were read. If an io.EOF happens after reading
// some but not all the bytes, Read returns io.ErrUnexpectedEOF.
func (b *Buffer) Read(r io.Reader, size int) error {
	if size <= 0 {
		return nil
	}

	var n int
	b.data, n = grow(b.data, size)
	_, err := io.ReadFull(r, b.data[n:n+size])
	return err
}

// Bytes returns a slice of length Len() holding the unread portion of
// the buffer. The slice is valid for use only until the next buffer
// modification.
func (b *Buffer) Bytes() []byte { return b.data }

// Len returns the number of bytes of the unread portion of Buffer.
func (b *Buffer) Len() int { return len(b.data) }

// Err returns the first error that was encountered by Buffer.
func (b *Buffer) Err() error { return b.err }

func (b *Buffer) setErr(err error) {
	if b.err == nil && err != nil {
		b.err = err
	}
}

// String decodes a 16-bit count-delimited string from Buffer.
func (b *Buffer) String() string {
	if b == nil || b.Err() != nil {
		return ""
	}

	data, buf, ok := consume(b.data, 2)
	if !ok {
		b.setErr(io.ErrUnexpectedEOF)
		return ""
	}

	size := int(binary.LittleEndian.Uint16(buf))
	if len(data) < size {
		b.setErr(io.ErrUnexpectedEOF)
		return ""
	}
	b.data = data[size:]
	return string(data[:size])
}

// Uint64 decodes a 64-bit integer from Buffer.
func (b *Buffer) Uint64() uint64 {
	if b == nil || b.Err() != nil {
		return 0
	}

	data, buf, ok := consume(b.data, 8)
	if !ok {
		b.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	b.data = data
	return binary.LittleEndian.Uint64(buf)
}

// Uint32 decodes a 32-bit integer from Buffer.
func (b *Buffer) Uint32() uint32 {
	if b == nil || b.Err() != nil {
		return 0
	}

	data, buf, ok := consume(b.data, 4)
	if !ok {
		b.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	b.data = data
	return binary.LittleEndian.Uint32(buf)
}

// Uint16 decodes a 16-bit integer from Buffer.
func (b *Buffer) Uint16() uint16 {
	if b == nil || b.Err() != nil {
		return 0
	}

	data, buf, ok := consume(b.data, 2)
	if !ok {
		b.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	b.data = data
	return binary.LittleEndian.Uint16(buf)
}

// Uint8 decodes a 8-bit integer from Buffer.
func (b *Buffer) Uint8() uint8 {
	if b == nil || b.Err() != nil {
		return 0
	}

	data, buf, ok := consume(b.data, 1)
	if !ok {
		b.setErr(io.ErrUnexpectedEOF)
		return 0
	}
	b.data = data
	return buf[0]
}

// PutString encodes a 16-bit count-delimited string to Buffer.
func (b *Buffer) PutString(v string) {
	b.data = PutString(b.data, v)
}

// PutUint64 encodes a unsigned 64-bit integer to Buffer.
func (b *Buffer) PutUint64(v uint64) {
	b.data = PutUint64(b.data, v)
}

// PutUint32 encodes a unsigned 32-bit integer to Buffer.
func (b *Buffer) PutUint32(v uint32) {
	b.data = PutUint32(b.data, v)
}

// PutUint16 encodes a unsigned 16-bit integer to Buffer.
func (b *Buffer) PutUint16(v uint16) {
	b.data = PutUint16(b.data, v)
}

// PutUint8 encodes a unsigned 8-bit integer to Buffer.
func (b *Buffer) PutUint8(v uint8) {
	b.data = PutUint8(b.data, v)
}
