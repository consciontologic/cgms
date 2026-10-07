import 'dart:ui' show SemanticsAction, PointerDeviceKind;
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'public_consumer_test.dart' show ArchiveAdapter;

Widget host({
  TableViewData? data,
  Widget? actions,
  bool cardDriven = false,
  ValueChanged<String>? activate,
  ValueChanged<TableDropTarget>? onTargetTap,
  Object? dragIdentity,
  Map<String, TableCardRole> cardRoles = const {},
  bool Function(TableCardDrag, TableDropTarget)? canDrop,
  bool Function(TableDropTarget)? isSuggestedTarget,
  TableDrawFeedback? drawFeedback,
  String? Function(TableCardDrag, TableDropTarget)? dropPreview,
  Object? actionsIdentity,
  ValueChanged<String>? select,
  ValueChanged<VisibleCard>? inspect,
  VoidCallback? coup,
  double scale = 1,
  bool rtl = false,
  bool showArt = false,
  Set<String> selectedIds = const {},
  Set<String> draggableIds = const {},
  void Function(TableCardDrag, TableDropTarget)? onDrop,
  Widget? footer,
  bool appBar = false,
  VoidCallback? cancelSelection,
  Widget? publicChooserTools,
}) => MaterialApp(
  theme: CgmsTheme.dark(),
  localizationsDelegates: CgmsLocalizations.localizationsDelegates,
  supportedLocales: CgmsLocalizations.supportedLocales,
  home: Directionality(
    textDirection: rtl ? TextDirection.rtl : TextDirection.ltr,
    child: MediaQuery(
      data: MediaQueryData(textScaler: TextScaler.linear(scale)),
      child: Scaffold(
        appBar: appBar ? AppBar(title: const Text('Online table')) : null,
        body: CgmsAdaptiveTable(
          data: data ?? ArchiveAdapter().projection,
          actions: cardDriven
              ? null
              : actions ?? const Text('Authoritative decisions'),
          onActivateCard: activate,
          onTargetTap: onTargetTap,
          dragIdentity: dragIdentity,
          cardRoles: cardRoles,
          canDrop: canDrop,
          isSuggestedTarget: isSuggestedTarget,
          drawFeedback: drawFeedback,
          dropPreview: dropPreview,
          onCancelSelection: cancelSelection,
          publicChooserTools: publicChooserTools,
          actionsIdentity: actionsIdentity,
          onSelect: select ?? (_) {},
          onInspect: inspect ?? (_) {},
          onCoup: coup,
          selectedCardIds: selectedIds,
          draggableCardIds: draggableIds,
          onCardDrop: onDrop,
          showArt: showArt,
          footer: footer,
        ),
      ),
    ),
  ),
);
Finder card(String id) => find.byKey(ValueKey('adaptive-select-$id'));
Finder pile([String group = 'clubs']) =>
    find.byKey(const ValueKey('adaptive-seat-archive-seat-a'));
FocusNode focus(WidgetTester t, Finder target) => t
    .widget<InkWell>(
      find.descendant(of: target, matching: find.byType(InkWell)).first,
    )
    .focusNode!;
Future<void> openPile(WidgetTester t, [String group = 'clubs']) async {
  if (find.byTooltip('Close actions').evaluate().isNotEmpty) {
    await t.tap(find.byTooltip('Close actions'));
    await t.pumpAndSettle();
  }
  await t.ensureVisible(pile(group));
  await t.tap(pile(group));
  await t.pumpAndSettle();
}

Future<void> twice(WidgetTester t, Finder target) async {
  await t.ensureVisible(target);
  final point = exposedPoint(t, target);
  await t.tapAt(point);
  await t.pump(const Duration(milliseconds: 50));
  await t.tapAt(point);
  await t.pumpAndSettle();
}

Offset exposedPoint(WidgetTester t, Finder target) {
  final card = t.widget<CgmsCard>(target);
  final rect = t.getRect(target);
  if (!card.handStackPreview) return rect.center;
  return Offset(
    Directionality.of(t.element(target)) == TextDirection.rtl
        ? rect.right - 24
        : rect.left + 24,
    rect.center.dy,
  );
}

TableViewData projection({
  List<VisibleCard>? hand,
  List<VisibleCard>? exposed,
  List<ComboView>? combos,
  PendingView? pending,
  String? gameId,
  List<VisibleCard> opponentExposed = const [],
}) {
  final b = ArchiveAdapter().projection, s = b.seats.first;
  return TableViewData(
    gameId: gameId ?? b.gameId,
    round: b.round,
    totalRounds: b.totalRounds,
    turnLabel: b.turnLabel,
    seats: [
      SeatView(
        id: s.id,
        name: s.name,
        number: s.number,
        exposed: exposed ?? s.exposed,
        concealed: s.concealed,
        finances: s.finances,
        status: s.status,
        isSelf: true,
        combos: combos ?? s.combos,
      ),
      for (final other in b.seats.skip(1))
        SeatView(
          id: other.id,
          name: other.name,
          number: other.number,
          exposed: opponentExposed,
          concealed: other.concealed,
          finances: other.finances,
          status: other.status,
        ),
    ],
    hand: hand ?? b.hand,
    events: b.events,
    objective: b.objective,
    actionLabel: b.actionLabel,
    actionExplanation: b.actionExplanation,
    phaseLabel: b.phaseLabel,
    pending: pending,
  );
}

class _RetainedActionWorkspace extends StatefulWidget {
  const _RetainedActionWorkspace({this.shortPrompt = false, this.onStay});

  final bool shortPrompt;
  final VoidCallback? onStay;

  @override
  State<_RetainedActionWorkspace> createState() =>
      _RetainedActionWorkspaceState();
}

class _RetainedActionWorkspaceState extends State<_RetainedActionWorkspace> {
  final draft = TextEditingController();
  final draftFocus = FocusNode();

  @override
  void dispose() {
    draft.dispose();
    draftFocus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: widget.shortPrompt
        ? [
            const Text('Choose whether to stay for the next round.'),
            FilledButton(
              onPressed: widget.onStay,
              child: const Text('Stay in game'),
            ),
            OutlinedButton(onPressed: () {}, child: const Text('Leave game')),
          ]
        : [
            const SizedBox(height: 400),
            TextField(
              controller: draft,
              focusNode: draftFocus,
              decoration: const InputDecoration(labelText: 'Action draft'),
            ),
            const SizedBox(height: 1000),
          ],
  );
}

void main() {
  testWidgets(
    'draw pile has its own safe space across live widths text and rotation',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final scale in [1.0, 2.0]) {
        for (final rtl in [false, true]) {
          await t.pumpWidget(host(cardDriven: true, scale: scale, rtl: rtl));
          for (final size in const [
            Size(1041, 948),
            Size(599, 900),
            Size(600, 900),
            Size(599, 900),
            Size(1049, 900),
            Size(1050, 900),
            Size(1049, 900),
            Size(390, 844),
            Size(844, 390),
            Size(319, 948),
            Size(1280, 400),
          ]) {
            await t.binding.setSurfaceSize(size);
            await t.pumpAndSettle();
            final draw = t.getRect(
              find.byKey(const ValueKey('adaptive-draw-pile')),
            );
            expect(draw.width, greaterThanOrEqualTo(48));
            expect(draw.height, greaterThanOrEqualTo(48));
            for (final seat in ArchiveAdapter().projection.seats) {
              final cell = t.getRect(
                find.byKey(ValueKey('adaptive-seat-${seat.id}')),
              );
              expect(
                draw.overlaps(cell),
                isFalse,
                reason:
                    'Pile must not paint or hit-test over ${seat.name} '
                    'at $size / $scale / RTL=$rtl',
              );
            }
            final table = t.getRect(
              find.byKey(const ValueKey('adaptive-public-square')),
            );
            final hand = t.getRect(
              find.byKey(const ValueKey('adaptive-hand-viewport')),
            );
            expect(hand.top - table.bottom, closeTo(4, .01));
            expect(
              t
                  .getSize(
                    find.byKey(
                      ValueKey('public-preview-${ArchiveAdapter.first.id}'),
                    ),
                  )
                  .width,
              greaterThanOrEqualTo(72),
            );
            expect(t.takeException(), isNull);
          }
        }
      }
    },
  );

  testWidgets('1041 draw pile and lower seat have independent touch targets', (
    t,
  ) async {
    addTearDown(() => t.binding.setSurfaceSize(null));
    await t.binding.setSurfaceSize(const Size(1041, 948));
    final targets = <TableDropTarget>[];
    final drops = <TableDropTarget>[];
    await t.pumpWidget(
      host(
        cardDriven: true,
        onTargetTap: targets.add,
        draggableIds: {ArchiveAdapter.jackFirst.id},
        onDrop: (_, target) => drops.add(target),
      ),
    );
    await t.pumpAndSettle();
    final draw = find.byKey(const ValueKey('adaptive-draw-pile'));
    final seat = find.byKey(const ValueKey('seat-target-3'));
    await t.tap(seat);
    await t.pumpAndSettle();
    await t.tap(draw);
    await t.pumpAndSettle();
    expect(targets.map((target) => target.zone), [
      TableDropZone.seat,
      TableDropZone.drawPile,
    ]);
    expect(targets.first.seat, 3);
    final pointer = await t.startGesture(
      exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
      kind: PointerDeviceKind.touch,
    );
    await t.pump(const Duration(milliseconds: 600));
    await pointer.moveTo(t.getCenter(draw));
    await t.pump();
    await pointer.up();
    await t.pumpAndSettle();
    expect(drops.single.zone, TableDropZone.drawPile);
    expect(targets.length, 2);
    expect(t.takeException(), isNull);
  });

  testWidgets(
    'formation header drags its exact members independent of selection',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      await t.binding.setSurfaceSize(const Size(1041, 948));
      final drops = <TableCardDrag>[];
      final targets = <TableDropTarget>[];
      await t.pumpWidget(
        host(
          cardDriven: true,
          selectedIds: {ArchiveAdapter.jackFirst.id},
          draggableIds: {
            ArchiveAdapter.queenFirst.id,
            ArchiveAdapter.queenSecond.id,
            ArchiveAdapter.jackFirst.id,
          },
          onDrop: (drag, target) {
            drops.add(drag);
            targets.add(target);
          },
        ),
      );
      await t.pumpAndSettle();
      final formation = find.byKey(
        ValueKey('formation-target-1-${ArchiveAdapter.formation.id}'),
      );
      await t.ensureVisible(formation);
      final pointer = await t.startGesture(
        t.getCenter(formation),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, 24));
      await t.pump();
      await pointer.moveTo(
        t.getCenter(find.byKey(const ValueKey('seat-target-3'))),
      );
      await t.pump();
      await pointer.up();
      await t.pumpAndSettle();
      expect(drops.single.cardIds, [
        ArchiveAdapter.queenFirst.id,
        ArchiveAdapter.queenSecond.id,
      ]);
      expect(targets.single.zone, TableDropZone.seat);
      expect(targets.single.seat, 3);
      expect(t.takeException(), isNull);
    },
  );

  testWidgets('formation drag cancels on revoked member and Escape', (t) async {
    addTearDown(() => t.binding.setSurfaceSize(null));
    await t.binding.setSurfaceSize(const Size(1041, 948));
    final drops = <TableCardDrag>[];
    var revoked = false;
    Widget build() => host(
      cardDriven: true,
      draggableIds: {
        ArchiveAdapter.queenFirst.id,
        if (!revoked) ArchiveAdapter.queenSecond.id,
      },
      onDrop: (drag, _) => drops.add(drag),
    );
    for (final cancel in ['revocation', 'Escape']) {
      revoked = false;
      await t.pumpWidget(build());
      await t.pumpAndSettle();
      final formation = find.byKey(
        ValueKey('formation-target-1-${ArchiveAdapter.formation.id}'),
      );
      await t.ensureVisible(formation);
      final pointer = await t.startGesture(
        t.getCenter(formation),
        kind: PointerDeviceKind.touch,
      );
      await t.pump(const Duration(milliseconds: 600));
      await pointer.moveTo(
        t.getCenter(find.byKey(const ValueKey('seat-target-3'))),
      );
      await t.pump();
      if (cancel == 'revocation') {
        revoked = true;
        await t.pumpWidget(build());
      } else {
        await t.sendKeyEvent(LogicalKeyboardKey.escape);
      }
      await t.pump();
      await pointer.up();
      await t.pumpAndSettle();
      expect(drops, isEmpty, reason: cancel);
      expect(t.takeException(), isNull);
    }
  });

  testWidgets(
    'draw feedback preserves pile geometry focus and immediate Coup',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      await t.binding.setSurfaceSize(const Size(1041, 948));
      var coups = 0;
      final targets = <TableDropTarget>[];
      Widget build(TableDrawFeedback? feedback) => host(
        cardDriven: true,
        drawFeedback: feedback,
        onTargetTap: targets.add,
        coup: () => coups++,
      );
      await t.pumpWidget(build(null));
      await t.pumpAndSettle();
      final draw = find.byKey(const ValueKey('adaptive-draw-pile'));
      final before = t.getRect(draw);
      final node = focus(t, card(ArchiveAdapter.jackFirst.id));
      node.requestFocus();
      await t.pump();
      await t.pumpWidget(
        build(
          TableDrawFeedback(
            id: 'confirmed-draw',
            seat: 1,
            destination: DrawDestination.concealedAce,
            startedAt: DateTime.now(),
          ),
        ),
      );
      await t.pump();
      expect(find.byKey(const ValueKey('draw-highlight')), findsOneWidget);
      expect(t.getRect(draw), before);
      expect(node.hasPrimaryFocus, isTrue);
      expect(coups, 0);
      expect(targets, isEmpty);
      await t.tap(draw);
      await t.pump();
      expect(targets.single.zone, TableDropZone.drawPile);
      await t.tap(find.byKey(const ValueKey('immediate-coup')));
      await t.pump();
      expect(coups, 1);
      await t.pumpAndSettle();
      expect(t.takeException(), isNull);
    },
  );

  testWidgets('missing choice highlights only its specific board destination', (
    t,
  ) async {
    var calls = 0;
    Widget build(bool missing) => host(
      cardDriven: true,
      isSuggestedTarget: (target) =>
          missing &&
          target.zone == TableDropZone.series &&
          target.seat == 1 &&
          target.suit == 'clubs',
      onTargetTap: (_) => calls++,
      onDrop: (_, __) => calls++,
    );
    Finder highlighted() => find.byWidgetPredicate(
      (widget) =>
          widget is DecoratedBox &&
          widget.position == DecorationPosition.foreground &&
          widget.decoration is BoxDecoration &&
          (widget.decoration as BoxDecoration).border ==
              Border.all(color: CgmsColors.cyan, width: 2),
    );
    await t.pumpWidget(build(true));
    await t.pumpAndSettle();
    expect(highlighted(), findsOneWidget);
    expect(
      find.descendant(
        of: highlighted(),
        matching: find.byKey(const ValueKey('formation-target-1-clubs')),
      ),
      findsOneWidget,
    );
    expect(calls, 0);
    await t.pumpWidget(build(false));
    await t.pumpAndSettle();
    expect(highlighted(), findsNothing);
    expect(calls, 0);
  });
  testWidgets(
    'a scrolled hand becoming fully visible renews only its scroll viewport',
    (t) async {
      final semantics = t.ensureSemantics();
      addTearDown(() => t.binding.setSurfaceSize(null));
      final data = projection(
        hand: [
          for (final suit in CardSuit.values)
            VisibleCard(
              id: 'retained-${suit.name}',
              rank: '10',
              suit: suit,
              copy: 1,
            ),
        ],
      );
      final selected = <String>{};
      var activations = 0;
      var drops = 0;
      Widget build() => host(
        data: data,
        cardDriven: true,
        selectedIds: selected,
        select: selected.add,
        activate: (_) => activations++,
        onDrop: (_, __) => drops++,
      );
      try {
        await t.binding.setSurfaceSize(const Size(390, 600));
        await t.pumpWidget(build());
        await t.pumpAndSettle();
        final target = card('retained-clubs');
        await t.ensureVisible(target);
        await t.tapAt(exposedPoint(t, target));
        await t.pumpAndSettle(const Duration(milliseconds: 350));
        await t.pumpWidget(build());
        await t.pumpAndSettle();
        final hand = find.byKey(const PageStorageKey('private-hand-scroll'));
        ScrollPosition position() => t
            .state<ScrollableState>(
              find
                  .descendant(of: hand, matching: find.byType(Scrollable))
                  .first,
            )
            .position;
        expect(position().maxScrollExtent, greaterThan(0));
        await t.drag(hand, const Offset(0, -70));
        await t.pumpAndSettle();
        expect(position().pixels, greaterThan(0));
        final cardElement = t.element(target);
        final cardFocus = focus(t, target);
        cardFocus.requestFocus();
        await t.pumpAndSettle();
        final semanticId = t
            .getSemantics(find.byKey(const ValueKey('select-retained-clubs')))
            .id;
        final scrolledPosition = position();

        await t.binding.setSurfaceSize(const Size(390, 1100));
        await t.pumpAndSettle();
        expect(position().maxScrollExtent, 0);
        expect(position().pixels, 0);
        expect(
          position(),
          isNot(same(scrolledPosition)),
          reason:
              'A no-longer-scrollable viewport must discard its stale native scroll compensation.',
        );
        expect(t.element(target), same(cardElement));
        expect(focus(t, target), same(cardFocus));
        expect(FocusManager.instance.primaryFocus, same(cardFocus));
        expect(
          t
              .getSemantics(find.byKey(const ValueKey('select-retained-clubs')))
              .id,
          semanticId,
        );
        expect(t.widget<CgmsCard>(target).selected, isTrue);
        expect(selected, {'retained-clubs'});
        final viewport = t.getRect(hand);
        expect(t.getRect(target).intersect(viewport), t.getRect(target));
        final fittedPosition = position();
        await t.binding.setSurfaceSize(const Size(390, 1150));
        await t.pumpAndSettle();
        expect(position(), same(fittedPosition));
        expect(t.element(target), same(cardElement));
        expect(FocusManager.instance.primaryFocus, same(cardFocus));
        expect(activations, 0);
        expect(drops, 0);
        expect(t.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'Escape outside table focus invalidates held pointer drag before release',
    (t) async {
      final external = FocusNode();
      addTearDown(external.dispose);
      final drops = <TableCardDrag>[];
      await t.pumpWidget(
        Focus(
          focusNode: external,
          onKeyEvent: (_, e) => e.logicalKey == LogicalKeyboardKey.escape
              ? KeyEventResult.handled
              : KeyEventResult.ignored,
          child: host(
            cardDriven: true,
            draggableIds: {ArchiveAdapter.jackFirst.id},
            onDrop: (d, _) => drops.add(d),
          ),
        ),
      );
      external.requestFocus();
      await t.pump();
      final pointer = await t.startGesture(
        exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -24));
      await t.pump();
      await pointer.moveTo(
        t.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
      );
      await t.pump();
      expect(external.hasPrimaryFocus, isTrue);
      await t.sendKeyEvent(LogicalKeyboardKey.escape);
      await t.pump();
      await pointer.up();
      await t.pumpAndSettle();
      expect(drops, isEmpty);
    },
  );
  testWidgets(
    'card-driven hand destination keeps heading and physical semantics separate',
    (t) async {
      final semantics = t.ensureSemantics();
      final targets = <TableDropTarget>[];
      await t.pumpWidget(host(cardDriven: true, onTargetTap: targets.add));
      final title = find.textContaining('Your hand ·');
      final node = t.getSemantics(title).getSemanticsData();
      expect(node.flagsCollection.isHeader, isTrue);
      expect(node.flagsCollection.isButton, isFalse);
      expect(node.headingLevel, 2);
      expect(node.label, 'Your hand · 2');
      expect(find.bySemanticsLabel('J of Clubs, copy 1'), findsOneWidget);
      await t.tap(title);
      await t.pumpAndSettle();
      expect(targets.single.zone, TableDropZone.hand);
      semantics.dispose();
    },
  );
  testWidgets(
    'accepted drag previews exact move and cancellation clears preview',
    (t) async {
      final semantics = t.ensureSemantics();
      final drops = <TableCardDrag>[];
      await t.pumpWidget(
        host(
          cardDriven: true,
          draggableIds: {ArchiveAdapter.jackFirst.id},
          canDrop: (_, target) => target.zone == TableDropZone.drawPile,
          dropPreview: (drag, target) =>
              'Preview ${drag.cardIds.single} at ${target.zone.name}',
          onDrop: (drag, _) => drops.add(drag),
        ),
      );
      await t.ensureVisible(card(ArchiveAdapter.jackFirst.id));
      await t.pumpAndSettle();
      final pointer = await t.startGesture(
        exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -24));
      await t.pump();
      await pointer.moveTo(
        t.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
      );
      await t.pump();
      expect(
        find.text('Preview archive-private-501 at drawPile'),
        findsOneWidget,
      );
      expect(
        t
            .getSemantics(find.text('Preview archive-private-501 at drawPile'))
            .getSemanticsData()
            .flagsCollection
            .isLiveRegion,
        isTrue,
      );
      await pointer.moveTo(const Offset(1, 1));
      await t.pump();
      expect(
        find.text('Preview archive-private-501 at drawPile'),
        findsNothing,
      );
      await pointer.up();
      await t.pumpAndSettle();
      expect(drops, isEmpty);
      semantics.dispose();
    },
  );
  testWidgets(
    'selecting and toggling at the same touch point never moves hand cards',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final width in [390.0, 599.0, 1041.0]) {
        await t.binding.setSurfaceSize(Size(width, 948));
        final selected = <String>{};
        final activations = <String>[];
        final inspections = <VisibleCard>[];
        await t.pumpWidget(
          StatefulBuilder(
            builder: (context, rebuild) => host(
              cardDriven: true,
              selectedIds: selected,
              activate: activations.add,
              inspect: inspections.add,
              select: (id) => rebuild(() {
                if (!selected.remove(id)) selected.add(id);
              }),
            ),
          ),
        );
        await t.pumpAndSettle();
        final source = card(ArchiveAdapter.jackFirst.id);
        await t.ensureVisible(source);
        await t.pumpAndSettle();
        final before = t.getRect(source);
        final point = exposedPoint(t, source);
        await t.tapAt(point);
        await t.pump(const Duration(milliseconds: 350));
        await t.pumpAndSettle();
        expect(selected, {ArchiveAdapter.jackFirst.id});
        expect(t.getRect(source), before, reason: '$width first selection');
        expect(find.byTooltip('Inspect selected cards'), findsOneWidget);
        await t.tap(find.byTooltip('Inspect selected cards'));
        await t.pumpAndSettle();
        expect(inspections.single.id, ArchiveAdapter.jackFirst.id);
        expect(selected, {ArchiveAdapter.jackFirst.id});
        await t.tapAt(point);
        await t.pump(const Duration(milliseconds: 350));
        await t.pumpAndSettle();
        expect(selected, isEmpty);
        expect(t.getRect(source), before, reason: '$width toggle');
        expect(activations, isEmpty);
      }
    },
  );
  testWidgets(
    '18 card opening and eight member loan previews fit at 200 percent after rotation',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      // This is a rendered-text boundary, so use the shipped readable typeface
      // instead of the test harness's square-glyph Ahem approximation.
      final loader = FontLoader('packages/cgms_ui/Atkinson Hyperlegible Next');
      loader.addFont(
        rootBundle.load(
          'packages/cgms_ui/assets/fonts/AtkinsonHyperlegibleNext-Regular.ttf',
        ),
      );
      await loader.load();
      final physicalCards = [
        for (var rank = 2; rank <= 10; rank++) '$rank[1,2]',
      ].join(' ');
      final openingPreview =
          'Release to submit Open Series\n'
          'Clubs: $physicalCards / Seat 1 · Clubs\n'
          'Copy numbers in brackets.\n'
          'Uses this turn’s opening allowance.';
      const loanPreview =
          'Release to submit Loan this intact formation\n'
          'Hearts: Q[1,2]; Diamonds: Q[1,2]; Clubs: Q[1,2]; Spades: Q[1,2] / Seat 2\n'
          'Copy numbers in brackets.\n'
          'No cards requested. Exact revision acceptance and successful responses resolve the loan and exclusive alliance together.';
      final drops = <TableCardDrag>[];
      for (final preview in [openingPreview, loanPreview]) {
        for (final rtl in [false, true]) {
          for (final size in [
            const Size(390, 844),
            const Size(844, 390),
            const Size(1041, 948),
          ]) {
            await t.binding.setSurfaceSize(size);
            await t.pumpWidget(
              host(
                cardDriven: true,
                scale: 2,
                rtl: rtl,
                draggableIds: {ArchiveAdapter.jackFirst.id},
                canDrop: (_, target) => target.zone == TableDropZone.hand,
                dropPreview: (_, __) => preview,
                onDrop: (drag, _) => drops.add(drag),
              ),
            );
            await t.pumpAndSettle();
            await t.ensureVisible(card(ArchiveAdapter.jackFirst.id));
            await t.pumpAndSettle();
            final source = exposedPoint(t, card(ArchiveAdapter.jackFirst.id));
            final pointer = await t.startGesture(
              source,
              kind: rtl ? PointerDeviceKind.touch : PointerDeviceKind.mouse,
            );
            if (rtl) await t.pump(const Duration(milliseconds: 400));
            // The lower right edge remains the hand's broad destination.
            await pointer.moveTo(Offset(size.width - 10, size.height - 10));
            await t.pump();
            expect(find.text(preview), findsOneWidget);
            final rect = t.getRect(find.text(preview));
            expect(
              rect.left,
              greaterThanOrEqualTo(8),
              reason: '$size rtl=$rtl',
            );
            expect(rect.top, greaterThanOrEqualTo(8), reason: '$size rtl=$rtl');
            expect(rect.right, lessThanOrEqualTo(size.width - 8));
            expect(rect.bottom, lessThanOrEqualTo(size.height - 8));
            final text = t.widget<Text>(find.text(preview));
            expect(text.maxLines, isNull);
            expect(text.overflow, isNull);
            expect(
              MediaQuery.textScalerOf(t.element(find.text(preview))).scale(14),
              28,
            );
            expect(drops, isEmpty);
            await t.sendKeyEvent(LogicalKeyboardKey.escape);
            await t.pump();
            await pointer.up();
            await t.pumpAndSettle();
            expect(find.text(preview), findsNothing);
            expect(drops, isEmpty);
            expect(t.takeException(), isNull);
          }
        }
      }
    },
  );
  testWidgets(
    'card interface scrolls header and cards when recovery leaves a tiny slot',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      await t.binding.setSurfaceSize(const Size(390, 60));
      await t.pumpWidget(host(cardDriven: true, scale: 2, rtl: true));
      await t.pumpAndSettle();
      expect(t.takeException(), isNull);
      expect(
        find.byKey(const PageStorageKey('card-board-scroll')),
        findsOneWidget,
      );
      final node = focus(t, card(ArchiveAdapter.jackFirst.id));
      node.requestFocus();
      await t.pumpAndSettle();
      expect(node.hasPrimaryFocus, isTrue);
      expect(t.takeException(), isNull);
    },
  );

  testWidgets(
    'card interface attachments stay with their authorized supporting series',
    (t) async {
      const bomb = VisibleCard(
        id: 'attached-bomb',
        rank: 'J',
        suit: CardSuit.hearts,
        copy: 1,
        seriesSuit: CardSuit.diamonds,
      );
      await t.pumpWidget(
        host(
          cardDriven: true,
          data: projection(exposed: [bomb], combos: []),
        ),
      );
      await t.pumpAndSettle();
      expect(
        find.byKey(const ValueKey('formation-target-1-diamonds')),
        findsOneWidget,
      );
      expect(
        find.byKey(const ValueKey('formation-target-1-hearts')),
        findsNothing,
      );
      expect(
        t
            .widget<CgmsCard>(
              find.byKey(const ValueKey('public-preview-attached-bomb')),
            )
            .card
            .suit,
        CardSuit.hearts,
      );
    },
  );

  testWidgets(
    'card interface chooser retains immediate card tools and independent inspection',
    (t) async {
      var activations = 0;
      final inspections = <VisibleCard>[];
      await t.pumpWidget(
        host(
          cardDriven: true,
          inspect: inspections.add,
          publicChooserTools: CgmsCard(
            card: ArchiveAdapter.queenFirst,
            width: 72,
            compact: true,
            showArt: false,
            onActivate: () => activations++,
          ),
        ),
      );
      await t.pumpAndSettle();
      await t.tap(find.byTooltip('Inspect public cards').first);
      await t.pumpAndSettle();
      final immediate = find
          .ancestor(
            of: find
                .byKey(ValueKey('select-${ArchiveAdapter.queenFirst.id}'))
                .last,
            matching: find.byType(CgmsCard),
          )
          .first;
      await twice(t, immediate);
      expect(activations, 1);
      final fullCard = card(ArchiveAdapter.second.id);
      final magnifier = find.descendant(
        of: fullCard,
        matching: find.byType(IconButton),
      );
      await t.ensureVisible(magnifier);
      await t.tap(magnifier);
      await t.pumpAndSettle();
      expect(inspections.single.id, ArchiveAdapter.second.id);
      expect(find.text('Close public cards'), findsNothing);
    },
  );

  testWidgets(
    'card interface direct physical public cards and series are accessible',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final selected = <String>[], activated = <String>[];
      final targets = <TableDropTarget>[];
      await t.pumpWidget(
        host(
          cardDriven: true,
          select: selected.add,
          activate: activated.add,
          onTargetTap: targets.add,
        ),
      );
      await t.pumpAndSettle();
      for (final id in [ArchiveAdapter.first.id, ArchiveAdapter.second.id]) {
        final face = find.byKey(ValueKey('public-preview-$id'));
        await t.ensureVisible(face);
        await t.tapAt(exposedPoint(t, face));
        await t.pump(const Duration(milliseconds: 350));
        await t.pumpAndSettle();
      }
      expect(selected, [ArchiveAdapter.first.id, ArchiveAdapter.second.id]);
      final series = find.byKey(const ValueKey('formation-target-1-clubs'));
      await t.ensureVisible(series);
      await t.tap(series);
      await t.pump(const Duration(milliseconds: 350));
      await t.pumpAndSettle();
      expect(targets.single.zone, TableDropZone.series);
      expect(targets.single.suit, 'clubs');
      expect(targets.single.seat, 1);
      final face = find.byKey(
        ValueKey('public-preview-${ArchiveAdapter.second.id}'),
      );
      await twice(t, face);
      expect(activated, [ArchiveAdapter.second.id]);
      expect(selected, hasLength(2));
    },
  );

  testWidgets(
    'card interface exact bundle rejects revoked members without substitution',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final drops = <TableCardDrag>[];
      final ids = {ArchiveAdapter.jackFirst.id, ArchiveAdapter.jackSecond.id};
      Widget build(Set<String> allowed) => host(
        cardDriven: true,
        selectedIds: ids,
        draggableIds: allowed,
        onDrop: (drag, _) => drops.add(drag),
      );
      await t.pumpWidget(build({ArchiveAdapter.jackFirst.id}));
      await t.pumpAndSettle();
      Future<void> drag() async {
        final pointer = await t.startGesture(
          exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
          kind: PointerDeviceKind.mouse,
        );
        await pointer.moveBy(const Offset(0, -24));
        await t.pump();
        await pointer.moveTo(
          t.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
        );
        await t.pump();
        await pointer.up();
        await t.pumpAndSettle();
      }

      await drag();
      expect(drops, isEmpty);
      await t.pumpWidget(build(ids));
      await t.pumpAndSettle();
      await drag();
      expect(drops.single.cardIds, unorderedEquals(ids));
    },
  );

  testWidgets(
    'card interface touch drag cancels by Escape and respects rejected targets',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      var drops = 0, selections = 0, cancellations = 0;
      Widget build(bool accepts) => host(
        cardDriven: true,
        select: (_) => selections++,
        cancelSelection: () => cancellations++,
        draggableIds: {ArchiveAdapter.jackFirst.id},
        canDrop: (_, target) =>
            accepts && target.zone == TableDropZone.drawPile,
        onDrop: (_, __) => drops++,
      );
      await t.pumpWidget(build(false));
      await t.pumpAndSettle();
      Future<void> drag({bool escape = false}) async {
        final pointer = await t.startGesture(
          exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
        );
        await t.pump(const Duration(milliseconds: 400));
        await pointer.moveTo(
          t.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
        );
        await t.pump();
        if (escape) {
          await t.sendKeyEvent(LogicalKeyboardKey.escape);
          await t.pump();
        }
        await pointer.up();
        await t.pumpAndSettle();
      }

      await drag();
      expect(drops, 0);
      await t.pumpWidget(build(true));
      await t.pumpAndSettle();
      focus(t, card(ArchiveAdapter.jackFirst.id)).requestFocus();
      await t.pump();
      await drag(escape: true);
      expect(drops, 0);
      expect(cancellations, 1);
      await drag();
      expect(drops, 1);
      expect(selections, 0);
    },
  );

  testWidgets(
    'card interface preserves focus roles and readable faces across live layouts',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      final semantics = t.ensureSemantics();

      var calls = 0;
      FocusNode? original;
      for (final rtl in [false, true]) {
        for (final scale in [1.0, 2.0]) {
          for (final size in [
            const Size(390, 844),
            const Size(599, 700),
            const Size(600, 700),
            const Size(1049, 948),
            const Size(1050, 948),
            const Size(844, 390),
            const Size(1280, 400),
          ]) {
            await t.binding.setSurfaceSize(size);
            await t.pumpWidget(
              host(
                cardDriven: true,
                rtl: rtl,
                scale: scale,
                selectedIds: {ArchiveAdapter.jackFirst.id},
                cardRoles: {ArchiveAdapter.jackFirst.id: TableCardRole.payment},
                select: (_) => calls++,
                activate: (_) => calls++,
              ),
            );
            await t.pumpAndSettle();
            final node = focus(t, card(ArchiveAdapter.jackFirst.id));
            original ??= node;
            expect(node, same(original));
            node.requestFocus();
            await t.pumpAndSettle();
            expect(node.hasPrimaryFocus, isTrue);
            expect(
              t
                  .getSize(
                    find.byKey(
                      ValueKey('public-preview-${ArchiveAdapter.first.id}'),
                    ),
                  )
                  .width,
              greaterThanOrEqualTo(72),
            );
            expect(
              t
                  .getSemantics(
                    find.byKey(
                      ValueKey('select-${ArchiveAdapter.jackFirst.id}'),
                    ),
                  )
                  .getSemanticsData()
                  .value,
              contains('Payment card'),
            );
            expect(find.byTooltip('Choose an action'), findsNothing);
            expect(
              t.takeException(),
              isNull,
              reason: '$size scale=$scale rtl=$rtl',
            );
          }
        }
      }
      expect(calls, 0);
      semantics.dispose();
    },
  );

  testWidgets(
    'card interface has no action subtree and double tap activates only',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final selections = <String>[], activations = <String>[];
      final inspections = <VisibleCard>[];
      await t.pumpWidget(
        host(
          cardDriven: true,
          select: selections.add,
          activate: activations.add,
          inspect: inspections.add,
        ),
      );
      await t.pumpAndSettle();
      expect(
        find.text('Authoritative decisions', skipOffstage: false),
        findsNothing,
      );
      expect(find.byTooltip('Choose an action'), findsNothing);
      await twice(t, card(ArchiveAdapter.jackFirst.id));
      expect(activations, [ArchiveAdapter.jackFirst.id]);
      expect(selections, isEmpty);
      expect(inspections, isEmpty);
      focus(t, card(ArchiveAdapter.jackFirst.id)).requestFocus();
      await t.pump();
      await t.sendKeyEvent(LogicalKeyboardKey.keyI);
      await t.pumpAndSettle();
      expect(inspections.single.id, ArchiveAdapter.jackFirst.id);
      await t.sendKeyEvent(LogicalKeyboardKey.space);
      await t.pumpAndSettle();
      expect(selections, [ArchiveAdapter.jackFirst.id]);
      await t.sendKeyEvent(LogicalKeyboardKey.enter);
      await t.pumpAndSettle();
      expect(activations, [
        ArchiveAdapter.jackFirst.id,
        ArchiveAdapter.jackFirst.id,
      ]);
    },
  );

  testWidgets(
    'card interface pile tap is a destination and drag revision cancels',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final destinations = <TableDropTarget>[];
      final drops = <TableCardDrag>[];
      Widget build(int version) => host(
        cardDriven: true,
        dragIdentity: version,
        draggableIds: {ArchiveAdapter.jackFirst.id},
        onTargetTap: destinations.add,
        onDrop: (drag, _) => drops.add(drag),
      );
      await t.pumpWidget(build(1));
      await t.pumpAndSettle();
      await t.tap(find.byKey(const ValueKey('adaptive-draw-pile')));
      await t.pumpAndSettle();
      expect(destinations.single.zone, TableDropZone.drawPile);
      final gesture = await t.startGesture(
        exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
        kind: PointerDeviceKind.mouse,
      );
      await gesture.moveBy(const Offset(0, -24));
      await t.pump();
      await t.pumpWidget(build(2));
      await t.pump();
      await gesture.moveTo(
        t.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
      );
      await t.pump();
      await gesture.up();
      await t.pumpAndSettle();
      expect(drops, isEmpty);
      expect(destinations, hasLength(1));
    },
  );

  Future<TestGesture> mouseDrag(WidgetTester t, Finder source) async {
    final point =
        t.widget(source) is CgmsCard &&
            t.widget<CgmsCard>(source).handStackPreview
        ? exposedPoint(t, source)
        : t.getCenter(source);
    final gesture = await t.startGesture(point, kind: PointerDeviceKind.mouse);
    await gesture.moveBy(const Offset(0, -24));
    await t.pump();
    return gesture;
  }

  Future<void> drop(WidgetTester t, TestGesture gesture, Finder target) async {
    await gesture.moveTo(t.getCenter(target));
    await t.pump();
    await gesture.up();
    await t.pumpAndSettle();
  }

  testWidgets(
    'card drag rejects changed game, revoked cards and pending state',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final ids = {ArchiveAdapter.jackFirst.id, ArchiveAdapter.jackSecond.id};
      final drops = <TableCardDrag>[];
      Widget board({TableViewData? data, Set<String>? allowed}) => host(
        data: data,
        draggableIds: allowed ?? ids,
        selectedIds: ids,
        onDrop: (drag, _) => drops.add(drag),
      );
      for (final update in [
        board(data: projection(gameId: 'a-different-game')),
        board(allowed: {ArchiveAdapter.jackFirst.id}),
        board(data: projection(hand: const [ArchiveAdapter.jackFirst])),
        board(
          data: projection(
            pending: const PendingView(
              title: 'Waiting for authority',
              detail: '',
              responders: [],
              responseIndex: 0,
            ),
          ),
        ),
      ]) {
        await t.pumpWidget(board());
        await t.pumpAndSettle();
        final gesture = await mouseDrag(t, card(ArchiveAdapter.jackFirst.id));
        await t.pumpWidget(update);
        await t.pumpAndSettle();
        await drop(
          t,
          gesture,
          find.byKey(const ValueKey('adaptive-draw-pile')),
        );
        expect(
          drops,
          isEmpty,
          reason: 'Reject the original bundle rather than trim revoked cards',
        );
      }
    },
  );

  testWidgets(
    'card drag is read only by default and never sources opponent cards',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      const enemy = VisibleCard(
        id: 'opponent-visible',
        rank: '9',
        suit: CardSuit.spades,
        copy: 1,
      );
      var drops = 0;
      await t.pumpWidget(host(onDrop: (_, __) => drops++));
      await t.pumpAndSettle();
      await drop(
        t,
        await mouseDrag(t, card(ArchiveAdapter.jackFirst.id)),
        find.byKey(const ValueKey('adaptive-draw-pile')),
      );
      expect(drops, 0);
      // An incorrect host allowlist must not promote public opponent identities
      // into draggable cards. There are no hidden-card source widgets at all.
      await t.pumpWidget(
        host(
          data: projection(opponentExposed: const [enemy]),
          draggableIds: {enemy.id, 'unseen-draw-card'},
          onDrop: (_, __) => drops++,
        ),
      );
      await t.pumpAndSettle();
      await drop(
        t,
        await mouseDrag(
          t,
          find.byKey(const ValueKey('public-preview-opponent-visible')).first,
        ),
        find.byKey(const ValueKey('adaptive-draw-pile')),
      );
      expect(drops, 0);
      expect(
        find.byKey(const ValueKey('adaptive-select-unseen-draw-card')),
        findsNothing,
      );
    },
  );

  testWidgets('card drag chooses nested public card and formation targets', (
    t,
  ) async {
    await t.binding.setSurfaceSize(const Size(1025, 948));
    addTearDown(() => t.binding.setSurfaceSize(null));
    final drops = <(TableCardDrag, TableDropTarget)>[];
    await t.pumpWidget(
      host(
        draggableIds: {
          ArchiveAdapter.jackFirst.id,
          ArchiveAdapter.jackSecond.id,
        },
        selectedIds: {ArchiveAdapter.jackSecond.id},
        onDrop: (drag, target) => drops.add((drag, target)),
      ),
    );
    await t.pumpAndSettle();
    await drop(
      t,
      await mouseDrag(t, card(ArchiveAdapter.jackFirst.id)),
      find.byKey(ValueKey('public-preview-${ArchiveAdapter.queenSecond.id}')),
    );
    expect(drops.single.$1.cardIds, [
      ArchiveAdapter.jackFirst.id,
    ], reason: 'Dragging an unselected source excludes the existing selection');
    expect(drops.single.$2.zone, TableDropZone.card);
    expect(drops.single.$2.cardId, ArchiveAdapter.queenSecond.id);
    expect(drops.single.$2.formationId, ArchiveAdapter.formation.id);
    expect(drops.single.$2.seat, 1);
    await drop(
      t,
      await mouseDrag(t, card(ArchiveAdapter.jackFirst.id)),
      find.text(ArchiveAdapter.formation.title),
    );
    expect(drops.last.$2.zone, TableDropZone.formation);
    expect(drops.last.$2.formationId, ArchiveAdapter.formation.id);
  });

  testWidgets('own public card drag works in preview and full chooser', (
    t,
  ) async {
    await t.binding.setSurfaceSize(const Size(1025, 948));
    addTearDown(() => t.binding.setSurfaceSize(null));
    final drops = <(TableCardDrag, TableDropTarget)>[];
    await t.pumpWidget(
      host(
        draggableIds: {ArchiveAdapter.second.id},
        onDrop: (drag, target) => drops.add((drag, target)),
      ),
    );
    await t.pumpAndSettle();
    await drop(
      t,
      await mouseDrag(
        t,
        find.byKey(ValueKey('public-preview-${ArchiveAdapter.second.id}')),
      ),
      find.byKey(const ValueKey('adaptive-hand-viewport')),
    );
    expect(drops.single.$1.cardIds, [ArchiveAdapter.second.id]);
    expect(drops.single.$2.zone, TableDropZone.hand);
    await openPile(t);
    final gesture = await mouseDrag(t, card(ArchiveAdapter.second.id));
    await t.pumpAndSettle();
    expect(
      t
          .widget<Opacity>(
            find
                .ancestor(
                  of: find.text('Close public cards'),
                  matching: find.byType(Opacity),
                )
                .first,
          )
          .opacity,
      0,
      reason: 'Hide the chooser without canceling its pointer',
    );
    await drop(t, gesture, find.byKey(const ValueKey('adaptive-draw-pile')));
    expect(drops, hasLength(2));
    expect(drops.last.$2.zone, TableDropZone.drawPile);
    expect(find.text('Close public cards'), findsNothing);
  });

  testWidgets('canceling a chooser card drag restores its inspection surface', (
    t,
  ) async {
    await t.binding.setSurfaceSize(const Size(1025, 948));
    addTearDown(() => t.binding.setSurfaceSize(null));
    var drops = 0;
    await t.pumpWidget(
      host(
        draggableIds: {ArchiveAdapter.second.id},
        onDrop: (_, __) => drops++,
      ),
    );
    await t.pumpAndSettle();
    await openPile(t);
    final gesture = await mouseDrag(t, card(ArchiveAdapter.second.id));
    await t.pumpAndSettle();
    final dialogOpacity = find
        .ancestor(
          of: find.text('Close public cards'),
          matching: find.byType(Opacity),
        )
        .first;
    expect(t.widget<Opacity>(dialogOpacity).opacity, 0);
    await gesture.cancel();
    await t.pumpAndSettle();
    expect(t.widget<Opacity>(dialogOpacity).opacity, 1);
    expect(drops, 0);
    await t.tap(find.text('Close public cards'));
    await t.pumpAndSettle();
    expect(find.text('Close public cards'), findsNothing);
    expect(t.takeException(), isNull);
  });

  testWidgets(
    'enabled card drags retain tap keyboard inspect focus and touch scroll',
    (t) async {
      await t.binding.setSurfaceSize(const Size(390, 844));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final cards = [
        for (var i = 0; i < 12; i++)
          VisibleCard(
            id: 'scroll-card-$i',
            rank: '8',
            suit: CardSuit.clubs,
            copy: i + 1,
          ),
      ];
      final selected = <String>[];
      final inspected = <String>[];
      var drops = 0;
      Widget board({bool enabled = true}) => host(
        data: projection(hand: cards),
        draggableIds: enabled ? cards.map((card) => card.id).toSet() : {},
        onDrop: (_, __) => drops++,
        select: selected.add,
        inspect: (card) => inspected.add(card.id),
      );
      await t.pumpWidget(board());
      await t.pumpAndSettle();
      final first = card(cards.first.id);
      await t.tapAt(exposedPoint(t, first));
      await t.pump(const Duration(milliseconds: 350));
      await t.pumpAndSettle();
      expect(selected, [cards.first.id]);
      final node = focus(t, first);
      node.requestFocus();
      await t.pump();
      await t.pumpWidget(board(enabled: false));
      await t.pumpAndSettle();
      expect(focus(t, first), same(node));
      expect(node.hasFocus, isTrue);
      await t.pumpWidget(board());
      await t.pumpAndSettle();
      await t.sendKeyEvent(LogicalKeyboardKey.space);
      await t.pumpAndSettle();
      expect(selected, [cards.first.id, cards.first.id]);
      await t.sendKeyEvent(LogicalKeyboardKey.keyI);
      await t.pumpAndSettle();
      expect(inspected, [cards.first.id]);
      await twice(t, first);
      expect(inspected, [cards.first.id, cards.first.id]);
      // A stationary hold must not become a delayed selection or a hand drop.
      final hold = await t.startGesture(exposedPoint(t, first));
      await t.pump(const Duration(milliseconds: 600));
      await hold.up();
      await t.pumpAndSettle();
      expect(drops, 0);
      expect(selected, hasLength(2));
      final horizontal = find.descendant(
        of: find.byKey(const PageStorageKey('hand-suit-scroll-clubs')),
        matching: find.byType(Scrollable),
      );
      final scroll = t.state<ScrollableState>(horizontal).position;
      final before = scroll.pixels;
      await t.dragFrom(
        exposedPoint(t, card(cards[3].id)),
        const Offset(-160, 0),
      );
      await t.pumpAndSettle();
      expect(scroll.pixels, greaterThan(before));
      expect(drops, 0);
      expect(t.takeException(), isNull);
    },
  );

  testWidgets(
    'mouse card drag prepares only current authorized selected cards',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1025, 948));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final drops = <(TableCardDrag, TableDropTarget)>[];
      var selections = 0;
      final ids = {ArchiveAdapter.jackFirst.id, ArchiveAdapter.jackSecond.id};
      await t.pumpWidget(
        host(
          draggableIds: {...ids, 'hidden-card'},
          selectedIds: {...ids, 'hidden-card'},
          onDrop: (drag, target) => drops.add((drag, target)),
          select: (_) => selections++,
        ),
      );
      await t.pumpAndSettle();
      final gesture = await t.startGesture(
        exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
        kind: PointerDeviceKind.mouse,
      );
      await gesture.moveBy(const Offset(0, -24));
      await t.pump();
      await gesture.moveTo(
        t.getCenter(find.byKey(const ValueKey('adaptive-seat-archive-seat-b'))),
      );
      await t.pump();
      await gesture.up();
      await t.pumpAndSettle();
      expect(drops, hasLength(1));
      expect(drops.single.$1.gameId, ArchiveAdapter().projection.gameId);
      expect(drops.single.$1.cardIds, unorderedEquals(ids));
      expect(drops.single.$2.zone, TableDropZone.seat);
      expect(drops.single.$2.seat, 2);
      expect(
        selections,
        0,
        reason: 'A drag prepares a draft, not a click or move',
      );
      expect(
        () => drops.single.$1.cardIds.add('hidden-card'),
        throwsUnsupportedError,
      );
    },
  );

  testWidgets('touch hold drag reaches neutral draw pile without selecting', (
    t,
  ) async {
    await t.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => t.binding.setSurfaceSize(null));
    TableDropTarget? target;
    var selections = 0;
    await t.pumpWidget(
      host(
        draggableIds: {ArchiveAdapter.jackFirst.id},
        onDrop: (_, value) => target = value,
        select: (_) => selections++,
      ),
    );
    await t.pumpAndSettle();
    final gesture = await t.startGesture(
      exposedPoint(t, card(ArchiveAdapter.jackFirst.id)),
    );
    await t.pump(const Duration(milliseconds: 400));
    await gesture.moveTo(
      t.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
    );
    await t.pump();
    await gesture.up();
    await t.pumpAndSettle();
    expect(target?.zone, TableDropZone.drawPile);
    expect(target?.cardId, isNull);
    expect(selections, 0);
  });

  testWidgets('artwork card touch hold past long press reaches draw pile', (
    t,
  ) async {
    await t.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => t.binding.setSurfaceSize(null));
    TableDropTarget? target;
    var selections = 0, activations = 0;
    await t.pumpWidget(
      host(
        cardDriven: true,
        showArt: true,
        activate: (_) => activations++,
        draggableIds: {ArchiveAdapter.jackFirst.id},
        onDrop: (_, value) => target = value,
        select: (_) => selections++,
      ),
    );
    await t.pumpAndSettle();
    final from =
        t.getTopLeft(card(ArchiveAdapter.jackFirst.id)) + const Offset(24, 24);
    final gesture = await t.startGesture(from, kind: PointerDeviceKind.touch);
    await t.pump(const Duration(milliseconds: 600));
    final destination = t.getCenter(
      find.byKey(const ValueKey('adaptive-draw-pile')),
    );
    for (var step = 1; step <= 12; step++) {
      await gesture.moveTo(Offset.lerp(from, destination, step / 12)!);
      await t.pump(const Duration(milliseconds: 16));
    }
    await gesture.up();
    await t.pumpAndSettle();
    expect(target?.zone, TableDropZone.drawPile);
    expect(selections, 0);
    expect(activations, 0);
  });

  testWidgets(
    'open actions retain readable painted public cards and a full hand target',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      const exposed = VisibleCard(
        id: 'readable-public-club',
        rank: '8',
        suit: CardSuit.clubs,
        copy: 1,
      );
      final handCards = [
        for (var i = 0; i < 17; i++)
          VisibleCard(
            id: 'readable-hand-$i',
            rank: '${i % 9 + 2}',
            suit: CardSuit.values[i % 4],
            copy: i ~/ 9 + 1,
          ),
      ];
      final data = projection(
        exposed: const [exposed],
        combos: const [],
        hand: handCards,
      );
      var submissions = 0;
      Rect painted(Finder target) {
        final box = t.renderObject<RenderBox>(target);
        return MatrixUtils.transformRect(
          box.getTransformTo(null),
          Offset.zero & box.size,
        );
      }

      for (final size in const [
        Size(1025, 948),
        Size(1356, 948),
        Size(1440, 948),
        Size(1440, 1400),
      ]) {
        final width = size.width;
        await t.pumpWidget(const SizedBox.shrink());
        await t.binding.setSurfaceSize(size);
        await t.pumpWidget(
          host(
            appBar: true,
            data: data,
            actions: Column(
              children: [
                const Text('Prepare your action'),
                const SizedBox(height: 240),
                FilledButton(
                  onPressed: () => submissions++,
                  child: const Text('Submit action'),
                ),
              ],
            ),
            footer: CgmsGameActions(
              primary: FilledButton(
                onPressed: () {},
                child: const Text('End turn'),
              ),
              secondary: OutlinedButton(
                onPressed: () {},
                child: const Text('Trade cards'),
              ),
            ),
          ),
        );
        await t.tap(find.byTooltip('Choose an action'));
        await t.pumpAndSettle();
        final public = painted(
          find.byKey(const ValueKey('public-preview-readable-public-club')),
        );
        expect(
          public.width,
          greaterThanOrEqualTo(64 - .01),
          reason: 'Painted card width at $width',
        );
        if (size.height > 948) {
          expect(
            public.width,
            greaterThan(64),
            reason: 'Public cards grow with ample height.',
          );
        }
        final title = find
            .descendant(of: pile(), matching: find.byType(Text))
            .first;
        expect(
          painted(title).height,
          greaterThanOrEqualTo(14),
          reason: 'Readable public metadata at $width',
        );
        final hand = t.getRect(
          find.byKey(const ValueKey('adaptive-hand-viewport')),
        );
        final first = t.getRect(card('readable-hand-0'));
        expect(first.intersect(hand).height, closeTo(first.height, .01));
        final control = find.widgetWithText(FilledButton, 'Submit action');
        await t.ensureVisible(control);
        await t.pumpAndSettle();
        expect(control.hitTestable(), findsOneWidget);
        expect(submissions, 0);
        expect(t.takeException(), isNull);
      }
    },
  );

  testWidgets('open actions retain the public table and adjacent usable hand', (
    t,
  ) async {
    addTearDown(() => t.binding.setSurfaceSize(null));
    for (final size in const [
      Size(1356, 948),
      Size(768, 1024),
      Size(390, 844),
      Size(844, 390),
    ]) {
      for (final scale in [1.0, 2.0]) {
        await t.pumpWidget(const SizedBox.shrink());
        await t.binding.setSurfaceSize(size);
        await t.pumpWidget(
          host(
            scale: scale,
            rtl: scale == 2,
            actions: const Column(
              children: [Text('Action controls'), SizedBox(height: 900)],
            ),
          ),
        );
        await t.tap(find.byTooltip('Choose an action'));
        await t.pumpAndSettle();
        final table = find.byKey(const ValueKey('adaptive-public-square'));
        expect(table, findsOneWidget, reason: '$size / $scale');
        final public = t.getRect(table);
        final hand = t.getRect(
          find.byKey(const ValueKey('adaptive-hand-viewport')),
        );
        expect(pile().hitTestable(), findsOneWidget);
        expect(public.height, greaterThanOrEqualTo(96));
        expect(hand.top, closeTo(public.bottom + 4, .01));
        expect(
          t.getRect(card(ArchiveAdapter.jackFirst.id)).intersect(hand).height,
          greaterThanOrEqualTo(48),
        );
        expect(find.text('Action controls').hitTestable(), findsOneWidget);
        expect(t.takeException(), isNull);
      }
    }
  });

  testWidgets('action workspace uses header toggle without a redundant strip', (
    t,
  ) async {
    addTearDown(() => t.binding.setSurfaceSize(null));
    const marker = ValueKey('action-content-start');
    await t.pumpWidget(host(actions: const SizedBox(key: marker, height: 100)));
    for (final width in [1440.0, 768.0, 319.0]) {
      await t.binding.setSurfaceSize(Size(width, 900));
      await t.pumpAndSettle();
      final opener = find.byTooltip('Choose an action');
      final toggleFocus = t
          .widget<IconButton>(
            find.byWidgetPredicate(
              (widget) =>
                  widget is IconButton && widget.tooltip == 'Choose an action',
            ),
          )
          .focusNode;
      await t.tap(opener);
      await t.pumpAndSettle();
      expect(find.text('Choose an action'), findsNothing);
      final closer = find.byTooltip('Close actions');
      expect(closer, findsOneWidget);
      final header = t.getRect(find.byKey(const ValueKey('game-header')));
      final target = t.getRect(closer);
      expect(header.contains(target.center), isTrue);
      expect(target.width, greaterThanOrEqualTo(48));
      expect(target.height, greaterThanOrEqualTo(48));
      final public = t.getRect(
        find.byKey(const ValueKey('adaptive-public-square')),
      );
      expect(public.top, closeTo(header.bottom, .01));
      expect(
        t.getRect(find.byKey(const ValueKey('adaptive-hand-viewport'))).bottom,
        closeTo(t.getRect(find.byKey(marker)).top - 4, .01),
        reason: 'Short action content must return unused space to the hand.',
      );
      expect(t.getRect(find.byKey(marker)).height, 100);
      toggleFocus!.requestFocus();
      await t.pumpAndSettle();
      await t.sendKeyEvent(LogicalKeyboardKey.space);
      await t.pumpAndSettle();
      expect(find.byTooltip('Choose an action'), findsOneWidget);
      expect(toggleFocus.hasFocus, isTrue);
      expect(find.byKey(marker).hitTestable(), findsNothing);
      expect(t.takeException(), isNull);
    }
  });

  testWidgets(
    'action workspace retains draft caret focus and scroll during ordinary updates',
    (t) async {
      await t.binding.setSurfaceSize(const Size(1440, 900));
      addTearDown(() => t.binding.setSurfaceSize(null));
      Widget subject() => host(
        actionsIdentity: 'command-form',
        actions: const _RetainedActionWorkspace(),
      );
      await t.pumpWidget(subject());
      await t.tap(find.byTooltip('Choose an action'));
      await t.pumpAndSettle();
      final workspace = find.byType(_RetainedActionWorkspace);
      final retained = t.state<_RetainedActionWorkspaceState>(workspace);
      final scrolling = find
          .ancestor(of: workspace, matching: find.byType(Scrollable))
          .first;
      retained.draftFocus.requestFocus();
      await t.pumpAndSettle();
      await t.enterText(find.byType(TextField), 'retained123');
      retained.draft.selection = const TextSelection.collapsed(offset: 4);
      await t.pumpAndSettle();
      final scroll = t.state<ScrollableState>(scrolling);
      final offset = scroll.position.pixels;
      expect(offset, greaterThan(0));

      await t.pumpWidget(subject());
      await t.pumpAndSettle();
      expect(t.state<ScrollableState>(scrolling), same(scroll));
      expect(scroll.position.pixels, closeTo(offset, .01));
      for (final width in [1050.0, 1049.0, 600.0, 599.0, 390.0, 1440.0]) {
        await t.binding.setSurfaceSize(Size(width, 900));
        await t.pumpAndSettle();
        expect(
          t.state<_RetainedActionWorkspaceState>(workspace),
          same(retained),
        );
        expect(t.state<ScrollableState>(scrolling), same(scroll));
        expect(retained.draft.text, 'retained123');
        expect(
          retained.draft.selection,
          const TextSelection.collapsed(offset: 4),
        );
        expect(retained.draftFocus.hasPrimaryFocus, isTrue);
        expect(scroll.position.pixels, greaterThan(0));
        expect(find.byType(TextField).hitTestable(), findsOneWidget);
        expect(t.takeException(), isNull);
      }
    },
  );

  testWidgets(
    'action workspace meaning replaces scroll viewport and retains action state',
    (t) async {
      // Browser hit targets and Flutter layout use the same logical-pixel units.
      t.view.devicePixelRatio = 1;
      addTearDown(t.view.resetDevicePixelRatio);
      await t.binding.setSurfaceSize(const Size(1440, 900));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final semantics = t.ensureSemantics();
      try {
        var stays = 0;
        Widget subject(String identity, {bool shortPrompt = false}) => host(
          actionsIdentity: identity,
          actions: _RetainedActionWorkspace(
            shortPrompt: shortPrompt,
            onStay: () => stays++,
          ),
        );
        await t.pumpWidget(subject('command-form'));
        await t.tap(find.byTooltip('Choose an action'));
        await t.pumpAndSettle();
        final workspace = find.byType(_RetainedActionWorkspace);
        final retained = t.state<_RetainedActionWorkspaceState>(workspace);
        retained.draft.value = const TextEditingValue(
          text: 'retained123',
          selection: TextSelection.collapsed(offset: 4),
        );
        final scrolling = find
            .ancestor(of: workspace, matching: find.byType(Scrollable))
            .first;
        final original = t.state<ScrollableState>(scrolling);
        original.position.jumpTo(314);
        await t.pumpAndSettle();
        expect(original.position.pixels, 314);

        // Equal-height content prevents Flutter's normal extent clamp from making
        // this pass accidentally. A new workspace must own a fresh scroll role.
        await t.pumpWidget(subject('round-choice'));
        await t.pumpAndSettle();
        final replacement = t.state<ScrollableState>(scrolling);
        expect(replacement, isNot(same(original)));
        expect(replacement.position.pixels, 0);
        expect(
          t.state<_RetainedActionWorkspaceState>(workspace),
          same(retained),
        );
        expect(retained.draft.text, 'retained123');
        expect(
          retained.draft.selection,
          const TextSelection.collapsed(offset: 4),
        );

        await t.pumpWidget(subject('command-form'));
        await t.pumpAndSettle();
        expect(t.state<ScrollableState>(scrolling).position.pixels, 314);
        await t.pumpWidget(subject('round-choice', shortPrompt: true));
        await t.pumpAndSettle();
        expect(
          t.state<_RetainedActionWorkspaceState>(workspace),
          same(retained),
        );
        final viewport = t.getRect(scrolling);
        for (final label in ['Stay in game', 'Leave game']) {
          final button = find.ancestor(
            of: find.text(label),
            matching: find.byWidgetPredicate((w) => w is ButtonStyleButton),
          );
          final painted = t.getRect(button);
          var node = t.getSemantics(button);
          var semanticBounds = node.rect;
          while (true) {
            final transform = node.transform;
            if (transform != null) {
              semanticBounds = MatrixUtils.transformRect(
                transform,
                semanticBounds,
              );
            }
            final parent = node.parent;
            if (parent == null) break;
            node = parent;
          }
          expect(semanticBounds.left, closeTo(painted.left, .01));
          expect(semanticBounds.top, closeTo(painted.top, .01));
          expect(semanticBounds.size, painted.size);
          expect(painted.width, greaterThanOrEqualTo(48));
          expect(painted.height, greaterThanOrEqualTo(48));
          expect(painted.intersect(viewport), painted);
          expect(button.hitTestable(), findsOneWidget);
        }
        await t.tap(find.widgetWithText(FilledButton, 'Stay in game'));
        await t.pumpAndSettle();
        expect(stays, 1);
        expect(retained.draft.text, 'retained123');
        expect(t.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets('game dock retains focus and complete targets through reflow', (
    t,
  ) async {
    // Exercise the bundled face used by the application. The square Ahem test
    // font turns even a single Trade word into an artificial three-line label.
    final loader = FontLoader('packages/cgms_ui/Atkinson Hyperlegible Next');
    for (final asset in [
      'AtkinsonHyperlegibleNext-Regular.ttf',
      'AtkinsonHyperlegibleNext-Bold.ttf',
    ]) {
      loader.addFont(rootBundle.load('packages/cgms_ui/assets/fonts/$asset'));
    }
    await loader.load();
    addTearDown(() => t.binding.setSurfaceSize(null));
    final primaryFocus = FocusNode();
    addTearDown(primaryFocus.dispose);
    final semantics = t.ensureSemantics();
    try {
      var submitted = 0;
      for (final rtl in [false, true]) {
        for (final scale in [1.0, 2.0]) {
          await t.binding.setSurfaceSize(const Size(390, 844));
          await t.pumpWidget(
            host(
              rtl: rtl,
              scale: scale,
              footer: CgmsGameActions(
                primary: FilledButton(
                  focusNode: primaryFocus,
                  onPressed: () => submitted++,
                  child: const Text('End turn'),
                ),
                secondary: OutlinedButton(
                  onPressed: () {},
                  child: const Text('Trade cards'),
                ),
                utilities: [
                  IconButton(
                    onPressed: () {},
                    tooltip: 'History',
                    icon: const Icon(Icons.history),
                  ),
                ],
              ),
            ),
          );
          await t.pumpAndSettle();
          primaryFocus.requestFocus();
          await t.pumpAndSettle();
          final primary = find.widgetWithText(FilledButton, 'End turn');
          final identity = t.getSemantics(primary).id;
          for (final size in const [
            Size(319, 568),
            Size(599, 844),
            Size(600, 844),
            Size(1049, 948),
            Size(1050, 948),
            Size(1440, 948),
            Size(844, 390),
            Size(600, 844),
            Size(599, 844),
            Size(390, 844),
          ]) {
            await t.binding.setSurfaceSize(size);
            await t.pumpAndSettle();
            final footer = t.getRect(
              find.byKey(const ValueKey('board-footer')),
            );
            for (final control in [
              primary,
              find.widgetWithText(OutlinedButton, 'Trade cards'),
              find.byTooltip('History'),
            ]) {
              final rect = t.getRect(control);
              expect(
                rect.height,
                greaterThanOrEqualTo(48),
                reason: '$size $scale $rtl',
              );
              expect(rect.width, greaterThanOrEqualTo(48));
              expect(rect.intersect(footer), rect, reason: '$size $scale $rtl');
              final label = find.descendant(
                of: control,
                matching: find.byType(Text),
              );
              if (label.evaluate().isNotEmpty) {
                final paragraph = t.renderObject<RenderBox>(label);
                expect(
                  paragraph.getMaxIntrinsicHeight(paragraph.size.width),
                  lessThanOrEqualTo(paragraph.size.height),
                );
              }
            }
            expect(primaryFocus.hasFocus, isTrue);
            expect(t.getSemantics(primary).id, identity);
            expect(submitted, 0);
            expect(t.takeException(), isNull);
          }
        }
      }
      await t.sendKeyEvent(LogicalKeyboardKey.enter);
      await t.pumpAndSettle();
      expect(submitted, 1);
    } finally {
      semantics.dispose();
    }
  });

  testWidgets(
    'public piles use the hand card scale in full-width roomy quadrants',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      final free = [
        for (final rank in [
          '2',
          '3',
          '4',
          '5',
          '6',
          '7',
          '8',
          '9',
          '10',
          'J',
          'Q',
        ])
          VisibleCard(
            id: 'wide-diamond-$rank',
            rank: rank,
            suit: CardSuit.diamonds,
            copy: 1,
          ),
        const VisibleCard(
          id: 'wide-club',
          rank: '8',
          suit: CardSuit.clubs,
          copy: 1,
        ),
      ];
      final kings = List.generate(
        5,
        (i) => VisibleCard(
          id: 'wide-king-$i',
          rank: 'K',
          suit: CardSuit.values[i % 4],
          copy: i ~/ 4 + 1,
        ),
      );
      final data = projection(
        exposed: [...free, ...kings],
        combos: [
          ComboView(
            id: 'wide-kings',
            title: '5 kings',
            stateLabel: 'Underground',
            description: 'Authorized public pile',
            cards: kings,
          ),
        ],
      );
      for (final rtl in [false, true]) {
        for (final art in [false, true]) {
          await t.binding.setSurfaceSize(const Size(821, 948));
          await t.pumpWidget(host(data: data, rtl: rtl, showArt: art));
          await t.pumpAndSettle();
          final state = t.state(find.byType(CgmsAdaptiveTable));
          for (final width in [821.0, 1440.0, 821.0]) {
            await t.binding.setSurfaceSize(Size(width, 948));
            await t.pumpAndSettle();
            final handWidth = t.getSize(card(data.hand.first.id)).width;
            final seat = t.getRect(
              find.byKey(const ValueKey('adaptive-seat-archive-seat-a')),
            );
            for (final face in [...free, ...kings]) {
              final box = t.renderObject<RenderBox>(
                find.byKey(ValueKey('public-preview-${face.id}')),
              );
              final painted = MatrixUtils.transformRect(
                box.getTransformTo(null),
                Offset.zero & box.size,
              );
              expect(
                painted.width,
                greaterThanOrEqualTo(handWidth * .9),
                reason: '${face.id} at $width',
              );
              expect(painted.width, lessThanOrEqualTo(handWidth + .01));
              expect(
                painted.width / painted.height,
                closeTo(face.artworkAspectRatio, .01),
              );
              expect(painted.intersect(seat).size, painted.size);
            }
            final table = t.getRect(
              find.byKey(const ValueKey('adaptive-public-square')),
            );
            final hand = t.getRect(
              find.byKey(const ValueKey('adaptive-hand-viewport')),
            );
            expect(table.left, closeTo(hand.left, .01));
            expect(table.right, closeTo(hand.right, .01));
            expect(hand.top - table.bottom, closeTo(4, .01));
            expect(t.state(find.byType(CgmsAdaptiveTable)), same(state));
            expect(t.takeException(), isNull);
          }
        }
      }
    },
  );

  testWidgets(
    'full-width public table and painted cards adapt with the hand across live widths',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      const exposed = VisibleCard(
        id: 'public-growth-club',
        rank: '8',
        suit: CardSuit.clubs,
        copy: 1,
      );
      final data = projection(exposed: const [exposed], combos: const []);
      final publicCard = find.byKey(
        const ValueKey('public-preview-public-growth-club'),
      );
      Rect paintedBounds(Finder finder) {
        final box = t.renderObject<RenderBox>(finder);
        return MatrixUtils.transformRect(
          box.getTransformTo(null),
          Offset.zero & box.size,
        );
      }

      for (final art in [false, true]) {
        for (final scale in [1.0, 2.0]) {
          await t.binding.setSurfaceSize(const Size(319, 948));
          await t.pumpWidget(
            host(data: data, scale: scale, rtl: scale == 2, showArt: art),
          );
          await t.pumpAndSettle();
          final board = t.state(find.byType(CgmsAdaptiveTable));
          final widths = <double, double>{};
          final faces = <double, Size>{};
          for (final width in [
            319.0,
            390.0,
            821.0,
            1440.0,
            821.0,
            390.0,
            319.0,
          ]) {
            await t.binding.setSurfaceSize(Size(width, 948));
            await t.pumpAndSettle();
            expect(t.state(find.byType(CgmsAdaptiveTable)), same(board));
            final square = t.getRect(
              find.byKey(const ValueKey('adaptive-public-square')),
            );
            final face = paintedBounds(publicCard);
            final hand = t.getRect(
              find.byKey(const ValueKey('adaptive-hand-viewport')),
            );
            expect(square.left, closeTo(hand.left, .01));
            expect(square.right, closeTo(hand.right, .01));
            expect(hand.top - square.bottom, closeTo(4, .01));
            expect(
              t.getRect(card(data.hand.first.id)).intersect(hand).height,
              greaterThanOrEqualTo(48),
            );
            if (widths.containsKey(width)) {
              expect(square.width, closeTo(widths[width]!, .01));
              expect(face.size.width, closeTo(faces[width]!.width, .01));
            }
            widths[width] = square.width;
            faces[width] = face.size;
            expect(t.takeException(), isNull);
          }
          for (final pair in [
            (319.0, 390.0),
            (390.0, 821.0),
            (821.0, 1440.0),
          ]) {
            expect(widths[pair.$2], greaterThan(widths[pair.$1]!));
            // Once public cards reach the shared hand scale, extra horizontal
            // room widens the quadrants instead of oversizing their cards.
            if (pair.$1 < 821) {
              expect(faces[pair.$2]!.width, greaterThan(faces[pair.$1]!.width));
              expect(
                faces[pair.$2]!.height,
                greaterThan(faces[pair.$1]!.height),
              );
            } else {
              expect(
                faces[pair.$2]!.width,
                closeTo(faces[pair.$1]!.width, .01),
              );
              expect(
                faces[pair.$2]!.height,
                closeTo(faces[pair.$1]!.height, .01),
              );
            }
          }
        }
      }
    },
  );

  testWidgets(
    'bounded actions preserve draft caret and leave the hand usable',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      final controller = TextEditingController();
      final node = FocusNode();
      addTearDown(controller.dispose);
      addTearDown(node.dispose);
      var commands = 0;
      await t.binding.setSurfaceSize(const Size(319, 948));
      await t.pumpWidget(
        host(
          actions: Column(
            children: [
              TextField(controller: controller, focusNode: node),
              const SizedBox(height: 900),
              FilledButton(
                onPressed: () => commands++,
                child: const Text('Submit retained draft'),
              ),
            ],
          ),
        ),
      );
      final pane = find.byKey(const ValueKey('adaptive-hand-viewport'));
      await t.tap(find.byTooltip('Choose an action'));
      await t.pumpAndSettle();
      expect(
        t.getRect(pane).top,
        closeTo(
          t
                  .getRect(find.byKey(const ValueKey('adaptive-public-square')))
                  .bottom +
              4,
          .01,
        ),
      );
      await t.enterText(find.byType(TextField), 'untouched draft');
      controller.selection = const TextSelection(
        baseOffset: 2,
        extentOffset: 7,
      );
      for (final size in const [
        Size(599, 900),
        Size(600, 900),
        Size(1049, 900),
        Size(1050, 900),
        Size(844, 390),
        Size(319, 948),
      ]) {
        await t.binding.setSurfaceSize(size);
        await t.pumpAndSettle();
        expect(node.hasPrimaryFocus, isTrue);
        expect(controller.text, 'untouched draft');
        expect(
          controller.selection,
          const TextSelection(baseOffset: 2, extentOffset: 7),
        );
        expect(
          t
              .getRect(card(ArchiveAdapter.jackFirst.id))
              .intersect(t.getRect(pane))
              .height,
          greaterThanOrEqualTo(48),
        );
      }
      await t.tap(find.byTooltip('Close actions'));
      await t.pumpAndSettle();
      expect(node.canRequestFocus, isFalse);
      expect(
        t
            .widget<IconButton>(find.widgetWithIcon(IconButton, Icons.tune))
            .focusNode!
            .hasPrimaryFocus,
        isTrue,
      );
      for (var i = 0; i < 20; i++) {
        await t.sendKeyEvent(LogicalKeyboardKey.tab);
        await t.pumpAndSettle();
        expect(node.hasFocus, isFalse);
      }
      await t.tap(find.byTooltip('Choose an action'));
      await t.pumpAndSettle();
      expect(controller.text, 'untouched draft');
      expect(
        controller.selection,
        const TextSelection(baseOffset: 2, extentOffset: 7),
      );
      expect(node.canRequestFocus, isTrue);
      expect(commands, 0);
      expect(t.takeException(), isNull);
    },
  );

  testWidgets(
    'full-width public overview has four equal quadrants and an initially visible hand',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final size in const [
        Size(319, 948),
        Size(320, 568),
        Size(390, 844),
        Size(768, 1024),
        Size(1440, 900),
        Size(844, 390),
      ]) {
        for (final scale in [1.0, 2.0]) {
          await t.binding.setSurfaceSize(size);
          await t.pumpWidget(
            host(
              scale: scale,
              rtl: scale == 2,
              actions: Column(
                children: List.generate(
                  30,
                  (i) => TextField(
                    decoration: InputDecoration(labelText: 'Draft $i'),
                  ),
                ),
              ),
            ),
          );
          await t.pumpAndSettle();
          final a = t.getRect(
            find.byKey(const ValueKey('adaptive-seat-archive-seat-a')),
          );
          final b = t.getRect(
            find.byKey(const ValueKey('adaptive-seat-archive-seat-b')),
          );
          final c = t.getRect(
            find.byKey(const ValueKey('adaptive-seat-archive-seat-c')),
          );
          expect(
            a.top,
            closeTo(b.top, .01),
            reason: 'two columns even at $size / $scale',
          );
          final empty = t.getRect(
            find.byKey(const ValueKey('adaptive-seat-empty-3')),
          );
          expect(c.top, closeTo(empty.top, .01));
          for (final cell in [b, c, empty]) {
            expect(cell.width, closeTo(a.width, 1 / 64));
            expect(cell.height, closeTo(a.height, 1 / 64));
          }
          final square = t.getRect(
            find.byKey(const ValueKey('adaptive-public-square')),
          );
          expect(square.left, closeTo(8, .01));
          expect(square.right, closeTo(size.width - 8, .01));
          expect(a.width, closeTo(square.width / 2, .01));
          expect(a.height, closeTo(square.height / 2, .01));
          expect(a.width, greaterThanOrEqualTo(48));
          expect(a.height, greaterThanOrEqualTo(48));
          final handViewport = t.getRect(
            find.byKey(const ValueKey('adaptive-hand-viewport')),
          );
          expect(handViewport.left, closeTo(square.left, .01));
          expect(handViewport.right, closeTo(square.right, .01));
          expect(handViewport.top - square.bottom, closeTo(4, .01));
          final firstCard = t.getRect(card(ArchiveAdapter.jackFirst.id));
          expect(
            firstCard.intersect(handViewport).height,
            greaterThanOrEqualTo(48),
          );
          expect(handViewport.bottom, lessThanOrEqualTo(size.height));
          expect(firstCard.top, greaterThanOrEqualTo(square.bottom));
          expect(t.takeException(), isNull);
        }
      }
    },
  );

  const suitHand = [
    VisibleCard(id: 'club-jack-1', rank: 'J', suit: CardSuit.clubs, copy: 1),
    VisibleCard(id: 'heart-two', rank: '2', suit: CardSuit.hearts, copy: 1),
    VisibleCard(id: 'spade-queen', rank: 'Q', suit: CardSuit.spades, copy: 1),
    VisibleCard(
      id: 'diamond-ten',
      rank: '10',
      suit: CardSuit.diamonds,
      copy: 1,
    ),
    VisibleCard(
      id: 'club-jack-2',
      rank: 'J',
      suit: CardSuit.clubs,
      copy: 2,
      style: CardStyle.drypoint,
    ),
    VisibleCard(
      id: 'heart-three',
      rank: '3',
      suit: CardSuit.hearts,
      copy: 1,
      style: CardStyle.rainGlaze,
    ),
    VisibleCard(
      id: 'club-eight',
      rank: '8',
      suit: CardSuit.clubs,
      copy: 1,
      style: CardStyle.emberGlaze,
    ),
    VisibleCard(id: 'club-king', rank: 'K', suit: CardSuit.clubs, copy: 1),
  ];
  Finder suitRow(CardSuit suit) =>
      find.byKey(ValueKey('hand-suit-${suit.name}'));

  testWidgets('wide hand groups stay compact and centered in both directions', (
    t,
  ) async {
    addTearDown(() => t.binding.setSurfaceSize(null));
    for (final width in [768.0, 1356.0, 1920.0]) {
      for (final rtl in [false, true]) {
        await t.binding.setSurfaceSize(Size(width, 948));
        await t.pumpWidget(
          host(
            data: projection(hand: suitHand),
            rtl: rtl,
          ),
        );
        await t.pumpAndSettle();
        final hearts = t.getRect(suitRow(CardSuit.hearts));
        final spades = t.getRect(suitRow(CardSuit.spades));
        final clubs = t.getRect(suitRow(CardSuit.clubs));
        final first = t.getRect(card('club-jack-1'));
        final last = t.getRect(card('club-king'));
        final fan = first.expandToInclude(last);
        expect(clubs.width, closeTo(fan.width, .01));
        final row = hearts.expandToInclude(spades);
        expect(row.center.dx, closeTo(width / 2, .01));
        expect(
          rtl ? hearts.left - spades.right : spades.left - hearts.right,
          closeTo(8, .01),
        );
        expect(t.takeException(), isNull);
      }
    }
  });

  testWidgets('a lone suit stays centered from the tablet boundary', (t) async {
    addTearDown(() => t.binding.setSurfaceSize(null));
    final clubs = suitHand
        .where((card) => card.suit == CardSuit.clubs)
        .toList();
    for (final width in [600.0, 1050.0, 1356.0]) {
      await t.binding.setSurfaceSize(Size(width, 948));
      await t.pumpWidget(host(data: projection(hand: clubs)));
      await t.pumpAndSettle();
      final group = t.getRect(suitRow(CardSuit.clubs));
      expect(group.center.dx, closeTo(width / 2, .01));
      expect(
        t.widget<Text>(find.text('Your hand · 4')).textAlign,
        TextAlign.center,
      );
      expect(t.takeException(), isNull);
    }
  });

  testWidgets(
    'private suit stacks have uniform contained artwork and compact two-row layout',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final width in [390.0, 768.0, 1440.0]) {
        await t.binding.setSurfaceSize(Size(width, 900));
        await t.pumpWidget(host(data: projection(hand: suitHand)));
        final size = t.getSize(card(suitHand.first.id));
        for (final c in suitHand) {
          expect(
            t.getSize(card(c.id)),
            size,
            reason: 'uniform ${c.id} at $width',
          );
          expect(size.width / size.height, closeTo(4 / 5, .001));
          final painting = t.getSize(find.byKey(ValueKey('card-art-${c.id}')));
          expect(
            painting.width / painting.height,
            closeTo(c.artworkAspectRatio, .001),
          );
          expect(painting.width, lessThanOrEqualTo(size.width));
          expect(painting.height, lessThanOrEqualTo(size.height));
        }
        final hearts = t.getRect(suitRow(CardSuit.hearts));
        final spades = t.getRect(suitRow(CardSuit.spades));
        final diamonds = t.getRect(suitRow(CardSuit.diamonds));
        final clubs = t.getRect(suitRow(CardSuit.clubs));
        expect(hearts.top, spades.top);
        expect(diamonds.top, clubs.top);
        expect(hearts.right, lessThanOrEqualTo(spades.left));
        expect(hearts.bottom, lessThanOrEqualTo(diamonds.top));
        final first = t.getRect(card('club-jack-1'));
        final second = t.getRect(card('club-jack-2'));
        expect(second.left - first.left, greaterThanOrEqualTo(48));
        expect(second.left, lessThan(first.right));
        expect(second.top, first.top);
        expect(t.takeException(), isNull);
      }
    },
  );

  testWidgets(
    'overlapped hand copies retain separate touch inspection and keyboard access in RTL',
    (t) async {
      await t.binding.setSurfaceSize(const Size(390, 500));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final selected = <String>[], inspected = <String>[];
      for (final rtl in [false, true]) {
        await t.pumpWidget(
          host(
            data: projection(hand: suitHand),
            rtl: rtl,
            selectedIds: const {'club-jack-1', 'club-jack-2', 'club-king'},
            select: selected.add,
            inspect: (c) => inspected.add(c.id),
          ),
        );
        for (final id in ['club-jack-1', 'club-jack-2', 'club-king']) {
          await t.ensureVisible(card(id));
          await t.pumpAndSettle();
          final rect = t.getRect(card(id));
          final exposed = Offset(
            rtl ? rect.right - 24 : rect.left + 24,
            rect.center.dy,
          );
          final copy = t.getRect(find.byKey(ValueKey('card-copy-$id')));
          final marker = t.getRect(find.byKey(ValueKey('card-selected-$id')));
          final strip = Rect.fromLTWH(
            rtl ? rect.right - 48 : rect.left,
            rect.top,
            48,
            rect.height,
          );
          expect(strip.contains(copy.center), isTrue);
          expect(strip.contains(marker.center), isTrue);
          final beforeTap = selected.length;
          await t.tapAt(exposed);
          await t.pump(const Duration(milliseconds: 350));
          expect(selected.last, id);
          expect(selected.length, beforeTap + 1);
          final before = selected.length;
          await t.tapAt(exposed);
          await t.pump(const Duration(milliseconds: 50));
          await t.tapAt(exposed);
          await t.pumpAndSettle();
          expect(inspected.last, id);
          expect(selected.length, before);
        }
        final last = card('club-king');
        final node = focus(t, last);
        focus(t, card('club-jack-1')).requestFocus();
        await t.pumpAndSettle();
        for (var i = 0; i < 10 && !node.hasPrimaryFocus; i++) {
          await t.sendKeyEvent(LogicalKeyboardKey.tab);
          await t.pumpAndSettle();
        }
        expect(node.hasPrimaryFocus, isTrue);
        expect(last.hitTestable(), findsOneWidget);
        final group = t.getRect(suitRow(CardSuit.clubs));
        final rect = t.getRect(last);
        expect(rect.left, greaterThanOrEqualTo(group.left - .01));
        expect(rect.right, lessThanOrEqualTo(group.right + .01));
        await t.sendKeyEvent(LogicalKeyboardKey.keyI);
        await t.pumpAndSettle();
        expect(inspected.last, 'club-king');
        for (final width in [599.0, 600.0, 1049.0, 1050.0, 390.0]) {
          await t.binding.setSurfaceSize(Size(width, 500));
          await t.pumpAndSettle();
          expect(node.hasPrimaryFocus, isTrue);
          final viewport = t.getRect(suitRow(CardSuit.clubs));
          final focusedCard = t.getRect(last);
          expect(
            focusedCard.left,
            greaterThanOrEqualTo(viewport.left - .01),
            reason: '$width/$rtl',
          );
          expect(
            focusedCard.right,
            lessThanOrEqualTo(viewport.right + .01),
            reason: '$width/$rtl',
          );
        }
        await t.pumpWidget(
          host(
            data: projection(hand: suitHand),
            rtl: rtl,
            scale: 2,
            select: selected.add,
            inspect: (c) => inspected.add(c.id),
          ),
        );
        await t.pumpAndSettle();
        expect(focus(t, last), same(node));
        expect(node.hasPrimaryFocus, isTrue);
        expect(
          t.getTopLeft(suitRow(CardSuit.spades)).dy,
          greaterThan(t.getTopLeft(suitRow(CardSuit.hearts)).dy),
        );
        expect(t.takeException(), isNull);
      }
    },
  );
  testWidgets(
    'queued focus reveal never moves the board behind a newer modal',
    (t) async {
      await t.binding.setSurfaceSize(const Size(390, 400));
      addTearDown(() => t.binding.setSurfaceSize(null));
      final actionFocus = FocusNode();
      final modalFocus = FocusNode();
      addTearDown(actionFocus.dispose);
      addTearDown(modalFocus.dispose);
      await t.pumpWidget(
        host(
          actions: Column(
            children: [
              TextButton(
                focusNode: actionFocus,
                onPressed: () {},
                child: const Text('Action target'),
              ),
              const SizedBox(height: 1000),
            ],
          ),
        ),
      );
      await t.tap(find.byTooltip('Choose an action'));
      await t.pumpAndSettle();
      final boardScroll = find
          .descendant(
            of: find.byKey(const PageStorageKey('adaptive-decisions')),
            matching: find.byType(Scrollable),
          )
          .first;
      await t.drag(boardScroll, const Offset(0, -900));
      await t.pumpAndSettle();
      final position = t.state<ScrollableState>(boardScroll).position;
      final before = position.pixels;
      actionFocus.requestFocus();
      FocusManager.instance.applyFocusChangesIfNeeded();
      showDialog<void>(
        context: t.element(find.text('Action target')),
        builder: (dialog) => AlertDialog(
          title: const Text('Newer modal'),
          actions: [
            TextButton(
              autofocus: true,
              focusNode: modalFocus,
              onPressed: () => Navigator.of(dialog).pop(),
              child: const Text('Close modal'),
            ),
          ],
        ),
      );
      await t.pumpAndSettle();
      expect(modalFocus.hasPrimaryFocus, isTrue);
      expect(position.pixels, before);
      expect(find.text('Close modal').hitTestable(), findsOneWidget);
      expect(t.takeException(), isNull);
    },
  );

  testWidgets('public card double tap magnifies without selecting', (t) async {
    final selected = <String>[], inspected = <String>[];
    await t.pumpWidget(
      host(select: selected.add, inspect: (c) => inspected.add(c.id)),
    );
    await openPile(t);
    await twice(t, card('archive-card-201'));
    expect(inspected, ['archive-card-201']);
    expect(selected, isEmpty);
    expect(find.text('Close public cards'), findsNothing);
    expect(t.takeException(), isNull);
  });
  testWidgets(
    'phone previews preserve full painting proportions and independent keyboard inspection',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final width in [360.0, 390.0, 599.0]) {
        for (final scale in [1.0, 2.0]) {
          await t.binding.setSurfaceSize(Size(width, 900));
          final selected = <String>[], inspected = <String>[];
          await t.pumpWidget(
            host(
              scale: scale,
              rtl: true,
              select: selected.add,
              inspect: (c) => inspected.add(c.id),
            ),
          );
          final target = card('archive-private-501');
          await t.ensureVisible(target);
          await t.pumpAndSettle();
          final size = t.getSize(
            find.byKey(const ValueKey('card-art-archive-private-501')),
          );
          expect(
            size.width / size.height,
            closeTo(ArchiveAdapter.jackFirst.artworkAspectRatio, .001),
          );
          expect(t.getSize(target).width, greaterThanOrEqualTo(48));
          expect(t.getSize(target).height, greaterThanOrEqualTo(48));
          await t.tap(target);
          await t.pump(const Duration(milliseconds: 350));
          expect(selected, ['archive-private-501']);
          focus(t, target).requestFocus();
          await t.pump();
          await t.sendKeyEvent(LogicalKeyboardKey.keyI);
          await t.pumpAndSettle();
          expect(inspected, ['archive-private-501']);
          expect(selected, ['archive-private-501']);
          expect(t.takeException(), isNull, reason: '$width/$scale');
        }
      }
    },
  );
  testWidgets(
    'hand inspection stays keyboard reachable after scrolling and resizing',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      final inspected = <String>[];
      await t.pumpWidget(host(inspect: (c) => inspected.add(c.id)));
      final target = card('archive-private-502');
      for (final width in [390.0, 600.0, 1050.0, 599.0]) {
        await t.binding.setSurfaceSize(Size(width, 400));
        await t.pumpAndSettle();
        if (find.byTooltip('Close actions').evaluate().isNotEmpty) {
          await t.tap(find.byTooltip('Close actions'));
          await t.pumpAndSettle();
        }
        await t.ensureVisible(pile());
        final node = focus(t, target);
        for (var i = 0; i < 30 && !node.hasPrimaryFocus; i++) {
          await t.sendKeyEvent(LogicalKeyboardKey.tab);
          await t.pumpAndSettle();
        }
        expect(node.hasPrimaryFocus, isTrue, reason: '$width');
        await t.sendKeyEvent(LogicalKeyboardKey.keyI);
        await t.pumpAndSettle();
        expect(inspected.last, 'archive-private-502');
      }
      expect(inspected.length, 4);
    },
  );
  testWidgets(
    'decision drafts and caret survive scrolling and exact class transitions',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      await t.pumpWidget(
        host(
          actions: const TextField(
            decoration: InputDecoration(labelText: 'Private trade draft'),
          ),
        ),
      );
      await t.tap(find.byTooltip('Choose an action'));
      await t.pumpAndSettle();
      await t.ensureVisible(find.byType(TextField));
      await t.enterText(find.byType(TextField), '2/3');
      final edit = t.widget<EditableText>(find.byType(EditableText));
      edit.controller.selection = const TextSelection(
        baseOffset: 1,
        extentOffset: 3,
      );
      for (final width in [
        599.0,
        600.0,
        1049.0,
        1050.0,
        1049.0,
        600.0,
        599.0,
      ]) {
        await t.binding.setSurfaceSize(Size(width, 900));
        await t.pumpAndSettle();
        expect(find.text('2/3'), findsOneWidget);
        expect(edit.focusNode.hasFocus, isTrue);
        expect(
          edit.controller.selection,
          const TextSelection(baseOffset: 1, extentOffset: 3),
        );
        expect(t.takeException(), isNull);
      }
    },
  );
  testWidgets(
    'resizing preserves exact physical card focus without dispatching',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      final selected = <String>[];
      await t.pumpWidget(host(select: selected.add));
      final node = focus(t, card('archive-private-501'))..requestFocus();
      await t.pump();
      for (final width in [
        1050.0,
        1049.0,
        600.0,
        599.0,
        600.0,
        1049.0,
        1050.0,
      ]) {
        await t.binding.setSurfaceSize(Size(width, 900));
        await t.pumpAndSettle();
        expect(node.hasPrimaryFocus, isTrue, reason: '$width');
        expect(selected, isEmpty);
        expect(t.takeException(), isNull);
      }
    },
  );
  testWidgets(
    'forward traversal reaches public group after short RTL resize sequence',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final rtl in [false, true]) {
        await t.pumpWidget(host(rtl: rtl));
        for (final width in [
          390.0,
          599.0,
          600.0,
          768.0,
          1049.0,
          1050.0,
          1440.0,
          844.0,
        ]) {
          await t.binding.setSurfaceSize(Size(width, 390));
          await t.pumpAndSettle();
          focus(t, card('archive-private-502')).requestFocus();
          await t.pump();
          bool reached() => find
              .descendant(
                of: pile(),
                matching: find.byElementPredicate(
                  (e) =>
                      identical(e, FocusManager.instance.primaryFocus?.context),
                ),
              )
              .evaluate()
              .isNotEmpty;
          for (var i = 0; i < 30 && !reached(); i++) {
            await t.sendKeyEvent(LogicalKeyboardKey.tab);
            await t.pumpAndSettle();
          }
          expect(reached(), isTrue, reason: '$width RTL=$rtl');
        }
      }
    },
  );
  testWidgets(
    'overlapped previews expose one accessible group and no invisible focus targets',
    (t) async {
      final semantics = t.ensureSemantics();
      await t.pumpWidget(host());
      expect(find.bySemanticsLabel('8 of Clubs, copy 1'), findsNothing);
      expect(t.getSize(pile()).width, greaterThanOrEqualTo(48));
      expect(
        focus(
          t,
          find.byKey(const ValueKey('public-preview-archive-card-201')),
        ).canRequestFocus,
        isFalse,
      );
      await openPile(t);
      final node = t
          .getSemantics(
            find.descendant(
              of: card('archive-card-201'),
              matching: find.byKey(const ValueKey('select-archive-card-201')),
            ),
          )
          .getSemanticsData();
      expect(node.label, '8 of Clubs, copy 1');
      expect(node.hasAction(SemanticsAction.focus), isTrue);
      semantics.dispose();
    },
  );
  testWidgets(
    'desktop and phone keep same continuous hierarchy and gameplay labels',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final width in [390.0, 768.0, 1440.0]) {
        await t.binding.setSurfaceSize(Size(width, 900));
        await t.pumpWidget(host());
        expect(find.byType(NavigationRail), findsNothing);
        expect(find.byKey(const ValueKey('nav-hand')), findsNothing);
        final public = t.getRect(
          find.byKey(const ValueKey('adaptive-public-square')),
        );
        final hand = t.getRect(
          find.byKey(const ValueKey('adaptive-hand-viewport')),
        );
        expect(public.bottom, lessThanOrEqualTo(hand.top));
        await t.tap(find.byTooltip('Choose an action'));
        await t.pumpAndSettle();
        expect(
          find.text('Authoritative decisions').hitTestable(),
          findsOneWidget,
        );
        final expandedHand = t.getRect(
          find.byKey(const ValueKey('adaptive-hand-viewport')),
        );
        expect(expandedHand.left, hand.left);
        expect(expandedHand.right, hand.right);
        expect(expandedHand.bottom, lessThan(hand.bottom));
        if (width == 390) {
          // The width-limited phone table already fits beside this short form;
          // opening it must preserve that public-card size.
          expect(expandedHand.top, hand.top);
        } else {
          expect(expandedHand.top, lessThan(hand.top));
        }
        expect(
          expandedHand.bottom,
          closeTo(t.getRect(find.text('Authoritative decisions')).top - 4, .01),
        );
        expect(
          expandedHand.top,
          closeTo(
            t
                    .getRect(
                      find.byKey(const ValueKey('adaptive-public-square')),
                    )
                    .bottom +
                4,
            .01,
          ),
        );
        await t.tap(find.byTooltip('Close actions'));
        await t.pumpAndSettle();
        expect(
          t.getRect(find.byKey(const ValueKey('adaptive-hand-viewport'))),
          hand,
        );
        expect(find.text('Imported public position'), findsNWidgets(3));
        expect(find.text(ArchiveAdapter.formation.stateLabel), findsOneWidget);
      }
    },
  );
  testWidgets(
    'public groups preserve physical copies and independent 48 pixel inspection',
    (t) async {
      final selected = <String>[], inspected = <String>[];
      await t.pumpWidget(
        host(select: selected.add, inspect: (c) => inspected.add(c.id)),
      );
      await openPile(t);
      for (final id in ['archive-card-201', 'archive-card-202']) {
        expect(card(id), findsOneWidget);
      }
      final inspect = find.descendant(
        of: card('archive-card-202'),
        matching: find.byType(IconButton),
      );
      expect(t.getSize(inspect).width, greaterThanOrEqualTo(48));
      expect(t.getSize(inspect).height, greaterThanOrEqualTo(48));
      await t.tap(inspect);
      await t.pumpAndSettle();
      expect(inspected, ['archive-card-202']);
      expect(selected, isEmpty);
      await openPile(t);
      await t.tap(card('archive-card-201'));
      await t.pump(const Duration(milliseconds: 350));
      await t.pumpAndSettle();
      expect(selected, ['archive-card-201']);
      expect(find.text('Close public cards'), findsNothing);
    },
  );
  testWidgets(
    'host modal inspector survives chooser dismissal and owns focus',
    (t) async {
      final node = FocusNode(debugLabel: 'modal-inspector');
      addTearDown(node.dispose);
      await t.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Builder(
            builder: (context) => Scaffold(
              body: CgmsAdaptiveTable(
                data: ArchiveAdapter().projection,
                actions: const Text('Decisions'),
                onSelect: (_) {},
                showArt: false,
                onInspect: (_) => showDialog<void>(
                  context: context,
                  builder: (dialog) => AlertDialog(
                    title: const Text('Card inspection'),
                    actions: [
                      TextButton(
                        autofocus: true,
                        focusNode: node,
                        onPressed: () => Navigator.of(dialog).pop(),
                        child: const Text('Close inspection'),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      );
      await openPile(t);
      await t.tap(
        find.descendant(
          of: card('archive-card-201'),
          matching: find.byType(IconButton),
        ),
      );
      await t.pumpAndSettle();
      expect(find.text('Card inspection'), findsOneWidget);
      expect(node.hasPrimaryFocus, isTrue);
      expect(find.text('Close public cards'), findsNothing);
      await t.sendKeyEvent(LogicalKeyboardKey.escape);
      await t.pumpAndSettle();
      expect(find.text('Card inspection'), findsNothing);
      expect(t.takeException(), isNull);
    },
  );
  testWidgets('lone public card keeps its own selectable semantics', (t) async {
    final semantics = t.ensureSemantics();
    await t.pumpWidget(
      host(
        data: projection(exposed: [ArchiveAdapter.first], combos: []),
      ),
    );
    await openPile(t);
    expect(
      t
          .getSemantics(
            find.descendant(
              of: card('archive-card-201'),
              matching: find.byKey(const ValueKey('select-archive-card-201')),
            ),
          )
          .getSemanticsData()
          .label,
      '8 of Clubs, copy 1',
    );
    semantics.dispose();
  });
  testWidgets(
    'Auto keeps physical selection across literal logical boundaries',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      final selected = <String>[];
      await t.pumpWidget(host(select: selected.add));
      await t.ensureVisible(card('archive-private-501'));
      await t.tapAt(exposedPoint(t, card('archive-private-501')));
      await t.pump(const Duration(milliseconds: 350));
      for (final (width, label) in [
        (599.0, 'phone'),
        (600.0, 'tablet'),
        (1049.0, 'tablet'),
        (1050.0, 'desktop'),
        (599.0, 'phone'),
      ]) {
        await t.binding.setSurfaceSize(Size(width, 900));
        await t.pumpAndSettle();
        expect(find.byKey(ValueKey('$label-table')), findsOneWidget);
        expect(selected, ['archive-private-501']);
      }
    },
  );
  testWidgets('Coup remains reachable including the public chooser', (t) async {
    var coups = 0;
    await t.pumpWidget(host(coup: () => coups++));
    await t.tap(find.byKey(const ValueKey('immediate-coup')));
    await openPile(t);
    await t.tap(find.widgetWithText(FilledButton, 'Declare Coup now').last);
    await t.pumpAndSettle();
    expect(coups, 2);
    expect(find.text('External adapter · Ada'), findsOneWidget);
  });
  testWidgets(
    'matrix retains class under 200 percent RTL text and short heights',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      for (final size in [
        const Size(320, 640),
        const Size(360, 800),
        const Size(390, 844),
        const Size(599, 900),
        const Size(600, 900),
        const Size(768, 1024),
        const Size(844, 390),
        const Size(1024, 768),
        const Size(1049, 900),
        const Size(1050, 900),
        const Size(1280, 600),
        const Size(1440, 900),
      ]) {
        await t.binding.setSurfaceSize(size);
        await t.pumpWidget(host(scale: 2, rtl: true, coup: () {}));
        await t.pumpAndSettle();
        expect(
          find.byKey(
            ValueKey(
              '${size.width < 600
                  ? 'phone'
                  : size.width < 1050
                  ? 'tablet'
                  : 'desktop'}-table',
            ),
          ),
          findsOneWidget,
        );
        expect(t.takeException(), isNull, reason: '$size');
      }
    },
  );
  testWidgets(
    'chooser refreshes authorized faces and styles across projection changes',
    (t) async {
      addTearDown(() => t.binding.setSurfaceSize(null));
      var calls = 0;
      final state = ValueNotifier(projection());
      addTearDown(state.dispose);
      await t.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: ValueListenableBuilder<TableViewData>(
              valueListenable: state,
              builder: (_, data, __) => CgmsAdaptiveTable(
                data: data,
                actions: const Text('Actions'),
                onSelect: (_) => calls++,
                onInspect: (_) => calls++,
                showArt: false,
              ),
            ),
          ),
        ),
      );
      await openPile(t);
      final stale = t.widget<CgmsCard>(card('archive-card-201'));
      await t.binding.setSurfaceSize(const Size(390, 844));
      await t.pumpAndSettle();
      expect(card('archive-card-201'), findsOneWidget);
      state.value = projection(
        exposed: [
          const VisibleCard(
            id: 'archive-card-202',
            rank: '8',
            suit: CardSuit.clubs,
            copy: 2,
            style: CardStyle.drypoint,
          ),
        ],
        combos: [],
      );
      await t.pumpAndSettle();
      expect(card('archive-card-201'), findsNothing);
      stale.onTap!();
      stale.onInspect!();
      expect(calls, 0);
      expect(
        t.widget<CgmsCard>(card('archive-card-202')).card.style,
        CardStyle.drypoint,
      );
      state.value = projection(exposed: [], combos: []);
      await t.pumpAndSettle();
      expect(card('archive-card-202'), findsNothing);
      expect(find.text('These cards are no longer exposed.'), findsOneWidget);
    },
  );
  testWidgets('removing the board dismisses its chooser and stale callbacks', (
    t,
  ) async {
    final visible = ValueNotifier(true);
    addTearDown(visible.dispose);
    var calls = 0;
    await t.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: Scaffold(
          body: ValueListenableBuilder<bool>(
            valueListenable: visible,
            builder: (_, show, __) => show
                ? CgmsAdaptiveTable(
                    data: ArchiveAdapter().projection,
                    actions: const Text('Actions'),
                    onSelect: (_) => calls++,
                    onInspect: (_) => calls++,
                    showArt: false,
                  )
                : const Text('Account changed'),
          ),
        ),
      ),
    );
    await openPile(t);
    final stale = t.widget<CgmsCard>(card('archive-card-201'));
    visible.value = false;
    await t.pumpAndSettle();
    expect(find.text('Account changed'), findsOneWidget);
    expect(find.text('Close public cards'), findsNothing);
    expect(card('archive-card-201'), findsNothing);
    stale.onTap!();
    stale.onInspect!();
    expect(calls, 0);
    expect(t.takeException(), isNull);
  });
  testWidgets('compact preview long press and double tap never select', (
    t,
  ) async {
    final selected = <String>[], inspected = <String>[];
    await t.pumpWidget(
      host(select: selected.add, inspect: (c) => inspected.add(c.id)),
    );
    final target = card('archive-private-501');
    await t.ensureVisible(target);
    await t.longPressAt(exposedPoint(t, target));
    await t.pumpAndSettle();
    expect(selected, isEmpty);
    expect(inspected, isEmpty);
    await twice(t, target);
    expect(selected, isEmpty);
    expect(inspected, ['archive-private-501']);
  });
  testWidgets(
    'pending operations disable selection while preserving exact inspection',
    (t) async {
      final selected = <String>[], inspected = <String>[];
      await t.pumpWidget(
        host(
          data: projection(
            pending: const PendingView(
              title: 'Waiting for reply',
              detail: 'Seat 2 must respond',
              responders: ['Seat 2'],
              responseIndex: 0,
            ),
          ),
          select: selected.add,
          inspect: (c) => inspected.add(c.id),
        ),
      );
      await openPile(t);
      final target = card('archive-card-201');
      expect(t.widget<CgmsCard>(target).onTap, isNull);
      await t.tap(
        find.descendant(of: target, matching: find.byType(IconButton)),
      );
      await t.pumpAndSettle();
      expect(inspected, ['archive-card-201']);
      expect(selected, isEmpty);
    },
  );
}
