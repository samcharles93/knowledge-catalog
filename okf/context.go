package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SummaryContext renders a progressive-disclosure prompt snippet from the
// bundle's root index.md, for use as agent context.
func SummaryContext(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return "No OKF bundle found.", nil
	}

	indexContent := "Root index.md not found."
	if data, err := os.ReadFile(filepath.Join(root, indexFileName)); err == nil {
		indexContent = string(data)
	}

	lines := []string{
		"# Project Knowledge Base (OKF)",
		fmt.Sprintf("**Bundle Location**: `%s`", filepath.Base(root)),
		"",
		"## Progressive Disclosure Index",
		"",
		indexContent,
		"",
		"---",
		"*(Agent Note: Use `okf get <concept_id>` or open specific `.md` files under `.okf/` for detailed schemas and architecture details.)*",
	}
	return strings.Join(lines, "\n"), nil
}

// ConceptContext renders the full prompt-ready markdown for a single
// concept, including a metadata header derived from its frontmatter.
func ConceptContext(root string, conceptID string) (string, error) {
	path := filepath.Join(root, strings.TrimLeft(conceptID, "/")+".md")
	if _, err := os.Stat(path); err != nil {
		return fmt.Sprintf("Concept '%s' not found in bundle.", conceptID), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	doc, err := ParseDocument(string(data))
	if err != nil {
		return "", err
	}

	fm := doc.Frontmatter
	lines := []string{
		fmt.Sprintf("# Concept: %s", stringFrontmatterOr(fm["title"], conceptID)),
		fmt.Sprintf("- **Type**: %s", stringFrontmatterOr(fm["type"], "Unknown")),
		fmt.Sprintf("- **Description**: %s", stringFrontmatterOr(fm["description"], "N/A")),
	}
	if resource := stringFrontmatterOr(fm["resource"], ""); resource != "" {
		lines = append(lines, fmt.Sprintf("- **Resource**: `%s`", resource))
	}
	if tags := frontmatterStringSlice(fm["tags"]); len(tags) > 0 {
		lines = append(lines, fmt.Sprintf("- **Tags**: %s", strings.Join(tags, ", ")))
	}

	return strings.Join(lines, "\n") + "\n\n" + doc.Body, nil
}

// frontmatterStringSlice extracts a []string from a frontmatter value,
// tolerating both the []string produced by normalizeYAMLValue and a raw
// []any with mixed element types.
func frontmatterStringSlice(v any) []string {
	switch val := v.(type) {
	case []string:
		return val
	case []any:
		out := make([]string, 0, len(val))
		for _, e := range val {
			if s, ok := e.(string); ok {
				out = append(out, s)
			} else {
				out = append(out, fmt.Sprint(e))
			}
		}
		return out
	default:
		return nil
	}
}
