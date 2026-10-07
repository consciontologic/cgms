# CGMS final draft — second frontend/backend assessment

Date: 2026-09-28. Method: Software Architect architecture review, with independent frontend, card-engine and financial-contract reviews. Scope: the amended rules, proposed API/state architecture, frontend guidance, scenario register and roadmap. This is a documentation assessment; no running Flutter client, Go engine, database or deployed service was verified.

The reviewed [FINAL_DRAFT](../design/drafts/FINAL_DRAFT.md) has 548 lines and SHA-256 `2b890a4b241f1e31d8b106167d155f96448b468b6b4c233c6d56bef6bb9ba5fa`. Line references below identify that snapshot; section links remain useful after later edits. The [first review](2026-09-28-final-draft-implementation-review.md) remains unchanged historical evidence. Its recommendations were accepted and incorporated; the new recommendations here are **proposals**, not silently adopted rules.

## Assessment

The documents are substantially more implementable. An engineer can now distinguish an activation from its later choices, committed costs from resolution costs, physical custody from lasting powers, and board closure from financial completion. The client and server have a shared visibility policy and a much clearer recovery contract. The proposed single Go authority with a durable match boundary remains an appropriate starting design; these findings do not justify a new stack or more services.

I would start the bounded simulator and infrastructure work. I would **not yet freeze full-game expected results or claim that every rule interaction has one answer**. Six P1 specification/contract gaps and one P2 edge case remain below. P1 means resolve before implementing or certifying that affected flow, not that all development must stop. None is evidence of a bug in an existing runtime.

Frontend readiness is strongest for the board structure, public/private projections, pending-decision UI and provisional results/settlement flow. Backend readiness is strongest for physical identity, staged effects, idempotency, game-scoped cash and durable recovery. Both still depend on a few missing adjudication rules. All **45 roadmap implementation tasks remain unchecked**; a larger specification is not implementation evidence.

## What the accepted recommendations changed

| First-review item | Adopted location and practical result |
|---|---|
| F01 — required choices | [§2.9](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order), API and state design persist actor/stage/decision IDs. Negotiator commits final movements together; Justice preserves sampled knowledge without rerolling; Dexter is an internal ace decision. |
| F02 — lifecycle | [§2.12](../design/drafts/FINAL_DRAFT.md#212-ability-lifecycle-and-costs) defines timing, support, allowance identity, commitment, cancellation and resolution costs. Named limits override the explicit player turn-cycle default. |
| F03 — victory admission | [API victory admission](../code/API-game.md#victory-admission) checks current eligibility rather than rejecting solely for unrelated versions. An already-confined player may declare a legal Coup. N05 below concerns a different game, not an unrelated version in the same game. |
| F04–F05 — money and settlement | [§4.4](../design/drafts/FINAL_DRAFT.md#44-financial-lifecycle) clears old Coup cash before replacement, separates phases, scopes voluntary transfers to their game and starts the next game with zero cash. Typed promises and explicit finish/refusal decisions are documented. |
| F06/F10 — cards and information | [§2.13](../design/drafts/FINAL_DRAFT.md#213-acquisition-and-visibility) specifies receiving zones, public captured Diamonds, retained public knowledge, public hand/Ace counts and private faces/breakdowns across clients and bots. |
| F07–F08 — persistent state | [§2.11](../design/drafts/FINAL_DRAFT.md#211-protection-and-doppelganger) and [§3.2](../design/drafts/FINAL_DRAFT.md#32-combos) separate bindings, privileges and Code ownership from physical card custody. Remote binding release does not expose a hidden partner. |
| F09 — Inflation | [§3.4](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces) rejects stacking and defines printed/effective values and transaction-before-removal order. N02 concerns its interaction with Fate, not reopening stacking. |
| Operating recommendations | [State/delivery](../design/DESIGN-go-state-and-delivery.md#proposed-design) adds resumable private exact-settlement progress and atomic final publication. The [roadmap](../planning/ROADMAP.md#accepted-review-traceability) maps every F item to work and evidence; match count defaults to three, tied totals share rank, and hibernation preserves untimed play. |

The acquisition, ability and financial tables are normative. The scenario register describes required traces and expected obligations; executable cross-layer fixtures still need to be built. Integration review also corrected substituted Kidnapper's physical return cost, hidden-mate assassination disclosure, ambiguous Inflation response targeting and broken section anchors. Those corrections are included in the reviewed snapshot.

## Remaining findings and recommendations

| ID | Priority | Gap | Affected roadmap work |
|---|---|---|---|
| N01 | P1 | Combat modifiers lack a complete comparison/payment formula | S06, D01–D03, P04 |
| N02 | P1 | Fate and persistent exposed Aces lack a shared custody/lifetime rule | S02/S04, D02–D03, O05, P04 |
| N03 | P1 | A confined participant can reach an undefined response/pass step | S06, D01/D08, O04–O05, P04 |
| N04 | P1 | Code's global effects are not classified against attack immunity | D03–D04, O03, P04 |
| N05 | P1 | Current-state command admission lacks an explicit origin-game fence | D06–D09, O03–O05 |
| N06 | P1 | Cross-game debt forgiveness lacks immutable score attribution | S07, D04–D05/D09, P04 |
| N07 | P2 | Barricade's delayed cards have an undefined scope of “use” | S02, D02–D03, P04 |

### N01 — Write the combat algebra, including virtual values and physical costs

**Evidence:** [§2.5](../design/drafts/FINAL_DRAFT.md#25-combat-basics--clubs--the-rule-of-even), lines 181–183; [§2.8](../design/drafts/FINAL_DRAFT.md#28-one-combat-procedure), lines 213–218; [§2.12](../design/drafts/FINAL_DRAFT.md#212-ability-lifecycle-and-costs), lines 317–320; [§3.4](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces), lines 420–425. Targets are frozen before responses, but defensive numerical/Advancement values are not explicitly attached to the selected subset, the whole series or a separate defense pool. Barricade requires an exchange without defining virtual-value payment.

**Counterexample:** a physical Club 5 plus Advancement 10 targets Hearts 7+8 protected by Barricade. The attack has 15 effective points but only 5 physical Club points to return. A numerical defensive ace raises a further question: exactly which total changes before revalidation?

**Recommendation:** define `A = effective committed Club sum + attack modifier` and `D = effective frozen target-card sum + applicable defense modifier`. Keep Advancement's defense Hearts-only; allow a defensive numerical ace to modify the targeted subset in any own number series under a club attack. Ordinary success requires `A = D`; Baron compares against the complete eligible series with its strict greater-than rule. An exchange returns the committed physical Clubs; virtual values never create cards or additional currency. A legal defensive increase can make an initially exact, frozen match fail. Keep Negotiator's already-committed modifiers fixed as accepted.

**Frontend/backend consequence and tests:** both layers need the same explainable calculation, not just a total. Test the 5+10 example, defensive numerical/Advancement changes, Dexter prevention, Baron, Inflation and Negotiator. Assert physical return IDs and usage separately from effective values.

### N02 — Define where persistent Aces live and what Fate can reset

**Evidence:** [Fate](../design/drafts/FINAL_DRAFT.md#32-combos), line 391, returns every controlled physical card and stops card-dependent abilities. [Aces](../design/drafts/FINAL_DRAFT.md#34-special-cards--aces), lines 416 and 426–427, retain Compensation until game end and Inflation until a qualifying transaction. Inflation targets an opponent, but its physical card controller and Fate treatment are unstated.

**Counterexample:** Bob is inflated when Alice applies Fate to him. Does the Ace stay, return and end the modifier, or return while a cardless modifier remains? The alternatives change both Bob's redraw count and subsequent Spade values. Compensation's stored remainder does not by itself answer whether new losses still accrue after Fate.

**Recommendation:** give resolved persistent Aces a public active-effect zone with separate source, target, physical custodian and expiry. Preserve Inflation and Compensation through Fate and exclude those active-effect cards from Fate's reset count; this is the most direct way to preserve their named lifetimes. Specify departure and game closure cleanup separately. If Fate should remove them instead, state that explicit exception and what happens to Compensation's spent game allowance and remainder. Either option is a new ruling and must be applied across all reset/acquisition rows.

**Frontend/backend consequence and tests:** effect badges, card conservation and exact matching must agree. Test Fate on caster and target, active Compensation with a remainder, departure, game end and eventual Ace recycling. Assert the draw count and physical card location, not just the effect flag.

### N03 — Skip participants who cannot respond, or explicitly permit their pass

**Evidence:** [§2.9](../design/drafts/FINAL_DRAFT.md#29-instant-effects--response-order), line 228, gives each other player a response/pass opportunity; lines 233 and 245 prohibit confined actions except Coup and preserve pending decisions without timeouts.

**Counterexample:** Alice's out-of-turn Ponzi is canceled by Confinement during Bob's turn. Bob starts another legal action before Alice's next scheduled turn. The response sequence reaches Alice, who is forbidden to act. Waiting for a pass could create a decision she cannot legally answer.

**Recommendation:** automatically omit confined and departed seats from the ordinary response sequence, while retaining the separate immediate Coup permission for an eligible confined player. This is a deterministic eligibility transition, not a timeout or inferred consent. Do not auto-pass an otherwise eligible disconnected player.

**Frontend/backend consequence and tests:** the UI must never display an impossible required input. Test a new action after out-of-turn confinement, a departed seat, reconnection during skipped positions and confinement expiry; preserve the remaining fixed seat order and Coup admission.

### N04 — Classify Code's economic declaration and recurring penalty

**Evidence:** [§1.7](../design/drafts/FINAL_DRAFT.md#17-attacking-trading--interacting-on-your-turn), line 96, protects initial players from every attack source. [Code](../design/drafts/FINAL_DRAFT.md#32-combos), line 397, is offensive and changes all opponents' economics, with a recurring −5 penalty. Neither that row nor the ability schedule identifies how a global declaration treats protected or confined opponents.

**Counterexample:** Alice can declare Code; Bob is attackable, while Carol has opened only Diamonds. Does Carol block the declaration, receive the price changes but avoid the penalty, or receive every effect? Her purchase preview and next round's ledger depend on the answer.

**Recommendation:** treat declaration as a non-attacking global economic action governed by Code's formation/Clubs prerequisites and the usual Underground waiver. Apply its price/bonus settings to all opponents. Treat its recurring −5 as hostile: at round end require the declarer's two current series and exclude initially protected or confined opponents. This is a proposed classification, not an inference that “offensive” already settles it.

**Frontend/backend consequence and tests:** publish per-seat effective economics and an explanation for penalty eligibility. Test mixed protected/unprotected seats, loss of initial protection, confinement across round end, Underground's waiver and loss/loan/replacement of Code.

### N05 — Bind every game command to a unique game instance

**Evidence:** the [API command envelope](../code/API-game.md#endpoints--methods), lines 38–53, is illustrative and carries no game identity; the command route is match-scoped. [Victory admission](../code/API-game.md#victory-admission), lines 132–143, deliberately accepts currently legal victories despite unrelated version changes. The [state design](../design/DESIGN-go-state-and-delivery.md#proposed-design) serializes the whole match, across games.

**Counterexample:** a first submission of a delayed Game 1 Coup command arrives after Game 2 begins. If the same player currently meets Game 2's conditions, current-state validation and a new command ID can accept it. Deduplicating already committed commands does not catch this first arrival. A nullified replacement also needs a fresh identity even when it occupies the same match slot.

**Recommendation:** require an immutable `game_id` for every game-scoped intent, including victory and settlement; bind turn/window/effect/decision IDs where relevant. Deduplicate first so a completed old command still returns its original result, then reject a new command for any different game instance. The victory exception waives unrelated aggregate-version changes **within the named game only**. Never automatically retarget a queued client intent to the next game.

**Frontend/backend consequence and tests:** clear or reconcile old pending UI operations at game transitions. Test a delayed first submission, duplicate committed submission, ordinary stale selections, a nullified replacement at the same slot, and a legal same-game Coup after an unrelated pass.

### N06 — Give cross-game forgiveness its own durable adjustment provenance

**Evidence:** [§4.2](../design/drafts/FINAL_DRAFT.md#42-debt-and-automatic-repayment), line 499, requires a nonspendable correction to an original negative obligation. [§4.4](../design/drafts/FINAL_DRAFT.md#44-financial-lifecycle), lines 523 and 527–528, replaces current-game scoring on Coup while freezing completed-game results.

**Counterexample:** Game 1 finalizes with Alice's −10 and debt 10. In Game 2 her creditor forgives 10, then Alice wins Coup. Rewriting Game 1 contradicts its frozen result; recording +10 as ordinary Game 2 score lets Coup erase the correction. Since the debt is already zero, unpaid-debt carry rules cannot restore it. Those paths yield different cumulative totals, 50 versus 40.

**Recommendation:** record an append-only, nonspendable match adjustment linked to the original debt and charged entry, with the forgiveness-creating game as its audit origin. Exclude this correction from subsequent Coup replacement; it is neither new cash nor a repayable receipt. Nullifying the game that created forgiveness restores both its correction and debt from the start snapshot. Preserve historical game results and show the match correction explicitly.

**Frontend/backend consequence and tests:** the match total needs an explainable adjustment line rather than a silently edited historical score. Test full/partial forgiveness, prior partial repayment, later Coup, nullification and duplicate commands. Assert score provenance, available cash and debt independently.

### N07 — Define Barricade's next-round availability per physical card

**Evidence:** [Barricade](../design/drafts/FINAL_DRAFT.md#33-non-combo-cards), line 408, says its five draws cannot be “used” before the next round; the [acquisition table](../design/drafts/FINAL_DRAFT.md#213-acquisition-and-visibility), lines 331 and 336, retains the restriction but gives trades ordinary use timing. [Victory](../design/drafts/FINAL_DRAFT.md#12-winning-conditions), line 52, counts all controlled hidden cards.

**Counterexample:** a Barricade draw supplies a missing Coup queen, or a card is transferred to another player before the next round. It is unclear whether the restriction blocks victory declaration, voluntary transfer, passive threshold counting, or only playing the card.

**Recommendation:** attach `available_from_round` to the physical card. Allow it to count as a possessed card toward an ordinary threshold, but prohibit voluntary trade/payment, voluntary opening/attachment and active use, including Coup, until that round. Mandatory exposure of captured or Fate-drawn Diamonds still occurs and does not clear the restriction. Preserve the marker through involuntary theft/capture, Fate, shuffle and redraw until the round boundary; it overrides Kidnapper's immediate-use permission. Forcibly exposed cards remain unavailable for active use. Keep concealed card metadata private when those cards change zones.

**Frontend/backend consequence and tests:** show a precise availability badge and reason for disabled actions. Test threshold count versus Coup activation, voluntary transfer rejection, forced theft, next-round unlock and the chosen Fate/recycle behavior. Do not guess from a disabled-button implementation.

## Engineering work and evidence still required

- **Shared executable scenarios:** convert the accepted register into versioned fixtures carrying input state, actors, IDs, exact ledgers, physical zones, quotas, event order and per-seat projections. The second-review rulings need their own expected outputs after adoption. Run identical traces through engine, simulator, server recovery and client journeys; do not maintain four independent rule interpretations.
- **Exact settlement performance:** the new private workspace/compressed audit design is a credible plan, not an implemented optimization. Prove small-ledger equivalence and crash/restart progress at large denominators. A finite calculation can still exceed request budgets. Publish measured cost without changing the exact rules.
- **Untimed product operation:** hibernation correctly preserves rules but cannot make an absent player finish settlement. Account for that in the lobby/results UX and decide allowance/reward handling before public or monetized release. This is a known accepted tradeoff, not a recommendation to introduce an automatic forfeit.
- **Runtime and release evidence:** no game balance, full-game duration, accessibility, race safety, load capacity, store readiness or production recovery has been demonstrated by this review. The historical Flutter analysis failure remains recorded and unresolved; documentation checks do not repair it.

## Recommended sequence

Resolve N01–N06 before freezing the affected full-rule/API contracts; settle N07 before implementing Barricade's draw restriction. Keep S01–S05 and isolated ledger/fixture foundations moving where they do not depend on these answers. Then implement one complete vertical scenario through engine, persistence, reconnect and frontend before broadening card coverage. Keep the existing roadmap as the sole execution checklist and require its evidence before checking any task complete.
