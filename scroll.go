package gink

// scrollOffset is the number of rows the global viewport is scrolled down.
// Zero means the top of content is at the top of the screen.
var scrollOffset int

// scrollContent is the detected content height from the last render pass,
// used to clamp the scroll offset.
var scrollContent int

// footerHeight is the height of the sticky footer (0 when AppShell is not used).
// Subtracted from the terminal height when computing scroll bounds.
var footerHeight int

// virtualHeight returns the height of the virtual render buffer. It is large
// enough to hold content that exceeds the visible terminal height — 512 rows
// is a generous upper bound for any practical TUI app.
func virtualHeight(termH int) int {
	if termH > 512 {
		return termH
	}
	return 512
}

// detectContentHeight returns the row index of the last non-blank row plus one,
// giving the true height of rendered content inside a virtual buffer.
func detectContentHeight(buf *Buffer) int {
	for row := buf.Height - 1; row >= 0; row-- {
		for _, cell := range buf.Cells[row] {
			if cell.Rune != ' ' {
				return row + 1
			}
		}
	}
	return 0
}

// applyScroll copies rows [offset, offset+termH) from virtual into a new
// screen-sized buffer. Rows past virtual.Height are left as blank spaces.
func applyScroll(virtual *Buffer, termW, termH, offset int) *Buffer {
	screen := NewBuffer(termW, termH)
	for row := 0; row < termH; row++ {
		src := row + offset
		if src < virtual.Height {
			copy(screen.Cells[row], virtual.Cells[src])
		}
	}
	return screen
}

// addScrollIndicators overlays ↑ and ↓ at the top-right and bottom-right
// corners of screen to signal that content exists above or below the current
// viewport. They are only shown when relevant.
func addScrollIndicators(screen *Buffer, offset, contentH int) {
	if screen.Width == 0 || screen.Height == 0 {
		return
	}
	right := screen.Width - 1
	style := NewStyle().Foreground(ColorBrightWhite).toTcell()
	if offset > 0 {
		screen.Cells[0][right] = Cell{Rune: '↑', Style: style}
	}
	if offset+screen.Height < contentH {
		screen.Cells[screen.Height-1][right] = Cell{Rune: '↓', Style: style}
	}
}

// scrollUp moves the viewport up by n rows, clamped at zero.
func scrollUp(n int) {
	scrollOffset -= n
	if scrollOffset < 0 {
		scrollOffset = 0
	}
}

// scrollDown moves the viewport down by n rows, clamped at the content bottom.
func scrollDown(n int) {
	scrollOffset += n
	clampScroll()
}

// availableHeight returns the number of screen rows available for scrollable content.
func availableHeight() int {
	h := currentTermSize.Height - footerHeight
	if h < 0 {
		h = 0
	}
	return h
}

// clampScroll ensures scrollOffset stays within [0, max(0, contentH-availH)].
func clampScroll() {
	avail := availableHeight()
	max := scrollContent - avail
	if max < 0 {
		max = 0
	}
	if scrollOffset > max {
		scrollOffset = max
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
}

// scrollToY adjusts scrollOffset so that virtual row y is visible.
func scrollToY(y int) {
	avail := availableHeight()
	if y < scrollOffset {
		scrollOffset = y
	} else if y >= scrollOffset+avail {
		scrollOffset = y - avail + 1
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}
}
