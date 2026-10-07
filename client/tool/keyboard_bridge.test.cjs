const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const source = fs.readFileSync(path.join(__dirname, '../web/keyboard_bridge.js'), 'utf8');

function bridge() {
  const tasks = [];
  let capture;
  const window = {addEventListener(type, listener, useCapture) {
    assert.equal(type, 'keydown'); assert.equal(useCapture, true); capture = listener;
  }};
  vm.runInNewContext(source, {window, setTimeout: task => tasks.push(task)});
  return {window, capture, flushOne: () => tasks.shift()(), flush: () => {while (tasks.length) tasks.shift()();}};
}
function event(overrides = {}) {
  return {key: 'Tab', eventPhase: 1, shiftKey: false, altKey: false, ctrlKey: false, metaKey: false,
    defaultPrevented: false, composedPath: () => [{tagName: 'INPUT'}, {tagName: 'FLUTTER-VIEW'}],
    preventDefault() {this.defaultPrevented = true;}, ...overrides};
}

test('handled forward and reverse traversal cancel native default synchronously', () => {
  for (const shiftKey of [false, true]) {
    const b = bridge(), key = event({shiftKey}); b.capture(key);
    b.window.cgmsCompleteTabTraversal(true);
    assert.equal(key.defaultPrevented, true);
  }
});
test('unhandled traversal can leave the Flutter view', () => {
  const b = bridge(), key = event(); b.capture(key);
  b.window.cgmsCompleteTabTraversal(false);
  assert.equal(key.defaultPrevented, false);
});
test('browser shortcuts and non-Flutter elements retain native keyboard behavior', () => {
  for (const overrides of [{ctrlKey: true}, {altKey: true}, {metaKey: true},
    {key: 'Enter'}, {composedPath: () => [{tagName: 'BUTTON'}, {tagName: 'BODY'}]}]) {
    const b = bridge(), key = event(overrides); b.capture(key);
    b.window.cgmsCompleteTabTraversal(true);
    assert.equal(key.defaultPrevented, false);
  }
});
test('late callbacks cannot cancel an expired or superseded event', () => {
  const b = bridge(), first = event(); b.capture(first); b.flush();
  b.window.cgmsCompleteTabTraversal(true); assert.equal(first.defaultPrevented, false);
  const second = event(), other = event({key: 'Enter'});
  b.capture(second); b.capture(other); b.window.cgmsCompleteTabTraversal(true);
  assert.equal(second.defaultPrevented, false); assert.equal(other.defaultPrevented, false);
});
test('an earlier cleanup cannot clear a newer active traversal event', () => {
  const b = bridge(), first = event(), second = event();
  b.capture(first); b.capture(second); b.flushOne(); b.window.cgmsCompleteTabTraversal(true);
  assert.equal(first.defaultPrevented, false); assert.equal(second.defaultPrevented, true);
});
test('finished dispatch cannot be cancelled before its cleanup task runs', () => {
  const b = bridge(), key = event(); b.capture(key); key.eventPhase = 0;
  b.window.cgmsCompleteTabTraversal(true);
  assert.equal(key.defaultPrevented, false);
});
