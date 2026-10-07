# P06 local product contracts — 2026-10-01

> Archive removal update (2026-10-02): the owner requested removal of every
> file and subdirectory in `client/screenshots/` except `cgmsart-phone.png`,
> including raw reports, manifests, logs and copied harnesses. This document
> remains a historical account of the recorded results; its capture-time counts
> do not imply that those raw files remain available. Removed archive links are
> plain text below. See the [retention report](2026-10-02-screenshot-retention.md)
> and its per-file inventory for the disposition.

Run ID: `p06-local-20261001`. Continues the current roadmap's local product gate;
P03 bot development and every Phase 5 production/native-release gate are excluded.
The preceding `styles-products-20261001` staging set is preserved.

## Scope and decisions

The owner selected fixed-term standalone Ad-free. Its exact duration and paid
Daily benefits are still pending; no product term, price, renewal rule, live
provider verifier or checkout flow is invented. P06 stays unchecked until its
complete applicable local gate is met.

The internal Ad-free contract accepts a verified absolute expiry, removes ads
without granting unlimited starts, and preserves independent earned access.
The private account response now exposes effective `ad_free_until`. A separate
checksum-pinned migration adds the provider kind without editing either earlier
SQL migration. Historical receipt recovery and previously promised benefits
remain intact.

The [updated applicability matrix](../design/DESIGN-economy-and-stores.md#platform-applicability-matrix)
records official Apple/Google research before SDK selection. In particular,
one-day paid access is not an Apple daily auto-renewing subscription, and a
hypothetical one-day Google prepaid model has a calculated 12-hour acknowledgement
deadline. Neither mapping is selected here. Account deletion, privacy disclosures
and real-store restore remain explicit native/release obligations.

## Verification

The final Flutter source passed 197 client tests, 58 shared-package tests,
analysis for both packages, and a release web build with local web resources.
The existing Cupertino icon-family build warning remains; it is not native font
acceptance evidence. No application source changed after this build.

| Check | Result | Captured log under `/tmp/agent-runs/` |
|---|---|---|
| Client tests, analysis, release web build | PASS; 197 tests | `p06-client-final--20261001T122553Z-2842421.log` |
| Shared package tests and analysis | PASS; 58 tests | `p06-shared-package--20261001T122651Z-2844722.log` |
| Economy PostgreSQL race suite | PASS; 19 tests plus two provider subtests | `p06-ad-free-pg-green--20261001T121900Z-2821520.log` |
| Ad-free owner consumes all three free starts | PASS; strengthened regression | `p06-owner-ad-free-cap-regression--20261001T122344Z-2836375.log` |
| Private HTTP expiry, earned redemption, revocation | PASS; real PostgreSQL race test | `p06-account-http--20261001T122011Z-2824990.log` |
| QA fixture unit race suite | PASS | `p06-ad-free-fixture-unit-green--20261001T122114Z-2828447.log` |
| QA fixture PostgreSQL race suite | PASS; 15 scenarios × three/four players × economy on/off; 14 Ad-free viewer combinations | `p06-fixture-pg--20261001T122234Z-2832077.log` |
| Economy/transport and fixture vet | PASS | `p06-ad-free-vet--20261001T121940Z-2823692.log`; `p06-fixture-vet--20261001T122402Z-2837625.log` |
| OpenAPI security/receipt/envelope contract tests | PASS; two tests | `p06-openapi-contracts--20261001T122312Z-2834719.log` |
| Policy source and local-link checks | PASS | `p06-policy-final-check--20261001T122402Z-2837592.log` |

The new tests were exercised before implementation: the missing Ad-free kind,
private expiry, fixture scenario and malformed-account validation failed before
their respective changes. Coverage includes restore, eight concurrent duplicate
provider notifications, account/environment binding, unknown/pending outcomes,
late events, exact expiry, revocation with independently valid access, historical
receipt recovery and migration-checksum rejection. Malformed successful account
responses retain the last confirmed view, disable new spending and allow receipt
recovery; valid refresh restores normal operation.

### Rendered account journeys

All **36 browser cases passed**, with **90 final screenshots** and no unexpected
page/console errors. Engines ran serially against the same real local Go
HTTP/WebSocket authority and disposable PostgreSQL. The
final manifest pins
all reports, screenshots, the served release assets and the exact final harness.

| Engine | Version | Cases / screenshots | Captured log under `/tmp/agent-runs/` |
|---|---|---|---|
| Chromium | 153.0.8010.12 | 12 / 30 | `p06-browser-chromium-verified--20261001T123609Z-2866591.log` |
| Firefox | 155.0 | 12 / 30 | `p06-browser-firefox-final--20261001T123759Z-2870473.log` |
| Containerized WebKit | 26.6 | 12 / 30 | `p06-browser-webkit-final--20261001T123929Z-2874215.log` |

The final `main.dart.js` SHA-256 is
`73147b2607cf96d531e9d26c98b93e0eb0f90baf276f3445299e671bf5cdb549`;
the account harness SHA-256 is
`0f2458875868064a3d4d7fc8003c7a00c5383f9f6b73d4a9659efa5ffb90ce2e`.

Each engine covers phone **390×844**, tablet **768×1024** and desktop
**1440×900**, normal presentation and RTL with 200% text. Standalone Ad-free
retains the free-start allowance; an explicit Daily redemption adds unlimited
starts without extending ad-free expiry; reload preserves both. Other seats
retain their own account benefits and the public board excludes entitlement
fields. The HTTP and fixture tests separately cover authenticated account
selection and game-snapshot boundaries. These are specific privacy regressions,
not an exhaustive privacy audit.

The retained journeys also exercise real room/start allowance consumption,
settlement rewards, streamed admission after the roster redeems passes,
insufficient Weekly funds, a lost purchase response, original-receipt recovery
after reload, and a failed post-purchase match refresh. Synthetic prior rewards
and a trusted absolute-expiry grant provide the starting states; subsequent
reads, redemptions, room actions and settlement commands use real authority paths.
One deliberately malformed HTTP 200 account response is rejected; keyboard
Refresh restores the valid view without any mutation request.

Accessibility checks actually run: named Flutter browser semantics; real Tab
reachability and Enter activation; focused controls scrolled into the viewport;
48-pixel minimum purchase/refresh/close targets; explicit navigation with zero
purchase requests; and visual inspection of normal, RTL/200%, invalid and recovered
account states. Exact expiry retention is asserted through the account API, with
exact formatted-label coverage in widget tests. Browser screenshots establish
readable tested states, not native screen-reader or complete WCAG acceptance.

Screenshot metadata and run reports originally lived in three engine directories;
the archive is now removed.
Original, now-removed screenshot filenames included `phone-standalone-ad-free.png`,
`tablet-ad-free-rtl-200-controls.png`, `desktop-combined-benefits-reloaded.png`
and `phone-invalid-account-refresh-recovered.png`. Prior redesign/card-style
before captures
and after captures
were captured as historical evidence; they are not presented as this build's
verification. This account-contract slice did not take a new before capture.

The final fixture process exited zero after graceful shutdown and its disposable
database resources were cleaned by the runner. Logs:
`p06-final-fixture--20261001T122707Z-2845577.log` and
`p06-fixture-cleanup--20261001T124120Z-2879054.log`.

### Recovery and review

An initial OpenAPI command named a nonexistent Python test module; its captured
failure was diagnosed and replaced with the repository's actual Go contract
tests. Two test-fixture const lint findings were repaired before final analysis.
The first Chromium run measured a partially clipped semantics node after direct
DOM focus skipped Flutter's scroll handling. A focused keyboard trace proved
that Tab reveals the full control. The acceptance runner now traverses real Tab
events, requires each control to be focused and inside the viewport, and retains
the 48-pixel assertion. Failed-run reports are preserved locally under
`/tmp/agent-runs/p06-chromium-first-failure/`; they are not passing evidence.
A second browser run exposed the same direct-DOM-focus problem after Refresh
was disabled and rebuilt: the attempted Enter produced no HTTP request. A
request-counted trace demonstrated that real Tab traversal followed by Enter
issued the second GET and cleared the error. Both refresh actions now use that
keyboard path; the failed capture is retained in
`/tmp/agent-runs/p06-chromium-refresh-failure/`. The first isolated trace used a
fixture identifier longer than the documented limit and received 400; shortening
that debug identifier repaired the setup without changing the validator.

Independent reviews covered the backend/migration/API and frontend/fixture
contracts. Documentation review found a stale policy-matrix fragment; a stable
explicit anchor repairs the link. No assertions were removed or relaxed to make
the checks pass. Final harness/screenshot review found no actionable issue.
The 498 pre-existing staged files were compared against the initial index
manifest before staging this slice; their index entries were intact. Protected
game rules, offline host/probe and both earlier SQL migrations were unchanged
from that index. The complete staging inventory was checked for secrets and
unrelated artifacts.

## Limits

Synthetic provider fixtures exercise trusted entitlement accounting; they do not
verify Apple/Google transactions or deliver live ads. Arbitrary fixture expiry
windows are not approved product durations. Browser evidence does not establish
native-device or screen-reader acceptance. No bot evaluation or production work
is included. This slice does not rerun the complete gameplay/capture, orientation,
breakpoint, contrast or reduced-motion matrix; previous evidence remains dated
and separate. Live purchases remain unavailable. The exact Ad-free duration and
paid Daily benefits are the remaining P06 decisions; the roadmap stays 38/46.
