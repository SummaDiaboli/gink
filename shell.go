package gink

// AppShell splits the application into a scrollable main area and a footer
// that is always pinned to the bottom of the terminal, even when the main
// content overflows and the user has scrolled down.
//
// Use this for hint bars, status lines, or any one-line UI chrome that must
// remain visible regardless of viewport position:
//
//	func App() gink.Element {
//	    main   := gink.Box(header, content, table)
//	    footer := gink.Text("Tab: focus  ·  Esc: quit", hintStyle)
//	    return gink.AppShell(main, footer)
//	}
func AppShell(main, footer Element) Element {
	return Element{
		Type:     "shell",
		Children: []Element{main, footer},
	}
}
