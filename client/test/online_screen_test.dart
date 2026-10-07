import 'dart:async';
import 'dart:ui' show SemanticsAction;

import 'package:cgms_demo/online/online_client.dart';
import 'package:cgms_demo/online/command_contract.dart';
import 'package:cgms_demo/online/online_screen.dart';
import 'package:cgms_demo/online/economy_panel.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'online/online_client_test.dart' show FakeTransport;

class ScreenTransport implements OnlineTransport {
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

void main() {
  for (final inflation in [false, true]) {
    testWidgets(
      'purchase blocks known payment shortfall with inflation $inflation',
      (tester) async {
        final client = ViewClient()
          ..publish(
            {
              'cards': [
                for (final rank in [9, 2])
                  {
                    'card': {
                      'id': 'payment-$rank',
                      'rank': rank,
                      'suit': 'spades',
                      'deck': 1,
                    },
                    'controller': 1,
                    'zone': 'hand',
                  },
              ],
            },
            online: {
              'rules_context': {
                'purchase_unit_price': 5,
                'inflation': inflation,
              },
            },
          );
        addTearDown(client.dispose);
        await showOnline(tester, client, expand: false);
        await tester.tap(
          find.byKey(const ValueKey('adaptive-select-payment-9')),
        );
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        await activatePrivateCard(tester, 'payment-2');
        final choice = find.byKey(const ValueKey('card-intent-purchase'));
        await tester.ensureVisible(choice);
        await tester.tap(choice);
        await tester.pumpAndSettle();
        final quantity = find.widgetWithText(TextField, 'Quantity');
        await tester.ensureVisible(quantity);
        await tester.enterText(quantity, inflation ? '2' : '3');
        await tester.pumpAndSettle();
        final confirm = find.byKey(const ValueKey('online-submit'));
        expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
        expect(
          find.textContaining('Effective payment ${inflation ? '11/2' : '11'}'),
          findsOneWidget,
        );
        expect(client.submissions, isEmpty);
        await tester.enterText(quantity, inflation ? '1' : '2');
        await tester.pumpAndSettle();
        expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
        await tester.ensureVisible(confirm);
        await tester.tap(confirm);
        await tester.pumpAndSettle();
        expect(client.submissions.single, {
          'type': 'purchase',
          'payload': {
            'quantity': inflation ? 1 : 2,
            'payment': ['payment-9', 'payment-2'],
          },
        });
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets('modifier-first attack accepts exact Clubs and target on board', (
    tester,
  ) async {
    Map<String, dynamic> card(
      String id,
      int rank,
      String suit,
      int owner, {
      String zone = 'series',
    }) => {
      'card': {'id': id, 'rank': rank, 'suit': suit, 'deck': 1},
      'controller': owner,
      'zone': zone,
      if (zone == 'attachment') 'allocation': 'clubs',
    };
    final client = ViewClient()
      ..publish({
        'turn': 3,
        'players': [
          for (final seat in [1, 2])
            {
              'seat': seat,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['clubs', 'hearts'],
            },
        ],
        'cards': [
          card('exile-source', 12, 'clubs', 1, zone: 'attachment'),
          card('attack-club', 7, 'clubs', 1),
          card('attack-target', 7, 'hearts', 2),
        ],
      });
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    await openPublicSeat(tester, 'exile-source');
    await activatePrivateCard(tester, 'exile-source');
    final choice = find.byKey(const ValueKey('card-intent-attack'));
    if (choice.evaluate().isNotEmpty) {
      await tester.ensureVisible(choice);
      await tester.tap(choice);
      await tester.pumpAndSettle();
    }
    final chooseClubs = find.text('Select attacking Clubs');
    expect(chooseClubs, findsOneWidget);
    await tester.ensureVisible(chooseClubs);
    await tester.tap(chooseClubs);
    await tester.pumpAndSettle();
    await openPublicSeat(tester, 'attack-club');
    await tester.tap(find.byKey(const ValueKey('adaptive-select-attack-club')));
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    await openPublicSeat(tester, 'attack-club');
    await activatePrivateCard(tester, 'attack-club');
    final chooseTargets = find.text('Choose the target card');
    await tester.ensureVisible(chooseTargets);
    await tester.tap(chooseTargets);
    await tester.pumpAndSettle();
    await openPublicSeat(tester, 'attack-target');
    await tester.tap(
      find.byKey(const ValueKey('adaptive-select-attack-target')),
    );
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    await openPublicSeat(tester, 'attack-target');
    await activatePrivateCard(tester, 'attack-target');
    expect(client.submissions, isEmpty);
    final confirm = find.byKey(const ValueKey('online-submit'));
    await tester.ensureVisible(confirm);
    await tester.tap(confirm);
    await tester.pumpAndSettle();
    expect(client.submissions.single['type'], 'attack');
    expect(client.submissions.single['payload'], {
      'selection': ['attack-club'],
      'targets': ['attack-target'],
      'suit': 'hearts',
      'exile': true,
    });
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'open online actions keep public cards readable and a full hand card visible',
    (tester) async {
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final client = ViewClient()
        ..publish({
          'players': [
            for (var seat = 1; seat <= 4; seat++)
              {'seat': seat, 'hand_count': 15, 'ace_count': 2},
          ],
          'cards': [
            for (var seat = 1; seat <= 4; seat++)
              for (var rank = 2; rank <= 8; rank++)
                {
                  'card': {
                    'id': 'readable-public-$seat-$rank',
                    'rank': rank,
                    'suit': 'diamonds',
                    'deck': seat,
                  },
                  'controller': seat,
                  'zone': 'series',
                },
            for (var rank = 2; rank <= 10; rank++)
              {
                'card': {
                  'id': 'readable-hand-$rank',
                  'rank': rank,
                  'suit': 'hearts',
                  'deck': 1,
                },
                'controller': 1,
                'zone': 'hand',
              },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'open-series');
      for (final size in [const Size(1025, 948), const Size(1440, 900)]) {
        tester.view.physicalSize = size;
        await tester.binding.setSurfaceSize(size);
        await tester.pumpAndSettle();
        final public = tester.getRect(
          find.byKey(const ValueKey('adaptive-public-square')),
        );
        for (var seat = 1; seat <= 4; seat++) {
          final box = tester.renderObject<RenderBox>(
            find.byKey(ValueKey('public-preview-readable-public-$seat-8')),
          );
          final painted = MatrixUtils.transformRect(
            box.getTransformTo(null),
            Offset.zero & box.size,
          );
          expect(
            painted.width,
            greaterThanOrEqualTo(64),
            reason: '$size seat $seat',
          );
          expect(painted.intersect(public).size, painted.size);
        }
        final hand = tester.getRect(
          find.byKey(const ValueKey('adaptive-hand-viewport')),
        );
        final firstCard = tester.getRect(
          find.byKey(const ValueKey('adaptive-select-readable-hand-2')),
        );
        expect(
          firstCard.intersect(hand).height,
          closeTo(firstCard.height, .01),
        );
        expect(hand.top, closeTo(public.bottom + 4, .01));
        expect(
          find.byKey(const ValueKey('adaptive-draw-pile')),
          findsOneWidget,
        );
        final action = find.textContaining('Series ·');
        await tester.ensureVisible(action);
        await tester.pumpAndSettle();
        expect(action.hitTestable(), findsOneWidget);
        final submit = find.byKey(const ValueKey('online-submit'));
        await tester.ensureVisible(submit);
        await tester.pumpAndSettle();
        final actionViewport = tester.getRect(
          find.ancestor(of: submit, matching: find.byType(Scrollable)).first,
        );
        expect(
          tester.getRect(submit).intersect(actionViewport).height,
          greaterThanOrEqualTo(48),
        );
        expect(tester.takeException(), isNull);
      }
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'face-down draw pile remains visible with actions open or closed',
    (tester) async {
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      final semantics = tester.ensureSemantics();
      try {
        await showOnline(tester, client);
        final pile = find.byKey(const ValueKey('adaptive-draw-pile'));
        for (final size in [const Size(1356, 948), const Size(390, 844)]) {
          tester.view.physicalSize = size;
          await tester.binding.setSurfaceSize(size);
          await tester.pumpAndSettle();
          expect(pile, findsOneWidget);
          final table = tester.getRect(
            find.byKey(const ValueKey('adaptive-public-square')),
          );
          final bounds = tester.getRect(pile);
          expect(bounds.intersect(table), bounds);
          expect(bounds.center.dx, closeTo(table.center.dx, .01));
          expect(bounds.center.dy, closeTo(table.center.dy, .01));
          final node = tester.getSemantics(pile).getSemanticsData();
          expect(node.label, 'Draw pile');
          expect(node.hasAction(SemanticsAction.tap), isTrue);
          expect(
            find.descendant(of: pile, matching: find.byType(ConcealedCard)),
            findsWidgets,
          );
          expect(
            find.descendant(of: pile, matching: find.byType(CgmsCard)),
            findsNothing,
          );
        }
        await tester.tap(find.byKey(const ValueKey('close-card-context')));
        await tester.pumpAndSettle();
        expect(pile, findsOneWidget);
        await tester.tap(pile);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        expect(client.submissions, isEmpty);
        expect(tester.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'short enlarged online actions retain usable table and card targets',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'required_actor': 1,
          'cards': [
            {
              'card': {
                'id': 'short-hand',
                'rank': 12,
                'suit': 'hearts',
                'deck': 1,
              },
              'controller': 1,
              'zone': 'hand',
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'open-series');
      await tester.binding.setSurfaceSize(const Size(844, 390));
      tester
          .widget<PopupMenuButton<String>>(
            find.byWidgetPredicate(
              (widget) =>
                  widget is PopupMenuButton<String> &&
                  widget.tooltip != 'Game options',
            ),
          )
          .onSelected!('text');
      await tester.pumpAndSettle();
      final picker = find.byKey(const ValueKey('online-submit'));
      final scrolling = find
          .ancestor(of: picker, matching: find.byType(Scrollable))
          .first;
      final scroll = tester.state<ScrollableState>(scrolling);
      expect(scroll.position.viewportDimension, greaterThanOrEqualTo(48));
      final public = tester.getRect(
        find.byKey(const ValueKey('adaptive-public-square')),
      );
      expect(public.height, greaterThanOrEqualTo(96));
      final hand = find.byKey(const ValueKey('adaptive-hand-viewport'));
      final handRect = tester.getRect(hand);
      final card = find
          .descendant(of: hand, matching: find.byType(CgmsCard))
          .first;
      expect(
        tester.getRect(card).intersect(handRect).height,
        greaterThanOrEqualTo(48),
      );
      await tester.ensureVisible(picker);
      await tester.pumpAndSettle();
      expect(
        tester.getRect(picker).intersect(tester.getRect(scrolling)).height,
        greaterThanOrEqualTo(48),
      );
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'contextual replacement choices keep 48 pixel semantic touch targets',
    (tester) async {
      final semantics = tester.ensureSemantics();
      debugDefaultTargetPlatformOverride = TargetPlatform.linux;
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      try {
        await showOnline(tester, client, expand: false);
        await preparePeople(tester);
        for (final label in [
          'Jack',
          'Queen',
          'King',
          'Clubs',
          'Diamonds',
          'Hearts',
          'Spades',
        ]) {
          final field = find.widgetWithText(OutlinedButton, label);
          await tester.ensureVisible(field);
          await tester.pumpAndSettle();
          final node = tester.getSemantics(field);
          expect(
            node.getSemanticsData().hasAction(SemanticsAction.tap),
            isTrue,
          );
          expect(node.rect.height, greaterThanOrEqualTo(48), reason: label);
          expect(node.rect.width, greaterThanOrEqualTo(48), reason: label);
        }
        expect(client.submissions, isEmpty);
      } finally {
        semantics.dispose();
        debugDefaultTargetPlatformOverride = null;
      }
    },
  );

  testWidgets('lobby headings and introduction have a centered hierarchy', (
    tester,
  ) async {
    final client = ViewClient()..account = {'id': 'owner'};
    addTearDown(client.dispose);
    await showOnline(tester, client);
    final strings = CgmsLocalizations.of(
      tester.element(find.byType(OnlineScreen)),
    );
    for (final label in [
      strings.onlineWelcomeTitle,
      strings.onlineWelcomeBody,
      strings.onlineSetupTitle,
      strings.onlineSetupBody,
      strings.onlineCreateRoom,
      strings.onlineJoinRoom,
      strings.onlineBotSetupHelp,
    ]) {
      final text = tester.widgetList<Text>(find.text(label)).first;
      expect(text.textAlign, TextAlign.center, reason: label);
    }
    final intro = tester.widget<Text>(find.text(strings.onlineWelcomeTitle));
    final section = tester
        .widgetList<Text>(find.text(strings.onlineCreateRoom))
        .first;
    expect(intro.style!.fontSize, greaterThan(section.style!.fontSize!));
    expect(section.style!.fontSize, inInclusiveRange(22, 24));
  });

  testWidgets(
    'room dropdowns use body typography in selected and open values',
    (tester) async {
      final client = ViewClient()..account = {'id': 'owner'};
      addTearDown(client.dispose);
      await showOnline(tester, client);
      final picker = find.byKey(const ValueKey('bot-difficulty'));
      final dropdown = tester.widget<DropdownButton<String>>(
        find.descendant(
          of: picker,
          matching: find.byType(DropdownButton<String>),
        ),
      );
      final selectedStyle =
          dropdown.style ??
          Theme.of(tester.element(picker)).textTheme.titleMedium!;
      expect(selectedStyle.fontSize, 15);
      expect(selectedStyle.fontWeight ?? FontWeight.normal, FontWeight.normal);
      expect(dropdown.isExpanded, isTrue);
      expect(dropdown.itemHeight, isNull);
      final players = tester.widget<DropdownButton<int>>(
        find.byType(DropdownButton<int>),
      );
      expect(players.style!.fontSize, 15);
      expect(players.style!.fontWeight ?? FontWeight.normal, FontWeight.normal);
      await tester.ensureVisible(picker);
      await tester.tap(picker);
      await tester.pumpAndSettle();
      final option = find.text('Advanced').hitTestable();
      final rich = find.descendant(of: option, matching: find.byType(RichText));
      final span = tester.widget<RichText>(rich).text as TextSpan;
      expect(span.style!.fontSize, 15);
      expect(span.style!.fontWeight ?? FontWeight.normal, FontWeight.normal);
      await tester.tap(option);
      await tester.pumpAndSettle();
      expect(
        tester.widget<DropdownButtonFormField<String>>(picker).initialValue,
        'advanced',
      );
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'contextual choices retain readable body type and intrinsic rows',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await preparePeople(tester);
      expect(find.byType(DropdownButtonFormField<String>), findsNothing);
      for (final label in [
        'Jack',
        'Queen',
        'King',
        'Clubs',
        'Diamonds',
        'Hearts',
        'Spades',
      ]) {
        final option = find.widgetWithText(OutlinedButton, label);
        final rich = tester.widget<RichText>(
          find.descendant(
            of: find.descendant(of: option, matching: find.text(label)),
            matching: find.byType(RichText),
          ),
        );
        expect(rich.text.style!.fontSize, 16);
        expect(rich.textScaler.scale(16), 16);
        expect(tester.getSize(option).height, greaterThanOrEqualTo(48));
      }
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'contextual purchase stays centered across live breakpoints without losing drafts',
    (tester) async {
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'purchase');
      final quantity = find.widgetWithText(TextField, 'Quantity');
      await tester.enterText(quantity, '12');
      for (final width in [
        599.0,
        600.0,
        768.0,
        1049.0,
        1050.0,
        1356.0,
        1920.0,
        390.0,
      ]) {
        tester.view.physicalSize = Size(width, 948);
        await tester.binding.setSurfaceSize(Size(width, 948));
        await tester.pumpAndSettle();
        final popup = tester.getRect(
          find.byKey(const ValueKey('contextual-popup')),
        );
        expect(popup.center.dx, closeTo(width / 2, .01));
        expect(popup.width, lessThanOrEqualTo(560));
        expect(popup.left, greaterThanOrEqualTo(0));
        expect(popup.right, lessThanOrEqualTo(width));
        final submit = find.byKey(const ValueKey('online-submit'));
        await tester.ensureVisible(submit);
        await tester.pumpAndSettle();
        expect(tester.getSize(submit).height, greaterThanOrEqualTo(48));
        expect(tester.widget<TextField>(quantity).controller!.text, '12');
        expect(client.submissions, isEmpty);
        expect(tester.takeException(), isNull);
      }
    },
  );

  testWidgets(
    'roomy desktop contextual controls fit without reserving an action workspace',
    (tester) async {
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'purchase');
      for (final size in [
        const Size(1280, 1400),
        const Size(1280, 900),
        const Size(1449, 1400),
      ]) {
        tester.view.physicalSize = size;
        await tester.binding.setSurfaceSize(size);
        await tester.pumpAndSettle();
        final popup = tester.getRect(
          find.byKey(const ValueKey('contextual-popup')),
        );
        expect(popup.width, lessThanOrEqualTo(560));
        expect(popup.center.dx, closeTo(size.width / 2, .01));
        final quantity = find.widgetWithText(TextField, 'Quantity');
        await tester.ensureVisible(quantity);
        await tester.pumpAndSettle();
        expect(
          tester.getRect(quantity).intersect(popup).height,
          greaterThanOrEqualTo(48),
        );
        final submit = find.byKey(const ValueKey('online-submit'));
        await tester.ensureVisible(submit);
        await tester.pumpAndSettle();
        expect(
          tester.getRect(submit).intersect(popup).height,
          greaterThanOrEqualTo(48),
        );
        for (final text in ['Purchase', 'Confirm Purchase']) {
          final rich = tester.widget<RichText>(
            find.descendant(
              of: find.text(text),
              matching: find.byType(RichText),
            ),
          );
          expect(rich.text.style!.fontSize, 16);
          expect(rich.textScaler.scale(16), 16);
        }
        expect(find.byKey(const ValueKey('online-action-form')), findsNothing);
        expect(
          find.byKey(
            const PageStorageKey('online-action-panel'),
            skipOffstage: false,
          ),
          findsNothing,
        );
        expect(tester.takeException(), isNull);
      }
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'contextual replacement choices wrap at 200 percent with full RTL touch rows',
    (tester) async {
      tester.view.devicePixelRatio = 1;
      tester.view.physicalSize = const Size(319, 844);
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final semantics = tester.ensureSemantics();
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await preparePeople(tester);
      final settings = tester.widget<PopupMenuButton<String>>(
        find.byWidgetPredicate(
          (w) =>
              w is PopupMenuButton<String> && w.tooltip == 'Display settings',
        ),
      );
      settings.onSelected!('rtl');
      settings.onSelected!('text');
      await tester.binding.setSurfaceSize(const Size(319, 844));
      await tester.pumpAndSettle();
      final popup = find.byKey(const ValueKey('contextual-popup'));
      for (final label in ['Diamonds', 'Hearts', 'Spades']) {
        final option = find.widgetWithText(OutlinedButton, label);
        await tester.ensureVisible(option);
        await tester.pumpAndSettle();
        final text = find.descendant(of: option, matching: find.text(label));
        final rich = tester.widget<RichText>(
          find.descendant(of: text, matching: find.byType(RichText)),
        );
        expect(rich.textScaler.scale(14), 28);
        expect(Directionality.of(tester.element(text)), TextDirection.rtl);
        final painter = TextPainter(
          text: rich.text,
          textDirection: TextDirection.rtl,
          textScaler: rich.textScaler,
        )..layout(maxWidth: tester.getSize(text).width);
        expect(
          tester.getSize(text).height,
          greaterThanOrEqualTo(painter.height),
        );
        painter.dispose();
        expect(tester.getSize(option).height, greaterThanOrEqualTo(48));
        expect(
          tester.getRect(option).intersect(tester.getRect(popup)).height,
          greaterThanOrEqualTo(48),
        );
        expect(tester.getRect(option).right, lessThanOrEqualTo(319));
      }
      final option = find.widgetWithText(OutlinedButton, 'Diamonds');
      await tester.ensureVisible(option);
      await tester.tap(option);
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<Semantics>(
              find
                  .ancestor(
                    of: option,
                    matching: find.byWidgetPredicate(
                      (w) => w is Semantics && w.properties.selected != null,
                    ),
                  )
                  .first,
            )
            .properties
            .selected,
        isTrue,
      );
      expect(tester.takeException(), isNull);
      expect(client.submissions, isEmpty);
      semantics.dispose();
    },
  );

  testWidgets(
    'server readiness gates the provisional starter without refresh',
    (tester) async {
      final client = ViewClient()
        ..publish({'round': 1}, online: {'server_pending': true});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      final end = find.widgetWithText(FilledButton, 'End Turn');
      expect(tester.widget<FilledButton>(end).onPressed, isNull);
      expect(find.text('Preparing the next move…'), findsWidgets);
      expect(
        commandTiming(
          'open-series',
          client.snapshot!.board,
          client.snapshot!.online,
        ),
        CommandTiming.automatic,
      );
      client.publish(
        {'round': 1, 'active': 2},
        online: {'server_pending': false},
      );
      await tester.pumpAndSettle();
      expect(tester.widget<FilledButton>(end).onPressed, isNull);
      expect(find.text('Preparing the next move…'), findsNothing);
      client.publish({'round': 1}, online: {'server_pending': false});
      await tester.pumpAndSettle();
      expect(tester.widget<FilledButton>(end).onPressed, isNotNull);
      expect(client.submissions, isEmpty);
      await tester.ensureVisible(end);
      await tester.tap(end);
      await tester.pumpAndSettle();
      expect(client.submissions, [
        {'type': 'end-turn', 'payload': {}},
      ]);
      expect(tester.takeException(), isNull);
    },
  );

  for (final scale in [1.0, 2.0]) {
    testWidgets('phone round choice accepts a real tap at text scale $scale', (
      tester,
    ) async {
      tester.platformDispatcher.textScaleFactorTestValue = scale;
      addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
      final client = DepartureClient()
        ..publish(
          {'active': 2},
          online: {'round_closed': true, 'departure_pending': true},
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.binding.setSurfaceSize(const Size(319, 844));
      await tester.pumpAndSettle();
      final stay = find.widgetWithText(FilledButton, 'Stay in game');
      await tester.ensureVisible(stay);
      await tester.pumpAndSettle();
      expect(tester.getSize(stay).height, greaterThanOrEqualTo(48));
      await tester.tap(stay);
      await tester.pump();
      expect(client.submissions, [
        {
          'type': 'departure-choice',
          'payload': {'accept': false},
        },
      ]);
      client.completion.complete();
      await tester.pumpAndSettle();
      expect(find.text('Stay in game'), findsNothing);
      expect(tester.takeException(), isNull);
    });
  }

  for (final leave in [false, true]) {
    testWidgets('round boundary explicitly chooses departure $leave once', (
      tester,
    ) async {
      final client = DepartureClient()
        ..publish(
          {'active': 2},
          online: {'round_closed': true, 'departure_pending': true},
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.byKey(const ValueKey('close-card-context')), findsOneWidget);
      expect(
        find.text('Round complete · Choose whether to stay'),
        findsOneWidget,
      );
      expect(client.submissions, isEmpty);
      for (final width in [319.0, 600.0, 1050.0]) {
        await tester.binding.setSurfaceSize(Size(width, 844));
        await tester.pumpAndSettle();
        expect(find.text('Stay in game'), findsOneWidget);
        expect(find.text('Leave game'), findsOneWidget);
        expect(client.submissions, isEmpty);
      }
      final stay = find.widgetWithText(FilledButton, 'Stay in game');
      final depart = find.widgetWithText(OutlinedButton, 'Leave game');
      final action = leave
          ? tester.widget<OutlinedButton>(depart).onPressed!
          : tester.widget<FilledButton>(stay).onPressed!;
      action();
      action();
      await tester.pump();
      expect(client.submissions, [
        {
          'type': 'departure-choice',
          'payload': {'accept': leave},
        },
      ]);
      expect(tester.widget<FilledButton>(stay).onPressed, isNull);
      expect(tester.widget<OutlinedButton>(depart).onPressed, isNull);
      client.completion.complete();
      await tester.pumpAndSettle();
      expect(find.text('Round complete · Waiting for players'), findsOneWidget);
      expect(find.text('Stay in game'), findsNothing);
      expect(find.widgetWithText(FilledButton, 'End Turn'), findsNothing);
      expect(client.submissions, hasLength(1));
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets(
    'round decision dismissal remains pending and reopens only at a new boundary',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'purchase');
      final quantity = find.widgetWithText(TextField, 'Quantity');
      await tester.enterText(quantity, '3');
      final draft = tester.widget<TextField>(quantity).controller!;
      client.publish(
        {},
        online: {'round_closed': true, 'departure_pending': true},
      );
      await tester.pumpAndSettle();
      expect(find.text('Stay in game'), findsOneWidget);
      expect(draft.text, '3');
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      client.publish(
        {},
        online: {'round_closed': true, 'departure_pending': true},
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(find.text('Pending decision'), findsOneWidget);
      expect(draft.text, '3');
      client.publish(
        {'round': 5},
        online: {'round_closed': true, 'departure_pending': true},
      );
      await tester.pumpAndSettle();
      expect(find.text('Stay in game'), findsOneWidget);
      client.publish({'round': 6});
      await tester.pumpAndSettle();
      expect(draft.text, '3');
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'bot room setup exposes beginner difficulty and preserves human rooms',
    (tester) async {
      final client = ViewClient()..account = {'id': 'owner'};
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(
        find.widgetWithText(FilledButton, 'Play against bots'),
        findsOneWidget,
      );
      expect(find.text('Bot difficulty'), findsOneWidget);
      expect(find.text('Beginner'), findsOneWidget);
      expect(
        find.widgetWithText(OutlinedButton, 'Create room'),
        findsOneWidget,
      );
      expect(find.text('Players'), findsOneWidget);
      expect(find.text('Games per match'), findsOneWidget);
    },
  );

  testWidgets(
    'bot room creation validates games and prevents duplicate requests',
    (tester) async {
      final transport = FakeTransport();
      final pending = Completer<OnlineResponse>();
      transport.handle = (_, path, _) async => path == '/v1/session'
          ? const OnlineResponse(200, {
              'account': {'id': 'owner'},
              'csrf_token': 'test-csrf',
            })
          : path == '/v1/rooms'
          ? await pending.future
          : const OnlineResponse(200, {'enabled': false});
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final games = find.byKey(const PageStorageKey('room-games'));
      final play = find.widgetWithText(FilledButton, 'Play against bots');
      await tester.enterText(games, '0');
      await tester.ensureVisible(play);
      await tester.tap(play);
      await tester.pumpAndSettle();
      expect(
        find.text('Enter a whole number of games from 1 to 100.'),
        findsOneWidget,
      );
      expect(
        transport.calls.where((call) => call['path'] == '/v1/rooms'),
        isEmpty,
      );
      await tester.enterText(games, '3');
      final create = tester.widget<FilledButton>(play).onPressed!;
      create();
      create();
      await tester.pump();
      final requests = transport.calls
          .where((call) => call['path'] == '/v1/rooms')
          .toList();
      expect(requests, hasLength(1));
      expect(
        requests.single['body'],
        containsPair('bot_difficulty', 'beginner'),
      );
      expect(requests.single['body'], containsPair('capacity', 4));
      expect(requests.single['body'], containsPair('games_per_match', 3));
      expect(tester.widget<FilledButton>(play).onPressed, isNull);
      expect(
        tester
            .widget<OutlinedButton>(
              find.widgetWithText(OutlinedButton, 'Create room'),
            )
            .onPressed,
        isNull,
      );
      expect(
        tester
            .widget<DropdownButtonFormField<String>>(
              find.byKey(const ValueKey('bot-difficulty')),
            )
            .onChanged,
        isNull,
      );
      expect(tester.widget<TextFormField>(games).enabled, isFalse);
      pending.complete(
        const OnlineResponse(200, {
          'id': 'bot-room',
          'owner_id': 'owner',
          'capacity': 4,
          'games': 3,
          'bot_difficulty': 'beginner',
          'members': [
            {'seat': 1},
            {'seat': 2, 'bot': true},
            {'seat': 3, 'bot': true},
            {'seat': 4, 'bot': true},
          ],
        }),
      );
      await tester.pumpAndSettle();
      expect(find.text('Bot · Beginner'), findsNWidgets(3));
      expect(find.text('Joined'), findsOneWidget);
      expect(find.text('Create invitation'), findsNothing);
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Start match'),
            )
            .onPressed,
        isNotNull,
      );
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('bot difficulty survives resizing and human creation omits it', (
    tester,
  ) async {
    final transport = FakeTransport()
      ..handle = (_, path, _) async => path == '/v1/session'
          ? const OnlineResponse(200, {
              'account': {'id': 'owner'},
              'csrf_token': 'test-csrf',
            })
          : const OnlineResponse(200, {
              'id': 'human-room',
              'owner_id': 'owner',
              'capacity': 4,
              'games': 3,
              'members': [
                {'seat': 1},
              ],
            });
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    final difficulty = find.byKey(const ValueKey('bot-difficulty'));
    await tester.ensureVisible(difficulty);
    await tester.tap(difficulty);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Advanced').last);
    await tester.pumpAndSettle();
    await tester.binding.setSurfaceSize(const Size(319, 844));
    await tester.pumpAndSettle();
    expect(find.text('Advanced'), findsOneWidget);
    final human = find.widgetWithText(OutlinedButton, 'Create room');
    await tester.ensureVisible(human);
    await tester.tap(human);
    await tester.pumpAndSettle();
    final body =
        transport.calls.singleWhere(
              (call) => call['path'] == '/v1/rooms',
            )['body']
            as Map;
    expect(body.containsKey('bot_difficulty'), isFalse);
    expect(find.text('Bot · Advanced'), findsNothing);
    expect(find.text('Create invitation'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('game bar semantics follow live RTL and text size changes', (
    tester,
  ) async {
    // Register an isolated alias so this visual regression does not change
    // the synthetic-font fixtures in unrelated tests in this file.
    const barFont = 'CGMS bar regression';
    final loader = FontLoader(barFont);
    for (final asset in [
      'AtkinsonHyperlegibleNext-Regular.ttf',
      'AtkinsonHyperlegibleNext-Bold.ttf',
    ]) {
      loader.addFont(rootBundle.load('packages/cgms_ui/assets/fonts/$asset'));
    }
    await loader.load();
    final semantics = tester.ensureSemantics();
    try {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      final base = CgmsTheme.dark();
      await showOnline(
        tester,
        client,
        expand: false,
        theme: base.copyWith(
          textTheme: base.textTheme.apply(fontFamily: barFont),
        ),
      );
      await tester.binding.setSurfaceSize(const Size(319, 948));
      await tester.pumpAndSettle();
      final ids = <String, int>{};
      void checkBounds(Finder target, String id) {
        final node = tester.getSemantics(target);
        var semanticBounds = node.rect;
        var ancestor = node;
        while (true) {
          if (ancestor.transform case final transform?) {
            semanticBounds = MatrixUtils.transformRect(
              transform,
              semanticBounds,
            );
          }
          final parent = ancestor.parent;
          if (parent == null) break;
          ancestor = parent;
        }
        // The semantics root maps logical geometry into physical pixels.
        final ratio = tester.view.devicePixelRatio;
        semanticBounds = Rect.fromLTRB(
          semanticBounds.left / ratio,
          semanticBounds.top / ratio,
          semanticBounds.right / ratio,
          semanticBounds.bottom / ratio,
        );
        final painted = tester.getRect(target);
        expect(semanticBounds.left, closeTo(painted.left, 1 / 64), reason: id);
        expect(semanticBounds.top, closeTo(painted.top, 1 / 64), reason: id);
        expect(
          semanticBounds.width,
          closeTo(painted.width, 1 / 64),
          reason: id,
        );
        expect(
          semanticBounds.height,
          closeTo(painted.height, 1 / 64),
          reason: id,
        );
      }

      void check() {
        for (final id in ['game-header', 'header-status']) {
          final marker = find.byWidgetPredicate(
            (widget) =>
                widget is Semantics && widget.properties.identifier == id,
          );
          final painted = tester.getSize(marker);
          final node = tester.getSemantics(marker);
          ids.putIfAbsent(id, () => node.id);
          expect(node.id, ids[id], reason: id);
          expect(node.rect.width, closeTo(painted.width, 1 / 64), reason: id);
          expect(node.rect.height, closeTo(painted.height, 1 / 64), reason: id);
          checkBounds(marker, id);
        }
        for (final (target, id) in [
          (find.byTooltip('Game options'), 'trade'),
          (find.widgetWithText(FilledButton, 'End Turn'), 'end turn'),
        ]) {
          checkBounds(target, id);
        }
        final primary = tester.getRect(
          find.descendant(
            of: find.widgetWithText(FilledButton, 'End Turn'),
            matching: find.byType(InkWell),
          ),
        );
        final secondary = tester.getRect(
          find.descendant(
            of: find.byTooltip('Game options'),
            matching: find.byType(InkWell),
          ),
        );
        expect(primary.height, greaterThanOrEqualTo(48));
        expect(secondary.height, greaterThanOrEqualTo(48));
        expect(primary.center.dy, closeTo(secondary.center.dy, 1 / 64));
      }

      check();
      for (final change in ['rtl', 'text', 'text', 'rtl']) {
        await tester.binding.setSurfaceSize(const Size(319, 948));
        await tester.pumpAndSettle();
        await tester.tap(find.byTooltip('Display settings'));
        await tester.pumpAndSettle();
        await tester.tap(
          find.widgetWithText(
            CheckedPopupMenuItem<String>,
            change == 'rtl' ? 'Right to left' : '200% text',
          ),
        );
        await tester.pumpAndSettle();
        check();
        await tester.binding.setSurfaceSize(const Size(390, 844));
        await tester.pumpAndSettle();
        check();
      }
      expect(tester.takeException(), isNull);
    } finally {
      semantics.dispose();
    }
  });

  testWidgets(
    'public semantics follow visible scrolled table through live RTL and text size changes',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.binding.setSurfaceSize(const Size(844, 390));
      await tester.pumpAndSettle();
      final marker = find.byWidgetPredicate(
        (w) => w is Semantics && w.properties.identifier == 'public-square',
      );
      void check() {
        final square = tester.getRect(
          find.byKey(const ValueKey('adaptive-public-square')),
        );
        final data = tester.getSemantics(marker).getSemanticsData();
        final bounds = tester.getSemantics(marker).rect;
        expect(
          data.label,
          CgmsLocalizations.of(tester.element(marker)).actionBoardPublic,
        );
        expect(bounds.width, greaterThanOrEqualTo(96));
        expect(bounds.height, greaterThanOrEqualTo(96));
        expect(bounds.width, lessThanOrEqualTo(square.width));
        expect(bounds.height, lessThanOrEqualTo(square.height));
        expect(
          find.byKey(const ValueKey('adaptive-draw-pile')),
          findsOneWidget,
        );
      }

      check();
      for (final change in [
        'Right to left',
        '200% text',
        '200% text',
        'Right to left',
      ]) {
        await tester.tap(find.byTooltip('Display settings'));
        await tester.pumpAndSettle();
        await tester.tap(
          find.widgetWithText(CheckedPopupMenuItem<String>, change),
        );
        await tester.pumpAndSettle();
        expect(find.byType(CheckedPopupMenuItem<String>), findsNothing);
        check();
      }
      expect(tester.takeException(), isNull);
      semantics.dispose();
    },
  );

  testWidgets(
    'footer buttons retain their full 48 pixel semantic hit targets',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      for (final size in [const Size(390, 844), const Size(844, 390)]) {
        await tester.binding.setSurfaceSize(size);
        await tester.pumpAndSettle();
        final footer = tester.getRect(utilityBar());
        for (final button in [
          find.byTooltip('Game options'),
          find.widgetWithText(FilledButton, 'End Turn'),
        ]) {
          final node = tester.getSemantics(button);
          final target = tester.getRect(button);
          expect(node.rect.height, greaterThanOrEqualTo(48));
          expect(target.intersect(footer).height, greaterThanOrEqualTo(48));
        }
        expect(tester.takeException(), isNull);
      }
      final settings = tester.widget<PopupMenuButton<String>>(
        find.byWidgetPredicate(
          (widget) =>
              widget is PopupMenuButton<String> &&
              widget.tooltip != 'Game options',
        ),
      );
      settings.onSelected!('rtl');
      settings.onSelected!('text');
      await tester.pumpAndSettle();
      for (final button in [
        find.byTooltip('Game options'),
        find.widgetWithText(FilledButton, 'End Turn'),
      ]) {
        expect(
          tester.getSemantics(button).rect.height,
          greaterThanOrEqualTo(48),
        );
      }
      for (final size in [const Size(1440, 900), const Size(1440, 1400)]) {
        await tester.binding.setSurfaceSize(size);
        await tester.pumpAndSettle();
        for (final label in ['End Turn']) {
          final text = find.text(label);
          final element = tester.element(text);
          final painter = TextPainter(
            text: TextSpan(
              text: label,
              style: DefaultTextStyle.of(element).style,
            ),
            textDirection: Directionality.of(element),
            textScaler: MediaQuery.textScalerOf(element),
          )..layout();
          expect(
            tester.getSize(text).height,
            greaterThanOrEqualTo(painter.height),
          );
          painter.dispose();
        }
        expect(tester.takeException(), isNull);
      }
      semantics.dispose();
    },
  );

  testWidgets(
    'required decision remains available after an open public chooser closes',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.tap(find.byTooltip('Inspect public cards').first);
      await tester.pumpAndSettle();
      expect(find.text('Close public cards'), findsOneWidget);
      client.publish({'required_actor': 1});
      await tester.pumpAndSettle();
      expect(find.text('Close public cards').hitTestable(), findsOneWidget);
      await tester.tap(find.text('Close public cards'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('close-card-context')), findsOneWidget);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'short enlarged viewport scrolls readable private cards while utilities remain reachable',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [
            {
              'card': {
                'id': 'short-hand',
                'rank': 12,
                'suit': 'hearts',
                'deck': 1,
              },
              'controller': 1,
              'zone': 'hand',
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.binding.setSurfaceSize(const Size(844, 390));
      tester
          .widget<PopupMenuButton<String>>(
            find.byWidgetPredicate(
              (w) =>
                  w is PopupMenuButton<String> &&
                  w.tooltip == 'Display settings',
            ),
          )
          .onSelected!('text');
      await tester.pumpAndSettle();
      final card = find.byKey(const ValueKey('adaptive-select-short-hand'));
      await tester.ensureVisible(card);
      await tester.pumpAndSettle();
      final footer = tester.getRect(utilityBar());
      final bounds = tester.getRect(card);
      expect(bounds.height, greaterThanOrEqualTo(48));
      expect(bounds.top, lessThan(footer.top));
      expect(bounds.bottom, lessThanOrEqualTo(footer.top));
      expect(card.hitTestable(), findsOneWidget);
      expect(footer.height, greaterThanOrEqualTo(48));
      for (final control in [
        find.byTooltip('Game options'),
        find.widgetWithText(FilledButton, 'End Turn'),
      ]) {
        final target = tester.getRect(control);
        expect(target.intersect(footer), target);
        expect(target.height, greaterThanOrEqualTo(48));
      }
      expect(tester.takeException(), isNull);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'closing actions allows later required and settlement decisions to reopen',
    (tester) async {
      final client = ViewClient()..publish({'required_actor': 1});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.byKey(const ValueKey('close-card-context')), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('close-card-context')), findsNothing);
      client.publish({'required_actor': 2});
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('close-card-context')), findsNothing);
      client.publish({'required_actor': 1});
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('close-card-context')), findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      client.publish({'phase': 'settlement'});
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('close-card-context')), findsOneWidget);
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
      await chooseGameOption(tester, 'Finish Settlement');
      expect(find.text('Confirm Finish Settlement'), findsOneWidget);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets('closed board announces status and disables End Turn', (
    tester,
  ) async {
    final client = ViewClient()..publish({'phase': 'settlement'});
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    final board = tester.widget<CgmsAdaptiveTable>(
      find.byType(CgmsAdaptiveTable),
    );
    expect(board.data.turnLabel, 'Board closed · settlement');
    expect(find.text('Board closed · settlement'), findsOneWidget);
    expect(
      find.ancestor(
        of: find.text('Board closed · settlement'),
        matching: find.byWidgetPredicate(
          (widget) =>
              widget is Semantics && widget.properties.liveRegion == true,
        ),
      ),
      findsOneWidget,
    );
    expect(find.widgetWithText(FilledButton, 'End Turn'), findsNothing);
    expect(client.submissions, isEmpty);
  });

  testWidgets(
    'settlement opens automatically from the unobstructed ordinary board',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      client.publish(
        {'phase': 'settlement'},
        online: {'ending': 'coup', 'declarer': 1},
      );
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('contextual-popup')), findsOneWidget);
      expect(find.text('Ending: coup · Declarer: seat 1'), findsOneWidget);
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
      await chooseGameOption(tester, 'Finish Settlement');
      expect(find.text('Confirm Finish Settlement'), findsOneWidget);
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('finalized standings and admission recovery open automatically', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish(
        {'phase': 'settlement'},
        online: {
          'financial_phase': 'finalized',
          'completed_games': 3,
          'ranks': [1, 1, 3],
        },
      );
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    expect(find.text('Match complete'), findsOneWidget);
    expect(find.text('Match ranks by seat: [1, 1, 3]'), findsOneWidget);
    await tester.tap(find.byKey(const ValueKey('close-card-context')));
    await tester.pumpAndSettle();
    client.publish(
      {'phase': 'settlement'},
      online: {'financial_phase': 'finalized', 'completed_games': 1},
      admissionWaiting: true,
    );
    await tester.pumpAndSettle();
    expect(find.textContaining('next game is waiting'), findsOneWidget);
    expect(find.text('Match complete'), findsNothing);
    expect(client.submissions, isEmpty);
  });

  testWidgets('superseded completion cannot reopen the next ordinary turn', (
    tester,
  ) async {
    final client = ViewClient()..publish({});
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    client.publish({'phase': 'settlement'});
    client.publish({'game_id': 'next-game'});
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
    expect(find.text('Board closed · settlement'), findsNothing);
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'ordinary authority updates preserve action draft focus and collapse',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      chooseAction(tester, 'purchase');
      await tester.pumpAndSettle();
      final quantity = find.widgetWithText(TextField, 'Quantity');
      await tester.ensureVisible(quantity);
      await tester.enterText(quantity, '12');
      final edit = tester.widget<EditableText>(
        find.descendant(of: quantity, matching: find.byType(EditableText)),
      );
      edit.controller.selection = const TextSelection(
        baseOffset: 0,
        extentOffset: 1,
      );
      final pending = <String, dynamic>{'operation_id': 'preserved-operation'};
      client.pendingUnknown = pending;
      client.busy = true;
      for (final width in [1449.0, 599.0, 600.0, 1049.0, 1050.0, 319.0]) {
        await tester.binding.setSurfaceSize(Size(width, 948));
        // The pending request intentionally keeps the progress indicator active.
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 300));
        final current = tester.widget<EditableText>(
          find.descendant(of: quantity, matching: find.byType(EditableText)),
        );
        expect(current.focusNode, same(edit.focusNode));
        expect(current.focusNode.hasFocus, isTrue);
        expect(current.controller, same(edit.controller));
        expect(current.controller.text, '12');
        expect(
          current.controller.selection,
          const TextSelection(baseOffset: 0, extentOffset: 1),
        );
        expect(client.unknownOperation, same(pending));
        expect(
          tester
              .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
              .onPressed,
          isNull,
        );
        expect(client.submissions, isEmpty);
        expect(tester.takeException(), isNull);
      }
      client.pendingUnknown = null;
      client.busy = false;
      client.publish({});
      await tester.pumpAndSettle();
      expect(edit.focusNode.hasFocus, isTrue);
      expect(edit.controller.text, '12');
      expect(
        edit.controller.selection,
        const TextSelection(baseOffset: 0, extentOffset: 1),
      );
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      client.publish({});
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(edit.controller.text, '12');
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'contextual choice keyboard focus reveals scrolled card parameters',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      try {
        await showOnline(tester, client, expand: false);
        await preparePeople(tester);
        await showActionGuidance(tester);
        await tester.binding.setSurfaceSize(const Size(390, 844));
        await tester.pumpAndSettle();
        final choice = find.widgetWithText(OutlinedButton, 'Diamonds');
        await tester.ensureVisible(choice);
        await tester.pumpAndSettle();
        expect(
          tester
              .getSemantics(choice)
              .getSemanticsData()
              .hasAction(SemanticsAction.focus),
          isTrue,
        );
        final scroll = find
            .ancestor(of: choice, matching: find.byType(Scrollable))
            .first;
        await tester.drag(scroll, const Offset(0, -700));
        await tester.pumpAndSettle();
        final viewport = tester.getRect(scroll);
        expect(tester.getBottomLeft(choice).dy, lessThan(viewport.top));
        final node = tester.getSemantics(choice);
        node.owner!.performAction(node.id, SemanticsAction.focus);
        await tester.pumpAndSettle();
        expect(FocusManager.instance.primaryFocus!.hasFocus, isTrue);
        final bounds = tester.getRect(choice);
        expect(bounds.top, greaterThanOrEqualTo(viewport.top));
        expect(bounds.bottom, lessThanOrEqualTo(viewport.bottom));
        await tester.sendKeyEvent(LogicalKeyboardKey.enter);
        await tester.pumpAndSettle();
        expect(
          tester
              .widget<Semantics>(
                find
                    .ancestor(
                      of: choice,
                      matching: find.byWidgetPredicate(
                        (w) => w is Semantics && w.properties.selected != null,
                      ),
                    )
                    .first,
              )
              .properties
              .selected,
          isTrue,
        );
        expect(client.submissions, isEmpty);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'dismissed contextual draft excludes keyboard focus and preserves caret on card reentry',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [
            {
              'card': {
                'id': 'draft-spade',
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
      await showOnline(tester, client, expand: false);
      await activatePrivateCard(tester, 'draft-spade');
      final purchase = find.byKey(const ValueKey('card-intent-purchase'));
      await tester.ensureVisible(purchase);
      await tester.tap(purchase);
      await tester.pumpAndSettle();
      final field = find.widgetWithText(TextField, 'Quantity');
      await tester.enterText(field, '123');
      final edit = tester.widget<EditableText>(
        find.descendant(of: field, matching: find.byType(EditableText)),
      );
      edit.controller.selection = const TextSelection(
        baseOffset: 1,
        extentOffset: 2,
      );
      final controller = edit.controller;
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      expect(find.byType(EditableText), findsNothing);
      for (var i = 0; i < 25; i++) {
        await tester.sendKeyEvent(LogicalKeyboardKey.tab);
        await tester.pumpAndSettle();
        expect(find.byType(EditableText), findsNothing);
      }
      await activatePrivateCard(tester, 'draft-spade');
      final restored = tester.widget<EditableText>(
        find.descendant(of: field, matching: find.byType(EditableText)),
      );
      expect(restored.controller, same(controller));
      expect(controller.text, '123');
      expect(
        controller.selection,
        const TextSelection(baseOffset: 1, extentOffset: 2),
      );
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    'required actor opens context initially and on authority changes',
    (tester) async {
      final client = ViewClient()..publish({'required_actor': 1});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final popup = find.byKey(const ValueKey('contextual-popup'));
      expect(popup, findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      client.publish({'required_actor': 2});
      await tester.pumpAndSettle();
      expect(popup, findsNothing);
      client.publish({'required_actor': 1});
      await tester.pumpAndSettle();
      expect(popup, findsOneWidget);
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      client.publish({'required_actor': 1});
      client.publish({'required_actor': 2});
      await tester.pumpAndSettle();
      expect(popup, findsNothing);
      client.publish({'required_actor': 1});
      client.snapshot = null;
      client.notifyListeners();
      await tester.pumpAndSettle();
      client.publish({});
      await tester.pumpAndSettle();
      expect(popup, findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'startup loading disables entry and automatically restores the session',
    (tester) async {
      final transport = FakeTransport();
      var attempts = 0;
      transport.handle = (_, path, _) async {
        if (path != '/v1/session')
          return const OnlineResponse(503, {'code': 'UNAVAILABLE'});
        attempts++;
        if (attempts == 1)
          return const OnlineResponse(503, {'code': 'UNAVAILABLE'});
        return const OnlineResponse(200, {
          'account': {'id': 'owner'},
          'csrf_token': 'private',
        });
      };
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: OnlineScreen(client: client),
        ),
      );
      await tester.pump();
      expect(
        find.text('Connecting to the table… attempt 1 of 8'),
        findsOneWidget,
      );
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Continue as guest'),
            )
            .onPressed,
        isNull,
      );
      expect(find.byKey(const ValueKey('online-error')), findsNothing);
      await tester.pump(const Duration(milliseconds: 250));
      await tester.pumpAndSettle();
      expect(find.textContaining('Connecting to the table'), findsNothing);
      expect(
        find.widgetWithText(OutlinedButton, 'Create room'),
        findsOneWidget,
      );
      expect(attempts, 2);
      expect(transport.calls.every((call) => call['method'] == 'GET'), isTrue);
    },
  );

  testWidgets(
    'open inspector follows captured card style and removes concealed identities',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final players = [
        {
          'seat': 1,
          'card_style': 'rain-glaze',
          'hand_count': 0,
          'ace_count': 0,
        },
        {'seat': 2, 'card_style': 'drypoint', 'hand_count': 8, 'ace_count': 2},
      ];
      final card = {
        'card': {
          'id': 'capture-diamond',
          'rank': 8,
          'deck': 2,
          'suit': 'diamonds',
        },
        'controller': 2,
        'zone': 'series',
        'available_from_round': 0,
      };
      final client = ViewClient()
        ..publish({
          'players': players,
          'cards': [card],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client);
      await openPublicSeat(tester, 'capture-diamond');
      await tester.tap(find.byKey(const ValueKey('magnify-capture-diamond')));
      await tester.pumpAndSettle();
      Finder inspector() => find.byWidgetPredicate(
        (w) =>
            w is CgmsCard &&
            w.key == const ValueKey('online-inspected-card') &&
            w.card.id == 'capture-diamond',
      );
      expect(
        tester.widget<CgmsCard>(inspector()).card.style,
        CardStyle.drypoint,
      );
      expect(
        tester.widget<CgmsCard>(inspector()).width,
        greaterThanOrEqualTo(200),
      );
      card['controller'] = 1;
      card['available_from_round'] = 5;
      client.publish({
        'players': players,
        'cards': [card],
      });
      await tester.pumpAndSettle();
      final captured = tester.widget<CgmsCard>(inspector()).card;
      expect(captured.style, CardStyle.rainGlaze);
      expect(captured.id, 'capture-diamond');
      expect(captured.copy, 2);
      expect(captured.status, 'Available from round 5');
      client.publish({'players': players, 'cards': []});
      await tester.pumpAndSettle();
      expect(inspector(), findsNothing);
      expect(find.text('Return to decisions'), findsNothing);
      expect(find.bySemanticsLabel('8 of Diamonds, copy 2'), findsNothing);
      client.publish({
        'players': players,
        'cards': [card],
      });
      await tester.pumpAndSettle();
      expect(
        inspector(),
        findsNothing,
        reason: 'A later reveal must not reopen a dismissed private context.',
      );
      await openPublicSeat(tester, 'capture-diamond');
      await tester.tap(find.byKey(const ValueKey('magnify-capture-diamond')));
      await tester.pumpAndSettle();
      expect(inspector(), findsOneWidget);
      client.publish({
        'game_id': 'next-game',
        'players': players,
        'cards': [card],
      });
      await tester.pumpAndSettle();
      expect(
        inspector(),
        findsNothing,
        reason:
            'A new game must not retarget an old inspector even if a fixture reuses an ID.',
      );
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
      semantics.dispose();
    },
  );

  testWidgets(
    'actionable offers follow selected action controls and precede general quotas',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {
            'proposals': [
              {
                'id': 'incoming-offer',
                'from': 2,
                'revision': 2,
                'status': 'offered',
                'terms': {
                  'to': 1,
                  'give': ['queen'],
                  'receive': [],
                },
              },
            ],
            'cards': [
              {
                'card': {
                  'id': 'queen',
                  'rank': 12,
                  'deck': 2,
                  'suit': 'hearts',
                },
                'controller': 2,
                'zone': 'formation',
                'allocation': 'x',
                'available_from_round': 0,
              },
            ],
          },
          online: {
            'rules_context': {
              'purchase_unit_price': 5,
              'quota_used': {'ponzi': true, 'fate': false},
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'purchase');
      final submit = tester
          .getTopLeft(find.byKey(const ValueKey('online-submit')))
          .dy;
      expect(
        tester
            .getTopLeft(
              find.text(
                'Unit price 5 · Quantity 1 · Effective payment 0 · Cost 5 · Excess -5',
              ),
            )
            .dy,
        lessThan(submit),
      );
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      await chooseGameOption(tester, 'Offers, promises & scores');
      final accept = find.widgetWithText(TextButton, 'Accept Offer');
      expect(accept, findsOneWidget);
      expect(client.submissions, isEmpty);
      expect(tester.widget<TextButton>(accept).onPressed, isNotNull);
      await tester.ensureVisible(accept);
      await tester.tap(accept);
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'accept-offer',
        'payload': {'offer_id': 'incoming-offer', 'revision': 2},
      });
      await chooseGameOption(tester, 'Card gestures');
      expect(find.text('Ponzi quota: used'), findsOneWidget);
    },
  );

  testWidgets(
    'settlement ending omits absent declarer and retains supplied seat',
    (tester) async {
      final client = ViewClient()
        ..publish({'phase': 'settlement'}, online: {'ending': 'ordinary'});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(find.text('Ending: ordinary'), findsOneWidget);
      expect(find.textContaining('seat null'), findsNothing);
      expect(find.textContaining('Declarer:'), findsNothing);
      client.publish(
        {'phase': 'settlement'},
        online: {'ending': 'coup', 'declarer': 2},
      );
      await tester.pumpAndSettle();
      expect(find.text('Ending: coup · Declarer: seat 2'), findsOneWidget);
      expect(find.text('Ending: ordinary'), findsNothing);
    },
  );

  testWidgets(
    'collapsed allowance retains state but excludes hidden keyboard targets',
    (tester) async {
      final client = ViewClient()..account = {'id': 'owner'};
      addTearDown(client.dispose);
      await showOnline(tester, client);
      final body = find.byType(EconomyPanel, skipOffstage: false);
      final retainedState = tester.state(body);
      bool focusIsInsideBody() {
        final bodyElement = tester.element(body);
        var inside = false;
        FocusManager.instance.primaryFocus?.context?.visitAncestorElements((
          element,
        ) {
          if (element == bodyElement) inside = true;
          return !inside;
        });
        return inside;
      }

      for (var i = 0; i < 20; i++) {
        await tester.sendKeyEvent(LogicalKeyboardKey.tab);
        await tester.pumpAndSettle();
        expect(
          focusIsInsideBody(),
          isFalse,
          reason: 'Collapsed allowance keyboard step $i',
        );
      }
      final disclosure = find.byType(ExpansionTile);
      await tester.ensureVisible(disclosure);
      await tester.tap(disclosure);
      await tester.pumpAndSettle();
      var reachedExpandedBody = false;
      for (var i = 0; i < 20 && !reachedExpandedBody; i++) {
        await tester.sendKeyEvent(LogicalKeyboardKey.tab);
        await tester.pumpAndSettle();
        reachedExpandedBody = focusIsInsideBody();
      }
      expect(reachedExpandedBody, isTrue);
      expect(tester.state(body), same(retainedState));
    },
  );

  testWidgets('selected costs precede submission and general quotas follow it', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish(
        {},
        online: {
          'rules_context': {
            'purchase_unit_price': 5,
            'quota_used': {'ponzi': true, 'fate': false},
          },
        },
      );
    addTearDown(client.dispose);
    await showOnline(tester, client, preparedAction: 'purchase');
    await showActionGuidance(tester);
    final submit = tester
        .getTopLeft(find.byKey(const ValueKey('online-submit')))
        .dy;
    expect(
      tester.getTopLeft(find.byKey(const ValueKey('online-cost-guidance'))).dy,
      greaterThan(submit),
    );
    expect(
      tester
          .getTopLeft(
            find.text(
              'Unit price 5 · Quantity 1 · Effective payment 0 · Cost 5 · Excess -5',
            ),
          )
          .dy,
      lessThan(submit),
    );
    final close = find.byKey(const ValueKey('close-card-context'));
    await tester.ensureVisible(close);
    await tester.tap(close);
    await tester.pumpAndSettle();
    await chooseGameOption(tester, 'Card gestures');
    expect(find.text('Ponzi quota: used'), findsOneWidget);
    expect(find.text('Fate quota: available'), findsOneWidget);
    expect(find.textContaining('Opening:'), findsOneWidget);
    expect(find.textContaining('Underground privilege:'), findsOneWidget);
    expect(client.submissions, isEmpty);
  });

  testWidgets(
    'anonymous entry does not present expected bootstrap sign-out as an error',
    (tester) async {
      final client = AnonymousEntryClient();
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(client.errorCode, 'UNAUTHENTICATED');
      expect(find.byKey(const ValueKey('online-error')), findsNothing);
      expect(find.text('Take your seat'), findsOneWidget);
      expect(
        find.text('Continue as a guest, or sign in to your account.'),
        findsOneWidget,
      );
    },
  );

  testWidgets('failed explicit login keeps authentication error feedback', (
    tester,
  ) async {
    final client = AnonymousEntryClient();
    addTearDown(client.dispose);
    await showOnline(tester, client);
    await tester.tap(find.widgetWithText(TextButton, 'Login'));
    await tester.pumpAndSettle();
    expect(client.loginCalls, 1);
    expect(find.byKey(const ValueKey('online-error')), findsOneWidget);
    expect(find.text('Sign in or continue as a guest.'), findsOneWidget);
  });

  testWidgets(
    'a pending guest request blocks another session request until completion',
    (tester) async {
      final client = DeferredGuestClient();
      addTearDown(client.dispose);
      await showOnline(tester, client);
      final guest = find.widgetWithText(FilledButton, 'Continue as guest');
      final start = tester.widget<FilledButton>(guest).onPressed!;
      start();
      start();
      await tester.pump();
      expect(client.guestCalls, 1);
      expect(
        client.busy,
        isFalse,
        reason: 'Session transport does not expose command busy state.',
      );
      expect(tester.widget<FilledButton>(guest).onPressed, isNull);
      client.completion.complete();
      await tester.pumpAndSettle();
      expect(tester.widget<FilledButton>(guest).onPressed, isNotNull);
    },
  );

  testWidgets(
    'room fields and entered values survive width transitions and large RTL text',
    (tester) async {
      final client = ViewClient()..account = {'id': 'owner'};
      addTearDown(client.dispose);
      await showOnline(tester, client);
      await tester.enterText(
        find.widgetWithText(TextField, 'Room code'),
        'private-room',
      );
      await tester.enterText(
        find.widgetWithText(TextField, 'Invitation'),
        'private-invitation',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Games per match'),
        '7',
      );
      for (final width in [
        320.0,
        390.0,
        599.0,
        600.0,
        768.0,
        1049.0,
        1050.0,
        1440.0,
      ]) {
        await tester.binding.setSurfaceSize(Size(width, 900));
        await tester.pumpAndSettle();
        expect(
          tester.takeException(),
          isNull,
          reason: 'room setup width $width',
        );
        expect(
          tester
              .widget<TextField>(find.widgetWithText(TextField, 'Room code'))
              .controller!
              .text,
          'private-room',
        );
        expect(
          tester
              .widget<TextField>(find.widgetWithText(TextField, 'Invitation'))
              .controller!
              .text,
          'private-invitation',
        );
        expect(find.text('7'), findsOneWidget);
      }
      final settings = tester.widget<PopupMenuButton<String>>(
        find.byWidgetPredicate(
          (widget) =>
              widget is PopupMenuButton<String> &&
              widget.tooltip != 'Game options',
        ),
      );
      settings.onSelected!('text');
      settings.onSelected!('rtl');
      await tester.binding.setSurfaceSize(const Size(320, 800));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Games per match'),
        'seven',
      );
      await tester.binding.setSurfaceSize(const Size(1050, 900));
      await tester.pumpAndSettle();
      expect(find.text('seven'), findsOneWidget);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'entry prevents duplicate authentication while a request is pending',
    (tester) async {
      final client = ViewClient()..busy = true;
      addTearDown(client.dispose);
      await tester.binding.setSurfaceSize(const Size(1280, 900));
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: OnlineScreen(client: client),
        ),
      );
      await tester.pump();
      expect(
        tester
            .widget<FilledButton>(
              find.widgetWithText(FilledButton, 'Continue as guest'),
            )
            .onPressed,
        isNull,
      );
      for (final button in tester.widgetList<TextButton>(
        find.byType(TextButton),
      )) {
        expect(button.onPressed, isNull);
      }
    },
  );

  Map<String, dynamic> physical(
    String id,
    int rank,
    String suit, {
    String zone = 'hand',
    String? allocation,
    int deck = 1,
    int owner = 1,
  }) => {
    'card': {'id': id, 'rank': rank, 'suit': suit, 'deck': deck},
    'controller': owner,
    'zone': zone,
    if (allocation != null) 'allocation': allocation,
  };
  Future<void> closeContext(WidgetTester t) async {
    final close = find.byKey(const ValueKey('close-card-context'));
    if (close.evaluate().isNotEmpty) {
      await t.tap(close);
      await t.pumpAndSettle();
    }
  }

  Future<void> tapPhysical(
    WidgetTester t,
    String id, {
    bool twice = false,
  }) async {
    await closeContext(t);
    final hand = find.byKey(ValueKey('adaptive-select-$id'));
    final f = hand.evaluate().isNotEmpty
        ? hand
        : find.byKey(ValueKey('public-preview-$id'));
    await t.ensureVisible(f);
    await t.pumpAndSettle();
    final rect = t.getRect(f);
    final point = Offset(rect.left + 24, rect.center.dy);
    await t.tapAt(point);
    if (twice) {
      await t.pump(const Duration(milliseconds: 80));
      await t.tapAt(point);
    }
    await t.pumpAndSettle(const Duration(milliseconds: 350));
  }

  Future<void> chooseIntent(WidgetTester t, String key) async {
    final choice = find.byKey(ValueKey('card-intent-$key'));
    await t.ensureVisible(choice);
    await t.tap(choice);
    await t.pumpAndSettle();
  }

  testWidgets(
    'every command has a contextual review without a universal action panel',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
      final prepared = <String>{};
      for (final type in commandFields.keys.where(
        (type) => type != 'finish-settlement',
      )) {
        chooseAction(tester, type);
        await tester.pumpAndSettle();
        expect(
          find.byKey(const ValueKey('online-submit')),
          findsOneWidget,
          reason: type,
        );
        expect(
          find.byKey(const ValueKey('contextual-popup')),
          findsOneWidget,
          reason: type,
        );
        expect(
          find.byKey(
            const PageStorageKey('online-action-panel'),
            skipOffstage: false,
          ),
          findsNothing,
        );
        expect(
          find.byKey(const ValueKey('online-action-status')),
          findsNothing,
        );
        expect(client.submissions, isEmpty, reason: type);
        prepared.add(type);
      }
      expect(
        prepared,
        unorderedEquals(
          commandFields.keys.where((type) => type != 'finish-settlement'),
        ),
      );
      chooseAction(tester, 'offer');
      await tester.pumpAndSettle();
      expect(find.text('Private card trade'), findsOneWidget);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets('invalid numeric draft explains choices and preserves input', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish(
        {},
        online: {
          'rules_context': {'purchase_unit_price': 5},
        },
      );
    addTearDown(client.dispose);
    await showOnline(tester, client);
    chooseAction(tester, 'purchase');
    await tester.pumpAndSettle();
    await tester.enterText(find.widgetWithText(TextField, 'Quantity'), 'two');
    expect(
      tester
          .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
          .onPressed,
      isNull,
    );
    expect(find.textContaining('Unit price'), findsOneWidget);
    expect(
      tester
          .widget<TextField>(find.widgetWithText(TextField, 'Quantity'))
          .controller!
          .text,
      'two',
    );
    expect(client.submissions, isEmpty);
  });

  testWidgets('selected abilities explain committed and resolution-only costs', (
    tester,
  ) async {
    final client = ViewClient()..publish({});
    addTearDown(client.dispose);
    await showOnline(tester, client);
    chooseAction(tester, 'ponzi');
    await tester.pumpAndSettle();
    await showActionGuidance(tester);
    expect(
      find.textContaining(
        'Return all four jacks after resolution, even when nothing is stolen.',
      ),
      findsOneWidget,
    );
    expect(
      find.textContaining(
        'Cancellation keeps the jacks but spends the turn-cycle allowance.',
      ),
      findsOneWidget,
    );
    chooseAction(tester, 'attack');
    await tester.pumpAndSettle();
    await showActionGuidance(tester);
    expect(
      find.textContaining(
        'Physical Clubs become used at legal declaration, including if the attack is canceled.',
      ),
      findsOneWidget,
    );
    expect(
      find.textContaining(
        'Exchange costs apply only after successful resolution.',
      ),
      findsOneWidget,
    );
    chooseAction(tester, 'infiltrator-attack');
    await tester.pumpAndSettle();
    await showActionGuidance(tester);
    expect(
      find.textContaining(
        'Return the Infiltrator jack at legal declaration; cancellation does not return it or its turn-cycle allowance.',
      ),
      findsOneWidget,
    );
    chooseAction(tester, 'purchase');
    await tester.pumpAndSettle();
    await showActionGuidance(tester);
    expect(
      find.textContaining(
        'Cancellation keeps every payment card and draws nothing.',
      ),
      findsOneWidget,
    );
    expect(
      client.submissions,
      isEmpty,
      reason: 'Guidance is presentation, not an activation.',
    );
  });
  testWidgets(
    'own exposed Clubs show current physical usage and restrictions',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 5,
          'cards': [
            for (final id in ['used', 'unused'])
              {
                'card': {
                  'id': id,
                  'rank': 5,
                  'deck': id == 'used' ? 1 : 2,
                  'suit': 'clubs',
                },
                'controller': 1,
                'zone': 'series',
                'used_turn': id == 'used' ? 5 : 4,
                'available_from_round': 6,
              },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client);
      await openPublicSeat(tester, 'used');
      final used = tester
          .widget<CgmsCard>(find.byKey(const ValueKey('adaptive-select-used')))
          .card;
      expect(used.status, contains('Used this turn'));
      expect(used.status, contains('Available from round 6'));
      final unused = tester
          .widget<CgmsCard>(
            find.byKey(const ValueKey('adaptive-select-unused')),
          )
          .card;
      expect(unused.status, isNot(contains('Used this turn')));
      final sameCards = client.snapshot!.board['cards'];
      client.publish({'turn': 6, 'cards': sameCards});
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<CgmsCard>(
              find.byKey(const ValueKey('adaptive-select-used')),
            )
            .card
            .status,
        isNot(contains('Used this turn')),
      );
    },
  );
  testWidgets('clearing a draft resets the displayed suit and submitted suit', (
    t,
  ) async {
    final c = ViewClient()
      ..publish({
        'cards': [physical('h', 5, 'hearts'), physical('d', 7, 'diamonds')],
      });
    addTearDown(c.dispose);
    await showOnline(t, c, expand: false);
    await tapPhysical(t, 'd', twice: true);
    await chooseIntent(t, 'open-series-diamonds');
    expect(find.text('Series · Diamonds'), findsOneWidget);
    await t.ensureVisible(find.text('Clear selection'));
    await t.tap(find.text('Clear selection'));
    await t.pumpAndSettle();
    expect(find.byKey(const ValueKey('online-submit')), findsNothing);
    await tapPhysical(t, 'h', twice: true);
    await chooseIntent(t, 'open-series-hearts');
    expect(find.text('Series · Hearts'), findsOneWidget);
    await t.ensureVisible(find.byKey(const ValueKey('online-submit')));
    await t.tap(find.byKey(const ValueKey('online-submit')));
    await t.pumpAndSettle();
    expect(c.submissions.single['payload']['suit'], 'hearts');
    expect(c.submissions.single['payload']['selection'], ['h']);
  });
  testWidgets(
    'switching action routes card choices into its visible selection',
    (t) async {
      final c = ViewClient()
        ..publish({
          'cards': [
            physical('payment-spade', 5, 'spades'),
            physical('heart', 11, 'hearts'),
            physical('support', 5, 'hearts', zone: 'series'),
          ],
        });
      addTearDown(c.dispose);
      await showOnline(t, c, expand: false);
      await tapPhysical(t, 'payment-spade', twice: true);
      await chooseIntent(t, 'purchase');
      expect(find.text('Payment · 5 of Spades, copy 1'), findsOneWidget);
      await tapPhysical(t, 'heart', twice: true);
      await chooseIntent(t, 'attach');
      expect(
        t
            .widget<CgmsCard>(
              find.byKey(const ValueKey('adaptive-select-payment-spade')),
            )
            .selected,
        isFalse,
      );
      expect(find.text('Selection · J of Hearts, copy 1'), findsOneWidget);
      await t.tap(find.text('Choose a supported series'));
      await t.pumpAndSettle();
      await t.tap(find.byKey(const ValueKey('formation-target-1-hearts')));
      await t.pumpAndSettle(const Duration(milliseconds: 350));
      await t.ensureVisible(find.byKey(const ValueKey('online-submit')));
      await t.tap(find.byKey(const ValueKey('online-submit')));
      await t.pumpAndSettle();
      expect(c.submissions.single['payload']['selection'], ['heart']);
      await tapPhysical(t, 'payment-spade', twice: true);
      await chooseIntent(t, 'purchase');
      await t.ensureVisible(find.text('Clear selection'));
      await t.tap(find.text('Clear selection'));
      await t.pumpAndSettle();
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
      expect(
        t
            .widget<CgmsCard>(
              find.byKey(const ValueKey('adaptive-select-payment-spade')),
            )
            .selected,
        isFalse,
      );
      await tapPhysical(t, 'payment-spade', twice: true);
      await chooseIntent(t, 'purchase');
      expect(find.text('Payment · 5 of Spades, copy 1'), findsOneWidget);
    },
  );
  testWidgets('clipboard denial explains recovery without a connection error', (
    tester,
  ) async {
    final client = ViewClient()..publishRoom(owner: true, seats: 1);
    addTearDown(client.dispose);
    tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      SystemChannels.platform,
      (call) async {
        if (call.method == 'Clipboard.setData')
          throw PlatformException(code: 'NotAllowedError');
        return null;
      },
    );
    addTearDown(
      () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        null,
      ),
    );
    await showOnline(tester, client);
    await tester.tap(find.byTooltip('Copy Room code'));
    await tester.pumpAndSettle();
    expect(
      find.text(
        'Clipboard access is unavailable. The code remains visible; check your browser permissions and try Copy again.',
      ),
      findsOneWidget,
    );
    expect(find.text('Room code: room'), findsOneWidget);
    expect(find.byKey(const ValueKey('online-error')), findsNothing);
  });
  testWidgets(
    'leaving substitution returns physical selection to the new formation',
    (t) async {
      final c = ViewClient()..publish({});
      addTearDown(c.dispose);
      await showOnline(t, c);
      await preparePeople(t);
      expect(find.text('Replacement rank'), findsOneWidget);
      await t.ensureVisible(find.text('Clear selection'));
      await t.tap(find.text('Clear selection'));
      await t.pumpAndSettle();
      chooseAction(t, 'open-formation');
      await t.pumpAndSettle();
      expect(find.text('Replacement rank'), findsNothing);
      final own = t
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .data
          .hand
          .firstWhere((card) => card.rank == 'K');
      await tapPhysical(t, own.id);
      await tapPhysical(t, own.id, twice: true);
      expect(find.textContaining('Combo members · K'), findsOneWidget);
      expect(c.submissions, isEmpty);
    },
  );
  testWidgets('room and invitation codes are labelled readable and copyable', (
    tester,
  ) async {
    final client = ViewClient()..publishRoom(owner: true, seats: 1);
    addTearDown(client.dispose);
    final semantics = tester.ensureSemantics();
    final copied = <String>[];
    tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
      SystemChannels.platform,
      (call) async {
        if (call.method == 'Clipboard.setData')
          copied.add((call.arguments as Map)['text'] as String);
        return null;
      },
    );
    addTearDown(
      () => tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
        SystemChannels.platform,
        null,
      ),
    );
    await showOnline(tester, client);
    final roomCode = find.byKey(const ValueKey('share-Room code'));
    expect(roomCode, findsOneWidget);
    expect(find.bySemanticsLabel('Room code: room'), findsOneWidget);
    await tester.tap(find.byTooltip('Copy Room code'));
    await tester.pumpAndSettle();
    expect(copied, ['room']);
    await tester.tap(find.text('Create invitation'));
    await tester.pumpAndSettle();
    expect(
      find.bySemanticsLabel('Invitation: private-invitation'),
      findsOneWidget,
    );
    await tester.tap(find.byTooltip('Copy Invitation'));
    await tester.pumpAndSettle();
    expect(copied, ['room', 'private-invitation']);
    semantics.dispose();
    expect(tester.takeException(), isNull);
  });
  testWidgets('room owner waits for capacity and members wait for the owner', (
    tester,
  ) async {
    final client = ViewClient()..publishRoom(owner: true, seats: 1);
    addTearDown(client.dispose);
    await showOnline(tester, client);
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Start match'),
          )
          .onPressed,
      isNull,
    );
    expect(find.text('Waiting for 3 more players.'), findsOneWidget);
    client.publishRoom(owner: true, seats: 4);
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Start match'),
          )
          .onPressed,
      isNotNull,
    );
    client.publishRoom(owner: false, seats: 4);
    await tester.pumpAndSettle();
    expect(find.text('Create invitation'), findsNothing);
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Start match'),
          )
          .onPressed,
      isNull,
    );
    expect(
      find.text('The room owner starts the match when everyone has joined.'),
      findsOneWidget,
    );
    client.publishRoom(owner: false, seats: 4, match: 'started');
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Enter match'),
          )
          .onPressed,
      isNotNull,
    );
  });
  testWidgets('uncertain start exposes original room recovery in the lobby', (
    tester,
  ) async {
    final client = ViewClient()..publishRoom(owner: true, seats: 4);
    client.pendingRoom = true;
    addTearDown(client.dispose);
    await showOnline(tester, client);
    expect(find.text('Recover room operation'), findsOneWidget);
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Start match'),
          )
          .onPressed,
      isNull,
    );
  });
  testWidgets(
    'pending response disables ordinary submission with timing guidance',
    (tester) async {
      final client = ViewClient()
        ..publish({'window_id': 'w', 'required_actor': 2});
      addTearDown(client.dispose);
      await showOnline(tester, client, preparedAction: 'open-series');
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(
        find.text(
          'Wait for the current response or decision before an ordinary action.',
        ),
        findsOneWidget,
      );
      chooseAction(tester, 'pass');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(find.text('Seat 2 must respond first.'), findsOneWidget);
      client.publish({'window_id': 'w', 'required_actor': 1});
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<OutlinedButton>(
              find.widgetWithText(OutlinedButton, 'Pass response'),
            )
            .onPressed,
        isNotNull,
      );
      expect(client.submissions, isEmpty);
    },
  );
  testWidgets(
    'effect input never offers pass and enables only scoped decision submission',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'window_id': 'w',
          'required_actor': 1,
          'decision_id': 'd',
          'decision_kind': 'justice',
          'choices': [<String>[]],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client);
      chooseAction(tester, 'pass');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(
        find.text(
          'Complete the current effect decision; this is not a response window.',
        ),
        findsOneWidget,
      );
      expect(find.text('Pass response'), findsNothing);
      await tester.tap(find.byKey(const ValueKey('close-card-context')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Pending decision'));
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('effect-choice-0')));
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNotNull,
      );
    },
  );
  testWidgets(
    'out-of-turn ordinary action waits while Ponzi stays available for authority validation',
    (t) async {
      final cards = [
        for (final suit in ['clubs', 'spades'])
          for (final copy in [1, 2])
            physical(
              '$suit-$copy',
              11,
              suit,
              deck: copy,
              zone: 'formation',
              allocation: 'ponzi',
            ),
      ];
      final c = ViewClient()
        ..publish({
          'active': 2,
          'players': [
            {
              'seat': 1,
              'history': ['clubs', 'hearts'],
              'hand_count': 0,
              'ace_count': 0,
            },
            {
              'seat': 2,
              'history': ['clubs', 'hearts'],
              'hand_count': 0,
              'ace_count': 0,
            },
          ],
          'cards': cards,
          'formations': [
            {
              'id': 'ponzi',
              'controller': 1,
              'spec': {
                'kind': 'ponzi',
                'cards': cards.map((p) => p['card']['id']).toList(),
              },
            },
          ],
        });
      addTearDown(c.dispose);
      await showOnline(t, c, preparedAction: 'open-series');
      expect(
        t
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(
        find.text('This action is available on your turn.'),
        findsOneWidget,
      );
      await tapPhysical(t, 'clubs-1', twice: true);
      await chooseIntent(t, 'ponzi');
      await t.tap(find.widgetWithText(OutlinedButton, 'Seat 2'));
      await t.pumpAndSettle();
      expect(
        t
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNotNull,
      );
      await t.ensureVisible(find.text('Clear selection'));
      await t.tap(find.text('Clear selection'));
      await t.pumpAndSettle();
      await tapPhysical(t, 'clubs-1', twice: true);
      await chooseIntent(t, 'offer-loan');
      expect(find.text('Loan this intact formation'), findsOneWidget);
      await t.tap(find.widgetWithText(OutlinedButton, 'Seat 2'));
      await t.pumpAndSettle();
      expect(
        t
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNotNull,
        reason:
            'Intact loan is allowed on borrower turn after exact recipient choice.',
      );
      expect(c.submissions, isEmpty);
    },
  );
  testWidgets(
    'ordinary victory may be declared out of turn only at idle boundary',
    (tester) async {
      final client = ViewClient()..publish({'active': 2});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      chooseAction(tester, 'declare-ordinary');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNotNull,
        reason:
            'An accepted transfer may establish victory on another seat’s turn.',
      );
      client.publish({
        'active': 2,
        'window_id': 'pending',
        'required_actor': 3,
      });
      await tester.pumpAndSettle();
      expect(
        commandTiming(
          'declare-ordinary',
          client.snapshot!.board,
          client.snapshot!.online,
        ),
        CommandTiming.pending,
      );
      client.publish({'active': 2}, online: {'automatic_pending': true});
      await tester.pumpAndSettle();
      expect(
        commandTiming(
          'declare-ordinary',
          client.snapshot!.board,
          client.snapshot!.online,
        ),
        CommandTiming.automatic,
      );
    },
  );
  testWidgets('short rotated recovery view remains scrollable at large text', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish({'window_id': 'w', 'required_actor': 2});
    client.pendingUnknown = {'command_id': 'unknown'};
    client.connected = false;
    client.errorCode = 'OUTCOME_UNKNOWN';
    addTearDown(client.dispose);
    await showOnline(tester, client);
    await tester.binding.setSurfaceSize(const Size(844, 390));
    tester
        .widget<PopupMenuButton<String>>(
          find.byWidgetPredicate(
            (widget) =>
                widget is PopupMenuButton<String> &&
                widget.tooltip != 'Game options',
          ),
        )
        .onSelected!('text');
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.ensureVisible(find.text('Check operation'));
    expect(find.text('Check operation'), findsOneWidget);
  });
  testWidgets(
    'next game allowance wait has accessible economy recovery instead of match complete',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {'phase': 'settlement'},
          online: {'financial_phase': 'finalized'},
          admissionWaiting: true,
        );
      client.account = {'id': 'owner'};
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(find.text('Match complete'), findsNothing);
      expect(find.text('Seat 1’s turn'), findsNothing);
      expect(find.textContaining('next game is waiting'), findsOneWidget);
      await tester.tap(find.byTooltip('Allowance and dirt'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('contextual-popup')), findsOneWidget);
      expect(find.byType(EconomyPanel), findsOneWidget);
    },
  );
  testWidgets(
    'allowance dialog preserves immediate Coup at large text and short height',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'active': 2,
          'cards': [
            for (final suit in ['hearts', 'diamonds'])
              for (final copy in [1, 2])
                {
                  'card': {
                    'id': '$suit-$copy',
                    'rank': 12,
                    'suit': suit,
                    'deck': copy,
                  },
                  'controller': 1,
                  'zone': 'hand',
                  'available_from_round': 4,
                },
          ],
        });
      client.account = {'id': 'owner'};
      addTearDown(client.dispose);
      await showOnline(tester, client);
      tester
          .widget<PopupMenuButton<String>>(
            find.byWidgetPredicate(
              (widget) =>
                  widget is PopupMenuButton<String> &&
                  widget.tooltip != 'Game options',
            ),
          )
          .onSelected!('text');
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Allowance and dirt'));
      await tester.pumpAndSettle();
      for (final size in [
        const Size(390, 844),
        const Size(768, 1024),
        const Size(844, 390),
      ]) {
        await tester.binding.setSurfaceSize(size);
        await tester.pumpAndSettle();
        final queen = find
            .descendant(
              of: find.byKey(const ValueKey('contextual-popup')),
              matching: find.byWidgetPredicate(
                (w) => w is CgmsCard && w.card.rank == 'Q',
              ),
            )
            .first;
        await tester.ensureVisible(queen);
        expect(tester.getSize(queen).width, greaterThanOrEqualTo(48));
        expect(tester.getSize(queen).height, greaterThanOrEqualTo(48));
        expect(
          queen.hitTestable(),
          findsOneWidget,
          reason:
              '$size queen=${tester.getRect(queen)} popup=${tester.getRect(find.byKey(const ValueKey('contextual-popup')))}',
        );
        expect(tester.takeException(), isNull);
      }
      final queen = find
          .descendant(
            of: find.byKey(const ValueKey('contextual-popup')),
            matching: find.byWidgetPredicate(
              (w) => w is CgmsCard && w.card.rank == 'Q',
            ),
          )
          .first;
      await tester.tap(queen);
      await tester.pump(const Duration(milliseconds: 80));
      await tester.tap(queen);
      await tester.pumpAndSettle();
      expect(client.submissions, isEmpty);
      await tester.ensureVisible(find.byKey(const ValueKey('online-submit')));
      await tester.tap(find.byKey(const ValueKey('online-submit')));
      await tester.pumpAndSettle();
      expect(client.submissions.single['type'], 'coup');
    },
  );
  testWidgets(
    'pending action pauses accepting offers while decline remains nonblocking',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'window_id': 'other-action',
          'required_actor': 2,
          'proposals': [
            {
              'id': 'p',
              'from': 2,
              'revision': 1,
              'status': 'offered',
              'terms': {'to': 1, 'give': [], 'receive': [], 'loan': false},
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(
        tester
            .widget<TextButton>(find.widgetWithText(TextButton, 'Accept Offer'))
            .onPressed,
        isNull,
      );
      expect(
        tester
            .widget<TextButton>(
              find.widgetWithText(TextButton, 'Decline Offer'),
            )
            .onPressed,
        isNotNull,
      );
    },
  );
  testWidgets('online setup provides real session and room controls', (
    tester,
  ) async {
    final client = OnlineClient(ScreenTransport());
    addTearDown(client.dispose);
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: OnlineScreen(client: client),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('Continue as guest'), findsOneWidget);
    expect(find.text('Create room'), findsNothing);
    expect(find.textContaining('Simulate'), findsNothing);
  });
  testWidgets(
    'numeric draft cannot move into a new game and clears all edits',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      chooseAction(tester, 'purchase');
      await tester.pumpAndSettle();
      await tester.enterText(find.widgetWithText(TextField, 'Quantity'), '7');
      client.publish({'game_id': 'replacement'});
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(find.textContaining('game or decision changed'), findsOneWidget);
      await tester.ensureVisible(find.text('Clear selection'));
      await tester.tap(find.text('Clear selection'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
      chooseAction(tester, 'purchase');
      await tester.pumpAndSettle();
      expect(
        tester
            .widget<TextField>(find.widgetWithText(TextField, 'Quantity'))
            .controller!
            .text,
        '1',
      );
    },
  );
  testWidgets('same-game decision replacement invalidates an edited response', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish({
        'window_id': 'w',
        'decision_id': 'd',
        'decision_kind': 'justice',
        'required_actor': 1,
        'choices': [<String>[]],
      });
    addTearDown(client.dispose);
    await showOnline(tester, client);
    await tester.tap(find.byKey(const ValueKey('effect-choice-0')));
    await tester.pumpAndSettle();
    expect(
      tester
          .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
          .onPressed,
      isNotNull,
    );
    client.publish({
      'window_id': 'w',
      'decision_id': 'replacement',
      'decision_kind': 'justice',
      'required_actor': 1,
      'choices': [<String>[]],
    });
    await tester.pumpAndSettle();
    expect(find.byKey(const ValueKey('online-submit')), findsNothing);
    expect(find.byKey(const ValueKey('effect-choice-0')), findsOneWidget);
    expect(client.submissions, isEmpty);
  });
  testWidgets(
    'quota status uses authority boolean rather than interval marker',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {},
          online: {
            'rules_context': {
              'purchase_unit_price': 5,
              'quotas': {'ponzi': 4},
              'quota_used': {'ponzi': true, 'fate': false},
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client);
      await chooseGameOption(tester, 'Card gestures');
      expect(find.text('Ponzi quota: used'), findsOneWidget);
      expect(find.text('Fate quota: available'), findsOneWidget);
      expect(find.text('Ponzi uses: 4'), findsNothing);
    },
  );
  testWidgets(
    'response confirmation retains exact pending actor and frozen target',
    (t) async {
      final client = ViewClient()
        ..publish(
          {
            'active': 2,
            'window_id': 'attack-window',
            'required_actor': 1,
            'players': [
              for (final seat in [1, 2])
                {
                  'seat': seat,
                  'hand_count': 0,
                  'ace_count': seat == 1 ? 1 : 0,
                  'history': ['clubs'],
                },
            ],
            'cards': [
              physical('response-ace', 1, 'diamonds', zone: 'concealed-ace'),
              physical('frozen-heart', 9, 'hearts', zone: 'series'),
              physical('committed-club', 9, 'clubs', zone: 'series', owner: 2),
            ],
          },
          online: {
            'rules_context': {
              'purchase_unit_price': 5,
              'pending': {
                'type': 'attack',
                'actor': 2,
                'target_seat': 1,
                'response_types': ['confinement'],
              },
              'combat': {
                'actor': 2,
                'defender': 1,
                'attack_cards': ['committed-club'],
                'target_cards': ['frozen-heart'],
                'physical_attack': {'numerator': '9', 'denominator': '1'},
                'attack_modifier': 0,
                'attack': {'numerator': '9', 'denominator': '1'},
                'physical_defense': {'numerator': '9', 'denominator': '1'},
                'defense_modifier': 0,
                'defense': {'numerator': '9', 'denominator': '1'},
                'comparison': 'equal',
              },
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(t, client, expand: false);
      await tapPhysical(t, 'response-ace', twice: true);
      await chooseIntent(t, 'confinement');
      expect(find.text('Confirm Confinement'), findsOneWidget);
      expect(find.text('Seat 2 attacks seat 1'), findsOneWidget);
      expect(find.text('Frozen targets · 9 of Hearts, copy 1'), findsOneWidget);
      expect(find.text('Committed Clubs · 9 of Clubs, copy 1'), findsOneWidget);
      expect(find.textContaining('Decision owner: seat 1'), findsOneWidget);
      expect(find.text('Source card · A of Diamonds, copy 1'), findsOneWidget);
      expect(client.submissions, isEmpty);
      await t.sendKeyEvent(LogicalKeyboardKey.escape);
      await t.pumpAndSettle();
      expect(client.snapshot!.board['window_id'], 'attack-window');
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets('clearing or losing an Ace resets the visible selected cost', (
    t,
  ) async {
    final cards = [
      physical('ace', 1, 'hearts', zone: 'concealed-ace'),
      physical('c', 5, 'clubs', zone: 'series'),
      physical('d', 5, 'diamonds', zone: 'series'),
    ];
    final c = ViewClient()..publish({'cards': cards});
    addTearDown(c.dispose);
    await showOnline(t, c, expand: false);
    await tapPhysical(t, 'ace', twice: true);
    await chooseIntent(t, 'compensation');
    expect(find.text('Source card · A of Hearts, copy 1'), findsOneWidget);
    expect(find.text('Source cards · A of Hearts, copy 1'), findsNothing);
    expect(find.text('Modifier · A of Hearts, copy 1'), findsNothing);
    expect(
      t
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .cardRoles['ace'],
      TableCardRole.source,
    );
    await t.ensureVisible(find.text('Clear selection'));
    await t.tap(find.text('Clear selection'));
    await t.pumpAndSettle();
    expect(
      t
          .widget<CgmsCard>(find.byKey(const ValueKey('adaptive-select-ace')))
          .selected,
      isFalse,
    );
    await tapPhysical(t, 'ace', twice: true);
    await chooseIntent(t, 'compensation');
    c.publish({'cards': cards.where((p) => p['card']['id'] != 'ace').toList()});
    await t.pumpAndSettle();
    expect(t.takeException(), isNull);
    expect(find.text('Source card · A of Hearts, copy 1'), findsNothing);
    expect(
      t
          .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
          .onPressed,
      isNull,
    );
    expect(c.submissions, isEmpty);
  });
  testWidgets('active match changes notify the host for reload-safe routing', (
    tester,
  ) async {
    final client = ViewClient();
    String? routed;
    addTearDown(client.dispose);
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: OnlineScreen(client: client, onMatchChanged: (id) => routed = id),
      ),
    );
    await tester.pumpAndSettle();
    client.matchId = 'match-route';
    client.publish({});
    await tester.pumpAndSettle();
    expect(routed, 'match-route');
  });
  testWidgets('active lobby changes notify the host without invitation data', (
    tester,
  ) async {
    final client = ViewClient();
    String? routed;
    addTearDown(client.dispose);
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: OnlineScreen(client: client, onRoomChanged: (id) => routed = id),
      ),
    );
    await tester.pumpAndSettle();
    client.publishRoom(owner: true, seats: 1);
    await tester.pumpAndSettle();
    expect(routed, 'room');
  });
  testWidgets('reload restores the authorized lobby after session recovery', (
    tester,
  ) async {
    final client = ViewClient();
    addTearDown(client.dispose);
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: OnlineScreen(client: client, roomId: 'room'),
      ),
    );
    await tester.pumpAndSettle();
    expect(client.restoredRooms, ['room']);
    expect(client.restoration, ['session', 'room:room']);
    expect(find.text('Room code: room'), findsOneWidget);
    expect(find.text('Waiting for 3 more players.'), findsOneWidget);
    expect(client.submissions, isEmpty);
  });
  testWidgets(
    'an explicit match route takes priority over a stale room route',
    (tester) async {
      final client = ViewClient();
      addTearDown(client.dispose);
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: OnlineScreen(client: client, roomId: 'room', matchId: 'match'),
        ),
      );
      await tester.pumpAndSettle();
      expect(client.restoration, ['session', 'match:match']);
      expect(client.restoredRooms, isEmpty);
      expect(client.submissions, isEmpty);
    },
  );
  testWidgets('signing in resumes a lobby route after an expired session', (
    tester,
  ) async {
    final client = ExpiredViewClient();
    addTearDown(client.dispose);
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: OnlineScreen(client: client, roomId: 'room'),
      ),
    );
    await tester.pumpAndSettle();
    expect(client.restoredRooms, isEmpty);
    expect(client.busy, isFalse);
    expect(find.byKey(const ValueKey('online-error')), findsOneWidget);
    await tester.ensureVisible(find.text('Login'));
    await tester.tap(find.text('Login'));
    await tester.pumpAndSettle();
    expect(client.restoredRooms, ['room']);
    expect(find.text('Room code: room'), findsOneWidget);
    expect(find.byKey(const ValueKey('online-error')), findsNothing);
  });
  testWidgets(
    'incoming promises and own receivables expose contextual actions',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {},
          financial: {
            'promises': [
              {
                'id': 'promise-1',
                'payer': 1,
                'recipient': 0,
                'status': 'offered',
                'mode': 'fixed',
                'award_id': 'award',
                'amount': {'numerator': '5', 'denominator': '2'},
              },
            ],
            'debts': [
              {
                'id': 'debt-1',
                'debtor': 1,
                'creditor': 0,
                'game_id': 'old',
                'remaining': {'numerator': '3', 'denominator': '2'},
              },
            ],
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(
        find.textContaining('Promise from seat 2 to seat 1'),
        findsOneWidget,
      );
      await tester.ensureVisible(find.text('Accept promise'));
      await tester.tap(find.text('Accept promise'));
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'promise-accept',
        'payload': {'promise_id': 'promise-1'},
      });
      await tester.ensureVisible(find.text('Forgive debt'));
      await tester.tap(find.text('Forgive debt'));
      await tester.pumpAndSettle();
      expect(find.widgetWithText(TextField, 'Debt Id'), findsNothing);
      expect(find.text('Confirm Forgive'), findsOneWidget);
      await tester.enterText(
        find.widgetWithText(TextField, 'Amount numerator'),
        '1',
      );
      await tester.ensureVisible(find.byKey(const ValueKey('online-submit')));
      await tester.tap(find.byKey(const ValueKey('online-submit')));
      await tester.pumpAndSettle();
      expect(client.submissions.last['payload']['debt_id'], 'debt-1');
    },
  );
  testWidgets('new proposal starts at valid first revision', (t) async {
    final c = ViewClient()
      ..publish({
        'cards': [physical('trade-card', 5, 'spades')],
        'players': [
          {
            'seat': 1,
            'history': ['clubs'],
            'hand_count': 0,
            'ace_count': 0,
          },
          {
            'seat': 2,
            'history': ['clubs', 'hearts'],
            'hand_count': 0,
            'ace_count': 0,
          },
        ],
      });
    addTearDown(c.dispose);
    await showOnline(t, c, expand: false);
    await tapPhysical(t, 'trade-card', twice: true);
    await chooseIntent(t, 'offer-trade');
    await t.tap(find.widgetWithText(OutlinedButton, 'Seat 2'));
    await t.pumpAndSettle();
    await t.ensureVisible(find.byKey(const ValueKey('online-submit')));
    await t.tap(find.byKey(const ValueKey('online-submit')));
    await t.pumpAndSettle();
    expect(c.submissions.single['type'], 'offer');
    expect(c.submissions.single['payload']['revision'], 1);
    expect(c.submissions.single['payload']['terms']['give'], ['trade-card']);
  });
  testWidgets(
    'settlement exposes finishing for explicit review before confirming',
    (tester) async {
      final client = ViewClient()..publish({'phase': 'settlement'});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
      await chooseGameOption(tester, 'Finish Settlement');
      expect(find.text('Confirm Finish Settlement'), findsOneWidget);
      expect(find.text('Confirm Nullify'), findsNothing);
    },
  );
  testWidgets(
    'finalized game waits for next game without claiming match complete',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {'phase': 'settlement'},
          online: {'financial_phase': 'finalized', 'completed_games': 1},
        );
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(
        find.text('Game settled. Waiting for the next game.'),
        findsOneWidget,
      );
      expect(find.text('Match complete'), findsNothing);
      expect(find.byKey(const ValueKey('online-submit')), findsNothing);
    },
  );
  testWidgets('finalized match shows standings without new game commands', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish(
        {'phase': 'settlement'},
        online: {
          'financial_phase': 'finalized',
          'completed_games': 3,
          'ranks': [1, 1, 3],
        },
      );
    addTearDown(client.dispose);
    await showOnline(tester, client);
    expect(find.text('Match complete'), findsOneWidget);
    expect(find.byKey(const ValueKey('online-submit')), findsNothing);
    expect(find.text('Match ranks by seat: [1, 1, 3]'), findsOneWidget);
  });
  testWidgets('Justice offers an exposed Queen as its physical cost', (
    t,
  ) async {
    final c = ViewClient()
      ..publish({
        'formations': [
          {
            'id': 'justice',
            'controller': 1,
            'spec': {
              'kind': 'justice',
              'cards': ['queen'],
            },
          },
        ],
        'cards': [
          physical(
            'queen',
            12,
            'hearts',
            zone: 'formation',
            allocation: 'justice',
          ),
        ],
      });
    addTearDown(c.dispose);
    await showOnline(t, c);
    chooseAction(t, 'justice');
    await t.pumpAndSettle();
    await t.ensureVisible(find.text('Consumed Queen'));
    await t.tap(find.text('Consumed Queen'));
    await t.pumpAndSettle();
    await tapPhysical(t, 'queen');
    await tapPhysical(t, 'queen', twice: true);
    expect(find.text('Consumed Queen · Q of Hearts, copy 1'), findsOneWidget);
    expect(c.submissions, isEmpty);
  });
  testWidgets(
    'private offer renders exact directions and physical terms before accept',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'proposals': [
            {
              'id': 'p',
              'from': 2,
              'revision': 2,
              'status': 'offered',
              'terms': {
                'to': 1,
                'give': ['queen'],
                'receive': [],
              },
            },
          ],
          'cards': [
            {
              'card': {'id': 'queen', 'rank': 12, 'deck': 2, 'suit': 'hearts'},
              'controller': 2,
              'zone': 'formation',
              'allocation': 'x',
              'available_from_round': 0,
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(find.text('You receive: Q of Hearts, copy 2'), findsOneWidget);
      expect(find.text('You give: none'), findsOneWidget);
      expect(find.text('Accept Offer'), findsOneWidget);
    },
  );

  testWidgets(
    'formation editor supplies substitution and protection controls',
    (t) async {
      final c = ViewClient()..publish({});
      addTearDown(c.dispose);
      await showOnline(t, c);
      await preparePeople(t);
      expect(find.text('Replacement rank'), findsOneWidget);
      expect(find.text('Replacement suit'), findsOneWidget);
      expect(find.textContaining('Protection ·'), findsOneWidget);
      expect(find.text('Choose a protection target'), findsOneWidget);
      expect(
        t
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(c.submissions, isEmpty);
    },
  );
  testWidgets('online respects inherited accessibility text scale', (
    tester,
  ) async {
    final client = ViewClient()..publish({});
    addTearDown(client.dispose);
    await tester.binding.setSurfaceSize(const Size(1280, 900));
    addTearDown(() => tester.binding.setSurfaceSize(null));
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: MediaQuery(
          data: const MediaQueryData(
            size: Size(1280, 900),
            textScaler: TextScaler.linear(2),
          ),
          child: OnlineScreen(client: client),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(
      MediaQuery.textScalerOf(
        tester.element(find.byType(CgmsAdaptiveTable)),
      ).scale(10),
      20,
    );
  });
  testWidgets('existing formation selection clears when authority removes it', (
    t,
  ) async {
    final c = ViewClient()
      ..publish({
        'formations': [
          {
            'id': 'known-code',
            'controller': 1,
            'spec': {
              'kind': 'code',
              'cards': ['q'],
            },
          },
        ],
        'cards': [
          physical(
            'q',
            12,
            'spades',
            zone: 'formation',
            allocation: 'known-code',
          ),
        ],
      });
    addTearDown(c.dispose);
    await showOnline(t, c, expand: false);
    await tapPhysical(t, 'q', twice: true);
    await chooseIntent(t, 'code');
    expect(find.text('Formation · Code'), findsOneWidget);
    c.publish({'formations': []});
    await t.pumpAndSettle();
    expect(t.takeException(), isNull);
    expect(find.text('Formation · Code'), findsNothing);
    expect(
      t
          .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
          .onPressed,
      isNull,
    );
    expect(c.submissions, isEmpty);
  });
  testWidgets('removed protected formation does not crash an edited draft', (
    t,
  ) async {
    final cards = [
      physical('qh1', 12, 'hearts'),
      physical('qh2', 12, 'hearts', deck: 2),
      physical(
        'code-q',
        12,
        'spades',
        zone: 'formation',
        allocation: 'known-code',
      ),
    ];
    final c = ViewClient()
      ..publish({
        'cards': cards,
        'formations': [
          {
            'id': 'known-code',
            'controller': 1,
            'spec': {
              'kind': 'code',
              'cards': ['code-q'],
            },
          },
        ],
      });
    addTearDown(c.dispose);
    await showOnline(t, c, expand: false);
    await tapPhysical(t, 'qh1');
    await tapPhysical(t, 'qh2', twice: true);
    await chooseIntent(t, 'open-formation-great-people');
    await t.ensureVisible(find.text('Choose a protection target'));
    await t.tap(find.text('Choose a protection target'));
    await t.pumpAndSettle();
    await tapPhysical(t, 'code-q');
    await tapPhysical(t, 'qh1', twice: true);
    expect(find.textContaining('Protection ·'), findsOneWidget);
    c.publish({'formations': [], 'cards': cards});
    await t.pumpAndSettle();
    expect(t.takeException(), isNull);
    expect(find.text('Protection · Choose your public target'), findsOneWidget);
    expect(
      t
          .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
          .onPressed,
      isNull,
    );
    expect(c.submissions, isEmpty);
  });
  testWidgets(
    'authority calculations custody draw order and corrections stay distinct',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {
            'effects': [
              {
                'kind': 'compensation',
                'source': 1,
                'target': 2,
                'custodian': 3,
                'expiry': 'closure',
              },
            ],
          },
          online: {
            'own_match_score': {'numerator': '4', 'denominator': '1'},
            'rules_context': {
              'purchase_unit_price': 8,
              'inflation': true,
              'combat': {
                'physical_attack': {'numerator': '10', 'denominator': '1'},
                'attack_modifier': 3,
                'attack': {'numerator': '13', 'denominator': '1'},
                'physical_defense': {'numerator': '20', 'denominator': '2'},
                'defense_modifier': 0,
                'defense': {'numerator': '10', 'denominator': '1'},
                'comparison': 'greater',
              },
            },
            'draw_queue': [
              {'seat': 2, 'remaining': 2},
              {'seat': 3, 'remaining': 1},
            ],
            'corrections': [
              {
                'amount': {'numerator': '1', 'denominator': '3'},
                'game_id': 'origin-game',
              },
            ],
            'ending': 'coup',
            'declarer': 2,
          },
          financial: {
            'cash': {'numerator': '10', 'denominator': '1'},
            'score': {'numerator': '3', 'denominator': '1'},
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client);
      expect(find.text('Your match score: 4'), findsOneWidget);

      expect(
        find.text('Compensation · owner 1 · target 2 · custodian 3 · closure'),
        findsOneWidget,
      );
      expect(
        find.text('Compensation draw · seat 2 · 2 remaining'),
        findsOneWidget,
      );
      expect(
        find.text('Compensation draw · seat 3 · 1 remaining'),
        findsOneWidget,
      );
      expect(
        find.text('Nonspendable match correction · 1/3 · origin origin-game'),
        findsOneWidget,
      );
      expect(find.text('Ending: coup · Declarer: seat 2'), findsOneWidget);
      expect(
        find.descendant(
          of: find.byKey(const ValueKey('contextual-popup')),
          matching: find.textContaining('Score 3 · Cash 10 · Debt 0'),
        ),
        findsOneWidget,
      );
      chooseAction(tester, 'purchase');
      await tester.pumpAndSettle();
      expect(
        find.text('Attack 10 + 3 = 13 · Defense 10 + 0 = 10 · Greater'),
        findsOneWidget,
      );
      expect(
        find.text(
          'Unit price 8 · Quantity 1 · Effective payment 0 · Cost 8 · Excess -8',
        ),
        findsOneWidget,
      );
    },
  );
  testWidgets(
    'primary online controls retain 48 pixel targets on desktop platforms',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.linux;
      final semantics = tester.ensureSemantics();
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      try {
        await showOnline(tester, client, preparedAction: 'open-series');
        for (final size in [
          const Size(390, 844),
          const Size(768, 1024),
          const Size(1440, 900),
        ]) {
          await tester.binding.setSurfaceSize(size);
          await tester.pumpAndSettle();
          await tester.ensureVisible(
            find.byKey(const ValueKey('contextual-popup'), skipOffstage: false),
          );
          await tester.pumpAndSettle();
          expect(find.byType(NavigationRail), findsNothing);
          final submit = find.byKey(const ValueKey('online-submit'));
          await tester.ensureVisible(submit);
          await tester.pumpAndSettle();
          expect(
            tester.getSemantics(submit).rect.height,
            greaterThanOrEqualTo(48),
            reason: 'submit at $size',
          );
          expect(
            tester
                .getSemantics(
                  find.byKey(
                    const ValueKey('contextual-popup'),
                    skipOffstage: false,
                  ),
                )
                .rect
                .height,
            greaterThanOrEqualTo(48),
            reason: 'navigation at $size',
          );
        }
      } finally {
        semantics.dispose();
        debugDefaultTargetPlatformOverride = null;
      }
    },
  );
  testWidgets(
    'every typed action form remains usable at phone tablet desktop sizes',
    (tester) async {
      final client = ViewClient()..publish({});
      addTearDown(client.dispose);
      await showOnline(tester, client);
      for (final size in [
        const Size(360, 800),
        const Size(600, 900),
        const Size(1280, 600),
      ]) {
        await tester.binding.setSurfaceSize(size);
        await tester.pumpAndSettle();
        var current = 'open-series';
        for (final action in commandFields.keys.where(
          (v) => v != 'finish-settlement',
        )) {
          chooseAction(tester, action);
          await tester.pump();
          expect(
            tester.takeException(),
            isNull,
            reason: '$size $action from $current',
          );
          current = action;
        }
      }
    },
  );
}

class DepartureClient extends ViewClient {
  final completion = Completer<void>();
  @override
  Future<void> submit(String type, Map<String, dynamic> payload) async {
    busy = true;
    notifyListeners();
    await super.submit(type, payload);
    await completion.future;
    busy = false;
    publish({}, online: {'round_closed': true, 'departure_pending': false});
  }
}

class ViewClient extends OnlineClient {
  ViewClient() : super(ScreenTransport());
  final submissions = <Map<String, dynamic>>[];
  final restoredRooms = <String>[];
  final restoration = <String>[];
  bool pendingRoom = false;
  Map<String, dynamic>? pendingUnknown;
  @override
  bool get hasPendingRoomOperation => pendingRoom;
  @override
  Map<String, dynamic>? get unknownOperation => pendingUnknown;
  void publishRoom({required bool owner, required int seats, String? match}) {
    account = {'id': owner ? 'owner' : 'member'};
    room = {
      'id': 'room',
      'owner_id': 'owner',
      'capacity': 4,
      'games': 3,
      'members': [
        for (var i = 1; i <= seats; i++) {'seat': i},
      ],
      if (match != null) 'match_id': match,
    };
    notifyListeners();
  }

  @override
  Future<void> restoreSession() async {
    restoration.add('session');
  }

  @override
  Future<void> watchMatch(String id) async {
    restoration.add('match:$id');
    matchId = id;
    publish({});
  }

  @override
  Future<void> refreshRoom(String id) async {
    restoration.add('room:$id');
    restoredRooms.add(id);
    publishRoom(owner: true, seats: 1);
  }

  @override
  Future<String> invite() async => 'private-invitation';
  @override
  Future<void> submit(String type, Map<String, dynamic> payload) async {
    submissions.add({'type': type, 'payload': payload});
  }

  void publish(
    Map<String, dynamic> board, {
    Map<String, dynamic> online = const {},
    Map<String, dynamic> financial = const {},
    bool admissionWaiting = false,
  }) {
    snapshot = AuthoritativeSnapshot.fromJson({
      'version': 1,
      'cursor': 1,
      'admission_waiting': admissionWaiting,
      'projection': {
        'board': {
          'game_id': 'g',
          'seat': 1,
          'phase': 'playing',
          'round': 4,
          'active': 1,
          'players': [
            {
              'seat': 1,
              'hand_count': 0,
              'ace_count': 0,
              'confined': false,
              'departed': false,
              'history': ['clubs'],
            },
          ],
          'cards': [],
          ...board,
        },
        'online': {
          'games_per_match': 3,
          'alliances': {'1': 0},
          ...online,
        },
        'debts': [],
        ...financial,
      },
    });
    connected = true;
    notifyListeners();
  }
}

class ExpiredViewClient extends ViewClient {
  bool sessionAvailable = false;
  @override
  Future<void> restoreSession() async {
    if (!sessionAvailable) {
      errorCode = 'UNAUTHENTICATED';
      throw StateError('UNAUTHENTICATED');
    }
    account = {'id': 'owner'};
    errorCode = null;
    await super.restoreSession();
  }

  @override
  Future<void> authenticate(
    String action,
    String username,
    String password,
  ) async {
    sessionAvailable = true;
    await restoreSession();
  }
}

class DeferredGuestClient extends ViewClient {
  final completion = Completer<void>();
  int guestCalls = 0;
  @override
  Future<void> guest() async {
    guestCalls++;
    await completion.future;
  }
}

class AnonymousEntryClient extends ViewClient {
  int loginCalls = 0;
  @override
  Future<void> restoreSession() async {
    errorCode = 'UNAUTHENTICATED';
    throw StateError('UNAUTHENTICATED');
  }

  @override
  Future<void> authenticate(
    String action,
    String username,
    String password,
  ) async {
    loginCalls++;
    errorCode = 'UNAUTHENTICATED';
    throw StateError('UNAUTHENTICATED');
  }
}

Future<void> showOnline(
  WidgetTester tester,
  OnlineClient client, {
  bool expand = true,
  ThemeData? theme,
  String? preparedAction,
}) async {
  await tester.binding.setSurfaceSize(const Size(1280, 900));
  addTearDown(() => tester.binding.setSurfaceSize(null));
  await tester.pumpWidget(
    MaterialApp(
      theme: theme ?? CgmsTheme.dark(),
      localizationsDelegates: CgmsLocalizations.localizationsDelegates,
      supportedLocales: CgmsLocalizations.supportedLocales,
      home: OnlineScreen(client: client),
    ),
  );
  await tester.pumpAndSettle();
  if (preparedAction != null) {
    chooseAction(tester, preparedAction);
    await tester.pumpAndSettle();
  } else if (expand &&
      client.snapshot != null &&
      find.byKey(const ValueKey('contextual-popup')).evaluate().isEmpty) {
    final options = tester.widget<PopupMenuButton<String>>(
      find.byWidgetPredicate(
        (w) => w is PopupMenuButton<String> && w.tooltip == 'Game options',
      ),
    );
    options.onSelected!('records');
    await tester.pumpAndSettle();
  }
}

Finder utilityBar() => find.byKey(const ValueKey('online-utilities'));

Future<void> showActionGuidance(WidgetTester tester) async {
  final help = find.byWidgetPredicate(
    (w) =>
        w is ExpansionTile &&
        w.title is Text &&
        (w.title as Text).data == 'Costs and timing',
  );
  if (help.evaluate().isEmpty) {
    chooseAction(tester, 'purchase');
    await tester.pumpAndSettle();
  }
  if (find.byKey(const ValueKey('online-cost-guidance')).evaluate().isEmpty) {
    final title = find
        .descendant(of: help, matching: find.text('Costs and timing'))
        .first;
    await tester.ensureVisible(title);
    await tester.tap(title);
    await tester.pumpAndSettle();
  }
}

Future<void> openPublicSeat(WidgetTester tester, String cardId) async {
  final close = find.byKey(const ValueKey('close-card-context'));
  if (close.evaluate().isNotEmpty) {
    await tester.tap(close);
    await tester.pumpAndSettle();
  }
  final board = tester.widget<CgmsAdaptiveTable>(
    find.byType(CgmsAdaptiveTable),
  );
  final seat = board.data.seats.firstWhere(
    (seat) =>
        seat.exposed.any((card) => card.id == cardId) ||
        seat.combos.any(
          (combo) => combo.cards.any((card) => card.id == cardId),
        ),
  );
  await tester.tap(
    find.descendant(
      of: find.byKey(ValueKey('adaptive-seat-${seat.id}')),
      matching: find.byTooltip('Inspect public cards'),
    ),
  );
  await tester.pumpAndSettle();
}

Future<void> activatePrivateCard(WidgetTester tester, String cardId) async {
  final card = find.byKey(ValueKey('adaptive-select-$cardId'));
  await tester.ensureVisible(card);
  await tester.tap(card);
  await tester.pump(const Duration(milliseconds: 80));
  await tester.tap(card);
  await tester.pumpAndSettle(const Duration(milliseconds: 350));
}

Future<void> preparePeople(WidgetTester tester) async {
  final client =
      tester.widget<OnlineScreen>(find.byType(OnlineScreen)).client
          as ViewClient;
  final snapshot = client.snapshot!;
  final close = find.byKey(const ValueKey('close-card-context'));
  if (close.evaluate().isNotEmpty) {
    await tester.tap(close);
    await tester.pumpAndSettle();
  }
  client.publish({
    ...snapshot.board,
    'cards': [
      ...snapshot.board['cards'] as List,
      for (final (id, rank, suit) in [
        ('context-king-clubs', 13, 'clubs'),
        ('context-king-diamonds', 13, 'diamonds'),
        ('context-queen-hearts', 12, 'hearts'),
      ])
        {
          'card': {'id': id, 'rank': rank, 'suit': suit, 'deck': 1},
          'controller': 1,
          'zone': 'hand',
        },
    ],
  }, online: snapshot.online);
  await tester.pumpAndSettle();
  for (final id in ['context-king-clubs', 'context-king-diamonds']) {
    final card = find.byKey(ValueKey('adaptive-select-$id'));
    await tester.ensureVisible(card);
    await tester.tap(card);
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
  }
  await activatePrivateCard(tester, 'context-queen-hearts');
  final intent = find.byKey(
    const ValueKey('card-intent-open-formation-people'),
  );
  await tester.ensureVisible(intent);
  await tester.tap(intent);
  await tester.pumpAndSettle();
}

Future<void> chooseGameOption(WidgetTester tester, String label) async {
  final options = find.byTooltip('Game options');
  await tester.ensureVisible(options);
  await tester.pumpAndSettle();
  await tester.tap(options);
  await tester.pumpAndSettle();
  await tester.tap(find.text(label).hitTestable());
  await tester.pumpAndSettle();
}

// Form/payload regression fixtures seed the existing draft transition directly.
// The actual Game options menu never exposes these card actions; real pointer,
// double-tap, keyboard and drag entry is exercised in card_play_screen_test.dart.
void chooseAction(
  WidgetTester tester,
  String action, {
  Map<String, dynamic> seed = const {},
}) {
  tester
      .widget<PopupMenuButton<String>>(
        find.byWidgetPredicate(
          (w) => w is PopupMenuButton<String> && w.tooltip == 'Game options',
        ),
      )
      .onSelected!(action);
}
