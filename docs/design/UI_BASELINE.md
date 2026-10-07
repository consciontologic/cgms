# CGMS UI baseline — shared vertical cgmsart

**Owner interaction revision, 2026-10-04; retained layout revision, 2026-10-02.** The owner selected
[`client/screenshots/cgmsart-phone.png`](../../client/screenshots/cgmsart-phone.png)
as the current layout reference for phone, tablet and desktop. This supersedes
the earlier v2 requirement for separate phone pages, tablet split panes and
desktop workspaces. All sizes retain the reference's vertical arrangement;
larger widths enlarge the cards and items. The earlier browser results remain
historical evidence for their respective builds, not the current composition.
Flutter web remains the development platform; native acceptance is separate.

## Authority and retained deliverables

Use this order when references differ:

1. [GAME_RULES.md](../project/GAME_RULES.md) governs legality, timing, visibility
   and financial behavior. Visual examples cannot amend the rules.
2. This baseline records the owner's current layout/interaction direction.
3. The [phone reference](../../client/screenshots/cgmsart-phone.png) is the
   primary current board layout: compact round/turn header, two-column public
   seat overview with card stacks and adjacent private hand. The 2026-10-04
   revision removes the permanent action region and retains compact non-card
   utilities, as described below.
   The [20 mockups](mockup/README.md) remain references for other flows.
   [cgmsart-decks](cgmsart-decks/README.md) remains the card/illustration authority.
4. The [Flutter implementation](../../client/README.md), especially
   [CgmsAdaptiveTable](../../client/packages/cgms_ui/lib/src/adaptive_table.dart),
   [CgmsActionBoard](../../client/packages/cgms_ui/lib/src/action_board.dart),
   [CgmsTheme](../../client/packages/cgms_ui/lib/src/theme.dart) and
   [approved flows](../../client/lib/approved_flow.dart), must implement this
   shared arrangement while preserving authoritative state and privacy.
5. The [cgms-art-direction skill](../../.agents/skills/cgms-art-direction/SKILL.md)
   explains applying this baseline and its semantic/accessibility constraints.

The acceptance preview was served at `http://127.0.0.1:7359/#/table`; that local
address is a review convenience, not a deployed service or permanent artifact.
The checked-in app and [run instructions](../../client/README.md#run) reproduce it.

## Shared vertical composition and acceptance contract

Use available width in Flutter logical pixels. Auto still classifies phone at
**width <600**, tablet at **600–1049**, desktop at **>=1050** for sizing and review;
the class never switches the board into different panes or hides the hand behind
navigation. Use the board's actual available height as well as width, including
space already consumed by the online header and notices.

At every width keep the public overview and private hand together in the viewport:

1. Round/turn and relevant pending/connection status.
2. One full-width table split into four equal public-seat quadrants, with its
   left/right edges aligned to the hand. Larger screens use wider rectangles;
   the owner's follow-up supersedes the earlier square-only constraint. Keep the
   2×2 arrangement even below 330 logical pixels; three-seat games leave the
   fourth quadrant empty. Fit decorative piles within their cell instead of
   allowing large holdings to make a row taller.
3. Private hand immediately adjacent to the public table, grouped by suit in
   horizontal overlapping stacks, with uniform
   card frames. This owner follow-up supersedes the reference's six-column hand.
4. Temporary card-specific choices or confirmation overlay the board only when
   necessary. No action panel, reserved composer space, toggle, tray or sidebar.
5. Compact non-card utilities: End Turn, explicit Pass, required decisions,
   records, financial workflows, ordinary declaration and match progression.

The public table uses the same width and card-size scale as the hand, including
live phone→821px→desktop resizing. Decorative public cards retain native artwork
proportions and shrink only when their actual quadrant cannot fit that size.
Dense piles cap their total fan depth so card counts cannot shrink previews
indefinitely. Once cards reach the shared hand scale, additional width gives the
quadrants more room. The table's height remains bounded independently of width;
do not enlarge it until it pushes the hand below the page. The
initial board must show the hand and a usable card target without page scrolling,
including the actual 319×948 preview. Short views, long hands, detailed actions
and 200% text may scroll within their own bounded areas; the public overview and
hand remain adjacent. Do not add a navigation rail, separate hand page or
horizontal workspace split. Primary touch controls remain at least 48 logical
pixels. Public decorative card-stack previews may shrink and overlap; each
quadrant's accessible control opens full details and supplied physical
cards for exact selection and inspection. Never make a narrow exposed rank tab
the only way to select a physical card. Private hand cards remain individually
selectable and inspectable. Hand frames use one 4:5 ratio; complete paintings fit
inside without cropping or stretching. Suit groups share two columns where space
allows and scroll horizontally when needed; enlarged text may use one column.
Expose at least 48 logical pixels per overlapping card for direct interaction,
preserve physical-copy selection, and scroll keyboard focus into view in both
directions. Inspectors show full artwork at its native ratio.

Top and bottom bars use the available width deliberately: round/turn context
stays at the leading side, layout/menu tools at the trailing side, and the bottom
dock separates utility actions from primary game actions. At ordinary text size,
keep the header in one centered horizontal row down to the 319px preview: use
smaller visible text/icons and compact selector spacing, retaining full round/turn
and Auto labels. Controls keep 48px targets; do not move the tool group below
the status group at a fixed width threshold. Enlarged accessibility text can wrap
within the status area without clipping or reducing the requested text scale.
The bottom dock can still reflow without forcing page scrolling. Bar
insets, gaps, icon sizes and control heights respond to available width and
height; opaque surfaces and subtle borders distinguish them from scrolling hand
content. Keep complete labels, visible focus and at least 48px targets at 200%
text scaling; preserve mounted controls while groups reflow. Paired primary
actions share the tallest natural label height; accessible hitboxes must follow
text-size changes made through settings menus.

State and focus stay in the mounted board across resizing and internal scrolling: preserve selected
physical IDs, raw form drafts/caret, game/decision IDs and pending operations.
Scrolling, resizing and opening or dismissing a group cannot submit commands,
pass a turn, change eligibility or expose concealed card identities.

| Class | Browser acceptance sizes | Shared arrangement |
|---|---|---|
| Phone | 319×948, 320×568, 390×844, 360×800, 599×900 | Four equal public quadrants with adjacent visible hand |
| Tablet | 768×1024, 1024×768, 600×900, 1049×900 | Same arrangement with height-aware item sizing |
| Desktop | 1440×900, 1280×720, 1050×900 | Same arrangement with height-aware item sizing |

Phone/tablet web remain the intended eventual native appearance. Actual native
lifecycle, platform input and device accessibility are separate R06N gates.

## Browser testing and evidence

P04 requires actual Flutter web browser tests, not screenshots or widget tests
alone. Run the table's viewports in Auto, plus live transitions 599↔600 and
1049↔1050 at height 900, 390×844↔844×390 and 768×1024↔1024×768.
Width determines item sizing even after rotation. Also exercise a 1280×600 desktop window
and 200% text scaling in each class, RTL, keyboard-only and touch input, and
browser screen-reader semantics. Record browser/version, viewport, input mode,
scenario and pass/fail; use Chromium, Firefox and WebKit coverage with any missing
coverage explicitly open rather than inferred from another browser.

At each representative class, execute setup/deal, opening and physical-card
inspection, attack/ordered response, private trade and acceptance, purchase,
loan/alliance, Confinement, immediate Coup, settlement/standings and reconnect
against shared authorized fixtures and the integrated authority. Preserve all
F01–F10/N01–N07/T01–T03 P04 obligations. Check pending/disabled/error states,
nonblocking offers, exact scores/cash/debt and art-off parity. Browser back/forward
and class transitions must retain game/decision IDs, draft selections, scroll or
return context and recoverable focus without duplicate commands, implicit passes,
lost cards or private data in DOM/semantics/tooltips/URLs/events.

Acceptance requires readable, unclipped text and controls, deliberate scrolling,
accessible per-card inspection, touch targets at least 48 logical pixels for
primary interactive controls, visible keyboard focus and access to legal actions
without hover. Eligible Coup stays immediately accessible even with a sheet open.
Large holdings and text sizes need not fit one screen. Test both normal and dense
states; compare approved images and executed journeys independently by class.

Existing v1 evidence at 390×844, 768×1024 and 1440×900 remains historical proof of
its shared board, including complete initial/opened-Hearts hands. It is not v2
acceptance. Native packaging, device lifecycle, performance, accessibility and
platform behavior remain separate mandatory R06N gates; browser testing cannot
prove them. ROADMAP is the sole sequenced completion checklist.

## Frozen theme and resources

[theme.dart](../../client/packages/cgms_ui/lib/src/theme.dart) is the executable token
source. Reuse `CgmsTheme.dark()` and `CgmsColors`; do not fork local palettes.

| Role | Value |
|---|---|
| Canvas / surface / raised | `#101322` / `#1D2335` / `#2B344B` |
| Primary text / secondary text | `#F5F1EA` / `#BAC4D6` |
| Primary action / secondary accent / focus | `#FF795E` / `#69D8E4` / `#DFFF7A` |
| Meaningful border | `#8595B0` |
| Hearts / Diamonds / Clubs / Spades | `#F5A4BE` / `#FFD275` / `#83DCD3` / `#B6B2FF` |
| Body / display fonts | Atkinson Hyperlegible Next / Barlow Condensed |

Keep bundled fonts, package asset paths, vector suit marks and live rank/suit
labels. Decision 0016 assigns the four existing styles—First Light, Rain Glaze,
Ember Glaze and Drypoint—distinctly across players for each entire match. All
authorized viewers use the same public assignment; visible cards use their
current controlling player's style, including after capture. New matches receive
fresh assignments; rebuilds, reconnects and games within a match do not reroll.
Cards have one shared identity-neutral concealed back regardless of style.
First Light remains the standalone-preview and unassigned-card fallback.
The full [cgmsart deck library](cgmsart-decks/README.md), source art,
reproduction tools, manifests and license notices remain retained resources.
Styles are automatic, with no purchase, currency unlock or entitlement gate.
Image-free rendering remains a failure/accessibility fallback, not a replacement
for the approved illustrated theme. Approval does not change asset rights.

Owner typography follow-up, 2026-10-03: the online lobby uses centered introductory
copy and a clear title/section hierarchy. Form labels and entered values retain
their reading direction. Online setup and gameplay dropdowns use regular 15px
body text instead of bold heading text, inherit accessibility scaling and allow
option rows to grow beyond their 48px minimum. Do not reduce touch targets to
make the visible type smaller or change the global gameplay heading scale.
Popup options inherit the same text scale and reading direction as their field.
The former permanent action form and its 2026-10-03 width, expansion and
reserved-height requirements are superseded by the 2026-10-04 revision. The
face-down pile, four quadrants, centered compact suit groups, public-card
readability and adjacent private hand remain required. On short/enlarged views,
scroll the vertical composition and contextual content deliberately instead of
shrinking all cards. Temporary popups must fit the viewport, retain keyboard
focus across live resizing, restore focus on dismissal, and allow targeting on
the underlying board. Their parameter choices must not block immediate Coup.

## Interaction and state contract

Use the existing narrow callbacks and viewer-filtered `TableViewData` contract.
The app/controller owns legality, selection and authoritative results; widgets
do not derive rules from images, totals or hidden card inventories.

Owner card-interaction revision, 2026-10-04: physical cards and concrete board
targets are the only entry points for card actions. Single tap/Space selects or
toggles a physical card; source, target, payment and modifier roles have distinct
visual and semantic treatment. Double tap/Enter enters that card or formation's
flow without triggering two selection changes. Ambiguity uses a concise choice.
Inspection uses the visible magnifier or `I`, independently of activation, and
never exposes an unused Ace to the game. Identical copies remain independent.

Dragging a card or exact selected bundle onto an own series/formation, opponent
card/formation/seat, hand or draw pile previews the exact move, physical cards,
costs and destination before release. A deliberate release of one complete,
unambiguous legal move submits exactly one operation; the server decides final
acceptance. This owner follow-up supersedes the earlier blanket confirmation
requirement. Do not repeat the chosen destination or add a confirmation after
an already complete drop. A touch hold
starts dragging; ordinary movement remains scrolling. Tapping the same selected
cards and destination performs the same preparation. Drop validation uses the
current immutable game/context and owned physical sources; invalid/cancelled
drags, Escape, scrolling and inspection cannot submit, pass or spend anything.

Temporary choices are specific to the move: combo identity, fixed Doppelganger
binding, protection, recipient/terms/revision, exact payment, purchase quantity,
Ace mode/value, Code settings, ordered effect choices or final confirmation.
Choose physical members and targets on the board wherever possible. There is no
permanent universal action form, including behind a toggle or inside a modal.
A necessary contextual commitment names exact cards, costs and targets. Replace
generic “Choose on the board” detours with specific missing-input guidance and
highlight the corresponding board targets. Source cards are not silently
replaced when their availability, ownership or visibility changes. A royal
attachment requires an explicitly chosen supported series; activation alone
cannot supply a default destination. Deliberate unused-Ace disposal always
requires its named destructive choice and confirmation, even if no other choice
is available at that boundary.

The first series opening includes all held number cards of that suit; additions
and historical restoration retain the chosen physical members. Historical and
current openings remain distinct. Whole-formation
take-back/loans retain their exact members. Barricade retains its initiating Queen
separately from five other cards; Dexter separates its Diamond payment and target.
Ordered effects use exact offered actor/stage choices, including a required empty
choice distinct from Pass. Required choices remain pending and reopenable after
dismissal. Explicit Pass remains reachable even without a card response.

Keep the draw pile in dedicated safe space outside every public seat's text,
cards and controls, with its shared face-down artwork and at least a 48px target.
Use the same safe layout rule at all widths, including 1041×948; never hide text
or shrink public cards to compensate for an overlapping pile.

The authority draws one card at each scheduled turn start, after Confinement
expires and subject to exhaustion. Feedback follows newly confirmed draw
transitions only: ordinary cards enter the hand, Aces the concealed-Ace area;
turn-drawn Diamonds stay concealed until a legal later addition. A brief themed
movement/glow is presentation only, never a draw command. Rebuild, resize,
refresh and reconnect cannot replay it, and trades/loans are not draws. Reduced
motion uses a quiet highlight and concise privacy-safe announcement without
taking focus or delaying immediate legal Coup.

Dragging an intact exposed own formation to another player proposes a private
loan of those exact members to that recipient. It never silently becomes an
attack or trade. Ask only for missing terms; acceptance binds the exact revision
and begins responses. Successful atomic resolution alone transfers custody and
establishes the alliance. Returns do not dissolve it. Combo-assisted attacks
remain a distinct card-driven activation with their attacking cards and targets.

Trades, loans and voluntary card returns start through the cards. Financial
promises/transfers/debt/settlement, End Turn, Stay/Leave, nullification and ordinary
victory declarations use compact utilities or temporary screens. Coup is card-
driven through the real eligible Queens, including equivalent Queen cards in
open contexts and menus; confinement and pending effects do not remove its
special admission. Ordinary and Coup drafts retain the server's same-game
victory admission instead of applying blanket version rejection.

See the complete [rule-command-gesture-test matrix](../reports/2026-10-04-card-interaction-redesign.md)
for action coverage and the exact verified build/browser limits.

Move suggestions are presentation hints, not legal-action guarantees. The server
retains timing, ownership, support, quota, cost, hidden-outcome and final legality
checks. Use only current viewer-authorized identities; never infer hidden faces,
stock counts or fixed binding metadata absent from the projection. These changes
do not amend game rules or establish new browser, native or release acceptance.

- Owner revision, 2026-10-02: compact reference-sized card stacks replace the
  previous minimum-100px/no-overlap board restriction. Keep cgmsart artwork and
  live rank/suit identities. Public stacks may overlap in previews; accessible
  group controls reveal every authorized physical card. Private hand cards stay
  individually usable. Full inspectors preserve the painting's native ratio.
  Compact board cards show the reference's small physical-copy number for
  authorized cards only. Full cards and inspector titles omit that visible
  number; all modes retain physical identity in authorized semantics/model data.
  Never display opaque internal card handles or concealed identities.
- Single tap/Space selects; double tap/Enter activates; `I` or the visible
  magnifier inspects independently. Compact cards must retain an accessible inspection action
  without a hit target covering their selection area. Group selectors must have
  readable full-card controls and keyboard focus containment/return. Rejected
  selection, long press and navigation must not submit or magnify unexpectedly.
- Preserve keyboard focus, semantic card identities, readable text scaling and
  individually usable inspector controls. Dense visual rank tabs are not a
  substitute for accessible per-card controls.
- Show explicit ordered responses. A timer, navigation, disconnection or demo
  preview never silently submits a pass. In the local host, opponent simulation
  controls identify themselves as simulation.
- Distinguish a nonblocking offer from its accepted response-bearing action;
  do not display a transfer or alliance as complete before successful resolution.
- Preserve declared versus resolved costs, exact chosen purchase quantities,
  immediate eligible Coup, and board closure versus settlement/finalization.
- Retain the last confirmed board during disconnection, pause ordinary actions,
  and reconcile an uncertain operation before allowing a duplicate submission.
  The demo illustrates this contract without durable networking or storage.

The [numbered storyboard](mockup/README.md) and implemented chooser cover setup,
deal, main board, combo inspection, opening, attack selection/responses/result,
trade proposal/pending/completion, purchase selection/result, loan/alliance,
Confinement, Coup, settlement, standings and reconnect. Independent examples are
separate branches, not twenty consecutive actions from one legal deal.

## Code ownership and cleanup

The owner requested relocation on 2026-10-01. Keep `client/` as the runnable reference host and
`client/packages/cgms_ui/` as the single reusable presentation package (`packages/cgms_ui` relative to the host).
No second Flutter app, copied theme or duplicate asset bundle is needed to freeze
the design. Future client integration follows [INTEGRATION.md](../../client/INTEGRATION.md):
consume or relocate the complete package with its assets, ARB localization,
notices and tests; adapt authoritative state without importing the demo controller.

Retain the twenty main mockups, current source, tests, current widget renders,
asset sources, manifests, licenses, accepted rules and decision history. Retire
the superseded gameplay v1/v2 images and their active layout guidance. The old
concept study is retained only under deck source provenance, not as a frontend
reference. Temporary build/cache outputs remain ignored and reproducible.

Older `CgmsTableView`, `CgmsTouchTableView` and `DemoApp(legacy: true)` code remains
only for existing API consumers and regression coverage. It does not define the
current v1 UX and is not exposed by the approved example chooser. Removing this
covered compatibility code is a separate migration; do not erase its tests just
to make cleanup pass.

## Verification and change policy

The [app README](../../client/README.md#validate) gives the commands and current evidence.
The later [direct-drop evidence](../reports/2026-10-04-direct-drop-revision.md)
records the complete-drop and scheduled-draw owner revision and its exact gates.
Retain host flow/routing/state tests, package behavior/semantics tests,
[asset validation](../../client/tool/verify_assets.py), and the independent public consumer
test. Review the normal board, opened-series board, pending response, inspection
and result at phone/tablet/desktop sizes after a relevant UI change.
The three [rendered board images](../../client/README.md#rendered-board-evidence) are
package-fixture evidence, not full-host or native-device captures.

The owner has authorized the 2026-10-04 card-interaction revision without a
further layout approval. Preserve the frozen theme/resources and semantic
contracts above while changing navigation, layouts and presentation components.
Update affected tests and class-specific review evidence during implementation;
do not erase historical evidence or compatibility tests to obtain a passing gate.
Further material changes beyond this scope follow the decision log/change policy.

P04 passed its adaptive web, authority journey, recovery and browser accessibility
gates on the recorded build. R06N retains final native engine,
packaging/device/platform and accessibility verification before R06/R07 release.
