package gink

import (
	"reflect"
	"strings"
	"testing"
)

func TestDisplayWidth_countsDisplayColumns(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "empty", in: "", want: 0},
		{name: "ascii", in: "abc", want: 3},
		{name: "cjk", in: "世界", want: 4},
		{name: "combining", in: "e\u0301", want: 1},
		{name: "flag", in: "🇺🇸", want: 2},
		{name: "skin tone", in: "👍🏽", want: 2},
		{name: "zwj", in: "👨‍👩‍👧‍👦", want: 2},
		{name: "control", in: "\t", want: 1},
		{name: "wide ambiguous", in: "\u2e3a", want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayWidth(tt.in); got != tt.want {
				t.Errorf("displayWidth(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestClusterWidth_clampsNonPositive(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{in: -3, want: 1},
		{in: 0, want: 1},
		{in: 2, want: 2},
	}

	for _, tt := range tests {
		if got := clusterWidth(tt.in); got != tt.want {
			t.Errorf("clusterWidth(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestSplitGraphemes_multiCodepointClusters(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "combining", in: "e\u0301x", want: []string{"e\u0301", "x"}},
		{name: "skin tone", in: "👍🏽", want: []string{"👍🏽"}},
		{name: "zwj", in: "👨‍👩‍👧‍👦", want: []string{"👨‍👩‍👧‍👦"}},
		{name: "flag", in: "🇺🇸", want: []string{"🇺🇸"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitGraphemes(tt.in)
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitGraphemes(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if joined := strings.Join(got, ""); joined != tt.in {
				t.Errorf("join = %q, want %q", joined, tt.in)
			}
		})
	}
}

func TestTruncateToWidth_neverSplitsCluster(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		maxWidth int
		want     string
	}{
		{name: "zero width", in: "abc", maxWidth: 0, want: ""},
		{name: "negative width", in: "abc", maxWidth: -1, want: ""},
		{name: "ascii exact", in: "abc", maxWidth: 3, want: "abc"},
		{name: "cjk boundary", in: "世界X", maxWidth: 4, want: "世界"},
		{name: "combining boundary", in: "e\u0301x", maxWidth: 1, want: "e\u0301"},
		{name: "skin tone boundary", in: "👍🏽X", maxWidth: 2, want: "👍🏽"},
		{name: "zwj boundary", in: "👨‍👩‍👧‍👦X", maxWidth: 2, want: "👨‍👩‍👧‍👦"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateToWidth(tt.in, tt.maxWidth); got != tt.want {
				t.Errorf("truncateToWidth(%q, %d) = %q, want %q", tt.in, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestTruncateWithEllipsis_widthAndBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		maxWidth int
		want     string
	}{
		{name: "zero width", in: "abc", maxWidth: 0, want: ""},
		{name: "one column", in: "abc", maxWidth: 1, want: "…"},
		{name: "exact fit", in: "abc", maxWidth: 3, want: "abc"},
		{name: "wide ellipsis", in: "世界世界", maxWidth: 5, want: "世界…"},
		{name: "skin tone ellipsis", in: "👍🏽XY", maxWidth: 3, want: "👍🏽…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateWithEllipsis(tt.in, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncateWithEllipsis(%q, %d) = %q, want %q", tt.in, tt.maxWidth, got, tt.want)
			}
			if w := displayWidth(got); w > tt.maxWidth {
				t.Errorf("displayWidth(%q) = %d, exceeds %d", got, w, tt.maxWidth)
			}
		})
	}
}

func TestPadToWidth_padsAndTruncatesByColumns(t *testing.T) {
	if got := padToWidth("Hi", 5); got != "Hi   " {
		t.Errorf("pad = %q, want %q", got, "Hi   ")
	}
	if got := padToWidth("界X", 2); got != "界" {
		t.Errorf("truncate = %q, want %q", got, "界")
	}
}

func TestByteOffsetAtGrapheme(t *testing.T) {
	s := "e\u0301x"
	tests := []struct {
		index int
		want  int
	}{
		{index: 0, want: 0},
		{index: 1, want: 3},
		{index: 2, want: 4},
		{index: 99, want: 4},
	}

	for _, tt := range tests {
		if got := byteOffsetAtGrapheme(s, tt.index); got != tt.want {
			t.Errorf("byteOffsetAtGrapheme(%q, %d) = %d, want %d", s, tt.index, got, tt.want)
		}
	}
}

func TestSpliceGraphemes(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		start       int
		end         int
		replacement string
		want        string
	}{
		{name: "delete grapheme", in: "e\u0301x", start: 0, end: 1, replacement: "", want: "x"},
		{name: "insert grapheme", in: "e\u0301x", start: 1, end: 1, replacement: "!", want: "e\u0301!x"},
		{name: "clamp range", in: "abc", start: -5, end: 99, replacement: "Z", want: "Z"},
		{name: "reversed range", in: "abc", start: 2, end: 1, replacement: "Z", want: "abZc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := spliceGraphemes(tt.in, tt.start, tt.end, tt.replacement); got != tt.want {
				t.Errorf("spliceGraphemes(%q, %d, %d, %q) = %q, want %q", tt.in, tt.start, tt.end, tt.replacement, got, tt.want)
			}
		})
	}
}
