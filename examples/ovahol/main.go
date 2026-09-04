// Command examples/ovahol demonstrates real-world integration of the ontology
// engine onto Ovahol's device dictionary and identity model.
//
// The engine is system-agnostic: it only executes what the vendor supplies.
// This example shows Ovahol supplying, in ONE place (its app layer):
//
//   - its taxonomy (fields + rules)          -> engine.NewEngine's taxonomy arg
//   - its device dictionary as a Catalog     -> engine.WithCatalog
//   - its identity conventions (the "Unknown"
//     placeholders + the controlled status set) -> engine.WithConventions
//   - an identity resolver that walks Ovahol's
//     FK chain (model -> brand -> manufacturer)  -> engine.WithIdentityResolver
//
// None of the Ovahol strings below are shipped by the engine; they live here,
// in the vendor's code, exactly as they'd live in the real Ovahol app.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ovahol/ontology"
)

// ovaholConventions is Ovahol's own identity conventions: the canonical
// "Unknown" placeholders and its controlled status vocabulary. These mirror
// what Ovahol's lifecycle exports already contain
// (device-records-lifecycle-*.json, models.csv, brands.csv, manufacturers.csv).
// The status list is Ovahol's real controlled status set.
var ovaholConventions = ontology.Conventions{
	UnknownDevice:       "Unknown Device",
	UnknownBrand:        "Unknown Brand",
	UnknownManufacturer: "Unknown Manufacturer",
	UnknownModelPrefix:  "Unknown Model - ",
	Statuses: []string{
		"In-Service",
		"Decommissioned",
		"Transferred",
		"Standby / Spare",
		"Under Maintenance",
		"Out of Service",
		"Disposed",
		"New / Commissioning",
	},
	// Source inventories use many free-text terms for a status. Every term here
	// reconciles to Ovahol's canonical status (case-insensitively). Anything
	// unmatched falls back to DefaultStatus.
	StatusSynonyms: map[string][]string{
		"In-Service":        {"functional", "functioning", "active", "in active service", "working"},
		"Under Maintenance": {"faulty", "not working", "broken down", "broken down and repairable", "maintenance"},
		"Out of Service":    {"out-of-service", "down", "offline"},
		"Standby / Spare":   {"standby", "spare", "reserve"},
		"Disposed":          {"discarded", "scrapped", "sold as scrap"},
		"Transferred":       {"relocated", "moved"},
	},
	DefaultStatus: "New / Commissioning", // conservative default for a freshly imported record
}

// ovaholResolver canonicalizes an inbound record's identity by walking
// Ovahol's FK chain: model -> brand -> manufacturer (and model -> device).
// It is Ovahol's own reference-data lookup; the engine only asks it for the
// canonical tuple, then applies Unknown conventions to what it can't resolve.
type ovaholResolver struct {
	modelByBrand map[string]string // model name -> brand name
	modelDevice  map[string]string // model name -> device name
	brandMfr     map[string]string // brand name -> manufacturer name
}

func (r *ovaholResolver) Resolve(in ontology.IdentityInput) (ontology.IdentityResult, bool) {
	model := strings.TrimSpace(in.Model)
	brand := r.modelByBrand[model]
	if brand == "" {
		// Unknown model: nothing resolves; the engine fills Unknown placeholders.
		return ontology.IdentityResult{}, false
	}
	mfr := r.brandMfr[brand]
	if mfr == "" {
		mfr = in.Manufacturer
	}
	return ontology.IdentityResult{
		Device:       r.modelDevice[model],
		Model:        model,
		Brand:        brand,
		Manufacturer: mfr,
	}, true
}

func main() {
	root := "."

	// Ovahol supplies its taxonomy (fields + rules). The taxonomy is Ovahol's
	// own artifact — this library does not ship it — so the example reads it
	// from wherever Ovahol's app keeps it, via $OVAHOL_TAXONOMY.
	taxPath := os.Getenv("OVAHOL_TAXONOMY")
	if taxPath == "" {
		log.Fatal("set OVAHOL_TAXONOMY to the path of your Ovahol taxonomy JSON " +
			"(e.g. the taxonomy export from the ovahol app); this library does not vendor it")
	}
	tax, err := ontology.LoadTaxonomyFile(taxPath)
	if err != nil {
		log.Fatalf("load taxonomy: %v", err)
	}

	// Ovahol supplies its device dictionary as a reconciliation catalog.
	deviceFile, err := os.Open(root + "/devices.csv")
	if err != nil {
		log.Fatalf("open devices.csv: %v", err)
	}
	defer deviceFile.Close()
	cat, err := (ontology.CSVCatalog{Columns: map[string]string{
		"name":                    "Name",
		"device_type":             "Device Type",
		"device_function":         "Device Function",
		"device_application_risk": "Application Risk",
		"emdn_code":               "EMDN Code",
		"emdn_term":               "EMDN Term",
	}}).Load(deviceFile)
	if err != nil {
		log.Fatalf("load catalog: %v", err)
	}

	// Ovahol supplies its identity reference data (models/brands/manufacturers).
	res := &ovaholResolver{
		modelByBrand: map[string]string{},
		modelDevice:  map[string]string{},
		brandMfr:     map[string]string{},
	}
	loadIdentityData(res, root)

	// Bind everything to one engine — the whole integration in one place.
	engine := ontology.NewEngine(tax,
		ontology.WithCatalog(cat),
		ontology.WithIdentityResolver(res),
		ontology.WithConventions(ovaholConventions),
	)

	inputs := []struct {
		in     ontology.Input
		model  string
		brand  string
		mfr    string
		status string
	}{
		// Exact dictionary device name + a real model FK. The source record's
		// brand/manufacturer are intentionally wrong; the resolver canonicalizes
		// them from the reference data (Plum 360 -> Medtronic -> Medtronic plc).
		{in: ontology.Input{DeviceName: "Infusion pump", SourceType: "infusion devices"}, model: "Plum 360", brand: "medtronic", mfr: "some legacy mfr", status: "in-service"},
		// Exact name, but no model/brand/manufacturer -> Unknown.
		{in: ontology.Input{DeviceName: "Multiparametric basic patient monitor", SourceType: "monitoring equipment"}},
		// Recognisable generic name (not a verbatim dictionary name) ->
		// classified by taxonomy rules into a valid type, identity Unknown.
		{in: ontology.Input{DeviceName: "ECG machine, 12-lead", SourceType: "monitoring equipment"}},
		// Device present only by EMDN code -> matched by EMDN; model FK resolves
		// a real chain (BS-200 -> Mindray -> Mindray Bio-Medical Electronics).
		{in: ontology.Input{DeviceName: "Chemistry analyzer", SourceType: "laboratory equipment", EMDNCode: "W0201019901"}, model: "BS-200", brand: "Mindray", status: "under maintenance"},
		// Completely unknown -> Unknown Device + Unknown Model.
		{in: ontology.Input{DeviceName: "Oscillating Thingamajig", SourceType: "mystery category"}},
	}

	for _, c := range inputs {
		// Step 1: classify (catalog-exact, then taxonomy rules).
		result := engine.Normalize(c.in)

		// Step 2: reconcile identity (FK resolution + Unknown conventions +
		// status normalization), threading the classified name as the device.
		id := engine.Reconcile(ontology.IdentityInput{
			DeviceName:   result.Name,
			Model:        c.model,
			Brand:        c.brand,
			Manufacturer: c.mfr,
			Status:       c.status,
		})

		// Ovahol's own function->category vocabulary is a vendor concern, so
		// it is derived explicitly when the dictionary row did not carry it.
		category := result.GetField(ontology.FieldDeviceCategory)
		if category == "" {
			category = ontology.CategoryForFunctionFor(result.GetField(ontology.FieldDeviceFunction), tax)
		}

		fmt.Println("------------------------------------------------------------------")
		fmt.Printf("Input:      %q (source=%q)\n", c.in.DeviceName, c.in.SourceType)
		fmt.Printf("Device:     %s\n", id.Device)
		fmt.Printf("Model:      %s\n", id.Model)
		fmt.Printf("Brand:      %s\n", id.Brand)
		fmt.Printf("Mfr:        %s\n", id.Manufacturer)
		fmt.Printf("Status:     %s\n", id.Status)
		fmt.Printf("Type:       %s\n", result.GetField(ontology.FieldDeviceType))
		fmt.Printf("Category:   %s\n", category)
		fmt.Printf("Function:   %s\n", result.GetField(ontology.FieldDeviceFunction))
		fmt.Printf("Risk:       %s\n", result.GetField(ontology.FieldDeviceApplicationRisk))
		fmt.Printf("EMDN:       %s / %s\n", result.EMDNCode, result.EMDNTerm)
		fmt.Printf("Source:     %s (confidence=%s)\n", result.MappingSource, result.Confidence)
	}

	fmt.Println("------------------------------------------------------------------")
	fmt.Println("Everything Ovahol-specific (taxonomy, dictionary, identity FK, Unknown")
	fmt.Println("conventions, status set) is supplied by the vendor and bound to the")
	fmt.Println("engine via options — the engine hardcodes none of it.")
}

// loadIdentityData parses Ovahol's reference CSVs (models/brands/manufacturers)
// into the resolver's FK index. In the real app this is a DB lookup; here it is
// a CSV snapshot for the example.
func loadIdentityData(r *ovaholResolver, root string) {
	for _, rec := range readCSV(root + "/models.csv") {
		if len(rec) < 3 {
			continue
		}
		name := strings.TrimSpace(rec[0])
		if name == "" || name == "Name" {
			continue
		}
		r.modelByBrand[name] = strings.TrimSpace(rec[1])
		r.modelDevice[name] = strings.TrimSpace(rec[2])
	}
	for _, rec := range readCSV(root + "/brands.csv") {
		if len(rec) < 2 {
			continue
		}
		name := strings.TrimSpace(rec[0])
		if name == "" || name == "Name" {
			continue
		}
		r.brandMfr[name] = strings.TrimSpace(rec[1])
	}
}

// readCSV reads a CSV file into rows of raw string fields.
func readCSV(path string) [][]string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Fatalf("parse %s: %v", path, err)
	}
	return rows
}
