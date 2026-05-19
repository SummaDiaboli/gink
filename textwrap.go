package gink

import "strings"

// TextWrapped renders s as one or more lines, wrapping at word boundaries so
// no line exceeds maxWidth runes. Existing newlines in s are always honoured.
// Words longer than maxWidth are hard-broken at the width boundary.
//
// When maxWidth is zero or negative, s is returned as a single unsplit line.
//
// An optional style applies to every line:
//
//	gink.TextWrapped(longDescription, 60, gink.NewStyle().Foreground(gink.ColorWhite))
func TextWrapped(s string, maxWidth int, styles ...Style) Element {
	var style Style
	if len(styles) > 0 {
		style = styles[0]
	}

	if maxWidth <= 0 {
		return Text(s, style)
	}

	lines := wrapText(s, maxWidth)
	elems := make([]Element, len(lines))
	for i, l := range lines {
		elems[i] = Text(l, style)
	}
	if len(elems) == 1 {
		return elems[0]
	}
	return Box(elems...)
}

// wrapText splits s on newlines then word-wraps each paragraph to maxWidth.
func wrapText(s string, maxWidth int) []string {
	paragraphs := strings.Split(s, "\n")
	var result []string
	for _, p := range paragraphs {
		result = append(result, wrapParagraph(p, maxWidth)...)
	}
	return result
}

// wrapParagraph greedily packs words onto lines of at most maxWidth runes.
// Words longer than maxWidth are hard-broken at the boundary.
func wrapParagraph(s string, maxWidth int) []string {
	if s == "" {
		return []string{""}
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var cur []rune

	flush := func() {
		lines = append(lines, string(cur))
		cur = nil
	}

	for _, word := range words {
		runes := []rune(word)

		// Hard-break any segment of the word that exceeds maxWidth.
		for len(runes) > maxWidth {
			if len(cur) > 0 {
				flush()
			}
			lines = append(lines, string(runes[:maxWidth]))
			runes = runes[maxWidth:]
		}
		if len(runes) == 0 {
			continue
		}

		// Try to append to the current line.
		if len(cur) == 0 {
			cur = runes
		} else if len(cur)+1+len(runes) <= maxWidth {
			cur = append(cur, ' ')
			cur = append(cur, runes...)
		} else {
			flush()
			cur = runes
		}
	}

	if len(cur) > 0 {
		flush()
	}

	return lines
}
