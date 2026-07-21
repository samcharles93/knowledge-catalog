---
description: Source module okf/paths.go (69 lines).
resource: okf/paths.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: paths.go
type: Module
---

# Module paths.go

**Path**: `okf/paths.go`  
**Lines**: 69

## Snippet Preview

```
package okf

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var segmentRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.\-]*$`)

func validateSegment(seg string) error {
	if !segmentRE.MatchString(seg) {
		return fmt.Errorf("invalid concept id segment: %q", seg)
	}
	return nil
}

// ParseConceptID splits a slash-separated concept identifier into its
// segments, validating each segment and rejecting empty or traversal input.
func ParseConceptID(id string) ([]string, error) {
	var parts []string
	for p := range strings.SplitSeq(id, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty concept id: %q", id)
	}
```
