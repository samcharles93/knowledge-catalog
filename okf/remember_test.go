package okf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRememberWritesValidatedConceptUnderTypeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	conceptID, err := Remember(root, RememberInput{
		Type:  "Rule",
		Title: "Always run golangci-lint before committing Go changes",
		Body:  "Run golangci-lint before every Go commit.",
	})
	if err != nil {
		t.Fatalf("Remember() error = %v", err)
	}
	if conceptID != "rules/Always_run_golangci_lint_before_committing_Go_changes" {
		t.Errorf("conceptID = %q, want rules/Always_run_golangci_lint_before_committing_Go_changes", conceptID)
	}

	path, err := ConceptPath(root, []string{"rules", "Always_run_golangci_lint_before_committing_Go_changes"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("concept file not written: %v", err)
	}
	doc, err := ParseDocument(string(data))
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if doc.Frontmatter["type"] != "Rule" {
		t.Errorf("type = %v, want Rule", doc.Frontmatter["type"])
	}
	if doc.Frontmatter["title"] != "Always run golangci-lint before committing Go changes" {
		t.Errorf("title = %v", doc.Frontmatter["title"])
	}
	if doc.Body != "Run golangci-lint before every Go commit." {
		t.Errorf("body = %q", doc.Body)
	}
}

func TestRememberRejectsReservedTypes(t *testing.T) {
	t.Parallel()

	for _, typ := range []string{"Codebase", "API", "Table", "Reference"} {
		t.Run(typ, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			_, err := Remember(root, RememberInput{Type: typ, Title: "should fail", Body: "x"})
			if err == nil {
				t.Fatalf("Remember() with reserved type %q should have errored", typ)
			}
			entries, _ := os.ReadDir(root)
			if len(entries) != 0 {
				t.Errorf("Remember() with reserved type %q should not write anything, found %v", typ, entries)
			}
		})
	}
}

func TestRememberFallsBackForCustomType(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	conceptID, err := Remember(root, RememberInput{
		Type:  "CustomThing",
		Title: "A custom memory",
		Body:  "Body text.",
	})
	if err != nil {
		t.Fatalf("Remember() error = %v", err)
	}
	if conceptID != "customthing/A_custom_memory" {
		t.Errorf("conceptID = %q, want customthing/A_custom_memory", conceptID)
	}
}

func TestRememberOverwritesOnSlugCollision(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	in := RememberInput{Type: "Concept", Title: "Same Title", Body: "first body"}
	if _, err := Remember(root, in); err != nil {
		t.Fatalf("first Remember() error = %v", err)
	}
	in.Body = "second body"
	conceptID, err := Remember(root, in)
	if err != nil {
		t.Fatalf("second Remember() error = %v", err)
	}

	path, err := ConceptPath(root, []string{"concepts", "Same_Title"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("concept file not found: %v", err)
	}
	doc, err := ParseDocument(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Body != "second body" {
		t.Errorf("body = %q, want %q (second call should overwrite, not duplicate)", doc.Body, "second body")
	}
	if conceptID != "concepts/Same_Title" {
		t.Errorf("conceptID = %q, want concepts/Same_Title", conceptID)
	}

	files, err := walkConceptFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Errorf("len(files) = %d, want 1 (collision should overwrite, not duplicate)", len(files))
	}
}

func TestRememberRejectsMissingTypeTitleOrBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   RememberInput
	}{
		{"missing type", RememberInput{Title: "t", Body: "b"}},
		{"missing title", RememberInput{Type: "Rule", Body: "b"}},
		{"missing body", RememberInput{Type: "Rule", Title: "t"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if _, err := Remember(root, tc.in); err == nil {
				t.Fatal("Remember() should have errored")
			}
			entries, _ := os.ReadDir(root)
			if len(entries) != 0 {
				t.Errorf("Remember() should not write anything on validation error, found %v", entries)
			}
		})
	}
}

func TestRememberRegeneratesIndexes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if _, err := Remember(root, RememberInput{Type: "Rule", Title: "Some Rule", Body: "body"}); err != nil {
		t.Fatalf("Remember() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "index.md")); err != nil {
		t.Errorf("bundle root index.md not regenerated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "rules", "index.md")); err != nil {
		t.Errorf("rules/index.md not regenerated: %v", err)
	}
}
