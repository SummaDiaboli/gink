package gink

import "strings"

// MenuItem is a single entry in a [NewMenu].
type MenuItem struct {
	Label    string
	Key      rune
	Disabled bool
}

// NewMenu returns a focusable bordered vertical menu component.
//
// items is the list of menu entries. onSelect is called with the chosen item
// when the user presses Enter or clicks an enabled row. onClose is called when
// the user presses Esc.
//
// Up/Down navigate between enabled items; disabled items are skipped. The
// highlighted row is rendered with the theme's Focused style. Each row shows
// the item label and its Key hint.
//
//	items := []gink.MenuItem{
//	    {Label: "New",  Key: 'n'},
//	    {Label: "Open", Key: 'o'},
//	    {Label: "Quit", Key: 'q', Disabled: true},
//	}
//	gink.C(gink.NewMenu(items, func(item gink.MenuItem) { handle(item) }, func() { close() }))
func NewMenu(items []MenuItem, onSelect func(MenuItem), onClose func(), styles ...Style) func() Element {
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

		cursor, setCursor := UseState(0)
		isFocused := UseFocus()
		theme := UseTheme()

		// Clamp cursor to a valid enabled item.
		cursor = clampToEnabled(items, cursor)

		// prevEnabled / nextEnabled skip over disabled items.
		prevEnabled := func(from int) int {
			for i := from - 1; i >= 0; i-- {
				if !items[i].Disabled {
					return i
				}
			}
			return from
		}
		nextEnabled := func(from int) int {
			for i := from + 1; i < len(items); i++ {
				if !items[i].Disabled {
					return i
				}
			}
			return from
		}

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyUp:
				setCursor(prevEnabled(cursor))
			case KeyDown:
				setCursor(nextEnabled(cursor))
			case KeyEnter:
				if !items[cursor].Disabled {
					onSelect(items[cursor])
				}
			case KeyEscape:
				onClose()
			}
		})

		// Map rendered data rows (after the top border) to item indices.
		UseClick(func(_, localY int) {
			itemIdx := localY - 1 // row 0 is the top border
			if itemIdx >= 0 && itemIdx < len(items) {
				item := items[itemIdx]
				if !item.Disabled {
					setCursor(itemIdx)
					onSelect(item)
				}
			}
		})

		// Compute column width: widest label + key hint ("  x").
		maxLen := 0
		for _, item := range items {
			l := len([]rune(item.Label)) + 4 // "  x " suffix
			if l > maxLen {
				maxLen = l
			}
		}

		top := "┌" + strings.Repeat("─", maxLen+2) + "┐"
		bot := "└" + strings.Repeat("─", maxLen+2) + "┘"

		rows := make([]Element, 0, len(items)+2)
		rows = append(rows, Text(top))

		for i, item := range items {
			hint := string(item.Key)
			label := item.Label
			// Pad so all rows are the same width.
			inner := label + strings.Repeat(" ", maxLen-len([]rune(label))-3) + " " + hint + " "
			line := "│ " + inner + "│"

			style := NewStyle()
			if item.Disabled {
				style = theme.Muted
			} else if i == cursor && isFocused {
				style = focusStyle
			}
			rows = append(rows, Text(line, style))
		}

		rows = append(rows, Text(bot))
		return Box(rows...)
	}
}

// clampToEnabled returns the nearest enabled item index at or after idx.
// Falls back to searching before idx if none found after. Returns idx
// unchanged only when all items are disabled (caller must guard against this).
func clampToEnabled(items []MenuItem, idx int) int {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(items) {
		idx = len(items) - 1
	}
	if !items[idx].Disabled {
		return idx
	}
	for i := idx + 1; i < len(items); i++ {
		if !items[i].Disabled {
			return i
		}
	}
	for i := idx - 1; i >= 0; i-- {
		if !items[i].Disabled {
			return i
		}
	}
	// All items disabled — return idx; Enter/click handlers guard against this.
	return idx
}
