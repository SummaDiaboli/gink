package gink

import (
	"strings"
	"testing"
)

// checkboxHarness creates a Harness whose root component owns the checked state.
func checkboxHarness(t *testing.T, initial bool) (*Harness, *bool, func(bool)) {
	t.Helper()
	var lastChecked bool
	var setChecked func(bool)

	h := NewHarness(t, func() Element {
		checked, setCheckedFn := UseState(initial)
		lastChecked = checked
		setChecked = setCheckedFn
		return C(NewCheckbox("Option", checked, func(v bool) { setCheckedFn(v) }))
	})
	return h, &lastChecked, setChecked
}

func TestNewCheckbox_rendersUnchecked(t *testing.T) {
	h, _, _ := checkboxHarness(t, false)
	defer h.Close()

	line := h.Line(0)
	if !strings.Contains(line, "[ ]") {
		t.Errorf("unchecked checkbox should render [ ]; got %q", line)
	}
	if !strings.Contains(line, "Option") {
		t.Errorf("label not rendered; got %q", line)
	}
}

func TestNewCheckbox_rendersChecked(t *testing.T) {
	h, _, _ := checkboxHarness(t, true)
	defer h.Close()

	line := h.Line(0)
	if !strings.Contains(line, "[x]") {
		t.Errorf("checked checkbox should render [x]; got %q", line)
	}
}

func TestNewCheckbox_spaceTogglesOff(t *testing.T) {
	h, lastChecked, _ := checkboxHarness(t, true)
	defer h.Close()

	h.SendRune(' ')

	if *lastChecked {
		t.Error("Space should toggle checked→unchecked")
	}
}

func TestNewCheckbox_spaceTogglesOn(t *testing.T) {
	h, lastChecked, _ := checkboxHarness(t, false)
	defer h.Close()

	h.SendRune(' ')

	if !*lastChecked {
		t.Error("Space should toggle unchecked→checked")
	}
}

func TestNewCheckbox_enterToggles(t *testing.T) {
	h, lastChecked, _ := checkboxHarness(t, false)
	defer h.Close()

	h.SendKey(KeyEnter)

	if !*lastChecked {
		t.Error("Enter should toggle unchecked→checked")
	}
}

func TestNewCheckbox_clickToggles(t *testing.T) {
	h, lastChecked, _ := checkboxHarness(t, false)
	defer h.Close()

	h.Click(0, 0)

	if !*lastChecked {
		t.Error("click should toggle unchecked→checked")
	}
}

func TestNewCheckbox_ignoredWhenUnfocused(t *testing.T) {
	var lastChecked bool
	h := NewHarness(t, func() Element {
		checked, setChecked := UseState(false)
		lastChecked = checked
		return Box(
			C(NewButton("Other", func() {})), // takes focus
			C(NewCheckbox("Option", checked, func(v bool) { setChecked(v) })),
		)
	})
	defer h.Close()

	h.SendKey(KeyEnter)

	if lastChecked {
		t.Error("unfocused checkbox should ignore keypresses")
	}
}

func TestNewCheckbox_focusStyleApplied(t *testing.T) {
	h, _, _ := checkboxHarness(t, false)
	defer h.Close()

	got := h.CellStyle(0, 0)
	if got == (Style{}).toTcell() {
		t.Error("focused checkbox should have non-default style")
	}
}

func TestNewCheckbox_customStyle(t *testing.T) {
	custom := NewStyle().Bold().Foreground(ColorBrightGreen)
	h := NewHarness(t, func() Element {
		checked, setChecked := UseState(false)
		return C(NewCheckbox("Option", checked, func(v bool) { setChecked(v) }, custom))
	})
	defer h.Close()

	got := h.CellStyle(0, 0)
	if got != custom.toTcell() {
		t.Errorf("custom style not applied: got %v, want %v", got, custom.toTcell())
	}
}
