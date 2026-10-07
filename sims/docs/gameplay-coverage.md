# Accepted gameplay coverage and evaluation evidence

Scope authorized 2026-09-30: D01–D04 and simulator/domain D08–D09, complete fair
bots and evaluation. The [roadmap](../../docs/planning/ROADMAP.md) owns completion.
This evidence record does not complete mixed persistence/server gates.

Historical campaign counts in earlier sections remain snapshots at their stated
cutoff. The [latest v3 evidence](#v3-verification-and-development-evidence) below
records later replay completion and new development failures separately.
Legacy v1 effect estimates also retain the later-discovered reconfiguration
tie-key confound; corrected v2 policies require new experimental evidence.

## Historical baseline and initial independent audit

[Phase 0 evidence](phase-0-implementation-evidence.md) records 2,000 replayed
partial prefixes and **zero completed games**. Preserve that category separately
from all later completed-game evidence. Healthy CodeGraph inspected actual Go
source before extension. The independent audit found implementation gaps, no
unresolved Q01–Q12 decision. Normal deal stopped before turn orchestration;
legal menus covered one/two-card ordinary combat; Ponzi resolution was unsupported.

## Rule-to-implementation and targeted-test matrix

Independent source review, 2026-09-30. Paths are under
`backend/internal/game/`; test names below have the `Test` prefix omitted only
for readability. These are concrete fixtures present in source, **not a claim
exhaustive coverage of all possible rule interactions**.
Further interactions below are additional coverage opportunities unless explicitly
identified as an unresolved gate; the final gate crosswalks supersede initial
audit notes. No accepted decision is reopened by a missing fixture.

Full-game autonomous exercise attribution remains **not established per rule**
for every row below. It requires recorded event/transition evidence, not inference
from a bot menu or test name. The matrix separates that missing evidence from
existing targeted fixtures.

| Rule / canonical clause | Implementation | Targeted fixture names | Coverage boundary / further evidence |
|---|---|---|---|
| Q01 / §§2.5–2.10 | combat.go | CombatRejectRetainsState; BombOrdinaryBaronCancellationCosts | Closed by AcceptedAbilityCancellationKeepsCostsAndQuota and JusticeExileInfiltratorConfinementCosts |
| Q02 / §2.9 | decision.go, lifecycle_board.go | DexterConfinementContinuation; LifecycleConfinementCannotSkipRoundThirteenSettlement | Rare nested Confinement/Coup combinations |
| Q03 / §§1.2,4.4 | lifecycle.go | LifecycleDeclarationPublicProof; LifecycleRejectsIllegalDeclarationWithoutPublishingOrMutation | Concurrent declaration ordering fixture |
| Q04 / §§2.1,2.11 | formation.go | NaturalFormationShapes; FormationOpeningBindingAndIsolation | Exhaustive Great People component interactions |
| Q05 / §§2.3–2.4 | proposal.go, state.go | IntactLoanPreservesBindingAllocation; LoanReturnCancellationPreservesCustodyAndAlliance | Closed by FourSeatDisjointAlliancesAndThirdPartnerRejection |
| Q06 / §§1.6–1.7,3.2 | opening_action.go, ability.go | OpeningCommandWaitsAndConsumesAllowance; AbilityFateReturnsCostAndPreservesPersistentAce | Reopening plus loan/return quota combinations |
| Q07 / §§1.1,1.4–1.5 | setup.go, draws.go | DealSeededRedealVectors; InitiativeNonNumbersRepeatedTieIsolationAndFrozenOrder; SupplyRecyclingAndExhaustion | Full lifecycle random-outcome replay attribution |
| Q08 / §§2.10,3.2–3.4 | decision.go, ability.go | NegotiatorStagesAndRetry; JusticeRandomOutcomeRestartRetry; DexterOffensiveModifierDecision | All staged effects across serialized restarts |
| Q09 / §§1.9,4.4 | lifecycle_finance.go, ledger.go | LifecycleCoupAndDepartureAndNullification; LifecycleDepartureRetainsDebtAndUnanimousVoidRestoresStart | Two/one/zero survivors plus later passive receipts |
| Q10 / §4.2 | settlement.go | SettlementBatchOracleAndResume; SettlementGoldenQ10; SettlementLargeDenominatorUsefulProgress | LifecycleTransferIntegratedReferenceAndSerializedResume compares integrated transfer against unbatched reference; broader generated graph coverage remains |
| Q11 / §2.9 | transition.go, lifecycle_operation.go | ResponsesSkipConfinedButWaitDisconnected; LifecycleOperationRetryDoesNotDuplicateDraw | Closed by MatchCompletedGameDuplicateResolvesBeforeCurrentGameAdmission |
| Q12 / §§4.3–4.4 | promises.go, lifecycle_operation.go | PromiseReputationBoundaries; PromiseFinalOffersAndPendingFinish; LifecycleFinancialOperationsRequirePaymentBeforeFinish | Autonomous negotiation behavior attribution |
| F01 | decision.go | JusticeRandomOutcomeRestartRetry; CoupInterruptsDecisionAndKeepsCosts; DexterDeclinePreservesDefensiveModifier | Coup/reconnect at every accepted-effect stage |
| F02 | ability.go, combat.go | AcceptedAbilityCancellationKeepsCostsAndQuota; AbilitySupportLossKeepsQuotaAndCost; PonziQuotaUsesOwnTurnCycle; FullBombAttackPhysicalCostAndShield | Completed §2.12 crosswalk and main/response timing matrices below |
| F03 | transition.go, lifecycle.go | CoupInterruptsDecisionAndKeepsCosts; LifecycleRejectsIllegalDeclarationWithoutPublishingOrMutation | Same-game stale unrelated version and dual declaration fixtures |
| F04 | ledger.go, lifecycle_finance.go | FinancialCoupNextGameAndDeparture; LifecycleCoupPreservesEarnedReceiptsBeforeAbandoningDraws | Departed creditor plus partial repayment plus Coup interaction |
| F05 | promises.go, lifecycle_operation.go | PromiseFinalOffersAndPendingFinish; LifecycleExplicitFinancialFinalization | Complete replay of financial wait/payment continuation |
| F06 | state.go, loan_return.go | CapturedDiamondForcedExposureKeepsRestriction; LoanPartialReturnRetainsRemainingCustodyProvenance; LoanIntactReturnRestoresFormationController | Special physical theft covered by SubstitutedKidnapperReturnsEveryPhysicalCostAndPreservesOwnBinding; acquisition invariants remain tested independently |
| F07 | formation.go, lifecycle.go | FormationEarnedUndergroundDoesNotConsumeSeriesOpening; AbilityCodeSettingsPersistAfterCardLoss | UndergroundLoanBreakReturnKeepsPrivilegeAndAllocatedIncome; CodeLoanBreakReturnRedeclarationPreservesOrReplacesDeclarer close the custody chain |
| F08 | state.go, formation.go | BindingReunionNeedsExplicitAllocation; AssassinationReleasesRemoteBinding; IntactLoanPreservesBindingAllocation | SubstitutedKidnapperReturnsEveryPhysicalCostAndPreservesOwnBinding verifies all physical costs, retained own binding, stolen-pair release, ten-point obligation |
| F09 | purchase.go, proposal.go | InflationNineteenTwentyAndNoCumulativeRemoval; VoluntaryTradeClearsInflationAfterTwentyPrinted; PurchaseInflationPriceAndPrivatePayment | Qualifying purchase followed by full-price purchase |
| F10 | observation.go, history.go | ProjectionHiddenEquivalence; GameplayObservationEquivalentHiddenRoyalSwap; HiddenReturnAndShuffleDoNotRevealIdentities | Every action menu category and report surface |
| N01 | combat.go | VirtualAttackPhysicalExchange; FullBaronAttackRequiresWholeSeriesAndScoresNoClubs; BarricadeVirtualExchange | Cross-product modifier fixtures |
| N02 | state.go, lifecycle.go | SourceDepartureKeepsTargetInflation; AbilityFateReturnsCostAndPreservesPersistentAce; CompensationDepartureExpiresRemainder | OrdinaryClosureDiscardsBothPersistentAceKinds verifies both live kinds at closure; Fate/recycling fixtures cover persistence |
| N03 | combat.go, lifecycle.go | ResponsesSkipConfinedButWaitDisconnected; ConfinementWaitsForScheduledTurnDraw | Reconnect after already skipped response seat |
| N04 | lifecycle.go, ability.go | LifecycleRoundIncomePrecedesCodeChargeAndResets; LifecycleCodePersistentEndSettings | CodePenaltyMixedProtectionConfinementAndUndergroundSupport covers mixed target exclusions and mandatory two-series penalty support |
| N05 | lifecycle_match.go, lifecycle_operation.go | MatchNullificationOperationKeepsRetryIdentity; MatchCompletedGameDuplicateResolvesBeforeCurrentGameAdmission | Historic retry, changed-body conflict and never-committed stale intent now verified |
| N06 | ledger.go | FinancialCrossGameCorrectionAndVoid; LifecycleCarriedForgivenessCoupAndVoidProvenance | Autonomous cross-game correction attribution |
| N07 | state.go, ability.go | BindingsAvailabilityAndFate; FateReturnRequiresShuffleAndPreservesMarkerOnRedraw; AbilityDexterRestrictedThresholdAndBindingRelease | RestrictedHoldingsCountForOrdinaryProofButBlockCoupUntilRoundBoundary adds explicit threshold/Coup distinction and round expiry |
| T01 | proposal.go | ProposalNonblockingRevisionAndAtomicTransfer; ProposalStaleRevisionAndExpiry; BrokenFormationCannotGrantLoanPrivilege | AcceptedTransferCancellationKeepsAceUseAndNoAlliance covers acceptor Confinement for Ace trade and combo loan; remaining proposal races separately listed |
| T02 | draws.go, lifecycle_board.go | DeferredMultipleRecycleCanonicalResume; DeferredLegalCoupAtEachBoundary; LifecycleCoupPreservesEarnedReceiptsBeforeAbandoningDraws | All ability loss events route through ordered queue |
| T03 | purchase.go | PurchaseResponseThenExactQuantity; PurchaseSupplyIncludesPaymentAndRejectsWithoutMutation; PurchaseCancellationRetainsPaymentAndQuantity | PurchasePostResponseSupplyLossNeverResizesChosenQuantity and RuleBoundaryGoldens/purchase close exact boundary evidence |

### Complete-game integration is a distinct evidence category

`backend/internal/sim/gameplay_test.go::TestGameplayNormalDealCompleteAndReplay`
contains normal-deal three- and four-player, two-game fixtures, using the explicitly
restricted `scripted-end-turn-test-only` policy. They exercise legitimate round-13
closure, financial finalization, next-game identity, and replay. They do not
exercise autonomous strategic competence, broad ability use, or balance.
`TestGameplayResumeRetainsPrefix` tests continuation; the worker-count fixture is
currently a **four-step prefix** comparison, not completed-match determinism.
Neither fixture is a held-out campaign or thousands of completed-game replays.

### Independent review findings and minimum closure tests

The N05 regression `TestMatchCompletedGameDuplicateResolvesBeforeCurrentGameAdmission`
failed in captured run `gameplay-n05-review-red--20260930T101055Z-2050022`:
`MatchLifecycle.ApplyOperation` checks the current game before consulting archived
command receipts. The committed original-game duplicate must instead resolve
without mutating the new game. A changed-body retry and never-committed old-game
intent must still reject. The coordinating agent repaired lookup ordering; the regression plus changed-body
and never-committed stale intents passed in `lifecycle-finance-review--20260930T101642Z-2085175`.

Independent review also identified missing replay decision-policy provenance
validation and a Justice hidden-selection planner path which can reject when no
eligible hidden card exists. Those are assigned to the runner/engine owners;
reviewer closure requires captured passing regressions, not an assumed repair.

Minimum additional conformance closure includes: the per-ability cost/cancellation
matrix; loan-derived lasting privileges and Code ownership chains; special
Kidnapper binding reassignment; mixed protection/Confinement Code penalties;
all required decision stages across restart/Coup; acquisition visibility for each
entry point; and complete-match worker determinism with actual autonomous policies.
These remain distinct from statistical sample-size and held-out-evaluation gates.

## Observed development evidence

Purchase tests first failed with unsupported transition in
`gameplay-purchase-red--20260930T094130Z-1886398`; green in
`gameplay-purchase-green--20260930T094240Z-1891941`. Independent review identified
missing successful-payment public history; the regression failed in
`gameplay-purchase-disclosure-red--20260930T094416Z-1901801` and passed in
`gameplay-purchase-disclosure-green--20260930T094424Z-1903583`.

Revisioned proposal tests failed before command support in
`gameplay-offer-red--20260930T094556Z-1910559`; focused tests passed in
`gameplay-offer-green3--20260930T094757Z-1919190` after concurrent compile gaps were
read and repaired. Broken formation loans were rejected after the independent
review regression `gameplay-loan-shape-red--20260930T095205Z-1949048` →
`gameplay-loan-shape-green--20260930T095218Z-1949819`. Completed qualifying trade
Inflation removal: red095324 → green095344. Captured logs are under
`/tmp/agent-runs/`; these are synthetic tests, not gameplay campaign outputs.

Final full-rule mapping, campaign counts, immutable artifact references and
findings must be appended only after actual execution. No balance or stronger-bot
claim follows from this in-progress record.

Independent integrated financial review added three focused fixtures:
`TestLifecycleTransferIntegratedReferenceAndSerializedResume` compares a ten-point
transfer with six-point recipient debt against the unbatched reference after JSON
restart at each budgeted step (cash 0/4/6), then checks duplicate and conflicting
retry behavior. `TestLifecycleDepartureRetainsDebtAndUnanimousVoidRestoresStart`
checks departure retaining a six-point unpaid obligation, positive-score removal,
departed-seat consent, exact nullification restoration, and no consumed match slot.
`TestLifecycleCarriedForgivenessCoupAndVoidProvenance` checks the accepted N06
−10 + 10 + 50 example through lifecycle commands, cash 50, immutable prior results,
and restoration on nullification. All lifecycle/match/forged-payment tests passed
with the race detector in `lifecycle-finance-review--20260930T101642Z-2085175`.
These additional existing-behavior conformance fixtures passed on first execution;
they are not presented as new red-before-green implementation evidence.

Cancellation matrix `TestAcceptedAbilityCancellationKeepsCostsAndQuota` exercises
accepted Ringleader, Richer sacrifice, Barricade sacrifice, Compensation, main
Inflation, Fate, Code, Kidnapper, and Dexter assassination followed by Confinement;
selected cards remain, accepted allowances remain consumed, and effects are not
applied. Opening allowance, Ace-supplier trade allowance, canceled loan alliance,
purchase payment, and loan-return custody are covered by separate fixtures.
`TestSubstitutedKidnapperReturnsEveryPhysicalCostAndPreservesOwnBinding` uses the
lifecycle interface to return all three allocated physical costs, retain their
substitute binding, expose both stolen kings/release their binding, and create the
exact ten-point obligation. Focused race run: `cancellation-matrix-final--20260930T102259Z-2120234`.

Security review found that one actor could assert other seats' nullification or
system-forgiveness consent. `TestLifecycleCannotAssertOtherPlayersFinancialConsent`
failed in `individual-consent-red--20260930T102351Z-2134460`; public commands now
record each actor's own assent, reject supplied foreign consent, and execute only
after all required seats assent. `TestSystemForgivenessRequiresRecordedIndividualConsents`
checks a departed participant, JSON continuation, duplicate idempotency, and exact
nonspendable correction. The low-level verified-consent financial primitives are
unchanged. Full game package passed with race detection in
`individual-consent-full--20260930T102517Z-2139608` (15.780 s).

`TestDepartureChoicesRemainPendingUntilEveryEligibleSeatAnswers` first failed in
`departure-choice-red--20260930T102739Z-2148547` and passed after adding recorded
per-seat `departure-choice` operations. An explicit `Accept=false` means stay;
absence of a response means unresolved. Only the final eligible seat's answer
applies the entire departing set atomically. Public collective `depart` intents
are rejected, and public `begin-round` requires the completed boundary.
`TestDepartureChoiceExplicitStaysRetryAndCoup` covers explicit stays, identical
retry, changed-body rejection, and legal Coup during a partially answered boundary
without applying a queued departure. JSON resume plus all-three simultaneous
departures are covered in the first fixture. Focused lifecycle/match race gate
passed in `departure-choice-final--20260930T102854Z-2157218` (2.023 s).

## Independent whole-gate assessment (2026-09-30, before campaign)

D01–D04 remain unchecked: the implemented behavior and passing focused tests do
not yet establish every clause of their gates. Specific remaining closure work:

- D01: explicit Justice cancellation retaining its chosen queen and quota;
  canceled Exile/Infiltrator tests asserting committed bypass usage/returned jack.
  The nine-ability Confinement matrix and existing Negotiator/Ponzi/combat tests
  establish substantial cancellation coverage, not every §2.12 timing case.
- D02: explicit People-to-People chain rejection and curated golden/exact lifecycle
  traces covering the listed rule families. Existing canonical serialization
  goldens do not alone supply the requested per-rule outcome traces.
- D03: Underground loan/break/return with allocated-income and lasting privilege;
  Code loan/break/return/redeclaration with unchanged versus replaced declarer;
  mixed protected/confined/eligible Code penalty targets. The complete §2.12
  legal/illegal activation matrix is broader than the current cancellation matrix.
- D04: four-player two-disjoint-pair and third-partner-rejection fixtures plus the
  mixed N04 penalty case above. Existing FIFO/reference/batching, correction,
  departure/nullification, and integrated exact-transfer evidence is substantive.

This assessment concerns domain evidence only. It neither completes D05–D07 nor
claims database transactions, server recovery, or concurrency gates. The canonical
roadmap remains the only completion checklist; these are evidence gaps rather
than reopened accepted-rule decisions.

F07/N04 closure fixtures added after the preceding assessment:
`TestUndergroundLoanBreakReturnKeepsPrivilegeAndAllocatedIncome` completes an
accepted loan, verifies both lasting privileges, exactly ten income from five
allocated kings, excludes hidden/unallocated kings, transfers income on intact
return, and gives no income after breakage/surviving partial return.
`TestCodeLoanBreakReturnRedeclarationPreservesOrReplacesDeclarer` checks old
settings/exemption through loan, penalty cessation/restoration on loan/return,
cessation on breakage, and replacement on the borrower's legal redeclaration.
`TestCodePenaltyMixedProtectionConfinementAndUndergroundSupport` checks a four-seat
round with protected/confined/eligible opponents: only the eligible fourth seat
owes five; earned Underground does not waive the penalty's two-series requirement.
These pre-existing behavior conformance fixtures passed on first run
`lifecycle-f07-matrix--20260930T103544Z-2189802`; the broader lifecycle/financial/
promise/reference-settlement race gate passed in
`lifecycle-f07-final--20260930T103559Z-2190978` (7.940 s).
The D04 assessment now awaits the independently owned exclusive-alliance fixtures;
its identified N04 penalty evidence gap is closed. This does not change D03's
remaining full §2.12 activation/timing matrix obligation.

## Completed §2.12 activation/timing crosswalk

The following supersedes the earlier D01/D03 timing-gap assessment. Test names
omit `Test`. `AbilityTimingSupportAndQuotaMatrix` independently admits each of
nine supported main abilities, then rejects other-turn, confined, closed-board,
pending-action, missing-number-support and consumed-quota variants, asserting no
state/event/cost mutation. `ResponseTimingSupportAndQuotaMatrix` does the same for
six response types with idle, wrong-responder, confined, closed-board,
consumed-quota and missing-support variants. Their fixtures first prove legal
admission, so an unrelated missing-card failure cannot stand in for timing proof.
`AcceptedAbilityCancellationKeepsCostsAndQuota` checks all nine main abilities'
Confinement cancellation. These matrices supplement named interaction tests below;
passive rows correctly have no activation/quota command to test.

| §2.12 schedule row | Legal/illegal support and timing fixtures | Costs, quota, cancellation / boundary fixtures |
|---|---|---|
| Accepted trade / combo loan | ProposalNonblockingRevisionAndAtomicTransfer; ProposalStaleRevisionAndExpiry; BrokenFormationCannotGrantLoanPrivilege | AcceptedTransferCancellationKeepsAceUseAndNoAlliance; FourSeatDisjointAlliancesAndThirdPartnerRejection |
| Deck purchase | PurchaseResponseThenExactQuantity; PurchaseSupplyIncludesPaymentAndRejectsWithoutMutation; PurchaseInflationPriceAndPrivatePayment | PurchaseCancellationRetainsPaymentAndQuantity; RuleBoundaryGoldens/purchase |
| Ordinary Clubs | CombatExactAndPhysicalUsage; CombatRejectRetainsState; CombatRevalidatesTwoSeriesAfterDexterPayment | CombatExactAndPhysicalUsage; BombOrdinaryBaronCancellationCosts |
| Doppelganger | FormationOpeningBindingAndIsolation; RegisteredOffensiveBindingUndergroundRequiresBothSupports; RegisteredBindingLoanSupportAndReunion | BindingReunionNeedsExplicitAllocation; AssassinationReleasesRemoteBinding; OpeningCancellationKeepsOpeningAllowance |
| Fate | AbilityTimingSupportAndQuotaMatrix/fate; AbilityFateReturnsCostAndPreservesPersistentAce | AcceptedAbilityCancellationKeepsCostsAndQuota/fate; RuleBoundaryGoldens/fate |
| Baron | FullBaronAttackRequiresWholeSeriesAndScoresNoClubs; BaronCombatAlgebra | BombOrdinaryBaronCancellationCosts; FullBaronAttackRequiresWholeSeriesAndScoresNoClubs |
| Underground | FormationUndergroundAndAttachments; FormationEarnedUndergroundDoesNotConsumeSeriesOpening | UndergroundLoanBreakReturnKeepsPrivilegeAndAllocatedIncome |
| People / Great People | FormationProtectionRestrictions; FullPeopleProtectsOnlyNamedSeriesAndBarsOwnClubs; RegisteredOffensiveBindingUndergroundRequiresBothSupports | FormationOpeningBindingAndIsolation; OpeningCancellationKeepsOpeningAllowance; PeopleProtectionRejectsChainsAndSelf |
| Coup | CoupInterruptsDecisionAndKeepsCosts; LifecycleRejectsIllegalDeclarationWithoutPublishingOrMutation; DeferredLegalCoupAtEachBoundary | LifecycleCoupAndDepartureAndNullification; LifecycleCoupPreservesEarnedReceiptsBeforeAbandoningDraws |
| Justice | RegisteredJusticeSubstitutionAndPonziActivation; JusticeRequiresClubSupportWithUnderground; JusticeUndergroundCannotWaiveClubs | JusticeExileInfiltratorConfinementCosts/justice; JusticeRandomOutcomeRestartRetry; JusticeContinuationAndCoupBoundary |
| Code declaration | AbilityTimingSupportAndQuotaMatrix/code; CodePenaltyMixedProtectionConfinementAndUndergroundSupport | AcceptedAbilityCancellationKeepsCostsAndQuota/code; CodeLoanBreakReturnRedeclarationPreservesOrReplacesDeclarer |
| Kidnapper | AbilityTimingSupportAndQuotaMatrix/kidnapper; AbilityKidnapperClosedSelectionMandatory | AcceptedAbilityCancellationKeepsCostsAndQuota/kidnapper; SubstitutedKidnapperReturnsEveryPhysicalCostAndPreservesOwnBinding |
| Ponzi | RegisteredJusticeSubstitutionAndPonziActivation; OutOfTurnPonziConfinement; PonziQuotaUsesOwnTurnCycle | LifecyclePonziExactAlliedShares; OutOfTurnPonziConfinement |
| Ringleader | AbilityTimingSupportAndQuotaMatrix/ringleader | AcceptedAbilityCancellationKeepsCostsAndQuota/ringleader; AbilityRingleaderCostAfterResponses; AbilitySupportLossKeepsQuotaAndCost |
| Richer income | AbilityRicherRequiresSingleAttachedKing; LifecycleRicherCountsAllAttachedKings | LifecycleRoundIncomePrecedesCodeChargeAndResets; passive, no activation quota |
| Richer sacrifice | AbilityTimingSupportAndQuotaMatrix/richer-sacrifice | AcceptedAbilityCancellationKeepsCostsAndQuota/richer-sacrifice |
| Exile | FullExileCommitsQuotaAndAvoidsBarricadeExchange | JusticeExileInfiltratorConfinementCosts/exile |
| Barricade passive | BarricadeVirtualExchange; FullExileCommitsQuotaAndAvoidsBarricadeExchange | VirtualAttackPhysicalExchange; passive exchange uses physical committed Clubs |
| Barricade sacrifice | AbilityTimingSupportAndQuotaMatrix/barricade-sacrifice; AbilityBarricadeDrawRestrictions | AcceptedAbilityCancellationKeepsCostsAndQuota/barricade-sacrifice |
| Infiltrator | InfiltratorDeclarationCommitsReturn | JusticeExileInfiltratorConfinementCosts/infiltrator |
| Negotiator | ResponseTimingSupportAndQuotaMatrix/negotiate; NegotiatorStagesAndRetry | CoupInterruptsDecisionAndKeepsCosts; ImportedNegotiatorRequiresStageActor |
| Dexter assassination | AbilityTimingSupportAndQuotaMatrix/dexter-assassination; AbilityDexterRestrictedThresholdAndBindingRelease | AcceptedAbilityCancellationKeepsCostsAndQuota/dexter-assassination; AbilityAssassinationCountsAttachedLossForCompensation |
| Dexter prevention | DexterDefensiveModifierPrevention; DexterOffensiveModifierDecision; DexterAdvancementAttackAndInvalidPayment | DexterDeclinePreservesDefensiveModifier; DexterPaymentCannotBeRecapturedAsFrozenTarget; ImportedDexterRequiresHostileTargetActor |
| Suicide Bomb passive/destruction | FullBombAttackPhysicalCostAndShield; BaronCombatAlgebra | BombOrdinaryBaronCancellationCosts; FullBombAttackPhysicalCostAndShield |
| Confinement | ResponseTimingSupportAndQuotaMatrix/confinement; ResponsesSkipConfinedButWaitDisconnected | DexterConfinementContinuation; ConfinementWaitsForScheduledTurnDraw; LifecycleConfinementCannotSkipRoundThirteenSettlement |
| Advancement | ResponseTimingSupportAndQuotaMatrix/advancement-defense; DexterAdvancementAttackAndInvalidPayment | DefenseModifierPhysicalUsageAndNoElimination; DexterDeclinePreservesDefensiveModifier |
| Inflation | AbilityTimingSupportAndQuotaMatrix/main-inflation; ResponseTimingSupportAndQuotaMatrix/inflation; DuplicateInflationResponseSpendsNothing | AcceptedAbilityCancellationKeepsCostsAndQuota/main-inflation; InflationNineteenTwentyAndNoCumulativeRemoval; PurchaseInflationPriceAndPrivatePayment |
| Compensation | AbilityTimingSupportAndQuotaMatrix/compensation; ResponseTimingSupportAndQuotaMatrix/compensation-response | AcceptedAbilityCancellationKeepsCostsAndQuota/compensation; CompensationResponseCountsOnlyLaterAttackLosses; RuleBoundaryGoldens/compensation |
| Numerical ace | ResponseTimingSupportAndQuotaMatrix/numerical-defense; DexterOffensiveModifierDecision | DefenseModifierPhysicalUsageAndNoElimination; VirtualAttackPhysicalExchange |
| Ace trade/exchange | ProposalNonblockingRevisionAndAtomicTransfer; ProposalStaleRevisionAndExpiry | AcceptedTransferCancellationKeepsAceUseAndNoAlliance/ace-trade |

D01 closure crosswalk: physical Club commitment, earlier bypass cost and
success-only-cost separation are covered by the ordinary/Baron/Exile/Infiltrator,
Justice and nine-ability cancellation fixtures above; Confinement tests separately
exercise active turn end, out-of-turn non-end, next scheduled expiry, skipped
confined/departed responses, disconnected waits and the independent confined Coup
exception. N01 exact effective values and physical-only exchange are covered by
BaronCombatAlgebra, VirtualAttackPhysicalExchange and the Dexter modifier suite.

D04 closure crosswalk: FourSeatDisjointAlliancesAndThirdPartnerRejection covers
exclusive pairs; LifecyclePonziExactAlliedShares covers paired halves;
SettlementOldestSystemAndInterleaving, SettlementBatchOracleAndResume,
SettlementLargeDenominatorUsefulProgress and SettlementGoldenQ10 cover ordered
FIFO, finite progress, exact batching/reference agreement and reproducible audits.
Lifecycle financial/departure/consent fixtures cover F04/F05; CodePenaltyMixedProtectionConfinementAndUndergroundSupport covers N04;
LifecycleCarriedForgivenessCoupAndVoidProvenance covers N06; FinancialBoundaries
and fresh-match construction cover separate-match isolation. Recommend D01 and
D04 domain-gate completion after the coordinating agent's final verification;
no persistence/server checkbox is implied. D03's named schedule rows now have
concrete legal/illegal and interaction evidence as mapped above, independent of
whether autonomous runs randomly exercise each row.

The added main-ability and response timing/support/quota matrices passed on first
execution (existing behavior, not invented red/green evidence). Final full game
package race run `timing-crosswalk-game-race--20260930T104541Z-2233261` passed in
27.990 s, including the new matrices, cancellation fixtures, F07/N04, alliance and
People-chain cases, and D02 rule-boundary goldens. This supports the domain-gate
recommendations above; campaign precision and server persistence remain separate.

## Final independent D02/D03 domain-gate decision

**Recommend PASS for D02 and D03**, after the bounded additions below and the
recorded passing checks; coordinating agent owns roadmap updates. This is the
named gate decision, not proof over every possible game state and not autonomous
frequency attribution.

D02 clause coverage:

- Q03/private invalid claims: LifecycleRejectsIllegalDeclarationWithoutPublishingOrMutation
  and LifecycleDeclarationPublicProof preserve private rejected claims and publish
  only sufficient accepted proof.
- Q04/protection: NaturalFormationShapes, FormationProtectionRestrictions,
  PeopleProtectionRejectsChainsAndSelf and FullPeopleProtectsOnlyNamedSeriesAndBarsOwnClubs
  cover components, forbidden protection chains and target scope.
- Q06/reset/reopening/history: OpeningHistoryAndAllowance,
  SecondHistoricalOpeningEndsProtectionPermanently, AbilityFateReturnsCostAndPreservesPersistentAce
  and RuleBoundaryGoldens/fate cover retained counters/history and exact reset cost.
- Q07/exhausted supply: SupplyRecyclingAndExhaustion and initiative tie/exhaustion
  fixtures preserve physical supply and committed chance outcomes.
- F06/F08: captured-Diamond, public-history, loan-return, registered binding support,
  reunion/assassination and special substituted Kidnapper fixtures cover destination,
  custody and binding distinctions.
- N02: SourceDepartureKeepsTargetInflation, CompensationDepartureExpiresRemainder,
  Fate fixtures and new OrdinaryClosureDiscardsBothPersistentAceKinds cover retained
  versus expired effect custody and complete physical conservation.
- N07: new RestrictedHoldingsCountForOrdinaryProofButBlockCoupUntilRoundBoundary
  proves restricted cards count toward thirteen-Diamond proof, cannot supply Coup,
  and become available only at round expiry before initiative. Existing Fate,
  captured-Diamond, attachment/binding and acquisition tests preserve markers.
- T02/T03: reviewed rule-boundary golden traces pin purchase, deferred Compensation
  order/exhaustion and Fate; semantic assertions run before digest comparisons.
  New PurchasePostResponseSupplyLossNeverResizesChosenQuantity proves failed final
  supply revalidation retains all payment, draws nothing and never reduces quantity.

D03 clause coverage: the full §2.12 crosswalk above covers every named combo,
non-combo and Ace row, including legal admission, timing/support rejection,
allowances and cost boundaries. Negotiator maximal matching, both Dexter modes,
Justice committed selection/restart, Compensation, Exile/Barricade and physical
modifier algebra have independent fixtures. F07 ownership/custody, allocated
Underground income and N04 mixed penalty exclusions have the F07 matrix; F09
Inflation and N02/N07 are covered above; T01 acceptance/cancellation/return and
T02/T03 exact boundaries have the proposal, cancellation and golden fixtures.
No new runtime changes were required for these final evidence additions.

`d02-restriction-closure--20260930T104855Z-2252938` passed the first two added
fixtures; `d02-d03-boundary-final--20260930T104934Z-2254574` passed all new boundary,
timing and rule-golden tests with race detection (6.679 s). The preceding full game
race run remains `timing-crosswalk-game-race--20260930T104541Z-2233261`; final additions
are tests only. Database/server, exhaustive D08 restart-stage coverage and measured
campaign conclusions remain distinct gates/evidence categories.

## Autonomous exercise attribution snapshot — engine 2.1

Read-only attribution from **completed, atomically published manifests**, captured
2026-09-30. This section does not infer execution from menus or targeted tests and
will not count later campaign shards until their manifests are published.
Restricted reproducible artifacts (ignored in Git):

- `sims/artifacts/gameplay-random-diagnostic-v2-1100/run/manifest.json`:
  three-player development population, one complete cyclic block, six finalized
  one-game matches. Both treatments use the same policies, so there are **three
  distinct trajectories, each executed twice**. Counts below use only jobs 0, 2,
  and 4. Its separate replay verified all six recorded executions.
- `sims/artifacts/gameplay-heldout-v2-20260930/p3-s0000-run/manifest.json`:
  three-player held-out population, one complete paired cyclic block, six
  finalized one-game matches (three rotations × two different treatments).
  `p3-s0000-replay`, `p3-s0000-workers`, and `p3-s0000-workers-replay` are completed
  verification artifacts, **not additional samples or blocks**. The original
  worker count is four and the comparison uses one.
- No completed four-player engine-2.1 shard was available at this snapshot.
  Earlier DEV3 engine-2.0 results remain valid evidence for that retired build,
  not additional current-build samples. No population or development/held-out
  pooling is performed here.

Every listed game reached the round-13 ending and explicit financial finalization;
none ended through ordinary victory or Coup. Development random seats elected
departure in all three distinct rotations. Finalized-game counts do not imply
rare-rule coverage. The campaign remains in progress: this is one held-out block,
not the requested 1,000 blocks per population, and supplies no strength or balance
conclusion.

| Accepted rows | Current autonomous evidence | Exact boundary of inference |
|---|---|---|
| Q07 | 3 development and 6 held-out committed normal deals and initiative resolutions; zero redeals | Verifies exercised normal paths with replay; tie, exhaustion and redeal branches still rely on targeted fixtures |
| Q06, Q04 | 104/309 accepted series-opening commands (development/held-out); 2 held-out formation openings and publicly functioning Kidnapper opportunities | No autonomous Great People, Underground, Justice or Ponzi attribution; reset/history edge cases remain fixture evidence |
| Q01, F02, N01 | 37/114 ordinary attack admissions, 1/2 Bomb attack admissions, 8/15 Dexter assassination admissions, 1/5 numerical-defense admissions | Admission/use is not proof that every declared attack or ability resolved successfully, nor that cancellation branches occurred |
| Q02, Q11, N03 | 3/5 Confinement admissions; complete turn/round progression; 6 development resolve-ready continuations | Replays verify continuation digests; disconnected waits and every skip ordering are targeted-test evidence, not inferred from completion |
| Q08, F01 | 3/8 Negotiator admissions, 3/9 explicit decision commands; held-out 17 Kidnapper admissions and 17 committed Kidnapper outcomes | Outcomes verify recorded selection continuations, not necessarily a successful capture; no autonomous Coup interruption, Justice selection or Dexter modifier-stage attribution |
| T03, F09 | 29/92 purchase admissions and matching 29/92 committed purchase outcomes; 0/5 main-Inflation admissions | Exact purchase completion path exercised; counts do not establish that an inflated purchase or qualifying Inflation removal occurred |
| T02, N02, N07 | Compensation admissions 2/5; one development Richer sacrifice and one ability-outcome continuation | Admission alone does not identify deferred Compensation receipts, marker expiry or persistent-Ace recycling; those remain separately tested |
| Q09, F04 | Development records explicit random-seat departure in all three rotations, followed by legitimate round-limit finalization | Does not establish departed-creditor/Coup interactions or nullification; no autonomous void observed |
| Q12, F05 | 9/18 explicit participant finish-settlement operations followed by 3/6 match finalizations | This proves completed settlement boundaries, not negotiated promise/payment/reputation branches; private financial histories are excluded from this public attribution |
| Q03, F03 | No ordinary or Coup declaration in these snapshots | Declaration correctness remains targeted-fixture evidence |
| Q05, Q10, F06–F08, N04–N06, T01 | No public command/effect evidence sufficient for specific accepted-branch attribution in this snapshot | Loans/binding custody, settlement oracle/batching, Code/Underground, historic retries/corrections and completed negotiated transfer must not be inferred from general game completion |
| F10 | Completed public reports omit hidden cards, seeds, private proposal terms and financial trajectories; private traces and economic snapshots are separate restricted artifacts | Privacy correctness is established by observational-equivalence/report tests and review, not by action frequency |

Numbers separated by `/` above always mean **three distinct development games /
six held-out games**, never pooled estimates. Ability names are accepted command
categories; the table explicitly distinguishes declaration/admission from a
recorded outcome. Digest-only events are not used to invent semantic outcomes.

Representative completed development job `games/000000/gameplay.json` uses
zero-based trace indices: 0 deal, 1 initiative, 2 first turn starts, 30 Ringleader
admission, 60 Compensation admission, 163 Richer sacrifice admission, 190 Dexter
assassination admission, 283 seat 3 departure choice, 474–476 explicit participant
settlement completion, 477 finalization. Its public ending is round-limit and
final scores are [64, 107, 0]. These references disclose no physical card identity,
private seed, proposal term or policy reasoning. Private economic diagnostics
remain under the restricted run directory; they are not exported into this matrix.

## Final campaign attribution and reporting verification

The [findings](gameplay-findings.md) extend the snapshot above to three complete
paired seed blocks per population: 18 three-player and 24 four-player games reached
board closure and financial finalization. Recorded-decision replay completed for
18 and 16; worker 1/4 comparisons matched 6 and 8. The last four-player replay was
canceled at the fixed resource cap. No incomplete replay is promoted to verified.
These counts exclude the historical Phase 0 partial checks, development trajectories,
worker reruns and the retired first campaign.

All three-player endings were round 13; four-player endings included 22 round 13 and
two accepted ordinary Diamond declarations. Three formation openings and 28
Kidnapper uses appeared in the three-player run manifests; Coup, tied ranks,
nullification and several rare accepted interactions still rely on targeted
conformance fixtures. The complete Q/F/N/T matrix remains a test crosswalk;
occurrence in a random sample is a separate column, never inferred from support.

The final CLI 2.2 two-game development tournament supplied eight finalized games
across four completed matches, with fresh IDs, retained completed ledgers and
nonempty carried debt verified at the next-game boundary. Two first-game menu
budget stops left their subsequent slots unstarted; all prefixes replayed.

Final review fixed pre-deal started counters and optional private diagnostic size
metadata without changing engine/policy files. A canonical rational-object analysis
regression then corrected campaign aggregation through an immutable analysis-only
artifact; original results, budgets and seed selection were preserved. See findings
for exact logs, version hashes, public projections and unmet evaluation targets.

## V3 verification and development evidence

The [latest findings](gameplay-findings.md#v3-development-investigation-and-verification)
and [prospective allocation](../experiments/gameplay-evidence-v3.md) distinguish
accepted-rule conformance, deterministic equivalence, policy fixtures and population
claims. None automatically completes P03 or online persistence gates.

| Evidence | Verified result | Limit |
|---|---|---|
| Original six v2 run manifests | All artifact checksums and byte counts agree; 18/24 finalized games, three paired blocks per population | Historical samples remain unchanged |
| Outstanding four-player replay | All eight games replayed with the pinned original binary; all 88 artifacts verified and outcomes byte-identical | Verification adds zero independent samples; historical totals become 18/24 completed-game replays |
| Canonical serialization and observation optimization | Legacy byte/error oracle, ASCII/Unicode quoting oracle and observation equality; fixed full transcripts unchanged | Two fixed development inputs, not exhaustive proof or sustained campaign throughput |
| Opportunity-policy regression fixtures | Early royal retention, free-gift acceptance, affordable reduced promise ranking and own-finance decision context | Local competence, not demonstrated strength or human-quality negotiation |
| Fair information boundary | Detached own cash/debt and public horizon; unrelated-party debt changes leave projection unchanged; existing hidden-card/menu noninterference | Tested channels only; no omniscient decisions authorized |
| Menu audit | Two independently legal, strategically distinct five-point payments; sampled menu retains only one | Action coverage remains sampled, with no uniform-all-actions or exhaustive strategic claim |
| New estimator and claim extraction | Exact per-game block contrasts; whole-block incomplete exclusion; homogeneous physical-seat contrasts; shared tied-win credit; treatment denominators; simultaneous intervals | No fresh confirmation cohort; one complete development block has no interval |
| Frozen replay selection | Schema-v2 dispatch verifies only preregistered logical shards; worker checks require selected replay; selection changes fail resume | Planned slots are not completed replay credit; historical schema-v1 behavior remains unchanged |
| Frozen analysis implementation | Schema-v3 retains checksummed runner/claim-extractor source copies; mid-campaign and resume source tampering prevent dispatch | New confirmation capability; no retrospective source-freeze claim for schema-v1/v2 cohorts |

The first new three-game development campaign planned **36/48** game slots across
two blocks per population. It started **18/22**, finalized **18/18**, completed
**6/4** matches, and left **18/26** slots never started. Complete paired blocks are
**1/0**; completed replay credit is **18/0**. Four four-player matches stopped,
two for transcript bytes and two for menu-generation operations. The first
three-player block's candidate-minus-economic effect is **+253/9 points per game**,
descriptive only; no four-player block effect exists. No finance-priority activation
was observed, so cross-game strategic improvement remains untested by these runs.

These failures identify practical evaluation limits, not automatic evidence of an
unbalanced game or a weaker candidate. Subsequent larger-cap development runs must
have new immutable manifests and remain separate controlled budget experiments.
Seat fairness, broad opponent robustness, dominance/counterplay and mechanic-specific
balance claims remain inconclusive until adequately powered confirmation exists.

Restricted artifacts are under `sims/artifacts/gameplay-historical-verification-v3/`,
`sims/artifacts/performance-20260930/`, and
`sims/artifacts/gameplay-evidence-v3-development/`. Shareable findings publish only
aggregate finalized outcomes, public-rule explanations and artifact references.

The subsequent controlled larger-budget rerun at
`sims/artifacts/gameplay-evidence-v3-budget-development/` finalized all **18/24**
planned games and **6/8** matches, with one complete development block per
population. Its explicit 1,000,000-operation menu cap and 32-MiB transcript profile
preserved all 14 prior committed prefixes, including unchanged full traces for
all ten previously completed matches. Checksummed run/replay artifacts passed
inspection. Three-player replay completed all 18 games; four-player replay was
canceled at the fixed wall allocation and earns zero verification credit.
The per-game effects **+253/9** and **−83/12** have no confidence intervals with
one block. Reruns of these same development inputs are not additional independent
samples. Completion under the larger envelope is supported for these inputs;
population strength, seat equivalence and balance remain inconclusive.

Additional schema-v3 calibration completed **36/48** games and **12/16** matches,
two new development blocks per population, with zero incomplete/never-started
slots and no replay dispatch. Together with the larger-budget first block this
is **three unique development blocks per population**, **54/72** games and
**18/24** matches, not five blocks obtained by counting reruns. Exact per-game
means are **460/27** and **341/36**; neither is confirmation. Three-block variance
projects K=42 halfwidths of **1.638/1.463 points per game** at 1,000 blocks, with
substantial variance-estimation uncertainty. The findings record sensitivity and
all underlying block effects. Separate homogeneous seat/robustness variances and
the untouched confirmation cohorts remain absent.

The separate four-player retention ablation completed **24/24** games and **8/8**
matches in one paired development block, with no replay dispatch. Disabling early
royal retention gained **20 points per game** against the otherwise unchanged
opportunity policy on that input. All four retained-policy baseline scores and
step counts reproduce the earlier candidate treatment; action content agrees
after job-identity normalization. A subsequently discovered legacy reconfiguration
tie-key defect makes isolated attribution to retention inconclusive; the contrast
is preserved and excluded from primary effect and independent-block counts.

The bargaining diagnostic completed **18/18** games and **6/6** matches, one
three-player block, with a **+4/3-point/game** focal contrast and no replay.
No selected trade offer/acceptance/loan or carried-debt priority activation occurred.
Private-source and artifact hashes passed inspection. Investigation identified
the legacy `gameplayActionKey` overwriting supplied reconfiguration specifications
with an existing formation's old specification. Distinct protection choices
collapsed to one tie key, allowing administrative-ID-dependent menu order to
affect decisions. Actual first divergent reconfigurations confirm this path.
This is an implementation defect; corrected behavior requires an explicit new
version and new evidence, not retrospective positive strength or balance claims.
Historical economic/pressure v1 comparisons share the legacy key and may be
affected; their original snapshots and preregistered thresholds remain preserved.

A separate aggregate scan of both calibration blocks found zero selected
carried-debt income-priority or affordable reduced-offer activations across all
6 three-player and 8 four-player candidate match artifacts. Combined with the
first budget-development block, financial strategy effectiveness therefore remains
unestablished throughout the primary development sample. Finance tie-key review
found no analogous concrete reconfiguration overwrite, but neither that review
nor the focused board regression establishes universal administrative-ID
invariance across all financial and negotiation paths.
A remaining finance-key limitation is explicit: an authorized public system-debt
forgiveness request for another debtor is absent from the observer's party-only
debt list, so its action key retains the raw debt reference. Current deterministic
policies rank forgiveness below declining; this review did not demonstrate a
selected-action effect in the primary cohort. It still prevents a claim of
universal identity invariance, especially for random financial selection.

The correction is now implemented as six explicit deterministic `@v2` policy
aliases. The legacy key remains source-identical to the pre-task staged version.
The corrected key retains the requested reconfiguration specification and uses
public formation kind plus sorted physical membership for references. Full Go
regression tests and `go vet` pass. Coverage includes an actual generated legal
menu with independently applicable competing protection reconfigurations, the
preserved legacy collision, corrected distinct keys, reordered-menu and renamed
formation fixtures, unchanged observation/menu under a hidden-card swap, and
financial alias operation/count equivalence. These are focused regression
results: no corrected-version full-game confirmation was collected, and they do
not establish universal finance or negotiation identity invariance.

Final v3 verification: full Go suite and vet pass; targeted race checks including
v2 ordering, privacy, replay, continuation, cancellation and resource bounds pass
(simulator 187.898 s). All 67 Python analysis/provenance/storage tests pass. Protected
logs: `sims/artifacts/verification-v2-final-20260930/`; final Python safe-run
`evidence-v3-final-python--20260930T125819Z-2590919`. These are regression gates,
not corrected-policy population evidence: new confirmation remains zero.
