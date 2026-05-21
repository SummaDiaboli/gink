package gink

import (
	"image"
	"image/color"
	"math"

	"github.com/gdamore/tcell/v2"
	xdraw "golang.org/x/image/draw"
)

// quadrantRunes maps a 4-bit mask to the Unicode block/quadrant character whose
// filled quadrants match the set bits. Bit layout: bit0=UL, bit1=UR, bit2=LL, bit3=LR.
var quadrantRunes = [16]rune{
	' ', '▘', '▝', '▀',
	'▖', '▌', '▞', '▛',
	'▗', '▚', '▐', '▜',
	'▄', '▙', '▟', '█',
}

// onesCount4 is a fast popcount lookup for 4-bit values.
var onesCount4 = [16]int{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}

// Image renders src as a block of terminal cells using Unicode quadrant block
// characters with true-colour foreground and background. Each cell covers a
// 2×2 pixel region, giving double the resolution of plain half-block rendering
// in both axes.
//
// The optimal two-colour palette and quadrant character are chosen per cell by
// minimising the sum of squared colour errors over all 16 possible foreground/
// background splits of the four pixels.
//
// width is the output width in terminal cells and must be greater than zero.
// height is the output height in terminal cells; pass 0 to derive it
// automatically from the image's aspect ratio, assuming a 2:1 cell
// pixel-height-to-width ratio (the typical terminal cell shape).
//
// The caller is responsible for registering any required image-format decoders:
//
//	import _ "image/png"
//	import _ "image/jpeg"
//
// Image does not use hooks and may be called outside a component function.
func Image(src image.Image, width, height int) Element {
	if src == nil || width <= 0 {
		return Box()
	}
	b := src.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return Box()
	}

	cellW := width
	var cellH int
	if height > 0 {
		cellH = height
	} else {
		// pixW = cellW*2 pixels; derive cellH from aspect ratio.
		pixW := cellW * 2
		pixH := pixW * b.Dy() / b.Dx()
		if pixH < 2 {
			pixH = 2
		}
		if pixH%2 != 0 {
			pixH++
		}
		cellH = pixH / 2
	}

	pixW := cellW * 2
	pixH := cellH * 2
	scaled := imageScale(src, pixW, pixH)

	rows := make([]Element, cellH)
	for cy := 0; cy < cellH; cy++ {
		cells := make([]Element, cellW)
		for cx := 0; cx < cellW; cx++ {
			// Sample the 2×2 pixel block for this cell: UL, UR, LL, LR.
			var px [4]color.NRGBA
			px[0] = scaled.NRGBAAt(cx*2, cy*2)
			px[1] = scaled.NRGBAAt(cx*2+1, cy*2)
			px[2] = scaled.NRGBAAt(cx*2, cy*2+1)
			px[3] = scaled.NRGBAAt(cx*2+1, cy*2+1)
			r, style := quadrantCell(px)
			cells[cx] = Text(string(r), style)
		}
		rows[cy] = Row(cells...)
	}
	return Box(rows...)
}

// quadrantCell finds the quadrant character and two-colour Style that best
// approximates the given 2×2 pixel block. It tries all 16 foreground/background
// splits, picks the one with the lowest sum-of-squared-colour-error, and breaks
// ties in favour of more foreground pixels (avoiding invisible spaces).
func quadrantCell(px [4]color.NRGBA) (rune, Style) {
	bestErr := math.MaxFloat64
	bestMask := uint8(0)
	var bestFg, bestBg color.NRGBA

	for mask := uint8(0); mask < 16; mask++ {
		var fgR, fgG, fgB, bgR, bgG, bgB float64
		fgN, bgN := 0, 0
		for i := 0; i < 4; i++ {
			p := px[i]
			if mask>>uint(i)&1 == 1 {
				fgR += float64(p.R); fgG += float64(p.G); fgB += float64(p.B)
				fgN++
			} else {
				bgR += float64(p.R); bgG += float64(p.G); bgB += float64(p.B)
				bgN++
			}
		}

		var fg, bg color.NRGBA
		if fgN > 0 {
			fg = color.NRGBA{R: uint8(fgR / float64(fgN)), G: uint8(fgG / float64(fgN)), B: uint8(fgB / float64(fgN)), A: 255}
		}
		if bgN > 0 {
			bg = color.NRGBA{R: uint8(bgR / float64(bgN)), G: uint8(bgG / float64(bgN)), B: uint8(bgB / float64(bgN)), A: 255}
		}

		var err float64
		for i := 0; i < 4; i++ {
			p := px[i]
			var c color.NRGBA
			if mask>>uint(i)&1 == 1 {
				c = fg
			} else {
				c = bg
			}
			dr := float64(p.R) - float64(c.R)
			dg := float64(p.G) - float64(c.G)
			db := float64(p.B) - float64(c.B)
			err += dr*dr + dg*dg + db*db
		}

		// Prefer more foreground pixels when error is tied, so solid-colour
		// blocks render as '█' rather than invisible spaces.
		if err < bestErr || (err == bestErr && onesCount4[mask] > onesCount4[bestMask]) {
			bestErr = err
			bestMask = mask
			bestFg = fg
			bestBg = bg
		}
	}

	style := Style{
		fg: tcell.NewRGBColor(int32(bestFg.R), int32(bestFg.G), int32(bestFg.B)),
		bg: tcell.NewRGBColor(int32(bestBg.R), int32(bestBg.G), int32(bestBg.B)),
	}
	return quadrantRunes[bestMask], style
}

// imageScale returns a Catmull-Rom scaled copy of src at (w, h) pixels.
// Catmull-Rom (bicubic) produces sharp results for both upscaling and
// downscaling, making it suitable for fitting high-resolution source images
// into the low-resolution terminal pixel grid.
func imageScale(src image.Image, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}
