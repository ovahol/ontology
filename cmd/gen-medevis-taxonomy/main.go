// Command gen-medevis-taxonomy reads the WHO MeDevIS device list (the
// reference devices.xlsx, 2653 rows) and emits examples/taxonomies/medevis.json
// using MeDevIS's own structure — NOT Ovahol's. This is the embedded default
// taxonomy (see default.go), so it must describe the real MeDevIS dimensions:
//
//	device_type        (the "Device type" column, 39 values)
//	service_type       (the "Service type" column, 12 values)
//	knowledge_level    (the "Knowledge level" column: Basic / General clinical / Specialized clinical)
//	reusable           (the "Reusable" column: Reusable / Single use)
//	emdn_code / emdn_term / gmdn_code / gmdn_term   (nomenclature lookups)
//
// Inference is derived from the rows: an exact device-name rule per distinct
// name (longest first) sets the full classification tuple verbatim, mirroring
// the migration guarantee that a known name resolves to its exact dictionary
// row. EMDN codes are deliberately not used for classification because (like
// Ovahol's) a single EMDN code maps to several distinct tuples — the device
// name is the disambiguator.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/ovahol/ontology"
)

type row struct {
	name        string
	deviceType  string
	serviceType string
	knowledge   string
	reusable    string
	emdnCode    string
	emdnTerm    string
	gmdnCode    string
	gmdnTerm    string
}

func clean(s string) string {
	return strings.TrimSpace(s)
}

func main() {
	inPath := flag.String("in", "devices.xlsx", "path to the MeDevIS reference xlsx")
	outPath := flag.String("out", "examples/taxonomies/medevis.json", "output taxonomy JSON path")
	flag.Parse()

	f, err := excelize.OpenFile(*inPath)
	if err != nil {
		log.Fatalf("open %s: %v", *inPath, err)
	}
	defer f.Close()
	sheet := f.GetSheetList()[0]
	rows, err := f.GetRows(sheet)
	if err != nil {
		log.Fatalf("read sheet: %v", err)
	}
	if len(rows) < 2 {
		log.Fatalf("no data rows in %s", *inPath)
	}
	header := rows[0]
	idx := func(prefix string) int {
		for i, h := range header {
			if h == prefix || strings.HasPrefix(h, prefix+"\n") || strings.HasPrefix(h, prefix+" ") {
				return i
			}
		}
		return -1
	}
	iName := idx("Device name")
	iType := idx("Device type")
	iService := idx("Service type")
	iKnowledge := idx("Knowledge level")
	iReusable := idx("Reusable")
	iEMDNC := idx("Nomenclature code (EMDN)")
	iEMDNT := idx("Nomenclature term (EMDN)")
	iGMDNC := idx("Nomenclature code (GMDN)")
	iGMDNT := idx("Nomenclature term (GMDN)")
	if iType < 0 || iService < 0 || iKnowledge < 0 || iReusable < 0 {
		log.Fatalf("missing expected MeDevIS columns (type=%d service=%d knowledge=%d reusable=%d)", iType, iService, iKnowledge, iReusable)
	}

	setType := map[string]bool{}
	setService := map[string]bool{}
	setKnowledge := map[string]bool{}
	setReusable := map[string]bool{}

	var data []row
	for _, r := range rows[1:] {
		if len(r) == 0 {
			continue
		}
		rec := row{
			name:        clean(val(r, iName)),
			deviceType:  clean(val(r, iType)),
			serviceType: clean(val(r, iService)),
			knowledge:   clean(val(r, iKnowledge)),
			reusable:    clean(val(r, iReusable)),
			emdnCode:    clean(val(r, iEMDNC)),
			emdnTerm:    clean(val(r, iEMDNT)),
			gmdnCode:    clean(val(r, iGMDNC)),
			gmdnTerm:    clean(val(r, iGMDNT)),
		}
		if rec.name == "" || rec.deviceType == "" {
			continue
		}
		setType[rec.deviceType] = true
		setService[rec.serviceType] = true
		setKnowledge[rec.knowledge] = true
		setReusable[rec.reusable] = true
		data = append(data, rec)
	}

	// Deduplicate names deterministically (rare): keep the first occurrence.
	collisionSet := map[string]int{}
	for i := range data {
		nm := data[i].name
		collisionSet[nm]++
	}
	seen := map[string]bool{}
	var distinct []row
	for i := range data {
		nm := data[i].name
		if seen[nm] {
			continue
		}
		seen[nm] = true
		r := data[i]
		// For any name that maps to >1 tuple, keep the single most common tuple.
		if collisionSet[nm] > 1 {
			r = mostCommonTuple(data, nm)
		}
		distinct = append(distinct, r)
	}

	// Build inference rules, longest name first.
	rules := make([]ontology.Rule, 0, len(distinct))
	for _, r := range distinct {
		rules = append(rules, ontology.Rule{
			Keywords: []string{r.name},
			Set: map[string]string{
				ontology.FieldDeviceType: r.deviceType,
				"service_type":           r.serviceType,
				"knowledge_level":        r.knowledge,
				"reusable":               r.reusable,
			},
			Name:          r.name,
			CanonicalName: r.name,
		})
	}
	sort.SliceStable(rules, func(i, j int) bool {
		return len(rules[i].Keywords[0]) > len(rules[j].Keywords[0])
	})

	tax := ontology.Taxonomy{
		ID:            "medevis",
		Name:          "WHO MeDevIS Medical Devices Reference",
		Version:       "2025.2.1",
		SchemaVersion: ontology.CurrentSchemaVersion,
		Fields: []ontology.FieldDef{
			{Key: ontology.FieldDeviceType, Label: "Device type", Required: true, AllowedValues: sortedKeys(setType)},
			{Key: "service_type", Label: "Service type", AllowedValues: sortedKeys(setService)},
			{Key: "knowledge_level", Label: "Knowledge level", AllowedValues: sortedKeys(setKnowledge)},
			{Key: "reusable", Label: "Reusable", AllowedValues: sortedKeys(setReusable)},
			{Key: emdnCodeField, Label: "EMDN code"},
			{Key: emdnTermField, Label: "EMDN term"},
			{Key: gmdnCodeField, Label: "GMDN code"},
			{Key: gmdnTermField, Label: "GMDN term"},
		},
		Inference: &ontology.InferenceRules{
			Rules:                   rules,
			GenericLegacyHeads:      []string{"analyser", "analyzer", "bag", "bottle", "chair", "cup", "device", "filter", "holder", "incubator", "kit", "machine", "mask", "meter", "monitor", "pump", "sensor", "set", "system", "tip", "tray", "unit"},
			LegacyDescriptorPhrases: []string{"adult", "automated", "bedside", "continuous", "disposable", "manual", "non invasive", "non-invasive", "portable", "reusable", "semi automated", "semi-automated", "single use", "single-use", "sterile"},
		},
		Source: fmt.Sprintf("devices.xlsx MeDevIS 2025 v2.1 (%d rows)", len(data)),
	}

	buf, err := json.MarshalIndent(tax, "", "  ")
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}
	if err := jsonRoundTrip(tax); err != nil {
		log.Fatalf("round-trip validation: %v", err)
	}
	if err := os.WriteFile(*outPath, append(buf, '\n'), 0644); err != nil {
		log.Fatalf("write %s: %v", *outPath, err)
	}
	fmt.Printf("wrote %s: %d fields, %d rules, %d rows\n", *outPath, len(tax.Fields), len(rules), len(data))
}

const (
	emdnCodeField = "emdn_code"
	emdnTermField = "emdn_term"
	gmdnCodeField = "gmdn_code"
	gmdnTermField = "gmdn_term"
)

func val(r []string, i int) string {
	if i < 0 || i >= len(r) {
		return ""
	}
	return r[i]
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func mostCommonTuple(data []row, name string) row {
	count := map[[4]string]int{}
	var first row
	for _, r := range data {
		if r.name != name {
			continue
		}
		t := [4]string{r.deviceType, r.serviceType, r.knowledge, r.reusable}
		count[t]++
		if first.name == "" {
			first = r
		}
	}
	best := first
	bestN := -1
	for _, r := range data {
		if r.name != name {
			continue
		}
		t := [4]string{r.deviceType, r.serviceType, r.knowledge, r.reusable}
		if count[t] > bestN {
			bestN = count[t]
			best = r
		}
	}
	return best
}

func jsonRoundTrip(tax ontology.Taxonomy) error {
	buf, err := json.Marshal(tax)
	if err != nil {
		return err
	}
	_, err = ontology.LoadTaxonomy(buf)
	return err
}
