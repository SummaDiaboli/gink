package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/salim/gink"
)

var titleStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
var labelStyle = gink.NewStyle().Foreground(gink.ColorWhite)
var countStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightYellow)
var hintStyle = gink.NewStyle().Foreground(gink.ColorWhite)

func App() gink.Element {
	count, setCount := gink.UseState(0)
	loading, setLoading := gink.UseState(false)
	name, setName := gink.UseState("")

	size := gink.UseTermSize()
	divider := strings.Repeat("─", size.Width)

	return gink.BoxWithGap(1,
		gink.Text("Gink Component Demo", titleStyle),
		gink.Text(divider),

		// Spinner visible while loading
		gink.Row(
			gink.Text("Status: "),
			func() gink.Element {
				if loading {
					return gink.C(gink.Spinner)
				}
				return gink.Text(fmt.Sprintf("Count is %d", count), countStyle)
			}(),
		),

		gink.Text(divider),

		gink.Row(
			gink.Text("Name:  "),
			gink.C(gink.NewInput(name, setName)),
		),

		gink.RowWithGap(2,
			gink.C(gink.NewButton("Increment", func() { setCount(count + 1) })),
			gink.C(gink.NewButton("Decrement", func() { setCount(count - 1) })),
			gink.C(gink.NewButton("Toggle Loading", func() { setLoading(!loading) })),
		),

		gink.Text("Tab to move focus  •  Enter/Space to press buttons  •  Esc to quit", hintStyle),
	)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}
