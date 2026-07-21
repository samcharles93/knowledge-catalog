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

// TestSQLExtractorExportBundlePrunesDroppedTables covers re-harvesting the
// same DDL file after a table was dropped/renamed: the extractor fully owns
// the database/ namespace for a given schema file, so stale table concepts
// from a previous version of the schema must not linger.
func TestSQLExtractorExportBundlePrunesDroppedTables(t *testing.T) {
	t.Parallel()

	sqlFile := filepath.Join(t.TempDir(), "schema.sql")
	if err := os.WriteFile(sqlFile, []byte("CREATE TABLE users (\n  id INT\n);"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	ext := SQLExtractor{SQLPath: sqlFile}
	if _, err := ext.ExportBundle(out); err != nil {
		t.Fatalf("ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "database", "users.md")); err != nil {
		t.Fatalf("database/users.md not written: %v", err)
	}

	if err := os.WriteFile(sqlFile, []byte("CREATE TABLE accounts (\n  id INT\n);"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ext.ExportBundle(out); err != nil {
		t.Fatalf("second ExportBundle() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "database", "users.md")); !os.IsNotExist(err) {
		t.Errorf("database/users.md should have been pruned after the table was dropped, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "database", "accounts.md")); err != nil {
		t.Errorf("database/accounts.md not written: %v", err)
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
