---
description: Source module okf/extract_codebase_test.go (91 lines).
resource: okf/extract_codebase_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: extract_codebase_test.go
type: Module
---

# Module extract_codebase_test.go

**Path**: `okf/extract_codebase_test.go`  
**Lines**: 91

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodebaseExtractorProducesOverviewAndModules(t *testing.T) {
	t.Parallel()

	proj := filepath.Join(t.TempDir(), "my-project")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
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
```
