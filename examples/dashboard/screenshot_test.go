package main

import (
	"os"
	"testing"
	"time"

	"github.com/SummaDiaboli/gink/ginktest"
	"github.com/SummaDiaboli/gink/internal/readmeimg"
)

// TestScreenshotDashboard regenerates the README images for the dashboard.
// It is skipped unless GINK_SCREENSHOTS is set:
//
//	GINK_SCREENSHOTS=1 go test ./examples/dashboard -run TestScreenshotDashboard -v
func TestScreenshotDashboard(t *testing.T) {
	if os.Getenv("GINK_SCREENSHOTS") == "" {
		t.Skip("set GINK_SCREENSHOTS=1 to regenerate docs/assets/dashboard.png and .gif")
	}
	t.Logf("rasterizing with %s", readmeimg.FontSource())

	resetState()
	h := ginktest.NewHarnessSize(t, App, 96, 48)
	defer h.Close()

	// Wait for the async metrics, then for the first live log entry so the
	// still image shows a fully populated dashboard. The loading branch hides
	// the events section entirely, so waiting for "Waiting for events…" to
	// appear is the reliable signal that metrics have arrived.
	ginktest.AwaitContains(t, h, "CPU", 2*time.Second)
	ginktest.AwaitContains(t, h, "Waiting for events…", 2*time.Second)
	ginktest.AwaitNotContains(t, h, "Waiting for events…", 5*time.Second)
	h.Tab() // focus the server list instead of the filter select

	if err := readmeimg.WritePNG("../../docs/assets/dashboard.png", readmeimg.Capture(h), "gink — dashboard"); err != nil {
		t.Fatalf("write dashboard PNG: %v", err)
	}

	// Animate a fresh instance: loading spinner → metrics → live events.
	resetState()
	g := ginktest.NewHarnessSize(t, App, 96, 48)
	defer g.Close()
	g.Tab()

	const (
		frames = 24
		delay  = 140 * time.Millisecond
	)
	shots := make([]*readmeimg.Snapshot, 0, frames)
	for i := 0; i < frames; i++ {
		time.Sleep(delay)
		g.Render()
		shots = append(shots, readmeimg.Capture(g))
	}

	if err := readmeimg.WriteGIF("../../docs/assets/dashboard.gif", shots, "gink — dashboard", int(delay/time.Millisecond)); err != nil {
		t.Fatalf("write dashboard GIF: %v", err)
	}
}
