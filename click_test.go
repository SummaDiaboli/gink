package gink_test

import (
	"testing"

	"github.com/SummaDiaboli/gink"
)

// labelComponent returns a component factory that renders "<name>:focused" or
// "<name>:blur" so focus state can be asserted via h.Contains.
func labelComponent(name string) func() gink.Element {
	return func() gink.Element {
		focused := gink.UseFocus()
		if focused {
			return gink.Text(name + ":focused")
		}
		return gink.Text(name + ":blur")
	}
}

// TestMouseClick_transfersFocusToClickedComponent verifies that clicking a
// component makes it the focused component.
func TestMouseClick_transfersFocusToClickedComponent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.C(labelComponent("A")),
			gink.C(labelComponent("B")),
		)
	})
	defer h.Close()

	if !h.Contains("A:focused") {
		t.Fatal("expected A to hold focus initially")
	}

	// B is rendered at row 1; click it.
	h.Click(0, 1)

	if !h.Contains("B:focused") {
		t.Error("expected B to be focused after click")
	}
	if !h.Contains("A:blur") {
		t.Error("expected A to lose focus after click on B")
	}
}

// TestMouseClick_noEffectOutsideFocusables verifies that clicking an empty
// area leaves focus unchanged.
func TestMouseClick_noEffectOutsideFocusables(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.C(labelComponent("A")), // only 1 row tall; rows 1+ are empty
		)
	})
	defer h.Close()

	h.Click(0, 5) // empty space

	if !h.Contains("A:focused") {
		t.Error("expected focus to remain on A after clicking empty space")
	}
}

// TestMouseClick_callsUseClickHandler verifies that UseClick fires when the
// component is clicked.
func TestMouseClick_callsUseClickHandler(t *testing.T) {
	clicked := false
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(func() gink.Element {
			gink.UseFocus()
			gink.UseClick(func(x, y int) {
				clicked = true
			})
			return gink.Size(5, 3, gink.Text("btn"))
		})
	})
	defer h.Close()

	h.Click(2, 1)

	if !clicked {
		t.Error("expected UseClick handler to be called on click")
	}
}

// TestMouseClick_localCoordinates verifies that UseClick receives coordinates
// relative to the component's top-left corner, not the screen.
func TestMouseClick_localCoordinates(t *testing.T) {
	var gotX, gotY int
	h := gink.NewHarness(t, func() gink.Element {
		return gink.C(func() gink.Element {
			gink.UseFocus()
			gink.UseClick(func(x, y int) {
				gotX = x
				gotY = y
			})
			return gink.Size(10, 5, gink.Text(""))
		})
	})
	defer h.Close()

	h.Click(3, 2)

	if gotX != 3 || gotY != 2 {
		t.Errorf("expected local coords (3, 2), got (%d, %d)", gotX, gotY)
	}
}

// TestMouseClick_rowLayoutFocusesCorrectComponent verifies that clicking within
// a horizontal row layout focuses the component at the clicked column, not
// the first component.
func TestMouseClick_rowLayoutFocusesCorrectComponent(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		// A occupies cols 0-9; B occupies cols 10-19.
		return gink.Row(
			gink.Width(10, gink.C(labelComponent("A"))),
			gink.Width(10, gink.C(labelComponent("B"))),
		)
	})
	defer h.Close()

	// Click inside B's area at column 12.
	h.Click(12, 0)

	if !h.Contains("B:focused") {
		t.Errorf("expected B to be focused after clicking col 12, screen: %v", h.Lines())
	}
	if !h.Contains("A:blur") {
		t.Error("expected A to lose focus after click on B")
	}
}

// TestMouseClick_useClickNotCalledWhenClickMisses verifies that a UseClick
// handler is not called when the click lands outside the component's bounds.
func TestMouseClick_useClickNotCalledWhenClickMisses(t *testing.T) {
	called := false
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.C(func() gink.Element {
				gink.UseFocus()
				gink.UseClick(func(x, y int) { called = true })
				return gink.Size(5, 1, gink.Text("A"))
			}),
			gink.C(func() gink.Element {
				gink.UseFocus()
				gink.UseClick(func(x, y int) { called = true })
				return gink.Size(5, 1, gink.Text("B"))
			}),
		)
	})
	defer h.Close()

	// Click at row 5 — outside both components.
	h.Click(0, 5)

	if called {
		t.Error("expected UseClick not to fire when clicking outside all components")
	}
}
