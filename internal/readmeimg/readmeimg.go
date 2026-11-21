// Package readmeimg rasterizes gink harness output into PNG and animated GIF
// images for documentation. It draws a minimal terminal-window chrome around
// the captured screen so screenshots look like a real app window rather than a
// wall of text.
//
// The PNG and GIF writers are used by the screenshot tests in the examples
// directory. Regenerate the README assets with:
//
//	GINK_SCREENSHOTS=1 go test ./examples/dashboard ./examples/form -run Screenshot -v
//
// The renderer prefers Cascadia Mono (installed with Windows Terminal) for its
// broad coverage of box-drawing, geometric, and braille glyphs, then falls back
// to embedded Go Mono.
package readmeimg

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"sync"

	"github.com/SummaDiaboli/gink"
	"github.com/gdamore/tcell/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Snapshot is a point-in-time copy of a harness screen and its cell styles.
type Snapshot struct {
	Width  int
	Height int
	Cells  [][]string
	Widths [][]int
	Styles [][]tcell.Style
}

// Capture copies the current contents and styles of h into a Snapshot.
func Capture(h *gink.Harness) *Snapshot {
	s := &Snapshot{
		Width:  h.Width,
		Height: h.Height,
		Cells:  make([][]string, h.Height),
		Widths: make([][]int, h.Height),
		Styles: make([][]tcell.Style, h.Height),
	}
	for y := 0; y < h.Height; y++ {
		row := make([]string, h.Width)
		widths := make([]int, h.Width)
		for x := 0; x < h.Width; x++ {
			row[x] = h.CellGrapheme(x, y)
			widths[x] = h.CellWidth(x, y)
			if widths[x] < 1 {
				widths[x] = 1
			}
		}
		s.Cells[y] = row
		s.Widths[y] = widths
		s.Styles[y] = make([]tcell.Style, h.Width)
		for x := 0; x < h.Width; x++ {
			s.Styles[y][x] = h.CellStyle(x, y)
		}
	}
	return s
}

const (
	chromePad  = 18
	titleBarH  = 34
	imageInset = 26

	fontSize      = 15.0
	titleFontSize = 12.0
)

var (
	fontOnce   sync.Once
	fontErr    error
	fontSource string

	face      font.Face
	titleFace font.Face
	cellW     int
	cellH     int
	ascent    int
)

var (
	canvasBG  = color.RGBA{0x08, 0x0b, 0x10, 0xff}
	windowBG  = color.RGBA{0x0d, 0x11, 0x17, 0xff}
	titleBG   = color.RGBA{0x16, 0x1b, 0x22, 0xff}
	edge      = color.RGBA{0x30, 0x36, 0x3d, 0xff}
	defaultFG = color.RGBA{0xc9, 0xd1, 0xd9, 0xff}
	defaultBG = color.RGBA{0x0a, 0x0e, 0x14, 0xff}
	titleFG   = color.RGBA{0x8b, 0x94, 0x9e, 0xff}
)

var trafficLights = []color.RGBA{
	{0xff, 0x5f, 0x56, 0xff},
	{0xff, 0xbd, 0x2e, 0xff},
	{0x27, 0xc9, 0x3f, 0xff},
}

// FontSource returns a human-readable description of the font used for
// rasterization. It is mainly useful in screenshot test logs.
func FontSource() string {
	if err := initFonts(); err != nil {
		return "unavailable: " + err.Error()
	}
	return fontSource
}

func initFonts() error {
	fontOnce.Do(func() {
		data, src, err := findFont()
		if err != nil {
			fontErr = err
			return
		}
		f, err := opentype.Parse(data)
		if err != nil {
			fontErr = err
			return
		}
		face, err = opentype.NewFace(f, &opentype.FaceOptions{
			Size:    fontSize,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			fontErr = err
			return
		}
		titleFace, err = opentype.NewFace(f, &opentype.FaceOptions{
			Size:    titleFontSize,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			fontErr = err
			return
		}
		fontSource = src

		m := face.Metrics()
		ascent = m.Ascent.Ceil()
		cellH = ascent + m.Descent.Ceil() + 2
		cellW = int(font.MeasureString(face, "M").Ceil()) + 1
	})
	return fontErr
}

func findFont() ([]byte, string, error) {
	candidates := []string{
		filepath.Join(os.Getenv("WINDIR"), "Fonts", "CascadiaMono.ttf"),
		filepath.Join(os.Getenv("WINDIR"), "Fonts", "CascadiaCode.ttf"),
		"/usr/share/fonts/truetype/cascadia/CascadiaMono.ttf",
		"/usr/share/fonts/opentype/cascadia/CascadiaMono.otf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationMono-Regular.ttf",
	}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if data, err := os.ReadFile(p); err == nil {
			return data, filepath.Base(p), nil
		}
	}
	return gomono.TTF, "Go Mono (embedded)", nil
}

// Image renders a snapshot as a PNG-ready RGBA image. The title is drawn in the
// window title bar.
func Image(s *Snapshot, title string) (*image.RGBA, error) {
	if err := initFonts(); err != nil {
		return nil, err
	}
	w, h := imageSize(s)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	render(img, s, title)
	return img, nil
}

// WritePNG renders a snapshot and writes it to path as a PNG.
func WritePNG(path string, s *Snapshot, title string) error {
	img, err := Image(s, title)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// WriteGIF renders each snapshot as a frame and writes an animated GIF.
// delay is the per-frame delay in milliseconds.
func WriteGIF(path string, frames []*Snapshot, title string, delayMS int) error {
	if len(frames) == 0 {
		return errors.New("readmeimg: no frames")
	}
	if err := initFonts(); err != nil {
		return err
	}
	if delayMS <= 0 {
		delayMS = 150
	}
	pal := gifPalette()
	g := &gif.GIF{
		LoopCount: 0,
		Config: image.Config{
			ColorModel: pal,
		},
	}
	for _, s := range frames {
		w, h := imageSize(s)
		frame := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		render(frame, s, title)
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, delayMS/10)
		g.Disposal = append(g.Disposal, gif.DisposalNone)
	}
	g.Config.Width, g.Config.Height = g.Image[0].Bounds().Dx(), g.Image[0].Bounds().Dy()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gif.EncodeAll(f, g)
}

func imageSize(s *Snapshot) (int, int) {
	winW := chromePad*2 + s.Width*cellW
	winH := titleBarH + chromePad*2 + s.Height*cellH
	return winW + imageInset*2, winH + imageInset*2
}

func render(dst draw.Image, s *Snapshot, title string) {
	winW := chromePad*2 + s.Width*cellW
	winH := titleBarH + chromePad*2 + s.Height*cellH
	w := winW + imageInset*2
	h := winH + imageInset*2

	draw.Draw(dst, image.Rect(0, 0, w, h), image.NewUniform(canvasBG), image.Point{}, draw.Src)

	shadow := image.Rect(imageInset+6, imageInset+8, imageInset+winW+8, imageInset+winH+8)
	draw.Draw(dst, shadow, image.NewUniform(color.NRGBA{0, 0, 0, 110}), image.Point{}, draw.Over)

	draw.Draw(dst, image.Rect(imageInset, imageInset, imageInset+winW, imageInset+winH), image.NewUniform(edge), image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(imageInset+1, imageInset+1, imageInset+winW-1, imageInset+winH-1), image.NewUniform(windowBG), image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(imageInset+1, imageInset+1, imageInset+winW-1, imageInset+titleBarH), image.NewUniform(titleBG), image.Point{}, draw.Src)
	draw.Draw(dst, image.Rect(imageInset+1, imageInset+titleBarH, imageInset+winW-1, imageInset+titleBarH+1), image.NewUniform(edge), image.Point{}, draw.Src)

	for i, c := range trafficLights {
		fillCircle(dst, imageInset+20+i*20, imageInset+titleBarH/2, 6, c)
	}
	drawCenteredText(dst, image.Rect(imageInset, imageInset, imageInset+winW, imageInset+titleBarH), title, titleFace, titleFG)

	ox := imageInset + chromePad
	oy := imageInset + titleBarH + chromePad
	draw.Draw(dst, image.Rect(ox, oy, ox+s.Width*cellW, oy+s.Height*cellH), image.NewUniform(defaultBG), image.Point{}, draw.Src)

	for y := 0; y < s.Height; y++ {
		base := oy + y*cellH + ascent + 1
		for x := 0; x < s.Width; {
			glyph := s.Cells[y][x]
			if glyph == "" {
				x++
				continue
			}
			span := s.Widths[y][x]
			if span < 1 {
				span = 1
			}
			if span > s.Width-x {
				span = s.Width - x
			}
			drawCell(dst, ox+x*cellW, oy+y*cellH, base, glyph, span, s.Styles[y][x])
			x += span
		}
	}
}

func drawCell(dst draw.Image, px, py, base int, glyph string, span int, style tcell.Style) {
	fg, bg, attr := style.Decompose()
	fgCol := mapColor(fg, defaultFG)
	bgCol := mapColor(bg, defaultBG)
	if attr&tcell.AttrReverse != 0 {
		fgCol, bgCol = bgCol, fgCol
	}

	spanW := cellW * span
	draw.Draw(dst, image.Rect(px, py, px+spanW, py+cellH), image.NewUniform(bgCol), image.Point{}, draw.Src)
	if glyph == "" || glyph == " " {
		return
	}

	adv := font.MeasureString(face, glyph).Ceil()
	xoff := (spanW - adv) / 2
	if xoff < 0 {
		xoff = 0
	}
	d := &font.Drawer{Dst: dst, Src: image.NewUniform(fgCol), Face: face}
	d.Dot = fixed.P(px+xoff, base)
	d.DrawString(glyph)
	if attr&tcell.AttrBold != 0 {
		d.Dot = fixed.P(px+xoff+1, base)
		d.DrawString(glyph)
	}
	if attr&tcell.AttrUnderline != 0 {
		uy := base + 2
		draw.Draw(dst, image.Rect(px, uy, px+spanW, uy+1), image.NewUniform(fgCol), image.Point{}, draw.Src)
	}
}

func mapColor(c tcell.Color, fallback color.RGBA) color.RGBA {
	if !c.Valid() {
		return fallback
	}
	r, g, b := c.RGB()
	if r < 0 || g < 0 || b < 0 {
		return fallback
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}
}

func fillCircle(dst draw.Image, cx, cy, r int, c color.Color) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r {
				dst.Set(x, y, c)
			}
		}
	}
}

func drawCenteredText(dst draw.Image, bounds image.Rectangle, text string, f font.Face, c color.Color) {
	w := font.MeasureString(f, text).Ceil()
	m := f.Metrics()
	x := bounds.Min.X + (bounds.Dx()-w)/2
	y := bounds.Min.Y + (bounds.Dy()+(m.Ascent-m.Descent).Ceil())/2
	d := &font.Drawer{Dst: dst, Src: image.NewUniform(c), Face: f}
	d.Dot = fixed.P(x, y)
	d.DrawString(text)
}

func gifPalette() color.Palette {
	p := color.Palette{
		canvasBG, windowBG, titleBG, edge, defaultBG, defaultFG, titleFG,
		trafficLights[0], trafficLights[1], trafficLights[2],
	}
	for i := 0; i < 32; i++ {
		v := uint8(i * 255 / 31)
		p = append(p, color.RGBA{v, v, v, 0xff})
	}
	named := []tcell.Color{
		tcell.ColorBlack, tcell.ColorMaroon, tcell.ColorGreen, tcell.ColorOlive,
		tcell.ColorNavy, tcell.ColorPurple, tcell.ColorTeal, tcell.ColorSilver,
		tcell.ColorGray, tcell.ColorRed, tcell.ColorLime, tcell.ColorYellow,
		tcell.ColorBlue, tcell.ColorFuchsia, tcell.ColorAqua, tcell.ColorWhite,
	}
	for _, c := range named {
		r, g, b := c.RGB()
		if r >= 0 && g >= 0 && b >= 0 {
			p = append(p, color.RGBA{uint8(r), uint8(g), uint8(b), 0xff})
		}
	}
	return dedupePalette(p)
}

func dedupePalette(p color.Palette) color.Palette {
	seen := make(map[color.RGBA]bool, len(p))
	out := make(color.Palette, 0, len(p))
	for _, c := range p {
		r, g, b, a := c.RGBA()
		key := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, key)
	}
	if len(out) == 0 {
		out = append(out, color.Black)
	}
	return out
}
