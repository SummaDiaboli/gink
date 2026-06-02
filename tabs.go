package gink

// NewTabs returns a focusable horizontal tab strip.
//
// tabs is the ordered list of tab labels. active is the currently selected tab
// — the parent owns it and receives updates via onChange. Left/Right navigate
// between tabs. Clicking a tab selects it directly. Input is ignored when the
// component is not focused.
//
// The active tab is highlighted with the theme's Focused style (or a custom
// style when provided). All tabs render on a single row separated by │.
//
//	tab, setTab := gink.UseState("Overview")
//	gink.C(gink.NewTabs([]string{"Overview", "Details", "Logs"}, tab, setTab))
func NewTabs(tabs []string, active string, onChange func(string), styles ...Style) func() Element {
	explicitStyle, hasExplicitStyle := optionalStyle(styles)
	return func() Element {
		focusStyle := resolveStyle(explicitStyle, hasExplicitStyle, UseTheme().Focused)

		isFocused := UseFocus()

		// Resolve current index; default to 0 if active is not found.
		idx := findIndex(tabs, active)

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyLeft:
				if idx > 0 {
					onChange(tabs[idx-1])
				}
			case KeyRight:
				if idx < len(tabs)-1 {
					onChange(tabs[idx+1])
				}
			}
		})

		// Build the rendered line to compute per-tab column offsets for click.
		// Each tab renders as " Label " with "│" separators between them.
		offsets := make([]int, len(tabs))
		col := 0
		for i, tab := range tabs {
			offsets[i] = col
			col += 1 + len([]rune(tab)) + 1 // space + label + space
			if i < len(tabs)-1 {
				col++ // separator
			}
		}

		UseClick(func(localX, _ int) {
			// Find which tab was clicked by checking column ranges.
			for i, tab := range tabs {
				start := offsets[i]
				end := start + 1 + len([]rune(tab)) + 1
				if localX >= start && localX < end {
					onChange(tab)
					return
				}
			}
		})

		// Render all tabs as a single Row of Text elements.
		elems := make([]Element, 0, len(tabs)*2-1)
		for i, tab := range tabs {
			style := NewStyle()
			if tab == active {
				if isFocused {
					style = focusStyle
				} else {
					style = NewStyle().Bold()
				}
			}
			elems = append(elems, Text(" "+tab+" ", style))
			if i < len(tabs)-1 {
				elems = append(elems, Text("│"))
			}
		}
		return Row(elems...)
	}
}
