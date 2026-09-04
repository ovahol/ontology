package ontology

// NamingRule defines a controlled-vocabulary naming convention.
type NamingRule struct {
	Field       string `json:"field"`
	Rule        string `json:"rule"`
	GoodExample string `json:"goodExample,omitempty"`
	Avoid       string `json:"avoid,omitempty"`
}

// DefaultDeviceSheetHeaders defines the column headers for the Devices sheet
// in workbook interchange. These are structural/engine-level, not vendor vocabulary.
var DefaultDeviceSheetHeaders = []string{
	"Name",
	"Device type",
	"Device category",
	"Device function",
	"Device application risk",
	"Legacy source name",
	"Source device type",
	"EMDN code",
	"EMDN term",
}

// DefaultAPIImportHeaders defines the column headers for the API import CSV.
// These are structural/engine-level, not vendor vocabulary.
var DefaultAPIImportHeaders = []string{
	"name",
	"device_type",
	"device_category",
	"device_function",
	"device_application_risk",
	"emdn_code",
	"emdn_term",
}
