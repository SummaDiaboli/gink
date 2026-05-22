package gink

import "github.com/gdamore/tcell/v2"

// Binding describes a keyboard shortcut for registration with [UseKeybinding].
// Set Key to [KeyRune] and Rune to the character for printable keys.
// Set Key to a special key constant (e.g. [KeyEnter], [KeyUp]) for non-printable keys.
type Binding struct {
	Key         tcell.Key // KeyRune for printable characters; special key constant otherwise
	Rune        rune   // the character; only used when Key == KeyRune
	Label       string // short display string shown in help, e.g. "n", "Ctrl+S", "↑"
	Description string // human-readable description, e.g. "New file"
}

type registeredBinding struct {
	Binding
	handler func()
}

// activeBindings accumulates bindings registered during the current render pass.
// Cleared at the start of each render and rebuilt as components run.
// Cache restoration in the reconciler ensures that bindings from cached
// (non-re-rendering) components are included, in tree order.
var activeBindings []registeredBinding

// UseKeybinding registers a named keyboard shortcut. It fires handler whenever
// the key described by b is pressed (globally, regardless of focus — same
// semantics as [UseKeyboard]). The binding is also recorded in the registry so
// [KeybindingHelp] can display it.
//
//	gink.UseKeybinding(gink.Binding{Key: gink.KeyRune, Rune: '?', Label: "?", Description: "Show help"}, func() {
//	    setShowHelp(true)
//	})
//	gink.UseKeybinding(gink.Binding{Key: gink.KeyEnter, Label: "Enter", Description: "Confirm"}, func() {
//	    confirm()
//	})
func UseKeybinding(b Binding, handler func()) {
	activeBindings = append(activeBindings, registeredBinding{b, handler})
	UseKeyboard(func(ev KeyEvent) {
		var matched bool
		// Treat Key == 0 (zero value) with a non-zero Rune as KeyRune so callers
		// that omit Key from the struct literal still get correct rune matching.
		isRune := b.Key == KeyRune || (b.Key == 0 && b.Rune != 0)
		if isRune {
			matched = ev.Key == KeyRune && ev.Rune == b.Rune
		} else {
			matched = ev.Key == b.Key
		}
		if matched {
			handler()
		}
	})
}

// KeybindingHelp returns an Element listing all keybindings registered so far
// in the current render pass. Call it after [UseKeybinding] (or in a component
// that renders after the components that register shortcuts) so all entries are
// visible. Typically mounted inside a help modal or a footer:
//
//	if showHelp {
//	    gink.C(gink.NewModal("Shortcuts", gink.KeybindingHelp(), nil, func() { setShowHelp(false) }))
//	}
func KeybindingHelp(styles ...Style) Element {
	labelStyle := NewStyle()
	if len(styles) > 0 {
		labelStyle = styles[0]
	}
	rows := make([]Element, len(activeBindings))
	for i, b := range activeBindings {
		rows[i] = Row(
			Width(10, Text(b.Label, labelStyle)),
			Text(b.Description),
		)
	}
	return Box(rows...)
}
