# ontology

**Ovahol Interchange Ontology** — a Go library that normalizes arbitrary device data from any healthcare system into Ovahol's controlled vocabulary.

Any system hoping to migrate onto Ovahol (or exchange data with it) runs its device records through this library and gets back canonical Ovahol terminology: common name, canonical device name, search aliases, device type, device family, device function, and application risk.

This is the interchange schema: one vocabulary, any source.

## Installation

```bash
go get github.com/ovahol/ontology@latest
```

Requires Go 1.27.0 or later.

## Quick start

```go
import "github.com/ovahol/ontology"

result := ontology.Normalize(ontology.Input{
    DeviceName: "ECG machine, portable, 12-lead",
    SourceType: "monitoring equipment",
    EMDNTerm:   "Electrocardiographs",
})
fmt.Println(result.CommonName)    // ECG machine
fmt.Println(result.CanonicalName) // Electrocardiography system
fmt.Println(result.OvaholType)    // Monitoring & Measurement Devices
fmt.Println(result.Family)        // Cardiac diagnostic systems
fmt.Println(result.Function)      // Additional Physiological Monitoring and Diagnostic
fmt.Println(result.Risk)          // Inappropriate therapy or misdiagnosis
fmt.Println(result.SearchAliases) // ECG machine, EKG machine, Electrocardiograph
fmt.Println(result.Confidence)    // high
```

Batch:

```go
results := ontology.NormalizeBatch([]ontology.Input{
    {DeviceName: "Infusion pump, volumetric", SourceType: "infusion devices"},
    {DeviceName: "Catheter, sterile, single-use, adult", SourceType: "catheters and related"},
})

apiRecords := ontology.ToAPIImportRecords(results) // deduplicated, API-ready
csvBytes, _ := ontology.ToCSV(results)
jsonBytes, _ := ontology.ToJSON(results)
```

## Interchange schema

The library defines a language-agnostic JSON interchange schema so non-Go systems can participate by producing/consuming JSON:

**Input** (what any system sends):

```json
{
  "device_name": "Infusion pump, volumetric",
  "source_type": "infusion devices",
  "emdn_code": "Z12010501",
  "emdn_term": "Volumetric infusion pumps"
}
```

**Result** (what Ovahol expects):

```json
{
  "common_name": "Infusion pump",
  "canonical_name": "Infusion or syringe pump system",
  "search_aliases": "Infusion pump, Infusion or syringe pump system, IV pump",
  "ovahol_device_type": "Treatment, Surgical & Life Support Devices",
  "ovahol_device_family": "Infusion and medication delivery systems",
  "device_function": "Surgical and Intensive Care",
  "device_application_risk": "Potential patient or operator injury",
  "legacy_source_name": "Infusion pump, volumetric",
  "source_type": "infusion devices",
  "emdn_code": "Z12010501",
  "emdn_term": "Volumetric infusion pumps",
  "mapping_source": "family_fallback",
  "confidence": "medium"
}
```

All `ovahol_*`, `device_function`, and `device_application_risk` values are drawn from a fixed controlled vocabulary — no free-text leakage.

## Controlled vocabulary

| Dimension | Count | Example values |
|-----------|-------|----------------|
| Device types | 8 | `Monitoring & Measurement Devices`, `Diagnostic & Imaging Devices`, `Treatment, Surgical & Life Support Devices`, ... |
| Device families | ~120 | `Cardiac diagnostic systems`, `Ultrasound systems`, `Infusion and medication delivery systems`, ... |
| Device functions | 9 | `Life Support`, `Surgical and Intensive Care`, `Analytical Laboratory`, ... |
| Application risks | 5 | `Potential patient death`, `Potential patient or operator injury`, `Inappropriate therapy or misdiagnosis`, ... |

See [`taxonomy.go`](./taxonomy.go) for the full lists and [`doc.go`](./doc.go) for package documentation.

`MappingSource` explains how the common name was derived:

| Value | Meaning |
|-------|---------|
| `specific_rule` | Matched a high-priority keyword rule (e.g. "oxygen concentrator") |
| `legacy_derived` | Parsed directly from the legacy device name |
| `family_fallback` | Fell back to the family default name |
| `unsupported_source_type` | Source type not in `SupportedSourceTypes` — no mapping possible |

`Confidence` is `high` / `medium` / `low` / `none`.

## Workbook interchange (bulk migration)

For migrating a full inventory spreadsheet, use the workbook API. Any spreadsheet with at least a device name column can be normalized:

```go
// Excel → normalized Excel + API CSV
csvPath, err := ontology.NormalizeWorkbook("legacy_inventory.xlsx", "normalized.xlsx")
fmt.Println(csvPath) // normalized.csv (next to the .xlsx)

// Or at the record level
records := ontology.NormalizeBatch(inputs)
csvBytes, _ := ontology.ToCSV(records)
apiRecords := ontology.ToAPIImportRecords(records)
```

The output workbook mirrors the reference template:

| Sheet | Contents |
|-------|----------|
| Devices | One row per device: common name, canonical name, aliases, type, family, function, risk, plus legacy source passthrough |
| API Import | Deduplicated, API-ready rows (`name`, `device_type`, `device_function`, `device_application_risk`, `emdn_code`, `emdn_term`) |
| Lookups | The 8 types, 9 functions, 5 risks (for Excel data validation) |
| Naming Rules | How common/canonical/alias names should be formed |
| Family Rules | All ~120 family grouping rules with match hints |
| Common Name Mapping Review | Legacy → mapped name audit trail |
| Family Naming Review | Per-family consistency check |
| Family Naming Audit | Detailed per-family audit |

Column headers are flexible on input — the library tolerates `Device name`, `name`, `common_name`, `device_type`, `emdn_code`, `Nomenclature code (EMDN)`, etc.

## CLI

```bash
go install github.com/ovahol/ontology/cmd/ontology@latest

ontology legacy.xlsx normalized.xlsx
# normalized workbook written to normalized.xlsx
# api import csv written to normalized.csv
```

## What's not in here

This library normalizes **vocabulary** — it does not:

- Resolve foreign keys (model IDs, location IDs, status IDs) — those are facility-specific and resolved by the importer after normalization.
- Validate business rules (e.g. "is this device allowed at this facility?") — that's the application's job.
- Handle non-device entities (work orders, training sessions, etc.) — use `facilityimport` in the Ovahol monorepo for the full facility lifecycle import.

Per the interchange philosophy, those are composition concerns for the application that consumes this library.

## Development

```bash
go build ./...
go test ./...
go vet ./...
```

## License

[MIT](./LICENSE)
