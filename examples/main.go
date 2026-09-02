package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ovahol/ontology"
)

func main() {
	// Example 1: single device normalization
	fmt.Println("=== Single device ===")
	result := ontology.Normalize(ontology.Input{
		DeviceName: "ECG machine, portable, 12-lead",
		SourceType: "monitoring equipment",
		EMDNTerm:   "Electrocardiographs",
	})
	printResult(result)

	// Example 2: batch from JSON (simulating interchange from another system)
	fmt.Println("\n=== Batch from JSON ===")
	jsonData := `[
		{"device_name": "Infusion pump, volumetric", "source_type": "infusion devices", "emdn_term": "Volumetric infusion pumps"},
		{"device_name": "Catheter, sterile, single-use, adult", "source_type": "catheters and related", "emdn_term": "Peripheral venous catheters"},
		{"device_name": "Chemistry analyzer", "source_type": "laboratory equipment"},
		{"device_name": "Oxygen concentrator, 5L", "source_type": "medical gas equipment"},
		{"device_name": "Autoclave, benchtop", "source_type": "cleaning disinfection sterilization equipment"}
	]`
	results, err := ontology.NormalizeJSON([]byte(jsonData))
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

	// Example 5: confidence and validation
	fmt.Println("\n=== Validation ===")
	invalid := ontology.Normalize(ontology.Input{
		DeviceName: "Something",
		SourceType: "unknown category xyz",
	})
	fmt.Printf("Unsupported source: valid=%v confidence=%s mapping_source=%s\n",
		invalid.IsValid(), invalid.Confidence, invalid.MappingSource)
}

func printResult(r ontology.Result) {
	fmt.Printf("  Input:    %q (%s)\n", r.LegacySourceName, r.SourceType)
	fmt.Printf("  Common:   %s\n", r.CommonName)
	fmt.Printf("  Canonical:%s\n", r.CanonicalName)
	fmt.Printf("  Aliases:  %s\n", r.SearchAliases)
	fmt.Printf("  Type:     %s\n", r.OvaholType)
	fmt.Printf("  Family:   %s\n", r.Family)
	fmt.Printf("  Function: %s\n", r.Function)
	fmt.Printf("  Risk:     %s\n", r.Risk)
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
