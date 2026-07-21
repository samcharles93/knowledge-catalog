package okf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const indexFileName = "index.md"

var reservedConceptFileNames = map[string]bool{
	"index.md": true,
	"log.md":   true,
}

var linkRE = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

// ConceptFile is one discovered concept document: its ID, on-disk path,
// raw bytes, and parsed contents. Err is set instead of Doc/Raw when the
// file could not be read (ParseDocument itself is self-healing and never
// fails).
type ConceptFile struct {
	ID   string
	Path string
	Raw  []byte
	Doc  Document
	Err  error
}

// walkConceptFiles discovers every concept document in a bundle (all .md
// files except reserved names like index.md/log.md), sorted by path. It is
// the single shared traversal used by index generation, validation, the
// viewer, and the MCP server, so "what counts as a concept file" only has
// to be defined once.
func walkConceptFiles(root string) ([]ConceptFile, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") || reservedConceptFileNames[d.Name()] {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	files := make([]ConceptFile, 0, len(paths))
	for _, path := range paths {
		id, err := ConceptID(root, path)
		if err != nil {
			continue
		}
		conceptID := strings.Join(id, "/")

		data, err := os.ReadFile(path)
		if err != nil {
			files = append(files, ConceptFile{ID: conceptID, Path: path, Err: err})
			continue
		}
		doc, err := ParseDocument(string(data))
		if err != nil {
			files = append(files, ConceptFile{ID: conceptID, Path: path, Err: err})
			continue
		}
		files = append(files, ConceptFile{ID: conceptID, Path: path, Raw: data, Doc: doc})
	}
	return files, nil
}

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

// SearchConcepts returns every concept in the bundle whose title, body, or
// tags contain query as a case-insensitive substring. An empty query
// matches every concept. Per SPEC.md §6.2 (okf_search_concepts): "Semantic/
// text search across tags, titles, and bodies" — this implements the text
// half; semantic (embedding-based) search is out of scope.
func SearchConcepts(root string, query string) ([]ConceptFile, error) {
	files, err := walkConceptFiles(root)
	if err != nil {
		return nil, err
	}
	lowerQuery := strings.ToLower(query)

	matches := make([]ConceptFile, 0, len(files))
	for _, f := range files {
		if f.Err == nil && conceptMatchesQuery(f, lowerQuery) {
			matches = append(matches, f)
		}
	}
	return matches, nil
}

func conceptMatchesQuery(f ConceptFile, lowerQuery string) bool {
	if lowerQuery == "" {
		return true
	}
	fm := f.Doc.Frontmatter
	title := stringFrontmatterOr(fm["title"], f.ID)
	if strings.Contains(strings.ToLower(title), lowerQuery) {
		return true
	}
	if strings.Contains(strings.ToLower(f.Doc.Body), lowerQuery) {
		return true
	}
	for _, tag := range frontmatterStringSlice(fm["tags"]) {
		if strings.Contains(strings.ToLower(tag), lowerQuery) {
			return true
		}
	}
	return false
}

// linkGraph is the resolved directed edges between concepts, keyed by
// concept ID, plus any links that didn't resolve to a known concept.
// Shared by ValidateBundle (link-integrity/orphan checks) and Backlinks
// (the "cited by" query) so link resolution is defined exactly once.
type linkGraph struct {
	outgoing map[string][]string
	incoming map[string][]string
	broken   []Issue
}

// buildLinkGraph resolves every markdown link in documents (relative to
// each doc's directory, or bundle-root-relative for a leading "/") into
// concept IDs. A link that doesn't resolve to any file in conceptFiles
// (and isn't the reserved "index" target) is recorded as a broken-link
// warning.
func buildLinkGraph(root string, concepts []ConceptFile, conceptFiles map[string]string) linkGraph {
	graph := linkGraph{outgoing: map[string][]string{}, incoming: map[string][]string{}}

	for _, c := range concepts {
		if c.Err != nil {
			continue
		}
		var links []string
		for _, m := range linkRE.FindAllStringSubmatch(c.Doc.Body, -1) {
			target := strings.TrimSpace(m[2])
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") ||
				strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "#") {
				continue
			}

			var targetClean string
			if rest, ok := strings.CutPrefix(target, "/"); ok {
				targetClean = strings.TrimSuffix(rest, ".md")
			} else {
				docDir := filepath.Dir(c.Path)
				resolved := filepath.Clean(filepath.Join(docDir, target))
				if rel, err := filepath.Rel(root, resolved); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					targetClean = strings.TrimSuffix(filepath.ToSlash(rel), ".md")
				} else {
					targetClean = target
				}
			}

			links = append(links, targetClean)
			graph.incoming[targetClean] = append(graph.incoming[targetClean], c.ID)

			if _, ok := conceptFiles[targetClean]; !ok && targetClean != "index" {
				graph.broken = append(graph.broken, Issue{
					ConceptID: c.ID, Path: c.Path,
					Message: fmt.Sprintf("Broken link target: '%s' (resolved as '%s')", target, targetClean),
					Warning: true,
				})
			}
		}
		graph.outgoing[c.ID] = links
	}

	return graph
}

// Backlinks returns the IDs of every concept that links to conceptID,
// sorted and deduplicated, resolved the same way ValidateBundle resolves
// links. An unknown or uncited conceptID returns an empty slice, not an
// error — SPEC.md §4.2 describes backlinks ("Cited By") as a graph query,
// not a validity check.
func Backlinks(root string, conceptID string) ([]string, error) {
	concepts, err := walkConceptFiles(root)
	if err != nil {
		return nil, err
	}
	conceptFiles := make(map[string]string, len(concepts))
	for _, c := range concepts {
		conceptFiles[c.ID] = c.Path
	}

	graph := buildLinkGraph(root, concepts, conceptFiles)

	seen := map[string]bool{}
	backlinks := make([]string, 0, len(graph.incoming[conceptID]))
	for _, id := range graph.incoming[conceptID] {
		if !seen[id] {
			seen[id] = true
			backlinks = append(backlinks, id)
		}
	}
	sort.Strings(backlinks)
	return backlinks, nil
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

	concepts, err := walkConceptFiles(root)
	if err != nil {
		report.Errors = append(report.Errors, Issue{Message: fmt.Sprintf("Failed to walk bundle: %v", err), Path: root})
		return report
	}
	report.TotalConcepts = len(concepts)

	conceptFiles := make(map[string]string, len(concepts))
	for _, c := range concepts {
		conceptFiles[c.ID] = c.Path
	}

	for _, c := range concepts {
		if c.Err != nil {
			report.Errors = append(report.Errors, Issue{
				ConceptID: c.ID, Path: c.Path,
				Message: fmt.Sprintf("Failed to read/parse document: %v", c.Err),
			})
			continue
		}
		if !utf8.Valid(c.Raw) {
			report.Errors = append(report.Errors, Issue{
				ConceptID: c.ID, Path: c.Path, Message: "File is not valid UTF-8",
			})
			continue
		}
		if bytes.ContainsRune(c.Raw, '\r') {
			report.Warnings = append(report.Warnings, Issue{
				ConceptID: c.ID, Path: c.Path,
				Message: "File does not use Unix (\\n) line endings", Warning: true,
			})
		}
		if err := c.Doc.Validate(); err != nil {
			report.Errors = append(report.Errors, Issue{
				ConceptID: c.ID, Path: c.Path, Message: err.Error(),
			})
			continue
		}
		for _, w := range c.Doc.Warnings() {
			report.Warnings = append(report.Warnings, Issue{
				ConceptID: c.ID, Path: c.Path, Message: w, Warning: true,
			})
		}
	}

	// Link resolution covers every readable concept (not just ones that
	// passed frontmatter validation) — a citation is still a citation even
	// if the citing doc is missing its `type` field.
	graph := buildLinkGraph(root, concepts, conceptFiles)
	report.Warnings = append(report.Warnings, graph.broken...)

	if report.TotalConcepts > 1 {
		for conceptID, path := range conceptFiles {
			if len(graph.outgoing[conceptID]) == 0 && len(graph.incoming[conceptID]) == 0 {
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
