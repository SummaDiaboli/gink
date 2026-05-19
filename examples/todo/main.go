// Todo list — demonstrates stateful child components, dynamic lists, and
// keyboard navigation without explicit buttons per item.
package main

import (
	"fmt"
	"log"

	"github.com/salim/gink"
)

// Todo holds a single task.
type Todo struct {
	text string
	done bool
}

var (
	titleStyle    = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	subtitleStyle = gink.NewStyle().Foreground(gink.ColorWhite)
	doneStyle     = gink.NewStyle().Foreground(gink.ColorWhite)
	pendingStyle  = gink.NewStyle().Foreground(gink.ColorBrightWhite)
	selectedStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightYellow)
	hintStyle     = gink.NewStyle().Foreground(gink.ColorWhite)
	emptyStyle    = gink.NewStyle().Foreground(gink.ColorWhite)
)

// TodoListComponent renders a navigable list of todos. It owns the
// selectedIdx state so navigation survives parent re-renders.
func TodoListComponent(todos []Todo, onToggle func(int), onDelete func(int)) func() gink.Element {
	return func() gink.Element {
		sel, setSel := gink.UseState(0)
		isFocused := gink.UseFocus()

		// Clamp selection when items are deleted.
		if len(todos) > 0 && sel >= len(todos) {
			sel = len(todos) - 1
		}

		gink.UseInput(func(ev gink.KeyEvent) {
			if !isFocused || len(todos) == 0 {
				return
			}
			switch ev.Key {
			case gink.KeyUp:
				if sel > 0 {
					setSel(sel - 1)
				}
			case gink.KeyDown:
				if sel < len(todos)-1 {
					setSel(sel + 1)
				}
			case gink.KeyEnter:
				onToggle(sel)
			}
			if ev.Rune == 'd' {
				onDelete(sel)
			}
		})

		if len(todos) == 0 {
			return gink.Text("  (no tasks — add one below)", emptyStyle)
		}

		rows := make([]gink.Element, len(todos))
		for i, todo := range todos {
			cursor := "  "
			check := "[ ]"
			style := pendingStyle
			if todo.done {
				check = "[x]"
				style = doneStyle
			}
			if i == sel && isFocused {
				cursor = "▶ "
				style = selectedStyle
			}
			rows[i] = gink.Text(fmt.Sprintf("%s%s %s", cursor, check, todo.text), style)
		}
		return gink.Box(rows...)
	}
}

func doneCount(todos []Todo) int {
	n := 0
	for _, t := range todos {
		if t.done {
			n++
		}
	}
	return n
}

func App() gink.Element {
	todos, setTodos := gink.UseState([]Todo{
		{text: "Buy groceries"},
		{text: "Walk the dog"},
		{text: "Read a book"},
	})
	inputText, setInputText := gink.UseState("")

	addTodo := func() {
		if inputText == "" {
			return
		}
		newList := make([]Todo, len(todos)+1)
		copy(newList, todos)
		newList[len(todos)] = Todo{text: inputText}
		setTodos(newList)
		setInputText("")
	}

	toggle := func(i int) {
		next := make([]Todo, len(todos))
		copy(next, todos)
		next[i].done = !next[i].done
		setTodos(next)
	}

	deleteTodo := func(i int) {
		if len(todos) == 0 {
			return
		}
		next := make([]Todo, 0, len(todos)-1)
		next = append(next, todos[:i]...)
		next = append(next, todos[i+1:]...)
		setTodos(next)
	}

	done := doneCount(todos)
	subtitle := fmt.Sprintf("%d task(s) · %d done · %d remaining", len(todos), done, len(todos)-done)

	return gink.BoxWithGap(1,
		gink.Text("Todo List", titleStyle),
		gink.Text(subtitle, subtitleStyle),

		gink.C(TodoListComponent(todos, toggle, deleteTodo)),

		gink.Row(
			gink.C(gink.NewInput(inputText, setInputText)),
			gink.Text("  "),
			gink.C(gink.NewButton("Add", addTodo)),
		),

		gink.Text("↑↓ navigate  ·  Enter toggle  ·  d delete  ·  Tab switch focus  ·  Esc quit", hintStyle),
	)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}
