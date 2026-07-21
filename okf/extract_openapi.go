package okf

// OpenAPIExtractor parses an OpenAPI 3.0 / Swagger spec (JSON or YAML) into
// one API concept per operation.
type OpenAPIExtractor struct {
	SpecPath string
}

var _ Extractor = OpenAPIExtractor{}

func (e OpenAPIExtractor) ExtractConcepts() (map[string]Document, error) {
	panic("not implemented")
}

func (e OpenAPIExtractor) ExportBundle(bundleRoot string) (int, error) {
	panic("not implemented")
}
