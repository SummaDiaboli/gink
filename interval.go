package gink

import "time"

// UseInterval calls fn on a repeating timer with the given duration.
// The timer is created once when the component mounts and is stopped
// automatically when d changes or the component unmounts.
//
// fn is always the latest closure — UseRef updates the callback reference
// on every render so the goroutine never calls a stale version of fn, even
// though the underlying effect only runs once (when d changes).
//
//	// Update a clock every second
//	gink.UseInterval(time.Second, func() {
//	    setNow(time.Now())
//	})
//
//	// Animate at ~60fps
//	gink.UseInterval(16*time.Millisecond, func() {
//	    setFrame(frame + 1)
//	})
//
// UseInterval is equivalent to UseEffect + UseRef + a ticker. Use UseEffect
// directly when you need to start or stop the timer based on a condition other
// than the duration changing.
func UseInterval(d time.Duration, fn func()) {
	// fnRef always points to the fn from the most recent render.
	// The goroutine reads through the ref, so it never calls a stale closure
	// even though the effect dep is only [d].
	fnRef := UseRef(fn)
	fnRef.Value = fn

	UseEffect(func() func() {
		ticker := time.NewTicker(d)
		go func() {
			for range ticker.C {
				fnRef.Value()
			}
		}()
		return ticker.Stop
	}, []any{d})
}
