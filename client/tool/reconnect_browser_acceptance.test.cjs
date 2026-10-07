const assert = require('node:assert/strict');
const test = require('node:test');
const {expectedOfflineConsole, classifyOfflineConsole, decisions, focusByKeyboard, gameOption, waitForCommandPresentation} = require('./reconnect_browser_acceptance.cjs');

function actionPage(selected = true, response = false, overlayOpen = false) {
  let active = null;
  const keys = [];
  const control = name => ({
    first() { return this; }, and() { return this; },
    count: async () => name === 'overlay' ? Number(overlayOpen) : name === 'card' ? Number(selected) : Number(response),
    waitFor: async () => assert.ok(name !== 'overlay' || overlayOpen),
    evaluateAll: async () => active === name,
    evaluate: async () => ({rect: {left: 8, top: 8, right: 108, bottom: 56, width: 100, height: 48}, clip: {left: 0, top: 0, right: 390, bottom: 844}, scroll: null}),
  });
  const page = {
    evaluate: async () => {}, viewportSize: () => ({width: 390, height: 844}),
    locator: name => {assert.equal(name, '[aria-current="true"]'); return {};},
    getByRole: (role, {name}) => {
      assert.equal(role, 'button');
      if (name === 'Close public cards') return {count: async () => 0};
      if (name instanceof RegExp && name.test('Close card context')) return control('overlay');
      if (name instanceof RegExp && name.test('4 of Clubs, copy 1')) return control('card');
      if (name instanceof RegExp && name.test('Pending decision')) return control('required');
      throw Error(`Unexpected permanent action navigation: ${name}`);
    },
    keyboard: {press: async key => {
      keys.push(key);
      if (key === 'Tab') active = selected ? 'card' : 'required';
      if (key === 'Enter' && active === 'card' || key === 'Space' && active === 'required') overlayOpen = true;
    }},
  };
  return {page, keys};
}

test('selected physical card enters contextual flow with Enter', async () => {
  const {page, keys} = actionPage(); await decisions(page);
  assert.deepEqual(keys, ['Tab', 'Enter']);
});
test('returning to open contextual review never toggles away its retained draft', async () => {
  const {page, keys} = actionPage(true, false, true); await decisions(page);
  assert.deepEqual(keys, []);
});
test('an automatically opened required decision remains pending without input', async () => {
  const {page, keys} = actionPage(false, true, true); await decisions(page);
  assert.deepEqual(keys, []);
});
test('dismissed context reopens through the same selected physical card', async () => {
  const {page, keys} = actionPage(true, false, false); await decisions(page);
  assert.deepEqual(keys, ['Tab', 'Enter']);
});
test('dismissed required decision reopens explicitly without answering it', async () => {
  const {page, keys} = actionPage(false, true, false); await decisions(page);
  assert.deepEqual(keys, ['Tab', 'Space']);
});

for (const enabled of [false, true]) {
  test(`financial option ${enabled ? 'retains a current review' : 'replaces a stale disabled review'} without confirming it`, async () => {
    let active = null, menuOpen = false, confirmed = enabled, busy = true;
    let releaseBusy;
    const appIdle = new Promise(resolve => {releaseBusy = () => {busy = false; resolve();};});
    const activations = [];
    const control = name => ({
      first() { return this; },
      count: async () => 1,
      isEnabled: async () => confirmed,
      click: async () => {activations.push([name, 'click']);},
      waitFor: async options => {if (name === 'option' && options?.state === 'hidden') assert.equal(menuOpen, false);},
      evaluateAll: async () => active === name,
      evaluate: async () => ({rect: {left: 8, top: 8, right: 108, bottom: 56, width: 100, height: 48}, clip: {left: 0, top: 0, right: 390, bottom: 844}, scroll: null}),
    });
    const page = {
      evaluate: async () => {},
      viewportSize: () => ({width: 390, height: 844}),
      getByRole: (role, {name} = {}) => {
        if (role === 'progressbar') return {waitFor: async options => {
          assert.deepEqual(options, {state: 'hidden', timeout: 15000});
          await appIdle;
        }};
        if (name === 'Close public cards') return {count: async () => 0};
        if (role === 'menuitem') {assert.equal(name, 'Finish Settlement'); return control('option');}
        if (name === 'Confirm Finish Settlement') return control('confirm');
        if (name === 'Clear selection' || name instanceof RegExp && name.test('Clear selection')) return control('clear');
        if (name instanceof RegExp && name.test('Close card context')) return control('overlay');
        if (name instanceof RegExp && name.test('Game options')) return control('menu');
        throw Error(`Unexpected control ${name}`);
      },
      keyboard: {press: async key => {
        if (key === 'Tab') {active = menuOpen ? 'option' : 'menu'; return;}
        activations.push([active, key]);
        if (active === 'clear') confirmed = false;
        if (active === 'menu') {assert.equal(busy, false, 'menu must snapshot ready command state'); menuOpen = true;}
        if (active === 'option') {menuOpen = false; confirmed = true;}
      }},
    };
    const pending = gameOption(page, 'Finish Settlement');
    await new Promise(resolve => setImmediate(resolve));
    assert.deepEqual(activations, [], 'no menu opens or confirmation occurs while presentation is busy');
    releaseBusy();
    await pending;
    assert.equal(confirmed, true);
    assert.deepEqual(activations, enabled ? [] : [['overlay', 'click'], ['menu', 'Space'], ['option', 'Enter']]);
  });
}

test('keyboard exposure fails after the existing bounded traversal instead of forcing DOM focus', async () => {
  const keys = [];
  const page = {evaluate: async () => {}, keyboard: {press: async key => keys.push(key)}};
  await assert.rejects(focusByKeyboard(page, {evaluateAll: async () => false}), /Keyboard traversal could not reach/);
  assert.equal(keys.length, 80);
  assert.ok(keys.every(key => key === 'Tab'));
});

test('offline console exceptions require active phase, exact failed GET URL and exact network message', () => {
  const url = 'http://127.0.0.1:8081/v1/matches/public-match';
  const chromium = 'Failed to load resource: net::ERR_INTERNET_DISCONNECTED';
  const failure = 'net::ERR_INTERNET_DISCONNECTED';
  assert.equal(expectedOfflineConsole(true, chromium, url, url, failure), true);
  assert.equal(expectedOfflineConsole(false, chromium, url, url, failure), false);
  assert.equal(expectedOfflineConsole(true, chromium, `${url}/commands`, url, failure), false);
  assert.equal(expectedOfflineConsole(true, chromium, url, url, undefined), false);
  assert.equal(expectedOfflineConsole(true, chromium, url, url, 'net::ERR_FAILED'), false);
  assert.equal(expectedOfflineConsole(true, 'Failed to load resource: net::ERR_FAILED', url, url, failure), false);
  assert.equal(expectedOfflineConsole(true, 'Exception: render overflowed', url, url, failure), false);
  assert.equal(expectedOfflineConsole(true, 'Failed to load resource: the server responded with a status of 500', url, url, failure), false);
});

test('WebKit internal network diagnostic requires the same actually failed GET in offline phase', () => {
  const url = 'http://127.0.0.1:8081/v1/matches/public-match';
  const failure = 'WebKit encountered an internal error';
  const message = `Failed to load resource: ${failure}`;
  assert.equal(expectedOfflineConsole(true, message, url, url, failure), true);
  assert.equal(expectedOfflineConsole(false, message, url, url, failure), false);
  assert.equal(expectedOfflineConsole(true, message, url, url, undefined), false);
  assert.equal(expectedOfflineConsole(true, message, `${url}/stream`, url, failure), false);
  assert.equal(expectedOfflineConsole(true, message, url, url, 'net::ERR_FAILED'), false);
});


test('bounded recovery console diagnostics consume distinct observed failed GETs exactly once', () => {
  const url = 'http://127.0.0.1:8081/v1/matches/public-match';
  const failures = Array.from({length: 8}, () => ({url, failure: 'net::ERR_INTERNET_DISCONNECTED'}));
  const messages = Array.from({length: 9}, () => ({phase: 'offline', message: 'Failed to load resource: net::ERR_INTERNET_DISCONNECTED', location: {url}}));
  const result = classifyOfflineConsole(messages, failures, url);
  assert.equal(result.expected.length, 8);
  assert.deepEqual(result.unexpected, [messages[8]]);
  const wrong = classifyOfflineConsole([{...messages[0], phase: 'online'}, {...messages[0], message: 'Unhandled exception'}], failures, url);
  assert.equal(wrong.expected.length, 0);
  assert.equal(wrong.unexpected.length, 2);
});

test('next command waits for its full reply and visible app busy completion before opening menus', async () => {
  const events = [];
  let releaseResponse, releaseBusy;
  const responseComplete = new Promise(resolve => {releaseResponse = resolve;});
  const appIdle = new Promise(resolve => {releaseBusy = resolve;});
  const reply = {finished: async () => {events.push('response-complete'); return responseComplete;}};
  const page = {
    evaluate: async () => {events.push('frame');},
    getByRole: role => {
      assert.equal(role, 'progressbar');
      return {waitFor: async options => {
        assert.deepEqual(options, {state: 'hidden', timeout: 15000});
        events.push('app-idle'); await appIdle;
      }};
    },
  };
  let complete = false;
  const pending = waitForCommandPresentation(page, reply).then(() => {complete = true;});
  assert.deepEqual(events, ['response-complete']);
  releaseResponse(null);
  await new Promise(resolve => setImmediate(resolve));
  assert.deepEqual(events, ['response-complete', 'frame', 'app-idle']);
  assert.equal(complete, false, 'authority receipt alone cannot freeze a disabled menu');
  releaseBusy(); await pending;
  assert.deepEqual(events, ['response-complete', 'frame', 'app-idle', 'frame']);
  assert.equal(complete, true);
});

test('incomplete command transport fails without waiting for presentation or resubmitting', async () => {
  const page = {evaluate: async () => assert.fail('must not continue after response failure')};
  await assert.rejects(waitForCommandPresentation(page, {finished: async () => new Error('connection closed')}), /full authority reply received/);
});
