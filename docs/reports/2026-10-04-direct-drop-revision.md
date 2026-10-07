# Direct drops and scheduled draw feedback — 2026-10-04

The owner revision supersedes the blanket confirmation requirement in the
[earlier interaction report](2026-10-04-card-interaction-redesign.md). Its rule and
command inventory remains useful historical coverage; the current interaction
contract is the [UI baseline](../design/UI_BASELINE.md). Accepted gameplay rules,
the four public quadrants, cgmsart and the removed permanent composer remain.

## Implemented behavior

The draw pile occupies a reserved band between the public seat rows. The same
layout rule applies across widths; public cards retain their size, each pile
target is at least 48 logical pixels, and short or enlarged layouts scroll to
reveal the complete target. The original overlap was reproduced in a disposable
1041×948 fixture before changing the layout.

The server already expires Confinement and draws one card in `BeginTurn`, subject
to its existing exhaustion rules. Those mechanics are unchanged. The additive
public `online.turn_started` boolean lets the client distinguish an authoritative
scheduled draw from other acquisitions. Only contiguous live events with that
transition and one acquired card generate feedback. A separate stream baseline
handles HTTP refresh arriving before the matching event; a short sequential
queue preserves multiple bot draws delivered in one rendering frame. Duplicate
events, rebuilds and resizing do not replay cues. Reconnect, game replacement
and closure clear the queue and establish a fresh baseline.

The pile gives a restrained amber glow with a privacy-safe live announcement.
The local announcement distinguishes hand from concealed-Ace acquisition;
opponents receive only the public seat and one-card count. Ordinary turn-drawn
Diamonds remain in hand. Reduced motion uses a static quiet highlight. Feedback
does not call a command, reveal a face, invent a deck count or move focus.

A complete legal drop submits once through the existing bound command envelope.
Before release, transient feedback identifies the physical cards, destination,
action and actual costs. Missing parameters retain a small action-specific
choice. Complete action or supported attachment choices commit directly; other
genuinely missing parameters retain their contextual commitment. Board targeting
uses precise guidance and highlights applicable targets. Tap selection,
double-tap/keyboard activation and independent inspection remain available.

First number-series opening includes every required held number of its suit.
Historical addition/restoration preserves the selected physical copies. Royal
attachment choices name only supported legal roles; passive effects are not
invented as actions. Self-sacrifice abilities reject unrelated opponent areas.
Ace suggestions preserve current-series prerequisites, round quotas, duplicate
Inflation restrictions and exact pending-actor response contexts. Invalid drops
never expose or discard an unused Ace.

The complete 18-card first-opening preview groups rank and suit with every copy
number in brackets, preserving exact identity without repeating long card names.
The transient preview stays within safe screen bounds and uses the configured
text size. Selection and inspection share the existing hand heading line, so a
first tap no longer shifts the physical card away from the next tap.

Returned Doppelganger kings retain their fixed pair and role in private
`board.own_bindings` metadata. Only currently controlled authorized cards receive
opaque handles; a separated partner controlled by someone else is omitted.
The client rejects attachment/rebinding suggestions and preserves valid reuse.
Historical projections may omit the additive field; present values remain exact.

Dragging a whole exposed formation to another player's area, including a nested
public-card or series target, proposes that exact intact loan. It cannot become
an attack. The proposal is private, revisioned and nonblocking. Neither proposing
nor accepting transfers the formation; successful resolution of the acceptance
response window performs the atomic transfer and establishes the alliance.
Combo-assisted attacks remain distinct card-driven activations.

The existing game/version/turn/window/effect/source bindings, same-game victory
exception, uncertain-operation reconciliation, explicit Pass/End Turn and
settlement controls are retained. Cancellation does not withdraw accepted play.

## Regression and authority evidence

The host gesture suite uses actual mouse/touch pointer sequences for direct
opening, exact duplicate copies, first-opening expansion, historical additions,
royal missing-role completion, contextual Aces, whole formation loans and nested
targets. Existing stale/revoked-source, game replacement, cancelled drop, Escape,
inspection, effect-stage and uncertain-submission tests run with these changes.

The draw tests cover contiguous events, repeated cursors, refresh ordering,
batched bot events rendered sequentially, reconnect, game changes, privacy,
actual destination and reduced-motion semantics. Layout tests measure pile/seat
separation and accessible targets through responsive and enlarged layouts.

The optional projection field retains historical journals that lack it while
rejecting changed values when present. Game/wire/matchstore/fixture race tests
and PostgreSQL receipt, version, recovery and privacy cases exercise the affected
authority boundary. Two deterministic QA scenarios force an Ace or ordinary
Diamond at the next scheduled draw; the engine performs the actual draw.

## Executed acceptance

Final source counts, build and served-asset hashes, exact commands/logs, browser
reports and curated capture locations are recorded in the
[compact verification record](2026-10-04-direct-drop-verification.json).
Browser work uses disposable authority on port 8081 and separate authenticated
seat contexts. No commands are submitted in an existing user match.

The source gates pass 475 host tests, 79 shared UI tests and 49 browser-support
tests, with both Flutter analyzers clean. Affected Go packages pass with the race
detector, six PostgreSQL recovery/privacy/receipt integration cases pass, and all
208 artwork inventory hashes remain intact.

All 176 cases pass on the final `b051251f` application bundle, with no unexpected
browser errors. Each report records its full asset and harness hashes; 14 curated
captures include pile separation, draw cues, exact drops and the complete loan
proposal/acceptance/response sequence. Disposable authority exits cleanly and
its temporary database container is removed; the ordinary preview remains up.

Local captures: [1041×948 pile](../../client/.browser-tools/artifacts/direct-drop-verified-geometry/chromium-390-pile-separated-1041x948.png),
[draw feedback](../../client/.browser-tools/artifacts/direct-drop-verified-draw/chromium-turn-draw-ace-glow.png),
[exact loan preview](../../client/.browser-tools/artifacts/direct-drop-verified-interactions/chromium-390-direct-loan-exact-members-preview.png),
and [natural bot settlement](../../client/.browser-tools/artifacts/direct-drop-verified-bots/online-bots-chromium-beginner-natural-match-settlement-complete.png).

| Engine | Executed coverage |
|---|---|
| Chromium | 44 gameplay journeys, six direct-drop/large-preview cases, two draw-feedback cases, two reconnect cases, 12 RTL physical-copy cases, 17 layout cases and 11 bot-room cases |
| Firefox | 17 layout cases and 48 keyboard inspection/selection cases on long-lived hand and public-card pages |
| WebKit | 17 layout cases |

The layout runs include 1041×948, live 599↔600 and 1049↔1050 transitions,
rotation, short windows and RTL at 200% text. Direct-interaction captures include
all 18 physical cards in a first opening at 200% text, including a short rotated
viewport. Firefox inspection preserves exact selections without command POSTs.
Expected deliberately injected network faults remain separately classified.

A normally dealt beginner match with one human seat and three bot seats completes
13 rounds and settlement (one completed game, final authority version 711).
Standard and advanced matches cover opening, bot responses,
reconnect, privacy and returned human control; their full settlement is not
claimed. Rare actions use explicitly authorized deterministic positions.

Final visual review also replaces a missing arrow glyph in the drop preview
with an ASCII separator. The report records the rebuilt assets and refreshed captures.

The ordinary preview is [CGMS Online](http://localhost:8080/#/online). Source is
under `client/` and `client/packages/cgms_ui/`; backend changes are limited to
additive projection compatibility and deterministic QA scenarios.

## Acceptance limits

Browser engine results apply only to the executed reports and final build.
Headless pointer/touch emulation and browser accessibility semantics do not
certify physical Android/iOS devices, native Safari or external screen readers.
Rare rule combinations beyond the listed deterministic browser journeys use
widget and authority tests; they are not claimed as naturally dealt play.
The pre-existing `client/pubspec.lock` is preserved byte-for-byte. It contains
older-SDK resolutions; verification uses the repository-pinned Flutter 3.47.5 /
Dart 3.13.4 dependency cache with `--no-pub`.

A pre-existing wording issue is recorded as a nonblocking tracking note: at a
round-limit closure, the Results popup shows departure acknowledgement text.
Finish Settlement remains available through Game options; no required settlement
action is unavailable. This wording correction is outside the interaction slice.
