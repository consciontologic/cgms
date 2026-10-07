---
name: cgms-rule-conformance
description: "Build or review CGMS accepted-rule fixtures, exact rational accounting, FIFO settlement and partial-coverage evidence."
---

# CGMS rule conformance and exact accounting

Use for accepted gameplay/ledger conformance, especially S02/S05–S07. This skill
specifies future runtime gates; preparation alone does not satisfy them.

1. Read relevant [GAME_RULES](../../../docs/project/GAME_RULES.md) clauses and
   [accepted register](../../../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md).
   Q01–Q12/F01–F10/N01–N07/T01–T03 are accepted. Classify each fixture as implemented,
   expected unsupported, or named experimental; never reopen accepted choices.
2. For each implemented trace record actor/inputs, physical zones, current versus
   historical openings, counters, decision IDs/stages, event order, exact ledgers
   and each seat's permitted observation. Couple expected outcomes to named clauses,
   not an implementation-derived snapshot alone. Use the
   [execution guide](../../../sims/docs/phase-0-execution-guide.md) for slice mapping.
3. Keep score, available cash, debts and match adjustments separate. Use reduced
   arbitrary-precision rational values with positive denominator and decimal-string
   serialization; clone mutable big-number state. Test legal paired Ponzi halves;
   thirds are generic arithmetic fixtures, not three-person alliance gameplay.
4. Implement a simple unbatched Q10 oracle before batching: fixed-seat gross
   receipts, FIFO recipient queue, oldest-created debt first, no new obligations
   and no reciprocal cancellation. Prove finite termination by scaling the finite
   rational system to integer units: each positive debt payment reduces total
   outstanding debt, and queue cleanup without payment is finite.
5. Differential-test optimized batches against the oracle on small exhaustive and
   generated ledgers, including reciprocal debts of 1 with receipt 1/D. Those can
   require 2D transfers; termination is not a practical cost bound. Validate exact
   end ledgers, queue order, receipt totals and compressed-audit reconstruction,
   not merely score equality. Retain resumable cursor/ledger/provenance progress;
   work limits pause without erasing obligations or replaying all prior work.
6. Cover F04/N06 boundary ledgers: next-game cash zero, old debts retained, separate
   matches reset, completed results immutable, forgiveness linked to original
   charge/creating game, Coup and nullification treatment distinct. Verify no
   duplicate debtor deduction and no manufactured creditor cash from correction.

Reject illegal declarations without costs; retain legally committed Clubs and
resolved response costs on later cancellation (Q01/F02). Include N01 virtual versus
physical values, N03 eligible response order, N02 effect zones, N07 physical
availability, and T02 deferred draws where the slice requires them. Test projections
for hidden marker/remote binding leaks alongside conservation and no-alias checks.

Conformance claims name covered clauses and excluded transitions. Complete fixture
coverage is a broader target, not permission to substitute approximations or pull
Phase 1 into Phase 0. Follow [format rules](../../../sims/docs/phase-0-format-spec.md)
for evidence data and [Phase 0 instructions](../../instructions/PHASE_0_SIMULATION.md)
for review/recovery.
