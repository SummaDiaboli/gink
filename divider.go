package gink

type DividerProps struct {
	Label string
	Style Style
}

// Divider is a component that renders a full-width horizontal rule using ─
// characters. The width adapts to the available layout width.
//
//	gink.C(gink.Divider)
func Divider() Element {
	return Element{Type: "divider", Props: DividerProps{}}
}

// DividerStyled returns a Divider component rendered with the given style.
// Use this to colour the rule or make it bold.
//
//	gink.C(gink.DividerStyled(gink.NewStyle().Foreground(gink.ColorBrightBlack)))
func DividerStyled(style Style) func() Element {
	return func() Element {
		return Element{Type: "divider", Props: DividerProps{Style: style}}
	}
}

// DividerWithLabel returns a Divider with a centered label inset into the
// line, e.g. ───── Label ─────. An optional Style applies to the full line.
//
//	gink.C(gink.DividerWithLabel("Section"))
//	gink.C(gink.DividerWithLabel("Section", gink.NewStyle().Bold()))
func DividerWithLabel(label string, styles ...Style) func() Element {
	return func() Element {
		style, _ := optionalStyle(styles)
		return Element{
			Type:  "divider",
			Props: DividerProps{Label: " " + label + " ", Style: style},
		}
	}
}
