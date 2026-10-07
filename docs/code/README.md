# 📘 `docs/code/` — codebase documentation

Per-module documentation that orients a new contributor (human or agent) on
**what** a piece of the codebase does, **why** it exists, and **how** to
extend it without breaking invariants.

## Files

- [Current game rules](../project/GAME_RULES.md) and [decision/test register](../design/RULES-IMPLEMENTATION-QUESTIONS.md) — accepted gameplay authority, Q01–Q12 decisions and adopted F01–F10/N01–N07/T01–T03 interaction obligations.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — proposed modular Go backend, engine, configuration and deployment boundaries.
- [`API-game.md`](API-game.md) — implemented local authenticated commands, safe projections and reconnect contract; production readiness remains open.
- [`MODULE-game.md`](MODULE-game.md) — implemented shared game engine and durable adapter, boundaries, test evidence and remaining integration work.
- [Third frontend/backend assessment](../reports/2026-09-28-final-draft-third-assessment.md) — historical source of the adopted T01–T03 rules; no runtime validation is implied.

For another significant module, follow the sections in `MODULE-game.md` and
distinguish implemented behavior from proposed contracts.

## When to write code docs

- A module exceeded ~500 lines and the agent had to read 3+ files to
  understand it.
- A new contract was introduced (RPC, event schema, public API).
- A non-obvious invariant exists that, if violated, would break things
  silently (race condition, idempotency requirement, etc.).

For one-off scripts or trivial helpers: no doc needed.
