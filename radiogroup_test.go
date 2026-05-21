package gink

import (
	"strings"
	"testing"
)

var radioOptions = []string{"Red", "Green", "Blue"}

// radioHarness creates a Harness whose root owns the selected string via UseState.
func radioHarness(t *testing.T, initial string) (*Harness, *string, func(string)) {
	t.Helper()
	var lastSel string
	var setSel func(string)

	h := NewHarness(t, func() Element {
		sel, setSelFn := UseState(initial)
		lastSel = sel
		setSel = setSelFn
		return C(NewRadioGroup(radioOptions, sel, func(s string) { setSelFn(s) }))
	})
	return h, &lastSel, setSel
}

func TestNewRadioGroup_rendersAllOptions(t *testing.T) {
	h, _, _ := radioHarness(t, "Red")
	defer h.Close()

	for _, opt := range radioOptions {
		if !h.Contains(opt) {
			t.Errorf("option %q not rendered; lines: %v", opt, h.Lines())
		}
	}
}

func TestNewRadioGroup_marksSelectedOption(t *testing.T) {
	h, _, _ := radioHarness(t, "Green")
	defer h.Close()

	found := false
	for _, line := range h.Lines() {
		if strings.Contains(line, "(●)") && strings.Contains(line, "Green") {
			found = true
		}
	}
	if !found {
		t.Errorf("selected option should have (●) glyph; lines: %v", h.Lines())
	}
}

func TestNewRadioGroup_unselectedOptionsHaveEmptyGlyph(t *testing.T) {
	h, _, _ := radioHarness(t, "Red")
	defer h.Close()

	for _, line := range h.Lines() {
		if strings.Contains(line, "Green") || strings.Contains(line, "Blue") {
			if !strings.Contains(line, "( )") {
				t.Errorf("unselected option should have ( ) glyph; line: %q", line)
			}
		}
	}
}

func TestNewRadioGroup_downMovesSelection(t *testing.T) {
	h, lastSel, _ := radioHarness(t, "Red")
	defer h.Close()

	h.SendKey(KeyDown)

	if *lastSel != "Green" {
		t.Errorf("Down should move to Green; got %q", *lastSel)
	}
}

func TestNewRadioGroup_upMovesSelection(t *testing.T) {
	h, lastSel, _ := radioHarness(t, "Blue")
	defer h.Close()

	h.SendKey(KeyUp)

	if *lastSel != "Green" {
		t.Errorf("Up should move to Green; got %q", *lastSel)
	}
}

func TestNewRadioGroup_downNoOpAtLastOption(t *testing.T) {
	h, lastSel, _ := radioHarness(t, "Blue")
	defer h.Close()

	h.SendKey(KeyDown)

	if *lastSel != "Blue" {
		t.Errorf("Down at last option should be no-op; got %q", *lastSel)
	}
}

func TestNewRadioGroup_upNoOpAtFirstOption(t *testing.T) {
	h, lastSel, _ := radioHarness(t, "Red")
	defer h.Close()

	h.SendKey(KeyUp)

	if *lastSel != "Red" {
		t.Errorf("Up at first option should be no-op; got %q", *lastSel)
	}
}

func TestNewRadioGroup_clickSelectsOption(t *testing.T) {
	h, lastSel, _ := radioHarness(t, "Red")
	defer h.Close()

	// Blue is on the third row (localY=2).
	h.Click(0, 2)

	if *lastSel != "Blue" {
		t.Errorf("clicking row 2 should select Blue; got %q", *lastSel)
	}
}

func TestNewRadioGroup_ignoredWhenUnfocused(t *testing.T) {
	var lastSel string
	h := NewHarness(t, func() Element {
		sel, setSel := UseState("Red")
		lastSel = sel
		return Box(
			C(NewButton("Other", func() {})), // takes focus
			C(NewRadioGroup(radioOptions, sel, func(s string) { setSel(s) })),
		)
	})
	defer h.Close()

	h.SendKey(KeyDown)

	if lastSel != "Red" {
		t.Errorf("unfocused radio group should ignore keypresses; got %q", lastSel)
	}
}

func TestNewRadioGroup_focusStyleOnSelectedRow(t *testing.T) {
	h, _, _ := radioHarness(t, "Red")
	defer h.Close()

	got := h.CellStyle(0, 0)
	if got == (Style{}).toTcell() {
		t.Error("focused selected row should have non-default style")
	}
}

func TestNewRadioGroup_customStyle(t *testing.T) {
	custom := NewStyle().Bold().Foreground(ColorBrightGreen)
	h := NewHarness(t, func() Element {
		sel, setSel := UseState("Red")
		return C(NewRadioGroup(radioOptions, sel, func(s string) { setSel(s) }, custom))
	})
	defer h.Close()

	got := h.CellStyle(0, 0)
	if got != custom.toTcell() {
		t.Errorf("custom style not applied: got %v, want %v", got, custom.toTcell())
	}
}
