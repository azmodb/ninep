package ninep

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
)

var errShutdown = errors.New("connection is shut down")

type Client struct {
	writer sync.Mutex // exclusive writer/encoder lock
	enc    *Encoder

	dec *Decoder

	mu       sync.Mutex // protects following
	tag      uint16
	pending  map[uint16]*Fcall
	closing  bool
	shutdown bool
	c        io.Closer
}

type Fcall struct {
	Args  *Message
	reply *Message
	err   error
	ch    chan<- *Fcall
}

func (f *Fcall) done() {
	select {
	case f.ch <- f:
		// ok
	default:
		// nothing
	}
}

func (f *Fcall) Err() error { return f.err }

func (f *Fcall) Reply() *Message { return f.reply }

func Dial(ctx context.Context, network, addr string) (*Client, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

func NewClient(rwc io.ReadWriteCloser) *Client {
	c := &Client{
		pending: make(map[uint16]*Fcall),
		enc:     NewEncoder(rwc),
		dec:     NewDecoder(rwc),
		c:       rwc,
	}
	go c.recv()
	return c
}

func (c *Client) Close() error {
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return errShutdown
	}
	c.closing = true
	c.mu.Unlock()
	return c.c.Close()
}

func (c *Client) Do(ctx context.Context, f *Fcall, ch chan<- *Fcall) {
	f.ch = ch
	c.send(f)
}

func (c *Client) next() (uint16, error) {
	c.tag++
	tag := c.tag
	return tag, nil
}

func (c *Client) send(f *Fcall) {
	c.mu.Lock()
	if c.shutdown || c.closing {
		f.err = errShutdown
		c.mu.Unlock()
		f.done()
		return
	}
	tag := f.Args.Tag
	if tag == 0 {
		tag, f.err = c.next()
		if f.err != nil {
			f.done()
		}
	}
	c.pending[tag] = f
	c.mu.Unlock()

	//f.Args.Tag = tag
	c.writer.Lock()
	err := c.enc.Encode(f.Args)
	c.writer.Unlock()
	if err != nil {
		c.mu.Lock()
		f = c.pending[tag]
		delete(c.pending, tag)
		c.mu.Unlock()
		if f != nil {
			f.err = err
			f.done()
		}
	}
}

func (c *Client) recv() {
	var err error
	for err == nil {
		resp := &Message{}
		if err = c.dec.Decode(resp); err != nil {
			break
		}

		tag := resp.Tag
		c.mu.Lock()
		f := c.pending[tag]
		delete(c.pending, tag)
		c.mu.Unlock()

		switch {
		case f == nil:
			// nothing
		case resp.Type == Rerror:
			f.err = errors.New(resp.Ename)
			f.done()
		default:
			f.reply = resp
			f.done()
		}
	}

	c.writer.Lock()
	c.mu.Lock()
	c.shutdown = true
	if err == io.EOF {
		if c.closing {
			err = errShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, f := range c.pending {
		f.err = err
		f.done()
	}
	c.mu.Unlock()
	c.writer.Unlock()
}
