---
description: Source module okf/fetch_test.go (65 lines).
resource: okf/fetch_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: fetch_test.go
type: Module
---

# Module fetch_test.go

**Path**: `okf/fetch_test.go`  
**Lines**: 65

## Snippet Preview

```
package okf

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchPageExtractsTitleLinksAndMarkdown(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><head><title>  Hello   World </title></head>` +
			`<body><h1>Hi</h1><a href="/other">Other</a><a href="https://example.com/x#frag">Ext</a></body></html>`))
	}))
	defer srv.Close()

	page, err := FetchPage(srv.URL)
	if err != nil {
		t.Fatalf("FetchPage() error = %v", err)
	}
	if page.Title != "Hello World" {
		t.Errorf("Title = %q, want %q", page.Title, "Hello World")
	}
	if !strings.Contains(page.Markdown, "Hi") {
		t.Errorf("Markdown = %q, want to contain heading text", page.Markdown)
	}
	wantLink := srv.URL + "/other"
```
