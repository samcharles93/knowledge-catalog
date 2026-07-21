---
description: Source module okf/viz.go (251 lines).
resource: okf/viz.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: viz.go
type: Module
---

# Module viz.go

**Path**: `okf/viz.go`  
**Lines**: 251

## Snippet Preview

```
package okf

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed viewer/viz.html viewer/viz.css viewer/viz.js
var vizAssets embed.FS

var typePalette = map[string]string{
	"Architecture": "#8b5cf6",
	"Service":      "#3b82f6",
	"API":          "#06b6d4",
	"Database":     "#10b981",
	"Table":        "#10b981",
	"Module":       "#f59e0b",
	"Class":        "#f59e0b",
	"Function":     "#f59e0b",
	"Rule":         "#ef4444",
	"Runbook":      "#ec4899",
	"Reference":    "#64748b",
	"Concept":      "#64748b",
}
```
