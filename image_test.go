package gink

import (
	"image"
	"image/color"
	"testing"
)

// solidImage returns a uniform-colour NRGBA image of the given dimensions.
func solidImage(c color.NRGBA, w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

// TestImage_explicitDimensions verifies that explicit width and height produce
// a block of exactly those cell dimensions. Checks via CellStyle since solid-
// colour cells may render as any quadrant character (including spaces with
// coloured backgrounds).
func TestImage_explicitDimensions(t *testing.T) {
	img := solidImage(color.NRGBA{R: 255, A: 255}, 10, 10)
	h := NewHarness(t, func() Element {
		return Image(img, 5, 3)
	})
	defer h.Close()

	defaultStyle := (Style{}).toTcell()

	// All 5 cells in row 0 must have a non-default style (image content).
	for x := 0; x < 5; x++ {
		if h.CellStyle(x, 0) == defaultStyle {
			t.Errorf("cell (%d, 0): expected non-default style from image", x)
		}
	}
	// Row 3 (one past the 3-cell-tall image) must be default (no image content).
	if h.CellStyle(0, 3) != defaultStyle {
		t.Errorf("cell (0, 3): expected default style outside image bounds; got %v", h.CellStyle(0, 3))
	}
}

// TestImage_autoHeightSquare verifies that a square image with height=0 produces
// the correct auto cellH. With quadrant rendering each cell covers 2×2 pixels,
// so cellH = width * srcH / srcW = 8 for a square image at width=8.
func TestImage_autoHeightSquare(t *testing.T) {
	// 10×10 at width=8: pixW=16, pixH=16, cellH=8.
	img := solidImage(color.NRGBA{R: 255, A: 255}, 10, 10)
	h := NewHarness(t, func() Element {
		return Image(img, 8, 0)
	})
	defer h.Close()

	if h.Line(7) == "" {
		t.Error("line 7 should have content (auto cellH=8 for 10×10 image at width=8)")
	}
	if h.Line(8) != "" {
		t.Errorf("line 8 should be empty (auto cellH=8); got %q", h.Line(8))
	}
}

// TestImage_autoHeightWide verifies aspect-ratio-preserving auto-height for a
// wide (2:1) image. cellH = width * srcH / srcW = 8*10/20 = 4.
func TestImage_autoHeightWide(t *testing.T) {
	// 20×10 at width=8: pixW=16, pixH=8, cellH=4.
	img := solidImage(color.NRGBA{G: 255, A: 255}, 20, 10)
	h := NewHarness(t, func() Element {
		return Image(img, 8, 0)
	})
	defer h.Close()

	if h.Line(3) == "" {
		t.Error("line 3 should have content (auto cellH=4 for 2:1 image at width=8)")
	}
	if h.Line(4) != "" {
		t.Errorf("line 4 should be empty (auto cellH=4); got %q", h.Line(4))
	}
}

// TestImage_solidColorUsesFullBlock verifies that a solid-colour image renders
// as '█' (full block) rather than spaces. The tiebreaker in quadrantCell
// prefers more foreground pixels when all 16 splits have equal error.
func TestImage_solidColorUsesFullBlock(t *testing.T) {
	img := solidImage(color.NRGBA{R: 200, A: 255}, 6, 4)
	h := NewHarness(t, func() Element {
		return Image(img, 6, 2)
	})
	defer h.Close()

	for i, r := range []rune(h.Line(0)) {
		if r != '█' {
			t.Errorf("cell %d: got %q, want █ for solid-colour image", i, r)
		}
	}
}

// TestImage_colorsTopIsForeground verifies that the top-pixel colour maps to
// the cell foreground and the bottom-pixel colour maps to the background, using
// the half-block character ▀ (mask=0b0011 = UL+UR fg, LL+LR bg).
func TestImage_colorsTopIsForeground(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for x := 0; x < 2; x++ {
		img.SetNRGBA(x, 0, color.NRGBA{R: 255, A: 255}) // top row: red
		img.SetNRGBA(x, 1, color.NRGBA{G: 255, A: 255}) // bottom row: green
	}

	h := NewHarness(t, func() Element {
		return Image(img, 2, 1)
	})
	defer h.Close()

	// The optimal split for (red, red, green, green) is mask=0b0011: upper half
	// fg=red, lower half bg=green → character '▀'.
	want := NewStyle().
		Foreground(NewRGBColor(255, 0, 0)).
		Background(NewRGBColor(0, 255, 0)).
		toTcell()
	got := h.CellStyle(0, 0)
	if got != want {
		t.Errorf("cell (0,0) style mismatch:\n got  %v\n want %v", got, want)
	}
}
