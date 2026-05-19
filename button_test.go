package gink

import "testing"

// TestNewButton_clickActivates verifies that clicking a button fires onPress.
func TestNewButton_clickActivates(t *testing.T) {
	pressed := false
	h := NewHarness(t, func() Element {
		return C(NewButton("OK", func() { pressed = true }))
	})
	defer h.Close()

	h.Click(0, 0)

	if !pressed {
		t.Error("expected button to activate on click")
	}
}

// TestNewButton_clickActivatesWhenAlreadyFocused verifies that clicking a
// button that is already focused also fires onPress.
func TestNewButton_clickActivatesWhenAlreadyFocused(t *testing.T) {
	count := 0
	h := NewHarness(t, func() Element {
		return C(NewButton("GO", func() { count++ }))
	})
	defer h.Close()

	h.Click(0, 0) // focus + press
	h.Click(0, 0) // already focused — should still press

	if count != 2 {
		t.Errorf("expected 2 presses, got %d", count)
	}
}
