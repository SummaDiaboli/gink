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
