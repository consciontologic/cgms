# CGMS simulation contract

Status: the shared engine and CLI now support accepted-rule gameplay, financial
finalization and subsequent games. D01–D04 targeted domain gates are recorded in
[the roadmap](../docs/planning/ROADMAP.md); persistence/server gates remain open.
Normal-deal development games have completed and replayed for both player counts,
including completed three-game matches. This is evidence of exercised behavior, not formal
proof of every interaction or established bot strength/balance.

Read [rule coverage](docs/gameplay-coverage.md), [bot specifications](docs/gameplay-bots.md),
[measured findings](docs/gameplay-findings.md), the [historical frozen evaluation plan](experiments/gameplay-heldout-v2.md), and the [v3 development and prospective confirmation plan](experiments/gameplay-evidence-v3.md). Historical
[Phase 0 evidence](docs/phase-0-implementation-evidence.md) remains bounded:
2,000 partial-prefix replay checks and zero completed games in that phase.
V3 completed 54/72 primary development games (three independent blocks each),
plus separate diagnostics. Review found v1 action-key collisions that confound
policy attribution; the versioned v2 repair has separate regression evidence.
Confirmation remains uncollected pending
the explicit measured resource allocation; P03 remains open.
The [shared-engine ADR](../docs/design/ADR-0001-shared-go-simulation-engine.md) and
[architecture](../docs/code/ARCHITECTURE.md) define component boundaries.

## Purpose and ownership

Phase 0 must produce a useful local CLI before the application foundation. It needs
only Go, versioned inputs and local files; PostgreSQL, Redis, Flutter, network access and LLMs are unnecessary.
The implemented `backend/cmd/sim` and future `backend/cmd/server` share `backend/internal/game`
within one `backend/go.mod`; `backend/internal/bots` contains bot policies.
Neither simulator nor bots reimplement adjudication. A bounded worker owns a
whole match and serializes its games, preserving cumulative scores, debts and
carried adjustments. Parallelize independent matches only. Standalone `run`
and ordinary `batch` entries start fresh one-game matches; tournaments retain
the recorded match state between their constituent games. Game count is visible
and fixed before match start, defaulting to three unless explicitly configured;
`run` and ordinary `batch` explicitly configure one-game research matches.
Financial finalization precedes every next game, with zero new-game available
points and surviving match scores/debts. Tied cumulative scores share rank;
record declaration outcome, game score leader and match rank separately.
Assign a stable immutable game-instance ID to every recorded game, including a
new ID for each nullified replacement at the same match slot. Commands and
decision records name that instance; replay never retargets an old intent.

| Path | Contents and source-control policy |
|---|---|
| `sims/scenarios/` | Small versioned JSON initial positions, legal action sequences and expected outcomes; synthetic hidden cards only. |
| `sims/experiments/` | Versioned JSON research plans: purpose, seed sets, lineups, seat schedule, parameter grid and interpretation profile. |
| `sims/docs/` | Optional investigation notes linking immutable run manifests; no second roadmap. |
| `sims/artifacts/<experiment-id>/<run-id>/` | Generated reports, traces and input snapshots; ignored by Git. |

Runtime code stays under `backend/`, rules under `config/rules/`, schemas under
`config/schemas/`, and operational helpers under `xops/`. Experiment grids reference
allowed parameters, not competing defaults. Curated scenarios and small evidence
summaries may be reviewed into Git; generated output must never be auto-staged.

## CLI surface

The binary is `cgms-sim` (the build helper names its output `sim`). Paths resolve
from the caller's directory. Versioned JSON definitions reject unknown fields and validate before games start.

| Subcommand | Required inputs and behavior |
|---|---|
| `run` | `--config`, `--seed`, `--players` (3 or 4), `--bots` (one versioned bot per seat), `--out`; optionally `--scenario`; produce one game. |
| `batch` | Same game inputs plus `--games` and `--workers`; derive an indexed seed schedule from the root seed, record every derived seed and preserve result ordering independently of workers. |
| `tournament` | `--experiment`, `--config`, `--seed`, `--out`; validate lineups, player count, complete seat rotations and games per match; report cumulative match scores separately from game declarations. |
| `sweep` | `--experiment`, `--config`, `--seed`, `--out`; enumerate a bounded, validated parameter grid with paired seed/seat blocks; retain each effective configuration and hash. |
| `replay` | `--manifest`, `--out`; verify artifact hashes and compatibility, replay recorded commands/random outcomes, compare every state/event digest, and stop at the first divergence. |
| `compare` | `--baseline`, `--candidate`, `--out`; compare compatible manifests and paired seed/seat blocks; expose changed rules, bots and interpretations rather than silently combining them. |

All game-producing commands accept `--interpretations`; omission selects the
accepted GAME_RULES profile without experimental replacements. Flags override
corresponding experiment inputs, validate,
and appear in the manifest. `--workers` changes scheduling only. Work limits are
experiment safeguards: a limit produces `budget-exhausted`, never a fabricated draw.

Equivalent invocation from the repository root (execute seeded work through the protected helper below):

```text
cgms-sim run --config config/rules/game-rules.json --seed 42 --players 3 --bots legal-random@v1,legal-random@v1,heuristic@v1 --out sims/artifacts/baseline/run-001
```

Return `0` only when the run completes and artifacts verify; nonzero codes are
`2` invalid input or incompatible version, `3` unsupported transition or experimental
adjudication required, `4` replay/invariant failure, `5` budget exhausted, `6` I/O
failure and `130` cancellation. Collect batch diagnostics but
return nonzero if any game is incomplete. Write an atomic manifest last with
`complete` or explicit partial/failure status; preserve completed games after
interruption. Reject existing output directories and print the manifest path.

## Running the implemented scope

Go **1.27.1**, Python 3 standard library and Linux are the tested host baseline.
The Go module has no third-party dependency. Build from the repository root:

```sh
pwd
xops/agent/safe-run.sh sim-build -- python3 xops/sim_build.py phase0-local
```

Use a fresh build ID; the helper refuses existing directories, builds twice with
fresh caches in different paths, and records compiler/source/binary provenance in
ignored `sims/artifacts/build/phase0-local/`. The binary is
`sims/artifacts/build/phase0-local/sim`. `GOTOOLCHAIN=local`, `GOPROXY=off`,
`GOSUMDB=off`, `CGO_ENABLED=0`, `-trimpath` and `-buildvcs=false` are enforced.
A supplied Go installation is required; no toolchain is downloaded.

Run the historical bounded-scenario contract evaluation (fresh output ID required;
this does not count completed games):

```sh
pwd
xops/agent/safe-run.sh sim-evaluate -- python3 xops/sim_evaluate.py sims/artifacts/build/phase0-local/sim phase0-local-evaluation --samples 1000
```

The helper creates mode-0600 run specifications and invokes `sim_private.py`.
Only opaque paths and redacted status reach safe-run logs. It exercises the
accepted Q10 thirds fixture, fair bot scenarios for both populations, batches at
one/four workers, replay, named tournaments, sweeps and comparisons. Expected
partial code 3 is checked, not suppressed. Generated manifests/reports and a
compact evaluation summary stay under the ignored output directory.

For an individual run, create a protected JSON specification beneath
`sims/artifacts/` with exactly `argv` (an array), `cwd` (an absolute directory), and
`capture` (an unused path beneath that same artifacts root); chmod it 0600.
Put the complete invocation from the table in `argv`, including the binary path.
Then run `xops/agent/safe-run.sh sim-private -- python3 xops/sim_private.py
sims/artifacts/REQUEST.json`. The helper preserves private child output and argv;
it refuses traversal, symlinks, duplicate keys and public-readable specifications.
Do not put seeds, hidden cards or direct simulator arguments in safe-run commands.

Curated runtime inputs are `q10-reciprocal-thirds.json`,
`narrow-combat-bots.json`, and `narrow-combat-bots-4p.json`. Files ending
`.preparation.json` are historical preparation inputs, not CLI scenarios.
The bot fixtures stop after accepted combat at the required post-action shuffle;
they do not infer a pass, turn end or victory. Normal generated deals use the complete gameplay adapter. Tournament later slots
remain not-started when a prior game is incomplete; successful finalization carries
match debt/corrections into the next fresh game identity with zero new cash. No
independent new game resets match debt. A supported engine action missing adapter
or policy support remains an explicit stop and a defect to investigate.

Standalone gameplay uses 100,000 policy operations per decision; menu generation
has its own explicit 100,000-operation cap. Either can stop a valid unfinished
game; larger card sets are not silently sampled down to avoid the stop.

`--max-steps` sets the deterministic transition budget (and standalone settlement
budget). Experiment settlement, bot-operation, input-digit/byte, output and wall
limits remain explicit. Equal output shares are reserved per logical match before
dispatch, independent of worker count; insufficient mandatory checkpoint capacity
is rejected before starting. Canonical JSON values are bounded at 4 MiB; replay
reads regular artifacts only, at most 64 MiB each/256 MiB aggregate. A budget stop
keeps exact committed obligations. Wall watchdogs affect stop status, not engine
rules. Metrics are adapter-only diagnostics; sampled Go heap is not peak RSS.

For settlement continuation, supply `kind: "settlement-checkpoint"`, the exact
checkpoint object, population, scenario identity/purpose/rule references and a
`parent` manifest path. The adapter verifies the parent transcript/configuration and exact checkpoint membership before starting, snapshots that provenance, and records the parent in the new manifest. Use a new output directory. Replay verifies the old prefix;
continuation resumes its retained queue, rather than recomputing prior payments.
Fixed transition scenarios retain scripted commands/random words as synthetic
fixture inputs; bot scenarios retain separate seeded policy-stream snapshots.
Ordinary replay never reruns a bot to decide its command.

For complete-game continuation use `kind: "gameplay-checkpoint"`, `gameplay`
containing the exact retained private gameplay transcript, and its `parent`
manifest path. Use `run --scenario` with a fresh output directory. Omitted seed,
population and policies inherit from the verified parent; explicit mismatches
reject. Logical match indices and original execution budgets are preserved.
`--max-steps` grants additional transitions, not a new game or policy decision.
Checkpoint inputs are private and must use the protected helper.

The complete-game campaign helper is `xops/sim_gameplay_evaluate.py BINARY PLAN
OUTPUT_ID [--resume | --analyze-only]`. Its strict private plan freezes both population profiles,
binary/build/config hashes, shard size, analysis seed/resamples, resource limits
and worker comparison units. It interleaves population shards, invokes the protected
CLI, verifies manifests/effective inputs, replays recorded games, and journals
immutable stage/analysis files. Resume retains failed units and continues later
planned units; it never overwrites or quietly retries a failure. Resource-limited
results retain all planned denominators and conditional uncertainty. Templates
under `sims/experiments/` use placeholder roots; the actual frozen seed inputs
remain mode-0600 files under ignored artifacts.

`--analyze-only` verifies saved frozen inputs, shard manifests and exact canonical
score objects, then writes separate immutable reanalysis files. It does not run
games, change campaign state, renew budgets or replace historical reports.


`cgms-engine-v2.2` is the current CLI; it corrects started-game counters at
pre-deal cancellation and caps optional restricted diagnostics without truncating
authoritative ledgers. The frozen v2.1 campaign retains its original binary and
inputs; these reporting repairs do not change game transitions or policies.
`cgms-engine-v2.1` added the complete-game adapter and authenticated empty-response
continuation. `cgms-gameplay-public-v2` omits private financial totals from ordinary
reports. The canonical accepted rules package remains unchanged (`cgms-engine-v1`
rules compatibility); runtime engine and legal-menu provenance separately identify
the implemented evaluator. Saved Phase 0 or development artifacts must be replayed
with their recorded compatible binary. A rules hash alone is not an engine-version
or policy-equivalence assertion. Earlier development public-v1 reports remain
restricted because review found private aggregate financial totals in them.

Runtime schemas are in `config/schemas/`; `python3 xops/sim_schema.py --check`
checks generated shapes against Go types. Go validators additionally enforce
canonical rationals, physical conservation and legal continuations. Manifests are
published last, after checksummed artifacts; writable I/O failures receive an
explicit failure manifest. Unpublishable storage failure returns 6 without claiming
publication. Raw artifacts are restricted; public reports omit seeds/hidden faces.
Use `--review-hold` to pin evidence at initial publication (the evaluation helper does this automatically). Retention previews protect held runs and referenced parents, never delete anything.

The scratch container recipe is `backend/Dockerfile.sim`; build using only the
prepared binary context:

```sh
pwd
xops/agent/safe-run.sh sim-image -- docker build --network none -f backend/Dockerfile.sim -t cgms-sim:phase0 sims/artifacts/build/phase0-local
```

For execution, put the `docker run` argv in a protected specification too. Use
`--network none --read-only --cap-drop ALL --security-opt no-new-privileges`, a
non-root `--user`, bounded `--pids-limit/--memory/--cpus`, read-only mounts for
`config/`, `docs/project/` and curated inputs, plus one writable output mount.
The container needs neither an application service nor a writable root filesystem.
The evidence record identifies the tested mounts/build and exact outcomes.

## Rules and reproducibility

The [current rules](../docs/project/GAME_RULES.md), accepted [change notes](../CHANGELOG.md)
and [resolved contract register](../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md) govern.
All Q01–Q12 and the adopted F01–F10/N01–N07/T01–T03 review decisions are accepted; a missing implementation is a coverage gap,
not unresolved gameplay. Reject unsupported transitions before execution, preserve
the last committed state and name the missing rule/implementation. Unknown or
incompatible rules/engine versions fail validation before starting a game.
Experiments may opt into named, versioned interpretations with explicit replacement
semantics; experimental runs cannot certify GAME_RULES. Keep `adjudication-required`
for a genuinely undefined future experimental case. If an earlier legal response
already resolved, preserve its effects, durable result and paused-action marker;
never repeat its cost or roll it back as a no-effect rejection. Record partial
coverage, experimental adjudications and exclusions; completion-only statistics
must show denominator and selection bias.

The transition core receives actions/random outcomes without clock, global randomness,
filesystem or database dependencies. Each of the 104 cards has a stable deck/suit/rank identity. Keep
historical openings separate from current series, per-club use separate from
card location, and response windows ordered. Persist effect/decision IDs, stages,
required actors, authorized options and committed random outcomes. Negotiator,
Justice and Dexter can require successive inputs without nested response windows;
the decision transcript is replay input and reconnect never rerolls a committed
result. Coup can interrupt a decision wait, including for a confined declarer,
but never an atomic mutation; ordinary declarations require the permitted
completed-action boundary and no Confinement. Represent exact score/debt amounts
as reduced arbitrary-precision rationals, serialized as decimal-string numerator
and positive denominator. Record debts in creation order and prevent duplicate
charges. Apply Q10: gross receipts in fixed seat order, creditor receipts FIFO,
oldest debt first, no automatic reciprocal cancellation and no new obligations
during settlement. Prove termination and optimized-batching equivalence; measure
work rather than inventing a cutoff that drops obligations. Reciprocal debts of
1 with a receipt of 1/D need 2D reference transfers: retain an unbatched oracle
for small cases, equivalent compressed batch audits and resumable exact progress
for large-D cases. A work-budget stop preserves progress and remains incomplete;
replay/resume must not repeatedly redo all prior expensive settlement. Exclusive alliances
permit only two equal Ponzi shares; third-fraction cases are generic arithmetic
fixtures. Debts persist within a match and reset for each separate new match.
Available points are game-scoped. Apply §4.4's full transition fields for Coup,
ordinary ending, departure, nullification, finalization and next-game start.
Bots must explicitly answer settlement offers/finish commands through the same
legal interface; silence cannot fabricate refusal or completion. Typed promises
name the origin game, award and condition. Reject gameplay after board closure
while permitting its specified financial decisions until finalization.
Record nonspendable cross-game forgiveness as a match adjustment linked to the
original retained charge/debt and its creating game, separate from immutable
completed results. Later Coup preserves that correction; nullifying its creating
game restores the prior debt/adjustment state. A current-game charge's correction
belongs to that current result and is replaced together with the charge by Coup.
Neither form creates cash or creditor receipts. Replay original charge lineage
and exact totals rather than modifying old score histories.

Implement acquisition/visibility from §2.13 and ability lifecycle from §2.12.
Public hand and separate concealed-Ace counts are intentional observations;
hidden faces and royal/eligible-card breakdowns remain private. Captured number
Diamonds are exposed even for Justice; other capture, trade and loan/return
destinations follow their named rows. Keep public identity history without
exposing another separated bound king's location. Lasting Underground privileges
and Code declarer/settings are player records, separately from current custody;
only allocated functioning Underground kings produce income. Split Doppelganger
bindings and nonstacking Inflation have explicit state and rule fixtures.
N01 combat compares effective Club total plus attack modifiers with frozen
target-card total plus applicable defense modifiers; separate those virtual
totals from physical Club return costs. A defensive numerical ace can affect any
own targeted number series, while Advancement defense is Hearts-only. N03
response traversal skips confined/departed seats, keeps the remaining order and
never skips an eligible disconnected player. N04 Code declaration changes all
opponents' settings as a non-attacking economic action; its hostile round penalty
requires the active, non-confined declarer's two series and excludes protected/confined opponents.

N02 active Inflation/Compensation Aces live in a public effect zone with source,
target, target-as-custodian and expiry, without gameplay reuse. Fate excludes them
from reset/draw counts and preserves Compensation's remainder. Caster departure
does not end another target's Inflation; target departure and board closure
discard affected active Aces under their cleanup rules. N07 `available_from_round`
belongs to physical Barricade draws and survives involuntary capture/theft,
Fate, shuffle and redraw. Before release, prohibit voluntary play/payment/transfer
and Coup but permit ordinary threshold counting/proof and mandatory Diamond
exposure. Filter hidden markers so observations cannot follow concealed cards.

T01 trade/loan proposals are nonblocking private records with offer ID, revision,
origin game/turn and status. They reserve no cards/allowances and expire on the
origin turn's end, board closure or party departure. Bots may withdraw/decline
unanswered offers through the same interface. Acceptance revalidates exact terms
at an idle boundary and starts the nonwithdrawable transfer action with the
acceptor as actor for responses/Confinement. Receiving an intact loan needs no
ability support. Final revalidation atomically transfers cards and grants any
alliance/loan-derived privilege only on success; cancellation retains prior
responses/committed Ace allowances and closes without transfer. Record proposal
and action states separately so a silent recipient does not block End turn.

T02 deferred Compensation draws use fixed ascending seat order after action
returns/shuffle. Complete each recipient's quantity, consuming earned five-loss
entitlements even when supply is exhausted, before the next recipient and before
point settlement or ordinary actions/victory. Record queue position, counters and
committed random outcomes. Coup may interrupt between atomic queue transitions,
preserving completed draws and abandoning the uncommitted remainder; simulation
must exercise the same boundaries as online recovery.

T03 purchases choose positive integer quantity and specific unrestricted number
Spades from hand/exposed series at an own-turn idle boundary. Effective payment
must cover quantity × current Code price under Inflation; excess gives no change.
No Underground/two-series prerequisite applies. Validate full supply including
returned payment, then use ordinary responses and final revalidation before
atomic payment/shuffle/exact-quantity draw and eligible Inflation removal.
Canceled/failed purchases pay/draw nothing and never silently resize quantity.
Cover the exhausted-supply case where one Spade 10 can buy a chosen one card,
but cannot supply a chosen two despite adequate value at price 5.

The [F01–F10/N01–N07/T01–T03 scenario pack](../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md#accepted-interaction-amendments-and-scenario-pack)
is shared across simulator, online command path and frontend reconnect journeys.
Each executable trace must name actor/inputs, expected zones/counters, exact ledgers, phase/
decision IDs, committed event order and all permitted per-seat observations.
Transport-only admission cases, such as same-game unrelated stale versions for
a legal Coup or delayed old-game first submissions, belong in server adapters
while reusing the same domain outcome. Original committed duplicates retain
their game/result even after the simulator progresses to another instance.

Each immutable manifest records:

- Rule-source revision and SHA-256, ruleset/schema/engine compatibility versions,
  and canonical effective configuration SHA-256; store the effective bytes too.
- Interpretation ID/version/content hash (or strict profile), scenario and
  experiment hashes, all overrides, run status, exclusions and reproduction command.
- Engine source revision, dirty patch hash if applicable, binary hash, Go toolchain,
  build flags/platform and dependency lock hash; bot versions/configuration hashes.
- Root and derived seeds; pinned PRNG algorithm/version, seed-derivation algorithm,
  shuffle/sampling algorithm versions, independent game/bot streams and draw counters.
- Ordered commands, response/effect choices and IDs/stages, committed random
  outcomes, immutable game identities and replacement lineage, financial-phase/
  settlement progress, proposal revisions/statuses, deferred-draw cursors,
  purchase quantities/payment identities, original charges/match adjustments, event/state schema versions,
  per-transition digests and final artifact checksums. Record a total order rather
  than relying on map iteration, goroutine scheduling or timestamps.

Phase 0 pins normalization, hash encoding, PRNG test vectors and state serialization
in golden fixtures. A seed alone is insufficient. Exact replay verifies recorded
decisions; re-running bots is a separate check. Incompatible versions fail explicitly.
Online games must keep future random state, seeds, deck order and private cards
server-private and use cryptographically secure, unbiased draws; a seeded
experimental PRNG is not automatically suitable for production. Full offline
forensic traces are restricted research artifacts;
ordinary reports use an explicit observer view, and omniscient exports are labeled.

## Bots, reports and proposed evaluation gates

The legal-random baseline samples a stably ordered legal-action set. The heuristic
baseline evaluates transparent features such as immediate score, exposed holdings,
protection, debt and declaration opportunities; reports record feature-based
reasons. Both receive only their seat's projected observation and public history.
Their legal-action menus must not reveal hidden identities or future draws.
Hidden-state permutation tests verify observational equivalence. An explicitly
named omniscient diagnostic mode is excluded from fair bot comparisons.

Later search/belief bots must beat baselines under equal information and budgets.
Optional LLM proposals need legality validation, non-LLM fallback, model/prompt/output
hashes and decision transcripts, with nondeterminism disclosed. Phase 0 needs no LLM.

Every run writes `report.md`, `manifest.json`, machine-readable outcomes and a
transition trace; batch/tournament reports link individual game reports. Markdown
includes purpose and executive summary, provenance/reproduction, interpretation
limitations, turn narratives with decision reasons, named observer views,
score/available-points/debt ledgers, ending reason, metrics, and tuning hypotheses
supported by the run. Full private traces are separate from shareable Markdown.
Comparisons distinguish paired evidence from exploration. Proposed gates below are
targets, not measurements or declarations of balance:

- All approved rule fixtures and card-conservation/accounting/visibility properties
  pass; zero accepted illegal actions or invariant failures. Cover the published
  scoring examples, response order, Q01–Q12, F01–F10, N01–N07 and T01–T03 outcomes, game-boundary
  ledgers, fixed match length/tied ranks and unsupported/experimental-case
  termination explicitly.
- Replay 1,000 seeded games per supported player count with identical transition
  digests, including worker-count changes; report incomplete cases separately.
- Evaluate each candidate on at least 1,000 held-out seed blocks per player count,
  complete seat rotations and fixed opponent lineups. Report paired score differences,
  declaration rates, seat effects and 95% confidence intervals clustered by seed block.
  Claim stronger play only when the preregistered score-difference interval excludes
  zero favorably; do not pool three- and four-player populations or tuning seeds.
- Pin deterministic operation budgets per bot for reproducible choices; collect
  p50/p95 decision time, games/second, peak memory, action counts, abort reasons and
  draw/redeal counts on recorded hardware. Set release performance limits after
  the first measured baseline; report budget exhaustion rather than hiding it.

Keep local runs for a proposed 30 days unless referenced by a review; retain cited
evidence and inputs until explicitly released. Future pruning previews paths and
never deletes curated inputs. Mobile parity is a separate gate: Go server/simulator
reuse does not establish parity with a Dart or native client engine.

## Phase 0 preparation and handoff

Historical preparation snapshot, written before Phase 0 implementation on
2026-09-30. Its then-absent-runtime statements describe that snapshot, not current
repository status. Read the measured Phase 0 evidence above and the authorized
extension in `.agents/instructions/PHASE_0_SIMULATION.md` before applying scope limits.
Read this section and the full CLI contract above; no conversation context is needed.
The preparation specifies implementation choices under the existing rules and
roadmap, without approving the proposed ADR or creating a second checklist.

Load in this order (all paths below are repository-relative):

| Order | Exact path | Purpose |
|---|---|---|
| 1 | [`AGENTS.md`](../AGENTS.md), [`xops/agent/session-bootstrap.sh`](../xops/agent/session-bootstrap.sh) | Operating rules, actual bootstrap/recovery and clean/dirty inventory. |
| 2 | [`docs/planning/ROADMAP.md`](../docs/planning/ROADMAP.md), [`.agents/instructions/ROADMAP_DISCIPLINE.md`](../.agents/instructions/ROADMAP_DISCIPLINE.md) | Sole S01–S10 sequence and complete acceptance gates. |
| 3 | [`docs/project/GAME_RULES.md`](../docs/project/GAME_RULES.md), [`docs/design/RULES-IMPLEMENTATION-QUESTIONS.md`](../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md), [`CHANGELOG.md`](../CHANGELOG.md) | Canonical gameplay and accepted Q01–Q12/F01–F10/N01–N07/T01–T03 decisions. |
| 4 | [`docs/design/ADR-0001-shared-go-simulation-engine.md`](../docs/design/ADR-0001-shared-go-simulation-engine.md), [`docs/code/ARCHITECTURE.md`](../docs/code/ARCHITECTURE.md), [`docs/code/MODULE-game.md`](../docs/code/MODULE-game.md), [`sims/README.md`](README.md) | Proposed shared-engine boundaries and unchanged full CLI contract. |
| 5 | [`.agents/instructions/PHASE_0_SIMULATION.md`](../.agents/instructions/PHASE_0_SIMULATION.md) | Scope, execution, recovery, review, validation and evidence rules. |
| 6 | [`sims/docs/phase-0-research.md`](docs/phase-0-research.md) | Inventory, cited comparison of simulation environments, engine/replay/accounting decisions and limits. |
| 7 | [`sims/docs/phase-0-operations-research.md`](docs/phase-0-operations-research.md), [`sims/docs/phase-0-experiment-research.md`](docs/phase-0-experiment-research.md) | Dated Go/toolchain, cancellation/build/testing evidence; fair bots, paired rotations, clustered intervals and claim limits. |
| 8 | [`sims/docs/phase-0-format-spec.md`](docs/phase-0-format-spec.md), [`config/rules/README.md`](../config/rules/README.md) | Configuration precedence/normalization, strict parsing, seeds/budgets/versions, replay/manifest schemas, privacy/retention and canonical defaults extraction. |
| 9 | [`config/schemas/phase-0-preparation.schema.json`](../config/schemas/phase-0-preparation.schema.json) | Structural schema for preparation inputs only; no runtime validation claim. |
| 10 | [`sims/scenarios/q10-reciprocal-thirds.preparation.json`](scenarios/q10-reciprocal-thirds.preparation.json), [`sims/scenarios/replay-encoding.preparation.json`](scenarios/replay-encoding.preparation.json) | Synthetic accounting-only oracle and encoding/hash/seed vectors; neither is a playable game or PRNG conformance result. |
| 11 | [`sims/experiments/baseline-3p.preparation.json`](experiments/baseline-3p.preparation.json), [`sims/experiments/baseline-4p.preparation.json`](experiments/baseline-4p.preparation.json) | Separate development pilots with complete cyclic rotations, paired blocks, bot/budget/analysis/output/retention specifications; not held-out strength evidence. |
| 12 | [`.agents/skills/cgms-simulation/SKILL.md`](../.agents/skills/cgms-simulation/SKILL.md) | Implementation procedure and partial-coverage boundary. |
| 13 | [`.agents/skills/cgms-deterministic-replay/SKILL.md`](../.agents/skills/cgms-deterministic-replay/SKILL.md), [`.agents/skills/cgms-rule-conformance/SKILL.md`](../.agents/skills/cgms-rule-conformance/SKILL.md), [`.agents/skills/cgms-simulation-experiments/SKILL.md`](../.agents/skills/cgms-simulation-experiments/SKILL.md) | Load the matching domain skill before replay, rule/accounting, or experiment work. |
| 14 | [`sims/docs/phase-0-execution-guide.md`](docs/phase-0-execution-guide.md), [`sims/docs/phase-0-validation.md`](docs/phase-0-validation.md) | Every S01–S10 dependency/path/clause/test/evidence/gate and preparation verification/implementation-only limits. |

Use the existing [skill index](../.agents/skills/README.md),
[tracking guide](../docs/tracking/README.md),
[tracking schema](../docs/tracking/tracking.schema.md), and [xops conventions](../xops/README.md)
for shared procedures rather than duplicating them. Generated `sims/artifacts/`
output is now ignored; no output exists or is needed for this preparation.
Promote inputs only after S03/S09 runtime schemas and semantic tests exist; retain
explicit preparation labels meanwhile. Start the subsequent implementation at S01,
without treating any documentation check as a completed runtime gate.
