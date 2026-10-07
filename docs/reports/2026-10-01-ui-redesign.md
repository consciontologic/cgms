# CGMS adaptive UI refinement — 2026-10-01

> Archive removal update (2026-10-02): the owner requested removal of every
> file and subdirectory in `client/screenshots/` except `cgmsart-phone.png`,
> including raw reports, manifests, logs and copied harnesses. This document
> remains a historical account of the recorded results; its capture-time counts
> do not imply that those raw files remain available. Removed archive links are
> plain text below. See the [retention report](2026-10-02-screenshot-retention.md)
> and its per-file inventory for the disposition.

This report covers a Design Partner redesign of the existing Flutter host in
`client/` and reusable package in `client/packages/cgms_ui/`. The starting tree
was clean at `07752481961540ab308aa06c0089b630361ca2d6`. No frontend migration,
backend rule change, dependency, artwork replacement or native release work was
part of this slice. `client/offline_host/` and `client/offline_probe/` are preserved.

## Direction and evidence

The [mockups](../design/mockup/README.md) supply the screen organization,
information hierarchy, public territories, private hand and explicit action
review direction. [cgmsart-decks](../design/cgmsart-decks/README.md) supplies the
card faces and illustrations; its artwork supersedes the paintings in the static
mockups, not their UX. Both references were inspected before implementation.
Current [rules](../project/GAME_RULES.md) and the [adaptive baseline](../design/UI_BASELINE.md)
govern incompatible static arrangements. No existing `.design/brief.md` or
`.design/*-report.md` was present.

The retained visual language is painterly and dark with opaque navy reading
surfaces, warm action controls, cyan context and lime keyboard focus. The accepted
palette, Atkinson/Barlow type, card faces, shared back and asset provenance remain.
The implementation uses Flutter/Dart and existing widgets without a new dependency.

Before evidence records the original sparse entry, undifferentiated setup,
phone hand list and flat public cards. The main improvements are:

- Entry and lobby: accepted card art introduces the game, while create/join tasks
  have separate panels, an explicit seat roster and owner/member waiting states.
  Account allowance details remain available below the primary room task and in
  the existing toolbar dialog. An ordinary cookie-free entry has no error banner;
  explicit failed authentication and expired deep links retain feedback.
- Gameplay: provided formation members and remaining suit cards form public
  groups, with visible physical-copy labels and individual inspection. The private
  hand is an illustrated gallery. Turn, round, phase and pending ownership remain
  visible above the workspace. Score, cash, debt and match standing stay distinct.
- Actions: the existing authority-validated command catalogue has named families.
  Fields, exact selected-action calculations and cost/timing guidance precede
  submission. Actionable records precede broader quota context, which remains available
  without interrupting the primary action sequence. Existing trade, loan, decision, settlement and result controls
  retain their rule behavior and receive the shared visual system.
- Interaction: complete panes survive breakpoint changes, preserving selection,
  raw text and form state. Navigation never submits or passes. Card inspection
  supports the I key and long press separately from selection. Inline inspection
  focuses the visible pane; a host modal retains its own focus scope.
- Shared components: `CgmsPanel` is an opaque Material reading surface with a
  directional accent. Named spacing, shape, divider and error tokens join the
  retained semantic palette. Menus, chips, dialogs, navigation, bottom sheets,
  notices and legacy feature surfaces use consistent controls and reading surfaces.

Phone uses a dedicated table/hand/decisions surface and selected bottom navigation;
tablet retains the public table beside a hand or decision companion; desktop
retains the public table, hand and decisions with keyboard navigation to each
workspace. Breakpoints remain below 600, 600–1049 and 1050+ CSS pixels. Headers
scroll inside short workspaces; large text can use additional local scrolling.

## Verification

Before `main.dart.js` SHA-256:
`4958712221bc8450841ef714142d1d0d531bd3fd0f6b705bcdb375e1b6bbd640`.
Final `main.dart.js` SHA-256:
`35eb57ead54f43912d98b33834f31da81cbfb0045797eb9e1283890e416f0345`.
Served `index.html`, bootstrap and application bytes were compared with the local
release output. A separate local Nginx instance served the exact production CSP;
the existing workstation services were preserved.

Flutter 3.47.5 from the project-local toolchain was used. Final source checks:

| Check | Result | Captured log |
|---|---|---|
| Host `flutter test` | 178 passed | `/tmp/agent-runs/host-record-order-final--20261001T090757Z-2353151.log` |
| Package `flutter test` | 53 passed | `/tmp/agent-runs/design-ending-package-final--20261001T090305Z-2339524.log` |
| Dart analysis: host/package lib and tests | No issues | `/tmp/agent-runs/host-record-order-analyze--20261001T090757Z-2353247.log` |
| Dart format check | 44 files, no changes | `/tmp/agent-runs/host-record-order-format-final--20261001T090757Z-2353219.log` |
| Asset verification | 52 art faces, back/fonts/import boundary | `/tmp/agent-runs/design-assets-diagnostics--20261001T082305Z-2240693.log` |
| Combined bootstrap, browser and reconnect diagnostics | 9 passed total, 0 skipped | `/tmp/agent-runs/design-final-harness-node-tests--20261001T095524Z-2484902.log` |
| `flutter build web --no-web-resources-cdn` | Passed | `/tmp/agent-runs/design-final-build--20261001T090841Z-2355255.log` |

The now-removed archived check logs
recorded these results alongside deliberate failing regressions and intermediate
checks. The final logs are `host-test.log`, `package-test.log`, `analysis.log`,
`format.log`, `assets-and-diagnostics.log`, `browser-tools.log` and `build.log`;
other check logs provide history rather than additional acceptance credit.
The scoped Git attribute preserves captured terminal log bytes, including
progress output and trailing spaces.

New regressions cover duplicate session requests, raw invalid draft retention,
action-family command preservation, invalid-number feedback, cost/control order,
expected anonymous entry versus failed login, complete pane retention across
breakpoints, public group copy semantics, inline versus modal inspection focus,
desktop workspace navigation, collapsed allowance keyboard exclusion with retained
state, inspection without accidental selection, absent/present settlement declarers,
and actionable offer priority over general quota reference text.
Red runs demonstrated the changed behavior before each repair; assertions were
retained. Cross-review checked host command/privacy paths and shared/adaptive
state retention independently.

Real checks found and repaired short-height heading overflow, a 2-pixel large-text
tablet overflow, a tablet navigation target below 48 pixels, missing Material for
panel ink, raw games text lost at a breakpoint, and inline inspector focus left
in a hidden pane. A final focus audit also found that a maintained collapsed
allowance body retained keyboard targets; it now excludes focus using the actual
expansion controller, and an explicit Tab traversal verifies both states.
Text contrast tests cover primary, muted, suit, action, context
and error foregrounds on the retained reading surfaces at 4.5:1 or greater.
Rendered screenshot review then found a black inspector title: assigning a custom
dialog text style had dropped the foreground color. A `RenderParagraph` regression
failed before applying colors after custom type styles and passed after repair;
the final build and browser evidence use that correction. Inspector screenshots
wait for the dialog transition so they depict the settled opaque surface. A final
settlement review also caught a missing declarer rendered as `seat null`; the
localized summary now omits the absent field and retains the supplied seat when
present, with both branches covered by a red/green regression. Focused trade
capture then showed actionable records below the full quota list; records now
precede that reference context. A regression verifies the ordering while
preserving selected-action review and the exact authorized offer command.

Source/SDK review verified that the browser reduced-motion preference reaches
Flutter accessibility flags. Standard Material controllers shorten their duration
to 5%; adaptive pane changes are immediate. This is not a claim that every
framework animation is removed. Online text-scale overrides preserve the platform
accessibility flags.

Independent manual inspection in the Codex in-app browser at 805×948 covered the
room setup/join composition, settings menu, allowance dialog, Tab into its
controls, and Escape restoring focus to the originating toolbar button. Browser
error/warning diagnostics for that inspection were empty. This was an
accessibility-tree and keyboard inspection, not operation of assistive technology.

Initial browser runs exposed obsolete harness assumptions about heading/group
roles, the old action ordering, and direct DOM focus on offscreen Flutter controls.
The runner now traverses actual keyboard focus, waits for the framework's focus
frame, and verifies the named selection. Long-draft checks refocus Flutter's
recreated editing input and continue typing before comparing its value. These
repairs preserve the command, state, ownership, exact-value and privacy assertions.
Failed and superseded browser attempts are retained separately from final passing
evidence. Invitation-bearing pages are omitted from failure screenshots and known
invitation values are redacted from failure diagnostics.
One final Chromium batch was mistakenly launched concurrently and hit real
admission limits. Those 429 failures remain under `failed/parallel-admission`;
the remaining authority runs were restarted against an isolated authority and
executed serially without changing limits or assertions.
One subsequent tablet process ended with exit 143 (SIGTERM) before its report or exit
breadcrumb was written. The interruption metadata remains in
`failed/interrupted-143-tablet`; the external signal's source was not established.
It receives no acceptance credit. Remaining journeys use individual captured
processes rather than a multi-command orchestration shell.

A Firefox joining probe reproduced a clipped invitation-input semantic rectangle:
DOM auto-scroll followed by a pointer click activated the visible Create room
control instead of the offscreen field. Real Tab traversal exposed the intended
fields, retained their exact values and produced no POST. The runner now uses
that keyboard path in Firefox and checks zero mutations and unchanged route
through joining-draft entry. The full Firefox phone journey then passed. This
changed the runner, not the frozen application: accepted Chromium/CSP reports
retain runner `c7b04cb568995454d5d54d8495ec2d4fa921fa4f8845abbd59961d9892759107`;
Firefox class/matrix reports record
`b84af592cf321340abbdaaca49db6f573a1c6d23f7c58010e39edf23bd6f893c`.
The finish probe then found a Firefox gallery forward-Tab cycle inside the
scrolling body. Actual Shift+Tab reaches Display settings and every header control;
the finish check uses that reverse path and retains its exact focus, Enter,
settings change, Escape and focus-return assertions. The forward-only failure
remains outside acceptance totals. Source review found no newly introduced Tab
handler or nested scope; the unchanged gallery uses Flutter's default traversal.
Without running the original build for this probe, the limitation cannot be
dated confidently. The final runner records
`4fcf983ef34d0c314126644a83eb05583258fed7ffa0ad6c6c28c858d4aed5b5`.
Per-report hashes preserve these distinctions; no single-runner claim is made.
The economy runner needed the same exact heading-role readiness selector after
the new header composition made its old text selector ambiguous. Its failed
zero-case attempt is retained; the eight economy assertions were unchanged.
Final economy reports use runner
`a168f03200bf1e7620f7796bf1caf31e0298e9dd72d33b92c5a7abb8eb7faeab`.

The separate [reconnect checker](../../client/tool/reconnect_browser_acceptance.cjs)
passed two scenarios in each of Chromium, Firefox and WebKit on the final build:
an opening purchase draft and a pending Justice decision. It combines browser
HTTP offline emulation with explicit closure of a transparent connection to the
real backend WebSocket. The failed application GET, paused mutation, explicit
keyboard Refresh, restored HTTP 200 and a new WebSocket route to the real backend
are observed.
The original physical card's selected semantics, quantity, game/turn/decision
scope and authority version survive; the page emits zero command POSTs.
Six passing cases and 18 captures
use checker SHA-256 `8565dee673450a2c4edc44e8eb9e54fcdf873c81b240f80e63ed4baf11fbd4d2`.
Expected network diagnostics require the actual failed GET, exact URL, captured
offline phase and exact browser message, and can be consumed only once. Page
exceptions and unrelated failures remain errors. Failed preliminary checker
attempts are retained separately. This verifies explicit browser Refresh recovery
at 390×844 in authorized fixtures; it does not prove automatic reconnect, an OS
network partition, a naturally dealt whole match, native devices or actual AT use.

The final manifest
indexes **25 passing reports and 444 passing case executions**, all against the
final application hash above, with zero unexpected browser errors. These are
repeated checks across engines and viewports, not 444 different product journeys.
Playwright 1.63.0 ran Chromium 153.0.8010.12, Firefox 155.0 and WebKit 26.6 against
the real local Go/Postgres authority. WebKit used the pinned Playwright container;
no native Safari acceptance is implied.

| Executed coverage | Reports | Passing cases |
|---|---:|---:|
| Phone, tablet and desktop journeys in all three engines | 9 | 264 |
| Adaptive viewport/transition matrix in all three engines | 3 | 99 |
| 320-pixel, long-draft, motion and dialog finish checks in all three engines | 3 | 18 |
| Exact production CSP served locally in all three engines | 3 | 3 |
| Additional Chromium three-seat journey | 1 | 30 |
| Allowance, recovery and private economy state in all three engines | 3 | 24 |
| Explicit Refresh reconnect with retained drafts in all three engines | 3 | 6 |
| **Total** | **25** | **444** |

Chromium contributes 170 cases; Firefox and WebKit contribute 137 each.
Clipboard-denial recovery is checked only in Chromium, explaining the 30 versus
29 cases in each full class journey. The baseline matrix executes 390×844,
360×800, 599×900, 600×900, 768×1024, 1024×768, 1049×900, 1050×900, 1280×720,
1280×600 and 1440×900, with breakpoint reversals, 844×390 rotation, and 200%
application text/RTL/art-off checks in all three device classes. Finish checks
add 320×800 and a retained 181-character mixed-direction draft.

Browser checks exercise real guest room creation, private invitation joining,
owner/member reload and dealing, then authorized fixtures for ordered play,
purchases, attacks, compensation, Confinement, Justice, Coup, trades, loans,
settlement and final ranks. Lost replies retain the original operation through
reload. Economy checks include 200% text/RTL dialogs at phone/tablet/desktop
sizes, one purchase after lost-response recovery, private balances and receipts,
and streamed allowance transitions.

The executed accessibility checks use browser semantics, real keyboard traversal
and activation, card inspection, Escape/Close focus restoration, browser history,
primary touch-target measurements of at least 48 pixels, private-offer semantics,
long drafts, RTL, text scaling and both browser motion preferences. Screenshot
review checks the rendered opaque surfaces, artwork, text and controls. Color
contrast is verified by the Flutter tests described above. These checks do not
constitute an automated or human screen-reader conformance certification.

## Before and after capture records

Capture reports were originally recorded in the now-removed design-partner-20261001 archive.
`before/` recorded the clean starting build; `after/` recorded the final passing
build; `failed/` and `superseded/` held other attempt reports outside acceptance totals.
The manifest counts 57 before captures and 350 accepted final captures, including
the supplemental `reconnect/` evidence. Another 280 failed or superseded captures
were recorded and excluded from acceptance; their images have since been removed.
The before reports contain 91 case
executions: three 30-case class journeys and one supplemental entry check.
Rule-boundary fixture states are directly comparable. Normal guest-room deals
exercise real shuffling, so their faces and starting seat can differ.

| Journey and viewport | Before run report | After run report |
|---|---|---|
| Phone 390×844 private hand after a real deal | Hand list report | Illustrated hand report |
| Room lobby | Lobby report | Seat roster and room controls report |
| Selected-action review | Decisions report | Fields and cost guidance report |
| Tablet 768×1024 public context and hand | Original tablet report | Persistent public table and hand companion report |
| Desktop 1440×900 workspaces after a real deal | Original desktop report | Public groups, hand and decisions report |

The settled 320-pixel inspector report
records the tested inspector state. Visual review at the time confirmed the
corrected readable title and opaque surface; other now-removed captures covered
private offers, loans, effect decisions, settlement, ranks, interrupted actions,
large text, RTL and art-off rendering.

## Boundaries

The command catalogue is timing-aware, but the current snapshot does not supply
an exhaustive per-action legal menu. This narrow contract gap remains explicit;
the interface does not infer concealed cards or implement a second legality engine.
Public group labels use supplied visible cards and formation labels, without
inventing totals or powers. Concealed hands and Aces remain authorized counts.

Browser semantics and keyboard checks are not operation of a screen reader.
Firefox gallery forward Tab cycles within the scrolling body after entry;
Shift+Tab reaches the header controls. The verified settings dismissal and focus
return use that reverse path, not a claim of correct forward wrap order.
Actual assistive-technology use, native Android/iOS, on-screen keyboards, device
safe areas and native lifecycle/store acceptance remain R06N. True browser zoom
was attempted in the in-app browser with Ctrl+plus twice: width 805, height 948,
device-pixel ratio 1.3 and visual-viewport scale 1 remained unchanged, so this is
recorded as unsupported/unverified; 200% application text scaling is separate.
The catalog is English ARB. Long strings and RTL layouts do not establish a full
translated-language review. Synthetic rule-boundary positions complement real
room creation/join/deal/reload, not a complete naturally dealt match.
The successful web build retains an existing warning about an unbundled
CupertinoIcons family; actual CGMS artwork, bundled text/suit fonts and inspected
web controls passed the asset and browser checks. Native icon coverage is unverified.

The sole sequenced checklist remains [ROADMAP.md](../planning/ROADMAP.md).
This refinement does not close native, production, store, bot-strength or other
broader gates.

## Follow-up — compact card proportions

The owner reported elongated cards at phone sizes. The starting tree was clean
at `90eef62`. Artwork used proportional `BoxFit.cover`; the distortion came from
the bordered component including additional copy/status and inspection rows.
A fresh 390×844 Chromium journey captured the
original hand report.
The new browser regression also reproduced a 136×210.797 selection face at
320×800 (height/width 1.55). The full bordered hand card was taller still because
its independent inspection button was inside that border.

Compact hand and public card faces now keep a 5:7 width/height ratio. Opaque live
rank/suit/copy masks remain within the illustrated face; status and inspection
sit outside its border. Stack headers retain the physical-copy label and selected
mark. Existing stacked-card logic exposes a full card whenever it has a status,
so statuses remain visible between physical copies. Noncompact inspectors,
thumbnail compatibility views, card art, commands and privacy projections retain
their existing behavior.

Regression tests first failed on the elongated component at widths 56, 100 and
144, with both stack-header modes. Final coverage also exercises 200% RTL text,
copy/rank containment, status separation, 48px inspection targets, stacked status
visibility and phone hand selection/inspection at widths 360, 390 and 599. A
separate regression caught an expanding opaque copy mask before acceptance; the
mask now wraps its rendered label plus padding without covering the illustration.

Final source checks on Flutter 3.47.5: **212 host tests and 65 package tests pass**,
both analyzers pass, formatting checks 44 files unchanged, all 208 artwork hashes
and import boundaries pass, and nine browser/bootstrap diagnostic tests pass.
The release build succeeds with the existing CupertinoIcons warning described
above. Captured logs
retain the red proportion/mask regressions and final checks.

The final application SHA-256 is
`438e367f51353f3493a795a0352267addd716ef02538a47e548cbfa7b52f387f`.
QA-served index/bootstrap/application bytes matched local build output; the
ordinary `localhost:8080` developer server also served the new application hash.
The focused `CGMS_CARD_SIZES_ONLY=1` browser mode checks complete visible face
rectangles, independent inspection, retained selection and zero command POSTs.
Raw browser rectangles were recorded in the removed reports; comparisons used 0.001 CSS-pixel
precision because a transformed 48px control can report 47.999996px.
Failed and intentional red runs were archived separately under `failed/`;
that directory is now removed.

The focused Firefox journey exposed a pre-existing Inspect keyboard failure after
inline inspection, return to Hand and resize from 320 to 360 pixels. A controlled
baseline/current comparison
reproduced the same missing Tab stop with the original application and the first
proportion revision. Pointer inspection and the card's I shortcut still worked;
neither comparison receives acceptance credit. A new widget regression reproduced
the missing focusability before repair. Navigation now settles outgoing control
focus while its pane is still visible, preserving its focusable semantics when the
retained pane returns. All ten adaptive tests, including existing modal focus and
draft-preservation checks, pass; independent review found no actionable issues.

A later Firefox landscape measurement exposed a harness boundary assumption:
Inspect occupied its full 48×48 pixels at y=342..390 in a pane ending at y=390.
The control was visible, but strict interior containment rejected its edge contact.
Inspection measurements now permit equality at the pane edge while retaining the
48px minimum and trial hit-test; face measurements keep their strict interior
check and 5:7 assertion. The failed run and raw geometry are preserved, and the
measurement adjustment received independent review. Reports retain each run's
exact harness hash rather than claiming identical harness bytes across runs.

The requested phone proportion checks pass at widths 320, 360, 390 and 599 in
Chromium, Firefox and WebKit: **20 phone cases per engine**, including normal,
200% application text and RTL where specified above. All use the final application
hash. The wider exploratory matrix is **48/48 Chromium, 48/48 WebKit, and 43/48
Firefox**. Firefox's report remains `failed`, not accepted as a complete matrix.
Nine public-table cases needed genuine Shift+Tab after forward traversal wrapped,
including phone RTL widths 390 and 599; this proves keyboard reachability rather
than correct forward Tab order. The tenth detour at public-table RTL 768×1024 failed in both directions. Pointer
inspection worked there, but the subsequent card-I diagnostic could not reach its
card, so that input receives no pass claim.

A bounded direct RTL baseline/final comparison
passed the same single 768×1024 case on both original and final application bytes.
It omitted the preceding resize/settings history, so the long-sequence failure's
origin remains **inconclusive**, not proven pre-existing or fixed. This is an open
broader keyboard-navigation finding; the phone visual scope and source checks
pass, but this follow-up does not certify all Firefox tablet keyboard traversal.
The failed long-run report, textual diagnostics, and superseded run summaries remain in the
artifact bundle and are excluded from complete-matrix acceptance.

The final 390×844 Chromium journey passes all **30 cases**, including normal
room creation/join/deal/reload, action and response flows, private offers, recovery,
selection and independent inspection. The real-deal
before hand report
and final hand report
record the compared real-deal runs; their then-captured screenshots showed the
changed silhouette. Different shuffles explain their different faces and hand
counts. The final hand and public-table screenshots were visually inspected then.

The evidence manifest
records hashes, complete runs, partial exploration, baselines and superseded work.
The task's disposable QA fixture shut down gracefully with exit 0 and port 8081
closed; the ordinary developer server on 8080 still serves the final application.
The unresolved Firefox failure breadcrumb is retained with its report reference.


## Follow-up: complete card paintings and deliberate magnification

The owner's next revision supersedes the preceding 5:7/masked-card presentation:
show full cgmsart paintings, remove visible Copy labels, and magnify deliberately
with double-tap or a small hollow corner affordance where gestures conflict.
The earlier evidence above remains historical and is not evidence of this revision.

The 208 bundled paintings share 52 native suit/rank dimensions across four styles;
they are unlabelled paintings with height/width ratios about 1.19–1.35. The image
box now follows those exact dimensions with `BoxFit.contain`. Rank/suit and the
selection mark sit above the painting; meaningful usage/availability text sits
outside it. Raw hand/series zone captions and visible copy numbers are removed,
including inspector titles and selection summaries. Authorized semantics retain
physical-copy identity. Full public cards no longer overlap or use clipped rank
tabs; groups reflow and scroll, preserving every physical card's reachability.
The maintained baseline and package/skill references record this owner revision.

Single tap selects only where the host permits. Double-tap, keyboard I and the
small hollow bottom-right magnifier inspect without changing a selection or
submitting a command. The corner's transparent 48px target is needed because
screen-reader double-tap activates the selected control. Inspectable paintings
are at least 100px wide, keeping the center outside that target. Rejected card
selection and long press no longer magnify. Online inspection now enlarges the
current authorized painting; capture/style changes, later concealment and new
game transitions continue to reconcile or dismiss it safely.

Fresh Chromium touch input exposed a real web-specific issue after widget tests
passed: Flutter 3.47.5's `ClickDebouncer` converts each click on an accessible DOM
button into a semantic tap and drops its raw pointer stream. Two touch clicks
87.8ms apart toggled selection twice instead of reaching `InkWell.onDoubleTap`.
The failed run and input diagnostic are retained without acceptance credit.
A regression dispatching `SemanticsAction.tap` reproduced premature selection.
The card now uses the same Flutter double-tap window for semantic taps: a single
activation selects after that window, a second cancels selection and magnifies.
Pending selection cancels on every inspection route, identity replacement and
disposal; callbacks consult the current widget. Raw gestures keep Flutter's
existing recognizer. A separate regression repaired native IconButton focus
semantics after the initial wrapper hid them.

Final source checks: **213 app tests, 75 package tests and 9 browser-support tests
pass**. Both analyzers pass, 44 Dart files require no formatting changes, and all
208 asset hashes/native dimensions plus the shared back/import boundary pass.
Independent review found no actionable correctness or privacy findings, with
three semantic-event regressions independently rerun. The release build uses
`--release --no-web-resources-cdn` and retains the existing CupertinoIcons warning.
Source checks and intentional red regressions were archived in `checks/`;
that directory is now removed.

Final application SHA-256:
`1ee104ac502fef5d001c05b419af1dd8c0025d23834c2c0df34785e622a037f7`.
The QA server on 8081 and ordinary development server on 8080 served identical
index/bootstrap/application bytes. All 269 files listed by the previous evidence
manifest still match their recorded hashes.

The focused browser matrix passes **12/12 each in Chromium, Firefox and WebKit**
(36 accepted cases): full hand/public card exposure at 320×800, 390×844,
768×1024 and 1440×900, plus phone 200% text and RTL. Each uses actual touchscreen
center selection, two taps 80ms apart for magnification, the 48px corner action,
retained drafts and zero authority-command POSTs. Normal 390px cases also use I.
All three reports contain zero unexpected browser errors and the final app hash.
DOM geometry measures the card including its header, not CanvasKit's decorative
image; exact image fit is established by asset/widget checks plus visual review.
Nested containment allows one CSS layout unit (1/64px), because the first failed
geometry run found a 0.0129px rounding difference between a transformed child and
its parent. The 48px minimum remains measured at 0.001px precision. That failed
run is preserved without credit.

The default approved phone demo also passes actual double-tap inspection with
no authority POSTs, with separate board/hand/inspector screenshots. Visual review
of those captures, Chromium's normal phone hand/inspector, and WebKit's RTL hand
and 200% public table confirms complete paintings, no visible Copy captions and
the small hollow corner affordance. The default demo's separate formation-details
button can wrap the final letters of “Underground” at 390px; it remains usable
and is recorded as a nonblocking typography follow-up outside the card-art scope.

These fresh-context checks do not resolve or replace the earlier long-sequence
Firefox tablet RTL keyboard finding, and do not establish native-device or actual
screen-reader certification. The old unresolved finding remains in the checkpoint
and failure breadcrumb rather than being marked fixed by unrelated green tests.

The final Chromium 390×844 quick journey passes all **30 cases**, including guest
room/deal, opening and action switching, responses, private offers, recovery,
settlement and standings. The report has zero unexpected browser errors; six
injected recovery faults and four initial signed-out 401 challenges are explicitly
classified. Together with the 36 focused cases and one default-demo case, this
revision has **67 accepted browser checks** on the same final application and
harness hashes. The new evidence manifest
separates accepted reports from failed diagnostic runs and records source/artifact
hashes. Both disposable QA batches shut down gracefully with exit 0; port 8081 is
closed, and the ordinary server on 8080 continues to serve the verified build.
