package gink

import "time"

// UseToast returns a toast Element and a show function. Mount the Element in
// your layout at the position where the toast should appear; call show to
// display a message that auto-dismisses after duration.
//
// The Element renders as an empty Box when no toast is active and as a styled
// text row when one is.  Calling show again before the previous toast expires
// replaces the message and resets the timer.
//
//	toastEl, showToast := gink.UseToast()
//	// In layout:
//	gink.Box(mainContent, toastEl)
//	// In an event handler:
//	showToast("Saved!", 3*time.Second)
func UseToast() (Element, func(string, time.Duration)) {
	msg, setMsg := UseState("")
	dur, setDur := UseState(time.Duration(0))
	gen, setGen := UseState(0)

	show := func(m string, d time.Duration) {
		setMsg(m)
		setDur(d)
		setGen(gen + 1)
	}

	// Each time gen changes (i.e. show was called), start a new dismissal
	// timer. The cleanup cancels any in-flight timer before the next one starts.
	UseEffect(func() func() {
		if msg == "" || gen == 0 {
			return nil
		}
		t := time.AfterFunc(dur, func() { setMsg("") })
		return func() { t.Stop() }
	}, []any{gen})

	theme := UseTheme()
	var el Element
	if msg == "" {
		el = Box()
	} else {
		el = Text(msg, theme.Focused)
	}
	return el, show
}
