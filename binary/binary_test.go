package binary

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"
)

var (
	string16 strings.Builder
	bytes16  []byte
)

func init() {
	string16 = strings.Builder{}
	for i := 0; i < math.MaxUint16; i++ {
		string16.WriteByte('a' + byte(i%26))
	}
	bytes16 = []byte(string16.String())
}

func TestConsume(t *testing.T) {
	for n, test := range []struct {
		data []byte
		size int

		want []byte
		buf  []byte
		ok   bool
	}{
		{[]byte("abcd"), 4, []byte(""), []byte("abcd"), true},
		{[]byte("abcd"), 3, []byte("d"), []byte("abc"), true},
		{[]byte("abcd"), 2, []byte("cd"), []byte("ab"), true},
		{[]byte("abcd"), 1, []byte("bcd"), []byte("a"), true},
		{[]byte("abcd"), 0, []byte("abcd"), []byte(""), true},

		{[]byte("abcd"), 5, []byte("abcd"), nil, false},
		{[]byte("abcd"), 6, []byte("abcd"), nil, false},
		{[]byte("abcd"), 7, []byte("abcd"), nil, false},
	} {
		data, buf, ok := consume(test.data, test.size)
		if ok != test.ok {
			t.Fatalf("consume(%d): expected result %v, have %v", n, test.ok, ok)
		}
		if !bytes.Equal(data, test.want) {
			t.Fatalf("consume(%d): expected data %q, got %q", n, test.want, data)
		}
		if !bytes.Equal(buf, test.buf) {
			t.Fatalf("consume(%d): expected buf %q, got %q", n, test.want, data)
		}
	}
}

func TestGrow(t *testing.T) {
	a := allocator{0, 8}

	for n, test := range []struct {
		data []byte
		size int

		want []byte
		off  int
	}{
		{make([]byte, 2), 8, make([]byte, 10, 10), 2},
		{make([]byte, 1), 8, make([]byte, 9, 9), 1},

		{make([]byte, 2), 4, make([]byte, 6, 8), 2},
		{make([]byte, 2), 3, make([]byte, 5, 7), 2},
		{make([]byte, 2), 2, make([]byte, 4, 6), 2},
		{make([]byte, 2), 1, make([]byte, 3, 5), 2},
		{make([]byte, 2), 0, make([]byte, 2, 2), 2},

		{make([]byte, 1), 4, make([]byte, 5, 6), 1},
		{make([]byte, 1), 3, make([]byte, 4, 5), 1},
		{make([]byte, 1), 2, make([]byte, 3, 4), 1},
		{make([]byte, 1), 1, make([]byte, 2, 3), 1},
		{make([]byte, 1), 0, make([]byte, 1, 1), 1},

		{[]byte{}, 4, make([]byte, 4), 0},
		{nil, 4, make([]byte, 4), 0},
	} {
		data, off := a.Grow(test.data, test.size)
		if len(data) != len(test.want) {
			t.Fatalf("grow(%d): expected len %d, got %d", n, len(test.want), len(data))
		}
		if cap(data) != cap(test.want) {
			t.Fatalf("grow(%d): expected cap %d, got %d", n, cap(test.want), cap(data))
		}
		if !bytes.Equal(data, test.want) {
			t.Fatalf("grow(%d): expected data %q, got %q", n, test.want, data)
		}
		if off != test.off {
			t.Fatalf("grow(%d): expected offset %d, got %d", n, test.off, off)
		}
	}
}

func sizeof(v interface{}) int {
	switch v := v.(type) {
	case uint64, *uint64:
		return 8
	case uint32, *uint32:
		return 4
	case uint16, *uint16:
		return 2
	case uint8, *uint8:
		return 1
	case string:
		return 2 + len(v)
	case *string:
		return 2 + len(*v)
	case testStruct, *testStruct:
		return 15
	}
	panic("sizeof: unsupported type")
}

func testCodec(t *testing.T, what string, src, dst interface{}) {
	size := sizeof(src)
	buf, n, err := Marshal(nil, src)
	if err != nil {
		t.Fatalf("marshal(%s): unexpected error: %v", what, err)
	}
	if n != size {
		t.Fatalf("marshal(%s): expected %d bytes written, got %d: ", what, size, n)
	}
	if len(buf) != size {
		t.Fatalf("marshal(%s): expected %d bytes, got %d: ", what, size, len(buf))
	}

	buf, err = Unmarshal(buf, dst)
	if err != nil {
		t.Fatalf("unmarshal(%s): unexpected error: %v", what, err)
	}
	if len(buf) != 0 {
		t.Fatalf("unmarshal(%s): expected empty buffer, got %d: ", what, len(buf))
	}
}

func checkValue(t *testing.T, what string, src, dst interface{}) {
	success := false
	switch src := src.(type) {
	case uint64:
		success = src == dst.(uint64)
	case uint32:
		success = src == dst.(uint32)
	case uint16:
		success = src == dst.(uint16)
	case uint8:
		success = src == dst.(uint8)
	case string:
		success = src == dst.(string)
	case *testStruct:
		success = reflect.DeepEqual(src, dst)
	}
	if !success {
		t.Fatalf("unmarshal(%s): expected value %v, got %v: ", what, src, dst)
	}
}

func TestUint64(t *testing.T) {
	for _, src := range []uint64{0, 42, math.MaxUint64} {
		var dst uint64

		testCodec(t, "uint64", src, &dst)
		checkValue(t, "uint64", src, dst)

		testCodec(t, "*uint64", &src, &dst)
		checkValue(t, "*uint64", src, dst)
	}
}

func TestUint32(t *testing.T) {
	for _, src := range []uint32{0, 42, math.MaxUint32} {
		var dst uint32

		testCodec(t, "uint32", src, &dst)
		checkValue(t, "uint32", src, dst)

		testCodec(t, "*uint32", &src, &dst)
		checkValue(t, "*uint32", src, dst)
	}
}

func TestUint16(t *testing.T) {
	for _, src := range []uint16{0, 42, math.MaxUint16} {
		var dst uint16

		testCodec(t, "uint16", src, &dst)
		checkValue(t, "uint16", src, dst)

		testCodec(t, "*uint16", &src, &dst)
		checkValue(t, "*uint16", src, dst)
	}
}

func TestUint8(t *testing.T) {
	for _, src := range []uint8{0, 42, math.MaxUint8} {
		var dst uint8

		testCodec(t, "uint8", src, &dst)
		checkValue(t, "uint8", src, dst)

		testCodec(t, "*uint8", &src, &dst)
		checkValue(t, "*uint8", src, dst)
	}
}

func TestString(t *testing.T) {
	for _, src := range []string{"", "hello world", string16.String()} {
		var dst string

		testCodec(t, "string", src, &dst)
		checkValue(t, "string", src, dst)

		testCodec(t, "*string", &src, &dst)
		checkValue(t, "*string", src, dst)
	}
}

type testStruct struct {
	Uint8  uint8
	Uint16 uint16
	Uint32 uint32
	Uint64 uint64
}

func (s testStruct) Marshal(data []byte) ([]byte, int, error) {
	data = PutUint8(data, s.Uint8)
	data = PutUint16(data, s.Uint16)
	data = PutUint32(data, s.Uint32)
	data = PutUint64(data, s.Uint64)
	return data, 15, nil
}

func (s *testStruct) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 15 {
		return data, io.ErrUnexpectedEOF
	}
	data, _ = guint8(data, &s.Uint8)
	data, _ = guint16(data, &s.Uint16)
	data, _ = guint32(data, &s.Uint32)
	data, _ = guint64(data, &s.Uint64)
	return data, nil
}

func TestMarshalerUnmarshaler(t *testing.T) {
	for _, src := range []testStruct{
		{math.MaxUint8, math.MaxUint16, math.MaxUint32, math.MaxUint64},
		{0, 0, 0, 0},
		{},
	} {
		var dst testStruct

		testCodec(t, "struct", &src, &dst)
		checkValue(t, "struct", &src, &dst)
	}
}

func testEncode(t *testing.T, buf *Buffer, v interface{}) {
	t.Helper()

	switch val := v.(type) {
	case string:
		buf.PutString(val)
	case uint64:
		buf.PutUint64(val)
	case uint32:
		buf.PutUint32(val)
	case uint16:
		buf.PutUint16(val)
	case uint8:
		buf.PutUint8(val)
	}
}

func checkBufferValue(t *testing.T, buf *Buffer, want interface{}) {
	t.Helper()

	switch want := want.(type) {
	case string:
		v := buf.String()
		if v != want {
			t.Errorf("buffer: expected string %q, got %q", want, v)
		}
	case uint64:
		v := buf.Uint64()
		if v != want {
			t.Errorf("buffer: expected uint64 %q, got %q", want, v)
		}
	case uint32:
		v := buf.Uint32()
		if v != want {
			t.Errorf("buffer: expected uint32 %q, got %q", want, v)
		}
	case uint16:
		v := buf.Uint16()
		if v != want {
			t.Errorf("buffer: expected uint16 %q, got %q", want, v)
		}
	case uint8:
		v := buf.Uint8()
		if v != want {
			t.Errorf("buffer: expected uint8 %q, got %q", want, v)
		}
	}
}

func TestBuffer(t *testing.T) {
	buf := &Buffer{}

	for _, test := range []struct {
		val interface{}
	}{
		{""}, {"hello world"}, {string16.String()},

		{uint64(0)}, {uint64(42)}, {uint64(math.MaxUint64)},
		{uint32(0)}, {uint32(42)}, {uint32(math.MaxUint32)},
		{uint16(0)}, {uint16(42)}, {uint16(math.MaxUint16)},
		{uint8(0)}, {uint8(42)}, {uint8(math.MaxUint8)},
	} {
		testEncode(t, buf, test.val)
		checkBufferValue(t, buf, test.val)
	}
}

func checkZero(t *testing.T, v interface{}) {
	switch val := v.(type) {
	case string:
		if val != "" {
			t.Errorf("buffer: expected empty string, got %q", val)
		}
	case uint64:
		if val != 0 {
			t.Errorf("buffer: expected zero uint64, got %q", val)
		}
	case uint32:
		if val != 0 {
			t.Errorf("buffer: expected zero uint32, got %q", val)
		}
	case uint16:
		if val != 0 {
			t.Errorf("buffer: expected zero uint16, got %q", val)
		}
	case uint8:
		if val != 0 {
			t.Errorf("buffer: expected zero uint8, got %q", val)
		}
	}
}

func TestNilBuffer(t *testing.T) {
	buf := &Buffer{}

	buf.String() // initial call

	checkZero(t, buf.String())
	checkZero(t, buf.Uint64())
	checkZero(t, buf.Uint32())
	checkZero(t, buf.Uint16())
	checkZero(t, buf.Uint8())
	buf.Reset()

	checkZero(t, buf.String())
	buf.Reset()
	checkZero(t, buf.Uint64())
	buf.Reset()
	checkZero(t, buf.Uint32())
	buf.Reset()
	checkZero(t, buf.Uint16())
	buf.Reset()
	checkZero(t, buf.Uint8())
	buf.Reset()

	buf.data = []byte{}

	checkZero(t, buf.String())
	buf.Reset()
	checkZero(t, buf.Uint64())
	buf.Reset()
	checkZero(t, buf.Uint32())
	buf.Reset()
	checkZero(t, buf.Uint16())
	buf.Reset()
	checkZero(t, buf.Uint8())
	buf.Reset()
}

func TestStringFunc(t *testing.T) {
	v := String(PutString(nil, ""))
	checkValue(t, "string", "", v)

	v = String(PutString(nil, "hello world"))
	checkValue(t, "string", "hello world", v)

	v = String(PutString(nil, string16.String()))
	checkValue(t, "string", string16.String(), v)
}

func TestBufferRead(t *testing.T) {
	buf := NewBuffer(nil)

	for _, r := range []struct {
		r   io.Reader
		len int
		err error
	}{
		{bytes.NewReader([]byte("hello world")), len("hello world"), nil},
		{bytes.NewReader(bytes16), len(bytes16), nil},
		{},

		{bytes.NewReader(bytes16), 42, nil},
		{bytes.NewReader(bytes16), 0, nil},

		{bytes.NewReader(bytes16), math.MaxUint16 + 1, io.ErrUnexpectedEOF},

		{bytes.NewReader(nil), 42, io.EOF},
	} {
		buf.Reset()

		if err := buf.Read(r.r, r.len); err != r.err {
			t.Errorf("read: expected read error %v: got %v", r.err, err)
			continue
		}

		if buf.Len() != r.len {
			t.Errorf("read: expected buffer length %d, got %d", r.len, buf.Len())
		}
	}
}
