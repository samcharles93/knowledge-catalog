package okf

// WebExtractor fetches a set of URLs and converts each into a Reference
// concept. Fetch defaults to FetchPage; tests may override it to avoid
// real network access.
type WebExtractor struct {
	URLs  []string
	Fetch func(url string) (Page, error)
}

var _ Extractor = WebExtractor{}

func (e WebExtractor) ExtractConcepts() (map[string]Document, error) {
	panic("not implemented")
}

func (e WebExtractor) ExportBundle(bundleRoot string) (int, error) {
	panic("not implemented")
}
