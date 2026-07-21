package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateVisualizationWritesGraphHTML(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{"type": "Architecture", "title": "Overview", "description": "Overview concept"},
		Body:        "# Overview\nContent.\n",
	}
	if err := os.WriteFile(filepath.Join(bundle, "overview.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "viz.html")
	stats, err := GenerateVisualization(bundle, out, "")
	if err != nil {
		t.Fatalf("GenerateVisualization() error = %v", err)
	}
	if stats.Concepts != 1 {
		t.Errorf("stats.Concepts = %d, want 1", stats.Concepts)
	}
	if stats.Edges != 0 {
		t.Errorf("stats.Edges = %d, want 0", stats.Edges)
	}
	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Overview") {
		t.Errorf("viz.html missing bundle data: %q", content)
	}
	if stats.Bytes != len(content) {
		t.Errorf("stats.Bytes = %d, want %d", stats.Bytes, len(content))
	}
}

func TestGenerateVisualizationLinksBetweenConcepts(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	a := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "A", "description": "Service A"},
		Body:        "See [B](b.md) for details.",
	}
	b := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "B", "description": "Service B"},
		Body:        "No links here.",
	}
	if err := os.WriteFile(filepath.Join(bundle, "a.md"), []byte(a.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "b.md"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "viz.html")
	stats, err := GenerateVisualization(bundle, out, "my-bundle")
	if err != nil {
		t.Fatalf("GenerateVisualization() error = %v", err)
	}
	if stats.Concepts != 2 {
		t.Errorf("stats.Concepts = %d, want 2", stats.Concepts)
	}
	if stats.Edges != 1 {
		t.Errorf("stats.Edges = %d, want 1", stats.Edges)
	}
	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "my-bundle") {
		t.Errorf("viz.html missing bundle name: %q", content)
	}
}

func TestGenerateVisualizationRejectsMissingBundle(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "missing")
	out := filepath.Join(t.TempDir(), "viz.html")
	if _, err := GenerateVisualization(root, out, ""); err == nil {
		t.Fatal("GenerateVisualization() error = nil, want not-found error")
	}
}

func TestGenerateVisualizationSkipsIndexFiles(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundle, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "viz.html")
	stats, err := GenerateVisualization(bundle, out, "")
	if err != nil {
		t.Fatalf("GenerateVisualization() error = %v", err)
	}
	if stats.Concepts != 0 {
		t.Errorf("stats.Concepts = %d, want 0", stats.Concepts)
	}
}
