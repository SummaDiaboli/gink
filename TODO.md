# Gink — Outstanding Work

## Layout Primitives

### Min/Max Sizing (global layout)
Constrain element dimensions at the layout level.
- `Width`, `Height` fields on element options (0 = auto)

### Text Wrapping
Wrap long text at a given column width rather than clipping.
- `TextWrapped(s string, maxWidth int, styles ...Style) Element`
- Reconciler splits content into multiple rows

---

## Component Library


### TextArea
Multi-line text input (NewInput is single-line only).
- `NewTextArea(value string, onChange func(string)) func() Element`
- Enter inserts a newline, Ctrl+Enter or a Submit button submits
- Fixed or auto-sizing height

### Table (interactive extensions)
- Horizontal scrolling for wide tables

---

## Hooks

---

## Nice to Have

- Mouse click support (tcell delivers mouse events; need hit-testing)
- Accessibility: screen reader hints via terminal title or alt text conventions

---

## Done

- Partial re-renders (dirty subtree tracking) ✓
- `ginktest` sub-package with public Harness API ✓
- Example apps: counter, todo, stopwatch, form ✓
- Example tests (full coverage) ✓
- Subprocess/JSON plugin system (`NewPlugin`, `NewPluginCmd`) ✓
- JS SDK (`sdk/js/gink.js`) ✓
- Python SDK (`sdk/python/gink.py`) ✓
- Divider / DividerWithLabel / DividerStyled ✓
- Padding / PaddingAll / PaddingXY ✓
- Styled examples (dividers, padding, consistent headers) ✓
- `UseAsync[T]` — async fn with loading/error state, stale-result cancellation ✓
- `ProgressBar` — fill bar with percentage label, optional style ✓
- `Badge` — inline `[ label ]` with optional style ✓
- `NewSelect` — single-line controlled option picker ✓
- `NewList` — scrollable viewport list with selection ✓
- `Border` / `BorderWithTitle` — line-drawing panel borders ✓
- `UseContext` / `SetContext` / `NewContext` — global shared state ✓
- Dashboard example — Clock, ServerList, MetricsPanel with all hooks ✓
- `Table` — bordered table with auto column widths, `MinWidth`/`MaxWidth` per column, truncation with `…` ✓
- `NewSelect` navigation changed to Left/Right arrows ✓
- `NewTable` — interactive table with row selection, Up/Down navigation, viewport scroll, focus highlight ✓
- Customisable focus highlight style on `NewList`, `NewTable`, `NewSelect` ✓
- `Constrain` / `MinWidth` / `MaxWidth` / `MinHeight` / `MaxHeight` — global layout constraints ✓
- Global scroll — virtual 512-row buffer, PgUp/PgDn, mouse wheel, scroll indicators ✓
- `AppShell` — sticky footer pinned outside the scroll viewport ✓
- Tab auto-scroll — scrolls to show focused component accounting for footer and component height ✓
- `NewScrollView` — scoped scroll region with Up/Down navigation, scroll indicators, fixed viewport height ✓
- `UseFocusWithin` — returns true when any descendant holds focus, for styling container borders ✓
