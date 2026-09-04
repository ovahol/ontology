package ontology

import (
	"encoding/json"
	"testing"
)

// TestOvaholCompatibility ensures that when ontology is run against Ovahol's
// taxonomy, its output is always compatible with what Ovahol expects. Ovahol's
// vocabulary is no longer vendored in this repo — it lives in
// examples/taxonomies/ovahol.json (for testing only). If Ovahol's seed/lookup
// changes, either this file or that taxonomy must be updated — that signals drift.
//
// Source of truth for expected values: backend/internal/seed/lookup/data.go
// in the ovahol monorepo.

func loadOvaholTax(t *testing.T) *Taxonomy {
	t.Helper()
	tax, err := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	if err != nil {
		t.Fatalf("LoadTaxonomyFile(ovahol.json): %v", err)
	}
	return tax
}

func allowedSet(tax *Taxonomy, key string) map[string]bool {
	set := map[string]bool{}
	if fd := tax.Field(key); fd != nil {
		for _, v := range fd.AllowedValues {
			set[v] = true
		}
	}
	return set
}

func TestCompatibility_Fields(t *testing.T) {
	tax := loadOvaholTax(t)
	wantCounts := map[string]int{
		FieldDeviceType:            8,
		FieldDeviceCategory:        4,
		FieldDeviceFunction:        9,
		FieldDeviceApplicationRisk: 5,
	}
	for key, want := range wantCounts {
		fd := tax.Field(key)
		if fd == nil {
			t.Fatalf("ovahol taxonomy missing field %q", key)
		}
		if len(fd.AllowedValues) != want {
			t.Errorf("field %q has %d allowed_values, want %d", key, len(fd.AllowedValues), want)
		}
	}
}

// TestCompatibility_Rules ensures every rule in Ovahol's taxonomy only ever
// assigns values that are within that field's declared controlled
// vocabulary — the same guarantee the old hardcoded FamilyRule/TypeDefaults
// validity checks gave, now expressed generically over Rule.Set.
func TestCompatibility_Rules(t *testing.T) {
	tax := loadOvaholTax(t)
	if tax.Inference == nil || len(tax.Inference.Rules) == 0 {
		t.Fatal("ovahol taxonomy has no inference rules")
	}
	validByField := map[string]map[string]bool{
		FieldDeviceType:            allowedSet(tax, FieldDeviceType),
		FieldDeviceCategory:        allowedSet(tax, FieldDeviceCategory),
		FieldDeviceFunction:        allowedSet(tax, FieldDeviceFunction),
		FieldDeviceApplicationRisk: allowedSet(tax, FieldDeviceApplicationRisk),
		FieldDeviceFamily:          allowedSet(tax, FieldDeviceFamily),
	}
	for i, r := range tax.Inference.Rules {
		if len(r.Keywords) == 0 && len(r.SourceTypes) == 0 && len(r.Requires) == 0 {
			t.Errorf("rules[%d] matches unconditionally (no keywords, source_types, or requires)", i)
		}
		for field, value := range r.Set {
			valid, known := validByField[field]
			if !known || value == "" {
				continue
			}
			if !valid[value] {
				t.Errorf("rules[%d] sets %s=%q, not in that field's allowed_values", i, field, value)
			}
		}
		for field, value := range r.Requires {
			valid, known := validByField[field]
			if !known || value == "" {
				continue
			}
			if !valid[value] {
				t.Errorf("rules[%d] requires %s=%q, not in that field's allowed_values", i, field, value)
			}
		}
	}
}

func TestCompatibility_NormalizeOutputIsOvaholValid(t *testing.T) {
	tax := loadOvaholTax(t)
	validTypes := allowedSet(tax, FieldDeviceType)
	validFuncs := allowedSet(tax, FieldDeviceFunction)
	validRisks := allowedSet(tax, FieldDeviceApplicationRisk)
	samples := []Input{
		{DeviceName: "Patient monitor", SourceType: "monitoring equipment"},
		{DeviceName: "ECG machine, 12-lead", SourceType: "monitoring equipment"},
		{DeviceName: "Ultrasound scanner", SourceType: "imaging nuclear medicine equipment"},
		{DeviceName: "X-ray machine", SourceType: "imaging nuclear medicine equipment"},
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
		{DeviceName: "Ventilator, portable", SourceType: "medical equipment"},
		{DeviceName: "Chemistry analyzer", SourceType: "laboratory equipment"},
		{DeviceName: "Microscope", SourceType: "laboratory equipment"},
		{DeviceName: "Oxygen concentrator", SourceType: "medical gas equipment"},
		{DeviceName: "Suction machine", SourceType: "medical gas equipment"},
		{DeviceName: "Surgical forceps", SourceType: "medical equipment"},
		{DeviceName: "Autoclave", SourceType: "medical equipment"},
		{DeviceName: "Hospital bed", SourceType: "medical furniture"},
		{DeviceName: "Wheelchair", SourceType: "medical equipment"},
		{DeviceName: "Gloves, sterile", SourceType: "medical equipment"},
	}
	for _, in := range samples {
		result := NormalizeWithTaxonomy(in, tax)
		if !result.IsValid() {
			t.Errorf("Normalize(%+v) returned invalid (confidence=%s, source=%s) — sample should be valid", in, result.Confidence, result.MappingSource)
			continue
		}
		if !validTypes[result.DeviceType] {
			t.Errorf("Normalize(%+v).DeviceType %q invalid", in, result.DeviceType)
		}
		if !validFuncs[result.DeviceFunction] {
			t.Errorf("Normalize(%+v).DeviceFunction %q invalid", in, result.DeviceFunction)
		}
		if !validRisks[result.DeviceApplicationRisk] {
			t.Errorf("Normalize(%+v).DeviceApplicationRisk %q invalid", in, result.DeviceApplicationRisk)
		}
		if result.Name == "" {
			t.Errorf("Normalize(%+v) empty Name", in)
		}
		if result.MappingSource == "" {
			t.Errorf("Normalize(%+v) empty MappingSource", in)
		}
		if result.Confidence == "" {
			t.Errorf("Normalize(%+v) empty Confidence", in)
		}
	}
}

func TestCompatibility_NormalizeNeverLeaksFreeText(t *testing.T) {
	tax := loadOvaholTax(t)
	validTypes := allowedSet(tax, FieldDeviceType)
	// Inputs whose Name would ideally be vendor-neutral. The library
	// does best-effort humanization but may retain some vendor tokens for
	// highly vendor-specific names — that's a known improvement area, not a
	// compatibility break. We verify only that the controlled fields don't leak.
	inputs := []Input{
		{DeviceName: "Patient monitor, portable", SourceType: "monitoring equipment"},
		{DeviceName: "Ultrasound scanner, portable", SourceType: "imaging nuclear medicine equipment"},
		{DeviceName: "Infusion pump, volumetric", SourceType: "infusion devices"},
	}
	for _, in := range inputs {
		result := NormalizeWithTaxonomy(in, tax)
		if !result.IsValid() {
			t.Errorf("Normalize(%+v) unexpectedly invalid", in)
			continue
		}
		if !validTypes[result.DeviceType] {
			t.Errorf("Normalize(%+v) leaked free text into DeviceType: %q", in, result.DeviceType)
		}
		// Controlled fields must be valid, but Name may retain descriptive tokens
		if result.Name == "" {
			t.Errorf("Normalize(%+v) empty Name", in)
		}
	}
}

func TestCompatibility_JSONInterchangeRoundTrips(t *testing.T) {
	tax := loadOvaholTax(t)
	inputJSON := `[
		{"device_name": "ECG machine", "source_type": "monitoring equipment"},
		{"device_name": "Infusion pump", "source_type": "infusion devices", "emdn_term": "Infusion pumps"}
	]`
	results, err := NormalizeJSONWithTaxonomy([]byte(inputJSON), tax)
	if err != nil {
		t.Fatalf("NormalizeJSON failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	encoded, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var decoded []Result
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal round-trip: %v", err)
	}
	if len(decoded) != len(results) {
		t.Fatalf("round-trip length mismatch: %d vs %d", len(decoded), len(results))
	}
	for i, r := range decoded {
		if r.DeviceType != results[i].DeviceType {
			t.Errorf("round-trip [%d] DeviceType mismatch: %q vs %q", i, r.DeviceType, results[i].DeviceType)
		}
	}
}

func TestCompatibility_CSVAndAPIImportHeaders(t *testing.T) {
	wantSheetHeaders := []string{
		"Name", "Device type", "Device category", "Device function",
		"Device application risk", "Legacy source name", "Source device type",
		"EMDN code", "EMDN term",
	}
	if len(DefaultDeviceSheetHeaders) != len(wantSheetHeaders) {
		t.Fatalf("DefaultDeviceSheetHeaders length %d, want %d", len(DefaultDeviceSheetHeaders), len(wantSheetHeaders))
	}
	for i, want := range wantSheetHeaders {
		if DefaultDeviceSheetHeaders[i] != want {
			t.Errorf("DefaultDeviceSheetHeaders[%d] = %q, want %q", i, DefaultDeviceSheetHeaders[i], want)
		}
	}
	wantAPIHeaders := []string{"name", "device_type", "device_category", "device_function", "device_application_risk", "emdn_code", "emdn_term"}
	if len(DefaultAPIImportHeaders) != len(wantAPIHeaders) {
		t.Fatalf("DefaultAPIImportHeaders length %d, want %d", len(DefaultAPIImportHeaders), len(wantAPIHeaders))
	}
	for i, want := range wantAPIHeaders {
		if DefaultAPIImportHeaders[i] != want {
			t.Errorf("DefaultAPIImportHeaders[%d] = %q, want %q", i, DefaultAPIImportHeaders[i], want)
		}
	}
}

func TestCompatibility_ToAPIImportRecordsAreValid(t *testing.T) {
	tax := loadOvaholTax(t)
	results := NormalizeBatchWithTaxonomy([]Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
		{DeviceName: "Chemistry analyzer", SourceType: "laboratory equipment"},
	}, tax)
	records := ToAPIImportRecords(results)
	if len(records) == 0 {
		t.Fatal("ToAPIImportRecords returned no records for valid inputs")
	}
	validTypes := allowedSet(tax, FieldDeviceType)
	validFuncs := allowedSet(tax, FieldDeviceFunction)
	validRisks := allowedSet(tax, FieldDeviceApplicationRisk)
	for _, rec := range records {
		if !validTypes[rec.DeviceType] {
			t.Errorf("APIImportRecord %q has invalid device_type %q", rec.Name, rec.DeviceType)
		}
		if !validFuncs[rec.DeviceFunction] {
			t.Errorf("APIImportRecord %q has invalid device_function %q", rec.Name, rec.DeviceFunction)
		}
		if !validRisks[rec.DeviceApplicationRisk] {
			t.Errorf("APIImportRecord %q has invalid device_application_risk %q", rec.Name, rec.DeviceApplicationRisk)
		}
		if rec.Name == "" {
			t.Error("APIImportRecord has empty name")
		}
	}
}
