package proto

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/azmodb/ninep/proto/internal/plan9"
)

var testFcalls = []plan9.Fcall{
	plan9.Fcall{Type: plan9.Tversion, Tag: plan9.NOTAG, Msize: math.MaxUint32, Version: maxVersionStr},
	plan9.Fcall{Type: plan9.Rversion, Tag: plan9.NOTAG, Msize: math.MaxUint32, Version: maxVersionStr},
	plan9.Fcall{Type: plan9.Tversion, Tag: plan9.NOTAG, Msize: 8192, Version: "9P2000"},
	plan9.Fcall{Type: plan9.Rversion, Tag: plan9.NOTAG, Msize: 8192, Version: "9P2000"},

	plan9.Fcall{Type: plan9.Tauth, Tag: plan9.NOTAG, Afid: plan9.NOFID, Uname: maxUnameStr,
		Aname: maxNameStr},
	plan9.Fcall{Type: plan9.Tauth, Tag: plan9.NOTAG, Afid: plan9.NOFID, Uname: "bootes", Aname: "tmp"},
	plan9.Fcall{Type: plan9.Rauth, Tag: plan9.NOTAG, Aqid: plan9Qids[0]},
	plan9.Fcall{Type: plan9.Rauth, Tag: plan9.NOTAG, Aqid: plan9Qids[1]},
	plan9.Fcall{Type: plan9.Rauth, Tag: plan9.NOTAG, Aqid: plan9Qids[2]},
	plan9.Fcall{Type: plan9.Tattach, Tag: plan9.NOTAG, Fid: 1, Afid: plan9.NOFID, Uname: maxUnameStr,
		Aname: maxNameStr},
	plan9.Fcall{Type: plan9.Tattach, Tag: plan9.NOTAG, Fid: 1, Afid: plan9.NOFID, Uname: "bootes",
		Aname: "tmp"},
	plan9.Fcall{Type: plan9.Rattach, Tag: plan9.NOTAG, Qid: plan9Qids[0]},
	plan9.Fcall{Type: plan9.Rattach, Tag: plan9.NOTAG, Qid: plan9Qids[1]},
	plan9.Fcall{Type: plan9.Rattach, Tag: plan9.NOTAG, Qid: plan9Qids[2]},

	plan9.Fcall{Type: plan9.Tflush, Tag: 42, Oldtag: 41},
	plan9.Fcall{Type: plan9.Tflush, Tag: math.MaxUint16, Oldtag: math.MaxUint16},
	plan9.Fcall{Type: plan9.Rflush, Tag: 42},
	plan9.Fcall{Type: plan9.Rflush, Tag: math.MaxUint16},

	plan9.Fcall{Type: plan9.Twalk, Tag: 42, Fid: 1, Newfid: 2, Wname: []string{"usr", "foo", "d"}},
	plan9.Fcall{Type: plan9.Twalk, Tag: 2, Fid: plan9.NOFID - 1, Newfid: plan9.NOFID,
		Wname: []string{
			maxNameStr, maxNameStr, maxNameStr, maxNameStr,
			maxNameStr, maxNameStr, maxNameStr, maxNameStr,
			maxNameStr, maxNameStr, maxNameStr, maxNameStr,
			maxNameStr, maxNameStr, maxNameStr, maxNameStr,
		},
	},
	plan9.Fcall{Type: plan9.Rwalk, Tag: 42, Wqid: []plan9.Qid{plan9Qids[1], plan9Qids[1]}},
	plan9.Fcall{Type: plan9.Rwalk, Tag: plan9.NOTAG,
		Wqid: []plan9.Qid{
			plan9Qids[0], plan9Qids[0], plan9Qids[0], plan9Qids[0],
			plan9Qids[0], plan9Qids[0], plan9Qids[0], plan9Qids[0],
			plan9Qids[0], plan9Qids[0], plan9Qids[0], plan9Qids[0],
			plan9Qids[0], plan9Qids[0], plan9Qids[0], plan9Qids[0],
		},
	},

	plan9.Fcall{Type: plan9.Topen, Tag: 42, Fid: 1, Mode: plan9.OREAD},
	plan9.Fcall{Type: plan9.Topen, Tag: plan9.NOTAG, Fid: plan9.NOFID, Mode: math.MaxUint8},
	plan9.Fcall{Type: plan9.Ropen, Tag: 42, Qid: plan9Qids[1], Iounit: 8192},
	plan9.Fcall{Type: plan9.Ropen, Tag: plan9.NOTAG, Qid: plan9Qids[0], Iounit: math.MaxUint32},
	plan9.Fcall{Type: plan9.Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0644, Mode: plan9.OREAD},
	plan9.Fcall{Type: plan9.Tcreate, Tag: plan9.NOTAG, Fid: plan9.NOFID, Name: maxNameStr,
		Perm: math.MaxUint32, Mode: math.MaxUint8},
	plan9.Fcall{Type: plan9.Rcreate, Tag: 42, Qid: plan9Qids[1], Iounit: 8192},
	plan9.Fcall{Type: plan9.Rcreate, Tag: plan9.NOTAG, Qid: plan9Qids[0], Iounit: math.MaxUint32},

	plan9.Fcall{Type: plan9.Tread, Tag: 42, Fid: 1, Offset: 0, Count: 8192},
	plan9.Fcall{Type: plan9.Tread, Tag: plan9.NOTAG, Fid: math.MaxUint32, Offset: math.MaxUint64,
		Count: math.MaxUint32},
	plan9.Fcall{Type: plan9.Tread, Tag: 0, Fid: 0, Offset: 0, Count: 0},

	plan9.Fcall{Type: plan9.Rread, Tag: 42, Count: uint32(len(testData)), Data: testData},
	plan9.Fcall{Type: plan9.Rread, Tag: plan9.NOTAG, Count: uint32(len(testData)),
		Data: testData},
	plan9.Fcall{Type: plan9.Rread, Tag: plan9.NOTAG, Count: 0, Data: nil},

	plan9.Fcall{Type: plan9.Twrite, Tag: 42, Fid: 1, Offset: 0, Count: uint32(len(testData)),
		Data: testData},
	plan9.Fcall{Type: plan9.Twrite, Tag: plan9.NOTAG, Fid: math.MaxUint32, Offset: math.MaxUint64,
		Count: uint32(len(testData)), Data: testData},
	plan9.Fcall{Type: plan9.Rwrite, Tag: 42, Count: 0},
	plan9.Fcall{Type: plan9.Rwrite, Tag: 42, Count: 8192},

	plan9.Fcall{Type: plan9.Tclunk, Tag: 42, Fid: 42},
	plan9.Fcall{Type: plan9.Tclunk, Tag: plan9.NOTAG, Fid: math.MaxUint32},
	plan9.Fcall{Type: plan9.Tclunk, Tag: 0, Fid: 0},
	plan9.Fcall{Type: plan9.Rclunk, Tag: 42},
	plan9.Fcall{Type: plan9.Rclunk, Tag: plan9.NOTAG},
	plan9.Fcall{Type: plan9.Rclunk, Tag: 0},
	plan9.Fcall{Type: plan9.Tremove, Tag: 42, Fid: 42},
	plan9.Fcall{Type: plan9.Tremove, Tag: plan9.NOTAG, Fid: math.MaxUint32},
	plan9.Fcall{Type: plan9.Tremove, Tag: 0, Fid: 0},
	plan9.Fcall{Type: plan9.Rremove, Tag: 42},
	plan9.Fcall{Type: plan9.Rremove, Tag: plan9.NOTAG},
	plan9.Fcall{Type: plan9.Rremove, Tag: 0},

	/*
		plan9.Fcall{Type: plan9.Tstat, Tag: 42, Fid: 42},
		//plan9.Fcall{Type: plan9.Rstat, Tag: 42, Stat: testDirBytes},
		//plan9.Fcall{Type: plan9.Rstat, Tag: plan9.NOTAG, Stat: testMaxDirBytes},
		//plan9.Fcall{Type: plan9.Twstat, Tag: 42, Fid: 42, Stat: testDirBytes},
		//plan9.Fcall{Type: plan9.Twstat, Tag: plan9.NOTAG, Fid: plan9.NOFID, Stat: testMaxDirBytes},
		plan9.Fcall{Type: plan9.Rwstat, Tag: 42},

		plan9.Fcall{Type: plan9.Rerror, Ename: "error message"},
	*/
}

var (
	plan9Dirs = []plan9.Dir{
		{
			Type: 0x01, Dev: 42, Qid: plan9.Qid{Type: plan9.QTFILE, Vers: 42, Path: 42}, Mode: 42,
			Atime: uint32(time.Now().Unix()), Mtime: uint32(time.Now().Unix()), Length: 42,
			Name: "tmp", Uid: "glenda", Gid: "lab", Muid: "bootes",
		},
		{
			Type: math.MaxUint16, Dev: math.MaxUint16, Qid: plan9Qids[0],
			Mode: math.MaxUint32, Atime: math.MaxUint32, Mtime: math.MaxUint32,
			Length: math.MaxUint64, Name: maxNameStr, Uid: maxUnameStr, Gid: maxUnameStr,
			Muid: maxUnameStr,
		},
		{},
	}

	plan9Qids = []plan9.Qid{
		{Type: math.MaxUint8, Vers: math.MaxUint32, Path: math.MaxUint64},
		{Type: 42, Vers: 42, Path: 42},
		{},
	}
)

func cmpQid(t *testing.T, q1 plan9.Qid, q2 Qid) {
	t.Helper()

	if q1.Type != q2.Type {
		t.Fatalf("compare qid: expected type %d, have %d", q1.Type, q2.Type)
	}
	if q1.Vers != q2.Version {
		t.Fatalf("compare qid: expected version %d, have %d", q1.Vers, q2.Version)
	}
	if q1.Path != q2.Path {
		t.Fatalf("compare qid: expected path %d, have %d", q1.Path, q2.Path)
	}
}

func cmpDirStat(t *testing.T, dir plan9.Dir, s Stat) {
	t.Helper()

	if dir.Type != s.Type {
		t.Fatalf("compare stat: expected type %d, have %d", dir.Type, s.Type)
	}
	if dir.Dev != s.Dev {
		t.Fatalf("compare stat: expected dev %d, have %d", dir.Dev, s.Dev)
	}
	cmpQid(t, dir.Qid, s.Qid)
	if uint32(dir.Mode) != s.Mode {
		t.Fatalf("compare mode: expected mode %d, have %d", dir.Mode, s.Mode)
	}
	if dir.Atime != s.Atime {
		t.Fatalf("compare stat: expected atime %d, have %d", dir.Atime, s.Atime)
	}
	if dir.Mtime != s.Mtime {
		t.Fatalf("compare stat: expected dev %d, have %d", dir.Mtime, s.Mtime)
	}
	if dir.Length != s.Length {
		t.Fatalf("compare stat: expected dev %d, have %d", dir.Length, s.Length)
	}
	if dir.Name != s.Name {
		t.Fatalf("compare stat: expected name %q, have %q", dir.Name, s.Name)
	}
	if dir.Uid != s.UID {
		t.Fatalf("compare stat: expected uid %q, have %q", dir.Uid, s.UID)
	}
	if dir.Gid != s.GID {
		t.Fatalf("compare stat: expected gid %q, have %q", dir.Gid, s.GID)
	}
	if dir.Muid != s.MUID {
		t.Fatalf("compare stat: expected name %q, have %q", dir.Muid, s.MUID)
	}
}

func TestStatCompat(t *testing.T) {
	for i, dir := range plan9Dirs {
		data, err := dir.Bytes()
		if err != nil {
			t.Fatalf("compat dir #%d: marshal plan9 dir failed: %v", i, err)
		}

		s := Stat{}
		if err = s.UnmarshalBinary(data); err != nil {
			t.Fatalf("compat dir %d: unmarshal stat failed: %v", i, err)
		}

		cmpDirStat(t, dir, s)
	}
}

func TestRstatFcallCompat(t *testing.T) {
	fcall := plan9.Fcall{Type: plan9.Rstat, Tag: 42}
	buf := newBuffer(64)
	enc := newEncoder(&buf)
	var err error

	for i, dir := range plan9Dirs {
		fcall.Stat, err = dir.Bytes()
		if err != nil {
			t.Fatalf("compat Rstat fcall #%d: marshal plan9 dir failed: %v", i, err)
		}

		s := Stat{}
		if err = s.UnmarshalBinary(fcall.Stat); err != nil {
			t.Fatalf("compat Rstat fcall %d: unmarshal stat failed: %v", i, err)
		}

		data, err := fcall.Bytes()
		if err != nil {
			t.Fatalf("compat Rstat fcall %d: marshal fcall failed: %v", i, err)
		}
		if err = enc.Rstat(fcall.Tag, s); err != nil {
			t.Fatalf("compat Rstat fcall %d: encode failed: %v", i, err)
		}
		if !bytes.Equal(data, buf) {
			t.Fatalf("compat Rstat %d: messages differ\nexpected %q\nhave     %q", i, data, buf)
		}
		buf = buf[:0]
	}
}

func TestTwstatFcallCompat(t *testing.T) {
	fcall := plan9.Fcall{Type: plan9.Twstat, Tag: 42, Fid: 42}
	buf := newBuffer(64)
	enc := newEncoder(&buf)
	var err error

	for i, dir := range plan9Dirs {
		fcall.Stat, err = dir.Bytes()
		if err != nil {
			t.Fatalf("compat Twstat fcall #%d: marshal plan9 dir failed: %v", i, err)
		}

		s := Stat{}
		if err = s.UnmarshalBinary(fcall.Stat); err != nil {
			t.Fatalf("compat Twstat fcall %d: unmarshal stat failed: %v", i, err)
		}

		data, err := fcall.Bytes()
		if err != nil {
			t.Fatalf("compat Twstat fcall %d: marshal fcall failed: %v", i, err)
		}
		if err = enc.Twstat(fcall.Tag, fcall.Fid, s); err != nil {
			t.Fatalf("compat Twstat fcall %d: encode failed: %v", i, err)
		}
		if !bytes.Equal(data, buf) {
			t.Fatalf("compat Twstat %d: messages differ\nexpected %q\nhave     %q", i, data, buf)
		}
		buf = buf[:0]
	}
}

func plan9QidToQid(q plan9.Qid) Qid {
	return Qid{q.Type, q.Vers, q.Path}
}

func plan9StatToStat(data []byte) Stat {
	d, err := plan9.UnmarshalDir(data)
	if err != nil {
		panic(fmt.Sprintf("cannot marshal plan9 dir: %v", err))
	}
	return Stat{
		Type:   d.Type,
		Dev:    d.Dev,
		Qid:    plan9QidToQid(d.Qid),
		Mode:   uint32(d.Mode),
		Atime:  d.Atime,
		Mtime:  d.Mtime,
		Length: d.Length,
		Name:   d.Name,
		UID:    d.Uid,
		GID:    d.Gid,
		MUID:   d.Muid,
	}
}

func TestBasicEncoderCompatiblity(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)

	for i, f := range testFcalls {
		buf.Reset()

		data, err := f.Bytes()
		if err != nil {
			t.Fatalf("basic enc test[#%d]: cannot marshal 9fans fcall: %v", i, err)
		}

		switch f.Type {
		case plan9.Tversion:
			err = enc.Tversion(f.Tag, f.Msize, f.Version)
		case plan9.Rversion:
			err = enc.Rversion(f.Tag, f.Msize, f.Version)
		case plan9.Tauth:
			err = enc.Tauth(f.Tag, f.Afid, f.Uname, f.Aname)
		case plan9.Rauth:
			err = enc.Rauth(f.Tag, plan9QidToQid(f.Aqid))
		case plan9.Tattach:
			err = enc.Tattach(f.Tag, f.Fid, f.Afid, f.Uname, f.Aname)
		case plan9.Rattach:
			err = enc.Rattach(f.Tag, plan9QidToQid(f.Qid))
		case plan9.Tflush:
			err = enc.Tflush(f.Tag, f.Oldtag)
		case plan9.Rflush:
			err = enc.Rflush(f.Tag)
		case plan9.Twalk:
			err = enc.Twalk(f.Tag, f.Fid, f.Newfid, f.Wname...)

		case plan9.Rwalk:
			qids := make([]Qid, 0, len(f.Wqid))
			for _, q := range f.Wqid {
				qids = append(qids, plan9QidToQid(q))
			}
			err = enc.Rwalk(f.Tag, qids...)

		case plan9.Topen:
			enc.Topen(f.Tag, f.Fid, f.Mode)
		case plan9.Ropen:
			enc.Ropen(f.Tag, plan9QidToQid(f.Qid), f.Iounit)
		case plan9.Tcreate:
			enc.Tcreate(f.Tag, f.Fid, f.Name, uint32(f.Perm), f.Mode)
		case plan9.Rcreate:
			enc.Rcreate(f.Tag, plan9QidToQid(f.Qid), f.Iounit)
		case plan9.Tread:
			enc.Tread(f.Tag, f.Fid, f.Offset, f.Count)
		case plan9.Rread:
			enc.Rread(f.Tag, f.Data)
		case plan9.Twrite:
			enc.Twrite(f.Tag, f.Fid, f.Offset, f.Data)
		case plan9.Rwrite:
			enc.Rwrite(f.Tag, f.Count)
		case plan9.Tclunk:
			enc.Tclunk(f.Tag, f.Fid)
		case plan9.Rclunk:
			enc.Rclunk(f.Tag)
		case plan9.Tremove:
			enc.Tremove(f.Tag, f.Fid)
		case plan9.Rremove:
			enc.Rremove(f.Tag)
			/*
				case plan9.Tstat:
					enc.Tstat(f.Tag, f.Fid)
				case plan9.Rstat:
					enc.Rstat(f.Tag, plan9StatToStat(f.Stat))
				case plan9.Twstat:
					enc.Twstat(f.Tag, f.Fid, plan9StatToStat(f.Stat))
				case plan9.Rwstat:
					enc.Rwstat(f.Tag)
				case plan9.Rerror:
					enc.Rerrorf(f.Tag, f.Ename)
			*/
		}
		if err != nil {
			t.Fatalf("basic enc test[#%d]: error encoding: %v", i, err)
		}

		if !bytes.Equal(data, buf.Bytes()) {
			t.Fatalf("basic enc test[#%d]: encoding differ\nwant: %q\nhave: %q",
				i, data, buf.Bytes())
		}
	}
}
