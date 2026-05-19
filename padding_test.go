package gink

import "testing"

func TestPadding_topOffsetsPushesContentDown(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Padding(Pad{Top: 2}, Text("Hello"))
	})
	defer h.Close()

	if h.Line(0) != "" {
		t.Errorf("line 0: got %q, want empty", h.Line(0))
	}
	if h.Line(1) != "" {
		t.Errorf("line 1: got %q, want empty", h.Line(1))
	}
	if h.Line(2) != "Hello" {
		t.Errorf("line 2: got %q, want %q", h.Line(2), "Hello")
	}
}

func TestPadding_leftOffsetsPushesContentRight(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Padding(Pad{Left: 3}, Text("Hi"))
	})
	defer h.Close()

	if h.Line(0) != "   Hi" {
		t.Errorf("got %q, want %q", h.Line(0), "   Hi")
	}
}

func TestPaddingAll_addsEqualSpacingOnAllSides(t *testing.T) {
	h := NewHarness(t, func() Element {
		return PaddingAll(1, Text("X"))
	})
	defer h.Close()

	if h.Line(0) != "" {
		t.Errorf("line 0 (top pad): got %q, want empty", h.Line(0))
	}
	if h.Line(1) != " X" {
		t.Errorf("line 1: got %q, want %q", h.Line(1), " X")
	}
	if h.Line(2) != "" {
		t.Errorf("line 2 (bottom pad): got %q, want empty", h.Line(2))
	}
}

func TestPaddingXY_appliesHorizontalAndVertical(t *testing.T) {
	h := NewHarness(t, func() Element {
		return PaddingXY(2, 1, Text("Hi"))
	})
	defer h.Close()

	if h.Line(0) != "" {
		t.Errorf("line 0 (top pad): got %q, want empty", h.Line(0))
	}
	if h.Line(1) != "  Hi" {
		t.Errorf("line 1: got %q, want %q", h.Line(1), "  Hi")
	}
}

func TestPadding_zeroPadIsTransparent(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Padding(Pad{}, Text("Hello"))
	})
	defer h.Close()

	if h.Line(0) != "Hello" {
		t.Errorf("got %q, want %q", h.Line(0), "Hello")
	}
}

func TestPadding_worksInsideBox(t *testing.T) {
	h := NewHarness(t, func() Element {
		return Box(
			Text("Above"),
			PaddingAll(1, Text("Padded")),
			Text("Below"),
		)
	})
	defer h.Close()

	if h.Line(0) != "Above" {
		t.Errorf("line 0: got %q, want %q", h.Line(0), "Above")
	}
	if h.Line(1) != "" {
		t.Errorf("line 1 (top pad): got %q, want empty", h.Line(1))
	}
	if h.Line(2) != " Padded" {
		t.Errorf("line 2: got %q, want %q", h.Line(2), " Padded")
	}
	if h.Line(3) != "" {
		t.Errorf("line 3 (bottom pad): got %q, want empty", h.Line(3))
	}
	if h.Line(4) != "Below" {
		t.Errorf("line 4: got %q, want %q", h.Line(4), "Below")
	}
}
