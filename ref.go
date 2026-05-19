package gink

// Ref holds a mutable value that persists across renders without triggering re-renders.
// Access the value via ref.Value — reads and writes are direct field access with no locking.
//
// The primary use case is keeping effect or interval callbacks up-to-date without
// listing them in a UseEffect dependency array. Because UseRef updates ref.Value every
// render, goroutines that read ref.Value always get the latest closure — avoiding the
// stale capture bug that occurs when a goroutine captures a variable directly.
//
//	countRef := gink.UseRef(count)
//	countRef.Value = count  // update to latest every render
//
//	gink.UseEffect(func() func() {
//	    go func() {
//	        fmt.Println(countRef.Value)  // always current, never stale
//	    }()
//	    return nil
//	}, []any{})
type Ref[T any] struct {
	Value T
}

// UseRef returns a stable *Ref[T] whose pointer identity is preserved across renders.
// Mutating ref.Value never schedules a re-render — use UseState if you need the UI
// to update when the value changes.
//
// initialValue is used only when the Ref is first created (on the component's first render).
//
//	ref := gink.UseRef("")          // string ref, initial value ""
//	ref := gink.UseRef(0)           // int ref, initial value 0
//	ref := gink.UseRef(myCallback)  // function ref
func UseRef[T any](initial T) *Ref[T] {
	if activeCtx == nil {
		panic("gink: UseRef called outside of a component render — hooks must be called at the top level of a component function")
	}

	ctx := activeCtx
	idx := ctx.hookIndex
	ctx.hookIndex++

	if idx >= len(ctx.slots) {
		// Store the pointer itself so the same Ref[T] is returned on every render.
		ctx.slots = append(ctx.slots, hookSlot{value: &Ref[T]{Value: initial}})
	}

	return ctx.slots[idx].value.(*Ref[T])
}
