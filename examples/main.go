package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ovahol/ontology"
)

func main() {
	// ontology is an engine: it needs a vendor taxonomy. Here we load an
	// example vendor taxonomy (the MeDevIS reference). Vendors ship their own
	// taxonomy in their own codebase and load it with LoadTaxonomyFile.
	tax, err := ontology.LoadTaxonomyFile("examples/taxonomies/medevis.json")
	if err != nil {
		log.Fatalf("load taxonomy: %v", err)
	}

	// Example 1: single device normalization
	fmt.Println("=== Single device ===")
	result := ontology.NormalizeWithTaxonomy(ontology.Input{
		DeviceName: "Patient monitor, multiparameter",
		SourceType: "Monitoring equipment",
	}, tax)
	printResult(result)

	// Example 2: batch from JSON (simulating interchange from another system)
	fmt.Println("\n=== Batch from JSON (WHO/Medevis default vocabulary) ===")
	jsonData := `[
		{"device_name": "Linear accelerator system", "source_type": "Radiotherapy-related equipment"},
		{"device_name": "Infusion pump, volumetric", "source_type": "Infusion devices"},
		{"device_name": "In vitro diagnostic rapid test kit", "source_type": "In vitro diagnostic tests"}
	]`
	results, err := ontology.NormalizeJSONWithTaxonomy([]byte(jsonData), tax)
	if err != nil {
		log.Fatal(err)
	}
	for i, r := range results {
		fmt.Printf("\n--- Device %d ---\n", i+1)
		printResult(r)
	}

	// Example 3: to API import records (deduplicated)
	fmt.Println("\n=== API import records (deduplicated) ===")
	apiRecords := ontology.ToAPIImportRecords(results)
	data, _ := json.MarshalIndent(apiRecords, "", "  ")
	fmt.Println(string(data))

	// Example 4: to CSV
	fmt.Println("\n=== CSV preview (first 3 lines) ===")
	csvData, _ := ontology.ToCSV(results)
	lines := splitLines(string(csvData), 4)
	for _, l := range lines {
		fmt.Println(l)
	}
}

func printResult(r ontology.Result) {
	fmt.Printf("  Input:    %q (%s)\n", r.LegacySourceName, r.SourceType)
	fmt.Printf("  Name:     %s\n", r.Name)
	for k, v := range r.Fields {
		fmt.Printf("  %s: %s\n", k, v)
	}
	fmt.Printf("  Source:   %s (confidence: %s)\n", r.MappingSource, r.Confidence)
}

func splitLines(s string, n int) []string {
	var out []string
	start := 0
	for i := 0; i < len(s) && len(out) < n; i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return out
}
