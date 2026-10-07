# Gameplay bot and menu specifications

Status: implementation/development specifications, not a held-out strength finding.
The original Phase 0 `narrow-single-pair-combat-v1` observations and partial-prefix
results retain their historical meaning. These policies use a separate
`sampled-all-categories-v2` menu and `authorized-finance-v1` financial interface.
The [existing roadmap](../../docs/planning/ROADMAP.md) remains the scope authority.

## Information and deterministic choices

Policies receive detached authorized observations, never engine state, draw order,
random outcomes, seeds, opponent concealed cards or unrelated financial records.
Known exposure history preserves faces, not their current concealed locations.
Loan parties see historical loan identities without remote current custody.
Financial seats are one-based at the policy boundary; nested canonical debt and
promise records retain their explicitly documented zero-based indices.

Candidate legality is independently checked by shared engine transitions. An
exhausted candidate or policy-operation budget returns an explicit error and no
fallback pass. Empty legal menus remain unsupported/waiting conditions; only a
selected `end-turn`, response `pass`, refusal or settlement finish records that
intent. Automatic financial work blocks discretionary choices. Adapter-supplied
random outcomes are independently recorded and replayed.

Policy sorting excludes administrative game/command/window identities. Semantic
action hashes resolve ties; a changed game ID alone must not alter choices. All
policy versions use the same candidate interface and operation accounting. Random
selection charges each inspected menu entry plus rejection-sampling words;
heuristics charge one inspection and one evaluation per menu entry. Candidate
generation has a separate explicit operation cap (runner default: 100,000).

## Actual menu sampling

This finite policy menu is **not exhaustive and not uniform over all legal physical
combinations**. An explicit engine command may select other legal combinations.
The `sampled-all-categories-v2` profile currently includes:

- One deterministic physical subset per achievable exact number-card total,
  computed by bounded integer-scaled dynamic programming; whole target series
  additionally supplied for Baron. No floating-point value comparisons.
- Natural formation templates and representative fixed Doppelganger substitutions
  for Fate, Baron, People, Justice and Kidnapper. Underground/Code use sorted
  natural-card prefixes of every eligible size; substituted Underground/Code
  representations are not sampled by this version.
- Supported series openings, royal attachments, formation openings, taking back,
  People target changes, ordinary/Baron/Bomb attacks and Infiltrator/Exile flags.
  Numerical Ace and Clubs Advancement modifiers are combined with available
  Infiltrator/Exile flags for ordinary, Baron and Bomb attacks. Ordinary matching
  and Baron superiority remain adjudicated by the engine; physical subset
  alternatives remain sampled, not exhaustive.
- Main abilities, legal defense responses, required exposed choices and abstract
  hidden Justice selection. Hidden outcome identities are resolved by the adapter
  after the policy commits to that authorized selection, never inspected by bots.
- Affordable deck-purchase quantities; a deterministic five-other-card Barricade
  sacrifice representative; Code diamond bonuses 1, 5 and 10 and prices 5 and 10.
- Named-party proposal accept/decline/withdraw, whole-formation loans, own one-card
  gifts, exchanges requesting only already-visible opponent cards, and intact or
  individual controlled loan returns. No guessed hidden identities or automatic
  offer acceptance. Arbitrary multi-card negotiations remain available through
  explicit engine commands but are not exhaustively sampled.
- Immediate valid Coup/ordinary declarations and explicit turn completion.

Finite sampled actions can miss valuable combinations. Coverage of an engine
fixture does not imply that an autonomous policy will discover that interaction.

## Policy versions

| Policy | Board priorities | Financial behavior |
|---|---|---|
| `gameplay-random@v1` | Uniform over canonical distinct entries of the finite supplied menu | Uniform explicit choice among authorized financial operations |
| `economic@v1` | Immediate ending; support/history; Underground; guaranteed Ringleader; useful combos/attachments; acquisitions; purchases early; conservative ordinary attacks | Honors affordable triggered commitments, offers an exact reduced amount when payable, explicitly answers reduced offers, finishes only after required decisions resolve |
| `pressure@v1` | Same legal foundations, higher ordinary attack/destruction/assassination/Fate/Inflation priority | Explicit opportunistic refusal preference preserves spendable points at the canonical reputation cost |

Within the complete-game adapter, historical `legal-random@v1` and `heuristic@v1`
names map to these new menu-aware random and economic implementations. The menu
version and engine source hash must therefore accompany a policy name; the same
name is not evidence of equivalence to the historical Phase 0 policy population.

Concrete board priorities are fixed in
[`gameplay_policy.go`](../../backend/internal/bots/gameplay_policy.go). Immediate
legal endings score 100,000; ordinary series opening 100 (pressure Clubs: 130);
Underground opening 160; Ringleader 85; ordinary formation opening 60; attachment
40 (king: 25); base ordinary attack 15 + 3 per target (pressure: +40); unknown
proposal creation −50; unchanged reconfiguration/takeback −100; explicit turn end
0. Other exact priorities and resource adjustments are source-defined and must be
hashed/frozen with the experiment. These are transparent initial heuristics, not
learned or proven optimal values.

Heuristics avoid placing a second king where it disables Richer income, avoid
unchanged reconfiguration loops and prefer cheaper Diamond assassination payment.
They do not perform lookahead, infer hidden hands, model beliefs or establish a
strategic advantage from apparent usage correlations. Board policy reasons
explicitly acknowledge unavailable financial context; financial choices separately
use own exact cash and party-only debts.

## Financial decisions and limits

The financial menu covers initial incoming promise acceptance, exact reduced final
offers, final accept/reject, actual payment, explicit refusal, creditor-authorized
forgiveness, voluntary transfers and settlement finish. It samples 40% prospective
ordinary/Coup award promises and representative reduced amounts (zero, 40% and
available cash). Acceptance is not payment. Existing debts are paid first;
forgiveness creates no cash. The economic policy avoids offering even a zero
voluntary payment when outstanding debt makes the payment interface unavailable.

No policy manufactures unanimous system-debt forgiveness consent, declares another
seat's intent or interprets absence as refusal. System forgiveness starts with a debtor request. Its exact requested debt identity,
amount and debtor are then visible to all participants; unrelated system debts
remain private. Each participant records only their own consent, and no debt is
forgiven before every required participant agrees. Initial promise rejection has no
separate engine operation; a recipient may leave an optional unaccepted offer
unaccepted. The runner solicits optional financial decisions during play and records explicit
declines without changing game state. Equivalent authorized observations are not
solicited again until their financial information or consent requests change.
Departure choices and nullification consent are also explicit participant decisions.
Implemented menus alone do not demonstrate autonomous promise creation; the latest
development pilots declined the available optional financial actions.

## Verification versus evaluation

Targeted tests exercise immediate endings, distinct economic/pressure choices,
legal defense, identity-independent tie-breaking, exact affordable/reduced promise
payment, automatic waits, explicit refusal, absence of invented system consent,
party-only financial privacy, hidden-card observation equivalence and detached
loan metadata. Captured red/green regression logs remain under `/tmp/agent-runs/`;
the coordinating evidence report records retained run references.

No stronger-bot or balance claim follows from these tests. Freeze actual source
hashes, policy parameters, menus, budgets, opponents, seed splits and analysis
before held-out execution. Report loops, unsupported menus, budget exhaustion and
incomplete matches in the planned denominator.

## Development menu optimization evidence

A fixed four-player synthetic development position with 200 prior command receipts
was CPU-profiled before and after caching repeated rejected candidates and final
sort hashes. Candidate generation, adjudication, budget counting, policy parameters
and sampling did not change. The engine still checks complete original state and
receipt history: no state-budget or idempotency checks were removed.

Three benchmark iterations measured 316.90 ms, 242.69 MB and 2,969,040 allocations
per menu before; 274.06 ms, 206.46 MB and 2,387,738 allocations afterward. This is
approximately 13.5% lower time, 14.9% fewer allocated bytes and 19.6% fewer
allocations on this fixture, not a game-throughput confidence interval. Canonical
serialization inside independently adjudicated candidates remains the dominant
cost. Captured pre-optimization three/four-player observations are protected by
`TestGameplayMenuOptimizationGolden`; candidate-budget behavior is unchanged.

Two isolated copies of the same development source, differing only in that menu
cache optimization, then executed and replayed the same mixed economic/pressure
normal-deal development games. Both games financially finalized. The complete
private transcripts were byte-identical: 452 transitions for three players and
832 for four players, including decisions, random outcomes and every digest.
Generated evidence remains ignored under
`sims/artifacts/dev-menu-profile/` and `sims/artifacts/dev-menu-equivalence/`.
These are development equivalence checks, not held-out evaluation units.

The later v2 modifier expansion preserves the historical v1 optimization result
above. Reviewed three/four-player development captures retained exactly 160/174
legal entries respectively, with no added or removed commands and no observation
change except `menu_coverage`: those fixtures contain no concealed Ace or attached
attack modifiers. Their golden checks retain the original v1 hashes after version
normalization as well as the full v2 hashes; separate modifier fixtures exercise
the expanded command combinations. This is interface regression evidence, not
held-out gameplay evidence.

The development fixture `TestGameplayPoliciesPreferAvailableFormationToAttachment`
checks an actually legal Kidnapper against competing attachment choices. Both
heuristics select the formation. With an unpaired royal they immediately attach
it instead of retaining an option for a future combo; this myopic priority can
suppress later formations. Three DEV3 games contained no formation openings.
That is an observed limitation, not evidence that forming combos is illegal or
that combos are intrinsically weak. No held-out tuning addressed it.

## Development random financial policy v2

`gameplay-random@v2` preserves v1 board-action sampling. Financial selection first
samples uniformly among the available operation **kinds**, sorted by name, then
uniformly among the semantically sorted legal entries of that kind. With `K`
kinds and `n_k` entries in kind `k`, an entry has probability `1/(K*n_k)`.
Explicit decline/wait choices participate only when present in the authorized
menu; missing choices are never manufactured. A growing collection of promise
offers therefore cannot reduce another available kind's class probability.

Operations charge the existing authorization scan, one grouping operation per
entry, and every rejection-sampling word for both selections. Insufficient budget
returns exhaustion without choosing a fallback. `gameplay-random@v1` retains its
original uniform-per-entry financial policy for historical replay. This change
responds to development-only promise churn; it changes neither frozen economic
nor pressure parameters and has no held-out strength claim.

## Financial diagnostic privacy

`games/<unit>/finance-private.json` uses
`cgms-gameplay-finance-private-v2`, `privacy: restricted` and
`observer: omniscient`. It retains detached exact rational ledger snapshots at
deal, round, financial-finalization and retained-checkpoint boundaries, including
party-level charges, debts, corrections and totals. Its closed shape is generated
as `config/schemas/finance-private.schema.json`; ledger parties use canonical
zero-based indices. It is forensic output, never a policy input or a shareable
report. Replay verifies the snapshot projection against the recorded transcript.

Public `cgms-gameplay-public-v2` output omits these trajectories entirely. Partial
outcomes also omit current cumulative scores/ranks because those can reveal
unfinalized private charges; completed game results remain public. Cumulative
scores and ranks appear only at a financially finalized boundary. Original v1
artifacts remain immutable and restricted and must not be published as though
they had the v2 privacy projection.

## Frozen-policy competence fixtures

`TestFrozenHeuristicsRemoveVisibleImmediateDiamondOpportunity` constructs an
actual engine menu with an opponent's 13 exposed number Diamonds and four-suit
history, an exact affordable attack removing one Diamond, an equally legal
unimportant Spade target, and End turn. Both frozen `economic@v1` and
`pressure@v1` select the Diamond attack. The test adjudicates Club usage and all
explicit response passes, then verifies the opponent holds 12 Diamonds and the
attacker acquired one. This controlled result follows the existing Diamond
priority; it does not establish general threat recognition or defense against
arbitrary immediate losses.

The matching competence crosswalk is:

- Immediate own ending and explicit operation exhaustion:
  `TestGameplayPolicyImmediateWinAndExplicitBudget`.
- Legal response defense and distinct policy priorities:
  `TestGameplayPoliciesDistinctAndLegalDefense`.
- Actual menu resource allocation toward an available formation, alongside the
  known unpaired-royal reservation limitation: `gameplay_formation_test.go`.
- Affordable triggered promise payment, reduced final offers and actual payment:
  `TestGameplayFinancePaysAffordableTriggeredPromise` and
  `TestGameplayFinanceReducedOfferAnswerAndActualPayment`.
- Existing debt consequences and no unpayable zero offer:
  `TestGameplayFinanceDoesNotOfferUnpayableZeroWhileInDebt`.
- Consent and privacy limits:
  `TestGameplayFinanceNoFabricatedSystemConsent` and
  `TestGameplayFinancePartyOnlyPrivacyAndExplicitRefusal`.

These are targeted competence checks, separate from held-out comparative results.
No policy parameter was changed to pass the visible-loss fixture.

Private financial diagnostic v2 retains a deterministic snapshot prefix under a
2 MiB diagnostic byte limit. `status: incomplete`,
`reason: diagnostic-byte-budget`, and `omitted_snapshots` identify excluded later
snapshots. This bounds duplicated forensic history only: authoritative ledger
obligations, exact values, gameplay checkpoint and settlement work are unchanged.
Replay reproduces both the retained prefix and omitted count.

## Opportunity development policies and controlled ablations

The added `opportunity@v1` is a prospective development candidate, not a proven
stronger policy. Frozen `economic@v1` and `pressure@v1` retain their original
scoring, financial choices and deterministic budget accounting. Current candidate
choices were motivated by the inspected historical development limitations; those
historical results are not independent confirmation of this candidate.

The candidate starts from economic priorities and changes these features:

- Reserve an otherwise attachable jack or queen through round nine (priority −2,
  below End turn), then release it from round ten. Kings retain the baseline's
  single-king income behavior. Already available formations remain preferred;
  retention is a heuristic for future formation opportunity, not a prediction of
  an unseen draw or a demonstrated optimal cutoff.
- Value a legal deck purchase at `4 + quantity`, including late purchases: number
  Spades have no direct ordinary end award. This favors more card opportunities,
  but does not model the cost of losing series support or all counterplay.
- Accept an ordinary free gift (priority 90) when the authorized named-party terms
  request no return. Do not treat an intact formation loan as a free gift: its
  permanent exclusive alliance disables Coup and may change Ponzi allocation.
- Prefer the largest affordable reduced offer **within the same promise**, after
  checking own older debts. This preserves payment's priority over negotiation:
  offered amounts are ranked only against the same promise's sampled amounts.
  Acceptance is still explicit; only actual payment counts toward reputation.
- Receive detached scalar own available cash, total outstanding own debt and public
  number of future games, excluding the current game. If future games remain and
  debt exceeds cash, add priorities 40 to Ringleader, 30 to Ponzi and 25 to opening
  Underground. This explicitly values earlier receipts that can discharge debt
  which would otherwise carry into later games. Cash expires at game closure;
  debt does not. This is a bounded heuristic, not a solved intertemporal model.

`opportunity-no-retain@v1` disables only the early jack/queen retention feature.
`opportunity-no-finance@v1` disables the **financial bundle**: contextual carried-debt
income bonuses and the new reduced-offer amount ranking. It retains the original
economic policy's affordability/debt legality checks, payments, explicit refusal
when payment is blocked, and proposal-churn avoidance. Neither ablation changes
the menu, accepted rules, seat schedule, exogenous draws or deterministic budget.
These names are separate immutable experimental treatments; compare them under
the same menu generator and budget.

The economy adapter receives the private match but exports only those authorized
scalars to the bot. The bot receives no opponent ledger, receivable list, root
seed, deck order or future outcome. Explicit simulator policy callbacks remain
honored even with opportunity participant names. Candidate reasons disclose
feature activation without printing private amounts. Finance ranking uses exact
rational comparisons. As with frozen policies, the charged policy operation is
one authorization inspection plus one action evaluation per menu entry, **not**
one CPU instruction or one rational comparison; ranking scans other legal amounts
inside an evaluation. Menu generation is a separate budget. Runtime profiling
must measure this work rather than infer equal CPU cost from equal operation counts.

### Negotiation and menu audit boundaries

The current generator supplies own single-card gifts, trades requesting exposed
opponent cards and whole exposed formation loans; it cannot propose a useful
exchange for an opponent's unknown hand card. It does not enumerate arbitrary
multi-card offers or arbitrary combo substitutions. These are menu limitations,
separate from policy evaluation. The candidate continues to decline valued
exchanges and loans absent a reciprocity/alliance model; it does not initiate
speculative offers. Thus a population of these policies will usually have no
voluntary negotiation, even where the menu contains a legal exchange. That is a
policy-population weakness, not evidence that accepted negotiation rules lack
benefit. Free-gift acceptance tests do not establish bargaining competence.

Formation priorities still group most combos at 60 and favor Underground at 160.
They do not simulate the next action, compare synergistic combo sequences or price
Coup exclusion against an incoming loan. Royal reservation makes one opportunity
cost explicit; it does not remedy every formation valuation gap. A future menu
change must be independently versioned and compared under fixed policies, because
changes to available physical subsets can otherwise masquerade as stronger bots.
No rule variant is justified by these fixture results alone.

### Verification and interpretation

`gameplay_opportunity_test.go` in bots exercises actual-menu early/late royal
retention with a same-menu ablation; free ordinary gifts versus permanent loans;
exact affordable reduced amounts; prior-debt refusal; late quantity preferences;
hidden-hand permutation invariance of observation, legal menu and decisions;
observation immutability; explicit budget exhaustion; party-only finance; and
future-game-dependent income priorities. The frozen visible-Diamond-threat fixture
also covers all three opportunity names. Simulator fixtures check scalar horizon
and debt projection, invariance to unrelated private debts, and preserved explicit
policy callbacks. Many unrelated promises cannot elevate offer ranking above an
existing payment decision.

The initial candidate tests failed before registration/implementation. A later
red regression exposed an explicit-callback bypass, and another exposed unrelated
promise multiplicity affecting final-offer ranking; both were repaired before
confirmation. Full bots/sim tests and vet passed before these small review fixes;
post-fix targeted race tests passed in
`opportunity-review-green--20260930T120123Z-2475477`. These checks support the stated
local decisions and privacy boundaries. They are not population strength,
negotiation quality, game balance or human-enjoyment evidence. Multi-game strategy
requires intact multi-game matches and independent seed-block comparisons.

## Explicit menu-generation allocations

Experiments may set `budgets.max_menu_operations_per_decision` to a positive
integer through 10,000,000. Omission or zero preserves the historical 100,000 cap
and omitted-field serialized form. The cap is forwarded to new matches and
continuations; it is separate from `max_bot_operations_per_decision`. Negative,
null or oversized values fail input validation. A larger cap does not change the
sampler or accepted actions when both allocations finish generating the menu,
but can make previously budget-stopped trajectories executable. It therefore
changes the experimental resource profile and completion population; it is not
an outcome-equivalence claim. Apply the same cap to every compared policy and
retain earlier incomplete units. A prospective 1,000,000 cap must be frozen in the
new campaign input before collection, rather than retroactively relabeling the
100,000-cap pilot complete.

## Bargaining development opponent v1

`bargaining@v1` is a separate prospective policy, leaving the opportunity versions
unchanged. It inherits opportunity's card, retention and own-economy priorities,
then evaluates ordinary proposals with an inspectable marginal-card utility.
This is not an expected-score estimator: number Diamonds use their printed rank
plus the current own ordinary-ending bonus; number Spades use `rank/2 + 1`;
other numbers use 1; Aces use 5; royals use 3, or 10 for Spades, plus 2 for kings.
A same-rank/suit twin already in own hand or attachment adds 12. These are declared
heuristic weights; even rule-derived terminal values are contingent on retention
and an ordinary ending. The utility difference must be positive to offer/accept.
Positive trades have priority `120 + min(difference,25)`.

The bot initiates at most one proposal per current turn, checking all its
party-visible proposal records including declined or withdrawn records. It never
initiates a gift or outgoing loan and does not manufacture reciprocal consent or
a promise. This bounds proposal churn; other legal actions remain available.
An incoming gift or ordinary trade is evaluated from its exact authenticated
terms. Named parties are authorized to know offered faces under GAME_RULES §2.13
and §2.15: the bot maps only those offered physical IDs through the static 104-card
universe. The lookup adds no custody, availability, opponent inventory or future
information. It requires matching game, turn, offer ID, revision, offered status
and recipient; unrelated or stale proposals cannot authorize that lookup.

Incoming natural Underground loans receive a fixed positive utility priority of
145 for their lasting privilege. Natural Kidnapper loans receive that priority
only with two own current series including Clubs. The bot declines other loan
kinds, substituted formations, or another incoming loan when an own loan record
already exists. It also declines a loan while holding at least three of the four
real Hearts/Diamonds Coup queens, explicitly reserving the near-Coup option that
permanent alliance would remove. This is a conservative alliance heuristic:
it does not model partner loyalty, full Ponzi consequences, early-game Coup
probability or bargaining equilibrium. No outgoing-loan benefit model is claimed.

Real generated-menu fixtures establish a mutually beneficial Club/Spade jack
exchange, including valuation of the privately offered face by its named
recipient, legal acceptance and complete physical transfer after explicit
responses. They cover same-turn proposal suppression after offered/declined/
withdrawn states and a supported Kidnapper loan versus near-Coup refusal.
Privacy tests change unrelated hidden cards and require identical observation,
menu and choice; outsiders receive no hidden offered face or private proposal.
Party/game/turn/revision mismatches are rejected, observation mutation is checked,
and insufficient operation budgets remain errors. The visible-Diamond-threat
fixture also covers bargaining. These are local competence checks; effectiveness
against the other policies requires separate fixed-seed-block experiments.

The expanded-menu continuation claim concerns the CLI's authenticated parent
execution limits and `continueFullGameplay`. The convenience `ResumeGameplay`
API has no recorded menu-allocation argument and therefore retains its legacy
100,000 menu default; do not use it to claim continuation at an expanded cap.

## Four-player development diagnosis after explicit budget expansion

The completed `gameplay-evidence-v3-budget-development/p4-s0000-run` contains
**one independent paired seed block**, four cyclic rotations per treatment and
three-game matches: 12 focal games per treatment, not 24 independent samples.
The opportunity-minus-economic focal match differences by rotation were −168,
−115, +26 and +174 points, totaling −83, or −83/12 per game. These are development
observations, not confirmation, seat-fairness estimates or causal feature effects.

Five concrete observations guide controlled ablations:

1. **Declaration trajectories matter.** Baseline unit `000000`, second game,
   declared Diamonds at transition 1226 in round 5 and scored 210; its candidate
   counterpart scored 30 in that game slot. Candidate unit `000003` faced opponent
   Diamond declarations at transitions 492 and 1316, rounds 6 and 11. These
   public events explain large realized outcome differences; they do not prove a
   particular retained card or earlier decision caused them.
2. **Retention operated without establishing stronger play.** Baseline focal
   players made 37 jack/queen attachments before round ten; the candidate made
   zero, then 29 later jack/queen attachments. Candidate focal players made five
   formation openings/reconfigurations against baseline zero, used Kidnapper 11
   times and Justice once, yet their aggregate paired score was lower. More
   formations alone are not a sufficient competence metric. The planned
   same-menu retention ablation isolates this feature prospectively.
3. **More purchases are not demonstrated useful value.** Focal purchases rose
   from 76 to 89, and requested card quantities from 103 to 170. The candidate
   made eight late purchases for 13 cards; the baseline made none. Unit `000007`,
   second game, bought 49 cards through 25 purchases and scored 40, against 41
   for its baseline counterpart. A separately frozen purchase ablation is a
   focused future evidence need; this single trajectory does not justify changing
   policy weights or accepted purchase rules.
4. **Finance remains an empirical coverage gap.** All 2,035 selected candidate
   board decisions carried the authorized economy context, but zero selected
   income-bonus features activated. No reduced-offer or free-gift acceptance
   features activated. Three-game completion therefore demonstrates the match
   horizon, not successful cross-game financial strategy.
5. **Expanded budgets recovered retained work.** Every one of the eight old-cap
   four-player traces exactly equals the corresponding expanded-cap trace prefix.
   Four already-completed units stayed unchanged; two output-limited and two
   menu-limited units progressed further. This verifies prefix preservation and
   supports the resource-bottleneck diagnosis. These reruns are not additional
   independent samples and do not erase the earlier incomplete outcomes.

Representative public action annotations are unit `000001`, first game,
transition 439 (candidate Kidnapper formation); unit `000003`, transitions 281
and 1802 (Kidnapper and Great People formations); and unit `000007`, transitions
351/355 (Justice formation/use) and 2927 (Kidnapper formation). Unit references
are relative to the shard above; `games/<unit>/gameplay-public.json` supplies the
shareable aggregate and ending record. Exact transcript indexing was checked
privately in the corresponding `gameplay.json`; those forensic files remain
restricted. No private card identities, seeds or financial terms are reproduced
here. Public action counts are repeated observations within the same block, not
independent samples or a causal mechanism estimate.

## Normal-deal bargaining diagnostic

The bounded `gameplay-evidence-v3-bargaining/p3-s0000-run` completed all 18 planned
three-player game slots: six complete three-game matches, with zero incomplete or
never-started slots. It compared focal bargaining against opportunity with two
bargaining opponents and complete cyclic rotations on one previously used
**development** seed block. This is a nonprimary diagnostic; it neither adds an
untouched confirmation block nor establishes bargaining strength.

Across the recorded selected decisions there were **zero offers, acceptances,
declines, withdrawals, loan offers or loan acceptances**. Consequently there were
zero repeated same-actor/turn offers, but this absence cannot validate churn
control during actual negotiation: no proposal was initiated. There were also
zero voluntary departures; all 612 selected departure-boundary decisions chose
to remain (510 bargaining and 102 opportunity decisions).

The own-economy context accompanied 6,606 bargaining and 1,296 opportunity board
decisions. Zero selected decisions activated the carried-debt income bonus or
reduced-final-offer ranking, and zero promise offers, acceptances, refusals,
reduced offers or payments were observed. These are aggregates of decisions
within the same block, not independent samples. No private terms or card
identities are reproduced.

The actual-menu synthetic trade and loan fixtures therefore remain the positive
execution evidence for the new negotiation policy; this normal-deal diagnostic
shows that autonomous negotiation was still not exercised in this population.
The transcript records chosen actions, so zero chosen offers alone does not
establish zero legal offers or identify whether menu restrictions, utility
thresholds or encountered holdings prevented them. A focused opportunity
frequency/utility audit and additional strategically varied development games
would be needed to separate those explanations. No policy was retuned from this
diagnostic and no general negotiation or cross-game-finance competence claim is
made.

## Confirmed v1 action-ordering defect and prospective v2 repair

**All preceding v1 comparative effects are confounded by an implementation
ordering defect.** They remain observations of their frozen executions, but cannot
isolate retention, bargaining or another policy feature. In particular, the
bargaining diagnostic can differ from opportunity despite selecting no proposals:
this is not evidence of an unobserved negotiation benefit.

The historical `gameplayActionKey` replaced a supplied reconfiguration spec with
the existing formation's spec when `FormationID` matched. Distinct legal requests
such as protecting Clubs versus Spades therefore received the same policy tie
key. Their incoming menu order could depend on administrative command/game IDs,
so paired treatments could choose different actions without a strategic feature
difference. This is a confirmed implementation defect, not demonstrated game
imbalance or bargaining strength. The actual-menu regression creates an existing
Great People formation and two separately legal next-turn protection requests;
it reproduces equal v1 keys and distinct repaired v2 keys.

Historical `@v1` behavior and artifacts are preserved. New deterministic versions
`economic@v2`, `pressure@v2`, `opportunity@v2`, `opportunity-no-retain@v2`,
`opportunity-no-finance@v2` and `bargaining@v2` share corrected semantic ordering.
Their strategic scoring and financial priorities delegate to the corresponding
v1 definitions while recorded decision policy labels remain v2. Financial action
ordering itself is unchanged. The separate v2 board key retains the requested
formation spec and represents existing references using formation kind plus
sorted physical membership. Requested protection references receive the same
normalization; it does not recursively follow protective links or retain raw
administrative formation IDs. The old helper is untouched.

Tests cover the actual legal-menu collision, distinct requested protections,
permuted menu/formation lists, changed administrative IDs, normalized requested
protection references, observation immutability, hidden-card permutation
invariance, financial alias decisions/labels and CLI version admission. The key
regression failed before repair in
`policy-order-v2-behavior-red--20260930T125338Z-2580204`; focused repaired checks
passed in `policy-order-v2-integration--20260930T125455Z-2582006` and
`policy-order-v2-real-menu--20260930T125555Z-2583833`.

No fresh normal-play v2 campaign was run under the exhausted development
allocation. Future controlled populations must use the same corrected ordering
for every deterministic treatment and opponent, with new frozen provenance and
fresh confirmation blocks. Earlier numerical effects must not be relabeled as
v2 evidence or pooled into v2 confirmation.
