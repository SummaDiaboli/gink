package gink

import "github.com/gdamore/tcell/v2"

// KeyEvent is the input event type passed to UseInput handlers.
// It wraps tcell internals so components do not need to import tcell directly.
type KeyEvent struct {
	Rune rune       // the character pressed; valid when Key == KeyRune
	Key  tcell.Key  // the key code; use Key* constants for special keys
}

// Key constants for use in UseInput handlers.
// Printable characters (letters, digits, space, punctuation) arrive with
// Key == KeyRune — match them via KeyEvent.Rune, not these constants.
const (
	KeyEnter      = tcell.KeyEnter      // Enter / Return
	KeyBackspace  = tcell.KeyBackspace2 // Backspace as BS  (0x08) — most terminals
	KeyBackspace2 = tcell.KeyBackspace  // Backspace as DEL (0x7F) — some terminals
	KeyEscape     = tcell.KeyEscape     // Escape
	KeyUp         = tcell.KeyUp         // Up arrow
	KeyDown       = tcell.KeyDown       // Down arrow
	KeyLeft       = tcell.KeyLeft       // Left arrow
	KeyRight      = tcell.KeyRight      // Right arrow
	KeyTab        = tcell.KeyTab        // Tab (consumed by focus system; not dispatched to UseInput)
	KeyPgUp       = tcell.KeyPgUp       // Page Up (consumed by scroll system; not dispatched to UseInput)
	KeyPgDn       = tcell.KeyPgDn       // Page Down (consumed by scroll system; not dispatched to UseInput)
	KeyHome       = tcell.KeyHome       // Home — move to start of line
	KeyEnd        = tcell.KeyEnd        // End  — move to end of line
	KeyRune       = tcell.KeyRune       // any printable character — use ev.Rune to identify it
	KeyCtrlV      = tcell.KeyCtrlV     // Ctrl+V — paste from clipboard
)

// inputHandlers holds the handlers registered by UseInput during the current render pass.
// Cleared before each render and rebuilt as component functions run.
var inputHandlers []func(KeyEvent)

// UseInput registers a keyboard event handler for the current render pass.
// The handler is called for every key event received after the render completes.
// Because handlers are rebuilt on every render, the closure always captures
// the latest state values — stale captures are not a problem.
//
// Tab and Shift+Tab are consumed by the focus system and are not dispatched here.
// Escape and Ctrl+C quit the application and are not dispatched here either.
//
//	// Match printable characters via ev.Rune
//	gink.UseInput(func(ev gink.KeyEvent) {
//	    switch ev.Rune {
//	    case '+': setCount(count + 1)
//	    case '-': setCount(count - 1)
//	    }
//	})
//
//	// Match special keys via ev.Key
//	gink.UseInput(func(ev gink.KeyEvent) {
//	    switch ev.Key {
//	    case gink.KeyUp:    moveUp()
//	    case gink.KeyDown:  moveDown()
//	    case gink.KeyEnter: confirm()
//	    }
//	})
func UseInput(fn func(KeyEvent)) {
	if activeCtx == nil {
		panic("gink: UseInput called outside of a component render — hooks must be called at the top level of a component function")
	}
	inputHandlers = append(inputHandlers, fn)
}

// dispatchKey fans a terminal key event out to all registered handlers.
// Global UseKeyboard handlers fire first, then UseInput handlers.
func dispatchKey(ev *tcell.EventKey) {
	ke := KeyEvent{Rune: ev.Rune(), Key: ev.Key()}
	for _, fn := range keyboardHandlers {
		fn(ke)
	}
	for _, fn := range inputHandlers {
		fn(ke)
	}
}
