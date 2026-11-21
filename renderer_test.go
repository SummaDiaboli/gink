package gink

import (
	"reflect"
	"testing"
)

func TestRenderer_emitsCombiningRunes(t *testing.T) {
	h := NewHarnessSize(t, func() Element { return Text("e\u0301") }, 4, 1)
	defer h.Close()

	main, comb, _, width := h.screen.GetContent(0, 0)
	if main != 'e' {
		t.Errorf("main rune = %q, want %q", main, 'e')
	}
	if !reflect.DeepEqual(comb, []rune{'\u0301'}) {
		t.Errorf("combining runes = %q, want %q", comb, []rune{'\u0301'})
	}
	if width != 1 {
		t.Errorf("width = %d, want 1", width)
	}
}

func TestRenderer_emitsWideRuneAndClearsContinuation(t *testing.T) {
	h := NewHarnessSize(t, func() Element { return Text("界") }, 4, 1)
	defer h.Close()

	main, comb, _, width := h.screen.GetContent(0, 0)
	if main != '界' {
		t.Errorf("main rune = %q, want %q", main, '界')
	}
	if len(comb) != 0 {
		t.Errorf("combining runes = %q, want none", comb)
	}
	if width != 2 {
		t.Errorf("width = %d, want 2", width)
	}

	nextMain, nextComb, _, nextWidth := h.screen.GetContent(1, 0)
	if nextMain != ' ' {
		t.Errorf("continuation main rune = %q, want space", nextMain)
	}
	if len(nextComb) != 0 {
		t.Errorf("continuation combining runes = %q, want none", nextComb)
	}
	if nextWidth != 1 {
		t.Errorf("continuation width = %d, want 1", nextWidth)
	}
}

func TestRenderer_emitsZwjEmojiCombiningRunes(t *testing.T) {
	cluster := "👨‍👩‍👧‍👦"
	h := NewHarnessSize(t, func() Element { return Text(cluster) }, 4, 1)
	defer h.Close()

	want := []rune(cluster)
	main, comb, _, width := h.screen.GetContent(0, 0)
	if main != want[0] {
		t.Errorf("main rune = %q, want %q", main, want[0])
	}
	if !reflect.DeepEqual(comb, want[1:]) {
		t.Errorf("combining runes = %q, want %q", comb, want[1:])
	}
	if width != 2 {
		t.Errorf("width = %d, want 2", width)
	}
}

func TestRenderer_wideToNarrowTransitionClearsStaleCell(t *testing.T) {
	var setVal func(string)
	h := NewHarnessSize(t, func() Element {
		val, setter := UseState("界")
		setVal = setter
		return Text(val)
	}, 4, 1)
	defer h.Close()

	setVal("X")
	h.Render()

	main, _, _, width := h.screen.GetContent(0, 0)
	if main != 'X' {
		t.Errorf("main rune = %q, want %q", main, 'X')
	}
	if width != 1 {
		t.Errorf("width = %d, want 1", width)
	}
	nextMain, _, _, nextWidth := h.screen.GetContent(1, 0)
	if nextMain != ' ' {
		t.Errorf("stale continuation main rune = %q, want space", nextMain)
	}
	if nextWidth != 1 {
		t.Errorf("stale continuation width = %d, want 1", nextWidth)
	}
}
