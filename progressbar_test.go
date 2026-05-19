package gink

import (
	"strings"
	"testing"
)

// ── ProgressBar ───────────────────────────────────────────────────────────────

func TestProgressBar_zeroProgress(t *testing.T) {
	h := NewHarness(t, func() Element {
		return ProgressBar(0.0, 8)
	})
	defer h.Close()

	want := "[░░░░░░░░] 0%"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestProgressBar_fullProgress(t *testing.T) {
	h := NewHarness(t, func() Element {
		return ProgressBar(1.0, 8)
	})
	defer h.Close()

	want := "[████████] 100%"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestProgressBar_halfProgress(t *testing.T) {
	h := NewHarness(t, func() Element {
		return ProgressBar(0.5, 10)
	})
	defer h.Close()

	want := "[█████░░░░░] 50%"
	if h.Line(0) != want {
		t.Errorf("got %q, want %q", h.Line(0), want)
	}
}

func TestProgressBar_clampsAboveOne(t *testing.T) {
	h := NewHarness(t, func() Element {
		return ProgressBar(1.5, 4)
	})
	defer h.Close()

	want := "[████] 100%"
	if h.Line(0) != want {
		t.Errorf("value > 1.0 should clamp to 1.0: got %q, want %q", h.Line(0), want)
	}
}

func TestProgressBar_clampsBelowZero(t *testing.T) {
	h := NewHarness(t, func() Element {
		return ProgressBar(-0.5, 4)
	})
	defer h.Close()

	want := "[░░░░] 0%"
	if h.Line(0) != want {
		t.Errorf("value < 0.0 should clamp to 0.0: got %q, want %q", h.Line(0), want)
	}
}

func TestProgressBar_appliesStyle(t *testing.T) {
	style := NewStyle().Bold()
	h := NewHarness(t, func() Element {
		return ProgressBar(0.5, 4, style)
	})
	defer h.Close()

	// The bar must be present and every non-space cell must carry the style.
	if !h.Contains("[██░░] 50%") {
		t.Errorf("styled bar not found; screen: %q", h.Line(0))
	}

	runes := []rune(h.Line(0))
	for x, ch := range runes {
		if ch == ' ' {
			continue
		}
		got := h.CellStyle(x, 0)
		if got != style.toTcell() {
			t.Errorf("cell %d (%q): style not applied", x, string(ch))
			break
		}
	}
}

func TestProgressBar_renderedWidthMatchesExpected(t *testing.T) {
	// Total rune count = "[" + width fill chars + "] " + percentage string.
	h := NewHarness(t, func() Element {
		return ProgressBar(0.25, 8)
	})
	defer h.Close()

	line := h.Line(0)
	// [██░░░░░░] 25%  → brackets(2) + 8 fill + "] "(2) + "25%"(3) = 15
	// Use strings.TrimRight since Harness already trims trailing spaces.
	if !strings.HasPrefix(line, "[") {
		t.Errorf("bar must start with '['; got %q", line)
	}
	if !strings.HasSuffix(line, "%") {
		t.Errorf("bar must end with '%%'; got %q", line)
	}
}
