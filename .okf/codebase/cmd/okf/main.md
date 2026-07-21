---
description: Source module cmd/okf/main.go (478 lines).
resource: cmd/okf/main.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: main.go
type: Module
---

# Module main.go

**Path**: `cmd/okf/main.go`  
**Lines**: 478

## Snippet Preview

```
// Command okf is the Open Knowledge Format toolkit CLI.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samcharles93/knowledge-catalog/okf"
)

// version, commit, and date are set via -ldflags at build time (see
// .goreleaser.yaml); they default to "dev"/"none"/"unknown" for `go build`
// and `go run` during development.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

```
