package okf

import (
	"os"
	"path/filepath"
	"testing"
)

func newDoc(title string) Document {
	return Document{
		Frontmatter: map[string]any{"type": "Module", "title": title},
		Body:        "# " + title + "\n",
	}
}

// TestExportBundlePrunesStaleFilesWithinOwnedPrefixes is a direct unit test
// of the shared exportBundle helper (independent of any specific extractor):
// re-exporting with a different concept set must delete stale files under
// the caller-declared owned prefixes, but must leave everything outside
// those prefixes untouched.
func TestExportBundlePrunesStaleFilesWithinOwnedPrefixes(t *testing.T) {
	t.Parallel()

	out := t.TempDir()

	first := map[string]Document{
		"codebase/pkg/sub/old": newDoc("old"),
		"api/keep":             newDoc("keep"), // outside the owned prefix
	}
	if _, err := exportBundle(out, func() (map[string]Document, error) { return first, nil }, []string{"codebase"}); err != nil {
		t.Fatalf("first exportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "pkg", "sub", "old.md")); err != nil {
		t.Fatalf("codebase/pkg/sub/old.md not written: %v", err)
	}

	second := map[string]Document{
		"codebase/new": newDoc("new"),
	}
	if _, err := exportBundle(out, func() (map[string]Document, error) { return second, nil }, []string{"codebase"}); err != nil {
		t.Fatalf("second exportBundle() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "codebase", "pkg", "sub", "old.md")); !os.IsNotExist(err) {
		t.Errorf("codebase/pkg/sub/old.md should have been pruned, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "pkg")); !os.IsNotExist(err) {
		t.Errorf("now-empty codebase/pkg directory should have been removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "new.md")); err != nil {
		t.Errorf("codebase/new.md not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "api", "keep.md")); err != nil {
		t.Errorf("api/keep.md outside the owned prefix should not have been touched: %v", err)
	}
}

// TestExportBundleNoPruningWhenPrefixesNil covers extractors like
// WebExtractor that are intentionally additive across runs: passing a nil
// prunePrefixes must leave pre-existing files untouched, however unrelated
// to the new concept set they are.
func TestExportBundleNoPruningWhenPrefixesNil(t *testing.T) {
	t.Parallel()

	out := t.TempDir()
	first := map[string]Document{"references/a": newDoc("a")}
	if _, err := exportBundle(out, func() (map[string]Document, error) { return first, nil }, nil); err != nil {
		t.Fatalf("first exportBundle() error = %v", err)
	}

	second := map[string]Document{"references/b": newDoc("b")}
	if _, err := exportBundle(out, func() (map[string]Document, error) { return second, nil }, nil); err != nil {
		t.Fatalf("second exportBundle() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "references", "a.md")); err != nil {
		t.Errorf("references/a.md should survive when prunePrefixes is nil: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "references", "b.md")); err != nil {
		t.Errorf("references/b.md not written: %v", err)
	}
}
