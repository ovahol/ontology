package ontology

// NamingRule defines a controlled-vocabulary naming convention.
type NamingRule struct {
	Field       string `json:"field"`
	Rule        string `json:"rule"`
	GoodExample string `json:"goodExample,omitempty"`
	Avoid       string `json:"avoid,omitempty"`
}
