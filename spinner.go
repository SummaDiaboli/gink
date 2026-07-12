package gink

import "time"

// spinnerFrameInterval is the delay between animation frames.
const spinnerFrameInterval = 80 * time.Millisecond

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
// useSpinnerFrame advances the spinner's animation state and returns the
// current frame glyph.
func useSpinnerFrame() string {
	frame, setFrame := UseState(0)
	UseInterval(spinnerFrameInterval, func() {
		setFrame((frame + 1) % len(spinnerFrames))
	})
	return spinnerFrames[frame]
}

func Spinner() Element {
	return Text(useSpinnerFrame())
}

// SpinnerWithStyle returns a Spinner component rendered with the given style.
// Because Spinner takes no parameters, SpinnerWithStyle is a factory function
// that closes over the style and returns a component function.
//
//	gink.C(gink.SpinnerWithStyle(gink.NewStyle().Foreground(gink.ColorBrightCyan)))
func SpinnerWithStyle(style Style) func() Element {
	return func() Element {
		return Text(useSpinnerFrame(), style)
	}
}
