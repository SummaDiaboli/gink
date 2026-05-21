package gink

import (
	"strings"
	"testing"
)

// TestUseKeybinding_dispatchesOnRune verifies that a rune binding fires its
// handler when the matching key is pressed.
func TestUseKeybinding_dispatchesOnRune(t *testing.T) {
	fired := false
	h := NewHarness(t, func() Element {
		UseKeybinding(Binding{Key: KeyRune, Rune: 'n', Label: "n", Description: "New"}, func() {
			fired = true
		})
		return Text("")
	})
	defer h.Close()

	h.SendRune('n')

	if !fired {
		t.Error("UseKeybinding handler should fire when matching rune is pressed")
	}
}

// TestUseKeybinding_dispatchesOnSpecialKey verifies that a special-key binding
// fires its handler when the matching key is pressed.
func TestUseKeybinding_dispatchesOnSpecialKey(t *testing.T) {
	fired := false
	h := NewHarness(t, func() Element {
		UseKeybinding(Binding{Key: KeyDown, Label: "↓", Description: "Down"}, func() {
			fired = true
		})
		return Text("")
	})
	defer h.Close()

	h.SendKey(KeyDown)

	if !fired {
		t.Error("UseKeybinding handler should fire when matching special key is pressed")
	}
}

// TestUseKeybinding_doesNotFireOnMismatch verifies that a binding's handler is
// not called when a different key is pressed.
func TestUseKeybinding_doesNotFireOnMismatch(t *testing.T) {
	fired := false
	h := NewHarness(t, func() Element {
		UseKeybinding(Binding{Key: KeyRune, Rune: 'n', Label: "n", Description: "New"}, func() {
			fired = true
		})
		return Text("")
	})
	defer h.Close()

	h.SendRune('x')

	if fired {
		t.Error("UseKeybinding handler should not fire for a different key")
	}
}

// TestUseKeybinding_multipleBindings verifies that only the matching binding's
// handler fires when multiple bindings are registered.
func TestUseKeybinding_multipleBindings(t *testing.T) {
	var aFired, bFired bool
	h := NewHarness(t, func() Element {
		UseKeybinding(Binding{Key: KeyRune, Rune: 'a', Label: "a", Description: "Alpha"}, func() {
			aFired = true
		})
		UseKeybinding(Binding{Key: KeyRune, Rune: 'b', Label: "b", Description: "Beta"}, func() {
			bFired = true
		})
		return Text("")
	})
	defer h.Close()

	h.SendRune('b')

	if !bFired {
		t.Error("binding 'b' should fire")
	}
	if aFired {
		t.Error("binding 'a' should not fire when 'b' is pressed")
	}
}

// TestKeybindingHelp_showsAllBindings verifies that KeybindingHelp renders the
// label and description for every registered binding after a render pass.
func TestKeybindingHelp_showsAllBindings(t *testing.T) {
	h := NewHarness(t, func() Element {
		UseKeybinding(Binding{Key: KeyRune, Rune: '?', Label: "?", Description: "Show help"}, func() {})
		UseKeybinding(Binding{Key: KeyRune, Rune: 'q', Label: "q", Description: "Quit"}, func() {})
		return Box(Text("app"), KeybindingHelp())
	})
	defer h.Close()

	lines := strings.Join(h.Lines(), "\n")
	for _, want := range []string{"?", "Show help", "q", "Quit"} {
		if !strings.Contains(lines, want) {
			t.Errorf("KeybindingHelp output missing %q", want)
		}
	}
}
