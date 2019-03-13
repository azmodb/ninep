package proto

import (
	"bufio"
	"io"
	"sync"
)

var bufPool = &sync.Pool{New: func() interface{} { return &Buffer{} }}

// Buffer is a variable-sized buffer of bytes used to decode 9P2000
// messages.
type Buffer struct {
	data    []byte
	err     error
	managed bool
}

// NewBuffer creates and initializes a new Buffer using data as its
// initial contents. The new Buffer takes ownership of data, and the
// caller should not use data after this call.
func NewBuffer(data []byte) *Buffer {
	return &Buffer{data: data, managed: false}
}

func newBuffer(size int, managed bool) (b *Buffer) {
	if managed {
		b = bufPool.Get().(*Buffer)
	} else {
		b = &Buffer{}
	}

	b.managed = managed
	b.grow(size)

	return b
}

// Len returns the number of bytes of the unread portion of the buffer.
func (b *Buffer) Len() int { return len(b.data) }

// Release resets to be empty and releases the buffer.
func (b *Buffer) Release() {
	if b == nil {
		return
	}

	managed := b.managed
	b.data = b.data[:0]
	b.err = nil
	b.managed = false

	if managed {
		bufPool.Put(b)
	}
}

// Err returns the first error that was encountered by the Buffer.
func (b *Buffer) Err() error { return b.err }

const maxInt = int64(^uint(0) >> 1)

const bufBootstrapSize = 128

func (b *Buffer) grow(n int) (off int) {
	if b.data == nil {
		if n < bufBootstrapSize {
			b.data = make([]byte, n, bufBootstrapSize)
		} else {
			b.data = make([]byte, n)
		}
		return 0
	}

	off = len(b.data)
	if off+n > cap(b.data) {
		capacity := n + cap(b.data)*2
		if capacity > 4*1024*1024 {
			capacity = n + cap(b.data)
		}
		if int64(capacity) >= maxInt {
			panic("buffer too large")
		}

		data := make([]byte, off, capacity)
		copy(data[:off], b.data)
		b.data = data
	}

	b.data = b.data[:off+n]
	return off
}

// consume consumes n bytes from Buffer. If an error occurred it returns
// false.
func (b *Buffer) consume(n int) ([]byte, bool) {
	if b.err != nil {
		return nil, false
	}
	if len(b.data) < n {
		b.err = io.ErrUnexpectedEOF
		return nil, false
	}
	data := b.data[:n]
	b.data = b.data[n:]
	return data, true
}

func (b *Buffer) uint64() uint64 {
	data, ok := b.consume(8)
	if !ok {
		return 0
	}
	return order.Uint64(data)
}

func (b *Buffer) uint32() uint32 {
	data, ok := b.consume(4)
	if !ok {
		return 0
	}
	return order.Uint32(data)
}

func (b *Buffer) uint16() uint16 {
	data, ok := b.consume(2)
	if !ok {
		return 0
	}
	return order.Uint16(data)
}

func (b *Buffer) uint8() uint8 {
	data, ok := b.consume(1)
	if !ok {
		return 0
	}
	return data[0]
}

func (b *Buffer) string(max uint16, throw error) string {
	size := b.uint16()
	if size == 0 {
		return ""
	}

	if max > 0 && size > max {
		b.err = throw
		return ""
	}

	if uint16(len(b.data)) < size {
		b.err = io.ErrUnexpectedEOF
		return ""
	}

	v := string(b.data[:size])
	b.data = b.data[size:]
	return v
}

func (b *Buffer) raw32() []byte {
	count := b.uint32()
	if count > MaxDataSize || int64(count) > maxInt { // TODO
		b.err = ErrDataTooLarge
		return nil
	}
	if count == 0 {
		return nil
	}

	data, ok := b.consume(int(count))
	if !ok {
		return nil
	}
	return data
}

func (b *Buffer) raw16() []byte {
	count := b.uint16()
	if count > maxStatSize || count > 1<<16-1 {
		b.err = ErrStatTooLarge
		return nil
	}
	if count == 0 {
		return nil
	}

	data, ok := b.consume(int(count))
	if !ok {
		return nil
	}
	return data
}

func (b *Buffer) version() (int64, string) {
	return int64(b.uint32()), b.string(MaxVersionSize, ErrVersionTooLarge)
}

// Tversion decodes a Tversion message from Buffer.
func (b *Buffer) Tversion() (msize int64, version string) {
	msize, version = b.version()
	return
}

// Rversion decodes a Rversion message from Buffer.
func (b *Buffer) Rversion() (msize int64, version string) {
	msize, version = b.version()
	return
}

// Tauth decodes a Tauth message from Buffer.
func (b *Buffer) Tauth() (afid uint32, uname, aname string) {
	afid = b.uint32()
	uname = b.string(MaxUserNameSize, ErrUserNameTooLarge)
	aname = b.string(MaxPathSize, ErrPathTooLarge)
	return
}

// Tattach decodes a Tattach message from Buffer.
func (b *Buffer) Tattach() (fid, afid uint32, uname, aname string) {
	fid = b.uint32()
	afid = b.uint32()
	uname = b.string(MaxUserNameSize, ErrUserNameTooLarge)
	aname = b.string(MaxPathSize, ErrPathTooLarge)
	return
}

// Rauth decodes a Rauth message from Buffer.
func (b *Buffer) Rauth() (typ uint8, version uint32, path uint64) {
	typ = b.uint8()
	version = b.uint32()
	path = b.uint64()
	return
}

// Rattach decodes a Rattach message from Buffer.
func (b *Buffer) Rattach() (typ uint8, version uint32, path uint64) {
	typ = b.uint8()
	version = b.uint32()
	path = b.uint64()
	return
}

// Rerror decodes a Rerror message from Buffer.
func (b *Buffer) Rerror() string {
	return b.string(MaxEnameSize, ErrEnameTooLarge)
}

// Tflush decodes a Tflush message from Buffer.
func (b *Buffer) Tflush() uint16 { return b.uint16() }

// Rflush decodes a Rflush message from Buffer.
//func (b *Buffer) Rflush() uint16 { return b.uint16() }

// Twalk decodes a Twalk message from Buffer.
func (b *Buffer) Twalk() (fid, newfid uint32, names []string) {
	fid = b.uint32()
	newfid = b.uint32()
	n := b.uint16()
	if n > MaxWalkElements {
		b.err = ErrWalkElements
		return
	}
	if n == 0 {
		return
	}

	names = make([]string, n)
	for i := range names {
		names[i] = b.string(MaxDirNameSize, ErrDirNameTooLarge)
	}
	return
}

// Rwalk decodes a Rwalk message from Buffer.
func (b *Buffer) Rwalk() (qids []Qid) {
	n := b.uint16()
	if n > MaxWalkElements {
		b.err = ErrWalkElements
		return
	}
	if n == 0 {
		return
	}

	qids = make([]Qid, n)
	for i := range qids {
		qids[i] = Qid{
			Type:    b.uint8(),
			Version: b.uint32(),
			Path:    b.uint64(),
		}
	}
	return
}

// Topen decodes a Topen message from Buffer.
func (b *Buffer) Topen() (fid uint32, mode uint8) {
	fid = b.uint32()
	mode = b.uint8()
	return
}

// Ropen decodes a Ropen message from Buffer.
func (b *Buffer) Ropen() (typ uint8, version uint32, path uint64, iounit uint32) {
	typ = b.uint8()
	version = b.uint32()
	path = b.uint64()
	iounit = b.uint32()
	return
}

// Tcreate decodes a Tcreate message from Buffer.
func (b *Buffer) Tcreate() (fid uint32, name string, perm uint32, mode uint8) {
	fid = b.uint32()
	name = b.string(MaxDirNameSize, ErrDirNameTooLarge)
	perm = b.uint32()
	mode = b.uint8()
	return
}

// Rcreate decodes a Rcreate message from Buffer.
func (b *Buffer) Rcreate() (typ uint8, version uint32, path uint64, iounit uint32) {
	typ = b.uint8()
	version = b.uint32()
	path = b.uint64()
	iounit = b.uint32()
	return
}

// Tread decodes a Tread message from Buffer.
func (b *Buffer) Tread() (fid uint32, offset uint64, count uint32) {
	fid = b.uint32()
	offset = b.uint64()
	count = b.uint32()
	return
}

// Rread decodes a Rread message from Buffer.
func (b *Buffer) Rread() (data []byte) {
	data = b.raw32()
	return
}

// Twrite decodes a Twrite message from Buffer.
func (b *Buffer) Twrite() (fid uint32, offset uint64, data []byte) {
	fid = b.uint32()
	offset = b.uint64()
	data = b.raw32()
	return
}

// Rwrite decodes a Rwrite message from Buffer.
func (b *Buffer) Rwrite() (count uint32) {
	count = b.uint32()
	return
}

// Tclunk decodes a Tclunk message from Buffer.
func (b *Buffer) Tclunk() (fid uint32) {
	fid = b.uint32()
	return
}

// Rclunk decodes a Rclunk message from Buffer.
//func (b *Buffer) Rclunk() uint16 { return b.uint16() }

// Tremove decodes a Tremove message from Buffer.
func (b *Buffer) Tremove() (fid uint32) {
	fid = b.uint32()
	return
}

// Rremove decodes a Rremove message from Buffer.
//func (b *Buffer) Rremove() uint16 { return b.uint16() }

// Tstat decodes a Tstat message from Buffer.
func (b *Buffer) Tstat() (fid uint32) {
	fid = b.uint32()
	return
}

// Rstat decodes a Rstat message from Buffer.
func (b *Buffer) Rstat() (data []byte) {
	data = b.raw16()
	return
}

// Twstat decodes a Twstat message from Buffer.
func (b *Buffer) Twstat() (fid uint32, data []byte) {
	fid = b.uint32()
	data = b.raw16()
	return
}

// Rwstat decodes a Rwstat message from Buffer.
func (b *Buffer) Rwstat() uint16 { return b.uint16() }

// Decoder provides an interface for reading a stream of 9P2000 messages
// from an io.Reader. Successive calls to the Next method of a Decoder
// will fetch and validate 9P2000 messages from the input stream, until
// EOF is encountered, or another error is encountered.
//
// Decoder is not safe for concurrent use. Usage of any Decoder method
// should be delegated to a single thread of execution or protected by a
// mutex.
type Decoder struct {
	// MaxMessageSize is the maximum size message that a Decoder will
	// accept. If MaxMessageSize is -1, a Decoder will accept any size
	// message.
	MaxMessageSize int64

	header [7]byte
	r      io.Reader
	err    error
}

// NewDecoder returns a new decoder that reads from the io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		MaxMessageSize: DefaultMaxMessageSize,
		r:              bufio.NewReader(r),
	}
}

// Reset discards any buffered data, resets all state, and switches the
// buffered reader to read from r.
func (d *Decoder) Reset(r io.Reader) {
	d.r = bufio.NewReader(r)
	d.err = nil
}

func (d *Decoder) decodeHeader() (int, FcallType, uint16, error) {
	if err := d.readFull(d.header[:]); err != nil {
		return 0, 0, 0, err
	}

	size := int64(order.Uint32(d.header[:4]))
	if size < headerLen {
		return 0, 0, 0, ErrMessageTooSmall
	}
	if size > d.MaxMessageSize || size > maxInt {
		return 0, 0, 0, ErrMessageTooLarge
	}

	typ := FcallType(d.header[4])
	if typ < Tversion || typ > Rwstat {
		return 0, 0, 0, ErrUnknownMessageType
	}

	return int(size - 7), typ, order.Uint16(d.header[5:7]), nil
}

// Next reads the next message from the input stream and stores it
// Buffer.
func (d *Decoder) Next() (FcallType, uint16, *Buffer, error) {
	size, typ, tag, err := d.decodeHeader()
	if err != nil {
		return 0, 0, nil, err
	}

	buf := newBuffer(size, true)
	if err = d.readFull(buf.data[:size]); err != nil {
		return 0, 0, nil, err
	}

	return typ, tag, buf, nil
}

func (d *Decoder) readFull(buf []byte) error {
	if d.err != nil {
		return d.err
	}
	_, d.err = io.ReadFull(d.r, buf)
	return d.err
}
