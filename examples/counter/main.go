package main

import (
	"fmt"
	"log"

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

	return gink.PaddingXY(2, 1,
		gink.BoxWithGap(1,
			gink.C(gink.DividerWithLabel("Gink Component Demo", titleStyle)),

			gink.Row(
				gink.Text("Status: ", labelStyle),
				func() gink.Element {
					if loading {
						return gink.C(gink.Spinner)
					}
					return gink.Text(fmt.Sprintf("Count is %d", count), countStyle)
				}(),
			),

			gink.C(gink.Divider),

			gink.Row(
				gink.Text("Name:  ", labelStyle),
				gink.C(gink.NewInput(name, setName)),
			),

			gink.RowWithGap(2,
				gink.C(gink.NewButton("Increment", func() { setCount(count + 1) })),
				gink.C(gink.NewButton("Decrement", func() { setCount(count - 1) })),
				gink.C(gink.NewButton("Toggle Loading", func() { setLoading(!loading) })),
			),

			gink.C(gink.Divider),

			gink.Text("Tab to move focus  •  Enter/Space to press buttons  •  Esc to quit", hintStyle),
		),
	)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}
