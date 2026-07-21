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

func keysOf(m map[string]Document) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
