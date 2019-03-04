//go:generate stringer -type FcallType -output strings.go

package proto

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Based on http://9p.io/sources/plan9/sys/include/fcall.h
const (
	Tversion FcallType = iota + 100 // size[4] Tversion tag[2] msize[4] version[s]
	Rversion                        // size[4] Rversion tag[2] msize[4] version[s]
	Tauth                           // size[4] Tauth tag[2] afid[4] uname[s] aname[s]
	Rauth                           // size[4] Rauth tag[2] aqid[13]
	Tattach                         // size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	Rattach                         // size[4] Rattach tag[2] qid[13]
	Terror                          // illegal
	Rerror                          // size[4] Rerror tag[2] ename[s]
	Tflush                          // size[4] Tflush tag[2] oldtag[2]
	Rflush                          // size[4] Rflush tag[2]
	Twalk                           // size[4] Twalk tag[2] fid[4] newfid[4] [2]wname[s]
	Rwalk                           // size[4] Rwalk tag[2] [2]wqid[13]
	Topen                           // size[4] Topen tag[2] fid[4] mode[1]
	Ropen                           // size[4] Ropen tag[2] qid[13] iounit[4]
	Tcreate                         // size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	Rcreate                         // size[4] Rcreate tag[2] qid[13] iounit[4]
	Tread                           // size[4] Tread tag[2] fid[4] offset[8] count[4]
	Rread                           // size[4] Rread tag[2] [4]data[count]
	Twrite                          // size[4] Twrite tag[2] fid[4] offset[8] [4]data[count]
	Rwrite                          // size[4] Rwrite tag[2] count[4]
	Tclunk                          // size[4] Tclunk tag[2] fid[4]
	Rclunk                          // size[4] Rclunk tag[2]
	Tremove                         // size[4] Tremove tag[2] fid[4]
	Rremove                         // size[4] Rremove tag[2]
	Tstat                           // size[4] Tstat tag[2] fid[4]
	Rstat                           // size[4] Rstat tag[2] stat[n]
	Twstat                          // size[4] Twstat tag[2] fid[4] stat[n]
	Rwstat                          // size[4] Rwstat tag[2]
)

// FcallType represents a 9P2000 message (fcall) type.
type FcallType uint8

var minSizeLUT = [28]uint32{
	13,                 // size[4] Tversion tag[2] msize[4] version[s]
	13,                 // size[4] Rversion tag[2] mversion[s]
	15,                 // size[4] Tauth tag[2] afid[4] uname[s] aname[s]
	20,                 // size[4] Rauth tag[2] aqid[13]
	19,                 // size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	20,                 // size[4] Rattach tag[2] qid[13]
	0,                  // illegal
	9,                  // size[4] Rerror tag[2] ename[s]
	9,                  // size[4] Tflush tag[2] oldtag[2]
	7,                  // size[4] Rflush tag[2]
	17,                 // size[4] Twalk tag[2] fid[4] newfid[4] [2]wname[s]
	9,                  // size[4] Rwalk tag[2] [2]wqid[13]
	12,                 // size[4] Topen tag[2] fid[4] mode[1]
	24,                 // size[4] Ropen tag[2] qid[13] iounit[4]
	18,                 // size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	24,                 // size[4] Rcreate tag[2] qid[13] iounit[4]
	FixedReadWriteSize, // size[4] Tread tag[2] fid[4] offset[8] count[4]
	11,                 // size[4] Rread tag[2] [4]data[count]
	FixedReadWriteSize, // size[4] Twrite tag[2] fid[4] offset[8] [4]data[count]
	11,                 // size[4] Rwrite tag[2] count[4]
	11,                 // size[4] Tclunk tag[2] fid[4]
	7,                  // size[4] Rclunk tag[2]
	11,                 // size[4] Tremove tag[2] fid[4]
	7,                  // size[4] Rremove tag[2]
	11,                 // size[4] Tstat tag[2] fid[4]
	9,                  // size[4] Rstat tag[2] stat[n]
	13,                 // size[4] Twstat tag[2] fid[4] stat[n]
	7,                  // size[4] Rwstat tag[2]
}

func fixedSize(typ FcallType) uint32 { return minSizeLUT[typ-100] }

const (
	// FixedReadWriteSize is the length of all fixed-width fields in a Twrite
	// or Tread message. Twrite and Tread messages are defined as
	//
	//     size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	//     size[4] Tread  tag[2] fid[4] offset[8] count[4]
	//
	FixedReadWriteSize = 4 + 1 + 2 + 4 + 8 + 4 // 23

	// To make the contents of a directory, such as returned by read(5),
	// easy to parse, each directory entry begins with a size field.
	// For consistency, the entries in Twstat and Rstat messages also
	// contain their size, which means the size appears twice.
	//
	//     http://9p.io/magic/man2html/5/stat
	//
	fixedStatSize = 2 + 4 + 13 + 4 + 4 + 4 + 8 + 4*2 // 47

	maxStatSize = 2 + fixedStatSize + MaxDirNameSize + 3*MaxUserNameSize

	// MaxDataSize is the maximum data size of a Twrite or Rread message.
	MaxDataSize = (1<<31 - 1) - (FixedReadWriteSize + 1) // ~ 2GB

	// DefaultMaxDataSize is the default maximum data size of a Twrite or
	// Rread message.
	DefaultMaxDataSize = 2 * 1024 * 1024 // ~ 2MB

	// MaxMessageSize is the maximum size of a 9P2000 message.
	MaxMessageSize = (FixedReadWriteSize + 1) + MaxDataSize

	// DefaultMaxMessageSize is the default maximum size of a 9P2000
	// message.
	DefaultMaxMessageSize = (FixedReadWriteSize + 1) + DefaultMaxDataSize

	// MinMessageSize is the minimum size of a 9P2000 message.
	MinMessageSize = 17 + MaxPathSize
)

// Defines proto package limits.
const (
	MaxPathSize     = MaxWalkElements * MaxDirNameSize
	MaxUserNameSize = 255
	MaxDirNameSize  = 255
	MaxVersionSize  = 8
	MaxEnameSize    = 1<<16 - 1

	MaxWalkElements = 16
)

// Defines 9P2000 protocol related errors.
const (
	ErrDirNameTooLarge  = ProtocolError("file or directory name too large")
	ErrUserNameTooLarge = ProtocolError("user name too large")
	ErrPathTooLarge     = ProtocolError("file tree name too large")
	ErrEnameTooLarge    = ProtocolError("error message too large")
	ErrVersionTooLarge  = ProtocolError("version identifier too large")
	ErrDataTooLarge     = ProtocolError("payload too large")
	ErrStatTooLarge     = ProtocolError("stat information too large")

	ErrUnknownMessageType = ProtocolError("unknown message type")
	ErrMessageTooLarge    = ProtocolError("message too large")
	ErrMessageTooSmall    = ProtocolError("message too small")

	ErrWalkElements = ProtocolError("maximum walk elements exceeded")
)

// ProtocolError records a 9P2000 protocol error.
type ProtocolError string

func (e ProtocolError) Error() string { return string(e) }

const headerLen = 7 // size[4] FcallType tag[2]

// Version defines the supported protocol version.
const Version = "9P2000"

var order = binary.LittleEndian

func putString(buf []byte, v string) int {
	order.PutUint16(buf[:2], uint16(len(v)))
	if len(v) == 0 {
		return 2
	}
	return 2 + copy(buf[2:], v)
}

func getString(buf []byte, v *string) int {
	size := int(order.Uint16(buf[:2]))
	*v = string(buf[2 : 2+size])
	return 2 + size
}

// Qid type field represents the type of a file (directory, etc.),
// represented as a bit vector corresponding to the high 8 bits of the
// file mode word.
const (
	QTDIR    = 0x80 // directories
	QTAPPEND = 0x40 // append only files
	QTEXCL   = 0x20 // exclusive use files
	QTMOUNT  = 0x10 // mounted channel
	QTAUTH   = 0x08 // authentication file (afid)
	QTTMP    = 0x04 // non-backed-up file
	QTFILE   = 0x00
)

// Qid represents the server's unique identification for the file being
// accessed. Two files on the same server hierarchy are the same if and
// only if their qids are the same.
type Qid struct {
	Type uint8 // type of a file (directory, etc)

	// Version is a version:number for a file. It is incremented every
	// time a file is modified. By convention, synthetic files usually
	// have a verison number of 0. Traditional files have a version
	// numb:r that is a hash of their modification time.
	Version uint32

	// Path is an integer unique among all files in the hierarchy. If
	// a file is deleted and recreated with the same name in the same
	// directory, the old and new path components of the qids should
	// be different.
	Path uint64
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (q Qid) MarshalBinary() ([]byte, error) {
	buf := [13]byte{}
	buf[0] = q.Type
	order.PutUint32(buf[1:5], q.Version)
	order.PutUint64(buf[5:13], q.Path)
	return buf[:], nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (q *Qid) UnmarshalBinary(data []byte) error {
	if len(data) < 13 {
		return io.ErrUnexpectedEOF
	}

	q.Type = data[0]
	q.Version = order.Uint32(data[1:5])
	q.Path = order.Uint64(data[5:13])
	return nil
}

func (q Qid) String() string {
	t := ""
	if q.Type&QTDIR != 0 {
		t += "d"
	}
	if q.Type&QTAPPEND != 0 {
		t += "a"
	}
	if q.Type&QTEXCL != 0 {
		t += "l"
	}
	if q.Type&QTMOUNT != 0 {
		t += "m"
	}
	if q.Type&QTAUTH != 0 {
		t += "A"
	}
	if q.Type&QTTMP != 0 {
		t += "t"
	}
	return fmt.Sprintf("(%.16x %d %q)", q.Path, q.Version, t)
}

// Stat describes a directory entry. It is contained in Rstat and Twstat
// messages. Tread requests on directories return a Stat structure for
// each directory entry.
type Stat struct {
	Type   uint16
	Dev    uint32
	Qid    Qid
	Mode   uint32
	Atime  uint32
	Mtime  uint32
	Length uint64
	Name   string
	Uid    string
	Gid    string
	Muid   string
}

func (s Stat) size() int {
	strSize := len(s.Name) + len(s.Uid) + len(s.Gid) + len(s.Muid)
	return 2 + fixedStatSize + strSize
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s Stat) MarshalBinary() ([]byte, error) {
	if len(s.Name) > MaxDirNameSize {
		return nil, ErrDirNameTooLarge
	}
	if len(s.Gid) > MaxUserNameSize {
		return nil, ErrUserNameTooLarge
	}
	if len(s.Uid) > MaxUserNameSize {
		return nil, ErrUserNameTooLarge
	}
	if len(s.Muid) > MaxUserNameSize {
		return nil, ErrUserNameTooLarge
	}

	data := make([]byte, s.size())
	if _, err := s.marshal(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s Stat) marshal(data []byte) (int, error) {
	if len(data) < 2+fixedStatSize {
		return 0, io.ErrUnexpectedEOF
	}

	order.PutUint16(data[0:2], uint16(len(data)-2))
	order.PutUint16(data[2:4], s.Type)
	order.PutUint32(data[4:8], s.Dev)
	data[8] = s.Qid.Type
	order.PutUint32(data[9:13], s.Qid.Version)
	order.PutUint64(data[13:21], s.Qid.Path)
	order.PutUint32(data[21:25], s.Mode)
	order.PutUint32(data[25:29], s.Atime)
	order.PutUint32(data[29:33], s.Mtime)
	order.PutUint64(data[33:41], s.Length)
	n := 41
	n += putString(data[n:], s.Name)
	n += putString(data[n:], s.Uid)
	n += putString(data[n:], s.Gid)
	n += putString(data[n:], s.Muid)

	return n, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *Stat) UnmarshalBinary(data []byte) error {
	if len(data) < 2+fixedStatSize {
		return io.ErrUnexpectedEOF
	}

	s.Type = order.Uint16(data[2:4])
	s.Dev = order.Uint32(data[4:8])
	s.Qid.Type = data[8]
	s.Qid.Version = order.Uint32(data[9:13])
	s.Qid.Path = order.Uint64(data[13:21])
	s.Mode = order.Uint32(data[21:25])
	s.Atime = order.Uint32(data[25:29])
	s.Mtime = order.Uint32(data[29:33])
	s.Length = order.Uint64(data[33:41])

	n := 41 + getString(data[41:], &s.Name)
	n += getString(data[n:], &s.Uid)
	n += getString(data[n:], &s.Gid)
	n += getString(data[n:], &s.Muid)
	return nil
}

func (s Stat) String() string {
	return fmt.Sprintf("%q %q %q %s q %v m %#o at %d mt %d l %d t %d d %d",
		s.Name, s.Uid, s.Gid, s.Muid, s.Qid, s.Mode,
		s.Atime, s.Mtime, s.Length, s.Type, s.Dev)
}
