# Astra execution guide for Phase 0

**Preparation handoff, 2026-09-30. No simulator is implemented.** Start at
[the loading order](../README.md#phase-0-preparation-and-handoff). This guide maps
work to [S01–S10](../../docs/planning/ROADMAP.md#phase-0--cli-simulation); it is not
a second roadmap, backlog or completion ledger. All ten items remain unchecked.
The task that produced this guide ends with preparation; a subsequent authorized
implementation task begins at S01.

## Starting conditions

Run the real repository bootstrap, inspect recovery and staged/unstaged work, read
AGENTS and the full canonical inputs. Inventory at preparation: no backend module,
engine, bots, simulator CLI, container, runtime schema, simulator result or benchmark.
The Flutter demo under `client/` exists and is outside this scope; it is
not a rules oracle. `config/rules/` now contains only a specification, and the
preparation schema/examples cannot be passed as runtime configurations.

[Research decisions](phase-0-research.md) separate evidence from proposed engineering;
[operations](phase-0-operations-research.md) verifies Go 1.27.1 as the dated S01
candidate, not an installed compiler. Recheck support/security and actual local
version before pinning; do not silently download via toolchain selection during
offline checks. Follow AGENTS for system installations. Architecture/ADR approval
status stays proposed; implement the user-authorized Phase 0 topology without
claiming broader deployment approval.

## Intended ownership and dependency map

All runtime paths here are future paths; discover before creating them.

| Boundary | Intended paths | Responsibility |
|---|---|---|
| One module/build | `backend/go.mod`; optional `backend/go.sum` when dependencies exist | Exact toolchain provenance and module graph; no separate sims module. |
| Pure state/rules | `backend/internal/game/`; `backend/testdata/` | Transition legality, observations, explicit chance, exact amounts and shared fixture corpus. |
| Policy | `backend/internal/bots/` | Seat-only legal-random and initial heuristic; no second adjudicator. |
| Runtime adapter | `backend/internal/sim/`; `backend/cmd/sim/` | CLI parsing, schedules, budgets, workers, checkpoint/artifact I/O and reports. |
| Canonical data | `config/rules/game-rules.json`; `config/schemas/` | Single defaults source and strict versioned runtime schemas. |
| Reviewed inputs | `sims/scenarios/`; `sims/experiments/` | Synthetic vectors and named experiments, versioned independently of generated runs. |
| Generated evidence | `sims/artifacts/<experiment-id>/<run-id>/` | Ignored restricted outputs and approved projected reports. |
| Operations | `xops/makefile/`; `xops/README.md`; planned `backend/Dockerfile.sim` | Thin dispatchers only after checks exist; simulator-only packaging, no app stack. |

S01 establishes tests; S02 builds state; S03 validates/hash-pins config; S04 builds
transition/observation/random contracts; S05/S06/S07 implement their listed scenarios;
S08 consumes those boundaries; S09 operates them; S10 packages the real CLI.
S04 may define a decision/draw queue using synthetic inputs before S06/S07 supply
full scenarios. Keep the dependent gate open until its required behavior is actually
proved. Do not mark S04 complete on type declarations alone.

## S01 — module, toolchain and first red/green evidence

Dependencies: authoritative docs and existing xops only. Intended paths:
`backend/go.mod`, `backend/internal/game/`, `backend/internal/game/*_test.go`.
Rule basis: §§1.1, 5; module purity from MODULE-game/ADR-0001.

Create a behavior test (for example double-deck physical identity/conservation)
that compiles and fails for a missing behavior assertion, then implement the
minimum to pass. A missing Go executable or syntax error is not the required red
behavior evidence. Record exact compiler/platform and real module-root test
invocation; choose supported Go from the dated research and pin provisioned compiler
checksums in build evidence. Prefer standard library, with any added dependency
justified and version/license verified. No invented module import domain: choose a
repository-appropriate module path from the actual remote/project conventions.

Gate evidence: failing assertion log and matching green run; actual documented test
command; pure-engine import boundary; pinned supported toolchain. Validate module
location, no duplicate modules and no filesystem/network/global RNG dependency.

## S02 — physical state and immutability

Dependencies: S01. Intended paths: `backend/internal/game/` card/state/zone/value
files and tests; `backend/testdata/cards/`. Rules §§0, 1.1, 1.6, 2.11–2.14, 3.4, 5;
Q04/Q06, F06/F08, N02/N07. Load rule-conformance skill.

Model 104 unique physical cards, two copies of each rank/suit; each belongs to one
zone. Represent hand, concealed Aces, exposed numbers/attachments/formations,
active-Ace effects, draw/discard and temporary initiative/reservation state without
duplicating cards. Reservations reference cards in their existing zone. Separate
control/custody, opening history, per-club use, binding state and availability.
Exercise acquisition, intact loan and surviving-card return movement **primitives**
against §2.13; full T01 proposal/action lifecycle remains Phase 1. Likewise Fate
movement/binding preservation can be tested through the state primitive without
claiming complete Fate activation.

Gate evidence: exact card partition after acquisition/loan/return, persistent
active/dormant-together/dormant-separated bindings, no new third binding, public
active-Ace custody distinct from playable inventory, restrictions surviving forced
movement/Fate/shuffle/redraw. Property tests generate valid partitions and reject
duplicates/missing cards. Mutation tests on successor state, observation and exact
amount internals cannot alter predecessor snapshots. Retained history is not current
support; restricted physical markers cannot leak hidden locations.

## S03 — canonical configuration, validation and normalization

Dependencies: S01/S02. Intended paths: `config/rules/game-rules.json`,
`config/schemas/rules.schema.json`, `backend/internal/game/` validated rule values,
`backend/internal/sim/` loader, `backend/testdata/config/`. Rules §§1.1, 2.12, 2.16,
3.2–3.5, 4; fixed match count and Q07/F09. Load format spec and rules README.

Extract the sole accepted defaults with source clauses/units/bounds; distinguish
Code's mutable in-game settings and experiment budgets. Add strict parser/schema
and semantic validator, explicit version compatibility, default/override resolution,
normalized canonical bytes and SHA-256. Separate rules hash from experiment hash.
No preparation status/version is runtime-compatible. Test unknown nested keys,
duplicate/case-varied keys, malformed UTF-8, trailing JSON, bad numeric/rational
forms, invalid player/count ranges, unsupported engine/rules and illegal strict
profile overrides. Include equivalent ordering/whitespace/default materialization
hash vectors and intentionally changed-value hash cases.

Gate evidence: invalid configs rejected before game execution; equivalent effective
configs hash equally; canonical bytes retained; bounds/units/version policy visible;
fixed positive match count defaults to three while run/batch explicitly select one.
The preparation vectors are seeds for this suite, not evidence it passes in Go.

## S04 — transitions, observations, decisions and random continuity

Dependencies: S02/S03. Intended paths: `backend/internal/game/` transition,
projection, decision/draw queue and random boundary types; `backend/internal/sim/`
seed/replay adapter; `backend/testdata/replay/` and `observations/`.
Rules §§2.9, 2.12–2.14, 5; F01/F10, F06/F08, N02/N07, T02.

Build pure accepted/rejected transitions, explicit decision kind/actor/stage/IDs,
committed random outcomes and continuation. Pin seed derivation, independent streams,
PRNG words, rejection sampler, shuffle and canonical state/event serialization.
Observation tests include public hand/Ace counts, concealed face/eligible-count
noninterference, public identity history without tracking hidden shuffle, remote
binding/restriction privacy and authorized active-Ace zones. Legal menus/errors
must obey the same visibility boundary.

F01 fixtures exercise staged Negotiator/Justice/Dexter **boundaries** (required
actors, valid choices, stale/wrong/duplicate IDs, durable selected result), not merely
stub records. Integrate the narrow effect paths needed to prove continuation; do
not claim every ability variant implemented. T02 fixtures start from an explicitly
valid post-action queue: seat 2/3 reach five losses, returned queen is sole supply,
seat 2 draws, seat 3 consumes entitlement without a card. Reverse initiative,
multiple draws, recycling and Coup at each legal atomic wait preserve completed
draws/counters. Point settlement follows the queue; ordinary victory waits.

Gate evidence: identical fixed inputs produce exact state/event hashes; continuation
retains stages and chosen random outcomes; per-seat projections obey §2.13; pinned
engine/random provenance private. Checkpoint round-trip of queue/counters/draws
preserves fixed-seat order, exhaustion and legal Coup boundaries. Database restart
is D08, but file/state continuation is required here. A digest match does not excuse
an incorrect expected rule trace.

## S05 — setup, initiative and openings

Dependencies: S02–S04. Intended paths: `backend/internal/game/` setup/opening;
`backend/testdata/setup/`; promoted `sims/scenarios/` fixtures. Rules §§0, 1.4–1.7,
2.13; Q06/Q07. Use replay and rule-conformance skills.

Deal twenty each and redeal **all** hands on any diamondless hand; separate Aces
and expose initial Diamonds. Seed vectors force zero/one/multiple redeals. Keep
all initiative draw-off cards aside (including non-numbers), resolve ties, then
return/shuffle; exhausted unresolved ties use rotating fixed-seat priority skipping
departures without disturbing resolved positions. Freeze order until next round.

Gate evidence: returned cards shuffle into draw supply, discard recycling only when
empty, partial mandatory draws create no entitlement; initial deal guarantee never
applies mid-game. One new series or combo opening allowance; opening second suit
permanently ends protection, loss of current support prevents attacking without
restoring protection; reopening historical empty series is addition without another
allowance. Unsupported combo activation remains explicitly outside coverage even
when the opening-allowance state supports its future representation.

## S06 — narrow combat, responses and paid-cost continuity

Dependencies: S04/S05. Intended paths: `backend/internal/game/` combat/response;
`backend/testdata/combat/`; promoted `sims/scenarios/`. Rules §§1.7, 2.5–2.10,
2.12–2.14, 3.4; Q01/Q02, F01/F02/F09, N01/N03.

Cover ordinary exact matching and target suit order, initial protection/two-current-
series prerequisites, physical unused Clubs, frozen targets and non-nested
fixed-seat responses. Enumerate the supported abilities and cancellation matrix.
Illegal declaration changes nothing; legal commitment marks Clubs used; later
failure keeps usage and previously paid costs but omits success-only costs.
Numerical/Advancement defense changes effective comparison without virtual payment
cards; N01 Club 5 + Advancement 10 against Barricaded Hearts 7+8 returns Club 5 and
both targets, not fabricated Clubs. Include exact half Spades under Inflation.

Confinement cancels the pending actor's action, ends only an active actor's turn,
expires before their next scheduled draw, and cannot prevent confined Coup. Skip
confined/departed response seats; keep eligible disconnected waits. F01 selected
response stages continue without duplicated choices/costs. F09 includes duplicate
Inflation rejection, exact odd halves and cancellation/qualifying transaction order
where covered; comprehensive ability/purchase variants remain D03.

Gate evidence: table-driven legal/illegal declaration, pass/response order, exact
comparison/physical exchange and continuation traces; invalid action no mutation;
paid-versus-resolution cost assertions. Every S06 named gate must pass; deferring
all Confinement/decision semantics would not constitute a narrow implementation.

## S07 — exact receipt, score/debt and resumable settlement

Dependencies: S02/S04, with S06 continuation inputs where needed. Intended paths:
`backend/internal/game/` amounts/ledger/reference/batching; `backend/testdata/ledger/`;
`backend/internal/sim/` checkpoint adapter. Rules §§2.4, 4.1–4.4, 5; Q05/Q09/Q10,
F04/F09, N06. Load conformance and replay skills plus research proof obligations.

Reproduce each worked §4.2 ledger: Coup −50 then next-game +50, charge 10 then receipt
6 then 8, older system 5 before player 10 with receipt 8, remaining 4 carried once
through Coup. Exact paired Ponzi shares of 11 are 11/2 each; generic reciprocal
1/3 cycle is not a three-party alliance. F04 cash 20 clears before new Coup debt 50;
zero next-game cash, departed exemption, ordinary finalization and nullification
match field-level expectations. N06 prior finalized −10 + later correction 10 +
Coup 50 totals 50, without forgiveness cash; same-game correction is replaced with
its charge; creating-game nullification restores prior adjustment/debt state.

Start with an unbatched FIFO oracle and frozen finite debt termination argument.
Compare candidate batching against expanded exact ordered traces, balances, scores,
IDs and remaining obligations. Include three-seat cycles/multiple debt heads/system
sinks. Freeze progress format and prove cancellation/restart equivalence at every
segment boundary; large-D test demonstrates retained useful progress without
rerunning the settled prefix. Measure operations/bit lengths/audit size. General
batching is optional only for shapes without a proof; exact resumable processing
and proven equivalent batch cases are not optional S07 gates.

Gate evidence: all named ledger fixtures, no duplicate charge, canonical rationals,
small-ledger differential/property/fuzz tests, compressed-audit expansion and
large-D resume. Full player-facing promise/F05 or T01 workflows remain later, but
scenario-level financial boundary transitions cannot be omitted. No financial
iteration cutoff, reciprocal netting, rounding or invented draw.

## S08 — fair baselines and named profiles

Dependencies: S04–S07. Intended paths: `backend/internal/bots/`,
`backend/testdata/bots/`, promoted `sims/experiments/`. Rules §2.13/§5, F10 with
F08/N07 hidden-state regressions; [experiment research](phase-0-experiment-research.md).

Provide legal-random over a canonical finite supported menu, transparent heuristic
with versioned feature weights/ties/reasons, independent engine validation and
per-seat RNG. Bots cannot inspect state or hidden legality side channels. Hidden-
state permutations preserving authorized history/counts must preserve menu, choice
and reason under the same policy stream. Adversarial bots forge cards, actor/stage
and game IDs to exercise rejection. Explicitly handle pass and supported financial
finish decisions; unsupported mandatory decisions stop, never auto-finish.

Gate evidence: fair observations, mutation isolation, independent illegal-action
rejection, reproducible policy vectors and named profiles with coverage/experimental
labels. No LLM/search or proven strength requirement. Preparation 3p/4p examples
are developmental plans, not accepted runtime inputs or measured results.

## S09 — full CLI, scheduling, replay and reports

Dependencies: S03–S08. Intended paths: `backend/cmd/sim/`,
`backend/internal/sim/`, runtime schemas under `config/schemas/`,
`backend/testdata/cli/`, `sims/artifacts/` (ignored), xops wrappers only when real.
Rule basis: §§1.1, 2.9, 4.2–4.4; F01/F04/F10, N05/N06, T02; full CLI README contract.

Implement **run, replay, batch, tournament, sweep, compare** and each required flag,
precedence rule, exit code and output. Replay consumes actual decisions/random
outcomes, verifies artifacts/versions and stops at first differing state/event hash.
Tournaments serialize games under one match owner and preserve cash/debt lifecycle;
no worker owns just one constituent game. Expand schedule before dispatch; compare
workers 1/2/larger for semantic equality. Sweep bounded grids, pairing fixed seeds
and rotations; compare rejects incompatible/unmatched inputs or clearly labels
explicit exploratory contrasts instead of pooling them.

CLI tests cover missing/invalid/unknown/duplicate fields, current-directory path
resolution, overrides, incompatible schemas, output-directory collision, malformed
or tampered replay, all nonzero exits and mixed child failures. Prove cancellation
while dispatching, blocked on result/write, choosing bots and settling; preserve
completed jobs, not-started denominators and exact resumable cursor. Relevant race
checks exercise shared configuration, workers, writer and shutdown. Fault tests
interrupt before/after artifact and manifest publication.

Reports include every field in README: purpose, summary, provenance/reproduction,
interpretation/coverage limits, observer narrative/reasons, separate exact ledgers,
ending reason/metrics/hypotheses and per-game links. Separate three/four-player
planned/started/finalized/stopped counts and complete seat rotations; no fabricated
zero/draw scores. Statistical unit is paired seed block, retaining whole rotations
and within-match dependence; test known paired means, bootstrap indices, empty
and selectively incomplete inputs. Complete-case estimates disclose selection bias.
Public Markdown never contains reconstructive seeds/private transcripts.
Implement and test the privacy-aware xops entry point described below before
running private CLI workloads through safe-run.

Gate evidence: printed commands reproduce actual fixtures, all six commands tested,
worker-count invariance, paired schedules/seat rotation and population-specific
incomplete denominators in reports, bounded cancellation, all outputs/hashes verified.
A useful report may compare supported scenario quantities/completion behavior; it
must not present a completed scenario as an independently playable full game.

## S10 — isolated packaging and operational evidence

Dependencies: S09. Intended paths: `backend/Dockerfile.sim`, existing xops dispatch
location, `sims/README.md` actual invocation sections, artifact/build evidence.
Gameplay unchanged; reproducibility/privacy/retention contract applies in full.

Package the existing CLI in a pinned simulator-only image with verified toolchain,
module graph and source provenance. Host and container need no PostgreSQL, Redis,
Flutter, application telemetry or network at execution. Provisioning is a separate
network-allowed step with explicit hashes; offline build uses preprovisioned inputs.
Run non-root, read-only input/root, designated output mount and explicit resource
limits. No unseen latest tags or untested executable examples. Preserve caller path
semantics and exact reproduction metadata across host/container path mappings.

Gate evidence: actual host/container commands and isolated successful scenario +
batch/comparison; runtime schema validation for manifest/outcome/trace/checkpoint;
private-versus-public export regression; retention preview protecting held evidence
and curated files; artifact-path escape/I/O/cancellation tests; two clean build
hashes and host/container semantic replay comparison with recorded limitations.
Claim platform atomicity/durability only on tested filesystem/runtime combinations.
Publish exact build flags/platform/dependency/toolchain/binary/image hashes and
performance baseline methodology; no guessed release performance limit.

## Coverage boundaries and unchanged broader targets

| Accepted family | Phase 0 evidence required | Remaining full coverage (unchanged roadmap) |
|---|---|---|
| Q01/Q02, F01/F02/F09, N01/N03 | S04/S06 supported decision, cost, combat and response fixtures | D01–D03/D08 all ability variants, persistence and adapter recovery |
| Q03/Q04/Q06/Q07, F06/F08, N02/N07 | S02 state/movement invariants; S04 views/draw boundaries; S05 setup/openings | D02/D03 complete victory/protection/reset/ability/purchase combinations |
| Q05/Q09/Q10, F04, N06 | S07 paired receipt and exact boundary/lineage scenarios | D04/D09 complete alliances/departures/settlement lifecycle integration |
| Q08 | S04 narrow F01 staged boundaries and S06 applicable continuation | D03 maximal Negotiator and all other card-operation variants |
| Q11/F03/N05 | Stable game/decision IDs and local continuation/replay | D06–D08/O04 authenticated duplicate ordering/current-state admission and real races |
| Q12/F05 | Format preserves named awards, phases and explicit decisions; supported financial paths stop honestly | D04/D09/O08 full promises/reputation and post-board workflows |
| F07/N04 | State representation retains lasting effects/economic identity when present; unsupported activations explicit | D03/D04 full Code/Underground lifecycle/economics |
| T01 | S02 loan/return movement and schema/provenance preservation | D03/D05–D08/O08 full proposal/acceptance lifecycle and concurrent races |
| T02 | S04 exact ordered deferred draw queues/exhaustion/continuation/Coup fixtures | D02/D03/D08/O05 full originating abilities and online recovery |
| T03 | Preserve explicit quantity/payment/supply in contracts; no silent resize | D02/D03/O03 full purchase implementation; S06 F09 fixtures may exercise minimal relevant pricing boundary |

The broad README targets remain **all approved Q/F/N/T fixtures**, **1,000 seeded
replays per supported player count**, and **at least 1,000 held-out paired seed
blocks per player count** with rotated seats and a favorable preregistered score
interval for stronger-play claims. Phase 0's narrow scope does not erase these or
make unimplemented scenarios pass. Maintain a coverage report listing each clause/
fixture as implemented-tested, unsupported, or experimental, with evidence links.
At Phase 0 exit state which broad targets were actually attempted/met, which remain
unfulfilled, and why; never label partial-rule results full conformance or strength.
Do not import D/O/P gameplay just to hide unsupported paths. The required Phase 0
exit is a reproducible accepted-rule scenario and useful honest batch/comparison
report, plus all S01–S10 gates—not complete independent playability.

## Commands and evidence at handoff

Available now, from the repository root:

```bash
pwd
xops/agent/session-bootstrap.sh
```

For future implementation, wrap each **verified real** build/test/CLI invocation
with the existing `xops/agent/safe-run.sh` contract; its tag is followed by `--` and
the actual command. Do not run fictitious Go targets or the illustrative README
CLI until implemented. Before every build/test/git invocation, confirm `pwd`.
Record exact cwd, tool versions, source/config hashes, exit and durable log.
For tests/builds, normal safe-run logging is appropriate. For private simulator
invocations, S09 must first add a privacy-aware xops entry point: safe-run receives
only an opaque run ID and protected run-spec path; the inner helper loads private
argv, captures/redacts subprocess output and retains the full command solely in
restricted artifacts. The current wrapper records argv in console, `.cmd` and
`last_failure.json`, so do not pass private seeds directly to it. Changing
`AVB_RUN_DIR` alone is insufficient. Verify synthetic seed/card canaries never
reach console, wrapper logs or tracking failure state on either success or failure.
No such helper exists in this preparation; its implementation and negative tests
are part of S09 before operating private runs.
Read a nonzero log, diagnose/fix, resume, then mark recovery resolved.

Preparation validation checks schema/meta-schema, examples and semantic schedules,
exact oracle arithmetic, encoding/hash vectors, local links, skill metadata,
unchanged CLI/roadmap/rules and complete diff hygiene. It cannot establish Go
conformance, optimizer equivalence, replay integrity, fair policies, cancellation,
races, measured performance, container isolation or reproducible binaries. Those
require the real tests above. Review/verification precede the coordinating parent's
tracking row and staging; humans commit/push. For this preparation's evidence see
[validation record](phase-0-validation.md).
