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

		isFocused := UseFocus()

		// Resolve current index; default to 0 if selected is not found.
		idx := 0
		for i, opt := range options {
			if opt == selected {
				idx = i
				break
			}
		}

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
			style := NewStyle()
			if opt == selected {
				glyph = "(●)"
				if isFocused {
					style = focusStyle
				} else {
					style = NewStyle().Bold()
				}
			}
			rows[i] = Text(glyph+" "+opt, style)
		}
		return Box(rows...)
	}
}
