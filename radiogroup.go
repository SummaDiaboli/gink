package gink

// NewRadioGroup returns a focusable vertical radio-button group.
//
// options is the full list of choices. selected is the currently active value
// — the parent owns it and receives updates via onChange. Up/Down navigate;
// the selection changes immediately on movement (no separate confirm step).
//
// Each option renders as "( ) Label" with "(●) Label" for the selected item.
// When focused the selected row is highlighted with the theme's Focused style
// (or a custom style when provided). Keypresses are ignored when the component
// is not focused.
//
//	color, setColor := gink.UseState("Red")
//	gink.C(gink.NewRadioGroup([]string{"Red", "Green", "Blue"}, color, setColor))
func NewRadioGroup(options []string, selected string, onChange func(string), styles ...Style) func() Element {
	explicitStyle, hasExplicitStyle := optionalStyle(styles)
	return func() Element {
		focusStyle := resolveStyle(explicitStyle, hasExplicitStyle, UseTheme().Focused)

		isFocused := UseFocus()

		// Resolve current index; default to 0 if selected is not found.
		idx := findIndex(options, selected)

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyUp:
				if idx > 0 {
					onChange(options[idx-1])
				}
			case KeyDown:
				if idx < len(options)-1 {
					onChange(options[idx+1])
				}
			}
		})

		UseClick(func(_, localY int) {
			if localY >= 0 && localY < len(options) {
				onChange(options[localY])
			}
		})

		rows := make([]Element, len(options))
		for i, opt := range options {
			glyph := "( )"
			if opt == selected {
				glyph = "(●)"
			}
			rows[i] = Text(glyph+" "+opt, itemStyle(opt == selected, isFocused, focusStyle))
		}
		return Box(rows...)
	}
}
