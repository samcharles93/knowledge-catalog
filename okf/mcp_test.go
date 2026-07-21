package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPListConceptsFormatsEachEntry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Users", "description": "User operations"},
		Body:        "body",
	}
	if err := os.WriteFile(filepath.Join(root, "users.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := MCPListConcepts(root)
	if err != nil {
		t.Fatalf("MCPListConcepts() error = %v", err)
	}
	if !strings.Contains(out, "`users`") || !strings.Contains(out, "[Service]") || !strings.Contains(out, "Users") {
		t.Errorf("MCPListConcepts() = %q", out)
	}
	if strings.Contains(out, "index") {
		t.Errorf("MCPListConcepts() should exclude index.md, got %q", out)
	}
}

func TestMCPListConceptsHandlesEmptyBundle(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	out, err := MCPListConcepts(root)
	if err != nil {
		t.Fatalf("MCPListConcepts() error = %v", err)
	}
	if out != "No concepts found." {
		t.Errorf("MCPListConcepts() = %q, want %q", out, "No concepts found.")
	}
}

func TestMCPGetConceptDelegatesToConceptContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Users", "description": "User operations"},
		Body:        "body",
	}
	if err := os.WriteFile(filepath.Join(root, "users.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := MCPGetConcept(root, "users")
	if err != nil {
		t.Fatalf("MCPGetConcept() error = %v", err)
	}
	if !strings.Contains(out, "# Concept: Users") {
		t.Errorf("MCPGetConcept() = %q", out)
	}
}

func TestMCPValidateBundleSummarizesReport(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	invalid := Document{Frontmatter: map[string]any{"title": "Missing type"}, Body: "body"}
	if err := os.WriteFile(filepath.Join(root, "bad.md"), []byte(invalid.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := MCPValidateBundle(root)
	if err != nil {
		t.Fatalf("MCPValidateBundle() error = %v", err)
	}
	if !strings.Contains(out, "Bundle Valid: false") || !strings.Contains(out, "Total Concepts: 1") {
		t.Errorf("MCPValidateBundle() = %q", out)
	}
}
