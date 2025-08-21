// New returns a default OS filesystem instance
package fs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// New returns a default OS filesystem instance
func New() FS {
	return OS{}
}

//go:generate mockgen -source=fs.go -destination=mocks/fs_mock.go -package=mocks

// FS defines the minimal file system operations we rely on (kept narrow for swapability).
type FS interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	MkdirAll(path string, perm fs.FileMode) error
	Stat(path string) (fs.FileInfo, error)
	Remove(path string) error
	Rename(oldPath, newPath string) error
}

// OS implements FS using the local operating system.
type OS struct {
	RootPath string // Optional root directory if no path is provided
}

func (o OS) ReadFile(path string) ([]byte, error) {
	if o.RootPath != "" {
		path = filepath.Join(o.RootPath, path)
	}
	return os.ReadFile(path)
}

func (o OS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if o.RootPath != "" {
		path = filepath.Join(o.RootPath, path)
	}
	return os.WriteFile(path, data, perm)
}
func (o OS) MkdirAll(path string, perm fs.FileMode) error {
	if o.RootPath != "" {
		path = filepath.Join(o.RootPath, path)
	}
	return os.MkdirAll(path, perm)
}
func (o OS) Stat(path string) (fs.FileInfo, error) {
	if o.RootPath != "" {
		path = filepath.Join(o.RootPath, path)
	}
	return os.Stat(path)
}
func (o OS) Remove(path string) error {
	if o.RootPath != "" {
		path = filepath.Join(o.RootPath, path)
	}
	return os.Remove(path)
}
func (o OS) Rename(oldPath, newPath string) error {
	if o.RootPath != "" {
		oldPath = filepath.Join(o.RootPath, oldPath)
		newPath = filepath.Join(o.RootPath, newPath)
	}
	return os.Rename(oldPath, newPath)
}

// EnsureDir ensures a directory relative or absolute exists.
func EnsureDir(fsys FS, dir string) error { return fsys.MkdirAll(dir, 0o755) }

// Join forms a path from parts (kept here to avoid importing filepath in callers).
func Join(parts ...string) string { return filepath.Join(parts...) }

// TMP implements FS using a temporary in-memory file system.
type TMP struct{}

func (TMP) ReadFile(path string) ([]byte, error)                       { return nil, os.ErrNotExist }
func (TMP) WriteFile(path string, data []byte, perm fs.FileMode) error { return nil }
func (TMP) MkdirAll(path string, perm fs.FileMode) error               { return nil }
func (TMP) Stat(path string) (fs.FileInfo, error)                      { return nil, os.ErrNotExist }
func (TMP) Remove(path string) error                                   { return nil }
func (TMP) Rename(o, n string) error                                   { return nil }
