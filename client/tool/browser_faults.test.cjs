const test=require('node:test');
const assert=require('node:assert/strict');
const {ExpectedAbort}=require('./browser_faults.cjs');
const url='http://127.0.0.1:8081/v1/matches/trade/commands/operation-1';
const message='/127.0.0.1:8081/v1/matches/trade/commands/operation-1 due to access control checks.';
test('only an actually aborted exact request in the active recovery page is expected',()=>{
 const fault=new ExpectedAbort();
 assert.equal(fault.consume(message),false);
 fault.begin();fault.record(url);
 assert.equal(fault.consume(message),true);
 assert.equal(fault.consume(message),false);
});
test('unrelated CSP, CORS and application errors remain failures',()=>{
 const fault=new ExpectedAbort();fault.begin();fault.record(url);
 for(const other of [message.replace('operation-1','operation-2'),message.replace('8081','8080'),'Refused to connect because CSP','TypeError: application failed']) assert.equal(fault.consume(other),false);
 assert.equal(new ExpectedAbort().consume(message),false);
});
test('completed fault phase cannot excuse later errors',()=>{
 const fault=new ExpectedAbort();fault.begin();fault.record(url);fault.end();
 assert.equal(fault.consume(message),false);
});
test('console errors always count, including messages without Error text',()=>{
 const {isBrowserError}=require('./browser_faults.cjs');
 assert.equal(isBrowserError('error','blocked by content security policy'),true);
 assert.equal(isBrowserError('error','Failed to load resource'),true);
 assert.equal(isBrowserError('log','Exception: overflowed'),true);
 assert.equal(isBrowserError('log','ordinary status'),false);
});
test('Chromium abort console diagnostic needs exact active request URL and exact message',()=>{
 const fault=new ExpectedAbort();const message='Failed to load resource: net::ERR_FAILED';
 fault.begin();fault.record(url);
 assert.equal(fault.consumeConsole(message,url+'-other'),false);
 assert.equal(fault.consumeConsole('Blocked by CSP',url),false);
 assert.equal(new ExpectedAbort().consumeConsole(message,url),false);
 assert.equal(fault.consumeConsole(message,url),true);
 assert.equal(fault.consumeConsole(message,url),false);
 fault.record(url);fault.end();assert.equal(fault.consumeConsole(message,url),false);
});
test('initial guest challenge requires observed401 and exact session URL/message',()=>{
 const {isInitialSessionChallenge}=require('./browser_faults.cjs');
 const message='Failed to load resource: the server responded with a status of 401 (Unauthorized)';
 const session='http://127.0.0.1:8081/v1/session';
 assert.equal(isInitialSessionChallenge(message,session,session,401),true);
 assert.equal(isInitialSessionChallenge(message,session,session,403),false);
 assert.equal(isInitialSessionChallenge(message,url,session,401),false);
 assert.equal(isInitialSessionChallenge('Blocked by CSP',session,session,401),false);
});
