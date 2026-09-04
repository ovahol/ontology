package ontology

import (
	"encoding/csv"
	"io"
	"strings"
)

// CatalogEntry is a single known device from a host system's dictionary. It
// is the system-agnostic view ontology needs for exact-match resolution
// without vendoring the full dictionary.
//
// Fields carries the host system's controlled-vocabulary values keyed by
// whatever dimension keys the vendor uses (the same keys taxonomy.Fields and
// Result.Fields use). A vendor may use "device_type", "device_function",
// "device_application_risk", "device_category", or any other set — ontology
// treats none as special.
//
// Name and ID are lookup identity (the canonical dictionary name and an
// opaque host identifier). EMDNCode/EMDNTerm are standard international
// device identifiers kept as first-class lookup keys.
type CatalogEntry struct {
	Name   string            `json:"name"`
	ID     string            `json:"id,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`

	EMDNCode string `json:"emdn_code,omitempty"`
	EMDNTerm string `json:"emdn_term,omitempty"`
}

// GetField returns e.Fields[key], or "" when absent.
func (e CatalogEntry) GetField(key string) string {
	if e.Fields == nil {
		return ""
	}
	return e.Fields[key]
}

// DeviceType returns Fields[FieldDeviceType], if the vendor uses that key.
func (e CatalogEntry) DeviceType() string { return e.GetField(FieldDeviceType) }

// DeviceCategory returns Fields[FieldDeviceCategory], if the vendor uses that key.
func (e CatalogEntry) DeviceCategory() string { return e.GetField(FieldDeviceCategory) }

// DeviceFunction returns Fields[FieldDeviceFunction], if the vendor uses that key.
func (e CatalogEntry) DeviceFunction() string { return e.GetField(FieldDeviceFunction) }

// DeviceApplicationRisk returns Fields[FieldDeviceApplicationRisk], if the vendor uses that key.
func (e CatalogEntry) DeviceApplicationRisk() string { return e.GetField(FieldDeviceApplicationRisk) }

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
//	        ID:   row.ID.String(),
//	        Fields: map[string]string{
//	            ontology.FieldDeviceType:            row.DeviceTypeName.String,
//	            ontology.FieldDeviceCategory:        row.DeviceCategoryName.String,
//	            ontology.FieldDeviceFunction:        row.DeviceFunctionName.String,
//	            ontology.FieldDeviceApplicationRisk: row.DeviceApplicationRisk.String,
//	        },
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
	// Prefer the device name: it is the disambiguating key. EMDN codes are
	// not unique in practice (a single code is shared across several devices
	// with different type/function/risk), so a name that is present and known
	// must win over a colliding EMDN code.
	if dn := strings.ToLower(Normalized(input.DeviceName)); dn != "" {
		if idx, ok := c.index[dn]; ok {
			e := c.entries[idx]
			return &e, true
		}
	}
	if code := strings.ToLower(Normalized(input.EMDNCode)); code != "" {
		if idx, ok := c.index["emdn:"+code]; ok {
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
//
// Deprecated: use NormalizeWithCatalogAndTaxonomy.
func NormalizeWithCatalog(input Input, cat Catalog) Result {
	return NormalizeWithCatalogAndTaxonomy(input, cat, nil)
}

// NormalizeWithCatalogAndTaxonomy is the vocab-less catalog entry point.
// It returns the catalog entry verbatim when matched (MappingSource=
// "catalog_exact", Confidence="high"); otherwise it falls back to taxonomy
// inference.
func NormalizeWithCatalogAndTaxonomy(input Input, cat Catalog, tax *Taxonomy) Result {
	if cat != nil {
		if entry, ok := cat.Find(input); ok && entry != nil {
			// Start from the catalog's own fields (any vendor dimensions),
			// then fill the Ovahol-style deprecated fixed fields if present.
			fields := make(map[string]string, len(entry.Fields)+1)
			for k, v := range entry.Fields {
				fields[k] = v
			}
			// Derive category from function via the taxonomy rules (a generic
			// function->category derivation) when the entry did not carry it.
			if fields[FieldDeviceCategory] == "" {
				if cat := CategoryForFunctionFor(entry.DeviceFunction(), tax); cat != "" {
					fields[FieldDeviceCategory] = cat
				}
			}
			return Result{
				Name:                  entry.Name,
				Fields:                fields,
				DeviceType:            fields[FieldDeviceType],
				DeviceCategory:        fields[FieldDeviceCategory],
				DeviceFunction:        fields[FieldDeviceFunction],
				DeviceApplication:     fields[FieldDeviceFunction],
				DeviceApplicationRisk: fields[FieldDeviceApplicationRisk],
				LegacySourceName:      input.DeviceName,
				SourceType:            input.SourceType,
				EMDNCode:              firstNonEmpty(input.EMDNCode, entry.EMDNCode),
				EMDNTerm:              firstNonEmpty(input.EMDNTerm, entry.EMDNTerm),
				MappingSource:         "catalog_exact",
				Confidence:            "high",
			}
		}
	}
	return NormalizeWithTaxonomy(input, tax)
}

// NormalizeBatchWithCatalog is the batch analogue.
//
// Deprecated: use NormalizeBatchWithCatalogAndTaxonomy.
func NormalizeBatchWithCatalog(inputs []Input, cat Catalog) []Result {
	return NormalizeBatchWithCatalogAndTaxonomy(inputs, cat, nil)
}

// NormalizeBatchWithCatalogAndTaxonomy is the vocab-less batch entry point.
func NormalizeBatchWithCatalogAndTaxonomy(inputs []Input, cat Catalog, tax *Taxonomy) []Result {
	out := make([]Result, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, NormalizeWithCatalogAndTaxonomy(in, cat, tax))
	}
	return out
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// CSVCatalog is a catalog loaded from a device dictionary CSV. The host
// system maps its own column names onto fields, keeping ontology
// system-agnostic while letting a migrator replay its dictionary exactly.
//
// Columns maps the value to read for each key of interest -> the CSV column
// header that holds it. Three special keys are lookup identity and are stored
// on the entry directly: "name" (required), "id", and "emdn_code"/"emdn_term".
// Every other key is a controlled-vocabulary dimension and is stored in
// CatalogEntry.Fields under that same key — so a vendor using
// "device_type"/"device_function"/... reads naturally, and any other
// dimension key works too.
type CSVCatalog struct {
	Columns map[string]string
}

// Load builds an InMemoryCatalog from the CSV reader using the column-name
// mapping described on CSVCatalog.
func (c CSVCatalog) Load(r io.Reader) (*InMemoryCatalog, error) {
	crd := csv.NewReader(r)
	crd.FieldsPerRecord = -1
	recs, err := crd.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(recs) == 0 {
		return &InMemoryCatalog{index: map[string]int{}}, nil
	}
	header := map[string]int{}
	for i, h := range recs[0] {
		header[strings.TrimSpace(h)] = i
	}
	val := func(rec []string, field string) string {
		hdr, ok := c.Columns[field]
		if !ok {
			return ""
		}
		i, ok := header[hdr]
		if !ok || i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[i])
	}
	var entries []CatalogEntry
	for _, rec := range recs[1:] {
		name := val(rec, "name")
		if name == "" {
			continue
		}
		entry := CatalogEntry{
			Name:     name,
			ID:       val(rec, "id"),
			EMDNCode: val(rec, "emdn_code"),
			EMDNTerm: val(rec, "emdn_term"),
		}
		// Every non-special key becomes a Fields dimension.
		for key := range c.Columns {
			switch key {
			case "name", "id", "emdn_code", "emdn_term":
				continue
			}
			if v := val(rec, key); v != "" {
				if entry.Fields == nil {
					entry.Fields = map[string]string{}
				}
				entry.Fields[key] = v
			}
		}
		entries = append(entries, entry)
	}
	return NewInMemoryCatalog(entries), nil
}
