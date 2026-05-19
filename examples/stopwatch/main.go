// Stopwatch — demonstrates UseInterval for real-time UI updates and
// UseState toggling between running/stopped modes.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SummaDiaboli/gink"
)

var (
	titleStyle   = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	timerStyle   = gink.NewStyle().Bold().Foreground(gink.ColorBrightYellow)
	runningStyle = gink.NewStyle().Foreground(gink.ColorBrightGreen)
	stoppedStyle = gink.NewStyle().Foreground(gink.ColorWhite)
	lapStyle     = gink.NewStyle().Foreground(gink.ColorBrightWhite)
	hintStyle    = gink.NewStyle().Foreground(gink.ColorWhite)
)

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	tenth := int(d.Milliseconds()/100) % 10
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%d", h, m, s, tenth)
	}
	return fmt.Sprintf("%02d:%02d.%d", m, s, tenth)
}

func App() gink.Element {
	elapsed, setElapsed := gink.UseState(time.Duration(0))
	running, setRunning := gink.UseState(false)
	laps, setLaps := gink.UseState([]time.Duration(nil))

	// Tick every 100 ms while running. UseInterval keeps the callback fresh
	// so it always sees the latest elapsed and running values.
	gink.UseInterval(100*time.Millisecond, func() {
		if running {
			setElapsed(elapsed + 100*time.Millisecond)
		}
	})

	startStop := func() { setRunning(!running) }

	reset := func() {
		setRunning(false)
		setElapsed(0)
		setLaps(nil)
	}

	lap := func() {
		if !running {
			return
		}
		next := append([]time.Duration{}, laps...)
		next = append(next, elapsed)
		setLaps(next)
	}

	btnLabel := "  Start  "
	statusStr := "Stopped"
	statusStyle := stoppedStyle
	if running {
		btnLabel = "   Stop  "
		statusStr = "Running"
		statusStyle = runningStyle
	}

	lapRows := make([]gink.Element, len(laps))
	for i, l := range laps {
		lapRows[i] = gink.Text(fmt.Sprintf("  Lap %d   %s", i+1, formatDuration(l)), lapStyle)
	}

	children := []gink.Element{
		gink.C(gink.DividerWithLabel("Stopwatch", titleStyle)),
		gink.Text(formatDuration(elapsed), timerStyle),
		gink.Text(statusStr, statusStyle),
		gink.C(gink.Divider),
		gink.RowWithGap(2,
			gink.C(gink.NewButton(btnLabel, startStop)),
			gink.C(gink.NewButton("  Lap    ", lap)),
			gink.C(gink.NewButton("  Reset  ", reset)),
		),
	}

	if len(lapRows) > 0 {
		children = append(children,
			gink.C(gink.DividerWithLabel("Laps")),
			gink.Box(lapRows...),
		)
	}

	children = append(children,
		gink.C(gink.Divider),
		gink.Text("Tab/ShiftTab move focus  ·  Enter/Space activate  ·  Esc quit", hintStyle),
	)

	return gink.PaddingXY(2, 1, gink.BoxWithGap(1, children...))
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}
