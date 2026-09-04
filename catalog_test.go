package ontology

import "testing"

func TestCatalogExactMatch(t *testing.T) {
	cat := NewInMemoryCatalog([]CatalogEntry{
		{
			Name:                  "ECG machine",
			DeviceType:            "Monitoring & Measurement Devices",
			DeviceCategory:        "Diagnostic",
			DeviceFunction:        "Additional Physiological Monitoring and Diagnostic",
			DeviceApplicationRisk: "Inappropriate therapy or misdiagnosis",
			ID:                    "device-uuid-1",
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
		{Name: "Infusion pump", DeviceType: "Treatment, Surgical & Life Support Devices", DeviceFunction: "Surgical and Intensive Care", DeviceApplicationRisk: "Potential patient or operator injury", EMDNCode: "Z1201"},
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
		{Name: "Known Device", DeviceType: "Support Equipment & Furniture", DeviceFunction: "Patient Related and Other", DeviceApplicationRisk: "Equipment damage"},
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
		{Name: "Weird Widget", DeviceType: "Consumables & Accessories", DeviceFunction: "Patient Related and Other", DeviceApplicationRisk: "Equipment damage"},
	})
	r := NormalizeWithCatalog(Input{DeviceName: "Weird Widget", SourceType: "unknown category xyz"}, cat)
	if r.MappingSource != "catalog_exact" || r.DeviceType != "Consumables & Accessories" {
		t.Fatalf("catalog should bypass unsupported_source_type: %+v", r)
	}
}
