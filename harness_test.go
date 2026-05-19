package gink

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// Harness renders a component tree into a tcell SimulationScreen so tests
// can inspect output and simulate input without a real terminal.
type Harness struct {
	rec    *Reconciler
	screen tcell.SimulationScreen
	r      *renderer
	root   Component
	Width  int
	Height int
	buf    *Buffer
	t      *testing.T
}

// NewHarness creates a Harness with an 80×24 simulation screen and renders the
// root component once. Call Render() to re-render after simulating input or
// state changes triggered by effects.
func NewHarness(t *testing.T, root Component) *Harness {
	t.Helper()
	return NewHarnessSize(t, root, 80, 24)
}

// NewHarnessSize creates a Harness with custom terminal dimensions.
func NewHarnessSize(t *testing.T, root Component, width, height int) *Harness {
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

	// Reset all global render state so tests are isolated from each other.
	focusedIdx = 0
	activeCtx = nil
	activePath = ""

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

// Render executes one full render pass: clears hook/effect/focus slices,
// walks the element tree, flushes the buffer, and runs pending effects.
func (h *Harness) Render() {
	h.t.Helper()
	inputHandlers = inputHandlers[:0]
	pendingEffects = pendingEffects[:0]
	focusables = focusables[:0]
	currentTermSize = TermSize{Width: h.Width, Height: h.Height}
	h.buf = h.rec.Render(C(h.root), h.Width, h.Height)
	h.r.flush(h.buf)
	runEffects()
	if len(focusables) > 0 && focusedIdx >= len(focusables) {
		focusedIdx = len(focusables) - 1
		// Focus was clamped — re-render once so components see the corrected index.
		inputHandlers = inputHandlers[:0]
		pendingEffects = pendingEffects[:0]
		focusables = focusables[:0]
		h.buf = h.rec.Render(C(h.root), h.Width, h.Height)
		h.r.flush(h.buf)
		runEffects()
	}
}

// Lines returns the buffer contents as a slice of strings, one per row,
// with trailing spaces trimmed.
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

// Line returns the trimmed string content of row y.
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

// CellStyle returns the tcell.Style of the cell at (x, y).
func (h *Harness) CellStyle(x, y int) tcell.Style {
	if y < 0 || y >= h.buf.Height || x < 0 || x >= h.buf.Width {
		return tcell.StyleDefault
	}
	return h.buf.Cells[y][x].Style
}

// SendRune simulates a printable character keypress and re-renders.
func (h *Harness) SendRune(r rune) {
	h.t.Helper()
	ev := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
	dispatchKey(ev)
	h.Render()
}

// SendKey simulates a special key event and re-renders.
func (h *Harness) SendKey(key tcell.Key) {
	h.t.Helper()
	ev := tcell.NewEventKey(key, 0, tcell.ModNone)
	dispatchKey(ev)
	h.Render()
}

// Tab advances focus to the next focusable component and re-renders.
func (h *Harness) Tab() {
	h.t.Helper()
	advanceFocus(1)
	h.Render()
}

// ShiftTab moves focus to the previous focusable component and re-renders.
func (h *Harness) ShiftTab() {
	h.t.Helper()
	advanceFocus(-1)
	h.Render()
}

// Enter simulates the Enter key and re-renders.
func (h *Harness) Enter() {
	h.t.Helper()
	h.SendKey(tcell.KeyEnter)
}

// Backspace simulates the Backspace key (BS variant) and re-renders.
func (h *Harness) Backspace() {
	h.t.Helper()
	h.SendKey(tcell.KeyBackspace2)
}

// Close releases the simulation screen.
func (h *Harness) Close() {
	h.screen.Fini()
}
