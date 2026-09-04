package ontology

import (
	"regexp"
	"strings"
)

var nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)
var nonAlnumPreserveCaseRe = regexp.MustCompile(`[^a-zA-Z0-9]+`)
var parenRe = regexp.MustCompile(`\([^)]*\)`)
var slashRe = regexp.MustCompile(`/`)

// Normalized converts text to lowercase alphanumeric tokens separated by single spaces.
func Normalized(text string) string {
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	replaced := nonAlnumRe.ReplaceAllString(lower, " ")
	parts := strings.Fields(replaced)
	return strings.Join(parts, " ")
}

// HasAny reports whether text contains any of the candidate tokens.
func HasAny(text string, tokens []string) bool {
	padded := " " + text + " "
	for _, token := range tokens {
		candidate := Normalized(token)
		if candidate == "" {
			continue
		}
		if strings.Contains(candidate, " ") || len(candidate) <= 4 {
			if strings.Contains(padded, " "+candidate+" ") {
				return true
			}
			continue
		}
		if strings.Contains(text, candidate) {
			return true
		}
	}
	return false
}

// InferFromKeywords infers the device type from keywords under default taxonomy.
func InferFromKeywords(text string) string {
	return InferFromKeywordsFor(text, nil)
}

// InferFromKeywordsFor infers the device type from text using the provided taxonomy.
func InferFromKeywordsFor(text string, tax *Taxonomy) string {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	if tax.Inference == nil || len(tax.Inference.TypeByKeyword) == 0 {
		return ""
	}
	for _, entry := range tax.Inference.TypeByKeyword {
		if HasAny(text, entry.Keywords) {
			if tax.Inference.TypeByCode != nil {
				if v, ok := tax.Inference.TypeByCode[entry.Type]; ok {
					return v
				}
			}
			if entry.Type != "" {
				if v, ok := TypeByCode[entry.Type]; ok {
					return v
				}
				return entry.Type
			}
		}
	}
	return ""
}

// IsSupportedSourceType checks if sourceType is recognized under default taxonomy.
func IsSupportedSourceType(sourceType string) bool {
	return IsSupportedSourceTypeFor(sourceType, nil)
}

// IsSupportedSourceTypeFor checks if sourceType is recognized under the given taxonomy.
func IsSupportedSourceTypeFor(sourceType string, tax *Taxonomy) bool {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	norm := Normalized(sourceType)
	if tax.Inference != nil && len(tax.Inference.SupportedSourceTypes) > 0 {
		_, ok := tax.Inference.SupportedSourceTypes[norm]
		return ok
	}
	if tax.Inference != nil && len(tax.Inference.SourceTypeMap) > 0 {
		_, ok := tax.Inference.SourceTypeMap[norm]
		return ok
	}
	for _, dt := range tax.DeviceTypes {
		if Normalized(dt.Name) == norm || Normalized(dt.Code) == norm {
			return true
		}
	}
	return false
}

// InferDeviceType infers the high-level classification under default taxonomy.
func InferDeviceType(deviceName, sourceType, emdnTerm string) string {
	return InferDeviceTypeFor(deviceName, sourceType, emdnTerm, nil)
}

// InferDeviceTypeFor infers the high-level classification under the given taxonomy.
func InferDeviceTypeFor(deviceName, sourceType, emdnTerm string, tax *Taxonomy) string {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	parts := []string{}
	if deviceName != "" {
		parts = append(parts, deviceName)
	}
	if sourceType != "" {
		parts = append(parts, sourceType)
	}
	if emdnTerm != "" {
		parts = append(parts, emdnTerm)
	}
	text := Normalized(strings.Join(parts, " "))
	source := Normalized(sourceType)

	// 1. Keyword-based matching
	keywordMatch := InferFromKeywordsFor(text, tax)
	if keywordMatch != "" {
		return keywordMatch
	}

	// 2. Direct source type map
	if tax.Inference != nil && len(tax.Inference.SourceTypeMap) > 0 {
		if v, ok := tax.Inference.SourceTypeMap[source]; ok {
			return v
		}
	}

	// 3. Direct match in taxonomy DeviceTypes
	for _, dt := range tax.DeviceTypes {
		if Normalized(dt.Name) == source || Normalized(dt.Code) == source {
			return dt.Name
		}
	}

	// 4. Match in taxonomy Fields allowed values
	for _, f := range tax.Fields {
		if f.Key == FieldDeviceType || f.Required {
			for _, val := range f.AllowedValues {
				if Normalized(val) == source {
					return val
				}
			}
		}
	}

	return ""
}

// InferFamilyRule matches a FamilyRule under default taxonomy.
func InferFamilyRule(deviceName, sourceType, emdnTerm, deviceType string) *FamilyRule {
	return InferFamilyRuleFor(deviceName, sourceType, emdnTerm, deviceType, nil)
}

// InferFamilyRuleFor matches a FamilyRule under the given taxonomy.
func InferFamilyRuleFor(deviceName, sourceType, emdnTerm, deviceType string, tax *Taxonomy) *FamilyRule {
	if deviceType == "" {
		return nil
	}
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	if tax.Inference == nil || len(tax.Inference.FamilyRules) == 0 {
		return nil
	}
	parts := []string{}
	if deviceName != "" {
		parts = append(parts, deviceName)
	}
	if emdnTerm != "" {
		parts = append(parts, emdnTerm)
	}
	text := Normalized(strings.Join(parts, " "))
	source := Normalized(sourceType)
	for i := range tax.Inference.FamilyRules {
		r := &tax.Inference.FamilyRules[i]
		if r.Type != deviceType {
			continue
		}
		if len(r.SourceTypes) > 0 && source != "" {
			if HasAny(source, r.SourceTypes) {
				return r
			}
		}
		if len(r.Keywords) > 0 && text != "" {
			if HasAny(text, r.Keywords) {
				return r
			}
		}
	}
	return nil
}

// InferSpecificNameRule matches a SpecificNameRule under default taxonomy.
func InferSpecificNameRule(deviceName, sourceType, emdnTerm string) *SpecificNameRule {
	return InferSpecificNameRuleFor(deviceName, sourceType, emdnTerm, nil)
}

// InferSpecificNameRuleFor matches a SpecificNameRule under the given taxonomy.
func InferSpecificNameRuleFor(deviceName, sourceType, emdnTerm string, tax *Taxonomy) *SpecificNameRule {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	if tax.Inference == nil || len(tax.Inference.SpecificNameRules) == 0 {
		return nil
	}
	parts := []string{}
	if deviceName != "" {
		parts = append(parts, deviceName)
	}
	if sourceType != "" {
		parts = append(parts, sourceType)
	}
	if emdnTerm != "" {
		parts = append(parts, emdnTerm)
	}
	text := Normalized(strings.Join(parts, " "))
	for i := range tax.Inference.SpecificNameRules {
		r := &tax.Inference.SpecificNameRules[i]
		if len(r.ExcludeKeywords) > 0 && HasAny(text, r.ExcludeKeywords) {
			continue
		}
		if HasAny(text, r.Keywords) {
			return r
		}
	}
	return nil
}

// InferredDefaults holds default metadata resolved from inference rules.
type InferredDefaults struct {
	Family            string
	Function          string
	Risk              string
	CommonNameHint    string
	CanonicalNameHint string
}

// InferDefaults infers default dimensions under default taxonomy.
func InferDefaults(deviceName, sourceType, emdnTerm, deviceType string) InferredDefaults {
	return InferDefaultsFor(deviceName, sourceType, emdnTerm, deviceType, nil)
}

// InferDefaultsFor infers default dimensions under the given taxonomy.
func InferDefaultsFor(deviceName, sourceType, emdnTerm, deviceType string, tax *Taxonomy) InferredDefaults {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	var defaults InferredDefaults
	if tax.Inference != nil && tax.Inference.TypeDefaults != nil {
		if d, ok := tax.Inference.TypeDefaults[deviceType]; ok {
			defaults = InferredDefaults{
				Function: d.Function,
				Risk:     d.Risk,
			}
		}
	}
	if rule := InferFamilyRuleFor(deviceName, sourceType, emdnTerm, deviceType, tax); rule != nil {
		if rule.Family != "" {
			defaults.Family = rule.Family
		}
		if rule.Function != "" {
			defaults.Function = rule.Function
		}
		if rule.Risk != "" {
			defaults.Risk = rule.Risk
		}
		if rule.CommonName != "" {
			defaults.CommonNameHint = rule.CommonName
		}
		if rule.CanonicalName != "" {
			defaults.CanonicalNameHint = rule.CanonicalName
		}
	}
	return defaults
}

// CategoryForFunction returns category under default taxonomy.
func CategoryForFunction(functionName string) string {
	return CategoryForFunctionFor(functionName, nil)
}

// CategoryForFunctionFor returns category under the given taxonomy.
func CategoryForFunctionFor(functionName string, tax *Taxonomy) string {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	for _, fn := range tax.DeviceFunctions {
		if fn.Name == functionName {
			return fn.Category
		}
	}
	return ""
}

// CleanLegacySegment removes descriptor phrases and unit artifacts under default taxonomy.
func CleanLegacySegment(text string) string {
	return CleanLegacySegmentFor(text, nil)
}

// CleanLegacySegmentFor removes descriptor phrases and unit artifacts under the given taxonomy.
func CleanLegacySegmentFor(text string, tax *Taxonomy) string {
	if text == "" {
		return ""
	}
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	var phrases []string
	if tax.Inference != nil && len(tax.Inference.LegacyDescriptorPhrases) > 0 {
		phrases = tax.Inference.LegacyDescriptorPhrases
	} else {
		for p := range LegacyDescriptorPhrases {
			phrases = append(phrases, p)
		}
	}
	cleaned := text
	for _, phrase := range phrases {
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(phrase) + `\b`)
		cleaned = re.ReplaceAllString(cleaned, " ")
	}
	cleaned = regexp.MustCompile(`\b\d+(?:[.,]\d+)?\s*(?:mm|cm|m|ml|l|g|kg|v|w|hz|fr|gauge|ch|french|f|kva|kva|mhz|khz)\b`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`\b\d+\s*x\s*\d+\b`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`\b(?:set|pack|box|pair|kit)\s+of\s+\d+\b`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`\b\d+\s*(?:pieces?|pcs?|units?|tests?)\b`).ReplaceAllString(cleaned, " ")
	cleaned = nonAlnumPreserveCaseRe.ReplaceAllString(cleaned, " ")
	return strings.Join(strings.Fields(cleaned), " ")
}

// HumanizeName title-cases text and preserves acronyms under default taxonomy.
func HumanizeName(text string) string {
	return HumanizeNameFor(text, nil)
}

// HumanizeNameFor title-cases text and preserves acronyms under the given taxonomy.
func HumanizeNameFor(text string, tax *Taxonomy) string {
	if text == "" {
		return ""
	}
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	t := text
	if tax.Inference != nil && tax.Inference.WordReplacements != nil {
		for k, v := range tax.Inference.WordReplacements {
			re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(k) + `\b`)
			t = re.ReplaceAllString(t, v)
		}
	}
	words := strings.Fields(t)
	if len(words) == 0 {
		return ""
	}
	acronymMap := make(map[string]bool)
	if tax.Inference != nil && len(tax.Inference.Acronyms) > 0 {
		for _, a := range tax.Inference.Acronyms {
			acronymMap[strings.ToLower(a)] = true
		}
	}
	loweredWords := []string{"and", "or", "for", "with", "of", "in", "on", "at", "to", "by", "the", "a", "an"}
	isLowered := func(w string) bool {
		for _, lw := range loweredWords {
			if w == lw {
				return true
			}
		}
		return false
	}
	for i, w := range words {
		lw := strings.ToLower(w)
		if acronymMap[lw] {
			words[i] = strings.ToUpper(lw)
		} else if i > 0 && isLowered(lw) {
			words[i] = lw
		} else {
			words[i] = strings.ToUpper(lw[:1]) + lw[1:]
		}
	}
	return strings.Join(words, " ")
}

// InferCommonNameFromLegacy extracts a common name from legacy text under default taxonomy.
func InferCommonNameFromLegacy(legacySourceName string) string {
	return InferCommonNameFromLegacyFor(legacySourceName, nil)
}

// InferCommonNameFromLegacyFor extracts a common name from legacy text under the given taxonomy.
func InferCommonNameFromLegacyFor(legacySourceName string, tax *Taxonomy) string {
	if legacySourceName == "" {
		return ""
	}
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	raw := parenRe.ReplaceAllString(legacySourceName, " ")
	raw = slashRe.ReplaceAllString(raw, " ")
	segments := strings.Split(raw, ",")
	first := CleanLegacySegmentFor(segments[0], tax)
	if first == "" {
		for _, seg := range segments[1:] {
			c := CleanLegacySegmentFor(seg, tax)
			if c != "" {
				first = c
				break
			}
		}
	}
	if first == "" {
		return ""
	}
	modifiers := []string{}
	for _, seg := range segments[1:] {
		c := CleanLegacySegmentFor(seg, tax)
		if c != "" && len(strings.Fields(c)) <= 2 {
			modifiers = append(modifiers, c)
		}
	}
	firstWords := strings.Fields(first)
	head := strings.ToLower(firstWords[len(firstWords)-1])

	heads := GenericLegacyHeads
	if tax.Inference != nil && len(tax.Inference.GenericLegacyHeads) > 0 {
		heads = make(map[string]struct{}, len(tax.Inference.GenericLegacyHeads))
		for _, h := range tax.Inference.GenericLegacyHeads {
			heads[Normalized(h)] = struct{}{}
		}
	}
	var candidate string
	if _, ok := heads[head]; ok && len(modifiers) > 0 {
		candidate = strings.Join(append(modifiers, first), " ")
	} else {
		candidate = first
	}
	candidate = HumanizeNameFor(candidate, tax)
	wc := len(strings.Fields(candidate))
	if wc >= 1 && wc <= 6 {
		return candidate
	}
	return ""
}

// RefineDescriptiveNames refines common and canonical names under default taxonomy.
func RefineDescriptiveNames(commonName, canonicalName, legacySourceName, emdnTerm string) (string, string) {
	return RefineDescriptiveNamesFor(commonName, canonicalName, legacySourceName, emdnTerm, nil)
}

// RefineDescriptiveNamesFor refines common and canonical names under the given taxonomy.
func RefineDescriptiveNamesFor(commonName, canonicalName, legacySourceName, emdnTerm string, tax *Taxonomy) (string, string) {
	if commonName == "" {
		return commonName, canonicalName
	}
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	if tax.Inference == nil || len(tax.Inference.NameRefinementRules) == 0 {
		return commonName, canonicalName
	}
	parts := []string{}
	if legacySourceName != "" {
		parts = append(parts, legacySourceName)
	}
	if emdnTerm != "" {
		parts = append(parts, emdnTerm)
	}
	text := Normalized(strings.Join(parts, " "))

	for _, r := range tax.Inference.NameRefinementRules {
		if strings.EqualFold(r.TargetName, commonName) {
			if len(r.Keywords) == 0 || HasAny(text, r.Keywords) {
				cName := r.CommonName
				canName := r.CanonicalName
				if canName == "" {
					canName = cName
				}
				return cName, canName
			}
		}
	}
	return commonName, canonicalName
}

// ResolvedRow holds intermediate resolution data for a row.
type ResolvedRow struct {
	Fields map[string]string

	// Deprecated: use Fields[FieldDeviceType] or GetField.
	DeviceType string
	// Deprecated: use Fields[FieldDeviceCategory].
	DeviceCategory string
	// Deprecated: use Fields[FieldDeviceFamily].
	DeviceFamily string
	// Deprecated: use Fields[FieldDeviceFunction].
	DeviceFunction string
	// Deprecated: use Fields[FieldDeviceApplicationRisk].
	DeviceApplicationRisk string

	Name          string
	CanonicalName string
	CommonNames   []string
	NamingSource  string
}

// GetField returns the taxonomy value for key, checking Fields first then deprecated fixed field.
func (r ResolvedRow) GetField(key string) string {
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
	case FieldDeviceFamily:
		return r.DeviceFamily
	case FieldDeviceFunction:
		return r.DeviceFunction
	case FieldDeviceApplicationRisk:
		return r.DeviceApplicationRisk
	}
	return ""
}

// Deprecated accessors (shim) for ResolvedRow.
func (r ResolvedRow) DeviceTypeAccessor() string     { return r.GetField(FieldDeviceType) }
func (r ResolvedRow) DeviceFamilyAccessor() string   { return r.GetField(FieldDeviceFamily) }
func (r ResolvedRow) DeviceFunctionAccessor() string { return r.GetField(FieldDeviceFunction) }
func (r ResolvedRow) DeviceApplicationRiskAccessor() string {
	return r.GetField(FieldDeviceApplicationRisk)
}
func (r ResolvedRow) DeviceCategoryAccessor() string { return r.GetField(FieldDeviceCategory) }

// ResolveRowNaming resolves names and taxonomy dimensions under default taxonomy.
func ResolveRowNaming(row map[string]string) ResolvedRow {
	return ResolveRowNamingFor(row, nil)
}

// ResolveRowNamingFor resolves names and taxonomy dimensions under the given taxonomy.
func ResolveRowNamingFor(row map[string]string, tax *Taxonomy) ResolvedRow {
	legacySourceName := row["Legacy source name"]
	sourceDeviceType := row["Source device type"]
	emdnTerm := row["EMDN term"]

	specificRule := InferSpecificNameRuleFor(legacySourceName, sourceDeviceType, emdnTerm, tax)
	inferredType := ""
	if specificRule != nil && specificRule.Type != "" {
		inferredType = specificRule.Type
	} else {
		inferredType = InferDeviceTypeFor(legacySourceName, sourceDeviceType, emdnTerm, tax)
	}
	if inferredType == "" {
		return ResolvedRow{NamingSource: "unsupported_source_type"}
	}

	defaults := InferDefaultsFor(legacySourceName, sourceDeviceType, emdnTerm, inferredType, tax)
	generatedCommon := InferCommonNameFromLegacyFor(legacySourceName, tax)
	defaultCommon := defaults.CommonNameHint
	defaultCanonical := defaults.CanonicalNameHint

	namingSource := "family_fallback"
	commonName := defaultCommon
	if specificRule != nil && specificRule.CommonName != "" {
		commonName = specificRule.CommonName
		namingSource = "specific_rule"
	} else if generatedCommon != "" {
		commonName = generatedCommon
		namingSource = "legacy_derived"
	}

	canonicalName := defaultCanonical
	if canonicalName == "" {
		canonicalName = commonName
	}
	if specificRule != nil && specificRule.CanonicalName != "" {
		canonicalName = specificRule.CanonicalName
	} else if namingSource == "specific_rule" || namingSource == "legacy_derived" {
		canonicalName = commonName
	}

	commonName, canonicalName = RefineDescriptiveNamesFor(commonName, canonicalName, legacySourceName, emdnTerm, tax)

	family := ""
	if specificRule != nil && specificRule.Family != "" {
		family = specificRule.Family
	} else {
		family = defaults.Family
	}

	category := CategoryForFunctionFor(defaults.Function, tax)
	fields := map[string]string{
		FieldDeviceType:            inferredType,
		FieldDeviceCategory:        category,
		FieldDeviceFamily:          family,
		FieldDeviceFunction:        defaults.Function,
		FieldDeviceApplicationRisk: defaults.Risk,
	}

	return ResolvedRow{
		Fields:                fields,
		DeviceType:            inferredType,
		DeviceCategory:        category,
		DeviceFamily:          family,
		DeviceFunction:        defaults.Function,
		DeviceApplicationRisk: defaults.Risk,
		Name:                  commonName,
		CanonicalName:         canonicalName,
		CommonNames:           BuildSearchAliasesFor(commonName, canonicalName, tax),
		NamingSource:          namingSource,
	}
}

// BuildSearchAliases builds search aliases under default taxonomy.
func BuildSearchAliases(commonName, canonicalName string) []string {
	return BuildSearchAliasesFor(commonName, canonicalName, nil)
}

// BuildSearchAliasesFor builds search aliases under the given taxonomy.
func BuildSearchAliasesFor(commonName, canonicalName string, tax *Taxonomy) []string {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	values := []string{}
	add := func(v string) {
		if v == "" {
			return
		}
		cleaned := strings.Join(strings.Fields(strings.TrimSpace(v)), " ")
		if cleaned == "" {
			return
		}
		lowered := strings.ToLower(cleaned)
		for _, existing := range values {
			if strings.ToLower(existing) == lowered {
				return
			}
		}
		values = append(values, cleaned)
	}
	add(commonName)
	add(canonicalName)

	if tax.Inference != nil && len(tax.Inference.SearchAliasRules) > 0 {
		combined := Normalized(strings.Join([]string{commonName, canonicalName}, " "))
		for _, r := range tax.Inference.SearchAliasRules {
			if HasAny(combined, r.Keywords) {
				for _, a := range r.Aliases {
					add(a)
				}
			}
		}
	}
	if len(values) == 0 {
		return nil
	}
	return values
}
