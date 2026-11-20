package gink

import "reflect"

// UseEffect schedules a side effect to run after the current render completes.
//
// The deps slice controls when the effect re-runs:
//
//   - nil       — runs after every render (React: no dep array)
//   - []any{}   — runs once when the component first mounts (React: empty array)
//   - []any{a}  — runs when a changes between renders (React: [a])
//
// fn may return a cleanup function. The cleanup runs before the next time the
// effect fires and when the component is removed from the tree, making it the
// right place to stop timers, cancel goroutines, or close connections.
// Component identity is positional: without keys, a different component at the
// same tree position reuses that position's hook state after cleanup.
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
	idx, _ := ctx.nextSlot(kindEffect)

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

	ctx.scope.state.pendingEffects = append(ctx.scope.state.pendingEffects, func() {
		if prevCleanup != nil {
			prevCleanup()
		}
		slot.cleanup = fn()
	})
}

func (s *runtimeState) queueCleanup(fn func()) {
	s.pendingCleanups = append(s.pendingCleanups, fn)
}

func (s *runtimeState) runCleanups() {
	cleanups := s.pendingCleanups
	s.pendingCleanups = nil
	for _, fn := range cleanups {
		fn()
	}
}

// runEffects executes all effects collected during the last render pass.
// Called by the runtime after flushing the buffer.
func (s *runtimeState) runEffects() {
	effects := s.pendingEffects
	s.pendingEffects = nil
	for _, fn := range effects {
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
