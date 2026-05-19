package gink

import (
	"strings"
	"testing"
)

// ── Constrain / MinWidth / MaxWidth / MinHeight / MaxHeight ───────────────────

func TestConstrain_minWidthPushesSiblingRight(t *testing.T) {
	// "Hi" is 2 chars; MinWidth=10 must make the Row place "|" at col 10.
	h := NewHarness(t, func() Element {
		return Row(MinWidth(10, Text("Hi")), Text("|"))
	})
	defer h.Close()

	line := []rune(h.Line(0))
	if len(line) <= 10 || line[10] != '|' {
		t.Errorf("separator should be at col 10; line: %q", h.Line(0))
	}
}

func TestConstrain_maxWidthClipsContent(t *testing.T) {
	h := NewHarness(t, func() Element {
		return MaxWidth(5, Text("Hello World"))
	})
	defer h.Close()

	if h.Contains("World") {
		t.Error("content beyond MaxWidth should be clipped")
	}
	if !h.Contains("Hello") {
		t.Error("content within MaxWidth should be visible")
	}
}

func TestConstrain_minHeightPushesSiblingDown(t *testing.T) {
	// Single-line text with MinHeight=3 must place the next sibling on row 3.
	h := NewHarness(t, func() Element {
		return Box(MinHeight(3, Text("Top")), Text("Bottom"))
	})
	defer h.Close()

	for i, line := range h.Lines() {
		if strings.Contains(line, "Bottom") {
			if i != 3 {
				t.Errorf("Bottom at row %d, want 3", i)
			}
			return
		}
	}
	t.Error("Bottom not found")
}

func TestConstrain_maxHeightClipsContent(t *testing.T) {
	h := NewHarness(t, func() Element {
		return MaxHeight(2, Box(Text("Line 1"), Text("Line 2"), Text("Line 3")))
	})
	defer h.Close()

	if h.Contains("Line 3") {
		t.Error("content beyond MaxHeight should be clipped")
	}
	if !h.Contains("Line 1") || !h.Contains("Line 2") {
		t.Error("content within MaxHeight should be visible")
	}
}

func TestConstrain_zeroMeansNoConstraint(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Constrain(Text("Hello"), 0, 0, 0, 0)
	})
	defer h.Close()

	if !h.Contains("Hello") {
		t.Error("unconstrained element should render normally")
	}
}

func TestConstrain_contentWithinBoundsNotClipped(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Constrain(Text("Hi"), 0, 20, 0, 5)
	})
	defer h.Close()

	if !h.Contains("Hi") {
		t.Error("content within bounds should not be clipped")
	}
}

func TestConstrain_maxWidthDoesNotAffectNarrowerContent(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Row(MaxWidth(20, Text("Hi")), Text("|"))
	})
	defer h.Close()

	// "Hi" is 2 chars wide; MaxWidth=20 should not pad it — "|" lands at col 2.
	line := []rune(h.Line(0))
	if len(line) < 3 || line[2] != '|' {
		t.Errorf("MaxWidth should not pad narrower content; line: %q", h.Line(0))
	}
}

func TestMinWidth_convenience(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Row(MinWidth(8, Text("X")), Text("|"))
	})
	defer h.Close()

	line := []rune(h.Line(0))
	if len(line) <= 8 || line[8] != '|' {
		t.Errorf("MinWidth convenience wrapper failed; line: %q", h.Line(0))
	}
}

func TestMaxWidth_convenience(t *testing.T) {
	h := NewHarness(t, func() Element {
		return MaxWidth(3, Text("Hello"))
	})
	defer h.Close()

	if h.Contains("lo") {
		t.Error("MaxWidth(3) should clip 'Hello' to 'Hel'")
	}
}

func TestMinHeight_convenience(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(MinHeight(2, Text("A")), Text("B"))
	})
	defer h.Close()

	for i, line := range h.Lines() {
		if strings.Contains(line, "B") {
			if i != 2 {
				t.Errorf("B at row %d, want 2", i)
			}
			return
		}
	}
	t.Error("B not found")
}

func TestMaxHeight_convenience(t *testing.T) {
	h := NewHarness(t, func() Element {
		return MaxHeight(1, Box(Text("A"), Text("B")))
	})
	defer h.Close()

	if h.Contains("B") {
		t.Error("MaxHeight(1) should clip second line")
	}
}
