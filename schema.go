package ontology

// Input is one raw device record from any external system.
// Every field is optional and free-text — the library normalizes it.
type Input struct {
	DeviceName string `json:"device_name,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	EMDNCode   string `json:"emdn_code,omitempty"`
	EMDNTerm   string `json:"emdn_term,omitempty"`
}

// Result is the minimal system-agnostic classification for one input.
// 4-field taxonomy + normalized name. Input echo fields are kept for
// workbook/CSV audit but are not part of the 4-field classification.
type Result struct {
	// Name is the normalized device name. Example: "ECG machine".
	Name string `json:"name"`

	// DeviceType is one of the 8 device types. Example: "Treatment, Surgical & Life Support Devices".
	DeviceType string `json:"device_type"`

	// DeviceCategory is one of the 4 high-level categories. Example: "Therapeutic", "Diagnostic", "Analytical", "Miscellaneous".
	// Derived from DeviceFunction's category.
	DeviceCategory string `json:"device_category"`

	// DeviceFunction is the device function / application. Example: "Life Support", "Surgical and Intensive Care".
	// Also exposed as device_application for backward compat with earlier naming.
	DeviceFunction string `json:"device_function"`

	// DeviceApplication is an alias for DeviceFunction (kept for the "device application" wording).
	// It always equals DeviceFunction. Use DeviceFunction going forward.
	DeviceApplication string `json:"device_application,omitempty"`

	// DeviceApplicationRisk is the application risk. Example: "Potential patient death".
	DeviceApplicationRisk string `json:"device_application_risk"`

	// Input echo — kept for interchange/audit, not part of device dictionary.
	LegacySourceName string `json:"legacy_source_name,omitempty"`
	SourceType       string `json:"source_type,omitempty"`
	EMDNCode         string `json:"emdn_code,omitempty"`
	EMDNTerm         string `json:"emdn_term,omitempty"`

	// Confidence and MappingSource are optional diagnostics (high/medium/low/none).
	Confidence    string `json:"confidence,omitempty"`
	MappingSource string `json:"mapping_source,omitempty"`
}

// InterchangeRecord is the row shape for workbook/CSV interchange (input + 4-field output).
type InterchangeRecord struct {
	Name                  string `json:"name" csv:"Name"`
	DeviceType            string `json:"device_type" csv:"Device type"`
	DeviceCategory        string `json:"device_category" csv:"Device category"`
	DeviceFunction        string `json:"device_function" csv:"Device function"`
	DeviceApplicationRisk string `json:"device_application_risk" csv:"Device application risk"`

	LegacySourceName string `json:"legacy_source_name" csv:"Legacy source name"`
	SourceType       string `json:"source_type" csv:"Source device type"`
	EMDNCode         string `json:"emdn_code" csv:"EMDN code"`
	EMDNTerm         string `json:"emdn_term" csv:"EMDN term"`

	MappingSource string `json:"mapping_source" csv:"Mapping source"`
	Confidence    string `json:"confidence" csv:"Confidence"`
}

// APIImportRecord is the deduplicated API-ready shape for Ovahol.
type APIImportRecord struct {
	Name                  string `json:"name"`
	DeviceType            string `json:"device_type"`
	DeviceCategory        string `json:"device_category"`
	DeviceFunction        string `json:"device_function"`
	DeviceApplicationRisk string `json:"device_application_risk"`
	EMDNCode              string `json:"emdn_code,omitempty"`
	EMDNTerm              string `json:"emdn_term,omitempty"`
}

func (r Result) ToAPIImportRecord() APIImportRecord {
	return APIImportRecord{
		Name:                  r.Name,
		DeviceType:            r.DeviceType,
		DeviceCategory:        r.DeviceCategory,
		DeviceFunction:        r.DeviceFunction,
		DeviceApplicationRisk: r.DeviceApplicationRisk,
		EMDNCode:              r.EMDNCode,
		EMDNTerm:              r.EMDNTerm,
	}
}

func (r Result) ToInterchangeRecord() InterchangeRecord {
	return InterchangeRecord{
		Name:                  r.Name,
		DeviceType:            r.DeviceType,
		DeviceCategory:        r.DeviceCategory,
		DeviceFunction:        r.DeviceFunction,
		DeviceApplicationRisk: r.DeviceApplicationRisk,
		LegacySourceName:      r.LegacySourceName,
		SourceType:            r.SourceType,
		EMDNCode:              r.EMDNCode,
		EMDNTerm:              r.EMDNTerm,
		MappingSource:         r.MappingSource,
		Confidence:            r.Confidence,
	}
}

func (r Result) IsValid() bool {
	return r.DeviceType != "" && r.MappingSource != "unsupported_source_type"
}

// SearchAliasesString is kept for compat (now no-op).
func (r Result) SearchAliasesString() string { return "" }
