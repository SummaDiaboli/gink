package gink

// NewList returns a focusable, scrollable list component.
//
// items is the full slice of option strings. selected is the index of the
// currently highlighted item — the parent owns this state and updates it via
// onSelect. height is the number of rows visible at once; items beyond that
// scroll into view as the selection moves.
//
// The list renders a ▶ cursor next to the selected item. When focused the
// selected row is rendered bold cyan; when not focused it is rendered bold so
// the selection remains visible without implying interactivity.
//
// The viewport scrolls automatically to keep the selected item visible. If
// selected is changed externally (e.g. reset to 0), the viewport adjusts on
// the next render.
//
//	sel, setSel := gink.UseState(0)
//	gink.C(gink.NewList(items, sel, setSel, 8))
//
// Combine with a detail view for a master-detail layout:
//
//	gink.Row(
//	    gink.C(gink.NewList(items, sel, setSel, 10)),
//	    gink.Text("  "),
//	    gink.Text(items[sel]),
//	)
func NewList(items []string, selected int, onSelect func(int), height int, styles ...Style) func() Element {
	hasExplicitStyle := len(styles) > 0
	explicitStyle := Style{}
	if hasExplicitStyle {
		explicitStyle = styles[0]
	}
	return func() Element {
		theme := UseTheme()
		focusStyle := theme.Focused
		if hasExplicitStyle {
			focusStyle = explicitStyle
		}

		offset, setOffset := UseState(0)
		isFocused := UseFocus()

		// Keep selected visible. Persist via setOffset so the viewport stays
		// stable across renders (e.g. after an external selection change).
		if selected < offset {
			offset = selected
			setOffset(offset)
		} else if height > 0 && selected >= offset+height {
			offset = selected - height + 1
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
			next := selected + move
			if next < 0 {
				next = 0
			} else if next >= len(items) {
				next = len(items) - 1
			}
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
				if selected < len(items)-1 {
					next := selected + 1
					onSelect(next)
					if next >= offset+height {
						setOffset(next - height + 1)
					}
				}
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
					onSelect(target)
				}
			}
		})

		rows := make([]Element, 0, end-offset+2)
		if hasAbove {
			rows = append(rows, Text("  ↑", theme.Muted))
		}
		for i, item := range items[offset:end] {
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
			rows = append(rows, Text(cursor+item, style))
		}
		if hasBelow {
			rows = append(rows, Text("  ↓", theme.Muted))
		}
		return Box(rows...)
	}
}
