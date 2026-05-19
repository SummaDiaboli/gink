'use strict';

/**
 * Gink plugin SDK for Node.js.
 *
 * A Gink plugin is a subprocess that communicates with the host over
 * newline-delimited JSON (NDJSON) on stdin/stdout. This module handles
 * the wire protocol so you can focus on building your UI.
 *
 * @example
 * const gink = require('./gink');
 *
 * let count = 0;
 *
 * gink.onMessage((msg) => {
 *   if (msg.type === 'render' || (msg.type === 'input' && msg.key === 'enter')) {
 *     if (msg.type === 'input') count++;
 *     gink.element(gink.box(
 *       gink.text(`Count: ${count}`, { bold: true }),
 *       gink.text('Press Enter to increment'),
 *     ));
 *   }
 *   if (msg.type === 'unmount') process.exit(0);
 * });
 */

const readline = require('readline');

const rl = readline.createInterface({ input: process.stdin, terminal: false });

let _handler = null;

/** Register the function that handles all messages from the Gink host. */
function onMessage(fn) {
  _handler = fn;
}

/** Send a raw object to the host as a JSON line on stdout. */
function send(obj) {
  process.stdout.write(JSON.stringify(obj) + '\n');
}

/**
 * Send an element tree to the host. Call this in response to a 'render'
 * message and any time you want to push an update (e.g. from a timer).
 *
 * @param {object} tree - An element built with text(), box(), or row().
 */
function element(tree) {
  send({ type: 'element', tree });
}

// ── Element builders ──────────────────────────────────────────────────────────

/**
 * A single line of text with optional style.
 *
 * @param {string} content
 * @param {{ bold?: boolean, italic?: boolean, underline?: boolean, fg?: string, bg?: string }} [style]
 */
function text(content, style = {}) {
  const node = { type: 'text', content };
  if (Object.keys(style).length > 0) node.style = style;
  return node;
}

/**
 * Stack children vertically (column direction).
 *
 * @param {...object} children
 */
function box(...children) {
  return { type: 'box', children: children.flat() };
}

/**
 * Stack children vertically with gap empty rows between each child.
 *
 * @param {number} gap
 * @param {...object} children
 */
function boxGap(gap, ...children) {
  return { type: 'box', gap, children: children.flat() };
}

/**
 * Lay children out horizontally (row direction).
 *
 * @param {...object} children
 */
function row(...children) {
  return { type: 'row', children: children.flat() };
}

/**
 * Lay children out horizontally with gap empty columns between each child.
 *
 * @param {number} gap
 * @param {...object} children
 */
function rowGap(gap, ...children) {
  return { type: 'row', gap, children: children.flat() };
}

// ── Wire up stdin ─────────────────────────────────────────────────────────────

rl.on('line', (line) => {
  if (!line.trim()) return;
  let msg;
  try { msg = JSON.parse(line); } catch { return; }
  if (_handler) _handler(msg);
});

module.exports = { onMessage, send, element, text, box, boxGap, row, rowGap };
