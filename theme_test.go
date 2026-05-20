package gink

import "testing"

// TestUseTheme_returnsDefaultWhenUnset verifies that UseTheme returns DefaultTheme
// before any custom theme is applied.
func TestUseTheme_returnsDefaultWhenUnset(t *testing.T) {
	var got Theme
	h := NewHarness(t, func() Element {
		got = UseTheme()
		return Text("")
	})
	defer h.Close()

	if got != DefaultTheme {
		t.Errorf("expected DefaultTheme, got %+v", got)
	}
}

// TestUseTheme_customThemePropagates verifies that a theme set via SetContext
// is returned by UseTheme in the same and descendant components.
func TestUseTheme_customThemePropagates(t *testing.T) {
	custom := DefaultTheme
	custom.Focused = NewStyle().Bold().Foreground(ColorBrightMagenta)

	var got Theme
	h := NewHarness(t, func() Element {
		SetContext(ThemeCtx, custom)
		got = UseTheme()
		return Text("")
	})
	defer h.Close()

	if got.Focused != custom.Focused {
		t.Errorf("expected custom Focused style, got %+v", got.Focused)
	}
}

// TestUseTheme_inheritedByDescendants verifies that a theme set in a parent
// component is visible to child components via UseTheme.
func TestUseTheme_inheritedByDescendants(t *testing.T) {
	custom := DefaultTheme
	custom.Accent = NewStyle().Bold().Foreground(ColorBrightYellow)

	var childGot Theme
	h := NewHarness(t, func() Element {
		SetContext(ThemeCtx, custom)
		return C(func() Element {
			childGot = UseTheme()
			return Text("")
		})
	})
	defer h.Close()

	if childGot.Accent != custom.Accent {
		t.Errorf("child expected custom Accent, got %+v", childGot.Accent)
	}
}

// TestNewList_usesThemeFocusStyle verifies that a NewList with no explicit style
// applies the theme's Focused style to the selected row when focused.
func TestNewList_usesThemeFocusStyle(t *testing.T) {
	custom := DefaultTheme
	custom.Focused = NewStyle().Bold().Foreground(ColorBrightMagenta)

	h := NewHarness(t, func() Element {
		SetContext(ThemeCtx, custom)
		sel, setSel := UseState(0)
		return C(NewList([]string{"alpha", "beta"}, sel, setSel, 4))
	})
	defer h.Close()

	// Row 0 is the selected+focused row; its style should match custom.Focused.
	got := h.CellStyle(2, 0) // col 2 is inside "alpha" text, past the cursor
	if got != custom.Focused.toTcell() {
		t.Errorf("NewList selected row style: got %v, want %v", got, custom.Focused.toTcell())
	}
}

// TestNewButton_usesThemeFocusStyle verifies that a focused NewButton with no
// explicit style renders using the theme's Focused style.
func TestNewButton_usesThemeFocusStyle(t *testing.T) {
	custom := DefaultTheme
	custom.Focused = NewStyle().Bold().Foreground(ColorBrightMagenta)

	h := NewHarness(t, func() Element {
		SetContext(ThemeCtx, custom)
		return C(NewButton("OK", func() {}))
	})
	defer h.Close()

	got := h.CellStyle(0, 0)
	if got != custom.Focused.toTcell() {
		t.Errorf("NewButton focused style: got %v, want %v", got, custom.Focused.toTcell())
	}
}
