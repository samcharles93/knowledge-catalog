package okf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPIExtractorExtractsOperations(t *testing.T) {
	t.Parallel()

	spec := filepath.Join(t.TempDir(), "openapi.json")
	body, err := json.Marshal(map[string]any{
		"openapi": "3.0.0",
		"info":    map[string]any{"title": "Test API", "version": "1.0"},
		"paths": map[string]any{
			"/users": map[string]any{
				"get": map[string]any{
					"summary":     "List users",
					"operationId": "listUsers",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(spec, body, 0o644); err != nil {
		t.Fatal(err)
	}

	ext := OpenAPIExtractor{SpecPath: spec}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	doc, ok := concepts["api/listUsers"]
	if !ok {
		t.Fatal("missing api/listUsers concept")
	}
	if doc.Frontmatter["type"] != "API" {
		t.Errorf("type = %v, want API", doc.Frontmatter["type"])
	}
	if doc.Frontmatter["description"] != "List users" {
		t.Errorf("description = %v, want %q", doc.Frontmatter["description"], "List users")
	}
}

func TestOpenAPIExtractorGeneratesOperationIDWhenMissing(t *testing.T) {
	t.Parallel()

	spec := filepath.Join(t.TempDir(), "openapi.yaml")
	yamlSpec := "openapi: 3.0.0\n" +
		"info:\n  title: Test API\n  version: '1.0'\n" +
		"paths:\n" +
		"  /users/{id}:\n" +
		"    delete:\n" +
		"      summary: Delete user\n"
	if err := os.WriteFile(spec, []byte(yamlSpec), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := OpenAPIExtractor{SpecPath: spec}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	doc, ok := concepts["api/delete__users__id_"]
	if !ok {
		t.Fatalf("expected generated operation id concept, got %v", keysOf(concepts))
	}
	if doc.Frontmatter["title"] != "DELETE /users/{id}" {
		t.Errorf("title = %v", doc.Frontmatter["title"])
	}
}

// TestOpenAPIExtractorExportBundlePrunesRemovedOperations covers re-harvesting
// the same spec path after an operation was renamed/removed: the extractor
// fully owns the api/ namespace for a given spec, so stale operation concepts
// from a previous version of the spec must not linger.
func TestOpenAPIExtractorExportBundlePrunesRemovedOperations(t *testing.T) {
	t.Parallel()

	spec := filepath.Join(t.TempDir(), "openapi.json")
	writeSpec := func(opID string) {
		t.Helper()
		body, err := json.Marshal(map[string]any{
			"openapi": "3.0.0",
			"info":    map[string]any{"title": "Test API", "version": "1.0"},
			"paths": map[string]any{
				"/users": map[string]any{
					"get": map[string]any{"operationId": opID},
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(spec, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeSpec("listUsers")
	out := t.TempDir()
	ext := OpenAPIExtractor{SpecPath: spec}
	if _, err := ext.ExportBundle(out); err != nil {
		t.Fatalf("ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "api", "listUsers.md")); err != nil {
		t.Fatalf("api/listUsers.md not written: %v", err)
	}

	writeSpec("getUsers")
	if _, err := ext.ExportBundle(out); err != nil {
		t.Fatalf("second ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "api", "listUsers.md")); !os.IsNotExist(err) {
		t.Errorf("api/listUsers.md should have been pruned after the operationId was renamed, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "api", "getUsers.md")); err != nil {
		t.Errorf("api/getUsers.md not written: %v", err)
	}
}

func keysOf(m map[string]Document) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
