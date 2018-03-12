package ninep

import (
	"sync"

	"github.com/azmodb/ninep/proto"
)

var requestPool = &sync.Pool{New: func() interface{} { return &request{} }}

type request struct {
	enc *encoder
	tag uint16
}

func newRequest(tag uint16, enc *encoder) *request {
	r := requestPool.Get().(*request)
	r.enc = enc
	r.tag = tag
	return r
}

func putRequest(r *request) {
	r.enc = nil
	r.tag = 0
	requestPool.Put(r)
}

func (r *request) Rerror(err error) { r.enc.Rerror(r.tag, err) }

func (r *request) Rerrorf(format string, args ...interface{}) {
	r.enc.Rerrorf(r.tag, format, args...)
}

type Tattach struct {
	proto.Tattach
	*request
}

func (r Tattach) Rattach(qid proto.Qid) {
	r.enc.Rattach(r.tag, qid)
}

type Tauth struct {
	proto.Tauth
	*request
}

func (r Tauth) Rauth(qid proto.Qid) {
	r.enc.Rauth(r.tag, qid)
}

type Twalk struct {
	proto.Twalk
	*request
}

func (r Twalk) Rwalk(qids ...proto.Qid) {
	r.enc.Rwalk(r.tag, qids...)
}

type Topen struct {
	proto.Topen
	*request
}

func (r Topen) Ropen(qid proto.Qid, iounit uint32) {
	r.enc.Ropen(r.tag, qid, iounit)
}

type Tcreate struct {
	proto.Tcreate
	*request
}

func (r Tcreate) Rcreate(qid proto.Qid, iounit uint32) {
	r.enc.Rcreate(r.tag, qid, iounit)
}

type Tread struct {
	proto.Tread
	*request
}

func (r Tread) Rread(data []byte) {
	r.enc.Rread(r.tag, data)
}

type Twrite struct {
	proto.Twrite
	*request
}

func (r Twrite) Rwrite(count uint32) {
	r.enc.Rwrite(r.tag, count)
}

type Tclunk struct {
	proto.Tclunk
	*request
}

func (r Tclunk) Rclunk() { r.enc.Rclunk(r.tag) }

type Tremove struct {
	proto.Tremove
	*request
}

func (r Tremove) Rremove() { r.enc.Rremove(r.tag) }

type Tstat struct {
	proto.Tstat
	*request
}

func (r Tstat) Rstat(stat proto.Stat) {
	r.enc.Rstat(r.tag, stat)
}

type Twstat struct {
	proto.Twstat
	*request
}

func (r Twstat) Rwstat() { r.enc.Rwstat(r.tag) }
