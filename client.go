package ninep

import (
	"errors"
	"io"
	"sync"

	"github.com/azmodb/ninep/proto"
)

// pool caches allocated but unused values for later reuse. It is used for
// both tags and fids.
//
// A pool is safe for use by multiple goroutines simultaneously.
type pool struct {
	mu    sync.Mutex
	cache []uint32
	cur   uint32
	limit uint32
}

func newPool(start, limit uint32) *pool {
	return &pool{cur: start, limit: limit}
}

// Next selects an arbitrary value from the pool, removes it from the
// pool, and returns it to the caller.
func (p *pool) Next() (uint32, bool) {
	p.mu.Lock()
	if len(p.cache) > 0 {
		v := p.cache[len(p.cache)-1]
		p.cache = p.cache[:len(p.cache)-1]
		p.mu.Unlock()
		return v, true
	}

	if p.cur == p.limit {
		p.mu.Unlock()
		return 0, false
	}

	v := p.cur
	p.cur++
	p.mu.Unlock()
	return v, true
}

// Put returns a value to the pool.
func (p *pool) Put(v uint32) {
	p.mu.Lock()
	p.cache = append(p.cache, v)
	p.mu.Unlock()
}

type call struct{}

func (c *call) done(err error) {}

type Client struct {
	//	writer sync.Mutex // exclusive encoder lock
	dec *proto.Decoder

	tag *pool

	mu       sync.Mutex // protects following
	pending  map[uint16]*call
	shutdown bool
	closing  bool
}

func (c *Client) Close() error { return nil }

var (
	errClientShutdown = errors.New("client is shut down")
	errOutOfTags      = errors.New("out of tags")
)

func (c *Client) send(call *call) {
	c.mu.Lock()
	if c.shutdown || c.closing {
		c.mu.Unlock()
		call.done(errClientShutdown)
		return
	}
	v, ok := c.tag.Next()
	if !ok {
		c.mu.Unlock()
		call.done(errOutOfTags)
		return
	}
	tag := uint16(v)
	c.pending[tag] = call
	c.mu.Unlock()
}

func (c *Client) handle(typ proto.FcallType, buf *proto.Buffer) error {
	return nil
}

func (c *Client) recv() (err error) {
	for err == nil {
		typ, tag, buf, decErr := c.dec.Next()
		if err = decErr; err != nil {
			break
		}

		c.mu.Lock()
		call := c.pending[tag]
		delete(c.pending, tag)
		c.mu.Unlock()

		switch {
		case call == nil:
			err = errors.New("no pending call: " + err.Error())
		case typ == proto.Rerror:
			call.done(errors.New(buf.Rerror()))
		default:
			call.done(c.handle(typ, buf))
		}
	}

	//c.writer.Lock()
	c.mu.Lock()
	c.shutdown = true
	if err == io.EOF {
		if !c.closing {
			err = io.ErrUnexpectedEOF
		} else {
			err = errClientShutdown
		}
	}
	for _, call := range c.pending {
		call.done(err)
	}
	c.mu.Unlock()
	//c.writer.Unlock()
	return err
}
