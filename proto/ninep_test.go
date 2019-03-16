package proto

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestStatEncoding(t *testing.T) {
	for _, in := range []Stat{
		Stat{
			Type: 0x01, Dev: 42, Qid: Qid{Type: 0x01, Version: 42, Path: 42},
			Mode: 42, Atime: uint32(time.Now().Unix()),
			Mtime: uint32(time.Now().Unix()), Length: 42,
			Name: "tmp", Uid: "glenda", Gid: "lab", Muid: "bootes",
		},
		Stat{
			Type: math.MaxUint16, Dev: math.MaxUint16,
			Qid:  plan9QidToQid(maxQid),
			Mode: math.MaxUint32, Atime: math.MaxUint32,
			Mtime: math.MaxUint32, Length: math.MaxUint64,
			Name: maxDirName, Uid: maxUserName,
			Gid: maxUserName, Muid: maxUserName,
		},
		Stat{},
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

func TestQidEncoding(t *testing.T) {
	for _, in := range []Qid{
		Qid{Type: 0x01, Version: 42, Path: 42},
		plan9QidToQid(maxQid),
		Qid{},
	} {
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
