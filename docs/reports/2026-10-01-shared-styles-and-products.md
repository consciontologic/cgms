# Shared match styles and revised products — 2026-10-01

> Archive removal update (2026-10-02): the owner requested removal of every
> file and subdirectory in `client/screenshots/` except `cgmsart-phone.png`,
> including raw reports, manifests, logs and copied harnesses. This document
> remains a historical account of the recorded results; its capture-time counts
> do not imply that those raw files remain available. Removed archive links are
> plain text below. See the [retention report](2026-10-02-screenshot-retention.md)
> and its per-file inventory for the disposition.

Run ID: `styles-products-20261001`. Base: `6248bb5`.

## Changes and authority

Decisions 0015–0016 remove cosmetic sales and currency unlocks, limit premium
products to Ad-free/Daily/Weekly/Monthly/Yearly, and make only earned Daily/Weekly
access currency-redeemable. The owner explicitly confirmed that earned access
retains ads and that styles last an entire match. No new price, renewal rule,
paid Daily benefit or standalone Ad-free duration was invented.

The public wire projection assigns distinct styles from the four existing
editions using a domain-separated hash of immutable public match ID and edition.
This affects neither private chance nor stored engine/checkpoint/replay data.
New matches recompute assignments; a permutation may coincidentally repeat.
Flutter resolves artwork from authoritative current control. Captured Diamonds
immediately adopt the capturer's style without changing physical identity.
All 208 native illustrations are exact copies of their provenance-pinned sources;
concealed backs, live identity labels and original artwork remain unchanged.

Inspectors resolve their identity against the latest authorized projection and
close when that face is no longer visible or the game changes. No command is
submitted by inspecting, changing styles or closing an inspector.

Earned products separate unlimited access from ad removal. New purchases accept
one/seven days; immutable historical two–six-day receipts remain recoverable.
An additive, checksum-pinned migration snapshots promised legacy ad-free expiry
once, never extending it from new earned access, and prevents old binaries from
reopening the revised policy. New cosmetic grants are rejected while historical
restore/revocation and uncertain operation safeguards remain.

The active docs and roadmap reflect the revised scope. GAME_RULES.md and both
client/offline tools are unchanged. P06 remains open for unresolved paid-product
terms; native/store/production and bot work are excluded.

## Verification

The final release asset is `eb6c045add5b6a04d247669f5fe0a474c570fa8163b524600392d9502b81520a`.
Intermediate failures and the earlier economy precheck use separate directories
and are not counted as final successful evidence.

- Host Flutter suite: 187 tests, analyzer and release web build passed together in `styles-final-build-tests--20261001T111621Z-2674921` after the final inactive-art repair.
- Shared package: 58 tests passed in `card-styles-package-final--20261001T110032Z-2634316`; package analysis passed in `card-styles-package-analysis--20261001T105547Z-2617220`.
- Exact inventory/source checks: all 208 illustrations, shared back, fonts and
  package import boundary passed in `styles-assets-final--20261001T110345Z-2644720`.
- Browser helper syntax and unit tests: 9 passed after the final helper correction in `styles-final-browser-helpers--20261001T114153Z-2742330`.
- Reconnect fault-classification tests: 8 passed after adding explicit style-stability assertions in `styles-reconnect-harness-unit--20261001T112248Z-2692401`.
- Existing OpenAPI checks: 2 passed in `shared-style-openapi-review--20261001T110420Z-2646177`.

Full logs live under `/tmp/agent-runs/` with `.log` and `.exit` suffixes. Browser
artifacts were archived under styles-economy-20261001, which is now removed; the
before manifest
pins the original served build and three representative viewports. The
final manifest
verifies 15 successful reports, 220 acceptance cases and 207 final screenshots,
plus ten separate focused regression passes. It pins the actual application,
helper and screenshot hashes and excludes intermediate artifacts.

Backend final verification:

- Serial real PostgreSQL race suite: economy 2.343s, transport 162.703s,
  matchstore 22.384s in `monetization-pg-serial-diagnostic--20261001T110335Z-2643657`.
- Final paid-kind assertions and load-diagnostic source: economy 1.165s,
  64-stream profile 18.621s in `monetization-final-focused-pg--20261001T110700Z-2652806`.
  The final seconds overlapped Flutter checks; this is not isolated capacity evidence.
- Fixture PostgreSQL race suite: 56 scenario/population/economy combinations and 28
  account checks in `shared-card-style-fixture-postgres--20261001T110833Z-2656952`.
- Wire/game/matchstore/transport/fixture unit suites passed in
  `shared-card-style-backend-full--20261001T105541Z-2616389`; wire pinned assignment
  and private-chance invariance rechecked in `shared-card-style-wire-final--20261001T105705Z-2622035`.

## Rendered application

Chromium 153.0.8010.12, Firefox 155.0 and containerized WebKit 26.6 passed all
nine capture/style cases against the real local Go HTTP/WebSocket authority and
disposable PostgreSQL. Phone 390×844 uses three players; tablet 768×1024 and
desktop 1440×900 use four. Every seated viewer agrees on assignments and capture
restyling, while card identity and privacy remain intact. Refresh and page reload
retain assignments. Starting positions are explicitly synthetic legal fixtures;
attacks and responses use the rendered Flutter controls and real commands.

All 24 earned-access browser cases passed across the same three engines. These
cover Daily/Weekly UI eligibility and cost, earned access retaining ads, private
account views, committed purchase with lost reply and receipt recovery, actual
finalization rewards, allowance waiting and streamed admission after redemption.
Weekly price/insufficient-funds UI is exercised in browsers; successful Weekly
debit/expiry and legacy migration are verified in real PostgreSQL tests.

Full real-authority journey/viewport suites passed: Chromium 55, Firefox 54 and
corrected WebKit 54 cases. They cover actual guest room creation/join/deal/reload
and fixture-assisted opening, purchase, private trade, intact loan, capture,
Compensation, Confinement, Justice, Coup, settlement and shared results. The
viewport matrix includes 360×800, 390×844, 599/600 and 1049/1050 boundaries,
768×1024, 1024×768, 1280×720/600, 1440×900 and 844×390 rotation. Selection
survives resizing, back/forward navigation, RTL and 200% text with artwork off.

All six reconnect cases passed across the three engines. Actual HTTP offline
failures and closure of a transparent real-backend WebSocket proxy are followed
by explicit refresh/reconnect. Match styles, selected physical card, raw quantity
draft, turn and pending Justice ownership remain stable; no POST or automatic
pass occurs. This is a browser interruption test, not an OS network partition.

All 18 finish cases passed across the three engines: 320×800 entry/setup/gallery,
a 181-character mixed-direction draft retained at 200% text without submission,
keyboard inspector opening/Escape dismissal and focus restoration under both
motion preferences, and the settings motion toggle with focus return.

Accessibility checks exercised Flutter DOM semantics, Tab/Shift+Tab traversal,
I/Enter/Space activation, Escape dismissal, focus restoration, selected-card
labels, primary 48-pixel touch targets, RTL, 200% text scaling, long mixed-direction
input and emulated reduced-motion preferences. These are targeted checks;
actual screen-reader operation, browser zoom and a full contrast audit were not
performed in this slice. No broad accessibility conformance claim is made.

The three Chromium after-capture layouts were visually inspected, alongside
phone recovered-pass, tablet account and desktop RTL/200%-text account views.
Review also inspected WebKit desktop capture and final lobby/Decisions views,
and Firefox RTL/200%-text account controls.
Public cards remain readable and account content scrolls within its dialog.
Before capture metadata used `before/`; final reports for the same
attack fixture used `card-styles/<engine>/<engine>-card-styles-report.json`.
The removed capture sequences contained separate before-capture, after-capture
and after-reload images. The Chromium report
recorded the phone, tablet and desktop checks.

## Failure recovery and evidence limits

The first final-build WebKit full journey reached the owner lobby after 49
passing cases, then timed out waiting for the invitation response after its
button click. A focused reproduction failed twice in ten attempts: DOM
autoscroll moved a clipped semantics control during pointer down/up, producing
no click and no invitation POST. Real Tab traversal and settled Flutter frames
before the same pointer click passed ten consecutive focused checks, each with
exactly one HTTP 200 response and no submission from navigation. The correction
is WebKit-only in the test helper; no application or assertion was weakened.
The focused report
records the reproduction and correction. Chromium/Firefox full journeys used
the original helper (`4fcf983…`); the final WebKit helper is `caf25dd…`. Its
Chromium/Firefox path is behaviorally unchanged, and both versions are pinned
in the final manifest rather than claimed to be byte-identical.

An archive permissions error was followed prematurely by a rerun, which was
stopped after overwriting some early screenshots/report data. The original
safe-run failure log and screenshot hashes remain. Mixed
failed/interrupted artifacts are explicitly quarantined and excluded from final
acceptance; the complete corrected run passed 54 cases in the fresh
`journeys/webkit-final/` directory (`styles-journeys-webkit-final-sync--20261001T114117Z-2740172`). This limitation
is recorded in the archive provenance
rather than presenting overwritten artifacts as the original run.

The initial baseline screenshot helper sought a phone/desktop Table button in
the tablet split layout. Its corrected class-specific navigation captured all
three viewports. Stopping the first fixture runner interrupted child schema
cleanup; its disposable runner resources were removed. The final fixture's owned
server child received SIGTERM; graceful shutdown and disposable cleanup finished
with runner exit 0 (`styles-final-fixture--20261001T110844Z-2657865`).

A concurrent broad backend integration run included obsolete economy expectations
while implementation was in flight and an operational-load HTTP503. Another
broad race run failed the 64-stream profile with status-only diagnostics. Those
runs are failed evidence; they do not establish a UI or economy regression, nor
is CPU contention proven causal. The coordinated final run and diagnostic
improvement are reported separately; a later pass does not erase these failures.

Browser semantics, keyboard/touch and rendered checks establish only the web
behavior actually exercised. No native device, native screen reader, live ads,
real payments, signed store sandbox or production acceptance is claimed.

The first rendered style check detected 32 static First Light artwork requests
from the hidden local-demo IndexedStack child beneath the online screen. These
were public synthetic examples, not concealed game data. The strict request
whitelist exposed unnecessary inactive rendering. Synchronous route parsing and
art gating now preserve the mounted local subtree, selections, last route and
artwork preference without loading its hidden deck. Four new regressions failed
before the repair and passed afterward; the 187-test suite and new release build
include that repair. Exactly three visible, public welcome illustrations are
allowed before the online snapshot (First Light Clubs 8, Hearts Q, Diamonds 4);
every other requested face must be authorized by the tested viewer projection,
including after capture/refresh/reload.

Markdown local-link, append-only decision and protected-path checks passed in
`styles-doc-contracts--20261001T111657Z-2676619`. The release compiler emitted
a Cupertino icon-family warning; actual artwork/font integrity
and rendered browser asset checks are recorded separately.
