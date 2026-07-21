---
description: Source module okf/extract_sql_test.go (112 lines).
resource: okf/extract_sql_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:53:05Z"
title: extract_sql_test.go
type: Module
---

# Module extract_sql_test.go

**Path**: `okf/extract_sql_test.go`  
**Lines**: 112

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLExtractorParsesCreateTable(t *testing.T) {
	t.Parallel()

	sqlFile := filepath.Join(t.TempDir(), "schema.sql")
	sql := "CREATE TABLE users (\n" +
		"  id INT PRIMARY KEY,\n" +
		"  email VARCHAR(255)\n" +
		");"
	if err := os.WriteFile(sqlFile, []byte(sql), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := SQLExtractor{SQLPath: sqlFile}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	doc, ok := concepts["database/users"]
	if !ok {
		t.Fatal("missing database/users concept")
	}
```
