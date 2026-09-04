// Command gen-ovahol-taxonomy generates examples/taxonomies/ovahol.json
// from the Ovahol device dictionary (devices.csv).
//
// The dictionary is the source of truth: every device row carries its exact
// (Device Type, Device Function, Application Risk) triple, so the generated
// taxonomy reconciles any inbound device onto Ovahol's own vocabulary.
// Keeping the taxonomy generated from the CSV guarantees it can never drift
// from the dictionary.
//
// Run from the repo root:
//
//	go run ./cmd/gen-ovahol-taxonomy
package main

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

//go:embed curated.json
var curatedJSON []byte

// curated holds hand-authored name-quality data that has no source in
// devices.csv (the dictionary has no acronym list, alias list, or
// name-refinement rules) but is needed for good output on inputs that don't
// match a dictionary name verbatim — e.g. "ECG machine, 12-lead" isn't a
// dictionary name, so it falls through to legacy-derived text parsing, which
// needs to know "ecg" stays uppercase. This is embedded (not regenerated
// from the CSV) so it survives every future run of this generator.
type curated struct {
	GenericLegacyHeads      []string             `json:"genericLegacyHeads"`
	LegacyDescriptorPhrases []string             `json:"legacyDescriptorPhrases"`
	NameRefinementRules     []nameRefinementRule `json:"nameRefinementRules"`
	SearchAliasRules        []searchAliasRule    `json:"searchAliasRules"`
	Acronyms                []string             `json:"acronyms"`
	WordReplacements        map[string]string    `json:"wordReplacements"`
}

func loadCurated() curated {
	var c curated
	if err := json.Unmarshal(curatedJSON, &c); err != nil {
		fmt.Fprintln(os.Stderr, "curated.json:", err)
		os.Exit(1)
	}
	return c
}

type fieldDef struct {
	Key           string   `json:"key"`
	Label         string   `json:"label"`
	Required      bool     `json:"required"`
	AllowedValues []string `json:"allowed_values"`
}

type rule struct {
	ID            string            `json:"id,omitempty"`
	Keywords      []string          `json:"keywords,omitempty"`
	Requires      map[string]string `json:"requires,omitempty"`
	Set           map[string]string `json:"set,omitempty"`
	Name          string            `json:"name,omitempty"`
	CanonicalName string            `json:"canonical_name,omitempty"`
}

type nameRefinementRule struct {
	TargetName    string   `json:"targetName"`
	Keywords      []string `json:"keywords,omitempty"`
	CommonName    string   `json:"commonName"`
	CanonicalName string   `json:"canonicalName,omitempty"`
}

type searchAliasRule struct {
	Keywords []string `json:"keywords"`
	Aliases  []string `json:"aliases"`
}

type inferenceRules struct {
	Rules                   []rule               `json:"rules"`
	GenericLegacyHeads      []string             `json:"genericLegacyHeads,omitempty"`
	LegacyDescriptorPhrases []string             `json:"legacyDescriptorPhrases,omitempty"`
	NameRefinementRules     []nameRefinementRule `json:"nameRefinementRules,omitempty"`
	SearchAliasRules        []searchAliasRule    `json:"searchAliasRules,omitempty"`
	Acronyms                []string             `json:"acronyms,omitempty"`
	WordReplacements        map[string]string    `json:"wordReplacements,omitempty"`
}

type taxonomy struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	SchemaVersion string          `json:"schemaVersion"`
	Fields        []fieldDef      `json:"fields"`
	Inference     *inferenceRules `json:"inference"`
	Source        string          `json:"source"`
}

type deviceRow struct {
	Name     string
	Type     string
	Function string
	Risk     string
	EMDNCode string
	EMDNTerm string
	GMDNCode string
	GMDNTerm string
}

var (
	// Device types present in devices.csv (canonical, in order of first appearance).
	deviceTypes = []string{}
	// Device functions present in devices.csv.
	deviceFunctions = []string{}
	// Application risks present in devices.csv.
	applicationRisks = []string{}
)

func main() {
	root, _ := os.Getwd()
	csvPath := flag.String("csv", filepath.Join(root, "devices.csv"), "path to the device dictionary CSV")
	outPath := flag.String("out", filepath.Join(root, "examples", "taxonomies", "ovahol.json"), "output taxonomy JSON path")
	flag.Parse()

	rows := readDevices(*csvPath)

	fields := buildFields(rows)
	rules := buildRules(rows)
	c := loadCurated()

	tax := &taxonomy{
		ID:            "ovahol-ontology",
		Name:          "Ovahol device dictionary reconciliation",
		Version:       "1.0.0",
		SchemaVersion: "1.0.0",
		Fields:        fields,
		Inference: &inferenceRules{
			Rules:                   rules,
			GenericLegacyHeads:      c.GenericLegacyHeads,
			LegacyDescriptorPhrases: c.LegacyDescriptorPhrases,
			NameRefinementRules:     c.NameRefinementRules,
			SearchAliasRules:        c.SearchAliasRules,
			Acronyms:                c.Acronyms,
			WordReplacements:        c.WordReplacements,
		},
		Source: "devices.csv",
	}

	data, err := json.MarshalIndent(tax, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "marshal:", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "mkdir:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	fmt.Printf("wrote %s (%d device names, %d rules, %d fields)\n",
		*outPath, len(rows), len(rules), len(fields))
}

func readDevices(path string) []deviceRow {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	recs, err := r.ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, "csv:", err)
		os.Exit(1)
	}
	if len(recs) < 2 {
		fmt.Fprintln(os.Stderr, "csv: empty")
		os.Exit(1)
	}
	header := map[string]int{}
	for i, h := range recs[0] {
		header[strings.TrimSpace(h)] = i
	}
	get := func(rec []string, col string) string {
		i, ok := header[col]
		if !ok || i >= len(rec) {
			return ""
		}
		return strings.TrimSpace(rec[i])
	}
	seen := map[string]bool{}
	var rows []deviceRow
	for _, rec := range recs[1:] {
		row := deviceRow{
			Name:     get(rec, "Name"),
			Type:     get(rec, "Device Type"),
			Function: get(rec, "Device Function"),
			Risk:     get(rec, "Application Risk"),
			EMDNCode: get(rec, "EMDN Code"),
			EMDNTerm: get(rec, "EMDN Term"),
			GMDNCode: get(rec, "GMDN Code"),
			GMDNTerm: get(rec, "GMDN Term"),
		}
		if row.Name == "" {
			continue
		}
		if seen[row.Name] {
			continue
		}
		seen[row.Name] = true
		rows = append(rows, row)
		addVocab(row.Type, row.Function, row.Risk)
	}
	return rows
}

func addVocab(t, f, r string) {
	if t != "" && !contains(deviceTypes, t) {
		deviceTypes = append(deviceTypes, t)
	}
	if f != "" && !contains(deviceFunctions, f) {
		deviceFunctions = append(deviceFunctions, f)
	}
	if r != "" && !contains(applicationRisks, r) {
		applicationRisks = append(applicationRisks, r)
	}
}

func buildFields(rows []deviceRow) []fieldDef {
	return []fieldDef{
		{
			Key:           "device_type",
			Label:         "Device Type",
			Required:      true,
			AllowedValues: deviceTypes,
		},
		{
			Key:           "device_function",
			Label:         "Device Function",
			Required:      true,
			AllowedValues: deviceFunctions,
		},
		{
			Key:           "device_application_risk",
			Label:         "Application Risk",
			Required:      true,
			AllowedValues: applicationRisks,
		},
		{
			Key:           "device_category",
			Label:         "Device Category",
			Required:      false,
			AllowedValues: []string{"Therapeutic", "Diagnostic", "Analytical", "Miscellaneous"},
		},
		{
			Key:           "emdn_code",
			Label:         "EMDN Code",
			Required:      false,
			AllowedValues: emdnCodes(rows),
		},
		{
			Key:           "emdn_term",
			Label:         "EMDN Term",
			Required:      false,
			AllowedValues: emdnTerms(rows),
		},
		{
			Key:           "gmdn_code",
			Label:         "GMDN Code",
			Required:      false,
			AllowedValues: []string{},
		},
		{
			Key:           "gmdn_term",
			Label:         "GMDN Term",
			Required:      false,
			AllowedValues: []string{},
		},
	}
}

func emdnCodes(rows []deviceRow) []string {
	m := map[string]bool{}
	for _, r := range rows {
		if r.EMDNCode != "" {
			m[r.EMDNCode] = true
		}
	}
	return keys(m)
}

func emdnTerms(rows []deviceRow) []string {
	m := map[string]bool{}
	for _, r := range rows {
		if r.EMDNTerm != "" {
			m[r.EMDNTerm] = true
		}
	}
	return keys(m)
}

func buildRules(rows []deviceRow) []rule {
	var rules []rule

	// Stage 1 — exact device-name reconciliation. Each dictionary device name
	// maps to one exact (type, function, risk) triple, verified unambiguous.
	for _, r := range rows {
		set := map[string]string{
			"device_type":             r.Type,
			"device_function":         r.Function,
			"device_application_risk": r.Risk,
		}
		if r.EMDNCode != "" {
			set["emdn_code"] = r.EMDNCode
		}
		if r.EMDNTerm != "" {
			set["emdn_term"] = r.EMDNTerm
		}
		rules = append(rules, rule{
			ID:            "ovahol.name." + slug(r.Name),
			Keywords:      []string{r.Name},
			Set:           set,
			Name:          r.Name,
			CanonicalName: r.Name,
		})
	}

	// Order exact-name rules longest-first so the most specific name is
	// evaluated before a shorter sibling it contains (e.g. "Conical
	// centrifuge tube" before "Centrifuge tube"). Keyword matching is fuzzy
	// (substring/token), so specificity ordering minimizes false hits when a
	// name shares tokens or substrings with another dictionary device.
	sort.SliceStable(rules, func(i, j int) bool {
		return len(norm(rules[i].Name)) > len(norm(rules[j].Name))
	})

	// Stage 2 — curated generic classification. Common medical-device
	// terminology is not always an exact dictionary name (an inbound record
	// may arrive as "ECG machine", "X-ray machine", "patient monitor"), so a
	// hand-curated keyword bucket classifies a recognizable device_type for
	// those. Exact dictionary names never reach here (Stage 1 already set all
	// three fields); these only fill the device_type for names the dictionary
	// does not spell out verbatim.
	rules = append(rules, curatedGenericRules()...)

	// Stage 3 — device_type -> device_function/risk defaults. Fires only for
	// inputs whose device_type was resolved (by an exact name or a generic
	// keyword) but whose function/risk are still unset. Gives every resolved
	// type a sensible dictionary-conformant default.
	for _, t := range deviceTypes {
		fn, rk, ok := typeDefault(t)
		if !ok {
			continue
		}
		rules = append(rules, rule{
			ID: "ovahol.type." + slug(t),
			Requires: map[string]string{
				"device_type": t,
			},
			Set: map[string]string{
				"device_function":         fn,
				"device_application_risk": rk,
			},
		})
	}

	// Stage 3 — device_function -> device_category. Pure derivation: the
	// category is a function of the function, matching Ovahol's own FK model.
	categoryRules := []struct{ function, category string }{
		{"Life Support", "Therapeutic"},
		{"Surgical and Intensive Care", "Therapeutic"},
		{"Physical Therapy and Treatment", "Therapeutic"},
		{"Surgical and Intensive Care Monitoring", "Diagnostic"},
		{"Additional Physiological Monitoring and Diagnostic", "Diagnostic"},
		{"Analytical Laboratory", "Analytical"},
		{"Laboratory Accessories", "Analytical"},
		{"Computers and Related", "Analytical"},
		{"Patient Related and Other", "Miscellaneous"},
	}
	for _, cr := range categoryRules {
		rules = append(rules, rule{
			ID: "ovahol.category." + slug(cr.function),
			Requires: map[string]string{
				"device_function": cr.function,
			},
			Set: map[string]string{
				"device_category": cr.category,
			},
		})
	}

	return rules
}

// curatedGenericRules returns the hand-curated keyword buckets that classify
// a recognizable device_type from common medical-device terminology. These
// are complementary to the dictionary: they recognise the generic names a
// migrating source system commonly uses for equipment whose exact device
// name is not in devices.csv. Matching is ordered most-specific-bucket-first,
// and each bucket only ever sets device_type (function/risk are filled by the
// type->defaults stage so the two stay consistent).
func curatedGenericRules() []rule {
	// bucket: device_type -> keywords. Later buckets are lower priority and
	// only fire when an earlier, more specific one did not.
	buckets := []struct {
		deviceType string
		keywords   []string
	}{
		{"Diagnostic & Imaging Devices", []string{
			"ultrasound", "sonograph", "x ray", "xray", "computed tomography",
			"ct scan", "mri", "magnetic resonance", "fluoroscopy", "mammograph",
			"radiograph", "nuclear medicine", "gamma camera", "endoscope",
			"bronchoscope", "cystoscope", "colposcope", "laparoscope",
			"arthroscope", "otoscope", "ophthalmoscope", "angiograph",
		}},
		{"Laboratory & IVD Equipment", []string{
			"ivd", "reagent", "assay", "hematology", "haematology",
			"chemistry analyzer", "chemistry analyser", "centrifuge",
			"microscope", "pcr", "blood gas analyzer", "laboratory",
			"specimen", "cuvette", "pipette", "ph meter",
		}},
		{"Medical Gas & Respiratory Devices", []string{
			"oxygen", "suction", "airway", "nebuliz", "nebulis", "flowmeter",
			"humidifier", "psa", "breathing circuit", "respiratory", "cpap",
			"bipap", "aspirator", "concentrator",
		}},
		{"Treatment, Surgical & Life Support Devices", []string{
			"ventilator", "infusion", "syringe pump", "defibrillator",
			"dialysis", "anaesthesia", "anesthesia", "electrosurgical",
			"pacemaker", "cardioverter", "apheresis", "resuscitator",
			"life support", "surgical", "implant", "catheter", "cannula",
			"trocar", "stent", "cautery",
		}},
		{"Monitoring & Measurement Devices", []string{
			"monitor", "ecg", "eeg", "tonometer", "audiometer", "spirometr",
			"pulmonary function", "bilirubinometer", "vital signs",
			"blood pressure", "stethoscope", "oximeter", "thermometer",
			"scale", "goniometer", "manometer",
		}},
		{"Sterilization & Infection Control Devices", []string{
			"steriliz", "autoclave", "disinfect", "decontamin", "ipc",
			"glove", "gown", "apron",
		}},
		{"Consumables & Accessories", []string{
			"dressing", "gauze", "bandage", "syringe", "needle", "drainage",
			"collection bag", "swab", "suture", "condom", "diaphragm",
			"intravenous line",
		}},
		{"Support Equipment & Furniture", []string{
			"wheelchair", "walker", "crutch", "prosthe", "orthotic",
			"mobility", "assistive", "rehabilitation", "physiotherapy",
			"hospital bed", "bedside", "cabinet", "trolley", "chart",
			"stationery", "waste", "lamp", "radiation shielding", "stretcher",
		}},
	}
	rules := make([]rule, 0, len(buckets))
	for _, b := range buckets {
		rules = append(rules, rule{
			ID:       "ovahol.generic." + slug(b.deviceType),
			Keywords: b.keywords,
			Set:      map[string]string{"device_type": b.deviceType},
		})
	}
	return rules
}

// typeDefault returns the curated device_function/application_risk default for
// a resolved device_type when only the type is known (generic-name input).
func typeDefault(dt string) (function, risk string, ok bool) {
	defaults := map[string][2]string{
		"Consumables & Accessories":                  {"Patient Related and Other", "Equipment damage"},
		"Diagnostic & Imaging Devices":               {"Additional Physiological Monitoring and Diagnostic", "Inappropriate therapy or misdiagnosis"},
		"Laboratory & IVD Equipment":                 {"Analytical Laboratory", "Inappropriate therapy or misdiagnosis"},
		"Medical Gas & Respiratory Devices":          {"Life Support", "Potential patient or operator injury"},
		"Monitoring & Measurement Devices":           {"Additional Physiological Monitoring and Diagnostic", "Inappropriate therapy or misdiagnosis"},
		"Sterilization & Infection Control Devices":  {"Patient Related and Other", "Potential patient or operator injury"},
		"Support Equipment & Furniture":              {"Patient Related and Other", "Equipment damage"},
		"Treatment, Surgical & Life Support Devices": {"Surgical and Intensive Care", "Potential patient or operator injury"},
	}
	d, found := defaults[dt]
	return d[0], d[1], found
}

func slug(s string) string {
	var b strings.Builder
	lastDash := false
	for _, c := range strings.ToLower(NormalizedASCII(s)) {
		if c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			b.WriteRune(c)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" || out == "-" {
		return "device"
	}
	return out
}

func contains(list []string, s string) bool {
	return slices.Contains(list, s)
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// NormalizedASCII strips surviving punctuation/apostrophes to ASCII letters
// and digits (kept simple for slug generation only).
func NormalizedASCII(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '\'' || r == '/' || r == '\\' || r == ',' || r == '.' || r == '(' || r == ')' || r == '&' || r == '-':
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// norm lowercases and space-normalizes a string for stable length ordering.
func norm(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
