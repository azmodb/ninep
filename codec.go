package ninep

import (
	"bufio"
	"io"
	"sync"

	"github.com/azmodb/ninep/proto"
)

func newDecoder(r io.Reader, opts ...proto.Option) proto.Decoder {
	return proto.NewDecoder(bufio.NewReader(r), opts...)
}

type encoder struct {
	mu  sync.Mutex // exclusive stream encoder
	enc proto.Encoder
	bw  *bufio.Writer
}

func newEncoder(w io.Writer, opts ...proto.Option) *encoder {
	bw := bufio.NewWriter(w)
	return &encoder{
		enc: proto.NewEncoder(bw, opts...),
		bw:  bw,
	}
}

func (e *encoder) flush(err error) error {
	if err != nil {
		return err
	}
	return e.bw.Flush()
}

func (e *encoder) Tversion(tag uint16, msize uint32, version string) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tversion(tag, msize, version))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rversion(tag uint16, msize uint32, version string) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rversion(tag, msize, version))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tauth(tag uint16, afid uint32, uname, aname string) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tauth(tag, afid, uname, aname))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rauth(tag uint16, qid proto.Qid) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rauth(tag, qid))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tattach(tag uint16, fid, afid uint32, uname, aname string) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tattach(tag, fid, afid, uname, aname))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rattach(tag uint16, qid proto.Qid) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rattach(tag, qid))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tflush(tag, oldtag uint16) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tflush(tag, oldtag))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rflush(tag uint16) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rflush(tag))
	e.mu.Unlock()
	return err
}

func (e *encoder) Twalk(tag uint16, fid, newfid uint32, names ...string) error {
	e.mu.Lock()
	err := e.flush(e.enc.Twalk(tag, fid, newfid, names...))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rwalk(tag uint16, qids ...proto.Qid) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rwalk(tag, qids...))
	e.mu.Unlock()
	return err
}

func (e *encoder) Topen(tag uint16, fid uint32, mode uint8) error {
	e.mu.Lock()
	err := e.flush(e.enc.Topen(tag, fid, mode))
	e.mu.Unlock()
	return err
}

func (e *encoder) Ropen(tag uint16, qid proto.Qid, iounit uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Ropen(tag, qid, iounit))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tcreate(tag, fid, name, perm, mode))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rcreate(tag uint16, qid proto.Qid, iounit uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rcreate(tag, qid, iounit))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tread(tag uint16, fid uint32, offset uint64, count uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tread(tag, fid, offset, count))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rread(tag uint16, data []byte) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rread(tag, data))
	e.mu.Unlock()
	return err
}

func (e *encoder) Twrite(tag uint16, fid uint32, offset uint64, data []byte) error {
	e.mu.Lock()
	err := e.flush(e.enc.Twrite(tag, fid, offset, data))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rwrite(tag uint16, count uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rwrite(tag, count))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tclunk(tag uint16, fid uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tclunk(tag, fid))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rclunk(tag uint16) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rclunk(tag))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tremove(tag uint16, fid uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tremove(tag, fid))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rremove(tag uint16) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rremove(tag))
	e.mu.Unlock()
	return err
}

func (e *encoder) Tstat(tag uint16, fid uint32) error {
	e.mu.Lock()
	err := e.flush(e.enc.Tstat(tag, fid))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rstat(tag uint16, stat proto.Stat) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rstat(tag, stat))
	e.mu.Unlock()
	return err
}

func (e *encoder) Twstat(tag uint16, fid uint32, stat proto.Stat) error {
	e.mu.Lock()
	err := e.flush(e.enc.Twstat(tag, fid, stat))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rwstat(tag uint16) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rwstat(tag))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rerrorf(tag uint16, format string, args ...interface{}) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rerrorf(tag, format, args...))
	e.mu.Unlock()
	return err
}

func (e *encoder) Rerror(tag uint16, perror error) error {
	e.mu.Lock()
	err := e.flush(e.enc.Rerror(tag, perror))
	e.mu.Unlock()
	return err
}

func (e *encoder) SetMaxMessageSize(size uint32) {
	e.mu.Lock()
	e.enc.SetMaxMessageSize(size)
	e.mu.Unlock()
}

func (e *encoder) MaxMessageSize() uint32 {
	e.mu.Lock()
	size := e.enc.MaxMessageSize()
	e.mu.Unlock()
	return size
}
