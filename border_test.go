package gink

import "testing"

// ── Border ────────────────────────────────────────────────────────────────────

func TestBorder_surroundsChildWithLineDrawingChars(t *testing.T) {
	// Child: "Hello" (5 wide, 1 tall) → total 7×3
	h := NewHarness(t, func() Element {
		return Border(Text("Hello"))
	})
	defer h.Close()

	if h.Line(0) != "┌─────┐" {
		t.Errorf("top border:    got %q, want %q", h.Line(0), "┌─────┐")
	}
	if h.Line(1) != "│Hello│" {
		t.Errorf("content row:   got %q, want %q", h.Line(1), "│Hello│")
	}
	if h.Line(2) != "└─────┘" {
		t.Errorf("bottom border: got %q, want %q", h.Line(2), "└─────┘")
	}
}

func TestBorder_contentIsInset(t *testing.T) {
	// Content "Hi" must appear at column 1 (not column 0).
	h := NewHarness(t, func() Element {
		return Border(Text("Hi"))
	})
	defer h.Close()

	if h.Line(1) != "│Hi│" {
		t.Errorf("content row: got %q, want %q", h.Line(1), "│Hi│")
	}
}

func TestBorder_multiLineChild(t *testing.T) {
	// Child: Box with two rows ("A" and "B"), each 1 wide → total 3×4
	h := NewHarness(t, func() Element {
		return Border(Box(Text("A"), Text("B")))
	})
	defer h.Close()

	if h.Line(0) != "┌─┐" {
		t.Errorf("top:    got %q, want %q", h.Line(0), "┌─┐")
	}
	if h.Line(1) != "│A│" {
		t.Errorf("row 1:  got %q, want %q", h.Line(1), "│A│")
	}
	if h.Line(2) != "│B│" {
		t.Errorf("row 2:  got %q, want %q", h.Line(2), "│B│")
	}
	if h.Line(3) != "└─┘" {
		t.Errorf("bottom: got %q, want %q", h.Line(3), "└─┘")
	}
}

func TestBorder_totalSizeIsChildPlusTwo(t *testing.T) {
	// "ABCDE" is 5 wide, 1 tall → border is 7 wide, 3 tall.
	h := NewHarness(t, func() Element {
		return Border(Text("ABCDE"))
	})
	defer h.Close()

	top := []rune(h.Line(0))
	if len(top) != 7 {
		t.Errorf("top border width: got %d, want 7", len(top))
	}
	if h.Line(3) != "" {
		t.Errorf("line 3 should be empty (border is only 3 tall); got %q", h.Line(3))
	}
}

func TestBorder_appliesStyleToBorderCells(t *testing.T) {
	style := NewStyle().Bold().Foreground(ColorBrightCyan)
	h := NewHarness(t, func() Element {
		return Border(Text("Hi"), style)
	})
	defer h.Close()

	// Corner cells and top-border dash must use the border style.
	corners := [][2]int{{0, 0}, {3, 0}, {0, 2}, {3, 2}} // ┌ ┐ └ ┘ for "Hi" (4 wide border)
	for _, pos := range corners {
		got := h.CellStyle(pos[0], pos[1])
		if got != style.toTcell() {
			t.Errorf("border cell (%d,%d): expected border style, got %v", pos[0], pos[1], got)
		}
	}
}

func TestBorder_contentKeepsOwnStyle(t *testing.T) {
	contentStyle := NewStyle().Foreground(ColorBrightRed)
	borderStyle := NewStyle().Foreground(ColorBrightCyan)
	h := NewHarness(t, func() Element {
		return Border(Text("Hi", contentStyle), borderStyle)
	})
	defer h.Close()

	// Content "H" is at column 1, row 1.
	got := h.CellStyle(1, 1)
	if got != contentStyle.toTcell() {
		t.Errorf("content cell (1,1): expected content style, got %v", got)
	}
}

func TestBorder_worksInsideBox(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(
			Text("Above"),
			Border(Text("Hi")),
			Text("Below"),
		)
	})
	defer h.Close()

	if h.Line(0) != "Above" {
		t.Errorf("line 0: got %q, want Above", h.Line(0))
	}
	if h.Line(1) != "┌──┐" {
		t.Errorf("line 1: got %q, want ┌──┐", h.Line(1))
	}
	if h.Line(2) != "│Hi│" {
		t.Errorf("line 2: got %q, want │Hi│", h.Line(2))
	}
	if h.Line(3) != "└──┘" {
		t.Errorf("line 3: got %q, want └──┘", h.Line(3))
	}
	if h.Line(4) != "Below" {
		t.Errorf("line 4: got %q, want Below", h.Line(4))
	}
}

// ── BorderWithTitle ───────────────────────────────────────────────────────────

func TestBorderWithTitle_showsTitleInTopBorder(t *testing.T) {
	// "Hello World" = 11 wide → border width = 13
	// Top: ┌─ Hi ─────┐  (remaining = 11 - 4 - 2 = 5 dashes after title)
	h := NewHarness(t, func() Element {
		return BorderWithTitle("Hi", Text("Hello World"))
	})
	defer h.Close()

	if h.Line(0) != "┌─ Hi ──────┐" {
		t.Errorf("top border with title: got %q, want %q", h.Line(0), "┌─ Hi ──────┐")
	}
	if h.Line(1) != "│Hello World│" {
		t.Errorf("content row: got %q, want %q", h.Line(1), "│Hello World│")
	}
	if h.Line(2) != "└───────────┘" {
		t.Errorf("bottom border: got %q, want %q", h.Line(2), "└───────────┘")
	}
}

func TestBorderWithTitle_titleFitsExactly(t *testing.T) {
	// Child "AB" = 2 wide → border width = 4
	// Top: ┌ ─ space T i t l e space ─ ┐ — title "X" (1 char), remaining = 2 - 4 - 1 = -3
	// Falls back to: ┌─ X ─┐  BUT that's 7 chars and child is only 2 wide.
	// So the border must widen to fit the title if needed.
	// ┌─ X ─┐ = 7 chars → child forced to at least 5 chars of inner space.
	// For a simple fallback: when title is too long, just use ┌ title ┐ or truncate.
	// Test with a child wide enough: "ABCDE" (5) → border 7, title "X" (1), remaining = 5-4-1=0
	// Top: ┌─ X ─┐  (no extra dashes)
	h := NewHarness(t, func() Element {
		return BorderWithTitle("X", Text("ABCDE"))
	})
	defer h.Close()

	if h.Line(0) != "┌─ X ─┐" {
		t.Errorf("exact-fit title: got %q, want %q", h.Line(0), "┌─ X ─┐")
	}
}

func TestBorderWithTitle_appliesStyleToBorderAndTitle(t *testing.T) {
	style := NewStyle().Bold()
	h := NewHarness(t, func() Element {
		return BorderWithTitle("T", Text("Hello"), style)
	})
	defer h.Close()

	// Title character "T" is at position 3 (after ┌, ─, space), row 0.
	got := h.CellStyle(3, 0)
	if got != style.toTcell() {
		t.Errorf("title cell (3,0): expected border style, got %v", got)
	}
}
