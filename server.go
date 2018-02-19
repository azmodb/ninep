package ninep

import (
	"errors"
	"net"
	"sync"

	"github.com/azmodb/ninep/proto"
)

/*
type session struct {
	writer sync.Mutex // exclusive connection writer
	enc    proto.Encoder

	dec proto.Decoder

	mu       sync.Mutex
	shutdown bool
	c        io.Closer
}

func newSession(conn net.Conn, msize uint32, log proto.Logger) *session {
	return &session{
		enc: proto.NewEncoder(
			conn,
			proto.WithMaxMessageSize(msize),
			proto.WithLogger(log),
		),
		dec: proto.NewDecoder(
			conn,
			proto.WithMaxMessageSize(msize),
			proto.WithLogger(log),
		),
		c: conn,
	}
}

// Close calls the underlying connection Close method. If the connection
// is already shutting down, Close is a noop.
func (s *session) Close() error {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return nil
	}

	err := s.c.Close()
	s.shutdown = true
	s.mu.Unlock()
	return err
}

func (s *session) Serve() error {
	m, err := s.dec.Decode()
	if err != nil {
		return err
	}
	tx, ok := m.(proto.Tversion)
	if !ok {
		s.rerrorf(m.Tag(), "expected Tversion message (%T)", m)
		return fmt.Errorf("expected Tversion message (%T)", m)
	}
	if err = s.handshake(tx); err != nil {
		return err
	}

	for err == nil {
		m, err := s.dec.Decode()
		if err != nil {
			return err
		}

		switch m := m.(type) {
		case proto.Tauth:
			s.rerrorf(m.Tag(), "authentification not required")
		case proto.Tflush:
			// TODO

		case proto.Tattach:
			// TODO
		case proto.Twalk:
			// TODO
		case proto.Rwalk:
			// TODO
		case proto.Topen:
			// TODO
		case proto.Tcreate:
			// TODO
		case proto.Tread:
			// TODO
		case proto.Twrite:
			// TODO
		case proto.Tclunk:
			// TODO
		case proto.Tstat:
			// TODO
		case proto.Twstat:
			// TODO

		default:
			s.rerrorf(m.Tag(), "unexpected message (%T)", m)
			return fmt.Errorf("unexpected message (%T)", m)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

//const protoVersion = "9P2000"

func (s *session) handshake(m proto.Tversion) error {
	version := m.Version()
	if version != protoVersion {
		version = "unknown"
	}

	msize := m.Msize()
	if msize > s.dec.MaxMessageSize() {
		msize = s.dec.MaxMessageSize()
	}

	s.writer.Lock()
	s.enc.SetMaxMessageSize(msize)
	s.dec.SetMaxMessageSize(msize)
	s.enc.Rversion(m.Tag(), msize, version)
	err := s.enc.Flush()
	s.writer.Unlock()
	return err
}

func (s *session) rerrorf(tag uint16, fmt string, args ...interface{}) {
	s.writer.Lock()
	s.enc.Rerrorf(tag, fmt, args...)
	s.enc.Flush()
	s.writer.Unlock()
}

func (s *session) rerror(tag uint16, err error) {
	s.writer.Lock()
	s.enc.Rerror(tag, err)
	s.enc.Flush()
	s.writer.Unlock()
}
*/

type Server struct {
	mu       sync.Mutex // protects following
	pending  map[int64]*conn
	freeid   map[int64]struct{}
	openConn int64
	maxConn  int64
	shutdown bool

	msize uint32
	log   proto.Logger
}

//type Option func(*Server) error

func NewServer(log proto.Logger) (*Server, error) {
	return &Server{
		freeid:   make(map[int64]struct{}),
		pending:  make(map[int64]*conn),
		openConn: 1,
		maxConn:  1<<63 - 1,

		msize: 24 + 2*1024*1024,
		//log:        log,
	}, nil
}

func (s *Server) addConn(c *conn) (int64, error) {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return 0, errors.New("server is shut down")
	}

	var id int64
	for id, _ = range s.freeid {
		delete(s.freeid, id)
	}
	if id == 0 {
		id = s.openConn
	}
	if id >= s.maxConn {
		s.mu.Unlock()
		return 0, errors.New("too many connections")
	}

	s.pending[id] = c
	s.openConn++
	s.mu.Unlock()
	return id, nil
}

func (s *Server) deleteConn(id int64) {
	s.mu.Lock()
	delete(s.pending, id)
	s.freeid[id] = struct{}{}
	s.mu.Unlock()
}

/*
func (s *Server) cleanupSessions() {
	s.mu.Lock()
	for id, sess := range s.pending {
		delete(s.pending, id)
		s.freeid[id] = struct{}{}
		sess.Close()
	}
	s.mu.Unlock()
}
*/

func (s *Server) Serve(listener net.Listener) error {
	//defer s.cleanupSessions()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(conn net.Conn) {
			c := newConn(conn, s.msize, s.log)
			defer c.Close()

			id, err := s.addConn(c)
			if err != nil {
				return // TODO: log
			}
			if err = c.Serve(); err != nil {
				s.printf("session[%d] error: %v", id, err)
			}
			s.deleteConn(id)
		}(conn)
	}
}

func (s *Server) printf(format string, args ...interface{}) {
	if s.log != nil {
		s.log.Printf(format, args...)
	}
}
