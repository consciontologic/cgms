# Module: shared game engine

Status: shared engine implemented for D01–D04 and local simulator/domain
continuations, with a separate PostgreSQL match adapter (2026-09-30). Public
server transport is implemented separately in Phase 2; adaptive web integration is implemented; native acceptance and release remain open. This document
specializes the [architecture](ARCHITECTURE.md) and [simulation ADR](../design/ADR-0001-shared-go-simulation-engine.md);
it does not approve the ADR or freeze Go symbol names.

## Purpose

Apply the [canonical CGMS rules](../project/GAME_RULES.md) consistently in the
authoritative server and standalone simulator through one deterministic Go engine.

## Public surface

These responsibilities are implemented in `state.go`, `transition.go`,
`projection.go`, `lifecycle*.go`, `ledger.go` and `settlement.go`. The table describes
boundaries; source declarations remain authoritative for exact Go signatures.

| Boundary | Kind | Purpose |
|---|---|---|
| Match/game state | Domain data | Physical card identities and zones, players, phases, quotas, pending decisions, proposals and exact financial records. |
| Rule configuration | Validated input | Pinned effective rules, interpretation and compatibility identity; validation/hash format is defined in S03. |
| Transition | Operation | Validate a command against state and rules, then produce new state, ordered domain events and explicit continuation requirements. |
| Seat observation | Query | Expose only that seat's authorized state, public counts and decision options; never another seat's private cards or random state. |
| Randomness | Explicit input/state | Reproducible versioned randomness with committed outcomes retained across continuation and replay. |

## Internals

A transition processes a declaration, response, pass or effect input without
mutating its input state. Pending action, response cursor, decision ID/stage,
required actor and random progress make interrupted effects resumable. Illegal
commands produce a domain rejection without partial changes. The online match
adapter owns authentication, membership, duplicate-command lookup, version/game
admission and the atomic database commit; the engine owns gameplay legality.
Human decisions never keep a database transaction open. The simulator supplies
the same engine with pinned configuration, seeded randomness and independently
validated bot commands. See [state/delivery](../design/DESIGN-go-state-and-delivery.md)
and the [game API](API-game.md) for adapter responsibilities.

## Invariants

- Conserve all 104 distinct cards across zones; preserve control, bindings,
  persistent-Ace custody and per-card availability through transfers and resets.
- Identical state, command, rules and random inputs under the same engine version
  yield identical outputs; no wall clock, global randomness, SQL or network I/O.
- Keep exact score/debt arithmetic, receipt order and provenance; spendable
  game-scoped points remain distinct from match score and product currency.
- Respect fixed response order and explicit decisions. Disconnection never
  implies a pass; Coup interrupts only at permitted atomic boundaries.
- Trade/loan proposals reserve nothing; only successful transfer creates an
  alliance or loan-derived privilege. Compensation follows its fixed-seat queue;
  purchases resolve the chosen quantity and eligible payment atomically (T01–T03).
- Seat views, events and bot inputs obey the same hidden-information policy.

## Tests

The [rule-to-test crosswalk](../../sims/docs/gameplay-coverage.md) maps accepted
Q/F/N/T decisions and the complete §2.12 schedule to concrete fixtures. Tests cover
physical conservation/aliasing, strict configuration, immutable transitions,
authorized observations, exact independent settlement-oracle agreement, response
and chance continuations, finalization and persistent multi-game state. Small
curated goldens live in `backend/testdata/`.

From repository root, execute `pwd` then
`xops/agent/safe-run.sh engine-tests -- bash -c 'cd backend && go test ./internal/game ./internal/bots'`.
Race/fuzz, CLI replay and actual autonomous campaigns are distinct evidence
categories; see [the simulation contract](../../sims/README.md). Database
retry/recovery tests are in the adapter below; host offline parity has separate P01 evidence.

## Dependencies

Validated rule data, explicit random state and exact arithmetic within the
single `backend/go.mod`. No PostgreSQL, Redis, Flutter, transport or
telemetry dependency belongs in the engine. Server/simulator adapters and bots
consume it; bots receive seat observations rather than authoritative state.

## Known issues / future work

D01–D04 targeted domain gates pass. The Phase 1 durable adapter extends local
continuations and multi-game orchestration to PostgreSQL transactions/recovery.
Phase 2 provides local public online transport and P04 integrates adaptive Flutter
web. Native-device acceptance and production release remain open. Actual complete-game evidence and measured
policy limitations are kept separate from historical Phase 0 partial-prefix
checks. Broad statistical strength/balance claims require the preregistered
campaign targets and precision; P01 owns host offline-engine comparison/parity and a provisional recommendation;
R06N owns final native integration, device evidence and engine approval. P04 web
integration uses authorized projections.

## Durable adapter (Phase 1)

`backend/internal/matchstore` is the implemented internal authority boundary.
`New(pool, rulesHash)` selects the accepted rules identity; `Create` pins an
`Envelope` and immutable member-to-seat mapping. `Submit` receives an already
authenticated actor identity and a complete `Intent` (ID, game ID, expected
aggregate version and board/financial request). It does not authenticate accounts.
Engine physical IDs remain private application identifiers, not the Phase 2 wire API.

Each command locks one match row, checks membership, retrieves an existing
actor/ID receipt, then admits a fresh intent against its named game and version.
A same-body retry returns the original result. A changed body conflicts. Current
same-game victory claims revalidate eligibility despite unrelated version changes.
Player actions cannot call the separate trusted `Continue` scheduler capability.
Neither disconnection nor elapsed time creates a player decision.

Snapshot, normalized exact ledger/provenance, game identities, private journal,
all seat projections/events and receipt commit in one PostgreSQL transaction.
`Snapshot` returns the authorized projection and cursor from one statement. Every
committed revision writes one envelope for each seat, including unchanged views.
No raw forensic journal is available through that member query. `Restore` is an
internal administrative capability. `SeatProjection` is an internal observation;
`wire` replaces physical IDs with seat/game/view capabilities; `transport` and
`delivery` authenticate snapshot, receipt and cursor subscriptions.

`ErrOutcomeUnknown` means retry/reconcile the identical identity on the primary.
`Reconcile` waits for the match lock; absence remains unknown, never a draw or a
license to mint a replacement ID. Server work has its own receipt namespace.
Committed random permutations come from `crypto/rand` and are stored with the
transition; offline experiment streams are not used for online entropy.

Compatibility pins cover envelope schema, adapter engine version, rules identity
and fixed match horizon. Unsupported checkpoints cannot execute. Existing receipts
remain readable without interpreting an unsupported checkpoint. `Up` verifies an
embedded migration checksum; `Down` is an explicitly destructive administrative
rollback and is never called by normal startup.

The journal retains the immutable initial checkpoint, accepted typed intents,
committed chance outcomes and complete private checkpoints. `VerifyJournal`
re-executes the stored trajectory without sampling new randomness and rejects
missing, changed or unused outcomes. This verifies recorded-outcome replay,
not the statistical quality of the entropy source or cryptographic tamper-proofing.
For wire commands, the receipt binds the original canonical wire body; the private
journal replays its separately stored decoded engine request. The trusted locked
codec admission step is not independently re-executed by that journal verifier.

The journal currently retains complete private checkpoints, and current financial
indices are rebuilt within each transaction. This favors inspectable correctness;
production storage sizing, compaction, statement timeouts and operational backup/
restore remain release work. Callers must supply bounded contexts and a correctly
configured primary pool. No deployed-service availability claim follows.

Tests: `store_integration_test.go` covers atomic faults, concurrency, historical
receipts, version refusal and PostgreSQL WAL crash recovery;
`network_integration_test.go` severs the actual COMMIT protocol before/after commit;
`recovery_integration_test.go` compares thirteen restored pending-stage successors,
including exact settlement, and checks duplicate/stale decisions;
`privacy_integration_test.go` varies hidden cards and nonparty financial terms;
`lifecycle_integration_test.go` covers declaration, proposal, purchase and match
boundaries. These complement the pure-engine rule/replay/continuation suite.

## Online adapter (Phase 2)

`identity` owns opaque sessions and in-place guest upgrades; `rooms` atomically
pins members and immutable match length using `CreateTx`. `SubmitChecked` decodes
seat-authorized card handles under the match lock after duplicate receipt lookup.
`delivery` serializes admitted commands with current active-player/round-order
priority; known victory claims skip ordinary rank queries at the next boundary.
An in-flight atomic transition or already-running bounded rank query completes
first. There is no victory batching timer or arrival-time fairness promise.
`transport.Start` advances automatic continuations but never human waits.
One matchstore transaction also writes finalized ordered-pair reputation outcomes.
Own cross-game corrections explain exact nonspendable adjustments; other players'
private corrections, current cash, debt terms and automatic workspaces stay hidden.
Board seats are one-based; projected financial debtor/payer indices retain the
engine's zero-based convention, declared by the protocol schema.

See [API-game](API-game.md) for retained-data/authentication decisions, actual
routes, schema, resource bounds, test scope and the single-process run contract.
