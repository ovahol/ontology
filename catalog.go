package ontology

import "strings"

// CatalogEntry is a single known device from the host system's dictionary
// (e.g. Ovahol's core_public.device). It is the system-agnostic view
// ontology needs to do exact-match resolution without vendoring the full
// dictionary into FamilyRules.
//
// Values are the same controlled vocabulary as Result: DeviceType maps to
// lookup.device_type.name, DeviceCategory to lookup.device_category.name,
// DeviceFunction to lookup.device_function.name, etc.
type CatalogEntry struct {
	Name                  string `json:"name"`
	DeviceType            string `json:"device_type"`
	DeviceCategory        string `json:"device_category,omitempty"`
	DeviceFunction        string `json:"device_function"`
	DeviceApplicationRisk string `json:"device_application_risk"`
	EMDNCode              string `json:"emdn_code,omitempty"`
	EMDNTerm              string `json:"emdn_term,omitempty"`

	// ID is opaque — Ovahol can store its device UUID here so the caller
	// can carry it through without ontology needing to understand it.
	ID string `json:"id,omitempty"`
}

// Catalog is the host system's read-only device dictionary. Ontology calls
// Find once per input; if it returns an entry ontology returns it verbatim
// (MappingSource="catalog_exact", Confidence="high") instead of running
// keyword inference. If nil is passed or Find returns !ok, ontology falls
// back to taxonomy inference.
//
// This keeps ontology system-agnostic and small: Ovahol implements this
// with a single SELECT, other systems can pass nil or an in-memory map.
// No FamilyRule vendoring required.
//
// Ovahol example (DB-backed, not vendored):
//
//	type dbCatalog struct{ q *Queries }
//	func (c *dbCatalog) Find(in ontology.Input) (*ontology.CatalogEntry, bool) {
//	    row, err := c.q.GetActiveDeviceByName(ctx, in.DeviceName)
//	    if err != nil { return nil, false }
//	    return &ontology.CatalogEntry{
//	        Name: row.Name,
//	        DeviceType: row.DeviceTypeName.String,
//	        DeviceCategory: row.DeviceCategoryName.String, // via device_function→category
//	        DeviceFunction: row.DeviceFunctionName.String,
//	        DeviceApplicationRisk: row.DeviceApplicationRisk.String,
//	        ID: row.ID.String(),
//	    }, true
//	}
//	// then: ontology.NormalizeWithCatalog(input, &dbCatalog{q})
type Catalog interface {
	Find(input Input) (*CatalogEntry, bool)
}

// InMemoryCatalog is a trivial Catalog backed by a slice. Useful for tests
// and for offline migrators that snapshot core_public.device into JSON.
type InMemoryCatalog struct {
	entries []CatalogEntry
	index   map[string]int
}

// NewInMemoryCatalog builds a catalog from entries. Names are indexed after
// Normalized() + lowercasing so "ECG machine" matches "ecg  machine".
func NewInMemoryCatalog(entries []CatalogEntry) *InMemoryCatalog {
	c := &InMemoryCatalog{
		entries: entries,
		index:   make(map[string]int, len(entries)*2),
	}
	for i, e := range entries {
		for _, key := range catalogKeys(e) {
			nk := strings.ToLower(Normalized(key))
			if nk == "" {
				continue
			}
			if _, ok := c.index[nk]; !ok {
				c.index[nk] = i
			}
		}
		if ek := strings.ToLower(Normalized(e.EMDNCode)); ek != "" {
			if _, ok := c.index["emdn:"+ek]; !ok {
				c.index["emdn:"+ek] = i
			}
		}
	}
	return c
}

func catalogKeys(e CatalogEntry) []string {
	return []string{e.Name}
}

// Find implements Catalog.
func (c *InMemoryCatalog) Find(input Input) (*CatalogEntry, bool) {
	if c == nil || len(c.entries) == 0 {
		return nil, false
	}
	if code := strings.ToLower(Normalized(input.EMDNCode)); code != "" {
		if idx, ok := c.index["emdn:"+code]; ok {
			e := c.entries[idx]
			return &e, true
		}
	}
	if dn := strings.ToLower(Normalized(input.DeviceName)); dn != "" {
		if idx, ok := c.index[dn]; ok {
			e := c.entries[idx]
			return &e, true
		}
	}
	if term := strings.ToLower(Normalized(input.EMDNTerm)); term != "" {
		if idx, ok := c.index[term]; ok {
			e := c.entries[idx]
			return &e, true
		}
	}
	return nil, false
}

// NormalizeWithCatalog is like Normalize but consults cat first.
func NormalizeWithCatalog(input Input, cat Catalog) Result {
	if cat != nil {
		if entry, ok := cat.Find(input); ok && entry != nil {
			category := entry.DeviceCategory
			if category == "" {
				category = CategoryForFunction(entry.DeviceFunction)
			}
			return Result{
				Name:                  entry.Name,
				DeviceType:            entry.DeviceType,
				DeviceCategory:        category,
				DeviceFunction:        entry.DeviceFunction,
				DeviceApplication:     entry.DeviceFunction,
				DeviceApplicationRisk: entry.DeviceApplicationRisk,
				LegacySourceName:      input.DeviceName,
				SourceType:            input.SourceType,
				EMDNCode:              firstNonEmpty(input.EMDNCode, entry.EMDNCode),
				EMDNTerm:              firstNonEmpty(input.EMDNTerm, entry.EMDNTerm),
				MappingSource:         "catalog_exact",
				Confidence:            "high",
			}
		}
	}
	return Normalize(input)
}

// NormalizeBatchWithCatalog is the batch analogue.
func NormalizeBatchWithCatalog(inputs []Input, cat Catalog) []Result {
	out := make([]Result, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, NormalizeWithCatalog(in, cat))
	}
	return out
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
