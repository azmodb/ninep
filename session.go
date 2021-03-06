package ninep

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"

	"github.com/azmodb/ninep/posix"
	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

type session struct {
	writer  sync.Mutex // exclusive stream encoder
	enc     *proto.Encoder
	encErr  error
	rlerror proto.Rlerror

	dec         *proto.Decoder
	maxDataSize uint32

	addr string
	srv  *service

	mu       sync.Mutex // protects following
	c        io.Closer
	shutdown bool
	donec    chan struct{}
}

func newSession(fs posix.FileSystem, conn net.Conn, msize, dsize uint32) *session {
	return &session{
		enc:         proto.NewEncoder(conn, msize),
		dec:         proto.NewDecoder(conn, msize),
		c:           conn,
		maxDataSize: dsize,

		addr:  conn.RemoteAddr().String(),
		srv:   newService(fs),
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

func (s *session) encode(tag uint16, m proto.Message) (err error) {
	if s.encErr != nil {
		err = s.encErr
		return err
	}
	if err = s.enc.Encode(tag, m); err != nil {
		s.encErr = err
	}
	return err
}

func (s *session) send(tag uint16, m proto.Message) {
	s.writer.Lock()
	err := s.encode(tag, m)
	s.writer.Unlock()
	if err != nil {
		log.Errorf("session: sending response: %v", err)
	}
}

func (s *session) rerror(tag uint16, e error) {
	s.writer.Lock()
	s.rlerror.Errno = uint32(newErrno(e))
	err := s.encode(tag, &s.rlerror)
	s.writer.Unlock()
	if err != nil {
		log.Errorf("session: sending error response: %v", err)
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

	s.srv.fidmap.Clear()
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
	log.Debugf("-> [%s] %s tag:%d %s", s.addr, tx.MessageType(), tag, tx)

	rx := fcall.Rx.(*proto.Rversion)
	if err = s.tversion(tx, rx); err != nil {
		if err == errVersionNotSupported {
			s.send(tag, rx)
		} else {
			s.rerror(tag, unix.EINVAL)
		}
		return err
	}
	log.Debugf("<- [%s] %s tag:%d %s", s.addr, rx.MessageType(), tag, rx)
	return s.encode(tag, rx)
}

func (s *session) serve() (err error) {
	defer close(s.donec)

	if err = s.handshake(); err != nil {
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
		go func(ctx context.Context, fcall *proto.Fcall) {
			log.Debugf("-> [%s] %s tag:%d %s", s.addr, fcall.Tx.MessageType(), tag, fcall.Tx)
			s.call(ctx, fcall)
			if fcall.Err != nil {
				log.Debugf("<- [%s] %s tag:%d %v", s.addr, fcall.Rx.MessageType(), tag, fcall.Err)
				s.rerror(tag, fcall.Err)
			} else {
				log.Debugf("<- [%s] %s tag:%d %s", s.addr, fcall.Rx.MessageType(), tag, fcall.Rx)
				s.send(tag, fcall.Rx)
			}
			wg.Done()
		}(ctx, fcall)
	}

	cancel()
	wg.Wait()
	return err
}

func (s *session) call(ctx context.Context, f *proto.Fcall) {
	var errno unix.Errno
	switch f.Tx.MessageType() {
	case proto.MessageTlattach:
		errno = s.srv.attach(ctx, f.Tx.(*proto.Tlattach), f.Rx.(*proto.Rlattach))
	case proto.MessageTlauth:
		errno = s.srv.auth(ctx, f.Tx.(*proto.Tlauth), f.Rx.(*proto.Rlauth))
	case proto.MessageTflush:
		errno = s.srv.flush(ctx, f.Tx.(*proto.Tflush), f.Rx.(*proto.Rflush))

	case proto.MessageTwalk:
		errno = s.srv.walk(ctx, f.Tx.(*proto.Twalk), f.Rx.(*proto.Rwalk))
	case proto.MessageTread:
		errno = s.srv.read(ctx, f.Tx.(*proto.Tread), f.Rx.(*proto.Rread))
	case proto.MessageTwrite:
		errno = s.srv.write(ctx, f.Tx.(*proto.Twrite), f.Rx.(*proto.Rwrite))
	case proto.MessageTclunk:
		errno = s.srv.clunk(ctx, f.Tx.(*proto.Tclunk), f.Rx.(*proto.Rclunk))
	case proto.MessageTremove:
		errno = s.srv.remove(ctx, f.Tx.(*proto.Tremove), f.Rx.(*proto.Rremove))

	case proto.MessageTstatfs:
		errno = s.srv.statfs(ctx, f.Tx.(*proto.Tstatfs), f.Rx.(*proto.Rstatfs))
	case proto.MessageTlopen:
		errno = s.srv.open(ctx, f.Tx.(*proto.Tlopen), f.Rx.(*proto.Rlopen))
	case proto.MessageTcreate:
		errno = s.srv.create(ctx, f.Tx.(*proto.Tlcreate), f.Rx.(*proto.Rlcreate))
	case proto.MessageTsymlink:
		errno = s.srv.symlink(ctx, f.Tx.(*proto.Tsymlink), f.Rx.(*proto.Rsymlink))
	case proto.MessageTmknod:
		errno = s.srv.mknod(ctx, f.Tx.(*proto.Tmknod), f.Rx.(*proto.Rmknod))
	case proto.MessageTrename:
		errno = s.srv.rename(ctx, f.Tx.(*proto.Trename), f.Rx.(*proto.Rrename))
	case proto.MessageTreadlink:
		errno = s.srv.readlink(ctx, f.Tx.(*proto.Treadlink), f.Rx.(*proto.Rreadlink))
	case proto.MessageTgetattr:
		errno = s.srv.getattr(ctx, f.Tx.(*proto.Tgetattr), f.Rx.(*proto.Rgetattr))
	case proto.MessageTsetattr:
		errno = s.srv.setattr(ctx, f.Tx.(*proto.Tsetattr), f.Rx.(*proto.Rsetattr))
	case proto.MessageTxattrwalk:
		errno = s.srv.xattrwalk(ctx, f.Tx.(*proto.Txattrwalk), f.Rx.(*proto.Rxattrwalk))
	case proto.MessageTxattrcreate:
		errno = s.srv.xattrcreate(ctx, f.Tx.(*proto.Txattrcreate), f.Rx.(*proto.Rxattrcreate))
	case proto.MessageTreaddir:
		errno = s.srv.readdir(ctx, f.Tx.(*proto.Treaddir), f.Rx.(*proto.Rreaddir))
	case proto.MessageTfsync:
		errno = s.srv.fsync(ctx, f.Tx.(*proto.Tfsync), f.Rx.(*proto.Rfsync))
	case proto.MessageTlock:
		errno = s.srv.lock(ctx, f.Tx.(*proto.Tlock), f.Rx.(*proto.Rlock))
	case proto.MessageTgetlock:
		errno = s.srv.getlock(ctx, f.Tx.(*proto.Tgetlock), f.Rx.(*proto.Rgetlock))
	case proto.MessageTlink:
		errno = s.srv.link(ctx, f.Tx.(*proto.Tlink), f.Rx.(*proto.Rlink))
	case proto.MessageTmkdir:
		errno = s.srv.mkdir(ctx, f.Tx.(*proto.Tmkdir), f.Rx.(*proto.Rmkdir))
	case proto.MessageTrenameat:
		errno = s.srv.renameat(ctx, f.Tx.(*proto.Trenameat), f.Rx.(*proto.Rrenameat))
	case proto.MessageTunlinkat:
		errno = s.srv.unlinkat(ctx, f.Tx.(*proto.Tunlinkat), f.Rx.(*proto.Runlinkat))
	}
	if errno != 0 {
		f.Err = errno
	}
}
