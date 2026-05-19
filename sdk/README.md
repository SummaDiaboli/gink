# Gink Plugin SDK

Gink plugins are subprocesses that render UI by exchanging newline-delimited
JSON (NDJSON) with the Go host over stdin/stdout. Any language that can read
stdin and write stdout can be a Gink plugin.

## Quick start

**Node.js**
```js
const gink = require('./gink');

gink.onMessage((msg) => {
  if (msg.type === 'render') {
    gink.element(gink.text('Hello from JS!'));
  }
  if (msg.type === 'unmount') process.exit(0);
});
```

**Python**
```python
import gink

def handle(msg):
    if msg['type'] == 'render':
        gink.element(gink.text('Hello from Python!'))
    elif msg['type'] == 'unmount':
        import sys; sys.exit(0)

gink.on_message(handle)
gink.run()
```

Wire it up in Go:
```go
gink.C(gink.NewPlugin("node", "my-plugin.js"))
gink.C(gink.NewPlugin("python", "my-plugin.py"))
```

---

## Protocol

All messages are JSON objects on a single line (`\n` terminated). Both
directions use the same framing.

### Host → Plugin

| `type`    | Fields                          | Description                                      |
|-----------|---------------------------------|--------------------------------------------------|
| `render`  | —                               | Sent once on mount. Plugin should respond with an element. |
| `input`   | `key`, `rune` (when key=`rune`) | A key event forwarded from the terminal.         |
| `unmount` | —                               | Plugin should clean up and exit.                 |

#### Input key names

| `key`       | Meaning                     |
|-------------|-----------------------------|
| `rune`      | Printable character — read `rune` field for the character |
| `enter`     | Enter / Return              |
| `backspace` | Backspace                   |
| `escape`    | Escape                      |
| `up`        | Up arrow                    |
| `down`      | Down arrow                  |
| `left`      | Left arrow                  |
| `right`     | Right arrow                 |

> The plugin only receives input when it holds focus. Tab/Shift-Tab are
> consumed by the Gink focus system and are never forwarded.

### Plugin → Host

| `type`    | Fields | Description                                        |
|-----------|--------|----------------------------------------------------|
| `element` | `tree` | A new element tree to render. Can be sent any time — not just in response to `render`. |

Sending `element` at any time triggers a re-render. Use this for timers,
async operations, or anything that produces UI updates independently of input.

---

## Element JSON

Elements mirror the Gink Go API:

```json
{ "type": "text", "content": "Hello", "style": { "bold": true, "fg": "brightCyan" } }

{ "type": "box", "gap": 1, "children": [ ... ] }

{ "type": "row", "gap": 2, "children": [ ... ] }
```

### Style fields

| Field       | Type    | Description                      |
|-------------|---------|----------------------------------|
| `bold`      | boolean | Bold text                        |
| `italic`    | boolean | Italic text                      |
| `underline` | boolean | Underlined text                  |
| `fg`        | string  | Foreground color (see below)     |
| `bg`        | string  | Background color (see below)     |

### Color names

`black` · `red` · `green` · `yellow` · `blue` · `magenta` · `cyan` · `white`

`brightRed` · `brightGreen` · `brightYellow` · `brightBlue` · `brightMagenta` · `brightCyan` · `brightWhite`

---

## Pushing updates from a timer

Plugins can push element updates at any time — not just in response to
`render` or `input`. Use this for real-time UIs:

**Python**
```python
import gink, threading, time

elapsed = 0

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

**Node.js**
```js
const gink = require('./gink');
let elapsed = 0;

gink.onMessage((msg) => {
  if (msg.type === 'render') {
    setInterval(() => {
      elapsed += 0.1;
      gink.element(gink.text(`Elapsed: ${elapsed.toFixed(1)}s`));
    }, 100);
    gink.element(gink.text('Starting...'));
  }
  if (msg.type === 'unmount') process.exit(0);
});
```
