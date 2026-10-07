# Gameplay evidence v3 — development allocation and prospective claims

Written 2026-09-30 before new confirmation collection. This extends evidence under
P03 in the existing roadmap; it neither changes accepted rules nor resumes the
exhausted v2 campaign. The old 5-point thresholds remain attached to old cohorts.

## Bounded development allocation

- Performance profiling, equivalence comparisons and throughput pilots: at most
  600 seconds execution wall time and 1 GiB generated artifacts.
- Historical eight-game four-player replay completion: at most 600 seconds and
  256 MiB additional artifacts, using the saved v2.1 binary. This adds verification,
  zero independent blocks and zero newly collected games.
- Candidate/ablation development pilots: at most 900 seconds and 2 GiB outputs,
  at most two policy development iterations. These are development only; all
  previously inspected results may inform development and cannot confirm it.
- Builds and regression checks are separate engineering work. Up to four workers;
  no purchased compute, system changes or unbounded background work.

Stop each allocation on elapsed time/storage or a recorded execution defect.
Preserve incomplete and never-started units; never replace difficult seeds.
Measure before requesting a single explicit confirmation time/storage allocation.
No confirmation collection is authorized by this development plan alone.

## Prospective claim definitions

All strength and economics experiments use fixed **three-game matches**, the
intended default horizon; average final match score divided by three is the
primary score unit. Compare focal candidates against unchanged frozen v1 policies
under the same menu generator and deterministic budgets. Equal policy-operation
caps do not make a changed menu an equivalent experiment.

Practical improvement: candidate-minus-baseline lower simultaneous confidence
bound exceeds **5 points per game**. Five points preserves the old meaningful
threshold and is one tenth of the canonical 50-point Coup award. A positive bound
below five supports only smaller improvement. A bound wholly below zero refutes
improvement for that named lineup; crossing zero remains inconclusive.

Seat fairness: identical policies, symmetric complete assignments; all pairwise
seat score differences must pass equivalence within **±5 points per game** and
seat match-win-share differences within **±0.05**. Shared first-place ties split
win credit equally. These are explicit product tolerances, not proven universal
balance standards. Failing to reject a zero difference does not pass equivalence.

Robustness: name each fixed opponent lineup and analyze it separately. Use full
permutations when testing independence from opponent order. Require practical
improvement on the designated primary lineup and noninferiority above **−5 points
per game** on every preregistered robustness lineup before claiming robustness.
Strategy dominance requires consistent advantage across the declared competent
roster; cyclic wins or matchup reversals indicate counterplay, not necessarily
imbalance. Raw ability frequencies are not causal evidence of rule defects.

Tactical and negotiation competence require rule-derived controlled fixtures,
observation/menu noninterference, legal chosen commands and inspectable feature
traces. Financial competence additionally requires completed three-game matches,
carried-debt behavior and controlled policy ablations; fixture success alone is
not an empirical strength claim.

## Confirmation freeze requirements

Before fresh confirmation seeds are used: freeze binary/source hashes, rules,
policy/menu versions, parameters, opponent roster, exact rotation schedules,
separate population roots, match horizon, budgets, outcomes, comparison family,
estimator and multiplicity correction, sample size, resource limits and stopping
rules. Target at least 1,000 independent untouched blocks per population;
development variance determines the projected precision/power, not observed
confirmation significance. Whole blocks contain all paired treatments/rotations.
Use a fixed sample; no favorable-result stopping or post-hoc cohort pooling.

Incomplete outcomes remain visible by treatment, with exact planned/started/
finalized/incomplete/never-started denominators. Complete-block estimates are
conditional and cannot establish unconditional strength when completion is
selective. Replay and worker reruns never increase independent sample counts.
Verify at least 1,000 completed-game replays per population plus worker equality.

Mechanic variants require a persistent effect across competent policies and a
controlled causal hypothesis. Any such variant gets a separate named/versioned
profile and paired experiment; canonical GAME_RULES remains unchanged.

## New-cohort inference implementation

`xops/sim_gameplay_evaluate.py` retains the frozen historical bootstrap unchanged.
The prospective helper takes one exact per-game contrast per independent block,
computes exact rational sample mean/variance, then uses a Student interval with
`B−1` degrees of freedom and critical probability `1−0.05/(2K)`. Here K is the
entire frozen comparison family, including both populations. These simultaneous
intervals use Bonferroni; no independence among comparisons is assumed. Inference
is exact under normal block means and asymptotic otherwise. Tail sensitivity and
completion must still be inspected; 1,000 blocks is not an automatic guarantee.

Equivalence requires strict interval inclusion within the prespecified bounds;
this is a conservative two-one-sided procedure at level `0.05/(2K)` per side.
Zero sample variance produces no confidence claim. Fewer than 1,000 independent
blocks remain descriptive under the default claim classifier. Known Student
quantiles at 1, 2, 10 and 999 degrees of freedom, exact mean/variance, multiplicity,
degenerate samples and non-significance-versus-equivalence have regression tests.
The exact family K and every named contrast must be frozen with the eventual
confirmation allocation before collecting confirmation outcomes.

Method references (accessed 2026-09-30): [NIST mean confidence limits](https://www.itl.nist.gov/div898/handbook/eda/section3/eda352.htm),
[Penn State simultaneous confidence intervals](https://online.stat.psu.edu/stat505/Lesson05),
and [Lakens, equivalence testing primer](https://pmc.ncbi.nlm.nih.gov/articles/PMC5502906/).
The tolerances, 1,000-block floor and conservative family design are CGMS choices.

## Frozen-menu audit

`TestSampledMenuAuditDistinctUsefulPayments` independently validates two legal
five-point Spade payments (one physical card versus two) and observes only one
in `sampled-all-categories-v2`. Remaining-card flexibility differs, so equal
payment total is not strategic equivalence. The sampler also chooses one physical
subset per attack total, prefixes for five-or-more queen/king formations, first
matching natural formation members, a restricted Code parameter grid, one-card
public trades and a fixed five-card Barricade sacrifice. These are declared
coverage restrictions, not uniform sampling over all accepted-rule actions.
Physical-ID ordering can select strategically different retained holdings.
No menu-generation behavior changes in this policy comparison; any expanded-menu
comparison must be separately versioned and paired under equal policy budgets.

## Development amendment after the first measured run

The first new development plan fixed two blocks/population, cyclic assignments,
three games/match, mixed pressure/economic opponents and opportunity versus economic.
It stopped at 320.32 seconds on a deterministic resource failure, not on its score.
The three-player first block finalized all 18 slots; the four-player first block
finalized 18/24 slots and 4/8 matches. Two matches reached the 4 MiB whole-match
transcript cap; two reached 100,000 menu candidates. The second block of each
population was never started. All artifacts and original plan remain immutable.

At that cutoff, remaining development allocation was 579.68 seconds from the original 900 seconds.
A symmetric rerun of the same development exogenous blocks will use explicit
larger menu and transcript budgets, with new binary/config hashes. This changes
the experimental completion envelope and is **not** an equivalence claim. It
changes no accepted rule or menu sampling. Old budgets/defaults remain available.
The deterministic budget failures must be repaired before a credible multi-game
confirmation allocation can be proposed. Financial feature activation and
completion are separate required diagnostics; positive completed-score differences
cannot excuse policy-dependent failures.

New runner schema `cgms-gameplay-campaign-v2` permits a frozen explicit list of
replay shard IDs. Confirmation must select a predetermined prefix containing at
least 1,000 game slots per population and keep worker-check shards selected.
Only actually completed and verified games earn credit; an incomplete selected
shard creates a shortfall, never a replacement chosen after observing results.
Historical schema-v1 continues replaying every shard unchanged. This reduces
large-campaign replay overhead while preserving each requested verification target.

## Proposed confirmation design (not yet authorized or started)

The resource request will offer a primary-only scope and the complete scoped
roster design below. Both use 1,000 fresh blocks per population, three-game matches,
fixed cyclic assignments, opportunity@v1 and frozen economic@v1/pressure@v1.
**Neither design claims independence from all opponent orders.** Full permutation
is a separately costed alternative requiring 6/24 rotations instead of 3/4.

The broader design has three fixed lineups per population, each with all 1,000
paired blocks. Mixed: baseline focal economic and alternating pressure/economic
opponents, exactly as the development template. Economic: every baseline seat is
economic. Pressure: every baseline seat is pressure. Each candidate treatment
replaces only focal p0 with opportunity. All comparisons use identical menu,
rules, budgets and match horizon. Share the same exogenous block identities
across lineups; do not count their correlated results as additional independent
blocks. Preserve all participant identities in seat schedules.

Primary strength: mixed-lineup improvement above 5 points/game in each population.
Robustness: candidate versus the named homogeneous policy must have lower bound
above −5 points/game in both homogeneous lineups. Also report each improvement
above 5 separately, without selecting the best lineup. Strategy performance may
be matchup-dependent; no equilibrium or human-strength interpretation.

Seat fairness uses only homogeneous baseline rows of the economic and pressure
cohorts. For each policy/population, all unordered physical-seat pairs yield
score/game and split-tie match-win-share contrasts. That is
`2 policies × (3+6 seat pairs) × 2 metrics  = 36` seat contrasts, plus 6 strength
contrasts, a fixed family of **K=42** simultaneous intervals across populations.
The primary-only option has just the two mixed-lineup strength contrasts (K=2)
and cannot supply confirmation of seat fairness or roster robustness.

At fixed 1,000 blocks, report development-based projected halfwidth and power with
sensitivity to development variance, and then actual interval widths. No sample
size is increased after inspecting confirmation outcomes. If a 5-point precision
goal is missed, the corresponding conclusion stays inconclusive. No replacement
seeds, imputed draws or favorable-result stopping. Development iteration cap is
two candidate versions; any revised candidate requires fresh confirmation roots.

Use shard size 1 or 2 only after measured worst-case aggregate artifact size confirms
replay's 256 MiB limit. Freeze an explicit first-shard schedule totaling at least
1,200 replay-eligible game slots/population (ceil rounding may exceed target) and
first-shard workers 1/4 checks. Distinct completed games in different lineups may count as replay verifications;
repeated replay/worker executions of the same planned game count once. Correlated
lineups never increase the independent seed-block count.
The larger budget/profile, final binary and all machine-readable input hashes
remain prerequisites to a final freeze. No confirmation seed is generated/read
until that freeze is ready.

## Prospective allocation transfer within the same total local cap

Before exhausting the new development allocation, unused verification/profiling
allowances are reassigned. The total new-task experimental envelope remains
**2,100 seconds and 3.25 GiB** (600 + 600 + 900 seconds originally), separate from the
immutable exhausted historical 1,600-second campaign. Profiling used approximately
215 seconds and retains a 350-second ceiling; historical verification used 110.421
seconds and retains a 150-second ceiling. Development now has a **1,600-second**
ceiling, including its 320.32 seconds already used. No new campaign allocation is
inferred from this transfer. Builds/regression tests remain engineering work.

This prospective transfer reserves enough bounded development work for additional
independent development blocks and a controlled retention ablation after the
larger-budget rerun. It does not change seeds or stopping rules of any already
started run. The active rerun retains its 480-second cap. No confirmation samples
have started; the large confirmation allocation still requires the one measured
resource decision requested by the owner.

## Measured resource request and next development inputs

The larger-budget development block completed all 18/24 games. Run collection
with four workers took 77.747/196.004 seconds; these are whole-shard wall times,
not CPU times to divide by four. Raw run artifacts occupied 21,739,451/43,767,555
bytes. Three-player replay took 84.795 seconds; four-player replay was canceled
after 122.091 seconds and receives zero verification credit. Sampled aggregate
process-tree RSS was 259,296 KiB; sampling started late, so this is not exact peak.

One-lineup cyclic 1,000-block collection projects 76.04 hours and 65.51 GB before
selected replays, manifests and contingency. The explicit owner request offers
100 hours/90 GiB for primary strength only, or 300 hours/230 GiB for the three-lineup
K=42 design. The latter projects 228.13 run hours and 196.52 GB before overhead.
These extrapolate one development block and may understate tail costs. Caps stop
collection without replacing incomplete seeds; neither option guarantees precision
or a positive conclusion. At request time about 294 GB was free. Full permutations
for one lineup project 369.87 hours and 306.08 GB raw, already exceeding available
storage; they are not silently substituted or authorized.

The remaining local development allocation uses two additional untouched
development blocks per population (84 slots), a 550-second fixed cap and no replay
dispatch, followed by a retention ablation only within the residual 1,600-second
development ceiling. These are development, never confirmation. Campaign schema-v3
freezes and retains analysis source bytes, checks them before dispatch/analysis
and rejects drift on resume; schemas v1/v2 preserve their historical behavior.

The broader design runs three separately frozen campaign directories, each with
one profile per population, under the single aggregate resource cap and a frozen
lineup execution order. Homogeneous pressure means every baseline participant,
including focal p0, uses pressure; replacing only p0 creates its candidate arm.
The bargaining policy is fixture-tested but outside this confirmation roster.
No evaluated population is called universally competent. Homogeneous-lineup
throughput has not been measured; the contingency and strict cap remain necessary.

Final confirmation source freeze must also include the executable aggregation and
claim-report wrapper that fixes K, named comparisons, score normalization and
classification. It is insufficient to freeze only low-level estimators and then
select claims manually after outcomes. Development summaries may use inspected
diagnostic calculations, but those cannot serve as confirmatory analysis.

Calibration completed all 84 planned games in 494.750 seconds, leaving 304.247
seconds under the existing development ceiling. Before its first dispatch, the
four-player retention ablation was assigned a 300-second cap (replacing the
unstarted 248-second draft). It compares opportunity-no-retain against opportunity
with the same block-zero inputs, pressure/economic opponents, menus, budgets and
three-game horizon: 24 planned slots, no replays. This reuses a development block
for a different controlled contrast, not a fourth independent primary block.

Final machine-readable confirmation profiles bind scope before collection using
IDs `confirmation-v3-primary-mixed-{3,4}p` or
`confirmation-v3-roster-{mixed,economic,pressure}-{3,4}p`. The aggregation wrapper
rejects a scope that differs from these frozen IDs, so its CLI cannot select a
smaller comparison family after seeing results. Its source belongs in
`analysis_sources` alongside the runner and claim extractor.

The retention ablation completed in 191.782 seconds, bringing actual development
execution to 1,487.532 seconds. Before dispatch, the remaining allowance assigns
105 seconds to a separate three-player bargaining diagnostic: opportunity versus
bargaining focal policy with two bargaining opponents, one reused development
block, three-game matches and 18 planned slots, no replay. This exercises the
second distinct policy candidate under the existing 1,600-second ceiling; it is
not pooled with the mixed-lineup primary or retention comparisons.

The resource scan also measured 29,579 existing files: 0.175–0.197 seconds per
full scan. The current monitor scans and then sleeps 0.2 seconds; it is not a
fixed five scans/second. A single-stat traversal prototype returned identical byte
totals in 0.080–0.092 seconds. The source change is made only after all active
frozen development runs close, with equivalence/error-path tests; this is not
a measured whole-campaign speedup. Retained immutable journals are quadratic in
unit count: an intentionally conservative 8,000 × 1.37 MB populated-snapshot
scenario adds roughly 11 GB per 2,000-shard campaign. Actual early snapshots are
smaller. Time/storage caps include this overhead and may prevent target attainment.

## Prospective correction after discovering treatment-sensitive tie-breaking

The normal bargaining diagnostic completed 18/18 games but selected no trades or
loans. Its score difference led to a concrete defect: the v1 semantic action key
replaced a requested formation reconfiguration with the existing formation spec.
Distinct protection choices therefore collapsed to the same key; menu ordering
that includes administrative identities could choose different actions across
otherwise equivalent treatments. All prior numbers remain preserved, but neither
the retention contrast nor the primary development score changes isolate their
intended policy changes reliably. They are not confirmation evidence.

The correction is prospective policy version **v2**, preserving all v1 behavior
and pinned artifacts. Economic, pressure, opportunity, both opportunity ablations
and bargaining receive the same corrected v2 ordering contract. Future confirmation
uses only those v2 versions; earlier references to v1 describe the original draft
and development runs. The confirmation wrapper enforces v2 before any collection.
Rules, sampled-menu generation and deterministic operation budgets remain fixed.
Regression fixtures test distinct requested reconfigurations, administrative-ID
renaming and menu-order invariance. This is a policy behavior change, not an
equivalent optimization. Old development variance is only a rough planning
sensitivity; it does not establish precision for the corrected policies.

The existing allocation question remains pending; no fresh confirmation roots or
results exist. The development execution allowance has only about 22 seconds left,
so a full corrected-policy comparison cannot be silently added. Implementation
regression checks are separate engineering work, not new population evidence.

Protected inspection briefly printed plan metadata to the operator tool transcript.
That transcript must retain the same restricted handling as forensic artifacts.
Actual private root values are absent from source and shareable report files;
the privacy scan checks those values directly rather than relying on filenames.

The older roughly 50.5-hour run-only projection concerned 14,000 one-game
slots. The new roughly 76-hour single-lineup projection concerns 42,000 games
within three-game financial matches. These are different workloads: the new
projection does not rerun the unchanged old campaign or negate the measured
fixed-input speedup. Both current resource projections were measured with v1
policies after performance changes; v2 trajectories may alter costs, and the
explicit caps are stopping limits rather than completion guarantees.

## Conditional power sensitivity, not validated v2 power

For a fixed 1,000 blocks, a normal approximation to the probability that the
simultaneous lower bound exceeds five is
`Phi((true_effect - 5) * sqrt(1000) / SD - t_critical)`. Using the three-block v1
standard deviations only as provisional planning inputs gives:

| Family and assumed SD | Power if true effect is +6/game, 3p / 4p | Power if true effect is +10/game, 3p / 4p |
|---|---:|---:|
| K=2, measured SD | 39.7% / 49.1% | >99.9% / >99.9% |
| K=42, measured SD | 10.3% / 15.2% | >99.9% / >99.9% |
| K=42, twice measured SD | 1.2% / 1.6% | 95.7% / 98.9% |

These hypothetical alternatives were chosen for prospective sensitivity, not
because they are established effects. Three clusters, the v1 attribution defect,
untested v2 trajectories and unmeasured homogeneous-lineup variance prevent treating
these as validated campaign power. They demonstrate why 1,000 blocks alone do not
guarantee a practical-strength conclusion, especially close to the +5 threshold.
Any approved allocation must reserve a bounded corrected-policy development check
before freezing confirmation roots; its actual variance and costs replace this
provisional input prospectively, without examining confirmation outcomes.
