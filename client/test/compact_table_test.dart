import 'package:cgms_demo/main.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  Future<void> chooseDesktop(WidgetTester tester) async {
    await tester.tap(find.byTooltip('Table layout'));
    await tester.pumpAndSettle();
    await tester.tap(
      find.widgetWithText(CheckedPopupMenuItem<String>, 'Desktop').last,
    );
    await tester.pumpAndSettle();
  }

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
  testWidgets(
    'combo scenario shows public formations and a private hand pattern',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(const DemoApp(legacy: true));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Scenarios'));
      await tester.pumpAndSettle();
      expect(find.text('Combos at the table'), findsOneWidget);
      await tester.ensureVisible(find.text('Combos at the table'));
      await tester.tap(find.text('Combos at the table'));
      await tester.pumpAndSettle();
      await chooseDesktop(tester);
      final table = tester
          .widget<CgmsTableView>(find.byType(CgmsTableView))
          .data;
      expect(table.hand, hasLength(15));
      expect(table.handCombos.single.title, 'Kidnapper');
      final publicCombos = table.seats.expand((seat) => seat.combos).toList();
      expect(
        publicCombos.map((combo) => combo.title),
        containsAll(['Great People', 'Fate']),
      );
      for (final combo in publicCombos) {
        expect(find.text(combo.title), findsOneWidget);
        for (final card in combo.cards) {
          expect(
            find.byWidgetPredicate(
              (w) => w is CgmsCard && w.card.id == card.id,
            ),
            findsOneWidget,
          );
        }
      }
      for (final combo in [...publicCombos, ...table.handCombos]) {
        final button = find.byKey(ValueKey('inspect-combo-${combo.id}'));
        await tester.ensureVisible(button);
        await tester.tap(button);
        await tester.pumpAndSettle();
        expect(find.byType(AlertDialog), findsOneWidget);
        expect(find.text(combo.description), findsOneWidget);
        for (final card in combo.cards) {
          expect(
            find.descendant(
              of: find.byType(AlertDialog),
              matching: find.byWidgetPredicate(
                (w) => w is CgmsCard && w.card.id == card.id,
              ),
            ),
            findsOneWidget,
          );
        }
        await tester.tap(find.text('Close inspection'));
        await tester.pumpAndSettle();
        expect(
          tester
              .widget<CgmsTableView>(find.byType(CgmsTableView))
              .data
              .hand
              .map((c) => c.id),
          table.hand.map((c) => c.id),
        );
      }
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    'public suit groups show full artwork and magnify every physical card',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(const DemoApp(legacy: true));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Scenarios'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Fifteen cards in hand'));
      await tester.tap(find.text('Fifteen cards in hand'));
      await tester.pumpAndSettle();
      await chooseDesktop(tester);

      final data = tester
          .widget<CgmsTableView>(find.byType(CgmsTableView))
          .data;
      for (final seat in data.seats) {
        final groupOrigins = <Offset>{};
        for (final suit in CardSuit.values) {
          final cards = seat.exposed
              .where((card) => card.suit == suit)
              .toList();
          expect(cards, hasLength(3));
          final positions = [
            for (final card in cards)
              tester.getRect(find.byKey(ValueKey('public-card-${card.id}'))),
          ];
          groupOrigins.add(positions.first.topLeft);
          for (var index = 1; index < positions.length; index++) {
            final previous = positions[index - 1];
            final current = positions[index];
            expect(current.left, previous.left, reason: '$suit shares a stack');
            expect(current.top - previous.top, greaterThanOrEqualTo(48));
            expect(current.top, greaterThanOrEqualTo(previous.bottom));
            expect(find.text('Copy ${cards[index - 1].copy}'), findsNothing);
          }
        }
        expect(
          groupOrigins,
          hasLength(4),
          reason: 'Each suit has its own distinct group',
        );
      }

      // Every full painting remains reachable and independently magnifiable.
      for (final seat in data.seats) {
        for (final card in seat.exposed) {
          final target = find.byKey(ValueKey('public-card-${card.id}'));
          await tester.ensureVisible(target);
          await tester.pumpAndSettle();
          await tester.tap(target);
          await tester.pump(const Duration(milliseconds: 50));
          await tester.tap(target);
          await tester.pumpAndSettle();
          expect(find.byType(CgmsCardInspector), findsOneWidget);
          final inspected = tester
              .widget<CgmsCardInspector>(find.byType(CgmsCardInspector))
              .card;
          expect(inspected.id, card.id);
          expect(inspected.rank, card.rank);
          expect(inspected.suit, card.suit);
          expect(inspected.copy, card.copy);
          await tester.tap(find.text('Close inspection'));
          await tester.pumpAndSettle();
        }
      }
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    '15 private and 48 full public cards remain reachable with persistent actions',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(const DemoApp(legacy: true));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Scenarios'));
      await tester.pumpAndSettle();
      await tester.ensureVisible(find.text('Fifteen cards in hand'));
      await tester.tap(find.text('Fifteen cards in hand'));
      await tester.pumpAndSettle();
      await chooseDesktop(tester);

      final cards = find.byType(CgmsCard);
      expect(cards, findsNWidgets(63));
      for (final element in cards.evaluate()) {
        final target = find.byWidget(element.widget);
        await tester.ensureVisible(target);
        await tester.pumpAndSettle();
        final rect = tester.getRect(target);
        expect(rect.top, greaterThanOrEqualTo(0));
        expect(rect.bottom, lessThanOrEqualTo(900));
        expect(rect.left, greaterThanOrEqualTo(0));
        expect(rect.right, lessThanOrEqualTo(1440));
      }
      final primary = find.byKey(const ValueKey('hand-primary-action'));
      expect(tester.getRect(primary).bottom, lessThanOrEqualTo(900));

      final seven = find.byKey(const ValueKey('hand-card-deck-1-clubs-7'));
      await tester.ensureVisible(seven);
      await tester.pumpAndSettle();
      final first = tester.getRect(cards.first);
      final handPosition = tester.getRect(seven);
      final handCards = tester
          .widget<CgmsTableView>(find.byType(CgmsTableView))
          .data
          .hand;
      final positions = {
        for (final card in handCards)
          card.id: tester.getRect(find.byKey(ValueKey('hand-card-${card.id}'))),
      };
      for (final suit in CardSuit.values) {
        final group = handCards.where((card) => card.suit == suit).toList();
        for (var index = 1; index < group.length; index++) {
          final before = positions[group[index - 1].id]!;
          final after = positions[group[index].id]!;
          expect(after.left, before.left);
          expect(after.top - before.top, greaterThanOrEqualTo(48));
          expect(after.top, greaterThanOrEqualTo(before.bottom));
        }
      }
      await tester.tapAt(Offset(handPosition.center.dx, handPosition.top + 24));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(tester.getRect(seven), handPosition);
      for (final card in handCards) {
        expect(
          tester.getRect(find.byKey(ValueKey('hand-card-${card.id}'))),
          positions[card.id],
        );
      }
      expect(tester.getRect(cards.first), first);
      expect(tester.getRect(primary).bottom, lessThanOrEqualTo(900));
      await tester.ensureVisible(primary);
      await tester.pumpAndSettle();
      await tester.tap(primary);
      await tester.pumpAndSettle();
      for (var response = 0; response < 3; response++) {
        await tester.ensureVisible(primary);
        await tester.pumpAndSettle();
        await tester.tap(primary);
        await tester.pumpAndSettle();
      }
      final table = tester.widget<CgmsTableView>(find.byType(CgmsTableView));
      expect(table.data.hand, hasLength(14));
      expect(
        table.data.seats.singleWhere((s) => s.isSelf).exposed,
        hasLength(13),
      );
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    'an unavailable own card can still be inspected in compact mode',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(const DemoApp(legacy: true));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Start guided game'));
      await tester.pumpAndSettle();
      await chooseDesktop(tester);
      final diamond = tester.getRect(
        find.byKey(const ValueKey('public-card-deck-1-diamonds-3')),
      );
      await tester.tapAt(Offset(diamond.center.dx, diamond.top + 24));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      final table = tester.widget<CgmsTableView>(find.byType(CgmsTableView));
      expect(table.data.selectedCardId, isNull);
      expect(find.byType(CgmsCardInspector), findsNothing);
      final card = find.byKey(const ValueKey('public-card-deck-1-diamonds-3'));
      await tester.ensureVisible(card);
      await tester.pumpAndSettle();
      await tester.tap(card);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(card);
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCardInspector), findsOneWidget);
      expect(
        tester
            .widget<CgmsCardInspector>(find.byType(CgmsCardInspector))
            .card
            .id,
        'deck-1-diamonds-3',
      );
      await tester.tap(find.text('Close inspection'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Display settings'));
      await tester.pumpAndSettle();
      await tester.tap(find.widgetWithText(SwitchListTile, 'Compact table'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Done'));
      await tester.pumpAndSettle();
      final comfortable = tester.widget<CgmsTableView>(
        find.byType(CgmsTableView),
      );
      expect(comfortable.compact, isFalse);
      expect(comfortable.data.gameId, table.data.gameId);
      expect(
        comfortable.data.hand.map((c) => c.id),
        table.data.hand.map((c) => c.id),
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('tablet public board scrolls while the hand stays in place', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(768, 1024);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(const DemoApp(legacy: true));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Scenarios'));
    await tester.pumpAndSettle();
    await tester.ensureVisible(find.text('Fifteen cards in hand'));
    await tester.tap(find.text('Fifteen cards in hand'));
    await tester.pumpAndSettle();
    await chooseDesktop(tester);
    final board = find.byKey(const ValueKey('compact-public-scroll'));
    final hand = find.byType(CgmsHandView);
    final before = tester.getRect(hand);
    expect(before.top, lessThan(700));
    final scroll = tester.widget<SingleChildScrollView>(board).controller!;
    expect(scroll.position.maxScrollExtent, greaterThan(0));
    await tester.drag(board, const Offset(0, -650));
    await tester.pumpAndSettle();
    expect(scroll.offset, greaterThan(0));
    expect(tester.getRect(hand), before);
    expect(tester.takeException(), isNull);
  });
}
