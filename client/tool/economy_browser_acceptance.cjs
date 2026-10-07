// Real local authority and PostgreSQL, with explicitly synthetic QA funding.
// Launch xops/postgres_integration.py --webfixture-economy before this harness.
const path = require('node:path');
const fs = require('node:fs');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const {ExpectedAbort, isBrowserError} = require('./browser_faults.cjs');
const {gameOption} = require('./reconnect_browser_acceptance.cjs');
if (!process.env.CGMS_CONTAINER) process.env.PLAYWRIGHT_BROWSERS_PATH = path.resolve(__dirname, '../.browser-tools/browsers');
const engines = require('../.browser-tools/node_modules/playwright');
const origin = process.env.CGMS_BROWSER_ORIGIN || 'http://127.0.0.1:8081';
const engine = process.env.CGMS_BROWSER || 'chromium';
const run = `economy-${engine}-${Date.now()}`;
const evidence = process.env.CGMS_EVIDENCE_DIR ? path.resolve(process.env.CGMS_EVIDENCE_DIR) : path.resolve(__dirname, '../.browser-tools/economy-evidence', run);
fs.mkdirSync(evidence, {recursive: true});
const sha256 = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const assetPath = path.resolve(__dirname, '../build/web/main.dart.js');
const entrypointPath = path.resolve(__dirname, '../build/web/index.html');
const bootstrapPath = path.resolve(__dirname, '../build/web/flutter_bootstrap.js');
const report = {
  run, engine, status: 'running', startedAt: new Date().toISOString(),
  diagnostics: 'pageerror-and-console-error-all-pages-v1',
  fixture: 'QA-only prior-day finalized ledgers fund 30 dirt through approved RewardFinalizedTx; seat 1 provider fixtures use approved confirmed 30-day Ad-free/24-hour Daily terms or a synthetic remaining expiry for Weekly/Monthly/Yearly; account reads, room starts and earned redemption use real authenticated HTTP and PostgreSQL; no live store verification',
  policy: {reward: 10, passDayPrice: 30}, nativeAcceptance: 'excluded',
  assetSha256: sha256(fs.readFileSync(assetPath)),
  entrypointSha256: sha256(fs.readFileSync(entrypointPath)),
  bootstrapSha256: sha256(fs.readFileSync(bootstrapPath)),
  harnessSha256: sha256(fs.readFileSync(__filename)),
  cases: [], screenshots: [], errors: [], expectedFaults: [],
};
const faults = new WeakMap();
const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
const endpoint = pathname => url => url.pathname === pathname;
// Flutter may insert a live-region mirror inside or beside its semantics host.
// Assert the visible control/text separately from its screen-reader announcement.
const content = (page, value, options) => page.locator('flt-semantics-host').getByText(value, options)
  .and(page.locator(':not(flt-announcement-polite):not(flt-announcement-assertive)'));

function diagnostics(page) {
  page.on('pageerror', error => {
    const fault = faults.get(page);
    if (fault?.registry.consume(error.message)) report.expectedFaults.push({phase: fault.phase, message: error.message});
    else report.errors.push({type: 'pageerror', message: error.message});
  });
  page.on('console', message => {
    if (!isBrowserError(message.type(), message.text())) return;
    const fault = faults.get(page);
    const event = {message: message.text(), location: message.location()};
    if (fault?.registry.consumeConsole(message.text(), message.location().url)) report.expectedFaults.push({phase: fault.phase, ...event});
    else report.errors.push({type: 'console', ...event});
  });
}
function beginFault(page, phase) {
  const registry = new ExpectedAbort(); registry.begin();
  faults.set(page, {phase, registry}); return registry;
}
async function stableAsset(context) {
  assert.equal(sha256(fs.readFileSync(__filename)), report.harnessSha256, 'acceptance harness changed during execution');
  for (const [name, file, expected] of [['main.dart.js', assetPath, report.assetSha256], ['index.html', entrypointPath, report.entrypointSha256], ['flutter_bootstrap.js', bootstrapPath, report.bootstrapSha256]]) {
    const response = await context.request.get(`${origin}/${name}`);
    assert.equal(response.status(), 200, `built ${name} must be served`);
    assert.equal(sha256(await response.body()), expected, `served ${name} changed during acceptance`);
    assert.equal(sha256(fs.readFileSync(file)), expected, `local ${name} changed during acceptance`);
  }
}
async function account(context) {
  const response = await context.request.get(`${origin}/v1/economy`);
  assert.equal(response.status(), 200, 'own account snapshot');
  return response.json();
}
async function fixture(browser, suffix, seat = 1, viewport = {width: 390, height: 844}, scenario = 'opening', players = 4) {
  const context = await browser.newContext({viewport, hasTouch: true});
  // The fixture run grammar is bounded to 32 characters. No credentials are logged.
  const fixtureRun = `${engine.slice(0, 7)}-${run.slice(-13)}-${suffix}`;
  const response = await context.request.post(`${origin}/__fixture/session`, {
    headers: {Origin: origin}, data: {scenario, players, seat, run: fixtureRun},
  });
  assert.equal(response.status(), 200, 'economy fixture session');
  const {match_id} = await response.json();
  await stableAsset(context);
  const page = await context.newPage(); diagnostics(page);
  await page.goto(`${origin}/#/online?match=${encodeURIComponent(match_id)}`);
  await page.getByRole('heading', {name: 'CGMS · Online', exact: true}).waitFor({timeout: 30000});
  await content(page, /Seat 1’s turn|Waiting for seat|Board closed|settlement/).first().waitFor({timeout: 30000});
  return {context, page, match_id};
}
async function pointerTarget(page, name) {
  const button = page.getByRole('button', {name, exact: true}).first();
  // Resize completion precedes Flutter's frame/semantics update in Firefox.
  // Wait for the toolbar to enter the new viewport; do not retry wrong hits.
  for (let frame = 0; frame < 60; frame++) {
    const box = await button.boundingBox(), viewport = page.viewportSize();
    if (box && box.x >= 0 && box.x + box.width <= viewport.width && box.y >= 0 && box.y + box.height <= viewport.height) break;
    await pause(50);
  }
  const hit = await button.evaluate(element => {
    const box = element.getBoundingClientRect();
    const target = document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2);
    return {correct: element === target || element.contains(target), target: target?.textContent};
  });
  assert.equal(hit.correct, true, `${name} pointer target intercepted by ${hit.target}`);
}
async function openPanel(page) {
  const wallet = page.getByRole('button', {name: 'Allowance and dirt', exact: true}).first();
  await pointerTarget(page, 'Allowance and dirt');
  await wallet.tap();
  await content(page, 'Earn 10 dirt once per finalized online game. Offline play earns no dirt; unfinished games do not refund starts.', {exact: true}).waitFor();
}
async function closePanel(page) {
  const close = page.getByRole('button', {name: 'Close', exact: true});
  await close.click(); await close.waitFor({state: 'hidden'}); await pause(350);
}
async function setting(page, name) {
  await page.getByRole('button', {name: 'Display settings', exact: true}).click();
  await page.getByRole('menuitemcheckbox', {name, exact: true}).click(); await pause(350);
}
async function days(page, count) {
  const dropdown = page.getByRole('button', {name: /^Earned access /});
  await dropdown.focus(); await page.keyboard.press('Space');
  const option = page.getByRole('menuitem', {name: count === 1 ? 'Daily · 1 day' : 'Weekly · 7 days', exact: true});
  for (let i = 0; i < 8 && !(await option.count()); i++) {await page.keyboard.press('ArrowDown'); await pause(80);}
  await option.waitFor(); await pause(350); await option.click(); await pause(350);
}
async function screenshot(page, name) {
  await pause(180);
  const file = `${name}.png`;
  await page.screenshot({path: path.join(evidence, file)}); report.screenshots.push(file);
}
async function target(page, name) {
  const button = page.getByRole('button', {name, exact: true});
  const box = await button.boundingBox();
  assert.ok(box && box.width >= 48 && box.height >= 48, `${name} must have a 48px touch target: ${JSON.stringify(box)}`);
  return {name, width: box.width, height: box.height};
}
async function keyboardFocus(page, name) {
  const button = page.getByRole('button', {name, exact: true});
  // DOM focus alone can bypass Flutter's Focus/Scrollable handling. Traverse
  // with real keyboard events, including after a disabled control is rebuilt.
  for (let step = 0; step < 12; step++) {
    await page.keyboard.press('Tab'); await pause(250);
    if (await button.evaluate(element => element === document.activeElement)) break;
  }
  assert.equal(await button.evaluate(element => element === document.activeElement), true, `${name} must be keyboard reachable`);
  const box = await button.boundingBox(), viewport = page.viewportSize();
  assert.ok(box && box.x >= 0 && box.y >= 0 && box.x + box.width <= viewport.width && box.y + box.height <= viewport.height, `${name} must scroll into the viewport`);
}
async function keyboardTargets(page) {
  const targets = [];
  for (const name of ['Spend 30 dirt', 'Refresh allowance', 'Close']) {
    await keyboardFocus(page, name);
    targets.push(await target(page, name));
  }
  return targets;
}
async function freshAccount(f) {
  const state = await account(f.context);
  assert.equal(state.enabled, true, 'run --webfixture-economy');
  assert.equal(state.reward, 10); assert.equal(state.pass_day_price, 30);
  assert.equal(state.benefits.dirt, 30); assert.equal(state.free_starts_remaining, 2);
  assert.equal(state.benefits.unlimited, false); assert.equal(state.benefits.no_ads, false);
  const reset = new Date(state.resets_at);
  assert.ok(reset.getTime() > Date.now());
  assert.equal(reset.getUTCHours(), 0); assert.equal(reset.getUTCMinutes(), 0);
  await openPanel(f.page);
  await content(f.page, '30 dirt', {exact: true}).waitFor();
  await content(f.page, '2 free online starts remaining today', {exact: true}).waitFor();
  await content(f.page, /^Free starts reset at .* UTC\.$/).waitFor();
  assert.equal(await content(f.page, /1970|Pass expiry:|Premium expiry:/).count(), 0, 'absent entitlements must not render epoch expiry');
  return state;
}
async function post(f, pathname, data) {
  // All browser seats share loopback's production IP bucket (20 burst, 5/s).
  // Pace synthetic setup instead of weakening admission or retrying 429s.
  await pause(350);
  if (!f.csrf) {
    const session = await f.context.request.get(`${origin}/v1/session`);
    assert.equal(session.status(), 200); f.csrf = (await session.json()).csrf_token;
    await pause(250);
  }
  const response = await f.context.request.post(`${origin}${pathname}`, {
    headers: {Origin: origin, 'X-CSRF-Token': f.csrf}, data,
  });
  assert.equal(response.status(), 200, `authenticated setup request ${pathname.split('/').slice(0, 3).join('/')}`);
  return response.json();
}
async function snapshot(f) {
  const response = await f.context.request.get(`${origin}/v1/matches/${f.match_id}`);
  assert.equal(response.status(), 200); return response.json();
}
async function decisions(page) {
  await require('./reconnect_browser_acceptance.cjs').decisions(page);
}
async function adFreeJourneys(browser) {
  for (const [layout, width, height] of [['phone', 390, 844], ['tablet', 768, 1024], ['desktop', 1440, 900]]) {
    const suffix = `ad-${layout}`;
    const owner = await fixture(browser, suffix, 1, {width, height}, 'ad-free');
    let purchasePosts = 0;
    owner.page.on('request', request => {
      if (request.method() === 'POST' && new URL(request.url()).pathname === '/v1/economy/passes') purchasePosts++;
    });
    const before = await account(owner.context);
    assert.equal(before.benefits.no_ads, true);
    assert.equal(before.benefits.unlimited, false, 'standalone Ad-free never grants unlimited starts');
    assert.equal(before.free_starts_remaining, 2);
    const adFreeRemaining = new Date(before.benefits.ad_free_until).getTime() - Date.now();
    assert.ok(adFreeRemaining > 29 * 86400000 && adFreeRemaining <= 30 * 86400000, 'standalone Ad-free uses the approved 30-day term');
    await openPanel(owner.page);
    await content(owner.page, 'Ad-free access active', {exact: true}).waitFor();
    await content(owner.page, /^Ad-free expiry: .* UTC$/).waitFor();
    assert.equal(await content(owner.page, 'Unlimited online starts', {exact: true}).count(), 0);
    await screenshot(owner.page, `${layout}-standalone-ad-free`);

    await closePanel(owner.page);
    await setting(owner.page, 'Right to left'); await setting(owner.page, '200% text');
    await openPanel(owner.page);
    await content(owner.page, /^Ad-free expiry: .* UTC$/).waitFor();
    await screenshot(owner.page, `${layout}-ad-free-rtl-200-top`);
    const targets = await keyboardTargets(owner.page);
    await screenshot(owner.page, `${layout}-ad-free-rtl-200-controls`);
    assert.equal(purchasePosts, 0, 'display navigation never redeems a product');
    await closePanel(owner.page);
    await setting(owner.page, 'Right to left'); await setting(owner.page, '200% text');
    await openPanel(owner.page);
    await owner.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).click();
    await content(owner.page, /^Pass confirmed · 30 dirt · expires .* UTC$/).waitFor();
    await content(owner.page, 'Unlimited online starts', {exact: true}).waitFor();
    await content(owner.page, 'Ad-free access active', {exact: true}).waitFor();
    const combined = await account(owner.context);
    assert.equal(combined.benefits.unlimited, true); assert.equal(combined.benefits.no_ads, true);
    assert.equal(combined.benefits.dirt, 0);
    assert.equal(combined.benefits.ad_free_until, before.benefits.ad_free_until, 'earned time never extends ad-free expiry');
    await owner.page.reload();
    await owner.page.getByRole('heading', {name: 'CGMS · Online', exact: true}).waitFor();
    await openPanel(owner.page);
    await content(owner.page, 'Unlimited online starts', {exact: true}).waitFor();
    await content(owner.page, /^Ad-free expiry: .* UTC$/).waitFor();
    await screenshot(owner.page, `${layout}-combined-benefits-reloaded`);
    assert.equal(purchasePosts, 1, 'one explicit redemption, no automatic purchase on reload');

    const outsider = await fixture(browser, suffix, 2, {width, height}, 'ad-free');
    const other = await account(outsider.context);
    assert.equal(other.benefits.no_ads, false); assert.equal(other.benefits.unlimited, false);
    assert.equal(new Date(other.benefits.ad_free_until).getTime(), 0);
    assert.equal(other.benefits.dirt, 30);
    for (const f of [owner, outsider]) {
      const publicBoard = (await snapshot(f)).projection.board;
      assert.doesNotMatch(JSON.stringify(publicBoard), /"(?:ad_free_until|no_ads|premium_until|pass_until|dirt)":/, 'private benefits never enter public game context');
      await stableAsset(f.context);
    }
    report.cases.push({case: 'standalone-ad-free-remains-capped-earned-access-combines-without-extension-reload-private-to-owner', layout, width, height, pass: true, touchTargets: targets, purchaseRequests: purchasePosts});
    await owner.context.close(); await outsider.context.close();
  }
}
async function paidAccessJourneys(browser) {
  for (const [layout, width, height] of [['phone', 390, 844], ['tablet', 768, 1024], ['desktop', 1440, 900]]) {
    for (const kind of ['daily', 'weekly', 'monthly', 'yearly']) {
      const suffix = `p${kind[0]}-${layout[0]}`;
      const owner = await fixture(browser, suffix, 1, {width, height}, `paid-${kind}`);
      let mutations = 0;
      owner.page.on('request', request => {if (request.method() === 'POST') mutations++;});
      const before = await account(owner.context);
      assert.equal(before.benefits.unlimited, true, `${kind} grants unlimited starts`);
      assert.equal(before.benefits.no_ads, true, `${kind} removes ads`);
      assert.equal(before.free_starts_remaining, 3, `${kind} fixture start did not consume free allowance`);
      assert.equal(before.benefits.dirt, 30);
      assert.equal(before.benefits.ad_free_until, before.benefits.premium_until);
      assert.equal(new Date(before.benefits.pass_until).getTime(), 0, 'paid grant is not an earned pass');
      const remaining = new Date(before.benefits.premium_until).getTime() - Date.now();
      assert.ok(remaining > 0);
      if (kind === 'daily') assert.ok(remaining > 23 * 3600000 && remaining <= 24 * 3600000, 'paid Daily has the approved 24-hour term');
      await openPanel(owner.page);
      await content(owner.page, 'Unlimited online starts', {exact: true}).waitFor();
      await content(owner.page, 'Ad-free access active', {exact: true}).waitFor();
      await content(owner.page, '3 free online starts remaining today', {exact: true}).waitFor();
      await content(owner.page, /^Ad-free expiry: .* UTC$/).waitFor();
      await content(owner.page, /^Premium expiry: .* UTC$/).waitFor();
      assert.equal(await content(owner.page, 'Ads remain active', {exact: true}).count(), 0);
      await screenshot(owner.page, `${layout}-paid-${kind}-active`);
      await keyboardFocus(owner.page, 'Refresh allowance');
      for (const terms of [
        'Ad-free: 30 days without ads. The 3 free online starts per day still apply unless you also have unlimited access.',
        'Paid Daily: 24 hours of unlimited online starts and no ads from confirmed purchase. Does not renew automatically.',
        'Paid Weekly, Monthly and Yearly: unlimited online starts and no ads.',
      ]) await content(owner.page, terms, {exact: true}).waitFor();
      await screenshot(owner.page, `${layout}-paid-${kind}-terms`);
      await closePanel(owner.page);
      await owner.page.reload();
      await owner.page.getByRole('heading', {name: 'CGMS · Online', exact: true}).waitFor();
      await openPanel(owner.page);
      await content(owner.page, 'Unlimited online starts', {exact: true}).waitFor();
      await content(owner.page, 'Ad-free access active', {exact: true}).waitFor();
      const restored = await account(owner.context);
      assert.deepEqual(restored.benefits, before.benefits, 'reload does not renew or alter paid access');
      assert.equal(restored.free_starts_remaining, 3);
      assert.equal(mutations, 0, 'reading terms and reloading never purchases or submits a game action');
      const outsider = await fixture(browser, suffix, 2, {width, height}, `paid-${kind}`);
      const other = await account(outsider.context);
      assert.equal(other.benefits.unlimited, false); assert.equal(other.benefits.no_ads, false);
      assert.equal(other.free_starts_remaining, 2);
      assert.equal(other.benefits.dirt, 30);
      assert.equal(new Date(other.benefits.ad_free_until).getTime(), 0);
      for (const f of [owner, outsider]) {
        assert.doesNotMatch(JSON.stringify(await snapshot(f)), /"(?:ConfirmedAt|confirmed_at|ad_free_until|no_ads|premium_until|pass_until|dirt)":/, 'paid account/provider data never enters the match response');
        await stableAsset(f.context);
      }
      report.cases.push({case: 'paid-access-removes-ads-keeps-free-allowance-private-and-stable-on-reload', product: kind, layout, width, height, pass: true, mutationRequests: mutations});
      await owner.context.close(); await outsider.context.close();
    }
  }
}
async function malformedAccountJourney(browser) {
  const f = await fixture(browser, 'invalid');
  await freshAccount(f);
  let mutations = 0, injected = 0;
  f.page.on('request', request => {if (request.method() === 'POST') mutations++;});
  const routePattern = endpoint('/v1/economy');
  await f.page.route(routePattern, async route => {
    const response = await route.fetch(); assert.equal(response.status(), 200);
    const data = await response.json(); injected++;
    // A successful status is not enough: reject bad price/benefit types before
    // changing the last confirmed account or allowing another spend.
    await route.fulfill({response, json: {...data, pass_day_price: 0, benefits: {...data.benefits, no_ads: 'false'}}});
  });
  await keyboardFocus(f.page, 'Refresh allowance'); await f.page.keyboard.press('Enter');
  await content(f.page, 'Could not refresh your allowance. Try again.', {exact: true}).waitFor();
  assert.equal(injected, 1, 'one deliberately malformed successful HTTP response');
  await content(f.page, '30 dirt', {exact: true}).waitFor();
  assert.equal(await f.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).isEnabled(), false);
  assert.equal(await f.page.getByRole('button', {name: 'Spend 0 dirt', exact: true}).count(), 0);
  assert.equal(mutations, 0, 'invalid account refresh never submits an action');
  await screenshot(f.page, 'phone-invalid-account-spending-disabled');
  await f.page.unroute(routePattern);
  await keyboardFocus(f.page, 'Refresh allowance'); await f.page.keyboard.press('Enter');
  await content(f.page, 'Could not refresh your allowance. Try again.', {exact: true}).waitFor({state: 'hidden'});
  assert.equal(await f.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).isEnabled(), true);
  assert.equal((await account(f.context)).benefits.dirt, 30);
  assert.equal(mutations, 0, 'successful recovery requires no purchase');
  await screenshot(f.page, 'phone-invalid-account-refresh-recovered');
  report.cases.push({case: 'malformed-successful-account-response-preserves-cache-disables-spending-and-recovers', pass: true, injectedResponses: injected, mutationRequests: mutations});
  await stableAsset(f.context); await f.context.close();
}
async function streamedAllowanceJourney(browser) {
  const roster = [];
  for (let seat = 1; seat <= 3; seat++) roster.push(await fixture(browser, 'wait', seat, {width: 390, height: 844}, 'settlement', 3));
  const owner = roster[0];
  for (const f of roster) assert.equal((await account(f.context)).benefits.dirt, 30);
  await pause(2000);
  // Consume the two remaining starts with normal authorized room operations.
  // The only synthetic data is the documented fixture's initial state/funding.
  for (let index = 0; index < 2; index++) {
    const room = await post(owner, '/v1/rooms', {command_id: `economy-room-${index}`, capacity: 3, games_per_match: 3});
    const invitation = await post(owner, `/v1/rooms/${room.id}/invitations`, {});
    for (const member of roster.slice(1)) await post(member, `/v1/rooms/${room.id}/join`, invitation);
    await post(owner, `/v1/rooms/${room.id}/matches`, {command_id: `economy-start-${index}`});
  }
  for (const f of roster) assert.equal((await account(f.context)).free_starts_remaining, 0);
  const before = await snapshot(owner);
  await decisions(owner.page);
  const promisePaid = owner.page.waitForResponse(r => new URL(r.url()).pathname === `/v1/matches/${owner.match_id}/commands` && r.request().method() === 'POST');
  await owner.page.getByRole('button', {name: 'Pay promise', exact: true}).click();
  assert.equal((await promisePaid).status(), 200);
  for (const f of roster) {
    // Give each scripted browser a fresh view before its independent command.
    // WebKit can defer background-page stream frames. The authority correctly
    // rejects a stale version; do not issue/retry that command to make green.
    await f.page.bringToFront(); await decisions(f.page);
    const loaded = f.page.waitForResponse(r => new URL(r.url()).pathname === `/v1/matches/${f.match_id}` && r.request().method() === 'GET');
    await f.page.getByRole('button', {name: 'Refresh connection', exact: true}).tap();
    const response = await loaded; assert.equal(response.status(), 200); await response.finished();
    await gameOption(f.page, 'Finish Settlement');
    const finish = f.page.getByRole('button', {name: 'Confirm Finish Settlement', exact: true});
    for (let frame = 0; frame < 60 && !(await finish.isEnabled()); frame++) await pause(50);
    assert.equal(await finish.isEnabled(), true); await pause(100);
    const confirmed = f.page.waitForResponse(r => new URL(r.url()).pathname === `/v1/matches/${f.match_id}/commands` && r.request().method() === 'POST');
    await finish.click();
    assert.equal((await confirmed).status(), 200); await pause(200);
  }
  const waiting = 'The next game is waiting for every player’s online access. Check your allowance; a player may need a pass, the UTC reset, or account access.';
  // Do not reload or press reconnect: the streamed finalization must trigger
  // the client's one-shot read of current admission state by itself.
  await content(owner.page, waiting, {exact: true}).waitFor({timeout: 20000});
  const stopped = await snapshot(owner);
  assert.equal(stopped.admission_waiting, true);
  assert.equal(stopped.projection.board.game_id, before.projection.board.game_id);
  assert.equal(stopped.projection.online.financial_phase, 'finalized');
  assert.equal(stopped.projection.online.completed_games, 1);
  assert.equal(await content(owner.page, /Seat \d+’s turn/).count(), 0, 'a finalized game has no current-turn owner');
  assert.equal(await content(owner.page, /free online starts remaining today|Pass expiry:/).count(), 0, 'generic waiting does not publish economic details');
  await screenshot(owner.page, 'phone-streamed-next-game-allowance-wait');
  for (const f of roster) {
    const earned = await account(f.context);
    assert.equal(earned.benefits.dirt, 40, 'actual finalized fixture game awards 10 exactly once');
    await openPanel(f.page);
    await content(f.page, '40 dirt', {exact: true}).waitFor();
    await f.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).tap();
    await content(f.page, /^Pass confirmed · 30 dirt · expires .* UTC$/).waitFor();
    await closePanel(f.page);
  }
  await content(owner.page, waiting, {exact: true}).waitFor({state: 'hidden', timeout: 20000});
  await content(owner.page, /Seat .*turn|Waiting for seat/).first().waitFor({timeout: 20000});
  const resumed = await snapshot(owner);
  assert.notEqual(resumed.projection.board.game_id, before.projection.board.game_id);
  assert.notEqual(resumed.admission_waiting, true);
  assert.equal(resumed.projection.online.completed_games, 1);
  for (const f of roster) {
    const state = await account(f.context);
    assert.equal(state.benefits.dirt, 10); assert.equal(state.benefits.unlimited, true);
    assert.equal(state.free_starts_remaining, 0, 'pass starts do not consume more free allowance');
  }
  await screenshot(owner.page, 'phone-streamed-next-game-after-roster-passes');
  report.cases.push({case: 'streamed-finalization-allowance-wait-and-roster-pass-unlock-without-reload', pass: true, players: 3, actualFinalizedReward: 10, remainingDirt: 10});
  for (const f of roster) {await stableAsset(f.context); await f.context.close();}
}

(async () => {
  assert.ok(engines[engine], 'supported browser engine');
  const browser = await engines[engine].launch({headless: true}); report.version = browser.version();
  try {
    const owner = await fixture(browser, 'main'); await freshAccount(owner);
    await days(owner.page, 7);
    assert.equal(await owner.page.getByRole('button', {name: 'Spend 210 dirt', exact: true}).isEnabled(), false);
    await days(owner.page, 1);
    const touchTargets = await keyboardTargets(owner.page);
    report.cases.push({case: 'approved-prices-utc-allowance-offline-explanation-and-insufficient-weekly', pass: true, touchTargets});

    for (const [width, height, layout] of [[390, 844, 'phone'], [768, 1024, 'tablet'], [1440, 900, 'desktop']]) {
      await closePanel(owner.page); await owner.page.setViewportSize({width, height});
      await openPanel(owner.page); await screenshot(owner.page, `${layout}-allowance`);
      await closePanel(owner.page);
      await setting(owner.page, 'Right to left'); await setting(owner.page, '200% text');
      for (const name of ['Display settings', 'Refresh connection', 'Allowance and dirt']) await pointerTarget(owner.page, name);
      await openPanel(owner.page);
      await screenshot(owner.page, `${layout}-rtl-200-allowance-top`);
      const targets = await keyboardTargets(owner.page);
      await screenshot(owner.page, `${layout}-rtl-200-allowance-controls`);
      await closePanel(owner.page); await setting(owner.page, 'Right to left'); await setting(owner.page, '200% text');
      await openPanel(owner.page);
      report.cases.push({case: 'adaptive-account-dialog-rtl-200-percent', layout, width, height, pass: true, touchTargets: targets});
    }

    await closePanel(owner.page); await owner.page.setViewportSize({width: 390, height: 844}); await openPanel(owner.page);
    const fault = beginFault(owner.page, 'lost-pass-response');
    const passPath = '/v1/economy/passes'; const passRoute = endpoint(passPath);
    let posted = null, receipt = null, postCount = 0;
    owner.page.on('request', request => {if (new URL(request.url()).pathname === passPath && request.method() === 'POST') postCount++;});
    await owner.page.route(passRoute, async route => {
      assert.equal(route.request().method(), 'POST'); posted = route.request().postDataJSON();
      const response = await route.fetch(); assert.equal(response.status(), 200); receipt = await response.json();
      fault.record(route.request().url()); await route.abort('failed');
    });
    await owner.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).tap();
    await owner.page.getByRole('button', {name: 'Recover pass purchase', exact: true}).waitFor({state: 'attached'});
    assert.equal(await owner.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).isEnabled(), false);
    assert.equal(posted.days, 1); assert.equal(receipt.cost, 30); assert.equal(receipt.operation_id, posted.operation_id);
    assert.equal((await account(owner.context)).benefits.dirt, 0, 'real transaction committed before dropped reply');
    await keyboardFocus(owner.page, 'Recover pass purchase');
    await screenshot(owner.page, 'phone-lost-pass-response');
    await owner.page.unroute(passRoute); await pause(200); fault.end();
    await owner.page.reload();
    await owner.page.getByRole('heading', {name: 'CGMS · Online', exact: true}).waitFor(); await openPanel(owner.page);
    await owner.page.getByRole('button', {name: 'Recover pass purchase', exact: true}).waitFor({state: 'attached'});
    await keyboardFocus(owner.page, 'Recover pass purchase');
    await screenshot(owner.page, 'phone-reload-pending-pass');
    const recovered = owner.page.waitForResponse(response => new URL(response.url()).pathname === `${passPath}/${posted.operation_id}` && response.request().method() === 'GET');
    await keyboardFocus(owner.page, 'Recover pass purchase'); await owner.page.keyboard.press('Enter');
    assert.deepEqual(await (await recovered).json(), receipt, 'recovery must return original immutable receipt');
    await content(owner.page, /^Pass confirmed · 30 dirt · expires .* UTC$/).waitFor();
    await content(owner.page, 'Unlimited online starts', {exact: true}).waitFor();
    await content(owner.page, 'Ads remain active', {exact: true}).waitFor();
    await content(owner.page, /^Pass expiry: .* UTC$/).waitFor();
    assert.equal(await owner.page.getByRole('button', {name: 'Recover pass purchase', exact: true}).count(), 0);
    const after = await account(owner.context);
    assert.equal(after.benefits.dirt, 0); assert.equal(after.benefits.unlimited, true); assert.equal(after.benefits.no_ads, false);
    assert.equal(new Date(after.benefits.pass_until).getTime(), new Date(receipt.expires_at).getTime());
    assert.equal(postCount, 1, 'recovery of a committed receipt must not submit another purchase');
    await screenshot(owner.page, 'phone-recovered-pass');
    report.cases.push({case: 'lost-pass-response-reload-keyboard-recovery-original-receipt-no-double-charge', pass: true, purchaseRequests: postCount, remainingDirt: after.benefits.dirt});

    const outsider = await fixture(browser, 'main', 2); await freshAccount(outsider);
    const privateReceipt = await outsider.context.request.get(`${origin}${passPath}/${posted.operation_id}`);
    assert.equal(privateReceipt.status(), 404); assert.equal((await privateReceipt.json()).code, 'OPERATION_NOT_FOUND');
    assert.equal(await content(outsider.page, 'Unlimited online starts', {exact: true}).count(), 0);
    await screenshot(outsider.page, 'phone-other-account-private-allowance');
    report.cases.push({case: 'other-seat-retains-private-balance-and-cannot-read-owner-receipt', pass: true, otherSeatDirt: 30});
    await stableAsset(outsider.context); await outsider.context.close();

    const refresh = await fixture(browser, 'refresh'); await freshAccount(refresh);
    const refreshFault = beginFault(refresh.page, 'confirmed-pass-followup-snapshot-failure');
    const matchRoute = endpoint(`/v1/matches/${refresh.match_id}`);
    let armed = false, failures = 0;
    await refresh.page.route(passRoute, async route => {
      const response = await route.fetch(); assert.equal(response.status(), 200); armed = true;
      await route.fulfill({response});
    });
    await refresh.page.route(matchRoute, async route => {
      if (armed && failures === 0 && route.request().method() === 'GET') {
        failures++; refreshFault.record(route.request().url()); await route.abort('failed');
      } else await route.continue();
    });
    await refresh.page.getByRole('button', {name: 'Spend 30 dirt', exact: true}).tap();
    await content(refresh.page, /^Pass confirmed · 30 dirt · expires .* UTC$/).waitFor();
    for (let i = 0; i < 50 && failures === 0; i++) await pause(100);
    assert.equal(failures, 1, 'fault must interrupt the post-purchase match refresh'); await pause(200);
    assert.equal(await refresh.page.getByRole('button', {name: 'Recover pass purchase', exact: true}).count(), 0);
    assert.equal(await content(refresh.page, 'Your pass purchase has an unknown result. Recover it before buying another pass.', {exact: true}).count(), 0);
    assert.equal((await account(refresh.context)).benefits.dirt, 0);
    await screenshot(refresh.page, 'phone-confirmed-pass-refresh-failure');
    await refresh.page.unroute(matchRoute); await refresh.page.unroute(passRoute); refreshFault.end();
    await closePanel(refresh.page);
    await refresh.page.getByRole('button', {name: 'Refresh connection', exact: true}).click();
    await openPanel(refresh.page); await content(refresh.page, 'Unlimited online starts', {exact: true}).waitFor();
    await content(refresh.page, 'Ads remain active', {exact: true}).waitFor();
    assert.equal((await account(refresh.context)).benefits.no_ads, false);
    report.cases.push({case: 'confirmed-spend-remains-known-after-followup-match-refresh-fails', pass: true, deliberatelyFailedSnapshots: failures});
    await stableAsset(refresh.context); await refresh.context.close();
    await stableAsset(owner.context); await owner.context.close();
    await streamedAllowanceJourney(browser);
    await adFreeJourneys(browser);
    await paidAccessJourneys(browser);
    await malformedAccountJourney(browser);
    assert.deepEqual(report.errors, [], 'unexpected browser diagnostics'); report.status = 'passed';
  } catch (error) {
    report.status = 'failed'; report.failure = error.message;
    for (const [i, context] of browser.contexts().entries()) for (const [j, page] of context.pages().entries()) {
      await screenshot(page, `failure-${i}-${j}`);
      fs.writeFileSync(path.join(evidence, `failure-${i}-${j}.aria.txt`), await page.locator('flt-semantics-host').ariaSnapshot());
    }
    throw error;
  } finally {
    report.finishedAt = new Date().toISOString();
    fs.writeFileSync(path.join(evidence, 'report.json'), JSON.stringify(report, null, 2));
    await browser.close(); console.log(JSON.stringify({status: report.status, evidence, assetSha256: report.assetSha256, cases: report.cases.length, errors: report.errors.length}));
  }
})().catch(error => {console.error(error); process.exitCode = 1;});
