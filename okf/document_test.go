package okf

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseDocumentRoundTrip(t *testing.T) {
	t.Parallel()

	source := "---\ntype: Service\ntitle: User Service\ndescription: User authentication and management.\ntags: [auth, users]\ntimestamp: 2026-07-22T00:00:00Z\n---\n\n# User Service\n\nService body details.\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if got := doc.Frontmatter["type"]; got != "Service" {
		t.Errorf("type = %v, want Service", got)
	}
	if got := doc.Frontmatter["tags"]; !reflect.DeepEqual(got, []string{"auth", "users"}) {
		t.Errorf("tags = %#v", got)
	}
	if !strings.HasPrefix(doc.Body, "# User Service") {
		t.Errorf("body = %q", doc.Body)
	}

	reparsed, err := ParseDocument(doc.String())
	if err != nil {
		t.Fatalf("ParseDocument(String()) error = %v", err)
	}
	if !reflect.DeepEqual(reparsed.Frontmatter, doc.Frontmatter) {
		t.Errorf("frontmatter after round trip = %#v, want %#v", reparsed.Frontmatter, doc.Frontmatter)
	}
	if strings.TrimSpace(reparsed.Body) != strings.TrimSpace(doc.Body) {
		t.Errorf("body after round trip = %q, want %q", reparsed.Body, doc.Body)
	}
}

func TestParseDocumentWithoutFrontmatter(t *testing.T) {
	t.Parallel()

	source := "# Heading\n\nNo frontmatter here.\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if len(doc.Frontmatter) != 0 || doc.Body != source {
		t.Errorf("ParseDocument() = %#v", doc)
	}
}

func TestParseDocumentSupportsQuotedScalarsAndInlineLists(t *testing.T) {
	t.Parallel()

	doc, err := ParseDocument("---\ntype: 'Rule'\ntitle: \"Security: baseline\"\ntags: [\"security\", 'must-have']\n---\nbody\n")
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	want := map[string]any{
		"type":  "Rule",
		"title": "Security: baseline",
		"tags":  []string{"security", "must-have"},
	}
	if !reflect.DeepEqual(doc.Frontmatter, want) {
		t.Errorf("frontmatter = %#v, want %#v", doc.Frontmatter, want)
	}
}

func TestDocumentValidate(t *testing.T) {
	t.Parallel()

	doc := Document{Frontmatter: map[string]any{"title": "Missing Type"}}
	if err := doc.Validate(); err == nil || !strings.Contains(err.Error(), "type") {
		t.Fatalf("Validate() error = %v, want missing type", err)
	}

	doc.Frontmatter["type"] = "Architecture"
	warnings := doc.Warnings()
	if len(warnings) != 1 || !strings.Contains(warnings[0], "description") {
		t.Fatalf("Warnings() = %v", warnings)
	}
}

// Self-healing parsing: OKF files may be poorly maintained or hand-edited.
// Malformed frontmatter must never fail a parse — it degrades to "no
// frontmatter", with the original text preserved verbatim as the body, so
// no content is ever silently dropped or lost behind an error.

func TestParseDocumentSelfHealsUnterminatedFrontmatter(t *testing.T) {
	t.Parallel()

	source := "---\ntype: Service\nstill in frontmatter\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v, want self-healed nil error", err)
	}
	if len(doc.Frontmatter) != 0 {
		t.Errorf("Frontmatter = %#v, want empty", doc.Frontmatter)
	}
	if doc.Body != source {
		t.Errorf("Body = %q, want original source preserved verbatim", doc.Body)
	}
}

func TestParseDocumentSelfHealsNonMappingFrontmatter(t *testing.T) {
	t.Parallel()

	source := "---\n- Service\n- API\n---\n\n# Invalid\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v, want self-healed nil error", err)
	}
	if len(doc.Frontmatter) != 0 {
		t.Errorf("Frontmatter = %#v, want empty", doc.Frontmatter)
	}
	if doc.Body != source {
		t.Errorf("Body = %q, want original source preserved verbatim", doc.Body)
	}
}

func TestParseDocumentSelfHealsInvalidYAML(t *testing.T) {
	t.Parallel()

	source := "---\ntype: [unterminated\n---\nbody\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v, want self-healed nil error", err)
	}
	if len(doc.Frontmatter) != 0 {
		t.Errorf("Frontmatter = %#v, want empty", doc.Frontmatter)
	}
	if doc.Body != source {
		t.Errorf("Body = %q, want original source preserved verbatim", doc.Body)
	}
}

func TestParseDocumentToleratesWhitespaceAroundDelimiters(t *testing.T) {
	t.Parallel()

	source := "  ---  \ntype: Service\n  ---  \n\nbody\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if got := doc.Frontmatter["type"]; got != "Service" {
		t.Errorf("type = %v, want Service (delimiter padding should be tolerated)", got)
	}
}

func TestParseDocumentThenStringRoundTripsSelfHealedInput(t *testing.T) {
	t.Parallel()

	malformed := "---\ntype: Service\nstill in frontmatter\n"
	doc, err := ParseDocument(malformed)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	serialized := doc.String()
	reparsed, err := ParseDocument(serialized)
	if err != nil {
		t.Fatalf("ParseDocument(String()) error = %v", err)
	}
	if !reflect.DeepEqual(reparsed, doc) {
		t.Errorf("self-healed round trip not idempotent: got %#v, want %#v", reparsed, doc)
	}
}

// Self-healing serialization: String() always produces well-formed output
// regardless of what shape the Document is in — missing trailing newlines,
// nil/empty frontmatter, or empty bodies are all normalized rather than
// surfaced as errors.

func TestDocumentStringAddsMissingTrailingNewline(t *testing.T) {
	t.Parallel()

	doc := Document{Frontmatter: map[string]any{"type": "Concept"}, Body: "body"}
	if got := doc.String(); !strings.HasSuffix(got, "body\n") {
		t.Errorf("String() = %q, want trailing newline", got)
	}
}

func TestDocumentStringDoesNotDoubleTrailingNewline(t *testing.T) {
	t.Parallel()

	doc := Document{Frontmatter: map[string]any{"type": "Concept"}, Body: "body\n"}
	if got := doc.String(); strings.HasSuffix(got, "body\n\n") {
		t.Errorf("String() = %q, want exactly one trailing newline", got)
	}
}

func TestDocumentStringOmitsFrontmatterBlockWhenEmpty(t *testing.T) {
	t.Parallel()

	doc := Document{Body: "just a plain markdown file\n"}
	got := doc.String()
	if strings.Contains(got, "---") {
		t.Errorf("String() = %q, want no frontmatter delimiters for a document with no frontmatter", got)
	}
	if got != doc.Body {
		t.Errorf("String() = %q, want body returned as-is", got)
	}
}

func TestDocumentStringHandlesEmptyBody(t *testing.T) {
	t.Parallel()

	doc := Document{Frontmatter: map[string]any{"type": "Concept"}, Body: ""}
	got := doc.String()
	if !strings.HasSuffix(got, "---\n\n") {
		t.Errorf("String() = %q, want frontmatter block with empty body and no dangling body newline", got)
	}
}

func TestParseDocumentThenStringPreservesFrontmatterlessRoundTrip(t *testing.T) {
	t.Parallel()

	source := "# Heading\n\nNo frontmatter here.\n"
	doc, err := ParseDocument(source)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}
	if got := doc.String(); got != source {
		t.Errorf("String() = %q, want %q (no frontmatter block invented)", got, source)
	}
}
