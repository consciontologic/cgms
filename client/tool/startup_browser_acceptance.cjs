// Run against the documented Compose stack, retaining all persistent volumes.
// Deliberately stops/restarts local project services; run serially with other QA.
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const {execFile, spawn} = require('node:child_process');
const {promisify} = require('node:util');
const {decisions, hand} = require('./reconnect_browser_acceptance.cjs');
const execute = promisify(execFile);
const root = path.resolve(__dirname, '../..');
const origin = 'http://localhost:8080';
const evidence = path.resolve(process.env.CGMS_STARTUP_EVIDENCE || path.join(__dirname, '../.browser-tools/evidence', `startup-${Date.now()}`));
const sha = buffer => crypto.createHash('sha256').update(buffer).digest('hex');
const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
const report = {schema: 'cgms-cold-start-browser-v1', status: 'running', startedAt: new Date().toISOString(), conditions: 'Local Compose, retained PostgreSQL volume, fresh Chromium contexts; stopped backend for delayed readiness and established-stream recovery. No cache deletion, manual reload recovery or restart recovery by user.', cases: [], serviceLogs: [], diagnostics: []};
let phase = 'setup';
let activePage, needsBackendRestore = false;
// safe-run's heartbeat can leave an inherited pipe open briefly after its
// process exits. Completion is its exit status, not the descendant pipe close.
function captureProcess(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {cwd: options.cwd || root, stdio: ['ignore', 'pipe', 'pipe']});
    let stdout = '', stderr = '';
    const timer = setTimeout(() => child.kill('SIGTERM'), options.timeout || 240000);
    child.stdout.on('data', chunk => { stdout += chunk; });
    child.stderr.on('data', chunk => { stderr += chunk; });
    child.on('error', error => { clearTimeout(timer); reject(error); });
    child.on('exit', (code, signal) => {
      clearTimeout(timer);
      child.stdout.destroy(); child.stderr.destroy();
      if (code === 0) resolve({stdout, stderr});
      else reject(Error(`Captured command failed (${code ?? signal}): ${stdout} ${stderr}`));
    });
  });
}
async function operation(tag, command, args) {
  const started = performance.now();
  const result = await captureProcess(path.join(root, 'xops/agent/safe-run.sh'), [tag, '--', command, ...args]);
  const log = result.stderr.match(/log\s*:\s*(\S+)/)?.[1] || result.stdout.match(/log\s*:\s*(\S+)/)?.[1];
  if (log) report.serviceLogs.push({tag, log});
  return Math.round(performance.now() - started);
}
const compose = (tag, ...args) => operation(tag, 'docker', ['compose', '-f', 'deploy/compose/compose.yaml', ...args]);
async function ready(context, timeout = 30000) {
  const deadline = performance.now() + timeout;
  while (performance.now() < deadline) {
    try { if ((await context.request.get(origin + '/health/ready', {timeout: 2000})).status() === 204) return; } catch (_) { /* Explicit bounded dependency probe. */ }
    await pause(100);
  }
  throw Error(`Authority did not become ready within ${timeout} milliseconds`);
}
function observe(page) {
  activePage = page;
  const data = {console: [], pageErrors: [], requests: [], failedRequests: [], responses: [], socketErrors: [], outagePaths: []};
  const start = performance.now();
  const route = url => new URL(url, origin).pathname; // Never retain session data or query strings.
  page.on('websocket', socket => socket.on('socketerror', message => data.socketErrors.push({phase, path: route(socket.url()), message})));
  page.on('pageerror', error => data.pageErrors.push({phase, message: error.message}));
  page.on('console', message => { if (message.type() === 'error') data.console.push({phase, message: message.text(), path: message.location().url ? route(message.location().url) : ''}); });
  page.on('request', request => data.requests.push({phase, method: request.method(), path: route(request.url()), ms: Math.round(performance.now() - start)}));
  page.on('requestfailed', request => data.failedRequests.push({phase, method: request.method(), path: route(request.url()), failure: request.failure()?.errorText, ms: Math.round(performance.now() - start)}));
  page.on('response', response => { if (response.status() >= 400 || route(response.url()).startsWith('/v1/')) data.responses.push({phase, path: route(response.url()), status: response.status(), ms: Math.round(performance.now() - start)}); });
  report.diagnostics.push(data);
  return data;
}
function verifyDiagnostics(data) {
  assert.deepEqual(data.pageErrors, [], 'uncaught browser exceptions');
  for (const entry of data.failedRequests) {
    assert.ok(/outage/.test(entry.phase) && entry.method === 'GET' && data.outagePaths.includes(entry.path) &&
      /^net::ERR_(ABORTED|EMPTY_RESPONSE|CONNECTION_RESET|CONNECTION_REFUSED|TIMED_OUT)$/.test(entry.failure),
    `unexpected failed request ${JSON.stringify(entry)}`);
  }
  for (const entry of data.socketErrors) {
    assert.ok(/outage/.test(entry.phase) && data.outagePaths.includes(entry.path) &&
      /^(Received bad response code: (502|503)|Error during WebSocket handshake: Unexpected response code: (502|503))$/.test(entry.message),
    `unexpected socket error ${JSON.stringify(entry)}`);
  }
  for (const entry of data.console) {
    const status = data.responses.find(r => r.phase === entry.phase && r.path === entry.path && r.status >= 400)?.status;
    const anonymous = entry.path === '/v1/session' && status === 401 && entry.message === 'Failed to load resource: the server responded with a status of 401 (Unauthorized)';
    const failed = data.failedRequests.find(r => r.phase === entry.phase && r.path === entry.path);
    const outageResource = /outage/.test(entry.phase) && data.outagePaths.includes(entry.path) && (
      (failed && entry.message === `Failed to load resource: ${failed.failure}`) ||
      ([502, 503].includes(status) && entry.message === `Failed to load resource: the server responded with a status of ${status} (${status === 502 ? 'Bad Gateway' : 'Service Unavailable'})`));
    const socketUrl = entry.message.match(/^WebSocket connection to '(ws:\/\/localhost:8080\/[^']+)' failed: Error during WebSocket handshake: Unexpected response code: (502|503)$/);
    const outageSocket = /outage/.test(entry.phase) && socketUrl && data.socketErrors.some(r => r.phase === entry.phase && r.path === new URL(socketUrl[1]).pathname && data.outagePaths.includes(r.path));
    assert.ok(anonymous || outageResource || outageSocket, `unexpected browser console diagnostic ${JSON.stringify(entry)}`);
  }
}

async function post(context, endpoint, body, csrf) {
  const response = await context.request.post(origin + endpoint, {headers: {Origin: origin, ...(csrf ? {'X-CSRF-Token': csrf} : {})}, data: body});
  assert.equal(response.status(), 200, `POST ${endpoint} status`);
  return response.json();
}
async function buildMatches(context) {
  const hashes = {};
  for (const file of ['index.html', 'keyboard_bridge.js', 'flutter_bootstrap.js', 'main.dart.js']) {
    const response = await context.request.get(origin + '/' + file);
    assert.equal(response.status(), 200);
    hashes[file] = sha(await response.body());
    assert.equal(hashes[file], sha(fs.readFileSync(path.join(root, 'client/build/web', file))), `served build ${file}`);
  }
  return hashes;
}
async function delayedSession(browser, reload) {
  const context = await browser.newContext({viewport: {width: 390, height: 844}, hasTouch: true});
  await post(context, '/v1/guests', {});
  const page = await context.newPage();
  const data = observe(page);
  data.outagePaths = ['/v1/session'];
  if (reload) {
    phase = 'reload-before-outage';
    await page.goto(origin + '/#/online');
    await page.getByRole('button', {name: 'Create room', exact: true}).waitFor();
  }
  phase = reload ? 'reload-outage' : 'session-outage';
  await compose(phase + '-stop', 'stop', 'backend');
  needsBackendRestore = true;
  const started = performance.now();
  const failedRead = Promise.race([
    page.waitForEvent('requestfailed', {predicate: request => new URL(request.url()).pathname === '/v1/session', timeout: 12000}),
    page.waitForResponse(response => new URL(response.url()).pathname === '/v1/session' && response.status() >= 500, {timeout: 12000}),
  ]);
  if (reload) await page.reload(); else await page.goto(origin + '/#/online');
  await failedRead;
  await page.locator('flt-semantics-host').getByText(/Connecting to the table… attempt/).waitFor();
  const failedAtMs = Math.round(performance.now() - started);
  const operationsBefore = data.requests.filter(r => r.method === 'POST').length;
  const [, readyAtMs, usableAtMs] = await Promise.all([
    compose(phase + '-start', 'start', 'backend').then(value => { needsBackendRestore = false; return value; }),
    ready(context).then(() => Math.round(performance.now() - started)),
    page.getByRole('button', {name: 'Create room', exact: true}).waitFor({timeout: 35000})
      .then(() => Math.round(performance.now() - started)),
  ]);
  assert.equal(await page.getByRole('button', {name: 'Create room', exact: true}).isEnabled(), true);
  assert.equal(data.requests.filter(r => r.method === 'POST').length, operationsBefore, 'recovery never replays a mutation');
  assert.ok(data.requests.filter(r => r.path === '/v1/session').length >= 2, 'automatic session retry observed');
  verifyDiagnostics(data);
  report.cases.push({case: reload ? 'reload-during-backend-outage' : 'delayed-backend-session-readiness', pass: true, failedAtMs, readyAtMs, usableAtMs, automaticReadAttempts: data.requests.filter(r => r.path === '/v1/session').length});
  console.log(`PASS ${phase}: ready ${readyAtMs}ms, usable ${usableAtMs}ms`);
  await context.close();
}
async function reconnect(browser) {
  phase = 'match-setup';
  const contexts = await Promise.all(Array.from({length: 3}, () => browser.newContext({viewport: {width: 390, height: 844}, hasTouch: true})));
  const actors = [];
  for (const context of contexts) actors.push(await post(context, '/v1/guests', {}));
  const room = await post(contexts[0], '/v1/rooms', {command_id: crypto.randomUUID(), capacity: 3, games_per_match: 3}, actors[0].csrf_token);
  for (let seat = 1; seat < 3; seat++) {
    const invitation = await post(contexts[0], `/v1/rooms/${room.id}/invitations`, {}, actors[0].csrf_token);
    await post(contexts[seat], `/v1/rooms/${room.id}/join`, {invitation: invitation.invitation}, actors[seat].csrf_token);
  }
  const match = await post(contexts[0], `/v1/rooms/${room.id}/matches`, {command_id: crypto.randomUUID()}, actors[0].csrf_token);
  const page = await contexts[0].newPage();
  const data = observe(page);
  data.outagePaths = [`/v1/matches/${match.match_id}`, `/v1/matches/${match.match_id}/stream`];
  // Observe successful native WebSocket handshakes without replacing or
  // intercepting the connection. Hiding a notice alone is not recovery proof.
  const network = await contexts[0].newCDPSession(page);
  await network.send('Network.enable');
  const socketPaths = new Map(), openedSockets = [];
  network.on('Network.webSocketCreated', event => socketPaths.set(event.requestId, new URL(event.url).pathname));
  network.on('Network.webSocketHandshakeResponseReceived', event => {
    if (event.response.status === 101 && socketPaths.get(event.requestId) === `/v1/matches/${match.match_id}/stream`) openedSockets.push(event.requestId);
  });
  async function waitForOpenedSocket(count) {
    const deadline = performance.now() + 40000;
    while (openedSockets.length < count && performance.now() < deadline) await pause(100);
    assert.ok(openedSockets.length >= count, 'real authority WebSocket handshake succeeds automatically');
  }
  await page.goto(origin + '/#/online?match=' + encodeURIComponent(match.match_id));
  await hand(page);
  await waitForOpenedSocket(1);
  const before = await (await contexts[0].request.get(`${origin}/v1/matches/${match.match_id}`)).json();
  const beforeVersion = before.version;
  const beforeGame = before.projection.board.game_id;
  phase = 'stream-outage';
  const started = performance.now();
  // Flutter merges this title and description into one text node in an
  // unnamed group; select its observed text, not a nonexistent group name.
  const disconnected = page.locator('flt-semantics-host').getByText(/Connection lost · last confirmed table · actions paused/);
  await Promise.all([
    compose('stream-outage-stop', 'stop', 'backend').then(() => { needsBackendRestore = true; }),
    disconnected.waitFor(),
  ]);
  const [, readyAtMs, usableAtMs] = await Promise.all([
    compose('stream-outage-start', 'start', 'backend').then(value => { needsBackendRestore = false; return value; }),
    ready(contexts[0]).then(() => Math.round(performance.now() - started)),
    Promise.all([waitForOpenedSocket(2), disconnected.waitFor({state: 'hidden', timeout: 40000})])
      .then(() => Math.round(performance.now() - started)),
  ]);
  const after = await (await contexts[0].request.get(`${origin}/v1/matches/${match.match_id}`)).json();
  assert.equal(after.projection.board.game_id, beforeGame);
  assert.ok(after.version >= beforeVersion);
  assert.equal(data.requests.filter(r => r.method === 'POST').length, 0, 'stream recovery submits no commands');
  const action = await decisions(page);
  await action.waitFor();
  assert.equal(await action.isEnabled(), true, 'action workspace is usable after reconnect');
  verifyDiagnostics(data);
  report.cases.push({case: 'established-match-reconnect-without-refresh-or-replayed-command', pass: true, readyAtMs, usableAtMs, persistedGamePreserved: true, commandPosts: 0, successfulSocketHandshakes: openedSockets.length});
  console.log(`PASS ${phase}: ready ${readyAtMs}ms, usable ${usableAtMs}ms`);
  for (const context of contexts) await context.close();
}
async function protocolResponses(browser) {
  const context = await browser.newContext();
  await post(context, '/v1/guests', {});
  phase = 'html-readiness-outage';
  let reads = 0;
  await context.route(origin + '/v1/session', route => ++reads === 1
    ? route.fulfill({status: 503, contentType: 'text/html', body: '<html>Service Unavailable</html>'})
    : route.continue());
  const page = await context.newPage();
  const data = observe(page);
  data.outagePaths = ['/v1/session'];
  await page.goto(origin + '/#/online');
  await page.getByRole('button', {name: 'Create room', exact: true}).waitFor();
  assert.equal(reads, 2, 'HTML proxy failure recovers with a bounded GET retry');
  verifyDiagnostics(data);
  report.cases.push({case: '503-html-proxy-response-recovers', pass: true, reads});
  await page.close();
  await context.unroute(origin + '/v1/session');
  phase = 'malformed-success'; reads = 0;
  await context.route(origin + '/v1/session', route => { reads++; return route.fulfill({status: 200, contentType: 'application/json', body: 'not JSON'}); });
  const bad = await context.newPage();
  const invalid = observe(bad);
  await bad.goto(origin + '/#/online');
  await bad.locator('flt-semantics-host').getByText('The server returned an unreadable response. Use Refresh connection to try again.', {exact: true}).waitFor();
  // Observe beyond the first two retry windows; terminal parsing errors must
  // not be mistaken for delayed readiness. Widget tests cover the full budget.
  await bad.waitForTimeout(1000);
  assert.equal(reads, 1);
  verifyDiagnostics(invalid);
  report.cases.push({case: 'malformed-200-response-is-terminal', pass: true, reads});
  await context.close();
}

async function finalizeStartupRun(report, {capture, restore, close, write}) {
  const errors = [];
  async function attempt(stage, action) {
    try { await action(); }
    catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      errors.push({stage, message});
      report.status = 'failed';
      report.failure ??= `${stage}: ${message}`;
      report.cleanupErrors = errors;
    }
  }
  // A failed screenshot, backend restart or report write must not prevent the
  // other cleanup obligations. Keep the original acceptance failure intact.
  await attempt('capture', capture);
  await attempt('restore', restore);
  await attempt('close', close);
  report.finishedAt = new Date().toISOString();
  await attempt('write', write);
  return errors;
}

async function run() {
  process.env.PLAYWRIGHT_BROWSERS_PATH = path.join(root, 'client/.browser-tools/browsers');
  const {chromium} = require('../.browser-tools/node_modules/playwright');
  fs.mkdirSync(evidence, {recursive: true});
  const browser = await chromium.launch({headless: true});
  report.browser = browser.version();
  const retainedVolume = async () => (await execute('docker', ['volume', 'inspect', 'cgms_postgres_data', '--format', '{{.Name}} {{.CreatedAt}}'], {cwd: root})).stdout.trim();
  let failure;
  try {
    report.volumeBefore = await retainedVolume();
    for (let launch = 1; launch <= 3; launch++) {
      phase = `cold-${launch}`;
      await operation(`startup-cold-${launch}-stop`, 'make', ['services.stop', 'MODE=local']);
      const rows = (await execute('docker', ['ps', '--filter', 'label=com.docker.compose.project=cgms', '--format', '{{.Names}}'], {cwd: root})).stdout.trim();
      assert.equal(rows, '', 'all project services stopped before independent cold launch');
      const context = await browser.newContext({viewport: {width: 390, height: 844}, hasTouch: true});
      const serviceStartedAt = new Date().toISOString();
      const launchStarted = performance.now();
      const [serviceCommandMs, healthReadyMs] = await Promise.all([
        operation(`startup-cold-${launch}-up`, 'make', ['services.up', 'MODE=local']),
        ready(context, 90000).then(() => Math.round(performance.now() - launchStarted)),
      ]);
      const serviceOutput = await execute('docker', ['compose', '-f', 'deploy/compose/compose.yaml', 'logs', '--no-color', '--since', serviceStartedAt, 'backend', 'web', 'postgres', 'redis'], {cwd: root, timeout: 15000, maxBuffer: 4 * 1024 * 1024});
      const serviceLog = path.join(evidence, `cold-${launch}-services.log`);
      fs.writeFileSync(serviceLog, serviceOutput.stdout + serviceOutput.stderr, {mode: 0o600});
      report.serviceLogs.push({tag: `cold-${launch}-services`, log: serviceLog});
      const buildHashes = await buildMatches(context);
      const page = await context.newPage();
      const data = observe(page);
      const started = performance.now();
      await page.goto(origin);
      await page.getByRole('heading', {name: /^(?:TABLE · ALL OPEN CARDS|The public table)$/}).waitFor({timeout: 30000});
      await page.getByRole('heading', {name: /^Your hand · /}).waitFor({timeout: 30000});
      const normalEntryMs = Math.round(performance.now() - started);
      await page.goto(origin + '/#/online');
      const guest = page.getByRole('button', {name: 'Continue as guest', exact: true});
      await guest.waitFor();
      await guest.click();
      await page.getByRole('button', {name: 'Create room', exact: true}).waitFor();
      const onlineUsableMs = Math.round(performance.now() - started);
      assert.equal(data.requests.filter(r => r.method === 'POST' && r.path === '/v1/guests').length, 1);
      verifyDiagnostics(data);
      report.cases.push({case: `cold-launch-${launch}`, pass: true, serviceStartedAt, serviceCommandMs, healthReadyMs, normalEntryMs, onlineUsableMs, buildHashes, guestPosts: 1});
      await context.close();
      console.log(`PASS cold launch ${launch}: health ${healthReadyMs}ms, wrapped command ${serviceCommandMs}ms, entry ${normalEntryMs}ms, online ${onlineUsableMs}ms`);
    }
    await delayedSession(browser, false);
    await delayedSession(browser, true);
    await reconnect(browser);
    await protocolResponses(browser);
    report.volumeAfter = await retainedVolume();
    assert.equal(report.volumeAfter, report.volumeBefore, 'persistent PostgreSQL volume identity is unchanged');
    report.status = 'passed';
  } catch (error) {
    failure = error;
    report.status = 'failed';
    report.failure = error.message;
  } finally {
    const cleanupErrors = await finalizeStartupRun(report, {
      capture: async () => {
        if (failure && activePage && !activePage.isClosed()) {
          report.failureSemantics = await activePage.locator('flt-semantics-host').ariaSnapshot({timeout: 5000});
          await activePage.screenshot({path: path.join(evidence, 'failure.png'), timeout: 5000});
        }
      },
      restore: async () => {
        if (needsBackendRestore) {
          await compose('startup-restore-backend', 'start', 'backend');
          needsBackendRestore = false;
        }
      },
      close: () => browser.close(),
      write: () => fs.writeFileSync(path.join(evidence, 'startup-browser-report.json'), JSON.stringify(report, null, 2) + '\n'),
    });
    if (cleanupErrors.length) {
      console.error('Startup cleanup failures:', JSON.stringify(cleanupErrors));
      if (!failure) failure = Error(`Startup cleanup failed: ${JSON.stringify(cleanupErrors)}`);
    }
  }
  if (failure) throw failure;
  console.log(JSON.stringify({status: report.status, cases: report.cases, report: path.join(evidence, 'startup-browser-report.json')}));
}
module.exports = {verifyDiagnostics, captureProcess, finalizeStartupRun};
if (require.main === module) run().catch(error => { console.error(error); process.exitCode = 1; });
