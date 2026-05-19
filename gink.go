// Package gink is a declarative, React-like TUI framework for Go.
// Components are plain functions that return Elements. State changes
// trigger re-renders automatically via the hooks system.
package gink

import "github.com/gdamore/tcell/v2"

// Component is a function that describes a piece of UI.
type Component func() Element

// Render starts the Gink runtime with the given root component.
// It blocks until the user quits (Escape or Ctrl+C).
func Render(root Component) error {
	r, err := newRenderer()
	if err != nil {
		return err
	}
	defer r.screen.Fini()

	rec := NewReconciler(r)

	doRender := func() {
		prevFocusables = append(prevFocusables[:0], focusables...)
		inputHandlers = inputHandlers[:0]
		pendingEffects = pendingEffects[:0]
		focusables = focusables[:0]
		w, h := r.screen.Size()
		currentTermSize = TermSize{Width: w, Height: h}
		rec.FooterBuf = nil
		virtual := rec.Render(C(root), w, virtualHeight(h))
		footer := rec.FooterBuf
		fh := 0
		if footer != nil {
			fh = footer.Height
		}
		footerHeight = fh
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
			for row := 0; row < fh && avail+row < h; row++ {
				copy(screen.Cells[avail+row], footer.Cells[row])
			}
		}
		r.flush(screen)
	}

	render := func() {
		doRender()
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
				doRender()
			}
		}
		focusChanged = false
		runEffects() // always after final flush so effects see the committed UI
		// Clamp after render when the full focusables list is known.
		// If clamping changed focusedIdx, schedule a re-render so components
		// immediately reflect the corrected focus state.
		if len(focusables) > 0 && focusedIdx >= len(focusables) {
			focusedIdx = len(focusables) - 1
			r.scheduleRender()
		}
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
				switch ev.Buttons() {
				case tcell.WheelUp:
					scrollUp(3)
					render()
				case tcell.WheelDown:
					scrollDown(3)
					render()
				}
			case *tcell.EventKey:
				switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return nil
			case tcell.KeyPgUp:
				scrollUp(currentTermSize.Height)
			case tcell.KeyPgDn:
				scrollDown(currentTermSize.Height)
			case tcell.KeyTab:
				advanceFocus(1)
			case tcell.KeyBacktab:
				advanceFocus(-1)
			default:
				dispatchKey(ev)
			}
				render()
			}

		case <-r.dirty:
			render()
		}
	}
}
