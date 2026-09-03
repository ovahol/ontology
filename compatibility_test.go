package ontology

import (
	"encoding/json"
	"testing"
)

// TestOvaholCompatibility ensures the ontology's output is always
// compatible with what Ovahol expects. If Ovahol's seed/lookup changes,
// these tests fail — that's intentional, it signals drift.
//
// Source of truth for expected values: backend/internal/seed/lookup/data.go
// in the ovahol monorepo. The ontology vendors these values statically in
// taxonomy.go so it has no runtime dependency on the monorepo, but the
// values must stay identical.

var expectedOvaholDeviceTypes = []DeviceType{
	{Name: "Monitoring & Measurement Devices", Code: "MONITORING_MEASUREMENT_DEVICES"},
	{Name: "Diagnostic & Imaging Devices", Code: "DIAGNOSTIC_IMAGING_DEVICES"},
	{Name: "Treatment, Surgical & Life Support Devices", Code: "TREATMENT_SURGICAL_LIFE_SUPPORT_DEVICES"},
	{Name: "Laboratory & IVD Equipment", Code: "LABORATORY_IVD_EQUIPMENT"},
	{Name: "Medical Gas & Respiratory Devices", Code: "MEDICAL_GAS_RESPIRATORY_DEVICES"},
	{Name: "Sterilization & Infection Control Devices", Code: "STERILIZATION_INFECTION_CONTROL_DEVICES"},
	{Name: "Support Equipment & Furniture", Code: "SUPPORT_EQUIPMENT_FURNITURE"},
	{Name: "Consumables & Accessories", Code: "CONSUMABLES_ACCESSORIES"},
}

var expectedOvaholDeviceFunctions = []DeviceFunction{
	{Name: "Life Support", Code: "LIFE_SUPPORT", Category: "Therapeutic", Score: 10},
	{Name: "Surgical and Intensive Care", Code: "SURGICAL_INTENSIVE_CARE", Category: "Therapeutic", Score: 9},
	{Name: "Physical Therapy and Treatment", Code: "PHYSICAL_THERAPY_TREATMENT", Category: "Therapeutic", Score: 8},
	{Name: "Surgical and Intensive Care Monitoring", Code: "CRITICAL_CARE_MONITORING", Category: "Diagnostic", Score: 7},
	{Name: "Additional Physiological Monitoring and Diagnostic", Code: "GENERAL_PHYSIOLOGICAL_MONITORING", Category: "Diagnostic", Score: 6},
	{Name: "Analytical Laboratory", Code: "ANALYTICAL_LABORATORY", Category: "Analytical", Score: 5},
	{Name: "Laboratory Accessories", Code: "LABORATORY_ACCESSORIES", Category: "Analytical", Score: 4},
	{Name: "Computers and Related", Code: "COMPUTERS_AND_IT", Category: "Analytical", Score: 3},
	{Name: "Patient Related and Other", Code: "PATIENT_RELATED_OTHER", Category: "Miscellaneous", Score: 2},
}

var expectedOvaholDeviceRisks = []DeviceApplicationRisk{
	{Description: "Potential patient death", ScorePoint: 5},
	{Description: "Potential patient or operator injury", ScorePoint: 4},
	{Description: "Inappropriate therapy or misdiagnosis", ScorePoint: 3},
	{Description: "Equipment damage", ScorePoint: 2},
	{Description: "No significant identified risk", ScorePoint: 1},
}

func TestCompatibility_DeviceTypes(t *testing.T) {
	if len(OvaholDeviceTypes) != len(expectedOvaholDeviceTypes) {
		t.Fatalf("device types count drifted: got %d, want %d. If ovahol added/removed a type, update both repos.", len(OvaholDeviceTypes), len(expectedOvaholDeviceTypes))
	}
	for i, want := range expectedOvaholDeviceTypes {
		got := OvaholDeviceTypes[i]
		if got.Name != want.Name || got.Code != want.Code {
			t.Errorf("device type [%d] mismatch: got %+v, want %+v", i, got, want)
		}
	}
	validNames := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validNames[dt.Name] = true
	}
	for code, name := range TypeByCode {
		if !validNames[name] {
			t.Errorf("TypeByCode[%q] = %q is not a valid Ovahol device type", code, name)
		}
	}
}

func TestCompatibility_DeviceFunctions(t *testing.T) {
	if len(DeviceFunctions) != len(expectedOvaholDeviceFunctions) {
		t.Fatalf("device functions count drifted: got %d, want %d", len(DeviceFunctions), len(expectedOvaholDeviceFunctions))
	}
	for i, want := range expectedOvaholDeviceFunctions {
		got := DeviceFunctions[i]
		if got.Name != want.Name || got.Code != want.Code || got.Category != want.Category || got.Score != want.Score {
			t.Errorf("device function [%d] mismatch: got %+v, want %+v", i, got, want)
		}
	}
	validFuncNames := make(map[string]bool, len(DeviceFunctions))
	for _, fn := range DeviceFunctions {
		validFuncNames[fn.Name] = true
	}
	for code, name := range FunctionByCode {
		if !validFuncNames[name] {
			t.Errorf("FunctionByCode[%q] = %q is not a valid device function", code, name)
		}
	}
}

func TestCompatibility_DeviceRisks(t *testing.T) {
	if len(DeviceApplicationRisks) != len(expectedOvaholDeviceRisks) {
		t.Fatalf("device risks count drifted: got %d, want %d", len(DeviceApplicationRisks), len(expectedOvaholDeviceRisks))
	}
	for i, want := range expectedOvaholDeviceRisks {
		got := DeviceApplicationRisks[i]
		if got.Description != want.Description || got.ScorePoint != want.ScorePoint {
			t.Errorf("device risk [%d] mismatch: got %+v, want %+v", i, got, want)
		}
	}
}

func TestCompatibility_FamilyRules(t *testing.T) {
	validTypes := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validTypes[dt.Name] = true
	}
	validFuncs := make(map[string]bool, len(DeviceFunctions))
	for _, fn := range DeviceFunctions {
		validFuncs[fn.Name] = true
	}
	validRisks := make(map[string]bool, len(DeviceApplicationRisks))
	for _, r := range DeviceApplicationRisks {
		validRisks[r.Description] = true
	}
	for i, rule := range FamilyRules {
		if !validTypes[rule.Type] {
			t.Errorf("FamilyRules[%d] (%s / %s) has invalid Type %q", i, rule.Type, rule.Family, rule.Type)
		}
		if !validFuncs[rule.Function] {
			t.Errorf("FamilyRules[%d] (%s / %s) has invalid Function %q", i, rule.Type, rule.Family, rule.Function)
		}
		if !validRisks[rule.Risk] {
			t.Errorf("FamilyRules[%d] (%s / %s) has invalid Risk %q", i, rule.Type, rule.Family, rule.Risk)
		}
		if rule.Family == "" {
			t.Errorf("FamilyRules[%d] has empty Family", i)
		}
		if rule.CommonName == "" {
			t.Errorf("FamilyRules[%d] (%s / %s) has empty CommonName", i, rule.Type, rule.Family)
		}
		if rule.CanonicalName == "" {
			t.Errorf("FamilyRules[%d] (%s / %s) has empty CanonicalName", i, rule.Type, rule.Family)
		}
	}
}

func TestCompatibility_SpecificNameRules(t *testing.T) {
	validTypes := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validTypes[dt.Name] = true
	}
	for i, rule := range SpecificNameRules {
		if rule.Type != "" && !validTypes[rule.Type] {
			t.Errorf("SpecificNameRules[%d] (%v) has invalid Type %q", i, rule.Keywords, rule.Type)
		}
		if len(rule.Keywords) == 0 {
			t.Errorf("SpecificNameRules[%d] has no Keywords", i)
		}
		if rule.CommonName == "" {
			t.Errorf("SpecificNameRules[%d] has empty CommonName", i)
		}
	}
}

func TestCompatibility_TypeDefaults(t *testing.T) {
	validTypes := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validTypes[dt.Name] = true
	}
	validFuncs := make(map[string]bool, len(DeviceFunctions))
	for _, fn := range DeviceFunctions {
		validFuncs[fn.Name] = true
	}
	validRisks := make(map[string]bool, len(DeviceApplicationRisks))
	for _, r := range DeviceApplicationRisks {
		validRisks[r.Description] = true
	}
	if len(OvaholTypeDefaults) != len(OvaholDeviceTypes) {
		t.Errorf("OvaholTypeDefaults has %d entries, want %d (one per device type)", len(OvaholTypeDefaults), len(OvaholDeviceTypes))
	}
	for typeName, def := range OvaholTypeDefaults {
		if !validTypes[typeName] {
			t.Errorf("OvaholTypeDefaults key %q is not a valid device type", typeName)
		}
		if !validFuncs[def.Function] {
			t.Errorf("OvaholTypeDefaults[%q].Function %q is not a valid function", typeName, def.Function)
		}
		if !validRisks[def.Risk] {
			t.Errorf("OvaholTypeDefaults[%q].Risk %q is not a valid risk", typeName, def.Risk)
		}
	}
}

func TestCompatibility_NormalizeOutputIsOvaholValid(t *testing.T) {
	validTypes := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validTypes[dt.Name] = true
	}
	validFuncs := make(map[string]bool, len(DeviceFunctions))
	for _, fn := range DeviceFunctions {
		validFuncs[fn.Name] = true
	}
	validRisks := make(map[string]bool, len(DeviceApplicationRisks))
	for _, r := range DeviceApplicationRisks {
		validRisks[r.Description] = true
	}
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
		result := Normalize(in)
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
	validTypes := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validTypes[dt.Name] = true
	}
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
		result := Normalize(in)
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
	inputJSON := `[
		{"device_name": "ECG machine", "source_type": "monitoring equipment"},
		{"device_name": "Infusion pump", "source_type": "infusion devices", "emdn_term": "Infusion pumps"}
	]`
	results, err := NormalizeJSON([]byte(inputJSON))
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
	if len(DeviceSheetHeaders) != len(wantSheetHeaders) {
		t.Fatalf("DeviceSheetHeaders length %d, want %d", len(DeviceSheetHeaders), len(wantSheetHeaders))
	}
	for i, want := range wantSheetHeaders {
		if DeviceSheetHeaders[i] != want {
			t.Errorf("DeviceSheetHeaders[%d] = %q, want %q", i, DeviceSheetHeaders[i], want)
		}
	}
	wantAPIHeaders := []string{"name", "device_type", "device_category", "device_function", "device_application_risk", "emdn_code", "emdn_term"}
	if len(APIImportHeaders) != len(wantAPIHeaders) {
		t.Fatalf("APIImportHeaders length %d, want %d", len(APIImportHeaders), len(wantAPIHeaders))
	}
	for i, want := range wantAPIHeaders {
		if APIImportHeaders[i] != want {
			t.Errorf("APIImportHeaders[%d] = %q, want %q", i, APIImportHeaders[i], want)
		}
	}
}

func TestCompatibility_ToAPIImportRecordsAreValid(t *testing.T) {
	results := NormalizeBatch([]Input{
		{DeviceName: "ECG machine", SourceType: "monitoring equipment"},
		{DeviceName: "Infusion pump", SourceType: "infusion devices"},
		{DeviceName: "Chemistry analyzer", SourceType: "laboratory equipment"},
	})
	records := ToAPIImportRecords(results)
	if len(records) == 0 {
		t.Fatal("ToAPIImportRecords returned no records for valid inputs")
	}
	validTypes := make(map[string]bool, len(OvaholDeviceTypes))
	for _, dt := range OvaholDeviceTypes {
		validTypes[dt.Name] = true
	}
	validFuncs := make(map[string]bool, len(DeviceFunctions))
	for _, fn := range DeviceFunctions {
		validFuncs[fn.Name] = true
	}
	validRisks := make(map[string]bool, len(DeviceApplicationRisks))
	for _, r := range DeviceApplicationRisks {
		validRisks[r.Description] = true
	}
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
