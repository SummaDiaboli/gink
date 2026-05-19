package gink

import "strings"

// BorderProps holds the configuration for a border element.
// Used internally by the reconciler; not part of the public API.
type BorderProps struct {
	Title string
	Style Style
}

// Border wraps child in a single-line box border using Unicode line-drawing characters:
//
//	┌─────────┐
//	│  child  │
//	└─────────┘
//
// The child is inset by one cell on every side. The total element size is
// the child width + 2 and child height + 2.
//
// An optional [Style] controls the color and decoration of the border
// characters only — the child retains its own style.
//
//	gink.Border(gink.C(gink.Spinner))
//	gink.Border(content, gink.NewStyle().Foreground(gink.ColorBrightCyan))
func Border(child Element, styles ...Style) Element {
	var style Style
	if len(styles) > 0 {
		style = styles[0]
	}
	return Element{
		Type:     "border",
		Props:    BorderProps{Style: style},
		Children: []Element{child},
	}
}

// BorderWithTitle wraps child in a border with a label inset into the top edge:
//
//	┌─ Title ────────┐
//	│     child      │
//	└────────────────┘
//
// The title is left-aligned after the top-left corner. If the child is too
// narrow to fit the title, the border widens to accommodate it.
// An optional [Style] applies to the border characters and the title text.
//
//	gink.BorderWithTitle("Preview", previewElement)
//	gink.BorderWithTitle("Stats", statsBox, gink.NewStyle().Bold())
func BorderWithTitle(title string, child Element, styles ...Style) Element {
	var style Style
	if len(styles) > 0 {
		style = styles[0]
	}
	return Element{
		Type:     "border",
		Props:    BorderProps{Title: title, Style: style},
		Children: []Element{child},
	}
}

// buildTopBorder constructs the top border line for a given inner width.
// With no title: ┌───┐
// With title:    ┌─ Title ─────┐  (title is left-aligned, remaining width filled with ─)
func buildTopBorder(title string, innerWidth int) string {
	if title == "" {
		return "┌" + strings.Repeat("─", innerWidth) + "┐"
	}

	titleRunes := []rune(title)

	// Layout: ┌ ─   title   ─ ... ┐
	// Fixed overhead for framing the title: "─ " before + " ─" after = 4 chars.
	overhead := 4
	remaining := innerWidth - overhead - len(titleRunes)

	if remaining < 0 {
		// Title doesn't fit cleanly — truncate it to leave room for the frame.
		maxTitle := innerWidth - overhead
		if maxTitle < 0 {
			maxTitle = 0
		}
		if len(titleRunes) > maxTitle {
			titleRunes = titleRunes[:maxTitle]
		}
		remaining = 0
	}

	return "┌─ " + string(titleRunes) + " ─" + strings.Repeat("─", remaining) + "┐"
}
