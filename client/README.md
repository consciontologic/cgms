# CGMS / cgmsart Flutter host

The host includes an authoritative online route at `#/online` and deterministic
local scenarios at `#/play`. Both use the adaptive composition; the normal `/`
entry opens `#/play` with Auto. Online state comes from the same-origin Go
API; the reusable `packages/cgms_ui` uses the selected phone-reference vertical
arrangement at every width, enlarging items on tablet/desktop under the
[UI baseline](../docs/design/UI_BASELINE.md). P04 browser acceptance
passed on the recorded release build. Local scenarios remain presentation examples
and do not prove multiplayer or native acceptance.

Online commands retain their original game, version and operation identity.
An uncertain operation blocks further mutations until its original receipt is
reconciled; explicit retry resends only that original envelope. Browser session
storage retains minimal pending command identifiers and authorized opaque card
handles, never credentials or complete game snapshots. Cookies bootstrap the
session; production requires the same-origin TLS deployment.

## Relocation and historical evidence

On 2026-10-01 the complete host moved from `docs/design/app/` directly into
`client/`, including hidden tracked files and the reusable `packages/cgms_ui`.
`offline_host/` and `offline_probe/` remain independent host tools. No compatibility
symlink or duplicate application exists. Old tracking rows describe executions
at the former path; their results are not relocated-build evidence. The owner
later removed the screenshot archive, including its JSON manifests.
Fresh verification is recorded in the [2026-10-01 report](../docs/reports/2026-10-01-nonproduction-verification.md), separately from those historical results.

## Shared vertical board

The owner selected [cgmsart-phone.png](screenshots/cgmsart-phone.png) on
2026-10-02 as the layout reference for every screen size. The normal entry and
online board keep a full-width table split into four equal public-seat quadrants,
with the private hand immediately adjacent. Ordinary viewports show the hand
without page scrolling; short windows, enlarged text and long hands scroll
deliberately. Online move choices use temporary overlays. Item sizing
uses both available width and height. The table and hand share aligned edges and
a common card-size scale; public previews shrink only to fit actual constraints.
The status bar anchors context and tools at opposite sides. Compact online
utilities retain explicit Pass/End Turn and financial access with full touch targets. The compact header keeps status and tools side by side with smaller
visible type/icons and full touch targets; only the bottom dock reflows its
groups. Enlarged status text may wrap within its area, keeping the hand visible.
All sections remain mounted during resizing;
there are no separate hand pages, split workspaces or navigation rails.

Public quadrant controls open accessible full-card details for exact physical
selection and inspection. The private hand uses equal-sized frames in horizontal
overlapping suit groups (Hearts, Spades, Diamonds, Clubs). Groups share two
columns where space permits, scroll horizontally for long holdings, and reflow
with enlarged text. Each physical card remains selectable and inspectable.
Canonical game rules, authorized projections, drafts, focus and pending command
identities remain above presentation. The source mockups guide other flows and
cgmsart-decks remains the artwork authority. The earlier
[redesign report](../docs/reports/2026-10-01-ui-redesign.md) describes its own
historical build, not the current layout.

## Card interaction model

The 2026-10-04 owner revision removes the authoritative online board's permanent
action area and all reserved composer space. Tap/Space selects physical cards;
double tap/Enter activates their relevant move. Drag the exact selected bundle
or tap those cards and a concrete destination. A temporary contextual choice
resolves only the ambiguity or parameters for that move, followed by explicit
confirmation. The magnifier and I inspect independently. Escape cancels only an
unsubmitted selection; dismissed required decisions stay pending and reopenable.

Public cards expose concrete series/formation/card targets, the hand is the
whole-formation take-back destination, and selected number Spades dropped on the
draw pile initiate an explicit-quantity purchase. Source, target, payment and
modifier roles remain distinct. Actual eligible Queens retain card-driven Coup
access inside temporary contexts. End Turn, explicit Pass, financial records and
match progression remain compact utilities. Suggestions do not authorize commands.

The [51-command matrix and verification report](../docs/reports/2026-10-04-card-interaction-redesign.md)
records complete routing coverage, actual browser journeys, final-build identity
and remaining native/engine limits. Historical `/play` storyboard examples are
presentation demonstrations; real server-authoritative play uses `/online`.

## Authoritative online development

From the repository root, select the [project-local Flutter SDK](#run), then run:

```bash
make services.init
(cd client && flutter build web --no-web-resources-cdn)
make services.up MODE=local
```

Build before starting Compose so the static-assets bind mount already exists
with your user ownership. The client is served at `http://localhost:8080/#/online`.
For later changes, `make web.re MODE=local` rebuilds the client and
`make go.re MODE=local` rebuilds the backend; both retain local data.
`make web.dev` provides the same-origin hot-reload workflow.

### Play against bots

Open `http://localhost:8080/#/online`, then **Continue as guest** or sign in.
Choose three or four players, the number of games and a **Bot difficulty**, then
select **Play against bots**. The lobby fills the remaining seats with bots;
select **Start match** when ready. **Create room** still creates a human lobby.

Bots run on the Go authority and see only their own cards and public information.
They take turns and answer required decisions automatically while human choices
remain yours. Reloading or reconnecting resumes the same durable match.
During setup, **Preparing the next move…** changes automatically to the actual
player's turn or required response; no refresh is needed.
At a round boundary, explicitly choose **Stay in game** or **Leave game**;
a temporary decision popup opens when your choice is required. Dismissing it
leaves the decision pending and available from the compact decision utility.
Beginner, Standard and Advanced reuse the existing policy tiers; these names are
not evidence of measured bot strength. Normal human online allowance and reward
rules apply. Browser bot play requires the backend; it is not offline web play.
Online bots also avoid purchases that gain no cards, preventing an empty-deck
buy-and-reopen loop without restricting human choices.

Use the documented `localhost` address for this local stack. `127.0.0.1` is a
different browser origin and is rejected by the configured origin check.

`node client/tool/bot_browser_acceptance.cjs`, run from the repository root with
ordinary services ready, exercises real guest bot matches without fixtures.
Set `CGMS_BROWSER=chromium`, `firefox` or `webkit` using the installed runtime;
run engines serially to respect the shared local request limit. Captures remain
under ignored `.browser-tools/artifacts/`; retain only the compact verification
summary after reviewing them.

For isolated browser acceptance,
build local web assets then run `python3 xops/postgres_integration.py --webfixture`
from the root. This QA-only server uses disposable PostgreSQL and synthetic
rule-boundary positions. Its session bootstrap is not part of the production API.
The browser runner is `tool/browser_acceptance.cjs`; its locally installed
Playwright runtime and reports live under ignored `.browser-tools/`.
Routine runs keep captures in ignored per-run artifact directories. The
[screenshot retention policy](../docs/reports/2026-10-02-screenshot-retention.md) explains how to preserve
compact reports outside this directory without staging disposable captures.

The repeatable browser runner uses Playwright 1.63.0. Install it locally with
`npm install --prefix .browser-tools playwright@1.63.0` from this directory and
install its Chromium/Firefox browsers into `.browser-tools/browsers` using
`PLAYWRIGHT_BROWSERS_PATH=$PWD/.browser-tools/browsers .browser-tools/node_modules/.bin/playwright install chromium firefox`.
The WebKit gate uses the matching official `mcr.microsoft.com/playwright:v1.63.0-noble`
container so it does not install workstation packages.

From the repository root, with the QA authority running:

```bash
CGMS_QUICK=1 CGMS_BROWSER=chromium CGMS_WIDTH=390 CGMS_HEIGHT=844 node client/tool/browser_acceptance.cjs
CGMS_MATRIX_ONLY=1 CGMS_BROWSER=firefox node client/tool/browser_acceptance.cjs
CGMS_FINISH_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_CARD_SIZES_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_CARD_DEMO_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_RTL_HAND_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_KEYBOARD_SEQUENCE_ONLY=1 CGMS_BROWSER=firefox node client/tool/browser_acceptance.cjs
CGMS_BROWSER_ORIGIN=http://127.0.0.1:8080 node client/tool/adaptive_entry_browser_acceptance.cjs
CGMS_QUADRANT_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_DIRECT_GEOMETRY_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_DIRECT_INTERACTIONS_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_TURN_DRAW_ONLY=1 CGMS_BROWSER=chromium node client/tool/browser_acceptance.cjs
CGMS_BROWSER=chromium node client/tool/reconnect_browser_acceptance.cjs
docker run --rm --network host --shm-size=1g -e CGMS_CONTAINER=1 -e CGMS_QUICK=1 -e CGMS_BROWSER=webkit -v "$PWD:/workspace" -w /workspace mcr.microsoft.com/playwright:v1.63.0-noble node client/tool/browser_acceptance.cjs
```

Repeat full journeys at 768×1024 and 1440×900 for every engine; use
`CGMS_PLAYERS=3` for the three-seat variation. Run these serially: concurrent
journeys from the same address intentionally encounter production admission
limits. Restart the disposable authority between engine batches to stay within
its bounded fixture-group count. Reports include the release asset hash and
explicit pass/fail status. These synthetic rule-boundary journeys complement
the normal guest-room/deal/reload journey; they are not complete natural-deal
matches or native-device evidence.

`CGMS_EVIDENCE_DIR` selects a separate output directory for a run. Finish checks
cover 320-pixel entry/gallery, a long mixed-direction draft at 200% text, keyboard
dialog dismissal/focus return and both browser motion preferences. The separate
reconnect runner tests an opening draft and a pending Justice decision at 390×844.
It interrupts browser HTTP and the real backend WebSocket proxy, then uses explicit
Refresh to verify the same physical selection and scope with no command POSTs.
This is transport emulation, not an OS network partition or automatic-reconnect test.

The direct-drop geometry runner covers 1041×948, live 599↔600 and 1049↔1050
transitions, rotation, short windows and RTL at 200% text. The interaction runner
checks pointer release commitment, contextual Confinement and intact loan
proposal/acceptance/response resolution. The scheduled-draw runner checks an Ace
and ordinary Diamond in separate authoritative scenarios, private announcements,
browser reduced motion and no replay on refresh/resize/reconnect. Every runner
verifies the served asset bytes before interacting. Current results are in the
[direct-drop report](../docs/reports/2026-10-04-direct-drop-revision.md).

Historical artwork-only checks use fresh browser contexts for 12 hand/public
cases: 320×800, 390×844, 768×1024, 1440×900, phone RTL and phone 200% text.
They verify single-tap selection and independent double-tap magnification on
exposed hand strips or full public cards, retained drafts, the 48px hollow corner
fallback on full public
chooser cards, no horizontal document overflow and no command POSTs. All cases
exercise keyboard I. Current online gameplay uses double tap for card action
activation; I and the independent inspect control magnify without submitting.
Temporary screenshots record complete paintings and visible-label cleanup;
Flutter canvas artwork itself is not exposed through DOM geometry. Package tests
verify native image-box ratios against all 208 bundled PNG headers and assert
no cropping or stretching. Full-card mode omits visible copy text; compact board
cards retain the selected reference's small physical-copy number. This run does not certify
all browser keyboard traversal or native screen-reader behavior.
`CGMS_CARD_DEMO_ONLY=1` separately captures the default approved phone board,
its hand and its double-tap inspector without authority commands.
`CGMS_RTL_HAND_ONLY=1` runs the existing browser runner against the Coup QA
fixture's four private Queens. It checks each exact physical copy through
right-edge touch selection and keyboard inspection, live width changes and
200% text, with zero command POSTs. This proves two-card suit overlap in RTL;
the normal-entry runner separately exercises dense 18-card horizontal scrolling.
The normal-entry runner opens `/` with Auto in fresh phone and desktop contexts,
checks the exact served build, and exercises live boundaries, rotation, short
viewports, touch navigation and keyboard inspection. The long keyboard sequence
retains the same hand/public selections through normal, 200% text and RTL modes;
it uses forward Tab without manually focusing the target. For automatic startup
and reconnect testing against the retained local Compose stack, use the
[cold-launch workflow](../deploy/README.md#rebuild-restart-and-verification).
The normal-entry checks also exercise distant-card keyboard scrolling in a short
200% viewport, modal reverse-Tab containment and trusted browser Tab events.
`web/keyboard_bridge.js` preserves Flutter's handled traversal result before the
browser can perform a competing move; it keeps unhandled view exit and browser
shortcuts intact. Its event-lifetime contracts run with
`node --test client/tool/keyboard_bridge.test.cjs` from the repository root.

For local economy acceptance, stop the ordinary QA fixture and use
`python3 xops/postgres_integration.py --webfixture-economy` from the repository
root, then `CGMS_BROWSER=chromium node client/tool/economy_browser_acceptance.cjs`.
This disposable authority uses the adopted 10/30 policy and ledger-funded test
accounts; the normal local Compose service keeps economy activation off by default.
Run engines serially to respect the same admission limits as normal requests.
The `ad-free` QA scenario gives seat 1 the approved 30-day trusted-provider term
while preserving the ordinary start cap. `paid-daily` uses the approved 24-hour
term; `paid-weekly`, `paid-monthly` and `paid-yearly` use synthetic 48-hour
remaining-time windows that do not define those products' period arithmetic.
All four paid access scenarios grant unlimited starts and remove ads. The runner
checks paid and earned benefits, private expiry, reload, malformed-account
recovery and keyboard access at 200% text across phone, tablet and desktop; see
the [approved paid-access report](../docs/reports/2026-10-01-approved-paid-access.md).
No store adapter, native purchase or live payment is activated.

All sizes keep the public/action/hand/footer vertical stack. Width selects
phone sizing below 600, tablet 600–1049 and desktop 1050+, independently of height;
it does not switch to a different layout.
Drafts and selected physical cards survive presentation changes. Switching actions
selects into the visible field; selected-action guidance explains declaration and
resolution costs, cancellation and timing. Own exposed Clubs used this turn are
marked from the authorized projection. Authorized owner/member lobbies reload
from public room-ID routes without retaining invitation tokens. Eligible Coup
remains reachable across all surfaces; ordinary actions wait for authority.

## Online requirement evidence

The online adapter renders the authority's projections and costs. It does not
compute a second rules engine or claim that the command chooser is an exhaustive
legal-action menu. The authority validates commands; unavailable, disconnected,
uncertain and stale-game draft states disable submission. New identities can be
entered explicitly; existing public formations, private offers, promises and
debts have contextual selection/actions.

| Requirement | Frontend evidence and boundary |
|---|---|
| F01–F03 | Browser opening/attack response traversal, three-stage Justice decisions, immediate Coup and final ranks; widgets preserve decision IDs and render declaration separately from score. Engine concurrency and legality matrices remain authoritative. |
| F04–F05 | Browser promise payment, all-seat finish, zero-cash next-game entry/reload and final standings; widget regressions distinguish payable debts from receivables and expose settlement promise/debt actions. |
| F06/F10 | Browser private-offer terms are visible only to parties, including semantic-tree checks against an unrelated seat; projections preserve authorized faces and public counts without fabricated opponent finances. |
| F07–F08 | Browser intact Underground loan/acceptance; authority contract widget checks effect owner/target/custodian; formation controls submit substitution/protection fields and discard removed formation references. Hidden binding identities remain server-owned. |
| F09/N01/N04 | Exact rational payment/cost/excess and combat formulas come from authoritative rules context; widget contract checks a changed purchase unit price and physical/modifier totals. Canonical Inflation/Code timing is verified by the shared engine. |
| N02/N03/N07 | Public effect expiry/custody and physical availability are displayed; Confinement cancels the pending hostile action in browsers. Eligible response seats and immediate Coup admission follow authoritative state; engine tests cover skipped/departed seats and restrictions. |
| N05 | Browser lost reply, reload and original-operation receipt recovery; unit tests reject reuse of edited drafts after game/decision replacement and preserve original retry envelopes. |
| N06 | Exact nonspendable match corrections retain origin-game identity separately from cash, game score and cumulative score; composed widget contract covers a one-third correction. Engine ledger tests establish its financial provenance. |
| T01 | Browser private proposal, named acceptance, acceptor-ordered responses, intact loan and non-party privacy; revision-one and exact physical terms have regressions. Offers do not reserve cards. |
| T02 | Browser return/shuffle/draw completion gives one private card to each of two entitled recipients, consumes counters and resumes play; the paired fixture engine test proves ascending queue order between automatic steps. Transient automatic steps are not represented as human prompts. |
| T03 | Browser chosen quantity, explicit physical Spade payment and ordered responses; exact authoritative price, effective payment and excess are rendered. |

The catalog is ARB-backed English; RTL and 200% scaling exercise layout and
semantics, not an unprovided translated-language catalog. Browser semantic-tree
checks do not replace R06N native screen-reader and device acceptance.

## Run

The verified SDK is the project-local Flutter 3.47.5 toolchain under
`.local/toolchains/flutter-3.47.5/flutter` (repository root). Select that toolchain
for these manual commands. From the repository root:

```bash
export PATH="$PWD/.local/toolchains/flutter-3.47.5/flutter/bin:$PATH"
cd client
flutter pub get
flutter run -d web-server --web-hostname=127.0.0.1 --web-port=7357
```

Open [the local demo](http://127.0.0.1:7357). Stop the server with Ctrl+C.
A release build can also be served locally:

```bash
flutter build web --no-web-resources-cdn
python3 -m http.server 7357 --bind 127.0.0.1 --directory build/web
```

Run only one server on the chosen port. The release build bundles its web
resources locally and does not register a service worker: the generated bootstrap
calls `_flutter.loader.load(...)` with an entrypoint callback and no service-worker
settings. The SDK may still
emit `flutter_service_worker.js`; its presence is not evidence of registration or
execution. This is not an installable PWA.

## Explore the approved flows

The default `#/play` opens Alice's 18-card hand and the public positions of Bob,
Carol and Deniz. The scenario chooser offers the same 20 examples as the PNGs.
Setup, card/combo inspection, scores and history remain within the same design.

The connected main sequence demonstrates:

1. Select and open the seven number Hearts; keep them in hand until the explicit
   response sequence resolves, then show 11 cards in hand and Hearts totalling 40.
2. Declare the exact 8 Clubs versus Bob's 8 Hearts attack. The Club becomes used
   on declaration; Bob's Heart remains until resolution. No score is awarded.
3. Offer Q Clubs for Bob's 9 Spades without reserving or moving either card.
   Acceptance starts Bob's action and ordered responses; successful resolution
   updates the hand and does not create an alliance.
4. Choose a purchase quantity and pay with 10 Spades. At a price of 5, buying one
   card leaves excess 5 with no change; payment and draw resolve atomically after
   responses.

Independent examples demonstrate a successful Underground loan and alliance,
Confinement cancelling an out-of-turn Ponzi action, immediate eligible Coup,
board closure followed by settlement, finalized match standings and an unknown
attack outcome that must be reconciled before another submission. These are
separate fixtures, not a claim that every image follows the previous deal.

The demo can explicitly simulate another seat's response for review. That control
does not give a real player permission to answer for opponents. No timer or
disconnection is treated as a pass. Opponent hand faces remain concealed; public
hand/Ace counts and physical copies are distinct.

## Historical v1 demo presentation

This describes existing behavior, not P04 adaptive acceptance. V2 replaces the
single-composition restriction with the baseline
[class-specific contract](../docs/design/UI_BASELINE.md#adaptive-composition-and-acceptance-contract).
The local v1 demo shares `CgmsActionBoard` across phone, tablet and desktop. The ordinary mobile fixture
keeps public stacks, action controls and the complete hand on one board. Wider
views enlarge card artwork and give the action workspace more room. The hand is
not hidden behind Table/Hand tabs. Large text, short viewports and denser states
use readable scrolling rather than clipping labels or omitting cards.

Each public stack is an accessible inspection target. Inspection exposes the
individual physical copies at a comfortable size; visible rank tabs stay compact.
Hand cards support selection and inspection. Concealed Aces use the shared back
and public count, never an inferred face.

Routes `#/play`, `#/phone`, `#/tablet` and `#/table` select the board presentation;
`#/setup` opens setup and `#/scenarios` opens the example chooser. Browser
Back/Forward preserves the in-memory board when changing layouts and board views.
Entering setup starts a fresh prepared room branch, and choosing an example
replaces the current branch. Refresh starts a new deterministic session; there
is no saved multiplayer match or durable command queue.

The earlier guided fixture and component/layout examples are retained for
compatibility and regression review. Tests can explicitly construct
`DemoApp(legacy: true)`; the approved application and its chooser expose no
navigation to the retired demo. Legacy widgets are not alternative UI targets.

## Source and reuse

```text
lib/
  main.dart                     Host entry and routing
  online/                       Authority transport, recovery, projection and UI
  approved_demo_controller.dart  Approved fixtures and local transitions
  approved_flow.dart             Approved flow composition
  demo_controller.dart           Earlier regression fixtures
packages/cgms_ui/                Reusable presentation package and assets
test/                           Host state, routing and responsive tests
tool/verify_assets.py            Asset hashes and package-boundary check
tool/browser_acceptance.cjs      Real-authority browser journeys and adaptive matrix
tool/browser_faults.test.cjs     Narrow expected-diagnostic regression tests
screenshots/                    Only the selected cgmsart-phone.png reference
```

Use the [integration guide](INTEGRATION.md) to embed package widgets through
`package:cgms_ui/cgms_ui.dart`. Controllers, services, navigation and authoritative
legality stay in the consuming application. Views never infer rules from artwork.

The package contains all 208 native paintings across First Light, Rain Glaze,
Ember Glaze and Drypoint, a shared back and local
fonts with their original license notices. A bundled Noto Sans Symbols2 fallback
keeps suit/interface glyphs local under the production CSP. Two physical copies
under the same controller share artwork but retain distinct IDs; different
controllers can have different editions. Live Flutter labels provide ranks and suits.
Online player styles come from `projection.board.players[].card_style`, remain
fixed for an entire match and follow the current authoritative controller of
each visible card after capture. All viewers agree; no purchase or unlock is
involved. Hidden cards always use the same neutral back. First Light remains
the standalone and unassigned fallback.
[Asset notices](packages/cgms_ui/assets/NOTICE.md) record provenance and limits.
Static package strings use ARB resources and `CgmsLocalizations`; the current
catalog is English. Dynamic scenario and event text belongs to the host adapter.

## Validate

From `client`:

```bash
dart format --output=none --set-exit-if-changed lib test packages/cgms_ui/lib packages/cgms_ui/test
flutter analyze
flutter test
python3 tool/verify_assets.py
node --test tool/bootstrap.test.cjs tool/browser_faults.test.cjs tool/reconnect_browser_acceptance.test.cjs
flutter build web --no-web-resources-cdn
```

From `client/packages/cgms_ui`:

```bash
flutter gen-l10n
flutter analyze
flutter test
```

Repository agents run long checks through `xops/agent/safe-run.sh` and follow
parent-only tracking/staging. The public consumer test verifies that the package
can be consumed without importing host routes or a demo controller. Test success and screenshot review alone do not establish native-device support,
screen-reader certification, full rules conformance or production readiness.
Online operation recovery has separate executed authority/browser evidence.

Historical local-demo verification on 2026-09-29: 85 host tests and 50 package tests pass; both analyzers,
formatting, asset hashes and the local-resource release web build pass. Browser
review at 390 × 844, 768 × 1024 and 1440 × 900 confirms the full hand/table layout.
The walkthrough exercises opening and attack responses, trade acceptance,
two-card purchase, Coup during a response, settlement, the next game, and
reconciliation of an uncertain action. The browser reports no console errors or
warnings. These actions use the local demonstration controller.
The freeze adds a regression check that all 20 examples remain available while
the approved chooser and menu contain no retired-demo navigation.

## Relocated application verification

The [2026-10-01 verification report](../docs/reports/2026-10-01-nonproduction-verification.md#independent-follow-up--nonprod-recheck-20261001)
records the verified direct move, independent backend/client repairs, red/green
regressions and exact remaining blockers. Final source checks passed **163 host
tests and 50 package tests**, both analyzers, formatting, asset validation and a
local-resource release build using Flutter 3.47.5/Dart 3.13.4. The normal
`make go.re MODE=local` / `make web.re MODE=local` workflow serves this build;
index/bootstrap/main bytes were compared with the local developer server.

The [retained verification summary](../docs/reports/2026-10-01-nonproduction-verification.md)
records **420 passing case executions**, 19 reports and 145
final screenshots on one application build and final harness. Its 396 baseline
cases cover Chromium, Firefox and WebKit phone/tablet/desktop journeys, all
breakpoint/rotation/RTL/200%-text/placeholder matrices, a Chromium three-player
variation and exact-CSP rendering. The 24 economy cases cover private account
views, spending, lost-reply/reload receipt recovery, finalized-game rewards and
streamed allowance wait/pass-based next-game admission. Unexpected diagnostics
are zero. Controls are exposed and hit-tested before their unchanged 48-pixel
minimum is checked; clipping a scrolled control is not mistaken for its full size.

These are real local-authority journeys with explicitly synthetic rule-boundary
states and initial economy funding. They do not establish a complete natural-deal
online match, native screen-reader/device certification, store acceptance or live
commerce. Failed attempts and superseded reports were excluded from final passing
totals. The original screenshot-directory manifests and logs were later deleted
at the owner's request; the retained report summarizes their historical results.

## Final browser references

The owner requested removal of every file and subdirectory in `client/screenshots`
except the [phone layout reference](screenshots/cgmsart-phone.png) on 2026-10-02.
Historical summaries remain under `docs/reports`; old archive paths are provenance
identifiers, not files still present. See the
[screenshot policy](../docs/reports/2026-10-02-screenshot-retention.md). Temporary browser captures are reviewed
and deleted after verification; do not force-add them to the staging set.

## Accepted online verification

Historical acceptance from 2026-09-30, before relocation and the latest repairs:

The former acceptance manifest recorded release hashes, browser versions and
case outcomes; it was deleted with the screenshot archive. The retained
[historical summary](../docs/reports/2026-10-01-nonproduction-verification.md)
describes that evidence. That build used Flutter 3.47.5 and Playwright 1.63.0.
Chromium 153.0.8010.12, Firefox 155.0 and WebKit 26.6 each passed all 18 full
journey checks at phone 390×844, tablet 768×1024 and desktop 1440×900 sizes.
Each engine also passed the 30-case breakpoint/rotation/RTL/200%-text/art-off
matrix. A three-seat Chromium run passed all 18 journeys; exact-production-CSP
local asset rendering passed separately: **271 browser case executions** total.

All final journeys capture page errors and console errors on fixture and guest
pages. Unexpected errors are zero; deliberately aborted recovery requests and
verified initial signed-out session challenges remain explicit evidence.
Primary controls measure at least 48×48 logical pixels, including the tablet
rail. Keyboard traversal, physical-card inspection, browser Back/Forward and
draft preservation are exercised. These checks use real authority plus synthetic
rule-boundary positions and a normal room/deal/reload flow; they do not claim a
complete natural-deal match or native screen-reader certification.

Independent final source verification passed 129 host tests, 50 package tests,
both analyzers, formatting and asset validation. Six browser-diagnostic regression
tests verify that unrelated application, CSP, CORS and resource errors still fail.
The requirement table above identifies paired engine evidence and remaining
native boundaries rather than treating browser coverage as a second rules engine.

## Online browser references

Historical local-authority results are summarized in the
[verification report](../docs/reports/2026-10-01-nonproduction-verification.md).
The original manifests and image files were deleted at the owner's request;
those results do not establish the
new shared vertical layout or native-device acceptance.

## Rendered board evidence

The sole retained screenshot is [Phone, 390×825](screenshots/cgmsart-phone.png),
a real Flutter widget render of the original board fixture with bundled art/fonts.
It is the owner's layout reference, not a fresh full-host browser acceptance
capture. Preserve its bytes; optional widget rendering must write temporary files
under ignored `.browser-tools/artifacts/widget-renders/` rather than replace it.
Full-host behavior is checked separately against the reference composition,
including real actions, scrolling, all width classes and enlarged text.
