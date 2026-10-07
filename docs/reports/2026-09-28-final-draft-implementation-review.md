# FINAL_DRAFT implementation review — frontend and backend

Date: 2026-09-28. Status: advisory review; recommendations have **not** been adopted as rules.

## Assessment

**The draft is a usable design baseline, but it is not yet a complete implementation contract.** The previous twelve decisions settled their named questions. They did not settle every interaction between those decisions, the inherited card rules, and the proposed online protocol. A frontend and backend built independently from this text can still disagree on legal actions, card locations, money and information disclosure.

I found **eight P1 issues and two P2 issues** below. P1 means resolve the issue before implementing the affected behavior and its expected-result tests; it does not mean all simulator work must stop. P2 means a narrower contract needs clarification before full conformance. These are specification findings, not observed production bugs or measured game-balance results.

The accepted changes remain intact: exclusive two-player alliances, no People-family protection of another People formation, and Negotiator's maximum exact match using committed and unused Clubs. I am not recommending reversing them.

### Evidence and limits

Reviewed the entire [final draft](../design/drafts/FINAL_DRAFT.md), the [Q01–Q12 register](../design/RULES-IMPLEMENTATION-QUESTIONS.md), [architecture](../code/ARCHITECTURE.md), [API](../code/API-game.md), [state/delivery design](../design/DESIGN-go-state-and-delivery.md), [simulation contract](../../sims/README.md), relevant [frontend information/interaction guidance](../../.agents/skills/cgms-art-direction/references/gameplay-layout.md), and roadmap. Separate reviewers examined frontend disclosure, backend transitions, and card-state interactions; the coordinating review examined financial boundaries and consolidated overlaps.

The reviewed `FINAL_DRAFT.md` has 440 lines and SHA-256 `cecea3335b163ad0dddc51857f64023b643437014a1566739d2cc3fdad828c21`. Line references below refer to that snapshot. The rule file was not changed by this review. The repository contains proposals, not an application engine to exercise. No Flutter/Go runtime, rendered interface, load test or security certification is claimed. The historical Flutter analysis failure remains unchanged and is not evidence about this rules document.

The Software Architect review method was used: trace user decisions, authority, private information, state transitions and recovery; distinguish a missing rule from a deliberate tradeoff. Findings rely on local source documents and explicit counterexamples, not external platform claims.

## Priorities and recommended fixes

| ID | Priority | Issue | Recommended fix |
|---|---|---|---|
| F01 | P1 | One response can require decisions from several people | Persist effect-specific decision stages, with an explicit Coup boundary. |
| F02 | P1 | Costs and usage limits do not cover cancellation/reconfiguration consistently | Publish a per-ability cost, quota and reset table. |
| F03 | P1 | Generic stale-version rejection can delay a rule-legal Coup | Validate victory declarations against current state without rejecting solely for an unrelated version change. |
| F04 | P1 | Score replacement does not fully define spendable balances | Specify complete ledger transitions for Coup and game boundaries. |
| F05 | P1 | End-game promises need commands after the board has ended | Separate board closure, voluntary settlement and financial finalization. |
| F06 | P1 | Captured-card destinations are incomplete; Justice conflicts with Diamond disclosure | Define destinations and disclosure per acquisition mechanism, with explicit precedence. |
| F07 | P1 | Persistent powers have ambiguous ownership when their combo is lent | Separate lasting player effects from physical card custody. |
| F08 | P2 | Doppelganger can remain bound after its kings are split | Define dormant, reunited and released binding states. |
| F09 | P1 | Two Inflation aces have no defined stacking/removal result | Permit one active effect per target and specify the qualifying removal transaction. |
| F10 | P2 | Public hidden-zone counts are not fully authorized | Approve one visibility matrix for snapshots, events and every client surface. |

### F01 — A response is not always one human decision

**Evidence:** [§2.9, lines 220–226](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order) says responses resolve immediately without nesting; [§2.10, lines 242–244](../design/drafts/FINAL_DRAFT.md#210-negotiator-resolution) requires an attacker choice followed by a defender choice inside one response. Dexter prevention and Justice's successive selections also require inputs. The [architecture, lines 99–103](../code/ARCHITECTURE.md#3-data-flow), principally describes an action and next responder.

**Failure case:** Negotiator finds several maximum club subsets. The attacker disconnects before choosing. The server needs to know who may answer, whether activation can be withdrawn, what has been consumed, and whether another player may declare Coup during that wait. “No nested response window” does not answer these questions.

**Recommendation:** Add a proposed pending-decision record: effect ID, decision ID, required actor, decision kind, stage and authorized options. These are inputs to one effect, not additional response opportunities. Freeze the accepted activation against withdrawal. Permit Coup while awaiting an input, but never halfway through an atomic effect mutation. For Negotiator, original Clubs remain used; commit added-Club usage, Spade transfer, jack return and cancellation together after the final choice. Specify analogous stages for Justice and Dexter, including when random outcomes become committed and visible. Any revealed committed random result must survive reconnect/retry; it must not be rerolled.

**Frontend/backend consequence:** The frontend shows the actual decision owner and resumable selection; the backend can accept exactly that input without holding a database transaction while a human thinks. **Acceptance cases:** reconnect, duplicate input and Coup at every stage; stale decision IDs; no duplicate payment, draw or usage. This extends the already proposed durable response model rather than replacing it.

### F02 — “Keep paid costs” is insufficient without a cost schedule

**Evidence:** [§2.8, line 214](../design/drafts/FINAL_DRAFT.md#28-one-combat-procedure) distinguishes paid and success-only costs. [§3.2](../design/drafts/FINAL_DRAFT.md#32-combos) and [§3.3](../design/drafts/FINAL_DRAFT.md#33-non-combo-cards) still use phrases such as “after use”; [§1.7, line 99](../design/drafts/FINAL_DRAFT.md#17-attacking-trading--interacting-on-your-turn) applies a general once-per-turn rule “where applicable.”

**Failure cases:** Confinement cancels Ponzi: do its four jacks return? Negotiator cancels a Baron attack using Infiltrator: is Baron exhausted and is the jack returned? Can Underground reconfigure used Kidnapper cards into another formation to steal again? An ordinary attack kills Suicide Bomb with at least five points, but the text does not name the ordinary Club-removal cost that Baron's “without a cost” exception waives.

**Recommendation:** Give every ability an explicit entry for timing, prerequisites, declaration cost, resolution cost, cancellation outcome, quota owner and reset boundary. My default would consume an activation allowance on accepted legal activation, while returning/discarding effect cards at successful resolution unless the ability explicitly commits them earlier. Infiltrator's bypass is an earlier committed effect, so its jack return should stand if the underlying attack is later canceled. Fate's king must still return before its specified shuffle. For Suicide Bomb, I recommend committing at least five Club points and returning those attacking Clubs on successful destruction; Baron waives that removal cost. These are proposed rulings requiring adoption, not conclusions already supplied by Q01.

Key ordinary ability quotas by player, ability and the stated turn/round/game interval; changing physical copies, formation, custody or using Fate must not silently refresh the same player's quota. Retain the explicit physical-card exception for ordinary Clubs. Specify the interval for each currently vague entry rather than applying a generic counter to everything.

**Frontend/backend consequence:** Both sides can explain costs before confirmation and display the same remaining uses. **Acceptance cases:** cancellation of every consumable ability; duplicate copies, loan/return and reconfiguration; ordinary versus Baron bomb destruction. Include a legal/illegal timing matrix rather than assuming every harmful effect follows identical timing.

### F03 — The proposed version check can change a victory race

**Evidence:** [§2.9, lines 220 and 234](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order) gives Coup immediate processing between effects. The [API, lines 37–45 and 77](../code/API-game.md#errors) and [transaction design, lines 29–36](../design/DESIGN-go-state-and-delivery.md#proposed-design) apply a generic expected-version check and resubmission policy.

**Failure case:** Alice submits Coup from version 42. An unrelated pass commits version 43 first. Alice still owns the required queens and remains eligible, but a blanket stale-version rejection makes her resnapshot and submit again. A subsequently processed victory may now end the game first. Serialization alone does not reconcile this with the rule's declaration policy.

**Recommendation:** Use a victory-specific admission contract: authenticate, deduplicate, lock, reject an already-ended game, and validate current holdings/eligibility. Do not reject solely because an unrelated aggregate version changed. Changed cards, alliance membership or other relevant eligibility must still cause rejection. Retain optimistic version checks for ordinary commands whose meaning depends on a stale selection. Apply this deliberately to ordinary threshold claims only at their permitted completed-action boundary.

**Frontend/backend consequence:** A stable Coup control can submit an intent without an unrelated pass creating an extra interaction. This preserves authoritative processing order, not network-arrival fairness. **Acceptance cases:** unrelated stale version succeeds when currently legal; lost queen, new alliance and terminal state fail; duplicate retry returns the original result; two valid declarations produce one ending.

### F04 — Financial boundaries need a complete transition, not only score arithmetic

**Evidence:** [§4, line 358](../design/drafts/FINAL_DRAFT.md#4-end-game-and-scoring) replaces active players' scores on Coup; [§4.2, lines 389–391](../design/drafts/FINAL_DRAFT.md#42-debt-and-automatic-repayment) correctly distinguishes recorded score, available points and debt. It does not explicitly say what happens to existing available points during that replacement or how unspent prior-game points are isolated at the next game. Ponzi may take current-game points only ([§3.2, line 311](../design/drafts/FINAL_DRAFT.md#32-combos)).

**Failure case:** Bob has 20 available points and no debt when Alice declares Coup. Does Bob's new −50 obligation use that 20, leaving debt 30, or does Coup clear available points before imposing debt 50? Both readings can replace recorded score with −50. They produce different future repayments. A single untagged available-balance field also cannot prove that a later Ponzi takes no previous-game money.

**Recommendation:** Write one authoritative transition table for ordinary ending, Coup, departure, nullification and next-game start, covering all financial fields. I recommend game-scoped spendable balances: on Coup, clear active participants' old-game available balances, preserve completed repayments, retain surviving unpaid-debt adjustments exactly once, then apply the new +50/−50 awards and settlement. Exempt departed participants as already required. After ordinary post-game settlement, close that game's available pool; begin the next game with zero spendable points, retaining completed scores and outstanding debts. These cash-boundary choices need explicit adoption; if money is instead intended to carry, preserve origin-game buckets and define which can fund transfers and repayments.

**Frontend/backend consequence:** The results screen can explain exactly why score, cash and debt differ; persistence/replay cannot choose a different reset order. **Acceptance cases:** Bob's 20-point example, prior-game debt partly paid before Coup, a departed creditor, and a new-game Ponzi with no current-game earnings.

### F05 — Promise settlement needs a legal post-game phase

**Evidence:** [§1.2, line 51](../design/drafts/FINAL_DRAFT.md#12-winning-conditions) ends the game on a valid declaration. [§4.3, lines 419–426](../design/drafts/FINAL_DRAFT.md#43-promises-and-reputation) requires actual payments or final decisions about awards, possibly after the game ends. The [API command families and ended-game error](../code/API-game.md#endpoints--methods) do not define the post-game mutation boundary.

**Failure case:** Alice promised Bob 40% of an ordinary ending award. That award exists only when the board ends. If all later commands return `GAME_ENDED`, Alice cannot pay. If generic transfer commands remain legal indefinitely, their attribution after the next game begins and their relationship to immutable completed-game results are unclear.

**Recommendation:** Separate `playing`, `board closed / settlement`, and `financially finalized`. Board closure is immediate and rejects further gameplay/victory actions. Settlement permits only specified financial commands, offer acceptance/rejection/refusal, and reads/reconnect; existing debts still have priority. Attribute transfers to that game, finalize its financial result, then start the next game. Provide an explicit finish-settlement action that requires the player to resolve their outstanding offers/promises, including an explicit refusal where appropriate. Never treat mere disconnection as refusal.

For v1, I recommend this sequential settlement model. It means unfinished human decisions can delay the next game under the accepted untimed policy. Supporting late payments across concurrent games is an alternative requiring separate wallet/provenance rules, not an incidental API addition. Use typed conditions and named award IDs for reputation-eligible promises; free-text bargaining alone should not require the server to interpret natural language.

**Frontend/backend consequence:** Results need a real settlement flow, not just a static winner screen; the API needs a phase-specific command allow-list. **Acceptance cases:** end-game award followed by sharing; recipient debt redirection; finish/refuse; retry after board closure; reconnect while an offer is pending; next-game admission only after financial finalization. Keep Trust/Anyhoo/Scam precedence already adopted in Q12.

### F06 — Capture does not always determine the receiving zone or disclosure

**Evidence:** [§1.4, line 71](../design/drafts/FINAL_DRAFT.md#14-deal-turn-order--draw) requires captured Diamonds to become public immediately. [Justice, line 308](../design/drafts/FINAL_DRAFT.md#32-combos) says only the recipient sees a captured hidden card. Ordinary Spade capture ([§2.6](../design/drafts/FINAL_DRAFT.md#26-attack-order)) and [Negotiator's payment, line 244](../design/drafts/FINAL_DRAFT.md#210-negotiator-resolution) do not name a receiving zone.

**Failure cases:** Justice captures a hidden 7♦: should all players see it or only the recipient? A player with only Clubs and Diamonds receives Spades: automatically exposing them could create a historical opening and current support, whereas placing them in hand does not. Clients can therefore disagree on victory eligibility as well as animation.

**Recommendation:** Publish an acquisition table for draw, ordinary capture, Justice, Kidnapper, Negotiator, trade, loan, return and Fate. My proposed defaults: captured number Diamonds enter the exposed Diamond series, overriding Justice's private-result sentence; other captured numbers/royals enter the hand unless a named rule says otherwise; acquired Aces enter the concealed Ace area. Transfers by trade/loan need their own stated rules, not an accidental extension of capture defaults. Opening and immediate-use exceptions remain explicit. Record already-public identities in authorized history even when their new zone is concealed; concealment cannot erase observed information.

**Frontend/backend consequence:** One transition determines zones, opening history and every permitted projection. **Acceptance cases:** hidden-Diamond Justice capture; ordinary and Negotiator Spade receipts into an unopened suit; public royal theft into hand; authorized history and reconnect show the same permitted information without leaking unselected cards.

### F07 — Persistent effects and loan custody disagree without an ownership rule

**Evidence:** [§2.4, line 170](../design/drafts/FINAL_DRAFT.md#24-what-counts-as-an-alliance) reserves a loaned combo's features to its recipient. [Underground and Code, lines 305 and 309](../design/drafts/FINAL_DRAFT.md#32-combos) preserve benefits/settings after cards are lost or the formation breaks.

**Failure cases:** Alice loans Underground to Bob: does Alice retain unlimited combo reconfiguration while Bob gains it too? Alice lends her declared Code to Bob: whose Diamond bonus/purchase price remains at the default, and who levies the recurring penalty? Also, Underground's “at least 5 kings” does not identify whether income counts allocated, attached, hidden or merely controlled kings.

**Recommendation:** Separate player-owned lasting effects from physical formations. Preserve an earned Underground privilege for its player throughout the game; explicitly grant a recipient their own privilege on accepting a valid Underground loan. This allows both partners to retain privileges and should be an acknowledged consequence. Pay Underground income only from kings allocated to that player's functioning Underground formation, excluding hidden kings and kings used elsewhere. If granting privileges to both is unwanted, that requires an explicit loan exception to persistence.

For Code, keep the recorded declarer and settings until another legal Code declaration. Lending its cards stops that former formation's recurring penalty and does not move its economic exemption silently. The recipient may redeclare Code under the usual rules. Amend the generic “only recipient uses features” sentence to distinguish the transferred formation from previously created lasting effects.

**Frontend/backend consequence:** Show the owner of a lasting rule separately from current card custody; storing only `combo.owner` is insufficient. **Acceptance cases:** loan, return and breakage of both combos; hidden/unallocated kings at round end; Code exemptions change only on the agreed declaration transition.

### F08 — A persistent Doppelganger pair can be separated

**Evidence:** [§2.11, lines 260–262](../design/drafts/FINAL_DRAFT.md#211-protection-and-doppelganger) keeps its assignment through Fate/deck returns. [Justice](../design/drafts/FINAL_DRAFT.md#32-combos) can capture one king, and Fate can later distribute the pair to different people. The assigned-together, assigned-separated and rejoined states are not defined.

**Failure case:** Alice's bound pair is split; Bob gets one king. May Bob use it as Richer or form a new Doppelganger with a third king? What happens if the original pair later reunites? The answer controls legality, not just how the pair is drawn on screen.

**Recommendation:** Add explicit binding states and require a king to belong to at most one persistent binding. My conservative recommendation is that separated bound kings remain physical royals but cannot join another formation or attachment; reunion permits legal reconfiguration into the original assigned slot. Fate, Justice and ordinary transfers preserve the binding; special Kidnapper theft permits reassignment; assassination releases the old binding. This is restrictive for isolated kings, but avoids silently erasing the accepted persistent role. A more permissive policy must explicitly say when other uses are released before reunion.

**Frontend/backend consequence:** Persist dormant binding metadata separately from active formations and reveal it only to authorized viewers; it must not identify who drew the other hidden king. **Acceptance cases:** split through Justice/Fate, attempted third-king pairing, reunion, assassination and recycling, hidden-state projection comparisons.

### F09 — Repeated Inflation changes exact matching unless stacking is defined

**Evidence:** [Inflation, line 338](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces) halves Spades and describes a removal trade. There are two such aces in two decks. [Negotiator, line 238](../design/drafts/FINAL_DRAFT.md#210-negotiator-resolution) uses Inflation-adjusted values.

**Failure case:** Different opponents apply both aces to one player. Is an 8♠ worth 4 or 2? Does a transaction spending 20 printed Spade points remove one effect or both? Does a forced Negotiator payment qualify? These choices change purchase quantity and legal attack/payment subsets.

**Recommendation:** Permit one active Inflation per target; reject a duplicate application without consuming it. Halve number-Spade values for currency, combat and Negotiator. Clear it only after one completed voluntary trade or deck purchase spending at least 20 printed Spade points; calculate that transaction with Inflation still active, then clear it. Do not count forced Negotiator payments or cumulative smaller trades toward removal. Keep spade-royal end-game awards at 10. This is a proposed explicit interpretation, not an already accepted stacking rule.

**Frontend/backend consequence:** The same source/effective value pair drives purchase previews, exact-match validation and the removal indicator. **Acceptance cases:** second ace, odd values, 19/20 printed points, canceled transaction, Negotiator payment and a subsequent uninflated purchase.

### F10 — Public counts need an explicit information policy

**Evidence:** [§3.4, line 328](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces) separates concealed Aces from the hand. The [API, line 61](../code/API-game.md#endpoints--methods) gives other players “counts” without identifying the zones. [Frontend guidance, lines 275–278](../../.agents/skills/cgms-art-direction/references/gameplay-layout.md#concealment-is-a-data-and-semantics-contract) deliberately defers permission to the authoritative contract; its example does not authorize an Ace count.

**Failure case:** One client exposes separate hand and Ace counts while another displays only aggregate hidden cards. Drawing an Ace then reveals its category to one set of observers but not the other.

**Recommendation:** Approve a visibility matrix covering each zone's count, card identity, effect metadata, event details and viewers. My tabletop-faithful recommendation is public hand size and public concealed-Ace count, with faces and the hidden royal/eligible-card breakdown private. This intentionally reveals the Ace category through its count and requires a product/rule decision. If aggregate-only counts are preferred, apply that equally to every client and event.

**Frontend/backend consequence:** The same information policy must cover snapshots, events, history, tooltips, accessibility labels, reconnect, bot observations and logs. **Acceptance cases:** Ace and non-Ace draws expose exactly the approved information; changing a hidden face within an authorized category does not change an observer's projection. Existing private-projection safeguards are valuable but do not select this policy.

## Deliberate choices and engineering risks

These are not additional claims that the rules contradict themselves.

- **Untimed play can remain pending indefinitely.** This is already an explicit Q11 decision. Unanimous resignation cannot rescue a game if a required participant never returns or consents. Preserve resumable state, show the exact waiting actor, and support technical hibernation without changing game outcomes. Any timeout, bot takeover, automatic abandonment result or reputation deadline is a separate proposed rule version. Monetized/public matchmaking needs an explicit abandoned-game allowance/reward policy before release; a transport timeout must not invent one.
- **Finite exact debt settlement is not necessarily cheap.** With A owing B 1, B owing A 1, and an external receipt of `1/D`, the documented unbatched FIFO procedure makes `2D` transfers. An independent arithmetic probe confirmed D = 3, 5, 10 and 1,000. Thus D = 1,000,000,000 implies two billion transfers in that model; this is a derivation, not a benchmark or a claim that normal matches produce that denominator. The architecture already calls for batching and safe pauses. Implement exact batching with reproducible compressed audit records, prove equivalence on small ledgers, and test bounded recovery on large values. A retry must resume progress rather than repeatedly roll back the same expensive calculation. Never round or forgive debt to fit a request timeout.
- **Match/product outcomes need a small final contract.** The simulator already proposes games-per-match input. Pin it in online match setup too, specify tied cumulative-score results, and distinguish the threshold declarer/Coup winner from highest-scoring player and match winner. I recommend a visible, fixed game count chosen before start, with three as the default, and shared rank on tied match scores. This is a proposed product clarification, not a reason to reopen accepted card mechanics.

No measured evidence currently establishes game balance, average session length, first-seat advantage or exploit resistance. Preserve the requested rules, then investigate those questions with the planned simulator instead of presenting review opinions as playtest results.

## Recommended implementation response

Resolve F01–F07 and F09 in a short rules amendment/contract addendum before writing expected results for those behaviors. Resolve F08/F10 before enabling the affected card/projection paths. Keep the current Q IDs as accepted decision history; record these as new interaction findings, not as a claim that the user never answered Q01–Q12. The recommendations in this report remain proposals until adopted.

The highest-value artifacts are:

1. One **ability table** containing actor/target eligibility, allowed phase, support, inputs, costs, quota identity, destinations, disclosure and cancellation outcomes. Include whether an already-confined player may declare Coup: “cannot counter Coup” and “cannot act while confined” should have an explicit precedence entry.
2. One **transition table** for action/response decision stages and for playing → board closure → settlement → finalization → next game. Include authoritative ordering, expected-version policy and financial state at each boundary.
3. One **visibility table** shared by server projections, client contracts and bots. Preserve known public history without leaking new hidden identities.
4. A **cross-layer scenario pack** containing every acceptance case above: expected card zones, counters, balances, event order and per-seat observations. Run the same traces through the simulator, server command path and reconnecting frontend, rather than writing separate interpretations in each layer.

The existing shared Go engine, match transaction boundary, idempotent commands, private projections and version pinning are appropriate foundations. Implement their proposed contracts; these findings call for more precise domain semantics, not microservices or a different stack. Safe work can proceed now on physical-card identity, basic deal/opening, deterministic transitions, exact ledger primitives and scenario tooling while the remaining rulings are resolved.
