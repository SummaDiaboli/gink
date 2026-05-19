package gink

import (
	"fmt"
	"testing"
)

// ── Global scroll ─────────────────────────────────────────────────────────────

// tallHarness creates a 80×10 harness with 30 rows of content — enough to
// force scrolling without needing a large harness allocation.
func tallHarness(t *testing.T) *Harness {
	t.Helper()
	items := make([]Element, 30)
	for i := range items {
		items[i] = Text(fmt.Sprintf("Row %02d", i))
	}
	content := Box(items...)
	return NewHarnessSize(t, func() Element { return content }, 80, 10)
}

func TestScroll_topContentVisibleInitially(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	if !h.Contains("Row 00") {
		t.Error("Row 00 should be visible at the top before any scrolling")
	}
}

func TestScroll_contentBeyondViewportNotVisibleInitially(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	if h.Contains("Row 29") {
		t.Error("Row 29 should not be visible before scrolling (content below fold)")
	}
}

func TestScroll_pageDownRevealsContentBelow(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	h.PageDown()

	if !h.Contains("Row 10") {
		t.Errorf("after PageDown, Row 10 should be visible; lines: %v", h.Lines())
	}
	if h.Contains("Row 00") {
		t.Error("Row 00 should have scrolled off screen after PageDown")
	}
}

func TestScroll_pageUpRestoresTopContent(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	h.PageDown()
	h.PageUp()

	if !h.Contains("Row 00") {
		t.Errorf("after PageDown+PageUp, Row 00 should be visible again; lines: %v", h.Lines())
	}
}

func TestScroll_pageUpAtTopIsNoOp(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	h.PageUp()

	if !h.Contains("Row 00") {
		t.Error("PageUp at top should be a no-op; Row 00 should still be visible")
	}
}

func TestScroll_clampedAtBottom(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	// Scroll far past the end — should clamp so Row 29 is visible.
	h.PageDown()
	h.PageDown()
	h.PageDown()
	h.PageDown()

	if !h.Contains("Row 29") {
		t.Errorf("after repeated PageDown, Row 29 should be visible; lines: %v", h.Lines())
	}
}

func TestScroll_noScrollWhenContentFits(t *testing.T) {
	// Content that fits within the harness should not require scrolling.
	h := NewHarness(t, func() Element {
		return Box(Text("Alpha"), Text("Beta"), Text("Gamma"))
	})
	defer h.Close()

	if !h.Contains("Alpha") || !h.Contains("Beta") || !h.Contains("Gamma") {
		t.Error("all content should be visible when it fits within the terminal")
	}
}

func TestScroll_showsDownIndicatorWhenContentBelow(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	// Content extends past the 10-row viewport — ↓ should appear.
	if !h.Contains("↓") {
		t.Error("↓ indicator should be visible when content extends below the viewport")
	}
}

func TestScroll_showsUpIndicatorWhenScrolledDown(t *testing.T) {
	h := tallHarness(t)
	defer h.Close()

	h.PageDown()

	if !h.Contains("↑") {
		t.Error("↑ indicator should appear after scrolling down")
	}
}

func TestScroll_noIndicatorsWhenContentFits(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(Text("Alpha"), Text("Beta"))
	})
	defer h.Close()

	if h.Contains("↑") || h.Contains("↓") {
		t.Error("scroll indicators should not appear when content fits in the terminal")
	}
}
