package okf

// CodebaseExtractor scans a local source tree and produces an architecture
// overview concept plus one Module concept per recognized source file.
type CodebaseExtractor struct {
	ProjectRoot string
}

var _ Extractor = CodebaseExtractor{}

func (e CodebaseExtractor) ExtractConcepts() (map[string]Document, error) {
	panic("not implemented")
}

func (e CodebaseExtractor) ExportBundle(bundleRoot string) (int, error) {
	panic("not implemented")
}
