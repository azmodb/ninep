package ninep

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/azmodb/ninep/proto"
)

type FileSystem interface {
	Access(ctx context.Context, auth File, user, root string) (File, error)
	//Auth(ctx context.Context, auth File, user, root string) (File, error)
}

type File interface {
	Create(ctx context.Context, name string, mode uint8, perm Perm) error
	Open(ctx context.Context, mode uint8) error
	Walk(ctx context.Context, names ...string) (File, error)
	Remove(ctx context.Context) error
	Clunk(ctx context.Context) error

	Stat(ctx context.Context) (proto.Stat, error)

	WriteAt(ctx context.Context, data []byte, offset int64) (int, error)
	ReadAt(ctx context.Context, data []byte, offset int64) (int, error)
}

type file struct {
	children map[string]*file
	data     []byte
	stat     *proto.Stat
}

func newFile() File { return &file{} }

func (f *file) Create(_ context.Context, name string, mode uint8, perm Perm) error {
	return nil
}

func (f *file) Open(_ context.Context, mode uint8) error { return nil }

func (f *file) Walk(_ context.Context, names ...string) (File, error) {
	return nil, nil
}

func (f *file) Remove(_ context.Context) error { return nil }

func (f *file) Clunk(_ context.Context) error { return nil }

func (f *file) Stat(_ context.Context) (proto.Stat, error) {
	return proto.Stat{
		Type: f.stat.Type,
		Dev:  f.stat.Dev,
		Qid: proto.Qid{
			Type:    f.stat.Qid.Type,
			Version: f.stat.Qid.Version,
			Path:    f.stat.Qid.Path,
		},
		Mode:   f.stat.Mode,
		Atime:  f.stat.Atime,
		Mtime:  f.stat.Mtime,
		Length: f.stat.Length,
		Name:   f.stat.Name,
		UID:    f.stat.UID,
		GID:    f.stat.GID,
		MUID:   f.stat.MUID,
	}, nil
}

func (f *file) WriteAt(_ context.Context, p []byte, offset int64) (int, error) {
	if f.stat.Mode&DMDIR != 0 {
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

	f.stat.Mtime = uint32(time.Now().Unix())
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

func (f *file) ReadAt(_ context.Context, p []byte, offset int64) (int, error) {
	if f.stat.Mode&DMDIR != 0 {
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

	f.stat.Atime = uint32(time.Now().Unix())
	return copy(p, f.data[offset:]), nil
}
