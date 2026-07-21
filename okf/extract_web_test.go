package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWebExtractorBuildsReferenceConcepts(t *testing.T) {
	t.Parallel()

	ext := WebExtractor{
		URLs: []string{"https://example.com/docs/guide"},
		Fetch: func(url string) (Page, error) {
			return Page{URL: url, Title: "Guide", Markdown: "# Guide\n\nContent."}, nil
		},
	}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	doc, ok := concepts["references/example_com_docs_guide"]
	if !ok {
		t.Fatalf("missing reference concept, got %v", keysOf(concepts))
	}
	if doc.Frontmatter["type"] != "Reference" {
		t.Errorf("type = %v, want Reference", doc.Frontmatter["type"])
	}
	if doc.Frontmatter["resource"] != "https://example.com/docs/guide" {
		t.Errorf("resource = %v", doc.Frontmatter["resource"])
	}
}

func TestWebExtractorSkipsFailedFetchesWithoutError(t *testing.T) {
	t.Parallel()

	ext := WebExtractor{
		URLs: []string{"https://bad.example.com"},
		Fetch: func(url string) (Page, error) {
			return Page{}, fmt.Errorf("connection refused")
		},
	}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v, want nil (failures are logged, not fatal)", err)
	}
	if len(concepts) != 0 {
		t.Errorf("len(concepts) = %d, want 0", len(concepts))
	}
}

// TestWebExtractorExportBundleDoesNotPruneAcrossRuns covers the intentional
// exception to the "extractor owns its namespace" pruning rule: web harvests
// are additive by design (users add --url flags across separate runs), so a
// later harvest of a different URL set must not delete references/ concepts
// from an earlier run.
func TestWebExtractorExportBundleDoesNotPruneAcrossRuns(t *testing.T) {
	t.Parallel()

	fetch := func(url string) (Page, error) {
		return Page{URL: url, Title: "Doc", Markdown: "# Doc"}, nil
	}
	out := t.TempDir()

	first := WebExtractor{URLs: []string{"https://example.com/a"}, Fetch: fetch}
	if _, err := first.ExportBundle(out); err != nil {
		t.Fatalf("first ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "references", "example_com_a.md")); err != nil {
		t.Fatalf("references/example_com_a.md not written: %v", err)
	}

	second := WebExtractor{URLs: []string{"https://example.com/b"}, Fetch: fetch}
	if _, err := second.ExportBundle(out); err != nil {
		t.Fatalf("second ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "references", "example_com_a.md")); err != nil {
		t.Errorf("references/example_com_a.md should survive a later harvest with a different URL set: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "references", "example_com_b.md")); err != nil {
		t.Errorf("references/example_com_b.md not written: %v", err)
	}
}
