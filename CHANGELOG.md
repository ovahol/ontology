# Changelog

All notable changes to `github.com/ovahol/ontology` will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial standalone extraction from `github.com/ovahol/ovahol/backend/internal/ontology`.
- Interchange schema: `Input`, `Result`, `InterchangeRecord`, `APIImportRecord` with JSON `snake_case` tags. `Result` carries a streamlined 4-field classification (`DeviceType`, `DeviceCategory`, `DeviceFunction`, `DeviceApplicationRisk`) plus the normalized `Name`; `DeviceApplication` is kept as an alias of `DeviceFunction` for backward-compat wording.
- High-level interchange API: `Normalize`, `NormalizeBatch`, `NormalizeJSON`, `NormalizeJSONFile`, `ToJSON`, `ToCSV`, `ToAPIImportRecords`, `ValidateInput`, `NormalizeCSV`.
- Catalog-backed exact-match resolution: `Catalog` interface, `InMemoryCatalog`, `CatalogEntry`, `NormalizeWithCatalog`, `NormalizeBatchWithCatalog` — lets a host system's existing device dictionary short-circuit taxonomy inference on exact hits (`MappingSource == "catalog_exact"`).
- Workbook interchange: `NormalizeWorkbook` (alias `UpdateOntology` for compat), `NormalizeCSV`.
- Standalone controlled vocabulary in `taxonomy.go` (8 device types, 4 device categories, 9 functions, 5 risks, ~140 family rules, 18 specific name rules) — no dependency on `ovahol` monorepo's `seed/lookup`. Family rules and canonical-name/alias inference remain internal (used by the workbook audit sheets) and are not exposed on `Result`.
- CLI `cmd/ontology`.
- Examples in `examples/`.
- Confidence scoring (`high`/`medium`/`low`/`none`) and `Result.IsValid()`.

### Ported from
- `scripts/update_ovahol_ontology.py` (Python/openpyxl) — the taxonomy generator.
- `backend/internal/ontology/{taxonomy,infer,workbook}.go` — the Go runtime used during facility import.
