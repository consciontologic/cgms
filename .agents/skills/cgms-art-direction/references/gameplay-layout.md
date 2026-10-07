# Gameplay presentation contract — compact shared board

[UI baseline](../../../../docs/design/UI_BASELINE.md) governs the current compact
quadrant/hand arrangement; the twenty mockups remain flow references.

This reference preserves semantic and rule boundaries for future authorized
engine integration; it does not assert that the bounded local demo implements
every rule below. Read the canonical [GAME_RULES](../../../../docs/project/GAME_RULES.md)
and [decision/test register](../../../../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md).
Rules govern legality, timing and privacy; the current UI baseline governs presentation.

## Adaptive composition

Follow the baseline's full-width public quadrants and adjacent visible hand at every width.
Use bounded internal areas for long hands and actions; do not make the player
scroll the page past public piles to reach their hand.
Keep public holdings and all local physical copies reachable and inspectable;
class changes preserve selections, pending decisions and immediate Coup access.
The fields below remain semantic facts, not a requirement to show every field
at once. Use authorized group inspection without leaking private data.

### Separate fields that must never collapse

| State | Semantic distinction to preserve |
|---|---|
| Seat order | Stable named seats determine the response order beginning left of the declarer; this is not round initiative. |
| Round initiative | The authoritative order freezes for the entire round despite changed diamond totals. |
| Currently open series | Current number-card holdings govern the two-series attack requirement and normal ability prerequisites. |
| Historical opening | Ordinary victory requires all four suits historically; losing a series does not erase that history. |
| Initial attack protection | Ends permanently when the second distinct series is first opened. It never reappears after losses. |
| Physical club usage | Each exposed Club has its own available/used state; one physical Club can make only one ordinary attack per turn. |
| Opening allowance | Normally one new series OR one combo opening/reconfiguration, with Underground's specific exception. |
| Recorded current score, match total, available points, debt | These remain separate values; available points are not score, and repayment is not a second penalty. |
| Board/financial phase | Playing, board-closed settlement and financial finalization have distinct permissions. |
| Ability allowance and lasting effect | Use/reset, effect owner and current formation remain distinct; custody changes cannot refresh the same player's quota. |

These distinctions come from authoritative state. Presentation code must not infer
missing permissions from color, card artwork, animation completion or a local
counter. Do not implement unresolved rule choices inside a widget.

## Cards and private information

Two standard decks create duplicate rank/suit cards. Give every physical card
a stable internal identity and preserve that identity across hand reordering,
selection, transfers and per-turn usage. Never key a widget only by `rank+suit`.
For two visible identical clubs, present distinguishable local labels such as
`5 of Clubs · copy 1` and `5 of Clubs · copy 2`; the internal identifier itself
must not encode or expose an opponent's concealed rank or suit.

The initial deal is 20 cards per player, then initial diamonds are exposed and
aces use their separate closed tray. Hands may later grow. Support at least a
20-card private hand stress case; do not mistake `20` for a maximum or put all
dealt cards into an ordinary hand when the ace rule places them separately.

Use the current compact hand and public-pile presentation. Physical
copies remain independently inspectable; no overlapping private fan hides rank,
selection or focus. Preserve the existing full-pile review controls and text
reflow within the selected reference's hierarchy. Larger
holdings and enlarged text may scroll without dropping authorized cards.

Sorting is a local view preference and must not mutate physical-card identity
or public reveal state. Keep keyboard focus and selection on the same physical
card when sorting or receiving a card. If a selected card leaves the owner's
control, announce that the selection changed and revalidate the action.

### Concealment is a data and semantics contract

- Build another player's view from an authorized public-state projection.
  Painting a back over a face is insufficient if hidden values remain in
  semantics, tooltips, debug labels, DOM inspection, alt text, cached thumbnails,
  analytics or client-visible payloads. Do not send concealed identities to
  unauthorized clients merely to simplify animation.
- Show public **hand size** and **concealed-Ace count** separately; the rule
  intentionally reveals Ace category through its count. Faces and hidden
  royal/eligible-card breakdowns stay private. A concealed slot can be
  `Concealed card`. Never name an eligible hidden ace/royal through a disabled
  action's reason, possible combo list, search result or predicted score.
- For Kidnapper, the game selects randomly among eligible closed aces/royals when
  any exist. Do not render a selectable set of hidden identities. Show only the
  authorized result after resolution; open-card choice is the separate fallback
  when no eligible closed cards exist, respecting protection.
- An owner's inspection of their own private card is not an in-game exposure.
  Do not synchronize private inspector state or announce it to other seats.
  Conversely, do not create a public peek affordance that bypasses the ace
  exposure rules. A numerical Ace declares an integer value of 2–10 for one
  attack or defense while its identity remains concealed. Normal revelation of
  the spent Ace does not undo the completed effect.
- Potential end-game points must distinguish known public holdings from private
  or conditional estimates. Do not leak a private royal through a public
  `14 / 15` progress change. Controlled hidden cards count toward victory; only
  a valid claim reveals the proof cards it needs, without a new response window.
  Invalid online claims are privately rejected without state changes or disclosure.
- Use GAME_RULES’ acquisition/visibility matrix for every arrival animation
  and history entry. Captured number Diamonds expose immediately, including a
  hidden Justice draw; other captured numbers/royals enter hand unless a named
  exception applies, and Aces enter the concealed tray. Trade/loan/return use
  their explicit destination rules. An already-public identity remains in
  authorized history after concealment; history must not reveal later hidden
  movement or the location of a split Doppelganger's other King.

## Actions and response order

The primary action names its consequence: `Declare attack`,
`Open Hearts`, `Offer trade`, `Pass response`. Avoid generic `Continue` for
different game actions. Present the acting player, selected physical cards,
target, enabling conditions and current legal explanation before submission.

For ordinary combat, guide target selection through **Hearts → Clubs → Spades
→ Diamonds**, omitting empty number series but not treating remaining royals
as a reopened number series. Respect Infiltrator's no-direct-Diamonds bypass,
Baron exceptions and individual-card limits. The exact-match display is two
labeled totals joined by `=` when equal, or `≠` with a correction message.
Do not introduce overpayment, suggest illegal targets or silently select extra
Clubs during ordinary selection. Negotiator's explicit forced adjustment may
use committed and unused Clubs to maximize an exact match; display that
authoritative change and payment. Initial protection and two-current-series
eligibility apply to all attacks, including combo and ace attacks.

The next legal decision should be understandable before opening a rule book.
Show the chosen cards, affected owner and immediate consequence first; expand
prerequisite detail through `Why available?` / `Why unavailable?`. Keep every
legal alternative discoverable, even when one action has the strongest visual
weight. Do not auto-select an opponent, silently spend a card, replace exact
matching with a recommendation, or use “best move” copy as if strategy were a
rule. A local selection/inspection does not consume the opening allowance.

For opening, show `Open one series` and `Open or reconfigure one combo` as
alternatives under the ordinary allowance. The first opening includes all held
numbers of that suit. Later additions, including restoring a historically opened
but empty series, do not consume the opening allowance. Show current supporting
number suits separately from combo card artwork: a remaining royal cannot keep an
enabling number series alive. Underground's unlimited combo opening/reconfiguration,
two-series waiver and Hearts exemption are separate named exceptions; they do
not permit attacks without two current series or remove other combos' support
suits. Aces use the separate closed tray and the actual one-ace-per-round limit,
not a generic repeatable power button.

For an ordinary declared action, preserve this durable pending information in
the current action/details surfaces:

```text
Alice declared attack on Bob · Hearts
Committed: Clubs 8 (copy 1) = 8
Target: Hearts 8 = 8
Response order: Bob [responding] → Carol [waiting] → Deniz [waiting]
```

The sequence is seat order beginning left of the declarer, independent of
initiative. Each other active, non-confined player has one permitted response or pass; a response
opens no nested window, and cancellation closes the
original window. Skip confined and departed seats by rule, while preserving
confined Coup access. Never skip an otherwise eligible disconnected seat. Only
legally permitted responses are offered. After all
responses, recheck legality and equality before showing a resolved capture or
elimination. A legal declaration marks committed Clubs used for the turn even
on failure/cancellation; keep resolved costs/effects. The Clubs remain exposed
unless an effect removes them. No retargeting or additional Clubs is allowed
after responses begin except Negotiator's explicit forced adjustment; render
this from authoritative state, not frontend-side accounting.

An effect may need several durable inputs. Bind its UI to the effect and decision
IDs in application state; show the required actor, stage and authorized options.
Display `Waiting for Alice to choose Clubs` separately from `Bob may
respond`. Negotiator can require the attacker's maximum exact Club selection
followed by the defender's matching Spades; added-Club usage, payment, jack
return and cancellation commit together after the final choice. Accepted
activation cannot be withdrawn. Justice and Dexter use their named input stages;
committed random results survive reconnect and retry, without a fresh draw.
Do not hold the UI in a generic loading state while another person decides.
Coup remains available at a permitted waiting boundary, never during an atomic
mutation. The ability lifecycle table supplies costs, allowance/reset ownership
and cancellation outcomes for every preview; cards changing custody, formation,
copies or Fate do not reset the same player's allowance.

Use `Waiting for Bob` without a countdown. The current rules add no action
budget, response clock, automatic pass, trading deadline or turn timer. A
transport timeout may show reconnect status; it is not a game forfeiture.

When another view is open, preserve access to permitted response/pass controls
and legally available instant actions through the baseline's current controls.
Do not show `Wait for your turn` as the only option when an ace, Ponzi or Coup
is usable outside the turn. A game-changing modal must either expose equivalent
permitted controls or yield to those controls without closing the user's unsent
work. Only reveal the local player's legal private actions; another seat's
private instant-action availability must not alter their public seat mark or board artwork.

### Victory declarations must keep their actual timing

- Ordinary thresholds are individual: at least 13 number diamonds or 15 royals,
  plus all four suits historically opened. They are declarations after the
  current action and responses resolve, not automatic wins on possession.
- Show `Declare diamond victory` / `Declare royal victory` from verified
  eligibility. Label progress separately from declaration. Never pool an ally's
  cards or reveal private holdings merely to make a meter appear complete.
- **Coup is immediate**, can occur outside the owner's turn, has its own
  Clubs-only series requirement, cannot use Doppelganger or allied membership,
  and opens no response window. It cannot overturn an already ended game.
- Keep the eligible owner's `Declare Coup` access in a consistent private action
  position on every gameplay view, including inspectors, pending responses and
  relevant dialogs. A modal must include an equivalent permitted Coup control
  or be replaced by a nonmodal inspector; do not make Escape-then-navigation a
  prerequisite. Do not disclose private eligibility to other seats.
- Do not add a warning phase, confirmation dialog, hold-to-confirm duration,
  animation wait or countdown to Coup. Prevent accidental activation through
  clear wording, adequate spacing and stable placement away from routine
  `Pass`/`End turn` controls, not a new rule delay. No global single-key shortcut.
- Resolve accepted declarations in the authoritative order and display the
  accepted result. Client arrival animation or local optimism cannot decide a
  victory race. A currently legal victory intent is not rejected solely for an
  unrelated stale aggregate version; holdings, eligibility and phase are still
  checked against current state. Ordinary thresholds keep their completed-action
  boundary. The frontend follows GAME_RULES’ digital ordering policy.

## Economy, alliances and results

Use a compact transaction record rather than a single large score. Every player inspection
shows these independently, with clear scope labels:

| Field | Example | Display rule |
|---|---:|---|
| Match total | `16` | Say whether the current game's recorded score is included; do not silently mix conventions. |
| Current-game score | `−4` | Signed recorded entries; may be negative. |
| Available points — this game | `0` | Game-scoped spendable/Ponzi-eligible positive balance; never substitute match score or a previous game's closed pool. |
| Outstanding debt | `4` | An already charged obligation, linked to creditor and creation order. |
| Possible ordinary end-game award | `+23` | Conditional estimate with included holdings/settings; not available points or a guaranteed final score. |

A transaction row includes event, gross receipt, allocations by named creditor,
remainder available and authoritative order. For example: `Earned 6 → paid Bob
6 → available 0 → debt remaining 4`. Do not animate a second `−6` score change:
the debt was charged when it arose. Incoming creditor payments are their own
receipts, subject to their debts. Keep exact fractional values: `5.5` may be
used for halves; use a fraction such as `11/3` with an accessible spoken value
when a decimal would round an exact obligation or receipt. An allied Ponzi
divides its gross award equally between the user and their sole partner.

Alliance details name the sole partner and show `Permanent for this game`,
`Coup unavailable` and the governing loan/controller. Successfully resolving an accepted combo loan
forms the alliance; proposals, acceptance alone, ordinary trades, promises and help do not. A loan
acceptance review must make the permanent effect clear. Returning the cards
does not end membership; only the recipient controls the loaned formation.
Previously earned player effects are separate: a successfully resolved Underground loan grants
its recipient a lasting privilege without deleting the lender's earned one.
A player already allied cannot join a different partner, so four players may have zero,
one or two disjoint pairs. Do not draw alliance chains or pool victory counts.

Deal state distinguishes `Draft offer`, `Proposed`, `Accepted action`,
`Completed`, `Declined`, `Withdrawn`, `Expired` and a currently invalid acceptance.
State named parties, exact terms and revision, and who may see them. Unanswered
offers reserve no cards or allowances, create no response window or alliance,
and never block End turn. Permit explicit decline/withdrawal and show expiry at
the originating turn's end, with earlier board-closure/departure expiry. This is
a rule boundary, not a wall-clock countdown or inferred refusal. Terms containing
concealed cards remain private to the named parties.

Acceptance starts the accepting player's ordinary action and response sequence;
it becomes nonwithdrawable while awaiting responses. Only successful atomic
resolution transfers cards and grants alliance/loan-derived privileges. Receiving
a dormant loan does not require its activation support. Reconnect restores the
same offer ID/revision/status; stale acceptance never silently accepts revised
terms or a new turn/game. Keep `Nonbinding promise` separate from these transfers;
the promise itself creates neither an alliance nor credit. Do not add escrow,
automatic acceptance or a new open-card-trading permission.
When no offers exist, show `No pending deals` with `Propose deal` only if legal.

For Ponzi, show the selected eligible opponent's **available positive
current-game points**, gross transfer, exact equal partner shares, each
recipient's own debt allocations and net available receipt. Low, zero or
negative-score targets are not forbidden solely for that reason; zero/negative
available points yield zero with no future claim. Do not add a cap or a
future-income meter. Other reward-sharing promises stay separate and voluntary.

Display one governing Code with its owner, diamond bonus and purchase price.
Separate `Economic settings persist` from `Round-end penalty active/inactive`;
breaking or lending the combo does not reset settings or silently move the
recorded declarer's economic exemption. Loss of its functioning formation stops
that formation's recurring penalty; the recipient may legally redeclare.
A later Code replaces the earlier governing Code without discarding its physical
cards merely for supersession.
For Underground, count income using kings allocated to the player's functioning
Underground formation, excluding hidden kings and kings used elsewhere.
Separately show income eligibility at `5+ allocated kings` and retained
benefits below five. Neither its income nor its other benefits requires Hearts;
it does not waive attack restrictions, other combos' support suits or ace rules.

Represent Doppelganger's persistent binding separately from the active formation:
separated bound Kings cannot join other formations or attachments, reunion permits
legal reconfiguration into the recorded slot, and only the named exceptions
reassign/release it. Show only the local seat's authorized binding information.
For Inflation show printed and effective number-Spade values and one active
effect per target. Its removal requires one completed voluntary trade or purchase
spending at least 20 printed Spade points; value that transaction while Inflation
still applies. Forced Negotiator payment and cumulative small trades do not clear it.

The results screen states **Game outcome** and **Match total** separately. An
ordinary ending retains earned points and adds appropriate card scoring; an
ordinary threshold declarer also gets `+50`. Coup replaces current-game scoring
with the prescribed `+50 / −50` result, preserves the defeated/quit exclusions,
completed-game scores and unpaid debts, and adds no ordinary card awards.
Show a reconciliation record, including any carried unpaid adjustment once,
so a score reset is not mistaken for debt forgiveness or a second charge.

Results have distinct **board closed / settlement** and **financially finalized**
states. At board closure, disable gameplay and new victory declarations immediately;
keep the specified financial and promise decisions reachable. Typed promises name
their condition, award ID, recipient and amount/percentage; free text is not a
server-interpreted obligation. Show explicit pay/offer/accept/reject/refuse and
finish-settlement controls as authorized. An unresolved required decision delays
finalization and the next game without a timer; disconnection is not refusal.
Clear active participants' old available pools before Coup awards, while preserving
completed repayments and the defined departed-player exemption. Financial
finalization closes the game's remaining available pool; the next game starts
with zero cash while completed scores and debt survive. A hibernated match can
resume the exact pending state without changing those outcomes.

Setup displays the fixed match game count (default three). Results distinguish
the threshold declarer/Coup winner, each game's score ranking and the cumulative
match result; tied cumulative scores share rank. Do not infer an overall winner
solely from who ended the final game.

Profiles keep the game's named fields—name, trust, anyhoo, scam and games
played—with textual descriptions. These are rule-defined reputation records,
not an invented credit rating or a substitute for game eligibility. Negotiated
promise records must stay visibly distinct from enforced Ponzi distributions.

## Accepted rule boundaries

The [GAME_RULES](../../../../docs/project/GAME_RULES.md) incorporates Q01–Q12 and
the adopted F01–F10/N01–N07/T01–T03 interaction amendments;
the [decision register](../../../../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md)
maps them to future tests. Historical proposals in the [change notes](../../../../CHANGELOG.md)
and earlier reviews do not override those decisions. Render authoritative outcomes
and preserve privacy; widgets must not implement a competing rules engine.

| Accepted topic | Design boundary |
|---|---|
| Failed/canceled attacks | Show committed Clubs used after a legal declaration, retain resolved costs/effects, and disallow retargeting/added Clubs except the explicit Negotiator adjustment. Initially illegal attempts spend nothing. |
| Confinement | Cancel that actor's pending action; only their own active turn ends. Show restriction/immunity until immediately before their next scheduled turn, without skipping it or inventing a countdown. Preserve Coup's exception. |
| Hidden-card victory | Count controlled physical cards once; reveal required proof only on a valid declaration. Invalid online claims stay private and change no state. Server order settles competing online declarations. |
| People / Great People; Doppelganger | Distinguish natural and substituted formation, the one-target limit and Great People's additional eligible-combo protection. Neither protects itself or another People-family combo. Keep each physical King's identity and persistent binding; separated Kings cannot join other formations, and hidden partner location stays private. |
| Negotiator and other card abilities | Show the forced maximum exact-match adjustment and staged choices from authoritative state. Read costs, player-owned quotas/reset and cancellation from the ability table. Justice preserves committed random results and exposes captured number Diamonds; other private captures follow the acquisition matrix. Keep Compensation's remainder and number-supported attachments. |
| Departure and accounting | Quitting/surrender takes effect before the next round; debts and completed payments persist. Resignation restores start-of-game finances only with all affected players' consent. Two players continue; one remaining ends with ordinary scoring and no threshold bonus. Browser lifecycle events make none of these decisions. |
| Deck and numerical Aces | Display the declared 2–10 numerical value while concealing Ace identity; do not count it as an opened series, threshold Diamond or currency. Observe return/shuffle, mandatory short-draw, rejected unsuppliable-purchase and rotating-seat draw-off rules. |
| Restoring an emptied series | Add number cards without spending the opening allowance; current eligibility returns only with number cards. Preserve opening history and lost initial protection through Fate. |
| Alliances and Ponzi | Show exclusive permanent pairs and reject a second partner; an allied user's award splits equally with the sole partner before each recipient's debt settlement. No card/score pooling. |
| Debt and reputation | Show exact gross receipts, oldest-debt repayments and FIFO creditor receipts without a second score charge. Separate explicit promises, Trust/Anyhoo/Scam outcomes and mandatory Ponzi shares. |
| Disconnects | Preserve a required pending decision until reconnect or unanimous resignation. No network timeout advances a response, passes, forfeits or declares a winner. |
| Board/financial lifecycle | Stop gameplay at board closure, permit the settlement command allow-list, and wait for explicit decisions before financial finalization/next-game admission. Current-game cash closes at finalization and resets before Coup awards as defined; completed repayments and debt are not erased. |
| Visibility and ownership | Public hand and concealed-Ace counts reveal category intentionally; hidden identities and eligible breakdowns remain private. Preserve authorized known history, and distinguish lasting Underground/Code effects from transferred formations. |
| Inflation | One active effect per target, shared effective values for combat/currency/Negotiator, and qualifying-transaction expiry after valuation; target departure/board closure also clean up the public effect zone, while Fate and source departure preserve it. |
| Match outcome | Pin the game count before play, default three; display shared rank for tied match scores and distinguish declaration from score/match results. |

Do not revive declined changes: combat overpayment, Coup warning/response
window, Ponzi cap, Code reversal on breakage, action budget or wall-clock trading deadline.
These are not visual-design improvements.

### Adopted N01–N07 frontend contract

Combat previews break down `A = effective committed Clubs + attack modifier`
and `D = effective frozen targets + applicable defense modifier`. Ordinary
success requires equality; Baron compares against the whole eligible series
with strict greater-than. Advancement defense is Hearts-only; numerical defense
may modify any targeted own number series. Exchanges return actual committed
Clubs, not invented cards equal to virtual value. Show response changes and
revalidation without offering retargeting. Negotiator payment excludes defensive
ace values and restricted cards; its attacker modifier remains fixed.

Display Code's global bonus/price settings for every opponent separately from
per-seat round-end penalty eligibility. Penalty needs the functioning declarer,
two current series and no Confinement; initially protected, confined and departed
opponents are excluded. Global settings persist under their own declaration rule.

Resolved Inflation and Compensation occupy a public active-effect zone with
source, target, physical custodian and expiry. The target is custodian, with no
right to replay or transfer the Ace. Fate excludes these cards from return/redraw
and preserves effects/counters. Inflation's caster departing does not end it;
target departure or board closure discards the affected Ace under the rules.
Do not animate an active Ace into a Fate hand or count it as a concealed draw.

A Barricade-drawn card has `available_from_round`. Show the local controller its
restriction and reason; no voluntary play, cost payment, opening/attachment,
trade/return or Coup may use it early. It still counts for ordinary possession
victory and scoring. Mandatory captured/Fate Diamond exposure occurs without
clearing the restriction. The marker follows forced capture/theft, Fate, shuffle
and redraw; Kidnapper cannot override it. Keep concealed metadata private;
next-round expiry is authoritative, never an animation or local clock.

Results show cross-game forgiveness as an exact nonspendable match adjustment
linked to the original finalized charge, without rewriting that game's score.
It survives later Coup and is reversed with debt if its creating game is
nullified. Same-game score corrections instead follow their replaceable charge;
no extra Coup credit. A correction adds neither spendable cash nor creditor
income. The UI renders the authoritative breakdown and never rebuilds the ledger.

### Adopted T01–T03 frontend contract

The proposal controls above follow GAME_RULES §2.15; exact party-visible terms,
revision and origin turn distinguish an unanswered offer from a committed action.
Never display an alliance or loan-derived Underground privilege before the
successful transfer event. A canceled transfer can retain committed Ace usage
without moving cards or granting either benefit.

Compensation draw animations follow ascending fixed seat order, completing each
recipient's quantity before the next. Show only authorized card faces and count
changes. Exhausted earned draws do not leave a later entitlement. A partially
processed automatic queue is not an ordinary-action or ordinary-victory boundary;
restore its committed progress after reconnect and preserve the separate Coup rule.

Deck purchase controls use explicit buyer-chosen quantity, physical payment,
effective price/value and excess-without-change. Eligible payment comes from the
buyer's unrestricted number Spades in hand or exposed series without Underground.
The server validates full supply including returned payment, then revalidates
after responses; never show a partial purchase or automatically reduce quantity.
Concealed payment selections remain private until publicly returned. At price 5,
one Spade 10 can buy a chosen one card even when returning it is the sole supply.

## Review checklist

- Preserve v2 class-specific composition and exact physical cards before/after actions.
- Compare authoritative permissions, timings, effect ownership and exact values.
- Verify visible labels, semantics, history and reconnect share the same authorized projection.
- Treat propositions, accepted actions, responses, results and settlement distinctly.
- Keep immediate legal declarations reachable without presentation-induced delay.
- Use [validation](validation.md) and the app's regression suites; a written contract is not runtime evidence.
