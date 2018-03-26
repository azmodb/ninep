package ninep

import "encoding/binary"

func gheader(b []byte) (size uint32, typ uint8, err error) {
	if len(b) < headerLen {
		return size, typ, errMessageTooSmall
	}

	size = binary.LittleEndian.Uint32(b[0:4])
	if size < headerLen {
		return size, typ, errMessageTooSmall
	}

	typ = b[4]
	if typ == Terror || typ < Tversion || typ > Rwstat {
		return size, typ, errInvalidMessageType
	}
	if size < minSizeLUT[typ-100] {
		return size, typ, errMessageTooSmall
	}
	return size, typ, nil
}

func gstring(b []byte) (string, []byte) {
	n := binary.LittleEndian.Uint16(b[0:2])
	if n == 0 {
		return "", b[2:]
	}
	return string(b[2 : 2+n]), b[2+n:]
}

func guint64(b []byte) (uint64, []byte) {
	v := binary.LittleEndian.Uint64(b)
	return v, b[8:]
}

func guint32(b []byte) (uint32, []byte) {
	v := binary.LittleEndian.Uint32(b)
	return v, b[4:]
}

func guint16(b []byte) (uint16, []byte) {
	v := binary.LittleEndian.Uint16(b)
	return v, b[2:]
}

var (
	puint64 = binary.LittleEndian.PutUint64
	puint32 = binary.LittleEndian.PutUint32
	puint16 = binary.LittleEndian.PutUint16
)

type buffer []byte

func newBuffer(size int) buffer {
	if size <= 0 {
		size = 256
	}
	b := make(buffer, 0, size)
	return b
}

func (b *buffer) WriteString(s string) (int, error) {
	n := b.grow(len(s))
	return copy((*b)[n:], s), nil
}

func (b *buffer) Write(p []byte) (int, error) {
	n := b.grow(len(p))
	return copy((*b)[n:], p), nil
}

func (b *buffer) WriteByte(p byte) error {
	n := b.grow(1)
	(*b)[n] = p
	return nil
}

func (b *buffer) grow(size int) int {
	n := len(*b)
	switch {
	case n == 0 && size < 256:
		*b = make([]byte, 256)
	case n+size > cap(*b):
		m := n + size*2*cap(*b)
		if m > 2*8192 {
			m = n + size
		}
		nb := make([]byte, n, m)
		copy(nb, *b)
		*b = nb
	}
	*b = (*b)[0 : n+size]
	return n
}
