package gink

// NewSelect returns a controlled single-line option picker component.
// "Controlled" means the parent owns the selected value — NewSelect renders
// value and calls onChange immediately when the user moves Up or Down.
//
// When focused, ◀ and ▶ arrows flank the current value to indicate it is
// navigable. When not focused, only the value is shown in brackets.
// Keypresses are ignored when the component is not focused.
//
// If value is not found in options the component behaves as if the first
// option is selected, so Down moves to options[1].
//
//	lang, setLang := gink.UseState("Go")
//	gink.C(gink.NewSelect([]string{"Go", "Rust", "Zig"}, lang, setLang))
//
// For a form with multiple fields, each NewSelect wrapped in C() receives
// focus independently via Tab:
//
//	gink.Box(
//	    gink.Row(gink.Text("Language: "), gink.C(gink.NewSelect(langs, lang, setLang))),
//	    gink.Row(gink.Text("Theme:    "), gink.C(gink.NewSelect(themes, theme, setTheme))),
//	)
func NewSelect(options []string, value string, onChange func(string), styles ...Style) func() Element {
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

		// Resolve the current index; default to 0 if value is not in options.
		idx := 0
		for i, opt := range options {
			if opt == value {
				idx = i
				break
			}
		}

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyLeft:
				if idx > 0 {
					onChange(options[idx-1])
				}
			case KeyRight:
				if idx < len(options)-1 {
					onChange(options[idx+1])
				}
			}
		})

		// Divide the widget at its midpoint: left half = previous, right half = next.
		// Width differs by focus state: focused adds "◀ " and " ▶" (4 extra chars).
		UseClick(func(localX, _ int) {
			var mid int
			if isFocused {
				mid = (len([]rune(value)) + 8) / 2
			} else {
				mid = (len([]rune(value)) + 4) / 2
			}
			if localX < mid {
				if idx > 0 {
					onChange(options[idx-1])
				}
			} else {
				if idx < len(options)-1 {
					onChange(options[idx+1])
				}
			}
		})

		if !isFocused {
			return Text("[ " + value + " ]")
		}
		return Text("[ ◀ "+value+" ▶ ]", focusStyle)
	}
}
