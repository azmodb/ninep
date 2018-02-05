package proto

import (
	"bufio"
	"fmt"
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
	Decode() (Message, error)
	Reset(r io.Reader)
}

type DecoderOption func(*decoder)

const headerLen = 7

type decoder struct {
	r   *bufio.Reader // stream reader
	err error
	buf [headerLen]byte

	maxSize  int64
	dataSize int64
}

func NewDecoder(r io.Reader, opts ...DecoderOption) Decoder {
	return newDecoder(r, opts...)
}

func newDecoder(r io.Reader, opts ...DecoderOption) *decoder {
	d := &decoder{}
	for _, opt := range opts {
		opt(d)
	}
	d.reset(r, defBufSize)
	return d
}

func (d *decoder) reset(r io.Reader, size int) {
	d.r = bufio.NewReaderSize(r, size)
	d.err = nil
}

func (d *decoder) Reset(r io.Reader) { d.reset(r, defBufSize) }

func (d *decoder) Decode() (Message, error) {
	if err := d.readFull(d.buf[:headerLen]); err != nil {
		return nil, err
	}
	size := int64(guint32(d.buf[:4]))
	if size < headerLen {
		return nil, errMessageTooSmall
	}
	if d.maxSize > 0 && size > d.maxSize {
		return nil, errMessageTooLarge
	}

	typ := d.buf[4]
	if typ == msgTerror || typ < msgTversion || typ > msgRwstat {
		d.discard(size - headerLen)
		return nil, errInvalidMessageType
	}
	if size < minSizeLUT64[typ-100] {
		d.discard(size - headerLen)
		return nil, errMessageTooSmall
	}

	data := make([]byte, size)
	copy(data[:headerLen], d.buf[:headerLen])
	if size > headerLen {
		if err := d.readFull(data[headerLen:]); err != nil {
			return nil, err
		}
	}
	return d.parse(typ, data)
}

func (d *decoder) parse(typ uint8, data []byte) (m Message, err error) {
	switch typ {
	default:
		panic(fmt.Sprintf("decoder: invalid type (%d)", typ))
	case msgTversion:
		m, err = parseTversion(data)
	case msgRversion:
		m, err = parseRversion(data)
	case msgTauth:
		m, err = parseTauth(data)
	case msgRauth:
		m = Rauth(data)
	case msgTattach:
		m, err = parseTattach(data)
	case msgRattach:
		m = Rattach(data)
	case msgTflush:
		m = Tflush(data)
	case msgRflush:
		m = Rflush(data)
	case msgTwalk:
		m, err = parseTwalk(data)
	case msgRwalk:
		m = Rwalk(data)
	case msgTopen:
		m = Topen(data)
	case msgRopen:
		m = Ropen(data)
	case msgTcreate:
		m, err = parseTcreate(data)
	case msgRcreate:
		m = Rcreate(data)
	case msgTread:
		m = Tread(data)
	case msgRread:
		m, err = parseRread(data, d.dataSize)
	case msgTwrite:
		m, err = parseTwrite(data, d.dataSize)
	case msgRwrite:
		m = Rwrite(data)
	case msgTclunk:
		m = Tclunk(data)
	case msgRclunk:
		m = Rclunk(data)
	case msgTremove:
		m = Tremove(data)
	case msgRremove:
		m = Rremove(data)
	case msgTstat:
		m = Tstat(data)
	case msgRstat:
		m, err = parseRstat(data)
	case msgTwstat:
		m, err = parseTwstat(data)
	case msgRwstat:
		m = Rwstat(data)
	case msgRerror:
		m = Rerror(data)
	}
	if err != nil {
		m.Reset()
		return nil, err
	}
	return m, err
}

func (d *decoder) readFull(buf []byte) error {
	if d.err != nil {
		return d.err
	}
	_, d.err = io.ReadFull(d.r, buf)
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
