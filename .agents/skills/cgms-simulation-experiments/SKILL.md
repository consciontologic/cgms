---
name: cgms-simulation-experiments
description: "Specify, execute or interpret CGMS fair-bot experiments with paired seed blocks, complete seat schedules and honest incomplete-run statistics."
---

# CGMS simulation experiments

Use for S08–S09 bots, tournament/sweep/compare, and later strength evaluation.
Preparation inputs describe future runs; do not execute nonexistent commands.
Read [experiment research and procedure](../../../sims/docs/phase-0-experiment-research.md)
for estimands, bootstrap formula and limitations, and the
[format specification](../../../sims/docs/phase-0-format-spec.md) for input fields.

1. Freeze purpose, implemented coverage, interpretations, primary metric, lineups,
   candidate/baseline versions, deterministic operation budgets, seed split and
   full rotation schedule before running. Preserve effective input hashes. Separate
   three- and four-player populations and tuning versus held-out blocks.
2. Enforce fair observations AND action menus. Bot code sees no full state or future
   deck/seed. Test hidden-state permutations, mutation attempts and malicious
   actions independently of policy generation. Legal-random is uniform only over
   its declared canonical menu; a restricted menu cannot claim all-rules coverage.
   Record heuristic feature reasons and deterministic tie-breaking; exclude
   omniscient diagnostics from fair comparisons.
3. Pair treatments by exogenous seed, fixed opponent lineup and complete rotations.
   Cyclic rotations balance seats but not all opponent orders; enumerate full
   permutations for all-order claims. Keep persistent matches intact in one worker;
   independent replication is the seed block, not each game/seat/turn.
4. Report exact paired match-score differences over complete blocks and resample
   whole blocks for 95% intervals using the pinned analysis procedure. Record
   planned/started/finalized units, every stop reason, never-started jobs and
   per-population denominators. Zero complete blocks means no estimate. Never
   impute stopped games as draws/zero scores or replace inconvenient seeds.
5. Label completion-only estimates conditional and potentially selection-biased.
   Distinguish declaration outcomes, game score leaders and tied match ranks.
   Keep performance denominators and incomplete work visible. Compare compatible
   provenance; retain unmatched units rather than silently dropping them.
6. The CLI target remains at least 1,000 held-out blocks per player count and a
   favorable preregistered score-difference interval for stronger-play claims.
   Fixed-sample intervals do not authorize optional stopping or picking the best
   held-out candidate. A partial-rule pilot may satisfy useful Phase 0 reporting
   while supporting no full-game strength/balance claim.

Gate tests include exact two-block paired calculations, missing rotations, an
incomplete block with retained cause, 3/4-player separation, bootstrap index
vectors, worker-count equality and cancellation publication. Share only projected
reports; preserve restricted forensic traces and cited evidence under the
[CLI privacy/retention contract](../../../sims/README.md). Follow the
[execution guide](../../../sims/docs/phase-0-execution-guide.md) and
[Phase 0 instructions](../../instructions/PHASE_0_SIMULATION.md); this skill adds no
roadmap or authority to expand gameplay coverage.
