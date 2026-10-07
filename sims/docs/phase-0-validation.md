# Phase 0 preparation validation record

**PASS — preparation only, 2026-09-30.** No S01–S10 runtime gate is complete.
Entry point: [handoff](../README.md#phase-0-preparation-and-handoff).
Run ID: `phase0-prep-20260930`.

## Verified artifacts and evidence

The initial worktree was clean. Discovery and CodeGraph distinguished the existing
Flutter demo from the absent Go engine. The prepared set contains four focused
skills, one scope/execution instruction, three cited research documents, a format
specification, an S01–S10 execution guide, this validation record, a rules-default
extraction specification, one preparation-only schema and four synthetic inputs.
Indexes, changelog, simulator status/handoff, architecture inventory wording and
the generated-artifact ignore rule were updated. GAME_RULES, its accepted decision
register and ROADMAP are unchanged; all ten S items remain unchecked.

| Check | Result and limits |
|---|---|
| JSON Schema | Draft 2020-12 meta-schema and all four preparation examples pass with locally available `jsonschema 4.19.2` on Python 3.14.4. No dependency was installed. |
| Semantic examples | Canonical reduced rationals, positive denominator, zero form, debt IDs/order/seat references, uint256 seed bound, bot features/weights, full cyclic rotations, unique IDs and bounded match counts pass. The pilots schedule 24 matches (3p) and 32 matches (4p), two treatments × four blocks × population rotations. |
| Rejection cases | 16 negative mutations pass: unknown top/nested fields, false runtime claim, invalid population, negative/overflow seed, missing/duplicate rotations, zero workers, insufficient match budget, omniscient policy, invalid retention, zero/nonreduced denominator, self debt and duplicate JSON key. These test preparation tooling, not an absent Go parser. |
| Accounting oracle | Independent Python Fraction calculation reproduces Q10's six payments for reciprocal unit debts and an initial 1/3 receipt. Final seat-1 score/cash are 1/3, other scores/cash zero, both debts cleared. This validates the example, not a Go settlement algorithm or optimizer. |
| Replay vectors | Three ASCII canonical encoding/SHA-256 vectors and one seed-tuple SHA-256 digest recompute exactly. No PRNG word, state/event replay, Go/Dart parity or runtime normalization claim. |
| Skills | All four new SKILL.md files pass the bundled skill-creator validator; names/descriptions, focused procedures and relative references reviewed. |
| Links and formatting | Local file/heading links and fenced-block pairing pass over changed/new Markdown; whitespace diff checks pass. External research URLs were opened during dated primary-source research; the Common Random Numbers finding is limited to the accessible publisher abstract. No claim that moving upstream pages are immutable. |
| Scope and preservation | Original CLI body from Purpose through retention is byte-preserved. Rules/register/roadmap are byte-identical to HEAD; backend remains absent. Generated `sims/artifacts/` paths match Git ignore. Preparation is labeled explicitly throughout. |
| Review and evidence handling | Complete changed/new text reviewed for secrets, unrelated files, generated artifacts, contradictory instructions and gate omissions. No secret-pattern findings. No simulator-generated output was created or staged. |

Primary local logs (temporary operational evidence, not committed simulation output):

- `/tmp/agent-runs/cgms-preparation-validation--20260930T080218Z-1545994.log`
- `/tmp/agent-runs/phase0-review-schema--20260930T080153Z-1545537.log`
- `/tmp/agent-runs/independent-prep-validation-py3--20260930T080422Z-1549937.log`
- `/tmp/agent-runs/independent-prep-skills--20260930T080422Z-1549984.log`

The coordinator's one-off validation script is
`/tmp/agent-runs/cgms-prep-validate.py`; it is a preparation check, not shipped code.
Its invocation was the existing safe-run wrapper with tag
`cgms-preparation-validation`, separator `--`, then `python3` and that script path,
from the repository root after confirming `pwd`. Independent verification inspected
and reran it and separately checked skills, scope and whitespace. Future sessions
can reconstruct the checks from the schema/examples and this record; temporary
logs/scripts are not a new project dependency or permanent simulator evidence.

## Review corrections and recovery

Independent review found that direct private CLI arguments would leak through the
existing safe-run console/command log and failure breadcrumb. Corrected the format,
execution and instruction documents to require a future S09 privacy-aware xops
entry point using an opaque protected run specification, private inner argv/output
and redacted diagnostics, with success/failure canary tests. No unimplemented helper
is presented as executable. Added official Go ChaCha8 API/implementation evidence.
Consolidated the duplicate Q10 proof/algorithm discussion into the main research
record and linked it from operations research.

Independent forward review exercised the future S07 large-denominator stop/resume
and S09 unsupported-F01 continuation cases using the skills/handoff. Both route to
exact preserved progress and explicit incomplete status, without netting debt,
repeating costs, resetting committed randomness or reopening accepted rules.
Reviewer and verifier returned PASS after corrections.

The verifier initially invoked unavailable `python` (exit 127). Its captured log
was read, the installed `python3` was selected, and the rerun passed. The failure
breadcrumb was marked resolved only after that evidence. An earlier skill-check
wrapper invocation omitted the `--` separator, returned usage 64 before making a
log, and was corrected after inspecting the wrapper; all four skill checks then
passed. Neither issue caused a skipped gate or source rollback.

## Implementation-only validation

Still required in the authorized implementation: red/green Go behavior tests;
104-card transition properties and alias isolation; all named rule/scenario gates;
PRNG/sampler/state/event goldens; runtime parser/schema negatives; reference/batched
settlement equivalence and compressed-audit expansion; large-D durable resume;
seat/menu/reason privacy and hostile bot tests; all six real CLI commands/exits;
worker-count determinism, cancellation/race/I/O failure tests; schema-valid runtime
manifests/checkpoints/reports; public export/retention tests; offline host/container
execution; reproducible clean builds and measured performance. No such results
are claimed here, and the full-conformance/1,000-game/1,000-block targets remain
visible and unfulfilled pending their prerequisites.

The completion tracking row and staged file inventory are recorded by the parent
under AGENTS; no commit or push is performed. The next implementation task can load
this handoff and begin S01 without relying on this conversation.
