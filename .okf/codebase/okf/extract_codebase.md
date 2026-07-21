---
description: Source module okf/extract_codebase.go (159 lines).
resource: okf/extract_codebase.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: extract_codebase.go
type: Module
---

# Module extract_codebase.go

**Path**: `okf/extract_codebase.go`  
**Lines**: 159

## Snippet Preview

```
package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var codebaseIgnoredDirs = map[string]bool{
	".git": true, ".venv": true, "node_modules": true, "__pycache__": true,
	".pytest_cache": true, "dist": true, "build": true, ".okf": true,
	"target": true, "vendor": true,
}

var codebaseSourceExtensions = map[string]bool{
	".py": true, ".ts": true, ".js": true, ".go": true, ".rs": true,
	".java": true, ".cpp": true, ".h": true,
}

// CodebaseExtractor scans a local source tree and produces an architecture
// overview concept plus one Module concept per recognized source file.
type CodebaseExtractor struct {
	ProjectRoot string
}

var _ Extractor = CodebaseExtractor{}

```
