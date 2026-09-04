package ontology

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func writeTestWorkbook(t *testing.T, path string, headers []string, rows [][]string) {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("save input workbook: %v", err)
	}
}

func TestNormalizeWorkbookWithCatalog(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.xlsx")
	writeTestWorkbook(t, in,
		[]string{"Device name", "Source device type", "Model", "Serial number"},
		[][]string{
			{"Known sensor", "industrial automation", "A-100", "SN-0001"},
			{"Mystery box", "unspecified", "", ""},
			{"Known sensor", "industrial automation", "A-200", "SN-0002"},
		},
	)

	// A foreign vendor's device dictionary: its only dimension is "device_tier",
	// and it has no device_type at all. A catalog hit must return it verbatim.
	cat := NewInMemoryCatalog([]CatalogEntry{
		{Name: "Known sensor", Fields: map[string]string{"device_tier": "Tier 1 - Critical"}},
	})
	tax := &Taxonomy{
		ID:            "acme",
		Version:       "1.0.0",
		SchemaVersion: "1.0.0",
		Fields: []FieldDef{
			{Key: "device_tier", Label: "Device tier", Required: true, AllowedValues: []string{"Tier 1 - Critical", "Tier 2 - General"}},
		},
	}

	out := filepath.Join(dir, "out.xlsx")
	if _, err := NormalizeWorkbookWithCatalogAndTaxonomy(in, out, cat, tax); err != nil {
		t.Fatalf("NormalizeWorkbookWithCatalogAndTaxonomy: %v", err)
	}

	f, err := excelize.OpenFile(out)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()
	// Find the devices sheet by looking for the row header that contains the
	// taxonomy's first field label ("Device tier") — the workbook keeps the
	// input sheet's name, so don't assume a literal name.
	var rows [][]string
	for _, sn := range f.GetSheetList() {
		r, err := f.GetRows(sn)
		if err != nil || len(r) == 0 {
			continue
		}
		for _, h := range r[0] {
			if strings.TrimSpace(h) == "Device tier" {
				rows = r
				break
			}
		}
		if rows != nil {
			break
		}
	}
	if rows == nil {
		t.Fatal("could not locate resolved devices sheet in output")
	}

	// Header names + the resolved device_tier per data row, ignoring which
	// column number; verify by locating the "Device tier" column.
	header := rows[0]
	tierCol := -1
	for i, h := range header {
		if strings.TrimSpace(h) == "Device tier" {
			tierCol = i
		}
	}
	if tierCol < 0 {
		t.Fatalf("output missing Device tier column, headers=%v", header)
	}

	// Each "Known sensor" row must carry the catalog's verbatim tier, and the
	// passthrough columns (Model, Serial number) must survive.
	knownTier := 0
	hasModelCol := false
	var modelVals []string
	for i, h := range header {
		if strings.TrimSpace(h) == "Model" {
			hasModelCol = true
			for _, r := range rows[1:] {
				if len(r) > i {
					modelVals = append(modelVals, r[i])
				}
			}
		}
	}
	for _, r := range rows[1:] {
		tier := ""
		if tierCol < len(r) {
			tier = r[tierCol]
		}
		if tier == "Tier 1 - Critical" {
			knownTier++
		}
	}
	if knownTier != 2 {
		t.Errorf("expected 2 catalog-exact rows resolved to Tier 1 - Critical, got %d", knownTier)
	}
	if !hasModelCol {
		t.Errorf("passthrough Model column was dropped")
	}
	for _, mv := range modelVals {
		if mv == "" {
			t.Errorf("passthrough Model value was dropped")
		}
	}
}

func TestNormalizeWorkbookWithNilCatalogIsTaxonomyOnly(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.xlsx")
	out := filepath.Join(dir, "out.xlsx")
	writeTestWorkbook(t, in,
		[]string{"Device name", "Source device type"},
		[][]string{{"Known sensor", "anything"}},
	)
	tax := &Taxonomy{
		ID:            "acme",
		Version:       "1.0.0",
		SchemaVersion: "1.0.0",
		Fields: []FieldDef{
			{Key: "device_tier", Label: "Device tier", Required: true, AllowedValues: []string{"Tier 1 - Critical"}},
		},
		Inference: &InferenceRules{Rules: []Rule{
			{Keywords: []string{"Known sensor"}, Set: map[string]string{"device_tier": "Tier 1 - Critical"}},
		}},
	}
	// nil catalog must behave identically to the taxonomy-only path (works off
	// rules, not the dictionary).
	if _, err := NormalizeWorkbookWithCatalogAndTaxonomy(in, out, nil, tax); err != nil {
		t.Fatalf("nil-catalog run: %v", err)
	}
	_ = os.Remove(out)
}
