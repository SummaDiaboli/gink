# Gink

A declarative, React-like TUI (terminal UI) framework for Go.

Gink brings component-based UI development to the terminal. You describe **what** your UI looks like using composable functions. Gink handles **how** it renders, diffs, and updates — no manual terminal escape sequences, no explicit event loop management, no MVC boilerplate.

Inspired by [Ink](https://github.com/vadimdemedes/ink) (React for CLI in JavaScript).

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Components](#components)
  - [Elements](#elements)
  - [Rules of Hooks](#rules-of-hooks)
- [Layout](#layout)
  - [Box](#box)
  - [Row](#row)
  - [Gaps](#gaps)
  - [Padding](#padding)
  - [Divider](#divider)
- [Styling](#styling)
- [Hooks](#hooks)
  - [UseState](#usestate)
  - [UseEffect](#useeffect)
  - [UseRef](#useref)
  - [UseInput](#useinput)
  - [UseInterval](#useinterval)
  - [UseTermSize](#usetermsize)
  - [UseFocus](#usefocus)
- [Built-in Components](#built-in-components)
  - [Spinner](#spinner)
  - [NewInput](#newinput)
  - [NewButton](#newbutton)
- [Testing](#testing)
  - [Writing tests](#writing-tests)
  - [Harness methods](#harness-methods)
  - [Assertions](#assertions)
  - [Async components](#async-components)
- [Plugins](#plugins)
  - [Writing a plugin](#writing-a-plugin)
  - [Plugin protocol](#plugin-protocol)
  - [Pushing updates](#pushing-updates)
- [Architecture](#architecture)

---

## Installation

```bash
go get github.com/salim/gink
```

Requires Go 1.21+ (generics are used for `UseState` and `UseRef`).

---

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/salim/gink"
)

func App() gink.Element {
    count, setCount := gink.UseState(0)

    gink.UseInput(func(ev gink.KeyEvent) {
        switch ev.Rune {
        case '+':
            setCount(count + 1)
        case '-':
            setCount(count - 1)
        }
    })

    return gink.Box(
        gink.Text("Counter", gink.NewStyle().Bold()),
        gink.Text(fmt.Sprintf("Count: %d", count)),
        gink.Text("Press +/- to change, Esc to quit"),
    )
}

func main() {
    if err := gink.Render(App); err != nil {
        log.Fatal(err)
    }
}
```

---

## Core Concepts

### Components

A **component** is a plain Go function that takes no arguments and returns an `Element`. It describes a piece of UI.

```go
func Greeting() gink.Element {
    return gink.Text("Hello, world!")
}
```

Components can use **hooks** to manage state and side effects. They are composable — nest them with `C()`:

```go
func App() gink.Element {
    return gink.Box(
        gink.C(Header),
        gink.C(Body),
        gink.C(Footer),
    )
}
```

**Parameterized components** — components that need arguments return a `func() Element` from a factory function:

```go
func Greeting(name string) func() gink.Element {
    return func() gink.Element {
        return gink.Text("Hello, " + name + "!")
    }
}

// Used as:
gink.C(Greeting("Alice"))
```

This pattern is used by all built-in interactive components (`NewInput`, `NewButton`).

### Elements

An **Element** is a lightweight, immutable description of a UI node. It is not the rendered output — it is the instruction that tells Gink what to render. Think of it as Gink's equivalent of a virtual DOM node.

Elements are created by calling layout and text functions:

```go
gink.Text("hello")
gink.Box(child1, child2)
gink.Row(child1, child2)
gink.C(MyComponent)
```

Elements form a tree. Gink walks the tree on every render, builds a virtual terminal buffer (a 2D grid of styled cells), diffs it against the previous buffer, and writes only the changed cells to the terminal.

### Rules of Hooks

Hooks must follow the same rules as React hooks:

1. **Only call hooks at the top level of a component function.** Do not call hooks inside `if` statements, loops, or nested functions. This ensures hooks are called in the same order on every render, which is how Gink tracks which state belongs to which hook.

2. **Only call hooks inside component functions.** Calling a hook outside a component panics with a clear error message.

```go
// Correct
func MyComponent() gink.Element {
    count, setCount := gink.UseState(0)   // always called
    gink.UseInput(func(ev gink.KeyEvent) {
        setCount(count + 1)
    })
    return gink.Text(fmt.Sprintf("%d", count))
}

// Wrong — conditional hook call
func MyComponent() gink.Element {
    if someCondition {
        count, _ := gink.UseState(0)   // breaks hook ordering
    }
    ...
}
```

---

## Layout

### Box

`Box` stacks children **vertically** (column direction), one per line.

```go
gink.Box(
    gink.Text("Line 1"),
    gink.Text("Line 2"),
    gink.Text("Line 3"),
)
```

Output:
```
Line 1
Line 2
Line 3
```

### Row

`Row` lays children out **horizontally**, side by side on the same line.

```go
gink.Row(
    gink.Text("Name: "),
    gink.Text("Alice"),
)
```

Output:
```
Name: Alice
```

Boxes and Rows are fully nestable. A `Row` containing a multi-line `Box` gets the correct height; a `Box` containing wide `Row`s gets the correct width.

```go
gink.Box(
    gink.Row(
        gink.Text("Left  "),
        gink.Text("Right"),
    ),
    gink.Row(
        gink.Text("A     "),
        gink.Text("B"),
    ),
)
```

Output:
```
Left  Right
A     B
```

### Gaps

Add spacing between children with `BoxWithGap` and `RowWithGap`. The gap value is the number of empty cells (rows for Box, columns for Row) inserted between each child.

```go
// 1 blank line between each child
gink.BoxWithGap(1,
    gink.Text("Section 1"),
    gink.Text("Section 2"),
    gink.Text("Section 3"),
)

// 3 spaces between each item
gink.RowWithGap(3,
    gink.Text("Alpha"),
    gink.Text("Beta"),
    gink.Text("Gamma"),
)
```

### Padding

`Padding` adds inner spacing around any element. The child is offset by the given amounts; the total element size grows to include the padding on each side.

```go
// Pad struct — set any combination of sides
gink.Padding(gink.Pad{Top: 1, Left: 2}, content)

// Equal spacing on all four sides
gink.PaddingAll(1, content)

// Horizontal (left/right) and vertical (top/bottom)
gink.PaddingXY(2, 1, content)
```

Wrap an entire app in `PaddingXY` to add breathing room without changing any individual component:

```go
func App() gink.Element {
    return gink.PaddingXY(2, 1,
        gink.Box(
            gink.Text("My App"),
            // ...
        ),
    )
}
```

### Divider

`Divider` renders a full-width horizontal rule (`─`) that automatically spans the current terminal width.

```go
// Plain rule
gink.C(gink.Divider)

// Rule with a centered label inset
gink.C(gink.DividerWithLabel("Section"))

// Rule with a style (color, bold, etc.)
gink.C(gink.DividerStyled(gink.NewStyle().Foreground(gink.ColorBrightBlack)))

// Label with a style
gink.C(gink.DividerWithLabel("Section", gink.NewStyle().Bold()))
```

`DividerWithLabel` centers the label and fills both sides with dashes:

```
───────────── Section ─────────────
```

If the terminal is too narrow to fit the label with dashes on both sides, the label is shown without dashes.

---

## Styling

Styles are built with `NewStyle()` and a chainable API. Pass a `Style` as the second argument to `Text`.

```go
// Bold cyan text
gink.Text("Hello", gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan))

// White text on red background
gink.Text("Error", gink.NewStyle().Foreground(gink.ColorWhite).Background(gink.ColorRed))

// Underlined
gink.Text("Link", gink.NewStyle().Underline())
```

**Style methods** — all return a new `Style` (immutable, chainable):

| Method | Description |
|---|---|
| `Foreground(Color)` | Set text color |
| `Background(Color)` | Set background color |
| `Bold()` | Bold text |
| `Underline()` | Underlined text |
| `Italic()` | Italic text (terminal support varies) |

**Color constants:**

| Normal | Bright |
|---|---|
| `ColorBlack` | — |
| `ColorRed` | `ColorBrightRed` |
| `ColorGreen` | `ColorBrightGreen` |
| `ColorYellow` | `ColorBrightYellow` |
| `ColorBlue` | `ColorBrightBlue` |
| `ColorMagenta` | `ColorBrightMagenta` |
| `ColorCyan` | `ColorBrightCyan` |
| `ColorWhite` | `ColorBrightWhite` |
| `ColorDefault` | — |

Store styles in package-level variables to avoid recreating them on every render:

```go
var titleStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
var errorStyle = gink.NewStyle().Foreground(gink.ColorBrightRed)
```

---

## Hooks

Hooks are functions that let components access Gink features — state, effects, keyboard input, and more. All hooks must be called at the top level of a component function (see [Rules of Hooks](#rules-of-hooks)).

### UseState

```go
value, setValue := gink.UseState(initialValue)
```

`UseState` adds local state to a component. Returns the current value and a setter function. Calling the setter schedules a re-render.

- `initialValue` — used only on the first render; ignored on subsequent renders
- `setValue(next)` — can be called from any goroutine; safe for use in effects and input handlers
- The value type is inferred from `initialValue` via generics

```go
func Counter() gink.Element {
    count, setCount := gink.UseState(0)
    name, setName := gink.UseState("Alice")

    gink.UseInput(func(ev gink.KeyEvent) {
        if ev.Rune == '+' {
            setCount(count + 1)
        }
    })

    return gink.Box(
        gink.Text(fmt.Sprintf("Count: %d", count)),
        gink.Text("Name: " + name),
    )
}
```

### UseEffect

```go
gink.UseEffect(func() func() {
    // side effect runs here
    return func() {
        // optional cleanup — runs before next effect or on unmount
    }
}, deps)
```

`UseEffect` runs a side effect after the UI has been rendered and flushed to the terminal. The `deps` argument controls when the effect re-runs:

| `deps` value | When effect runs |
|---|---|
| `nil` | After every render |
| `[]any{}` | Once on mount only |
| `[]any{a, b}` | When `a` or `b` change between renders |

The function passed to `UseEffect` may return a cleanup function. The cleanup runs before the next time the effect fires, and is the correct place to stop timers, close channels, or release resources.

```go
// Run once on mount
gink.UseEffect(func() func() {
    fmt.Println("component mounted")
    return nil
}, []any{})

// Reconnect when host changes
gink.UseEffect(func() func() {
    conn := connect(host)
    return func() { conn.Close() }
}, []any{host})

// Run after every render
gink.UseEffect(func() func() {
    log.Printf("rendered at %s", time.Now())
    return nil
}, nil)
```

### UseRef

```go
ref := gink.UseRef(initialValue)
// Read:  ref.Value
// Write: ref.Value = newValue
```

`UseRef` returns a `*Ref[T]` whose pointer identity is stable across renders. Unlike `UseState`, writing to `ref.Value` does **not** schedule a re-render.

The primary use case is holding values that effects need to read without being in the dependency array — particularly callbacks, which would otherwise capture stale values:

```go
func MyComponent() gink.Element {
    count, setCount := gink.UseState(0)

    // Without UseRef, the effect would capture the initial count=0 forever.
    // With UseRef, the goroutine always reads the latest count.
    countRef := gink.UseRef(count)
    countRef.Value = count

    gink.UseEffect(func() func() {
        go func() {
            // countRef.Value is always current because it's updated every render
            fmt.Printf("current count: %d\n", countRef.Value)
        }()
        return nil
    }, []any{})

    return gink.Text(fmt.Sprintf("Count: %d", count))
}
```

`UseRef` is used internally by `UseInterval` to keep the callback fresh without restarting the timer on every render.

### UseInput

```go
gink.UseInput(func(ev gink.KeyEvent) {
    // handle key event
})
```

`UseInput` registers a keyboard event handler for the current render pass. The handler is called for every key event received after the render completes. Handlers are rebuilt on every render — the closure always captures the latest state values.

`KeyEvent` fields:

| Field | Type | Description |
|---|---|---|
| `Rune` | `rune` | The character pressed (valid when `Key == KeyRune`) |
| `Key` | `tcell.Key` | The key code (for special keys) |

**Matching printable characters** — use `ev.Rune`:

```go
gink.UseInput(func(ev gink.KeyEvent) {
    switch ev.Rune {
    case '+':
        setCount(count + 1)
    case '-':
        setCount(count - 1)
    case 'q':
        os.Exit(0)
    }
})
```

**Matching special keys** — use `ev.Key` with the exported key constants:

```go
gink.UseInput(func(ev gink.KeyEvent) {
    switch ev.Key {
    case gink.KeyUp:
        moveUp()
    case gink.KeyDown:
        moveDown()
    case gink.KeyEnter:
        confirm()
    case gink.KeyBackspace, gink.KeyBackspace2:
        deleteChar()
    }
})
```

**Key constants:**

| Constant | Description |
|---|---|
| `KeyEnter` | Enter / Return |
| `KeyBackspace` | Backspace (BS, 0x08) |
| `KeyBackspace2` | Backspace (DEL, 0x7F) — some terminals send this |
| `KeyEscape` | Escape |
| `KeyUp` | Up arrow |
| `KeyDown` | Down arrow |
| `KeyLeft` | Left arrow |
| `KeyRight` | Right arrow |
| `KeyTab` | Tab (consumed by focus system; components don't receive it) |
| `KeyRune` | Any printable character — match via `ev.Rune` instead |

> **Note:** Tab and Shift+Tab are consumed by Gink's focus system and are not dispatched to `UseInput` handlers.

### UseInterval

```go
gink.UseInterval(duration, func() {
    // called on every tick
})
```

`UseInterval` calls `fn` on a repeating timer. The timer is created once when the component mounts and stopped automatically when the duration changes or the component unmounts.

The callback always receives the latest state values — `UseInterval` uses `UseRef` internally to keep the callback reference current, avoiding the stale closure problem that would occur with a raw `UseEffect`.

```go
func Clock() gink.Element {
    now, setNow := gink.UseState(time.Now())

    gink.UseInterval(time.Second, func() {
        setNow(time.Now())
    })

    return gink.Text(now.Format("15:04:05"))
}
```

> **Tip:** `UseInterval` is equivalent to `UseEffect` + `UseRef` + a ticker, but without the boilerplate. Use `UseEffect` directly only when you need finer control over when the timer starts and stops.

### UseTermSize

```go
size := gink.UseTermSize()
// size.Width  int
// size.Height int
```

`UseTermSize` returns the current terminal dimensions. The component automatically re-renders when the terminal is resized — no subscription or manual handling needed.

```go
func App() gink.Element {
    size := gink.UseTermSize()
    divider := strings.Repeat("─", size.Width)

    return gink.Box(
        gink.Text("My App"),
        gink.Text(divider),
        gink.Text(fmt.Sprintf("Terminal: %dx%d", size.Width, size.Height)),
    )
}
```

`TermSize` can be stored in a `UseEffect` dep array to trigger effects on resize:

```go
gink.UseEffect(func() func() {
    fmt.Printf("resized to %dx%d\n", size.Width, size.Height)
    return nil
}, []any{size})
```

### UseFocus

```go
isFocused := gink.UseFocus()
```

`UseFocus` registers the current component as a focusable element and returns whether it currently holds focus. Focusable components are cycled through in tree order using Tab (forward) and Shift+Tab (backward).

The first focusable component in the tree receives focus on startup. If the number of focusable components changes (e.g. conditional rendering), focus is clamped to the last valid index.

Use `isFocused` to:
- Gate input handling — ignore keypresses when not focused
- Change visual appearance to indicate focused state

```go
func SelectableItem(label string) gink.Element {
    isFocused := gink.UseFocus()

    gink.UseInput(func(ev gink.KeyEvent) {
        if !isFocused {
            return // ignore input when not focused
        }
        if ev.Key == gink.KeyEnter {
            fmt.Println("selected:", label)
        }
    })

    style := gink.NewStyle()
    if isFocused {
        style = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
    }

    return gink.Text(label, style)
}
```

> **Note:** `UseFocus` does not consume a hook slot — it does not participate in the ordered slot array used by `UseState`, `UseEffect`, and `UseRef`. It can be called in any position.

---

## Built-in Components

### Spinner

`Spinner` displays an animated braille spinner (`⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏`) that cycles at 80ms per frame.

```go
// Default spinner
gink.C(gink.Spinner)

// With style
gink.C(gink.SpinnerWithStyle(gink.NewStyle().Foreground(gink.ColorBrightCyan)))
```

Use conditionally to indicate loading:

```go
func App() gink.Element {
    loading, _ := gink.UseState(true)

    if loading {
        return gink.Row(
            gink.C(gink.Spinner),
            gink.Text(" Loading..."),
        )
    }
    return gink.Text("Done!")
}
```

### NewInput

`NewInput` is a controlled single-line text input component. The parent component owns the state.

```go
gink.C(gink.NewInput(value, onChange))
```

| Argument | Type | Description |
|---|---|---|
| `value` | `string` | Current text content |
| `onChange` | `func(string)` | Called with the full new string on each keystroke |

- Handles both `KeyBackspace` keycodes for cross-terminal compatibility
- Shows a block cursor `█` when focused
- Cyan brackets when focused, default color when not

```go
func Form() gink.Element {
    name, setName := gink.UseState("")
    email, setEmail := gink.UseState("")

    return gink.BoxWithGap(1,
        gink.Row(gink.Text("Name:  "), gink.C(gink.NewInput(name, setName))),
        gink.Row(gink.Text("Email: "), gink.C(gink.NewInput(email, setEmail))),
    )
}
```

Tab moves focus between input fields automatically.

### NewButton

`NewButton` is a focusable button that activates on Enter or Space.

```go
gink.C(gink.NewButton(label, onPress))
```

| Argument | Type | Description |
|---|---|---|
| `label` | `string` | Text displayed inside the button |
| `onPress` | `func()` | Called when the button is activated |

- Renders as `[ label ]`
- Bold cyan when focused, default when not
- Activated by Enter or Space when focused

```go
gink.RowWithGap(2,
    gink.C(gink.NewButton("Save", func() { save() })),
    gink.C(gink.NewButton("Cancel", func() { cancel() })),
)
```

---

## Testing

Gink ships a companion package, `github.com/salim/gink/ginktest`, that lets you test components without a real terminal. It follows the same pattern as [`net/http/httptest`](https://pkg.go.dev/net/http/httptest) — import it only in `_test.go` files.

```bash
go get github.com/salim/gink/ginktest
```

### Writing tests

Create a `Harness`, simulate user input, and assert on the screen contents:

```go
package main

import (
    "testing"

    "github.com/salim/gink/ginktest"
)

// Focus order: Input(0) · Increment(1) · Decrement(2)

func TestCounter_increment(t *testing.T) {
    h := ginktest.NewHarness(t, App)
    defer h.Close()

    h.Tab()   // → Increment button
    h.Enter() // press it

    ginktest.AssertContains(t, h, "Count is 1")
}

func TestCounter_typeInInput(t *testing.T) {
    h := ginktest.NewHarness(t, App)
    defer h.Close()

    // Input has focus by default (focus index 0).
    h.SendRune('A')
    h.SendRune('l')
    h.SendRune('i')

    ginktest.AssertContains(t, h, "Ali")
}
```

Document the focus order in a comment at the top of each test file. This makes the `Tab()` sequences readable without having to trace through the component tree.

### Harness methods

| Method | Description |
|---|---|
| `Tab()` | Advance focus to the next focusable component |
| `ShiftTab()` | Move focus to the previous focusable component |
| `Enter()` | Press Enter on the focused component |
| `Backspace()` | Press Backspace |
| `SendRune(r)` | Type a printable character |
| `SendKey(key)` | Send a special key (arrow keys, Escape, etc.) |
| `Render()` | Force a re-render without input (for async polling) |
| `Contains(s)` | Returns true if any screen line contains s |
| `Lines()` | Full screen as `[]string`, one entry per row |
| `Line(y)` | Trimmed content of row y |
| `CellStyle(x, y)` | Raw `tcell.Style` at a cell (for color/style assertions) |
| `Close()` | Release the simulation screen |

Arrow keys and other special keys use the `gink.Key*` constants:

```go
h.SendKey(gink.KeyDown)
h.SendKey(gink.KeyUp)
h.SendKey(gink.KeyEscape)
```

### Assertions

```go
ginktest.AssertContains(t, h, "text")     // fails if text is absent
ginktest.AssertNotContains(t, h, "text")  // fails if text is present
```

Both print the full screen contents on failure so you can see exactly what was rendered.

### Async components

Components that use `UseInterval` or launch goroutines inside `UseEffect` update state asynchronously. Use `AwaitContains` to poll until the expected text appears:

```go
func TestClock_ticks(t *testing.T) {
    h := ginktest.NewHarness(t, Clock)
    defer h.Close()

    // Wait up to 500ms for the timer to fire at least once.
    ginktest.AwaitContains(t, h, "00:01", 500*time.Millisecond)
}
```

`AwaitContains` re-renders every 50 ms and fails the test if the timeout elapses before the text appears.

---

## Plugins

Gink can embed UI written in other languages via subprocess plugins. A plugin is any executable that speaks newline-delimited JSON (NDJSON) on stdin/stdout. Client SDKs for Node.js and Python are provided in `sdk/`.

```go
// Wire up a plugin — it becomes a normal Gink component in the tree.
gink.C(gink.NewPlugin("node", "my-plugin.js"))
gink.C(gink.NewPlugin("python", "my-plugin.py"))

// Pre-configure the process (env vars, working directory, etc.)
cmd := exec.Command("node", "plugin.js")
cmd.Dir = "/path/to/plugin"
cmd.Env = append(os.Environ(), "API_KEY=secret")
gink.C(gink.NewPluginCmd(cmd))
```

While the plugin is starting up, a spinner is displayed. If the process fails to start, an error message is shown in place of the spinner.

### Writing a plugin

**Node.js**

```js
const gink = require('./sdk/js/gink');

let count = 0;

gink.onMessage((msg) => {
  if (msg.type === 'render' || (msg.type === 'input' && msg.key === 'enter')) {
    if (msg.type === 'input') count++;
    gink.element(
      gink.boxGap(1,
        gink.text('JS Counter', { bold: true, fg: 'brightCyan' }),
        gink.text(`Count: ${count}`, { fg: 'brightYellow' }),
        gink.text('Press Enter to increment'),
      )
    );
  }
  if (msg.type === 'unmount') process.exit(0);
});
```

**Python**

```python
import gink

count = 0

def handle(msg):
    global count
    if msg['type'] == 'render':
        show()
    elif msg['type'] == 'input' and msg.get('key') == 'enter':
        count += 1
        show()
    elif msg['type'] == 'unmount':
        import sys; sys.exit(0)

def show():
    gink.element(gink.box(
        gink.text(f'Count: {count}', bold=True, fg='brightYellow'),
        gink.text('Press Enter to increment'),
        gap=1,
    ))

gink.on_message(handle)
gink.run()
```

### Plugin protocol

All messages are single JSON objects terminated by `\n`.

**Host → Plugin**

| `type`    | Fields                               | Description |
|-----------|--------------------------------------|-------------|
| `render`  | —                                    | Sent once on mount. Plugin should respond with an element. |
| `input`   | `key`, `rune` (when `key` = `rune`)  | A key event forwarded from the terminal (only when the plugin holds focus). |
| `unmount` | —                                    | Plugin should clean up and exit. |

Input key names: `enter`, `backspace`, `escape`, `up`, `down`, `left`, `right`, `rune` (read the `rune` field for the character).

**Plugin → Host**

| `type`    | Fields | Description |
|-----------|--------|-------------|
| `element` | `tree` | A new element tree to display. Can be sent at any time. |

**Element JSON**

```json
{ "type": "text", "content": "Hello", "style": { "bold": true, "fg": "brightCyan" } }

{ "type": "box", "gap": 1, "children": [ ... ] }

{ "type": "row", "gap": 0, "children": [ ... ] }
```

Style color names: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, and the `bright` variants (`brightRed`, `brightCyan`, etc.).

### Pushing updates

Plugins can send `element` messages at any time — not just in response to `render` or `input`. Use this for timers or async work:

```python
import gink, threading, time

elapsed = 0.0

def tick():
    global elapsed
    while True:
        time.sleep(0.1)
        elapsed += 0.1
        gink.element(gink.text(f'Elapsed: {elapsed:.1f}s'))

def handle(msg):
    if msg['type'] == 'render':
        threading.Thread(target=tick, daemon=True).start()
        gink.element(gink.text('Starting...'))
    elif msg['type'] == 'unmount':
        import sys; sys.exit(0)

gink.on_message(handle)
gink.run()
```

See `sdk/README.md` for the full protocol specification.

---

## Architecture

Gink is built on five layers:

```
┌───────────────────────────────────────────┐
│              User Components              │
│     Plain functions returning Elements    │
└──────────────────┬────────────────────────┘
                   │ Element tree
                   ▼
┌───────────────────────────────────────────┐
│               Reconciler                  │
│  Walks tree → invokes components →        │
│  manages hook stores by tree path →       │
│  paints into virtual Buffer               │
└──────────────────┬────────────────────────┘
                   │ Buffer (2D cell grid)
                   ▼
┌───────────────────────────────────────────┐
│            Diff Engine (buffer.go)        │
│  Compares new Buffer against previous →   │
│  emits only changed CellUpdates           │
└──────────────────┬────────────────────────┘
                   │ []CellUpdate
                   ▼
┌───────────────────────────────────────────┐
│            Renderer (tcell)               │
│  Writes changed cells to real terminal    │
└───────────────────────────────────────────┘
```

**Hook identity** — Each component instance is identified by its position in the element tree (its path, e.g. `"root/0/1/0"`). The reconciler maintains a `map[string]*renderContext` so each component has its own hook slot array. Hooks are matched by call order within a render, which is why their order must not change between renders.

**Rendering pipeline** — On each render pass:
1. `inputHandlers`, `pendingEffects`, and `focusables` slices are cleared
2. The element tree is walked; components are called, hooks are evaluated, cells are painted
3. The new `Buffer` is diffed against the previous one
4. Only changed cells are written to the terminal via tcell
5. Effects registered during the walk are executed in order

**Event loop** — `tcell.PollEvent` runs in a dedicated goroutine and sends events to a buffered channel. The main loop selects between that channel and a `dirty` channel (written to by state setters). This ensures that timer-driven re-renders (from `UseInterval`) are not blocked waiting for a keypress.
