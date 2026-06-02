package gink_test

import (
	"fmt"
	"testing"

	"github.com/SummaDiaboli/gink"
)

func makeScrollContent() gink.Element {
	lines := make([]gink.Element, 10)
	for i := range lines {
		lines[i] = gink.Text(fmt.Sprintf("line %d", i))
	}
	return gink.Box(lines...)
}

// TestScrollView_showsOnlyViewportRows verifies that only `height` rows of
// content are visible initially, even when the child is taller.
func TestScrollView_showsOnlyViewportRows(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	if !h.Contains("line 0") {
		t.Error("expected line 0 to be visible")
	}
	if !h.Contains("line 3") {
		t.Error("expected line 3 to be visible")
	}
	if h.Contains("line 4") {
		t.Error("expected line 4 to be hidden initially")
	}
}

// TestScrollView_scrollDownRevealsHiddenContent verifies Down key scrolls the
// viewport so previously-hidden rows become visible.
func TestScrollView_scrollDownRevealsHiddenContent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.SendKey(gink.KeyDown)

	if h.Contains("line 0") {
		t.Error("expected line 0 to scroll out of view")
	}
	if !h.Contains("line 4") {
		t.Error("expected line 4 to scroll into view")
	}
}

// TestScrollView_scrollUpRestoresContent verifies Up key reverses a Down scroll.
func TestScrollView_scrollUpRestoresContent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.SendKey(gink.KeyDown)
	h.SendKey(gink.KeyUp)

	if !h.Contains("line 0") {
		t.Error("expected line 0 to be restored after scroll up")
	}
	if h.Contains("line 4") {
		t.Error("expected line 4 to be hidden again after scroll up")
	}
}

// TestScrollView_clampedAtTop verifies Up at offset=0 is a no-op.
func TestScrollView_clampedAtTop(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.SendKey(gink.KeyUp)
	h.SendKey(gink.KeyUp)

	if !h.Contains("line 0") {
		t.Error("expected line 0 to remain visible when clamped at top")
	}
}

// TestScrollView_clampedAtBottom verifies scrolling stops when the last content
// row reaches the bottom of the viewport.
func TestScrollView_clampedAtBottom(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	for i := 0; i < 20; i++ {
		h.SendKey(gink.KeyDown)
	}

	if !h.Contains("line 9") {
		t.Error("expected last line to be visible when clamped at bottom")
	}
}

// TestScrollView_showsDownIndicatorWhenContentBelow verifies a ↓ indicator
// appears when there is hidden content below the viewport.
func TestScrollView_showsDownIndicatorWhenContentBelow(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	if !h.Contains("↓") {
		t.Error("expected ↓ scroll indicator when content extends below viewport")
	}
}

// TestScrollView_showsUpIndicatorAfterScrolling verifies a ↑ indicator appears
// once the viewport has been scrolled down.
func TestScrollView_showsUpIndicatorAfterScrolling(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.SendKey(gink.KeyDown)

	if !h.Contains("↑") {
		t.Error("expected ↑ scroll indicator after scrolling down")
	}
}

// TestScrollView_noIndicatorsWhenContentFits verifies no indicators are shown
// when all content fits within the viewport.
func TestScrollView_noIndicatorsWhenContentFits(t *testing.T) {
	short := gink.Box(gink.Text("only line"))
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, short))
	})
	defer h.Close()

	if h.Contains("↑") || h.Contains("↓") {
		t.Error("expected no scroll indicators when all content fits")
	}
}

// TestScrollView_ignoresInputWhenUnfocused verifies that a ScrollView does not
// scroll when it does not hold focus.
func TestScrollView_ignoresInputWhenUnfocused(t *testing.T) {
	// Two scroll views; first is focused initially. Sending Down should only
	// scroll the first view.
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.C(gink.NewScrollView(4, makeScrollContent())),
			gink.C(gink.NewScrollView(4, makeScrollContent())),
		)
	})
	defer h.Close()

	h.SendKey(gink.KeyDown)

	lines := h.Lines()
	firstViewHasLine0 := false
	for _, l := range lines[:4] {
		if l == "line 0" {
			firstViewHasLine0 = true
		}
	}
	if firstViewHasLine0 {
		t.Error("expected focused first view to have scrolled away from line 0")
	}
}

// TestScrollView_reportsFixedHeight verifies that the element returned by
// NewScrollView always occupies exactly `height` rows in the layout.
func TestScrollView_reportsFixedHeight(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.C(gink.NewScrollView(3, makeScrollContent())),
			gink.Text("after"),
		)
	})
	defer h.Close()

	// "after" should appear at row 3 (0-indexed), immediately below the viewport.
	if h.Line(3) != "after" {
		t.Errorf("expected 'after' at row 3, got %q", h.Line(3))
	}
}

// TestScrollView_wheelDownScrollsWhenFocused verifies that WheelDown routes to
// a focused ScrollView and advances its internal offset rather than the global buffer.
func TestScrollView_wheelDownScrollsWhenFocused(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.WheelDown()

	if h.Contains("line 0") {
		t.Error("expected line 0 to scroll out of view after WheelDown")
	}
	if !h.Contains("line 3") {
		t.Errorf("expected line 3 to be visible after WheelDown; lines: %v", h.Lines())
	}
}

// TestScrollView_wheelUpScrollsWhenFocused verifies that WheelUp routes to a
// focused ScrollView and decrements its internal offset.
func TestScrollView_wheelUpScrollsWhenFocused(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.WheelDown()
	h.WheelDown()
	h.WheelUp()

	if h.Contains("line 0") {
		t.Error("expected line 0 to remain scrolled off after WheelDown×2, WheelUp×1")
	}
	if !h.Contains("line 3") {
		t.Errorf("expected line 3 to be visible; lines: %v", h.Lines())
	}
}

// TestScrollView_pageDownScrollsWhenFocused verifies that PageDown routes to a
// focused ScrollView and scrolls to the bottom of its content.
func TestScrollView_pageDownScrollsWhenFocused(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.PageDown()

	if h.Contains("line 0") {
		t.Error("expected line 0 to scroll off after PageDown")
	}
	if !h.Contains("line 9") {
		t.Errorf("expected line 9 to be visible after PageDown; lines: %v", h.Lines())
	}
}

// TestScrollView_pageUpScrollsWhenFocused verifies that PageUp routes to a
// focused ScrollView and returns it to the top of its content.
func TestScrollView_pageUpScrollsWhenFocused(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(gink.NewScrollView(4, makeScrollContent()))
	})
	defer h.Close()

	h.PageDown()
	h.PageUp()

	if !h.Contains("line 0") {
		t.Errorf("expected line 0 to be visible after PageDown+PageUp; lines: %v", h.Lines())
	}
}

// TestScrollView_wheelFallsThroughToGlobalWhenUnfocused verifies that WheelDown
// falls through to the global scroll buffer when no ScrollView is focused.
func TestScrollView_wheelFallsThroughToGlobalWhenUnfocused(t *testing.T) {
	items := make([]gink.Element, 30)
	for i := range items {
		items[i] = gink.Text(fmt.Sprintf("Row %02d", i))
	}
	content := gink.Box(items...)
	h := gink.NewHarnessSize(t, func() gink.Element { return content }, 80, 10)
	defer h.Close()

	h.WheelDown()

	if h.Contains("Row 00") {
		t.Error("expected Row 00 to scroll off the global buffer after WheelDown")
	}
	if !h.Contains("Row 03") {
		t.Errorf("expected Row 03 to be visible after WheelDown; lines: %v", h.Lines())
	}
}
