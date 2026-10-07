// Real browser / real authority smoke and adaptive matrix; synthetic fixtures.
// The QA-only authority must be launched with xops/postgres_integration.py --webfixture.
const path=require('node:path');
const fs=require('node:fs');
const {ExpectedAbort,isBrowserError,isInitialSessionChallenge}=require('./browser_faults.cjs');
const recoveryFaults=new WeakMap();
const privateLobbyPages=new WeakSet();
const invitationValues=new Set();
const safeDiagnostic=value=>{let text=require('./bot_browser_acceptance.cjs').safeDiagnostic(value);for(const invitation of invitationValues) text=text.replaceAll(invitation,'[invitation redacted]');return text;};
if(!process.env.CGMS_CONTAINER) process.env.PLAYWRIGHT_BROWSERS_PATH=path.resolve(__dirname,'../.browser-tools/browsers');
const engines=require('../.browser-tools/node_modules/playwright');
const origin=process.env.CGMS_BROWSER_ORIGIN||'http://127.0.0.1:8081';
const engine=process.env.CGMS_BROWSER||'chromium';
const players=Number(process.env.CGMS_PLAYERS||4);
const otherSeats=Array.from({length:players-1},(_,i)=>i+2);
const acceptorResponses=[...otherSeats.filter(s=>s!==2),1];
const run=`${engine}-${Date.now()}`;
const evidence=process.env.CGMS_EVIDENCE_DIR ? path.resolve(process.env.CGMS_EVIDENCE_DIR) : path.resolve(__dirname,'../.browser-tools/artifacts',run);fs.mkdirSync(evidence,{recursive:true});
const report={diagnostics:'pageerror-and-console-error-all-pages-v1',engine,run,status:'running',viewport:{width:Number(process.env.CGMS_WIDTH||(process.env.CGMS_FINISH_ONLY?320:390)),height:Number(process.env.CGMS_HEIGHT||(process.env.CGMS_FINISH_ONLY?800:844))},input:'touch and keyboard',players,fixture:'synthetic rule-boundary positions plus normal guest room deal',assetSha256:require('node:crypto').createHash('sha256').update(fs.readFileSync(path.resolve(__dirname,'../build/web/main.dart.js'))).digest('hex'),cases:[],errors:[],expectedFaults:[],expectedAuthenticationChallenges:[]};
report.startedAt=new Date().toISOString();
report.screenshots=[];
report.harnessSha256=Object.fromEntries(['browser_acceptance.cjs','browser_faults.cjs','reconnect_browser_acceptance.cjs'].map(file=>[file,require('node:crypto').createHash('sha256').update(fs.readFileSync(path.join(__dirname,file))).digest('hex')]));
report.webAssetsSha256=Object.fromEntries(['index.html','keyboard_bridge.js','flutter_bootstrap.js','main.dart.js'].map(file=>[file,require('node:crypto').createHash('sha256').update(fs.readFileSync(path.resolve(__dirname,'../build/web',file))).digest('hex')]));
const pause=ms=>new Promise(resolve=>setTimeout(resolve,ms));
// Let Flutter complete post-frame pane focus and DOM-to-framework focus events
// before issuing a keyboard command to the next control.
const focusFrame=page=>page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));
async function focusByKeyboard(page,target,key='Tab') {
  const visited=new Set();
  for(let step=0;step<80;step++) {
    if(await target.evaluateAll(elements=>elements.some(element=>element===document.activeElement))) return;
    visited.add(await page.evaluate(()=>{const e=document.activeElement;return `${e?.tagName}#${e?.id}: ${e?.getAttribute('aria-label')||(e?.getAttribute('role')==='button'?e.textContent?.slice(0,120):'')}`;}));
    await page.keyboard.press(key);await focusFrame(page);
  }
  const focus=await target.evaluateAll(elements=>{const describe=e=>e?{tag:e.tagName,id:e.id,tabIndex:e.tabIndex,label:e.getAttribute('aria-label')||(e.getAttribute('role')==='button'?e.textContent?.slice(0,160):null)}:null;return {active:describe(document.activeElement),targets:elements.map(describe)};});
  throw Error(`Keyboard traversal could not reach the requested control: ${JSON.stringify({...focus,key,visited:[...visited]})}`);
}
async function capture(page,name) {
  const file=`${engine}-${report.viewport.width}-${name}.png`;
  // Capture completed artwork paint, not the first semantics frame. WebSocket
  // streams remain open; request idle waits only for finite HTTP asset loads.
  await page.waitForLoadState('networkidle');await focusFrame(page);
  await page.screenshot({path:path.join(evidence,file)});
  report.screenshots.push(file);
}
async function fixture(browser,scenario,seat=1,seatCount=players,size={width:Number(process.env.CGMS_WIDTH||390),height:Number(process.env.CGMS_HEIGHT||844)}) {
  const context=await browser.newContext({viewport:size,hasTouch:true});
  const result=await context.request.post(`${origin}/__fixture/session`,{headers:{Origin:origin},data:{scenario,players:seatCount,seat,run}});
  if(!result.ok()) throw Error(`fixture ${result.status()}`);
  const {match_id}=await result.json();
  const page=await context.newPage();
  page.on('pageerror',e=>{const fault=recoveryFaults.get(page);if(fault?.consume(e.message)) report.expectedFaults.push({phase:'lost-reply-recovery',message:e.message});else report.errors.push(e.message);});
  page.on('console',m=>{if(!isBrowserError(m.type(),m.text())) return;const fault=recoveryFaults.get(page);if(fault?.consumeConsole(m.text(),m.location().url)) report.expectedFaults.push({phase:'lost-reply-recovery',message:m.text(),location:m.location()});else report.errors.push(JSON.stringify({message:m.text(),location:m.location()}));});
  await page.goto(`${origin}/#/online?match=${encodeURIComponent(match_id)}`);
  await page.getByRole('heading',{name:'CGMS · Online',exact:true}).waitFor({timeout:30000});
  await page.locator('flt-semantics-host').getByText(/Seat 1’s turn|Waiting for seat|Board closed|board_closed_settlement|settlement/).first().waitFor({timeout:30000});
  // The first authoritative response can automatically expand a required
  // decision. Finish initial asset/frame work before a helper decides whether
  // the same form needs expansion; otherwise its click can reverse that change.
  await page.waitForLoadState('networkidle');await focusFrame(page);
  return {context,page,match_id};
}
async function closePublicCards(page) {
  const close=page.getByRole('button',{name:'Close public cards',exact:true});
  if(await close.count()) {await close.click();await close.waitFor({state:'hidden'});}
}
async function closeActions(page) {return require('./reconnect_browser_acceptance.cjs').closeActions(page);}
async function cancelDraft(page) {return require('./reconnect_browser_acceptance.cjs').cancelDraft(page);}
async function selectedPhysical(page, pattern=/^(?:[AKQJ]|[2-9]|10) of /) {
  const selected=page.getByRole('button',{name:pattern}).and(page.locator('[aria-current="true"]'));
  if(await selected.count()===0) throw Error('Expected selected physical card identity was not retained');
}

async function table(page) {
  await closePublicCards(page);
  await closeActions(page);
  await page.locator('[flt-semantics-identifier="public-square"]').waitFor();
  await focusFrame(page);
}
async function hand(page) {
  await closePublicCards(page);
  await closeActions(page);
  await exposeRecord(page,page.locator('[flt-semantics-identifier="private-hand-viewport"]'));
  await focusFrame(page);
}
async function decisions(page) {return require('./reconnect_browser_acceptance.cjs').decisions(page);}
async function revealCard(page, pattern) {
  const target=page.getByRole('button',{name:pattern}).first();
  if(await target.count()) {await focusByKeyboard(page,target);await exposeRecord(page,target);return target;}
  await table(page);
  const groups=page.getByRole('button',{name:'Inspect public cards',exact:true});
  const groupCount=await groups.count();
  for(let index=0;index<groupCount;index++) {
    const group=groups.nth(index);await exposeRecord(page,group);await group.click();
    await page.getByRole('button',{name:'Close public cards',exact:true}).waitFor();
    if(await target.count()) {await exposeRecord(page,target);return target;}
    await closePublicCards(page);
  }
  throw Error(`Authorized physical card is not reachable through public seats: ${pattern}`);
}
async function confirmed(f, predicate) {
  for(let i=0;i<100;i++) {
    const r=await f.context.request.get(`${origin}/v1/matches/${encodeURIComponent(f.match_id)}`);
    if(r.ok()) {const state=await r.json();if(predicate(state)) return state;}
    await pause(500);
  }
  throw Error('Authoritative transition was not confirmed');
}
async function card(page, pattern) {
  // A temporary pending-action explanation may cover the private hand.
  // Dismiss it explicitly while preserving the required decision and draft.
  await closeActions(page);
  const target=await revealCard(page,pattern);
  const close=page.getByRole('button',{name:'Close public cards',exact:true});
  const chooser=await close.count();
  await target.tap();
  // Public selection closes the route after the card's single/double-tap
  // window. Do not reuse its disappearing semantics for the next physical card.
  if(chooser) {await close.waitFor({state:'hidden'});await focusFrame(page);}
  else await pause(350); // Existing private-card single/double-tap window.
}
async function returnToDecisions(page, keyboard=false) {
  const back=page.getByRole('button',{name:/^Close card context(?: Close card context)?$/});
  await back.waitFor();
  if(keyboard) await activate(page,back); else await back.click();
  await back.waitFor({state:'hidden'});await focusFrame(page);
}

async function activate(page,target,key='Space') {
  await exposeRecord(page,target);
  await focusByKeyboard(page,target);await focusFrame(page);
  const geometry=await scrollGeometry(target),unit=1/64;
  if(geometry.rect.width<=0||geometry.rect.height<=0||geometry.rect.left<geometry.clip.left-unit||geometry.rect.right>geometry.clip.right+unit||geometry.rect.top<geometry.clip.top-unit||geometry.rect.bottom>geometry.clip.bottom+unit) throw Error(`Keyboard activation target is outside its painted viewport: ${JSON.stringify(geometry)}`);
  if(!(await target.evaluateAll(elements=>elements.some(element=>element===document.activeElement)))) throw Error('Keyboard activation target lost focus before input');
  await page.keyboard.press(key);await focusFrame(page);
}
async function typeText(page,label,value) {
  const field=page.getByRole('textbox',{name:label,exact:true});
  // All engines must focus the painted field through real keyboard traversal;
  // DOM locator autofill can scroll a semantics node without its Flutter canvas.
  // Flutter's native editing INPUT has browser-default padding/borders beyond
  // its canvas bounds. Its immediate semantic parent is the painted field.
  const paintedField=field.locator('xpath=parent::flt-semantics');
  await exposeRecord(page,paintedField);await focusByKeyboard(page,field);await focusFrame(page);
  const geometry=await scrollGeometry(paintedField),unit=1/64;
  if(geometry.rect.width<=0||geometry.rect.height<=0||geometry.rect.left<geometry.clip.left-unit||geometry.rect.right>geometry.clip.right+unit||geometry.rect.top<geometry.clip.top-unit||geometry.rect.bottom>geometry.clip.bottom+unit) throw Error(`Keyboard text field is outside its painted viewport: ${JSON.stringify(geometry)}`);
  if(!(await field.evaluateAll(elements=>elements.some(element=>element===document.activeElement)))) throw Error('Keyboard text field lost focus before entry');
  await page.keyboard.press('ControlOrMeta+A');await page.keyboard.type(value,{delay:15});await focusFrame(page);
  if(await field.inputValue()!==value) throw Error(`Keyboard entry did not retain the ${label} draft`);
}
async function guest(browser) {
  const context=await browser.newContext({viewport:report.viewport,hasTouch:true});
  const page=await context.newPage();
  page.on('pageerror',e=>{const fault=recoveryFaults.get(page);if(fault?.consume(e.message)) report.expectedFaults.push({phase:'lost-room-start-recovery',message:e.message});else report.errors.push(e.message);});
  let initial=true;const initialErrors=[];const sessionUrl=`${origin}/v1/session`;
  page.on('console',m=>{if(!isBrowserError(m.type(),m.text())) return;const event={message:m.text(),location:m.location()};if(initial) initialErrors.push(event);else {const fault=recoveryFaults.get(page);if(fault?.consumeConsole(m.text(),m.location().url)) report.expectedFaults.push({phase:'lost-room-start-recovery',...event});else report.errors.push(JSON.stringify(event));}});
  const challengeResponse=page.waitForResponse(r=>r.url()===sessionUrl && r.request().method()==='GET');
  await page.goto(`${origin}/#/online`);
  const challenge=await challengeResponse;await challenge.finished();
  await page.getByRole('button',{name:'Continue as guest',exact:true}).waitFor();
  if(!report.screenshots.includes(`${engine}-${report.viewport.width}-entry.png`)) await capture(page,'entry');
  if(challenge.status()!==401) throw Error(`Fresh guest session challenge returned ${challenge.status()}`);
  initial=false;let consumed=false;
  for(const event of initialErrors){
    if(!consumed && isInitialSessionChallenge(event.message,event.location.url,sessionUrl,challenge.status())){consumed=true;report.expectedAuthenticationChallenges.push({phase:'initial-signed-out-session',status:challenge.status(),...event});}
    else report.errors.push(JSON.stringify(event));
  }
  await page.getByRole('button',{name:'Continue as guest',exact:true}).click();
  await page.getByRole('button',{name:'Create room',exact:true}).waitFor();return {context,page};
}
async function setting(page, name) {
  await closePublicCards(page);
  await page.getByRole('button',{name:'Display settings',exact:true}).click();
  const option=page.getByRole('menuitemcheckbox',{name,exact:true});
  await option.click();await option.waitFor({state:'hidden'});
  // Popup reverse transition lasts 300ms; its closing overlay otherwise
  // consumes the following tap on the newly exposed action expander.
  await pause(350);await focusFrame(page);
}
async function action(page,name) {return require('./reconnect_browser_acceptance.cjs').selectIntent(page,name);}
async function gameOption(page,name) {return require('./reconnect_browser_acceptance.cjs').gameOption(page,name);}
async function records(page) {
  await closeActions(page);await closePublicCards(page);
  await activate(page,page.getByRole('button',{name:/Game options/}));
  const option=page.getByRole('menuitem',{name:'Offers, promises & scores',exact:true});
  await option.waitFor();await pause(350);await activate(page,option,'Enter');
  await option.waitFor({state:'hidden'});await pause(350);await focusFrame(page);
}
async function guidance(page) {
  await decisions(page);
  const toggle=page.getByRole('button',{name:'Costs and timing',exact:true});
  const explanation=page.locator('flt-semantics-host').getByText(/The authority validates your choices/);
  if(!(await explanation.count())) await activate(page,toggle);
  await pause(200);await focusFrame(page);
}
async function purchaseGestures(f) {
  const page=f.page,posts=[];
  page.on('request',request=>{if(request.method()==='POST'&&new URL(request.url()).pathname.endsWith('/commands')) posts.push(request.url());});
  const before=await confirmed(f,()=>true);
  const source=page.getByRole('button',{name:/^10 of Spades, copy 1/}).first();
  const pile=page.getByRole('button',{name:/^Draw pile(?: Draw pile)?$/});
  await pile.tap();await focusFrame(page);await closeActions(page);
  if(posts.length) throw Error('Tapping draw pile submitted an unauthorized draw');
  await exposeRecord(page,source);let box=await source.boundingBox();
  const from={x:box.x+24,y:box.y+24};
  await page.touchscreen.tap(from.x,from.y);await pause(80);await page.touchscreen.tap(from.x,from.y);
  await page.getByRole('button',{name:/^Close card context/}).waitFor();
  await selectedPhysical(page,/^10 of Spades, copy 1/);
  await capture(page,'card-double-tap-contextual-choices');
  await closeActions(page);await cancelDraft(page);
  const drag=async(end,{cancel=false,captureName}={})=>{
    await exposeRecord(page,source);const rect=await source.boundingBox();
    await page.mouse.move(rect.x+24,rect.y+24);await page.mouse.down();
    await page.mouse.move(end.x,end.y,{steps:12});await focusFrame(page);
    if(captureName) await capture(page,captureName);
    if(cancel) await page.keyboard.press('Escape');
    await page.mouse.up();await page.mouse.move(1,1);await pause(350);await focusFrame(page);
  };
  await drag({x:20,y:20});
  if(await page.getByRole('button',{name:/^Confirm /}).count()) throw Error('Invalid drop prepared a different action');
  box=await pile.boundingBox();const destination={x:box.x+box.width/2,y:box.y+box.height/2};
  await drag(destination,{cancel:true,captureName:'drag-destinations-cancel'});
  if(await page.getByRole('button',{name:'Confirm Purchase',exact:true}).count()) throw Error('Escape cancelled drag still prepared a purchase');
  await drag(destination,{captureName:'drag-purchase-preview'});
  await page.getByRole('button',{name:'Confirm Purchase',exact:true}).waitFor();
  await capture(page,'drag-purchase-review');
  report.pointerReviewFocusBeforeEscape=await page.evaluate(()=>({tag:document.activeElement?.tagName,role:document.activeElement?.getAttribute('role'),documentFocused:document.hasFocus()}));
  await cancelDraft(page);
  await page.getByRole('button',{name:/^Close card context/}).waitFor({state:'hidden',timeout:2000});
  // The same concrete source and draw pile also prepare the same purchase by taps.
  await card(page,/^10 of Spades, copy 1/);await pile.tap();await focusFrame(page);
  await page.getByRole('button',{name:'Confirm Purchase',exact:true}).waitFor();
  await cancelDraft(page);
  if(engine==='chromium') {
    await exposeRecord(page,source);const rect=await source.boundingBox();
    const cdp=await page.context().newCDPSession(page);
    await cdp.send('Input.dispatchTouchEvent',{type:'touchStart',touchPoints:[{x:rect.x+24,y:rect.y+24}]});
    await pause(600);await capture(page,'touch-hold-active');
    for(let step=1;step<=12;step++) {
      await cdp.send('Input.dispatchTouchEvent',{type:'touchMove',touchPoints:[{x:rect.x+24+(destination.x-rect.x-24)*step/12,y:rect.y+24+(destination.y-rect.y-24)*step/12}]});
      await pause(30);
    }
    await capture(page,'touch-drag-over-pile');
    await cdp.send('Input.dispatchTouchEvent',{type:'touchEnd',touchPoints:[]});await cdp.detach();
    await page.getByRole('button',{name:'Confirm Purchase',exact:true}).waitFor();
    await capture(page,'touch-hold-drag-purchase');
  } else {await card(page,/^10 of Spades, copy 1/);await pile.tap();}
  const after=await confirmed(f,()=>true);
  if(posts.length||after.version!==before.version) throw Error('Selection, drag cancellation or popup dismissal changed authority');
  report.cases.push({case:'actual-double-tap-pointer-drag-invalid-drop-Escape-and-equivalent-tap-destination',pass:true,commandPosts:0,touchHoldDrag:engine==='chromium',version:before.version});
}
async function tradeFocusBoundary(trade) {
    await card(trade.page,/^5 of Spades, copy 1/);await action(trade.page,'Offer');
    await activate(trade.page,trade.page.getByRole('button',{name:'Seat 2',exact:true}));
    const offerDraft=trade.page.getByRole('button',{name:'Seat 2',exact:true});
    await focusByKeyboard(trade.page,offerDraft);
    const draftFocusId=await offerDraft.getAttribute('id');
    await offerDraft.evaluate(e=>{
      const chain=node=>{const result=[];for(let n=node;n&&result.length<12;n=n.parentElement) result.push({tag:n.tagName,id:n.id});return result;};
      const diagnostics={target:e,beforeChain:chain(e),mutations:[],chain};
      diagnostics.observer=new MutationObserver(records=>{for(const record of records) {
        const removed=[...record.removedNodes].filter(n=>n===e||n.contains?.(e));
        const added=[...record.addedNodes].filter(n=>n===e||n.contains?.(e));
        if(removed.length||added.length) diagnostics.mutations.push({parent:record.target.id,removed:removed.map(n=>({tag:n.tagName,id:n.id})),added:added.map(n=>({tag:n.tagName,id:n.id}))});
      }});
      diagnostics.observer.observe(document.querySelector('flt-semantics-host'),{subtree:true,childList:true});window.__cgmsFocusDiagnostics=diagnostics;
    });
    for(const [width,height] of [[599,900],[600,900],[1049,900],[1050,900],[1440,900],[390,844]]) {
      await trade.page.setViewportSize({width,height});await focusFrame(trade.page);
      const focusState=await offerDraft.evaluate(e=>{
        const d=window.__cgmsFocusDiagnostics;const r=e.getBoundingClientRect();
        return {targetId:e.id,activeId:document.activeElement?.id,activeTag:document.activeElement?.tagName,focused:document.activeElement===e,sameNode:d.target===e,documentFocused:document.hasFocus(),beforeChain:d.beforeChain,afterChain:d.chain(e),mutations:d.mutations.slice(-12),bounds:{x:r.x,y:r.y,width:r.width,height:r.height}};
      });
      if(focusState.targetId!==draftFocusId||!focusState.focused) throw Error(`Adaptive transition lost focused concrete trade recipient at ${width}x${height}: ${JSON.stringify({draftFocusId,...focusState})}`);
      await trade.page.getByText(/Give · 5 of Spades, copy 1/).waitFor();
      report.cases.push({case:'focused-physical-trade-draft-survives-boundary',width,height,pass:true});
    }
}
async function tradeChecks(browser) {
    const trade=await fixture(browser,'trade');
    await tradeFocusBoundary(trade);
    const submitOffer=trade.page.getByRole('button',{name:'Confirm Offer',exact:true});
    await focusByKeyboard(trade.page,submitOffer);
    const [tradeReply]=await Promise.all([
      trade.page.waitForResponse(response=>response.request().method()==='POST'&&new URL(response.url()).pathname.endsWith('/commands')),
      trade.page.keyboard.press('Space'),
    ]);
    if(tradeReply.status()!==200) throw Error('Explicit trade offer was not acknowledged');
    await submitOffer.waitFor({state:'hidden'});await focusFrame(trade.page);
    await records(trade.page);
    await trade.page.getByRole('group',{name:/Private offer · revision/}).first().waitFor();
    await capture(trade.page,'trade-pending');
    await focusByKeyboard(trade.page,trade.page.getByRole('button',{name:'Withdraw Offer',exact:true}));
    await capture(trade.page,'trade-pending-focused');
    const outsider=await fixture(browser,'trade',3);await records(outsider.page);await pause(150);
    if(await outsider.page.getByRole('group',{name:/Private offer/}).count()) throw Error('Private proposal leaked to an unrelated seat');
    if((await outsider.page.locator('flt-semantics-host').ariaSnapshot()).includes('5 of Spades')) throw Error('Private offered face leaked to an unrelated seat');
    await outsider.context.close();report.cases.push({case:'private-offer-semantics-boundary',pass:true});
    const acceptor=await fixture(browser,'trade',2);await records(acceptor.page);
    await acceptor.page.getByRole('group',{name:/You receive: 5 of Spades/}).waitFor();
    await capture(acceptor.page,'trade-review');
    await focusByKeyboard(acceptor.page,acceptor.page.getByRole('button',{name:'Accept Offer',exact:true}));
    await capture(acceptor.page,'trade-review-focused');
    await activate(acceptor.page,acceptor.page.getByRole('button',{name:'Accept Offer',exact:true}));
    await acceptor.page.locator('flt-semantics-host').getByText('Waiting for seat 3',{exact:true}).waitFor();
    await responses(browser,'trade',acceptorResponses);
    await acceptor.page.locator('flt-semantics-host').getByText('Seat 1’s turn',{exact:true}).waitFor();
    report.cases.push({case:'private-trade-acceptance',pass:true});
    await trade.context.close();await acceptor.context.close();
}
async function responses(browser,scenario,seats=otherSeats,last='Seat 1’s turn') {
  for(const [index,seat] of seats.entries()) {
    const other=await fixture(browser,scenario,seat);await decisions(other.page);
    await activate(other.page,other.page.getByRole('button',{name:'Pass response',exact:true}));
    await other.page.locator('flt-semantics-host').getByText(index<seats.length-1 ? `Waiting for seat ${seats[index+1]}` : last,{exact:true}).waitFor();
    await other.context.close();
  }
}
async function scrollGeometry(target) {
  return target.evaluate(e=>{
      const rect=e.getBoundingClientRect().toJSON();
      const clip={left:0,top:0,right:innerWidth,bottom:innerHeight};
      let scroll=null;
      for(let p=e.parentElement;p;p=p.parentElement) {
        const style=getComputedStyle(p),r=p.getBoundingClientRect();
        if(!r.width||!r.height) continue;
        if(/auto|scroll|hidden|clip/.test(style.overflowY)) {
          clip.top=Math.max(clip.top,r.top);clip.bottom=Math.min(clip.bottom,r.bottom);
        }
        if(/auto|scroll|hidden|clip/.test(style.overflowX)) {
          clip.left=Math.max(clip.left,r.left);clip.right=Math.min(clip.right,r.right);
        }
        // Pointer interaction can switch a Flutter scroll group's DOM overflow
        // to hidden while its canvas still accepts normal wheel scrolling.
        if(!scroll&&p.getAttribute('role')==='group'&&/auto|scroll|hidden|clip/.test(style.overflowY)&&r.bottom>0&&r.top<innerHeight) scroll=r.toJSON();
      }
      return {rect,clip,scroll};
  });
}
async function exposeRecord(page,target) {
  await target.waitFor();
  // Flutter transforms and CSS clips quantize differently by one 1/64px layout
  // unit (observed recovery-button bottom156.200 vs clip156.1875). This only
  // compares containment; target dimensions still must independently be >=48px.
  const layoutUnit=1/64;
  for(let attempt=0;attempt<12;attempt++) {
    const g=await scrollGeometry(target);
    if(g.rect.left>=g.clip.left-layoutUnit&&g.rect.right<=g.clip.right+layoutUnit&&g.rect.top>=g.clip.top-layoutUnit&&g.rect.bottom<=g.clip.bottom+layoutUnit) return;
    if(!g.scroll) break;
    const top=Math.max(0,g.scroll.top),bottom=Math.min(page.viewportSize().height,g.scroll.bottom);
    await page.mouse.move(Math.max(1,Math.min(page.viewportSize().width-1,g.scroll.left+g.scroll.width/2)),(top+bottom)/2);
    await page.mouse.wheel(0,g.rect.top+g.rect.height/2-(top+bottom)/2);await pause(150);
  }
  throw Error(`Requested control could not be fully exposed: ${JSON.stringify({text:await target.textContent(),geometry:await scrollGeometry(target),viewport:page.viewportSize()})}`);
}
async function touchTargets(page) {
  // Flutter clips semantic rectangles at scroll-pane edges. Keep every original
  // control, expose it with real wheel input, then check its interactive bounds.
  const controls=await page.getByRole('button').evaluateAll(elements=>elements.map(e=>({id:e.id,label:e.getAttribute('aria-label')||e.textContent})));
  const measured=[];
  for(const control of controls) {
    const target=page.locator(`#${control.id}`);
    const geometry=()=>scrollGeometry(target);
    let exposed=false;
    for(let attempt=0;attempt<12;attempt++) {
      let g=await geometry();
      if(g.scroll) {
        const top=Math.max(0,g.scroll.top),bottom=Math.min(page.viewportSize().height,g.scroll.bottom);
        const x=Math.max(1,Math.min(page.viewportSize().width-1,g.scroll.left+g.scroll.width/2));
        await page.mouse.move(x,(top+bottom)/2);
        await page.mouse.wheel(0,g.rect.top+g.rect.height/2-(top+bottom)/2);
        await pause(150);g=await geometry();
      }
      if(g.rect.left>=g.clip.left&&g.rect.right<=g.clip.right&&g.rect.top>=g.clip.top&&g.rect.bottom<=g.clip.bottom) {
        await target.click({trial:true});
        const final=await geometry();
        if(final.rect.left<final.clip.left||final.rect.right>final.clip.right||final.rect.top<final.clip.top||final.rect.bottom>final.clip.bottom) continue;
        measured.push({label:control.label,width:final.rect.width,height:final.rect.height});exposed=true;break;
      }
      if(!g.scroll) break;
    }
    if(!exposed) throw Error(`Primary target cannot be fully exposed: ${control.label}`);
  }
  return measured;
}
async function onlineBarGeometry(page, board, hand, {rtl=false,textScale=1}={}) {
  const marker=name=>page.locator(`[flt-semantics-identifier="${name}"]`);
  const boxes={};
  for(const name of ['game-header','header-status']) {
    await marker(name).waitFor();boxes[name]=await marker(name).boundingBox();
    if(!boxes[name]) throw Error(`Online header region is not painted: ${name}`);
  }
  const unit=1/64,right=box=>box.x+box.width,bottom=box=>box.y+box.height;
  const inside=(a,b)=>a.x>=b.x-unit&&a.y>=b.y-unit&&right(a)<=right(b)+unit&&bottom(a)<=bottom(b)+unit;
  if(!inside(boxes['game-header'],board)||bottom(boxes['game-header'])>hand.y+unit) throw Error('Compact header overlaps or exceeds reclaimed board');
  for(const name of ['footer-primary','footer-secondary']) if(await marker(name).count()) throw Error('Permanent action composer remains allocated');
  const utilityFooter=marker('game-footer');
  const utilityNames=await utilityFooter.getByRole('button').evaluateAll(elements=>elements.map(element=>element.getAttribute('aria-label')||element.textContent));
  if(utilityNames.some(name=>! /^(?:Game options(?: Game options)?|End Turn|Pass response|Pending decision|Results & settlement)$/.test(name))) throw Error(`Card action leaked into permanent utilities: ${JSON.stringify(utilityNames)}`);
  if(await utilityFooter.getByRole('textbox').count()) throw Error('Permanent utilities retained composer fields');
  for(const name of ['Prepare your action','Play cards','Choose an action']) if(await page.getByRole('button',{name,exact:true}).count()) throw Error(`Removed permanent action entry remains: ${name}`);
  const controls=[];
  for(const name of ['End Turn']) {
    const target=page.getByRole('button',{name,exact:true});
    const g=await scrollGeometry(target);
    if(g.rect.width<48-1e-5||g.rect.height<48-1e-5) throw Error('Retained turn utility target below48px');
    controls.push({name,width:g.rect.width,height:g.rect.height});
  }
  return {rtl,textScale,regions:boxes,controls,noPermanentActionPanel:true};
}
async function compactOnlineGeometry(page, {publicVisible=true,rtl=false,textScale=1}={}) {
  const marker=name=>page.locator(`[flt-semantics-identifier="${name}"]`);
  const board=marker('board-viewport'), hand=marker('private-hand-viewport');
  await board.waitFor();await hand.waitFor();await focusFrame(page);
  const boardBox=await board.boundingBox(), handBox=await hand.boundingBox();
  const unit=1/64, bottom=box=>box.y+box.height, right=box=>box.x+box.width;
  // Independently rounded Firefox child/table edges can differ by one1/60px
  // layout unit plus float roundoff. Scope this to cell containment only.
  const cellUnit=engine==='firefox'?1/60+1/16384:unit;
  report.cellContainmentToleranceCssPx=cellUnit;
  const inside=(a,b,tolerance=unit)=>a.x>=b.x-tolerance&&a.y>=b.y-tolerance&&right(a)<=right(b)+tolerance&&bottom(a)<=bottom(b)+tolerance;
  if(!boardBox||!handBox||!inside(boardBox,{x:0,y:0,...page.viewportSize()})||(handBox.x<boardBox.x-unit||right(handBox)>right(boardBox)+unit)) throw Error(`Online board or hand exceeds the actual host slot: ${JSON.stringify({boardBox,handBox,viewport:page.viewportSize()})}`);
  // Measure without wheel, focus or scrollIntoView. Clipped semantics alone do
  // not prove an actionable painted card; intersect every ancestor viewport.
  const handCards=hand.getByRole('button',{name:/^[AKQJ0-9]+ of (Hearts|Spades|Diamonds|Clubs), copy /});
  const visible=[];
  for(const card of await handCards.all()) {
    const g=await scrollGeometry(card);
    const width=Math.min(g.rect.right,g.clip.right)-Math.max(g.rect.left,g.clip.left);
    const height=Math.min(g.rect.bottom,g.clip.bottom)-Math.max(g.rect.top,g.clip.top);
    if(width>=48-unit&&height>=48-unit) visible.push({label:await card.getAttribute('aria-label')||await card.textContent(),width,height});
  }
  const needsHandScroll=!visible.length;
  if(needsHandScroll&&!(await handCards.count())) throw Error('Online hand lost its physical card nodes');
  const footer=[];
  for(const button of await page.getByRole('button',{name:'End Turn',exact:true}).all()) {
    const g=await scrollGeometry(button);
    const paintedWidth=Math.min(g.rect.right,g.clip.right)-Math.max(g.rect.left,g.clip.left);
    const paintedHeight=Math.min(g.rect.bottom,g.clip.bottom)-Math.max(g.rect.top,g.clip.top);
    if(paintedWidth>=48-unit) {
      if(paintedHeight<48-unit) throw Error(`Footer action has a painted target below48px: ${JSON.stringify({label:await button.textContent(),geometry:g,paintedWidth,paintedHeight})}`);
      footer.push({label:await button.getAttribute('aria-label')||await button.textContent(),width:paintedWidth,height:paintedHeight});
    }
  }
  if(!footer.length) throw Error('Online footer has no initially visible48px primary control');
  const bars=await onlineBarGeometry(page,boardBox,handBox,{rtl,textScale});
  const result={board:boardBox,hand:handBox,initiallyVisibleCards:visible,footer,bars};
  const seats=page.locator('[flt-semantics-identifier^="public-seat-"]');
  if(publicVisible) {
    const table=await marker('public-square').boundingBox(), cells=await seats.all();
    if(!table||(table.x<boardBox.x-unit||right(table)>right(boardBox)+unit)||Math.abs(table.x-handBox.x)>unit||Math.abs(right(table)-right(handBox))>unit||Math.abs(table.x-boardBox.x)>unit||Math.abs(right(table)-right(boardBox))>unit) throw Error(`Online public table must fill the board width and align with both hand edges: ${JSON.stringify({table,boardBox,handBox})}`);
    if(cells.length!==4) throw Error(`Expected four public-seat controls, got ${cells.length}`);
    const boxes=await Promise.all(cells.map(cell=>cell.boundingBox()));
    const pileSpace=await marker('draw-pile-space').boundingBox();
    if(!pileSpace||pileSpace.height<48||Math.abs(pileSpace.width-table.width)>unit) throw Error('Public rows lost their dedicated draw-pile spacing');
    const rowHeight=(table.height-pileSpace.height)/2;
    const clippedTable=handBox.y-bottom(table)>8+unit;
    for(const box of boxes) if(!box||Math.abs(box.width-table.width/2)>unit||box.width<48-unit||(!clippedTable&&(Math.abs(box.height-rowHeight)>unit||!inside(box,table,cellUnit)||box.height<48-unit))) throw Error(`Public quadrant width or unclipped geometry changed: ${JSON.stringify({table,boxes,pileSpace,cellUnit,clippedTable})}`);
    if(!clippedTable) {
      const top=boxes.filter(box=>Math.abs(box.y-table.y)<=unit),lower=boxes.filter(box=>Math.abs(box.y-table.y-rowHeight-pileSpace.height)<=unit);
      if(top.length!==2||lower.length!==2||Math.abs(top[0].x-top[1].x)<table.width/2-unit||Math.abs(lower[0].x-lower[1].x)<table.width/2-unit) throw Error(`Public quadrants do not form two equal rows: ${JSON.stringify(boxes)}`);
      if(handBox.y<bottom(table)-unit||handBox.y-bottom(table)>8+unit) throw Error(`Private hand is not adjacent to public table: ${JSON.stringify({table,handBox})}`);
    } else {
      // Short/large-text viewports retain card sizes and intentionally scroll.
      // A clipped semantics rectangle cannot measure the full physical row.
      // Prove every quadrant remains reachable at its complete48px inspection
      // target, including those initially below the viewport.
      const accessible=[];
      for(const cell of cells) {
        const control=cell.getByRole('button',{name:'Inspect public cards',exact:true});
        await focusByKeyboard(page,control);await focusFrame(page);
        const g=await scrollGeometry(control);
        if(g.rect.width<48-unit||g.rect.height<48-unit||!inside({x:g.rect.left,y:g.rect.top,width:g.rect.width,height:g.rect.height},{x:g.clip.left,y:g.clip.top,width:g.clip.right-g.clip.left,height:g.clip.bottom-g.clip.top})) throw Error(`Scrolled public quadrant control is not fully exposed: ${JSON.stringify(g)}`);
        accessible.push({width:g.rect.width,height:g.rect.height,fullBoundsRevealed:true});
      }
      result.scrolledQuadrants=accessible;
    }
    if(bottom(bars.regions['game-header'])>table.y+unit) throw Error('Online header overlaps the public table');
    result.table=table;result.quadrants=boxes;result.pileSpace=pileSpace;
  } else if(await seats.count()) throw Error('Covered public quadrants remain interactive beneath actions');
  if(needsHandScroll) {
    const physical=handCards.first();
    await focusByKeyboard(page,physical);await focusFrame(page);
    const revealed=await scrollGeometry(physical);
    if(revealed.rect.width<48||revealed.rect.height<48||!inside({x:revealed.rect.left,y:revealed.rect.top,width:revealed.rect.width,height:revealed.rect.height},{x:revealed.clip.left,y:revealed.clip.top,width:revealed.clip.right-revealed.clip.left,height:revealed.clip.bottom-revealed.clip.top})) throw Error(`Scrollable hand card is not fully visible after keyboard focus: ${JSON.stringify(revealed)}`);
    result.scrolledHandCard={width:revealed.rect.width,height:revealed.rect.height,fullBoundsRevealed:true};
  }
  return result;
}
async function pileSeparation(page) {
  const pile=page.getByRole('button',{name:/^Draw pile(?: Draw pile)?$/});
  await exposeRecord(page,pile);
  // Flutter clips semantics at the board scroll edge. Reveal the complete
  // painted pile with real wheel input before measuring its accessible target.
  let bounds=await pile.boundingBox();
  for(let attempt=0;bounds&&(bounds.width<48||bounds.height<48)&&attempt<12;attempt++) {
    const g=await scrollGeometry(pile);
    if(!g.scroll)break;
    const top=Math.max(0,g.scroll.top),bottom=Math.min(page.viewportSize().height,g.scroll.bottom);
    await page.mouse.move(g.scroll.left+g.scroll.width/2,(top+bottom)/2);
    await page.mouse.wheel(0,g.rect.top+g.rect.height/2-(top+bottom)/2);
    await pause(150);await focusFrame(page);bounds=await pile.boundingBox();
  }
  if(!bounds||bounds.width<48||bounds.height<48) throw Error(`Draw pile lost its 48px target: ${JSON.stringify(await scrollGeometry(pile))}`);
  const seats=await page.locator('[flt-semantics-identifier^="public-seat-"]').all();
  const boxes=await Promise.all(seats.map(seat=>seat.boundingBox()));
  const overlaps=(a,b)=>Math.min(a.x+a.width,b.x+b.width)>Math.max(a.x,b.x)+1/64&&Math.min(a.y+a.height,b.y+b.height)>Math.max(a.y,b.y)+1/64;
  if(boxes.some(box=>box&&overlaps(bounds,box))) throw Error(`Draw pile overlaps a public seat: ${JSON.stringify({pile:bounds,seats:boxes})}`);
  const hit=await pile.evaluate(element=>{
    const r=element.getBoundingClientRect(),target=document.elementFromPoint(r.x+r.width/2,r.y+r.height/2);
    return target===element||element.contains(target);
  });
  if(!hit) throw Error('Draw pile center cannot receive pointer input');
  return {pile:bounds,seats:boxes,hit};
}
async function directDropGeometryChecks(browser) {
  const f=await fixture(browser,'card-style',1,4,{width:1041,height:948}),page=f.page;
  const posts=[];page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/commands'))posts.push(r.url());});
  for(const [width,height] of [[1041,948],[319,948],[390,844],[599,900],[600,900],[599,900],[1049,900],[1050,900],[1049,900],[768,1024],[1024,768],[844,390],[1280,600],[1440,900]]) {
    await page.setViewportSize({width,height});await focusFrame(page);
    const geometry=await pileSeparation(page);
    await capture(page,`pile-separated-${width}x${height}`);
    report.cases.push({case:'draw-pile-safe-spacing-and-hit-test',width,height,pass:true,...geometry});
  }
  for(const name of ['Right to left','200% text'])await setting(page,name);
  for(const [width,height] of [[390,844],[1041,948],[1440,900]]) {
    await page.setViewportSize({width,height});await focusFrame(page);
    const geometry=await pileSeparation(page);await capture(page,`pile-separated-rtl200-${width}`);
    report.cases.push({case:'draw-pile-safe-spacing-rtl-text200',width,height,pass:true,...geometry});
  }
  if(posts.length)throw Error('Pile layout testing submitted a command');
  await f.context.close();
}
async function dragToSeat(page,source,seat,{cancel=false,captureName,checkPreview,revealDestination=false}={}) {
  if(revealDestination)await focusByKeyboard(page,source);
  await exposeRecord(page,source);
  const start=await source.boundingBox();
  const target=page.locator(`[flt-semantics-identifier="public-seat-${seat}"]`);
  let end=await target.boundingBox();
  if(!start||!end)throw Error('Physical drag source or public destination missing');
  // Empty lower part of this concrete seat, clear of series/card subtargets.
  await page.mouse.move(start.x+Math.min(24,start.width/2),start.y+Math.min(24,start.height/2));
  await page.mouse.down();
  if(revealDestination) {
    await page.mouse.move(start.x+36,start.y+24,{steps:3});await focusFrame(page);
    await exposeRecord(page,target);end=await target.boundingBox();
  }
  await page.mouse.move(end.x+end.width*0.72,end.y+end.height-12,{steps:16});
  await focusFrame(page);
  if(checkPreview)await checkPreview();
  if(captureName)await capture(page,captureName);
  if(cancel)await page.keyboard.press('Escape');
  await page.mouse.up();await page.mouse.move(1,1);await pause(350);await focusFrame(page);
}
async function largeOpeningPreviewChecks(browser) {
  const f=await fixture(browser,'opening-large',1,4,{width:390,height:844}),page=f.page,posts=[];
  page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/commands'))posts.push(r.postDataJSON());});
  const state=await confirmed(f,()=>true);
  if(state.projection.board.cards.filter(c=>c.controller===1&&c.zone==='hand').length!==18)throw Error('Large opening must contain both physical copies of all nine numbers');
  await setting(page,'200% text');
  for(const [width,height,rtl] of [[390,844,false],[844,390,false],[1041,948,true]]) {
    await page.setViewportSize({width,height});if(rtl)await setting(page,'Right to left');await focusFrame(page);
    const source=page.getByRole('button',{name:/^2 of Clubs, copy 1/}).first();
    await dragToSeat(page,source,1,{cancel:true,revealDestination:true,captureName:`large-opening-preview-${width}x${height}-text200${rtl?'-rtl':''}`,checkPreview:async()=>{
      const preview=page.getByText(/^Release to submit Open Series/).last();
      await preview.waitFor();const box=await preview.boundingBox();
      if(!box||box.x<0||box.y<0||box.x+box.width>width+1/64||box.y+box.height>height+1/64)throw Error(`Exact large opening preview leaves viewport: ${JSON.stringify(box)}`);
      const text=await preview.textContent();
      if(!text.includes('Clubs:'))throw Error('Preview omitted exact suit');
      for(let rank=2;rank<=10;rank++)if(!text.includes(`${rank}[1,2]`))throw Error(`Preview omitted physical rank/copies ${rank}`);
      if(!text.includes('Copy numbers in brackets.'))throw Error('Preview omitted physical copy legend');
    }});
    report.cases.push({case:'full-18-card-first-opening-exact-preview-text200-cancel',width,height,rtl,pass:true});
  }
  if(posts.length)throw Error('Large preview cancellation submitted a command');
  await f.context.close();
}
async function directDropInteractionChecks(browser) {
  const f=await fixture(browser,'opening',1,4,{width:1041,height:948}),page=f.page,posts=[];
  page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/commands'))posts.push(r.postDataJSON());});
  const source=page.getByRole('button',{name:/^4 of Clubs, copy 1/}).first();
  await exposeRecord(page,source);let box=await source.boundingBox();
  await page.touchscreen.tap(box.x+24,box.y+24);await pause(350);
  await selectedPhysical(page,/^4 of Clubs, copy 1/);
  await page.touchscreen.tap(box.x+24,box.y+24);await pause(350);
  if(await source.getAttribute('aria-current')==='true')throw Error('Second single tap did not toggle exact source');
  await page.touchscreen.tap(box.x+24,box.y+24);await pause(80);await page.touchscreen.tap(box.x+24,box.y+24);
  await page.getByRole('button',{name:/^Close card context/}).waitFor();await closeActions(page);await cancelDraft(page);
  await dragToSeat(page,source,1,{cancel:true,captureName:'direct-opening-cancel-preview'});
  if(posts.length)throw Error('Cancelled opening drop submitted');
  await dragToSeat(page,source,1,{captureName:'direct-opening-exact-preview'});
  await confirmed(f,s=>Boolean(s.projection.board.window_id));
  if(posts.length!==1||posts[0].type!=='open-series'||posts[0].payload.suit!=='clubs'||posts[0].payload.selection.length!==1)throw Error('Complete opening drop did not send exactly one exact command');
  if(await page.getByRole('button',{name:'Confirm Open Series',exact:true}).count())throw Error('Complete opening drop retained redundant confirmation');
  await capture(page,'direct-opening-accepted');
  await responses(browser,'opening');await confirmed(f,s=>!s.projection.board.window_id);
  report.cases.push({case:'single-tap-toggle-double-tap-Escape-and-direct-opening-drop',pass:true,commandPosts:1});
  await f.context.close();

  const confinement=await fixture(browser,'confinement',2,4,{width:1041,height:948}),acePosts=[];
  confinement.page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/commands'))acePosts.push(r.postDataJSON());});
  await closeActions(confinement.page);
  const ace=confinement.page.getByRole('button',{name:/^A of Diamonds, copy 1/}).first();
  await dragToSeat(confinement.page,ace,3);
  if(acePosts.length)throw Error('Confinement targeted an unrelated opponent');
  await dragToSeat(confinement.page,ace,1,{captureName:'direct-confinement-pending-actor-preview'});
  await confirmed(confinement,s=>s.projection.board.players.find(p=>p.seat===1).confined===true);
  if(acePosts.length!==1||acePosts[0].type!=='confinement')throw Error('Complete Confinement drop did not submit once');
  await capture(confinement.page,'direct-confinement-resolved');
  report.cases.push({case:'Ace-drop-rejects-wrong-actor-and-confines-exact-pending-actor',pass:true,commandPosts:1});
  await confinement.context.close();

  const loan=await fixture(browser,'loan',1,4,{width:1041,height:948}),loanPosts=[];
  loan.page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/commands'))loanPosts.push(r.postDataJSON());});
  const before=await confirmed(loan,()=>true),formation=before.projection.board.formations.find(f=>f.spec.kind==='underground');
  const formationSource=loan.page.locator(`[flt-semantics-identifier="public-formation-${formation.id}"]`);
  await dragToSeat(loan.page,formationSource,2,{captureName:'direct-loan-exact-members-preview'});
  const proposed=await confirmed(loan,s=>(s.projection.board.proposals||[]).some(p=>p.status==='offered'));
  if(loanPosts.length!==1||loanPosts[0].type!=='offer'||loanPosts[0].payload.terms.loan!==true||loanPosts[0].payload.terms.to!==2)throw Error('Formation drop became something other than one exact loan proposal');
  if(proposed.projection.board.window_id||proposed.projection.board.formations.find(f=>f.id===formation.id).controller!==1)throw Error('Proposal transferred formation or began responses early');
  await records(loan.page);await capture(loan.page,'direct-loan-private-proposal');
  const borrower=await fixture(browser,'loan',2,4,{width:1041,height:948});await records(borrower.page);
  await activate(borrower.page,borrower.page.getByRole('button',{name:'Accept Offer',exact:true}));
  const accepted=await confirmed(borrower,s=>Boolean(s.projection.board.window_id));
  if(accepted.projection.board.formations.find(f=>f.id===formation.id).controller!==1)throw Error('Acceptance transferred loan before responses');
  await capture(borrower.page,'direct-loan-accepted-pending-responses');
  await responses(browser,'loan',acceptorResponses);
  const resolved=await confirmed(borrower,s=>!s.projection.board.window_id&&s.projection.board.formations.some(f=>f.id===formation.id&&f.controller===2));
  if(resolved.projection.online.rules_context.underground!==true||resolved.projection.online.alliances['1']!==2||resolved.projection.online.alliances['2']!==1)throw Error('Successful Underground loan did not establish the exclusive pair and persistent privilege');
  await closeActions(borrower.page);await capture(borrower.page,'direct-loan-alliance-resolved');
  report.cases.push({case:'intact-formation-drop-private-loan-exact-acceptance-responses-atomic-alliance',pass:true,commandPosts:1});
  await loan.context.close();await borrower.context.close();
}
async function observeDrawAnnouncements(page) {
  // Observe rendered accessibility feedback only; never invoke application code.
  await page.evaluate(()=>{
    window.drawObservations=[];let active='';
    const sample=()=>{
      const labels=[...document.querySelectorAll('flt-semantics[aria-label]')]
        .map(e=>e.getAttribute('aria-label'))
        .filter(s=>/^Turn draw:|^Seat \d+ drew one card\.$/.test(s));
      const current=labels.join('|');
      if(current&&current!==active)window.drawObservations.push(current);
      active=current;
    };
    new MutationObserver(sample).observe(document.documentElement,{subtree:true,childList:true,attributes:true,attributeFilter:['aria-label']});sample();
  });
}
async function scheduledDrawChecks(browser) {
  for(const [scenario,zone,quiet] of [['turn-draw-ace','concealed-ace',false],['turn-draw-diamond','hand',true]]) {
    const owner=await fixture(browser,scenario,1,4,{width:1041,height:948});
    const recipient=await fixture(browser,scenario,2,4,{width:1041,height:948});
    const posts=[];
    for(const f of [owner,recipient]) {
      if(quiet) {await f.page.emulateMedia({reducedMotion:'reduce'});await focusFrame(f.page);}
      await observeDrawAnnouncements(f.page);
      f.page.on('request',r=>{if(r.method()==='POST'&&new URL(r.url()).pathname.endsWith('/commands'))posts.push(r.postDataJSON().type);});
    }
    const before=await confirmed(recipient,()=>true);
    await activate(owner.page,owner.page.getByRole('button',{name:'End Turn',exact:true}));
    const label=zone==='hand'?'Turn draw: one card added to your hand.':'Turn draw: one card added to your concealed Ace area.';
    await recipient.page.locator(`flt-semantics[aria-label="${label}"]`).waitFor({timeout:10000});
    // Do not wait for network-idle and miss this deliberately brief feedback.
    const screenshot=`${engine}-${scenario}-${quiet?'quiet':'glow'}.png`;
    await recipient.page.screenshot({path:path.join(evidence,screenshot)});report.screenshots.push(screenshot);
    const after=await confirmed(recipient,s=>s.projection.online.turn_started===true&&s.projection.board.active===2);
    const acquired=after.projection.board.cards.filter(c=>c.controller===2&&c.zone===zone);
    if(acquired.length!==1||after.version<=before.version)throw Error('Scheduled draw did not reach its exact authorized destination');
    const observer=await confirmed(owner,()=>true);
    if((observer.projection.board.cards||[]).some(c=>c.controller===2&&['hand','concealed-ace'].includes(c.zone)))throw Error('Opponent draw revealed a hidden face');
    await pause(1200);
    for(const f of [owner,recipient]) {
      await f.page.setViewportSize({width:599,height:900});await f.page.setViewportSize({width:600,height:900});
      await f.page.getByRole('button',{name:'Refresh connection',exact:true}).click();await pause(1300);
      const observations=await f.page.evaluate(()=>window.drawObservations);
      if(observations.length!==1||observations[0]!== (f===recipient?label:'Seat 2 drew one card.'))throw Error(`Draw feedback repeated or leaked identity: ${JSON.stringify(observations)}`);
      await f.page.reload();await f.page.getByRole('heading',{name:'CGMS · Online',exact:true}).waitFor();await observeDrawAnnouncements(f.page);await pause(1300);
      if((await f.page.evaluate(()=>window.drawObservations)).length)throw Error('Reconnect replayed an old draw');
    }
    if(posts.length!==1||posts[0]!=='end-turn')throw Error('Feedback issued an extra gameplay command');
    report.cases.push({case:'scheduled-authoritative-draw-once-private-destination-refresh-resize-reconnect',scenario,zone,reducedMotion:quiet,commandPosts:1,pass:true});
    await owner.context.close();await recipient.context.close();
  }
}
async function publicGrowthOnlineChecks(browser) {
  // This legal fixture has public artwork and the actor's own Queen/Ace, plus
  // other seats' hidden hands/Aces. No command is needed to review its layout.
  const f=await fixture(browser,'card-style',1,4,{width:319,height:948}),page=f.page,posts=[];
  page.on('request',request=>{if(request.method()==='POST') posts.push(new URL(request.url()).pathname);});
  await closeActions(page);
  const url=page.url(),timeOrigin=await page.evaluate(()=>performance.timeOrigin);
  let navigations=0;page.on('framenavigated',frame=>{if(frame===page.mainFrame()) navigations++;});
  const snapshot=async()=>{
    const response=await f.context.request.get(`${origin}/v1/matches/${encodeURIComponent(f.match_id)}`);
    if(response.status()!==200) throw Error(`Public-growth snapshot failed with HTTP${response.status()}`);
    return response.json();
  };
  const before=await snapshot(),measurements=[];report.publicGrowthMeasurements=measurements;
  const handIdentities=()=>page.locator('[flt-semantics-identifier="private-hand-viewport"]').getByRole('button',{name:/^[AKQJ0-9]+ of (Hearts|Spades|Diamonds|Clubs), copy /}).allTextContents();
  const originalHand=await handIdentities();
  const expectedHand=['A of Hearts, copy 1 concealed-ace','Q of Hearts, copy 1 hand'];
  if(JSON.stringify(originalHand)!==JSON.stringify(expectedHand)) throw Error('Public-growth fixture must expose exactly the acting seat own Queen and Ace');
  const expandingWidths=[319,390,500,559,560,600,821,1050,1440];
  const widths=[...expandingWidths,...expandingWidths.slice(0,-1).reverse()];
  const footerAction=page.getByRole('button',{name:/^Q of Hearts, copy 1/}).first();
  let footerFocusId;
  for(const width of widths) {
    await page.setViewportSize({width,height:948});await focusFrame(page);
    const geometry=await compactOnlineGeometry(page);
    measurements.push({width,height:948,...geometry});
    if(footerFocusId===undefined) {
      await focusByKeyboard(page,footerAction);await focusFrame(page);
      footerFocusId=await footerAction.getAttribute('id');
      if(!footerFocusId) throw Error('Focused online footer action has no stable semantics identity');
    }
    if(await footerAction.getAttribute('id')!==footerFocusId||!(await footerAction.evaluate(element=>element===document.activeElement))) throw Error(`Online footer keyboard focus was lost during live resize at${width}`);
    if(JSON.stringify(await handIdentities())!==JSON.stringify(originalHand)) throw Error('Live public growth changed authorized physical hand identities');
    const semantics=await page.locator('flt-semantics-host').ariaSnapshot();
    if(/(?:Q|A) of (?:Clubs|Spades|Diamonds), copy/.test(semantics)) throw Error('Live public growth exposed a concealed opponent hand or Ace identity');
    if(page.url()!==url||await page.evaluate(()=>performance.timeOrigin)!==timeOrigin) throw Error('Public growth reloaded or navigated the document');
    await capture(page,`public-growth-${measurements.length}-${width}x948`);
  }
  const peak=Math.floor(measurements.length/2);
  for(let index=1;index<=peak;index++) {
    const before=measurements[index-1],after=measurements[index];
    const widthDelta=after.width-before.width;
    if(Math.abs(after.table.width-before.table.width-widthDelta)>1/64||Math.abs(after.quadrants[0].width-before.quadrants[0].width-widthDelta/2)>1/64) throw Error(`Online public table and equal quadrants must follow each exact width change at fixed height: ${JSON.stringify({before,after,widthDelta})}`);
  }
  for(let index=peak+1;index<measurements.length;index++) {
    const original=measurements[measurements.length-1-index],resized=measurements[index];
    for(const surface of ['table','hand']) for(const dimension of ['x','y','width','height']) {
      if(Math.abs(resized[surface][dimension]-original[surface][dimension])>=1) throw Error(`Shrinking did not restore the same online ${surface} ${dimension}`);
    }
    for(const name of ['game-header','header-status']) for(const dimension of ['x','y','width','height']) {
      if(Math.abs(resized.bars.regions[name][dimension]-original.bars.regions[name][dimension])>=1) throw Error(`Shrinking did not restore online ${name} ${dimension}`);
    }
  }
  // Activate the same focused physical card after resizing. Enter opens only
  // its contextual interpretations and never issues an authority command.
  await page.keyboard.press('Enter');await focusFrame(page);
  await page.getByRole('button',{name:/^Close card context/}).waitFor();
  await closeActions(page);
  const toggle=footerAction;
  await focusByKeyboard(page,toggle);await focusFrame(page);
  const headerFocusId=await toggle.getAttribute('id'),boundaryMeasurements=[];
  for(const width of [599,600,1049,1050,1049,600,599]) {
    await page.setViewportSize({width,height:948});await focusFrame(page);
    const geometry=await compactOnlineGeometry(page);
    if(await toggle.getAttribute('id')!==headerFocusId||!(await toggle.evaluate(element=>element===document.activeElement))) throw Error(`Physical card focus was lost at live${width}px boundary`);
    boundaryMeasurements.push({width,bars:geometry.bars});
  }
  await page.keyboard.press('Enter');await focusFrame(page);
  if(await page.getByRole('button',{name:/^Close card context/}).count()!==1) throw Error('Resized physical card must open exactly one contextual popup');
  const after=await snapshot();
  if(navigations||posts.length||before.version!==after.version||before.projection.board.game_id!==after.projection.board.game_id) throw Error('Live public growth navigated, issued a command or changed authority state');
  report.cases.push({case:'online-live-full-width-public-table-and-preview-composition-grow-at-fixed-height',measurements,navigations,commandPosts:posts.length,beforeVersion:before.version,afterVersion:after.version,pass:true,
    artworkMeasurement:'Browser DOM measures table and seat composition; transformed widget paint bounds verify decorative card scaling, with these captures for visual browser review.'});
  report.cases.push({case:'online-live-boundaries-preserve-physical-card-focus-and-action-authority',footerFocusId,headerFocusId,boundaryMeasurements,footerActivation:'Card Enter opens its contextual moves',headerActivation:'Card Enter opens one contextual popup',commandPosts:posts.length,navigations,pass:true});
  console.log('Online public growth and physical card focus passed live319→390→500→559→560→600→821→1050→1440 and reverse at948, plus599↔600 and1049↔1050 focus transitions');
  await f.context.close();
}
async function contextLayoutGeometry(page) {
  const result={};
  for(const [key,identifier] of Object.entries({hand:'private-hand-viewport',table:'public-square'})) result[key]=await page.locator(`[flt-semantics-identifier="${identifier}"]`).boundingBox();
  return result;
}
async function quadrantOnlineChecks(browser) {
  report.fixture='Real online host with authoritative synthetic purchase and card-style positions, real toolbar and notices; no gameplay commands';
  report.input='viewport geometry, real keyboard scrolling and contextual card drafts';
  report.accessibilityBoundary='Unclipped quadrant geometry, plus full48px quadrant controls and physical cards reached by keyboard scrolling on short screens; app RTL/text scaling and retained draft/focus checks, no native device or screen-reader claim';
  const modes=[
    {name:'normal',settings:[],sizes:[[319,948],[320,568],[390,844],[768,1024],[1041,948],[1440,900],[844,390]]},
    {name:'rtl-text200',settings:['Right to left','200% text'],sizes:[[319,948],[320,568],[390,844],[768,1024],[1440,900],[1440,1200],[844,390]]},
  ];
  for(const mode of modes) for(const [width,height] of mode.sizes) {
    const f=await fixture(browser,'purchase',1,4,{width,height}),page=f.page,posts=[];
    page.on('request',request=>{if(request.method()==='POST') posts.push(new URL(request.url()).pathname);});
    for(const name of mode.settings) await setting(page,name);
    await closeActions(page);
    const rtl=mode.settings.includes('Right to left');
    const textScale=mode.settings.includes('200% text')?2:1;
    const initial=await compactOnlineGeometry(page,{rtl,textScale});
    const beforeResponse=await f.context.request.get(`${origin}/v1/matches/${encodeURIComponent(f.match_id)}`);
    if(beforeResponse.status()!==200) throw Error(`Geometry before snapshot failed with HTTP${beforeResponse.status()}`);
    const before=await beforeResponse.json();
    await capture(page,`quadrants-online-${width}x${height}-${mode.name}`);
    await card(page,/^10 of Spades, copy 1/);
    const selectedGeometry=await contextLayoutGeometry(page);
    await action(page,'Purchase');
    const quantity=page.getByRole('textbox',{name:'Quantity',exact:true});
    await focusByKeyboard(page,quantity);await page.keyboard.press('ControlOrMeta+A');await page.keyboard.type('3');
    if(await quantity.inputValue()!=='3') throw Error('Purchase quantity draft did not accept keyboard input');
    const expanded=await contextLayoutGeometry(page);
    if(JSON.stringify(selectedGeometry.hand)!==JSON.stringify(expanded.hand)||JSON.stringify(selectedGeometry.table)!==JSON.stringify(expanded.table)) throw Error('Contextual popup reserved permanent board space');
    await capture(page,`quadrants-actions-${width}x${height}-${mode.name}`);
    await closeActions(page);
    const closed=await contextLayoutGeometry(page);
    if(JSON.stringify(selectedGeometry.hand)!==JSON.stringify(closed.hand)) throw Error('Closing actions did not restore the original hand geometry');
    await decisions(page);await focusByKeyboard(page,quantity);
    if(await quantity.inputValue()!=='3') throw Error('Closing and reopening actions lost the quantity draft');
    const afterResponse=await f.context.request.get(`${origin}/v1/matches/${encodeURIComponent(f.match_id)}`);
    if(afterResponse.status()!==200) throw Error(`Geometry after snapshot failed with HTTP${afterResponse.status()}`);
    const after=await afterResponse.json();
    if(after.version!==before.version||after.projection.board.game_id!==before.projection.board.game_id||posts.length) throw Error(`Presentation interactions changed authority or submitted a command: ${JSON.stringify({beforeVersion:before.version,afterVersion:after.version,posts})}`);
    report.cases.push({case:'online-card-board-responsive-geometry-and-contextual-draft',mode:mode.name,width,height,pass:true,initial,expanded,rawDraftRetained:true,pagePosts:0});
    console.log(`Online quadrants passed ${width}x${height} ${mode.name}`);
    for(const [file,expected] of Object.entries(report.webAssetsSha256)) {
      const response=await f.context.request.get(`${origin}/${file}`);
      if(!response.ok()||require('node:crypto').createHash('sha256').update(await response.body()).digest('hex')!==expected) throw Error(`Online quadrant browser served stale ${file}`);
    }
    await f.context.close();
  }
  await publicGrowthOnlineChecks(browser);
}
async function cardSizeChecks(browser) {
  report.fixture='Synthetic opening hand and public attack table; selection and magnification only, without authority commands';
  report.accessibilityBoundary='Real touch selection/double tap on exposed hand strips or full public cards, corner magnifier, keyboard I, browser semantics and app text/RTL settings; no native-device or screen-reader claim';
  report.artworkVerification='Native image dimensions and no-crop composition are covered by package widget tests; these browser screenshots require visual review because CanvasKit excludes the decorative painting/header from DOM semantics';
  const modes=[
    {name:'normal',settings:[],sizes:[[320,800],[390,844],[768,1024],[1440,900]]},
    {name:'text200',settings:['200% text'],sizes:[[390,844]]},
    {name:'rtl',settings:['Right to left'],sizes:[[390,844]]},
  ];
  // Browser transforms can report 48 CSS px as 47.999996185302734.
  // Compare at 0.001px precision, far below a rendered pixel, while retaining
  // raw dimensions in the report; the target minimum remains 48 CSS px.
  const cssPixel=value=>Math.round(value*1000)/1000;
  const fullyExposed=(g,allowVerticalEdge)=>g.rect.width>0&&g.rect.height>0&&cssPixel(g.rect.left)>=cssPixel(g.clip.left)&&cssPixel(g.rect.right)<=cssPixel(g.clip.right)&&(allowVerticalEdge?cssPixel(g.rect.top)>=cssPixel(g.clip.top)&&cssPixel(g.rect.bottom)<=cssPixel(g.clip.bottom):cssPixel(g.rect.top)>cssPixel(g.clip.top)&&cssPixel(g.rect.bottom)<cssPixel(g.clip.bottom));
  report.geometryComparisonPrecisionCssPx=.001;
  // CSS box sizes quantize to 1/64px while Flutter transforms retain the
  // fractional painting height. Allow one layout unit for nested containment;
  // this does not relax the independently measured 48px target minimum.
  const containmentUnit=1/64;
  report.nestedContainmentToleranceCssPx=containmentUnit;
  // Center with real wheel input before measuring: a Flutter semantic
  // rectangle at a scroll edge may describe only the clipped part of a card.
  const measure=async(page,target,label,allowVerticalEdge=false,handEdge=null)=>{
    await target.waitFor();
    if(handEdge) {await focusByKeyboard(page,target);await focusFrame(page);}
    await pause(350);
    let lastGeometry;
    for(let attempt=0;attempt<12;attempt++) {
      let g=await scrollGeometry(target);
      if(g.scroll) {
        const top=Math.max(0,g.scroll.top),bottom=Math.min(page.viewportSize().height,g.scroll.bottom);
        const x=Math.max(1,Math.min(page.viewportSize().width-1,g.scroll.left+g.scroll.width/2));
        await page.mouse.move(x,(top+bottom)/2);
        await page.mouse.wheel(0,g.rect.top+g.rect.height/2-(top+bottom)/2);
        await pause(150);g=await scrollGeometry(target);
      }
      lastGeometry=g;
      if(fullyExposed(g,allowVerticalEdge)) {
        await target.click({trial:true,...(handEdge?{position:{x:handEdge==='right'?g.rect.width-24:24,y:32}}:{})});
        const final=await scrollGeometry(target);
        lastGeometry=final;
        if(fullyExposed(final,allowVerticalEdge)) return final.rect;
      }
      if(!g.scroll) break;
    }
    throw Error(`Card-size measurement could not fully expose ${label}: ${JSON.stringify({viewport:page.viewportSize(),geometry:lastGeometry})}`);
  };
  for(const scenario of [
    {fixture:'opening',surface:'Hand',identity:'4 of Clubs, copy 1',art:'clubs-4.png',action:'Open Series'},
    {fixture:'attack',surface:'Table',identity:'8 of Clubs, copy 1',art:'clubs-8.png',action:'Attack'},
  ]) {
    const artBytes=fs.readFileSync(path.resolve(__dirname,'../packages/cgms_ui/assets/art',scenario.art));
    const artwork={asset:scenario.art,width:artBytes.readUInt32BE(16),height:artBytes.readUInt32BE(20)};
    artwork.aspectRatio=artwork.width/artwork.height;
    for(const mode of modes) for(const [width,height] of mode.sizes) {
      const f=await fixture(browser,scenario.fixture,1,players,{width,height}),page=f.page,commands=[];
      page.on('request',request=>{if(request.method()==='POST'&&new URL(request.url()).pathname.endsWith('/commands')) commands.push(request.url());});
      const showSurface=async()=>{
        if(scenario.surface==='Table') {await table(page);await page.getByRole('button',{name:'Inspect public cards',exact:true}).first().click();await page.getByRole('button',{name:'Close public cards',exact:true}).waitFor();await revealCard(page,new RegExp(`^${scenario.identity}`));}
        else {await hand(page);await focusFrame(page);}
      };
      for(const name of mode.settings) await setting(page,name);
      await showSurface();
      const selected=page.getByRole('button',{name:new RegExp(`^${scenario.identity}`)}).first();
      const inspect=page.getByRole('button',{name:new RegExp(`^Inspect ${scenario.identity}`)}).first();
      const handEdge=scenario.surface==='Hand'?(mode.name==='rtl'?'right':'left'):null;
      const measureSelected=()=>measure(page,selected,scenario.identity,false,handEdge);
      const touchPoint=face=>({x:face.x+(handEdge==='left'?24:handEdge==='right'?face.width-24:face.width/2),y:face.y+(handEdge?Math.min(32,face.height/2):face.height/2)});
      let face=await measureSelected();
      // The face includes the rank/suit header. Never compare this whole DOM
      // rectangle to the native painting ratio or claim it measures image fit.
      const initialPoint=touchPoint(face);
      await page.touchscreen.tap(initialPoint.x,initialPoint.y);
      if(scenario.surface==='Table') await page.getByRole('button',{name:'Close public cards',exact:true}).waitFor({state:'hidden'});
      else await pause(350);
      await focusFrame(page);await decisions(page);
      await selectedPhysical(page);
      await showSurface();face=await measureSelected();
      if(cssPixel(face.width)<48||cssPixel(face.height)<48) throw Error(`Card selection target below 48px: ${JSON.stringify(face)}`);
      // The selected reference uses compact private cards with double-tap/I
      // inspection. Full public chooser cards retain the separate magnifier.
      let corner=null;
      if(scenario.surface==='Table') {
        const inspection=await measure(page,inspect,`Magnify ${scenario.identity}`,true);
        if(cssPixel(inspection.width)<48||cssPixel(inspection.height)<48) throw Error(`Magnifier target below 48px: ${JSON.stringify(inspection)}`);
        face=await measureSelected();
        corner=(await scrollGeometry(inspect)).rect;
        if(corner.left<face.left-containmentUnit||corner.right>face.right+containmentUnit||corner.top<face.top-containmentUnit||corner.bottom>face.bottom+containmentUnit) throw Error(`Magnifier is outside the card: ${JSON.stringify({face,corner})}`);
      }
      const center=touchPoint(face);
      if(corner&&center.x>=corner.left&&center.x<=corner.right&&center.y>=corner.top&&center.y<=corner.bottom) throw Error('Magnifier intercepts the center selection target');
      const reflow=await page.evaluate(()=>({width:innerWidth,documentWidth:document.documentElement.scrollWidth}));
      if(reflow.documentWidth>reflow.width) throw Error(`Card reflow overflow: ${JSON.stringify(reflow)}`);
      await page.mouse.move(1,1);await pause(350);
      await capture(page,`native-cards-${scenario.surface.toLowerCase()}-${width}x${height}-${mode.name}`);
      await page.touchscreen.tap(center.x,center.y);await pause(80);await page.touchscreen.tap(center.x,center.y);
      await page.getByRole('button',{name:/^Close card context/}).waitFor();
      await closeActions(page);await showSurface();await focusByKeyboard(page,selected);await page.keyboard.press('i');
      await page.getByRole('button',{name:/^Close card context/}).waitFor();await pause(350);
      if(width===390&&mode.name==='normal') await capture(page,`native-inspector-${scenario.surface.toLowerCase()}-${width}x${height}`);
      await returnToDecisions(page);await selectedPhysical(page);
      if(corner) {
        await showSurface();await measure(page,inspect,`Magnify ${scenario.identity}`,true);
        await inspect.tap();await returnToDecisions(page);
        await selectedPhysical(page);
      }
      await showSurface();await focusByKeyboard(page,selected);await page.keyboard.press('i');
      await returnToDecisions(page,true);await selectedPhysical(page);
      if(commands.length) throw Error('Card selection or magnification submitted an authority command');
      report.cases.push({case:'reference-card-selection-double-tap-and-keyboard-inspection',surface:scenario.surface,mode:mode.name,width,height,pass:true,face:{width:face.width,height:face.height},artwork,corner:corner?{width:corner.width,height:corner.height}:null,selectionRetained:true,keyboardI:true,measurement:'fully exposed card semantics including header; native artwork fit requires widget tests plus visual review',commandPosts:0});
      console.log(`Native cards passed ${scenario.surface} ${width}x${height} ${mode.name}`);
      await f.context.close();
    }
  }
}

async function demoCardChecks(browser) {
  report.fixture='Default approved local demo at 390x844; visual and magnification checks only';
  report.accessibilityBoundary='Touch input and screenshot review; no screen-reader or native-device claim';
  const context=await browser.newContext({viewport:{width:390,height:844},hasTouch:true});
  const page=await context.newPage(),posts=[];
  page.on('pageerror',error=>report.errors.push(error.message));
  page.on('console',message=>{if(isBrowserError(message.type(),message.text())) report.errors.push(message.text());});
  page.on('request',request=>{if(request.method()==='POST')posts.push(new URL(request.url()).pathname);});
  await page.goto(`${origin}/#/play`);
  await page.locator('flt-semantics-host').getByRole('button').first().waitFor({timeout:30000});
  await capture(page,'native-demo-table-390x844');
  await hand(page);await focusFrame(page);
  const queen=page.getByRole('button',{name:/^Q of Clubs, copy 1/}).first();
  for(let attempt=0;attempt<40&&!(await queen.count());attempt++) {
    await page.mouse.move(195,500);await page.mouse.wheel(0,500);await pause(150);
  }
  await exposeRecord(page,queen);await page.mouse.move(1,1);await pause(350);
  await capture(page,'native-demo-hand-390x844');
  const rect=await queen.boundingBox(),center={x:rect.x+rect.width/2,y:rect.y+rect.height/2};
  await page.touchscreen.tap(center.x,center.y);await pause(80);await page.touchscreen.tap(center.x,center.y);
  const close=page.getByRole('button',{name:'Close inspection',exact:true});
  await close.waitFor();await pause(350);await capture(page,'native-demo-inspector-390x844');
  await close.click();await close.waitFor({state:'hidden'});
  if(posts.length) throw Error(`Default demo magnification made a POST: ${posts.join(',')}`);
  report.cases.push({case:'default-demo-phone-table-hand-and-double-tap-inspector',pass:true,width:390,height:844,commandPosts:0});
  await context.close();
}
async function keyboardSequenceChecks(browser) {
  report.fixture='Long-lived opening/attack pages through complete normal, text200 and RTL viewport sequences; no authority commands';
  const normal=[[320,800],[360,800],[390,844],[599,900],[600,900],[768,1024],[1049,900],[1050,900],[1440,900],[844,390]];
  for(const scenario of [
    {fixture:'opening',surface:'Hand',identity:'4 of Clubs, copy 1',action:'Open Series'},
    {fixture:'attack',surface:'Table',identity:'8 of Clubs, copy 1',action:'Attack'},
  ]) {
    const f=await fixture(browser,scenario.fixture),page=f.page,commands=[];
    page.on('request',request=>{if(request.method()==='POST'&&new URL(request.url()).pathname.endsWith('/commands')) commands.push(request.url());});
    const show=async()=>{if(scenario.surface==='Table') {await table(page);await revealCard(page,new RegExp(`^${scenario.identity}`));}else {await hand(page);await focusFrame(page);}};
    await show();
    const selected=page.getByRole('button',{name:new RegExp(`^${scenario.identity}`)}).first();
    await selected.click();
    if(scenario.surface==='Table') await page.getByRole('button',{name:'Close public cards',exact:true}).waitFor({state:'hidden'});
    // A directly visible public card has no chooser route to wait for. Wait
    // for this exact single-tap selection before resize/inspection can cancel
    // its pending double-tap interval. The retained-identity checks stay exact.
    await selected.and(page.locator('[aria-current="true"]')).waitFor({state:'visible',timeout:2000});
    for(const mode of [
      {name:'normal',sizes:normal},
      {name:'text200',sizes:[[320,800],[390,844],[768,1024],[1440,1200]]},
      {name:'rtl',sizes:normal},
    ]) {
      if(mode.name==='text200') await setting(page,'200% text');
      if(mode.name==='rtl') {await setting(page,'200% text');await setting(page,'Right to left');}
      for(const [width,height] of mode.sizes) {
        await page.setViewportSize({width,height});await focusFrame(page);await show();
        await focusByKeyboard(page,selected);await page.keyboard.press('i');
        const back=page.getByRole('button',{name:/^Close card context/});await back.waitFor();
        await returnToDecisions(page,true);await selectedPhysical(page);
        if(commands.length) throw Error('Keyboard inspection or resize submitted an authority command');
        report.cases.push({case:'long-sequence-forward-keyboard-inspection',surface:scenario.surface,mode:mode.name,width,height,pass:true,commandPosts:0});
      }
      await capture(page,`keyboard-sequence-${scenario.surface}-${mode.name}`);
    }
    await f.context.close();
  }
}
async function rtlHandChecks(browser) {
  report.fixture='Four authorized private Queens, two physical copies per suit, from the real Coup QA position; no authority commands';
  report.accessibilityBoundary='RTL overlapping-hand keyboard and emulated-touch input; this four-card fixture does not establish dense horizontal overflow';
  const f=await fixture(browser,'coup'),page=f.page,commands=[];
  page.on('request',request=>{if(request.method()==='POST'&&new URL(request.url()).pathname.endsWith('/commands')) commands.push(request.url());});
  // Keep a neutral selection so all four physical Queens can be selected and
  // inspected independently without creating or confirming a formation.
  await setting(page,'Right to left');
  const identities=['Q of Hearts, copy 1','Q of Hearts, copy 2','Q of Diamonds, copy 1','Q of Diamonds, copy 2'];
  const targetFor=identity=>page.getByRole('button',{name:new RegExp(`^${identity}(?: |$)`)});
  const selection=target=>target.evaluate(e=>e.getAttribute('aria-current')??e.getAttribute('aria-selected')??e.getAttribute('aria-pressed'));
  const reveal=async target=>{
    await focusByKeyboard(page,target);await focusFrame(page);
    const g=await scrollGeometry(target),unit=1/64;
    if(g.rect.left<g.clip.left-unit||g.rect.right>g.clip.right+unit||g.rect.top<g.clip.top-unit||g.rect.bottom>g.clip.bottom+unit) throw Error(`RTL hand card is outside its nested painted viewport: ${JSON.stringify(g)}`);
    if(g.rect.width<48||g.rect.height<48) throw Error(`RTL hand card target is below48px: ${JSON.stringify(g.rect)}`);
    return g.rect;
  };
  for(const mode of [
    {name:'rtl',sizes:[[390,844],[599,900],[600,900],[768,1024],[1049,900],[1050,900],[1440,900],[390,844]]},
    {name:'rtl-text200',sizes:[[390,844],[768,1024],[1440,1200],[390,844]]},
  ]) {
    if(mode.name==='rtl-text200') await setting(page,'200% text');
    for(const [width,height] of mode.sizes) {
      await page.setViewportSize({width,height});await focusFrame(page);await hand(page);
      const cards=page.getByRole('button',{name:/^Q of (Hearts|Diamonds), copy [12](?: |$)/});
      if(await cards.count()!==4) throw Error('RTL fixture must expose all four private physical Queens');
      const measured=[];
      for(const identity of identities) {
        const target=targetFor(identity),rect=await reveal(target);
        if(await target.count()!==1) throw Error(`RTL physical identity is not unique: ${identity}`);
        measured.push({identity,width:rect.width,height:rect.height});
        const first=measured[0];
        if(Math.abs(rect.width-first.width)>1/64||Math.abs(rect.height-first.height)>1/64||Math.abs(rect.width/rect.height-.8)>.001) throw Error(`RTL hand footprints differ: ${JSON.stringify(measured)}`);
        if(await selection(target)!=='false') throw Error(`RTL physical copy unexpectedly selected: ${identity}`);
        const point={x:rect.right-24,y:rect.top+Math.min(32,rect.height/2)};
        await page.touchscreen.tap(point.x,point.y);await pause(350);await focusFrame(page);
        if(await selection(target)!=='true') throw Error(`RTL exposed strip did not select ${identity}`);
        for(const other of identities.filter(value=>value!==identity)) if(await selection(targetFor(other))!=='false') throw Error(`RTL exposed strip selected adjacent copy ${other}`);
        await reveal(target);await page.keyboard.press('i');
        await page.getByRole('button',{name:/^Close card context(?: Close card context)?$/}).waitFor();
        // The full inspection face keeps copy identity in its own semantics.
        const inspectionSnapshot=await page.locator('flt-semantics-host').ariaSnapshot();
        if(!inspectionSnapshot.split('\n').some(line=>line.trim()===`- text: ${identity} hand`)) throw Error(`RTL inspector did not expose the exact physical copy ${identity}`);
        await returnToDecisions(page,true);
        if(await selection(target)!=='true') throw Error(`RTL inspection lost selected copy ${identity}`);
        const again=await reveal(target);
        await page.touchscreen.tap(again.right-24,again.top+Math.min(32,again.height/2));await pause(350);await focusFrame(page);
        if(await selection(target)!=='false') throw Error(`RTL exposed strip did not clear exactly ${identity}`);
      }
      const overlaps=[];
      for(const suit of ['Hearts','Diamonds']) {
        const first=targetFor(`Q of ${suit}, copy 1`),second=targetFor(`Q of ${suit}, copy 2`);
        await reveal(first);await reveal(second);
        const a=(await scrollGeometry(first)).rect,b=(await scrollGeometry(second)).rect;
        const step=a.right-b.right;
        if(Math.abs(a.top-b.top)>1/64||step<48-1/64||step>=a.width) throw Error(`RTL ${suit} must overlap from right to left with48px exposed strips: ${JSON.stringify({a,b,step})}`);
        overlaps.push({suit,step,cardWidth:a.width});
      }
      if(commands.length) throw Error('RTL inspection, selection or resizing submitted an authority command');
      report.cases.push({case:'rtl-overlapping-suit-copies-remain-independently-usable',mode:mode.name,width,height,pass:true,physicalCopies:identities,measured,overlaps,keyboardI:true,touchSelection:true,commandPosts:0});
    }
    await capture(page,`overlapping-hand-${mode.name}`);
  }
  await f.context.close();
}
async function handCollapseChecks(browser) {
  report.fixture='Four authorized Queen copies from the real Coup fixture; narrow220px diagnostic at200% text forces a same-suit horizontal scroll before widening390px';
  report.accessibilityBoundary='Chromium DOM/paint captures with keyboard focus and actual touch selection;220px is a stress diagnostic, not a native-device claim';
  const identity=/^Q of Hearts, copy 2(?: |$)/;
  const geometry=target=>target.evaluate(element=>{
    const hand=element.closest('[flt-semantics-identifier="private-hand-viewport"]');
    const r=element.getBoundingClientRect(),h=hand.getBoundingClientRect(),scrollGroups=[];
    for(let node=element.parentElement;node;node=node.parentElement) {
      const style=getComputedStyle(node);
      if(/auto|scroll|hidden/.test(style.overflowX)) scrollGroups.push({id:node.id,scrollLeft:node.scrollLeft,scrollWidth:node.scrollWidth,clientWidth:node.clientWidth});
    }
    return {id:element.id,x:r.x-h.x,width:r.width,height:r.height,selected:element.getAttribute('aria-current'),focused:element===document.activeElement,scrollGroups};
  });
  const prepare=async width=>{
    const f=await fixture(browser,'coup',1,4,{width:390,height:844}),page=f.page,posts=[];
    page.on('request',request=>{if(request.method()==='POST'&&new URL(request.url()).pathname.endsWith('/commands'))posts.push(request.url());});
    await setting(page,'Right to left');await setting(page,'200% text');
    await page.setViewportSize({width,height:844});await focusFrame(page);
    const target=page.getByRole('button',{name:identity});
    const beforeReveal=await geometry(target);
    await focusByKeyboard(page,target);await focusFrame(page);
    const g=await scrollGeometry(target),unit=1/64;
    if(g.rect.left<g.clip.left-unit||g.rect.right>g.clip.right+unit||g.rect.top<g.clip.top-unit||g.rect.bottom>g.clip.bottom+unit) throw Error(`Horizontal diagnostic source was not fully exposed: ${JSON.stringify(g)}`);
    await page.touchscreen.tap(g.rect.right-24,g.rect.top+32);await pause(350);
    await focusByKeyboard(page,target);await focusFrame(page);
    return {f,page,target,posts,beforeReveal};
  };
  const baseline=await prepare(390),expected=await geometry(baseline.target);
  await capture(baseline.page,'horizontal-fresh-wide-baseline');
  const narrow=await prepare(220),before=await geometry(narrow.target);
  if(!before.scrollGroups.some(group=>group.scrollWidth>group.clientWidth+1)||Math.abs(before.x-narrow.beforeReveal.x)<=1) throw Error(`Horizontal diagnostic did not actually scroll the clipped physical card: ${JSON.stringify({beforeReveal:narrow.beforeReveal,before})}`);
  if(before.selected!=='true'||!before.focused) throw Error('Horizontal diagnostic must start with selected focused physical copy');
  const authorityBefore=await confirmed(narrow.f,()=>true);
  await capture(narrow.page,'horizontal-scrolled-narrow');
  await narrow.page.setViewportSize({width:390,height:844});await focusFrame(narrow.page);
  const after=await geometry(narrow.target);
  await capture(narrow.page,'horizontal-collapsed-wide');
  if(after.id!==before.id||after.selected!=='true'||!after.focused) throw Error(`Horizontal collapse lost physical identity, selection or focus: ${JSON.stringify({before,after})}`);
  for(const key of ['x','width','height']) if(Math.abs(after[key]-expected[key])>1/64) throw Error(`Horizontal collapse has stale ${key}: ${JSON.stringify({expected,after})}`);
  const authorityAfter=await confirmed(narrow.f,()=>true);
  if(baseline.posts.length||narrow.posts.length||authorityAfter.version!==authorityBefore.version) throw Error('Horizontal collapse submitted an authority command');
  report.cases.push({case:'horizontal-scroll-collapse-preserves-physical-card-alignment-selection-and-focus',pass:true,beforeReveal:narrow.beforeReveal,before,after,expected,commandPosts:0});
  await baseline.f.context.close();await narrow.f.context.close();
}
async function finishChecks(browser) {
  report.fixture='Normal guest entry plus retained local gallery; no synthetic game or gameplay commands in this finish run';
  report.accessibilityBoundary='Browser DOM semantics, keyboard, focus and emulated motion preferences; actual screen-reader operation and browser zoom are not tested';
  const f=await guest(browser), page=f.page;
  const mutations=[];
  page.on('request',request=>{if(request.method()==='POST') mutations.push(new URL(request.url()).pathname);});
  const checkReflow=async()=>{
    const size=await page.evaluate(()=>({width:innerWidth,documentWidth:document.documentElement.scrollWidth}));
    if(size.width!==320||size.documentWidth>size.width) throw Error(`320px document reflow failed: ${JSON.stringify(size)}`);
  };
  const restoredFocus=async(target,label)=>{
    for(let frame=0;frame<30&&!(await target.evaluate(element=>element===document.activeElement));frame++) await pause(30);
    if(!(await target.evaluate(element=>element===document.activeElement))) {
      const active=await page.evaluate(()=>document.activeElement?.getAttribute('aria-label')||document.activeElement?.textContent||document.activeElement?.tagName);
      throw Error(`${label} did not restore keyboard focus (active: ${active})`);
    }
  };
  await checkReflow();await capture(page,'finish-room-setup-320');
  report.cases.push({case:'entry-and-room-setup-320-css-pixels',pass:true});
  const longDraft='room-'+('long-display-value-'.repeat(9))+'مرحبا';
  await typeText(page,'Room code',longDraft);
  const roomField=page.getByRole('textbox',{name:'Room code',exact:true});
  if(await roomField.inputValue()!==longDraft) throw Error('Long room-code draft was not entered');
  await setting(page,'200% text');
  // Flutter recreates the browser editing input while its controller retains
  // the draft. Refocus the visible field and continue editing to test that state.
  await roomField.click();await focusFrame(page);
  if(await roomField.inputValue()!==longDraft) throw Error('Long room-code draft was lost when text scale changed');
  await page.keyboard.press('End');await page.keyboard.type('-continued');
  const continuedDraft=longDraft+'-continued';
  if(await roomField.inputValue()!==continuedDraft) throw Error('Continued editing discarded the long room-code draft');
  await checkReflow();await capture(page,'finish-long-draft-320-text200');
  await setting(page,'200% text');
  await roomField.click();await focusFrame(page);
  if(await roomField.inputValue()!==continuedDraft) throw Error('Long room-code draft was lost when normal scale returned');
  if(mutations.length) throw Error(`Display-only room changes submitted an action: ${mutations.join(',')}`);
  report.cases.push({case:'long-mixed-direction-draft-retained-across-text-scaling-without-submission',pass:true,characters:longDraft.length});
  await page.goto(`${origin}/#/gallery`);
  await page.getByRole('button',{name:'Gallery',exact:true}).waitFor();
  for(const preference of ['no-preference','reduce']) {
    await page.emulateMedia({reducedMotion:preference});
    if(await page.evaluate(()=>matchMedia('(prefers-reduced-motion: reduce)').matches)!==(preference==='reduce')) throw Error('Reduced-motion media preference was not applied');
    await checkReflow();
    const queen=page.getByRole('button',{name:/^Q of Hearts, copy 1/}).first();
    await focusByKeyboard(page,queen);await page.keyboard.press('i');
    const close=page.getByRole('button',{name:'Close inspection',exact:true});
    await close.waitFor();
    // Semantics can appear before Material's modal fade/scale has completed.
    await pause(350);
    await capture(page,`finish-gallery-inspector-${preference}`);
    await page.keyboard.press('Escape');await close.waitFor({state:'hidden'});
    await restoredFocus(queen,`Inspector Escape (${preference})`);
    await page.keyboard.press('i');await close.waitFor();await close.click();await close.waitFor({state:'hidden'});
    await restoredFocus(queen,`Inspector close (${preference})`);
    report.cases.push({case:'gallery-inspector-keyboard-escape-close-focus-return',preference,pass:true});
  }
  await capture(page,'finish-gallery-320');
  report.cases.push({case:'gallery-320-css-pixels-and-both-motion-preference-interactions',pass:true});
  const settings=page.getByRole('button',{name:'Display settings',exact:true});
  // The header precedes the gallery cards. Firefox's forward traversal wraps
  // inside the body; reverse traversal reaches the header controls directly.
  const settingsKey=engine==='firefox'?'Shift+Tab':'Tab';
  await focusByKeyboard(page,settings,settingsKey);await page.keyboard.press('Enter');
  const done=page.getByRole('button',{name:'Done',exact:true});await done.waitFor();
  const reduced=page.getByRole('switch',{name:/Reduced motion/});
  await reduced.click();await reduced.click();
  await page.keyboard.press('Escape');await done.waitFor({state:'hidden'});
  await restoredFocus(settings,'Display settings Escape');
  report.cases.push({case:'gallery-display-settings-motion-toggle-and-escape-focus-return',keyboardPath:settingsKey,pass:true});
  if(mutations.length) throw Error('Gallery finish checks submitted an authority command');
  await f.context.close();
}
(async()=>{
  const browser=await engines[engine].launch({headless:true});report.version=browser.version();
  try {
    const verificationContext=await browser.newContext();
    for(const [file,expected] of Object.entries(report.webAssetsSha256)) {
      const response=await verificationContext.request.get(`${origin}/${file}`);
      if(!response.ok()||require('node:crypto').createHash('sha256').update(await response.body()).digest('hex')!==expected) throw Error(`Browser serves stale ${file}`);
    }
    await verificationContext.close();report.servedBuildVerified=true;
    if(process.env.CGMS_DIRECT_GEOMETRY_ONLY) {
      await directDropGeometryChecks(browser);
      if(report.errors.length)throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_DIRECT_INTERACTIONS_ONLY) {
      await directDropInteractionChecks(browser);
      await largeOpeningPreviewChecks(browser);
      if(report.errors.length)throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_TURN_DRAW_ONLY) {
      await scheduledDrawChecks(browser);
      if(report.errors.length)throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_TRADE_ONLY) {
      await tradeChecks(browser);report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_TRADE_FOCUS_ONLY) {
      const f=await fixture(browser,'trade');await tradeFocusBoundary(f);
      report.status='passed';await f.context.close();console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_GESTURES_ONLY) {
      const f=await fixture(browser,'purchase');
      for(const [file,expected] of Object.entries(report.webAssetsSha256)) {
        const response=await f.context.request.get(`${origin}/${file}`);
        if(!response.ok()||require('node:crypto').createHash('sha256').update(await response.body()).digest('hex')!==expected) throw Error(`Gesture browser serves stale ${file}`);
      }
      report.servedBuildVerified=true;
      await purchaseGestures(f);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';await f.context.close();console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_QUADRANT_ONLY) {
      await quadrantOnlineChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_HAND_COLLAPSE_ONLY) {
      await handCollapseChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_RTL_HAND_ONLY) {
      await rtlHandChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_KEYBOARD_SEQUENCE_ONLY) {
      await keyboardSequenceChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_CARD_DEMO_ONLY) {
      await demoCardChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_CARD_SIZES_ONLY) {
      await cardSizeChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_FINISH_ONLY) {
      await finishChecks(browser);
      if(report.errors.length) throw Error(report.errors.join('; '));
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(process.env.CGMS_CSP_ONLY) {
      report.fixture='Local static final build under the exact checked-in CSP; no authority gameplay in this run';
      report.input='render/semantics and served-byte verification';
      const page=await browser.newPage({viewport:report.viewport});
      page.on('pageerror',e=>report.errors.push(e.message));
      page.on('console',m=>{if(m.type()==='error') report.errors.push(JSON.stringify({message:m.text(),location:m.location()}));});
      const response=await page.goto(`${origin}/#/play`);
      const csp=response.headers()['content-security-policy'];
      if(!csp) throw Error('CSP response header missing');
      await page.locator('flt-semantics-host').getByRole('button').first().waitFor({timeout:30000});
      await pause(1000);
      for(const [file,expected] of Object.entries(report.webAssetsSha256)) {
        const asset=await page.request.get(`${origin}/${file}`);
        const servedHash=require('node:crypto').createHash('sha256').update(await asset.body()).digest('hex');
        if(!asset.ok()||servedHash!==expected) throw Error(`CSP surface serves a different ${file}`);
      }
      if(report.errors.length) throw Error(report.errors.join('; '));
      await page.screenshot({path:path.join(evidence,`${engine}-csp.png`)});
      report.cases.push({case:'production-csp-local-assets-render',pass:true});
      report.status='passed';console.log(JSON.stringify(report));return;
    }
    if(!process.env.CGMS_EFFECTS_ONLY) {
    const {page,context}=await fixture(browser,'opening');
    for(const surface of ['Table','Hand','Decisions']) {
      await (surface==='Table'?table(page):surface==='Hand'?hand(page):decisions(page));
      await pause(150);await capture(page,`opening-${surface.toLowerCase()}`);
    }
    await card(page,/^4 of Clubs, copy 1/);await action(page,'Offer');
    await page.getByText(/Give · 4 of Clubs, copy 1/).waitFor();
    await cancelDraft(page);
    await hand(page);
    await activate(page,page.getByRole('button',{name:/^4 of Clubs, copy 1/}));await pause(350);
    const inspection=page.getByRole('button',{name:/^4 of Clubs, copy 1/}).first();
    await focusByKeyboard(page,inspection);await page.keyboard.press('i');
    await page.getByRole('button',{name:/^Close card context(?: Close card context)?$/}).waitFor();
    await returnToDecisions(page,true);
    await selectedPhysical(page);
    report.cases.push({case:'action-switch-routes-cards-to-visible-selection',pass:true});
    report.cases.push({case:'keyboard-card-inspection-preserves-selection',pass:true});
    const sizes=[[390,844],[360,800],[599,900],[600,900],[768,1024],[1024,768],[1049,900],[1050,900],[1280,720],[1440,900],[1280,600],[599,900],[600,900],[599,900],[1049,900],[1050,900],[1049,900],[390,844],[844,390],[390,844],[768,1024],[1024,768]];
    for(const [width,height] of (process.env.CGMS_QUICK ? [] : sizes)) {
      await page.setViewportSize({width,height});await pause(120);
      await decisions(page);
      await selectedPhysical(page);
      if([390,768,1440].includes(width)) await page.screenshot({path:path.join(evidence,`${engine}-${width}x${height}.png`)});
      report.cases.push({case:'adaptive-selection',width,height,pass:true});
    }
    if(!process.env.CGMS_QUICK) {
      for(const name of ['Artwork','Right to left','200% text']) await setting(page,name);
      for(const [width,height] of [[390,844],[768,1024],[1440,900]]) {
        await page.setViewportSize({width,height});await decisions(page);
        await selectedPhysical(page);
        await page.screenshot({path:path.join(evidence,`${engine}-${width}-rtl-200-placeholder.png`)});
        report.cases.push({case:'rtl-200-placeholder-selection',width,height,pass:true});
      }
      for(const name of ['Artwork','Right to left','200% text']) await setting(page,name);
    }
    await page.setViewportSize(report.viewport);await decisions(page);

    const onlineUrl=page.url();
    await page.goto(`${origin}/#/play`);await pause(200);
    await page.goBack();await page.getByRole('heading',{name:'CGMS · Online',exact:true}).waitFor();
    await decisions(page);await selectedPhysical(page);
    await page.goForward();await pause(200);await page.goBack();
    await page.getByRole('heading',{name:'CGMS · Online',exact:true}).waitFor();await decisions(page);
    if(page.url()!==onlineUrl) throw Error('Browser history lost match URL');
    report.cases.push({case:'browser-back-forward-retains-match-and-draft',pass:true});
    const focusLabels=[];
    for(let i=0;i<12;i++) {
      await page.keyboard.press('Tab');await pause(30);
      focusLabels.push(await page.evaluate(()=>{const e=document.activeElement;return e?.getAttribute('aria-label')||e?.textContent||e?.tagName;}));
    }
    if(new Set(focusLabels.filter(Boolean)).size<3) throw Error(`Keyboard traversal stuck: ${JSON.stringify(focusLabels)}`);
    report.cases.push({case:'keyboard-tab-traversal',pass:true,distinctStops:new Set(focusLabels.filter(Boolean)).size});
    await page.screenshot({path:path.join(evidence,`${engine}-${report.viewport.width}-keyboard-focus.png`)});
    await closeActions(page);
    const targets=await touchTargets(page);
    const smallTargets=targets.filter(t=>t.width<48-1e-5 || t.height<48-1e-5);
    if(smallTargets.length) throw Error(`Primary targets below48: ${JSON.stringify(smallTargets)}`);
    report.cases.push({case:'primary-touch-targets-at-least-48',pass:true,measurement:'each original control exposed by scrolling and trial hit-tested',count:targets.length,minimumWidth:Math.min(...targets.map(t=>t.width)),minimumHeight:Math.min(...targets.map(t=>t.height)),targets});
    await action(page,'Open Series');
    await guidance(page);
    await page.locator('flt-semantics-host').getByText(/Costs and timing[\s\S]*Own turn: normally open one new number series/).waitFor();
    report.cases.push({case:'selected-opening-cost-and-timing-guidance',pass:true});
    await page.getByText('Series · Clubs',{exact:true}).waitFor();
    if(await page.getByRole('button',{name:/^Suit /}).count()) throw Error('Physical opening was retargetable through a universal suit form');
    await capture(page,'opening-card-confirmation');
    report.cases.push({case:'physical-opening-preserves-exact-suit-and-all-held-card-cost',pass:true});
    await activate(page,page.getByRole('button',{name:'Confirm Open Series',exact:true}));
    await page.locator('flt-semantics-host').getByText('Waiting for seat 2',{exact:true}).waitFor();
    // Accepted commands clear their draft. The same table must not offer a
    // fresh ordinary move while the authority is collecting responses.
    if(await page.getByRole('button',{name:'Confirm Open Series',exact:true}).count()) throw Error('Accepted opening retained a confirmation control during pending responses');
    await card(page,/^4 of Clubs, copy 1/);await decisions(page);
    if(await page.getByRole('button',{name:/^Open Series(?: · .*)?$/}).count()) throw Error('Ordinary opening remained offered during pending responses');
    report.cases.push({case:'ordinary-action-disabled-during-response',pass:true});
    for(const seat of otherSeats) {
      const other=await fixture(browser,'opening',seat);await decisions(other.page);await pause(200);
      await activate(other.page,other.page.getByRole('button',{name:'Pass response',exact:true}));
      await other.page.locator('flt-semantics-host').getByText(seat<players ? `Waiting for seat ${seat+1}` : 'Seat 1’s turn',{exact:true}).waitFor();
      await other.context.close();
    }
    await page.locator('flt-semantics-host').getByText('Seat 1’s turn',{exact:true}).waitFor();
    report.cases.push({case:'opening-ordered-responses',pass:true});
    await context.close();
    if(process.env.CGMS_MATRIX_ONLY) {if(report.errors.length) throw Error(report.errors.join('; '));report.status='passed';console.log(JSON.stringify(report));return;}
    const purchase=await fixture(browser,'purchase');
    await purchaseGestures(purchase);
    await typeText(purchase.page,'Quantity','1');
    await purchase.page.getByText(/Payment · 10 of Spades, copy 1/).waitFor();
    await decisions(purchase.page);
    await activate(purchase.page,purchase.page.getByRole('button',{name:'Confirm Purchase',exact:true}));
    await purchase.page.locator('flt-semantics-host').getByText('Waiting for seat 2',{exact:true}).waitFor();
    await responses(browser,'purchase');
    await purchase.page.locator('flt-semantics-host').getByText('Seat 1’s turn',{exact:true}).waitFor();
    report.cases.push({case:'purchase-explicit-quantity-payment-responses',pass:true});
    await purchase.context.close();
    await tradeChecks(browser);
    const recovery=await fixture(browser,'trade');const replayed=[];
    const fault=new ExpectedAbort();fault.begin();recoveryFaults.set(recovery.page,fault);
    await recovery.page.route('**/commands',async route=>{replayed.push(route.request().postData());await route.fetch();fault.record(route.request().url());await route.abort('failed');});
    await recovery.page.route('**/commands/*',async route=>{fault.record(route.request().url());await route.abort('failed');});
    await activate(recovery.page,recovery.page.getByRole('button',{name:'End Turn',exact:true}));
    await recovery.page.getByRole('button',{name:'Check operation',exact:true,disabled:false}).waitFor({timeout:15000});
    // Submission already has a bounded two-attempt identical-envelope protocol.
    // Resize must add no attempts after that initial budget has completed.
    const originalAttempts=replayed.length;
    if(originalAttempts!==2||new Set(replayed).size!==1) throw Error('Initial lost-reply submission exceeded its identical-envelope two-attempt contract');
    for(const [width,height] of [[599,900],[600,900],[1049,900],[1050,900],[1049,900],[599,900]]) {
      await recovery.page.setViewportSize({width,height});await focusFrame(recovery.page);
      await recovery.page.getByRole('button',{name:'Check operation',exact:true}).waitFor();
      if(replayed.length!==originalAttempts||new Set(replayed).size!==1) throw Error('Adaptive transition replayed or changed an uncertain authority command');
      report.cases.push({case:'pending-original-operation-retained-across-boundary',width,height,pass:true,initialIdenticalAttempts:originalAttempts,additionalCommandPosts:0});
    }
    await recovery.page.setViewportSize({width:844,height:390});await setting(recovery.page,'200% text');
    const recoverControl=recovery.page.getByRole('button',{name:'Check operation',exact:true});
    // DOM scrollIntoView cannot establish that Flutter's painted viewport moved.
    // Scroll the notice region with real input and activate the visible control.
    await exposeRecord(recovery.page,recoverControl);
    const recoveryBounds=await recoverControl.boundingBox();
    if(!recoveryBounds||recoveryBounds.y<0||recoveryBounds.y+recoveryBounds.height>390) throw Error('Recovery action clipped in short large-text view');
    await recovery.page.screenshot({path:path.join(evidence,`${engine}-recovery-844x390-200.png`)});
    const shortReconcile=recovery.page.waitForRequest(r=>r.url().includes('/commands/')&&r.method()==='GET');
    await recoverControl.tap();await shortReconcile;
    await recovery.page.getByRole('button',{name:'Check operation',exact:true}).waitFor();
    report.cases.push({case:'short-rotated-large-text-recovery-control',pass:true,viewport:{width:844,height:390},textScale:2});
    await setting(recovery.page,'200% text');await recovery.page.setViewportSize(report.viewport);
    await recovery.page.reload();await recovery.page.getByRole('button',{name:'Check operation',exact:true}).waitFor();
    await recovery.page.unroute('**/commands');await recovery.page.unroute('**/commands/*');
    await activate(recovery.page,recovery.page.getByRole('button',{name:'Check operation',exact:true}));
    await recovery.page.getByRole('button',{name:'Check operation',exact:true}).waitFor({state:'hidden'});
    if(replayed.length!==originalAttempts || new Set(replayed).size!==1) throw Error('Receipt recovery replayed or changed the original command envelope');
    fault.end();
    report.cases.push({case:'lost-reply-reload-original-operation-reconciliation',pass:true});await recovery.context.close();
    }
    const loan=await fixture(browser,'loan');
    await card(loan.page,/^K of Hearts, copy 1/);await decisions(loan.page);
    await activate(loan.page,loan.page.getByRole('button',{name:/^Loan this intact formation/}).first());
    await activate(loan.page,loan.page.getByRole('button',{name:'Seat 2',exact:true}));
    await loan.page.getByText(/Give · .*K of Hearts, copy 1/).waitFor();
    const submitLoan=loan.page.getByRole('button',{name:'Confirm Offer',exact:true});
    await focusByKeyboard(loan.page,submitLoan);
    const [loanReply]=await Promise.all([
      loan.page.waitForResponse(response=>response.request().method()==='POST'&&new URL(response.url()).pathname.endsWith('/commands')),
      loan.page.keyboard.press('Space'),
    ]);
    if(loanReply.status()!==200) throw Error('Explicit loan offer was not acknowledged');
    await submitLoan.waitFor({state:'hidden'});await focusFrame(loan.page);
    await records(loan.page);await loan.page.getByRole('group',{name:/Private offer · revision/}).waitFor();
    const borrower=await fixture(browser,'loan',2);await records(borrower.page);
    await activate(borrower.page,borrower.page.getByRole('button',{name:'Accept Offer',exact:true}));
    await borrower.page.locator('flt-semantics-host').getByText('Waiting for seat 3',{exact:true}).waitFor();
    await responses(browser,'loan',acceptorResponses);
    const resolvedLoan=await confirmed(borrower,s=>!s.projection.board.window_id);
    if(!(resolvedLoan.projection.board.formations||[]).some(f=>f.controller===2&&f.spec.kind==='underground')) throw Error('Accepted intact loan did not transfer the Underground formation');
    await closeActions(borrower.page);await capture(borrower.page,'loan-resolved-table');
    report.cases.push({case:'intact-loan-alliance-underground',pass:true});await loan.context.close();await borrower.context.close();
    const attack=await fixture(browser,'attack');
    await cancelDraft(attack.page);await closeActions(attack.page);
    for(const [width,height] of [[390,844],[768,1024],[1440,900]]) {
      await attack.page.setViewportSize({width,height});await focusFrame(attack.page);
      if(await attack.page.getByRole('button',{name:/^Close card context/}).count()) throw Error('Idle board retained a contextual overlay');
      await capture(attack.page,`unobstructed-public-cards-idle-${width}x${height}`);
    }
    await attack.page.setViewportSize(report.viewport);await focusFrame(attack.page);
    await card(attack.page,/^8 of Clubs, copy 1/);
    await action(attack.page,'Attack');await guidance(attack.page);
    await attack.page.locator('flt-semantics-host').getByText(/Costs and timing[\s\S]*Own turn; choose unused exposed Clubs and a legal target\. Physical Clubs become used at legal declaration, including if the attack is canceled\. Exchange costs apply only after successful resolution\./).waitFor();
    await attack.page.screenshot({path:path.join(evidence,`${engine}-${process.env.CGMS_WIDTH||390}-attack-costs.png`)});
    report.cases.push({case:'selected-attack-declaration-cancellation-and-resolution-costs',pass:true});
    await decisions(attack.page);await activate(attack.page,attack.page.getByRole('button',{name:'Choose the target card',exact:true}));
    await table(attack.page);await card(attack.page,/^8 of Hearts, copy 1/);await decisions(attack.page);
    await capture(attack.page,'attack-exact-target-confirmation');
    await activate(attack.page,attack.page.getByRole('button',{name:'Confirm Attack',exact:true}));
    await attack.page.locator('flt-semantics-host').getByText('Waiting for seat 2',{exact:true}).waitFor();
    await capture(attack.page,'attack-pending-response');
    await responses(browser,'attack');await confirmed(attack,s=>!s.projection.board.window_id);
    report.cases.push({case:'physical-attack-ordered-response-exchange',pass:true});await attack.context.close();
    const compensated=[];const compensationEvidence=[];
    for(const seat of [2,3]) {
      const f=await fixture(browser,'compensation',seat);
      const before=await confirmed(f,()=>true);
      const own=before.projection.board.players.find(p=>p.seat===seat);
      compensated.push({...f,seat,privateBefore:own.hand_count+own.ace_count,countsBefore:{hand:own.hand_count,aces:own.ace_count}});
    }
    await responses(browser,'compensation');
    for(const f of compensated) {
      const after=await confirmed(f,s=>!s.projection.board.window_id && !(s.projection.online.draw_queue||[]).length && s.projection.board.players.some(p=>p.seat===f.seat && p.hand_count+p.ace_count===f.privateBefore+1));
      if(after.projection.board.effects.some(e=>e.kind==='compensation'&&e.losses!==0)) throw Error('Compensation entitlement was not consumed');
      const ownAfter=after.projection.board.players.find(p=>p.seat===f.seat);
      compensationEvidence.push({seat:f.seat,before:f.countsBefore,after:{hand:ownAfter.hand_count,aces:ownAfter.ace_count}});
      const acquired=after.projection.board.cards.filter(c=>c.controller===f.seat && ['hand','concealed-ace'].includes(c.zone));
      if(acquired.length!==f.privateBefore+1) throw Error('Compensation private card projection does not match acquired count');
      await hand(f.page);
      for(const entry of acquired){
        const c=entry.card;const rank=({1:'A',11:'J',12:'Q',13:'K'})[c.rank]||String(c.rank);const suit=c.suit[0].toUpperCase()+c.suit.slice(1);
        await f.page.getByRole('button',{name:new RegExp(`^${rank} of ${suit}, copy ${c.deck} ${entry.zone}`)}).waitFor();
      }
      await f.context.close();
    }
    report.cases.push({case:'compensation-return-shuffle-two-recipients-resume',pass:true,recipients:compensationEvidence,orderEvidence:'paired fixture engine test proves ascending queue; browser verifies each private holding including concealed Aces and completed counters'});
    const confinement=await fixture(browser,'confinement',2);
    await card(confinement.page,/^A of Diamonds, copy 1/);
    await action(confinement.page,'Confinement');
    await confinement.page.getByText(/Source card · A of Diamonds, copy 1/).waitFor();
    await capture(confinement.page,'defence-confinement-context');
    await activate(confinement.page,confinement.page.getByRole('button',{name:'Confirm Confinement',exact:true}));
    await confirmed(confinement,s=>s.projection.board.players.find(p=>p.seat===1).confined===true);
    report.cases.push({case:'confinement-cancels-hostile-action',pass:true});await confinement.context.close();
    const justice=await fixture(browser,'justice');await decisions(justice.page);
    await capture(justice.page,'effect-decision');
    if(await justice.page.getByRole('button',{name:'Pass response',exact:true}).count()) throw Error('Response pass is offered during effect input');
    await justice.page.locator('flt-semantics-host').getByText('Choose how to resolve Justice',{exact:true}).waitFor();
    report.cases.push({case:'effect-input-is-not-response-pass',pass:true});
    for(const rank of Array.from({length:players-1},(_,i)=>i+4)) {
      const before=await confirmed(justice,s=>Boolean(s.projection.board.decision_id));
      const physical=before.projection.board.cards.find(p=>p.card.rank===rank&&p.card.suit==='hearts');
      if(!physical||(before.projection.board.choices||[]).every(choice=>choice.length!==1||choice[0]!==physical.card.id)) throw Error('Justice fixture lacks the expected current authorized physical choice');
      await decisions(justice.page);
      await activate(justice.page,justice.page.getByRole('button',{name:`${rank} of Hearts, copy 1`,exact:true}));
      await decisions(justice.page);await activate(justice.page,justice.page.getByRole('button',{name:'Confirm Decision',exact:true}));
      await confirmed(justice,s=>s.projection.board.decision_id!==before.projection.board.decision_id);
    }
    report.cases.push({case:'justice-staged-distinct-card-selections',pass:true});await justice.context.close();
    const coup=await fixture(browser,'coup');
    await card(coup.page,/^Q of Hearts, copy 1/);await action(coup.page,'Coup');
    await activate(coup.page,coup.page.getByRole('button',{name:'Confirm Coup',exact:true}));
    await coup.page.locator('flt-semantics-host').getByText(/Board closed|settlement/).first().waitFor();
    report.cases.push({case:'immediate-coup',pass:true});await coup.context.close();
    const settlement=await fixture(browser,'settlement');await decisions(settlement.page);
    await capture(settlement.page,'settlement');
    await activate(settlement.page,settlement.page.getByRole('button',{name:'Pay promise',exact:true}));
    await confirmed(settlement,s=>s.projection.promises[0].status==='paid');
    await gameOption(settlement.page,'Finish Settlement');
    await activate(settlement.page,settlement.page.getByRole('button',{name:'Confirm Finish Settlement',exact:true}));
    await pause(300);
    for(const seat of otherSeats) {
      const other=await fixture(browser,'settlement',seat);await decisions(other.page);
      await gameOption(other.page,'Finish Settlement');
      await activate(other.page,other.page.getByRole('button',{name:'Confirm Finish Settlement',exact:true}));
      await pause(400);await other.context.close();
    }
    await confirmed(settlement,s=>s.projection.online.completed_games===1);
    await settlement.page.reload();
    await settlement.page.locator('flt-semantics-host').getByText(/Seat .*turn|Waiting for seat/).first().waitFor();
    report.cases.push({case:'promise-payment-settlement-next-game-reload',pass:true});await settlement.context.close();
    const finalSettlement=await fixture(browser,'final-settlement');await decisions(finalSettlement.page);
    await activate(finalSettlement.page,finalSettlement.page.getByRole('button',{name:'Pay promise',exact:true}));await confirmed(finalSettlement,s=>s.projection.promises[0].status==='paid');
    await gameOption(finalSettlement.page,'Finish Settlement');
    await activate(finalSettlement.page,finalSettlement.page.getByRole('button',{name:'Confirm Finish Settlement',exact:true}));await pause(300);
    for(const seat of otherSeats) {const other=await fixture(browser,'final-settlement',seat);await gameOption(other.page,'Finish Settlement');await activate(other.page,other.page.getByRole('button',{name:'Confirm Finish Settlement',exact:true}));await pause(400);await other.context.close();}
    await confirmed(finalSettlement,s=>Array.isArray(s.projection.online.ranks)&&s.projection.online.ranks.length===players);
    await decisions(finalSettlement.page);
    await finalSettlement.page.getByText(/Match ranks by seat:/).waitFor();
    await capture(finalSettlement.page,'final-standings');
    report.cases.push({case:'final-match-shared-ranks',pass:true});await finalSettlement.context.close();
    if(process.env.CGMS_EFFECTS_ONLY) {if(report.errors.length) throw Error(report.errors.join('; '));report.status='passed';console.log(JSON.stringify(report));return;}
    const owner=await guest(browser);
    await capture(owner.page,'room-setup');
    if(players===3) {await owner.page.getByRole('button',{name:'Players 4',exact:true}).click();await pause(350);await owner.page.getByRole('menuitem',{name:'3',exact:true}).click();await owner.page.getByRole('button',{name:'Players 3',exact:true}).waitFor();await pause(350);}
    const created=owner.page.waitForResponse(r=>r.url()===`${origin}/v1/rooms` && r.request().method()==='POST');
    await owner.page.getByRole('button',{name:'Create room',exact:true}).click();
    const room=await (await created).json();const members=[];
    await owner.page.locator('flt-semantics-host').getByText(`Room code: ${room.id}`,{exact:false}).waitFor();
    await owner.page.waitForURL(url=>new URLSearchParams(url.hash.split('?')[1]).get('room')===room.id);
    await owner.page.reload();
    await owner.page.locator('flt-semantics-host').getByText(`Room code: ${room.id}`,{exact:false}).waitFor();
    await owner.page.getByRole('button',{name:'Create invitation',exact:true}).waitFor();
    // Capture before an invitation is issued: invitation credentials must never
    // become screenshot artifacts even though the authorized owner can read them.
    await capture(owner.page,'room-lobby');
    if(await owner.page.getByRole('button',{name:'Create room',exact:true}).count()) throw Error('Reload discarded the active lobby');
    report.cases.push({case:'owner-lobby-reload-restores-authorized-room',pass:true});
    if(await owner.page.getByRole('textbox',{name:'Room code',exact:true}).count()) throw Error('Room code must not expose a phantom editable field');
    if(engine==='chromium') {
      // Headless Chromium defaults to denied clipboard writes even on a focused
      // secure loopback page. Exercise the denial UI before granting only write
      // access for this origin; no clipboard read or transport stubbing is used.
      await owner.page.bringToFront();
      await owner.page.getByRole('button',{name:'Copy Room code',exact:true}).tap();
      await owner.page.locator('flt-semantics-host').getByText('Clipboard access is unavailable. The code remains visible; check your browser permissions and try Copy again.',{exact:true}).waitFor();
      await owner.page.locator('flt-semantics-host').getByText(`Room code: ${room.id}`,{exact:false}).waitFor();
      if(await owner.page.getByText('Connection unavailable. Retry when the authority is reachable.',{exact:true}).count()) throw Error('Clipboard denial was reported as authority failure');
      await owner.context.grantPermissions(['clipboard-write'],{origin});
      report.cases.push({case:'clipboard-denial-retains-code-and-specific-recovery-guidance',pass:true});
      report.clipboardPermission='Chromium: explicit clipboard-write for loopback origin after real denial; no clipboard-read';
    }
    await owner.page.getByRole('button',{name:'Copy Room code',exact:true}).tap();
    await owner.page.locator('flt-semantics-host').getByText('Room code copied',{exact:true}).waitFor();
    if(await owner.page.getByRole('button',{name:'Start match',exact:true}).isEnabled()) throw Error('Owner can start before room fills');
    for(let n=0;n<players-1;n++) {
      const invitationButton=owner.page.getByRole('button',{name:'Create invitation',exact:true});
      if(engine==='webkit') {
        // The copy snackbar can leave this Flutter control clipped below the
        // viewport. WebKit's DOM autoscroll can change the semantics node
        // between pointerdown/up, dropping click entirely. Real Tab traversal
        // lets Flutter reveal the control before the actual pointer action.
        await focusByKeyboard(owner.page,invitationButton);await focusFrame(owner.page);
        report.invitationPointerPreparation='WebKit: real Tab traversal and settled Flutter frames before pointer click';
      }
      const invitationResponse=owner.page.waitForResponse(r=>r.url().endsWith('/invitations'));
      await invitationButton.click();
      const invitation=(await (await invitationResponse).json()).invitation;
      invitationValues.add(invitation);privateLobbyPages.add(owner.page);
      await owner.page.locator('flt-semantics-host').getByText(`Invitation: ${invitation}`,{exact:false}).waitFor();
      if(await owner.page.getByRole('textbox',{name:'Invitation',exact:true}).count()) throw Error('Invitation must not expose a phantom editable field');
      if(n===0) {
        await owner.page.getByRole('button',{name:'Copy Invitation',exact:true}).tap();
        await owner.page.locator('flt-semantics-host').getByText('Invitation copied',{exact:true}).waitFor();
      }
      const member=await guest(browser);members.push(member);
      privateLobbyPages.add(member.page);
      const draftRoute=member.page.url();let draftMutations=0;
      const observeDraft=request=>{if(request.method()==='POST') draftMutations++;};
      member.page.on('request',observeDraft);
      await typeText(member.page,'Room code',room.id);await typeText(member.page,'Invitation',invitation);
      if(engine==='firefox') await focusByKeyboard(member.page,member.page.getByRole('button',{name:'Join room',exact:true}));
      member.page.off('request',observeDraft);
      if(draftMutations!==0 || member.page.url()!==draftRoute) throw Error('Join draft entry or navigation unexpectedly submitted an action');
      await member.page.getByRole('button',{name:'Join room',exact:true}).click();
      await member.page.getByRole('button',{name:'Refresh seats',exact:true}).waitFor();
      if(n===0) {
        await member.page.waitForURL(url=>new URLSearchParams(url.hash.split('?')[1]).get('room')===room.id);
        if(member.page.url().includes(invitation)) throw Error('Invitation leaked into lobby URL');
        await member.page.reload();
        await member.page.locator('flt-semantics-host').getByText(`Room code: ${room.id}`,{exact:false}).waitFor();
        report.cases.push({case:'member-lobby-reload-preserves-role-without-invitation',pass:true});
      }
      if(await member.page.getByRole('button',{name:'Create invitation',exact:true}).count()) throw Error('Member was shown owner invitation controls');
      if(await member.page.getByRole('button',{name:'Start match',exact:true}).isEnabled()) throw Error('Member was shown enabled owner start');
    }
    await owner.page.getByRole('button',{name:'Refresh seats',exact:true}).click();
    const ownerStart=owner.page.getByRole('button',{name:'Start match',exact:true});
    // Refresh includes an HTTP read and a Flutter frame before the button updates.
    for(let i=0;i<100 && !(await ownerStart.isEnabled());i++) await pause(50);
    if(!(await ownerStart.isEnabled())) throw Error('Full room owner start unavailable');
    report.cases.push({case:'room-capacity-and-owner-affordances',pass:true});
    report.cases.push({case:'accessible-static-room-invitation-and-copy-feedback',pass:true});
    const roomFault=new ExpectedAbort();roomFault.begin();recoveryFaults.set(owner.page,roomFault);
    const started=[];
    await owner.page.route('**/rooms/*/matches',async route=>{started.push(route.request().postData());await route.fetch();roomFault.record(route.request().url());await route.abort('failed');});
    await owner.page.getByRole('button',{name:'Start match',exact:true}).click();
    await owner.page.getByRole('button',{name:'Recover room operation',exact:true}).waitFor({timeout:15000});
    if(await owner.page.getByRole('button',{name:'Start match',exact:true}).isEnabled()) throw Error('Unknown start allows another start');
    await owner.page.unroute('**/rooms/*/matches');
    const recovered=owner.page.waitForRequest(r=>r.url().endsWith(`/rooms/${encodeURIComponent(room.id)}/matches`)&&r.method()==='POST');
    await owner.page.getByRole('button',{name:'Recover room operation',exact:true}).click();
    const recoveryRequest=await recovered;
    if(started.length!==1||recoveryRequest.postData()!==started[0]) throw Error('Room recovery changed original operation');
    roomFault.end();report.cases.push({case:'lost-start-reply-lobby-original-operation-recovery',pass:true});
    await owner.page.waitForURL(/match=/);
    await owner.page.reload();await owner.page.locator('flt-semantics-host').getByText(/Seat .*turn|Waiting for seat/).first().waitFor();
    const member=members[0];await member.page.getByRole('button',{name:'Refresh seats',exact:true}).click();
    await member.page.getByRole('button',{name:'Enter match',exact:true}).click();await member.page.waitForURL(/match=/);
    for(const surface of ['Table','Hand','Decisions']) {
      await (surface==='Table'?table(owner.page):surface==='Hand'?hand(owner.page):decisions(owner.page));await pause(150);
      await owner.page.screenshot({path:path.join(evidence,`${engine}-${process.env.CGMS_WIDTH||390}-normal-${surface.toLowerCase()}.png`)});
    }
    report.cases.push({case:'normal-guest-room-deal-entry-reload',pass:true});
    await owner.context.close();for(const member of members) await member.context.close();
    if(report.errors.length) throw Error(`Browser errors: ${report.errors.join('; ')}`);
    report.status='passed';
  } catch(error) {report.status='failed';report.failure=safeDiagnostic(error.message); for(const [i,ctx] of browser.contexts().entries()) {for(const p of ctx.pages()){if(privateLobbyPages.has(p)){report.privateFailureCaptureOmitted=true;continue;}await p.screenshot({path:path.join(evidence,`${engine}-failure-${i}.png`)});report.failurePageCount=(report.failurePageCount||0)+1;}} throw new Error(safeDiagnostic(error.stack||error)); } finally {fs.writeFileSync(path.join(evidence,`${engine}-${process.env.CGMS_GESTURES_ONLY?'gestures':process.env.CGMS_QUADRANT_ONLY?'quadrants':process.env.CGMS_CARD_DEMO_ONLY?'card-demo':process.env.CGMS_CARD_SIZES_ONLY?'card-sizes':process.env.CGMS_FINISH_ONLY?'finish':process.env.CGMS_CSP_ONLY?'csp':process.env.CGMS_MATRIX_ONLY?'matrix':(process.env.CGMS_WIDTH||390)}${players===3?'-3p':''}-report.json`),JSON.stringify(report,null,2));await browser.close();}
  console.log(JSON.stringify(report));
})().catch(e=>{console.error(e);process.exitCode=1});
