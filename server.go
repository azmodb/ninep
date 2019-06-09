package ninep

import (
	"errors"
	"io"
	"math"
	"net"
	"sync"

	"github.com/azmodb/ninep/proto"
	"github.com/azmodb/pkg/pool"
	"golang.org/x/sys/unix"
)

type Server struct {
	mu       sync.Mutex // protects sessions map
	sessions map[int64]io.Closer
	sid      *pool.Generator

	maxMessageSize uint32
	maxDataSize    uint32
}

func NewServer(service Service, opts ...Option) *Server {
	return &Server{
		// TODO: max concurrent sessions
		sid:      pool.NewGenerator(1, math.MaxUint16),
		sessions: make(map[int64]io.Closer),

		maxMessageSize: proto.MaxMessageSize,
		maxDataSize:    proto.MaxDataSize,
	}
}

func (s *Server) Listen(listener net.Listener) (err error) {
	wg := &sync.WaitGroup{}

	for err == nil {
		conn, lerr := listener.Accept()
		if err = lerr; err != nil {
			break
		}

		id, ok := s.sid.Get()
		if !ok {
			conn.Close()
			err = errors.New("out of sessions")
			break
		}
		s.mu.Lock()
		s.sessions[id] = conn
		s.mu.Unlock()

		wg.Add(1)
		go func(conn net.Conn, id int64) {
			sess := newSession(conn, s.maxMessageSize, s.maxDataSize)

			err := sess.serve()
			if err == io.EOF || err == io.ErrUnexpectedEOF {
			}
			if err == unix.EINVAL {
			}

			conn.Close()
			s.mu.Lock()
			delete(s.sessions, id)
			s.mu.Unlock()

			wg.Done()
		}(conn, id)
	}

	s.mu.Lock()
	for id, c := range s.sessions {
		c.Close()
		delete(s.sessions, id)
	}
	s.mu.Unlock()

	wg.Wait()
	return err
}
