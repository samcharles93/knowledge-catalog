---
description: Source module okf/extract_codebase_test.go (231 lines).
resource: okf/extract_codebase_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: extract_codebase_test.go
type: Module
---

# Module extract_codebase_test.go

**Path**: `okf/extract_codebase_test.go`  
**Lines**: 231

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
```
