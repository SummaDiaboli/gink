package main

import (
	"testing"

	"github.com/salim/gink/ginktest"
)

// Focus order in App: Input(0) · Increment(1) · Decrement(2) · Toggle Loading(3)

func TestCounter_initialRender(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	ginktest.AssertContains(t, h, "Count is 0")
}

func TestCounter_increment(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab()   // → Increment
	h.Enter()

	ginktest.AssertContains(t, h, "Count is 1")
}

func TestCounter_decrement(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab()   // → Increment
	h.Enter() // count = 1
	h.Tab()   // → Decrement
	h.Enter() // count = 0

	ginktest.AssertContains(t, h, "Count is 0")
}

func TestCounter_decrementBelowZero(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab()   // → Increment
	h.Tab()   // → Decrement
	h.Enter()

	ginktest.AssertContains(t, h, "Count is -1")
}

func TestCounter_multipleIncrements(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab() // → Increment
	h.Enter()
	h.Enter()
	h.Enter()

	ginktest.AssertContains(t, h, "Count is 3")
}

func TestCounter_toggleLoadingHidesCount(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab() // → Increment
	h.Tab() // → Decrement
	h.Tab() // → Toggle Loading
	h.Enter()

	ginktest.AssertNotContains(t, h, "Count is")
}

func TestCounter_nameInput(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Input has focus by default.
	h.SendRune('A')
	h.SendRune('l')
	h.SendRune('i')

	ginktest.AssertContains(t, h, "Ali")
}
