import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

const _public = VisibleCard(
  id: 'public-copy-2',
  rank: '8',
  suit: CardSuit.clubs,
  copy: 2,
);
final _hand = List.generate(
  15,
  (i) => VisibleCard(
    id: 'private-$i',
    rank: '${2 + i % 9}',
    suit: CardSuit.hearts,
    copy: 1 + i ~/ 9,
  ),
);

TableViewData _data({
  String gameId = 'consumer-one',
  String? selected,
  ConnectionStateView connection = ConnectionStateView.connected,
  bool empty = false,
  bool actionEnabled = true,
  bool densePublic = false,
  int eventCount = 1,
  PendingView? pending,
}) => TableViewData(
  gameId: gameId,
  round: 4,
  totalRounds: 13,
  turnLabel: 'Ada has the turn',
  seats: empty
      ? []
      : [
          for (var i = 0; i < 3; i++)
            SeatView(
              id: 'seat-$i',
              name: i == 0 ? 'Ada' : 'Player with a longer name $i',
              number: i + 1,
              isSelf: i == 0,
              exposed: [
                if (i == 0) _public,
                if (densePublic)
                  for (var rank = 2; rank <= 10; rank++)
                    VisibleCard(
                      id: 'seat-$i-public-$rank',
                      rank: '$rank',
                      suit: i == 0
                          ? CardSuit.clubs
                          : i == 1
                          ? CardSuit.diamonds
                          : CardSuit.spades,
                      copy: 1,
                    ),
              ],
              concealed: ConcealedCards(
                handCount: i == 0 ? 15 : 7,
                aceCount: 1,
              ),
              finances: FinancialView(
                matchScore: ExactPoints.integer(0),
                gameScore: ExactPoints.integer(0),
                available: ExactPoints.integer(0),
                debt: ExactPoints.integer(0),
              ),
              status: 'Public status supplied by consumer',
            ),
        ],
  hand: empty ? [] : _hand,
  selectedCardId: selected,
  events: [
    for (var i = 0; i < eventCount; i++)
      i == 0
          ? 'The consumer supplied this event.'
          : 'Earlier consumer event $i.',
  ],
  objective: 'Review the next move',
  actionLabel: pending == null
      ? 'Confirm supplied move'
      : 'Resolve supplied response',
  actionExplanation: 'The consumer decides what this action means.',
  actionEnabled: actionEnabled,
  phaseLabel: 'Ready',
  connection: connection,
  pending: pending,
);

Widget _shell(Widget child, {double scale = 1}) => MaterialApp(
  theme: CgmsTheme.dark(),
  localizationsDelegates: CgmsLocalizations.localizationsDelegates,
  supportedLocales: CgmsLocalizations.supportedLocales,
  home: MediaQuery(
    data: MediaQueryData(
      textScaler: TextScaler.linear(scale),
      disableAnimations: true,
    ),
    child: Scaffold(body: child),
  ),
);

CgmsTouchTableView _view(
  TableViewData data, {
  ValueChanged<String>? select,
  ValueChanged<VisibleCard>? inspect,
  VoidCallback? primary,
  VoidCallback? reconnect,
}) => CgmsTouchTableView(
  data: data,
  showArt: false,
  onSelect: select ?? (_) {},
  onInspect: inspect ?? (_) {},
  onSelectHand: select,
  onClearHandSelection: () {},
  onPrimary: primary ?? () {},
  onTrade: () {},
  onScores: () {},
  onEndTurn: () {},
  onReconnect: reconnect ?? () {},
  endTurnEnabled: false,
  endTurnDisabledReason: 'The supplied turn cannot finish.',
);

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUpAll(() async {
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

  void size(WidgetTester tester, Size value) {
    tester.view.physicalSize = value;
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
  }

  testWidgets('phone preserves hand filter and keeps the move dock visible', (
    tester,
  ) async {
    size(tester, const Size(390, 788));
    String? selected;
    var moves = 0;
    await tester.pumpWidget(
      _shell(
        StatefulBuilder(
          builder: (context, setState) => _view(
            _data(selected: selected),
            select: (id) => setState(() => selected = id),
            primary: () => moves++,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();
    final action = find.byKey(const ValueKey('touch-primary-action'));
    final dockPosition = tester.getRect(action);
    expect(dockPosition.bottom, lessThanOrEqualTo(788));
    await tester.tap(find.byKey(const ValueKey('touch-tab-hand')));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('hand-filter-hearts')));
    await tester.pumpAndSettle();
    final first = tester.getRect(
      find.byKey(const ValueKey('hand-card-private-0')),
    );
    await tester.tapAt(Offset(first.center.dx, first.top + 24));
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(selected, 'private-0');
    expect(tester.getRect(action).bottom, dockPosition.bottom);
    await tester.drag(
      find.byKey(const ValueKey('touch-hand-scroll')),
      const Offset(0, -350),
    );
    await tester.pumpAndSettle();
    expect(tester.getRect(action).bottom, dockPosition.bottom);
    await tester.tap(find.byKey(const ValueKey('touch-tab-activity')));
    await tester.pumpAndSettle();
    expect(find.text('The consumer supplied this event.'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('touch-tab-hand')));
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<ChoiceChip>(find.byKey(const ValueKey('hand-filter-hearts')))
          .selected,
      isTrue,
    );
    await tester.tap(action);
    expect(moves, 1);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'rejected public selection stays inspectable and new games reset local seat',
    (tester) async {
      size(tester, const Size(390, 788));
      VisibleCard? inspected;
      String? selected;
      var game = 'one';
      late StateSetter update;
      await tester.pumpWidget(
        _shell(
          StatefulBuilder(
            builder: (context, setState) {
              update = setState;
              return _view(
                _data(
                  gameId: game,
                  selected: game == 'one' ? 'private-0' : null,
                ),
                select: (id) => selected = id,
                inspect: (card) => inspected = card,
              );
            },
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('public-card-public-copy-2')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(selected, _public.id);
      final card = find.byKey(const ValueKey('public-card-public-copy-2'));
      await tester.tap(card);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(card);
      await tester.pumpAndSettle();
      expect(inspected, same(_public));
      final seatButton = tester.widget<OutlinedButton>(
        find.byKey(const ValueKey('touch-seat-seat-0')),
      );
      final focusBorder = seatButton.style!.side!.resolve({
        WidgetState.focused,
      })!;
      expect(focusBorder.color, CgmsColors.lime);
      expect(focusBorder.width, 3);
      await tester.tap(find.byKey(const ValueKey('touch-tab-hand')));
      await tester.pumpAndSettle();
      final opponent = find.byKey(const ValueKey('touch-seat-seat-1'));
      await tester.ensureVisible(opponent);
      await tester.tap(opponent);
      await tester.pumpAndSettle();
      expect(
        find.text('Player with a longer name 1 · exposed cards'),
        findsOneWidget,
      );
      update(() => game = 'two');
      await tester.pumpAndSettle();
      expect(find.text('Ada · exposed cards'), findsOneWidget);
      expect(find.byIcon(Icons.zoom_in), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'tablet independently scrolls its public position and private hand',
    (tester) async {
      size(tester, const Size(768, 968));
      await tester.pumpWidget(_shell(_view(_data(densePublic: true))));
      await tester.pumpAndSettle();
      final public = find.byKey(const ValueKey('touch-public-scroll'));
      final hand = find.byKey(const ValueKey('touch-hand-scroll'));
      expect(find.byKey(const ValueKey('touch-tab-hand')), findsNothing);
      expect(
        tester.getRect(public).right,
        lessThanOrEqualTo(tester.getRect(hand).left),
      );
      final before = tester.getRect(
        find.byKey(const ValueKey('public-card-public-copy-2')),
      );
      await tester.drag(hand, const Offset(0, -600));
      await tester.pumpAndSettle();
      expect(
        tester.getRect(find.byKey(const ValueKey('public-card-public-copy-2'))),
        before,
      );
      expect(
        tester.widget<SingleChildScrollView>(hand).controller!.offset,
        greaterThan(0),
      );
      final handOffset = tester
          .widget<SingleChildScrollView>(hand)
          .controller!
          .offset;
      await tester.drag(public, const Offset(0, -400));
      await tester.pumpAndSettle();
      expect(
        tester.widget<SingleChildScrollView>(public).controller!.offset,
        greaterThan(0),
      );
      await tester.tap(find.byKey(const ValueKey('touch-seat-seat-1')));
      await tester.pumpAndSettle();
      expect(
        tester.widget<SingleChildScrollView>(public).controller!.offset,
        0,
      );
      expect(
        tester.widget<SingleChildScrollView>(hand).controller!.offset,
        handOffset,
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'disconnected pending decisions stay visible and actions await reconnect',
    (tester) async {
      final semantics = tester.ensureSemantics();
      size(tester, const Size(390, 788));
      var reconnected = 0;
      await tester.pumpWidget(
        _shell(
          _view(
            _data(
              connection: ConnectionStateView.disconnected,
              pending: const PendingView(
                title: 'A response is waiting',
                detail: 'No response has resolved.',
                responders: ['Ada', 'Bea', 'Cem'],
                responseIndex: 1,
              ),
            ),
            reconnect: () => reconnected++,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('A response is waiting'), findsWidgets);
      expect(
        tester
            .widget<FilledButton>(
              find.byKey(const ValueKey('touch-primary-action')),
            )
            .onPressed,
        isNull,
      );
      await tester.tap(find.text('Restore connection'));
      expect(reconnected, 1);
      await tester.tap(find.byKey(const ValueKey('touch-tab-activity')));
      await tester.pumpAndSettle();
      expect(find.text('Ada · Passed'), findsOneWidget);
      expect(find.text('Bea · Awaiting response'), findsOneWidget);
      expect(find.text('Cem · Waiting'), findsOneWidget);
      for (final status in [
        'Ada · Passed',
        'Bea · Awaiting response',
        'Cem · Waiting',
      ]) {
        expect(find.bySemanticsLabel(status), findsOneWidget);
      }
      expect(tester.takeException(), isNull);
      semantics.dispose();
    },
  );

  for (final scale in [1.0, 2.0]) {
    testWidgets(
      'disabled move details exposes the explanation after scrolling at $scale text',
      (tester) async {
        size(tester, const Size(390, 788));
        var requests = 0;
        await tester.pumpWidget(
          _shell(
            _view(
              _data(actionEnabled: false, eventCount: 25),
              primary: () => requests++,
            ),
            scale: scale,
          ),
        );
        await tester.pumpAndSettle();
        expect(
          find.text('The consumer decides what this action means.'),
          findsNothing,
        );
        final details = find.byKey(const ValueKey('touch-move-details'));
        await tester.ensureVisible(details);
        await tester.tap(details);
        await tester.pumpAndSettle();
        expect(
          find.text('The consumer decides what this action means.'),
          findsOneWidget,
        );
        final scroll = find.byKey(
          ValueKey(
            scale == 1 ? 'touch-activity-scroll' : 'touch-natural-scroll',
          ),
        );
        await tester.drag(scroll, const Offset(0, -600));
        await tester.pumpAndSettle();
        expect(
          tester.widget<SingleChildScrollView>(scroll).controller!.offset,
          greaterThan(0),
        );
        await tester.ensureVisible(details);
        await tester.tap(details);
        await tester.pumpAndSettle();
        expect(
          find
              .text('The consumer decides what this action means.')
              .hitTestable(),
          findsOneWidget,
        );
        expect(
          tester
              .widget<FilledButton>(
                find.byKey(const ValueKey('touch-primary-action')),
              )
              .onPressed,
          isNull,
        );
        expect(requests, 0);
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('large text excludes inactive panels from keyboard focus', (
    tester,
  ) async {
    size(tester, const Size(390, 788));
    await tester.pumpWidget(_shell(_view(_data(), select: (_) {}), scale: 2));
    await tester.pumpAndSettle();
    final card = find.byKey(
      const ValueKey('hand-card-private-0'),
      skipOffstage: false,
    );
    final ink = find.descendant(
      of: card,
      matching: find.byType(InkWell, skipOffstage: false),
      skipOffstage: false,
    );
    expect(
      tester.widget<InkWell>(ink.first).focusNode!.canRequestFocus,
      isFalse,
    );
    final hand = find.byKey(const ValueKey('touch-tab-hand'));
    await tester.ensureVisible(hand);
    await tester.tap(hand);
    await tester.pumpAndSettle();
    expect(
      tester.widget<InkWell>(ink.first).focusNode!.canRequestFocus,
      isTrue,
    );
    expect(tester.takeException(), isNull);
  });

  for (final entry in [
    (const Size(390, 788), 2.0),
    (const Size(844, 334), 1.0),
  ]) {
    testWidgets(
      'touch content stays reachable at ${entry.$1} and ${entry.$2} text',
      (tester) async {
        size(tester, entry.$1);
        await tester.pumpWidget(
          _shell(_view(_data(empty: true)), scale: entry.$2),
        );
        await tester.pumpAndSettle();
        expect(
          find.byKey(const ValueKey('touch-natural-scroll')),
          findsOneWidget,
        );
        final hand = find.byKey(const ValueKey('touch-tab-hand'));
        await tester.ensureVisible(hand);
        await tester.tap(hand);
        await tester.pumpAndSettle();
        expect(find.text('Your hand is empty'), findsOneWidget);
        final action = find.byKey(const ValueKey('touch-primary-action'));
        await tester.ensureVisible(action);
        await tester.pumpAndSettle();
        expect(action.hitTestable(), findsOneWidget);
        expect(tester.takeException(), isNull);
      },
    );
  }
}
