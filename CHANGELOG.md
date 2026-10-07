# Changelog

Changes to the CGMS rules are recorded here. The current working reference is
[GAME_RULES.md](docs/project/GAME_RULES.md); the
[third draft](docs/design/drafts/THIRD_DRAFT.md),
[second draft](docs/design/drafts/SECOND_DRAFT.md) and
[first draft](docs/design/drafts/FIRST_DRAFT.md) retain their historical rules.
These decisions supersede conflicting recommendations in the historical
[review](docs/reports/CGMS-game-review.md) and the removed `suggestions.md` document.

## [Unreleased]

### Changed — direct drops and scheduled draw feedback

- A deliberate complete, unambiguous legal card drop commits one operation after
  showing its exact cards, costs and destination. Contextual choices collect only
  genuinely missing inputs; board targeting no longer repeats the destination.
- Give the face-down draw pile safe space between public seat rows, keeping
  seat text, controls, public card sizes and its accessible target intact.
- Present newly confirmed scheduled-turn draws once with destination-aware,
  privacy-safe themed feedback and a reduced-motion highlight. Draws remain
  server-owned; refresh, reconnect, trades and loans do not replay feedback.
- Treat intact public formation drops onto another player as private loan
  proposals. Exact revision acceptance and successful response resolution
  retain exclusive alliance and atomic transfer rules.
- Keep hand-card positions stable through selection, fit exact large-bundle
  previews at enlarged text sizes, and preserve returned Doppelganger roles
  using private authorized binding metadata.

### Changed — cards as the gameplay interface

- Remove the permanent online action/composer/review workspace and its reserved
  space. Single tap selects physical cards, double tap or Enter activates their
  flow, and drag/drop or taps choose concrete board destinations.
- Keep inspection independent through the magnifier or I. Show source, target,
  payment and modifier roles; use only temporary move-specific choices and
  confirmation, retaining exact physical copies and server validation.
- Cover combo opening/reconfiguration/activation/take-back, royal abilities,
  Aces, attacks/defences, ordered choices, chosen-quantity purchases, private
  revisioned trades, intact loans and voluntary returns through their cards.
- Keep immediate Coup reachable through eligible Queens while contexts are open;
  retain explicit Pass, End Turn, required decisions and financial utilities.
  Cancelled/stale gestures never submit or withdraw an accepted action.
- Keep whole-series destinations distinct from individual cards and attached
  royals. Disable purchase confirmation for a known exact payment shortfall,
  including Inflation fractions, while preserving server-only supply validation.
- Require an explicit attachment destination and keep deliberate unused-Ace
  disposal behind its named choice and confirmation at the server's legal boundaries.
- Reclaim board/hand space, preserve centered suit groups and readable cgmsart,
  and support live resizing, independent focus/inspection and enlarged scrolling.

### Changed — online moves from cards

- Select cards or drag them onto the table, another player, the hand or draw pile
  to choose a contextual move, review its costs and targets, and confirm it.
- Replace the generic action dropdown with card-based choices, explicit combo
  roles and protection, purchase quantities, Ace modes and ordered responses;
  start trades from the cards and keep financial decisions and ordinary victory
  declarations in game options and records.
- Retain readable public cards and the adjacent hand while reviewing a move,
  with keyboard/tap alternatives and checks against stale or unauthorized cards.
  Hidden identities and final legality remain controlled by the server.
- Refresh missing response context in saved matches without advancing play,
  retain older journal receipts, and expose the final Kidnapper card choice.

### Fixed — public card readability during move review

- Preserve readable public-card previews and a complete first hand card at
  ordinary tablet/desktop heights during move review; bound and scroll temporary
  contexts instead of collapsing the table into tiny thumbnails.

### Fixed — visible draw pile and balanced table layout

- Show the face-down draw-pile marker using the shared card back, without exposing
  hidden identities or inventing a remaining-card count.
- Keep the public table visible while preparing an action, with the private hand
  immediately below it and temporary contexts for the selected move.
- Bring wide-screen hand suit groups together and center tablet/desktop compact
  utilities while preserving card selections, drafts and touch targets.

### Fixed — backend log flood

- Keep successful internal operations and expected cancellations out of default
  logs while retaining metrics/traces and visible warnings for failures,
  deadlines and unknown commits.
- Avoid repeatedly restoring unchanged matches that are waiting for a human or
  have finished; resume automatic work when the saved match version changes.

### Added — service lifecycle commands

- Add `make up`, `make down` and `make restart` for Nginx/Flutter web, backend,
  PostgreSQL, Redis and configured telemetry. Startup initializes missing secrets
  and builds Flutter/Go; restart rebuilds and recreates the stack. Shutdown
  retains containers and persistent volumes. Existing scoped commands remain.

### Fixed — compact online action workspace

- Remove the redundant action heading strip and use the game-header toggle to
  open or close the workspace.
- Arrange action controls in responsive rows with smaller text and icons;
  keep full touch targets and put extended rule guidance behind a disclosure.
- Let short action panels return unused space to the hand while longer forms
  retain bounded scrolling and their draft/focus state.

### Fixed — online typography and alignment

- Center the online lobby introduction and section headings, with smaller
  secondary headings and balanced explanatory text.
- Use regular body text in setup and gameplay dropdowns; allow enlarged or
  wrapped options to grow while retaining full touch targets and draft choices.
- Keep display text-size and reading-direction settings in open dropdowns;
  bound long menus and center the gameplay action form on larger screens.

### Added — online bot matches

- Add online matches against two or three server-controlled bots, with existing
  difficulty policies, private observations and durable turn/response recovery.
  The lobby exposes bot setup while retaining human room creation.
- Show the human Stay/Leave decision at round boundaries instead of an apparent
  bot-turn stall, retaining drafts and guarding against duplicate submissions.
- Reset the decision scroll viewport at round-workspace transitions so web
  accessibility and touch targets remain aligned with the visible buttons.
- Keep ordinary play controls disabled during server setup, including initiative
  and turn initialization, and enable them automatically when play is ready.
- Exclude zero-growth purchases from online bot suggestions so an empty deck
  cannot cause the observed repeated buy-and-reopen turn; preserve human rules
  and existing offline policy behavior.
- Avoid spending the online worker deadline validating unchosen bot candidates;
  retain tied-choice compatibility and stop planning when cancelled.

### Fixed — compact public quadrants and visible hand

- Keep the compact header in one centered row with smaller visible typography
  and icons, full labels and unchanged touch targets; remove the forced status/tool
  split at near-tablet widths.

- Balance the top status and tools across the available width, and group bottom
  actions in an opaque responsive dock. Insets, gaps and control sizing adapt;
  compact layouts reflow while preserving touch targets and keyboard focus.
  Paired actions stay equally tall with wrapped labels, and accessible bounds
  stay aligned after text-size changes.

- Align the public table to the hand's full width, with four equal rectangular
  quadrants on larger screens. Public cards now share the hand's size scale
  where space permits; dense piles retain bounded overlap and complete artwork.
  This owner revision supersedes the square-only table constraint.
- Grow the public table and its card previews as the available width increases,
  including live tablet/desktop resizing, while retaining space for the adjacent
  hand and bottom controls. Short viewports still constrain the public area.
- Keep the public overview in four equal quadrants at every width,
  size previews to the available height, and reserve adjacent space for the hand.
  Public piles and detailed actions no longer push the hand down a long page.
- Remove the obsolete `client/screenshots` archive, including JSON and logs;
  retain only `cgmsart-phone.png` and compact verification reports under `docs/reports`.

### Fixed — private hand sizing and suit stacks

- Give every hand card the same frame size and horizontally overlap cards of
  the same suit to save space. Keep complete artwork, separate physical-card
  selection, keyboard/touch inspection and readable exposed targets; long suit
  groups scroll horizontally.

### Changed — shared phone-reference layout

- Adopt the owner's `client/screenshots/cgmsart-phone.png` layout at every
  screen width: compact round header, two-column public card stacks, action
  region, private hand and bottom actions. Tablet and desktop
  enlarge the same arrangement. This supersedes separate workspace panes and
  the earlier prohibition on overlapping public previews.
- Open public stacks in an accessible physical-card chooser; preserve private
  card inspection, action drafts, focus and pending commands across resizing.
  Isolate expandable form scroll state, dismiss card choosers when their board
  is removed, and validate their cards against the current authorized view.
- Keep focused controls visible after scrolling and resizing. Announce closed
  boards, open settlement/completion controls automatically and disable End Turn
  outside active play without dispatching commands.
- Remove every other screenshot, including historical captures, at the owner's
  request. Keep the reference byte-identical and retain lightweight reports,
  source mockups and card artwork. Disposable captures stay ignored.

### Fixed — normal-entry adaptation and startup recovery

- Route the normal Auto-selected demonstration through the shared adaptive
  table, with phone, tablet and desktop boundaries at 600 and 1050 logical
  pixels. Preserve physical-card selections and retained workspaces while
  changing composition, and keep controls usable with enlarged text.
- Recover session bootstrap and interrupted match reads automatically with
  bounded backoff and visible progress. Authentication and malformed successful
  responses fail clearly; uncertain commands retain their original identities
  and are never replayed by readiness recovery.
- Retire only superseded generated screenshot captures, preserve their
  lightweight outcomes and required visual references, and exclude future
  disposable captures from ordinary staging. Reachable Git history was
  inspected; the available evidence does not establish a push rejection cause.
  See the [repair report](docs/reports/2026-10-02-adaptive-startup-recovery.md)
  and [retention inventory](docs/reports/2026-10-02-screenshot-retention.md).

### Fixed — full card artwork and deliberate magnification

- Match card image boxes to each painting's native proportions, showing the
  complete artwork without stretching, cropping, opaque labels or overlapping
  piles. Keep small rank/suit identifiers outside the painting and hide visible
  copy numbers and raw zone labels, including in inspectors.
- Single tap selects; double-tap magnifies without changing selection. Keep
  keyboard I and a small hollow corner magnifier with an accessible 48px target.
  Long press and rejected selections no longer open an inspector.
- Keep inspection keyboard-reachable after returning to retained phone panes
  and resizing, and preserve authorized physical identities and status.

### Changed — approved paid access terms

- Set standalone Ad-free to 30 days without removing the online-start cap, and
  paid Daily to a non-renewing 24-hour unlimited/ad-free pass from confirmed
  purchase. Paid Weekly, Monthly and Yearly also provide unlimited starts and
  remove ads. Earned Daily/Weekly continue to retain ads unless separate ad-free
  access is active.
- Validate fixed-term confirmation and expiry through the trusted provider
  contract, preserve historical receipts, and prevent restores/notifications
  from extending access. Added paid-tier allowance, expiry, revocation and
  privacy regressions and localized account explanations. Live checkout and
  provider SDK integration remain unavailable. See the
  [approved paid-access report](docs/reports/2026-10-01-approved-paid-access.md)
  for final executed coverage and limitations.

### Changed — fixed-term Ad-free account contracts

- Added the local verified-provider contract for fixed-term standalone Ad-free,
  preserving the normal free-start cap and keeping its expiry independent of
  earned Daily/Weekly access. Restores and duplicate notifications do not extend
  access; revocation preserves separately valid benefits. The new migration
  preserves both earlier SQL migrations and historical receipts.
- Added a separate private ad-free expiry to the account contract and refreshed
  official Apple/Google policy applicability. Decision 0018 subsequently resolves
  the Ad-free duration and paid Daily benefits; live checkout stays unavailable.
- Reject malformed account snapshots without losing the last confirmed view.
  New spending waits for a successful refresh, while uncertain receipt recovery
  remains available. Disabled account payloads no longer crash the allowance
  dialog. See the [P06 local report](docs/reports/2026-10-01-p06-local-contracts.md)
  for that slice's executed checks and then-unresolved decisions.

### Changed — shared match styles and earned access

- Removed cosmetic sales/unlocks from the active product model. The only premium
  products are Ad-free, Daily, Weekly, Monthly and Yearly; only Daily/Weekly can
  be redeemed with earned dirt. Earned access retains ads unless a separate
  ad-free benefit is active. Unspecified paid-product terms remain unavailable.
- Replaced new one–six-day redemption choices with Daily/Weekly. Preserved
  historical receipt recovery and original promised ad-free expiry without
  extending it through new currency redemptions; retired cosmetic grants cannot
  be newly issued. The additive migration rejects older policy readers.
- Assigned the four original card-art styles distinctly across players for an
  entire match, consistently for every authorized viewer. Captured cards use
  their new controller's style. All 208 original illustrations are bundled;
  card identities, rules, live labels and the shared concealed back are unchanged.
- Reconciled open inspectors with the latest authorized projection so transferred
  cards cannot retain stale appearance or no-longer-visible faces.
- Stopped inactive local-demo artwork loading behind online views while retaining
  local routes, selections and artwork preferences across navigation.
- Updated active product, API, art and roadmap documentation under superseding
  decisions 0015–0016. Earlier reports retain their original tested policy.
  The [implementation report](docs/reports/2026-10-01-shared-styles-and-products.md)
  records fresh checks, failures and remaining product/release boundaries.

### Changed — cgmsart adaptive interface refinement

- Applied the mockups' layout and interaction direction together with the
  cgmsart-decks illustrations in the existing Flutter client. Entry and room
  setup now use illustrated introductions, separate create/join panels and a
  visible seat roster; allowance details remain available after the room tasks.
- Grouped public cards by their supplied formations and suits, replaced the
  phone's long hand list with an illustrated gallery, and clarified turn and
  decision context. Phone navigation, tablet companions and desktop workspaces
  retain selections, form drafts and inspection behavior across size changes.
- Unified opaque panels, control shapes, spacing and semantic colors. Action
  menus have named families; inspection supports keyboard and double-tap;
  navigation and dialogs retain usable focus. Session requests reject duplicate
  taps, expected guest entry avoids a false error, and invalid numeric drafts
  keep their contents with validation feedback.
- Kept action costs before submission, moved actionable trade/settlement records
  ahead of general quota context, and omitted absent declarers from ending labels.
- Retained the rules, authority checks, private projections, artwork and offline
  tools. The [redesign verification report](docs/reports/2026-10-01-ui-redesign.md)
  separates rendered browser evidence from native and assistive-technology gates.

### Fixed — independent non-production follow-up

- Re-rank competing ordinary commands when an active player's intent arrives
  during database ranking, while preserving victory priority. Database failures
  during membership checks now report unavailable instead of falsely rejecting
  the player's authorization.
- Prune impossible Baron and already-used Ace candidates from offline bot menus,
  preserving the legal choices and existing computation budgets on dense boards.
- Restore authorized owner/member lobbies after reload using public room IDs;
  keep invitation tokens out of routes. Reset card-selection destinations when
  changing actions or removing a Doppelganger substitution.
- Explain selected-action timing, committed costs, resolution costs and
  cancellation through localized text, and identify own Clubs already used this
  turn from the authority's card projection.
- Correct stale integration documentation and the local first-start command
  sequence for the app directly under `client/`.

### Changed — relocated client and local economy integration

- Moved the complete Flutter host directly to `client/`, preserving `cgms_ui`,
  offline tools, cgmsart and historical evidence; updated local build, preview,
  browser, Compose and documentation paths.
- Adopted 10 dirt per eligible finalized online game and 30 dirt per pass day
  (up to six days per redemption). Added private allowance/balance views,
  explicit pass spending and durable receipt recovery after lost replies.
  Offline games earn no dirt; store purchases and promotional entry points remain disabled.
- Fixed victory priority when an ordinary command is being ranked and overlapping
  economy room-start deadlocks. Waiting matches now retry admission without
  rapid repeated writes and resume after access changes or the UTC reset.
- Made room ownership/fullness and decision timing visible, retained legal
  out-of-turn victory declarations, and exposed recovery for uncertain starts.
  Room/invitation codes now expose their values to screen readers and have
  labelled Copy controls instead of empty unnamed selectable fields on web.
  Short rotated views and large text now scroll recovery/status content.
  Clipboard denial has specific recovery feedback; account dialogs preserve
  48-pixel controls, and startup Retry works under the strict content security policy.
  Transparent accessibility labels no longer intercept adjacent RTL toolbar controls.
- Distinguished settled games awaiting admission from completed matches, refreshed
  allowance status after streamed finalization, removed stale turn ownership after
  board closure, and hid absent entitlement dates.
  Fresh browser evidence is recorded separately from the original acceptance build.

### Added — adaptive authoritative online play

- Added distinct phone, tablet and desktop online workspaces with cgmsart,
  accessible physical-card inspection, touch controls, large text and RTL layout.
- Connected room/deal, ordered decisions, combat, private trade and loans,
  purchases, Coup, financial settlement and standings to the game authority.
- Preserved original commands across uncertain replies and reloads, while showing
  authorized costs, quotas, effects, scores, cash and debts separately.
- Bundled suit-symbol fallback fonts so the online client works under the
  production content security policy without fetching remote fonts.
- Retained the local demo as a separate presentation reference; native-device
  and store acceptance remain final release gates.

### Added — local operations and recovery evidence

- Added private optional telemetry, strict service configuration, pinned Compose
  images, separate Nginx development/TLS profiles and bounded restart commands.
  The supported Alpine backend image passes an isolated production rehearsal.
- Added private snapshot/restore and archived-reader rollback drills, including
  exact financial state and economy-schema rejection. Real host/storage loss,
  off-host recovery and agreed RPO/RTO acceptance remain open.
- Measured 100 four-seat matches with 400 live subscriptions and real gameplay
  operations. The 20/80 commands-per-second schedules complete; 200 saturates
  admission/database capacity and is explicitly rejected as a target. Published
  local measurements do not establish a production SLA or causal tracing cost.

### Added — online economy contracts and browser acceptance support

- Adopted UTC allowance reset, once-per-start charging, once-per-finalized-game
  rewards and extending ad-free dirt passes. Transactional contracts cover
  concurrent limits, rollback, duplicate recovery and pinned per-match policy;
  the initial slice left amounts and activation pending. The 2026-10-01 entry
  above records the subsequently adopted amounts and explicit local activation.
- Added seat-authorized quota status, purchase economics and public committed
  combat calculations, plus a loopback-only synthetic fixture server for real
  three/four-player browser interaction tests.
- Recorded dated Apple/Google store applicability research. Native purchase
  verification, promotional channels and release acceptance remain open.
- Added an internal hashed-code premium grant contract with atomic eligibility,
  expiry, redemption limits and duplicate recovery; public redemption is disabled.
- Added internal provider reconciliation and entitlement projection with synthetic
  restore/revocation fixtures, cross-account and sandbox isolation, and a distinct
  economy save schema that prevents older readers bypassing pinned policy.

### Added — verified provisional offline engine comparison

- Completed the P01 host gate with independent Dart/Go card-play comparison:
  six complete games, exact state/event/seat-observation parity and restart checks.
  Shared Go is the provisional recommendation; the Dart candidate deliberately
  covers a restricted policy, and final native measurements remain R06N.

### Added — durable offline host play

- Added a network-independent shared-Go local runtime with private atomic saves,
  guarded seat projections, three legal heuristic tiers and a persisted beginner
  tutorial. Host process restart/death and complete-game verification passed;
  native packaging, device lifecycle and calibrated bot strength remain deferred.

### Changed — web-first adaptive development plan

- Adopted UI baseline v2: distinct desktop, phone and tablet presentation in the
  existing Flutter web host/package, preserving cgmsart, rules and privacy.
- Separated P01 host comparison and P04 browser acceptance from final R06N native
  packaging, device, lifecycle, performance and platform verification before
  release. Existing evidence is retained; no frontend implementation is claimed.

### Added — offline engine comparison tooling

- Added an independent Dart financial comparison candidate and Go differential
  fixtures for three-game financial lifecycles, exact debt receipts, forgiveness,
  Coup and nullification. This extends P01 research evidence; full Dart card play,
  Android/iOS runtime acceptance and the offline engine choice remain open.
- Added a local Go/C/Dart comparison probe with synthetic board transition,
  private-observation and bounded baseline-bot parity checks, host measurements
  and optional Android ARM64 cross-compilation. This is P01 preparation;
  the offline engine choice, mobile runtime and full Dart rules parity remain open.

### Added — online multiplayer authority

- Added guest/account sessions, in-place guest upgrades, private invitations and
  atomic three/four-seat match starts with a visible immutable match length.
- Added typed HTTP commands, opaque seat-authorized card handles and authenticated
  cursor replay/WebSockets. Identical committed commands remain recoverable after
  visibility changes and game transitions; disconnection never plays for a human.
- Added bounded command/socket handling, optional Redis presence fallback, exact
  finalized reputation outcomes and private score-correction explanations.
  The local server survives restart with its identity and handle key preserved;
  client integration, monetization and public deployment remain open.

### Fixed — execution evidence capture

- A failed output-capture process now fails the safe-run wrapper even when the
  wrapped command succeeds, preventing truncated test logs from being marked green.

### Added — durable match authority

- Added internal PostgreSQL match persistence with atomic exact-ledger updates,
  seat-specific snapshots, immutable command receipts and recoverable pending work.
  Identical retries retain their original result across game boundaries and lost
  commit acknowledgements; server randomness and financial progress survive restart.
- Added bounded real-database migration, concurrency, network-fault, crash-recovery
  and privacy tests. The online transport now builds on this durable boundary.

### Added — accepted gameplay domain and simulator

- Extended the shared pure Go engine with full turn/ability/proposal lifecycles,
  exact financial finalization and fixed multi-game matches; added authorized
  gameplay policies, recorded complete-game replay/continuation and explicit
  incomplete outcomes. See [rule-to-test evidence](sims/docs/gameplay-coverage.md).
- Added frozen seed-block campaign execution and separate public versus restricted
  diagnostics. Development games demonstrate legitimate endings; held-out target
  attainment and precision are reported in [measured findings](sims/docs/gameplay-findings.md). No human-enjoyment, universal
  bot-strength, balance or production-persistence claim is implied.

### Changed — gameplay evaluation and development policies

- Reduced measured serialization and redundant menu-generation costs with fixed-input
  transcript equivalence checks; added explicit larger menu/transcript budgets for
  multi-game experiments while preserving historical defaults.
- Added opportunity and bargaining policies, controlled ablations, authorized-view
  privacy fixtures, frozen analysis-source provenance, and prospective seed-block
  strength/equivalence estimators. Historical replay completion and development
  results remain distinct from confirmation; see the [v3 plan](sims/experiments/gameplay-evidence-v3.md).

- Found a v1 formation-reconfiguration tie-key collision that confounds policy
  attribution; preserve its results and policies, and introduce an explicit v2
  ordering repair with regression tests. Corrected-policy strength remains unconfirmed.

### Added — bounded Phase 0 simulator

- Added one pinned Go module with pure physical-card transitions, exact resumable
  ledgers, strict configuration/hashes, fair baseline policies and the seven CLI
  operations. Versioned fixtures, private replay/provenance, incomplete-result
  reports, resource/cancellation checks and scratch-container packaging are covered
  by [implementation evidence](sims/docs/phase-0-implementation-evidence.md).
  Full games, stronger bots, balance and production readiness remain unclaimed.

### Added — Phase 0 preparation

- Added the [Astra simulator handoff](sims/README.md#phase-0-preparation-and-handoff):
  primary-source research, four repository-local skills, execution instructions,
  configuration/schema specifications, synthetic preparation examples and a full
  S01–S10 evidence guide. Generated simulation artifacts are ignored. No Go module,
  simulator, bot, container, measured result or runtime gate is delivered by this work.

### Changed

- Froze the owner-approved Flutter preview as [UI baseline v1](docs/design/UI_BASELINE.md),
  including the UI/UX and cgmsart theme. The runnable reference stays in
  `client/`; all twenty mockups remain supporting flow storyboards.
  Future fixes and integration follow the baseline's change policy. Full rules,
  backend/native integration, accessibility and production gates remain open.

### Removed

- Removed the superseded gameplay v1/v2 images and retired their layout guidance.
  Moved the historical concept study to deck source provenance, where its bytes
  remain unchanged. The deck library, twenty main mockups, licensing and
  game-rule history remain retained. Removed navigation into earlier demo
  layouts while preserving covered compatibility APIs and regression fixtures.

### Added

- Imported the 20 approved [mobile flow mockups](docs/design/mockup/README.md)
  with ordered filenames and source hashes. The Flutter adaptation uses compact
  public stacks and a complete private hand together across phone, tablet and
  desktop layouts.
- Renamed the illustrated identity, asset directories, references and derived
  sheet labels to `cgmsart`; native card illustrations and font licenses remain
  unchanged.

- The [Flutter web demo](client/README.md) now offers Auto, Phone,
  Tablet and Desktop table modes. Screen-size changes select distinct player
  workspaces while preserving cards, selected moves and pending responses.

### Third assessment — accepted finalization

- Published [GAME_RULES.md](docs/project/GAME_RULES.md) as the sole current
  gameplay authority and moved all four rule drafts into
  [docs/design/drafts/](docs/design/drafts/README.md), preserving historical
  bodies with authority notices and relative links updated.
- Completed the project [charter](docs/project/CHARTER.md),
  [decision log](docs/project/DECISION_LOG.md) and [glossary](docs/project/GLOSSARY.md)
  from the retained templates. Product requirements, proposed technical choices
  and unverified implementation gates remain distinct.
- Adopted every recommendation from the
  [third assessment](docs/reports/2026-09-28-final-draft-third-assessment.md):

| Decision | Adopted rule / contract |
|---|---|
| T01 — Offer lifecycle | Nonblocking revisioned trade/loan proposals reserve nothing and expire with their originating turn. Explicit decline/withdrawal races with acceptance; acceptance starts the acceptor's response-bearing action, and only successful atomic transfer grants an alliance or loan-derived privilege. |
| T02 — Compensation order | Deferred draws follow ascending fixed seat order, complete each recipient before the next, consume earned entitlement under exhaustion, preserve committed draws through restart and precede ordinary victory. Coup retains its atomic-boundary exception. |
| T03 — Purchase contract | Buyer chooses an affordable positive integer quantity using unrestricted number Spades from hand or exposed series; excess gives no change. Revalidate exact price and full supply after responses, then return/shuffle/draw atomically. |

All current contracts and scenario obligations use these accepted rulings;
the dated reports retain their original proposal wording as history. No runtime
roadmap task is completed by this documentation finalization.

### Second assessment — accepted interaction amendments

The user accepted all seven recommendations in the
[second assessment](docs/reports/2026-09-28-final-draft-second-assessment.md).
Its original snapshot remains historical. The
[third assessment](docs/reports/2026-09-28-final-draft-third-assessment.md)
reviewed this adoption; its T01–T03 recommendations are now adopted above.

| Decision | Adopted rule / contract |
|---|---|
| N01 — Combat algebra | Effective A/D comparison, numerical defense across number series, Hearts-only Advancement defense, frozen targets and physical Club exchanges distinct from virtual values. |
| N02 — Persistent Aces | Public effect-zone source/target/custodian/expiry, excluded from Fate count. Preserve effects/counters through Fate and Inflation source departure; target departure/board closure discards affected Aces. |
| N03 — Response eligibility | Skip confined/departed seats deterministically; retain confined Coup and eligible disconnected waits, with no inferred pass or timeout. |
| N04 — Code classification | Non-attacking global declaration applies price/bonus settings to every opponent. Hostile round penalty needs declarer support/two current series/no Confinement; exempt initially protected, confined and departed opponents. |
| N05 — Game identity | Require immutable game ID on every game-scoped command, including victory/settlement; authorized duplicate result first, then reject different-instance submissions. Replacements get fresh IDs; never retarget queued commands. |
| N06 — Forgiveness provenance | Preserve finalized results with linked nonspendable match corrections for cross-game forgiveness; survive Coup, reverse with creating-game nullification. Same-game correction follows its replaceable charge, preventing an extra Coup credit. |
| N07 — Restricted cards | Physical availability marker follows forced movement, Fate/shuffle/redraw until next round; forbid premature active use/voluntary transfer/payment. Ordinary possession counts and mandatory Diamond exposure remain allowed without clearing restriction. |

Rules, roadmap gates, technical contracts and frontend guidance are synchronized;
all 45 runtime tasks remain unchecked. This pass provides no runtime, rendered
UI, gameplay conformance or measured balance evidence.

### Added

- A [shared game engine module document](docs/code/MODULE-game.md) describing
  proposed boundaries, invariants and test obligations; it replaces the generic
  module template without claiming runtime implementation.
- A [second frontend/backend assessment](docs/reports/2026-09-28-final-draft-second-assessment.md)
  after applying the first review's recommendations to rules, contracts,
  frontend guidance and the implementation roadmap. This documentation pass
  provides no runtime test, rendered-interface or game-balance evidence.
- A [frontend/backend implementation review](docs/reports/2026-09-28-final-draft-implementation-review.md)
  of the final rules, identifying interaction gaps and proposed fixes with
  acceptance scenarios. It was advisory when written; the accepted amendments
  below now adopt its recommendations. The original report remains historical.

- The [final draft](docs/design/drafts/FINAL_DRAFT.md), derived from the third draft
  with the accepted Q01–Q12 rulings and the designer's three overrides. The
  [decision register](docs/design/RULES-IMPLEMENTATION-QUESTIONS.md) now records
  resolved rules and the tests future implementation must satisfy.
- A proposed [Go backend architecture](docs/code/ARCHITECTURE.md), API and
  state/delivery contracts, a shared-engine simulation ADR and CLI contract,
  rule-to-test/open-question register, and a simulator-first
  [implementation roadmap](docs/planning/ROADMAP.md). This is planning only;
  no backend, simulator, migrations or deployment stack is implemented. That
  planning slice did not adopt open rules; the final-draft decisions below do.

- Four complete [cgmsart card-art editions](docs/design/cgmsart-decks/README.md),
  with 208 reviewed card faces, one shared back, separate rank/suit overlays,
  contact sheets, comparisons, mobile-size previews and font/source provenance.
  These are cosmetic exploration exports; native artwork resolution is disclosed,
  production references are preserved, and the two-deck rules are unchanged.
- A reusable [CGMS art-direction skill](.agents/skills/cgms-art-direction/SKILL.md)
  defines the original cgmsart identity, exact theme tokens, accessible
  component candidates and worked layout examples, Flutter/PWA constraints, cited visual research,
  and button/card/layout validation. This adds design guidance only; game rules
  remain unchanged and the frontend prototype is a separate task.

### Changed

- Adopted F01–F10 and the operational recommendations from the first implementation
  review in the final rules, technical contracts and existing roadmap. Runtime
  implementation remains unchecked; one shared scenario contract must verify
  simulator, server command and reconnecting frontend observations.
- Softened the cgmsart reference background to matte blue-charcoal with sparse,
  subdued pigment. Cards, objects, board groups, labels and numbers now lead the
  visual hierarchy; vivid foreground artwork and meaningful action/state accents
  remain. Updated the skill, reference specifications and reusable instructions
  to maintain this balance. The static image remains an art study, not approved UX.
- Clarified that cgmsart's artistic concept is retained but its pictured UI is
  not approved for implementation. Palette, painted cards and material remain
  references; screen arrangements, navigation, sidebars, density, candidate widgets
  and prose geometry are unapproved UX examples. Future flows and responsive
  layouts must be designed from player tasks, preserving game, visibility and
  accessibility invariants. Historical art reviews and generation prompts do not
  grant UX approval or impose another approval step on authorized work.
- Recast cgmsart's scene as a functional painted game board: public holdings,
  exact scores, grouped combos, active alliances, pending deals and event history
  surround the current action. Removed physical-room/table/chair staging while
  preserving the approved card design, pigment palette and expressive controls.
  The concept remains static design guidance, with no Flutter implementation.
- Refined cgmsart to reserve human faces for King, Queen and Jack card art.
  Replaced player portraits with contextual scenery/items and compact seat
  identity, without reserving space for oversized decorative replacements.
  Preserved card design, colors, painted backgrounds, controls and game rules;
  updated the static concept and reusable specifications.
- Updated design links for `FIRST_DRAFT.md`, `SECOND_DRAFT.md`, `THIRD_DRAFT.md`
  and `IDEA.docx`; preserved the user's removal of the moodboard, suggestions
  and historical prototype record. Research retains its prior observations.
- Replaced Civic Ledger frontend guidance with cgmsart: illustration-led
  compositions, expressive cards/suit marks/buttons, chromatic dark atmosphere,
  new type/color/material tokens, detailed art briefs and visual-first review.
  Updated source research, accessibility and official Flutter/PWA contracts;
  added an original static concept study and recast artwork-hidden inspection
  as a fallback check. Game rules and the existing prototype are unchanged.

### Final draft — accepted interaction amendments

The user accepted all recommendations in the [first frontend/backend review](docs/reports/2026-09-28-final-draft-implementation-review.md).
These amendments extend Q01–Q12; they do not erase those earlier decisions or
turn the historical review into evidence of implemented behavior.

| Decision | Adopted rule or contract |
|---|---|
| F01 — Decisions inside effects | Persist effect/decision IDs, required actor, stage and authorized options. Accepted activation cannot be withdrawn; input stages do not create response windows. Coup can interrupt while awaiting input, never halfway through an atomic mutation. Committed random results survive retry/reconnect. |
| F02 — Costs and quotas | Use the authoritative per-ability lifecycle table for timing, prerequisites, declaration/resolution costs, cancellation and reset. Legal activation consumes the specified allowance; reconfiguration, alternate copies, custody changes and Fate cannot refresh that player's quota. Preserve earlier committed Infiltrator/Fate costs and the physical-Club exception. Successful ordinary Suicide Bomb destruction returns its committed attack Clubs; Baron waives that removal. |
| F03 — Victory admission | Check victory declarations against current authorized holdings, eligibility, phase and ended state. An unrelated stale aggregate version alone does not reject a currently legal declaration; ordinary selected commands retain their version checks. |
| F04 — Financial scope | Available points belong to one game. Coup clears active players' prior available pool before replacement awards, preserves completed repayments and handles surviving unpaid adjustments once. Financial finalization closes the available pool; the next game begins at zero cash while completed scores and debts persist. Departed players keep their defined exemption. |
| F05 — Settlement | Close board play immediately, then allow only the specified voluntary financial/offer decisions until financial finalization. Typed promises reference named awards. Finish-settlement requires resolved obligations or explicit refusal; disconnection is neither refusal nor consent. The next game waits for finalization. |
| F06 — Acquisition | Apply the acquisition/destination table and disclosure precedence: captured number Diamonds expose immediately, including Justice; other captured numbers/royals enter hand unless explicitly excepted, and Aces enter the concealed tray. Preserve authorized public history when a card later becomes concealed. |
| F07 — Lasting powers | Separate Underground's earned player privilege and Code's recorded declarer/settings from card custody. A valid Underground loan grants the recipient their own lasting privilege; income counts only kings allocated to their functioning formation. Code's exemption changes by legal declaration, while loss of the former formation ends its recurring penalty. |
| F08 — Doppelganger binding | Maintain one persistent binding per King. Separated bound Kings remain physical royals but cannot join another formation or attachment; reunion allows legal reconfiguration into their recorded slot. Preserve bindings through Fate, Justice and ordinary transfers; apply the named Kidnapper/assassination exceptions. Do not disclose a hidden partner's location. |
| F09 — Inflation | One active effect per target; reject duplicates without consumption. Use half-value number Spades for currency, combat and Negotiator. Clear only after one completed voluntary trade or purchase spending at least 20 printed Spade points, valuing that transaction before removal. Forced payments and accumulated smaller trades do not qualify; royal end-game awards remain 10. |
| F10 — Visibility | Publish separate hand and concealed-Ace counts while keeping faces and hidden royal/eligible-card breakdowns private. Apply the same matrix to snapshots, events/history, UI/semantics, reconnect, bots and logs. |
| Operations and match outcomes | Preserve untimed pending decisions through technical hibernation. Implement resumable, exact debt batching with compressed audit and small-ledger equivalence tests; no rounding or silent forgiveness. Pin a visible game count before the match (default three), share rank on tied match scores, and distinguish the ending declarer from score and match results. Public matchmaking/reward policy still needs its planned release contract. |

### Final draft — accepted decisions

The designer accepted the Q01–Q12 recommendations, with the protection,
alliance and Negotiator changes recorded below. These replace the third draft's
open interpretations and supersede conflicting historical suggestions. Balance
playtesting and implementation verification remain future work.

| Decision | Accepted rule |
|---|---|
| Q01 — Failed/canceled attacks | A legal declaration marks committed Clubs used for the turn, even if the attack fails or is canceled. Keep resolved costs/effects and the Clubs unless an effect removes them; charge success-only costs only on success. Initially illegal declarations spend nothing. No retargeting or added Clubs after responses begin, except Negotiator's explicit forced adjustment. |
| Q02 — Confinement | Cancel the confined player's pending action and end the current turn only if they are the active player. The action restriction and attack immunity expire immediately before their next scheduled turn; preserve Coup's exception. |
| Q03 — Hidden victory proof | Controlled hidden cards count once each. Reveal the cards needed to prove a valid threshold, verify historical opening of all four suits, and open no new response window. Reject an invalid online claim privately without changing game state. |
| Q04 — People-family protection | Natural twin Queens of Hearts form Great People; the Doppelganger-assisted form is People. Either protects one controlled Clubs, Spades or Diamonds series; protected Clubs cannot attack. Great People may instead protect one other eligible combo against kidnapping/assassination. Neither can protect itself, any other People or Great People combo, Hearts or Suicide Bomb. Changing the target is reconfiguration. Two Doppelganger Kings fill one fixed replacement slot, remain two physical royals, transfer together under Kidnapper's special rule and cease substituting if either is assassinated. |
| Q05 — Exclusive alliances | Each player can have at most one partner at a time: alliances are exclusive two-player pairs permanent for that game. Four players can have zero, one or two disjoint alliances. An existing member cannot accept another partner; returning the loan does not dissolve membership. Allied Ponzi splits equally between its user and that sole partner; any alliance still excludes Coup. |
| Q06 — Reopening/Fate | Restore an emptied historical series by adding cards without using the opening allowance. Fate resets the controlled physical cards, separates drawn Aces and exposes drawn number Diamonds without another guaranteed-Diamond redeal; preserve financial state, alliance, opening history, lost initial protection and usage counters. Card-dependent abilities stop with lost support; explicitly persistent benefits survive. |
| Q07 — Deck/Aces | Shuffle cards returned to the deck into the draw supply before the next draw; ordinary discards wait until that supply empties. Exhausted mandatory draws take what remains with no future entitlement; reject unsuppliable optional purchases after accounting for returned payment cards. Exhausted initiative draw-offs use rotating seat priority. A numerical Ace declares integer 2–10 for one attack or defense while concealed; it is not an opened series, threshold Diamond or Spade currency, and normal spent-card revelation does not undo its effect. |
| Q08 — Card details | Rename Forger to **Negotiator**. Its compulsory Spades transfer cancels the attack; when an exact subset is unavailable, the attacker must adjust using committed and unused Clubs to maximize the exact match. Only if no positive common sum exists does that adjustment fail without payment or consumption. Preserve the smaller-entire-series payment case and remove Baron's strength exception for that attack. Define Dexter's shared round limit and threshold/cost checks, private random Justice selection, Compensation's counted losses/remainder, and number-supported Exile/Barricade attachments in the final card reference. |
| Q09 — Departure/accounting | Defeat is voluntary surrender under quitting rules, effective before the next round; preserve negative results, debts, earlier scores and completed transfers, while forfeiting positive current-game score/available points and returning controlled cards. Loans held by others remain there. Continue with two players; end with ordinary scoring and no threshold bonus at one. Unanimous resignation restores start-of-game finances with consent from all affected players, including departed players. Forgive only identified unpaid debt, with creditor consent or unanimity for system debt, and a non-spendable score correction. Debt carries within a match; a separate match starts clean. |
| Q10 — Repayment order | Calculate simultaneous gross awards first, process receipts in fixed seat order and resulting creditor receipts FIFO, paying each recipient's oldest debt first. Settle fully before another action; use exact arithmetic, no automatic reciprocal netting and no rounding/cutoff that erases obligations. Payment decreases debt and creates no new obligation; optimized processing must preserve results. |
| Q11 — Online order/reconnection | One authoritative server order determines the first valid declaration; client timestamps do not. Coup may interrupt pending responses but never an ended game, and ordinary-action priority does not delay victory declarations. Preserve required pending decisions during disconnection with no automatic pass or forfeiture; reconnect or unanimously resign. |
| Q12 — Reputation | Record an explicitly accepted promise with recipient, amount/reward percentage and condition. The 40% threshold uses the promised amount. Full actual payment earns Trust; an accepted and paid reduction of at least 40% earns Anyhoo; rejection of a reduced final offer or refusal of an agreed transfer earns Scam. An accepted payment below 40% earns neither positive classification. Settle payer debts first, count recipient gross receipt, exclude mandatory Ponzi shares, aggregate promises to at most one outcome per ordered pair after game end and settlement of all triggered promises, and ignore conditions that never occurred. Refusal/rejected final offers take Scam precedence; pending promises defer classification. |

Current rule discovery, architecture/planning references and frontend rule
guidance now point to GAME_RULES and the resolved decision register. Earlier
drafts, review findings and the historical Flutter failure remain preserved;
this documentation update implements no runtime behavior.

### Third draft — historical accepted decisions

This section records what the third draft adopted at the time. GAME_RULES
is now authoritative, including the replacement of unrestricted partner
references with exclusive two-player alliances.

- Added the third draft as the current rules reference, carrying forward the
  second draft except for the changes below. Previous drafts keep their rule
  text; their navigation points to the current version.
- A player cannot be attacked before first opening a second distinct number
  series. The designer confirmed that this initial protection ends permanently
  at that opening; later card losses do not restore it.
- Initiating any attack requires two number series currently open. This applies
  to attacks through combos, non-combo cards, and aces as well as ordinary clubs.
  Underground does not bypass the attack restrictions. Coup retains its separate
  victory-declaration rules.
- Underground uniquely needs no exposed Hearts to open or function, including
  earning income and using its other benefits. Other defensive combos still need
  Hearts, and offensive combos retain their Clubs support requirements. The
  exception is stated in opening rules, enabling-suit rules, the card reference,
  diagrams, scoring, and implementation notes.
- Allied players may use Ponzi. The designer confirmed an automatic equal split
  among its user and all of that user's alliance partners. Each receives their
  share directly, then applies their own existing debt repayments. Exact shares
  may be fractional; no recipient receives a rounding advantage. An unallied
  user receives the full stolen amount, subject to their debts. Ponzi remains
  uncapped, takes only available positive current-game points once, cannot use
  Doppelganger, and can be canceled by Confinement. Alliance membership still
  disqualifies Coup; other promised reward splits remain voluntary.

The designer's responses to the second-draft gameplay assessment are recorded
here so that declined proposals are not treated as accepted changes:

| Review item | Decision for the third draft |
|---|---|
| 1 — Opening attack | Adopt initial protection until the second distinct series is first opened, plus two-current-series eligibility for initiating any attack. |
| 2 — Exact matching | Dismiss the proposed change. Keep exact matching and the existing attack order; do not add overpayment. |
| 3 — Persistent combo benefits | Add Underground's unique Hearts exemption. No reversal of Code persistence or Underground's retained non-income benefits is adopted. |
| 4 — Coup and Ponzi | Allow allied Ponzi with automatic equal sharing; retain uncapped individual Ponzi. No Coup warning step, Coup scoring change, or Ponzi cap is adopted. |
| 5 — Turn duration | Dismiss for now. No action budget, trading deadline, or new pacing restriction is adopted. |
| 6 — Alliance incentives | Dismiss the broader redesign. Keep the existing permanent combo-loan alliance and ownership rules, with only the accepted Ponzi exception. |
| 7 — Bookkeeping and unresolved rulings | Request recommendations. The proposals below remain unadopted and do not change the playable rules. |

### Third draft — historical proposals for review item 7

These were unadopted suggestions when the third draft was written, and the table
preserves that proposal history. Their rule topics are now resolved by the
final-draft decisions above; use those decisions rather than treating this table
as a pending approval gate. Presentation suggestions remain non-normative.

| Issue | Suggested mitigation |
|---|---|
| Score versus available points versus debt | Use four visible fields per player: match score, current-game score, available points, and outstanding debt. Keep creditor and age details in one transaction ledger. Show debt as an obligation already charged, never as another deduction from score. Preserve exact fractional allied-Ponzi shares. |
| Repayment and Coup bookkeeping | Record each receipt once with gross earnings, debt allocations, and the remainder available. Use a fixed Coup reconciliation worksheet showing which game entries are removed and which unpaid adjustments survive. Keep the current repayment order and economics. |
| Invalidated or canceled attacks | Validate the initial declaration, then mark its committed clubs used for the turn. If a response cancels it or breaks exact matching, the attempt fails without retargeting or adding clubs. Keep the clubs on the table unless a resolved effect explicitly removes them; capture/exchange happens only on a successful attack. Resolved response costs remain spent. |
| Confinement timing | Cancel the confined actor's pending action. End the turn only if that actor is the active player; otherwise the active player continues. End confinement's action restriction and immunity at the start of the confined player's next scheduled turn, without skipping that turn. Preserve Coup's explicit declaration exception. |
| People / Great People | Use Great People for natural twin queens of hearts and People for the Doppelganger-assisted version. Explicitly define whether each grants immunity, whether it protects itself, and how many targets it protects before rewriting the contradictory row. Do not infer these protection decisions from the names. |
| Hidden-card victory verification | Allow controlled hidden cards to count, but reveal every hidden physical card used to substantiate a threshold claim and count each physical card once. Verify the required suit-opening history. Validation follows the completed action/response window and does not open another response window. An invalid claim does not end the game. |
| Deck-return placement | Remove chosen insertion positions: mix returned cards into the draw supply before its next draw. Batch returns from one action into one shuffle, while obeying any card-specific earlier shuffle instruction. This removes predictable top-deck gifts without shuffling after each individual returned card. |

**Bookkeeping example:** a player charged 10 points who later earns 6 has a
current-game score of −4, zero available points, and debt of 4. Their creditor
receives 6. The 6-point repayment is not a second score deduction.

**Proposed failed-attack example:** 8 clubs attack 8 hearts; Advancement raises
the defense to 18. The attack fails, its clubs remain on the table but are used
for the turn, and the response remains spent. This outcome is a proposal, not
yet the third draft's adopted rule.

**Proposed Confinement example:** during Alice's turn, Bob declares Ponzi and
Carol confines him. Ponzi fails, Alice continues her turn, and Bob's confinement
expires when his next scheduled turn starts. This timing remains a proposal.

### Second draft — historical decisions

The entries below describe the second draft at that time. The later third-draft
decisions and current GAME_RULES take precedence, including their resolutions
of questions described below as then-unresolved.

#### Added

- A consolidated second draft with current rule text instead of inline amendments.
- A redeal of all initial hands whenever any player has no number diamonds.
- System and player debts with named creditors, automatic repayment from new
  points, and oldest-debt-first priority. Unpaid debts survive game transitions
  and Coup. A player creditor receives points only when actually paid.

#### Changed

- Coup requires only that its owner has opened Clubs, rather than two currently
  open series. Its card recipe, alliance prohibition, and ban on Doppelganger
  remain. It resolves immediately without a response window and must be declared
  before any other valid victory declaration ends the game.
- Coup resets current-game scoring: +50 for the owner and −50 for affected
  opponents, keeping defeated/quit exclusions. Earlier current-game points do
  not count, and no ordinary end-game card points or threshold bonus are added.
  Completed-game totals remain. The designer explicitly selected this reset
  when asked to resolve the conflict between request items 2 and 16.
- Ordinary victory is based on the first valid declaration, after the pending
  action and its responses resolve, rather than automatically reserving a win
  merely by holding the threshold cards.
- Ordinary combat usage is per physical club card, once per turn. Unused clubs
  can attack the same or another opponent, checking suit order for every attack.
- Other actions use a single pending action and one permitted response per other
  player in seat order. Responses resolve immediately without nested windows;
  Coup is the explicit immediate exception.
- The ordinary opening allowance is one new series OR one combo opening or
  reconfiguration per turn. Existing combos may still function and multiple
  combos may stay open. Underground retains unlimited combo opening and
  reconfiguration, except for Doppelganger's fixed assignment.
- Underground explicitly waives the two-open-series prerequisite for opening
  and using other combos. It does not waive Hearts/Clubs support, ace
  prerequisites, or the historical four-suit victory condition. Its previous
  income threshold and retained non-income benefits remain.
- Only the latest declared Code governs. Its settings persist even after its
  combo breaks; a new Code nullifies the earlier settings and recurring effect.
  Physical cards are not discarded merely because their Code is superseded.

#### Fixed

- Initiative draw-off cards are set aside until all positions are resolved and
  then shuffled back. Remaining ties require another draw-off.
- Suit-opening history and current series are stated separately: all four suits
  previously opened for ordinary victory; two currently open series for actions
  that require them, subject to the named exceptions.
- Ponzi explicitly requires both jacks of clubs and both jacks of spades. It
  transfers available positive current-game points once. Low, zero, and negative
  scores do not prevent targeting; zero/negative scores yield no benefit, and
  there is no ongoing future claim. No minimum positive score is invented.
- Existing Coup immunity, Ponzi cancellation by Confinement, individual card
  control, permanent accepted combo loans, Kidnapper targeting, Inflation's
  persistence, and Baron's round-long Suicide Bomb restriction are consolidated.
- Rule diagrams follow the declaration timing and per-card attack usage. The
  attack-order diagram now follows the existing written hearts → clubs → spades
  → diamonds order and Infiltrator's existing prohibition on bypassing to diamonds.

#### Decision notes and explanations

| Request item | Applied decision or explanation |
|---|---|
| 1 | Coup has Clubs-only series eligibility, is immediate and uncounterable, and cannot overturn an already declared victory. Ordinary declarations do not reserve victory before their qualifying action resolves. |
| 2, 16 | Clarification selected a current-game reset to +50/−50, without ordinary end-game scoring. Coup's owner wins that game; cumulative totals still determine the match. |
| 3 | Redeal if any initial hand has no number diamonds. Keep draw-off cards aside until every initiative position is settled. |
| 4 | Each physical club may attack once; unused clubs may attack the same or a different opponent. |
| 5 | Explanation only: a legal response can make an initially valid attack fail its legality or exact-match check. No new failed-attack spending/refund rule has been approved. |
| 6 | One action, one response opportunity per other player, immediate response resolution, no response-to-response windows. Ordinary declaration priority is active player, then turn order; Coup's immediate victory declaration is the exception. |
| 7 | Coup immunity and Ponzi's vulnerability to Confinement were already in the first draft. They are preserved, with Coup's obsolete response delay removed. |
| 8 | One combo opening/reconfiguration instead of a new number series; existing combos and multiple active combos are allowed; Underground keeps its exception. |
| 9 | Historical openings satisfy the four-suit victory requirement. Current openings govern the two-series action prerequisite. No new rule for restoring an emptied series or charging an opening allowance for additions has been adopted. |
| 10 | Explanation only: control determines whose cards count; the existing loan rule already answers that. Whether hidden cards qualify is a separate explicitness question. No new hidden-card counting or reveal procedure is adopted here. |
| 11 | Accepting a combo loan already forms an explicit permanent alliance for that game. The earlier suggestion was redundant. No transitive-alliance or whole-table membership rule is added. |
| 12 | Debt names its system/player creditor. Pay what is available, defer the shortfall, and automatically repay the oldest obligation from future receipts. Creditors score only actual payments; debtors record gross earnings without a second repayment deduction. Worked examples cover carryover and Coup resets. Follow-up decisions confirm unpaid debt survives Coup and age determines repayment order. |
| 13 | Second draft becomes the working document; the first remains the historical reference. Explanations and unresolved decisions live here, separately from rule text. |
| 14, 15 | Ponzi uses the four specified jacks; no score-based targeting restriction. A positive but low score only yields that small amount; zero/negative yields nothing. |
| 17, 18 | Code survives being broken. Only a newer Code replaces it, with no simultaneous governing Codes or stacked penalties. |
| 19 | Underground's exemption extends to opening and using other combos without two series currently open. |

**An invalidated attack:** suppose an attacker commits clubs worth 8 against
hearts worth 8. A legally usable Advancement ace adds 10 defense, so the declared
8-point attack no longer matches that 18-point defense. Confinement instead
cancels an action outright. The unresolved question is whether the committed
clubs count as used, whether retargeting or adding clubs is permitted, and which
costs remain paid. The earlier proposed automatic consumption rule has not been
adopted; it was requested for explanation only.

**Response example:** Alice declares Ponzi against Carol. Bob gets the first
response opportunity. If he plays Confinement to protect Carol, it immediately
cancels Ponzi and closes that action's window; nobody gets a fresh window to
counter Bob's Confinement. If Bob passes, Carol may respond next. If all pass,
Ponzi resolves. A valid Coup is immediate and ends the game instead of waiting
through this sequence.

**Hidden holdings versus control:** if Alice lends Bob a combo, Bob controls its
cards, as the first draft already states. Separately, if Bob has 14 exposed
royals and one closed royal, the question is whether the closed royal permits a
15-royal declaration. This update preserves the existing control wording and
does not invent a new hidden-card eligibility or reveal rule.

### Historical remaining design questions

This list records questions carried forward when the third draft was current.
Q01–Q12 in GAME_RULES and the decision register resolve the listed rule topics;
they are no longer pending approval. The balance/playtest item remains future
validation, not an unresolved rule decision.

- Failed or canceled attacks: committed-card usage, refunds, and retargeting.
- Confinement's exact expiry and how confinement outside the target's own turn
  interacts with turn progression.
- Explicit hidden-card victory eligibility and verification; exactly tied
  simultaneous victory declarations in tabletop play.
- People/Great People naming and protection scope, including its contradictory
  natural-versus-Doppelganger wording; physical Doppelganger components and
  overlap/removal rules.
- Forger payment/refusal details, Dexter modes and limits, Fate's exact reset
  scope, Justice's closed-card selection, Compensation loss accounting, and
  Exile/Barricade attachment requirements.
- Defeat, quitting/resignation and debt interactions, negotiated reversal of
  penalties, alliance chains, and fewer than three remaining players.
- Numerical ace values/exposure, general deck-return placement and empty piles,
  and draw-off exhaustion before all positions have been settled.
- Balance and player-agency questions require separate three- and four-player
  playtests. No unapproved recovery action, Code reversibility, or ace/trade
  restriction is introduced in this revision.
