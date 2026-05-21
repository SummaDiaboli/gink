package gink

import "sync/atomic"

// UseAsync runs fn in a goroutine and returns the result, a loading flag, and
// any error. It is the idiomatic alternative to wiring UseEffect + UseState
// together for every data-fetching component.
//
// On the first render (and whenever deps change) loading is true and value is
// the zero value for T. Once fn returns, the component re-renders with the
// result and loading set to false.
//
// deps follows the same rules as [UseEffect]:
//   - nil       — re-runs after every render
//   - []any{}   — runs once when the component mounts
//   - []any{a}  — re-runs whenever a changes
//
// Note: fn itself is not compared between renders. If fn closes over values
// that change independently of deps, include those values in deps explicitly.
//
// When deps change before fn returns, the in-flight goroutine is cancelled:
// its result is silently discarded so stale data never overwrites a newer result.
//
//	// Fetch a user record when userID changes.
//	user, loading, err := gink.UseAsync(func() (User, error) {
//	    return db.GetUser(userID)
//	}, []any{userID})
//	if loading {
//	    return gink.C(gink.Spinner)
//	}
//	if err != nil {
//	    return gink.Text("error: " + err.Error(), errStyle)
//	}
//	return gink.Text(user.Name)
func UseAsync[T any](fn func() (T, error), deps []any) (value T, loading bool, err error) {
	var zero T
	val, setVal := UseState(zero)
	isLoading, setLoading := UseState(true)
	errState, setErr := UseState[error](nil)

	firstRenderRef := UseRef(true)
	isFirstRender := firstRenderRef.Value
	firstRenderRef.Value = false

	depsRef := UseRef[[]any](nil)
	depsChanged := !isFirstRender && (deps == nil || depsRef.Value == nil || !depsEqual(depsRef.Value, deps))
	depsRef.Value = deps

	if depsChanged {
		val = zero
		setVal(zero)
		isLoading = true
		setLoading(true)
		errState = nil
		setErr(nil)
	}

	UseEffect(func() func() {
		// cancelled is written by the cleanup (main goroutine) and read by the
		// async goroutine after fn() returns, so it must be accessed atomically.
		var cancelled int32

		go func() {
			result, e := fn()
			if atomic.LoadInt32(&cancelled) == 1 {
				return
			}
			setVal(result)
			setErr(e)
			setLoading(false)
		}()

		return func() { atomic.StoreInt32(&cancelled, 1) }
	}, deps)

	return val, isLoading, errState
}
