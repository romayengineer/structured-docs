package fsys

import (
	"io/fs"
	"os"
	"path/filepath"
)

type FS interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadDir(name string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm os.FileMode) error
	RemoveAll(path string) error
	Walk(root string, fn filepath.WalkFunc) error
}

type OS struct{}

func (OS) ReadFile(name string) ([]byte, error)  { return os.ReadFile(name) }
func (OS) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (OS) ReadDir(name string) ([]fs.DirEntry, error)   { return os.ReadDir(name) }
func (OS) MkdirAll(path string, perm os.FileMode) error  { return os.MkdirAll(path, perm) }
func (OS) RemoveAll(path string) error                    { return os.RemoveAll(path) }
func (OS) Walk(root string, fn filepath.WalkFunc) error   { return filepath.Walk(root, fn) }
