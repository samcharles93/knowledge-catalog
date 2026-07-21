package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var codebaseIgnoredDirs = map[string]bool{
	".git": true, ".venv": true, "node_modules": true, "__pycache__": true,
	".pytest_cache": true, "dist": true, "build": true, ".okf": true,
	"target": true, "vendor": true,
}

var codebaseSourceExtensions = map[string]bool{
	".py": true, ".ts": true, ".js": true, ".go": true, ".rs": true,
	".java": true, ".cpp": true, ".h": true,
}

// CodebaseExtractor scans a local source tree and produces an architecture
// overview concept plus one Module concept per recognized source file.
type CodebaseExtractor struct {
	ProjectRoot string
}

var _ Extractor = CodebaseExtractor{}

// isRepoRoot reports whether root is the whole project's root, as opposed
// to a harvested subtree (e.g. `-src ./src`, a documented, supported
// usage), by checking for a ".git" entry directly under it. Only a
// repo-root harvest owns the singleton architecture/overview concept: a
// subtree harvest producing (and overwriting) a project-wide overview
// scoped to just that subtree would be actively misleading.
func isRepoRoot(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func (e CodebaseExtractor) ExtractConcepts() (map[string]Document, error) {
	root, err := filepath.Abs(e.ProjectRoot)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)

	readme := ""
	if data, err := os.ReadFile(filepath.Join(root, "README.md")); err == nil {
		readme = string(data)
	}

	var paths []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && codebaseIgnoredDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if codebaseSourceExtensions[filepath.Ext(path)] {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)

	concepts := map[string]Document{}
	var moduleLinks []string

	for _, path := range paths {
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			continue
		}
		relSlash := filepath.ToSlash(relPath)
		conceptID := "codebase/" + strings.TrimSuffix(relSlash, filepath.Ext(relSlash))

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := splitLines(string(data))
		firstLines := strings.Join(firstN(lines, 30), "\n")
		ext := strings.TrimPrefix(filepath.Ext(path), ".")
		name := filepath.Base(path)

		concepts[conceptID] = Document{
			Frontmatter: map[string]any{
				"type":        "Module",
				"title":       name,
				"description": fmt.Sprintf("Source module %s (%d lines).", relSlash, len(lines)),
				"resource":    relSlash,
				"tags":        []string{ext, "source"},
				"timestamp":   now,
			},
			Body: fmt.Sprintf(
				"# Module %s\n\n**Path**: `%s`  \n**Lines**: %d\n\n## Snippet Preview\n\n```\n%s\n```\n",
				name, relSlash, len(lines), firstLines,
			),
		}
		moduleLinks = append(moduleLinks, fmt.Sprintf("* [%s](/%s.md) - `%s`", name, conceptID, relSlash))
	}

	if isRepoRoot(root) {
		projectName := filepath.Base(root)
		overviewBody := fmt.Sprintf("# Overview\n\n%s\n\n# Codebase Navigation\n\n", readme)
		overviewBody += strings.Join(firstN(moduleLinks, 50), "\n")

		concepts["architecture/overview"] = Document{
			Frontmatter: map[string]any{
				"type":        "Architecture",
				"title":       projectName + " Overview",
				"description": fmt.Sprintf("Root architecture and project structure for %s.", projectName),
				"resource":    root,
				"tags":        []string{"overview", "architecture", "codebase"},
				"timestamp":   now,
			},
			Body: overviewBody,
		}
	}

	return concepts, nil
}

// ExportBundle only prunes on a repo-root harvest. Concept IDs are always
// root-relative ("codebase/" + path from whatever root was harvested), with
// nothing to distinguish "this subtree's slice of codebase/" from the whole
// namespace — so only a repo-root harvest, which walks the entire tree, is
// actually authoritative for codebase/ (and architecture/overview) as a
// whole. A subtree harvest (-src ./src, a documented, supported usage) only
// ever sees its own slice and must not prune anything: doing so would
// delete concepts belonging to the rest of the tree, exactly as it did
// before this fix.
func (e CodebaseExtractor) ExportBundle(bundleRoot string) (int, error) {
	root, err := filepath.Abs(e.ProjectRoot)
	if err != nil {
		return 0, err
	}
	var prunePrefixes []string
	if isRepoRoot(root) {
		prunePrefixes = []string{"codebase", "architecture/overview"}
	}
	return exportBundle(bundleRoot, e.ExtractConcepts, prunePrefixes)
}

func firstN(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
