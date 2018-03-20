package ninep

import (
	"bufio"
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

// Option configures a encoder/decoder.
type Option func(interface{})

type writer interface {
	WriteString(string) (int, error)
	io.ByteWriter
	io.Writer
}

// NewEncoder returns a new encoder that will transmit on the io.Writer.
func NewEncoder(w io.Writer, opts ...Option) *Encoder {
	bw, ok := w.(writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}

	e := &Encoder{w: bw}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Encode transmits the data item represented by Message. Passing a nil
// pointer to Encoder will panic, as they cannot be transmitted.
func (e *Encoder) Encode(m *Message) error {
	if m.Type == Terror || m.Type < Tversion || m.Type > Rwstat {
		return errInvalidMessageType
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
		e.pqid(m.Qid)
	case Rerror:
		e.pstring(m.Ename)
	case Tflush:
		e.puint16(m.Oldtag)
	case Twalk:
	case Rwalk:
	case Topen:
		e.puint32(m.Fid)
		e.puint8(m.Mode)
	case Tcreate:
		e.puint32(m.Fid)
		e.pstring(m.Name)
		e.puint32(m.Perm)
		e.puint8(m.Mode)
	case Ropen, Rcreate:
		e.pqid(m.Qid)
		e.puint32(m.Iounit)
	case Tread:
		e.puint32(m.Fid)
		e.puint64(m.Offset)
		e.puint32(m.Count)
	case Rread:
		e.puint32(m.Count)
		e.pdata(m.Data)
	case Twrite:
		e.puint32(m.Fid)
		e.puint64(m.Offset)
		e.puint32(m.Count)
		e.pdata(m.Data)
	case Rwrite:
		e.puint32(m.Count)
	case Tclunk, Tremove, Tstat:
		e.puint32(m.Fid)
	case Rstat:
		e.pstat(m.Stat)
	case Twstat:
		e.puint32(m.Fid)
		e.pstat(m.Stat)
	case Rflush, Rclunk, Rremove, Rwstat:
		// nothing
	default:
		panic("encode: unexpected message type")
	}
	return e.err
}

func (e *Encoder) pheader(size uint32, typ uint8, tag uint16) {
	puint32(e.buf[:4], size)
	e.buf[4] = typ
	puint16(e.buf[5:7], tag)
	e.write(e.buf[:7])
}

func (e *Encoder) pstat(v Stat) {
	puint16(e.buf[0:2], uint16(v.len()))
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
	e.pstring(v.UID)
	e.pstring(v.GID)
	e.pstring(v.MUID)
}

func (e *Encoder) pqid(v Qid) {
	e.buf[0] = v.Type
	puint32(e.buf[1:5], v.Version)
	puint64(e.buf[5:13], v.Path)
	e.write(e.buf[:13])
}

func (e *Encoder) pdata(v []byte) {
	puint32(e.buf[:4], uint32(len(v)))
	e.write(e.buf[:4])
	e.write(v)
}

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
