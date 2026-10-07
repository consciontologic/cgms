import 'dart:ui' show Tristate;

import 'package:cgms_demo/main.dart';
import 'package:cgms_demo/approved_flow.dart';
import 'package:cgms_demo/approved_demo_controller.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart' show DebugSemanticsDumpOrder;
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUpAll(() async {
    for (final entry in {
      'packages/cgms_ui/Barlow Condensed': ['BarlowCondensed-Bold.ttf'],
      'packages/cgms_ui/Atkinson Hyperlegible Next': [
        'AtkinsonHyperlegibleNext-Regular.ttf',
        'AtkinsonHyperlegibleNext-Bold.ttf',
      ],
    }.entries) {
      final loader = FontLoader(entry.key);
      for (final asset in entry.value) {
        loader.addFont(rootBundle.load('packages/cgms_ui/assets/fonts/$asset'));
      }
      await loader.load();
    }
  });
  Future<void> start(
    WidgetTester tester,
    double width, {
    double scale = 1,
    double height = 844,
  }) async {
    tester.view.physicalSize = Size(width, height);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    await tester.pumpWidget(DemoApp(initialTextScale: scale));
    await tester.pumpAndSettle();
  }

  testWidgets(
    'compact header keeps status and tools centered in one responsive row',
    (tester) async {
      await start(tester, 319, height: 948);
      final initial = tester.state(find.byType(CgmsAdaptiveTable));
      final toggle = find.byWidgetPredicate(
        (widget) =>
            widget is IconButton && widget.tooltip == 'Choose an action',
      );
      final focus = tester.widget<IconButton>(toggle).focusNode!;
      focus.requestFocus();
      await tester.pump();
      final labels = ['Round 7 / 13', 'Your turn', 'Auto'];
      double? phoneFont, phoneIcon;
      Finder marker(String name) => find.byWidgetPredicate(
        (widget) => widget is Semantics && widget.properties.identifier == name,
      );
      for (final width in [
        319.0,
        320.0,
        390.0,
        500.0,
        599.0,
        600.0,
        821.0,
        1050.0,
        1440.0,
        1049.0,
        600.0,
        599.0,
        319.0,
      ]) {
        tester.view.physicalSize = Size(width, 948);
        await tester.pumpAndSettle();
        final status = tester.getRect(marker('header-status'));
        final tools = tester.getRect(marker('header-tools'));
        final header = tester.getRect(marker('game-header'));
        expect(
          status.center.dy,
          closeTo(tools.center.dy, .01),
          reason: 'One centered header row at $width',
        );
        expect(
          status.height,
          lessThanOrEqualTo(28),
          reason: 'Round and turn share one line at normal text size',
        );
        expect(
          tester.getRect(find.text(labels[0])).center.dy,
          closeTo(tester.getRect(find.text(labels[1])).center.dy, .01),
        );
        expect(status.right, lessThanOrEqualTo(tools.left));
        for (final label in labels) {
          final target = find.text(label);
          final bounds = tester.getRect(target);
          expect(header.contains(bounds.topLeft), isTrue);
          expect(header.contains(bounds.bottomRight), isTrue);
          final paragraph = tester.renderObject<RenderBox>(target);
          expect(
            paragraph.getMaxIntrinsicHeight(paragraph.size.width),
            lessThanOrEqualTo(paragraph.size.height),
          );
        }
        for (final target in [
          toggle,
          find.byTooltip('Table layout'),
          find.byTooltip('Game menu'),
        ]) {
          final bounds = tester.getRect(target);
          expect(bounds.width, greaterThanOrEqualTo(48));
          expect(bounds.height, greaterThanOrEqualTo(48));
          expect(header.contains(bounds.topLeft), isTrue);
          expect(header.contains(bounds.bottomRight), isTrue);
          expect(
            tester.getSemantics(target).rect.height,
            greaterThanOrEqualTo(48),
          );
        }
        final round = find.descendant(
          of: find.text(labels.first),
          matching: find.byType(RichText),
        );
        final font = tester.widget<RichText>(round).text.style!.fontSize!;
        final icon = tester
            .getSize(find.descendant(of: toggle, matching: find.byType(Icon)))
            .width;
        if (width == 319 && phoneFont == null) {
          phoneFont = font;
          phoneIcon = icon;
        }
        if (width >= 1050) {
          expect(font, greaterThan(phoneFont!));
          expect(icon, greaterThan(phoneIcon!));
        }
        expect(focus.hasFocus, isTrue);
        expect(
          identical(tester.state(find.byType(CgmsAdaptiveTable)), initial),
          isTrue,
        );
        expect(tester.takeException(), isNull);
      }
    },
  );

  testWidgets(
    'desktop platform paints complete responsive game action targets',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.linux;
      final semantics = tester.ensureSemantics();
      try {
        await start(tester, 390, height: 948);
        final primary = find.widgetWithText(FilledButton, 'End turn');
        final ink = find.descendant(
          of: primary,
          matching: find.byType(InkWell),
        );
        final phoneHeight = tester.getSize(ink).height;
        for (final width in [1049.0, 1050.0, 1440.0, 599.0, 600.0, 390.0]) {
          tester.view.physicalSize = Size(width, 948);
          await tester.pumpAndSettle();
          for (final button in [
            primary,
            find.widgetWithText(OutlinedButton, 'Trade cards'),
          ]) {
            final painted = tester.getSize(
              find.descendant(of: button, matching: find.byType(InkWell)),
            );
            expect(
              painted.height,
              greaterThanOrEqualTo(48),
              reason: 'Linux painted target at $width',
            );
            expect(
              tester.getSemantics(button).rect.height,
              greaterThanOrEqualTo(48),
            );
            if (width >= 1050) {
              expect(painted.height, greaterThanOrEqualTo(56));
              expect(painted.height, greaterThan(phoneHeight));
            }
          }
          expect(tester.takeException(), isNull);
        }
      } finally {
        semantics.dispose();
        debugDefaultTargetPlatformOverride = null;
      }
    },
  );

  testWidgets(
    'game bars distribute controls and grow targets with available width',
    (tester) async {
      await start(tester, 390, height: 948);
      final end = find.widgetWithText(FilledButton, 'End turn');
      final phoneTarget = tester.getRect(end);
      final initial = tester.state(find.byType(CgmsAdaptiveTable));
      tester.view.physicalSize = const Size(1440, 948);
      await tester.pumpAndSettle();
      final hand = tester.getRect(
        find.byKey(const ValueKey('adaptive-hand-viewport')),
      );
      final menu = tester.getRect(find.byTooltip('Game menu'));
      final primary = tester.getRect(end);
      final history = tester.getRect(find.byTooltip('History'));
      expect(menu.right, greaterThan(hand.right - 32));
      expect(primary.right, greaterThan(hand.right - 32));
      expect(primary.height, greaterThan(phoneTarget.height));
      expect(history.left, lessThan(hand.left + 32));
      expect(primary.left - history.right, greaterThan(600));
      expect(
        identical(tester.state(find.byType(CgmsAdaptiveTable)), initial),
        isTrue,
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'reference board keeps four equal quadrants above a visible uniform hand',
    (tester) async {
      await start(tester, 390, height: 825);
      double? previousCardWidth;
      for (final width in [390.0, 768.0, 1440.0]) {
        tester.view.physicalSize = Size(width, 1000);
        await tester.pumpAndSettle();
        expect(find.byType(NavigationRail), findsNothing);
        expect(find.byKey(const ValueKey('nav-hand')), findsNothing);
        final public = find.byKey(const ValueKey('adaptive-public-square'));
        final hand = find.byKey(const PageStorageKey('adaptive-hand'));
        expect(public, findsOneWidget);
        expect(hand, findsOneWidget);
        final publicRect = tester.getRect(public);
        final handRect = tester.getRect(hand);
        expect(publicRect.left, closeTo(handRect.left, .01));
        expect(publicRect.right, closeTo(handRect.right, .01));
        expect(publicRect.width, closeTo(width - 16, .01));
        expect(handRect.top - publicRect.bottom, inInclusiveRange(0, 24));
        expect(handRect.bottom, lessThanOrEqualTo(1000));
        final board = tester.widget<CgmsAdaptiveTable>(
          find.byType(CgmsAdaptiveTable),
        );
        final seats = board.data.seats
            .map(
              (seat) => tester.getRect(
                find.byKey(ValueKey('adaptive-seat-${seat.id}')),
              ),
            )
            .toList();
        expect(seats, hasLength(4));
        // Compare layout sizes directly: subtracting fractional global origins
        // can introduce floating-point differences in otherwise equal cells.
        expect(
          board.data.seats
              .map(
                (seat) => tester.getSize(
                  find.byKey(ValueKey('adaptive-seat-${seat.id}')),
                ),
              )
              .toSet(),
          hasLength(1),
        );
        expect(seats.map((seat) => seat.top).toSet(), hasLength(2));
        expect(seats.map((seat) => seat.left).toSet(), hasLength(2));
        for (var index = 0; index < seats.length; index++) {
          final seat = seats[index];
          expect(seat.width, closeTo(publicRect.width / 2, .01));
          expect(seat.height, closeTo(publicRect.height / 2, .01));
          expect(
            seat.left,
            closeTo(publicRect.left + index % 2 * publicRect.width / 2, .01),
          );
          expect(
            seat.top,
            closeTo(publicRect.top + index ~/ 2 * publicRect.height / 2, .01),
          );
        }
        expect(seats.first.width, greaterThanOrEqualTo(48));
        expect(seats.first.height, greaterThanOrEqualTo(48));
        final cards = find.descendant(
          of: hand,
          matching: find.byType(CgmsCard),
        );
        expect(cards, findsNWidgets(18));
        for (var i = 0; i < 18; i++) {
          expect(tester.getSize(cards.at(i)), tester.getSize(cards.first));
        }
        expect(tester.getSize(cards.first).aspectRatio, closeTo(.8, .001));
        expect(
          tester.getRect(cards.first).intersect(handRect).height,
          greaterThanOrEqualTo(48),
        );
        for (final suit in CardSuit.values) {
          final group = find.byKey(ValueKey('hand-suit-${suit.name}'));
          final members = find.descendant(
            of: group,
            matching: find.byType(CgmsCard),
          );
          expect(members.evaluate(), isNotEmpty);
          for (final element in members.evaluate()) {
            expect((element.widget as CgmsCard).card.suit, suit);
          }
          if (members.evaluate().length > 1) {
            final first = tester.getRect(members.at(0));
            final next = tester.getRect(members.at(1));
            expect(next.top, closeTo(first.top, .01));
            expect(next.left - first.left, greaterThanOrEqualTo(48));
            expect(next.left, lessThan(first.right));
          }
        }
        final cardWidth = tester.getSize(cards.first).width;
        expect(cardWidth, greaterThanOrEqualTo(48));
        // Available height caps growth before a wider desktop can push the hand
        // out of view; the tablet still grows from the phone at this height.
        if (width == 768) {
          expect(cardWidth, greaterThan(previousCardWidth!));
        } else if (previousCardWidth != null) {
          expect(cardWidth, greaterThanOrEqualTo(previousCardWidth));
        }
        previousCardWidth = cardWidth;
      }
      tester.view.physicalSize = const Size(1440, 1200);
      await tester.pumpAndSettle();
      final largerCard = find
          .descendant(
            of: find.byKey(const PageStorageKey('adaptive-hand')),
            matching: find.byType(CgmsCard),
          )
          .first;
      expect(tester.getSize(largerCard).width, greaterThan(previousCardWidth!));
    },
  );

  testWidgets(
    'narrow entry and live boundaries keep the hand adjacent to the table above bounded actions',
    (tester) async {
      await start(tester, 319, height: 948);
      final initial = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      for (final size in const [
        Size(319, 948),
        Size(320, 568),
        Size(599, 900),
        Size(600, 900),
        Size(1049, 900),
        Size(1050, 900),
        Size(1049, 900),
        Size(600, 900),
        Size(599, 900),
        Size(844, 390),
      ]) {
        tester.view.physicalSize = size;
        await tester.pumpAndSettle();
        final public = tester.getRect(
          find.byKey(const ValueKey('adaptive-public-square')),
        );
        final hand = find.byKey(const ValueKey('adaptive-hand-viewport'));
        final handRect = tester.getRect(hand);
        final firstCard = find
            .descendant(of: hand, matching: find.byType(CgmsCard))
            .first;
        expect(public.left, closeTo(handRect.left, .01), reason: '$size');
        expect(public.right, closeTo(handRect.right, .01), reason: '$size');
        expect(public.width, closeTo(size.width - 16, .01), reason: '$size');
        expect(
          handRect.top - public.bottom,
          inInclusiveRange(0, 24),
          reason: '$size',
        );
        expect(
          handRect.bottom,
          lessThanOrEqualTo(size.height),
          reason: '$size',
        );
        expect(
          tester.getRect(firstCard).intersect(handRect).height,
          greaterThanOrEqualTo(48),
          reason: '$size',
        );

        await tester.tap(find.byTooltip('Choose an action'));
        await tester.pumpAndSettle();
        final actions = tester.getRect(
          find.byKey(const PageStorageKey('adaptive-decisions')),
        );
        final expandedHand = tester.getRect(hand);
        final expandedPublic = tester.getRect(
          find.byKey(const ValueKey('adaptive-public-square')),
        );
        expect(expandedPublic.top, public.top, reason: '$size');
        expect(expandedHand.top, closeTo(expandedPublic.bottom + 4, .01));
        expect(actions.top, closeTo(expandedHand.bottom + 4, .01));
        expect(actions.bottom, closeTo(handRect.bottom, .01));
        expect(expandedHand.left, handRect.left);
        expect(expandedHand.right, handRect.right);
        expect(expandedHand.bottom, lessThan(handRect.bottom));
        expect(firstCard.hitTestable(), findsOneWidget, reason: '$size');
        final close = find.byTooltip('Close actions');
        expect(tester.getSize(close).width, greaterThanOrEqualTo(48));
        expect(tester.getSize(close).height, greaterThanOrEqualTo(48));
        await tester.tap(close);
        await tester.pumpAndSettle();
        final current = tester.widget<CgmsAdaptiveTable>(
          find.byType(CgmsAdaptiveTable),
        );
        expect(current.data.gameId, initial.data.gameId);
        expect(
          current.data.hand.map((card) => card.id),
          initial.data.hand.map((card) => card.id),
        );
        expect(current.data.pending, isNull);
        expect(current.selectedCardIds, isEmpty);
        expect(tester.getRect(hand), handRect, reason: '$size');
        expect(tester.takeException(), isNull, reason: '$size');
      }
    },
  );

  testWidgets(
    'every hand card keeps focused semantics after horizontal scrolling and resizing',
    (tester) async {
      final semantics = tester.ensureSemantics();
      await start(tester, 390, height: 900);
      for (final width in [390.0, 768.0, 1440.0, 390.0]) {
        tester.view.physicalSize = Size(width, 900);
        await tester.pumpAndSettle();
        final hand = find.byKey(const PageStorageKey('adaptive-hand'));
        await tester.ensureVisible(hand);
        await tester.pumpAndSettle();
        final cards = find.descendant(
          of: hand,
          matching: find.byType(CgmsCard),
        );
        expect(cards, findsNWidgets(18));
        for (var index = 0; index < 18; index++) {
          final target = cards.at(index);
          final model = tester.widget<CgmsCard>(target).card;
          final focus = tester
              .widget<InkWell>(
                find.descendant(of: target, matching: find.byType(InkWell)),
              )
              .focusNode!;
          for (var tabs = 0; tabs < 100 && !focus.hasPrimaryFocus; tabs++) {
            await tester.sendKeyEvent(LogicalKeyboardKey.tab);
            await tester.pumpAndSettle();
          }
          expect(focus.hasPrimaryFocus, isTrue, reason: '$width ${model.id}');
          final data = tester
              .getSemantics(find.byKey(ValueKey('select-${model.id}')))
              .getSemanticsData();
          expect(
            data.flagsCollection.isFocused,
            Tristate.isTrue,
            reason: '$width ${model.id}',
          );
          final siblings = cards
              .evaluate()
              .map((e) => (e.widget as CgmsCard).card)
              .where((c) => c.suit == model.suit)
              .map(
                (c) =>
                    tester.getSemantics(find.byKey(ValueKey('select-${c.id}'))),
              )
              .toList();
          final parent = tester
              .getSemantics(find.byKey(ValueKey('select-${model.id}')))
              .parent!;
          final order = parent
              .debugListChildrenInOrder(DebugSemanticsDumpOrder.traversalOrder)
              .where((node) => siblings.any((sibling) => sibling.id == node.id))
              .map((node) => node.id)
              .toList();
          expect(
            order,
            siblings.map((node) => node.id).toList(),
            reason: 'stable semantic order $width ${model.id}',
          );
          final rect = tester.getRect(target);
          final group = tester.getRect(
            find.byKey(ValueKey('hand-suit-${model.suit.name}')),
          );
          expect(
            rect.left,
            greaterThanOrEqualTo(group.left - .01),
            reason: '$width ${model.id}',
          );
          expect(
            rect.right,
            lessThanOrEqualTo(group.right + .01),
            reason: '$width ${model.id}',
          );
          expect(rect.top, greaterThanOrEqualTo(-.01));
          expect(rect.bottom, lessThanOrEqualTo(900.01));
        }
        expect(tester.takeException(), isNull);
      }
      semantics.dispose();
    },
  );

  testWidgets('keyboard reaches the last hand card in a short enlarged view', (
    tester,
  ) async {
    await start(tester, 844, scale: 2, height: 390);
    await tester.ensureVisible(
      find.byKey(const PageStorageKey('adaptive-hand')),
    );
    await tester.pumpAndSettle();
    final hand = find.byKey(const PageStorageKey('adaptive-hand'));
    final cards = find.descendant(of: hand, matching: find.byType(CgmsCard));
    expect(cards.evaluate().length, greaterThan(4));
    final lastCard = cards.last;
    final focus = tester
        .widget<InkWell>(
          find.descendant(of: lastCard, matching: find.byType(InkWell)).first,
        )
        .focusNode!;
    final handViewport = tester.getRect(
      find.byKey(const ValueKey('adaptive-hand-viewport')),
    );
    expect(handViewport.height, greaterThanOrEqualTo(48));
    expect(tester.getRect(lastCard).bottom, greaterThan(handViewport.bottom));
    for (var i = 0; i < 100 && !focus.hasPrimaryFocus; i++) {
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.pumpAndSettle();
    }
    expect(focus.hasPrimaryFocus, isTrue);
    expect(
      tester.getRect(lastCard).intersect(handViewport).height,
      greaterThanOrEqualTo(48),
    );
    await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
    await tester.pumpAndSettle();
    expect(find.byType(Dialog), findsOneWidget);
    await tester.sendKeyEvent(LogicalKeyboardKey.escape);
    await tester.pumpAndSettle();
    expect(find.byType(Dialog), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets('normal entry Auto follows logical widths and live boundaries', (
    tester,
  ) async {
    await start(tester, 390);
    expect(find.text('Auto'), findsOneWidget);
    for (final (width, kind) in [
      (390.0, 'phone'),
      (599.0, 'phone'),
      (600.0, 'tablet'),
      (1049.0, 'tablet'),
      (1050.0, 'desktop'),
      (1440.0, 'desktop'),
      (1049.0, 'tablet'),
      (600.0, 'tablet'),
      (599.0, 'phone'),
    ]) {
      tester.view.physicalSize = Size(width, 900);
      await tester.pumpAndSettle();
      expect(find.byKey(ValueKey('$kind-table')), findsOneWidget);
      final picker = find.byTooltip('Table layout');
      expect(
        find.descendant(
          of: picker,
          matching: find.byIcon(switch (kind) {
            'phone' => Icons.phone_android,
            'tablet' => Icons.tablet_mac,
            _ => Icons.desktop_windows,
          }),
        ),
        width >= 1050 ? findsOneWidget : findsNothing,
      );
      expect(
        find.descendant(of: picker, matching: find.text('Auto')),
        findsOneWidget,
      );
      expect(
        find.descendant(
          of: picker,
          matching: find.byIcon(Icons.arrow_drop_down),
        ),
        findsOneWidget,
      );
      expect(tester.takeException(), isNull);
    }
  });

  testWidgets('200 percent toolbar fits the complete Auto label and controls', (
    tester,
  ) async {
    await start(tester, 390, scale: 2);
    final toolbar = find.byTooltip('Table layout');
    final label = find.text('Auto');
    final labelBox = tester.renderObject<RenderBox>(label);
    final paragraph = tester.renderObject<RenderBox>(
      find.descendant(of: label, matching: find.byType(RichText)),
    );
    expect(
      paragraph.getMaxIntrinsicHeight(paragraph.size.width),
      lessThanOrEqualTo(labelBox.size.height),
    );
    final barRect = tester.getRect(toolbar);
    final labelRect = tester.getRect(label);
    expect(labelRect.top - barRect.top, greaterThanOrEqualTo(12));
    expect(barRect.bottom - labelRect.bottom, greaterThanOrEqualTo(12));
    expect(
      tester.getSize(find.byTooltip('Table layout')).height,
      greaterThanOrEqualTo(48),
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'large text header labels are fully painted above the public square',
    (tester) async {
      await start(tester, 390, scale: 2);
      for (final size in const [
        Size(319, 568),
        Size(390, 844),
        Size(1440, 900),
        Size(844, 390),
      ]) {
        tester.view.physicalSize = size;
        await tester.pumpAndSettle();
        final tableTop = tester
            .getRect(find.byKey(const ValueKey('adaptive-public-square')))
            .top;
        Rect headerRegion(String name) => tester.getRect(
          find.byWidgetPredicate(
            (widget) =>
                widget is Semantics && widget.properties.identifier == name,
          ),
        );
        expect(
          headerRegion('header-status').center.dy,
          closeTo(headerRegion('header-tools').center.dy, .01),
        );
        for (final label in ['Round 7 / 13', 'Your turn', 'Auto']) {
          final target = find.text(label);
          final bounds = tester.getRect(target);
          expect(
            bounds.top,
            greaterThanOrEqualTo(0),
            reason: '$label at $size',
          );
          expect(
            bounds.bottom,
            lessThanOrEqualTo(tableTop),
            reason: '$label at $size',
          );
          final paragraph = tester.renderObject<RenderBox>(target);
          expect(
            paragraph.getMaxIntrinsicHeight(paragraph.size.width),
            lessThanOrEqualTo(paragraph.size.height),
          );
        }
        final hand = tester.getRect(
          find.byKey(const ValueKey('adaptive-hand-viewport')),
        );
        final first = find
            .descendant(
              of: find.byKey(const PageStorageKey('adaptive-hand')),
              matching: find.byType(CgmsCard),
            )
            .first;
        expect(
          tester.getRect(first).intersect(hand).height,
          greaterThanOrEqualTo(48),
        );
        expect(tester.takeException(), isNull);
      }
    },
  );

  testWidgets(
    'Auto uses logical pixels and preserves response state on rotation',
    (tester) async {
      await start(tester, 390);
      tester.view.devicePixelRatio = 2;
      tester.view.physicalSize = const Size(780, 1688);
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('phone-table')), findsOneWidget);
      await openActions(tester);
      await tester.pumpAndSettle();
      final primary = find.byKey(const ValueKey('approved-primary'));
      await tester.ensureVisible(primary);
      await tester.tap(primary);
      await tester.pumpAndSettle();
      final initial = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      expect(initial.data.pending, isNotNull);
      expect(initial.data.pending!.responseIndex, 0);
      expect(initial.selectedCardIds, hasLength(7));
      final game = initial.data.gameId;
      final hand = initial.data.hand.map((card) => card.id).toList();
      for (final (width, height, kind) in [
        (844.0, 390.0, 'tablet'),
        (768.0, 1024.0, 'tablet'),
        (1024.0, 768.0, 'tablet'),
        (1280.0, 600.0, 'desktop'),
        (390.0, 844.0, 'phone'),
      ]) {
        tester.view.physicalSize = Size(width * 2, height * 2);
        await tester.pumpAndSettle();
        expect(find.byKey(ValueKey('$kind-table')), findsOneWidget);
        final board = tester.widget<CgmsAdaptiveTable>(
          find.byType(CgmsAdaptiveTable),
        );
        expect(board.data.gameId, game);
        expect(board.data.hand.map((card) => card.id), hand);
        expect(board.selectedCardIds, initial.selectedCardIds);
        expect(
          board.data.pending!.responseIndex,
          initial.data.pending!.responseIndex,
        );
        expect(tester.takeException(), isNull);
      }
      await tester.ensureVisible(primary);
      await tester.tap(primary);
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .data
            .pending!
            .responseIndex,
        1,
      );
    },
  );

  testWidgets(
    'phone overview uses paired seats and complete inspectable paintings',
    (tester) async {
      await start(tester, 390);
      final board = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      final seats = board.data.seats;
      final first = tester.getRect(
        find.byKey(ValueKey('adaptive-seat-${seats[0].id}')),
      );
      final second = tester.getRect(
        find.byKey(ValueKey('adaptive-seat-${seats[1].id}')),
      );
      expect(first.top, second.top);
      expect(first.right, closeTo(second.left, .01));
      final cards = seats.first.exposed
          .where((card) => card.suit == CardSuit.diamonds)
          .toList();
      expect(cards.length, greaterThan(1));
      final earlier = tester.getRect(
        find.byKey(ValueKey('card-art-${cards[0].id}')),
      );
      final next = tester.getRect(
        find.byKey(ValueKey('card-header-${cards[1].id}')),
      );
      expect(
        earlier.width / earlier.height,
        closeTo(cards[0].artworkAspectRatio, .001),
      );
      expect(earlier.bottom, greaterThan(next.top));
      final quadrant = find.byKey(ValueKey('adaptive-seat-${seats.first.id}'));
      expect(tester.getSize(quadrant).width, greaterThanOrEqualTo(48));
      expect(tester.getSize(quadrant).height, greaterThanOrEqualTo(48));
      await tester.tap(quadrant);
      await tester.pumpAndSettle();
      final chooserCards = tester.widgetList<CgmsCard>(
        find.descendant(
          of: find.byType(AlertDialog),
          matching: find.byType(CgmsCard),
        ),
      );
      expect(
        chooserCards.map((card) => card.card.id).toSet(),
        seats.first.exposed.map((card) => card.id).toSet(),
      );
      final inspect = find.byKey(ValueKey('magnify-${cards[0].id}'));
      final target = tester.getRect(inspect);
      final fullArt = tester.getRect(
        find.byKey(ValueKey('card-art-${cards[0].id}')).last,
      );
      expect(tester.getSize(inspect).height, greaterThanOrEqualTo(48));
      expect(tester.getSize(inspect).width, greaterThanOrEqualTo(48));
      expect(target.bottom, closeTo(fullArt.bottom, .001));
      await tester.ensureVisible(inspect);
      await tester.tap(inspect);
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCardInspector), findsOneWidget);
      expect(find.text('Inspect card'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'approved route remains usable with RTL and 200 percent text in every class',
    (tester) async {
      final controller = ApprovedDemoController();
      addTearDown(controller.dispose);
      addTearDown(() => tester.binding.setSurfaceSize(null));
      for (final size in [
        const Size(390, 844),
        const Size(768, 1024),
        const Size(1440, 600),
      ]) {
        await tester.binding.setSurfaceSize(size);
        await tester.pumpWidget(
          MaterialApp(
            theme: CgmsTheme.dark(),
            localizationsDelegates: CgmsLocalizations.localizationsDelegates,
            supportedLocales: CgmsLocalizations.supportedLocales,
            home: Directionality(
              textDirection: TextDirection.rtl,
              child: MediaQuery(
                data: MediaQueryData(
                  size: size,
                  textScaler: const TextScaler.linear(2),
                ),
                child: AnimatedBuilder(
                  animation: controller,
                  builder: (context, _) => ApprovedFlowView(
                    controller: controller,
                    route: '/play',
                    tableRoute: '/play',
                    layoutMenu: const Text('Auto'),
                    onNavigate: (_) {},
                    onSettings: () {},
                    showArt: false,
                  ),
                ),
              ),
            ),
          ),
        );
        await tester.pumpAndSettle();
        expect(
          Directionality.of(tester.element(find.byType(CgmsAdaptiveTable))),
          TextDirection.rtl,
        );
        Rect headerRegion(String name) => tester.getRect(
          find.byWidgetPredicate(
            (widget) =>
                widget is Semantics && widget.properties.identifier == name,
          ),
        );
        final status = headerRegion('header-status');
        final tools = headerRegion('header-tools');
        expect(status.center.dy, closeTo(tools.center.dy, .01));
        expect(tools.right, lessThanOrEqualTo(status.left));
        await tester.ensureVisible(
          find.byKey(const PageStorageKey('adaptive-hand')),
        );
        await tester.pumpAndSettle();
        final card = find.byKey(
          ValueKey('adaptive-select-${controller.table.hand.first.id}'),
        );
        await tester.ensureVisible(card);
        expect(tester.getSize(card).width, greaterThanOrEqualTo(48));
        tester
            .widget<InkWell>(
              find.descendant(of: card, matching: find.byType(InkWell)),
            )
            .focusNode!
            .requestFocus();
        await tester.pump();
        await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
        await tester.pumpAndSettle();
        expect(find.byType(CgmsCardInspector), findsOneWidget);
        await tester.tap(find.text('Close inspection'));
        await tester.pumpAndSettle();
        expect(find.byType(CgmsCardInspector), findsNothing);
        await openActions(tester);
        final primary = find.byKey(const ValueKey('approved-primary'));
        await tester.ensureVisible(primary);
        expect(primary.hitTestable(), findsOneWidget);
        expect(tester.getSize(primary).height, greaterThanOrEqualTo(48));
        expect(tester.takeException(), isNull, reason: 'RTL at200%: $size');
      }
    },
  );

  testWidgets('adaptive route preserves public totals and combo inspection', (
    tester,
  ) async {
    await start(tester, 390);
    expect(find.text('Diamonds 18'), findsOneWidget);
    final board = tester.widget<CgmsAdaptiveTable>(
      find.byType(CgmsAdaptiveTable),
    );
    final owner = board.data.seats.firstWhere(
      (seat) => seat.combos.any((combo) => combo.title == 'Underground'),
    );
    final quadrant = find.byKey(ValueKey('adaptive-seat-${owner.id}'));
    await tester.tap(quadrant);
    await tester.pumpAndSettle();
    final combo = find.text('Inspect Underground');
    await tester.ensureVisible(combo);
    await tester.tap(combo);
    await tester.pumpAndSettle();
    expect(find.byType(CgmsComboInspector), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('direct scores route reveals scores from phone overview', (
    tester,
  ) async {
    await start(tester, 390);
    await tester.ensureVisible(
      find.byKey(const PageStorageKey('adaptive-hand')),
    );
    await tester.pumpAndSettle();
    final card = find.byKey(const ValueKey('adaptive-select-deck-1-clubs-Q'));
    tester
        .widget<InkWell>(
          find.descendant(of: card, matching: find.byType(InkWell)).first,
        )
        .focusNode!
        .requestFocus();
    await tester.pump();
    await tester.binding.handlePushRoute('/scores');
    await tester.pumpAndSettle();
    expect(find.text('Scores and obligations'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('keyboard inspector contains focus and Escape closes it', (
    tester,
  ) async {
    await start(tester, 390);
    await tester.ensureVisible(
      find.byKey(const PageStorageKey('adaptive-hand')),
    );
    await tester.pumpAndSettle();
    final card = find.byKey(const ValueKey('adaptive-select-deck-1-clubs-Q'));
    final focus = tester
        .widget<InkWell>(
          find.descendant(of: card, matching: find.byType(InkWell)).first,
        )
        .focusNode!;
    focus.requestFocus();
    await tester.pumpAndSettle();
    await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
    await tester.pumpAndSettle();
    expect(find.byType(CgmsCardInspector), findsOneWidget);
    expect(
      FocusManager.instance.primaryFocus!.context!
          .findAncestorWidgetOfExactType<CgmsCardInspector>(),
      isNotNull,
    );
    await tester.sendKeyEvent(LogicalKeyboardKey.escape);
    await tester.pumpAndSettle();
    expect(find.byType(CgmsCardInspector), findsNothing);
    expect(find.byKey(const PageStorageKey('adaptive-hand')), findsOneWidget);
  });

  for (final size in const [Size(390, 844), Size(768, 1024), Size(1440, 900)]) {
    testWidgets(
      'complete initial and opened hand are fully reachable at $size',
      (tester) async {
        await start(tester, size.width, height: size.height);
        Future<void> expectVisibleHand(int count) async {
          await tester.ensureVisible(
            find.byKey(const PageStorageKey('adaptive-hand')),
          );
          await tester.pumpAndSettle();
          final board = tester.widget<CgmsAdaptiveTable>(
            find.byType(CgmsAdaptiveTable),
          );
          expect(board.data.hand, hasLength(count));
          for (final card in board.data.hand) {
            final target = find.byKey(ValueKey('adaptive-select-${card.id}'));
            await tester.ensureVisible(target);
            await tester.pumpAndSettle();
            final rect = tester.getRect(target);
            expect(
              rect.top,
              greaterThanOrEqualTo(0),
              reason: '${card.id}: $count cards',
            );
            expect(
              rect.bottom,
              lessThanOrEqualTo(size.height),
              reason: '${card.id}: $count cards',
            );
          }
        }

        await expectVisibleHand(18);
        await openExamples(tester);
        await tester.pumpAndSettle();
        final example = find.byKey(const ValueKey('approved-example-6'));
        await tester.ensureVisible(example);
        await tester.tap(example);
        await tester.pumpAndSettle();
        await expectVisibleHand(11);
        expect(tester.takeException(), isNull);
      },
    );
  }

  for (final width in [390.0, 768.0, 1440.0]) {
    testWidgets(
      'approved default shares all players and complete hand at $width',
      (tester) async {
        await start(tester, width);
        final board = tester.widget<CgmsAdaptiveTable>(
          find.byType(CgmsAdaptiveTable),
        );
        expect(board.data.hand, hasLength(18));
        expect(board.data.seats, hasLength(4));
        expect(board.data.seats.map((s) => s.name), [
          'Alice',
          'Bob',
          'Carol',
          'Deniz',
        ]);
        expect(find.byKey(const ValueKey('touch-tab-hand')), findsNothing);
        expect(find.byTooltip('Choose an action'), findsOneWidget);
        expect(find.byKey(const ValueKey('approved-primary')), findsNothing);
        expect(board.selectedCardIds, isEmpty);
        expect(board.data.pending, isNull);
        expect(find.byTooltip('Board actions'), findsNothing);
        expect(tester.takeException(), isNull);
        if (width == 390) {
          await tester.ensureVisible(
            find.byKey(const PageStorageKey('adaptive-hand')),
          );
          await tester.pumpAndSettle();
          for (final card in board.data.hand) {
            final target = find.byKey(ValueKey('adaptive-select-${card.id}'));
            await tester.ensureVisible(target);
            await tester.pumpAndSettle();
            final rect = tester.getRect(target);
            expect(rect.top, greaterThanOrEqualTo(0));
            expect(rect.bottom, lessThanOrEqualTo(844), reason: card.id);
          }
        }
      },
    );
  }
  testWidgets('opening and explicit response actions update the same board', (
    tester,
  ) async {
    await start(tester, 390);
    Future<void> primary() async {
      await openActions(tester);
      await tester.pumpAndSettle();
      final button = find.byKey(const ValueKey('approved-primary'));
      await tester.ensureVisible(button);
      await tester.tap(button);
      await tester.pumpAndSettle();
    }

    await openActions(tester); // choose Open Hearts and select its seven cards
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .selectedCardIds,
      hasLength(7),
    );
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data
          .pending,
      isNull,
    );
    await primary(); // opening awaits responses
    const pendingStatus =
        'Waiting for responses · Waiting for Bob. Each pass below is an explicit local simulation. Cards move only when the action resolves.';
    expect(find.text(pendingStatus), findsOneWidget);
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data
          .pending!
          .responseIndex,
      0,
    );
    expect(
      find.descendant(
        of: find.byWidgetPredicate(
          (widget) =>
              widget is Semantics && widget.properties.liveRegion == true,
        ),
        matching: find.text(pendingStatus),
      ),
      findsOneWidget,
    );
    expect(find.text('Simulate Bob passing'), findsOneWidget);
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data
          .hand,
      hasLength(18),
    );
    for (var i = 0; i < 3; i++) {
      await primary();
    }
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data
          .hand,
      hasLength(11),
    );
    final details = find.byTooltip('Hearts opened details');
    await tester.ensureVisible(details);
    await tester.tap(details);
    await tester.pumpAndSettle();
    expect(
      find.text(
        'Seven physical cards moved to your Hearts series. Your opening allowance is used; your 11 other hand cards remain private.',
      ),
      findsOneWidget,
    );
    expect(tester.takeException(), isNull);
  });
  testWidgets(
    'twenty examples are accessible and navigation preserves their game',
    (tester) async {
      await start(tester, 768);
      await openExamples(tester);
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('approved-example-20')), findsOneWidget);
      final coup = find.byKey(const ValueKey('approved-example-17'));
      await tester.ensureVisible(coup);
      await tester.tap(coup);
      await tester.pumpAndSettle();
      final before = tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data;
      expect(before.hand.where((c) => c.rank == 'Q'), hasLength(4));
      await tester.binding.handlePushRoute('/tablet');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .data
            .gameId,
        before.gameId,
      );
      await tester.binding.handlePushRoute('/phone');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .data
            .hand
            .map((c) => c.id),
        before.hand.map((c) => c.id),
      );
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets('large text and artwork fallback retain actionable controls', (
    tester,
  ) async {
    await start(tester, 390, scale: 2);
    await tester.ensureVisible(find.byTooltip('Game menu'));
    await tester.tap(find.byTooltip('Game menu'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Display settings'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(SwitchListTile, 'Card artwork'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Done'));
    await tester.pumpAndSettle();
    expect(
      tester.widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable)).showArt,
      isFalse,
    );
    await openActions(tester);
    await tester.pumpAndSettle();
    final primary = find.byKey(const ValueKey('approved-primary'));
    await tester.ensureVisible(primary);
    await tester.tap(primary);
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .selectedCardIds,
      hasLength(7),
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'approved chooser exposes exactly twenty flows and no legacy navigation',
    (tester) async {
      await start(tester, 768);
      await openExamples(tester);
      await tester.pumpAndSettle();
      expect(find.byType(OutlinedButton), findsNWidgets(20));
      for (var example = 1; example <= 20; example++) {
        expect(
          find.byKey(ValueKey('approved-example-$example')),
          findsOneWidget,
        );
      }
      expect(find.text('Earlier demo scenarios'), findsNothing);
      await tester.ensureVisible(find.byTooltip('Game menu'));
      await tester.tap(find.byTooltip('Game menu'));
      await tester.pumpAndSettle();
      expect(find.text('Room setup'), findsOneWidget);
      expect(find.text('Game rules'), findsOneWidget);
      expect(find.text('Display settings'), findsOneWidget);
      expect(find.text('Earlier demo scenarios'), findsNothing);
    },
  );

  testWidgets('Coup stays reachable after canceling an out-of-turn Ponzi', (
    tester,
  ) async {
    await start(tester, 390);
    await openExamples(tester);
    await tester.pumpAndSettle();
    final choice = find.byKey(const ValueKey('approved-example-17'));
    await tester.ensureVisible(choice);
    await tester.tap(choice);
    await tester.pumpAndSettle();
    await openActions(tester);
    await tester.pumpAndSettle();
    final branch = find.text('Try Coup during a response');
    await tester.ensureVisible(branch);
    await tester.tap(branch);
    await tester.pumpAndSettle();
    final confine = find.byKey(const ValueKey('approved-confinement'));
    await tester.ensureVisible(confine);
    await tester.tap(confine);
    await tester.pumpAndSettle();
    final coup = find.byKey(const ValueKey('immediate-coup'));
    await tester.ensureVisible(coup);
    await tester.tap(coup);
    await tester.pumpAndSettle();
    expect(find.text('Alice wins by Coup'), findsOneWidget);
    final closed = tester.widget<CgmsAdaptiveTable>(
      find.byType(CgmsAdaptiveTable),
    );
    expect(closed.data.phaseLabel, 'Alice wins by Coup');
    expect(closed.data.turnLabel, contains('Board closed'));
    expect(closed.onCoup, isNull);
    expect(closed.selectedCardIds, isEmpty);
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data
          .pending,
      isNull,
    );
  });

  testWidgets(
    'non-actionable cards magnify only through independent inspection',
    (tester) async {
      await start(tester, 390);
      var board = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      board.onSelect('deck-1-clubs-6');
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCardInspector), findsNothing);
      board.onInspect(
        board.data.seats
            .expand((seat) => seat.exposed)
            .firstWhere((card) => card.id == 'deck-1-clubs-6'),
      );
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCardInspector), findsOneWidget);
      Navigator.of(tester.element(find.byType(CgmsCardInspector))).pop();
      await tester.pumpAndSettle();
      board = tester.widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable));
      board.onSelect('deck-1-clubs-Q');
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCardInspector), findsNothing);
      board.onInspect(
        board.data.hand.firstWhere((card) => card.id == 'deck-1-clubs-Q'),
      );
      await tester.pumpAndSettle();
      expect(find.byType(CgmsCardInspector), findsOneWidget);
      expect(board.data.hand, hasLength(18));
    },
  );

  testWidgets('direct room navigation accepts name and match length', (
    tester,
  ) async {
    await start(tester, 390);
    await tester.binding.handlePushRoute('/setup');
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextFormField), 'Guest');
    final five = find.widgetWithText(ChoiceChip, '5 games');
    await tester.ensureVisible(five);
    await tester.tap(five);
    await tester.pumpAndSettle();
    expect(tester.widget<ChoiceChip>(five).selected, isTrue);
    final startButton = find.byKey(const ValueKey('approved-start-match'));
    await tester.ensureVisible(startButton);
    await tester.tap(startButton);
    await tester.pumpAndSettle();
    final data = tester
        .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
        .data;
    expect(data.seats.first.name, 'Guest');
    expect(data.turnLabel, contains('Game 1 of 5'));
  });
}

Future<void> openActions(WidgetTester tester) async {
  if (find.byTooltip('Close actions').evaluate().isEmpty) {
    await tester.tap(find.byTooltip('Choose an action'));
    await tester.pumpAndSettle();
  }
  final primary = find.byKey(const ValueKey('approved-primary'));
  if (primary.evaluate().isEmpty &&
      find.byTooltip('Board actions').evaluate().isNotEmpty) {
    await tester.ensureVisible(find.byTooltip('Board actions'));
    await tester.tap(find.byTooltip('Board actions'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(PopupMenuItem<String>, 'Open Hearts'));
    await tester.pumpAndSettle();
  }
  await tester.ensureVisible(
    find.byKey(const PageStorageKey('adaptive-decisions')),
  );
  await tester.pumpAndSettle();
}

Future<void> openExamples(WidgetTester tester) async {
  final direct = find.byTooltip('Flow examples');
  if (direct.evaluate().isNotEmpty) {
    await tester.ensureVisible(direct);
    await tester.tap(direct);
  } else {
    await tester.ensureVisible(find.byTooltip('Game menu'));
    await tester.tap(find.byTooltip('Game menu'));
    await tester.pumpAndSettle();
    await tester.tap(
      find.widgetWithText(PopupMenuItem<String>, 'Flow examples'),
    );
  }
}
