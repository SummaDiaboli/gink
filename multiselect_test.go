package gink

import (
	"strings"
	"testing"
)

var multiItems = []string{"Alpha", "Bravo", "Charlie", "Delta", "Echo"}

// multiHarness creates a Harness whose root owns the selection state.
func multiHarness(t *testing.T, initialSel []bool, height int) (*Harness, *[]bool, func(int)) {
	t.Helper()
	var lastSel []bool
	var onToggle func(int)

	h := NewHarness(t, func() Element {
		sel, setSel := UseState(initialSel)
		lastSel = sel
		onToggle = func(i int) {
			next := make([]bool, len(sel))
			copy(next, sel)
			next[i] = !next[i]
			setSel(next)
		}
		return C(NewMultiSelect(multiItems, sel, onToggle, height))
	})
	return h, &lastSel, onToggle
}

func noneSelected(n int) []bool { return make([]bool, n) }

func TestNewMultiSelect_rendersVisibleItems(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 3)
	defer h.Close()

	if !h.Contains("Alpha") || !h.Contains("Bravo") || !h.Contains("Charlie") {
		t.Errorf("expected first three items; lines: %v", h.Lines()[:3])
	}
	if h.Contains("Delta") || h.Contains("Echo") {
		t.Error("items beyond viewport height must not be rendered")
	}
}

func TestNewMultiSelect_rendersUncheckedGlyph(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	for _, line := range h.Lines() {
		for _, item := range multiItems {
			if strings.Contains(line, item) && !strings.Contains(line, "[ ]") && !strings.Contains(line, "[x]") {
				t.Errorf("line missing checkbox glyph: %q", line)
			}
		}
	}
}

func TestNewMultiSelect_rendersCheckedGlyph(t *testing.T) {
	sel := noneSelected(5)
	sel[1] = true // Bravo checked
	h, _, _ := multiHarness(t, sel, 5)
	defer h.Close()

	found := false
	for _, line := range h.Lines() {
		if strings.Contains(line, "[x]") && strings.Contains(line, "Bravo") {
			found = true
		}
	}
	if !found {
		t.Errorf("checked item should render [x]; lines: %v", h.Lines())
	}
}

func TestNewMultiSelect_downAdvancesCursor(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	h.SendKey(KeyDown)

	// After one Down, cursor is on Bravo — Space should toggle Bravo.
	h.SendRune(' ')

	// Bravo (index 1) should now be checked.
	found := false
	for _, line := range h.Lines() {
		if strings.Contains(line, "[x]") && strings.Contains(line, "Bravo") {
			found = true
		}
	}
	if !found {
		t.Error("Down then Space should check Bravo")
	}
}

func TestNewMultiSelect_spaceTogglesCurrentItem(t *testing.T) {
	h, lastSel, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	h.SendRune(' ') // toggle Alpha (cursor starts at 0)

	if !(*lastSel)[0] {
		t.Error("Space should toggle Alpha (index 0)")
	}
}

func TestNewMultiSelect_spaceTogglesOff(t *testing.T) {
	sel := noneSelected(5)
	sel[0] = true
	h, lastSel, _ := multiHarness(t, sel, 5)
	defer h.Close()

	h.SendRune(' ')

	if (*lastSel)[0] {
		t.Error("Space should untoggle already-checked item")
	}
}

func TestNewMultiSelect_upNoOpAtFirstItem(t *testing.T) {
	h, lastSel, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	h.SendKey(KeyUp)
	h.SendRune(' ') // should still toggle index 0

	if !(*lastSel)[0] {
		t.Error("Up at first item should be no-op; cursor should stay at 0")
	}
}

func TestNewMultiSelect_downNoOpAtLastItem(t *testing.T) {
	h, lastSel, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	for i := 0; i < 10; i++ {
		h.SendKey(KeyDown)
	}
	h.SendRune(' ') // should toggle index 4 (Echo)

	if !(*lastSel)[4] {
		t.Error("Down past last item should stop at last; Space should toggle Echo")
	}
}

func TestNewMultiSelect_scrollsDownWhenCursorLeavesViewport(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 2)
	defer h.Close()

	h.SendKey(KeyDown)
	h.SendKey(KeyDown) // cursor=2, should scroll so Charlie is visible

	if !h.Contains("Charlie") {
		t.Errorf("viewport should scroll to keep cursor visible; lines: %v", h.Lines())
	}
	if h.Contains("Alpha") {
		t.Error("Alpha should have scrolled out of view")
	}
}

func TestNewMultiSelect_scrollIndicatorDown(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 3)
	defer h.Close()

	if !h.Contains("↓") {
		t.Errorf("expected ↓ indicator when items extend below viewport; lines: %v", h.Lines())
	}
}

func TestNewMultiSelect_scrollIndicatorUp(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 3)
	defer h.Close()

	h.SendKey(KeyDown)
	h.SendKey(KeyDown)
	h.SendKey(KeyDown)

	if !h.Contains("↑") {
		t.Errorf("expected ↑ indicator when items hidden above; lines: %v", h.Lines())
	}
}

func TestNewMultiSelect_noScrollIndicatorsWhenAllFit(t *testing.T) {
	h, _, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	if h.Contains("↑") || h.Contains("↓") {
		t.Errorf("no scroll indicators expected when all items fit; lines: %v", h.Lines())
	}
}

func TestNewMultiSelect_ignoredWhenUnfocused(t *testing.T) {
	var lastSel []bool
	h := NewHarness(t, func() Element {
		sel, setSel := UseState(noneSelected(5))
		lastSel = sel
		toggle := func(i int) {
			next := make([]bool, len(sel))
			copy(next, sel)
			next[i] = !next[i]
			setSel(next)
		}
		return Box(
			C(NewButton("Other", func() {})),
			C(NewMultiSelect(multiItems, sel, toggle, 5)),
		)
	})
	defer h.Close()

	h.SendRune(' ')

	for i, v := range lastSel {
		if v {
			t.Errorf("unfocused multiselect should ignore keypresses; index %d toggled", i)
		}
	}
}

func TestNewMultiSelect_clickTogglesItem(t *testing.T) {
	h, lastSel, _ := multiHarness(t, noneSelected(5), 5)
	defer h.Close()

	h.Click(0, 1) // click Bravo (row 1)

	if !(*lastSel)[1] {
		t.Error("clicking row 1 should toggle Bravo")
	}
}
