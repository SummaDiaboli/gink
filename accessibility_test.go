package gink

import "testing"

// TestUseAccessibility_labelForFocusedComponent verifies that the label
// registered by UseAccessibility is returned for the focused component.
func TestUseAccessibility_labelForFocusedComponent(t *testing.T) {
	h := NewHarness(t, func() Element {
		UseAccessibility("save button")
		UseFocus()
		return Text("[ Save ]")
	})
	defer h.Close()

	if h.AccessibilityLabel() != "save button" {
		t.Errorf("got %q, want %q", h.AccessibilityLabel(), "save button")
	}
}

// TestUseAccessibility_labelFollowsFocus verifies that the returned label
// updates to match whichever component currently holds focus.
func TestUseAccessibility_labelFollowsFocus(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(
			C(func() Element {
				UseAccessibility("first")
				UseFocus()
				return Text("[ A ]")
			}),
			C(func() Element {
				UseAccessibility("second")
				UseFocus()
				return Text("[ B ]")
			}),
		)
	})
	defer h.Close()

	if h.AccessibilityLabel() != "first" {
		t.Errorf("initial: got %q, want %q", h.AccessibilityLabel(), "first")
	}

	h.Tab()

	if h.AccessibilityLabel() != "second" {
		t.Errorf("after Tab: got %q, want %q", h.AccessibilityLabel(), "second")
	}
}

// TestUseAccessibility_emptyWhenNotSet verifies that the label is empty when
// the focused component did not call UseAccessibility.
func TestUseAccessibility_emptyWhenNotSet(t *testing.T) {
	h := NewHarness(t, func() Element {
		UseFocus()
		return Text("no hint")
	})
	defer h.Close()

	if h.AccessibilityLabel() != "" {
		t.Errorf("got %q, want empty when UseAccessibility is not called", h.AccessibilityLabel())
	}
}

// TestUseAccessibility_onlyFocusedComponentLabel verifies that the label of
// an unfocused component is not returned.
func TestUseAccessibility_onlyFocusedComponentLabel(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(
			C(func() Element {
				UseAccessibility("focused label")
				UseFocus()
				return Text("[ A ]")
			}),
			C(func() Element {
				UseAccessibility("unfocused label")
				UseFocus()
				return Text("[ B ]")
			}),
		)
	})
	defer h.Close()

	// Focus starts on the first component.
	if h.AccessibilityLabel() != "focused label" {
		t.Errorf("got %q, want %q", h.AccessibilityLabel(), "focused label")
	}
}
