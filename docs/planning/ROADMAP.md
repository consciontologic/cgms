# CGMS roadmap

The sole sequenced implementation checklist. Revised 2026-09-30 for web-first
adaptive development and deferred native acceptance; retaining the 2026-09-28
architecture and Go backend plan, aligned with accepted
[GAME_RULES](../project/GAME_RULES.md) rules and the adopted F01–F10/N01–N07/T01–T03 review
amendments. Rule decisions are accepted; implementation and evidence are pending. Checked items have runtime evidence; unchecked items remain future work.
Writing a plan does not complete an implementation phase. The broader
[project brief](../project/INITIATE.md) supplies product context; its complete
documentation assignment is not claimed finished in this slice.

## Goal

Make CGMS rules reproducible and inspectable in a useful CLI simulator first,
then deliver a durable Go online authority and the client/product integration
needed for a Docker Compose production v1.

## Non-goals

Phase 0 runtime is bounded by the checked gates below. Kubernetes/Helm,
microservices, distributed match-owner leases and mandatory LLMs are deferred.
Future experiments must not silently change accepted GAME_RULES gameplay rules.

## Status snapshot

Flutter web is the current development/testing platform. The owner revision of
2026-10-04 in the [UI baseline](../design/UI_BASELINE.md) selects one shared
vertical board across phone, tablet and desktop, retaining cgmsart. The existing local
preview and completed evidence remain valid within their recorded scope;
adaptive online integration passed its recorded three-engine browser acceptance.
Native Android/iOS acceptance is deferred to R06N immediately before R06/R07.
Unavailable devices or Xcode do not block P01 host research or P04 web work.

| Phase | Items | Done | Status |
|---|---|---|---|
| 0 — CLI simulation | 10 | 10 | Complete (bounded scenarios) |
| 1 — Complete rules and durable matches | 9 | 9 | Complete (local durable authority) |
| 2 — Online API and multiplayer recovery | 8 | 8 | Complete; independent admission-priority/database repairs reverified |
| 3 — Offline and product integration | 6 | 5 | P06 local product contracts passed; P03 allocation remains open |
| 4 — Observability and operational workflow | 5 | 5 | Complete within documented local operational criteria |
| 5 — Production v1 | 8 | 2 | Bounded load and rollback verified; release gates remain open |

Independent follow-up at `dd2858c` (2026-10-01): the requested direct migration
was already committed and has been checked against its original tracked inventory.
Within O04/P02/P04, repair newly demonstrated admission-priority, database-error,
bot-menu budget, action-draft, cost guidance and lobby reload deficiencies while
preserving original historical evidence. Touched areas include engine candidate
pruning, delivery/storage boundaries, client route/state/tests and active documentation. Verification uses deterministic red/green regressions, independent
review, real-database/race checks and a fresh final-build browser matrix. P03
allocation remains blocked; decisions 0015–0016 replace P06 catalogue/channel
questions with the five-product model and shared match styles;
the same Phase 5 exclusions above apply. No new bot evaluation budget is implied.

## Direct card drops and automatic draw feedback — 2026-10-04

Owner follow-up supersedes the blanket card-move confirmation requirement below.
The current scope retains the four public quadrants, adjacent private hand,
cgmsart and the removed permanent composer, with server-owned rule resolution.

- [x] Reserve safe draw-pile space across responsive layouts and verify 1041×948,
  live breakpoint transitions, short/rotated views and RTL at 200% text.
- [x] Present confirmed scheduled-turn draws once, with privacy-safe destinations,
  reduced motion, and no animation-triggered commands or reconnect replay.
- [x] Preview exact actions/cards/costs/destinations before release and submit one
  complete unambiguous legal drop; retain small choices for missing parameters.
- [x] Cover own-area numbers/royals/combos, context-correct Aces and intact combo
  loans through exact proposal acceptance, responses and atomic alliance creation.
- [x] Reverify authority-bound stale/invalid/cancelled gestures, physical copies,
  uncertain submissions, explicit utilities, focused keyboard and touch input.
- [x] Pass frontend/shared analyzers and tests, affected authority/privacy/retry
  gates, final-build browser scenarios and one-human/three-bot settlement.
- [x] Update baseline/changelog/evidence, independently review the complete diff,
  append tracking and stage without committing or pushing.

## Cards as the gameplay interface — 2026-10-04

Owner revision superseding the permanent action workspace requirements below.
Scope: existing `client/` and reusable `cgms_ui`; accepted rules, artwork,
authority, privacy and financial progression remain unchanged.

- [x] Audit the full command catalogue and record the rule → command → gesture → test matrix.
- [x] Remove the permanent online action/composer/review area and its layout reservation.
- [x] Implement physical selection, double-tap/Enter activation, drag/tap destinations,
  independent inspection, card-specific choices and explicit commitment.
- [x] Preserve exact source/payment/target roles, authorized ordered decisions,
  immediate card-driven Coup, explicit Pass, financial utilities and stale/unknown operation safeguards.
- [x] Complete final frontend/shared analyzer and regression gates, rebuilt isolated
  browser journeys and a disposable one-human/three-bot match; capture actual coverage.
- [x] Review complete diff/staging set, update final evidence, track and stage.

Coverage, executed verification and platform limits: [interaction evidence](../reports/2026-10-04-card-interaction-redesign.md).
Historical checked P04 results apply to their recorded builds; this revision does
not claim native or unexecuted browser-engine acceptance.

## Card-driven online play — 2026-10-03

**Goal:** replace the generic action form with card selection, contextual moves
and drag/drop preparation for series, combos, abilities, combat and responses.
**Non-goals:** change game rules, automate human decisions, expose hidden cards,
or replace server validation with a client legality engine.
**Touched files:** online screen/interaction helpers and tests; shared adaptive
table/models and tests; public rules/decision projection, cached-snapshot and
historical-journal compatibility, wire/privacy tests and OpenAPI schema; browser
acceptance helpers; current UI baseline, changelog and evidence report.
**Test plan:** rule-mapped intent/payload cases, pointer drag and tap/keyboard
equivalence, ordered combat/effect responses, ambiguous choices, restricted
cards, stale drafts, privacy, full frontend gates and rebuilt live preview.

- [x] Map all card-driven intents and response stages to current command contracts.
- [x] Add accessible card drag sources and table/deck drop targets with regression tests.
- [x] Replace the generic action chooser with selection-driven contextual moves and explicit draft confirmation.
- [x] Integrate combo/special-card choices, attacks, defences and required-effect decisions while preserving financial/trade workflows.
- [x] Review, verify regression suites and live play surfaces, update evidence and stage.

**Risks:** a gesture can be ambiguous and a visible candidate is not proof of
legality. Drops only prepare drafts; exact choices, costs and confirmation stay
visible, and the authority revalidates immutable game/decision/card handles.
Keep tap/keyboard alternatives, untouched pending state and honest coverage limits.

## Readable public cards with open actions — 2026-10-03

**Goal:** repair the public-card shrinkage exposed by the owner's 1025×948
preview while retaining the deck, adjacent hand and usable action controls.
**Scope:** adaptive height allocation, rendered-size regressions and UI evidence.
**Test plan:** measure transformed public-card bounds with open actions, keep a
complete first private card visible at ordinary sizes, and recheck short views,
enlarged text, RTL, state retention and the live human-plus-three-bot preview.

- [x] Reproduce the 24px public cards and repair the competing height budgets.
- [x] Review, run frontend checks, inspect the rebuilt preview and stage evidence. 383 frontend tests and both analyzers pass; normal tablet-class browser preview verifies the larger cards and restored draft.

## Visible table and balanced online layout — 2026-10-03

**Goal:** show the face-down draw pile, keep the table visible while preparing an action, remove excess space
between hand suit groups, and center related controls on tablet and desktop.
**Scope:** shared adaptive table, online form, matching tests and existing UI evidence.
**Test plan:** reproduce hidden table and stretched hand gaps; verify action-row
geometry, selection/draft retention, short windows, text scaling and RTL; run
frontend suites/analyzers and inspect the rebuilt human-plus-three-bot preview.

- [x] Add the neutral draw-pile marker with privacy/semantics coverage; restore table visibility and compact centered hand groups with regressions.
- [x] Center wide action controls while preserving drafts and mobile behavior.
- [x] Review, verify the live preview, update evidence, track and stage. Normal browser preview and 381 frontend tests pass; viewport-emulation limits are recorded in the UI evidence report.

## Compact online action workspace — 2026-10-03

**Goal:** remove the redundant visible “Choose an action” strip, reduce action
text/icon density and arrange status, selectors and primary controls in rows so
ordinary desktop actions do not need panel scrolling.
**Scope:** existing online screen/tests, local presentation and evidence. Preserve
the prior staged typography fix, action ownership, drafts and pending operations.
**Test plan:** failing layout regressions, full frontend suite/analyzer, normal
preview at desktop/tablet/phone, live resize, keyboard, enlarged text and RTL.
**Risks:** reparenting loses focus or drafts; fixed heights clip scaled labels;
hidden guidance obscures required response controls. Keep stable stateful controls,
full touch targets, visible required actions and accessible optional details.

- [x] Implement compact responsive form and matching regression coverage.
- [x] Verify the rebuilt preview and inspect source/test review findings.
- [x] Retain concise evidence, inspect all staged changes, track and stage.
  299 host and 76 shared tests, both analyzers, final build and local Chromium
  browser review passed. Run `action-density-20261003`; extends the
  [online UI report](../reports/2026-10-03-online-typography.md).

## Online typography and alignment — 2026-10-03

**Goal:** give the online lobby a centered introduction and clear heading hierarchy,
with readable, modest dropdown typography in setup and gameplay, bounded menus
and a centered gameplay action form.
**Non-goals:** game rules, bot policies, public-table composition and global theme changes.
**Files:** existing online screen and widget tests, UI baseline, changelog and a dated
verification report. Preserve other work and the sole phone-reference screenshot.
**Test plan:** reproduce inherited bold dropdown text and left-only hierarchy;
regress actual opened menus, enlarged text, RTL, keyboard/touch targets and drafts;
run the online widget suite and analyzer, build and inspect the served preview.
**Risks:** fixed menu heights clipping enlarged labels and rebuilds losing drafts;
retain intrinsic row heights, 48px targets, state ownership and command boundaries.

- [x] Add failing regressions and fix local heading alignment and dropdown text.
- [x] Verify phone/tablet/desktop preview, menus, keyboard and 200% text behavior.
- [x] Review the complete diff, update evidence/tracking and stage.
  Verification: 297 full-suite frontend cases plus one additional semantic-target
  case, analyzer/format/build and in-app Chromium review passed. Evidence:
  [typography report](../reports/2026-10-03-online-typography.md),
  run `online-type-20261003`.

## Play against bots — 2026-10-02

**Goal:** start a real online match as one human against two or three fair bots
from the normal Flutter lobby, and continue playing through human/bot turns,
responses and reconnects. Scope is existing room/transport/runtime and client
integration. Reuse the offline beginner/standard/advanced policies and the Go
authority. This is not P03 strength evaluation, a new game engine, an LLM service,
offline web support or a change to economy/gameplay rules.

**Files:** backend room, delivery/runtime, transport and server lifecycle modules,
their tests and protocol contract; client online client/screen/localizations/tests;
existing browser acceptance, operational guidance and this roadmap.

**Contract:** an optional room bot difficulty fills the remaining seats with
server-controlled bots; omitted difficulty retains human rooms. The owner starts
through the existing operation. Persist bot ownership and use the same validated,
idempotent command/persistence/projection boundaries as human play.

**Risks:** duplicate or stalled bot work, lost responses, leaked private cards,
spoofed bot actors and reconnect races. Bound work, serialize authoritative state,
derive policies from authorized observations, and test failures before retrying.

- [x] Add room/bot ownership and runtime integration with red/green tests for
  turn/response progress, privacy, idempotence and restart/reconnect recovery.
- [x] Expose bot setup and seat labels in the existing lobby with regression
  tests for validation, pending operations and preserved human-room behavior.
  Round boundaries now expose an explicit Stay/Leave decision; bots record their
  independent Stay choices after the human answers to preserve first-click
  version validity. Existing saved views and immutable journal receipts remain
  compatible with the added round metadata. Authoritative readiness gates
  provisional starter controls, and decision workspace changes reset web
  semantic scroll geometry while retaining form state.
  Online suggestions reject zero-growth purchases to break the reproduced
  empty-deck buy/reopen cycle; human rules and frozen offline policies stay intact.
- [x] Verify actual browser creation, human commands and autonomous bot replies
  at phone/desktop sizes, reload recovery and normal local startup configuration.
- [x] Review full diff and required gates; retain concise evidence, remove
  disposable captures, update tracking and stage without committing or pushing.
  Verification: 411 frontend/support tests, complete Go/vet and seven-package
  PostgreSQL/Redis race gates, plus 27 browser cases across nine real bot matches.
  [Evidence and remaining attribution limits](../reports/2026-10-02-online-bots-verification.json)
  are retained under run `online-bots-20261002`; 97 disposable captures removed.

## Shared vertical reference revision — 2026-10-02

**Compact header row — `compact-header-20261002`.** Owner revision: keep the
round/turn text and header tools side by side and vertically centered at compact
and near-tablet widths, using smaller visible type/icons with deliberate spacing.
Preserve 48px targets, full Auto/round/turn labels, enlarged-text accessibility,
focus/state and the existing footer/table/hand. Scope: existing shared header,
normal selector and tests, browser harnesses and current UI guidance; no rules,
backend, assets or footer redesign. Risk: clipping or shrinking accessibility
text; use measured width budgets and stable controls instead of a forced two-row
breakpoint.

- [x] Reproduce the compact two-row header; implement and verify the centered
  compact row with regression tests for typography, targets, focus and text scale.
- [x] Verify normal/online browser flows, live compact and class boundaries,
  rotation, RTL/200%, keyboard/touch, and the final served build; review visuals.
- [x] Remove disposable captures, preserve earlier staged work, review, track
  and stage the completed revision.

**Responsive bars — `responsive-bars-20261002`.** Balance status and tools across
available header width, and group bottom utilities and primary actions in an
opaque responsive dock with deliberate compact reflow, spacing and control sizes.
Scope: shared board chrome, normal/online host slots, existing widget/browser
regressions and current UI guidance. No gameplay, assets or startup changes.
Touched files: `adaptive_table.dart`, normal/online host widgets, their tests,
existing browser harnesses, baseline, changelog and compact verification report.
Risks: clipped text/targets, displaced hand, lost focus or duplicate commands;
retain stable control trees, natural text sizing and bounded viewport budgets.

- [x] Reproduce intrinsic left clustering; implement responsive bars with
  red/green alignment, target-size, live focus and state regressions.
- [x] Verify normal/online browser layouts at phone/tablet/desktop, live threshold
  transitions, rotation, short height, RTL/200%, keyboard and touch; inspect visuals.
- [x] Remove temporary captures, record final evidence, review, track and stage.

Verified on build `e9011d03…`: 276 host, 76 shared UI and 31 Node tests,
both analyzers, and 357 browser cases across Chromium, Firefox and WebKit.
The [verification record](../reports/2026-10-02-responsive-bars-verification.json)
retains root causes, failed-attempt diagnoses, final geometry and cleanup evidence.
322 temporary captures were inventoried and removed; the sole phone reference
and 1,248 protected artwork/design files are unchanged. Web evidence only.

**Full-width follow-up — `public-width-20261002`.** The owner selected a public
table as wide as the hand, with four equal rectangular quadrants. This supersedes
the square-only constraint below. Use the hand's card-size scale for public
previews where space permits, and shrink only for actual height/pile constraints.
Scope: shared board, matching widget/host/browser tests and current UI guidance;
no gameplay, authority, connectivity or asset changes. Preserve mounted state,
privacy, equal cells, adjacent visible hand and keyboard/touch controls.

- [x] Reproduce width/painted-card mismatch; implement aligned full-width cells
  and proportional cards with red/green regressions, including dense piles.
- [x] Verify final normal/online browser behavior, short/RTL/200% layouts and
  live resizing; remove captures, update evidence, review, track and stage.
  Evidence: [379 source tests, 351 browser cases and final cleanup](../reports/2026-10-02-public-width-verification.json).

**Public-size follow-up — `public-growth-20261002`.** Make the public square and
its painted cards grow smoothly with wider screens at the same height, including
the owner's 821×948 preview. Keep the equal quadrants and adjacent visible hand.
Scope: shared board sizing, widget/browser regressions and current UI guidance;
gameplay, assets and host state remain unchanged. The main risk is consuming the
hand's space, so retain short-height, RTL/200%, focus and pending-state gates.

- [x] Reproduce capped public sizing, then implement width-responsive growth
  with transformed card-paint bounds and initial hand-visibility regressions.
- [x] Verify live resizing without reload in the normal and online browser hosts,
  run relevant source checks, remove temporary captures, review, track and stage.
  Evidence: [378 source/support tests and351 browser cases](../reports/2026-10-02-public-growth-verification.json).

**Current correction: public quadrants and adjacent hand.** Keep the public
overview in one compact square split into four equal seats at every width,
including the actual 319px preview. Reserve visible space for the private hand
so tall public piles or action forms cannot push it below the page.

**Scope:** shared Flutter board, necessary host constraints, matching widget and
browser regressions, current UI guidance, and screenshot-directory cleanup.
Gameplay, artwork, privacy and command identity stay authoritative and unchanged.

**Tests:** reproduce the 319px one-column regression, assert equal 2×2 square
geometry and initial hand visibility at phone/tablet/desktop and short heights;
exercise resize boundaries, RTL, 200% text, touch, keyboard, drafts and pending
operations against the actual served build. Public previews can shrink while
full details and physical-card controls remain accessible.

- [x] Implement bounded public quadrants and adjacent hand with red/green tests.
- [x] Delete all files and subdirectories under `client/screenshots` except
  `cgmsart-phone.png`; repair active links and retain compact reports elsewhere.
- [x] Verify source and real-browser gates, review the complete diff, track and stage.
  Evidence: [377 source/support tests and 388 browser cases](../reports/2026-10-02-quadrant-verification.json),
  run `quadrant-board-20261002`.

**Risk control:** retain mounted state and stable physical-card semantics;
confine dense or enlarged content to deliberate internal scrolling or details
without making the player scroll the page to find their hand.

Owner follow-up: normalize private hand card sizes and horizontally overlap
cards by suit (Hearts, Spades, Diamonds, Clubs), superseding the six-column hand.
Keep full artwork inspection, physical identities, selection and focus.

- [x] Implement uniform hand frames and accessible horizontal suit stacks with
  red/green regressions; verify the served browser build and stage this follow-up.

The owner now selects `client/screenshots/cgmsart-phone.png` as the board layout
for every width: compact public two-column seat overview, action area, private
hand and bottom actions in one vertical sequence. Tablet/desktop enlarge items
without changing this arrangement. This supersedes the preceding repair's
separate-pane requirement; it does not undo verified startup recovery, privacy,
state retention or command identity. All other screenshot images are explicitly
authorized for deletion, including historical captures and other widget renders;
keep the phone reference unchanged, artwork/mockups/fixtures, and lightweight
verification summaries. No Git history rewriting is part of this work.

- [x] Implement the shared vertical layout and proportional item sizing with
  red/green Flutter regressions and preserved state/selection/focus/privacy.
- [x] Delete all other screenshots, repair active references, prevent future
  staging of generated captures and preserve the selected image's hash.
- [x] Verify normal entry and online browser behavior across widths, resizing,
  keyboard/touch, short/rotated views and enlarged text; review, track and stage.

Evidence: [shared vertical revision](../reports/2026-10-02-shared-vertical-layout.md)
and [401-case verification](../reports/2026-10-02-shared-vertical-verification.json).

## Adaptive and startup repair — 2026-10-02

Historical preceding repair: its composition and screenshot-retention policy
are superseded by the owner revision above; its startup safeguards remain active.

Within P04/V05, restore the normal entry's Auto composition at 600/1050 logical
pixel boundaries, compact mockup-derived public context, dedicated phone hand
and decisions, and state/focus continuity. Repair demonstrated first-load
session readiness and reconnect recovery with bounded read-only retries.
Inventory screenshot purposes and Git history before removing only disposable
captures; keep historical outcomes and required references. Touched paths are
the existing Flutter host/package/tests, browser runners, startup wiring and
active verification documents. Preserve staged work, gameplay/privacy authority,
artwork, persistent data and all unapproved Phase 5/P03 exclusions.

Acceptance order: reproduce the served build and startup fault; add red/green
regressions; implement; run host/package/backend checks; verify normal-entry
direct loads, live boundaries/rotation/short viewports and online accessibility,
privacy/pending commands; verify three independent cold starts plus delayed
readiness and reload/reconnect; review the whole staging set and stage. All
browser captures go to ignored artifacts. Report timings, hashes and remaining
limits in the [repair report](../reports/2026-10-02-adaptive-startup-recovery.md).
Existing phase counts remain unchanged: this repairs accepted local gates and
does not establish native acceptance or confirm a remote push will succeed.

- [x] Inventory screenshot purposes, references and reachable Git blobs; retire
  disposable captures, preserve protected evidence and add narrow output rules.
- [x] Restore and verify normal-entry Auto composition, boundary transitions,
  keyboard/touch access and retained private state on the final served build.
- [x] Verify bounded startup/reconnect recovery through regression coverage,
  three independent cold launches and delayed-readiness/reload cases.
- [x] Review final integrated evidence and the complete staging set, reconcile
  recovery records, append tracking and stage without committing or pushing.

## Non-production execution sequence — 2026-10-01

Goal: relocate the complete Flutter host directly into `client/`, verify the live
non-production implementation and repair demonstrated backend/client gaps.
Phase 5 (including R06N), deployments, live commerce, native release acceptance
and unapproved P03 allocations are excluded; existing production work is retained.
Touched paths: `client/`, affected `xops/`/Compose wiring, backend rules/authority/
economy integration, active documentation and dated acceptance evidence.
Sequence within the existing gates: inventory and migrate P04 host/package;
restore and test local build/start wiring; review and repair F/N/T backend and P04
journeys; complete decidable P03/P05/P06 local contracts; independently review,
verify final integrated state, then record exact remaining blockers and stage.
Tests: Flutter host/package analysis and tests, asset checks, local release build,
real PostgreSQL/Redis integration and Go race checks, final real-authority browser
journeys and the complete v2 viewport/transition matrix across three engines.
Risks: relocation may invalidate relative paths/caches; preserve historical hashes
and generate new evidence. Gate checkboxes stay unchanged until their complete
acceptance passes. Missing product/resource decisions do not stop independent work.

## Canonical design and proposed touched paths

| Concern | Design | Proposed implementation paths |
|---|---|---|
| Boundaries/configuration/deployment | [Architecture](../code/ARCHITECTURE.md) | `backend/`, `client/`, `client/offline_probe/`, `config/`, `deploy/` |
| Simulator/engine choice | [Simulation ADR](../design/ADR-0001-shared-go-simulation-engine.md), [CLI contract](../../sims/README.md) | `backend/cmd/sim/`, `backend/internal/game/`, `backend/internal/bots/`, `sims/` |
| Rules and acceptance fixtures | [Resolved rule register](../design/RULES-IMPLEMENTATION-QUESTIONS.md) | `backend/internal/game/`, `backend/testdata/` |
| Transactions/retries/streams | [State/delivery design](../design/DESIGN-go-state-and-delivery.md), [API contract](../code/API-game.md) | `backend/internal/{matches,storage,transport}/`, `backend/migrations/` |
| Developer lifecycle | [Existing xops conventions](../../xops/README.md) | Thin Make targets and `xops/makefile/` Python dispatchers |

Runtime paths above are proposed. Preserve existing assets, templates,
historical rules, tracking and unrelated user changes.

## Accepted review traceability

The [accepted amendment/scenario register](../design/RULES-IMPLEMENTATION-QUESTIONS.md#accepted-interaction-amendments-and-scenario-pack)
connects all three accepted reviews' recommendations to this single checklist. Preserve
its shared trace inputs and expected card zones, counters, exact ledgers, phase/
decision IDs, ordered events and per-seat observations across engine, simulator,
server and reconnecting client. No row below completes a runtime task.

| Accepted finding / operating decision | Implementation items | Required evidence before those items close |
|---|---|---|
| F01 — effect decisions and Coup boundary | S04, S06, D08, O03–O05, P04 | Negotiator/Justice/Dexter actor/stage inputs, committed randomness, duplicate/stale IDs, restart and Coup at every legal wait; no transaction held while a human decides. |
| F02 — costs, timing and quotas | S06, D01–D03, O03, P04 | Per-ability §2.12 legal/illegal matrix, declaration versus resolution costs, cancellation, physical-Club exception, player quotas across copy/loan/reconfiguration/Fate, ordinary versus Baron Suicide Bomb. |
| F03 — current-state victory admission | D06–D09, O03–O05 | Within the named game, unrelated stale version preserves a legal Coup; changed eligibility and closed board reject; ordinary boundary remains required; two claims and retries yield one ending. N05 rejects different-game first submissions. |
| F04 — complete financial transitions | S07, D04–D07, D09, P04 | Coup clears old cash before replacement, exact surviving debt adjustments, departed exemption, origin-game cash, zero-cash next game, departure/nullification fixtures. |
| F05 — sequential financial settlement | D04–D09, O03, O08, P04 | Typed award promises, board-closed allow-list, pay/accept/reject/refuse/finish, debt priority, pending reconnect, next-game admission only after finalization. |
| F06 — acquisition destinations | S02, S04, D02–D03, O05, O08, P04 | Justice Diamond disclosure, captured Spades to hand, distinct trade/loan/return rules, opening history and retained public knowledge. |
| F07 — lasting effects and custody | D03, D05, O05, O08, P04 | Underground privilege persists for earner and valid borrower; allocated income kings only; Code settings/declarer versus physical custodian and recurring penalty. |
| F08 — dormant Doppelganger bindings | S02, D02–D03, D05, O05, P04 | Split/reunited/released states, one binding per king, forbidden third pairing, assassination/recycling and private-location projections. |
| F09 — one Inflation and removal order | S06–S07, D03, O03, P04 | Duplicate no-op, exact half values, 19/20 printed threshold, canceled/forced/cumulative transactions, active-price-then-clear ordering. |
| F10 — authorized counts and projections | S04, S08, D05, O03, O05, P04, V01 | Public hand/Ace counts and private faces/breakdowns consistently across snapshots, events, history, screen readers, bots and logs. |
| N01 — effective combat and physical cost | S06, D01–D03, O03, P04 | `A`/`D` formulas, frozen targets, any-series numerical defense versus Hearts-only Advancement, Baron comparison, modifier prevention and physical-only exchange returns. |
| N02 — persistent-Ace custody and reset | S02, S04, D02–D03, D05, O05, P04 | Public source/target/custodian/expiry effect zone, Fate exclusion/counts, Compensation remainder, caster/target departure distinction and end cleanup preserve all 104 cards. |
| N03 — eligible response traversal | S06, D01, D08, O04–O05, P04 | Skip confined/departed seats without requiring their pass; preserve fixed order, eligible disconnected waits, expiry and separate confined Coup permission. |
| N04 — Code economics and hostile penalty | D03–D04, O03, P04 | Non-attacking settings affect all opponents; recurring penalty requires declarer's two series and excludes protected/confined seats; Underground, custody and redeclaration cases. |
| N05 — immutable game command scope | D05–D09, O03–O05, P04 | Required game IDs, duplicate-before-game-check order, delayed wrong-game first submission, fresh nullified replacement identity, related decision IDs and no automatic client retargeting. |
| N06 — forgiveness provenance | S07, D04–D05, D09, O03, P04 | Cross-game nonspendable match correction against retained original charge, partial repayment, Coup survival, current-game correction replacement, nullification restoration and duplicate prevention. |
| N07 — physical availability restriction | S02, S04, D02–D03, O03, O05, P04 | Physical round marker blocks voluntary use/Coup, permits ordinary threshold/proof and mandatory Diamond exposure, survives forced movement/Fate/shuffle/redraw, overrides immediate use and preserves privacy. |
| T01 — proposal and accepted-transfer lifecycle | D03, D05–D08, O03, O05, O08, P04 | Private revisioned nonblocking offers; no reservation; origin-turn/closure/departure expiry; withdraw/decline/accept races; acceptor actor/Confinement; final revalidation, no partial transfer, and alliance/Underground privilege only on success. |
| T02 — deferred Compensation draw order | S04, D02–D03, D08, O05, P04 | Fixed ascending-seat queue, complete recipient quantities, return/shuffle first, exhausted entitlements consumed, committed random progress and distinct Coup/ordinary-action boundaries. |
| T03 — explicit deck purchase quantity | D02–D03, O03, P04 | Buyer-chosen positive quantity; eligible hand/exposed physical Spades; exact Code/Inflation cost and full supply including returned payment; post-response revalidation and atomic payment/shuffle/draw, no partial purchase. |
| Exact settlement cost and recovery | S07, S09, D04, D08–D09, R02–R04 | Small-ledger reference equivalence, compressed reproducible audits and durable progress; large-D restart resumes without rounding, loss or repeated rollback of all work. |
| Untimed waits and release economy | D08, O04, P04–P05, R06 | Resumable exact waiting actor and hibernation; explicit abandoned-game allowance/reward policy before public/monetized release, without inventing automatic game outcomes. |
| Fixed match length and tied ranks | S03, S09, D09, O02, P04 | Visible immutable positive game count with default three; shared ranks on ties; declaration, game score leader and cumulative match rank shown separately. |

## Assessment status

All Q01–Q12, F01–F10, N01–N07 and T01–T03 decisions are accepted in
[GAME_RULES.md](../project/GAME_RULES.md). The first, second and
[third assessments](../reports/2026-09-28-final-draft-third-assessment.md)
remain historical evidence of how the contracts evolved; their former proposal
status does not reopen these decisions. The accepted matrix above connects every
amendment to this checklist and its required evidence. Runtime completion is recorded only by the checkboxes below; documentation
consolidation does not complete an implementation gate.

## Phase 0 — CLI simulation

**Scope:** `phase-0`. **Dependencies:** existing rules and this plan; no Flutter,
PostgreSQL, Redis, production telemetry or LLM prerequisite.
**Goal:** useful accepted-rule scenarios and explicitly named experiments.

- [x] **S01** Create one `backend/go.mod`, pin a supported toolchain and establish pure-engine tests. Gate: observe a deliberately failing behavior test before its minimal implementation; document the real test command.
- [x] **S02** Model distinct physical cards, zones and immutable state. Gate: 104 cards, two copies per rank/suit, conservation after acquisition/loan/return moves, persistent binding states, active-Ace effect zones, physical availability markers through forced moves/Fate and no state alias mutation (F06/F08/N02/N07).
- [x] **S03** Add canonical rules JSON/schema, bounds, units, compatibility and normalized hash vectors. Gate: unknown/invalid configuration fails; equivalent effective configs hash equally.
- [x] **S04** Add transition, observation and seeded-random boundaries. Gate: fixed inputs repeat exact state/event hashes; effect decision IDs/stages and committed random outcomes survive continuation; §2.13 projections expose approved counts only; engine/random versions and private replay provenance are retained (F01/F10); T02 queue/counters/draw outcomes preserve fixed-seat ordering, exhausted supply and legal Coup boundaries.
- [x] **S05** Implement deal/redeal, initial diamonds, initiative and series opening. Gate: no-diamond redeal, draw-off isolation, frozen order, opening allowance and current/history protection fixtures; returned-card shuffling, exhausted supplies and rotating initiative priority follow Q07.
- [x] **S06** Implement narrow ordinary-combat and response/pass scenarios. Gate: exact matching, attack order, physical-club usage and non-nested eligible-seat responses; Q01/Q02 and F01/F02/F09 costs, Confinement and continuation distinguish legal activation from resolution; N01 algebra separates virtual values from physical returns, and N03 skips ineligible seats without auto-passing disconnections.
- [x] **S07** Add exact rational score/debt receipt scenarios. Gate: exact paired Ponzi halves, generic rational thirds, oldest-first repayment and no duplicate score deduction; reproduce the rules' worked ledgers, F04 game-scoped cash transitions and N06 charge/correction provenance through Coup/nullification; establish an unbatched reference and exact batched equivalence cases with durable-progress format.
- [x] **S08** Add legal-random/initial heuristic bots and named experiment profiles. Gate: fair observations, independent legality validation and explicit incomplete/experimental labels; no claimed full rule coverage.
- [x] **S09** Implement `run`, `replay`, `batch`, `tournament`, `sweep`, `compare` and Markdown reports. Gate: printed commands reproduce fixtures; worker count preserves outcomes; seat rotation and three/four-player incomplete denominators appear in reports.
- [x] **S10** Package a simulator container and document host/container use. Gate: isolated execution needs no application services or network; output schemas, retention and provenance satisfy the CLI contract.

**Test plan:** pure Go unit/property/fuzz tests over implemented transitions,
golden config/replay/projection fixtures, CLI exit/report tests and bounded
batch cancellation. Commands become executable in S01/S10, not in this plan.
**Exit:** reproduce one accepted-rule scenario and a useful batch/comparison
report, stating rule coverage and all stopped/experimental runs. Phase 0 is a
partial simulator, not proof the complete GAME_RULES is independently playable.

## Phase 1 — Complete rules and durable matches

**Scope:** `phase-1`. **Dependencies:** Phase 0 and the accepted GAME_RULES
contract. All Q01–Q12, F01–F10, N01–N07 and T01–T03 decisions are accepted; their implementation and
verification are tracked by the gates below.
**Goal:** fully covered rules and recoverable authoritative matches.

The 2026-09-30 owner authorization extends shared-engine/CLI work to D01–D04
and the simulator/domain portions of D08–D09, plus bots and evaluation. Database
and server gates were excluded by that earlier authorization. The subsequent
2026-09-30 instruction authorizes the remaining Phase 1 D05–D09 implementation;
Phase 2 transport/authentication remains a separate phase. [Gameplay coverage evidence](../../sims/docs/gameplay-coverage.md)
separates implementation, targeted tests and autonomous exercise; it is not a
second roadmap. Phase 0’s 2,000 partial-prefix replays remain historical partial
evidence, with zero completed games.

**D05–D09 execution plan (2026-09-30).** Goal: durable authoritative match
transitions survive retries and restart. Non-goals: HTTP/WebSocket delivery,
account authentication, Redis, client work, or expanding the P03 compute allocation.
Touched files: `backend/internal/matchstore/`, backend module dependencies,
`xops/` database test runner, and existing architecture/roadmap documentation.
Test plan: real PostgreSQL migration/rollback, competing connections, fault injection,
receipt reconciliation, private projections, persisted continuation and lifecycle
fixtures, followed by Go regression/race/vet and independent review.
Execution follows the five existing gates below: schema and compatibility; locked
atomic writes; durable retry receipts; restartable work; finalization and next game.
Risks: stale retries, ambiguous commit acknowledgements, projection leakage and
partial settlement; mitigate with primary-database receipts, explicit version pins,
seat-scoped reads, and atomically persisted complete engine checkpoints.

- [x] **D01** Implement and verify Q01/Q02 cancellation and Confinement. Gate: spent committed clubs, retained paid costs, no success-only costs on failure, active-player-only turn ending, next-scheduled-turn expiry and confined Coup exception match accepted fixtures; the F02 per-ability cost/quota/timing matrix covers every consumable cancellation; N01 modifier changes and N03 eligible-seat traversal preserve one result.
- [x] **D02** Implement and verify Q03/Q04/Q06/Q07 victory, protection, reset, reopening, deck and aces. Gate: golden cases cover private invalid claims, no People protection chains, Fate retained state, opening history, exhausted supplies, F06 destinations, F08 bindings, N02 persistent-Ace reset/cleanup and N07 restricted-card movement/threshold/Coup distinctions; T02 deferred-draw order and T03 chosen purchase quantity/payment/supply have exact traces.
- [x] **D03** Cover every combo/non-combo/ace activation, support and usage rule, including Q08. Gate: Negotiator maximal exact matching with committed/unused clubs, Dexter modes, Justice selection, Compensation and Exile/Barricade fixtures; legal/illegal coverage follows §2.12; F07 lasting ownership versus custody, allocated Underground income, F09 Inflation and committed randomness have expected outcomes; N01/N02/N04/N07 govern comparison/cost, Ace lifetime, Code penalty eligibility and physical availability; T01 accepted transfers, T02 draw queues and T03 purchases follow their canonical lifecycles.
- [x] **D04** Implement Q05/Q09/Q10 alliance sharing, departure/resignation/forgiveness, debt cascades and Coup reconciliation. Gate: exclusive pairs, paired Ponzi halves, FIFO receipt cycles with oldest-first debts, termination and batching-equivalence evidence, reproducible compressed audits, F04/F05 boundary ledgers, N04 Code penalty eligibility, N06 immutable correction provenance and clean separate-match starts.
- [x] **D05** Add PostgreSQL migrations for snapshot, members, command results, journal, projections and ledger. Gate: real-database migration/rollback fixtures retain pinned versions, exact rational values, unique game IDs, origin-game balances, charge/correction lineage, typed promises, phases, game-scoped decision records, lasting effects/bindings, active-Ace zones, physical availability, revisioned turn-bound proposal status/terms, draw-queue progress and resumable settlement progress.
- [x] **D06** Implement the match-row transaction boundary. Gate: competing commands cannot duplicate cards, scores or winners; snapshot, debts and events commit/roll back together; N05 rejects new wrong-game intents, and F03 current-state victory admission tolerates unrelated versions within that game but rejects changed eligibility; T01 proposal revision/acceptance/withdrawal/expiry races serialize and final transfer revalidation is atomic.
- [x] **D07** Add idempotency and unknown-commit reconciliation. Gate: connection loss before/after commit produces one effect; identical ID/body replays its original result, changed body conflicts, duplicate lookup precedes game-instance/version checks; an old committed command and delayed old-game first submission have distinct outcomes; duplicate accepted proposals/purchases never repeat transfers, alliance/privilege creation or draws.
- [x] **D08** Persist pending responses, effect decisions and randomness through restart. Gate: resume every response/decision stage, committed random result and exact settlement batch; stale decision inputs fail, Coup interrupts legal waits, and retries never repeat costs or restart already committed expensive work (F01/F02); retain T01 offer revision/accepted action and T02 queue position/counters/random draws through every restart boundary.
- [x] **D09** Complete board closure, financial finalization and next-game transitions with recoverable settlement. Gate: F03–F05 concurrent declarations yield one board ending; only specified settlement commands remain legal; finish resolves triggered promises; next game starts at zero cash after finalization with a fresh N05 game ID, including nullified replacement; N06 corrections, fixed match length and tied ranks hold; large-D recovery preserves exact progress without financial cutoffs.

Phase 1 persistence evidence (2026-09-30): `backend/internal/matchstore`
implements the five durable gates above. The complete real-PostgreSQL race suite
passed in 16.547 seconds, including migration/rollback constraints, atomic fault
rollback, concurrent victory/proposal commands, actual COMMIT wire interruption,
PostgreSQL SIGKILL/WAL recovery, thirteen restored pending-stage successors,
private-observation metamorphic checks, replay-journal corruption checks, exact
51-digit-denominator settlement progress, and completed/void next-game boundaries.
Full backend unit tests, vet, and game/bot/simulator replay regression race checks
passed; independent review found no remaining blockers. Evidence logs:
`phase1-final-pg-race--20260930T140315Z-2687063`,
`phase1-final-matchstore-vet--20260930T140340Z-2688436`,
`phase1-full-backend--20260930T135727Z-2674593`, and
`phase1-race-regression--20260930T135809Z-2677013` under `/tmp/agent-runs/`.
Reproduce with `python3 xops/postgres_integration.py -race ./internal/matchstore`.
The [adapter contract](../code/MODULE-game.md#durable-adapter-phase-1) explains
private/internal APIs and operational limitations. Public transport, identity,
opaque card handles and deployment remain Phase 2/later work. This implements
persistence correctness, not a bot-strength/balance result or production load claim.

Simulator/domain progress for D08–D09 also retains complete-game local replay.
See the linked coverage
record for separate targeted, autonomous and campaign evidence. The [measured
findings](../../sims/docs/gameplay-findings.md) report 18 three-player and 24
four-player finalized held-out games (three seed blocks each), plus separate
multi-game integration. The v3 investigation completed the eight outstanding historical
replays, measured serialization/menu bottlenecks, added tested policy candidates,
and completed three independent paired development blocks per population under
explicit larger budgets and three-game matches, plus separate retention and
bargaining comparisons. Review found v1 action-key collisions affecting attribution;
the v2 repair requires fresh evaluation. These development results do not confirm
strength or balance.
See the [prospective resource and inference plan](../../sims/experiments/gameplay-evidence-v3.md).
The 1,000-block targets remain unmet; P03 stays unchecked.

**Test plan:** rule regression matrix, deterministic replay, real PostgreSQL
transaction/fault tests, constraints and race checks on exercised paths.
**Exit:** every GAME_RULES transition is implemented with rule evidence;
committed state survives process failure and experimental/unsupported profiles
cannot masquerade as accepted-rule coverage.

## Phase 2 — Online API and multiplayer recovery

**Scope:** `phase-2`. **Dependencies:** Phase 1 for complete games; room/identity
contracts can be developed independently. **Goal:** confidential online play
with bounded requests and recoverable streams.

**Phase 2 execution plan (2026-09-30).** Goal: authenticated three/four-player
HTTP and WebSocket play with recoverable commands and confidential projections.
Non-goals: public deployment, monetization/allowance policy, Flutter integration,
or expanding P03 experiment allocations. Files: `internal/identity`, `rooms`,
`wire`, `delivery`, `presence`, `transport`, `cmd/server`, additive matchstore
adapters/migrations, protocol schemas and existing docs/test harnesses.
Tests: real PostgreSQL/Redis, authenticated multi-client journeys, malformed and
forged inputs, private-information metamorphism, command/cursor replay, arbitration,
slow readers, bounded overload, cancellation and shutdown under Go race detection.
Checklist: implement O01 identity/retention decisions, O02 atomic rooms/start,
O03 typed wire contracts, O04 ordered admission, O05 durable catch-up/streams,
O06 lifecycle bounds, O07 optional presence, O08 financial/reputation journeys.
Risks: account deletion versus indefinite waits, cross-seat handles, double-start,
lost wake-ups and unbounded workers. Keep pseudonymous shared match records,
derive actors from sessions, preserve PostgreSQL authority and bound all queues.

- [x] **O01** Decide sessions, guest migration and account deletion boundaries. Gate: invalid/expired credentials and cross-account access fail; retention conflicts are resolved before schema finalization.
- [x] **O02** Add rooms, seats, invitations and match start. Gate: concurrent joins/start obey capacity and pin a visible immutable games-per-match count (default three); account-scoped room-creation identity prevents duplicate rooms after a lost response.
- [x] **O03** Implement typed HTTP/stream schemas and localized error codes. Gate: OpenAPI/golden fixtures match routes, phase allow-lists, required immutable game/related decision IDs, typed award conditions and same-game F03 admission; authorized calculation/correction fields explain N01/N04/N06/N07; T01 typed offer revisions/statuses, T02 queue progress and T03 explicit quantity/payment schemas match rules; malformed/oversize inputs and forged actor/card IDs reveal no private state.
- [x] **O04** Implement and verify Q11 digital arbitration/disconnection. Gate: authoritative declaration order, ordinary-action priority without delaying victory, Coup during pending responses/effect choices, N05 game-scoped current-state admission, N03 skipped ineligible seats, competing declarations and indefinitely preserved disconnected/hibernated state have deterministic tests.
- [x] **O05** Add seat snapshot/cursor, replay and authenticated WebSockets. Gate: connect-during-mutation, duplicate/reordered delivery, pending decisions/settlement and expired cursors converge to the same §2.13 safe view, including public counts/effect zones and known history without private binding/restriction-location leaks; clients reconcile old operations without retargeting across game IDs; T01 private offer terms/status and T02 ordered draw projections converge after restart.
- [x] **O06** Add bounded queues, socket loops, cancellation and shutdown. Gate: slow readers cannot stall matches; overload is explicit; goroutines return to baseline after reconnect/shutdown tests.
- [x] **O07** Add Redis presence/optional hints with bounded local fallback. Gate: Redis loss cannot change authoritative game state, allowances or command recovery.
- [x] **O08** Add trade/loan acceptance, Q12 agreements/reputation and scripted three/four-client games. Gate: T01 nonblocking proposals do not reserve cards or block End turn; revision/expiry/withdrawal/acceptance races and canceled transfers create no early alliance/privilege; successful exclusive loan alliances hold; accepted-promise/40%-boundary/exclusive reputation outcomes match Q12 and remain voluntary transfers; F05 pay/offer/accept/reject/refuse/finish commands work after board closure, retain origin-game attribution and block early next-game; API and engine scenario traces match.

Phase 2 evidence (2026-09-30): 75/75 top-level tests passed in the independently
run six-package PostgreSQL/Redis race suite, zero skipped and no race reports.
The run covers real database rollback/crash/COMMIT faults, concurrent migration,
identity/session deletion and guest upgrade, atomic room creation/start, three/four-
client trade/loan/settlement scenarios, exact 40% reputation boundaries, private
capabilities, durable event gaps and TCP WebSocket reconnect, Redis kill/restart,
slow readers, admission overload and concurrent lifecycle shutdown.
`phase2-independent-captured-race.full.log` under `/tmp/agent-runs/` retains the
full output; `phase2-independent-captured-race--20260930T153326Z-2798539`
records the successful run. A preceding run had 73 passes and two failures:
asynchronous authority-lock release and reputation DDL outside the migration
transaction. Both causes were fixed and the complete suite rerun. The repaired
migration is checksum-pinned version 2 under the original transaction/lock.
A separate real-process server smoke test passed under the race detector in
3.327 seconds (`server-smoke-scoped-race--20260930T153555Z-2804898`), proving
TCP startup, duplicate-authority rejection, SIGTERM shutdown and preserved
identity/handle key across restart. Full backend tests and vet passed
(`phase2-full-backend--20260930T153052Z-2791710`); final wire/transport race and
vet passed (`phase2-final-vet--20260930T153527Z-2803147`).
One intervening verbose capture run was discarded because the system `tee`
crashed and truncated evidence. The successful 75-test run writes its full log
directly; `safe-run.sh` now rejects capture failures, with three regression tests.
The [API contract](../code/API-game.md) records auth/retention decisions, limits,
wire schemas, run instructions and exact verification scope. This is targeted
local scenario evidence, not production capacity, human usability or P03 bot
strength/balance confirmation. Flutter, economy and deployment gates remain open.

**Test plan:** real PostgreSQL/Redis integration, authorization/contracts,
Go race detector, fault injection and scripted multi-client journeys.
**Exit:** retries/reconnects cannot duplicate effects or expose concealed
information, and identity/online admission decisions are recorded.

## Phase 3 — Offline and product integration

**Scope:** `phase-3`. **Dependencies:** Phase 0 fixtures, Phase 2 API.
**Goal:** develop and test the adaptive Flutter web client and product contracts,
while progressing host-based offline research without trusting local online outcomes.
P04 can start now from the existing host and Phase 2 API, independently of P01,
P02, P03, native devices and iOS infrastructure. P01 informs P02's local engine
adapter; P03 retains its separate resource-allocation constraint. P05/P06 contracts
and Phase 4 can proceed without native verification. Production web gameplay
still requires online authority; local demo/research fixtures are not offline web play.
Only ROADMAP carries completion checkboxes; baseline and integration criteria are
acceptance specifications, not additional sequenced checklists.

**P01 historical execution slice (2026-09-30, before sequencing revision).** Goal: make the native candidate measurable
through a stateless C ABI and Dart fixture runner before selecting an offline
engine. Non-goals: choosing the engine, a complete Dart port, changing UI v1,
claiming mobile acceptance or expanding P03's resource allocation. Touched paths:
`backend/internal/offlineprobe`, `backend/cmd/offlineffi`,
`client/offline_probe`, `xops/offline_probe.py` and
existing documentation. Steps: add failing transition/privacy/invalid-input
tests; implement the bounded probe and caller-owned C buffers; compare Go and
Dart FFI results; build Android ARM64; record reproducible evidence and gaps.
Test plan: Go race/vet, byte-exact state/event/seat-observation parity, malformed
requests, repeated allocation/free, isolated-process restart and host timings.
Risks: omniscient research fixtures must never become online or UI payloads;
host measurements cannot establish Android/iOS lifecycle or device budgets.
No connected Android device was found; iOS measurements are unavailable on this
Linux host. The subsequent host comparison below closes P01; target-device gates belong to R06N.

P01 retained earlier evidence: the [native comparison probe](../../client/offline_probe/README.md)
passes synthetic three/four-player board state, event, private-observation and
narrow heuristic parity through Dart FFI, including malformed buffers, pending
resume, repeated calls and concurrent isolates. Android ARM64/API 23 compilation
passes. The independent Dart financial candidate now compares exact ledgers,
seat-limited financial observations, totals/ranks and finish decisions with native
Go across three finalized financial games for both populations, reciprocal FIFO
thirds, large rationals, forgiveness/Coup and nullification; every-frame JSON
restart and fresh-process/isolate checks cover these synthetic scenarios. This is
not a dealt-and-played card match, complete Dart board/bot engine, promise journey
or production save system. Android execution, APK packaging, iOS and full-gameplay
Dart/native comparison were unverified at that earlier checkpoint. Those limited
fixtures alone did not establish an engine choice.

P01 current evidence (2026-09-30): an independent Dart card-play candidate and
Go/native now complete natural three-game matches for three and four seats,
covering six games, 1,233 exact full-state/event/seat-observation frames and 1,231
restart successors in each of two cohorts. Independent review and verification
replayed all boundaries and passed Go race/vet, Dart analysis and four runner
tests. Four isolated timing processes and memory results are recorded in the
[probe README](../../client/offline_probe/README.md).
This supports a provisional shared-Go recommendation. The independent candidate
explicitly supports a restricted passive card-play policy, not complete Dart
combat/ability/promise rules or competent bots. Final device measurements and
packaging remain R06N; P02 owns production saves/tutorials. P03's existing
allocation blocker remains unchanged.

- [x] **P01** Compare independent Dart engine/bots with Go/native FFI using host-executable full-game fixtures. Gate: exact state/event/seat-observation parity, complete three/four-player card-play matches, restart and documented host memory/decision-latency comparison support a provisional engine recommendation with explicit limits. Existing bounded evidence is retained; final engine approval, packaging and Android/iOS device measurements move to R06N.
- [x] **P02** Add offline saved games, difficulty tiers and beginner tutorial. Gate: host tests prove durable save/resume, network-independent local engine/bots and tutorial journeys; lowest-tier bots remain legal/competent; web play still requires online authority. Real-device airplane-mode, lifecycle and save acceptance remain required in R06N.
- [ ] **P03** Evaluate stronger bots against baselines. Gate: held-out seeds, rotated seats, separate three/four-player estimates, confidence intervals and decision budgets; optional LLM research first inventories hardware and reports nondeterminism.
- [x] **P04** Build distinct desktop web, phone and tablet presentations and integrate full Flutter rules journeys within [UI baseline v2](../design/UI_BASELINE.md), using cgmsart assets and ARB localization. Gate: purpose-built composition, navigation, interactions and components for each class; phone/tablet web appearance and journeys faithfully specify eventual native apps; the baseline breakpoint/viewport/transition matrix, browser keyboard/screen-reader, touch, RTL, text-scale, placeholder-parity and shared scenario journeys pass; show decision owner/stage, legal costs/quotas, lasting effect owner, known versus hidden cards, separate score/cash/debt, pending settlement and declaration/score/rank under F01–F10; N01–N07 add explainable calculations, effect custody/availability, eligible response prompts, Code economics, game-transition operation reconciliation and distinct match corrections; T01 proposal versus accepted-action states, T02 authoritative draw order and T03 explicit purchase quantity/excess are exercised; reuse `client/` and its `packages/cgms_ui`, sharing state, gameplay logic and suitable primitives without requiring one shared screen. Native accessibility/platform acceptance belongs to R06N; existing demo evidence alone cannot close P04.

P04 Design Partner refinement on 2026-10-01 uses the primary mockup layout and
cgmsart-decks artwork together within the adaptive baseline. The
[redesign report](../reports/2026-10-01-ui-redesign.md) records the implemented
entry/lobby, public groups, hand gallery, action review, shared surfaces and focus
repairs, with comparable before/after evidence and exact verification limits.
This is a frontend refinement of the existing checked gate; native, production,
store and other unchecked acceptance gates retain their existing status.

P04 independently reverified after the earlier follow-up repairs on 2026-10-01:
[retained verification summary](../reports/2026-10-01-nonproduction-verification.md)
records **396 baseline and 24 economy passing cases**, 19 reports and
145 final screenshots across Chromium, Firefox and WebKit. Nine class/engine
full journeys, three adaptive matrices, a full three-seat variation and three
strict-CSP checks use one final build/harness; 163 host/50 package tests and both
analyzers pass. The [follow-up report](../reports/2026-10-01-nonproduction-verification.md#independent-follow-up--nonprod-recheck-20261001)
records backend race/database checks, preserved failed/superseded browser evidence,
UI regressions and native/production limits. Earlier relocation results remain
historical results in the report; the original screenshot-directory manifests
and logs were deleted under the owner's current cleanup instruction.

Historical P04 acceptance (2026-09-30): [accepted online client](../../client/README.md#accepted-online-verification)
and [retained historical summary](../reports/2026-10-01-nonproduction-verification.md).
Nine full engine/class runs, three adaptive matrices, a three-seat variation and
strict-CSP asset rendering passed 271 browser case executions on one release build;
129 host and 50 package tests passed independent verification. Synthetic rule
boundaries complement normal room/deal/reload; native acceptance remains R06N.

- [x] **P05** Decide/reset/test online economy boundaries. Gate: day timezone, charged game starts, eligible completion, abandoned-game allowance/reward handling under untimed waits, dirt awards, pass stacking and refunds have explicit examples; atomic allowance reservation and unique reward IDs withstand concurrency; offline earns no dirt.
- [x] **P06** Research current official store policies and implement applicable Ad-free/Daily/Weekly/Monthly/Yearly and account contracts under decision 0015. Gate: dated Apple/Google applicability matrix before SDK selection; free caps, paid access and earned Daily/Weekly access have eligibility, atomic redemption, expiry, restore, duplicate-notification and ambiguous-provider-result contract tests with provider fixtures; historical receipts remain recoverable, new cosmetics/promotions are excluded; real native SDK/store sandbox, signing and device acceptance are deferred to R06N/R06; product durations, benefits and dirt prices are explicit validated decisions.

P06 approved-term implementation (completed 2026-10-01, decision 0018): apply 30-day
standalone Ad-free and non-renewing 24-hour paid Daily access anchored to confirmed
purchase; verify unlimited starts and ad removal for every paid duration. Change
the economy provider contract, additive migration, account copy/ARB, QA fixtures,
browser acceptance and active product/API docs. Sequence: add failing contract
tests; implement and preserve historical receipt recovery; verify all paid tiers
and unchanged earned access; run independent review, final integrated build and
three-engine account journeys; then close only P06 if its whole local gate passes.
Tests include exact durations, expiry/revocation/duplicates/restore, allowance
charging, private projections and phone/tablet/desktop keyboard/text-scale checks.
Preserve existing staged work, game rules, artwork, offline tools and all prior
SQL checksums. P03, Phase 5, live payment/ad SDKs, prices, other renewal rules and
native/store acceptance remain excluded.

Prior P06 local continuation (2026-10-01, before decision 0018): finish the remaining non-bot,
non-production product contracts. Resolve paid Daily benefits and the exact
fixed term of standalone Ad-free with the owner; do not infer prices or renewal
rules. Refresh official Apple/Google applicability before selecting any SDK.
Touched areas: economy provider/account contracts and migrations, their tests,
the existing account UI/ARB where required, API/product documentation and dated
verification evidence. Sequence: record the decisions; add failing eligibility,
expiry, restore/revocation and benefit-separation tests; implement the narrow
contracts; verify account privacy and recovery with the real local authority;
review and stage. Verification uses affected Go/PostgreSQL race tests, Flutter
tests/analysis/build if UI changes, and rendered phone/tablet/desktop account
checks. Risk controls: preserve historical receipts and migration checksums,
never turn synthetic provider fixtures into live verification, and leave the
gate unchecked until every applicable local criterion passes. P03, Phase 5,
live payments, native release and production changes are excluded.

Prior P06 continuation evidence: standalone fixed-term Ad-free gained a verified
absolute-expiry contract independent of online-start allowance, a separate
private account expiry, and duplicate/restore/revocation/migration regressions.
The client rejects malformed account snapshots while preserving confirmed
state and receipt recovery. [The dated local report](../reports/2026-10-01-p06-local-contracts.md)
records those executed checks and the refreshed official policy matrix. At that
handoff the exact Ad-free term and paid Daily benefits were still pending;
decision 0018 resolves both. The [approved-term report](../reports/2026-10-01-approved-paid-access.md)
records the completed local gate: 268 Flutter tests, both analyses, release build,
economy/HTTP/fixture PostgreSQL race suites and 72 final-build browser cases
across Chromium, Firefox and WebKit, with 162 screenshots. P06 is checked; the
roadmap is **39/46**, with P03 and six Phase 5 gates remaining. Provider fixtures
do not establish native/store acceptance or longer-tier calendar-period mapping.

P05 fresh local acceptance: decision 0014, real PostgreSQL/Redis concurrency and
race suites, and [24 economy browser cases across three engines](../reports/2026-10-01-nonproduction-verification.md)
pass on the relocated final build. Account privacy, explicit spending, uncertain
receipt recovery, no duplicate charge, once-only rewards and streamed allowance
wait/pass-based next-game admission are exercised. Synthetic initial funding is
labelled; actual finalized game rewards and pass operations use the real authority.
Offline earns no dirt. No SDK, native/store sandbox or live commerce is implied.

2026-10-01 remaining decisions: **P03** stays unchecked because the owner
explicitly kept compute allocation blocked; neither 300 hours/230 GiB nor
100 hours/90 GiB is authorized. The 67 bounded simulator preparation tests passed,
with zero new evaluation samples and no new strength claim. **P06** terms are now
approved by decision 0018 and their local implementation is verified. The
existing approved 30-dirt daily rate is preserved (Daily 30, Weekly 210). Cosmetic sales/unlocks and new promotional campaigns
are explicitly excluded by decision 0015; they are no longer missing catalogue
or channel decisions. Earned Daily/Weekly retain ads unless a separate ad-free
entitlement is active. Decisions 0015–0016 authorize the current product/style
implementation without implying any native/store or production acceptance.
The [economy design](../design/DESIGN-economy-and-stores.md) records historical
store-policy research and remaining paid-product boundaries. Preserve the earlier
[nonproduction verification report](../reports/2026-10-01-nonproduction-verification.md)
as evidence of the policy tested then, not proof of the revised behavior.


P02 host implementation in [client/offline_host](../../client/offline_host/README.md)
passed independent review and verification: durable fsync/rename saves, exclusive
ownership, corruption/version rejection, authorized projections, uncertain-save
recovery, real process restart/SIGKILL and a persisted engine-backed first-turn
tutorial. Cold race tests completed natural three/four-player games in 544/1,017
transitions. Three legal heuristic tiers are implemented; calibrated ordinal
strength is not claimed. The host bridge is not native packaging or an offline
Flutter screen, and web remains online-only. The independent 2026-10-01 follow-up
reopened P02 after a random dense-board budget failure; impossible Baron/spent-Ace
candidates are now pruned without changing the budget or legal menu. A deterministic
regression and the complete final Go race suite, including three/four-player
offline journeys, passed; see the follow-up verification report for exact logs
and the unavailable original-checkpoint limitation.

**Test plan:** host cross-engine differential tests, adaptive Flutter web browser
journeys and economy/provider-contract fault tests. Native offline/device and real
store sandbox checks are mandatory final-stage R06N/R06 work.
**Exit:** online authority remains independent of local results; paid/free
boundaries have product decisions and test evidence. LLMs remain optional.

## Phase 4 — Observability and operational workflow

**Scope:** `phase-4`. **Dependencies:** Phase 2; required before release.
**Goal:** predictable inspection, configuration and developer operations.

- [x] **V01** Add safe logs/traces/metrics around request, transaction, engine and delivery boundaries. Gate: no private hands/seeds/tokens; queue/lock waits and unknown commits are observable; Flutter support/overhead is verified before choosing its instrumentation.
- [x] **V02** Containerize the OTel collector/storage/UI stack behind private access. Gate: one telemetry setting drives exporter and Compose profile behavior; enabled/disabled tests preserve logs/readiness without repeated failed exports.
- [x] **V03** Add service configuration schemas, overlays, secret references and health checks. Gate: invalid config fails startup; Compose provides wiring, service values live in dedicated files and credentials stay private.
- [x] **V04** Add Nginx web/API/WebSocket routes and separate local/production profiles. Gate: hot reload/debug/source maps, Flutter assets/workers, CSP/CORS/HSTS and TLS are tested against named criteria, not an undefined security grade.
- [x] **V05** Add thin `web.re`, `go.re` and service restart targets using `xops/makefile/`. Gate: readiness waits are bounded, failure logs survive, browser cache refresh is documented, repeats are safe and persistent volumes remain intact.

Current operational evidence is consolidated in [deploy/VALIDATION.md](../../deploy/VALIDATION.md).
Independent review and verification passed backend privacy/telemetry, strict
configuration, private enabled/disabled services and bounded restart behavior;
real collector recreations also passed. V01's Flutter evidence is an isolated
web-only candidate spike, with no production SDK selection or native overhead
claim. V04's named CanvasKit criteria pass: same-origin assets and font MIME,
exact-CSP rendering, trusted TLS/headers/origin checks, authenticated WebSockets
and development reload/static-profile recovery. Emitted worker files are routed;
unused worker-renderer execution is not claimed. The observed transient debug
hot-restart assertion remains documented in the validation report.

**Test plan:** container/configuration negatives, redaction, telemetry disable
and restart/failure drills. **Exit:** instrumentation switches require no code
change, and all documented operational commands actually exist and pass gates.

## Phase 5 — Production v1

**Scope:** `phase-5`. **Dependencies:** Phases 1–4 and release-critical decisions.
**Goal:** a verified single-host Docker Compose release with known limits.

- [ ] **R01** Produce pinned images and production Compose settings. Gate: clean-host rehearsal uses explicit service files/secrets, least privilege and private database/admin networks; versions/licenses are recorded.
- [x] **R02** Load-test proposed capacity/latency/memory budgets. Gate: publish hardware/workload, p50/p95/p99, queue/DB saturation and telemetry overhead; revise unachievable targets explicitly.
- [ ] **R03** Test overload, Redis/telemetry loss, database outage and process/host restart. Gate: no false success, duplicate effects or unbounded queues; state exactly which tested failures preserve acknowledged writes.
- [ ] **R04** Run backup and timed restore drills. Gate: coherent snapshot/ledger/config/command-result recovery meets agreed RPO/RTO; report possible acknowledged-data loss under host/storage failure.
- [x] **R05** Rehearse migration/rollback with active pinned matches. Gate: old/new readers are compatible, partial backfills resume, original engines remain available and unsupported states are never silently reinterpreted.
- [ ] **R06N** Complete deferred Android and iOS verification after adaptive web/product integration and operational work, before R06/R07 release approval. Gate: package and execute both offline engine candidates under matched workloads on representative real Android/iOS phones and tablets; record OS/device/toolchain versions, memory and decision-latency budgets/results, then confirm or revise P01's provisional engine choice and rerun affected parity tests. Prove signed packaging/install/upgrade, airplane-mode bot play/save/restore, background/foreground, suspension/process death, reconnection and uncertain-command recovery; verify safe areas, orientation, system back/navigation, keyboard, touch/gestures, native screen readers, rendering/frame performance and platform/plugin behavior. Compare native screenshots and complete journeys against accepted P04 phone/tablet web references; justify platform-specific differences without changing rules/privacy. Exercise actual purchase/restore/provider sandbox flows and required platform acceptance. Missing devices or iOS build/signing infrastructure block this gate and release, never earlier web development; compilation, emulation or browser evidence alone cannot close it.
- [ ] **R06** Complete baseline security/privacy/store release verification and operator guides. Gate: R06N passes; recheck dated official policies, declarations, deletion, purchases and mobile signing with actual evidence; review optional polish without delaying core correctness work.
- [ ] **R07** Run full v1 acceptance and release rehearsal. Gate: R06N and R06 pass; rule coverage, simulator/backend/client parity, reproducible builds, recovery evidence and known limits are archived; no readiness claim rests only on this plan.

R02's published baseline combines 100 four-seat matches, 400 live subscriptions
and real opening/offer/transfer/response/turn operations. Both telemetry cohorts
complete the 20 and 80 commands/second schedules; 200 is explicitly rejected as
a capacity target after pool/admission saturation. Timings include drain and
report retries, queue/transaction tails and RSS. Sequential shared-host samples
do not establish causal instrumentation overhead or production SLAs. Independent
short-fixture verification passes. R03's retained-storage process restart and
individual outages pass, but an actual host restart remains unverified; R01's
fresh-workspace rehearsal does not substitute for a fresh host.

**Test plan:** full regression/integration/client suites, sustained load,
fault injection, clean Compose deploy, restore/rollback and submission evidence.
**Exit:** accepted targets and required gates pass; the operator understands
single-host availability limitations. Kubernetes/Helm remain post-v1.

## Risks and execution discipline

Primary risks: incomplete rule implementation, exact debt cascade cost, online ordering,
private-data leakage and Go/offline-engine divergence. All F01–F10, N01–N07 and T01–T03 recommendations have been adopted;
new findings from further reviews must be recorded distinctly, never presented
as missing user answers to those decisions. Use the rule register,
deterministic fixtures, atomic effects and measured capacity gates. A passing
test for an invented rule is not conformance.

For behavior slices, use red-green-refactor and rerun affected tests. Roadmap
items describe outcomes; break larger items into bounded working steps without
creating a competing authoritative checklist. Review and verify before checking
an item; update snapshot counts immediately when its whole gate passes.
Only the coordinating parent tracks/stages; humans commit and push.
