package readmeimg

import (
	"image"
	"image/color"
	"testing"

	"github.com/SummaDiaboli/gink"
	"github.com/gdamore/tcell/v2"
)

func TestCapture_wideGraphemeSpans(t *testing.T) {
	h := gink.NewHarnessSize(t, func() gink.Element {
		return gink.Row(gink.Text("界"), gink.Text("X"))
	}, 4, 1)
	defer h.Close()

	s := Capture(h)

	if got := s.Cells[0][0]; got != "界" {
		t.Errorf("cell 0 = %q, want %q", got, "界")
	}
	if got := s.Widths[0][0]; got != 2 {
		t.Errorf("cell 0 width = %d, want 2", got)
	}
	if got := s.Cells[0][1]; got != "" {
		t.Errorf("cell 1 = %q, want empty", got)
	}
	if got := s.Widths[0][1]; got != 1 {
		t.Errorf("cell 1 width = %d, want 1", got)
	}
	if got := s.Cells[0][2]; got != "X" {
		t.Errorf("cell 2 = %q, want %q", got, "X")
	}
	if got := s.Widths[0][2]; got != 1 {
		t.Errorf("cell 2 width = %d, want 1", got)
	}
	if got := s.Styles[0][0]; got != h.CellStyle(0, 0) {
		t.Errorf("cell 0 style = %v, want %v", got, h.CellStyle(0, 0))
	}
}

func TestRender_skipsContinuationAndAdvancesBySpan(t *testing.T) {
	if err := initFonts(); err != nil {
		t.Fatalf("initFonts: %v", err)
	}

	style := tcell.StyleDefault.Background(tcell.ColorRed)
	s := &Snapshot{
		Width:  4,
		Height: 1,
		Cells:  [][]string{{"界", "", "X", " "}},
		Widths: [][]int{{2, 1, 1, 1}},
		Styles: [][]tcell.Style{{style, style, style, style}},
	}
	w, h := imageSize(s)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	render(img, s, "")

	want := mapColor(tcell.ColorRed, defaultBG)
	ox := imageInset + chromePad
	oy := imageInset + titleBarH + chromePad
	for _, x := range []int{0, 1, 2} {
		px := ox + x*cellW + cellW/2
		py := oy + cellH - 2
		got, ok := img.At(px, py).(color.RGBA)
		if !ok || got != want {
			t.Errorf("cell %d background = %v, want %v", x, got, want)
		}
	}

	s.Widths[0][0] = 99
	render(img, s, "")
}

func TestDrawCell_spanControlsBackgroundAndUnderline(t *testing.T) {
	if err := initFonts(); err != nil {
		t.Fatalf("initFonts: %v", err)
	}

	style := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorYellow).Bold(true).Underline(true)
	img := image.NewRGBA(image.Rect(0, 0, 3*cellW, cellH))
	drawCell(img, 0, 0, ascent+1, "界", 2, style)

	wantBG := mapColor(tcell.ColorBlue, defaultBG)
	for x := 0; x < 2*cellW; x++ {
		got, ok := img.At(x, cellH-2).(color.RGBA)
		if !ok || got != wantBG {
			t.Fatalf("background at x=%d = %v, want %v", x, got, wantBG)
		}
	}

	wantFG := mapColor(tcell.ColorYellow, defaultFG)
	uy := ascent + 1 + 2
	if got, ok := img.At(2*cellW-1, uy).(color.RGBA); !ok || got != wantFG {
		t.Errorf("underline last pixel = %v, want %v", got, wantFG)
	}
	if got := img.At(2*cellW, uy); got != (color.RGBA{}) {
		t.Errorf("underline leaked past span: got %v", got)
	}
}
