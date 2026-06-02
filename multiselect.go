package gink

// NewMultiSelect returns a focusable, scrollable multi-selection list.
//
// items is the full slice of option strings. selected is a parallel bool slice
// where selected[i] indicates whether items[i] is checked — the parent owns
// both slices. onToggle(i) is called when the user toggles item i; the parent
// is responsible for flipping selected[i] and re-rendering.
//
// height is the number of rows visible at once; items beyond that scroll into
// view as the internal cursor moves. Up/Down move the cursor; Space toggles
// the item at the cursor. Scroll indicators (↑/↓) appear when items are
// hidden above or below the viewport. Input is ignored when not focused.
//
// Each item renders as "[ ] Label" or "[x] Label". The cursor row is
// highlighted with the theme's Focused style (or a custom style).
//
//	sel, setSel := gink.UseState(make([]bool, len(items)))
//	toggle := func(i int) {
//	    next := append([]bool(nil), sel...)
//	    next[i] = !next[i]
//	    setSel(next)
//	}
//	gink.C(gink.NewMultiSelect(items, sel, toggle, 8))
func NewMultiSelect(items []string, selected []bool, onToggle func(int), height int, styles ...Style) func() Element {
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

		theme := UseTheme()
		cursor, setCursor := UseState(0)
		offset, setOffset := UseState(0)
		isFocused := UseFocus()

		// Clamp cursor to valid range when items shrink externally.
		if cursor >= len(items) {
			cursor = len(items) - 1
			setCursor(cursor)
		}
		if cursor < 0 {
			cursor = 0
			setCursor(cursor)
		}

		// Keep cursor visible in viewport.
		if cursor < offset {
			offset = cursor
			setOffset(offset)
		} else if height > 0 && cursor >= offset+height {
			offset = cursor - height + 1
			setOffset(offset)
		}

		end := offset + height
		if end > len(items) {
			end = len(items)
		}

		hasAbove := offset > 0
		hasBelow := end < len(items)

		registerScrollHandler(func(delta int) bool {
			if !isFocused {
				return false
			}
			move := delta
			if height > 0 {
				if move > height {
					move = height
				} else if move < -height {
					move = -height
				}
			}
			next := cursor + move
			if next < 0 {
				next = 0
			} else if next >= len(items) {
				next = len(items) - 1
			}
			if next != cursor {
				setCursor(next)
			}
			return true
		})

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyUp:
				if cursor > 0 {
					next := cursor - 1
					setCursor(next)
					if next < offset {
						setOffset(next)
					}
				}
			case KeyDown:
				if cursor < len(items)-1 {
					next := cursor + 1
					setCursor(next)
					if next >= offset+height {
						setOffset(next - height + 1)
					}
				}
			}
			if ev.Rune == ' ' {
				onToggle(cursor)
			}
		})

		UseClick(func(_, localY int) {
			clickY := localY
			if hasAbove {
				clickY = localY - 1
			}
			if clickY >= 0 && clickY < (end-offset) {
				target := offset + clickY
				if target >= 0 && target < len(items) {
					setCursor(target)
					onToggle(target)
				}
			}
		})

		rows := make([]Element, 0, end-offset+2)
		if hasAbove {
			rows = append(rows, Text("  ↑", theme.Muted))
		}
		for i, item := range items[offset:end] {
			actualIdx := offset + i
			checked := actualIdx < len(selected) && selected[actualIdx]
			glyph := "[ ]"
			if checked {
				glyph = "[x]"
			}
			style := NewStyle()
			if actualIdx == cursor && isFocused {
				style = focusStyle
			}
			rows = append(rows, Text(glyph+" "+item, style))
		}
		if hasBelow {
			rows = append(rows, Text("  ↓", theme.Muted))
		}
		return Box(rows...)
	}
}
