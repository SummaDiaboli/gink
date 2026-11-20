// Package gink is a declarative, React-like TUI framework for Go.
// Components are plain functions that return Elements. State changes
// trigger re-renders automatically via the hooks system.
package gink

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

var renderMu sync.Mutex

// Component is a function that describes a piece of UI.
type Component func() Element

// renderPass resets per-frame render state, runs the reconciler, and paints
// the result into the screen buffer. Shared by Render's event loop and
// [Harness.Render] so the two stay in sync.
func renderPass(r *renderer, rec *Reconciler, root Component, w, h int) {
	activeRenderer.Store(r)
	defer activeRenderer.Store(nil)

	s := &r.state
	s.prevFocusables = append(s.prevFocusables[:0], s.focusables...)
	s.inputHandlers = s.inputHandlers[:0]
	s.keyboardHandlers = s.keyboardHandlers[:0]
	s.clickHandlers = s.clickHandlers[:0]
	s.scrollHandlers = s.scrollHandlers[:0]
	s.pendingEffects = s.pendingEffects[:0]
	s.focusables = s.focusables[:0]
	s.focusBarrier = ""
	s.currentTermSize = TermSize{Width: w, Height: h}
	s.accessibilityHints = map[string]string{}
	s.activeBindings = s.activeBindings[:0]
	rec.FooterBuf = nil
	virtual := rec.Render(C(root), w, virtualHeight(h))
	footer := rec.FooterBuf
	s.footerHeight = 0
	if footer != nil {
		s.footerHeight = footer.Height
	}
	avail := s.availableHeight()
	s.scrollContent = detectContentHeight(virtual)
	s.clampScroll()
	main := applyScroll(virtual, w, avail, s.scrollOffset)
	addScrollIndicators(main, s.scrollOffset, s.scrollContent)
	screen := NewBuffer(w, h)
	for row := 0; row < avail && row < h; row++ {
		copy(screen.Cells[row], main.Cells[row])
	}
	if footer != nil {
		for row := 0; row < s.footerHeight && avail+row < h; row++ {
			copy(screen.Cells[avail+row], footer.Cells[row])
		}
	}
	r.flush(screen)
	lastAccessibilityLabel.Store(s.resolveAccessibilityLabel())
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
	renderMu.Lock()
	defer renderMu.Unlock()

	renderPass(r, rec, root, w, h)
	s := &r.state
	// Auto-scroll only when Tab/Shift+Tab changed the focused component —
	// never override an explicit PageUp/PageDown or mouse wheel scroll.
	if s.focusChanged && s.focusedIdx < len(s.focusables) {
		f := s.focusables[s.focusedIdx]
		fy := f.y
		fh := 1
		if cache, ok := rec.cellCache[f.path]; ok {
			fh = cache.h
		}
		avail := s.availableHeight()
		bottomY := fy + fh - 1
		if avail > 0 && (fy < s.scrollOffset || bottomY >= s.scrollOffset+avail) {
			if fy < s.scrollOffset || fh > avail {
				s.scrollToY(fy)
			} else {
				s.scrollToY(bottomY)
			}
			renderPass(r, rec, root, w, h)
		}
	}
	prevIdx := s.focusedIdx
	s.snapFocusToBarrier()
	if s.focusedIdx != prevIdx {
		renderPass(r, rec, root, w, h) // reflect snapped focus in components
	}
	s.focusChanged = false
	s.runCleanups()
	s.runEffects() // always after final flush so effects see the committed UI
	// Clamp after render when the full focusables list is known.
	// If clamping changed focusedIdx, schedule a re-render so components
	// immediately reflect the corrected focus state.
	if len(s.focusables) > 0 && s.focusedIdx >= len(s.focusables) {
		s.focusedIdx = len(s.focusables) - 1
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
		r.screen.SetTitle(r.state.currentAccessibilityLabel)
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
		s := &r.state
		switch ev.Buttons() {
		case tcell.WheelUp:
			s.scrollBy(-wheelDelta)
		case tcell.WheelDown:
			s.scrollBy(wheelDelta)
		case tcell.Button1:
			mx, my := ev.Position()
			s.dispatchClick(mx, my)
		default:
			return // no re-render needed
		}
		render()
	}

	// handleKey processes a key event and reports whether Render should exit.
	handleKey := func(ev *tcell.EventKey) bool {
		s := &r.state
		switch ev.Key() {
		case tcell.KeyCtrlC:
			return true
		case tcell.KeyEscape:
			if s.focusBarrier != "" {
				s.dispatchKey(ev)
			} else {
				return true
			}
		case tcell.KeyPgUp:
			s.scrollBy(-s.currentTermSize.Height)
		case tcell.KeyPgDn:
			s.scrollBy(s.currentTermSize.Height)
		case tcell.KeyTab:
			s.advanceFocus(1)
		case tcell.KeyBacktab:
			s.advanceFocus(-1)
		default:
			s.dispatchKey(ev)
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
