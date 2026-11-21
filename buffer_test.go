package gink

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestNewBuffer_dimensions(t *testing.T) {
	buf := NewBuffer(10, 5)
	if buf.Width != 10 || buf.Height != 5 {
		t.Errorf("got %dx%d, want 10x5", buf.Width, buf.Height)
	}
	if len(buf.Cells) != 5 {
		t.Errorf("got %d rows, want 5", len(buf.Cells))
	}
	for y, row := range buf.Cells {
		if len(row) != 10 {
			t.Errorf("row %d: got %d cols, want 10", y, len(row))
		}
	}
}

func TestNewBuffer_initializedToSpace(t *testing.T) {
	buf := NewBuffer(5, 3)
	for y, row := range buf.Cells {
		for x, cell := range row {
			if cell.Rune != ' ' {
				t.Errorf("cell (%d,%d): got %q, want space", x, y, cell.Rune)
			}
			if cell.Style != tcell.StyleDefault {
				t.Errorf("cell (%d,%d): style is not default", x, y)
			}
		}
	}
}

func TestDiff_identicalBuffers(t *testing.T) {
	buf := NewBuffer(5, 3)
	updates := Diff(buf, buf)
	if len(updates) != 0 {
		t.Errorf("identical buffers: got %d updates, want 0", len(updates))
	}
}

func TestDiff_singleChange(t *testing.T) {
	prev := NewBuffer(5, 3)
	next := NewBuffer(5, 3)
	next.Cells[1][2] = Cell{Rune: 'X', Style: tcell.StyleDefault}

	updates := Diff(prev, next)
	if len(updates) != 1 {
		t.Fatalf("got %d updates, want 1", len(updates))
	}
	u := updates[0]
	if u.X != 2 || u.Y != 1 {
		t.Errorf("update at (%d,%d), want (2,1)", u.X, u.Y)
	}
	if u.Cell.Rune != 'X' {
		t.Errorf("update rune %q, want 'X'", u.Cell.Rune)
	}
}

func TestDiff_multipleChanges(t *testing.T) {
	prev := NewBuffer(5, 3)
	next := NewBuffer(5, 3)
	next.Cells[0][0] = Cell{Rune: 'A', Style: tcell.StyleDefault}
	next.Cells[2][4] = Cell{Rune: 'Z', Style: tcell.StyleDefault}

	updates := Diff(prev, next)
	if len(updates) != 2 {
		t.Errorf("got %d updates, want 2", len(updates))
	}
}

func TestDiff_noChanges(t *testing.T) {
	prev := NewBuffer(4, 4)
	next := NewBuffer(4, 4)
	// Both initialized to spaces — no diff expected.
	updates := Diff(prev, next)
	if len(updates) != 0 {
		t.Errorf("got %d updates, want 0", len(updates))
	}
}

func TestBuffer_wideGraphemeMetadata(t *testing.T) {
	buf := NewBuffer(3, 1)
	style := NewStyle().Bold().toTcell()

	width := setBufferGrapheme(buf, 0, 0, "界", style)
	if width != 2 {
		t.Fatalf("width: got %d, want 2", width)
	}
	if got := buf.Cells[0][0].Grapheme; got != "界" {
		t.Errorf("base grapheme: got %q, want %q", got, "界")
	}
	if buf.Cells[0][0].Width != 2 || buf.Cells[0][0].Continuation {
		t.Errorf("base cell metadata: got %+v", buf.Cells[0][0])
	}
	if !buf.Cells[0][1].Continuation {
		t.Errorf("cell 1 should be a continuation cell")
	}

	updates := Diff(NewBuffer(3, 1), buf)
	if len(updates) != 2 {
		t.Errorf("updates: got %d, want 2", len(updates))
	}
}

func TestDiff_sizeChange_returnsAllNextCells(t *testing.T) {
	prev := NewBuffer(5, 3)
	next := NewBuffer(8, 4)
	next.Cells[0][0] = Cell{Rune: 'A', Style: tcell.StyleDefault}

	updates := Diff(prev, next)
	if len(updates) != next.Width*next.Height {
		t.Errorf("on resize: got %d updates, want %d (all next cells)", len(updates), next.Width*next.Height)
	}
}

func TestDiff_styleChange(t *testing.T) {
	prev := NewBuffer(5, 3)
	next := NewBuffer(5, 3)
	next.Cells[0][0] = Cell{Rune: ' ', Style: tcell.StyleDefault.Bold(true)}

	updates := Diff(prev, next)
	if len(updates) != 1 {
		t.Errorf("style change: got %d, want 1", len(updates))
	}
}

func TestSetBufferCell_skipsWideClusterThatDoesNotFit(t *testing.T) {
	buf := NewBuffer(4, 1)
	setBufferCell(buf, 3, 0, Cell{Rune: '界', Grapheme: "界", Width: 2})

	cell := buf.Cells[0][3]
	if cell.Grapheme != " " || cell.Width != 1 || cell.Continuation {
		t.Errorf("cell 3: got %+v, want blank", cell)
	}
}

func TestCellRunes(t *testing.T) {
	tests := []struct {
		name     string
		cell     Cell
		wantMain rune
		wantComb []rune
	}{
		{name: "no grapheme", cell: Cell{Rune: 'A'}, wantMain: 'A'},
		{name: "single rune", cell: Cell{Rune: '界', Grapheme: "界"}, wantMain: '界'},
		{name: "combining", cell: Cell{Rune: 'e', Grapheme: "e\u0301"}, wantMain: 'e', wantComb: []rune{'\u0301'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main, comb := cellRunes(tt.cell)
			if main != tt.wantMain {
				t.Errorf("main = %q, want %q", main, tt.wantMain)
			}
			if !reflect.DeepEqual(comb, tt.wantComb) {
				t.Errorf("comb = %q, want %q", comb, tt.wantComb)
			}
		})
	}
}

func TestClearBufferCellSpan_walksBackFromContinuation(t *testing.T) {
	buf := NewBuffer(4, 1)
	setBufferGrapheme(buf, 0, 0, "界", tcell.StyleDefault)

	clearBufferCellSpan(buf, 1, 0)

	for x := 0; x < 2; x++ {
		cell := buf.Cells[0][x]
		if cell.Rune != ' ' || cell.Grapheme != " " || cell.Width != 1 || cell.Continuation {
			t.Errorf("cell %d: got %+v, want blank", x, cell)
		}
	}
}

func TestSetBufferGrapheme_multiRuneCluster(t *testing.T) {
	buf := NewBuffer(4, 1)
	cluster := "👨‍👩‍👧‍👦"

	if width := setBufferGrapheme(buf, 0, 0, cluster, tcell.StyleDefault); width != 2 {
		t.Fatalf("width = %d, want 2", width)
	}
	if got := buf.Cells[0][0].Grapheme; got != cluster {
		t.Errorf("base grapheme = %q, want %q", got, cluster)
	}
	if got := buf.Cells[0][0].Rune; got != []rune(cluster)[0] {
		t.Errorf("base rune = %q, want %q", got, []rune(cluster)[0])
	}
	if buf.Cells[0][0].Width != 2 {
		t.Errorf("base width = %d, want 2", buf.Cells[0][0].Width)
	}
	if !buf.Cells[0][1].Continuation {
		t.Error("cell 1 should be a continuation cell")
	}
}

func TestWriteBufferText_reportsFullWidthWhenClipped(t *testing.T) {
	buf := NewBuffer(3, 1)

	if width := writeBufferText(buf, 0, 0, "世界X", tcell.StyleDefault); width != 5 {
		t.Errorf("width = %d, want 5", width)
	}
	if got := buf.Cells[0][0].Grapheme; got != "世" {
		t.Errorf("cell 0 = %q, want %q", got, "世")
	}
	if got := buf.Cells[0][2].Grapheme; got != " " {
		t.Errorf("cell 2 = %q, want blank", got)
	}
}
