package okf

import (
	"fmt"
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
