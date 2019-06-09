// Package posix provides the definitions and functions for interacting
// with the filesystem.
//
// The package has been named posix to avoid clashing with the standard
// unix (golang.org/x/sys/unix) package.
package posix

import "golang.org/x/sys/unix"

// FileSystem is the filesystem interface. Any simulated or real system
// should implement this interface.
type FileSystem interface {
}

// File represents a file in the filesystem and is a set of operations
// corresponding to a single node.
type File interface {
	//ReadDir(p []byte, offset int64) (int, error)

	WriteAt(p []byte, offset int64) (int, error)
	ReadAt(p []byte, offset int64) (int, error)

	Close() error
}

// Stat describes a file system object.
type Stat = unix.Stat_t

// StatFS describes a file system.
type StatFS = unix.Statfs_t
