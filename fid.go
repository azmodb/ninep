package ninep

import "github.com/azmodb/ninep/proto"

type Fid struct {
	c   *Client
	num uint32

	qid proto.Qid
}

func newFid(c *Client, num uint32) *Fid {
	return &Fid{c: c, num: num}
}

func (f *Fid) Close() error {
	if f.num != 0 && f.num != proto.NoFid {
		f.c.fidpool.Put(f.num)
	}
	return nil
}
