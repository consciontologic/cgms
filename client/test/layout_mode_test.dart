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
      await (FontLoader(entry.key)..addFont(
            rootBundle.load('packages/cgms_ui/assets/fonts/${entry.value}'),
          ))
          .load();
    }
  });

  Future<void> start(
    WidgetTester tester, {
    double width = 390,
    double height = 1024,
    double textScale = 1,
  }) async {
    tester.view.physicalSize = Size(width, height);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(DemoApp(legacy: true, initialTextScale: textScale));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Scenarios'));
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Combos at the table'));
    await tester.tap(find.text('Combos at the table'));
    await tester.pumpAndSettle();
  }

  TableViewData data(WidgetTester tester) =>
      find.byType(CgmsTableView).evaluate().isNotEmpty
      ? tester.widget<CgmsTableView>(find.byType(CgmsTableView)).data
      : tester.widget<CgmsTouchTableView>(find.byType(CgmsTouchTableView)).data;

  Object? route(WidgetTester tester) => tester
      .widget<MaterialApp>(find.byType(MaterialApp))
      .routerDelegate!
      .currentConfiguration;

  Future<void> choose(WidgetTester tester, String label) async {
    await tester.tap(find.byTooltip('Table layout'));
    await tester.pumpAndSettle();
    await tester.tap(
      find.widgetWithText(CheckedPopupMenuItem<String>, label).last,
    );
    await tester.pumpAndSettle();
  }

  Future<void> resize(WidgetTester tester, double width) async {
    tester.view.physicalSize = Size(width, 1024);
    await tester.pumpAndSettle();
  }

  void expectLayout(WidgetTester tester, String layout) {
    expect(find.byKey(ValueKey('demo-layout-$layout')), findsOneWidget);
    if (layout == 'desktop') {
      expect(find.byType(CgmsTableView), findsOneWidget);
      expect(find.byType(CgmsTouchTableView), findsNothing);
    } else {
      expect(find.byType(CgmsTouchTableView), findsOneWidget);
      expect(find.byType(CgmsTableView), findsNothing);
      if (layout == 'phone') {
        expect(find.byKey(const ValueKey('touch-tab-hand')), findsOneWidget);
      } else {
        expect(find.byKey(const ValueKey('touch-tab-hand')), findsNothing);
        final table = tester.getRect(
          find.byKey(const ValueKey('touch-public-scroll')),
        );
        final hand = tester.getRect(
          find.byKey(const ValueKey('touch-hand-scroll')),
        );
        expect(table.right, lessThanOrEqualTo(hand.left));
      }
    }
    expect(tester.takeException(), isNull);
  }

  testWidgets('Auto starts with the phone layout in a narrow browser', (
    tester,
  ) async {
    await start(tester);
    expect(find.byKey(const ValueKey('demo-layout-phone')), findsOneWidget);
    expect(find.byType(CgmsTouchTableView), findsOneWidget);
    expect(find.byKey(const ValueKey('touch-tab-hand')), findsOneWidget);
    expect(find.byTooltip('Table layout'), findsOneWidget);
    final picker = tester.getRect(find.byTooltip('Table layout'));
    expect(picker.height, greaterThanOrEqualTo(48));
    expect(picker.width, greaterThanOrEqualTo(48));
    expect(route(tester), '/play');
    expect(tester.takeException(), isNull);
  });

  testWidgets('Auto switches real layouts at both width boundaries', (
    tester,
  ) async {
    await start(tester);
    final before = data(tester);
    for (final (width, layout) in [
      (599.0, 'phone'),
      (600.0, 'tablet'),
      (1049.0, 'tablet'),
      (1050.0, 'desktop'),
      (1440.0, 'desktop'),
      (768.0, 'tablet'),
      (390.0, 'phone'),
    ]) {
      await resize(tester, width);
      expectLayout(tester, layout);
      expect(route(tester), '/play');
      expect(data(tester).gameId, before.gameId);
      expect(
        data(tester).hand.map((card) => card.id),
        before.hand.map((card) => card.id),
      );
      expect(data(tester).handCombos.single.title, 'Kidnapper');
    }
  });

  testWidgets(
    'manual modes persist across resize with distinct canvas widths',
    (tester) async {
      await start(tester, width: 1440);
      final gameId = data(tester).gameId;
      await choose(tester, 'Phone');
      for (final width in [1440.0, 768.0, 390.0]) {
        await resize(tester, width);
        expectLayout(tester, 'phone');
        expect(route(tester), '/phone');
        expect(
          tester.getSize(find.byType(CgmsTouchTableView)).width,
          lessThanOrEqualTo(440),
        );
        expect(data(tester).gameId, gameId);
      }
      await choose(tester, 'Tablet');
      for (final width in [390.0, 1440.0, 768.0]) {
        await resize(tester, width);
        expectLayout(tester, 'tablet');
        expect(route(tester), '/tablet');
        final canvasWidth = tester
            .getSize(find.byType(CgmsTouchTableView))
            .width;
        expect(canvasWidth, greaterThanOrEqualTo(768));
        expect(canvasWidth, lessThanOrEqualTo(1024));
        expect(data(tester).gameId, gameId);
        if (width == 390) {
          final preview = find.byKey(const ValueKey('tablet-preview-scroll'));
          final scroll = tester
              .widget<SingleChildScrollView>(preview)
              .controller!;
          expect(scroll.position.maxScrollExtent, greaterThan(0));
          await tester.drag(preview, const Offset(-350, 0));
          await tester.pumpAndSettle();
          expect(scroll.offset, greaterThan(0));
          expect(data(tester).gameId, gameId);
        }
      }
      await choose(tester, 'Desktop');
      for (final width in [1440.0, 390.0]) {
        await resize(tester, width);
        expectLayout(tester, 'desktop');
        expect(route(tester), '/table');
        expect(data(tester).gameId, gameId);
      }
    },
  );

  testWidgets(
    'selection and explicit pending responses survive every layout mode',
    (tester) async {
      await start(tester);
      final originalGame = data(tester).gameId;
      await tester.tap(find.byKey(const ValueKey('touch-tab-hand')));
      await tester.pumpAndSettle();
      final seven = find.byKey(const ValueKey('hand-card-deck-1-clubs-7'));
      await tester.ensureVisible(seven);
      final rect = tester.getRect(seven);
      await tester.tapAt(Offset(rect.center.dx, rect.top + 24));
      await tester.pump(const Duration(milliseconds: 350));
      await tester.pumpAndSettle();
      for (final label in ['Tablet', 'Desktop', 'Phone']) {
        await choose(tester, label);
        expect(data(tester).selectedCardId, 'deck-1-clubs-7');
        expect(data(tester).gameId, originalGame);
        expect(data(tester).hand, hasLength(15));
      }
      final primary = find.byKey(const ValueKey('touch-primary-action'));
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(data(tester).pending!.responseIndex, 0);
      for (final label in ['Desktop', 'Tablet', 'Auto · screen size']) {
        await choose(tester, label);
        expect(data(tester).pending!.responseIndex, 0);
        expect(data(tester).gameId, originalGame);
        expect(data(tester).hand, hasLength(15));
      }
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(data(tester).pending!.responseIndex, 1);
      await resize(tester, 768);
      expectLayout(tester, 'tablet');
      expect(data(tester).pending!.responseIndex, 1);
      await tester.tap(primary);
      await tester.pumpAndSettle();
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(data(tester).pending, isNull);
      expect(data(tester).hand, hasLength(14));
      expect(data(tester).seats.first.exposed, hasLength(13));
      expect(data(tester).handCombos.single.title, 'Kidnapper');
    },
  );

  testWidgets(
    'Table navigation and scenario changes remember the manual mode',
    (tester) async {
      await start(tester, width: 1440);
      await choose(tester, 'Tablet');
      final before = data(tester);
      await tester.tap(find.byKey(const ValueKey('touch-scores')));
      await tester.pumpAndSettle();
      expect(find.byType(CgmsScoresView), findsOneWidget);
      await tester.tap(find.text('Table').first);
      await tester.pumpAndSettle();
      expect(route(tester), '/tablet');
      expect(data(tester).gameId, before.gameId);
      await tester.tap(find.byTooltip('Demo menu'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Scenarios'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('An answer is owed'));
      await tester.tap(find.text('An answer is owed'));
      await tester.pumpAndSettle();
      expectLayout(tester, 'tablet');
      expect(route(tester), '/tablet');
      expect(data(tester).gameId, isNot(before.gameId));
      expect(data(tester).pending!.responseIndex, 0);
    },
  );

  testWidgets(
    'restored browser routes retain state and the legacy touch alias',
    (tester) async {
      await start(tester, width: 1440);
      final gameId = data(tester).gameId;
      for (final (path, layout) in [
        ('/phone', 'phone'),
        ('/tablet', 'tablet'),
        ('/table', 'desktop'),
        ('/play', 'desktop'),
      ]) {
        await tester.binding.handlePushRoute(path);
        await tester.pumpAndSettle();
        expect(route(tester), path);
        expectLayout(tester, layout);
        expect(data(tester).gameId, gameId);
        expect(data(tester).hand, hasLength(15));
      }
      await tester.binding.handlePushRoute('/touch');
      await tester.pumpAndSettle();
      expect(route(tester), '/touch');
      expect(find.byType(CgmsTouchTableView), findsOneWidget);
      expect(data(tester).gameId, gameId);
      await resize(tester, 390);
      expect(find.byKey(const ValueKey('touch-tab-hand')), findsOneWidget);
      expect(route(tester), '/touch');
      expect(data(tester).hand, hasLength(15));
    },
  );

  for (final (width, height, scale, tablet) in [
    (390.0, 844.0, 2.0, false),
    (768.0, 390.0, 1.0, false),
    (390.0, 844.0, 2.0, true),
  ]) {
    testWidgets(
      'layout fallback at $width x $height scale $scale manualTablet $tablet remains usable',
      (tester) async {
        await start(tester, width: width, height: height, textScale: scale);
        if (tablet) await choose(tester, 'Tablet');
        expect(tester.takeException(), isNull);
        final hand = find.byKey(const ValueKey('touch-tab-hand'));
        await tester.ensureVisible(hand);
        await tester.tap(hand);
        await tester.pumpAndSettle();
        final card = find.byKey(const ValueKey('hand-card-deck-1-clubs-7'));
        await tester.ensureVisible(card);
        final rect = tester.getRect(card);
        await tester.tapAt(Offset(rect.center.dx, rect.top + 24));
        await tester.pump(const Duration(milliseconds: 350));
        await tester.pumpAndSettle();
        expect(data(tester).selectedCardId, 'deck-1-clubs-7');
        final primary = find.byKey(const ValueKey('touch-primary-action'));
        await tester.ensureVisible(primary);
        await tester.tap(primary);
        await tester.pumpAndSettle();
        expect(data(tester).pending!.responseIndex, 0);
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('layout menu explains every mode and Escape preserves the game', (
    tester,
  ) async {
    await start(tester);
    final before = data(tester);
    await tester.tap(find.byTooltip('Table layout'));
    await tester.pumpAndSettle();
    for (final label in ['Auto · screen size', 'Phone', 'Tablet', 'Desktop']) {
      expect(find.text(label), findsWidgets);
    }
    await tester.sendKeyEvent(LogicalKeyboardKey.escape);
    await tester.pumpAndSettle();
    expect(find.byType(CheckedPopupMenuItem<String>), findsNothing);
    expect(route(tester), '/play');
    expect(data(tester).gameId, before.gameId);
    expect(
      data(tester).hand.map((card) => card.id),
      before.hand.map((card) => card.id),
    );
  });
}
