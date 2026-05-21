package gink

import (
	"strings"
	"testing"
)

type testMenuItem struct {
	Label    string
	Key      rune
	Disabled bool
}

var menuItems = []MenuItem{
	{Label: "New File", Key: 'n'},
	{Label: "Open",     Key: 'o'},
	{Label: "Save",     Key: 's'},
	{Label: "Quit",     Key: 'q', Disabled: true},
}

// menuHarness creates a Harness with a NewMenu. lastSel receives the most
// recently selected item; closed is set when Esc fires onClose.
func menuHarness(t *testing.T) (*Harness, *MenuItem, *bool) {
	t.Helper()
	var lastSel MenuItem
	closed := false

	h := NewHarness(t, func() Element {
		return C(NewMenu(menuItems, func(item MenuItem) { lastSel = item }, func() { closed = true }))
	})
	return h, &lastSel, &closed
}

func TestNewMenu_rendersAllEnabledItems(t *testing.T) {
	h, _, _ := menuHarness(t)
	defer h.Close()

	for _, item := range menuItems {
		if !h.Contains(item.Label) {
			t.Errorf("item %q not rendered; lines: %v", item.Label, h.Lines())
		}
	}
}

func TestNewMenu_rendersKeyHint(t *testing.T) {
	h, _, _ := menuHarness(t)
	defer h.Close()

	// Each item should show its key hint somewhere on its line.
	for _, item := range menuItems {
		for _, line := range h.Lines() {
			if strings.Contains(line, item.Label) {
				if !strings.ContainsRune(line, item.Key) {
					t.Errorf("key hint %q not shown for %q; line: %q", item.Key, item.Label, line)
				}
				break
			}
		}
	}
}

func TestNewMenu_hasBorder(t *testing.T) {
	h, _, _ := menuHarness(t)
	defer h.Close()

	// A border character should appear on the first line.
	first := h.Line(0)
	if !strings.ContainsAny(first, "┌─┐") {
		t.Errorf("expected top border on first line; got %q", first)
	}
}

func TestNewMenu_downMovesHighlight(t *testing.T) {
	h, _, _ := menuHarness(t)
	defer h.Close()

	h.SendKey(KeyDown)

	// After one Down, "Open" should be highlighted (non-default style on that row).
	for i, line := range h.Lines() {
		if strings.Contains(line, "Open") {
			if h.CellStyle(2, i) == (Style{}).toTcell() {
				t.Error("Open should be highlighted after Down")
			}
			return
		}
	}
	t.Error("Open not found in rendered lines")
}

func TestNewMenu_enterSelectsItem(t *testing.T) {
	h, lastSel, _ := menuHarness(t)
	defer h.Close()

	h.SendKey(KeyEnter)

	if lastSel.Label != "New File" {
		t.Errorf("Enter should select highlighted item; got %q", lastSel.Label)
	}
}

func TestNewMenu_downThenEnterSelectsSecondItem(t *testing.T) {
	h, lastSel, _ := menuHarness(t)
	defer h.Close()

	h.SendKey(KeyDown)
	h.SendKey(KeyEnter)

	if lastSel.Label != "Open" {
		t.Errorf("expected Open; got %q", lastSel.Label)
	}
}

func TestNewMenu_escFiresOnClose(t *testing.T) {
	h, _, closed := menuHarness(t)
	defer h.Close()

	h.SendKey(KeyEscape)

	if !*closed {
		t.Error("Esc should fire onClose")
	}
}

func TestNewMenu_disabledItemSkippedByDown(t *testing.T) {
	h, lastSel, _ := menuHarness(t)
	defer h.Close()

	// Navigate to Save (index 2), then Down should skip Quit (disabled).
	h.SendKey(KeyDown)
	h.SendKey(KeyDown)
	h.SendKey(KeyDown) // would land on Quit — should skip or stop

	h.SendKey(KeyEnter)

	if lastSel.Label == "Quit" {
		t.Error("disabled item should not be selectable")
	}
}

func TestNewMenu_clickSelectsItem(t *testing.T) {
	h, lastSel, _ := menuHarness(t)
	defer h.Close()

	// Find the row that contains "Open" and click it.
	for i, line := range h.Lines() {
		if strings.Contains(line, "Open") {
			h.Click(2, i)
			break
		}
	}

	if lastSel.Label != "Open" {
		t.Errorf("clicking Open should select it; got %q", lastSel.Label)
	}
}

func TestNewMenu_clickDisabledItemDoesNotSelect(t *testing.T) {
	h, lastSel, _ := menuHarness(t)
	defer h.Close()

	for i, line := range h.Lines() {
		if strings.Contains(line, "Quit") {
			h.Click(2, i)
			break
		}
	}

	if lastSel.Label == "Quit" {
		t.Error("clicking disabled item should not fire onSelect")
	}
}
