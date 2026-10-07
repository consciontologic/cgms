// Real bot matches through the ordinary lobby. No seeded sessions, fixture
// routes, synthetic cards, client-side bot commands or transport response stubs.
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const {isBrowserError, isInitialSessionChallenge} = require('./browser_faults.cjs');
const {activate, closeActions, closePublicCards, decisions, focusByKeyboard, hand, section, revealCard, selectIntent, gameOption, cancelDraft, waitForCommandPresentation} = require('./reconnect_browser_acceptance.cjs');

const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
const frame = page => page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
const digest = value => crypto.createHash('sha256').update(value).digest('hex');
const fileHash = file => digest(fs.readFileSync(file));
const privateZones = new Set(['hand', 'concealed-ace', 'draw']);
const safeDiagnostic = value => String(value).replace(/\b(?:A|K|Q|J|\d{1,2}) of (?:Hearts|Spades|Diamonds|Clubs), copy \d+/g, '[private card]')
  .replace(/h_[A-Za-z0-9_-]{43}/g, '[card capability]');

class ProgressBudget {
  constructor(startedAt) {
    this.startedAt = startedAt; this.humanMs = 0; this.interventions = 0;
    this.lastVersion = null; this.lastProgressWaitMs = 0; this.metrics = {};
  }
  beforeHuman() {assert.ok(this.interventions < 128, '128 human-intervention cap reached');}
  recordHuman(durationMs) {
    this.beforeHuman(); assert.ok(durationMs >= 0, 'measured human duration is nonnegative');
    this.humanMs += durationMs; this.interventions++;
  }
  check(now, version) {
    const elapsedMs = now - this.startedAt, unattendedMs = elapsedMs - this.humanMs;
    if (this.lastVersion === null || version > this.lastVersion) {
      this.lastVersion = version; this.lastProgressWaitMs = unattendedMs;
    }
    Object.assign(this.metrics, {elapsedMs, humanInteractionMs: this.humanMs, unattendedMs,
      noProgressMs: unattendedMs - this.lastProgressWaitMs, interventions: this.interventions, version});
    assert.ok(elapsedMs < 300000, '300s overall progress budget exceeded');
    assert.ok(unattendedMs < 90000, '90s cumulative unattended server budget exceeded');
    assert.ok(this.metrics.noProgressMs < 15000, '15s without authoritative version progress');
    return this.metrics;
  }
}

function verifyPrivacy(projection, expectedSeat = 1) {
  const board = projection.board;
  assert.equal(board.seat, expectedSeat, 'snapshot belongs to the human seat');
  // The real undealt starting board serializes its empty Go slice as null.
  const cards = board.cards ?? [];
  assert.ok(Array.isArray(cards), 'authorized card projection is a card list or undealt null');
  const hidden = cards.filter(card => privateZones.has(card.zone) && card.controller !== expectedSeat);
  assert.equal(hidden.length, 0, 'opponent hands, concealed Aces and draw identities must never reach this client');
  return {visibleCards: cards.length, ownPrivateCards: cards.filter(card => privateZones.has(card.zone)).length};
}
function handDigest(projection) {
  verifyPrivacy(projection);
  return digest(JSON.stringify(projection.board.cards.filter(card => ['hand', 'concealed-ace'].includes(card.zone))
    .map(card => [card.card.id, card.card.rank, card.card.suit, card.card.deck, card.zone]).sort((a, b) => a[0].localeCompare(b[0]))));
}
function publicState(snapshot) {
  const {board, online} = snapshot.projection;
  return {version: snapshot.version, round: board.round, turn: board.turn, active: board.active,
    requiredActor: board.required_actor, phase: board.phase, window: Boolean(board.window_id),
    decision: Boolean(board.decision_id), automaticPending: online.automatic_pending === true,
    serverPending: online.server_pending === true,
    roundClosed: online.round_closed === true, departureChoicePending: online.departure_pending === true,
    ending: online.ending || null, financialPhase: online.financial_phase, completedGames: online.completed_games,
    botSeats: online.bot_seats, difficulty: online.bot_difficulty, privacy: verifyPrivacy(snapshot.projection)};
}
const ownBoundary = snapshot => snapshot.projection.board.phase === 'playing' && snapshot.projection.board.active === 1 &&
  !snapshot.projection.board.window_id && !snapshot.projection.board.required_actor && !snapshot.projection.online.automatic_pending &&
  !snapshot.projection.online.server_pending && !snapshot.projection.online.round_closed;
function openingOutcome(before, after, cards) {
  if (after.version <= before.version) return null;
  if (after.projection.board.phase !== 'playing') return 'closed';
  const heldIn = zone => cards.every(card => (after.projection.board.cards || []).some(current => current.controller === 1 && current.zone === zone &&
    current.card.rank === card.card.rank && current.card.suit === card.card.suit && current.card.deck === card.card.deck));
  if (!after.projection.board.window_id && heldIn('series')) return 'opened';
  if (after.projection.board.players.some(player => player.seat === 1 && player.confined) && heldIn('hand')) return 'confined';
  return null;
}
const cardName = placed => {
  const {rank, suit, deck} = placed.card;
  return new RegExp(`^${({1: 'A', 11: 'J', 12: 'Q', 13: 'K'})[rank] || rank} of ${suit[0].toUpperCase() + suit.slice(1)}, copy ${deck}(?: |$)`);
};
const usableTouchTarget = ({rect, clip}) =>
  rect.left >= clip.left - 1 / 64 && rect.right <= clip.right + 1 / 64 &&
  rect.top >= clip.top - 1 / 64 && rect.bottom <= clip.bottom + 1 / 64 &&
  Math.min(rect.right, clip.right) - Math.max(rect.left, clip.left) >= 48 - 1 / 64 &&
  Math.min(rect.bottom, clip.bottom) - Math.max(rect.top, clip.top) >= 48 - 1 / 64 &&
  (rect.left + rect.right) / 2 >= clip.left && (rect.left + rect.right) / 2 <= clip.right &&
  (rect.top + rect.bottom) / 2 >= clip.top && (rect.top + rect.bottom) / 2 <= clip.bottom;
async function targetGeometry(target) {
  return target.evaluate(element => {
    const metrics = node => ({tag: node.tagName, role: node.getAttribute('role'),
      top: node.style.top, left: node.style.left, transform: node.style.transform,
      scrollTop: node.scrollTop, scrollLeft: node.scrollLeft,
      scrollHeight: node.scrollHeight, scrollWidth: node.scrollWidth,
      clientHeight: node.clientHeight, clientWidth: node.clientWidth,
      offsetWidth: node.offsetWidth, offsetHeight: node.offsetHeight,
      rect: node.getBoundingClientRect().toJSON()});
    const rect = element.getBoundingClientRect().toJSON();
    const clip = {left: 0, top: 0, right: innerWidth, bottom: innerHeight}, ancestors = [], hierarchy = [];
    for (let parent = element.parentElement; parent; parent = parent.parentElement) {
      const style = getComputedStyle(parent), box = parent.getBoundingClientRect();
      hierarchy.push({...metrics(parent), overflowX: style.overflowX, overflowY: style.overflowY});
      if (!box.width || !box.height) continue;
      if (/auto|scroll|hidden|clip/.test(style.overflowY)) {clip.top = Math.max(clip.top, box.top); clip.bottom = Math.min(clip.bottom, box.bottom);}
      if (/auto|scroll|hidden|clip/.test(style.overflowX)) {clip.left = Math.max(clip.left, box.left); clip.right = Math.min(clip.right, box.right);}
      if (/auto|scroll|hidden|clip/.test(style.overflowX + style.overflowY)) ancestors.push({role: parent.getAttribute('role'), rect: box.toJSON()});
    }
    return {rect, clip, ancestors, target: metrics(element), targetHasFocus: element === document.activeElement, hierarchy};
  });
}

async function selectMenu(page, control, value) {
  await activate(page, control);
  const option = page.getByRole('menuitem', {name: value, exact: true});
  // Dropdown route animation is 300ms in the pinned Flutter SDK. A semantic
  // option can exist before the painted pointer target reaches its position.
  await option.waitFor(); await pause(350);
  await focusByKeyboard(page, option); await page.keyboard.press('Enter');
  await option.waitFor({state: 'hidden'}); await pause(350); await frame(page);
}
const cardLabel = placed => {
  const {rank, suit, deck} = placed.card;
  return `${({1: 'A', 11: 'J', 12: 'Q', 13: 'K'})[rank] || rank} of ${suit[0].toUpperCase() + suit.slice(1)}, copy ${deck}`;
};
function effectChoiceLabel(board, choice) {
  assert.ok(Array.isArray(choice), 'effect supplies an ordered authorized choice');
  if (!choice.length) return 'Continue without spending a card';
  return choice.map(id => {
    const card = board.cards.find(card => card.card.id === id);
    assert.ok(card, 'effect choice references an authorized visible card');
    return cardLabel(card);
  }).join(', ');
}
function draftPlan(board) {
  const held = (board.cards || []).filter(card => card.controller === board.seat && card.zone === 'hand' &&
    (card.available_from_round || 0) <= board.round);
  const payment = held.find(card => card.card.suit === 'spades' && card.card.rank >= 2 && card.card.rank <= 10);
  if (payment) return {kind: 'purchase', card: payment, field: 'Quantity'};
  const royal = held.find(card => card.card.rank >= 11 && card.card.rank <= 13);
  if (royal) return {kind: 'offer', card: royal, field: null};
  assert.ok(held.length, 'normal human hand has a current selectable card');
  return {kind: 'selection', card: held[0], field: null};
}
async function selectPrivateCard(page, target) {
  // Real random hands have horizontal suit overflow. Keyboard focus delegates
  // reveal to Flutter's two-axis hand scroller; the generic form helper only
  // wheels vertically and cannot expose an arbitrary physical copy here.
  await focusByKeyboard(page, target); await frame(page);
  const visible = await target.evaluate(element => {
    const rect = element.getBoundingClientRect();
    let left = Math.max(0, rect.left), top = Math.max(0, rect.top), right = Math.min(innerWidth, rect.right), bottom = Math.min(innerHeight, rect.bottom);
    for (let parent = element.parentElement; parent; parent = parent.parentElement) {
      const style = getComputedStyle(parent), clip = parent.getBoundingClientRect();
      if (!clip.width || !clip.height) continue;
      if (/auto|scroll|hidden|clip/.test(style.overflowX)) {left = Math.max(left, clip.left); right = Math.min(right, clip.right);}
      if (/auto|scroll|hidden|clip/.test(style.overflowY)) {top = Math.max(top, clip.top); bottom = Math.min(bottom, clip.bottom);}
    }
    return {width: right - left, height: bottom - top};
  });
  assert.ok(visible.width >= 48 - 1 / 64 && visible.height >= 48 - 1 / 64, 'keyboard reveals a painted48px physical-card target');
  await page.keyboard.press('Space'); await pause(350); await frame(page);
  assert.equal(await target.getAttribute('aria-current'), 'true', 'requested physical card is selected through keyboard input');
}

async function run() {
  if (!process.env.CGMS_CONTAINER) process.env.PLAYWRIGHT_BROWSERS_PATH = path.resolve(__dirname, '../.browser-tools/browsers');
  const engines = require('../.browser-tools/node_modules/playwright');
  const engine = process.env.CGMS_BROWSER || 'chromium';
  const origin = process.env.CGMS_BROWSER_ORIGIN || 'http://localhost:8080';
  const runId = `online-bots-${engine}-${Date.now()}`;
  const evidence = path.resolve(process.env.CGMS_EVIDENCE_DIR || path.join(__dirname, '../.browser-tools/artifacts', runId));
  fs.mkdirSync(evidence, {recursive: true});
  const report = {run: runId, engine, origin, startedAt: new Date().toISOString(), status: 'running',
    fixture: 'None: ordinary authenticated guest lobby, random production deal, real server bot policy',
    diagnostics: 'All page errors, failed requests and HTTP errors fail except one verified initial session401 per fresh context',
    privacy: 'Only counts and SHA256 digests retained; no sessions, credentials, private card identities or full snapshots in this report',
    harnessSha256: Object.fromEntries(['bot_browser_acceptance.cjs', 'browser_faults.cjs', 'reconnect_browser_acceptance.cjs'].map(file => [file, fileHash(path.join(__dirname, file))])),
    assetSha256: fileHash(path.resolve(__dirname, '../build/web/main.dart.js')),
    cases: [], matches: [], errors: [], expectedAuthenticationChallenges: [], screenshots: [], captureErrors: []};
  const browser = await engines[engine].launch({headless: true});
  report.version = browser.version();
  try {
    for (const config of [
      {difficulty: 'beginner', players: 4, viewport: {width: 390, height: 844}},
      {difficulty: 'standard', players: 4, viewport: {width: 1440, height: 900}},
      {difficulty: 'advanced', players: 4, viewport: {width: 600, height: 900}},
    ]) {
      const started = Date.now();
      const result = {...config, timings: {}, operations: [], projectionsChecked: 0, states: [], transitions: []}; report.matches.push(result);
      const context = await browser.newContext({viewport: config.viewport, hasTouch: true});
      const bundle = await context.request.get(`${origin}/main.dart.js`);
      assert.equal(bundle.status(), 200, 'ordinary preview serves the application bundle');
      result.servedAssetSha256 = digest(await bundle.body());
      assert.equal(result.servedAssetSha256, report.assetSha256, 'served build matches tested workspace bundle');
      const streams = [];
      await context.routeWebSocket(url => url.origin === origin.replace(/^http/, 'ws') && /^\/v1\/matches\/[^/]+\/stream$/.test(url.pathname), route => {
        // Transparent forwarding; only the reconnect check closes this stream.
        streams.push({page: route, server: route.connectToServer()});
      });
      const page = await context.newPage(); page.setDefaultTimeout(15000);
      let initial = true, phase = 'entry';
      const initialConsole = [], pendingResponses = [];
      const manifestRequests = new Map();
      const sessionUrl = `${origin}/v1/session`;
      page.on('pageerror', error => report.errors.push({difficulty: config.difficulty, phase, message: safeDiagnostic(error.message)}));
      page.on('console', message => {
        if (!isBrowserError(message.type(), message.text())) return;
        const entry = {difficulty: config.difficulty, phase, message: safeDiagnostic(message.text()), location: message.location()};
        if (initial) initialConsole.push(entry); else report.errors.push(entry);
      });
      page.on('requestfailed', request => {
        const atMs = Date.now() - started;
        if (manifestRequests.has(request)) Object.assign(manifestRequests.get(request), {failedAtMs: atMs, error: request.failure()?.errorText});
        report.errors.push({difficulty: config.difficulty, phase, atMs,
          failedRequest: {path: new URL(request.url()).pathname, method: request.method(), error: request.failure()?.errorText}});
      });
      page.on('requestfinished', request => {
        if (manifestRequests.has(request)) manifestRequests.get(request).finishedAtMs = Date.now() - started;
      });
      page.on('request', request => {
        if (new URL(request.url()).pathname === '/assets/FontManifest.json') {
          const entry = {phase, startedAtMs: Date.now() - started};
          manifestRequests.set(request, entry); (result.fontManifestRequests ||= []).push(entry);
        }
        if (request.method() !== 'POST') return;
        const body = request.postDataJSON() || {};
        result.operations.push({path: new URL(request.url()).pathname.replace(/\/rooms\/[^/]+/, '/rooms/:room').replace(/\/matches\/[^/]+/, '/matches/:match'),
          type: body.type || null, ...(typeof body.payload?.accept === 'boolean' ? {accept: body.payload.accept} : {}),
          ...(Number.isInteger(body.expected_version) ? {expectedVersion: body.expected_version} : {}),
          idDigest: body.command_id ? digest(body.command_id) : null, atMs: Date.now() - started});
      });
      page.on('response', response => {
        const request = response.request();
        if (manifestRequests.has(request)) Object.assign(manifestRequests.get(request), {status: response.status(), responseAtMs: Date.now() - started});
        if (response.status() >= 400 && !(initial && response.url() === sessionUrl && request.method() === 'GET' && response.status() === 401))
          report.errors.push({difficulty: config.difficulty, phase, httpStatus: response.status(), path: new URL(response.url()).pathname});
        if (request.method() === 'GET' && /^\/v1\/matches\/[^/]+$/.test(new URL(response.url()).pathname) && response.status() === 200)
          pendingResponses.push(response.json().then(snapshot => {verifyPrivacy(snapshot.projection); result.projectionsChecked++;}).catch(error => report.errors.push({difficulty: config.difficulty, phase, privacyError: safeDiagnostic(error.message)})));
      });
      page.on('websocket', socket => socket.on('framereceived', ({payload}) => {
        try {
          const message = JSON.parse(String(payload));
          if (message.projection) {verifyPrivacy(message.projection); result.projectionsChecked++;}
        } catch (error) {report.errors.push({difficulty: config.difficulty, phase, streamError: safeDiagnostic(error.message)});}
      }));
      const capture = async name => {
        await page.waitForLoadState('networkidle'); await frame(page);
        const filename = `online-bots-${engine}-${config.difficulty}-${name}.png`;
        await page.screenshot({path: path.join(evidence, filename)}); report.screenshots.push(filename);
      };
      const challengePromise = page.waitForResponse(response => response.url() === sessionUrl && response.request().method() === 'GET');
      await page.goto(`${origin}/#/online`);
      const challenge = await challengePromise; await challenge.finished();
      await page.getByRole('button', {name: 'Continue as guest', exact: true}).waitFor({timeout: 30000});
      assert.equal(challenge.status(), 401, 'fresh context has no identity');
      initial = false; let consumed = false;
      for (const entry of initialConsole) {
        if (!consumed && isInitialSessionChallenge(entry.message, entry.location.url, sessionUrl, 401)) {
          consumed = true; report.expectedAuthenticationChallenges.push(entry);
        } else report.errors.push(entry);
      }
      await activate(page, page.getByRole('button', {name: 'Continue as guest', exact: true}));
      await page.getByRole('button', {name: 'Create room', exact: true}).waitFor();
      result.timings.guestReadyMs = Date.now() - started;
      // Deliberate RED gate on the prior build: the ordinary lobby had no bot
      // entry. No room or match is created before this assertion succeeds.
      await page.getByRole('button', {name: 'Play against bots', exact: true}).waitFor();
      if (config.players === 3) await selectMenu(page, page.getByRole('button', {name: 'Players 4', exact: true}), '3');
      const games = page.getByRole('textbox', {name: 'Games per match', exact: true});
      await section(page, games.locator('xpath=parent::flt-semantics')); await focusByKeyboard(page, games); await page.keyboard.press('ControlOrMeta+A'); await page.keyboard.type('1');
      assert.equal(await games.inputValue(), '1');
      if (config.difficulty !== 'beginner') await selectMenu(page, page.getByRole('button', {name: 'Bot difficulty Beginner', exact: true}), config.difficulty[0].toUpperCase() + config.difficulty.slice(1));
      const roomPromise = page.waitForResponse(response => response.url() === `${origin}/v1/rooms` && response.request().method() === 'POST');
      phase = 'create-room'; await activate(page, page.getByRole('button', {name: 'Play against bots', exact: true}));
      const roomResponse = await roomPromise; assert.equal(roomResponse.status(), 200, 'bot room created');
      const room = await roomResponse.json();
      assert.equal(room.bot_difficulty, config.difficulty); assert.equal(room.capacity, config.players); assert.equal(room.members.length, config.players);
      assert.equal(room.games, 1, 'existing match-length control configures the bot room');
      assert.equal(room.members.filter(member => member.bot === true).length, config.players - 1);
      assert.ok(room.members.every(member => (member.bot === true) === (member.seat !== 1)), 'only opponent seats are server bots');
      assert.equal(await page.getByRole('button', {name: 'Create invitation', exact: true}).count(), 0, 'bot room does not invite human replacements');
      await page.getByRole('button', {name: 'Start match', exact: true}).waitFor();
      assert.equal(await page.getByRole('button', {name: 'Start match', exact: true}).isEnabled(), true, 'bots fill room immediately');
      await capture('lobby');
      result.timings.roomReadyMs = Date.now() - started;
      await page.waitForLoadState('networkidle');
      (result.reloads ||= []).push({phase: 'room-reload', startedAtMs: Date.now() - started}); phase = 'room-reload';
      await page.reload(); await page.getByRole('button', {name: 'Start match', exact: true}).waitFor({timeout: 30000});
      const startPromise = page.waitForResponse(response => response.url().endsWith(`/rooms/${room.id}/matches`) && response.request().method() === 'POST');
      phase = 'start-match'; await activate(page, page.getByRole('button', {name: 'Start match', exact: true}));
      const startResponse = await startPromise; assert.equal(startResponse.status(), 200);
      await page.waitForURL(/match=/);
      const matchId = new URLSearchParams(page.url().split('?')[1]).get('match');
      assert.ok(matchId, 'match URL is shareable without an invitation or credential');
      result.matchId = matchId;
      const matchUrl = `${origin}/v1/matches/${encodeURIComponent(matchId)}`;
      let lastSnapshotReadAt = 0;
      const snapshot = async () => {
        // Enforce pacing across every caller, including immediate before/after
        // reads around an acknowledged command. Never retry a rejected read.
        await pause(Math.max(0, 500 - (Date.now() - lastSnapshotReadAt)));
        lastSnapshotReadAt = Date.now();
        const response = await context.request.get(matchUrl, {timeout: 15000}); assert.equal(response.status(), 200, 'authenticated own snapshot');
        const data = await response.json(); verifyPrivacy(data.projection); result.projectionsChecked++;
        assert.deepEqual(data.projection.online.bot_seats, Array.from({length: config.players - 1}, (_, i) => i + 2));
        assert.equal(data.projection.online.bot_difficulty, config.difficulty);
        result.lastState = publicState(data);
        if (result.transitions.at(-1)?.version !== data.version) result.transitions.push({...result.lastState, atMs: Date.now() - started});
        return data;
      };
      const until = async (predicate, label, timeout = 30000) => {
        const deadline = Date.now() + timeout;
        let current;
        // The ordinary API admits five requests/second per IP and account.
        // Sample at two/second so UI reads/commands retain their own budget.
        do {current = await snapshot(); if (predicate(current)) return current; await pause(500);} while (Date.now() < deadline);
        throw Error(`${label} exceeded ${timeout}ms: ${JSON.stringify(publicState(current))}`);
      };
      const activateWithReply = async target => {
        const [reply] = await Promise.all([
          page.waitForResponse(response => response.url() === `${matchUrl}/commands` && response.request().method() === 'POST'),
          activate(page, target),
        ]);
        await waitForCommandPresentation(page, reply);
        return reply;
      };
      const progressBudget = (label, startedAt = Date.now()) => {
        const budget = new ProgressBudget(startedAt);
        (result.progressBudgets ||= {})[label] = budget.metrics;
        return budget;
      };
      const answerHuman = async (state, budget) => {
        const {board, online} = state.projection;
        if (online.automatic_pending || online.server_pending) return false;
        if (!online.departure_pending && board.required_actor !== 1) return false;
        budget.beforeHuman(); const humanStarted = Date.now();
        const count = result.operations.length;
        let type, reply;
        if (online.departure_pending) {
          assert.equal(online.round_closed, true, 'Stay requires an authoritative round boundary');
          const stay = page.getByRole('button', {name: 'Stay in game', exact: true});
          await stay.waitFor(); result.stayBeforeFocus = await targetGeometry(stay);
          await focusByKeyboard(page, stay); await frame(page);
          // The full-width round-choice control is directly touchable. Prove
          // its fully painted48px target and center instead of wheeling a form that
          // is already visible; retain numeric bounds if that proof fails.
          const visibleDeadline = Date.now() + 15000;
          do {result.stayGeometry = await targetGeometry(stay); if (usableTouchTarget(result.stayGeometry)) break; await frame(page);} while (Date.now() < visibleDeadline);
          assert.ok(usableTouchTarget(result.stayGeometry), `Stay must have a painted48px touch target: ${JSON.stringify(result.stayGeometry)}`);
          (result.stayGeometries ||= []).push(result.stayGeometry);
          [reply] = await Promise.all([
            page.waitForResponse(response => response.url() === `${matchUrl}/commands` && response.request().method() === 'POST'),
            stay.tap(),
          ]);
          type = 'departure-choice';
          result.stayChoices = (result.stayChoices || 0) + 1;
        } else if (board.required_actor === 1 && board.decision_id) {
          await decisions(page);
          // Choose only the current authority's complete ordered choice. Justice
          // has a separate concealed outcome and never requires a hidden ID.
          const closedJustice = board.decision_kind === 'justice';
          await cancelDraft(page);
          await decisions(page);
          const choice = closedJustice ? 'Choose concealed Justice outcome' : effectChoiceLabel(board, board.choices?.[0]);
          await activate(page, page.getByRole('button', {name: choice, exact: true}));
          reply = await activateWithReply(page.getByRole('button', {name: `Confirm ${closedJustice ? 'Decision Closed' : 'Decision'}`, exact: true}));
          type = closedJustice ? 'decision-closed' : 'decision';
          result.effectDecisions = (result.effectDecisions || 0) + 1;
        } else if (board.required_actor === 1) {
          const pass = page.getByRole('button', {name: 'Pass response', exact: true});
          // Newly required responses expand after a Flutter frame. Waiting for
          // the control avoids racing that automatic expansion closed.
          await pass.waitFor(); await pause(200); await frame(page); reply = await activateWithReply(pass); type = 'pass';
        } else return false;
        assert.equal(reply.status(), 200, `human ${type} acknowledged by authority`);
        const acknowledgement = await reply.json();
        assert.ok(acknowledgement.version > state.version, 'human acknowledgement advances its original version');
        await until(next => next.version >= acknowledgement.version, `human ${type} accepted`);
        assert.equal(result.operations.length, count + 1, 'one human activation sends one command');
        assert.equal(result.operations.at(-1).type, type);
        if (type === 'departure-choice') assert.equal(result.operations.at(-1).accept, false, 'Stay does not surrender the human seat');
        result.operations.at(-1).acknowledgedVersion = acknowledgement.version;
        // Exclude only a real, acknowledged human interaction. No-op helper
        // calls and all waiting for automatic authority progress still count.
        budget.recordHuman(Date.now() - humanStarted);
        return true;
      };
      const finishClosed = async state => {
        assert.notEqual(state.projection.board.phase, 'playing', 'terminal branch requires authoritative board closure');
        assert.ok(state.projection.online.ending, 'authority records the ending reason');
        const budget = progressBudget('settlement');
        for (;;) {
          budget.check(Date.now(), state.version);
          if (state.projection.online.completed_games === 1) {
            assert.equal(state.projection.online.financial_phase, 'finalized');
            assert.equal(state.projection.online.ranks.length, config.players, 'completed match has all standings');
            await page.locator('flt-semantics-host').getByText('Match complete', {exact: false}).waitFor();
            report.cases.push({case: `${config.difficulty}-authoritative-early-ending-settles-and-shows-standings`, pass: true, ending: state.projection.online.ending});
            return state;
          }
          if (!state.projection.online.automatic_pending && !state.projection.online.server_pending && !state.projection.online.finished) {
            budget.beforeHuman(); const humanStarted = Date.now();
            await gameOption(page, 'Finish Settlement');
            const count = result.operations.length;
            const reply = await activateWithReply(page.getByRole('button', {name: 'Confirm Finish Settlement', exact: true}));
            assert.equal(reply.status(), 200, 'human settlement acknowledged');
            const acknowledgement = await reply.json();
            assert.ok(acknowledgement.version > state.version, 'settlement receipt advances its original version');
            await until(next => next.version >= acknowledgement.version, 'human settlement accepted');
            assert.equal(result.operations.length, count + 1, 'settlement confirmation is one human command');
            assert.equal(result.operations.at(-1).type, 'finish-settlement');
            result.operations.at(-1).acknowledgedVersion = acknowledgement.version;
            budget.recordHuman(Date.now() - humanStarted);
          } else await pause(500);
          state = await snapshot();
        }
      };
      const reachHuman = async label => {
        const budget = progressBudget(label);
        for (;;) {
          const state = await snapshot();
          budget.check(Date.now(), state.version);
          if (state.projection.board.phase !== 'playing') return finishClosed(state);
          // The dealt projection can precede the first BeginTurn. Require the
          // app's streamed state to agree before preparing a human action;
          // another seat may actually start and request a human response.
          const endTurn = page.getByRole('button', {name: 'End Turn', exact: true});
          if (ownBoundary(state) && await endTurn.count() && await endTurn.isEnabled()) return state;
          if (!(await answerHuman(state, budget))) await pause(500);
        }
      };
      const initialOperations = result.operations.length;
      const first = await reachHuman('random starter reaches first human turn');
      result.initialHumanResponses = result.operations.length - initialOperations;
      if (first.projection.board.phase !== 'playing') {await capture('early-match-complete'); await Promise.all(pendingResponses); await context.close(); continue;}
      await page.getByRole('button', {name: 'End Turn', exact: true}).waitFor();
      await page.waitForLoadState('networkidle'); await frame(page);
      result.timings.playableMs = Date.now() - started; result.states.push(publicState(first));
      const beforeDraftPosts = result.operations.length;
      const draft = draftPlan(first.projection.board);
      result.draftKind = draft.kind;
      await decisions(page);
      await cancelDraft(page);
      await closeActions(page);
      await hand(page);
      const selectedCard = draft.card;
      const selectedTarget = page.locator('[flt-semantics-identifier="private-hand-viewport"]').getByRole('button', {name: cardName(selectedCard)});
      await selectPrivateCard(page, selectedTarget);
      await decisions(page);
      if (draft.kind === 'purchase') await selectIntent(page, 'Purchase');
      if (draft.kind === 'offer') {
        await selectIntent(page, 'Offer');
        await activate(page, page.getByRole('button', {name: 'Bot 3', exact: true}));
      }
      const draftField = draft.field ? page.getByRole('textbox', {name: draft.field, exact: true}) : null;
      if (draftField) {
        // Native INPUT borders exceed the painted field: expose its parent,
        // then enter the retained numeric draft through real keyboard input.
        await section(page, draftField.locator('xpath=parent::flt-semantics'));
        await focusByKeyboard(page, draftField); await page.keyboard.press('ControlOrMeta+A'); await page.keyboard.type('3');
        assert.equal(await draftField.inputValue(), '3', 'numeric draft is entered before resize/reconnect');
      }
      const selected = draft.kind === 'selection' ? selectedTarget
        : page.getByRole('button', {name: `Confirm ${draft.kind === 'purchase' ? 'Purchase' : 'Offer'}`, exact: true});
      await selected.waitFor();
      const beforeDraft = await snapshot();
      const handHash = handDigest(beforeDraft.projection);
      for (const width of [599, 600, 1049, 1050, 390]) {
        await page.setViewportSize({width, height: 900}); await frame(page);
        await decisions(page); await selected.waitFor();
        if (draftField) assert.equal(await draftField.inputValue(), '3', 'resize preserves numeric draft');
        assert.equal(await selectedTarget.getAttribute('aria-current'), 'true', 'resize preserves the same physical card selection');
        assert.equal(handDigest((await snapshot()).projection), handHash, 'resize preserves physical hand');
      }
      phase = 'reconnect';
      const connections = streams.length; assert.ok(connections > 0, 'real authority stream connected');
      const reconnectRead = page.waitForResponse(response => response.url() === matchUrl && response.request().method() === 'GET');
      const reconnectAt = Date.now();
      for (const stream of [...streams]) {
        await stream.page.close({code: 1012, reason: 'bot acceptance reconnect'});
        await stream.server.close({code: 1012, reason: 'bot acceptance reconnect'});
      }
      assert.equal((await reconnectRead).status(), 200, 'automatic reconnect fetches authority');
      await page.locator('flt-semantics-host').getByText('Connection lost · last confirmed table · actions paused', {exact: false}).waitFor({state: 'hidden'});
      const streamDeadline = Date.now() + 10000;
      while (streams.length === connections && Date.now() < streamDeadline) await pause(50);
      assert.ok(streams.length > connections, 'automatic reconnect reopens real authority stream');
      await decisions(page); await selected.waitFor();
      if (draftField) assert.equal(await draftField.inputValue(), '3', 'reconnect preserves numeric draft');
      assert.equal(await selectedTarget.getAttribute('aria-current'), 'true', 'reconnect preserves the same physical card selection');
      assert.equal(handDigest((await snapshot()).projection), handHash, 'reconnect preserves physical hand');
      assert.equal((await snapshot()).version, beforeDraft.version, 'resize/reconnect never plays for the human');
      assert.equal(result.operations.length, beforeDraftPosts, 'drafts, resizes and reconnect issue no commands');
      result.timings.reconnectMs = Date.now() - reconnectAt;
      report.cases.push({case: `${config.difficulty}-live-boundaries-and-stream-reconnect-retain-draft-selection-and-private-hand`, pass: true, handDigest: handHash, draftKind: draft.kind, numericField: draft.field});
      // Complete the current document's asset load before deliberately replacing
      // it. Any failed request remains unexpected; no cancellation allowlist.
      await page.waitForLoadState('networkidle');
      result.reloads.push({phase: 'match-reload', startedAtMs: Date.now() - started}); phase = 'match-reload';
      await page.reload(); await page.getByRole('button', {name: 'End Turn', exact: true}).waitFor({timeout: 30000});
      assert.equal(handDigest((await snapshot()).projection), handHash, 'reload preserves physical hand and match');
      assert.equal(result.operations.length, beforeDraftPosts, 'reload never resubmits a command');
      await page.setViewportSize(config.viewport); await frame(page);
      phase = 'human-opening';
      const opening = await snapshot();
      const suit = ['hearts', 'spades', 'clubs'].find(suit => opening.projection.board.cards.some(card => card.controller === 1 && card.zone === 'hand' && card.card.suit === suit && card.card.rank >= 2 && card.card.rank <= 10));
      assert.ok(suit, 'normal initial deal supplies a legal unopened number suit');
      const openingCards = opening.projection.board.cards.filter(card => card.controller === 1 && card.zone === 'hand' && card.card.suit === suit && card.card.rank >= 2 && card.card.rank <= 10);
      // A single physical number card suggests opening its whole held suit.
      await decisions(page);
      await cancelDraft(page);
      await closeActions(page);
      await hand(page);
      await selectPrivateCard(page, page.locator('[flt-semantics-identifier="private-hand-viewport"]').getByRole('button', {name: cardName(openingCards[0])}));
      await selectIntent(page, 'Open Series');
      const beforeOpeningPosts = result.operations.length, openingAt = Date.now();
      const openingBudget = progressBudget('human-opening', openingAt); openingBudget.beforeHuman();
      const openingReply = await activateWithReply(page.getByRole('button', {name: 'Confirm Open Series', exact: true}));
      assert.equal(openingReply.status(), 200, 'human opening acknowledged');
      const openingAcknowledgement = await openingReply.json();
      assert.ok(openingAcknowledgement.version > opening.version, 'opening receipt advances its original version');
      assert.equal(result.operations.length, beforeOpeningPosts + 1, 'one opening activation sends one command before human-time credit');
      assert.equal(result.operations.at(-1).type, 'open-series');
      result.operations.at(-1).acknowledgedVersion = openingAcknowledgement.version;
      openingBudget.recordHuman(Date.now() - openingAt);
      let resolved, outcome;
      for (;;) {
        const current = await snapshot(); openingBudget.check(Date.now(), current.version);
        outcome = openingOutcome(opening, current, openingCards);
        if (outcome) {resolved = current; break;}
        if (!(await answerHuman(current, openingBudget))) await pause(500);
      }
      assert.ok(resolved, 'bots resolve or legally cancel the human opening within bounded progress');
      assert.equal(result.operations.slice(beforeOpeningPosts).filter(operation => operation.type === 'open-series').length, 1, 'one human opening activation sends one opening command');
      assert.ok(resolved.version > opening.version + 1, 'authority makes a bot response without a client bot command');
      if (outcome === 'opened') assert.ok(ownBoundary(resolved), 'successful human opening retains the turn');
      result.timings.botResponsesMs = Date.now() - openingAt; result.states.push(publicState(resolved));
      report.cases.push({case: `${config.difficulty}-${config.players}p-human-opening-and-automatic-bot-responses`, pass: true, outcome, responseVersionAdvance: resolved.version - opening.version, selectedCardCount: openingCards.length});
      if (outcome === 'closed') {await finishClosed(resolved); await capture('early-match-complete'); await Promise.all(pendingResponses); await context.close(); continue;}
      if (outcome === 'confined') {
        resolved = await reachHuman('legal confinement returns control');
        if (resolved.projection.board.phase !== 'playing') {await capture('early-match-complete'); await Promise.all(pendingResponses); await context.close(); continue;}
      }
      await closeActions(page); await capture('human-opening-resolved');
      phase = 'bot-turns';
      const endTurnPosts = result.operations.length, cycleAt = Date.now();
      const cycleBudget = progressBudget('human-turn-cycle', cycleAt); cycleBudget.beforeHuman();
      await section(page, page.getByRole('button', {name: 'End Turn', exact: true}));
      const [commandReply] = await Promise.all([
        page.waitForResponse(response => response.url() === `${matchUrl}/commands` && response.request().method() === 'POST'),
        page.getByRole('button', {name: 'End Turn', exact: true}).tap(),
      ]);
      assert.equal(commandReply.status(), 200, 'touch end turn accepted');
      const endAcknowledgement = await commandReply.json();
      assert.ok(endAcknowledgement.version > resolved.version, 'end-turn receipt advances its original version');
      assert.equal(result.operations.length, endTurnPosts + 1, 'one touch activation sends one end-turn');
      assert.equal(result.operations.at(-1).type, 'end-turn');
      result.operations.at(-1).acknowledgedVersion = endAcknowledgement.version;
      cycleBudget.recordHuman(Date.now() - cycleAt);
      let humanResponses = 0, returned;
      for (;;) {
        const current = await snapshot();
        cycleBudget.check(Date.now(), current.version);
        if (current.projection.board.turn > resolved.projection.board.turn && ownBoundary(current)) {returned = current; break;}
        if (current.projection.board.phase !== 'playing') {returned = await finishClosed(current); break;}
        if (await answerHuman(current, cycleBudget)) humanResponses++; else await pause(500);
      }
      assert.ok(returned, 'all bot seats finish their turns and return control within bounded progress');
      const terminal = returned.projection.board.phase !== 'playing';
      // Initiative can move the human earlier in the following round; a valid
      // return is a later turn, not a fixed capacity-sized turn counter delta.
      if (!terminal) assert.ok(returned.projection.board.turn > resolved.projection.board.turn, 'control returns on a later human turn');
      if (result.stayChoices && !terminal) assert.ok(returned.projection.board.round > opening.projection.board.round, 'explicit Stay advances into the next round');
      const cycleCommands = result.operations.slice(endTurnPosts).filter(operation => operation.type);
      assert.ok(returned.version > resolved.version + cycleCommands.length, 'authority progresses without client bot commands');
      const sentCommands = result.operations.filter(operation => operation.type);
      assert.equal(new Set(sentCommands.map(operation => operation.idDigest)).size, sentCommands.length, 'no duplicate human operation IDs');
      assert.ok(sentCommands.every(operation => ['open-series', 'end-turn', 'pass', 'decision', 'decision-closed', 'departure-choice', 'finish-settlement'].includes(operation.type)), 'browser sends only explicitly activated human actions');
      result.timings.botTurnCycleMs = Date.now() - cycleAt;
      result.humanInterventions = humanResponses;
      result.humanResponses = cycleCommands.filter(operation => operation.type === 'pass').length;
      result.states.push(publicState(returned));
      if (!terminal) {
        await closeActions(page); await page.getByRole('button', {name: 'End Turn', exact: true}).waitFor();
        const endTurn = page.getByRole('button', {name: 'End Turn', exact: true});
        // API reads can lead the streamed Flutter frame. Require the visible
        // app itself to reach readiness rather than testing one premature frame.
        await page.waitForFunction(element => element.getAttribute('aria-disabled') !== 'true', await endTurn.elementHandle(), {timeout: 15000});
        assert.equal(await endTurn.isEnabled(), true, 'human regains usable controls');
      }
      await capture(terminal ? 'early-match-complete' : 'human-turn-returned');
      await Promise.all(pendingResponses);
      report.cases.push({case: `${config.difficulty}-${config.players}p-bots-play-${terminal ? 'complete-match' : 'return-control'}-and-privacy`, pass: true,
        humanResponses: result.humanResponses, humanCommandCount: cycleCommands.length,
        botVersionAdvance: returned.version - resolved.version - cycleCommands.length, elapsedMs: result.timings.botTurnCycleMs});
      if (config.difficulty === 'beginner') {
        // Keep the one-human/three-bot room alive through its natural board
        // closure and financial finalization. Every human boundary still uses
        // the visible End Turn, explicit response/Stay and settlement controls.
        let finalState = returned, turnsEnded = 0;
        while (finalState.projection.board.phase === 'playing') {
          if (++turnsEnded > 64) throw Error('Natural bot match exceeded64human turns');
          finalState = await reachHuman(`complete-match-boundary-${turnsEnded}`);
          if (finalState.projection.board.phase !== 'playing') break;
          await closeActions(page);
          const reply = await activateWithReply(page.getByRole('button', {name: 'End Turn', exact: true}));
          assert.equal(reply.status(), 200, 'full-match explicit End Turn acknowledged');
          const acknowledgement = await reply.json();
          result.operations.at(-1).acknowledgedVersion = acknowledgement.version;
          await until(next => next.version >= acknowledgement.version, 'full-match end turn receipt visible');
          finalState = await reachHuman(`complete-match-next-${turnsEnded}`);
        }
        if (finalState.projection.online.completed_games !== 1) finalState = await finishClosed(finalState);
        assert.equal(finalState.projection.online.financial_phase, 'finalized');
        assert.equal(finalState.projection.online.ranks.length, 4);
        result.completeMatch = {turnsEnded, ending: finalState.projection.online.ending,
          completedGames: finalState.projection.online.completed_games, finalVersion: finalState.version};
        await capture('natural-match-settlement-complete');
        report.cases.push({case: 'one-human-three-bots-natural-board-closure-and-complete-settlement', pass: true, ...result.completeMatch});
      }
      await context.close();
    }
    for (const match of report.matches) {
      const commands = match.operations.filter(operation => operation.type);
      assert.equal(new Set(commands.map(command => command.idDigest)).size, commands.length, 'every outcome preserves unique human operation IDs');
    }
    assert.deepEqual(report.errors, [], 'no unexpected browser diagnostics or privacy violations');
    report.status = 'passed';
  } catch (error) {
    report.status = 'failed'; report.failure = safeDiagnostic(error.stack || error.message);
    for (const [index, context] of browser.contexts().entries()) for (const page of context.pages()) {
      const filename = `online-bots-${engine}-failure-${index}.png`;
      try {await page.screenshot({path: path.join(evidence, filename)}); report.screenshots.push(filename);}
      catch (captureError) {report.captureErrors.push({stage: 'screenshot', message: safeDiagnostic(captureError.message)});}
      // Do not print the full semantic tree: it contains the human's private
      // cards. The ignored failure capture stays local for deliberate review.
    }
    throw new Error(report.failure);
  } finally {
    report.finishedAt = new Date().toISOString();
    // Final serialization also covers diagnostic URL/location/error fields.
    fs.writeFileSync(path.join(evidence, `${engine}-bots-report.json`), safeDiagnostic(JSON.stringify(report, null, 2)));
    await browser.close();
  }
  console.log(JSON.stringify({status: report.status, engine, cases: report.cases.length, matches: report.matches.length, report: path.join(evidence, `${engine}-bots-report.json`)}));
}

module.exports = {verifyPrivacy, handDigest, publicState, ownBoundary, openingOutcome, safeDiagnostic, usableTouchTarget, ProgressBudget, effectChoiceLabel, draftPlan};
if (require.main === module) run().catch(error => {console.error(error); process.exitCode = 1;});
