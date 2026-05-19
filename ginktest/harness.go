// Package ginktest provides testing utilities for Gink components.
// It follows the same pattern as [net/http/httptest] — import it only in
// _test.go files.
//
// # Quick start
//
//	import "github.com/SummaDiaboli/gink/ginktest"
//
//	func TestMyComponent(t *testing.T) {
//	    h := ginktest.NewHarness(t, MyComponent)
//	    defer h.Close()
//
//	    h.Tab()          // move focus to next component
//	    h.Enter()        // press Enter on the focused component
//	    h.SendRune('x')  // type a character
//
//	    ginktest.AssertContains(t, h, "expected text")
//	    ginktest.AssertNotContains(t, h, "unexpected text")
//	}
//
// # How it works
//
// [NewHarness] creates an 80×24 virtual terminal backed by a tcell
// SimulationScreen. No real terminal is needed — tests run headlessly in CI.
// Each call to an input method ([Harness.Tab], [Harness.Enter], [Harness.SendRune],
// etc.) dispatches the event and immediately re-renders the component tree, so
// assertions can be made synchronously right after each interaction.
//
// # Async components
//
// Components that use [gink.UseInterval] or [gink.UseEffect] with goroutines
// update state asynchronously. Use [AwaitContains] instead of [AssertContains]
// to poll until the expected text appears or a timeout elapses:
//
//	ginktest.AwaitContains(t, h, "Running", 500*time.Millisecond)
//
// # Focus order
//
// Tab cycles focus through [gink.UseFocus] components in tree order, starting
// from the first focusable component. Document the focus order in a comment at
// the top of each test file so readers can follow the Tab/ShiftTab sequences:
//
//	// Focus order: Name(0) · Email(1) · Subject(2) · Submit(3)
package ginktest

import (
	"strings"
	"testing"
	"time"

	"github.com/SummaDiaboli/gink"
)

// NewHarness creates a test harness with an 80×24 simulation screen and
// renders root once. Use [gink.Harness.Render] to re-render after state
// changes or input.
func NewHarness(t *testing.T, root gink.Component) *gink.Harness {
	t.Helper()
	return gink.NewHarness(t, root)
}

// NewHarnessSize creates a test harness with custom terminal dimensions.
// Use this when the component's layout depends on the terminal size
// (e.g. components that use [gink.UseTermSize]).
func NewHarnessSize(t *testing.T, root gink.Component, width, height int) *gink.Harness {
	t.Helper()
	return gink.NewHarnessSize(t, root, width, height)
}

// AssertContains fails the test if the rendered screen does not contain s.
// The full screen contents are printed on failure.
func AssertContains(t *testing.T, h *gink.Harness, s string) {
	t.Helper()
	if !h.Contains(s) {
		t.Errorf("expected screen to contain %q\nscreen:\n%s", s, dump(h))
	}
}

// AssertNotContains fails the test if the rendered screen contains s.
// The full screen contents are printed on failure.
func AssertNotContains(t *testing.T, h *gink.Harness, s string) {
	t.Helper()
	if h.Contains(s) {
		t.Errorf("expected screen NOT to contain %q\nscreen:\n%s", s, dump(h))
	}
}

// AwaitContains re-renders every 50 ms until the screen contains s or timeout
// elapses. Use this for components that update asynchronously via
// [gink.UseInterval] or [gink.UseEffect] goroutines.
//
//	ginktest.AwaitContains(t, h, "Running", 500*time.Millisecond)
func AwaitContains(t *testing.T, h *gink.Harness, s string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		h.Render()
		if h.Contains(s) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Errorf("timed out after %s waiting for screen to contain %q\nscreen:\n%s", timeout, s, dump(h))
}

// AwaitNotContains re-renders every 50 ms until the screen no longer contains
// s or timeout elapses. Use this to wait for async content to disappear, such
// as a loading indicator being replaced by real data.
//
//	ginktest.AwaitNotContains(t, h, "Loading…", 2*time.Second)
func AwaitNotContains(t *testing.T, h *gink.Harness, s string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		h.Render()
		if !h.Contains(s) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Errorf("timed out after %s waiting for screen to stop containing %q\nscreen:\n%s", timeout, s, dump(h))
}

func dump(h *gink.Harness) string {
	lines := h.Lines()
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
