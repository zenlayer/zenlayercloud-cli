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

type multiTable struct {
	sections []*tableSection
	current  *tableSection
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

// buildFromDict populates the current section with scalar key-value pairs and
// recursively creates sub-sections for complex (non-scalar) values.
func (mt *multiTable) buildFromDict(m map[string]interface{}, indent int) {
	scalars, complex := groupKeys(m)
	switch {
	case len(scalars) == 1:
		// Single scalar → vertical key | value row (no header).
		k := scalars[0]
		mt.addRow([]string{k, formatScalar(m[k])})
	case len(scalars) > 1:
		// Multiple scalars → horizontal header + single data row.
		mt.addHeader(scalars)
		row := make([]string, len(scalars))
		for i, k := range scalars {
			row[i] = formatScalar(m[k])
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
				mt.addRow([]string{formatScalar(item)})
			}
		}
	default:
		mt.newSection(title, indent)
		mt.addRow([]string{formatScalar(val)})
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
		// When an element has nested columns, start a new sub-table section for
		// subsequent elements so nested sub-sections follow the correct row.
		if !first && len(complexCols) > 0 {
			mt.newSection(title, indent)
			if len(scalarCols) > 0 {
				mt.addHeader(scalarCols)
			}
		}
		first = false

		if len(scalarCols) > 0 {
			row := make([]string, len(scalarCols))
			for i, col := range scalarCols {
				row[i] = formatScalar(m[col])
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

func groupKeys(m map[string]interface{}) (scalars, complex []string) {
	for k, v := range m {
		if isScalar(v) {
			scalars = append(scalars, k)
		} else {
			complex = append(complex, k)
		}
	}
	sort.Strings(scalars)
	sort.Strings(complex)
	return
}

func groupKeysFromList(items []interface{}) (scalars, complex []string) {
	scalarSet := make(map[string]struct{})
	complexSet := make(map[string]struct{})
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		for k, v := range m {
			if isScalar(v) {
				scalarSet[k] = struct{}{}
			} else {
				complexSet[k] = struct{}{}
			}
		}
	}
	for k := range scalarSet {
		scalars = append(scalars, k)
	}
	for k := range complexSet {
		complex = append(complex, k)
	}
	sort.Strings(scalars)
	sort.Strings(complex)
	return
}

// isScalar returns true when v can be rendered inline as a single value.
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

// allMaps returns true when every element of items is a map[string]interface{}.
func allMaps(items []interface{}) bool {
	for _, item := range items {
		if _, ok := item.(map[string]interface{}); !ok {
			return false
		}
	}
	return true
}

// formatScalar converts a scalar value to its display string.
func formatScalar(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
