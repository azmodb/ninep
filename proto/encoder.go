package proto

import (
	"bufio"
	"fmt"
	"io"
)

// Encoder writes 9P2000 message to the underlying output stream. If an
// error occures, encoding to a Encoder, no more data will be accepted
// and all subsequent calls will return the error. After all data has
// been encoded, the client should call the Flush method to guarantee
// all data has been forworded to the underlying output stream.
//
// Encoder is not safe for concurrent use. Usage of any Encoder method
// should be protected by a mutex.
type Encoder interface {
	Tversion(tag uint16, msize uint32, version string)
	Rversion(tag uint16, msize uint32, version string)
	Tauth(tag uint16, afid uint32, uname, aname string)
	Rauth(tag uint16, qid Qid)
	Tattach(tag uint16, fid, afid uint32, uname, aname string)
	Rattach(tag uint16, qid Qid)
	Tflush(tag, oldtag uint16)
	Rflush(tag uint16)
	Twalk(tag uint16, fid, newfid uint32, names ...string)
	Rwalk(tag uint16, qids ...Qid)
	Topen(tag uint16, fid uint32, mode uint8)
	Ropen(tag uint16, qid Qid, iouint uint32)
	Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8)
	Rcreate(tag uint16, qid Qid, iouint uint32)
	Tread(tag uint16, fid uint32, offset uint64, count uint32)
	Rread(tag uint16, data []byte)
	Twrite(tag uint16, fid uint32, offset uint64, data []byte)
	Rwrite(tag uint16, count uint32)
	Tclunk(tag uint16, fid uint32)
	Rclunk(tag uint16)
	Tremove(tag uint16, fid uint32)
	Rremove(tag uint16)
	Tstat(tag uint16, fid uint32)
	Rstat(tag uint16, stat Stat)
	Twstat(tag uint16, fid uint32, stat Stat)
	Rwstat(tag uint16)

	Rerrorf(tag uint16, format string, args ...interface{})
	Rerror(tag uint16, err error)

	SetMaxMessageSize(size uint32)
	MaxMessageSize() uint32

	Reset(w io.Writer)
	Err() error
	Flush() error
}

const defBufSize = 4096

type encoder struct {
	w   *bufio.Writer
	err error
	buf [20]byte

	maxSize  int64
	dataSize int64

	logger Logger
}

// NewEncoder returns a new encoder that will transmit on the io.Writer.
func NewEncoder(w io.Writer, opts ...Option) Encoder {
	return newEncoder(w, opts...)
}

func newEncoder(w io.Writer, opts ...Option) *encoder {
	e := &encoder{
		maxSize:  defaultMaxMessageLen, // TODO: unused
		dataSize: defaultMaxDataLen,    // TODO: unused
	}
	for _, opt := range opts {
		opt(e)
	}
	e.reset(w, defBufSize)
	return e
}

func (e *encoder) reset(w io.Writer, size int) {
	e.w = bufio.NewWriterSize(w, size)
	e.err = nil
}

func (e *encoder) Reset(w io.Writer) { e.reset(w, defBufSize) }

func (e *encoder) Err() error { return e.err }

func (e *encoder) SetMaxMessageSize(size uint32) {
	e.maxSize = int64(size)
}

func (e *encoder) MaxMessageSize() uint32 {
	return uint32(e.maxSize)
}

func (e *encoder) Tversion(tag uint16, msize uint32, version string) {
	size := minSizeLUT[msgTversion-100] + uint32(len(version))
	e.pheader(size, msgTversion, tag)
	e.puint32(msize)
	e.pstring(version)
	e.printf("<- Tversion tag:%d msize:%d version:%q", tag, msize, version)
}

func (e *encoder) Rversion(tag uint16, msize uint32, version string) {
	size := minSizeLUT[msgRversion-100] + uint32(len(version))
	e.pheader(size, msgRversion, tag)
	e.puint32(msize)
	e.pstring(version)
	e.printf("<- Rversion tag:%d msize:%d version:%q", tag, msize, version)
}

func (e *encoder) Tauth(tag uint16, afid uint32, uname, aname string) {
	size := minSizeLUT[msgTauth-100] + uint32(len(uname)+len(aname))
	e.pheader(size, msgTauth, tag)
	e.puint32(afid)
	e.pstring(uname)
	e.pstring(aname)
	e.printf("<- Tauth tag:%d afid:%d uname:%q aname:%q", tag, afid, uname, aname)
}

func (e *encoder) Rauth(tag uint16, aqid Qid) {
	size := minSizeLUT[msgRauth-100]
	e.pheader(size, msgRauth, tag)
	e.pqid(aqid)
	e.printf("<- Rauth tag:%d qid:[%q]", tag, aqid)
}

func (e *encoder) Tattach(tag uint16, fid, afid uint32, uname, aname string) {
	size := minSizeLUT[msgTattach-100] + uint32(len(uname)+len(aname))
	e.pheader(size, msgTattach, tag)
	e.puint32(fid)
	e.puint32(afid)
	e.pstring(uname)
	e.pstring(aname)
	e.printf("<- Tattach tag:%d fid:%d afid:%d uname:%q aname:%q", tag, fid, afid, uname, aname)
}

func (e *encoder) Rattach(tag uint16, qid Qid) {
	size := minSizeLUT[msgRattach-100]
	e.pheader(size, msgRattach, tag)
	e.pqid(qid) // TODO: entries check
	e.printf("<- Rattach tag:%d qid:[%q]", tag, qid)
}

func (e *encoder) Tflush(tag, oldtag uint16) {
	size := minSizeLUT[msgTflush-100]
	e.pheader(size, msgTflush, tag)
	e.puint16(oldtag)
	e.printf("<- Tflush tag:%d oldtag:%d", tag, oldtag)
}

func (e *encoder) Rflush(tag uint16) {
	size := minSizeLUT[msgRflush-100]
	e.pheader(size, msgRflush, tag)
	e.printf("<- Rflush tag:%d", tag)
}

func (e *encoder) Twalk(tag uint16, fid, newfid uint32, names ...string) {
	size := minSizeLUT[msgTwalk-100]
	for _, n := range names {
		size += 2 + uint32(len(n))
	}

	e.pheader(uint32(size), msgTwalk, tag)
	e.puint32(fid)
	e.puint32(newfid)

	puint16(e.buf[:2], uint16(len(names))) // TODO: entries check
	e.write(e.buf[:2])
	for _, n := range names {
		e.pstring(n)
	}

	e.printf("<- Twalk tag:%d fid:%d newfid:%d wname:%q", tag, fid, newfid, names)
}

func (e *encoder) Rwalk(tag uint16, qids ...Qid) {
	size := minSizeLUT[msgRwalk-100] + uint32(13*len(qids))
	e.pheader(size, msgRwalk, tag)

	puint16(e.buf[:2], uint16(len(qids)))
	e.write(e.buf[:2])
	for _, q := range qids {
		e.pqid(q)
	}

	e.printf("<- Rwalk tag:%d wqid:%q", tag, qids)
}

func (e *encoder) Topen(tag uint16, fid uint32, mode uint8) {
	size := minSizeLUT[msgTopen-100]
	e.pheader(size, msgTopen, tag)
	e.puint32(fid)
	e.puint8(mode)
	e.printf("<- Topen tag:%d fid:%d mode:%d", tag, fid, mode)
}

func (e *encoder) Ropen(tag uint16, qid Qid, iounit uint32) {
	size := minSizeLUT[msgRopen-100]
	e.pheader(size, msgRopen, tag)
	e.pqid(qid)
	e.puint32(iounit)
	e.printf("<- Ropen tag:%d qid:%q iounit:%d", tag, qid, iounit)
}

func (e *encoder) Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) {
	size := minSizeLUT[msgTcreate-100] + uint32(len(name))
	e.pheader(size, msgTcreate, tag)
	e.puint32(fid)
	e.pstring(name)
	e.puint32(perm)
	e.puint8(mode)
	e.printf("<- Tcreate tag:%d fid:%d name:%q perm:%d mode:%d", tag, fid, name, perm, mode)
}

func (e *encoder) Rcreate(tag uint16, qid Qid, iounit uint32) {
	size := minSizeLUT[msgRcreate-100]
	e.pheader(size, msgRcreate, tag)
	e.pqid(qid)
	e.puint32(iounit)
	e.printf("<- Rcreate tag:%d qid:%q iounit:%d", tag, qid, iounit)
}

func (e *encoder) Tread(tag uint16, fid uint32, offset uint64, count uint32) {
	size := minSizeLUT[msgTread-100]
	e.pheader(size, msgTread, tag)
	e.puint32(fid)
	e.puint64(offset)
	e.puint32(count)
	e.printf("<- Tread tag:%d fid:%d offset:%d count:%d", tag, fid, offset, count)
}

func (e *encoder) Rread(tag uint16, data []byte) {
	size := minSizeLUT[msgRread-100] + uint32(len(data))
	e.pheader(size, msgRread, tag)
	e.pdata(data)
	e.printf("<- Rread tag:%d count:%d", tag, len(data))
}

func (e *encoder) Twrite(tag uint16, fid uint32, offset uint64, data []byte) {
	size := minSizeLUT[msgTwrite-100] + uint32(len(data))
	e.pheader(size, msgTwrite, tag)
	e.puint32(fid)
	e.puint64(offset)
	e.pdata(data)
	e.printf("<- Twrite tag:%d fid:%d offset:%d count:%d", tag, fid, offset, len(data))
}

func (e *encoder) Rwrite(tag uint16, count uint32) {
	size := minSizeLUT[msgRwrite-100]
	e.pheader(size, msgRwrite, tag)
	e.puint32(count)
	e.printf("<- Rwrite tag:%d count:%d", tag, count)
}

func (e *encoder) Tclunk(tag uint16, fid uint32) {
	size := minSizeLUT[msgTclunk-100]
	e.pheader(size, msgTclunk, tag)
	e.puint32(fid)
	e.printf("<- Tclunk tag:%d fid:%d", tag, fid)
}

func (e *encoder) Rclunk(tag uint16) {
	size := minSizeLUT[msgRclunk-100]
	e.pheader(size, msgRclunk, tag)
	e.printf("<- Rclunk tag:%d", tag)
}

func (e *encoder) Tremove(tag uint16, fid uint32) {
	size := minSizeLUT[msgTremove-100]
	e.pheader(size, msgTremove, tag)
	e.puint32(fid)
	e.printf("<- Tremove tag:%d fid:%d", tag, fid)
}

func (e *encoder) Rremove(tag uint16) {
	size := minSizeLUT[msgRremove-100]
	e.pheader(size, msgRremove, tag)
	e.printf("<- Rremove tag:%d", tag)
}

func (e *encoder) Tstat(tag uint16, fid uint32) {
	size := minSizeLUT[msgTstat-100]
	e.pheader(size, msgTstat, tag)
	e.puint32(fid)
	e.printf("<- Tstat tag:%d fid:%d", tag, fid)
}

func (e *encoder) Rstat(tag uint16, stat Stat) {
	strLen := len(stat.Name) + len(stat.Uid) + len(stat.Gid) + len(stat.Muid)
	statLen := uint16(fixedStatLen + strLen)
	size := minSizeLUT[msgRstat-100] + uint32(strLen)
	e.pheader(size, msgRstat, tag)
	e.puint16(2 + statLen)
	e.pstat(stat, statLen)
	e.printf("<- Rstat tag:%d stat:\"%s\"", tag, stat)
}

func (e *encoder) Twstat(tag uint16, fid uint32, stat Stat) {
	strLen := len(stat.Name) + len(stat.Uid) + len(stat.Gid) + len(stat.Muid)
	statLen := uint16(fixedStatLen + strLen)
	size := minSizeLUT[msgTwstat-100] + uint32(strLen)
	e.pheader(size, msgTwstat, tag)
	e.puint32(fid)
	e.puint16(2 + statLen)
	e.pstat(stat, statLen)
	e.printf("<- Twstat tag:%d fid:%d stat:\"%s\"", tag, fid, stat)
}

func (e *encoder) Rwstat(tag uint16) {
	size := minSizeLUT[msgRwstat-100]
	e.pheader(size, msgRwstat, tag)
	e.printf("<- Rwstat tag:%d", tag)
}

func (e *encoder) Rerrorf(tag uint16, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	e.rerror(tag, msg)
}

func (e *encoder) Rerror(tag uint16, err error) {
	e.rerror(tag, err.Error())
}

func (e *encoder) rerror(tag uint16, msg string) {
	size := minSizeLUT[msgRerror-100] + uint32(len(msg))
	e.pheader(size, msgRerror, tag)
	e.pstring(msg)
	e.printf("<- Rerror tag:%d ename:%q", tag, msg)
}

func (e *encoder) pheader(size uint32, typ uint8, tag uint16) {
	puint32(e.buf[:4], size)
	e.buf[4] = typ
	puint16(e.buf[5:7], tag)
	e.write(e.buf[:7])
}

func (e *encoder) pstring(v string) {
	puint16(e.buf[:2], uint16(len(v)))
	e.write(e.buf[:2])
	e.writeString(v)
}

func (e *encoder) pdata(v []byte) {
	puint32(e.buf[:4], uint32(len(v)))
	e.write(e.buf[:4])
	e.write(v)
}

func (e *encoder) pqid(v Qid) {
	e.buf[0] = v.Type
	puint32(e.buf[1:5], v.Version)
	puint64(e.buf[5:13], v.Path)
	e.write(e.buf[:13])
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
	e.pstring(v.Uid)
	e.pstring(v.Gid)
	e.pstring(v.Muid)
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

func (e *encoder) writeString(p string) {
	if e.err != nil || len(p) == 0 {
		return
	}
	_, e.err = e.w.WriteString(p)
}

func (e *encoder) write(p []byte) {
	if e.err != nil || len(p) == 0 {
		return
	}
	_, e.err = e.w.Write(p)
}

func (e *encoder) Flush() error {
	if e.err != nil {
		return e.err
	}
	e.err = e.w.Flush()
	return e.err
}

func (e *encoder) printf(format string, args ...interface{}) {
	if e.logger != nil {
		e.logger.Printf(format, args...)
	}
}
