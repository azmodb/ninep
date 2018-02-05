package proto

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"time"
)

var (
	testStats = []Stat{
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
	}
	testQids = []Qid{
		{
			Type:    math.MaxUint8,
			Version: math.MaxUint32,
			Path:    math.MaxUint64,
		},
		{
			Type:    0x01,
			Version: 42,
			Path:    42,
		},
		{},
	}
)

func TestQidMarshalUnmarshalBinary(t *testing.T) {
	for i, in := range testQids {
		data, err := in.MarshalBinary()
		if err != nil {
			t.Fatalf("binary#%d: cannot marshal qid: %v", i, err)
		}

		out := Qid{}
		if err = out.UnmarshalBinary(data); err != nil {
			t.Fatalf("binary#%d: cannot unmarshal qid: %v", i, err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("binary#%d: in/out differ:\n%s\n%s", i, in, out)
		}
	}
}

func TestStatMarshalUnmarshalBinary(t *testing.T) {
	for i, in := range testStats {
		data, err := in.MarshalBinary()
		if err != nil {
			t.Fatalf("binary#%d: cannot marshal stat: %v", i, err)
		}

		out := Stat{}
		if err = out.UnmarshalBinary(data); err != nil {
			t.Fatalf("binary#%d: cannot unmarshal stat: %v", i, err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Fatalf("binary#%d: in/out differ:\n%s\n%s", i, in, out)
		}
	}
}

func TestStatMarshalEncoderCompatiblity(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := newEncoder(buf)

	for i, stat := range testStats {
		buf.Reset()

		strLen := len(stat.Name) + len(stat.Uid) + len(stat.Gid) + len(stat.Muid)
		statLen := uint16(fixedStatLen + strLen)
		enc.pstat(stat, statLen)
		if err := enc.Flush(); err != nil {
			t.Fatalf("compat#%d: cannot encode stat structure: %v", i, err)
		}

		data, err := stat.MarshalBinary()
		if err != nil {
			t.Fatalf("compat#%d: cannot marshal stat structure: %v", i, err)
		}

		if !bytes.Equal(buf.Bytes(), data) {
			t.Fatalf("compare#%d: encode and marshal result differ\n%q\n%q", i, buf.Bytes(), data)
		}
	}
}
