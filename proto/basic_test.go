package proto

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	"9fans.net/go/plan9"
)

var (
	testMaxQid = plan9.Qid{Type: math.MaxUint8, Vers: math.MaxUint32, Path: math.MaxUint64}
	testQid    = plan9.Qid{Type: 2, Vers: 1, Path: 42}

	testMaxDir = plan9.Dir{
		Type:   math.MaxUint16,
		Dev:    math.MaxUint16,
		Qid:    testMaxQid,
		Mode:   math.MaxUint32,
		Atime:  math.MaxUint32,
		Mtime:  math.MaxUint32,
		Length: math.MaxUint64,
		Name:   "tmp",
		Uid:    "glenda",
		Gid:    "glenda",
		Muid:   "glenda",
	}
	testDir = plan9.Dir{
		Type:   42,
		Dev:    42,
		Qid:    plan9.Qid{42, 42, 42},
		Mode:   42,
		Atime:  42,
		Mtime:  42,
		Length: 42,
		Name:   "tmp",
		Uid:    "glenda",
		Gid:    "glenda",
		Muid:   "glenda",
	}
	testMaxDirBytes, _ = testMaxDir.Bytes()
	testDirBytes, _    = testDir.Bytes()

	testData = make([]byte, 8192)
)

var testFcalls = []plan9.Fcall{
	plan9.Fcall{Type: plan9.Tversion, Tag: NOTAG, Msize: 8192, Version: "9P2000"},
	plan9.Fcall{Type: plan9.Rversion, Tag: NOTAG, Msize: 8192, Version: "9P2000"},

	plan9.Fcall{Type: plan9.Tauth, Tag: NOTAG, Afid: NOFID, Uname: "bootes", Aname: "tmp"},
	plan9.Fcall{Type: plan9.Rauth, Tag: NOTAG, Aqid: testMaxQid},
	plan9.Fcall{Type: plan9.Tattach, Tag: NOTAG, Fid: 1, Afid: NOFID, Uname: "bootes",
		Aname: "tmp"},
	plan9.Fcall{Type: plan9.Rattach, Tag: NOTAG, Qid: testMaxQid},

	plan9.Fcall{Type: plan9.Tflush, Tag: 42, Oldtag: 41},
	plan9.Fcall{Type: plan9.Tflush, Tag: math.MaxUint16, Oldtag: math.MaxUint16},
	plan9.Fcall{Type: plan9.Rflush, Tag: 42},
	plan9.Fcall{Type: plan9.Rflush, Tag: math.MaxUint16},

	plan9.Fcall{Type: plan9.Twalk, Tag: 42, Fid: 1, Newfid: 2,
		Wname: []string{"usr", "glenda", "lib"}},
	plan9.Fcall{Type: plan9.Twalk, Tag: 2, Fid: math.MaxUint32 - 1, Newfid: math.MaxUint32,
		Wname: []string{"usr", "glenda", "lib"}},
	plan9.Fcall{Type: plan9.Rwalk, Tag: 42,
		Wqid: []plan9.Qid{testQid, testQid, testMaxQid}},

	plan9.Fcall{Type: plan9.Topen, Tag: 42, Fid: 1, Mode: plan9.OREAD},
	plan9.Fcall{Type: plan9.Topen, Tag: 42, Fid: 1, Mode: plan9.OWRITE},
	plan9.Fcall{Type: plan9.Topen, Tag: 42, Fid: 1, Mode: plan9.ORDWR},
	plan9.Fcall{Type: plan9.Ropen, Tag: 42, Qid: testQid, Iounit: 8192},
	plan9.Fcall{Type: plan9.Ropen, Tag: 42, Qid: testMaxQid, Iounit: 8192},
	plan9.Fcall{Type: plan9.Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0644,
		Mode: plan9.OREAD},
	plan9.Fcall{Type: plan9.Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0660,
		Mode: plan9.OWRITE},
	plan9.Fcall{Type: plan9.Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0600,
		Mode: plan9.ORDWR},
	plan9.Fcall{Type: plan9.Rcreate, Tag: 42, Qid: testQid, Iounit: 8192},
	plan9.Fcall{Type: plan9.Rcreate, Tag: 42, Qid: testMaxQid, Iounit: 8192},

	plan9.Fcall{Type: plan9.Tread, Tag: 42, Fid: 42, Offset: 0, Count: 8192},
	plan9.Fcall{Type: plan9.Tread, Tag: 42, Fid: 42, Offset: 8192, Count: 8192},
	plan9.Fcall{Type: plan9.Tread, Tag: 42, Fid: 42, Offset: 0, Count: 0},
	plan9.Fcall{Type: plan9.Rread, Tag: 42, Count: uint32(len(testData)), Data: testData},

	plan9.Fcall{Type: plan9.Twrite, Tag: 42, Fid: 42, Offset: 0,
		Count: uint32(len(testData)), Data: testData},
	plan9.Fcall{Type: plan9.Twrite, Tag: 42, Fid: 42, Offset: 8192,
		Count: uint32(len(testData)), Data: testData},
	plan9.Fcall{Type: plan9.Twrite, Tag: 42, Fid: 42, Offset: 0,
		Count: 0, Data: []byte{}},
	plan9.Fcall{Type: plan9.Rwrite, Tag: 42, Count: 0},
	plan9.Fcall{Type: plan9.Rwrite, Tag: 42, Count: 8192},

	plan9.Fcall{Type: plan9.Tclunk, Tag: 42, Fid: 42},
	plan9.Fcall{Type: plan9.Rclunk, Tag: 42},
	plan9.Fcall{Type: plan9.Tremove, Tag: 42, Fid: 42},
	plan9.Fcall{Type: plan9.Rremove, Tag: 42},

	plan9.Fcall{Type: plan9.Tstat, Tag: 42, Fid: 42},
	plan9.Fcall{Type: plan9.Rstat, Tag: 42, Stat: testDirBytes},
	plan9.Fcall{Type: plan9.Rstat, Tag: NOTAG, Stat: testMaxDirBytes},
	plan9.Fcall{Type: plan9.Twstat, Tag: 42, Fid: 42, Stat: testDirBytes},
	plan9.Fcall{Type: plan9.Twstat, Tag: NOTAG, Fid: NOFID, Stat: testMaxDirBytes},
	plan9.Fcall{Type: plan9.Rwstat, Tag: 42},

	plan9.Fcall{Type: plan9.Rerror, Ename: "error message"},
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
		Uid:    d.Uid,
		Gid:    d.Gid,
		Muid:   d.Muid,
	}
}

func TestStatCompatiblity(t *testing.T) {
	for i, d := range []plan9.Dir{
		{
			Type:   42,
			Dev:    42,
			Qid:    plan9.Qid{42, 42, 42},
			Mode:   42,
			Atime:  42,
			Mtime:  42,
			Length: 42,
			Name:   "tmp",
			Uid:    "glenda",
			Gid:    "glenda",
			Muid:   "glenda",
		},
		{
			Type:   math.MaxUint16,
			Dev:    math.MaxUint16,
			Qid:    testMaxQid,
			Mode:   math.MaxUint32,
			Atime:  math.MaxUint32,
			Mtime:  math.MaxUint32,
			Length: math.MaxUint64,
			Name:   "tmp",
			Uid:    "glenda",
			Gid:    "glenda",
			Muid:   "glenda",
		},
		{},
	} {
		data1, err := d.Bytes()
		if err != nil {
			t.Fatalf("binary#%d: cannot marshal dir: %v", i, err)
		}

		stat := Stat{
			Type:   d.Type,
			Dev:    d.Dev,
			Qid:    plan9QidToQid(d.Qid),
			Mode:   uint32(d.Mode),
			Atime:  d.Atime,
			Mtime:  d.Mtime,
			Length: d.Length,
			Name:   d.Name,
			Uid:    d.Uid,
			Gid:    d.Gid,
			Muid:   d.Muid,
		}
		data2, err := stat.MarshalBinary()
		if err != nil {
			t.Fatalf("binary#%d: cannot marshal stat: %v", i, err)
		}

		if !bytes.Equal(data1, data2) {
			t.Fatalf("binary#%d: marshal result differ\n%q\n%q", i, data1, data2)
		}
	}
}

func TestEncoderBasicCompatiblity(t *testing.T) {
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
			enc.Tversion(f.Tag, f.Msize, f.Version)
		case plan9.Rversion:
			enc.Rversion(f.Tag, f.Msize, f.Version)
		case plan9.Tauth:
			enc.Tauth(f.Tag, f.Afid, f.Uname, f.Aname)
		case plan9.Rauth:
			enc.Rauth(f.Tag, plan9QidToQid(f.Aqid))
		case plan9.Tattach:
			enc.Tattach(f.Tag, f.Fid, f.Afid, f.Uname, f.Aname)
		case plan9.Rattach:
			enc.Rattach(f.Tag, plan9QidToQid(f.Qid))
		case plan9.Tflush:
			enc.Tflush(f.Tag, f.Oldtag)
		case plan9.Rflush:
			enc.Rflush(f.Tag)
		case plan9.Twalk:
			enc.Twalk(f.Tag, f.Fid, f.Newfid, f.Wname...)
		case plan9.Rwalk:
			qids := make([]Qid, 0, len(f.Wqid))
			for _, q := range f.Wqid {
				qids = append(qids, plan9QidToQid(q))
			}
			enc.Rwalk(f.Tag, qids...)
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
		}
		if err = enc.Flush(); err != nil {
			t.Fatalf("basic enc test[#%d]: error flushing encoder: %v", i, err)
		}

		if !bytes.Equal(data, buf.Bytes()) {
			t.Fatalf("basic enc test[#%d]: encoding differ\nwant: %q\nhave: %q",
				i, data, buf.Bytes())
		}
	}
}
