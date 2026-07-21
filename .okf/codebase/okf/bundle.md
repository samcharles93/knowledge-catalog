---
description: Source module okf/bundle.go (481 lines).
resource: okf/bundle.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: bundle.go
type: Module
---

# Module bundle.go

**Path**: `okf/bundle.go`  
**Lines**: 481

## Snippet Preview

```
package okf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const indexFileName = "index.md"

var reservedConceptFileNames = map[string]bool{
	"index.md": true,
	"log.md":   true,
}

var linkRE = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// ConceptFile is one discovered concept document: its ID, on-disk path,
// raw bytes, and parsed contents. Err is set instead of Doc/Raw when the
// file could not be read (ParseDocument itself is self-healing and never
// fails).
type ConceptFile struct {
	ID   string
	Path string
	Raw  []byte
```
