# Adaptive entry and startup recovery — 2026-10-02

Run `adaptive-startup-20261002`. Existing staged card-artwork changes are
preserved. The evidence below separates application repairs from the unresolved
push diagnosis. All three application repair and local verification scopes passed;
no commit or push was attempted.

## Reproduced causes

- The normal `/` entry redirects to `#/play` with Auto but renders the legacy
  combined action board, bypassing the online adaptive composition. Live
  390/1440/599/600/1049/1050/390 widths retain that board. The local shell also
  used obsolete 700/1100 classification thresholds. Before-fix Chromium
  153.0.8010.12 became usable in 739 ms with no console/network errors once
  the stopped proxy was started.
- Served `index.html`, `flutter_bootstrap.js`, and `main.dart.js` matched the
  local build byte-for-byte. Before-fix main JavaScript SHA-256:
  `1ee104ac502fef5d001c05b419af1dd8c0025d23834c2c0df34785e622a037f7`.
  This reproduction is not attributable to a stale browser cache.
- The initial proxy was stopped (8080 refused connections). Documented
  `services.up` brought it up successfully. Separately, with the backend stopped,
  a fresh online page entered a connection error at 2,552 ms; the backend became
  ready at 3,347 ms, but another five seconds left that page stuck. Session
  restoration performed one read and had no
  automatic readiness recovery. Normal initialization latency is distinct from
  this demonstrated persistent error.
- Historical Firefox keyboard traversal had competing browser and Flutter Tab
  handling: Flutter's asynchronous cancellation could arrive after a native
  focus move. A pre-bootstrap helper now cancels only a Tab event handled by
  Flutter's normal action, within its current dispatch. Cleanup runs at the
  next task because Chromium can drain microtasks between listeners. Browser
  shortcuts, external controls and unhandled view exit are preserved.
- Geometry-based traversal across independently scrolling workspaces could skip
  distant cards in short, enlarged-text views. A red/green normal-route widget
  test demonstrates this at 844×390 and 200% text. Each retained workspace now
  has stable widget-order traversal. The unchanged 48-case Firefox sequence
  passes, including the previously failing final RTL landscape case.
- WebKit exposed a second, distinct keyboard defect: a one-control modal wraps
  back to the same control, which Flutter reports as an unsuccessful traversal.
  Native reverse Tab then escaped the inspector. The normal focus action now
  consumes unchanged focus only inside a closed-loop scope. Four action tests
  cover forward/reverse modal wrap and preserve forward/reverse browser exit.

## Evidence policy

Disposable captures and raw browser reports live under ignored
`client/.browser-tools/artifacts/` or the existing ignored evidence directories.
Curated lightweight results are recorded here. Required mockups, artwork,
fixtures, current widget renders and retained historical failures are preserved.
No push, commit or history rewrite is authorized or attempted.

## Implemented repairs

The approved local flow now uses `CgmsAdaptiveTable`, the same composition as
the online route. Available Flutter logical width selects phone below 600,
tablet through 1049, and desktop from 1050. Phone has separate public, hand and
decision workspaces with bottom navigation; tablet keeps public context beside
a rail-selected companion; desktop keeps all three workspaces visible. Physical
card artwork remains complete and nonoverlapping, with compact seat grouping,
independent scrolling and text-aware controls. The local shell's former
700/1100 thresholds are corrected. Inspectors receive keyboard focus so Escape
can dismiss them. The 200% Auto toolbar is sized from the enlarged text.

Retained pane identity preserves host-owned state, form controllers and editing
selection. Browser assertions cover a focused raw Offer Id draft across both
boundaries, selected physical cards, and an uncertain command's original
envelope. The pre-existing submit protocol makes at most two identical initial
attempts; resizing and receipt recovery must add zero command POSTs.

Session restoration coalesces callers and performs bounded read-only recovery
with visible attempts and explicit terminal failure. Match stream recovery keeps
the last confirmed projection and uses an eight-attempt outage budget, reset
only by a validated event or five seconds of an open socket. Generation and
disposal checks prevent stale callbacks. Readiness recovery never replays a
mutation. Malformed session envelopes are validated before changing actor or
pending-operation state; HTTP 401/403 are terminal independently of body codes.

The [screenshot retention report](2026-10-02-screenshot-retention.md) and
[per-file inventory](2026-10-02-screenshot-inventory.csv) record all dispositions
and historical commit evidence. Cleanup removed 262 superseded captures
(31,627,881 bytes), preserving 1,544 screenshot hashes and 1,907 protected
evidence/design/art files. No reachable blob is at least 50 MiB. Removed committed
screenshots remain in `0775248`, `6248bb5` and `90eef62`; the available evidence
does not establish the reported push obstruction or a need for history rewriting.

## Verified application behavior

The release web build uses the project-local Flutter 3.47.5 SDK and same-origin
assets (`flutter build web --no-web-resources-cdn`). Every browser run compares
the actual served index, keyboard bridge, bootstrap and main bundle with the
current build. The final main bundle SHA-256 is
`51a4a4c90cdeaa699ffb30ec9812cfa09bbe821b1517cdbbb63c19ccfb7baf8b`.
Complete hashes and compact case summaries are retained in the
[verification record](2026-10-02-adaptive-startup-verification.json).

| Gate | Result |
|---|---|
| Flutter host | 244 tests; analyzer clean |
| Shared UI package | 75 tests; analyzer clean |
| Browser support | 24 diagnostic, bootstrap, bridge and runner tests pass; runner syntax clean |
| Backend | `go test ./...` and `go vet ./...` pass |
| Local operations | 8 service-operation Python tests pass |
| Artwork and boundaries | All 208 artwork hashes/native dimensions, local fonts and import boundaries pass |
| Normal `/` entry, Auto | Chromium 153.0.8010.12, Firefox 155.0 and WebKit 26.6: 52 cases each |
| Full online journey | Chromium: 67 cases, zero unexpected browser errors |
| Online adaptive/RTL/accessibility matrices | Firefox and WebKit: 33 cases each, zero unexpected browser errors |
| Historical forward-only keyboard sequence | Firefox: 48 cases, zero commands submitted |
| Cold startup and fault recovery | 8 browser cases, zero unexpected diagnostics |

The entry checks start from fresh phone and desktop contexts, preserve physical
card selection through live 599↔600 and 1049↔1050 transitions, and exercise
360/390-wide phones, 768/1024 tablets, 1280/1440 desktops, rotation and 844×390
short landscape. Real pointer/touch and keyboard actions reach the hand and
inspector, including a distant offscreen card at 200% text and reverse Tab in
its modal. Reviewed final captures show separate phone workspaces, tablet rail
and two panes, and persistent desktop public/hand/decision panes.

Online checks also cover RTL with 200% text, measured 48-logical-pixel primary
touch targets, browser back/forward, private projection/outsider boundaries,
response windows, guest room/deal/reload, settlement, focused Offer Id text and
caret across six layout changes, and pending-command identity. Six resizes and
receipt recovery add zero command POSTs beyond the existing two-attempt initial
submission contract. Lost replies and signed-out 401 responses are deliberate
fault cases, reported separately from unexpected errors. No assertion was
removed or weakened to turn a failing gate green.

## Cold launch and readiness timings

Each independent launch first stops the project's services and verifies none
remain running, then executes the documented `make services.up MODE=local`.
PostgreSQL's existing volume and data are retained. Each launch opens a fresh
Chromium context, navigates the normal root, and creates a usable online guest
room without manual refresh or restart. This was local loopback functional QA
with concurrent browser/test activity, not a performance benchmark or a clean
machine installation.

| Launch | Health ready | Wrapped startup command | Normal root usable | Online usable from root navigation |
|---|---:|---:|---:|---:|
| 1 | 11,335 ms | 16,489 ms | 939 ms | 1,607 ms |
| 2 | 10,593 ms | 15,856 ms | 770 ms | 1,209 ms |
| 3 | 9,968 ms | 15,053 ms | 724 ms | 1,201 ms |

Health readiness is measured independently of command-wrapper completion;
application timings begin after service startup, at browser root navigation.
The startup runner uses process exit rather than inherited output-pipe closure
for its command timer. No fixed sleep is used as a readiness gate.

| Deliberate outage/fault | Observed outcome |
|---|---|
| Fresh session, backend initially stopped | Error observed 2,356 ms; backend ready 3,374 ms; automatically usable 3,728 ms, two reads |
| Reload during backend outage | Error observed 2,196 ms; ready 2,925 ms; automatically usable 3,058 ms, three outage reads plus one pre-outage read |
| Established match stream interrupted | Backend ready 1,228 ms; automatically usable 1,511 ms; two real HTTP 101 socket handshakes across the outage, same persisted game, zero command POSTs |
| HTML 503 proxy response | Two reads, automatically recovered |
| Malformed successful session response | One read, explicit terminal error; no invalid session accepted |

First-run service logs, console diagnostics, failed requests and timings were
captured. The final startup run has zero uncaught page errors and zero unexpected
diagnostics. Expected observations were three anonymous-session 401s, one
intentional 502, one intentional 503 and two aborted reads during outages.
All four ordinary services ended healthy. The retained PostgreSQL volume's
identity and creation timestamp are unchanged. Service logs contain the existing
Redis `vm.overcommit_memory` warning; it did not prevent readiness. No workstation
sysctl or other system configuration was changed.

The reproduced defect was client readiness/reconnect behavior. Stopped-service
startup ordering itself passed, so there is no unsupported claim of a repaired
Go service sequencing bug. Read-only recovery has a finite budget: a dependency
that remains unavailable beyond it produces a clear terminal error and explicit
retry control, rather than an infinite loop.

## Review, recovery and limits

Independent review covered the adaptive/focus changes, session validation,
reconnect generation/budget rules, transport error handling, privacy and command
identity. Red/green regressions include the ordinary-entry layout, offscreen
keyboard traversal, singleton modal loop, transient readiness, malformed session
envelopes and authentication rejection. Diagnostic harness cleanup is reviewed
separately so a failed diagnostic cannot prevent backend restoration. Its nine
focused tests pass. The successful eight-case final browser run predates only
that runner cleanup patch; application code, assertions and built assets are
unchanged. The verification record preserves this provenance.

The previously unresolved Firefox evidence is retained as history and is closed
by the same 48-case forward traversal sequence on this build. Its earlier
43/48 and diagnostic failures are not reclassified as passes. A temporary QA
fixture reached its configured one-hour lifetime after completed runs; its
captured `TimeoutExpired` was read before starting a fresh isolated fixture.
This is separate from ordinary application startup. The final isolated fixture
was intentionally terminated after acceptance; its interruption and schema-cleanup
warning were inspected, and port 8081 plus its disposable container, volume and
network were confirmed removed. Ordinary persistent services were unaffected.
Recovery records preserve those diagnoses and the final passing evidence.

All generated captures from this task remain ignored. Required design/game
rules/card-art authorities and preserved evidence hashes were checked again.
The complete prospective staging set includes existing work from the prior
408-file staging window; it was reviewed for unrelated additions and secrets.
See the verification record for exact log IDs and final staging evidence.

These results cover local Flutter web in three real browser engines. They do
not certify native devices, managed/public deployment, independent networks,
full assistive-technology behavior or release readiness. Existing roadmap
release gates remain open. The push obstruction remains unconfirmed: the remote
HEAD is reachable, no ≥50 MiB historical blob was found, and no push was attempted.
A human must retain the actual rejection from any future `make git` failure
before a further diagnosis; agents must not rewrite published history.
