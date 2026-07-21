---
description: Source module okf/extract.go (141 lines).
resource: okf/extract.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: extract.go
type: Module
---

# Module extract.go

**Path**: `okf/extract.go`  
**Lines**: 141

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
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
//
// prunePrefixes lists the concept ID prefixes this invocation is
// authoritative for (e.g. "codebase" for a codebase harvest): any existing
// bundle file under one of those prefixes that is absent from the fresh
// concepts map is stale (its source was renamed/removed since the last
// harvest) and gets deleted, along with any directory left empty by that
// deletion. Pass nil to skip pruning entirely, for extractors whose harvests
// are additive across runs (e.g. WebExtractor, built up via repeated --url
// flags) rather than a full snapshot of a single source.
```
