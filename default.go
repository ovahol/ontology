package ontology

import _ "embed"

//go:embed examples/taxonomies/medevis.json
var defaultTaxonomyJSON []byte

// DefaultTaxonomy returns the built-in reference taxonomy used when a caller
// does not supply one. It is the WHO-derived MeDevIS vocabulary. ontology is
// an engine: vendors should provide their own taxonomy via LoadTaxonomy /
// LoadTaxonomyFile. This default exists only so nil-taxonomy callers still
// produce useful, standards-aligned results and is not a vendor-specific
// vocabulary.
func DefaultTaxonomy() *Taxonomy {
	t, err := LoadTaxonomy(defaultTaxonomyJSON)
	if err != nil {
		panic("ontology: failed to load embedded default taxonomy: " + err.Error())
	}
	return t
}
