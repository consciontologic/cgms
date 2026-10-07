# Approved paid access — 2026-10-01

> Archive removal update (2026-10-02): the owner requested removal of every
> file and subdirectory in `client/screenshots/` except `cgmsart-phone.png`,
> including raw reports, manifests, logs and copied harnesses. This document
> remains a historical account of the recorded results; its capture-time counts
> do not imply that those raw files remain available. Removed archive links are
> plain text below. See the [retention report](2026-10-02-screenshot-retention.md)
> and its per-file inventory for the disposition.

Run ID: `p06-approved-20261001`. Continues P06 after owner decision 0018 and
preserves the preceding 596-file staging set. P03 bots and all Phase 5 work are
excluded. This report supersedes the two pending product questions in the
[earlier local-contract report](2026-10-01-p06-local-contracts.md), retaining its
original evidence and limitations.

## Approved behavior

| Source | Unlimited online starts | Ad removal | Term |
|---|---|---|---|
| Standalone paid Ad-free | No; separate unlimited access can coexist | Yes | 30 × 24 hours from confirmed purchase |
| Paid Daily | Yes | Yes | 24 hours from confirmed purchase; non-renewing |
| Paid Weekly / Monthly / Yearly | Yes | Yes | Verified provider period expiry; no new calendar/renewal mapping |
| Earned Daily / Weekly | Yes | No; separate ad-free access can coexist | One / seven days, retaining existing stacking |

Earned prices remain Daily 30 dirt and Weekly 210 dirt. Ad-free, Monthly and
Yearly cannot be redeemed with dirt. No cosmetics, sold currency, promotional
products or prices are introduced. The normal allowance remains three free
online starts per UTC day, and paid unlimited access does not consume it.

The local provider contract validates the original confirmation instant and
expiry of new fixed-term purchases. Confirmation metadata remains private and
trusted, never accepted from a client purchase payload. Restore and duplicate
notifications preserve time; revocation removes only the relevant benefit.
Historical receipts remain recoverable. All paid unlimited kinds feed both the
private benefit projection and atomic game-start eligibility.

The account dialog explains the distinct paid and earned benefits through ARB
localization, while eligibility still comes from authoritative account flags.
Existing opaque surfaces, controls, scrolling, focus and device layouts are
retained. Live paid checkout remains unavailable.

## Verification

The final Flutter source passed 210 client tests, 58 shared-package tests,
analysis for both packages, and a release web build with local web resources.
The existing Cupertino icon-family build warning remains; it does not establish
native font acceptance. No application source changed after this build.

| Check | Result | Captured log under `/tmp/agent-runs/` |
|---|---|---|
| Client tests, analysis, release web build | PASS; 210 tests | `approved-client-final--20261001T133227Z-2972701.log` |
| Shared-package tests and analysis | PASS; 58 tests | `approved-shared-package--20261001T133402Z-2977091.log` |
| Economy PostgreSQL race suite | PASS; 24 tests plus 10 subtests | `p06-terms-pg-green--20261001T133711Z-2984604.log` |
| Authenticated economy HTTP race suite | PASS; six tests plus six subtests | `approved-account-http--20261001T133856Z-2988904.log` |
| QA fixture unit race suite | PASS; nine tests | `paid-fixture-unit-green--20261001T133513Z-2980099.log` |
| QA fixture PostgreSQL race suite | PASS; 10 tests plus two subtests; 76 initial groups and 70 paid-viewer combinations | `paid-fixture-pg-final--20261001T133919Z-2990654.log` |
| Economy, transport and fixture vet | PASS | `p06-terms-vet--20261001T133742Z-2986410.log`; `approved-transport-vet--20261001T134125Z-2996585.log`; `paid-fixture-vet--20261001T133514Z-2980488.log` |
| OpenAPI security/receipt/envelope contracts | PASS; two tests | `approved-openapi--20261001T133856Z-2988920.log` |

Regression coverage includes exact confirmation/expiry boundaries, rejection of
missing or wrong fixed terms, future confirmation, duplicate/restore stability,
unknown outcomes, revocation, late events, environment/account binding and
historical receipt preservation. Each of the four paid tiers admits five starts
without charging the free allowance. Earned access retains ads after paid access
is revoked. HTTP tests reject cross-account projection and exclude private
provider metadata. The additive migration leaves the three earlier SQL files
unchanged, preserves stored receipt bytes and rejects old writers/checksum drift.

Thirteen added widget cases cover all four benefit combinations and the real
account dialog at three viewport sizes, both directions and normal/200% text.
Every terms paragraph scrolls into view; controls retain 48-pixel targets, and
navigation does not send a purchase. These tests verify the supplied English
catalog in RTL layout, not an unprovided translation.

### Rendered account journeys

All **72 browser cases passed**, with **162 screenshots** and no unexpected
page/console errors. Engines ran serially against the final release build and
the real local Go HTTP/WebSocket authority with disposable PostgreSQL. The
final manifest
pins the source, served assets, harness, reports, screenshots and successful logs.

| Engine | Version | Cases / screenshots | Captured log under `/tmp/agent-runs/` |
|---|---|---|---|
| Chromium | 153.0.8010.12 | 24 / 54 | `approved-browser-chromium--20261001T134138Z-2997812.log` |
| Firefox | 155.0 | 24 / 54 | `approved-browser-firefox--20261001T134411Z-3004532.log` |
| Containerized WebKit | 26.6 | 24 / 54 | `approved-browser-webkit--20261001T134638Z-3011834.log` |

The final `main.dart.js` SHA-256 is
`e9fad9c954d44776cc728c89b754aac306601c30cbd2eef7dbeb2f6993d37fd3`;
the account harness SHA-256 is
`116cbb0420b2a77ff3f248b1bca92a9229fe3c96e08bb7ef75e272dc0a1ac59f`.

Each engine exercises phone **390×844**, tablet **768×1024** and desktop
**1440×900**. All four paid tiers show unlimited/ad-free access and preserve the
three free starts after fixture admission and reload. Another seat retains its
own benefits; entitlement/provider fields remain absent from match responses.
Reading terms, navigating and reloading make zero mutation requests. These are
specific privacy regressions, not an exhaustive privacy audit.

The retained journeys cover standalone Ad-free with the ordinary allowance,
combined earned/ad-free access, earned-only access retaining ads, insufficient
Weekly funds, actual room/start allowance consumption, settlement rewards and
streamed admission after the roster redeems passes. Lost purchase replies,
original-receipt recovery after reload, failed follow-up match refresh and a
deliberately malformed account response remain recoverable through real API
paths. Synthetic prior rewards and verified-provider fixtures supply the initial
state; they do not establish live provider integration.

Accessibility checks actually run: named Flutter browser semantics; real Tab
reachability and Enter activation; focused-control viewport assertions;
48-pixel purchase/refresh/close targets; normal and RTL/200% account layouts;
and visual inspection of paid, pending, recovered and malformed account states.
Independent review inspected 16 Chromium and 12 Firefox screenshots, with parent
review also covering all three WebKit device classes. Firefox's pending-reload
capture records the top scroll position; adjacent captures show focused recovery
and its confirmed receipt. The real keyboard/receipt checks passed. This capture
alone does not demonstrate the recovery button's visual focus after reload.

Historical capture reports originally lived in three engine directories;
the archive is now removed.
Original, now-removed image examples: `phone-paid-weekly-terms.png`, `tablet-paid-monthly-active.png`,
`desktop-paid-yearly-terms.png` and `tablet-ad-free-rtl-200-controls.png`.
The prior account captures
are the preceding-build reference; this slice did not capture a new before build.
Historical design/card-style results are summarized in the reports linked below.

The final fixture runner exited zero after graceful shutdown and cleaned its
disposable database resources. Logs: `approved-final-fixture--20261001T134126Z-2996895.log`
and `approved-fixture-cleanup--20261001T134905Z-3020324.log`.
P06's complete applicable local gate is now met, bringing the roadmap to **39/46**;
P03 and the six remaining Phase 5 gates stay open.

### Recovery and independent review

Contract tests failed before implementation for the missing fixed terms,
Daily provider kind and migration. An initial invalid-year test failed during
JSON serialization; using a serializable out-of-range year exposed the intended
contract failure before implementation. New widget tests failed on the absent
terms. The longer copy also exposed an old broad text matcher, repaired to assert
the exact allowance label, and a settings-menu test target, repaired to tap the
actual menu item. No test was skipped or assertion weakened.

Independent review caught harness scroll assumptions before final browser runs.
The harness now traverses real Tab events to reveal controls and recovery actions
instead of bypassing Flutter focus/scroll handling. Screenshot crops at intentional
scroll positions are distinguished from inaccessible content. One parent vet
invocation used a root-relative wrapper from `backend/`; the shell's missing-path
error was inspected, the path corrected, and the captured vet run passed.
A final roadmap tally helper initially omitted the `V`-prefixed Phase 4 items.
After inspecting that failure, the corrected identifier pattern verified 39/46
and the exact seven remaining gate IDs; no roadmap assertion was relaxed.

Independent backend/migration/API, frontend/fixture, active-document and Chromium
visual reviews found no actionable outstanding issue. The original 596-file
staging set was inventoried, and all 3,776 original index entries remained intact
before staging this slice. Game rules, offline host/probe and the three earlier
economy SQL migrations were byte-identical to the original index.
The complete final candidate inventory contains 764 paths (661 PNGs and 103 text
files), including the preserved staged work. Credential-signature and PNG-metadata
checks found no candidate secrets or unexpected payloads. Final review log:
`approved-final-review-fixed--20261001T135215Z-3026123.log`.

## Scope limits

Paid fixtures are trusted synthetic provider data through the real local
authority and PostgreSQL; they are not signed Apple/Google purchase evidence.
Weekly/Monthly/Yearly fixture remaining-time windows do not define store period
arithmetic. Native devices, screen readers, live ads/payments, store SDKs,
signing and production acceptance are deferred to their existing gates. The
dated [official policy matrix](../design/DESIGN-economy-and-stores.md#platform-applicability-matrix)
remains the research basis; no SDK is selected here.

The existing [before/after design evidence](2026-10-01-shared-styles-and-products.md)
remains historical. This slice changes account terms and eligibility, not game
rules or artwork. Browser account evidence does not revalidate the complete
gameplay/capture/breakpoint/orientation/motion matrix.
