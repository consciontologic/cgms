const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

test('retry loading works when CSP blocks inline event handlers', () => {
  const html = fs.readFileSync(path.join(__dirname, '../web/index.html'), 'utf8');
  assert.doesNotMatch(html, /\son[a-z]+\s*=/i, 'strict CSP must not report blocked inline handlers');
  const listeners = new Map();
  const retry = {hidden: true, addEventListener: (name, callback) => listeners.set(name, callback)};
  let reloads = 0;
  let showSlowLoading;
  const bootstrap = fs.readFileSync(path.join(__dirname, '../web/flutter_bootstrap.js'), 'utf8')
    .replace('{{flutter_js}}', '').replace('{{flutter_build_config}}', '');
  vm.runInNewContext(bootstrap, {
    document: {getElementById: id => id === 'retry' ? retry : {textContent: ''}},
    location: {reload: () => reloads++},
    setTimeout: callback => {showSlowLoading = callback; return 1;},
    clearTimeout: () => {},
    _flutter: {loader: {load: () => {}}},
  });
  showSlowLoading();
  assert.equal(retry.hidden, false);
  // Only listeners installed by the allowed external script can execute here.
  listeners.get('click')?.();
  assert.equal(reloads, 1, 'the visible recovery control must reload under strict CSP');
});
