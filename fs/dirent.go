// +build darwin linux

package fs

import (
	"encoding/binary"
	"unsafe"

	"github.com/azmodb/ninep/proto"
	"golang.org/x/sys/unix"
)

const direntBlockSize = 4 * 1024

// readDirent reads directory entries from the given file descriptor. The
// order of the directory entries is not specified.
func readDirent(fd int) (entries []proto.Dirent, err error) {
	buf := make([]byte, direntBlockSize)
	pos, size, offset := 0, 0, uint64(1)

	for {
		if pos >= size {
			pos = 0
			if size, err = unix.ReadDirent(fd, buf); err != nil {
				return nil, err
			}
			if size <= 0 {
				break // EOF
			}
		}

		ents, consumed, off := parse(buf[pos:size], offset)
		offset = off
		pos += consumed
		entries = append(entries, ents...)
	}

	return entries, err
}

// parse parses directory entries in buf. It returns the number of bytes
// consumed from buf and the current offset.
func parse(buf []byte, offset uint64) ([]proto.Dirent, int, uint64) {
	var entries []proto.Dirent
	n := len(buf)

	for len(buf) > 0 {
		dirent := (*unix.Dirent)(unsafe.Pointer(&buf[0]))
		reclen, ok := direntReclen(buf, dirent)
		if !ok {
			return entries, n, offset
		}
		rec := buf[:reclen]
		buf = buf[reclen:]

		ino, ok := direntIno(rec, dirent)
		if !ok {
			return entries, n, offset
		}
		if ino == 0 { // file absent in directory
			continue
		}

		typ, ok := direntType(rec, dirent)
		if !ok {
			return entries, n, offset
		}
		if typ == unix.DT_UNKNOWN { // TODO(mason)
			continue
		}

		name, ok := direntName(rec, dirent)
		if !ok {
			return entries, n, offset
		}

		entries = append(entries, proto.Dirent{
			Qid: struct {
				Type    uint8
				Version uint32
				Path    uint64
			}{proto.UnixDirTypeToQidType(typ), 0, ino},
			Offset: offset,
			Type:   typ,
			Name:   name,
		})
		offset++
	}

	return entries, n - len(buf), offset
}

func direntReclen(buf []byte, dirent *unix.Dirent) (uint64, bool) {
	offset := unsafe.Offsetof(dirent.Reclen)
	size := unsafe.Sizeof(dirent.Reclen)
	return decode(buf, offset, size)
}

func direntType(buf []byte, dirent *unix.Dirent) (uint8, bool) {
	offset := unsafe.Offsetof(dirent.Type)
	size := unsafe.Sizeof(dirent.Type)
	v, ok := decode(buf, offset, size)
	if !ok {
		return 0, false
	}
	return uint8(v), true
}

func direntIno(buf []byte, dirent *unix.Dirent) (uint64, bool) {
	offset := unsafe.Offsetof(dirent.Ino)
	size := unsafe.Sizeof(dirent.Ino)
	return decode(buf, offset, size)
}

func direntName(buf []byte, dirent *unix.Dirent) (string, bool) {
	const namelen = unsafe.Sizeof(dirent.Name)
	name := (*[namelen]byte)(unsafe.Pointer(&dirent.Name[0]))
	size := uintptr(0)
	for ; size < namelen; size++ {
		if name[size] == 0 {
			break
		}
	}
	return string(name[:size]), true
}

var order = binary.LittleEndian

func decode(buf []byte, offset, size uintptr) (uint64, bool) {
	if len(buf) < int(offset+size) {
		return 0, false
	}

	switch size {
	case 1:
		return uint64(buf[offset]), true
	case 2:
		return uint64(order.Uint16(buf[offset:])), true
	case 8:
		return order.Uint64(buf[offset:]), true
	}
	panic("dirent: decode with unsupported size")
}
