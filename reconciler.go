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
	wasFocusedWithin  bool             // isFocusedWithinPath result at render time
	focusablePaths    []focusable      // focusables entries added by this subtree
	inputHandlerCache []func(KeyEvent) // inputHandlers entries added by this subtree
	clickHandlerCache []clickHandler   // clickHandlers entries added by this subtree
	footerBuf         *Buffer          // non-nil when this subtree rendered a "shell" element
}

// Reconciler walks an Element tree and paints it into a Buffer.
// It owns the per-component hook stores, keyed by tree path.
type Reconciler struct {
	hooks       map[string]*renderContext
	renderer    *renderer
	cellCache   map[string]componentCache
	dirtySnap   map[string]bool // snapshot taken at the start of each Render pass
	FooterBuf   *Buffer         // set when a "shell" element is rendered; nil otherwise
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
	if isFocusedWithinPath(path) != cache.wasFocusedWithin {
		return true
	}
	if rec.dirtySnap[path] {
		return true
	}
	// Descendant dirty → must traverse into this component to reach it.
	prefix := path + "/"
	for dp := range rec.dirtySnap {
		if strings.HasPrefix(dp, prefix) {
			return true
		}
	}
	// Ancestor component dirty → parent produced a new closure with updated props.
	parts := strings.Split(path, "/")
	for i := 1; i < len(parts); i++ {
		if rec.dirtySnap[strings.Join(parts[:i], "/")] {
			return true
		}
	}
	return false
}

func (rec *Reconciler) Render(root Element, w, h int) *Buffer {
	buf := NewBuffer(w, h)
	rec.dirtySnap = rec.renderer.snapshotDirty()
	rec.renderElement(root, buf, 0, 0, "root")
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
			// Restore focusables using relative offsets from cache so nested
			// components land at the correct absolute position.
			for _, f := range cache.focusablePaths {
				focusables = append(focusables, focusable{
					path: f.path,
					y:    y + f.y,
					x:    x + f.x,
					w:    f.w,
					h:    f.h,
				})
			}
			inputHandlers = append(inputHandlers, cache.inputHandlerCache...)
			clickHandlers = append(clickHandlers, cache.clickHandlerCache...)
			if cache.footerBuf != nil {
				rec.FooterBuf = cache.footerBuf
			}
			return cache.w, cache.h
		}

		// Snapshot global slice lengths before rendering so we can extract
		// the entries this subtree adds.
		focusablesBefore := len(focusables)
		inputsBefore := len(inputHandlers)
		clicksBefore := len(clickHandlers)

		activeY = y
		activeX = x
		activePath = path
		activeCtx = rec.hooksFor(path)
		child := fn()
		activeCtx = nil
		activePath = ""

		w, h := rec.renderElement(child, buf, x, y, path+"/0")

		// Backfill component dimensions into its own focusable entry so the
		// hit-tester has accurate bounds for mouse click dispatch.
		for i := focusablesBefore; i < len(focusables); i++ {
			if focusables[i].path == path {
				focusables[i].w = w
				focusables[i].h = h
				break
			}
		}

		// Build cell cache. Store focusable positions relative to (x, y) so
		// they can be restored correctly even if the component moves later.
		relFocusables := make([]focusable, len(focusables)-focusablesBefore)
		for i, f := range focusables[focusablesBefore:] {
			relFocusables[i] = focusable{
				path: f.path,
				y:    f.y - y,
				x:    f.x - x,
				w:    f.w,
				h:    f.h,
			}
		}
		cache := componentCache{
			w: w, h: h, focusedIdx: focusedIdx,
			wasFocusedWithin:  isFocusedWithinPath(path),
			focusablePaths:    relFocusables,
			inputHandlerCache: append([]func(KeyEvent){}, inputHandlers[inputsBefore:]...),
			clickHandlerCache: append([]clickHandler{}, clickHandlers[clicksBefore:]...),
			cells:             make([][]Cell, h),
			footerBuf:         rec.FooterBuf,
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

	case "padding":
		props := el.Props.(Pad)
		if len(el.Children) == 0 {
			return 0, 0
		}
		cw, ch := rec.renderElement(el.Children[0], buf, x+props.Left, y+props.Top, path+"/0")
		return cw + props.Left + props.Right, ch + props.Top + props.Bottom

	case "scrollview":
		props := el.Props.(ScrollViewProps)
		if props.Height <= 0 {
			return 0, 0
		}

		// Render child into a tall sub-buffer to capture its full content.
		subH := props.Height * 8
		if subH < 256 {
			subH = 256
		}
		subBuf := NewBuffer(buf.Width, subH)
		cw, ch := rec.renderElement(props.Child, subBuf, 0, 0, path+"/content")

		// Clamp offset so the viewport never scrolls past the last content row.
		offset := props.Offset
		maxOffset := ch - props.Height
		if maxOffset < 0 {
			maxOffset = 0
		}
		if offset > maxOffset {
			offset = maxOffset
		}
		if offset < 0 {
			offset = 0
		}

		// Copy the visible slice of the sub-buffer into the main buffer.
		for row := 0; row < props.Height; row++ {
			src := row + offset
			if src < subBuf.Height {
				for col := 0; col < cw && x+col < buf.Width; col++ {
					if y+row < buf.Height {
						buf.Cells[y+row][x+col] = subBuf.Cells[src][col]
					}
				}
			}
		}

		// Place scroll indicators in a dedicated column to the right of the content
		// so they never overwrite rendered text. The extra column is only reported
		// in the returned width when content overflows the viewport.
		indW := 0
		if ch > props.Height {
			indW = 1
			indCol := x + cw
			if indCol < buf.Width {
				indStyle := NewStyle().Foreground(ColorBrightWhite).toTcell()
				if offset > 0 && y < buf.Height {
					buf.Cells[y][indCol] = Cell{Rune: '↑', Style: indStyle}
				}
				if offset+props.Height < ch && y+props.Height-1 < buf.Height {
					buf.Cells[y+props.Height-1][indCol] = Cell{Rune: '↓', Style: indStyle}
				}
			}
		}

		return cw + indW, props.Height

	case "shell":
		if len(el.Children) < 2 {
			return 0, 0
		}
		mw, mh := rec.renderElement(el.Children[0], buf, x, y, path+"/main")
		// Render footer into a separate buffer so the runtime can pin it to
		// the bottom of the screen outside the scrollable viewport.
		footerBuf := NewBuffer(buf.Width, 32)
		_, fh := rec.renderElement(el.Children[1], footerBuf, 0, 0, path+"/footer")
		if fh > 0 {
			trimmed := NewBuffer(buf.Width, fh)
			for row := 0; row < fh; row++ {
				copy(trimmed.Cells[row], footerBuf.Cells[row])
			}
			rec.FooterBuf = trimmed
		}
		return mw, mh

	case "constrain":
		props := el.Props.(ConstrainProps)
		if len(el.Children) == 0 {
			return 0, 0
		}

		// Render child into a temporary buffer to capture its natural size
		// without writing directly into the main buffer.
		// Push the real screen offset so that any UseFocus calls inside the
		// child record the correct absolute position for hit-testing.
		subBuf := NewBuffer(buf.Width, buf.Height)
		renderOffsetX += x
		renderOffsetY += y
		cw, ch := rec.renderElement(el.Children[0], subBuf, 0, 0, path+"/0")
		renderOffsetX -= x
		renderOffsetY -= y

		// Compute constrained output dimensions.
		w, h := cw, ch
		if props.MinWidth > 0 && w < props.MinWidth {
			w = props.MinWidth
		}
		if props.MaxWidth > 0 && w > props.MaxWidth {
			w = props.MaxWidth
		}
		if props.MinHeight > 0 && h < props.MinHeight {
			h = props.MinHeight
		}
		if props.MaxHeight > 0 && h > props.MaxHeight {
			h = props.MaxHeight
		}

		// Copy from sub-buffer into main buffer, clipping to the constrained size.
		copyW := cw
		if props.MaxWidth > 0 && copyW > props.MaxWidth {
			copyW = props.MaxWidth
		}
		copyH := ch
		if props.MaxHeight > 0 && copyH > props.MaxHeight {
			copyH = props.MaxHeight
		}
		for row := 0; row < copyH; row++ {
			for col := 0; col < copyW; col++ {
				if y+row < buf.Height && x+col < buf.Width {
					buf.Cells[y+row][x+col] = subBuf.Cells[row][col]
				}
			}
		}
		return w, h

	case "border":
		props := el.Props.(BorderProps)
		if len(el.Children) == 0 {
			return 0, 0
		}
		style := props.Style.toTcell()

		// Render the child inset by one cell on each side.
		childW, childH := rec.renderElement(el.Children[0], buf, x+1, y+1, path+"/0")

		// Draw top border.
		for i, r := range []rune(buildTopBorder(props.Title, childW)) {
			if x+i < buf.Width && y < buf.Height {
				buf.Cells[y][x+i] = Cell{Rune: r, Style: style}
			}
		}

		// Draw left and right vertical edges.
		for row := 0; row < childH; row++ {
			if x < buf.Width && y+1+row < buf.Height {
				buf.Cells[y+1+row][x] = Cell{Rune: '│', Style: style}
			}
			if x+childW+1 < buf.Width && y+1+row < buf.Height {
				buf.Cells[y+1+row][x+childW+1] = Cell{Rune: '│', Style: style}
			}
		}

		// Draw bottom border.
		bottom := "└" + strings.Repeat("─", childW) + "┘"
		for i, r := range []rune(bottom) {
			if x+i < buf.Width && y+childH+1 < buf.Height {
				buf.Cells[y+childH+1][x+i] = Cell{Rune: r, Style: style}
			}
		}

		return childW + 2, childH + 2
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
