package gink

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

type renderer struct {
	screen     tcell.Screen
	dirty      chan struct{}
	current    *Buffer // last buffer written to the terminal
	mu         sync.Mutex
	dirtyPaths map[string]bool
}

// markDirty records that the component at path needs re-rendering and schedules a render pass.
// Safe to call from any goroutine.
func (r *renderer) markDirty(path string) {
	r.mu.Lock()
	if r.dirtyPaths == nil {
		r.dirtyPaths = make(map[string]bool)
	}
	r.dirtyPaths[path] = true
	r.mu.Unlock()
	r.scheduleRender()
}

// snapshotDirty atomically returns the current dirty set and resets it.
func (r *renderer) snapshotDirty() map[string]bool {
	r.mu.Lock()
	snap := r.dirtyPaths
	r.dirtyPaths = nil
	r.mu.Unlock()
	return snap
}

func newRenderer() (*renderer, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}
	screen.EnableMouse()
	return &renderer{
		screen: screen,
		dirty:  make(chan struct{}, 1),
	}, nil
}

// scheduleRender signals that a re-render is needed. Safe to call from any goroutine.
func (r *renderer) scheduleRender() {
	select {
	case r.dirty <- struct{}{}:
	default:
	}
}

// flush diffs next against the last rendered buffer and writes only changed cells.
func (r *renderer) flush(next *Buffer) {
	var updates []CellUpdate
	if r.current == nil {
		// First render — write everything.
		updates = make([]CellUpdate, 0, next.Width*next.Height)
		for y, row := range next.Cells {
			for x, cell := range row {
				updates = append(updates, CellUpdate{X: x, Y: y, Cell: cell})
			}
		}
	} else {
		updates = Diff(r.current, next)
	}

	for _, u := range updates {
		r.screen.SetContent(u.X, u.Y, u.Cell.Rune, nil, u.Cell.Style)
	}
	if len(updates) > 0 {
		r.screen.Show()
	}

	r.current = next
}
