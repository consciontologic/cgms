# Phase 0 input, replay and artifact format specification

Status: **preparation only** (2026-09-30). The schema and examples linked from the
[handoff](../README.md#phase-0-preparation-and-handoff) validate design inputs, not
executable simulator configurations. S03/S09 must implement and test distinct
runtime schemas before promoting any example. No command accepts these files yet.
The existing [CLI contract](../README.md#planned-cli-surface) is preserved in full.

## Preparation files and promotion

[Preparation schema](../../config/schemas/phase-0-preparation.schema.json) uses
JSON Schema 2020-12, local `$defs`, closed objects and three tagged kinds:
`scenario`, `experiment`, `replay-vector`. Every example must carry
`status: preparation-only` and `runtime_validated: false`. This marker is never an
accepted runtime configuration version. The scenario is a small exact-ledger
oracle input, not an engine snapshot or a playable game. The experiment demonstrates
a bounded pilot, not the 1,000-block held-out target. Replay vectors pin only the
encoding/hash/seed-derivation portions computable without Go; PRNG words and actual
state/event traces remain S04 work. Schema defaults never manufacture values.

At implementation, introduce closed runtime schemas under `config/schemas/` for
rules, scenario, experiment, manifest, outcome, trace record and checkpoint.
Names intended: `rules.schema.json`, `scenario.schema.json`,
`experiment.schema.json`, `manifest.schema.json`, `outcome.schema.json`,
`trace.schema.json`, `checkpoint.schema.json`. These paths are **planned**.
Reuse common rational/identity/version definitions with local references. Publish
schema dialect and content hashes. Reject runtime loading of preparation tags.

Runtime scenario promotion requires a complete 104-card partition or a named
accounting-only fixture type, immutable game/match identity, valid phase/support/
quotas, typed commands with actors/decision IDs, committed random inputs, expected
per-step events/zones/counters/exact ledgers and every seat projection. Each case
names its rule clauses and Q/F/N/T IDs, coverage requirements and expected stop.
An arithmetic fixture can complete its scenario without completing a game; both
statuses must be separate in results. No synthetic state bypasses invariants merely
because it lives in testdata.

## Strict parsing and semantic checks

Before mutation, validate UTF-8, one JSON value, no duplicate keys, exact key case,
unknown-field rejection at every nesting level, no NaN/Infinity, bounded input
bytes/depth/array sizes and exact numeric types. Reject fraction floats, stringified
booleans, non-integer counts and unsupported versions. Paths resolve relative to
the caller as the CLI contract says; snapshot resolved bytes into private artifacts
so replay does not depend on a later working directory. Do not fetch schemas or
follow remote references at runtime. Resolve curated paths within the designated
input root; output artifact names are relative and reject traversal/symlink escapes.

Canonical rational input is `{ "numerator": "-3", "denominator": "2" }`, decimal
strings, no plus sign/leading zeros/negative zero, denominator positive, gcd=1,
zero=0/1. Financial state requires canonical values; semantically equivalent raw
configuration fractions may be reduced during effective-config normalization,
with original bytes retained and the resulting canonical form validated. Domain
amount signs depend on type: debt/receipt/cash nonnegative, score/correction signed.
Preparation examples use only canonical values. Input digit/byte limits protect
parsing; internal arithmetic growth yields checkpointed resource exhaustion, not
rounded amounts or an arbitrary cap on legal obligations.

Cross-field checks include unique IDs and physical cards, exactly 3/4 seats,
seat indices in range, debt creation order, non-self creditor references, supported
rule/engine/version intersection, compatible scenario coverage, complete lineups
and rotations, disjoint development/held-out blocks, no omniscient bot in a fair
population, bounded product of grid × blocks × rotations × matches × games,
positive fixed games per match, rational canonicality and output privacy policy.
Explicitly validate these in Go; JSON Schema alone does not establish them.

## Effective configuration and version identities

Rules defaults come only from future `config/rules/game-rules.json`, as specified
in [rules configuration](../../config/rules/README.md). Resolve canonical defaults,
selected config, corresponding experiment values, then corresponding explicit CLI
flags. Validate the merged result, record every override and retain source hashes.
No environment variable silently overrides gameplay. Existing service YAML
precedence in ARCHITECTURE is for services, not this simulator. `--interpretations`
omitted means accepted GAME_RULES, with no experimental replacements.

Separate these identities:

| Identity | Meaning and compatibility gate |
|---|---|
| rule source revision + SHA-256 | Exact GAME_RULES bytes used, including accepted amendments; source change requires explicit review. |
| ruleset + interpretation ID/version/hash | Accepted constants versus an explicitly named clause replacement; strict accepted profile contains no replacement. |
| engine + state/event/command schema versions | Exact supported transition semantics and serialization; unknown versions fail before starting/replaying. |
| PRNG + derivation + sampler + shuffle versions | All must match for recomputation; recorded random outcomes still need algorithm/provenance validation. |
| bot ID/version/config hash | Distinct from engine; changing weights changes identity and evaluation treatment. |
| scenario/experiment/coverage hashes | Exact fixture, schedule, budgets, supported transition set and exclusion policy. |
| analysis version/seed | Interval method/resampling identity, independent of match randomness. |

An effective **rules** hash excludes scheduling, hardware, output paths and policy
weights. An effective **experiment** hash includes full schedule, bots, budgets,
analysis and privacy/retention decisions. A run manifest records workers and wall
limits even though those are excluded from deterministic state/outcome digests.
Changing an operation budget changes the experiment; changing workers alone must
not change completed semantic outcomes. Compatibility uses a maintained exact
allow-list of tested version tuples, not optimistic major-version guessing. Unknown
rules/engine/version → exit 2. Supported version with unsupported transition → 3.

## Canonical encoding, hashes and random streams

Engineering protocol `cgms-canonical-v1`: normalize effective defaults and rational
representations first; object keys are ASCII and lexically sorted; JSON strings
are UTF-8, using `\b`, `\t`, `\n`, `\f`, `\r` for those controls, lowercase
`\u00xx` for other U+0000–U+001F controls, and `\"`/`\\` for quote/backslash.
Emit other Unicode characters literally (including U+2028/U+2029); reject unpaired
surrogates. No HTML/slash escaping, Unicode normalization, whitespace or terminal newline. Integers are base-10, no exponent,
within ±(2^53−1); larger quantities are canonical decimal strings. No floats in
domain data. Arrays preserve order; schema-designated sets are sorted by stable ID
before encoding. This constrained domain agrees with RFC 8785; a generic encoder
outside this domain is not automatically JCS. Hash SHA-256 over these exact bytes,
render lowercase 64-hex. File checksums instead hash stored bytes including any
newline. Do not confuse file checksums with semantic hashes. The ASCII config
vectors are preparation checks, not cross-language or runtime certification.

Engineering protocol `cgms-seed-v1`: root seed is a canonical unsigned 256-bit
integer decimal string (0 ≤ value < 2^256). SHA-256 of the canonical JSON array
below yields the 32 seed bytes, used in digest byte order without endianness reversal:

```text
["cgms-seed-v1", root_seed, population, pairing_id, block_id, rotation_id,
 match_index, game_slot, replacement_index, stream_kind, seat]
```

Population is integer 3/4; IDs and root are strings; match/slot/replacement/seat
are integers. Match index and game slot start at 0; replacement starts at 0 and
increments on each nullification. Seat 0 denotes an environment stream, otherwise
1..population. Stream kinds: `deal`, `initiative`, `effect`, `bot`. Each bot gets
its own seat stream; environment substreams persist independently. Pairing ID and
block schedule are frozen independently of candidate/config hashes. Rotation ID
identifies the exact enumerated mapping; counterpart treatments share the same
exogenous tuples. Within a slot, redeals advance `deal`, never reset it. Unconsumed
stream state, counters and algorithm version go in private checkpoints. Analysis
uses its own root seed and domain `cgms-analysis-v1`, with tuple
`["cgms-analysis-v1", analysis_seed, population, pairing_id]`; it never consumes a
match stream. S04 adds actual Go ChaCha8 output vectors and marshal/resume tests.

Sampling `cgms-u64-reject-v1`: for 1 ≤ n ≤ 2^64−1 set t = 2^64 mod n using exact
arithmetic; draw unsigned x until x ≥ t; choose x mod n. Every consumed word,
including rejections, increments the counter. n=0 is invalid. Shuffle
`cgms-fisher-yates-v1`: for i from len−1 down to 1, swap i with sample(i+1).
Empty/singleton shuffles consume zero words. Candidates have canonical internal
physical-ID order before sampling; manifests retain before/after counters and
chosen identities privately. Do not depend on a library's unspecified shuffle or
bounded-number convenience method. Property tests check permutations; golden
vectors check exact draws and rejections, including n near 2^63+1.

## Command preservation and completion semantics

The six commands retain every input in the README table: run's `--config --seed
--players --bots --out` and optional `--scenario`; batch adds `--games --workers`;
tournament/sweep use `--experiment --config --seed --out`; replay uses `--manifest
--out`; compare uses `--baseline --candidate --out`. Every game-producing command
accepts `--interpretations`; flags override corresponding experiment inputs and
are manifested. `--workers` is scheduling only. Do not invent a seventh command
or present any as executable before S09. Host/container examples become tested
commands only in S01/S10 as applicable.

`run` and ordinary `batch` start independent one-game matches; tournaments retain
scores/debts across their fixed game count (default three), finalizing finances
before zero-cash next games. Nullification restores the start snapshot and gets a
new immutable game ID in the same slot. Report declaration, game score leader and
tied cumulative rank separately. CLI attempts a new output directory and rejects
an existing one. The output path of the published manifest is printed.

| Exit | Required meaning / test |
|---|---|
| 0 | Requested operation completed and artifacts verify. A completed accounting scenario must say `scenario-complete`, with game completion false; never count it as a finalized game. |
| 2 | Invalid input, existing output request or incompatible versions; no game starts. |
| 3 | Unsupported accepted transition, or genuinely undefined experimental adjudication; preserve prefix and rule ID. |
| 4 | Replay divergence or invariant failure; stop at first divergence, retain diagnosis. |
| 5 | Work/resource budget exhausted; resumable incomplete progress, no fabricated ending. |
| 6 | I/O failure; do not claim publication if even the failure manifest cannot be written. |
| 130 | Cancellation; preserve completed units and private progress wherever storage permits. |

Batch collects diagnostics and returns nonzero for **any** incomplete game. To make
mixed failures deterministic, preparation chooses aggregate priority
6 > 4 > 130 > 3 > 5 after preflight input/version errors (2); retain every child
code and status. This priority is an adapter choice, not gameplay. A successful
`compare` of partial manifests may return 0 for a valid **analysis artifact**, but
must preserve the inputs' incomplete statuses and must not mark their games
complete. Plain `replay` verifies the committed prefix of an incomplete run and
returns its incomplete status; it cannot convert a stopped game to complete.

Recovery is deliberately not an undocumented `resume` CLI flag: S09 may first
expose explicit continuation through a validated scenario/checkpoint input using
`run --scenario` with a new output directory and recorded parent manifest, once
that runtime scenario variant and integration test exist. Complete prior artifacts
remain immutable. Replay is forensic verification, not a command to rerun bots or
repeat settlement. S07 still must test checkpoint/resume internally even before
this adapter exists.

## Runtime artifacts and privacy

All rows below are **S09/S10 format obligations**, not existing output schemas.
Every object is closed, versioned and strict; unknown kinds fail before replay.

| File / record | Required content |
|---|---|
| `manifest.json` (private) | Complete CLI-contract manifest: rule source/revision/hash; rules/schema/engine/interpretation versions + hashes; effective bytes; scenario/experiment/override hashes; coverage/exclusions; reproduction argv and cwd; engine source/dirty patch/binary/dependency lock/toolchain/platform/build flags; bot identities/config hashes; all root/derived streams and counters; status; immutable game IDs/replacements; artifact paths/bytes/SHA-256/privacy; parent checkpoint provenance. Also planned/started/finalized/unsupported/budget/canceled/not-started counts per population. |
| `trace.jsonl` (private) | Versioned sequence number, game ID, command ID, actor, kind, expected boundary, action/window/effect/decision IDs/stage, validated input and authorized choices/results, random commitment IDs/outcomes/counters, ordered events, pre/post state/event hashes; relevant proposal revision, draw cursor, purchase quantity/payment, charge/correction lineage and financial progress. |
| `outcomes.json` | Per-index population/block/rotation/treatment/match/game identity; scenario status separate from game and match status; exact scores/cash/debts; declarations/ranks; stop reason/unsupported rule; work and timing; observer and privacy label. Shareable projection omits private replay inputs. |
| `report.md` | Purpose/executive summary; provenance and private reproduction reference; coverage/interpretations; named observer narratives and bot reasons; exact score/cash/debt ledgers; ending reason; counts/denominators, metrics and tuning hypotheses; child report links for batch/tournament. No private seeds/commands embedded in shareable output. |
| checkpoint + audit segments | Version tuple; parent hashes; last committed public boundary; private working state, FIFO partially consumed head, debts, score accumulators, decisions and RNG state; ordered compressed segment metadata and cursor; deferred draws/counters and replacement identity. Resume continues exactly once; gameplay cannot overtake settlement. |
| public export manifest | Explicit `not-replayable` projection, observer ID, allowed aggregate/source-version hashes and redacted artifact list. Never claim it can substitute for the restricted forensic manifest. |

Canonical state hashing includes all state relevant to future transitions, including
usage/history, effects/bindings/restrictions, pending decisions, draw queue, private
random state and exact finance lineage; excludes scheduling/timestamps. Ordered
event hashing includes schema version and the full authorized internal event order.
Private hashes alone can still fingerprint secrets; expose only approved aggregate
references in public output. Use synthetic identities in curated fixtures.

Private files use restrictive permissions (directory 0700, files 0600 on POSIX).
Shareable reports are built from an explicit `public` or `seat-N` view; seat-N
exports remain private to that seat unless deliberately authorized for sharing.
Omniscient diagnostics are labeled restricted and excluded from fair comparisons.
Generated artifacts, copied inputs, seeds, pprof and dumps belong only under ignored
`sims/artifacts/`, never tracking CSV, global logs or Git. Operational logs use
safe IDs/statuses, not raw commands containing private seeds. The existing safe-run
wrapper logs argv/output and copies a failed command into tracking state. S09 must
provide a tested privacy-aware xops entry point accepting an opaque protected
run-spec path; the inner helper retains private argv/output under artifacts and
passes only redacted diagnostics outward. No such helper exists yet. A public reproduction
reference points to restricted evidence, rather than publishing a reconstructive
seed/deck/transcript. Full local manifests still retain the complete command.

Publish in an exclusively reserved new directory: write/close/check/hash temporary
siblings, finalize artifacts, then publish the verified manifest atomically last.
Readers treat no manifest as unpublished and a partial manifest as incomplete.
Checkpoint generations publish atomically with prior-hash linkage; verify artifact
paths and hashes before consuming them. Flush file and parent directory where
crash durability is claimed; test supported filesystem/platform behavior. A process
kill can leave unpublished files; do not manufacture a success manifest on recovery.

Retention is the existing proposed **30 days**, expressed explicitly in experiment
policy; `review_hold` protects cited runs/inputs until explicitly released, even
past expiry. Record creation/expiry timestamps outside semantic digests. Future
pruning previews exact paths, checks holds and root containment, and never deletes
curated `sims/scenarios/` or `sims/experiments/`. No pruning is performed here.
