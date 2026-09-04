// Ovahol migration conventions.
//
// When a record being migrated onto Ovahol cannot be resolved to a known
// entity, the migration emits the canonical "Unknown" placeholder so lookups
// never fail and no free-text leaks into controlled/identity fields. These
// are Ovahol-specific conventions; they mirror what Ovahol's own lifecycle
// exports (device-records-lifecycle-*.json) already contain:
//
//	device        -> "Unknown Device"
//	brand         -> "Unknown Brand"
//	manufacturer  -> "Unknown Manufacturer"
//	model         -> "Unknown Model - <device name>"
//
// The model placeholder is a template because the model name must stay
// distinct per device so the reconciliation FK chain (model -brand>
// device) remains resolvable even for unknown models.
package main

import "strings"

// Canonical Ovahol "unknown" placeholders.
const (
	// UnknownDevice is used when an inbound device cannot be reconciled to
	// any entry in the device dictionary. It is itself a dictionary device
	// (device_type: Support Equipment & Furniture).
	UnknownDevice = "Unknown Device"

	// UnknownBrand is used when the inbound record carries no resolvable brand.
	UnknownBrand = "Unknown Brand"

	// UnknownManufacturer is used when the inbound record carries no
	// resolvable manufacturer.
	UnknownManufacturer = "Unknown Manufacturer"

	// UnknownModelPrefix prefixes the model placeholder. The full value is
	// UnknownModelPrefix + device name, keeping each unknown model unique per
	// device.
	UnknownModelPrefix = "Unknown Model - "
)

// UnknownModel returns the canonical placeholder model name for a device. It
// embeds the device name so that, even when the model is unknown, the
// model->device foreign key remains unique and resolvable.
func UnknownModel(deviceName string) string {
	if deviceName == "" {
		return UnknownModelPrefix + UnknownDevice
	}
	return UnknownModelPrefix + deviceName
}

// MigrationRecord is the identity/classification view produced when
// reconciling one inbound device onto Ovahol, with the Unknown conventions
// applied so every field is non-empty. It is the shape a migrator writes into
// Ovahol (device_name/model_name/brand_name/manufacturer_name mirror the
// lifecycle export).
type MigrationRecord struct {
	Name            string `json:"device_name"`
	Model           string `json:"model_name"`
	Brand           string `json:"brand_name"`
	Manufacturer    string `json:"manufacturer_name"`
	DeviceType      string `json:"device_type"`
	DeviceCategory  string `json:"device_category"`
	DeviceFunction  string `json:"device_function"`
	ApplicationRisk string `json:"application_risk"`
	EMDNCode        string `json:"emdn_code"`
	EMDNTerm        string `json:"emdn_term"`
	GMDNCode        string `json:"gmdn_code"`
	GMDNTerm        string `json:"gmdn_term"`
	MappingSource   string `json:"mapping_source"`
	Confidence      string `json:"confidence"`
}

// ApplyUnknownConventions fills the identity fields with the canonical
// Unknown placeholders whenever they are empty or unresolved, and returns the
// completed MigrationRecord. deviceName is the reconciled device name
// (result.Name); modelName/brandName/manufacturerName come from the source
// system's model FK lookup and may be empty/unknown.
func ApplyUnknownConventions(deviceName, modelName, brandName, manufacturerName, emdnCode, emdnTerm string) MigrationRecord {
	if deviceName == "" {
		deviceName = UnknownDevice
	}
	if strings.TrimSpace(modelName) == "" {
		modelName = UnknownModel(deviceName)
	}
	if strings.TrimSpace(brandName) == "" {
		brandName = UnknownBrand
	}
	if strings.TrimSpace(manufacturerName) == "" {
		manufacturerName = UnknownManufacturer
	}
	return MigrationRecord{
		Name:         deviceName,
		Model:        modelName,
		Brand:        brandName,
		Manufacturer: manufacturerName,
		EMDNCode:     emdnCode,
		EMDNTerm:     emdnTerm,
	}
}
