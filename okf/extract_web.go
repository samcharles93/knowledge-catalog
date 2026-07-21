package okf

import (
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"
)

// WebExtractor fetches a set of URLs and converts each into a Reference
// concept. Fetch defaults to FetchPage; tests may override it to avoid
// real network access.
type WebExtractor struct {
	URLs  []string
	Fetch func(url string) (Page, error)
}

var _ Extractor = WebExtractor{}

func (e WebExtractor) ExtractConcepts() (map[string]Document, error) {
	fetch := e.Fetch
	if fetch == nil {
		fetch = FetchPage
	}
	now := time.Now().UTC().Format(time.RFC3339)
	concepts := map[string]Document{}

	for _, u := range e.URLs {
		page, err := fetch(u)
		if err != nil {
			// A single unreachable URL shouldn't fail the whole harvest.
			continue
		}

		title := page.Title
		if title == "" {
			title = u
		}

		slug := u
		if parsed, err := url.Parse(u); err == nil {
			slug = parsed.Host + strings.TrimSuffix(parsed.Path, "/")
		}
		conceptID := "references/" + slugify(slug)

		concepts[conceptID] = Document{
			Frontmatter: map[string]any{
				"type":        "Reference",
				"title":       title,
				"description": "Reference documentation from " + u,
				"resource":    u,
				"tags":        []string{"web", "reference"},
				"timestamp":   now,
			},
			Body: fmt.Sprintf("# %s\n\n**Source**: [%s](%s)\n\n%s\n", title, u, u, strings.TrimSpace(page.Markdown)),
		}
	}

	return concepts, nil
}

// ExportBundle never prunes: web harvests are additive by design (users
// accumulate --url flags across separate runs), so a run over a different
// URL set must not delete references/ concepts an earlier run produced.
func (e WebExtractor) ExportBundle(bundleRoot string) (int, error) {
	return exportBundle(bundleRoot, e.ExtractConcepts, nil)
}

func slugify(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, s)
}
