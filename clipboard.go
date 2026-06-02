package gink

import (
	"strings"

	"github.com/atotto/clipboard"
)

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
// Unlike most gink hooks, UseClipboard is a convenience helper — it holds no
// internal state and does not call UseState or UseEffect. It can be called at
// the top level of a component and the returned functions used freely in
// UseInput handlers, UseEffect callbacks, or event handlers.
//
// [NewInput] and [NewTextArea] handle Ctrl+V paste automatically. Use
// UseClipboard directly when you need explicit control — for example, to copy
// content to the clipboard with a custom shortcut:
//
//	_, write := gink.UseClipboard()
//	gink.UseKeybinding(gink.Binding{Rune: 'y', Label: "y", Description: "Copy"}, func() {
//	    write(selectedText)
//	})
// normalizeNewlines replaces all line-ending sequences (\r\n, \n, \r) in s
// with replacement. Call it on pasted text before inserting into a component:
//
//	text = normalizeNewlines(text, "\n")  // textarea: keep as newlines
//	text = normalizeNewlines(text, " ")   // textinput: collapse to spaces
func normalizeNewlines(s, replacement string) string {
	s = strings.ReplaceAll(s, "\r\n", replacement)
	s = strings.ReplaceAll(s, "\n", replacement)
	s = strings.ReplaceAll(s, "\r", replacement)
	return s
}

func UseClipboard() (read func() string, write func(string)) {
	read = func() string {
		s, _ := clipRead()
		return s
	}
	write = func(s string) {
		_ = clipWrite(s)
	}
	return read, write
}

