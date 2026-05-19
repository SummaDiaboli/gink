package gink

// ConstrainProps holds the min/max size constraints for a constrain element.
// Zero means no constraint on that dimension.
type ConstrainProps struct {
	MinWidth, MaxWidth, MinHeight, MaxHeight int
}

// Constrain wraps child with optional min/max size constraints.
// Pass 0 for any value to leave that dimension unconstrained.
// The constrained size is reported to the parent, so siblings in a [Row] or
// [Box] are positioned correctly.
//
//	// Reserve at least 20 columns for a label, clip at 40.
//	gink.Constrain(label, 20, 40, 0, 0)
func Constrain(child Element, minW, maxW, minH, maxH int) Element {
	return Element{
		Type:     "constrain",
		Props:    ConstrainProps{minW, maxW, minH, maxH},
		Children: []Element{child},
	}
}

// MinWidth wraps child so it reports at least n columns wide to its parent,
// pushing siblings in a [Row] to the right without clipping content.
//
//	gink.Row(gink.MinWidth(20, label), value)
func MinWidth(n int, child Element) Element { return Constrain(child, n, 0, 0, 0) }

// MaxWidth wraps child so content beyond n columns is not rendered.
//
//	gink.MaxWidth(40, longText)
func MaxWidth(n int, child Element) Element { return Constrain(child, 0, n, 0, 0) }

// MinHeight wraps child so it reports at least n rows tall to its parent,
// pushing siblings in a [Box] downward without adding visible content.
//
//	gink.Box(gink.MinHeight(3, header), body)
func MinHeight(n int, child Element) Element { return Constrain(child, 0, 0, n, 0) }

// MaxHeight wraps child so content beyond n rows is not rendered.
//
//	gink.MaxHeight(5, scrollableContent)
func MaxHeight(n int, child Element) Element { return Constrain(child, 0, 0, 0, n) }
