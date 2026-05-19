package gink

import "reflect"

// pendingEffects collects effects registered during a render pass.
// They are executed after the buffer is flushed to the terminal, never during render,
// matching React's guarantee that effects see the committed UI.
var pendingEffects []func()

// UseEffect schedules a side effect to run after the current render completes.
//
// The deps slice controls when the effect re-runs:
//
//   - nil       — runs after every render (React: no dep array)
//   - []any{}   — runs once when the component first mounts (React: empty array)
//   - []any{a}  — runs when a changes between renders (React: [a])
//
// fn may return a cleanup function. The cleanup runs before the next time the
// effect fires (or when the component unmounts), making it the right place to
// stop timers, cancel goroutines, or close connections.
//
//	// Run once on mount
//	gink.UseEffect(func() func() {
//	    conn := connect(host)
//	    return func() { conn.Close() }
//	}, []any{})
//
//	// Re-run when host changes
//	gink.UseEffect(func() func() {
//	    conn := connect(host)
//	    return func() { conn.Close() }
//	}, []any{host})
func UseEffect(fn func() func(), deps []any) {
	if activeCtx == nil {
		panic("gink: UseEffect called outside of a component render — hooks must be called at the top level of a component function")
	}

	ctx := activeCtx
	idx := ctx.hookIndex
	ctx.hookIndex++

	if idx >= len(ctx.slots) {
		ctx.slots = append(ctx.slots, hookSlot{})
	}

	slot := &ctx.slots[idx]

	// slot.prevDeps == nil means this slot has never run. Always run on first render
	// regardless of deps. Without this guard, depsEqual(nil, []any{}) returns true
	// and mount-only effects ([]any{}) would be silently skipped.
	if deps != nil && slot.prevDeps != nil && depsEqual(slot.prevDeps, deps) {
		return // deps unchanged — skip
	}

	prevCleanup := slot.cleanup
	slot.prevDeps = deps
	slot.cleanup = nil

	pendingEffects = append(pendingEffects, func() {
		if prevCleanup != nil {
			prevCleanup()
		}
		slot.cleanup = fn()
	})
}

// runEffects executes all effects collected during the last render pass.
// Called by the runtime after flushing the buffer.
func runEffects() {
	for _, fn := range pendingEffects {
		fn()
	}
}

// depsEqual returns true if prev and next have the same length and all corresponding
// values are deeply equal. Uses reflect.DeepEqual so struct, slice, and pointer
// values compare correctly.
func depsEqual(prev, next []any) bool {
	if len(prev) != len(next) {
		return false
	}
	for i := range prev {
		if !reflect.DeepEqual(prev[i], next[i]) {
			return false
		}
	}
	return true
}
