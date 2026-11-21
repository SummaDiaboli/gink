package gink

import "strings"

// TextWrapped renders s as one or more lines, wrapping at word boundaries so
// no line exceeds maxWidth display columns, except for an individual grapheme
// wider than maxWidth. Existing newlines in s are always honoured. Words
// longer than maxWidth are hard-broken at grapheme boundaries.
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

// wrapParagraph greedily packs words onto lines of at most maxWidth display
// columns. Words longer than maxWidth are hard-broken at grapheme boundaries.
func wrapParagraph(s string, maxWidth int) []string {
	if s == "" {
		return []string{""}
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var cur strings.Builder
	curWidth := 0

	flush := func() {
		lines = append(lines, cur.String())
		cur.Reset()
		curWidth = 0
	}

	for _, word := range words {
		wordWidth := displayWidth(word)
		if wordWidth > maxWidth {
			if curWidth > 0 {
				flush()
			}
			lines = append(lines, hardBreakWord(word, maxWidth)...)
			continue
		}

		if curWidth == 0 {
			cur.WriteString(word)
			curWidth = wordWidth
		} else if curWidth+1+wordWidth <= maxWidth {
			cur.WriteByte(' ')
			cur.WriteString(word)
			curWidth += 1 + wordWidth
		} else {
			flush()
			cur.WriteString(word)
			curWidth = wordWidth
		}
	}

	if curWidth > 0 {
		flush()
	}

	return lines
}

func hardBreakWord(word string, maxWidth int) []string {
	var lines []string
	var cur strings.Builder
	curWidth := 0

	for _, cluster := range splitGraphemes(word) {
		w := displayWidth(cluster)
		if w < 1 {
			w = 1
		}
		if curWidth > 0 && curWidth+w > maxWidth {
			lines = append(lines, cur.String())
			cur.Reset()
			curWidth = 0
		}
		cur.WriteString(cluster)
		curWidth += w
		if curWidth >= maxWidth {
			lines = append(lines, cur.String())
			cur.Reset()
			curWidth = 0
		}
	}

	if curWidth > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}
