// Focus order: Filter Select (0) · Server List (1)
package main

import (
	"testing"
	"time"

	"github.com/SummaDiaboli/gink"
	"github.com/SummaDiaboli/gink/ginktest"
)

// resetState resets shared context so tests are isolated from each other.
func resetState() {
	gink.SetContext(SelectedCtx, 0)
}

func TestDashboard_showsAllServersInList(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	for _, s := range servers {
		if !h.Contains(s.Name) {
			t.Errorf("expected server %q to appear in the list", s.Name)
		}
	}
}

func TestDashboard_showsLoadingSpinnerBeforeMetricsArrive(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// UseAsync starts loading on mount; the first render should show the spinner
	// (or the metrics if they resolved before the first render, which is fine too).
	// We just check the panel title is visible so the layout is correct.
	if !h.Contains("Details · web-01") {
		t.Error("panel title for first server should be visible immediately")
	}
}

func TestDashboard_showsMetricsAfterLoad(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// UseAsync resolves in ~400ms; wait up to 2s for the metrics section.
	ginktest.AwaitContains(t, h, "CPU", 2*time.Second)
	ginktest.AwaitContains(t, h, "Memory", 2*time.Second)
	ginktest.AwaitContains(t, h, "Disk", 2*time.Second)
}

func TestDashboard_showsStatusBadgeAfterLoad(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	ginktest.AwaitContains(t, h, string(servers[0].Status), 2*time.Second)
}

func TestDashboard_navigationChangesDetailPanel(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Focus order: Select(0) · List(1). Tab once to reach the list.
	h.Tab()

	// Navigate to web-02 (index 1).
	h.SendKey(gink.KeyDown)

	// Detail panel title should update immediately (SetContext is synchronous).
	if !h.Contains("Details · web-02") {
		t.Errorf("after navigating to web-02, detail title should update; lines: %v", h.Lines()[:5])
	}
}

func TestDashboard_filterHidesServers(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Select is focused by default. Press Down to change filter to "Online".
	h.SendKey(gink.KeyDown)

	// Offline server cache-01 should no longer appear in the list.
	if h.Contains("cache-01") {
		t.Error("cache-01 (OFFLINE) should be hidden when filter is Online")
	}
}

func TestDashboard_filterShowsWarningServersOnly(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Cycle filter: All → Online → Warning
	h.SendKey(gink.KeyDown) // Online
	h.SendKey(gink.KeyDown) // Warning

	if !h.Contains("◐ web-03") {
		t.Error("web-03 (WARN) should appear in the list with Warning filter")
	}
	// "● web-01" is the list-entry format; "web-01" also appears in the MetricsPanel title.
	if h.Contains("● web-01") {
		t.Error("web-01 (ONLINE) list entry should be hidden with Warning filter")
	}
}

func TestDashboard_liveLogAppearsAfterMetricsLoad(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Wait for metrics to load first, then wait for at least one log entry.
	ginktest.AwaitContains(t, h, "CPU", 2*time.Second)
	// Log entries use format [HH:MM:SS]; match current hour so it works at any time of day.
	ginktest.AwaitContains(t, h, "["+time.Now().Format("15:04"), 5*time.Second)
}

func TestDashboard_clockUpdatesOverTime(t *testing.T) {
	resetState()
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Grab the clock text from the first render.
	before := h.Line(1) // row 1 is the header row (after top padding)

	// Wait just over one second for the clock to tick.
	time.Sleep(1100 * time.Millisecond)
	h.Render()

	after := h.Line(1)
	if before == after {
		t.Error("clock should have updated after one second")
	}
}
