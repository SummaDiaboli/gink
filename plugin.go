package gink

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"sync"
)

var pluginErrorStyle = NewStyle().Foreground(ColorBrightRed)

// pluginConn guards subprocess stdin behind a mutex so the render goroutine
// (effect cleanup) and the event goroutine (UseInput) can write concurrently.
type pluginConn struct {
	mu    sync.Mutex
	stdin io.WriteCloser
}

func (c *pluginConn) send(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stdin != nil {
		c.stdin.Write([]byte(msg + "\n")) //nolint:errcheck
	}
}

func (c *pluginConn) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stdin != nil {
		c.stdin.Close()
		c.stdin = nil
	}
}

// NewPlugin returns a component that runs command as a subprocess plugin.
// The subprocess communicates with the host over newline-delimited JSON on
// stdin/stdout. See sdk/ for client libraries and the full protocol spec.
//
//	gink.C(gink.NewPlugin("node", "my-plugin.js"))
//	gink.C(gink.NewPlugin("python", "my-plugin.py"))
func NewPlugin(command string, args ...string) func() Element {
	return NewPluginCmd(exec.Command(command, args...))
}

// NewPluginCmd is like NewPlugin but accepts a pre-configured *exec.Cmd,
// allowing you to set Env, Dir, or other fields before the process starts.
//
//	cmd := exec.Command("node", "plugin.js")
//	cmd.Dir = "/path/to/plugin"
//	cmd.Env = append(os.Environ(), "API_KEY=secret")
//	gink.C(gink.NewPluginCmd(cmd))
func NewPluginCmd(cmd *exec.Cmd) func() Element {
	return func() Element {
		tree, setTree := UseState[*Element](nil)
		errMsg, setErr := UseState("")
		isFocused := UseFocus()
		conn := UseRef[*pluginConn](nil)

		// Mount once: start the process, run the reader goroutine, send the
		// initial render request. Cleanup sends unmount and waits for the process.
		UseEffect(func() func() {
			stdin, err := cmd.StdinPipe()
			if err != nil {
				setErr("stdin pipe: " + err.Error())
				return nil
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				setErr("stdout pipe: " + err.Error())
				return nil
			}
			if err := cmd.Start(); err != nil {
				setErr("start: " + err.Error())
				return nil
			}

			c := &pluginConn{stdin: stdin}
			conn.Value = c

			// Reader goroutine: parse element messages and push them into state.
			go func() {
				scanner := bufio.NewScanner(stdout)
				for scanner.Scan() {
					var env struct {
						Type string          `json:"type"`
						Tree json.RawMessage `json:"tree"`
					}
					if json.Unmarshal(scanner.Bytes(), &env) != nil {
						continue
					}
					if env.Type == "element" && len(env.Tree) > 0 {
						el := parsePluginElement(env.Tree)
						setTree(&el)
					}
				}
			}()

			c.send(`{"type":"render"}`)

			return func() {
				c.send(`{"type":"unmount"}`)
				c.close()
				cmd.Wait() //nolint:errcheck
			}
		}, []any{})

		// Forward keyboard input to the plugin when focused.
		UseInput(func(ev KeyEvent) {
			if !isFocused || conn.Value == nil {
				return
			}
			var msg []byte
			if ev.Key == KeyRune {
				msg, _ = json.Marshal(map[string]any{
					"type": "input",
					"key":  "rune",
					"rune": string(ev.Rune),
				})
			} else if name := pluginKeyName(ev); name != "" {
				msg, _ = json.Marshal(map[string]any{
					"type": "input",
					"key":  name,
				})
			}
			if msg != nil {
				conn.Value.send(string(msg))
			}
		})

		switch {
		case errMsg != "":
			return Text("⚠ plugin error: "+errMsg, pluginErrorStyle)
		case tree == nil:
			return C(Spinner) // show spinner while plugin initialises
		default:
			return *tree
		}
	}
}

// pluginKeyName maps a KeyEvent to its protocol wire name.
// Returns "" for keys that are not forwarded to plugins.
func pluginKeyName(ev KeyEvent) string {
	switch ev.Key {
	case KeyEnter:
		return "enter"
	case KeyBackspace, KeyBackspace2:
		return "backspace"
	case KeyEscape:
		return "escape"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyLeft:
		return "left"
	case KeyRight:
		return "right"
	default:
		return ""
	}
}

// ── Element JSON parsing ──────────────────────────────────────────────────────

type pluginElementJSON struct {
	Type     string            `json:"type"`
	Content  string            `json:"content"`
	Style    *pluginStyleJSON  `json:"style"`
	Gap      int               `json:"gap"`
	Children []json.RawMessage `json:"children"`
}

type pluginStyleJSON struct {
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	Underline bool   `json:"underline"`
	Fg        string `json:"fg"`
	Bg        string `json:"bg"`
}

func parsePluginElement(raw json.RawMessage) Element {
	var ej pluginElementJSON
	if err := json.Unmarshal(raw, &ej); err != nil {
		return Text("plugin parse error: "+err.Error(), pluginErrorStyle)
	}

	switch ej.Type {
	case "text":
		return Text(ej.Content, parsePluginStyle(ej.Style))

	case "box":
		children := make([]Element, len(ej.Children))
		for i, cr := range ej.Children {
			children[i] = parsePluginElement(cr)
		}
		return BoxWithGap(ej.Gap, children...)

	case "row":
		children := make([]Element, len(ej.Children))
		for i, cr := range ej.Children {
			children[i] = parsePluginElement(cr)
		}
		return RowWithGap(ej.Gap, children...)

	default:
		return Text("unknown plugin element: "+ej.Type, pluginErrorStyle)
	}
}

func parsePluginStyle(s *pluginStyleJSON) Style {
	if s == nil {
		return NewStyle()
	}
	style := NewStyle()
	if s.Bold {
		style = style.Bold()
	}
	if s.Italic {
		style = style.Italic()
	}
	if s.Underline {
		style = style.Underline()
	}
	if s.Fg != "" {
		style = style.Foreground(parsePluginColor(s.Fg))
	}
	if s.Bg != "" {
		style = style.Background(parsePluginColor(s.Bg))
	}
	return style
}

func parsePluginColor(name string) Color {
	switch name {
	case "black":
		return ColorBlack
	case "red":
		return ColorRed
	case "green":
		return ColorGreen
	case "yellow":
		return ColorYellow
	case "blue":
		return ColorBlue
	case "magenta":
		return ColorMagenta
	case "cyan":
		return ColorCyan
	case "white":
		return ColorWhite
	case "brightRed":
		return ColorBrightRed
	case "brightGreen":
		return ColorBrightGreen
	case "brightYellow":
		return ColorBrightYellow
	case "brightBlue":
		return ColorBrightBlue
	case "brightMagenta":
		return ColorBrightMagenta
	case "brightCyan":
		return ColorBrightCyan
	case "brightWhite":
		return ColorBrightWhite
	default:
		return ColorDefault
	}
}
