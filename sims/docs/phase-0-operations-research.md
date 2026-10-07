# Phase 0 operations and Go research

Preparation evidence, accessed **2026-09-30**. No runtime, dependency installation,
container or executable simulator is delivered here. Implementation gates remain
S01–S10 in [ROADMAP](../../docs/planning/ROADMAP.md); the complete
[CLI contract](../README.md) remains authoritative. Gameplay, including Q10 and
accepted amendments, comes from [GAME_RULES](../../docs/project/GAME_RULES.md) and
the [accepted register](../../docs/design/RULES-IMPLEMENTATION-QUESTIONS.md).
“Decision” below means an implementation design choice, not a new game rule.

## Verified sources and decisions

| Evidence (all accessed 2026-09-30) | Supported finding | CGMS decision, alternative and limitation |
|---|---|---|
| [Go release history](https://go.dev/doc/devel/release), [official downloads](https://go.dev/dl/) | Official pages list Go 1.27.1, released 2026-09-01, and Go 1.26.8. A major line remains supported until two newer major lines exist. Downloads publish platform-specific SHA-256 values. | S01 target is Go **1.27.1**, subject to a fresh security/support check when implementation starts. Pin the exact archive checksum/platform in build provenance; do not claim an installed local version from this web check. Supported 1.26.8 is an alternative only if a documented compatibility constraint warrants it. Do not create go.mod during preparation. |
| [Go toolchains](https://go.dev/doc/toolchain) | Automatic toolchain selection can download another toolchain; `GOTOOLCHAIN=local` selects the bundled toolchain. A module's language/toolchain declarations alone do not guarantee the exact compiler used. | Provision and verify the selected compiler before offline execution; use process-local `GOTOOLCHAIN=local` and check the actual version. Record language minimum separately from exact build compiler. Do not use global `go env -w`. Explicit download during provisioning is an alternative, not offline operation. |
| [Go modules reference](https://go.dev/ref/mod) | Module versions/checksums and module cache/vendor facilities support dependency verification; module resolution may access proxies/checksum services. | Prefer standard-library engine, bots and CLI initially. No unverified third-party version is selected here. If schema validation requires a dependency, review its primary documentation, license, supported dialect and pinned release at S03; record module graph and go.sum hash. Prepopulate verified dependencies or vendor during provisioning. Prove an offline build with network disabled; cached success is not proof of a clean offline provisioning path. |
| [math/big](https://pkg.go.dev/math/big) | Rat implements arbitrary-precision rational arithmetic; operations mutate their receiver. Shallow Rat copies are unsupported; Rat.Set makes a copy. Zero denominators panic in fraction construction. | Encapsulate rationals, deep-copy at state boundaries and never return writable numerator/denominator internals. Validate denominator before construction; serialize reduced signed decimal numerator and positive decimal denominator, zero as 0/1. No float conversion in accounting. Fixed-width scaled integers are unsuitable for arbitrary denominators; decimal rounding changes Q10. |
| [Pipelines and cancellation](https://go.dev/blog/pipelines) | Go's primary example bounds concurrent work and propagates cancellation so blocked senders can stop; unconstrained parallel work consumes unbounded resources. | One bounded worker owns an entire match, including all games and financial continuation. Feed stable match IDs through a bounded queue; collect by index rather than completion order. Channel/context mechanics are adapter concerns, outside the pure engine. A goroutine per game would violate ownership and can lose carried debts. |
| [os.Rename](https://pkg.go.dev/os#Rename) | Rename can replace an existing file; even same-directory rename is not atomic on non-Unix platforms. | Publish immutable artifacts via temporary sibling files and manifest-last commit on tested local Unix filesystems. Reserve a new output directory exclusively; never overwrite an earlier run. Portable atomicity requires platform-specific validation or explicit rejection of unsupported durability guarantees. Rename alone does not prove power-loss durability. |
| [encoding/json](https://pkg.go.dev/encoding/json) | The legacy decoder can ignore unknown fields and accepts duplicate members; later duplicate values can replace or merge earlier ones. DisallowUnknownFields only addresses part of strict validation. | S03/S09 must reject duplicate keys, unknown keys, wrong case, trailing values, malformed UTF-8 and noncanonical numeric encodings before normalization. Apply structural schema plus semantic cross-field validation. Plain struct decode is insufficient; changing JSON libraries does not remove the need for negative fixtures. |
| [Go fuzzing](https://go.dev/doc/security/fuzz/) | Coverage-guided fuzzing minimizes failures and persists reproducing inputs. Targets should be deterministic, fast and independent of shared/global state. | Fuzz isolated transition sequences, malformed config/replay records and small ledgers. Constrain generator size, use per-case streams, preserve the minimized reproducer in curated testdata after privacy review. A fuzz time budget is engineering coverage, not a proof of all rules or settlement complexity. |
| [Race detector](https://go.dev/doc/articles/race_detector) | Detection is dynamic and covers executed paths; supported platforms and C compiler requirements constrain execution. | Run relevant worker, cancellation, publication and shared-bot configuration tests under the detector in S09. Record platform/compiler and exact checks. Race-free output still needs worker-count deterministic digest tests. Do not require a C compiler in the final offline simulator container merely because the test environment needs one. |
| [Go diagnostics](https://go.dev/doc/diagnostics) | CPU, heap, blocking and execution-trace tools diagnose different resource/latency problems and can perturb measurements. | Profile fixed seeded workloads separately from headline throughput. Record build/hardware/workers, workload hashes, warm-up, repetitions, allocations and profile settings. Debug first divergence with a minimal replay and logical IDs before inspecting scheduler traces. No performance claim until measured. |
| [Reproducible Go toolchains](https://go.dev/blog/rebuild), [go build implementation/help](https://go.dev/src/cmd/go/internal/work/build.go) | Build reproducibility requires eliminating relevant environmental variation; Go build flags control path and VCS metadata. Reproducibility of the compiler is not automatically reproducibility of every application. | S10 records exact source and dirty patch hash, dependencies, compiler checksum, GOOS/GOARCH, CGO, flags and binary hash. Prefer CGO disabled for the runtime if dependencies permit, path trimming, deterministic metadata and no current-time injection. Compare two clean builds from different paths; compare host/container state digests separately. |
| [Docker none network](https://docs.docker.com/engine/network/drivers/none/), [container run reference](https://docs.docker.com/reference/cli/docker/container/run) | Network-none leaves loopback only; read-only root, explicit mounts and resource controls constrain a container's operating environment. | S10 packages a prebuilt pinned image, no application services. Test with network disabled, input read-only, output-only write mount, non-root identity and bounded CPU/memory/processes. Pin the image digest and record runtime version. Image acquisition/build provisioning can need network; executing an already provisioned run must not. No Dockerfile is created in preparation. |

## Exact settlement and immutable-state design

The [research decision and proof obligations](phase-0-research.md#exact-settlement-decision-and-proof-obligation)
are the single detailed preparation procedure for Q10 termination, unbatched
reference, exact batching, compressed audit and resumable settlement. Use that
procedure with S07's fixture/evidence mapping rather than maintaining a second
algorithm specification here. The math/big source evidence above explains the
implementation's mutable-value hazard; it does not prove a CGMS optimization.

For immutability tests, save a deep snapshot, apply an accepted and a rejected
action, mutate returned observation/test copies, then prove all retained states
and original rationals unchanged. Exercise shared slice backing arrays, map
values, nested effect/decision records and Rat/Int internals. Copy-on-write is
an alternative only after equivalent alias-isolation evidence; initial deep
copies are simpler to review for this small 104-card model.

## Worker ownership, cancellation and publication

Create the complete logical match/seat/seed schedule before dispatch. Each match
owns its game and bot streams; neither stream derives from worker ID, elapsed
time or completion order. Workers return immutable indexed outcomes; the
coordinator emits stable ordering. Compare worker counts 1, 2 and a larger
bounded value for commands, outcome hashes and replay digests. Timing/resource
metadata may vary and must be outside semantic digests.

Deterministic operation budgets govern bot choices and simulation progress.
Wall-clock watchdogs and memory ceilings protect the host but may stop different
amounts of work on different machines: record that limitation and incomplete
denominators. Cancellation stops dispatch, lets atomic work reach a checkpoint,
joins owned workers and persists finished games plus exact pending progress.
No disconnected actor is auto-passed and no finish/refusal is fabricated.
Test cancellation during queue backpressure, bot selection, settlement segments,
trace writing and manifest publication; verify no goroutine/channel deadlock.

Publication protocol for the future S09 adapter: exclusively create the requested
new run directory; write bounded temporary files there; close/check writes,
hash exact bytes and synchronize according to the supported filesystem contract;
publish finalized files; write, verify and atomically publish the manifest last.
Where crash durability is claimed, synchronize the containing directory as well
and fault-test the actual filesystem. A missing final manifest means unfinished
publication; recovery must not promote it silently to complete. Preserve valid
completed artifacts on cancellation; explicit partial manifests retain diagnostics.
I/O failure may prevent even a partial manifest, which must be reported as such.
Follow the existing CLI exit codes and no-overwrite contract exactly.

## Validation and evidence obligations

| Layer | Implementation evidence required |
|---|---|
| Unit/property | Red-before-green S01 behavior test; card conservation, alias isolation, exact arithmetic, decision continuation and visibility invariants for supported transitions. |
| Golden | Strict parser rejection vectors; canonical config bytes/hashes; normalized rationals; committed random vectors; observation and state/event digests with declared schema versions. |
| Differential | Reference versus optimized settlement including audit expansion; prefix checkpoint/resume equivalence; legal generator versus independent adjudicator. |
| CLI | Every required flag/precedence and exit code; unknown versions; existing output rejection; corrupted manifest/hash; first divergence; partial reports and explicit incomplete denominator. |
| Concurrency | Whole-match ownership, worker-count invariance, bounded queue saturation, cancellation and relevant race tests. |
| Offline/build | Provisioned host/container isolation; no service/network dependency; repeat clean binary build hashes; semantic host/container replay equivalence. |
| Performance | Fixed workload hashes, player populations, work counts, denominator size, settlement reference versus compressed work, resume cost, allocations, peak memory and p50/p95 decisions; no arbitrary release limit before baseline. |

Generated profiles, debug dumps, checkpoints, traces and full replay manifests
can expose hidden cards, future random state and deck order. Keep them restricted
under ignored `sims/artifacts/`; share reports only through named observer
projection and test that export removes private state/seeds that would reconstruct
it. Public summaries may identify hashes without publishing private inputs.
Retention follows the CLI contract: proposed 30-day local retention, protected
review-cited evidence until explicitly released, and preview-only selection before
future pruning. No pruning tool or executable simulator command exists here.

These are design obligations, not runtime evidence. Preparation can validate
document links, schema syntax and example structure; it cannot certify Go tests,
alias safety, optimization equivalence, shutdown behavior, platform atomicity,
binary reproducibility or offline container operation. Broad conformance targets
remain recorded even when Phase 0 supports only selected scenarios; unsupported
accepted transitions stop explicitly and remain implementation coverage gaps.
