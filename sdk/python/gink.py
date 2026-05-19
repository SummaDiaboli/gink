"""
Gink plugin SDK for Python.

A Gink plugin is a subprocess that communicates with the host over
newline-delimited JSON (NDJSON) on stdin/stdout. This module handles
the wire protocol so you can focus on building your UI.

Example
-------
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
        ))

    gink.on_message(handle)
    gink.run()
"""

import json
import sys

_handler = None


def on_message(fn):
    """Register the function that handles all messages from the Gink host."""
    global _handler
    _handler = fn


def send(obj):
    """Send a raw dict to the host as a JSON line on stdout."""
    sys.stdout.write(json.dumps(obj) + '\n')
    sys.stdout.flush()


def element(tree):
    """
    Send an element tree to the host. Call this in response to a 'render'
    message and any time you want to push an update (e.g. from a timer thread).
    """
    send({'type': 'element', 'tree': tree})


# ── Element builders ──────────────────────────────────────────────────────────

def text(content, *, bold=False, italic=False, underline=False, fg='', bg=''):
    """
    A single line of text with optional style.

    Color names: black, red, green, yellow, blue, magenta, cyan, white,
    brightRed, brightGreen, brightYellow, brightBlue, brightMagenta,
    brightCyan, brightWhite.
    """
    node = {'type': 'text', 'content': content}
    style = {}
    if bold:      style['bold'] = True
    if italic:    style['italic'] = True
    if underline: style['underline'] = True
    if fg:        style['fg'] = fg
    if bg:        style['bg'] = bg
    if style:
        node['style'] = style
    return node


def box(*children, gap=0):
    """Stack children vertically (column direction), with optional row gap."""
    return {'type': 'box', 'gap': gap, 'children': list(children)}


def row(*children, gap=0):
    """Lay children out horizontally (row direction), with optional column gap."""
    return {'type': 'row', 'gap': gap, 'children': list(children)}


# ── Event loop ────────────────────────────────────────────────────────────────

def run():
    """
    Start the plugin event loop. Reads NDJSON from stdin and dispatches each
    message to the handler registered with on_message(). Blocks until stdin
    closes. Always call this last, after registering your handler.
    """
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        try:
            msg = json.loads(line)
        except json.JSONDecodeError:
            continue
        if _handler:
            _handler(msg)
