---
description: Source module okf/viz_test.go (109 lines).
resource: okf/viz_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: viz_test.go
type: Module
---

# Module viz_test.go

**Path**: `okf/viz_test.go`  
**Lines**: 109

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateVisualizationWritesGraphHTML(t *testing.T) {
	t.Parallel()

	bundle := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{"type": "Architecture", "title": "Overview", "description": "Overview concept"},
		Body:        "# Overview\nContent.\n",
	}
	if err := os.WriteFile(filepath.Join(bundle, "overview.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "viz.html")
	stats, err := GenerateVisualization(bundle, out, "")
	if err != nil {
		t.Fatalf("GenerateVisualization() error = %v", err)
	}
	if stats.Concepts != 1 {
		t.Errorf("stats.Concepts = %d, want 1", stats.Concepts)
	}
	if stats.Edges != 0 {
```
