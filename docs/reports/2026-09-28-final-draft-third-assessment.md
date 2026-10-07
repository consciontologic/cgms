# CGMS final draft — third frontend/backend assessment

Date: 2026-09-28. Method: Software Architect architecture review, including independent frontend and card-engine reviews and a financial-contract integration pass. Scope: the amended rules, proposed API/state architecture, scenario register, frontend guidance and roadmap. This is a documentation assessment; no running Flutter client, Go engine, database or deployed service was verified.

The reviewed [FINAL_DRAFT](../design/drafts/FINAL_DRAFT.md) has 578 lines and SHA-256 `472167b8fe28fa3e5883c4f9ca0fed66efcefbc91e977da2486b1660d58c3219`. Line references identify that snapshot; section links remain useful after later edits. The [first review](2026-09-28-final-draft-implementation-review.md) and [second assessment](2026-09-28-final-draft-second-assessment.md) retain their original bodies as historical evidence. F01–F10 and N01–N07 are now accepted and incorporated. The new T01–T03 recommendations below are **proposals**, not silently adopted rules.

## Assessment

The specification is ready for bounded implementation, with fewer unresolved interactions than the previous review. I found no remaining contradiction in the inspected propagation of N01–N07. Combat values and physical costs now agree; persistent Aces have an explicit home and cleanup policy; response traversal cannot demand an illegal pass; Code has per-player economic and penalty eligibility; commands cannot migrate into another game; forgiveness retains its score provenance; and restricted cards retain their restrictions through forced movement.

I would start the simulator and shared fixtures now. I would resolve **two P1 contract gaps** before implementing the affected offer and automatic-draw flows, and make **one P2 purchase clarification** before freezing its API. These findings do not require a different architecture or a halt to independent work. They are specification findings, not demonstrated runtime bugs. All **45 roadmap implementation tasks remain unchecked**.

From a frontend developer's perspective, the documents now support explaining legal choices, waiting actors, physical versus effective values, private information, game transitions and provisional versus final results. The main remaining interaction problem is the lifetime of an unanswered trade or loan proposal. From a backend developer's perspective, the proposed state boundary can represent the accepted rules, but automatic draws still need one authoritative order and purchases need a precise quantity contract. A deterministic engine must encode these choices; it cannot derive them from a database lock or a client animation.

## Accepted changes and integration checks

| Accepted item | Applied contract and review result |
|---|---|
| N01 — combat algebra | [§2.5](../design/drafts/FINAL_DRAFT.md#25-combat-basics--clubs--the-rule-of-even) separates effective A/D totals, frozen targets and physical exchange costs. Numerical defense, Hearts-only Advancement, Baron, Suicide Bomb and Negotiator use explicit exceptions. Required traces include Club 5 plus Advancement 10 against Barricaded Hearts 7+8. |
| N02 — persistent Aces | [§3.4](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces) places Inflation and Compensation in a public active-effect zone with target custody and separate source/expiry. Fate excludes them from its count. Target departure and board closure discard them; source departure alone does not end another player's Inflation. |
| N03 — response eligibility | [§2.9](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order) skips confined/departed seats while preserving eligible disconnected waits and the confined Coup exception. API/state guidance reflects that difference. |
| N04 — Code classification | [Code](../design/drafts/FINAL_DRAFT.md#32-combos) applies global economic settings separately from its hostile recurring penalty. The latter requires the active, non-confined declarer's functioning formation and two current series, and excludes protected/confined/departed targets. |
| N05 — game identity | The [API envelope and admission policy](../code/API-game.md#endpoints--methods) require immutable game identity, preserve committed duplicate results and reject new submissions for another game. Replacement games receive fresh IDs. Client guidance forbids retargeting queued intentions. |
| N06 — forgiveness attribution | [§4.2](../design/drafts/FINAL_DRAFT.md#42-debt-and-automatic-repayment) separates cross-game match corrections from same-game corrections, records charge lineage and restores debt/corrections on nullification. A later Coup preserves a correction to a finalized game's charge without creating cash. |
| N07 — restricted cards | [§2.14](../design/drafts/FINAL_DRAFT.md#214-restricted-cards) covers physical costs as well as abilities, preserves markers through Fate/theft/shuffle, and distinguishes passive threshold counting from active Coup. Negotiator's eligible payment pool excludes restricted Spades. |

These checks compare written contracts, not executable outcomes. Integration also clarified Fate's treatment of exposed unassigned cards, end cleanup of active Aces, and restricted payment eligibility. Those details are necessary to make the accepted recommendations agree across sections.

## Remaining findings and recommendations

| ID | Priority | Remaining issue | Affected roadmap work |
|---|---|---|---|
| T01 | P1 | Ordinary trade and loan proposals have no complete lifecycle | D03, D05–D08, O03, O05, O08, P04 |
| T02 | P1 | Simultaneously earned Compensation draws have no recipient order | S04, D02–D03, D08, O05, P04 |
| T03 | P2 | Deck purchases do not explicitly distinguish chosen quantity from maximum affordable quantity | D02–D03, O03, P04 |

P1 means resolve before certifying that affected flow; P2 is a narrower contract clarification. The recommendations below belong in the rules and shared fixtures if accepted, rather than being independently invented in frontend and backend implementations.

### T01 — Separate proposals from accepted actions

**Evidence:** [§1.7](../design/drafts/FINAL_DRAFT.md#17-attacking-trading--interacting-on-your-turn), line 105, limits who initiates ordinary trades. [§2.3](../design/drafts/FINAL_DRAFT.md#23-lending-combos-to-allies), line 168, allows loan offer/acceptance on either party's turn. [§2.9](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order), line 238, requires completing an action before starting another. The [API](../code/API-game.md#endpoints--methods), lines 70–80 and 90–95, defines offer/accept commands and atomic acceptance revalidation, but no decline, withdrawal, expiry or blocking policy. The ban on withdrawing an accepted activation does not define an unanswered proposal.

**Counterexample:** Alice offers Bob a hidden card. Bob stays silent. Can Alice continue playing or end her turn? If she can, may Bob accept after her turn, or after the offered card moves? The client cannot determine whether to show a blocking decision, a dismissible offer or an expired notification.

**Recommendation:** make proposals nonblocking records with immutable offer ID, revision, exact terms and origin game/turn. Reserve no cards or allowances before acceptance. Allow the maker to withdraw and the recipient to decline; expire ordinary trade and loan proposals at the end of the turn in which they were created, or earlier on board closure or party departure. A loan can be proposed afresh on either participant's turn under §2.3. Changes to terms require a new revision and fresh acceptance. Pending proposals must not create an alliance or block End turn.

Acceptance should revalidate current ownership, availability, transfer prerequisites/quotas, turn eligibility and alliance status at a legal idle boundary. Merely receiving an intact loan does not require the support needed to use its ability; check that support at ability use under §2.3. Once accepted, the transfer starts the ordinary action/response procedure, becomes nonwithdrawable and commits transfers only at successful resolution. Define the accepting player as that action's actor, including for Confinement, and preserve the original offer's turn eligibility. This actor/response choice is a proposed ruling, not something the current API already settles. Serialize acceptance against withdrawal/expiry so exactly one wins; reject a stale revision without revealing hidden state. Offer terms containing hidden cards are visible only to the authorized parties.

**Frontend/backend consequence and tests:** distinguish pending proposal, accepted action, completed, declined, withdrawn, expired and invalid states. Exercise unanswered offer plus End turn, acceptance versus withdrawal, cards moved or restricted after proposal, changed terms, duplicate acceptance, stale turn/game IDs, Confinement after acceptance and reconnect with the same authorized offer state. Assert that the alliance appears only after a successful loan transfer and that no client display reserves cards the authority has not reserved.

### T02 — Order deferred Compensation draws across recipients

**Evidence:** [§2.9](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order), line 241, defers Compensation draws until action completion. Justice transfers its selected cards atomically at line 251. [Compensation](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces), line 452, creates a draw for each five qualifying losses but does not order multiple recipients. [§4.2](../design/drafts/FINAL_DRAFT.md#42-debt-and-automatic-repayment), line 518, specifies a fixed seat order for monetary receipts, not card draws.

**Counterexample:** Bob and Carol each have active Compensation with remainder four. Alice's Justice captures one exposed number card from each, of different ranks, and returns its queen. The rest of the draw/discard supply is exhausted. Both players earn a draw, but only one can receive that queen. The current rules do not determine which one; even ample supply leaves the identity of their respective cards order-dependent.

**Recommendation:** build one deferred draw queue in ascending fixed seat order, independent of initiative or target-selection order. After the action's returns and required shuffle, complete each recipient's owed draws before advancing to the next. Follow ordinary recycling/exhaustion rules; an unavailable mandatory draw creates no later entitlement. Finish this automatic draw phase before point settlement and ordinary victory checks, as §2.9 already requires. Store the queue position and committed random outcomes if execution spans durable transitions; do not expose an intermediate draw state as an ordinary-action boundary.

**Frontend/backend consequence and tests:** animate the authoritative order and disclose each card only to its permitted viewer. Test two recipients with one available card, adequate supply with distinct identities, more than one owed draw, reversed initiative/Justice declarer positions, exhaustion/recycling, and crash/replay without duplicate draws. Assert Compensation counters, physical conservation, private projections and the completed boundary before ordinary victory. Preserve the existing Coup interruption rules if the queue uses separate atomic transitions.

### T03 — Specify the purchase quantity and payment contract

**Evidence:** [§1.5](../design/drafts/FINAL_DRAFT.md#15-deck--reshuffling), line 81, checks a purchase's “requested quantity.” The [Spades reference](../design/drafts/FINAL_DRAFT.md#35-numbers-suits), line 459, permits one card for 5 or more points and says 10 can draw two. It does not state whether the buyer chooses a quantity below the affordable maximum or must receive the maximum. The existing examples suggest buyer choice, so this is a clarification rather than a demonstrated contradiction.

**Counterexample:** the draw and discard piles are empty and the buyer pays one unrestricted Spade 10 at price 5. Returning that payment supplies one card. A mandatory maximum quantity of two makes the purchase reject; a chosen quantity of one makes it legal. Implementing either interpretation changes available actions near deck exhaustion.

**Recommendation:** explicitly let the buyer choose integer `q ≥ 1` at an own-turn idle boundary, with unrestricted number-Spade payment totaling at least `q × current price`. Excess payment gives no change. Permit payment from hand or the buyer's exposed Spades without Underground: the ordinary exposed-card trade ban concerns player-to-player trades. This zone permission is useful clarification of the currency rule, not a separately demonstrated contradiction. Use current Code/Inflation values, validate that supply including returned payment can satisfy all `q` draws, and commit payment/return/shuffle/draw only on legal resolution after applicable responses. Preserve existing rejection-without-payment and transaction-before-Inflation-removal rules.

**Frontend/backend consequence and tests:** expose quantity as an explicit command field and show quantity, effective cost, physical payment and excess before acceptance. Cover 5/9 for one card, 10 for a chosen one or two, the exhausted-supply example, Code prices, fractional Inflation values, the 20-printed-point removal threshold, exposed payment without Underground and restricted payment rejection. Revalidate supply and payment after responses so no partial purchase can commit.

## Engineering risks that remain after rule clarification

- **Shared executable evidence is the next priority.** The register now contains substantial normative detail, but prose cannot prove those rules compose. Implement versioned traces with exact physical zones, quotas, event order, ledgers and per-seat observations, then run the same traces through the engine, recovery path and client. Include a restricted Barricade draw blocking the first opening of its suit: all held numbers must be opened together, while the restricted card cannot be opened. That is a consequence of existing rules, not another unresolved decision.
- **Untimed waiting remains a deliberate product tradeoff.** Eligible disconnected response/settlement participants can still keep a game pending. N03 removes impossible inputs, not all waits. Hibernation/reconnect UI and the roadmap's abandoned-game economy policy need implementation and release evidence; a timeout must not invent a pass or refusal.
- **Exact settlement remains a performance and recovery obligation.** Cross-game corrections improve attribution but do not prove FIFO batching, large-denominator cost or restart progress. Keep the unbatched reference and exact equivalence/fault gates already in the roadmap.
- **No runtime or release readiness claim follows from this review.** Game balance, duration, accessibility, security enforcement, database atomicity, race behavior, load, backups and store integration remain unverified. The historical Flutter analysis failure is still recorded as unresolved; documentation checks do not repair it.

## Recommended next step

Resolve T01/T02 before freezing their state machines, and settle T03 alongside the purchase schema. Continue independent S01–S05 foundations now, then implement one complete accepted scenario through transition, persistence, reconnect and frontend before broad card coverage. Keep the existing roadmap as the sole execution checklist. Another general prose expansion is less valuable than executable evidence once these bounded contracts are settled.
