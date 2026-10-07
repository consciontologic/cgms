// Supplemental real-authority reconnect check; deliberately separate from the
// frozen main browser runner. This emulates browser HTTP loss plus a closed
// transparent WebSocket proxy, not an operating-system network partition.
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const {isBrowserError} = require('./browser_faults.cjs');

const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
const frame = page => page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
const sha256 = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
const scope = snapshot => Object.fromEntries(['game_id', 'phase', 'active', 'window_id', 'decision_id', 'required_actor', 'decision_kind'].map(key => [key, snapshot.projection.board[key] ?? null]));
const disconnectedTitle = 'Connection lost · last confirmed table · actions paused';
const relativeHandCardGeometry = target => target.evaluate(element => {
  const hand=element.closest('[flt-semantics-identifier="private-hand-viewport"]');
  if(!hand) throw Error('Physical card is not in the authorized private hand');
  const card=element.getBoundingClientRect(),parent=hand.getBoundingClientRect();
  return {x:card.x-parent.x,y:card.y-parent.y,width:card.width,height:card.height};
});
const ariaStates = target => target.evaluate(element => Object.fromEntries(element.getAttributeNames().filter(name => name.startsWith('aria-')).map(name => [name, element.getAttribute(name)])));

// Exact diagnostics and exact failed-GET URL, only during the explicit offline
// phase. No application exceptions, generic aborts or HTTP status errors pass.
function expectedOfflineConsole(active, message, url, matchUrl, failedGetError) {
  return active && url === matchUrl && [
    'net::ERR_INTERNET_DISCONNECTED',
    'WebKit encountered an internal error',
  ].includes(failedGetError) && message === `Failed to load resource: ${failedGetError}`;
}

function classifyOfflineConsole(messages, failures, matchUrl) {
  const unused = [...failures], expected = [], unexpected = [];
  for (const item of messages) {
    const index = unused.findIndex(failed => expectedOfflineConsole(item.phase === 'offline', item.message, item.location.url, matchUrl, failed.failure) && failed.url === item.location.url);
    if (index < 0) unexpected.push(item);
    else {
      const [failed] = unused.splice(index, 1);
      expected.push({...item, failedGetError: failed.failure});
    }
  }
  return {expected, unexpected};
}

async function traceGeometry(page, target, phase) {
  if (!process.env.CGMS_RECONNECT_TRACE) return;
  const geometry=await target.evaluate(element=>{
    const ancestors=[];
    for(let node=element;node;node=node.parentElement) {
      const style=getComputedStyle(node);
      ancestors.push({id:node.id,tag:node.tagName,identifier:node.getAttribute('flt-semantics-identifier'),role:node.getAttribute('role'),ariaLabel:node.getAttribute('aria-label'),ariaHidden:node.getAttribute('aria-hidden'),tabIndex:node.getAttribute('tabindex'),childIds:[...node.children].map(child=>child.id||child.tagName),rect:node.getBoundingClientRect().toJSON(),transform:style.transform,position:style.position,top:style.top,left:style.left,scrollTop:node.scrollTop,scrollLeft:node.scrollLeft,scrollHeight:node.scrollHeight,clientHeight:node.clientHeight,overflow:style.overflow,focused:node===document.activeElement});
    }
    return ancestors;
  });
  fs.appendFileSync(path.join(process.env.CGMS_EVIDENCE_DIR,'geometry-trace.jsonl'),JSON.stringify({phase,geometry})+'\n');
  await page.screenshot({path:path.join(process.env.CGMS_EVIDENCE_DIR,`geometry-${phase}.png`)});
}
async function focusByKeyboard(page, target) {
  for (let step = 0; step < 80; step++) {
    if (await target.evaluateAll(elements => elements.some(element => element === document.activeElement))) return;
    await page.keyboard.press('Tab'); await frame(page);
  }
  throw Error('Keyboard traversal could not reach requested control');
}
async function section(page, target) {
  await target.waitFor({state: 'attached', timeout: 30000});
  const viewport = page.viewportSize(), unit = 1 / 64;
  let lastGeometry; const geometryHistory=[];
  for (let step = 0; step < 24; step++) {
    const geometry = await target.evaluate(element => {
      const rect = element.getBoundingClientRect().toJSON();
      const clip = {left: 0, top: 0, right: innerWidth, bottom: innerHeight};
      let scroll; const clippingAncestors=[];
      for (let parent = element.parentElement; parent; parent = parent.parentElement) {
        const style = getComputedStyle(parent), box = parent.getBoundingClientRect();
        if (!box.width || !box.height) continue;
        if (/auto|scroll|hidden|clip/.test(style.overflowX + style.overflowY)) clippingAncestors.push({tag:parent.tagName,id:parent.id,role:parent.getAttribute('role'),rect:box.toJSON(),overflowX:style.overflowX,overflowY:style.overflowY,scrollTop:parent.scrollTop,scrollHeight:parent.scrollHeight,clientHeight:parent.clientHeight,transform:style.transform});
        if (/auto|scroll|hidden|clip/.test(style.overflowY)) {
          clip.top = Math.max(clip.top, box.top); clip.bottom = Math.min(clip.bottom, box.bottom);
          // Scroll the ancestor that actually clips this card, not an inner
          // hand scroller whose content is already fully exposed within it.
          if (!scroll && parent.getAttribute('role') === 'group' &&
              (rect.top < Math.max(0, box.top) - 1 / 64 ||
               rect.bottom > Math.min(innerHeight, box.bottom) + 1 / 64)) scroll = box.toJSON();
        }
        if (/auto|scroll|hidden|clip/.test(style.overflowX)) {
          clip.left = Math.max(clip.left, box.left); clip.right = Math.min(clip.right, box.right);
        }
      }
      return {rect, clip, scroll, clippingAncestors, focused:document.activeElement===element, activeTag:document.activeElement?.tagName, activeId:document.activeElement?.id};
    });
    lastGeometry=geometry; if(step<3||step===23)geometryHistory.push(geometry);
    const {rect, clip, scroll} = geometry;
    if (rect.width > 0 && rect.height > 0 && rect.top >= clip.top - unit && rect.bottom <= clip.bottom + unit && rect.left >= clip.left - unit && rect.right <= clip.right + unit) return;
    if (!scroll) throw Error(`Control has no painted scrolling path: ${JSON.stringify(geometry)}`);
    const top = Math.max(0, scroll.top), bottom = Math.min(viewport.height, scroll.bottom);
    await page.mouse.move(Math.max(1, Math.min(viewport.width - 1, scroll.left + scroll.width / 2)), (top + bottom) / 2);
    await page.mouse.wheel(0, rect.top + rect.height / 2 - (top + bottom) / 2);
    await pause(150); await frame(page);
  }
  throw Error(`Control could not be exposed by bounded wheel input: ${JSON.stringify({lastGeometry,geometryHistory})}`);
}
async function activate(page, target, key = 'Space') {
  await section(page, target); await focusByKeyboard(page, target); await frame(page);
  await page.keyboard.press(key); await frame(page);
}
async function closeActions(page) {
  const close = page.getByRole('button', {name: /^Close card context(?: Close card context)?$/});
  if (!(await close.count())) return;
  if(process.env.CGMS_RECONNECT_DISMISS==='escape') await page.keyboard.press('Escape');
  else await close.click();
  await close.waitFor({state: 'hidden'}); await frame(page);
}
async function cancelDraft(page) {
  await page.keyboard.press('Escape'); await frame(page);
}
async function closePublicCards(page) {
  const close = page.getByRole('button', {name: 'Close public cards', exact: true});
  if (!(await close.count())) return;
  await focusByKeyboard(page, close);
  await page.keyboard.press('Space');
  await close.waitFor({state: 'hidden'});
  await frame(page);
}
async function hand(page) {
  await closePublicCards(page);
  await closeActions(page);
  // The scrollable hand container can be taller than its clipped viewport.
  // Reveal an actual physical card, keeping any existing selection unchanged.
  const cards = page.locator('[flt-semantics-identifier="private-hand-viewport"]')
    .getByRole('button', {name: /^(?:[AKQJ]|[2-9]|10) of /});
  if (await cards.count()) {
    const selected = cards.and(page.locator('[aria-current="true"]'));
    const target = (await selected.count() ? selected : cards).first();
    await traceGeometry(page,target,'dismissed-before-focus');
    await focusByKeyboard(page, target); await pause(150); await frame(page);
    await traceGeometry(page,target,'focused-before-scroll'); await section(page,target);
  }
}
async function decisions(page) {
  await closePublicCards(page);
  const context = page.getByRole('button', {name: /^Close card context(?: Close card context)?$/});
  if (await context.count()) return;
  const selected = page.getByRole('button', {name: /^(?:[AKQJ]|[2-9]|10) of /})
    .and(page.locator('[aria-current="true"]')).first();
  if (await selected.count()) {
    await focusByKeyboard(page, selected);
    await page.keyboard.press('Enter'); await frame(page);
    await context.waitFor(); return;
  }
  const required = page.getByRole('button', {name: /^(?:Pending decision|Results & settlement)$/});
  if (await required.count()) {await activate(page, required); await context.waitFor();}
}
async function revealCard(page, pattern) {
  const target = page.getByRole('button', {name: pattern});
  if (await target.count()) {
    await focusByKeyboard(page, target);
    return target;
  }
  await closePublicCards(page);
  await closeActions(page);
  const groups = page.getByRole('button', {name: 'Inspect public cards', exact: true});
  const count = await groups.count();
  for (let index = 0; index < count; index++) {
    const group = groups.nth(index);
    await focusByKeyboard(page, group);
    await page.keyboard.press('Space');
    await page.getByRole('button', {name: 'Close public cards', exact: true}).waitFor();
    if (await target.count()) {
      await focusByKeyboard(page, target);
      return target;
    }
    await closePublicCards(page);
  }
  throw Error(`Authorized physical card is not reachable through public seats: ${pattern}`);
}
async function selectIntent(page, name) {
  await decisions(page);
  const confirm = page.getByRole('button', {name: `Confirm ${name}`, exact: true});
  if (await confirm.count()) return;
  const renderedName = name === 'Offer' ? 'Trade cards' : name;
  const escaped = renderedName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const option = page.getByRole('button', {name: new RegExp(`^${escaped}(?: · .*)?$`)}).first();
  await option.waitFor(); await activate(page, option);
  await confirm.waitFor();
}
// Receiving a command response can precede OnlineClient's refresh and the
// Flutter frame that clears busy. PopupMenuItem.enabled is captured when the
// menu opens, so wait for presentation to settle before creating that menu.
// This only observes delivery/readiness; it never retries an operation.
async function waitForCommandPresentation(page, reply) {
  if (reply) assert.equal(await reply.finished(), null, 'full authority reply received');
  await frame(page);
  await page.getByRole('progressbar').waitFor({state: 'hidden', timeout: 15000});
  await frame(page);
}
async function gameOption(page, name) {
  await waitForCommandPresentation(page);
  const existing = page.getByRole('button', {name: `Confirm ${name}`, exact: true});
  if (await existing.count() && await existing.isEnabled()) return;
  if (name === 'Offer') {await selectIntent(page, name); return;}
  await closePublicCards(page); await closeActions(page);
  const menu = page.getByRole('button', {name: /Game options/});
  await activate(page, menu);
  const option = page.getByRole('menuitem', {name, exact: true});
  await option.waitFor(); await pause(350);
  await focusByKeyboard(page, option); await page.keyboard.press('Enter');
  await option.waitFor({state: 'hidden'}); await pause(350); await frame(page);
  await page.getByRole('button', {name: `Confirm ${name}`, exact: true}).waitFor();
}

async function run() {
  if (!process.env.CGMS_CONTAINER) process.env.PLAYWRIGHT_BROWSERS_PATH = path.resolve(__dirname, '../.browser-tools/browsers');
  const engines = require('../.browser-tools/node_modules/playwright');
  const engine = process.env.CGMS_BROWSER || 'chromium';
  const origin = process.env.CGMS_BROWSER_ORIGIN || 'http://127.0.0.1:8081';
  const runId = `${engine.slice(0, 7)}-reconnect-${Date.now()}`;
  const evidence = path.resolve(process.env.CGMS_EVIDENCE_DIR || path.join(__dirname, '../.browser-tools/artifacts', runId));
  fs.mkdirSync(evidence, {recursive: true});
  const report = {
    engine, run: runId, status: 'running', startedAt: new Date().toISOString(),
    viewport: {width: 390, height: 844}, players: 4,
    fixture: 'Authorized purchase and Justice decision positions on real local QA authority',
    interruption: 'Playwright context.setOffline plus explicit closure of transparent real-backend WebSocket proxy; no OS network partition',
    diagnostics: 'All page exceptions fail; exact expected failed-GET console diagnostics allowed only inside offline phase',
    assetSha256: sha256(path.resolve(__dirname, '../build/web/main.dart.js')),
    harnessSha256: {'reconnect_browser_acceptance.cjs': sha256(__filename), 'browser_faults.cjs': sha256(path.join(__dirname, 'browser_faults.cjs'))},
    cases: [], errors: [], expectedFaults: [], screenshots: [], captureErrors: [],
  };
  const browser = await engines[engine].launch({headless: true});
  report.version = browser.version();
  try {
    for (const scenario of ((process.env.CGMS_RECONNECT_FOCUS_ONLY||process.env.CGMS_RECONNECT_DISMISS) ? ['purchase'] : ['purchase', 'justice'])) {
      const context = await browser.newContext({viewport: report.viewport, hasTouch: true});
      const response = await context.request.post(`${origin}/__fixture/session`, {
        headers: {Origin: origin}, data: {scenario, players: 4, seat: 1, run: runId},
      });
      assert.equal(response.status(), 200, `${scenario} fixture session`);
      const {match_id: matchId} = await response.json();
      const matchUrl = `${origin}/v1/matches/${encodeURIComponent(matchId)}`;
      const streams = [];
      let offline = false;
      await context.routeWebSocket(url => url.origin === origin.replace(/^http/, 'ws') && url.pathname === `/v1/matches/${encodeURIComponent(matchId)}/stream`, route => {
        if (offline) {void route.close({code: 1012, reason: 'offline acceptance phase'}); return;}
        // No frame handler is installed: Playwright forwards real server data
        // unchanged in both directions until this test closes the transport.
        streams.push({page: route, server: route.connectToServer()});
      });
      const page = await context.newPage();
      const mutations = [], offlineFailures = [], offlineConsole = [], snapshotResponses = [];
      page.on('pageerror', error => report.errors.push({scenario, phase: offline ? 'offline' : 'online', message: error.message}));
      page.on('console', message => {
        if (!isBrowserError(message.type(), message.text())) return;
        const item = {scenario, message: message.text(), location: message.location(), phase: offline ? 'offline' : 'online'};
        if (offline) offlineConsole.push(item);
        else report.errors.push({...item, phase: 'online'});
      });
      page.on('request', request => {if (request.method() === 'POST') mutations.push({path: new URL(request.url()).pathname});});
      page.on('requestfailed', request => {
        const item = {method: request.method(), url: request.url(), failure: request.failure()?.errorText};
        if (offline && request.method() === 'GET' && request.url() === matchUrl) offlineFailures.push(item);
        else report.errors.push({scenario, unexpectedRequestFailure: item, phase: offline ? 'offline' : 'online'});
      });
      page.on('response', response => {
        if (response.url() === matchUrl && response.request().method() === 'GET') snapshotResponses.push({status: response.status(), phase: offline ? 'offline' : 'online'});
      });
      const capture = async name => {
        await frame(page);
        const file = `${engine}-390-reconnect-${scenario}-${name}.png`;
        await page.screenshot({path: path.join(evidence, file)}); report.screenshots.push(file);
      };
      const snapshot = async () => {
        const result = await context.request.get(matchUrl);
        assert.equal(result.status(), 200, 'authority snapshot'); return result.json();
      };
      await page.goto(`${origin}/#/online?match=${encodeURIComponent(matchId)}`);
      await page.getByRole('heading', {name: 'CGMS · Online', exact: true}).waitFor({timeout: 30000});
      await page.locator('flt-semantics-host').getByText(/Seat 1’s turn|Waiting for seat/).first().waitFor({timeout: 30000});
      await page.waitForLoadState('networkidle');
      const selectedAction = scenario === 'purchase' ? 'Purchase' : 'Decision';
      const originalCardName = scenario === 'purchase' ? /^10 of Spades, copy 1/ : /^4 of Hearts, copy 1/;
      await closeActions(page);
      const originalCard = await revealCard(page, originalCardName);
      const beforeSelectionStates = await ariaStates(originalCard);
      await originalCard.click();
      if (scenario === 'justice') await page.getByRole('button', {name: 'Close public cards', exact: true}).waitFor({state: 'hidden'});
      await pause(350); await decisions(page);
      await selectIntent(page, selectedAction);
      if (scenario === 'purchase') {
        const quantity = page.getByRole('textbox', {name: 'Quantity', exact: true});
        await focusByKeyboard(page, quantity); await frame(page);
        await page.keyboard.press('ControlOrMeta+A'); await frame(page); await page.keyboard.type('3');
        assert.equal(await quantity.inputValue(), '3', 'quantity draft entered before interruption');
      }
      const selected = page.getByRole('button', {name: `Confirm ${selectedAction}`, exact: true});
      await selected.waitFor();
      const selectedStates = await ariaStates(await revealCard(page, originalCardName));
      // Live browser inspection of this Flutter build records selected cards as
      // aria-current, not aria-selected or aria-pressed on the button role.
      const selectionAttributes = ['aria-current'];
      assert.equal(beforeSelectionStates['aria-current'], 'false', 'original physical card starts unselected');
      assert.equal(selectedStates['aria-current'], 'true', 'original physical card has observed selected semantics');
      await decisions(page);
      await selected.waitFor();
      const submit = page.getByRole('button', {name: `Confirm ${selectedAction}`, exact: true});
      assert.equal(await submit.isEnabled(), scenario !== 'purchase', 'only a fully paid selected action is initially enabled');
      const before = await snapshot();
      const assignedStyles = before.projection.board.players.map(player => player.card_style);
      assert.equal(new Set(assignedStyles).size, 4, 'four distinct public styles before interruption');
      assert.ok(assignedStyles.every(style => ['first-light', 'rain-glaze', 'ember-glaze', 'drypoint'].includes(style)));
      if (scenario === 'justice') assert.ok(before.projection.board.decision_id, 'Justice has a pending decision');
      assert.ok(streams.length > 0, 'real backend stream opened');
      const connectedCardGeometry=scenario==='purchase'?await relativeHandCardGeometry(originalCard):null;
      await capture('connected-draft');
      await traceGeometry(page,originalCard,'connected-popup');
      if (process.env.CGMS_RECONNECT_FOCUS_ONLY) {
        await hand(page);
        await capture('focused-card-after-context');
        report.cases.push({case:'focused-card-full-bounds-after-context',pass:true});
        await context.close(); continue;
      }

      offline = true;
      await context.setOffline(true);
      for (const stream of streams) {
        await stream.page.close({code: 1012, reason: 'offline acceptance phase'});
        await stream.server.close({code: 1012, reason: 'offline acceptance phase'});
      }
      // CgmsNotice intentionally merges title and guidance into one group.
      const semantics = page.locator('flt-semantics-host');
      await semantics.getByText(disconnectedTitle).waitFor();
      await semantics.getByText('Your last confirmed table is still visible. Reconnecting automatically; no action is passed or resent.').waitFor();
      // Prove the existing eight-read budget really stops before an explicit
      // user retry. No sleeps or unbounded retries substitute for readiness.
      await semantics.getByText('Connection is still unavailable after bounded recovery. Use Refresh connection to try again.').waitFor();
      await semantics.getByText('Your last confirmed table is still visible. Use Refresh connection to retry; no action is passed or resent.').waitFor();
      assert.equal(offlineFailures.length, 8, 'automatic reconnect stops after exactly eight failed GETs');
      assert.equal(await submit.isEnabled(), false, 'mutation disabled during stream interruption');
      await selected.waitFor();
      // Flutter may include its visible tooltip a second time in the accessible
      // name of the same focused toolbar button.
      const refresh = page.getByRole('button', {name: /^Refresh connection(?: Refresh connection)?$/});
      await capture('offline');
      await traceGeometry(page,originalCard,'offline-popup');
      await focusByKeyboard(page, refresh); await frame(page);
      await Promise.all([
        page.waitForEvent('requestfailed', {predicate: request => request.method() === 'GET' && request.url() === matchUrl}),
        page.keyboard.press('Enter'),
      ]);
      await frame(page);
      assert.equal(offlineFailures.length, 9, 'explicit user retry makes one immediate failed GET after the exhausted budget');
      assert.equal(await submit.isEnabled(), false, 'mutation stays disabled after offline refresh');
      assert.equal(mutations.length, 0, 'offline navigation or refresh never submits or passes');

      const streamCount = streams.length;
      const restoredReadPromise = page.waitForResponse(response => response.url() === matchUrl && response.request().method() === 'GET');
      await context.setOffline(false);
      offline = false;
      // The restarted bounded recovery owns the next read and stream; restoring
      // transport requires no second Refresh, action, or page reload.
      const restoredRead = await restoredReadPromise;
      assert.equal(restoredRead.status(), 200, 'automatic restored-transport retry succeeds');
      const classified = classifyOfflineConsole(offlineConsole, offlineFailures, matchUrl);
      report.expectedFaults.push(...classified.expected);
      report.errors.push(...classified.unexpected);
      await semantics.getByText(disconnectedTitle).waitFor({state: 'hidden'});
      for (let attempt = 0; attempt < 50 && streams.length === streamCount; attempt++) await pause(50);
      assert.ok(streams.length > streamCount, 'new stream reconnects to real backend');
      await page.getByRole('heading', {name: selectedAction, exact: true}).waitFor();
      await selected.waitFor();
      assert.equal(await submit.isEnabled(), scenario !== 'purchase', 'restored snapshot retains known underpayment and enables valid decisions');
      if (scenario === 'purchase') {
        const quantity = page.getByRole('textbox', {name: 'Quantity', exact: true});
        await focusByKeyboard(page, quantity); await frame(page);
        assert.equal(await quantity.inputValue(), '3', 'raw quantity draft survives transport interruption');
        assert.equal(await submit.isEnabled(), false, 'restored quantity three remains unaffordable with exact ten-point payment');
        await page.keyboard.press('ControlOrMeta+A'); await page.keyboard.type('2'); await frame(page);
        assert.equal(await quantity.inputValue(), '2', 'player can correct retained quantity without changing physical payment');
        assert.equal(await submit.isEnabled(), true, 'corrected affordable purchase is enabled after reconnect');
        assert.equal(mutations.length, 0, 'correcting a draft never submits the purchase');
      }
      await traceGeometry(page,originalCard,'restored-popup');
      if (scenario === 'purchase') await hand(page);
      if(process.env.CGMS_RECONNECT_DISMISS) {
        assert.equal(mutations.length,0,'diagnostic dismissal never submits');
        report.cases.push({case:'diagnostic-keyboard-dismissal-full-card-bounds',pass:true});
        await context.close();continue;
      }
      const restoredCardGeometry=scenario==='purchase'?await relativeHandCardGeometry(originalCard):null;
      if(connectedCardGeometry) for(const [key,value] of Object.entries(connectedCardGeometry)) assert.ok(Math.abs(restoredCardGeometry[key]-value)<=1/64,`reconnect preserves single-card hand geometry ${key}: ${value} versus ${restoredCardGeometry[key]}`);
      const restoredStates = await ariaStates(await revealCard(page, originalCardName));
      for (const name of selectionAttributes) assert.equal(restoredStates[name], selectedStates[name], `original physical card retains ${name} after reconnect`);
      await decisions(page); await selected.waitFor();
      const after = await snapshot();
      assert.equal(after.version, before.version, 'no automatic authoritative command while disconnected');
      assert.deepEqual(scope(after), scope(before), 'game, turn and decision scope preserved');
      assert.deepEqual(after.projection.board.players.map(player => player.card_style), assignedStyles, 'shared match styles survive reconnect');
      assert.equal(mutations.length, 0, 'all page POSTs remain absent including automatic pass');
      assert.equal(offlineFailures.length, 9, 'eight bounded automatic GET failures plus one explicit retry, with no extra offline requests');
      await capture('restored-draft');
      report.cases.push({case: `${scenario}-bounded-offline-retry-restores-draft-selection-and-authority-scope`, automaticReadLimit: 8, explicitOfflineRefreshes: 1, restoredTransportRefreshes: 0, pass: true, beforeVersion: before.version, afterVersion: after.version, scope: scope(after), assignedStyles, physicalCardSelection: {card: originalCardName.source, connectedCardGeometry, restoredCardGeometry, selectionAttributes, beforeSelectionStates, selectedStates, restoredStates}, snapshotResponses, realStreamConnections: streams.length, offlineFailures, pagePosts: mutations.length});
      await context.close();
    }
    assert.deepEqual(report.errors, [], 'unexpected browser diagnostics');
    report.status = 'passed';
  } catch (error) {
    report.status = 'failed'; report.failure = error.stack || error.message;
    for (const [i, context] of browser.contexts().entries()) {
      for (const page of context.pages()) {
        try {await page.screenshot({path: path.join(evidence, `${engine}-reconnect-failure-${i}.png`)});}
        catch (captureError) {report.captureErrors.push({context: i, stage: 'screenshot', message: captureError.message});}
      }
    }
    throw error;
  } finally {
    report.finishedAt = new Date().toISOString();
    fs.writeFileSync(path.join(evidence, `${engine}-reconnect-report.json`), JSON.stringify(report, null, 2));
    await browser.close();
  }
  console.log(JSON.stringify(report));
}

module.exports = {expectedOfflineConsole, classifyOfflineConsole, decisions, focusByKeyboard, hand, section, revealCard, closePublicCards, closeActions, activate, selectIntent, gameOption, cancelDraft, waitForCommandPresentation};
if (require.main === module) run().catch(error => {console.error(error); process.exitCode = 1;});
