package gink

import "strings"

// NewTable returns a focusable, scrollable table component with row selection.
//
// cols defines column headers and width constraints (same as [Table]).
// rows is the full data set. selected is the index of the highlighted row —
// the parent owns this state and updates it via onSelect. height is the number
// of data rows visible at once; rows beyond that scroll into view as the
// selection moves.
//
// Up/Down navigate rows; input is ignored when the component is not focused.
// The selected row is rendered with a ▶ cursor and bold cyan style when
// focused, or bold when not focused.
//
//	sel, setSel := gink.UseState(0)
//	gink.C(gink.NewTable(cols, rows, sel, setSel, 8))
func NewTable(cols []Column, rows [][]string, selected int, onSelect func(int), height int, styles ...Style) func() Element {
	focusStyle := NewStyle().Bold().Foreground(ColorBrightCyan)
	if len(styles) > 0 {
		focusStyle = styles[0]
	}
	return func() Element {
		offset, setOffset := UseState(0)
		isFocused := UseFocus()

		// Keep selected visible — silent clamp, no extra render needed.
		if selected < offset {
			offset = selected
		} else if height > 0 && selected >= offset+height {
			offset = selected - height + 1
		}

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyUp:
				if selected > 0 {
					next := selected - 1
					onSelect(next)
					if next < offset {
						setOffset(next)
					}
				}
			case KeyDown:
				if selected < len(rows)-1 {
					next := selected + 1
					onSelect(next)
					if next >= offset+height {
						setOffset(next - height + 1)
					}
				}
			}
		})

		widths := tableCalcWidths(cols, rows)

		end := offset + height
		if end > len(rows) {
			end = len(rows)
		}

		headers := make([]string, len(cols))
		for i, col := range cols {
			headers[i] = col.Header
		}

		lines := []Element{
			Text(tableBorderTop(widths)),
			Text(tableFormatRow(headers, widths)),
			Text(tableBorderMid(widths)),
		}

		for i, row := range rows[offset:end] {
			actualIdx := offset + i
			cursor := "  "
			style := NewStyle()
			if actualIdx == selected {
				cursor = "▶ "
				if isFocused {
					style = focusStyle
				} else {
					style = NewStyle().Bold()
				}
			}
			lines = append(lines, Text(cursor+tableFormatRow(row, widths), style))
		}

		lines = append(lines, Text(tableBorderBot(widths)))
		return Box(lines...)
	}
}

// Column defines one column in a [Table].
// MinWidth and MaxWidth constrain the computed content width; zero means no constraint.
//
//	cols := []gink.Column{
//	    {Header: "Name",   MinWidth: 12},
//	    {Header: "Status", MaxWidth: 8},
//	    {Header: "Region"},
//	}
type Column struct {
	Header   string
	MinWidth int // 0 = fit to content
	MaxWidth int // 0 = no limit
}

// Table renders a bordered table with a header row and zero or more data rows.
// Column widths are computed from the widest content (header or any cell), then
// clamped by [Column.MinWidth] and [Column.MaxWidth]. Content that exceeds
// MaxWidth is truncated with a "…" suffix.
//
// An optional style applies to all border and header-separator characters.
//
//	cols := []gink.Column{
//	    {Header: "Name",   MinWidth: 12},
//	    {Header: "Status", MaxWidth: 8},
//	}
//	rows := [][]string{
//	    {"web-01", "ONLINE"},
//	    {"web-02", "WARN"},
//	}
//	gink.Table(cols, rows, gink.NewStyle().Foreground(gink.ColorBrightCyan))
func Table(cols []Column, rows [][]string, styles ...Style) Element {
	var style Style
	if len(styles) > 0 {
		style = styles[0]
	}

	widths := tableCalcWidths(cols, rows)

	headers := make([]string, len(cols))
	for i, col := range cols {
		headers[i] = col.Header
	}

	lines := []Element{
		Text(tableBorderTop(widths), style),
		Text(tableFormatRow(headers, widths), style),
		Text(tableBorderMid(widths), style),
	}
	for _, row := range rows {
		lines = append(lines, Text(tableFormatRow(row, widths), style))
	}
	lines = append(lines, Text(tableBorderBot(widths), style))

	return Box(lines...)
}

// tableCalcWidths returns the computed content width for each column.
func tableCalcWidths(cols []Column, rows [][]string) []int {
	widths := make([]int, len(cols))
	for i, col := range cols {
		w := len([]rune(col.Header))
		for _, row := range rows {
			if i < len(row) {
				if cw := len([]rune(row[i])); cw > w {
					w = cw
				}
			}
		}
		if col.MinWidth > 0 && w < col.MinWidth {
			w = col.MinWidth
		}
		if col.MaxWidth > 0 && w > col.MaxWidth {
			w = col.MaxWidth
		}
		widths[i] = w
	}
	return widths
}

// tableFormatRow formats cells into a │-delimited row, padding or truncating each
// cell to its column width with one space of padding on each side.
func tableFormatRow(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for i, w := range widths {
		var content string
		if i < len(cells) {
			content = tableTruncate(cells[i], w)
		}
		pad := w - len([]rune(content))
		parts[i] = " " + content + strings.Repeat(" ", pad) + " "
	}
	return "│" + strings.Join(parts, "│") + "│"
}

func tableBorderTop(widths []int) string {
	segs := make([]string, len(widths))
	for i, w := range widths {
		segs[i] = strings.Repeat("─", w+2)
	}
	return "┌" + strings.Join(segs, "┬") + "┐"
}

func tableBorderMid(widths []int) string {
	segs := make([]string, len(widths))
	for i, w := range widths {
		segs[i] = strings.Repeat("─", w+2)
	}
	return "├" + strings.Join(segs, "┼") + "┤"
}

func tableBorderBot(widths []int) string {
	segs := make([]string, len(widths))
	for i, w := range widths {
		segs[i] = strings.Repeat("─", w+2)
	}
	return "└" + strings.Join(segs, "┴") + "┘"
}

// tableTruncate clips s to maxWidth runes, replacing the last rune with "…" when
// truncation occurs. Returns s unchanged when len([]rune(s)) ≤ maxWidth.
func tableTruncate(s string, maxWidth int) string {
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 1 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-1]) + "…"
}
