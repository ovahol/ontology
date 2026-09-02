package ontology

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		in   Input
		want string // expected OvaholType
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
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Normalize(tc.in)
			if result.OvaholType != tc.want {
				t.Errorf("Normalize(%+v) OvaholType = %q, want %q (full result: %+v)", tc.in, result.OvaholType, tc.want, result)
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

func TestNormalizeConfidence(t *testing.T) {
	r := Normalize(Input{DeviceName: "ECG machine", SourceType: "monitoring equipment"})
	if r.Confidence == "none" {
		t.Errorf("expected non-none confidence for valid input, got %q", r.Confidence)
	}
	r2 := Normalize(Input{DeviceName: "Widget", SourceType: "unknown xyz"})
	if r2.Confidence != "none" {
		t.Errorf("expected none confidence for unsupported, got %q", r2.Confidence)
	}
}

func TestNormalizeBatch(t *testing.T) {
	inputs := []Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
	}
	results := NormalizeBatch(inputs)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestNormalizeJSON(t *testing.T) {
	data := `[{"device_name":"ECG machine","source_type":"monitoring equipment"}]`
	results, err := NormalizeJSON([]byte(data))
	if err != nil {
		t.Fatalf("NormalizeJSON: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].OvaholType != "Monitoring & Measurement Devices" {
		t.Errorf("unexpected type: %q", results[0].OvaholType)
	}
}

func TestToAPIImportRecords(t *testing.T) {
	inputs := []Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"}, // duplicate
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
	}
	results := NormalizeBatch(inputs)
	apiRecords := ToAPIImportRecords(results)
	// Two unique combinations (ECG and infusion pump)
	if len(apiRecords) != 2 {
		t.Errorf("expected 2 deduped API records, got %d: %+v", len(apiRecords), apiRecords)
	}
}

func TestToCSV(t *testing.T) {
	results := NormalizeBatch([]Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
	})
	data, err := ToCSV(results)
	if err != nil {
		t.Fatalf("ToCSV: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty CSV")
	}
}

func TestLookupVocabulary(t *testing.T) {
	if len(OvaholDeviceTypes) != 8 {
		t.Errorf("expected 8 device types, got %d", len(OvaholDeviceTypes))
	}
	if len(DeviceFunctions) != 9 {
		t.Errorf("expected 9 device functions, got %d", len(DeviceFunctions))
	}
	if len(DeviceApplicationRisks) != 5 {
		t.Errorf("expected 5 risks, got %d", len(DeviceApplicationRisks))
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
