package gink

import "strings"

// NewTextArea returns a focusable multi-line text input component.
// "Controlled" means the parent owns the value string and receives updates via
// onChange. height is the number of visible rows; content that exceeds this
// scrolls as the cursor moves.
//
// Cursor movement: Left/Right within a line, Up/Down between lines, Home/End.
// Enter splits the line at the cursor. Backspace deletes the character before
// the cursor, merging lines when at column 0.
//
// An optional style controls the cursor highlight (defaults to reverse video):
//
//	text, setText := gink.UseState("")
//	gink.C(gink.NewTextArea(text, setText, 5))
func NewTextArea(value string, onChange func(string), height int, styles ...Style) func() Element {
	hasExplicitStyle := len(styles) > 0
	explicitStyle := Style{}
	if hasExplicitStyle {
		explicitStyle = styles[0]
	}

	return func() Element {
		cursorStyle := explicitStyle
		if !hasExplicitStyle {
			cursorStyle = UseTheme().Cursor
		}

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
		lineRunes := []rune(lines[curLine])
		if curCol > len(lineRunes) {
			curCol = len(lineRunes)
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
			lr := []rune(ls[cl])
			if cc > len(lr) {
				cc = len(lr)
			}

			switch ev.Key {
			case KeyLeft:
				if cc > 0 {
					setCurCol(cc - 1)
				} else if cl > 0 {
					setCurLine(cl - 1)
					setCurCol(len([]rune(ls[cl-1])))
				}

			case KeyRight:
				if cc < len(lr) {
					setCurCol(cc + 1)
				} else if cl < nl-1 {
					setCurLine(cl + 1)
					setCurCol(0)
				}

			case KeyUp:
				if cl > 0 {
					setCurLine(cl - 1)
					prevLen := len([]rune(ls[cl-1]))
					if cc > prevLen {
						setCurCol(prevLen)
					} else {
						setCurCol(cc)
					}
				}

			case KeyDown:
				if cl < nl-1 {
					setCurLine(cl + 1)
					nextLen := len([]rune(ls[cl+1]))
					if cc > nextLen {
						setCurCol(nextLen)
					} else {
						setCurCol(cc)
					}
				}

			case KeyHome:
				setCurCol(0)

			case KeyEnd:
				setCurCol(len(lr))

			case KeyEnter:
				before := string(lr[:cc])
				after := string(lr[cc:])
				newLines := make([]string, 0, nl+1)
				newLines = append(newLines, ls[:cl]...)
				newLines = append(newLines, before, after)
				newLines = append(newLines, ls[cl+1:]...)
				onChange(strings.Join(newLines, "\n"))
				setCurLine(cl + 1)
				setCurCol(0)

			case KeyBackspace, KeyBackspace2:
				if cc > 0 {
					newLine := string(lr[:cc-1]) + string(lr[cc:])
					newLines := make([]string, nl)
					copy(newLines, ls)
					newLines[cl] = newLine
					onChange(strings.Join(newLines, "\n"))
					setCurCol(cc - 1)
				} else if cl > 0 {
					prevLine := ls[cl-1]
					prevLen := len([]rune(prevLine))
					merged := prevLine + string(lr)
					newLines := make([]string, 0, nl-1)
					newLines = append(newLines, ls[:cl-1]...)
					newLines = append(newLines, merged)
					newLines = append(newLines, ls[cl+1:]...)
					onChange(strings.Join(newLines, "\n"))
					setCurLine(cl - 1)
					setCurCol(prevLen)
				}

			case KeyCtrlV:
				text := clipRead()
				text = strings.ReplaceAll(text, "\r\n", "\n")
				text = strings.ReplaceAll(text, "\r", "\n")
				pastedLines := strings.Split(text, "\n")
				// Merge first pasted line into current line at cursor.
				firstPart := string(lr[:cc]) + pastedLines[0]
				lastPart := pastedLines[len(pastedLines)-1] + string(lr[cc:])
				newLines := make([]string, 0, nl+len(pastedLines)-1)
				newLines = append(newLines, ls[:cl]...)
				if len(pastedLines) == 1 {
					newLines = append(newLines, firstPart+string(lr[cc:]))
				} else {
					newLines = append(newLines, firstPart)
					newLines = append(newLines, pastedLines[1:len(pastedLines)-1]...)
					newLines = append(newLines, lastPart)
				}
				newLines = append(newLines, ls[cl+1:]...)
				onChange(strings.Join(newLines, "\n"))
				// Move cursor to end of pasted content.
				newCurLine := cl + len(pastedLines) - 1
				newCurCol := len([]rune(pastedLines[len(pastedLines)-1]))
				if len(pastedLines) == 1 {
					newCurCol = cc + len([]rune(pastedLines[0]))
				}
				setCurLine(newCurLine)
				setCurCol(newCurCol)

			case KeyRune:
				newLine := string(lr[:cc]) + string(ev.Rune) + string(lr[cc:])
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
			lr := []rune(lines[src])
			if !isFocused || src != curLine {
				elems[i] = Text(string(lr))
				continue
			}
			// Render cursor on the active line.
			before := string(lr[:curCol])
			var curChar string
			if curCol < len(lr) {
				curChar = string(lr[curCol])
			} else {
				curChar = " "
			}
			after := ""
			if curCol < len(lr) {
				after = string(lr[curCol+1:])
			}
			elems[i] = Row(Text(before), Text(curChar, cursorStyle), Text(after))
		}

		return Box(elems...)
	}
}
