# Operational validation — 2026-09-30

This is local evidence, not production approval or full roadmap completion.
Commands are documented in [README.md](README.md). All private archives,
credentials, logs and binaries remain under ignored `.local/` or agent run logs.

## Runtime and bounded load

Host: Intel Core Ultra 9 285H, Linux, 16 visible CPUs, 31,811,008 kB RAM;
Go 1.27.1, `GOMAXPROCS=4`. Disposable PostgreSQL 16.13 runs with a 2-CPU,
512-MiB limit; the Go test process shares the workstation. This is not a dedicated
4-vCPU/8-GiB host measurement. All timings include local host contention.

`TestOperationalCapacityAndTelemetryLoad` creates 100 four-seat fixtures and
400 authenticated WebSocket attempts, followed by 300 admitted lightweight
nullification-consent transitions at 20/second over 15 seconds per case.
It uses real HTTP, WebSocket, PostgreSQL, dispatch queues and engine adapters.
It does not model a complete gameplay mix, slow internet clients, lengthy
settlement work, long-duration capacity or mobile devices. Logs are serialized
into an in-memory measurement writer; exporter-on uses a local OTLP sink, not
production storage or network latency. Reported HTTP response latency includes
commit and response serialization; it is not an isolated commit timer.

| Explicit stream cap / tracing | Admitted / rejected sockets | HTTP p50 / p95 / p99 (ms) | Transaction p95 (ms) | Engine p95 (ms) | Queue p95 (ms) | Lock p95 (ms) | Process RSS (kB) |
|---|---:|---:|---:|---:|---:|---:|---:|
| 64 / off | 64 / 336 (429) | 10.252 / 19.021 / 40.604 | 16.671 | 0.561 | 2.073 | 0.614 | 36,984 |
| 400 / off | 400 / 0 | 9.111 / 13.480 / 17.528 | 11.779 | 0.420 | 1.772 | 0.501 | 69,488 |
| 400 / on | 400 / 0 | 8.907 / 14.796 / 20.013 | 12.274 | 0.401 | 2.219 | 0.682 | 70,864 |

Pool capacity was 16; empty-acquisition counts during the workloads were 5, 6,
and 6; aggregate acquisition durations were 20.06, 26.11 and 25.81 ms. These
counts do not establish database saturation thresholds. Exporter-on p95 response
latency was approximately 9.8% higher than off in this one sequential pair;
repeat randomized longer trials before attributing that difference solely to
instrumentation. No load target is promoted to an SLA from these samples.
`max_streams` is an optional validated 1–4096 setting, default 64; the 400 case
uses an explicit test configuration. Default admission rejection is a policy
limit, not evidence of physical inability to serve 400 sockets.

Evidence: `operations-capacity-400--20260930T204744Z-377075.log`.

## Sustained mixed authority workload with 400 live subscriptions

`TestOperationalGameplayMix` supplements the earlier consent-only probe with
100 synthetic four-seat matches and 400 authenticated, continuously drained
WebSocket subscriptions. Setup opens a Diamonds series with all three ordered
responses. Each measured cycle declares a real card offer, declines it, and
transfers one unit of available cash in each direction through the durable
financial scheduler. After pressure, all matches complete an accepted card
transfer and end a turn; card ownership and conserved balances are asserted,
and one complete canonical journal is replayed. This is a repeatable business-
operation mix, not naturally dealt full games or combat/large-debt coverage.
It does not consume the P03 bot-research allocation or estimate playing strength.

Each tracing cohort schedules 30-second stages at 20, 80 and 200 commands/second.
A maximum of one cycle per match prevents an unbounded client backlog. Each actor
makes at most four HTTP attempts/second, below the existing five/second rate
policy. GETs for current versions and financial-continuation polling add load;
reported command latency is the successful POST attempt including its response,
not client pacing, automatic-wait time or earlier rejected attempts. Only explicit
429 non-commit responses are retried, at most five attempts with identical body
and command identity. Other errors fail the test. All refusals and skipped
scheduling opportunities remain visible.

| Tracing / target commands per second | Successful commands | Stage plus drain (s) | Skipped cycles | HTTP 429 | POST p50 / p95 / p99 (ms) | RSS (kB) |
|---|---:|---:|---:|---:|---:|---:|
| Off / 20 | 600 | 31.547 | 0 | 0 | 13.869 / 24.702 / 45.573 | 77,872 |
| Off / 80 | 2,400 | 31.560 | 0 | 0 | 13.522 / 35.685 / 86.247 | 78,932 |
| Off / 200 | 2,744 | 34.582 | 814 | 212 | 22.704 / 177.074 / 354.555 | 83,732 |
| On / 20 | 600 | 31.550 | 0 | 0 | 10.954 / 17.878 / 44.672 | 83,492 |
| On / 80 | 2,400 | 31.565 | 0 | 0 | 12.275 / 19.560 / 33.294 | 83,500 |
| On / 200 | 3,612 | 32.913 | 597 | 334 | 17.233 / 83.896 / 503.567 | 83,588 |

| Tracing / target | Pool peak / 16 | Empty acquisitions | Aggregate pool acquire time (ms) | Queue p50 / p95 / p99 (ms) | Transaction p95 (ms) |
|---|---:|---:|---:|---:|---:|
| Off / 20 | 10 | 0 | 15.994 | 0.005 / 0.020 / 1.945 | 20.905 |
| Off / 80 | 14 | 255 | 57.991 | 0.004 / 1.815 / 3.001 | 26.192 |
| Off / 200 | 16 | 7,633 | 35,560.478 | 1.214 / 5.066 / 20.949 | 141.287 |
| On / 20 | 9 | 0 | 14.350 | 0.004 / 0.015 / 1.642 | 15.469 |
| On / 80 | 12 | 0 | 18.098 | 0.004 / 1.770 / 3.184 | 15.760 |
| On / 200 | 16 | 7,073 | 61,505.782 | 1.009 / 3.820 / 21.894 | 46.920 |

The 200/second schedule is **not achieved**: completed throughput including drain
is approximately 79.3/second off and 109.7/second on, with both the 16-connection
pool and 32-active-HTTP admission bound reached. It must not be adopted as a
capacity target for this workload/configuration. The 20/second proposed load and
80/second schedule completed without HTTP refusals in both cohorts; including
final draining they deliver approximately 19.0 and 76.0 commands/second.
Queue and pool measurements include automatic scheduler work, not just requests.

Hardware and PostgreSQL limits are the same as above. These remain sequential
samples on a shared development host: frontend builds/browser checks and separate
service restarts overlapped portions of the run. The CPU-heavy offline race gate
finished before these cohorts. Lower tracing-on latency is not evidence that
instrumentation improves performance; a causal overhead estimate remains
inconclusive without repeated randomized runs on an isolated host. The local OTLP
sink does not model a remote collector outage or production storage cost.

Evidence: `mixed-live-subscriptions--20260930T212802Z-556603.log`: all six stages,
all 100 ending assertions per cohort, and journal replay passed in 263.890 seconds
overall (132.51 / 131.37 seconds per cohort). Every cohort retained 400 live
subscriptions; 38,852 / 44,035 ordered, authorized event frames were validated.
The preceding HTTP-only cohort, `gameplay-mix-sustained--20260930T212333Z-535675.log`,
passed 20/80 stages but failed at 200 when its original client treated admission
429 as a fatal result. That failure is retained; the combined test measures
bounded retry and scheduling behavior explicitly rather than claiming no overload.
Run the default short integration fixture for correctness, or opt into measured
stages with `CGMS_GAMEPLAY_STAGE_SECONDS=30` and a nine-minute test timeout.

## TLS, proxy, telemetry and failure boundaries

Nginx was updated from 1.27.5 to the verified stable 1.30.5 image digest recorded
in Compose. A trusted local test CA and hostname verification passed HTTPS
readiness/assets, security headers, origin rejection, private-metrics denial,
three-seat match start, authenticated WebSocket upgrade, quiet socket survival
past 75 seconds, and two backend recreations preserving session/room/match state.
Redis loss preserved PostgreSQL-backed authenticated reads. PostgreSQL outage
refused writes and readiness; recovery retained acknowledged state and repeated
room-command identity produced the identical result. The outage drill exposed
and repaired authentication dependency failures being misreported as 401;
they now return 503 without exposing database error details.

Evidence: `nginx-supported-tls--20260930T204432Z-362690.log`. These are process
and service failures on retained local storage, not destructive host/storage
loss, arbitrary packet faults or high-availability evidence. Load admission
and unit tests establish bounded queues/connections; no claim is made that
all failure combinations were tested. Collector loss/cooldown is tested at the
exporter boundary; enabled Compose was separately verified through actual Tempo
traces and Prometheus metrics.

Flutter 3.38.5 proxy reload timed out. A workspace-local, official checksum-pinned
Flutter 3.47.5/Dart 3.13.4 archive passes host and package tests, analysis and web
build. Real Nginx-proxied hot reload completed in 293 ms, and repeated successfully
in 390 ms after the Nginx 1.30.5 upgrade; hot restart returned
390 ms with one transient disposed-view debug assertion. The old SDK installation
is untouched. Fresh browser rendering under the exact production CSP passed
without console errors in a loopback HTTP fixture; HTTPS protocol validation
is separate. CanvasKit assets/WebAssembly load and source-map denial are tested;
the current build does not exercise a worker renderer.

Final restart regression checks reject explicit null optional settings and require
a selected telemetry service's private readiness endpoint, rather than merely
a running container. Two actual collector recreations passed in 2.148 and
0.639 seconds, followed by independent health probes; the original running
service set was restored and volumes retained. Evidence:
`ops-economy-verifier-collector--20260930T211328Z-499947.log`.

## V04 named acceptance for the shipped CanvasKit build

The current renderer is CanvasKit. A different skwasm renderer is not a condition
for accepting this build. Its emitted support files are routed and byte-checked,
without claiming that their worker paths execute.

| Criterion | Evidence / result |
|---|---|
| Same-origin bootstrap, application JS, CanvasKit JS/WASM and font routing | 17 emitted JS/WASM/font artifacts have correct MIME and SHA-256 equality with the current build; both trusted TLS and exact-CSP HTTP probes pass. |
| Emitted worker assets | `flutter_service_worker.js` has correct JS MIME and matching bytes. Worker/service-worker execution is not inferred. No separately named renderer-worker asset is emitted by this build. |
| Production source-map isolation | `/main.dart.js.map` returns 404. |
| Production CSP rendering | Final local-font build rendered under exact CSP with zero console/page errors and matching served digest; `p04-local-font-csp--20260930T213305Z-572563.log`, asset SHA-256 `82a3f9f7e765f1126ea1aa49ecf4af28c77a0af0e8f77213ccceacac70b01fba`. |
| HTTPS and browser policy | Trusted test CA and hostname verification, HSTS, foreign-origin rejection and private endpoint denial pass. |
| WebSocket forwarding and bounded idle survival | Authenticated 101, heartbeat survival beyond 75 seconds, and two backend recreations preserving durable state pass. |
| Development proxy | Workspace-pinned Flutter 3.47.5 hot reload through Nginx 1.30.5 completes in 390 ms; quit restores the static proxy. |
| Development restart limitation | Hot restart completes in 390 ms but emitted one transient upstream disposed-view debug assertion. This remains disclosed; it is not a production-rendering result. |

The asset probe exposed a real missing TTF mapping: Nginx served the fallback
font as `text/plain`. All three profiles now explicitly map `.ttf` to `font/ttf`;
the test retains MIME and byte assertions. Evidence:
`tls-font-mime--20260930T212456Z-543413.log` and
`csp-after-port-release--20260930T212803Z-556954.log`.
The local-profile recreation first hit a transient occupied port 8080; after
checking that the port had been released, recreation and the same probes passed.

## Grouped process restart on retained storage

`xops/test_operations_integration.py --group-restart` requires a quiescent,
acknowledged match projection before stopping backend, PostgreSQL and Redis
together. It verifies unavailability while stopped, starts all three with bounded
health checks, and compares the complete match projection, authenticated room
and a pre-stop room-command receipt after recovery. The interrupted-stop unit
test proves the recovery operation is attempted in `finally`, including an
uncertain stop result. The initial drill deliberately failed its equality check
because startup automatic work had not reached a stable checkpoint; the revised
probe waits for three equal snapshots before measuring recovery.

The actual grouped restart passed in **13.599 seconds**, with identical durable
state and idempotent receipt. Recovery uses `compose start --wait`, preserving
the existing container configuration rather than implicitly recreating it with a
different overlay. Evidence:
`grouped-retained-container-restart--20260930T213316Z-573401.log`.
The earlier quiescent drill also passed in 14.160 seconds.
This is a graceful service-process restart approximation on retained local
storage, not a workstation reboot, abrupt power loss, disk loss or an off-host
recovery claim. Redis, telemetry and database individual-outage evidence above
retains its original limits.

## Flutter instrumentation decision evidence

The isolated [spike](../xops/telemetry_spike/) pins Workiva `opentelemetry` 0.18.11;
it changes no production application dependency. Publisher status: traces beta,
metrics alpha, logs unimplemented, web platform. Flutter widget-test compilation,
release web compilation and actual Chromium execution pass. After 1,000 warm-up
spans, 30 blocks of 1,000 spans measured per-block mean microseconds/span (not individual-span tails):

| Mode | p50 | p95 | p99 |
|---|---:|---:|---:|
| No-op API provider | 0.2 | 0.4 | 0.6 |
| Recording provider, no exporter | 42.101 | 44.4 | 45.199 |

This isolates span creation/end, not export, frames, background lifecycle or
native device overhead. All spans contain a constant synthetic name and no game
state. No production Flutter SDK is selected: production adoption needs bounded
export/failure behavior and platform checks beyond this web-only candidate spike.
Source: [publisher documentation](https://pub.dev/packages/opentelemetry).

## Backup and rollback

The private managed-backup drill restored 26 tables and three operational matches
into a new network-isolated PostgreSQL container in 3.555 seconds. Its live
fixture had no financial rows or command receipts, so a separate real PostgreSQL
integration fixture supplies a pinned active match, exact long-denominator debt,
Coup command receipt and projections. A custom-format archive restored all 17
matchstore tables byte-equivalently (canonical row JSON), preserved private
checkpoint/configuration/version and returned the same result on duplicate retry.
The populated fixture used 52,539 bytes and 315.8 ms for dump/restore/comparison.
These small-fixture timings do not estimate large-database disaster recovery.

Evidence: `backup-restore-drill--20260930T202925Z-291212.log` and
`backup-fixture-green--20260930T203301Z-305469.log`. The proposed RPO 15 minutes /
RTO 60 minutes remain unaccepted until the owner adopts them. Local backup loss
shares the host failure domain; transactions after the snapshot are absent.
Secret recovery, off-host copies and destructive host/storage-loss drills remain
outside this evidence.

`xops/rollback_rehearsal.py` archives reader revision
`706454f0b1f9da81a69e0f124e5b057b6ee13b34`, compiles that reader without changing
Git state, and runs it beside the current reader against one active pinned
checkpoint. Both preserve state/version, replay compatible migration identities,
reject unknown migration/engine identities, and retain the original engine after
the unsupported fixture identity is removed. The compiled old reader is retained
under `.local/rollback/`. Existing transactional migration tests cover failed DDL
rollback; this change introduces no resumable data backfill.

Evidence: `rollback-engine-rehearsal--20260930T205429Z-405565.log` and
`economy-rollback-rehearsal--20260930T210851Z-477044.log`. The latter creates an
actual economy-enabled match and proves the archived reader rejects its new
envelope schema with unsupported-checkpoint status, while the current reader
retains identical state and pinned policy. Existing unconfigured matches remain
compatible. No destructive schema `Down` or financial-history deletion was used.

## Supported backend base refresh

Final dependency review found Alpine 3.20 beyond routine support (upstream end
2026-04-01, security patches only on request). The Dockerfile now pins the actual
Docker Official Image for Alpine **3.24.2** at
`sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6`.
The tag was checked against [official image metadata](https://github.com/docker-library/official-images/blob/master/library/alpine)
and pulled by Docker; `/etc/alpine-release` in the base and the rebuilt application
image both report 3.24.2. The [upstream support table](https://alpinelinux.org/releases/)
currently lists the 3.24 branch through 2028-06-01. The
[release notes](https://alpinelinux.org/posts/Alpine-3.24.0-released.html) were reviewed;
the server is a statically linked Go binary, so changes to desktop packages,
Python packaging and host bootloaders do not alter this container contract.

Installed package metadata records BusyBox/ssl_client 1.37.0-r31 and apk-tools/
libapk 3.0.8-r0 as GPL-2.0-only; musl 1.2.6-r2 as MIT; OpenSSL libraries 3.5.8-r0
as Apache-2.0; zlib 1.3.2-r0 as Zlib; and the certificate bundle 20260909-r0 as
MPL-2.0 AND MIT. The complete 16 installed package records were captured in
`alpine-base-package-record--20260930T213845Z-597484.log`; this is package metadata,
not a full release SBOM or a vulnerability scan.

`make services.config` rebuilt the binary through project tooling and passed
configuration validation. The existing isolated production rehearsal rebuilt the
container from the new digest and passed in **16.440 seconds**: fresh project,
secrets and PostgreSQL volume, trusted TLS hostname verification, guest/room
creation, private service ports, nonroot/read-only/capability-dropped backend.
Its containers/networks were removed; volume
`cgms-rehearsal-2360bebfa9_postgres_data` remains for the audit. The ordinary
8080 application and separate 8081 fixture were not restarted or modified.
Evidence: `alpine324-isolated-rehearsal--20260930T213852Z-598222.log`,
`alpine-built-image-identity--20260930T213951Z-604896.log` and
`alpine-base-application-gates--20260930T213912Z-601650.log` (server build plus
config/transport/telemetry tests passed). This repairs the identified stale base;
true fresh-host and full release inventory/security acceptance remain separate.

Independent verification repeated the new-base rehearsal in **14.452 seconds**
and inspected the resulting application image: Alpine 3.24.2, UID 1000,
read-only root, zero effective capabilities and `NoNewPrivs: 1`. Evidence:
`p04-verifier-supported-base--20260930T214041Z-608798.log` and
`p04-verifier-base-identity--20260930T214103Z-612317.log`.

## Remaining release gates

Fresh-workspace production rehearsal (`fresh-stack-rehearsal-fixed--20260930T210001Z-427142.log`) passed in 14.289 seconds with new secrets and PostgreSQL volume, certificate/hostname verification, guest/room creation, private port checks, and nonroot/read-only/capability-dropped backend. Its containers/networks were removed; project `cgms-rehearsal-2e4b1c6c35` retains its own volume. This uses the existing Docker host.

R01 still needs a true fresh-host rehearsal, complete image/base/dependency
security and license inventory, release provenance and operator acceptance.
R02 now has a bounded mixed-workload capacity and pool-saturation baseline;
long-duration natural gameplay, heavy settlement and causal telemetry-overhead
estimation remain outside these samples. R03 covers grouped service-process
restart on retained storage, but not a real host reboot or destructive storage loss. R04 awaits accepted
recovery limits and off-host/secret recovery. R05 includes the archived-reader rejection evidence for the new economy schema
fence; older binaries cannot run economy-enabled matches.
Native/store acceptance and release approval remain the roadmap's later gates.
