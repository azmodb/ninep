package ninep

import (
	"sync"

	"github.com/azmodb/ninep/proto"
)

type encoder struct {
	writer sync.Mutex // exclusive writer lock
	e      *proto.Encoder
}

func (e *encoder) Tversion(tag uint16, msize int64, version string) error {
	e.writer.Lock()
	e.e.Tversion(tag, msize, version)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tauth(tag uint16, afid uint32, uname, aname string) error {
	e.writer.Lock()
	e.e.Tauth(tag, afid, uname, aname)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tattach(tag uint16, fid, afid uint32, uname, aname string) error {
	e.writer.Lock()
	e.e.Tattach(tag, fid, afid, uname, aname)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tflush(tag, oldtag uint16) error {
	e.writer.Lock()
	e.e.Tflush(tag, oldtag)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Twalk(tag uint16, fid, newfid uint32, names ...string) error {
	e.writer.Lock()
	e.e.Twalk(tag, fid, newfid, names...)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Topen(tag uint16, fid uint32, mode uint8) error {
	e.writer.Lock()
	e.e.Topen(tag, fid, mode)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tcreate(tag uint16, fid uint32, name string, perm uint32, mode uint8) error {
	e.writer.Lock()
	e.e.Tcreate(tag, fid, name, perm, mode)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tread(tag uint16, fid uint32, offset uint64, count uint32) error {
	e.writer.Lock()
	e.e.Tread(tag, fid, offset, count)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Twrite(tag uint16, fid uint32, offset uint64, data []byte) error {
	e.writer.Lock()
	e.e.Twrite(tag, fid, offset, data)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tclunk(tag uint16, fid uint32) error {
	e.writer.Lock()
	e.e.Tclunk(tag, fid)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tremove(tag uint16, fid uint32) error {
	e.writer.Lock()
	e.e.Tremove(tag, fid)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Tstat(tag uint16, fid uint32) error {
	e.writer.Lock()
	e.e.Tstat(tag, fid)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Twstat(tag uint16, fid uint32, data []byte) error {
	e.writer.Lock()
	e.e.Twstat(tag, fid, data)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rversion(tag uint16, msize int64, version string) error {
	e.writer.Lock()
	e.e.Rversion(tag, msize, version)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rauth(tag uint16, typ uint8, version uint32, path uint64) error {
	e.writer.Lock()
	e.e.Rauth(tag, typ, version, path)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rattach(tag uint16, typ uint8, version uint32, path uint64) error {
	e.writer.Lock()
	e.e.Rattach(tag, typ, version, path)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rerror(tag uint16, ename string) error {
	e.writer.Lock()
	e.e.Rerror(tag, ename)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rflush(tag uint16) error {
	e.writer.Lock()
	e.e.Rflush(tag)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rwalk(tag uint16, qids ...proto.Qid) error {
	e.writer.Lock()
	e.e.Rwalk(tag, qids...)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Ropen(tag uint16, typ uint8, version uint32, path uint64, iounit uint32) error {
	e.writer.Lock()
	e.e.Ropen(tag, typ, version, path, iounit)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rcreate(tag uint16, typ uint8, version uint32, path uint64, iounit uint32) error {
	e.writer.Lock()
	e.e.Rcreate(tag, typ, version, path, iounit)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rread(tag uint16, data []byte) error {
	e.writer.Lock()
	e.e.Rread(tag, data)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rwrite(tag uint16, count uint32) error {
	e.writer.Lock()
	e.e.Rwrite(tag, count)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rclunk(tag uint16) error {
	e.writer.Lock()
	e.e.Rclunk(tag)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rremove(tag uint16) error {
	e.writer.Lock()
	e.e.Rremove(tag)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rstat(tag uint16, data []byte) error {
	e.writer.Lock()
	e.e.Rstat(tag, data)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}

func (e *encoder) Rwstat(tag uint16) error {
	e.writer.Lock()
	e.e.Rwstat(tag)
	err := e.e.Flush()
	e.writer.Unlock()
	return err
}
