package ninep

import (
	"errors"
	"io"
	"os"
	"time"
)

type File interface {
	Create(name string, flag int, perm os.FileMode) error
	Stat() (os.FileInfo, error)

	io.WriterAt
	io.ReaderAt
	io.Closer
}

type file struct {
	children map[string]*file
	data     []byte

	mode  os.FileMode
	mtime time.Time
	atime time.Time
	name  string
}

func newFile(mode os.FileMode) File { return &file{} }

func (f *file) Create(name string, flag int, perm os.FileMode) error {
	return nil
}

func (f file) Stat() (os.FileInfo, error) { return f, nil }

func (f *file) WriteAt(p []byte, offset int64) (int, error) {
	if f.mode&os.ModeDir != 0 {
		return 0, errors.New("is a directory")
	}
	if offset < 0 {
		return 0, errors.New("negative offset")
	}
	size := int64(len(f.data))
	if offset > size {
		offset = size
	}
	if f.data == nil {
		f.data = make([]byte, len(p))
	}

	f.mtime = time.Now()
	n := copy(f.data[offset:], p)
	if n < len(p) {
		data := grow(f.data, len(f.data)+len(p)-n)
		copy(data[len(f.data):], p[n:])
		f.data = data
	}

	return len(p), nil
}

func grow(data []byte, size int) []byte {
	if cap(data) < size {
		b := make([]byte, size)
		copy(b, data)
		data = b
	}
	return data[:size]
}

func (f *file) ReadAt(p []byte, offset int64) (int, error) {
	if f.mode&os.ModeDir != 0 {
		return 0, errors.New("is a directory")
	}
	if offset < 0 {
		return 0, errors.New("negative offset")
	}
	size := int64(len(f.data))
	if offset > size {
		return 0, io.EOF
	}
	if int64(len(p)) > size-offset {
		p = p[:size-offset]
	}

	f.atime = time.Now()
	return copy(p, f.data[offset:]), nil
}

func (f *file) Close() error { return nil }

// file implements os.FileInfo
func (f file) IsDir() bool        { return f.mode&os.ModeDir != 0 }
func (f file) Size() int64        { return int64(len(f.data)) }
func (f file) Name() string       { return f.name }
func (f file) Mode() os.FileMode  { return f.mode }
func (f file) Sys() interface{}   { return nil }
func (f file) ModTime() time.Time { return f.mtime }

func (f file) AccessTime() time.Time { return f.atime }
