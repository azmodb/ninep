package proto

import "encoding/binary"

const errOverflow = Error("size of field exceeds size of message")

type verifierFunc func([]byte) error

func vfield(buf []byte, offset, n int, fn verifierFunc) error {
	if offset+2 > len(buf) {
		return errOverflow
	}
	size := int(guint16(buf[offset : offset+2]))

	for i := 0; i < n; i++ {
		offset += 2 + size
		if offset+2 > len(buf) {
			return errOverflow
		}
		size = int(guint16(buf[offset : offset+2]))
	}

	if offset+2+size > len(buf) {
		return errOverflow
	}
	if fn == nil {
		return nil
	}
	return fn(buf[offset+2 : offset+2+size])
}

func vdata(buf []byte, offset int, max int64) error {
	if offset+4 > len(buf) {
		return errOverflow
	}
	size := int64(guint32(buf[offset : offset+4]))
	if max > 0 && size > max {
		return errDataTooLarge
	}
	if size > maxDataLen { // maxDataLen == ~ 2GB
		return errDataTooLarge
	}

	if offset+4+int(size) > len(buf) {
		return errOverflow
	}
	return nil
}

func gfield(m []byte, offset, n int) []byte {
	size := int(guint16(m[offset : offset+2]))
	for i := 0; i < n; i++ {
		offset += size + 2
		size = int(guint16(m[offset : offset+2]))
	}
	return m[offset+2 : offset+2+size]
}

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
