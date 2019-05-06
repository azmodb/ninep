package proto

import (
	"bufio"
	"io"
	"io/ioutil"

	"github.com/azmodb/ninep/binary"
)

// Decoder manages the receipt of type and data information read from the
// remote side of a connection.
type Decoder struct {
	header  [HeaderSize]byte
	r       io.Reader
	buf     *binary.Buffer
	pending int
	err     error

	MaxMessageSize uint32
}

// NewDecoder returns a new decoder that decodes from the io.Reader.
// Msize defines the maximum message size if greater than zero.
func NewDecoder(r io.Reader, msize uint32) *Decoder {
	return &Decoder{
		buf: binary.NewBuffer(nil),
		r:   bufio.NewReader(r),

		MaxMessageSize: msize,
	}
}

// Reset discards any buffered data, resets all state, and switches the
// Decoder to read from r.
func (d *Decoder) Reset(r io.Reader) {
	d.r = bufio.NewReader(r)
	d.buf.Reset()
	d.pending = 0
	d.err = nil
}

func (d *Decoder) maxMessageSize() uint32 {
	if d.MaxMessageSize <= 0 || d.MaxMessageSize > MaxMessageSize {
		return MaxMessageSize
	}
	return d.MaxMessageSize
}

const maxInt = int64(^uint(0) >> 1)

// DecodeHeader decodes the next 9P header from the input stream.
// DecodeHeader and Decode must be called in pairs to read 9P messages
// from the connection.
func (d *Decoder) DecodeHeader(h *Header) error {
	if err := d.readFull(d.header[:]); err != nil {
		return err
	}

	size := binary.Uint32(d.header[:4])
	if size < HeaderSize {
		return ErrMessageTooSmall
	}
	if int64(size) > maxInt || size > d.maxMessageSize() {
		d.discard(size - HeaderSize)
		return ErrMessageTooLarge
	}

	d.pending = int(size) - HeaderSize
	if h != nil {
		h.Type = MessageType(d.header[4])
		h.Tag = binary.Uint16(d.header[5:7])
	}
	return nil
}

// Decode decodes the next 9P message from the input stream and stores it
// in the data represented by the empty interface m. If m is nil, the
// value will be discarded. Otherwise, the value underlying m must be a
// pointer to the correct type for the next 9P message received. If the
// input is at EOF, Decode returns io.EOF and does not modify e.
func (d *Decoder) Decode(m Message) error {
	if d.err != nil {
		return d.err
	}
	if d.pending <= 0 {
		return nil
	}

	if p, ok := m.(Payloader); ok {
		return d.decPayload(p)
	}
	return d.decMessage(d.pending, m)
}

func (d *Decoder) decMessage(size int, m Message) error {
	if d.err != nil {
		return d.err
	}

	if err := d.buf.Read(d.r, size); err != nil {
		d.err = err
		return err
	}

	if m == nil {
		return d.err
	}

	m.Decode(d.buf)
	return d.buf.Err()
}

func (d *Decoder) decPayload(p Payloader) (err error) {
	fixed := p.FixedLen()
	if fixed > 0 {
		if err = d.decMessage(fixed, p); err != nil {
			return err
		}
	}

	// read payload length into header buffer
	if err = d.readFull(d.header[:4]); err != nil {
		return err
	}
	size := binary.Uint32(d.header[:4])
	if size > 0 {
		payload := make([]byte, size)
		if err = d.readFull(payload); err != nil {
			payload = nil
			return err
		}
		p.PutPayload(payload)
	}
	return err
}

func (d *Decoder) readFull(buf []byte) error {
	if d.err != nil {
		return d.err
	}
	_, d.err = io.ReadFull(d.r, buf)
	return d.err
}

func (d *Decoder) discard(size uint32) {
	if d.err != nil {
		return
	}
	if size <= 0 {
		return
	}
	_, d.err = io.CopyN(ioutil.Discard, d.r, int64(size))
}
