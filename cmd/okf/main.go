// Command okf is the Open Knowledge Format toolkit CLI.
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/samcharles93/knowledge-catalog/okf"
)

// version, commit, and date are set via -ldflags at build time (see
// .goreleaser.yaml); they default to "dev"/"none"/"unknown" for `go build`
// and `go run` during development.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

const usage = "usage: okf <init|validate|harvest|context|visualize|mcp|version> [flags]"

// run executes a single CLI invocation and returns the process exit code:
// 0 on success, 1 when a validation-type command reports failure, 2 for
// usage errors (unknown command or bad flags).
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(stderr, usage)
		return 2
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], stderr)
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "harvest":
		return runHarvest(args[1:], stderr)
	case "context":
		return runContext(args[1:], stdout, stderr)
	case "visualize":
		return runVisualize(args[1:], stderr)
	case "mcp":
		return runMCP(args[1:], stderr)
	case "version":
		_, _ = fmt.Fprintf(stdout, "okf %s (commit %s, built %s)\n", version, commit, date)
		return 0
	default:
		_, _ = fmt.Fprintln(stderr, usage)
		return 2
	}
}

func newFlagSet(name string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	return fs
}

func runInit(args []string, stderr io.Writer) int {
	fs := newFlagSet("init", stderr)
	path := fs.String("path", ".okf", "Target bundle directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	abs, err := initBundle(*path)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprintf(stderr, "Initialized OKF Knowledge Bundle at: %s\n", abs)
	return 0
}

// initBundle scaffolds a starter bundle: architecture/services/rules
// directories, a default config.yaml, and a sample architecture concept,
// then regenerates indexes. Existing files are left untouched.
func initBundle(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	for _, dir := range []string{"architecture", "services", "rules"} {
		if err := os.MkdirAll(filepath.Join(abs, dir), 0o755); err != nil {
			return "", err
		}
	}

	configPath := filepath.Join(abs, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, []byte("name: project-knowledge\nversion: '1.0'\n"), 0o644); err != nil {
			return "", err
		}
	}

	overviewPath := filepath.Join(abs, "architecture", "system-overview.md")
	if _, err := os.Stat(overviewPath); os.IsNotExist(err) {
		doc := okf.Document{
			Frontmatter: map[string]any{
				"type":        "Architecture",
				"title":       "System Overview",
				"description": "High level system components and architecture principles.",
				"tags":        []string{"overview", "architecture"},
			},
			Body: "# System Overview\n\nWelcome to the project OKF Knowledge Base.\n",
		}
		if err := os.WriteFile(overviewPath, []byte(doc.String()), 0o644); err != nil {
			return "", err
		}
	}

	if _, err := okf.RegenerateIndexes(abs); err != nil {
		return "", err
	}
	return abs, nil
}

func runValidate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("validate", stderr)
	bundle := fs.String("bundle", ".okf", "Path to bundle root")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	report := okf.ValidateBundle(*bundle)
	_, _ = fmt.Fprintf(stdout, "Validation Report for %s:\n", *bundle)
	_, _ = fmt.Fprintf(stdout, "- Total Concepts: %d\n", report.TotalConcepts)
	_, _ = fmt.Fprintf(stdout, "- Valid: %t\n", report.Valid())
	if len(report.Errors) > 0 {
		_, _ = fmt.Fprintln(stdout, "\nErrors:")
		for _, e := range report.Errors {
			_, _ = fmt.Fprintf(stdout, "  [%s] %s\n", e.ConceptID, e.Message)
		}
	}
	if len(report.Warnings) > 0 {
		_, _ = fmt.Fprintln(stdout, "\nWarnings:")
		for _, w := range report.Warnings {
			_, _ = fmt.Fprintf(stdout, "  [%s] %s\n", w.ConceptID, w.Message)
		}
	}

	if !report.Valid() {
		return 1
	}
	return 0
}

// stringSliceFlag implements flag.Value to support a repeatable flag, e.g.
// --url a --url b, mirroring argparse's action="append".
type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, ",") }
func (s *stringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func runHarvest(args []string, stderr io.Writer) int {
	fs := newFlagSet("harvest", stderr)
	typ := fs.String("type", "", "Extractor type: codebase, openapi, sql, web")
	src := fs.String("src", "", "Source file or directory path")
	out := fs.String("out", ".okf", "Output bundle root directory")
	var urls stringSliceFlag
	fs.Var(&urls, "url", "URL for web harvester (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	var extractor okf.Extractor
	switch *typ {
	case "codebase":
		root := *src
		if root == "" {
			root = "."
		}
		extractor = okf.CodebaseExtractor{ProjectRoot: root}
	case "openapi":
		extractor = okf.OpenAPIExtractor{SpecPath: *src}
	case "sql":
		extractor = okf.SQLExtractor{SQLPath: *src}
	case "web":
		extractor = okf.WebExtractor{URLs: urls}
	default:
		_, _ = fmt.Fprintf(stderr, "unknown harvest type %q; want codebase, openapi, sql, or web\n", *typ)
		return 2
	}

	n, err := extractor.ExportBundle(*out)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprintf(stderr, "Harvested %d concepts into %s\n", n, *out)
	return 0
}

func runContext(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("context", stderr)
	bundle := fs.String("bundle", ".okf", "Path to bundle root")
	concept := fs.String("concept", "", "Specific concept ID to fetch")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	var text string
	var err error
	if *concept != "" {
		text, err = okf.ConceptContext(*bundle, *concept)
	} else {
		text, err = okf.SummaryContext(*bundle)
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprintln(stdout, text)
	return 0
}

func runVisualize(args []string, stderr io.Writer) int {
	fs := newFlagSet("visualize", stderr)
	bundle := fs.String("bundle", ".okf", "Path to bundle root")
	out := fs.String("out", "", "Output HTML path (default: <bundle>/viz.html)")
	name := fs.String("name", "", "Bundle display name")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	outPath := *out
	if outPath == "" {
		outPath = filepath.Join(*bundle, "viz.html")
	}
	stats, err := okf.GenerateVisualization(*bundle, outPath, *name)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	_, _ = fmt.Fprintf(stderr, "Wrote %d concept(s), %d edge(s) -> %s\n", stats.Concepts, stats.Edges, outPath)
	return 0
}

func runMCP(args []string, stderr io.Writer) int {
	fs := newFlagSet("mcp", stderr)
	bundle := fs.String("bundle", ".okf", "Path to bundle root")
	addr := fs.String("addr", ":8080", "HTTP listen address for the MCP Streamable HTTP server")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	srv := okf.NewMCPServer(*bundle)
	_, _ = fmt.Fprintf(stderr, "Serving OKF MCP server for %s on http://%s\n", *bundle, *addr)
	if err := http.ListenAndServe(*addr, srv); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
