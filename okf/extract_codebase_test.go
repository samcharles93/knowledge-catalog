package okf

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRealGitRepo runs `git init` (and, if commit is true, adds and commits
// every file currently in dir) so tests can exercise the git-ls-files-based
// ignore path for real, rather than the markAsRepoRoot fake ".git" marker
// (which only satisfies isRepoRoot's presence check and deliberately is
// NOT a working repo, so gitTrackedFiles falls back to the plain walk).
func initRealGitRepo(t *testing.T, dir string, commit bool) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	if commit {
		run("add", "-A")
		run("commit", "-q", "-m", "initial")
	}
}

// markAsRepoRoot creates the ".git" marker CodebaseExtractor uses to decide
// whether a harvested root is the whole-project root (and therefore owns
// the singleton architecture/overview concept) or a harvested subtree.
func markAsRepoRoot(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestCodebaseExtractorProducesOverviewAndModules(t *testing.T) {
	t.Parallel()

	proj := filepath.Join(t.TempDir(), "my-project")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	markAsRepoRoot(t, proj)
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

// TestCodebaseExtractorSkipsGitignoredFiles is a security regression test:
// a gitignored file sitting in the tree (e.g. a local secrets file) must
// never be harvested and baked as plaintext into a committed concept
// document.
func TestCodebaseExtractorSkipsGitignoredFiles(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	if err := os.WriteFile(filepath.Join(proj, ".gitignore"), []byte("secret.go\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "secret.go"), []byte("package main\n\nconst APIKey = \"sk-should-never-leak\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initRealGitRepo(t, proj, true)

	ext := CodebaseExtractor{ProjectRoot: proj}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	if _, ok := concepts["codebase/secret"]; ok {
		t.Error("codebase/secret should not have been harvested: it's gitignored")
	}
	for id, doc := range concepts {
		if strings.Contains(doc.Body, "sk-should-never-leak") {
			t.Errorf("concept %q leaked gitignored secret content: %q", id, doc.Body)
		}
	}
	if _, ok := concepts["codebase/main"]; !ok {
		t.Error("missing codebase/main concept (the non-ignored file should still be harvested)")
	}
}

// TestCodebaseExtractorSkipsGitignoredDirectories covers a whole
// gitignored directory (e.g. an agent's scratch/worktree directory), not
// just a single ignored file.
func TestCodebaseExtractorSkipsGitignoredDirectories(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	if err := os.WriteFile(filepath.Join(proj, ".gitignore"), []byte(".scratch/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	scratchDir := filepath.Join(proj, ".scratch", "nested")
	if err := os.MkdirAll(scratchDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scratchDir, "leftover.go"), []byte("package scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initRealGitRepo(t, proj, true)

	ext := CodebaseExtractor{ProjectRoot: proj}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	for id := range concepts {
		if strings.Contains(id, "scratch") {
			t.Errorf("concept %q from a gitignored directory should not have been harvested", id)
		}
	}
	if _, ok := concepts["codebase/main"]; !ok {
		t.Error("missing codebase/main concept")
	}
}

// TestCodebaseExtractorGitScopesToSubtree covers harvesting a subtree
// (-src ./internal, a documented, supported usage) of a larger git repo:
// concept IDs must stay relative to the harvested subtree, matching the
// non-git walk's existing behavior, not the repo root.
func TestCodebaseExtractorGitScopesToSubtree(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	if err := os.WriteFile(filepath.Join(proj, "root.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(proj, "internal")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "lib.go"), []byte("package internal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	initRealGitRepo(t, proj, true)

	ext := CodebaseExtractor{ProjectRoot: sub}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	if _, ok := concepts["codebase/lib"]; !ok {
		t.Errorf("missing codebase/lib concept, got %v", keysOf(concepts))
	}
	if _, ok := concepts["codebase/root"]; ok {
		t.Error("codebase/root should not appear when harvesting the internal/ subtree only")
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
	markAsRepoRoot(t, proj)
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

// TestCodebaseExtractorSkipsOverviewForSubtreeHarvest guards against
// harvesting a subtree (e.g. `-src ./src`, a documented, supported usage)
// silently overwriting the whole-project architecture/overview with a
// narrow one scoped to that subtree.
func TestCodebaseExtractorSkipsOverviewForSubtreeHarvest(t *testing.T) {
	t.Parallel()

	proj := filepath.Join(t.TempDir(), "my-project")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	// Deliberately no markAsRepoRoot(t, proj): this simulates harvesting a
	// subtree of a larger project, not the project root itself.
	if err := os.WriteFile(filepath.Join(proj, "main.py"), []byte("def hello(): pass\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := CodebaseExtractor{ProjectRoot: proj}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	if _, ok := concepts["architecture/overview"]; ok {
		t.Error("architecture/overview should not be produced for a subtree (non-repo-root) harvest")
	}
	if _, ok := concepts["codebase/main"]; !ok {
		t.Error("missing codebase/main concept")
	}
}

// TestCodebaseExtractorExportBundlePrunesStaleModules covers the scenario
// that motivated this fix: re-harvesting the same repo root after files
// were renamed/deleted (e.g. a language migration) must not leave the old
// concept files behind.
func TestCodebaseExtractorExportBundlePrunesStaleModules(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	markAsRepoRoot(t, proj)
	if err := os.WriteFile(filepath.Join(proj, "a.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	ext := CodebaseExtractor{ProjectRoot: proj}
	if _, err := ext.ExportBundle(out); err != nil {
		t.Fatalf("ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "a.md")); err != nil {
		t.Fatalf("codebase/a.md not written: %v", err)
	}

	if err := os.Remove(filepath.Join(proj, "a.go")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "b.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ext.ExportBundle(out); err != nil {
		t.Fatalf("second ExportBundle() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "codebase", "a.md")); !os.IsNotExist(err) {
		t.Errorf("codebase/a.md should have been pruned after a.go was removed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "b.md")); err != nil {
		t.Errorf("codebase/b.md not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "architecture", "overview.md")); err != nil {
		t.Errorf("architecture/overview.md should survive a repo-root re-harvest: %v", err)
	}
}

// TestCodebaseExtractorExportBundleSubtreeDoesNotClobberOverview covers the
// other half of the fix: a later subtree harvest into the same bundle must
// not delete or replace the project-wide overview a prior repo-root harvest
// produced.
func TestCodebaseExtractorExportBundleSubtreeDoesNotClobberOverview(t *testing.T) {
	t.Parallel()

	proj := t.TempDir()
	markAsRepoRoot(t, proj)
	if err := os.WriteFile(filepath.Join(proj, "README.md"), []byte("# Root Project\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "app.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	rootExt := CodebaseExtractor{ProjectRoot: proj}
	if _, err := rootExt.ExportBundle(out); err != nil {
		t.Fatalf("root ExportBundle() error = %v", err)
	}
	before, err := os.ReadFile(filepath.Join(out, "architecture", "overview.md"))
	if err != nil {
		t.Fatalf("architecture/overview.md not written by root harvest: %v", err)
	}

	sub := filepath.Join(proj, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "lib.go"), []byte("package sub\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	subExt := CodebaseExtractor{ProjectRoot: sub}
	if _, err := subExt.ExportBundle(out); err != nil {
		t.Fatalf("subtree ExportBundle() error = %v", err)
	}

	after, err := os.ReadFile(filepath.Join(out, "architecture", "overview.md"))
	if err != nil {
		t.Fatalf("architecture/overview.md should still exist after subtree harvest: %v", err)
	}
	if string(before) != string(after) {
		t.Errorf("architecture/overview.md changed after subtree harvest:\nbefore=%s\nafter=%s", before, after)
	}
	if _, err := os.Stat(filepath.Join(out, "codebase", "lib.md")); err != nil {
		t.Errorf("codebase/lib.md not written by subtree harvest: %v", err)
	}
	// Subtree concept IDs are root-relative with nothing to namespace them
	// apart from the rest of the tree, so a subtree harvest must never
	// prune codebase/ at all: it only ever sees its own slice and would
	// otherwise delete concepts belonging to the rest of the project.
	if _, err := os.Stat(filepath.Join(out, "codebase", "app.md")); err != nil {
		t.Errorf("codebase/app.md from the earlier root harvest should survive a subtree harvest: %v", err)
	}
}
