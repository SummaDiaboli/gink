package gink

import "strings"

// focusable pairs a component's tree path with its position and dimensions in
// the virtual render buffer, enabling focus-aware auto-scrolling and mouse
// click hit-testing.
type focusable struct {
	path string
	y    int
	x    int
	w    int // component width; 0 means not yet backfilled
	h    int // component height; 0 means not yet backfilled
}

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
	scope := activeCtx.scope
	state := scope.state
	myIdx := len(state.focusables)
	state.focusables = append(state.focusables, focusable{
		path: scope.path,
		y:    scope.y + scope.offsetY,
		x:    scope.x + scope.offsetX,
	})
	return myIdx == state.focusedIdx
}

// UseFocusWithin returns true when the currently focused component is this
// component or any of its descendants, without registering an extra Tab stop.
//
// It reads the focusable list from the previous render pass, so it is always
// one frame behind on the very first render (invisible in practice). Use it to
// style container elements — borders, panels — based on whether focus is inside:
//
//	isFocused := gink.UseFocusWithin()
//	style := gink.NewStyle()
//	if isFocused {
//	    style = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
//	}
//	return gink.BorderWithTitle("Panel", child, style)
func UseFocusWithin() bool {
	if activeCtx == nil {
		panic("gink: UseFocusWithin called outside of a component render — hooks must be called at the top level of a component function")
	}
	return activeCtx.scope.state.isFocusedWithinPath(activeCtx.scope.path)
}

// isFocusedWithinPath returns true if prevFocusables[focusedIdx] is at or
// within the subtree rooted at path. Used by UseFocusWithin and the reconciler
// cache-invalidation check.
func (s *runtimeState) isFocusedWithinPath(path string) bool {
	if s.focusedIdx >= len(s.prevFocusables) {
		return false
	}
	target := s.prevFocusables[s.focusedIdx].path
	prefix := path + "/"
	return target == path || strings.HasPrefix(target, prefix)
}

// UseFocusBarrier registers the current component as a focus barrier.
// While a barrier is active, Tab and Shift+Tab cycle focus only within this
// component's subtree, and focus is automatically snapped inside when the
// barrier first appears. The barrier is cleared at the start of each render
// pass, so it must be called unconditionally on every render.
//
// Intended for modal dialogs that must trap focus until dismissed:
//
//	func NewModal(...) func() Element {
//	    return func() Element {
//	        gink.UseFocusBarrier()
//	        ...
//	    }
//	}
func UseFocusBarrier() {
	if activeCtx == nil {
		panic("gink: UseFocusBarrier called outside of a component render — hooks must be called at the top level of a component function")
	}
	activeCtx.scope.state.focusBarrier = activeCtx.scope.path
}

// snapFocusToBarrier moves focusedIdx to the first focusable within the
// barrier subtree when focus is currently outside it. Called after each
// render pass when focusBarrier is non-empty.
func (s *runtimeState) snapFocusToBarrier() {
	if s.focusBarrier == "" {
		return
	}
	prefix := s.focusBarrier + "/"
	inBarrier := func(path string) bool {
		return path == s.focusBarrier || strings.HasPrefix(path, prefix)
	}
	if s.focusedIdx < len(s.focusables) && inBarrier(s.focusables[s.focusedIdx].path) {
		return
	}
	for i, f := range s.focusables {
		if inBarrier(f.path) {
			s.focusedIdx = i
			s.focusChanged = true
			return
		}
	}
}

// advanceFocus moves focus forward (delta=1) or backward (delta=-1).
// When a focusBarrier is active it cycles only within that subtree;
// otherwise it wraps around all focusables.
// Called by the runtime on Tab / Shift+Tab — not part of the public API.
func (s *runtimeState) advanceFocus(delta int) {
	if len(s.focusables) == 0 {
		return
	}
	if s.focusBarrier == "" {
		s.focusedIdx = (s.focusedIdx + delta + len(s.focusables)) % len(s.focusables)
		s.focusChanged = true
		return
	}
	prefix := s.focusBarrier + "/"
	var indices []int
	for i, f := range s.focusables {
		if f.path == s.focusBarrier || strings.HasPrefix(f.path, prefix) {
			indices = append(indices, i)
		}
	}
	if len(indices) == 0 {
		return
	}
	cur := -1
	for i, idx := range indices {
		if idx == s.focusedIdx {
			cur = i
			break
		}
	}
	if cur == -1 {
		s.focusedIdx = indices[0]
	} else {
		s.focusedIdx = indices[(cur+delta+len(indices))%len(indices)]
	}
	s.focusChanged = true
}
