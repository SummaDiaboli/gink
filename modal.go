package gink

// ModalAction is a labelled button in a [NewModal].
type ModalAction struct {
	Label   string
	OnPress func()
}

// NewModal returns a modal dialog component that traps focus within its
// subtree until dismissed.
//
// title appears in the top border. content is rendered above the action
// buttons. actions is a slice of [ModalAction] values; each becomes a button
// rendered in a horizontal row at the bottom. onClose is called when the user
// presses Escape.
//
// While the modal is mounted, Tab and Shift+Tab cycle only between the
// modal's own buttons — focus cannot reach components outside the modal.
// Focus is automatically moved inside the modal on the first render.
//
//	gink.C(gink.NewModal(
//	    "Confirm",
//	    gink.Text("Delete this file?"),
//	    []gink.ModalAction{
//	        {Label: "Delete", OnPress: func() { doDelete() }},
//	        {Label: "Cancel", OnPress: func() { setOpen(false) }},
//	    },
//	    func() { setOpen(false) },
//	))
func NewModal(title string, content Element, actions []ModalAction, onClose func(), styles ...Style) func() Element {
	return func() Element {
		UseFocusBarrier()

		UseInput(func(ev KeyEvent) {
			if ev.Key == KeyEscape {
				onClose()
			}
		})

		buttons := make([]Element, len(actions))
		for i, action := range actions {
			a := action
			buttons[i] = C(NewButton(a.Label, a.OnPress))
		}

		var inner Element
		if len(buttons) > 0 {
			inner = Box(content, Row(buttons...))
		} else {
			inner = content
		}
		return BorderWithTitle(title, inner)
	}
}
