# Card interaction redesign — owner revision, 2026-10-04

This report separates the command audit and acceptance obligations from executed
verification. The owner supersedes all permanent action/composer/review areas.
The accepted [rules](../project/GAME_RULES.md) and
[decision register](../design/RULES-IMPLEMENTATION-QUESTIONS.md) are unchanged.
Both documents were read in full for this audit. Discovery used CodeGraph before
source inspection; the 51 client command names and field sets agree with
`backend/internal/wire/codec.go` and the actor/lifecycle adapter.

## Gesture and commitment contract

- A tap selects an exact physical card; another tap toggles it. Selections may
  contain source cards, payments and targets with different visible roles.
- Double tap or the card's keyboard/assistive activation opens only the applicable
  interpretation or parameter flow. Full-card inspection remains independent and
  never calls `expose-unused-ace`.
- Dragging one card or its selected bundle to an own series/formation, opponent
  card/formation/seat, own hand or draw pile prepares the same intent as tapping
  those sources and destination. Opponent cards are targets, never drag sources.
- Card moves require explicit contextual confirmation; compact non-card utilities
  retain their explicit actions. Escape, outside dismissal,
  inspection, drag cancellation and invalid drops cannot submit or imply Pass.
  Dismissing an accepted effect's decision leaves that decision pending/reopenable.
- No permanent action panel, reserved slot, universal modal composer, replacement
  toolbar, action tray or sidebar is part of the revised interface.

All rows below inherit this commitment/cancellation behavior. “Own” means an
own-turn idle boundary; “Response” means the authority's current required actor;
“Inside” means the exact current effect stage. Client suggestions are hints.

## Rule → command → gesture → test matrix

Test references identify executable suites/fixtures and assertions; they do not
claim a browser journey for every matrix row. `I` is
`client/test/card_intents_test.dart`, `W` is
`client/test/card_play_screen_test.dart` plus `client/test/online_screen_test.dart`,
`C` is `client/test/online/online_client_test.dart`, `A` is
`client/test/adaptive_table_test.dart` plus shared-package card interaction tests.
Domain authority remains in `backend/internal/game/`; durable/privacy tests remain
in `backend/internal/matchstore/` and `backend/internal/wire/`.

| Rules / action | Wire command and exact payload fields | Cards / destination / extra choice | Timing and acceptance assertions |
|---|---|---|---|
| §§1.6, 2.12 series first opening/addition/restoration | `open-series`: `suit, selection` | Select numbers → own suit; first opening includes every held number of that suit; confirm exact physical set | Own. Historical restoration/additions consume no new-opening allowance. I/W opening and restricted-card cases; QA `opening`. |
| §§1.6, 2.11 combos: Fate, Baron, Underground, People/Great People, Justice, Code, Kidnapper, Ponzi | `open-formation`: `formation, formation_id` | Select exact members; double tap or own-formation target; choose kind only when ambiguous, existing formation when reconfiguring | Own; opening allowance/two current suits/support apply, with Underground's precise exceptions. I/W natural/substituted and repeated-formation cases. |
| §§2.11 Doppelganger | Same `open-formation`; `formation.substitute={kings,slot:{rank,suit}}` | Two selected kings explicitly fill one recorded slot in the chosen combo | No standalone two-king opening, no Coup/Ponzi substitution, no component overlap or hidden binding leak; fixed binding survives returns. I/W fixed pair/role assertions. |
| §§2.11 People protection | Same `open-formation`; `formation.protection={series}` or `{formation}` | Select a concrete own target; natural Great People may choose another eligible formation | No Hearts, Bomb, self or either People variant; protected Clubs cannot attack. Reconfiguration consumes its allowance. I/W protection/removal cases. |
| §§1.6, 2.11 whole combo take-back | `take-back`: `selection, formation_id` | Formation → own hand, with the exact whole formation shown | Own; never number series/attached royals; fixed binding persists. I/W whole-set and multiple-formation cases. |
| §§1.6, 3.3 royal attachment | `attach`: `selection, suit` | Own eligible royal(s) → explicitly chosen supported own series; destination suit distinct from printed suit | Own; untargeted activation cannot default to Hearts or confirm before choosing a series. Exact named support and component exclusivity; no take-back. I/W royal-to-different-suit cases. |
| §§2.12, 3.3 Ringleader | `ringleader`: `selection, ace, value` | Attached Diamond King activation; selection identifies that King; no ordinary extra Ace/value UI | Once/game, own; +10 and discard only on resolution. I/W cost/quota; authority rejects unused optional fields where inapplicable. |
| §§2.12, 3.3 Richer | `richer-sacrifice`: `selection` | Attached non-Diamond King activation, explicitly sacrifice that King | Own, once/round; support requires exactly one attached King; discard/draw up to 2 only on resolution. +2 round income is passive. I/W costs/support. |
| §§2.12, 3.3 Barricade | `barricade-sacrifice`: `selection` | Hearts Queen attached to Hearts plus five OTHER cards; Queen first in payload, visually separate payment roles | Own, turn-cycle; discard Queen, return 5, restricted replacement draws on resolution. Defensive exchange is passive. I/W Queen ordering/restrictions. |
| §§2.12, 3.3 Dexter assassination | `dexter-assassination`: `selection, targets` | Attached Diamond Jack enters flow; one exposed number Diamond is payment; one exposed enemy royal is target | Own; prepayment Diamonds ≥20, shared round allowance; Great People/protection/attack checks. I/W payment versus target, no hidden royal. |
| §§2.12, 3.2 Fate | `fate`: `selection, target_seat, formation_id` | Choose specific Fate formation, consumed King, enemy seat | Own, turn-cycle; return chosen King before reset; active-effect Aces excluded; bindings/history/quotas/restrictions persist. I/W exact formation/cost and authority Fate tests. |
| §§2.12, 3.2 Code | `code`: `formation_id, value, price` | Specific Code formation activation; choose Diamond bonus 1–10 and price 5–10 | Own, turn-cycle; economic settings persist independently of passive round penalty. I/W numeric fields/quota. |
| §§2.12, 3.4 Compensation | `compensation`: `ace` | Own concealed Hearts Ace activation; select named mode | Own, shared Ace/round plus activation/game; later eligible series losses only; persistent effect and deferred queue are automatic. I/W Ace modes; QA `compensation`. |
| §§2.12, 3.4 Inflation | `main-inflation`: `ace, target_seat` | Own concealed Spades Ace → eligible opponent, choose named mode | Own; one active effect/target; exact halving, qualifying printed-Spade transaction expiry. I/W Ace/target/cost. |
| §§2.12, 3.2 Kidnapper activation | `kidnapper`: `formation_id, target_seat` | Specific twin-Jack formation → opponent seat | Own, once/turn; server chooses concealed outcome, with no client hidden pool. I/W exact formation; authority random/privacy fixtures. |
| §§2.9, 3.2 Kidnapper exposed outcome | `kidnapper-outcome`: `selection, pending_action_id` | Tap only currently authorized exposed choice on board | Inside, required actor; only if no eligible concealed candidates. Do not reroll recorded outcome. W revoked/replaced choices, wrong actor/window. |
| §§2.4, 2.12, 3.2 Ponzi | `ponzi`: `value` | Exact natural Ponzi formation → opponent; wire `value` is target seat | Any idle including outside own turn; turn-cycle; two current series/eligible target, uncapped exact allied half shares. I/W out-of-turn entry and authority quotas. |
| §§2.5–2.8 ordinary combat | `attack`: `selection, targets, ace, value, suit, exile, infiltrator` | Own available exposed Clubs → frozen enemy number cards of one suit; optional selected Ace/modifiers | Own; A=D exactly at declaration and resolution, committed physical Clubs remain used on cancel; attack order/initial protection. I/W exact target bundle; QA `attack`, `combat`. |
| §§2.5, 3.4 offensive Ace modifiers | `ace, value` on the applicable attack command | Exact concealed Clubs Ace with Advancement (`value: 0`) adds 10 once; numerical Ace declares 2–10 with its face concealed | Own attack; shared one-Ace/player/round allowance, two current series, committed physical Clubs and frozen targets. W real gesture modes 0 and 7 assert exact opaque Ace/cost/target payload, no early submission or exposure. Domain `TestDexterAdvancementAttackAndInvalidPayment` verifies arithmetic/prevention/revalidation. |
| §§2.8, 3.3 Infiltrator | `infiltrator-attack`: same attack fields; or `infiltrator` on applicable attack | Attached Clubs Jack enters modifier flow; actual Clubs and non-Diamond target remain explicit | Own, turn-cycle; Jack returned at accepted bypass, retained cost on later cancellation. I/W bypass and Diamond exclusion. |
| §§2.8, 3.3 Exile | `exile` on applicable attack; no separate command | Attached Clubs Queen offers attack modifier with exact Clubs/targets | Own, round; bypass Barricade only; no implicit toggle from unrelated selection. I/W modifier/cost semantics. |
| §§2.5, 3.2 Baron | `baron-attack`: same attack fields | Specific Baron and Clubs → whole enemy series; confirmation names every target | Own, round; A > whole D, no combat points, protection/order still apply. I/W whole-series distinction. |
| §§2.5, 3.3 Bomb destruction | `bomb-attack`, `baron-bomb-attack`: same attack fields | Clubs (plus applicable Baron) → exposed attached Hearts Jack | Own; A≥5 separate from number matching; no numerical/Advancement defense; success returns ordinary Clubs but not Baron's. I/W Bomb distinction. |
| §§2.9, 3.2 Justice | `justice`: `ace` | Specific Justice Queen activates its formation; oddly named `ace` is that Queen's cost handle | Own, turn-cycle; choose/reserve in fixed seat order; return Queen at final atomic resolution. I/W Queen cost; QA `justice`. |
| §§2.9, 3.2 Justice closed choice | `decision-closed`: `pending_action_id, decision_id` | Current Justice choice offers concealed sampling without faces/IDs | Inside; authority samples once; no hidden rank/eligible-count leak, no duplicate reroll. W ordered/private and C recovery. |
| §§2.9–2.10 effect input | `decision`: `pending_action_id, decision_id, selection, ace, value, suit, target_seat` | Board choice from exact authorized current options; stage-specific parameters only | Inside. Negotiator: attacker maximal Clubs then defender matching Spades. Dexter: prevent/pass with payable Diamond. Justice: ordered exposed target or explicitly empty choice. Empty selection is not Pass. I/W actor/stage/options/staleness. |
| §§2.9, 3.4 response Compensation | `compensation-response`: `pending_action_id, ace` | Own Hearts Ace during authorized incoming attack | Response; before losses, shared Ace and game quota. I/W pending response types. |
| §§2.5, 3.4 numerical defense | `numerical-defense`: `pending_action_id, ace, value` | Concealed Ace → own targeted number subset; choose integer 2–10 | Response to incoming number combat; retain concealed identity until resolved/canceled. Never on Bomb. I/W numeric limits/privacy. |
| §§2.5, 3.4 defensive Advancement | `advancement-defense`: `pending_action_id, ace` | Concealed Clubs Ace → targeted own Hearts | Response, Hearts only +10; shared Ace quota; no Bomb target. I/W exact applicable response. |
| §§2.9, 3.4 Confinement | `confinement`: `pending_action_id, ace` | Concealed Diamond Ace → pending actor via current action target | Response; may protect self/third party; cancels actor action, never Coup; Dexter prevention is inside resolution. I/W response timing; QA `confinement`. |
| §§2.9, 3.4 response Inflation | `inflation`: `pending_action_id, ace` | Concealed Spades Ace → incoming attack's actor | Response only to attack on controller; no arbitrary target; duplicate active target invalid. I/W authorized response. |
| §§2.9–2.10 Negotiator activation | `negotiate`: `pending_action_id` | Attached own Spades Jack activates response | Incoming Clubs attack only, once/round, payable unrestricted Spades; no positive common match consumes nothing. I/W and ordered decision tests. |
| §§2.16 purchases | `purchase`: `quantity, payment` | Number Spades from hand/exposed own series → face-down draw pile; choose positive quantity | Own; known payment shortfall disables confirmation using the projected price and exact Inflation fractions. Server validates hidden supply before acceptance and after responses. No free draw, automatic max quantity or invented deck count; cancellation pays nothing. I/W; QA `purchase`. |
| §§2.3, 2.15 card trade/combo loan proposal | `offer`: `offer_id, revision, terms` where terms=`to,give,receive,loan` | Selected cards/whole exposed formation → recipient; private exact terms popup | Ordinary own turn; intact loan either party's turn; idle only. Revision1/new ID, new revision for edits, origin-turn expiry; nonblocking, no reservations. I/W trade/loan; QA `trade`, `loan`. |
| §2.15 proposal controls | `withdraw-offer`, `decline-offer`, `accept-offer`: `offer_id, revision` | Current private proposal record/context | Exact revision/party; acceptance actor starts response-bearing transfer; accepted action cannot withdraw. Final revalidation and atomic transfer/alliance. W stale terms/record action and C retries. |
| §§2.3–2.4, 2.13 voluntary return | `return-loan`: `selection, target_seat` | Own borrowed formation/cards → lender seat | Legal own boundary; only currently controlled borrowed cards; whole intact formation or surviving exposed/unassigned cards; no remote reclaim/automatic reformation. I/W lender/exact members. |
| §3.4 deliberate unused-Ace exposure | `expose-unused-ace`: `ace` | Own concealed Ace → explicit destructive choice, even when the sole choice → confirmation | Playing/nondeparted, including off-turn, confinement, restricted cards and pending effects; deferred server work blocks input. Authority rejects an already committed Ace; its hidden commitment identity is not added to projections. Never inspection, selection, drag preview, invalid gesture or generic clear. I/W real activation and cancellation cases. |
| §§1.2, 2.9, 3.2 Coup | `coup`: empty payload | Four actual permitted Queens have immediate card-driven activation even over another temporary flow | Any legal atomic boundary, including confined/pending; historical Clubs, unallied, unrestricted real Queens. Special same-game version tolerance. I/W modal/confinement and C admission; QA `coup`. |
| §§1.2, 2.9 ordinary declarations | `declare-ordinary`: `threshold` | Compact utility; choose Diamonds or Royals | Completed-action boundary, all-four history/private authority proof; no blanket stale-version rejection. W/C invalid claim/privacy. |
| §§1.4, 2.9 turn and response | `end-turn`: empty; `pass`: `pending_action_id` | Compact explicit utilities; Pass visible even without card response | End Turn own idle; Pass exact required responder only, never effect answer or inferred silence. W real taps, pending actor/empty options. |
| §1.9 round departure | `departure-choice`: `accept` | Compact Stay/Leave choice | Authority round boundary only; explicit choice, never disconnect/weak holdings. W touch/large text; bot browser journey. |
| §§4.3–4.4 promise offer | `promise-offer`: `promise_id,recipient,award_id,mode,amount,condition` | Temporary financial utility | Playing only; typed award-occurrence condition matches game/award/payer; no free-text adjudication. W promises and domain lifecycle. |
| §§4.3–4.4 promise decisions | `promise-accept`: `promise_id`; `promise-final-offer`: `promise_id,amount`; `promise-final-answer`: `promise_id,accept`; `promise-pay`, `promise-refuse`: `promise_id` | Exact private record and financial terms | Authorized payer/recipient; explicit acceptance/payment/refusal; silence pending. W/C; QA `settlement`, `final-settlement`. |
| §§4.2–4.4 transfers/forgiveness | `voluntary-transfer`: `recipient,amount`; `forgive`: `debt_id,amount` | Temporary financial utilities; exact rational amount | Proper playing/origin-game settlement boundary and participant permission; automatic oldest-debt settlement, no hidden card dependency. W/C authority ledger/retry tests. |
| §§1.9, 4.4 nullification/finalization | `nullify`, `finish-settlement`: empty payload | Compact financial decisions | All affected explicit consent/finishes; no automatic refusal; board closure distinct from financial finalization and next immutable game. W/C; QA `final-settlement`. |

## Authority and privacy obligations

The command envelope is `{command_id,game_id,expected_version,type,payload}`.
Payload handles are current actor/game/view capabilities; historical handles
never authorize a present selection. An unsubmitted draft binds game, version,
turn, pending action, decision, actor and current cards/options. Live changes
must revoke affected sources/targets and disable stale confirmations. A victory
intent remains a special case: authenticate/deduplicate first, bind the same game,
then validate current eligibility without rejecting solely for unrelated version
advancement. Do not apply ordinary selection-draft version rules to Coup.

`OnlineClient` freezes and stores a submitted envelope before sending; uncertain
submission blocks other mutation, reconciles its original receipt and retries
only that envelope. Canceling presentation cannot withdraw this accepted/unknown
operation. Reconnect preserves pending actors, costs, options, random outcomes
and deferred draws. `online.server_pending`/`automatic_pending` describe server
work; they are not permission for a client draw, response or timeout.

The server owns deal, initiative, begin-turn draws, random Kidnapper outcome,
Justice closed sampling, purchase/Fate/Richer/Barricade random outcomes,
Compensation queue, passive awards, settlement and next-game creation. None gets
a new card gesture that bypasses its rule timing. A draw-pile tap alone submits
nothing. No remaining-deck count is fabricated. A concealed numerical Ace's face
must be absent from opponent feedback, popups, semantics, logs and payloads.

## Existing executable verification paths

- Frontend: host tests/analyzer under `client/`; shared tests/analyzer under
  `client/packages/cgms_ui/`; meaningful pointer/touch/keyboard cases augment the
  intent unit suites rather than replacing them.
- Normal preview: repository-local `make web.re MODE=local` rebuilds/recreates
  the web service; `http://localhost:8080/#/online` is the configured local origin.
  Compare served `main.dart.js` with `client/build/web/main.dart.js` before capture.
- Natural deal: fresh browser identity, four seats, one game, **Play against bots**
  then **Start match** gives one human plus three server bots. Existing runner:
  `client/tool/bot_browser_acceptance.cjs`, migrated to card entry for this owner
  revision. Set `CGMS_BROWSER` and `CGMS_EVIDENCE_DIR`.
  Never reuse an existing user match for evidence moves.
- Disposable deterministic authority: from repository root, safe-run
  `python3 xops/postgres_integration.py --webfixture`; it binds
  `http://127.0.0.1:8081` and serves `client/build/web`. QA-only
  `POST /__fixture/session` requires that origin and JSON fields
  `{scenario,players,seat,run}`. Its response supplies `match_id`; session cookies
  stay private. Separate seat contexts share the same scenario/players/run.
- Supported rule fixtures in `backend/cmd/webfixture/main.go`: opening, purchase,
  attack, combat, decision (Dexter), confinement, justice, compensation, coup,
  loan (Underground), trade, settlement, final-settlement and card-style, plus
  economy fixtures. These are synthetic authorized scenarios, distinct from the
  required natural-deal bot match. Existing runners include
  `client/tool/browser_acceptance.cjs` and `reconnect_browser_acceptance.cjs`.
- Existing QA fixture coverage does not include every rare combo/modifier or
  Negotiator adjustment. Widget/domain evidence must say so; adding an isolated
  deterministic browser fixture cannot be reported as a naturally dealt flow.

## Host regression mapping

`client/test/card_play_screen_test.dart` drives real pointer gestures for idle
selection, independently selected duplicate copies, deliberate double tap,
mouse/touch drag cancellation, Spades-to-pile purchase, exact Great People
protection, Barricade Queen/payment roles, fixed public Doppelganger
reconfiguration, required-response closure, pending-decision dismissal/reopening,
ordered empty-choice commitment and same-window ownership revocation. Queued
confirmations cannot submit after the game or authorized target changes.
It also verifies distinct Justice formations/consumed Queens, neutral source
versus opponent-target roles, explicit private offer revisions and Escape after a
pointer-created popup when focus moves to a utility outside the board.
An additional two-view Negotiator regression activates the attached Jack, selects
the attacker's exact Clubs, switches to the authoritative defender/payment stage,
rejects a queued old-stage callback and commits the exact Spade payment. Dismissal
and reopening at either stage cannot answer or Pass. Modifier-first attack tests
start from Exile, add Clubs and a target on the board, and assert the exact payload.

Review corrections include exact protection target/member summaries, editable
Clubs after modifier-first activation, irrelevant response-drop rejection and
Escape even when a pointer leaves focus at the browser/Flutter root. A 600 ms
artwork-card touch hold followed by paced pointer movement has its own regression.
Resizing a longer private-offer context reproduced a clipped focused recipient;
the regression now preserves its element, focus node, scope, semantic identity,
48-pixel visible bounds and exact revision across breakpoints and rotation.
An independent browser mutation diagnostic also reproduced ancestor detachment
and native focus loss. A stable popup semantic group and focus revelation after
reflow repair these separate problems without replacing the selected control.
The dedicated reconnect browser check additionally reproduced two stale browser
accessibility offsets after the connection notice disappeared: 106 pixels when a
tooltip reparented the route semantics, and 22 pixels when the inner hand viewport
changed from scrolling to fitting its contents. A permanent semantic group above
the Navigator/Overlay prevents that route reparenting. The inner hand viewport
renews only on the scrolling-to-fitting transition, preserving its card subtree,
controller, physical identities, selection and Flutter focus. Widget regressions
and the real reconnect browser gate verify these properties, native focus and
the exact card position relative to the hand. After reconnect and pointer popup
dismissal, that position remains 88 pixels; no command is submitted. The fix uses
public Flutter APIs without modifying the pinned SDK.
Response and ordered-decision confirmations repeat the current action/actor,
required player/stage, committed Clubs and frozen authorized targets. The same
snapshot-backed summary is used in the pending context, so choosing an Ace cannot
hide the identity of the action being countered. Named Aces appear once as source
costs; numerical combat Aces retain their modifier role.

Purchase confirmation has a red/green regression for both ordinary payment and
fractional Inflation: a deliberately unaffordable quantity stays editable with
confirmation disabled, then an affordable quantity submits the same exact cards.
This comparison reuses the projected price and payment summary; hidden supply
continues to be validated by the server at admission and resolution.

Untargeted royal activation has a regression requiring a concrete own series
before confirmation; a previous implicit Hearts destination is removed. Real
card activation also verifies deliberate unused-Ace exposure while off-turn,
confined, restricted and during a pending effect. Even a sole disposal hint
requires choosing the named destructive action before its final confirmation.
The offensive-Ace gesture tests select actual Clubs and frozen Hearts targets,
then choose Advancement or numerical 7, asserting no premature exposure or
command before confirmation and sending only its authorized opaque handle.
The server retains a numerical Ace's concealed identity until combat resolves
or is canceled.

Series headers carry seat/suit identity instead of aliasing the first projected
card, which may be an attached royal of another suit. Header activation and
whole-series targeting resolve the exact current number members; individual-card
targeting remains separate. Invalid drops onto a specific card, series or
formation are consumed as invalid there, so an enclosing seat cannot silently
reinterpret the rejected destination as a different move.

`client/test/online_screen_test.dart` retains catalogue/payload, authority,
financial, custody and layout regressions through temporary contexts. Its
`chooseAction` helper seeds a draft for the exhaustive command-contract sweep;
that sweep is not evidence of physical entry for every command. Card-entry
coverage is provided by the gesture cases, the intent suite and shared pointer
and keyboard tests. Removed universal controls are not recreated for the test
harness. The same file checks contextual targets at 48 pixels, readable public
cards and adjacent hand, 599/600 and 1049/1050 transitions, 200% RTL choices,
keyboard/assistive focus revelation, private inspection revocation, explicit
Stay/Leave/Pass, and stateful financial/settlement screens.

## Executed verification

Source checks and final-build browser coverage are recorded separately below.
The [compact verification record](2026-10-04-card-interaction-verification.json)
contains exact logs, source/build/report hashes, all 124 passing browser cases
and 16 curated captures from 132 retained final-build screenshots.

The repository-pinned Flutter 3.47.5 toolchain is used for tests, analysis and the
web bundle. `client/pubspec.lock` now records its SDK-selected transitive versions;
no runtime dependency was added. CodeGraph's file watcher resolved the new
protection-summary, typed target activation and exact payment helpers after the edits; no folder
move or manual index rebuild was needed.

Executed source and authority checks (the backend is unchanged):

- `backend/internal/game` and `backend/internal/wire` pass in
  `/tmp/agent-runs/card-interface-rules-privacy--20261004T102249Z-118916.log`.
- Six real-Postgres recovery/admission/privacy cases pass in
  `/tmp/agent-runs/card-interface-persistence--20261004T102249Z-118913.log`:
  `TestEveryPendingStageRestoresAndAdvances`, `TestCompetingVersionsAndPrivacy`,
  `TestDurableReceiptAndConcurrentAdmission`,
  `TestOldGameReceiptAndFreshReplacement`,
  `TestSnapshotRefreshesDerivedPendingContextWithoutAdvancingMatch` and
  `TestCompletedGameAdvancesWithinMatchAndKeepsReceipts`.
- The shared package's 76 tests pass in
  `/tmp/agent-runs/card-owner-entry-final-package--20261004T123129Z-457870.log`.
- All 437 host tests pass, including the physical-card flow tests and the shared
  adaptive gesture suite, in
  `/tmp/agent-runs/card-owner-entry-final-host--20261004T123129Z-457853.log`.
- Both whole-directory analyzers report no issues:
  `/tmp/agent-runs/card-owner-entry-host-analyze--20261004T123157Z-459326.log`
  and `/tmp/agent-runs/card-owner-entry-package-analyze--20261004T123157Z-459340.log`.
  Formatting checks cover 17 changed Dart files with zero rewrites in
  `/tmp/agent-runs/card-owner-entry-final-format--20261004T123142Z-458706.log`.
- All 47 browser-runner support tests pass with zero skips in
  `/tmp/agent-runs/card-interface-final-browser-support--20261004T122242Z-440024.log`.
- All 154 filesystem links in the changed documentation resolve, recorded in
  `/tmp/agent-runs/card-interface-final-evidence-audit--20261004T125439Z-504494.log`.
- A direct catalogue-to-matrix check finds all 51 command names, with no omissions:
  `/tmp/agent-runs/card-interface-final-evidence-audit--20261004T125439Z-504494.log`.
  The same audit confirms all 26 changed client files still match their verified
  hashes, checks the full candidate set and finds no common secret patterns in
  additions. Backend rules and artwork are unchanged. Independent source review
  found no remaining actionable findings.

The final build's `main.dart.js` SHA-256 is
`6329b9b8b42ad72702d1250db3ed65d5c62b06df32ed02768a3d23e04df5e7f0`.
All four local web assets match the normal preview on port 8080 and the isolated
authority on port 8081. The build log is
`/tmp/agent-runs/card-owner-final-build--20261004T123300Z-461575.log`.

Chromium 153.0.8010.12 passes 69 grouped gameplay cases and two strict reconnect
cases on this build. The gameplay run has zero unexpected errors; its six exact
injected lost-reply faults and four observed anonymous-session challenges are
recorded separately. Reconnect deliberately interrupts the browser's network and
its real backend stream; each scenario preserves the same game/version, physical
selection and pending input with zero command posts. It retains the invalid
purchase quantity until the player corrects it and does not answer or pass
Justice. The selected card returns to the exact relative hand position, including
after pointer dismissal and keyboard focus.

Firefox 155.0 and WebKit 26.6 each pass 15 grouped geometry/context/keyboard
cases with zero unexpected errors and zero gameplay posts. Each covers 13
normal/RTL-200% viewport configurations, live public-card growth and focus across
599↔600 and 1049↔1050. Short and rotated windows reveal the hand through deliberate
scrolling. WebKit runs in the existing Playwright 1.63.0 Linux container; no OS
packages were installed. These are layout and keyboard checks, not the complete
Chromium gameplay journey or native Safari acceptance.

The natural bot suite passes 11 grouped cases across three fresh four-seat
matches, each with exactly one human and three server bots. It uses ordinary
guest lobbies, random production deals and real bot policies on the isolated
authority, with no synthetic hand or existing user match. Beginner play reaches
the round limit after 12 additional human turns and completes settlement:
one completed game, finalized financial state, four ranks and final version 728.
Standard play opens successfully and returns control after bot cycles; Advanced
legally confines the attempted opening, then returns control through normal
responses. Those two shorter journeys are not claimed as complete matches.
The report retains privacy projection checks and counts/digests rather than
credentials or full private snapshots. Coordinating visual review also covers
the populated 17-card phone hand, centered 15-card desktop hand and natural
match's completed settlement.

The final RTL physical-copy suite adds 12 passing cases: independently select
identical copies through emulated touch, retain selection through `I` inspection,
preserve 48-pixel exposed targets and cross live widths at 200% text. All six
browser reports use the same final build and have zero unexpected errors.

The coordinating review inspected final-build phone/tablet/desktop idle captures,
card-specific choices, drag destinations/cancellation, exact attack confirmation,
Confinement's pending actor/committed Clubs/frozen target summary and Justice's
ordered choice. The idle captures have no temporary context or permanent panel.
Captures remain in the ignored `client/.browser-tools/artifacts/` evidence folders.
The owned disposable authority, database containers, volumes and network were
removed after verification; port 8081 is closed. The normal port-8080 preview
still serves all four verified final assets. No existing user match was used.

Naturally dealt one-human/three-bot play is recorded separately from deterministic
rule fixtures. Synthetic fixture coverage does not prove every rare combo was
naturally dealt. Browser device emulation does not establish native Android/iOS,
physical touch-device or external screen-reader acceptance. Rare combo/modifier
and Negotiator paths beyond the listed browser journeys retain widget/domain
coverage; they are not claimed as browser-playtested flows.
