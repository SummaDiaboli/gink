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
