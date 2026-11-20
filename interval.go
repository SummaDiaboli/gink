package gink

import (
	"sync"
	"time"
)

type intervalCallback struct {
	mu sync.RWMutex
	fn func()
}

func (c *intervalCallback) set(fn func()) {
	c.mu.Lock()
	c.fn = fn
	c.mu.Unlock()
}

func (c *intervalCallback) call() {
	c.mu.RLock()
	fn := c.fn
	c.mu.RUnlock()
	if fn != nil {
		fn()
	}
}

// UseInterval calls fn on a repeating timer with the given duration.
// The timer is created once when the component mounts and is stopped
// automatically when d changes or the component unmounts.
//
// fn is always the latest closure — a mutex-protected ref updates the callback
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
	callback := UseRef(&intervalCallback{fn: fn})
	callback.Value.set(fn)

	UseEffect(func() func() {
		ticker := time.NewTicker(d)
		done := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker.C:
					callback.Value.call()
				case <-done:
					return
				}
			}
		}()
		return func() {
			ticker.Stop()
			close(done)
		}
	}, []any{d})
}
