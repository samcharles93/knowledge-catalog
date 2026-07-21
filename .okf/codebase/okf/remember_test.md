---
description: Source module okf/remember_test.go (165 lines).
resource: okf/remember_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T18:15:48Z"
title: remember_test.go
type: Module
---

# Module remember_test.go

**Path**: `okf/remember_test.go`  
**Lines**: 165

## Snippet Preview

```
package okf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRememberWritesValidatedConceptUnderTypeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	conceptID, err := Remember(root, RememberInput{
		Type:  "Rule",
		Title: "Always run golangci-lint before committing Go changes",
		Body:  "Run golangci-lint before every Go commit.",
	})
	if err != nil {
		t.Fatalf("Remember() error = %v", err)
	}
	if conceptID != "rules/Always_run_golangci_lint_before_committing_Go_changes" {
		t.Errorf("conceptID = %q, want rules/Always_run_golangci_lint_before_committing_Go_changes", conceptID)
	}

	path, err := ConceptPath(root, []string{"rules", "Always_run_golangci_lint_before_committing_Go_changes"})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
```
