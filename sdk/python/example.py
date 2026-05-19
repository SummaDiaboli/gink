"""
Example Gink plugin — a counter that increments on Enter.
Run from your Go app with:

    gink.C(gink.NewPlugin("python", "example.py"))
"""

import sys
import os

sys.path.insert(0, os.path.dirname(__file__))
import gink

count = 0


def show():
    gink.element(
        gink.box(
            gink.text('Python Plugin Counter', bold=True, fg='brightCyan'),
            gink.text(f'Count: {count}', bold=True, fg='brightYellow'),
            gink.text('Press Enter to increment  ·  Esc to quit', fg='white'),
            gap=1,
        )
    )


def handle(msg):
    global count
    if msg['type'] == 'render':
        show()
    elif msg['type'] == 'input' and msg.get('key') == 'enter':
        count += 1
        show()
    elif msg['type'] == 'unmount':
        sys.exit(0)


gink.on_message(handle)
gink.run()
