---
description: Source module okf/fetch.go (155 lines).
resource: okf/fetch.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: fetch.go
type: Module
---

# Module fetch.go

**Path**: `okf/fetch.go`  
**Lines**: 155

## Snippet Preview

```
package okf

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/net/html/charset"
)

const (
	fetchUserAgent   = "okf-reference-agent/0.1 (+https://github.com/samcharles93/knowledge-catalog)"
	fetchTimeout     = 10 * time.Second
	maxMarkdownBytes = 40 * 1024
)

var (
	titleRE = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	hrefRE  = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'\s]+)["']`)
)

// Page is a fetched web page reduced to its title, markdown body, and the
// absolute, fragment-stripped links found on it.
```
