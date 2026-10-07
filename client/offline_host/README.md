# P02 local engine and durable host journeys

This is the Linux/macOS **host harness** for the provisionally selected shared
Go engine. It is separate from the P01 research probe: production local sessions
use the full lifecycle and chance adapter, own durable storage, and expose only
a seat projection to presentation code. The executable opens no sockets and
requires no account, server, online allowance or entitlement. Offline outcomes
never earn dirt or become online results. `host.dart` imports `dart:io`; it
cannot provide an offline web route. Flutter web still requires online authority.

Android/iOS FFI packaging, actual airplane-mode/device save lifecycle acceptance
and final engine approval remain R06N. This executable and its desktop process
bridge are not mobile packaging or an offline Flutter screen.

## Reproduce with existing tools

From the repository root, build the host and run its tests:

```sh
xops/agent/safe-run.sh offline-tests -- bash -c 'cd backend && go test ./internal/offline ./cmd/offlinehost -count=1 && go test ./internal/matchstore -run TestLocalCapability -count=1 && go vet ./internal/offline ./cmd/offlinehost && go build -o ../sims/artifacts/offline-host ./cmd/offlinehost'
xops/agent/safe-run.sh offline-dart-analysis -- dart analyze client/offline_host
```

Run the Dart process/restart tutorial with a **fresh private directory**. This
command computes the canonical rules fingerprint and supplies argv directly:

```sh
xops/agent/safe-run.sh offline-dart-journey -- python3 - <<'PY'
import hashlib, subprocess
from pathlib import Path
root = Path.cwd()
rules = hashlib.sha256((root/'docs/project/GAME_RULES.md').read_bytes()).hexdigest()
subprocess.run(['dart', 'run', 'client/offline_host/host_test.dart',
               str(root/'sims/artifacts/offline-host'),
               str(root/'sims/artifacts/my-fresh-offline-journey'), rules], check=True)
PY
```

Only the launcher supplies a save directory and rules fingerprint. Requests
cannot supply paths, private saves, random outcomes, account identities or bot
seats. The host locks its directory for its lifetime; a second owner rejects.
The OS releases that lock after process death. Windows storage currently fails
closed until an equivalent durable lock implementation is verified.

## Runtime contract

Each input is one strict JSON line, maximum 64 KiB; unknown and duplicate keys
reject. Output is either `{schema: "cgms-offline-view-v1", view: ...}` or a small
error code. One bounded scheduler or bot transition runs per `step`; waiting for
a human leaves the state unchanged. No timer substitutes a human choice.

| Request kind | Inputs and behavior |
|---|---|
| `new` | Fresh `slot`, `population` 3/4, `games` 1–3, `difficulty`; creates a human at seat 1. Existing/unreadable slots reject. |
| `tutorial` | Same inputs; deals and initializes a first-turn lesson, assigning the current active seat to the learner. |
| `load` | Existing `slot`; returns only that session's human-seat projection. |
| `act` | `slot`, expected `version`, `action` (`command` or `operation`); actor must equal the saved human seat. |
| `step` | `slot`, expected `version`; advances automatic work or one bot decision, never the human. |

`View` contains version, difficulty, authorized board/financial projection,
legal action menu, tutorial step and plain-language instruction. Finite menus
sample physical combinations and are not an exhaustive enumeration of every
legal combination. The actor adapter accepts the full authorized command
families; arbitrary valid human actions need not be restricted to menu samples.
Budget exhaustion is an error, never a silent pass. Bots receive only their own
authorized board/party financial observations. The engine may use private state
to validate an action and generate chance; policy selection cannot inspect it.

The local adapter temporarily reuses `matchstore`'s **pure** envelope, scheduler,
actor validation and projection functions through two explicitly local wrappers.
It never constructs a Store, database connection, transport route or economy
service. Offline saves reject the online economy envelope schema. This import
is shared implementation, not runtime network or online authorization.

## Saves and failure semantics

Saves are versioned private JSON files, limited to 4 MiB. Admission validates the
save schema, session/engine versions, canonical rules fingerprint, checksum and
engine state. The SHA-256 checksum detects corruption; it is not authentication
against the workstation owner. Full saves contain hidden cards and private
chance and must remain in app-owned storage, never a response, log or URL.

A save writes a fresh mode-0600 temporary file, syncs it, atomically renames it,
then syncs the directory. Invalid data or failures before rename preserve the
previous save. A directory-sync failure **after** rename has an explicitly
uncertain durability outcome (`save_outcome_unknown_reload`); reload the slot
before deciding what to retry. Every host request reloads storage and every
mutation saves before publishing its projection. Version checks prevent a stale
view from replaying a newly identified action. No automatic save migration or
backup recovery is claimed; incompatible or damaged saves fail closed.

Slot names are restricted to lowercase letters/digits/hyphens. `os.Root`
prevents symlink/path escape; nonregular/oversized files reject, and nonblocking
opens avoid hanging on a substituted FIFO. A kernel lock prevents two host
processes from overwriting each other's state. Abrupt process death can leave an
unreferenced temporary file; it cannot make the loader treat it as a save slot.

## Difficulty and tutorial scope

- `beginner`: the existing economic heuristic builds support, uses legal defenses,
  takes a proven win, avoids takeback cycles and explicitly finishes turns.
- `standard`: the pressure heuristic values attacks more and handles selected
  financial commitments differently.
- `advanced`: the opportunity heuristic retains early royals for formations and
  uses additional acquisition/financial priorities.

All tiers use identical rules and hidden-information restrictions; none is an
intentionally illegal or passive policy. Tests prove immediate-win behavior in
all tiers, productive beginner openings, different royal-retention choices,
no idle financial churn, and complete beginner games for both populations.
These are transparent heuristic tiers, **not experimentally calibrated ordinal
strength levels**; comparative bot-strength claims and confirmation campaigns
remain P03. The bounded host tests are verification, not an independent bot study.

The first-turn tutorial uses a real deal and the same legal engine: select a
series opening, observe ordered explicit opponent passes, then end the turn.
Guided passes apply only while teaching that response window. Instructions and
progress persist through each action. Tests reopen after every step; the Dart
journey additionally restarts the executable and kills/reopens a process after
a saved completed lesson. This is one concrete beginner journey, not a complete
curriculum or proof of mobile lifecycle acceptance.
