package gink

// TermSize holds the current terminal dimensions in character cells.
type TermSize struct {
	Width  int // number of columns
	Height int // number of rows
}

// currentTermSize is set before every render pass so all calls to UseTermSize
// within a render see dimensions consistent with the buffer being painted.
var currentTermSize TermSize

// UseTermSize returns the current terminal dimensions.
//
// The component re-renders automatically on resize — no subscription or event
// handling is needed. UseTermSize simply reads the dimensions that were captured
// at the start of the current render pass.
//
// TermSize can be passed as a UseEffect dependency to run a side effect on resize:
//
//	size := gink.UseTermSize()
//
//	// Full-width divider that adapts on resize
//	divider := strings.Repeat("─", size.Width)
//
//	// Side effect when terminal size changes
//	gink.UseEffect(func() func() {
//	    log.Printf("resized to %dx%d", size.Width, size.Height)
//	    return nil
//	}, []any{size})
//
// Note: UseTermSize does not consume a hook slot. It reads a package-level
// variable set by the runtime and has no per-render state of its own.
func UseTermSize() TermSize {
	if activeCtx == nil {
		panic("gink: UseTermSize called outside of a component render — hooks must be called at the top level of a component function")
	}
	return currentTermSize
}
