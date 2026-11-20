package gink

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// TestingT is the subset of [*testing.T] that [Harness] requires.
// [*testing.T] satisfies this interface automatically — no additional
// import is needed beyond the standard `import "testing"`.
// The interface exists so that the gink package itself does not import
// the testing package, keeping it out of production binaries.
type TestingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

// Harness renders a component tree into a tcell SimulationScreen so tests
// can inspect output and simulate input without a real terminal.
// Construct one with [NewHarness] or [NewHarnessSize]; close it with [Harness.Close].
//
// In most cases, prefer the [github.com/SummaDiaboli/gink/ginktest] package, which
// wraps Harness with assertion helpers ([ginktest.AssertContains] etc.).
// Use Harness directly only when you need lower-level access such as
// [Harness.Line], [Harness.CellStyle], or custom screen dimensions.
type Harness struct {
	rec    *Reconciler
	screen tcell.SimulationScreen
	r      *renderer
	root   Component
	Width  int
	Height int
	buf    *Buffer
	t      TestingT
}

// NewHarness creates a Harness with an 80×24 simulation screen and renders
// root once. It is the standard entry point for component tests.
//
//	h := gink.NewHarness(t, MyComponent)
//	defer h.Close()
func NewHarness(t TestingT, root Component) *Harness {
	t.Helper()
	return NewHarnessSize(t, root, 80, 24)
}

// NewHarnessSize creates a Harness with custom terminal dimensions.
// Use this when component layout depends on the terminal size
// (e.g. components that call [UseTermSize]).
func NewHarnessSize(t TestingT, root Component, width, height int) *Harness {
	t.Helper()

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("harness: screen init: %v", err)
	}
	screen.SetSize(width, height)

	r := &renderer{
		screen: screen,
		dirty:  make(chan struct{}, 1),
	}

	h := &Harness{
		rec:    NewReconciler(r),
		screen: screen,
		r:      r,
		root:   root,
		Width:  width,
		Height: height,
		t:      t,
	}
	h.Render()
	return h
}

// Render executes one full render pass and updates the internal buffer.
// Input methods ([Tab], [Harness.SendRune], [Harness.Enter], etc.) call
// Render automatically, so you only need to call it explicitly when waiting
// for async state updates — e.g. polling between [time.Sleep] calls in
// timing-sensitive tests. For most cases prefer [ginktest.AwaitContains].
func (h *Harness) Render() {
	h.t.Helper()
	commitRender(h.r, h.rec, h.root, h.Width, h.Height, func() {
		renderPass(h.r, h.rec, h.root, h.Width, h.Height)
	})
	h.buf = h.r.current
}

// doScroll dispatches a scroll event: focused handlers get first pick;
// only when none consume it does the global buffer move.
func (h *Harness) doScroll(delta int) {
	h.t.Helper()
	h.r.state.scrollBy(delta)
	h.Render()
}

// PageDown scrolls the viewport down by one screen height and re-renders.
// A focused scrollable component receives the event first; only when none
// consumes it does the global scroll buffer move.
func (h *Harness) PageDown() { h.t.Helper(); h.doScroll(h.Height) }

// PageUp scrolls the viewport up by one screen height and re-renders.
// A focused scrollable component receives the event first; only when none
// consumes it does the global scroll buffer move.
func (h *Harness) PageUp() { h.t.Helper(); h.doScroll(-h.Height) }

// WheelDown simulates a mouse wheel-down event and re-renders.
// A focused scrollable component receives the event first; only when none
// consumes it does the global scroll buffer move.
func (h *Harness) WheelDown() { h.t.Helper(); h.doScroll(wheelDelta) }

// WheelUp simulates a mouse wheel-up event and re-renders.
// A focused scrollable component receives the event first; only when none
// consumes it does the global scroll buffer move.
func (h *Harness) WheelUp() { h.t.Helper(); h.doScroll(-wheelDelta) }

// Lines returns the buffer contents as a slice of strings, one per terminal
// row, with trailing spaces trimmed. The slice length equals Harness.Height.
func (h *Harness) Lines() []string {
	rows := make([]string, h.buf.Height)
	for y, row := range h.buf.Cells {
		runes := make([]rune, h.buf.Width)
		for x, cell := range row {
			runes[x] = cell.Rune
		}
		rows[y] = strings.TrimRight(string(runes), " ")
	}
	return rows
}

// Line returns the trimmed string content of row y (0-indexed from the top).
// Returns an empty string if y is out of bounds.
func (h *Harness) Line(y int) string {
	lines := h.Lines()
	if y < 0 || y >= len(lines) {
		return ""
	}
	return lines[y]
}

// Contains returns true if any line in the buffer contains the given substring.
func (h *Harness) Contains(s string) bool {
	for _, line := range h.Lines() {
		if strings.Contains(line, s) {
			return true
		}
	}
	return false
}

// CellStyle returns the tcell.Style of the cell at column x, row y.
// Returns tcell.StyleDefault if the coordinates are out of bounds.
// Use this to assert on colors, bold, or underline when text content
// alone is not sufficient to verify correct rendering.
func (h *Harness) CellStyle(x, y int) tcell.Style {
	if y < 0 || y >= h.buf.Height || x < 0 || x >= h.buf.Width {
		return tcell.StyleDefault
	}
	return h.buf.Cells[y][x].Style
}

// SendRune simulates a printable character keypress and re-renders.
// The event is dispatched to whichever component's [UseInput] handler
// is active for the current focus state.
func (h *Harness) SendRune(r rune) {
	h.t.Helper()
	ev := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
	h.r.state.dispatchKey(ev)
	h.Render()
}

// SendKey simulates a special key event and re-renders.
// Use [Harness.Tab], [Harness.Enter], and [Harness.Backspace] for the most
// common keys. Use SendKey directly for arrow keys and other special keys:
//
//	h.SendKey(gink.KeyDown)
func (h *Harness) SendKey(key tcell.Key) {
	h.t.Helper()
	ev := tcell.NewEventKey(key, 0, tcell.ModNone)
	h.r.state.dispatchKey(ev)
	h.Render()
}

// Tab advances focus to the next [UseFocus] component in tree order and re-renders.
// Focus wraps around from the last focusable back to the first.
func (h *Harness) Tab() {
	h.t.Helper()
	h.r.state.advanceFocus(1)
	h.Render()
}

// ShiftTab moves focus to the previous [UseFocus] component in tree order and re-renders.
// Focus wraps around from the first focusable back to the last.
func (h *Harness) ShiftTab() {
	h.t.Helper()
	h.r.state.advanceFocus(-1)
	h.Render()
}

// Enter simulates the Enter key and re-renders. Equivalent to [Harness.SendKey](gink.KeyEnter).
func (h *Harness) Enter() {
	h.t.Helper()
	h.SendKey(tcell.KeyEnter)
}

// Backspace simulates a Backspace keypress and re-renders.
// Sends the DEL keycode (0x7F) which most terminals use for Backspace.
func (h *Harness) Backspace() {
	h.t.Helper()
	h.SendKey(tcell.KeyBackspace2)
}

// Click simulates a mouse left-click at screen position (x, y) and re-renders.
// It transfers focus to the focusable component at that position (if any) and
// fires any [UseClick] handler registered by that component.
func (h *Harness) Click(x, y int) {
	h.t.Helper()
	h.r.state.dispatchClick(x, y)
	h.Render()
}

// AccessibilityLabel returns the accessibility label registered by the currently
// focused component via [UseAccessibility], or an empty string if none was set.
// The label is resolved after each render, so it reflects the state after the
// most recent interaction or [Harness.Render] call.
func (h *Harness) AccessibilityLabel() string {
	return h.r.state.currentAccessibilityLabel
}

// Close releases the simulation screen. Call it with defer immediately after
// [NewHarness] to ensure the screen is freed even if the test panics.
func (h *Harness) Close() {
	h.screen.Fini()
}
