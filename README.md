# CGMS

CGMS (Card Game Mafia State) is a card game about public power, private hands,
negotiation and lasting obligations. The [UI baseline](docs/design/UI_BASELINE.md)
preserves cgmsart and one compact board across phone, tablet and desktop, with
four equal public quadrants and the private hand directly below. Uniform hand cards
stack horizontally by suit; item sizing respects available width and height.
Flutter web is the current development/testing platform, using the existing
[host and reusable package](client/README.md). Local demonstration scenarios and
the online route share this composition. Native Android/iOS verification remains
mandatory at the final pre-release stage.

## Current scope

The shared Go rules engine and CLI simulator live in `backend/`. The internal
PostgreSQL match adapter persists exact finances, pending decisions, private
chance outcomes and idempotent command results. It is tested against a disposable
real database. Phase 2 adds opaque sessions, atomic rooms/start, typed HTTP commands,
private card handles, authenticated cursor replay/WebSockets and bounded Redis
presence fallback. See the [online API and local run guide](docs/code/API-game.md).
Local paid/free allowance and dirt-pass boundaries have adopted decisions and
real-authority verification; live commerce, public deployment and production capacity
remain outside this acceptance.
The [Flutter client](client/README.md) connects `#/online` to the local authority;
see its [integration guide](client/INTEGRATION.md) for the presentation boundary.
Its [bot-match lobby](client/README.md#play-against-bots) lets one human play
against two or three server-controlled opponents using the same rules and saves.
Simulation implementation and evidence are documented in [sims](sims/README.md).
The experimental [offline comparison probe](client/offline_probe/README.md)
records host Go/Dart parity and a provisional shared-Go recommendation;
[offline_host](client/offline_host/README.md) provides verified local saves/tutorials.
Final native engine/device acceptance remains a separate release gate.

Run the durable adapter tests from the repository root:

```bash
python3 xops/postgres_integration.py --redis -p 1 -race ./internal/identity ./internal/rooms ./internal/matchstore ./internal/transport ./internal/delivery ./internal/presence
```

The runner requires Docker and the pinned local PostgreSQL image described in
[xops](xops/README.md), fails if unavailable, and cleans up its disposable resources.
Keep `-p 1` when combining packages: the match-store crash test deliberately
restarts the shared disposable database.

## Documentation

- [`docs/project/GAME_RULES.md`](docs/project/GAME_RULES.md) — canonical game rules, including accepted Q01–Q12 and F01–F10/N01–N07/T01–T03 amendments.
- [UI baseline v2](docs/design/UI_BASELINE.md) — adaptive UI/UX requirements, preserved theme, browser acceptance and change policy.
- [`docs/project/README.md`](docs/project/README.md) — completed charter, decision log and glossary; [earlier drafts](docs/design/drafts/README.md) remain historical.
- [`docs/code/ARCHITECTURE.md`](docs/code/ARCHITECTURE.md) — proposed Go authority, data ownership and deployment boundaries.
- [`sims/README.md`](sims/README.md) — CLI simulation, reproducibility and bot evaluation contracts.
- [`docs/design/RULES-IMPLEMENTATION-QUESTIONS.md`](docs/design/RULES-IMPLEMENTATION-QUESTIONS.md) — accepted decisions and their implementation/test obligations.
- [`docs/reports/2026-09-28-final-draft-third-assessment.md`](docs/reports/2026-09-28-final-draft-third-assessment.md) — third frontend/backend assessment; its recommendations are now adopted, and all three reports remain historical evidence.
- [`AGENTS.md`](AGENTS.md) — rules every AI coding assistant follows in this repo.
- [`docs/planning/ROADMAP.md`](docs/planning/ROADMAP.md) — the plan.
- [`docs/tracking/README.md`](docs/tracking/README.md) — how the tracking log works.
- [`docs/tracking/context.md`](docs/tracking/context.md) — project context pack.
- [`.agents/skills/README.md`](.agents/skills/README.md) — curated skill library.

## Common commands

```bash
make help            # list available targets
make up              # build/start Flutter web (Nginx), backend, PostgreSQL, Redis and enabled telemetry
make down            # stop all services; retain database volumes
make restart         # rebuild/recreate the configured stack; retain data
make git.dry         # preview pending commits (read-only)
make git             # commit pending tracking rows + push
make track.add ACTION=note SUMMARY="..."
```

## License

See [`LICENSE`](LICENSE).
