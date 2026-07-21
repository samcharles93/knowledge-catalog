---
description: Source module okf/synthesizer_test.go (45 lines).
resource: okf/synthesizer_test.go
tags:
    - go
    - source
timestamp: "2026-07-21T17:36:27Z"
title: synthesizer_test.go
type: Module
---

# Module synthesizer_test.go

**Path**: `okf/synthesizer_test.go`  
**Lines**: 45

## Snippet Preview

```
package okf

import (
	"strings"
	"testing"
)

func TestSynthesizeDescriptionEmpty(t *testing.T) {
	t.Parallel()

	if got := SynthesizeDescription("services", nil); got != "" {
		t.Errorf("SynthesizeDescription(nil) = %q, want empty", got)
	}
}

func TestSynthesizeDescriptionUsesFirstChildWithDescription(t *testing.T) {
	t.Parallel()

	children := []ChildSummary{
		{Title: "Auth Service", Description: "Handles authentication."},
		{Title: "Billing Service", Description: "Handles billing."},
	}
	got := SynthesizeDescription("services", children)
	want := "Directory containing 2 items, including: Auth Service."
	if got != want {
		t.Errorf("SynthesizeDescription() = %q, want %q", got, want)
	}
}

func TestSynthesizeDescriptionFallsBackWithoutDescriptions(t *testing.T) {
```
