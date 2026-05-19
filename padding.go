package gink

// Pad specifies inner spacing around a child element on each side, in cells.
type Pad struct {
	Top, Right, Bottom, Left int
}

// Padding wraps child with inner spacing on each side as specified by p.
// The child is offset by (Left, Top) cells; the total element size grows by
// the padding on each side.
//
//	gink.Padding(gink.Pad{Top: 1, Left: 2}, content)
func Padding(p Pad, child Element) Element {
	return Element{Type: "padding", Props: p, Children: []Element{child}}
}

// PaddingAll wraps child with equal spacing on all four sides.
//
//	gink.PaddingAll(1, content)
func PaddingAll(n int, child Element) Element {
	return Padding(Pad{Top: n, Right: n, Bottom: n, Left: n}, child)
}

// PaddingXY wraps child with horizontal spacing x (left and right) and
// vertical spacing y (top and bottom).
//
//	gink.PaddingXY(2, 1, content)
func PaddingXY(x, y int, child Element) Element {
	return Padding(Pad{Top: y, Right: x, Bottom: y, Left: x}, child)
}
