package gink

import (
	"fmt"
	"testing"
	"time"
)

// ── UseContext / SetContext ───────────────────────────────────────────────────

func TestUseContext_returnsDefaultValue(t *testing.T) {
	ctx := NewContext("default")

	h := NewHarness(t, func() Element {
		val := UseContext(ctx)
		return Text(val)
	})
	defer h.Close()

	if h.Line(0) != "default" {
		t.Errorf("got %q, want default", h.Line(0))
	}
}

func TestUseContext_triggersRerenderOnSet(t *testing.T) {
	ctx := NewContext("initial")

	h := NewHarness(t, func() Element {
		val := UseContext(ctx)
		return Text(val)
	})
	defer h.Close()

	SetContext(ctx, "updated")
	h.Render()

	if h.Line(0) != "updated" {
		t.Errorf("after SetContext: got %q, want updated", h.Line(0))
	}
}

func TestUseContext_multipleSetCallsUpdateCorrectly(t *testing.T) {
	ctx := NewContext(0)

	h := NewHarness(t, func() Element {
		n := UseContext(ctx)
		return Text(fmt.Sprintf("%d", n))
	})
	defer h.Close()

	for i := 1; i <= 3; i++ {
		SetContext(ctx, i)
		h.Render()
		if h.Line(0) != fmt.Sprintf("%d", i) {
			t.Errorf("after SetContext(%d): got %q, want %d", i, h.Line(0), i)
		}
	}
}

func TestUseContext_allWatchingComponentsUpdate(t *testing.T) {
	ctx := NewContext("x")

	makeComp := func(prefix string) func() Element {
		return func() Element {
			val := UseContext(ctx)
			return Text(prefix + ":" + val)
		}
	}

	h := NewHarness(t, func() Element {
		return Box(C(makeComp("A")), C(makeComp("B")))
	})
	defer h.Close()

	SetContext(ctx, "y")
	h.Render()

	if !h.Contains("A:y") {
		t.Errorf("component A did not update; lines: %v", h.Lines()[:2])
	}
	if !h.Contains("B:y") {
		t.Errorf("component B did not update; lines: %v", h.Lines()[:2])
	}
}

func TestUseContext_independentContextsDoNotInterfere(t *testing.T) {
	ctxA := NewContext("alpha")
	ctxB := NewContext("beta")

	h := NewHarness(t, func() Element {
		a := UseContext(ctxA)
		b := UseContext(ctxB)
		return Box(Text(a), Text(b))
	})
	defer h.Close()

	SetContext(ctxA, "ALPHA")
	h.Render()

	if !h.Contains("ALPHA") {
		t.Error("ctxA value should have updated to ALPHA")
	}
	if !h.Contains("beta") {
		t.Error("ctxB value should remain beta after only ctxA was set")
	}
}

func TestUseContext_setFromGoroutineTriggersRerender(t *testing.T) {
	ctx := NewContext("initial")

	h := NewHarness(t, func() Element {
		val := UseContext(ctx)
		return Text(val)
	})
	defer h.Close()

	go SetContext(ctx, "from-goroutine")

	awaitContains(t, h, "from-goroutine", 500*time.Millisecond)
}

func TestUseContext_worksAlongsideUseState(t *testing.T) {
	ctx := NewContext("global")

	h := NewHarness(t, func() Element {
		local, setLocal := UseState("local")
		global := UseContext(ctx)
		_ = setLocal
		return Box(Text("local:"+local), Text("global:"+global))
	})
	defer h.Close()

	SetContext(ctx, "GLOBAL")
	h.Render()

	if !h.Contains("local:local") {
		t.Error("local state should be unchanged")
	}
	if !h.Contains("global:GLOBAL") {
		t.Error("global context should have updated")
	}
}
