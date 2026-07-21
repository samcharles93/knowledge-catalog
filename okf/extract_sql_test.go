package okf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLExtractorParsesCreateTable(t *testing.T) {
	t.Parallel()

	sqlFile := filepath.Join(t.TempDir(), "schema.sql")
	sql := "CREATE TABLE users (\n" +
		"  id INT PRIMARY KEY,\n" +
		"  email VARCHAR(255)\n" +
		");"
	if err := os.WriteFile(sqlFile, []byte(sql), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := SQLExtractor{SQLPath: sqlFile}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	doc, ok := concepts["database/users"]
	if !ok {
		t.Fatal("missing database/users concept")
	}
	if doc.Frontmatter["type"] != "Table" {
		t.Errorf("type = %v, want Table", doc.Frontmatter["type"])
	}
	if !strings.Contains(doc.Body, "email") {
		t.Errorf("body = %q, want column listing", doc.Body)
	}
}

func TestSQLExtractorHandlesMultipleTablesAndQuotedNames(t *testing.T) {
	t.Parallel()

	sqlFile := filepath.Join(t.TempDir(), "schema.sql")
	sql := "CREATE TABLE `orders` (\n  id INT\n);\n\n" +
		"CREATE TABLE IF NOT EXISTS \"public\".\"products\" (\n  sku TEXT\n);"
	if err := os.WriteFile(sqlFile, []byte(sql), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := SQLExtractor{SQLPath: sqlFile}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	if _, ok := concepts["database/orders"]; !ok {
		t.Error("missing database/orders concept")
	}
	if _, ok := concepts["database/products"]; !ok {
		t.Error("missing database/products concept (schema-qualified name should be stripped)")
	}
}

func TestSQLExtractorReturnsEmptyWhenNoTablesFound(t *testing.T) {
	t.Parallel()

	sqlFile := filepath.Join(t.TempDir(), "schema.sql")
	if err := os.WriteFile(sqlFile, []byte("SELECT 1;"), 0o644); err != nil {
		t.Fatal(err)
	}

	ext := SQLExtractor{SQLPath: sqlFile}
	concepts, err := ext.ExtractConcepts()
	if err != nil {
		t.Fatalf("ExtractConcepts() error = %v", err)
	}
	if len(concepts) != 0 {
		t.Errorf("len(concepts) = %d, want 0", len(concepts))
	}
}
