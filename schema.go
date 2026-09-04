package ontology

import (
	"sort"
	"strings"
)

// Input is one raw device record from any external system.
// Every field is optional and free-text — the library normalizes it.
type Input struct {
	DeviceName string `json:"device_name,omitempty"`
	SourceType string `json:"source_type,omitempty"`
	EMDNCode   string `json:"emdn_code,omitempty"`
	EMDNTerm   string `json:"emdn_term,omitempty"`
}

// Field keys for the pluggable Fields map. The 4 fixed dimensions remain the
// canonical vocabulary, but callers may extend Fields without schema changes.
const (
	FieldDeviceType            = "device_type"
	FieldDeviceCategory        = "device_category"
	FieldDeviceFunction        = "device_function"
	FieldDeviceApplicationRisk = "device_application_risk"
	FieldDeviceFamily          = "device_family"
)

// Result is the minimal system-agnostic classification for one input.
// Fields is the pluggable primary storage; fixed accessors are deprecated shims
// that proxy to Fields for backward compatibility.
type Result struct {
	// Name is the normalized device name. Example: "ECG machine".
	Name string `json:"name"`

	// Fields holds taxonomy dimensions. Keys use Field* constants above.
	// For backward compat the same values are also mirrored to the deprecated
	// fixed fields (DeviceType, DeviceCategory, etc.) so existing callers
	// reading struct fields continue to work. Prefer Fields going forward.
	Fields map[string]string `json:"fields,omitempty"`

	// Deprecated: use Fields[FieldDeviceType] or GetField(FieldDeviceType).
	DeviceType string `json:"device_type"`

	// Deprecated: use Fields[FieldDeviceCategory].
	DeviceCategory string `json:"device_category"`

	// Deprecated: use Fields[FieldDeviceFunction].
	DeviceFunction string `json:"device_function"`

	// Deprecated: alias for DeviceFunction (kept for the "device application" wording).
	// It always equals DeviceFunction. Use Fields[FieldDeviceFunction] going forward.
	DeviceApplication string `json:"device_application,omitempty"`

	// Deprecated: use Fields[FieldDeviceApplicationRisk].
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

// GetField returns Fields[key] if present, otherwise falls back to the
// deprecated fixed field for that key (so old JSON without "fields" still works).
func (r Result) GetField(key string) string {
	if r.Fields != nil {
		if v, ok := r.Fields[key]; ok && v != "" {
			return v
		}
	}
	switch key {
	case FieldDeviceType:
		return r.DeviceType
	case FieldDeviceCategory:
		return r.DeviceCategory
	case FieldDeviceFunction:
		return r.DeviceFunction
	case FieldDeviceApplicationRisk:
		return r.DeviceApplicationRisk
	case FieldDeviceFamily:
		// Family not stored on Result historically; check Fields only.
		return ""
	}
	return ""
}

// SetField sets Fields[key] and mirrors to the deprecated fixed field if it is
// a known taxonomy dimension.
func (r *Result) SetField(key, value string) {
	if r.Fields == nil {
		r.Fields = make(map[string]string)
	}
	r.Fields[key] = value
	switch key {
	case FieldDeviceType:
		r.DeviceType = value
	case FieldDeviceCategory:
		r.DeviceCategory = value
	case FieldDeviceFunction:
		r.DeviceFunction = value
		r.DeviceApplication = value
	case FieldDeviceApplicationRisk:
		r.DeviceApplicationRisk = value
	}
}

// syncFieldsFromFixed populates Fields from fixed fields when Fields is nil
// or missing keys (compat for JSON without "fields").
func (r *Result) syncFieldsFromFixed() {
	if r.Fields == nil {
		r.Fields = make(map[string]string)
	}
	if r.DeviceType != "" && r.Fields[FieldDeviceType] == "" {
		r.Fields[FieldDeviceType] = r.DeviceType
	}
	if r.DeviceCategory != "" && r.Fields[FieldDeviceCategory] == "" {
		r.Fields[FieldDeviceCategory] = r.DeviceCategory
	}
	if r.DeviceFunction != "" && r.Fields[FieldDeviceFunction] == "" {
		r.Fields[FieldDeviceFunction] = r.DeviceFunction
	}
	if r.DeviceApplicationRisk != "" && r.Fields[FieldDeviceApplicationRisk] == "" {
		r.Fields[FieldDeviceApplicationRisk] = r.DeviceApplicationRisk
	}
}

// syncFixedFromFields mirrors Fields back to deprecated fixed fields (used after
// constructing via Fields). Keeps both representations consistent.
func (r *Result) syncFixedFromFields() {
	if r.Fields == nil {
		return
	}
	if v, ok := r.Fields[FieldDeviceType]; ok {
		r.DeviceType = v
	}
	if v, ok := r.Fields[FieldDeviceCategory]; ok {
		r.DeviceCategory = v
	}
	if v, ok := r.Fields[FieldDeviceFunction]; ok {
		r.DeviceFunction = v
		r.DeviceApplication = v
	}
	if v, ok := r.Fields[FieldDeviceApplicationRisk]; ok {
		r.DeviceApplicationRisk = v
	}
}

// InterchangeRecord is the row shape for workbook/CSV interchange (input + 4-field output).
// Fields enables dynamic columns beyond the 4 fixed ones.
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

	// Fields holds any additional dynamic columns (including the 4 fixed
	// dimensions mirrored for forward compat). When present, CSV/workbook
	// writers may emit extra columns.
	Fields map[string]string `json:"fields,omitempty" csv:"-"`
}

// GetField for InterchangeRecord (mirrors Result.GetField semantics).
func (r InterchangeRecord) GetField(key string) string {
	if r.Fields != nil {
		if v, ok := r.Fields[key]; ok && v != "" {
			return v
		}
	}
	switch key {
	case FieldDeviceType:
		return r.DeviceType
	case FieldDeviceCategory:
		return r.DeviceCategory
	case FieldDeviceFunction:
		return r.DeviceFunction
	case FieldDeviceApplicationRisk:
		return r.DeviceApplicationRisk
	}
	return ""
}

// APIImportRecord is the deduplicated API-ready shape for Ovahol.
// Fields enables pluggable taxonomy dimensions; fixed fields remain as deprecated shims.
type APIImportRecord struct {
	Name                  string `json:"name"`
	DeviceType            string `json:"device_type"`
	DeviceCategory        string `json:"device_category"`
	DeviceFunction        string `json:"device_function"`
	DeviceApplicationRisk string `json:"device_application_risk"`
	EMDNCode              string `json:"emdn_code,omitempty"`
	EMDNTerm              string `json:"emdn_term,omitempty"`
	Fields                map[string]string `json:"fields,omitempty"`
}

// GetField for APIImportRecord.
func (r APIImportRecord) GetField(key string) string {
	if r.Fields != nil {
		if v, ok := r.Fields[key]; ok && v != "" {
			return v
		}
	}
	switch key {
	case FieldDeviceType:
		return r.DeviceType
	case FieldDeviceCategory:
		return r.DeviceCategory
	case FieldDeviceFunction:
		return r.DeviceFunction
	case FieldDeviceApplicationRisk:
		return r.DeviceApplicationRisk
	}
	return ""
}

// DedupKey returns a generalized deduplication key for an API import record.
// It generalizes the old hardcoded 4-field join (Name + 4 taxonomy fields)
// to include any additional Fields entries sorted by key, so pluggable
// dimensions participate in deduplication without code changes.
func (r APIImportRecord) DedupKey() string {
	// Collect all taxonomy-relevant keys: fixed + dynamic.
	fieldMap := make(map[string]string)
	if r.Fields != nil {
		for k, v := range r.Fields {
			fieldMap[k] = v
		}
	}
	// Ensure fixed fields are represented even if Fields missing them (compat).
	if _, ok := fieldMap[FieldDeviceType]; !ok && r.DeviceType != "" {
		fieldMap[FieldDeviceType] = r.DeviceType
	}
	if _, ok := fieldMap[FieldDeviceCategory]; !ok && r.DeviceCategory != "" {
		fieldMap[FieldDeviceCategory] = r.DeviceCategory
	}
	if _, ok := fieldMap[FieldDeviceFunction]; !ok && r.DeviceFunction != "" {
		fieldMap[FieldDeviceFunction] = r.DeviceFunction
	}
	if _, ok := fieldMap[FieldDeviceApplicationRisk]; !ok && r.DeviceApplicationRisk != "" {
		fieldMap[FieldDeviceApplicationRisk] = r.DeviceApplicationRisk
	}
	keys := make([]string, 0, len(fieldMap))
	for k := range fieldMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+1)
	parts = append(parts, r.Name)
	for _, k := range keys {
		parts = append(parts, k+"="+fieldMap[k])
	}
	return strings.Join(parts, "\x00")
}

func (r Result) ToAPIImportRecord() APIImportRecord {
	fields := make(map[string]string)
	if r.Fields != nil {
		for k, v := range r.Fields {
			fields[k] = v
		}
	}
	// Ensure canonical 4 are present (compat: if Fields empty, use fixed).
	if _, ok := fields[FieldDeviceType]; !ok && r.DeviceType != "" {
		fields[FieldDeviceType] = r.DeviceType
	}
	if _, ok := fields[FieldDeviceCategory]; !ok && r.DeviceCategory != "" {
		fields[FieldDeviceCategory] = r.DeviceCategory
	}
	if _, ok := fields[FieldDeviceFunction]; !ok && r.DeviceFunction != "" {
		fields[FieldDeviceFunction] = r.DeviceFunction
	}
	if _, ok := fields[FieldDeviceApplicationRisk]; !ok && r.DeviceApplicationRisk != "" {
		fields[FieldDeviceApplicationRisk] = r.DeviceApplicationRisk
	}
	// Remove empty family if not set.
	if v, ok := fields[FieldDeviceFamily]; ok && v == "" {
		delete(fields, FieldDeviceFamily)
	}
	if len(fields) == 0 {
		fields = nil
	}
	return APIImportRecord{
		Name:                  r.Name,
		DeviceType:            r.GetField(FieldDeviceType),
		DeviceCategory:        r.GetField(FieldDeviceCategory),
		DeviceFunction:        r.GetField(FieldDeviceFunction),
		DeviceApplicationRisk: r.GetField(FieldDeviceApplicationRisk),
		EMDNCode:              r.EMDNCode,
		EMDNTerm:              r.EMDNTerm,
		Fields:                fields,
	}
}

func (r Result) ToInterchangeRecord() InterchangeRecord {
	fields := make(map[string]string)
	if r.Fields != nil {
		for k, v := range r.Fields {
			fields[k] = v
		}
	}
	if _, ok := fields[FieldDeviceType]; !ok && r.DeviceType != "" {
		fields[FieldDeviceType] = r.DeviceType
	}
	if _, ok := fields[FieldDeviceCategory]; !ok && r.DeviceCategory != "" {
		fields[FieldDeviceCategory] = r.DeviceCategory
	}
	if _, ok := fields[FieldDeviceFunction]; !ok && r.DeviceFunction != "" {
		fields[FieldDeviceFunction] = r.DeviceFunction
	}
	if _, ok := fields[FieldDeviceApplicationRisk]; !ok && r.DeviceApplicationRisk != "" {
		fields[FieldDeviceApplicationRisk] = r.DeviceApplicationRisk
	}
	if len(fields) == 0 {
		fields = nil
	}
	return InterchangeRecord{
		Name:                  r.Name,
		DeviceType:            r.GetField(FieldDeviceType),
		DeviceCategory:        r.GetField(FieldDeviceCategory),
		DeviceFunction:        r.GetField(FieldDeviceFunction),
		DeviceApplicationRisk: r.GetField(FieldDeviceApplicationRisk),
		LegacySourceName:      r.LegacySourceName,
		SourceType:            r.SourceType,
		EMDNCode:              r.EMDNCode,
		EMDNTerm:              r.EMDNTerm,
		MappingSource:         r.MappingSource,
		Confidence:            r.Confidence,
		Fields:                fields,
	}
}

func (r Result) IsValid() bool {
	dt := r.GetField(FieldDeviceType)
	return dt != "" && r.MappingSource != "unsupported_source_type"
}

// SearchAliasesString is kept for compat (now no-op).
func (r Result) SearchAliasesString() string { return "" }

// Deprecated accessors (shim): kept for callers that used struct fields as
// accessors conceptually. Prefer GetField(Field*).
// These are deprecated aliases; they proxy to Fields via GetField.

func (r Result) DeviceTypeAccessor() string            { return r.GetField(FieldDeviceType) }
func (r Result) DeviceFamilyAccessor() string          { return r.GetField(FieldDeviceFamily) }
func (r Result) DeviceFunctionAccessor() string        { return r.GetField(FieldDeviceFunction) }
func (r Result) DeviceApplicationRiskAccessor() string { return r.GetField(FieldDeviceApplicationRisk) }
func (r Result) DeviceCategoryAccessor() string        { return r.GetField(FieldDeviceCategory) }
