---
description: Source module okf/document_test.go (225 lines).
resource: okf/document_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: document_test.go
type: Module
---

# Module document_test.go

**Path**: `okf/document_test.go`  
**Lines**: 225

## Snippet Preview

```
package okf

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDocumentRoundTrip(t *testing.T) {
	t.Parallel()

	source := "---\ntype: Service\ntitle: User Service\ndescription: User authentication and management.\ntags: [auth, users]\ntimestamp: 2026-07-22T00:00:00Z\n---\n\n# User Service\n\nService body details.\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if got := doc.Frontmatter["type"]; got != "Service" {
		t.Errorf("type = %v, want Service", got)
	}
	if got := doc.Frontmatter["tags"]; !reflect.DeepEqual(got, []string{"auth", "users"}) {
		t.Errorf("tags = %#v", got)
	}
	if !strings.HasPrefix(doc.Body, "# User Service") {
		t.Errorf("body = %q", doc.Body)
	}

	reparsed, err := ParseDocument(doc.String())
	if err != nil {
		t.Fatalf("ParseDocument(String()) error = %v", err)
	}
```
