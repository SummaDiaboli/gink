package gink

import "github.com/atotto/clipboard"

// clipRead and clipWrite are the active clipboard backend functions.
// Swappable in tests without touching the system clipboard.
var clipRead  = defaultClipRead
var clipWrite = defaultClipWrite

func defaultClipRead() (string, error)  { return clipboard.ReadAll() }
func defaultClipWrite(s string) error   { return clipboard.WriteAll(s) }

// UseClipboard returns a read and write function for the system clipboard.
// read returns the current clipboard contents (empty string on error).
// write sets the clipboard contents.
//
// To paste clipboard content on Ctrl+V, use the built-in support in
// [NewInput] and [NewTextArea]. Use UseClipboard directly when you need
// explicit control — for example, to copy selected text to the clipboard:
//
//	read, write := gink.UseClipboard()
//	gink.UseKeybinding(gink.Binding{Key: gink.KeyRune, Rune: 'y', Label: "y", Description: "Copy"}, func() {
//	    write(selectedText)
//	})
func UseClipboard() (read func() string, write func(string)) {
	if activeCtx == nil {
		panic("gink: UseClipboard called outside of a component render — hooks must be called at the top level of a component function")
	}
	read = func() string {
		s, _ := clipRead()
		return s
	}
	write = func(s string) {
		_ = clipWrite(s)
	}
	return read, write
}

