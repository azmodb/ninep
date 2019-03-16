package ninep

import (
	"context"

	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/log"
)

type Fid struct {
	c   *Client
	num uint32

	//	mu     sync.Mutex // protects following
	qid    proto.Qid
	iounit uint32
}

func newFid(c *Client, num uint32) *Fid {
	return &Fid{c: c, num: num}
}

func (f *Fid) Close() error {
	tag, fcall, err := f.c.register(false)
	if err != nil {
		return err
	}
	defer func() {
		if f.num != 0 && f.num != proto.NoFid {
			f.c.fidpool.Put(f.num)
		}
		fcall.reset()
	}()

	log.Debugf("<- Tclunk tag:%d fid:%d", tag, f.num)
	if err = f.c.enc.Tclunk(tag, f.num); err != nil {
		return err
	}
	if err = fcall.wait(context.TODO()); err != nil {
		return err
	}
	log.Debugf("-> Rclunk tag%d", tag)
	return err
}

func (f *Fid) Open(flag int) error {
	tag, fcall, err := f.c.register(false)
	if err != nil {
		return err
	}
	defer fcall.reset()

	mode := uint8(0)

	log.Debugf("<- Topen tag:%d fid:%d mode:%s", tag, f.num, mode)
	if err = f.c.enc.Topen(tag, f.num, mode); err != nil {
		return err
	}
	if err = fcall.wait(context.TODO()); err != nil {
		return err
	}

	typ, version, path, iounit := fcall.buf.Ropen()
	if err = fcall.buf.Err(); err != nil {
		return err
	}
	f.qid = proto.Qid{Type: typ, Version: version, Path: path}
	f.iounit = iounit
	log.Debugf("-> Ropen tag:%d qid:(%s) iounit:%d", tag, f.qid, iounit)
	return err
}
