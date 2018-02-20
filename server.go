package ninep

import (
	"context"
	"errors"
	"net"
	"sync"

	"github.com/azmodb/ninep/proto"
)

type FileServer interface {
	Tattach(ctx context.Context, m *Tattach) bool
	/*
	   	Twalk(ctx context.Context, m *Twalk)
	    Tclunk(ctx context.Context, m *Tclunk)
	    Tcreate(ctx context.Context, m *Tcreate)
	    Topen(ctx context.Context, m *Topen)
	    Tremove(ctx context.Context, m *Tremove)
	    Tread(ctx context.Context, m *Tread)
	    Twrite(ctx context.Context, m *Twrite)
	    Tstat(ctx context.Context, m *Tstat)
	    Twstat(ctx context.Context, m *Twstat)
	*/
}

type Server struct {
	mu       sync.Mutex // protects following
	freeid   map[int64]struct{}
	pending  map[int64]*conn
	opened   int64
	maxConn  int64
	shutdown bool

	fs    FileServer
	msize uint32
	log   proto.Logger
}

//type Option func(*Server) error

func NewServer(fs FileServer, log proto.Logger) (*Server, error) {
	return &Server{
		freeid:  make(map[int64]struct{}),
		pending: make(map[int64]*conn),
		maxConn: 1<<63 - 1, // TODO

		fs:    fs,
		msize: 24 + 2*1024*1024,
		log:   log,
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
		s.opened++
		id = s.opened
	}
	if id >= s.maxConn {
		s.mu.Unlock()
		return 0, errors.New("too many connections")
	}

	s.pending[id] = c
	s.mu.Unlock()
	return id, nil
}

func (s *Server) deleteConn(id int64) {
	s.mu.Lock()
	s.freeid[id] = struct{}{}
	delete(s.pending, id)
	s.mu.Unlock()
}

func (s *Server) destroyConns() {
	s.mu.Lock()
	for id, c := range s.pending {
		delete(s.pending, id)
		s.freeid[id] = struct{}{}
		c.Close()
	}
	s.mu.Unlock()
}

func (s *Server) Serve(listener net.Listener) error {
	defer s.destroyConns()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func(conn net.Conn) {
			c := newConn(s, conn)
			defer c.Close()

			id, err := s.addConn(c)
			if err != nil {
				s.printf("conn[%d] error: %v", id, err)
				return
			}
			s.printf("insert conn[%d]", id)
			if err = c.Serve(); err != nil {
				s.printf("conn[%d] error: %v", id, err)
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
