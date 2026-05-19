package gink

// Badge renders an inline label enclosed in brackets, suitable for status
// indicators, category tags, and counts.
//
// An optional [Style] controls color and decoration of the entire badge:
//
//	gink.Badge("OK", gink.NewStyle().Foreground(gink.ColorBrightGreen))
//	gink.Badge("ERR", gink.NewStyle().Background(gink.ColorRed).Foreground(gink.ColorWhite))
//
// Badges are pure elements — no hooks, no state — so they can appear anywhere
// in a layout tree without needing [C].
func Badge(label string, styles ...Style) Element {
	return Text("[ "+label+" ]", styles...)
}
