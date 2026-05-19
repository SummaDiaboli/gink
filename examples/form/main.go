// Contact form — demonstrates multiple controlled inputs, form validation,
// and conditional rendering (form view vs. success view).
package main

import (
	"log"
	"strings"

	"github.com/SummaDiaboli/gink"
)

var (
	titleStyle   = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	labelStyle   = gink.NewStyle().Foreground(gink.ColorBrightWhite)
	errorStyle   = gink.NewStyle().Bold().Foreground(gink.ColorBrightRed)
	successStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightGreen)
	mutedStyle   = gink.NewStyle().Foreground(gink.ColorWhite)
	hintStyle    = gink.NewStyle().Foreground(gink.ColorWhite)
)

func App() gink.Element {
	name, setName := gink.UseState("")
	email, setEmail := gink.UseState("")
	subject, setSubject := gink.UseState("")
	body, setBody := gink.UseState("")
	errMsg, setErrMsg := gink.UseState("")
	submitted, setSubmitted := gink.UseState(false)

	if submitted {
		return gink.PaddingXY(2, 1,
			gink.BoxWithGap(1,
				gink.C(gink.DividerWithLabel("Contact Form", titleStyle)),
				gink.Text("✓ Message sent!", successStyle),
				gink.Text("Thanks, "+name+". We'll reply to "+email+" shortly.", mutedStyle),
				gink.C(gink.NewButton("Send another", func() {
					setSubmitted(false)
					setName("")
					setEmail("")
					setSubject("")
					setBody("")
					setErrMsg("")
				})),
				gink.C(gink.Divider),
				gink.Text("Tab to focus  ·  Enter/Space to activate  ·  Esc to quit", hintStyle),
			),
		)
	}

	submit := func() {
		switch {
		case strings.TrimSpace(name) == "":
			setErrMsg("Name is required")
		case strings.TrimSpace(email) == "":
			setErrMsg("Email is required")
		case !strings.Contains(email, "@"):
			setErrMsg("Email must contain @")
		case strings.TrimSpace(subject) == "":
			setErrMsg("Subject is required")
		case strings.TrimSpace(body) == "":
			setErrMsg("Message body is required")
		default:
			setErrMsg("")
			setSubmitted(true)
		}
	}

	// Clear the error as soon as any field changes.
	clearErr := func(set func(string)) func(string) {
		return func(v string) {
			set(v)
			if errMsg != "" {
				setErrMsg("")
			}
		}
	}

	var statusLine gink.Element
	if errMsg != "" {
		statusLine = gink.Text("⚠  "+errMsg, errorStyle)
	} else {
		statusLine = gink.Text("Fill in all fields, then press Submit.", mutedStyle)
	}

	return gink.PaddingXY(2, 1,
		gink.BoxWithGap(1,
			gink.C(gink.DividerWithLabel("Contact Form", titleStyle)),

			gink.Row(gink.Text("Name     ", labelStyle), gink.C(gink.NewInput(name, clearErr(setName)))),
			gink.Row(gink.Text("Email    ", labelStyle), gink.C(gink.NewInput(email, clearErr(setEmail)))),
			gink.Row(gink.Text("Subject  ", labelStyle), gink.C(gink.NewInput(subject, clearErr(setSubject)))),
			gink.Row(gink.Text("Message  ", labelStyle), gink.C(gink.NewInput(body, clearErr(setBody)))),

			gink.C(gink.Divider),

			statusLine,

			gink.C(gink.NewButton("  Submit  ", submit)),

			gink.C(gink.Divider),

			gink.Text("Tab to move between fields  ·  Esc to quit", hintStyle),
		),
	)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}
