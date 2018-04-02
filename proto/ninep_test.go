package proto

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

var (
	maxQid   = Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64}
	emptyQid = Qid{}

	emptyStat = Stat{}

	maxVersionStr string
	maxUnameStr   string
	maxNameStr    string
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	maxVersionStr = randString(maxVersionLen)
	maxUnameStr = randString(maxUnameLen)
	maxNameStr = randString(maxNameLen)
}

func TestQidEncoding(t *testing.T) {
	for _, in := range []Qid{
		Qid{Type: QTFILE, Version: 42, Path: 42},
		maxQid,
		emptyQid,
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

func TestStatEncoding(t *testing.T) {
	for _, in := range []Stat{
		Stat{
			Type: 0x01, Dev: 42, Qid: Qid{Type: QTFILE, Version: 42, Path: 42},
			Mode: 42, Atime: uint32(time.Now().Unix()),
			Mtime: uint32(time.Now().Unix()), Length: 42,
			Name: "tmp", UID: "glenda", GID: "lab", MUID: "bootes",
		},
		Stat{
			Type: math.MaxUint16, Dev: math.MaxUint16, Qid: maxQid,
			Mode: math.MaxUint32, Atime: math.MaxUint32,
			Mtime: math.MaxUint32, Length: math.MaxUint64,
			Name: maxNameStr, UID: maxUnameStr, GID: maxUnameStr,
			MUID: maxUnameStr,
		},
		emptyStat,
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

func TestStatLimit(t *testing.T) {
	s := Stat{}
	if s.len() != fixedStatLen {
		t.Fatalf("stat: expected empty len %d, have %d", fixedStatLen, s.len())
	}
}
