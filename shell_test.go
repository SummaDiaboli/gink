package gink

import (
	"fmt"
	"testing"
)

// ── AppShell ──────────────────────────────────────────────────────────────────

func TestAppShell_footerAlwaysVisibleWithoutScroll(t *testing.T) {
	h := NewHarness(t, func() Element {
		return AppShell(Text("Main content"), Text("Footer hint"))
	})
	defer h.Close()

	if !h.Contains("Main content") {
		t.Error("main content should be visible")
	}
	if !h.Contains("Footer hint") {
		t.Error("footer should be visible")
	}
}

func TestAppShell_footerVisibleAfterScrollingMainContent(t *testing.T) {
	// Build main content taller than the harness so it scrolls.
	items := make([]Element, 30)
	for i := range items {
		items[i] = Text(fmt.Sprintf("Row %02d", i))
	}
	main := Box(items...)

	h := NewHarnessSize(t, func() Element {
		return AppShell(main, Text("Sticky Footer"))
	}, 80, 10)
	defer h.Close()

	// Scroll down so the top rows are off screen.
	h.PageDown()

	// Footer must still be visible regardless of scroll position.
	if !h.Contains("Sticky Footer") {
		t.Errorf("footer should remain visible after scrolling; lines: %v", h.Lines())
	}
}

func TestAppShell_mainContentScrollsIndependentlyOfFooter(t *testing.T) {
	items := make([]Element, 30)
	for i := range items {
		items[i] = Text(fmt.Sprintf("Item %02d", i))
	}
	main := Box(items...)

	h := NewHarnessSize(t, func() Element {
		return AppShell(main, Text("Footer"))
	}, 80, 10)
	defer h.Close()

	// Item 00 visible before scroll.
	if !h.Contains("Item 00") {
		t.Error("Item 00 should be visible before scrolling")
	}

	h.PageDown()

	// Item 00 scrolled off; footer still present.
	if h.Contains("Item 00") {
		t.Error("Item 00 should be off screen after scrolling down")
	}
	if !h.Contains("Footer") {
		t.Error("footer should remain on screen after scrolling")
	}
}

// ── Focus auto-scroll ─────────────────────────────────────────────────────────

func TestScroll_tabScrollsToOffScreenFocusable(t *testing.T) {
	// Build a component with two focusables separated by enough content that
	// the second one is below the initial viewport.
	filler := make([]Element, 20)
	for i := range filler {
		filler[i] = Text(fmt.Sprintf("Filler %02d", i))
	}
	h := NewHarnessSize(t, func() Element {
		return Box(
			append(
				[]Element{C(NewButton("Top Button", func() {}))},
				append(filler, C(NewButton("Bottom Button", func() {})))...,
			)...,
		)
	}, 80, 10)
	defer h.Close()

	// Initially Top Button is visible and focused.
	if !h.Contains("Top Button") {
		t.Error("Top Button should be visible initially")
	}
	if h.Contains("Bottom Button") {
		t.Error("Bottom Button should be off screen initially")
	}

	// Tab to Bottom Button — viewport should auto-scroll to show it.
	h.Tab()

	if !h.Contains("Bottom Button") {
		t.Errorf("Bottom Button should be visible after Tab; lines: %v", h.Lines())
	}
}
