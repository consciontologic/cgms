# Nonproduction implementation and verification — 2026-10-01

> Archive removal update (2026-10-02): the owner requested removal of every
> file and subdirectory in `client/screenshots/` except `cgmsart-phone.png`,
> including raw reports, manifests, logs and copied harnesses. This document
> remains a historical account of the recorded results; its capture-time counts
> do not imply that those raw files remain available. Removed archive links are
> plain text below. See the [retention report](2026-10-02-screenshot-retention.md)
> and its per-file inventory for the disposition.

Run: `nonprod-20261001`. Starting tree was clean at `5f3ea28`; no commit or push
is performed by this run. [ROADMAP](../planning/ROADMAP.md) remains the sole
sequenced checklist. This report records evidence and limitations, not additional
gates. Phase 5 is excluded, including R01–R07/R06N, live commerce, deployment,
native/store release acceptance and host-reboot/production recovery drills.

## Application relocation

All 141 originally tracked application files, including hidden files, moved from
`docs/design/app/` directly into `client/`. Source/destination inventory found no
collisions. `client/offline_host/` and `client/offline_probe/` remain in place;
`client/packages/cgms_ui/` remains the reusable presentation boundary. There is no
`client/app/`, root `app/`, duplicate application or compatibility symlink.

Active paths were updated in Make/Python dispatchers, Compose static assets,
the loopback fixture server, browser harness, package/document links and skills.
CI had no independent old-path reference. Historical tracking rows, decisions,
reports and original browser manifests retain their execution paths and hashes.
The old tree is absent. Independent review compared every relocated tracked file
and confirmed that historical screenshots/manifests and artwork retain their
original bytes. New tests and browser evidence belong to the relocated build.

The local Flutter 3.47.5 SDK resolves the host and package. Normal tooling updated
the host's stale lockfile. Analysis also required six constant `FormatException`
expressions in the existing offline probe; its behavior is unchanged.
`make go.re MODE=local` and `make web.re MODE=local` ran successfully with the
project-local Flutter toolchain. The final online route was inspected at
`http://localhost:8080/#/online` through the real local Nginx/backend stack.
Existing local volumes were retained and economy activation remains off by default.
Fresh byte comparison against the normal developer server after the final build
passed `nonprod-local-release-final--20261001T061638Z-1832566`; index, bootstrap
and main script exactly match `client/build/web/`.

## Demonstrated defects and repairs

| Defect | Root repair and regression evidence |
|---|---|
| An ordinary command could overtake a victory arriving during database ranking | Dispatcher reselects against the live queue at the mutex-protected dispatch boundary. `backend-victory-red--20261001T051511Z-1613050` failed before the fix; `backend-victory-green--20261001T051527Z-1614495` passed. |
| Overlapping economy-enabled room starts deadlocked while upgrading shared identity locks | Acquire sorted exclusive identity locks before room/economy locks; economy operations use the same identity-first order. Real PostgreSQL RED reproduced SQLSTATE 40P01 (`backend-rooms-lock-red--20261001T051726Z-1624032`); GREEN `backend-rooms-lock-green--20261001T051816Z-1628118`. |
| Allowance-denied next-game admission repeatedly attempted writes and gave no useful client status | Bounded scheduler waits until an entitlement wake or at most 30 seconds/UTC reset; a private-safe, snapshot-only common wait flag exposes no account balance or identity. Regression, restart and real database evidence cover admission without partial charges. |
| Closed boards still announced the previous active player’s turn | Non-playing board projections no longer supply a turn label; the shared status strip omits empty turn semantics while keeping the actual phase. Screenshot review and widget RED `closed-board-turn-red--20261001T061254Z-1820406` reproduced the misleading ownership; GREEN `closed-board-turn-green--20261001T061329Z-1822035` reran all 154 host/50 package tests, both analyzers and the release build. |
| Streamed finalization did not fetch the snapshot-only wait flag | Refresh advisory data after `online.financial_phase` becomes finalized/void with games remaining. The corrected regression uses the actual settlement board plus financial phase; duplicates do not poll, game/cursor do not change. RED `stream-financial-boundary-red--20261001T053604Z-1702452`; GREEN `stream-final-boundary-green--20261001T053636Z-1704474`. |
| Room controls hid ownership/fullness and lost-start recovery | Nonowners cannot create invitations/start; incomplete rosters show a waiting reason; the original start operation remains recoverable even with a room loaded. |
| Browser accessibility exposed visible room/invitation codes as empty unnamed disabled fields | Replaced selectable/editable display fields with labelled static text and explicit Copy actions/feedback. Widget regression validates semantic values and exact clipboard content. An initial read-only-field fix passed widgets but still exposed an empty web input; actual browser testing rejected it. Final RED `room-code-value-red--20261001T054713Z-1742861`; GREEN `room-code-value-green--20261001T054746Z-1744553` (31 screen tests, analyzer, release build). |
| Ordinary actions and contextual trade acceptance appeared available during required decisions | Shared timing hints disable and explain incompatible actions and distinguish response from effect input. Server remains authoritative. Legal out-of-turn ordinary victory and draw-queue Coup remain available; financial automation blocks Coup. Independent review caught and repaired an overly broad own-turn restriction. |
| Short rotated screens at 200% text overflowed recovery/status content | Constrained scrolling preserves recovery controls and status; all three purpose-built compositions and cgmsart remain intact. |
| Clipboard permission denial incorrectly reported a network failure | Clipboard failures now give a localized permission/retry message without changing connection status or hiding the code. RED `clipboard-denial-red--20261001T055450Z-1764963`; GREEN `clipboard-denial-green--20261001T055516Z-1766671` verifies 32 screen tests, analysis and the final build. Chromium headless copy tests explicitly exercise denial, then grant only origin-scoped clipboard-write; Firefox/WebKit use their native user-gesture path. |
| Absent pass/premium timestamps displayed a 1970 expiry | Hide the backend's empty epoch sentinel; actual expiries remain UTC and server benefit flags decide access. RED `economy-expiry-ui-red--20261001T052810Z-1676054`; focused widget/client GREEN followed. |
| Account dialog inherited compact desktop density and reduced touch targets to 40 pixels | Standard density is applied at the dialog route so Spend, Refresh, Close and eligible Coup retain the 48-pixel minimum. Browser failure was reproduced by widget RED `economy-dialog-targets-red--20261001T055838Z-1776864`; GREEN `economy-dialog-targets-green--20261001T060205Z-1786272` passed all 35 affected tests, analysis and release build. |
| RTL toolbar clicks activated a neighboring button | Browser `elementFromPoint` and pointer events showed transparent Flutter leaf-button label text overflowing its 48-pixel semantics box. Clip only tappable leaf-button semantic overflow, retaining names and leaving nested controls untouched. Original browser failure `gates-economy-browser-chromium-fixed--20261001T060247Z-1788431`; diagnosis `gates-economy-rtl-event-debug-scoped--20261001T060535Z-1797508`; final build `semantics-hitbox-build--20261001T060649Z-1800675`. Browser hit-target regression is included in final economy acceptance. |
| Startup Retry used an inline handler rejected by strict CSP in Firefox | Registered the handler in the existing external bootstrap script without relaxing CSP. RED `nonprod-bootstrap-csp-red--20261001T060023Z-1781828`; GREEN `nonprod-bootstrap-node-final--20261001T060141Z-1785367` proves the retry reload and absence of inline handlers. |
| Strict configuration schema rejected the implemented economy flag | Added the optional boolean and a runtime/schema field/type regression. RED `economy-config-schema-red--20261001T053537Z-1701051`; race GREEN `economy-config-schema-green--20261001T053605Z-1702603`. |

All run names refer to captured `.cmd`, `.log` and `.exit` files under
`/tmp/agent-runs/`. Expected RED runs remain evidence; later passes do not erase
them. The first stream regression incorrectly modeled board finalization;
independent review corrected it before accepting the final result.

## Backend scope and reproducible checks

The review traced the accepted rules through shared engine, lifecycle, exact
ledger, matchstore, HTTP/stream transport and client projections. Existing sound
implementations were retained. The [accepted traceability matrix](../planning/ROADMAP.md#accepted-review-traceability)
and [client requirement map](../../client/README.md#online-requirement-evidence)
remain the detailed mapping for F01–F10/N01–N07/T01–T03.

| Area | Exercised implementation/evidence |
|---|---|
| F01–F03, F09, N01/N03/N04/N07 | Ability timing/cancellation matrices, victory legality and current-state priority, immutable decisions, physical costs/availability and exact prices in `internal/game`; ordered transport admission and client timing regressions. |
| F04–F05, N05–N06 | Lifecycle/settlement/review suites, origin-game receipts/corrections, duplicate-before-game-scope checks, pending work/restart tests in `internal/matchstore` and `internal/transport`. No rounded accounting or retargeted commands. |
| F06–F08, F10, N02 | Acquisition/custody/binding engine cases, privacy integration and online projection tests. Opponent faces and private finances remain absent from unauthorized snapshots/events/UI semantics. |
| T01–T03 | Proposal revision/expiry, loan/alliance lifecycle, accepted-transfer responses, deferred draw order and explicit purchase quantities, plus integrated browser scenarios. |
| Resource ownership and shutdown | Delivery cancellation/queue/race tests; stream/runtime supervisor lifecycle, durable cursor ordering, no implicit human pass, bounded allowance retry cache. |
| Economy and provider boundaries | Concurrent allowance/pass/reward transactions, outer rollback, receipt lookup, account deletion, expiry/restore/duplicate/revocation and ambiguous-result provider fixtures; no public provider or promotion mutation route. |

Fresh pre-browser backend checks passed:

- `backend-final-unit-race--20261001T052456Z-1656531`: game, matchstore,
  transport, delivery, rooms, identity, presence, wire and economy under `-race`.
- `backend-final-pg-race--20261001T052457Z-1656987`: real PostgreSQL/Redis race
  integration for matchstore, rooms, identity, presence and economy.
- `backend-final-transport-pg-race--20261001T052458Z-1657384`: named nonproduction
  transport suites (automatic work, origin/auth/schema, streams, privacy telemetry,
  economy, runtime/supervisor, concurrent starts, durable events and HTTP).
- `backend-final-vet--20261001T052543Z-1661870`: backend static checks.
- Independent reviewer: focused real PostgreSQL HTTP/admission race check
  `review-economy-http-focused--20261001T053446Z-1697608`, plus independent
  economy/rooms/matchstore race and client/config review tests.

One broad reviewer transport invocation accidentally included excluded Phase 5
capacity tests and failed `TestOperationalCapacityAndTelemetryLoad/streams=64/exporter=false`
at `load_integration_test.go:201` with HTTP 503. Captured run
`review-economy-pg--20261001T053136Z-1687006` remains **failed**; it is not converted
to a pass by the narrower authorized checks. An earlier broad backend audit run
also encountered that excluded load case. No production capacity acceptance is
claimed and no load assertions were weakened or removed.

A broad filesystem-only Markdown diagnostic also failed
`nonprod-doc-links--20261001T063507Z-1907120`: unchanged agent templates/vendor
adapters contain legacy relative links, a missing `CLAUDE.md` reference and a
provider-specific `/memories/session/` link. These are outside the application
relocation and backend/client work; they are recorded in tracking rather than
silently counted as repaired. The scoped changed/moved documentation check passed
375 links across 29 Markdown files after final documentation updates
(`nonprod-final-files-audit--20261001T064648Z-1940647`).
The broad diagnostic is retained as failed and is not part of P04/P05 acceptance.

## Economy decisions and remaining limits

Decision 0014 adopts **10 dirt per eligible finalized online game** and **30 dirt
per pass day**, up to six days/180 dirt per redemption. The adopted UTC reset,
three charged free starts, unfinished-game no-refund/no-reward policy and ad-free
pass stacking are preserved. Offline earns no dirt. Authenticated account and pass
routes expose only the caller's account; explicit local activation selects the
fixed adopted policy. The client stores an account-scoped original pass operation
before sending it and recovers its receipt without an automatic second spend.

**P05 passed** all decidable local acceptance. The final
economy manifest
contains three reports, 48 screenshots and **24/24 browser cases**, with no
unexpected diagnostics, using the same final build and harness. Each engine
exercises private balance/UTC allowance, explicit one-to-six-day pricing, lost
purchase response followed by reload and receipt-only recovery, no duplicate
charge, and a known committed pass despite a failed subsequent refresh. A real
three-seat game finalizes and earns 10 dirt per account; exhausted admission
appears through the stream, then all players spend 30 dirt and the next game
starts through streamed updates without reload/manual refresh. Initial funding
uses explicitly synthetic prior-day finalized ledgers; normal authenticated
HTTP and PostgreSQL handle the tested commands.

Final logs: `gates-economy-browser-frozen-chromium--20261001T061947Z-1841721`,
`gates-economy-browser-frozen-firefox--20261001T062039Z-1844166`, and
`gates-economy-browser-frozen-webkit--20261001T061855Z-1837740`.

P03 remains blocked by the owner's explicit choice to withhold a new compute
allocation. The previously proposed 300 hours/230 GiB and 100 hours/90 GiB were
not approved. Bounded preparation ran 67 simulator analysis/orchestration tests
(`gates-p03-preparation-discover--20261001T052850Z-1679375`); these are mocked
preparation tests, **zero new evaluation samples**. Existing incomplete campaigns
and replay checks do not establish calibrated strength. Held-out paired seeds,
seat rotation, separate three/four-player results, confidence intervals, decision
budgets and complete/incomplete denominators remain required for future evidence.

P06's dated [official Apple/Google applicability assessment](../design/DESIGN-economy-and-stores.md)
was refreshed before any SDK selection. Local provider/promotion fixtures cover
entitlements, redemption, restore, duplicates, expiry and ambiguous results.
The owner kept the cosmetic catalogue and promotional channels undecided, so
purchase/redemption entry points remain disabled and P06 stays unchecked.
Signed native purchases, real store sandbox/device acceptance and all production
work remain excluded. Browser accessibility checks cover semantics and layout,
not certification with native screen readers or an unprovided translation catalogue.

## Final integrated evidence

Independent verifier `final-verifier-go-race--20261001T053839Z-1712406` passed
the full Go race suite and vet: 19 tested packages plus four without tests,
including simulator (533.926 seconds) and offline runtime (300.526 seconds).
Integration-tag suites are separately represented by the real database runs above;
unit-only execution does not substitute for database evidence.
Independent final checks passed **154 host tests, 50 reusable-package tests and
both analyzers** on the final source (`final-review-host-tests--20261001T060426Z-1793704`,
`final-review-package-tests--20261001T060445Z-1794860`,
`final-review-host-analyze--20261001T060445Z-1794880`,
`final-review-package-analyze--20261001T060454Z-1795553`). Seven Node bootstrap
and diagnostic tests passed `final-review-bootstrap-fault-tests--20261001T060427Z-1793919`.
The final closed-board status repair reran those full Flutter suites and build
(`closed-board-turn-green--20261001T061329Z-1822035`), with an independent review
of the change and red/green evidence. Independent Chromium, Firefox and containerized
WebKit DOM probes additionally verified the semantic leaf clipping, unchanged
accessible names, keyboard focus and exclusion of nested controls.
Manual inspection through the normal local developer stack confirmed the actual
room code appears in browser semantics, ownership/waiting controls are correct,
and the account dialog opens and closes. The old empty text field is absent.
Offline host process/restart/tutorial checks and the independent Go/Dart board,
finance and complete-game restart probes passed on the pinned toolchain. Probe
artifact: ignored `sims/artifacts/offline-probe/nonprod-verifier-20261001/report.json`;
captured log `final-verifier-offline-probe-corrected--20261001T053936Z-1716917`.
These are host checks, with no Android/iOS device acceptance.

Browser harness repairs preserved the product assertions: visible content queries
exclude Flutter's duplicate live-announcement mirror; responsive tests wait for
new viewport geometry before hit testing; synthetic setup requests are paced
within the unchanged server admission limits. Multi-seat settlement setup explicitly
refreshes a seat before its confirmation to avoid intentionally rejected stale
versions. The finalization-to-allowance-wait and pass-to-next-game assertions still
require streamed updates without reload or a manual refresh. Failed harness runs
remain in captured logs; no repeated command was used to turn a 409 into success.

**P04 reverified:** the final baseline manifest
contains 16 passing reports, **340/340 case executions** and 74 attributed final
screenshots. Chromium contributed 132, Firefox 104 and WebKit 104, with zero
unexpected browser errors. Each engine covers phone, tablet and desktop full
journeys plus all 22 ordered viewport/transition samples and RTL/200%/placeholder
matrix cases; Chromium adds a full three-seat variation. Exact-CSP local rendering
passed separately in all three engines. Final manifest integrity passed
`nonprod-final-manifest--20261001T063815Z-1921294`.

The normal room/deal/reload flow and synthetic rule-boundary positions both use
the real local authority. The matrix exercises legal/blocked actions, decisions,
trades, loans, settlement, next-game transitions, keyboard/touch, private/public
semantics and uncertain-command recovery. Final phone/tablet/desktop screenshots
were visually reviewed. This proves the recorded browser scope, not a complete
natural-deal match, native accessibility certification or store acceptance.

Final build SHA-256: `aecc22cf28a1cdcf7e6ea6164f76be21a9e7706990d416e82e5d14f6acdcddf8`.
Index/bootstrap/harness and every archived report/screenshot hash are recorded
in the manifests. Together, P04 and P05 contribute **364 browser case executions**.
Final integrity check `nonprod-final-integrity--20261001T064617Z-1939387`
verified all 19 reports, 122 screenshots, their build/harness hashes and the
roadmap counts. The final file audit also verified every original relocated path,
78 unchanged asset/history files (excluding Markdown whose relative links changed),
and the complete 350-file proposed staging set for unexpected generated files,
symlinks and secret signatures. An earlier audit incorrectly froze relocation
links inside `NOTICE.md`; the corrected audit preserves actual historical bytes
while checking the changed links at their new location.
Firefox and WebKit ran on separate loopback fixture processes with isolated
PostgreSQL schemas and per-process admission state; server limits were unchanged.
Owned QA processes exited cleanly and disposable resources were removed. The
normal developer server at `localhost:8080` remains available.

Roadmap totals are **38/46**, with Phase 3 **4/6**. P05 is newly complete and P04
was reopened, repaired and reverified. P03 and P06 remain blocked exactly as
recorded above; all Phase 5 operations/native release acceptance remain excluded.


## Independent follow-up — nonprod-recheck-20261001

This follow-up started from a clean tree at `dd2858c`, where the direct relocation
and preceding results above were already committed. Those historical results and
hashes remain unchanged. Live inspection confirmed the accepted 10-dirt reward,
30-dirt pass-day policy and decision 0014's outstanding P03/P06 decisions. No
production operations or newly allocated bot evaluation ran.

All 141 original tracked application paths and modes map directly into `client/`,
including hidden files, with no old tree, duplicate host or compatibility symlink.
The offline tools and reusable package remain present. Independent migration
verification passed `final-review-migration-fixed--20261001T071459Z-2030357`,
including 79 immutable artwork/font/license/history files. Its first diagnostic
incorrectly required relocated Markdown links in `NOTICE.md` to be byte-identical;
the failed diagnostic remains recorded and the corrected scope preserves actual
historical execution evidence. Original 19 reports/122 screenshot hashes also
passed `recheck-prior-evidence-fixed--20261001T065851Z-1976352`.

### Additional demonstrated repairs

| Defect | Repair and reproducible evidence |
|---|---|
| A newly queued active-player ordinary command could lose to an older lower-priority command while database ranking was in progress | Re-rank when ordinary queue membership changes before admission; victory still wins. Deterministic RED `review2-dispatch-active-red--20261001T065451Z-1961828`; GREEN `review2-dispatch-active-green--20261001T065515Z-1964403`; independent race review `final-review-dispatcher--20261001T070050Z-1983968`. |
| Membership lookup database failure was misclassified as unauthorized | Only absent membership maps to unauthorized; infrastructure faults preserve the sanitized unavailable response. Real PostgreSQL RED `review2-priority-membership-red--20261001T065915Z-1977512`; final database/race run below passes. |
| Impossible Baron and spent-Ace variants exhausted the fixed local bot menu budget | Prune impossible candidates before the unchanged 50,000-attempt limit. Deterministic dense-board RED `review2-menu-budget-red--20261001T070316Z-1991424`; the regression compares all 1,765 legal choices against a larger-budget menu and verifies state immutability. Ten repeats passed `review2-menu-regression-repeat--20261001T070558Z-1999095`; independent race check passed `final-review-menu--20261001T070547Z-1998452`. |
| Switching from Purchase payment selection to Open Series silently selected cards into a hidden field | Rebind the visible selection destination to the selected action; clear inactive substitution targets. RED `client-draft-regression-red--20261001T065552Z-1966270`; GREEN `client-draft-regression-green--20261001T065657Z-1968864`. Removed-substitution RED `client-substitution-regression-red--20261001T070113Z-1985354`; independent GREEN `final-review-client-fixed--20261001T070420Z-1994523`. |
| Reloading an unfinished room lost the lobby and the user's owner/member context | Persist only the public room ID in the route; authorize through session bootstrap before restoring room state, including explicit sign-in recovery. Invitation tokens stay out of the route. Lobby RED `client-lobby-regression-red--20261001T065815Z-1974801`, GREEN `client-lobby-regression-green--20261001T065859Z-1976666`; expired-session RED `client-lobby-auth-red--20261001T070244Z-1988795`, independent GREEN `final-review-client-fixed--20261001T070420Z-1994523`. |
| Action forms omitted declaration/resolution/cancellation cost explanations and own Clubs' usage marker | Add ARB-backed selected-action guidance derived from accepted §2.12 and a privacy-safe own-Club marker from the existing `used_turn` projection. RED `client-cost-guidance-red--20261001T070638Z-2003482`; GREEN `client-cost-guidance-green--20261001T070942Z-2014711`; independent final guidance/marker check `final-review-costs--20261001T071148Z-2019800`. |

The original broad race run `review2-full-backend-race--20261001T065735Z-1971950`
failed the random four-player offline journey at step 346 with menu-budget
exhaustion. Its private checkpoint was not retained by that existing test, so the
deterministic dense-board regression establishes the same mechanism, not replay
of that exact original board. This failed run remains failed. No budget increase,
assertion weakening, automatic game pass or legal-action removal was used.
A suspected suit-menu reset was tested and refuted; its sound behavior remains.

### Fresh backend, economy and preparation verification

- Full final `go test -race`: `review2-full-backend-final-race--20261001T070616Z-2000479`
  passed all 19 tested packages (four have no tests), including complete three- and
  four-player offline journeys (268.555 seconds) and simulator tests (464.982 seconds).
- Named real PostgreSQL/Redis matchstore/transport race checks passed
  `review2-game-final-pg-race--20261001T070617Z-2001262`; HTTP/stream, authorization,
  immutable receipts, trades/loans, settlement, restart and supervisor lifecycle
  also passed `review2-final-http-pg-race--20261001T065703Z-1969861`.
- Vet passed `review2-final-menu-vet--20261001T070559Z-1999400`; rebuilt offline
  host/Dart analysis, process/restart/tutorial/stale-command checks passed
  `review2-offline-host-final-journey--20261001T071136Z-2019024`.
- Independent economy PostgreSQL race checks passed
  `economy-audit-pg-race--20261001T065452Z-1962516`, covering atomic redemption,
  entitlement/restore/expiry/duplicates, account binding, rollback, ambiguous
  provider outcomes and sandbox isolation. No new local contract gap was found.
- P03's 67 mocked preparation tests passed
  `economy-audit-p03-prep--20261001T065515Z-1964306`: zero new evaluation samples,
  no calibrated strength claim, and no approval of either proposed allocation.
- Nineteen dispatcher/tooling tests, thirteen integration-runner tests, seven
  Node bootstrap/diagnostic tests and the 52-artwork/package-boundary validation
  passed in the `recheck-tooling`, `recheck-integration-runner`,
  `recheck-web-bootstrap` and `recheck-assets` captured runs.

Official policy applicability was rechecked on 2026-10-01 before any SDK choice.
[Apple's review guidelines](https://developer.apple.com/app-store/review/guidelines/)
still require applicable IAP/restore handling and distinguish storefront-specific
external purchase exceptions. [Google's payments policy](https://support.google.com/googleplay/android-developer/answer/9858738?hl=en)
and [points/currency guidance](https://support.google.com/googleplay/android-developer/answer/10281818?hl=en)
distinguish earned points from sold digital currency and regional programs.
[Google's billing security guidance](https://developer.android.com/google/play/billing/security)
requires backend verification, unique purchase identity, account binding and
correct pending/revoked handling. [Apple offer-code documentation](https://developer.apple.com/help/app-store-connect/manage-in-app-purchases/create-offer-codes-for-in-app-purchases)
was readable; the Google promotional-code help page was verified only through
official search results because its full page returned an error/challenge.
No claim of a full fresh Google promotional-page read is made. Existing applicable
local contracts remain valid; catalogue/channel decisions and native/store work
remain unresolved/excluded respectively.


### Final client verification

After detecting a separate IDE agent's Flutter 3.38.5 commands overwriting shared
dependency state, the checks explicitly selected both `PATH` and `FLUTTER_ROOT`
for the pinned Flutter 3.47.5/Dart 3.13.4 toolchain. Host and package dependency
configurations were verified afterward; both lockfiles match the starting commit.
A concurrent test import edit was repaired narrowly to retain the existing
`EconomyPanel` import. Final source checks passed **163 host tests and 50 reusable
package tests**, with both analyzers reporting no issues and 44 Dart files formatting without changes:

- `client-review-pinned-analyze-final--20261001T071712Z-2035623`
- `client-review-pinned-tests-final--20261001T071724Z-2036186`
- `client-review-pinned-package-analyze--20261001T071750Z-2038183`
- `client-review-pinned-package-tests--20261001T071759Z-2038701`
- `client-review-format-final--20261001T071451Z-2029880`

Independent review passed the nine focused selection/lobby/guidance/usage tests
and found no remaining correctness/privacy findings. Asset checks preserve
cgmsart, shared card backs, local fonts and the reusable package boundary.


The final local developer rebuilds passed
`recheck-final-local-backend--20261001T071423Z-2027509` and
`recheck-final-web--20261001T071832Z-2040226`. Normal Nginx-served index, bootstrap
and main script exactly matched `client/build/web/` in
`recheck-served-final-assets--20261001T071949Z-2044515`; both dependency
configurations still used Flutter 3.47.5/Dart 3.13.4. Manual browser inspection at
`localhost:8080` created a real room, then reloaded it into the same authorized
owner lobby with correct seat count, invitation permission and disabled start.

The fresh browser wrapper initially probed a nonexistent fixture health endpoint;
the fixture intentionally exposes assets and its QA/API routes. Readiness now
requires its root index bytes to equal the final local build. A subsequent port
probe was corrected to distinguish closed TCP connections from a live listener.
These captured setup failures executed no browser cases. The first actual browser
attempt then exposed a harness traversal error when returning from Purchase to
Open Series: the helper searched downward for an option above the current item.
Keyboard traversal now goes upward for that target, retaining the original
visibility/selection/submission assertions. The failed report and screenshot are
retained in the new evidence folder's `failures/` subtree and excluded from
passing-case totals.


The first browser failure also exposed a wrapper-only cleanup error: Go reused a
cached `webfixture` executable outside the assumed `/exe/` path, so the wrapper
never sent the intended signal. The independently identified owned fixture exited
within 0.1 seconds after SIGTERM; its disposable resources were removed and verified
absent (`recheck-owned-fixture-cleanup--20261001T072638Z-2061127`). No backend
shutdown defect was demonstrated. The corrected wrapper identifies owned descendant
executables through `/proc/<pid>/exe` and signals the actual fixture, allowing its
runner to perform normal cleanup.

The next browser attempt reached Open Series correctly but the new heading-only
locator did not match Flutter's merged explanatory text node. The corrected
locator targets the rendered semantics subtree and, for Attack, asserts the full
declaration/cancellation/resolution wording rather than only a prefix. That failed
report also remains under `failures/`; neither failure is counted as a pass.

Independent review of the first final Chromium phone report and all seven
associated PNGs passed `final-review-phone-evidence--20261001T073151Z-2074266`:
30 passing cases, no unexpected diagnostics, and exact current build/harness
hashes. Its screenshot review confirmed readable guidance, public/private surface
separation and reachable landscape/200%-text recovery. Later engine/class reports
must independently pass; the phone result is not extrapolated to them.

The subsequent Firefox desktop run failed the 48-pixel touch check because a
previous button's semantics rectangle was clipped to 8 pixels at the pane edge
after keyboard navigation. This was independently reproduced in
`client-touch-firefox-red--20261001T074528Z-2110688`. The failing report/screenshot
remain under `failures/clipped-target-firefox-1440/`. Earlier passing browser
reports from the preceding harness are preserved under `superseded/` and excluded
from the final acceptance totals; they are not relabelled as failed or rewritten.

The geometry correction preserves every button from the original semantic-control
set. It exposes each through actual wheel input, checks containment against all
ancestor/viewport clipping, performs a trial pointer hit-test without activation,
and retains the unchanged 48-pixel minimum. No small controls are filtered out.
Independent review found no weakened assertions. Focused Firefox desktop GREEN
`client-touch-firefox-green--20261001T074956Z-2120846` measured all 14 controls,
including Clear selection at 551.6×48 and Submit at 551.6×49, then completed ordered
opening responses with no unexpected errors. This checks currently exposed controls,
not every control hidden inside unvisited menus.

For the final rerun, engines use separate loopback ports 8131/8132/8133 and separate
disposable PostgreSQL instances. Each engine's journeys remain serial; production
admission limits are unchanged. The existing integration runner owns resources;
a temporary wrapper changes only the fixture's loopback listen address. Screenshot
names include engine and viewport, preventing cross-engine archive collisions.


### Final follow-up acceptance

The final follow-up manifest
contains **420/420 passing case executions**, 19 reports and
145 final screenshots with zero unexpected browser diagnostics.
Baseline accounts for 396 cases and economy for 24; Chromium contributes
162, Firefox 129 and WebKit 129. Each engine covers its three representative layouts,
22 ordered viewport/transition samples and RTL/200%/placeholder cases, economy
journeys and strict-CSP rendering. Chromium additionally covers three-player play.
All final reports identify the same application and current applicable harness;
3 failed attempts and 9 superseded reports are separately preserved and
excluded from passing totals. Reference screenshots were visually reviewed.

Final captured engine runs:

- `recheck-final-parallel-chromium--20261001T075050Z-2123189`
- `recheck-final-parallel-firefox--20261001T075050Z-2123190`
- `recheck-final-parallel-webkit--20261001T075050Z-2123198`

Main build SHA-256: `4958712221bc8450841ef714142d1d0d531bd3fd0f6b705bcdb375e1b6bbd640`.
Every archived report/screenshot and current entrypoint/bootstrap/harness digest
is recorded in the manifest. Owned fixtures exit cleanly and dispose their test
resources; normal local developer data/services remain intact.

O04, P02 and P04 were reopened for demonstrated deficiencies, repaired and
reclosed against fresh evidence. P05 remains complete after independent local
reverification; it was already checked at this follow-up's starting commit.
The roadmap remains **38/46**, with Phase 3 **4/6**. P03 stays unchecked because
compute allocation is explicitly withheld; P06 stays unchecked because the
cosmetic catalogue and promotional channels are undecided. No calibrated bot
strength, native/store acceptance or Phase 5 completion is claimed. Verified
changes are eligible for tracking/staging; no commit or push is performed.

Final evidence checks passed `recheck-final-evidence-audit--20261001T080609Z-2193348`
and `recheck-final-runtime-integrity--20261001T080643Z-2194507`: all current build
bytes still match the normal local server, all three engine batches exited zero,
all six owned fixtures shut down cleanly, and the original report prefix and
dependency lockfiles remain unchanged. Independent review verified all 247
manifest-listed file hashes and exact artifact coverage. Manifest SHA-256:
`4d766379ed4006dd22ddb6a36437c8173614bd8f0a4d48536de4ae8269e5315b`.
`make roadmap.status` independently confirmed 38/46 in
`recheck-roadmap-status--20261001T080618Z-2193692`.

The complete pre-staging inventory passed
`recheck-complete-diff-audit--20261001T080717Z-2195927`: 26 changed tracked files
and 248 new evidence files (216 PNGs and 32 JSON files, including preserved
failures/superseded evidence), with no unexpected paths, generated dependency
directories or nonempty sensitive JSON fields. Added-text credential checks and
198 local documentation links also passed. The coordinating run is
`nonprod-recheck-20261001`; P03/P06 blocker rows remain separate from the verified
completion row.
