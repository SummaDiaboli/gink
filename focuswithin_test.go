package gink_test

import (
	"strings"
	"testing"

	"github.com/SummaDiaboli/gink"
)

// focusWithinInner is a leaf focusable used to exercise UseFocusWithin.
func focusWithinInner() gink.Element {
	gink.UseFocus()
	return gink.Text("inner")
}

// focusWithinPanel wraps focusWithinInner and reports its own focus-within
// state so tests can assert on the rendered label.
func focusWithinPanel() gink.Element {
	within := gink.UseFocusWithin()
	label := "panel:unfocused"
	if within {
		label = "panel:focused"
	}
	return gink.Box(gink.Text(label), gink.C(focusWithinInner))
}

// TestUseFocusWithin_trueWhenDescendantFocused verifies that UseFocusWithin
// returns true when the focused component is a descendant of the caller.
func TestUseFocusWithin_trueWhenDescendantFocused(t *testing.T) {
	h := gink.NewHarness(t, focusWithinPanel)
	defer h.Close()
	h.Render() // second pass so prevFocusables is seeded

	if !h.Contains("panel:focused") {
		t.Errorf("expected panel:focused\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}
}

// TestUseFocusWithin_falseWhenSiblingFocused verifies that UseFocusWithin
// returns false when a sibling (not a descendant) holds focus.
func TestUseFocusWithin_falseWhenSiblingFocused(t *testing.T) {
	siblingFocusable := func() gink.Element {
		gink.UseFocus()
		return gink.Text("sibling")
	}
	panelB := func() gink.Element {
		within := gink.UseFocusWithin()
		if within {
			return gink.Text("B:focused")
		}
		return gink.Text("B:unfocused")
	}
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(gink.C(siblingFocusable), gink.C(panelB))
	})
	defer h.Close()
	h.Render()

	if !h.Contains("B:unfocused") {
		t.Errorf("expected B:unfocused when sibling holds focus\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}
}

// TestUseFocusWithin_trueAfterTabToDescendant verifies that UseFocusWithin
// updates correctly when Tab moves focus into the component's subtree.
func TestUseFocusWithin_trueAfterTabToDescendant(t *testing.T) {
	otherFocusable := func() gink.Element {
		gink.UseFocus()
		return gink.Text("other")
	}
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(gink.C(otherFocusable), gink.C(focusWithinPanel))
	})
	defer h.Close()
	h.Render()

	// Initially other is focused; panel should not report focused-within.
	if h.Contains("panel:focused") {
		t.Error("expected panel:unfocused when other holds initial focus")
	}

	h.Tab() // moves focus to focusWithinInner inside the panel
	h.Render()

	if !h.Contains("panel:focused") {
		t.Errorf("expected panel:focused after Tab\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}
}

// TestUseFocusWithin_falseAfterTabAway verifies that UseFocusWithin resets to
// false when Tab moves focus out of the subtree.
func TestUseFocusWithin_falseAfterTabAway(t *testing.T) {
	otherFocusable := func() gink.Element {
		gink.UseFocus()
		return gink.Text("other")
	}
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(gink.C(focusWithinPanel), gink.C(otherFocusable))
	})
	defer h.Close()
	h.Render()

	// focusWithinInner is focused first; panel should report focused-within.
	if !h.Contains("panel:focused") {
		t.Errorf("expected panel:focused initially\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}

	h.Tab() // moves focus to other
	h.Render()

	if h.Contains("panel:focused") {
		t.Errorf("expected panel:unfocused after Tab away\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}
}
