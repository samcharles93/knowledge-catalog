---
description: Source module okf/bundle_test.go (331 lines).
resource: okf/bundle_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: bundle_test.go
type: Module
---

# Module bundle_test.go

**Path**: `okf/bundle_test.go`  
**Lines**: 331

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"reflect"
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
```
