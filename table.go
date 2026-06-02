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
	hasExplicitStyle := len(styles) > 0
	explicitStyle := Style{}
	if hasExplicitStyle {
		explicitStyle = styles[0]
	}
	return func() Element {
		focusStyle := explicitStyle
		if !hasExplicitStyle {
			focusStyle = UseTheme().Focused
		}

		offset, setOffset := UseState(0)
		colOffset, setColOffset := UseState(0)
		isFocused := UseFocus()

		termW := UseTermSize().Width
		widths := tableCalcWidths(cols, rows)

		// Determine whether horizontal scrolling is needed.
		totalW := tableRowWidth(widths, 0, len(widths))
		needsHScroll := totalW > termW

		// Compute the visible column range given colOffset and terminal width.
		colStart, colEnd := 0, len(cols)
		if needsHScroll {
			colStart = colOffset
			colEnd = tableVisibleColEnd(widths, colOffset, termW)
		}

		// Keep selected row visible — silent clamp, no extra render needed.
		if selected < offset {
			offset = selected
		} else if height > 0 && selected >= offset+height {
			offset = selected - height + 1
		}

		end := offset + height
		if end > len(rows) {
			end = len(rows)
		}

		UseScroll(func(delta int) bool {
			if !isFocused {
				return false
			}
			next := clampIndex(selected+clampDelta(delta, height), len(rows))
			if next != selected {
				onSelect(next)
			}
			return true
		})

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
			case KeyLeft:
				if needsHScroll && colOffset > 0 {
					setColOffset(colOffset - 1)
				}
			case KeyRight:
				if needsHScroll && colOffset < len(cols)-1 {
					setColOffset(colOffset + 1)
				}
			}
		})

		// top-border + header + mid-border occupy the first 3 rows; data rows follow.
		const tableHeaderRows = 3
		UseClick(func(_, localY int) {
			dataRow := localY - tableHeaderRows
			if dataRow >= 0 && dataRow < (end-offset) {
				target := offset + dataRow
				if target < len(rows) {
					onSelect(target)
				}
			}
		})

		visWidths := widths[colStart:colEnd]

		headers := make([]string, colEnd-colStart)
		for i, col := range cols[colStart:colEnd] {
			headers[i] = col.Header
		}

		hasLeft := colStart > 0
		hasRight := colEnd < len(cols)
		lines := []Element{
			Text(tableBorderTopScroll(visWidths, hasLeft, hasRight)),
			Text(tableFormatRow(headers, visWidths)),
			Text(tableBorderMid(visWidths, offset > 0)),
		}

		for i, row := range rows[offset:end] {
			actualIdx := offset + i
			style := NewStyle()
			visRow := make([]string, colEnd-colStart)
			for j := range visRow {
				if colStart+j < len(row) {
					visRow[j] = row[colStart+j]
				}
			}
			runes := []rune(tableFormatRow(visRow, visWidths))
			if actualIdx == selected {
				runes[0] = '▶'
				if isFocused {
					style = focusStyle
				} else {
					style = NewStyle().Bold()
				}
			}
			lines = append(lines, Text(string(runes), style))
		}

		lines = append(lines, Text(tableBorderBot(visWidths, end < len(rows))))
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
		Text(tableBorderMid(widths, false), style),
	}
	for _, row := range rows {
		lines = append(lines, Text(tableFormatRow(row, widths), style))
	}
	lines = append(lines, Text(tableBorderBot(widths, false), style))

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
	return tableBorderTopScroll(widths, false, false)
}

// tableBorderTopScroll builds the top border, replacing corners with ◀/▶ when
// columns are hidden to the left or right respectively.
func tableBorderTopScroll(widths []int, hasLeft, hasRight bool) string {
	segs := make([]string, len(widths))
	for i, w := range widths {
		segs[i] = strings.Repeat("─", w+2)
	}
	left, right := "┌", "┐"
	if hasLeft {
		left = "◀"
	}
	if hasRight {
		right = "▶"
	}
	return left + strings.Join(segs, "┬") + right
}

// tableVisibleColEnd returns the exclusive end index of the columns that fit
// within maxWidth starting from colOffset. Always returns at least colOffset+1
// so at least one column is shown.
func tableVisibleColEnd(widths []int, colOffset, maxWidth int) int {
	used := 2 // left and right border characters
	end := colOffset
	for end < len(widths) {
		w := widths[end] + 2 // content + two spaces of padding
		if end > colOffset {
			w++ // column separator
		}
		if used+w > maxWidth {
			break
		}
		used += w
		end++
	}
	if end <= colOffset {
		end = colOffset + 1
	}
	return end
}

// tableRowWidth returns the character width of a rendered table row spanning
// columns [start, end).
func tableRowWidth(widths []int, start, end int) int {
	if end <= start {
		return 2
	}
	total := 2 // outer borders
	for i := start; i < end && i < len(widths); i++ {
		total += widths[i] + 2 // content + padding
		if i > start {
			total++ // separator
		}
	}
	return total
}

// tableBorderMid builds the header-separator border, replacing the right corner
// with ↑ when rows are hidden above the current viewport.
func tableBorderMid(widths []int, hasAbove bool) string {
	segs := make([]string, len(widths))
	for i, w := range widths {
		segs[i] = strings.Repeat("─", w+2)
	}
	right := "┤"
	if hasAbove {
		right = "↑"
	}
	return "├" + strings.Join(segs, "┼") + right
}

// tableBorderBot builds the bottom border, replacing the right corner with ↓
// when rows are hidden below the current viewport.
func tableBorderBot(widths []int, hasBelow bool) string {
	segs := make([]string, len(widths))
	for i, w := range widths {
		segs[i] = strings.Repeat("─", w+2)
	}
	right := "┘"
	if hasBelow {
		right = "↓"
	}
	return "└" + strings.Join(segs, "┴") + right
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
