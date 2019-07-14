package proto

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/azmodb/ninep/binary"
)

var customTestPackets = []packet{
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

func TestQidCodec(t *testing.T) {
	buf := make([]byte, 0, 13)
	for n, test := range []struct {
		in  *Qid
		out *Qid
	}{
		{&Qid{math.MaxUint8, math.MaxUint32, math.MaxUint64}, &Qid{}},
		{&Qid{}, &Qid{}},
	} {
		in, out := test.in, test.out
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
