---
description: Source module okf/extract_test.go (113 lines).
resource: okf/extract_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: extract_test.go
type: Module
---

# Module extract_test.go

**Path**: `okf/extract_test.go`  
**Lines**: 113

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"testing"
)

func newDoc(title string) Document {
	return Document{
		Frontmatter: map[string]any{"type": "Module", "title": title},
		Body:        "# " + title + "\n",
	}
}

// TestExportBundlePrunesStaleFilesWithinOwnedPrefixes is a direct unit test
// of the shared exportBundle helper (independent of any specific extractor):
// re-exporting with a different concept set must delete stale files under
// the caller-declared owned prefixes, but must leave everything outside
// those prefixes untouched.
func TestExportBundlePrunesStaleFilesWithinOwnedPrefixes(t *testing.T) {
	t.Parallel()

	out := t.TempDir()

	first := map[string]Document{
		"codebase/pkg/sub/old": newDoc("old"),
		"api/keep":             newDoc("keep"), // outside the owned prefix
	}
	if _, err := exportBundle(out, func() (map[string]Document, error) { return first, nil }, []string{"codebase"}); err != nil {
```
