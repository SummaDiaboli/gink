package gink

import "sync"

// Context is a typed container for shared state that any component in the tree
// can read without prop drilling. Create one with [NewContext], read its value
// with [UseContext], and update it with [SetContext].
//
// Unlike React's context, this implementation uses a global registry — there
// is one value per context object, shared across the entire application. This
// is the natural fit for TUI apps, which have a single screen and a single
// logical state space (theme, current user, global error state, etc.).
//
//	// Package-level context — created once, shared everywhere.
//	var ThemeCtx = gink.NewContext(DefaultTheme)
//
//	// In a component:
//	theme := gink.UseContext(ThemeCtx)
//
//	// From an event handler or goroutine:
//	gink.SetContext(ThemeCtx, DarkTheme)
type Context[T any] struct {
	mu       sync.Mutex
	value    T
	watchers map[string]*renderContext // path → renderContext for each subscribed component
}

// NewContext creates a new Context with the given default value.
// The returned pointer should typically be stored in a package-level variable
// so all components can reference the same context object.
//
//	var ThemeCtx = gink.NewContext(Theme{Primary: gink.ColorBrightCyan})
//	var CurrentUser = gink.NewContext(User{Name: "guest"})
func NewContext[T any](defaultValue T) *Context[T] {
	return &Context[T]{
		value:    defaultValue,
		watchers: make(map[string]*renderContext),
	}
}

// UseContext reads the current value of ctx and subscribes the calling
// component to future updates. When [SetContext] is called with a new value,
// every subscribed component is scheduled for re-render automatically.
//
// UseContext does not consume a hook slot — it may be called in any order
// relative to [UseState], [UseEffect], and other hooks.
//
//	theme := gink.UseContext(ThemeCtx)
//	return gink.Text("Hello", theme.TitleStyle)
func UseContext[T any](ctx *Context[T]) T {
	if activeCtx == nil {
		panic("gink: UseContext called outside of a component render — hooks must be called at the top level of a component function")
	}
	ctx.mu.Lock()
	ctx.watchers[activeCtx.path] = activeCtx
	val := ctx.value
	ctx.mu.Unlock()
	return val
}

// SetContext updates the value of ctx and schedules a re-render for every
// component that called [UseContext] with this context. Safe to call from
// any goroutine, including [UseEffect] or [UseInterval] callbacks.
//
//	gink.SetContext(ThemeCtx, DarkTheme)
//	gink.SetContext(ErrorCtx, fmt.Errorf("connection lost"))
func SetContext[T any](ctx *Context[T], value T) {
	ctx.mu.Lock()
	ctx.value = value
	watchers := make([]*renderContext, 0, len(ctx.watchers))
	for _, w := range ctx.watchers {
		watchers = append(watchers, w)
	}
	ctx.mu.Unlock()

	// Mark each subscribed component dirty outside the lock to avoid holding
	// ctx.mu while markDirty acquires the renderer's mutex. This is safe
	// because gink's render loop is single-goroutine: renderContext pointers
	// in the snapshot remain valid until the next render cycle.
	for _, w := range watchers {
		w.renderer.markDirty(w.path)
	}
}
