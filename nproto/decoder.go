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
	r       *bufio.Reader // stream reader
	err     error
	maxSize int64
	buf     [headerLen]byte
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

	size, _, err := gheader(d.buf[:headerLen])
	if err != nil {
		d.discard(int64(size) - headerLen)
		return err
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
	m.Type, b = b[4], b[5:]
	m.Tag, b = guint16(b)

	switch m.Type {
	case Tversion, Rversion:
		m.Msize, b = guint32(b)
		m.Version, b = gstring(b)
	case Tauth:
		m.Afid, b = guint32(b)
		m.Uname, b = gstring(b)
		m.Aname, b = gstring(b)
	case Tattach:
		m.Fid, b = guint32(b)
		m.Afid, b = guint32(b)
		m.Uname, b = gstring(b)
		m.Aname, b = gstring(b)
	case Rauth, Rattach:
		b = m.Qid.unmarshal(b)
	case Rerror:
		m.Ename, b = gstring(b)
	case Tflush:
		m.Oldtag, b = guint16(b)
	case Twalk:
		m.Fid, b = guint32(b)
		m.Newfid, b = guint32(b)
		n, b := guint16(b)
		if n > maxWalkNames {
			return errMaxWalkNames
		}
		m.Wname = make([]string, n)
		for i := range m.Wname {
			m.Wname[i], b = gstring(b)
		}
	case Rwalk:
		n, b := guint16(b)
		if n > maxWalkNames {
			return errMaxWalkNames
		}
		m.Wqid = make([]Qid, n)
		for i := range m.Wqid {
			b = m.Wqid[i].unmarshal(b)
		}
	case Topen:
		m.Fid, b = guint32(b)
		m.Mode = b[0]
	case Tcreate:
		m.Fid, b = guint32(b)
		m.Name, b = gstring(b)
		m.Perm, b = guint32(b)
		m.Mode = b[0]
	case Ropen, Rcreate:
		b = m.Qid.unmarshal(b)
		m.Iounit, b = guint32(b)
	case Tread:
		m.Fid, b = guint32(b)
		m.Offset, b = guint64(b)
		m.Count, b = guint32(b)
	case Rread:
		m.Count, b = guint32(b)
		m.Data = b // TODO
	case Twrite:
		m.Fid, b = guint32(b)
		m.Offset, b = guint64(b)
		m.Count, b = guint32(b)
		m.Data = b // TODO
	case Rwrite:
		m.Count, b = guint32(b)
	case Tclunk, Tremove, Tstat:
		m.Fid, b = guint32(b)
	case Rstat:
		b = m.Stat.unmarshal(b[4:])
		return nil
	case Twstat:
		m.Fid, b = guint32(b)
		b = m.Stat.unmarshal(b[4:])
		return nil
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
