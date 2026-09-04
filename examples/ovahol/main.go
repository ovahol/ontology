// Command examples/ovahol demonstrates seamless migration onto Ovahol's
// device dictionary using the generated ovahol taxonomy plus the device
// dictionary loaded as a reconciliation catalog.
//
// The essence of Ovahol's dictionary: pick the right Model/device and every
// classification dimension (device type, function, application risk, EMDN,
// GMDN) falls out by foreign-key from that one row. ontology reconciles an
// inbound record onto that row when the device name (or EMDN) matches the
// dictionary, and only falls back to keyword inference for records the
// dictionary does not recognise.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ovahol/ontology"
)

func main() {
	// Paths below are relative to the repo root (run with: go run ./examples/ovahol).
	root := "."

	// 1) Load the generated Ovahol taxonomy (fields + rules).
	tax, err := ontology.LoadTaxonomyFile(root + "/examples/taxonomies/ovahol.json")
	if err != nil {
		log.Fatalf("load taxonomy: %v", err)
	}

	// 2) Load the device dictionary as a reconciliation catalog. Each CSV
	//    column is mapped to its CatalogEntry field; "name" is required.
	f, err := os.Open(root + "/devices.csv")
	if err != nil {
		log.Fatalf("open devices.csv: %v", err)
	}
	defer f.Close()
	cat, err := (ontology.CSVCatalog{Columns: map[string]string{
		"name":                    "Name",
		"device_type":             "Device Type",
		"device_function":         "Device Function",
		"device_application_risk": "Application Risk",
		"emdn_code":               "EMDN Code",
		"emdn_term":               "EMDN Term",
	}}).Load(f)
	if err != nil {
		log.Fatalf("load catalog: %v", err)
	}

	// 3) Reconcile inbound records. Exact matches hit the catalog
	//    (catalog_exact / high); everything else falls back to taxonomy rules.
	//
	// modelName/brandName/manufacturerName would normally be pulled from the
	// source system's model-row FK chain. When they are missing, the Unknown
	// conventions fill them in.
	inputs := []struct {
		in    ontology.Input
		model string
		brand string
		mfr   string
	}{
		// Exact dictionary device name, known model FK -> verbatim row + names.
		{in: ontology.Input{DeviceName: "Infusion pump", SourceType: "infusion devices"}, model: "Volumetric 10", brand: "Bbraun", mfr: "B. Braun Melsungen AG"},
		// Exact name, but no model/brand/manufacturer resolved -> Unknown.
		{in: ontology.Input{DeviceName: "Multiparametric basic patient monitor", SourceType: "monitoring equipment"}},
		// Recognisable generic name (not a verbatim dictionary name) ->
		// classified by taxonomy rules into a valid type. Unresolved ids fall
		// back to Unknown.
		{in: ontology.Input{DeviceName: "ECG machine, 12-lead", SourceType: "monitoring equipment"}},
		// Device present only by EMDN code -> matched by EMDN.
		{in: ontology.Input{DeviceName: "Chemistry analyzer", SourceType: "laboratory equipment", EMDNCode: "W0201019901"}, model: "AU480", brand: "Beckman", mfr: "Beckman Coulter"},
		// Completely unknown device -> Unknown Device + Unknown Model.
		{in: ontology.Input{DeviceName: "Oscillating Thingamajig", SourceType: "mystery category"}},
	}

	for _, c := range inputs {
		result := ontology.NormalizeWithCatalogAndTaxonomy(c.in, cat, tax)
		// The engine returns catalog hits verbatim (system-agnostic: it does
		// not invent dimensions the catalog lacks). Ovahol's own function->
		// category vocabulary is a vendor concern, so the ovahol layer derives
		// it explicitly here when the dictionary row did not carry it.
		category := result.GetField(ontology.FieldDeviceCategory)
		function := result.GetField(ontology.FieldDeviceFunction)
		if category == "" && strings.TrimSpace(function) != "" {
			category = ontology.CategoryForFunctionFor(function, tax)
		}
		rec := ApplyUnknownConventions(result.Name, c.model, c.brand, c.mfr, result.EMDNCode, result.EMDNTerm)
		rec.DeviceType = result.GetField(ontology.FieldDeviceType)
		rec.DeviceCategory = category
		rec.DeviceFunction = function
		rec.ApplicationRisk = result.GetField(ontology.FieldDeviceApplicationRisk)
		rec.MappingSource = result.MappingSource
		rec.Confidence = result.Confidence
		fmt.Println("------------------------------------------------------------")
		fmt.Printf("Input:     %q (source=%q)\n", c.in.DeviceName, c.in.SourceType)
		fmt.Printf("Device:    %s\n", rec.Name)
		fmt.Printf("Model:     %s\n", rec.Model)
		fmt.Printf("Brand:     %s\n", rec.Brand)
		fmt.Printf("Mfr:       %s\n", rec.Manufacturer)
		fmt.Printf("Type:      %s | Category: %s\n", rec.DeviceType, rec.DeviceCategory)
		fmt.Printf("Function:  %s | Risk: %s\n", rec.DeviceFunction, rec.ApplicationRisk)
		fmt.Printf("EMDN:      %s / %s\n", rec.EMDNCode, rec.EMDNTerm)
		fmt.Printf("Source:    %s (confidence=%s)\n", rec.MappingSource, rec.Confidence)
	}

	// Raw dump of the catalog size for reference.
	fmt.Println("------------------------------------------------------------")
	fmt.Println("Device dictionary loaded as catalog; inbound records are")
	fmt.Println("reconciled verbatim when the name/EMDN matches it. Identity")
	fmt.Println("fields that cannot be resolved fall back to the Unknown")
	fmt.Println("convention (Unknown device/brand/manufacturer/model).")
}
