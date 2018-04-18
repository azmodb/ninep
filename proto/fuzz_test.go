package proto

import (
	"strings"
	"testing"
)

func TestCrashers(t *testing.T) {
	crashers := []string{
	//		"#\x00usr\x06\x00glenda\x03\x10\x00ib",
	//		"\x1c\x00bootes\x03\x00tmp",
	//		"0000x39",
	//		"t0\x85\xe9tr\x85",
	//		"0i\x80tpui",
	}

	for _, data := range crashers {
		dec := NewDecoder(strings.NewReader(data))
		m, err := dec.Decode()
		if err == nil {
			t.Fatalf("fuzz: expected non-nil error, got %v", err)
		}
		if m != nil {
			t.Fatalf("fuzz: expected nil message, got %+v", m)
		}
	}
}
