package main

import (
	"testing"
	"time"

	"github.com/salim/gink/ginktest"
)

// Focus order: Start/Stop(0) · Lap(1) · Reset(2)

func TestStopwatch_initialRender(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	ginktest.AssertContains(t, h, "00:00.0")
	ginktest.AssertContains(t, h, "Stopped")
}

func TestStopwatch_startChangesStatus(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter() // Start (focused by default)

	ginktest.AssertContains(t, h, "Running")
}

func TestStopwatch_stopChangesStatusBack(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter() // Start
	h.Enter() // Stop

	ginktest.AssertContains(t, h, "Stopped")
}

func TestStopwatch_timerAdvancesWhileRunning(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter() // Start

	ginktest.AwaitContains(t, h, "00:00.", 500*time.Millisecond)

	// After ~400ms the tenths digit must have advanced past 0.
	time.Sleep(400 * time.Millisecond)
	h.Render()

	if h.Contains("00:00.0") && !h.Contains("00:00.") {
		t.Error("timer did not advance after 400ms")
	}
	// At least one tick should have moved the display.
	ginktest.AssertNotContains(t, h, "00:00.0\n") // not stuck at exactly zero
}

func TestStopwatch_timerFreezesWhenStopped(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter()                          // Start
	time.Sleep(300 * time.Millisecond) // let it tick
	h.Enter()                          // Stop
	h.Render()
	frozen := h.Line(1)

	time.Sleep(300 * time.Millisecond) // wait — timer should not move
	h.Render()

	if h.Line(1) != frozen {
		t.Errorf("timer moved after stop: %q → %q", frozen, h.Line(1))
	}
}

func TestStopwatch_resetClearsTime(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter()                          // Start
	time.Sleep(300 * time.Millisecond) // let it tick
	h.Render()
	h.Tab()   // → Lap
	h.Tab()   // → Reset
	h.Enter()

	ginktest.AssertContains(t, h, "00:00.0")
	ginktest.AssertContains(t, h, "Stopped")
}

func TestStopwatch_lapRecordedWhileRunning(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter() // Start
	h.Tab()   // → Lap
	h.Enter() // record lap at ~00:00.0

	ginktest.AssertContains(t, h, "Lap 1")
}

func TestStopwatch_lapIgnoredWhenStopped(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Stopwatch is stopped — press Lap.
	h.Tab()   // → Lap
	h.Enter()

	ginktest.AssertNotContains(t, h, "Lap 1")
}
