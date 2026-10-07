# CGMS shared project context

## Project identity

- **Name:** CGMS (Card Game Mafia State).
- **Purpose:** Three- or four-player cards, negotiation, private holdings and persistent debts.
- **Current implementation:** Shared Go rules engine/CLI simulator, PostgreSQL durable match authority, and local single-process HTTP/WebSocket server with sessions/rooms/recovery; the [Flutter client](../../client/README.md) integrates the authority with adaptive online UI and retains local presentation scenarios.
- **Planned stack:** Go, Flutter web/iOS/Android, PostgreSQL, Redis and Nginx; production v1 uses Docker Compose.

## Key paths

| Concern | Authority |
|---|---|
| Operating rules | [AGENTS.md](../../AGENTS.md) |
| Current rules | [Design index](../design/README.md), currently [GAME_RULES.md](../project/GAME_RULES.md) |
| Accepted changes/proposals | [CHANGELOG.md](../../CHANGELOG.md) |
| Architecture and APIs | [Architecture](../code/ARCHITECTURE.md), [game API](../code/API-game.md) |
| State/recovery contract | [State and delivery design](../design/DESIGN-go-state-and-delivery.md) |
| Sole implementation plan | [ROADMAP.md](../planning/ROADMAP.md) |
| Simulation | [CLI contract](../../sims/README.md), [proposed ADR](../design/ADR-0001-shared-go-simulation-engine.md) |
| Accepted rule decisions/tests | [Rule register](../design/RULES-IMPLEMENTATION-QUESTIONS.md); Q01–Q12/F01–F10/N01–N07/T01–T03 adopted |
| Project scope and terminology | [Charter](../project/CHARTER.md), [decision log](../project/DECISION_LOG.md), [glossary](../project/GLOSSARY.md) |
| Historical implementation assessment | [Third frontend/backend assessment](../reports/2026-09-28-final-draft-third-assessment.md), recommendations now adopted |
| Art direction | [cgmsart skill](../../.agents/skills/cgms-art-direction/SKILL.md) |
| Approved UI/UX and theme | [UI baseline v2](../design/UI_BASELINE.md), adaptive revision authorized 2026-09-30; adaptive online implementation and retained v1 scenarios; [mockups 01–20](../design/mockup/README.md) remain supporting storyboards |
| Local Flutter demo | [Demo](../../client/README.md), [integration boundary](../../client/INTEGRATION.md) |
| Tracking and operations | [Tracking schema](tracking.schema.md), [xops](../../xops/README.md) |

## Active context

The live roadmap is authoritative. Phases 0–2 and 4 have bounded local evidence;
P01/P02 host gates passed. P04 passed 340 final-build browser checks after the
direct move into `client/` and demonstrated room/decision/recovery usability repairs. Native
acceptance and all Phase 5 production work are excluded from this execution.
P03 remains explicitly allocation-blocked. The owner approved 10 dirt per eligible
finalized online game and 30 dirt per pass day on 2026-10-01. Decisions 0015–0016
remove cosmetic sales/unlocks and new promotions, restrict products to Ad-free and
Daily/Weekly/Monthly/Yearly, and assign shared card styles for an entire match.
Only earned Daily/Weekly access is currency-redeemable; it retains ads. Standalone
Ad-free lasts 30 days without unlimited starts. Decision 0018 makes paid Daily a
non-renewing 24-hour unlimited/ad-free pass and reaffirms ad removal for all paid
Weekly/Monthly/Yearly access. The existing 30-dirt/day rate
continues to calculate Daily 30 and Weekly 210. P05 local economy integration passed real database race and 24 three-engine
browser cases. P06 approved terms passed 268 Flutter tests, affected real-database
race suites and 72 final-build account cases across Chromium, Firefox and WebKit;
see the [approved-term report](../reports/2026-10-01-approved-paid-access.md). The roadmap
is 39/46, with only P03 and six Phase 5 gates open. Live commerce remains disabled.
Historical simulation development blocks and replay checks do not establish bot
strength or balance. Refer to the v3 inference plan and fresh gate evidence.

GAME_RULES accepts the Q01–Q12 recommendations with exclusive two-player
alliances, a ban on People-family protection of other People-family combos,
and Negotiator replacing Forger with forced maximum exact matching using
committed and unused Clubs. The roadmap and linked suites record their implementation evidence; acceptance
of a rule alone is never runtime evidence.

The first frontend/backend review's F01–F10 recommendations are now adopted:
durable decisions within an effect, explicit ability costs and quota ownership,
current-state victory admission, game-scoped available points, board closure
followed by voluntary settlement and financial finalization, acquisition and
visibility matrices, player-owned lasting powers, split Doppelganger bindings,
and non-stacking Inflation. Public hand and concealed-Ace counts are deliberate;
faces and eligible-card breakdowns remain private. Online setup fixes the match
game count before play (default three) and tied match scores share rank. Technical
hibernation and resumable exact debt batching must preserve outcomes; they do
not authorize timeouts, rounded debt or automatic refusal. The engine, persistence and API suites exercise these contracts; native/client
and deployed-service acceptance remain separate roadmap obligations.

The second assessment's N01–N07 are now accepted too: explicit combat algebra
and physical exchanges, public persistent-Ace custody outside Fate inventory,
response omission for confined/departed seats, Code's global settings versus
eligible hostile penalties, immutable game IDs on commands, cross-game
nonspendable forgiveness corrections, and per-card next-round availability.
Old report bodies remain historical; future findings must use new IDs and
must not present these accepted decisions as unanswered questions.

The third assessment's T01–T03 are accepted in GAME_RULES §§2.9, 2.15 and 2.16:
revisioned, nonblocking, turn-scoped trade/loan proposals; accepting-player action
identity with success-only transfers/alliances/privileges; ascending-seat deferred
Compensation draws; and chosen-quantity atomic purchases with exact pricing.
Four prior drafts live in [the archive](../design/drafts/README.md). Historical
reports retain their old wording/hashes and relocated evidence links.

## Next session

Read session recovery state and the live roadmap before work. The complete Flutter
host is directly in `client/`, with reusable `client/packages/cgms_ui/` and the
preserved offline tools. Use the local developer/browser commands in its README.
P03 allocation remains owner-blocked; no exhausted campaign may be restarted.
Use current P05/P06 evidence and decisions 0014–0018 for product boundaries.
Cosmetic catalogue/promo channels are excluded; unspecified paid-product terms
and final native/production acceptance remain distinct.
The human commits/pushes via `make git`; preserve all concurrent changes.

## Conventions and boundaries

Game rules and accepted changes override illustrative visuals and historical
reviews. Historical drafts and proposals do not override GAME_RULES.
Native offline play must not grant online rewards; dirt/entitlements are distinct from game
score/debt. [UI baseline v2](../design/UI_BASELINE.md) governs distinct adaptive
Flutter compositions and the preserved cgmsart theme. Static mockups and the
current v1 implementation support review but cannot override v2 requirements
or canonical rules. Proposed commands/config paths are
unavailable until implemented.

Preserve user and concurrent edits. Only the coordinating parent tracks and
stages; humans commit and push. Service credentials and private game state
must not enter source control, logs or public event/replay payloads.

## External services

No external runtime provider is selected or integrated. Identity, purchase,
store, advertisement and client telemetry choices require their scheduled
policy/contract work; no SDK or endpoint is implied by this plan.
