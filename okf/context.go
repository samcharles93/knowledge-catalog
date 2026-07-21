package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
// concept, including a metadata header derived from its frontmatter. If
// conceptID doesn't resolve to a file directly, it falls back to an exact,
// case-sensitive match against a concept's `aliases` frontmatter list
// (SPEC.md §3.1: "aliases ... Alternative lookup names").
func ConceptContext(root string, conceptID string) (string, error) {
	concept, found, err := resolveConcept(root, conceptID)
	if err != nil {
		return "", err
	}
	if !found {
		return fmt.Sprintf("Concept '%s' not found in bundle.", conceptID), nil
	}

	fm := concept.Doc.Frontmatter
	lines := []string{
		fmt.Sprintf("# Concept: %s", stringFrontmatterOr(fm["title"], concept.ID)),
		fmt.Sprintf("- **Type**: %s", stringFrontmatterOr(fm["type"], "Unknown")),
		fmt.Sprintf("- **Description**: %s", stringFrontmatterOr(fm["description"], "N/A")),
	}
	if resource := stringFrontmatterOr(fm["resource"], ""); resource != "" {
		lines = append(lines, fmt.Sprintf("- **Resource**: `%s`", resource))
	}
	if tags := frontmatterStringSlice(fm["tags"]); len(tags) > 0 {
		lines = append(lines, fmt.Sprintf("- **Tags**: %s", strings.Join(tags, ", ")))
	}
	if backlinks, err := Backlinks(root, concept.ID); err == nil && len(backlinks) > 0 {
		cited := make([]string, len(backlinks))
		for i, id := range backlinks {
			cited[i] = "`" + id + "`"
		}
		lines = append(lines, fmt.Sprintf("- **Cited by**: %s", strings.Join(cited, ", ")))
	}

	return strings.Join(lines, "\n") + "\n\n" + concept.Doc.Body, nil
}

// resolveConcept finds a concept by ID, trying the literal ID as a file
// path first. If no file exists there, it scans the bundle for a concept
// whose `aliases` frontmatter list contains conceptID as an exact,
// case-sensitive match. The bool return reports whether a concept was
// found at all.
func resolveConcept(root string, conceptID string) (ConceptFile, bool, error) {
	path := filepath.Join(root, strings.TrimLeft(conceptID, "/")+".md")
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return ConceptFile{}, false, err
		}
		doc, err := ParseDocument(string(data))
		if err != nil {
			return ConceptFile{}, false, err
		}
		return ConceptFile{ID: strings.TrimLeft(conceptID, "/"), Path: path, Doc: doc}, true, nil
	}

	concepts, err := walkConceptFiles(root)
	if err != nil {
		return ConceptFile{}, false, err
	}
	for _, c := range concepts {
		if c.Err == nil && slices.Contains(frontmatterStringSlice(c.Doc.Frontmatter["aliases"]), conceptID) {
			return c, true, nil
		}
	}
	return ConceptFile{}, false, nil
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
