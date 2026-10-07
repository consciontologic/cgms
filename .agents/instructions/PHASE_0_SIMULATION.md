# CGMS Phase 0 implementation instructions

Scope: future work on `backend/cmd/sim`, `backend/internal/game`,
`backend/internal/bots`, `backend/internal/sim`, simulation schemas/rules,
curated `sims/` inputs and associated xops helpers. Load explicitly from
[sims handoff](../../sims/README.md#phase-0-preparation-and-handoff); clients do not
necessarily auto-load this directory. Applies to Astra and other implementers.
This file does not authorize implementation during a preparation-only task.

## Authorized gameplay extension (2026-09-30)

The owner explicitly authorized D01–D04 domain rules, simulator/domain portions
of D08–D09, complete bots and measured evaluation beyond Phase 0. For that task,
Phase-0-only exclusions in this handoff do not restrict accepted gameplay work.
Keep PostgreSQL/server durability gates unchecked; no online infrastructure is
a prerequisite. Historical Phase 0 evidence remains partial-sequence evidence.
All purity, exact accounting, privacy, replay and verification requirements below
continue to apply. The existing roadmap remains the sole progress checklist.

## Authority and scope

[AGENTS](../../AGENTS.md) governs operations; [GAME_RULES](../../docs/project/GAME_RULES.md)
governs gameplay; [ROADMAP](../../docs/planning/ROADMAP.md) is the only sequenced
checklist. Load [roadmap discipline](ROADMAP_DISCIPLINE.md), the full
[CLI contract](../../sims/README.md), and the
[execution guide](../../sims/docs/phase-0-execution-guide.md). All Q/F/N/T decisions
are accepted. An unsupported mechanic is not an unanswered rule question.
Use CodeGraph first for source questions; distinguish Flutter demo behavior from
authoritative Go implementation. Do not approve the proposed ADR by implication.

Keep one backend module and pure engine; no application service or LLM dependency.
No Phase 1 transaction/server implementation is needed to close Phase 0.
Never weaken an S gate, mark a partial item complete, or mark runtime work complete
for documentation. Extend the implementation's coverage record as features land;
state the exact unsupported remainder in every experiment. A scenario may finish
while its enclosing game remains unimplemented/incomplete.

## Execution and recovery

Follow the [simulation skill](../skills/cgms-simulation/SKILL.md) plus topic skills
from the handoff. Read the whole current S gate before each slice; write meaningful
failing tests before behavior; use only actually implemented commands. Store the
real invocations and red/green logs as evidence. Do not generate mock results.

One worker owns a whole match. Transitions and settlement checkpoints are atomic
logical boundaries; cancellation cannot manufacture pass/refusal/game ending.
Persist last committed state, pending decision and exact private settlement cursor
before a controlled budget stop. Do not replay all expensive work on resume or
roll back prior legal responses on reaching an unsupported transition.

Use the existing safe-run and failure recovery loop for builds/tests. For private
simulator runs, first implement S09's privacy-aware xops entry point from the
execution guide: the current wrapper logs argv/output and failure state. Supply
only an opaque ID/protected run-spec path to safe-run; keep full private argv and
redacted subprocess handling inside that tested helper. Do not pass seeds or
private cards directly to existing wrapper logs. For interrupted agent work,
write repository tracking checkpoint with scope, step, last command, completed
and remaining gates, evidence references and next action. For simulator work,
use its separate private checkpoint under `sims/artifacts/`; do not put hidden
cards or seeds in tracking state/CSV. Recovery validates versions and hashes,
then continues the original logical IDs. Keep user/concurrent edits intact.

## Review, validation and evidence

At a completed implementation phase boundary, use implementer → reviewer → verifier
as the repository workflow requires; independent delegates return evidence only.
Review full current/new files, not only the staged diff. Rule review checks clauses,
negative cases, permitted visibility and costs before accepting golden expectations.
Verifier runs applicable unit/property/fuzz/golden/differential/CLI/cancellation/race
checks against the same candidate; failing gates enter recovery, never assertion
weakening or blind retries. No performance or full-conformance claim follows from
static files or a green narrow unit test.

[Format specification](../../sims/docs/phase-0-format-spec.md) supplies status,
privacy, provenance and retention contracts. Synthetic curated inputs only;
generated traces, seeds, debug profiles, manifests and copied configs remain ignored.
Public reports use approved observer projections and incomplete denominators.
Retain review-cited evidence until explicitly released; preview before future pruning.

Only after the full selected scope and its review/checks pass, inspect the complete
staging set for unrelated edits/secrets/generated output, append the one completion
row and stage under AGENTS. Never commit/push. Update ROADMAP only for actual
implementation gates, immediately when fully evidenced; the guide has no checkboxes
or independent completion state to maintain.
