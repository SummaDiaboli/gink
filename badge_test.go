package gink

import "testing"

// ── Badge ─────────────────────────────────────────────────────────────────────

func TestBadge_rendersLabel(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Badge("OK")
	})
	defer h.Close()

	want := "[ OK ]"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestBadge_emptyLabel(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Badge("")
	})
	defer h.Close()

	want := "[  ]"
	if h.Line(0) != want {
		t.Errorf("empty label: got %q, want %q", h.Line(0), want)
	}
}

func TestBadge_appliesStyle(t *testing.T) {
	style := NewStyle().Bold().Foreground(ColorBrightRed)
	h := NewHarness(t, func() Element {
		return Badge("ERR", style)
	})
	defer h.Close()

	if !h.Contains("[ ERR ]") {
		t.Errorf("badge text not found; line: %q", h.Line(0))
	}

	// Every cell in the badge must carry the given style.
	runes := []rune(h.Line(0))
	for x := range len(runes) {
		got := h.CellStyle(x, 0)
		if got != style.toTcell() {
			t.Errorf("cell %d: style not applied (got %v)", x, got)
			break
		}
	}
}

func TestBadge_noStyleRendersDefaultStyle(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Badge("hi")
	})
	defer h.Close()

	// Without a style argument every cell must use the default tcell style.
	runes := []rune(h.Line(0))
	for x := range len(runes) {
		got := h.CellStyle(x, 0)
		if got != (Style{}).toTcell() {
			t.Errorf("cell %d: expected default style, got %v", x, got)
			break
		}
	}
}
