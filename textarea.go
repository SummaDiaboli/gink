package gink

import "strings"

// NewTextArea returns a focusable multi-line text input component.
// "Controlled" means the parent owns the value string and receives updates via
// onChange. height is the number of visible rows; content that exceeds this
// scrolls as the cursor moves.
//
// Cursor movement: Left/Right within a line, Up/Down between lines, Home/End.
// Enter splits the line at the cursor. Backspace deletes the grapheme before
// the cursor, merging lines when at column 0.
//
// An optional style controls the cursor highlight (defaults to reverse video):
//
//	text, setText := gink.UseState("")
//	gink.C(gink.NewTextArea(text, setText, 5))
func NewTextArea(value string, onChange func(string), height int, styles ...Style) func() Element {
	explicitStyle, hasExplicitStyle := optionalStyle(styles)

	return func() Element {
		cursorStyle := resolveStyle(explicitStyle, hasExplicitStyle, UseTheme().Cursor)

		curLine, setCurLine := UseState(0)
		curCol, setCurCol := UseState(0)
		offset, setOffset := UseState(0)
		isFocused := UseFocus()
		clipRead, _ := UseClipboard()

		lines := strings.Split(value, "\n")
		nLines := len(lines)

		// Clamp cursor to valid position when value changes externally.
		if curLine >= nLines {
			curLine = nLines - 1
			setCurLine(curLine)
		}
		curClusters := splitGraphemes(lines[curLine])
		if curCol > len(curClusters) {
			curCol = len(curClusters)
			setCurCol(curCol)
		}

		// Scroll viewport to keep cursor visible.
		if curLine < offset {
			offset = curLine
			setOffset(offset)
		} else if height > 0 && curLine >= offset+height {
			offset = curLine - height + 1
			setOffset(offset)
		}

		UseInput(func(ev KeyEvent) {
			if !isFocused {
				return
			}
			ls := strings.Split(value, "\n")
			nl := len(ls)
			cl := curLine
			if cl >= nl {
				cl = nl - 1
			}
			cc := curCol
			clusters := splitGraphemes(ls[cl])
			if cc > len(clusters) {
				cc = len(clusters)
			}

			switch ev.Key {
			case KeyLeft:
				if cc > 0 {
					setCurCol(cc - 1)
				} else if cl > 0 {
					setCurLine(cl - 1)
					setCurCol(len(splitGraphemes(ls[cl-1])))
				}

			case KeyRight:
				if cc < len(clusters) {
					setCurCol(cc + 1)
				} else if cl < nl-1 {
					setCurLine(cl + 1)
					setCurCol(0)
				}

			case KeyUp:
				if cl > 0 {
					setCurLine(cl - 1)
					prevLen := len(splitGraphemes(ls[cl-1]))
					if cc > prevLen {
						setCurCol(prevLen)
					} else {
						setCurCol(cc)
					}
				}

			case KeyDown:
				if cl < nl-1 {
					setCurLine(cl + 1)
					nextLen := len(splitGraphemes(ls[cl+1]))
					if cc > nextLen {
						setCurCol(nextLen)
					} else {
						setCurCol(cc)
					}
				}

			case KeyHome:
				setCurCol(0)

			case KeyEnd:
				setCurCol(len(clusters))

			case KeyEnter:
				line := ls[cl]
				off := byteOffsetAtGrapheme(line, cc)
				before := line[:off]
				after := line[off:]
				newLines := make([]string, 0, nl+1)
				newLines = append(newLines, ls[:cl]...)
				newLines = append(newLines, before, after)
				newLines = append(newLines, ls[cl+1:]...)
				onChange(strings.Join(newLines, "\n"))
				setCurLine(cl + 1)
				setCurCol(0)

			case KeyBackspace, KeyBackspace2:
				if cc > 0 {
					newLine := spliceGraphemes(ls[cl], cc-1, cc, "")
					newLines := make([]string, nl)
					copy(newLines, ls)
					newLines[cl] = newLine
					onChange(strings.Join(newLines, "\n"))
					setCurCol(cc - 1)
				} else if cl > 0 {
					prevLine := ls[cl-1]
					prevLen := len(splitGraphemes(prevLine))
					merged := prevLine + ls[cl]
					newLines := make([]string, 0, nl-1)
					newLines = append(newLines, ls[:cl-1]...)
					newLines = append(newLines, merged)
					newLines = append(newLines, ls[cl+1:]...)
					onChange(strings.Join(newLines, "\n"))
					setCurLine(cl - 1)
					setCurCol(prevLen)
				}

			case KeyCtrlV:
				text := normalizeNewlines(clipRead(), "\n")
				pastedLines := strings.Split(text, "\n")
				line := ls[cl]
				off := byteOffsetAtGrapheme(line, cc)
				firstPart := line[:off] + pastedLines[0]
				lastPart := pastedLines[len(pastedLines)-1] + line[off:]
				newLines := make([]string, 0, nl+len(pastedLines)-1)
				newLines = append(newLines, ls[:cl]...)
				if len(pastedLines) == 1 {
					newLines = append(newLines, firstPart+line[off:])
				} else {
					newLines = append(newLines, firstPart)
					newLines = append(newLines, pastedLines[1:len(pastedLines)-1]...)
					newLines = append(newLines, lastPart)
				}
				newLines = append(newLines, ls[cl+1:]...)
				onChange(strings.Join(newLines, "\n"))
				newCurLine := cl + len(pastedLines) - 1
				newCurCol := len(splitGraphemes(pastedLines[len(pastedLines)-1]))
				if len(pastedLines) == 1 {
					newCurCol = cc + len(splitGraphemes(pastedLines[0]))
				}
				setCurLine(newCurLine)
				setCurCol(newCurCol)

			case KeyRune:
				newLine := spliceGraphemes(ls[cl], cc, cc, string(ev.Rune))
				newLines := make([]string, nl)
				copy(newLines, ls)
				newLines[cl] = newLine
				onChange(strings.Join(newLines, "\n"))
				setCurCol(cc + 1)
			}
		})

		// Build exactly height row elements, padding with blank lines when needed.
		elems := make([]Element, height)
		for i := range elems {
			src := offset + i
			if src >= nLines {
				elems[i] = Text("")
				continue
			}
			clusters := splitGraphemes(lines[src])
			if !isFocused || src != curLine {
				elems[i] = Text(strings.Join(clusters, ""))
				continue
			}
			before := ""
			if curCol > 0 {
				before = strings.Join(clusters[:min(curCol, len(clusters))], "")
			}
			curChar := " "
			if curCol >= 0 && curCol < len(clusters) {
				curChar = clusters[curCol]
			}
			after := ""
			if curCol+1 < len(clusters) {
				after = strings.Join(clusters[curCol+1:], "")
			}
			elems[i] = Row(Text(before), Text(curChar, cursorStyle), Text(after))
		}

		return Box(elems...)
	}
}
