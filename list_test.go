package gink

import (
	"strings"
	"testing"
)

// ── NewList ───────────────────────────────────────────────────────────────────

var listItems = []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo"}

// listHarness creates a Harness whose root component owns the selection via
// UseState, so that calling setSel triggers a proper reconciler re-render.
func listHarness(t *testing.T, initialSel, height int) (*Harness, *int, func(int)) {
	t.Helper()
	var lastSel int
	var setSel func(int)

	h := NewHarness(t, func() Element {
		sel, setSelFn := UseState(initialSel)
		lastSel = sel
		setSel = setSelFn
		return C(NewList(listItems, sel, func(i int) { setSelFn(i) }, height))
	})
	return h, &lastSel, setSel
}

func TestNewList_rendersVisibleItems(t *testing.T) {
	h, _, _ := listHarness(t, 0, 3)
	defer h.Close()

	// With height=3 only the first three items should be visible.
	if !h.Contains("Alpha") || !h.Contains("Bravo") || !h.Contains("Charlie") {
		t.Errorf("expected first three items; lines: %v", h.Lines()[:3])
	}
	if h.Contains("Delta") || h.Contains("Echo") {
		t.Error("items beyond viewport height must not be rendered")
	}
}

func TestNewList_highlightsSelectedItem(t *testing.T) {
	h, _, _ := listHarness(t, 1, 5)
	defer h.Close()

	// Selected item (Bravo, index 1) must have the cursor indicator.
	found := false
	for _, line := range h.Lines() {
		runes := []rune(line)
		if len(runes) > 0 && runes[0] == '▶' && strings.Contains(line, "Bravo") {
			found = true
		}
	}
	if !found {
		t.Errorf("selected item should have cursor indicator; lines: %v", h.Lines()[:5])
	}
}

func TestNewList_downAdvancesSelection(t *testing.T) {
	h, lastSel, _ := listHarness(t, 0, 5)
	defer h.Close()

	h.SendKey(KeyDown)

	if *lastSel != 1 {
		t.Errorf("after Down: sel=%d, want 1", *lastSel)
	}
}

func TestNewList_upDecrementsSelection(t *testing.T) {
	h, lastSel, _ := listHarness(t, 2, 5)
	defer h.Close()

	h.SendKey(KeyUp)

	if *lastSel != 1 {
		t.Errorf("after Up: sel=%d, want 1", *lastSel)
	}
}

func TestNewList_downNoOpAtLastItem(t *testing.T) {
	h, lastSel, _ := listHarness(t, len(listItems)-1, 5)
	defer h.Close()

	h.SendKey(KeyDown)

	if *lastSel != len(listItems)-1 {
		t.Errorf("Down at last item should be no-op; sel=%d", *lastSel)
	}
}

func TestNewList_upNoOpAtFirstItem(t *testing.T) {
	h, lastSel, _ := listHarness(t, 0, 5)
	defer h.Close()

	h.SendKey(KeyUp)

	if *lastSel != 0 {
		t.Errorf("Up at first item should be no-op; sel=%d", *lastSel)
	}
}

func TestNewList_scrollsDownWhenSelectionLeavesViewport(t *testing.T) {
	// height=2: items 0 and 1 visible initially.
	h, _, _ := listHarness(t, 0, 2)
	defer h.Close()

	h.SendKey(KeyDown) // sel=1
	h.SendKey(KeyDown) // sel=2 — must scroll so Charlie is visible

	if !h.Contains("Charlie") {
		t.Errorf("viewport should scroll to keep selection visible; lines: %v", h.Lines()[:2])
	}
	if h.Contains("Alpha") {
		t.Error("Alpha should have scrolled out of view")
	}
}

func TestNewList_scrollsUpWhenSelectionLeavesViewport(t *testing.T) {
	// Start at last item (Echo, index 4), height=2.
	h, _, _ := listHarness(t, 4, 2)
	defer h.Close()

	h.SendKey(KeyUp) // sel=3
	h.SendKey(KeyUp) // sel=2
	h.SendKey(KeyUp) // sel=1
	h.SendKey(KeyUp) // sel=0 — viewport should scroll up to show Alpha

	if !h.Contains("Alpha") {
		t.Errorf("viewport should scroll up to show selected item; lines: %v", h.Lines()[:2])
	}
}

func TestNewList_ignoredWhenUnfocused(t *testing.T) {
	var lastSel int
	h := NewHarness(t, func() Element {
		sel, setSel := UseState(0)
		lastSel = sel
		return Box(
			C(NewButton("Other", func() {})), // takes focus
			C(NewList(listItems, sel, func(i int) { setSel(i) }, 3)),
		)
	})
	defer h.Close()

	h.SendKey(KeyDown)

	if lastSel != 0 {
		t.Errorf("unfocused list should ignore keypresses; sel=%d", lastSel)
	}
}

func TestNewList_customFocusStyle(t *testing.T) {
	custom := NewStyle().Bold().Foreground(ColorBrightGreen)
	h := NewHarness(t, func() Element {
		sel, setSel := UseState(0)
		return C(NewList(listItems, sel, func(i int) { setSel(i) }, 5, custom))
	})
	defer h.Close()

	got := h.CellStyle(0, 0)
	if got != custom.toTcell() {
		t.Errorf("custom focus style not applied: got %v, want %v", got, custom.toTcell())
	}
}

func TestNewList_showsFocusStyleOnSelectedItem(t *testing.T) {
	h, _, _ := listHarness(t, 0, 3)
	defer h.Close()

	// First component is focused — selected item row should carry non-default style.
	got := h.CellStyle(0, 0)
	if got == (Style{}).toTcell() {
		t.Error("focused selected item should have non-default style")
	}
}

// ── NewList vertical scroll indicators ───────────────────────────────────────

// TestNewList_vScrollDownIndicator verifies that a ↓ row appears when items
// extend below the visible viewport.
func TestNewList_vScrollDownIndicator(t *testing.T) {
	// height=3 shows Alpha, Bravo, Charlie; Delta and Echo are below.
	h, _, _ := listHarness(t, 0, 3)
	defer h.Close()

	if !h.Contains("↓") {
		t.Errorf("expected ↓ indicator when items extend below viewport; lines: %v", h.Lines())
	}
}

// TestNewList_vScrollUpIndicator verifies that a ↑ row appears when items
// are hidden above the current viewport.
func TestNewList_vScrollUpIndicator(t *testing.T) {
	// height=3; scroll to sel=3 so Alpha leaves the viewport.
	h, _, _ := listHarness(t, 0, 3)
	defer h.Close()

	h.SendKey(KeyDown)
	h.SendKey(KeyDown)
	h.SendKey(KeyDown)

	if !h.Contains("↑") {
		t.Errorf("expected ↑ indicator when items are hidden above viewport; lines: %v", h.Lines())
	}
}

// TestNewList_vScrollNoIndicatorsWhenAllFit verifies that no indicators appear
// when all items fit within the viewport height.
func TestNewList_vScrollNoIndicatorsWhenAllFit(t *testing.T) {
	// height=5 fits all 5 listItems exactly.
	h, _, _ := listHarness(t, 0, 5)
	defer h.Close()

	if h.Contains("↑") {
		t.Errorf("unexpected ↑ indicator when all items fit; lines: %v", h.Lines())
	}
	if h.Contains("↓") {
		t.Errorf("unexpected ↓ indicator when all items fit; lines: %v", h.Lines())
	}
}

func TestNewList_ClickSelection(t *testing.T) {
	// 1. With height=5 (no up/down indicators, offset=0)
	h, lastSel, _ := listHarness(t, 0, 5)
	defer h.Close()

	// Click on Bravo (index 1), which is at localY=1
	h.Click(0, 1)
	if *lastSel != 1 {
		t.Errorf("expected sel=1 after clicking localY=1 when hasAbove=false; got %d", *lastSel)
	}

	// 2. With height=3, selected=3 (Delta), so offset=1 (items: Bravo, Charlie, Delta), hasAbove=true
	h2, lastSel2, _ := listHarness(t, 3, 3)
	defer h2.Close()

	// Screen:
	// Line 0: "  ↑"
	// Line 1: "  Bravo"
	// Line 2: "  Charlie"
	// Line 3: "▶ Delta"
	// Line 4: "  ↓"

	// Click on Bravo (which is at Line 1 on the screen, index 1 in items)
	h2.Click(0, 1)
	if *lastSel2 != 1 {
		t.Errorf("expected sel=1 after clicking Bravo at localY=1 when hasAbove=true; got %d", *lastSel2)
	}

	// Click on the up indicator (Line 0 on screen), selection should NOT change (should remain 1)
	h2.Click(0, 0)
	if *lastSel2 != 1 {
		t.Errorf("clicking scroll up indicator changed selection to %d", *lastSel2)
	}
}
