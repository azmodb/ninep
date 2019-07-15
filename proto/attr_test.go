package proto

import (
	"bytes"
	"reflect"
	"runtime"
	"testing"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

func statfsBytesEqual(t *testing.T, st *unix.Statfs_t, rx *Rstatfs) {
	t.Helper()
	b1, b2 := &binary.Buffer{}, &binary.Buffer{}

	if err := EncodeRstatfs(b1, st); err != nil {
		t.Fatalf("statfs: unexpected marshal error: %v", err)
	}

	rx.Encode(b2)
	if err := b2.Err(); err != nil {
		t.Fatalf("statfs: unexpected marshal error: %v", err)
	}

	if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
		t.Fatalf("statfs: buffers differ\nwant %q\ngot  %q", b2.Bytes(), b1.Bytes())
	}
}

func statBytesEqual(t *testing.T, st *unix.Stat_t, rx *Rgetattr) {
	t.Helper()
	b1, b2 := &binary.Buffer{}, &binary.Buffer{}

	if err := EncodeRgetattr(b1, 0, st); err != nil {
		t.Fatalf("stat: unexpected marshal error: %v", err)
	}

	rx.Encode(b2)
	if err := b2.Err(); err != nil {
		t.Fatalf("stat: unexpected marshal error: %v", err)
	}

	if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
		t.Fatalf("stat: buffers differ\nwant %q\ngot  %q", b2.Bytes(), b1.Bytes())
	}

	attr := &Rgetattr{Stat_t: &unix.Stat_t{}}
	attr.Decode(b1)
	if err := b1.Err(); err != nil {
		t.Fatalf("stat: unexpected unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(rx, attr) {
		t.Fatalf("stat: attr differ\nwant %q\ngot  %q", rx, attr)
	}
}

func TestEmptyStatEncoding(t *testing.T) {
	st := &unix.Stat_t{}
	rx := &Rgetattr{Stat_t: &unix.Stat_t{}}

	statBytesEqual(t, st, rx)
}

func TestStatEncoding(t *testing.T) {
	s1, s2 := &unix.Stat_t{}, &unix.Stat_t{}
	unix.Stat("attr_test.go", s1)
	var valid uint64
	b := &binary.Buffer{}

	if err := EncodeRgetattr(b, GetAttrBasic, s1); err != nil {
		t.Fatalf("encode: unexpected rgetattr encode error: %v", err)
	}
	if err := DecodeRgetattr(b, &valid, s2); err != nil {
		t.Fatalf("decode: unexpected rgetattr encode error: %v", err)
	}

	//if !reflect.DeepEqual(s1, s2) {
	//	t.Fatalf("stat: attr differ\nwant %+v\ngot  %+v", s1, s2)
	//}

	if b.Len() != 0 {
		t.Fatalf("codec: expected empty buffer, have %d", b.Len())
	}
}

func TestEmptyStatfsEncoding(t *testing.T) {
	st := &unix.Statfs_t{}
	rx := &Rstatfs{}
	if runtime.GOOS == "darwin" {
		rx.NameLength = 255
	}

	statfsBytesEqual(t, st, rx)
}
