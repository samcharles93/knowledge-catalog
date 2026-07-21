package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCreatesStarterBundle(t *testing.T) {
	t.Parallel()

	bundle := filepath.Join(t.TempDir(), ".okf")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--path", bundle}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(init) = %d, stderr = %q", code, stderr.String())
	}
	for _, path := range []string{
		"config.yaml",
		"index.md",
		filepath.Join("architecture", "system-overview.md"),
		"services",
		"rules",
	} {
		if _, err := os.Stat(filepath.Join(bundle, path)); err != nil {
			t.Errorf("starter path %q: %v", path, err)
		}
	}
	if !strings.Contains(stderr.String(), "Initialized OKF Knowledge Bundle") {
		t.Errorf("stderr = %q", stderr.String())
	}
}

func TestRunValidateUsesExitStatusForValidity(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundle, "invalid.md"), []byte("---\ntitle: Missing Type\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"validate", "--bundle", bundle}, &stdout, &stderr); code != 1 {
		t.Fatalf("run(validate) = %d, want 1; stdout = %q; stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "Valid: false") || !strings.Contains(stdout.String(), "Missing required frontmatter keys: type") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunContextWritesSummary(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundle, "index.md"), []byte("# Services\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run([]string{"context", "--bundle", bundle}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(context) = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "# Project Knowledge Base (OKF)") || !strings.Contains(stdout.String(), "# Services") {
		t.Errorf("stdout = %q", stdout.String())
	}
}

func TestRunHarvestCodebaseWritesBundle(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), ".okf")
	var stdout, stderr bytes.Buffer
	code := run([]string{"harvest", "--type", "codebase", "--src", src, "--out", out}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(harvest) = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Harvested") {
		t.Errorf("stderr = %q, want harvest summary", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(out, "index.md")); err != nil {
		t.Errorf("index.md not written: %v", err)
	}
}

func TestRunHarvestRejectsUnknownType(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := run([]string{"harvest", "--type", "bogus", "--out", t.TempDir()}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run(harvest bogus) = %d, want 2; stderr = %q", code, stderr.String())
	}
}

func TestRunVisualizeWritesHTML(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	if err := os.WriteFile(filepath.Join(bundle, "overview.md"), []byte("---\ntype: Concept\ntitle: Overview\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "viz.html")
	var stdout, stderr bytes.Buffer
	code := run([]string{"visualize", "--bundle", bundle, "--out", out}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(visualize) = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "1 concept") {
		t.Errorf("stderr = %q, want concept count", stderr.String())
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("viz.html not written: %v", err)
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := run([]string{"unknown"}, &stdout, &stderr); code != 2 {
		t.Fatalf("run(unknown) = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage") {
		t.Errorf("stderr = %q, want usage", stderr.String())
	}
}
