package fsys

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestNewMemFS_HasRootDir(t *testing.T) {
	m := NewMemFS()
	if _, ok := m.dirs["."]; !ok {
		t.Error("expected root dir '.' to exist")
	}
}

func TestMemFS_WriteRead(t *testing.T) {
	m := NewMemFS()
	err := m.WriteFile("hello.txt", []byte("world"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	data, err := m.ReadFile("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Errorf("expected 'world', got %q", string(data))
	}
}

func TestMemFS_WriteRead_Nested(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("a/b/c.txt", []byte("nested"), 0644)
	data, err := m.ReadFile("a/b/c.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nested" {
		t.Errorf("expected 'nested', got %q", string(data))
	}
}

func TestMemFS_ReadFileNotFound(t *testing.T) {
	m := NewMemFS()
	_, err := m.ReadFile("nonexistent.txt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMemFS_ReadFile_NormalizesPath(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("dir/file.txt", []byte("data"), 0644)
	data, err := m.ReadFile("dir/./file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("expected 'data', got %q", string(data))
	}
}

func TestMemFS_ReadDir(t *testing.T) {
	m := NewMemFS()
	m.MkdirAll("root", 0755)
	m.WriteFile("root/a.txt", []byte("a"), 0644)
	m.WriteFile("root/b.txt", []byte("b"), 0644)
	m.WriteFile("root/sub/c.txt", []byte("c"), 0644)

	entries, err := m.ReadDir("root")
	if err != nil {
		t.Fatal(err)
	}

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	if !names["a.txt"] {
		t.Error("expected a.txt in root dir")
	}
	if !names["b.txt"] {
		t.Error("expected b.txt in root dir")
	}
	if !names["sub"] {
		t.Error("expected sub/ in root dir")
	}
}

func TestMemFS_ReadDir_Subdir(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("dir/a.txt", []byte("a"), 0644)
	m.WriteFile("dir/b.txt", []byte("b"), 0644)

	entries, err := m.ReadDir("dir")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestMemFS_ReadDir_Nested(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("dir/sub/a.txt", []byte("a"), 0644)
	m.WriteFile("dir/sub/b.txt", []byte("b"), 0644)

	entries, err := m.ReadDir("dir/sub")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestMemFS_ReadDir_NotFound(t *testing.T) {
	m := NewMemFS()
	_, err := m.ReadDir("nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestMemFS_MkdirAll(t *testing.T) {
	m := NewMemFS()
	err := m.MkdirAll("a/b/c", 0755)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.dirs["a/b/c"]; !ok {
		t.Error("expected a/b/c dir to exist")
	}
	if _, ok := m.dirs["a/b"]; !ok {
		t.Error("expected a/b dir to exist (parent)")
	}
	if _, ok := m.dirs["a"]; !ok {
		t.Error("expected a dir to exist (parent)")
	}
}

func TestMemFS_RemoveAll_File(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("file.txt", []byte("data"), 0644)
	m.RemoveAll("file.txt")

	_, err := m.ReadFile("file.txt")
	if err == nil {
		t.Error("expected file to be removed")
	}
}

func TestMemFS_RemoveAll_Dir(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("dir/a.txt", []byte("a"), 0644)
	m.WriteFile("dir/sub/b.txt", []byte("b"), 0644)

	m.RemoveAll("dir")

	_, err := m.ReadFile("dir/a.txt")
	if err == nil {
		t.Error("expected nested file to be removed")
	}
	_, err = m.ReadFile("dir/sub/b.txt")
	if err == nil {
		t.Error("expected deeply nested file to be removed")
	}

	if _, ok := m.dirs["dir"]; ok {
		t.Error("expected dir to be removed from directories")
	}
}

func TestMemFS_RemoveAll_Idempotent(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("f.txt", []byte("data"), 0644)
	err := m.RemoveAll("f.txt")
	if err != nil {
		t.Fatal(err)
	}
	err = m.RemoveAll("f.txt")
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemFS_Walk(t *testing.T) {
	m := NewMemFS()
	m.MkdirAll("dir/sub", 0755)
	m.WriteFile("dir/a.txt", []byte("a"), 0644)
	m.WriteFile("dir/sub/b.txt", []byte("b"), 0644)

	var paths []string
	err := m.Walk("dir", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(paths) != 4 {
		t.Fatalf("expected 4 paths, got %v", paths)
	}
}

func TestMemFS_Walk_NonexistentRoot(t *testing.T) {
	m := NewMemFS()
	err := m.Walk("nonexistent", func(path string, info os.FileInfo, err error) error {
		t.Error("callback should not be called for nonexistent root")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemFS_Walk_StopsOnCallbackError(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("f1.txt", []byte("a"), 0644)
	m.WriteFile("f2.txt", []byte("b"), 0644)

	err := m.Walk(".", func(path string, info os.FileInfo, err error) error {
		return filepath.SkipDir
	})
	if err != filepath.SkipDir {
		t.Fatalf("expected filepath.SkipDir, got %v", err)
	}
}

func TestMemFS_ReadDir_Sorted(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("mydir/z.txt", []byte("z"), 0644)
	m.WriteFile("mydir/a.txt", []byte("a"), 0644)
	m.WriteFile("mydir/m.txt", []byte("m"), 0644)

	entries, err := m.ReadDir("mydir")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Name() != "a.txt" || entries[1].Name() != "m.txt" || entries[2].Name() != "z.txt" {
		t.Errorf("expected sorted order, got %v", entryNames(entries))
	}
}

func TestMemFS_ReadDir_Deduplicates(t *testing.T) {
	m := NewMemFS()
	m.WriteFile("root/a.txt", []byte("a"), 0644)

	entries, err := m.ReadDir("root")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, e := range entries {
		if e.Name() == "a.txt" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected a.txt once, got %d", count)
	}
}

func entryNames(entries []fs.DirEntry) []string {
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	return names
}
