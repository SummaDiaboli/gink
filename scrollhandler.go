package gink

// scrollHandlers is rebuilt on every render pass, in tree order.
// Cleared at the start of each render and refilled as component functions run.
var scrollHandlers []func(delta int) bool

// registerScrollHandler registers a handler that receives scroll events routed
// through [dispatchScroll]. delta is positive for scroll-down, negative for
// scroll-up. The handler returns true if it consumed the event, which prevents
// the global scroll buffer from moving.
func registerScrollHandler(fn func(delta int) bool) {
	if activeCtx == nil {
		panic("gink: registerScrollHandler called outside of a component render — hooks must be called at the top level of a component function")
	}
	scrollHandlers = append(scrollHandlers, fn)
}

// dispatchScroll delivers a scroll event to each registered handler in tree
// order and returns true as soon as one handler consumes it. A positive delta
// scrolls down; a negative delta scrolls up.
func dispatchScroll(delta int) bool {
	for _, fn := range scrollHandlers {
		if fn(delta) {
			return true
		}
	}
	return false
}
