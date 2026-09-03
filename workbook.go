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
		Font: &excelize.Font{Bold: true, Color: "FFFFFF", Size: 10},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"1F4B99"}, Pattern: 1},
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
			"Name":                  get("Name", "Common name", "common_name"),
			"Canonical device name": get("Canonical device name", "canonical_device_name"),
			"Common names":          get("Common names", "Search aliases", "search_aliases", "common_names"),
			"Device type":           get("Device type", "device_type", "Ovahol device type", "ovahol_device_type"),
			"Device family":         get("Device family", "device_family", "Ovahol device family", "ovahol_device_family"),
			"Device function":       get("Device function", "device_function"),
			"Device application risk": get("Device application risk", "device_application_risk"),
			"Legacy source name":    get("Legacy source name", "Device name", "name", "legacy_source_name"),
			"Source device type":    get("Source device type", "source_type"),
			"EMDN code":             get("EMDN code", "emdn_code", emdnCodeKey),
			"EMDN term":             get("EMDN term", "emdn_term", emdnTermKey),
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

func rebuildDevicesSheet(f *excelize.File, sheetName string, rows []map[string]string) (string, error) {
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

	for colIdx, header := range DeviceSheetHeaders {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, header)
		styleHeader(f, sheetName, cell)
	}
	for rowIdx, row := range rows {
		resolved := ResolveRowNaming(row)
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

func rebuildAPIImportSheet(f *excelize.File, devicesSheet string) error {
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
	for colIdx, header := range APIImportHeaders {
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
		key := strings.Join(values, "\x00")
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

func rebuildLookupsSheet(f *excelize.File) error {
	sheet := "Lookups"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	headers := []struct{ Cell, Value string }{
		{"A1", "Device types"},
		{"C1", "Device categories"},
		{"E1", "Device functions"},
		{"H1", "Device application risks"},
	}
	for _, h := range headers {
		f.SetCellValue(sheet, h.Cell, h.Value)
		styleHeader(f, sheet, h.Cell)
	}
	for i, dt := range DeviceTypes {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), dt.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), dt.Code)
	}
	for i, dc := range DeviceCategories {
		f.SetCellValue(sheet, fmt.Sprintf("C%d", i+2), dc.Name)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", i+2), dc.Code)
	}
	for i, fn := range DeviceFunctions {
		f.SetCellValue(sheet, fmt.Sprintf("E%d", i+2), fn.Name)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", i+2), fn.Code)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", i+2), fn.Category)
	}
	for i, r := range DeviceApplicationRisks {
		f.SetCellValue(sheet, fmt.Sprintf("H%d", i+2), r.Description)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", i+2), r.ScorePoint)
	}
	f.SetColWidth(sheet, "A", "A", 44)
	f.SetColWidth(sheet, "B", "B", 40)
	f.SetColWidth(sheet, "C", "C", 24)
	f.SetColWidth(sheet, "D", "D", 24)
	f.SetColWidth(sheet, "E", "E", 44)
	f.SetColWidth(sheet, "F", "F", 32)
	f.SetColWidth(sheet, "G", "G", 24)
	f.SetColWidth(sheet, "H", "H", 42)
	f.SetColWidth(sheet, "I", "I", 12)
	return nil
}

func rebuildNamingRulesSheet(f *excelize.File) error {
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
	for i, r := range NamingRules {
		setCell(f, sheet, i+2, 1, r.Field)
		setCell(f, sheet, i+2, 2, r.Rule)
		setCell(f, sheet, i+2, 3, r.GoodExample)
		setCell(f, sheet, i+2, 4, r.Avoid)
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	if len(NamingRules) > 0 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:D%d", len(NamingRules)+1), nil)
	}
	f.SetColWidth(sheet, "A", "A", 20)
	f.SetColWidth(sheet, "B", "B", 72)
	f.SetColWidth(sheet, "C", "C", 34)
	f.SetColWidth(sheet, "D", "D", 42)
	return nil
}

func rebuildFamilyRulesSheet(f *excelize.File) error {
	sheet := "Family Rules"
	if idx, err := f.GetSheetIndex(sheet); err == nil {
		_ = idx
		f.DeleteSheet(sheet)
	}
	newIdx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	_ = newIdx
	headers := []string{"Device type", "Device family", "Suggested name", "Suggested canonical name", "Default device function", "Default application risk", "Match hints"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		styleHeader(f, sheet, cell)
	}
	for i, r := range FamilyRules {
		hints := []string{}
		if len(r.SourceTypes) > 0 {
			sort.Strings(r.SourceTypes)
			hints = append(hints, "source: "+strings.Join(r.SourceTypes, ", "))
		}
		if len(r.Keywords) > 0 {
			hints = append(hints, "keywords: "+strings.Join(r.Keywords, ", "))
		}
		values := []string{r.Type, r.Family, r.CommonName, r.CanonicalName, r.Function, r.Risk, strings.Join(hints, " | ")}
		for colIdx, v := range values {
			setCell(f, sheet, i+2, colIdx+1, v)
		}
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, fmt.Sprintf("A1:G%d", len(FamilyRules)+1), nil)
	f.SetColWidth(sheet, "A", "A", 34)
	f.SetColWidth(sheet, "B", "B", 34)
	f.SetColWidth(sheet, "C", "C", 28)
	f.SetColWidth(sheet, "D", "D", 42)
	f.SetColWidth(sheet, "E", "E", 34)
	f.SetColWidth(sheet, "F", "F", 34)
	f.SetColWidth(sheet, "G", "G", 72)
	return nil
}

func rebuildCommonNameMappingReviewSheet(f *excelize.File, devicesSheet string) error {
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
		resolved := ResolveRowNaming(rmap)
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
		count    int
		common   map[string]bool
		canonical map[string]bool
		aliases  map[string]bool
		sources  map[string]bool
		legacy   []string
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
		Type, Family string
		Count int
		Status, CommonStd, CanonicalStd, AliasStd string
		DistinctCommon, DistinctCanonical, DistinctAlias int
		Sources, Sample string
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

func applyValidations(f *excelize.File, sheet string, maxRow int) error {
	typeEnd := 1 + len(DeviceTypes)
	categoryEnd := 1 + len(DeviceCategories)
	funcEnd := 1 + len(DeviceFunctions)
	riskEnd := 1 + len(DeviceApplicationRisks)
	validations := []struct {
		SqRef, Formula string
	}{
		{fmt.Sprintf("B2:B%d", maxRow), fmt.Sprintf("Lookups!$A$2:$A$%d", typeEnd)},
		{fmt.Sprintf("C2:C%d", maxRow), fmt.Sprintf("Lookups!$C$2:$C$%d", categoryEnd)},
		{fmt.Sprintf("D2:D%d", maxRow), fmt.Sprintf("Lookups!$E$2:$E$%d", funcEnd)},
		{fmt.Sprintf("E2:E%d", maxRow), fmt.Sprintf("Lookups!$H$2:$H$%d", riskEnd)},
	}
	for _, v := range validations {
		dv := excelize.NewDataValidation(true)
		dv.SetSqref(v.SqRef)
		dv.SetSqrefDropList(v.Formula)
		f.AddDataValidation(sheet, dv)
	}
	return nil
}

func writeAPIImportCSV(f *excelize.File, devicesSheet, outputPath string) (string, error) {
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
	w.Write(APIImportHeaders)
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

// UpdateOntology is the main entrypoint, mirroring Python's main().
// NormalizeWorkbook is the workbook interchange entry point for bulk migration.
// It reads any spreadsheet with at least a device name column, normalizes
// every row through the ontology, and writes a fully populated
// workbook plus a deduplicated API-import CSV.
//
// UpdateOntology is kept as an alias for backward compatibility with the
// original Python script's naming.
func NormalizeWorkbook(inputPath, outputPath string) (string, error) {
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
	devicesSheet, err := rebuildDevicesSheet(f, firstSheet, rows)
	if err != nil {
		return "", err
	}
	if err := rebuildAPIImportSheet(f, devicesSheet); err != nil {
		return "", err
	}
	if err := rebuildLookupsSheet(f); err != nil {
		return "", err
	}
	if err := rebuildNamingRulesSheet(f); err != nil {
		return "", err
	}
	if err := rebuildFamilyRulesSheet(f); err != nil {
		return "", err
	}
	if err := rebuildCommonNameMappingReviewSheet(f, devicesSheet); err != nil {
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
	if err := applyValidations(f, devicesSheet, maxRow); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", err
	}
	if err := f.SaveAs(outputPath); err != nil {
		return "", err
	}
	csvPath, err := writeAPIImportCSV(f, devicesSheet, outputPath)
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

// NormalizeCSV reads a CSV file with interchange columns (at minimum
// "Legacy source name" or "Device name" and "Source device type") and
// returns normalized Results. The CSV header row is flexible — the same
// tolerant header matching as the workbook path is used.
func NormalizeCSV(path string) ([]Result, error) {
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
		results = append(results, Normalize(in))
	}
	return results, nil
}
