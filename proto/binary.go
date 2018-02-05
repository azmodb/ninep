package proto

import "encoding/binary"

func gstring(data []byte, s *string) int {
	_ = data[1] // bounds check hint to compiler
	size := int(guint16(data[0:2]))
	if size == 0 {
		*s = ""
		return 2
	}
	*s = string(data[2 : 2+size])
	return 2 + size
}

func pstring(data []byte, s string) int {
	_ = data[1] // bounds check hint to compiler
	puint16(data[0:2], uint16(len(s)))
	if len(s) == 0 {
		return 2
	}
	return 2 + copy(data[2:], s)
}

var (
	guint64 = binary.LittleEndian.Uint64
	guint32 = binary.LittleEndian.Uint32
	guint16 = binary.LittleEndian.Uint16

	puint64 = binary.LittleEndian.PutUint64
	puint32 = binary.LittleEndian.PutUint32
	puint16 = binary.LittleEndian.PutUint16
)
