package gink_test

import (
	"testing"

	"github.com/SummaDiaboli/gink"
)

// TestWidth_expandsNarrowContent verifies that Width pads a narrow element to
// the exact column count, pushing siblings in a Row to the correct position.
func TestWidth_expandsNarrowContent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Row(gink.Width(10, gink.Text("hi")), gink.Text("X"))
	})
	defer h.Close()

	// "hi" is 2 wide; Width(10) should report 10, so "X" lands at col 10.
	if h.Line(0)[10] != 'X' {
		t.Errorf("expected 'X' at column 10, got %q (line: %q)", h.Line(0)[10], h.Line(0))
	}
}

// TestWidth_clipsWideContent verifies that Width clips content exceeding n
// columns and reports exactly n to the parent.
func TestWidth_clipsWideContent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Row(gink.Width(4, gink.Text("hello")), gink.Text("X"))
	})
	defer h.Close()

	// "hello" clipped to 4 → "hell"; "X" at col 4.
	if h.Line(0)[4] != 'X' {
		t.Errorf("expected 'X' at column 4, got %q (line: %q)", h.Line(0)[4], h.Line(0))
	}
	if h.Contains("hello") {
		t.Error("expected 'hello' to be clipped, but it appeared unclipped")
	}
}

// TestWidth_doesNotAffectHeight verifies that Width leaves the element's height
// unchanged.
func TestWidth_doesNotAffectHeight(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.Width(10, gink.Text("hi")),
			gink.Text("after"),
		)
	})
	defer h.Close()

	// Single-line text → height 1; "after" should be on row 1.
	if h.Line(1) != "after" {
		t.Errorf("expected 'after' on row 1, got %q", h.Line(1))
	}
}

// TestHeight_expandsShortContent verifies that Height pads a short element to
// the exact row count, pushing siblings in a Box to the correct row.
func TestHeight_expandsShortContent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.Height(3, gink.Text("hi")),
			gink.Text("after"),
		)
	})
	defer h.Close()

	// "hi" takes 1 row; Height(3) reports 3, so "after" lands at row 3.
	if h.Line(3) != "after" {
		t.Errorf("expected 'after' on row 3, got %q", h.Line(3))
	}
}

// TestHeight_clipsTallContent verifies that Height clips content exceeding n
// rows and reports exactly n to the parent.
func TestHeight_clipsTallContent(t *testing.T) {
	content := gink.Box(gink.Text("row0"), gink.Text("row1"), gink.Text("row2"))
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.Height(2, content),
			gink.Text("after"),
		)
	})
	defer h.Close()

	// Height(2) clips after row1; "after" at row 2, "row2" never rendered.
	if h.Line(2) != "after" {
		t.Errorf("expected 'after' on row 2, got %q", h.Line(2))
	}
	if h.Contains("row2") {
		t.Error("expected 'row2' to be clipped")
	}
}

// TestHeight_doesNotAffectWidth verifies that Height leaves the element's width
// unchanged.
func TestHeight_doesNotAffectWidth(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Row(
			gink.Height(3, gink.Text("hi")),
			gink.Text("X"),
		)
	})
	defer h.Close()

	// "hi" is 2 wide; Height doesn't change width, so "X" at col 2.
	if h.Line(0)[2] != 'X' {
		t.Errorf("expected 'X' at column 2, got %q (line: %q)", h.Line(0)[2], h.Line(0))
	}
}

// TestSize_setsBothDimensions verifies that Size forces exact width and height
// simultaneously.
func TestSize_setsBothDimensions(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.Row(gink.Size(6, 2, gink.Text("hi")), gink.Text("X")),
			gink.Text("after"),
		)
	})
	defer h.Close()

	// Width=6 → "X" at col 6; Height=2 → "after" at row 2.
	if h.Line(0)[6] != 'X' {
		t.Errorf("expected 'X' at column 6, got %q", h.Line(0)[6])
	}
	if h.Line(2) != "after" {
		t.Errorf("expected 'after' on row 2, got %q", h.Line(2))
	}
}
