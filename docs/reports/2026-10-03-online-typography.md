# Online typography and alignment

Run: `online-type-20261003`. Scope: online lobby hierarchy and dropdown presentation.
Baseline: `de66e5894d63dc470b9ab05c1ca863a443b7bd1a`.

## Reproduction and cause

Opened the normal `http://localhost:8080/#/online` preview and its Bot difficulty
menu at desktop and measured narrow widths. The served and local release bundle
matched SHA-256 `210401b46cfe9c7df944dd1d88a403d50739ce32a3ce2d716624210392a58234`;
this was the current committed implementation, not a stale build.

Online dropdowns omitted an explicit text style. Flutter's legacy dropdown widget
therefore inherited the theme's 16px bold title style for selected values and
popup rows. Lobby introduction and section widgets also consistently used leading
alignment and the same large heading treatment, weakening the visual hierarchy.
The desktop action form and its long dropdown additionally used nearly the full
viewport width and height, leaving short option labels at the extreme leading edge.

The fix is local to the online screen: regular 15px body text for selectors,
intrinsic menu row heights with 48px minimum targets, centered introductory and
section text, and distinct secondary headings. Form labels retain their reading
direction. Global theme tokens, artwork, game rules and command ownership stay intact.
The gameplay form is centered at a maximum 640px reading width; menus scroll
within 60% of the viewport height, up to 400px.

## Verification

The first rebuilt preview exposed an additional overlay issue at 319×844 with
the in-app **200% text** setting: form text enlarged, but popup choices retained
ordinary-size text. The earlier widget fixture supplied enlarged text above the
Navigator, masking the route's loss of the screen-local setting. The corrected
fixture toggles the actual setting from an ordinary-size app root. Popup items
now carry the effective screen-local text scale and direction into the route.
Selected values also grow to their wrapped height without taking the height of
the longest unselected option.

Six new regression cases cover centered hierarchy; selected/open lobby type;
conditional action selectors; centered desktop width and menu height; actual
200%/RTL overlay, wrapping and keyboard selection; and closed semantic targets.
The full host suite passed **297 tests**, including the first five new cases and
the complete **82-case online suite**. The sixth, test-only semantic target case
was added afterward and passed separately: **298 distinct frontend tests passed**.
Final formatting and analyzer checks passed. No production source changed after
the full-suite/build freeze. Independent review reported no actionable findings.

The final release bundle served by the rebuilt ordinary local web service matches
the local file: SHA-256
`c70989b49b849851821f6c6152d02003be8903dbdf7dcaf719049ee84cc70c1b`.
The verified online source SHA-256 is
`8e5c5930e5350143a92a654998bf0c8cea540fde386bd0fed07af9b14278caa8`.

Browser review used the normal online route and an existing bot match in the
in-app Chromium browser. The centered lobby was reviewed at 319×844, 390×844,
768×1024 and 1440×900 logical viewports across the builds; the final build was
rechecked at narrow, tablet and restored desktop sizes. A difficulty selection
survived live resizing. The final gameplay form measured 640px wide and centered
at 1280px; its open popup measured 680×400px including route padding. At 319×844,
the actual 200%/RTL settings produced 30px, right-aligned, scrolling options and
fully wrapped long selections. Keyboard selection and Escape dismissal worked.
No game command was submitted during this typography review, and no browser
console warnings or errors were recorded. The preview remains open at normal
text size and direction with the temporary viewport override removed.

A suspected 44px closed target proved to be a partially clipped semantic box at
the scroll boundary. Once fully visible it measured 640×58px; the added regression
checks both dimensions against 48px. Its first diagnostic run failed because a
test platform override was not reset before framework teardown; a `finally`
cleanup fixed the fixture, with no production change or weakened assertion.

Captured logs under `/tmp/agent-runs/`:

| Check | Log filename | Result |
|---|---|---|
| Original typography regression | `online-typography-red-corrected--20261003T064627Z-3384027.log` | Expected RED |
| Selected text clipping | `online-typography-selected-red--20261003T065043Z-3392784.log` | Expected RED |
| Screen-local overlay scale | `online-typography-overlay-red--20261003T065543Z-3404557.log` | Expected RED |
| Desktop form bounds | `online-typography-bounds-red--20261003T065742Z-3409321.log` | Expected RED |
| Full host suite | `online-type-full-host--20261003T065939Z-3413950.log` | 297 passed |
| Online suite | `online-typography-complete-suite--20261003T065935Z-3413539.log` | 82 passed; included above |
| Additional semantic target | `online-typography-target-verified--20261003T070553Z-3427649.log` | 1 passed |
| Final analyzer | `online-typography-target-analysis--20261003T070555Z-3427826.log` | Passed |
| Final format | `online-typography-target-format--20261003T070553Z-3427590.log` | Passed |
| Build/recreate web | `online-type-final-web--20261003T065939Z-3413939.log` | Passed; persistent volumes retained |
| Served build identity | `online-type-final-identity--20261003T070048Z-3417662.log` | Matched |

## Evidence limits and retention

Browser observations cover one local Chromium surface, not a fresh Firefox/WebKit
matrix, native releases or accessibility certification. The lobby tab reports
devicePixelRatio 1.3; acceptance dimensions use observed CSS/logical width, not
the viewport override alone. Screenshots were inspected inline without creating
image files. The existing `client/screenshots/cgmsart-phone.png` reference and
earlier reports are preserved. No backend change was made in this slice; earlier
bot verification remains separate evidence. No unresolved blocker remains for
this typography fix; broader roadmap gates retain their existing status.

## Compact action workspace follow-up

Run: `action-density-20261003`; preserves the eight files staged by the earlier
typography run. Reproduced the owner’s current match at 1449×948 logical pixels
in the normal online route. The action workspace had a redundant 48px title/close
strip, a narrow vertical form, repeated preparation heading, and always-expanded
guidance/quota text. Its text also inherited the table artwork size multiplier.
The ordinary action controls consequently required substantial inner scrolling
despite unused horizontal space. Browser console warnings/errors were empty.

This follow-up removes the strip, shares open/close on the existing header tool,
and compacts the form into responsive rows with optional guidance. It supersedes
the earlier 640px action-column decision. Action text is 14px with 18px local
icons and user accessibility scaling, independent of the artwork multiplier.
Selected quantities, calculated costs, timing restrictions, required responses
and actionable records stay outside the optional guidance disclosure.

The first browser rebuild confirmed compact controls but exposed a second cause:
the fixed table-height slot left unused space below the now-short form and still
squeezed the hand. A new failing adjacency regression measured the hand at
551.86px rather than the expected 177px beneath its short fixture. The action
slot now sizes naturally up to its former height cap, with both public/action
subtrees retained offstage. Short forms return height to the hand; long forms
retain bounded scrolling. Closing actions restores the original table geometry.
Independent review found no blocking issues in either change.

Final verification passed **299 host tests + 76 shared UI tests = 375**. The
83-case online suite is included in the host count. Both analyzers and format
checks passed. Regressions cover the removed strip, header-toggle focus and
48px target, natural panel/hand adjacency, desktop rows without action scrolling,
14px user-scaled text, actual 200%/RTL dropdowns, and pending draft/controller/
caret/focus preservation through 599↔600 and 1049↔1050 logical boundaries.

Browser review of the final build covered 1449×948 desktop, 768×1024 tablet,
319×948 phone, 1280×600 short desktop, and the phone's actual 200%/RTL settings.
The desktop now shows the full current 17-card hand and basic action controls
without vertical scrolling. Tablet rows reflow; narrow/short views and enlarged
text retain bounded scrolling without displacing the hand or clipping wrapped
controls. Opened dropdowns honor enlarged RTL text and Escape dismissal; keyboard
selection persisted through resizing. Guidance expands and collapses, and the
header opens/closes the workspace while retaining the draft. Review changed no
authoritative game state, submitted no command/pass, and restored the original
Hearts selection, normal text size/direction and browser viewport. Console
warnings/errors were empty; the existing match preview remains open.

This follow-up again used only the in-app Chromium browser; Firefox/WebKit,
native and hardware-touch acceptance were not rerun. Pointer interactions and
widget touch-target tests are separate from physical-device certification.
Screenshots were reviewed inline and no captures were written to the repository.

Final served/local bundle SHA-256:
`2973f9847adc8dba5ae8681349184dc0b7c1114cb07fc0518f7f122c4cac9f30`.
Online source:
`aab123ad2a0da3d2e6ed858f04c869fa71d6dc00e0bbe02531a8701cbcdaf366`;
shared adaptive source:
`70622d66104df8b789eea949fd9614b462b2d5a755602e21891673b9b63cbe93`.
The service rebuild retained persistent volumes.

Final logs under `/tmp/agent-runs/`:

| Check | Log filename | Result |
|---|---|---|
| Redundant strip | `action-strip-red-corrected--20261003T073544Z-3479590.log` | Expected RED |
| Desktop row | `action-density-red--20261003T073522Z-3478748.log` | Expected RED |
| Artwork-scaled title | `action-density-title-red--20261003T074019Z-3489629.log` | Expected RED: 16.88px rather than 14px |
| Wasted panel height | `action-height-red--20261003T074316Z-3497372.log` | Expected RED |
| Final host suite | `action-density-host-final--20261003T074841Z-3510273.log` | 299 passed |
| Final shared UI suite | `action-density-final-ui--20261003T074533Z-3502048.log` | 76 passed |
| Host analyzer | `action-density-final-analysis--20261003T074533Z-3502085.log` | Passed |
| Shared analyzer | `action-density-final-ui-analysis--20261003T074602Z-3503735.log` | Passed |
| Build/recreate web | `action-density-final-web--20261003T074533Z-3502036.log` | Passed |
| Served build identity | `action-density-final-identity--20261003T074702Z-3506892.log` | Matched |

Captured diagnostic failures were inspected before retrying. Existing tests were
adapted to explicitly expand optional guidance and to assert the new natural
height plus exact restored table geometry. A keyboard-inspection test now closes
the action workspace before seeking the public table, since hidden public content
is properly offstage. A deliberately busy spinner required bounded test animation
pumps rather than waiting for it to settle. New geometry uses 0.01px comparison
tolerance for floating-point addition. A wrong Flutter working directory and a
tooltip-widget cast in initial test setup were corrected. No tests were skipped
or removed; all affected behavior assertions run in the passing full suite.

## Draw pile, table visibility and wide-screen spacing follow-up

Run: `visible-deck-20261003`. The owner clarified that the missing deck meant the
face-down draw pile. The supplied 1356×948 online-match screenshot also exposed
hand suit groups stretched across half-screen columns and left-anchored action
rows. Inspection found a separate visibility defect: opening actions deliberately
put the entire public table offstage.

The existing identity-neutral `ConcealedCard` and shared `back.png` provide the
pile marker at the center of the public table. Neither the backend observation nor `TableViewData` supplies a public
stock count, so this change identifies the draw-pile location without inferring
remaining supply, exposing card identities, or inventing a draw command. The
marker has one localized accessible label and no gameplay callback. Game rules,
authoritative state and bot behavior remain unchanged.

The shared board retains its four public quadrants and adjacent private hand
while decisions stay mounted below them in a bounded area. Wide hand columns
use the longest actual suit fan rather than half the window width. The online
form uses a centered 960px maximum width; status, input and command rows center
from tablet width while field contents keep their reading direction.

New regressions first reproduced absent draw-pile markup, hidden public content,
stretched hand columns and off-center command geometry. The initial integration
checks additionally caught header crowding and lost action space in short,
enlarged-text windows; these required layout repair rather than suppressing
checks. The draw-pile semantic test uses explicit `finally` cleanup, and the
centering test updates both physical and logical viewport metrics.

Final verification passed **305 host tests + 76 shared UI tests = 381**. The
147 targeted adaptive-table, approved-flow and online-screen cases are included
in the host count. Both analyzers, Dart formatting, `git diff --check`, and the
208-card artwork/shared-back/font/package-boundary validation passed. Independent
review found no remaining actionable issues. Short-window regressions enforce
usable 48px card/action targets with enlarged text; existing breakpoint,
selection, draft, focus and RTL coverage remains green.

The rebuilt human-plus-three-bot match was inspected in the in-app Chromium
browser at its normal approximately 1138×948 logical viewport. The face-down
pile is visible with actions both open and closed, the private suit groups stay
together in the center, and the action rows are centered. The action picker
opened and dismissed with Escape. Closing/reopening the workspace retained the
selected Hearts ace and Open Series/Hearts draft. No command, response pass or
turn was submitted. The normal viewport and original selection were restored;
the preview remains open. Captured browser warning/error entries were empty.

Responsive browser acceptance is **inconclusive** for this follow-up: temporary
viewport overrides at the browser's approximately 1.3 device-pixel ratio produced
a mismatch between canvas screenshots and semantic element bounds, followed by
intermittent browser-control timeouts. Resetting the override and reloading
restored normal rendering, which was checked again. The repository's bootstrap
uses the default Flutter engine loader without custom canvas/DPR handling. This
observation does not establish the cause or a product regression; no full live
phone/tablet sweep is claimed. Responsive widget results are separate evidence.
Firefox/WebKit, native and hardware-touch acceptance were not rerun. Screenshots
were reviewed inline; no captures were added to the repository.

`make web.re` rebuilt the final source and recreated Nginx while retaining
volumes. Web, backend, PostgreSQL and Redis subsequently reported healthy.
Served/local bundle SHA-256 both equal
`aafd3e998f532f86b9e22e5719a2978af7a2db4ef8003627dd855dcf79211634`.
Online source:
`38a0ca3af6fec9aa2f1336afb18a90e37afd1b1906cb547b0c7adb7d14ba342e`;
shared adaptive source:
`d0f0e87173c6ac4139da61101f40e5305fc631028bf7edf5bf3a0d3bf877e044`.

Final logs under `/tmp/agent-runs/`:

| Check | Log filename | Result |
|---|---|---|
| Missing draw pile | `draw-pile-red--20261003T092452Z-3678159.log` | Expected RED |
| Centered action rows | `centered-actions-red--20261003T092045Z-3668125.log` | Expected RED |
| Targeted UI regressions | `table-final-geometry-targeted--20261003T093749Z-3706515.log` | 147 passed |
| Final host/shared UI suites | `visible-deck-full-tests--20261003T093833Z-3708080.log` | 305 + 76 passed |
| Both analyzers | `visible-deck-final-analysis--20261003T093833Z-3708124.log` | Passed |
| Format/assets | `visible-deck-final-static--20261003T093928Z-3711452.log` | Passed |
| Build/recreate web | `visible-deck-final-web--20261003T093833Z-3708050.log` | Passed |
| Served build identity | `visible-deck-final-identity--20261003T095120Z-3732861.log` | Matched |

## Public-card readability correction

Run: `public-card-size-20261003`. The owner reported that the public cards became
tiny after the deck/table-visibility change. The normal-view screenshot from the
preceding follow-up did show this regression; its passing tests verified table
visibility, not adequate painted card dimensions with actions open.

The action-open layout capped the entire public table at 18% of the available
workspace height. Each seat then fitted its 180px reference composition into
half of that strip, shrinking both artwork and metadata. The new integrated
1025×948 regression measures transformed render bounds, reproducing a public
card only **24.42px wide** against the corrected minimum of 64px. It also requires
a complete first private card to remain visible, the deck to remain mounted,
adjacency to remain intact and action controls to remain reachable without
submitting a command. RED evidence:
`/tmp/agent-runs/public-cards-online-red--20261003T100030Z-3749808.log`.

The corrected budget reserves a complete first private card and readable public
previews before assigning the remaining action viewport. Public cards reach at
least 64px in the ordinary tablet/desktop regression cases and grow toward the
shared hand scale on taller screens. Actions retain bounded scrolling where
needed. The minimum action viewport follows the original user text scale,
independently of artwork scaling, so 200% phone controls remain tappable.

The first integration pass caught two real enlarged-phone tap regressions: an
action picker and Stay button could be clipped by a 48px action viewport. Their
unchanged tests pass after reserving 112px for actions at 200% text. The earlier
no-scroll form assertions remain in a genuinely roomy 1280×1400 scenario;
ordinary-size regressions instead verify public-card readability, the full first
private card and scroll-reachable submit controls. A phone hierarchy assertion
now requires the table to retain its exact height when the short form fits;
tablet/desktop movement, adjacency and close/restore assertions remain.

Final verification passed **307 host tests + 76 shared UI tests = 383**. Both
analyzers and formatting passed. Independent review found no actionable issues
in the final allocation. The final web build completed and the served/local
bundle hashes matched:
`7939cd630d92e17c00107a879a6921a24dfaaad73c7996e6c2a5bb692eede403`.
Adaptive source SHA-256:
`51bd31783141c5dad6afb253d87082f4a2823295a51f006e505e805cd0113b72`.

| Check | Log filename under `/tmp/agent-runs/` | Result |
|---|---|---|
| Rendered-size regression | `public-cards-online-red--20261003T100030Z-3749808.log` | Expected RED: 24.42px |
| Public/hand bounds and enlarged real taps | `public-card-control-scale-green--20261003T100557Z-3760707.log` | 7 passed |
| Final full host/shared UI suites | `public-cards-final-tests-fixed--20261003T100807Z-3766683.log` | 307 + 76 passed |
| Both analyzers/format | `public-cards-final-analysis--20261003T100641Z-3762288.log` | Passed |
| Build/recreate web | `public-cards-final-web--20261003T100641Z-3762349.log` | Passed |
| Served build identity | `public-cards-build-identity--20261003T100947Z-3770213.log` | Matched |

Live review used the existing in-app Chromium match at its normal **1009×948
logical viewport** (approximately 1.3 DPR), in the same tablet class as the
owner's 1025px screenshot, and subsequently at **1449×948**. After the rebuilt page finished loading, the public
cards and seat labels were visibly larger, the deck remained centered, and the
first private-hand row was fully visible above bounded action controls. Bot 2's
public-card details opened correctly. The owner's latest Compensation/Hearts
ace draft was restored after reload; no game command, pass or end-turn was
submitted. The preview remains open and captured warning/error logs were empty.
All four main containers reported healthy. No viewport override or repository
screenshot was created for this follow-up; the broader cross-browser/native
acceptance limits recorded above remain open.

## Card-driven play follow-up — 2026-10-03

Run: `card-driven-play-20261003`. The owner requested removal of the generic
action picker in favor of selecting and dragging cards, including combos,
special cards, attacks and defences. The current 609-line game rules, accepted
rule decisions and existing command/payload contracts were reviewed before
mapping card selections to moves. No game rules or bot policy changed.

The online workspace now begins with a neutral selection. Own current cards
can be dragged to table seats, specific public cards/formations, the hand or
the draw pile. Mouse dragging and a touch hold are supported; ordinary touch
scrolling, tap selection, keyboard selection and explicit destination buttons
remain available. A drop only prepares a draft or offers ambiguous choices;
it never sends a game command. Only confirmation submits the reviewed move.
Public-card inspection is independent of the current draft.

Contextual intents cover whole-suit opening, royal attachments, natural and
substituted formations, explicit People protection, take-back, Spade purchases,
abilities, ordinary/Baron/Bomb/Infiltrator attacks, Ace modes and ordered
responses. Effect decisions use only the exact server-supplied physical choices;
an empty effect choice remains a decision, never a response pass. Financial,
trade and victory operations remain available through records and Game options.
Intentionally discarding an unused Ace stays an explicitly named Game option.

Public combat context now includes opaque current-view handles for committed
Clubs and frozen targets. These are bound through the same privacy codec as
visible board cards, cloned at projection boundaries, and never inferred from
hidden or historical identities. Public pending-action context distinguishes
combat, purchases and abilities so response suggestions match the pending move.
The server still validates exact cards, costs, quotas and legality.

Independent review found and repaired a queued-confirmation race that could
otherwise apply a draft after the client had advanced to another game, and a
combo-inspection callback that unnecessarily replaced a prepared draft. Current
game, phase, window, decision, ownership, physical-card availability and response
category are checked again before confirmation. Protection choices are cleared
when a changed combo makes them incompatible.

The review also found that removing the universal picker exposed a missing
Kidnapper outcome entry point. After responses finish, the authority now exposes
the actor's ordered selection of eligible public cards when no concealed
outcome remains. This is an explicit `kidnapper-outcome` stage, with exact offered
choices and fresh actor/choice validation before confirmation. Concealed random
outcomes remain automatic; the client never derives or selects a hidden card.

The rebuilt live match exposed an upgrade seam that unit projection tests alone
did not cover: saved seat projections lacked the newly added response context.
Snapshot reads now refresh derived projection data from the same-version
checkpoint, without writing events or advancing play. A real PostgreSQL
regression covers a legacy cached projection. Historical journal comparison
also tolerates only absent additive context from genuine older receipts;
present altered fields must still fail verification and receipts stay immutable.

Regression evidence includes actual mouse drag and tap preparation, explicit
confirmation payloads, ambiguous formation choices, restricted/hidden cards,
owned formation identity, ordered effect decisions, incompatible protection,
stale window and same-window ownership changes. The five browser acceptance
helpers were migrated to the new real UI; helper unit tests and syntax checks
were run. Their complete cross-browser workflows were not rerun in this slice.
Live preview evidence below is limited to the existing one-human/three-bot
Chromium match; no human move, pass or end-turn was submitted.

Final frontend verification passed **359 host tests + 76 shared UI tests = 435**,
both analyzers, and formatting for all nine affected Dart sources/tests. The
browser helper regressions passed **25 tests**, and all five migrated scripts
passed syntax checks. Asset verification passed for all 208 cards, shared back,
styles and bundled fonts. Tests and the final release preview use the workspace
Flutter 3.47.5 toolchain. The pre-existing staged lockfile was preserved after
tool-generated dependency resolution. The final web and served bundles matched:
`92ce13d92efede31248116e561a1fea5ff09904c6cb1041d9cae3bc629e222db`.

| Check | Log filename under `/tmp/agent-runs/` | Result |
|---|---|---|
| Host non-combat context regression | `card-response-host-red--20261003T111401Z-3905282.log` | Expected RED: absent pending-action caption |
| Kidnapper exposed-choice regression | `kidnapper-ui-red--20261003T111947Z-3918732.log` | Expected RED: absent offered-choice control |
| Final host/shared suites, analyzers and format | `card-driven-final-client--20261003T112322Z-3929356.log` | 357 + 76 passed; both analyzers and format passed |
| Duplicate combo choice regression | `duplicate-combo-red--20261003T113422Z-3960464.log` | Expected RED: duplicate Fate keys |
| Final host suite after combo fix | `card-driven-final-choices--20261003T113714Z-3967937.log` | 358 passed; analyzer and format passed |
| Visible Pass response regression | `pending-pass-layout-red--20261003T114242Z-3979450.log` | Expected RED: only 10px of the control visible |
| Final host suite after response row fix | `card-driven-visible-response--20261003T114459Z-3983929.log` | 359 passed; analyzer and format passed |
| Browser helper regressions | `contextual-browser-helper-final--20261003T110815Z-3895149.log` | 25 passed |
| Main browser helper syntax | `main-card-play-harness-syntax--20261003T110654Z-3892486.log` | Passed |
| Card/style/font assets | `card-play-static-fixed--20261003T110419Z-3887354.log` | Passed |
| Final release web build | `card-driven-visible-response--20261003T114459Z-3983929.log` | Passed |
| Final served build identity | `card-driven-visible-identity--20261003T114609Z-3987510.log` | Matched |
| Public context, game/wire/matchstore | `response-context-final--20261003T112501Z-3935386.log` | Passed, including Kidnapper projection→handle→intent→Apply |
| Legacy cached views with PostgreSQL and race detector | `snapshot-context-shapes--20261003T113146Z-3951134.log` | Three independent old shapes and existing readiness suites passed |
| Historical journal compatibility | `historical-context-final--20261003T113402Z-3959613.log` | Focused race tests passed; altered/partial/duplicate fields rejected |
| Full Go suites and affected race checks | `card-driven-final-go--20261003T113247Z-3954083.log` | Normal suites and game/wire/transport race checks passed; matchstore recovery below |
| Final matchstore race suite and full Go vet | `card-driven-go-recovery--20261003T113535Z-3963084.log` | Passed after the concurrently developed duplicate-field regression was fixed |
| Final backend build/recreate | `card-driven-release-backend--20261003T113542Z-3963951.log` | Passed |

Live input review used the normal **1035×948 logical viewport**, approximately
1.3 DPR. It verified neutral Ace selection, public-card inspection and dismissal,
an actual Ace drag to the table with no applicable response, and clearing the
local selection without advancing the match. The rebuilt UI rejected that
inapplicable drop without preparing or sending a command. Public previews,
the centered draw pile and the complete first private-hand row remained visible;
the generic action selector was absent. Browser warning/error capture was empty.
The valid purchase-drag/explicit-confirm path and ordered combat/ability payloads
were exercised by the automated widget/contract tests, not by playing the
owner's waiting response in the live match.

After the backend compatibility repair, Refresh connection updated the original
match to **Seat 2 · Open Series** and **No card response is available. Pass to
continue.** The round, waiting actor and hand stayed unchanged. Selecting an
inapplicable Ace no longer offered unrelated combat defences or an implicit
discard. Same-kind combo choices also have distinct friendly ordinals and
stable internal keys; their callbacks retain the corresponding formation IDs.

Final visual inspection also caught the required Pass response partly below
the action viewport at 1035×948 after adding the pending-action caption. Neutral
Game options now shares the command row with response/selection controls; a
rendered-bounds regression requires the entire Pass control to remain visible
without scrolling at that size while preserving the public table and hand.
The final rebuilt preview visually confirmed the complete Pass button above
the footer, with Game options and Clear selection on the same row. The existing
human/three-bot match remains open at the unchanged waiting response; captured
warning/error logs were empty and all four service containers were healthy.
