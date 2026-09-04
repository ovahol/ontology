package ontology

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CurrentSchemaVersion is the current ontology schema version. Vendors
// define their own Taxonomy.ID/Version in their own codebase; this is only
// used for semver compatibility checks when a taxonomy is loaded.
const CurrentSchemaVersion = "1.0.0"

// CurrentTaxonomyID is kept for migration helpers but no longer embedded.
const CurrentTaxonomyID = "ovahol-ontology"

// CurrentTaxonomyVersion is an alias for CurrentSchemaVersion.
const CurrentTaxonomyVersion = CurrentSchemaVersion

// TaxonomyVersion is kept for backward-compat.
const TaxonomyVersion = CurrentSchemaVersion

// FieldDef defines a custom taxonomy classification field / dimension.
type FieldDef struct {
	Key           string   `json:"key"`
	Label         string   `json:"label,omitempty"`
	Required      bool     `json:"required,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
}

// NameRefinementRule defines a keyword-based common/canonical name transformation rule.
type NameRefinementRule struct {
	TargetName    string   `json:"targetName"`
	Keywords      []string `json:"keywords,omitempty"`
	CommonName    string   `json:"commonName"`
	CanonicalName string   `json:"canonicalName,omitempty"`
}

// SearchAliasRule defines keyword-triggered search aliases.
type SearchAliasRule struct {
	Keywords []string `json:"keywords"`
	Aliases  []string `json:"aliases"`
}

// InferenceRules contains all rule-based matching, grouping, and inference configuration.
type InferenceRules struct {
	TypeByKeyword []struct {
		Keywords []string `json:"keywords"`
		Type     string   `json:"type"`
	} `json:"typeByKeyword,omitempty"`
	SourceTypeMap        map[string]string   `json:"sourceTypeMap,omitempty"`
	SupportedSourceTypes map[string]struct{} `json:"supportedSourceTypes,omitempty"`
	FamilyRules          []FamilyRule        `json:"familyRules,omitempty"`
	SpecificNameRules    []SpecificNameRule  `json:"specificNameRules,omitempty"`
	TypeDefaults         map[string]struct {
		Function string `json:"function"`
		Risk     string `json:"risk"`
	} `json:"typeDefaults,omitempty"`
	TypeByCode              map[string]string    `json:"typeByCode,omitempty"`
	FunctionByCode          map[string]string    `json:"functionByCode,omitempty"`
	GenericLegacyHeads      []string             `json:"genericLegacyHeads,omitempty"`
	LegacyDescriptorPhrases []string             `json:"legacyDescriptorPhrases,omitempty"`
	NameRefinementRules     []NameRefinementRule `json:"nameRefinementRules,omitempty"`
	Acronyms                []string             `json:"acronyms,omitempty"`
	WordReplacements        map[string]string    `json:"wordReplacements,omitempty"`
	SearchAliasRules        []SearchAliasRule    `json:"searchAliasRules,omitempty"`
}

// Taxonomy is the complete specification of a vendor's controlled vocabulary,
// classification dimensions, and inference rules.
type Taxonomy struct {
	ID                     string                  `json:"id"`
	Name                   string                  `json:"name,omitempty"`
	Version                string                  `json:"version,omitempty"`
	SchemaVersion          string                  `json:"schemaVersion,omitempty"`
	DeviceTypes            []DeviceType            `json:"deviceTypes,omitempty"`
	DeviceCategories       []DeviceCategory        `json:"deviceCategories,omitempty"`
	DeviceFunctions        []DeviceFunction        `json:"deviceFunctions,omitempty"`
	DeviceApplicationRisks []DeviceApplicationRisk `json:"deviceApplicationRisks,omitempty"`
	NamingRules            []NamingRule            `json:"namingRules,omitempty"`
	Fields                 []FieldDef              `json:"fields,omitempty"`
	Inference              *InferenceRules         `json:"inference,omitempty"`
	Source                 string                  `json:"source,omitempty"`
	Counts                 map[string]int          `json:"counts,omitempty"`
}

// CurrentTaxonomy is deprecated: ontology no longer ships a hardcoded runtime vocab.
// Vendors provide their own Taxonomy via LoadTaxonomy / LoadTaxonomyFile.
// Kept as stub to ease migration — returns empty taxonomy with current ID/version.
func CurrentTaxonomy() *Taxonomy {
	return &Taxonomy{
		ID:            CurrentTaxonomyID,
		Version:       CurrentSchemaVersion,
		SchemaVersion: CurrentSchemaVersion,
	}
}

// parseSemver parses a strict semver "MAJOR.MINOR.PATCH" (no leading "v").
func parseSemver(v string) (major, minor, patch int, err error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, 0, 0, fmt.Errorf("empty version")
	}
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver %q: want MAJOR.MINOR.PATCH", v)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid semver %q: %w", v, err)
	}
	if major < 0 || minor < 0 || patch < 0 {
		return 0, 0, 0, fmt.Errorf("invalid semver %q: negative component", v)
	}
	return major, minor, patch, nil
}

// compareSemver returns -1 if a<b, 0 if a==b, 1 if a>b.
func compareSemver(a, b string) (int, error) {
	ma, mi, pa, err := parseSemver(a)
	if err != nil {
		return 0, err
	}
	mb, mib, pb, err := parseSemver(b)
	if err != nil {
		return 0, err
	}
	if ma != mb {
		if ma < mb {
			return -1, nil
		}
		return 1, nil
	}
	if mi != mib {
		if mi < mib {
			return -1, nil
		}
		return 1, nil
	}
	if pa != pb {
		if pa < pb {
			return -1, nil
		}
		return 1, nil
	}
	return 0, nil
}

// isBreakingMismatch reports whether loaded vs current is a breaking change
// (different major version).
func isBreakingMismatch(loaded, current string) (bool, error) {
	ma, _, _, err := parseSemver(loaded)
	if err != nil {
		return false, err
	}
	mb, _, _, err := parseSemver(current)
	if err != nil {
		return false, err
	}
	return ma != mb, nil
}

// Validate checks id, version, and taxonomy payload for well-formedness.
func (t *Taxonomy) Validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return fmt.Errorf("taxonomy: missing id")
	}
	ver := t.effectiveVersion()
	if strings.TrimSpace(ver) == "" {
		return fmt.Errorf("taxonomy: missing version/schemaVersion")
	}
	if _, _, _, err := parseSemver(ver); err != nil {
		return fmt.Errorf("taxonomy: %w", err)
	}
	if len(t.Fields) > 0 {
		hasRequired := false
		for _, f := range t.Fields {
			if strings.TrimSpace(f.Key) == "" {
				return fmt.Errorf("taxonomy: field missing key")
			}
			if f.Required {
				hasRequired = true
				if len(f.AllowedValues) == 0 {
					return fmt.Errorf("taxonomy: required field %q missing allowed_values", f.Key)
				}
			}
		}
		if !hasRequired && len(t.Fields) > 0 {
			return fmt.Errorf("taxonomy: at least one field must be required")
		}
		return nil
	}
	// Ovahol-shaped or minimal taxonomy
	if len(t.DeviceTypes) == 0 && len(t.DeviceCategories) == 0 && len(t.DeviceFunctions) == 0 && len(t.DeviceApplicationRisks) == 0 && t.Inference == nil {
		return fmt.Errorf("taxonomy: at least one field, device type, category, or inference rule must be defined")
	}
	return nil
}

func (t *Taxonomy) effectiveVersion() string {
	if strings.TrimSpace(t.SchemaVersion) != "" {
		return strings.TrimSpace(t.SchemaVersion)
	}
	return strings.TrimSpace(t.Version)
}

func (t *Taxonomy) normalizeVersionAlias() {
	v := strings.TrimSpace(t.Version)
	sv := strings.TrimSpace(t.SchemaVersion)
	if v == "" && sv != "" {
		t.Version = sv
	}
	if sv == "" && v != "" {
		t.SchemaVersion = v
	}
	if v != "" && sv != "" && v != sv {
		// Prefer schemaVersion as canonical when both present but differ.
		t.Version = sv
	}
}

// Migrate upgrades a taxonomy decoded from an older minor/patch to the current
// schema. It returns an error on breaking (major) mismatch. On success the
// returned taxonomy has its version fields set to CurrentSchemaVersion.
func Migrate(t *Taxonomy) (*Taxonomy, error) {
	if t == nil {
		return nil, fmt.Errorf("taxonomy: nil")
	}
	t.normalizeVersionAlias()
	ver := t.effectiveVersion()
	if ver == "" {
		return nil, fmt.Errorf("taxonomy: missing version/schemaVersion")
	}
	if _, _, _, err := parseSemver(ver); err != nil {
		return nil, fmt.Errorf("taxonomy: %w", err)
	}
	breaking, err := isBreakingMismatch(ver, CurrentSchemaVersion)
	if err != nil {
		return nil, err
	}
	if breaking {
		return nil, fmt.Errorf("taxonomy: breaking version mismatch: got %q, want %q (major version differs)", ver, CurrentSchemaVersion)
	}
	cmp, err := compareSemver(ver, CurrentSchemaVersion)
	if err != nil {
		return nil, err
	}
	if cmp == 0 {
		t.Version = CurrentSchemaVersion
		t.SchemaVersion = CurrentSchemaVersion
		return t, nil
	}
	// Minor/patch migration: accept older or newer patch within same major.
	t.Version = CurrentSchemaVersion
	t.SchemaVersion = CurrentSchemaVersion
	return t, nil
}

// LoadTaxonomy decodes, validates, and migrates a taxonomy JSON document.
func LoadTaxonomy(data []byte) (*Taxonomy, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("taxonomy: empty document")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var t Taxonomy
	if err := dec.Decode(&t); err != nil {
		return nil, fmt.Errorf("taxonomy: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("taxonomy: trailing content after JSON object")
	}
	t.normalizeVersionAlias()
	if strings.TrimSpace(t.ID) == "" {
		return nil, fmt.Errorf("taxonomy: missing id")
	}
	if strings.TrimSpace(t.effectiveVersion()) == "" {
		return nil, fmt.Errorf("taxonomy: missing version/schemaVersion")
	}
	ver := t.effectiveVersion()
	if _, _, _, err := parseSemver(ver); err != nil {
		return nil, fmt.Errorf("taxonomy: %w", err)
	}
	if breaking, err := isBreakingMismatch(ver, CurrentSchemaVersion); err != nil {
		return nil, err
	} else if breaking {
		return nil, fmt.Errorf("taxonomy: breaking version mismatch: got %q, want %q (major version differs)", ver, CurrentSchemaVersion)
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return Migrate(&t)
}

// LoadTaxonomyFile reads a file and delegates to LoadTaxonomy.
func LoadTaxonomyFile(path string) (*Taxonomy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("taxonomy: read %s: %w", path, err)
	}
	return LoadTaxonomy(data)
}

// LoadEmbeddedTaxonomy is deprecated: no embedded vocab any more.
// Vendors must load from their own file via LoadTaxonomyFile.
func LoadEmbeddedTaxonomy() (*Taxonomy, error) {
	return nil, fmt.Errorf("taxonomy: no embedded taxonomy — provide your own taxonomy file via LoadTaxonomyFile")
}

// MustLoadEmbeddedTaxonomy is deprecated: see LoadEmbeddedTaxonomy.
func MustLoadEmbeddedTaxonomy() *Taxonomy {
	panic("taxonomy: no embedded taxonomy — provide your own taxonomy file")
}
