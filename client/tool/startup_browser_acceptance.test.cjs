const test = require('node:test');
const assert = require('node:assert/strict');
const {verifyDiagnostics, captureProcess, finalizeStartupRun} = require('./startup_browser_acceptance.cjs');
const fresh = () => ({pageErrors: [], failedRequests: [], console: [], responses: [], socketErrors: [], outagePaths: ['/v1/session']});
test('only the observed anonymous-session challenge is expected', () => {
  const data = fresh();
  data.console.push({phase: 'cold-1', path: '/v1/session', message: 'Failed to load resource: the server responded with a status of 401 (Unauthorized)'});
  assert.throws(() => verifyDiagnostics(data));
  data.responses.push({phase: 'cold-1', path: '/v1/session', status: 401});
  assert.doesNotThrow(() => verifyDiagnostics(data));
});
test('expected readiness response must have exact phase path status and text', () => {
  const data = fresh();
  data.console.push({phase: 'session-outage', path: '/v1/session', message: 'Failed to load resource: the server responded with a status of 503 (Service Unavailable)'});
  assert.throws(() => verifyDiagnostics(data));
  data.responses.push({phase: 'session-outage', path: '/v1/session', status: 503});
  assert.doesNotThrow(() => verifyDiagnostics(data));
  data.responses[0].status = 403;
  assert.throws(() => verifyDiagnostics(data));
});
test('network failure exception is limited to the expected outage GET', () => {
  const data = fresh();
  data.failedRequests.push({phase: 'session-outage', method: 'GET', path: '/v1/session', failure: 'net::ERR_ABORTED'});
  assert.doesNotThrow(() => verifyDiagnostics(data));
  for (const change of [{method: 'POST'}, {path: '/main.dart.js'}, {path: '/v1/rooms'}, {phase: 'cold-1'}, {failure: 'net::ERR_BLOCKED_BY_CLIENT'}]) {
    const copy = structuredClone(data);
    Object.assign(copy.failedRequests[0], change);
    assert.throws(() => verifyDiagnostics(copy));
  }
});
test('application errors and unrelated resource diagnostics always fail', () => {
  const data = fresh();
  data.pageErrors.push({phase: 'session-outage', message: 'application failed'});
  assert.throws(() => verifyDiagnostics(data));
  data.pageErrors = [];
  data.console.push({phase: 'session-outage', path: '/v1/session', message: 'Failed to load resource: arbitrary error'});
  assert.throws(() => verifyDiagnostics(data));
});
test('socket failures require exact allowed match stream and handshake error', () => {
  const data = fresh();
  data.outagePaths.push('/v1/matches/m/stream');
  data.socketErrors.push({phase: 'stream-outage', path: '/v1/matches/m/stream', message: 'Received bad response code: 502'});
  assert.doesNotThrow(() => verifyDiagnostics(data));
  data.socketErrors[0].path = '/v1/matches/other/stream';
  assert.throws(() => verifyDiagnostics(data));
});

test('completed process is not timed by an inherited descendant capture pipe', async () => {
  const started = performance.now();
  await captureProcess(process.execPath, ['-e', `require('node:child_process').spawn(process.execPath, ['-e', 'setTimeout(() => {}, 3000)'], {stdio: 'inherit', detached: true}).unref(); process.exit(0);`], {timeout: 1000});
  assert.ok(performance.now() - started < 1500, 'parent exit, not the three-second descendant pipe lifetime, determines completion');
});

test('failed diagnostic capture still restores backend, closes browser and saves original failure', async () => {
  const calls = [], report = {status: 'failed', failure: 'original acceptance failure'};
  let saved;
  const errors = await finalizeStartupRun(report, {
    capture: async () => { calls.push('capture'); throw Error('screenshot unavailable'); },
    restore: async () => { calls.push('restore'); },
    close: async () => { calls.push('close'); },
    write: async () => { calls.push('write'); saved = structuredClone(report); },
  });
  assert.deepEqual(calls, ['capture', 'restore', 'close', 'write']);
  assert.equal(saved.failure, 'original acceptance failure');
  assert.deepEqual(errors, [{stage: 'capture', message: 'screenshot unavailable'}]);
  assert.deepEqual(saved.cleanupErrors, errors);
});

test('failed backend restoration still closes browser and records an unsuccessful run', async () => {
  const calls = [], report = {status: 'passed'};
  let saved;
  const errors = await finalizeStartupRun(report, {
    capture: async () => { calls.push('capture'); },
    restore: async () => { calls.push('restore'); throw Error('backend start failed'); },
    close: async () => { calls.push('close'); },
    write: async () => { calls.push('write'); saved = structuredClone(report); },
  });
  assert.deepEqual(calls, ['capture', 'restore', 'close', 'write']);
  assert.equal(saved.status, 'failed');
  assert.deepEqual(errors, [{stage: 'restore', message: 'backend start failed'}]);
  assert.deepEqual(saved.cleanupErrors, errors);
});

test('report write failure cannot bypass browser cleanup or disappear', async () => {
  let closed = false;
  const report = {status: 'passed'};
  const errors = await finalizeStartupRun(report, {
    capture: async () => {}, restore: async () => {},
    close: async () => { closed = true; },
    write: async () => { assert.equal(closed, true); throw Error('report write failed'); },
  });
  assert.equal(report.status, 'failed');
  assert.deepEqual(errors, [{stage: 'write', message: 'report write failed'}]);
});
