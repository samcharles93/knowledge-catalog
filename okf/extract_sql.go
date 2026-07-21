package okf

// SQLExtractor parses CREATE TABLE statements out of a SQL DDL file into
// one Table concept per table.
type SQLExtractor struct {
	SQLPath string
}

var _ Extractor = SQLExtractor{}

func (e SQLExtractor) ExtractConcepts() (map[string]Document, error) {
	panic("not implemented")
}

func (e SQLExtractor) ExportBundle(bundleRoot string) (int, error) {
	panic("not implemented")
}
