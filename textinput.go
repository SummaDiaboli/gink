package gink


// NewInput returns a controlled single-line text input component.
// "Controlled" means the parent component owns the state — NewInput renders
// value and calls onChange with the updated string on each keystroke, but never
// stores the string itself.
//
// The component is focusable: Tab moves focus to/from it. When focused, it
// displays a block cursor and cyan brackets. When not focused, keypresses are ignored.
// Both Backspace keycodes (BS and DEL) are handled for cross-terminal compatibility.
//
//	name, setName := gink.UseState("")
//	gink.C(gink.NewInput(name, setName))
//
// For multiple inputs in a form, wrap each in C() and they will receive focus
// independently via Tab:
//
//	gink.BoxWithGap(1,
//	    gink.Row(gink.Text("Name:  "), gink.C(gink.NewInput(name, setName))),
//	    gink.Row(gink.Text("Email: "), gink.C(gink.NewInput(email, setEmail))),
//	)
func NewInput(value string, onChange func(string)) func() Element {
	return func() Element {
		isFocused := UseFocus()
		clipRead, _ := UseClipboard()

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyBackspace, KeyBackspace2: // handle both BS and DEL variants
				runes := []rune(value)
				if len(runes) > 0 {
					onChange(string(runes[:len(runes)-1]))
				}
			case KeyCtrlV:
				// Collapse newlines — single-line input cannot display them.
				text := normalizeNewlines(clipRead(), " ")
				onChange(value + text)
			case KeyRune:
				onChange(value + string(ev.Rune))
			}
		})

		borderStyle := NewStyle()
		if isFocused {
			borderStyle = NewStyle().Foreground(ColorBrightCyan)
		}

		cursor := ""
		if isFocused {
			cursor = "█"
		}

		return Row(
			Text("[ ", borderStyle),
			Text(value+cursor),
			Text(" ]", borderStyle),
		)
	}
}
