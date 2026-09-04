package ontology

import "testing"

// ovaholTax loads Ovahol's taxonomy from its example file. Ovahol's vocabulary
// is no longer the built-in default — it lives in examples/taxonomies/ovahol.json
// for use/testing only.

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want string // expected DeviceType
	}{
		{
			name: "ECG via monitoring equipment",
			in:   Input{DeviceName: "ECG machine, portable", SourceType: "monitoring equipment"},
			want: "Monitoring & Measurement Devices",
		},
		{
			name: "infusion pump",
			in:   Input{DeviceName: "Infusion pump, volumetric", SourceType: "infusion devices"},
			want: "Treatment, Surgical & Life Support Devices",
		},
		{
			name: "ultrasound via EMDN",
			in:   Input{DeviceName: "Ultrasound scanner", SourceType: "imaging nuclear medicine equipment", EMDNTerm: "Diagnostic ultrasound systems"},
			want: "Diagnostic & Imaging Devices",
		},
		{
			name: "unsupported source type",
			in:   Input{DeviceName: "Something", SourceType: "unknown category xyz"},
			want: "",
		},
		{
			name: "lab analyzer",
			in:   Input{DeviceName: "Chemistry analyzer", SourceType: "laboratory equipment"},
			want: "Laboratory & IVD Equipment",
		},
		{
			name: "oxygen concentrator via specific rule",
			in:   Input{DeviceName: "Oxygen concentrator, 5L", SourceType: "medical gas equipment"},
			want: "Medical Gas & Respiratory Devices",
		},
	}
	tax, err := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	if err != nil {
		t.Fatalf("load ovahol taxonomy: %v", err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeWithTaxonomy(tc.in, tax)
			if got := result.GetField(FieldDeviceType); got != tc.want {
				t.Errorf("Normalize(%+v) device_type = %q, want %q (full result: %+v)", tc.in, got, tc.want, result)
			}
			if tc.want == "" && result.IsValid() {
				t.Errorf("expected invalid result for unsupported source type, got valid: %+v", result)
			}
			if tc.want != "" && !result.IsValid() {
				t.Errorf("expected valid result, got invalid: %+v", result)
			}
		})
	}
}

func TestNormalizeWithMedevisDefault(t *testing.T) {
	// When no taxonomy is supplied, the engine uses the WHO/Medevis default.
	tests := []struct {
		name string
		in   Input
		want string
	}{
		{
			name: "radiotherapy source type maps to itself",
			in:   Input{DeviceName: "Linear accelerator system", SourceType: "Radiotherapy-related equipment"},
			want: "Radiotherapy-related equipment",
		},
		{
			name: "monitoring equipment source type",
			in:   Input{DeviceName: "Patient monitor", SourceType: "monitoring equipment"},
			want: "Monitoring equipment",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := NormalizeWithTaxonomy(tc.in, nil)
			if got := result.GetField(FieldDeviceType); got != tc.want {
				t.Errorf("Normalize(%+v) device_type = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNormalizeConfidence(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	r := NormalizeWithTaxonomy(Input{DeviceName: "ECG machine", SourceType: "monitoring equipment"}, tax)
	if r.Confidence == "none" {
		t.Errorf("expected non-none confidence for valid input, got %q", r.Confidence)
	}
	r2 := NormalizeWithTaxonomy(Input{DeviceName: "Widget", SourceType: "unknown xyz"}, tax)
	if r2.Confidence != "none" {
		t.Errorf("expected none confidence for unsupported, got %q", r2.Confidence)
	}
}

func TestNormalizeBatch(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	inputs := []Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
	}
	results := NormalizeBatchWithTaxonomy(inputs, tax)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestNormalizeJSON(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	data := []byte(`[{"device_name":"ECG machine","source_type":"monitoring equipment"}]`)
	results, err := NormalizeJSONWithTaxonomy(data, tax)
	if err != nil {
		t.Fatalf("NormalizeJSON: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if got := results[0].GetField(FieldDeviceType); got != "Monitoring & Measurement Devices" {
		t.Errorf("unexpected type: %q", got)
	}
}

func TestToAPIImportRecords(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	inputs := []Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"}, // duplicate
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
	}
	results := NormalizeBatchWithTaxonomy(inputs, tax)
	apiRecords := ToAPIImportRecords(results)
	// Two unique combinations (ECG and infusion pump)
	if len(apiRecords) != 2 {
		t.Errorf("expected 2 deduped API records, got %d: %+v", len(apiRecords), apiRecords)
	}
}

func TestToCSV(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	results := NormalizeBatchWithTaxonomy([]Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
	}, tax)
	data, err := ToCSV(results)
	if err != nil {
		t.Fatalf("ToCSV: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty CSV")
	}
}

func TestDefaultTaxonomyIsMedevis(t *testing.T) {
	tax := DefaultTaxonomy()
	if tax == nil {
		t.Fatal("expected non-nil default taxonomy")
	}
	if tax.ID != "medevis" {
		t.Errorf("expected default taxonomy id 'medevis', got %q", tax.ID)
	}
	fd := tax.Field(FieldDeviceType)
	if fd == nil || len(fd.AllowedValues) != 39 {
		t.Errorf("expected 39 medevis device types, got %d", len(fd.AllowedValues))
	}
}

func TestValidateInput(t *testing.T) {
	if err := ValidateInput(Input{}); err == nil {
		t.Error("expected error for empty input")
	}
	if err := ValidateInput(Input{DeviceName: "something"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgnosticJSONFields(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	r := NormalizeWithTaxonomy(Input{DeviceName: "ECG machine", SourceType: "monitoring equipment"}, tax)
	if r.Name == "" {
		t.Error("expected non-empty Name")
	}
	if r.GetField(FieldDeviceType) == "" {
		t.Error("expected non-empty device_type")
	}
	if r.GetField(FieldDeviceCategory) == "" {
		t.Error("expected non-empty device_category")
	}
	if r.GetField(FieldDeviceFunction) == "" {
		t.Error("expected non-empty device_function")
	}
	// Ensure legacy top-level keys not present, streamlined keys present.
	// The pluggable "fields" map carries the dimensions, so we only reject
	// legacy top-level keys (ovahol_*, common_name, canonical_name,
	// search_aliases), not the presence of device_family inside fields.
	data, _ := ToJSON([]Result{r})
	s := string(data)
	if contains(s, "\"ovahol_device_type\"") || contains(s, "\"ovahol_device_family\"") || contains(s, "\"common_name\"") || contains(s, "\"canonical_name\"") || contains(s, "\"search_aliases\"") {
		t.Errorf("JSON still contains legacy top-level keys: %s", s)
	}
	if !contains(s, "\"name\"") || !contains(s, FieldDeviceType) || !contains(s, FieldDeviceCategory) || !contains(s, FieldDeviceFunction) || !contains(s, FieldDeviceApplicationRisk) {
		t.Errorf("JSON missing streamlined keys: %s", s)
	}
	// Pluggable Fields map is the sole storage.
	if r.Fields == nil {
		t.Error("expected non-nil Fields map (pluggable storage)")
	}
	if r.GetField(FieldDeviceType) != r.Fields[FieldDeviceType] {
		t.Errorf("GetField/Fields mismatch for device_type")
	}
}

func TestStreamlinedFields(t *testing.T) {
	tax, _ := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	r := NormalizeWithTaxonomy(Input{DeviceName: "Infusion pump", SourceType: "infusion devices"}, tax)
	if r.GetField(FieldDeviceType) != "Treatment, Surgical & Life Support Devices" {
		t.Errorf("unexpected device_type: %q", r.GetField(FieldDeviceType))
	}
	if r.GetField(FieldDeviceCategory) != "Therapeutic" {
		t.Errorf("expected Therapeutic category, got %q", r.GetField(FieldDeviceCategory))
	}
	if r.GetField(FieldDeviceFunction) == "" || r.GetField(FieldDeviceApplicationRisk) == "" {
		t.Errorf("expected function/risk, got %q / %q", r.GetField(FieldDeviceFunction), r.GetField(FieldDeviceApplicationRisk))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}
