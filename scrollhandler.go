package gink

// wheelDelta is the number of rows scrolled per mouse wheel tick.
const wheelDelta = 3

// UseScroll registers a handler that receives scroll events routed through
// [dispatchScroll]. delta is positive for scroll-down, negative for scroll-up.
// The handler returns true if it consumed the event, which prevents the global
// scroll buffer from moving.
func UseScroll(fn func(delta int) bool) {
	if activeCtx == nil {
		panic("gink: UseScroll called outside of a component render — hooks must be called at the top level of a component function")
	}
	activeCtx.scope.state.scrollHandlers = append(activeCtx.scope.state.scrollHandlers, fn)
}

// dispatchScroll delivers a scroll event to each registered handler in tree
// order and returns true as soon as one handler consumes it. A positive delta
// scrolls down; a negative delta scrolls up.
func (s *runtimeState) dispatchScroll(delta int) bool {
	for _, fn := range s.scrollHandlers {
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

// clampCursor clamps cursor to [0, n-1] (0 when n <= 0), persisting the
// change via setCursor only when it actually moved.
func clampCursor(cursor, n int, setCursor func(int)) int {
	if cursor >= n {
		cursor = n - 1
		setCursor(cursor)
	}
	if cursor < 0 {
		cursor = 0
		setCursor(cursor)
	}
	return cursor
}

// viewportOffset returns the viewport start offset that keeps idx visible
// within a page of size height. Returns offset unchanged when idx is already
// in view.
func viewportOffset(idx, offset, height int) int {
	if idx < offset {
		return idx
	}
	if height > 0 && idx >= offset+height {
		return idx - height + 1
	}
	return offset
}

// findIndex returns the index of the first element in items equal to value,
// or 0 if value is not found.
func findIndex(items []string, value string) int {
	for i, s := range items {
		if s == value {
			return i
		}
	}
	return 0
}
