package ninep

import "encoding/binary"

var (
	puint64 = binary.LittleEndian.PutUint64
	puint32 = binary.LittleEndian.PutUint32
	puint16 = binary.LittleEndian.PutUint16
)

func puint8(b []byte, v uint8) {
	_ = b[0] // bounds check hint to compiler
	b[0] = v
}

func pstring(b []byte, s string) int {
	_ = b[2+len(s)-1] // bounds check hint to compiler
	puint16(b[0:2], uint16(len(s)))
	if len(s) == 0 {
		return 2
	}
	return 2 + copy(b[2:], s)
}

func guint64(b []byte, v *uint64) []byte {
	*v = binary.LittleEndian.Uint64(b)
	return b[8:]
}

func guint32(b []byte, v *uint32) []byte {
	*v = binary.LittleEndian.Uint32(b)
	return b[4:]
}

func guint16(b []byte, v *uint16) []byte {
	*v = binary.LittleEndian.Uint16(b)
	return b[2:]
}

func guint8(b []byte, v *uint8) []byte {
	*v = b[0]
	return b[1:]
}

func gstring(b []byte, s *string) []byte {
	size := int(binary.LittleEndian.Uint16(b))
	if size == 0 {
		*s = ""
		return b[2:]
	}
	*s = string(b[2 : 2+size])
	return b[2+size:]
}
