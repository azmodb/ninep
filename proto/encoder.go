package proto

import (
	"bufio"
	"io"
	"math"

	"github.com/azmodb/ninep/binary"
)

// Encoder manages the transmission of type and data information to the
// other side of a connection. Encoder is safe for concurrent use by
// multiple goroutines.
type Encoder struct {
	MaxMessageSize uint32

	buf *binary.Buffer
	w   *bufio.Writer
	err error
}

// NewEncoder returns a new encoder that will transmit on the io.Writer.
// Msize defines the maximum message size if greater than zero.
func NewEncoder(w io.Writer, msize uint32) *Encoder {
	return &Encoder{
		buf: binary.NewBuffer(nil),
		w:   bufio.NewWriter(w),

		MaxMessageSize: msize,
	}
}

// Reset discards any unflushed buffered data, clears any error, and
// resets Encoder to write its output to w.
func (e *Encoder) Reset(w io.Writer) {
	e.w = bufio.NewWriter(w)
	e.buf.Reset()
	e.err = nil
}

func (e *Encoder) maxMessageSize() uint32 {
	if e.MaxMessageSize <= 0 || e.MaxMessageSize > MaxMessageSize {
		return MaxMessageSize
	}
	return e.MaxMessageSize
}

func (e *Encoder) encHeader(size uint32, h *Header) error {
	e.buf.Reset() // reset message buffer

	e.buf.PutUint32(size)
	h.Encode(e.buf)
	if err := e.buf.Err(); err != nil {
		return err
	}
	return e.write(e.buf.Bytes())
}

func (e *Encoder) encMessage(m Message) error {
	e.buf.Reset() // reset message buffer

	m.Encode(e.buf)
	if err := e.buf.Err(); err != nil {
		return err
	}
	return e.write(e.buf.Bytes())
}

func (e *Encoder) encPayload(data []byte) error {
	e.buf.Reset() // reset message buffer

	e.buf.PutUint32(uint32(len(data)))
	e.write(e.buf.Bytes())
	e.write(data)
	return e.err
}

// Encode encodes a 9P header and message to the connection.
func (e *Encoder) Encode(h *Header, m Message) error {
	if e.err != nil {
		return e.err
	}
	if HeaderSize+m.Len() > math.MaxUint32 {
		return ErrMessageTooLarge
	}

	size := HeaderSize + uint32(m.Len())
	if size > e.maxMessageSize() {
		return ErrMessageTooLarge
	}

	err := e.encHeader(size, h)
	if err != nil {
		return err
	}

	if err = e.encMessage(m); err != nil {
		return err
	}
	if p, ok := m.(Payloader); ok {
		e.encPayload(p.Payload())
	}

	return e.flush()
}

func (e *Encoder) write(v []byte) error {
	if e.err != nil {
		return e.err
	}
	if e.w != nil {
		_, e.err = e.w.Write(v)
	}
	return e.err
}

func (e *Encoder) flush() error {
	if e.err != nil {
		if e.w != nil {
			e.w.Flush()
		}
		return e.err
	}
	if e.w != nil {
		e.err = e.w.Flush()
	}
	return e.err
}
