package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var rememberDirs = map[string]string{
	"Architecture": "architecture",
	"Service":      "services",
	"Rule":         "rules",
	"Runbook":      "runbooks",
	"Concept":      "concepts",
}

// rememberReservedTypes are owned by a harvest Extractor's pruning: a
// manually-written concept placed there would be silently deleted by the
// next matching harvest run.
var rememberReservedTypes = map[string]bool{
	"Codebase":  true, // owned by CodebaseExtractor, codebase/
	"API":       true, // owned by OpenAPIExtractor, api/
	"Table":     true, // owned by SQLExtractor, database/
	"Reference": true, // owned by WebExtractor, references/
}

// RememberInput captures a single deliberate, free-form memory (a coding
// rule, session insight, runbook step, etc.) to be written as one OKF
// concept document. Unlike an Extractor, this is a one-shot write, not a
// repeatable mechanical harvest: no ExtractConcepts/ExportBundle/pruning.
type RememberInput struct {
	Type        string   // required; Rule, Runbook, Concept, Service, Architecture, or a custom type
	Title       string   // required; derives the concept ID via slugify
	Body        string   // required; markdown body
	Description string   // optional
	Resource    string   // optional; file path, PR link, or session reference
	Tags        []string // optional
}

// Remember validates and writes a single concept document into bundleRoot
// under a type-derived top-level directory, then regenerates bundle
// indexes. It returns the concept ID the document was written under.
//
// Codebase, API, Table, and Reference are rejected: those namespaces are
// owned by a harvest Extractor's pruning (see pruneStaleConcepts in
// extract.go), so a manually-written concept placed there would be
// silently deleted the next time that harvest type runs.
func Remember(bundleRoot string, in RememberInput) (string, error) {
	if strings.TrimSpace(in.Type) == "" {
		return "", fmt.Errorf("remember: type is required")
	}
	if rememberReservedTypes[in.Type] {
		return "", fmt.Errorf("remember: type %q is owned by a harvest extractor and would be pruned by the next harvest; use Rule, Runbook, Concept, Service, Architecture, or a custom type instead", in.Type)
	}
	if strings.TrimSpace(in.Title) == "" {
		return "", fmt.Errorf("remember: title is required")
	}
	if strings.TrimSpace(in.Body) == "" {
		return "", fmt.Errorf("remember: body is required")
	}

	conceptID := rememberDir(in.Type) + "/" + slugify(in.Title)

	fm := map[string]any{
		"type":      in.Type,
		"title":     in.Title,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if in.Description != "" {
		fm["description"] = in.Description
	}
	if in.Resource != "" {
		fm["resource"] = in.Resource
	}
	if len(in.Tags) > 0 {
		fm["tags"] = in.Tags
	}

	doc := Document{Frontmatter: fm, Body: in.Body}
	if err := doc.Validate(); err != nil {
		return "", err
	}

	path, err := ConceptPath(bundleRoot, strings.Split(conceptID, "/"))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(doc.String()), 0o644); err != nil {
		return "", err
	}
	if _, err := RegenerateIndexes(bundleRoot); err != nil {
		return conceptID, err
	}
	return conceptID, nil
}

// rememberDir maps a concept type to its top-level bundle directory,
// falling back to a slugified-lowercase form of the type itself for
// anything not in rememberDirs (custom types are never harvest-owned, so
// no pruning risk applies to the fallback).
func rememberDir(conceptType string) string {
	if dir, ok := rememberDirs[conceptType]; ok {
		return dir
	}
	return strings.ToLower(slugify(conceptType))
}
