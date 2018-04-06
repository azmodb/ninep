package proto

import (
	"bytes"
	"math"
	"testing"
	"time"

	"9fans.net/go/plan9"
)

var (
	plan9Dirs = []plan9.Dir{
		{
			Type: 0x01, Dev: 42, Qid: plan9.Qid{Type: QTFILE, Vers: 42, Path: 42}, Mode: 42,
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
			t.Fatal("compat dir %d: unmarshal stat failed: %v", i, err)
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
			t.Fatal("compat Rstat fcall %d: unmarshal stat failed: %v", i, err)
		}

		data, err := fcall.Bytes()
		if err != nil {
			t.Fatal("compat Rstat fcall %d: marshal fcall failed: %v", i, err)
		}
		if err = enc.Rstat(fcall.Tag, s); err != nil {
			t.Fatal("compat Rstat fcall %d: encode failed: %v", i, err)
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
			t.Fatal("compat Twstat fcall %d: unmarshal stat failed: %v", i, err)
		}

		data, err := fcall.Bytes()
		if err != nil {
			t.Fatal("compat Twstat fcall %d: marshal fcall failed: %v", i, err)
		}
		if err = enc.Twstat(fcall.Tag, fcall.Fid, s); err != nil {
			t.Fatal("compat Twstat fcall %d: encode failed: %v", i, err)
		}
		if !bytes.Equal(data, buf) {
			t.Fatalf("compat Twstat %d: messages differ\nexpected %q\nhave     %q", i, data, buf)
		}
		buf = buf[:0]
	}
}
