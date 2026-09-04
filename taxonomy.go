// Package ontology — controlled vocabulary.
//
// This file is generated from scripts/update_ovahol_ontology.py. The
// standalone library vendors the vocabulary statically so it has no
// external dependency. To regenerate, run `go generate` or re-run the
// Python generator and copy the output.
package ontology

// DeviceType is one of Ovahol's 8 canonical device types.
type DeviceType struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// DeviceFunction is one of Ovahol's device functions, grouped by category.
type DeviceFunction struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Category string `json:"category"`
	Score    int    `json:"score"`
}

// DeviceApplicationRisk is one of Ovahol's 5 application risk levels.
type DeviceApplicationRisk struct {
	Description string `json:"description"`
	ScorePoint  int    `json:"score_point"`
}

// DeviceTypes is the canonical list of 8 device types.
var DeviceTypes = []DeviceType{
	{Name: "Monitoring & Measurement Devices", Code: "MONITORING_MEASUREMENT_DEVICES"},
	{Name: "Diagnostic & Imaging Devices", Code: "DIAGNOSTIC_IMAGING_DEVICES"},
	{Name: "Treatment, Surgical & Life Support Devices", Code: "TREATMENT_SURGICAL_LIFE_SUPPORT_DEVICES"},
	{Name: "Laboratory & IVD Equipment", Code: "LABORATORY_IVD_EQUIPMENT"},
	{Name: "Medical Gas & Respiratory Devices", Code: "MEDICAL_GAS_RESPIRATORY_DEVICES"},
	{Name: "Sterilization & Infection Control Devices", Code: "STERILIZATION_INFECTION_CONTROL_DEVICES"},
	{Name: "Support Equipment & Furniture", Code: "SUPPORT_EQUIPMENT_FURNITURE"},
	{Name: "Consumables & Accessories", Code: "CONSUMABLES_ACCESSORIES"},
}

// DeviceFunctions is the canonical list of device functions.
var DeviceFunctions = []DeviceFunction{
	{Name: "Life Support", Code: "LIFE_SUPPORT", Category: "Therapeutic", Score: 10},
	{Name: "Surgical and Intensive Care", Code: "SURGICAL_INTENSIVE_CARE", Category: "Therapeutic", Score: 9},
	{Name: "Physical Therapy and Treatment", Code: "PHYSICAL_THERAPY_TREATMENT", Category: "Therapeutic", Score: 8},
	{Name: "Surgical and Intensive Care Monitoring", Code: "CRITICAL_CARE_MONITORING", Category: "Diagnostic", Score: 7},
	{Name: "Additional Physiological Monitoring and Diagnostic", Code: "GENERAL_PHYSIOLOGICAL_MONITORING", Category: "Diagnostic", Score: 6},
	{Name: "Analytical Laboratory", Code: "ANALYTICAL_LABORATORY", Category: "Analytical", Score: 5},
	{Name: "Laboratory Accessories", Code: "LABORATORY_ACCESSORIES", Category: "Analytical", Score: 4},
	{Name: "Computers and Related", Code: "COMPUTERS_AND_IT", Category: "Analytical", Score: 3},
	{Name: "Patient Related and Other", Code: "PATIENT_RELATED_OTHER", Category: "Miscellaneous", Score: 2},
}

// DeviceApplicationRisks is the canonical list of 5 application risk levels.
var DeviceApplicationRisks = []DeviceApplicationRisk{
	{Description: "Potential patient death", ScorePoint: 5},
	{Description: "Potential patient or operator injury", ScorePoint: 4},
	{Description: "Inappropriate therapy or misdiagnosis", ScorePoint: 3},
	{Description: "Equipment damage", ScorePoint: 2},
	{Description: "No significant identified risk", ScorePoint: 1},
}

// DeviceCategory is one of the 4 high-level device categories.
type DeviceCategory struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// DeviceCategories is the canonical list of 4 device categories.
var DeviceCategories = []DeviceCategory{
	{Name: "Therapeutic", Code: "THERAPEUTIC"},
	{Name: "Diagnostic", Code: "DIAGNOSTIC"},
	{Name: "Analytical", Code: "ANALYTICAL"},
	{Name: "Miscellaneous", Code: "MISCELLANEOUS"},
}

// NamingRule mirrors Python NAMING_RULES
type NamingRule struct {
	Field       string `json:"field"`
	Rule        string `json:"rule"`
	GoodExample string `json:"goodExample,omitempty"`
	Avoid       string `json:"avoid,omitempty"`
}

var NamingRules = []NamingRule{
{Field: "Common name", Rule: "Use the short term clinicians, nurses, and biomedical teams naturally search for in daily work.", GoodExample: "Patient monitor", Avoid: "Multiparameter physiological surveillance platform"},
{Field: "Common name", Rule: "Keep it singular, generic, and manufacturer-neutral.", GoodExample: "Infusion pump", Avoid: "B. Braun Space pump"},
{Field: "Common name", Rule: "Prefer 2 to 5 words and put the core noun first when possible.", GoodExample: "Ultrasound machine", Avoid: "Advanced portable diagnostic ultrasound imaging system"},
{Field: "Canonical device name", Rule: "Use a controlled generic descriptor that is more precise than the common name but still product-neutral.", GoodExample: "Multiparameter patient monitoring system", Avoid: "Patient monitor"},
{Field: "Canonical device name", Rule: "Describe what the device is, not how a specific vendor markets it.", GoodExample: "Diagnostic ultrasound system", Avoid: "Acuson Freestyle"},
{Field: "Canonical device name", Rule: "Avoid packaging, single-use state, quantity, and procurement wording unless it defines the device class itself.", GoodExample: "Peripheral intravenous catheter", Avoid: "Catheter, sterile, single-use, adult"},
{Field: "Search aliases", Rule: "Store common abbreviations, alternate spellings, and local terms separated by commas.", GoodExample: "ECG machine, EKG machine, electrocardiograph", Avoid: "Leave empty when users search with other terms"},
}

var DeviceSheetHeaders = []string{
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

var APIImportHeaders = []string{
"name",
"device_type",
"device_category",
"device_function",
"device_application_risk",
"emdn_code",
"emdn_term",
}

var SupportedSourceTypes = map[string]struct{}{
"imaging nuclear medicine equipment": {},
"infusion devices": {},
"laboratory equipment": {},
"measurement devices": {},
"medical equipment": {},
"medical furniture": {},
"medical gas equipment": {},
"monitoring equipment": {},
}

var DirectSourceTypeMap = map[string]string{
"assistive products": "Support Equipment & Furniture",
"cleaning disinfection sterilization equipment": "Sterilization & Infection Control Devices",
"cleaning disinfection sterilization solutions": "Sterilization & Infection Control Devices",
"consumables for medical equipment": "Consumables & Accessories",
"dressings gauzes bandages pads": "Consumables & Accessories",
"imaging nuclear medicine equipment": "Diagnostic & Imaging Devices",
"in vitro diagnostic tests": "Laboratory & IVD Equipment",
"infection prevention and control ipc": "Sterilization & Infection Control Devices",
"infusion devices": "Treatment, Surgical & Life Support Devices",
"laboratory equipment": "Laboratory & IVD Equipment",
"medical charts stationery": "Support Equipment & Furniture",
"medical furniture": "Support Equipment & Furniture",
"medical gas equipment": "Medical Gas & Respiratory Devices",
"monitoring equipment": "Monitoring & Measurement Devices",
"quality assurance calibration maintenance devices": "Support Equipment & Furniture",
"radiotherapy related equipment": "Treatment, Surgical & Life Support Devices",
"rehabilitation devices": "Support Equipment & Furniture",
"respiratory accessories": "Medical Gas & Respiratory Devices",
"simulators trainers": "Support Equipment & Furniture",
"software": "Support Equipment & Furniture",
"surgical instruments trays and bowls": "Treatment, Surgical & Life Support Devices",
"sutures": "Consumables & Accessories",
"syringes and needles": "Consumables & Accessories",
"tubes and cannulae": "Consumables & Accessories",
"vector control equipment": "Support Equipment & Furniture",
"vehicles": "Support Equipment & Furniture",
"waste management products": "Support Equipment & Furniture",
}

var GenericLegacyHeads = map[string]struct{}{
"analyser": {},
"analyzer": {},
"bag": {},
"bottle": {},
"chair": {},
"cup": {},
"device": {},
"filter": {},
"holder": {},
"incubator": {},
"kit": {},
"machine": {},
"mask": {},
"meter": {},
"monitor": {},
"pump": {},
"sensor": {},
"set": {},
"system": {},
"tip": {},
"tray": {},
"unit": {},
}

var LegacyDescriptorPhrases = map[string]struct{}{
"adult": {},
"automated": {},
"bedside": {},
"continuous": {},
"disposable": {},
"long life battery": {},
"long-life battery": {},
"manual": {},
"mri safe": {},
"mri-safe": {},
"non invasive": {},
"non-invasive": {},
"portable": {},
"professional use": {},
"reusable": {},
"semi automated": {},
"semi-automated": {},
"single use": {},
"single-use": {},
"sterile": {},
}

// FamilyRule defines a device family grouping rule
type FamilyRule struct {
	Type          string   `json:"type"`
	Family        string   `json:"family"`
	CommonName    string   `json:"commonName,omitempty"`
	CanonicalName string   `json:"canonicalName,omitempty"`
	Function      string   `json:"function,omitempty"`
	Risk          string   `json:"risk,omitempty"`
	SourceTypes   []string `json:"sourceTypes,omitempty"`
	Keywords      []string `json:"keywords,omitempty"`
}

var FamilyRules = []FamilyRule{
// Function/Risk overrides paired with the SpecificNameRule Type overrides
// above, for devices.csv names whose correct Function/Risk isn't the
// DeviceTypeDefaults default for their Type.
{Type: "Diagnostic & Imaging Devices", Family: "Imaging workflow and visualization accessories", CommonName: "Multi-syringe contrast media injector", CanonicalName: "Multi-syringe contrast media injector", Function: "Surgical and Intensive Care Monitoring", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"multi syringe contrast media injector"}},
{Type: "Diagnostic & Imaging Devices", Family: "Angiography and catheterization imaging systems", CommonName: "Cardiac catheterization laboratory system", CanonicalName: "Cardiac catheterization laboratory system", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"cardiac catheterization laboratory system"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Sample preparation and laboratory support", CommonName: "CO2 blood gas electrode", CanonicalName: "CO2 blood gas electrode", Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"co2 blood gas electrode"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Diagnostic instrument monitor card", CanonicalName: "Diagnostic instrument monitor card", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"diagnostic instrument monitor card"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Dialysis and extracorporeal therapy systems", CommonName: "Extracorporeal membrane oxygenation (ECMO)", CanonicalName: "Extracorporeal membrane oxygenation (ECMO)", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"extracorporeal membrane oxygenation ecmo"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Medical label / report printer", CanonicalName: "Medical label / report printer", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"medical label report printer"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Oxygen supply systems", CommonName: "MRI-compatible oxygen canister", CanonicalName: "MRI-compatible oxygen canister", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"mri compatible oxygen canister"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Infusion and medication delivery systems", CommonName: "Parenteral / enteral solution bag", CanonicalName: "Parenteral / enteral solution bag", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"parenteral enteral solution bag"}},
{Type: "Laboratory & IVD Equipment", Family: "Sample preparation and laboratory support", CommonName: "RNA/DNA extraction system", CanonicalName: "RNA/DNA extraction system", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"rna dna extraction system"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Blood bag scale", CanonicalName: "Blood bag scale", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"blood bag scale"}},
{Type: "Laboratory & IVD Equipment", Family: "Sample preparation and laboratory support", CommonName: "Blood collection pipette", CanonicalName: "Blood collection pipette", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"blood collection pipette"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory containers and disposables", CommonName: "Capillary blood tube", CanonicalName: "Capillary blood tube", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"capillary blood tube"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory containers and disposables", CommonName: "Cell culture tube", CanonicalName: "Cell culture tube", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"cell culture tube"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Cotton applicator sticks", CanonicalName: "Cotton applicator sticks", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"cotton applicator sticks"}},
{Type: "Monitoring & Measurement Devices", Family: "Critical care monitoring systems", CommonName: "CT display monitor", CanonicalName: "CT display monitor", Function: "Surgical and Intensive Care Monitoring", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"ct display monitor"}},
{Type: "Diagnostic & Imaging Devices", Family: "General surgical and interventional instruments", CommonName: "Diagnostic examination set", CanonicalName: "Diagnostic examination set", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"diagnostic examination set"}},
{Type: "Monitoring & Measurement Devices", Family: "Critical care monitoring systems", CommonName: "Dialysis central monitor", CanonicalName: "Dialysis central monitor", Function: "Surgical and Intensive Care Monitoring", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"dialysis central monitor"}},
{Type: "Diagnostic & Imaging Devices", Family: "Imaging, ophthalmic, and procedural consumables", CommonName: "Fiducial marker (radiotherapy)", CanonicalName: "Fiducial marker (radiotherapy)", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"fiducial marker radiotherapy"}},
{Type: "Support Equipment & Furniture", Family: "Medical furniture and fixtures", CommonName: "Fixed surgical light", CanonicalName: "Fixed surgical light", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"fixed surgical light"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Hot air sterilizer", CanonicalName: "Hot air sterilizer", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"hot air sterilizer"}},
{Type: "Support Equipment & Furniture", Family: "Medical furniture and fixtures", CommonName: "Mobile surgical light", CanonicalName: "Mobile surgical light", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"mobile surgical light"}},
{Type: "Monitoring & Measurement Devices", Family: "Therapeutic energy systems", CommonName: "Nd: Yag laser", CanonicalName: "Nd: Yag laser", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"nd yag laser"}},
{Type: "Monitoring & Measurement Devices", Family: "Airway gas monitoring accessories", CommonName: "Respiratory gas monitor", CanonicalName: "Respiratory gas monitor", Function: "Surgical and Intensive Care Monitoring", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"respiratory gas monitor"}},
{Type: "Consumables & Accessories", Family: "Laboratory storage and thermal equipment", CommonName: "Specimen warming cabinet", CanonicalName: "Specimen warming cabinet", Function: "Laboratory Accessories", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"specimen warming cabinet"}},
{Type: "Diagnostic & Imaging Devices", Family: "Suction and aspiration systems", CommonName: "Surgery suction system", CanonicalName: "Surgery suction system", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"surgery suction system"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory containers and disposables", CommonName: "Test tube rack", CanonicalName: "Test tube rack", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"test tube rack"}},
{Type: "Monitoring & Measurement Devices", Family: "Therapeutic energy systems", CommonName: "Argon laser", CanonicalName: "Argon laser", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"argon laser"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory safety cabinets and hoods", CommonName: "Biosafety cabinet", CanonicalName: "Biosafety cabinet", Function: "Laboratory Accessories", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"biosafety cabinet"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Hybridization incubator", CanonicalName: "Hybridization incubator", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"hybridization incubator"}},
{Type: "Consumables & Accessories", Family: "Injection and infusion supplies", CommonName: "Intravenous line", CanonicalName: "Intravenous line", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"intravenous line"}},
{Type: "Support Equipment & Furniture", Family: "Medical furniture and fixtures", CommonName: "Surgical light", CanonicalName: "Surgical light", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"surgical light"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory containers and disposables", CommonName: "Funnel", CanonicalName: "Funnel", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"funnel"}},
{Type: "Monitoring & Measurement Devices", Family: "Critical care monitoring systems", CommonName: "Patient monitor", CanonicalName: "Multiparameter patient monitoring system", Function: "Surgical and Intensive Care Monitoring", Risk: "Potential patient or operator injury", SourceTypes: []string{"monitoring equipment"}, Keywords: []string{"monitor", "telemetry", "transducer", "vital signs"}},
{Type: "Monitoring & Measurement Devices", Family: "Cardiac diagnostic systems", CommonName: "ECG machine", CanonicalName: "Electrocardiography system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"ecg", "ekg", "holter"}},
{Type: "Monitoring & Measurement Devices", Family: "Neurophysiology monitoring systems", CommonName: "EEG machine", CanonicalName: "Electroencephalography system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"eeg", "electroencephalograph", "photostimulator"}},
{Type: "Monitoring & Measurement Devices", Family: "Neurophysiology accessories", CommonName: "EEG accessory", CanonicalName: "Electroencephalography accessory", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{"accessories"}, Keywords: []string{"electro cap", "electrode positioning", "electrode eeg", "headset egg", "eeg headset"}},
{Type: "Monitoring & Measurement Devices", Family: "Anthropometric and clinical measurement tools", CommonName: "Clinical measuring device", CanonicalName: "Anthropometric or bedside measuring device", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{"measurement devices"}, Keywords: []string{"caliper", "goniometer", "measuring board", "muac", "ruler", "scale", "stadiometer", "stopwatch", "tape measure", "thermometer", "timer", "pedometer"}},
{Type: "Monitoring & Measurement Devices", Family: "Physiological assessment tools", CommonName: "Diagnostic monitor", CanonicalName: "Physiological assessment device", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"tonometer", "audiometer", "bilirubinometer", "spirom", "blood pressure", "stethoscope", "calliper"}},
{Type: "Diagnostic & Imaging Devices", Family: "Ultrasound systems", CommonName: "Ultrasound machine", CanonicalName: "Diagnostic ultrasound system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"ultrasound", "sonograph"}},
{Type: "Diagnostic & Imaging Devices", Family: "Radiography and fluoroscopy systems", CommonName: "X-ray machine", CanonicalName: "Diagnostic radiography or fluoroscopy system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"x ray", "xray", "radiograph", "fluoroscopy", "mammograph"}},
{Type: "Diagnostic & Imaging Devices", Family: "Advanced imaging systems", CommonName: "Imaging scanner", CanonicalName: "Advanced diagnostic imaging system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"ct", "computed tomography", "mri", "magnetic resonance", "gamma camera", "pet", "spect"}},
{Type: "Diagnostic & Imaging Devices", Family: "Angiography and catheterization imaging systems", CommonName: "Angiography system", CanonicalName: "Angiography or catheterization imaging system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"angiography", "catheterization lab", "fluorescein angiography"}},
{Type: "Diagnostic & Imaging Devices", Family: "Imaging workflow and visualization accessories", CommonName: "Imaging accessory", CanonicalName: "Diagnostic imaging accessory or visualization aid", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"contrast injector", "injector, contrast", "negatoscope", "digital storage media", "fiducial marker", "film, dosimetry system"}},
{Type: "Diagnostic & Imaging Devices", Family: "Ophthalmic imaging systems", CommonName: "Ophthalmic imaging unit", CanonicalName: "Ophthalmic diagnostic imaging system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"optical coherence tomography", "oct unit"}},
{Type: "Diagnostic & Imaging Devices", Family: "Nuclear medicine probes and detectors", CommonName: "Gamma probe", CanonicalName: "Nuclear medicine probe or detector", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"gamma probe", "gamma probes"}},
{Type: "Diagnostic & Imaging Devices", Family: "Endoscopic visualization systems", CommonName: "Endoscopic scope", CanonicalName: "Diagnostic endoscope or endoscopic visualization system", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"endoscope", "bronchoscope", "cystoscope", "colposcope", "laparoscope", "arthroscope", "otoscope", "ophthalmoscope"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory analyzers", CommonName: "Laboratory analyzer", CanonicalName: "Clinical laboratory analyzer", Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"analyzer", "analyser", "hematology", "chemistry", "blood gas", "pcr"}},
{Type: "Laboratory & IVD Equipment", Family: "Rapid diagnostic tests", CommonName: "Rapid test kit", CanonicalName: "Rapid diagnostic or point-of-care test", Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{"in vitro diagnostic tests"}, Keywords: []string{"rdt", "rapid diagnostic test", "rapid test", "point of care", "poc"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory microscopy and pathology equipment", CommonName: "Laboratory microscope", CanonicalName: "Microscopy or pathology laboratory device", Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"microscope", "histology", "pathology", "stain"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory safety cabinets and hoods", CommonName: "Lab safety cabinet", CanonicalName: "Laboratory biosafety cabinet or fume hood", Function: "Laboratory Accessories", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"biosafety", "fume hood", "grossing bench", "grossing table", "biological hood", "cabinet, biosafety"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory storage and thermal equipment", CommonName: "Lab freezer or bath", CanonicalName: "Laboratory thermal control or cold storage device", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"freezer", "plasma thaw", "cryoprecipitate", "water bath", "hot water", "deep freezer", "thawing"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory flow and cell analysis systems", CommonName: "Flow cytometer", CanonicalName: "Flow or cell analysis laboratory system", Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"flow cytometer", "cell counter", "counting chamber", "haemoglobinometer", "hemoglobinometer", "spectrometer"}},
{Type: "Laboratory & IVD Equipment", Family: "Sample preparation and laboratory support", CommonName: "Lab prep equipment", CanonicalName: "Laboratory sample preparation and support equipment", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"centrifuge", "incubator", "sample", "specimen", "pipette"}},
{Type: "Laboratory & IVD Equipment", Family: "Laboratory containers and disposables", CommonName: "Lab consumable", CanonicalName: "Laboratory container, filter, or disposable", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"bottle", "filter", "funnel", "applicator sticks", "film, sealing", "test tubes", "paper filter"}},
{Type: "Laboratory & IVD Equipment", Family: "IVD reagents and assays", CommonName: "IVD reagent", CanonicalName: "In vitro diagnostic reagent or assay kit", Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"reagent", "assay", "ivd", "solution ivd"}},
{Type: "Laboratory & IVD Equipment", Family: "General laboratory support equipment", CommonName: "Laboratory support equipment", CanonicalName: "General laboratory support equipment", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{"laboratory equipment"}, Keywords: []string{}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Infusion and medication delivery systems", CommonName: "Infusion pump", CanonicalName: "Infusion or syringe pump system", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"infusion", "syringe pump"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Defibrillation and resuscitation systems", CommonName: "Defibrillator", CanonicalName: "Defibrillation or resuscitation system", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"defibrillator", "resuscitator", "cardioverter"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Cardiac pacing and circulatory support systems", CommonName: "Cardiac support system", CanonicalName: "Cardiac pacing or circulatory support system", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"pacemaker", "balloon pump", "aortic pump", "cardiopulmonary bypass", "cardiac ablation", "cable electrode", "lead"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Anesthesia and critical care therapy systems", CommonName: "Anesthesia machine", CanonicalName: "Anesthesia or critical care therapy system", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"anaesthesia", "anesthesia", "ventilator", "icu"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Patient warming and cooling systems", CommonName: "Thermal therapy system", CanonicalName: "Patient warming or cooling therapy system", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"warmer", "cooling mattress", "hypothermia", "blood heater", "fluid warmer", "phototherapy"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Neonatal therapy and support systems", CommonName: "Neonatal support device", CanonicalName: "Neonatal therapy or support system", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"newborn", "neonatal", "incubator", "fhr detector", "foetal heart"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Compression and immobilization therapy devices", CommonName: "Compression therapy device", CanonicalName: "Compression, immobilization, or support therapy device", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"compression", "binder", "tourniquet", "garment", "calf compression"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Ophthalmic therapy systems", CommonName: "Ophthalmic treatment device", CanonicalName: "Ophthalmic therapeutic or procedure system", Function: "Surgical and Intensive Care", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"corneal topography", "keratometer", "gonioscopy", "pachymeter", "phacoemulsification", "fundus camera", "lensmeter", "optical biometer", "biometer", "prism test", "fusion test"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Neuromuscular and rehabilitation therapy systems", CommonName: "Therapy device", CanonicalName: "Neuromuscular or therapeutic support device", Function: "Physical Therapy and Treatment", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"biofeedback", "electromiography", "emg", "monofilament"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Medication and access therapy consumables", CommonName: "Therapy access consumable", CanonicalName: "Therapeutic access or medication delivery consumable", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"closed system drug transfer", "lancet", "filter, leukocyte removal", "enema"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Dialysis and extracorporeal therapy systems", CommonName: "Dialysis machine", CanonicalName: "Dialysis or extracorporeal therapy system", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"dialysis", "apheresis", "extracorporeal"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Therapeutic energy systems", CommonName: "Therapeutic energy system", CanonicalName: "Electrosurgical or radiotherapy energy system", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"electrosurgical", "radiotherapy"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "General therapeutic and life-support devices", CommonName: "Therapy support equipment", CanonicalName: "General therapeutic or life-support equipment", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{"implantable devices", "medical equipment", "accessories"}, Keywords: []string{}},
{Type: "Medical Gas & Respiratory Devices", Family: "Oxygen supply systems", CommonName: "Oxygen system", CanonicalName: "Medical oxygen supply system", Function: "Life Support", Risk: "Potential patient death", SourceTypes: []string{}, Keywords: []string{"oxygen", "psa", "concentrator", "flowmeter"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Suction and aspiration systems", CommonName: "Suction machine", CanonicalName: "Medical suction or aspiration system", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"suction", "aspirator"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Respiratory masks and interfaces", CommonName: "Respiratory mask", CanonicalName: "Respiratory or anesthesia mask/interface", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"mask", "venturi", "manual resuscitator"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Breathing circuits and ventilator accessories", CommonName: "Ventilation circuit", CanonicalName: "Breathing circuit or ventilator accessory", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"ventilation circuit", "ventilator set", "peep", "valve assembly"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Airway gas monitoring accessories", CommonName: "Airway gas accessory", CanonicalName: "Airway gas monitoring or absorber accessory", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"carbon dioxide absorber", "end tidal co2", "colorimetric"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Medical gas outlets and cylinders", CommonName: "Medical gas supply", CanonicalName: "Medical gas outlet, cylinder, or supply item", Function: "Life Support", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"nitrous oxide", "gas equipment"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Respiratory mouthpieces and adjuncts", CommonName: "Respiratory adjunct", CanonicalName: "Respiratory mouthpiece or airway adjunct", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"mouth gag"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Respiratory consumables and materials", CommonName: "Respiratory consumable", CanonicalName: "Respiratory gas or airway consumable", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"gas tamponade"}},
{Type: "Medical Gas & Respiratory Devices", Family: "Airway and breathing consumables", CommonName: "Breathing circuit consumable", CanonicalName: "Airway or breathing circuit consumable", Function: "Surgical and Intensive Care", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"airway", "breathing circuit", "cpap", "bipap", "humidifier", "nebul"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Surgical instrument sets", CommonName: "Surgical instrument set", CanonicalName: "Reusable surgical instrument set", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{"surgical instruments trays and bowls"}, Keywords: []string{"surgical instrument", "tray", "bowl", "set"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Reusable surgical hand instruments", CommonName: "Surgical instrument", CanonicalName: "Reusable surgical or procedure instrument", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{"surgical instruments trays and bowls"}, Keywords: []string{"adenotome", "knife", "blade", "chisel", "handle", "forceps", "scissors", "curette", "speculum"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Biopsy and sampling instruments", CommonName: "Biopsy instrument", CanonicalName: "Biopsy, aspiration, or sampling instrument", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"biopsy", "bone marrow", "cervical cytology", "aspiration set"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Dilation and access instruments", CommonName: "Dilating instrument", CanonicalName: "Dilation, access, or obturation instrument", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"bougie", "dilator", "obturator"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Clinical trays and bowls", CommonName: "Procedure bowl or tray", CanonicalName: "Clinical bowl, tray, or sterilization basket", Function: "Surgical and Intensive Care", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"bowl", "basket", "tray", "emesis", "dish", "gallipot"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Clamps, clips, and approximators", CommonName: "Surgical clamp or clip", CanonicalName: "Surgical clamp, clip, or approximator", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"approximator", "clamp", "clip", "haemoclip", "endoclip", "raney"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Hooks, retractors, and dissectors", CommonName: "Surgical retractor or dissector", CanonicalName: "Surgical hook, retractor, or dissector", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"hook", "retractor", "dissector", "spatula"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Powered and fixation instruments", CommonName: "Surgical drill or driver", CanonicalName: "Powered surgical or fixation instrument", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"drill", "driver", "screw", "craniotomy", "crosslink", "compressor"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Electrosurgery and stapling accessories", CommonName: "Electrosurgery accessory", CanonicalName: "Electrosurgery, clipping, or stapling accessory", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"electrode", "stapler", "cautery"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Drainage and thoracic access devices", CommonName: "Chest tube kit", CanonicalName: "Drainage or thoracic access device", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"chest tube", "chest aspiration", "thoracentesis", "paracentesis", "drainage"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Ophthalmic and ENT procedure devices", CommonName: "ENT or ophthalmic device", CanonicalName: "ENT or ophthalmic procedure device", Function: "Surgical and Intensive Care", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"otoacoustic", "autorefractor", "vitrectomy", "ophthalmic", "ent"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Surgical drapes and gowns", CommonName: "Surgical drape or gown", CanonicalName: "Surgical drape, gown, or theatre textile", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"infection prevention and control ipc"}, Keywords: []string{"drape", "gown", "cap, surgical"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Procedure support devices", CommonName: "Procedure room equipment", CanonicalName: "Surgical or interventional support equipment", Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{"medical equipment"}, Keywords: []string{"fundus camera", "ergometer", "cross cylinder", "stress test"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Surgical device accessories", CommonName: "Surgical accessory", CanonicalName: "Surgical or interventional device accessory", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{"accessories"}, Keywords: []string{"hardware accessory", "cautery", "case,", "probe"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Surgical planning and modelling tools", CommonName: "Foam cutter", CanonicalName: "Hot-wire foam cutter for clinical modelling", Function: "Surgical and Intensive Care", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"hot wire foam", "foam cutter"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "General surgical and interventional instruments", CommonName: "Surgical instrument", CanonicalName: "General surgical or interventional instrument", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{"surgical instruments trays and bowls", "catheters and related", "oral devices"}, Keywords: []string{}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Interventional access devices", CommonName: "Catheter or cannula", CanonicalName: "Interventional access device", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"trocar", "catheter", "cannula", "laparoscopy", "access port"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Implants and implantable therapeutic devices", CommonName: "Implantable device", CanonicalName: "Implantable surgical or therapeutic device", Function: "Surgical and Intensive Care", Risk: "Potential patient death", SourceTypes: []string{"implantable devices"}, Keywords: []string{"implant", "pacemaker", "annuloplasty", "septal", "aneurysm clip", "defibrillator implantable"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Reproductive and contraceptive procedure devices", CommonName: "Contraceptive device", CanonicalName: "Reproductive or contraceptive procedure device", Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury", SourceTypes: []string{"contraception devices"}, Keywords: []string{"intra uterine", "iud", "levonorgestrel", "cervical cap"}},
{Type: "Treatment, Surgical & Life Support Devices", Family: "Dental and oral procedure instruments", CommonName: "Dental instrument", CanonicalName: "Dental or oral procedure instrument", Function: "Surgical and Intensive Care", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{"oral devices"}, Keywords: []string{"dental", "oral", "root", "burr", "mirror", "elevator"}},
{Type: "Support Equipment & Furniture", Family: "Mobility and transfer aids", CommonName: "Mobility aid", CanonicalName: "Patient mobility or transfer aid", Function: "Physical Therapy and Treatment", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"wheelchair", "walker", "crutch", "transfer", "lift"}},
{Type: "Support Equipment & Furniture", Family: "Communication and cognitive support aids", CommonName: "Communication aid", CanonicalName: "Assistive communication or cognitive support device", Function: "Patient Related and Other", Risk: "No significant identified risk", SourceTypes: []string{}, Keywords: []string{"alarm", "communication", "audiobook", "braille", "screen reader", "speaking", "mobile phone", "personal digital assistant", "orientation board", "watch", "electronic navigation", "gesture to voice"}},
{Type: "Support Equipment & Furniture", Family: "Vision and hearing assistive aids", CommonName: "Vision or hearing aid", CanonicalName: "Assistive vision or hearing product", Function: "Patient Related and Other", Risk: "No significant identified risk", SourceTypes: []string{}, Keywords: []string{"contact lens", "spectacle", "prism glasses", "magnifier", "telescope", "headset", "hearing aid", "trial lens", "visual aids"}},
{Type: "Support Equipment & Furniture", Family: "Orthotic and prosthetic supports", CommonName: "Orthotic or prosthetic aid", CanonicalName: "Orthotic, prosthetic, or support aid", Function: "Physical Therapy and Treatment", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"orthoses", "prostheses", "prosthetic", "stump socks", "off-loading", "ulcer management", "collar", "conformer", "orbital", "hip protector"}},
{Type: "Support Equipment & Furniture", Family: "Daily living and self-care aids", CommonName: "Daily living aid", CanonicalName: "Assistive daily living or self-care aid", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"feeding cup", "straws", "pill organizers", "suppository inserter", "grooming", "grabber", "cup, medicine"}},
{Type: "Support Equipment & Furniture", Family: "Positioning and pressure relief aids", CommonName: "Pressure relief aid", CanonicalName: "Pressure relief or positioning aid", Function: "Physical Therapy and Treatment", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"pressure relief", "mattress", "cushion", "standing frame"}},
{Type: "Support Equipment & Furniture", Family: "Environmental accessibility aids", CommonName: "Accessibility aid", CanonicalName: "Environmental accessibility or support aid", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"handrail", "ramps", "toilet seat", "seat, shower", "mat, anti-slip", "tricycles"}},
{Type: "Support Equipment & Furniture", Family: "General rehabilitation and assistive aids", CommonName: "Assistive product", CanonicalName: "General rehabilitation or assistive product", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"medical equipment", "accessories", "implantable devices", "assistive products", "rehabilitation devices", "oral devices"}, Keywords: []string{}},
{Type: "Support Equipment & Furniture", Family: "Continence and daily living aids", CommonName: "Continence aid", CanonicalName: "Continence or daily living assistive product", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"incontinence", "ostomy", "stoma"}},
{Type: "Support Equipment & Furniture", Family: "Rehabilitation therapy devices", CommonName: "Rehab device", CanonicalName: "Rehabilitation therapy device", Function: "Physical Therapy and Treatment", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"rehabilitation", "physiotherapy", "therapy"}},
{Type: "Sterilization & Infection Control Devices", Family: "Sterilization equipment", CommonName: "Sterilizer", CanonicalName: "Sterilization or decontamination equipment", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"autoclave", "steriliz", "washer disinfector", "decontamin"}},
{Type: "Sterilization & Infection Control Devices", Family: "Disinfection and sterilization consumables", CommonName: "Disinfection chemical", CanonicalName: "Disinfection or sterilization consumable", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"disinfect", "antimicrobial", "cleaning", "peracetic", "decontaminating solution"}},
{Type: "Sterilization & Infection Control Devices", Family: "Protective barriers and PPE", CommonName: "Barrier PPE", CanonicalName: "Protective barrier or PPE item", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"glove", "mask", "gown", "apron", "clogs", "laboratory coat", "cap, surgical", "headwear"}},
{Type: "Sterilization & Infection Control Devices", Family: "Surgical drapes and covers", CommonName: "Surgical drape", CanonicalName: "Surgical drape or sterile cover", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"drape, surgical", "shoe cover", "shoe covers", "surgical trousers", "trousers surgical"}},
{Type: "Support Equipment & Furniture", Family: "Medical furniture and fixtures", CommonName: "Clinical furniture", CanonicalName: "Hospital furniture or bedside fixture", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"medical furniture"}, Keywords: []string{"bed", "cabinet", "couch", "bedscreen"}},
{Type: "Support Equipment & Furniture", Family: "Clinical charts and stationery", CommonName: "Clinical record form", CanonicalName: "Clinical chart, register, or record form", Function: "Patient Related and Other", Risk: "No significant identified risk", SourceTypes: []string{"medical charts stationery"}, Keywords: []string{"chart", "card", "stationery", "form"}},
{Type: "Support Equipment & Furniture", Family: "Protective aprons and shields", CommonName: "Protective apron or shield", CanonicalName: "Radiation or barrier protection apron or shield", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{"personal protective equipment radiation protection equipment"}, Keywords: []string{"apron", "shielding", "breast protection", "radiation source handling"}},
{Type: "Support Equipment & Furniture", Family: "Transport and field support", CommonName: "Medical transport vehicle", CanonicalName: "Ambulance or field support vehicle", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{"vehicles"}, Keywords: []string{"ambulance", "vehicle"}},
{Type: "Support Equipment & Furniture", Family: "Waste and environmental control", CommonName: "Waste management item", CanonicalName: "Waste or environmental support item", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"waste", "biohazard"}},
{Type: "Support Equipment & Furniture", Family: "Vector control and environmental health equipment", CommonName: "Vector control equipment", CanonicalName: "Vector control or environmental health device", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"vector control equipment"}, Keywords: []string{"vector"}},
{Type: "Support Equipment & Furniture", Family: "Administrative and support equipment", CommonName: "Operations support equipment", CanonicalName: "Administrative, IT, or general operations support equipment", Function: "Computers and Related", Risk: "No significant identified risk", SourceTypes: []string{}, Keywords: []string{"computer", "printer", "clock", "data logger", "lamp"}},
{Type: "Support Equipment & Furniture", Family: "Radiation protection and shielding", CommonName: "Radiation shielding", CanonicalName: "Radiation protection or shielding device", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"radiation shielding", "shielding", "radiation source handling"}},
{Type: "Support Equipment & Furniture", Family: "General non-clinical supplies", CommonName: "Support supply", CanonicalName: "General non-clinical support supply", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"non medical devices"}, Keywords: []string{"ear cotton", "marker", "loudspeaker", "box cutter"}},
{Type: "Support Equipment & Furniture", Family: "General facility and support assets", CommonName: "Facility support item", CanonicalName: "General facility, utility, or support asset", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"medical equipment", "personal protective equipment radiation protection equipment", "accessories", "medical charts stationery"}, Keywords: []string{}},
{Type: "Support Equipment & Furniture", Family: "Calibration and test analyzers", CommonName: "Biomedical tester", CanonicalName: "Biomedical test or calibration analyzer", Function: "Laboratory Accessories", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"calibration", "testing system", "tester", "analyzer", "analyser"}},
{Type: "Support Equipment & Furniture", Family: "Simulators and training devices", CommonName: "Training simulator", CanonicalName: "Simulation or training device", Function: "Laboratory Accessories", Risk: "No significant identified risk", SourceTypes: []string{}, Keywords: []string{"simulator", "trainer", "phantom"}},
{Type: "Support Equipment & Furniture", Family: "Clinical information systems", CommonName: "Clinical information system", CanonicalName: "Clinical workflow or information software", Function: "Computers and Related", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"software", "emr", "ehr", "clinical information system"}},
{Type: "Support Equipment & Furniture", Family: "Imaging and laboratory software", CommonName: "Diagnostic informatics software", CanonicalName: "Imaging or laboratory informatics software", Function: "Computers and Related", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"pacs", "lis", "quality control software", "photoscreener software", "programmer"}},
{Type: "Consumables & Accessories", Family: "Wound care and dressings", CommonName: "Dressing", CanonicalName: "Wound care dressing or absorbent supply", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"dressing", "gauze", "bandage", "pad", "swab"}},
{Type: "Consumables & Accessories", Family: "Monitoring and neurodiagnostic accessories", CommonName: "Monitoring accessory", CanonicalName: "Monitoring or neurodiagnostic accessory", Function: "Patient Related and Other", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"electrode", "lead", "patient cable", "eeg", "tens"}},
{Type: "Consumables & Accessories", Family: "Imaging, ophthalmic, and procedural consumables", CommonName: "Procedure consumable", CanonicalName: "Imaging, ophthalmic, or procedure consumable", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"gel", "fluorescein", "lens", "phacoemulsification", "ophthalmoscope", "irrigation", "coupling"}},
{Type: "Consumables & Accessories", Family: "Airway and laryngoscopy accessories", CommonName: "Airway accessory", CanonicalName: "Airway or laryngoscopy accessory", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"laryngoscope", "airway", "ear tips"}},
{Type: "Consumables & Accessories", Family: "Dialysis and extracorporeal consumables", CommonName: "Dialysis consumable", CanonicalName: "Dialysis or extracorporeal therapy consumable", Function: "Patient Related and Other", Risk: "Potential patient or operator injury", SourceTypes: []string{}, Keywords: []string{"haemodialysis", "hemodialysis", "heat exchanger", "filter, haemodialysis"}},
{Type: "Consumables & Accessories", Family: "Labels, packaging, and handling supplies", CommonName: "Label or package supply", CanonicalName: "Label, packaging, or handling consumable", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"label", "carry-bag", "cover", "box"}},
{Type: "Consumables & Accessories", Family: "Injection and infusion supplies", CommonName: "Injection supply", CanonicalName: "Injection or infusion consumable", Function: "Patient Related and Other", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"syringe", "needle", "infusion set"}},
{Type: "Consumables & Accessories", Family: "Sutures and wound closure supplies", CommonName: "Suture", CanonicalName: "Suture or wound closure consumable", Function: "Patient Related and Other", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{"sutures"}, Keywords: []string{"suture"}},
{Type: "Consumables & Accessories", Family: "Tubes, cannulae, and airway consumables", CommonName: "Tube or cannula", CanonicalName: "Tube, cannula, or airway consumable", Function: "Patient Related and Other", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"tube", "cannula", "airway"}},
{Type: "Consumables & Accessories", Family: "Collection, drainage, and ostomy supplies", CommonName: "Drainage or collection bag", CanonicalName: "Collection, drainage, or ostomy consumable", Function: "Patient Related and Other", Risk: "Inappropriate therapy or misdiagnosis", SourceTypes: []string{}, Keywords: []string{"drainage", "collection", "ostomy", "stoma", "bag"}},
{Type: "Consumables & Accessories", Family: "Procedure packs and accessories", CommonName: "Procedure accessory", CanonicalName: "Procedure accessory or disposable supply", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"accessory", "procedure", "single use", "pack"}},
{Type: "Consumables & Accessories", Family: "Reproductive health supplies", CommonName: "Contraceptive device", CanonicalName: "Reproductive health or contraception supply", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{}, Keywords: []string{"condom", "diaphragm", "cervical cap", "spermicide"}},
{Type: "Consumables & Accessories", Family: "General device accessories and disposables", CommonName: "Device accessory", CanonicalName: "General medical device accessory or disposable", Function: "Patient Related and Other", Risk: "Equipment damage", SourceTypes: []string{"accessories", "consumables for medical equipment"}, Keywords: []string{}},
}

type SpecificNameRule struct {
	Keywords        []string `json:"keywords,omitempty"`
	ExcludeKeywords []string `json:"excludeKeywords,omitempty"`
	CommonName      string   `json:"commonName,omitempty"`
	CanonicalName   string   `json:"canonicalName,omitempty"`
	Type            string   `json:"type,omitempty"`
	Family          string   `json:"family,omitempty"`
}

var SpecificNameRules = []SpecificNameRule{
// The following overrides pin exact Ovahol device dictionary names to their
// production Type (see devices.csv) so keyword inference can't outvote them.
// Ordered longest-keyword-first so multi-word overrides aren't shadowed by
// shorter overrides later in this block (e.g. "Fixed surgical light" before
// "Surgical light").
{Keywords: []string{"blood glucose meter point of care"}, ExcludeKeywords: []string{}, CommonName: "Blood glucose meter (point-of-care)", CanonicalName: "Blood glucose meter (point-of-care)", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"point of care pt inr analyzer"}, ExcludeKeywords: []string{}, CommonName: "Point-of-care pt / inr analyzer", CanonicalName: "Point-of-care pt / inr analyzer", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"in situ hybridization staining platform"}, ExcludeKeywords: []string{}, CommonName: "In situ hybridization staining platform", CanonicalName: "In situ hybridization staining platform", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"multi syringe contrast media injector"}, ExcludeKeywords: []string{}, CommonName: "Multi-syringe contrast media injector", CanonicalName: "Multi-syringe contrast media injector", Type: "Diagnostic & Imaging Devices"},
{Keywords: []string{"cardiac catheterization laboratory system"}, ExcludeKeywords: []string{}, CommonName: "Cardiac catheterization laboratory system", CanonicalName: "Cardiac catheterization laboratory system", Type: "Diagnostic & Imaging Devices"},
{Keywords: []string{"co2 blood gas electrode"}, ExcludeKeywords: []string{}, CommonName: "CO2 blood gas electrode", CanonicalName: "CO2 blood gas electrode", Type: "Treatment, Surgical & Life Support Devices"},
{Keywords: []string{"diagnostic instrument monitor card"}, ExcludeKeywords: []string{}, CommonName: "Diagnostic instrument monitor card", CanonicalName: "Diagnostic instrument monitor card", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"extracorporeal membrane oxygenation ecmo"}, ExcludeKeywords: []string{}, CommonName: "Extracorporeal membrane oxygenation (ECMO)", CanonicalName: "Extracorporeal membrane oxygenation (ECMO)", Type: "Treatment, Surgical & Life Support Devices"},
{Keywords: []string{"medical label report printer"}, ExcludeKeywords: []string{}, CommonName: "Medical label / report printer", CanonicalName: "Medical label / report printer", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"mri compatible infusion pump"}, ExcludeKeywords: []string{}, CommonName: "MRI-compatible infusion pump", CanonicalName: "MRI-compatible infusion pump", Type: "Treatment, Surgical & Life Support Devices"},
{Keywords: []string{"mri compatible oxygen canister"}, ExcludeKeywords: []string{}, CommonName: "MRI-compatible oxygen canister", CanonicalName: "MRI-compatible oxygen canister", Type: "Medical Gas & Respiratory Devices"},
{Keywords: []string{"next generation sequencing system"}, ExcludeKeywords: []string{}, CommonName: "Next generation sequencing system", CanonicalName: "Next generation sequencing system", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"parenteral enteral solution bag"}, ExcludeKeywords: []string{}, CommonName: "Parenteral / enteral solution bag", CanonicalName: "Parenteral / enteral solution bag", Type: "Treatment, Surgical & Life Support Devices"},
{Keywords: []string{"radiation protection apron rack"}, ExcludeKeywords: []string{}, CommonName: "Radiation protection apron rack", CanonicalName: "Radiation protection apron rack", Type: "Support Equipment & Furniture"},
{Keywords: []string{"rna dna extraction system"}, ExcludeKeywords: []string{}, CommonName: "RNA/DNA extraction system", CanonicalName: "RNA/DNA extraction system", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"surgical instrument trolley table"}, ExcludeKeywords: []string{}, CommonName: "Surgical instrument trolley / table", CanonicalName: "Surgical instrument trolley / table", Type: "Support Equipment & Furniture"},
{Keywords: []string{"blood bag scale"}, ExcludeKeywords: []string{}, CommonName: "Blood bag scale", CanonicalName: "Blood bag scale", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"blood collection pipette"}, ExcludeKeywords: []string{}, CommonName: "Blood collection pipette", CanonicalName: "Blood collection pipette", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"capillary blood tube"}, ExcludeKeywords: []string{}, CommonName: "Capillary blood tube", CanonicalName: "Capillary blood tube", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"cell culture tube"}, ExcludeKeywords: []string{}, CommonName: "Cell culture tube", CanonicalName: "Cell culture tube", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"clinical instrument stand"}, ExcludeKeywords: []string{}, CommonName: "Clinical instrument stand", CanonicalName: "Clinical instrument stand", Type: "Support Equipment & Furniture"},
{Keywords: []string{"cotton applicator sticks"}, ExcludeKeywords: []string{}, CommonName: "Cotton applicator sticks", CanonicalName: "Cotton applicator sticks", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"ct display monitor"}, ExcludeKeywords: []string{}, CommonName: "CT display monitor", CanonicalName: "CT display monitor", Type: "Monitoring & Measurement Devices"},
{Keywords: []string{"diagnostic examination set"}, ExcludeKeywords: []string{}, CommonName: "Diagnostic examination set", CanonicalName: "Diagnostic examination set", Type: "Diagnostic & Imaging Devices"},
{Keywords: []string{"dialysis central monitor"}, ExcludeKeywords: []string{}, CommonName: "Dialysis central monitor", CanonicalName: "Dialysis central monitor", Type: "Monitoring & Measurement Devices"},
{Keywords: []string{"fiducial marker radiotherapy"}, ExcludeKeywords: []string{}, CommonName: "Fiducial marker (radiotherapy)", CanonicalName: "Fiducial marker (radiotherapy)", Type: "Diagnostic & Imaging Devices"},
{Keywords: []string{"fixed surgical light"}, ExcludeKeywords: []string{}, CommonName: "Fixed surgical light", CanonicalName: "Fixed surgical light", Type: "Support Equipment & Furniture"},
{Keywords: []string{"histology staining rack"}, ExcludeKeywords: []string{}, CommonName: "Histology staining rack", CanonicalName: "Histology staining rack", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"hot air sterilizer"}, ExcludeKeywords: []string{}, CommonName: "Hot air sterilizer", CanonicalName: "Hot air sterilizer", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"iv infusion pole"}, ExcludeKeywords: []string{}, CommonName: "IV infusion pole", CanonicalName: "IV infusion pole", Type: "Support Equipment & Furniture"},
{Keywords: []string{"mobile surgical light"}, ExcludeKeywords: []string{}, CommonName: "Mobile surgical light", CanonicalName: "Mobile surgical light", Type: "Support Equipment & Furniture"},
{Keywords: []string{"mri equipment cart"}, ExcludeKeywords: []string{}, CommonName: "MRI equipment cart", CanonicalName: "MRI equipment cart", Type: "Support Equipment & Furniture"},
{Keywords: []string{"mri positioning footrest"}, ExcludeKeywords: []string{}, CommonName: "MRI positioning footrest", CanonicalName: "MRI positioning footrest", Type: "Support Equipment & Furniture"},
{Keywords: []string{"nd yag laser"}, ExcludeKeywords: []string{}, CommonName: "Nd: Yag laser", CanonicalName: "Nd: Yag laser", Type: "Monitoring & Measurement Devices"},
{Keywords: []string{"ophthalmic examination table"}, ExcludeKeywords: []string{}, CommonName: "Ophthalmic examination table", CanonicalName: "Ophthalmic examination table", Type: "Support Equipment & Furniture"},
{Keywords: []string{"ophthalmic instrument stand"}, ExcludeKeywords: []string{}, CommonName: "Ophthalmic instrument stand", CanonicalName: "Ophthalmic instrument stand", Type: "Support Equipment & Furniture"},
{Keywords: []string{"respiratory gas monitor"}, ExcludeKeywords: []string{}, CommonName: "Respiratory gas monitor", CanonicalName: "Respiratory gas monitor", Type: "Monitoring & Measurement Devices"},
{Keywords: []string{"specimen warming cabinet"}, ExcludeKeywords: []string{}, CommonName: "Specimen warming cabinet", CanonicalName: "Specimen warming cabinet", Type: "Consumables & Accessories"},
{Keywords: []string{"surgery suction system"}, ExcludeKeywords: []string{}, CommonName: "Surgery suction system", CanonicalName: "Surgery suction system", Type: "Diagnostic & Imaging Devices"},
{Keywords: []string{"test tube rack"}, ExcludeKeywords: []string{}, CommonName: "Test tube rack", CanonicalName: "Test tube rack", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"treatment examination table"}, ExcludeKeywords: []string{}, CommonName: "Treatment / examination table", CanonicalName: "Treatment / examination table", Type: "Support Equipment & Furniture"},
{Keywords: []string{"argon laser"}, ExcludeKeywords: []string{}, CommonName: "Argon laser", CanonicalName: "Argon laser", Type: "Monitoring & Measurement Devices"},
{Keywords: []string{"bedside cabinet"}, ExcludeKeywords: []string{}, CommonName: "Bedside cabinet", CanonicalName: "Bedside cabinet", Type: "Support Equipment & Furniture"},
{Keywords: []string{"biosafety cabinet"}, ExcludeKeywords: []string{}, CommonName: "Biosafety cabinet", CanonicalName: "Biosafety cabinet", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"cassette printer"}, ExcludeKeywords: []string{}, CommonName: "Cassette printer", CanonicalName: "Cassette printer", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"cell counter"}, ExcludeKeywords: []string{}, CommonName: "Cell counter", CanonicalName: "Cell counter", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"defibrillator trolley"}, ExcludeKeywords: []string{}, CommonName: "Defibrillator trolley", CanonicalName: "Defibrillator trolley", Type: "Support Equipment & Furniture"},
{Keywords: []string{"dialysis chair"}, ExcludeKeywords: []string{}, CommonName: "Dialysis chair", CanonicalName: "Dialysis chair", Type: "Support Equipment & Furniture"},
{Keywords: []string{"hybridization incubator"}, ExcludeKeywords: []string{}, CommonName: "Hybridization incubator", CanonicalName: "Hybridization incubator", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"immunochemistry analyzer"}, ExcludeKeywords: []string{}, CommonName: "Immunochemistry analyzer", CanonicalName: "Immunochemistry analyzer", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"instrument cabinet"}, ExcludeKeywords: []string{}, CommonName: "Instrument cabinet", CanonicalName: "Instrument cabinet", Type: "Support Equipment & Furniture"},
{Keywords: []string{"instruments trolley"}, ExcludeKeywords: []string{}, CommonName: "Instruments trolley", CanonicalName: "Instruments trolley", Type: "Support Equipment & Furniture"},
{Keywords: []string{"intravenous line"}, ExcludeKeywords: []string{}, CommonName: "Intravenous line", CanonicalName: "Intravenous line", Type: "Consumables & Accessories"},
{Keywords: []string{"kick bucket"}, ExcludeKeywords: []string{}, CommonName: "Kick bucket", CanonicalName: "Kick bucket", Type: "Support Equipment & Furniture"},
{Keywords: []string{"medicine cabinet"}, ExcludeKeywords: []string{}, CommonName: "Medicine cabinet", CanonicalName: "Medicine cabinet", Type: "Support Equipment & Furniture"},
{Keywords: []string{"microscope cabinet"}, ExcludeKeywords: []string{}, CommonName: "Microscope cabinet", CanonicalName: "Microscope cabinet", Type: "Support Equipment & Furniture"},
{Keywords: []string{"storage cabinet"}, ExcludeKeywords: []string{}, CommonName: "Storage cabinet", CanonicalName: "Storage cabinet", Type: "Support Equipment & Furniture"},
{Keywords: []string{"surgical light"}, ExcludeKeywords: []string{}, CommonName: "Surgical light", CanonicalName: "Surgical light", Type: "Support Equipment & Furniture"},
{Keywords: []string{"funnel"}, ExcludeKeywords: []string{}, CommonName: "Funnel", CanonicalName: "Funnel", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"spectrophotometer"}, ExcludeKeywords: []string{}, CommonName: "Spectrophotometer", CanonicalName: "Spectrophotometer", Type: "Laboratory & IVD Equipment"},
{Keywords: []string{"cytoflowmeter", "cytoflowmeters"}, ExcludeKeywords: []string{}, CommonName: "Flow cytometer", CanonicalName: "Flow cytometer", Type: "Laboratory & IVD Equipment", Family: "Flow and cell analysis systems"},
{Keywords: []string{"blood flow meter", "blood flow meters"}, ExcludeKeywords: []string{}, CommonName: "Blood flow meter", CanonicalName: "Blood flow meter", Type: "Monitoring & Measurement Devices", Family: "Physiological assessment tools"},
{Keywords: []string{"nasal cannula", "nasal cannulas"}, ExcludeKeywords: []string{}, CommonName: "Oxygen nasal cannula", CanonicalName: "Oxygen nasal cannula", Type: "Medical Gas & Respiratory Devices", Family: "Respiratory masks and interfaces"},
{Keywords: []string{"clinical chemistry"}, ExcludeKeywords: []string{}, CommonName: "Chemistry analyzer", CanonicalName: "Clinical chemistry analyzer", Type: "Laboratory & IVD Equipment", Family: "Laboratory analyzers"},
{Keywords: []string{"anaesthesia system", "anesthesia system"}, ExcludeKeywords: []string{}, CommonName: "Anesthesia machine", CanonicalName: "Anesthesia workstation", Type: "Treatment, Surgical & Life Support Devices", Family: "Anesthesia and critical care therapy systems"},
{Keywords: []string{"ventilator"}, ExcludeKeywords: []string{"analyser", "analyzer", "sensor", "set", "circuit", "tubing", "accessory"}, CommonName: "Ventilator", CanonicalName: "Mechanical ventilator", Type: "Treatment, Surgical & Life Support Devices", Family: "Anesthesia and critical care therapy systems"},
{Keywords: []string{"blood glucose meter"}, ExcludeKeywords: []string{}, CommonName: "Blood glucose meter", CanonicalName: "Blood glucose monitoring device", Type: "Monitoring & Measurement Devices", Family: ""},
{Keywords: []string{"foetal monitor", "fetal monitor", "fhr monitor", "foetal heart rate"}, ExcludeKeywords: []string{}, CommonName: "Fetal monitor", CanonicalName: "Fetal monitoring system", Type: "Monitoring & Measurement Devices", Family: ""},
{Keywords: []string{"haemoglobin monitoring", "hemoglobin monitoring"}, ExcludeKeywords: []string{}, CommonName: "Hemoglobin monitor", CanonicalName: "Hemoglobin monitoring device", Type: "Monitoring & Measurement Devices", Family: ""},
{Keywords: []string{"oxygen concentrator"}, ExcludeKeywords: []string{}, CommonName: "Oxygen concentrator", CanonicalName: "Medical oxygen concentrator", Type: "Medical Gas & Respiratory Devices", Family: "Oxygen supply systems"},
{Keywords: []string{"thorpe tube", "medical medicinal gas pipeline systems", "medical gas pipeline", "oxygen flowmeter"}, ExcludeKeywords: []string{"cytoflowmeter", "blood flow meter", "nasal cannula"}, CommonName: "Oxygen flowmeter", CanonicalName: "Medical oxygen flowmeter", Type: "Medical Gas & Respiratory Devices", Family: "Oxygen supply systems"},
{Keywords: []string{"dialysis solution", "dialysate"}, ExcludeKeywords: []string{}, CommonName: "Dialysis solution", CanonicalName: "Dialysis or dialysate solution", Type: "Consumables & Accessories", Family: "Dialysis and extracorporeal consumables"},
{Keywords: []string{"catheter, suction", "catheter tip, suction", "yankauer"}, ExcludeKeywords: []string{}, CommonName: "Suction catheter", CanonicalName: "Suction catheter or Yankauer tip", Type: "Consumables & Accessories", Family: "Airway and laryngoscopy accessories"},
{Keywords: []string{"bottle, suction system", "bottle, vacuum, collection"}, ExcludeKeywords: []string{}, CommonName: "Suction bottle", CanonicalName: "Suction collection bottle", Type: "Medical Gas & Respiratory Devices", Family: "Suction and aspiration systems"},
{Keywords: []string{"bulb, suction"}, ExcludeKeywords: []string{}, CommonName: "Bulb suction device", CanonicalName: "Manual bulb suction device", Type: "Medical Gas & Respiratory Devices", Family: "Suction and aspiration systems"},
{Keywords: []string{"chair, haemodialysis", "chair, hemodialysis"}, ExcludeKeywords: []string{}, CommonName: "Dialysis chair", CanonicalName: "Hemodialysis treatment chair", Type: "Support Equipment & Furniture", Family: "Medical furniture and fixtures"},
{Keywords: []string{"central monitor, haemodialysis system", "central monitor, hemodialysis system"}, ExcludeKeywords: []string{}, CommonName: "Dialysis central monitor", CanonicalName: "Hemodialysis central monitoring system", Type: "Monitoring & Measurement Devices", Family: "Critical care monitoring systems"},
}

var TypeByCode = map[string]string{
"MONITORING_MEASUREMENT_DEVICES": "Monitoring & Measurement Devices",
"DIAGNOSTIC_IMAGING_DEVICES": "Diagnostic & Imaging Devices",
"TREATMENT_SURGICAL_LIFE_SUPPORT_DEVICES": "Treatment, Surgical & Life Support Devices",
"LABORATORY_IVD_EQUIPMENT": "Laboratory & IVD Equipment",
"MEDICAL_GAS_RESPIRATORY_DEVICES": "Medical Gas & Respiratory Devices",
"STERILIZATION_INFECTION_CONTROL_DEVICES": "Sterilization & Infection Control Devices",
"SUPPORT_EQUIPMENT_FURNITURE": "Support Equipment & Furniture",
"CONSUMABLES_ACCESSORIES": "Consumables & Accessories",
"CLINICAL_MONITORING_ASSESSMENT": "Monitoring & Measurement Devices",
"DIAGNOSTIC_IMAGING_VISUALIZATION": "Diagnostic & Imaging Devices",
"LABORATORY_IN_VITRO_DIAGNOSTICS": "Laboratory & IVD Equipment",
"THERAPEUTIC_LIFE_SUPPORT": "Treatment, Surgical & Life Support Devices",
"SURGICAL_INTERVENTIONAL": "Treatment, Surgical & Life Support Devices",
"REHABILITATION_MOBILITY_ASSISTIVE": "Support Equipment & Furniture",
"MEDICAL_GAS_RESPIRATORY_SUCTION": "Medical Gas & Respiratory Devices",
"INFECTION_PREVENTION_DECONTAMINATION_STERILIZATION": "Sterilization & Infection Control Devices",
"FACILITY_UTILITY_ENVIRONMENTAL_SUPPORT": "Support Equipment & Furniture",
"BIOMEDICAL_TEST_CALIBRATION_QUALITY_ASSURANCE": "Support Equipment & Furniture",
"DIGITAL_HEALTH_CLINICAL_SOFTWARE": "Support Equipment & Furniture",
"CONSUMABLES_ACCESSORIES_PROCEDURE_SUPPLIES": "Consumables & Accessories",
}

var FunctionByCode = map[string]string{
"LIFE_SUPPORT": "Life Support",
"SURGICAL_INTENSIVE_CARE": "Surgical and Intensive Care",
"PHYSICAL_THERAPY_TREATMENT": "Physical Therapy and Treatment",
"CRITICAL_CARE_MONITORING": "Surgical and Intensive Care Monitoring",
"GENERAL_PHYSIOLOGICAL_MONITORING": "Additional Physiological Monitoring and Diagnostic",
"ANALYTICAL_LABORATORY": "Analytical Laboratory",
"LABORATORY_ACCESSORIES": "Laboratory Accessories",
"COMPUTERS_AND_IT": "Computers and Related",
"PATIENT_RELATED_OTHER": "Patient Related and Other",
}

var DeviceTypeDefaults = map[string]struct{Function, Risk string}{
"Monitoring & Measurement Devices": {Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis"},
"Diagnostic & Imaging Devices": {Function: "Additional Physiological Monitoring and Diagnostic", Risk: "Inappropriate therapy or misdiagnosis"},
"Laboratory & IVD Equipment": {Function: "Analytical Laboratory", Risk: "Inappropriate therapy or misdiagnosis"},
"Treatment, Surgical & Life Support Devices": {Function: "Surgical and Intensive Care", Risk: "Potential patient or operator injury"},
"Medical Gas & Respiratory Devices": {Function: "Life Support", Risk: "Potential patient or operator injury"},
"Sterilization & Infection Control Devices": {Function: "Patient Related and Other", Risk: "Potential patient or operator injury"},
"Support Equipment & Furniture": {Function: "Patient Related and Other", Risk: "Equipment damage"},
"Consumables & Accessories": {Function: "Patient Related and Other", Risk: "Equipment damage"},
}
// Deprecated: use DeviceTypes
var OvaholDeviceTypes = DeviceTypes

// Deprecated: use DeviceTypeDefaults
var OvaholTypeDefaults = DeviceTypeDefaults
