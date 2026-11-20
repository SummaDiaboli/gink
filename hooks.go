package gink

import "fmt"

// hookKind identifies which hook function created a slot, so slot reuse across
// renders can be checked against the hook actually being called this time.
type hookKind int

const (
	kindState hookKind = iota + 1
	kindEffect
	kindRef
)

func (k hookKind) String() string {
	switch k {
	case kindState:
		return "UseState"
	case kindEffect:
		return "UseEffect"
	case kindRef:
		return "UseRef"
	default:
		return "unknown"
	}
}

// hookSlot holds the persisted state for one hook call across renders.
// UseState uses value; UseEffect uses prevDeps and cleanup; UseRef stores a *Ref pointer.
// Slots are matched to hook calls by their position in the call order (hookIndex),
// which is why hooks must always be called in the same order every render.
type hookSlot struct {
	kind     hookKind
	value    any
	prevDeps []any  // UseEffect: dependency values from the last run; nil = never ran
	cleanup  func() // UseEffect: cleanup function returned by the last effect
}

type renderScope struct {
	path    string
	x       int
	y       int
	offsetX int
	offsetY int
	state   *runtimeState
}

// renderContext tracks hook state for one component instance.
// The reconciler creates one per unique tree path and resets hookIndex to 0
// before each render, enforcing the same call-order contract as React hooks.
type renderContext struct {
	slots         []hookSlot
	hookIndex     int
	prevHookCount int // hooks used on the previous render; -1 means no previous render yet
	renderer      *renderer
	path          string
	scope         *renderScope
}

// nextSlot returns the index of the slot for the next hook call of the given kind
// and whether the slot was just created (true on this component's first render
// reaching this call). On later renders it asserts the slot at this position was
// created by the same kind of hook, panicking immediately if a different hook was
// called here than last time — this catches hooks being reordered between renders
// (e.g. swapping a UseState and UseEffect call) even when the total hook count
// stays the same, which a count-only check would miss.
func (ctx *renderContext) nextSlot(kind hookKind) (int, bool) {
	idx := ctx.hookIndex
	ctx.hookIndex++

	if idx >= len(ctx.slots) {
		ctx.slots = append(ctx.slots, hookSlot{kind: kind})
		return idx, true
	}

	if ctx.slots[idx].kind != kind {
		panic(fmt.Sprintf(
			"gink: component at path %q called %s at hook index %d, but the previous render called %s there — hooks must be called in the same order every render",
			ctx.path, kind, idx, ctx.slots[idx].kind,
		))
	}
	return idx, false
}

// checkHookCount panics if this render used a different number of hooks than the
// previous render did. Called once after a component function returns. This is
// necessary in addition to nextSlot's kind check because skipping or adding a hook
// call conditionally shifts every later index without ever hitting a kind mismatch
// at a stable position.
func (ctx *renderContext) checkHookCount() {
	if ctx.prevHookCount != -1 && ctx.prevHookCount != ctx.hookIndex {
		panic(fmt.Sprintf(
			"gink: component at path %q called a different number of hooks (%d) than the previous render (%d) — hooks must be called unconditionally in the same order every render",
			ctx.path, ctx.hookIndex, ctx.prevHookCount,
		))
	}
	ctx.prevHookCount = ctx.hookIndex
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
	idx, isNew := ctx.nextSlot(kindState)
	if isNew {
		ctx.slots[idx].value = initial
	}

	// Use comma-ok so nil interface values (e.g. UseState[error](nil)) don't panic.
	val, ok := ctx.slots[idx].value.(T)
	if !ok {
		var zero T
		val = zero
	}

	setter := func(next T) {
		ctx.slots[idx].value = next
		ctx.renderer.markDirty(ctx.path)
	}

	return val, setter
}
