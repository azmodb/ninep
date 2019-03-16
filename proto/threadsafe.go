package proto

import (
	"io"
	"sync"
)

// Encoder writes 9P2000 messages to the underlying output stream. If an
// error occures, encoding to a Encoder, no more data will be accepted
// and all subsequent calls will return the error. After all data has
// been written, the client should call the Flush method to guarantee
// all data has been forwarded to the underlying io.Writer.
//
// Encoder is safe for concurrent use.
type Encoder struct {
	mu  sync.Mutex // exclusive writer lock
	enc *BufferedEncoder
}

// NewEncoder returns a new encoder that will transmit on the
// io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{enc: NewBufferedEncoder(w)}
}

// SetMaxMessageSize sets the maximum message size that a Encoder will
// accept.
func (e *Encoder) SetMaxMessageSize(size uint32) {
	e.mu.Lock()
	e.enc.MaxMessageSize = int64(size)
	e.mu.Unlock()
}

// Reset discards any unflushed buffered data, clears any error, and
// resets b to write its output to w.
func (e *Encoder) Reset(w io.Writer) {
	e.mu.Lock()
	e.enc.Reset(w)
	e.mu.Unlock()
}

// Tversion writes a new Tversion message to the underlying io.Writer.
func (e *Encoder) Tversion(tag uint16, msize int64, version string) error {
	e.mu.Lock()
	e.enc.Tversion(tag, msize, version)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rversion writes a new Rversion message to the underlying io.Writer.
func (e *Encoder) Rversion(tag uint16, msize int64, version string) error {
	e.mu.Lock()
	e.enc.Rversion(tag, msize, version)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tauth writes a new Tauth message to the underlying io.Writer.
func (e *Encoder) Tauth(tag uint16, afid uint32, uname, aname string) error {
	e.mu.Lock()
	e.enc.Tauth(tag, afid, uname, aname)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rauth writes a new Rauth message to the underlying io.Writer.
func (e *Encoder) Rauth(tag uint16, typ uint8, version uint32, path uint64) error {
	e.mu.Lock()
	e.enc.Rauth(tag, typ, version, path)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tattach writes a new Tattach message to the underlying io.Writer.
func (e *Encoder) Tattach(tag uint16, fid, afid uint32, uname, aname string) error {
	e.mu.Lock()
	e.enc.Tattach(tag, fid, afid, uname, aname)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rattach writes a new Rattach message to the underlying io.Writer.
func (e *Encoder) Rattach(tag uint16, typ uint8, version uint32, path uint64) error {
	e.mu.Lock()
	e.enc.Rattach(tag, typ, version, path)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rerror writes a new Rerror message to the underlying io.Writer.
func (e *Encoder) Rerror(tag uint16, ename string) error {
	e.mu.Lock()
	e.enc.Rerror(tag, ename)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tflush writes a new Tflush message to the underlying io.Writer.
func (e *Encoder) Tflush(tag, oldtag uint16) error {
	e.mu.Lock()
	e.enc.Tflush(tag, oldtag)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rflush writes a new Rflush message to the underlying io.Writer.
func (e *Encoder) Rflush(tag uint16) error {
	e.mu.Lock()
	e.enc.Rflush(tag)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Twalk writes a new Twalk message to the underlying io.Writer.
func (e *Encoder) Twalk(tag uint16, fid, newfid uint32, names ...string) error {
	e.mu.Lock()
	e.enc.Twalk(tag, fid, newfid, names...)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rwalk writes a new Rwalk message to the underlying io.Writer.
func (e *Encoder) Rwalk(tag uint16, qids ...Qid) error {
	e.mu.Lock()
	e.enc.Rwalk(tag, qids...)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Topen writes a new Topen message to the underlying io.Writer.
func (e *Encoder) Topen(tag uint16, fid uint32, mode uint8) error {
	e.mu.Lock()
	e.enc.Topen(tag, fid, mode)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Ropen writes a new Ropen message to the underlying io.Writer.
func (e *Encoder) Ropen(tag uint16, typ uint8, version uint32, path uint64, iounit uint32) error {
	e.mu.Lock()
	e.enc.Ropen(tag, typ, version, path, iounit)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tcreate writes a new Tcreate message to the underlying io.Writer.
func (e *Encoder) Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) error {
	e.mu.Lock()
	e.enc.Tcreate(tag, fid, name, perm, mode)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rcreate writes a new Rcreate message to the underlying io.Writer.
func (e *Encoder) Rcreate(tag uint16, typ uint8, version uint32, path uint64, iounit uint32) error {
	e.mu.Lock()
	e.enc.Rcreate(tag, typ, version, path, iounit)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tread writes a new Tread message to the underlying io.Writer.
func (e *Encoder) Tread(tag uint16, fid uint32, offset uint64, count uint32) error {
	e.mu.Lock()
	e.enc.Tread(tag, fid, offset, count)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rread writes a new Rread message to the underlying io.Writer.
func (e *Encoder) Rread(tag uint16, data []byte) error {
	e.mu.Lock()
	e.enc.Rread(tag, data)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Twrite writes a new Twrite message to the underlying io.Writer.
func (e *Encoder) Twrite(tag uint16, fid uint32, offset uint64, data []byte) error {
	e.mu.Lock()
	e.enc.Twrite(tag, fid, offset, data)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rwrite writes a new Rwrite message to the underlying io.Writer.
func (e *Encoder) Rwrite(tag uint16, count uint32) error {
	e.mu.Lock()
	e.enc.Rwrite(tag, count)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tclunk writes a new Tclunk message to the underlying io.Writer.
func (e *Encoder) Tclunk(tag uint16, fid uint32) error {
	e.mu.Lock()
	e.enc.Tclunk(tag, fid)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rclunk writes a new Rclunk message to the underlying io.Writer.
func (e *Encoder) Rclunk(tag uint16) error {
	e.mu.Lock()
	e.enc.Rclunk(tag)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tremove writes a new Tremove message to the underlying io.Writer.
func (e *Encoder) Tremove(tag uint16, fid uint32) error {
	e.mu.Lock()
	e.enc.Tremove(tag, fid)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rremove writes a new Rremove message to the underlying io.Writer.
func (e *Encoder) Rremove(tag uint16) error {
	e.mu.Lock()
	e.enc.Rremove(tag)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Tstat writes a new Tstat message to the underlying io.Writer.
func (e *Encoder) Tstat(tag uint16, fid uint32) error {
	e.mu.Lock()
	e.enc.Tstat(tag, fid)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rstat writes a new Rstat message to the underlying io.Writer.
func (e *Encoder) Rstat(tag uint16, data []byte) error {
	e.mu.Lock()
	e.enc.Rstat(tag, data)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Twstat writes a new Twstat message to the underlying io.Writer.
func (e *Encoder) Twstat(tag uint16, fid uint32, data []byte) error {
	e.mu.Lock()
	e.enc.Twstat(tag, fid, data)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}

// Rwstat writes a new Rwstat message to the underlying io.Writer.
func (e *Encoder) Rwstat(tag uint16) error {
	e.mu.Lock()
	e.enc.Rwstat(tag)
	err := e.enc.Flush()
	e.mu.Unlock()
	return err
}
