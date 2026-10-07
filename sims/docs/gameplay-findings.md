# Accepted gameplay: implementation and measured findings

**Latest v3 conclusion:** practical bot strength, financial competence and game
balance remain **inconclusive**. Investigation reproduced a policy-ordering defect
that confounds legacy comparisons and repaired it in explicitly versioned v2
policies. Fixed-input performance improvements and historical replay completion
are supported; no corrected-policy confirmation sample has been collected.
See [final v3 status](#final-v3-status-and-verification) for current gates and limits.

The original report through “Reproducible executions” below is a **historical
snapshot** of the exhausted v2 allocation. The appended [v3 investigation](#v3-development-investigation-and-verification)
records subsequent verification and development separately; it does not enlarge
the historical independent sample or retrospectively change its thresholds.
The latest investigation found a legacy reconfiguration tie-key defect: reported
v1 strength/ablation contrasts remain descriptive and confounded. Prospective
confirmation requires corrected v2 policies on both sides and fresh inputs.

Evidence date: 2026-09-30. Collection stopped at the preregistered 1,600-second
resource cap; this is an underpowered campaign, not the requested full sample. The engine, targeted
conformance evidence, autonomous integration, policy competence and balance are
separate claims. Historical Phase 0 produced 2,000 replayed partial sequences and
**zero complete games**; none are relabeled here.

## What is implemented and checked

D01–D04 shared-domain gates pass. The pure engine supports complete turns,
formation/ability/response and accepted transfer lifecycles, explicit chance and
choices, ordinary/Coup endings, departures/nullification, exact finance and
subsequent games. D08–D09 simulator portions retain pending decisions, committed
randomness, settlement progress, fixed match length and fresh identities through
local artifacts. Their online persistence/transaction gates remain unchecked;
no PostgreSQL, Redis, API, Flutter or production service was introduced.

The [coverage crosswalk](gameplay-coverage.md) maps accepted Q/F/N/T decisions and
every §2.12 schedule row to tests. Independent review repaired historic-game retry
ordering, match snapshot aliasing, consent impersonation, all-skipped response
continuation and public financial disclosure. Pure transitions have no clock,
filesystem/database or global-random dependency. Every observed game still tests
only its particular path; targeted fixtures cover rare interactions. Passing this
matrix is not a mathematical proof of all state combinations.

## Frozen versions and provenance

Current campaign: `cgms-engine-v2.1`, `sampled-all-categories-v2`, public stats
`cgms-gameplay-public-v2`; economic/pressure policies remain `@v1`. Binary SHA-256:
`61be758da8269e25969d4fe17bf567572d1e4994478031b2818e63ccd1b8f0e4`.
Two clean, offline, different-path builds were identical. Exact compiler/source/
configuration hashes are retained in
`sims/artifacts/build/gameplay-candidate-v2-20260930/build-provenance.json`.
The accepted rules package was not changed. Engine/menu/source versions must
accompany policy names; historical saved binaries replay their own artifacts.

The [original plan](../experiments/gameplay-heldout-v1.md) and
[defect-repair amendment](../experiments/gameplay-heldout-v2.md) preregister 1,000
seed blocks per population, complete cyclic rotations, paired treatments, exact
focal score differences and 10,000 seed-block bootstrap resamples. A proposed
5-point meaningful improvement and 5-point interval half-width are **thresholds**,
not measured findings. No full opponent-order independence claim is made: cyclic
rotations do not enumerate all permutations. Roots and forensic traces remain
restricted; public reports omit hidden cards and private financial terms.

## Development games and defects discovered

These development results are not held-out estimates:

| Evidence | Actually observed | Limits |
|---|---|---|
| Initial CLI engine-v2 pilot | 6/6 three-player and 8/8 four-player games finalized and separately replayed | One seed block/population; retired implementation; old financial report projections remain restricted |
| DEV3 integration | Three-player two-game match and one four-player game finalized/replayed; workers 1/4 yielded identical entire transcripts | Three distinct games, not six independent samples; no formation openings or voluntary economic deals |
| Random-opponent v1 diagnostic | Six planned, three started, zero board-ended/finalized; three never started | Missing empty-response adapter continuation and first-turn promise churn discovered; preserved failure evidence |
| Final CLI2.2 two-game tournament | 12 planned slots; 10 started; 8 finalized/replayed games; 4/6 complete two-game matches | Two first games hit the menu-generation budget; two subsequent slots never started; one development block |
| Random-opponent v2 diagnostic | Six planned/finalized/replayed, comprising three distinct cyclic trajectories repeated twice | One three-player development block; engine and random sampler changed jointly; no causal strength estimate |

The random-v1 trace reached 712 promise offers and 170 acceptances within a
1,056-step first turn. This was a policy/scheduling interaction under legal
negotiation, not evidence that accepted rules force nontermination. The separately
versioned random-v2 sampler first chooses a financial operation class uniformly,
then an entry within it; it cannot turn a missing response into a pass. The adapter
repair records authenticated `resolve-ready` when the response list is exhausted.
New random diagnostics exercised that continuation and completed, without tuning
the frozen economic/pressure policies.

The random-v2 player explicitly departed early in all three distinct rotations
and scored zero; pressure led those three games. This explains why beating that
baseline would be weak evidence of robust competence. All three games ended at
round 13, with 28/28/27 started turns and 478/394/370 transitions. Run wall time was
105.105 seconds, sampled Go heap 17,997,360 bytes (not peak RSS), and artifacts
6,368,128 bytes. The six repeated executions recorded 1,956 policy decisions and
35,034 operations. See restricted
`sims/artifacts/gameplay-random-diagnostic-v2-1100/inspection.md` for public-event
annotations and `economics-review-private.json` for separately private diagnostics.

## Held-out results and unmet targets

The corrected, checksum-verified report is
`sims/artifacts/gameplay-heldout-v2-20260930/reanalysis-000037-0002-report.md`.
Collection used 1,600.623 seconds and 203,214,556 bytes at the stop (below the
1.75 GiB cap). Analysis files add a small amount afterward. The final four-player
replay received cancellation at the wall limit; committed run results survived.
All six run shards finalized; no gameplay failure was omitted.

| Measure | Three players | Four players |
|---|---:|---:|
| Planned games / one-game matches | 6,000 | 8,000 |
| Started / board-ended / financially finalized games | 18 / 18 / 18 | 24 / 24 / 24 |
| Finalized matches | 18 | 24 |
| Complete paired seed blocks | 3 / 1,000 | 3 / 1,000 |
| Never-started game slots | 5,982 | 7,976 |
| Unknown slots / voided instances | 0 / 0 | 0 / 0 |
| Fully replay-verified finalized games | 18 | 16 |
| Worker 1/4 identical finalized games | 6 | 8 |
| Remaining blocks to target | 997 | 997 |
| Remaining completed-game replay checks to 1,000 target | 982 | 984 |

Worker reruns and replay executions are not additional game samples. Eight
four-player games finalized but their replay shard was interrupted; no full replay
claim is made for that shard. A separate committed-deal audit found zero denominator
discrepancies in all 42 frozen-v2.1 games, despite the later v2.2 reporting edge fix.

The preregistered primary estimate is pressure-minus-economic focal finalized
match score, paired within each rotation and averaged within each seed block:

| Population | Mean effect (points) | Seed-block bootstrap nominal 95% interval |
|---|---:|---:|
| Three players | −110/9 (−12.22) | [−86, 25] |
| Four players | 111/4 (27.75) | [2, 54] |

These use the frozen 10,000-resample analysis. With only **three independent blocks**,
the percentile intervals have poor calibration and cannot substantiate a strong
95%-coverage claim. Their widths (111 and 52 points) exceed the proposed 10-point
full-width precision goal. Even the positive four-player interval does not justify
a stronger-bot claim: the sample target and precision goal were missed, opponents
were limited, and the three-player effect is unfavorable. Results condition on the
completed prefix of the planned campaign; resource-limited selection and fixed
opponent composition limit generalization. No tuning used these held-out outcomes.
P03 remains unchecked. Seat, matchup and balance conclusions remain hypotheses.

The original v1 campaign is preserved separately: six three-player games finalized
before a defect-driven stop, with a later recorded-decision replay of all six;
zero four-player games started. These were not pooled with the amended campaign.
The original v2 aggregate report mistakenly labeled valid score-object outcomes
unverified because the Python decoder expected strings. An actual-CLI-shape
regression caught this; a strict canonical rational decoder and `--analyze-only`
produced the correction above, verifying original manifests/frozen inputs and
preserving all original artifacts/state/budgets. The statistical method, policies,
seed selection and game results did not change.

Measured run-only wall times were 173.233 seconds for the three three-player
blocks and 371.950 seconds for the three four-player blocks. A linear run-only
projection for 1,000 blocks each is about 50.5 hours at this configuration, before
replay/verification/storage overhead; it is a rough projection, not a measured
full-campaign cost. The observed campaign consumed about 194 MiB for six blocks
including verification and frozen planning files. Profile serialization/menu work
and measure larger development pilots before allocating a much longer continuation.
The saved state preserves every unstarted unit. Do not silently extend the exhausted
allocation or tune the frozen candidate while continuing this held-out cohort.

## Observed gameplay patterns

Across the held-out run manifests, all 18 three-player games ended at round 13;
22 of 24 four-player games ended at round 13 and two by an ordinary Diamonds
declaration. No Coup, voided instance, tied score leader or redeal occurred in
this sample. These absences are observed-use gaps, not missing targeted fixtures.
Three-player games took 39 started turns and 467–648 transitions; four-player
games took 13–52 turns and 396–1,164 transitions. No run shard hit an engine wait,
loop or budget stop. The campaign's final replay was interrupted by the wall cap.

Physical attacks/purchases were common: 357/283 in three-player games and 817/631
in four-player games. Three-player runs recorded 75 Dexter assassinations, 28
Kidnapper uses, 24 Confinements, 17 Compensations, 14 main Inflations and 13 Bomb
attacks; four-player runs recorded 92/0/37/30/13/29 respectively. Three formation
openings occurred in three-player play and none in four-player play. The public
menu recorded 140 Kidnapper formation opportunity visits, not 140 distinct states
or chances to win. Raw counts do not establish usefulness, dominance or causality.
The heavy attachment/purchase pattern and royal-retention weakness motivate a
controlled development-only policy ablation; changing rules is not justified.

Seat game-score leaders were 8/6/4 (three players) and 6/7/3/8 (four players).
Descriptive mean scores by seat were 105.22/97.67/82.56 and
83.79/87.96/79.13/85.08. Only three independent seed blocks underlie these numbers;
paired treatments share seeds and opponents, so the 42 games are not 42 independent
seat-effect observations. These apparent differences cannot establish a seat
advantage. The two declarers were seat 2 and also led their game scores; all other
leaders won on finalized score without a successful declaration. The campaign uses
one-game matches, so its cumulative match scores equal game scores. Multi-game
persistence evidence is reported separately below.

Representative restricted artifact references (public projection files):

- `gameplay-heldout-v2-20260930/p3-s0000-run/games/000000/gameplay-public.json`:
  one formation opening, seven Kidnapper uses, eight Dexter assassinations and
  22 purchases; round-13 ending after 39 turns/648 transitions; final scores
  121/56/47. This illustrates legal ability integration, not Kidnapper causality.
- `gameplay-heldout-v2-20260930/p4-s0002-run/games/000000/gameplay-public.json`:
  23 purchases and 13 attacks precede seat 2's round-4 Diamonds declaration;
  13 started turns/443 transitions; final scores 19/206/47/37. This is a genuine
  board ending and financial finalization. Its shard replay did not finish before
  the resource cap, so this example is not labeled fully replay-verified.
- `gameplay-final-v22-multigame/run/games/000002/gameplay-public.json`:
  first-game legal-menu budget stop, preserved partial state and no invented
  subsequent game. Its replay retains the original incomplete status.

Private financial trajectories were inspected separately: completed ledgers,
nonempty carried debts and exact settlement receipts persist across game boundaries.
The heuristics rarely undertake voluntary financial deals, so these trajectories
and final score patterns cannot establish whether loans, promises or forgiveness
are strategically balanced. Exact financial terms and hidden holdings remain in
restricted artifacts; targeted oracle/interaction fixtures provide the conformance
evidence that this narrow policy population does not supply autonomously.

## Bot competence: supported and unsupported conclusions

Both heuristics select an immediately legal ending, use legal defense, avoid a
second attached king that disables single-king income, prefer an already available
combo over attachment, account for assassination cost, and handle required exact
financial decisions. A controlled public-threat fixture gives an opponent visible
13 Diamonds and two equally legal attacks; both remove its Diamond to leave 12,
with the correct committed physical Club cost. These are targeted competence
checks, not general lookahead or opponent modelling.

A documented weakness is immediate attachment of an unpaired royal, sacrificing
future combo options. The three DEV3 games had no formation openings; this does
not show that combos themselves are weak. Board policies lack a strategic debt
model and deliberately avoid unvalued negotiations; financial policies see only
their own authorized cash/debt information. Pressure's refusal preference incurs
canonical reputation consequences. No policy sees opponent hidden hands, future
draws or private seeds. Menus sample representative physical combinations rather
than every possible combination, and can exhaust their explicit operation cap.

Prioritize simple development-only improvements—royal retention, opportunity
valuation and proposal evaluation—before expensive hidden-state search. Search or
belief modelling has not yet been justified by held-out gains. Broader lineups,
nontransitive matchups and controlled paired ablations are needed before claiming
robustness. Automated play cannot establish enjoyment, clarity or social dynamics;
human tests should focus on negotiation comprehension, refusal incentives,
response/settlement downtime and repetitive strategies.

## Verification actually exercised

- Complete Go test suite and `go vet ./...`: `gameplay-v21-suite-verified--20260930T110243Z-2301502`.
- Full game race matrix: `timing-crosswalk-game-race--20260930T104541Z-2233261`; later D02/D03 boundary, resolve-ready, privacy, finance-artifact and bot sampler/competence race checks passed separately.
- Exact settlement oracle fuzz: 185 executions in 10 seconds; rejected-command immutability fuzz: 3,532 executions in 10 seconds. These bounded counts are not exhaustive verification.
- The original twenty campaign Python integrity/resume/denominator/cluster tests and the original nine generated schema families: `gameplay-python-final--20260930T110852Z-2313289`, `gameplay-schema-final--20260930T110852Z-2313288`.
- CLI tests cover strict input/override/version validation, no overwrite, cancellation and completed-prefix retention, nonzero exits, replay tampering, pending continuation, hidden-state observations and worker ordering.

Final CLI2.2 full Go suite and vet passed (`gameplay-v22-final-suite--20260930T111736Z-2339165`); the additive public/private schema packaging regression subsequently passed, and all ten runtime schema families were verified (`gameplay-final-schema-contract--20260930T112605Z-2362763`). The committed-deal denominator regression passed race checking, including cancellation before first deal and between game allocation and deal. Optional private financial diagnostics retain a bounded prefix with explicit omitted counts; authoritative exact ledger obligations remain in the checkpoint.

The final exact-boundary diagnostic-cap race regression passed (`gameplay-private-terminal-fixed--20260930T112810Z-2368260`). The analysis correction first passed 23 Python tests including actual rational-object output, checksummed saved-manifest reanalysis, immutable correction publication and tamper rejection (`evaluator-analysis-final--20260930T113317Z-2386273`). Final independent review added a canceled-replay metadata regression; all 24 evaluator tests passed (`evaluator-stage-green--20260930T113918Z-2397326`), and hardened reanalysis reproduced identical counts/estimates without rerunning games. The final binary also passed 18 CLI contract runs (`gameplay-release-cli-contracts--20260930T113314Z-2385941`); these accounting/partial checks add zero complete games.

Final independent review closed the metadata finding with 24 passing tests (`review-campaign-status-closure--20260930T113953Z-2398100`); no remaining findings in that review scope. Source/report privacy scans, documentation links and whitespace checks passed.

Captured red-before-green logs establish behavioral changes; additional
conformance fixtures that passed on first execution are labeled as such in the
coverage record. No skipped assertions or silently manufactured endings were used.
Generated artifacts remain ignored; only curated fixtures and this concise evidence
are source-controlled. No commit or push is performed by the agent.

## Container and persistent-match evidence

The engine2.1 host/container development tournament finalized four of six games;
two stopped at the menu-generation budget. All six transcripts and outcomes were
identical, and container replay verified the four completed games and two partial
prefixes. The image ran without network/application services, with read-only root,
dropped capabilities, non-root user, one CPU and 256 MiB memory. It used a read-only
workspace mount and one writable output mount. This is separate development evidence,
not an addition to held-out denominators.

The final CLI2.2 two-game tournament finalized four matches, with fresh game IDs,
retained completed ledgers and unchanged carried debts across next-game entry;
a nonempty carried debt was observed. Eight finalized games and both interrupted
prefixes replayed. No tied final ranks occurred; ties have targeted fixture coverage.
It used 119.466 seconds, 46,821,688 bytes sampled Go heap (not RSS), 7,990,803 bytes
of artifacts, 3,292 transitions and 62,120 policy operations. Restricted inspection:
`sims/artifacts/gameplay-final-v22-multigame/inspection.md`.

Final release is CLI/engine `cgms-engine-v2.2`, binary SHA-256
`37b7de0e05e8648ce92f6e29ea4fcf47594cf10e0e9d0962c2cfab68cbcbe287`.
Its rule engine and bot source files are byte-identical to the frozen campaign;
changes correct committed-deal counts, cap optional private diagnostics and bundle
public/private output schemas. Two clean builds in different paths agreed.
The final scratch image `cgms-sim:gameplay-release` was built with network disabled
and ran the Q10 accounting fixture on host/container with equal outcomes and exit
0. Twelve schemas were archived (ten generated families plus rules/experiment).
Execution used only read-only config/project-doc/scenario/experiment mounts and
one output mount, network disabled and the constraints above. This accounting
smoke is **zero additional completed games**.

Privacy review found no actual held-out roots or analysis seeds in source or
shareable report files. One process diagnostic inadvertently exposed a seed in the
operator tool transcript; it was not copied into source, report files, bots or
training inputs. Keep that operator transcript restricted alongside forensic
artifacts. Seeds and hidden-state traces must not be shared as ordinary reports.

## Reproducible executions

From repository root, the following commands were actually executed. Seeded argv
is retained only in protected specifications; paths are local private inputs.

```sh
pwd
xops/agent/safe-run.sh gameplay-candidate-v2-build -- python3 xops/sim_build.py gameplay-candidate-v2-20260930
xops/agent/safe-run.sh gameplay-heldout-campaign-v2 -- python3 xops/sim_gameplay_evaluate.py sims/artifacts/build/gameplay-candidate-v2-20260930/sim sims/artifacts/gameplay-candidate-v2-inputs/plan.json gameplay-heldout-v2-20260930
xops/agent/safe-run.sh gameplay-v21-host-tournament -- python3 xops/sim_private.py sims/artifacts/gameplay-container-v21/host-tournament.spec.json
xops/agent/safe-run.sh gameplay-v21-container-tournament -- python3 xops/sim_private.py sims/artifacts/gameplay-container-v21/container-tournament.spec.json
xops/agent/safe-run.sh gameplay-final-v22-multigame -- python3 xops/sim_private.py sims/artifacts/gameplay-final-v22-multigame/run-spec.json
xops/agent/safe-run.sh gameplay-release-build -- python3 xops/sim_build.py gameplay-release-20260930
xops/agent/safe-run.sh gameplay-release-container-build -- docker build --network none -f backend/Dockerfile.sim -t cgms-sim:gameplay-release sims/artifacts/build/gameplay-release-20260930
xops/agent/safe-run.sh gameplay-release-host -- python3 xops/sim_private.py sims/artifacts/gameplay-release-smoke/host.json
xops/agent/safe-run.sh gameplay-release-container -- python3 xops/sim_private.py sims/artifacts/gameplay-release-smoke/container.json
xops/agent/safe-run.sh gameplay-final-corrected-analysis -- python3 xops/sim_gameplay_evaluate.py sims/artifacts/build/gameplay-candidate-v2-20260930/sim sims/artifacts/gameplay-candidate-v2-inputs/plan.json gameplay-heldout-v2-20260930 --analyze-only
```

Fresh execution requires fresh build/output/spec capture IDs. A campaign resume
uses its existing immutable inputs and `--resume`; it never erases failed units.
An exhausted frozen allocation requires an explicitly documented continuation
allocation, not editing old results or silently extending until a favorable interval.

## V3 development investigation and verification

The prospective [v3 allocation and claim definitions](../experiments/gameplay-evidence-v3.md)
separate implementation equivalence, policy development and untouched confirmation.
No new confirmation sample has been collected in this investigation. Canonical
GAME_RULES is unchanged; balance variants have not been justified or adopted.
Legacy v1 policy score contrasts below are preserved evidence, not clean causal
feature effects after the subsequently discovered tie-key defect. The prospective
confirmation wrapper requires uniformly corrected `economic@v2`, `pressure@v2`
and `opportunity@v2`; it rejects old or mixed policy versions.
Prospective campaign schema-v3 additionally freezes and retains the actual Python
runner and claim-extraction source bytes. Resume and subsequent dispatch reject
analysis-source drift; older campaign schemas remain unchanged. A statistical
method name alone is not sufficient analysis implementation provenance.

### Historical verification completed

All checksummed artifacts in the six original v2 run manifests were independently
verified. They still contain **18 three-player and 24 four-player finalized games,
three independent paired blocks per population**. The eight outstanding games in
`gameplay-heldout-v2-20260930/p4-s0002-run` now replay successfully with the original
saved binary, SHA-256
`61be758da8269e25969d4fe17bf567572d1e4994478031b2818e63ccd1b8f0e4`.
The new replay verified all 88 published artifacts and produced outcomes
byte-identical to that source shard. Its restricted manifest and summary are
`sims/artifacts/gameplay-historical-verification-v3/replay/manifest.json` and
`verification-summary.json` in the same verification directory.

Verification took **110.421 seconds**, **70,228 KiB peak child RSS**, and
**14,572,568 bytes** at measurement, within the separate 600-second/256-MiB cap.
Historical completed-game replay totals are now **18/24**; remaining targets are
**982/976**. This contributes zero new games or independent blocks. The original
canceled replay and original campaign state remain intact.

### Measured equivalent optimization

Profiling identified repeated canonical serialization/parsing and discarded
narrow-menu generation in optional out-of-turn observation. Removing the redundant
untyped JSON round trip, emitting unchanged canonical string bytes in chunks, and
projecting the same observation before generating the actual optional menu reduced
cost without changing rules, sampled actions, decisions or deterministic budgets
on the fixed development inputs.

| Fixed development input | Original run | Optimized run | Optimized replay | Transcript bytes |
|---|---:|---:|---:|---:|
| Three players, 546 steps | 18.837 s | 8.050 s | 1.737 s | 935,311 |
| Four players, 623 steps | 29.358 s | 14.183 s | 2.075 s | 1,069,122 |

These are approximately **2.34×/2.07×** run speedups on one fixed fixture per
population. Full transcript bytes matched the original captures; inspection also
confirmed identical final/projection captures. The menu benchmark fell from
280,773,463 to 119,128,588 ns/op, with allocated bytes per operation falling from
205,659,404 to 117,579,405. Final projection test peak RSS was 99,516 KiB.
Evidence: `sims/artifacts/performance-20260930/measurement-summary.json`, the cited
logs, profiles and private captures. Shared-host timings and fixture size limit
extrapolation; these do not establish sustained three-game campaign throughput.

### First three-game development campaign: incomplete

The frozen first development run compared `opportunity@v1` against frozen
`economic@v1`, with pressure/economic opponents, complete cyclic rotations,
three-game matches, the unchanged sampled menu, 100,000 menu operations and the
historical 4-MiB transcript cap. It planned two blocks per population and stopped
on a deterministic resource failure after **320.317 seconds**, not on a score.
All run-artifact lengths and checksums were independently checked.

| Denominator | Three players | Four players |
|---|---:|---:|
| Planned games / matches | 36 / 12 | 48 / 16 |
| Started games | 18 | 22 |
| Board-ended / financially finalized games | 18 / 18 | 18 / 18 |
| Complete matches | 6 | 4 |
| Complete paired blocks | 1 / 2 | 0 / 2 |
| Never-started slots | 18 | 26 |
| Verified completed-game replays | 18 | 0 |

Thus 84 slots were planned, 40 started, 36 finalized and 44 never started.
The second blocks account for 42 unstarted slots; two later slots were never
entered after first-block match interruption. Four four-player matches stopped:
two at the whole-match transcript cap (2,544 and 2,528 committed steps retained),
two at the menu-generation cap (1,117 and 1,398 steps). None is imputed as a draw
or silently replaced. Four-player completed matches alone cannot form the paired
block; its primary effect estimate is unavailable.

The sole complete three-player block gives **+253/9 points per game** for the
candidate, with **no confidence interval** and no confirmation claim. The saved
legacy runner reports +253/3 match-total points; dividing by the frozen three-game
horizon produces the per-game number above. Selection, one-block uncertainty and
the limited opponent roster prevent strength or balance conclusions.
Restricted provenance: `sims/artifacts/gameplay-evidence-v3-development/state.json`,
`analysis-000009.json`, and the two first-shard run manifests. Its original 78,286,922
bytes and incomplete evidence are retained.

### Competence and interpretation limits

Focused fixtures support that the new candidate reserves early jacks/queens for
formation opportunities, accepts free ordinary gifts without creating a loan
alliance, ranks affordable reduced promise delivery, and uses detached own cash,
owed debt and public remaining-game count. Frozen economic/pressure policies remain
available, and retention/finance ablations have separate policy identities.
Privacy fixtures exclude opponents' finances and hidden-state/menu interference;
legal action selection remains independently adjudicated.

These are **supported fixture-level behaviors**, not demonstrated practical bot
strength. The first development campaign did not activate the new carried-debt
priority feature; it therefore supplies no empirical evidence that this feature
improves multi-game decisions. Legal negotiation, loan valuation, opponent
robustness, strategy counterplay and seat equivalence remain **inconclusive**.
No observed ability frequency identifies a causal balance defect.

The sampled-menu audit independently found strategically distinct legal Spade
payments with only one represented, alongside declared restrictions in attack
subsets, formations, Code settings, trade subsets and Barricade sacrifices.
Policy comparisons share this generator. Raising resource bounds is a separately
named budget experiment; it cannot be reported as equivalent behavior when it
allows previously interrupted paths to proceed. Wider menu coverage requires its
own versioned controlled comparison.

| Claim at this development cutoff | Assessment | Basis |
|---|---|---|
| Equivalent optimization improves fixed-input runtime | Supported for the two named fixtures | Unchanged transcripts with 2.34×/2.07× measured run speedups |
| Original budgets reliably support this three-game evaluation | Refuted | Four four-player matches stopped at explicit transcript/menu limits |
| Opportunity policy implements the targeted local decisions | Supported for named fixtures | Legal-selection, retention, gift, promise and private own-finance tests |
| Candidate exceeds the practical +5-point/game threshold | Inconclusive | One three-player block without an interval; zero complete four-player blocks |
| Financial priority improves cross-game strategy | Inconclusive | No activation in the first development campaign |
| Seat effects lie within ±5 points/game and ±0.05 win share | Inconclusive | No homogeneous-policy confirmation cohort |
| Strategy advantages demonstrate healthy counterplay or a persistent rule imbalance | Inconclusive | No sufficiently diverse competent confirmation population or controlled mechanic variant |

Focused human-playtest questions remain whether proposal terms and refusals are
understandable, repeated offers create pressure or downtime, permanent alliances
offer meaningful choices against Coup opportunities, and players can anticipate
cash expiry and carried debt. These questions cannot be answered by deterministic
completion or bot score comparisons.

### Controlled larger-budget rerun

A new immutable development campaign reran the same exogenous first-block inputs
under **1,000,000 menu operations** and the named **32-MiB gameplay artifact
profile**, with four workers and unchanged three-game horizon, policies and sampled
menu. This is an operational-budget experiment, not new independent replication or
an equivalent-policy/availability claim for paths previously cut short.

| Larger-budget result | Three players | Four players |
|---|---:|---:|
| Planned / started / finalized games | 18 / 18 / 18 | 24 / 24 / 24 |
| Complete matches / paired blocks | 6 / 1 | 8 / 1 |
| Incomplete games / never-started slots | 0 / 0 | 0 / 0 |
| Candidate-minus-baseline points per game | +253/9 | −83/12 |
| Confidence interval | Unavailable: one block | Unavailable: one block |
| Completed replay credit | 18 | 0 |
| Run wall time | 77.747 s | 196.004 s |
| Replay wall time | 84.795 s, complete | 122.091 s, canceled |
| Run-directory bytes | 21,739,451 | 43,767,555 |

All 14 match transcripts preserve their prior committed prefixes. The six
three-player and four four-player matches that already completed have unchanged
full traces. All run and replay artifact checksums/lengths passed independent
inspection. Four-player replay reached the fixed wall allocation; its manifest
retains the source's finalized-game count, but a canceled verification receives
**zero completed replay credit**. No game was replaced or imputed.

The larger-budget campaign used **480.683 seconds** and **131,039,647 bytes** at
its recorded stop. Its binary hash is
`c17f367da6489f06c2b91399e107b13d737ef1ff90fcc7dafc4dce15af18c5b2`;
restricted evidence is `sims/artifacts/gameplay-evidence-v3-budget-development/`,
including immutable journals, run/replay manifests and analysis. The original
failed campaign remains separate. The favorable three-player and unfavorable
four-player development point estimates do not establish either practical strength
or weakness: each population still supplies only one independent development block.

The completion concern is **resolved for these fixed inputs under the named larger
budgets**, not proven absent throughout a confirmation population. This experiment
does not show whether increasing only one cap would suffice across all inputs.
It supports using the larger envelope for further development, while retaining
completion and resource stops as explicit confirmation outcomes.

### Additional development blocks and precision calibration

The fixed 550-second calibration allocation completed two additional development
blocks per population in **494.750 seconds**, retaining **135,302,088 bytes**.
All **36/48** planned games finalized in **12/16** complete three-game matches;
there were no incomplete games or never-started slots. No replays were selected
for this calibration run. Its schema-v3 source freeze and checksummed run manifests
are retained under `sims/artifacts/gameplay-evidence-v3-calibration/`.

Combining these unique blocks with the larger-budget development block, and
counting the repeated first block only once, gives **three independent development
blocks per population**, comprising **54/72** games in **18/24** matches. These are
compatible development runs with recorded build provenance, not a new confirmation
cohort. The untouched-confirmation sample size remains zero.

| Population | Exact per-game block effects | Development mean | Exact sample variance |
|---|---|---:|---:|
| Three players | 253/9, 218/9, −11/9 | 460/27 (17.037) | 61681/243 |
| Four players | −83/12, 199/12, 75/4 | 341/36 (9.472) | 21883/108 |

No confidence claim is assigned to these three-block development averages. The
change of sign across blocks cautions against selecting the favorable individual
game or run. The original historical cohorts, failed smaller-budget campaign and
reruns are not added to these independent denominators.

Using these development standard deviations, the prospective 1,000-block Student
halfwidth projection is **1.131/1.010 points per game** for primary-only K=2, and
**1.638/1.463** for the broader K=42 family, separately for three/four players.
At half the measured standard deviation, K=42 projections become **0.819/0.732**;
at twice the measured deviation they become **3.275/2.926**. These are planning
sensitivities from only three blocks, not confidence bounds on future precision or
power guarantees. Homogeneous-policy seat and robustness cohorts have not supplied
their own variance estimates. The 1,000-block floor is not reduced on this basis;
actual confirmation intervals, tail behavior and completion still govern claims.

The new protected confirmation wrapper, `xops/sim_confirmation.py`, authenticates
schema-v3 inputs and retained analysis sources before deriving contrasts. Exact
frozen profile identities bind primary K=2 versus roster K=42 before collection;
all-pressure baselines include the focal seat. It keeps population roots distinct,
reports unknown dispatch separately from never-started work, requires complete
planned blocks and completed replay/worker verification for supported claims, and
separates positive effects from practically meaningful improvement. The wrapper
itself must be included in the future source freeze; its availability is not a
claim that confirmation has run.

### Controlled early-retention ablation

The four-player retention ablation finalized **24/24 games in 8/8 three-game
matches**, one complete paired development block, in **191.782 seconds**, retaining
**44,364,676 bytes**. It compares focal `opportunity-no-retain@v1` against
`opportunity@v1`, with pressure/economic/pressure opponents, the same exogenous
block, rotations, one-million menu-operation cap and 32-MiB artifact profile.
No replay was selected. Its run manifest and artifact checksums were verified
under `sims/artifacts/gameplay-evidence-v3-retention/`.

Removing early royal retention improved the focal result by **20 points per game
in this block** (60 match-total points). The retained-policy baseline reproduces
the original candidate's exact scores and step counts in all four rotations;
selected action content and reasons also agree after normalizing job-derived
identifiers. This initially suggested a retention cost on the development input,
but the later tie-key defect below prevents clean attribution solely to retention.
There is no confidence interval with one block. The ablation is not pooled into
the primary opportunity-versus-economic effects above and does not constitute
fresh confirmation of a revised candidate. Formation-preservation benefits remain
fixture-supported while their net strategic value is empirically uncertain.

### Bargaining diagnostic and discovered comparison defect

A separate three-player diagnostic compared focal `bargaining@v1` with
`opportunity@v1` against two bargaining opponents. It completed **18/18 games in
6/6 three-game matches**, one development block, in **90.175 seconds**, retaining
**25,080,266 bytes**. Its per-game focal difference was **+4/3**, with no confidence
interval and no replay dispatch. Frozen live/copy analysis source hashes and all
run artifacts were independently verified before subsequent source development.
Restricted evidence: `sims/artifacts/gameplay-evidence-v3-bargaining/`.

No trade offer, acceptance or loan command was selected in this diagnostic, and
the carried-debt priority feature did not activate. Five uses of the separate
Negotiator ability are not five trade proposals. Consequently the score difference
cannot demonstrate the new bargaining feature's effectiveness. Actual generated
trade/loan regression fixtures exercise party-authorized terms and transfers;
normal-play negotiation competence remains empirically unestablished.

Investigation of the unexpected difference found an **implementation defect in
the legacy policy tie key**. On an `open-formation` reconfiguration, the key builder
replaced the supplied new formation specification with the existing record's old
specification. Different protection targets therefore received the same key.
Their remaining menu order depended on command hashes containing administrative
game identity. In the recorded paired games, first differing reconfigurations
occurred at steps **305, 456 and 1,824**, all referencing formations already
created earlier. The first pair selected Spade versus Club protection despite
identical relevant heuristic priority. This is not evidence of a canonical rule
imbalance or successful bargaining.

The defect affects interpretation of legacy policy comparisons, including the
retention ablation: the observed score contrasts are preserved, but isolated
causal attribution to the advertised feature is **inconclusive**. A correction
must preserve the requested reconfiguration in its key, retain frozen legacy
behavior for old artifacts, and use a named new policy behavior version. Fresh
development/confirmation under that version is required before treating feature
effects or the three-block precision projection as validated for the corrected
evaluator. Prior byte-equivalent performance measurements remain evidence about
the old fixed-input execution, not about corrected policy decisions.
The historical `economic@v1`/`pressure@v1` cohorts use the same legacy key and are
potentially affected too. Their original values, preregistered thresholds, sample
denominators and immutable artifacts remain unchanged; this later diagnostic
adds an interpretation limitation rather than rewriting their historical results.

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

### Final v3 status and verification

| Current claim | Assessment | Exact scope |
|---|---|---|
| Equivalent optimizations reduce runtime | Supported | Two unchanged fixed-input transcripts; 2.34×/2.07× run speedups |
| Outstanding historical replays are complete | Supported | Eight additional checks; historical totals 18/24, still three independent blocks each |
| Legacy action keys distinguish legal reconfigurations | Refuted | Two distinct accepted protection choices collide in v1 |
| Version 2 repairs that collision | Supported | Actual legal-menu regression, requested-spec preservation, named ID/menu-order and privacy fixtures |
| Candidate gains exceed +5 points/game | Inconclusive | Legacy contrasts are confounded; zero corrected-policy confirmation blocks |
| Multi-game financial strategy is stronger | Inconclusive | No selected new financial-feature activation in the primary development sample |
| Seats are equivalent within the declared tolerances | Inconclusive | Homogeneous-policy confirmation has not run |
| Strategies demonstrate healthy counterplay or a persistent rule imbalance | Inconclusive | No valid broad-roster confirmation or controlled rule-variant comparison |

Development execution used **1,577.707 seconds** of its amended 1,600-second cap.
The small remainder cannot fund a corrected full comparison. The separate historical
verification and profiling remain within their allocations; the old exhausted
campaign was never extended. The single measured confirmation allocation request
remains unanswered: up to 300 hours/230 GiB for the broader design or
100 hours/90 GiB for primary strength only. No confirmation roots were generated.
Canonical GAME_RULES is unchanged and no rule variant is recommended for adoption
from these confounded, development-only comparisons.

Two clean builds of the v2-policy simulator agree on SHA-256
`1f5bf861b5f41dfdd2a3a83acfe9a1fcdc558b3c1877f06ae934487c1484ecad`;
provenance and source snapshots are under
`sims/artifacts/build/gameplay-evidence-v3-policy-v2-20260930/`. This binary was not
used to relabel older game artifacts. Full Go tests passed (simulator 43.120 s),
`go vet ./...` passed, and targeted race checks passed (simulator 187.898 s),
including version/order, privacy, replay, continuation, cancellation and budgets.
Evidence: `sims/artifacts/verification-v2-final-20260930/`. All **67 Python**
analysis, provenance, denominator, source-freeze, confirmation and storage tests
passed in `evidence-v3-final-python--20260930T125819Z-2590919`. Independent
design, estimator, denominator, policy and conclusion reviews found no remaining
blocking implementation finding; this does not clear the unmet evidence targets.

A separate storage-monitor optimization returned identical readable-tree totals
on 29,579 files, reducing median scan time from 0.17924 s to 0.07815 s (2.29×).
It retries one whole scan on an atomic-rename missing-file race and discards the
partial total. Persistent scan errors stop/reap the child and record unavailable
storage rather than inventing zero. That stricter failure behavior is documented
separately from readable-tree equivalence; it is not a whole-campaign speed claim.
