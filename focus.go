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

// focusables is rebuilt on every render pass, in tree order.
// focusedIdx is the index of the currently focused component within that list.
var focusables []focusable
var focusedIdx int

// prevFocusables is a snapshot of focusables from the previous render pass.
// UseFocusWithin reads it so it can check descendant focus before children
// have rendered in the current pass.
var prevFocusables []focusable

// activePath, activeY, and activeX are set by the reconciler before calling
// each component function so UseFocus can register the correct path and position.
var activePath string
var activeY int
var activeX int

// renderOffsetX and renderOffsetY accumulate the translation from sub-buffer
// rendering (Constrain, Width, Height, Size). When a container renders its
// child into a temporary buffer at (0,0), it pushes (x,y) onto the offsets so
// UseFocus still records the correct absolute screen position.
var renderOffsetX, renderOffsetY int

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
	focusables = append(focusables, focusable{
		path: activePath,
		y:    activeY + renderOffsetY,
		x:    activeX + renderOffsetX,
	})
	return myIdx == focusedIdx
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
	return isFocusedWithinPath(activePath)
}

// isFocusedWithinPath returns true if prevFocusables[focusedIdx] is at or
// within the subtree rooted at path. Used by UseFocusWithin and the reconciler
// cache-invalidation check.
func isFocusedWithinPath(path string) bool {
	if focusedIdx >= len(prevFocusables) {
		return false
	}
	target := prevFocusables[focusedIdx].path
	prefix := path + "/"
	return target == path || strings.HasPrefix(target, prefix)
}

// focusChanged is set when Tab/Shift+Tab changes the focused component so
// the next render knows to auto-scroll to the new focus position. It is
// cleared after the scroll check so manual scrolling is never overridden.
var focusChanged bool

// focusBarrier is the tree path of the active focus barrier (e.g. a modal).
// When non-empty, Tab/Shift+Tab cycle only within that subtree. Reset to ""
// at the start of each render pass and re-registered by the component that
// owns the barrier.
var focusBarrier string

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
	focusBarrier = activePath
}

// snapFocusToBarrier moves focusedIdx to the first focusable within the
// barrier subtree when focus is currently outside it. Called after each
// render pass when focusBarrier is non-empty.
func snapFocusToBarrier() {
	if focusBarrier == "" {
		return
	}
	prefix := focusBarrier + "/"
	inBarrier := func(path string) bool {
		return path == focusBarrier || strings.HasPrefix(path, prefix)
	}
	if focusedIdx < len(focusables) && inBarrier(focusables[focusedIdx].path) {
		return // already inside
	}
	for i, f := range focusables {
		if inBarrier(f.path) {
			focusedIdx = i
			focusChanged = true
			return
		}
	}
}

// advanceFocus moves focus forward (delta=1) or backward (delta=-1).
// When a focusBarrier is active it cycles only within that subtree;
// otherwise it wraps around all focusables.
// Called by the runtime on Tab / Shift+Tab — not part of the public API.
func advanceFocus(delta int) {
	if len(focusables) == 0 {
		return
	}
	if focusBarrier == "" {
		focusedIdx = (focusedIdx + delta + len(focusables)) % len(focusables)
		focusChanged = true
		return
	}
	prefix := focusBarrier + "/"
	var indices []int
	for i, f := range focusables {
		if f.path == focusBarrier || strings.HasPrefix(f.path, prefix) {
			indices = append(indices, i)
		}
	}
	if len(indices) == 0 {
		return
	}
	cur := -1
	for i, idx := range indices {
		if idx == focusedIdx {
			cur = i
			break
		}
	}
	if cur == -1 {
		focusedIdx = indices[0]
	} else {
		focusedIdx = indices[(cur+delta+len(indices))%len(indices)]
	}
	focusChanged = true
}
