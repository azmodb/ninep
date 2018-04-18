package proto

import (
	"bytes"
	"math"
	"testing"
)

func testVersion(t *testing.T, isTversion bool) {
	t.Helper()

	type versionInterface interface {
		Tag() uint16
		Msize() uint32
		Version() string
	}

	tests := []struct {
		tag     uint16
		msize   uint32
		version string
	}{
		{math.MaxUint8, math.MaxUint32, maxVersionStr},
		{},
		{42, 8192, "9P2000"},
	}

	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)
	dec := NewDecoder(buf)
	var err error
	for _, tx := range tests {
		if isTversion {
			err = enc.Tversion(tx.tag, tx.msize, tx.version)
		} else {
			err = enc.Rversion(tx.tag, tx.msize, tx.version)
		}
		if err != nil {
			t.Fatalf("Tversion encode: %v", err)
		}

		m, err := dec.Decode()
		if err != nil {
			t.Fatalf("Tversion decode: %v", err)
		}
		var rx versionInterface
		if isTversion {
			rx = m.(Tversion)
		} else {
			rx = m.(Rversion)
		}
		if tx.tag != rx.Tag() {
			t.Fatalf("Tversion: expected tag %d, got %d", tx.tag, rx.Tag())
		}
		if tx.msize != rx.Msize() {
			t.Fatalf("Tversion: expected msize %d, got %d", tx.msize, rx.Msize())
		}
		if tx.version != rx.Version() {
			t.Fatalf("Tversion: expected version %q, got %q", tx.version, rx.Version())
		}
	}
}

func TestVersion(t *testing.T) {
	testVersion(t, true)
	testVersion(t, false)
}
