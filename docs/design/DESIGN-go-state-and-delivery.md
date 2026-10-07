# Design: Go match state and delivery

- **Status**: single-process Phase 1/2 implementation verified locally; deployment pending
- **Author**: Codex
- **Date**: 2026-09-28
- **Supersedes**: none

## Problem

Retries, lost sockets and competing declarations must not duplicate cards,
scores or winners. A human response window can last much longer than a database
transaction. The [architecture](../code/ARCHITECTURE.md) needs a durable boundary
covering a match's games, balances and surviving debts.

## Goals

One committed order, atomic game effects, confidential projections and safe
recovery after an unknown commit. The local implementation follows this design; digital arbitration
and disconnection follow accepted Q11/F01–F10/N01–N07/T01–T03 in the [resolved rule register](RULES-IMPLEMENTATION-QUESTIONS.md)
and [GAME_RULES](../project/GAME_RULES.md). The roadmap records database/race/scenario evidence; production rollout remains pending.

## Non-goals

Distributed actor leases, a message broker, offline-result trust and changing
the game's untimed turn/response rules.

## Proposed design

**Durable unit.** One PostgreSQL match row serializes every match mutation,
including game completion and debt carryover. At READ COMMITTED, first lock
that row `FOR UPDATE`, then load snapshot and ledger through the same `sql.Tx`.
No in-memory state or Redis value justifies a write. Check a unique
`(match_id, actor_id, command_id)` record before game-instance or expected-version
validation; the same ID/body returns its original result, a different body is a
conflict. Require immutable `game_id` in every game-scoped intent and fingerprint,
including victory and settlement. After duplicate lookup, reject a new intent
whose game differs from the current instance. IDs for turns/windows/actions/
effects/decisions belong to that game and must match when relevant. Next-game
lifecycle requests name the finalized predecessor and generate a fresh game ID;
a nullified replacement also gets a fresh ID despite reusing its match slot.
The first game gets its ID at match start. Queued client intent never migrates
to a later game automatically. For ordinary commands validate expected version
and selected state. Victory intents instead validate the current board phase, holdings and eligibility:
never reject Coup solely for an unrelated version change within its named game.
Ordinary threshold claims get the same admission treatment only at a permitted completed-action
boundary. Relevant eligibility changes still reject them. Effect inputs must
also match the live effect/decision/stage and required actor.
Commit the snapshot/version, ledger/debt allocations, deterministic events,
per-seat delivery payloads and command result together. Unique event/effect
keys prevent duplicate results. Follow Go's transaction-handle discipline and
PostgreSQL's locking behavior. [Go transactions](https://go.dev/doc/database/execute-transactions),
[PostgreSQL locks](https://www.postgresql.org/docs/current/explicit-locking.html)
(accessed 2026-09-28).

Serialize account/economy mutations by sorted account IDs before a match lock
where both are needed; all such paths use that order, including game rewards.
Promo redemption locks its code row before sorted account rows; no path may
acquire a code row after locking an account or match. Test competing redemption,
revocation and account-lifecycle operations against this global order.
Prefer a durable reward intent in the final-game transaction and an idempotent
economy worker afterward. Its effect ID is `(game_id, player_id, reward_kind)`;
consuming the intent and crediting dirt happen in one database transaction.
Any failed/unknown commit is reconciled by that effect ID before replay.
Do not claim a final-game response means the asynchronous reward is delivered.

**Human timing.** Persist declaration, response cursor and effect-specific
decision records: effect ID, decision ID, required actor, kind, stage and
authorized options. Commit and release the lock after every command. A response
can need several inputs without creating a nested response window. Accepted
activations cannot be withdrawn. A canceled action closes its window. Coup may
run at the next serialized transition even while an effect waits for input,
including a confined actor's legal Coup; ordinary declarations require a
completed action and an actor who is not confined. Never interrupt an atomic
mutation halfway. Accepted digital victory order is the first valid
declaration processed in the authoritative server order; persist that order
and terminal result atomically. Client timestamps are untrusted. Preserve
active-player/turn priority for competing ordinary action declarations without
delaying valid victory declarations. Disconnection/silence preserves pending
state indefinitely for reconnection or unanimous resignation; a socket timeout
never auto-passes, quits or forfeits. Cursor traversal omits confined and departed
seats deterministically, with no impossible required pass. Preserve the remaining
fixed seat order; disconnected but eligible seats remain pending. Re-evaluate
normal eligibility after Confinement's scheduled expiry, retaining its separate
Coup exception. Test the final-response crash/retry
boundary against Q01/Q08/F01/F02 without repeating costs or club usage.
Negotiator records attacker and defender choices, then commits added-Club usage,
Spade transfer, jack return and cancellation together; original committed Clubs
remain used throughout. Justice records each fixed-seat selection and reserves
each committed
random result, transferring all selected cards and returning the queen only at
final resolution. Coup preserves committed knowledge but abandons uncommitted
movements. Dexter records its required prevention decision. Store visibility
with the committed result so reconnect never rerolls or discloses an unselected
card. Ability costs, quotas and resets come from GAME_RULES §2.12; counters
are keyed by player/ability/interval except ordinary physical-Club use.

**Proposals and accepted transfers.** GAME_RULES §2.15 proposals are nonblocking
records with offer ID, revision, exact terms/parties, origin game/turn and status;
they reserve no cards or allowances. Only named parties see concealed terms.
Creation/revision needs the permitted idle/turn boundary and eligible parties;
a revision cannot extend expiry. Maker withdrawal, recipient decline, acceptance
and deterministic expiry serialize under the match lock. Expire unanswered
offers at origin-turn end, board closure or party departure; no clock deadline
or inferred acceptance/refusal is involved.

Acceptance names the exact revision, rechecks current ownership/availability,
transfer restrictions/quotas, original turn and alliance compatibility, and
requires no pending action or automatic work. Invalid current terms make
acceptance unavailable without blocking play. Receiving an intact loan requires
no ability support; support is checked when using it. Legal acceptance starts
one nonwithdrawable action whose actor is the accepting player, including for
response order and Confinement. Ace suppliers spend their allowance at acceptance;
no opening quota is spent. Final resolution revalidates and atomically transfers
all agreed cards plus any alliance/loan-derived Underground privilege. Cancellation
or invalid resolution closes that offer without movement/privilege and retains
earlier committed allowances/responses; retry never reopens or reapplies it.

**Deferred draws.** After the action/responses and required returns/shuffle,
create the Compensation draw queue by fixed ascending seat. Finish each
recipient's owed quantity before advancing. Each draw follows normal recycling/
exhaustion; consumed five-loss entitlements create no later claim when supply is
empty. Persist queue ID, originating action/game, recipient position, remaining
quantity, counters and committed random draws if work spans transactions. Every
draw/cursor update commits together with its authorized projection. Complete the
queue before point settlement or another ordinary action/victory. Coup alone
may interrupt between atomic queue steps, preserving committed draws and closing
the uncommitted remainder with the board. This queue is distinct from the exact
financial-settlement workspace described below.

**Deck purchases.** GAME_RULES §2.16 binds explicit positive integer quantity
and specific unrestricted number Spades from hand/exposed series. At the own-turn
idle boundary validate exact effective value against quantity × current price
and full draw supply including returned payment. No generic two-series or
Underground requirement applies. The buyer is the ordinary action actor; payment
is a resolution cost. Revalidate unchanged selections/quantity after responses,
then commit payment return, shuffle, all draws and any qualifying post-pricing
Inflation removal together. No partial purchase, automatic resize or second
draw on retry. Hidden payment selections remain private until public return;
server random outcomes and per-seat events obey the same transaction boundary.

Physical state also records active-effect Aces with source/target/custodian/expiry
and physical-card `available_from_round`. Inflation's target is its physical
custodian; Compensation targets its owner. These exposed effect-zone cards have
no reusable gameplay control and are excluded from Fate/reset counts, preserving
the named effect and Compensation remainder. Source departure leaves another
target's Inflation active; target departure ends/discards its active effects,
and board closure cleans up all active Aces after required end evaluation.
Restriction markers survive involuntary movement, Fate, shuffle and redraw until
the recorded round; they do not reveal concealed custody. Mandatory Diamond
exposure still occurs. Enforce §2.14 for voluntary action/Coup eligibility while
allowing ordinary threshold counting and proof.
If a future experimental interpretation has undefined settlement only after
a legal response resolved, persist its effects and a recoverable adjudication
marker as one successful transition; retry returns the same result. Reject
before execution only when the incoming command has no resolved effect.
Accepted Q01–Q12/F01–F10/N01–N07/T01–T03 are not unresolved branches; production profiles must fully
implement GAME_RULES, and incompatible versions fail before match start.

**Replay identity/randomness.** Persist engine/rules/interpretation hashes,
commands and actual random outcomes. Go owns a versioned deterministic random
stream in simulations; production uses a cryptographically secure generator
and unbiased sampling, with its seeds/state kept secret. An unpredictable
seed alone does not make a weak PRNG suitable for online hidden draws.
Transaction rollback must not advance the durable stream. Once a random result
has been committed or revealed, its durable outcome belongs to the effect and
survives continuation/reconnect without resampling. Stored outcomes make replay
independent of
later random-library changes; old engines remain available for old matches.

**Unknown commit/retry.** A lost commit response is `OUTCOME_UNKNOWN`, not a
rollback. Read the durable command result on the primary with fresh bounded
context, retaining the same operation identity. If still uncertain, leave the
client pending. Replaying the identical ID is safe only through the same
locked/unique command boundary, even if the earlier request is still running.
The client operation coordinator alone owns automatic command retry: at most
two submissions and two reconciliation reads within one five-second deadline;
the server gets the remaining budget capped at two seconds per request.
Server SQL, proxy and transport layers make one attempt: maximum mutation
attempts are 2 × 1 × 1 = 2. Reconnect does not create a new command or reset
this budget. On exhaustion, keep the ID for a later explicit reconciliation.
Validation, authorization, illegal action and ID/body conflict are permanent;
confirmed rolled-back transient database contention/unavailability may replay;
ordinary-command version conflict needs resync and player intent confirmation;
victory current-state admission does not create that conflict for unrelated
changes. Ambiguous
outcomes always reconcile before replay. These time budgets bound transport
work, not game turns.

**Delivery.** Persist recipient-filtered events with stable `(match_id, seq)`
identities. For each seat emit a safe envelope per sequence (possibly no-op),
so gap detection does not infer a secret action from missing sequence numbers.
Never publish the authoritative private event and ask clients to hide it.
Apply GAME_RULES §2.13 to snapshots, envelopes, histories and bots: public hand
and concealed-Ace counts; private faces/hidden royal breakdown; public captured
number Diamonds, including Justice captures. Preserve already-public identities
in authorized history, but never infer/reveal another concealed king's location
from a persistent Doppelganger binding. Model lasting player effects separately
from card custody and filter both deliberately. Include immutable game identity
in snapshots and game events; retain history's original IDs across nullification.
Keep physical selections, effective combat values/modifiers and physical costs
separate, following §2.5 algebra. Code's global economic settings apply to all
opponents; derive hostile recurring-penalty eligibility from its active,
non-confined declarer's current two series and each opponent's protection/Confinement at round end.
Snapshot reads return state and cursor from one consistent database snapshot;
subscribe from the next sequence. Buffer boundedly during catch-up, deduplicate
events, and resnapshot if retention removed the cursor. Client acknowledgement
means locally applied projection, not a game effect or safe global deletion.
An unknown/lost acknowledgement replays events with the same identities;
durable command results and client sequence deduplication make it harmless.
Polling committed events repairs a lost wake-up; Redis Pub/Sub can only be a
hint, since its delivery is at-most-once. [Redis delivery semantics](https://redis.io/docs/latest/develop/pubsub/)
(accessed 2026-09-28). No broker ACK boundary or lease exists in v1.

**Lifecycle/capacity.** Proposed starting caps: 100 active match queues,
32 pending commands per match, 16 admitted transaction workers, 20 SQL
connections, 400 sockets and 64 outbound envelopes per socket. Acquire global
and match admission before spawning work; reject overload explicitly. A
process-owned supervisor drains queues; request cancellation removes work not
yet admitted and bounds admitted DB work. Each socket owns one read and one
write loop, canceling and joining both on close. A full socket buffer forces
resync, never blocks a match. Worker/sweeper contexts derive from process
shutdown, each DB call has a deadline, and workers are joined. Stop admission,
drain within a proposed ten-second grace, cancel remaining work and reconcile
any unknown results on restart. These are initial limits to load-test.

**Financial phases.** Pin the visible positive game count at match start
(default three). Persist `playing`, `board_closed_settlement` and
`financially_finalized` separately. Board closure immediately freezes card play
and further victory claims; it does not close the financial command path.
Typed promises name origin game, award, recipient, amount/percentage and
condition. During settlement accept only the specified payments, final offers,
acceptances/rejections/refusals, financial consent and finish commands. Settle
automatic debts before spendable offers; never infer refusal from absence.
Finalize only after required players explicitly finish their resolved promises
and debt work is complete. Next-game admission then starts zero available cash,
preserving completed scores and outstanding match debts. Do not run concurrent
games to bypass pending human settlement. Final match scores award shared rank
for ties, separately from threshold/Coup declaration outcomes.

Use GAME_RULES §4.4 as the field-level transition source: Coup clears active
participants' old current-game cash, preserves completed transfers, carries
surviving unpaid adjustments exactly once and applies replacement awards;
departed accounts retain their exemption. Ordinary finalization closes the
origin game's spendable pool. Departure, nullification and forgiveness each
have explicit ledger transitions. Record origin-game attribution on every
award/transfer and retain final results; an untagged match-wide wallet is not
sufficient to enforce current-game Ponzi or sequential settlement.
For cross-game forgiveness, append a nonspendable match adjustment with unique
effect ID, original debt/charged-entry lineage, exact amount and creating game ID.
It corrects a charge retained in a completed result without rewriting that result,
survives subsequent Coup replacement and never feeds cash or repayment queues.
Current-game forgiveness instead corrects that game's still-recorded charge;
Coup replaces both charge and correction together. Preserve carried-charge
lineage to prevent duplicate corrections. Nullification restores debt and the
effective adjustment set from its start snapshot; retain void entries for audit.
Compute cumulative totals from finalized/current game results plus effective
match adjustments, exposing the components to clients.

**Accounting order.** Exclusive alliances contain at most one partner per
player; allied Ponzi first creates two equal gross receipts, not a transitive
recipient set. Compute simultaneous awards before repayment, enqueue their
receipts in fixed seat order, then process creditor receipts FIFO with each
recipient's oldest debt first. Reciprocal debt is not canceled automatically.
No new obligation arises from settlement, and every positive transfer reduces
outstanding debt by at least one unit of the finite rational inputs' common
denominator; use this termination argument and prove optimized batching
equivalent. The finite bound can be impractical: two reciprocal debts of 1
with receipt 1/D require 2D unbatched transfers. Implement exact batching with
reproducible compressed audit records; prove equivalence against unbatched
small cases. Persist a settlement identity, frozen inputs, FIFO/debt cursor,
exact accumulators, version and committed progress at safe batch boundaries in
a private settlement workspace. Public score/cash/debt results remain at the
last complete effect until final publication; progress is not a partly paid
public award. Retrying must resume committed work, not continually restart the same over-budget
transaction. Each workspace/cursor/audit commit is atomic; final publication
atomically replaces the authoritative ledger/snapshot and emits its safe events.
The initiating command's durable result may report committed pending settlement;
completion is a later identified event/state transition, never a falsely final
award. Reads/reconnect/reconciliation can report progress, but no gameplay or
next financial operation overtakes it. Measure work and rational growth;
technical limits must preserve
recoverable settlement without accepting a later game action or dropping debt.
Debts and departed-player records survive games within the match; a separate
match begins clean. Resignation restores the start-of-game financial state only
with every affected player's consent; forgiveness affects unpaid remainders
without reversing completed transfers. Snapshot, ledger and consent records
need the same transactional/fault-test boundary.

## Alternatives considered

An in-memory actor alone loses acknowledged state on crash. A Redis lease
does not stop a paused old owner from writing; it would need authoritative
fencing. Full event sourcing adds replay/migration obligations beyond the
snapshot plus journal needed here. Separate services/queues are deferred until
measurements justify their additional effect boundaries.

## Migration / rollout

Start with simulation, then a single backend replica and real PostgreSQL
fault tests. Use expand/backfill/validate/contract schema migrations, retaining
old snapshot readers while old matches drain. Do not silently reinterpret a
live match after deploy. Backups preserve snapshots, ledgers, journal, command
identity and pinned configuration together. A primary outage stops mutations;
Redis/telemetry outages permit core play with bounded local fallback. Stale
replicas never answer correctness or command-outcome queries.

## Risks

Test every F01–F10, N01–N07 and T01–T03 cross-layer scenario in the rule register, two simultaneous
victories, duplicate Ponzi, transaction crash before/
after commit, lost success responses, reconnect gaps, stale snapshots, Redis
loss, slow clients, queue overload and shutdown during settlement. Exercise
race/leak checks and cycle-heavy exact debt settlement. Q10 settles the cycle
economics and ordering; termination, practical work bounds and batching
equivalence still require evidence. Never silently forgive debt. The engine/durable/transport suites provide targeted runtime conformance evidence;
production capacity and deployment remain unverified;
accepted rules do not complete the proposed architecture.
In particular, delayed first submissions from an earlier/nullified game must
reject, old committed duplicates must replay their original result, and same-game
Coup still tolerates unrelated versions. Exercise active-effect card conservation
through Fate/departure/closure, confined-seat traversal, Code eligibility,
forgiveness before Coup/nullification and hidden restriction-marker movement.
Exercise T01 acceptance/revision/withdrawal/expiry races, canceled loan with no
early alliance/privilege, T02 exhausted cross-player draws and restart/Coup at
queue steps, and T03 chosen quantities/payment zones with post-response supply
revalidation. All use the same accepted engine traces and immutable game scope.

## Open questions

Retention/deletion horizons and measured capacity remain product/operational
decisions. Q09–Q11 already settle match/debt lifetime, declaration ordering
and disconnection. Future automatic abandoned-game termination would require
a separate rule revision; it is not a blocker for accepted untimed play.
Persist hibernated waits without changing outcomes. Public/monetized release
still needs the abandoned-game allowance/reward policy recorded, without using
a technical timeout as a substitute for a rule change.
See the resolved rule register and roadmap implementation gates.


## Implemented local limits (2026-09-30)

The [API run contract](../code/API-game.md) records the current bounds and
selected identity/retention policy. `cmd/server` holds a PostgreSQL advisory lock
to reject accidental duplicate startup and shuts down on that connection's loss.
It does not implement distributed fencing: never run a hot standby against the
same authority. Redis only stores transient hashed presence hints. Durable cursor
polling means dropped hints cannot lose game transitions. Catch-up accepts at
most 1,000 revisions of lag, 128 events and 1 MiB; outside that window, resnapshot.
A persistent private handle key keeps original wire receipts replayable across
normal restarts. Database backup/recovery must preserve it with match state.

Queued victory declarations are selected before ordinary rank queries already
queued at that selection boundary. New arrivals do not interrupt an atomic
transition or a rank query already running. Ordinary priorities are recomputed
from the authoritative active player/round order. Two-second command contexts,
bounded queues and socket ownership are implementation limits, never human-turn
deadlines. Existing disconnected decisions remain durable indefinitely.
