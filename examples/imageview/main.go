// Package main demonstrates gink's Image renderer by fetching a random photo
// from picsum.photos and displaying it using Unicode half-block characters.
// Press R to fetch a new random image; Escape to quit.
package main

import (
	"image"
	_ "image/jpeg"
	"log"
	"net/http"

	"github.com/SummaDiaboli/gink"
)

var titleStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
var mutedStyle = gink.NewStyle().Foreground(gink.ColorWhite)
var errStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightRed)

// fetchImage downloads a random image from picsum.photos at the given pixel
// dimensions. picsum always returns a JPEG.
func fetchImage(pixW, pixH int) (image.Image, error) {
	url := "https://picsum.photos/" + itoa(pixW) + "/" + itoa(pixH)
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	return img, err
}

func App() gink.Element {
	size := gink.UseTermSize()
	seed, setSeed := gink.UseState(0)

	// R refreshes the image.
	gink.UseKeyboard(func(ev gink.KeyEvent) {
		if ev.Rune == 'r' || ev.Rune == 'R' {
			setSeed(seed + 1)
		}
	})

	// Render at most 40 cells wide and 20 cells tall so the image fits
	// comfortably in a standard 80×24 terminal alongside the header.
	imgCellW := size.Width / 2
	if imgCellW > 40 {
		imgCellW = 40
	}
	if imgCellW < 10 {
		imgCellW = 10
	}
	imgCellH := 18 // explicit height keeps the image within the visible screen

	// Fetch at 10× the output pixel dimensions so Catmull-Rom has plenty of
	// source data to anti-alias from. Each cell covers 2×2 pixels with
	// quadrant rendering, so the effective pixel grid is imgCellW*2 × imgCellH*2.
	fetchPx := imgCellW * 20
	pixW := fetchPx
	pixH := fetchPx // square fetch; Catmull-Rom handles the aspect rescale

	img, loading, fetchErr := gink.UseAsync(func() (image.Image, error) {
		return fetchImage(pixW, pixH)
	}, []any{seed})

	header := gink.Row(
		gink.Text("picsum.photos viewer", titleStyle),
		gink.Text("  R = new image  •  Esc = quit", mutedStyle),
	)

	if loading {
		return gink.BoxWithGap(1,
			header,
			gink.C(gink.Spinner),
		)
	}
	if fetchErr != nil {
		return gink.BoxWithGap(1,
			header,
			gink.Text("fetch error: "+fetchErr.Error(), errStyle),
		)
	}

	return gink.BoxWithGap(1,
		header,
		gink.Image(img, imgCellW, imgCellH),
	)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}

// itoa converts a non-negative integer to its decimal string representation
// without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
