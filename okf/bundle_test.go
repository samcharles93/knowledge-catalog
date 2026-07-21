package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegenerateIndexes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	services := filepath.Join(root, "services")
	if err := os.MkdirAll(services, 0o755); err != nil {
		t.Fatal(err)
	}
	doc := Document{
		Frontmatter: map[string]any{
			"type": "Service", "title": "Auth Service", "description": "Handles authentication.",
		},
		Body: "# Auth Service\n",
	}
	if err := os.WriteFile(filepath.Join(services, "auth.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	written, err := RegenerateIndexes(root)
	if err != nil {
		t.Fatalf("RegenerateIndexes() error = %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("len(written) = %d, want 2", len(written))
	}
	index, err := os.ReadFile(filepath.Join(root, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), "[services](services/index.md)") {
		t.Errorf("root index = %q", index)
	}
}

func TestValidateBundleFindsInvalidDocumentAndBrokenLink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	valid := Document{
		Frontmatter: map[string]any{"type": "Architecture", "title": "Overview", "description": "System overview"},
		Body:        "Links to [Broken Target](/services/nonexistent.md).",
	}
	invalid := Document{Frontmatter: map[string]any{"title": "Invalid"}, Body: "Missing type."}
	for name, doc := range map[string]Document{"overview.md": valid, "invalid.md": invalid} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(doc.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	report := ValidateBundle(root)
	if report.Valid() {
		t.Fatal("ValidateBundle() is valid, want invalid")
	}
	if !issuesContain(report.Errors, "type") {
		t.Errorf("errors = %v, want missing type", report.Errors)
	}
	if !issuesContain(report.Warnings, "Broken link target") {
		t.Errorf("warnings = %v, want broken link", report.Warnings)
	}
}

func TestValidateBundleAcceptsRelativeAndExternalLinks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	serviceDir := filepath.Join(root, "services")
	if err := os.MkdirAll(serviceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	overview := Document{
		Frontmatter: map[string]any{"type": "Architecture", "title": "Overview", "description": "System overview"},
		Body:        "See [Users](../services/users.md), [section](#details), and [website](https://example.com).",
	}
	users := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Users", "description": "User operations"},
		Body:        "See [Overview](/architecture/overview.md).",
	}
	architectureDir := filepath.Join(root, "architecture")
	if err := os.MkdirAll(architectureDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, doc := range map[string]Document{
		filepath.Join(architectureDir, "overview.md"): overview,
		filepath.Join(serviceDir, "users.md"):         users,
	} {
		if err := os.WriteFile(path, []byte(doc.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	report := ValidateBundle(root)
	if !report.Valid() || len(report.Warnings) != 0 {
		t.Fatalf("ValidateBundle() = errors %v, warnings %v", report.Errors, report.Warnings)
	}
}

func TestValidateBundleReportsMissingDirectory(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "missing")
	report := ValidateBundle(root)
	if report.Valid() || !issuesContain(report.Errors, "does not exist") {
		t.Fatalf("ValidateBundle() = %+v, want missing-directory error", report)
	}
}

func TestValidateBundleWarnsAboutOrphanConcepts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, name := range []string{"one.md", "two.md"} {
		doc := Document{
			Frontmatter: map[string]any{"type": "Concept", "title": name, "description": "An isolated concept"},
			Body:        "No links.",
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(doc.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	report := ValidateBundle(root)
	if !report.Valid() {
		t.Fatalf("ValidateBundle() errors = %v", report.Errors)
	}
	if got := countIssuesContaining(report.Warnings, "Orphan concept"); got != 2 {
		t.Fatalf("orphan warnings = %d, want 2; warnings = %v", got, report.Warnings)
	}
}

func TestContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.md"), []byte("# Service\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	doc := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Users", "description": "User operations", "tags": []string{"auth", "users"}},
		Body:        "# Users\n\nDetails.\n",
	}
	if err := os.WriteFile(filepath.Join(root, "users.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := SummaryContext(root)
	if err != nil {
		t.Fatalf("SummaryContext() error = %v", err)
	}
	if !strings.Contains(summary, "Progressive Disclosure Index") {
		t.Errorf("summary = %q", summary)
	}
	concept, err := ConceptContext(root, "users")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if !strings.Contains(concept, "# Concept: Users") || !strings.Contains(concept, "auth, users") {
		t.Errorf("concept = %q", concept)
	}
}

func issuesContain(issues []Issue, substring string) bool {
	for _, issue := range issues {
		if strings.Contains(issue.Message, substring) {
			return true
		}
	}
	return false
}

func countIssuesContaining(issues []Issue, substring string) int {
	count := 0
	for _, issue := range issues {
		if strings.Contains(issue.Message, substring) {
			count++
		}
	}
	return count
}
