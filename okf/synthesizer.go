package okf

import (
	"fmt"
	"strings"
)

// ChildSummary is a (title, description) pair for a directory entry, used
// as input to description synthesis when generating index.md files.
type ChildSummary struct {
	Title       string
	Description string
}

// SynthesizeDescription produces a deterministic, vendor-neutral summary of
// a directory's contents from its child concepts. relPath is accepted for
// forward compatibility with model-backed synthesis strategies but is not
// currently used.
func SynthesizeDescription(relPath string, children []ChildSummary) string {
	if len(children) == 0 {
		return ""
	}

	for _, c := range children {
		if c.Description != "" {
			return fmt.Sprintf("Directory containing %d items, including: %s.", len(children), children[0].Title)
		}
	}
	return fallbackDescription(children)
}

func fallbackDescription(children []ChildSummary) string {
	n := len(children)
	limit := min(n, 5)

	var titles []string
	for _, c := range children[:limit] {
		if c.Title != "" {
			titles = append(titles, c.Title)
		}
	}
	titleStr := strings.Join(titles, ", ")
	if titleStr == "" {
		titleStr = "concepts"
	}

	more := ""
	if n > 5 {
		more = fmt.Sprintf(" and %d more", n-5)
	}
	return fmt.Sprintf("Contains %d concepts (%s%s).", n, titleStr, more)
}
