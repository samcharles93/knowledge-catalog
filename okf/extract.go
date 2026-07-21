package okf

import (
	"os"
	"path/filepath"
	"strings"
)

// Extractor turns some external, vendor-neutral source into OKF concept
// documents and can export them directly into a bundle.
type Extractor interface {
	// ExtractConcepts returns the concept documents keyed by concept ID.
	ExtractConcepts() (map[string]Document, error)
	// ExportBundle writes the extracted concepts into bundleRoot and
	// regenerates indexes, returning the number of concepts written.
	ExportBundle(bundleRoot string) (int, error)
}

// exportBundle writes the concepts produced by extract into bundleRoot (one
// {concept_id}.md file per concept) and regenerates the bundle's indexes.
// Shared by every Extractor's ExportBundle method.
//
// prunePrefixes lists the concept ID prefixes this invocation is
// authoritative for (e.g. "codebase" for a codebase harvest): any existing
// bundle file under one of those prefixes that is absent from the fresh
// concepts map is stale (its source was renamed/removed since the last
// harvest) and gets deleted, along with any directory left empty by that
// deletion. Pass nil to skip pruning entirely, for extractors whose harvests
// are additive across runs (e.g. WebExtractor, built up via repeated --url
// flags) rather than a full snapshot of a single source.
func exportBundle(bundleRoot string, extract func() (map[string]Document, error), prunePrefixes []string) (int, error) {
	concepts, err := extract()
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(bundleRoot, 0o755); err != nil {
		return 0, err
	}
	if err := pruneStaleConcepts(bundleRoot, prunePrefixes, concepts); err != nil {
		return 0, err
	}
	for id, doc := range concepts {
		path := filepath.Join(bundleRoot, id+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return 0, err
		}
		if err := os.WriteFile(path, []byte(doc.String()), 0o644); err != nil {
			return 0, err
		}
	}
	if _, err := RegenerateIndexes(bundleRoot); err != nil {
		return len(concepts), err
	}
	return len(concepts), nil
}

// ownedByPrefix reports whether conceptID falls under one of the given
// owned prefixes, matching the prefix itself exactly or any of its
// descendants (a "/"-delimited child).
func ownedByPrefix(conceptID string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if conceptID == prefix || strings.HasPrefix(conceptID, prefix+"/") {
			return true
		}
	}
	return false
}

// pruneStaleConcepts deletes existing bundle files under prunePrefixes that
// are no longer present in keep, then removes any directories left empty by
// those deletions.
func pruneStaleConcepts(bundleRoot string, prunePrefixes []string, keep map[string]Document) error {
	if len(prunePrefixes) == 0 {
		return nil
	}
	existing, err := walkConceptFiles(bundleRoot)
	if err != nil {
		return err
	}
	for _, cf := range existing {
		if !ownedByPrefix(cf.ID, prunePrefixes) {
			continue
		}
		if _, ok := keep[cf.ID]; ok {
			continue
		}
		if err := os.Remove(cf.Path); err != nil {
			return err
		}
	}
	return removeEmptyDirs(bundleRoot)
}

// removeEmptyDirs removes empty directories nested under root, bottom-up,
// leaving root itself in place even if it ends up empty.
func removeEmptyDirs(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := removeEmptySubdir(filepath.Join(root, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// removeEmptySubdir removes dir, and everything nested under it, bottom-up,
// if and only if it ends up containing no concept files. A stale index.md
// left over from a previous RegenerateIndexes call doesn't count: it would
// itself be deleted (or simply not regenerated) once the directory has no
// concepts left, so it shouldn't keep an otherwise-empty directory alive.
func removeEmptySubdir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			if err := removeEmptySubdir(filepath.Join(dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	entries, err = os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		// A subdirectory that survived recursion still has concepts of its
		// own; a file that isn't a reserved index/log is a concept too.
		// Either means dir isn't actually empty.
		if entry.IsDir() || !reservedConceptFileNames[entry.Name()] {
			return nil
		}
	}
	return os.RemoveAll(dir)
}
