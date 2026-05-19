package gink

// NewButton returns a focusable button component that activates on Enter or Space.
// The button renders as "[ label ]" and changes to bold cyan when focused.
//
// onPress is called synchronously in the input dispatch phase — it is safe to
// call state setters from it, which will schedule a re-render.
//
//	gink.C(gink.NewButton("Submit", func() { submit() }))
//
// Multiple buttons in a row with spacing:
//
//	gink.RowWithGap(2,
//	    gink.C(gink.NewButton("Save",   func() { save() })),
//	    gink.C(gink.NewButton("Cancel", func() { cancel() })),
//	)
//
// Tab moves focus between buttons (and other focusable components) automatically.
func NewButton(label string, onPress func()) func() Element {
	return func() Element {
		isFocused := UseFocus()

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			if ev.Key == KeyEnter || ev.Rune == ' ' {
				onPress()
			}
		})

		UseClick(func(_, _ int) {
			onPress()
		})

		style := NewStyle()
		if isFocused {
			style = NewStyle().Bold().Foreground(ColorBrightCyan)
		}

		return Text("[ "+label+" ]", style)
	}
}
