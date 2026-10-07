import 'package:cgms_demo/main.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUpAll(() async {
    for (final entry in {
      'packages/cgms_ui/Barlow Condensed': 'BarlowCondensed-Bold.ttf',
      'packages/cgms_ui/Atkinson Hyperlegible Next':
          'AtkinsonHyperlegibleNext-Regular.ttf',
    }.entries) {
      final loader = FontLoader(entry.key)
        ..addFont(
          rootBundle.load('packages/cgms_ui/assets/fonts/${entry.value}'),
        );
      await loader.load();
    }
  });

  Future<void> loadScenario(
    WidgetTester tester, {
    String title = 'Combos at the table',
    Size size = const Size(390, 844),
    double textScale = 1,
  }) async {
    tester.view.physicalSize = size;
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(DemoApp(legacy: true, initialTextScale: textScale));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Scenarios'));
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text(title));
    await tester.tap(find.text(title));
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Table layout'));
    await tester.pumpAndSettle();
    await tester.tap(
      find.widgetWithText(CheckedPopupMenuItem<String>, 'Desktop').last,
    );
    await tester.pumpAndSettle();
  }

  Future<void> openTouch(WidgetTester tester) async {
    final viewport = tester.view.physicalSize / tester.view.devicePixelRatio;
    final mode = viewport.width >= 700 && viewport.height >= 620
        ? 'Tablet'
        : 'Phone';
    await tester.tap(find.byTooltip('Table layout'));
    await tester.pumpAndSettle();
    await tester.tap(
      find.widgetWithText(CheckedPopupMenuItem<String>, mode).last,
    );
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('touch-table-view')), findsOneWidget);
  }

  TableViewData touchData(WidgetTester tester) =>
      tester.widget<CgmsTouchTableView>(find.byType(CgmsTouchTableView)).data;

  Future<void> openTab(WidgetTester tester, String name) async {
    final tab = find.byKey(ValueKey('touch-tab-$name'));
    await tester.ensureVisible(tab);
    await tester.tap(tab);
    await tester.pumpAndSettle();
  }

  Future<void> selectSeven(WidgetTester tester) async {
    final card = find.byKey(const ValueKey('hand-card-deck-1-clubs-7'));
    await tester.ensureVisible(card);
    final rect = tester.getRect(card);
    await tester.tapAt(Offset(rect.center.dx, rect.top + 24));
    await tester.pump(const Duration(milliseconds: 350));
    await tester.pumpAndSettle();
    expect(touchData(tester).selectedCardId, 'deck-1-clubs-7');
  }

  testWidgets(
    'Touch layout retains the combo scenario when switching layouts',
    (tester) async {
      await loadScenario(tester);
      final before = tester
          .widget<CgmsTableView>(find.byType(CgmsTableView))
          .data;
      expect(before.hand, hasLength(15));
      expect(before.handCombos.single.title, 'Kidnapper');
      expect(find.byTooltip('Table layout'), findsOneWidget);
      await openTouch(tester);
      expect(find.byType(CgmsTableView), findsNothing);
      expect(touchData(tester).gameId, before.gameId);
      expect(touchData(tester).hand, hasLength(15));
      expect(touchData(tester).handCombos.single.title, 'Kidnapper');
      await tester.tap(find.byTooltip('Table layout'));
      await tester.pumpAndSettle();
      await tester.tap(
        find.widgetWithText(CheckedPopupMenuItem<String>, 'Desktop').last,
      );
      await tester.pumpAndSettle();
      final after = tester
          .widget<CgmsTableView>(find.byType(CgmsTableView))
          .data;
      expect(after.gameId, before.gameId);
      expect(
        after.hand.map((card) => card.id),
        before.hand.map((card) => card.id),
      );
      expect(after.handCombos.single.title, 'Kidnapper');
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('phone tabs retain selection and resolve the hand addition', (
    tester,
  ) async {
    await loadScenario(tester);
    await openTouch(tester);
    await openTab(tester, 'activity');
    expect(find.textContaining('Public combo review:'), findsOneWidget);
    await openTab(tester, 'hand');
    await selectSeven(tester);
    await openTab(tester, 'table');
    expect(touchData(tester).selectedCardId, 'deck-1-clubs-7');
    await openTab(tester, 'hand');
    final primary = find.byKey(const ValueKey('touch-primary-action'));
    expect(tester.getRect(primary).bottom, lessThanOrEqualTo(844));
    await tester.tap(primary);
    await tester.pumpAndSettle();
    expect(touchData(tester).pending!.responders, ['Mara', 'Ivo', 'Nia']);
    for (var response = 0; response < 3; response++) {
      expect(touchData(tester).pending!.responseIndex, response);
      expect(touchData(tester).hand, hasLength(15));
      await tester.tap(primary);
      await tester.pumpAndSettle();
    }
    final after = touchData(tester);
    expect(after.pending, isNull);
    expect(after.hand, hasLength(14));
    expect(after.seats.first.exposed, hasLength(13));
    expect(after.handCombos.single.title, 'Kidnapper');
    expect(after.seats.first.finances.gameScore, ExactPoints.integer(0));
    expect(tester.takeException(), isNull);
  });

  testWidgets('phone seat picker opens the correct public combo inspector', (
    tester,
  ) async {
    await loadScenario(tester);
    await openTouch(tester);
    final before = touchData(tester);
    await tester.tap(find.byKey(const ValueKey('touch-seat-mara')));
    await tester.pumpAndSettle();
    final inspect = find.byKey(
      const ValueKey('inspect-combo-mara-great-people'),
    );
    await tester.ensureVisible(inspect);
    await tester.tap(inspect);
    await tester.pumpAndSettle();
    expect(find.byType(AlertDialog), findsOneWidget);
    expect(
      find.text(before.seats[1].combos.single.description),
      findsOneWidget,
    );
    final dialogCards = tester.widgetList<CgmsCard>(
      find.descendant(
        of: find.byType(AlertDialog),
        matching: find.byType(CgmsCard),
      ),
    );
    expect(dialogCards.map((card) => card.card.id), [
      'deck-1-hearts-Q',
      'deck-2-hearts-Q',
    ]);
    await tester.tap(find.text('Close inspection'));
    await tester.pumpAndSettle();
    expect(
      touchData(tester).hand.map((card) => card.id),
      before.hand.map((card) => card.id),
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'disconnected touch response remains pending until an explicit pass',
    (tester) async {
      await loadScenario(tester, title: 'A quiet connection');
      await openTouch(tester);
      expect(touchData(tester).connection, ConnectionStateView.disconnected);
      expect(touchData(tester).pending!.responseIndex, 0);
      final primary = find.byKey(const ValueKey('touch-primary-action'));
      final button = find.descendant(
        of: primary,
        matching: find.byWidgetPredicate(
          (widget) => widget is ButtonStyleButton,
        ),
        matchRoot: true,
      );
      expect(tester.widget<ButtonStyleButton>(button).onPressed, isNull);
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(touchData(tester).pending!.responseIndex, 0);
      final reconnect = find.text('Restore connection');
      await tester.ensureVisible(reconnect);
      await tester.tap(reconnect);
      await tester.pumpAndSettle();
      expect(touchData(tester).connection, ConnectionStateView.connected);
      expect(touchData(tester).pending!.responseIndex, 0);
      expect(tester.widget<ButtonStyleButton>(button).onPressed, isNotNull);
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(touchData(tester).pending!.responseIndex, 1);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('tablet keeps the private hand beside the selected public seat', (
    tester,
  ) async {
    await loadScenario(tester, size: const Size(768, 1024));
    await openTouch(tester);
    final board = find.byKey(const ValueKey('touch-public-scroll'));
    final hand = find.byKey(const ValueKey('touch-hand-scroll'));
    final boardRect = tester.getRect(board);
    final handRect = tester.getRect(hand);
    expect(boardRect.right, lessThanOrEqualTo(handRect.left));
    expect(boardRect.bottom, lessThanOrEqualTo(1024));
    expect(handRect.bottom, lessThanOrEqualTo(1024));
    await selectSeven(tester);
    final selectedHandRect = tester.getRect(hand);
    await tester.tap(find.byKey(const ValueKey('touch-seat-mara')));
    await tester.pumpAndSettle();
    expect(find.text('Great People'), findsOneWidget);
    expect(tester.getRect(hand), selectedHandRect);
    expect(touchData(tester).selectedCardId, 'deck-1-clubs-7');
    expect(touchData(tester).hand, hasLength(15));
    expect(tester.takeException(), isNull);
  });

  testWidgets('scores and scenario navigation return to the touch table', (
    tester,
  ) async {
    await loadScenario(tester, title: 'An answer is owed');
    await openTouch(tester);
    final before = touchData(tester);
    await tester.tap(find.byKey(const ValueKey('touch-scores')));
    await tester.pumpAndSettle();
    expect(find.byType(CgmsScoresView), findsOneWidget);
    final back = find.text('Back to the table');
    await tester.ensureVisible(back);
    await tester.tap(back);
    await tester.pumpAndSettle();
    expect(touchData(tester).gameId, before.gameId);
    expect(
      touchData(tester).pending!.responseIndex,
      before.pending!.responseIndex,
    );
    await tester.tap(find.byTooltip('Demo menu'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Scenarios'));
    await tester.pumpAndSettle();
    await openTouch(tester);
    expect(touchData(tester).gameId, before.gameId);
    expect(
      touchData(tester).pending!.responseIndex,
      before.pending!.responseIndex,
    );
    await tester.tap(find.byTooltip('Demo menu'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Scenarios'));
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('An answer is owed'));
    await tester.tap(find.text('An answer is owed'));
    await tester.pumpAndSettle();
    expect(touchData(tester).gameId, isNot(before.gameId));
    expect(touchData(tester).pending!.responseIndex, 0);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'phone at 200 percent text keeps tabs and the legal action reachable',
    (tester) async {
      await loadScenario(tester, textScale: 2);
      await openTouch(tester);
      for (final tab in ['activity', 'table', 'hand']) {
        await openTab(tester, tab);
        expect(tester.takeException(), isNull);
      }
      await selectSeven(tester);
      final primary = find.byKey(const ValueKey('touch-primary-action'));
      await tester.ensureVisible(primary);
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(touchData(tester).pending!.responseIndex, 0);
      expect(tester.takeException(), isNull);
    },
  );

  for (final size in [const Size(320, 568), const Size(844, 390)]) {
    testWidgets(
      'short touch viewport $size keeps the primary action reachable',
      (tester) async {
        await loadScenario(tester, size: size);
        await openTouch(tester);
        expect(tester.takeException(), isNull);
        await openTab(tester, 'hand');
        await selectSeven(tester);
        final primary = find.byKey(const ValueKey('touch-primary-action'));
        await tester.ensureVisible(primary);
        await tester.tap(primary);
        await tester.pumpAndSettle();
        expect(touchData(tester).pending!.responseIndex, 0);
        expect(tester.takeException(), isNull);
      },
    );
  }
}
