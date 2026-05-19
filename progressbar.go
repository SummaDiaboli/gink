package gink

import (
	"fmt"
	"strings"
)

// ProgressBar renders a horizontal progress indicator.
//
// value is clamped to [0.0, 1.0]. width is the number of fill characters
// inside the brackets — the total rendered width is width+2 (brackets) plus
// the percentage label. Pass a width of 10–20 for typical use:
//
//	gink.ProgressBar(0.5, 20)          // [██████████░░░░░░░░░░] 50%
//	gink.ProgressBar(done/total, 16, barStyle)
//
// An optional [Style] controls color and decoration of the entire bar,
// including brackets and percentage text.
func ProgressBar(value float64, width int, styles ...Style) Element {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	filled := int(value * float64(width))
	empty := width - filled

	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "] " +
		fmt.Sprintf("%d%%", int(value*100))

	return Text(bar, styles...)
}
