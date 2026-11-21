package gink

// virtualBufferHeight is the minimum height of the virtual render buffer —
// a generous upper bound for any practical TUI app.
const virtualBufferHeight = 512

// virtualHeight returns the height of the virtual render buffer. It is large
// enough to hold content that exceeds the visible terminal height.
func virtualHeight(termH int) int {
	if termH > virtualBufferHeight {
		return termH
	}
	return virtualBufferHeight
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
		setBufferRune(screen, right, 0, '↑', style)
	}
	if offset+screen.Height < contentH {
		setBufferRune(screen, right, screen.Height-1, '↓', style)
	}
}

func (s *runtimeState) scrollUp(n int) {
	s.scrollOffset -= n
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
}

func (s *runtimeState) scrollDown(n int) {
	s.scrollOffset += n
	s.clampScroll()
}

func (s *runtimeState) availableHeight() int {
	h := s.currentTermSize.Height - s.footerHeight
	if h < 0 {
		h = 0
	}
	return h
}

func (s *runtimeState) clampScroll() {
	avail := s.availableHeight()
	max := s.scrollContent - avail
	if max < 0 {
		max = 0
	}
	if s.scrollOffset > max {
		s.scrollOffset = max
	}
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
}

func (s *runtimeState) scrollBy(delta int) {
	if s.dispatchScroll(delta) {
		return
	}
	if delta > 0 {
		s.scrollDown(delta)
	} else {
		s.scrollUp(-delta)
	}
}

func (s *runtimeState) scrollToY(y int) {
	avail := s.availableHeight()
	if y < s.scrollOffset {
		s.scrollOffset = y
	} else if y >= s.scrollOffset+avail {
		s.scrollOffset = y - avail + 1
	}
	if s.scrollOffset < 0 {
		s.scrollOffset = 0
	}
}
