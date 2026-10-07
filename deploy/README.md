# Local operational stack

This stack is a single-process CGMS authority. It does not establish release
readiness, a public deployment, store integration or native acceptance.
Use the repository root. Docker Compose, Go and Flutter must already exist.
Nothing installs OS packages or changes the workstation configuration.

## First start and service configuration

```
python3 xops/install_flutter_local.py  # Linux x64; verified 3.47.5 archive
export PATH="$PWD/.local/toolchains/flutter-3.47.5/flutter/bin:$PATH"
make up
```

`make up` initializes missing private secrets (preserving existing values),
builds Flutter web and the Go backend, then starts Nginx, backend, PostgreSQL,
Redis and any enabled telemetry services with bounded readiness checks.
`make restart` performs the same builds and forces container recreation so
code and configuration changes take effect. Both retain persistent volumes.
`make down` stops all services, including telemetry, retaining containers and
volumes. These are aliases for `services.up`, `services.up` and `services.stop`;
all accept `MODE=local` (default) or `MODE=production`. Production still requires
the TLS files described below. `services.init` remains available separately.

The existing system Flutter installation remains untouched. The local pin passed real reverse-proxy hot reload (293 ms) after the original
3.38.5 baseline timed out. Hot restart completed in 390 ms but emitted one
transient disposed-view debug assertion; reload and a fresh page remained
interactive. When switching SDKs, remove only ignored `build/unit_test_assets` and
`build/test_cache` if tests report incompatible compiled shader assets, then
rerun tests. Do not alternate SDKs concurrently in one checkout.

The client restores cookie sessions automatically with up to eight read attempts
(two-second request deadlines and 250 ms–4 s backoff, under 32 seconds per session
read recovery). During recovery it displays the current attempt and disables
session actions. An absent cookie is a normal signed-out state; authentication
rejections do not loop. Malformed successful responses are terminal and cannot
replace an existing account or pending command. Established match streams retain
the last confirmed projection and reconnect with up to eight consecutive
attempts. A validated state event or a socket that remains open for five seconds
resets that outage budget; merely returning HTTP 200 does not. Exhaustion shows a connection error
and an explicit retry control. Mutations are never replayed by this recovery;
unknown operations keep their original receipt/identity safeguards.

The local web address is `http://localhost:8080`. Do not open a second backend
against the same database; startup enforces one database-wide authority lease.

`config/backend/local.json` and `production.json` are complete service overlays,
validated by `backend/internal/config` against the documented
`config/schemas/service.schema.json` contract. Unknown or duplicate fields,
missing/null settings, bad listener ports, insecure production origins and bad
export endpoints fail validation/startup. `CGMS_CONFIG` selects this strict mode.
The old environment-only launch remains supported for existing integration tests.
Compose only wires files, volumes and networks; PostgreSQL, Redis, Nginx and
telemetry values live under `config/`.

`services.init` creates `.local/secrets/` without replacing existing values.
The database password and URL and the Redis password/config and URL are separate
private files. Do not print them or commit `.local/`. Backend files must not be
world-readable. Grafana's password is group-readable only so its configured
supplementary group can read it. Docker Compose file secrets are bind mounts;
source permissions matter, and `uid/gid/mode` declarations do not remap them.
The backend uses the invoking user's UID/GID. PostgreSQL durable storage lives
in the `cgms_postgres_data` named volume; Redis deliberately has no persistence.
The restricted Docker build context includes only the compiled server binary.

## Telemetry

Set `observability.enabled` in the selected backend file and run
`python3 xops/makefile/service_ops.py up --config <file>`. This same file drives
both the backend exporter and Compose's `telemetry` profile. Enabling starts
collector, Tempo, Prometheus and Grafana; disabling stops those containers,
preserves their volumes, retains logs/health/metrics and creates no exporter.
Use `services.up` when changing the setting; a web-only restart cannot change
a running backend's instrumentation. Grafana is authenticated and bound only
to `127.0.0.1:3000`; collector, Tempo, Prometheus, PostgreSQL and Redis have no
published ports. Nginx denies `/internal/`, including backend metrics.

Boundary logs and spans accept only fixed event/outcome names, internally minted
operation IDs, UTC times and durations. They never record URLs, SQL, client IDs,
raw errors, commands, hands, seeds, cards, tokens, cookies or receipt payloads.
Successful boundaries and expected cancellations use DEBUG, below the default
INFO threshold; failures, deadlines and unknown commits use WARN. Load tests
explicitly opt into DEBUG capture for latency distributions. Metrics and
enabled traces still record every boundary. The automatic-match scheduler keeps
a bounded cache of idle match versions to avoid restoring unchanged matches on
every scan; changed versions, process restart and entitlement wakeups recheck work.
Prometheus scrapes bounded counter/duration series and runtime heap/goroutine
measurements. Engine timing wraps the adapter call: the pure engine has no clock
or telemetry import. Queue and row-lock waits and unknown commit outcomes are
observable. Exports have a 256-span queue, one-second timeout, disabled retries,
and a 60-second failure cooldown with one sanitized log per attempted batch.
Spans can be dropped under saturation/loss; telemetry cannot establish durability.

No production Flutter telemetry SDK is selected or integrated. The isolated
web-only candidate spike measures span overhead in Chromium; see
[validation evidence](VALIDATION.md#flutter-instrumentation-decision-evidence).
Native support, export failures and frame-time overhead remain unverified.

## TLS and web profiles

`make services.up MODE=production` uses the production overlay and listens on
loopback `https://localhost:8443`. Supply `.local/secrets/tls_certificate` and
`.local/secrets/tls_key`; use a real hostname/origin/certificate for an authorized
public deployment. The checked-in production origin is a local validation
fixture, not a domain selection. Do not disable certificate validation for
public use. The TLS container health probe deliberately verifies local process
readiness independently of the certificate hostname; the integration probe
performs certificate/hostname verification using its explicitly supplied CA.

Local HTTP omits HSTS/CSP and permits development source maps.
After a local `services.up`, `make web.dev` runs Flutter's web-server on the
private Docker bridge and uses a same-origin Nginx development overlay for its
scripts, source maps and debug WebSockets. Type `r` plus Enter for reload,
`R` for restart or `q` to stop. It restores the static local web profile on exit.
Do not run `web.dev` while the production profile is active. Production uses
TLS 1.2/1.3, HSTS, same-origin API and WebSocket routes, `nosniff`, no referrer,
CSP and denies source maps. Build with `--no-web-resources-cdn` so Flutter's
CanvasKit/worker assets remain same-origin. The CSP allows WebAssembly compilation
and blob workers; it grants no inline script permission. A fresh Chromium in-app browser rendered the full app under this exact CSP
without console errors in a loopback HTTP fixture; certificate-verified HTTPS
protocol tests are separate. The current CanvasKit build does not exercise a
WebAssembly worker renderer; its policy permission is not a worker execution test.

Nginx resolves backend container addresses through Docker DNS every five seconds.
Recreating the backend therefore does not leave Nginx pinned to a dead IP.
The dedicated Compose settings opt into `proxy_client_ip`; Nginx overwrites
`X-Real-IP` so each real client has its own admission bucket. Never enable this
setting on a backend reachable by untrusted direct clients. Legacy/direct mode
ignores forwarded IP headers. Quiet game sockets send a bounded heartbeat every
20 seconds, below Nginx's 70-second read timeout. Cookies, invitations and
WebSocket headers never enter Nginx access logs.
The integration script opens a real three-seat match and verifies authenticated
WebSocket upgrade through Nginx, with an optional trusted local TLS certificate.

## Rebuild, restart and verification

```
make up                    # initialize, build and start the configured stack
make down                  # stop the whole stack; retain containers and volumes
make restart               # rebuild and recreate the whole stack; retain data
make go.re                 # static server build; bounded backend recreation
make web.re                # Flutter local-assets build; web recreation
make service.re SERVICE=redis
make services.config
make services.stop         # retains named volumes
python3 -m unittest discover -s xops/makefile -p test_service_ops.py
python3 xops/test_operations_integration.py --failures
python3 xops/test_operations_integration.py --base https://localhost:8443 \
  --certificate .local/secrets/tls_certificate --restart --idle 75
```

Pass `MODE=production` consistently while the production profile is running.
Readiness is bounded, command output and failure/timeout evidence are retained
under private `.local/logs/`, and no operational command removes a volume.
`services.up` recreates services to apply changed mounted configs. `go.re` and
`service.re` recreate only their selected service. Repeated backend recreations
must retain sessions and room/match IDs; the `--restart` probe checks this twice.
After `web.re`, hard-reload the browser; if an old service worker controls the
page, unregister it and clear this site's cache via developer tools, then reload.
This is an explicit local browser action, not a server promise to evict caches.

Cold-start browser acceptance uses the actual local Compose stack:

```bash
xops/agent/safe-run.sh startup-browser -- node client/tool/startup_browser_acceptance.cjs
node --test client/tool/startup_browser_acceptance.test.cjs
```

Run it after building the client, serially with other tests using port 8080.
It stops all project services and runs the documented `services.up` three times,
retaining PostgreSQL volumes and prior data. Each run uses a fresh Chromium
context, checks served artifact hashes, and records service/browser timings.
Health readiness is measured independently from service-command completion;
normal-entry and online-usability timings begin at browser navigation.
Additional cases stop only the backend to exercise delayed session readiness,
reload during an outage, and automatic established-match reconnect. The harness
creates ordinary guest accounts and a three-seat match in the retained local
store; it does not delete them. Private service logs and JSON browser diagnostics
are written under ignored `client/.browser-tools/evidence/`; no screenshot is
required. Deliberate outage diagnostics are admitted only for the exact read or
stream path and observed status/error; asset and application errors fail.

Run database integration packages serially: the server and transport suites
exercise the same database-wide singleton lock and must not run concurrently
against one disposable database (`-p 1`). Do not weaken the singleton check.

Container image digests were resolved from the actually pulled images. Upstream
API/configuration references checked during implementation:
[OTel Go](https://opentelemetry.io/docs/languages/go/),
[OTLP HTTP exporter](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp),
[Collector Docker](https://opentelemetry.io/docs/collector/install/docker/),
[Tempo](https://grafana.com/docs/tempo/latest/docker-example/),
[Compose secrets](https://docs.docker.com/compose/how-tos/use-secrets/), and
[Nginx WebSockets](https://nginx.org/en/docs/http/websocket.html).


## Private backup and isolated restore rehearsal

```
python3 xops/makefile/backup_ops.py backup
python3 xops/makefile/backup_ops.py drill
python3 xops/makefile/backup_ops.py drill --backup .local/backups/<run>
```

Backups contain private game/session state and are created under mode-0700
`.local/backups/` with private files. They use one consistent PostgreSQL logical
snapshot, preserve every table and copy explicit service/rule files plus the
workspace server binary into a SHA256 manifest. These filesystem artifacts are
explicitly unverified workspace snapshots: they may differ from the running
container, particularly after a build or with a custom/production overlay. The
database snapshot itself preserves each match’s pinned rules/configuration; the
operator must retain deployed images/config separately for runtime recovery. Credentials are not copied: retain secret
files separately in an operator-managed secure recovery system. Local backups
share the workstation failure domain and are not off-host disaster protection.
Only restore trusted archives; checksums detect accidental modification, not
malicious replacement of both archive and manifest.

The drill verifies file hashes and restores into a fresh, network-isolated
PostgreSQL container with a 512-MiB ephemeral data filesystem, then removes that
container. It never overwrites the running database or removes its volumes.
This is a bounded small-fixture drill, not a general large-database restore tool.
The report records elapsed time, restored tables/matches/receipts/debts and
constraint validation. The separate matchstore integration fixture compares
all restored table rows, exact debt rationals, pinned configuration, snapshots
and duplicate command receipts. Production recovery still requires a tested
secret restoration process, off-host retention, capacity and application replay.

RPO 15 minutes / RTO 60 minutes remain proposed acceptance targets until owner
adoption. A logical snapshot omits transactions acknowledged after its snapshot;
host/storage loss can also destroy every backup kept on that host. No automatic
backup scheduling or production recovery policy is installed by these commands.
References: [pg_dump consistency](https://www.postgresql.org/docs/16/app-pgdump.html)
and [pg_restore](https://www.postgresql.org/docs/16/app-pgrestore.html).


## Version and license inventory (2026-09-30)

Compose pins image content digests; tags below identify the versions inspected.
The backend image contains the locally built binary, so its binary hash and
source revision must also be retained for release provenance. This list covers
primary products, not a complete image SBOM or transitive license audit.

| Component | Inspected version | Primary license source |
|---|---|---|
| PostgreSQL | 16.13 | [PostgreSQL License](https://www.postgresql.org/about/licence/) |
| Redis | 7.4.8 | [RSALv2 or SSPLv1](https://redis.io/legal/licenses/); this version is source-available, not BSD licensed |
| Nginx | 1.30.5 stable | [BSD-style license](https://nginx.org/LICENSE); upgraded from the initial 1.27.5 fixture |
| OTel Collector | 0.161.0 | [Apache-2.0](https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/v0.161.0/LICENSE) |
| Prometheus | 3.10.0 | [Apache-2.0](https://raw.githubusercontent.com/prometheus/prometheus/v3.10.0/LICENSE) |
| Tempo | 2.10.3 | [AGPLv3](https://grafana.com/licensing/) |
| Grafana OSS | 12.4.1 | [AGPLv3](https://grafana.com/licensing/) |
| Backend base | Alpine 3.24.2; digest `294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6` | Individual package licenses; not one distribution-wide license. [Supported branch](https://alpinelinux.org/releases/) and [official image metadata](https://github.com/docker-library/official-images/blob/master/library/alpine). |
| Flutter local toolchain | 3.47.5 / Dart 3.13.4 | Bundled LICENSE files; official archive checksum pinned in installer |

The Nginx stable version and official image tag were checked against
[upstream downloads](https://nginx.org/en/download.html) and
[Docker Official Images metadata](https://github.com/docker-library/official-images/blob/master/library/nginx).
Alpine 3.20 was beyond routine support and has been replaced with the supported
3.24.2 base. The 3.24 branch currently lists main/community support and an end
date of 2028-06-01; future community-package policy must still be monitored.
The [3.24 release notes](https://alpinelinux.org/posts/Alpine-3.24.0-released.html)
were reviewed; this minimal image runs a statically built Go binary and its
BusyBox health check, not the changed desktop/Python/bootloader components.
A fresh host rehearsal, complete SBOM/security scan and operator acceptance
are still required before R01 release completion.


For an isolated fresh-workspace/secrets/volume production rehearsal, run
`python3 xops/rehearse_clean_stack.py`. It uses the existing Docker daemon,
loopback ephemeral HTTPS port and local test certificate, retains a private
report under `.local/rehearsals/`, removes only its project containers/networks,
and retains its own PostgreSQL volume. This is not a fresh OS/host rehearsal.
For archived-HEAD reader compatibility, run `python3 xops/rollback_rehearsal.py`;
it compiles into `.local/rollback/` and uses a disposable test database. Its
scope includes rejection of economy-enabled matches by the archived reader,
while the current reader preserves the pinned economy configuration and state.
Existing matches without economy remain compatible with both readers.
See [measured validation](VALIDATION.md).
