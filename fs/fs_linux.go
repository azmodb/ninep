package fs

import (
	"runtime"

	"github.com/azmodb/pkg/log"
	"golang.org/x/sys/unix"
)

// setfsid changes the values of the caller's file system uid and gid
// that the Linux kernel uses to check for all accesses to the file
// system.
//
// setfsid wires the calling goroutine to its current operating system
// thread. The calling goroutine will always execute in that thread,
// and no other goroutine will execute in it.
func (fs *osFS) setfsid(uid, gid uint32) {
	runtime.LockOSThread()
	if err := unix.Setfsuid(int(uid)); err != nil {
		runtime.UnlockOSThread()
		log.Panicf("setfsuid: %v", err)
	}
	if err := unix.Setfsgid(int(gid)); err != nil {
		runtime.UnlockOSThread()
		log.Panicf("setfsgid: %v", err)
	}
}

func (fs *osFS) resetfsid() {
	if err := unix.Setfsuid(int(fs.uid)); err != nil {
		runtime.UnlockOSThread()
		log.Panicf("resetfsuid: %v", err)
	}
	if err := unix.Setfsgid(int(fs.gid)); err != nil {
		runtime.UnlockOSThread()
		log.Panicf("resetfsgid: %v", err)
	}
	runtime.UnlockOSThread()
}
