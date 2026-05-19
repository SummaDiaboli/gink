package gink_test

import (
	"testing"

	"github.com/SummaDiaboli/gink"
)

// TestTextWrapped_shortStringFitsOnOneLine verifies that text shorter than
// maxWidth is not split.
func TestTextWrapped_shortStringFitsOnOneLine(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.TextWrapped("hello", 20)
	})
	defer h.Close()

	if h.Line(0) != "hello" {
		t.Errorf("expected 'hello' on line 0, got %q", h.Line(0))
	}
	if h.Line(1) != "" {
		t.Errorf("expected empty line 1, got %q", h.Line(1))
	}
}

// TestTextWrapped_wrapsAtWordBoundary verifies that a string longer than
// maxWidth is split between words.
func TestTextWrapped_wrapsAtWordBoundary(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.TextWrapped("hello world", 8)
	})
	defer h.Close()

	if h.Line(0) != "hello" {
		t.Errorf("line 0: expected 'hello', got %q", h.Line(0))
	}
	if h.Line(1) != "world" {
		t.Errorf("line 1: expected 'world', got %q", h.Line(1))
	}
}

// TestTextWrapped_packsMultipleWordsPerLine verifies that words are packed
// greedily — as many as fit on one line before wrapping.
func TestTextWrapped_packsMultipleWordsPerLine(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.TextWrapped("one two three four", 11)
	})
	defer h.Close()

	// "one two" = 7, "three" = 5, "one two thr" = 11 — fits exactly
	// "one two thr" = 11, next "ee" no...
	// "one two" = 7 ≤ 11, "one two three" = 13 > 11 → line 0 = "one two"
	// "three four" = 10 ≤ 11 → line 1 = "three four"
	if h.Line(0) != "one two" {
		t.Errorf("line 0: expected 'one two', got %q", h.Line(0))
	}
	if h.Line(1) != "three four" {
		t.Errorf("line 1: expected 'three four', got %q", h.Line(1))
	}
}

// TestTextWrapped_hardBreaksWordLongerThanMaxWidth verifies that a single word
// exceeding maxWidth is split at the width boundary.
func TestTextWrapped_hardBreaksWordLongerThanMaxWidth(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.TextWrapped("abcdefgh", 5)
	})
	defer h.Close()

	if h.Line(0) != "abcde" {
		t.Errorf("line 0: expected 'abcde', got %q", h.Line(0))
	}
	if h.Line(1) != "fgh" {
		t.Errorf("line 1: expected 'fgh', got %q", h.Line(1))
	}
}

// TestTextWrapped_respectsExistingNewlines verifies that \n in the input
// always starts a new line regardless of width.
func TestTextWrapped_respectsExistingNewlines(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.TextWrapped("line one\nline two", 40)
	})
	defer h.Close()

	if h.Line(0) != "line one" {
		t.Errorf("line 0: expected 'line one', got %q", h.Line(0))
	}
	if h.Line(1) != "line two" {
		t.Errorf("line 1: expected 'line two', got %q", h.Line(1))
	}
}

// TestTextWrapped_emptyString verifies that an empty string produces a single
// blank line rather than panicking or producing no output.
func TestTextWrapped_emptyString(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.TextWrapped("", 20)
	})
	defer h.Close()

	if h.Line(0) != "" {
		t.Errorf("expected empty line 0, got %q", h.Line(0))
	}
}

// TestTextWrapped_correctHeightInLayout verifies that TextWrapped reports the
// right height to its parent so sibling elements are positioned correctly.
func TestTextWrapped_correctHeightInLayout(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		return gink.Box(
			gink.TextWrapped("hello world foo bar", 10),
			gink.Text("after"),
		)
	})
	defer h.Close()

	// "hello" (5) + " world" = 11 > 10 → line 0 = "hello world"? No.
	// "hello" = 5 ≤ 10, "hello world" = 11 > 10 → line 0 = "hello"
	// "world foo" = 9 ≤ 10, "world foo bar" = 13 > 10 → line 1 = "world foo"
	// line 2 = "bar"
	// "after" should appear at row 3
	if h.Line(3) != "after" {
		t.Errorf("expected 'after' at row 3, got %q (wrapped text took wrong height)", h.Line(3))
	}
}

// TestTextWrapped_unicodeCharacters verifies that wrapping counts runes, not
// bytes, so multi-byte characters are handled correctly.
func TestTextWrapped_unicodeCharacters(t *testing.T) {
	h := gink.NewHarness(t, func() gink.Element {
		// Each emoji is 1 rune; "😀😀😀" = 3 runes, maxWidth = 2
		return gink.TextWrapped("😀😀 😀😀", 4)
	})
	defer h.Close()

	if h.Line(0) != "😀😀" {
		t.Errorf("line 0: expected '😀😀', got %q", h.Line(0))
	}
	if h.Line(1) != "😀😀" {
		t.Errorf("line 1: expected '😀😀', got %q", h.Line(1))
	}
}
