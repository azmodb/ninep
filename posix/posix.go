package posix

import (
	"io"
	"os"

	"golang.org/x/sys/unix"
)

type pfile struct {
	cache []Record // directory info cache TODO: reload if needed?
	f     *os.File
}

func (f *pfile) WriteAt(p []byte, offset int64) (int, error) {
	if f == nil || f.f == nil {
		return 0, unix.EBADF
	}
	return f.f.WriteAt(p, offset)
}

func (f *pfile) ReadAt(p []byte, offset int64) (int, error) {
	if f == nil || f.f == nil {
		return 0, unix.EBADF
	}
	return f.f.ReadAt(p, offset)
}

func (f *pfile) ReadDir(offset int64) ([]Record, error) {
	if f == nil || f.f == nil {
		return nil, unix.EBADF
	}

	if offset < 0 {
		return nil, unix.EINVAL
	}
	if len(f.cache) == 0 { // initialize directory cache
		cache, err := readDir(int(f.f.Fd()))
		if err != nil {
			return nil, err
		}
		f.cache = cache
	}

	if offset >= int64(len(f.cache)) {
		return nil, io.EOF
	}

	return f.cache[offset:], nil
}

func (f *pfile) Close() error {
	if f == nil || f.f == nil {
		return unix.EBADF
	}
	err := f.f.Close()
	f.f = nil
	return err
}
