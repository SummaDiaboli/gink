package gink

import (
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
		t.Errorf("style change: got %d updates, want 1", len(updates))
	}
}
