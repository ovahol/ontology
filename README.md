# ontology

**ontology** is a Go engine for normalizing free-text device records into a controlled vocabulary. It ships no vocabulary of its own — every vendor (Ovahol, or anyone else) defines their own **Taxonomy**: the classification dimensions they care about and the rules that infer them from a device name, source type, and EMDN code/term. ontology just executes it.

That's the whole design: **engine + data**. The taxonomy is JSON, authored and versioned in the vendor's own codebase, not vendored into this repo.

## Installation

```bash
go get github.com/ovahol/ontology@latest
```

Requires Go 1.27.0 or later.

## Quick start

```go
import "github.com/ovahol/ontology"

tax, err := ontology.LoadTaxonomyFile("my_taxonomy.json")

result := ontology.NormalizeWithTaxonomy(ontology.Input{
    DeviceName: "ECG machine, portable, 12-lead",
    SourceType: "monitoring equipment",
    EMDNTerm:   "Electrocardiographs",
}, tax)

fmt.Println(result.Name)   // ECG machine
fmt.Println(result.Fields) // map[device_type:... device_category:... device_function:... device_application_risk:...]
```

Or bind a taxonomy (and optionally a catalog) to an `Engine` once, and reuse it:

```go
engine := ontology.NewEngine(tax, ontology.WithCatalog(myCatalog))
result := engine.Normalize(input)
results := engine.NormalizeBatch(inputs)
```

The `Engine` also reconciles identity — the device name, model, brand,
manufacturer, and a controlled status — using only vendor-supplied conventions
and an optional FK resolver (see [identity reconciliation](#identity-reconciliation)):

```go
engine := ontology.NewEngine(tax,
    ontology.WithCatalog(myCatalog),
    ontology.WithConventions(myConventions),       // Unknown placeholders + status set
    ontology.WithIdentityResolver(myResolver),     // walks model->brand->manufacturer
)
id := engine.Reconcile(ontology.IdentityInput{DeviceName: result.Name, Model: dm.Model, Status: "Active"})
```

If you don't have a taxonomy yet, `ontology.Normalize(input)` (no taxonomy argument) falls back to an embedded neutral reference taxonomy derived from WHO's MeDevIS nomenclature — useful for getting started, not a substitute for your own vocabulary.

## Taxonomy: your vocabulary, your rules

A `Taxonomy` is two things, and nothing else:

- **`Fields`** — the classification dimensions you want out. Each field has a `key`, optional `allowed_values` (its controlled vocabulary), and whether it's `required`. There's no fixed set of dimensions and no dimension name means anything special to the engine — `device_type` is just a convention, not a keyword ontology recognizes.
- **`Inference.Rules`** — an ordered list of rules. Each rule says: *if the input matches this (keywords, source type, or fields already resolved by an earlier rule), set these field values.* The first rule to set a given field wins; later rules only fill in what's still unset. That's how multi-stage inference works — no engine-level concept of "type then function then category," just rule order.

```json
{
  "id": "acme-devices",
  "version": "1.0.0",
  "fields": [
    { "key": "device_type", "required": true, "allowed_values": ["Imaging", "Monitoring", "Surgical"] },
    { "key": "risk_tier", "allowed_values": ["Low", "Medium", "High"] }
  ],
  "inference": {
    "rules": [
      { "keywords": ["ultrasound", "x-ray", "mri"], "set": { "device_type": "Imaging" } },
      { "keywords": ["ecg", "patient monitor"], "set": { "device_type": "Monitoring" } },
      { "requires": { "device_type": "Imaging" }, "set": { "risk_tier": "Medium" } }
    ]
  }
}
```

Load it with `ontology.LoadTaxonomyFile("acme.json")`. A vendor with two dimensions declares two; a vendor with eight declares eight — nothing here is Ovahol-shaped or MeDevIS-shaped. `Taxonomy.Validate()` checks it's well-formed (an id, a semver version, at least one required field with its vocabulary spelled out).

Rules also support `source_types` (match against the normalized source type instead of/alongside keywords), `exclude_keywords`, and `name` / `canonical_name` (to assign the device's display name directly instead of deriving it from the legacy text). See [`taxonomy_version.go`](./taxonomy_version.go) for the full `Rule`/`FieldDef` shapes, and [`examples/taxonomies/ovahol.json`](./examples/taxonomies/ovahol.json) for a taxonomy with ~290 rules across 5 dimensions as a fully worked example.

If none of a field's rules fire but the taxonomy declares `allowed_values` for it, the engine tries one more thing for free: matching the normalized source type directly against those allowed values. This is why the MeDevIS default taxonomy also classifies anything whose source type happens to already be one of its 39 device type names — even though its exact-name rules already reconcile every known device.

## Result shape

```go
result.Name                  // normalized display name, e.g. "ECG machine"
result.Fields                // map[string]string — every field the taxonomy resolved, keyed by field key
result.MappingSource         // "specific_rule" | "legacy_derived" | "family_fallback" | "unsupported_source_type" | "catalog_exact"
result.Confidence            // "high" | "medium" | "low" | "none"
```

`Result.Fields` is the **sole** storage for taxonomy dimensions — a `Result` carries no fixed dimension fields. Read `result.Fields["device_type"]` or `result.GetField("device_type")`; a taxonomy with different keys (like `risk_tier`) works the same way.

```json
{
  "name": "Infusion pump",
  "fields": {
    "device_type": "Treatment, Surgical & Life Support Devices",
    "device_category": "Therapeutic",
    "device_function": "Surgical and Intensive Care",
    "device_application_risk": "Potential patient or operator injury"
  },
  "legacy_source_name": "Infusion pump, volumetric",
  "source_type": "infusion devices",
  "mapping_source": "family_fallback",
  "confidence": "medium"
}
```

## Catalog (exact-match resolution)

If the host system already has a device dictionary (e.g. Ovahol's `core_public.device`), pass it in as a `Catalog` so exact matches skip rule inference entirely and return the dictionary's values verbatim:

```go
cat := ontology.NewInMemoryCatalog([]ontology.CatalogEntry{
    {
        Name: "ECG machine",
        Fields: map[string]string{
            "device_type":             "Monitoring & Measurement Devices",
            "device_category":         "Diagnostic",
            "device_function":         "Additional Physiological Monitoring and Diagnostic",
            "device_application_risk": "Inappropriate therapy or misdiagnosis",
        },
    },
})

result := ontology.NormalizeWithCatalogAndTaxonomy(ontology.Input{DeviceName: "ECG machine"}, cat, tax)
// result.MappingSource == "catalog_exact", result.Confidence == "high"
```

Matching is by EMDN code, then device name, then EMDN term (all normalized/lowercased). A catalog miss falls back to ordinary taxonomy inference; a `nil` catalog always falls back. Implement `Catalog` yourself (e.g. backed by a single SQL lookup) to avoid vendoring the whole dictionary — see [`catalog.go`](./catalog.go) for the interface and a DB-backed example. `NormalizeBatchWithCatalogAndTaxonomy` is the batch analogue, and `Engine`/`WithCatalog` wraps both together.

## Identity reconciliation

Classification tells you what a device *is*; reconciliation tells you how it
*exists* in your system — its device name, model, brand, manufacturer, and
status. The engine completes identity the same way it classifies: it only
executes what you configure, so no identity vocabulary is hardcoded here.

Two vendor-supplied pieces bind to `Engine.Reconcile`:

- **`Conventions`** — your "Unknown" placeholder templates and your controlled
  status vocabulary:

  ```go
  conv := ontology.Conventions{
      UnknownDevice:       "Unknown Device",
      UnknownBrand:        "Unknown Brand",
      UnknownManufacturer: "Unknown Manufacturer",
      UnknownModelPrefix:  "Unknown Model - ", // per-device unknown model, keeps FK resolvable
      Statuses: []string{
          "In-Service", "Decommissioned", "Transferred", "Standby / Spare",
          "Under Maintenance", "Out of Service", "Disposed", "New / Commissioning",
      },
      // Source inventories use many free-text terms for a status; each maps
      // (case-insensitively) to Ovahol's canonical status. Anything unmatched
      // falls back to DefaultStatus.
      StatusSynonyms: map[string][]string{
          "In-Service":        {"functional", "functioning", "active", "in active service", "working"},
          "Under Maintenance": {"faulty", "not working", "broken down", "broken down and repairable"},
          "Out of Service":    {"down", "offline"},
          "Standby / Spare":   {"standby", "spare", "reserve"},
          "Disposed":          {"discarded", "scrapped"},
      },
      DefaultStatus: "New / Commissioning", // conservative default on ingest
  }
  ```

- **`IdentityResolver`** — an optional interface that maps an inbound identity
  tuple to the canonical entity names via your own reference-data foreign keys
  (e.g. model -> brand -> manufacturer); return empty strings for anything you
  can't resolve and the engine fills the Unknown placeholders.

```go
engine := ontology.NewEngine(tax,
    ontology.WithCatalog(myCatalog),
    ontology.WithConventions(conv),
    ontology.WithIdentityResolver(myResolver),
)
result := engine.Normalize(ontology.Input{DeviceName: "ECG machine"})
id := engine.Reconcile(ontology.IdentityInput{
    DeviceName:   result.Name, // the classified name
    Model:        src.Model,   Brand: src.Brand, Manufacturer: src.Manufacturer,
    Status:       src.Status,
})
// id.Device / id.Model / id.Brand / id.Manufacturer / id.Status all non-empty
// where Applies; status is normalized into Statuses (case-insensitively, via
// StatusSynonyms) or DefaultStatus on a miss.
```

`ReconcileIdentity(in, res, conv)` is the resolver-free, package-level
counterpart. A zero-value `Conventions` invents no strings and normalizes no
status, so the mechanism is fully opt-in.

## Workbook interchange (bulk migration)

For migrating a full inventory spreadsheet, use the workbook API. Any spreadsheet with at least a device name column can be normalized:

```go
// Excel → normalized Excel + API CSV
csvPath, err := ontology.NormalizeWorkbookWithTaxonomy("legacy_inventory.xlsx", "normalized.xlsx", tax)
fmt.Println(csvPath) // <output>.api_import.csv, next to the .xlsx
```

If your host system has a device dictionary, reconcile the workbook against it
before falling back to rules — known rows resolve verbatim, unknown rows still
classify by taxonomy:

```go
csvPath, err := ontology.NormalizeWorkbookWithCatalogAndTaxonomy("inventory.xlsx", "normalized.xlsx", myCatalog, tax)
```

Anything the taxonomy doesn't classify — Model, Manufacturer, Serial number, or
any other input column — is passed through to the output untouched; the engine
never drops caller data it doesn't recognize.

The output workbook mirrors the reference template:

| Sheet | Contents |
|-------|----------|
| *(first sheet, original name preserved)* | One row per device: name, one column per `tax.Fields` field the taxonomy declares (any dimensions), plus legacy source passthrough |
| API Import | Deduplicated, API-ready rows |
| Lookups | One column per taxonomy field that declares `allowed_values` — however many dimensions the taxonomy has (used for Excel dropdown validation on those columns) |
| Naming Rules | `Taxonomy.NamingRules`, if the vendor supplies any |
| Inference Rules | The vendor's `Inference.Rules` list verbatim (keywords/source types/requires → sets), for audit |
| Common Name Mapping Review | Legacy → mapped name audit trail |
| Family Naming Review / Audit | Per-`device_family` consistency check, when the taxonomy uses that conventional key |

Column headers are flexible on input — the library tolerates `Device name`, `name`, `common_name`, `device_type`, `emdn_code`, `Nomenclature code (EMDN)`, etc.

`NormalizeWorkbook`/`NormalizeCSV` (no taxonomy argument) fall back to the embedded default taxonomy, same as `Normalize`.

## CLI

```bash
go install github.com/ovahol/ontology/cmd/ontology@latest

ontology --taxonomy my_taxonomy.json legacy.xlsx normalized.xlsx
# normalized workbook written to normalized.xlsx
# api import csv written to normalized.xlsx.api_import.csv
```

Omit `--taxonomy`/`--taxonomy-dir` to use the embedded WHO/MeDevIS default.

## What's not in here

This library normalizes **vocabulary** — it does not:

- Own any vendor's vocabulary. Ovahol's, MeDevIS's, and any other taxonomy under `examples/taxonomies/` are examples/fixtures for this repo's own tests, not vocabulary this library ships as *the* vocabulary.
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
