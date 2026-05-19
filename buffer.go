package gink

import "github.com/gdamore/tcell/v2"

type Cell struct {
	Rune  rune
	Style tcell.Style
}

type Buffer struct {
	Width  int
	Height int
	Cells  [][]Cell
}

func NewBuffer(w, h int) *Buffer {
	cells := make([][]Cell, h)
	for i := range cells {
		cells[i] = make([]Cell, w)
		for j := range cells[i] {
			cells[i][j] = Cell{Rune: ' ', Style: tcell.StyleDefault}
		}
	}
	return &Buffer{Width: w, Height: h, Cells: cells}
}

// CellUpdate describes a single cell that needs to be written to the terminal.
type CellUpdate struct {
	X, Y int
	Cell Cell
}

// Diff compares prev and next and returns only the cells that changed.
// If the buffers differ in size, all cells in next are returned (resize case).
func Diff(prev, next *Buffer) []CellUpdate {
	if prev.Width != next.Width || prev.Height != next.Height {
		updates := make([]CellUpdate, 0, next.Width*next.Height)
		for y, row := range next.Cells {
			for x, cell := range row {
				updates = append(updates, CellUpdate{X: x, Y: y, Cell: cell})
			}
		}
		return updates
	}

	var updates []CellUpdate
	for y := range next.Cells {
		for x := range next.Cells[y] {
			if prev.Cells[y][x] != next.Cells[y][x] {
				updates = append(updates, CellUpdate{X: x, Y: y, Cell: next.Cells[y][x]})
			}
		}
	}
	return updates
}
