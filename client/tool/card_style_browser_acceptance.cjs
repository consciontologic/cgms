// Real Flutter/HTTP/PostgreSQL capture flow from a synthetic legal position.
// Start xops/postgres_integration.py --webfixture-economy first.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const assert = require('node:assert/strict');
const {isBrowserError} = require('./browser_faults.cjs');
const {decisions, revealCard, closePublicCards, closeActions, focusByKeyboard, activate, selectIntent} = require('./reconnect_browser_acceptance.cjs');
if (!process.env.CGMS_CONTAINER) process.env.PLAYWRIGHT_BROWSERS_PATH = path.resolve(__dirname, '../.browser-tools/browsers');
const engines = require('../.browser-tools/node_modules/playwright');
const engine = process.env.CGMS_BROWSER || 'chromium';
const origin = process.env.CGMS_BROWSER_ORIGIN || 'http://127.0.0.1:8081';
const output = path.resolve(process.env.CGMS_EVIDENCE_DIR || 'client/.browser-tools/card-styles');
fs.mkdirSync(output, {recursive: true});
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const report = {engine, status: 'running', fixture: 'legal diamond capture with synthetic starting position; real commands and authority', cases: [], screenshots: [], errors: [], assets: {}};
for (const file of ['main.dart.js', 'index.html', 'flutter_bootstrap.js']) report.assets[file] = hash(fs.readFileSync(path.resolve(__dirname, '../build/web', file)));
report.harness = hash(fs.readFileSync(__filename));
const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
const styles = board => board.players.map(p => p.card_style);
// These three public welcome illustrations can load before the session snapshot.
const welcomeArt = ['clubs-8', 'hearts-Q', 'diamonds-4'].map(card => `/assets/packages/cgms_ui/assets/art/${card}.png`);
const cardPath = (style, suit, rank) => `/assets/packages/cgms_ui/assets/art/${style === 'first-light' ? '' : style + '/'}${suit}-${rank}.png`;
async function table(page) {
  await closePublicCards(page);
  await closeActions(page);
  await page.locator('[flt-semantics-identifier="public-square"]').waitFor();
}
async function snapshot(f) {
  const response = await f.context.request.get(`${origin}/v1/matches/${encodeURIComponent(f.match_id)}`);
  assert.equal(response.status(), 200); return (await response.json()).projection.board;
}
async function screenshot(page, name) {
  await page.waitForLoadState('networkidle');
  await page.screenshot({path: path.join(output, name)}); report.screenshots.push(name);
}
async function fixture(browser, players, seat, viewport, run, scenario = 'card-style') {
  const context = await browser.newContext({viewport, hasTouch: true});
  const response = await context.request.post(`${origin}/__fixture/session`, {headers: {Origin: origin}, data: {scenario, players, seat, run}});
  assert.equal(response.status(), 200, 'card-style fixture');
  const {match_id} = await response.json();
  const page = await context.newPage(); const art = new Set();
  page.on('pageerror', e => report.errors.push(e.message));
  page.on('console', m => {if (isBrowserError(m.type(), m.text())) report.errors.push(m.text());});
  page.on('request', r => {const p = new URL(r.url()).pathname; if (p.includes('/assets/art/')) art.add(p);});
  page.on('response', r => {if (r.status() >= 400 && new URL(r.url()).pathname.startsWith('/assets/')) report.errors.push(`asset ${r.status()} ${new URL(r.url()).pathname}`);});
  await page.goto(`${origin}/#/online?match=${encodeURIComponent(match_id)}`);
  await page.getByRole('heading', {name: 'CGMS · Online', exact: true}).waitFor({timeout: 30000});
  await table(page); await page.waitForLoadState('networkidle');
  return {context, page, match_id, art};
}
async function chooseAttack(page) {
  for (const pattern of [/^8 of Clubs, copy 1/, /^8 of Diamonds, copy 1/]) {
    const card = await revealCard(page, pattern);
    await card.click();
    const close = page.getByRole('button', {name: 'Close public cards', exact: true});
    await close.waitFor({state: 'hidden'}); await pause(350);
  }
  await selectIntent(page, 'Attack');
}

async function refreshed(f) {
  await f.page.getByRole('button', {name: 'Refresh connection', exact: true}).click();
  await pause(350);
}
(async () => {
  const browser = await engines[engine].launch({headless: true}); report.version = browser.version();
  try {
    for (const [label, width, height, players] of [['phone',390,844,3], ['tablet',768,1024,4], ['desktop',1440,900,4]]) {
      const run = `style-${Date.now()}-${players}`; const roster = [];
      for (let seat = 1; seat <= players; seat++) roster.push(await fixture(browser, players, seat, {width,height}, run));
      const owner = roster[0]; const initial = await snapshot(owner); const assigned = styles(initial);
      assert.equal(new Set(assigned).size, players); assert.ok(assigned.every(s => ['first-light','rain-glaze','ember-glaze','drypoint'].includes(s)));
      const target = initial.cards.find(c => c.card.suit === 'diamonds' && c.card.rank === 8 && c.controller === 2);
      assert.ok(target, 'public capture target');
      for (const f of roster) {
        const board = await snapshot(f); assert.deepEqual(styles(board), assigned);
        assert.ok(board.cards.every(c => !['hand','concealed-ace','draw'].includes(c.zone) || c.controller === board.seat), 'authorized concealed projection');
        assert.ok(f.art.has(cardPath(assigned[1], 'diamonds', 8)), 'target renders defender style for every viewer');
        const permitted = new Set([...welcomeArt, ...board.cards.map(c => cardPath(assigned[c.controller-1] || 'first-light', c.card.suit, ({1:'A',11:'J',12:'Q',13:'K'})[c.card.rank] || c.card.rank))]);
        f.permitted = permitted;
        assert.deepEqual([...f.art].filter(p => !permitted.has(p)), [], 'no unauthorized artwork requested');
      }
      await screenshot(owner.page, `${engine}-${label}-before-capture.png`);
      await chooseAttack(owner.page);
      await decisions(owner.page); await activate(owner.page, owner.page.getByRole('button', {name: 'Confirm Attack', exact: true}));
      await owner.page.locator('flt-semantics-host').getByText('Waiting for seat 2', {exact:true}).waitFor();
      for (const f of roster.slice(1)) {
        await f.page.bringToFront(); await refreshed(f); await decisions(f.page);
        const pass = f.page.getByRole('button', {name:'Pass response',exact:true}); await pass.waitFor(); await activate(f.page, pass);
        await pause(250);
      }
      for (const f of roster) {
        await f.page.bringToFront(); await refreshed(f); await table(f.page); await f.page.waitForLoadState('networkidle');
        const board = await snapshot(f); assert.deepEqual(styles(board), assigned); assert.equal(board.window_id, undefined);
        assert.ok(board.cards.every(c => !['hand','concealed-ace','draw'].includes(c.zone) || c.controller === board.seat), 'authorized concealed projection after capture');
        const captured = board.cards.find(c => c.card.suit==='diamonds' && c.card.rank===8 && c.card.deck===target.card.deck);
        assert.ok(captured); assert.equal(captured.controller, 1); assert.equal(captured.zone, 'series');
        assert.ok(f.art.has(cardPath(assigned[0], 'diamonds',8)), 'captured card renders new controller style for every viewer');
        for (const c of board.cards) f.permitted.add(cardPath(assigned[c.controller-1] || 'first-light', c.card.suit, ({1:'A',11:'J',12:'Q',13:'K'})[c.card.rank] || c.card.rank));
        assert.deepEqual([...f.art].filter(p => !f.permitted.has(p)), [], 'no unauthorized artwork during capture');
      }
      await screenshot(owner.page, `${engine}-${label}-after-capture.png`);
      await owner.page.reload(); await table(owner.page);
      await screenshot(owner.page, `${engine}-${label}-after-reload.png`);
      const reloaded = await snapshot(owner); assert.deepEqual(styles(reloaded), assigned);
      assert.ok(reloaded.cards.every(c => !['hand','concealed-ace','draw'].includes(c.zone) || c.controller === reloaded.seat), 'authorized concealed projection after reload');
      assert.deepEqual([...owner.art].filter(p => !owner.permitted.has(p)), [], 'no unauthorized artwork after reload');
      report.cases.push({case:'distinct-shared-styles-capture-identity-privacy-refresh-reload', label, players, viewport:{width,height}, pass:true});
      for (const [file, expected] of Object.entries(report.assets)) {
        const r = await owner.context.request.get(`${origin}/${file}`); assert.equal(hash(await r.body()),expected,`final served ${file}`);
      }
      for (const f of roster) await f.context.close();
      const reference = await fixture(browser, 4, 1, {width,height}, `reference-${Date.now()}`, 'attack');
      await screenshot(reference.page, `${engine}-${label}-reference-table.png`);
      await reference.context.close();
    }
    assert.deepEqual(report.errors, []); assert.equal(hash(fs.readFileSync(__filename)),report.harness);
    report.status = 'passed';
  } catch (e) {report.status='failed'; report.failure=e.message; throw e;}
  finally {fs.writeFileSync(path.join(output,`${engine}-card-styles-report.json`),JSON.stringify(report,null,2)); await browser.close();}
  console.log(JSON.stringify(report));
})().catch(e => {console.error(e); process.exitCode=1;});
