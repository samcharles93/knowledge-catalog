package okf

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SQLExtractor parses CREATE TABLE statements out of a SQL DDL file into
// one Table concept per table.
type SQLExtractor struct {
	SQLPath string
}

var _ Extractor = SQLExtractor{}

var sqlCreateTableRE = regexp.MustCompile(
	"(?is)CREATE\\s+TABLE\\s+(?:IF\\s+NOT\\s+EXISTS\\s+)?([`\"\\w.]+)\\s*\\((.*?)\\);",
)

var sqlSkipLinePrefixes = []string{"--", "PRIMARY", "CONSTRAINT", "FOREIGN", "KEY", "INDEX", ")"}

func (e SQLExtractor) ExtractConcepts() (map[string]Document, error) {
	data, err := os.ReadFile(e.SQLPath)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)

	concepts := map[string]Document{}
	for _, match := range sqlCreateTableRE.FindAllStringSubmatch(string(data), -1) {
		rawName, body := match[1], match[2]

		// Split on "." before trimming quotes/backticks, not after — a
		// naive strip-then-split on a schema-qualified, quoted identifier
		// like "public"."products" leaves a stray quote stuck to the table
		// name (the inner quotes around the dot survive a strip that only
		// touches the outer edges of the whole string).
		segments := strings.Split(rawName, ".")
		cleanName := strings.Trim(segments[len(segments)-1], "`\"")

		var colLines []string
		for _, line := range splitLines(body) {
			line = strings.TrimSpace(line)
			if line == "" || hasAnyPrefix(line, sqlSkipLinePrefixes) {
				continue
			}
			colName := ""
			if fields := strings.Fields(line); len(fields) > 0 {
				colName = fields[0]
			}
			colLines = append(colLines, fmt.Sprintf("| `%s` | `%s` |", colName, line))
		}

		tableBody := fmt.Sprintf(
			"# Table `%s`\n\nExtracted from SQL DDL `%s`.\n\n## Columns\n\n| Column | Full Definition |\n|---|---|\n%s\n",
			cleanName, filepath.Base(e.SQLPath), strings.Join(colLines, "\n"),
		)

		concepts["database/"+cleanName] = Document{
			Frontmatter: map[string]any{
				"type":        "Table",
				"title":       cleanName,
				"description": fmt.Sprintf("Database table %s.", cleanName),
				"resource":    e.SQLPath,
				"tags":        []string{"database", "table", "sql"},
				"timestamp":   now,
			},
			Body: tableBody,
		}
	}
	return concepts, nil
}

// ExportBundle fully owns the database/ namespace for this schema file:
// re-harvesting after a table was dropped or renamed prunes its stale
// concept file.
func (e SQLExtractor) ExportBundle(bundleRoot string) (int, error) {
	return exportBundle(bundleRoot, e.ExtractConcepts, []string{"database"})
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
