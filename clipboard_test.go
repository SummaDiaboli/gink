package gink

import "testing"

// TestUseClipboard_readWrite verifies that write followed by read returns the
// same string via the hook's returned functions.
func TestUseClipboard_readWrite(t *testing.T) {
	// Intercept clipboard calls so the test does not touch the system clipboard.
	var stored string
	clipRead = func() (string, error) { return stored, nil }
	clipWrite = func(s string) error { stored = s; return nil }
	t.Cleanup(func() { clipRead = defaultClipRead; clipWrite = defaultClipWrite })

	var read func() string
	var write func(string)
	h := NewHarness(t, func() Element {
		read, write = UseClipboard()
		return Text("")
	})
	defer h.Close()

	write("hello clipboard")
	if got := read(); got != "hello clipboard" {
		t.Errorf("expected %q, got %q", "hello clipboard", got)
	}
}

// TestNewInput_pasteCtrlV verifies that Ctrl+V appends the clipboard content
// to the current input value.
func TestNewInput_pasteCtrlV(t *testing.T) {
	clipRead = func() (string, error) { return "pasted", nil }
	clipWrite = func(s string) error { return nil }
	t.Cleanup(func() { clipRead = defaultClipRead; clipWrite = defaultClipWrite })

	value := ""
	setValue := func(s string) { value = s }
	h := NewHarness(t, func() Element {
		return C(NewInput(value, setValue))
	})
	defer h.Close()

	h.SendKey(KeyCtrlV)

	if value != "pasted" {
		t.Errorf("expected %q after Ctrl+V, got %q", "pasted", value)
	}
}

// TestNewInput_pasteStripsNewlines verifies that multi-line clipboard content
// has its newlines replaced with spaces when pasted into a single-line input.
func TestNewInput_pasteStripsNewlines(t *testing.T) {
	clipRead = func() (string, error) { return "line1\nline2", nil }
	clipWrite = func(s string) error { return nil }
	t.Cleanup(func() { clipRead = defaultClipRead; clipWrite = defaultClipWrite })

	value := ""
	setValue := func(s string) { value = s }
	h := NewHarness(t, func() Element {
		return C(NewInput(value, setValue))
	})
	defer h.Close()

	h.SendKey(KeyCtrlV)

	if value != "line1 line2" {
		t.Errorf("expected %q after Ctrl+V, got %q", "line1 line2", value)
	}
}

// TestNewTextArea_pasteSingleLine verifies that single-line clipboard content
// is inserted at the cursor position in a textarea.
func TestNewTextArea_pasteSingleLine(t *testing.T) {
	clipRead = func() (string, error) { return "world", nil }
	clipWrite = func(s string) error { return nil }
	t.Cleanup(func() { clipRead = defaultClipRead; clipWrite = defaultClipWrite })

	value := "hello "
	setValue := func(s string) { value = s }
	h := NewHarness(t, func() Element {
		return C(NewTextArea(value, setValue, 3))
	})
	defer h.Close()

	h.SendKey(KeyEnd) // move cursor to end of "hello "
	h.SendKey(KeyCtrlV)

	if value != "hello world" {
		t.Errorf("expected %q after Ctrl+V, got %q", "hello world", value)
	}
}

// TestNewTextArea_pasteMultiLine verifies that multi-line clipboard content
// splits correctly at the cursor position.
func TestNewTextArea_pasteMultiLine(t *testing.T) {
	clipRead = func() (string, error) { return "B\nC", nil }
	clipWrite = func(s string) error { return nil }
	t.Cleanup(func() { clipRead = defaultClipRead; clipWrite = defaultClipWrite })

	value := "A\nD"
	setValue := func(s string) { value = s }
	h := NewHarness(t, func() Element {
		return C(NewTextArea(value, setValue, 4))
	})
	defer h.Close()

	// Cursor starts at line 0, col 1 (end of "A"). Paste "B\nC" there.
	h.SendKey(KeyEnd) // move to end of first line
	h.SendKey(KeyCtrlV)

	if value != "AB\nC\nD" {
		t.Errorf("expected %q after Ctrl+V, got %q", "AB\nC\nD", value)
	}
}
