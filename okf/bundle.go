package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const indexFileName = "index.md"

var reservedConceptFileNames = map[string]bool{
	"index.md": true,
	"log.md":   true,
}

var linkRE = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// Issue is a single validation error or warning attached to a concept.
type Issue struct {
	ConceptID string
	Path      string
	Message   string
	Warning   bool
}

// Report is the result of validating an OKF bundle.
type Report struct {
	TotalConcepts int
	Errors        []Issue
	Warnings      []Issue
}

// Valid reports whether the bundle has no validation errors (warnings are
// permitted).
func (r Report) Valid() bool {
	return len(r.Errors) == 0
}

// RegenerateIndexes rebuilds index.md for every directory in the bundle that
// contains concept documents, returning the paths written.
func RegenerateIndexes(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil, nil
	}

	dirs, err := directoriesToIndex(root)
	if err != nil {
		return nil, err
	}
	sort.Slice(dirs, func(i, j int) bool {
		di, dj := pathDepth(root, dirs[i]), pathDepth(root, dirs[j])
		if di != dj {
			return di > dj
		}
		return dirs[i] < dirs[j]
	})

	dirDescriptions := map[string]string{}
	var written []string

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return written, err
		}

		var rows []indexRow
		var pairs []ChildSummary
		for _, ent := range entries {
			name := ent.Name()
			if name == indexFileName {
				continue
			}
			full := filepath.Join(dir, name)
			switch {
			case !ent.IsDir() && strings.HasSuffix(name, ".md"):
				data, err := os.ReadFile(full)
				if err != nil {
					continue
				}
				doc, err := ParseDocument(string(data))
				if err != nil {
					continue
				}
				title := stringFrontmatterOr(doc.Frontmatter["title"], strings.TrimSuffix(name, ".md"))
				desc := stringFrontmatterOr(doc.Frontmatter["description"], "")
				typ := stringFrontmatterOr(doc.Frontmatter["type"], "Concept")
				rows = append(rows, indexRow{Type: typ, Title: title, Link: name, Description: desc})
				pairs = append(pairs, ChildSummary{Title: title, Description: desc})
			case ent.IsDir():
				desc := dirDescriptions[full]
				rows = append(rows, indexRow{
					Type:        "Subdirectories",
					Title:       name,
					Link:        name + "/" + indexFileName,
					Description: desc,
				})
				pairs = append(pairs, ChildSummary{Title: name, Description: desc})
			}
		}

		if len(rows) == 0 {
			continue
		}

		indexPath := filepath.Join(dir, indexFileName)
		if err := os.WriteFile(indexPath, []byte(buildIndexText(rows)), 0o644); err != nil {
			return written, err
		}
		written = append(written, indexPath)

		if dir == root {
			continue
		}

		if len(pairs) == 1 && pairs[0].Description != "" {
			dirDescriptions[dir] = pairs[0].Description
		} else {
			rel, _ := filepath.Rel(root, dir)
			dirDescriptions[dir] = SynthesizeDescription(rel, pairs)
		}
	}

	return written, nil
}

type indexRow struct {
	Type        string
	Title       string
	Link        string
	Description string
}

func directoriesToIndex(root string) ([]string, error) {
	dirSet := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		cur := filepath.Dir(path)
		for {
			dirSet[cur] = true
			if cur == root {
				break
			}
			cur = filepath.Dir(cur)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(dirSet))
	for d := range dirSet {
		dirs = append(dirs, d)
	}
	return dirs, nil
}

func pathDepth(root, dir string) int {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return 0
	}
	return len(strings.Split(filepath.ToSlash(rel), "/"))
}

func stringFrontmatterOr(v any, fallback string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return fallback
}

func buildIndexText(entries []indexRow) string {
	grouped := map[string][]indexRow{}
	var types []string
	for _, e := range entries {
		typ := e.Type
		if typ == "" {
			typ = "Other"
		}
		if _, ok := grouped[typ]; !ok {
			types = append(types, typ)
		}
		grouped[typ] = append(grouped[typ], e)
	}
	sort.Strings(types)

	sections := make([]string, 0, len(types))
	for _, typ := range types {
		rows := grouped[typ]
		sort.SliceStable(rows, func(i, j int) bool {
			return strings.ToLower(rows[i].Title) < strings.ToLower(rows[j].Title)
		})
		lines := []string{"# " + typ, ""}
		for _, r := range rows {
			suffix := ""
			if r.Description != "" {
				suffix = " - " + r.Description
			}
			lines = append(lines, fmt.Sprintf("* [%s](%s)%s", r.Title, r.Link, suffix))
		}
		sections = append(sections, strings.Join(lines, "\n"))
	}
	return strings.Join(sections, "\n\n") + "\n"
}

// ValidateBundle walks a bundle root, parses every concept document,
// checks required/recommended frontmatter, link integrity, and orphan
// concepts, and returns the aggregated report.
func ValidateBundle(root string) Report {
	var report Report

	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		report.Errors = append(report.Errors, Issue{
			Message: fmt.Sprintf("Bundle directory does not exist: %s", root),
			Path:    root,
		})
		return report
	}

	conceptFiles := map[string]string{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if reservedConceptFileNames[d.Name()] {
			return nil
		}
		id, err := ConceptID(root, path)
		if err != nil {
			return nil
		}
		conceptFiles[strings.Join(id, "/")] = path
		return nil
	})
	report.TotalConcepts = len(conceptFiles)

	documents := map[string]Document{}
	for conceptID, path := range conceptFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			report.Errors = append(report.Errors, Issue{
				ConceptID: conceptID, Path: path,
				Message: fmt.Sprintf("Failed to read/parse document: %v", err),
			})
			continue
		}
		doc, err := ParseDocument(string(data))
		if err != nil {
			report.Errors = append(report.Errors, Issue{
				ConceptID: conceptID, Path: path,
				Message: fmt.Sprintf("Failed to read/parse document: %v", err),
			})
			continue
		}
		if err := doc.Validate(); err != nil {
			report.Errors = append(report.Errors, Issue{
				ConceptID: conceptID, Path: path, Message: err.Error(),
			})
			continue
		}
		for _, w := range doc.Warnings() {
			report.Warnings = append(report.Warnings, Issue{
				ConceptID: conceptID, Path: path, Message: w, Warning: true,
			})
		}
		documents[conceptID] = doc
	}

	outgoingLinks := map[string][]string{}
	incomingLinks := map[string][]string{}
	for conceptID, doc := range documents {
		var links []string
		for _, m := range linkRE.FindAllStringSubmatch(doc.Body, -1) {
			target := strings.TrimSpace(m[2])
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") ||
				strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
				continue
			}

			var targetClean string
			if rest, ok := strings.CutPrefix(target, "/"); ok {
				targetClean = strings.TrimSuffix(rest, ".md")
			} else {
				docDir := filepath.Dir(conceptFiles[conceptID])
				resolved := filepath.Clean(filepath.Join(docDir, target))
				if rel, err := filepath.Rel(root, resolved); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					targetClean = strings.TrimSuffix(filepath.ToSlash(rel), ".md")
				} else {
					targetClean = target
				}
			}

			links = append(links, targetClean)
			incomingLinks[targetClean] = append(incomingLinks[targetClean], conceptID)

			if _, ok := conceptFiles[targetClean]; !ok && targetClean != "index" {
				report.Warnings = append(report.Warnings, Issue{
					ConceptID: conceptID, Path: conceptFiles[conceptID],
					Message: fmt.Sprintf("Broken link target: '%s' (resolved as '%s')", target, targetClean),
					Warning: true,
				})
			}
		}
		outgoingLinks[conceptID] = links
	}

	if report.TotalConcepts > 1 {
		for conceptID, path := range conceptFiles {
			if len(outgoingLinks[conceptID]) == 0 && len(incomingLinks[conceptID]) == 0 {
				report.Warnings = append(report.Warnings, Issue{
					ConceptID: conceptID, Path: path,
					Message: "Orphan concept: has no incoming or outgoing links in bundle graph",
					Warning: true,
				})
			}
		}
	}

	return report
}
