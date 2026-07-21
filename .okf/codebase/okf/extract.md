---
description: Source module okf/extract.go (42 lines).
resource: okf/extract.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: extract.go
type: Module
---

# Module extract.go

**Path**: `okf/extract.go`  
**Lines**: 42

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
)

// Extractor turns some external, vendor-neutral source into OKF concept
// documents and can export them directly into a bundle.
type Extractor interface {
	// ExtractConcepts returns the concept documents keyed by concept ID.
	ExtractConcepts() (map[string]Document, error)
	// ExportBundle writes the extracted concepts into bundleRoot and
	// regenerates indexes, returning the number of concepts written.
	ExportBundle(bundleRoot string) (int, error)
}

// exportBundle writes the concepts produced by extract into bundleRoot (one
// {concept_id}.md file per concept) and regenerates the bundle's indexes.
// Shared by every Extractor's ExportBundle method.
func exportBundle(bundleRoot string, extract func() (map[string]Document, error)) (int, error) {
	concepts, err := extract()
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(bundleRoot, 0o755); err != nil {
		return 0, err
	}
	for id, doc := range concepts {
		path := filepath.Join(bundleRoot, id+".md")
```
