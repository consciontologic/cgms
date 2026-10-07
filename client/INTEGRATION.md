# Integrating `cgms_ui`

[UI baseline](../docs/design/UI_BASELINE.md) is the adaptive presentation contract.
Continue development in `client/` with its local `packages/cgms_ui`
dependency. Share authoritative adapters, gameplay logic, state, action intents,
privacy-filtered models, theme and primitives in one vertical board at all
widths; tablet and desktop enlarge the phone-reference items. The existing API examples
below describe v1. `CgmsAdaptiveTable` adds the online adaptive composition. Never copy the
demo controller into reusable widgets or fork the theme. Both the approved local
flow and online route now consume `CgmsAdaptiveTable`; retained legacy primitives
are not the normal Auto entry composition.

P04 runs through Flutter web independently of P01/native toolchains. Phone/tablet
web must reproduce intended native appearance and user journeys using the
baseline's explicit class and browser acceptance matrix. Keep route/session,
selection and pending-operation state above presentation switching; changing
class must not remount authoritative state or repeat submissions. Go remains
online authority. P01/P02 research adapters cannot enable offline web gameplay.
Native adapters, packaging, lifecycle, performance and device acceptance are
final-stage R06N checks, not inferred from package portability or web tests.

`cgms_ui` is a portable Flutter presentation package. The demo host consumes it
through `path: packages/cgms_ui`; all package assets, fonts, localization sources,
models, theme and views travel with that folder. It has no dependency on demo
routes, browser APIs, mock services or a global controller. The host is web only;
portability of the package does not constitute verification of native builds.

## Online adapter boundary

`OnlineClient` in `lib/online/online_client.dart` is host-owned. It bootstraps the same-origin session,
creates/joins rooms, subscribes to authorized snapshots and resubmits only an
explicitly reconciled original operation. `OnlineScreen` owns typed drafts and
physical-card selection; game, window and decision identities are captured on
the first edit and are never silently replaced. `projection.dart` converts only
viewer-authorized data into package models, using exact rational amounts and
separating payables from receivables. `financesKnown` prevents unavailable opponent
balances from appearing as zero-valued public facts.

`CgmsAdaptiveTable` takes injected table data, selection/inspection callbacks,
eligible immediate Coup, and an action widget. It owns only presentation state.
All sizes use the same public overview, action region, private hand and bottom
actions in one vertical scroll. Public stacks open accessible full-card choosers;
resizing adjusts item sizes without switching or hiding workspaces. Host
routing keeps the online screen mounted during local navigation and writes the
active match or lobby's public room ID to the URL for restoration after reload.
Session recovery precedes the authorized room lookup; invitation tokens and room
contents never enter that URL. Match locations take precedence over lobby locations.

The online host applies `VisualDensity.standard` so desktop platform defaults do
not shrink primary targets below 48 logical pixels. Preserve that density when
embedding the adaptive view for mixed pointer/touch use. Public pile controls and primary actions must remain at least 48 logical pixels;
compact preview rank tabs are not individual touch targets. The shared
text theme includes a locally bundled symbols fallback for strict same-origin CSP.

## Add it to an application

Use the existing host for P04; do not create a parallel Flutter application.
The following dependency example is for future authorized external consumers.
For an existing application, add a local dependency whose path points from its
`pubspec.yaml` to this package:

```yaml
dependencies:
  flutter:
    sdk: flutter
  cgms_ui:
    path: ../relative/path/to/client/packages/cgms_ui
```

The example path is intentionally application-relative; replace it with the real
location. Then run `flutter pub get` in that application. To relocate, move the
entire `packages/cgms_ui/` directory, including `pubspec.yaml`, `lib/`, `assets/`,
`LICENSE`, localization configuration and notices; change the dependency path
and rerun `flutter pub get`. Do not copy individual feature-screen files.

Import only the package barrel:

```dart
import 'package:cgms_ui/cgms_ui.dart';
```

For example, a consumer can embed the reusable setup view without importing any
demo code:

```dart
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';

class MatchSetupPage extends StatelessWidget {
  const MatchSetupPage({super.key, required this.createMatch});

  final void Function(String playerLabel, int gameCount) createMatch;

  @override
  Widget build(BuildContext context) => MaterialApp(
    theme: CgmsTheme.dark(),
    localizationsDelegates: CgmsLocalizations.localizationsDelegates,
    supportedLocales: CgmsLocalizations.supportedLocales,
    home: Scaffold(body: CgmsSetupView(onStart: createMatch)),
  );
}
```

In an existing app, put the view under its existing `MaterialApp` and register
the theme/delegates there rather than nesting another application.

Internal `src/` files are implementation details. The public consumer test in
[`test/public_consumer_test.dart`](test/public_consumer_test.dart) is the executable
example of an independent adapter supplying the same table view.

## Ownership and API

| Reusable package | Embedding application / demo host |
|---|---|
| Theme tokens, typography and shared styles | Application entry point, theme selection and preference persistence |
| Visible/concealed cards and presentation components | Authorized projections, physical-card handles and state adapters |
| Setup, table, score, trade and inspection views | Navigation, dialogs, controller lifecycle and callback wiring |
| Typed presentation data and action affordances | Command validation, operation identities, services and updates |
| Native art, fonts, licenses and ARB catalogs | Dynamic event/reason localization and locale selection |
| Exact point representation and display | Authoritative scoring, debt settlement and ledger history |

The approved board is `CgmsActionBoard`; shared feature views include
`CgmsSetupView`, `CgmsScoresView`,
`CgmsTradeView` and `CgmsCardInspector`. Their callbacks express player intent;
the consuming application chooses how to perform it. `CgmsActionBoard`
receives a `TableViewData` and callbacks for selection, inspection,
the primary action, trading, scores, ending the turn and reconnecting.

For the approved board, supply `onSelectHand` and `onClearHandSelection`, and
return selected physical IDs through `selectedCardIds` for multi-card drafts.
`TableViewData.selectedCardId` remains supported for a single selection. Keep
the full flat hand visible; public piles open individual-copy review. The adapter
owns every allowed action and response window, including actions near the hand.
Counts describe supplied visible cards, not inferred combat strength or legality.
`endTurnEnabled` and `endTurnExplanation` retain an explicit disabled action and
its reason in the board's default footer.

The retained compatibility `CgmsHandView` has local suit filters and the earlier
`CgmsTableView` has navigation anchors. Those browsing features belong to
the [earlier compact table](#compatibility-earlier-compact-table), not the current v1
board. Existing consumers should key those browsing widgets by immutable game ID
so local filters reset between games; filtering never changes card ownership.

Construct a fresh authorized projection when state changes and rebuild the
view. Keep controllers and service subscriptions in the application. Views do
not fetch endpoints, decide online legality or navigate to host routes.

## Theme and localization setup

Register `CgmsLocalizations.localizationsDelegates` and
`CgmsLocalizations.supportedLocales` on the embedding `MaterialApp` or
`MaterialApp.router`. Apply the public cgmsart theme so its package-aware font
families and color tokens reach the views. The host entry point shows the
complete composition, and the public consumer test independently wires it.

All package-owned static component and feature-view text lives under
`lib/src/l10n/cgms_en.arb`. After changing ARB resources, run from the package
directory:

```bash
flutter pub get
flutter gen-l10n
flutter analyze
flutter test
```

Add translated ARB catalogs and regenerate before advertising another locale.
The package does not translate injected player labels, scenario descriptions,
event lines or availability reasons. Map server error/reason codes through the
consumer's localized catalog before creating display state; do not display raw
backend internals or secret eligibility details.

## Current v1 action board

The current v1 mobile/tablet/desktop implementation uses the exported
`CgmsActionBoard`; this API description does not mandate its composition for v2. It accepts `TableViewData` and the existing table callbacks,
plus `selectedCardIds` for multiple highlighted physical cards and optional
`header`, `actionPanel`, `footer` and `publicGroupLabels` presentation content.
The host supplies those regions without moving game rules into the package.
All public formations and the complete flat hand grid remain in one composition;
compact public stacks open an inspector with individually accessible cards.
See the host's `approved_flow.dart` and package action-board tests for concrete
composition and responsive examples. The earlier `CgmsTableView` API below
remains available for existing consumers and regression fixtures.

## Compatibility: earlier compact table

Pass `compact: true` to `CgmsTableView` to use a shared four-seat board with the
stacked hand and one action area beside it on wide screens, or below on narrow
screens. The default remains `false` for existing
consumers. `CgmsCard` and `CgmsHandView` also expose independent `compact` options;
all are exported through `package:cgms_ui/cgms_ui.dart`.

```dart
CgmsTableView(
  data: adapter.projection,
  compact: true,
  onSelect: adapter.selectPublicCard,
  onSelectHand: adapter.selectHandCard,
  onClearHandSelection: adapter.clearSelection,
  onInspect: openInspector,
  onPrimary: adapter.submitAction,
  onTrade: openTrade,
  onScores: openScores,
  onEndTurn: adapter.endTurn,
  onReconnect: adapter.reconnect,
);
```

The compact board uses the public `CgmsSuitStacks` component, which can also be
embedded independently:

```dart
CgmsSuitStacks(
  cards: adapter.visiblePublicCards,
  selectedCardId: adapter.selectedPhysicalId,
  onTap: openInspector,
);
```

It preserves supplied order within each suit and never merges physical copies.
`cardWidth` and `showArt` are optional. Each covered card keeps a readable,
keyboard-accessible identity header; only covered artwork is clipped. Status
cards expand naturally, and columns wrap when text grows.
`cardKeyPrefix` defaults to `public-card`; `expandSelectedCard` defaults to `true`.
The compact hand uses `hand-card` and `false` to retain stable geometry while its
reserved check slot communicates selection. The lower-level `CgmsCard(compact: true, stackHeader: true, ...)` places copy and status
above artwork for this presentation. Concealed holdings remain counts and the
shared back; they must never be passed to these visible-card components.

`VisibleCard.style` selects one of the four existing editions. Online adapters
resolve it from `projection.board.players[].card_style` and the card's current
authoritative `controller`, never the viewer's theme or original owner. This
public assignment lasts an entire match and is independent of private chance,
card handles and entitlements. Capture changes the displayed style without
changing identity. Missing/unassigned style uses First Light; concealed cards
still use only the neutral back. Reconcile open inspectors against each fresh
authorized projection so neither stale artwork nor a no-longer-visible face
remains displayed after a transfer.

Compact mode changes presentation only. A tapped own public card gets a local
inspection reference even if the adapter rejects the selection; only
`data.selectedCardId` marks a card as selected for a move. Inspection remains
available during pending/disconnected states. Hand selection, filters and exact
physical copies keep their existing contracts. The table owns its board scroll
controller and hand jump focus; the embedding view should supply normal vertical
page scrolling. On narrower widths, the public board scrolls independently.
Large text removes that height cap and expands card labels naturally. Consumers
may expose the density preference through their own settings; it is unrelated
to navigation or game state. Fonts must be loaded when measuring geometry in
widget tests, since Flutter's default Ahem test font has different metrics.

## Combo projections

Supply `SeatView.combos` for exposed formations and `TableViewData.handCombos`
for private patterns visible to the viewer. Both default to empty for existing
adapters. Each `ComboView` contains `id`, `title`, `stateLabel`, `description`
and authorized constituent `cards`. Those cards must reference the same physical
IDs already present in that owner's exposed holdings or the viewer's hand.
The table groups public allocated cards exactly once and keeps private candidates
in the hand. Supply only formations the viewer is authorized to see; the package
does not detect recipes, verify support or decide activation eligibility.

Wire `CgmsTableView.onInspectCombo` to your router or dialog. The host uses
`showDialog` with `CgmsComboInspector(combo: combo)`. `CgmsComboGroup` is the
standalone formation component, with `onInspectCard`, optional `onInspect` and
`showArt`; all are public-barrel exports. Without an inspection callback, public
groups show the supplied description inline. Static inspection/pattern labels
come from package ARB resources; combo names, states and explanations come from
the localized adapter. Keep potential hand patterns distinct from opened or
active formations, and preserve support, turn and consumed-allowance state from
the real authority. The approved host contains a scripted Ponzi demonstration;
the legacy combo-inspection fixtures and package do not implement a general
combo activation engine.

## Touch table integration

This section documents the retained compatibility API only. The current v1 phone,
tablet and desktop UI all use `CgmsActionBoard`, as described above.

`CgmsTouchTableView` is a second public presentation of `TableViewData`. Place it
in a **bounded** region such as `Scaffold.body` or `Expanded`; it owns its scroll
regions and action/navigation areas. Do not wrap the whole widget in an
unbounded `SingleChildScrollView`. The original `CgmsTableView` API and layout
remain available for the desktop overview.

```dart
Scaffold(
  body: SafeArea(
    child: CgmsTouchTableView(
      data: adapter.table,
      onSelect: adapter.selectPublicCard,
      onSelectHand: adapter.selectHandCard,
      onClearHandSelection: adapter.clearSelection,
      onInspect: openCardInspector,
      onInspectCombo: openComboInspector,
      onPrimary: adapter.submitPresentedAction,
      onTrade: openTrade,
      onScores: openScores,
      onEndTurn: adapter.requestEndTurn,
      endTurnEnabled: adapter.canEndTurn,
      endTurnDisabledReason: adapter.endTurnReason,
      onReconnect: adapter.reconnect,
    ),
  ),
)
```

The callbacks above belong to your adapter/router; they are example consumer
names, not additional package APIs. Phone tabs, the viewed seat and scroll
positions are presentation state. Card custody, turn/response actors, action
availability, exact finances and combo visibility still come from the adapter.
At ordinary text sizes tablets use independent public/hand panes with
Table/Activity navigation; larger text
and short windows favor normal scrolling. Use stable immutable game IDs and
physical card IDs across projections.

The explicit legacy regression fixture retains older presentation APIs;
normal Auto uses <600 / 600–1049 / >=1050 solely for sizing.
In the explicit `DemoApp(legacy: true)` regression fixture, the **Table layout**
menu and preview frames belong to the host. Its `#/play` selects Auto: below
600px the compatibility touch view uses phone controls, from 600px tablet sizing,
and from 1050px desktop sizing. This legacy API does not define the current
shared vertical normal-entry layout.
Explicit `#/phone` (maximum 440px), `#/tablet` (768–1024px canvas) and `#/table`
(Desktop) routes provide review overrides. The legacy `#/touch` route continues
to choose between phone and tablet. Only an explicitly forced tablet preview
needs horizontal canvas scrolling in a narrow browser; Auto uses the actual
available space. The touch view still chooses its accessible natural scrolling
fallback for large text and short windows.

That legacy host reuses one adapter across its modes. Existing consumers of these
APIs can keep the same data and callbacks during migration to `CgmsActionBoard`.
Keep the compatibility touch view bounded and give its desktop overview ordinary
page scrolling. The package itself has no demo mode menu, preview frame or
routing dependency. Do not use these retired compositions for new CGMS screens.

The router remembers the chosen mode for returns from trade, scores and scenario
pages. Browser history changes views without dispatching a game command, and a
page refresh starts a fresh local demo. Projection-owned selection and pending
actions survive mode changes; view-local seat/tab/scroll state may reset when
switching presentation families. A production router can use different paths
without changing the package.

## Assets and fonts

Images are package-owned and declared in `pubspec.yaml`. Load an image using its
package-relative name and `package: 'cgms_ui'`, for example:

```dart
Image.asset('assets/back.png', package: 'cgms_ui');
```

For direct bundle access, use the Flutter asset key
`packages/cgms_ui/assets/back.png`. Never use filesystem paths, the demo working
directory or a path outside the package. The components apply package-qualified
font families; retain those qualifications when extending the theme.
See Flutter's [asset and package-asset documentation](https://docs.flutter.dev/ui/assets/assets-and-images)
for the supported bundle conventions.

Retain `assets/manifest.json`, [NOTICE.md](packages/cgms_ui/assets/NOTICE.md), the
font OFL notices and `LICENSE` when relocating or distributing the package.
The manifest maps copied paintings back to the original cgmsart deck crops
and records checksums. Live labels are separate from paintings; concealed cards
always use the shared back. Card identity is data supplied by the authorized
projection and must not be inferred from art, edition or asset filename.

## Authoritative state adapters

The online host uses the implemented [API](../docs/code/API-game.md) and
[OpenAPI schema](../docs/api/online.openapi.json); local demonstration routes keep
their deterministic adapters. Preserve these boundaries when extending the
online projection adapter or embedding the package in another application:

1. Receive an authenticated, seat-filtered snapshot and its event cursor. Map
   only authorized data into presentation models. Another player's concealed
   cards become counts, never private `VisibleCard` objects hidden by widgets.
2. Preserve the opaque ID of every physical card. Equal rank/suit pairs can be
   two different cards; selection and command payloads use physical handles.
3. Map exact rational amounts without passing through `double`. Preserve
   numerator/denominator strings from the authoritative transport and construct exact
   values. Keep game score, match score, spendable points, debts and adjustments
   separate; display formatting must never change settlement amounts.
4. Supply action availability and a player-readable reason. Selection is local
   UI state; execution sends a typed intent to the authority and awaits its
   result. An enabled button is advisory, not an authorization decision.
5. Replace the projection only after committed results/events arrive. Handle
   loading, pending, failure and disconnection explicitly. An uncertain commit
   remains pending and must be reconciled before resubmission.

All game-scoped commands must retain their origin `game_id`, operation ID and
relevant turn/action/window/effect/decision IDs. Ordinary actions also bind to
the selected state version. The host's fixture reset is not a production game
transition: real game IDs come from the authority. Never retarget a stale intent
to a newly created or nullified replacement game. Reconnect cannot invent a new
operation ID, pass for a disconnected player, or reveal another seat's cards.

The online authority and host implement authentication, authorization, safe errors,
deduplication, reconciliation, ordered event application and resnapshot behavior.
These responsibilities do not belong in the widgets. The implemented
[API](../docs/code/API-game.md) and
[delivery contract](../docs/design/DESIGN-go-state-and-delivery.md) describe the online
boundary used by `#/online`; the local demonstration remains a separate route.

## Trade and finance semantics

Keep a proposal's status separate from whether accepting it is currently legal.
A proposal reserves no cards. Acceptance binds to its exact revision and origin
turn/game, starts an action with the acceptor as actor, and does not immediately
move cards. Only successful resolution moves all agreed cards and creates a
loan-derived alliance or privilege. Pending offers expire at rule boundaries,
not because a UI timer elapsed.

Keep `playing`, board-closed settlement and financially finalized states
distinct. A result screen may show a closed board while payments remain pending.
Do not turn absence into refusal or allow next-game transition before required
settlement. New-game available points start at zero; retained match scores and
unpaid debts remain. Show the declaration winner, game score leader and eventual
match rank separately where they differ. Debt/reputation, Coup, departure and
nullification accounting remain the authority's responsibility; do not reproduce
that accounting in the presentation package.

## Navigation and browser behavior

The demo host owns its Router and hash locations. Its URLs identify presentation
views, never card IDs, private terms or concealed identities. Browser Back/Forward
preserves the active board across layout/view changes. Entering setup creates a
fresh prepared room branch; choosing an example also replaces the current branch.
Refresh resets local state.
No package widget assumes that routing scheme.
The host uses Flutter's [Router navigation model](https://docs.flutter.dev/ui/navigation)
to synchronize browser locations with its view state.

In a larger application, wire `onTrade`, `onScores`, inspection and setup callbacks
to the application's router, sheets or dialogs. The same feature view can be a
route body or an embedded pane. Authentication redirects, membership checks,
deep-link recovery and persisted session restoration are application concerns.
Do not interpret a navigation pop as a game-action cancellation or withdrawal.

## Validation and remaining work

Use the [README commands](README.md#validate) to run host and
package tests and build web. The alternative-consumer test deliberately has no
imports from `cgms_demo` or package internals. It checks the dependency boundary;
it does not claim integration with an existing main application.

From `client`, `python3 tool/verify_assets.py` checks all 208 native
painting/source hashes, exact four-style inventory, shared back, font notices and package import boundary using
only Python's standard library. Run it after copying or relocating assets.

The online host now uses the typed authority adapter, account/session handling,
original-operation recovery and durable match navigation described above. Local
`#/play` examples still use simulated transitions. Consult the README's explicit
requirement/evidence table and browser manifest before extending acceptance claims.
The current localization catalog is English; other language catalogs, asset-rights
and release approval remain separate work. Native platform behavior and performance
remain the R06N gate and are not established by browser or widget tests.
