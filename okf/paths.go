package okf

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var segmentRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.\-]*$`)

func validateSegment(seg string) error {
	if !segmentRE.MatchString(seg) {
		return fmt.Errorf("invalid concept id segment: %q", seg)
	}
	return nil
}

// ParseConceptID splits a slash-separated concept identifier into its
// segments, validating each segment and rejecting empty or traversal input.
func ParseConceptID(id string) ([]string, error) {
	var parts []string
	for p := range strings.SplitSeq(id, "/") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty concept id: %q", id)
	}
	for _, p := range parts {
		if err := validateSegment(p); err != nil {
			return nil, err
		}
	}
	return parts, nil
}

// ConceptPath resolves a concept ID to its on-disk markdown path under root.
func ConceptPath(root string, id []string) (string, error) {
	if len(id) == 0 {
		return "", fmt.Errorf("concept id must have at least one segment")
	}
	for _, seg := range id {
		if err := validateSegment(seg); err != nil {
			return "", err
		}
	}
	dirs := id[:len(id)-1]
	name := id[len(id)-1]
	parts := make([]string, 0, len(dirs)+2)
	parts = append(parts, root)
	parts = append(parts, dirs...)
	parts = append(parts, name+".md")
	return filepath.Join(parts...), nil
}

// ConceptID derives a concept ID from a markdown file path relative to root.
func ConceptID(root string, path string) ([]string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("path %q is outside bundle root %q", path, root)
	}
	rel = strings.TrimSuffix(rel, filepath.Ext(rel))
	return strings.Split(filepath.ToSlash(rel), "/"), nil
}
