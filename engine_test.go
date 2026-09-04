package ontology

import (
	"testing"
)

func TestEngineWithOvaholTaxonomy(t *testing.T) {
	tax, err := LoadTaxonomyFile("examples/taxonomies/ovahol.json")
	if err != nil {
		t.Fatalf("LoadTaxonomyFile error: %v", err)
	}

	engine := NewEngine(tax)
	res := engine.Normalize(Input{
		DeviceName: "Patient monitor, multiparameter",
		SourceType: "monitoring equipment",
	})

	if res.DeviceType != "Monitoring & Measurement Devices" {
		t.Errorf("expected DeviceType 'Monitoring & Measurement Devices', got %q", res.DeviceType)
	}
	if res.Confidence != "high" {
		t.Errorf("expected high confidence, got %q", res.Confidence)
	}
}

func TestEngineWithMedevisTaxonomy(t *testing.T) {
	tax, err := LoadTaxonomyFile("examples/taxonomies/medevis.json")
	if err != nil {
		t.Fatalf("LoadTaxonomyFile error: %v", err)
	}

	engine := NewEngine(tax)
	result := engine.Normalize(Input{
		DeviceName: "Linear accelerator system",
		SourceType: "Radiotherapy-related equipment",
	})
	if result.Confidence == "none" || result.DeviceType != "Radiotherapy-related equipment" {
		t.Errorf("expected valid classification from medevis taxonomy, got %+v", result)
	}
}

func TestEngineWithCustomVendorRules(t *testing.T) {
	customTax := &Taxonomy{
		ID:            "vendor-alpha",
		Name:          "Alpha Corp Taxonomy",
		Version:       "1.0.0",
		SchemaVersion: "1.0.0",
		Fields: []FieldDef{
			{
				Key:           "device_tier",
				Label:         "Device Tier",
				Required:      true,
				AllowedValues: []string{"Tier 1 - Critical", "Tier 2 - General"},
			},
		},
		Inference: &InferenceRules{
			TypeByKeyword: []struct {
				Keywords []string `json:"keywords"`
				Type     string   `json:"type"`
			}{
				{
					Keywords: []string{"cryo pump", "cryo-cooler"},
					Type:     "Cryogenic Systems",
				},
			},
			SourceTypeMap: map[string]string{
				"cryo devices": "Cryogenic Systems",
			},
			LegacyDescriptorPhrases: []string{"ultra-low", "50k"},
			Acronyms:                []string{"cryo", "mri"},
			NameRefinementRules: []NameRefinementRule{
				{
					TargetName:    "CRYO pump",
					Keywords:      []string{"ultra low", "ultra-low"},
					CommonName:    "Ultra-low Cryo Pump",
					CanonicalName: "Ultra-low Temperature Cryogenic Pump",
				},
			},
			SearchAliasRules: []SearchAliasRule{
				{
					Keywords: []string{"cryo pump"},
					Aliases:  []string{"Cryo Chiller", "Cryogenic Vacuum Pump"},
				},
			},
		},
	}

	if err := customTax.Validate(); err != nil {
		t.Fatalf("Taxonomy.Validate() error: %v", err)
	}

	engine := NewEngine(customTax)

	res := engine.Normalize(Input{
		DeviceName: "Cryo pump, ultra-low, 50K",
		SourceType: "cryo devices",
	})

	if res.DeviceType != "Cryogenic Systems" {
		t.Errorf("expected DeviceType 'Cryogenic Systems', got %q", res.DeviceType)
	}
	if res.Name != "Ultra-low Cryo Pump" {
		t.Errorf("expected refined common name 'Ultra-low Cryo Pump', got %q", res.Name)
	}

	// Test ResolveRowNamingFor for full naming refinement & search aliases
	row := map[string]string{
		"Legacy source name": "Cryo pump, ultra-low, 50K",
		"Source device type": "cryo devices",
	}
	resolved := ResolveRowNamingFor(row, customTax)
	if resolved.Name != "Ultra-low Cryo Pump" {
		t.Errorf("expected resolved Name 'Ultra-low Cryo Pump', got %q", resolved.Name)
	}
	if resolved.CanonicalName != "Ultra-low Temperature Cryogenic Pump" {
		t.Errorf("expected resolved CanonicalName 'Ultra-low Temperature Cryogenic Pump', got %q", resolved.CanonicalName)
	}

	hasAlias := false
	for _, a := range resolved.CommonNames {
		if a == "Cryo Chiller" {
			hasAlias = true
			break
		}
	}
	if !hasAlias {
		t.Errorf("expected custom search alias 'Cryo Chiller' in CommonNames: %v", resolved.CommonNames)
	}
}

func TestEngineWithCatalogOverride(t *testing.T) {
	cat := NewInMemoryCatalog([]CatalogEntry{
		{
			Name:       "custom-sensor",
			DeviceType: "Sensors & Actuators",
		},
	})

	engine := NewEngine(nil, WithCatalog(cat))

	res := engine.Normalize(Input{
		DeviceName: "custom-sensor",
	})

	if res.DeviceType != "Sensors & Actuators" {
		t.Errorf("expected DeviceType from catalog 'Sensors & Actuators', got %q", res.DeviceType)
	}
	if res.MappingSource != "catalog_exact" {
		t.Errorf("expected MappingSource 'catalog_exact', got %q", res.MappingSource)
	}
	if res.Confidence != "high" {
		t.Errorf("expected Confidence 'high', got %q", res.Confidence)
	}
}

func TestDynamicFieldsDedupKey(t *testing.T) {
	rec1 := APIImportRecord{
		Name: "Test Device",
		Fields: map[string]string{
			"tier":   "Tier 1",
			"domain": "Radiology",
		},
	}
	rec2 := APIImportRecord{
		Name: "Test Device",
		Fields: map[string]string{
			"domain": "Radiology",
			"tier":   "Tier 1",
		},
	}

	if rec1.DedupKey() != rec2.DedupKey() {
		t.Errorf("expected equal dedup keys regardless of map iteration order, got %q vs %q",
			rec1.DedupKey(), rec2.DedupKey())
	}
}
