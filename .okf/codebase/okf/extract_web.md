---
description: Source module okf/extract_web.go (77 lines).
resource: okf/extract_web.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: extract_web.go
type: Module
---

# Module extract_web.go

**Path**: `okf/extract_web.go`  
**Lines**: 77

## Snippet Preview

```
package okf

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// WebExtractor fetches a set of URLs and converts each into a Reference
// concept. Fetch defaults to FetchPage; tests may override it to avoid
// real network access.
type WebExtractor struct {
	URLs  []string
	Fetch func(url string) (Page, error)
}

var _ Extractor = WebExtractor{}

func (e WebExtractor) ExtractConcepts() (map[string]Document, error) {
	fetch := e.Fetch
	if fetch == nil {
		fetch = FetchPage
	}
	now := time.Now().UTC().Format(time.RFC3339)
	concepts := map[string]Document{}

	for _, u := range e.URLs {
		page, err := fetch(u)
```
