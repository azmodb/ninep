package ninep

import (
	"encoding/binary"
	"errors"
	"io"
)

// Encoder writes 9P2000 message to the underlying output stream. If an
// error occures, encoding to a Encoder, no more data will be accepted
// and all subsequent calls will return the error.
//
// Encoder is not safe for concurrent use. Usage of any Encoder method
// should be protected by a mutex.
type Encoder struct {
	buf [20]byte
	w   writer
	err error
}

type Option func(interface{})

type writer interface {
	WriteString(string) (int, error)
	io.ByteWriter
	io.Writer
}

func NewEncoder(w io.Writer, opts ...Option) *Encoder {
	br, ok := w.(writer)
	if !ok {
		// TODO
	}

	e := &Encoder{w: br}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Encoder) Encode(m *Message) error {
	if m.Type == Terror || m.Type < Tversion || m.Type > Rwstat {
		return errors.New("unexpected message type")
	}
	return e.encode(m)
}

func (e *Encoder) encode(m *Message) error {
	e.pheader(sizeFunc[m.Type](m), m.Type, m.Tag)

	switch m.Type {
	case Tversion, Rversion:
		e.puint32(m.Msize)
		e.pstring(m.Version)
	case Tauth:
		e.puint32(m.Afid)
		e.pstring(m.Uname)
		e.pstring(m.Aname)
	case Tattach:
		e.puint32(m.Fid)
		e.puint32(m.Afid)
		e.pstring(m.Uname)
		e.pstring(m.Aname)
	case Rauth, Rattach:

	}
	panic("encode: unexpected message type")
}

func (e *Encoder) pheader(size uint32, typ uint8, tag uint16) {
	puint32(e.buf[:4], size)
	e.buf[4] = typ
	puint16(e.buf[5:7], tag)
	e.write(e.buf[:7])
}

/*
func (e *Encoder) pstat(v Stat, size uint16) {
	puint16(e.buf[0:2], size)
	puint16(e.buf[2:4], v.Type)
	puint32(e.buf[4:8], v.Dev)
	e.write(e.buf[:8])
	e.pqid(v.Qid)
	puint32(e.buf[0:4], v.Mode)
	puint32(e.buf[4:8], v.Atime)
	puint32(e.buf[8:12], v.Mtime)
	puint64(e.buf[12:20], v.Length)
	e.write(e.buf[:20])
	e.pstring(v.Name)
	e.pstring(v.Uid)
	e.pstring(v.Gid)
	e.pstring(v.Muid)
}

func (e *Encoder) pqid(v Qid) {
	e.buf[0] = v.Type
	puint32(e.buf[1:5], v.Version)
	puint64(e.buf[5:13], v.Path)
	e.write(e.buf[:13])
}
*/

func (e *Encoder) pstring(v string) {
	puint16(e.buf[:2], uint16(len(v)))
	e.write(e.buf[:2])
	e.writeString(v)
}

func (e *Encoder) puint64(v uint64) {
	puint64(e.buf[:8], v)
	e.write(e.buf[:8])
}

func (e *Encoder) puint32(v uint32) {
	puint32(e.buf[:4], v)
	e.write(e.buf[:4])
}

func (e *Encoder) puint16(v uint16) {
	puint16(e.buf[:2], v)
	e.write(e.buf[:2])
}

func (e *Encoder) puint8(v uint8) {
	e.writeByte(v)
}

func (e *Encoder) writeString(p string) {
	if e.err != nil || len(p) == 0 {
		return
	}
	_, e.err = e.w.WriteString(p)
}

func (e *Encoder) write(p []byte) {
	if e.err != nil || len(p) == 0 {
		return
	}
	_, e.err = e.w.Write(p)
}

func (e *Encoder) writeByte(p byte) {
	if e.err != nil {
		return
	}
	e.err = e.w.WriteByte(p)
}

var (
	puint64 = binary.LittleEndian.PutUint64
	puint32 = binary.LittleEndian.PutUint32
	puint16 = binary.LittleEndian.PutUint16
)

func puint8(b []byte, v uint8) {
	_ = b[0] // bounds check hint to compiler
	b[0] = v
}

func pstring(b []byte, s string) int {
	_ = b[2+len(s)-1] // bounds check hint to compiler
	puint16(b[0:2], uint16(len(s)))
	if len(s) == 0 {
		return 2
	}
	return 2 + copy(b[2:], s)
}

func guint64(b []byte, v *uint64) []byte {
	*v = binary.LittleEndian.Uint64(b)
	return b[8:]
}

func guint32(b []byte, v *uint32) []byte {
	*v = binary.LittleEndian.Uint32(b)
	return b[4:]
}

func guint16(b []byte, v *uint16) []byte {
	*v = binary.LittleEndian.Uint16(b)
	return b[2:]
}

func guint8(b []byte, v *uint8) []byte {
	*v = b[0]
	return b[1:]
}

func gstring(b []byte, s *string) []byte {
	size := int(binary.LittleEndian.Uint16(b))
	if size == 0 {
		*s = ""
		return b[2:]
	}
	*s = string(b[2 : 2+size])
	return b[2+size:]
}
