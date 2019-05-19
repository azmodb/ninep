package proto

import (
	"bytes"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/azmodb/ninep/binary"
)

func testEncoderMaxMessageSize(t *testing.T, e *Encoder, msize uint32, want error) {
	t.Helper()

	if err := e.encHeader(msize, 0, 0); err != want {
		t.Errorf("encoder MaxMessageSize: expected error %q, got %q", want, err)
	}
	if err := e.flush(); err != nil {
		t.Errorf("encoder MaxMessageSize: encode header: %v", err)
	}
}

func TestDecoderHeaderMessageSize(t *testing.T) {
	e := &Encoder{buf: binary.NewBuffer(nil)}

	e.MaxMessageSize = 0
	testEncoderMaxMessageSize(t, e, MaxMessageSize+1, ErrMessageTooLarge)
	testEncoderMaxMessageSize(t, e, MaxMessageSize, nil)
	testEncoderMaxMessageSize(t, e, MinMessageSize, nil)

	e.MaxMessageSize = MaxMessageSize
	testEncoderMaxMessageSize(t, e, MaxMessageSize+1, ErrMessageTooLarge)
	testEncoderMaxMessageSize(t, e, MaxMessageSize, nil)
	testEncoderMaxMessageSize(t, e, MinMessageSize, nil)

	e.MaxMessageSize = MaxMessageSize + 1
	testEncoderMaxMessageSize(t, e, MaxMessageSize+1, ErrMessageTooLarge)

	e.MaxMessageSize = MinMessageSize
	testEncoderMaxMessageSize(t, e, MinMessageSize+1, ErrMessageTooLarge)
	testEncoderMaxMessageSize(t, e, MinMessageSize, nil)
}

func TestHeaderEncoding(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf, 0)
	dec := NewDecoder(buf, 0)

	for n, test := range []struct {
		size uint32
		t    MessageType
		tag  uint16
	}{
		{MaxMessageSize, math.MaxUint8, math.MaxUint16},
		{8192, 100, math.MaxUint16},
		{MinMessageSize, 0, 0},
	} {
		if err := enc.encHeader(test.size, uint8(test.t), test.tag); err != nil {
			t.Fatalf("header(%d): unexpected header encoding error: %v", n, err)
		}
		if err := enc.flush(); err != nil {
			t.Fatalf("header(%d): unexpected header flushing error: %v", n, err)
		}

		dec.Reset(buf)
		v, tag, err := dec.DecodeHeader()
		if err != nil {
			t.Errorf("header(%d): unexpected header decoding error: %v", n, err)
		}

		if v != test.t {
			t.Errorf("header(%d): expected message type %d, got %d", n, test.t, v)
		}
		if tag != test.tag {
			t.Errorf("header(%d): expected tag %d, got %d", n, tag, test.tag)
		}
	}
}

func TestMessageEncoding(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf, 0)
	dec := NewDecoder(buf, 0)

	for n, test := range []struct {
		Type MessageType
		Tag  uint16
		M    Message
		R    Message
	}{
		{
			Type: MessageTversion, Tag: math.MaxUint16,
			M: &Tversion{math.MaxUint32, "9P2000.L"},
			R: &Tversion{},
		},
		{
			Type: MessageTversion, Tag: math.MaxUint16,
			M: &Tversion{8192, "9P2000.L"},
			R: &Tversion{},
		},
		{
			Type: MessageRversion, Tag: math.MaxUint16,
			M: &Rversion{math.MaxUint32, "9P2000.L"},
			R: &Rversion{},
		},
		{
			Type: MessageTversion, Tag: 0,
			M: &Tversion{0, ""},
			R: &Tversion{},
		},
	} {
		if err := enc.Encode(test.Tag, test.M); err != nil {
			t.Fatalf("msg(%d): unexpected message encoding error: %v", n, err)
		}

		v, tag, err := dec.DecodeHeader()
		if err != nil {
			t.Fatalf("msg(%d): unexpected message decoding error: %v", n, err)
		}

		if v != test.Type {
			t.Errorf("msg(%d): expected message type %d, got %d", n, test.Type, v)
		}
		if tag != test.Tag {
			t.Errorf("msg(%d): expected tag %d, got %d", n, test.Tag, tag)
		}

		if err := dec.Decode(test.R); err != nil {
			t.Fatalf("msg(%d): unexpected message decode error: %v", n, err)
		}
		if !reflect.DeepEqual(test.M, test.R) {
			t.Errorf("msg(%d): messages differ\n%+v\n+%v", n, test.M, test.R)
		}
	}
}

func TestPayloadEncoding(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf, 0)
	dec := NewDecoder(buf, 0)

	for n, test := range []struct {
		Type MessageType
		Tag  uint16
		M    Payloader
		R    Payloader
	}{
		{
			Type: MessageRreaddir, Tag: math.MaxUint16,
			M: &Rreaddir{},
			R: &Rreaddir{},
		},
		{
			Type: MessageRreaddir, Tag: math.MaxUint16,
			M: &Rreaddir{bytes16},
			R: &Rreaddir{},
		},

		{
			Type: MessageRread, Tag: math.MaxUint16,
			M: &Rread{},
			R: &Rread{},
		},
		{
			Type: MessageRread, Tag: math.MaxUint16,
			M: &Rread{bytes16},
			R: &Rread{},
		},

		{
			Type: MessageTwrite, Tag: math.MaxUint16,
			M: &Twrite{},
			R: &Twrite{},
		},
		{
			Type: MessageTwrite, Tag: math.MaxUint16,
			M: &Twrite{Fid: math.MaxUint32, Offset: math.MaxUint64},
			R: &Twrite{},
		},
		{
			Type: MessageTwrite, Tag: math.MaxUint16,
			M: &Twrite{Data: bytes16},
			R: &Twrite{},
		},
		{
			Type: MessageTwrite, Tag: math.MaxUint16,
			M: &Twrite{Fid: math.MaxUint32, Offset: math.MaxUint64, Data: bytes16},
			R: &Twrite{},
		},
	} {
		if err := enc.Encode(test.Tag, test.M); err != nil {
			t.Fatalf("msg(%d): unexpected message encoding error: %v", n, err)
		}

		v, tag, err := dec.DecodeHeader()
		if err != nil {
			t.Fatalf("msg(%d): unexpected message decoding error: %v", n, err)
		}

		if v != test.Type {
			t.Errorf("msg(%d): expected message type %d, got %d", n, test.Type, v)
		}
		if tag != test.Tag {
			t.Errorf("msg(%d): expected tag %d, got %d", n, test.Tag, tag)
		}

		if err := dec.Decode(test.R); err != nil {
			t.Fatalf("msg(%d): unexpected message decode error: %v", n, err)
		}
		if !reflect.DeepEqual(test.M, test.R) {
			t.Errorf("msg(%d): messages differ\n%+v\n+%v", n, test.M, test.R)
		}
	}
}

func TestDecoderMessageTooSmall(t *testing.T) {
	data := make([]byte, 0, HeaderSize)
	data = binary.PutUint32(data, HeaderSize-1)
	data = data[:HeaderSize]

	buf := bytes.NewBuffer(data)
	dec := NewDecoder(buf, 0)

	_, _, err := dec.DecodeHeader()
	if err != ErrMessageTooSmall {
		t.Fatalf("decoder: expected ErrMessageTooSmall, got %T", err)
	}
}

func TestDecoderMessageTooLarge(t *testing.T) {
	maxMessageSize := uint32(32)

	data := make([]byte, 0, HeaderSize)
	data = binary.PutUint32(data, maxMessageSize+1)
	data = data[:HeaderSize]

	buf := bytes.NewBuffer(data)
	dec := NewDecoder(buf, maxMessageSize)

	_, _, err := dec.DecodeHeader()
	if err != ErrMessageTooLarge {
		t.Fatalf("decoder: expected ErrMessageTooLarge, got %v", err)
	}
}

var errTest = errors.New("test error")

func TestErrDecoder(t *testing.T) {
	dec := NewDecoder(nil, 0)
	dec.err = errTest

	if _, _, err := dec.DecodeHeader(); err != errTest {
		t.Errorf("decoder: expected decode error, got %v", err)
	}
	if err := dec.Decode(nil); err != errTest {
		t.Errorf("decoder: expected decode error, got %v", err)
	}
}

func TestNilMessageDecode(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf, 0)
	dec := NewDecoder(buf, 0)

	m := Tversion{MessageSize: 8192, Version: Version}
	if err := enc.Encode(42, &m); err != nil {
		t.Fatalf("encode: %v", err)
	}

	if _, _, err := dec.DecodeHeader(); err != nil {
		t.Fatalf("decode header: %v", err)
	}
	if err := dec.Decode(nil); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("decode nil message: expected empty buffer")
	}
}
