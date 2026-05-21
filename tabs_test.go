package gink

import (
	"strings"
	"testing"
)

var tabLabels = []string{"Files", "Git", "Terminal"}

// tabsHarness creates a Harness whose root owns the active tab via UseState.
func tabsHarness(t *testing.T, initial string) (*Harness, *string, func(string)) {
	t.Helper()
	var lastActive string
	var setActive func(string)

	h := NewHarness(t, func() Element {
		active, setActiveFn := UseState(initial)
		lastActive = active
		setActive = setActiveFn
		return C(NewTabs(tabLabels, active, func(s string) { setActiveFn(s) }))
	})
	return h, &lastActive, setActive
}

func TestNewTabs_rendersAllLabels(t *testing.T) {
	h, _, _ := tabsHarness(t, "Files")
	defer h.Close()

	line := h.Line(0)
	for _, label := range tabLabels {
		if !strings.Contains(line, label) {
			t.Errorf("tab %q not rendered; line: %q", label, line)
		}
	}
}

func TestNewTabs_rightMovesToNextTab(t *testing.T) {
	h, lastActive, _ := tabsHarness(t, "Files")
	defer h.Close()

	h.SendKey(KeyRight)

	if *lastActive != "Git" {
		t.Errorf("Right should move to Git; got %q", *lastActive)
	}
}

func TestNewTabs_leftMovesToPrevTab(t *testing.T) {
	h, lastActive, _ := tabsHarness(t, "Terminal")
	defer h.Close()

	h.SendKey(KeyLeft)

	if *lastActive != "Git" {
		t.Errorf("Left should move to Git; got %q", *lastActive)
	}
}

func TestNewTabs_rightNoOpAtLastTab(t *testing.T) {
	h, lastActive, _ := tabsHarness(t, "Terminal")
	defer h.Close()

	h.SendKey(KeyRight)

	if *lastActive != "Terminal" {
		t.Errorf("Right at last tab should be no-op; got %q", *lastActive)
	}
}

func TestNewTabs_leftNoOpAtFirstTab(t *testing.T) {
	h, lastActive, _ := tabsHarness(t, "Files")
	defer h.Close()

	h.SendKey(KeyLeft)

	if *lastActive != "Files" {
		t.Errorf("Left at first tab should be no-op; got %q", *lastActive)
	}
}

func TestNewTabs_clickSelectsTab(t *testing.T) {
	h, lastActive, _ := tabsHarness(t, "Files")
	defer h.Close()

	// Find the column position of "Terminal" and click it.
	line := h.Line(0)
	x := strings.Index(line, "Terminal")
	if x < 0 {
		t.Fatal("Terminal not found in rendered line")
	}
	h.Click(x, 0)

	if *lastActive != "Terminal" {
		t.Errorf("clicking Terminal should select it; got %q", *lastActive)
	}
}

func TestNewTabs_ignoredWhenUnfocused(t *testing.T) {
	var lastActive string
	h := NewHarness(t, func() Element {
		active, setActive := UseState("Files")
		lastActive = active
		return Box(
			C(NewButton("Other", func() {})), // takes focus
			C(NewTabs(tabLabels, active, func(s string) { setActive(s) })),
		)
	})
	defer h.Close()

	h.SendKey(KeyRight)

	if lastActive != "Files" {
		t.Errorf("unfocused tabs should ignore keypresses; got %q", lastActive)
	}
}

func TestNewTabs_activeTabHasFocusStyle(t *testing.T) {
	h, _, _ := tabsHarness(t, "Files")
	defer h.Close()

	// The active tab cell should carry the focused style.
	line := h.Line(0)
	x := strings.Index(line, "Files")
	if x < 0 {
		t.Fatal("Files not found in rendered line")
	}
	got := h.CellStyle(x, 0)
	if got == (Style{}).toTcell() {
		t.Error("active tab should have non-default style when focused")
	}
}

func TestNewTabs_customStyle(t *testing.T) {
	custom := NewStyle().Bold().Foreground(ColorBrightGreen)
	h := NewHarness(t, func() Element {
		active, setActive := UseState("Files")
		return C(NewTabs(tabLabels, active, func(s string) { setActive(s) }, custom))
	})
	defer h.Close()

	line := h.Line(0)
	x := strings.Index(line, "Files")
	if x < 0 {
		t.Fatal("Files not found in rendered line")
	}
	got := h.CellStyle(x, 0)
	if got != custom.toTcell() {
		t.Errorf("custom style not applied: got %v, want %v", got, custom.toTcell())
	}
}
