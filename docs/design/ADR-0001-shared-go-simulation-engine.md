# ADR-0001: Share one Go rules engine between server and simulator

- **Status**: proposed
- **Date**: 2026-09-28
- **Deciders**: pending project-owner review
- **Supersedes**: N/A
- **Superseded by**: N/A

## Context

CGMS needs an early simulator and future authoritative Go backend. Hidden information,
ordered responses, fractional debts and evolving versioned rules make separate engines costly
to reconcile. Phase 0 requires no PostgreSQL, Redis, Flutter, network or LLM. The
[simulation contract](../../sims/README.md) defines outputs/gates; the
[architecture](../code/ARCHITECTURE.md) defines the system boundary.

## Decision

Recommend one `backend/go.mod`, a pure `backend/internal/game` engine shared by
`backend/cmd/server` and `backend/cmd/sim`, and seat-limited `backend/internal/bots`.

## Consequences

- Shared adjudication reduces drift; deterministic fixtures must still prove adapter/replay parity.
- Root `sims/` owns scenarios, experiments, notes and reports, without another runtime or default-config source.
- Go simplifies testing/bounded parallelism; exploratory reporting takes more code than notebooks.
- Exact arithmetic/provenance have costs; measure realistic action branching before optimization.
- Online seeds/private state stay confidential; full simulation traces are explicit research exports.
- Mobile reuse remains unresolved: evaluate native integration versus a client engine with differential fixtures.
- Revisit after measured Phase 0 friction/profiling; Python analysis must retain Go adjudication and versioned artifacts.
- Architecture approval, implementation, replay and baseline evaluation gates remain pending.
- [GAME_RULES](../project/GAME_RULES.md) resolves Q01–Q12 and incorporates the accepted F01–F10 and N01–N07 interaction amendments; their shared engine/API/client scenario fixtures remain future work. Accepting rule recommendations does not approve this architecture automatically.

## Considered options

- **Go throughout (recommended)** — best engine reuse and one toolchain; direct
  unit/property/race tests, but more effort for exploratory analysis.
- **Python engine** — rapid experiments, but duplicates future Go rules and hidden-information/accounting/replay maintenance.
- **Hybrid Go engine + Python analysis** — preserves adjudication and rich analysis;
  defer additional environments/adapters until measured needs justify them. Never duplicate rules.
- **Documents/manual playtests only** — low maintenance, but cannot deliver CLI simulation or automated reproducibility.

## Current delivery sequence (2026-09-30)

The shared Go implementation and local authority have the evidence recorded in
[ROADMAP](../planning/ROADMAP.md); the original proposal above remains its decision
basis. P01 now compares offline candidates on the host and records a provisional
recommendation. P04 adaptive Flutter web can proceed independently. Final native
packaging, Android/iOS device measurements and engine approval belong to R06N
before release; this sequence does not select an offline engine.
