package okf

// Extractor turns some external, vendor-neutral source into OKF concept
// documents and can export them directly into a bundle.
type Extractor interface {
	// ExtractConcepts returns the concept documents keyed by concept ID.
	ExtractConcepts() (map[string]Document, error)
	// ExportBundle writes the extracted concepts into bundleRoot and
	// regenerates indexes, returning the number of concepts written.
	ExportBundle(bundleRoot string) (int, error)
}
