package posix

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

var expectedRecords = map[string]uint8{
	".":     unix.DT_DIR,
	"..":    unix.DT_DIR,
	"dir1":  unix.DT_DIR,
	"file1": unix.DT_REG,
	"file2": unix.DT_REG,
	"file3": unix.DT_REG,
	"file4": unix.DT_REG,
}

func TestReadDir(t *testing.T) {
	dir, err := os.Open(filepath.Join("testdata", "dirent"))
	if err != nil {
		t.Fatalf("redDir: %v", err)
	}
	defer dir.Close()

	records, err := readDir(int(dir.Fd()))
	if err != nil {
		t.Fatalf("redDir: %v", err)
	}

	if len(records) != len(expectedRecords) {
		t.Fatalf("redDir: expected %d records, got %d", len(expectedRecords), len(records))
	}

	for i, entry := range records {
		typ, found := expectedRecords[entry.Name]
		if !found {
			t.Fatalf("redDir: found unexpected name %q", entry.Name)
		}
		if typ != entry.Type {
			t.Fatalf("redDir: expected dirent type %d, got %d", entry.Type, typ)
		}

		if entry.Offset != uint64(i+1) {
			t.Fatalf("redDir: expected offset %d, got %d", i+1, entry.Offset)
		}
	}
}
