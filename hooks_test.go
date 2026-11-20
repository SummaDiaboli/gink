package gink

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ── UseState ──────────────────────────────────────────────────────────────────

func TestUseState_initialValue(t *testing.T) {
	h := NewHarness(t, func() Element {
		count, _ := UseState(42)
		return Text(fmt.Sprintf("%d", count))
	})
	defer h.Close()

	if h.Line(0) != "42" {
		t.Errorf("initial value: got %q, want 42", h.Line(0))
	}
}

func TestUseState_setterUpdatesValue(t *testing.T) {
	var set func(int)

	h := NewHarness(t, func() Element {
		count, setCount := UseState(0)
		set = setCount
		return Text(fmt.Sprintf("%d", count))
	})
	defer h.Close()

	set(7)
	h.Render()

	if h.Line(0) != "7" {
		t.Errorf("after setter: got %q, want 7", h.Line(0))
	}
}

func TestUseState_multipleStateVars(t *testing.T) {
	h := NewHarness(t, func() Element {
		a, _ := UseState("hello")
		b, _ := UseState("world")
		return Row(Text(a), Text(" "), Text(b))
	})
	defer h.Close()

	if h.Line(0) != "hello world" {
		t.Errorf("multiple states: got %q, want 'hello world'", h.Line(0))
	}
}

func TestUseState_independentAcrossInstances(t *testing.T) {
	makeComp := func(initial int) func() Element {
		return func() Element {
			val, _ := UseState(initial)
			return Text(fmt.Sprintf("%d", val))
		}
	}

	h := NewHarness(t, func() Element {
		return Box(
			C(makeComp(1)),
			C(makeComp(2)),
			C(makeComp(3)),
		)
	})
	defer h.Close()

	if h.Line(0) != "1" || h.Line(1) != "2" || h.Line(2) != "3" {
		t.Errorf("independent instances: lines %v, want [1 2 3]", h.Lines()[:3])
	}
}

func TestUseState_preservedAcrossRerenders(t *testing.T) {
	var set func(string)

	h := NewHarness(t, func() Element {
		val, setVal := UseState("initial")
		set = setVal
		return Text(val)
	})
	defer h.Close()

	set("updated")
	h.Render()
	h.Render() // second re-render must keep "updated", not reset to "initial"

	if h.Line(0) != "updated" {
		t.Errorf("state after two renders: got %q, want updated", h.Line(0))
	}
}

// ── UseEffect ────────────────────────────────────────────────────────────────

func TestUseEffect_runsOnFirstRender(t *testing.T) {
	ran := false

	h := NewHarness(t, func() Element {
		UseEffect(func() func() {
			ran = true
			return nil
		}, nil)
		return Text("")
	})
	defer h.Close()

	if !ran {
		t.Error("UseEffect with nil deps did not run on first render")
	}
}

func TestUseEffect_mountOnly_runsOnce(t *testing.T) {
	count := 0

	h := NewHarness(t, func() Element {
		UseEffect(func() func() {
			count++
			return nil
		}, []any{}) // mount-only
		return Text("")
	})
	defer h.Close()

	h.Render()
	h.Render()

	if count != 1 {
		t.Errorf("mount-only effect ran %d times, want 1", count)
	}
}

func TestUseEffect_runsWhenDepChanges(t *testing.T) {
	var set func(int)
	effectCount := 0

	h := NewHarness(t, func() Element {
		val, setVal := UseState(0)
		set = setVal
		UseEffect(func() func() {
			effectCount++
			return nil
		}, []any{val})
		return Text(fmt.Sprintf("%d", val))
	})
	defer h.Close()

	initial := effectCount
	set(1)
	h.Render()

	if effectCount != initial+1 {
		t.Errorf("effect ran %d times after dep change, want %d", effectCount, initial+1)
	}
}

func TestUseEffect_skipsWhenDepUnchanged(t *testing.T) {
	effectCount := 0

	h := NewHarness(t, func() Element {
		val, _ := UseState(42) // never changes
		UseEffect(func() func() {
			effectCount++
			return nil
		}, []any{val})
		return Text("")
	})
	defer h.Close()

	after := effectCount
	h.Render()
	h.Render()

	if effectCount != after {
		t.Errorf("effect ran %d extra times with unchanged dep, want 0", effectCount-after)
	}
}

func TestUseEffect_cleanupCalledBeforeNextEffect(t *testing.T) {
	var set func(int)
	var log []string

	h := NewHarness(t, func() Element {
		val, setVal := UseState(0)
		set = setVal
		UseEffect(func() func() {
			log = append(log, fmt.Sprintf("run:%d", val))
			return func() {
				log = append(log, fmt.Sprintf("cleanup:%d", val))
			}
		}, []any{val})
		return Text("")
	})
	defer h.Close()

	set(1)
	h.Render()

	// Expected: run:0, then on second render: cleanup:0, run:1
	if len(log) < 3 {
		t.Fatalf("effect log too short: %v", log)
	}
	if log[0] != "run:0" {
		t.Errorf("log[0] = %q, want run:0", log[0])
	}
	if log[1] != "cleanup:0" {
		t.Errorf("log[1] = %q, want cleanup:0", log[1])
	}
	if log[2] != "run:1" {
		t.Errorf("log[2] = %q, want run:1", log[2])
	}
}

func TestUseEffect_nilDepsRunsEveryRender(t *testing.T) {
	count := 0
	var set func(int)

	h := NewHarness(t, func() Element {
		_, setVal := UseState(0)
		set = setVal
		UseEffect(func() func() {
			count++
			return nil
		}, nil)
		return Text("")
	})
	defer h.Close()

	// With partial render, nil-dep effects run whenever the component re-renders.
	// Trigger two re-renders via state changes.
	before := count
	set(1)
	h.Render()
	set(2)
	h.Render()

	if count != before+2 {
		t.Errorf("nil deps: ran %d extra times, want 2", count-before)
	}
}

// ── UseRef ───────────────────────────────────────────────────────────────────

func TestUseRef_returnsStablePointer(t *testing.T) {
	var first, second *Ref[int]
	var set func(int)

	h := NewHarness(t, func() Element {
		_, setVal := UseState(0)
		set = setVal
		ref := UseRef(0)
		if first == nil {
			first = ref
		} else {
			second = ref
		}
		return Text("")
	})
	defer h.Close()

	// Trigger a re-render via state change so the component runs a second time.
	set(1)
	h.Render()

	if first != second {
		t.Error("UseRef returned different pointer on second render")
	}
}

func TestUseRef_mutationPersistsAcrossRenders(t *testing.T) {
	var ref *Ref[int]

	h := NewHarness(t, func() Element {
		ref = UseRef(0)
		return Text("")
	})
	defer h.Close()

	ref.Value = 99
	h.Render()

	if ref.Value != 99 {
		t.Errorf("ref.Value after mutation + re-render: got %d, want 99", ref.Value)
	}
}

func TestUseRef_doesNotTriggerRerender(t *testing.T) {
	renderCount := 0
	var ref *Ref[int]

	h := NewHarness(t, func() Element {
		renderCount++
		ref = UseRef(0)
		return Text("")
	})
	defer h.Close()

	before := renderCount
	ref.Value = 42 // should NOT schedule a render
	// No Render() call — check that dirty channel is empty
	select {
	case <-h.r.dirty:
		t.Error("UseRef mutation scheduled a re-render unexpectedly")
	default:
		// correct — channel is empty
	}
	_ = before
}

// ── hook order violations ───────────────────────────────────────────────────

func TestHookOrder_panicsWhenHookSkippedConditionally(t *testing.T) {
	var set func(bool)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when a hook call is skipped conditionally, got none")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "different number of hooks") {
			t.Errorf("panic message = %q, want it to mention hook count mismatch", msg)
		}
	}()

	h := NewHarness(t, func() Element {
		callExtra, setCallExtra := UseState(false)
		set = setCallExtra
		if callExtra {
			UseState(0)
		}
		return Text("")
	})
	defer h.Close()

	set(true)
	h.Render()
}

func TestHookOrder_panicsWhenHooksReordered(t *testing.T) {
	var set func(bool)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when hooks are reordered between renders, got none")
		}
		msg := fmt.Sprintf("%v", r)
		if !strings.Contains(msg, "previous render called") {
			t.Errorf("panic message = %q, want it to mention a hook kind mismatch", msg)
		}
	}()

	h := NewHarness(t, func() Element {
		swap, setSwap := UseState(false)
		set = setSwap
		if swap {
			UseEffect(func() func() { return nil }, nil)
			UseState(0)
		} else {
			UseState(0)
			UseEffect(func() func() { return nil }, nil)
		}
		return Text("")
	})
	defer h.Close()

	set(true)
	h.Render()
}

func TestHookOrder_noPanicWhenHooksConsistent(t *testing.T) {
	var set func(int)

	h := NewHarness(t, func() Element {
		count, setCount := UseState(0)
		set = setCount
		UseEffect(func() func() { return nil }, []any{count})
		UseRef(0)
		return Text(fmt.Sprintf("%d", count))
	})
	defer h.Close()

	set(1)
	h.Render()
	set(2)
	h.Render()

	if h.Line(0) != "2" {
		t.Errorf("consistent hook usage: got %q, want 2", h.Line(0))
	}
}

// ── UseInput ─────────────────────────────────────────────────────────────────

func TestUseInput_runeHandlerCalled(t *testing.T) {
	var last rune

	h := NewHarness(t, func() Element {
		UseInput(func(ev KeyEvent) {
			last = ev.Rune
		})
		return Text("")
	})
	defer h.Close()

	h.SendRune('x')

	if last != 'x' {
		t.Errorf("UseInput: got rune %q, want 'x'", last)
	}
}

func TestUseInput_specialKeyHandlerCalled(t *testing.T) {
	pressed := false

	h := NewHarness(t, func() Element {
		UseInput(func(ev KeyEvent) {
			if ev.Key == KeyUp {
				pressed = true
			}
		})
		return Text("")
	})
	defer h.Close()

	h.SendKey(KeyUp)

	if !pressed {
		t.Error("UseInput: KeyUp handler was not called")
	}
}

func TestUseInput_handlerSeesCurrentState(t *testing.T) {
	var set func(int)
	var handlerSaw int

	h := NewHarness(t, func() Element {
		count, setCount := UseState(0)
		set = setCount
		UseInput(func(ev KeyEvent) {
			handlerSaw = count // must see current count, not stale
		})
		return Text(fmt.Sprintf("%d", count))
	})
	defer h.Close()

	set(5)
	h.Render()
	h.SendRune('a')

	if handlerSaw != 5 {
		t.Errorf("handler saw count=%d, want 5 (stale closure bug)", handlerSaw)
	}
}

func TestUseInput_multipleHandlers(t *testing.T) {
	calls := 0

	h := NewHarness(t, func() Element {
		UseInput(func(ev KeyEvent) { calls++ })
		UseInput(func(ev KeyEvent) { calls++ })
		return Text("")
	})
	defer h.Close()

	h.SendRune('z')

	if calls != 2 {
		t.Errorf("multiple handlers: got %d calls, want 2", calls)
	}
}

// ── UseInterval ───────────────────────────────────────────────────────────────

func TestUseInterval_callbackFires(t *testing.T) {
	fired := make(chan struct{}, 5)

	h := NewHarness(t, func() Element {
		UseInterval(10*time.Millisecond, func() {
			select {
			case fired <- struct{}{}:
			default:
			}
		})
		return Text("")
	})
	defer h.Close()

	select {
	case <-fired:
		// passed
	case <-time.After(300 * time.Millisecond):
		t.Error("UseInterval callback did not fire within 300ms")
	}
}

func TestUseInterval_callbackNotStale(t *testing.T) {
	var set func(int)
	latest := make(chan int, 10)

	h := NewHarness(t, func() Element {
		count, setCount := UseState(0)
		set = setCount
		UseInterval(10*time.Millisecond, func() {
			select {
			case latest <- count:
			default:
			}
		})
		return Text(fmt.Sprintf("%d", count))
	})
	defer h.Close()

	// Change state, re-render so the ref is updated
	set(99)
	h.Render()

	// Wait for the interval to fire and check it sees the updated value
	deadline := time.After(300 * time.Millisecond)
	for {
		select {
		case v := <-latest:
			if v == 99 {
				return // saw the updated value — no stale closure
			}
		case <-deadline:
			t.Error("UseInterval still seeing stale count after state update")
			return
		}
	}
}

// ── UseFocus ─────────────────────────────────────────────────────────────────

func TestUseFocus_firstComponentFocusedByDefault(t *testing.T) {
	var aFocused, bFocused bool

	compA := func() Element {
		aFocused = UseFocus()
		return Text("A")
	}
	compB := func() Element {
		bFocused = UseFocus()
		return Text("B")
	}

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	if !aFocused {
		t.Error("first focusable (A) should be focused by default")
	}
	if bFocused {
		t.Error("second focusable (B) should not be focused initially")
	}
}

func TestUseFocus_tabAdvancesFocus(t *testing.T) {
	var aFocused, bFocused bool

	compA := func() Element {
		aFocused = UseFocus()
		return Text("A")
	}
	compB := func() Element {
		bFocused = UseFocus()
		return Text("B")
	}

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	h.Tab()

	if aFocused {
		t.Error("after Tab, A should not be focused")
	}
	if !bFocused {
		t.Error("after Tab, B should be focused")
	}
}

func TestUseFocus_tabWrapsAround(t *testing.T) {
	var aFocused bool

	compA := func() Element {
		aFocused = UseFocus()
		return Text("A")
	}
	compB := func() Element {
		UseFocus()
		return Text("B")
	}

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	h.Tab() // focus B
	h.Tab() // wrap back to A

	if !aFocused {
		t.Error("focus should have wrapped back to A after two Tabs")
	}
}

func TestUseFocus_shiftTabMovesFocusBack(t *testing.T) {
	var aFocused, bFocused bool

	compA := func() Element { aFocused = UseFocus(); return Text("A") }
	compB := func() Element { bFocused = UseFocus(); return Text("B") }

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	h.Tab()      // A→B
	h.ShiftTab() // B→A

	if !aFocused {
		t.Error("after Tab then ShiftTab, A should be focused")
	}
	if bFocused {
		t.Error("after Tab then ShiftTab, B should not be focused")
	}
}

func TestUseFocus_clampsWhenFocusablesShrink(t *testing.T) {
	var setShowBoth func(bool)
	var aFocused bool

	compA := func() Element { aFocused = UseFocus(); return Text("A") }
	// compB registers as focusable but we only care about A's focus after B is removed.
	compB := func() Element { UseFocus(); return Text("B") }

	h := NewHarness(t, func() Element {
		val, set := UseState(true)
		setShowBoth = set
		if val {
			return Box(C(compA), C(compB))
		}
		return Box(C(compA))
	})
	defer h.Close()

	h.Tab() // move focus to B (focusedIdx = 1)

	// Remove B from the tree. focusedIdx=1 is now out of bounds (only A remains).
	setShowBoth(false)
	h.Render()

	// After clamping and the automatic re-render, A should hold focus.
	if !aFocused {
		t.Error("after focusable count shrinks, focus should clamp to A (last valid index)")
	}
}

func TestUseEffect_cleanupRunsOnUnmount(t *testing.T) {
	var setShow func(bool)
	cleanups := 0

	child := func() Element {
		UseEffect(func() func() {
			return func() { cleanups++ }
		}, []any{})
		return Text("child")
	}

	h := NewHarness(t, func() Element {
		show, set := UseState(true)
		setShow = set
		if show {
			return Box(Text("parent"), C(child))
		}
		return Box(Text("parent"))
	})
	defer h.Close()

	if cleanups != 0 {
		t.Fatalf("cleanups before unmount = %d, want 0", cleanups)
	}

	setShow(false)
	h.Render()

	if cleanups != 1 {
		t.Errorf("cleanups after unmount = %d, want 1", cleanups)
	}
}

func TestUseEffect_cleanupStopsGoroutineOnUnmount(t *testing.T) {
	var setShow func(bool)
	stopped := make(chan struct{})

	child := func() Element {
		UseEffect(func() func() {
			done := make(chan struct{})
			go func() {
				defer close(stopped)
				<-done
			}()
			return func() { close(done) }
		}, []any{})
		return Text("worker")
	}

	h := NewHarness(t, func() Element {
		show, set := UseState(true)
		setShow = set
		if show {
			return C(child)
		}
		return Text("hidden")
	})
	defer h.Close()

	setShow(false)
	h.Render()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("effect goroutine was not stopped on unmount")
	}
}

func TestUnmountRemountResetsHookState(t *testing.T) {
	var setShow func(bool)
	var setChild func(int)
	runs, cleanups := 0, 0

	child := func() Element {
		count, set := UseState(0)
		setChild = set
		UseEffect(func() func() {
			runs++
			return func() { cleanups++ }
		}, []any{})
		return Text(fmt.Sprintf("count:%d", count))
	}

	h := NewHarness(t, func() Element {
		show, set := UseState(true)
		setShow = set
		if show {
			return C(child)
		}
		return Text("hidden")
	})
	defer h.Close()

	setChild(5)
	h.Render()
	if !h.Contains("count:5") {
		t.Fatal("expected child state update to render")
	}

	setShow(false)
	h.Render()
	if cleanups != 1 {
		t.Fatalf("cleanups after unmount = %d, want 1", cleanups)
	}

	setShow(true)
	h.Render()

	if !h.Contains("count:0") {
		t.Errorf("child state after remount should reset, got lines %v", h.Lines())
	}
	if runs != 2 {
		t.Errorf("effect runs after remount = %d, want 2", runs)
	}
	if cleanups != 1 {
		t.Errorf("cleanups after remount = %d, want 1", cleanups)
	}
}
