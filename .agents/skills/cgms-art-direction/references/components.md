# cgmsart components and interaction — shared vertical board

Read the [UI baseline](../../../../docs/design/UI_BASELINE.md). The implemented
[public package API](../../../../client/packages/cgms_ui/README.md) and
source below replace the former candidate widget inventory, card dimensions,
notched-button examples and proposed navigation system. Reuse current components;
do not build those retired examples as a separate redesign.

## Existing component boundaries

| Responsibility | Current authority |
|---|---|
| Shared public quadrants and private hand | [CgmsAdaptiveTable](../../../../client/packages/cgms_ui/lib/src/adaptive_table.dart) |
| Live rank/suit, semantic copy identity, artwork, focus and inspection | [CgmsCard and inspectors](../../../../client/packages/cgms_ui/lib/src/components.dart) |
| Colors, typography and control states | [CgmsTheme](../../../../client/packages/cgms_ui/lib/src/theme.dart) |
| Authorized presentation data and exact values | [Models](../../../../client/packages/cgms_ui/lib/src/models.dart) |
| Portable labels | [ARB source](../../../../client/packages/cgms_ui/lib/src/l10n/cgms_en.arb) and generated localizations |
| Flow actions, menus, scores and history | [ApprovedFlowView](../../../../client/lib/approved_flow.dart) |

The host owns legal selection and action content, including optional header,
action and footer slots. The board renders supplied physical identities and
multi-selection; it does not calculate game permissions or combat rules.
Adapter-supplied public totals must remain distinct from physical return costs.
Legacy package views remain compatibility surfaces, not alternate default UI.

## Card interaction and accessible piles

The current phone reference uses overlapping public suit/combo previews and a
compact private hand grouped into horizontal suit stacks with uniform 4:5 frames.
Each hand card exposes at least 48 logical pixels for selection; complete paintings
fit without crop/stretch and keyboard focus scrolls into view. Public previews
fit inside four equal quadrants in a table aligned to the hand's full width,
using the hand's card-size scale where space permits and sized to keep the adjacent hand
visible without page scrolling. Each quadrant must open full-card controls for
every authorized physical ID; thin rank tabs cannot be the only selection path.
A group is presentation, never a rule-derived aggregation or concealed breakdown.
Private cards keep independent selection and inspection. Compact board cards
show the reference's small copy number, only for authorized cards; full cards
omit visible copy labels. Physical identity remains in semantics/model data;
inspectors retain full native-ratio artwork.

Single tap selects only when applicable. Double-tap and keyboard **I** inspect
independently; rejected selection and long press do not inspect. Compact controls
must expose an accessible inspection action without an overlay consuming the
selection hit area. Enter/Space activates the focused control, never a global
submission. Group choosers and inspectors contain keyboard focus and dismiss
without changing gameplay. Re-resolve chooser members from the current authorized
projection so removed/private cards cannot linger in a captured dialog snapshot.

Selection, usage and focus coexist: cyan selection/check, explicit used state,
lime focus. A used attacking Club stays exposed; it must not look discarded.
Public hand size and concealed-Ace count are separate. Hidden identities,
royal/eligible-card breakdowns and split-card partner locations cannot leak in
labels, tooltips, semantics, asset paths or public view models.

## State meanings to retain

| State | Meaning and behavior |
|---|---|
| Default | A real verb/target or passive fact; no fake affordance on decorative art |
| Hover | Supplementary preview only, never sole access to rules or an action |
| Pressed | Stable target; activation occurs once through the intended callback |
| Focus | Clearly visible independently of selection; retained during unrelated updates |
| Selected | Supplied physical IDs stay visible until changed/revalidated |
| Disabled | No mutation; specific reason remains readable or in accessible Details |
| Submitting | Context retained; duplicate submission blocked; no success claim |
| Pending | Exact required actor/response stage shown; no countdown or automatic pass |
| Error/unknown | Preserve valid draft and distinguish rejection from unknown acceptance |
| Resolved | Past-tense result and durable history only after authoritative resolution |

Disabled/busy blocks invocation; press overrides hover while focus and selection
remain independent. Revalidate a draft when its cards or permissions change.
Artwork, animation or a dialog dismissal cannot resolve an action.

For the demonstrated 8♣→8♥ attack, declaration consumes that physical Club's turn
use. Successful resolution returns the Heart to the deck and awards no points;
submission and response waiting are not success. Keep this explanation in the
existing action/details/history surfaces, within the shared vertical hierarchy.

A legally available Coup remains immediate, including permitted out-of-turn and
pending-action cases. Never add confirmation, a hold gesture, countdown,
response window or animation delay. Modal policy and callback admission must
respect the canonical rules; the bounded demo does not certify full engine coverage.

## Motion, fallback and feedback

Retain the approved stable controls and reduced-motion behavior. Motion may
connect an action to its result but cannot grant permission, change timing or
be the only result record. No animated grain, pressure timers, fake progress or
reward effects. Haptic/audio feedback is supplementary only.

With artwork disabled or unavailable, identity, ownership, values, status,
selection and focus remain live. Restore artwork for visual regression review:
an art-off check does not replace acceptance of the illustrated baseline.
Inspectors and menus announce their purpose, preserve focus and dismiss only
where allowed; closing them must not pass, accept, resign or replay a request.

## Shared presentation

The owner-selected vertical arrangement is common to phone, tablet and desktop.
Reuse the authorized models and controls; size and reflow them for available
width/text scale while retaining public/action/hand/footer order and state.
