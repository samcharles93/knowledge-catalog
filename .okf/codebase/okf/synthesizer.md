---
description: Source module okf/synthesizer.go (52 lines).
resource: okf/synthesizer.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: synthesizer.go
type: Module
---

# Module synthesizer.go

**Path**: `okf/synthesizer.go`  
**Lines**: 52

## Snippet Preview

```
package okf

import (
	"fmt"
	"strings"
)

// ChildSummary is a (title, description) pair for a directory entry, used
// as input to description synthesis when generating index.md files.
type ChildSummary struct {
	Title       string
	Description string
}

// SynthesizeDescription produces a deterministic, vendor-neutral summary of
// a directory's contents from its child concepts. relPath is accepted for
// forward compatibility with model-backed synthesis strategies but is not
// currently used.
func SynthesizeDescription(relPath string, children []ChildSummary) string {
	if len(children) == 0 {
		return ""
	}

	for _, c := range children {
		if c.Description != "" {
			return fmt.Sprintf("Directory containing %d items, including: %s.", len(children), children[0].Title)
		}
	}
	return fallbackDescription(children)
}
```
