import 'package:cgms_demo/main.dart';
import 'package:cgms_demo/online/online_client.dart';
import 'package:cgms_demo/online/online_screen.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'online_screen_test.dart' show ViewClient;

void main() {
  testWidgets(
    'hovering a tooltip preserves the physical card semantic ancestor chain',
    (tester) async {
      tester.platformDispatcher.defaultRouteNameTestValue = '/online';
      addTearDown(tester.platformDispatcher.clearDefaultRouteNameTestValue);
      await tester.binding.setSurfaceSize(const Size(390, 844));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final semantics = tester.ensureSemantics();
      final client = ViewClient()
        ..publish({
          'cards': [
            {
              'card': {
                'id': 'tooltip-held-spade',
                'rank': 10,
                'suit': 'spades',
                'deck': 1,
              },
              'controller': 1,
              'zone': 'hand',
            },
          ],
        });
      addTearDown(client.dispose);
      await tester.pumpWidget(DemoApp(onlineClient: client));
      await tester.pumpAndSettle();
      final card = find.byKey(
        const ValueKey('adaptive-select-tooltip-held-spade'),
      );
      final control = find.byKey(const ValueKey('select-tooltip-held-spade'));
      final cardElement = tester.element(card);
      final focus = tester
          .widget<InkWell>(
            find.descendant(of: card, matching: find.byType(InkWell)).first,
          )
          .focusNode!;
      focus.requestFocus();
      await tester.pumpAndSettle();
      await tester.tap(card);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      List<int> ancestorIds() {
        var node = tester.getSemantics(control);
        final ids = <int>[node.id];
        while (node.parent != null) {
          node = node.parent!;
          ids.add(node.id);
        }
        return ids;
      }

      final before = ancestorIds();
      final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await mouse.addPointer(location: Offset.zero);
      await mouse.moveTo(tester.getCenter(find.byTooltip('Game options')));
      await tester.pumpAndSettle(const Duration(milliseconds: 700));
      expect(find.text('Game options'), findsOneWidget);
      expect(
        ancestorIds(),
        before,
        reason:
            'A tooltip must not reparent the route and reset native semantic scroll offsets.',
      );
      await mouse.moveTo(Offset.zero);
      await tester.pumpAndSettle(const Duration(milliseconds: 700));
      expect(ancestorIds(), before);
      expect(tester.element(card), same(cardElement));
      expect(FocusManager.instance.primaryFocus, same(focus));
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        {'tooltip-held-spade'},
      );
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
      await mouse.removePointer();
      semantics.dispose();
    },
  );

  testWidgets('initial online deep link never loads the hidden local deck', (
    tester,
  ) async {
    tester.platformDispatcher.defaultRouteNameTestValue =
        '/online?match=existing-match';
    addTearDown(tester.platformDispatcher.clearDefaultRouteNameTestValue);
    final client = _RouteClient();
    addTearDown(client.dispose);
    final assets = _RecordingAssets();
    await tester.pumpWidget(
      DefaultAssetBundle(
        bundle: assets,
        child: DemoApp(onlineClient: client),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.byType(OnlineScreen), findsOneWidget);
    expect(client.watched, ['existing-match']);
    expect(
      assets.artRequests,
      isEmpty,
      reason:
          'Neither an initial default frame nor the inactive local pane may request face artwork.',
    );
    final hidden = tester.widget<CgmsAdaptiveTable>(
      find.byKey(const ValueKey('approved-board'), skipOffstage: false),
    );
    expect(hidden.showArt, isFalse);
    await tester.pumpWidget(const SizedBox());
    expect(
      client.wasDisposed,
      isFalse,
      reason: 'The injected client belongs to its caller.',
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'online round trip retains local selection and artwork preference',
    (tester) async {
      await tester.binding.setSurfaceSize(const Size(1440, 900));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      final client = _RouteClient();
      addTearDown(client.dispose);
      await tester.pumpWidget(DemoApp(onlineClient: client));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Choose an action'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.byTooltip('Board actions'));
      await tester.tap(find.byTooltip('Board actions'));
      await tester.pumpAndSettle();
      await tester.tap(
        find.widgetWithText(PopupMenuItem<String>, 'Open Hearts'),
      );
      await tester.pumpAndSettle();
      final board = find.byKey(
        const ValueKey('approved-board'),
        skipOffstage: false,
      );
      final originalElement = tester.element(board);
      final original = tester.widget<CgmsAdaptiveTable>(board);
      expect(original.selectedCardIds, hasLength(7));
      expect(original.data.pending, isNull);
      expect(original.showArt, isTrue);
      await tester.binding.handlePushRoute('/online');
      await tester.pumpAndSettle();
      expect(tester.element(board), same(originalElement));
      expect(tester.widget<CgmsAdaptiveTable>(board).showArt, isFalse);
      expect(
        find.descendant(
          of: board,
          matching: find.byType(Image, skipOffstage: false),
          skipOffstage: false,
        ),
        findsNothing,
      );
      await tester.binding.handlePushRoute('/play');
      await tester.pumpAndSettle();
      final returned = tester.widget<CgmsAdaptiveTable>(board);
      expect(tester.element(board), same(originalElement));
      expect(returned.showArt, isTrue);
      expect(returned.selectedCardIds, original.selectedCardIds);
      expect(returned.data.gameId, original.data.gameId);
      expect(returned.data.pending, isNull);
      await tester.ensureVisible(find.byTooltip('Game menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Game menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Display settings'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(SwitchListTile, 'Card artwork'));
      await tester.tap(find.text('Done'));
      await tester.pumpAndSettle();
      await tester.binding.handlePushRoute('/online');
      await tester.pumpAndSettle();
      await tester.binding.handlePushRoute('/play');
      await tester.pumpAndSettle();
      expect(tester.widget<CgmsAdaptiveTable>(board).showArt, isFalse);
      expect(
        tester.widget<CgmsAdaptiveTable>(board).selectedCardIds,
        original.selectedCardIds,
      );
      expect(tester.takeException(), isNull);
    },
  );

  for (final legacy in [false, true]) {
    testWidgets(
      'inactive gallery retains its route without artwork, legacy=$legacy',
      (tester) async {
        tester.platformDispatcher.defaultRouteNameTestValue = '/gallery';
        addTearDown(tester.platformDispatcher.clearDefaultRouteNameTestValue);
        final client = _RouteClient();
        addTearDown(client.dispose);
        await tester.pumpWidget(DemoApp(legacy: legacy, onlineClient: client));
        await tester.pumpAndSettle();
        final galleryCard = find.byWidgetPredicate(
          (w) => w is CgmsCard && w.card.id == 'gallery-c7-1',
          skipOffstage: false,
        );
        final originalElement = tester.element(galleryCard);
        expect(tester.widget<CgmsCard>(galleryCard).showArt, isTrue);
        await tester.binding.handlePushRoute('/online');
        await tester.pumpAndSettle();
        expect(tester.element(galleryCard), same(originalElement));
        expect(tester.widget<CgmsCard>(galleryCard).showArt, isFalse);
        await tester.binding.handlePushRoute('/gallery');
        await tester.pumpAndSettle();
        expect(tester.element(galleryCard), same(originalElement));
        expect(tester.widget<CgmsCard>(galleryCard).showArt, isTrue);
        expect(tester.takeException(), isNull);
      },
    );
  }

  Future<void> chooseDesktop(WidgetTester tester) async {
    await tester.tap(find.byTooltip('Table layout'));
    await tester.pumpAndSettle();
    await tester.tap(
      find.widgetWithText(CheckedPopupMenuItem<String>, 'Desktop').last,
    );
    await tester.pumpAndSettle();
  }

  for (final (title, count) in [
    ('Ten cards in hand', 10),
    ('Fifteen cards in hand', 15),
  ]) {
    testWidgets(
      '$title loads a full table and preserves the hand on navigation',
      (tester) async {
        await tester.pumpWidget(const DemoApp(legacy: true));
        await tester.pumpAndSettle();
        await tester.tap(find.text('Scenarios'));
        await tester.pumpAndSettle();
        await tester.ensureVisible(find.text(title));
        await tester.tap(find.text(title));
        await tester.pumpAndSettle();
        await chooseDesktop(tester);
        var table = tester.widget<CgmsTableView>(find.byType(CgmsTableView));
        expect(table.data.hand, hasLength(count));
        expect(
          table.data.seats.every((seat) => seat.exposed.length >= 10),
          isTrue,
        );
        await tester.tap(find.text('Trade').first);
        await tester.pumpAndSettle();
        expect(find.text('Propose trade'), findsNothing);
        expect(find.text('Trade unavailable'), findsOneWidget);
        await tester.tap(find.text('Table').first);
        await tester.pumpAndSettle();
        table = tester.widget<CgmsTableView>(find.byType(CgmsTableView));
        expect(table.data.hand, hasLength(count));
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('setup validates name and enters a connected table', (
    tester,
  ) async {
    await tester.pumpWidget(const DemoApp(legacy: true));
    await tester.pumpAndSettle();
    expect(find.text('TAKE YOUR SEAT.'), findsOneWidget);
    await tester.enterText(find.byType(TextFormField), 'River');
    await tester.ensureVisible(find.text('Start guided game'));
    await tester.tap(find.text('Start guided game'));
    await tester.pumpAndSettle();
    await chooseDesktop(tester);
    expect(find.textContaining('River'), findsWidgets);
    expect(find.text('YOUR MOVE.'), findsOneWidget);
  });

  testWidgets(
    'empty projection explains unavailable trade instead of offering inert action',
    (tester) async {
      await tester.pumpWidget(const DemoApp(legacy: true));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Scenarios'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Nothing up your sleeve'));
      await tester.tap(find.text('Nothing up your sleeve'));
      await tester.pumpAndSettle();
      await chooseDesktop(tester);
      await tester.tap(find.text('Trade').first);
      await tester.pumpAndSettle();
      expect(find.text('Propose trade'), findsNothing);
      expect(find.text('Trade unavailable'), findsOneWidget);
    },
  );

  for (final width in [390.0, 768.0, 1440.0]) {
    testWidgets('setup reflows at $width and 200 percent text', (tester) async {
      tester.view.physicalSize = Size(width, 1024);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(const DemoApp(legacy: true, initialTextScale: 2));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.ensureVisible(find.text('Start guided game'));
      expect(find.text('Start guided game'), findsOneWidget);
    });
  }
}

class _RecordingAssets extends CachingAssetBundle {
  final artRequests = <String>[];
  @override
  Future<ByteData> load(String key) {
    if (key.contains('/assets/art/')) artRequests.add(key);
    return rootBundle.load(key);
  }
}

class _RouteClient extends OnlineClient {
  _RouteClient() : super(_RouteTransport()) {
    snapshot = AuthoritativeSnapshot.fromJson({
      'version': 1,
      'cursor': 1,
      'projection': {
        'board': {
          'game_id': 'online-game',
          'seat': 1,
          'phase': 'playing',
          'round': 1,
          'active': 1,
          'players': [],
          'cards': [],
        },
        'online': {'games_per_match': 3},
      },
    });
    connected = true;
  }
  final watched = <String>[];
  bool wasDisposed = false;
  @override
  Future<void> restoreSession() async {}
  @override
  Future<void> watchMatch(String id) async {
    watched.add(id);
  }

  @override
  void dispose() {
    wasDisposed = true;
    super.dispose();
  }
}

class _RouteTransport implements OnlineTransport {
  @override
  Future<OnlineResponse> request(
    String method,
    String path, {
    Map<String, dynamic>? body,
    Map<String, String> headers = const {},
  }) async => const OnlineResponse(401, {'code': 'UNAUTHENTICATED'});
  @override
  Stream<Map<String, dynamic>> events(String path) => const Stream.empty();
  @override
  void close() {}
}
