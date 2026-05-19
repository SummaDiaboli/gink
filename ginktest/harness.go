// Package ginktest provides testing utilities for Gink components.
// It follows the same pattern as net/http/httptest — import it only in
// _test.go files.
//
//	import "github.com/salim/gink/ginktest"
//
//	func TestMyComponent(t *testing.T) {
//	    h := ginktest.NewHarness(t, MyComponent)
//	    defer h.Close()
//
//	    h.Tab()
//	    h.Enter()
//	    ginktest.AssertContains(t, h, "expected text")
//	}
package ginktest

import (
	"strings"
	"testing"
	"time"

	"github.com/salim/gink"
)

// NewHarness creates a test harness with an 80×24 simulation screen and
// renders root once. Use Render() to re-render after state changes or input.
func NewHarness(t *testing.T, root gink.Component) *gink.Harness {
	t.Helper()
	return gink.NewHarness(t, root)
}

// NewHarnessSize creates a test harness with custom terminal dimensions.
func NewHarnessSize(t *testing.T, root gink.Component, width, height int) *gink.Harness {
	t.Helper()
	return gink.NewHarnessSize(t, root, width, height)
}

// AssertContains fails the test if the screen does not contain s.
func AssertContains(t *testing.T, h *gink.Harness, s string) {
	t.Helper()
	if !h.Contains(s) {
		t.Errorf("expected screen to contain %q\nscreen:\n%s", s, dump(h))
	}
}

// AssertNotContains fails the test if the screen contains s.
func AssertNotContains(t *testing.T, h *gink.Harness, s string) {
	t.Helper()
	if h.Contains(s) {
		t.Errorf("expected screen NOT to contain %q\nscreen:\n%s", s, dump(h))
	}
}

// AwaitContains re-renders every 50 ms until the screen contains s or timeout
// elapses. Useful for asserting on UseInterval or async UseEffect output.
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

func dump(h *gink.Harness) string {
	lines := h.Lines()
	// Trim trailing blank lines for readability.
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}
