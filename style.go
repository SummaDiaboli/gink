package gink

import "github.com/gdamore/tcell/v2"

// Color is an alias for tcell.Color. Use the Color* constants defined below
// rather than constructing tcell colors directly, so components stay free of
// a tcell import.
type Color = tcell.Color

// Standard and bright terminal color constants.
// Pass these to Style.Foreground and Style.Background.
const (
	ColorDefault = tcell.ColorDefault // terminal's default color

	// Standard (dim) colors
	ColorBlack   = tcell.ColorBlack
	ColorRed     = tcell.ColorMaroon
	ColorGreen   = tcell.ColorGreen
	ColorYellow  = tcell.ColorOlive
	ColorBlue    = tcell.ColorNavy
	ColorMagenta = tcell.ColorPurple
	ColorCyan    = tcell.ColorTeal
	ColorWhite   = tcell.ColorSilver

	// Bright variants
	ColorBrightRed     = tcell.ColorRed
	ColorBrightGreen   = tcell.ColorLime
	ColorBrightYellow  = tcell.ColorYellow
	ColorBrightBlue    = tcell.ColorBlue
	ColorBrightMagenta = tcell.ColorFuchsia
	ColorBrightCyan    = tcell.ColorAqua
	ColorBrightWhite   = tcell.ColorWhite
)

// Style describes the visual appearance of a Text element.
// All methods return a new Style value, making styles safe to share and chain:
//
//	var titleStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
//	var errorStyle = titleStyle.Foreground(gink.ColorBrightRed) // titleStyle is unchanged
type Style struct {
	fg        tcell.Color
	bg        tcell.Color
	bold      bool
	underline bool
	italic    bool
	reverse   bool
}

// NewStyle returns a Style with terminal default colors and no decoration.
// Use it as the starting point for building a style:
//
//	gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
func NewStyle() Style {
	return Style{fg: tcell.ColorDefault, bg: tcell.ColorDefault}
}

// Foreground returns a copy of the style with the text color set to c.
func (s Style) Foreground(c Color) Style { s.fg = c; return s }

// Background returns a copy of the style with the background color set to c.
func (s Style) Background(c Color) Style { s.bg = c; return s }

// Bold returns a copy of the style with bold text enabled.
func (s Style) Bold() Style { s.bold = true; return s }

// Underline returns a copy of the style with underline enabled.
func (s Style) Underline() Style { s.underline = true; return s }

// Italic returns a copy of the style with italic enabled.
// Note: italic rendering depends on terminal support and may display as reverse video.
func (s Style) Italic() Style { s.italic = true; return s }

// Reverse returns a copy of the style with foreground and background colors
// swapped. Commonly used to render a text cursor.
func (s Style) Reverse() Style { s.reverse = true; return s }

// toTcell converts the Gink style to a tcell.Style for use in the renderer.
func (s Style) toTcell() tcell.Style {
	ts := tcell.StyleDefault.Foreground(s.fg).Background(s.bg)
	if s.bold {
		ts = ts.Bold(true)
	}
	if s.underline {
		ts = ts.Underline(true)
	}
	if s.italic {
		ts = ts.Italic(true)
	}
	if s.reverse {
		ts = ts.Reverse(true)
	}
	return ts
}

// NewRGBColor returns a 24-bit true-colour Color for use with
// [Style.Foreground] and [Style.Background]. r, g, b are in the range 0–255.
// Requires a terminal with true-colour support (most modern terminals).
func NewRGBColor(r, g, b int32) Color {
	return tcell.NewRGBColor(r, g, b)
}

// optionalStyle returns the first element of styles and true, or a zero Style
// and false when styles is empty. Use it in component factories to extract the
// optional style parameter:
//
//	explicitStyle, hasExplicit := optionalStyle(styles)
func optionalStyle(styles []Style) (Style, bool) {
	if len(styles) > 0 {
		return styles[0], true
	}
	return Style{}, false
}

// resolveStyle returns explicit when hasExplicit is true, otherwise fallback.
// Pair with [optionalStyle] to implement the optional focus-style override that
// most interactive components expose:
//
//	focusStyle := resolveStyle(explicitStyle, hasExplicit, theme.Focused)
func resolveStyle(explicit Style, hasExplicit bool, fallback Style) Style {
	if hasExplicit {
		return explicit
	}
	return fallback
}

// itemStyle returns the style for a list or table row.
// When the row is not selected it returns the zero style. When selected and
// focused it returns focusStyle; when selected but unfocused it returns Bold so
// the selection remains visible without implying interactivity.
func itemStyle(selected, focused bool, focusStyle Style) Style {
	if !selected {
		return NewStyle()
	}
	if focused {
		return focusStyle
	}
	return NewStyle().Bold()
}

// TextProps holds the content and style for a text element.
// Used internally by the reconciler; not part of the public API.
type TextProps struct {
	Content string
	Style   Style
}
