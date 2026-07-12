// Package gink is a declarative, React-like TUI framework for Go.
// Components are plain functions that return Elements. State changes
// trigger re-renders automatically via the hooks system.
package gink

import "github.com/gdamore/tcell/v2"

// Component is a function that describes a piece of UI.
type Component func() Element

// renderPass resets per-frame render state, runs the reconciler, and paints
// the result into the screen buffer. Shared by Render's event loop and
// [Harness.Render] so the two stay in sync.
func renderPass(r *renderer, rec *Reconciler, root Component, w, h int) {
	prevFocusables = append(prevFocusables[:0], focusables...)
	inputHandlers = inputHandlers[:0]
	keyboardHandlers = keyboardHandlers[:0]
	clickHandlers = clickHandlers[:0]
	scrollHandlers = scrollHandlers[:0]
	pendingEffects = pendingEffects[:0]
	focusables = focusables[:0]
	focusBarrier = ""
	renderOffsetX = 0
	renderOffsetY = 0
	currentTermSize = TermSize{Width: w, Height: h}
	accessibilityHints = map[string]string{}
	activeBindings = activeBindings[:0]
	rec.FooterBuf = nil
	virtual := rec.Render(C(root), w, virtualHeight(h))
	footer := rec.FooterBuf
	footerHeight = 0
	if footer != nil {
		footerHeight = footer.Height
	}
	avail := availableHeight()
	scrollContent = detectContentHeight(virtual)
	clampScroll()
	main := applyScroll(virtual, w, avail, scrollOffset)
	addScrollIndicators(main, scrollOffset, scrollContent)
	screen := NewBuffer(w, h)
	for row := 0; row < avail && row < h; row++ {
		copy(screen.Cells[row], main.Cells[row])
	}
	if footer != nil {
		for row := 0; row < footerHeight && avail+row < h; row++ {
			copy(screen.Cells[avail+row], footer.Cells[row])
		}
	}
	r.flush(screen)
	resolveAccessibilityLabel()
}

// commitRender runs renderPass, then applies auto-scroll-to-focus, focus-
// barrier snapping, and focus clamping — each re-running renderPass if it
// changed state. Shared by Render's event loop and [Harness.Render].
//
// onClamp is invoked if focus had to be clamped because the focusable count
// shrank. Render's event loop schedules an async re-render via r.scheduleRender
// (consumed by its r.dirty select case); Harness has no event loop, so it
// re-renders synchronously instead.
func commitRender(r *renderer, rec *Reconciler, root Component, w, h int, onClamp func()) {
	renderPass(r, rec, root, w, h)
	// Auto-scroll only when Tab/Shift+Tab changed the focused component —
	// never override an explicit PageUp/PageDown or mouse wheel scroll.
	if focusChanged && focusedIdx < len(focusables) {
		f := focusables[focusedIdx]
		fy := f.y
		fh := 1
		if cache, ok := rec.cellCache[f.path]; ok {
			fh = cache.h
		}
		avail := availableHeight()
		bottomY := fy + fh - 1
		if avail > 0 && (fy < scrollOffset || bottomY >= scrollOffset+avail) {
			if fy < scrollOffset || fh > avail {
				scrollToY(fy)
			} else {
				scrollToY(bottomY)
			}
			renderPass(r, rec, root, w, h)
		}
	}
	prevIdx := focusedIdx
	snapFocusToBarrier()
	if focusedIdx != prevIdx {
		renderPass(r, rec, root, w, h) // reflect snapped focus in components
	}
	focusChanged = false
	runEffects() // always after final flush so effects see the committed UI
	// Clamp after render when the full focusables list is known.
	// If clamping changed focusedIdx, schedule a re-render so components
	// immediately reflect the corrected focus state.
	if len(focusables) > 0 && focusedIdx >= len(focusables) {
		focusedIdx = len(focusables) - 1
		onClamp()
	}
}

// Render starts the Gink runtime with the given root component.
// It blocks until the user quits (Escape or Ctrl+C).
func Render(root Component) error {
	r, err := newRenderer()
	if err != nil {
		return err
	}
	defer r.screen.Fini()

	rec := NewReconciler(r)

	render := func() {
		w, h := r.screen.Size()
		commitRender(r, rec, root, w, h, r.scheduleRender)
		r.screen.SetTitle(currentAccessibilityLabel)
	}

	render()

	// PollEvent blocks indefinitely. Running it in a goroutine lets us
	// select between terminal events and dirty signals from state updates,
	// so goroutines (timers, async effects) can trigger re-renders without
	// waiting for the next keypress.
	events := make(chan tcell.Event, 1) // buffered so the polling goroutine never blocks on exit
	go func() {
		for {
			ev := r.screen.PollEvent()
			if ev == nil {
				close(events)
				return
			}
			events <- ev
		}
	}()

	handleMouse := func(ev *tcell.EventMouse) {
		switch ev.Buttons() {
		case tcell.WheelUp:
			scrollBy(-wheelDelta)
		case tcell.WheelDown:
			scrollBy(wheelDelta)
		case tcell.Button1:
			mx, my := ev.Position()
			dispatchClick(mx, my)
		default:
			return // no re-render needed
		}
		render()
	}

	// handleKey processes a key event and reports whether Render should exit.
	handleKey := func(ev *tcell.EventKey) bool {
		switch ev.Key() {
		case tcell.KeyCtrlC:
			return true
		case tcell.KeyEscape:
			if focusBarrier != "" {
				dispatchKey(ev)
			} else {
				return true
			}
		case tcell.KeyPgUp:
			scrollBy(-currentTermSize.Height)
		case tcell.KeyPgDn:
			scrollBy(currentTermSize.Height)
		case tcell.KeyTab:
			advanceFocus(1)
		case tcell.KeyBacktab:
			advanceFocus(-1)
		default:
			dispatchKey(ev)
		}
		render()
		return false
	}

	for {
		select {
		case ev, ok := <-events:
			if !ok {
				return nil
			}
			switch ev := ev.(type) {
			case *tcell.EventResize:
				r.screen.Sync()
				render()
			case *tcell.EventMouse:
				handleMouse(ev)
			case *tcell.EventKey:
				if handleKey(ev) {
					return nil
				}
			}

		case <-r.dirty:
			render()
		}
	}
}
