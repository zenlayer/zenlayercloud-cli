package output

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// colPadding is the total horizontal padding applied inside each cell (2 spaces each side).
const colPadding = 4

// longTextThreshold is the minimum rune-count at which a top-level scalar value
// is extracted as a footnote instead of occupying an inline table cell.
const longTextThreshold = 60

// TableFormatter formats output as a human-readable table using box-drawing
// borders, matching the style of Tencent Cloud CLI (tccli).
//
// Visual layout:
//
//	---------------------------
//	|       instanceSet       |
//	+--------+---------+------+
//	|   id   |  name   | ...  |
//	+--------+---------+------+
//	|  ins-1 |  Alice  | ...  |
//	+--------+---------+------+
//
// Nested objects are rendered as indented sub-sections wrapped with || pipes.
// When the table is wider than the terminal, single-row sections are
// automatically converted to vertical key|value layout (tccli auto-reformat).
type TableFormatter struct{}

// Format implements Formatter for table output.
func (f *TableFormatter) Format(w io.Writer, data interface{}) error {
	var m map[string]interface{}
	switch v := data.(type) {
	case map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		m = v
	case map[string]string:
		if len(v) == 0 {
			return nil
		}
		m = make(map[string]interface{}, len(v))
		for k, val := range v {
			m[k] = val
		}
	case nil:
		fmt.Fprintln(w, "<nil>")
		return nil
	default:
		fmt.Fprintln(w, data)
		return nil
	}

	mt := &multiTable{}
	mt.newSection("", 0)
	mt.buildFromDict(m, 0)
	mt.renderAll(w)
	return nil
}

// ─── tableSection ─────────────────────────────────────────────────────────────

type tableSection struct {
	title     string
	headers   []string
	rows      [][]string
	indent    int
	maxWidths []int // max content rune-width per column (without padding/borders)
}

func (s *tableSection) addHeader(cols []string) {
	s.headers = cols
	s.updateWidths(cols)
}

func (s *tableSection) addRow(row []string) {
	s.rows = append(s.rows, row)
	s.updateWidths(row)
}

func (s *tableSection) updateWidths(cols []string) {
	for i, c := range cols {
		l := utf8.RuneCountInString(c)
		if i >= len(s.maxWidths) {
			s.maxWidths = append(s.maxWidths, l)
		} else if l > s.maxWidths[i] {
			s.maxWidths[i] = l
		}
	}
}

// naturalColWidths returns per-column cell widths with colPadding applied.
// Cell width = max_content_rune_width + colPadding.
func (s *tableSection) naturalColWidths() []int {
	ws := make([]int, len(s.maxWidths))
	for i, w := range s.maxWidths {
		ws[i] = w + colPadding
	}
	return ws
}

// naturalInnerWidth is the sum of natural column widths, floored by the
// minimum width needed to display the title. Returns 0 for empty sections.
func (s *tableSection) naturalInnerWidth() int {
	ws := s.naturalColWidths()
	if len(ws) == 0 {
		if s.title == "" {
			return 0
		}
		return utf8.RuneCountInString(s.title) + colPadding
	}
	total := 0
	for _, w := range ws {
		total += w
	}
	if minT := utf8.RuneCountInString(s.title) + colPadding; total < minT {
		total = minT
	}
	return total
}

// totalWidth returns the total rendered width including indent pipe chars.
func (s *tableSection) totalWidth() int {
	return s.naturalInnerWidth() + s.indent*2
}

// convertToVertical flips a single-row table with headers into a 2-column
// key | value layout, matching tccli's auto-reformat behaviour when the table
// is wider than the terminal.
func (s *tableSection) convertToVertical() {
	if len(s.rows) != 1 || len(s.headers) == 0 {
		return
	}
	newRows := make([][]string, len(s.headers))
	for i, h := range s.headers {
		val := ""
		if i < len(s.rows[0]) {
			val = s.rows[0][i]
		}
		newRows[i] = []string{h, val}
	}
	s.headers = nil
	s.rows = newRows
	s.maxWidths = nil
	for _, row := range s.rows {
		s.updateWidths(row)
	}
}

// ─── multiTable ───────────────────────────────────────────────────────────────

type footnote struct {
	key   string
	value string
}

type multiTable struct {
	sections  []*tableSection
	current   *tableSection
	footnotes []footnote
}

func (mt *multiTable) newSection(title string, indent int) {
	s := &tableSection{title: title, indent: indent}
	mt.sections = append(mt.sections, s)
	mt.current = s
}

func (mt *multiTable) addHeader(cols []string) {
	if mt.current != nil {
		mt.current.addHeader(cols)
	}
}

func (mt *multiTable) addRow(row []string) {
	if mt.current != nil {
		mt.current.addRow(row)
	}
}

// buildFromDict populates the current section with inline key-value pairs and
// recursively creates sub-sections for complex (non-inline) values.
// At indent 0, scalar values exceeding longTextThreshold are collected as
// footnotes and printed after the table rather than in a cell.
func (mt *multiTable) buildFromDict(m map[string]interface{}, indent int) {
	inlines, complex := groupKeys(m)

	var short, long []string
	for _, k := range inlines {
		if indent == 0 && (len(inlines) > 1 || len(complex) > 0) &&
			utf8.RuneCountInString(formatValue(m[k])) > longTextThreshold {
			long = append(long, k)
		} else {
			short = append(short, k)
		}
	}
	for _, k := range long {
		mt.footnotes = append(mt.footnotes, footnote{key: k, value: formatValue(m[k])})
	}

	switch {
	case len(short) == 1:
		k := short[0]
		mt.addRow([]string{k, formatValue(m[k])})
	case len(short) > 1:
		mt.addHeader(short)
		row := make([]string, len(short))
		for i, k := range short {
			row[i] = formatValue(m[k])
		}
		mt.addRow(row)
	}
	for _, k := range complex {
		mt.buildValue(k, m[k], indent+1)
	}
}

// buildValue creates a sub-section for a complex (non-scalar) value.
func (mt *multiTable) buildValue(title string, v interface{}, indent int) {
	if v == nil {
		return
	}
	switch val := v.(type) {
	case map[string]interface{}:
		mt.newSection(title, indent)
		mt.buildFromDict(val, indent)
	case []interface{}:
		if len(val) == 0 {
			return
		}
		if allMaps(val) {
			mt.buildFromList(title, val, indent)
		} else {
			mt.newSection(title, indent)
			for _, item := range val {
				mt.addRow([]string{fmt.Sprintf("%v", item)})
			}
		}
	default:
		mt.newSection(title, indent)
		mt.addRow([]string{fmt.Sprintf("%v", val)})
	}
}

// buildFromList creates a titled section with column headers derived from the
// union of scalar keys across all items in the list.
func (mt *multiTable) buildFromList(title string, items []interface{}, indent int) {
	scalarCols, complexCols := groupKeysFromList(items)

	mt.newSection(title, indent)
	if len(scalarCols) > 0 {
		mt.addHeader(scalarCols)
	}

	first := true
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		// When an element has nested columns, start a new (untitled) sub-table
		// section so nested sub-sections follow the correct row.
		// No title on continuation sections to avoid repeating the parent label.
		if !first && len(complexCols) > 0 {
			mt.newSection("", indent)
			if len(scalarCols) > 0 {
				mt.addHeader(scalarCols)
			}
		}
		first = false

		if len(scalarCols) > 0 {
			row := make([]string, len(scalarCols))
			for i, col := range scalarCols {
				row[i] = formatValue(m[col])
			}
			mt.addRow(row)
		}
		for _, col := range complexCols {
			if val, exists := m[col]; exists {
				mt.buildValue(col, val, indent+1)
			}
		}
	}
}

// ─── rendering ───────────────────────────────────────────────────────────────

func (mt *multiTable) renderAll(w io.Writer) {
	if len(mt.sections) == 0 {
		return
	}
	maxW := 0
	for _, s := range mt.sections {
		if tw := s.totalWidth(); tw > maxW {
			maxW = tw
		}
	}
	if maxW == 0 {
		return
	}

	// Auto-reformat: convert single-row sections to vertical key|value layout
	// when the table would be wider than the terminal, matching tccli behaviour.
	if termW := terminalWidth(); maxW > termW {
		for _, s := range mt.sections {
			s.convertToVertical()
		}
		maxW = 0
		for _, s := range mt.sections {
			if tw := s.totalWidth(); tw > maxW {
				maxW = tw
			}
		}
	}

	fmt.Fprintln(w, strings.Repeat("-", maxW))
	for _, s := range mt.sections {
		renderSection(w, s, maxW)
	}
	for _, fn := range mt.footnotes {
		fmt.Fprintf(w, "\n%s:\n%s\n", fn.key, fn.value)
	}
}

func renderSection(w io.Writer, s *tableSection, maxW int) {
	ind := strings.Repeat("|", s.indent)
	innerW := maxW - s.indent*2

	if s.title != "" {
		fmt.Fprintln(w, ind+makeCenterStr(s.title, innerW, "|", "|")+ind)
	}

	if len(s.headers) == 0 && len(s.rows) == 0 {
		if s.title != "" {
			fmt.Fprintln(w, ind+"+"+strings.Repeat("-", innerW-2)+"+"+ind)
		}
		return
	}

	colWs := scaleWidths(s.naturalColWidths(), innerW)

	if len(s.headers) > 0 {
		fmt.Fprintln(w, ind+makeRowSep(colWs)+ind)
		fmt.Fprintln(w, ind+makeHeaderRow(s.headers, colWs)+ind)
	}

	if len(s.rows) > 0 {
		fmt.Fprintln(w, ind+makeRowSep(colWs)+ind)
		for _, row := range s.rows {
			fmt.Fprintln(w, ind+makeDataRow(row, colWs)+ind)
		}
		fmt.Fprintln(w, ind+makeRowSep(colWs)+ind)
	}
}

// makeRowSep returns a +---+---+ separator line.
//
// Rendering rule (mirrors tccli):
//   - first column: "+" + "-"*(w-2) + "+"  → w chars
//   - others:       "-"*(w-1)      + "+"  → w chars
//
// Total rune-width == sum(colWs).
func makeRowSep(colWs []int) string {
	var sb strings.Builder
	for i, w := range colWs {
		if i == 0 {
			sb.WriteByte('+')
			writeRep(&sb, '-', w-2)
			sb.WriteByte('+')
		} else {
			writeRep(&sb, '-', w-1)
			sb.WriteByte('+')
		}
	}
	return sb.String()
}

// makeHeaderRow returns a "| col1 | col2 |" header line with centered text.
// Total rune-width == sum(colWs).
func makeHeaderRow(headers []string, colWs []int) string {
	var sb strings.Builder
	for i, h := range headers {
		w := colWs[i]
		if i == 0 {
			sb.WriteString(makeCenterStr(h, w, "|", "|"))
		} else {
			sb.WriteString(makeCenterStr(h, w, "", "|"))
		}
	}
	return sb.String()
}

// makeDataRow returns a "|  val1  |  val2  |" data line with left-aligned text.
// Total rune-width == sum(colWs).
func makeDataRow(row []string, colWs []int) string {
	var sb strings.Builder
	for i, w := range colWs {
		val := ""
		if i < len(row) {
			val = row[i]
		}
		if i == 0 {
			sb.WriteString(makeAlignLeft(val, w, "|", "|"))
		} else {
			sb.WriteString(makeAlignLeft(val, w, "", "|"))
		}
	}
	return sb.String()
}

// makeCenterStr centers text in a cell of total rune-width w.
// Postcondition: utf8.RuneCountInString(result) == w.
func makeCenterStr(text string, w int, leftEdge, rightEdge string) string {
	edgeW := len(leftEdge) + len(rightEdge) // edges are single-byte ASCII
	inner := w - edgeW
	if inner <= 0 {
		return leftEdge + rightEdge
	}
	textW := utf8.RuneCountInString(text)
	if textW > inner {
		text = string([]rune(text)[:inner])
		textW = inner
	}
	pad := inner - textW
	padL := pad / 2
	padR := pad - padL
	return leftEdge + strings.Repeat(" ", padL) + text + strings.Repeat(" ", padR) + rightEdge
}

// makeAlignLeft left-aligns text in a cell with 2-space left padding.
// Postcondition: utf8.RuneCountInString(result) == w.
func makeAlignLeft(text string, w int, leftEdge, rightEdge string) string {
	const leftPad = 2
	edgeW := len(leftEdge) + len(rightEdge) // edges are single-byte ASCII
	available := w - edgeW - leftPad
	if available < 0 {
		available = 0
	}
	textW := utf8.RuneCountInString(text)
	if textW > available {
		text = string([]rune(text)[:available])
		textW = available
	}
	rightPad := available - textW
	return leftEdge + strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad) + rightEdge
}

// scaleWidths scales column widths proportionally so their sum equals
// targetTotal, matching tccli's column-width calculation logic.
func scaleWidths(widths []int, targetTotal int) []int {
	if len(widths) == 0 {
		return widths
	}
	natural := 0
	for _, w := range widths {
		natural += w
	}
	if natural == targetTotal {
		return append([]int(nil), widths...)
	}

	const minColW = 2
	scaled := make([]int, len(widths))
	for i, w := range widths {
		scaled[i] = int(float64(w)/float64(natural)*float64(targetTotal) + 0.5)
		if scaled[i] < minColW {
			scaled[i] = minColW
		}
	}

	// Fix rounding so sum(scaled) == targetTotal exactly.
	for iter := 0; iter < len(scaled)*3; iter++ {
		off := 0
		for _, w := range scaled {
			off += w
		}
		off -= targetTotal
		if off == 0 {
			break
		}
		for i := range scaled {
			if off > 0 && scaled[i] > minColW {
				scaled[i]--
				off--
			} else if off < 0 {
				scaled[i]++
				off++
			}
			if off == 0 {
				break
			}
		}
	}
	return scaled
}

func writeRep(sb *strings.Builder, c byte, n int) {
	for i := 0; i < n; i++ {
		sb.WriteByte(c)
	}
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func groupKeys(m map[string]interface{}) (inlines, complex []string) {
	for k, v := range m {
		if isInlineValue(v) {
			inlines = append(inlines, k)
		} else {
			complex = append(complex, k)
		}
	}
	sort.Strings(inlines)
	sort.Strings(complex)
	return
}

func groupKeysFromList(items []interface{}) (inlines, complex []string) {
	inlineSet := make(map[string]struct{})
	complexSet := make(map[string]struct{})
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		for k, v := range m {
			if isInlineValue(v) {
				// Only add to inlines if not already classified as complex by another item.
				if _, alreadyComplex := complexSet[k]; !alreadyComplex {
					inlineSet[k] = struct{}{}
				}
			} else {
				complexSet[k] = struct{}{}
				delete(inlineSet, k) // complex takes priority if a previous item was inline
			}
		}
	}
	for k := range inlineSet {
		inlines = append(inlines, k)
	}
	for k := range complexSet {
		complex = append(complex, k)
	}
	sort.Strings(inlines)
	sort.Strings(complex)
	return
}

// isScalar returns true when v is a primitive value (not a collection).
func isScalar(v interface{}) bool {
	switch v.(type) {
	case nil, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, string:
		return true
	default:
		return false
	}
}

// isInlineValue reports whether v can be rendered as a single-line string.
// Extends isScalar to cover scalar arrays and small flat-object arrays.
func isInlineValue(v interface{}) bool {
	switch val := v.(type) {
	case nil, bool, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, string:
		return true
	case []interface{}:
		return isInlineArray(val)
	default:
		return false
	}
}

// isInlineArray reports whether an array is renderable as a compact inline string.
// Scalar arrays are always inline. Object arrays are inline when every item is a
// flat map (all-scalar fields, ≤ 4 fields per item) and the list has ≤ 10 items.
func isInlineArray(arr []interface{}) bool {
	if len(arr) == 0 {
		return true
	}
	if isAllScalar(arr) {
		return true
	}
	if !allMaps(arr) || len(arr) > 10 {
		return false
	}
	for _, item := range arr {
		m := item.(map[string]interface{})
		if len(m) > 4 {
			return false
		}
		for _, mv := range m {
			if !isScalar(mv) {
				return false
			}
		}
	}
	return true
}

func isAllScalar(arr []interface{}) bool {
	for _, item := range arr {
		if !isScalar(item) {
			return false
		}
	}
	return true
}

// allMaps returns true when every element of items is a map[string]interface{}.
func allMaps(items []interface{}) bool {
	for _, item := range items {
		if _, ok := item.(map[string]interface{}); !ok {
			return false
		}
	}
	return true
}

// formatValue renders any value accepted by isInlineValue.
func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	if arr, ok := v.([]interface{}); ok {
		return formatInlineArray(arr)
	}
	return fmt.Sprintf("%v", v)
}

// formatInlineArray renders a scalar or small-object array as a compact string.
func formatInlineArray(arr []interface{}) string {
	if len(arr) == 0 {
		return ""
	}
	if isAllScalar(arr) {
		parts := make([]string, len(arr))
		for i, item := range arr {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(parts, ", ")
	}
	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			parts = append(parts, fmt.Sprintf("%v", item))
			continue
		}
		parts = append(parts, formatInlineObject(m))
	}
	return strings.Join(parts, ", ")
}

// formatInlineObject renders a flat map as "primaryVal(secondary1,secondary2,...)"
// using identifier-like fields as the primary, or "val1/val2" when none found.
func formatInlineObject(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	primary := findPrimaryKey(keys)
	if primary == "" || len(keys) == 1 {
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = fmt.Sprintf("%v", m[k])
		}
		return strings.Join(parts, "/")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%v", m[primary]))
	others := make([]string, 0, len(keys)-1)
	for _, k := range keys {
		if k != primary {
			others = append(others, fmt.Sprintf("%v", m[k]))
		}
	}
	sb.WriteByte('(')
	sb.WriteString(strings.Join(others, ","))
	sb.WriteByte(')')
	return sb.String()
}

// findPrimaryKey returns the most identifier-like key from a sorted key list,
// scanning common suffixes in order of specificity.
func findPrimaryKey(keys []string) string {
	for _, suffix := range []string{"type", "id", "name", "code", "key"} {
		for _, k := range keys {
			if strings.HasSuffix(strings.ToLower(k), suffix) {
				return k
			}
		}
	}
	return ""
}
