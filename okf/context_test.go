package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConceptContextResolvesByAliasWhenIDMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{
			"type": "Service", "title": "User Service", "description": "Handles users.",
			"aliases": []string{"UserMgr", "AuthService"},
		},
		Body: "# User Service\n",
	}
	if err := os.WriteFile(filepath.Join(root, "user-service.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "UserMgr")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if !strings.Contains(out, "# Concept: User Service") {
		t.Errorf("ConceptContext() = %q, want alias-resolved concept", out)
	}
}

func TestConceptContextAliasLookupIsCaseSensitive(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "User Service", "aliases": []string{"UserMgr"}},
		Body:        "body",
	}
	if err := os.WriteFile(filepath.Join(root, "user-service.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "usermgr")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if !strings.Contains(out, "not found in bundle") {
		t.Errorf("ConceptContext() = %q, want not-found (alias match is case-sensitive)", out)
	}
}

func TestConceptContextPrefersDirectIDOverAlias(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	direct := Document{Frontmatter: map[string]any{"type": "Service", "title": "Direct Hit"}, Body: "body"}
	aliased := Document{Frontmatter: map[string]any{"type": "Service", "title": "Aliased", "aliases": []string{"users"}}, Body: "body"}
	if err := os.WriteFile(filepath.Join(root, "users.md"), []byte(direct.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "other.md"), []byte(aliased.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "users")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if !strings.Contains(out, "# Concept: Direct Hit") {
		t.Errorf("ConceptContext() = %q, want direct ID match to win over alias", out)
	}
}

func TestConceptContextUnresolvedIDStillReportsNotFound(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "users.md"), []byte((Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Users"},
		Body:        "body",
	}).String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "nonexistent")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if out != "Concept 'nonexistent' not found in bundle." {
		t.Errorf("ConceptContext() = %q", out)
	}
}

func TestConceptContextIncludesCitedByWhenBacklinksExist(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := Document{Frontmatter: map[string]any{"type": "Service", "title": "Target"}, Body: "target body"}
	citer := Document{Frontmatter: map[string]any{"type": "Service", "title": "Citer"}, Body: "See [Target](target.md)."}
	if err := os.WriteFile(filepath.Join(root, "target.md"), []byte(target.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "citer.md"), []byte(citer.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "target")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if !strings.Contains(out, "Cited by") || !strings.Contains(out, "`citer`") {
		t.Errorf("ConceptContext() = %q, want Cited by section referencing citer", out)
	}
}

func TestConceptContextOmitsCitedByWhenNoBacklinks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{Frontmatter: map[string]any{"type": "Concept", "title": "Alone"}, Body: "body"}
	if err := os.WriteFile(filepath.Join(root, "alone.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "alone")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if strings.Contains(out, "Cited by") {
		t.Errorf("ConceptContext() = %q, want no Cited by section", out)
	}
}
