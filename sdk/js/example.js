'use strict';

/**
 * Example Gink plugin — a counter that increments on Enter.
 * Run from your Go app with:
 *
 *   gink.C(gink.NewPlugin("node", "example.js"))
 */

const gink = require('./gink');

let count = 0;

function render() {
  gink.element(
    gink.boxGap(1,
      gink.text('JS Plugin Counter', { bold: true, fg: 'brightCyan' }),
      gink.text(`Count: ${count}`, { bold: true, fg: 'brightYellow' }),
      gink.text('Press Enter to increment  ·  Esc to quit', { fg: 'white' }),
    ),
  );
}

gink.onMessage((msg) => {
  switch (msg.type) {
    case 'render':
      render();
      break;

    case 'input':
      if (msg.key === 'enter') {
        count++;
        render();
      }
      break;

    case 'unmount':
      process.exit(0);
  }
});
