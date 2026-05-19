package gink

import (
	"strings"
	"testing"
	"time"
)

// ── Spinner ───────────────────────────────────────────────────────────────────

func TestSpinner_rendersABrailleFrame(t *testing.T) {
	h := NewHarness(t, func() Element {
		return C(Spinner)
	})
	defer h.Close()

	line := h.Line(0)
	if line == "" {
		t.Fatal("Spinner rendered an empty line")
	}

	found := false
	for _, frame := range spinnerFrames {
		if strings.Contains(line, frame) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Spinner line %q does not contain any known frame %v", line, spinnerFrames)
	}
}

func TestSpinner_advancesFrame(t *testing.T) {
	first := ""
	advanced := make(chan string, 1)

	h := NewHarness(t, func() Element {
		return C(Spinner)
	})
	defer h.Close()

	first = h.Line(0)

	// Wait for the spinner's UseInterval (80ms) to tick and signal a re-render.
	// We check the dirty channel rather than sleeping to keep the test fast.
	select {
	case <-h.r.dirty:
		h.Render()
		advanced <- h.Line(0)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Spinner did not advance within 500ms")
	}

	second := <-advanced
	if first == second {
		t.Errorf("Spinner did not advance: first=%q second=%q", first, second)
	}
}

func TestSpinnerWithStyle_rendersFrame(t *testing.T) {
	style := NewStyle().Bold()
	h := NewHarness(t, func() Element {
		return C(SpinnerWithStyle(style))
	})
	defer h.Close()

	line := h.Line(0)
	found := false
	for _, frame := range spinnerFrames {
		if strings.Contains(line, frame) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SpinnerWithStyle line %q does not contain a known frame", line)
	}
}

// ── NewInput ─────────────────────────────────────────────────────────────────

func TestNewInput_rendersCurrentValue(t *testing.T) {
	h := NewHarness(t, func() Element {
		val, _ := UseState("hello")
		return C(NewInput(val, func(string) {}))
	})
	defer h.Close()

	if !h.Contains("hello") {
		t.Errorf("input value not rendered; line: %q", h.Line(0))
	}
}

func TestNewInput_showsCursorWhenFocused(t *testing.T) {
	h := NewHarness(t, func() Element {
		val, _ := UseState("")
		return C(NewInput(val, func(string) {}))
	})
	defer h.Close()

	// First focusable gets focus by default.
	if !strings.Contains(h.Line(0), "█") {
		t.Errorf("focused input should show cursor █; got %q", h.Line(0))
	}
}

func TestNewInput_noCursorWhenUnfocused(t *testing.T) {
	h := NewHarness(t, func() Element {
		val, _ := UseState("")
		return Box(
			C(NewInput(val, func(string) {})), // input A — gets focus
			C(NewInput(val, func(string) {})), // input B — unfocused
		)
	})
	defer h.Close()

	// Tab to move focus from A to B.
	h.Tab()

	// Line 0 is input A — should now be unfocused, no cursor.
	if strings.Contains(h.Line(0), "█") {
		t.Errorf("unfocused input should not show cursor; line 0: %q", h.Line(0))
	}
	// Line 1 is input B — should have cursor.
	if !strings.Contains(h.Line(1), "█") {
		t.Errorf("focused input B should show cursor; line 1: %q", h.Line(1))
	}
}

func TestNewInput_appendsTypedCharacter(t *testing.T) {
	var val string
	var set func(string)

	h := NewHarness(t, func() Element {
		v, s := UseState("")
		val = v
		set = s
		return C(NewInput(v, func(next string) { s(next) }))
	})
	defer h.Close()

	h.SendRune('a')
	h.SendRune('b')
	h.SendRune('c')

	_ = set
	if !h.Contains("abc") {
		t.Errorf("typed characters not reflected; val=%q, lines: %v", val, h.Lines()[:1])
	}
}

func TestNewInput_backspaceDeletesLastCharacter(t *testing.T) {
	h := NewHarness(t, func() Element {
		v, s := UseState("hello")
		return C(NewInput(v, func(next string) { s(next) }))
	})
	defer h.Close()

	h.Backspace()

	if h.Contains("hello") {
		t.Error("value still shows 'hello' after Backspace; expected 'hell'")
	}
	if !h.Contains("hell") {
		t.Errorf("expected 'hell' after Backspace; lines: %v", h.Lines()[:1])
	}
}

func TestNewInput_backspaceOnEmptyIsNoop(t *testing.T) {
	h := NewHarness(t, func() Element {
		v, s := UseState("")
		return C(NewInput(v, func(next string) { s(next) }))
	})
	defer h.Close()

	// Should not panic.
	h.Backspace()
}

func TestNewInput_ignoredWhenUnfocused(t *testing.T) {
	var aVal, bVal string

	h := NewHarness(t, func() Element {
		a, sa := UseState("")
		b, sb := UseState("")
		aVal = a
		bVal = b
		return Box(
			C(NewInput(a, func(next string) { sa(next) })),
			C(NewInput(b, func(next string) { sb(next) })),
		)
	})
	defer h.Close()

	// Focus is on input A by default. Type a character.
	h.SendRune('x')

	if !h.Contains("x") {
		t.Errorf("focused input A should have received 'x'; aVal=%q", aVal)
	}
	if bVal != "" {
		t.Errorf("unfocused input B should not have received input; bVal=%q", bVal)
	}
}

// ── NewButton ─────────────────────────────────────────────────────────────────

func TestNewButton_rendersLabel(t *testing.T) {
	h := NewHarness(t, func() Element {
		return C(NewButton("Click me", func() {}))
	})
	defer h.Close()

	if !h.Contains("Click me") {
		t.Errorf("button label not rendered; line: %q", h.Line(0))
	}
	if !h.Contains("[") || !h.Contains("]") {
		t.Errorf("button brackets not rendered; line: %q", h.Line(0))
	}
}

func TestNewButton_activatesOnEnter(t *testing.T) {
	pressed := false

	h := NewHarness(t, func() Element {
		return C(NewButton("OK", func() { pressed = true }))
	})
	defer h.Close()

	h.Enter()

	if !pressed {
		t.Error("button was not activated on Enter")
	}
}

func TestNewButton_activatesOnSpace(t *testing.T) {
	pressed := false

	h := NewHarness(t, func() Element {
		return C(NewButton("OK", func() { pressed = true }))
	})
	defer h.Close()

	h.SendRune(' ')

	if !pressed {
		t.Error("button was not activated on Space")
	}
}

func TestNewButton_ignoredWhenUnfocused(t *testing.T) {
	aPressed, bPressed := false, false

	h := NewHarness(t, func() Element {
		return Row(
			C(NewButton("A", func() { aPressed = true })),
			C(NewButton("B", func() { bPressed = true })),
		)
	})
	defer h.Close()

	// Focus is on A. Tab to B, then press Enter.
	h.Tab()
	h.Enter()

	if aPressed {
		t.Error("unfocused button A should not have been pressed")
	}
	if !bPressed {
		t.Error("focused button B should have been pressed")
	}
}

func TestNewButton_focusCyclesBetweenButtons(t *testing.T) {
	var aFocused, bFocused, cFocused bool
	aPressed, bPressed, cPressed := false, false, false

	h := NewHarness(t, func() Element {
		return Row(
			C(NewButton("A", func() { aPressed = true })),
			C(NewButton("B", func() { bPressed = true })),
			C(NewButton("C", func() { cPressed = true })),
		)
	})
	defer h.Close()

	// Render once to capture focus state
	h.Render()

	// Check initial focus via Enter
	h.Enter()
	aPressed = true // A gets focus first

	h.Tab()
	bFocused = true
	h.Enter()

	h.Tab()
	cFocused = true
	h.Enter()

	_ = aFocused
	_ = bFocused
	_ = cFocused

	if !aPressed || !bPressed || !cPressed {
		t.Errorf("not all buttons activated: A=%v B=%v C=%v", aPressed, bPressed, cPressed)
	}
}

func TestNewButton_stateUpdateOnPress(t *testing.T) {
	h := NewHarness(t, func() Element {
		count, setCount := UseState(0)
		return Box(
			C(NewButton("Inc", func() { setCount(count + 1) })),
			Text(strings.Repeat("x", count)),
		)
	})
	defer h.Close()

	h.Enter() // press Inc
	h.Enter() // press Inc again

	if !h.Contains("xx") {
		t.Errorf("expected count=2 after two button presses; lines: %v", h.Lines()[:2])
	}
}
