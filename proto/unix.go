package proto

import (
	"fmt"
	"io"
	"os"

	"github.com/azmodb/ninep/binary"
	"golang.org/x/sys/unix"
)

// Dirent structure defines the format of 9P2000.L directory entries.
type Dirent struct {
	Qid
	Offset uint64
	Type   uint8
	Name   string
}

// String implements fmt.Stringer.
func (m Dirent) String() string {
	return fmt.Sprintf("%s offset:%d type:%d name:%q", m.Qid, m.Offset, m.Type, m.Name)
}

// Len returns the length of the message in bytes.
func (m Dirent) Len() int { return 13 + 8 + 1 + 2 + len(m.Name) }

// Reset resets all state.
func (m *Dirent) Reset() { *m = Dirent{} }

// Encode encodes to the given binary.Buffer.
func (m Dirent) Encode(buf *binary.Buffer) {
	m.Qid.Encode(buf)
	buf.PutUint64(m.Offset)
	buf.PutUint8(m.Type)
	buf.PutString(m.Name)
}

// Decode decodes from the given binary.Buffer.
func (m *Dirent) Decode(buf *binary.Buffer) {
	m.Qid.Decode(buf)
	m.Offset = buf.Uint64()
	m.Type = buf.Uint8()
	m.Name = buf.String()
}

// Marshal implements binary.Marshaler.
func (m Dirent) Marshal(data []byte) ([]byte, int, error) {
	data = binary.PutUint8(data, m.Qid.Type)
	data = binary.PutUint32(data, m.Qid.Version)
	data = binary.PutUint64(data, m.Qid.Path)
	data = binary.PutUint64(data, m.Offset)
	data = binary.PutUint8(data, m.Type)
	data = binary.PutString(data, m.Name)
	return data, m.Len(), nil
}

// Unmarshal implements binary.Unmarshaler.
func (m *Dirent) Unmarshal(data []byte) ([]byte, error) {
	if len(data) < 24 {
		return data, io.ErrUnexpectedEOF
	}
	m.Qid = Qid{
		Type:    binary.Uint8(data[:1]),
		Version: binary.Uint32(data[1:5]),
		Path:    binary.Uint64(data[5:13]),
	}
	m.Offset = binary.Uint64(data[13:21])
	m.Type = binary.Uint8(data[21:22])
	m.Name = binary.String(data[22:])
	return data, nil
}

// Stat describes a file system object.
type Stat = unix.Stat_t

// StatFS describes a file system.
type StatFS = unix.Statfs_t

func encodeTimespec(buf *binary.Buffer, t unix.Timespec) {
	sec, nsec := t.Unix()
	buf.PutUint64(uint64(sec))
	buf.PutUint64(uint64(nsec))
}

func decodeTimespec(buf *binary.Buffer) unix.Timespec {
	return unix.Timespec{
		Sec:  int64(buf.Uint64()),
		Nsec: int64(buf.Uint64()),
	}
}

// UnixDirTypeToQidType converts an unix directory type.
func UnixDirTypeToQidType(typ uint8) uint8 {
	switch typ {
	//case unix.DT_BLK: ???
	case unix.DT_CHR, unix.DT_FIFO, unix.DT_SOCK:
		return TypeAppendOnly
	case unix.DT_DIR:
		return TypeDirectory
	case unix.DT_LNK:
		return TypeSymlink
	}
	return TypeRegular
}

// UnixFileTypeToQidType converts an unix file type.
func UnixFileTypeToQidType(mode uint32) uint8 {
	switch mode & unix.S_IFMT {
	//case unix.S_IFBLK: ???
	case unix.S_IFCHR, unix.S_IFIFO, unix.S_IFSOCK:
		return TypeAppendOnly
	case unix.S_IFDIR:
		return TypeDirectory
	case unix.S_IFLNK:
		return TypeSymlink
	}
	return TypeRegular
}

// FileMode are flags corresponding to file modes. These correspond to
// bits sent over the wire and also to mode_t bits.
type FileMode uint32

// Flags corresponding to file modes. These correspond to bits sent over
// the wire and also correspond to mode_t bits.
const (
	// FileModeMask is a mask of all the file mode bits of FileMode.
	FileModeMask FileMode = 0170000

	// ModeRegular is a mode bit for regular files.
	ModeRegular FileMode = 0100000

	// ModeDirectory is a mode bit for directories.
	ModeDirectory FileMode = 040000

	// ModeSymlink is a mode bit for a symlink.
	ModeSymlink FileMode = 0120000

	// ModeBlockDevice is a mode bit for block devices.
	ModeBlockDevice FileMode = 060000

	// ModeCharacterDevice is a mode bit for a character device.
	ModeCharacterDevice FileMode = 020000

	// ModeSocket is an mode bit for a socket.
	ModeSocket FileMode = 0140000

	// ModeNamedPipe is a mode bit for a named pipe.
	ModeNamedPipe FileMode = 010000

	// ModeGroupID has several special uses. For a directory, it
	// indicates that BSD semantics is to be used for that directory:
	// files created there inherit their group ID from the directory,
	// not from the effective group ID of the creating process, and
	// directories created there will also get the ModeGroupID bit
	// set.
	ModeGroupID FileMode = 0x400
	ModeUserID  FileMode = 0x800
	ModeSticky  FileMode = 0x200

	ModePerm = 0777 // Unix permission bits
)

// NewFileMode returns a FileMode from an os.FileMode.
func NewFileMode(mode os.FileMode) FileMode {
	m := FileMode(mode & os.ModePerm)

	switch {
	case mode&os.ModeCharDevice != 0:
		m |= ModeCharacterDevice
	case mode&os.ModeDevice != 0:
		m |= ModeBlockDevice
	case mode&os.ModeNamedPipe != 0:
		m |= ModeNamedPipe
	case mode&os.ModeSymlink != 0:
		m |= ModeSymlink
	case mode&os.ModeSocket != 0:
		m |= ModeSocket
	case mode&os.ModeDir != 0:
		m |= ModeDirectory
	default:
		m |= ModeRegular
	}

	if mode&os.ModeSetgid != 0 {
		m |= ModeGroupID
	}
	if mode&os.ModeSetuid != 0 {
		m |= ModeUserID
	}
	if mode&os.ModeSticky != 0 {
		m |= ModeSticky
	}
	return m
}

// NewFileModeUnix returns a FileMode from a mode_t.
func NewFileModeUnix(mode uint32) FileMode {
	m := FileMode(mode & ModePerm)

	switch mode & unix.S_IFMT {
	case unix.S_IFCHR:
		m |= ModeCharacterDevice
	case unix.S_IFBLK:
		m |= ModeBlockDevice
	case unix.S_IFIFO:
		m |= ModeNamedPipe
	case unix.S_IFLNK:
		m |= ModeSymlink
	case unix.S_IFSOCK:
		m |= ModeSocket
	case unix.S_IFDIR:
		m |= ModeDirectory
	default:
		m |= ModeRegular
	}

	if mode&unix.S_ISGID != 0 {
		m |= ModeGroupID
	}
	if mode&unix.S_ISUID != 0 {
		m |= ModeUserID
	}
	if mode&unix.S_ISVTX != 0 {
		m |= ModeSticky
	}
	return m
}

// FileMode converts a 9P2000.L file mode to an os.FileMode.
func (m FileMode) FileMode() os.FileMode {
	mode := os.FileMode(m & 0777)

	switch m & FileModeMask {
	case ModeCharacterDevice:
		mode |= os.ModeDevice | os.ModeCharDevice
	case ModeBlockDevice:
		mode |= os.ModeDevice
	case ModeNamedPipe:
		mode |= os.ModeNamedPipe
	case ModeSymlink:
		mode |= os.ModeSymlink
	case ModeSocket:
		mode |= os.ModeSocket
	case ModeDirectory:
		mode |= os.ModeDir
	case ModeRegular:
		// nothing to do
	}

	if m&ModeGroupID != 0 {
		mode |= os.ModeSetgid
	}
	if m&ModeUserID != 0 {
		mode |= os.ModeSetuid
	}
	if m&ModeSticky != 0 {
		mode |= os.ModeSticky
	}
	return mode
}

const rwx = "rwxrwxrwx"

func (m FileMode) String() string {
	mask := m & FileModeMask
	buf := [32]byte{}
	w := 0

	if mask == ModeCharacterDevice {
		buf[w], w = 'D', w+1
		buf[w], w = 'c', w+1
	}
	if mask == ModeBlockDevice {
		buf[w], w = 'D', w+1
	}
	if mask == ModeDirectory {
		buf[w], w = 'd', w+1
	}
	if mask == ModeSymlink {
		buf[w], w = 'L', w+1
	}
	if mask == ModeSocket {
		buf[w], w = 'S', w+1
	}
	if mask == ModeNamedPipe {
		buf[w], w = 'p', w+1
	}

	if m&ModeGroupID != 0 {
		buf[w], w = 'g', w+1
	}
	if m&ModeUserID != 0 {
		buf[w], w = 'u', w+1
	}
	if m&ModeSticky != 0 {
		buf[w], w = 't', w+1
	}
	if w == 0 {
		buf[w] = '-'
		w++
	}

	for i, c := range rwx {
		if m&(1<<uint(9-1-i)) != 0 {
			buf[w] = byte(c)
		} else {
			buf[w] = '-'
		}
		w++
	}
	return string(buf[:w])
}

// Flag is the mode passed to Open and Create operations. These
// correspond to bits sent over the wire.
type Flag uint32

const (
	// FlagModeMask is a mask of valid Flag mode bits.
	FlagModeMask Flag = 0x3

	// FlagReadOnly flag is indicating read-only mode.
	FlagReadOnly Flag = 0x0

	// FlagWriteOnly flag is indicating write-only mode.
	FlagWriteOnly Flag = 0x1

	// FlagReadWrite flag is indicating read-write mode.
	FlagReadWrite Flag = 0x2

	// FlagAppend opens the file in append mode. Before each write, the
	// file offset is positioned at the end of the file. The
	// modification of the file offset and the write operation are
	// performed as a single atomic step.
	FlagAppend Flag = 0x400

	// FlagTruncate is indicating to truncate file when opened, if
	// possible.
	FlagTruncate Flag = 0x200

	// FlagSync is indicating to open for synchronous I/O.
	FlagSync Flag = 0x101000

	// internal used flags
	flagCreate    Flag = 0x40
	flagExclusive Flag = 0x80
)

// NewFlag returns a Flag from an int compatible with open(2).
func NewFlag(flags int) Flag {
	f := Flag(flags & unix.O_ACCMODE)

	switch flags & unix.O_ACCMODE {
	case unix.O_RDONLY:
		f |= FlagReadOnly
	case unix.O_RDWR:
		f |= FlagReadWrite
	case unix.O_WRONLY:
		f |= FlagWriteOnly
	}

	if flags&unix.O_CREAT != 0 {
		f |= flagCreate
	}
	if flags&unix.O_EXCL != 0 {
		f |= flagExclusive
	}

	if flags&unix.O_APPEND != 0 {
		f |= FlagAppend
	}
	if flags&unix.O_TRUNC != 0 {
		f |= FlagTruncate
	}
	if flags&unix.O_SYNC != 0 {
		f |= FlagSync
	}
	return f
}

// Flags converts a Flag to an int compatible with open(2).
func (f Flag) Flags() int {
	flags := int(f & FlagModeMask)

	switch f & FlagModeMask {
	case FlagReadOnly:
		flags |= unix.O_RDONLY
	case FlagReadWrite:
		flags |= unix.O_RDWR
	case FlagWriteOnly:
		flags |= unix.O_WRONLY
	}

	if f&flagCreate != 0 {
		flags |= unix.O_CREAT
	}
	if f&flagExclusive != 0 {
		flags |= unix.O_EXCL
	}

	if f&FlagAppend != 0 {
		flags |= unix.O_APPEND
	}
	if f&FlagTruncate != 0 {
		flags |= unix.O_TRUNC
	}
	if f&FlagSync != 0 {
		flags |= unix.O_SYNC
	}
	return flags
}
