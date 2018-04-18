package ninep

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"

	"github.com/azmodb/ninep/proto"
)

// Conn represents an 9P2000 RPC client. There may be multiple outstanding
// calls associated with a single client, and a client may be used by
// multiple goroutines simultaneously.
type Conn struct {
	enc   *encoder // exclusive 9P2000 stream encoder
	dec   proto.Decoder
	c     io.Closer
	msize int

	mu       sync.Mutex // protects following
	pending  map[uint16]chan<- interface{}
	freetag  map[uint16]struct{}
	freefid  map[uint32]struct{}
	tag      uint16
	fid      uint32
	closing  bool
	shutdown bool
}

// ConnOption configures how we set up the connection.
type ConnOption func(*Conn) error

// Dial connects to an 9P2000 server at the specified network address. For
// TCP and UDP networks, the address has the form "host:port". For unix
// networks, the address must be a file system path.
func Dial(ctx context.Context, network, addr string, opts ...ConnOption) (*Conn, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	return NewConn(conn, opts...)
}

// NewConn returns a new client to handle requests to the set of services
// at the other end of the connection.
func NewConn(rwc io.ReadWriteCloser, opts ...ConnOption) (*Conn, error) {
	c := &Conn{
		pending: make(map[uint16]chan<- interface{}),
		freetag: make(map[uint16]struct{}),
		freefid: make(map[uint32]struct{}),

		enc: newEncoder(rwc),
		dec: newDecoder(rwc),
		c:   rwc,
	}
	go c.recv()

	if err := c.version(proto.DefaultMaxMessageLen, proto.Version); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// Close calls the underlying connection Close method. If the connection
// is already shutting down, an error is returned.
func (c *Conn) Close() error {
	c.mu.Lock()
	if c.closing {
		c.mu.Unlock()
		return errors.New("connection is shut down")
	}
	c.closing = true
	c.mu.Unlock()
	return c.c.Close()
}

func send(ch chan<- interface{}, m interface{}) {
	select {
	case ch <- m:
	default:
	}
}

func wait(ch <-chan interface{}, rx proto.Message) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("received invalid response type")
		}
	}()

	m := <-ch
	switch rx := m.(type) {
	case proto.Rerror:
		return proto.Error(rx.Ename())
	case error:
		return rx
	}

	switch rx := rx.(type) {
	case *proto.Rversion:
		*rx = m.(proto.Rversion)
	case *proto.Rauth:
		*rx = m.(proto.Rauth)
	case *proto.Rattach:
		*rx = m.(proto.Rattach)
	case *proto.Rflush:
		*rx = m.(proto.Rflush)
	case *proto.Rwalk:
		*rx = m.(proto.Rwalk)
	case *proto.Ropen:
		*rx = m.(proto.Ropen)
	case *proto.Rcreate:
		*rx = m.(proto.Rcreate)
	case *proto.Rread:
		*rx = m.(proto.Rread)
	case *proto.Rwrite:
		*rx = m.(proto.Rwrite)
	case *proto.Rclunk:
		*rx = m.(proto.Rclunk)
	case *proto.Rremove:
		*rx = m.(proto.Rremove)
	case *proto.Rstat:
		*rx = m.(proto.Rstat)
	case *proto.Rwstat:
		*rx = m.(proto.Rwstat)
	default:
		err = errors.New("received invalid response type")
	}
	return err
}

// version request negotiates the protocol version and message size to
// be used on the connection and initializes the connection for I/O.
func (c *Conn) version(msize uint32, version string) (err error) {
	ch := make(chan interface{}, 1)
	c.pending[NOTAG] = ch

	log.Printf("<- Tversion tag:%d msize:%d version:%q", NOTAG, msize, version)
	if err = c.enc.Tversion(NOTAG, msize, version); err != nil {
		delete(c.pending, NOTAG)
		return err
	}

	var rx proto.Rversion
	if err = wait(ch, &rx); err != nil {
		return err
	}

	msize = rx.Msize()
	if msize > c.dec.MaxMessageSize() {
		return errors.New("invalid msize in version response")
	}
	if msize != c.dec.MaxMessageSize() {
		c.enc.SetMaxMessageSize(msize)
		c.dec.SetMaxMessageSize(msize)
	}
	c.msize = int(msize)

	log.Printf("-> Rversion tag:%d msize:%d version:%q", rx.Tag(), msize, rx.Version())
	return nil
}

// Auth is used to authenticate users on a connection. If the server does
// require authentication, it returns a fid defining a file of type
// proto.QTAUTH that may be read and written to execute an authentication
// protocol. That protocol's definition is not part of 9P2000 itself.
func (c *Conn) Auth(ctx context.Context, user, root string) (*Fid, error) {
	panic("not implemented")
}

// Access identifies the user (uname) and may select the file tree to
// access (aname). The afid argument specifies a fid previously
// established by an auth message. Access returns the root directory of
// the desired file tree.
func (c *Conn) Access(ctx context.Context, fid *Fid, user, root string) (*Fid, error) {
	tag, ch, err := c.register()
	if err != nil {
		return nil, err
	}
	num, err := c.nextFid()
	if err != nil {
		return nil, err
	}

	afid := NOFID
	if fid != nil {
		afid = fid.num
	}

	log.Printf("<- Tattach tag:%d fid:%d afid:%d uname:%q aname:%q", tag, num, afid, user, root)
	if err = c.enc.Tattach(tag, num, afid, user, root); err != nil {
		c.deregister(tag)
		return nil, err
	}

	var rx proto.Rattach
	if err = wait(ch, &rx); err != nil {
		return nil, err
	}

	f := &Fid{num: num, qid: rx.Qid(), c: c}
	log.Printf("-> Rattach tag:%d qid:%s", tag, f.qid)
	return f, err
}

func (c *Conn) nextFid() (uint32, error) {
	var fid uint32
	for fid = range c.freefid {
		delete(c.freefid, fid)
	}
	if fid == 0 {
		c.fid++
		if c.fid == NOFID {
			return 0, errors.New("out of fids")
		}
		fid = c.fid
	}
	return fid, nil
}

func (c *Conn) putFid(fid uint32) {
	c.mu.Lock()
	if fid != 0 && fid != NOFID {
		c.freefid[fid] = struct{}{}
	}
	c.mu.Unlock()
}

func (c *Conn) nextTag() (uint16, error) {
	var tag uint16
	for tag = range c.freetag {
		delete(c.freetag, tag)
	}
	if tag == 0 {
		c.tag++
		if c.tag == NOTAG {
			return 0, errors.New("out of tags")
		}
		tag = c.tag
	}
	return tag, nil
}

func (c *Conn) register() (uint16, <-chan interface{}, error) {
	c.mu.Lock()
	tag, err := c.nextTag()
	if err != nil {
		c.mu.Unlock()
		return 0, nil, err
	}

	ch := make(chan interface{}, 1)
	c.pending[tag] = ch
	c.mu.Unlock()

	return tag, ch, nil
}

func (c *Conn) deregister(tag uint16) chan<- interface{} {
	c.mu.Lock()
	if tag != 0 && tag != NOTAG {
		c.freetag[tag] = struct{}{}
	}
	ch := c.pending[tag]
	delete(c.pending, tag)
	c.mu.Unlock()
	return ch
}

func (c *Conn) recv() {
	var err error
	for err == nil {
		var m proto.Message
		if m, err = c.dec.Decode(); err != nil {
			break
		}

		ch := c.deregister(m.Tag())
		if ch == nil {
			// We've got no pending channel. That usually means that
			// encode() partially failed, and the channel was already
			// removed.
			continue
		}

		send(ch, m)
	}

	c.enc.mu.Lock()
	c.mu.Lock()
	c.shutdown = true
	if err == io.EOF {
		if c.closing {
			err = errors.New("connection is shut down")
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, ch := range c.pending {
		send(ch, err)
	}
	c.mu.Unlock()
	c.enc.mu.Unlock()
}

// Fid identifies a file on the file server. A new Fid is created when
// the user attaches to the file server, or when walking to a file.
type Fid struct {
	qid proto.Qid
	num uint32
	c   *Conn
}

func (f *Fid) Walk(ctx context.Context, names ...string) (*Fid, error) {
	tag, ch, err := f.c.register()
	if err != nil {
		return nil, err
	}
	newfid, err := f.c.nextFid()
	if err != nil {
		return nil, err
	}

	log.Printf("<- Twalk tag:%d fid:%d newfid:%d names:%v", tag, f.num, newfid, names)
	if err = f.c.enc.Twalk(tag, f.num, newfid, names...); err != nil {
		f.c.deregister(tag)
		return nil, err
	}

	var rx proto.Rwalk
	if err = wait(ch, &rx); err != nil {
		return nil, err
	}

	qids := rx.Wqid()
	log.Printf("-> Rwalk tag:%d qids:%v", tag, qids)
	fid := &Fid{num: newfid, c: f.c}
	if len(qids) > 0 {
		fid.qid = qids[len(qids)-1]
	}
	return fid, nil
}

func (f *Fid) Create(ctx context.Context, name string, mode uint8, perm Perm) error {
	tag, ch, err := f.c.register()
	if err != nil {
		return err
	}

	log.Printf("<- Tcreate tag:%d fid:%d name:%q perm:%q mode:%d", tag, f.num, name, perm, mode)
	if err = f.c.enc.Tcreate(tag, f.num, name, uint32(perm), mode); err != nil {
		f.c.deregister(tag)
		return err
	}

	var rx proto.Rcreate
	if err = wait(ch, &rx); err != nil {
		return err
	}
	f.qid = rx.Qid()
	log.Printf("-> Rcreate tag:%d qid:%s iounit:%d", tag, f.qid, rx.Iounit())

	return nil
}

func (f *Fid) Open(ctx context.Context, mode uint8) error {
	tag, ch, err := f.c.register()
	if err != nil {
		return err
	}

	log.Printf("<- Topen tag:%d fid:%d mode:%d", tag, f.num, mode)
	if err = f.c.enc.Topen(tag, f.num, mode); err != nil {
		f.c.deregister(tag)
		return err
	}

	var rx proto.Ropen
	if err = wait(ch, &rx); err != nil {
		return err
	}
	f.qid = rx.Qid()
	log.Printf("-> Ropen tag:%d qid:%s iounit:%d", tag, f.qid, rx.Iounit())
	return nil
}

func (f *Fid) Remove(ctx context.Context) error {
	tag, ch, err := f.c.register()
	if err != nil {
		return err
	}

	log.Printf("<- Tremove tag:%d fid:%d", tag, f.num)
	if err = f.c.enc.Tremove(tag, f.num); err != nil {
		f.c.deregister(tag)
		return err
	}

	var rx proto.Rremove
	if err = wait(ch, &rx); err != nil {
		return err
	}
	log.Printf("-> Rremove tag:%d", tag)

	f.c.putFid(f.num)
	f.num = NOFID
	return nil
}

func (f *Fid) WriteAt(ctx context.Context, data []byte, offset int64) (int, error) {
	done, n := 0, len(data)
	first := true
	for done < n || first {
		want := n - done
		if want > f.c.msize {
			want = f.c.msize
		}
		m, err := f.writeAt(ctx, data[done:done+want], offset)
		done += m
		if err != nil {
			return done, err
		}
		offset += int64(m)
		first = false
	}
	return done, nil
}

func (f *Fid) writeAt(ctx context.Context, data []byte, offset int64) (int, error) {
	tag, ch, err := f.c.register()
	if err != nil {
		return 0, err
	}

	log.Printf("<- Twrite tag:%d fid:%d offset:%d count:%d", tag, f.num, offset, len(data))
	if err = f.c.enc.Twrite(tag, f.num, uint64(offset), data); err != nil {
		f.c.deregister(tag)
		return 0, err
	}

	var rx proto.Rwrite
	if err = wait(ch, &rx); err != nil {
		return 0, err
	}
	count := int(rx.Count())
	log.Printf("-> Rwrite tag:%d count:%d", tag, count)
	return count, nil
}

func (f *Fid) ReadAt(ctx context.Context, data []byte, offset int64) (int, error) {
	n := len(data)
	if n > f.c.msize {
		n = f.c.msize
	}

	tag, ch, err := f.c.register()
	if err != nil {
		return 0, err
	}

	log.Printf("<- Tread tag:%d fid:%d offset:%d count:%d", tag, f.num, offset, n)
	if err = f.c.enc.Tread(tag, f.num, uint64(offset), uint32(n)); err != nil {
		f.c.deregister(tag)
		return 0, err
	}

	var rx proto.Rread
	if err = wait(ch, &rx); err != nil {
		return 0, err
	}
	count := len(rx.Data())
	if count == 0 {
		return 0, io.EOF
	}
	log.Printf("-> Rread tag:%d count:%d", tag, count)
	return copy(data, rx.Data()), nil
}

// Stat inquires about the file identified by fid. The reply will contain
// a machine-independent directory entry.
func (f *Fid) Stat(ctx context.Context) (proto.Stat, error) {
	tag, ch, err := f.c.register()
	if err != nil {
		return proto.Stat{}, err
	}

	log.Printf("<- Tstat tag:%d fid:%d", tag, f.num)
	if err = f.c.enc.Tstat(tag, f.num); err != nil {
		f.c.deregister(tag)
		return proto.Stat{}, err
	}

	var rx proto.Rstat
	if err = wait(ch, &rx); err != nil {
		return proto.Stat{}, err
	}
	stat := rx.Stat()
	log.Printf("-> Rstat tag:%d stat:%v", tag, stat)
	return stat, nil
}

// Close informs the file server that the current file is no longer
// needed by the client.
func (f *Fid) Close() error {
	tag, ch, err := f.c.register()
	if err != nil {
		return err
	}

	log.Printf("<- Tclunk tag:%d fid:%d", tag, f.num)
	if err = f.c.enc.Tclunk(tag, f.num); err != nil {
		f.c.deregister(tag)
		return err
	}

	var rx proto.Rclunk
	if err = wait(ch, &rx); err != nil {
		return err
	}
	log.Printf("-> Rclunk tag:%d", tag)

	return nil
}
