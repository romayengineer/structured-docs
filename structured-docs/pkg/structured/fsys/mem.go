package fsys

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type MemFS struct {
	files map[string][]byte
	dirs  map[string]struct{}
}

func NewMemFS() *MemFS {
	return &MemFS{
		files: make(map[string][]byte),
		dirs:  map[string]struct{}{".": {}},
	}
}

func (m *MemFS) addParentDirs(path string) {
	dir := filepath.Dir(path)
	if dir == "." {
		return
	}
	m.dirs[dir] = struct{}{}
	m.addParentDirs(dir)
}

func (m *MemFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	name = filepath.Clean(name)
	m.files[name] = data
	m.addParentDirs(name)
	return nil
}

func (m *MemFS) ReadFile(name string) ([]byte, error) {
	name = filepath.Clean(name)
	data, ok := m.files[name]
	if !ok {
		return nil, fmt.Errorf("file %s: %w", name, fs.ErrNotExist)
	}
	return data, nil
}

func (m *MemFS) ReadDir(name string) ([]fs.DirEntry, error) {
	name = filepath.Clean(name)
	if _, ok := m.dirs[name]; !ok {
		return nil, fmt.Errorf("directory %s: %w", name, fs.ErrNotExist)
	}

	prefix := name
	if prefix != "." {
		prefix += "/"
	}

	var entries []fs.DirEntry
	seen := make(map[string]bool)

	for p := range m.files {
		if !strings.HasPrefix(p, prefix) && p != name {
			continue
		}
		rel := strings.TrimPrefix(p, prefix)
		if rel == "" || strings.Contains(rel, "/") {
			continue
		}
		if !seen[rel] {
			seen[rel] = true
			entries = append(entries, memDirEntry{name: rel, isDir: false})
		}
	}

	for d := range m.dirs {
		if !strings.HasPrefix(d, prefix) || d == name {
			continue
		}
		rel := strings.TrimPrefix(d, prefix)
		if rel == "" || strings.Contains(rel, "/") {
			continue
		}
		if !seen[rel] {
			seen[rel] = true
			entries = append(entries, memDirEntry{name: rel, isDir: true})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return entries, nil
}

func (m *MemFS) MkdirAll(path string, perm os.FileMode) error {
	path = filepath.Clean(path)
	m.dirs[path] = struct{}{}
	m.addParentDirs(path)
	return nil
}

func (m *MemFS) RemoveAll(path string) error {
	path = filepath.Clean(path)
	for p := range m.files {
		if p == path || strings.HasPrefix(p, path+"/") {
			delete(m.files, p)
		}
	}
	for d := range m.dirs {
		if d == path || strings.HasPrefix(d, path+"/") {
			delete(m.dirs, d)
		}
	}
	return nil
}

func (m *MemFS) Walk(root string, fn filepath.WalkFunc) error {
	root = filepath.Clean(root)

	_, dirOk := m.dirs[root]
	_, fileOk := m.files[root]
	if !dirOk && !fileOk {
		return nil
	}

	var allPaths []string
	for p := range m.files {
		if p == root || strings.HasPrefix(p, root+"/") {
			allPaths = append(allPaths, p)
		}
	}
	for d := range m.dirs {
		if d == root || strings.HasPrefix(d, root+"/") {
			allPaths = append(allPaths, d)
		}
	}

	sort.Strings(allPaths)

	seen := make(map[string]bool)
	for _, p := range allPaths {
		if seen[p] {
			continue
		}
		seen[p] = true

		_, isDir := m.dirs[p]
		var info fs.FileInfo
		if isDir {
			info = memFileInfo{name: filepath.Base(p), isDir: true}
		} else {
			info = memFileInfo{name: filepath.Base(p), size: int64(len(m.files[p]))}
		}

		if err := fn(p, info, nil); err != nil {
			return err
		}
	}

	return nil
}

type memDirEntry struct {
	name  string
	isDir bool
}

func (e memDirEntry) Name() string               { return e.name }
func (e memDirEntry) IsDir() bool                { return e.isDir }
func (e memDirEntry) Type() fs.FileMode          { return 0 }
func (e memDirEntry) Info() (fs.FileInfo, error) { return memFileInfo{name: e.name, isDir: e.isDir}, nil }

type memFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (i memFileInfo) Name() string       { return i.name }
func (i memFileInfo) Size() int64         { return i.size }
func (i memFileInfo) Mode() os.FileMode  { return 0644 }
func (i memFileInfo) ModTime() time.Time { return time.Time{} }
func (i memFileInfo) IsDir() bool        { return i.isDir }
func (i memFileInfo) Sys() interface{}   { return nil }
