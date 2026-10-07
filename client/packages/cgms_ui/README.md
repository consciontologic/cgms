# CGMS presentation package

This is the single reusable presentation package for the
[CGMS UI baseline](../../../docs/design/UI_BASELINE.md). The 2026-10-04 owner
revision adds card-driven interaction to the phone-reference vertical board at
every width, enlarging items on tablet/desktop. Keep the package at
`client/packages/cgms_ui/`; no duplicate
app or theme is required. Browser/native acceptance remains separate.

`cgms_ui` provides platform-neutral Flutter themes, cards, viewer-filtered
presentation models and reusable setup, table, trade and score views. Import
only `package:cgms_ui/cgms_ui.dart`; the package never imports the demo host,
browser APIs, endpoints or an authoritative game engine.

The [host integration guide](../../INTEGRATION.md) documents relocation,
adapters, assets and the remaining integration work. Fonts and illustration
are loaded through `packages/cgms_ui/` asset paths. Package-owned English
resources are in `lib/src/l10n/cgms_en.arb`; generated Dart is included.
The [demo README](../../README.md) contains host commands, and the independent
[consumer test](../../test/public_consumer_test.dart) imports only the public API.

When relocating, retain this complete folder, including `pubspec.yaml`,
localization configuration, assets, [LICENSE](LICENSE) and
[asset notices](assets/NOTICE.md). Bundled fonts use SIL OFL 1.1; UI acceptance
does not change the generated card art's provenance or licensing status.

```dart
MaterialApp(
  theme: CgmsTheme.dark(),
  localizationsDelegates: CgmsLocalizations.localizationsDelegates,
  supportedLocales: CgmsLocalizations.supportedLocales,
  home: consumerView,
);
```

## Shared adaptive presentation

Online hosts omit `CgmsAdaptiveTable.actions` for the card interface. No action
workspace, offstage form, layout reservation or action toggle is built in this
mode. Existing demo/API consumers may still supply `actions` for compatibility.
The direct public quadrants and adjacent compact hand share the existing theme
and art; short/enlarged screens scroll instead of reducing physical controls.

`onSelect` receives an exact physical card on single tap or Space.
`onActivateCard` receives that card on double tap or Enter, without preceding
selection changes. `onInspect` remains independent through I, assistive custom
actions, public full-card inspection and the selected-card magnifier. Card roles
(`source`, `target`, `payment`, `modifier`) expose both distinct icons/borders
and localized semantics; hosts own their meaning and legal timing.

`onTargetTap` provides the same own/opponent card, series, formation, seat, hand
and draw-pile destinations as drag/drop. `TableDropZone.series` includes the
public suit and seat; it never aliases an arbitrary first card. The separate
`onActivateTarget` callback activates a series or formation by its typed identity
on double tap/Enter. Hosts resolve its current authorized members, keeping a
whole series distinct from an individual number card or attached royal.
`onCardDrop` receives the exact selected authorized bundle. In card mode, a
revoked member invalidates the entire drag rather than silently reducing it.
`dragIdentity` is captured in `TableCardDrag.intentIdentity` and rechecked at
drop, alongside game identity and current ownership. Supply the host's current
version/turn/window/effect/decision identity. Optional `canDrop` filters and
highlights contextual destinations; it remains a presentation hint, never a
legality guarantee. Touch requires a hold before drag so ordinary swipes scroll.
Escape clears the active drag and invokes `onCancelSelection`; no cancellation,
inspection, target or gesture callback submits a command inside this package.
The optional `dropPreview` callback supplies a viewer-safe move label for the
exact accepted hover; its temporary live-region feedback clears on leaving,
cancellation or invalidation.

The normal local entry and online host use `CgmsAdaptiveTable`. Its public piles
contain only supplied cards/formations; no hidden inventories or eligibility are
reconstructed. All sizes retain the public overview and adjacent private hand.
Online decisions appear temporarily above that board; compact non-card utilities
follow the hand. Larger widths enlarge the same items.
Drafts, caret, selection, focus and pending operations survive resizing.

Hosts supply `publicGroupLabels` (keyed by `seatId:suitName`) and
`onInspectCombo` for authorized captions and details. Public stack controls open
a chooser of full physical cards; that chooser tracks current authorized data.
For compatibility hosts that supply `actions`, `showDecisions` requests exposure
of that region without submitting or replacing the other sections.
Those hosts can supply `actionsIdentity` for a meaningful decision-workspace change,
such as entering round departure choices. It creates a fresh scroll viewport
with separate saved offsets while retaining the action subtree's state. Keep
this identity stable across ordinary snapshots and resizing; do not use a
command version or current actor as its identity.

`CgmsPanel` supplies a shared opaque reading surface with an optional directional
accent. `CgmsSpacing`, `CgmsShape` and `CgmsColors` centralize spacing, control and
panel geometry, separators and error color. Cards, account panels, notices,
dialogs and menus use the same existing cgmsart palette and bundled fonts.
All `CgmsCard` modes support independent keyboard **I** and visible magnifier
inspection through `onInspect`. With `onActivate`, double tap and Enter open the
card action; neither also selects or submits. Without `onActivate`, legacy
callers retain double-tap inspection. Single tap selects only, and long press
does not magnify. Full inspection cards keep
a hollow bottom-right magnifier with a transparent 48px accessible target;
its tooltip says “Magnify card”. Compact board cards
use an accessible inspection action without consuming their selection area;
public previews expose one >=48px group target.

The image box matches `VisibleCard.artworkAspectRatio`, sourced from all 52
native suit/rank dimensions (shared across four styles). `BoxFit.contain` shows
the complete painting without cropping, stretching or padding. Live rank/suit
and selection sit above it; meaningful status sits outside it. Compact board
cards retain the reference's small copy number, for authorized cards only.
Full cards and inspector titles omit visible copy numbers while retaining them
in semantics and model identity. `displayStatus` removes incidental zone names. Large text
expands the header width while preserving the painting's native proportions.

`VisibleCard.style` accepts `CardStyle.firstLight`, `rainGlaze`, `emberGlaze` or
`drypoint`. Adapters supply the current controller's public match assignment;
the package never assigns styles or derives ownership from an illustration.
An authoritative transfer therefore changes the painting while retaining the
physical ID, live rank/suit labels, semantic copy, status, selection and inspection keys.
All four sets of 52 native crops are bundled; First Light remains the default
for standalone consumers and legacy projections. Unknown wire styles also
resolve to that default through `CardStyle.fromWire`. `ConcealedCard` accepts no
identity or style and always uses the same shared back. Style is decorative and
does not add card identity information to semantics.

## Shared action board

`CgmsActionBoard` remains a portable board API over the same authorized models.
This legacy compatibility surface is not the current online gameplay interface.
Its composition follows the older compact reference: public two-column seats,
stack previews, action area, complete private hand and footer in vertical order.
Long holdings and larger text may scroll/reflow; exact physical selection and
inspection stay accessible. Older table/touch APIs remain compatibility surfaces.

```dart
CgmsActionBoard(
  data: adapter.projection,
  selectedCardIds: adapter.selectedPhysicalIds,
  onSelect: adapter.selectPublicCard,
  onSelectHand: adapter.selectHandCard,
  onClearHandSelection: adapter.clearSelection,
  onInspect: openCardInspector,
  onInspectCombo: openComboInspector,
  onPrimary: adapter.submitAction,
  onTrade: openTrade,
  onScores: openScores,
  onEndTurn: adapter.endTurn,
  onReconnect: adapter.reconnect,
  endTurnEnabled: adapter.canEndTurn,
);
```

Only the public package import is required. `actionPanel`, `header` and `footer`
are optional widget slots for the embedding app; the app owns availability and
localization inside those slots. `selectedCardIds` supports multi-card drafts
alongside `data.selectedCardId`. Optional `publicGroupLabels`, keyed by
`seatId:suitName`, carry adapter-calculated display totals; the view itself
does not infer combat strength or card effects. Supplied combo membership keeps
its cards out of ordinary suit piles, so each exposed physical card appears once.

Public pile previews preserve supplied suit/combo membership and open full-card
controls for individual selection and inspection. Private hand cards retain
single-tap selection, independent keyboard `I` inspection, contextual double-tap
actions when `onActivateCard` is supplied, and authorized
physical identity. Disconnection and pending states prevent ordinary selection
callbacks without preventing authorized inspection.

Host and package tests exercise compact/reference geometry, all width classes,
card and chooser interactions, text growth and short-view scrolling. Optional capture
runs using `--dart-define=CGMS_CAPTURE_BOARD=true` write disposable images under
`../../.browser-tools/artifacts/widget-renders/`; never overwrite the retained
phone reference. Those widget renders do not certify browser/native behavior.

## Compatibility collections

The following APIs are retained for existing consumers and regression coverage.
They do not override the current shared vertical composition.

`CgmsHandView` accepts only cards authorized for the viewer. It preserves
physical identities and supplied hand order, offers All/suit filters with
counts, and keeps an existing selection inspectable when a filter hides its
card. The consumer owns selection; the widget owns only the browsing filter.
Key this view by immutable game ID to reset its filter between games.

```dart
CgmsHandView(
  key: ValueKey(gameId),
  cards: playerHand,
  selectedCardId: selectedPhysicalId,
  onSelect: adapter.selectHandCard,
  onClearSelection: adapter.clearSelection,
  onInspect: openCardInspector,
);
```

When `onSelect` is supplied, single tap selects. Double-tap or the corner
magnifier inspects through `onInspect`, including when selection is unavailable.
The selected-card summary keeps Clear; if filtering hides the selected card,
its full preview remains available for magnification. `showArt: false` retains
live ranks, vector suits, meaningful status and semantic physical identity.

Use `compact: true` on `CgmsTableView` for one public board, smaller cards,
a stacked hand, adjacent action area and an expandable journal. Narrow layouts
provide a separately scrolling public board; larger text returns to natural document flow.
`CgmsHandView` and `CgmsCard` also accept `compact: true` independently. Compact
selection summaries follow the hand so selection never shifts its cards. Public
card review retains an explicit inspection control even for rejected moves;
reviewing a card does not manufacture an authoritative selection.

`CgmsSuitStacks(cards:, onTap:, onInspect:)` presents complete visible cards in
one group per suit. Optional `selectedCardId`, `cardWidth` and `showArt` configure
it. Cards never overlap; native image proportions and meaningful status remain
visible. Rank/suit headers grow with text scale. Input order within each suit
and every physical ID are preserved. `CgmsCard.stackHeader` remains accepted for
compatibility; every mode now places identity outside the painting. Neither API
infers legal actions or open series.

Compact hands use the same suit stacks with `cardKeyPrefix: 'hand-card'` and
`expandSelectedCard: false`; selection never shifts the remaining hand cards.
Both options are available on `CgmsSuitStacks` (defaults: `public-card`, `true`).

`ComboView` carries adapter-provided `id`, `title`, `stateLabel`, `description`
and constituent `cards`. Pass exposed groups through `SeatView.combos`, private
candidates through `TableViewData.handCombos`, and handle
`CgmsTableView.onInspectCombo`. Constituent IDs refer to existing holdings;
public groups are rendered once and private candidates remain private.
`CgmsComboGroup` and `CgmsComboInspector` also work independently. Their labels
and ability state are supplied by the consumer, with no recipe/legality engine.

## Compatibility: earlier phone and tablet interaction

`CgmsTouchTableView` accepts the same viewer-filtered `TableViewData` and narrow
action callbacks as the desktop table. Place it in a bounded `Scaffold.body` or
`Expanded`. It owns scroll regions, a seat switcher, Table/Hand/Activity navigation
and the action dock. Tablet layouts show a focused public position beside the
private hand. Enlarged text and short windows use natural scrolling.

The component stores only browsing state. Supply `onSelect` for public cards,
optional `onSelectHand` / `onClearHandSelection` for the hand, `onInspect` and
optional `onInspectCombo` for detail views, and `onPrimary`, `onTrade`, `onScores`,
`onEndTurn`, `onReconnect` for consumer actions. `endTurnEnabled` and
`endTurnDisabledReason` describe ending availability; `showArt` retains the
existing artwork fallback. `data.actionEnabled` and connection state gate the
primary intent. No callback is submitted by changing tabs or seats.

The demo-specific route and menu live in the host. See the
[touch integration example](../../INTEGRATION.md#touch-table-integration) for
wiring. A consumer can retain both table presentations without duplicating its
adapter, fixtures or game state.

## Package checks

From this package directory:

```sh
flutter pub get
flutter gen-l10n
dart format --output=none --set-exit-if-changed lib test
flutter analyze
flutter test
```

The tests cover exact arithmetic, hidden-card rendering, distinct physical
copies, fallback artwork, selection and inspection, suit filters and counts,
keyboard activation and jump focus, game-ID filter reset, reduced motion,
200% text at 390 pixels, and semantic color contrast. Browser screenshots and
the host's wider responsive and scenario checks are documented in the host.
