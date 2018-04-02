package proto

import (
	"fmt"
	"io"
)

// Encoder writes 9P2000 messages to the underlying output stream. If an
// error occures, encoding to a Encoder, no more data will be accepted
// and all subsequent calls will return the error.
//
// Encoder is not safe for concurrent use. Usage of any Encoder method
// should be protected by a mutex.
type Encoder interface {
	Tversion(tag uint16, msize uint32, version string) error
	Rversion(tag uint16, msize uint32, version string) error
	Tauth(tag uint16, afid uint32, uname, aname string) error
	Rauth(tag uint16, qid Qid) error
	Tattach(tag uint16, fid, afid uint32, uname, aname string) error
	Rattach(tag uint16, qid Qid) error
	Tflush(tag, oldtag uint16) error
	Rflush(tag uint16) error
	Twalk(tag uint16, fid, newfid uint32, names ...string) error
	Rwalk(tag uint16, qids ...Qid) error
	Topen(tag uint16, fid uint32, mode uint8) error
	Ropen(tag uint16, qid Qid, iouint uint32) error
	Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) error
	Rcreate(tag uint16, qid Qid, iouint uint32) error
	Tread(tag uint16, fid uint32, offset uint64, count uint32) error
	Rread(tag uint16, data []byte) error
	Twrite(tag uint16, fid uint32, offset uint64, data []byte) error
	Rwrite(tag uint16, count uint32) error
	Tclunk(tag uint16, fid uint32) error
	Rclunk(tag uint16) error
	Tremove(tag uint16, fid uint32) error
	Rremove(tag uint16) error
	Tstat(tag uint16, fid uint32) error
	//Rstat(tag uint16, stat Stat) error
	//Twstat(tag uint16, fid uint32, stat Stat) error
	Rwstat(tag uint16) error

	Rerrorf(tag uint16, format string, args ...interface{}) error
	Rerror(tag uint16, err error) error
}

type Option func(interface{})

type encoder struct {
	w   io.Writer
	err error
	buf [20]byte
}

// NewEncoder returns a new encoder that will transmit on the io.Writer.
func NewEncoder(w io.Writer, opts ...Option) Encoder {
	return newEncoder(w, opts...)
}

func newEncoder(w io.Writer, opts ...Option) *encoder {
	e := &encoder{w: w}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *encoder) version(typ uint8, tag uint16, msize uint32, version string) error {
	size := minSize(typ) + uint32(len(version))
	e.pheader(size, typ, tag)
	e.puint32(msize)
	e.pstring(version)
	return e.err
}

func (e *encoder) Tversion(tag uint16, msize uint32, version string) error {
	return e.version(msgTversion, tag, msize, version)
}

func (e *encoder) Rversion(tag uint16, msize uint32, version string) error {
	return e.version(msgRversion, tag, msize, version)
}

func (e *encoder) Tauth(tag uint16, afid uint32, uname, aname string) error {
	size := minSize(msgTauth) + uint32(len(uname)+len(aname))
	e.pheader(size, msgTauth, tag)
	e.puint32(afid)
	e.pstring(uname)
	e.pstring(aname)
	return e.err
}

func (e *encoder) Tattach(tag uint16, fid, afid uint32, uname, aname string) error {
	size := minSize(msgTattach) + uint32(len(uname)+len(aname))
	e.pheader(size, msgTattach, tag)
	e.puint32(fid)
	e.puint32(afid)
	e.pstring(uname)
	e.pstring(aname)
	return e.err
}

func (e *encoder) rqid(typ uint8, tag uint16, qid Qid) error {
	e.pheader(minSize(typ), msgRauth, tag)
	e.pqid(qid)
	return e.err
}

func (e *encoder) Rauth(tag uint16, aqid Qid) error {
	return e.rqid(msgRauth, tag, aqid)
}

func (e *encoder) Rattach(tag uint16, qid Qid) error {
	return e.rqid(msgRattach, tag, qid)
}

func (e *encoder) Tflush(tag, oldtag uint16) error {
	e.pheader(minSize(msgTflush), msgTflush, tag)
	e.puint16(oldtag)
	return e.err
}

func (e *encoder) rheader(typ uint8, tag uint16) error {
	e.pheader(minSize(typ), typ, tag)
	return e.err
}

func (e *encoder) Rflush(tag uint16) error {
	return e.rheader(msgRflush, tag)
}

func (e *encoder) Twalk(tag uint16, fid, newfid uint32, names ...string) error {
	size := minSize(msgTwalk)
	for _, n := range names {
		size += 2 + uint32(len(n))
	}

	e.pheader(size, msgTwalk, tag)
	e.puint32(fid)
	e.puint32(newfid)

	puint16(e.buf[:2], uint16(len(names))) // TODO: entries check
	e.write(e.buf[:2])
	for _, n := range names {
		e.pstring(n)
	}
	return e.err
}

func (e *encoder) Rwalk(tag uint16, qids ...Qid) error {
	size := minSize(msgRwalk) + uint32(13*len(qids))
	e.pheader(size, msgRwalk, tag)

	puint16(e.buf[:2], uint16(len(qids)))
	e.write(e.buf[:2])
	for _, q := range qids {
		e.pqid(q)
	}
	return e.err
}

func (e *encoder) Topen(tag uint16, fid uint32, mode uint8) error {
	e.pheader(minSize(msgTopen), msgTopen, tag)
	e.puint32(fid)
	e.puint8(mode)
	return e.err
}

func (e *encoder) ropen(typ uint8, tag uint16, qid Qid, iounit uint32) error {
	e.pheader(minSize(typ), typ, tag)
	e.pqid(qid)
	e.puint32(iounit)
	return e.err
}

func (e *encoder) Ropen(tag uint16, qid Qid, iounit uint32) error {
	return e.ropen(msgRopen, tag, qid, iounit)
}

func (e *encoder) Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) error {
	size := minSize(msgTcreate) + uint32(len(name))
	e.pheader(size, msgTcreate, tag)
	e.puint32(fid)
	e.pstring(name)
	e.puint32(perm)
	e.puint8(mode)
	return e.err
}

func (e *encoder) Rcreate(tag uint16, qid Qid, iounit uint32) error {
	return e.ropen(msgRcreate, tag, qid, iounit)
}

func (e *encoder) Tread(tag uint16, fid uint32, offset uint64, count uint32) error {
	e.pheader(minSize(msgTread), msgTread, tag)
	e.puint32(fid)
	e.puint64(offset)
	e.puint32(count)
	return e.err
}

func (e *encoder) Rread(tag uint16, data []byte) error {
	size := minSize(msgRread) + uint32(len(data))
	e.pheader(size, msgRread, tag)
	e.pdata(data)
	return e.err
}

func (e *encoder) Twrite(tag uint16, fid uint32, offset uint64, data []byte) error {
	size := minSize(msgTwrite) + uint32(len(data))
	e.pheader(size, msgTwrite, tag)
	e.puint32(fid)
	e.puint64(offset)
	e.pdata(data)
	return e.err
}

func (e *encoder) Rwrite(tag uint16, count uint32) error {
	e.pheader(minSize(msgRwrite), msgRwrite, tag)
	e.puint32(count)
	return e.err
}

func (e *encoder) rfid(typ uint8, tag uint16, fid uint32) error {
	e.pheader(minSize(typ), typ, tag)
	e.puint32(fid)
	return e.err
}

func (e *encoder) Tclunk(tag uint16, fid uint32) error {
	return e.rfid(msgTclunk, tag, fid)
}

func (e *encoder) Rclunk(tag uint16) error {
	return e.rheader(msgRclunk, tag)
}

func (e *encoder) Tremove(tag uint16, fid uint32) error {
	return e.rfid(msgTremove, tag, fid)
}

func (e *encoder) Rremove(tag uint16) error {
	return e.rheader(msgRremove, tag)
}

func (e *encoder) Tstat(tag uint16, fid uint32) error {
	return e.rfid(msgTstat, tag, fid)
}

func (e *encoder) Rstat(tag uint16, stat Stat) error {
	panic("TODO")
}

func (e *encoder) Twstat(tag uint16, fid uint32, stat Stat) error {
	panic("TODO")
}

func (e *encoder) Rwstat(tag uint16) error {
	return e.rheader(msgRwstat, tag)
}

func (e *encoder) Rerrorf(tag uint16, format string, args ...interface{}) error {
	return e.rerror(tag, fmt.Sprintf(format, args...))
}

func (e *encoder) Rerror(tag uint16, err error) error {
	return e.rerror(tag, err.Error())
}

func (e *encoder) rerror(tag uint16, msg string) error {
	size := minSize(msgRerror) + uint32(len(msg))
	e.pheader(size, msgRerror, tag)
	e.pstring(msg)
	return e.err
}

func (e *encoder) pheader(size uint32, typ uint8, tag uint16) {
	puint32(e.buf[:4], size)
	e.buf[4] = typ
	puint16(e.buf[5:7], tag)
	e.write(e.buf[:7])
}

func (e *encoder) pstat(v Stat, size uint16) {
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
	e.pstring(v.UID)
	e.pstring(v.GID)
	e.pstring(v.MUID)
}

func (e *encoder) pqid(v Qid) {
	e.buf[0] = v.Type
	puint32(e.buf[1:5], v.Version)
	puint64(e.buf[5:13], v.Path)
	e.write(e.buf[:13])
}

func (e *encoder) pdata(v []byte) {
	puint32(e.buf[:4], uint32(len(v)))
	e.write(e.buf[:4])
	e.write(v)
}

func (e *encoder) pstring(v string) {
	puint16(e.buf[:2], uint16(len(v)))
	e.write(e.buf[:2])
	e.writeString(v)
}

func (e *encoder) puint64(v uint64) {
	puint64(e.buf[:8], v)
	e.write(e.buf[:8])
}

func (e *encoder) puint32(v uint32) {
	puint32(e.buf[:4], v)
	e.write(e.buf[:4])
}

func (e *encoder) puint16(v uint16) {
	puint16(e.buf[:2], v)
	e.write(e.buf[:2])
}

func (e *encoder) puint8(v uint8) {
	e.buf[0] = v
	e.write(e.buf[:1])
}

func (e *encoder) write(v []byte) {
	if e.err != nil {
		return
	}
	_, e.err = e.w.Write(v)
}

func (e *encoder) writeString(v string) {
	if e.err != nil {
		return
	}
	_, e.err = io.WriteString(e.w, v)
}
