package proto

import (
	"io"
	"sync"
)

// Decoder provides an interface for reading a stream of 9P2000 messages
// from an io.Reader. Successive calls to the Decode method of a Decoder
// will fetch and validate 9P2000 messages from the input stream, until
// EOF is encountered, or another error is encountered.
//
// Decoder is not safe for concurrent use. Usage of any Decoder method
// should be delegated to a single thread of execution or protected by a
// mutex.
type Decoder interface {
	// Decode reads the next value from the input stream and stores it
	// in the data represented by Message. If m is nil, the value will
	// be discarded.
	Decode() (Message, error)
}

type decoder struct {
	buf [headerLen]byte
	r   io.Reader
	err error
}

// NewDecoder returns a new decoder that reads from the io.Reader.
func NewDecoder(r io.Reader, opts ...Option) Decoder {
	return newDecoder(r, opts...)
}

func newDecoder(r io.Reader, opts ...Option) *decoder {
	return &decoder{r: r}
}

func (d *decoder) Decode() (Message, error) {
	if err := d.readFull(d.buf[:headerLen]); err != nil {
		return nil, err
	}

	size, _, err := gheader(d.buf[:headerLen])
	if err != nil {
		d.discard(int64(size) - headerLen)
		return nil, err
	}

	data := make([]byte, size)
	copy(data[:headerLen], d.buf[:headerLen])
	if size > headerLen {
		if err = d.readFull(data[headerLen:]); err != nil {
			return nil, err
		}
	}
	//return unmarshal(data, m)
	return nil, nil
}

func (d *decoder) readFull(buf []byte) error {
	if d.err != nil {
		return d.err
	}
	_, d.err = io.ReadFull(d.r, buf)
	return d.err
}

func (d *decoder) discard(size int64) error {
	if d.err != nil {
		return d.err
	}
	if size <= 0 {
		return nil
	}
	_, d.err = io.CopyN(devNull{}, d.r, size)
	return d.err
}

type devNull struct{}

var devNullPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 8192)
		return &b
	},
}

func (devNull) Write(p []byte) (int, error) {
	return len(p), nil
}

func (devNull) ReadFrom(r io.Reader) (int64, error) {
	bufp := devNullPool.Get().(*[]byte)
	defer devNullPool.Put(bufp)

	var n int64
	for {
		m, err := r.Read(*bufp)
		n += int64(m)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return n, err
		}
	}
}
