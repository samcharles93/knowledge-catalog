package okf

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPListConcepts renders a one-line-per-concept listing (id, type, title,
// description) for every concept in the bundle, excluding index/log files.
func MCPListConcepts(bundleRoot string) (string, error) {
	var paths []string
	err := filepath.WalkDir(bundleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") || reservedConceptFileNames[d.Name()] {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(paths)

	if len(paths) == 0 {
		return "No concepts found.", nil
	}

	lines := make([]string, 0, len(paths))
	for _, path := range paths {
		id, err := ConceptID(bundleRoot, path)
		if err != nil {
			continue
		}
		relID := strings.Join(id, "/")

		data, err := os.ReadFile(path)
		if err != nil {
			lines = append(lines, fmt.Sprintf("- `%s` (Parse Error)", relID))
			continue
		}
		doc, err := ParseDocument(string(data))
		if err != nil {
			lines = append(lines, fmt.Sprintf("- `%s` (Parse Error)", relID))
			continue
		}

		fm := doc.Frontmatter
		typ := stringFrontmatterOr(fm["type"], "Concept")
		title := stringFrontmatterOr(fm["title"], relID)
		desc := stringFrontmatterOr(fm["description"], "")
		lines = append(lines, fmt.Sprintf("- `%s` [%s]: %s - %s", relID, typ, title, desc))
	}
	return strings.Join(lines, "\n"), nil
}

// MCPGetConcept renders the full prompt context for a single concept ID.
func MCPGetConcept(bundleRoot string, conceptID string) (string, error) {
	return ConceptContext(bundleRoot, conceptID)
}

// MCPGetContext renders the bundle's progressive-disclosure summary.
func MCPGetContext(bundleRoot string) (string, error) {
	return SummaryContext(bundleRoot)
}

// MCPValidateBundle renders a human-readable validation report summary.
func MCPValidateBundle(bundleRoot string) (string, error) {
	report := ValidateBundle(bundleRoot)
	lines := []string{
		fmt.Sprintf("Bundle Valid: %t", report.Valid()),
		fmt.Sprintf("Total Concepts: %d", report.TotalConcepts),
	}
	if len(report.Errors) > 0 {
		lines = append(lines, "\nErrors:")
		for _, e := range report.Errors {
			lines = append(lines, fmt.Sprintf("- [%s] %s", e.ConceptID, e.Message))
		}
	}
	if len(report.Warnings) > 0 {
		lines = append(lines, "\nWarnings:")
		for _, w := range report.Warnings {
			lines = append(lines, fmt.Sprintf("- [%s] %s", w.ConceptID, w.Message))
		}
	}
	return strings.Join(lines, "\n"), nil
}

// MCPServer exposes an OKF bundle over the MCP Streamable HTTP transport,
// via the official github.com/modelcontextprotocol/go-sdk — no stdio
// transport is supported. It serves both the four okf_* tools and, as MCP
// resources, every concept document in the bundle so hosts can browse them
// directly rather than only reaching them through a tool call.
type MCPServer struct {
	BundleRoot string
	handler    *sdk.StreamableHTTPHandler
}

var _ http.Handler = (*MCPServer)(nil)

// NewMCPServer constructs a server rooted at bundleRoot. The bundle is
// rescanned for every new session, so tools and resources always reflect
// the current on-disk contents.
func NewMCPServer(bundleRoot string) *MCPServer {
	s := &MCPServer{BundleRoot: bundleRoot}
	s.handler = sdk.NewStreamableHTTPHandler(func(*http.Request) *sdk.Server {
		return newBundleServer(bundleRoot)
	}, nil)
	return s
}

// ServeHTTP delegates to the SDK's Streamable HTTP handler, which manages
// session IDs, protocol version negotiation, and the JSON/SSE response
// modes per the MCP spec.
func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

type getConceptArgs struct {
	ConceptID string `json:"concept_id" jsonschema:"Concept ID relative to bundle root"`
}

func newBundleServer(bundleRoot string) *sdk.Server {
	server := sdk.NewServer(&sdk.Implementation{Name: "okf-knowledge-server", Version: "0.1.0"}, nil)

	sdk.AddTool(server, &sdk.Tool{
		Name:        "okf_list_concepts",
		Description: "List all knowledge concepts in the OKF bundle with their type, title, and metadata.",
	}, func(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		text, err := MCPListConcepts(bundleRoot)
		if err != nil {
			return nil, nil, err
		}
		return textResult(text), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "okf_get_concept",
		Description: "Get full content and metadata for a specific concept ID (e.g., 'architecture/system-overview').",
	}, func(_ context.Context, _ *sdk.CallToolRequest, args getConceptArgs) (*sdk.CallToolResult, any, error) {
		text, err := MCPGetConcept(bundleRoot, args.ConceptID)
		if err != nil {
			return nil, nil, err
		}
		return textResult(text), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "okf_get_context",
		Description: "Get summary prompt context for progressive disclosure in coding agents.",
	}, func(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		text, err := MCPGetContext(bundleRoot)
		if err != nil {
			return nil, nil, err
		}
		return textResult(text), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "okf_validate_bundle",
		Description: "Validate OKF bundle for structural integrity, frontmatter schema compliance, and link health.",
	}, func(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		text, err := MCPValidateBundle(bundleRoot)
		if err != nil {
			return nil, nil, err
		}
		return textResult(text), nil, nil
	})

	for _, c := range discoverConcepts(bundleRoot) {
		conceptID := c.ID
		server.AddResource(&sdk.Resource{
			URI:         "okf:///" + conceptID,
			Name:        c.Title,
			Description: c.Description,
			MIMEType:    "text/markdown",
		}, func(_ context.Context, req *sdk.ReadResourceRequest) (*sdk.ReadResourceResult, error) {
			text, err := MCPGetConcept(bundleRoot, conceptID)
			if err != nil {
				return nil, err
			}
			return &sdk.ReadResourceResult{
				Contents: []*sdk.ResourceContents{{URI: req.Params.URI, MIMEType: "text/markdown", Text: text}},
			}, nil
		})
	}

	return server
}

func textResult(text string) *sdk.CallToolResult {
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: text}}}
}

type conceptSummary struct {
	ID          string
	Title       string
	Description string
}

// discoverConcepts walks bundleRoot for concept documents to register as
// MCP resources. It self-heals like the rest of the package: a missing
// bundle directory or an unreadable/malformed file just yields fewer
// entries rather than failing server startup.
func discoverConcepts(bundleRoot string) []conceptSummary {
	var paths []string
	_ = filepath.WalkDir(bundleRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") || reservedConceptFileNames[d.Name()] {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	sort.Strings(paths)

	summaries := make([]conceptSummary, 0, len(paths))
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

		summaries = append(summaries, conceptSummary{
			ID:          conceptID,
			Title:       stringFrontmatterOr(doc.Frontmatter["title"], conceptID),
			Description: stringFrontmatterOr(doc.Frontmatter["description"], ""),
		})
	}
	return summaries
}
