package gink

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ── Text ─────────────────────────────────────────────────────────────────────

func TestText_rendersAtOrigin(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Text("hello")
	})
	defer h.Close()

	if h.Line(0) != "hello" {
		t.Errorf("line 0: got %q, want %q", h.Line(0), "hello")
	}
}

func TestText_doesNotOverflowBuffer(t *testing.T) {
	h := NewHarnessSize(t, func() Element {
		return Text("abcde")
	}, 3, 5) // narrower than the string
	defer h.Close()

	line := h.Line(0)
	if len([]rune(line)) > 3 {
		t.Errorf("text overflowed narrow buffer: %q", line)
	}
}

func TestText_unicodeHandledCorrectly(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Text("héllo")
	})
	defer h.Close()

	if !h.Contains("héllo") {
		t.Errorf("unicode text not found; lines: %v", h.Lines())
	}
}

func TestText_emptyString(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Text("")
	})
	defer h.Close()

	if h.Line(0) != "" {
		t.Errorf("empty text: line 0 = %q, want empty", h.Line(0))
	}
}

// ── Box ──────────────────────────────────────────────────────────────────────

func TestBox_stacksChildrenVertically(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(
			Text("first"),
			Text("second"),
			Text("third"),
		)
	})
	defer h.Close()

	if h.Line(0) != "first" {
		t.Errorf("line 0: got %q, want %q", h.Line(0), "first")
	}
	if h.Line(1) != "second" {
		t.Errorf("line 1: got %q, want %q", h.Line(1), "second")
	}
	if h.Line(2) != "third" {
		t.Errorf("line 2: got %q, want %q", h.Line(2), "third")
	}
}

func TestBox_singleChild(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(Text("only"))
	})
	defer h.Close()

	if h.Line(0) != "only" {
		t.Errorf("line 0: got %q, want %q", h.Line(0), "only")
	}
}

func TestBoxWithGap_insertsBlankLines(t *testing.T) {
	h := NewHarness(t, func() Element {
		return BoxWithGap(2,
			Text("A"),
			Text("B"),
		)
	})
	defer h.Close()

	// A at row 0, 2 blank rows, B at row 3
	if h.Line(0) != "A" {
		t.Errorf("line 0: got %q, want A", h.Line(0))
	}
	if h.Line(1) != "" {
		t.Errorf("line 1 (gap): got %q, want empty", h.Line(1))
	}
	if h.Line(2) != "" {
		t.Errorf("line 2 (gap): got %q, want empty", h.Line(2))
	}
	if h.Line(3) != "B" {
		t.Errorf("line 3: got %q, want B", h.Line(3))
	}
}

// ── Row ──────────────────────────────────────────────────────────────────────

func TestRow_laysOutChildrenHorizontally(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Row(
			Text("hello "),
			Text("world"),
		)
	})
	defer h.Close()

	if h.Line(0) != "hello world" {
		t.Errorf("line 0: got %q, want %q", h.Line(0), "hello world")
	}
}

func TestRow_allOnSameLine(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Row(Text("A"), Text("B"), Text("C"))
	})
	defer h.Close()

	if h.Line(0) != "ABC" {
		t.Errorf("line 0: got %q, want ABC", h.Line(0))
	}
	if h.Line(1) != "" {
		t.Errorf("line 1: got %q, want empty (row should not span multiple lines)", h.Line(1))
	}
}

func TestRowWithGap_insertsSpaces(t *testing.T) {
	h := NewHarness(t, func() Element {
		return RowWithGap(3, Text("A"), Text("B"))
	})
	defer h.Close()

	// A, then 3 spaces, then B
	if h.Line(0) != "A   B" {
		t.Errorf("line 0: got %q, want %q", h.Line(0), "A   B")
	}
}

// ── Nesting ──────────────────────────────────────────────────────────────────

func TestNested_rowInsideBox(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(
			Row(Text("A"), Text("B")),
			Row(Text("C"), Text("D")),
		)
	})
	defer h.Close()

	if h.Line(0) != "AB" {
		t.Errorf("line 0: got %q, want AB", h.Line(0))
	}
	if h.Line(1) != "CD" {
		t.Errorf("line 1: got %q, want CD", h.Line(1))
	}
}

func TestNested_boxInsideRow(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Row(
			Text("X"),
			Box(Text("A"), Text("B")),
		)
	})
	defer h.Close()

	// X and A are on the same row; B is on the row below X's position.
	if h.Line(0) != "XA" {
		t.Errorf("line 0: got %q, want XA", h.Line(0))
	}
	if h.Line(1) != " B" {
		t.Errorf("line 1: got %q, want ' B'", h.Line(1))
	}
}

// ── Style ────────────────────────────────────────────────────────────────────

func TestStyle_boldAppliedToCells(t *testing.T) {
	boldStyle := NewStyle().Bold()
	h := NewHarness(t, func() Element {
		return Text("hi", boldStyle)
	})
	defer h.Close()

	style := h.CellStyle(0, 0)
	_, _, attrs := style.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Error("expected bold attribute on cell (0,0)")
	}
}

func TestStyle_unstyledCellsHaveDefaultStyle(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Text("hi")
	})
	defer h.Close()

	style := h.CellStyle(0, 0)
	if style != tcell.StyleDefault {
		t.Errorf("unstyled text: cell (0,0) style is not default")
	}
}

func TestStyle_chaining(t *testing.T) {
	base := NewStyle()
	styled := base.Bold().Underline().Foreground(ColorBrightRed)

	// base must be unchanged
	if base == styled {
		t.Error("Style.Bold() mutated the receiver instead of returning a new value")
	}
}

// ── Component (C) ────────────────────────────────────────────────────────────

func TestC_wrapsComponentFunction(t *testing.T) {
	inner := func() Element { return Text("inner") }

	h := NewHarness(t, func() Element {
		return C(inner)
	})
	defer h.Close()

	if !h.Contains("inner") {
		t.Errorf("component output not found; lines: %v", h.Lines())
	}
}

func TestC_nestedComponents(t *testing.T) {
	child := func() Element { return Text("child") }

	h := NewHarness(t, func() Element {
		return Box(
			Text("parent"),
			C(child),
		)
	})
	defer h.Close()

	if !h.Contains("parent") || !h.Contains("child") {
		t.Errorf("nested components missing; lines: %v", h.Lines())
	}
}

// ── TermSize ─────────────────────────────────────────────────────────────────

func TestUseTermSize_returnsConfiguredDimensions(t *testing.T) {
	var got TermSize

	h := NewHarnessSize(t, func() Element {
		got = UseTermSize()
		return Text("")
	}, 120, 40)
	defer h.Close()

	if got.Width != 120 || got.Height != 40 {
		t.Errorf("UseTermSize: got %dx%d, want 120x40", got.Width, got.Height)
	}
}

func TestUseTermSize_usedForDivider(t *testing.T) {
	h := NewHarnessSize(t, func() Element {
		size := UseTermSize()
		return Text(strings.Repeat("─", size.Width))
	}, 10, 5)
	defer h.Close()

	if h.Line(0) != "──────────" {
		t.Errorf("divider: got %q, want 10 dashes", h.Line(0))
	}
}
