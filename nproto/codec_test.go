package ninep

import (
	"bytes"
	"crypto/rand"
	"math"
	"reflect"
	"testing"
	"time"
)

var (
	testStats = []Stat{
		Stat{
			Type: math.MaxUint16, Dev: math.MaxUint16, Qid: testQids[0], Mode: math.MaxUint32,
			Atime: math.MaxUint32, Mtime: math.MaxUint32, Length: math.MaxUint64,
			Name: "/tmp", UID: "glenda", GID: "lab", MUID: "bootes",
		},
		Stat{
			Type: 0x01, Dev: 42, Qid: testQids[1], Mode: 42,
			Atime: uint32(time.Now().Unix()), Mtime: uint32(time.Now().Unix()), Length: 42,
			Name: "/tmp", UID: "glenda", GID: "lab", MUID: "bootes",
		},
		Stat{},
	}

	testQids = []Qid{
		Qid{Type: QTDIR, Version: math.MaxUint32, Path: math.MaxUint64},
		Qid{Type: QTFILE, Version: 42, Path: 42},
		Qid{},
	}
)

func TestQidEncoding(t *testing.T) {
	for _, in := range testQids {
		data, err := in.MarshalBinary()
		if err != nil {
			t.Fatalf("qid: marshal failed: %v", err)
		}

		out := Qid{}
		err = out.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("qid: unmarshal failed: %v", err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("qid: expected %+v, got %+v", in, out)
		}
	}
}

func TestStatEncoding(t *testing.T) {
	for _, in := range testStats {
		data, err := in.MarshalBinary()
		if err != nil {
			t.Fatalf("stat: marshal failed: %v", err)
		}

		out := Stat{}
		err = out.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("stat: unmarshal failed: %v", err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("stat: expected %+v, got %+v", in, out)
		}
	}
}

var testData = make([]byte, 8192)

func init() {
	if _, err := rand.Read(testData); err != nil {
		panic("out of entropy")
	}
}

func TestMessageEncoding(t *testing.T) {
	for _, in := range []Message{
		Message{Type: Tversion, Msize: 8192, Version: "9P2000"},
		Message{Type: Rversion, Msize: 8192, Version: "9P2000"},

		Message{Type: Tauth, Tag: NOTAG, Afid: NOFID, Uname: "bootes", Aname: "tmp"},
		Message{Type: Rauth, Tag: NOTAG, Qid: testQids[0]},

		Message{Type: Tattach, Tag: NOTAG, Fid: 1, Afid: NOFID, Uname: "bootes",
			Aname: "tmp"},
		Message{Type: Rattach, Tag: NOTAG, Qid: testQids[0]},

		Message{Type: Tflush, Tag: 42, Oldtag: 41},
		Message{Type: Tflush, Tag: math.MaxUint16, Oldtag: math.MaxUint16},
		Message{Type: Rflush, Tag: 42},
		Message{Type: Rflush, Tag: math.MaxUint16},

		Message{Type: Twalk, Tag: 42, Fid: 1, Newfid: 2,
			Wname: []string{"usr", "glenda", "lib"}},
		Message{Type: Twalk, Tag: 2, Fid: math.MaxUint32 - 1, Newfid: math.MaxUint32,
			Wname: []string{"usr", "glenda", "lib"}},
		Message{Type: Rwalk, Tag: 42, Wqid: testQids},

		Message{Type: Topen, Tag: 42, Fid: 1, Mode: OREAD},
		Message{Type: Topen, Tag: 42, Fid: 1, Mode: OWRITE},
		Message{Type: Topen, Tag: 42, Fid: 1, Mode: ORDWR},
		Message{Type: Ropen, Tag: 42, Qid: testQids[0], Iounit: 8192},
		Message{Type: Ropen, Tag: 42, Qid: testQids[2], Iounit: 8192},

		Message{Type: Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0644,
			Mode: OREAD},
		Message{Type: Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0660,
			Mode: OWRITE},
		Message{Type: Tcreate, Tag: 42, Fid: 1, Name: "file", Perm: 0600,
			Mode: ORDWR},
		Message{Type: Rcreate, Tag: 42, Qid: testQids[0], Iounit: 8192},
		Message{Type: Rcreate, Tag: 42, Qid: testQids[2], Iounit: 8192},

		Message{Type: Tread, Tag: 42, Fid: 42, Offset: 0, Count: 8192},
		Message{Type: Tread, Tag: 42, Fid: 42, Offset: 8192, Count: 8192},
		Message{Type: Tread, Tag: 42, Fid: 42, Offset: 0, Count: 0},

		Message{Type: Rread, Tag: 42, Count: uint32(len(testData)), Data: testData},

		Message{Type: Twrite, Tag: 42, Fid: 42, Offset: 0,
			Count: uint32(len(testData)), Data: testData},
		Message{Type: Twrite, Tag: 42, Fid: 42, Offset: 8192,
			Count: uint32(len(testData)), Data: testData},
		Message{Type: Twrite, Tag: 42, Fid: 42, Offset: 0,
			Count: 0, Data: []byte{}},

		Message{Type: Rwrite, Tag: 42, Count: 0},
		Message{Type: Rwrite, Tag: 42, Count: 8192},

		Message{Type: Tclunk, Tag: 42, Fid: 42},
		Message{Type: Rclunk, Tag: 42},
		Message{Type: Tremove, Tag: 42, Fid: 42},
		Message{Type: Rremove, Tag: 42},

		Message{Type: Tstat, Tag: 42, Fid: 42},
		Message{Type: Rstat, Tag: 42, Stat: testStats[0]},
		Message{Type: Rstat, Tag: 42, Stat: testStats[1]},
		Message{Type: Rstat, Tag: 42, Stat: testStats[2]},

		Message{Type: Twstat, Tag: 42, Fid: 42, Stat: testStats[0]},
		Message{Type: Twstat, Tag: 42, Fid: 42, Stat: testStats[1]},
		Message{Type: Twstat, Tag: 42, Fid: 42, Stat: testStats[2]},

		Message{Type: Rflush, Tag: 42},
		Message{Type: Rclunk, Tag: 42},
		Message{Type: Rremove, Tag: 42},
		Message{Type: Rwstat, Tag: 42},
	} {
		data, err := in.MarshalBinary()
		if err != nil {
			t.Fatalf("message: marshal failed: %v", err)
		}

		out := Message{}
		err = out.UnmarshalBinary(data)
		if err != nil {
			t.Fatalf("message: unmarshal failed: %v", err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("message: expected %+v, got %+v", in, out)
		}

		out = Message{}
		dec := NewDecoder(bytes.NewReader(data))
		if err = dec.Decode(&out); err != nil {
			t.Fatalf("message: decode failed: %v", err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("message: expected %+v, got %+v", in, out)
		}
	}
}
