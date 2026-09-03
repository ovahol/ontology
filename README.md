# ontology

**Ovahol Interchange Ontology** — a Go library that normalizes arbitrary device data from any healthcare system into Ovahol's controlled vocabulary.

Any system hoping to migrate onto Ovahol (or exchange data with it) runs its device records through this library and gets back canonical Ovahol terminology: normalized name, device type, device category, device function, and application risk.

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
fmt.Println(result.Name)                  // ECG machine
fmt.Println(result.DeviceType)            // Monitoring & Measurement Devices
fmt.Println(result.DeviceCategory)        // Diagnostic
fmt.Println(result.DeviceFunction)        // Additional Physiological Monitoring and Diagnostic
fmt.Println(result.DeviceApplicationRisk) // Inappropriate therapy or misdiagnosis
fmt.Println(result.Confidence)            // high
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
  "name": "Infusion pump",
  "device_type": "Treatment, Surgical & Life Support Devices",
  "device_category": "Therapeutic",
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

`device_application` is also present as an alias of `device_function`, kept for callers using the older "device application" wording.

All `device_type`, `device_category`, `device_function`, and `device_application_risk` values are drawn from a fixed controlled vocabulary — no free-text leakage.

## Controlled vocabulary

| Dimension | Count | Example values |
|-----------|-------|----------------|
| Device types | 8 | `Monitoring & Measurement Devices`, `Diagnostic & Imaging Devices`, `Treatment, Surgical & Life Support Devices`, ... |
| Device categories | 4 | `Therapeutic`, `Diagnostic`, `Analytical`, `Miscellaneous` |
| Device functions | 9 | `Life Support`, `Surgical and Intensive Care`, `Analytical Laboratory`, ... |
| Application risks | 5 | `Potential patient death`, `Potential patient or operator injury`, `Inappropriate therapy or misdiagnosis`, ... |

Internally the library also carries ~140 family grouping rules (used for name/family inference and in the workbook's audit sheets — see below) and 18 specific-name rules, but `Family` is not part of the public `Result`.

See [`taxonomy.go`](./taxonomy.go) for the full lists and [`doc.go`](./doc.go) for package documentation.

`MappingSource` explains how the common name was derived:

| Value | Meaning |
|-------|---------|
| `specific_rule` | Matched a high-priority keyword rule (e.g. "oxygen concentrator") |
| `legacy_derived` | Parsed directly from the legacy device name |
| `family_fallback` | Fell back to the family default name |
| `unsupported_source_type` | Source type not in `SupportedSourceTypes` — no mapping possible |

`Confidence` is `high` / `medium` / `low` / `none`.

## Catalog (exact-match resolution)

If the host system already has a device dictionary (e.g. Ovahol's `core_public.device`), pass it in as a `Catalog` so exact matches skip taxonomy inference entirely and return the dictionary's values verbatim:

```go
cat := ontology.NewInMemoryCatalog([]ontology.CatalogEntry{
    {
        Name:                  "ECG machine",
        DeviceType:            "Monitoring & Measurement Devices",
        DeviceCategory:        "Diagnostic",
        DeviceFunction:        "Additional Physiological Monitoring and Diagnostic",
        DeviceApplicationRisk: "Inappropriate therapy or misdiagnosis",
    },
})

result := ontology.NormalizeWithCatalog(ontology.Input{DeviceName: "ECG machine"}, cat)
// result.MappingSource == "catalog_exact", result.Confidence == "high"
```

Matching is by EMDN code, then device name, then EMDN term (all normalized/lowercased). A catalog miss falls back to ordinary taxonomy inference; a `nil` catalog always falls back. Implement `Catalog` yourself (e.g. backed by a single SQL lookup) to avoid vendoring the whole dictionary — see [`catalog.go`](./catalog.go) for the interface and a DB-backed example. `NormalizeBatchWithCatalog` is the batch analogue.

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
| Devices | One row per device: name, type, category, function, risk, plus legacy source passthrough |
| API Import | Deduplicated, API-ready rows (`name`, `device_type`, `device_category`, `device_function`, `device_application_risk`, `emdn_code`, `emdn_term`) |
| Lookups | The 8 types, 4 categories, 9 functions, 5 risks (for Excel data validation) |
| Naming Rules | How names should be formed |
| Family Rules | All ~140 internal family grouping rules with match hints (used for inference, not exposed on `Result`) |
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
