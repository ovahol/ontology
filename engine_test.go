package ontology

import (
	"os"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
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

// TestMeDevisDefaultReconciliation proves the embedded default (MeDevIS)
// taxonomy reconciles every real MeDevIS row to its exact (device_type,
// service_type, knowledge_level, reusable) tuple by device name alone — the
// same migration guarantee the Ovahol dictionary test gives, but for a
// genuinely different vendor whose vocabulary sits in its own field keys.
// It reads the reference devices.xlsx, so it is skipped when that (gitignored,
// local-only) file is absent.
func TestMeDevisDefaultReconciliation(t *testing.T) {
	if _, err := os.Stat("devices.xlsx"); err != nil {
		t.Skip("devices.xlsx not present; skipping local MeDevIS reconciliation")
	}
	tax := DefaultTaxonomy()
	f, err := excelize.OpenFile("devices.xlsx")
	if err != nil {
		t.Fatalf("open devices.xlsx: %v", err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		t.Fatalf("read sheet: %v", err)
	}
	count := 0
	dup := map[string]int{}
	for _, r := range rows {
		if len(r) == 0 || strings.TrimSpace(r[0]) == "" {
			continue
		}
		dup[strings.TrimSpace(r[0])]++
	}
	for i, r := range rows {
		if i == 0 || len(r) == 0 {
			continue
		}
		name := strings.TrimSpace(r[0])
		if name == "" {
			continue
		}
		// Names that appear more than once in the source carry conflicting
		// tuples; the taxonomy collapses each to one (most common) tuple, so
		// per-row comparison is only meaningful for unique names.
		if dup[name] > 1 {
			continue
		}
		res := NormalizeWithTaxonomy(Input{DeviceName: name}, tax)
		got := [4]string{
			res.GetField("device_type"),
			res.GetField("service_type"),
			res.GetField("knowledge_level"),
			res.GetField("reusable"),
		}
		want := [4]string{
			strings.TrimSpace(r[4]), strings.TrimSpace(r[5]),
			strings.TrimSpace(r[6]), strings.TrimSpace(r[2]),
		}
		if got != want {
			t.Errorf("row %d %q -> %v, want %v (source=%s)", i, name, got, want, res.MappingSource)
			continue
		}
		count++
	}
	if count == 0 {
		t.Fatal("no rows reconciled")
	}
	t.Logf("reconciled %d MeDevIS rows", count)
}

// TestEngineWithCustomVendorRules exercises a vendor whose taxonomy has a
// dimension ontology has never heard of ("device_tier") and no device_type
// at all. That's the point: nothing in the engine special-cases field names,
// so a vendor's own shape works exactly like Ovahol's or MeDevIS's.
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
			Rules: []Rule{
				{
					Keywords:    []string{"cryo pump", "cryo-cooler"},
					SourceTypes: []string{"cryo devices"},
					Set:         map[string]string{"device_tier": "Tier 1 - Critical"},
				},
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

	if res.DeviceType != "" {
		t.Errorf("vendor declares no device_type field; expected empty deprecated DeviceType accessor, got %q", res.DeviceType)
	}
	if res.Fields["device_tier"] != "Tier 1 - Critical" {
		t.Errorf("expected Fields[device_tier] = 'Tier 1 - Critical', got %q (fields: %+v)", res.Fields["device_tier"], res.Fields)
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
	if resolved.Fields["device_tier"] != "Tier 1 - Critical" {
		t.Errorf("expected resolved Fields[device_tier] = 'Tier 1 - Critical', got %q", resolved.Fields["device_tier"])
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
			Name: "custom-sensor",
			Fields: map[string]string{
				FieldDeviceType: "Sensors & Actuators",
			},
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
	// The catalog only declares device_type; the engine must NOT invent
	// other dimensions (e.g. device_category) from some other vendor's
	// vocabulary. A catalog hit is returned verbatim.
	if got := res.GetField(FieldDeviceCategory); got != "" {
		t.Errorf("engine invented a category (%q) the catalog did not declare", got)
	}
	if res.Fields == nil || len(res.Fields) != 1 {
		t.Errorf("expected exactly the catalog's single field, got %+v", res.Fields)
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
