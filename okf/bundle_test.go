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

func TestValidateBundleFlagsInvalidUTF8(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// 0xff is not valid UTF-8 in any position.
	invalid := append([]byte("---\ntype: Service\n---\n\nbroken: "), 0xff, 0xfe)
	if err := os.WriteFile(filepath.Join(root, "bad.md"), invalid, 0o644); err != nil {
		t.Fatal(err)
	}

	report := ValidateBundle(root)
	if report.Valid() {
		t.Fatal("ValidateBundle() is valid, want invalid UTF-8 error")
	}
	if !issuesContain(report.Errors, "UTF-8") {
		t.Errorf("errors = %v, want UTF-8 error", report.Errors)
	}
}

func TestValidateBundleWarnsAboutNonUnixLineEndings(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	crlf := "---\r\ntype: Service\r\ntitle: CRLF Doc\r\ndescription: Uses CRLF.\r\n---\r\n\r\nbody\r\n"
	if err := os.WriteFile(filepath.Join(root, "crlf.md"), []byte(crlf), 0o644); err != nil {
		t.Fatal(err)
	}

	report := ValidateBundle(root)
	if !report.Valid() {
		t.Fatalf("ValidateBundle() errors = %v, want valid (CRLF is a warning, not an error)", report.Errors)
	}
	if !issuesContain(report.Warnings, "line ending") {
		t.Errorf("warnings = %v, want line-ending warning", report.Warnings)
	}
}

func TestSearchConceptsMatchesTitleTagsAndBody(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	byTitle := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Auth Service", "description": "Handles login."},
		Body:        "Nothing relevant here.",
	}
	byTag := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Billing", "description": "Handles billing.", "tags": []string{"AuthZ"}},
		Body:        "Nothing relevant here.",
	}
	byBody := Document{
		Frontmatter: map[string]any{"type": "Rule", "title": "Session Rule", "description": "Session handling."},
		Body:        "Must always validate the auth token before proceeding.",
	}
	noMatch := Document{
		Frontmatter: map[string]any{"type": "Concept", "title": "Unrelated", "description": "Nothing to do with it."},
		Body:        "Completely unrelated content.",
	}
	for name, doc := range map[string]Document{
		"by-title.md": byTitle, "by-tag.md": byTag, "by-body.md": byBody, "no-match.md": noMatch,
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(doc.String()), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	matches, err := SearchConcepts(root, "AUTH")
	if err != nil {
		t.Fatalf("SearchConcepts() error = %v", err)
	}
	got := map[string]bool{}
	for _, m := range matches {
		got[m.ID] = true
	}
	for _, want := range []string{"by-title", "by-tag", "by-body"} {
		if !got[want] {
			t.Errorf("SearchConcepts(%q) missing %q, got %v", "AUTH", want, got)
		}
	}
	if got["no-match"] {
		t.Errorf("SearchConcepts(%q) unexpectedly matched %q", "AUTH", "no-match")
	}
}

func TestSearchConceptsEmptyQueryMatchesEverything(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{Frontmatter: map[string]any{"type": "Concept", "title": "Anything"}, Body: "body"}
	if err := os.WriteFile(filepath.Join(root, "one.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	matches, err := SearchConcepts(root, "")
	if err != nil {
		t.Fatalf("SearchConcepts() error = %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("SearchConcepts(\"\") = %d matches, want 1", len(matches))
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
