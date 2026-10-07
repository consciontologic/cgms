import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

// This consumer deliberately imports neither cgms_demo nor cgms_ui/src.
// A real application's adapter can make the same projection from its services.
class ArchiveAdapter extends ChangeNotifier {
  var requests = 0;
  String? selectedId;
  String? inspectedCardId;
  final inspectedCombos = <String>[];

  static const first = VisibleCard(
    id: 'archive-card-201',
    rank: '8',
    suit: CardSuit.clubs,
    copy: 1,
  );
  static const second = VisibleCard(
    id: 'archive-card-202',
    rank: '8',
    suit: CardSuit.clubs,
    copy: 2,
  );
  static const queenFirst = VisibleCard(
    id: 'archive-royal-401',
    rank: 'Q',
    suit: CardSuit.hearts,
    copy: 1,
  );
  static const queenSecond = VisibleCard(
    id: 'archive-royal-402',
    rank: 'Q',
    suit: CardSuit.hearts,
    copy: 2,
  );
  static const jackFirst = VisibleCard(
    id: 'archive-private-501',
    rank: 'J',
    suit: CardSuit.clubs,
    copy: 1,
  );
  static const jackSecond = VisibleCard(
    id: 'archive-private-502',
    rank: 'J',
    suit: CardSuit.clubs,
    copy: 2,
  );
  static const formation = ComboView(
    id: 'archive-formation-7',
    title: 'Great People',
    stateLabel: 'Exposed formation · Hearts support unavailable',
    description:
        'These twin heart queens remain exposed. The archive adapter reports '
        'that their protection is inactive because Hearts support was lost.',
    cards: [queenFirst, queenSecond],
  );
  static const candidate = ComboView(
    id: 'archive-candidate-17',
    title: 'Kidnapper',
    stateLabel: 'Private candidate · not opened',
    description:
        'The viewer can inspect these twin club jacks in their own hand. '
        'The consumer must validate opening and ability eligibility separately.',
    cards: [jackFirst, jackSecond],
  );

  TableViewData get projection => TableViewData(
    gameId: 'archive-game-17',
    round: 4,
    totalRounds: 13,
    turnLabel: 'External adapter · Ada',
    seats: [
      _seat('archive-seat-a', 'Ada', 1, isSelf: true),
      _seat('archive-seat-b', 'Bea', 2),
      _seat('archive-seat-c', 'Cem', 3),
    ],
    hand: const [jackFirst, jackSecond],
    handCombos: const [candidate],
    events: [
      requests == 0
          ? 'Authorized archive snapshot loaded.'
          : 'External adapter received the request.',
    ],
    objective: 'Inspect an independent projection',
    actionLabel: 'Request adapter update',
    actionExplanation: 'The consumer owns this callback and its next state.',
    phaseLabel: 'Archive review',
    actionEnabled: true,
    selectedCardId: selectedId,
  );

  SeatView _seat(String id, String name, int number, {bool isSelf = false}) =>
      SeatView(
        id: id,
        name: name,
        number: number,
        isSelf: isSelf,
        exposed: isSelf
            ? const [first, second, queenFirst, queenSecond]
            : const [],
        combos: isSelf ? const [formation] : const [],
        concealed: ConcealedCards(handCount: isSelf ? 2 : 3, aceCount: 0),
        finances: FinancialView(
          matchScore: ExactPoints.fromStrings('11', '2'),
          gameScore: ExactPoints.fromStrings('3', '2'),
          available: ExactPoints.integer(0),
          debt: ExactPoints.fromStrings('1', '2'),
        ),
        status: 'Imported public position',
      );

  void select(String id) {
    selectedId = id;
    notifyListeners();
  }

  void inspectCard(VisibleCard card) => inspectedCardId = card.id;

  void inspectCombo(ComboView combo) => inspectedCombos.add(combo.id);

  void requestUpdate() {
    requests++;
    notifyListeners();
  }
}

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
  testWidgets('public package renders and dispatches an alternative adapter', (
    tester,
  ) async {
    final adapter = ArchiveAdapter();
    addTearDown(adapter.dispose);
    tester.view.physicalSize = const Size(1200, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: Scaffold(
          body: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ListenableBuilder(
              listenable: adapter,
              builder: (context, _) => CgmsTableView(
                data: adapter.projection,
                showArt: false,
                onSelect: adapter.select,
                onInspect: adapter.inspectCard,
                onInspectCombo: (combo) {
                  adapter.inspectCombo(combo);
                  showDialog<void>(
                    context: context,
                    builder: (_) =>
                        CgmsComboInspector(combo: combo, showArt: false),
                  );
                },
                onPrimary: adapter.requestUpdate,
                onTrade: () {},
                onScores: () {},
                onEndTurn: () {},
                onReconnect: () {},
              ),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('External adapter · Ada'), findsOneWidget);
    expect(find.text('02  /  Bea'), findsOneWidget);
    expect(find.text('03  /  Cem'), findsOneWidget);
    expect(find.text('Mara'), findsNothing);
    expect(find.text('Ivo'), findsNothing);
    expect(find.text('Nia'), findsNothing);
    expect(adapter.projection.seats.first.finances.matchScore.format(), '11/2');

    await tester.tap(find.text('Request adapter update'));
    await tester.pumpAndSettle();
    expect(adapter.requests, 1);
    expect(find.text('External adapter received the request.'), findsOneWidget);

    final copyTwo = find.byKey(const ValueKey('select-archive-card-202'));
    await tester.ensureVisible(copyTwo);
    await tester.tap(copyTwo);
    await tester.pump(const Duration(milliseconds: 350));
    await tester.pumpAndSettle();
    expect(adapter.selectedId, 'archive-card-202');
    expect(adapter.projection.selectedCardId, isNot(ArchiveAdapter.first.id));
    final selectedCard = find.ancestor(
      of: copyTwo,
      matching: find.byType(CgmsCard),
    );
    expect(tester.widget<CgmsCard>(selectedCard).selected, isTrue);
    expect(
      find.descendant(of: selectedCard, matching: find.byIcon(Icons.check)),
      findsOneWidget,
    );

    expect(
      tester
          .widgetList<CgmsComboGroup>(find.byType(CgmsComboGroup))
          .map((group) => group.combo.id),
      [ArchiveAdapter.formation.id],
    );
    final beforeInspection = adapter.projection;
    final queenCopyTwo = find.byKey(
      const ValueKey('combo-card-archive-royal-402'),
    );
    await tester.ensureVisible(queenCopyTwo);
    await tester.tap(queenCopyTwo);
    await tester.pump(const Duration(milliseconds: 350));
    expect(adapter.inspectedCardId, isNull);
    expect(adapter.selectedId, beforeInspection.selectedCardId);
    await tester.tap(queenCopyTwo);
    await tester.pump(const Duration(milliseconds: 50));
    await tester.tap(queenCopyTwo);
    await tester.pumpAndSettle();
    expect(adapter.inspectedCardId, ArchiveAdapter.queenSecond.id);

    for (final combo in [ArchiveAdapter.formation, ArchiveAdapter.candidate]) {
      final inspect = find.byKey(ValueKey('inspect-combo-${combo.id}'));
      await tester.ensureVisible(inspect);
      await tester.tap(inspect);
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CgmsComboInspector>(find.byType(CgmsComboInspector))
            .combo,
        same(combo),
      );
      final inspectedCards = tester
          .widgetList<CgmsCard>(
            find.descendant(
              of: find.byType(CgmsComboInspector),
              matching: find.byType(CgmsCard),
            ),
          )
          .map((card) => card.card.id);
      expect(inspectedCards, combo.cards.map((card) => card.id));
      await tester.tap(find.text('Close inspection'));
      await tester.pumpAndSettle();
    }
    expect(adapter.inspectedCombos, [
      ArchiveAdapter.formation.id,
      ArchiveAdapter.candidate.id,
    ]);
    expect(adapter.requests, 1);
    expect(adapter.projection.selectedCardId, beforeInspection.selectedCardId);
    expect(
      adapter.projection.hand.map((card) => card.id),
      beforeInspection.hand.map((card) => card.id),
    );
    expect(
      adapter.projection.seats.first.exposed.map((card) => card.id),
      beforeInspection.seats.first.exposed.map((card) => card.id),
    );
    expect(adapter.projection.seats.first.finances.matchScore.format(), '11/2');
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'standalone combo inspector reflows at 390px with 200 percent text',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final adapter = ArchiveAdapter();
      addTearDown(adapter.dispose);
      const combo = ComboView(
        id: 'archive-long-formation',
        title: 'Great People from a long named archived match',
        stateLabel: 'Exposed formation · awaiting restored Hearts support',
        description:
            'Both physical queens belong to this supplied formation. Inspection '
            'does not restore support, open a formation, select an action or '
            'change the external consumer state. Long explanatory text remains '
            'available through the dialog scroll area.',
        cards: [ArchiveAdapter.queenFirst, ArchiveAdapter.queenSecond],
      );
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          builder: (context, child) => MediaQuery(
            data: MediaQuery.of(
              context,
            ).copyWith(textScaler: const TextScaler.linear(2)),
            child: child!,
          ),
          home: Scaffold(
            body: Builder(
              builder: (context) => SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: CgmsComboGroup(
                  combo: combo,
                  showArt: false,
                  onInspectCard: adapter.inspectCard,
                  onInspect: () {
                    adapter.inspectCombo(combo);
                    showDialog<void>(
                      context: context,
                      builder: (_) => const CgmsComboInspector(
                        combo: combo,
                        showArt: false,
                      ),
                    );
                  },
                ),
              ),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      final inspect = find.byKey(
        const ValueKey('inspect-combo-archive-long-formation'),
      );
      await tester.ensureVisible(inspect);
      await tester.tap(inspect);
      await tester.pumpAndSettle();
      expect(find.byType(CgmsComboInspector), findsOneWidget);
      expect(tester.takeException(), isNull);
      final constituentCards = find.descendant(
        of: find.byType(CgmsComboInspector),
        matching: find.byType(CgmsCard),
      );
      expect(constituentCards, findsNWidgets(2));
      final cardRects = [
        for (final card in constituentCards.evaluate())
          tester.getRect(find.byWidget(card.widget)),
      ];
      for (final rect in cardRects) {
        expect(
          rect.width,
          greaterThanOrEqualTo(176),
          reason: 'Large text needs wider cards to keep the full suit readable',
        );
      }
      expect(cardRects[1].top, greaterThanOrEqualTo(cardRects[0].bottom));
      await tester.ensureVisible(find.text(combo.description));
      await tester.pumpAndSettle();
      expect(find.text(combo.description).hitTestable(), findsOneWidget);
      final close = find.widgetWithText(TextButton, 'Close inspection');
      expect(tester.getRect(close).height, greaterThanOrEqualTo(48));
      expect(tester.getRect(close).left, greaterThanOrEqualTo(0));
      expect(tester.getRect(close).right, lessThanOrEqualTo(390));
      expect(tester.getRect(close).bottom, lessThanOrEqualTo(844));
      await tester.tap(close);
      await tester.pumpAndSettle();
      expect(find.byType(CgmsComboInspector), findsNothing);
      expect(adapter.inspectedCombos, [combo.id]);
      expect(adapter.requests, 0);
      expect(adapter.selectedId, isNull);
      expect(tester.takeException(), isNull);
    },
  );
}
