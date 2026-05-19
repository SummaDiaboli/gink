package main

import (
	"testing"

	"github.com/salim/gink"
	"github.com/salim/gink/ginktest"
)

// Focus order in App: TodoList(0) · Input(1) · Add(2)
// TodoList is focused by default; arrow keys and Enter work immediately.

func TestTodo_initialRender(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	for _, want := range []string{"Buy groceries", "Walk the dog", "Read a book"} {
		ginktest.AssertContains(t, h, want)
	}
}

func TestTodo_toggleFirstTodo(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Enter() // toggle the selected (first) todo

	ginktest.AssertContains(t, h, "[x]")
}

func TestTodo_navigateDownAndToggle(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendKey(gink.KeyDown) // select second todo
	h.Enter()

	ginktest.AssertContains(t, h, "[x] Walk the dog")
	ginktest.AssertNotContains(t, h, "[x] Buy groceries")
}

func TestTodo_deleteFirstTodo(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('d')

	ginktest.AssertNotContains(t, h, "Buy groceries")
	ginktest.AssertContains(t, h, "Walk the dog")
}

func TestTodo_addTodo(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab() // → Input
	h.SendRune('G')
	h.SendRune('o')
	h.Tab()   // → Add
	h.Enter()

	ginktest.AssertContains(t, h, "Go")
	ginktest.AssertContains(t, h, "4 task")
}

func TestTodo_emptyAddIsNoop(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab() // → Input (leave empty)
	h.Tab() // → Add
	h.Enter()

	ginktest.AssertNotContains(t, h, "4 task")
}

func TestTodo_subtitleTracksDoneCount(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	ginktest.AssertContains(t, h, "0 done")

	h.Enter() // toggle first todo

	ginktest.AssertContains(t, h, "1 done")
}

func TestTodo_deleteAllShowsEmptyMessage(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('d')
	h.SendRune('d')
	h.SendRune('d')

	ginktest.AssertContains(t, h, "no tasks")
}
