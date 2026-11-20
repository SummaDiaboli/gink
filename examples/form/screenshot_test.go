package main

import (
	"os"
	"testing"

	"github.com/SummaDiaboli/gink/ginktest"
	"github.com/SummaDiaboli/gink/internal/readmeimg"
)

// TestScreenshotForm regenerates the README form screenshot. It is skipped
// unless GINK_SCREENSHOTS is set:
//
//	GINK_SCREENSHOTS=1 go test ./examples/form -run TestScreenshotForm -v
func TestScreenshotForm(t *testing.T) {
	if os.Getenv("GINK_SCREENSHOTS") == "" {
		t.Skip("set GINK_SCREENSHOTS=1 to regenerate docs/assets/form.png")
	}
	t.Logf("rasterizing with %s", readmeimg.FontSource())

	h := ginktest.NewHarnessSize(t, App, 80, 21)
	defer h.Close()

	typeString(h, "Ada Lovelace")
	h.Tab()
	typeString(h, "ada@example.com")
	h.Tab()
	typeString(h, "Terminal UIs in Go")
	h.Tab()
	typeString(h, "Gink makes composing TUIs feel like React.")
	h.Tab() // focus Submit

	if err := readmeimg.WritePNG("../../docs/assets/form.png", readmeimg.Capture(h), "gink — form components"); err != nil {
		t.Fatalf("write form PNG: %v", err)
	}
}

func typeString(h interface {
	SendRune(rune)
}, s string) {
	for _, r := range s {
		h.SendRune(r)
	}
}
