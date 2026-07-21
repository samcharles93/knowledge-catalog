---
description: Source module okf/mcp.go (242 lines).
resource: okf/mcp.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: mcp.go
type: Module
---

# Module mcp.go

**Path**: `okf/mcp.go`  
**Lines**: 242

## Snippet Preview

```
package okf

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPListConcepts renders a one-line-per-concept listing (id, type, title,
// description) for every concept in the bundle, excluding index/log files.
func MCPListConcepts(bundleRoot string) (string, error) {
	files, err := walkConceptFiles(bundleRoot)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "No concepts found.", nil
	}

	lines := make([]string, 0, len(files))
	for _, f := range files {
		if f.Err != nil {
			lines = append(lines, fmt.Sprintf("- `%s` (Parse Error)", f.ID))
			continue
		}
		fm := f.Doc.Frontmatter
		typ := stringFrontmatterOr(fm["type"], "Concept")
```
