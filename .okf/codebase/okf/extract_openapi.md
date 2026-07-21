---
description: Source module okf/extract_openapi.go (109 lines).
resource: okf/extract_openapi.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: extract_openapi.go
type: Module
---

# Module extract_openapi.go

**Path**: `okf/extract_openapi.go`  
**Lines**: 109

## Snippet Preview

```
package okf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// OpenAPIExtractor parses an OpenAPI 3.0 / Swagger spec (JSON or YAML) into
// one API concept per operation.
type OpenAPIExtractor struct {
	SpecPath string
}

var _ Extractor = OpenAPIExtractor{}

var openAPIMethods = []string{"get", "post", "put", "delete", "patch"}

var openAPIOpIDCleanRE = regexp.MustCompile(`[^A-Za-z0-9_-]`)

func (e OpenAPIExtractor) ExtractConcepts() (map[string]Document, error) {
	data, err := os.ReadFile(e.SpecPath)
	if err != nil {
```
