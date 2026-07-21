---
description: Source module okf/context_test.go (135 lines).
resource: okf/context_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: context_test.go
type: Module
---

# Module context_test.go

**Path**: `okf/context_test.go`  
**Lines**: 135

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConceptContextResolvesByAliasWhenIDMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{
			"type": "Service", "title": "User Service", "description": "Handles users.",
			"aliases": []string{"UserMgr", "AuthService"},
		},
		Body: "# User Service\n",
	}
	if err := os.WriteFile(filepath.Join(root, "user-service.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := ConceptContext(root, "UserMgr")
	if err != nil {
		t.Fatalf("ConceptContext() error = %v", err)
	}
	if !strings.Contains(out, "# Concept: User Service") {
		t.Errorf("ConceptContext() = %q, want alias-resolved concept", out)
```
