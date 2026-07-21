---
description: Source module cmd/okf/main_test.go (299 lines).
resource: cmd/okf/main_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: main_test.go
type: Module
---

# Module main_test.go

**Path**: `cmd/okf/main_test.go`  
**Lines**: 299

## Snippet Preview

```
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCreatesStarterBundle(t *testing.T) {
	t.Parallel()

	bundle := filepath.Join(t.TempDir(), ".okf")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--path", bundle}, &stdout, &stderr); code != 0 {
		t.Fatalf("run(init) = %d, stderr = %q", code, stderr.String())
	}
	for _, path := range []string{
		"config.yaml",
		"index.md",
		filepath.Join("architecture", "system-overview.md"),
		"services",
		"rules",
	} {
		if _, err := os.Stat(filepath.Join(bundle, path)); err != nil {
			t.Errorf("starter path %q: %v", path, err)
		}
	}
	if !strings.Contains(stderr.String(), "Initialized OKF Knowledge Bundle") {
```
