// Package ontology is a minimal system-agnostic device interchange ontology.
//
// Any system can run its device records through this library and get back
// the 4-field classification Ovahol needs for its device dictionary:
// device_type, device_category, device_function (aka device_application),
// and device_application_risk.
//
// # Quick start
//
//	import "github.com/ovahol/ontology"
//
//	result := ontology.Normalize(ontology.Input{
//	    DeviceName: "ECG machine, portable, 12-lead",
//	    SourceType: "monitoring equipment",
//	    EMDNTerm:   "Electrocardiographs",
//	})
//	// result.Name                  -> "ECG machine"
//	// result.DeviceType            -> "Monitoring & Measurement Devices"
//	// result.DeviceCategory        -> "Diagnostic"
//	// result.DeviceFunction        -> "Additional Physiological Monitoring and Diagnostic"
//	// result.DeviceApplicationRisk -> "Inappropriate therapy or misdiagnosis"
//
// For bulk imports from spreadsheets, use NormalizeWorkbook or NormalizeCSV:
//
//	results, err := ontology.NormalizeWorkbook("legacy_inventory.xlsx", "normalized.xlsx")
//
// # Interchange schema
//
// The library defines a language-agnostic JSON interchange schema so
// non-Go systems can participate by producing/consuming JSON:
//
//	{
//	  "device_name": "Infusion pump, volumetric",
//	  "source_type": "infusion devices",
//	  "emdn_code": "Z1201",
//	  "emdn_term": "Volumetric infusion pumps"
//	}
//
// Normalizes to:
//
//	{
//	  "name": "Infusion pump",
//	  "device_type": "Treatment, Surgical & Life Support Devices",
//	  "device_category": "Therapeutic",
//	  "device_function": "Surgical and Intensive Care",
//	  "device_application_risk": "Potential patient or operator injury",
//	  "mapping_source": "family_fallback"
//	}
//
// Output field names align with Ovahol's lookup tables
// (device_type, device_category, device_function, device_application_risk)
// so the same payload can be used for generic interchange or direct import.
// device_application is an alias for device_function.
//
// # Controlled vocabulary
//
// Taxonomy is intentionally small and stable:
//
//   - 8 device types
//   - 4 device categories (Therapeutic, Diagnostic, Analytical, Miscellaneous)
//   - 9 device functions (each belongs to a category)
//   - 5 application risks
//
// See taxonomy.go for the full lists. Every output value is drawn from
// this vocabulary — no free-text leakage.
package ontology
