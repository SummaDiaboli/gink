package gink

// ScrollViewProps is the internal props passed from the NewScrollView component
// to the reconciler's "scrollview" element handler.
type ScrollViewProps struct {
	Height int
	Offset int
	Child  Element
}

// NewScrollView returns a focusable component that clips its child to a fixed
// viewport height and lets the user scroll the content with Up/Down.
//
// height is the number of rows the viewport occupies in the layout. child is
// any Element — wrap multiple children in [Box] when needed.
//
// When focused, Up scrolls up one row and Down scrolls down one row. Scroll
// indicators (↑ / ↓) appear at the right edge of the viewport when content
// exists above or below the visible area.
//
//	gink.C(gink.NewScrollView(8, gink.Box(rows...)))
func NewScrollView(height int, child Element) func() Element {
	return func() Element {
		offset, setOffset := UseState(0)
		isFocused := UseFocus()

		registerScrollHandler(func(delta int) bool {
			if !isFocused {
				return false
			}
			newOffset := offset + delta
			if newOffset < 0 {
				newOffset = 0
			}
			setOffset(newOffset) // reconciler clamps at content bottom
			return true
		})

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			switch ev.Key {
			case KeyUp:
				if offset > 0 {
					setOffset(offset - 1)
				}
			case KeyDown:
				setOffset(offset + 1) // reconciler clamps at content bottom
			}
		})

		return Element{
			Type:  "scrollview",
			Props: ScrollViewProps{Height: height, Offset: offset, Child: child},
		}
	}
}
