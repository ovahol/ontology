package ontology

// Input is one raw device record from any external system.
// Every field is optional and free-text — the library normalizes it.
// Populate whatever you have; empty fields are treated as absent.
//
// For JSON interchange, field names are snake_case to match Ovahol's API
// convention. A minimal valid input is just DeviceName or SourceType or
// EMDNTerm — any single signal is enough to attempt classification.
type Input struct {
	// DeviceName is the free-text device name as it appears in the source
	// system. Examples: "ECG machine, portable", "Catheter, sterile, single-use",
	// "B. Braun Space pump".
	DeviceName string `json:"device_name,omitempty"`

	// SourceType is the source system's category/type for the device.
	// Examples: "monitoring equipment", "infusion devices", "laboratory equipment".
	// If your system has no explicit type, leave empty.
	SourceType string `json:"source_type,omitempty"`

	// EMDNCode is the European Medical Device Nomenclature code, if known.
	// Example: "Z12010501". Used for passthrough, not inference.
	EMDNCode string `json:"emdn_code,omitempty"`

	// EMDNTerm is the EMDN term/descriptor for the device.
	// Example: "Electrocardiographs". Used alongside DeviceName for inference.
	EMDNTerm string `json:"emdn_term,omitempty"`
}

// Result is the normalized Ovahol-standard record for one input device.
// Every Ovahol-controlled field is drawn from the controlled vocabulary
// defined in taxonomy.go — no free-text values leak into these fields
// except LegacySourceName/SourceType/EMDN which echo the input.
//
// For JSON interchange, field names are snake_case.
type Result struct {
	// CommonName is the short, clinician-friendly name (2-5 words, singular,
	// manufacturer-neutral). Example: "ECG machine", "Infusion pump".
	CommonName string `json:"common_name"`

	// CanonicalName is the precise generic descriptor. Example:
	// "Electrocardiography system", "Infusion or syringe pump system".
	CanonicalName string `json:"canonical_name"`

	// SearchAliases is a comma-separated list of alternate names, abbreviations,
	// and local terms. Example: "ECG machine, EKG machine, Electrocardiograph".
	SearchAliases string `json:"search_aliases"`

	// OvaholType is one of the 8 Ovahol device types. Example:
	// "Monitoring & Measurement Devices".
	OvaholType string `json:"ovahol_device_type"`

	// Family is the Ovahol device family within the type. Example:
	// "Cardiac diagnostic systems".
	Family string `json:"ovahol_device_family"`

	// Function is the device function. Example:
	// "Additional Physiological Monitoring and Diagnostic".
	Function string `json:"device_function"`

	// Risk is the device application risk. Example:
	// "Inappropriate therapy or misdiagnosis".
	Risk string `json:"device_application_risk"`

	// LegacySourceName echoes the original DeviceName input.
	LegacySourceName string `json:"legacy_source_name"`

	// SourceType echoes the original SourceType input.
	SourceType string `json:"source_type"`

	// EMDNCode echoes the original EMDNCode input.
	EMDNCode string `json:"emdn_code"`

	// EMDNTerm echoes the original EMDNTerm input.
	EMDNTerm string `json:"emdn_term"`

	// MappingSource explains how the name was derived. One of:
	//   "specific_rule"            — matched a SpecificNameRule
	//   "legacy_derived"           — parsed from the legacy device name
	//   "family_fallback"          — fell back to the family default name
	//   "unsupported_source_type"  — source type is not in SupportedSourceTypes
	MappingSource string `json:"mapping_source"`

	// Confidence indicates how much signal was available:
	//   "high"   — specific rule or strong keyword match
	//   "medium" — family keyword match
	//   "low"    — type-level default only
	//   "none"   — unsupported source type, no mapping possible
	Confidence string `json:"confidence"`
}

// InterchangeRecord is the full row shape used in workbook interchange
// (Excel/CSV). It combines Input fields with Result fields so a single
// row carries both the legacy source and the normalized output, matching
// the Devices sheet layout in the reference workbook.
//
// This is the canonical interchange schema for bulk migration: any system
// can produce a spreadsheet or CSV with at least LegacySourceName and
// SourceType columns, run it through NormalizeWorkbook/NormalizeCSV, and
// get back a fully populated workbook ready for Ovahol API import.
type InterchangeRecord struct {
	CommonName    string `json:"common_name" csv:"Common name"`
	CanonicalName string `json:"canonical_name" csv:"Canonical device name"`
	SearchAliases string `json:"search_aliases" csv:"Search aliases"`
	OvaholType    string `json:"ovahol_device_type" csv:"Ovahol device type"`
	Family        string `json:"ovahol_device_family" csv:"Ovahol device family"`
	Function      string `json:"device_function" csv:"Device function"`
	Risk          string `json:"device_application_risk" csv:"Device application risk"`

	LegacySourceName string `json:"legacy_source_name" csv:"Legacy source name"`
	SourceType       string `json:"source_type" csv:"Source device type"`
	EMDNCode         string `json:"emdn_code" csv:"EMDN code"`
	EMDNTerm         string `json:"emdn_term" csv:"EMDN term"`

	MappingSource string `json:"mapping_source" csv:"Mapping source"`
}

// APIImportRecord is the deduplicated, API-ready shape derived from
// InterchangeRecords. It contains only the fields Ovahol's device-model
// creation API requires — the minimal interchange for programmatic
// migration without a workbook.
type APIImportRecord struct {
	Name                  string `json:"name"`
	DeviceType            string `json:"device_type"`
	DeviceFunction        string `json:"device_function"`
	DeviceApplicationRisk string `json:"device_application_risk"`
	EMDNCode              string `json:"emdn_code"`
	EMDNTerm              string `json:"emdn_term"`
}

// ToAPIImportRecord converts a Result to its API import shape.
func (r Result) ToAPIImportRecord() APIImportRecord {
	return APIImportRecord{
		Name:                  r.CommonName,
		DeviceType:            r.OvaholType,
		DeviceFunction:        r.Function,
		DeviceApplicationRisk: r.Risk,
		EMDNCode:              r.EMDNCode,
		EMDNTerm:              r.EMDNTerm,
	}
}

// ToInterchangeRecord converts a Result to its workbook row shape.
func (r Result) ToInterchangeRecord() InterchangeRecord {
	return InterchangeRecord{
		CommonName:       r.CommonName,
		CanonicalName:    r.CanonicalName,
		SearchAliases:    r.SearchAliases,
		OvaholType:       r.OvaholType,
		Family:           r.Family,
		Function:         r.Function,
		Risk:             r.Risk,
		LegacySourceName: r.LegacySourceName,
		SourceType:       r.SourceType,
		EMDNCode:         r.EMDNCode,
		EMDNTerm:         r.EMDNTerm,
		MappingSource:    r.MappingSource,
	}
}

// IsValid reports whether the result contains a usable Ovahol mapping.
// An unsupported source type yields an invalid result (all Ovahol fields empty).
func (r Result) IsValid() bool {
	return r.OvaholType != "" && r.MappingSource != "unsupported_source_type"
}
