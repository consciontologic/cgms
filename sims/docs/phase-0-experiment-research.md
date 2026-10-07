# Phase 0 experiment and fair-bot research

Status: **preparation specification, not implemented or measured**. Access date for
all external sources below: **2026-09-30**. The [CLI contract](../README.md),
[GAME_RULES](../../docs/project/GAME_RULES.md), accepted
[Q/F/N/T register](../../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md), and
[S01–S10 roadmap](../../docs/planning/ROADMAP.md) take precedence. This document
explains experiment design for S08–S09, not a second implementation checklist.

## Evidence and alternatives

| Source (primary) | Supported finding | CGMS decision and limitation |
|---|---|---|
| [Lanctot et al., OpenSpiel, 2019, revised 2020](https://arxiv.org/abs/1908.09453), and [original `spiel.h` interface](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h) | An environment can distinguish world state, per-player observation/information state, legal actions, transitions and chance. It accommodates multiplayer and general-sum games. | Adopt the separation of responsibilities, not OpenSpiel as a dependency. CGMS decision stages, exact debt ledgers, acquisition table and partial coverage need explicit domain types. The linked `master` implementation is a dated inspection reference, not a pinned dependency or executable API promise. |
| [Glasserman and Yao, Common Random Numbers, 1992](https://pubsonline.informs.org/doi/abs/10.1287/mnsc.38.6.884) | Shared randomness is a variance-reduction method for simulation differences; advantageous variance depends on the compared systems, not simply identical seed labels. | Pair baseline/candidate by initial seed and seat schedule. Preserve stream identities and report empirical variation. Divergent policies can consume different draws; pairing is still a design choice, not proof of reduced variance or identical post-divergence card histories. Publisher abstract was accessible; no inaccessible theorem is claimed. |
| [Burch et al., AIVAT, original preprint](https://arxiv.org/abs/1612.06915) | A value estimator can reduce stochastic-game evaluation variance using heuristic values and known policies while retaining unbiasedness under its construction. | Defer AIVAT. Phase 0 uses raw exact match scores and transparent paired differences. CGMS is a partial multiplayer environment with no validated value estimator or full policy probabilities. Poker results cannot establish CGMS sample efficiency. |
| [Agarwal et al., Statistical Precipice, NeurIPS 2021](https://papers.nips.cc/paper/2021/file/f514cec81cb148559cf475e7426eed5e-Paper.pdf), [authors' implementation](https://github.com/google-research/rliable) | Point estimates obscure uncertainty; evaluation selection and aggregate summaries can mislead. The authors provide stratified-bootstrap interval methods and distribution summaries. | Report intervals and raw distributions; freeze tuning before held-out evaluation. Do not directly transplant task-stratified RL bootstraps: CGMS's correlated unit is the complete paired seed block, containing every rotation and match. No Python/RL dependency is needed. |
| [Cameron and Miller, Cluster-Robust Inference, 2015, author repository](https://escholarship.org/uc/item/1jq5d0pq) | Within-cluster dependence can invalidate independence-based precision; few clusters are a material limitation. | Resample whole independent seed blocks, preserving all outcomes inside each block. The source discusses regression inference, not a ready-made CGMS estimator; the concrete bootstrap and score estimand below are engineering choices. Small pilots remain descriptive. |

These sources motivate interface boundaries and statistical caution. Numerical
thresholds, rotation policy, privacy tests, output fields and the following
procedure are **CGMS engineering specifications** derived from the repository's
contract, not claims that the cited authors prescribe this particular game design.
No package installation or external framework is required by this preparation.

## Fair decision boundary

The engine owns adjudication. A bot receives a detached projected observation,
a stable legal-action/decision menu for its seat, public history, its own prior
observations and its own random stream. It receives neither a state pointer nor
callbacks allowing full-state inspection. Menus, rejection reasons and feature
explanations are part of the information boundary. Do not use a full-state legal
menu that happens to be accompanied by a redacted observation.

Use GAME_RULES §2.13 and F10 as the observation allow-list: public hand counts and
separate concealed-Ace counts are intentional; another player's concealed faces,
royal counts, eligible-target breakdowns, future random state and deck order are
not. Public historical identity does not authorize revealing a hidden card's
current location. F08 remote bound kings and N07 unreleased-card markers require
special regressions. T01 private offer terms belong only to the parties. Maintain
public active-Ace zones (N02) without exposing unrelated private cards.

For each supported decision stage, construct pairs of valid states differing only
in another player's unobservable cards/order/markers, preserving the observer's
entire permitted history and public counts. Compare serialized observation,
ordered menu, bot feature explanations and chosen action using the same bot
stream. Outcomes must be equal before a legitimately observable difference occurs.
Add an adversarial bot returning altered actor, physical card, stage, game ID or
proposal revision; the engine independently rejects it without effect. Attempt
mutation of the supplied observation and verify canonical state remains unchanged.
These tests are stronger than checking that a JSON field named `hand` is absent.
They are necessary evidence, not a proof covering untested information channels.

Legal-random samples uniformly over the versioned, canonical finite action menu
using the pinned unbiased bounded-integer sampler. Distinct physical choices remain
distinct; remove duplicate encodings of the same action. Uniformity over actions
is not uniformity over strategic categories. If exhaustive legal enumeration is
not implemented for a stage, report that coverage limitation; do not market a
restricted generator as uniform over all legal GAME_RULES actions. Preserve cheap
nonblocking offer-decline/withdraw paths and explicit settlement/finish decisions.
An empty menu is terminal only when the engine's phase says so; missing coverage
is unsupported, not an implicit pass, draw, refusal or finished settlement.

The first heuristic is transparent and versioned: use only projected immediate
score opportunities, exposed holdings, protection, known debt and legal declarations.
Record named feature contributions and deterministic tie-breaking. It may prefer
an immediately legal declaration or a safe opening, but may not simulate future
hidden cards from the omniscient state. Freeze feature definitions/weights before
evaluation; deterministic operation limits bound choices, while wall time is a
measured operational result. Omniscient diagnostics use an explicit separate mode
and cannot enter fair comparisons. Search, learned beliefs and LLMs are deferred;
no strength claim is needed to complete S08's initial heuristic.

## Experiment plan and paired blocks

Before execution, retain the versioned experiment input, its hash, purpose,
coverage profile, rules/interpretation versions, primary metric, candidate and
baseline bot configurations, fixed opponent lineup, budgets and exclusion policy.
Resolve overrides once and preserve the effective configuration. A sweep declares
its finite grid and maximum work before scheduling; it is exploratory unless a
single comparison was preregistered independently.

Keep **three-player and four-player populations separate** at every aggregate,
not merely in a footnote. Each block is identified by `(population, experiment,
lineup, seed-block-id)` and contains both baseline and candidate treatments and
all preregistered rotations. Replace the focal baseline with the focal candidate
while keeping the other seats' policies fixed. Rotations map bot identities to
fixed physical seats; initiative is still determined by Q07, never assigned by
this experiment design. All games within a match remain owned by one worker.

A complete cyclic rotation has `n` rotations of one ordered lineup: identity `j`
occupies physical seat `(j + r) mod n` for every `r` in `0..n-1`. This balances each
identity across seats but does **not** balance every neighbor/opponent order.
For heterogeneous opponents, prefer an explicitly enumerated full permutation
schedule (`3! = 6` or `4! = 24` assignments), retaining repeated policy identities
as distinct participant IDs. A cyclic-only experiment is valid only with its
fixed relative-order limitation stated; never label it all-order balanced. Store
the full expanded schedule, reject missing/duplicated rotation IDs, and pair the
same schedule across treatments and parameter values. Do not rotate seats between
games of a persistent match and accidentally change the underlying match rules.

Use independent held-out seed blocks, disjoint from development/tuning seeds.
Freeze the held-out schedule before execution and retain all scheduled identities,
including jobs never started. Domain-separate seed derivation for population,
block, rotation, match, game/replacement and bot streams, following the replay
specification. Counterpart treatments share exogenous seed identities; changing
the candidate must not change the opponent's initial stream seed through a whole
configuration hash. A manifest still records both different configuration hashes.
Within-match debt and score carryover creates dependence; a new match resets it.
No within-match game is an independent replicate.

`run` and ordinary `batch` configure fresh one-game research matches. Tournament
plans fix the visible number of finalized, non-nullified games before each match
(default three); nullified replacements get fresh game IDs and do not consume a
slot. Match score includes the actual §4.4 lifecycle and N06 adjustments. Track
ordinary threshold declarations, Coup outcomes, game score leaders and shared
match ranks separately. None is interchangeable with the others as “win rate.”

## Estimands, intervals and incomplete work

For a given population and fixed lineup let `B` be scheduled independent blocks,
`R` the preregistered rotation count, and `Y[t,b,r]` the focal player's exact final
match score for treatment `t` in `{candidate, baseline}`. Define `C[b] = 1` only
when both treatments have financially finalized all required matches/rotations
and verified their artifacts. Engine errors are never valid outcomes.

For a complete block:

```text
D[b] = (1/R) * sum_r (Y[candidate,b,r] - Y[baseline,b,r])
B_complete = sum_b C[b]
paired_mean = sum_{b:C[b]=1} D[b] / B_complete
block_completion = B_complete / B
```

Retain exact rational sums and counts; decimal formatting is presentation only.
When `B_complete = 0`, the estimate and interval are **not available**, never zero.
The primary paired summary weights each complete block equally. Do not silently
weight blocks by the number of games finishing or surviving rotations. Also report
per-seat differences and declared subgroup metrics without treating those extra
rows as independent replication or new confirmatory tests.

Engineering choice for the 95% interval: percentile bootstrap of complete seed
blocks, 10,000 resamples, with a separately pinned analysis stream and sampler.
For each replicate sample `B_complete` block indices with replacement; include each
selected block's complete treatment/rotation data and recompute the mean. Sort the
replicate means and take nearest-rank quantiles `ceil(0.025*K)` and `ceil(0.975*K)`
using one-based indexing, recording `K`, indices and analysis version. Cluster by
the shared seed identity even if reused across several predefined lineups; either
analyze lineups separately or preregister fixed lineup weights inside the block.
Do not bootstrap individual turns, players, rotations or games independently.
Intervals need enough independent blocks and representative sampling; degenerate
or tiny pilots cannot establish certainty. Report block count and raw values.

Every report shows the following counts separately for each population/treatment:

- Planned blocks, matches and logical game slots; started matches and actual game
  instances, including nullified replacements; finalized matches and games.
- Complete paired blocks and unpaired completions; unsupported-transition,
  adjudication-required, budget-exhausted, canceled, replay/invariant and I/O
  failures; jobs never started; exact reason/rule ID where available.
- Declaration rates with numerator and denominator explicitly named (for example,
  valid ordinary declarations divided by finalized games), plus completed/planned
  fractions. A match-rank statistic uses finalized matches, not game instances.

Completion-only estimates are conditional on completing **both** treatments.
If one policy reaches unimplemented rules or costly settlement more often, removal
is outcome-dependent selection. Do not impute incomplete scores as zero, count
stops as draws, replace hard seeds, or infer all-game strength from survivors.
Completion-rate differences are results in their own right. Since uncapped Ponzi
and unresolved future receipts offer no assumed universal score bound, do not
invent worst-case numeric intervals for missing final scores. Report missingness,
conditional estimates and the absence of an unconditional estimate. Deterministic
resume may complete the original units without changing identities; larger-budget
reruns form a new manifested analysis, preserve old failures, and apply the new
budget symmetrically. Error fixes invalidate old confirmatory runs for comparison
until the same frozen schedule is rerun with compatible versions.

The CLI contract's target of **at least 1,000 held-out seed blocks per player count**
and favorable preregistered score-difference interval remains intact for claims of
stronger play. It is not evidence that 1,000 guarantees useful power or that Phase
0 must implement all rules. Fixed-sample inference forbids repeatedly stopping
when the interval first excludes zero. Pick a candidate using tuning data, then
use a fresh held-out comparison; if many candidates are tested confirmatorily,
preregister a multiplicity procedure or make no family-wide significance claim.
Bots may be stronger against one fixed lineup and weaker elsewhere. Scores in a
partial implementation cannot prove game balance, human enjoyment, equilibrium,
commercial demand or parity with an offline client.

## Operational evidence and acceptance relationship

Schedule bounded whole-match workers; stream per-match artifacts rather than hold
all traces in memory. With the same logical schedule, outcomes and canonical
transition digests must match across worker counts; elapsed time is not part of
that equality. Record machine/platform/build, operations, p50/p95 decision latency,
games/second with an explicit finalized-game denominator, peak memory, aborts and
draw/redeal counts. Measure incomplete work too: good throughput obtained by
aborting difficult matches is not a performance improvement.

Cancellation and resource limits preserve committed per-match progress and known
completed units. Publish the status manifest atomically last, return the existing
CLI nonzero code for incomplete work, and preserve planned-but-unstarted units.
Comparison rejects incompatible provenance or unmatched pairs instead of joining
by accidental filesystem order. Markdown reports name the observer; private bot
transcripts and omniscient traces are separate restricted artifacts. Curated
inputs use synthetic cards. Retention follows the CLI contract's proposed 30-day
local policy, preserving cited evidence until explicitly released; no automatic
pruning of curated inputs or cited manifests.

S08 evidence consists of fair-observation/permutation tests, adversarial action
validation, deterministic policy decisions and honest coverage labels. S09 evidence
adds schedule validation, 3/4-player report fixtures, whole-block statistical test
vectors, zero/incomplete-block cases, deterministic worker counts and cancellation.
A statistical test vector should include two equal-size blocks with known exact
paired differences, an incomplete block excluded with its reason retained, and a
case where unequal completed rotations would change the answer if incorrectly
pooled. Test bootstrap resampling indices independently of the summary formatter.

Phase 0's exit remains one reproducible accepted-rule scenario plus a useful
batch/comparison report with all coverage and stop reasons. The broad rule matrix,
1,000-game replay target per supported player count and 1,000-block strength target
are recorded unfulfilled where their full prerequisites do not exist; they are
neither waived nor a reason to pull Phase 1 gameplay into this task. A comparison
of supported scenario quantities or completion behavior can be useful without
claiming full-game strength. All runtime/statistical tests, performance numbers,
fair-bot guarantees and host/container behavior remain unvalidated until actual
implementation executes them.
