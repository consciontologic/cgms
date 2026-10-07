# CGMS rules implementation contract and resolved decisions

## Authority

[GAME_RULES.md](../project/GAME_RULES.md) is the current rules authority. It incorporates
all twelve accepted recommendations and the owner's amendments: People formations
cannot protect one another, alliances are exclusive pairs, and Negotiator replaces
Forger with a maximal exact-match adjustment that may include unused clubs.
[All four drafts](drafts/) remain historical; [CHANGELOG.md](../../CHANGELOG.md)
records the supersession. This document retains the Q IDs for traceability and
translates accepted rules into required tests. No runtime or test implementation
is claimed by resolving the decisions.

Exact matching, uncapped Ponzi, immediate Coup and persistent Code settings remain
accepted. Action budgets, wall-clock trading deadlines and automatic timeout forfeits remain
outside the rules. None of Q01–Q12 remains an open rule-design blocker.
The owner subsequently accepted all recommendations in the
[first implementation review](../reports/2026-09-28-final-draft-implementation-review.md).
F01–F10 below record the adopted interaction amendments and operational gates;
they refine the twelve decisions instead of reopening them. Adoption is not
evidence of implementation, balance or production readiness.
The owner also accepted N01–N07 from the
[second assessment](../reports/2026-09-28-final-draft-second-assessment.md).
T01–T03 from the [third assessment](../reports/2026-09-28-final-draft-third-assessment.md)
are accepted as well and govern proposal lifecycle, deferred draws and purchases.
The dated reports remain historical evidence; the accepted contracts and shared
fixtures below supersede their proposal status without claiming runtime completion.

## Accepted invariant and test matrix

Each scenario requires a deterministic domain test, followed by persistence and
retry tests where an operation changes durable state. References name sections of
GAME_RULES. The decision register below supplies the detailed expectations.

| Contract | Rule reference | Required test evidence |
|---|---|---|
| Two decks, distinct physical cards | §§1.1, 1.7, 5 | Duplicate rank/suit cards remain distinct; every physical card occupies one location after a resolved transition. |
| Initial deal and initiative | §§1.4–1.5 | Any diamondless initial hand redeals every hand; tie cards stay out until places resolve; round order stays frozen; exhausted supplies use Q07. |
| Historical opening differs from current holdings | §§0, 1.6–1.7 | Losing a second series never restores attack protection; losing it prevents initiating attacks; all-four-suit history survives loss and Fate; restoring an old series consumes no opening allowance. |
| Attack checks apply to every attack source | §§1.7, 2.8, 3 | Clubs, combos, non-combos and aces reject protected targets and attackers with fewer than two current series; Underground provides no bypass. |
| Opening allowance and enabling suits | §§1.6, 2.1, 3.2 | One new series OR one combo opening/reconfiguration per turn; Underground waives combo allowance/two-series requirements and its own Hearts support only, preserving other prerequisites. |
| Exact combat and per-card usage | §§1.7, 2.5–2.10 | Exact totals and suit order; Infiltrator cannot bypass to Diamonds; legally committed clubs count as used even on failure/cancellation; illegal declarations spend nothing; no retargeting or card changes except Negotiator's defined adjustment. |
| One response sequence, staged effect decisions | §2.9 | One permitted response or pass per eligible other player in seat order; omit confined/departed seats but preserve eligible disconnected waits; a response may need several actor-specific inputs without nesting; persist stages and committed randomness; cancellation closes the pending action. |
| First valid declared victory | §§1.2, 2.9, 4 | Thresholds do not win automatically; controlled hidden cards count with sufficient public proof; invalid claims change/reveal nothing in digital play; no response to verification. |
| Coup is immediate and separate | §§1.2, 3.2, 4 | Clubs-opening eligibility suffices; alliance membership and Doppelganger invalidate Coup; no counter; +50/−50 replaces current-game awards while completed games remain. |
| Exclusive permanent alliance pairs | §§1.8, 2.3–2.4, 2.15 | Successful resolution of an accepted combo loan forms one exclusive pair; proposal/acceptance alone does not. Already-allied players cannot join another; returning cards preserves the pair; only recipient controls/counts loaned cards. |
| Allied Ponzi distributes before repayment | §§2.4, 3.2, 4.2 | Steal available positive current-game points once without a cap; split into halves for exactly the user and partner; zero/negative targets yield zero; debts do not reduce another's gross share. |
| Protection and component exclusivity | §§2.1, 2.11, 3.2 | Natural Great People versus substituted People have distinct protection; one target, no self/People-chain protection; no component overlap; two replacement kings fill one slot. |
| Hidden random selection | §§3.2, 3.4, 5 | Kidnapper chooses randomly among eligible hidden aces/royals and permits open selection only when none exist; Justice obeys its rank exclusions; selection/projections conceal unselected identities. |
| Code persistence differs from active penalty | §§3.2, 4.1, 5 | Non-attacking declaration sets all opponents' economics; hostile round penalty requires declarer's two current series and excludes protected/confined opponents. Breaking or lending the governing combo preserves settings/declarer exemption but stops its recurring penalty; a later legal declaration replaces settings/owner without stacking or discarding old physical cards. |
| Underground income differs from other benefits | §§3.2, 4.1 | Income counts only kings allocated to the player's functioning Underground, stops below five and resumes at five; earned privileges remain player-owned, including after loan, and a valid recipient earns their own; Hearts loss never disables it. |
| Score, available points and debt are distinct | §4.2 | Reproduce accounting examples; exact rationals; oldest debt first; gross earnings once; no second debtor subtraction; creditor gains only actual receipts. |
| Debt survives games and Coup, not separate matches | §§4–4.2 | Preserve unpaid obligations within the match; carry only unpaid adjustments removed by Coup; completed payments stay; no double −50; new matches start clean. |
| Ordinary ending and round receipts | §§1.10, 3.3, 3.5, 4.1 | Round 13 ends normally; eligible round receipts, end-game diamond/spade-royal values and one threshold bonus apply; Coup omits those end-game awards. |
| Financial phase and game identity | §§4.3–4.4 | Board closure stops cards/victory but permits specified settlement commands; origin-game available balances close on finalization; next game starts at zero cash after required finishes, preserving scores/debts. Every game-scoped command binds an immutable instance; duplicates replay before game checks, new wrong-game intents reject, and nullified replacement IDs differ. |
| Match outcomes | §§1.1, 4 | Visible fixed game count chosen before start, default three; report declaration winner separately from score leader, and shared match rank for equal cumulative scores. |
| Nonblocking proposals and atomic accepted transfer | §2.15 | Origin-turn offer expiry, private revisioned terms, withdrawal/decline/acceptance races, acceptor response actor, committed Ace allowances and final revalidation produce no partial transfer or early alliance/privilege. |
| Ordered deferred Compensation draws | §2.9 | Fixed ascending-seat queue, complete each recipient's quantity, counters consumed even when exhausted, durable random progress and Coup versus ordinary-action boundaries. |
| Chosen deck purchase quantity | §2.16 | Positive chosen quantity, unrestricted hand/exposed number-Spade payment, exact Code/Inflation value, full supply including returns and response-before-atomic-resolution prevent partial purchase. |

## Resolved decision register

All entries are **resolved — accepted for GAME_RULES, implementation unverified**.
The tests below are acceptance obligations, not evidence that a backend exists.
Keep these IDs in scenarios, implementation work and future rule revisions.

### Q01 — Failed or canceled attack settlement

- **Rules:** §§1.7, 2.8–2.10.
- **Decision:** A legal declaration spends each committed club's use for the turn, even on failure or cancellation; the cards remain unless an effect removes them. Preserve paid costs and resolved responses; do not charge success-only costs after failure. An illegal declaration spends nothing. No retargeting or changing clubs after responses start, except Q08 Negotiator's required adjustment.
- **Required evidence:** Reject an illegal declaration without usage/cost changes; cancel a legal attack and retain usage and paid costs; resolve Advancement changing defense from 8 to 18 without additions or retargeting; verify unused clubs remain usable later and success-only costs are omitted on failure. Cover Negotiator's explicit exception separately.

### Q02 — Confinement outside the active turn and expiry

- **Rules:** §§2.9, 3.4.
- **Decision:** Cancel the target's pending action. End the current turn only if that target is the active player. Action restriction and attack immunity expire at the start of the target's next scheduled turn, before they act; no extra skipped turn. Coup retains its explicit exception.
- **Required evidence:** Confine Bob's out-of-turn Ponzi during Alice's turn, preserving Alice's turn; confine the active player and end their turn; verify immunity/restriction until but not during the next scheduled turn, and verify Coup's exception. N03 skips confined/departed response seats without auto-passing disconnected eligible players.

### Q03 — Hidden holdings and victory proof

- **Rules:** §§1.2, 2.4, 4.
- **Decision:** Controlled hidden cards count. Reveal sufficient qualifying cards to prove a valid threshold and verify four-suit opening history; each physical card counts once. Digitally reject invalid claims privately without changing state or revealing cards. Tabletop invalid claims do not end play; information already revealed cannot be retracted. Verification opens no response window.
- **Required evidence:** Valid hidden-card proof reveals only the selected sufficient set; loaned cards count for the controller only; duplicate physical cards cannot be claimed twice; history is still required; invalid claims leak no identities; a later declaration cannot replace a terminal result.

### Q04 — People, Great People and Doppelganger components

- **Rules:** §§2.1, 2.11, 3.2.
- **Decision:** Natural twin Hearts queens form Great People; a Hearts queen plus a Doppelganger replacement forms People. Either protects one controlled Clubs, Spades or Diamonds series from club attacks including Baron; protected Clubs cannot attack. Great People may instead protect one other combo from kidnapping/assassination; People cannot. Neither can protect itself, either People variant, Hearts or Suicide Bomb. No People protection chains. Changing target is reconfiguration. Two Doppelganger kings fill one fixed replacement slot in one combo and remain two physical royals for victory; no component overlap or virtual extra card. The recorded assignment survives Fate/returns except the explicit removal exceptions. Kidnapper transfers both replacement kings and permits reassignment, while assassinating one breaks substitution and releases the binding. Under F08, the survivor stays in its current zone/control; a remote hidden survivor is not exposed or moved. Other cards in a broken formation remain; reactivation requires legal reconfiguration.
- **Required evidence:** Natural/substitute protection table, one-target limit, no People-to-People protection in any variant pairing, no self/Hearts/Suicide Bomb protection, protected-Clubs attack rejection, reconfiguration allowance, physical count, assignment persistence through Fate/returns and both removal outcomes.

### Q05 — Exclusive two-player alliances

- **Rules:** §§1.8, 2.3–2.4, 3.2.
- **Decision:** Alliances contain exactly two players. Neither may form another alliance during that game; existing partners may continue lending. A four-player game has zero, one or two disjoint pairs; a three-player game has at most one. Returning cards does not dissolve an alliance. Allied Ponzi splits equally between its user and sole partner; alliance membership still bars Coup. Trading alone creates no alliance.
- **Required evidence:** Form A–B, reject B–C and A–C, permit further A–B lending, permit C–D in a four-player game and start with zero pairs. Validate atomic successful loan resolution after acceptance/responses, custody, pair persistence after returns/departure, exactly two Ponzi recipients and Coup rejection. Thirds are arithmetic fixtures only, not legal multi-partner Ponzi examples.

### Q06 — Reopening and Fate's reset scope

- **Rules:** §§0, 1.4–1.6, 2.11, 3.2.
- **Decision:** Restore a historically opened empty series as adding cards without spending the opening allowance; current eligibility returns only with number cards. Fate resets the controlled playable-card zones: return those physical cards and redraw that count, separating aces and exposing drawn number Diamonds. N02 excludes active-effect Aces from both return and count; N07 preserves each physical availability marker through the reset. Do not guarantee a Diamond or redeal. Preserve scores, debts, alliances, opening history, lost initial protection, usage counters and fixed Doppelganger assignments. Supporting cards lost disable dependent abilities; explicitly persistent benefits remain.
- **Required evidence:** Restore a series before/after another opening without extra allowance; Fate with hidden/loaned/attached cards preserves card count and controller accounting; a Diamondless redraw is accepted; enumerate retained state/counters and disabled versus persistent benefits.

### Q07 — Deck return, exhaustion and numerical aces

- **Rules:** §§1.4–1.5, 3.4–3.5.
- **Decision:** Shuffle returned cards into the draw supply before the next draw; ordinary discards stay separate until the draw pile empties. Mandatory draws take available cards with no later entitlement if supplies exhaust. Reject optional purchases unable to supply their quantity, accounting for payment cards returned first. Finish exhausted initiative ties by rotating seat priority. A numerical ace declares integer 2–10 for one attack/defense while its identity stays hidden; it does not open a series, count as threshold Diamonds or serve as Spade currency. Normal revelation on discard does not invalidate its completed use.
- **Required evidence:** No chosen insertion or top-deck gift; returned-card shuffle and discard recycling conserve cards; partial/zero mandatory draw, purchase affordability/supply including returns, exhausted initiative and rotating tie order; ace range/visibility, one-use lifecycle and all excluded counting roles.

### Q08 — Negotiator and remaining card-operation details

- **Rules:** §§2.8–2.10, 3.2–3.4.
- **Decision — Negotiator (formerly Forger):** While under a club attack, transfer exposed number Spades matching the attack value to cancel it; if the entire unrestricted Spade payment pool is smaller, transfer all of that eligible pool. Apply current values including Inflation; keep already committed combat modifiers and numerical-ace values fixed. If no exact subset matches while that eligible pool is sufficient, the attacker must reselect clubs for the greatest positive exact common value with a defender Spade subset, using originally committed clubs plus otherwise unused exposed clubs. Additions/swaps are permitted only for this adjustment; no new ace may be added. The attacker chooses among maximal club subsets, then the defender chooses an equal-valued Spade subset. Originally committed clubs remain used and newly selected clubs become used. No target change or new response window. If no positive common sum exists, activation fails without transfer, jack return or round-allowance consumption and the original attack continues with its original effects. On success the attacker must accept; return Negotiator to the deck. Only unrestricted Spades enter the payment pool and only unrestricted unused Clubs may be added under N07. These are card transfers, not score payments; retain the once-per-round, one-opponent limit and Spades support requirement. Baron loses its special strength for this negotiation on legal activation; a rejected activation leaves the original attack unchanged.
- **Decision — other abilities:** Dexter has one activation per round across offense/defense: 20+ exposed Diamonds allows one exposed-royal assassination; 1–10 allows blocking one hostile ace including Confinement; 11–19 allows neither. Check threshold before returning one number Diamond as cost; defensive prevention opens no new response window and cannot block Coup. Justice selects exposed targets directly or concealed cards randomly among eligible cards, in fixed seat order excluding already captured ranks; no eligible card yields nothing. Captured number Diamonds are exposed under §2.13; only the recipient sees other newly captured hidden identities. Compensation counts involuntary post-activation losses from exposed series including attached royals, carrying remainders across actions/rounds; draw after the action per five losses in T02 fixed-seat order, excluding trades, payments, sacrifices, combo losses and Fate. Exile attaches to Clubs and Barricade to Hearts with a number card required; Barricade is passive, Exile bypasses it for one attack per round, the same-queen exception persists and unsupported remnants confer no ability.
- **Required evidence:** Negotiator exact payment, shortfall, mandatory greatest-common-sum with additions/swaps, refusal rejection, original/new usage, fixed modifiers/Inflation/numerical-ace values, ties between maximal subsets, no-positive-match no-op preserving round allowance, unchanged target/window, Baron and cost/limit/support cases. Dexter thresholds before cost, shared limit, exposed-only targets, Confinement and Coup; Justice constrained pools and concealed projections; Compensation five-loss/remainder/attachment and every exclusion; Exile/Barricade suit support, passive/active timing, exception and support loss.

### Q09 — Defeat, departure, resignation and forgiveness

- **Rules:** §§1.8–1.9, 4–4.3.
- **Decision:** Defeat is voluntary surrender under quit rules; weak holdings, empty series or negative score never eliminate automatically. Quit takes effect before the next round: forfeit positive current-game score/available points, retain negative results/debts, discard active-effect Aces targeting the departed player under N02, return their remaining controlled cards, leave loaned-out cards with recipients and preserve completed transfers/prior games. Departed players' debts and alliance records remain; later receipts settle normally. Two players may continue; apply boundary departures together, and fewer than two remaining ends by ordinary scoring without automatic threshold bonus. With zero active players there are no end-game card awards. Unanimous resignation restores start-of-game financial state, including repaid starting debts, and needs consent from every affected player including departed ones; the nullified game creates no competitive result or reputation outcome. Forgiveness cancels an identified unpaid remainder by creditor consent (unanimous consent for system penalties), with non-spendable score correction and no reversal of completed transfers. Debts survive games in one match; a new match starts clean.
- **Required evidence:** Surrender timing and no automatic elimination; positive/negative ledgers and both loan directions; departed creditor receipts and pair persistence; two/one/zero remaining; resignation restores the financial snapshot with all affected consent and no reputation outcome; partial repayment then forgiveness, system consent, Coup carryover and separate-match reset.

### Q10 — Cyclic debt receipts and settlement termination

- **Rules:** §4.2.
- **Decision:** Calculate simultaneous gross awards first, admit receipts in fixed seat order, then process creditor receipts FIFO; each recipient pays oldest debt first. Finish settlement before another game action. No automatic reciprocal-debt cancellation, rounding, dropped obligations or arbitrary transfer cutoff. Each positive payment reduces outstanding debt and creates no new obligation. A finite set of exact rational debts/receipts has a common denominator D, so every positive reduction is at least 1/D; this supplies a finite bound. Efficient batching must preserve the same allocations, ordering and results.
- **Required evidence:** A and B each owe the other 1; A receives 1/3; both debts end paid and A retains 1/3. Add A→B→C→A, multiple oldest debts, simultaneous paired Ponzi/Coup awards and exact deterministic score entries. Prove optimized/unoptimized equivalence and measure worst-case cost; a technical limit must preserve recoverable state, never forgive obligations or claim a completed settlement.

### Q11 — Digital declaration admission and reconnect policy

- **Rules:** §§1.2, 1.9, 2.9, 5.
- **Decision:** One authoritative server order establishes the first valid declaration; client timestamps do not. Coup can interrupt a pending response or effect-decision wait, never an ended board or an atomic mutation; an already-confined player retains this explicit exception. Current-state victory admission does not reject solely for an unrelated version change within the named immutable game instance; N05 rejects a new wrong-game intent after original-result duplicate lookup. Ordinary thresholds still require a completed-action boundary and an actor who is not confined. Ordinary-action priority cannot delay valid victory declarations; ordinary thresholds still require resolved actions. Untimed disconnection/silence preserves pending state for reconnection or unanimous resignation, without automatic pass, quit, forfeiture or wall-clock turn/trade deadline. Unanswered proposals still expire at their origin-turn boundary under T01. Automatically ending abandoned public games would require a future rule change.
- **Required evidence:** Competing valid declarations yield one winner; Coup interrupts every pending-response position but not terminal state; ordinary-action priority does not starve declarations; disconnect before/after commit and reconnect preserve cursor/costs; indefinite absence changes no game state by itself.

### Q12 — Reputation agreement and settlement events

Apply GAME_RULES §4.3 after game end and settlement of all triggered promises for the ordered pair. Rejected final offers/refusals take Scam precedence over other payments; otherwise aggregate full payment earns Trust, otherwise accepted/paid aggregate delivery of at least 40% earns Anyhoo. Pending promises defer classification; zero realized amounts earn none. Cover multiple named awards and mixed payment/refusal outcomes.

- **Rules:** §§1.8, 4.3, 5.
- **Decision:** Record an explicitly accepted promise with recipient, amount or named reward percentage and condition. Forty percent means 40% of the promised amount. Outcomes are exclusive: full actual transfer earns trust; an accepted/paid reduced offer of at least 40% earns anyhoo; a rejected reduced final offer or refused agreed transfer earns scam; accepted payment below 40% earns neither positive label. Settle payer debts before determining offers; recipient gross receipt counts even when redirected into their debts. Mandatory Ponzi shares earn no reputation. Aggregate promises for at most one outcome per ordered payer/recipient pair per game; an unmet explicit condition earns none. Promises stay voluntary, not automatic transfers or player debt. Disconnection does not classify a pending promise as refusal; nullified games have no reputation outcome.
- **Required evidence:** Accepted promise identity/amount/condition, percentage denominator and exact 40% boundary, mutually exclusive outcomes, payer and recipient debts, no Ponzi reputation, ordered-pair aggregation, pending disconnected promises, nullification and unmet conditions including Coup. No profile mutation is complete until durable event/retry tests pass.

## Accepted interaction amendments and scenario pack

F01–F10, N01–N07 and T01–T03 are **accepted; implementation and cross-layer verification pending**.
The canonical mechanics are GAME_RULES §§2.5, 2.8–2.16, 3.2, 3.4 and 4.2–4.4. This
register supplies scenario identities for the [roadmap](../planning/ROADMAP.md),
not a second implementation checklist. Each trace must specify initial state,
inputs and authorized actor, card zones, support and counters, exact score/cash/
debt values, phase and decision IDs, committed event order, and every seat's
observation. Reuse the same vectors in pure-engine/simulator, server commands,
restart/reconnect and frontend journeys; do not independently invent expectations.

| ID | Accepted contract | Required cross-layer scenarios |
|---|---|---|
| F01 | An accepted effect persists effect/decision IDs, required actor, kind, stage and authorized options; no withdrawal. Several inputs remain one response. Coup may interrupt a wait between atomic transitions. Negotiator final added-Club usage, Spade transfer, jack return and cancellation commit together. Justice samples/reserves each hidden result durably and transfers cards with the queen return only at final resolution; Coup preserves acquired knowledge but abandons uncommitted movement. Dexter prevention is its own effect stage. | Duplicate/reordered/stale and wrong-actor input; reconnect at every stage; Coup at every wait; same committed random result after retry; no duplicated card, cost or use. |
| F02 | §2.12 is the explicit timing/support/cost/quota matrix. Accepted legal activation consumes its allowance; success costs wait for successful resolution unless explicitly earlier. Infiltrator's committed jack return and Fate's pre-shuffle king return persist. Ordinary bomb destruction needs A≥5 under N01 and returns its committed physical Clubs on success; virtual modifiers create no extra return cost and Baron waives Club removal. Generic quotas use player/ability/turn-cycle, initially game start and resetting at that player's scheduled turn; named round/game limits override. Passive effects have no activation quota. Ordinary Clubs retain physical-card usage. Ponzi starts only at idle action boundaries, including out of turn. | Cancellation of every consumable ability; illegal activation without cost; duplicate copies, loan/return, reconfiguration and Fate cannot reset player allowance; ordinary versus Baron bomb; explicit legal/illegal phase and support cases including non-combos' named support and confined Coup. |
| F03 | Authorize, lock, deduplicate and check the named game instance before current victory eligibility. Within that game, unrelated aggregate version changes alone do not reject Coup or ordinary claims at a legal completed-action boundary; ordinary selected commands retain optimistic concurrency. | Stale unrelated pass then valid Coup succeeds; lost queen, new alliance, closed board and ineligible ordinary boundary reject; duplicate result unchanged; two declarations create one board ending. |
| F04 | §4.4 governs complete financial fields. Coup clears active participants' available cash before replacement awards, preserves completed repayments and carries only surviving unpaid adjustments once; departed players remain exempt. Ordinary finalization closes origin-game cash; next game starts with zero cash and retained scores/debts. | Bob has 20 cash before Coup and receives a fresh −50 obligation after cash clears; partially repaid prior-game debt, departed creditor, no duplicate adjustment; new-game Ponzi sees no prior-game cash; departure/nullification match every field. |
| F05 | Separate playing, board-closed settlement and financially finalized. Only specified origin-game financial commands survive board closure. Use typed conditions and named award IDs; finish requires explicit resolution of triggered offers/promises, with no automatic refusal on disconnect. Start the next game only after required finishes and automatic debt settlement. | Award then promised sharing; recipient debt redirection; final accept/reject/refuse and finish; duplicate after board closure; pending reconnect; reject early next-game and late finalized-game transfer; reputation precedence and unmet conditions stay Q12. |
| F06 | §2.13 is the acquisition table. Captured number Diamonds enter exposed Diamonds, overriding Justice privacy; other captured numbers/royals go to hand and Aces to their concealed area unless a named rule overrides. Trade and loan/return follow their own rows. Public history survives concealed custody. | Hidden-Diamond Justice capture becomes public; ordinary/Negotiator Spades received in hand do not open the suit; public royal into hand remains historical knowledge; free-card trade versus intact exposed loan/return; support, loan acceptance timing and surviving-card return without remote reclaim or auto-reformation. |
| F07 | Lasting effects belong to players independently of custody. Underground privilege remains for its earner and a successfully completed loan grants its recipient their own; income uses only kings allocated to their functioning formation. Code settings/declarer exemption persist until legal redeclaration; lending stops the old formation's recurring penalty without silently moving the exemption. | Underground loan/return/breakage, both players' lasting privileges, hidden/unallocated kings excluded from income; Code loan/breakage/return/redeclaration and unchanged versus replaced owner; quota identity never follows cards. |
| F08 | One king belongs to at most one binding. Separated bound kings remain physical royals but cannot join other formations/attachments; reunion permits legal reconfiguration into the original assigned slot. Fate/Justice/transfers preserve binding; special Kidnapper permits reassignment; assassination releases it. | Justice/Fate split, forbidden third-king pairing/attachment, reunion with legal allowance, assassination/recycling and Kidnapper reassignment; substituted Kidnapper returns all allocated physical cards; binding release does not expose or move a remote survivor and projections cannot locate the other hidden king. |
| F09 | At most one Inflation per target; reject duplicates without consumption. Halve number-Spade values for currency/combat/Negotiator. A completed voluntary trade/deck purchase spending at least 20 printed Spade points clears it after pricing that transaction while active. Forced payments and cumulative small trades do not clear it; Spade-royal awards stay 10. | Second ace; odd exact fractions; printed 19 versus 20; canceled transaction; forced Negotiator payment; several small trades; qualifying purchase then subsequent full-value purchase. |
| F10 | Public hand size and separate concealed-Ace count; faces and hidden royal/eligible-card counts private. The same §2.13 visibility matrix governs state, events, history, accessibility, bots and reconnect. | Ace/non-Ace draws change only authorized counts; permutation of a private face within its visible category preserves every observer projection; no tooltip/label/log or binding metadata leak. |
| N01 | Combat uses `A = effective committed Club sum + attack modifier` and `D = effective frozen target-card sum + applicable defense modifier`; ordinary success requires equality. Baron compares against the complete eligible series using its strict greater-than rule. Advancement defense applies only to Hearts; a defensive numerical ace may modify the targeted subset of any own number series. Exchanges return the committed physical Clubs; virtual values create no payment cards. Negotiator retains committed modifiers. | Club 5 plus Advancement 10 versus Barricaded Hearts 7+8 returns the physical Club 5 only; defensive numerical/Advancement changes invalidate a frozen match; Dexter prevention, Baron, Inflation and Negotiator use the same calculation and separate physical usage/cost assertions. |
| N02 | Resolved Inflation/Compensation Aces occupy a public active-effect zone with source, target, physical custodian and expiry, without reusable gameplay control. Target is custodian, Compensation self-targets. Fate preserves these effects/cards and excludes them from reset/draw counts. Source departure does not end another target's Inflation; target departure discards that target's active Aces, and board closure cleans up all active Aces after required end evaluation. Named expiry still applies. | Fate on caster/target and Compensation with remainder preserves the exact effect, card location and redraw count; caster versus target departure; ordinary/Coup closure and eventual recycling conserve all 104 cards without new Ace allowance or reuse. |
| N03 | Omit confined and departed players from ordinary response traversal; remaining seats keep fixed order. A skipped player receives no required pass decision. Disconnection alone never removes an otherwise eligible responder; legal confined Coup remains an independent exception. | Out-of-turn confinement followed by another player's action; departed seat; disconnected eligible seat still pending; reconnect does not undo a skip; scheduled confinement expiry restores eligibility. |
| N04 | Code declaration is a non-attacking global economic action governed by its formation/Clubs prerequisites and usual Underground waiver. Its settings affect all opponents. The recurring −5 is hostile: require the active, non-confined declarer's two current series at round end and exclude initially protected or confined opponents. Existing ownership/custody and functioning-formation rules remain. | Mixed protected/unprotected/confined targets receive correct prices/bonuses and penalty exclusions; protection loss and confinement expiry; Underground declaration waiver without penalty prerequisite waiver; Code break/loan/return/redeclaration. |
| N05 | Every game-scoped intent, including victory and settlement, carries immutable `game_id`; dependent turn/window/action/effect/decision IDs belong to it. Authorize and deduplicate first, then reject a new wrong-game intent before current eligibility; unrelated-version victory tolerance applies only within that named game. Every next game/nullified replacement has a fresh authority-generated ID, even at the same configured slot; queued intent never retargets automatically. | Delayed Game 1 first-submission Coup rejects in Game 2; committed Game 1 duplicate returns original result; stale ordinary selection rejects; nullified replacement ID differs; same-game legal Coup after unrelated pass succeeds; changed-body reused command ID conflicts. |
| N06 | Cross-game forgiveness appends a nonspendable match adjustment against a charge retained in a completed result, linked to debt/charged-entry lineage and the forgiveness-creating game. It survives subsequent Coup without rewriting the old result or creating cash/repayable receipt. Current-game forgiveness instead corrects its current charge, so Coup replaces both together. Nullification restores start debt and effective adjustments while retaining void audit records. | Game 1 score −10 and debt 10, Game 2 forgiveness 10 then Coup +50 produces cumulative 50 (−10+10+50), cash 50 and debt 0 absent other entries; partial forgiveness after partial repayment; current-game charge/correction both replaced; carried-charge lineage; nullification reverses that game's correction and restores debt; duplicate has one effect. |
| N07 | Each Barricade draw gets physical `available_from_round`; before then it counts for ordinary thresholds/proof but cannot voluntarily trade/pay/open/attach/activate or supply Coup. Mandatory captured/Fate-drawn Diamond exposure still occurs. Restriction survives involuntary movement, Fate, shuffle and redraw, overriding Kidnapper immediate use; hidden metadata stays private. | Missing Coup queen remains unusable until next round; ordinary hidden threshold counts; trade/payment/attachment reject; forced Diamond exposure remains restricted; theft/Fate/recycle/redraw preserve marker and release at recorded round; projections do not track concealed marked cards. |
| T01 | §2.15 proposals are private, nonblocking revisioned records with immutable offer ID and origin game/turn. No pre-acceptance cards/allowances are reserved. Withdraw/decline serialize against acceptance; unanswered proposals expire at origin-turn end, board closure or party departure. Acceptance checks exact terms/revision, eligibility, ownership, availability, transfer quotas and alliance compatibility at an idle boundary. The accepting player is transfer actor for responses/Confinement; supplied Aces consume their owners' allowances. Final revalidation and atomic successful transfer alone create an alliance/loan-derived privilege; receipt of a loan needs no activation support. Accepted actions cannot be withdrawn; cancellation/failure closes without transfer, keeping prior committed costs/responses. | Silence plus End turn; acceptance versus withdrawal/expiry; changed revision; stale game/turn; cards moved/restricted; loan receipt without ability support; Confinement of acceptor on another player's turn; canceled loan gives no alliance/Underground privilege; duplicate/reconnect preserves status/action; private terms stay private. |
| T02 | After action/responses and required returns/shuffle, queue earned Compensation draws by fixed ascending seat, complete one recipient's quantity before the next, then settle points. Each draw uses normal recycling/exhaustion; consume earned five-loss entitlements even without supply and keep only the remainder. Persist cursor/counters/random outcomes if staged. No ordinary action/victory interleaves; Coup between atomic steps preserves completed draws and abandons the rest on closure. | Seats 2/3 each reach five losses with only returned Justice queen available: seat 2 gets queen, seat 3 no card/entitlement, both counters zero. Also adequate distinct cards, multiple draws per seat, reversed initiative/target order, recycling, crash/replay, authorized projections and Coup at every staged boundary. |
| T03 | §2.16 buyer chooses integer quantity q≥1 and exact unrestricted number Spades from hand or exposed series at an own-turn idle boundary; no Underground or generic two-series prerequisite. Effective payment ≥ q×Code price with Inflation applied, no change. Check supply including returned payment before acceptance and again after responses. Buyer is actor; payment is resolution-only. Success atomically returns/shuffles payment, draws exactly q and clears qualifying Inflation after pricing; cancellation/failure pays/draws nothing and never resizes q. Hidden payment becomes public only on return. | Price 5: payment 5/9 buys one; 10 buys chosen one/two; empty piles plus single Spade 10 supports q=1, rejects q=2 without payment. Hand/exposed payments, Code price, odd Inflation fractions and removal threshold, restricted cards, changed supply after responses and retry preserve quantity, conservation and privacy. |

Intact combo lending/return is a custody transfer, not a new opening or
reconfiguration; it does not consume an opening allowance. Acceptance is at a
clear action boundary on either party's turn and starts the T01 transfer action;
support is not needed merely to receive an intact loan, but ability use still
needs its support/timing/quota. Success alone grants loan-derived privileges.
Returning surviving separate royals exposes them
unassigned, without reforming the old combo or reclaiming cards now held elsewhere.

The review's operational recommendations are accepted as implementation gates:

- Preserve resumable untimed waits and identify the exact waiting actor; technical
  hibernation must not alter outcomes. Before public/monetized release, decide
  abandoned-game allowance/reward handling separately from the accepted no-forfeit
  rule; a transport timeout cannot supply that product policy.
- Exact cyclic settlement is finite but can be expensive: reciprocal debts of 1
  and a receipt of 1/D need 2D unbatched transfers. Prove batched equivalence on
  small ledgers; keep reproducible compressed audit records and durable progress.
  Large-D restart tests must resume rather than repeatedly roll back identical
  work, with no financial cutoff, rounding or forgiveness.
- Pin a visible positive games-per-match count before start, default three.
  Shared rank on tied cumulative scores is distinct from threshold/Coup winner
  and highest game score. Test configuration, immutability, sequential settlement,
  next-game admission and tied final ranks.

## Partial simulator and full-conformance status

Q01–Q12, F01–F10, N01–N07 and T01–T03 are accepted rule/contract decisions
incorporated into GAME_RULES. Dated reports retain their original proposal wording
as historical evidence. The [roadmap traceability](../planning/ROADMAP.md#accepted-review-traceability)
maps all adopted amendments to existing implementation tasks; adoption alone does
not satisfy their executable evidence or complete any runtime gate.

A partial simulator may exercise implemented scenarios under **partial rules
coverage**. An unimplemented GAME_RULES transition is a coverage limitation,
not a reopened Q01–Q12 decision. Fail explicitly before executing unsupported
operations; never fabricate a completed game or reset prior resolved effects.

Retain safe `adjudication-required` handling for a future genuinely undefined
experimental transition, with a named/versioned interpretation and diagnostic
reference. Preserve the last committed state; if earlier responses resolved,
retain their effects and a recoverable paused-action marker without exposing
hidden identities or charging costs twice. Unknown/incompatible ruleset or
engine versions fail validation before a game starts. Experimental profiles
cannot certify GAME_RULES conformance or create real rewards/reputation.

Full GAME_RULES conformance and online readiness remain **unverified pending
implementation, rule fixtures, persistence/fault tests and release gates**.
Accepted rule decisions do not demonstrate those engineering outcomes.
