package gink

import "testing"

// TestUseKeyboard_firesWhenComponentUnfocused verifies that a UseKeyboard
// handler fires even when the component does not hold focus.
func TestUseKeyboard_firesWhenComponentUnfocused(t *testing.T) {
	var keyboardFired, inputFired bool
	h := NewHarness(t, func() Element {
		return Box(
			C(NewButton("btn", func() {})), // takes focus away from the watcher
			C(func() Element {
				isFocused := UseFocus()
				UseInput(func(ev KeyEvent) {
					if !isFocused {
						return
					}
					inputFired = true
				})
				UseKeyboard(func(ev KeyEvent) {
					keyboardFired = true
				})
				return Text("watcher")
			}),
		)
	})
	defer h.Close()

	h.SendRune('x')

	if !keyboardFired {
		t.Error("UseKeyboard should fire even when component is not focused")
	}
	if inputFired {
		t.Error("UseInput should not fire when component is not focused")
	}
}

// TestUseKeyboard_firesFromNonFocusableComponent verifies that UseKeyboard
// works in a component that never calls UseFocus.
func TestUseKeyboard_firesFromNonFocusableComponent(t *testing.T) {
	fired := false
	h := NewHarness(t, func() Element {
		UseKeyboard(func(ev KeyEvent) {
			if ev.Rune == 'q' {
				fired = true
			}
		})
		return C(NewButton("btn", func() {}))
	})
	defer h.Close()

	h.SendRune('q')

	if !fired {
		t.Error("UseKeyboard should fire from a non-focusable root component")
	}
}

// TestUseKeyboard_receivesSpecialKeys verifies that UseKeyboard fires for
// special keys (arrows etc.), not just printable runes.
func TestUseKeyboard_receivesSpecialKeys(t *testing.T) {
	var got KeyEvent
	h := NewHarness(t, func() Element {
		UseKeyboard(func(ev KeyEvent) {
			got = ev
		})
		return Text("")
	})
	defer h.Close()

	h.SendKey(KeyDown)

	if got.Key != KeyDown {
		t.Errorf("expected KeyDown, got key=%v rune=%v", got.Key, got.Rune)
	}
}

// TestUseKeyboard_multipleHandlersFire verifies that all registered UseKeyboard
// handlers fire, not just the first one.
func TestUseKeyboard_multipleHandlersFire(t *testing.T) {
	count := 0
	h := NewHarness(t, func() Element {
		UseKeyboard(func(KeyEvent) { count++ })
		UseKeyboard(func(KeyEvent) { count++ })
		return Text("")
	})
	defer h.Close()

	h.SendRune('x')

	if count != 2 {
		t.Errorf("expected both UseKeyboard handlers to fire, got count=%d", count)
	}
}

// TestUseKeyboard_firesBeforeUseInput verifies that global keyboard handlers
// run before focus-gated UseInput handlers in the same dispatch cycle.
func TestUseKeyboard_firesBeforeUseInput(t *testing.T) {
	var order []string
	h := NewHarness(t, func() Element {
		UseKeyboard(func(KeyEvent) { order = append(order, "keyboard") })
		return C(func() Element {
			isFocused := UseFocus()
			UseInput(func(ev KeyEvent) {
				if !isFocused {
					return
				}
				order = append(order, "input")
			})
			return Text("focused")
		})
	})
	defer h.Close()

	h.SendRune('x')

	if len(order) != 2 || order[0] != "keyboard" || order[1] != "input" {
		t.Errorf("expected [keyboard input], got %v", order)
	}
}
