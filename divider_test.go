package gink

import (
	"strings"
	"testing"
)

func TestDivider_fillsTerminalWidth(t *testing.T) {
	h := NewHarnessSize(t, func() Element { return C(Divider) }, 20, 5)
	defer h.Close()

	want := strings.Repeat("─", 20)
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestDivider_adaptsToWidth(t *testing.T) {
	for _, width := range []int{10, 40, 80} {
		h := NewHarnessSize(t, func() Element { return C(Divider) }, width, 3)
		want := strings.Repeat("─", width)
		if h.Line(0) != want {
			t.Errorf("width=%d: got %q, want %q", width, h.Line(0), want)
		}
		h.Close()
	}
}

func TestDividerWithLabel_centersLabel(t *testing.T) {
	// Width=14, inner=" Hi "=4, dashes=10, left=5, right=5
	h := NewHarnessSize(t, func() Element { return C(DividerWithLabel("Hi")) }, 14, 3)
	defer h.Close()

	want := "───── Hi ─────"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestDividerWithLabel_asymmetricWhenOdd(t *testing.T) {
	// Width=13, inner=" Hi "=4, dashes=9, left=4, right=5
	h := NewHarnessSize(t, func() Element { return C(DividerWithLabel("Hi")) }, 13, 3)
	defer h.Close()

	want := "──── Hi ─────"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestDividerWithLabel_fallsBackWhenTooNarrow(t *testing.T) {
	// Width=4, inner=" Hi "=4, dashes=0 — falls back to just the label.
	h := NewHarnessSize(t, func() Element { return C(DividerWithLabel("Hi")) }, 4, 3)
	defer h.Close()

	if !h.Contains("Hi") {
		t.Error("expected label to appear even when terminal is too narrow for dashes")
	}
}

func TestDividerStyled_rendersCorrectWidth(t *testing.T) {
	style := NewStyle().Bold()
	h := NewHarnessSize(t, func() Element { return C(DividerStyled(style)) }, 10, 3)
	defer h.Close()

	want := strings.Repeat("─", 10)
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestDivider_fillsRemainingRowWidth(t *testing.T) {
	h := NewHarnessSize(t, func() Element {
		return Row(
			C(func() Element { return Text("LEFT") }),
			C(Divider),
			C(func() Element { return Text("RIGHT") }),
		)
	}, 20, 3)
	defer h.Close()

	want := "LEFT" + strings.Repeat("─", 11) + "RIGHT"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestDividerWithLabel_wideLabelCentersByColumns(t *testing.T) {
	h := NewHarnessSize(t, func() Element { return C(DividerWithLabel("世界")) }, 10, 3)
	defer h.Close()

	if got := h.Line(0); got != "── 世界 ──" {
		t.Errorf("line = %q, want %q", got, "── 世界 ──")
	}
	if got := h.CellGrapheme(3, 0); got != "世" {
		t.Errorf("cell 3 = %q, want %q", got, "世")
	}
	if got := h.CellWidth(3, 0); got != 2 {
		t.Errorf("cell 3 width = %d, want 2", got)
	}
	if got := h.CellGrapheme(4, 0); got != "" {
		t.Errorf("cell 4 = %q, want empty", got)
	}
}

func TestDividerWithLabel_narrowWideLabelDoesNotSplitCluster(t *testing.T) {
	h := NewHarnessSize(t, func() Element { return C(DividerWithLabel("世界")) }, 5, 3)
	defer h.Close()

	if displayWidth(h.Line(0)) > 5 {
		t.Errorf("line width = %d, exceeds 5 (%q)", displayWidth(h.Line(0)), h.Line(0))
	}
	if got := h.CellGrapheme(1, 0); got != "世" {
		t.Errorf("cell 1 = %q, want %q", got, "世")
	}
	if got := h.CellGrapheme(2, 0); got != "" {
		t.Errorf("cell 2 = %q, want empty", got)
	}
	if got := h.CellGrapheme(3, 0); got != "界" {
		t.Errorf("cell 3 = %q, want %q", got, "界")
	}
}
