package gink

// accessibilityHints maps component path to its accessibility label for the
// current render pass. Cleared before each render and rebuilt by UseAccessibility.
var accessibilityHints = map[string]string{}

// currentAccessibilityLabel holds the label of the focused component after the
// most recent render pass. Empty if the focused component did not register one.
var currentAccessibilityLabel string

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
	accessibilityHints[activePath] = label
}

// AccessibilityLabel returns the accessibility label of the focused component
// from the most recent render pass, or an empty string if none was registered.
// Primarily useful in tests:
//
//	if h.AccessibilityLabel() != "Save button" {
//	    t.Error("unexpected label")
//	}
func AccessibilityLabel() string {
	return currentAccessibilityLabel
}

// resolveAccessibilityLabel sets currentAccessibilityLabel from the hints map
// using the focused component's path. Called after each render pass.
func resolveAccessibilityLabel() {
	currentAccessibilityLabel = ""
	if focusedIdx < len(focusables) {
		path := focusables[focusedIdx].path
		if label, ok := accessibilityHints[path]; ok {
			currentAccessibilityLabel = label
		}
	}
}
