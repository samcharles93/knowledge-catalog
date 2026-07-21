---
description: Source module okf/paths_test.go (68 lines).
resource: okf/paths_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: paths_test.go
type: Module
---

# Module paths_test.go

**Path**: `okf/paths_test.go`  
**Lines**: 68

## Snippet Preview

```
package okf

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseConceptID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{name: "nested", input: "services/user-service", want: []string{"services", "user-service"}},
		{name: "surrounding slashes", input: "/services/users/", want: []string{"services", "users"}},
		{name: "underscore and dot", input: "api/v1_users.legacy", want: []string{"api", "v1_users.legacy"}},
		{name: "empty", input: "///", wantErr: true},
		{name: "traversal", input: "../outside", wantErr: true},
		{name: "space", input: "services/user service", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseConceptID(tt.input)
			if (err != nil) != tt.wantErr {
```
