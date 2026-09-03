// Package ontology ports scripts/update_ovahol_ontology.py to Go.
// Every device_type, device_function, and device_application_risk is validated
// against the controlled vocabulary so the pipeline stays consistent.
package ontology

import (
	"regexp"
	"strings"
)

var nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)
var parenRe = regexp.MustCompile(`\([^)]*\)`)
var slashRe = regexp.MustCompile(`/`)

func Normalized(text string) string {
	if text == "" {
		return ""
	}
	lower := strings.ToLower(text)
	replaced := nonAlnumRe.ReplaceAllString(lower, " ")
	parts := strings.Fields(replaced)
	return strings.Join(parts, " ")
}

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

func InferFromKeywords(text string) string {
	if HasAny(text, []string{"software", "pacs", "electronic medical record", "emr", "ehr", "clinical information system"}) {
		return TypeByCode["DIGITAL_HEALTH_CLINICAL_SOFTWARE"]
	}
	if HasAny(text, []string{"calibration", "quality control", "testing system", "tester", "test equipment", "patient simulator", "simulator", "phantom", "electrical safety"}) {
		return TypeByCode["BIOMEDICAL_TEST_CALIBRATION_QUALITY_ASSURANCE"]
	}
	if HasAny(text, []string{"ultrasound", "x ray", "xray", "ct ", "computed tomography", "mri", "magnetic resonance", "fluoroscopy", "mammograph", "radiograph", "nuclear medicine", "gamma camera", "endoscope", "bronchoscope", "cystoscope", "colposcope", "laparoscope", "arthroscope", "otoscope", "ophthalmoscope"}) {
		return TypeByCode["DIAGNOSTIC_IMAGING_VISUALIZATION"]
	}
	if HasAny(text, []string{"ivd", "reagent", "assay", "hematology", "chemistry analyzer", "chemistry analyser", "centrifuge", "microscope", "pcr", "blood gas", "laboratory", "specimen", "sample"}) {
		return TypeByCode["LABORATORY_IN_VITRO_DIAGNOSTICS"]
	}
	if HasAny(text, []string{"oxygen", "suction", "airway", "nebul", "flowmeter", "humidifier", "psa", "breathing circuit", "respiratory", "cpap", "bipap", "aspirator"}) {
		return TypeByCode["MEDICAL_GAS_RESPIRATORY_SUCTION"]
	}
	if HasAny(text, []string{"ventilator", "infusion", "syringe pump", "defibrillator", "dialysis", "anaesthesia", "anesthesia", "electrosurgical", "pacemaker", "cardioverter", "apheresis", "resuscitator", "life support"}) {
		return TypeByCode["THERAPEUTIC_LIFE_SUPPORT"]
	}
	if HasAny(text, []string{"monitor", "ecg", "eeg", "tonometer", "audiometer", "spirom", "pulmonary function", "bilirubinometer", "vital signs", "blood pressure", "stethoscope"}) {
		return TypeByCode["CLINICAL_MONITORING_ASSESSMENT"]
	}
	if HasAny(text, []string{"wheelchair", "walker", "crutch", "prosthe", "ortho", "mobility", "assistive", "rehabilitation", "physiotherapy"}) {
		return TypeByCode["REHABILITATION_MOBILITY_ASSISTIVE"]
	}
	if HasAny(text, []string{"steriliz", "autoclave", "disinfect", "decontamin", "ipc", "glove", "mask", "gown", "apron"}) {
		return TypeByCode["INFECTION_PREVENTION_DECONTAMINATION_STERILIZATION"]
	}
	if HasAny(text, []string{"surgical", "instrument", "tray", "trocar", "catheter", "cannula", "implant", "clip", "stent", "laparoscopy", "dental", "oral", "iud", "intra uterine", "intra uterine system", "anoscope", "scope sheath"}) {
		return TypeByCode["SURGICAL_INTERVENTIONAL"]
	}
	if HasAny(text, []string{"dressing", "gauze", "bandage", "pad", "syringe", "needle", "tube", "drainage", "bag", "collection", "accessory", "swab", "suture", "condom", "diaphragm", "cervical cap"}) {
		return TypeByCode["CONSUMABLES_ACCESSORIES_PROCEDURE_SUPPLIES"]
	}
	if HasAny(text, []string{"bed", "cabinet", "ambulance", "vehicle", "chart", "stationery", "waste", "clock", "printer", "data logger", "lamp", "radiation shielding", "shielding"}) {
		return TypeByCode["FACILITY_UTILITY_ENVIRONMENTAL_SUPPORT"]
	}
	return ""
}

func IsSupportedSourceType(sourceType string) bool {
	_, ok := SupportedSourceTypes[Normalized(sourceType)]
	return ok
}

func InferDeviceType(deviceName, sourceType, emdnTerm string) string {
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
	keywordMatch := InferFromKeywords(text)
	switch source {
	case "medical equipment":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["THERAPEUTIC_LIFE_SUPPORT"]
	case "accessories":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["CONSUMABLES_ACCESSORIES_PROCEDURE_SUPPLIES"]
	case "collection devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["CONSUMABLES_ACCESSORIES_PROCEDURE_SUPPLIES"]
	case "catheters and related":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["CONSUMABLES_ACCESSORIES_PROCEDURE_SUPPLIES"]
	case "contraception devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["CONSUMABLES_ACCESSORIES_PROCEDURE_SUPPLIES"]
	case "endoscopes and related devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["DIAGNOSTIC_IMAGING_VISUALIZATION"]
	case "implantable devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["SURGICAL_INTERVENTIONAL"]
	case "measurement devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["CLINICAL_MONITORING_ASSESSMENT"]
	case "non medical devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["FACILITY_UTILITY_ENVIRONMENTAL_SUPPORT"]
	case "oral devices":
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["SURGICAL_INTERVENTIONAL"]
	case "personal protective equipment radiation protection equipment":
		if HasAny(text, []string{"radiation", "shielding"}) {
			return TypeByCode["FACILITY_UTILITY_ENVIRONMENTAL_SUPPORT"]
		}
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["INFECTION_PREVENTION_DECONTAMINATION_STERILIZATION"]
	case "solutions and reagents":
		if HasAny(text, []string{"disinfect", "steriliz", "cleaning", "decontamin"}) {
			return TypeByCode["INFECTION_PREVENTION_DECONTAMINATION_STERILIZATION"]
		}
		if keywordMatch != "" {
			return keywordMatch
		}
		return TypeByCode["LABORATORY_IN_VITRO_DIAGNOSTICS"]
	}
	if keywordMatch != "" {
		return keywordMatch
	}
	if v, ok := DirectSourceTypeMap[source]; ok {
		return v
	}
	return ""
}

func InferFamilyRule(deviceName, sourceType, emdnTerm, deviceType string) *FamilyRule {
	if deviceType == "" {
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
	source := Normalized(sourceType)
	for i := range FamilyRules {
		r := &FamilyRules[i]
		if r.Type != deviceType {
			continue
		}
		if source != "" {
			for _, st := range r.SourceTypes {
				if Normalized(st) == source {
					return r
				}
			}
		}
		if HasAny(text, r.Keywords) {
			return r
		}
	}
	return nil
}

type Defaults struct {
	Family           string
	Function         string
	Risk             string
	CommonNameHint   string
	CanonicalNameHint string
}

func InferDefaults(deviceName, sourceType, emdnTerm, deviceType string) Defaults {
	if deviceType == "" {
		return Defaults{}
	}
	r := InferFamilyRule(deviceName, sourceType, emdnTerm, deviceType)
	td, ok := DeviceTypeDefaults[deviceType]
	if !ok {
		td = struct{ Function, Risk string }{}
	}
	if r != nil {
		return Defaults{Family: r.Family, Function: r.Function, Risk: r.Risk, CommonNameHint: r.CommonName, CanonicalNameHint: r.CanonicalName}
	}
	return Defaults{Function: td.Function, Risk: td.Risk}
}

func CategoryForFunction(name string) string {
	for _, f := range DeviceFunctions {
		if f.Name == name {
			return f.Category
		}
	}
	return ""
}

func CleanLegacySegment(value string) string {
	cleaned := parenRe.ReplaceAllString(value, "")
	cleaned = strings.ReplaceAll(cleaned, "/", " / ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.Trim(cleaned, " ,-/")
	lowered := strings.ToLower(cleaned)
	if _, ok := LegacyDescriptorPhrases[lowered]; ok {
		return ""
	}
	// also check exact phrase match (same as above, kept for parity)
	return cleaned
}

func HumanizeName(value string) string {
	phrase := strings.Join(strings.Fields(value), " ")
	phrase = strings.TrimSpace(phrase)
	if phrase == "" {
		return phrase
	}
	replacements := map[string]string{
		"anaesthesia": "anesthesia",
		"analyser":    "analyzer",
		"haemodialysis": "hemodialysis",
		"haemoglobin":   "hemoglobin",
		"foetal":        "fetal",
	}
	lowered := strings.ToLower(phrase)
	for old, nw := range replacements {
		lowered = strings.ReplaceAll(lowered, old, nw)
	}
	words := strings.Fields(lowered)
	acronyms := map[string]bool{"ecg": true, "eeg": true, "iv": true, "mri": true, "aed": true, "cpap": true, "bipap": true, "psa": true, "ent": true, "oct": true, "ivd": true, "it": true}
	out := []string{}
	for i, w := range words {
		if acronyms[w] {
			out = append(out, strings.ToUpper(w))
		} else if i == 0 {
			if len(w) > 0 {
				out = append(out, strings.ToUpper(w[:1])+w[1:])
			} else {
				out = append(out, w)
			}
		} else {
			out = append(out, w)
		}
	}
	return strings.Join(out, " ")
}

func InferSpecificNameRule(deviceName, sourceType, emdnTerm string) *SpecificNameRule {
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
	for i := range SpecificNameRules {
		r := &SpecificNameRules[i]
		if len(r.Keywords) > 0 && !HasAny(text, r.Keywords) {
			continue
		}
		if len(r.ExcludeKeywords) > 0 && HasAny(text, r.ExcludeKeywords) {
			continue
		}
		return r
	}
	return nil
}

func InferCommonNameFromLegacy(deviceName string) string {
	if deviceName == "" {
		return ""
	}
	rawParts := strings.Split(deviceName, ",")
	parts := []string{}
	for _, p := range rawParts {
		cleaned := CleanLegacySegment(p)
		if cleaned != "" {
			parts = append(parts, cleaned)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		candidate := HumanizeName(parts[0])
		wc := len(strings.Fields(candidate))
		if wc >= 1 && wc <= 5 {
			return candidate
		}
		return ""
	}
	first := parts[0]
	modifiers := parts[1:]
	firstTokens := strings.Fields(Normalized(first))
	head := ""
	if len(firstTokens) > 0 {
		head = firstTokens[len(firstTokens)-1]
	}
	var candidate string
	if _, ok := GenericLegacyHeads[head]; ok && len(modifiers) > 0 {
		candidate = strings.Join(append(modifiers, first), " ")
	} else {
		candidate = first
	}
	candidate = HumanizeName(candidate)
	wc := len(strings.Fields(candidate))
	if wc >= 1 && wc <= 6 {
		return candidate
	}
	return ""
}

func RefineDescriptiveNames(commonName, canonicalName, legacySourceName, emdnTerm string) (string, string) {
	if commonName == "" {
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
	apply := func(common string, canonical ...string) (string, string) {
		c := common
		c2 := common
		if len(canonical) > 0 && canonical[0] != "" {
			c2 = canonical[0]
		}
		return c, c2
	}
	switch commonName {
	case "Thermometer":
		if HasAny(text, []string{"non contact", "non-contact", "infrared"}) {
			return apply("Infrared thermometer")
		}
		if HasAny(text, []string{"tympanic", "ear"}) {
			return apply("Ear thermometer")
		}
		if HasAny(text, []string{"digital"}) {
			return apply("Digital thermometer")
		}
		if HasAny(text, []string{"laboratory", "glass"}) {
			return apply("Laboratory thermometer")
		}
		if HasAny(text, []string{"water bath"}) {
			return apply("Water bath thermometer")
		}
		if HasAny(text, []string{"calibrated"}) {
			return apply("Calibrated thermometer")
		}
		return apply("Clinical thermometer")
	case "Scale":
		if HasAny(text, []string{"stadiometer", "altimeter", "height"}) {
			return apply("Height and weight scale")
		}
		if HasAny(text, []string{"blood bag", "donor blood", "blood scales"}) {
			return apply("Blood bag scale")
		}
		if HasAny(text, []string{"neonatal", "infant"}) {
			return apply("Neonatal scale")
		}
		if HasAny(text, []string{"analytical"}) {
			return apply("Analytical balance")
		}
		if HasAny(text, []string{"technical"}) {
			return apply("Technical balance")
		}
		if HasAny(text, []string{"beamtype", "beam type", "beam"}) {
			return apply("Beam balance scale")
		}
		if HasAny(text, []string{"digital"}) {
			return apply("Digital weighing scale")
		}
		return apply("Weighing scale")
	case "Pipette":
		if HasAny(text, []string{"stand"}) {
			return apply("Pipette stand")
		}
		if HasAny(text, []string{"filler", "wheel run"}) {
			return apply("Pipette filler")
		}
		if HasAny(text, []string{"micropipette"}) {
			return apply("Micropipette")
		}
		if HasAny(text, []string{"serological", "dilution"}) {
			return apply("Serological pipette")
		}
		if HasAny(text, []string{"blood"}) {
			return apply("Blood pipette")
		}
		if HasAny(text, []string{"esr"}) {
			return apply("ESR pipette")
		}
		if HasAny(text, []string{"pasteur"}) {
			return apply("Pasteur pipette")
		}
		if HasAny(text, []string{"digital"}) {
			return apply("Digital pipette")
		}
		if HasAny(text, []string{"automatic"}) {
			return apply("Automatic pipette")
		}
		return apply("Laboratory pipette")
	case "Tube":
		if HasAny(text, []string{"capillary", "blood collection"}) {
			return apply("Capillary blood tube")
		}
		if HasAny(text, []string{"oxygen administration", "medical gases"}) {
			return apply("Oxygen tubing")
		}
		if HasAny(text, []string{"pcr"}) {
			return apply("PCR tube")
		}
		if HasAny(text, []string{"microtube", "microvial", "micro tube"}) {
			return apply("Lab microtube")
		}
		if HasAny(text, []string{"centrifuge"}) {
			return apply("Centrifuge tube")
		}
		if HasAny(text, []string{"sample collection", "sample analysis", "sample analyses"}) {
			return apply("Sample tube")
		}
		if HasAny(text, []string{"test tube"}) {
			return apply("Test tube")
		}
		return apply("Laboratory tube")
	case "Table":
		if HasAny(text, []string{"operating tables", "operating table"}) {
			return apply("Operating table")
		}
		if HasAny(text, []string{"grossing"}) {
			return apply("Grossing table")
		}
		if HasAny(text, []string{"neonatal resuscitation"}) {
			return apply("Neonatal resuscitation table")
		}
		if HasAny(text, []string{"childbirth", "maternal", "delivery"}) {
			return apply("Delivery table")
		}
		if HasAny(text, []string{"ophthalmology"}) {
			return apply("Ophthalmic instrument table")
		}
		if HasAny(text, []string{"functional exploration", "therapeutic interventions", "surgical instruments"}) {
			return apply("Instrument table")
		}
		return apply("Procedure table")
	case "Procedure table":
		if HasAny(text, []string{"examination", "treatment"}) {
			return apply("Examination table")
		}
		if HasAny(text, []string{"slit lamp"}) {
			return apply("Slit lamp table")
		}
		if HasAny(text, []string{"instrument"}) {
			return apply("Instrument table")
		}
		return apply("Procedure table")
	case "Cabinet":
		if HasAny(text, []string{"biosafety", "biological hoods", "biological cabinets"}) {
			return apply("Biosafety cabinet")
		}
		if HasAny(text, []string{"warming", "blanket warmer", "contrast media", "fluids heating"}) {
			return apply("Warming cabinet")
		}
		if HasAny(text, []string{"bedside"}) {
			return apply("Bedside cabinet")
		}
		if HasAny(text, []string{"medicines", "emergency medicines", "medicine"}) {
			return apply("Medicine cabinet")
		}
		if HasAny(text, []string{"microscope"}) {
			return apply("Microscope cabinet")
		}
		if HasAny(text, []string{"instrument"}) {
			return apply("Instrument cabinet")
		}
		return apply("Storage cabinet")
	case "Warming cabinet":
		if HasAny(text, []string{"blanket warmer"}) {
			return apply("Blanket warmer")
		}
		if HasAny(text, []string{"contrast media"}) {
			return apply("Contrast media warmer")
		}
		return apply("Warming cabinet")
	case "Rack":
		if HasAny(text, []string{"test tube"}) {
			return apply("Test tube rack")
		}
		if HasAny(text, []string{"esr", "erythrocyte sedimentation rate"}) {
			return apply("ESR pipette rack")
		}
		if HasAny(text, []string{"radiation shielding apron", "body protection"}) {
			return apply("Apron rack")
		}
		if HasAny(text, []string{"drying"}) {
			return apply("Drying rack")
		}
		if HasAny(text, []string{"retinoscopy lens", "ophthalmic lenses"}) {
			return apply("Lens rack")
		}
		if HasAny(text, []string{"staining"}) {
			return apply("Staining rack")
		}
		return apply("Storage rack")
	case "Lamp":
		if HasAny(text, []string{"slit lamp"}) {
			return apply("Slit lamp")
		}
		if HasAny(text, []string{"phototherapy"}) {
			return apply("Phototherapy lamp")
		}
		if HasAny(text, []string{"mobile scialytic", "operating room mobile"}) {
			return apply("Mobile surgical light")
		}
		if HasAny(text, []string{"fixed scialytic", "operating room ceiling"}) {
			return apply("Fixed surgical light")
		}
		if HasAny(text, []string{"scialytic lamp", "operating room", "double single head"}) {
			return apply("Surgical light")
		}
		if HasAny(text, []string{"flashlight"}) {
			return apply("Medical flashlight")
		}
		if HasAny(text, []string{"examination", "light sources"}) {
			return apply("Examination lamp")
		}
		return apply("Medical lamp")
	case "Ventilator":
		if HasAny(text, []string{"portable"}) {
			return apply("Portable ventilator")
		}
		if HasAny(text, []string{"neonatal", "paediatric", "pediatric"}) {
			return apply("Neonatal ventilator")
		}
		if HasAny(text, []string{"hospital use"}) {
			return apply("Hospital ventilator")
		}
		return apply("Ventilator", "Mechanical ventilator")
	case "Ultrasound scanner":
		if HasAny(text, []string{"bladder volume", "bladder"}) {
			return apply("Bladder scanner")
		}
		if HasAny(text, []string{"cardiology", "cardiovascular", "cardiac"}) {
			return apply("Cardiac ultrasound machine")
		}
		if HasAny(text, []string{"portable"}) {
			return apply("Portable ultrasound machine")
		}
		return apply("Ultrasound machine")
	case "Oxygen flowmeter":
		if HasAny(text, []string{"cytoflowmeter", "cytoflowmeters"}) {
			return apply("Flow cytometer")
		}
		if HasAny(text, []string{"blood flow meter", "blood flow meters"}) {
			return apply("Blood flow meter")
		}
		if HasAny(text, []string{"nasal cannula", "nasal cannulas"}) {
			return apply("Oxygen nasal cannula")
		}
		return apply("Oxygen flowmeter")
	case "Microscope":
		if HasAny(text, []string{"operative"}) {
			return apply("Operating microscope")
		}
		return apply("Laboratory microscope")
	case "Refrigerator":
		if HasAny(text, []string{"blood bank"}) {
			return apply("Blood bank refrigerator")
		}
		return apply("Medical refrigerator")
	case "Centrifuge":
		if HasAny(text, []string{"refrigerated"}) {
			return apply("Refrigerated centrifuge")
		}
		return apply("Laboratory centrifuge")
	case "Laser":
		if HasAny(text, []string{"photocoagulator"}) {
			return apply("Laser photocoagulator")
		}
		if HasAny(text, []string{"argon", "ophthalmic"}) {
			return apply("Argon laser")
		}
		if HasAny(text, []string{"carbon dioxide", "co2"}) {
			return apply("CO2 surgical laser")
		}
		if HasAny(text, []string{"nd yag", "neodymium"}) {
			return apply("Nd:YAG laser")
		}
		if HasAny(text, []string{"trabeculoplasty"}) {
			return apply("Trabeculoplasty laser")
		}
		return apply("Surgical laser")
	case "Cart":
		if HasAny(text, []string{"linen", "laundry", "clean"}) {
			return apply("Clean linen cart")
		}
		if HasAny(text, []string{"linen", "laundry", "soiled"}) {
			return apply("Soiled linen cart")
		}
		if HasAny(text, []string{"mri", "magnetic resonance imaging"}) {
			return apply("MRI equipment cart")
		}
		return apply("Medical cart")
	case "Trolley":
		if HasAny(text, []string{"defibrillator"}) {
			return apply("Defibrillator trolley")
		}
		return apply("Medical trolley")
	case "Bath":
		if HasAny(text, []string{"plasma thaw"}) {
			return apply("Plasma thawing bath")
		}
		if HasAny(text, []string{"thermostatic water bath"}) {
			return apply("Water bath")
		}
		return apply("Laboratory bath")
	case "Bed":
		if HasAny(text, []string{"intensive care", "resuscitation", "icu"}) {
			return apply("ICU bed")
		}
		if HasAny(text, []string{"pediatric"}) {
			return apply("Pediatric bed")
		}
		return apply("Hospital bed")
	case "Calliper":
		if HasAny(text, []string{"castroviejo"}) {
			return apply("Castroviejo caliper")
		}
		if HasAny(text, []string{"ophthalmic"}) {
			return apply("Ophthalmic caliper")
		}
		return apply("Measuring caliper")
	case "Fetal monitor":
		if HasAny(text, []string{"detector", "doppler"}) {
			return apply("Fetal Doppler")
		}
		if HasAny(text, []string{"continuous", "portable"}) {
			return apply("Portable fetal monitor")
		}
		if HasAny(text, []string{"bedside"}) {
			return apply("Bedside fetal monitor")
		}
		return apply("Fetal monitor")
	case "Oven":
		if HasAny(text, []string{"dry air steriliser", "dry air sterilizers", "hot air"}) {
			return apply("Hot air sterilizer")
		}
		if HasAny(text, []string{"microwave"}) {
			return apply("Laboratory microwave")
		}
		return apply("Laboratory oven")
	case "Stretcher":
		if HasAny(text, []string{"foldable"}) {
			return apply("Foldable stretcher")
		}
		if HasAny(text, []string{"mri safe", "mri-safe"}) {
			return apply("MRI-safe stretcher")
		}
		return apply("Patient stretcher")
	case "Warmer":
		if HasAny(text, []string{"radiant"}) {
			return apply("Radiant warmer")
		}
		if HasAny(text, []string{"heating pad"}) {
			return apply("Neonatal warming pad")
		}
		if HasAny(text, []string{"sleeping bag"}) {
			return apply("Neonatal warming bag")
		}
		return apply("Patient warmer")
	case "Chemistry analyzer":
		if HasAny(text, []string{"point of care", "point-of-care", "rapid tests", "poc"}) {
			return apply("Point-of-care chemistry analyzer")
		}
		return apply("Chemistry analyzer", "Clinical chemistry analyzer")
	case "Laboratory bath":
		if HasAny(text, []string{"plasma thaw"}) {
			return apply("Plasma thawing bath")
		}
		return apply("Water bath")
	case "Biometer":
		if HasAny(text, []string{"pachymeter"}) {
			return apply("Pachymeter")
		}
		if HasAny(text, []string{"optical biometer"}) {
			return apply("Optical biometer")
		}
		return apply("Biometer")
	case "Amber screw cap bottle":
		if HasAny(text, []string{"glass bottles"}) {
			return apply("Amber sample bottle")
		}
		return apply("Clinical bottle")
	case "Bracelet":
		if HasAny(text, []string{"body temperature monitoring probes"}) {
			return apply("Temperature monitoring bracelet")
		}
		return apply("Patient bracelet")
	case "Medicine cabinet":
		if HasAny(text, []string{"emergency"}) {
			return apply("Emergency medicine cabinet")
		}
		return apply("Medicine cabinet")
	case "Counter":
		if HasAny(text, []string{"limited panel"}) {
			return apply("Limited-panel cell counter")
		}
		if HasAny(text, []string{"cell counting", "cell counting instruments"}) {
			return apply("Cell counter")
		}
		return apply("Counter")
	case "Defibrillator":
		if HasAny(text, []string{"automatic", "aed", "automated external"}) {
			return apply("AED", "Automated external defibrillator")
		}
		return apply("Defibrillator")
	}
	return commonName, canonicalName
}

type ResolvedRow struct {
	DeviceType            string
	DeviceCategory        string
	DeviceFamily          string
	DeviceFunction        string
	DeviceApplicationRisk string
	Name                  string
	CanonicalName         string
	CommonNames           []string
	NamingSource          string
}

func ResolveRowNaming(row map[string]string) ResolvedRow {
	legacySourceName := row["Legacy source name"]
	sourceDeviceType := row["Source device type"]
	emdnTerm := row["EMDN term"]
	specificRule := InferSpecificNameRule(legacySourceName, sourceDeviceType, emdnTerm)
	ovaholType := ""
	if specificRule != nil && specificRule.Type != "" {
		ovaholType = specificRule.Type
	} else {
		ovaholType = InferDeviceType(legacySourceName, sourceDeviceType, emdnTerm)
	}
	if ovaholType == "" {
		// No signal from the specific-name rules, source-type mapping, or
		// device name/EMDN term keyword matching — genuinely unclassifiable,
		// not merely an unrecognized source type string.
		return ResolvedRow{NamingSource: "unsupported_source_type"}
	}
	defaults := InferDefaults(legacySourceName, sourceDeviceType, emdnTerm, ovaholType)
	generatedCommon := InferCommonNameFromLegacy(legacySourceName)
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
	commonName, canonicalName = RefineDescriptiveNames(commonName, canonicalName, legacySourceName, emdnTerm)
	family := ""
	if specificRule != nil && specificRule.Family != "" {
		family = specificRule.Family
	} else {
		family = defaults.Family
	}
	category := CategoryForFunction(defaults.Function)
	return ResolvedRow{
		DeviceType:            ovaholType,
		DeviceCategory:        category,
		DeviceFamily:          family,
		DeviceFunction:        defaults.Function,
		DeviceApplicationRisk: defaults.Risk,
		Name:                  commonName,
		CanonicalName:         canonicalName,
		CommonNames:           BuildSearchAliases(commonName, canonicalName),
		NamingSource:          namingSource,
	}
}

func BuildSearchAliases(commonName, canonicalName string) []string {
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
	combined := Normalized(strings.Join([]string{commonName, canonicalName}, " "))
	if strings.Contains(combined, "electrocardiography") || strings.Contains(combined, "ecg") {
		add("ECG machine")
		add("EKG machine")
		add("Electrocardiograph")
	}
	if strings.Contains(combined, "electroencephalography") || strings.Contains(combined, "eeg") {
		add("EEG machine")
		add("EEG accessory")
	}
	if strings.Contains(combined, "ultrasound") {
		add("Ultrasound system")
		add("Sonography system")
	}
	if strings.Contains(combined, "patient monitoring") || strings.Contains(combined, "patient monitor") {
		add("Vital signs monitor")
	}
	if strings.Contains(combined, "x ray") || strings.Contains(combined, "radiography") {
		add("Xray machine")
		add("Radiography system")
	}
	if strings.Contains(combined, "endoscope") || strings.Contains(combined, "endoscopic visualization") {
		add("Endoscope")
		add("Endoscopy system")
		add("Diagnostic scope")
	}
	if strings.Contains(combined, "angiography") {
		add("Cath lab imaging system")
		add("Angio system")
	}
	if strings.Contains(combined, "infusion") && strings.Contains(combined, "pump") {
		add("IV pump")
	}
	if strings.Contains(combined, "defibrillation") || strings.Contains(combined, "defibrillator") {
		add("AED")
	}
	if strings.Contains(combined, "dialysis") {
		add("Dialysis unit")
	}
	if strings.Contains(combined, "electrosurgical or radiotherapy energy system") {
		add("Therapeutic energy unit")
		add("Energy therapy system")
	}
	if strings.Contains(combined, "oxygen") && strings.Contains(combined, "system") {
		add("Oxygen equipment")
	}
	if strings.Contains(combined, "suction") || strings.Contains(combined, "medical suction") {
		add("Aspirator")
	}
	if strings.Contains(combined, "sterilization") || strings.Contains(combined, "sterilizer") {
		add("Autoclave")
	}
	if strings.Contains(combined, "clinical information software") {
		add("Clinical system")
	}
	if strings.Contains(combined, "clinical workflow or information software") {
		add("Clinical software")
		add("Health information system")
	}
	if strings.Contains(combined, "imaging or laboratory software") {
		add("Diagnostic system software")
	}
	if strings.Contains(combined, "imaging or laboratory informatics software") {
		add("Diagnostic software")
		add("Imaging software")
		add("Laboratory software")
	}
	if strings.Contains(combined, "rapid diagnostic or point of care test") {
		add("Rapid test")
		add("RDT")
		add("Point-of-care test")
		add("POCT")
	}
	if strings.Contains(combined, "clinical laboratory analyzer") {
		add("Laboratory analyzer")
		add("Lab analyzer")
	}
	if strings.Contains(combined, "general laboratory support") {
		add("Lab equipment")
		add("Lab support equipment")
	}
	if strings.Contains(combined, "laboratory sample preparation and support equipment") {
		add("Sample prep equipment")
		add("Laboratory support equipment")
	}
	if strings.Contains(combined, "in vitro diagnostic reagent or assay kit") {
		add("Diagnostic reagent")
		add("Assay kit")
	}
	if strings.Contains(combined, "biomedical test or calibration analyzer") {
		add("Calibration tester")
	}
	if strings.Contains(combined, "reusable surgical or procedure instrument") {
		add("Surgical tool")
		add("Surgical hand instrument")
	}
	if strings.Contains(combined, "general surgical or interventional instrument") {
		add("Procedure instrument")
		add("Interventional instrument")
	}
	if strings.Contains(combined, "general medical device accessory or disposable") {
		add("Medical consumable")
		add("Device accessory")
	}
	if strings.Contains(combined, "monitoring or neurodiagnostic accessory") {
		add("Monitoring lead")
		add("Electrode accessory")
	}
	if strings.Contains(combined, "dialysis or extracorporeal therapy consumable") {
		add("Dialysis set")
	}
	if strings.Contains(combined, "drainage or thoracic access device") {
		add("Chest drain")
	}
	if strings.Contains(combined, "airway or laryngoscopy accessory") {
		add("Laryngoscope accessory")
	}
	if strings.Contains(combined, "laboratory biosafety cabinet or fume hood") {
		add("Biosafety cabinet")
		add("Fume hood")
	}
	if strings.Contains(combined, "laboratory thermal control or cold storage device") {
		add("Lab freezer")
		add("Lab bath")
	}
	if strings.Contains(combined, "cardiac pacing or circulatory support system") {
		add("Pacemaker system")
		add("Circulatory support device")
	}
	if strings.Contains(combined, "patient warming or cooling therapy system") {
		add("Patient warmer")
		add("Cooling therapy device")
	}
	if strings.Contains(combined, "general therapeutic or life support equipment") {
		add("Therapeutic device")
		add("Life-support equipment")
	}
	if strings.Contains(combined, "neonatal therapy or support system") {
		add("Neonatal support unit")
	}
	if strings.Contains(combined, "compression immobilization or support therapy device") {
		add("Compression therapy device")
	}
	if strings.Contains(combined, "ophthalmic therapeutic or procedure system") {
		add("Ophthalmic treatment system")
	}
	if strings.Contains(combined, "ophthalmic diagnostic imaging system") {
		add("OCT machine")
	}
	if strings.Contains(combined, "assistive communication or cognitive support device") {
		add("Communication aid")
	}
	if strings.Contains(combined, "vision or hearing assistive aid") {
		add("Sensory aid")
	}
	if strings.Contains(combined, "assistive vision or hearing product") {
		add("Sensory aid")
		add("Assistive sensory product")
	}
	if strings.Contains(combined, "orthotic prosthetic or support aid") {
		add("Orthotic aid")
		add("Prosthetic aid")
	}
	if strings.Contains(combined, "assistive daily living or self care aid") {
		add("Daily living aid")
	}
	if strings.Contains(combined, "pressure relief or positioning aid") {
		add("Positioning aid")
	}
	if strings.Contains(combined, "environmental accessibility or support aid") {
		add("Accessibility aid")
	}
	if strings.Contains(combined, "clinical chart") || strings.Contains(combined, "record form") || strings.Contains(combined, "register") {
		add("Growth chart")
		add("Clinical form")
	}
	if strings.Contains(combined, "administrative or general facility support equipment") {
		add("Support equipment")
		add("General support equipment")
	}
	if strings.Contains(combined, "protective apron barrier or shield") {
		add("Radiation apron")
		add("Protective shield")
	}
	if strings.Contains(combined, "radiation or barrier protection apron or shield") {
		add("Protective apron")
		add("Barrier protection")
	}
	if strings.Contains(combined, "respiratory or anesthesia mask interface") {
		add("Anesthesia mask")
		add("Respiratory mask")
	}
	if strings.Contains(combined, "breathing circuit or ventilator accessory") {
		add("Breathing circuit")
		add("Ventilator accessory")
	}
	if strings.Contains(combined, "airway or breathing circuit consumable") {
		add("Airway accessory")
	}
	if strings.Contains(combined, "airway gas monitoring or absorber accessory") {
		add("CO2 detector")
		add("CO2 absorber")
	}
	if strings.Contains(combined, "medical gas outlet cylinder or supply item") {
		add("Medical gas supply")
	}
	if strings.Contains(combined, "surgical drape or sterile cover") {
		add("Sterile drape")
	}
	if strings.Contains(combined, "general facility utility or support asset") {
		add("Facility support equipment")
	}
	if strings.Contains(combined, "hospital furniture or bedside fixture") {
		add("Hospital furniture")
		add("Medical furniture")
	}
	if strings.Contains(combined, "ambulance or field support vehicle") {
		add("Transport vehicle")
		add("Ambulance")
	}
	if strings.Contains(combined, "administrative it or general operations support equipment") {
		add("Administrative support equipment")
		add("Support equipment")
	}
	if len(values) == 0 {
		return nil
	}
	return values
}

// Deprecated: use InferDeviceType
func InferOvaholType(deviceName, sourceType, emdnTerm string) string {
	return InferDeviceType(deviceName, sourceType, emdnTerm)
}
