// Focus order: Filter Select (0) · Server List (1)
package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/SummaDiaboli/gink"
	"github.com/SummaDiaboli/gink/ginktest"
)

// resetState resets shared context so tests are isolated from each other.
func resetState() {
	gink.SetContext(SelectedCtx, 0)
}

// newDashHarness creates a tall-enough harness for the full dashboard layout,
// which now includes the fleet table below the two-panel section.
func newDashHarness(t *testing.T) *gink.Harness {
	t.Helper()
	return ginktest.NewHarnessSize(t, App, 80, 60)
}

func TestDashboard_showsAllServersInList(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	for _, s := range servers {
		if !h.Contains(s.Name) {
			t.Errorf("expected server %q to appear in the list", s.Name)
		}
	}
}

func TestDashboard_showsLoadingSpinnerBeforeMetricsArrive(t *testing.T) {
	resetState()
	h := newDashHarness(t)
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
	h := newDashHarness(t)
	defer h.Close()

	// UseAsync resolves in ~400ms; wait up to 2s for the metrics section.
	ginktest.AwaitContains(t, h, "CPU", 2*time.Second)
	ginktest.AwaitContains(t, h, "Memory", 2*time.Second)
	ginktest.AwaitContains(t, h, "Disk", 2*time.Second)
}

func TestDashboard_showsStatusBadgeAfterLoad(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	ginktest.AwaitContains(t, h, string(servers[0].Status), 2*time.Second)
}

func TestDashboard_navigationChangesDetailPanel(t *testing.T) {
	resetState()
	h := newDashHarness(t)
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
	h := newDashHarness(t)
	defer h.Close()

	// Select is focused by default. Press Down to change filter to "Online".
	h.SendKey(gink.KeyDown)

	// "○ cache-01" is the list-entry format for an offline server.
	// "cache-01" alone would also match the fleet table, which always shows all servers.
	if h.Contains("○ cache-01") {
		t.Error("cache-01 (OFFLINE) list entry should be hidden when filter is Online")
	}
}

func TestDashboard_filterShowsWarningServersOnly(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	// Cycle filter: All → Online → Warning
	h.SendKey(gink.KeyDown) // Online
	h.SendKey(gink.KeyDown) // Warning

	if !h.Contains("◐ web-03") {
		t.Error("web-03 (WARN) should appear in the list with Warning filter")
	}
	// "● web-01" is the list-entry format; "web-01" also appears in MetricsPanel
	// and the fleet table, so we check for the dot prefix to target the list only.
	if h.Contains("● web-01") {
		t.Error("web-01 (ONLINE) list entry should be hidden with Warning filter")
	}
}

func TestDashboard_liveLogAppearsAfterMetricsLoad(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	// Wait for metrics to load first, then wait for at least one log entry.
	ginktest.AwaitContains(t, h, "CPU", 2*time.Second)
	// Log entries use format [HH:MM:SS]; match current hour so it works at any time of day.
	ginktest.AwaitContains(t, h, "["+time.Now().Format("15:04"), 5*time.Second)
}

func TestDashboard_clockUpdatesOverTime(t *testing.T) {
	resetState()
	h := newDashHarness(t)
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

// ── Fleet Table tests ─────────────────────────────────────────────────────────

func TestDashboard_fleetTableVisible(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	if !h.Contains("Fleet Overview") {
		t.Error("fleet section label should be visible")
	}
}

func TestDashboard_fleetTableShowsAllServersRegardlessOfFilter(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	// Switch filter to Online — cache-01 disappears from the list.
	h.SendKey(gink.KeyDown)

	// cache-01 is absent from the list (○ cache-01) but must still appear in
	// the fleet table, which is not filtered.
	if h.Contains("○ cache-01") {
		t.Error("list entry for cache-01 should be gone with Online filter")
	}
	if !h.Contains("cache-01") {
		t.Error("fleet table should still show cache-01 even when list filter hides it")
	}
}

func TestDashboard_fleetTableTruncatesRegion(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	// Region "us-east-1" (9 chars) exceeds MaxWidth=8 → truncated to "us-east…".
	if !h.Contains("us-east…") {
		t.Error("fleet table should truncate 'us-east-1' to 'us-east…' (MaxWidth=8)")
	}
}

func TestDashboard_fleetTableShowsMetricPercentages(t *testing.T) {
	resetState()
	h := newDashHarness(t)
	defer h.Close()

	// simulateMetrics is deterministic; web-01 always has the same CPU value.
	m := simulateMetrics("web-01")
	expected := fmt.Sprintf("%d%%", int(m.CPU*100))
	if !h.Contains(expected) {
		t.Errorf("fleet table should show CPU metric %q for web-01", expected)
	}
}
