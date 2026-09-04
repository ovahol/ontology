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
//
// Any input column that isn't recognized as one of the known device/taxonomy
// fields (model, manufacturer, serial number, location, whatever the source
// inventory happens to carry) is not classification data ontology understands
// — but it isn't discarded either. It's kept verbatim, under its original
// header text, in both the returned rows and the extraHeaders list (in
// original column order), so rebuildDevicesSheet can pass it straight through
// to the output. ontology only ever adds classification columns; it never
// drops caller data it doesn't recognize.
func extractDeviceRows(f *excelize.File, sheetName string) (result []map[string]string, extraHeaders []string, err error) {
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
	consumed := map[int]bool{}
	find := func(keys ...string) int {
		for _, k := range keys {
			if idx, ok := headerMap[k]; ok {
				consumed[idx] = true
				return idx
			}
		}
		return -1
	}
	get := func(row []string, keys ...string) string {
		idx := find(keys...)
		if idx >= 0 && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
		return ""
	}
	// Resolve every known field once (against the header row only, via find's
	// consumed side effect) so `consumed` is complete before we decide what's
	// left over as passthrough.
	find("Name", "Common name", "common_name")
	find("Canonical device name", "canonical_device_name")
	find("Common names", "Search aliases", "search_aliases", "common_names")
	find("Device type", "device_type", "Ovahol device type", "ovahol_device_type")
	find("Device family", "device_family", "Ovahol device family", "ovahol_device_family")
	find("Device function", "device_function")
	find("Device application risk", "device_application_risk")
	find("Legacy source name", "Device name", "name", "legacy_source_name")
	find("Source device type", "source_type")
	find("EMDN code", "emdn_code", emdnCodeKey)
	find("EMDN term", "emdn_term", emdnTermKey)
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if h == "" || consumed[i] {
			continue
		}
		extraHeaders = append(extraHeaders, h)
	}

	for _, row := range rows[1:] {
		// Also allow "Device name" / "name" etc. — support both new agnostic and legacy Ovahol headers for backward compat
		m := map[string]string{
			"Name":                    get(row, "Name", "Common name", "common_name"),
			"Canonical device name":   get(row, "Canonical device name", "canonical_device_name"),
			"Common names":            get(row, "Common names", "Search aliases", "search_aliases", "common_names"),
			"Device type":             get(row, "Device type", "device_type", "Ovahol device type", "ovahol_device_type"),
			"Device family":           get(row, "Device family", "device_family", "Ovahol device family", "ovahol_device_family"),
			"Device function":         get(row, "Device function", "device_function"),
			"Device application risk": get(row, "Device application risk", "device_application_risk"),
			"Legacy source name":      get(row, "Legacy source name", "Device name", "name", "legacy_source_name"),
			"Source device type":      get(row, "Source device type", "source_type"),
			"EMDN code":               get(row, "EMDN code", "emdn_code", emdnCodeKey),
			"EMDN term":               get(row, "EMDN term", "emdn_term", emdnTermKey),
		}
		for _, h := range extraHeaders {
			if idx, ok := headerMap[h]; ok && idx < len(row) {
				m[h] = strings.TrimSpace(row[idx])
			}
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
	return result, extraHeaders, nil
}

func colLetter(n int) string {
	letter, _ := excelize.ColumnNumberToName(n)
	return letter
}

// deviceSheetLayout is the 1-based column layout of the Devices sheet: Name,
// then one column per tax.Fields entry in taxonomy order, then the fixed
// input-echo columns. It's computed once and shared by every function that
// reads or writes the Devices sheet, so "where is device_type's column"
// is answered by field key (works for any taxonomy shape) rather than by
// matching literal header text (which breaks the moment a taxonomy picks its
// own Label wording).
//
// EMDN code/term get one column each, not two: if the taxonomy itself
// declares "emdn_code"/"emdn_term" as fields (as a dictionary-derived
// taxonomy naturally does, since rules resolve them from a device-name
// match), that field's column IS the EMDN column — emdnCodeIsField/
// emdnTermIsField record this so callers know to leave the raw input echo
// out of it rather than clobbering the resolved value. Only a taxonomy that
// doesn't model EMDN at all gets a separate, input-echo-only column.
type deviceSheetLayout struct {
	fieldCol                         map[string]int
	legacyCol, sourceCol             int
	emdnCodeCol, emdnTermCol         int
	emdnCodeIsField, emdnTermIsField bool
	totalFixed                       int
}

func newDeviceSheetLayout(tax *Taxonomy) deviceSheetLayout {
	l := deviceSheetLayout{fieldCol: map[string]int{}}
	col := 1 // column 1 is always Name
	for _, fd := range tax.Fields {
		col++
		l.fieldCol[fd.Key] = col
	}
	l.legacyCol = col + 1
	l.sourceCol = col + 2
	next := col + 2
	if c, ok := l.fieldCol["emdn_code"]; ok {
		l.emdnCodeCol = c
		l.emdnCodeIsField = true
	} else {
		next++
		l.emdnCodeCol = next
	}
	if c, ok := l.fieldCol["emdn_term"]; ok {
		l.emdnTermCol = c
		l.emdnTermIsField = true
	} else {
		next++
		l.emdnTermCol = next
	}
	l.totalFixed = next
	return l
}

// deviceSheetHeaders returns the Devices sheet's header row for tax, with
// extraHeaders (passthrough columns ontology didn't recognize) appended.
func deviceSheetHeaders(tax *Taxonomy, extraHeaders []string) []string {
	layout := newDeviceSheetLayout(tax)
	headers := []string{"Name"}
	for _, fd := range tax.Fields {
		label := fd.Label
		if label == "" {
			label = fd.Key
		}
		headers = append(headers, label)
	}
	headers = append(headers, "Legacy source name", "Source device type")
	if !layout.emdnCodeIsField {
		headers = append(headers, "EMDN code")
	}
	if !layout.emdnTermIsField {
		headers = append(headers, "EMDN term")
	}
	headers = append(headers, extraHeaders...)
	return headers
}

// colAt reads row[col-1], treating col==0 (a field the taxonomy doesn't
// declare) or an out-of-range row as simply absent rather than an error.
func colAt(row []string, col int) string {
	if col <= 0 || col-1 >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[col-1])
}

func rebuildDevicesSheet(f *excelize.File, sheetName string, rows []map[string]string, extraHeaders []string, tax *Taxonomy, resolve rowResolver) (string, error) {
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

	layout := newDeviceSheetLayout(tax)
	headers := deviceSheetHeaders(tax, extraHeaders)
	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheetName, cell, header)
		styleHeader(f, sheetName, cell)
	}
	for rowIdx, row := range rows {
		resolved := resolve(row, tax)
		values := make([]string, len(headers))
		values[0] = resolved.Name
		for _, fd := range tax.Fields {
			values[layout.fieldCol[fd.Key]-1] = resolved.Fields[fd.Key]
		}
		values[layout.legacyCol-1] = row["Legacy source name"]
		values[layout.sourceCol-1] = row["Source device type"]
		// EMDN: if the taxonomy resolved emdn_code/emdn_term as a field (e.g.
		// from a dictionary-name match), that value wins; the raw input echo
		// only fills in when the taxonomy didn't resolve one.
		if values[layout.emdnCodeCol-1] == "" {
			values[layout.emdnCodeCol-1] = row["EMDN code"]
		}
		if values[layout.emdnTermCol-1] == "" {
			values[layout.emdnTermCol-1] = row["EMDN term"]
		}
		for i, h := range extraHeaders {
			values[layout.totalFixed+i] = row[h]
		}
		for colIdx, v := range values {
			setCell(f, sheetName, rowIdx+2, colIdx+1, v)
		}
	}
	// freeze, filter, widths
	f.SetPanes(sheetName, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	maxRow := len(rows) + 1
	if maxRow >= 2 {
		f.AutoFilter(sheetName, fmt.Sprintf("A1:%s%d", colLetter(len(headers)), maxRow), nil)
	}
	f.SetColWidth(sheetName, "A", "A", 28)
	for i := range tax.Fields {
		col := colLetter(i + 2)
		f.SetColWidth(sheetName, col, col, 30)
	}
	f.SetColWidth(sheetName, colLetter(layout.legacyCol), colLetter(layout.legacyCol), 48)
	f.SetColWidth(sheetName, colLetter(layout.sourceCol), colLetter(layout.sourceCol), 34)
	f.SetColWidth(sheetName, colLetter(layout.emdnCodeCol), colLetter(layout.emdnCodeCol), 18)
	f.SetColWidth(sheetName, colLetter(layout.emdnTermCol), colLetter(layout.emdnTermCol), 44)
	for i := range extraHeaders {
		col := colLetter(layout.totalFixed + i + 1)
		f.SetColWidth(sheetName, col, col, 24)
	}
	f.SetRowHeight(sheetName, 1, 18)
	return sheetName, nil
}

// apiImportHeaders returns ["name", <tax.Fields keys in order>, "emdn_code",
// "emdn_term"] — the API Import sheet/CSV's dynamic column set. A vendor's
// own field keys are used verbatim, since this feeds whatever API expects
// those exact keys.
func apiImportHeaders(tax *Taxonomy) []string {
	layout := newDeviceSheetLayout(tax)
	headers := []string{"name"}
	for _, fd := range tax.Fields {
		headers = append(headers, fd.Key)
	}
	if !layout.emdnCodeIsField {
		headers = append(headers, "emdn_code")
	}
	if !layout.emdnTermIsField {
		headers = append(headers, "emdn_term")
	}
	return headers
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
	headers := apiImportHeaders(tax)
	for colIdx, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, header)
		styleHeader(f, sheet, cell)
	}
	// read devices sheet rows, by position — the Devices sheet's classification
	// columns are keyed by tax.Fields order, not by matching header text, so
	// this works regardless of what Label a taxonomy chose for its fields.
	rows, err := f.GetRows(devicesSheet)
	if err != nil {
		return err
	}
	if len(rows) < 1 {
		return nil
	}
	layout := newDeviceSheetLayout(tax)
	seen := map[string]bool{}
	outRow := 2
	for _, row := range rows[1:] {
		name := colAt(row, 1)
		if name == "" {
			continue
		}
		fieldVals := make(map[string]string, len(tax.Fields))
		requiredOK := true
		for _, fd := range tax.Fields {
			v := colAt(row, layout.fieldCol[fd.Key])
			fieldVals[fd.Key] = v
			if fd.Required && v == "" {
				requiredOK = false
			}
		}
		if !requiredOK {
			continue
		}
		emdnCode := colAt(row, layout.emdnCodeCol)
		emdnTerm := colAt(row, layout.emdnTermCol)

		rec := APIImportRecord{Name: name, Fields: fieldVals, EMDNCode: emdnCode, EMDNTerm: emdnTerm}
		key := rec.DedupKey()
		if seen[key] {
			continue
		}
		seen[key] = true

		values := []string{name}
		for _, fd := range tax.Fields {
			values = append(values, fieldVals[fd.Key])
		}
		if !layout.emdnCodeIsField {
			values = append(values, emdnCode)
		}
		if !layout.emdnTermIsField {
			values = append(values, emdnTerm)
		}
		for colIdx, v := range values {
			setCell(f, sheet, outRow, colIdx+1, v)
		}
		outRow++
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	if outRow > 2 {
		f.AutoFilter(sheet, fmt.Sprintf("A1:%s%d", colLetter(len(headers)), outRow-1), nil)
	}
	f.SetColWidth(sheet, "A", "A", 34)
	for i := range tax.Fields {
		col := colLetter(i + 2)
		f.SetColWidth(sheet, col, col, 30)
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

func rebuildCommonNameMappingReviewSheet(f *excelize.File, devicesSheet string, tax *Taxonomy, resolve rowResolver) error {
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
	layout := newDeviceSheetLayout(tax)
	headers := []string{"Legacy source name", "Name"}
	for _, fd := range tax.Fields {
		label := fd.Label
		if label == "" {
			label = fd.Key
		}
		headers = append(headers, label)
	}
	if !layout.emdnCodeIsField {
		headers = append(headers, "EMDN code")
	}
	if !layout.emdnTermIsField {
		headers = append(headers, "EMDN term")
	}
	headers = append(headers, "Mapping source")
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
	outRow := 2
	for _, row := range rows[1:] {
		legacy := colAt(row, layout.legacyCol)
		source := colAt(row, layout.sourceCol)
		emdnTerm := colAt(row, layout.emdnTermCol)
		// Re-resolve to get naming_source
		rmap := map[string]string{
			"Legacy source name": legacy,
			"Source device type": source,
			"EMDN term":          emdnTerm,
		}
		resolved := resolve(rmap, tax)
		values := []string{legacy, colAt(row, 1)}
		for _, fd := range tax.Fields {
			values = append(values, colAt(row, layout.fieldCol[fd.Key]))
		}
		if !layout.emdnCodeIsField {
			values = append(values, colAt(row, layout.emdnCodeCol))
		}
		if !layout.emdnTermIsField {
			values = append(values, emdnTerm)
		}
		values = append(values, resolved.NamingSource)
		for colIdx, v := range values {
			setCell(f, sheet, outRow, colIdx+1, v)
		}
		outRow++
	}
	f.SetPanes(sheet, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})
	f.AutoFilter(sheet, fmt.Sprintf("A1:%s%d", colLetter(len(headers)), outRow-1), nil)
	f.SetColWidth(sheet, "A", "A", 52)
	f.SetColWidth(sheet, "B", "B", 28)
	for i := range tax.Fields {
		col := colLetter(i + 3)
		f.SetColWidth(sheet, col, col, 30)
	}
	return nil
}

func rebuildFamilyNamingReviewSheet(f *excelize.File, devicesSheet string, tax *Taxonomy) error {
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
	// This sheet only produces rows for taxonomies that use the conventional
	// device_type/device_family keys — it's an optional, Ovahol-shaped
	// consistency audit, not something every taxonomy needs. A taxonomy
	// without device_family (or without either key) just gets an empty sheet.
	layout := newDeviceSheetLayout(tax)
	typeCol, famCol := layout.fieldCol[FieldDeviceType], layout.fieldCol[FieldDeviceFamily]
	if typeCol == 0 || famCol == 0 {
		return nil
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
		dt := colAt(row, typeCol)
		fam := colAt(row, famCol)
		if dt == "" || fam == "" {
			continue
		}
		k := keyFor(dt, fam)
		if _, ok := grouped[k]; !ok {
			grouped[k] = &agg{common: map[string]bool{}, canonical: map[string]bool{}, aliases: map[string]bool{}, sources: map[string]bool{}, legacySet: map[string]bool{}}
		}
		a := grouped[k]
		a.count++
		if v := colAt(row, 1); v != "" {
			a.common[v] = true
		}
		if v := colAt(row, layout.sourceCol); v != "" {
			a.sources[v] = true
		}
		if v := colAt(row, layout.legacyCol); v != "" && !a.legacySet[v] {
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

func rebuildFamilyNamingAuditSheet(f *excelize.File, devicesSheet string, tax *Taxonomy) error {
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
	// Same optional-sheet caveat as rebuildFamilyNamingReviewSheet: only
	// meaningful for taxonomies using the conventional device_type/
	// device_family keys.
	layout := newDeviceSheetLayout(tax)
	typeCol, famCol := layout.fieldCol[FieldDeviceType], layout.fieldCol[FieldDeviceFamily]
	type agg struct {
		common, canonical, aliases map[string]bool
	}
	grouped := map[string]*agg{}
	if typeCol != 0 && famCol != 0 {
		for _, row := range rows[1:] {
			dt := colAt(row, typeCol)
			fam := colAt(row, famCol)
			if dt == "" || fam == "" {
				continue
			}
			k := dt + "\x00" + fam
			if _, ok := grouped[k]; !ok {
				grouped[k] = &agg{common: map[string]bool{}, canonical: map[string]bool{}, aliases: map[string]bool{}}
			}
			a := grouped[k]
			if v := colAt(row, 1); v != "" {
				a.common[v] = true
			}
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

// applyValidations wires an Excel dropdown for every Devices sheet field
// column to its Lookups column, for every field the taxonomy declares
// allowed values for — however many dimensions it has, not just the 4
// conventional ones.
func applyValidations(f *excelize.File, sheet string, maxRow int, tax *Taxonomy, lookupCol map[string]string) error {
	layout := newDeviceSheetLayout(tax)
	for _, fd := range tax.Fields {
		if len(fd.AllowedValues) == 0 {
			continue
		}
		lCol, ok := lookupCol[fd.Key]
		if !ok {
			continue
		}
		dCol := colLetter(layout.fieldCol[fd.Key])
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
	csvPath := strings.TrimSuffix(outputPath, filepath.Ext(outputPath)) + ".api_import.csv"
	file, err := os.Create(csvPath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	w := csv.NewWriter(file)
	w.Write(apiImportHeaders(tax))
	layout := newDeviceSheetLayout(tax)
	seen := map[string]bool{}
	for _, row := range rows[1:] {
		name := colAt(row, 1)
		if name == "" {
			continue
		}
		values := []string{name}
		requiredOK := true
		for _, fd := range tax.Fields {
			v := colAt(row, layout.fieldCol[fd.Key])
			values = append(values, v)
			if fd.Required && v == "" {
				requiredOK = false
			}
		}
		if !requiredOK {
			continue
		}
		if !layout.emdnCodeIsField {
			values = append(values, colAt(row, layout.emdnCodeCol))
		}
		if !layout.emdnTermIsField {
			values = append(values, colAt(row, layout.emdnTermCol))
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

// NormalizeWorkbook normalizes an input workbook using the embedded default
// (MeDevis) taxonomy.
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
	return normalizeWorkbook(inputPath, outputPath, tax, taxonomyRowResolver)
}

// NormalizeWorkbookWithCatalogAndTaxonomy normalizes a workbook with the given
// vendor taxonomy, reconciling each row against the device dictionary cat
// before falling back to taxonomy rules. This is the migration path for a host
// system that already has a device dictionary: known names resolve verbatim
// (MappingSource="catalog_exact"/"catalog_fuzzy"), and everything else falls
// back to taxonomy inference. A nil cat behaves exactly like
// NormalizeWorkbookWithTaxonomy.
func NormalizeWorkbookWithCatalogAndTaxonomy(inputPath, outputPath string, cat Catalog, tax *Taxonomy) (string, error) {
	return normalizeWorkbook(inputPath, outputPath, tax, makeCatalogRowResolver(cat))
}

// normalizeWorkbook is the shared workbook pipeline. resolve decides how each
// row is turned into a ResolvedRow (taxonomy rules alone, or catalog-first).
func normalizeWorkbook(inputPath, outputPath string, tax *Taxonomy, resolve rowResolver) (string, error) {
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
	rows, extraHeaders, err := extractDeviceRows(f, firstSheet)
	if err != nil {
		return "", err
	}
	devicesSheet, err := rebuildDevicesSheet(f, firstSheet, rows, extraHeaders, tax, resolve)
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
	if err := rebuildCommonNameMappingReviewSheet(f, devicesSheet, tax, resolve); err != nil {
		return "", err
	}
	if err := rebuildFamilyNamingReviewSheet(f, devicesSheet, tax); err != nil {
		return "", err
	}
	if err := rebuildFamilyNamingAuditSheet(f, devicesSheet, tax); err != nil {
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
