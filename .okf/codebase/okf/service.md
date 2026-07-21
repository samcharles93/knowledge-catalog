---
description: Source module okf/service.go (213 lines).
resource: okf/service.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: service.go
type: Module
---

# Module service.go

**Path**: `okf/service.go`  
**Lines**: 213

## Snippet Preview

```
package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const launchdLabel = "com.samcharles93.okf-mcp"

const systemdUnitName = "okf-mcp.service"

// ServiceConfig describes an `okf mcp` invocation to register as a
// background OS service.
type ServiceConfig struct {
	BundleRoot string
	Addr       string
	BinPath    string
}

// SystemdUnit renders a systemd user unit file that runs `okf mcp` as a
// background service on Linux.
func SystemdUnit(cfg ServiceConfig) string {
	return fmt.Sprintf(`[Unit]
Description=OKF Knowledge Bundle MCP server
After=network.target

[Service]
Type=simple
```
