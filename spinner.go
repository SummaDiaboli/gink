package gink

import "time"

// spinnerFrames is the braille animation sequence. Each frame is one Unicode character.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner is a component that displays an animated braille spinner cycling at 80ms per frame.
// It uses UseInterval internally and manages its own frame state.
//
// Usage (no parameters needed):
//
//	gink.C(gink.Spinner)
//
// To show a spinner conditionally:
//
//	if loading {
//	    return gink.Row(gink.C(gink.Spinner), gink.Text(" Loading..."))
//	}
func Spinner() Element {
	frame, setFrame := UseState(0)
	UseInterval(80*time.Millisecond, func() {
		setFrame((frame + 1) % len(spinnerFrames))
	})
	return Text(spinnerFrames[frame])
}

// SpinnerWithStyle returns a Spinner component rendered with the given style.
// Because Spinner takes no parameters, SpinnerWithStyle is a factory function
// that closes over the style and returns a component function.
//
//	gink.C(gink.SpinnerWithStyle(gink.NewStyle().Foreground(gink.ColorBrightCyan)))
func SpinnerWithStyle(style Style) func() Element {
	return func() Element {
		frame, setFrame := UseState(0)
		UseInterval(80*time.Millisecond, func() {
			setFrame((frame + 1) % len(spinnerFrames))
		})
		return Text(spinnerFrames[frame], style)
	}
}
