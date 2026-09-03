package ontology

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Normalize is the primary interchange entry point. It takes one raw device
// record from any external system and returns the canonical system-agnostic
// record. This is the function any migrator calls per row.
//
// Example:
//
//	result := ontology.Normalize(ontology.Input{
//	    DeviceName: "Catheter, sterile, single-use, adult",
//	    SourceType: "catheters and related",
//	    EMDNTerm:   "Peripheral venous catheters",
//	})
//	// result.IsValid() == true
//	// result.Name == "Peripheral intravenous catheter" (or similar)
//	// result.DeviceType == "Consumables & Accessories"
func Normalize(in Input) Result {
	row := map[string]string{
		"Legacy source name": in.DeviceName,
		"Source device type": in.SourceType,
		"EMDN term":          in.EMDNTerm,
	}

	resolved := ResolveRowNaming(row)

	// Derive confidence
	confidence := confidenceFor(resolved)

	return Result{
		Name:                  resolved.Name,
		DeviceType:            resolved.DeviceType,
		DeviceCategory:        resolved.DeviceCategory,
		DeviceFunction:        resolved.DeviceFunction,
		DeviceApplication:     resolved.DeviceFunction,
		DeviceApplicationRisk: resolved.DeviceApplicationRisk,
		LegacySourceName:      in.DeviceName,
		SourceType:            in.SourceType,
		EMDNCode:              in.EMDNCode,
		EMDNTerm:              in.EMDNTerm,
		MappingSource:         resolved.NamingSource,
		Confidence:            confidence,
	}
}

func confidenceFor(r ResolvedRow) string {
	switch r.NamingSource {
	case "unsupported_source_type":
		return "none"
	case "specific_rule":
		return "high"
	case "legacy_derived":
		if r.DeviceCategory != "" {
			return "high"
		}
		return "medium"
	case "family_fallback":
		if r.DeviceCategory != "" {
			return "medium"
		}
		return "low"
	default:
		return "low"
	}
}

// NormalizeBatch normalizes a slice of inputs, preserving order.
// Invalid results (unsupported source type) are included with Confidence "none"
// so the caller can decide how to handle them — filter, report, or override.
func NormalizeBatch(inputs []Input) []Result {
	out := make([]Result, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, Normalize(in))
	}
	return out
}

// NormalizeJSON reads a JSON array of Input objects and returns normalized Results.
// Input JSON uses the same snake_case field names as Input's json tags.
//
// Example input file:
//
//	[
//	  {"device_name": "ECG machine", "source_type": "monitoring equipment"},
//	  {"device_name": "Infusion pump", "source_type": "infusion devices"}
//	]
func NormalizeJSON(data []byte) ([]Result, error) {
	var inputs []Input
	if err := json.Unmarshal(data, &inputs); err != nil {
		// Try single object
		var single Input
		if err2 := json.Unmarshal(data, &single); err2 != nil {
			return nil, fmt.Errorf("ontology: invalid JSON: %w", err)
		}
		return []Result{Normalize(single)}, nil
	}
	return NormalizeBatch(inputs), nil
}

// NormalizeJSONFile reads a JSON file of Input objects and returns Results.
func NormalizeJSONFile(path string) ([]Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ontology: read %s: %w", path, err)
	}
	return NormalizeJSON(data)
}

// ToJSON marshals results to indented JSON.
func ToJSON(results []Result) ([]byte, error) {
	return json.MarshalIndent(results, "", "  ")
}

// ToCSV writes results to CSV using the interchange column order.
// Returns the CSV bytes.
func ToCSV(results []Result) ([]byte, error) {
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	headers := []string{
		"Name", "Device type", "Device category", "Device function",
		"Device application risk", "Legacy source name", "Source device type",
		"EMDN code", "EMDN term", "Mapping source", "Confidence",
	}
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, r := range results {
		row := []string{
			r.Name, r.DeviceType, r.DeviceCategory, r.DeviceFunction, r.DeviceApplicationRisk,
			r.LegacySourceName, r.SourceType, r.EMDNCode, r.EMDNTerm,
			r.MappingSource, r.Confidence,
		}
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
// Only valid results (IsValid() == true) are included. Deduplication is by
// (Name, DeviceType, DeviceCategory, DeviceFunction, DeviceApplicationRisk).
func ToAPIImportRecords(results []Result) []APIImportRecord {
	seen := make(map[string]struct{})
	var out []APIImportRecord
	for _, r := range results {
		if !r.IsValid() {
			continue
		}
		rec := r.ToAPIImportRecord()
		key := strings.Join([]string{
			rec.Name, rec.DeviceType, rec.DeviceCategory, rec.DeviceFunction,
			rec.DeviceApplicationRisk,
		}, "\x00")
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
		return out[i].DeviceType < out[j].DeviceType
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
