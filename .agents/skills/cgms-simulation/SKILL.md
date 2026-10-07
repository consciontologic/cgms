---
name: cgms-simulation
description: "Implement or review CGMS Phase 0 CLI simulation slices with explicit partial coverage, pure engine boundaries and CLI acceptance evidence."
---

# CGMS simulation

Use for S01–S10 implementation/review and explicitly authorized shared-engine
gameplay extensions. The 2026-09-30 extension covers D01–D04 and simulator/domain
D08–D09, bots and evaluation; it excludes frontend and online infrastructure.
Loading this skill does not authorize runtime work during a preparation-only task.

1. Load [Phase 0 instructions](../../instructions/PHASE_0_SIMULATION.md), then
   [execution guide](../../../sims/docs/phase-0-execution-guide.md). The existing
   [roadmap](../../../docs/planning/ROADMAP.md) is the only sequenced checklist;
   select its authorized slice and retain its complete acceptance gate.
2. Inventory actual runtime and test files before choosing intended paths. Use
   available CodeGraph first for source questions; distinguish contracts from code.
   Runtime belongs under `backend/`; do not put a second engine in `sims/` or bots.
3. Translate accepted rule clauses into red-first behavior fixtures. Model actions,
   response passes, required decisions, chance outcomes and continuation as explicit
   boundaries. Pure transitions receive inputs and return state/events; no clock,
   filesystem, shared random generator or scheduler dependency. Protect against
   slice/map/big-number alias mutation and keep physical card identity stable.
4. Preserve the full [CLI contract](../../../sims/README.md). Reject unknown input,
   incompatible versions and unsupported transitions honestly. A resolved earlier
   response remains committed when later work stops. Unsupported accepted rules
   are coverage gaps, not `adjudication-required` or successful no-ops. Named
   experimental replacements cannot certify the accepted profile.
5. One bounded worker owns an entire match. Test cancellation at durable boundaries,
   committed expensive-work resume, deterministic worker-count outcomes, partial
   manifest publication and retained completed units. Preserve input precedence,
   immutable game IDs, financial finalization and CLI exit statuses.
6. Use [format specification](../../../sims/docs/phase-0-format-spec.md) for strict
   inputs/provenance. Preparation examples are not runtime certification. Only
   publish runnable host/container commands after their actual gates execute.

Before claiming a slice complete, retain failed-before/passed-after evidence,
focused properties/fuzz seeds, relevant golden/CLI tests and an independent review.
Fail the review if an empty legal menu becomes a fabricated ending, a worker owns
only a game of a persistent match, budget exhaustion drops debt, or a documentation
check is offered as runtime evidence. Broad conformance/strength targets remain
visible when unfulfilled; they neither shrink nor expand Phase 0's narrow coverage.
Route replay, accounting and evaluation work to their focused CGMS skills.
