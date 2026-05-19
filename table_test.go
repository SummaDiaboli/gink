package gink

import (
	"strings"
	"testing"
)

// ── Table ─────────────────────────────────────────────────────────────────────

func TestTable_rendersColumnHeaders(t *testing.T) {
	cols := []Column{{Header: "Name"}, {Header: "Status"}}
	h := NewHarness(t, func() Element { return Table(cols, nil) })
	defer h.Close()

	if !h.Contains("Name") {
		t.Error("header 'Name' should be visible")
	}
	if !h.Contains("Status") {
		t.Error("header 'Status' should be visible")
	}
}

func TestTable_rendersDataRows(t *testing.T) {
	cols := []Column{{Header: "Name"}, {Header: "Status"}}
	rows := [][]string{
		{"web-01", "ONLINE"},
		{"web-02", "WARN"},
	}
	h := NewHarness(t, func() Element { return Table(cols, rows) })
	defer h.Close()

	if !h.Contains("web-01") {
		t.Error("row value 'web-01' should be visible")
	}
	if !h.Contains("WARN") {
		t.Error("row value 'WARN' should be visible")
	}
}

func TestTable_borderStructure(t *testing.T) {
	// Single column "A" (1 wide) → inner segment width = 1+2 = 3 dashes.
	cols := []Column{{Header: "A"}}
	h := NewHarness(t, func() Element { return Table(cols, nil) })
	defer h.Close()

	if h.Line(0) != "┌───┐" {
		t.Errorf("top border:    got %q, want %q", h.Line(0), "┌───┐")
	}
	if h.Line(2) != "├───┤" {
		t.Errorf("header sep:    got %q, want %q", h.Line(2), "├───┤")
	}
	if h.Line(3) != "└───┘" {
		t.Errorf("bottom border: got %q, want %q", h.Line(3), "└───┘")
	}
}

func TestTable_columnWidthFitsContent(t *testing.T) {
	// Header "N" (1) < content "Alexander" (9) → column width = 9.
	// Top border: ┌ + ─×(9+2) + ┐ = 13 chars.
	cols := []Column{{Header: "N"}}
	rows := [][]string{{"Alexander"}}
	h := NewHarness(t, func() Element { return Table(cols, rows) })
	defer h.Close()

	if !h.Contains("Alexander") {
		t.Error("content 'Alexander' should appear without truncation")
	}
	top := []rune(h.Line(0))
	if len(top) != 13 {
		t.Errorf("top border width: got %d, want 13", len(top))
	}
}

func TestTable_minWidthExpandsColumn(t *testing.T) {
	// Header "ID" (2) and content "01" (2) are both narrower than MinWidth=10.
	// Column width = 10. Top border: ┌ + ─×12 + ┐ = 14 chars.
	cols := []Column{{Header: "ID", MinWidth: 10}}
	rows := [][]string{{"01"}}
	h := NewHarness(t, func() Element { return Table(cols, rows) })
	defer h.Close()

	top := []rune(h.Line(0))
	if len(top) != 14 {
		t.Errorf("top border width: got %d, want 14 (MinWidth=10 → col width 10)", len(top))
	}
}

func TestTable_maxWidthTruncatesContent(t *testing.T) {
	// "Alexander" (9) capped at MaxWidth=5 → "Alex…" (5 runes).
	cols := []Column{{Header: "Name", MaxWidth: 5}}
	rows := [][]string{{"Alexander"}}
	h := NewHarness(t, func() Element { return Table(cols, rows) })
	defer h.Close()

	if !h.Contains("Alex…") {
		t.Error("content should be truncated to 'Alex…' when MaxWidth=5")
	}
	// Top border: ┌ + ─×(5+2) + ┐ = 9 chars — column is capped, not expanded.
	top := []rune(h.Line(0))
	if len(top) != 9 {
		t.Errorf("top border width: got %d, want 9 (MaxWidth=5)", len(top))
	}
}

func TestTable_maxWidthAlsoCapsShorterHeader(t *testing.T) {
	// Header "LongHeader" (10) capped at MaxWidth=5 → "Long…".
	cols := []Column{{Header: "LongHeader", MaxWidth: 5}}
	h := NewHarness(t, func() Element { return Table(cols, nil) })
	defer h.Close()

	if !h.Contains("Long…") {
		t.Error("header should be truncated to 'Long…' when MaxWidth=5")
	}
}

func TestTable_emptyRowsHasFourLines(t *testing.T) {
	// top + header + separator + bottom = 4 lines (indices 0-3); line 4 is empty.
	cols := []Column{{Header: "Name"}, {Header: "Value"}}
	h := NewHarness(t, func() Element { return Table(cols, nil) })
	defer h.Close()

	if h.Line(4) != "" {
		t.Errorf("line 4 should be empty (table has 4 lines with no rows); got %q", h.Line(4))
	}
}

func TestTable_multipleColumnsOnSameLine(t *testing.T) {
	// Both cell values must appear on the same line (the data row).
	cols := []Column{{Header: "Name"}, {Header: "Status"}}
	rows := [][]string{{"web-01", "ONLINE"}}
	h := NewHarness(t, func() Element { return Table(cols, rows) })
	defer h.Close()

	// Line 3 is the first (and only) data row.
	line := h.Line(3)
	if !strings.Contains(line, "web-01") || !strings.Contains(line, "ONLINE") {
		t.Errorf("data row should contain both values; got %q", line)
	}
}

func TestTable_appliesStyleToBorderCells(t *testing.T) {
	style := NewStyle().Bold().Foreground(ColorBrightCyan)
	cols := []Column{{Header: "A"}}
	h := NewHarness(t, func() Element { return Table(cols, nil, style) })
	defer h.Close()

	// Top-left corner ┌ is at cell (0, 0).
	got := h.CellStyle(0, 0)
	if got != style.toTcell() {
		t.Errorf("border cell (0,0): expected border style, got %v", got)
	}
}
