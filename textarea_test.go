package gink_test

import (
	"testing"

	"github.com/SummaDiaboli/gink"
)

// TestNewTextArea_rendersValueAsLines verifies that a multi-line value is
// split across rows in the viewport.
func TestNewTextArea_rendersValueAsLines(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		val, _ := gink.UseState("line one\nline two")
		return gink.C(gink.NewTextArea(val, func(string) {}, 3))
	})
	defer h.Close()

	if !h.Contains("line one") {
		t.Error("expected 'line one' to be visible")
	}
	if !h.Contains("line two") {
		t.Error("expected 'line two' to be visible")
	}
}

// TestNewTextArea_typingInsertsCharacter verifies that a printable keypress
// inserts the character at the cursor (initially col 0 of line 0).
func TestNewTextArea_typingInsertsCharacter(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('h')
	h.SendRune('i')

	if !h.Contains("hi") {
		t.Errorf("expected 'hi' after typing, screen: %v", h.Lines())
	}
}

// TestNewTextArea_backspaceRemovesCharacter verifies Backspace deletes the
// character before the cursor.
func TestNewTextArea_backspaceRemovesCharacter(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('h')
	h.SendRune('i')
	h.Backspace()

	if !h.Contains("h") {
		t.Error("expected 'h' after backspace")
	}
	if h.Contains("hi") {
		t.Error("expected 'i' to be deleted")
	}
}

// TestNewTextArea_enterInsertsNewline verifies Enter splits the current line.
func TestNewTextArea_enterInsertsNewline(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('a')
	h.Enter()
	h.SendRune('b')

	if !h.Contains("a") {
		t.Error("expected 'a' on first line")
	}
	if !h.Contains("b") {
		t.Error("expected 'b' on second line")
	}
	// They must be on different rows.
	aRow, bRow := -1, -1
	for i, l := range h.Lines() {
		if l == "a" {
			aRow = i
		}
		if l == "b" {
			bRow = i
		}
	}
	if aRow == -1 || bRow == -1 || aRow == bRow {
		t.Errorf("expected 'a' and 'b' on different rows, got aRow=%d bRow=%d", aRow, bRow)
	}
}

// TestNewTextArea_rightMovesAndTypingInserts verifies that pressing Right moves
// the cursor and the next typed character is inserted at the new position.
func TestNewTextArea_rightMovesAndTypingInserts(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('a')
	h.SendRune('b') // value = "ab", cursor at col 2
	h.SendKey(gink.KeyLeft)
	h.SendKey(gink.KeyLeft) // cursor back at col 0
	h.SendKey(gink.KeyRight) // cursor at col 1
	h.SendRune('x')          // insert between 'a' and 'b' → "axb"

	if !h.Contains("axb") {
		t.Errorf("expected 'axb' after insert, screen: %v", h.Lines())
	}
}

// TestNewTextArea_homeMovesToLineStart verifies the Home key places the cursor
// at column 0 so the next typed character is prepended to the line.
func TestNewTextArea_homeMovesToLineStart(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('a')
	h.SendRune('b')         // "ab", cursor col 2
	h.SendKey(gink.KeyHome) // cursor col 0
	h.SendRune('x')         // "xab"

	if !h.Contains("xab") {
		t.Errorf("expected 'xab', screen: %v", h.Lines())
	}
}

// TestNewTextArea_endMovesToLineEnd verifies the End key places the cursor
// after the last character so typing appends to the line.
func TestNewTextArea_endMovesToLineEnd(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('a')
	h.SendRune('b')           // "ab", cursor col 2
	h.SendKey(gink.KeyHome)   // cursor col 0
	h.SendKey(gink.KeyEnd)    // cursor back at col 2
	h.SendRune('x')           // "abx"

	if !h.Contains("abx") {
		t.Errorf("expected 'abx', screen: %v", h.Lines())
	}
}

// TestNewTextArea_downMovesToNextLine verifies Down arrow moves to the next line.
func TestNewTextArea_downMovesToNextLine(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 3))
	})
	defer h.Close()

	h.SendRune('a')
	h.Enter()               // "a\n", cursor on line 1
	h.SendRune('b')         // "a\nb"
	h.SendKey(gink.KeyUp)   // back to line 0, col 1
	h.SendKey(gink.KeyDown) // line 1 col 0 (col clamped since "b" is at col 1 now)
	h.SendRune('x')         // insert at line 1 col 1 → "a\nbx"

	if !h.Contains("bx") {
		t.Errorf("expected 'bx' on second line, screen: %v", h.Lines())
	}
}

// TestNewTextArea_viewportScrollsWithCursor verifies that the viewport scrolls
// when the cursor moves below the visible area.
func TestNewTextArea_viewportScrollsWithCursor(t *testing.T) {
	val, setVal := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv func(string)
		val, sv = gink.UseState(val)
		setVal = sv
		return gink.C(gink.NewTextArea(val, setVal, 2))
	})
	defer h.Close()

	// Build 4 lines by pressing Enter.
	for _, ch := range "line0" {
		h.SendRune(ch)
	}
	h.Enter()
	for _, ch := range "line1" {
		h.SendRune(ch)
	}
	h.Enter()
	for _, ch := range "line2" {
		h.SendRune(ch)
	}

	// Viewport is 2 rows; cursor is on line 2. "line2" must be visible.
	if !h.Contains("line2") {
		t.Errorf("expected 'line2' to scroll into view, screen: %v", h.Lines())
	}
	// "line0" should have scrolled out.
	if h.Contains("line0") {
		t.Error("expected 'line0' to scroll out of view")
	}
}

// TestNewTextArea_ignoresInputWhenUnfocused verifies that a TextArea that does
// not hold focus does not process keypresses.
func TestNewTextArea_ignoresInputWhenUnfocused(t *testing.T) {
	val1, setVal1 := "", func(string) {}
	val2, setVal2 := "", func(string) {}
	h := gink.NewHarness(t, func() gink.Element {
		var sv1, sv2 func(string)
		val1, sv1 = gink.UseState(val1)
		val2, sv2 = gink.UseState(val2)
		setVal1, setVal2 = sv1, sv2
		return gink.Box(
			gink.C(gink.NewTextArea(val1, setVal1, 2)),
			gink.C(gink.NewTextArea(val2, setVal2, 2)),
		)
	})
	defer h.Close()

	// First textarea is focused; typing should only go to val1.
	h.SendRune('x')

	lines := h.Lines()
	// val1 on rows 0-1, val2 on rows 2-3.
	foundInFirst := lines[0] == "x" || lines[1] == "x"
	foundInSecond := lines[2] == "x" || lines[3] == "x"

	if !foundInFirst {
		t.Error("expected 'x' in first textarea (focused)")
	}
	if foundInSecond {
		t.Error("expected second textarea to ignore keypresses (unfocused)")
	}
}

// TestNewTextArea_reportsFixedHeight verifies that NewTextArea always occupies
// exactly height rows in the layout, so siblings are positioned correctly.
func TestNewTextArea_reportsFixedHeight(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		val, _ := gink.UseState("")
		return gink.Box(
			gink.C(gink.NewTextArea(val, func(string) {}, 3)),
			gink.Text("after"),
		)
	})
	defer h.Close()

	if h.Line(3) != "after" {
		t.Errorf("expected 'after' at row 3, got %q", h.Line(3))
	}
}
