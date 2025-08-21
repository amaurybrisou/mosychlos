package pdf

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:generate mockgen -source=pdf.go -destination=mocks/pdf_mock.go -package=mocks

type Converter interface {
	Convert(mdPath string) (string, error)
}

// converter converts markdown files to PDF via pandoc with engine fallbacks.
type converter struct {
	engines  []string // preferred engines order
	sanitize bool     // unicode sanitization fallback
}

// Option configures the Converter.
type Option func(*converter)

// WithEngines overrides engine order (default: xelatex, lualatex).
func WithEngines(e []string) Option { return func(c *converter) { c.engines = e } }

// WithSanitize enables/disables unicode sanitization fallback (default true).
func WithSanitize(v bool) Option { return func(c *converter) { c.sanitize = v } }

// New returns a Converter with defaults.
func New(opts ...Option) *converter {
	c := &converter{engines: []string{"xelatex", "lualatex"}, sanitize: true}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Convert turns a markdown file into a PDF. Returns path to PDF.
// The input markdown file is read and converted to PDF format using pandoc.
// output path is the same as the input path with .pdf extension.
func (c *converter) Convert(mdPath string) (string, error) {
	if mdPath == "" {
		return "", errors.New("pdf: empty markdown path")
	}
	if _, err := exec.LookPath("pandoc"); err != nil {
		return "", fmt.Errorf("pandoc not found in PATH: %w", err)
	}
	base := strings.TrimSuffix(mdPath, filepath.Ext(mdPath))
	pdfPath := base + ".pdf"

	for _, eng := range c.engines {
		if eng == "" {
			continue
		}
		if _, err := exec.LookPath(eng); err != nil {
			continue
		}
		if out, err := exec.Command("pandoc", mdPath, "-o", pdfPath, "--pdf-engine="+eng).CombinedOutput(); err == nil {
			return pdfPath, nil
		} else {
			slog.Warn("pdf: engine failed", "engine", eng, "err", err, "output", string(out))
		}
	}

	if c.sanitize {
		if sanitized, serr := sanitizeMarkdownUnicode(mdPath); serr == nil {
			defer func() { _ = os.Remove(sanitized) }()
			if out, err := exec.Command("pandoc", sanitized, "-o", pdfPath).CombinedOutput(); err == nil {
				return pdfPath, nil
			} else {
				slog.Error("pdf: sanitized pandoc failed", "err", err, "output", string(out))
			}
		}
	}

	// Final attempt plain pandoc.
	if out, err := exec.Command("pandoc", mdPath, "-o", pdfPath).CombinedOutput(); err != nil {
		return "", fmt.Errorf("pandoc failed: %w (output: %s)", err, string(out))
	}
	return pdfPath, nil
}

func sanitizeMarkdownUnicode(mdPath string) (string, error) {
	b, err := os.ReadFile(mdPath)
	if err != nil {
		return "", err
	}
	s := string(b)
	repl := strings.NewReplacer(
		"≤", "<=",
		"≥", ">=",
		"—", "--",
		"–", "-",
		"•", "-",
		"×", "x",
		"“", "\"",
		"”", "\"",
		"‘", "'",
		"’", "'",
		" ", " ",
		"‑", "-",
	)
	s = repl.Replace(s)
	tmp := mdPath + ".ascii.md"
	if err := os.WriteFile(tmp, []byte(s), 0o644); err != nil {
		return "", err
	}
	return tmp, nil
}
