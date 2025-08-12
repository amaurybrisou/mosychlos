package persist

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	pfs "github.com/amaurybrisou/mosychlos/pkg/fs"
)

//go:generate mockgen -source=manager.go -destination=mocks/manager_mock.go -package=mocks

// FileSystem narrows fs dependency (interface duplication avoids module cycle if relocated).
type FileSystem interface {
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, os.FileMode) error
	MkdirAll(string, os.FileMode) error
	Rename(string, string) error
}

// PDFConverter minimal subset we use from pdf.Converter.
type PDFConverter interface {
	Convert(mdPath string) (string, error)
}

// Manager handles artifact persistence relative to a base directory.
type Manager struct {
	base  string
	fs    FileSystem
	pdfc  PDFConverter
	nowFn func() time.Time
}

// Option configures Manager.
type Option func(*Manager)

func WithFS(f FileSystem) Option {
	return func(m *Manager) {
		if f != nil {
			m.fs = f
		}
	}
}
func WithPDF(c PDFConverter) Option { return func(m *Manager) { m.pdfc = c } }
func WithNow(fn func() time.Time) Option {
	return func(m *Manager) {
		if fn != nil {
			m.nowFn = fn
		}
	}
}

// New returns a Manager.
func New(base string, opts ...Option) *Manager {
	m := &Manager{base: base, fs: pfs.OS{}, nowFn: time.Now}
	for _, o := range opts {
		o(m)
	}
	return m
}

// EnsureDir ensures (base/rel) exists.
func (m *Manager) EnsureDir(rel string) (string, error) {
	if rel == "" {
		return m.base, m.fs.MkdirAll(m.base, 0o755)
	}
	full := filepath.Join(m.base, rel)
	return full, m.fs.MkdirAll(full, 0o755)
}

// WriteJSON writes object as pretty JSON (atomic) and returns path.
func (m *Manager) WriteJSON(rel string, v any) (string, error) {
	if rel == "" {
		return "", errors.New("persist: empty rel path")
	}
	path := filepath.Join(m.base, rel)
	if err := m.fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	tmp := path + ".tmp"
	if err := m.fs.WriteFile(tmp, b, 0o644); err != nil {
		return "", err
	}
	if err := m.fs.Rename(tmp, path); err != nil {
		slog.Warn("persist: rename failed fallback", "err", err)
		return "", err
	}
	return path, nil
}

// ReadJSON reads JSON file into out.
func (m *Manager) ReadJSON(rel string, out any) error {
	path := filepath.Join(m.base, rel)
	b, err := m.fs.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// WriteMarkdown writes markdown content.
func (m *Manager) WriteMarkdown(rel string, content string) (string, error) {
	if rel == "" {
		return "", errors.New("persist: empty rel path")
	}
	path := filepath.Join(m.base, rel)
	if err := m.fs.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp := path + ".tmp"
	if err := m.fs.WriteFile(tmp, []byte(content), 0o644); err != nil {
		return "", err
	}
	if err := m.fs.Rename(tmp, path); err != nil {
		slog.Warn("persist: rename failed fallback", "err", err)
		return "", err
	}
	return path, nil
}

// WritePDFFromMarkdown creates a markdown file then converts to PDF; returns PDF path.
func (m *Manager) WritePDFFromMarkdown(relMD string, content string) (string, error) {
	if m.pdfc == nil {
		return "", errors.New("persist: pdf converter nil")
	}
	mdPath, err := m.WriteMarkdown(relMD, content)
	if err != nil {
		return "", err
	}
	pdfPath, err := m.pdfc.Convert(mdPath)
	if err != nil {
		return "", fmt.Errorf("pdf convert: %w", err)
	}
	return pdfPath, nil
}

// TimestampedName builds a file name with current timestamp prefix (YYYYMMDD-HHMMSS).
func (m *Manager) TimestampedName(base, ext string) string {
	ts := m.nowFn().Format("20060102-150405")
	return fmt.Sprintf("%s-%s.%s", ts, base, ext)
}
