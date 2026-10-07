# Complete accepted-gameplay evaluation v1 — preregistration

This plan is written before any held-out outcomes are inspected. The runner freezes
its exact inputs, binary, source/build provenance and analysis parameters in a
restricted campaign manifest. Development results never enter held-out estimates.
The existing ROADMAP remains the only implementation roadmap.

## Questions and estimands

Primary question, separately for three and four players: does replacing focal
`economic@v1` with `pressure@v1` improve final match score against the fixed mixed
lineup? Primary outcome is candidate minus baseline focal final match score,
averaged over every planned cyclic seat rotation within a seed block. The minimum
practically meaningful improvement is **5 points per game**; the desired 95%
interval half-width is **5 points**. These are proposed decision thresholds, not
measured results or established design requirements. A stronger-policy claim
requires the held-out interval lower bound above zero, at least the target block
count, satisfactory precision and no material incomplete selection; a practically
meaningful claim additionally requires its lower bound above 5. Test populations
separately; do not pool them to rescue an inconclusive result.

Secondary descriptive outcomes: legitimate board endings, financial finalization,
match completion, declaration type and seat, game score leaders, tied final ranks,
seat-specific results, rounds/turns, redeals, public ability opportunities/uses and
aggregate economic trajectories. Declaration and score leadership are distinct.
Raw use/win correlations are not causal estimates. No rule-variant treatment is
included. Candidate success against this lineup is not robustness to all opponents.

## Frozen design

Target **1,000 held-out seed blocks per population**, identifiers `heldout-0000`
through `heldout-0999`; each block has both treatments and all cyclic seat
rotations (six games at three players, eight at four). Thus 6,000 + 8,000 planned
single-game matches, not 1,000 games per block. Multi-game persistence is checked
separately using development integration runs; this primary comparison does not
estimate long-match strategy. Baseline focal p0 is economic; p1 pressure, p2
economic, and p3 pressure where present. Candidate changes p0 only to pressure.
This deliberately controls lineup composition but is a limited policy population.
Complete cyclic rotations balance seats, **not every opponent ordering**; no
opponent-order independence claim is allowed. Further lineups/random opponents
are development diagnostics and separately reported, never pooled here.

Development roots and identifiers are distinct from the private held-out root;
held-out seeds are never used for tuning. The final private profiles select
`sampled-all-categories-v2`, equal 100,000 deterministic policy operations per
decision, 100,000 menu operations, 10,000 transition limit per match, 100,000
settlement-step budget, four independent-match workers and one game per match.
The canonical 4 MiB transcript cap remains an explicit incomplete condition.
Policies, priorities and sampling are frozen through binary/source hashes. No
search or hidden-state inference is introduced. Every treatment uses the same
committed chance derivation for each block/rotation until choices diverge.

## Analysis and stopping

The sampling unit is a whole seed block. Use exact rational paired scores and
10,000 deterministic seed-block percentile bootstrap resamples for clustered
95% intervals (`python-mt19937-seed-block-percentile-v1`). One complete block has
no interval; few-block intervals are descriptive and unreliable. Incomplete
blocks stay in all planned denominators; complete-only estimates are explicitly
conditional and potentially biased. Report planned, started, board-ended,
financially finalized and never-started games and finalized matches separately,
plus every incomplete status and verified complete block count.

Replay each completed shard from recorded decisions/chance, without rerunning
bots. The target is at least 1,000 genuinely completed-game replays per population,
reported separately from historical Phase 0's 2,000 partial-prefix checks. The
first shard of each population is also rerun with one worker and compared with
four-worker outcomes and full transcripts, then replayed. Repeated worker runs
are determinism evidence, not additional independent samples.

Initial local allocation: **1,800 seconds cumulative campaign wall time and
2 GiB campaign storage**, including replay and worker checks. Development pilots
measured roughly 35–62 seconds per match serially (one or two games), around
1–2 MiB per raw transcript, with additional report/artifact overhead. This bounded
allocation is expected to be far below the full target; it tests execution and
produces a resumable honest sample, not permission to claim adequate power.
Population shards alternate in fixed order, one block per shard. Stop only for
these resource limits, explicit cancellation, or a recorded execution/verification
failure—not for favorable results. Preserve every planned unit and immutable
old outputs. A failed shard is retained, never silently retried or dropped.
Resume may continue later unaffected units with the exact same frozen inputs;
additional resource allocation requires a separately recorded continuation plan,
not changing the frozen plan silently. No result-dependent sample extension.

## Interpretation and next evidence

Report evidence-backed correctness findings, policy limitations and balance
hypotheses separately. Failure to meet targets is an explicit unmet target.
Automated play cannot establish enjoyment, clarity or social negotiation quality.
Focused human sessions should examine proposal comprehension, refusal/reputation
incentives, downtime under responses/settlement, and whether any observed passive
or repetitive strategy is enjoyable. Controlled paired ablations, broader lineups
and full permutations are follow-ups, not causal conclusions from usage counts.
