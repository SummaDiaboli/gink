package gink

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// ── UseAsync ──────────────────────────────────────────────────────────────────

func TestUseAsync_initialStateIsLoading(t *testing.T) {
	// Block the async fn so it never resolves during this test.
	block := make(chan struct{})

	h := NewHarness(t, func() Element {
		_, loading, _ := UseAsync(func() (string, error) {
			<-block
			return "", nil
		}, []any{})
		if loading {
			return Text("loading")
		}
		return Text("done")
	})
	defer func() {
		close(block)
		h.Close()
	}()

	if !h.Contains("loading") {
		t.Error("initial render must have loading=true before async fn completes")
	}
}

func TestUseAsync_returnsValueOnSuccess(t *testing.T) {
	h := NewHarness(t, func() Element {
		val, loading, err := UseAsync(func() (string, error) {
			return "hello", nil
		}, []any{})
		if loading {
			return Text("loading")
		}
		if err != nil {
			return Text("error")
		}
		return Text(val)
	})
	defer h.Close()

	awaitContains(t, h, "hello", 500*time.Millisecond)
}

func TestUseAsync_returnsErrorOnFailure(t *testing.T) {
	h := NewHarness(t, func() Element {
		_, loading, err := UseAsync(func() (string, error) {
			return "", errors.New("oops")
		}, []any{})
		if loading {
			return Text("loading")
		}
		if err != nil {
			return Text("err:" + err.Error())
		}
		return Text("ok")
	})
	defer h.Close()

	awaitContains(t, h, "err:oops", 500*time.Millisecond)
}

func TestUseAsync_rerunsWhenDepsChange(t *testing.T) {
	var setID func(int)

	h := NewHarness(t, func() Element {
		id, setIDFn := UseState(1)
		setID = setIDFn
		val, loading, _ := UseAsync(func() (string, error) {
			return fmt.Sprintf("result-%d", id), nil
		}, []any{id})
		if loading {
			return Text("loading")
		}
		return Text(val)
	})
	defer h.Close()

	awaitContains(t, h, "result-1", 500*time.Millisecond)

	setID(2)
	h.Render()

	awaitContains(t, h, "result-2", 500*time.Millisecond)
}

func TestUseAsync_doesNotRerunWhenDepsUnchanged(t *testing.T) {
	var runs int32
	var triggerRerender func(int)

	h := NewHarness(t, func() Element {
		_, setN := UseState(0)
		triggerRerender = setN
		_, _, _ = UseAsync(func() (string, error) {
			atomic.AddInt32(&runs, 1)
			return "ok", nil
		}, []any{}) // mount-only deps
		return Text("")
	})
	defer h.Close()

	// Wait for the mount run to complete.
	awaitContains(t, h, "", 200*time.Millisecond)
	before := atomic.LoadInt32(&runs)

	// Re-render via unrelated state change — deps haven't changed.
	triggerRerender(1)
	h.Render()
	time.Sleep(50 * time.Millisecond)

	if got := atomic.LoadInt32(&runs); got != before {
		t.Errorf("UseAsync re-ran %d extra times with unchanged deps, want 0", got-before)
	}
}

func TestUseAsync_ignoresStaleResult(t *testing.T) {
	// The fn for id=1 sleeps 200ms; the fn for id=2 returns immediately.
	// We switch to id=2 while id=1 is still running and verify that id=1's
	// result never overwrites the resolved id=2 result.
	var setID func(int)

	h := NewHarness(t, func() Element {
		id, setIDFn := UseState(1)
		setID = setIDFn
		val, loading, _ := UseAsync(func() (string, error) {
			capturedID := id
			if capturedID == 1 {
				time.Sleep(200 * time.Millisecond)
			}
			return fmt.Sprintf("result-%d", capturedID), nil
		}, []any{id})
		if loading {
			return Text(fmt.Sprintf("loading-%d", id))
		}
		return Text(val)
	})
	defer h.Close()

	// Switch to id=2 while id=1 is still sleeping.
	setID(2)
	h.Render()

	// id=2 resolves quickly; wait for it.
	awaitContains(t, h, "result-2", 500*time.Millisecond)

	// Wait long enough for the stale id=1 goroutine to finish and try to write.
	time.Sleep(300 * time.Millisecond)
	h.Render()

	if h.Contains("result-1") {
		t.Errorf("stale result from id=1 overwrote id=2\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}
	if !h.Contains("result-2") {
		t.Errorf("result-2 disappeared after stale result attempt\nscreen:\n%s", strings.Join(h.Lines(), "\n"))
	}
}
