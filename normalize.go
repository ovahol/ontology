package ontology

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// NormalizeWithTaxonomy is the single-record entry point: it takes one raw
// device record from any external system and returns the canonical system-
// agnostic result. Vendors supply their own taxonomy (LoadTaxonomyFile); a nil
// tax falls back to the embedded default.
//
// Example:
//
//	result := ontology.NormalizeWithTaxonomy(ontology.Input{
//	    DeviceName: "Catheter, sterile, single-use, adult",
//	    SourceType: "catheters and related",
//	    EMDNTerm:   "Peripheral venous catheters",
//	}, tax)
//	// result.IsValid() == true
//	// result.Name == "Peripheral intravenous catheter" (or similar)
//	// result.Fields[ontology.FieldDeviceType] == "Consumables & Accessories"
//
// This is the function any migrator calls per row.
func NormalizeWithTaxonomy(in Input, tax *Taxonomy) Result {
	row := map[string]string{
		"Legacy source name": in.DeviceName,
		"Source device type": in.SourceType,
		"EMDN term":          in.EMDNTerm,
	}

	resolved := ResolveRowNamingFor(row, tax)

	// Derive confidence
	confidence := confidenceFor(resolved)

	fields := resolved.Fields
	if len(fields) == 0 {
		fields = nil
	}
	return Result{
		Name:             resolved.Name,
		Fields:           fields,
		LegacySourceName: in.DeviceName,
		SourceType:       in.SourceType,
		EMDNCode:         in.EMDNCode,
		EMDNTerm:         in.EMDNTerm,
		MappingSource:    resolved.NamingSource,
		Confidence:       confidence,
	}
}

func confidenceFor(r ResolvedRow) string {
	switch r.NamingSource {
	case "unsupported_source_type":
		return "none"
	case "specific_rule":
		return "high"
	case "legacy_derived":
		if r.GetField(FieldDeviceCategory) != "" {
			return "high"
		}
		return "medium"
	case "family_fallback":
		if r.GetField(FieldDeviceCategory) != "" {
			return "medium"
		}
		return "low"
	default:
		return "low"
	}
}

// NormalizeBatchWithTaxonomy normalizes a slice of inputs, preserving order.
// Invalid results (unsupported source type) are included with Confidence "none"
// so the caller can decide how to handle them — filter, report, or override.
func NormalizeBatchWithTaxonomy(inputs []Input, tax *Taxonomy) []Result {
	out := make([]Result, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, NormalizeWithTaxonomy(in, tax))
	}
	return out
}

// NormalizeJSONWithTaxonomy reads a JSON array (or single object) of Input
// objects and returns normalized Results. Input JSON uses the same snake_case
// field names as Input's json tags.
//
// Example input:
//
//	[
//	  {"device_name": "ECG machine", "source_type": "monitoring equipment"},
//	  {"device_name": "Infusion pump", "source_type": "infusion devices"}
//	]
func NormalizeJSONWithTaxonomy(data []byte, tax *Taxonomy) ([]Result, error) {
	var inputs []Input
	if err := json.Unmarshal(data, &inputs); err != nil {
		// Try single object
		var single Input
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("ontology: invalid JSON: %w", err)
		}
		return []Result{NormalizeWithTaxonomy(single, tax)}, nil
	}
	return NormalizeBatchWithTaxonomy(inputs, tax), nil
}

// NormalizeJSONFileWithTaxonomy reads a JSON file of Input objects and returns
// normalized Results.
func NormalizeJSONFileWithTaxonomy(path string, tax *Taxonomy) ([]Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ontology: read %s: %w", path, err)
	}
	return NormalizeJSONWithTaxonomy(data, tax)
}

// ToJSON marshals results to indented JSON.
func ToJSON(results []Result) ([]byte, error) {
	return json.MarshalIndent(results, "", "  ")
}

// ToCSV writes results to CSV. Columns are taxonomy-driven: Name, then one
// column per resolved Fields key (sorted), then the input echo and
// diagnostics. The first row's field set defines the columns; missing values
// are written as empty cells.
func ToCSV(results []Result) ([]byte, error) {
	fieldKeys := make(map[string]struct{})
	for _, r := range results {
		for k := range r.Fields {
			fieldKeys[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(fieldKeys))
	for k := range fieldKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headers := []string{"Name"}
	headers = append(headers, keys...)
	headers = append(headers, "Legacy source name", "Source device type",
		"EMDN code", "EMDN term", "Mapping source", "Confidence")

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, r := range results {
		row := []string{r.Name}
		for _, k := range keys {
			row = append(row, r.Fields[k])
		}
		row = append(row, r.LegacySourceName, r.SourceType, r.EMDNCode, r.EMDNTerm,
			r.MappingSource, r.Confidence)
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

// ToAPIImportRecords deduplicates Results into API-ready import records.
// Only valid results (IsValid() == true) are included. Deduplication is driven
// by APIImportRecord.DedupKey(), which joins the name with every resolved
// Fields entry (sorted by key), so any taxonomy dimensions participate in
// deduplication without code changes.
func ToAPIImportRecords(results []Result) []APIImportRecord {
	seen := make(map[string]struct{})
	var out []APIImportRecord
	for _, r := range results {
		if !r.IsValid() {
			continue
		}
		rec := r.ToAPIImportRecord()
		key := rec.DedupKey()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, rec)
	}
	// Deterministic order
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].DedupKey() < out[j].DedupKey()
	})
	return out
}

// ValidateInput checks whether an Input has enough signal to attempt normalization.
// Returns a descriptive error if the input is empty; nil otherwise.
// Normalize itself never returns an error — it always produces a Result (possibly
// with Confidence "none"). Use this helper if you want to reject empty rows early.
func ValidateInput(in Input) error {
	if strings.TrimSpace(in.DeviceName) == "" &&
		strings.TrimSpace(in.SourceType) == "" &&
		strings.TrimSpace(in.EMDNTerm) == "" {
		return fmt.Errorf("ontology: input has no device_name, source_type, or emdn_term")
	}
	return nil
}
