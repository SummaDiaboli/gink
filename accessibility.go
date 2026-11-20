package gink

import "sync/atomic"

var lastAccessibilityLabel atomic.Value

// UseAccessibility registers a plain-text description for the current component.
// After each render the runtime surfaces the focused component's label as the
// terminal window title, giving screen readers and automation tools a concise
// description of what is currently active.
//
//	gink.UseAccessibility("Server list — use Up/Down to navigate, Enter to select")
//
// The label is also readable in tests via [Harness.AccessibilityLabel] and the
// package-level [AccessibilityLabel] function, so component behaviour can be
// verified without inspecting screen pixels.
func UseAccessibility(label string) {
	if activeCtx == nil {
		panic("gink: UseAccessibility called outside of a component render — hooks must be called at the top level of a component function")
	}
	activeCtx.scope.state.accessibilityHints[activeCtx.scope.path] = label
}

// AccessibilityLabel returns the accessibility label of the focused component
// from the most recent render pass, or an empty string if none was registered.
// Primarily useful in tests:
//
//	if h.AccessibilityLabel() != "Save button" {
//	    t.Error("unexpected label")
//	}
func AccessibilityLabel() string {
	if v := lastAccessibilityLabel.Load(); v != nil {
		return v.(string)
	}
	return ""
}

func (s *runtimeState) resolveAccessibilityLabel() string {
	s.currentAccessibilityLabel = ""
	if s.focusedIdx < len(s.focusables) {
		path := s.focusables[s.focusedIdx].path
		if label, ok := s.accessibilityHints[path]; ok {
			s.currentAccessibilityLabel = label
		}
	}
	return s.currentAccessibilityLabel
}
