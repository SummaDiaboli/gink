package gink

// NewCheckbox returns a focusable checkbox component.
//
// checked is the current state — the parent owns it and receives updates via
// onChange. Space or Enter toggles the value; clicking the component also
// toggles it.
//
// When focused the label is rendered with the theme's Focused style (or a
// custom style when provided). The glyph is "[ ]" when unchecked and "[x]"
// when checked.
//
//	checked, setChecked := gink.UseState(false)
//	gink.C(gink.NewCheckbox("Enable notifications", checked, func(v bool) { setChecked(v) }))
func NewCheckbox(label string, checked bool, onChange func(bool), styles ...Style) func() Element {
	explicitStyle, hasExplicitStyle := optionalStyle(styles)
	return func() Element {
		focusStyle := resolveStyle(explicitStyle, hasExplicitStyle, UseTheme().Focused)

		isFocused := UseFocus()

		toggle := func() { onChange(!checked) }

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			if ev.Key == KeyEnter || ev.Rune == ' ' {
				toggle()
			}
		})

		UseClick(func(_, _ int) {
			toggle()
		})

		glyph := "[ ]"
		if checked {
			glyph = "[x]"
		}

		style := NewStyle()
		if isFocused {
			style = focusStyle
		}

		return Text(glyph+" "+label, style)
	}
}
