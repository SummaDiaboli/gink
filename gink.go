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

	render := func() {
		inputHandlers = inputHandlers[:0]
		pendingEffects = pendingEffects[:0]
		focusables = focusables[:0]
		w, h := r.screen.Size()
		currentTermSize = TermSize{Width: w, Height: h}
		buf := rec.Render(C(root), w, h)
		r.flush(buf)
		runEffects() // always after flush so effects see the committed UI
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
			case *tcell.EventKey:
				switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlC:
				return nil
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
