package gink

// focusable pairs a component's tree path with its Y position in the virtual
// render buffer, enabling focus-aware auto-scrolling.
type focusable struct {
	path string
	y    int
}

// focusables is rebuilt on every render pass, in tree order.
// focusedIdx is the index of the currently focused component within that list.
var focusables []focusable
var focusedIdx int

// activePath and activeY are set by the reconciler before calling each
// component function so UseFocus can register the correct path and position.
var activePath string
var activeY int

// UseFocus registers the current component as a focusable element and returns
// whether it currently holds focus.
//
// Focusable components are cycled in tree order (the order their UseFocus calls
// execute during a render pass). Tab moves focus forward; Shift+Tab moves it back.
// The first focusable component in the tree receives focus on startup.
//
// Use the returned bool to gate input handling and change visual appearance:
//
//	isFocused := gink.UseFocus()
//
//	gink.UseInput(func(ev gink.KeyEvent) {
//	    if !isFocused {
//	        return  // ignore keypresses when not focused
//	    }
//	    if ev.Key == gink.KeyEnter {
//	        submit()
//	    }
//	})
//
//	style := gink.NewStyle()
//	if isFocused {
//	    style = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
//	}
//
// Note: UseFocus does not consume a hook slot. It appends to a render-scoped
// slice rather than the component's ordered slot array, so its call position
// relative to UseState/UseEffect/UseRef does not matter.
func UseFocus() bool {
	if activeCtx == nil {
		panic("gink: UseFocus called outside of a component render — hooks must be called at the top level of a component function")
	}
	myIdx := len(focusables)
	focusables = append(focusables, focusable{path: activePath, y: activeY})
	return myIdx == focusedIdx
}

// focusChanged is set when Tab/Shift+Tab changes the focused component so
// the next render knows to auto-scroll to the new focus position. It is
// cleared after the scroll check so manual scrolling is never overridden.
var focusChanged bool

// advanceFocus moves focus forward (delta=1) or backward (delta=-1), wrapping around.
// Called by the runtime on Tab / Shift+Tab — not part of the public API.
func advanceFocus(delta int) {
	if len(focusables) == 0 {
		return
	}
	focusedIdx = (focusedIdx + delta + len(focusables)) % len(focusables)
	focusChanged = true
}
