package proto

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestConvertMode(t *testing.T) {
	for i, mode := range []os.FileMode{
		os.ModeDevice | os.ModeCharDevice | 0644,
		os.ModeDir | 0755,
		os.ModeSymlink | 0644,
		os.ModeSocket | 0644,
		os.ModeNamedPipe | 0644,
		os.ModeDevice | 0644,
		0664, 0775, 0222,

		os.ModeDir | os.ModeSticky | 0755,
		os.ModeDir | os.ModeSetgid | 0755,
		os.ModeDir | os.ModeSetuid | 0755,

		os.ModeDevice | os.ModeCharDevice,
		os.ModeNamedPipe,
		os.ModeSocket,
		os.ModeDevice,
		os.ModeDir,
		os.ModeSymlink,
		0, /* os.ModeRegular */
	} {
		m := NewMode(mode)

		if mode != m.FileMode() {
			t.Fatalf("mode converter (#%.4d): expected file mode %d, got %d", i, mode, m.FileMode())
		}

		v := m.FileMode()
		if mode.String() != v.String() {
			t.Fatalf("mode converter (#%.4d): expected file mode %q, got %q", i, mode, v)
		}

		if mode.String() != m.String() {
			t.Fatalf("mode converter (#%.4d): expected file mode %q, got %q", i, mode, m)
		}
	}
}

func TestConvertFileModeUnix(t *testing.T) {
	for i, test := range []struct {
		mode uint32
		want Mode
	}{
		{unix.S_IFDIR | unix.S_ISGID | 0755, ModeDirectory | ModeGroupID | 0755},
		{unix.S_IFDIR | unix.S_ISUID | 0755, ModeDirectory | ModeUserID | 0755},
		{unix.S_IFDIR | unix.S_ISVTX | 0755, ModeDirectory | ModeSticky | 0755},

		{unix.S_IFCHR | 0644, ModeCharacterDevice | 0644},
		{unix.S_IFIFO | 0644, ModeNamedPipe | 0644},
		{unix.S_IFDIR | 0644, ModeDirectory | 0644},
		{unix.S_IFBLK | 0644, ModeBlockDevice | 0644},
		{unix.S_IFSOCK | 0644, ModeSocket | 0644},
		{unix.S_IFLNK | 0644, ModeSymlink | 0644},
		{unix.S_IFREG | 0644, ModeRegular | 0644},

		{unix.S_IFCHR, ModeCharacterDevice},
		{unix.S_IFIFO, ModeNamedPipe},
		{unix.S_IFDIR, ModeDirectory},
		{unix.S_IFBLK, ModeBlockDevice},
		{unix.S_IFSOCK, ModeSocket},
		{unix.S_IFLNK, ModeSymlink},
		{unix.S_IFREG, ModeRegular},
	} {
		m := NewModeUnix(test.mode)
		if m != test.want {
			t.Fatalf("mode converter (#%.4d): expected file mode %d, got %d", i, test.want, m)
		}
	}
}

func TestConvertFlag(t *testing.T) {
	for i, flags := range []int{
		os.O_APPEND | os.O_WRONLY,
		os.O_TRUNC | os.O_WRONLY,
		os.O_SYNC | os.O_WRONLY,

		os.O_RDONLY,
		os.O_RDWR,
		os.O_WRONLY,

		os.O_APPEND,
		os.O_APPEND | os.O_RDWR,

		os.O_CREATE,
		os.O_EXCL,
	} {
		f := NewFlag(flags)

		if flags != f.Flags() {
			t.Fatalf("flag converter (#%.4d): expected flag %d, got %d", i, flags, f.Flags())
		}
	}
}
