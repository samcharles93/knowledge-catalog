package okf

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPListConcepts renders a one-line-per-concept listing (id, type, title,
// description) for every concept in the bundle, excluding index/log files.
func MCPListConcepts(bundleRoot string) (string, error) {
	files, err := walkConceptFiles(bundleRoot)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "No concepts found.", nil
	}

	lines := make([]string, 0, len(files))
	for _, f := range files {
		if f.Err != nil {
			lines = append(lines, fmt.Sprintf("- `%s` (Parse Error)", f.ID))
			continue
		}
		fm := f.Doc.Frontmatter
		typ := stringFrontmatterOr(fm["type"], "Concept")
		title := stringFrontmatterOr(fm["title"], f.ID)
		desc := stringFrontmatterOr(fm["description"], "")
		lines = append(lines, fmt.Sprintf("- `%s` [%s]: %s - %s", f.ID, typ, title, desc))
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

// MCPSearchConcepts renders a one-line-per-match listing of concepts whose
// title, body, or tags contain query (case-insensitive substring match).
func MCPSearchConcepts(bundleRoot string, query string) (string, error) {
	matches, err := SearchConcepts(bundleRoot, query)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "No matching concepts found.", nil
	}

	lines := make([]string, 0, len(matches))
	for _, f := range matches {
		fm := f.Doc.Frontmatter
		typ := stringFrontmatterOr(fm["type"], "Concept")
		title := stringFrontmatterOr(fm["title"], f.ID)
		desc := stringFrontmatterOr(fm["description"], "")
		lines = append(lines, fmt.Sprintf("- `%s` [%s]: %s - %s", f.ID, typ, title, desc))
	}
	return strings.Join(lines, "\n"), nil
}

// MCPServer exposes an OKF bundle over the MCP Streamable HTTP transport,
// via the official github.com/modelcontextprotocol/go-sdk — no stdio
// transport is supported. It serves the five okf_* tools defined by
// SPEC.md §6.2, and, as MCP resources, every concept document in the
// bundle so hosts can browse them directly rather than only reaching them
// through a tool call.
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
		return newBundleServer(s.BundleRoot)
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

type searchConceptsArgs struct {
	Query string `json:"query" jsonschema:"Case-insensitive substring to search for across concept tags, titles, and bodies"`
}

type rememberArgs struct {
	Type        string   `json:"type" jsonschema:"Concept type: Rule, Runbook, Concept, Service, Architecture, or a custom type. Codebase/API/Table/Reference are rejected -- those are owned by harvest and would be pruned."`
	Title       string   `json:"title" jsonschema:"Short human title; used to derive the concept ID"`
	Body        string   `json:"body" jsonschema:"Markdown body content of the memory"`
	Description string   `json:"description,omitempty" jsonschema:"Optional one-line description"`
	Resource    string   `json:"resource,omitempty" jsonschema:"Optional file path, PR link, or session reference"`
	Tags        []string `json:"tags,omitempty" jsonschema:"Optional list of tags"`
}

// MCPRemember writes a new free-form concept (rule, insight, runbook step,
// etc.) into the bundle and returns a short confirmation with the concept
// ID and a preview of what was written.
func MCPRemember(bundleRoot string, in RememberInput) (string, error) {
	conceptID, err := Remember(bundleRoot, in)
	if err != nil {
		return "", err
	}
	preview := in.Body
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	return fmt.Sprintf("Remembered `%s` [%s]: %s\n\n%s", conceptID, in.Type, in.Title, preview), nil
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
		Name:        "okf_validate",
		Description: "Validate OKF bundle for structural integrity, frontmatter schema compliance, and link health.",
	}, func(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, any, error) {
		text, err := MCPValidateBundle(bundleRoot)
		if err != nil {
			return nil, nil, err
		}
		return textResult(text), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "okf_search_concepts",
		Description: "Search concepts by a case-insensitive substring across tags, titles, and bodies.",
	}, func(_ context.Context, _ *sdk.CallToolRequest, args searchConceptsArgs) (*sdk.CallToolResult, any, error) {
		text, err := MCPSearchConcepts(bundleRoot, args.Query)
		if err != nil {
			return nil, nil, err
		}
		return textResult(text), nil, nil
	})

	sdk.AddTool(server, &sdk.Tool{
		Name:        "okf_remember",
		Description: "Capture a free-form memory (coding rule, session insight, runbook step, etc.) as a new validated OKF concept document.",
	}, func(_ context.Context, _ *sdk.CallToolRequest, args rememberArgs) (*sdk.CallToolResult, any, error) {
		text, err := MCPRemember(bundleRoot, RememberInput{
			Type: args.Type, Title: args.Title, Body: args.Body,
			Description: args.Description, Resource: args.Resource, Tags: args.Tags,
		})
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
	files, err := walkConceptFiles(bundleRoot)
	if err != nil {
		return nil
	}

	summaries := make([]conceptSummary, 0, len(files))
	for _, f := range files {
		if f.Err != nil {
			continue
		}
		summaries = append(summaries, conceptSummary{
			ID:          f.ID,
			Title:       stringFrontmatterOr(f.Doc.Frontmatter["title"], f.ID),
			Description: stringFrontmatterOr(f.Doc.Frontmatter["description"], ""),
		})
	}
	return summaries
}
