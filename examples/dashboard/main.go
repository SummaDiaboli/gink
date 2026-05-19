// Dashboard — a server-monitoring TUI demonstrating UseContext for shared
// selection state, UseAsync for async data loading, UseInterval for live
// updates, NewList, NewSelect, ProgressBar, Badge, Border, and Dividers.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SummaDiaboli/gink"
)

// ── Data model ────────────────────────────────────────────────────────────────

// Status is the operational state of a server.
type Status string

const (
	Online  Status = "ONLINE"
	Warning Status = "WARN"
	Offline Status = "OFFLINE"
)

// Server is a static description of one server in the fleet.
type Server struct {
	Name   string
	Status Status
	Region string
}

// Metrics holds the runtime measurements fetched for a server.
type Metrics struct {
	CPU    float64
	Memory float64
	Disk   float64
	Uptime time.Duration
}

var servers = []Server{
	{Name: "web-01", Status: Online, Region: "us-east-1"},
	{Name: "web-02", Status: Online, Region: "us-east-1"},
	{Name: "web-03", Status: Warning, Region: "eu-west-1"},
	{Name: "db-primary", Status: Online, Region: "us-east-1"},
	{Name: "db-replica", Status: Online, Region: "us-west-2"},
	{Name: "cache-01", Status: Offline, Region: "us-east-1"},
	{Name: "cache-02", Status: Online, Region: "eu-west-1"},
}

// ── Shared context ────────────────────────────────────────────────────────────

// SelectedCtx holds the servers-slice index of the focused server.
// ServerList writes to it; MetricsPanel reads from it.
var SelectedCtx = gink.NewContext(0)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle   = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	sectionStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	valueStyle   = gink.NewStyle().Foreground(gink.ColorBrightWhite)
	labelStyle   = gink.NewStyle().Foreground(gink.ColorWhite)
	dimStyle     = gink.NewStyle().Foreground(gink.ColorWhite)
	hintStyle    = gink.NewStyle().Foreground(gink.ColorWhite)
	clockStyle   = gink.NewStyle().Foreground(gink.ColorBrightWhite)
	onlineStyle  = gink.NewStyle().Foreground(gink.ColorBrightGreen)
	warningStyle = gink.NewStyle().Foreground(gink.ColorBrightYellow)
	offlineStyle = gink.NewStyle().Foreground(gink.ColorBrightRed)
	logStyle     = gink.NewStyle().Foreground(gink.ColorWhite)
)

// ── Clock ─────────────────────────────────────────────────────────────────────

// Clock renders the current wall-clock time, updated every second.
func Clock() gink.Element {
	now, setNow := gink.UseState(time.Now())
	gink.UseInterval(time.Second, func() {
		setNow(time.Now())
	})
	return gink.Text(now.Format("Mon 02 Jan 2006  15:04:05"), clockStyle)
}

// ── ServerList ────────────────────────────────────────────────────────────────

// ServerList renders the left panel: a status-filter Select above a scrollable
// list of servers. Selecting a server updates SelectedCtx so MetricsPanel
// reacts without any prop threading.
//
// Note: Divider is intentionally avoided inside this panel because it uses
// UseTermSize (full terminal width), which would widen the border to fill the
// screen and leave no room for the right panel.
func ServerList() gink.Element {
	sel := gink.UseContext(SelectedCtx)
	filter, setFilter := gink.UseState("All")

	filterOpts := []string{"All", "Online", "Warning", "Offline"}

	// Map visible positions back to the real servers slice.
	var visible []int
	for i, s := range servers {
		keep := filter == "All" ||
			(filter == "Online" && s.Status == Online) ||
			(filter == "Warning" && s.Status == Warning) ||
			(filter == "Offline" && s.Status == Offline)
		if keep {
			visible = append(visible, i)
		}
	}

	// Find where sel sits in the filtered view; default to 0 when hidden.
	listSel := 0
	for i, idx := range visible {
		if idx == sel {
			listSel = i
			break
		}
	}

	items := make([]string, len(visible))
	for i, idx := range visible {
		s := servers[idx]
		dot := "● "
		switch s.Status {
		case Warning:
			dot = "◐ "
		case Offline:
			dot = "○ "
		}
		items[i] = dot + s.Name
	}

	var listElem gink.Element
	if len(visible) == 0 {
		listElem = gink.Text("  (no servers)", dimStyle)
	} else {
		listElem = gink.C(gink.NewList(items, listSel, func(i int) {
			gink.SetContext(SelectedCtx, visible[i])
		}, 8))
	}

	return gink.BorderWithTitle("Servers",
		gink.BoxWithGap(1,
			gink.Row(gink.Text("Filter ", labelStyle), gink.C(gink.NewSelect(filterOpts, filter, setFilter))),
			listElem,
		),
		titleStyle,
	)
}

// ── MetricsPanel ──────────────────────────────────────────────────────────────

// MetricsPanel renders the right panel. It uses UseAsync to fetch metrics
// whenever the selected server changes, showing a spinner during the load.
// UseInterval appends a new log entry every two seconds once metrics arrive.
//
// Section headers use styled text rather than DividerWithLabel to avoid
// expanding the panel to full terminal width.
func MetricsPanel() gink.Element {
	sel := gink.UseContext(SelectedCtx)
	server := servers[sel]

	// Fetch metrics asynchronously; re-fetches whenever sel changes.
	metrics, loading, _ := gink.UseAsync(func() (Metrics, error) {
		time.Sleep(400 * time.Millisecond) // simulate network latency
		return simulateMetrics(server.Name), nil
	}, []any{sel})

	// Live event log — reset on server change, then grows via UseInterval.
	logs, setLogs := gink.UseState([]string(nil))
	gink.UseEffect(func() func() {
		setLogs(nil)
		return nil
	}, []any{sel})

	gink.UseInterval(2*time.Second, func() {
		if loading {
			return
		}
		entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), randomLog(server.Name))
		next := append([]string{entry}, logs...)
		if len(next) > 5 {
			next = next[:5]
		}
		setLogs(next)
	})

	panelTitle := fmt.Sprintf("Details · %s", server.Name)

	if loading {
		return gink.BorderWithTitle(panelTitle,
			gink.PaddingXY(1, 0,
				gink.Row(gink.C(gink.Spinner), gink.Text("  Fetching metrics…", dimStyle)),
			),
			titleStyle,
		)
	}

	statusStyle := onlineStyle
	switch server.Status {
	case Warning:
		statusStyle = warningStyle
	case Offline:
		statusStyle = offlineStyle
	}

	var logSection gink.Element
	if len(logs) == 0 {
		logSection = gink.Text("  Waiting for events…", dimStyle)
	} else {
		rows := make([]gink.Element, len(logs))
		for i, l := range logs {
			rows[i] = gink.Text(l, logStyle)
		}
		logSection = gink.Box(rows...)
	}

	const barWidth = 20

	return gink.BorderWithTitle(panelTitle,
		gink.BoxWithGap(1,
			gink.Row(
				gink.Text("Status  ", labelStyle),
				gink.Badge(string(server.Status), statusStyle),
				gink.Text("    Region  ", labelStyle),
				gink.Text(server.Region, valueStyle),
			),
			gink.Row(
				gink.Text("Uptime  ", labelStyle),
				gink.Text(formatUptime(metrics.Uptime), valueStyle),
			),
			gink.Text("Metrics", sectionStyle),
			gink.Row(gink.Text("CPU     ", labelStyle), gink.ProgressBar(metrics.CPU, barWidth)),
			gink.Row(gink.Text("Memory  ", labelStyle), gink.ProgressBar(metrics.Memory, barWidth)),
			gink.Row(gink.Text("Disk    ", labelStyle), gink.ProgressBar(metrics.Disk, barWidth)),
			gink.Text("Events", sectionStyle),
			logSection,
		),
		titleStyle,
	)
}

// ── Fleet Table ───────────────────────────────────────────────────────────────

// fleetCols defines the Table columns with MinWidth and MaxWidth constraints.
// MinWidth=12 on Server ensures short names ("web-01") are padded to a uniform
// width. MaxWidth=8 on Region truncates "us-east-1" → "us-east…".
var fleetCols = []gink.Column{
	{Header: "Server", MinWidth: 12},
	{Header: "Status"},
	{Header: "Region", MaxWidth: 8},
	{Header: "CPU"},
	{Header: "Memory"},
	{Header: "Disk"},
}

// fleetTable builds a summary Table of all servers with their current metrics.
// It is a plain function (not a component) because it uses no hooks — no C()
// wrapper is needed when calling it from App.
func fleetTable() gink.Element {
	rows := make([][]string, len(servers))
	for i, s := range servers {
		m := simulateMetrics(s.Name)
		rows[i] = []string{
			s.Name,
			string(s.Status),
			s.Region,
			fmt.Sprintf("%d%%", int(m.CPU*100)),
			fmt.Sprintf("%d%%", int(m.Memory*100)),
			fmt.Sprintf("%d%%", int(m.Disk*100)),
		}
	}
	return gink.BoxWithGap(1,
		gink.Text("Fleet Overview", sectionStyle),
		gink.Table(fleetCols, rows, titleStyle),
	)
}

// ── App ───────────────────────────────────────────────────────────────────────

// App is the root component. It composes the header, two-column layout, fleet
// summary table, and footer hint bar. Tab cycles focus between the filter
// Select and the List.
func App() gink.Element {
	return gink.PaddingXY(2, 1,
		gink.BoxWithGap(1,
			gink.Row(
				gink.Text("◆ Gink Dashboard", titleStyle),
				gink.Text("   "),
				gink.C(Clock),
			),
			gink.C(gink.Divider),
			gink.Row(
				gink.C(ServerList),
				gink.Text("  "),
				gink.C(MetricsPanel),
			),
			gink.C(gink.Divider),
			fleetTable(),
			gink.C(gink.Divider),
			gink.Text("Tab: switch focus  ·  ↑↓: navigate  ·  Esc: quit", hintStyle),
		),
	)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// simulateMetrics produces deterministic-ish metric values for a server name
// so the dashboard shows varied but stable readings without a real backend.
func simulateMetrics(name string) Metrics {
	h := 0
	for _, c := range name {
		h = h*31 + int(c)
	}
	if h < 0 {
		h = -h
	}
	return Metrics{
		CPU:    float64(h%55+10) / 100.0,
		Memory: float64((h/7)%45+20) / 100.0,
		Disk:   float64((h/13)%35+30) / 100.0,
		Uptime: time.Duration(int64(time.Hour) * int64(h%480+12)),
	}
}

// formatUptime renders a duration as "Xd Yh Zm" or "Yh Zm".
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

var logTemplates = []func(int) string{
	func(v int) string { return fmt.Sprintf("GET /api/health 200 %dms", v%80+5) },
	func(v int) string { return fmt.Sprintf("POST /api/data 201 %dms", v%150+20) },
	func(v int) string { return fmt.Sprintf("DB query completed in %dms", v%40+2) },
	func(v int) string { return fmt.Sprintf("Cache hit ratio %.0f%%", float64(v%25+72)) },
	func(v int) string { return fmt.Sprintf("Active goroutines: %d", v%30+8) },
	func(v int) string { return fmt.Sprintf("GC pause %dms", v%4+1) },
	func(v int) string { return fmt.Sprintf("TLS handshake completed %dms", v%20+5) },
	func(v int) string { return fmt.Sprintf("Replica lag: %dms", v%10+1) },
}

// randomLog generates a plausible log line, varying by server name and time.
func randomLog(name string) string {
	seed := int(time.Now().Unix()/2) + len(name)*7
	fn := logTemplates[seed%len(logTemplates)]
	return fn(seed)
}
