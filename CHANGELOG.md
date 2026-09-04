# Changelog

All notable changes to `github.com/ovahol/ontology` will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed — ontology is now an engine; vendors bring their own taxonomy

Previously the *shape* of a taxonomy was still fixed by this library (8 device
types, 4 categories, 9 functions, 5 risks, plus `FamilyRule`/`SpecificNameRule`
Go structs with those exact field names) even after the vocabulary itself
moved into JSON in v0.2.0 — a second vendor (MeDevIS) had to reshape its own
domain into Ovahol's dimensions to participate. That's fixed now:

- **`Taxonomy` is just `Fields` + `Inference.Rules`.** `Fields` is a vendor's
  own list of classification dimensions (`FieldDef{Key, Label, Required,
  AllowedValues}`) — no fixed count, no special dimension names. `Rule`
  (`Keywords`, `SourceTypes`, `ExcludeKeywords`, `Requires`, `Set`, `Name`,
  `CanonicalName`) replaces `FamilyRule`/`SpecificNameRule`/`TypeDefaults`/
  `TypeByKeyword`/`SourceTypeMap`: one generic rule shape where `Set` assigns
  whatever field keys the vendor named, and `Requires` gates a rule on fields
  an earlier rule already resolved — which is how multi-stage inference
  (type → function/risk → category) is expressed, purely through rule order,
  with no such staging built into the engine.
- Removed `DeviceType`, `DeviceCategory`, `DeviceFunction`,
  `DeviceApplicationRisk`, `FamilyRule`, `SpecificNameRule` as fixed Taxonomy
  structs, and `TypeByCode`/`FunctionByCode` indirection (now baked directly
  into each taxonomy's rules at authoring time).
- New generic-engine test: `TestEngineWithCustomVendorRules` now exercises a
  vendor whose only dimension is `device_tier` (not `device_type` at all) to
  prove the engine doesn't special-case any field name — it previously only
  proved keyword rules could set an arbitrary *value* for the still-hardcoded
  `device_type` key.
- `Result.IsValid()` no longer checks `FieldDeviceType` by name; it checks
  whether any fields resolved at all, so it works for any taxonomy shape.
- `Engine` (`NewEngine`, `WithCatalog`) — a small object binding a taxonomy
  (and optional catalog) so callers don't have to thread `*Taxonomy` through
  every call.
- `DefaultTaxonomy()` — an embedded, neutral reference taxonomy derived from
  WHO's MeDevIS nomenclature, used only when no taxonomy is supplied
  (`Normalize`, `NormalizeWorkbook`, etc. without a taxonomy argument now mean
  "use the default," not "use Ovahol's vocabulary"). Ovahol's own vocabulary
  moved to `examples/taxonomies/ovahol.json`, used only by this repo's tests.
- Workbook: the old "Family Rules" sheet is now "Inference Rules" and dumps
  the vendor's actual rule list (whatever fields it sets) instead of assuming
  `Type`/`Family`/`Function`/`Risk` columns. "Lookups" now writes one column
  per taxonomy field with declared `allowed_values`, however many the vendor
  has, instead of 4 fixed columns with vendor-specific sub-columns (code,
  category, score) that don't generalize.
- Removed the `taxonomy` subpackage (a second, unused, competing `Taxonomy`
  type) and `cmd/ontology`'s dependency on it.

### Migration
- `Normalize`/`NormalizeBatch`/`NormalizeJSON`/`NormalizeWorkbook`/
  `NormalizeCSV` still work with no taxonomy argument (now backed by
  `DefaultTaxonomy()` instead of a vendored Ovahol vocabulary) — prefer the
  `*WithTaxonomy` / `Engine` variants and load your own taxonomy.
- If you were relying on Ovahol's specific vocabulary being built in, load
  it explicitly from `examples/taxonomies/ovahol.json` (or your own copy)
  instead.

### Corrected — the embedded MeDevIS default taxonomy now uses MeDevIS's own structure

The default taxonomy (embedded in `default.go`, source `devices.xlsx`) had been
forced into Ovahol's field shape: MeDevIS's real "Knowledge level" column
(Basic/General clinical/Specialized clinical) was mislabeled `device_function`,
and its "Reusable" column (Reusable/Single use) was mislabeled
`device_application_risk`. `examples/taxonomies/medevis.json` is regenerated
from the reference `devices.xlsx` (2653 rows) with MeDevIS's real dimensions:

- `device_type` (39) — required — the "Device type" column
- `service_type` (12) — the "Service type" column
- `knowledge_level` (3) — the "Knowledge level" column
- `reusable` (2) — the "Reusable" column
- `emdn_code` / `emdn_term` / `gmdn_code` / `gmdn_term` (nomenclature lookups)

It now ships 2649 exact-name inference rules (one per distinct device name,
longest first) that reconcile every known MeDevIS device to its exact tuple —
verified by `TestMeDevisDefaultReconciliation`, which reconciles all 2645
unique-name rows from the reference file. EMDN codes are deliberately not used
for classification (a single EMDN code maps to several distinct tuples). New
generator: `cmd/gen-medevis-taxonomy`.

## [0.2.0] - 2026-09-03

### Added
- Interchange schema: `Input`, `Result`, `InterchangeRecord`, `APIImportRecord` with JSON `snake_case` tags. `Result` carries a streamlined 4-field classification (`DeviceType`, `DeviceCategory`, `DeviceFunction`, `DeviceApplicationRisk`) plus the normalized `Name`; `DeviceApplication` is kept as an alias of `DeviceFunction` for backward-compat wording.
- High-level interchange API: `Normalize`, `NormalizeBatch`, `NormalizeJSON`, `NormalizeJSONFile`, `ToJSON`, `ToCSV`, `ToAPIImportRecords`, `ValidateInput`, `NormalizeCSV`.
- Catalog-backed exact-match resolution: `Catalog` interface, `InMemoryCatalog`, `CatalogEntry`, `NormalizeWithCatalog`, `NormalizeBatchWithCatalog` — lets a host system's existing device dictionary short-circuit taxonomy inference on exact hits (`MappingSource == "catalog_exact"`).
- Workbook interchange: `NormalizeWorkbook` (alias `UpdateOntology` for compat), `NormalizeCSV`.
- Standalone controlled vocabulary in `taxonomy.go` (8 device types, 4 device categories, 9 functions, 5 risks, ~140 family rules, 18 specific name rules) — no dependency on `ovahol` monorepo's `seed/lookup`. Family rules and canonical-name/alias inference remain internal (used by the workbook audit sheets) and are not exposed on `Result`.
- Confidence scoring (`high`/`medium`/`low`/`none`) and `Result.IsValid()`.

## [0.1.0] - 2026-09-02

### Added
- Initial standalone extraction from `github.com/ovahol/ovahol/backend/internal/ontology`.
- CLI `cmd/ontology`.
- Examples in `examples/`.

### Ported from
- `scripts/update_ovahol_ontology.py` (Python/openpyxl) — the taxonomy generator.
- `backend/internal/ontology/{taxonomy,infer,workbook}.go` — the Go runtime used during facility import.
