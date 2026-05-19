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
func NewList(items []string, selected int, onSelect func(int), height int) func() Element {
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
				if selected < len(items)-1 {
					next := selected + 1
					onSelect(next)
					if next >= offset+height {
						setOffset(next - height + 1)
					}
				}
			}
		})

		end := offset + height
		if end > len(items) {
			end = len(items)
		}

		rows := make([]Element, end-offset)
		for i, item := range items[offset:end] {
			actualIdx := offset + i
			cursor := "  "
			style := NewStyle()
			if actualIdx == selected {
				cursor = "▶ "
				if isFocused {
					style = NewStyle().Bold().Foreground(ColorBrightCyan)
				} else {
					style = NewStyle().Bold()
				}
			}
			rows[i] = Text(cursor+item, style)
		}
		return Box(rows...)
	}
}
