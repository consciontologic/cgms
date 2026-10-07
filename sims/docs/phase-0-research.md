# Phase 0 research and engineering decisions

Preparation recorded 2026-09-30. This is implementation guidance, not a runtime,
rule amendment, architecture approval or additional roadmap. Load the
[handoff](../README.md#phase-0-preparation-and-handoff) first.
[GAME_RULES](../../docs/project/GAME_RULES.md) controls gameplay;
[ROADMAP](../../docs/planning/ROADMAP.md) controls sequencing and gates.

## Inventory and authority

At preparation start, Git was clean at `2626c38`; session bootstrap reported the
prior failure resolved and a previous UI staging checkpoint. CodeGraph exploration
of `backend/internal/game` and `backend/cmd/sim` returned Flutter demo symbols.
Filesystem discovery confirmed no `backend/` or `config/`, and only `sims/README.md`
in `sims/`. The existing Flutter controllers model selected presentation flows;
they are not the shared engine or an independent rules oracle. xops supplies real
bootstrap, safe-run and tracking helpers. This preparation adds documents, skills,
a preparation-only schema and synthetic examples; it adds no Go module/runtime.

All Q01–Q12, F01–F10, N01–N07 and T01–T03 are accepted. ADR-0001 remains proposed;
this task authorizes preparing its roadmap topology, not silently changing its
approval status. Complete rules, databases, server idempotency and production
security stay with their scheduled phases. Phase 0 still needs the exact S01–S10
gates, including narrow continuations/accounting that involve later full-rule
features. A synthetic post-effect fixture is valid boundary evidence, not proof
that the entire originating ability is implemented.

## Primary evidence register

All URLs below were accessed **2026-09-30**. Sources support the findings in the
second column; the third column is a CGMS engineering choice, not a source claim.
Moving `master` documentation is design evidence, not a dependency lock.

| Source | Supported finding | Decision, alternative and limitation for CGMS |
|---|---|---|
| [OpenSpiel paper, Lanctot et al., v6](https://arxiv.org/abs/1908.09453v6) | Supports sequential, chance, imperfect-information and general-sum games under a shared research framework. | Reuse its separation of game/state/action/information concepts. Do not import its C++/Python stack into one Go module or infer equilibrium/balance claims for CGMS. |
| [OpenSpiel concepts](https://openspiel.readthedocs.io/en/latest/concepts.html) | Explicit chance nodes, legal actions and state transitions; `ApplyAction` mutates, `Child` creates a successor. | Prefer an immutable Go transition boundary; explicit random requests/outcomes. Copy-on-write is an optimization only after alias tests. A mutable shared match state is simpler but too easy to leak into bot/search branches. |
| [OpenSpiel primary API implementation](https://github.com/google-deepmind/open_spiel/blob/master/open_spiel/spiel.h) | Distinguishes state, observation/information-state interfaces and serialization; legal actions are part of the game API. | Define seat projections and decision kinds before bots. A CGMS multi-stage response cannot be collapsed into one opaque bot call: F01/T02 need stable waits and interruption points. |
| [OpenSpiel developer guide](https://github.com/google-deepmind/open_spiel/blob/master/docs/developer_guide.md), [API reference](https://github.com/google-deepmind/open_spiel/blob/master/docs/api_reference.md) | Game implementations receive common tests; game/state serialization is exposed. | Add reusable invariant/trajectory/serialization tests as well as rule-specific fixtures. Serialization alone does not prove compatibility, privacy or ledger conformance. |
| [PettingZoo paper, Terry et al., v7](https://arxiv.org/abs/2009.14471v7), [AEC API](https://pettingzoo.farama.org/api/aec/) | AEC makes sequential agent turns explicit; API distinguishes termination and truncation and supports action masking. | Use explicit required actor, stage and supported legal choices. Keep a budget stop separate from a rules ending. Do not copy a generic agent-cycle advance that skips eligible disconnected seats or loses CGMS's Coup boundary. |
| [Random123 authors' project](https://random123.com/), [primary implementation documentation](https://random123.com/releases/1.02/docs/index.html) | Counter-based generators obtain indexed random values without walking a shared mutable sequence. | Independent logical streams remove scheduling dependence. Counter-based Philox is an alternative if random access later dominates; choose pinned Go ChaCha8 streams now to avoid a new engine dependency, with algorithm vectors and explicit counters. Neither choice removes the need to pin sampling. |
| [Go rand/v2 API](https://pkg.go.dev/math/rand/v2), [ChaCha8 implementation](https://go.dev/src/math/rand/v2/chacha8.go) | Local ChaCha8 sources expose a 32-byte seed, Uint64 and binary state encoding. Interleaving Read and Uint64 has undefined bit-return order. | Pin Uint64-only generation and state marshal/resume vectors. Standard-library convenience samplers can evolve separately; use the specified sampler. The package is for simulation, not security-sensitive draws. |
| [RFC 8785](https://www.rfc-editor.org/rfc/rfc8785.html) | JCS defines deterministic JSON serialization, constrained number handling and key ordering; it does not normalize Unicode. | Use a constrained JCS-compatible domain (ASCII machine keys, safe integer JSON numbers, rationals/large integers as strings), with explicit semantic normalization before hashing. Ordinary sorted JSON alone is not a claim of full JCS implementation. |
| [JSON Schema 2020-12](https://json-schema.org/draft/2020-12), [object closure](https://json-schema.org/understanding-json-schema/reference/object) | Closed objects reject additional properties; composition needs care about where closure applies. | Use closed tagged objects and local schema references; duplicate keys, canonical rationals, reference integrity and seat schedules require semantic checks beyond schema. No runtime network schema resolution. |

Current toolchain, exact arithmetic, worker lifecycle, offline build, atomic
publication and Go testing evidence are in
[operations research](phase-0-operations-research.md). Original evaluation papers,
reference bot implementations and estimator limitations are in
[experiment research](phase-0-experiment-research.md). These are parts of this
research record, not alternate contracts.

## Engine and partial-coverage decision

Use proposed `backend/internal/game` for rules, `backend/internal/bots` for seat-only
policies and `backend/internal/sim` for execution/artifacts, composed by
`backend/cmd/sim`. Public symbol names are intentionally not frozen by preparation.
Transition input is a fully validated state, command, pinned rule profile and
explicit random result; output is a new state, ordered events and either a next
required decision, automatic work, completed scenario/game, or typed rejection.
The core never performs file, clock, network or global-random operations.

Separate legality from implementation coverage. Record supported transitions and
fixture IDs in a versioned coverage manifest. An unsupported accepted transition
stops with code 3 before that transition mutates anything; already committed
responses/costs remain with a paused-action marker. Do not omit unsupported actions
from a bot menu and call the remaining policy full legal-random. A supported-slice
menu must identify its restricted action universe and stop when the next mandatory
step is unsupported. Full rule legality must never be guessed by the runner.

Represent card inventory, current control, history, binding, restrictions and
active-Ace zones separately. Deep-copy slices/maps and exact amounts at the
boundary. Expose immutable values to policies. Canonical physical identifiers may
encode suit/rank internally; concealed public handles must not. F10 approves hand
and concealed-Ace counts, not hidden rank counts. Use noninterference tests on
observations, menus, reasons, errors and public traces, preserving already-known
history when constructing equivalent hidden states.

## Replay protocol decision

The concrete format is in [format specification](phase-0-format-spec.md). Adopt:

- A private event/decision transcript as replay input; bot recomputation is a
  different diagnostic. Record command IDs, immutable game/replacement IDs,
  required actor/stage, random choice IDs/results, counters and before/after hashes.
- Pin ChaCha8 algorithm identity plus exact Go implementation/toolchain, using only
  its `Uint64` stream for simulation. Never mix `Read` and `Uint64`. Check published
  implementation behavior at S04 and freeze independent word vectors before use.
- Derive streams from a versioned SHA-256 tuple, not worker IDs or completion order.
  Separate environment/deal/initiative/effect streams and each seat's policy stream.
  Paired variants reuse the block schedule; consuming more bot randomness cannot
  change environment randomness. Different policies may still diverge in draw count;
  paired seeds are variance control, not identical trajectories.
- Pin rejection sampling and descending Fisher–Yates order; sort eligible physical
  candidates by internal ID before sampling. Random counters count rejected words
  too. Persist chosen outcomes before crossing a decision boundary.
- Semantic normalization precedes canonical serialization and SHA-256. Preserve
  meaningfully ordered queues/debts/events; sort only explicitly unordered sets.
  Keep wall times, worker count and output paths out of state/config rule digests,
  but retain them in provenance. Hashes detect drift, not authenticity.
- Reject unknown compatibility versions before replay. Compare inputs, decisions,
  RNG counters, then state/events at each transition. First divergence reports its
  index, rule/decision ID and expected/actual digest; private field-level diff is
  restricted. Never auto-regenerate a failed golden fixture.

Alternative: seed-only reruns are cheap but cannot distinguish policy, RNG,
serialization and rule changes; unversioned event sourcing merely moves the drift.
Exact replay checks consistency, not correctness; pair with independent expected
ledgers, card properties and manually reviewed rule fixtures.

## Exact settlement decision and proof obligation

GAME_RULES §4.2/Q10 supplies the normative algorithm, not an external finance model.
Use arbitrary-precision reduced rationals behind immutable values. A private
mutable `big.Rat` receiver must never alias a previous state or returned amount.
Reference settlement admits precomputed gross receipts by ascending seat, consumes
one FIFO receipt against that recipient's debts by immutable creation sequence,
and appends actual player-creditor receipts to the tail. System receipts stop.
Charge once on obligation creation; subsequent repayment changes cash/debt, not
another debtor score debit. Credit every actual gross creditor receipt exactly.

For finite frozen input, scale by a common denominator D; every positive payment
reduces total debt by at least 1/D and creates no debt. There are finitely many
payments; zero amounts must be skipped without enqueueing zero receipts. Processing
non-paying receipts is also finite because only a payment adds a new receipt.
This is a termination proof, not a practical cost bound: reciprocal unit debts
with receipt 1/D perform 2D transfers, exponentially many in D's bit length.

First implement the unbatched oracle for bounded generated ledgers. Optimize only
repeated FIFO segments with a demonstrated invariant: identical active debt heads,
queue shape and transfer amounts until the next debt exhaustion or queue boundary.
Skip an exact integer number of whole repeats bounded by that first change; handle
the remainder using the reference rules. A two-seat shortcut is not evidence for
arbitrary three-seat cycles, multiple debts or interleaved receipts. Fall back to
resumable reference processing for unproven shapes, never reciprocal netting.

Compressed audit segments must retain algorithm version, starting cursor/state
hash, ordered one-cycle payments, repetition count, exact accumulated gross receipt
and debt deltas, ending cursor/hash, and linkage to prior segment. Expansion on
small cases must equal the reference **ordered** trace, not just final cash.
Checkpoints retain working ledger, FIFO queue including partially processed head,
debt cursors, score accumulators, sequence counters and audit prefix. The public
ledger remains at its committed boundary until exact settlement completes; no
command or Coup interleaves with unfinished automatic settlement. Resume validates
compatibility/hash and continues that workspace without paying twice or restarting
expensive settled prefixes. Budget exhaustion preserves it and reports incomplete.

Test small D=1,2,3,7; large decimal denominators; system sinks; oldest debt changes;
three-seat cycles; simultaneous halves; zero debt/receipt; interruption at every
segment boundary. Include F04 cash clearing, N06 correction lineage and void-game
restoration. Thirds are arithmetic cases, never three-person Ponzi alliances.
No external source proves the future CGMS optimizer correct; differential traces,
proof of each skip rule and restart tests are mandatory S07 evidence.

## Decision closure and limits

Every requested research area routes to a concrete decision: engine/coverage and
replay above; accounting above; schema/provenance/privacy in the format spec;
bots/statistics in experiment research; worker ownership, recovery, toolchain,
container and test/profiling in operations research. Remaining uncertainties are
implementation measurements: branching cost, rational bit growth, throughput,
file-system durability and stronger-bot evidence. These do not reopen accepted
rules. Pin practical input/work budgets, measure before tuning and retain stopped
runs. No balance, complete-playability, performance or reproducible-binary result
is established by this preparation.
