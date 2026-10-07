import 'dart:io';
import 'dart:ui' as ui;

import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

VisibleCard card(String rank, CardSuit suit, [int copy = 1]) => VisibleCard(
  id: '${suit.name}-$rank-$copy',
  rank: rank,
  suit: suit,
  copy: copy,
);

final hand = [
  card('Q', CardSuit.clubs),
  card('Q', CardSuit.diamonds),
  card('Q', CardSuit.hearts),
  card('J', CardSuit.hearts),
  card('J', CardSuit.diamonds),
  card('K', CardSuit.spades, 2),
  for (final rank in ['2', '3', '4', '5', '7', '9', '10'])
    card(rank, CardSuit.hearts),
  for (final rank in ['2', '4', '6', '9', '10']) card(rank, CardSuit.spades),
];
final underground = ComboView(
  id: 'underground',
  title: 'Underground',
  stateLabel: '5 kings',
  description: 'An exposed formation supplied by the adapter.',
  cards: [
    card('K', CardSuit.hearts),
    card('K', CardSuit.hearts, 2),
    card('K', CardSuit.clubs),
    card('K', CardSuit.diamonds),
    card('K', CardSuit.spades),
  ],
);
final ponzi = ComboView(
  id: 'ponzi',
  title: 'Ponzi',
  stateLabel: '4 jacks',
  description: 'An exposed formation supplied by the adapter.',
  cards: [
    card('J', CardSuit.clubs),
    card('J', CardSuit.clubs, 2),
    card('J', CardSuit.spades),
    card('J', CardSuit.spades, 2),
  ],
);

TableViewData fixture({
  ConnectionStateView connection = ConnectionStateView.connected,
  PendingView? pending,
  String gameId = 'game-2',
}) {
  SeatView seat(
    int number,
    String name,
    List<VisibleCard> exposed, {
    List<ComboView> combos = const [],
  }) => SeatView(
    id: 'seat-$number',
    number: number,
    name: name,
    exposed: exposed,
    combos: combos,
    isSelf: number == 1,
    status: '',
    concealed: ConcealedCards(handCount: number == 1 ? 18 : 17, aceCount: 1),
    finances: FinancialView(
      matchScore: ExactPoints.integer(22),
      gameScore: ExactPoints.integer(6),
      available: ExactPoints.integer(6),
      debt: ExactPoints.integer(0),
    ),
  );
  return TableViewData(
    gameId: gameId,
    round: 7,
    totalRounds: 13,
    turnLabel: 'Your turn',
    seats: [
      seat(1, 'Alice', [
        card('10', CardSuit.diamonds),
        card('8', CardSuit.diamonds),
        card('8', CardSuit.clubs),
      ]),
      seat(2, 'Bob', [
        card('9', CardSuit.diamonds),
        card('7', CardSuit.diamonds),
        card('8', CardSuit.hearts),
        card('6', CardSuit.hearts),
      ]),
      seat(
        3,
        'Carol',
        [
          card('8', CardSuit.diamonds, 2),
          card('6', CardSuit.diamonds),
          ...underground.cards,
        ],
        combos: [underground],
      ),
      seat(
        4,
        'Deniz',
        [
          card('7', CardSuit.diamonds, 2),
          card('5', CardSuit.diamonds),
          card('6', CardSuit.clubs),
          ...ponzi.cards,
        ],
        combos: [ponzi],
      ),
    ],
    hand: hand,
    events: const [],
    objective: 'Choose an action',
    actionLabel: 'Declare attack',
    actionExplanation: '8 Clubs matches 8 Hearts.',
    phaseLabel: 'Your turn',
    actionEnabled: true,
    connection: connection,
    pending: pending,
  );
}

Widget shell(Widget child, {double scale = 1}) => MaterialApp(
  theme: CgmsTheme.dark(),
  localizationsDelegates: CgmsLocalizations.localizationsDelegates,
  supportedLocales: CgmsLocalizations.supportedLocales,
  builder: (context, child) => MediaQuery(
    data: MediaQuery.of(context).copyWith(textScaler: TextScaler.linear(scale)),
    child: child!,
  ),
  home: Scaffold(
    body: RepaintBoundary(
      key: const ValueKey('capture-board'),
      child: ColoredBox(color: CgmsColors.canvas, child: child),
    ),
  ),
);

Widget board({
  TableViewData? data,
  ValueChanged<String>? selected,
  ValueChanged<VisibleCard>? inspected,
  VoidCallback? primary,
  Set<String> selection = const {},
}) => CgmsActionBoard(
  data: data ?? fixture(),
  showArt: const bool.fromEnvironment('CGMS_CAPTURE_BOARD'),
  selectedCardIds: selection,
  onSelect: selected ?? (_) {},
  onSelectHand: selected,
  onInspect: inspected ?? (_) {},
  onPrimary: primary ?? () {},
  onTrade: () {},
  onScores: () {},
  onEndTurn: () {},
  onReconnect: () {},
  actionPanel: const SizedBox(height: 56, child: Text('Choose an action')),
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUpAll(() async {
    // Measure the bundled product typography, rather than Flutter's square
    // test font, when asserting how many illustrated cards fit the viewport.
    for (final entry in {
      'Barlow Condensed': ['BarlowCondensed-Bold.ttf'],
      'Atkinson Hyperlegible Next': [
        'AtkinsonHyperlegibleNext-Regular.ttf',
        'AtkinsonHyperlegibleNext-Bold.ttf',
      ],
    }.entries) {
      final loader = FontLoader('packages/cgms_ui/${entry.key}');
      for (final filename in entry.value) {
        loader.addFont(
          rootBundle.load('packages/cgms_ui/assets/fonts/$filename'),
        );
      }
      await loader.load();
    }
    await (FontLoader(
      'MaterialIcons',
    )..addFont(rootBundle.load('fonts/MaterialIcons-Regular.otf'))).load();
  });
  for (final size in [
    const Size(390, 844),
    const Size(768, 1024),
    const Size(1440, 900),
  ]) {
    testWidgets(
      'all public piles and complete flat hand remain together at $size',
      (tester) async {
        tester.view.physicalSize = size;
        tester.view.devicePixelRatio = 1;
        addTearDown(tester.view.resetPhysicalSize);
        addTearDown(tester.view.resetDevicePixelRatio);
        await tester.pumpWidget(shell(board()));
        await tester.pumpAndSettle();
        if (const bool.fromEnvironment('CGMS_CAPTURE_BOARD')) {
          await tester.runAsync(() async {
            final context = tester.element(find.byType(CgmsActionBoard));
            final all = [...hand, ...fixture().seats.expand((s) => s.exposed)];
            for (final card in all) {
              await precacheImage(
                AssetImage(
                  'assets/art/${card.suit.name}-${card.rank}.png',
                  package: 'cgms_ui',
                ),
                context,
              );
            }
          });
          await tester.pumpAndSettle();
          final boundary = tester.renderObject<RenderRepaintBoundary>(
            find.byKey(const ValueKey('capture-board')),
          );
          await tester.runAsync(() async {
            final image = await boundary.toImage();
            final bytes = await image.toByteData(
              format: ui.ImageByteFormat.png,
            );
            final name = size.width < 600
                ? 'phone'
                : size.width < 1000
                ? 'tablet'
                : 'desktop';
            final output = Directory(
              '../../.browser-tools/artifacts/widget-renders',
            );
            await output.create(recursive: true);
            await File(
              '${output.path}/cgmsart-$name.png',
            ).writeAsBytes(bytes!.buffer.asUint8List());
            image.dispose();
          });
        }
        expect(tester.takeException(), isNull);
        for (final seat in fixture().seats) {
          expect(
            find.byKey(ValueKey('action-seat-${seat.id}')),
            findsOneWidget,
          );
          for (final value in seat.exposed) {
            expect(
              find.byKey(ValueKey('board-public-${value.id}')),
              findsOneWidget,
            );
          }
        }
        for (final value in hand) {
          final target = find.byKey(ValueKey('board-hand-${value.id}'));
          expect(target, findsOneWidget);
          expect(tester.getRect(target).right, lessThanOrEqualTo(size.width));
          await tester.ensureVisible(target);
          await tester.pumpAndSettle();
          // Scroll translation can leave sub-pixel floating point residue.
          expect(tester.getRect(target).top, greaterThanOrEqualTo(-.000001));
          expect(
            tester.getRect(target).bottom,
            lessThanOrEqualTo(size.height + .000001),
          );
        }
        expect(
          tester
              .getTopLeft(find.byKey(const ValueKey('action-seat-seat-1')))
              .dy,
          tester
              .getTopLeft(find.byKey(const ValueKey('action-seat-seat-2')))
              .dy,
        );
        expect(find.byType(TabBar), findsNothing);
        expect(find.byType(ChoiceChip), findsNothing);
      },
    );
  }

  testWidgets(
    'multi selection retains all seven Hearts and every physical card',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final selected = hand
          .where(
            (c) => c.suit == CardSuit.hearts && int.tryParse(c.rank) != null,
          )
          .map((c) => c.id)
          .toSet();
      String? submitted;
      await tester.pumpWidget(
        shell(board(selection: selected, selected: (id) => submitted = id)),
      );
      final cards = tester.widgetList<CgmsCard>(find.byType(CgmsCard));
      expect(
        cards.where((c) => c.selected).map((c) => c.card.id).toSet(),
        selected,
      );
      final target = find.byKey(ValueKey('board-hand-${hand.first.id}'));
      await tester.ensureVisible(target);
      await tester.tap(target);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(submitted, hand.first.id);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('keyboard inspection does not submit a thumbnail selection', (
    tester,
  ) async {
    String? selected;
    VisibleCard? inspected;
    await tester.pumpWidget(
      shell(
        board(
          selected: (id) => selected = id,
          inspected: (card) => inspected = card,
        ),
      ),
    );
    final target = find.byKey(ValueKey('board-hand-${hand.first.id}'));
    await tester.ensureVisible(target);
    final ink = tester.widget<InkWell>(
      find.descendant(of: target, matching: find.byType(InkWell)).first,
    );
    ink.focusNode!.requestFocus();
    await tester.pump();
    await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
    expect(inspected?.id, hand.first.id);
    expect(selected, isNull);
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    expect(selected, hand.first.id);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'keyboard inspects a single public card without selecting its attack',
    (tester) async {
      String? selected;
      VisibleCard? inspected;
      await tester.pumpWidget(
        shell(
          board(
            selected: (id) => selected = id,
            inspected: (c) => inspected = c,
          ),
        ),
      );
      final target = find.byKey(const ValueKey('board-public-clubs-8-1'));
      await tester.ensureVisible(target);
      await tester.tap(target);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(selected, 'clubs-8-1');
      selected = null;
      tester
          .widget<InkWell>(
            find.descendant(of: target, matching: find.byType(InkWell)).first,
          )
          .focusNode!
          .requestFocus();
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
      expect(inspected?.id, 'clubs-8-1');
      expect(selected, isNull);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'complete physical copies remain separately selectable and double-tap inspectable',
    (tester) async {
      String? chosen;
      VisibleCard? inspected;
      await tester.pumpWidget(
        shell(
          board(selected: (id) => chosen = id, inspected: (c) => inspected = c),
        ),
      );
      final first = find.byKey(
        ValueKey('board-public-${underground.cards[0].id}'),
      );
      final target = find.byKey(
        ValueKey('board-public-${underground.cards[1].id}'),
      );
      expect(first, findsOneWidget);
      expect(target, findsOneWidget);
      expect(
        tester.getRect(first).bottom,
        lessThanOrEqualTo(tester.getRect(target).top),
      );
      await tester.ensureVisible(target);
      await tester.tap(target);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(chosen, underground.cards[1].id);
      expect(inspected, isNull);
      chosen = null;
      await tester.tap(target);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(target);
      await tester.pumpAndSettle();
      expect(inspected?.id, underground.cards[1].id);
      expect(chosen, isNull);
      expect(find.byIcon(Icons.zoom_in), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'disconnection and pending responses preserve inspection without selection',
    (tester) async {
      var selections = 0;
      VisibleCard? inspected;
      await tester.pumpWidget(
        shell(
          board(
            data: fixture(
              connection: ConnectionStateView.disconnected,
              pending: const PendingView(
                title: 'Attack pending',
                detail: 'Waiting for Bob',
                responders: ['Bob', 'Carol', 'Deniz'],
                responseIndex: 0,
              ),
            ),
            selected: (_) => selections++,
            inspected: (c) => inspected = c,
          ),
        ),
      );
      await tester.ensureVisible(
        find.byKey(ValueKey('board-hand-${hand.first.id}')),
      );
      await tester.tap(find.byKey(ValueKey('board-hand-${hand.first.id}')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(selections, 0);
      expect(inspected, isNull);
      await tester.tap(find.byKey(ValueKey('board-hand-${hand.first.id}')));
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(find.byKey(ValueKey('board-hand-${hand.first.id}')));
      await tester.pumpAndSettle();
      expect(inspected?.id, hand.first.id);
      expect(selections, 0);
      expect(find.text('Attack pending'), findsOneWidget);
      expect(find.text('Waiting for Bob'), findsOneWidget);
      expect(tester.takeException(), isNull);
    },
  );

  for (final size in [
    const Size(390, 844),
    const Size(390, 480),
    const Size(768, 600),
  ]) {
    testWidgets('large text and short windows retain every identity at $size', (
      tester,
    ) async {
      tester.view.physicalSize = size;
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      await tester.pumpWidget(shell(board(), scale: 2));
      expect(tester.takeException(), isNull);
      for (final value in hand) {
        final target = find.byKey(ValueKey('board-hand-${value.id}'));
        expect(target, findsOneWidget);
        await tester.ensureVisible(target);
        await tester.pump();
        expect(tester.getRect(target).left, greaterThanOrEqualTo(0));
        expect(tester.getRect(target).right, lessThanOrEqualTo(size.width));
      }
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      expect(tester.takeException(), isNull);
    });
  }
}
