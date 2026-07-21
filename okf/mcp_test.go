package okf

import (
	"net/http"
	"net/http/httptest"
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

func TestMCPSearchConceptsFormatsMatches(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	match := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Auth Service", "description": "Handles login."},
		Body:        "body",
	}
	noMatch := Document{
		Frontmatter: map[string]any{"type": "Concept", "title": "Unrelated", "description": "Nothing to do with it."},
		Body:        "body",
	}
	if err := os.WriteFile(filepath.Join(root, "auth.md"), []byte(match.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "unrelated.md"), []byte(noMatch.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := MCPSearchConcepts(root, "auth")
	if err != nil {
		t.Fatalf("MCPSearchConcepts() error = %v", err)
	}
	if !strings.Contains(out, "`auth`") || !strings.Contains(out, "Auth Service") {
		t.Errorf("MCPSearchConcepts() = %q", out)
	}
	if strings.Contains(out, "Unrelated") {
		t.Errorf("MCPSearchConcepts() should not match unrelated concept, got %q", out)
	}
}

func TestMCPSearchConceptsHandlesNoMatches(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{Frontmatter: map[string]any{"type": "Concept", "title": "Something"}, Body: "body"}
	if err := os.WriteFile(filepath.Join(root, "one.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := MCPSearchConcepts(root, "no-such-term")
	if err != nil {
		t.Fatalf("MCPSearchConcepts() error = %v", err)
	}
	if out != "No matching concepts found." {
		t.Errorf("MCPSearchConcepts() = %q, want %q", out, "No matching concepts found.")
	}
}

func TestMCPRememberWritesConceptAndReturnsConfirmation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	out, err := MCPRemember(root, RememberInput{
		Type:  "Rule",
		Title: "Always run golangci-lint",
		Body:  "Run golangci-lint before every Go commit.",
	})
	if err != nil {
		t.Fatalf("MCPRemember() error = %v", err)
	}
	if !strings.Contains(out, "rules/Always_run_golangci_lint") || !strings.Contains(out, "Rule") ||
		!strings.Contains(out, "Run golangci-lint before every Go commit.") {
		t.Errorf("MCPRemember() = %q", out)
	}

	data, err := os.ReadFile(filepath.Join(root, "rules", "Always_run_golangci_lint.md"))
	if err != nil {
		t.Fatalf("concept file not written: %v", err)
	}
	doc, err := ParseDocument(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Frontmatter["type"] != "Rule" {
		t.Errorf("type = %v, want Rule", doc.Frontmatter["type"])
	}
}

func TestMCPRememberPropagatesValidationError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	_, err := MCPRemember(root, RememberInput{Type: "Codebase", Title: "should fail", Body: "x"})
	if err == nil {
		t.Fatal("MCPRemember() with a reserved type should have errored")
	}
}

func TestNewMultiBundleServerRoutesByName(t *testing.T) {
	t.Parallel()

	mux := NewMultiBundleServer(map[string]string{
		"tau":         t.TempDir(),
		"archie-core": t.TempDir(),
	})
	for _, name := range []string{"tau", "archie-core"} {
		req := httptest.NewRequest(http.MethodGet, "/"+name+"/", nil)
		h, pattern := mux.Handler(req)
		if h == nil || pattern == "" {
			t.Errorf("no handler matched for /%s/, pattern = %q", name, pattern)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/unregistered/", nil)
	_, pattern := mux.Handler(req)
	if pattern != "" {
		t.Errorf("unexpected route match for an unregistered bundle name: pattern = %q", pattern)
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
