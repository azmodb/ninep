package proto

import (
	"bytes"
	"encoding/binary"
	"unicode/utf8"
)

const errOverflow = Error("size of field exceeds size of message")

var separator = []byte("/")

type verifier func([]byte) error

func verify(buf []byte, offset, n int, fn verifier) error {
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

func verifyVersion(version []byte) error {
	if len(version) > maxVersionLen {
		return errVersionTooLarge
	}
	if !utf8.Valid(version) {
		return errVersionInvalidUTF8
	}
	return nil
}
func verifyUname(uname []byte) error {
	if len(uname) > maxUnameLen {
		return errUnameTooLarge
	}
	if !utf8.Valid(uname) {
		return errUnameInvalidUTF8
	}
	return nil
}

func verifyName(name []byte) error {
	if len(name) > maxNameLen {
		return errElemTooLarge
	}
	if !utf8.Valid(name) {
		return errElemInvalidUTF8
	}
	if bytes.Contains(name, separator) {
		return errPathName
	}
	return nil
}

func verifyPath(name []byte) (err error) {
	for len(name) > 0 && name[0] == '/' {
		name = name[1:]
	}
	if len(name) == 0 {
		return nil
	}
	if len(name) > maxPathLen {
		return errPathTooLarge
	}
	if bytes.Count(name, separator) > maxWalkElem {
		return errMaxWalkElem
	}

	elems := bytes.Split(name, separator)
	for _, elem := range elems {
		if err = verifyName(elem); err != nil {
			return err
		}
	}
	return nil
}

func verifyData(buf []byte, offset int, max int64) error {
	if offset+4 > len(buf) {
		return errOverflow
	}
	size := int64(guint32(buf[offset : offset+4]))
	if max > 0 && size > max {
		return errDataTooLarge
	}
	if size > MaxDataLen { // MaxDataLen == ~ 2GB
		return errDataTooLarge
	}

	if offset+4+int(size) > len(buf) {
		return errOverflow
	}
	return nil
}

func gheader(b []byte, msize int64) (uint32, uint8, error) {
	if len(b) < headerLen {
		return 0, 0, errMessageTooSmall
	}

	size := binary.LittleEndian.Uint32(b[0:4])
	if size < headerLen {
		return 0, 0, errMessageTooSmall
	}
	if int64(size) > msize {
		return 0, 0, errMessageTooLarge
	}

	typ := b[4]
	if typ == msgTerror || typ < msgTversion || typ > msgRwstat {
		return 0, 0, errInvalidMessageType
	}
	if size < minSize(typ) {
		return 0, 0, errMessageTooSmall
	}
	return size, typ, nil
}

func gstat(b []byte) (s Stat) {
	s.Type = guint16(b[2:4])
	s.Dev = guint32(b[4:8])
	s.Qid.UnmarshalBinary(b[8:21])
	s.Mode = guint32(b[21:25])
	s.Atime = guint32(b[25:29])
	s.Mtime = guint32(b[29:33])
	s.Length = guint64(b[33:41])

	n := 41 + gstring(b[41:], &s.Name)
	n += gstring(b[n:], &s.UID)
	n += gstring(b[n:], &s.GID)
	n += gstring(b[n:], &s.MUID)
	return s
}

func gqid(b []byte) (q Qid) {
	q.Type = b[0]
	q.Version = guint32(b[1:5])
	q.Path = guint64(b[5:13])
	return q
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
	n := int(guint16(data[0:2]))
	if n == 0 {
		*s = ""
		return 2
	}
	*s = string(data[2 : 2+n])
	return 2 + n
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
