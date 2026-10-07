# S07 bounded accounting evidence

This is scenario-level GAME_RULES §§4.2–4.4/Q10/F04/N06 coverage, not complete
alliance, promise, departure/card-cleanup or game conformance. Financial seat
indices are zero-based; system creditor is -1. External schemas/adapters must
preserve that convention or translate explicitly at their boundary.

Observed TDD failures and repaired behavior (logs under `/tmp/agent-runs/`):

- `s07-amount-red--20260930T081507Z-1582708.log`: exact thirds returned zero;
  `s07-amount-green--20260930T081538Z-1583490.log` passed.
- `s07-charge-red--20260930T081610Z-1584589.log`: charge did not debit/create debt.
- `s07-settlement-red--20260930T081648Z-1585900.log`: six-point receipt did not pay;
  `s07-settlement-green--20260930T081803Z-1589274.log` passed both slices.
- `s07-boundary-red--20260930T081820Z-1590802.log`: Coup cash and forgiveness;
  `s07-boundary-green--20260930T081932Z-1593704.log` passed.
- `s07-lineage-red--20260930T082207Z-1599775.log`: changed creditor retry accepted;
  `s07-lineage-green--20260930T082219Z-1600338.log` passed.
- `s07-validation-red--20260930T082305Z-1604793.log`: missing charge provenance;
  `s07-validation-green--20260930T082321Z-1605503.log` passed.
- `s07-head-red--20260930T082426Z-1608285.log`: no partially consumed head;
  `s07-final--20260930T082447Z-1609497.log` passed after durable head accounting.

Reproduce from repository root:

```sh
pwd
xops/agent/safe-run.sh s07-tests -- go -C backend test ./internal/game -run 'Amount|Ledger|Settlement|Financial|Forgiveness|ImmediateCharge|ChargeRetry' -count=1
xops/agent/safe-run.sh s07-race -- go -C backend test -race ./internal/game -run 'Amount|Ledger|Settlement|Financial|Forgiveness|ImmediateCharge|ChargeRetry' -count=1
xops/agent/safe-run.sh s07-fuzz -- go -C backend test ./internal/game -run '^$' -fuzz FuzzSettlementEquivalence -fuzztime=2s -parallel=2
```

Final race run `s07-race--20260930T082456Z-1609864.log` passed. Final bounded
fuzz run `s07-final-fuzz--20260930T082457Z-1610885.log` passed 52 executions
(41 new interesting inputs); an earlier pre-canonical-hash run passed 557.
These are observed test counts, not balance/performance release thresholds.

Tests cover exact paired 11/2 Ponzi shares, generic thirds, 6 then 8 payments,
oldest system-first payment, current versus prior-game charges, retained unpaid
4 through Coup, next-game zero cash, departure exemption, explicit settlement
finish, immutable completed results, nonspendable forgiveness and nullification.
The curated Q10 golden specifies six ordered repayments and exact final balances.
Two-/three-seat simple cycles at D=1,2,3,7 compare optimized ledgers and expanded
ordered payment traces with the unbatched oracle. Twenty-five small interleaved
receipt/system-sink cases and generated ledgers exercise fallback behavior.
JSON serialization/resume is tested at every segment boundary, with corrupt digest
rejection. A 51-digit denominator (approximately 167 bits) requires two optimized
durable operations, two compressed audit segments; the initial measured serialized
checkpoint was 2804 bytes before the explicit head flag was added. This measurement
is historical and not a fixed golden size.

Termination: scale finite debts/receipts to common denominator D. Every positive
payment reduces total remaining debt by at least 1/D. No new debts enter settlement;
only payments enqueue further receipts, so cleanup is finite. This is not a
practical bound. The optimizer skips only complete repetitions of a simple directed
cycle with one queued receipt, identical oldest debt heads and unchanged transfer
size. The minimum floor(debt/receipt) bounds repetitions at the first exhaustion.
Other shapes use the reference procedure. Each durable segment has canonical
before/after hashes, a prior-segment link, exact receipts/payments and repeat count.

Integration obligations: preserve the cursor and audit in private artifacts; use
`BeginSettlement`/`ResumeSettlement` with bounded budgets for CLI/cancellation.
`SettleReceipts` and lifecycle wrappers are synchronous conveniences for bounded
curated scenarios, not a cancellation interface. `Committed` remains unchanged
until Done; commands/Coup must not interleave with working settlement. A partially
consumed FIFO head and its already-credited flag survive resume. The engine owns no
clock, filesystem or cancellation signal. F05 promise workflows are later scope;
scenario `FinishSettlement` assumes no unresolved promises. Callers retain the
start snapshot for `Nullify` and audit/history alongside convenience ledger results.

## Review hardening

`TestFinancialValidationRejectsCorruptProvenance` first failed 16 malicious-input
cases in `s07-review-red--20260930T082922Z-1624897.log`, then passed in
`s07-review-green--20260930T083008Z-1627737.log`. Validation now requires one debt
per charge, complete origin-game history, unique ordered debt identities, dimensionally
valid completed results and corrections linked to their exact debt/charge/debtor,
creating game, operation and original amount. Aggregate corrections plus remaining
debt cannot exceed the original charge. Mutators validate before indexing state;
`MatchScores` returns nil on corrupt input instead of panicking or inventing a total.

The synchronous convenience limitation is now enforced, superseding the advisory
wording above: `SettleReceipts` refuses cyclic or larger-than-64-debt/receipt shapes
before mutation with `ErrSettlementRequiresContinuation`. The bounded cursor remains
the supported path for arbitrary legal cycles. Lifecycle wrappers return their
original ledger on this admission error, with no partial charge/transfer/Ponzi.
The regression failed in `s07-sync-red--20260930T083030Z-1628348.log` and passed in
`s07-sync-green--20260930T083054Z-1629268.log`. The synchronous acyclic shape has at
most four seats, bounded splits and no denominator-dependent circulation.

Resume now revalidates ledger/provenance, seat indices, amounts and progress flags
in addition to its digest. Rehashed malformed snapshots failed the regression in
`s07-resume-validation-red--20260930T083118Z-1631612.log` before the fix. Final review
regressions and race checks passed in `s07-review-final--20260930T083133Z-1632230.log`.
