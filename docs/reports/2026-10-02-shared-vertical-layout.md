# Shared vertical phone-reference layout — 2026-10-02

Run: `vertical-board-20261002`. The owner explicitly selected
[`cgmsart-phone.png`](../../client/screenshots/cgmsart-phone.png) for phone,
tablet and desktop, with larger items at wider widths, and requested deletion
of every other screenshot. This replaces the preceding repair's separate
phone pages, tablet panes and desktop workspaces. It does not undo the
[startup/read-only recovery repair](2026-10-02-adaptive-startup-recovery.md).


The subsequent private-hand refinement below supersedes this build's six-column
hand; the other results remain evidence for the recorded build.

## Implementation and reproduced regressions

The normal entry and online host now use one mounted vertical board: compact
round/status header, two-column public seats with overlapping card-stack
previews, action region, private hand, bottom actions. At normal 390px the
18-card reference hand uses six columns. Wider screens enlarge the same
arrangement. Auto still uses 600/1050 logical-pixel classes for sizing; it no
longer changes navigation or replaces entire workspaces.

Public groups open a chooser with full, individually selectable physical cards.
The chooser reads the current authorized projection, validates callbacks and
is removed with its owning board. Available Coup remains immediate. Compact
cards preserve single-tap selection, independent double-tap/keyboard inspection,
long-press suppression and the reference's small authorized copy numbers.
The demo preserves its existing scenario and uses the same vector suits/theme.

Online forms are expandable within the action region. Their controllers remain
mounted, while collapsed descendants are excluded from keyboard focus. A
separate PageStorage bucket prevents the ExpansionTile's saved boolean from
being read as a text-field scroll offset; the first candidate reproduced that
`bool`/`double` exception in browser and widget tests. Public empty formations,
financial/status labels, enlarged rank headers and required-decision expansion
are covered by regressions alongside selections, drafts/caret and pending IDs.

The original normal-entry browser check failed before implementation because
the phone action region was hidden behind the previous workspace navigation.
Updated tests exercise the requested composition while preserving the previous
privacy, command-identity, focus and gameplay assertions. Test harnesses use
real wheel/keyboard interaction with Flutter's painted scroll area and public
choosers instead of removed Table/Hand/Decisions buttons.

The full browser sequence also reproduced an offscreen dropdown after real
scrolling, resizing and history navigation: Tab moved focus to Suit while its
top remained at -296px, and Space opened an inaccessible menu. A matching widget
regression verifies the entire focused dropdown enters the painted viewport;
browser acceptance checks its geometry before keyboard activation. Wheel input
is not used to bypass this keyboard requirement.

## Verification

Final web build: `6f0c164d19ef4d9c693fef07adba2041024b6fd2ee29b763bcaf3fa951391e57`
(`main.dart.js` SHA-256). HTTP reads of `index.html`, `keyboard_bridge.js`,
`flutter_bootstrap.js` and `main.dart.js` on both the normal 8080 entry and the
isolated 8081 QA authority match the local release assets. The browser checks use
Auto and fresh contexts; captures are reviewed and removed rather than retained.

| Gate | Result |
|---|---|
| Flutter host | 258 tests pass; analyzer clean |
| Shared package | 75 tests pass; analyzer clean |
| Browser support | 29 tests pass, including action/required-response expansion and exact fault-diagnostic budgets |
| Format, assets | 48 Dart files unchanged by format check; all 208 artwork hashes/native dimensions pass |
| Normal entry | Chromium 153.0.8010.12, Firefox 155.0, WebKit 26.6: 64 cases each, no browser errors |
| Cold startup and automatic recovery | Eight cases pass, including three independent stopped-service launches |
| Online gameplay | Chromium 68 cases; Firefox and WebKit 34 cases each; zero unexpected browser errors |
| Long keyboard sequence | Firefox 48 cases pass, including live resizing, RTL and 200% text |
| Physical card inspection | Chromium 12 hand/public cases pass: pointer, double tap, keyboard, selection retention, zero command POSTs |
| Bounded reconnect | Two cases pass with original drafts/selections and zero command POSTs |
| Shared card styles | Three real capture/privacy/refresh/reload cases pass at phone, tablet and desktop sizes |

Normal-entry coverage includes direct phone/desktop loading; public/action/hand
vertical order and no replacement tabs; six hand columns with card widths
59, 119.4 and 230.3 logical pixels at 390/768/1440; live 599↔600 and
1049↔1050 transitions; both rotations; short viewports; 200% text; touch selection;
keyboard inspection and modal reverse traversal. Selections survive every
transition. The short-view keyboard check requires at least 48px of painted
card visibility, not merely a semantic node outside the viewport.

Cold-launch checks retain the same PostgreSQL volume and use new browser
contexts. Health readiness took 10.504–11.754 seconds; normal entry was usable
812–823ms after navigation, and online guest setup at 1.267–1.388 seconds from
root navigation. The deliberately unavailable session recovered automatically
at 3.742 seconds, outage reload at 2.387 seconds, and established-stream outage
at 1.526 seconds with the same persisted game and zero command POSTs. HTML503
recovery and terminal malformed200 handling also pass. Captured service logs
remain local; no persistent data was removed.

The inspected service logs contain the previously recorded Redis memory-overcommit
warning and no additional error/fatal/panic entries. No host sysctl was changed.
The release build retains the existing unused CupertinoIcons font warning;
asset verification and browser resource checks pass.

The supplemental reconnect check passes both opening and Justice: eight automatic
failed reads stop at the documented budget; one explicit retry while still offline
starts a fresh bounded attempt, then transport restoration recovers automatically
without another refresh. Draft quantity 3, the original physical selection, game,
decision, style scope and authority version remain; two real stream connections
and zero command POSTs are recorded for each case.

The longer journey reproduced a closed-board regression after immediate Coup:
the phase header was blank, settlement stayed collapsed, and End Turn remained
enabled. Five host regressions cover closed status, completion/final standings,
admission, superseded callbacks and ordinary-update draft/focus retention. The
fix announces closed status, opens meaningful completion transitions and disables
End Turn outside play without dispatching any command.

A supplemental inspection failure was isolated to automation: clicking an
offscreen inline Return control made the browser scroll its semantic overlay
horizontally while Flutter's canvas stayed fixed. Real wheel exposure before
touch, and keyboard Tab/Space with geometry assertions, preserve alignment.
These actual input paths replace offscreen locator autoscrolling.

The final complete journey passes opening and ordered responses, explicit purchase
payment, private trade/loan, attacks, compensation, Confinement, Justice, immediate
Coup, settlement, next-game reload, final ranks, owner/member lobby reload and
normal room deal. It preserves focused draft selections across all breakpoints
and uncertain operation identity across resizing/reload without extra commands.
The 48-case Firefox traversal and 12-case card inspection checks pass on this same
build. Browser results include only exact injected outage and observed anonymous
401 challenges as expected diagnostics.
Native-device and actual screen-reader certification remain separate roadmap gates.

The [compact verification record](2026-10-02-shared-vertical-verification.json)
retains all 401 browser case outcomes, build/report hashes, timings and conditions.
All owned QA authorities were stopped gracefully with exit 0; ordinary services
remain running. Prior backend test/vet results remain in the preceding repair
record; this follow-up changes frontend code only.

## Screenshot disposition and preserved work

The [inventory and retention report](2026-10-02-screenshot-retention.md) records
all removals and the unchanged phone SHA-256. Source artwork, mockups, fixtures,
application assets and lightweight historical reports/manifests remain.
The final scan confirms the phone reference is the only screenshot: 3,609
removed paths (including prior and current ignored captures), 487,153,597 recorded
bytes plus 808 earlier captures with unknown byte sizes. All 1,907 protected
artwork/mockup/non-image evidence hashes, 22 ignore checks and screenshot links
pass. Narrow ignore
rules exclude screenshot outputs with only the selected phone exception;
optional widget captures write to ignored browser artifacts and cannot replace
that reference.

Previously staged artwork, startup, privacy and other work is preserved. No
commit, push or history rewrite is performed. Historical screenshot blobs remain
in Git; the previous investigation did not establish a push-size rejection.


## Private-hand follow-up — `hand-suits-20261002`

The owner reported uneven hand-card sizes and requested horizontal stacking by
suit. The prior browser build reproduced widths of 59px but heights of 70.20–79.42px
at 390px: board cards inherited each painting's individual native aspect ratio.
The hand now uses uniform 4:5 frames containing uncropped, unstretched artwork.
Hearts, Spades, Diamonds and Clubs form horizontal overlapping groups, arranged
in two columns where space permits. Long groups scroll horizontally; enlarged
text can reflow to one column. Exposed strips retain at least 48 logical pixels
when fully revealed. Exact physical-copy labels, selection markers and focus
remain on the exposed leading edge in both directions. Inspectors retain native
artwork proportions.

Browser testing also demonstrated a keyboard regression in the initial overlap
implementation. After horizontal scrolling and resizing 390→768, the framework's
geometry-based semantics ordering removed and reinserted a focused Spade node.
Chromium and Firefox then left focus on BODY, and I no longer opened inspection.
A DOM mutation probe established the removal/reinsertion; a widget regression
independently reproduced changed semantics traversal order. Explicit ordinal
semantics keys now preserve the physical-card order while hit-testing keeps the
painted overlap order. A separate regression revealed a retained focused card
outside its suit viewport after shrinking; layout changes now trigger the same
mounted/current-route/current-focus guarded reveal. No engine or SDK patch,
blind retry, focus reassignment, or command replay is used.

The original size/overlap, resize-visibility and semantics-order tests failed
before their fixes. Previous center-tap tests now target the exposed card strip,
retaining exact-ID, one-selection-per-tap and no-selection-on-double-tap assertions.
The normal-entry browser gate traverses all 18 cards at 390/768/1440 before measuring
uniform frames, verifies nested clipping, horizontal scrolling, exact first/
middle/last number-card selection, keyboard inspection, live boundaries/rotation,
short viewports, 200% text and zero command POSTs. The online 12-case inspection
runner also covers RTL and independent full-card public inspectors; its opening
fixture contains one hand card, so it alone does not verify overlap in RTL.

The final source gates pass: 261 host tests, 76 package tests, 29 browser-support
tests, both analyzers, five-file formatting, and all 208 artwork hashes/native
dimensions. The release bundle's `main.dart.js` SHA-256 is
`2af8cae7ed3fd7f24d01f4c1f2d5b7294523ab2dc681874a3d723a4885b4b607`.
Chromium 153.0.8010.12, Firefox 155.0 and WebKit 26.6 each pass 74 normal-entry
cases on this build, with no browser errors or command POSTs. The online
inspection runner passes another 12 cases. At 390/768/1440 logical-pixel widths,
the hand frames measure approximately 72×90, 141.78×177.22 and 172.8×216 pixels.
Dense phone/desktop captures and the live tablet preview were visually reviewed.

The supplemental RTL fixture uses four private Queens, with two physical copies
in each of two suits. Chromium and Firefox each pass 12 additional cases: every
copy is selected by its exposed right edge, inspected through real Tab/I, and
cleared independently. Uniform frames, nested clipping, live 599/600 and
1049/1050 transitions, larger widths and 200% text all pass without commands.
This brings the follow-up to **258 browser cases**. The four-card RTL fixture
proves overlap; dense horizontal overflow is browser-tested in LTR and
widget-tested in both directions.

Two automation failures were corrected without changing application behavior:
semantic rectangles clipped by scrolling must be measured after keyboard reveal,
and wheel input must land in the painted hand viewport. A 30px clipped rectangle
had placed the wheel below the viewport; the helper now exposes at least 48px
of headroom. Existing strict clipping, selection and command assertions remain.
The supplemental runner also corrected its inspector-text locator and its
Firefox fixture preparation: forward Tab could not wrap to an earlier item in
the action menu, so setup now uses the painted menu's upward wheel and touch.
The hand's keyboard assertions remain unchanged. Both final 12-case runs pass
on the same runner source and release build.

All 235 temporary captures from this follow-up were removed after representative
visual review (97,689,601 bytes); compact reports retain the measured evidence.
The sole retained screenshot, `cgmsart-phone.png`, keeps its original SHA-256.
After acceptance, the owned QA runner was explicitly stopped with SIGTERM.
Its existing signal handler reports `InterruptedError` and exit 1; this was
inspected as intentional shutdown, not a failed browser case. The cleanup path
completed: no disposable integration containers, volumes or networks remain,
8081 is closed, and all four ordinary Compose services remain healthy on 8080.

Final verification is recorded in
[the hand-suit verification manifest](2026-10-02-hand-suit-verification.json).
This follow-up changes frontend presentation only. Earlier backend/startup results
remain historical evidence; native devices and screen readers are not certified.

## Compact public quadrants and visible hand follow-up

Run: `quadrant-board-20261002`. The owner's narrow preview measured 319×948
Flutter logical pixels (device pixel ratio 1.3). Its previous public grid switched
to one column below 330px of board width, while individual piles grew with card
count. The outer document scroll placed the first private card at y=957.35,
entirely below the viewport. The new initial-paint browser regression captured
that failure before scrolling, focus traversal or any reveal helper.

The shared board now budgets the actual host width and height, after online
headers and notices. Four equal public quadrants stay in a bounded square, with
the private hand immediately below. Dense hands scroll within their own pane;
long actions open within the public area and retain their mounted controllers.
Whole-quadrant controls open full authorized public-card details, allowing the
decorative previews to shrink without shrinking their touch targets. Existing
equal 4:5 hand cards and horizontal suit stacks remain.

Review reproduced another regression introduced by the bounded action overlay:
closing it could leave the host's form marked expanded, preventing a later
required decision or Trade action from reopening it. Closing now synchronizes
that state, and regressions exercise the actual Close actions control. Queued
opens validate the current game, generation, required state and route.
Required decisions arriving behind a public chooser prepare the panel beneath
the modal without changing focus, so dismissing the chooser exposes the decision.

Actual browser checks also found two issues that widget geometry alone missed.
The public semantic parent retained its old 120.24px extent after a live
RTL/200% change while the canvas and its children shrank to 110px. This persisted
for 120 animation frames. Giving the semantic boundary tight constraints inside
the square restores matching DOM and canvas geometry without replacing physical
card or seat identities. Material button padding and Wrap centering also clipped
a footer target to 47.5px (37px at enlarged text). Footer buttons now share a
reserved height of at least 48px, with enough additional room for enlarged label
lines; short landscape retains its hand target. Regression checks measure the
painted targets, semantic bounds and natural label height.
Visual review additionally caught enlarged Round/turn/Auto text clipped by the
old 48px header slot. The header now lays out at its natural height above the
remaining board area; that area's actual constraints determine the square and
hand budget. Enlarged text is not reduced to make the old toolbar height fit.

At the same height, tablet and desktop card sizes can reach the same height cap.
Tests require nondecreasing sizes across those widths and a strict increase when
both available dimensions grow, while retaining equal-card, full keyboard reveal,
48px target and initial hand-visibility assertions.

The owner also removed the remaining screenshot archive: 660 files and 250
subdirectories. Only the unchanged phone reference remains in `client/screenshots`;
all future output types there are ignored except that file. Historical counts
above describe their recorded runs, not current archive availability. The
[retention report](2026-10-02-screenshot-retention.md) records the exact inventory
and unresolved historical push diagnosis; no push or history rewrite was attempted.

The final source gates pass: **270 host tests, 76 package tests and 31 Node
support tests**, both analyzers, seven-file formatting and all 208 artwork
hashes/native dimensions. The final release bundle is
`69e8d529655e7556ed37afc879bbb3cd518ec53fe98bd8d5281509e831494a11`
(`main.dart.js` SHA-256). Normal entry on port 8080 matches this bundle and
passes **88 cases in each of Chromium, Firefox and WebKit**. These cover fresh
319/320px loading, equal quadrants, initial hand visibility, full authorized
public details, all 18 uniform hand frames, suit scrolling, touch and keyboard,
live boundaries, rotation, short heights and 200% text. No browser errors or
command POSTs occurred. The live 319×948 preview and enlarged desktop/short
landscape captures were visually reviewed; header and footer labels fit.

Online acceptance runs serially, as required by the harness documentation.
An earlier concurrent run saturated the shared fixture's HTTP rate limit;
those failures remain failures, with no expanded error allowlist. The bounded
action pane also required the harness to reveal painted controls before real
keyboard input, verify checkbox/dropdown state and assert exact draft values.
A native input's 8px browser border/padding overhang is not the painted Flutter
field; containment checks use its semantic field rectangle while input still
uses Tab and keyboard events. These corrections preserve strict target,
selection, authority-version and duplicate-command assertions.

Final browser acceptance totals **388 passing cases** on this bundle:
264 normal-entry cases, the complete 68-case online journey, 39 online geometry
cases (13 per engine), 12 physical-card inspection cases, two bounded reconnect
cases and three shared-style capture/privacy/reload cases. The online matrix
includes the real host toolbar, untouched initial geometry and exact draft
retention while opening/closing actions. The opening and Justice reconnect cases
each stop after eight failed automatic reads; an explicit offline retry also
remains bounded, and restoring transport recovers without refresh, reload or any
command POST. Exact injected faults and anonymous-session 401 challenges are
recorded separately from unexpected browser errors, of which there were none.

The [verification manifest](2026-10-02-quadrant-verification.json) retains every
case outcome, measurements, report/build hashes, test logs and inspected failure
history. Dense 18-card normal-entry browser coverage is LTR; the online geometry
fixture has one authorized private card. Widget regressions cover hand overlap
in both directions. Native devices and screen readers remain separate gates.
Earlier backend and three-cold-launch results above remain evidence for their
recorded builds, not new runs against this presentation-only follow-up.

After visual review, all **809 temporary captures / 173,736,401 bytes** from this
follow-up were inventoried and removed. The phone reference is unchanged, all
1,248 protected external hashes still match, and no images remain in browser
artifact output directories. The owned QA authority stopped gracefully with
exit 0; port 8081 and all disposable integration resources are gone. The four
ordinary services remain healthy and port 8080 serves the verified final bundle.

The complete diff and secret-pattern review passed. All local links in changed
documents resolve. A broader filesystem-only scan found 46 unresolved targets
across 21 unchanged agent-framework documents, each byte-identical to HEAD;
these are recorded in a separate `framework-docs` tracking note for follow-up.

## Public card growth at a fixed viewport height

The owner's 821×948 preview exposed a remaining sizing cap. The public square
used the smaller of 360px and 36% of the viewport height, so widening from 390px
to 821px or 1440px at the same height left it at 341.27px. A new browser regression
failed on that exact sequence against the preceding `69e8d529` bundle; a widget
regression independently reproduced the cap and now also checks transformed
painted-card bounds.

The public area's height share now grows continuously with width, from the
existing phone density to a larger desktop allocation. The remaining-height
budget still reserves the adjacent hand and bottom controls. Short viewports
constrain growth; the four equal quadrants, mounted controllers, hand sizing,
authorized projections and command callbacks are unchanged. This is a sizing
correction, not a cache or reload repair.

The final `main.dart.js` SHA-256 is
`3a25b7a5e052fbfd9f26694a8a26d18dd08243fe83c149c43a9be5c022777811`.
The normal preview and disposable online host both served that exact bundle
during acceptance.
At a constant 948px height, the normal public square measures approximately
303, 341.27, 403.53 and 492.95px at widths 319, 390, 821 and 1440px respectively;
the reverse sequence restores the same sizes without reload. Widget checks
measure painted card transforms, with artwork both enabled and disabled and
RTL/200% text. Browser checks measure the real public and seat bounds, while
visual review confirms the painted previews enlarge at 821px and 1440px.

All 271 host tests, 76 shared-package tests and 31 Node support tests pass, as
do both analyzers and formatting. One existing equal-cell assertion exposed
floating-point subtraction noise in global rectangles; it now compares exact
layout sizes directly, retaining its single-distinct-size assertion. The
analyzer's unnecessary-import failure was fixed before its clean rerun.
The first online growth fixture also correctly exposed the viewer's own
concealed Ace alongside their Queen; its new expectation was corrected to
require both exact authorized identities and exclude opponents' private cards.
The failed attempts are retained in the compact verification record.

The [public-growth manifest](2026-10-02-public-growth-verification.json) records
the final browser outcomes, build/report hashes, measurements, source logs,
failure diagnoses and cleanup. Earlier backend, cold-launch and reconnect
results above remain evidence for their recorded builds; those workflows were
not rerun for this presentation-only change. Native devices and screen readers
remain separate acceptance gates.

Final browser acceptance passes **351 cases**: 103 normal-entry cases and 14
online cases in each of Chromium, Firefox and WebKit, with zero unexpected
errors. Live resizing preserves the document and private identities, sends no
command POSTs and leaves the online authority version unchanged. The new build
is loaded in the existing preview; resizing itself requires no refresh.

After visual review, all **248 temporary captures / 61,526,125 bytes** were
inventoried and removed. Only the unchanged phone reference remains under
`client/screenshots`; all 1,248 protected external hashes still match. The
disposable QA host exited cleanly with no remaining integration resources;
the four ordinary services remain healthy and port 8080 serves the final build.

## Full-width public table and shared card scale

The subsequent review showed that width-responsive growth alone did not align
the public table with the hand. At 821×948, the hand was 805px wide while the
public square was 403.53px wide. A fixed 144×144 drawing inside each player cell,
followed by another fit for each pile, reduced three-pile public faces to about
65.26px versus the hand's 97.51px. Increasing the outer border alone would have
left that second bottleneck intact.

The owner explicitly selected a table as wide as the hand, with four equal
rectangular quadrants on larger screens. The table now spans the hand's exact
left/right edges. Its height remains independently bounded, retaining the
adjacent hand and footer. Public drawings use their actual cell width and the
same 72px base card width as the hand, scaled to the host's available size.
Piles fit down only when necessary, preserve each painting's native ratio and
cap total overlap depth at 16 unscaled pixels. Full authorized physical cards
remain available through the existing quadrant chooser.

New red/green regressions reproduce both the misaligned table and the small
painted faces. The dense-pile test covers eleven Diamonds, a Club and five Kings;
every painted public card must retain 90–100% of the hand width in roomy 821px
and 1440px views, including RTL and artwork fallback. Equal-cell checks now
assert full-width rectangular coverage. Cards that have reached the shared hand
scale stay at that size when additional width is available, matching the hand's
own behavior. Native proportions, containment, reverse resizing, visible hand,
focus, drafts, pending commands and privacy checks remain.

The [full-width manifest](2026-10-02-public-width-verification.json) records the
final bundle, source/browser results, inspected failures and capture cleanup.
The preceding sections remain historical evidence for their recorded builds.

Final acceptance uses `main.dart.js` SHA-256
`0f3908b52ecf21f4e232b7d538164bb61a8774c44ddf3a5603ea4a71aca6fe1e`:
**272 host + 76 shared-package + 31 Node tests** pass, both analyzers are clean,
and **351 browser cases** pass across Chromium, Firefox and WebKit (103 normal
entry + 14 online cases per engine), with zero unexpected browser errors.
At 948px height, the table widths are 303, 374, 805 and 1424px at viewport widths
319, 390, 821 and 1440px. Its height stops growing at approximately 470.73px;
the hand shares both horizontal edges and remains 4px below it. Reverse resizing
restores table and hand bounds, retains authorized card identities and sends no
commands or document navigations.

One new Firefox containment assertion initially failed because independently
rounded DOM rectangles differed by 0.016677856px at their bottom edges. The
captured probe and independent review established numerical rounding, with no
visible overflow. Only Firefox cell-edge containment allows one 1/60px layout
unit plus a 1/16384px floating-point margin; all other geometry, 48px target,
privacy, state and interaction checks retain their limits. The corrected run
passes all 103 normal-entry cases. Failed evidence is recorded in the manifest.

After visual review, **226 temporary captures / 67,513,510 bytes** were
inventoried and removed, leaving 5,553 unique inventory paths. The sole phone
reference and all 1,248 protected external hashes remain unchanged. The
disposable QA host exited successfully and left no integration resources;
the four ordinary services remain healthy and port 8080 serves the verified
bundle. A separate preview tab shows the new build, preserving the previous
tab's test position. Resizing the new build requires no reload.
