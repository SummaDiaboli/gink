package gink

import "strings"

// Divider is a component that renders a full-width horizontal rule using ─
// characters. The width adapts automatically to the terminal via UseTermSize.
//
//	gink.C(gink.Divider)
func Divider() Element {
	size := UseTermSize()
	return Text(strings.Repeat("─", size.Width))
}

// DividerStyled returns a Divider component rendered with the given style.
// Use this to colour the rule or make it bold.
//
//	gink.C(gink.DividerStyled(gink.NewStyle().Foreground(gink.ColorBrightBlack)))
func DividerStyled(style Style) func() Element {
	return func() Element {
		size := UseTermSize()
		return Text(strings.Repeat("─", size.Width), style)
	}
}

// DividerWithLabel returns a Divider with a centered label inset into the
// line, e.g. ───── Label ─────. An optional Style applies to the full line.
//
//	gink.C(gink.DividerWithLabel("Section"))
//	gink.C(gink.DividerWithLabel("Section", gink.NewStyle().Bold()))
func DividerWithLabel(label string, styles ...Style) func() Element {
	return func() Element {
		size := UseTermSize()
		var style Style
		if len(styles) > 0 {
			style = styles[0]
		}
		inner := " " + label + " "
		dashes := size.Width - len([]rune(inner))
		if dashes < 2 {
			// Terminal too narrow to fit the label with dashes — just show the label.
			return Text(inner, style)
		}
		left := dashes / 2
		right := dashes - left
		return Text(strings.Repeat("─", left)+inner+strings.Repeat("─", right), style)
	}
}
