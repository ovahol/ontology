package ontology

import (
	"os"
	"strings"
	"testing"
)

func TestCatalogExactMatch(t *testing.T) {
	cat := NewInMemoryCatalog([]CatalogEntry{
		{
			Name: "ECG machine",
			Fields: map[string]string{
				"device_type":             "Monitoring & Measurement Devices",
				"device_category":         "Diagnostic",
				"device_function":         "Additional Physiological Monitoring and Diagnostic",
				"device_application_risk": "Inappropriate therapy or misdiagnosis",
			},
			ID: "device-uuid-1",
		},
	})

	// exact hit — should bypass taxonomy, return catalog verbatim
	r := NormalizeWithCatalog(Input{DeviceName: "ECG machine", SourceType: "monitoring equipment"}, cat)
	if r.MappingSource != "catalog_exact" {
		t.Fatalf("expected catalog_exact, got %q", r.MappingSource)
	}
	if r.Name != "ECG machine" || r.DeviceType != "Monitoring & Measurement Devices" {
		t.Errorf("unexpected catalog result: %+v", r)
	}
	if r.Confidence != "high" {
		t.Errorf("expected high confidence for catalog hit, got %q", r.Confidence)
	}
	// EMDN code match already covers alias via normalized name
	// now alias via common EMDN is primary
	// EMDN code match
	r3 := NormalizeWithCatalog(Input{EMDNCode: " Z1201 "}, NewInMemoryCatalog([]CatalogEntry{
		{Name: "Infusion pump", EMDNCode: "Z1201", Fields: map[string]string{
			"device_type":             "Treatment, Surgical & Life Support Devices",
			"device_function":         "Surgical and Intensive Care",
			"device_application_risk": "Potential patient or operator injury",
		}},
	}))
	if r3.MappingSource != "catalog_exact" {
		t.Fatalf("expected catalog hit via EMDN, got %q", r3.MappingSource)
	}
}

func TestCatalogFallback(t *testing.T) {
	tax, err := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	if err != nil {
		t.Fatalf("load taxonomy: %v", err)
	}
	cat := NewInMemoryCatalog([]CatalogEntry{
		{Name: "Known Device", Fields: map[string]string{
			"device_type":             "Support Equipment & Furniture",
			"device_function":         "Patient Related and Other",
			"device_application_risk": "Equipment damage",
		}},
	})
	// miss → fallback to taxonomy
	r := NormalizeWithCatalogAndTaxonomy(Input{DeviceName: "Infusion pump", SourceType: "infusion devices"}, cat, tax)
	if r.MappingSource == "catalog_exact" {
		t.Fatalf("expected taxonomy fallback, got catalog_exact")
	}
	if r.DeviceType != "Treatment, Surgical & Life Support Devices" {
		t.Errorf("fallback inference wrong: %+v", r)
	}
	// nil catalog → always taxonomy
	rNil := NormalizeWithCatalogAndTaxonomy(Input{DeviceName: "ECG machine", SourceType: "monitoring equipment"}, nil, tax)
	if rNil.MappingSource == "catalog_exact" {
		t.Error("nil catalog should not return catalog_exact")
	}
}

func TestCatalogUnsupportedSourceBypass(t *testing.T) {
	// even unsupported source types are resolved if catalog has them
	cat := NewInMemoryCatalog([]CatalogEntry{
		{Name: "Weird Widget", Fields: map[string]string{
			"device_type":             "Consumables & Accessories",
			"device_function":         "Patient Related and Other",
			"device_application_risk": "Equipment damage",
		}},
	})
	r := NormalizeWithCatalog(Input{DeviceName: "Weird Widget", SourceType: "unknown category xyz"}, cat)
	if r.MappingSource != "catalog_exact" || r.DeviceType != "Consumables & Accessories" {
		t.Fatalf("catalog should bypass unsupported_source_type: %+v", r)
	}
}

// TestCSVCatalogLoad exercises the generic device-dictionary CSV loader with
// Ovahol's own column layout.
func TestCSVCatalogLoad(t *testing.T) {
	csvData := `Name,Device Type,Device Function,Application Risk,EMDN Code,EMDN Term
Infusion pump,"Treatment, Surgical & Life Support Devices","Surgical and Intensive Care","Potential patient or operator injury",Z12030301,INFUSION PUMPS
ECG machine,"Monitoring & Measurement Devices","Additional Physiological Monitoring and Diagnostic","Inappropriate therapy or misdiagnosis",Z11010101,ECG MACHINES
`
	cat, err := (CSVCatalog{Columns: map[string]string{
		"name":                    "Name",
		"device_type":             "Device Type",
		"device_function":         "Device Function",
		"device_application_risk": "Application Risk",
		"emdn_code":               "EMDN Code",
		"emdn_term":               "EMDN Term",
	}}).Load(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("CSVCatalog.Load: %v", err)
	}
	if cat == nil || len(cat.entries) != 2 {
		t.Fatalf("expected 2 catalog entries, got %d", len(cat.entries))
	}
	r := NormalizeWithCatalog(Input{DeviceName: "Infusion pump"}, cat)
	if r.MappingSource != "catalog_exact" || r.DeviceType != "Treatment, Surgical & Life Support Devices" {
		t.Fatalf("catalog load reconcile failed: %+v", r)
	}
	// EMDN code lookup on a loaded entry.
	r2 := NormalizeWithCatalog(Input{EMDNCode: "Z11010101"}, cat)
	if r2.MappingSource != "catalog_exact" || r2.Name != "ECG machine" {
		t.Fatalf("EMDN lookup via loaded catalog failed: %+v", r2)
	}
}

// TestDictionaryReconciliation is the core migration guarantee: every device
// in Ovahol's dictionary (devices.csv) reconciles to its exact dictionary
// (type, function, risk) when a device name is supplied — both through the
// exact catalog path and through the generated taxonomy's rules alone.
func TestDictionaryReconciliation(t *testing.T) {
	dictionary := []CatalogEntry{}
	f, err := os.Open("devices.csv")
	if err != nil {
		t.Fatalf("open devices.csv: %v (run tests from the module root)", err)
	}
	defer f.Close()
	cat, err := (CSVCatalog{Columns: map[string]string{
		"name":                    "Name",
		"device_type":             "Device Type",
		"device_function":         "Device Function",
		"device_application_risk": "Application Risk",
		"emdn_code":               "EMDN Code",
		"emdn_term":               "EMDN Term",
	}}).Load(f)
	if err != nil {
		t.Fatalf("load dictionary: %v", err)
	}
	dictionary = cat.entries

	tax, err := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	if err != nil {
		t.Fatalf("load taxonomy: %v", err)
	}

	// Path A: exact catalog reconciliation (the seamless migration path).
	batch := make([]Input, 0, len(dictionary))
	for _, e := range dictionary {
		batch = append(batch, Input{DeviceName: e.Name, EMDNCode: e.EMDNCode})
	}
	results := NormalizeBatchWithCatalogAndTaxonomy(batch, cat, tax)
	for i, e := range dictionary {
		r := results[i]
		if r.DeviceType != e.DeviceType() || r.DeviceFunction != e.DeviceFunction() || r.DeviceApplicationRisk != e.DeviceApplicationRisk() {
			t.Errorf("[catalog] %q -> (%s|%s|%s), want (%s|%s|%s)",
				e.Name, r.DeviceType, r.DeviceFunction, r.DeviceApplicationRisk,
				e.DeviceType(), e.DeviceFunction(), e.DeviceApplicationRisk())
		}
		if r.MappingSource != "catalog_exact" {
			t.Errorf("[catalog] %q mapping source %q, want catalog_exact", e.Name, r.MappingSource)
		}
	}

	// Path B: taxonomy rules alone (no catalog) must also reconcile every
	// dictionary name to its exact triple.
	rulesResults := NormalizeBatchWithTaxonomy(batch, tax)
	for i, e := range dictionary {
		r := rulesResults[i]
		if r.DeviceType != e.DeviceType() || r.DeviceFunction != e.DeviceFunction() || r.DeviceApplicationRisk != e.DeviceApplicationRisk() {
			t.Errorf("[rules] %q -> (%s|%s|%s), want (%s|%s|%s)",
				e.Name, r.DeviceType, r.DeviceFunction, r.DeviceApplicationRisk,
				e.DeviceType(), e.DeviceFunction(), e.DeviceApplicationRisk())
		}
	}
}

// TestCatalogPrefersNameOverCollidingEMDN guards the fix that a known device
// name wins over a shared EMDN code that otherwise maps to several devices.
func TestCatalogPrefersNameOverCollidingEMDN(t *testing.T) {
	cat := NewInMemoryCatalog([]CatalogEntry{
		{Name: "Infusion pump", EMDNCode: "Z1203", Fields: map[string]string{
			"device_type":             "Treatment, Surgical & Life Support Devices",
			"device_function":         "Surgical and Intensive Care",
			"device_application_risk": "Potential patient or operator injury",
		}},
		{Name: "Saline stand", EMDNCode: "Z1203", Fields: map[string]string{
			"device_type":             "Support Equipment & Furniture",
			"device_function":         "Patient Related and Other",
			"device_application_risk": "Equipment damage",
		}},
	})
	r := NormalizeWithCatalog(Input{DeviceName: "Infusion pump", EMDNCode: "Z1203"}, cat)
	if r.Name != "Infusion pump" || r.DeviceType != "Treatment, Surgical & Life Support Devices" {
		t.Fatalf("name should win over colliding EMDN, got: %+v", r)
	}
}
