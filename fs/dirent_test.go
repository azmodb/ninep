package fs

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

var expectedEntries = map[string]uint8{
	".":     unix.DT_DIR,
	"..":    unix.DT_DIR,
	"dir1":  unix.DT_DIR,
	"file1": unix.DT_REG,
	"file2": unix.DT_REG,
	"file3": unix.DT_REG,
	"file4": unix.DT_REG,
}

func TestReadDirent(t *testing.T) {
	dir, err := os.Open(filepath.Join("testdata", "dirent"))
	if err != nil {
		t.Fatalf("readDirent: %v", err)
	}
	defer dir.Close()

	entries, err := readDirent(int(dir.Fd()))
	if err != nil {
		t.Fatalf("readDirent: %v", err)
	}

	if len(entries) != len(expectedEntries) {
		t.Fatalf("readDirent: expected %d entries, got %d", len(expectedEntries), len(entries))
	}

	for i, entry := range entries {
		typ, found := expectedEntries[entry.Name]
		if !found {
			t.Fatalf("readDirent: found unexpected name %q", entry.Name)
		}
		if typ != entry.Type {
			t.Fatalf("readDirent: expected dirent type %d, got %d", entry.Type, typ)
		}

		if entry.Offset != uint64(i+1) {
			t.Fatalf("readDirent: expected offset %d, got %d", i+1, entry.Offset)
		}
		if entry.Qid.Version != 0 {
			t.Fatalf("readDirent: expected version 0, got %d", entry.Qid.Version)
		}
	}
}

// func TestReadDirFull(t *testing.T) {
// 	wd, _ := os.Getwd()
// 	root := filepath.Join(wd, "testdata")
// 	fs, err := NewFileSystem(root)
// 	if err != nil {
// 		t.Fatalf("readdir: cannot create filesystem: %v", err)
// 	}
// 	defer fs.Close()

// 	dir, err := fs.Open("/dirent", os.O_RDONLY, 42, 42)
// 	if err != nil {
// 		t.Fatalf("readdir: cannot open root: %v", err)
// 	}
// 	defer dir.Close()

// 	data := make([]byte, 256)
// 	n, err := dir.ReadDir(data, 0)
// 	if err != nil {
// 		t.Fatalf("readdir: read directory entries: %v", err)
// 	}
// 	data = data[:n]

// 	entries, err := unmarshalDirents(data)
// 	if err != nil {
// 		t.Fatalf("readdir: unmarshal directory entries: %v", err)
// 	}

// 	if len(entries) != len(expectedEntries) {
// 		t.Fatalf("readDirent: expected %d entries, got %d", len(expectedEntries), len(entries))
// 	}

// 	for i, entry := range entries {
// 		typ, found := expectedEntries[entry.Name]
// 		if !found {
// 			t.Fatalf("readDirent: found unexpected name %q", entry.Name)
// 		}
// 		if typ != entry.Type {
// 			t.Fatalf("readDirent: expected dirent type %d, got %d", entry.Type, typ)
// 		}

// 		if entry.Offset != uint64(i+1) {
// 			t.Fatalf("readDirent: expected offset %d, got %d", i+1, entry.Offset)
// 		}
// 		if entry.Qid.Version != 0 {
// 			t.Fatalf("readDirent: expected version 0, got %d", entry.Qid.Version)
// 		}
// 	}
// }

// func TestReadDirOffset(t *testing.T) {
// 	wd, _ := os.Getwd()
// 	root := filepath.Join(wd, "testdata")
// 	fs, err := NewFileSystem(root)
// 	if err != nil {
// 		t.Fatalf("readdir offset: cannot create filesystem: %v", err)
// 	}
// 	defer fs.Close()

// 	dir, err := fs.Open("/dirent", os.O_RDONLY, 42, 42)
// 	if err != nil {
// 		t.Fatalf("readdir offset: cannot open root: %v", err)
// 	}
// 	defer dir.Close()

// 	var entries []*proto.Dirent
// 	data := make([]byte, 32)
// 	offset := int64(0)
// 	for {
// 		n, err := dir.ReadDir(data, offset)
// 		if n <= 0 || err == unix.EIO {
// 			break
// 		}
// 		if err != nil {
// 			t.Fatalf("readdir offset: read directry entries failed: %v", err)
// 			break
// 		}

// 		ents, err := unmarshalDirents(data[:n])
// 		if err != nil {
// 			t.Fatalf("readdir offset: unmarshal directory entries: %v", err)
// 		}

// 		offset = int64(ents[len(ents)-1].Offset)
// 		entries = append(entries, ents...)
// 	}

// 	if len(entries) != len(expectedEntries) {
// 		t.Fatalf("readDirent: expected %d entries, got %d", len(expectedEntries), len(entries))
// 	}

// 	for i, entry := range entries {
// 		typ, found := expectedEntries[entry.Name]
// 		if !found {
// 			t.Fatalf("readDirent: found unexpected name %q", entry.Name)
// 		}
// 		if typ != entry.Type {
// 			t.Fatalf("readDirent: expected dirent type %d, got %d", entry.Type, typ)
// 		}

// 		if entry.Offset != uint64(i+1) {
// 			t.Fatalf("readDirent: expected offset %d, got %d", i+1, entry.Offset)
// 		}
// 		if entry.Qid.Version != 0 {
// 			t.Fatalf("readDirent: expected version 0, got %d", entry.Qid.Version)
// 		}
// 	}
// }
