package main

import (
	"testing"

	"github.com/SummaDiaboli/gink/ginktest"
)

// Focus order: Name(0) · Email(1) · Subject(2) · Message(3) · Submit(4)

func TestForm_initialRender(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	ginktest.AssertContains(t, h, "Contact Form")
	ginktest.AssertContains(t, h, "Name")
	ginktest.AssertContains(t, h, "Email")
	ginktest.AssertContains(t, h, "Subject")
	ginktest.AssertContains(t, h, "Message")
}

func TestForm_emptySubmitShowsError(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.Tab()
	h.Tab()
	h.Tab()
	h.Tab()   // → Submit
	h.Enter()

	ginktest.AssertContains(t, h, "Name is required")
}

func TestForm_missingEmailShowsError(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('A') // Name
	h.Tab()         // → Email (leave empty)
	h.Tab()         // → Subject
	h.Tab()         // → Message
	h.Tab()         // → Submit
	h.Enter()

	ginktest.AssertContains(t, h, "Email is required")
}

func TestForm_invalidEmailShowsError(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('A') // Name
	h.Tab()
	h.SendRune('n') // Email — no @
	h.SendRune('o')
	h.Tab() // → Subject
	h.Tab() // → Message
	h.Tab() // → Submit
	h.Enter()

	ginktest.AssertContains(t, h, "Email must contain @")
}

func TestForm_missingSubjectShowsError(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('A') // Name
	h.Tab()
	h.SendRune('a') // Email
	h.SendRune('@')
	h.SendRune('b')
	h.Tab() // → Subject (leave empty)
	h.Tab() // → Message
	h.Tab() // → Submit
	h.Enter()

	ginktest.AssertContains(t, h, "Subject is required")
}

func TestForm_missingBodyShowsError(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('A') // Name
	h.Tab()
	h.SendRune('a') // Email
	h.SendRune('@')
	h.SendRune('b')
	h.Tab()
	h.SendRune('H') // Subject
	h.Tab()         // → Message (leave empty)
	h.Tab()         // → Submit
	h.Enter()

	ginktest.AssertContains(t, h, "Message body is required")
}

func TestForm_validSubmitShowsSuccess(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('A') // Name: A
	h.Tab()
	h.SendRune('a') // Email: a@b
	h.SendRune('@')
	h.SendRune('b')
	h.Tab()
	h.SendRune('H') // Subject: Hi
	h.SendRune('i')
	h.Tab()
	h.SendRune('M') // Message: Msg
	h.SendRune('s')
	h.SendRune('g')
	h.Tab()   // → Submit
	h.Enter()

	ginktest.AssertContains(t, h, "Message sent!")
	ginktest.AssertContains(t, h, "Thanks, A.")
}

func TestForm_sendAnotherResetsForm(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	h.SendRune('A')
	h.Tab()
	h.SendRune('a')
	h.SendRune('@')
	h.SendRune('b')
	h.Tab()
	h.SendRune('H')
	h.Tab()
	h.SendRune('M')
	h.Tab()
	h.Enter()

	ginktest.AssertContains(t, h, "Message sent!")

	h.Enter() // press "Send another"

	ginktest.AssertContains(t, h, "Contact Form")
	ginktest.AssertNotContains(t, h, "Message sent!")
}

func TestForm_typingClearsError(t *testing.T) {
	h := ginktest.NewHarness(t, App)
	defer h.Close()

	// Trigger an error by submitting empty.
	h.Tab()
	h.Tab()
	h.Tab()
	h.Tab()
	h.Enter()
	ginktest.AssertContains(t, h, "Name is required")

	// Navigate back to Name and type something.
	h.ShiftTab()
	h.ShiftTab()
	h.ShiftTab()
	h.ShiftTab()
	h.SendRune('X')

	ginktest.AssertNotContains(t, h, "Name is required")
}
