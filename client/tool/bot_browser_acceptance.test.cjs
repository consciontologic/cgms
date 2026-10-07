const {test} = require('node:test');
const assert = require('node:assert/strict');
const {verifyPrivacy, handDigest, publicState, ownBoundary, openingOutcome, safeDiagnostic, usableTouchTarget, ProgressBudget, effectChoiceLabel, draftPlan} = require('./bot_browser_acceptance.cjs');
const placed = (id, controller = 1, zone = 'hand') => ({controller, zone, card: {id, rank: 4, suit: 'hearts', deck: 1}});
const projection = cards => ({board: {seat: 1, cards, phase: 'playing', active: 1, required_actor: 0}, online: {bot_seats: [2, 3], bot_difficulty: 'standard'}});
test('contextual browser effect choice keeps authority order and rejects unseen identities', () => {
  const board = {cards: [placed('first'), {...placed('second'), card: {id: 'second', rank: 13, suit: 'clubs', deck: 2}}]};
  assert.equal(effectChoiceLabel(board, ['second', 'first']), 'K of Clubs, copy 2, 4 of Hearts, copy 1');
  assert.equal(effectChoiceLabel(board, []), 'Continue without spending a card');
  assert.throws(() => effectChoiceLabel(board, ['hidden']), /authorized visible card/);
  assert.throws(() => effectChoiceLabel(board, undefined), /ordered authorized choice/);
});
test('random-hand draft uses legal Spade payment or a Royal offer and never invents a card', () => {
  const heart = placed('heart');
  const spade = {...placed('spade'), card: {id: 'spade', rank: 10, suit: 'spades', deck: 1}};
  const royal = {...placed('royal'), card: {id: 'royal', rank: 12, suit: 'hearts', deck: 1}};
  const board = {seat: 1, round: 3, cards: [heart, royal, spade]};
  assert.deepEqual(draftPlan(board), {kind: 'purchase', card: spade, field: 'Quantity'});
  assert.deepEqual(draftPlan({...board, cards: [heart, royal, {...spade, available_from_round: 4}]}), {kind: 'offer', card: royal, field: null});
  assert.deepEqual(draftPlan({...board, cards: [heart, {...spade, controller: 2}]}), {kind: 'selection', card: heart, field: null});
  assert.throws(() => draftPlan({...board, cards: [{...spade, zone: 'draw'}]}), /current selectable card/);
});
test('bot acceptance rejects every remote hidden card zone and wrong viewer', () => {
  for (const zone of ['hand', 'concealed-ace', 'draw']) assert.throws(() => verifyPrivacy(projection([placed('hidden', 2, zone)])), /must never reach/);
  const wrong = projection([]); wrong.board.seat = 2; assert.throws(() => verifyPrivacy(wrong), /human seat/);
  assert.deepEqual(verifyPrivacy(projection([placed('own'), placed('public', 2, 'series')])), {visibleCards: 2, ownPrivateCards: 1});
  assert.deepEqual(verifyPrivacy(projection(null)), {visibleCards: 0, ownPrivateCards: 0});
  assert.throws(() => verifyPrivacy(projection({})), /card list/);
});
test('hand digest detects physical changes and ignores public cards and ordering', () => {
  const cards = [placed('private-a'), placed('private-b', 1, 'concealed-ace')];
  const hash = handDigest(projection(cards));
  assert.equal(hash, handDigest(projection([...cards].reverse().concat(placed('public', 2, 'series')))));
  assert.notEqual(hash, handDigest(projection([placed('replacement'), cards[1]])));
  assert.notEqual(hash, handDigest(projection([placed('private-a', 1, 'series'), cards[1]])));
  assert.match(hash, /^[a-f0-9]{64}$/);
});
test('reported state excludes private card identities and requires a usable human boundary', () => {
  const state = {version: 7, projection: projection([placed('secret-card')])};
  assert.ok(ownBoundary(state));
  assert.ok(!JSON.stringify(publicState(state)).includes('secret-card'));
  for (const key of ['window_id', 'required_actor']) {
    state.projection.board[key] = key === 'window_id' ? 'pending' : 1; assert.ok(!ownBoundary(state)); delete state.projection.board[key];
  }
  state.projection.online.automatic_pending = true; assert.ok(!ownBoundary(state));
  state.projection.online.automatic_pending = false; state.projection.online.round_closed = true; assert.ok(!ownBoundary(state));
});
test('provisional active human seat is not ready while the next server move is pending', () => {
  const state = {version: 1, projection: projection([placed('own')])};
  state.projection.board.round = 1; state.projection.board.turn = 1;
  state.projection.online.server_pending = true;
  assert.equal(ownBoundary(state), false, 'a dealt provisional active1 must not qualify as the actual first human turn');
  assert.equal(publicState(state).serverPending, true, 'retained evidence distinguishes next-server readiness');
  state.projection.online.server_pending = false;
  assert.equal(ownBoundary(state), true);
});
test('opening acceptance distinguishes success, proven confinement and authoritative closure', () => {
  const card = placed('original');
  const before = {version: 3, projection: projection([card])};
  const after = {version: 5, projection: projection([placed('rotated-handle', 1, 'series')])};
  after.projection.board.players = [{seat: 1, confined: false}];
  assert.equal(openingOutcome(before, after, [card]), 'opened');
  after.projection.board.cards = [placed('rotated-handle')];
  assert.equal(openingOutcome(before, after, [card]), null, 'unchanged private cards do not prove a successful opening');
  after.projection.board.players[0].confined = true;
  assert.equal(openingOutcome(before, after, [card]), 'confined');
  after.projection.board.cards[0].controller = 2;
  assert.equal(openingOutcome(before, after, [card]), null, 'confinement also requires those same physical cards to remain private');
  after.projection.board.phase = 'settlement';
  assert.equal(openingOutcome(before, after, [card]), 'closed');
  after.version = before.version; assert.equal(openingOutcome(before, after, [card]), null);
});
test('all retained diagnostic strings redact card labels and current-view capabilities', () => {
  const capability = `h_${'x'.repeat(43)}`;
  const message = `Exception: Q of Hearts, copy 2; 10 of Clubs, copy 1; ${capability}`;
  const safe = safeDiagnostic(message);
  assert.equal(safe, 'Exception: [private card]; [private card]; [card capability]');
  assert.ok(!safe.includes(capability));
  assert.equal(safeDiagnostic('HTTP429 unavailable'), 'HTTP429 unavailable');
});
test('touch activation requires48painted pixels in both axes and a visible center', () => {
  const rect = {left: 8, top: 314, right: 382, bottom: 363};
  const clip = {left: 8, top: 114, right: 382, bottom: 485};
  assert.ok(usableTouchTarget({rect, clip}));
  assert.ok(!usableTouchTarget({rect, clip: {...clip, bottom: 350}}), 'short visible strip is not a touch target');
  assert.ok(!usableTouchTarget({rect, clip: {...clip, right: 55}}), 'narrow visible strip is not a touch target');
  assert.ok(!usableTouchTarget({rect, clip: {...clip, right: 100}}), 'tap center must remain painted even when48pixels remain');
  assert.ok(!usableTouchTarget({rect, clip: {...clip, right: 300}}), 'visible center and48pixels do not excuse a partially clipped control');
});
test('progress budget excludes acknowledged human work without resetting cumulative unattended time', () => {
  const budget = new ProgressBudget(0); let now = 0, version = 1;
  budget.check(now, version);
  for (let i = 0; i < 60; i++) {
    budget.beforeHuman(); budget.recordHuman(1500); now += 1500;
    now += 500; budget.check(now, ++version);
  }
  assert.deepEqual(budget.metrics, {elapsedMs: 120000, humanInteractionMs: 90000,
    unattendedMs: 30000, noProgressMs: 0, interventions: 60, version: 61});
});
test('frequent unchanged polls fail after15seconds without authoritative progress', () => {
  const budget = new ProgressBudget(0); budget.check(0, 1);
  for (let now = 500; now < 15000; now += 500) budget.check(now, 1);
  assert.throws(() => budget.check(15000, 1), /15s without authoritative version progress/);
  assert.equal(budget.metrics.humanInteractionMs, 0, 'no-op polling never receives human-time credit');
});
test('continued authority progress still exhausts the90second unattended budget', () => {
  const budget = new ProgressBudget(0);
  for (let now = 0; now < 90000; now += 10000) budget.check(now, now / 10000);
  assert.throws(() => budget.check(90000, 9), /90s cumulative unattended/);
});
test('human work cannot bypass the300second wall or128intervention caps', () => {
  const wall = new ProgressBudget(0); wall.check(0, 1);
  wall.beforeHuman(); wall.recordHuman(279000);
  assert.throws(() => wall.check(300000, 2), /300s overall/);
  const actions = new ProgressBudget(0);
  for (let i = 0; i < 128; i++) {actions.beforeHuman(); actions.recordHuman(1);}
  assert.throws(() => actions.beforeHuman(), /128 human-intervention/);
  assert.throws(() => actions.recordHuman(1), /128 human-intervention/);
});
test('human interaction time is excluded from the no-progress watchdog only when recorded', () => {
  const budget = new ProgressBudget(0); budget.check(0, 1);
  budget.beforeHuman(); budget.recordHuman(20000);
  assert.equal(budget.check(21000, 1).noProgressMs, 1000);
  assert.throws(() => budget.check(35000, 1), /15s without authoritative version progress/);
});
