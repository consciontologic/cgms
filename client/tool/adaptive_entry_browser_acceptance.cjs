// Normal application entry, without fixtures or layout overrides. Captures are
// disposable; only the reviewed JSON result belongs in retained evidence.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const {isBrowserError} = require('./browser_faults.cjs');
if (!process.env.CGMS_CONTAINER) process.env.PLAYWRIGHT_BROWSERS_PATH = path.resolve(__dirname, '../.browser-tools/browsers');
const engines = require('../.browser-tools/node_modules/playwright');
const engine = process.env.CGMS_BROWSER || 'chromium';
const origin = process.env.CGMS_BROWSER_ORIGIN || 'http://127.0.0.1:8080';
const run = `${engine}-adaptive-entry-${Date.now()}`;
const evidence = path.resolve(process.env.CGMS_EVIDENCE_DIR || path.join(__dirname, '../.browser-tools/artifacts', run));
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const report = {run, engine, status: 'running', cases: [], errors: [], assets: {}, screenshots: [],
  harnessSha256: hash(fs.readFileSync(__filename)),
  scope: 'Normal / entry and Auto; shared phone-reference vertical layout at every width, real touch/keyboard and live resizes. Local demonstration, no native-device claim.'};
const frame = page => page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
const button = (page, name) => page.getByRole('button', {
  name: name === 'Close actions' ? /^Close actions(?: Close actions)?$/ : name,
  exact: true,
});
const publicHeading = page => page.locator('[flt-semantics-identifier="public-square"]');
// Flutter briefly includes the visible tooltip in the button's accessible name
// after pointer activation. Match its observed exact forms, not arbitrary text.
const actionHeading = page => page.getByRole('button', {name: /^Choose an action(?: Choose an action)?$/});
const handHeading = page => page.getByRole('heading', {name: /^Your hand · /});
const handPane = page => page.locator('[flt-semantics-identifier="private-hand-viewport"]');
const publicTable = page => page.locator('[flt-semantics-identifier="public-square"]');
async function expose(page, target) {
  await target.waitFor();
  for (let attempt = 0; attempt < 30; attempt++) {
    const viewport=page.viewportSize();
    const geometry=await target.evaluate(element=>{
      const rect=element.getBoundingClientRect().toJSON();
      const clip={left:0,top:0,right:innerWidth,bottom:innerHeight};let scroll;
      for(let p=element.parentElement;p;p=p.parentElement) {
        const style=getComputedStyle(p),r=p.getBoundingClientRect();
        if(/auto|scroll|hidden|clip/.test(style.overflowY)) {
          clip.top=Math.max(clip.top,r.top);clip.bottom=Math.min(clip.bottom,r.bottom);
          if(!scroll&&p.getAttribute('role')==='group') scroll=r.toJSON();
        }
      }
      return {rect,clip,scroll};
    });
    const {rect,clip,scroll}=geometry;
    if(rect.height>0&&rect.top>=clip.top-1/64&&rect.bottom<=clip.bottom+1/64) return;
    assert.ok(scroll,`offscreen control must have a deliberate scroll area: ${JSON.stringify(geometry)}`);
    const top=Math.max(0,scroll.top),bottom=Math.min(viewport.height,scroll.bottom);
    await page.mouse.move(Math.max(1,Math.min(viewport.width-1,scroll.left+scroll.width/2)),(top+bottom)/2);
    await page.mouse.wheel(0,rect.top+rect.height/2-(top+bottom)/2);
    await frame(page);
  }
  throw new Error(`Could not expose vertical board control ${await target.textContent()}`);
}
async function nav(page, name) {
  const close=button(page,'Close actions');
  if(name==='Table'&&await close.count()) {await close.click();await frame(page);}
  if(name==='Decisions'&&!(await close.count())) {await actionHeading(page).click();await frame(page);}
  await expose(page, name === 'Table' ? publicHeading(page) : name === 'Hand' ? handHeading(page) : actionHeading(page));
  await frame(page);
}
async function keyboardFocus(page, target) {
  const visited = new Set();
  for (let n = 0; n < 100; n++) {
    if (await target.evaluateAll(elements => elements.some(e => e === document.activeElement))) return;
    visited.add(await page.evaluate(() => {
      const e = document.activeElement;
      return `${e?.tagName}#${e?.id}: ${e?.getAttribute('aria-label') || e?.textContent?.slice(0,100)}`;
    }));
    await page.keyboard.press('Tab'); await frame(page);
  }
  const requested=await target.evaluateAll(elements=>elements.map(element=>({
    text:element.getAttribute('aria-label')||element.textContent,
    html:element.outerHTML,rect:element.getBoundingClientRect().toJSON(),
  })));
  throw new Error(`Keyboard could not reach normal-entry control ${JSON.stringify(requested)}: ${JSON.stringify([...visited])}`);
}
async function assertRevealedCard(target) {
  const geometry=await target.evaluate(element=>{
    const rect=element.getBoundingClientRect().toJSON();
    const clip={left:0,top:0,right:innerWidth,bottom:innerHeight};
    for(let parent=element.parentElement;parent;parent=parent.parentElement) {
      const style=getComputedStyle(parent),bounds=parent.getBoundingClientRect();
      if(/auto|scroll|hidden|clip/.test(style.overflowX)) {
        clip.left=Math.max(clip.left,bounds.left);clip.right=Math.min(clip.right,bounds.right);
      }
      if(/auto|scroll|hidden|clip/.test(style.overflowY)) {
        clip.top=Math.max(clip.top,bounds.top);clip.bottom=Math.min(clip.bottom,bounds.bottom);
      }
    }
    return {rect,clip};
  });
  const {rect,clip}=geometry,unit=1/64;
  assert.ok(rect.left>=clip.left-unit && rect.right<=clip.right+unit && rect.top>=clip.top-unit && rect.bottom<=clip.bottom+unit,
    `focused card is fully painted inside its nested suit/board clips: ${JSON.stringify(geometry)}`);
  return rect;
}
async function capture(page, label) {
  await page.waitForLoadState('networkidle'); await frame(page);
  const file = `${label}.png`;
  await page.screenshot({path: path.join(evidence, file)}); report.screenshots.push(file);
}
async function barGeometry(page, width, height) {
  const controls={},geometryEpsilon=1e-5;
  for(const name of ['Table layout','Game menu','Trade cards','End turn','History','Scores']) {
    // A stationary pointer can reveal a tooltip as the dock moves under it.
    // Flutter then includes that exact tooltip twice in the accessible name.
    const label=name==='Table layout'?/^Table layout/:
      ['Game menu','History','Scores'].includes(name)?new RegExp(`^${name}(?: ${name})?$`):name;
    const bounds=await button(page,label).boundingBox();
    // DOM edge subtraction can report 47.999999px for a painted 48px target.
    assert.ok(bounds&&bounds.width>=48-geometryEpsilon&&bounds.height>=48-geometryEpsilon,
      `${name} retains a complete 48px target: ${JSON.stringify(bounds)}`);
    assert.ok(bounds.x>=-1/60&&bounds.y>=-1/60&&bounds.x+bounds.width<=width+1/60&&bounds.y+bounds.height<=height+1/60,
      `${name} stays fully visible inside the viewport: ${JSON.stringify({width,height,bounds})}`);
    controls[name]=bounds;
  }
  if(width>=600) {
    for(const name of ['Game menu','End turn']) {
      const bounds=controls[name],gap=width-bounds.x-bounds.width;
      assert.ok(gap>=8-1/60&&gap<=56,
        `${name} anchors to the trailing edge instead of clustering at the start: ${JSON.stringify({width,gap,bounds})}`);
    }
    assert.ok(controls.History.x+controls.History.width<controls['Trade cards'].x,
      'roomy action dock separates utility controls from primary actions');
  }
  const table=await publicTable(page).boundingBox(),hand=await handPane(page).boundingBox();
  assert.ok(controls['Game menu'].y+controls['Game menu'].height<=table.y+1/60,'header tools do not overlap the table');
  assert.ok(controls['Table layout'].y+controls['Table layout'].height<=table.y+1/60,'layout picker does not overlap the table');
  for(const name of ['Trade cards','End turn','History','Scores']) {
    assert.ok(controls[name].y>=hand.y+hand.height-1/60,`${name} does not overlay the private hand`);
  }
  return controls;
}
async function responsiveBarChecks(browser, selectPage) {
  const context=await browser.newContext({viewport:{width:1440,height:948},hasTouch:true});
  const page=await context.newPage();selectPage(page);
  const posts=[];
  page.on('request',request=>{if(request.method()==='POST') posts.push(new URL(request.url()).pathname);});
  page.on('pageerror',error=>report.errors.push({type:'pageerror',message:error.message}));
  page.on('console',message=>{if(isBrowserError(message.type(),message.text())) report.errors.push({type:'console',message:message.text()});});
  page.on('requestfailed',request=>report.errors.push({type:'network',path:new URL(request.url()).pathname,failure:request.failure()}));
  await page.goto(`${origin}/`);await button(page,'Game menu').waitFor({timeout:30000});
  await page.waitForLoadState('networkidle');await frame(page);
  const response=await context.request.get(`${origin}/main.dart.js`);
  assert.equal(response.status(),200);report.assets['main.dart.js']=hash(await response.body());
  assert.equal(report.assets['main.dart.js'],hash(fs.readFileSync(path.resolve(__dirname,'../build/web/main.dart.js'))));
  const timeOrigin=await page.evaluate(()=>performance.timeOrigin),measurements=[];
  report.barMeasurements=measurements;
  // Start at the reported roomy layout so the old intrinsic cluster fails
  // before viewport changes or focus revelation can alter its initial paint.
  for(const width of [1440,1050,1049,821,600,599,560,559,500,390,320,319,320,390,500,559,560,599,600,821,1049,1050,1440]) {
    await page.setViewportSize({width,height:948});await frame(page);
    const controls=await barGeometry(page,width,948);
    const status=page.locator('[flt-semantics-identifier="header-status"]');
    const statusBox=await status.boundingBox(),toolsBox=await page.locator('[flt-semantics-identifier="header-tools"]').boundingBox();
    assert.ok(Math.abs(statusBox.y+statusBox.height/2-toolsBox.y-toolsBox.height/2)<=1/60,
      `normal-size header status and tools share one centered row at${width}: ${JSON.stringify({statusBox,toolsBox})}`);
    assert.ok(statusBox.height<=28,'default round and turn stay on one readable line, including319px');
    assert.match((await status.innerText()).replace(/\s+/g,' '),/Round 7 \/ 13 Your turn/,'compact row retains full round and turn labels');
    assert.ok(statusBox.x+statusBox.width<=toolsBox.x+1/60,'status and tools never overlap');
    measurements.push({width,height:948,controls,header:{status:statusBox,tools:toolsBox}});
    assert.equal(await page.evaluate(()=>performance.timeOrigin),timeOrigin,'responsive bars do not reload the game');
  }
  const phone=measurements.find(m=>m.width===390),desktop=measurements[0];
  assert.ok(desktop.controls['End turn'].height>phone.controls['End turn'].height,
    'roomy action controls grow beyond phone sizing');
  assert.ok(desktop.controls['Game menu'].height>phone.controls['Game menu'].height,
    'roomy header controls grow beyond phone sizing');
  for(let i=Math.floor(measurements.length/2)+1;i<measurements.length;i++) for(const name of Object.keys(measurements[i].controls)) {
    const original=measurements.find(m=>m.width===measurements[i].width).controls[name];
    for(const axis of ['x','y','width','height']) assert.ok(Math.abs(original[axis]-measurements[i].controls[name][axis])<1/60,
      `reverse resize restores ${name} ${axis}`);
  }
  const history=button(page,'History');await keyboardFocus(page,history);
  const focusedId=await history.getAttribute('id');
  for(const width of [599,600,1049,1050,1440]) {
    await page.setViewportSize({width,height:948});await frame(page);
    assert.equal(await history.getAttribute('id'),focusedId,'bar control retains its semantic identity');
    assert.ok(await history.evaluate(e=>e===document.activeElement),'bar keyboard focus survives live reflow');
  }
  await page.keyboard.press('Enter');await page.getByText('Game history',{exact:true}).waitFor();
  await button(page,'Close').waitFor();
  await page.keyboard.press('Escape');await button(page,'Close').waitFor({state:'hidden'});
  assert.deepEqual(posts,[],'bar layout and history never submit commands');
  await capture(page,'responsive-bars-desktop');
  for(const width of [821,599,560,500,390,319]) {
    await page.setViewportSize({width,height:948});await frame(page);
    await capture(page,`responsive-bars-${width}`);
  }
  report.cases.push({case:'responsive-bars-distribution-sizing-focus-and-reverse-resize',measurements,commandPosts:posts.length,pass:true});
  await context.close();
}
async function initiallyVisibleHand(page) {
  // Check the initial paint before any scroll, focus reveal or helper can hide
  // the regression where public piles pushed the entire hand below the page.
  const viewport=page.viewportSize(),heading=await handHeading(page).boundingBox();
  const first=page.getByRole('button',{name:/^(?:[2-9]|10|[AJQK]) of (?:Hearts|Diamonds|Clubs|Spades), copy/}).first();
  const card=await first.boundingBox();
  assert.ok(heading&&heading.y>=0&&heading.y+heading.height<=viewport.height,
    `the hand heading is visible on initial paint: ${JSON.stringify({viewport,heading})}`);
  assert.ok(card&&card.y>=0&&Math.min(card.y+card.height,viewport.height)-card.y>=48,
    `a usable hand card is visible without scrolling the page: ${JSON.stringify({viewport,card})}`);
  report.cases.push({case:'hand-visible-on-initial-paint-without-page-scroll',...viewport,heading,card,pass:true});
}
async function composition(page, width, height) {
  await page.setViewportSize({width, height}); await frame(page);
  const close=button(page,'Close actions');
  if(await close.count()) {
    await close.click();
    // Flutter can defer WebKit accessibility activation past two browser frames.
    // Await the single activation's outcome; never click again to force a pass.
    await close.waitFor({state:'hidden',timeout:3000});
    await publicHeading(page).waitFor({timeout:3000});
    await frame(page);
  }
  assert.equal(await publicHeading(page).count(), 1, 'public overview remains in the shared board');
  assert.equal(await actionHeading(page).count(), 1, 'action controls remain immediately available');
  assert.equal(await handHeading(page).count(), 1, 'private hand remains in the shared board');
  // Responsive text animates while resizing. Read related geometry in one
  // browser task so the table and its children describe the same painted frame.
  const {table,hand,board,cells}=await page.getByRole('button',{name:/^View (?:Alice|Bob|Carol|Deniz) public cards/}).evaluateAll(elements=>{
    const box=id=>document.querySelector(`[flt-semantics-identifier="${id}"]`)?.getBoundingClientRect().toJSON();
    return {table:box('public-square'),hand:box('private-hand-viewport'),board:box('board-viewport'),cells:elements.map(e=>e.getBoundingClientRect().toJSON())};
  });
  assert.equal(cells.length,4,'all four public quadrants remain visible at every width');
  const unit=1/64;
  // Firefox rounds row position and child height independently. At768px the
  // bottom delta .019669px is .011353 position + .008316 height; both satisfy
  // the retained1/64 component checks below. Containment must allow their sum.
  // Raw CSS still declares equal237.429px halves of a474.857px table.
  const cellUnit=engine==='firefox'?2*unit+1/16384:unit;
  report.cellContainmentToleranceCssPx=cellUnit;
  assert.ok(table&&hand&&board&&Math.abs(table.x-hand.x)<=unit&&Math.abs(table.x+table.width-hand.x-hand.width)<=unit&&Math.abs(table.x-board.x)<=unit&&Math.abs(table.x+table.width-board.x-board.width)<=unit,
    `public table spans the full board width and aligns with both hand edges: ${JSON.stringify({table,hand,board})}`);
  for(const cell of cells) {
    assert.ok(Math.abs(cell.width-table.width/2)<=unit&&Math.abs(cell.height-table.height/2)<=unit,`four equal public quadrants each fill one quarter of the table: ${JSON.stringify({table,cell})}`);
    assert.ok(cell.width>=48&&cell.height>=48,'whole public quadrant is an accessible touch target');
    assert.ok(cell.left>=table.x-cellUnit&&cell.right<=table.x+table.width+cellUnit&&cell.top>=table.y-cellUnit&&cell.bottom<=table.y+table.height+cellUnit,
      `each public quadrant stays inside the public table: ${JSON.stringify({cell,table,cellUnit})}`);
    assert.ok(cell.left>=-unit&&cell.right<=width+unit&&cell.top>=-unit&&cell.bottom<=height+unit,'each public quadrant is painted inside the viewport');
  }
  assert.equal(new Set(cells.map(c=>Math.round(c.x))).size,2,'public seats never collapse to one column');
  assert.equal(new Set(cells.map(c=>Math.round(c.y))).size,2,'public seats form exactly two equal rows');
  const top=cells.filter(cell=>Math.abs(cell.y-table.y)<=unit),bottom=cells.filter(cell=>Math.abs(cell.y-table.y-table.height/2)<=unit);
  assert.ok(top.length===2&&bottom.length===2&&Math.abs(top[0].x-top[1].x)>=table.width/2-unit&&Math.abs(bottom[0].x-bottom[1].x)>=table.width/2-unit,
    'public quadrants fill both rows without gaps or overlap');
  assert.ok(hand.y>=table.y+table.height-unit&&hand.y-(table.y+table.height)<=24,'hand sits immediately below the public table');
  assert.ok(hand.y>=0&&hand.y+hand.height<=height+unit&&hand.height>=48,'private hand retains a visible viewport');
  for (const name of ['Table','Hand','Decisions']) assert.equal(await button(page,name).count(),0, 'no replacement workspace tabs or rail');
  assert.match(await button(page, /^Table layout/).innerText(), /Auto/);
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth > innerWidth);
  assert.equal(overflow, false, 'no horizontal document overflow');
  await barGeometry(page,width,height);
  report.cases.push({case:'full-width-equal-public-quadrants-with-adjacent-visible-hand',width,height,table,hand,cells,pass:true});
}
async function publicGrowthChecks(browser, selectPage) {
  const context=await browser.newContext({viewport:{width:319,height:948},hasTouch:true});
  const page=await context.newPage();selectPage(page);
  const posts=[];
  page.on('request',request=>{if(request.method()==='POST') posts.push(new URL(request.url()).pathname);});
  page.on('pageerror',error=>report.errors.push({type:'pageerror',message:error.message}));
  page.on('console',message=>{if(isBrowserError(message.type(),message.text())) report.errors.push({type:'console',message:message.text()});});
  page.on('requestfailed',request=>report.errors.push({type:'network',path:new URL(request.url()).pathname,failure:request.failure()}));
  await page.goto(`${origin}/`);await button(page,'Game menu').waitFor({timeout:30000});
  await page.waitForLoadState('networkidle');await frame(page);
  const response=await context.request.get(`${origin}/main.dart.js`);
  assert.equal(response.status(),200);
  report.assets['main.dart.js']=hash(await response.body());
  assert.equal(report.assets['main.dart.js'],hash(fs.readFileSync(path.resolve(__dirname,'../build/web/main.dart.js'))));
  const url=page.url(),timeOrigin=await page.evaluate(()=>performance.timeOrigin);
  let navigations=0;page.on('framenavigated',f=>{if(f===page.mainFrame()) navigations++;});
  const identities=()=>page.getByRole('button',{name:/^(?:[2-9]|10|[AJQK]) of (?:Hearts|Diamonds|Clubs|Spades), copy/}).allTextContents();
  const originalHand=await identities();assert.equal(originalHand.length,18);
  const measurements=[];report.publicGrowthMeasurements=measurements;
  for(const width of [319,390,821,1440,821,390,319]) {
    await composition(page,width,948);await initiallyVisibleHand(page);
    const table=await publicTable(page).boundingBox();
    const cells=await page.getByRole('button',{name:/^View (?:Alice|Bob|Carol|Deniz) public cards/}).evaluateAll(elements=>elements.map(e=>e.getBoundingClientRect().toJSON()));
    const hand=await handPane(page).boundingBox();
    const footer=await Promise.all(['Trade cards','End turn'].map(async name=>{
      const control=page.getByRole('button',{name:new RegExp(`^${name}$`,'i')});
      const bounds=await control.boundingBox();
      assert.ok(bounds&&bounds.width>=48&&bounds.height>=48&&bounds.y>=0&&bounds.y+bounds.height<=948+1/64,`footer remains painted and usable at${width}: ${JSON.stringify(bounds)}`);
      return bounds;
    }));
    measurements.push({width,height:948,table,cells,hand,footer});
    assert.deepEqual(await identities(),originalHand,'live width changes preserve every authorized physical hand card');
    assert.equal(page.url(),url);assert.equal(await page.evaluate(()=>performance.timeOrigin),timeOrigin,'resize preserves the document without refresh');
    await capture(page,`public-growth-${measurements.length}-${width}x948`);
  }
  for(let index=1;index<4;index++) {
    const before=measurements[index-1],after=measurements[index];
    assert.ok(after.table.width>before.table.width+1,`public table must grow with available width at fixed height: ${JSON.stringify({before,after})}`);
    assert.ok(after.cells[0].width>before.cells[0].width+1,'public card preview composition grows with its equal seat');
  }
  for(let index=4;index<7;index++) for(const surface of ['table','hand']) for(const dimension of ['x','y','width','height']) {
    assert.ok(Math.abs(measurements[index][surface][dimension]-measurements[6-index][surface][dimension])<1,
      `shrinking restores the same ${surface} ${dimension}`);
  }
  assert.equal(navigations,0,'live resizing never navigates or reloads');assert.deepEqual(posts,[]);
  report.cases.push({case:'live-full-width-public-table-and-preview-composition-grow-at-fixed-height',measurements,navigations,commandPosts:posts.length,pass:true,
    artworkMeasurement:'Browser DOM measures table and seat composition; transformed widget paint bounds verify decorative card scaling, with these captures for visual browser review.'});
  await context.close();
}
async function runChecks() {
  fs.mkdirSync(evidence, {recursive: true});
  const browser = await engines[engine].launch({headless: true}); report.version = browser.version();
  let activePage;
  try {
    // Trusted key events expose microtask checkpoints between native listeners;
    // a synchronous VM dispatch alone cannot establish this bridge contract.
    const bridgeContext = await browser.newContext();
    const bridgePage = await bridgeContext.newPage();
    // Native inputs receive pointer focus in all three engines; Safari's button
    // click-focus preference is independent of the keyboard bridge contract.
    await bridgePage.setContent('<input id="outside" aria-label="Outside"><flutter-view><input id="start" aria-label="Start"><input id="next" aria-label="Next"></flutter-view><input id="after" aria-label="After">');
    await bridgePage.addScriptTag({path: path.resolve(__dirname,'../web/keyboard_bridge.js')});
    await bridgePage.evaluate(() => {
      window.bridgeHandled = true;
      window.addEventListener('keydown', event => {
        if (event.key === 'Tab') window.cgmsCompleteTabTraversal(window.bridgeHandled);
      }, true);
    });
    for (const key of ['Tab','Shift+Tab']) {
      await bridgePage.locator('#start').click(); await bridgePage.keyboard.press(key);
      assert.equal(await bridgePage.evaluate(() => document.activeElement.id),'start');
    }
    await bridgePage.locator('#outside').click(); await bridgePage.keyboard.press('Tab');
    assert.equal(await bridgePage.evaluate(() => document.activeElement.id),'start','outside controls retain native traversal');
    await bridgePage.evaluate(() => {window.bridgeHandled = false;});
    await bridgePage.locator('#start').click(); await bridgePage.keyboard.press('Shift+Tab');
    assert.equal(await bridgePage.evaluate(() => document.activeElement.id),'outside','unhandled reverse traversal exits the view');
    await bridgePage.locator('#next').click(); await bridgePage.keyboard.press('Tab');
    assert.equal(await bridgePage.evaluate(() => document.activeElement.id),'after','unhandled forward traversal exits the view');
    report.cases.push({case:'trusted-native-tab-bridge-handled-reverse-outside-and-exit',pass:true});
    await bridgeContext.close();
    if (process.env.CGMS_BRIDGE_ONLY) {
      report.status = 'passed';
      console.log(JSON.stringify({status: report.status, engine, cases: report.cases.length}));
      return;
    }
    await responsiveBarChecks(browser,page=>{activePage=page;});
    if(process.env.CGMS_BARS_ONLY) {
      assert.deepEqual(report.errors,[]);report.status='passed';
      console.log(JSON.stringify({status:report.status,engine,cases:report.cases.length}));return;
    }
    await publicGrowthChecks(browser,page=>{activePage=page;});
    if(process.env.CGMS_PUBLIC_GROWTH_ONLY) {
      assert.deepEqual(report.errors,[]);report.status='passed';
      console.log(JSON.stringify({status:report.status,engine,cases:report.cases.length}));return;
    }
    for(const viewport of [{width:319,height:948},{width:320,height:568}]) {
      const context=await browser.newContext({viewport,hasTouch:true});
      const page=await context.newPage();activePage=page;
      page.on('pageerror',error=>report.errors.push({type:'pageerror',message:error.message}));
      page.on('console',message=>{if(isBrowserError(message.type(),message.text())) report.errors.push({type:'console',message:message.text()});});
      page.on('requestfailed',request=>report.errors.push({type:'network',path:new URL(request.url()).pathname,failure:request.failure()}));
      await page.goto(`${origin}/`);await button(page,'Game menu').waitFor({timeout:30000});
      const response=await context.request.get(`${origin}/main.dart.js`);
      report.assets['main.dart.js']=hash(await response.body());
      assert.equal(report.assets['main.dart.js'],hash(fs.readFileSync(path.resolve(__dirname,'../build/web/main.dart.js'))));
      await initiallyVisibleHand(page);
      await composition(page,viewport.width,viewport.height);
      await capture(page,`fresh-${viewport.width}-${viewport.height}-visible-hand`);
      const beforeActions=await handPane(page).boundingBox();
      await actionHeading(page).tap();await button(page,'Close actions').waitFor();
      assert.deepEqual(await handPane(page).boundingBox(),beforeActions,'opening actions preserves the visible hand rectangle');
      await initiallyVisibleHand(page);
      await button(page,'Close actions').tap();await button(page,'Close actions').waitFor({state:'hidden'});
      assert.deepEqual(await handPane(page).boundingBox(),beforeActions,'closing actions preserves the hand rectangle');
      await page.getByRole('button',{name:/^View Bob public cards/}).tap();
      await button(page,'Close public cards').waitFor();
      const visibleCards=await page.getByRole('button',{name:/^(?:[2-9]|10|[AJQK]) of (?:Hearts|Diamonds|Clubs|Spades), copy/}).allTextContents();
      assert.deepEqual(visibleCards,['9 of Diamonds, copy 1','7 of Diamonds, copy 1','8 of Hearts, copy 1','6 of Hearts, copy 1'],
        'one compact public quadrant opens all and only its authorized physical cards');
      await button(page,'Close public cards').tap();await button(page,'Close public cards').waitFor({state:'hidden'});
      report.cases.push({case:'narrow-public-chooser-and-actions-preserve-visible-hand',...viewport,publicCardCount:4,pass:true});
      await context.close();
    }
    // Both phone-first and desktop-first fresh contexts are acceptance paths.
    for (const initialWidth of [390, 1440]) {
      const context = await browser.newContext({viewport: {width: initialWidth, height: 900}, hasTouch: true});
      const page = await context.newPage(); activePage = page;
      const posts=[];
      page.on('request',request=>{if(request.method()==='POST') posts.push(new URL(request.url()).pathname);});
      page.on('pageerror', error => report.errors.push({type: 'pageerror', message: error.message}));
      page.on('console', message => {if (isBrowserError(message.type(), message.text())) report.errors.push({type: 'console', message: message.text()});});
      page.on('requestfailed', request => report.errors.push({type: 'network', path: new URL(request.url()).pathname, failure: request.failure()}));
      const start = Date.now(); await page.goto(`${origin}/`);
      await button(page, 'Game menu').waitFor({timeout: 30000});
      await composition(page,initialWidth,900);
      report.cases.push({case: 'fresh-normal-entry', initialWidth, usableMs: Date.now() - start, pass: true});
      await capture(page,`fresh-${initialWidth}-reference-board`);
      for (const file of ['index.html', 'keyboard_bridge.js', 'flutter_bootstrap.js', 'main.dart.js']) {
        const response = await context.request.get(`${origin}/${file}`);
        const local = hash(fs.readFileSync(path.resolve(__dirname, '../build/web', file)));
        assert.equal(response.status(), 200); assert.equal(hash(await response.body()), local, `served ${file} matches current build`);
        report.assets[file] = local;
      }
      const sizesByClass=[];
      for(const [width,height] of [[390,900],[768,900],[1440,900],[1440,1200]]) {
        await composition(page,width,height); await nav(page,'Hand');
        const privateCards=page.getByRole('button',{name:/^(?:[2-9]|10|[AJQK]) of (?:Hearts|Diamonds|Clubs|Spades), copy/});
        const boxes=await privateCards.evaluateAll(elements=>elements.map(e=>({
          ...e.getBoundingClientRect().toJSON(),label:e.getAttribute('aria-label')||e.textContent,
        })));
        assert.equal(boxes.length,18,'the complete reference hand remains individually accessible');
        // Flutter clips semantic rectangles at scroll edges. Reveal each exact
        // physical card with real keyboard traversal before measuring its frame;
        // a clipped semantic sliver is not the card's visual size.
        const measured=[];
        for(const box of boxes) {
          const target=page.getByRole('button',{name:box.label,exact:true});
          await keyboardFocus(page,target); await frame(page);
          const rect=await assertRevealedCard(target); measured.push({...rect,label:box.label});
          const first=measured[0];
          assert.ok(Math.abs(rect.width-first.width)<1 && Math.abs(rect.height-first.height)<1,
            `fully revealed hand cards have equal footprints at ${width}: ${JSON.stringify(measured)}`);
        }
        const first=measured[0];
        for(const suit of ['Hearts','Spades','Diamonds','Clubs']) {
          const group=boxes.filter(box=>box.label.includes(` of ${suit},`));
          for(let index=1;index<group.length;index++) {
            const previous=group[index-1],current=group[index];
            assert.ok(Math.abs(previous.y-current.y)<1,`${suit} stays in one horizontal stack`);
            const step=Math.abs(current.x-previous.x);
            // Compare complete semantic frames only; edge-clipped ones were
            // individually measured and exercised by the traversal above.
            if(Math.abs(previous.width-first.width)<1 && Math.abs(current.width-first.width)<1) {
              assert.ok(step>=47.999 && step<first.width,`${suit} overlaps while exposing a 48px target: ${step}/${first.width}`);
            }
          }
        }
        sizesByClass.push({width,height,cardWidth:first.width,cardHeight:first.height});
      }
      // At equal height, desktop cards may reach the same height budget as
      // tablet cards. More available width AND height must enlarge them.
      assert.ok(sizesByClass[1].cardWidth>sizesByClass[0].cardWidth && sizesByClass[2].cardWidth>=sizesByClass[1].cardWidth && sizesByClass[3].cardWidth>sizesByClass[2].cardWidth,
        `cards enlarge within both viewport budgets: ${JSON.stringify(sizesByClass)}`);
      report.cases.push({case:'uniform-hand-cards-in-horizontal-suit-stacks',initialWidth,measurements:sizesByClass,pass:true});
      await composition(page,initialWidth,900);await nav(page, 'Hand');
      // The normal approved scenario admits Hearts as an opening selection;
      // a Queen is deliberately ineligible here and must not be forced selected.
      const selected = page.getByRole('button', {name: /^2 of Hearts, copy 1/}).first();
      await keyboardFocus(page,selected); await frame(page);
      await selected.tap({position:{x:24,y:32}});
      await page.waitForFunction(() => [...document.querySelectorAll('[role="button"]')].some(e => /^2 of Hearts, copy 1/.test(e.getAttribute('aria-label') || e.textContent) && e.getAttribute('aria-current') === 'true'));
      // A height-capped desktop hand can fit the entire suit without overflow.
      // Exercise wheel revelation at a real overflowing phone width in both
      // phone-first and desktop-first contexts, retaining the same selection.
      await composition(page,390,900);
      await keyboardFocus(page,selected);
      // Selecting an eligible opening card reveals the action region. Return
      // the painted hand strip to view before exercising horizontal wheel input.
      await frame(page); await expose(page,selected); await frame(page);
      const beforeScroll=await selected.boundingBox();
      await page.mouse.move(beforeScroll.x+24,beforeScroll.y+32);
      await page.mouse.wheel(240,0); await frame(page);
      const afterScroll=await selected.boundingBox();
      assert.ok(afterScroll.x<beforeScroll.x-1,`wheel scroll reveals later cards in the horizontal suit stack: ${JSON.stringify({beforeScroll,afterScroll})}`);
      assert.equal(await selected.getAttribute('aria-current'),'true','horizontal scrolling preserves physical selection');
      report.cases.push({case:'horizontal-suit-scroll-preserves-selection',initialWidth,width:390,pass:true});
      const sizes = initialWidth === 390
        ? [[319,948],[320,568],[390,844],[360,800],[599,900],[600,900],[599,900],[600,900],[1049,900],[1050,900],[1049,900],[1050,900],[1440,900],[1280,720],[1280,600],[390,844],[844,390],[390,844],[768,1024],[1024,768]]
        : [[1440,900],[390,844],[1440,900]];
      for (const [width,height] of sizes) {
        await composition(page,width,height);
        assert.equal(await selected.getAttribute('aria-current'), 'true', 'physical card selection survives every transition');
      }
      for (const [width,height] of [[390,844],[768,1024],[1440,900]]) {
        await composition(page,width,height); await capture(page,`${initialWidth}-${width}-hand`);
        await nav(page, 'Table');
        await publicHeading(page).waitFor(); await capture(page,`${initialWidth}-${width}-table`);
        await nav(page,'Decisions'); await actionHeading(page).waitFor();
        await capture(page,`${initialWidth}-${width}-decisions`);
      }
      await composition(page,390,844);
      await keyboardFocus(page, selected); await page.keyboard.press('i');
      await button(page, 'Close inspection').waitFor(); await page.keyboard.press('Escape');
      await button(page, 'Close inspection').waitFor({state: 'hidden'});
      report.cases.push({case: 'normal-entry-keyboard-inspect-and-escape', initialWidth, pass: true});
      // First/middle/last eligible number cards exercise the exposed part of a
      // suit stack. Real keyboard traversal must reveal cards outside its
      // horizontal viewport before touch or inspection can succeed.
      for(const rank of ['2','5','10']) {
        const target=page.getByRole('button',{name:new RegExp(`^${rank} of Hearts, copy 1`)});
        await keyboardFocus(page,target); await frame(page);
        const rect=await assertRevealedCard(target);
        assert.ok(rect.x>=0 && rect.x+48<=page.viewportSize().width,'keyboard reveals the exposed hand-card target');
        if(await target.getAttribute('aria-current')!=='true') {
          await target.tap({position:{x:24,y:32}});
          await page.waitForFunction(label=>[...document.querySelectorAll('[role="button"]')].some(e=>(e.getAttribute('aria-label')||e.textContent).startsWith(label)&&e.getAttribute('aria-current')==='true'),`${rank} of Hearts, copy 1`);
        }
        await keyboardFocus(page,target); await page.keyboard.press('i');
        await button(page,'Close inspection').waitFor(); await page.keyboard.press('Escape');
        await button(page,'Close inspection').waitFor({state:'hidden'});
        assert.equal(await target.getAttribute('aria-current'),'true','inspection preserves the exact physical selection');
        report.cases.push({case:'suit-stack-physical-card-touch-and-keyboard',rank,initialWidth,pass:true});
      }
      await expose(page,button(page,'Game menu')); await button(page, 'Game menu').tap();
      await page.getByRole('menuitem', {name: 'Display settings', exact: true}).tap();
      const textSize = button(page, /^Text size/); await keyboardFocus(page, textSize);
      await page.keyboard.press('Space');
      await page.getByRole('menuitem', {name: '200%', exact: true}).click();
      await keyboardFocus(page, button(page, 'Done')); await page.keyboard.press('Enter');
      await button(page, 'Done').waitFor({state: 'hidden'});
      for (const [width,height] of [[390,844],[768,1024],[1440,900],[844,390]]) {
        await composition(page,width,height);
        await capture(page,`${initialWidth}-${width}-text200`);
        report.cases.push({case:'normal-entry-text200-touch-navigation',width,height,pass:true});
      }
      if (initialWidth === 390) {
        await composition(page,844,390);
        const pane=await handPane(page).boundingBox();
        await page.mouse.move(pane.x+pane.width/2,pane.y+pane.height/2);await page.mouse.wheel(0,-100000);await frame(page);
        // The hand stays on screen while its own content scrolls. Real forward
        // traversal must reveal the last card inside this bounded viewport.
        const lastCard = page.getByRole('button', {name:/^(?:[2-9]|10|[AJQK]) of (?:Hearts|Diamonds|Clubs|Spades), copy/}).last();
        const before = await lastCard.boundingBox();
        assert.ok(pane.y>=0&&pane.y+pane.height<=390+1/64,'hand pane remains on screen before traversal');
        await keyboardFocus(page,lastCard); await frame(page);
        const after = await lastCard.boundingBox();
        const visibleHeight=Math.min(after.y+after.height,pane.y+pane.height)-Math.max(after.y,pane.y);
        assert.ok(visibleHeight>=48, 'keyboard exposes at least a complete touch-target height of the distant card');
        await capture(page,'text200-short-offscreen-keyboard');
        await page.keyboard.press('i');
        const close = button(page,'Close inspection'); await close.waitFor();
        await page.keyboard.press('Shift+Tab'); await frame(page);
        assert.ok(await close.evaluate(element => {
          const dialog = element.closest('[role="dialog"],[role="alertdialog"]');
          return dialog ? dialog.contains(document.activeElement) : element === document.activeElement;
        }), 'reverse Tab remains inside the card inspector');
        await keyboardFocus(page,close);
        await page.keyboard.press('Escape'); await close.waitFor({state:'hidden'});
        report.cases.push({case:'offscreen-card-after-pointer-short-text200-and-modal-reverse-tab',width:844,height:390,before,after,pass:true});
      }
      assert.deepEqual(posts,[],'viewing, scrolling, selecting and inspecting never sends commands');
      report.cases.push({case:'hand-interactions-without-command-posts',initialWidth,commandPosts:0,pass:true});
      await context.close();
    }
    assert.deepEqual(report.errors, []); report.status = 'passed';
  } catch (error) {
    report.status = 'failed'; report.failure = error.message;
    if (activePage && !activePage.isClosed()) {
      await activePage.screenshot({path: path.join(evidence, 'failure.png')});
      fs.writeFileSync(path.join(evidence, 'failure-semantics.txt'), await activePage.locator('flt-semantics-host').ariaSnapshot());
    }
    throw error;
  } finally {
    fs.writeFileSync(path.join(evidence, 'report.json'), JSON.stringify(report, null, 2));
    await browser.close();
  }
  console.log(JSON.stringify({status: report.status, cases: report.cases.length, evidence, errors: report.errors}));
}
runChecks().catch(error => { console.error(error); process.exitCode = 1; });
