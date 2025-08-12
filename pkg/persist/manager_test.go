package persist

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/pdf"
)

func TestManagerJSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := New(dir)
	type sample struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	in := sample{A: 3, B: "x"}
	name := "test/sample.json"
	if _, err := m.WriteJSON(name, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	var out sample
	if err := m.ReadJSON(name, &out); err != nil {
		t.Fatalf("read: %v", err)
	}
	if out != in {
		t.Fatalf("mismatch: %#v != %#v", out, in)
	}
}

func TestManagerMarkdown(t *testing.T) {
	dir := t.TempDir()
	m := New(dir)
	md := "# Title\nBody"
	rel := "report/r.md"
	p, err := m.WriteMarkdown(rel, md)
	if err != nil {
		t.Fatalf("markdown: %v", err)
	}
	b, _ := os.ReadFile(p)
	if string(b) != md {
		t.Fatalf("content mismatch")
	}
}

func TestTimestampedName(t *testing.T) {
	dir := t.TempDir()
	fixed := func() time.Time { return time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC) }
	m := New(dir, WithNow(fixed))
	name := m.TimestampedName("contextpack", "json")
	exp := "20250102-030405-contextpack.json"
	if name != exp {
		t.Fatalf("expected %s got %s", exp, name)
	}
}

// NOTE: PDF conversion relies on pandoc & engines; we skip if not present to keep tests hermetic.
func TestPDFSkipIfMissing(t *testing.T) {
	dir := t.TempDir()
	m := New(dir, WithPDF(pdf.New()))
	if _, err := exec.LookPath("pandoc"); err != nil {
		t.Skip("pandoc not installed")
	}
	rel := filepath.Join("pdf", "doc.md")
	if _, err := m.WritePDFFromMarkdown(rel, "# hi"); err != nil {
		t.Fatalf("pdf: %v", err)
	}
}
