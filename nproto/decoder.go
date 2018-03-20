package ninep

import (
	"bufio"
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
type Decoder struct {
	r   *bufio.Reader // stream reader
	err error
	buf [headerLen]byte
}

// NewDecoder returns a new decoder that reads from the io.Reader.
func NewDecoder(r io.Reader, opts ...Option) *Decoder {
	d := &Decoder{r: bufio.NewReader(r)}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// Decode reads the next value from the input stream and stores it in
// the data represented by Message. If m is nil, the value will be
// discarded.
func (d *Decoder) Decode(m *Message) error {
	if err := d.readFull(d.buf[:headerLen]); err != nil {
		return err
	}
	size := int64(guint32(d.buf[:4]))
	if size < headerLen {
		return errMessageTooSmall
	}
	//if size > d.maxSize {
	//	return errMessageTooLarge
	//}

	typ := d.buf[4]
	if typ == Terror || typ < Tversion || typ > Rwstat {
		d.discard(size - headerLen)
		return errInvalidMessageType
	}
	if size < minSizeLUT64[typ-100] {
		d.discard(size - headerLen)
		return errMessageTooSmall
	}

	data := make([]byte, size)
	copy(data[:headerLen], d.buf[:headerLen])
	if size > headerLen {
		if err := d.readFull(data[headerLen:]); err != nil {
			return err
		}
	}
	return unmarshal(data, m)
}

func unmarshal(b []byte, m *Message) error {
	m.Type = b[4]
	m.Tag = guint16(b[5:7])

	switch m.Type {
	case Tversion, Rversion:
		m.Msize = guint32(b[7:11])
		gstring(b[11:], &m.Version)
	case Tauth:
		m.Afid = guint32(b[7:11])
		n := 11 + gstring(b[11:], &m.Uname)
		gstring(b[n:], &m.Aname)
	case Tattach:
		m.Fid = guint32(b[7:11])
		m.Afid = guint32(b[11:15])
		n := 15 + gstring(b[15:], &m.Uname)
		gstring(b[n:], &m.Aname)
	case Rauth, Rattach:
		m.Qid.unmarshal(b[7:20])
	case Rerror:
		gstring(b[7:], &m.Ename)
	case Tflush:
		m.Oldtag = guint16(b[7:9])
	case Twalk:
		m.Fid = guint32(b[7:11])
	case Rwalk:
	case Topen:
		m.Fid = guint32(b[7:11])
		m.Mode = b[11]
	case Tcreate:
		m.Fid = guint32(b[7:11])
		n := 11 + gstring(b[11:], &m.Name)
		m.Perm = guint32(b[n:])
		m.Mode = b[n+4]
	case Ropen, Rcreate:
		m.Qid.unmarshal(b[7:20])
		m.Iounit = guint32(b[20:24])
	case Tread:
		m.Fid = guint32(b[7:11])
		m.Offset = guint64(b[11:19])
		m.Count = guint32(b[19:23])
	case Rread:
		m.Count = guint32(b[7:11])
		m.Data = gdata(b[11:], m.Data)
	case Twrite:
		m.Fid = guint32(b[7:11])
		m.Offset = guint64(b[11:19])
		m.Count = guint32(b[19:23])
		m.Data = gdata(b[23:], m.Data)
	case Rwrite:
		m.Count = guint32(b[7:11])
	case Tclunk, Tremove, Tstat:
		m.Fid = guint32(b[7:11])
	case Rstat:
		m.Stat.unmarshal(b[7:])
	case Twstat:
		m.Fid = guint32(b[7:11])
		m.Stat.unmarshal(b[11:])
	case Rflush, Rclunk, Rremove, Rwstat:
		// nothing
	}
	return nil
}

func (d *Decoder) readFull(buf []byte) error {
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

func (d *Decoder) discard(size int64) error {
	if d.err != nil {
		return d.err
	}
	if size <= 0 {
		return nil
	}
	_, d.err = io.CopyN(devNull{}, d.r, size)
	return d.err
}
