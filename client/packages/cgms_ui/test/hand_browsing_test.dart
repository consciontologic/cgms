import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

final _hand = List<VisibleCard>.generate(
  15,
  (index) => VisibleCard(
    id: 'hand-$index',
    rank: index == 4 ? '2' : '${2 + index ~/ 4}',
    suit: CardSuit.values[index % 4],
    copy: index == 4 ? 2 : 1,
  ),
);

Widget _shell(
  Widget child, {
  double textScale = 1,
  EdgeInsets padding = const EdgeInsets.all(16),
}) => MaterialApp(
  theme: CgmsTheme.dark(),
  localizationsDelegates: CgmsLocalizations.localizationsDelegates,
  supportedLocales: CgmsLocalizations.supportedLocales,
  home: Scaffold(
    body: MediaQuery(
      data: MediaQueryData(
        textScaler: TextScaler.linear(textScale),
        disableAnimations: true,
      ),
      child: SingleChildScrollView(padding: padding, child: child),
    ),
  ),
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUpAll(() async {
    // Measure compact rows with the real package fonts instead of Ahem.
    for (final font in {
      'Barlow Condensed': 'BarlowCondensed-Bold.ttf',
      'Atkinson Hyperlegible Next': 'AtkinsonHyperlegibleNext-Regular.ttf',
    }.entries) {
      await (FontLoader('packages/cgms_ui/${font.key}')..addFont(
            rootBundle.load('packages/cgms_ui/assets/fonts/${font.value}'),
          ))
          .load();
    }
  });

  for (final compact in [false, true]) {
    testWidgets(
      'hand separates selection and double-tap magnification compact=$compact',
      (tester) async {
        final selected = <String>[];
        final inspected = <VisibleCard>[];
        await tester.pumpWidget(
          _shell(
            CgmsHandView(
              cards: [_hand.first],
              compact: compact,
              showArt: false,
              onSelect: selected.add,
              onInspect: inspected.add,
            ),
          ),
        );
        final target = find.byKey(ValueKey('hand-card-${_hand.first.id}'));
        await tester.tap(target);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        expect(selected, [_hand.first.id]);
        expect(inspected, isEmpty);
        await tester.tap(target);
        await tester.pump(const Duration(milliseconds: 50));
        await tester.tap(target);
        await tester.pumpAndSettle();
        expect(inspected, [_hand.first]);
        expect(selected, [_hand.first.id]);
        expect(find.byIcon(Icons.zoom_in), findsNothing);
      },
    );
  }

  testWidgets(
    'magnified card hides physical-copy labels while preserving semantics',
    (tester) async {
      final semantics = tester.ensureSemantics();
      await tester.pumpWidget(
        _shell(CgmsCardInspector(card: _hand[4], showArt: false)),
      );
      expect(
        find.textContaining(RegExp('copy', caseSensitive: false)),
        findsNothing,
      );
      expect(find.bySemanticsLabel('2 of Hearts, copy 2'), findsWidgets);
      semantics.dispose();
    },
  );

  testWidgets(
    'suit stacks preserve full paintings and keyboard activates each copy',
    (tester) async {
      final semantics = tester.ensureSemantics();
      const first = VisibleCard(
        id: 'consumer-hearts-copy-one',
        rank: '8',
        suit: CardSuit.hearts,
        copy: 1,
      );
      const second = VisibleCard(
        id: 'consumer-hearts-copy-two',
        rank: '8',
        suit: CardSuit.hearts,
        copy: 2,
      );
      const club = VisibleCard(
        id: 'consumer-club',
        rank: '3',
        suit: CardSuit.clubs,
        copy: 1,
      );
      final activated = <VisibleCard>[];
      await tester.pumpWidget(
        _shell(
          CgmsSuitStacks(
            cards: const [first, club, second],
            onTap: activated.add,
            showArt: false,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCard), findsNWidgets(3));
      final firstRect = tester.getRect(
        find.byKey(const ValueKey('public-card-consumer-hearts-copy-one')),
      );
      final secondRect = tester.getRect(
        find.byKey(const ValueKey('public-card-consumer-hearts-copy-two')),
      );
      expect(firstRect.left, secondRect.left);
      expect(secondRect.top, greaterThan(firstRect.top));
      expect(secondRect.top, greaterThanOrEqualTo(firstRect.bottom));
      expect(
        find.textContaining(RegExp('copy', caseSensitive: false)),
        findsNothing,
      );
      final coveredSemantics = tester.getSemantics(
        find.byKey(const ValueKey('select-consumer-hearts-copy-one')),
      );
      expect(coveredSemantics.label, '8 of Hearts, copy 1');
      expect(coveredSemantics.rect.height, greaterThanOrEqualTo(48));
      final art = tester.getRect(
        find.byKey(const ValueKey('card-art-consumer-hearts-copy-one')),
      );
      expect(art.bottom, lessThanOrEqualTo(secondRect.top));
      expect(art.width / art.height, closeTo(first.artworkAspectRatio, .001));

      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.pump();
      expect(
        tester
            .widget<InkWell>(
              find.descendant(
                of: find.byKey(
                  const ValueKey('public-card-consumer-hearts-copy-one'),
                ),
                matching: find.byType(InkWell),
              ),
            )
            .focusNode!
            .hasFocus,
        isTrue,
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(activated, [first]);
      final secondInk = tester.widget<InkWell>(
        find.descendant(
          of: find.byKey(const ValueKey('select-consumer-hearts-copy-two')),
          matching: find.byType(InkWell),
        ),
      );
      for (var i = 0; i < 6 && !secondInk.focusNode!.hasFocus; i++) {
        await tester.sendKeyEvent(LogicalKeyboardKey.tab);
        await tester.pump();
      }
      expect(secondInk.focusNode!.hasFocus, isTrue);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(activated, [first, second]);

      // Each copy has an independent full-card pointer target.
      await tester.tapAt(Offset(secondRect.center.dx, secondRect.top + 24));
      expect(activated, [first, second, second]);
      expect(tester.takeException(), isNull);
      semantics.dispose();
    },
  );

  testWidgets(
    'stack headers, long statuses and selection survive 200 percent text',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      const status =
          'Waiting for review from a player with a long display name';
      const cards = [
        VisibleCard(id: 'plain', rank: '10', suit: CardSuit.hearts, copy: 1),
        VisibleCard(
          id: 'status',
          rank: '10',
          suit: CardSuit.hearts,
          copy: 2,
          status: status,
        ),
        VisibleCard(id: 'selected', rank: '7', suit: CardSuit.hearts, copy: 1),
        VisibleCard(id: 'last', rank: '8', suit: CardSuit.hearts, copy: 1),
      ];
      await tester.pumpWidget(
        _shell(
          CgmsSuitStacks(
            cards: cards,
            selectedCardId: 'selected',
            onTap: (_) {},
            showArt: false,
          ),
          textScale: 2,
        ),
      );
      await tester.pumpAndSettle();
      Finder card(String id) => find.byKey(ValueKey('public-card-$id'));
      final plainArt = find.byKey(const ValueKey('card-art-plain'));
      final statusRect = tester.getRect(card('status'));
      final selectedRect = tester.getRect(card('selected'));
      final lastRect = tester.getRect(card('last'));
      expect(
        tester.getRect(plainArt).bottom,
        lessThanOrEqualTo(statusRect.top),
      );
      expect(statusRect.bottom, lessThanOrEqualTo(selectedRect.top));
      expect(
        tester.getRect(find.text(status)).bottom,
        lessThan(selectedRect.top),
      );
      expect(selectedRect.bottom, lessThanOrEqualTo(lastRect.top));
      final selectedMark = find.descendant(
        of: card('selected'),
        matching: find.byIcon(Icons.check),
      );
      expect(selectedMark, findsOneWidget);
      expect(tester.getRect(selectedMark).bottom, lessThan(lastRect.top));
      expect(tester.widget<CgmsCard>(card('selected')).selected, isTrue);
      await tester.ensureVisible(selectedMark);
      await tester.pumpAndSettle();
      for (var index = 0; index < 3; index++) {
        await tester.sendKeyEvent(LogicalKeyboardKey.tab);
        await tester.pump();
      }
      expect(
        tester
            .widget<InkWell>(
              find.descendant(
                of: card('selected'),
                matching: find.byType(InkWell),
              ),
            )
            .focusNode!
            .hasFocus,
        isTrue,
      );
      final header = tester.getRect(
        find.byKey(const ValueKey('card-header-selected')),
      );
      final art = tester.getRect(
        find.byKey(const ValueKey('card-art-selected')),
      );
      expect(header.bottom, lessThanOrEqualTo(art.top));
      expect(
        art.width / art.height,
        closeTo(cards[2].artworkAspectRatio, .001),
      );
      expect(tester.widget<Icon>(selectedMark).color, CgmsColors.cyan);
      expect(tester.widget<CgmsCard>(card('selected')).selected, isTrue);
      expect(tester.getRect(selectedMark).top, greaterThanOrEqualTo(0));
      expect(tester.getRect(selectedMark).bottom, lessThanOrEqualTo(844));
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('empty suit stacks expose no cards or actions', (tester) async {
    var actions = 0;
    await tester.pumpWidget(
      _shell(
        CgmsSuitStacks(
          cards: const [],
          onTap: (_) => actions++,
          showArt: false,
        ),
      ),
    );
    expect(find.byType(CgmsCard), findsNothing);
    expect(find.byType(InkWell), findsNothing);
    expect(actions, 0);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'compact hand double-tap inspects each physical copy without a select adapter',
    (tester) async {
      VisibleCard? inspected;
      await tester.pumpWidget(
        _shell(
          CgmsHandView(
            cards: _hand,
            compact: true,
            onInspect: (card) => inspected = card,
            showArt: false,
          ),
        ),
      );
      final copy = find.byKey(const ValueKey('hand-card-hand-4'));
      await tester.ensureVisible(copy);
      await tester.tap(copy);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(inspected, isNull);
      await _doubleTap(tester, copy);
      expect(inspected?.id, 'hand-4');
      expect(inspected?.copy, 2);
      expect(
        find.byKey(const ValueKey('hand-selection-summary')),
        findsNothing,
      );
      expect(tester.takeException(), isNull);
    },
  );

  for (final textScale in [1.0, 2.0]) {
    testWidgets(
      'compact hand stacks fifteen cards without moving selection at $textScale text',
      (tester) async {
        tester.view.physicalSize = const Size(1200, 844);
        tester.view.devicePixelRatio = 1;
        addTearDown(tester.view.resetPhysicalSize);
        addTearDown(tester.view.resetDevicePixelRatio);
        String? selected;
        await tester.pumpWidget(
          _shell(
            StatefulBuilder(
              builder: (context, setState) => CgmsHandView(
                cards: _hand,
                compact: true,
                selectedCardId: selected,
                onSelect: (id) => setState(() => selected = id),
                onClearSelection: () => setState(() => selected = null),
                onInspect: (_) {},
                showArt: false,
              ),
            ),
            padding: EdgeInsets.zero,
            textScale: textScale,
          ),
        );
        await tester.pumpAndSettle();
        final initialRects = [
          for (final card in _hand)
            tester.getRect(find.byKey(ValueKey('hand-card-${card.id}'))),
        ];
        expect(
          initialRects.take(4).map((rect) => rect.left).toSet(),
          hasLength(4),
        );
        for (var index = 4; index < _hand.length; index++) {
          expect(initialRects[index].left, initialRects[index - 4].left);
          expect(
            initialRects[index].top,
            greaterThan(initialRects[index - 4].top),
          );
          expect(
            initialRects[index].top,
            greaterThanOrEqualTo(initialRects[index - 4].bottom - .000001),
          );
        }
        expect(initialRects.last.right, lessThanOrEqualTo(1200));
        expect(find.text('Showing 15 of 15 cards'), findsOneWidget);
        for (final rect in initialRects) {
          expect(rect.width, greaterThanOrEqualTo(48));
          expect(rect.height, greaterThanOrEqualTo(48));
          if (textScale == 1) {
            expect(rect.width, 108);
          }
        }
        await tester.tapAt(
          Offset(initialRects[4].center.dx, initialRects[4].top + 24),
        );
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        expect(selected, 'hand-4');
        expect(find.text('Selected: 2 Hearts'), findsOneWidget);
        for (var index = 0; index < _hand.length; index++) {
          expect(
            tester.getRect(
              find.byKey(ValueKey('hand-card-${_hand[index].id}')),
            ),
            initialRects[index],
          );
        }
        final selectedMark = find.descendant(
          of: find.byKey(const ValueKey('hand-card-hand-4')),
          matching: find.byIcon(Icons.check),
        );
        expect(selectedMark, findsOneWidget);
        expect(
          tester.getRect(selectedMark).bottom,
          lessThanOrEqualTo(initialRects[8].top),
        );
        expect(
          tester
              .getTopLeft(find.byKey(const ValueKey('hand-selection-summary')))
              .dy,
          greaterThanOrEqualTo(
            initialRects
                .map((rect) => rect.bottom)
                .reduce((a, b) => a > b ? a : b),
          ),
        );
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets(
    'compact hand retains filtered selection controls and accessible filters at large text',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      String? selected = 'hand-4';
      VisibleCard? inspected;
      await tester.pumpWidget(
        _shell(
          StatefulBuilder(
            builder: (context, setState) => CgmsHandView(
              cards: _hand,
              compact: true,
              selectedCardId: selected,
              onSelect: (id) => setState(() => selected = id),
              onClearSelection: () => setState(() => selected = null),
              onInspect: (card) => inspected = card,
              showArt: false,
            ),
          ),
          textScale: 2,
        ),
      );
      final filter = find.byKey(const ValueKey('hand-filter-spades'));
      expect(tester.getSize(filter).height, greaterThanOrEqualTo(48));
      await tester.ensureVisible(filter);
      await tester.tap(filter);
      await tester.pumpAndSettle();
      expect(find.text('Showing 3 of 15 cards'), findsOneWidget);
      expect(find.byKey(const ValueKey('hand-card-hand-4')), findsNothing);
      expect(find.byKey(const ValueKey('select-hand-4')), findsOneWidget);
      expect(
        find.text('Your selection is outside this filter.'),
        findsOneWidget,
      );
      final inspect = find.byKey(const ValueKey('inspect-hand-selection'));
      await tester.ensureVisible(inspect);
      await _doubleTap(tester, inspect);
      expect(inspected?.id, 'hand-4');
      final clear = find.byKey(const ValueKey('clear-hand-selection'));
      await tester.ensureVisible(clear);
      await tester.tap(clear);
      await tester.pumpAndSettle();
      expect(selected, isNull);
      expect(
        find.byKey(const ValueKey('hand-selection-summary')),
        findsNothing,
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'hand filters preserve physical copies and show exact visible count',
    (tester) async {
      await tester.pumpWidget(
        _shell(CgmsHandView(cards: _hand, onInspect: (_) {}, showArt: false)),
      );
      expect(find.text('Showing 15 of 15 cards'), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('hand-filter-hearts')));
      await tester.pump();
      expect(find.text('Showing 4 of 15 cards'), findsOneWidget);
      expect(find.byType(CgmsCard), findsNWidgets(4));
      expect(find.byKey(const ValueKey('select-hand-0')), findsOneWidget);
      expect(find.byKey(const ValueKey('select-hand-4')), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('hand-filter-all')));
      await tester.pump();
      expect(find.byType(CgmsCard), findsNWidgets(15));
    },
  );

  testWidgets(
    'selected card remains inspectable and clearable behind another suit filter',
    (tester) async {
      String? selected;
      VisibleCard? inspected;
      await tester.pumpWidget(
        _shell(
          StatefulBuilder(
            builder: (context, setState) => CgmsHandView(
              cards: _hand,
              selectedCardId: selected,
              onSelect: (id) => setState(() => selected = id),
              onClearSelection: () => setState(() => selected = null),
              onInspect: (card) => inspected = card,
              showArt: false,
            ),
          ),
        ),
      );
      await tester.tap(find.byKey(const ValueKey('select-hand-0')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(selected, 'hand-0');
      await tester.tap(find.byKey(const ValueKey('hand-filter-spades')));
      await tester.pump();
      expect(find.byKey(const ValueKey('hand-card-hand-0')), findsNothing);
      expect(find.byKey(const ValueKey('select-hand-0')), findsOneWidget);
      expect(find.text('Selected: 2 Hearts'), findsOneWidget);
      expect(
        find.text('Your selection is outside this filter.'),
        findsOneWidget,
      );
      await _doubleTap(
        tester,
        find.byKey(const ValueKey('inspect-hand-selection')),
      );
      expect(inspected?.id, 'hand-0');
      await tester.tap(find.byKey(const ValueKey('clear-hand-selection')));
      await tester.pump();
      expect(selected, isNull);
      expect(find.text('Selected: 2 Hearts'), findsNothing);
    },
  );

  testWidgets('keyboard can choose a filter and select a physical card', (
    tester,
  ) async {
    String? selected;
    await tester.pumpWidget(
      _shell(
        CgmsHandView(
          cards: _hand,
          onSelect: (id) => selected = id,
          onInspect: (_) {},
          showArt: false,
        ),
      ),
    );
    await tester.sendKeyEvent(LogicalKeyboardKey.tab);
    await tester.sendKeyEvent(LogicalKeyboardKey.tab);
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    await tester.pump();
    expect(find.text('Showing 4 of 15 cards'), findsOneWidget);
    final card = find.byKey(const ValueKey('select-hand-0'));
    tester
        .widget<InkWell>(
          find.descendant(of: card, matching: find.byType(InkWell)),
        )
        .focusNode!
        .requestFocus();
    await tester.pump();
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    expect(selected, 'hand-0');
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'dense public groups count supplied faces and keep every copy reachable',
    (tester) async {
      final cards = _hand.take(12).toList();
      await tester.pumpWidget(
        _shell(
          CgmsExposedCardsView(cards: cards, showArt: false, onInspect: (_) {}),
        ),
      );
      expect(find.text('Hearts · 3 cards'), findsOneWidget);
      expect(find.text('Diamonds · 3 cards'), findsOneWidget);
      expect(find.byType(CgmsCard), findsNWidgets(12));
      expect(find.byKey(const ValueKey('select-hand-0')), findsOneWidget);
      expect(find.byKey(const ValueKey('select-hand-4')), findsOneWidget);
    },
  );

  testWidgets(
    '390px hand at 200 percent has usable filters, empty state and no overflow',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(
        _shell(
          CgmsHandView(
            cards: _hand.where((card) => card.suit != CardSuit.spades).toList(),
            selectedCardId: 'hand-0',
            onSelect: (_) {},
            onClearSelection: () {},
            onInspect: (_) {},
            showArt: false,
          ),
          textScale: 2,
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.ensureVisible(
        find.byKey(const ValueKey('hand-filter-spades')),
      );
      await tester.tap(find.byKey(const ValueKey('hand-filter-spades')));
      await tester.pumpAndSettle();
      expect(find.text('Showing 0 of 12 cards'), findsOneWidget);
      expect(find.text('No cards match this filter.'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  for (final width in [390.0, 1200.0]) {
    for (final scale in [1.0, 2.0]) {
      testWidgets(
        'dense suit group geometry at $width pixels and $scale text',
        (tester) async {
          tester.view.physicalSize = Size(width, 844);
          tester.view.devicePixelRatio = 1;
          addTearDown(tester.view.resetPhysicalSize);
          addTearDown(tester.view.resetDevicePixelRatio);
          await tester.pumpWidget(
            _shell(
              CgmsExposedCardsView(
                cards: _hand.take(12).toList(),
                cardWidth: 96,
                showArt: false,
                onInspect: (_) {},
              ),
              textScale: scale,
            ),
          );
          await tester.pumpAndSettle();
          final hearts = tester.getTopLeft(find.text('Hearts · 3 cards'));
          final diamonds = tester.getTopLeft(find.text('Diamonds · 3 cards'));
          final clubs = tester.getTopLeft(find.text('Clubs · 3 cards'));
          if (width > 1000) {
            expect(diamonds.dy, closeTo(hearts.dy, 0.01));
            expect(diamonds.dx, greaterThan(hearts.dx));
            expect(clubs.dx, hearts.dx);
            expect(clubs.dy, greaterThan(hearts.dy));
          } else {
            expect(diamonds.dx, hearts.dx);
            expect(diamonds.dy, greaterThan(hearts.dy));
            expect(clubs.dy, greaterThan(diamonds.dy));
          }
          expect(find.byType(CgmsCard), findsNWidgets(12));
          expect(tester.getSize(find.byType(CgmsCard).first).width, 112);
          expect(tester.takeException(), isNull);
        },
      );
    }
  }

  testWidgets(
    'table jumps to hand without animation and resets its filter for a new game',
    (tester) async {
      var actions = 0;
      Widget table(String gameId) => CgmsTableView(
        data: _tableData(gameId),
        showArt: false,
        onSelect: (_) {},
        onSelectHand: (_) {},
        onClearHandSelection: () {},
        onInspect: (_) {},
        onPrimary: () => actions++,
        onTrade: () {},
        onScores: () {},
        onEndTurn: () {},
        onReconnect: () {},
        endTurnEnabled: false,
        endTurnExplanation: 'The adapter has no end-turn action in this state.',
      );
      await tester.pumpWidget(_shell(table('game-a')));
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      await tester.pump();
      expect(
        Scrollable.of(
          tester.element(find.byType(CgmsHandView)),
        ).position.pixels,
        greaterThan(0),
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.pump();
      expect(
        FocusManager.instance.primaryFocus?.context
            ?.findAncestorWidgetOfExactType<ChoiceChip>()
            ?.key,
        const ValueKey('hand-filter-all'),
      );
      await tester.tap(find.byKey(const ValueKey('hand-filter-hearts')));
      await tester.pump();
      expect(find.text('Showing 4 of 15 cards'), findsOneWidget);
      await tester.ensureVisible(
        find.byKey(const ValueKey('hand-primary-action')),
      );
      await tester.tap(find.byKey(const ValueKey('hand-primary-action')));
      expect(actions, 1);
      final endTurn = find.widgetWithText(OutlinedButton, 'End turn & score');
      expect(tester.widget<OutlinedButton>(endTurn).onPressed, isNull);
      expect(
        find.text('The adapter has no end-turn action in this state.'),
        findsOneWidget,
      );
      await tester.pumpWidget(_shell(table('game-b')));
      await tester.pump();
      expect(find.text('Showing 15 of 15 cards'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );
}

TableViewData _tableData(String gameId) => TableViewData(
  gameId: gameId,
  round: 4,
  totalRounds: 13,
  turnLabel: 'A supplied turn label',
  seats: [
    SeatView(
      id: 'viewer',
      name: 'Viewer',
      number: 1,
      isSelf: true,
      exposed: _hand
          .take(12)
          .map(
            (card) => VisibleCard(
              id: 'public-${card.id}',
              rank: card.rank,
              suit: card.suit,
              copy: card.copy,
            ),
          )
          .toList(),
      concealed: const ConcealedCards(handCount: 15, aceCount: 1),
      finances: FinancialView(
        matchScore: ExactPoints.integer(0),
        gameScore: ExactPoints.integer(0),
        available: ExactPoints.integer(0),
        debt: ExactPoints.integer(0),
      ),
      status: 'Public cards supplied by the adapter',
    ),
  ],
  hand: _hand,
  events: [],
  objective: 'Review a supplied hand',
  actionLabel: 'Confirm supplied action',
  actionExplanation: 'The adapter decides what this action does.',
  actionEnabled: true,
  phaseLabel: 'Ready',
);

Future<void> _doubleTap(WidgetTester tester, Finder target) async {
  await tester.ensureVisible(target);
  await tester.pumpAndSettle();
  await tester.tap(target);
  await tester.pump(const Duration(milliseconds: 50));
  await tester.tap(target);
  await tester.pumpAndSettle();
}
