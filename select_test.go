package gink

import "testing"

// ── NewSelect ─────────────────────────────────────────────────────────────────

var selectOpts = []string{"Apple", "Banana", "Cherry"}

// selectHarness creates a Harness whose root component owns the selection via
// UseState so that calling onChange triggers a proper reconciler re-render.
func selectHarness(t *testing.T, initial string) (*Harness, *string) {
	t.Helper()
	var lastVal string
	h := NewHarness(t, func() Element {
		val, setVal := UseState(initial)
		lastVal = val
		return C(NewSelect(selectOpts, val, func(v string) { setVal(v) }))
	})
	return h, &lastVal
}

func TestNewSelect_rendersCurrentValue(t *testing.T) {
	h, _ := selectHarness(t, "Banana")
	defer h.Close()

	if !h.Contains("Banana") {
		t.Errorf("expected current value to be rendered; line: %q", h.Line(0))
	}
}

func TestNewSelect_showsFocusIndicatorWhenFocused(t *testing.T) {
	h, _ := selectHarness(t, "Apple")
	defer h.Close()

	// First component is focused by default — navigation arrows must be visible.
	if !h.Contains("◀") || !h.Contains("▶") {
		t.Errorf("focused select should show navigation arrows; line: %q", h.Line(0))
	}
}

func TestNewSelect_noArrowsWhenUnfocused(t *testing.T) {
	h := NewHarness(t, func() Element {
		val, setVal := UseState("Apple")
		return Box(
			C(NewButton("First", func() {})), // takes focus ahead of the Select
			C(NewSelect(selectOpts, val, func(v string) { setVal(v) })),
		)
	})
	defer h.Close()

	if h.Contains("◀") || h.Contains("▶") {
		t.Errorf("unfocused select must not show navigation arrows; screen: %v", h.Lines())
	}
}

func TestNewSelect_rightMovesToNextOption(t *testing.T) {
	h, lastVal := selectHarness(t, "Apple")
	defer h.Close()

	h.SendKey(KeyRight)

	if *lastVal != "Banana" {
		t.Errorf("after Right: got %q, want Banana", *lastVal)
	}
}

func TestNewSelect_leftMovesToPreviousOption(t *testing.T) {
	h, lastVal := selectHarness(t, "Banana")
	defer h.Close()

	h.SendKey(KeyLeft)

	if *lastVal != "Apple" {
		t.Errorf("after Left: got %q, want Apple", *lastVal)
	}
}

func TestNewSelect_rightNoOpAtLastOption(t *testing.T) {
	h, lastVal := selectHarness(t, "Cherry")
	defer h.Close()

	h.SendKey(KeyRight)

	if *lastVal != "Cherry" {
		t.Errorf("Right at last option should be no-op; got %q", *lastVal)
	}
}

func TestNewSelect_leftNoOpAtFirstOption(t *testing.T) {
	h, lastVal := selectHarness(t, "Apple")
	defer h.Close()

	h.SendKey(KeyLeft)

	if *lastVal != "Apple" {
		t.Errorf("Left at first option should be no-op; got %q", *lastVal)
	}
}

func TestNewSelect_ignoredWhenUnfocused(t *testing.T) {
	var lastVal string
	h := NewHarness(t, func() Element {
		val, setVal := UseState("Apple")
		lastVal = val
		return Box(
			C(NewButton("First", func() {})), // takes focus
			C(NewSelect(selectOpts, val, func(v string) { setVal(v) })),
		)
	})
	defer h.Close()

	h.SendKey(KeyRight)

	if lastVal != "Apple" {
		t.Errorf("unfocused select should not respond to keys; got %q", lastVal)
	}
}

func TestNewSelect_tabCyclesFocusToSelect(t *testing.T) {
	h, lastVal := func() (*Harness, *string) {
		var lastVal string
		h := NewHarness(t, func() Element {
			val, setVal := UseState("Apple")
			lastVal = val
			return Box(
				C(NewButton("First", func() {})),
				C(NewSelect(selectOpts, val, func(v string) { setVal(v) })),
			)
		})
		return h, &lastVal
	}()
	defer h.Close()

	h.Tab() // move focus to the Select
	h.SendKey(KeyRight)

	if *lastVal != "Banana" {
		t.Errorf("after Tab+Right: got %q, want Banana", *lastVal)
	}
}

func TestNewSelect_customFocusStyle(t *testing.T) {
	custom := NewStyle().Bold().Foreground(ColorBrightGreen)
	h, _ := func() (*Harness, *string) {
		var lastVal string
		h := NewHarness(t, func() Element {
			val, setVal := UseState("Apple")
			lastVal = val
			return C(NewSelect(selectOpts, val, func(v string) { setVal(v) }, custom))
		})
		return h, &lastVal
	}()
	defer h.Close()

	// Select is focused by default — check the cell style matches the custom style.
	got := h.CellStyle(0, 0)
	if got != custom.toTcell() {
		t.Errorf("custom focus style not applied: got %v, want %v", got, custom.toTcell())
	}
}

func TestNewSelect_clickRightHalfAdvances(t *testing.T) {
	h, lastVal := selectHarness(t, "Apple")
	defer h.Close()

	// Focused: "[ ◀ Apple ▶ ]" — 13 chars, midpoint 6. Click col 8 (right half).
	h.Click(8, 0)

	if *lastVal != "Banana" {
		t.Errorf("click right half: got %q, want Banana", *lastVal)
	}
}

func TestNewSelect_clickLeftHalfGoesBack(t *testing.T) {
	h, lastVal := selectHarness(t, "Banana")
	defer h.Close()

	// Focused: "[ ◀ Banana ▶ ]" — 14 chars, midpoint 7. Click col 2 (left half).
	h.Click(2, 0)

	if *lastVal != "Apple" {
		t.Errorf("click left half: got %q, want Apple", *lastVal)
	}
}

func TestNewSelect_clickRightNoOpAtLastOption(t *testing.T) {
	h, lastVal := selectHarness(t, "Cherry")
	defer h.Close()

	h.Click(12, 0) // right half — already at last option

	if *lastVal != "Cherry" {
		t.Errorf("click right at last option should be no-op; got %q", *lastVal)
	}
}

func TestNewSelect_unknownValueDefaultsToFirst(t *testing.T) {
	// When value is not in options, Down should behave as if at index 0.
	h, lastVal := selectHarness(t, "Unknown")
	defer h.Close()

	h.SendKey(KeyRight)

	if *lastVal != "Banana" {
		t.Errorf("unknown value: after Right got %q, want Banana", *lastVal)
	}
}
