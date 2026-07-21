---
description: Source module okf/extract_web_test.go (84 lines).
resource: okf/extract_web_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: extract_web_test.go
type: Module
---

# Module extract_web_test.go

**Path**: `okf/extract_web_test.go`  
**Lines**: 84

## Snippet Preview

```
package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestWebExtractorBuildsReferenceConcepts(t *testing.T) {
	t.Parallel()

	ext := WebExtractor{
		URLs: []string{"https://example.com/docs/guide"},
		Fetch: func(url string) (Page, error) {
			return Page{URL: url, Title: "Guide", Markdown: "# Guide\n\nContent."}, nil
		},
	}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	doc, ok := concepts["references/example_com_docs_guide"]
	if !ok {
		t.Fatalf("missing reference concept, got %v", keysOf(concepts))
	}
	if doc.Frontmatter["type"] != "Reference" {
		t.Errorf("type = %v, want Reference", doc.Frontmatter["type"])
	}
	if doc.Frontmatter["resource"] != "https://example.com/docs/guide" {
```
