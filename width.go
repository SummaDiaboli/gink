package gink

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

func displayWidth(s string) int {
	width := 0
	state := -1
	for len(s) > 0 {
		_, rest, w, next := uniseg.FirstGraphemeClusterInString(s, state)
		width += clusterWidth(w)
		s = rest
		state = next
	}
	return width
}

func clusterWidth(w int) int {
	if w < 1 {
		return 1
	}
	return w
}

func splitGraphemes(s string) []string {
	clusters := make([]string, 0)
	state := -1
	for len(s) > 0 {
		cluster, rest, _, next := uniseg.FirstGraphemeClusterInString(s, state)
		clusters = append(clusters, cluster)
		s = rest
		state = next
	}
	return clusters
}

func truncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if displayWidth(s) <= maxWidth {
		return s
	}

	var b strings.Builder
	width := 0
	state := -1
	for len(s) > 0 {
		cluster, rest, w, next := uniseg.FirstGraphemeClusterInString(s, state)
		w = clusterWidth(w)
		if width+w > maxWidth {
			break
		}
		b.WriteString(cluster)
		width += w
		s = rest
		state = next
	}
	return b.String()
}

func truncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if displayWidth(s) <= maxWidth {
		return s
	}
	if maxWidth == 1 {
		return "…"
	}
	return truncateToWidth(s, maxWidth-1) + "…"
}

func padToWidth(s string, width int) string {
	if w := displayWidth(s); w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return truncateToWidth(s, width)
}

func byteOffsetAtGrapheme(s string, index int) int {
	if index <= 0 {
		return 0
	}

	idx := 0
	offset := 0
	state := -1
	for len(s) > 0 {
		cluster, rest, _, next := uniseg.FirstGraphemeClusterInString(s, state)
		if idx == index {
			return offset
		}
		idx++
		offset += len(cluster)
		s = rest
		state = next
	}
	return offset
}

func spliceGraphemes(s string, start, end int, replacement string) string {
	clusters := splitGraphemes(s)
	if start < 0 {
		start = 0
	}
	if end > len(clusters) {
		end = len(clusters)
	}
	if end < start {
		end = start
	}
	left := byteOffsetAtGrapheme(s, start)
	right := byteOffsetAtGrapheme(s, end)
	return s[:left] + replacement + s[right:]
}

func blankCell(style tcell.Style) Cell {
	return Cell{Rune: ' ', Grapheme: " ", Width: 1, Style: style}
}

func blankDefaultCell() Cell {
	return blankCell(tcell.StyleDefault)
}

func continuationCell(style tcell.Style) Cell {
	return Cell{Rune: ' ', Grapheme: "", Width: 0, Continuation: true, Style: style}
}

func setBufferCell(buf *Buffer, x, y int, cell Cell) {
	if buf == nil || y < 0 || y >= buf.Height || x < 0 || x >= buf.Width {
		return
	}
	width := cell.Width
	if width < 1 {
		width = 1
	}
	if cell.Continuation || (width > 1 && x+width > buf.Width) {
		return
	}
	cell.Width = width
	cell.Continuation = false
	buf.Cells[y][x] = cell
	for i := 1; i < width && x+i < buf.Width; i++ {
		buf.Cells[y][x+i] = continuationCell(cell.Style)
	}
}

func clearBufferCellSpan(buf *Buffer, x, y int) {
	if buf == nil || y < 0 || y >= buf.Height || x < 0 || x >= buf.Width {
		return
	}
	if buf.Cells[y][x].Continuation {
		for x > 0 && buf.Cells[y][x].Continuation {
			x--
		}
	}
	width := buf.Cells[y][x].Width
	if width < 1 {
		width = 1
	}
	for i := 0; i < width && x+i < buf.Width; i++ {
		buf.Cells[y][x+i] = blankCell(tcell.StyleDefault)
	}
}

func setBufferRune(buf *Buffer, x, y int, r rune, style tcell.Style) {
	clearBufferCellSpan(buf, x, y)
	setBufferCell(buf, x, y, Cell{
		Rune:     r,
		Grapheme: string(r),
		Width:    1,
		Style:    style,
	})
}

func setBufferGrapheme(buf *Buffer, x, y int, grapheme string, style tcell.Style) int {
	if buf == nil || y < 0 || y >= buf.Height || x < 0 || x >= buf.Width || grapheme == "" {
		return 0
	}

	r, _ := utf8.DecodeRuneInString(grapheme)
	if r == utf8.RuneError && len(grapheme) > 0 && grapheme != string(utf8.RuneError) {
		r = []rune(grapheme)[0]
	}
	width := displayWidth(grapheme)
	setBufferCell(buf, x, y, Cell{
		Rune:     r,
		Grapheme: grapheme,
		Width:    width,
		Style:    style,
	})
	return width
}

func writeBufferText(buf *Buffer, x, y int, s string, style tcell.Style) int {
	col := 0
	state := -1
	for len(s) > 0 {
		cluster, rest, w, next := uniseg.FirstGraphemeClusterInString(s, state)
		if w < 1 {
			w = 1
		}
		if y >= 0 && y < buf.Height && x+col+w <= buf.Width {
			setBufferGrapheme(buf, x+col, y, cluster, style)
		}
		col += w
		s = rest
		state = next
	}
	return col
}
