package gink

// clickHandler associates a component path with a UseClick callback.
type clickHandler struct {
	path string
	fn   func(int, int)
}

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
	activeCtx.scope.state.clickHandlers = append(activeCtx.scope.state.clickHandlers, clickHandler{path: activeCtx.scope.path, fn: fn})
}

// dispatchClick handles a mouse left-click at screen position (mx, my).
// It finds the innermost focusable whose bounds contain the click, transfers
// focus to it, and fires any registered UseClick handler for that component.
func (s *runtimeState) dispatchClick(mx, my int) {
	vy := my + s.scrollOffset

	// Find the last focusable in tree order (innermost) whose full (x, y, w, h)
	// bounds contain (mx, vy). Taking the last match gives priority to inner
	// components over their ancestors when bounds overlap.
	targetIdx := -1
	for i, f := range s.focusables {
		if f.w == 0 || f.h == 0 {
			continue
		}
		if mx >= f.x && mx < f.x+f.w && vy >= f.y && vy < f.y+f.h {
			targetIdx = i
		}
	}

	if targetIdx < 0 {
		return
	}

	f := s.focusables[targetIdx]
	s.focusedIdx = targetIdx
	s.focusChanged = true

	localX := mx - f.x
	localY := vy - f.y
	for _, ch := range s.clickHandlers {
		if ch.path == f.path {
			ch.fn(localX, localY)
		}
	}
}
