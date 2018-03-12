package ninep

import (
	"io"
	"sync"

	"github.com/azmodb/ninep/proto"
)

type encoder struct {
	writer sync.Mutex // exclusive connection encoder
	proto.Encoder
}

// newEncoder returns a new thread-safe encoder that will transmit on the
// io.Writer.
func newEncoder(w io.Writer, opts ...proto.Option) *encoder {
	return &encoder{Encoder: proto.NewEncoder(w, opts...)}
}

func (e *encoder) Err() error {
	e.writer.Lock()
	err := e.Encoder.Err()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rerrorf(tag uint16, fmt string, args ...interface{}) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rerrorf(tag, fmt, args...)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rerror(tag uint16, err error) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rerror(tag, err)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rauth(tag uint16, qid proto.Qid) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rauth(tag, qid)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rattach(tag uint16, qid proto.Qid) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rattach(tag, qid)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rflush(tag uint16) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rflush(tag)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rwalk(tag uint16, qids ...proto.Qid) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rwalk(tag, qids...)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Ropen(tag uint16, qid proto.Qid, iounit uint32) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Ropen(tag, qid, iounit)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rcreate(tag uint16, qid proto.Qid, iounit uint32) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rcreate(tag, qid, iounit)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rread(tag uint16, data []byte) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rread(tag, data)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rwrite(tag uint16, count uint32) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rwrite(tag, count)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rclunk(tag uint16) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rclunk(tag)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rremove(tag uint16) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rremove(tag)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rstat(tag uint16, stat proto.Stat) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rstat(tag, stat)
	e.Flush()
	e.writer.Unlock()
}

func (e *encoder) Rwstat(tag uint16) {
	e.writer.Lock()
	if e.Err() != nil {
		e.writer.Unlock()
		return
	}
	e.Rwstat(tag)
	e.Flush()
	e.writer.Unlock()
}
