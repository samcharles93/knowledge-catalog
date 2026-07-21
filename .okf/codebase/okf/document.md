---
description: Source module okf/document.go (180 lines).
resource: okf/document.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: document.go
type: Module
---

# Module document.go

**Path**: `okf/document.go`  
**Lines**: 180

## Snippet Preview

```
package okf

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const frontmatterDelim = "---"

var (
	requiredFrontmatterKeys    = []string{"type"}
	recommendedFrontmatterKeys = []string{"title", "description"}
)

// Document is a parsed OKF markdown file: YAML frontmatter plus body text.
type Document struct {
	Frontmatter map[string]any
	Body        string
}

// splitLines mimics Python's str.splitlines(): it splits on line
// boundaries without producing a trailing empty element for a final
// newline, and normalizes CRLF/CR to LF first.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
```
