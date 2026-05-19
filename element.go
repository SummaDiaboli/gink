package gink

// Element is a lightweight, immutable description of a UI node.
// It is not rendered output — it describes what Gink should render.
// Elements form a tree that the reconciler walks on every render pass.
// Create elements using Text, Box, Row, and C — never construct this struct directly.
type Element struct {
	Type     string
	Props    any
	Children []Element
}

// C wraps a component function into an Element so it can appear in a tree.
// The fn argument must be a component function: a func that takes no arguments
// and returns an Element. It may use hooks.
//
//	gink.C(MyComponent)
//	gink.C(NewButton("Submit", onSubmit))  // parameterized component factory
func C(fn func() Element) Element {
	return Element{Type: "component", Props: fn}
}

// Text returns an Element that renders a single line of text.
// An optional Style controls color, weight, and decoration.
//
//	gink.Text("Hello")
//	gink.Text("Error", gink.NewStyle().Foreground(gink.ColorBrightRed).Bold())
func Text(s string, styles ...Style) Element {
	var style Style
	if len(styles) > 0 {
		style = styles[0]
	}
	return Element{Type: "text", Props: TextProps{Content: s, Style: style}}
}

// Box stacks children vertically (top to bottom), one per line.
//
//	gink.Box(
//	    gink.Text("Line 1"),
//	    gink.Text("Line 2"),
//	)
func Box(children ...Element) Element {
	return Element{Type: "box", Props: BoxProps{Direction: DirectionColumn}, Children: children}
}

// Row lays children out horizontally (left to right) on the same line.
//
//	gink.Row(
//	    gink.Text("Label: "),
//	    gink.Text(value),
//	)
func Row(children ...Element) Element {
	return Element{Type: "box", Props: BoxProps{Direction: DirectionRow}, Children: children}
}

// BoxWithGap stacks children vertically with gap empty lines between each child.
//
//	gink.BoxWithGap(1,
//	    gink.Text("Section A"),
//	    gink.Text("Section B"),  // separated by 1 blank line
//	)
func BoxWithGap(gap int, children ...Element) Element {
	return Element{Type: "box", Props: BoxProps{Direction: DirectionColumn, Gap: gap}, Children: children}
}

// RowWithGap lays children out horizontally with gap empty columns between each child.
//
//	gink.RowWithGap(3,
//	    gink.Text("Alpha"),
//	    gink.Text("Beta"),  // separated by 3 spaces
//	)
func RowWithGap(gap int, children ...Element) Element {
	return Element{Type: "box", Props: BoxProps{Direction: DirectionRow, Gap: gap}, Children: children}
}
