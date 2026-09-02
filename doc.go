// Package ontology is the Ovahol interchange ontology — a Go library that
// normalizes arbitrary device data from any healthcare system into Ovahol's
// controlled vocabulary.
//
// Any system hoping to migrate onto Ovahol (or exchange data with it) can
// run its device records through this library and get back canonical Ovahol
// terminology: common name, canonical device name, search aliases, device
// type, device family, device function, and application risk.
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
//	// result.CommonName       -> "ECG machine"
//	// result.CanonicalName    -> "Electrocardiography system"
//	// result.OvaholType       -> "Monitoring & Measurement Devices"
//	// result.Family           -> "Cardiac diagnostic systems"
//	// result.Function         -> "Additional Physiological Monitoring and Diagnostic"
//	// result.Risk             -> "Inappropriate therapy or misdiagnosis"
//	// result.SearchAliases    -> "ECG machine, EKG machine, Electrocardiograph, ..."
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
//	  "common_name": "Infusion pump",
//	  "canonical_name": "Infusion or syringe pump system",
//	  "search_aliases": "Infusion pump, Infusion or syringe pump system, IV pump",
//	  "ovahol_device_type": "Treatment, Surgical & Life Support Devices",
//	  "ovahol_device_family": "Infusion and medication delivery systems",
//	  "device_function": "Surgical and Intensive Care",
//	  "device_application_risk": "Potential patient or operator injury",
//	  "mapping_source": "family_fallback"
//	}
//
// # Controlled vocabulary
//
// Ovahol's taxonomy is intentionally small and stable:
//
//   - 8 device types (e.g. "Diagnostic & Imaging Devices")
//   - ~120 device families (e.g. "Ultrasound systems")
//   - 9 device functions (e.g. "Life Support")
//   - 5 application risks (e.g. "Potential patient death")
//
// See taxonomy.go for the full lists. Every output value is drawn from
// this vocabulary — no free-text leakage.
//
// # Port provenance
//
// Ported from scripts/update_ovahol_ontology.py (Python/openpyxl) to pure
// Go. The Python script remains as the taxonomy generator; this library is
// the runtime normalization path used during facility import and by any
// external migrator.
package ontology
