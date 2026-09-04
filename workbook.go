package ontology

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xuri/excelize/v2"
)

// styleHeader applies the header style (1F4B99, white bold, centered).
func styleHeader(f *excelize.File, sheet, cell string) error {
	style, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"1F4B99"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "D9D9D9", Style: 1},
			{Type: "top", Color: "D9D9D9", Style: 1},
			{Type: "bottom", Color: "D9D9D9", Style: 1},
			{Type: "right", Color: "D9D9D9", Style: 1},
		},
	})
	if err != nil {
		return err
	}
	return f.SetCellStyle(sheet, cell, cell, style)
}

func setCell(f *excelize.File, sheet string, row, col int, value string) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	f.SetCellValue(sheet, cell, value)
}

// extractDeviceRows reads the first sheet's header row and returns row maps.
// It mimics Python's extract_device_rows which tolerates alternative header names.
func extractDeviceRows(f *excelize.File, sheetName string) ([]map[string]string, []string, error) {
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("sheet %s is empty", sheetName)
	}
	headers := rows[0]
	headerMap := map[string]int{}
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}
	// EMDN code/term may have prefix like "Nomenclature code (EMDN)"
	emdnCodeKey := "EMDN code"
	emdnTermKey := "EMDN term"
	for _, h := range headers {
		if strings.HasPrefix(h, "Nomenclature code (EMDN)") {
			emdnCodeKey = h
		}
		if strings.HasPrefix(h, "Nomenclature term (EMDN)") {
			emdnTermKey = h
		}
	}
	find := func(keys ...string) int {
		for _, k := range keys {
			if idx, ok := headerMap[k]; ok {
				return idx
			}
		}
		return -1
	}
	var result []map[string]string
	for _, row := range rows[1:] {
		get := func(keys ...string) string {
			idx := find(keys...)
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		// Also allow "Device name" / "name" etc. — support both new agnostic and legacy Ovahol headers for backward compat
		m := map[string]string{
			"Name":                    get("Name", "Common name", "common_name"),
			"Canonical device name":   get("Canonical device name", "canonical_device_name"),
			"Common names":            get("Common names", "Search aliases", "search_aliases", "common_names"),
			"Device type":             get("Device type", "device_type", "Ovahol device type", "ovahol_device_type"),
			"Device family":           get("Device family", "device_family", "Ovahol device family", "ovahol_device_family"),
			"Device function":         get("Device function", "device_function"),
			"Device application risk": get("Device application risk", "device_application_risk"),
			"Legacy source name":      get("Legacy source name", "Device name", "name", "legacy_source_name"),
			"Source device type":      get("Source device type", "source_type"),
			"EMDN code":               get("EMDN code", "emdn_code", emdnCodeKey),
			"EMDN term":               get("EMDN term", "emdn_term", emdnTermKey),
		}
		// skip entirely empty rows
		empty := true
		for _, v := range m {
			if v != "" {
				empty = false
				break
			}
		}
		if empty {
			continue
		}
		result = append(result, m)
	}
	return result, headers, nil
}

func colLetter(n int) string {
	letter, _ := excelize.ColumnNumberToName(n)
	return letter
}

func rebuildDevicesSheet(f *excelize.File, sheetName string, rows []map[string]string, tax *Taxonomy) (string, error) {
	// Delete and recreate as first sheet
	idx, err := f.GetSheetIndex(sheetName)
	if err == nil {
		f.DeleteSheet(sheetName)
		_ = idx
	}
	// create will be at end, then move to 0
	newIdx, err := f.NewSheet(sheetName)
	if err != nil {
		return "", err
	}
	f.SetActiveSheet(newIdx)
	// Move to first position by reordering via SetSheetVisible trick? excelize MoveSheet
	if err := f.MoveSheet(sheetName, "Sheet1"); err != nil {
		// fallback: try moving to front via index - excelize MoveSheet may not exist in this version, use SetSheetProps?
		// If MoveSheet not available, we just keep creation order; not critical.
	}

	for colIdx, header := range DefaultDeviceSheetHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, header)
		styleHeader(f, sheetName, cell)
	}
	for rowIdx, row := range rows {
		resolved := ResolveRowNamingFor(row, tax)
		values := []string{
			resolved.Name,
			resolved.DeviceType,
			resolved.DeviceCategory,
			resolved.DeviceFunction,
			resolved.DeviceApplicationRisk,
			row["Legacy source name"],
			row["Source device type"],
			row["EMDN code"],
			row["EMDN term"],
		}
		for colIdx, v := range values {
			setCell(f, sheetName, rowIdx+2, colIdx+1, v)
		}
	}
	// freeze, filter, widths
	f.SetPanes(sheetName, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	maxRow := len(rows) + 1
	if maxRow >= 2 {
		f.AutoFilter(sheetName, fmt.Sprintf("A1:I%d", maxRow), nil)
	}
	widths := map[string]float64{"A": 28, "B": 34, "C": 24, "D": 34, "E": 34, "F": 48, "G": 34, "H": 18, "I": 44}
	for col, w := range widths {
		f.SetColWidth(sheetName, col, col, w)
	}
	f.SetRowHeight(sheetName, 1, 18)
	return sheetName, nil
}

func rebuildAPIImportSheet(f *excelize.File, devicesSheet string, tax *Taxonomy) error {
	sheet := "API Import"
	// delete if exists
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	for colIdx, header := range DefaultAPIImportHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, header)
		styleHeader(f, sheet, cell)
	}
	// read devices sheet rows
	rows, err := f.GetRows(devicesSheet)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return nil
	}
	headerMap := map[string]int{}
	for i, h := range rows[0] {
		headerMap[h] = i
	}
	seen := map[string]bool{}
	outRow := 2
	for _, row := range rows[1:] {
		get := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		getWithFallback := func(keys ...string) string {
			for _, k := range keys {
				if v := get(k); v != "" {
					return v
				}
			}
			return ""
		}
		values := []string{
			getWithFallback("Name", "Common name"),
			getWithFallback("Device type", "Ovahol device type"),
			getWithFallback("Device category"),
			getWithFallback("Device function"),
			getWithFallback("Device application risk"),
			getWithFallback("EMDN code"),
			getWithFallback("EMDN term"),
		}
		// skip if any of first 5 empty (requires name + 4 taxonomy fields)
		if values[0] == "" || values[1] == "" || values[2] == "" || values[3] == "" || values[4] == "" {
			continue
		}
		// Generalized dedup: use APIImportRecord.DedupKey() so pluggable Fields participate automatically.
		// For workbook rows we synthesize an APIImportRecord and delegate key generation.
		rec := APIImportRecord{
			Name:                  values[0],
			DeviceType:            values[1],
			DeviceCategory:        values[2],
			DeviceFunction:        values[3],
			DeviceApplicationRisk: values[4],
			EMDNCode:              values[5],
			EMDNTerm:              values[6],
			Fields: map[string]string{
				FieldDeviceType:            values[1],
				FieldDeviceCategory:        values[2],
				FieldDeviceFunction:        values[3],
				FieldDeviceApplicationRisk: values[4],
			},
		}
		key := rec.DedupKey()
		if seen[key] {
			continue
		}
		seen[key] = true
		for colIdx, v := range values {
			setCell(f, sheet, outRow, colIdx+1, v)
		}
		outRow++
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	if outRow > 2 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:G%d", outRow-1), nil)
	}
	widths := map[string]float64{"A": 34, "B": 24, "C": 34, "D": 34, "E": 18, "F": 42, "G": 42}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
	return nil
}

// rebuildLookupsSheet writes one column per taxonomy field that declares
// AllowedValues, in Fields order — however many dimensions the vendor's
// taxonomy has. It returns the column letter used for each field key so
// applyValidations can wire up dropdowns for the fields the Devices sheet
// has dedicated columns for.
func rebuildLookupsSheet(f *excelize.File, tax *Taxonomy) (map[string]string, error) {
	sheet := "Lookups"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return nil, err
	}
	_ = newIdx
	colFor := make(map[string]string)
	col := 1
	for _, fd := range tax.Fields {
		if len(fd.AllowedValues) == 0 {
			continue
		}
		letter := colLetter(col)
		header := fd.Label
		if header == "" {
			header = fd.Key
		}
		f.SetCellValue(sheet, letter+"1", header)
		styleHeader(f, sheet, letter+"1")
		for i, v := range fd.AllowedValues {
			f.SetCellValue(sheet, fmt.Sprintf("%s%d", letter, i+2), v)
		}
		f.SetColWidth(sheet, letter, letter, 44)
		colFor[fd.Key] = letter
		col++
	}
	return colFor, nil
}

func rebuildNamingRulesSheet(f *excelize.File, tax *Taxonomy) error {
	sheet := "Naming Rules"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	headers := []string{"Field", "Rule", "Good example", "Avoid"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		styleHeader(f, sheet, cell)
	}
	for i, r := range tax.NamingRules {
		setCell(f, sheet, i+2, 1, r.Field)
		setCell(f, sheet, i+2, 2, r.Rule)
		setCell(f, sheet, i+2, 3, r.GoodExample)
		setCell(f, sheet, i+2, 4, r.Avoid)
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	if len(tax.NamingRules) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:D%d", len(tax.NamingRules)+1), nil)
	}
	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 72)
	f.SetColWidth(sheet, "C", "C", 34)
	f.SetColWidth(sheet, "D", "D", 42)
	return nil
}

// rebuildInferenceRulesSheet dumps the vendor's ordered rule list verbatim —
// Keywords/SourceTypes/Requires (when the rule matches) and Set/Name (what it
// assigns) — as a debug/audit view. Unlike the old Ovahol-specific "Family
// Rules" sheet, this makes no assumption about which fields a rule sets.
func rebuildInferenceRulesSheet(f *excelize.File, tax *Taxonomy) error {
	sheet := "Inference Rules"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	headers := []string{"When: keywords", "When: source types", "When: requires", "Sets", "Name", "Canonical name"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		styleHeader(f, sheet, cell)
	}
	var rules []Rule
	if tax.Inference != nil {
		rules = tax.Inference.Rules
	}
	for i, r := range rules {
		requires := make([]string, 0, len(r.Requires))
		for k, v := range r.Requires {
			requires = append(requires, k+"="+v)
		}
		sort.Strings(requires)
		sets := make([]string, 0, len(r.Set))
		for k, v := range r.Set {
			sets = append(sets, k+"="+v)
		}
		sort.Strings(sets)
		values := []string{
			strings.Join(r.Keywords, ", "),
			strings.Join(r.SourceTypes, ", "),
			strings.Join(requires, ", "),
			strings.Join(sets, ", "),
			r.Name,
			r.CanonicalName,
		}
		for colIdx, v := range values {
			setCell(f, sheet, i+2, colIdx+1, v)
		}
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, fmt.Sprintf("A1:F%d", len(rules)+1), nil)
	f.SetColWidth(sheet, "A", "A", 60)
	f.SetColWidth(sheet, "B", "B", 34)
	f.SetColWidth(sheet, "C", "C", 34)
	f.SetColWidth(sheet, "D", "D", 60)
	f.SetColWidth(sheet, "E", "E", 34)
	f.SetColWidth(sheet, "F", "F", 42)
	return nil
}

func rebuildCommonNameMappingReviewSheet(f *excelize.File, devicesSheet string, tax *Taxonomy) error {
	sheet := "Common Name Mapping Review"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	headers := []string{"Legacy source name", "Name", "Device type", "Device category", "Device function", "Device application risk", "EMDN code", "EMDN term", "Mapping source"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		styleHeader(f, sheet, cell)
	}
	rows, err := f.GetRows(devicesSheet)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return nil
	}
	headerMap := map[string]int{}
	for i, h := range rows[0] {
		headerMap[h] = i
	}
	outRow := 2
	for _, row := range rows[1:] {
		get := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return row[idx]
			}
			return ""
		}
		// Re-resolve to get naming_source
		rmap := map[string]string{
			"Legacy source name": get("Legacy source name"),
			"Source device type": get("Source device type"),
			"EMDN term":          get("EMDN term"),
		}
		resolved := ResolveRowNamingFor(rmap, tax)
		values := []string{
			get("Legacy source name"),
			get("Name"),
			get("Device type"),
			get("Device category"),
			get("Device function"),
			get("Device application risk"),
			get("EMDN code"),
			get("EMDN term"),
			resolved.NamingSource,
		}
		for colIdx, v := range values {
			setCell(f, sheet, outRow, colIdx+1, v)
		}
		outRow++
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, fmt.Sprintf("A1:I%d", outRow-1), nil)
	widths := map[string]float64{"A": 52, "B": 28, "C": 34, "D": 24, "E": 34, "F": 34, "G": 18, "H": 44, "I": 18}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
	return nil
}

func rebuildFamilyNamingReviewSheet(f *excelize.File, devicesSheet string) error {
	sheet := "Family Naming Review"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	headers := []string{"Device type", "Device family", "Device count", "Consistency status", "Common name standard", "Canonical name standard", "Common names standard", "Distinct common names", "Distinct canonical names", "Distinct alias strings", "Source device types seen", "Sample legacy source names"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		styleHeader(f, sheet, cell)
	}
	rows, err := f.GetRows(devicesSheet)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return nil
	}
	headerMap := map[string]int{}
	for i, h := range rows[0] {
		headerMap[h] = i
	}
	type agg struct {
		count     int
		common    map[string]bool
		canonical map[string]bool
		aliases   map[string]bool
		sources   map[string]bool
		legacy    []string
		legacySet map[string]bool
	}
	grouped := map[string]*agg{}
	keyFor := func(t, fam string) string { return t + "\x00" + fam }
	for _, row := range rows[1:] {
		get := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return row[idx]
			}
			return ""
		}
		dt := get("Device type")
		fam := get("Device family")
		if dt == "" || fam == "" {
			continue
		}
		k := keyFor(dt, fam)
		if _, ok := grouped[k]; !ok {
			grouped[k] = &agg{common: map[string]bool{}, canonical: map[string]bool{}, aliases: map[string]bool{}, sources: map[string]bool{}, legacySet: map[string]bool{}}
		}
		a := grouped[k]
		a.count++
		if v := get("Name"); v != "" {
			a.common[v] = true
		}
		if v := get("Canonical device name"); v != "" {
			a.canonical[v] = true
		}
		if v := get("Common names"); v != "" {
			a.aliases[v] = true
		}
		if v := get("Source device type"); v != "" {
			a.sources[v] = true
		}
		if v := get("Legacy source name"); v != "" && !a.legacySet[v] {
			a.legacySet[v] = true
			a.legacy = append(a.legacy, v)
		}
	}
	type rowData struct {
		Type, Family                                     string
		Count                                            int
		Status, CommonStd, CanonicalStd, AliasStd        string
		DistinctCommon, DistinctCanonical, DistinctAlias int
		Sources, Sample                                  string
	}
	var outRows []rowData
	for k, a := range grouped {
		parts := strings.Split(k, "\x00")
		dt, fam := parts[0], parts[1]
		commonVals := sortedKeys(a.common)
		canonicalVals := sortedKeys(a.canonical)
		aliasVals := sortedKeys(a.aliases)
		status := "Needs review"
		if len(commonVals) <= 1 && len(canonicalVals) <= 1 && len(aliasVals) <= 1 {
			status = "Consistent"
		}
		commonStd := ""
		if len(commonVals) > 0 {
			commonStd = commonVals[0]
		}
		canonicalStd := ""
		if len(canonicalVals) > 0 {
			canonicalStd = canonicalVals[0]
		}
		aliasStd := ""
		if len(aliasVals) > 0 {
			aliasStd = aliasVals[0]
		}
		sources := strings.Join(sortedKeys(a.sources), ", ")
		sample := strings.Join(limit(a.legacy, 5), " | ")
		outRows = append(outRows, rowData{dt, fam, a.count, status, commonStd, canonicalStd, aliasStd, len(commonVals), len(canonicalVals), len(aliasVals), sources, sample})
	}
	sort.Slice(outRows, func(i, j int) bool {
		if outRows[i].Type != outRows[j].Type {
			return outRows[i].Type < outRows[j].Type
		}
		return outRows[i].Family < outRows[j].Family
	})
	for i, r := range outRows {
		values := []string{r.Type, r.Family, fmt.Sprintf("%d", r.Count), r.Status, r.CommonStd, r.CanonicalStd, r.AliasStd, fmt.Sprintf("%d", r.DistinctCommon), fmt.Sprintf("%d", r.DistinctCanonical), fmt.Sprintf("%d", r.DistinctAlias), r.Sources, r.Sample}
		for colIdx, v := range values {
			setCell(f, sheet, i+2, colIdx+1, v)
		}
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, fmt.Sprintf("A1:L%d", len(outRows)+1), nil)
	widths := map[string]float64{"A": 34, "B": 38, "C": 12, "D": 18, "E": 28, "F": 42, "G": 54, "H": 18, "I": 20, "J": 18, "K": 42, "L": 64}
	for col, w := range widths {
		f.SetColWidth(sheet, col, col, w)
	}
	return nil
}

func rebuildFamilyNamingAuditSheet(f *excelize.File, devicesSheet string) error {
	sheet := "Family Naming Audit"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	for i, h := range []string{"Metric", "Value"} {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		styleHeader(f, sheet, cell)
	}
	rows, err := f.GetRows(devicesSheet)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return nil
	}
	headerMap := map[string]int{}
	for i, h := range rows[0] {
		headerMap[h] = i
	}
	type agg struct {
		common, canonical, aliases map[string]bool
	}
	grouped := map[string]*agg{}
	for _, row := range rows[1:] {
		get := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return row[idx]
			}
			return ""
		}
		dt := get("Device type")
		fam := get("Device family")
		if dt == "" || fam == "" {
			continue
		}
		k := dt + "\x00" + fam
		if _, ok := grouped[k]; !ok {
			grouped[k] = &agg{common: map[string]bool{}, canonical: map[string]bool{}, aliases: map[string]bool{}}
		}
		a := grouped[k]
		if v := get("Name"); v != "" {
			a.common[v] = true
		}
		if v := get("Canonical device name"); v != "" {
			a.canonical[v] = true
		}
		if v := get("Common names"); v != "" {
			a.aliases[v] = true
		}
	}
	familiesTotal := len(grouped)
	commonConsistent := 0
	canonicalConsistent := 0
	aliasConsistent := 0
	fullyConsistent := 0
	for _, a := range grouped {
		if len(a.common) <= 1 {
			commonConsistent++
		}
		if len(a.canonical) <= 1 {
			canonicalConsistent++
		}
		if len(a.aliases) <= 1 {
			aliasConsistent++
		}
		if len(a.common) <= 1 && len(a.canonical) <= 1 && len(a.aliases) <= 1 {
			fullyConsistent++
		}
	}
	metrics := [][]string{
		{"Families total", fmt.Sprintf("%d", familiesTotal)},
		{"Families with consistent common name", fmt.Sprintf("%d", commonConsistent)},
		{"Families with consistent canonical name", fmt.Sprintf("%d", canonicalConsistent)},
		{"Families with consistent alias string", fmt.Sprintf("%d", aliasConsistent)},
		{"Families fully consistent across all naming fields", fmt.Sprintf("%d", fullyConsistent)},
		{"Families requiring review", fmt.Sprintf("%d", familiesTotal-fullyConsistent)},
	}
	for i, m := range metrics {
		setCell(f, sheet, i+2, 1, m[0])
		setCell(f, sheet, i+2, 2, m[1])
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, fmt.Sprintf("A1:B%d", len(metrics)+1), nil)
	f.SetColWidth(sheet, "A", "A", 48)
	f.SetColWidth(sheet, "B", "B", 16)
	return nil
}

// applyValidations wires an Excel dropdown for each of the Devices sheet's 4
// conventional columns (B–E) to its Lookups column, for whichever of those
// fields the taxonomy actually declares allowed values for — a taxonomy
// missing one (or all) of them just doesn't get a dropdown there.
func applyValidations(f *excelize.File, sheet string, maxRow int, tax *Taxonomy, lookupCol map[string]string) error {
	devicesCol := map[string]string{
		FieldDeviceType:            "B",
		FieldDeviceCategory:        "C",
		FieldDeviceFunction:        "D",
		FieldDeviceApplicationRisk: "E",
	}
	for key, dCol := range devicesCol {
		lCol, ok := lookupCol[key]
		if !ok {
			continue
		}
		fd := tax.Field(key)
		if fd == nil || len(fd.AllowedValues) == 0 {
			continue
		}
		dv := excelize.NewDataValidation(true)
		dv.SetSqref(fmt.Sprintf("%s2:%s%d", dCol, dCol, maxRow))
		dv.SetSqrefDropList(fmt.Sprintf("Lookups!$%s$2:$%s$%d", lCol, lCol, len(fd.AllowedValues)+1))
		f.AddDataValidation(sheet, dv)
	}
	return nil
}

func writeAPIImportCSV(f *excelize.File, devicesSheet, outputPath string, tax *Taxonomy) (string, error) {
	rows, err := f.GetRows(devicesSheet)
	if err != nil {
		return "", err
	}
	if len(rows) < 1 {
		return "", nil
	}
	headerMap := map[string]int{}
	for i, h := range rows[0] {
		headerMap[h] = i
	}
	csvPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".api_import.csv"
	file, err := os.Create(csvPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	w.Write(DefaultAPIImportHeaders)
	seen := map[string]bool{}
	for _, row := range rows[1:] {
		get := func(key string) string {
			if idx, ok := headerMap[key]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		values := []string{
			get("Name"),
			get("Device type"),
			get("Device function"),
			get("Device application risk"),
			get("EMDN code"),
			get("EMDN term"),
		}
		if values[0] == "" || values[1] == "" || values[2] == "" || values[3] == "" {
			continue
		}
		key := strings.Join(values, "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		w.Write(values)
	}
	w.Flush()
	return csvPath, w.Error()
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func limit(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// NormalizeWorkbook normalizes an input workbook using the built-in Ovahol rules.
func NormalizeWorkbook(inputPath, outputPath string) (string, error) {
	return NormalizeWorkbookWithTaxonomy(inputPath, outputPath, nil)
}

// NormalizeWorkbookAs is the vendor-aware workbook entry point. It accepts
// a loaded *Taxonomy and normalizes the workbook using that taxonomy.
func NormalizeWorkbookAs(inputPath, outputPath string, tax interface{}) (string, error) {
	if t, ok := tax.(*Taxonomy); ok {
		return NormalizeWorkbookWithTaxonomy(inputPath, outputPath, t)
	}
	return NormalizeWorkbook(inputPath, outputPath)
}

// NormalizeWorkbookWithTaxonomy normalizes a workbook with the given vendor taxonomy.
func NormalizeWorkbookWithTaxonomy(inputPath, outputPath string, tax *Taxonomy) (string, error) {
	if tax == nil {
		tax = DefaultTaxonomy()
	}
	if err := tax.Validate(); err != nil {
		return "", err
	}
	f, err := excelize.OpenFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("open input: %w", err)
	}
	defer f.Close()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return "", fmt.Errorf("no sheets in workbook")
	}
	firstSheet := sheets[0]
	rows, _, err := extractDeviceRows(f, firstSheet)
	if err != nil {
		return "", err
	}
	devicesSheet, err := rebuildDevicesSheet(f, firstSheet, rows, tax)
	if err != nil {
		return "", err
	}
	if err := rebuildAPIImportSheet(f, devicesSheet, tax); err != nil {
		return "", err
	}
	lookupCol, err := rebuildLookupsSheet(f, tax)
	if err != nil {
		return "", err
	}
	if err := rebuildNamingRulesSheet(f, tax); err != nil {
		return "", err
	}
	if err := rebuildInferenceRulesSheet(f, tax); err != nil {
		return "", err
	}
	if err := rebuildCommonNameMappingReviewSheet(f, devicesSheet, tax); err != nil {
		return "", err
	}
	if err := rebuildFamilyNamingReviewSheet(f, devicesSheet); err != nil {
		return "", err
	}
	if err := rebuildFamilyNamingAuditSheet(f, devicesSheet); err != nil {
		return "", err
	}
	// validations on Devices sheet
	rr, _ := f.GetRows(devicesSheet)
	maxRow := len(rr)
	if maxRow < 2 {
		maxRow = 2
	}
	if err := applyValidations(f, devicesSheet, maxRow, tax, lookupCol); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", err
	}
	if err := f.SaveAs(outputPath); err != nil {
		return "", err
	}
	csvPath, err := writeAPIImportCSV(f, devicesSheet, outputPath, tax)
	if err != nil {
		return "", err
	}
	return csvPath, nil
}

// UpdateOntology is an alias for NormalizeWorkbook, kept for compatibility
// with code that used the original Python-derived name.
func UpdateOntology(inputPath, outputPath string) (string, error) {
	return NormalizeWorkbook(inputPath, outputPath)
}

// NormalizeCSV reads a CSV file with interchange columns and returns normalized Results.
func NormalizeCSV(path string) ([]Result, error) {
	return NormalizeCSVWithTaxonomy(path, nil)
}

// NormalizeCSVWithTaxonomy normalizes a CSV file using the given vendor taxonomy.
func NormalizeCSVWithTaxonomy(path string, tax *Taxonomy) ([]Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ontology: open CSV %s: %w", path, err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ontology: read CSV %s: %w", path, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("ontology: CSV %s is empty", path)
	}
	headers := rows[0]
	headerMap := make(map[string]int, len(headers))
	for i, h := range headers {
		headerMap[strings.TrimSpace(h)] = i
	}
	find := func(keys ...string) int {
		for _, k := range keys {
			if idx, ok := headerMap[k]; ok {
				return idx
			}
		}
		return -1
	}
	emdnCodeKey := "EMDN code"
	emdnTermKey := "EMDN term"
	for _, h := range headers {
		if strings.HasPrefix(h, "Nomenclature code (EMDN)") {
			emdnCodeKey = h
		}
		if strings.HasPrefix(h, "Nomenclature term (EMDN)") {
			emdnTermKey = h
		}
	}
	var results []Result
	for _, row := range rows[1:] {
		get := func(keys ...string) string {
			idx := find(keys...)
			if idx >= 0 && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}
		in := Input{
			DeviceName: get("Legacy source name", "Device name", "name", "Name", "common_name"),
			SourceType: get("Source device type", "Device type", "source_type"),
			EMDNCode:   get("EMDN code", "emdn_code", emdnCodeKey),
			EMDNTerm:   get("EMDN term", "emdn_term", emdnTermKey),
		}
		// Skip entirely empty rows
		if in.DeviceName == "" && in.SourceType == "" && in.EMDNCode == "" && in.EMDNTerm == "" {
			continue
		}
		results = append(results, NormalizeWithTaxonomy(in, tax))
	}
	return results, nil
}
