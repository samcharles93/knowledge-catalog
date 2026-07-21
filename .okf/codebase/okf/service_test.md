---
description: Source module okf/service_test.go (325 lines).
resource: okf/service_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: service_test.go
type: Module
---

# Module service_test.go

**Path**: `okf/service_test.go`  
**Lines**: 325

## Snippet Preview

```
package okf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testServiceConfig() ServiceConfig {
	return ServiceConfig{
		BundleRoot: "/home/sam/project/.okf",
		Addr:       ":8080",
		BinPath:    "/usr/local/bin/okf",
	}
}

func TestSystemdUnitRendersExecStartWithBundleAndAddr(t *testing.T) {
	t.Parallel()

	unit := SystemdUnit(testServiceConfig())
	if !strings.Contains(unit, "[Unit]") || !strings.Contains(unit, "[Service]") || !strings.Contains(unit, "[Install]") {
		t.Fatalf("SystemdUnit() missing standard sections:\n%s", unit)
	}
	wantExec := "ExecStart=/usr/local/bin/okf mcp --bundle /home/sam/project/.okf --addr :8080"
	if !strings.Contains(unit, wantExec) {
		t.Errorf("SystemdUnit() = %q, want ExecStart %q", unit, wantExec)
	}
	if !strings.Contains(unit, "Restart=on-failure") {
```
