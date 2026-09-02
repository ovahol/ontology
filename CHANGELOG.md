# Changelog

All notable changes to `github.com/ovahol/ontology` will be documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial standalone extraction from `github.com/ovahol/ovahol/backend/internal/ontology`.
- Interchange schema: `Input`, `Result`, `InterchangeRecord`, `APIImportRecord` with JSON `snake_case` tags.
- High-level interchange API: `Normalize`, `NormalizeBatch`, `NormalizeJSON`, `NormalizeJSONFile`, `ToJSON`, `ToCSV`, `ToAPIImportRecords`, `ValidateInput`, `NormalizeCSV`.
- Workbook interchange: `NormalizeWorkbook` (alias `UpdateOntology` for compat), `NormalizeCSV`.
- Standalone controlled vocabulary in `taxonomy.go` (8 device types, 9 functions, 5 risks, ~120 family rules, 18 specific name rules) — no dependency on `ovahol` monorepo's `seed/lookup`.
- CLI `cmd/ontology`.
- Examples in `examples/`.
- Confidence scoring (`high`/`medium`/`low`/`none`) and `Result.IsValid()`.

### Ported from
- `scripts/update_ovahol_ontology.py` (Python/openpyxl) — the taxonomy generator.
- `backend/internal/ontology/{taxonomy,infer,workbook}.go` — the Go runtime used during facility import.
