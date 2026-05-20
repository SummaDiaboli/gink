package gink

// Theme holds the named styles that built-in components use when no explicit
// style override is provided. Assign a custom Theme to [ThemeCtx] to restyle
// the entire application at once:
//
//	var myTheme = gink.DefaultTheme
//	myTheme.Focused = gink.NewStyle().Bold().Foreground(gink.ColorBrightGreen)
//	gink.SetContext(gink.ThemeCtx, myTheme)
type Theme struct {
	// Focused is applied to the selected/active row or widget when it has focus.
	Focused Style
	// Accent is used for emphasis, headings, and decorative highlights.
	Accent Style
	// Muted is used for secondary or de-emphasised text.
	Muted Style
	// Error is used for error messages and destructive actions.
	Error Style
	// Success is used for confirmations and positive feedback.
	Success Style
	// Warning is used for warnings and cautionary feedback.
	Warning Style
	// Cursor is applied to the text cursor rendered inside editable fields.
	Cursor Style
}

// DefaultTheme is the built-in theme used when no custom theme has been set.
var DefaultTheme = Theme{
	Focused: NewStyle().Bold().Foreground(ColorBrightCyan),
	Accent:  NewStyle().Bold().Foreground(ColorBrightBlue),
	Muted:   NewStyle().Foreground(ColorWhite),
	Error:   NewStyle().Bold().Foreground(ColorBrightRed),
	Success: NewStyle().Bold().Foreground(ColorBrightGreen),
	Warning: NewStyle().Bold().Foreground(ColorBrightYellow),
	Cursor:  NewStyle().Reverse(),
}

// ThemeCtx is the application-wide theme context. Set it once at startup (or
// at any time via [SetContext]) to change the look of all built-in components.
//
//	gink.SetContext(gink.ThemeCtx, myTheme)
var ThemeCtx = NewContext(DefaultTheme)

// UseTheme returns the current [Theme] from [ThemeCtx]. Call it inside any
// component function to read the active theme:
//
//	theme := gink.UseTheme()
//	return gink.Text("hello", theme.Accent)
func UseTheme() Theme {
	return UseContext(ThemeCtx)
}
