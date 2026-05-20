# Gink

A declarative, React-like TUI (terminal UI) framework for Go.

Gink brings component-based UI development to the terminal. You describe **what** your UI looks like using composable functions. Gink handles **how** it renders, diffs, and updates — no manual escape sequences, no explicit event loop, no MVC boilerplate.

Inspired by [Ink](https://github.com/vadimdemedes/ink) (React for CLI in JavaScript).

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Layout](#layout)
- [Styling](#styling)
- [Theming](#theming)
- [Hooks](#hooks)
- [Built-in Components](#built-in-components)
- [Scroll](#scroll)
- [Testing](#testing)
- [Plugins](#plugins)
- [Architecture](#architecture)

---

## Installation

```bash
go get github.com/SummaDiaboli/gink
```

Requires Go 1.25+ (generics are used for `UseState`, `UseAsync`, and `UseRef`; `golang.org/x/image` is used for image scaling).

---

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/SummaDiaboli/gink"
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

A **component** is a plain Go function that returns an `Element`. It describes a piece of UI and can call hooks for state and side effects.

```go
func Greeting() gink.Element {
    return gink.Text("Hello, world!")
}
```

Compose components with `C()`:

```go
func App() gink.Element {
    return gink.Box(
        gink.C(Header),
        gink.C(Body),
        gink.C(Footer),
    )
}
```

**Parameterized components** — components that need arguments return a `func() gink.Element` from a factory:

```go
func Greeting(name string) func() gink.Element {
    return func() gink.Element {
        return gink.Text("Hello, " + name + "!")
    }
}

gink.C(Greeting("Alice"))
```

All built-in interactive components (`NewInput`, `NewList`, `NewTable`, etc.) follow this factory pattern.

### Rules of Hooks

1. **Only call hooks at the top level** of a component function — not inside `if`, loops, or nested functions.
2. **Only call hooks inside component functions** — calling a hook outside a component panics.

---

## Layout

### Box and Row

`Box` stacks children **vertically**; `Row` lays them out **horizontally**.

```go
gink.Box(gink.Text("Line 1"), gink.Text("Line 2"))
gink.Row(gink.Text("Left: "), gink.Text("value"))
```

Add uniform spacing between children with `BoxWithGap` and `RowWithGap`:

```go
gink.BoxWithGap(1, child1, child2, child3)   // 1 blank row between each
gink.RowWithGap(2, child1, child2, child3)   // 2 blank cols between each
```

### Padding

```go
gink.PaddingAll(1, content)          // equal on all sides
gink.PaddingXY(2, 1, content)        // horizontal=2, vertical=1
gink.Padding(gink.Pad{Top: 1, Left: 2}, content)
```

### Divider

Full-width horizontal rules that span the current terminal width.

```go
gink.C(gink.Divider)
gink.C(gink.DividerWithLabel("Section"))
gink.C(gink.DividerStyled(gink.NewStyle().Foreground(gink.ColorBrightBlack)))
```

### Border

Draw a line-art border around any content.

```go
gink.Border(content)
gink.BorderWithTitle("Panel", content, gink.NewStyle().Bold())
```

### Constrain, Width, Height, Size

Enforce dimension constraints without changing the child's layout logic.

```go
// Exact dimensions
gink.Width(20, label)           // always 20 cols wide
gink.Height(5, content)         // always 5 rows tall
gink.Size(20, 5, content)       // both at once

// Bounds
gink.MinWidth(10, label)        // at least 10 cols
gink.MaxWidth(40, longText)     // clip at 40 cols
gink.MinHeight(3, header)       // at least 3 rows
gink.MaxHeight(8, list)         // clip at 8 rows

// Full control
gink.Constrain(child, minW, maxW, minH, maxH)
```

### TextWrapped

Word-wrap a string to a given column width. Existing `\n` characters are always honoured; words longer than the width are hard-broken.

```go
gink.TextWrapped(longDescription, 60)
gink.TextWrapped(longDescription, 60, gink.NewStyle().Foreground(gink.ColorWhite))
```

### AppShell

Pin a footer to the bottom of the screen, outside the scrollable viewport.

```go
func App() gink.Element {
    main := gink.PaddingXY(2, 1, gink.Box(/* ... */))
    footer := gink.Text("Tab: focus  ·  Esc: quit", hintStyle)
    return gink.AppShell(main, footer)
}
```

---

## Styling

Styles are built with `NewStyle()` and a chainable API. Pass a `Style` as the optional last argument to `Text` and most built-in components.

```go
gink.Text("Title", gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan))
```

| Method | Description |
|---|---|
| `Foreground(Color)` | Text color |
| `Background(Color)` | Background color |
| `Bold()` | Bold text |
| `Underline()` | Underlined text |
| `Italic()` | Italic (terminal support varies) |
| `Reverse()` | Swap foreground and background (used for cursors) |

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

**True-colour (24-bit):**

```go
red := gink.NewRGBColor(255, 0, 0)
gink.Text("RGB!", gink.NewStyle().Foreground(red))
```

---

## Theming

Gink has a built-in theme system that lets you define a consistent colour palette and have all built-in components pick it up automatically.

```go
// Declare a custom theme
myTheme := gink.Theme{
    Focused: gink.NewStyle().Bold().Foreground(gink.ColorBrightMagenta),
    Accent:  gink.NewStyle().Bold().Foreground(gink.ColorBrightBlue),
    Muted:   gink.NewStyle().Foreground(gink.ColorWhite),
    Error:   gink.NewStyle().Bold().Foreground(gink.ColorBrightRed),
    Success: gink.NewStyle().Bold().Foreground(gink.ColorBrightGreen),
    Warning: gink.NewStyle().Bold().Foreground(gink.ColorBrightYellow),
    Cursor:  gink.NewStyle().Reverse(),
}

// Apply it at the root of your app
func App() gink.Element {
    gink.SetContext(gink.ThemeCtx, myTheme)
    // ...
}
```

Read the current theme from any component with `UseTheme`:

```go
func MyComponent() gink.Element {
    theme := gink.UseTheme()
    return gink.Text("Accent text", theme.Accent)
}
```

Built-in interactive components (`NewButton`, `NewList`, `NewSelect`, `NewTable`, `NewTextArea`) read `theme.Focused` (or `theme.Cursor` for text cursors) automatically. Pass an explicit style as the optional last argument to override the theme for a specific instance.

**`Theme` fields:**

| Field | Default | Used by |
|---|---|---|
| `Focused` | Bold bright-cyan | `NewButton`, `NewList`, `NewSelect`, `NewTable` highlight |
| `Accent` | Bold bright-blue | — (available for custom components) |
| `Muted` | White | — |
| `Error` | Bold bright-red | — |
| `Success` | Bold bright-green | — |
| `Warning` | Bold bright-yellow | — |
| `Cursor` | Reverse video | `NewTextArea` cursor, `NewInput` cursor |

---

## Hooks

All hooks must be called at the top level of a component function.

### UseState

```go
value, setValue := gink.UseState(initialValue)
```

Local state for a component. Calling `setValue` schedules a re-render. Safe to call from any goroutine.

### UseEffect

```go
gink.UseEffect(func() func() {
    // side effect
    return func() { /* cleanup */ }
}, deps)
```

Runs after the UI is flushed. `deps` controls when it re-runs:

| `deps` | When runs |
|---|---|
| `nil` | Every render |
| `[]any{}` | Once on mount |
| `[]any{a, b}` | When `a` or `b` change |

### UseRef

```go
ref := gink.UseRef(initialValue)
ref.Value = newValue // does not trigger re-render
```

Stable pointer across renders. Use for values that effects need without triggering re-renders on change.

### UseInput

```go
gink.UseInput(func(ev gink.KeyEvent) {
    switch ev.Key {
    case gink.KeyUp:   moveUp()
    case gink.KeyDown: moveDown()
    }
    switch ev.Rune {
    case '+': increment()
    }
})
```

Registers a keyboard handler for the current render pass. Tab, Shift+Tab, Escape, and Ctrl+C are consumed by the runtime and not dispatched here.

**Key constants:**

| Constant | Description |
|---|---|
| `KeyEnter` | Enter / Return |
| `KeyBackspace` / `KeyBackspace2` | Backspace (two keycodes for cross-terminal compat) |
| `KeyEscape` | Escape |
| `KeyUp` / `KeyDown` / `KeyLeft` / `KeyRight` | Arrow keys |
| `KeyHome` / `KeyEnd` | Home / End |
| `KeyPgUp` / `KeyPgDn` | Page Up / Down (consumed by global scroll) |
| `KeyTab` | Tab (consumed by focus system) |
| `KeyRune` | Printable character — read `ev.Rune` |

### UseInterval

```go
gink.UseInterval(time.Second, func() {
    setNow(time.Now())
})
```

Repeating timer. Always captures the latest state — no stale closure problem.

### UseTermSize

```go
size := gink.UseTermSize()
// size.Width, size.Height
```

Returns current terminal dimensions. Component re-renders automatically on resize.

### UseFocus

```go
isFocused := gink.UseFocus()
```

Registers the component as a Tab stop and returns whether it currently holds focus. Tab cycles forward, Shift+Tab cycles backward. When a focused component is off-screen, the viewport auto-scrolls to reveal it.

### UseFocusWithin

```go
isFocused := gink.UseFocusWithin()
```

Returns true when any **descendant** component holds focus — without adding an extra Tab stop. Use this to highlight a container (border, panel) when focus is inside it.

```go
func Panel() gink.Element {
    focused := gink.UseFocusWithin()
    style := gink.NewStyle().Foreground(gink.ColorCyan)
    if focused {
        style = gink.NewStyle().Bold().Foreground(gink.ColorBrightWhite)
    }
    return gink.BorderWithTitle("My Panel", content, style)
}
```

### UseAsync

```go
result, loading, err := gink.UseAsync(func() (T, error) {
    return fetchData()
}, deps)
```

Runs an async function and returns its result, a loading flag, and any error. Re-fetches whenever `deps` change. Stale results from superseded fetches are discarded automatically.

```go
data, loading, err := gink.UseAsync(func() ([]Item, error) {
    return api.List(query)
}, []any{query})

if loading {
    return gink.Row(gink.C(gink.Spinner), gink.Text(" Loading…"))
}
```

### UseContext

Share state across the component tree without prop drilling.

```go
// Declare at package level
var ThemeCtx = gink.NewContext("light")

// Write from any component
gink.SetContext(ThemeCtx, "dark")

// Read from any component
theme := gink.UseContext(ThemeCtx)
```

### UseKeyboard

Global keyboard handler — fires for every keypress regardless of which component holds focus. Useful for app-wide shortcuts.

```go
gink.UseKeyboard(func(ev gink.KeyEvent) {
    if ev.Rune == 'q' {
        os.Exit(0)
    }
    if ev.Key == gink.KeyF1 {
        showHelp()
    }
})
```

Unlike `UseInput`, `UseKeyboard` is not gated by focus.

### UseClick

```go
gink.UseClick(func(x, y int) {
    // x, y are terminal cell coordinates
})
```

Fires on left mouse button clicks within the component's rendered area. Focus is transferred to the clicked component automatically; `UseClick` is called after focus transfer.

### UseAccessibility

Registers a human-readable label for the currently focused component. The label is exposed via the terminal title (readable by screen readers that monitor the title) and via `Harness.AccessibilityLabel()` in tests.

```go
func SearchBox() func() gink.Element {
    return func() gink.Element {
        query, setQuery := gink.UseState("")
        gink.UseAccessibility("Search box — type to filter results")
        return gink.C(gink.NewInput(query, setQuery))
    }
}
```

---

## Built-in Components

All interactive components (`New*`) are factory functions — they return `func() gink.Element` and must be wrapped with `C()`.

### Spinner

```go
gink.C(gink.Spinner)
gink.C(gink.SpinnerWithStyle(gink.NewStyle().Foreground(gink.ColorBrightCyan)))
```

Animated braille spinner at 80 ms per frame.

### NewInput

Single-line controlled text input.

```go
name, setName := gink.UseState("")
gink.C(gink.NewInput(name, setName))
```

Shows a block cursor `█` when focused. Tab moves between inputs automatically.

### NewTextArea

Multi-line controlled text input with cursor, line splitting/merging, and viewport scrolling.

```go
body, setBody := gink.UseState("")
gink.C(gink.NewTextArea(body, setBody, 5)) // 5 visible rows
```

Keys: Left/Right (within line), Up/Down (between lines), Home/End, Enter (split line), Backspace (delete/merge).

### NewButton

```go
gink.C(gink.NewButton("Save", func() { save() }))
```

Activates on Enter or Space. Bold cyan when focused.

### NewList

Scrollable selectable list.

```go
items := []string{"Alpha", "Beta", "Gamma"}
sel, setSel := gink.UseState(0)
gink.C(gink.NewList(items, sel, func(i int) { setSel(i) }, 8)) // 8 visible rows
```

Up/Down to navigate, Enter to select. Optional focus style as last argument.

### NewSelect

Single-line option picker cycled with Left/Right arrows.

```go
options := []string{"Small", "Medium", "Large"}
size, setSize := gink.UseState("Medium")
gink.C(gink.NewSelect(options, size, setSize))
```

### Table

Static bordered table with auto-sized columns.

```go
cols := []gink.Column{
    {Header: "Name", MinWidth: 12},
    {Header: "Status"},
    {Header: "Region", MaxWidth: 8},
}
rows := [][]string{
    {"web-01", "ONLINE", "us-east-1"},
}
gink.Table(cols, rows)
```

### NewTable

Interactive version of `Table` — focusable, with Up/Down row selection, viewport scroll, and horizontal scrolling when columns exceed the terminal width.

```go
sel, setSel := gink.UseState(0)
gink.C(gink.NewTable(cols, rows, sel, setSel, 10)) // 10 visible rows
```

Keys: Up/Down to move the row selection; Left/Right to scroll columns when the table is wider than the terminal. `◀`/`▶` corner indicators appear when there is hidden content in either direction.

### NewScrollView

A fixed-height scrollable viewport around any content. Up/Down to scroll when focused.

```go
gink.C(gink.NewScrollView(8, gink.Box(lines...)))
```

Scroll indicators (↑ ↓) appear when content extends beyond the viewport.

### ProgressBar

```go
gink.ProgressBar(0.72, 20) // 72% fill, 20 cols wide
gink.ProgressBar(0.72, 20, gink.NewStyle().Foreground(gink.ColorBrightGreen))
```

### Badge

```go
gink.Badge("ONLINE", gink.NewStyle().Foreground(gink.ColorBrightGreen))
// renders: [ ONLINE ]
```

### Image

Render an `image.Image` as a block of Unicode quadrant characters with true-colour foreground and background. Each terminal cell covers a 2×2 pixel region, giving double the resolution of plain half-block rendering.

```go
import (
    _ "image/jpeg"
    _ "image/png"
)

// width: output width in cells; height: output height in cells (0 = auto from aspect ratio)
gink.Image(img, 40, 20)
gink.Image(img, 40, 0) // auto-height
```

The source image is scaled with Catmull-Rom (bicubic) interpolation before rendering, so high-resolution sources produce sharp output. Fetch source images at a higher resolution than the output dimensions for best results — the scaler uses the extra detail for anti-aliasing.

```go
img, loading, err := gink.UseAsync(func() (image.Image, error) {
    resp, _ := http.Get("https://example.com/photo.jpg")
    defer resp.Body.Close()
    img, _, _ := image.Decode(resp.Body)
    return img, nil
}, []any{})

if !loading && err == nil {
    return gink.Image(img, 40, 20)
}
```

See `examples/imageview/` for a live demo that fetches random photos from [picsum.photos](https://picsum.photos) and refreshes on `R`.

---

## Scroll

Gink renders into a 512-row virtual buffer and clips it to the terminal height. When content is taller than the terminal:

- **PgUp / PgDn** — scroll the viewport by one screen height
- **Mouse wheel** — scroll 3 rows at a time
- **Tab / Shift+Tab** — auto-scroll to the newly focused component

Scroll indicators (↑ ↓) appear at the top-right and bottom-right corners of the screen.

Use `AppShell` to pin a footer below the scrollable area:

```go
return gink.AppShell(mainContent, footerBar)
```

---

## Testing

Gink ships a companion package, `ginktest`, for testing components without a real terminal.

```bash
go get github.com/SummaDiaboli/gink/ginktest
```

```go
import "github.com/SummaDiaboli/gink/ginktest"

func TestCounter_increment(t *testing.T) {
    h := ginktest.NewHarness(t, App)
    defer h.Close()

    h.Tab()   // move focus to Increment button
    h.Enter() // press it

    ginktest.AssertContains(t, h, "Count is 1")
}
```

**Harness methods:**

| Method | Description |
|---|---|
| `Tab()` / `ShiftTab()` | Move focus forward / backward |
| `Enter()` | Press Enter |
| `Backspace()` | Press Backspace |
| `SendRune(r)` | Type a printable character |
| `SendKey(key)` | Send a special key (arrow keys, Home, End, …) |
| `PageDown()` / `PageUp()` | Scroll viewport |
| `Click(x, y)` | Simulate a left mouse click at cell (x, y) |
| `Render()` | Force a re-render (for async polling) |
| `Contains(s)` | True if any line contains `s` |
| `Lines()` | Screen as `[]string`, one per row |
| `Line(y)` | Trimmed content of row `y` |
| `CellStyle(x, y)` | Raw `tcell.Style` at a cell |
| `AccessibilityLabel()` | Current accessibility label (set via `UseAccessibility`) |
| `Close()` | Release the simulation screen |

**Assertions:**

```go
ginktest.AssertContains(t, h, "text")
ginktest.AssertNotContains(t, h, "text")
ginktest.AssertLine(t, h, 0, "exact line text")          // exact match on row y
ginktest.AssertLineContains(t, h, 0, "partial text")     // substring match on row y
ginktest.AwaitContains(t, h, "text", 2*time.Second)      // poll until present
ginktest.AwaitNotContains(t, h, "text", 2*time.Second)   // poll until absent
```

You can also use `gink.NewHarness` directly from the main package for lower-level access.

---

## Plugins

Gink can embed UI written in other languages via subprocess plugins — any executable that speaks newline-delimited JSON on stdin/stdout. Client SDKs for Node.js and Python are provided in `sdk/`.

```go
gink.C(gink.NewPlugin("node", "my-plugin.js"))
gink.C(gink.NewPlugin("python", "my-plugin.py"))
```

See `sdk/README.md` for the full protocol specification.

---

## Architecture

```
┌──────────────────────────────────────────┐
│            User Components               │
│   Plain functions returning Elements     │
└─────────────────┬────────────────────────┘
                  │ Element tree
                  ▼
┌──────────────────────────────────────────┐
│              Reconciler                  │
│  Walks tree → calls components →         │
│  manages hook stores by tree path →      │
│  paints into a virtual Buffer            │
└─────────────────┬────────────────────────┘
                  │ Virtual buffer (512 rows)
                  ▼
┌──────────────────────────────────────────┐
│         Scroll + Compose layer           │
│  Clips virtual buffer to viewport →      │
│  overlays footer → adds indicators       │
└─────────────────┬────────────────────────┘
                  │ Screen buffer
                  ▼
┌──────────────────────────────────────────┐
│          Diff Engine (buffer.go)         │
│  Compares against previous buffer →      │
│  emits only changed cells               │
└─────────────────┬────────────────────────┘
                  │ Changed cells
                  ▼
┌──────────────────────────────────────────┐
│           Renderer (tcell)               │
│  Writes changed cells to the terminal    │
└──────────────────────────────────────────┘
```

**Hook identity** — Each component instance is identified by its position in the element tree (e.g. `"root/0/1/0"`). The reconciler keeps a `map[string]*renderContext` so each component has its own ordered hook slot array.

**Partial re-renders** — Components are cached after each render. On subsequent passes, a component is only re-rendered if its own state changed, a descendant's state changed, or its focus-within status changed. Unchanged subtrees are restored from cache in O(cells) time.

**Event loop** — `tcell.PollEvent` runs in a goroutine and sends events to a buffered channel. The main loop selects between terminal events and a `dirty` channel written by state setters, so async updates (timers, goroutines) trigger re-renders without waiting for a keypress.
