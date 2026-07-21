---
description: Source module okf/context.go (126 lines).
resource: okf/context.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: context.go
type: Module
---

# Module context.go

**Path**: `okf/context.go`  
**Lines**: 126

## Snippet Preview

```
package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// SummaryContext renders a progressive-disclosure prompt snippet from the
// bundle's root index.md, for use as agent context.
func SummaryContext(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "No OKF bundle found.", nil
	}

	indexContent := "Root index.md not found."
	if data, err := os.ReadFile(filepath.Join(root, indexFileName)); err == nil {
		indexContent = string(data)
	}

	lines := []string{
		"# Project Knowledge Base (OKF)",
		fmt.Sprintf("**Bundle Location**: `%s`", filepath.Base(root)),
		"",
		"## Progressive Disclosure Index",
		"",
		indexContent,
```
