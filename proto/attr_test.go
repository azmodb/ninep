package proto

import (
	"bytes"
	"testing"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

func attrBytesEqual(t *testing.T, st *unix.Stat_t, rx *Rgetattr) {
	t.Helper()
	b1, b2 := &binary.Buffer{}, &binary.Buffer{}

	if err := EncodeStat(b1, 0, st); err != nil {
		t.Fatalf("stat: unexpected marshal error: %v", err)
	}

	rx.Encode(b2)
	if err := b2.Err(); err != nil {
		t.Fatalf("stat: unexpected marshal error: %v", err)
	}

	if !bytes.Equal(b1.Bytes(), b2.Bytes()) {
		t.Fatalf("stat: buffers differ\nwant %q\ngot  %q",
			b2.Bytes(), b1.Bytes())
	}
}

func TestEmptyStatEncoding(t *testing.T) {
	st := &unix.Stat_t{}
	rx := &Rgetattr{}

	attrBytesEqual(t, st, rx)
}

func TestStatToRgetattr(t *testing.T) {
	name, st := "attr_test.go", &unix.Stat_t{}
	if err := unix.Stat(name, st); err != nil {
		t.Fatalf("stat: %v", err)
	}

	rx := StatToRgetattr(st)

	attrBytesEqual(t, st, rx)
}
