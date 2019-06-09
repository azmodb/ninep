package posix

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func checkRecords(t *testing.T, records []Record, num int, want map[string]uint8) {
	if len(records) != num {
		t.Fatalf("ReadDir: expected %d records, got %d", num, len(records))
	}

	for _, rec := range records {
		typ, found := want[rec.Name]
		if !found {
			t.Fatalf("ReadDir: found unexpected name %q", rec.Name)
		}
		if typ != rec.Type {
			t.Fatalf("ReadDir: expected dirent type %d, got %d", rec.Type, typ)
		}

		//if rec.Offset != uint64(i+1) {
		//	t.Fatalf("ReadDir: expected offset %d, got %d", i+1, rec.Offset)
		//}
	}
}

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
		t.Fatalf("readDir: %v", err)
	}
	defer dir.Close()

	records, err := readDir(int(dir.Fd()))
	if err != nil {
		t.Fatalf("readDir: %v", err)
	}

	checkRecords(t, records, len(expectedRecords), expectedRecords)
}

func TestReadDirFull(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "dirent"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	dir := &pfile{f: f}
	defer dir.Close()

	records, err := dir.ReadDir(0)
	if err != nil {
		t.Errorf("ReadDir: unexpected error: %v", err)
	}

	checkRecords(t, records, len(expectedRecords), expectedRecords)
}

func TestReadDirOffset(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "dirent"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	dir := &pfile{f: f}
	defer dir.Close()

	for _, test := range []struct {
		offset int64

		err error
	}{
		{0, nil},
		{1, nil},
		{2, nil},
		{3, nil},
		{4, nil},
		{5, nil},
		{6, nil},
		{7, io.EOF},
	} {
		records, err := dir.ReadDir(test.offset)
		if err != test.err {
			t.Errorf("ReadDir: unexpected error: %v", err)
		}

		want := len(expectedRecords) - int(test.offset)
		checkRecords(t, records, want, expectedRecords)
	}
}
