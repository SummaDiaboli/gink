package gink

import (
	"testing"
	"time"
)

func TestUseToast_initiallyEmpty(t *testing.T) {
	h := NewHarness(t, func() Element {
		el, _ := UseToast()
		return el
	})
	defer h.Close()

	if h.Contains("toast") {
		t.Error("toast should not be visible before show is called")
	}
}

func TestUseToast_showRendersMessage(t *testing.T) {
	var show func(string, time.Duration)
	h := NewHarness(t, func() Element {
		el, showFn := UseToast()
		show = showFn
		return el
	})
	defer h.Close()

	show("saved!", time.Second)
	h.Render()

	if !h.Contains("saved!") {
		t.Errorf("toast message not rendered; lines: %v", h.Lines())
	}
}

func TestUseToast_autoDismissesAfterDuration(t *testing.T) {
	var show func(string, time.Duration)
	h := NewHarness(t, func() Element {
		el, showFn := UseToast()
		show = showFn
		return el
	})
	defer h.Close()

	show("bye!", 50*time.Millisecond)
	h.Render()

	if !h.Contains("bye!") {
		t.Fatal("toast should be visible immediately after show")
	}

	// Wait for auto-dismiss then re-render.
	time.Sleep(150 * time.Millisecond)
	h.Render()

	if h.Contains("bye!") {
		t.Error("toast should have been dismissed after duration elapsed")
	}
}

func TestUseToast_showAgainResetsTimer(t *testing.T) {
	var show func(string, time.Duration)
	h := NewHarness(t, func() Element {
		el, showFn := UseToast()
		show = showFn
		return el
	})
	defer h.Close()

	show("first", 60*time.Millisecond)
	h.Render()

	// Call show again before the first toast expires.
	time.Sleep(20 * time.Millisecond)
	show("second", 60*time.Millisecond)
	h.Render()

	if !h.Contains("second") {
		t.Fatal("second toast should be visible")
	}

	// Wait for "second" timer (60ms from now) to fire, then re-render.
	time.Sleep(150 * time.Millisecond)
	h.Render()

	if h.Contains("second") {
		t.Error("second toast should have been dismissed")
	}
}
