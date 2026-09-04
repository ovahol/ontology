package ontology

// rowResolver turns one inventory row (keyed by the workbook's canonical
// column names) into a ResolvedRow, given a taxonomy. It is the seam that lets
// the workbook migration path be catalog-aware: the catalog variant reconciles
// each row against a device dictionary (MappingSource="catalog_exact" /
// "catalog_fuzzy") before falling back to taxonomy rules, while the taxonomy
// variant uses rules only. Both produce the same ResolvedRow shape, so every
// sheet-builder consumes them identically.
type rowResolver func(row map[string]string, tax *Taxonomy) ResolvedRow

// taxonomyRowResolver resolves a row by taxonomy rules alone.
func taxonomyRowResolver(row map[string]string, tax *Taxonomy) ResolvedRow {
	return ResolveRowNamingFor(row, tax)
}

// makeCatalogRowResolver returns a row resolver that reconciles each row
// against cat first (exact, then typo-tolerant), falling back to taxonomy
// inference on a miss. A nil cat behaves exactly like the taxonomy resolver.
func makeCatalogRowResolver(cat Catalog) rowResolver {
	return func(row map[string]string, tax *Taxonomy) ResolvedRow {
		if cat == nil {
			return ResolveRowNamingFor(row, tax)
		}
		res := NormalizeWithCatalogAndTaxonomy(Input{
			DeviceName: row["Legacy source name"],
			SourceType: row["Source device type"],
			EMDNCode:   row["EMDN code"],
			EMDNTerm:   row["EMDN term"],
		}, cat, tax)
		return ResolvedRow{
			Fields:        res.Fields,
			Name:          res.Name,
			CanonicalName: res.Name,
			NamingSource:  res.MappingSource,
		}
	}
}
