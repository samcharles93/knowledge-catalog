---
description: Source module okf/extract_openapi_test.go (84 lines).
resource: okf/extract_openapi_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: extract_openapi_test.go
type: Module
---

# Module extract_openapi_test.go

**Path**: `okf/extract_openapi_test.go`  
**Lines**: 84

## Snippet Preview

```
package okf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPIExtractorExtractsOperations(t *testing.T) {
	t.Parallel()

	spec := filepath.Join(t.TempDir(), "openapi.json")
	body, err := json.Marshal(map[string]any{
		"openapi": "3.0.0",
		"info":    map[string]any{"title": "Test API", "version": "1.0"},
		"paths": map[string]any{
			"/users": map[string]any{
				"get": map[string]any{
					"summary":     "List users",
					"operationId": "listUsers",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(spec, body, 0o644); err != nil {
		t.Fatal(err)
```
