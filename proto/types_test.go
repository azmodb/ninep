package proto

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

var customTestPackets = []packet{
	{&Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64}, &Qid{}},
	{&Qid{}, &Qid{}},

	{&Dirent{
		Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64},
		math.MaxUint64,
		math.MaxUint8,
		string16.String(),
	}, &Dirent{}},
	{&Dirent{}, &Dirent{}},

	{&Rgetattr{
		Valid:       math.MaxUint64,
		Qid:         Qid{Type: math.MaxUint8, Version: math.MaxUint32, Path: math.MaxUint64},
		Mode:        math.MaxUint32,
		Uid:         math.MaxUint32,
		Gid:         math.MaxUint32,
		Nlink:       math.MaxUint64,
		Rdev:        math.MaxUint64,
		Size:        math.MaxUint64,
		BlockSize:   math.MaxUint64,
		Blocks:      math.MaxUint64,
		Atime:       unix.NsecToTimespec(math.MaxInt64),
		Mtime:       unix.NsecToTimespec(math.MaxInt64),
		Ctime:       unix.NsecToTimespec(math.MaxInt64),
		Btime:       unix.NsecToTimespec(math.MaxInt64),
		Gen:         math.MaxUint64,
		DataVersion: math.MaxUint64,
	}, &Rgetattr{}},
	{&Rgetattr{}, &Rgetattr{}},

	{&Twalk{}, &Twalk{}},
	{&Twalk{Fid: math.MaxUint32, NewFid: math.MaxUint32, Names: testNames}, &Twalk{}},

	{testRwalk, &Rwalk{}},
	{&Rwalk{}, &Rwalk{}},
}

func init() {
	testPackets = append(customTestPackets, generatedLegacyPackets...)
	testPackets = append(testPackets, generatedLinuxPackets...)

	string16 = strings.Builder{}
	for i := 0; i < math.MaxUint16; i++ {
		if i < math.MaxUint8 {
			string8.WriteByte('a' + byte(i%26))
		}
		string16.WriteByte('a' + byte(i%26))
	}
	bytes16 = []byte(string16.String())

	testNames = make([]string, MaxNames)
	testQids = make([]Qid, MaxNames)
	for i := 0; i < MaxNames; i++ {
		testNames[i] = string8.String()
		testQids[i] = testQid
	}
	*testRwalk = Rwalk(testQids)
}

type packet struct {
	in  Message
	out Message
}

var (
	string16 strings.Builder
	string8  strings.Builder
	bytes16  []byte

	testQid     = Qid{math.MaxUint8, math.MaxUint32, math.MaxUint64}
	testQids    []Qid
	testRwalk   = &Rwalk{}
	testNames   []string
	testPackets []packet
)

func TestCodec(t *testing.T) {
	buf := &binary.Buffer{}
	for n, test := range testPackets {
		test.in.Encode(buf)
		if err := buf.Err(); err != nil {
			t.Errorf("encode(%d): unexpected error: %v", n, err)
			continue
		}

		test.out.Decode(buf)
		if err := buf.Err(); err != nil {
			t.Errorf("decode(%d): unexpected error: %v", n, err)
			continue
		}

		if !reflect.DeepEqual(test.in, test.out) {
			t.Errorf("codec(%d): test message differ\n%v\n%v", n,
				test.in, test.out)
		}

		if test.in.Len() != test.out.Len() {
			t.Errorf("codec(%d): expected message size %d, got %d",
				n, test.in.Len(), test.out.Len())
		}
		test.out.Reset()
	}
}

func TestRgetattrLen(t *testing.T) {
	buf := &binary.Buffer{}
	m := &Rgetattr{}

	m.Encode(buf)
	if err := buf.Err(); err != nil {
		t.Fatalf("rgetattr: unexpected encode error: %v", err)
	}
	if buf.Len() != m.Len() {
		t.Fatalf("rgetattr: expected length %d, got %d", buf.Len(), m.Len())
	}
}

func TestQidCodec(t *testing.T) {
	buf := make([]byte, 0, 13)
	for n, test := range customTestPackets[:2] {
		in, out := test.in.(*Qid), test.out.(*Qid)
		buf = buf[:0]

		buf, m, err := in.Marshal(buf)
		if err != nil {
			t.Fatalf("qid(%d): marshal error: %v", n, err)
		}
		if m != 13 {
			t.Fatalf("qid(%d): expected marshal length 13, got %d", n, m)
		}

		if _, err = out.Unmarshal(buf); err != nil {
			t.Fatalf("qid(%d): unmarshal error: %v", n, err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Errorf("qid(%d): qid differ\n%v\n%v", n, in, out)
		}

	}
}

func TestDirentCodec(t *testing.T) {
	buf := make([]byte, 0, 24)
	for n, test := range customTestPackets[2:4] {
		in, out := test.in.(*Dirent), test.out.(*Dirent)
		buf = buf[:0]

		buf, m, err := in.Marshal(buf)
		if err != nil {
			t.Fatalf("dirent(%d): marshal error: %v", n, err)
		}
		if m != 24 {
			t.Fatalf("dirent(%d): expected marshal length 24, got %d", n, m)
		}

		if _, err = out.Unmarshal(buf); err != nil {
			t.Fatalf("dirent(%d): unmarshal error: %v", n, err)
		}

		if !reflect.DeepEqual(in, out) {
			t.Errorf("dirent(%d): dirent differ\n%v\n%v", n, in, out)
		}

	}
}
