package gink

import (
	"fmt"
	"strings"
	"testing"
)

func focusRoot(prefix string) Component {
	return func() Element {
		return Box(C(func() Element {
			if UseFocus() {
				return Text(prefix + "A:focused")
			}
			return Text(prefix + "A:blur")
		}), C(func() Element {
			if UseFocus() {
				return Text(prefix + "B:focused")
			}
			return Text(prefix + "B:blur")
		}))
	}
}

func scrollRoot() Component {
	return func() Element {
		children := make([]Element, 20)
		for i := range children {
			children[i] = Text(fmt.Sprintf("line-%02d", i))
		}
		return Box(children...)
	}
}

func TestRuntimeStateIsolation_interleavedFocus(t *testing.T) {
	t.Parallel()

	a := NewHarness(t, focusRoot("a"))
	defer a.Close()
	b := NewHarness(t, focusRoot("b"))
	defer b.Close()

	if !a.Contains("aA:focused") || !b.Contains("bA:focused") {
		t.Fatalf("initial focus incorrect: a=%v b=%v", a.Lines(), b.Lines())
	}

	a.Tab()
	if !a.Contains("aB:focused") || !b.Contains("bA:focused") {
		t.Fatalf("after a.Tab: a=%v b=%v", a.Lines(), b.Lines())
	}

	b.Tab()
	if !a.Contains("aB:focused") || !b.Contains("bB:focused") {
		t.Fatalf("after b.Tab: a=%v b=%v", a.Lines(), b.Lines())
	}
}

func TestRuntimeStateIsolation_scrollOffsets(t *testing.T) {
	t.Parallel()

	a := NewHarnessSize(t, scrollRoot(), 20, 5)
	defer a.Close()
	b := NewHarnessSize(t, scrollRoot(), 20, 5)
	defer b.Close()

	a.PageDown()
	if !strings.HasPrefix(a.Line(0), "line-05") {
		t.Fatalf("a scroll line: %q", a.Line(0))
	}
	if !strings.HasPrefix(b.Line(0), "line-00") {
		t.Fatalf("b scroll line: %q", b.Line(0))
	}

	b.PageDown()
	b.PageDown()
	if !strings.HasPrefix(a.Line(0), "line-05") {
		t.Fatalf("a scroll line after b scroll: %q", a.Line(0))
	}
	if !strings.HasPrefix(b.Line(0), "line-10") {
		t.Fatalf("b scroll line after b scroll: %q", b.Line(0))
	}
}

func TestRuntimeStateIsolation_effectQueues(t *testing.T) {
	t.Parallel()

	a := &renderer{}
	b := &renderer{}
	aCleanups := 0
	bCleanups := 0
	aEffects := 0
	bEffects := 0

	a.state.queueCleanup(func() { aCleanups++ })
	b.state.queueCleanup(func() { bCleanups++ })
	a.state.runCleanups()

	if aCleanups != 1 || bCleanups != 0 {
		t.Fatalf("cleanups: a=%d b=%d", aCleanups, bCleanups)
	}

	b.state.pendingEffects = append(b.state.pendingEffects, func() { bEffects++ })
	a.state.pendingEffects = append(a.state.pendingEffects, func() { aEffects++ })
	b.state.runEffects()

	if aEffects != 0 || bEffects != 1 {
		t.Fatalf("effects: a=%d b=%d", aEffects, bEffects)
	}
}

func TestRuntimeStateIsolation_contextValues(t *testing.T) {
	t.Parallel()

	ctx := NewContext("default")
	var setA func(string)
	var setB func(string)
	root := func(set *func(string)) Component {
		return func() Element {
			value, setValue := UseState("default")
			*set = setValue
			current := UseContext(ctx)
			if value != "default" && value != current {
				SetContext(ctx, value)
				current = value
			}
			return Text(current)
		}
	}

	a := NewHarness(t, root(&setA))
	defer a.Close()
	b := NewHarness(t, root(&setB))
	defer b.Close()

	setA("alpha")
	a.Render()
	setB("beta")
	b.Render()

	if a.Line(0) != "alpha" {
		t.Fatalf("a context: %q", a.Line(0))
	}
	if b.Line(0) != "beta" {
		t.Fatalf("b context: %q", b.Line(0))
	}

	a.Render()
	if a.Line(0) != "alpha" || b.Line(0) != "beta" {
		t.Fatalf("contexts after extra render: a=%q b=%q", a.Line(0), b.Line(0))
	}
}

func TestRuntimeStateIsolation_parallelHarnesses(t *testing.T) {
	names := []string{"one", "two", "three", "four"}
	for _, name := range names {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			h := NewHarness(t, focusRoot(name))
			defer h.Close()

			h.Tab()
			if !h.Contains(name + "B:focused") {
				t.Fatalf("focus after Tab: %v", h.Lines())
			}
			h.Tab()
			if !h.Contains(name + "A:focused") {
				t.Fatalf("focus after wrap: %v", h.Lines())
			}
		})
	}
}
