package proto

import (
	"bufio"
	"io"
)

// BufferedEncoder writes 9P2000 messages to the underlying output stream.
// If an error occures, encoding to a BufferedEncoder, no more data will
// be accepted and all subsequent calls will return the error. After all
// data has been written, the client should call the Flush method to
// guarantee all data has been forwarded to the underlying io.Writer.
//
// Encoder is not safe for concurrent use. Usage of any Encoder method
// should be protected by a mutex.
type BufferedEncoder struct {
	// MaxMessageSize is the maximum size message that a Encoder will
	// accept. If MaxMessageSize is -1, a Encoder will accept any size
	// message.
	MaxMessageSize int64

	w   *bufio.Writer
	buf [8]byte
	err error
}

// NewBufferedEncoder returns a new encoder that will transmit on the
// io.Writer.
func NewBufferedEncoder(w io.Writer) *BufferedEncoder {
	return &BufferedEncoder{
		MaxMessageSize: DefaultMaxMessageSize,
		w:              bufio.NewWriter(w),
	}
}

// Reset discards any unflushed buffered data, clears any error, and
// resets b to write its output to w.
func (e *BufferedEncoder) Reset(w io.Writer) {
	e.w = bufio.NewWriter(w)
	e.err = nil
}

// Err returns the first non-EOF error that was encountered by the
// Encoder.
func (e *BufferedEncoder) Err() error { return e.err }

func (e *BufferedEncoder) setErr(err error) {
	if e.err == nil && err != nil {
		e.err = err
	}
}

func (e *BufferedEncoder) version(typ FcallType, tag uint16, msize int64, version string) {
	if msize > MaxMessageSize {
		e.setErr(ErrMessageTooLarge)
		return
	}
	if e.MaxMessageSize >= 0 && msize > e.MaxMessageSize {
		e.setErr(ErrMessageTooLarge)
		return
	}

	size := fixedSize(typ) + uint32(len(version))
	e.writeHeader(size, typ, tag)
	e.writeUint32(uint32(msize))
	e.writeString(version)
}

// Tversion writes a new Tversion message to the underlying io.Writer.
func (e *BufferedEncoder) Tversion(tag uint16, msize int64, version string) {
	e.version(Tversion, tag, msize, version)
}

// Rversion writes a new Rversion message to the underlying io.Writer.
func (e *BufferedEncoder) Rversion(tag uint16, msize int64, version string) {
	e.version(Rversion, tag, msize, version)
}

// Tauth writes a new Tauth message to the underlying io.Writer.
func (e *BufferedEncoder) Tauth(tag uint16, afid uint32, uname, aname string) {
	if len(aname) > MaxPathSize {
		e.setErr(ErrPathTooLarge)
		return
	}
	if len(uname) > MaxUserNameSize {
		e.setErr(ErrUserNameTooLarge)
		return
	}

	size := fixedSize(Tauth) + uint32(len(uname)) + uint32(len(aname))
	e.writeHeader(size, Tauth, tag)
	e.writeUint32(afid)
	e.writeString(uname)
	e.writeString(aname)
}

// Rauth writes a new Rauth message to the underlying io.Writer.
func (e *BufferedEncoder) Rauth(tag uint16, typ uint8, version uint32, path uint64) {
	e.writeHeader(fixedSize(Rauth), Rauth, tag)
	e.writeUint8(typ)
	e.writeUint32(version)
	e.writeUint64(path)
}

// Tattach writes a new Tattach message to the underlying io.Writer.
func (e *BufferedEncoder) Tattach(tag uint16, fid, afid uint32, uname, aname string) {
	if len(aname) > MaxPathSize {
		e.setErr(ErrPathTooLarge)
		return
	}
	if len(uname) > MaxUserNameSize {
		e.setErr(ErrUserNameTooLarge)
		return
	}

	size := fixedSize(Tattach) + uint32(len(uname)) + uint32(len(aname))
	e.writeHeader(size, Tattach, tag)
	e.writeUint32(fid)
	e.writeUint32(afid)
	e.writeString(uname)
	e.writeString(aname)
}

// Rattach writes a new Rattach message to the underlying io.Writer.
func (e *BufferedEncoder) Rattach(tag uint16, typ uint8, version uint32, path uint64) {
	e.writeHeader(fixedSize(Rattach), Rattach, tag)
	e.writeUint8(typ)
	e.writeUint32(version)
	e.writeUint64(path)
}

// Rerror writes a new Rerror message to the underlying io.Writer.
func (e *BufferedEncoder) Rerror(tag uint16, ename string) {
	if len(ename) > MaxEnameSize {
		e.setErr(ErrEnameTooLarge)
		return
	}

	size := fixedSize(Rerror) + uint32(len(ename))
	e.writeHeader(size, Rerror, tag)
	e.writeString(ename)
}

// Tflush writes a new Tflush message to the underlying io.Writer.
func (e *BufferedEncoder) Tflush(tag, oldtag uint16) {
	e.writeHeader(fixedSize(Tflush), Tflush, tag)
	e.writeUint16(oldtag)
}

// Rflush writes a new Rflush message to the underlying io.Writer.
func (e *BufferedEncoder) Rflush(tag uint16) {
	e.writeHeader(fixedSize(Rflush), Rflush, tag)
}

// Twalk writes a new Twalk message to the underlying io.Writer.
func (e *BufferedEncoder) Twalk(tag uint16, fid, newfid uint32, names ...string) {
	if len(names) > MaxWalkElements {
		e.setErr(ErrWalkElements)
		return
	}

	size := fixedSize(Twalk)
	for _, name := range names {
		size += 2 + uint32(len(name))
	}

	e.writeHeader(size, Twalk, tag)
	e.writeUint32(fid)
	e.writeUint32(newfid)
	e.writeUint16(uint16(len(names)))
	for _, name := range names {
		e.writeString(name)
	}
}

// Rwalk writes a new Rwalk message to the underlying io.Writer.
func (e *BufferedEncoder) Rwalk(tag uint16, qids ...Qid) {
	if len(qids) > MaxWalkElements {
		e.setErr(ErrWalkElements)
		return
	}

	size := fixedSize(Rwalk) + uint32(13*len(qids))
	e.writeHeader(size, Rwalk, tag)
	e.writeUint16(uint16(len(qids)))
	for _, qid := range qids {
		e.writeUint8(qid.Type)
		e.writeUint32(qid.Version)
		e.writeUint64(qid.Path)
	}
}

// Topen writes a new Topen message to the underlying io.Writer.
func (e *BufferedEncoder) Topen(tag uint16, fid uint32, mode uint8) {
	e.writeHeader(fixedSize(Topen), Topen, tag)
	e.writeUint32(fid)
	e.writeUint8(uint8(mode))
}

// Ropen writes a new Ropen message to the underlying io.Writer.
func (e *BufferedEncoder) Ropen(tag uint16, typ uint8, version uint32, path uint64, iounit uint32) {
	e.writeHeader(fixedSize(Ropen), Ropen, tag)
	e.writeUint8(typ)
	e.writeUint32(version)
	e.writeUint64(path)
	e.writeUint32(iounit)
}

// Tcreate writes a new Tcreate message to the underlying io.Writer.
func (e *BufferedEncoder) Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) {
	if len(name) > MaxDirNameSize {
		e.setErr(ErrDirNameTooLarge)
		return
	}

	size := fixedSize(Tcreate) + uint32(len(name))
	e.writeHeader(size, Tcreate, tag)
	e.writeUint32(fid)
	e.writeString(name)
	e.writeUint32(perm)
	e.writeUint8(uint8(mode))
}

// Rcreate writes a new Rcreate message to the underlying io.Writer.
func (e *BufferedEncoder) Rcreate(tag uint16, typ uint8, version uint32, path uint64, iounit uint32) {
	e.writeHeader(fixedSize(Rcreate), Rcreate, tag)
	e.writeUint8(typ)
	e.writeUint32(version)
	e.writeUint64(path)
	e.writeUint32(iounit)
}

// Tread writes a new Tread message to the underlying io.Writer.
func (e *BufferedEncoder) Tread(tag uint16, fid uint32, offset uint64, count uint32) {
	e.writeHeader(fixedSize(Tread), Tread, tag)
	e.writeUint32(fid)
	e.writeUint64(offset)
	e.writeUint32(count)
}

// Rread writes a new Rread message to the underlying io.Writer.
func (e *BufferedEncoder) Rread(tag uint16, data []byte) {
	if len(data) > MaxDataSize { // TODO
		e.setErr(ErrDataTooLarge)
		return
	}

	count := uint32(len(data))
	size := fixedSize(Rread) + count
	e.writeHeader(size, Rread, tag)
	e.writeUint32(count)
	e.write(data)
}

// Twrite writes a new Twrite message to the underlying io.Writer.
func (e *BufferedEncoder) Twrite(tag uint16, fid uint32, offset uint64, data []byte) {
	if len(data) > MaxDataSize { // TODO
		e.setErr(ErrDataTooLarge)
		return
	}

	count := uint32(len(data))
	size := fixedSize(Twrite) + count
	e.writeHeader(size, Twrite, tag)
	e.writeUint32(fid)
	e.writeUint64(offset)
	e.writeUint32(count)
	e.write(data)
}

// Rwrite writes a new Rwrite message to the underlying io.Writer.
func (e *BufferedEncoder) Rwrite(tag uint16, count uint32) {
	e.writeHeader(fixedSize(Rwrite), Rwrite, tag)
	e.writeUint32(count)
}

// Tclunk writes a new Tclunk message to the underlying io.Writer.
func (e *BufferedEncoder) Tclunk(tag uint16, fid uint32) {
	e.writeHeader(fixedSize(Tclunk), Tclunk, tag)
	e.writeUint32(fid)
}

// Rclunk writes a new Rclunk message to the underlying io.Writer.
func (e *BufferedEncoder) Rclunk(tag uint16) {
	e.writeHeader(fixedSize(Rclunk), Rclunk, tag)
}

// Tremove writes a new Tremove message to the underlying io.Writer.
func (e *BufferedEncoder) Tremove(tag uint16, fid uint32) {
	e.writeHeader(fixedSize(Tremove), Tremove, tag)
	e.writeUint32(fid)
}

// Rremove writes a new Rremove message to the underlying io.Writer.
func (e *BufferedEncoder) Rremove(tag uint16) {
	e.writeHeader(fixedSize(Rremove), Rremove, tag)
}

// Tstat writes a new Tstat message to the underlying io.Writer.
func (e *BufferedEncoder) Tstat(tag uint16, fid uint32) {
	e.writeHeader(fixedSize(Tstat), Tstat, tag)
	e.writeUint32(fid)
}

// Rstat writes a new Rstat message to the underlying io.Writer.
func (e *BufferedEncoder) Rstat(tag uint16, data []byte) {
	size := len(data)
	if size > maxStatSize {
		e.setErr(ErrStatTooLarge)
		return
	}

	e.writeHeader(fixedSize(Rstat)+uint32(size), Rstat, tag)
	e.writeUint16(uint16(size))
	e.write(data)
}

// Twstat writes a new Twstat message to the underlying io.Writer.
func (e *BufferedEncoder) Twstat(tag uint16, fid uint32, data []byte) {
	size := len(data)
	if size > maxStatSize {
		e.setErr(ErrStatTooLarge)
		return
	}

	e.writeHeader(fixedSize(Twstat)+uint32(size), Twstat, tag)
	e.writeUint32(fid)
	e.writeUint16(uint16(size))
	e.write(data)
}

// Rwstat writes a new Rwstat message to the underlying io.Writer.
func (e *BufferedEncoder) Rwstat(tag uint16) {
	e.writeHeader(fixedSize(Rwstat), Rwstat, tag)
}

func (e *BufferedEncoder) writeHeader(size uint32, typ FcallType, tag uint16) {
	order.PutUint32(e.buf[:4], size)
	e.buf[4] = byte(typ)
	order.PutUint16(e.buf[5:7], tag)
	e.write(e.buf[:7])
}

func (e *BufferedEncoder) writeString(v string) {
	if e.err != nil {
		return
	}

	order.PutUint16(e.buf[:2], uint16(len(v)))
	if _, e.err = e.w.Write(e.buf[:2]); e.err != nil {
		return
	}
	_, e.err = io.WriteString(e.w, v)
}

func (e *BufferedEncoder) write(buf []byte) {
	if e.err != nil {
		return
	}
	_, e.err = e.w.Write(buf)
}

func (e *BufferedEncoder) writeUint64(v uint64) {
	order.PutUint64(e.buf[:8], v)
	e.write(e.buf[:8])
}

func (e *BufferedEncoder) writeUint32(v uint32) {
	order.PutUint32(e.buf[:4], v)
	e.write(e.buf[:4])
}

func (e *BufferedEncoder) writeUint16(v uint16) {
	order.PutUint16(e.buf[:2], v)
	e.write(e.buf[:2])
}

func (e *BufferedEncoder) writeUint8(v uint8) {
	e.buf[0] = v
	e.write(e.buf[:1])
}

// Flush writes any buffered data to the underlying io.Writer.
func (e *BufferedEncoder) Flush() (err error) {
	if err = e.err; err != nil {
		return err
	}
	if err = e.w.Flush(); err != nil {
		e.setErr(err)
	}
	return err
}
