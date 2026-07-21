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
	t.Parallel()

	children := []ChildSummary{
		{Title: "one"}, {Title: "two"}, {Title: "three"},
		{Title: "four"}, {Title: "five"}, {Title: "six"},
	}
	got := SynthesizeDescription("services", children)
	if !strings.HasPrefix(got, "Contains 6 concepts (one, two, three, four, five and 1 more).") {
		t.Errorf("SynthesizeDescription() = %q", got)
	}
}
