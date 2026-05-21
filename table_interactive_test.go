package gink

import (
	"strings"
	"testing"
)

// ── NewTable ──────────────────────────────────────────────────────────────────

var tableCols = []Column{
	{Header: "Name", MinWidth: 8},
	{Header: "Status"},
	{Header: "Region", MaxWidth: 6},
}

var tableRows = [][]string{
	{"alpha", "ONLINE", "us-east-1"},
	{"bravo", "WARN", "eu-west-1"},
	{"charlie", "OFFLINE", "us-west-2"},
	{"delta", "ONLINE", "ap-south-1"},
	{"echo", "ONLINE", "us-east-1"},
}

// tableHarness creates a Harness whose root owns selection via UseState.
func tableHarness(t *testing.T, initialSel, height int) (*Harness, *int) {
	t.Helper()
	var lastSel int
	h := NewHarness(t, func() Element {
		sel, setSel := UseState(initialSel)
		lastSel = sel
		return C(NewTable(tableCols, tableRows, sel, func(i int) { setSel(i) }, height))
	})
	return h, &lastSel
}

func TestNewTable_rendersHeaders(t *testing.T) {
	h, _ := tableHarness(t, 0, 5)
	defer h.Close()

	if !h.Contains("Name") || !h.Contains("Status") || !h.Contains("Region") {
		t.Errorf("column headers must be visible; lines: %v", h.Lines())
	}
}

func TestNewTable_rendersVisibleRows(t *testing.T) {
	h, _ := tableHarness(t, 0, 3)
	defer h.Close()

	// height=3: only the first three data rows should be visible.
	if !h.Contains("alpha") || !h.Contains("bravo") || !h.Contains("charlie") {
		t.Errorf("expected first three rows; lines: %v", h.Lines())
	}
	if h.Contains("delta") || h.Contains("echo") {
		t.Error("rows beyond viewport height must not be rendered")
	}
}

func TestNewTable_highlightsSelectedRow(t *testing.T) {
	h, _ := tableHarness(t, 1, 5)
	defer h.Close()

	// Row 1 (bravo) is selected — it must carry the cursor indicator.
	found := false
	for _, line := range h.Lines() {
		if strings.Contains(line, "▶") && strings.Contains(line, "bravo") {
			found = true
		}
	}
	if !found {
		t.Errorf("selected row should have cursor indicator; lines: %v", h.Lines())
	}
}

func TestNewTable_downAdvancesSelection(t *testing.T) {
	h, lastSel := tableHarness(t, 0, 5)
	defer h.Close()

	h.SendKey(KeyDown)

	if *lastSel != 1 {
		t.Errorf("after Down: sel=%d, want 1", *lastSel)
	}
}

func TestNewTable_upDecrementsSelection(t *testing.T) {
	h, lastSel := tableHarness(t, 2, 5)
	defer h.Close()

	h.SendKey(KeyUp)

	if *lastSel != 1 {
		t.Errorf("after Up: sel=%d, want 1", *lastSel)
	}
}

func TestNewTable_downNoOpAtLastRow(t *testing.T) {
	h, lastSel := tableHarness(t, len(tableRows)-1, 5)
	defer h.Close()

	h.SendKey(KeyDown)

	if *lastSel != len(tableRows)-1 {
		t.Errorf("Down at last row should be no-op; sel=%d", *lastSel)
	}
}

func TestNewTable_upNoOpAtFirstRow(t *testing.T) {
	h, lastSel := tableHarness(t, 0, 5)
	defer h.Close()

	h.SendKey(KeyUp)

	if *lastSel != 0 {
		t.Errorf("Up at first row should be no-op; sel=%d", *lastSel)
	}
}

func TestNewTable_scrollsDownWhenSelectionLeavesViewport(t *testing.T) {
	h, _ := tableHarness(t, 0, 2)
	defer h.Close()

	h.SendKey(KeyDown) // sel=1
	h.SendKey(KeyDown) // sel=2 — charlie must scroll into view

	if !h.Contains("charlie") {
		t.Errorf("viewport should scroll to keep selection visible; lines: %v", h.Lines())
	}
	if h.Contains("alpha") {
		t.Error("alpha should have scrolled out of view")
	}
}

func TestNewTable_scrollsUpWhenSelectionLeavesViewport(t *testing.T) {
	h, _ := tableHarness(t, 4, 2)
	defer h.Close()

	h.SendKey(KeyUp)
	h.SendKey(KeyUp)
	h.SendKey(KeyUp)
	h.SendKey(KeyUp) // sel=0 — alpha must scroll back into view

	if !h.Contains("alpha") {
		t.Errorf("viewport should scroll up; lines: %v", h.Lines())
	}
}

func TestNewTable_ignoredWhenUnfocused(t *testing.T) {
	var lastSel int
	h := NewHarness(t, func() Element {
		sel, setSel := UseState(0)
		lastSel = sel
		return Box(
			C(NewButton("Other", func() {})), // takes focus ahead of NewTable
			C(NewTable(tableCols, tableRows, sel, func(i int) { setSel(i) }, 3)),
		)
	})
	defer h.Close()

	h.SendKey(KeyDown)

	if lastSel != 0 {
		t.Errorf("unfocused table should ignore keypresses; sel=%d", lastSel)
	}
}

func TestNewTable_truncatesRegionColumn(t *testing.T) {
	h, _ := tableHarness(t, 0, 5)
	defer h.Close()

	// MaxWidth=6 on Region: "us-east-1" (9 chars) → "us-ea…" (5 chars + ellipsis)
	if !h.Contains("us-ea…") {
		t.Errorf("region column should be truncated to MaxWidth=6; lines: %v", h.Lines())
	}
}

func TestNewTable_customFocusStyle(t *testing.T) {
	custom := NewStyle().Bold().Foreground(ColorBrightGreen)
	h := NewHarness(t, func() Element {
		sel, setSel := UseState(0)
		return C(NewTable(tableCols, tableRows, sel, func(i int) { setSel(i) }, 5, custom))
	})
	defer h.Close()

	for y, line := range h.Lines() {
		if strings.Contains(line, "alpha") {
			got := h.CellStyle(0, y)
			if got != custom.toTcell() {
				t.Errorf("custom focus style not applied: got %v, want %v", got, custom.toTcell())
			}
			return
		}
	}
	t.Error("could not find row containing 'alpha'")
}

func TestNewTable_showsFocusStyleOnSelectedRow(t *testing.T) {
	h, _ := tableHarness(t, 0, 5)
	defer h.Close()

	// Locate the row containing "alpha" and check its style is non-default.
	for y, line := range h.Lines() {
		if strings.Contains(line, "alpha") {
			got := h.CellStyle(0, y)
			if got == (Style{}).toTcell() {
				t.Error("focused selected row should have non-default style")
			}
			return
		}
	}
	t.Error("could not find row containing 'alpha'")
}

// ── NewTable horizontal scrolling ─────────────────────────────────────────────

// narrowTableHarness creates a 25-column harness. At that width the three
// standard test columns do not all fit at once (total natural width = 31),
// so horizontal scrolling is exercised.
func narrowTableHarness(t *testing.T) *Harness {
	t.Helper()
	h := NewHarnessSize(t, func() Element {
		sel, setSel := UseState(0)
		return C(NewTable(tableCols, tableRows, sel, func(i int) { setSel(i) }, 3))
	}, 25, 20)
	return h
}

// TestNewTable_hScrollRight_showsHiddenColumn verifies that pressing Right
// shifts the column viewport so the initially-hidden Region column becomes visible.
func TestNewTable_hScrollRight_showsHiddenColumn(t *testing.T) {
	h := narrowTableHarness(t)
	defer h.Close()

	if h.Contains("Region") {
		t.Skip("all columns already fit; increase column widths or reduce harness width")
	}

	h.SendKey(KeyRight)

	if !h.Contains("Region") {
		t.Errorf("after Right: Region column should be visible; lines: %v", h.Lines())
	}
}

// TestNewTable_hScrollRight_hidesFirstColumn verifies that after scrolling right
// the leftmost column (Name) is no longer rendered.
func TestNewTable_hScrollRight_hidesFirstColumn(t *testing.T) {
	h := narrowTableHarness(t)
	defer h.Close()

	h.SendKey(KeyRight)

	if h.Contains("Name") {
		t.Errorf("after Right: Name column should be hidden; lines: %v", h.Lines())
	}
}

// TestNewTable_hScrollLeft_restoresFirstColumn verifies that pressing Left after
// Right returns the view to the original column set.
func TestNewTable_hScrollLeft_restoresFirstColumn(t *testing.T) {
	h := narrowTableHarness(t)
	defer h.Close()

	h.SendKey(KeyRight)
	h.SendKey(KeyLeft)

	if !h.Contains("Name") {
		t.Errorf("after Right+Left: Name column should be visible again; lines: %v", h.Lines())
	}
	if h.Contains("Region") {
		t.Errorf("after Right+Left: Region column should be hidden again; lines: %v", h.Lines())
	}
}

// TestNewTable_hScrollNoOpAtFirstColumn verifies that Left does nothing when the
// leftmost column is already at column 0.
func TestNewTable_hScrollNoOpAtFirstColumn(t *testing.T) {
	h := narrowTableHarness(t)
	defer h.Close()

	h.SendKey(KeyLeft)

	if !h.Contains("Name") {
		t.Errorf("Left at first column should be no-op; lines: %v", h.Lines())
	}
}

// TestNewTable_hScrollNoOpWhenAllFit verifies that Right is a no-op when all
// columns are already visible (wide terminal).
func TestNewTable_hScrollNoOpWhenAllFit(t *testing.T) {
	// Default 80-wide harness — all tableCols fit.
	h, _ := tableHarness(t, 0, 3)
	defer h.Close()

	h.SendKey(KeyRight)

	if !h.Contains("Name") {
		t.Errorf("Name should still be visible after Right when all columns fit; lines: %v", h.Lines())
	}
	if !h.Contains("Region") {
		t.Errorf("Region should still be visible after Right when all columns fit; lines: %v", h.Lines())
	}
}

// TestNewTable_hScrollRightIndicator verifies that the top border contains a ▶
// indicator when columns are hidden to the right.
func TestNewTable_hScrollRightIndicator(t *testing.T) {
	h := narrowTableHarness(t)
	defer h.Close()

	if h.Contains("Region") {
		t.Skip("all columns already fit")
	}
	if !strings.Contains(h.Line(0), "▶") {
		t.Errorf("top border should contain ▶ when columns are hidden to the right; line 0: %q", h.Line(0))
	}
}

// ── NewTable vertical scroll indicators ──────────────────────────────────────

// TestNewTable_vScrollDownIndicator verifies that the bottom border shows ↓
// when rows are hidden below the visible viewport.
func TestNewTable_vScrollDownIndicator(t *testing.T) {
	// height=3, 5 rows: delta and echo are below the viewport from the start.
	h, _ := tableHarness(t, 0, 3)
	defer h.Close()

	// Bottom border is at line height+3 = 6 (top + header + mid + 3 data rows).
	botLine := h.Line(6)
	if !strings.Contains(botLine, "↓") {
		t.Errorf("bottom border should contain ↓ when rows are hidden below; got %q", botLine)
	}
}

// TestNewTable_vScrollUpIndicator verifies that the mid border (header separator)
// shows ↑ when rows are hidden above the current viewport.
func TestNewTable_vScrollUpIndicator(t *testing.T) {
	h, _ := tableHarness(t, 0, 3)
	defer h.Close()

	// Three downs: sel=3 pushes offset to 1, scrolling alpha out of view.
	h.SendKey(KeyDown)
	h.SendKey(KeyDown)
	h.SendKey(KeyDown)

	// Mid border is always at line 2.
	midLine := h.Line(2)
	if !strings.Contains(midLine, "↑") {
		t.Errorf("mid border should contain ↑ when rows are hidden above; got %q", midLine)
	}
}

// TestNewTable_vScrollNoIndicatorsWhenAllFit verifies that neither ↑ nor ↓
// appears when all rows fit within the viewport height.
func TestNewTable_vScrollNoIndicatorsWhenAllFit(t *testing.T) {
	// height=5 fits all 5 rows exactly — no scrolling needed.
	h, _ := tableHarness(t, 0, 5)
	defer h.Close()

	// Mid border at line 2, bottom border at line 2+1+5=8.
	if strings.Contains(h.Line(2), "↑") {
		t.Errorf("mid border should not show ↑ when all rows fit; got %q", h.Line(2))
	}
	if strings.Contains(h.Line(8), "↓") {
		t.Errorf("bottom border should not show ↓ when all rows fit; got %q", h.Line(8))
	}
}

// TestNewTable_hScrollLeftIndicator verifies that the top border contains a ◀
// indicator when columns are hidden to the left.
func TestNewTable_hScrollLeftIndicator(t *testing.T) {
	h := narrowTableHarness(t)
	defer h.Close()

	h.SendKey(KeyRight)

	if !strings.Contains(h.Line(0), "◀") {
		t.Errorf("top border should contain ◀ after scrolling right; line 0: %q", h.Line(0))
	}
}
