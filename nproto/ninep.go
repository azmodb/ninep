package ninep

import (
	"fmt"
	"io"
)

// Based on http://9p.io/sources/plan9/sys/include/fcall.h
const (
	Tversion = iota + 100 // size[4] Tversion tag[2] msize[4] version[s]
	Rversion              // size[4] Rversion tag[2] msize[4] version[s]
	Tauth                 // size[4] Tauth tag[2] afid[4] uname[s] aname[s]
	Rauth                 // size[4] Rauth tag[2] aqid[13]
	Tattach               // size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
	Rattach               // size[4] Rattach tag[2] qid[13]
	Terror                // illegal
	Rerror                // size[4] Rerror tag[2] ename[s]
	Tflush                // size[4] Tflush tag[2] oldtag[2]
	Rflush                // size[4] Rflush tag[2]
	Twalk                 // size[4] Twalk tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])
	Rwalk                 // size[4] Rwalk tag[2] nwqid[2] nwqid*(wqid[13])
	Topen                 // size[4] Topen tag[2] fid[4] mode[1]
	Ropen                 // size[4] Ropen tag[2] qid[13] iounit[4]
	Tcreate               // size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
	Rcreate               // size[4] Rcreate tag[2] qid[13] iounit[4]
	Tread                 // size[4] Tread tag[2] fid[4] offset[8] count[4]
	Rread                 // size[4] Rread tag[2] count[4] data[count]
	Twrite                // size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	Rwrite                // size[4] Rwrite tag[2] count[4]
	Tclunk                // size[4] Tclunk tag[2] fid[4]
	Rclunk                // size[4] Rclunk tag[2]
	Tremove               // size[4] Tremove tag[2] fid[4]
	Rremove               // size[4] Rremove tag[2]
	Tstat                 // size[4] Tstat tag[2] fid[4]
	Rstat                 // size[4] Rstat tag[2] stat[n]
	Twstat                // size[4] Twstat tag[2] fid[4] stat[n]
	Rwstat                // size[4] Rwstat tag[2]
)

var (
	minSizeLUT = [28]uint32{
		13,                // size[4] Tversion tag[2] msize[4] version[s]
		13,                // size[4] Rversion tag[2] mversion[s]
		15,                // size[4] Tauth tag[2] afid[4] uname[s] aname[s]
		20,                // size[4] Rauth tag[2] aqid[13]
		19,                // size[4] Tattach tag[2] fid[4] afid[4] uname[s] aname[s]
		20,                // size[4] Rattach tag[2] qid[13]
		0,                 // illegal
		9,                 // size[4] Rerror tag[2] ename[s]
		9,                 // size[4] Tflush tag[2] oldtag[2]
		7,                 // size[4] Rflush tag[2]
		17,                // size[4] Twalk tag[2] fid[4] newfid[4] nwname[2] nwname*(wname[s])
		9,                 // size[4] Rwalk tag[2] nwqid[2] nwqid*(wqid[13])
		12,                // size[4] Topen tag[2] fid[4] mode[1]
		24,                // size[4] Ropen tag[2] qid[13] iounit[4]
		18,                // size[4] Tcreate tag[2] fid[4] name[s] perm[4] mode[1]
		24,                // size[4] Rcreate tag[2] qid[13] iounit[4]
		fixedReadWriteLen, // size[4] Tread tag[2] fid[4] offset[8] count[4]
		11,                // size[4] Rread tag[2] count[4] data[count]
		fixedReadWriteLen, // size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
		11,                // size[4] Rwrite tag[2] count[4]
		11,                // size[4] Tclunk tag[2] fid[4]
		7,                 // size[4] Rclunk tag[2]
		11,                // size[4] Tremove tag[2] fid[4]
		7,                 // size[4] Rremove tag[2]
		11,                // size[4] Tstat tag[2] fid[4]
		11 + fixedStatLen, // size[4] Rstat tag[2] stat[n]
		15 + fixedStatLen, // size[4] Twstat tag[2] fid[4] stat[n]
		7,                 // size[4] Rwstat tag[2]
	}

	sizeFunc = map[uint8]func(m *Message) uint32{
		Tversion: func(m *Message) uint32 {
			return minSizeLUT[Tversion-100] + uint32(len(m.Version))
		},
		Rversion: func(m *Message) uint32 {
			return minSizeLUT[Tversion-100] + uint32(len(m.Version))
		},
		Tauth: func(m *Message) uint32 {
			return minSizeLUT[Tauth-100] + uint32(len(m.Uname)+len(m.Aname))
		},
		Rauth: func(m *Message) uint32 {
			return minSizeLUT[Rauth-100]
		},
		Tattach: func(m *Message) uint32 {
			return minSizeLUT[Tattach-100] + uint32(len(m.Uname)+len(m.Aname))
		},
		Rattach: func(m *Message) uint32 {
			return minSizeLUT[Rattach-100]
		},
		Terror: func(m *Message) uint32 { return 0 }, // illegal
		Rerror: func(m *Message) uint32 {
			return minSizeLUT[Rerror-100] + uint32(len(m.Ename))
		},
		Tflush: func(m *Message) uint32 { return minSizeLUT[Tflush-100] },
		Rflush: func(m *Message) uint32 { return minSizeLUT[Rflush-100] },
		Twalk: func(m *Message) uint32 {
			size := minSizeLUT[Twalk-100]
			for _, n := range m.Wname {
				size += 2 + uint32(len(n))
			}
			return size
		},
		Rwalk: func(m *Message) uint32 {
			return minSizeLUT[Rwalk-100] + uint32(13*len(m.Wqid))
		},
		Topen: func(m *Message) uint32 { return minSizeLUT[Topen-100] },
		Ropen: func(m *Message) uint32 { return minSizeLUT[Ropen-100] },
		Tcreate: func(m *Message) uint32 {
			return minSizeLUT[Tcreate-100] + uint32(len(m.Name))
		},
		Rcreate: func(m *Message) uint32 { return minSizeLUT[Rcreate-100] },
		Tread:   func(m *Message) uint32 { return minSizeLUT[Tread-100] },
		Rread:   func(m *Message) uint32 { return minSizeLUT[Rread-100] },
		Twrite: func(m *Message) uint32 {
			return minSizeLUT[Twrite-100] + uint32(len(m.Data))
		},
		Rwrite:  func(m *Message) uint32 { return minSizeLUT[Rwrite-100] },
		Tclunk:  func(m *Message) uint32 { return minSizeLUT[Tclunk-100] },
		Rclunk:  func(m *Message) uint32 { return minSizeLUT[Rclunk-100] },
		Tremove: func(m *Message) uint32 { return minSizeLUT[Tremove-100] },
		Rremove: func(m *Message) uint32 { return minSizeLUT[Rremove-100] },
		Tstat:   func(m *Message) uint32 { return minSizeLUT[Tstat-100] },
		Rstat:   func(m *Message) uint32 { return 0 }, // TODO
		Twstat:  func(m *Message) uint32 { return 0 }, // TODO
		Rwstat:  func(m *Message) uint32 { return minSizeLUT[Rwstat-100] },
	}
)

// Defines some limits on how big some arbitrarily-long 9P2000 fields can
// be.
const (
	// size[4] Twrite tag[2] fid[4] offset[8] count[4] data[count]
	// size[4] Tread  tag[2] fid[4] offset[8] count[4]
	fixedReadWriteLen = 4 + 1 + 2 + 4 + 8 + 4 // 23

	// To make the contents of a directory, such as returned by read(5),
	// easy to parse, each directory entry begins with a size field.
	// For consistency, the entries in Twstat and Rstat messages also
	// contain their size, which means the size appears twice.
	//
	//  http://9p.io/magic/man2html/5/stat
	//
	fixedStatLen = 2 + 4 + 13 + 4 + 4 + 4 + 8 + 4*2 // 47

)

// NOFID is a reserved fid used in a Tattach request for the afid field,
// that indicates that the client does not wish to authenticate his
// session.
const NOFID = ^uint32(0)

// NOTAG is the tag for Tversion and Rversion requests.
const NOTAG = ^uint16(0)

// Flags for the mode field in Topen and Tcreate messages.
const (
	OREAD   = 0  // open read-only
	OWRITE  = 1  // open write-only
	ORDWR   = 2  // open read-write
	OEXEC   = 3  // read but check execute permission
	OTRUNC  = 16 // or'ed in (except for exec), truncate file first
	OCEXEC  = 32 // or'ed in, close on exec
	ORCLOSE = 64 // or'ed in, remove on close
)

// File and directory modes.
const (
	DMDIR    = 0x80000000 // mode bit for directories
	DMAPPEND = 0x40000000 // mode bit for append only files
	DMEXCL   = 0x20000000 // mode bit for exclusive use files
	DMMOUNT  = 0x10000000 // mode bit for mounted channel
	DMAUTH   = 0x08000000 // mode bit for authentication file
	DMTMP    = 0x04000000 // mode bit for non-backed-up file
	DMREAD   = 0x4        // mode bit for read permission
	DMWRITE  = 0x2        // mode bit for write permission
	DMEXEC   = 0x1        // mode bit for execute permission
)

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

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s Stat) MarshalBinary() ([]byte, error) {
	data := make([]byte, s.len())
	s.marshal(data)
	return data, nil
}

func (s Stat) len() int {
	strLen := len(s.Name) + len(s.Uid) + len(s.Gid) + len(s.Muid)
	return 2 + fixedStatLen + strLen
}

func (s Stat) marshal(b []byte) {
	puint16(b[0:2], uint16(len(b)))
	puint16(b[2:4], s.Type)
	puint32(b[4:8], s.Dev)
	s.Qid.marshal(b[8:21])
	puint32(b[21:25], s.Mode)
	puint32(b[25:29], s.Atime)
	puint32(b[29:33], s.Mtime)
	puint64(b[33:41], s.Length)

	n := 41 + pstring(b[41:], s.Name)
	n += pstring(b[n:], s.Uid)
	n += pstring(b[n:], s.Gid)
	_ = pstring(b[n:], s.Muid)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *Stat) UnmarshalBinary(data []byte) error {
	if len(data) < 2+fixedStatLen {
		return io.ErrUnexpectedEOF
	}

	s.unmarshal(data)
	return nil
}

func (s *Stat) unmarshal(b []byte) []byte {
	b = guint16(b[2:], &s.Type)
	b = guint32(b, &s.Dev)
	b = s.Qid.unmarshal(b)
	b = guint32(b, &s.Mode)
	b = guint32(b, &s.Atime)
	b = guint32(b, &s.Mtime)
	b = guint64(b, &s.Length)

	b = gstring(b, &s.Name)
	b = gstring(b, &s.Uid)
	b = gstring(b, &s.Gid)
	b = gstring(b, &s.Muid)
	return b
}

func (s Stat) String() string {
	return fmt.Sprintf("type:0x%x dev:0x%x qid:%q mode:%d atime:%d mtime:%d length:%d name:%q uid:%q gid:%q muid:%q",
		s.Type, s.Dev, s.Qid, s.Mode, s.Atime, s.Mtime, s.Length, s.Name, s.Uid, s.Gid, s.Muid)
}

// Qid represents the server's unique identification for the file being
// accessed. Two files on the same server hierarchy are the same if and
// only if their qids are the same.
type Qid struct {
	Type uint8 // type of a file (directory, etc)

	// Version is a version number for a file. It is incremented every
	// time a file is modified. By convention, synthetic files usually
	// have a verison number of 0. Traditional files have a version
	// number that is a hash of their modification time.
	Version uint32

	// Path is an integer unique among all files in the hierarchy. If
	// a file is deleted and recreated with the same name in the same
	// directory, the old and new path components of the qids should
	// be different.
	Path uint64
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (q Qid) MarshalBinary() ([]byte, error) {
	data := make([]byte, 13)
	q.marshal(data)
	return data, nil
}

func (q Qid) marshal(b []byte) {
	b[0] = q.Type
	puint32(b[1:5], q.Version)
	puint64(b[5:13], q.Path)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (q *Qid) UnmarshalBinary(data []byte) error {
	if len(data) < 13 {
		return io.ErrUnexpectedEOF
	}

	data = guint8(data, &q.Type)
	data = guint32(data, &q.Version)
	data = guint64(data, &q.Path)
	return nil
}

func (q *Qid) unmarshal(b []byte) []byte {
	b = guint8(b, &q.Type)
	b = guint32(b, &q.Version)
	b = guint64(b, &q.Path)
	return b
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
	return fmt.Sprintf("%.16x %d %q", q.Path, q.Version, t)
}

type Message struct {
	Type    uint8
	Fid     uint32
	Tag     uint16
	Msize   uint32   // Tversion, Rversion
	Version string   // Tversion, Rversion
	Oldtag  uint16   // Tflush
	Ename   string   // Rerror
	Qid     Qid      // Rauth, Rattach, Ropen, Rcreate
	Iounit  uint32   // Ropen, Rcreate
	Afid    uint32   // Tauth, Tattach
	Uname   string   // Tauth, Tattach
	Aname   string   // Tauth, Tattach
	Perm    uint32   // Tcreate
	Name    string   // Tcreate
	Mode    uint8    // Tcreate, Topen
	Newfid  uint32   // Twalk
	Wname   []string // Twalk
	Wqid    []Qid    // Rwalk
	Offset  uint64   // Tread, Twrite
	Count   uint32   // Tread, Rwrite
	Data    []byte   // Twrite, Rread, Twstat, Rstat
}

func (m Message) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (m *Message) UnmarshalBinary(data []byte) error {
	return nil
}

func (m Message) String() string { return "" }
