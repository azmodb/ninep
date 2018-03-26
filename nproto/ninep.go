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

var minSizeLUT = [28]uint32{
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
	9 + fixedStatLen,  // size[4] Rstat tag[2] stat[n]
	13 + fixedStatLen, // size[4] Twstat tag[2] fid[4] stat[n]
	7,                 // size[4] Rwstat tag[2]
}

var sizeFunc = map[uint8]func(m *Message) uint32{
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
	Rread: func(m *Message) uint32 {
		return minSizeLUT[Rread-100] + uint32(len(m.Data))
	},
	Twrite: func(m *Message) uint32 {
		return minSizeLUT[Twrite-100] + uint32(len(m.Data))
	},
	Rwrite:  func(m *Message) uint32 { return minSizeLUT[Rwrite-100] },
	Tclunk:  func(m *Message) uint32 { return minSizeLUT[Tclunk-100] },
	Rclunk:  func(m *Message) uint32 { return minSizeLUT[Rclunk-100] },
	Tremove: func(m *Message) uint32 { return minSizeLUT[Tremove-100] },
	Rremove: func(m *Message) uint32 { return minSizeLUT[Rremove-100] },
	Tstat:   func(m *Message) uint32 { return minSizeLUT[Tstat-100] },
	Rstat: func(m *Message) uint32 {
		s := m.Stat
		strLen := len(s.Name) + len(s.UID) + len(s.GID) + len(s.MUID)
		return 2 + minSizeLUT[Rstat-100] + uint32(strLen)
	},
	Twstat: func(m *Message) uint32 {
		s := m.Stat
		strLen := len(s.Name) + len(s.UID) + len(s.GID) + len(s.MUID)
		return 2 + minSizeLUT[Twstat-100] + uint32(strLen)
	},
	Rwstat: func(m *Message) uint32 { return minSizeLUT[Rwstat-100] },
}

// Defines some limits on how big some arbitrarily-long 9P2000 fields can
// be.
const (
	// maxWalkName is the maximum allowed number of path elements in a
	// Twalk request.
	maxWalkNames = 16

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

	headerLen = 7 // size[4] type[1] tag[2]
)

// Error represents a 9P2000 protocol error.
type Error string

func (e Error) Error() string { return string(e) }

const (
	errMaxWalkNames = Error("maximum walk elements exceeded")

	errInvalidMessageType = Error("invalid message type")
	errMessageTooLarge    = Error("message too large")
	errMessageTooSmall    = Error("message too small")
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
	UID    string
	GID    string
	MUID   string
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (s Stat) MarshalBinary() ([]byte, error) {
	size := s.len()
	buf := newBuffer(size)
	enc := NewEncoder(&buf)
	enc.pstat(s, uint16(size))
	if enc.err != nil {
		return nil, enc.err
	}
	return buf, nil
}

func (s Stat) len() int {
	strLen := len(s.Name) + len(s.UID) + len(s.GID) + len(s.MUID)
	return fixedStatLen + strLen
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (s *Stat) UnmarshalBinary(data []byte) error {
	if len(data) < 2+fixedStatLen {
		return io.ErrUnexpectedEOF
	}

	s.unmarshal(data[2:])
	return nil
}

func (s *Stat) unmarshal(b []byte) []byte {
	s.Type, b = guint16(b)
	s.Dev, b = guint32(b)
	b = s.Qid.unmarshal(b)
	s.Mode, b = guint32(b)
	s.Atime, b = guint32(b)
	s.Mtime, b = guint32(b)
	s.Length, b = guint64(b)
	s.Name, b = gstring(b)
	s.UID, b = gstring(b)
	s.GID, b = gstring(b)
	s.MUID, b = gstring(b)
	return b
}

func (s Stat) String() string {
	return fmt.Sprintf("type:0x%x dev:0x%x qid:%q mode:%d atime:%d mtime:%d length:%d name:%q uid:%q gid:%q muid:%q",
		s.Type, s.Dev, s.Qid, s.Mode, s.Atime, s.Mtime, s.Length, s.Name, s.UID, s.GID, s.MUID)
}

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
	buf := newBuffer(13)
	enc := NewEncoder(&buf)
	enc.pqid(q)
	if enc.err != nil {
		return nil, enc.err
	}
	return buf, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (q *Qid) UnmarshalBinary(data []byte) error {
	if len(data) < 13 {
		return io.ErrUnexpectedEOF
	}

	q.unmarshal(data)
	return nil
}

func (q *Qid) unmarshal(b []byte) []byte {
	q.Type, b = b[0], b[1:]
	q.Version, b = guint32(b)
	q.Path, b = guint64(b)
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

// Message represents a 9P2000 message and is used to access fields
// common to all 9P2000 messages.
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
	Data    []byte   // Twrite, Rread
	Stat    Stat     // Twstat, Rstat
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (m Message) MarshalBinary() ([]byte, error) {
	buf := newBuffer(128) // should be enough for most messages
	enc := NewEncoder(&buf)
	if err := enc.Encode(&m); err != nil {
		return nil, err
	}
	return buf, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (m *Message) UnmarshalBinary(data []byte) error {
	if _, _, err := gheader(data); err != nil {
		return err
	}
	return unmarshal(data, m)
}

func (m Message) String() string {
	switch m.Type {
	case Tversion:
		return fmt.Sprintf("Tversion tag:%d msize:%d version:%q",
			m.Tag, m.Msize, m.Version)
	case Rversion:
		return fmt.Sprintf("Rversion tag:%d msize:%d version:%q",
			m.Tag, m.Msize, m.Version)
	case Tauth:
		return fmt.Sprintf("Tauth tag:%d afid:%d uname:%q aname:%q",
			m.Tag, m.Afid, m.Uname, m.Aname)
	case Rauth:
		return fmt.Sprintf("Rauth tag:%d qid:%q", m.Tag, m.Qid)
	case Tattach:
		return fmt.Sprintf("Tattach tag:%d fid:%d afid:%d uname:%q aname:%q",
			m.Tag, m.Fid, m.Afid, m.Uname, m.Aname)
	case Rattach:
		return fmt.Sprintf("Rattach tag:%d qid:%q", m.Tag, m.Qid)
	case Rerror:
		return fmt.Sprintf("Rerror tag:%d ename:%q", m.Tag, m.Ename)
	case Tflush:
		return fmt.Sprintf("Tflush tag:%d oldtag:%d", m.Tag, m.Oldtag)
	case Rflush:
		return fmt.Sprintf("Rflush tag:%d", m.Tag)
	case Twalk:
		return fmt.Sprintf("Twalk tag:%d fid:%d newfid:%d wname:%v",
			m.Tag, m.Fid, m.Newfid, m.Wname)
	case Rwalk:
		return fmt.Sprintf("Rwalk tag:%d wqid:%v", m.Tag, m.Wqid)
	case Topen:
		return fmt.Sprintf("Topen tag:%d fid:%d mode:%d", m.Tag, m.Fid, m.Mode)
	case Ropen:
		return fmt.Sprintf("Ropen tag:%d qid:%q iouint:%d", m.Tag, m.Qid, m.Iounit)
	case Tcreate:
		return fmt.Sprintf("Tcreate tag:%d fid:%d name:%q perm:%d mode:%d",
			m.Tag, m.Fid, m.Name, m.Perm, m.Mode)
	case Rcreate:
		return fmt.Sprintf("Rcreate tag:%d qid:%q iouint:%d", m.Tag, m.Qid, m.Iounit)
	case Tread:
		return fmt.Sprintf("Tread tag:%d fid:%d offset:%d count:%d",
			m.Tag, m.Fid, m.Offset, m.Count)
	case Rread:
		return fmt.Sprintf("Rread tag:%d count:%d", m.Tag, m.Count)
	case Twrite:
		return fmt.Sprintf("Twrite tag:%d fid:%d offset:%d count:%d",
			m.Tag, m.Fid, m.Offset, len(m.Data))
	case Rwrite:
		return fmt.Sprintf("Rwrite tag:%d count:%d", m.Tag, m.Count)
	case Tclunk:
		return fmt.Sprintf("Tclunk tag:%d fid:%d", m.Tag, m.Fid)
	case Rclunk:
		return fmt.Sprintf("Rclunk tag:%d", m.Tag)
	case Tremove:
		return fmt.Sprintf("Tremove tag:%d fid:%d", m.Tag, m.Fid)
	case Rremove:
		return fmt.Sprintf("Rremove tag:%d", m.Tag)
	case Tstat:
		return fmt.Sprintf("Tstat tag:%d fid:%d", m.Tag, m.Fid)
	case Rstat:
		return fmt.Sprintf("Rstat tag:%d stat:%q", m.Tag, m.Stat)
	case Twstat:
		return fmt.Sprintf("Twstat tag:%d fid:%d stat:%q", m.Tag, m.Fid, m.Stat)
	case Rwstat:
		return fmt.Sprintf("FidRwstat tag:%d", m.Tag)
	}
	return fmt.Sprintf("unknown type %d", m.Type)
}
