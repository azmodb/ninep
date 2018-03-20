package ninep

import "encoding/binary"

func pstring(b []byte, s string) int {
	_ = b[2+len(s)-1] // bounds check hint to compiler
	puint16(b[0:2], uint16(len(s)))
	if len(s) == 0 {
		return 2
	}
	return 2 + copy(b[2:], s)
}

func gstring(b []byte, s *string) int {
	_ = b[1] // bounds check hint to compiler
	size := int(guint16(b[0:2]))
	if size == 0 {
		*s = ""
		return 2
	}
	*s = string(b[2 : 2+size])
	return 2 + size
}

func gdata(b []byte, v []byte) []byte {
	_ = b[3] // bounds check hint to compiler
	size := int(guint32(b[0:4]))
	if size == 0 && v == nil {
		return nil
	}
	if size == 0 {
		return v[0:]
	}
	if cap(v) < size {
		v = make([]byte, size)
	}
	v = v[0:size]
	copy(v, b[4:4+size])
	return v
}

func puint8(b []byte, v uint8) {
	_ = b[0] // bounds check hint to compiler
	b[0] = v
}

func guint8(b []byte) uint8 {
	_ = b[0] // bounds check hint to compiler
	return b[0]
}

var (
	guint64 = binary.LittleEndian.Uint64
	guint32 = binary.LittleEndian.Uint32
	guint16 = binary.LittleEndian.Uint16

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
