package gink

import (
	"fmt"
	"strings"
)

// componentCache stores the rendered output of a component so it can be
// restored on subsequent renders when the component's state has not changed.
type componentCache struct {
	w, h              int
	cells             [][]Cell         // [row][col], relative to the component's top-left corner
	focusedIdx        int              // global focusedIdx at time of render
	focusablePaths    []string         // focusables entries added by this subtree
	inputHandlerCache []func(KeyEvent) // inputHandlers entries added by this subtree
}

// Reconciler walks an Element tree and paints it into a Buffer.
// It owns the per-component hook stores, keyed by tree path.
type Reconciler struct {
	hooks     map[string]*renderContext
	renderer  *renderer
	cellCache map[string]componentCache
}

func NewReconciler(r *renderer) *Reconciler {
	return &Reconciler{
		hooks:     make(map[string]*renderContext),
		renderer:  r,
		cellCache: make(map[string]componentCache),
	}
}

// hooksFor returns the renderContext for a given tree path, creating it on first use.
func (rec *Reconciler) hooksFor(path string) *renderContext {
	ctx, ok := rec.hooks[path]
	if !ok {
		ctx = &renderContext{renderer: rec.renderer, path: path}
		rec.hooks[path] = ctx
	}
	ctx.hookIndex = 0
	return ctx
}

// needsRender returns true if the component at path must be re-rendered:
// no cached output exists (first render), focus changed, the component itself
// is dirty, any descendant is dirty, or any ancestor component is dirty
// (ancestor dirty means the parent produced a new closure with updated props).
func (rec *Reconciler) needsRender(path string) bool {
	cache, cached := rec.cellCache[path]
	if !cached {
		return true
	}
	if cache.focusedIdx != focusedIdx {
		return true
	}
	if rec.renderer.dirtyPaths[path] {
		return true
	}
	// Descendant dirty → must traverse into this component to reach it.
	prefix := path + "/"
	for dp := range rec.renderer.dirtyPaths {
		if strings.HasPrefix(dp, prefix) {
			return true
		}
	}
	// Ancestor component dirty → parent produced a new closure with updated props.
	parts := strings.Split(path, "/")
	for i := 1; i < len(parts); i++ {
		if rec.renderer.dirtyPaths[strings.Join(parts[:i], "/")] {
			return true
		}
	}
	return false
}

func (rec *Reconciler) Render(root Element, w, h int) *Buffer {
	buf := NewBuffer(w, h)
	rec.renderElement(root, buf, 0, 0, "root")
	rec.renderer.dirtyPaths = nil
	return buf
}

// renderElement paints el into buf at (x, y) and returns (width, height) of the painted area.
// path uniquely identifies this element's position in the tree across renders.
func (rec *Reconciler) renderElement(el Element, buf *Buffer, x, y int, path string) (int, int) {
	switch el.Type {
	case "component":
		fn := el.Props.(func() Element)

		if !rec.needsRender(path) {
			// Restore cached cells and side effects at the current position.
			cache := rec.cellCache[path]
			for row := 0; row < cache.h; row++ {
				for col := 0; col < cache.w; col++ {
					if y+row < buf.Height && x+col < buf.Width {
						buf.Cells[y+row][x+col] = cache.cells[row][col]
					}
				}
			}
			focusables = append(focusables, cache.focusablePaths...)
			inputHandlers = append(inputHandlers, cache.inputHandlerCache...)
			return cache.w, cache.h
		}

		// Snapshot global slice lengths before rendering so we can extract
		// the entries this subtree adds.
		focusablesBefore := len(focusables)
		inputsBefore := len(inputHandlers)

		activePath = path
		activeCtx = rec.hooksFor(path)
		child := fn()
		activeCtx = nil
		activePath = ""

		w, h := rec.renderElement(child, buf, x, y, path+"/0")

		// Build cell cache relative to (x, y).
		cache := componentCache{
			w: w, h: h, focusedIdx: focusedIdx,
			focusablePaths:    append([]string{}, focusables[focusablesBefore:]...),
			inputHandlerCache: append([]func(KeyEvent){}, inputHandlers[inputsBefore:]...),
			cells:             make([][]Cell, h),
		}
		for row := 0; row < h; row++ {
			cache.cells[row] = make([]Cell, w)
			for col := 0; col < w; col++ {
				if y+row < buf.Height && x+col < buf.Width {
					cache.cells[row][col] = buf.Cells[y+row][x+col]
				}
			}
		}
		rec.cellCache[path] = cache

		return w, h

	case "text":
		props := el.Props.(TextProps)
		style := props.Style.toTcell()
		runes := []rune(props.Content)
		for i, ch := range runes {
			if x+i < buf.Width && y < buf.Height {
				buf.Cells[y][x+i] = Cell{Rune: ch, Style: style}
			}
		}
		return len(runes), 1

	case "box":
		props, _ := el.Props.(BoxProps)

		if props.Direction == DirectionRow {
			return rec.renderRow(el.Children, props.Gap, buf, x, y, path)
		}
		return rec.renderColumn(el.Children, props.Gap, buf, x, y, path)
	}

	return 0, 0
}

func (rec *Reconciler) renderColumn(children []Element, gap int, buf *Buffer, x, y int, path string) (int, int) {
	curY := y
	maxW := 0
	for i, child := range children {
		childPath := fmt.Sprintf("%s/%d", path, i)
		w, h := rec.renderElement(child, buf, x, curY, childPath)
		curY += h
		if w > maxW {
			maxW = w
		}
		if i < len(children)-1 {
			curY += gap
		}
	}
	return maxW, curY - y
}

func (rec *Reconciler) renderRow(children []Element, gap int, buf *Buffer, x, y int, path string) (int, int) {
	curX := x
	maxH := 0
	for i, child := range children {
		childPath := fmt.Sprintf("%s/%d", path, i)
		w, h := rec.renderElement(child, buf, curX, y, childPath)
		curX += w
		if h > maxH {
			maxH = h
		}
		if i < len(children)-1 {
			curX += gap
		}
	}
	return curX - x, maxH
}
