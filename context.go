package gink

import "sync"

// Context is a typed container for shared state that any component in the tree
// can read without prop drilling. Create one with [NewContext], read its value
// with [UseContext], and update it with [SetContext].
//
// Values set during a render are scoped to that renderer; values set outside a
// render update the process-wide default used by renderers without a local
// override.
//
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
	values   map[*renderer]T
	watchers map[*renderer]map[string]*renderContext
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
		values:   make(map[*renderer]T),
		watchers: make(map[*renderer]map[string]*renderContext),
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
	r := activeCtx.renderer
	ctx.mu.Lock()
	if ctx.values == nil {
		ctx.values = make(map[*renderer]T)
	}
	if ctx.watchers == nil {
		ctx.watchers = make(map[*renderer]map[string]*renderContext)
	}
	val, ok := ctx.values[r]
	if !ok {
		val = ctx.value
	}
	byPath := ctx.watchers[r]
	if byPath == nil {
		byPath = make(map[string]*renderContext)
		ctx.watchers[r] = byPath
	}
	byPath[activeCtx.path] = activeCtx
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
	r := activeRenderer.Load()
	ctx.mu.Lock()
	if r != nil {
		if ctx.values == nil {
			ctx.values = make(map[*renderer]T)
		}
		ctx.values[r] = value
	} else {
		ctx.value = value
	}
	var watchers []*renderContext
	if r != nil {
		for _, w := range ctx.watchers[r] {
			watchers = append(watchers, w)
		}
	} else {
		for _, byPath := range ctx.watchers {
			for _, w := range byPath {
				watchers = append(watchers, w)
			}
		}
	}
	ctx.mu.Unlock()

	for _, w := range watchers {
		w.renderer.markDirty(w.path)
	}
}
