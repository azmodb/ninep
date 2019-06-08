package ninep

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

type fidmap struct{}

func newFidmap() *fidmap { return nil }

func (f *fidmap) Clear() {}

type session struct {
	writer  sync.Mutex // exclusive stream encoder
	enc     *proto.Encoder
	encErr  error
	rlerror proto.Rlerror

	dec    *proto.Decoder
	fidmap *fidmap

	maxDataSize uint32

	mu       sync.Mutex // protects following
	c        io.Closer
	shutdown bool
	donec    chan struct{}
}

func newSession(rwc io.ReadWriteCloser, msize, dsize uint32) *session {
	return &session{
		enc: proto.NewEncoder(rwc, msize),
		dec: proto.NewDecoder(rwc, msize),
		c:   rwc,

		fidmap:      newFidmap(),
		maxDataSize: dsize,

		donec: make(chan struct{}),
	}
}

var errSessionShutdown = errors.New("session is shut down")

func (s *session) Close() error {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return errSessionShutdown
	}
	s.shutdown = true
	err := s.c.Close()
	<-s.donec
	s.mu.Unlock()
	return err
}

func (s *session) encode(tag uint16, m proto.Message) error {
	if s.encErr != nil {
		err := s.encErr
		return err
	}
	err := s.enc.Encode(tag, m)
	if err != nil {
		s.encErr = err
	}
	return err
}

func (s *session) send(tag uint16, m proto.Message) {
	s.writer.Lock()
	err := s.encode(tag, m)
	s.writer.Unlock()
	if err != nil {
		log.Debugf("session: sending response: %v", err)
	}
}

func (s *session) rerror(tag uint16, e error) {
	s.writer.Lock()
	s.rlerror.Errno = uint32(newErrno(e))
	err := s.encode(tag, &s.rlerror)
	s.writer.Unlock()
	if err != nil {
		log.Debugf("session: sending error response: %v", err)
	}
}

func newErrno(err error) unix.Errno {
	if err == nil {
		return 0
	}

	switch e := err.(type) {
	case unix.Errno:
		return e
	case *os.PathError:
		return newErrno(e.Err)
	case *os.SyscallError:
		return newErrno(e.Err)
	}
	return unix.EIO // defaults to IO error
}

func (s *session) tversion(tx *proto.Tversion, rx *proto.Rversion) error {
	if s.enc.MaxMessageSize != s.dec.MaxMessageSize {
		log.Panic("encoder and decoder maximal message size differ")
	}

	if tx.MessageSize < proto.MinMessageSize {
		return proto.ErrMessageTooSmall
	}
	if tx.MessageSize > s.dec.MaxMessageSize {
		tx.MessageSize = s.dec.MaxMessageSize
	}

	rx.MessageSize = tx.MessageSize
	rx.Version = proto.Version
	if tx.Version != proto.Version {
		// http://9p.io/magic/man2html/5/version
		//
		// If the server does not understand the client's version string,
		// it should respond with an Rversion message (not Rerror) with
		// the version string the 7 characters `unknown`.
		rx.Version = "unknown"
		return errVersionNotSupported
	}

	s.fidmap.Clear()
	s.enc.MaxMessageSize = rx.MessageSize
	s.dec.MaxMessageSize = rx.MessageSize
	s.maxDataSize = rx.MessageSize - (proto.FixedReadWriteSize + 1)
	return nil
}

func (s *session) handshake() error {
	mtype, tag, err := s.dec.DecodeHeader()
	if err != nil {
		return err
	}

	if mtype != proto.MessageTversion {
		s.dec.Decode(nil) // discard pending message body
		return unix.EINVAL
	}

	fcall := mustAlloc(proto.MessageTversion)
	defer proto.Release(fcall)

	tx := fcall.Tx.(*proto.Tversion)
	if err = s.dec.Decode(tx); err != nil {
		return err
	}
	log.Debugf("-> %s tag:%d %s", tx.MessageType(), tag, tx)

	rx := fcall.Rx.(*proto.Rversion)
	if err = s.tversion(tx, rx); err != nil {
		if err == errVersionNotSupported {
			s.send(tag, rx)
		} else {
			s.rerror(tag, unix.EINVAL)
		}
		return err
	}
	log.Debugf("<- %s tag:%d %s", rx.MessageType(), tag, rx)
	return s.encode(tag, rx)
}

func (s *session) serve() (err error) {
	if err = s.handshake(); err != nil {
		close(s.donec)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	for err == nil {
		mtype, tag, decErr := s.dec.DecodeHeader()
		if err = decErr; err != nil {
			break
		}

		fcall, ok := proto.Alloc(mtype)
		if !ok {
			err = unix.EINVAL
			s.rerror(tag, err)
			break
		}

		if err = s.dec.Decode(fcall.Tx); err != nil {
			proto.Release(fcall)
			s.rerror(tag, err)
			break
		}

		wg.Add(1)
		go s.call(ctx, fcall, wg)
	}

	cancel()
	wg.Wait()
	close(s.donec)

	return err
}

func (s *session) call(ctx context.Context, fcall *proto.Fcall, wg *sync.WaitGroup) {
	wg.Done()
}
