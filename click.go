package gink

// clickHandler associates a component path with a UseClick callback.
type clickHandler struct {
	path string
	fn   func(int, int)
}

// clickHandlers is rebuilt on every render pass, in tree order.
// Cleared at the start of each render and refilled as component functions run.
var clickHandlers []clickHandler

// UseClick registers a handler that is called when this component is clicked.
// x and y are coordinates local to the component's top-left corner.
// The component must also call [UseFocus] to be included in hit-testing.
//
//	gink.UseFocus()
//	gink.UseClick(func(x, y int) {
//	    // handle click at local position (x, y)
//	})
func UseClick(fn func(x, y int)) {
	if activeCtx == nil {
		panic("gink: UseClick called outside of a component render — hooks must be called at the top level of a component function")
	}
	clickHandlers = append(clickHandlers, clickHandler{path: activePath, fn: fn})
}

// dispatchClick handles a mouse left-click at screen position (mx, my).
// It finds the innermost focusable whose bounds contain the click, transfers
// focus to it, and fires any registered UseClick handler for that component.
//
// Hit-testing is two-pass:
//  1. Exact: find the last focusable whose full (x, y, w, h) bounds contain the click.
//  2. Y-only fallback: if nothing matched exactly (e.g. the component's rendered text
//     is narrower than the panel the user perceives as clickable), find the last
//     focusable whose Y range contains the click. This lets users click anywhere in a
//     list row without needing to land precisely on the text.
func dispatchClick(mx, my int) {
	vy := my + scrollOffset

	// Pass 1: exact (x + y) hit-test. Take the last (innermost) match.
	targetIdx := -1
	for i, f := range focusables {
		if f.w == 0 || f.h == 0 {
			continue
		}
		if mx >= f.x && mx < f.x+f.w && vy >= f.y && vy < f.y+f.h {
			targetIdx = i
		}
	}

	// Pass 2: Y-only fallback when nothing matched exactly.
	if targetIdx < 0 {
		for i, f := range focusables {
			if f.h == 0 {
				continue
			}
			if vy >= f.y && vy < f.y+f.h {
				targetIdx = i
			}
		}
	}

	if targetIdx < 0 {
		return
	}

	f := focusables[targetIdx]
	focusedIdx = targetIdx
	focusChanged = true

	localX := mx - f.x
	localY := vy - f.y
	for _, ch := range clickHandlers {
		if ch.path == f.path {
			ch.fn(localX, localY)
		}
	}
}
