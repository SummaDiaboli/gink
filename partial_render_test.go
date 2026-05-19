package gink

import (
	"fmt"
	"testing"
)

// ── Helpers ───────────────────────────────────────────────────────────────────

// callCounter returns a component that increments *count each time it renders.
func callCounter(count *int, inner func() Element) func() Element {
	return func() Element {
		*count++
		return inner()
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestPartialRender_firstRenderCallsAllComponents(t *testing.T) {
	countA, countB, countC := 0, 0, 0

	h := NewHarness(t, func() Element {
		return Box(
			C(callCounter(&countA, func() Element { return Text("A") })),
			C(callCounter(&countB, func() Element { return Text("B") })),
			C(callCounter(&countC, func() Element { return Text("C") })),
		)
	})
	defer h.Close()

	if countA != 1 || countB != 1 || countC != 1 {
		t.Errorf("first render: A=%d B=%d C=%d, want all 1", countA, countB, countC)
	}
}

func TestPartialRender_cleanSiblingNotCalled(t *testing.T) {
	var setA func(int)
	countB := 0

	compA := func() Element {
		val, set := UseState(0)
		setA = set
		return Text(fmt.Sprintf("A:%d", val))
	}
	compB := callCounter(&countB, func() Element { return Text("B") })

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	afterInit := countB

	setA(1)
	h.Render()

	if countB != afterInit {
		t.Errorf("clean sibling B was called %d extra time(s), want 0", countB-afterInit)
	}
}

func TestPartialRender_dirtyComponentIsCalled(t *testing.T) {
	var setA func(int)
	countA := 0

	compA := func() Element {
		countA++
		val, set := UseState(0)
		setA = set
		return Text(fmt.Sprintf("A:%d", val))
	}

	h := NewHarness(t, func() Element {
		return C(compA)
	})
	defer h.Close()

	afterInit := countA

	setA(99)
	h.Render()

	if countA != afterInit+1 {
		t.Errorf("dirty compA called %d extra times, want 1", countA-afterInit)
	}
}

func TestPartialRender_outputCorrectAfterPartialRender(t *testing.T) {
	var setA func(int)

	compA := func() Element {
		val, set := UseState(0)
		setA = set
		return Text(fmt.Sprintf("A:%d", val))
	}
	compB := func() Element { return Text("B:static") }

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	setA(42)
	h.Render()

	if !h.Contains("A:42") {
		t.Errorf("dirty component output not updated; lines: %v", h.Lines()[:2])
	}
	if !h.Contains("B:static") {
		t.Errorf("clean component output disappeared; lines: %v", h.Lines()[:2])
	}
}

func TestPartialRender_noStateChange_noExtraRenders(t *testing.T) {
	countA, countB := 0, 0

	h := NewHarness(t, func() Element {
		return Box(
			C(callCounter(&countA, func() Element { return Text("A") })),
			C(callCounter(&countB, func() Element { return Text("B") })),
		)
	})
	defer h.Close()

	afterInit := countA + countB

	// Re-render with no state changes — no component should be called.
	h.Render()

	if countA+countB != afterInit {
		t.Errorf("no-state-change re-render called components extra times: A=%d B=%d", countA, countB)
	}
}

func TestPartialRender_multipleDirtyPaths(t *testing.T) {
	var setA func(int)
	var setB func(int)
	countA, countB, countC := 0, 0, 0

	compA := func() Element {
		countA++
		val, set := UseState(0)
		setA = set
		return Text(fmt.Sprintf("A:%d", val))
	}
	compB := func() Element {
		countB++
		val, set := UseState(0)
		setB = set
		return Text(fmt.Sprintf("B:%d", val))
	}
	compC := callCounter(&countC, func() Element { return Text("C") })

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB), C(compC))
	})
	defer h.Close()

	afterInitA, afterInitB, afterInitC := countA, countB, countC

	setA(1)
	setB(2)
	h.Render()

	if countA != afterInitA+1 {
		t.Errorf("compA: got %d extra calls, want 1", countA-afterInitA)
	}
	if countB != afterInitB+1 {
		t.Errorf("compB: got %d extra calls, want 1", countB-afterInitB)
	}
	if countC != afterInitC {
		t.Errorf("clean compC was called %d extra times, want 0", countC-afterInitC)
	}
	if !h.Contains("A:1") || !h.Contains("B:2") {
		t.Errorf("updated output missing; lines: %v", h.Lines()[:3])
	}
}

func TestPartialRender_deeplyNestedDirtyComponent(t *testing.T) {
	var setLeaf func(int)
	countSibling := 0

	leaf := func() Element {
		val, set := UseState(0)
		setLeaf = set
		return Text(fmt.Sprintf("leaf:%d", val))
	}
	sibling := callCounter(&countSibling, func() Element { return Text("sibling") })

	// Deep nesting: root → box → box → leaf  (sibling alongside inner box)
	h := NewHarness(t, func() Element {
		return Box(
			Box(
				Box(C(leaf)),
			),
			C(sibling),
		)
	})
	defer h.Close()

	afterInit := countSibling

	setLeaf(7)
	h.Render()

	if countSibling != afterInit {
		t.Errorf("sibling at same level re-rendered %d extra times, want 0", countSibling-afterInit)
	}
	if !h.Contains("leaf:7") {
		t.Errorf("deep leaf output not updated; lines: %v", h.Lines()[:2])
	}
}

func TestPartialRender_consecutiveUpdates(t *testing.T) {
	var setA func(int)
	countB := 0

	compA := func() Element {
		val, set := UseState(0)
		setA = set
		return Text(fmt.Sprintf("A:%d", val))
	}
	compB := callCounter(&countB, func() Element { return Text("B") })

	h := NewHarness(t, func() Element {
		return Box(C(compA), C(compB))
	})
	defer h.Close()

	afterInit := countB

	// Multiple consecutive state changes to compA — B should never be called.
	setA(1)
	h.Render()
	setA(2)
	h.Render()
	setA(3)
	h.Render()

	if countB != afterInit {
		t.Errorf("B re-rendered across 3 consecutive updates (%d extra calls), want 0", countB-afterInit)
	}
	if !h.Contains("A:3") {
		t.Errorf("final value A:3 not found; lines: %v", h.Lines()[:2])
	}
}

func TestPartialRender_cachedCellsRestoredCorrectly(t *testing.T) {
	var setA func(string)

	compA := func() Element {
		val, set := UseState("hello")
		setA = set
		return Text(val)
	}
	compB := func() Element {
		return Text("world")
	}

	h := NewHarness(t, func() Element {
		return Row(C(compA), Text(" "), C(compB))
	})
	defer h.Close()

	// Verify initial
	if !h.Contains("hello world") {
		t.Fatalf("initial render: expected 'hello world', got %q", h.Line(0))
	}

	setA("goodbye")
	h.Render()

	// compB is clean — its cached cells ("world") must still appear
	if !h.Contains("goodbye world") {
		t.Errorf("after partial render: expected 'goodbye world', got %q", h.Line(0))
	}
}
