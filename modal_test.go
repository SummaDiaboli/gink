package gink

import (
	"strings"
	"testing"
)

// modalHarness builds a layout with an outside button plus a modal, so that
// focus-trap behaviour can be observed.
func modalHarness(t *testing.T, onClose func(), actions []ModalAction) *Harness {
	t.Helper()
	return NewHarness(t, func() Element {
		return Box(
			C(NewButton("Outside", func() {})),
			C(NewModal("Dialog", Text("Modal content"), actions, onClose)),
		)
	})
}

func TestNewModal_rendersTitle(t *testing.T) {
	h := modalHarness(t, func() {}, nil)
	defer h.Close()

	if !h.Contains("Dialog") {
		t.Errorf("modal title not rendered; lines: %v", h.Lines())
	}
}

func TestNewModal_rendersContent(t *testing.T) {
	h := modalHarness(t, func() {}, nil)
	defer h.Close()

	if !h.Contains("Modal content") {
		t.Errorf("modal content not rendered; lines: %v", h.Lines())
	}
}

func TestNewModal_hasBorder(t *testing.T) {
	h := modalHarness(t, func() {}, nil)
	defer h.Close()

	found := false
	for _, line := range h.Lines() {
		if strings.ContainsAny(line, "┌┐└┘│─") {
			found = true
			break
		}
	}
	if !found {
		t.Error("modal should have a border")
	}
}

func TestNewModal_rendersActionButtons(t *testing.T) {
	actions := []ModalAction{
		{Label: "OK", OnPress: func() {}},
		{Label: "Cancel", OnPress: func() {}},
	}
	h := modalHarness(t, func() {}, actions)
	defer h.Close()

	if !h.Contains("OK") {
		t.Error("OK button not rendered")
	}
	if !h.Contains("Cancel") {
		t.Error("Cancel button not rendered")
	}
}

func TestNewModal_escFiresOnClose(t *testing.T) {
	closed := false
	h := modalHarness(t, func() { closed = true }, nil)
	defer h.Close()

	h.SendKey(KeyEscape)

	if !closed {
		t.Error("Esc should fire onClose")
	}
}

func TestNewModal_actionButtonFiresCallback(t *testing.T) {
	pressed := false
	actions := []ModalAction{
		{Label: "Confirm", OnPress: func() { pressed = true }},
	}
	h := modalHarness(t, func() {}, actions)
	defer h.Close()

	// Focus starts inside the modal (barrier snaps it in). Press Enter.
	h.SendKey(KeyEnter)

	if !pressed {
		t.Error("action button should fire its callback on Enter")
	}
}

func TestNewModal_focusTrapPreventsTabEscape(t *testing.T) {
	actions := []ModalAction{
		{Label: "A", OnPress: func() {}},
		{Label: "B", OnPress: func() {}},
	}
	h := modalHarness(t, func() {}, actions)
	defer h.Close()

	// Tab through several times — focus must never leave the modal.
	for i := 0; i < 6; i++ {
		h.Tab()
		for _, line := range h.Lines() {
			if strings.Contains(line, "[ Outside ]") {
				// Check Outside button is not focused (no focus style on it).
				x := strings.Index(line, "Outside")
				y := 0
				for row, l := range h.Lines() {
					if l == line {
						y = row
						break
					}
				}
				if h.CellStyle(x, y) != (Style{}).toTcell() {
					t.Errorf("Tab iteration %d: Outside button should not be focused", i+1)
				}
			}
		}
	}
}

func TestNewModal_focusSnapsIntoModalOnOpen(t *testing.T) {
	actions := []ModalAction{
		{Label: "OK", OnPress: func() {}},
	}
	h := modalHarness(t, func() {}, actions)
	defer h.Close()

	// On initial render the focus barrier should have snapped focus into the
	// modal. Pressing Enter should fire the OK action, not the Outside button.
	pressed := false
	// Rebuild with a trackable action.
	h2 := NewHarness(t, func() Element {
		return Box(
			C(NewButton("Outside", func() {})),
			C(NewModal("D", Text(""), []ModalAction{{Label: "OK", OnPress: func() { pressed = true }}}, func() {})),
		)
	})
	defer h2.Close()

	h2.SendKey(KeyEnter)

	if !pressed {
		t.Error("focus should snap into modal on open; Enter should fire modal action")
	}
}
