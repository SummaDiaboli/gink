package gink

// wheelDelta is the number of rows scrolled per mouse wheel tick.
const wheelDelta = 3

// scrollHandlers is rebuilt on every render pass, in tree order.
// Cleared at the start of each render and refilled as component functions run.
var scrollHandlers []func(delta int) bool

// UseScroll registers a handler that receives scroll events routed through
// [dispatchScroll]. delta is positive for scroll-down, negative for scroll-up.
// The handler returns true if it consumed the event, which prevents the global
// scroll buffer from moving.
func UseScroll(fn func(delta int) bool) {
	if activeCtx == nil {
		panic("gink: UseScroll called outside of a component render — hooks must be called at the top level of a component function")
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

// clampDelta clamps delta to the range [-limit, limit].
// If limit <= 0 the delta is returned unchanged.
func clampDelta(delta, limit int) int {
	if limit <= 0 {
		return delta
	}
	if delta > limit {
		return limit
	}
	if delta < -limit {
		return -limit
	}
	return delta
}

// clampIndex clamps idx to [0, n-1]. Returns 0 when n <= 0.
func clampIndex(idx, n int) int {
	if idx < 0 {
		return 0
	}
	if n > 0 && idx >= n {
		return n - 1
	}
	return idx
}
