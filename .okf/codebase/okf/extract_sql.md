---
description: Source module okf/extract_sql.go (92 lines).
resource: okf/extract_sql.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: extract_sql.go
type: Module
---

# Module extract_sql.go

**Path**: `okf/extract_sql.go`  
**Lines**: 92

## Snippet Preview

```
package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SQLExtractor parses CREATE TABLE statements out of a SQL DDL file into
// one Table concept per table.
type SQLExtractor struct {
	SQLPath string
}

var _ Extractor = SQLExtractor{}

var sqlCreateTableRE = regexp.MustCompile(
	"(?is)CREATE\\s+TABLE\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?([`\"\\w.]+)\\s*\\((.*?)\\);",
)

var sqlSkipLinePrefixes = []string{"--", "PRIMARY", "CONSTRAINT", "FOREIGN", "KEY", "INDEX", ")"}

func (e SQLExtractor) ExtractConcepts() (map[string]Document, error) {
	data, err := os.ReadFile(e.SQLPath)
	if err != nil {
		return nil, err
	}
```
