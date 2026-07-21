package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodebaseExtractorProducesOverviewAndModules(t *testing.T) {
	t.Parallel()

	proj := filepath.Join(t.TempDir(), "my-project")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "README.md"), []byte("# My App\nSample app readme.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "main.py"), []byte("def hello(): pass\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := CodebaseExtractor{ProjectRoot: proj}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	overview, ok := concepts["architecture/overview"]
	if !ok {
		t.Fatal("missing architecture/overview concept")
	}
	if !strings.Contains(overview.Body, "Sample app readme.") {
		t.Errorf("overview body = %q, want readme content", overview.Body)
	}
	module, ok := concepts["codebase/main"]
	if !ok {
		t.Fatal("missing codebase/main concept")
	}
	if module.Frontmatter["type"] != "Module" {
		t.Errorf("module type = %v, want Module", module.Frontmatter["type"])
	}
}

func TestCodebaseExtractorIgnoresVendorDirectories(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	ignored := filepath.Join(proj, "node_modules", "pkg")
	if err := os.MkdirAll(ignored, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignored, "index.js"), []byte("module.exports = {};\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := CodebaseExtractor{ProjectRoot: proj}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	for id := range concepts {
		if strings.Contains(id, "node_modules") {
			t.Errorf("concept %q should have been ignored", id)
		}
	}
}

func TestCodebaseExtractorExportBundleWritesFiles(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	if err := os.WriteFile(filepath.Join(proj, "app.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	ext := CodebaseExtractor{ProjectRoot: proj}
	n, err := ext.ExportBundle(out)
	if err != nil {
		t.Fatalf("ExportBundle() error = %v", err)
	}
	if n < 2 {
		t.Fatalf("ExportBundle() wrote %d concepts, want at least 2", n)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "app.md")); err != nil {
		t.Errorf("codebase/app.md not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "index.md")); err != nil {
		t.Errorf("index.md not regenerated: %v", err)
	}
}
