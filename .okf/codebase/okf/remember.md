---
description: Source module okf/remember.go (111 lines).
resource: okf/remember.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: remember.go
type: Module
---

# Module remember.go

**Path**: `okf/remember.go`  
**Lines**: 111

## Snippet Preview

```
package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var rememberDirs = map[string]string{
	"Architecture": "architecture",
	"Service":      "services",
	"Rule":         "rules",
	"Runbook":      "runbooks",
	"Concept":      "concepts",
}

// rememberReservedTypes are owned by a harvest Extractor's pruning: a
// manually-written concept placed there would be silently deleted by the
// next matching harvest run.
var rememberReservedTypes = map[string]bool{
	"Codebase":  true, // owned by CodebaseExtractor, codebase/
	"API":       true, // owned by OpenAPIExtractor, api/
	"Table":     true, // owned by SQLExtractor, database/
	"Reference": true, // owned by WebExtractor, references/
}

// RememberInput captures a single deliberate, free-form memory (a coding
// rule, session insight, runbook step, etc.) to be written as one OKF
```
