# 🛠 `xops/` — agent & makefile ops

Everything in this tree is **plain bash or `python3` stdlib** — no
third-party deps, no virtual envs to set up. Cross-platform where the
emoji and ANSI escapes don't matter (Linux + macOS first-class; Windows
runs via Git Bash / WSL).

## Layout

```
xops/
├── README.md         ← you are here
├── init/             ← the scaffolder (framework-only; stripped from scaffolded projects)
│   ├── scaffold.sh
│   └── scaffold_codex.py
├── agent/            ← runtime scripts agents call directly
│   ├── tracking_append.sh
│   ├── safe-run.sh
│   ├── session-bootstrap.sh
│   └── run-with-retry.sh
├── lib/              ← shared bash helpers (emoji logger)
│   └── log.sh
└── makefile/         ← python3 dispatchers the Makefile calls
    ├── _common.py
    ├── git_ops.py
    ├── track_ops.py
    ├── roadmap_ops.py
    └── codegraph_ops.py
```

## Conventions

- **All scripts log with emojis.** The shared logger is
  [`lib/log.sh`](lib/log.sh) for bash; Python scripts use the constants in
  [`makefile/_common.py`](makefile/_common.py).
- **Exit codes matter.** `0` = ok, `1` = expected-failure (e.g. "no
  pending rows"), `2+` = real error.
- **All scripts are idempotent.** Re-running them on a clean state is a
  no-op.
- **Atomic writes.** Anything that touches `docs/tracking/tracking.csv` or
  `docs/tracking/state/*.json` uses `flock(1)` or `os.replace()`.

## Key scripts

The framework-only `init/scaffold_codex.py` generates Codex agent TOML,
prompt-backed skills and correctly escaped MCP configuration. It needs Python
3.9+ and only the standard library. The scaffold tests additionally use Python
3.11+'s `tomllib` to parse generated TOML.

| Path | Purpose |
|---|---|
| [`agent/safe-run.sh`](agent/safe-run.sh) | Crash-safe wrapper for risky commands. Output survives a killed terminal. |
| [`agent/session-bootstrap.sh`](agent/session-bootstrap.sh) | Print orienting context at agent session start. |
| [`agent/tracking_append.sh`](agent/tracking_append.sh) | Validated, atomic CSV appender for `docs/tracking/tracking.csv`. |
| [`agent/run-with-retry.sh`](agent/run-with-retry.sh) | Wrap a flaky command in bounded retries with backoff. |
| [`makefile/_common.py`](makefile/_common.py) | Shared helpers for the Python make dispatchers. |
| [`makefile/git_ops.py`](makefile/git_ops.py) | `make git` / `make git.dry`. |
| [`makefile/track_ops.py`](makefile/track_ops.py) | `make track.add` / `make track.list`. |
| [`makefile/roadmap_ops.py`](makefile/roadmap_ops.py) | `make roadmap.status`. |
| [`makefile/codegraph_ops.py`](makefile/codegraph_ops.py) | `make codeg` — initialize or update the CodeGraph index. |

## Add a new Makefile target

1. Add a thin target to the root `Makefile` (one line, dispatches to
   `xops/makefile/<module>.py <subcommand>`).
2. Add (or extend) the Python module in `xops/makefile/`.
3. Import `_common` for the emoji logger + standard subprocess helpers.
4. No new dependencies — `python3` standard library only.

## Add a new agent script

1. Create `xops/agent/<name>.sh` (`set -euo pipefail`, `source lib/log.sh`).
2. Make it executable (`chmod +x`).
3. Document it in this README and in the appropriate skill file under
   [`.agents/skills/`](../.agents/skills/).

## Phase 0 simulator helpers

For P01 native-candidate measurements, `python3 xops/offline_probe.py RUN_ID`
builds a local C shared library, exports synthetic Go fixtures, checks Dart FFI
parity in two fresh processes and writes a hashed report beneath ignored
`sims/artifacts/offline-probe/`. It requires existing Go, Dart and C tools and
does not install dependencies. Optional `--android-ndk PATH` cross-compiles ARM64;
it does not claim device execution. See the [probe guide](../client/offline_probe/README.md)
for scope, ABI ownership, reproduction and outstanding P01 gates.

- `sim_build.py BUILD_ID`: pinned offline Go build in two fresh paths, provenance and
  binary equality; outputs only under ignored `sims/artifacts/build/`.
- `sim_private.py SPEC.json`: protected argv/output capture for seeded workloads.
  Pass only the specification path through `agent/safe-run.sh`.
- `sim_evaluate.py BINARY OUTPUT_ID [--samples N]`: verify expected complete/partial
  statuses, replay, worker equivalence and public reports; no full-game claim.
- `sim_schema.py [--check]`: regenerate or verify closed runtime JSON shapes.

Use fresh output IDs. See [host/container instructions](../sims/README.md#running-the-implemented-scope).

## Complete-game evaluation

- `sim_gameplay_evaluate.py BINARY PRIVATE_PLAN OUTPUT_ID [--resume | --analyze-only]`: frozen,
  population-interleaved seed-block campaign, protected execution, manifest/input
  verification, completed-game replay, worker equivalence and clustered analysis.
  Preserves all planned/incomplete units and refuses replacement outputs.
- `python3 -m unittest discover -s xops -p 'test_sim_gameplay*.py'`: campaign
  admission, integrity, denominator, cancellation/resume and analysis checks.

See the [preregistered design](../sims/experiments/gameplay-heldout-v1.md); a
resource-limited campaign is not a strength or balance conclusion.

`--analyze-only` verifies saved frozen inputs, shard manifests and exact canonical
score objects, then writes separate immutable reanalysis files. It does not run
games, change campaign state, renew budgets or replace historical reports.

## PostgreSQL integration tests

For local browser QA, first build the existing Flutter host, then run
`python3 xops/postgres_integration.py --webfixture` through `agent/safe-run.sh`.
It serves that build and the real API at `http://127.0.0.1:8081`, using a
disposable database and isolated schema for at most 3600 seconds. Ordinary
integration test runs retain their 900-second bound. Stop the fixture with
SIGTERM/SIGINT; the runner removes its own container, network and volume.
The development-only `backend/cmd/webfixture` must never be deployed. A same-origin
JSON `POST /__fixture/session` with `scenario`, `players` (3/4) and `seat` (1-based)
selects a synthetic account via an HttpOnly cookie and returns `match_id`.
Scenarios are `trade`, `settlement`, `opening`, `purchase`, `combat`, `decision`,
`coup`, `loan`, `confinement`, `justice`, `attack`, `final-settlement` and
`compensation`, `card-style`, `opening-large`, `turn-draw-ace` and `turn-draw-diamond`;
separate browser contexts represent
separate seats. An optional alphanumeric/hyphen `run` (at most 32 characters)
isolates browser runs; all seats sharing a run use the same fixture match.
At most 128 fixture groups can be created per server lifetime. These fixtures
prove targeted interactions, not full normal-deal games. The helper refuses
nonliteral/nonloopback binds, foreign Host/Origin and unsupported scenarios.
No session token or database credential appears in URLs or response JSON.
The two turn-draw fixtures start at a human boundary with deterministic supply;
the real End Turn and server BeginTurn still perform the draw. They exercise
concealed-Ace versus ordinary Diamond acquisition without synthetic UI events.
`opening-large` holds both physical copies of all nine Clubs numbers for an
18-card first-opening preview and enlarged-text/scrolling acceptance.

Run `python3 xops/postgres_integration.py -race ./internal/matchstore -timeout=3m`
through `agent/safe-run.sh` (replace package with the integration package). The
runner adds `-tags=integration -count=1`, defaults Go timeout to five minutes and
enforces a 900-second child wall deadline. Docker, the pinned locally available
image, startup or cleanup failures fail the run; they never skip tests.

A unique container, bridge network and disk-backed Docker volume are removed on
normal exit, failure or handled SIGINT/SIGTERM. SIGKILL of the runner cannot execute
cleanup; leaked resources are identifiable by `cgms-integration-` names. No image
is pulled automatically. The pinned digest in the script was inspected from the
local `postgres:16-alpine` image. PostgreSQL durability settings remain enabled.
The generated password is passed through environment, never command arguments;
the port is published only on `127.0.0.1`. Tests receive
`CGMS_TEST_DATABASE_URL` and `CGMS_TEST_POSTGRES_CONTAINER`; do not log the URL.
The latter permits tests to kill/start their exact disposable server and validate
WAL recovery. `--self-test` verifies a committed row survives SIGKILL/restart.

The pgxpool dependency is pinned to `github.com/jackc/pgx/v5 v5.11.0`, verified
against the [official project](https://github.com/jackc/pgx) and
[release](https://github.com/jackc/pgx/releases/tag/v5.11.0); Go checksum
verification is recorded in `backend/go.sum`.

Add `--redis` for the optional presence integration suite, for example
`python3 xops/postgres_integration.py --redis -race ./internal/presence`.
This creates a separate loopback-only, password-protected Redis container from
the locally verified digest pinned in the runner, with 128 MiB container memory
and 64 MiB key memory limits. It exports `CGMS_TEST_REDIS_URL` and
`CGMS_TEST_REDIS_CONTAINER` only to tests and removes Redis with the PostgreSQL
resources. Never log either database URL. Redis has no persistence: tests kill
and restart it to verify bounded local presence fallback; PostgreSQL remains
sole authority for commands, finance and recovery. Production Redis URLs use
private plaintext `redis://` transport and must stay on a trusted local network.

Online dependencies are pinned to `github.com/coder/websocket v1.8.15` and
`golang.org/x/crypto v0.57.0` after checking the
[official WebSocket project](https://github.com/coder/websocket) and
[Go Argon2 package](https://pkg.go.dev/golang.org/x/crypto/argon2).

## Service operations

The [operational stack](../deploy/README.md) documents `up`, `down`, `restart`, `services.init`,
`services.up`, `services.stop`, `services.config`, `go.re`, `web.re`,
`service.re` and `web.dev`. Thin Make targets dispatch to `makefile/service_ops.py`
and `makefile/web_dev.py`; private command/failure logs live under `.local/logs/`.
The commands retain persistent volumes. See the deployment guide for local/TLS
profiles, secret-file setup and the limits of the current verification evidence.
`up` and `restart` use `services.up`: initialize missing secrets, build Flutter
web and Go, then recreate the configured stack and wait for readiness. `down`
uses `services.stop`, stopping core and telemetry services without deleting
containers or volumes. `MODE=production` selects the existing TLS overlay.
