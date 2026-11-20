package gink

// UseKeyboard registers a global key handler that fires for every key event,
// regardless of which component currently holds focus. Use it for app-wide
// shortcuts (Ctrl+S, ?, Escape sequences) that should work from anywhere.
//
// Unlike [UseInput], the handler does NOT need to guard with an isFocused check,
// and the component does NOT need to call [UseFocus]. UseKeyboard handlers fire
// before [UseInput] handlers in the same dispatch cycle.
//
//	// Quit on 'q' from anywhere in the app.
//	gink.UseKeyboard(func(ev gink.KeyEvent) {
//	    if ev.Rune == 'q' {
//	        os.Exit(0)
//	    }
//	})
//
//	// Toggle a help overlay with '?'.
//	gink.UseKeyboard(func(ev gink.KeyEvent) {
//	    if ev.Rune == '?' {
//	        setShowHelp(!showHelp)
//	    }
//	})
func UseKeyboard(fn func(KeyEvent)) {
	if activeCtx == nil {
		panic("gink: UseKeyboard called outside of a component render — hooks must be called at the top level of a component function")
	}
	activeCtx.scope.state.keyboardHandlers = append(activeCtx.scope.state.keyboardHandlers, fn)
}
