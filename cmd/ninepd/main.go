package main

import (
	"context"
	"flag"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"

	ninep "github.com/azmodb/ninep"
	"github.com/azmodb/ninep/posix"
	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

// Option configures how we set up a 9P2000.L file server before it starts.
type Option struct {
	Export string
	user   string

	//Network string
	Address string
}

func newOptionFlag() *Option {
	o := &Option{}
	flag.StringVar(&o.Export, "export", "/tmp", "export path (comm-seperated)")
	flag.StringVar(&o.Address, "addr", "0.0.0.0:5640", "bind to host address")
	flag.StringVar(&o.user, "user", "", "run-as user")
	return o
}

func (o *Option) UID() int { return 0 }
func (o *Option) GID() int { return 0 }

func main() {
	opt := newOptionFlag()
	flag.Parse()

	fs, err := posix.Open(opt.Export, opt.UID(), opt.GID())
	if err != nil {
		log.Fatalf("cannot open local filesystem: %v", err)
	}
	defer fs.Close()

	ctx := context.Background()
	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", opt.Address)
	if err != nil {
		log.Fatalf("cannot announce network: %v", err)
	}

	c := &onceCloser{Closer: listener}
	s := ninep.NewServer(fs)
Loop:
	for {
		select {
		case err = <-listen(s, listener):
			c.Close()
			break Loop
		case <-shutdown():
			c.Close()
		}
	}

	log.Debugf("listener received error: %v", err)
	log.Infof("9P2000.L server is shut down")
}

func shutdown() <-chan os.Signal {
	ch := make(chan os.Signal)
	signal.Notify(ch, unix.SIGTERM, unix.SIGINT, unix.SIGKILL)
	return ch
}

type onceCloser struct {
	sync.Once
	io.Closer
}

func (c *onceCloser) Close() (err error) {
	c.Once.Do(func() { err = c.Close() })
	return err
}

func listen(s *ninep.Server, listener net.Listener) <-chan error {
	ch := make(chan error, 1)
	go func() { ch <- s.Listen(listener) }()
	return ch
}
