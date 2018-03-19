package ninep

import (
	"math"
	"reflect"
	"testing"
	"time"
)

var testQids = []Qid{
	Qid{Type: QTDIR, Version: math.MaxUint32, Path: math.MaxUint64},
	Qid{Type: QTFILE, Version: 42, Path: 42},
	Qid{},
}

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
	for _, in := range []Stat{
		{
			Type:   math.MaxUint16,
			Dev:    math.MaxUint16,
			Qid:    testQids[0],
			Mode:   math.MaxUint32,
			Atime:  math.MaxUint32,
			Mtime:  math.MaxUint32,
			Length: math.MaxUint64,
			Name:   "/tmp",
			Uid:    "glenda",
			Gid:    "lab",
			Muid:   "bootes",
		},
		{
			Type:   0x01,
			Dev:    42,
			Qid:    testQids[1],
			Mode:   42,
			Atime:  uint32(time.Now().Unix()),
			Mtime:  uint32(time.Now().Unix()),
			Length: math.MaxUint32,
			Name:   "/tmp",
			Uid:    "glenda",
			Gid:    "lab",
			Muid:   "bootes",
		},
		{},
	} {
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
