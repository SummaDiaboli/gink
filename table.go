package gink

import "strings"

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
