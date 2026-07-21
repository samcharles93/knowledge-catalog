---
description: Source module okf/mcp_test.go (136 lines).
resource: okf/mcp_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: mcp_test.go
type: Module
---

# Module mcp_test.go

**Path**: `okf/mcp_test.go`  
**Lines**: 136

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPListConceptsFormatsEachEntry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	doc := Document{
		Frontmatter: map[string]any{"type": "Service", "title": "Users", "description": "User operations"},
		Body:        "body",
	}
	if err := os.WriteFile(filepath.Join(root, "users.md"), []byte(doc.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.md"), []byte("# Index\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	out, err := MCPListConcepts(root)
	if err != nil {
		t.Fatalf("MCPListConcepts() error = %v", err)
	}
	if !strings.Contains(out, "`users`") || !strings.Contains(out, "[Service]") || !strings.Contains(out, "Users") {
		t.Errorf("MCPListConcepts() = %q", out)
```
