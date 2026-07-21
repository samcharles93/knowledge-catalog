package okf

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

//go:embed viewer/viz.html viewer/viz.css viewer/viz.js
var vizAssets embed.FS

var typePalette = map[string]string{
	"Architecture": "#8b5cf6",
	"Service":      "#3b82f6",
	"API":          "#06b6d4",
	"Database":     "#10b981",
	"Table":        "#10b981",
	"Module":       "#f59e0b",
	"Class":        "#f59e0b",
	"Function":     "#f59e0b",
	"Rule":         "#ef4444",
	"Runbook":      "#ec4899",
	"Reference":    "#64748b",
	"Concept":      "#64748b",
}

const defaultNodeColor = "#94a3b8"

var vizLinkRE = regexp.MustCompile(`\]\(([^)\s]+\.md)(?:#[A-Za-z0-9_\-]*)?\)`)

// VizStats summarizes a generated visualization: how many concepts and
// edges were rendered, and the resulting HTML file size in bytes.
type VizStats struct {
	Concepts int
	Edges    int
	Bytes    int
}

type vizConcept struct {
	ID          string
	Type        string
	Title       string
	Description string
	Resource    string
	Tags        []string
	Body        string
	LinksTo     []string
}

func (c vizConcept) toNode() map[string]any {
	color, ok := typePalette[c.Type]
	if !ok {
		color = defaultNodeColor
	}
	label := c.Title
	if label == "" {
		label = c.ID
	}
	return map[string]any{
		"data": map[string]any{
			"id":          c.ID,
			"label":       label,
			"type":        c.Type,
			"description": c.Description,
			"resource":    c.Resource,
			"tags":        c.Tags,
			"color":       color,
			"size":        30 + min(60, len(c.Body)/200),
		},
	}
}

// extractVizLinks finds markdown links to other .md files in body and
// resolves them, relative to docDir, into bundle-root-relative concept IDs.
// Links outside the bundle, external links, and bundle-absolute links (the
// viewer only follows relative links, matching the Python reference) are
// skipped rather than erroring.
func extractVizLinks(body, docDir, bundleRoot string) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range vizLinkRE.FindAllStringSubmatch(body, -1) {
		target := m[1]
		if strings.Contains(target, "://") || strings.HasPrefix(target, "/") {
			continue
		}
		resolved := filepath.Clean(filepath.Join(docDir, target))
		rel, err := filepath.Rel(bundleRoot, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		id := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func walkVizConcepts(bundleRoot string) ([]vizConcept, error) {
	var paths []string
	err := filepath.WalkDir(bundleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") || d.Name() == indexFileName {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	concepts := make([]vizConcept, 0, len(paths))
	for _, path := range paths {
		id, err := ConceptID(bundleRoot, path)
		if err != nil {
			continue
		}
		conceptID := strings.Join(id, "/")

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		doc, err := ParseDocument(string(data))
		if err != nil {
			continue
		}

		fm := doc.Frontmatter
		tags := frontmatterStringSlice(fm["tags"])
		if tags == nil {
			tags = []string{}
		}

		concepts = append(concepts, vizConcept{
			ID:          conceptID,
			Type:        stringFrontmatterOr(fm["type"], "Unknown"),
			Title:       stringFrontmatterOr(fm["title"], conceptID),
			Description: stringFrontmatterOr(fm["description"], ""),
			Resource:    stringFrontmatterOr(fm["resource"], ""),
			Tags:        tags,
			Body:        doc.Body,
			LinksTo:     extractVizLinks(doc.Body, filepath.Dir(path), bundleRoot),
		})
	}
	return concepts, nil
}

func buildVizGraph(concepts []vizConcept) map[string]any {
	ids := make(map[string]bool, len(concepts))
	for _, c := range concepts {
		ids[c.ID] = true
	}

	nodes := make([]any, 0, len(concepts))
	bodies := make(map[string]string, len(concepts))
	typeSet := map[string]bool{}
	for _, c := range concepts {
		nodes = append(nodes, c.toNode())
		bodies[c.ID] = c.Body
		typeSet[c.Type] = true
	}

	edges := make([]any, 0)
	seenEdges := map[[2]string]bool{}
	for _, c := range concepts {
		for _, target := range c.LinksTo {
			if target == c.ID || !ids[target] {
				continue
			}
			key := [2]string{c.ID, target}
			if seenEdges[key] {
				continue
			}
			seenEdges[key] = true
			edges = append(edges, map[string]any{
				"data": map[string]any{
					"id":     fmt.Sprintf("%s__%s", c.ID, target),
					"source": c.ID,
					"target": target,
				},
			})
		}
	}

	types := make([]string, 0, len(typeSet))
	for t := range typeSet {
		types = append(types, t)
	}
	sort.Strings(types)

	return map[string]any{
		"nodes":   nodes,
		"edges":   edges,
		"bodies":  bodies,
		"types":   types,
		"palette": typePalette,
	}
}

// GenerateVisualization walks a bundle, builds a concept/link graph, and
// writes a standalone interactive HTML visualization to outPath.
func GenerateVisualization(bundleRoot string, outPath string, bundleName string) (VizStats, error) {
	info, err := os.Stat(bundleRoot)
	if err != nil || !info.IsDir() {
		return VizStats{}, fmt.Errorf("bundle directory not found: %s", bundleRoot)
	}

	concepts, err := walkVizConcepts(bundleRoot)
	if err != nil {
		return VizStats{}, err
	}
	graph := buildVizGraph(concepts)

	tmpl, err := vizAssets.ReadFile("viewer/viz.html")
	if err != nil {
		return VizStats{}, err
	}
	css, err := vizAssets.ReadFile("viewer/viz.css")
	if err != nil {
		return VizStats{}, err
	}
	js, err := vizAssets.ReadFile("viewer/viz.js")
	if err != nil {
		return VizStats{}, err
	}

	name := bundleName
	if name == "" {
		abs, err := filepath.Abs(bundleRoot)
		if err != nil {
			abs = bundleRoot
		}
		name = filepath.Base(abs)
	}
	nameJSON, err := json.Marshal(name)
	if err != nil {
		return VizStats{}, err
	}
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return VizStats{}, err
	}

	html := string(tmpl)
	html = strings.Replace(html, "/*__VIZ_CSS__*/", string(css), 1)
	html = strings.Replace(html, "/*__VIZ_JS__*/", string(js), 1)
	html = strings.Replace(html, "__BUNDLE_NAME__", string(nameJSON), 1)
	html = strings.Replace(html, "__BUNDLE_DATA__", string(graphJSON), 1)

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return VizStats{}, err
	}
	if err := os.WriteFile(outPath, []byte(html), 0o644); err != nil {
		return VizStats{}, err
	}

	return VizStats{
		Concepts: len(concepts),
		Edges:    len(edges(graph)),
		Bytes:    len(html),
	}, nil
}

func edges(graph map[string]any) []any {
	e, _ := graph["edges"].([]any)
	return e
}
