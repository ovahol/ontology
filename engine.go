package ontology

// Option configures an Engine instance.
type Option func(*Engine)

// WithCatalog sets the Catalog used for exact-match dictionary lookups.
func WithCatalog(cat Catalog) Option {
	return func(e *Engine) {
		e.catalog = cat
	}
}

// Engine is the system-agnostic ontology normalization and inference engine.
// It executes vendor-defined taxonomies, vocabularies, and rules.
type Engine struct {
	taxonomy *Taxonomy
	catalog  Catalog
}

// NewEngine creates a new Engine configured with the given vendor taxonomy and options.
func NewEngine(tax *Taxonomy, opts ...Option) *Engine {
	e := &Engine{
		taxonomy: tax,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Taxonomy returns the engine's loaded taxonomy definition.
func (e *Engine) Taxonomy() *Taxonomy {
	return e.taxonomy
}

// Catalog returns the engine's catalog if configured.
func (e *Engine) Catalog() Catalog {
	return e.catalog
}

// Normalize normalizes a single device input using the engine's taxonomy and catalog.
func (e *Engine) Normalize(in Input) Result {
	if e.catalog != nil {
		return NormalizeWithCatalogAndTaxonomy(in, e.catalog, e.taxonomy)
	}
	return NormalizeWithTaxonomy(in, e.taxonomy)
}

// NormalizeBatch normalizes multiple device inputs.
func (e *Engine) NormalizeBatch(inputs []Input) []Result {
	if e.catalog != nil {
		return NormalizeBatchWithCatalogAndTaxonomy(inputs, e.catalog, e.taxonomy)
	}
	return NormalizeBatchWithTaxonomy(inputs, e.taxonomy)
}

// NormalizeJSON normalizes JSON input (single object or array).
func (e *Engine) NormalizeJSON(data []byte) ([]Result, error) {
	return NormalizeJSONWithTaxonomy(data, e.taxonomy)
}

// NormalizeJSONFile reads and normalizes a JSON file of inputs.
func (e *Engine) NormalizeJSONFile(path string) ([]Result, error) {
	return NormalizeJSONFileWithTaxonomy(path, e.taxonomy)
}

// NormalizeWorkbook normalizes an Excel spreadsheet using the engine's taxonomy.
func (e *Engine) NormalizeWorkbook(inputPath, outputPath string) (string, error) {
	return NormalizeWorkbookWithTaxonomy(inputPath, outputPath, e.taxonomy)
}

// NormalizeCSV normalizes a CSV file using the engine's taxonomy.
func (e *Engine) NormalizeCSV(path string) ([]Result, error) {
	return NormalizeCSVWithTaxonomy(path, e.taxonomy)
}
