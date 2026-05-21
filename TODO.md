# Gink — Outstanding Work

## Layout Primitives

---

## Hooks

---

## Nice to Have

- `Image` improvements:
  - Sixel protocol (Windows Terminal, WezTerm, xterm) for true raster output at full terminal resolution
  - Kitty graphics protocol (Kitty, WezTerm)
  - iTerm2 inline images (iTerm2, WezTerm on macOS)
  - Auto-detect terminal protocol support via `$TERM_PROGRAM`, `$KITTY_WINDOW_ID`, `$WT_SESSION` env vars
  - Bilinear interpolation option alongside Catmull-Rom (faster for large images)
  - `LoadImage(path string)` / `LoadImageURL(url string)` convenience helpers
- Video rendering — play video frames in the terminal at a target framerate, using the same protocol stack as image rendering

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
- `TextWrapped` — word-wrap with hard-break, newline support, and optional style ✓
- `Width` / `Height` / `Size` — exact-dimension wrappers over Constrain ✓
- `NewTextArea` — multi-line input with cursor, Up/Down/Home/End, viewport scroll, line split/merge ✓
- `Style.Reverse()` — reverse video for cursor rendering ✓
- `KeyHome` / `KeyEnd` constants ✓
- Mouse click support — `UseClick` hook, focus transfer on left-click, exact hit-testing ✓
- Click-to-select on `NewList`, `NewTable`, `NewButton`, `NewSelect` ✓
- `UseKeyboard` — global key handler independent of focus ✓
- Color themes — `Theme` struct, `ThemeCtx`, `UseTheme()`, built-in components read from theme ✓
- `ginktest.AssertLine` / `AssertLineContains` — line-specific assertion helpers ✓
- `NewTable` horizontal scrolling — Left/Right shifts column viewport, ◀/▶ border indicators ✓
- `UseAccessibility` — registers screen-reader label; exposed via terminal title and `Harness.AccessibilityLabel()` ✓
- `Image` — quadrant-block true-colour renderer (2×2px/cell), Catmull-Rom scaling, auto-height, picsum example, `NewRGBColor` helper ✓
- `NewCheckbox` — toggle rendered as `[ ]/[x]`; Space/Enter/click toggles; theme focus highlight ✓
- `NewRadioGroup` — `( )/(●)` vertical option list; Up/Down + Enter/click selects ✓
- `NewTabs` — horizontal tab strip with Left/Right nav and active-tab highlight ✓
- `NewMultiSelect` — scrollable `[ ]/[x]` list; Space toggles; Up/Down navigates ✓
- `UseToast` — transient notification hook with auto-dismiss timer ✓
- `NewMenu` — bordered key-hint menu; Up/Down + Enter selects; Esc closes; disabled items skipped ✓
- `NewTree` — hierarchical `▶/▼` expand/collapse list; Right/Left/Up/Down/Enter navigation ✓
- `NewModal` — focus-trapping dialog with title, content, and action buttons; Esc closes ✓
- `UseFocusBarrier` — traps Tab/Shift+Tab within a component subtree (used by `NewModal`) ✓
