package gink

// hookSlot holds the persisted state for one hook call across renders.
// UseState uses value; UseEffect uses prevDeps and cleanup; UseRef stores a *Ref pointer.
// Slots are matched to hook calls by their position in the call order (hookIndex),
// which is why hooks must always be called in the same order every render.
type hookSlot struct {
	value    any
	prevDeps []any  // UseEffect: dependency values from the last run; nil = never ran
	cleanup  func() // UseEffect: cleanup function returned by the last effect
}

// renderContext tracks hook state for one component instance.
// The reconciler creates one per unique tree path and resets hookIndex to 0
// before each render, enforcing the same call-order contract as React hooks.
type renderContext struct {
	slots     []hookSlot
	hookIndex int
	renderer  *renderer
	path      string
}

// activeCtx is set by the reconciler immediately before invoking a component
// function and cleared after. Rendering is synchronous so no locking is needed.
var activeCtx *renderContext

// UseState adds local state to a component. Returns the current value and a setter.
// Calling the setter with a new value schedules a re-render.
//
// The type parameter T is inferred from initialValue — no explicit type annotation needed.
// initialValue is used only on the first render; it is ignored on subsequent renders.
// The setter is safe to call from any goroutine (e.g. from UseEffect or UseInterval).
//
//	count, setCount := gink.UseState(0)
//	name,  setName  := gink.UseState("Alice")
//	open,  setOpen  := gink.UseState(false)
func UseState[T any](initial T) (T, func(T)) {
	if activeCtx == nil {
		panic("gink: UseState called outside of a component render — hooks must be called at the top level of a component function")
	}
	ctx := activeCtx
	idx := ctx.hookIndex
	ctx.hookIndex++

	if idx >= len(ctx.slots) {
		ctx.slots = append(ctx.slots, hookSlot{value: initial})
	}

	val := ctx.slots[idx].value.(T)

	setter := func(next T) {
		ctx.slots[idx].value = next
		ctx.renderer.markDirty(ctx.path)
	}

	return val, setter
}
