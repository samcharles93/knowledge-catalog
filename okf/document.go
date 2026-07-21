package okf

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const frontmatterDelim = "---"

var (
	requiredFrontmatterKeys    = []string{"type"}
	recommendedFrontmatterKeys = []string{"title", "description"}
)

// Document is a parsed OKF markdown file: YAML frontmatter plus body text.
type Document struct {
	Frontmatter map[string]any
	Body        string
}

// splitLines mimics Python's str.splitlines(): it splits on line
// boundaries without producing a trailing empty element for a final
// newline, and normalizes CRLF/CR to LF first.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}

// noFrontmatter returns source as-is: the self-healing fallback used
// whenever a leading "---" isn't found, isn't closed, or doesn't parse to
// a YAML mapping. Malformed input never produces a parse error.
func noFrontmatter(source string) (Document, error) {
	return Document{Frontmatter: map[string]any{}, Body: source}, nil
}

// ParseDocument parses raw markdown source into a Document, splitting out
// any leading "---" delimited YAML frontmatter block. Frontmatter that is
// unterminated, not a YAML mapping, or otherwise unparseable self-heals to
// "no frontmatter": the original source is preserved verbatim as the body.
func ParseDocument(source string) (Document, error) {
	lines := splitLines(source)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != frontmatterDelim {
		return noFrontmatter(source)
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == frontmatterDelim {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return noFrontmatter(source)
	}

	fmText := strings.Join(lines[1:endIdx], "\n")
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(fmText), &raw); err != nil {
		return noFrontmatter(source)
	}
	if raw == nil {
		raw = map[string]any{}
	}
	frontmatter := normalizeYAMLMap(raw)

	body := strings.Join(lines[endIdx+1:], "\n")
	body = strings.TrimPrefix(body, "\n")

	return Document{Frontmatter: frontmatter, Body: body}, nil
}

// normalizeYAMLMap converts a yaml.v3-decoded map so that homogeneous
// string sequences (e.g. tags: [a, b]) come out as []string rather than
// []any, matching how callers construct Document literals by hand.
func normalizeYAMLMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = normalizeYAMLValue(v)
	}
	return out
}

func normalizeYAMLValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return normalizeYAMLMap(val)
	case []any:
		strs := make([]string, len(val))
		allStrings := true
		for i, elem := range val {
			s, ok := elem.(string)
			if !ok {
				allStrings = false
				break
			}
			strs[i] = s
		}
		if allStrings {
			return strs
		}
		normalized := make([]any, len(val))
		for i, elem := range val {
			normalized[i] = normalizeYAMLValue(elem)
		}
		return normalized
	default:
		return v
	}
}

// String serializes the Document back to frontmatter + body markdown text.
// A document with no frontmatter serializes to its body alone (no
// delimiters invented); the body always ends with exactly one trailing
// newline.
func (d Document) String() string {
	body := d.Body
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}

	if len(d.Frontmatter) == 0 {
		return body
	}

	fmBytes, err := yaml.Marshal(d.Frontmatter)
	if err != nil {
		return body
	}
	fmText := strings.TrimRight(string(fmBytes), "\n")

	if body == "\n" {
		return fmt.Sprintf("%s\n%s\n%s\n\n", frontmatterDelim, fmText, frontmatterDelim)
	}
	return fmt.Sprintf("%s\n%s\n%s\n\n%s", frontmatterDelim, fmText, frontmatterDelim, body)
}

// Validate checks that required frontmatter keys are present.
func (d Document) Validate() error {
	var missing []string
	for _, k := range requiredFrontmatterKeys {
		v, ok := d.Frontmatter[k]
		if !ok || isZeroFrontmatterValue(v) {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required frontmatter keys: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Warnings reports missing recommended (non-required) frontmatter keys.
func (d Document) Warnings() []string {
	var warnings []string
	for _, k := range recommendedFrontmatterKeys {
		v, ok := d.Frontmatter[k]
		if !ok || isZeroFrontmatterValue(v) {
			warnings = append(warnings, fmt.Sprintf("missing recommended frontmatter key: %q", k))
		}
	}
	return warnings
}

func isZeroFrontmatterValue(v any) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return s == ""
	}
	return false
}
