# Phase 0 implementation evidence

Phase 0 S01–S10 gates passed, 2026-09-30. Bounded scenarios only; no complete-game conformance claim.

## S01 compiler and red/green

One module: `github.com/metaphy6/cgms/backend`, derived from the actual Git remote.
The language minimum and provisioned compiler are Go 1.27.1, linux/amd64.
`GOTOOLCHAIN=local` prevents implicit compiler download; tests also work with
`GOPROXY=off GOSUMDB=off`. Standard-library dependencies only so far.
Official [downloads](https://go.dev/dl/) freshly checked on 2026-09-30 list 1.27.1.
The existing Snap installation was used without system changes:

- GOROOT: `/snap/go/11295`.
- Go driver SHA-256: `30969f97169d7f43fe6a085873d75613adc21e30818a8c61d95bd27275df4624`.
- Compiler SHA-256: `3780fc07130d6d45096373acf062cc316940d253e2a1d892ece03a3bdd0d21d4`.
- Official linux/amd64 archive checksum: `63d339f0da5ab53635a56f2490a7984dfe12dfcff22ad749f63edaf590168445` (reference only; no archive downloaded).

Actual repository-root invocation:

```bash
pwd
xops/agent/safe-run.sh phase0-s01-behavior-red -- env GOTOOLCHAIN=local go -C backend test ./internal/game -run TestDoubleDeckPhysicalIdentity -count=1
```

The compiling stub failed: `physical inventory = 0, want 104`.
After implementing enumeration, the same test passed with tag `phase0-s01-green`
and process-local `GOPROXY=off GOSUMDB=off`.
Logs: `/tmp/agent-runs/phase0-s01-behavior-red--20260930T081505Z-1582515.log`
and `/tmp/agent-runs/phase0-s01-green--20260930T081519Z-1583050.log`.
An earlier go.mod tidy requirement was a setup failure, not red behavior evidence.
Independent reviewer found no blocker; isolated reviewer test log:
`/tmp/agent-runs/s01-review--20260930T081620Z-1584987.log`.

## Coverage limits

The implemented Phase 0 scope is complete. No complete-game conformance,
1,000-complete-game replay, 1,000-block strength, balance or release performance gate is claimed.
Preparation files retain their historical preparation-only status.

## S10 initial isolated packaging evidence

The initial implementation snapshot (`s10-initial`, before later runtime edits)
was built with Go 1.27.1 on linux/amd64, `CGO_ENABLED=0`, `GOTOOLCHAIN=local`,
`GOPROXY=off`, `GOSUMDB=off`, `-a -trimpath -buildvcs=false`, and link-time source
revision/dirty-patch hashes. `python3 xops/sim_build.py s10-initial` produced two
fresh-cache builds in different source directories with identical binary SHA-256
`33229a91932f56a120fc4ad4fb0c0edb9fbd1f1df5213d624695c3a51cad7963`.
A third clean build inside a scratch container with network disabled, read-only
source/toolchain mounts and an output-only writable mount produced that same hash.
The helper records compiler executable, compiler tool, module, source snapshot,
tracked diff plus untracked-source content, flags and binary hashes in ignored
`build-provenance.json`; it never injects the current time into the binary.

Docker client/server 29.7.2 built `backend/Dockerfile.sim` without network or base
image downloads. The initial image ID was
`sha256:7d8adbcd702926414e703d5ac7a487974e884062eee74da9a4698b173a9f2ac1`.
The actual scenario executions used `--network none --read-only --cap-drop ALL
--security-opt no-new-privileges --pids-limit 64 --memory 256m --cpus 1`, an explicit
non-root UID/GID, read-only inputs and a designated output mount. Private CLI
arguments were supplied only through protected `xops/sim_private.py` specifications.
No application services were started.

Host and container both completed the Q10 reciprocal-thirds accounting scenario
and replayed the host's recording with exit 0. Every manifest artifact's length and
SHA-256 was inspected. Outcomes, initial state, checkpoint and trace bytes matched
across both original runs and both replays. The checkpoint SHA-256 was
`13bd6c9d4c287c945d3bee68edb1727f2286d07f5fd03a5f145ce6ce97ff13e4`;
trace SHA-256 was
`afd8fd359293230db827a38a595ae6d7e3cca2ae377f26b4e461ee99ae6292c6`.
These are accounting-scenario/replay checks, not completed-game evidence.
Final source changes require a fresh build and repetition before S10 closes.

Durable safe-run logs: `sim-build-s10--20260930T083849Z-1662965`,
`sim-image-s10--20260930T083914Z-1666826`,
`sim-host-s10--20260930T084004Z-1668711`,
`sim-container-s10--20260930T084008Z-1668844`,
`sim-offline-build-s10--20260930T084037Z-1670172`,
`sim-host-replay-s10--20260930T084115Z-1671844`, and
`sim-container-replay-s10--20260930T084120Z-1672171`.
Private specs, manifests and the compact machine-readable evidence remain ignored
under `sims/artifacts/s10-isolation/`; clean-build inputs/results are under
`sims/artifacts/build/s10-initial/`.

## S03 configuration gate

Independent review and race verification passed canonical strict decoding, accepted
rules allowlisting, bounds/units and consumed normalization/hash fixtures. Logs:
`review-s03-rules--20260930T084810Z-1693179` and
`review-s03-random-bots--20260930T084810Z-1693203` (canonical/randomstream portions;
its separately failing bot fixture was repaired and independently rerun).
Configuration coverage does not imply every configured ability is implemented.

### Retention preview safety

`RetentionPreview` only proposes expired run directories; it never deletes files or rewrites manifests. It excludes review holds, unexpired runs, and runs referenced by any manifest in the scanned artifact tree (single parent manifest or exactly two comparison parents separated by ` | `). Relative references resolve against recorded CWD. Ambiguous reference syntax fails closed. Enumeration excludes symlinks; existing parent-reference symlinks resolve only for protection. Ancestors containing protected evidence are excluded too. References outside the scanned tree cannot be discovered; callers must preview the complete artifact root before any separate deletion procedure.

Observed behavioral assertion red: `retention-red--20260930T085741Z-1713332`; race-test green: `retention-final--20260930T085815Z-1714834`. Tests cover expired, unexpired, equal expiry, held, comparison-referenced, symlinked runs, immutable manifests and ambiguous references. These are retention safety fixtures, not gameplay coverage or measured balance results.

## S02/S04 gate verification

Independent review repaired snapshot-admission panics, inconsistent saved decision
choices, deferred-draw admission and public identity-history retention before gate
acceptance. The verifier inspected the full S02/S04 clauses and named fixtures;
`verify-s02-s04-race--20260930T085926Z-1717712` passed game, canonical, randomstream,
bots and simulator race tests. `verify-s02-s04-schema--20260930T085926Z-1717733`
verified generated runtime schemas. S02 covers physical conservation, binding
lifecycle, active-Ace custody, availability markers and aliasing. S04 covers narrow
transition/projection goldens, effect continuation/retry/Coup, deterministic streams
and T02 entitlement/recycle/exhaustion/continuation. Complete-game conformance and
full reconnect history for unimplemented effects remain outside this evidence.

## S06/S08 gate verification

Independent verifier `verify-s06-s08-final--20260930T090502Z-1736483` passed targeted
race tests across game/bots/sim after review repairs. Combat now revalidates support
at resolution without refunding already paid costs, Justice requires Clubs or the
accepted Underground waiver, numerical/Advancement effects offer eligible Dexter,
and Inflation tests cover exact odd halves, nineteen/twenty, cumulative small spends
and duplicate no-op activation. Shared legal-menu versioning, projection/alias
privacy, independent legality checks, bounded uniform sampling and named 3p/4p
cyclic pilots passed. The menu remains a named one/two-Club subset; Bomb cost/algebra
fixtures do not implement a full Bomb response loop; surviving Ponzi card-to-ledger
resolution and full turn/game progression stop explicitly. These are not stronger-bot
or complete-rule claims.

## S05 setup and S07 accounting

S05 review/verification exercised recursive numeric draw-offs (isolated drawn cards,
ignored non-numbers, repeated ties and exhausted supply), fixed order, diamondless
redeals, whole-suit first opening and historical reopening. The consumed synthetic
seed fixtures exercise zero, one and two redeals without reseeding. The fixed
second-historical-suit protection regression failed at `phase0-protection-red`
and passed after correcting the four-suit check; final setup/initiative race checks
are included in `s04-admission-race-final--20260930T084631Z-1689141`.

S07's complete red/green and oracle record is in
[the curated ledger evidence](../../backend/testdata/ledger/EVIDENCE.md).
Independent review/verification accepted that gate before marking it. The final
51-digit-denominator example uses two durable operations, two compressed segments,
and a 2,826-byte serialized checkpoint. This proves bounded work for that cycle,
not a general constant-time settlement claim.

## Final source verification

`phase0-final-format--20260930T091107Z-1760599` passed all packages, vet and generated
schema consistency. Independent `final-full-race--20260930T091127Z-1762422` passed
the complete race suite. Five existing fuzz targets, two workers and two requested
seconds each, passed in `final-bounded-fuzz--20260930T091150Z-1766856`:

| Target | Observed executions | Observed elapsed |
|---|---:|---:|
| Decode | 40,316 | 3.040 s |
| SettlementEquivalence | 68 | 2.207 s |
| StatePartition | 10,171 | 2.040 s |
| AmountArithmetic | 110,719 | 2.058 s |
| RejectedCommandPreservesState | 955 | 2.798 s |

The total is 162,229 bounded fuzz executions, not exhaustive proof. Import review
(`final-pure-engine--20260930T091151Z-1767051`,
`final-engine-imports--20260930T091204Z-1768023`) found no production engine clock,
filesystem, SQL, network or global-random dependency; the decoder's `io` use is EOF
handling only. No benchmark functions existed; runtime metrics are measured by
the adapter and excluded from semantic outcomes.

Later adapter-only continuation/hold changes passed complete tests, vet and schema
checks in `phase0-final-continuation-format--20260930T091538Z-1777571` and independent
targeted race/review in `verify-s09-continuation-hold--20260930T091602Z-1778904`.
The engine was unchanged after the full race/fuzz run. Current replay also records
its verifier binary separately from the original recording provenance; targeted
replay checks passed `phase0-final-verifier-provenance--20260930T091639Z-1781310`.

Representative adapter assertion-red/green pairs include strict command flags and
rehashed outcomes, seeded bot transcripts, match-slot denominators, settlement and
output budgets, malformed schedule/checkpoint rejection, match-index stream
derivation, scenario/experiment provenance, publication I/O, paired identity and
parent-checkpoint provenance. Each recovery read the captured failure before the
fix; expected assertion failures are not counted as green validation. Full private
workload arguments never entered those logs.

Privacy verification ran the actual safe-run script in an isolated fixture:
canaries stayed out of console, command/log/exit files and failure breadcrumbs,
while the protected child log retained them. Traversal and launch-signal races
were repaired; logs `privacy-audit-red--20260930T090206Z-1725765`,
`privacy-audit-verified--20260930T090310Z-1730825` and
`privacy-path-final--20260930T090336Z-1732069` record the result. Cooperative SIGINT
returned 130 within the tested one-second bound (five-second outer cap); arbitrary
uncooperative external programs are not covered by that result.

## S10 final frozen-source isolation verification

Final build `phase0-final-0916` produced SHA-256
`7d558db167ae61cbb6ddcd9306c9e3f677d7b6c256003600107ba73836bebba4`
from two clean caches in different workspace paths. A third build inside the scratch
container, with mounted Go 1.27.1 toolchain/source and network disabled, produced the
same binary. The image is
`sha256:eb55dc17ccf77fa81952c0cfdac86af11d1aa43e731891c1e89ffc57d968829e`.
The build records actual source revision, dirty patch/source snapshot hashes,
compiler/module hashes and flags in ignored `build-provenance.json`.

Host and container executed `q10-reciprocal-thirds`; both host and container replayed
the host recording. All four exits were 0. Inspection verified every artifact's
length/hash (23/23/25/25 artifacts), identical initial/checkpoint/trace/outcome bytes,
and final-binary verifier provenance. All four manifests were published with
`--review-hold`; no post-publication edits were needed. Container runs used
`--network none`, read-only root filesystem, nonroot UID 1000, dropped capabilities,
no-new-privileges and no application services. CLI arguments/output remained in
restricted specs/captures under ignored `sims/artifacts/phase0-final2-isolation/`.
The public report explicitly records zero finalized games and one completed accounting
scenario; this evidence establishes neither complete-game execution nor balance.

Captured successful runs: `phase0-final2-build--20260930T091658Z-1782077`,
`phase0-final2-image--20260930T091728Z-1786395`,
`phase0-final2-host--20260930T091753Z-1787509`,
`phase0-final2-container--20260930T091753Z-1787569`,
`phase0-final2-offline-build--20260930T091753Z-1787541`,
`phase0-final2-host-replay--20260930T091803Z-1788716` and
`phase0-final2-container-replay--20260930T091803Z-1788695`.
Earlier `s10-initial` and `phase0-final-0911` runs are superseded candidate evidence.

The final runtime JSON-schema inspection additionally passed
`phase0-final-runtime-schemas--20260930T091941Z-1793284`: one rules input, two
experiment profiles, three curated scenarios, four isolated manifests, eight
initial/checkpoint states, eight trace records, four metric records and four
outcomes. All four isolated manifests had review holds at publication.

## S09 final executable evaluation and phase exit

Final binary `7d558db167ae61cbb6ddcd9306c9e3f677d7b6c256003600107ba73836bebba4`
passed all 18 protected workload invocations in 246.642 seconds:

```sh
pwd
xops/agent/safe-run.sh phase0-final-evaluation-v2 -- python3 xops/sim_evaluate.py sims/artifacts/build/phase0-final-0916/sim phase0-final-evaluation-0917 --samples 1000
```

Use new build/output IDs to reproduce; existing output is never overwritten.
The helper exercised Q10 run/replay; 3p/4p bot run/replay; 1,000-match batches at
one and four workers in each population; full replay of each one-worker batch;
paired comparisons; complete cyclic pilot tournaments; and two-value policy
sweeps. Outcomes, schedules, initial states, checkpoints and traces were byte-equal
across worker counts. Every manifest artifact checksum/length was inspected.
All 18 manifests recorded review holds at publication.

There were **1,000 replayed seeded partial prefixes per population, zero finalized
games**. Each 3p/4p bot fixture commits three/four supported actions respectively
and stops at the next unsupported boundary. Q10 returns 0 with scenario-complete
and game-complete false; batch/tournament/sweep return 3; successful paired
analysis returns 0 without upgrading its inputs. No score interval is estimated
from these incomplete blocks. The 1,000-complete-game and 1,000-held-out-block
strength targets remain unfulfilled; no game balance/stronger-play claim follows.

`phase0-final-report-inspection--20260930T092223Z-1805622` separately validated
18 manifests, 6,048 outcomes, 88 sampled trace records, 52 sampled initial/checkpoint
objects and 16 metrics objects against runtime schemas. It inspected public
comparison text, both population denominators, all review holds and zero finalized
counts. The workload log is `phase0-final-evaluation-v2--20260930T091737Z-1786724`;
restricted manifests, reports and the compact summary remain under
`sims/artifacts/phase0-final-evaluation-0917/`.

Measured worker-execution interval (excluding artifact publication and verification):

| Population / workers | Actions | Bot calls (including terminal unsupported menu) | p50 / p95 call time | Sampled Go heap | Worker interval |
|---|---:|---:|---:|---:|---:|
| 3p / 1 | 3000 | 4000 | 1307 / 3157 µs | 122461072 bytes | 13761416 µs |
| 3p / 4 | 3000 | 4000 | 1502 / 3116 µs | 124186896 bytes | 3937769 µs |
| 4p / 1 | 4000 | 5000 | 1383 / 3376 µs | 155127112 bytes | 19969689 µs |
| 4p / 4 | 4000 | 5000 | 1543 / 3445 µs | 175477024 bytes | 5648891 µs |

These diagnostics are hardware/workload-specific; sampled heap is not peak RSS,
and the worker interval is not end-to-end throughput. Finalized games/second is
zero. Proposed release thresholds remain separate from these measurements.

Completion means the Phase 0 roadmap gates and explicit partial-simulator exit,
not Phase 1. Unsupported card/finance orchestration, full-turn progression, all
ability combinations, proposals/promises and full durable server semantics remain
unchecked Phase 1 work. No Flutter/database/cache/telemetry/LLM service was added.
