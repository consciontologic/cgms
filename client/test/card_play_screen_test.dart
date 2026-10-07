import 'dart:ui' show PointerDeviceKind, SemanticsAction, Tristate;
import 'package:cgms_demo/online/card_intents.dart';
import 'package:cgms_demo/online/card_play_choices.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/semantics.dart' show CustomSemanticsAction;
import 'package:flutter_test/flutter_test.dart';
import 'online_screen_test.dart'
    show ViewClient, showOnline, chooseGameOption, chooseAction, openPublicSeat;

Map<String, dynamic> playCard(
  String id,
  int rank,
  String suit, {
  int controller = 1,
  String zone = 'hand',
}) => {
  'card': {'id': id, 'rank': rank, 'suit': suit, 'deck': 1},
  'controller': controller,
  'zone': zone,
};

void main() {
  for (final kind in [PointerDeviceKind.mouse, PointerDeviceKind.touch]) {
    testWidgets(
      'large first-opening preview names all 18 exact copies and costs before $kind release',
      (tester) async {
        final cards = [
          for (var rank = 2; rank <= 10; rank++)
            for (var copy = 1; copy <= 2; copy++)
              {
                ...playCard('club-$rank-$copy', rank, 'clubs'),
                'card': {
                  'id': 'club-$rank-$copy',
                  'rank': rank,
                  'suit': 'clubs',
                  'deck': copy,
                },
              },
        ];
        final client = ViewClient()
          ..publish({
            'turn': 1,
            'cards': cards,
            'players': [
              {
                'seat': 1,
                'hand_count': 18,
                'ace_count': 0,
                'history': ['diamonds'],
              },
            ],
          });
        addTearDown(client.dispose);
        await showOnline(tester, client, expand: false);
        final source = find.byKey(const ValueKey('adaptive-select-club-2-1'));
        await tester.ensureVisible(source);
        await tester.pumpAndSettle();
        final pointer = await tester.startGesture(
          tester.getTopLeft(source) + const Offset(8, 8),
          kind: kind,
        );
        if (kind == PointerDeviceKind.touch)
          await tester.pump(const Duration(milliseconds: 600));
        await pointer.moveBy(const Offset(0, -25));
        await tester.pump();
        await pointer.moveTo(
          tester.getCenter(find.byKey(const ValueKey('seat-target-1'))),
        );
        await tester.pump();
        final preview = tester
            .widget<Text>(find.textContaining('Release to submit Open Series'))
            .data!;
        expect(
          preview,
          contains(
            'Clubs: ${[for (var rank = 2; rank <= 10; rank++) '$rank[1,2]'].join(' ')} / Seat 1',
          ),
        );
        expect(preview, contains('Copy numbers in brackets.'));
        expect(preview, isNot(contains('→')));
        expect(preview, contains('Uses this turn’s opening allowance.'));
        expect(client.submissions, isEmpty);
        await tester.sendKeyEvent(LogicalKeyboardKey.escape);
        await pointer.up();
        await tester.pumpAndSettle();
        expect(client.submissions, isEmpty);
      },
    );

    testWidgets(
      'complete own-area $kind drop previews all first-opening cards and submits once',
      (tester) async {
        final client = ViewClient()
          ..publish({
            'turn': 1,
            'cards': [
              playCard('own-four-1', 4, 'hearts'),
              {
                ...playCard('own-four-2', 4, 'hearts'),
                'card': {
                  'id': 'own-four-2',
                  'rank': 4,
                  'suit': 'hearts',
                  'deck': 2,
                },
              },
            ],
          });
        addTearDown(client.dispose);
        await showOnline(tester, client, expand: false);
        final pointer = await tester.startGesture(
          tester.getCenter(
            find.byKey(const ValueKey('adaptive-select-own-four-1')),
          ),
          kind: kind,
        );
        if (kind == PointerDeviceKind.touch)
          await tester.pump(const Duration(milliseconds: 600));
        await pointer.moveBy(const Offset(0, -25));
        await tester.pump();
        await pointer.moveTo(
          tester.getCenter(find.byKey(const ValueKey('seat-target-1'))),
        );
        await tester.pump();
        expect(client.submissions, isEmpty);
        final preview = find.textContaining('Release to submit Open Series');
        expect(preview, findsOneWidget);
        expect(
          tester.widget<Text>(preview).data,
          contains('Uses this turn’s opening allowance.'),
        );
        expect(
          tester.widget<Text>(preview).data,
          contains('Hearts: 4[1,2]'),
          reason:
              'Both physical first-opening copies are previewed before release',
        );
        await pointer.up();
        await tester.pumpAndSettle();
        expect(client.submissions, [
          {
            'type': 'open-series',
            'payload': {
              'suit': 'hearts',
              'selection': ['own-four-1', 'own-four-2'],
            },
          },
        ]);
        expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
        await tester.pump();
        expect(client.submissions, hasLength(1));
      },
    );
  }

  testWidgets(
    'own-area royal drop asks only the missing supported series and commits that choice',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [
            playCard('broad-king', 13, 'spades'),
            playCard('support-h', 4, 'hearts', zone: 'series'),
            playCard('support-d', 4, 'diamonds', zone: 'series'),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await dragCard(tester, 'broad-king', 'seat-target-1');
      expect(client.submissions, isEmpty);
      expect(find.text('Choose on the board'), findsNothing);
      await tester.tap(
        find.widgetWithText(OutlinedButton, 'Attach to Diamonds'),
      );
      await tester.pumpAndSettle();
      expect(client.submissions, [
        {
          'type': 'attach',
          'payload': {
            'selection': ['broad-king'],
            'suit': 'diamonds',
          },
        },
      ]);
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
    },
  );

  testWidgets(
    'held direct drop is revoked by a new turn or disconnect without submission',
    (tester) async {
      final cards = [playCard('revoked-heart', 5, 'hearts')];
      final client = ViewClient()..publish({'turn': 1, 'cards': cards});
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      for (final disconnect in [false, true]) {
        client.publish({'turn': 1, 'cards': cards});
        await tester.pumpAndSettle();
        final pointer = await tester.startGesture(
          tester.getCenter(
            find.byKey(const ValueKey('adaptive-select-revoked-heart')),
          ),
          kind: PointerDeviceKind.mouse,
        );
        await pointer.moveBy(const Offset(0, -25));
        await tester.pump();
        await pointer.moveTo(
          tester.getCenter(find.byKey(const ValueKey('seat-target-1'))),
        );
        await tester.pump();
        if (disconnect) {
          client.connected = false;
          client.notifyListeners();
        } else {
          client.publish({'turn': 2, 'cards': cards});
        }
        await tester.pump();
        await pointer.up();
        await tester.pumpAndSettle();
        expect(client.submissions, isEmpty);
      }
    },
  );

  testWidgets(
    'Confinement drop uses the pending actor while protecting another player',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {
            'turn': 1,
            'active': 2,
            'window_id': 'third-party-attack',
            'required_actor': 1,
            'players': [
              for (final seat in [1, 2, 3])
                {
                  'seat': seat,
                  'hand_count': 0,
                  'ace_count': seat == 1 ? 1 : 0,
                  'history': ['clubs', 'hearts'],
                },
            ],
            'cards': [
              playCard('response-ace', 1, 'diamonds', zone: 'concealed-ace'),
              playCard('support-h', 4, 'hearts', zone: 'series'),
              playCard('support-c', 4, 'clubs', zone: 'series'),
            ],
          },
          online: {
            'rules_context': {
              'purchase_unit_price': 5,
              'pending': {
                'actor': 2,
                'target_seat': 3,
                'type': 'attack',
                'response_types': ['confinement'],
              },
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      await dragCard(tester, 'response-ace', 'seat-target-2');
      expect(client.submissions, [
        {
          'type': 'confinement',
          'payload': {
            'pending_action_id': 'third-party-attack',
            'ace': 'response-ace',
          },
        },
      ]);
      expect(find.text('Confirm Confinement'), findsNothing);
    },
  );

  testWidgets(
    'exposed formation header drops exact private loan proposal without transfer',
    (tester) async {
      final ids = ['loan-j1', 'loan-j2'];
      final client = ViewClient()
        ..publish({
          'turn': 1,
          'players': [
            for (final seat in [1, 2])
              {
                'seat': seat,
                'hand_count': 0,
                'ace_count': 0,
                'history': ['diamonds'],
              },
          ],
          'cards': [
            for (final id in ids)
              {
                ...playCard(id, 11, 'clubs', zone: 'formation'),
                'allocation': 'kid',
              },
          ],
          'formations': [
            {
              'id': 'kid',
              'controller': 1,
              'spec': {'kind': 'kidnapper', 'cards': ids},
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final pointer = await tester.startGesture(
        tester.getCenter(find.byKey(const ValueKey('formation-target-1-kid'))),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -25));
      await tester.pump();
      await pointer.moveTo(
        tester.getCenter(find.byKey(const ValueKey('seat-target-2'))),
      );
      await tester.pump();
      expect(find.textContaining('exclusive alliance'), findsOneWidget);
      expect(client.submissions, isEmpty);
      await pointer.up();
      await tester.pumpAndSettle();
      expect(client.submissions, hasLength(1));
      final command = client.submissions.single;
      expect(command['type'], 'offer');
      expect(command['payload']['revision'], 1);
      expect(command['payload']['terms'], {
        'to': 2,
        'give': ids,
        'receive': <String>[],
        'loan': true,
      });
      expect(client.snapshot!.board['formations'][0]['controller'], 1);
      expect(client.snapshot!.online['alliances'], {'1': 0});
      expect(find.text('Confirm Ponzi'), findsNothing);
    },
  );

  testWidgets(
    'formation header over nested enemy series proposes loan and never activates attack',
    (tester) async {
      final ids = ['loan-j1', 'loan-j2'];
      final client = ViewClient()
        ..publish({
          'turn': 1,
          'players': [
            for (final seat in [1, 2])
              {
                'seat': seat,
                'hand_count': 0,
                'ace_count': 0,
                'history': ['diamonds'],
              },
          ],
          'cards': [
            for (final id in ids)
              {
                ...playCard(id, 11, 'clubs', zone: 'formation'),
                'allocation': 'kid',
              },
            playCard(
              'enemy-series',
              5,
              'hearts',
              zone: 'series',
              controller: 2,
            ),
          ],
          'formations': [
            {
              'id': 'kid',
              'controller': 1,
              'spec': {'kind': 'kidnapper', 'cards': ids},
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final pointer = await tester.startGesture(
        tester.getCenter(find.byKey(const ValueKey('formation-target-1-kid'))),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -25));
      await tester.pump();
      await pointer.moveTo(
        tester.getCenter(
          find.byKey(const ValueKey('formation-target-2-hearts')),
        ),
      );
      await tester.pump();
      expect(find.textContaining('exclusive alliance'), findsOneWidget);
      expect(client.submissions, isEmpty);
      await pointer.up();
      await tester.pumpAndSettle();
      expect(client.submissions, hasLength(1));
      final command = client.submissions.single;
      expect(command['type'], 'offer');
      expect(command['payload']['revision'], 1);
      expect(command['payload']['terms'], {
        'to': 2,
        'give': ids,
        'receive': <String>[],
        'loan': true,
      });
      expect(client.snapshot!.board['formations'][0]['controller'], 1);
      expect(client.snapshot!.online['alliances'], {'1': 0});
      expect(find.text('Confirm Ponzi'), findsNothing);
    },
  );

  testWidgets('royal activation requires an explicitly chosen own series', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish({
        'cards': [
          playCard('attach-king', 13, 'spades'),
          playCard('attach-heart', 6, 'hearts', zone: 'series'),
          playCard('attach-spade', 5, 'spades', zone: 'series'),
        ],
      });
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    await doubleTapCard(tester, 'attach-king');
    await tester.tap(find.byKey(const ValueKey('card-intent-attach')));
    await tester.pumpAndSettle();
    final confirm = find.byKey(const ValueKey('online-submit'));
    expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
    expect(
      find.widgetWithText(OutlinedButton, 'Attach to Hearts'),
      findsOneWidget,
    );
    expect(client.submissions, isEmpty);
    await tester.tap(find.text('Choose a supported series'));
    await tester.pumpAndSettle();
    await tester.tap(find.byKey(const ValueKey('formation-target-1-spades')));
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
    await tester.ensureVisible(confirm);
    await tester.tap(confirm);
    await tester.pumpAndSettle();
    expect(client.submissions.single, {
      'type': 'attach',
      'payload': {
        'selection': ['attach-king'],
        'suit': 'spades',
      },
    });
  });

  for (final situation in ['off-turn', 'confined', 'restricted', 'pending']) {
    testWidgets(
      'deliberate Ace disposal remains card-driven while $situation',
      (tester) async {
        final client = ViewClient()
          ..publish(
            {
              'active': situation == 'off-turn' ? 2 : 1,
              'players': [
                {
                  'seat': 1,
                  'hand_count': 0,
                  'ace_count': 1,
                  'confined': situation == 'confined',
                  'departed': false,
                  'history': ['clubs'],
                },
                {'seat': 2, 'hand_count': 0, 'ace_count': 0, 'history': []},
              ],
              if (situation == 'pending') ...{
                'window_id': 'someone-elses-action',
                'decision_id': 'someone-elses-choice',
                'decision_kind': 'justice',
                'required_actor': 2,
              },
              'cards': [
                {
                  ...playCard(
                    'discard-ace',
                    1,
                    'hearts',
                    zone: 'concealed-ace',
                  ),
                  if (situation == 'restricted') 'available_from_round': 5,
                },
              ],
            },
            online: {
              'rules_context': {
                'ace_used_this_round': true,
                'purchase_unit_price': 5,
              },
            },
          );
        addTearDown(client.dispose);
        await showOnline(tester, client, expand: false);
        final aceCard = find.byKey(
          const ValueKey('adaptive-select-discard-ace'),
        );
        await tester.tap(aceCard);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        await tester.tap(find.byTooltip('Inspect selected cards'));
        await tester.pumpAndSettle();
        expect(find.text('Confirm Expose Unused Ace'), findsNothing);
        expect(client.submissions, isEmpty);
        await tester.tap(find.byKey(const ValueKey('close-card-context')));
        await tester.pumpAndSettle();
        await doubleTapCard(tester, 'discard-ace');
        expect(find.text('Confirm Expose Unused Ace'), findsNothing);
        expect(client.submissions, isEmpty);
        await tester.tap(
          find.byKey(const ValueKey('card-intent-expose-unused-ace')),
        );
        await tester.pumpAndSettle();
        expect(find.text('Confirm Expose Unused Ace'), findsOneWidget);
        expect(client.submissions, isEmpty);
        await tester.sendKeyEvent(LogicalKeyboardKey.escape);
        await tester.pumpAndSettle();
        expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
        expect(client.submissions, isEmpty);
        await doubleTapCard(tester, 'discard-ace');
        await tester.tap(
          find.byKey(const ValueKey('card-intent-expose-unused-ace')),
        );
        await tester.pumpAndSettle();
        final confirm = find.byKey(const ValueKey('online-submit'));
        expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
        await tester.ensureVisible(confirm);
        await tester.tap(confirm);
        await tester.pumpAndSettle();
        expect(client.submissions.single, {
          'type': 'expose-unused-ace',
          'payload': {'ace': 'discard-ace'},
        });
      },
    );
  }

  for (final mode in [0, 7]) {
    testWidgets(
      'offensive Ace mode $mode preserves exact Clubs targets and concealed cost',
      (tester) async {
        final client = ViewClient()
          ..publish({
            'turn': 2,
            'players': [
              {
                'seat': 1,
                'hand_count': 0,
                'ace_count': 1,
                'history': ['clubs'],
              },
              {
                'seat': 2,
                'hand_count': 0,
                'ace_count': 0,
                'history': ['hearts'],
              },
            ],
            'cards': [
              playCard(
                'attack-ace',
                1,
                mode == 0 ? 'clubs' : 'spades',
                zone: 'concealed-ace',
              ),
              playCard(
                'physical-club',
                mode == 0 ? 2 : 5,
                'clubs',
                zone: 'series',
              ),
              playCard(
                'target-five',
                5,
                'hearts',
                controller: 2,
                zone: 'series',
              ),
              playCard(
                'target-seven',
                7,
                'hearts',
                controller: 2,
                zone: 'series',
              ),
            ],
          });
        addTearDown(client.dispose);
        await showOnline(tester, client, expand: false);
        await doubleTapCard(tester, 'attack-ace');
        await tester.tap(find.byKey(const ValueKey('card-intent-attack')));
        await tester.pumpAndSettle();
        final selection = find.text('Select attacking Clubs');
        await tester.ensureVisible(selection);
        await tester.tap(selection);
        await tester.pumpAndSettle();
        await openPublicSeat(tester, 'physical-club');
        await tester.tap(
          find.byKey(const ValueKey('adaptive-select-physical-club')),
        );
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        await openPublicSeat(tester, 'physical-club');
        await doubleTapCard(tester, 'physical-club');
        final targets = find.text('Choose the target card');
        await tester.ensureVisible(targets);
        await tester.tap(targets);
        await tester.pumpAndSettle();
        await tester.tap(
          find.byKey(const ValueKey('formation-target-2-hearts')),
        );
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        final confirm = find.byKey(const ValueKey('online-submit'));
        expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
        final value = find.widgetWithText(
          OutlinedButton,
          mode == 0 ? 'Advancement · +10' : '7',
        );
        await tester.ensureVisible(value);
        await tester.tap(value);
        await tester.pumpAndSettle();
        final table = tester.widget<CgmsAdaptiveTable>(
          find.byType(CgmsAdaptiveTable),
        );
        expect(table.cardRoles['attack-ace'], TableCardRole.modifier);
        expect(table.cardRoles['physical-club'], TableCardRole.source);
        expect(table.cardRoles['target-five'], TableCardRole.target);
        expect(table.cardRoles['target-seven'], TableCardRole.target);
        expect(
          client.snapshot!.board['cards']
              .where((card) => card['card']['id'] == 'attack-ace')
              .single['zone'],
          'concealed-ace',
        );
        expect(client.submissions, isEmpty);
        expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
        await tester.ensureVisible(confirm);
        await tester.tap(confirm);
        await tester.pumpAndSettle();
        expect(client.submissions.single, {
          'type': 'attack',
          'payload': {
            'selection': ['physical-club'],
            'targets': ['target-five', 'target-seven'],
            'suit': 'hearts',
            'ace': 'attack-ace',
            'value': mode,
          },
        });
        expect(tester.takeException(), isNull);
      },
    );
  }

  testWidgets(
    'idle online board has no action workspace or reserved controls',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [playCard('held-heart', 6, 'hearts')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(
        find.byKey(
          const PageStorageKey('online-action-panel'),
          skipOffstage: false,
        ),
        findsNothing,
      );
      expect(
        find.byKey(const ValueKey('online-action-form'), skipOffstage: false),
        findsNothing,
      );
      expect(find.text('Play cards'), findsNothing);
      expect(find.text('Prepare your action'), findsNothing);
      expect(find.byTooltip('Open actions'), findsNothing);
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      final card = find.byKey(const ValueKey('adaptive-select-held-heart'));
      await tester.tap(card);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'selecting an already focused private card keeps it visible through every scroll viewport',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final client = ViewClient()
        ..publish({
          'cards': [
            for (final suit in ['hearts', 'spades', 'diamonds', 'clubs'])
              for (final rank in [2, 3, 4, 5, 6, 10])
                playCard('focus-$suit-$rank', rank, suit),
          ],
        });
      addTearDown(client.dispose);
      try {
        await showOnline(tester, client, expand: false);
        await tester.binding.setSurfaceSize(const Size(390, 844));
        await tester.pumpAndSettle();
        final card = find.byKey(
          const ValueKey('adaptive-select-focus-spades-10'),
        );
        final ink = tester.widget<InkWell>(
          find.descendant(of: card, matching: find.byType(InkWell)).first,
        );
        ink.focusNode!.requestFocus();
        await tester.pumpAndSettle();
        final focus = FocusManager.instance.primaryFocus;
        void visible(String step) {
          final bounds = tester.getRect(card);
          for (final scroll
              in find
                  .ancestor(of: card, matching: find.byType(Scrollable))
                  .evaluate()) {
            final widget = scroll.widget as Scrollable;
            final viewport = tester.getRect(
              find.byElementPredicate((element) => element == scroll),
            );
            if (widget.axisDirection == AxisDirection.down ||
                widget.axisDirection == AxisDirection.up) {
              expect(
                bounds.top,
                greaterThanOrEqualTo(viewport.top - .01),
                reason: '$step ${widget.axisDirection} $viewport',
              );
              expect(
                bounds.bottom,
                lessThanOrEqualTo(viewport.bottom + .01),
                reason: '$step ${widget.axisDirection} $viewport',
              );
            }
          }
        }

        visible('before selection');
        expect(find.byTooltip('Inspect selected cards'), findsNothing);
        await tester.sendKeyEvent(LogicalKeyboardKey.space);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        expect(FocusManager.instance.primaryFocus, same(focus));
        expect(focus!.hasFocus, isTrue);
        expect(
          tester
              .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
              .selectedCardIds,
          {'focus-spades-10'},
        );
        expect(find.byTooltip('Inspect selected cards'), findsOneWidget);
        visible('after selection');
        expect(client.submissions, isEmpty);
        expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
        expect(tester.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'connection notice removal preserves private card semantic geometry and focus',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final client = CardConnectionClient()
        ..publish({
          'cards': [
            for (final suit in ['hearts', 'spades', 'diamonds', 'clubs'])
              for (final rank in [10])
                playCard('reconnect-$suit-$rank', rank, suit),
          ],
        });
      addTearDown(client.dispose);
      try {
        await showOnline(tester, client, expand: false);
        await tester.binding.setSurfaceSize(const Size(390, 844));
        await tester.pumpAndSettle();
        final card = find.byKey(
          const ValueKey('adaptive-select-reconnect-spades-10'),
        );
        final ink = tester.widget<InkWell>(
          find.descendant(of: card, matching: find.byType(InkWell)).first,
        );
        ink.focusNode!.requestFocus();
        await tester.pumpAndSettle();
        await tester.tap(card);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        final focus = ink.focusNode!;
        final cardElement = tester.element(card);
        final semanticId = tester
            .getSemantics(
              find.byKey(const ValueKey('select-reconnect-spades-10')),
            )
            .id;
        void check(String step) {
          var node = tester.getSemantics(
            find.byKey(const ValueKey('select-reconnect-spades-10')),
          );
          var rect = node.rect;
          while (true) {
            if (node.transform case final transform?)
              rect = MatrixUtils.transformRect(transform, rect);
            final parent = node.parent;
            if (parent == null) break;
            node = parent;
          }
          final ratio = tester.view.devicePixelRatio;
          rect = Rect.fromLTRB(
            rect.left / ratio,
            rect.top / ratio,
            rect.right / ratio,
            rect.bottom / ratio,
          );
          // Semantics reports the visible portion of a physical card. Match
          // the same nested viewport clipping used by its painted content.
          var painted = tester.getRect(card);
          for (final scroll
              in find
                  .ancestor(of: card, matching: find.byType(Scrollable))
                  .evaluate()) {
            painted = painted.intersect(
              tester.getRect(
                find.byElementPredicate((element) => element == scroll),
              ),
            );
          }
          expect(painted.width, greaterThanOrEqualTo(48), reason: step);
          expect(painted.height, greaterThanOrEqualTo(48), reason: step);
          expect(rect.left, closeTo(painted.left, 1 / 64), reason: step);
          expect(rect.top, closeTo(painted.top, 1 / 64), reason: step);
          expect(rect.width, closeTo(painted.width, 1 / 64), reason: step);
          expect(rect.height, closeTo(painted.height, 1 / 64), reason: step);
          expect(tester.element(card), same(cardElement), reason: step);
          expect(
            tester
                .getSemantics(
                  find.byKey(const ValueKey('select-reconnect-spades-10')),
                )
                .id,
            semanticId,
            reason: step,
          );
          expect(FocusManager.instance.primaryFocus, same(focus), reason: step);
          expect(focus.hasFocus, isTrue, reason: step);
          expect(
            tester
                .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
                .selectedCardIds,
            {'reconnect-spades-10'},
            reason: step,
          );
          expect(client.submissions, isEmpty, reason: step);
        }

        check('connected');
        client.setConnection(false);
        await tester.pumpAndSettle();
        expect(
          find.text(
            CgmsLocalizations.of(tester.element(card)).onlineConnectionLost,
          ),
          findsOneWidget,
        );
        check('notice visible');
        client.setConnection(true);
        await tester.pumpAndSettle();
        expect(
          find.text(
            CgmsLocalizations.of(tester.element(card)).onlineConnectionLost,
          ),
          findsNothing,
        );
        check('notice removed');
        expect(tester.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'neutral taps distinguish an own source from an opponent target',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'players': [
            {
              'seat': 1,
              'hand_count': 1,
              'ace_count': 0,
              'history': ['clubs'],
            },
            {
              'seat': 2,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['diamonds'],
            },
          ],
          'cards': [
            playCard('tap-source', 7, 'clubs'),
            playCard(
              'tap-target',
              7,
              'diamonds',
              controller: 2,
              zone: 'series',
            ),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-tap-source')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await openPublicSeat(tester, 'tap-target');
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-tap-target')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      final table = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      expect(table.selectedCardIds, {'tap-source', 'tap-target'});
      expect(table.cardRoles['tap-source'], TableCardRole.source);
      expect(table.cardRoles['tap-target'], TableCardRole.target);
      expect(table.draggableCardIds, contains('tap-source'));
      expect(table.draggableCardIds, isNot(contains('tap-target')));
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'card-driven trade explicitly replaces the current proposal revision',
    (tester) async {
      final semantics = tester.ensureSemantics();
      try {
        final client = ViewClient()
          ..publish({
            'players': [
              {
                'seat': 1,
                'hand_count': 1,
                'ace_count': 0,
                'history': ['clubs'],
              },
              {
                'seat': 2,
                'hand_count': 0,
                'ace_count': 0,
                'history': ['diamonds'],
              },
            ],
            'cards': [playCard('trade-queen', 12, 'hearts')],
            'proposals': [
              {
                'id': 'existing-offer',
                'from': 1,
                'revision': 3,
                'status': 'offered',
                'terms': {
                  'to': 2,
                  'give': ['trade-queen'],
                  'receive': <String>[],
                  'loan': false,
                },
              },
            ],
          });
        addTearDown(client.dispose);
        await showOnline(tester, client, expand: false);
        await doubleTapCard(tester, 'trade-queen');
        expect(
          find.byKey(const ValueKey('card-intent-offer-trade')),
          findsNothing,
        );
        expect(find.text('Offer'), findsOneWidget);
        final confirm = find.byKey(const ValueKey('online-submit'));
        expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
        final recipient = find.widgetWithText(OutlinedButton, 'Seat 2');
        Finder selectedMarker(Finder choice) => find
            .ancestor(
              of: choice,
              matching: find.byWidgetPredicate(
                (widget) =>
                    widget is Semantics && widget.properties.selected != null,
              ),
            )
            .first;
        expect(
          tester
              .getSemantics(selectedMarker(recipient))
              .getSemanticsData()
              .flagsCollection
              .isSelected,
          Tristate.isFalse,
        );
        await tester.ensureVisible(recipient);
        await tester.tap(recipient);
        await tester.pumpAndSettle();
        expect(
          tester
              .getSemantics(selectedMarker(recipient))
              .getSemanticsData()
              .flagsCollection
              .isSelected,
          Tristate.isTrue,
        );
        expect(client.submissions, isEmpty);
        final replacement = find.widgetWithText(
          OutlinedButton,
          'Replace offer to seat 2 · revision 4',
        );
        await tester.ensureVisible(replacement);
        await tester.tap(replacement);
        await tester.pumpAndSettle();
        expect(
          tester
              .getSemantics(selectedMarker(replacement))
              .getSemanticsData()
              .flagsCollection
              .isSelected,
          Tristate.isTrue,
        );
        expect(client.submissions, isEmpty);
        await tester.ensureVisible(confirm);
        await tester.tap(confirm);
        await tester.pumpAndSettle();
        expect(client.submissions.single, {
          'type': 'offer',
          'payload': {
            'offer_id': 'existing-offer',
            'revision': 4,
            'terms': {
              'to': 2,
              'give': ['trade-queen'],
              'receive': <String>[],
              'loan': false,
            },
          },
        });
        expect(tester.takeException(), isNull);
      } finally {
        semantics.dispose();
      }
    },
  );

  testWidgets(
    'focused trade recipient keeps its identity and visibility across live breakpoints and rotation',
    (tester) async {
      final semantics = tester.ensureSemantics();
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final client = ViewClient()
        ..publish({
          'players': [
            {
              'seat': 1,
              'hand_count': 1,
              'ace_count': 0,
              'history': ['clubs'],
            },
            {
              'seat': 2,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['diamonds'],
            },
          ],
          'cards': [playCard('focused-trade', 12, 'hearts')],
          'proposals': [
            {
              'id': 'focused-offer',
              'from': 1,
              'revision': 3,
              'status': 'offered',
              'terms': {
                'to': 2,
                'give': ['focused-trade'],
                'receive': <String>[],
                'loan': false,
              },
            },
          ],
        });
      addTearDown(client.dispose);
      try {
        await showOnline(tester, client, expand: false);
        tester.view.physicalSize = const Size(390, 844);
        await tester.binding.setSurfaceSize(const Size(390, 844));
        await tester.pumpAndSettle();
        await doubleTapCard(tester, 'focused-trade');
        expect(
          find.byKey(const ValueKey('card-intent-offer-trade')),
          findsNothing,
        );
        expect(find.text('Offer'), findsOneWidget);
        final recipient = find.widgetWithText(OutlinedButton, 'Seat 2');
        await tester.ensureVisible(recipient);
        await tester.tap(recipient);
        await tester.pumpAndSettle();
        final replacement = find.widgetWithText(
          OutlinedButton,
          'Replace offer to seat 2 · revision 4',
        );
        await tester.ensureVisible(replacement);
        await tester.tap(replacement);
        await tester.pumpAndSettle();
        await tester.ensureVisible(recipient);
        await tester.pumpAndSettle();
        final recipientElement = tester.element(recipient);
        final semanticsNode = tester.getSemantics(recipient);
        semanticsNode.owner!.performAction(
          semanticsNode.id,
          SemanticsAction.focus,
        );
        await tester.pumpAndSettle();
        final focus = FocusManager.instance.primaryFocus!;
        final scope = focus.enclosingScope;
        expect(focus.hasFocus, isTrue);
        expect(focus.context, isNotNull);
        final initialPopup = tester.getRect(
          find.byKey(const ValueKey('contextual-popup')),
        );
        final initialControl = tester.getRect(recipient);
        expect(
          initialControl.top,
          greaterThanOrEqualTo(initialPopup.top),
          reason: 'initial portrait',
        );
        expect(
          initialControl.bottom,
          lessThanOrEqualTo(initialPopup.bottom),
          reason: 'initial portrait',
        );
        expect(
          find.ancestor(
            of: find.byElementPredicate((element) => element == focus.context),
            matching: recipient,
          ),
          findsOneWidget,
        );
        for (final size in [
          const Size(599, 844),
          const Size(600, 844),
          const Size(599, 844),
          const Size(1049, 844),
          const Size(1050, 844),
          const Size(1049, 844),
          const Size(844, 390),
          const Size(390, 844),
        ]) {
          tester.view.physicalSize = size;
          await tester.binding.setSurfaceSize(size);
          await tester.pumpAndSettle();
          expect(
            tester.element(recipient),
            same(recipientElement),
            reason: '$size',
          );
          expect(
            FocusManager.instance.primaryFocus,
            same(focus),
            reason: '$size',
          );
          expect(focus.enclosingScope, same(scope), reason: '$size');
          expect(focus.hasFocus, isTrue, reason: '$size');
          expect(
            tester.getSemantics(recipient).id,
            semanticsNode.id,
            reason: '$size',
          );
          final popup = tester.getRect(
            find.byKey(const ValueKey('contextual-popup')),
          );
          final control = tester.getRect(recipient);
          expect(control.top, greaterThanOrEqualTo(popup.top), reason: '$size');
          expect(
            control.bottom,
            lessThanOrEqualTo(popup.bottom),
            reason: '$size',
          );
          expect(control.left, greaterThanOrEqualTo(0), reason: '$size');
          expect(control.right, lessThanOrEqualTo(size.width), reason: '$size');
          expect(control.height, greaterThanOrEqualTo(48), reason: '$size');
          final marker = find
              .ancestor(
                of: recipient,
                matching: find.byWidgetPredicate(
                  (widget) =>
                      widget is Semantics && widget.properties.selected != null,
                ),
              )
              .first;
          expect(
            tester
                .getSemantics(marker)
                .getSemanticsData()
                .flagsCollection
                .isSelected,
            Tristate.isTrue,
            reason: '$size',
          );
          final table = tester.widget<CgmsAdaptiveTable>(
            find.byType(CgmsAdaptiveTable),
          );
          expect(table.selectedCardIds, {'focused-trade'}, reason: '$size');
          expect(client.submissions, isEmpty, reason: '$size');
          expect(tester.takeException(), isNull, reason: '$size');
        }
        final confirm = find.byKey(const ValueKey('online-submit'));
        await tester.ensureVisible(confirm);
        await tester.tap(confirm);
        await tester.pumpAndSettle();
        expect(client.submissions.single, {
          'type': 'offer',
          'payload': {
            'offer_id': 'focused-offer',
            'revision': 4,
            'terms': {
              'to': 2,
              'give': ['focused-trade'],
              'receive': <String>[],
              'loan': false,
            },
          },
        });
      } finally {
        semantics.dispose();
      }
    },
  );

  for (final activation in [
    'double tap',
    'selected double tap',
    'keyboard',
    'assistive',
  ]) {
    testWidgets(
      'series header $activation activates exact number members rather than its first royal',
      (tester) async {
        final semantics = tester.ensureSemantics();
        try {
          final client = ViewClient()
            ..publish({
              'turn': 2,
              'players': [
                {
                  'seat': 1,
                  'hand_count': 0,
                  'ace_count': 0,
                  'history': ['clubs'],
                },
              ],
              'cards': [
                {
                  ...playCard('first-exile', 12, 'clubs', zone: 'attachment'),
                  'allocation': 'clubs',
                },
                playCard('series-club', 8, 'clubs', zone: 'series'),
              ],
            });
          addTearDown(client.dispose);
          await showOnline(tester, client, expand: false);
          final header = find.byKey(const ValueKey('formation-target-1-clubs'));
          if (activation == 'selected double tap') {
            await tester.tap(header);
            await tester.pumpAndSettle(const Duration(milliseconds: 350));
            expect(
              tester
                  .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
                  .selectedCardIds,
              {'series-club'},
            );
            expect(
              find.byKey(const ValueKey('contextual-popup')),
              findsNothing,
            );
          }
          if (activation == 'double tap' ||
              activation == 'selected double tap') {
            await tester.tap(header);
            await tester.pump(const Duration(milliseconds: 50));
            await tester.tap(header);
            await tester.pumpAndSettle(const Duration(milliseconds: 350));
          } else {
            final node = tester.getSemantics(header);
            if (activation == 'keyboard') {
              node.owner!.performAction(node.id, SemanticsAction.focus);
              await tester.pumpAndSettle();
              await tester.sendKeyEvent(LogicalKeyboardKey.enter);
            } else {
              final action = node
                  .getSemanticsData()
                  .customSemanticsActionIds!
                  .singleWhere(
                    (id) =>
                        CustomSemanticsAction.getAction(id)?.label ==
                        'Card actions',
                  );
              node.owner!.performAction(
                node.id,
                SemanticsAction.customAction,
                action,
              );
            }
            await tester.pumpAndSettle();
          }
          final table = tester.widget<CgmsAdaptiveTable>(
            find.byType(CgmsAdaptiveTable),
          );
          expect(table.selectedCardIds, {'series-club'});
          expect(table.cardRoles['first-exile'], isNull);
          expect(find.text('Confirm Attack'), findsOneWidget);
          expect(client.submissions, isEmpty);
          expect(tester.takeException(), isNull);
        } finally {
          semantics.dispose();
        }
      },
    );
  }

  testWidgets(
    'own series tap accepts held additions when a cross-suit attachment is projected first',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'players': [
            {
              'seat': 1,
              'hand_count': 2,
              'ace_count': 0,
              'history': ['hearts'],
            },
          ],
          'cards': [
            {
              ...playCard('first-attached', 13, 'clubs', zone: 'attachment'),
              'allocation': 'hearts',
            },
            playCard('public-heart', 3, 'hearts', zone: 'series'),
            playCard('held-heart-5', 5, 'hearts'),
            playCard('held-heart-7', 7, 'hearts'),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-held-heart-5')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await tester.tap(find.byKey(const ValueKey('formation-target-1-hearts')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.text('Confirm Open Series'), findsOneWidget);
      expect(client.submissions, isEmpty);
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'open-series',
        'payload': {
          'suit': 'hearts',
          'selection': ['held-heart-5'],
        },
      });
    },
  );

  testWidgets(
    'royal pointer drop targets a whole series without choosing its first attachment',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'players': [
            {
              'seat': 1,
              'hand_count': 1,
              'ace_count': 0,
              'history': ['spades'],
            },
          ],
          'cards': [
            {
              ...playCard('first-attached', 13, 'clubs', zone: 'attachment'),
              'allocation': 'spades',
            },
            playCard('public-heart', 3, 'spades', zone: 'series'),
            playCard('new-royal', 13, 'spades'),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final source = find.byKey(const ValueKey('adaptive-select-new-royal'));
      final destination = find.byKey(
        const ValueKey('formation-target-1-spades'),
      );
      final pointer = await tester.startGesture(
        tester.getCenter(source),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -24));
      await tester.pump();
      await pointer.moveTo(tester.getCenter(destination));
      await tester.pump();
      await pointer.up();
      await tester.pumpAndSettle();
      expect(find.text('Confirm Attach'), findsNothing);
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(client.submissions.single, {
        'type': 'attach',
        'payload': {
          'suit': 'spades',
          'selection': ['new-royal'],
        },
      });
    },
  );

  testWidgets(
    'invalid series drop never falls through to the enclosing own seat',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'players': [
            {
              'seat': 1,
              'hand_count': 1,
              'ace_count': 0,
              'history': ['spades'],
            },
          ],
          'cards': [
            playCard('held-heart', 4, 'hearts'),
            playCard('public-spade', 5, 'spades', zone: 'series'),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final source = find.byKey(const ValueKey('adaptive-select-held-heart'));
      final destination = find.byKey(
        const ValueKey('formation-target-1-spades'),
      );
      final pointer = await tester.startGesture(
        tester.getCenter(source),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -24));
      await tester.pump();
      await pointer.moveTo(tester.getCenter(destination));
      await tester.pump();
      await pointer.up();
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(client.submissions, isEmpty);
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        isEmpty,
      );
    },
  );

  testWidgets(
    'prepared attack on a series header commits its number members and never an attached royal',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 2,
          'players': [
            {
              'seat': 1,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['clubs'],
            },
            {
              'seat': 2,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['hearts'],
            },
          ],
          'cards': [
            playCard('attacking-8', 8, 'clubs', zone: 'series'),
            {
              ...playCard(
                'enemy-first-royal',
                13,
                'spades',
                controller: 2,
                zone: 'attachment',
              ),
              'allocation': 'hearts',
            },
            playCard('enemy-3', 3, 'hearts', controller: 2, zone: 'series'),
            playCard('enemy-5', 5, 'hearts', controller: 2, zone: 'series'),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await openPublicSeat(tester, 'attacking-8');
      await doubleTapCard(tester, 'attacking-8');
      final attack = find.byKey(const ValueKey('card-intent-attack'));
      if (attack.evaluate().isNotEmpty) {
        await tester.ensureVisible(attack);
        await tester.tap(attack);
        await tester.pumpAndSettle();
      }
      final target = find.text('Choose the target card');
      await tester.ensureVisible(target);
      await tester.tap(target);
      await tester.pumpAndSettle();
      await tester.tap(find.byKey(const ValueKey('formation-target-2-hearts')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      final table = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      expect(table.cardRoles['enemy-3'], TableCardRole.target);
      expect(table.cardRoles['enemy-5'], TableCardRole.target);
      expect(table.cardRoles['enemy-first-royal'], isNull);
      expect(client.submissions, isEmpty);
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'attack',
        'payload': {
          'selection': ['attacking-8'],
          'targets': ['enemy-3', 'enemy-5'],
          'suit': 'hearts',
        },
      });
    },
  );

  testWidgets(
    'double tap activates once while identical physical copies stay selected independently',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [
            playCard('copy-one', 6, 'hearts'),
            {
              ...playCard('copy-two', 6, 'hearts'),
              'card': {
                'id': 'copy-two',
                'rank': 6,
                'suit': 'hearts',
                'deck': 2,
              },
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.tap(find.byKey(const ValueKey('adaptive-select-copy-one')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        {'copy-one'},
      );
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      await doubleTapCard(tester, 'copy-two');
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        {'copy-one', 'copy-two'},
      );
      expect(find.byKey(const ValueKey('contextual-popup')), findsOneWidget);
      expect(client.submissions, isEmpty);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets('cancelled mouse and touch drags never submit or open a draft', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish({
        'turn': 1,
        'cards': [playCard('money', 10, 'spades')],
      });
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    for (final kind in [PointerDeviceKind.mouse, PointerDeviceKind.touch]) {
      final source = find.byKey(const ValueKey('adaptive-select-money'));
      final gesture = await tester.startGesture(
        tester.getCenter(source),
        kind: kind,
      );
      if (kind == PointerDeviceKind.touch)
        await tester.pump(const Duration(milliseconds: 600));
      await gesture.moveBy(const Offset(0, -30));
      await tester.pump();
      await gesture.cancel();
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions, isEmpty, reason: '$kind cancellation');
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
    }
  });

  testWidgets(
    'dismissing an ordered decision preserves it and offers reopening without Pass',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'window_id': 'w',
          'decision_id': 'd',
          'required_actor': 1,
          'decision_kind': 'dexter-prevention',
          'choices': [
            [],
            ['payment'],
          ],
          'cards': [playCard('payment', 3, 'diamonds', zone: 'series')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.byKey(const ValueKey('effect-choice-0')), findsOneWidget);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(find.text('Pass response'), findsNothing);
      expect(client.snapshot!.board['decision_id'], 'd');
      expect(client.submissions, isEmpty);
      final reopen = find.text('Pending decision');
      await tester.ensureVisible(reopen);
      await tester.tap(reopen);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.byKey(const ValueKey('effect-choice-0')), findsOneWidget);
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets(
    'tablet pending response fits above the footer without scrolling',
    (tester) async {
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final client = ViewClient()
        ..publish(
          {
            'active': 2,
            'window_id': 'response',
            'required_actor': 1,
            'players': [
              for (var seat = 1; seat <= 4; seat++)
                {'seat': seat, 'hand_count': 17, 'ace_count': 2},
            ],
            'cards': [
              for (var seat = 1; seat <= 4; seat++)
                for (var rank = 2; rank <= 8; rank++)
                  playCard(
                    'public-$seat-$rank',
                    rank,
                    'diamonds',
                    controller: seat,
                    zone: 'series',
                  ),
              for (var n = 0; n < 17; n++)
                playCard(
                  'hand-$n',
                  2 + n % 9,
                  ['hearts', 'spades', 'diamonds', 'clubs'][n % 4],
                ),
            ],
          },
          online: {
            'rules_context': {
              'purchase_unit_price': 5,
              'pending': {
                'type': 'open-series',
                'actor': 2,
                'response_types': <String>[],
              },
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      const size = Size(1035, 948);
      tester.view.physicalSize = size;
      await tester.binding.setSurfaceSize(size);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      final pass = find.widgetWithText(OutlinedButton, 'Pass response');
      final scroll = find
          .ancestor(of: pass, matching: find.byType(Scrollable))
          .first;
      expect(tester.state<ScrollableState>(scroll).position.pixels, 0);
      final viewport = tester.getRect(scroll);
      final passRect = tester.getRect(pass);
      expect(passRect.intersect(viewport), passRect);
      expect(passRect.height, greaterThanOrEqualTo(48));
      expect(pass.hitTestable(), findsOneWidget);
      final menu = find.byTooltip('Game options');
      expect(tester.getRect(menu).height, greaterThanOrEqualTo(48));
      expect(tester.getRect(menu).intersect(viewport), tester.getRect(menu));
      final public = tester.getRect(
        find.byKey(const ValueKey('adaptive-public-square')),
      );
      final hand = tester.getRect(
        find.byKey(const ValueKey('adaptive-hand-viewport')),
      );
      final firstCard = tester.getRect(
        find.byKey(const ValueKey('adaptive-select-hand-0')),
      );
      expect(firstCard.intersect(hand), firstCard);
      expect(hand.top, closeTo(public.bottom + 4, .01));
      final publicCard = tester.renderObject<RenderBox>(
        find.byKey(const ValueKey('public-preview-public-1-8')),
      );
      expect(
        publicCard.localToGlobal(Offset(publicCard.size.width, 0)).dx -
            publicCard.localToGlobal(Offset.zero).dx,
        greaterThanOrEqualTo(64),
      );
      await tester.tap(pass);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions.single, {
        'type': 'pass',
        'payload': {'pending_action_id': 'response'},
      });
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'resolved response closes its required context without sending Pass',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {'window_id': 'response', 'required_actor': 1},
          online: {
            'rules_context': {
              'pending': {
                'type': 'open-series',
                'actor': 2,
                'response_types': <String>[],
              },
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.text('Pass response'), findsOneWidget);
      expect(find.byKey(const ValueKey('contextual-popup')), findsOneWidget);
      client.publish({});
      await tester.pumpAndSettle();
      expect(find.text('Pass response'), findsNothing);
      expect(find.text('Pending decision'), findsNothing);
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'returned fixed pair rejects attachment and reopens exact bound combo by drop',
    (tester) async {
      const ids = ['returned-k1', 'returned-k2', 'returned-j'];
      final client = ViewClient()
        ..publish({
          'cards': [
            playCard(ids[0], 13, 'clubs'),
            playCard(ids[1], 13, 'diamonds'),
            playCard(ids[2], 11, 'clubs'),
            playCard('support-h', 5, 'hearts', zone: 'series'),
            playCard('support-c', 5, 'clubs', zone: 'series'),
          ],
          'own_bindings': [
            {
              'kings': ids.take(2).toList(),
              'slot': {'rank': 11, 'suit': 'clubs'},
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await dragCard(tester, ids[0], 'formation-target-1-hearts');
      expect(client.submissions, isEmpty);
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      for (final id in ids) {
        await tester.tap(find.byKey(ValueKey('adaptive-select-$id')));
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
      }
      await dragCard(tester, ids.last, 'seat-target-1');
      expect(client.submissions.single, {
        'type': 'open-formation',
        'payload': {
          'formation_id': startsWith('formation-'),
          'formation': {
            'kind': 'kidnapper',
            'cards': ids,
            'substitute': {
              'kings': ids.take(2).toList(),
              'slot': {'rank': 11, 'suit': 'clubs'},
            },
          },
        },
      });
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
    },
  );

  testWidgets(
    'public Doppelganger reconfiguration preserves its fixed binding without an editable role',
    (tester) async {
      const ids = ['binding-king-c', 'binding-king-d', 'binding-jack-c'];
      final client = ViewClient()
        ..publish({
          'cards': [
            for (final (id, rank, suit) in [
              ('binding-king-c', 13, 'clubs'),
              ('binding-king-d', 13, 'diamonds'),
              ('binding-jack-c', 11, 'clubs'),
            ])
              {
                ...playCard(id, rank, suit, zone: 'formation'),
                'allocation': 'kidnapper-fixed',
              },
          ],
          'formations': [
            {
              'id': 'kidnapper-fixed',
              'controller': 1,
              'spec': {
                'kind': 'kidnapper',
                'cards': ids,
                'substitute': {
                  'kings': ids.take(2).toList(),
                  'slot': {'rank': 11, 'suit': 'clubs'},
                },
              },
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      for (final id in ids.take(2)) {
        await openPublicSeat(tester, id);
        await tester.tap(find.byKey(ValueKey('adaptive-select-$id')));
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
      }
      await openPublicSeat(tester, ids.last);
      await doubleTapCard(tester, ids.last);
      final reconfigure = find.byKey(
        const ValueKey('card-intent-open-formation-kidnapper'),
      );
      await tester.ensureVisible(reconfigure);
      await tester.tap(reconfigure);
      await tester.pumpAndSettle();
      expect(find.text('Replacement rank · 11 · Clubs'), findsOneWidget);
      expect(find.widgetWithText(OutlinedButton, 'Jack'), findsNothing);
      expect(find.widgetWithText(OutlinedButton, 'Hearts'), findsNothing);
      expect(find.text('Select the two Doppelganger kings'), findsNothing);
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'open-formation',
        'payload': {
          'formation_id': 'kidnapper-fixed',
          'formation': {
            'kind': 'kidnapper',
            'cards': ids,
            'substitute': {
              'kings': ids.take(2).toList(),
              'slot': {'rank': 11, 'suit': 'clubs'},
            },
          },
        },
      });
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'Coup remains directly card-driven inside an ambiguous move choice',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [
            playCard('choice-money', 10, 'spades'),
            for (final suit in ['hearts', 'diamonds'])
              for (final copy in [1, 2])
                {
                  ...playCard('$suit-$copy', 12, suit),
                  'card': {
                    'id': '$suit-$copy',
                    'rank': 12,
                    'suit': suit,
                    'deck': copy,
                  },
                },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await doubleTapCard(tester, 'choice-money');
      expect(find.byType(CardPlayChoices), findsOneWidget);
      // A live restriction invalidates the unrelated purchase source, never Coup.
      client.publish({
        'cards': [
          for (final card in client.snapshot!.board['cards'] as List)
            if (card['card']['id'] == 'choice-money')
              {...card as Map<String, dynamic>, 'available_from_round': 5}
            else
              card,
        ],
      });
      await tester.pumpAndSettle();
      final queen = find.descendant(
        of: find.byKey(const ValueKey('contextual-popup')),
        matching: find.byWidgetPredicate(
          (w) => w is CgmsCard && w.card.id == 'hearts-1',
        ),
      );
      expect(queen, findsOneWidget);
      await tester.ensureVisible(queen);
      await tester.pumpAndSettle();
      expect(queen.hitTestable(), findsOneWidget);
      await tester.tap(queen);
      await tester.pump(const Duration(milliseconds: 80));
      await tester.tap(queen);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.text('Confirm Coup'), findsOneWidget);
      expect(client.submissions, isEmpty);
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(client.submissions, [
        {'type': 'coup', 'payload': {}},
      ]);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'an explicitly chosen Exile is a modifier distinct from committed Clubs',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 2,
          'cards': [
            playCard('combat-club', 7, 'clubs', zone: 'series'),
            {
              ...playCard('combat-exile', 12, 'clubs', zone: 'attachment'),
              'allocation': 'clubs',
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await openPublicSeat(tester, 'combat-exile');
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-combat-exile')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await openPublicSeat(tester, 'combat-club');
      await doubleTapCard(tester, 'combat-club');
      final attack = find.byKey(const ValueKey('card-intent-attack'));
      if (attack.evaluate().isNotEmpty) {
        await tester.ensureVisible(attack);
        await tester.tap(attack);
        await tester.pumpAndSettle();
      }
      final table = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      expect(table.cardRoles['combat-exile'], TableCardRole.modifier);
      expect(table.cardRoles['combat-club'], TableCardRole.source);
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'two Justice formations keep the chosen source and consumed Queen distinct',
    (tester) async {
      const suits = ['hearts', 'diamonds', 'clubs', 'spades'];
      final client = ViewClient()
        ..publish({
          'cards': [
            for (final f in ['justice-a', 'justice-b'])
              for (final suit in suits)
                {
                  ...playCard('$f-$suit', 12, suit, zone: 'formation'),
                  'allocation': f,
                },
          ],
          'formations': [
            for (final f in ['justice-a', 'justice-b'])
              {
                'id': f,
                'controller': 1,
                'spec': {
                  'kind': 'justice',
                  'cards': [for (final suit in suits) '$f-$suit'],
                },
              },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      for (final f in ['justice-a', 'justice-b']) {
        await openPublicSeat(tester, '$f-hearts');
        await tester.tap(find.byKey(ValueKey('adaptive-select-$f-hearts')));
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
      }
      await openPublicSeat(tester, 'justice-b-hearts');
      await doubleTapCard(tester, 'justice-b-hearts');
      final choice = find.byKey(
        const ValueKey('card-intent-justice-formation-justice-b'),
      );
      await tester.ensureVisible(choice);
      await tester.tap(choice);
      await tester.pumpAndSettle();
      final chooseQueen = find.text('Consumed Queen');
      await tester.ensureVisible(chooseQueen);
      await tester.tap(chooseQueen);
      await tester.pumpAndSettle();
      await openPublicSeat(tester, 'justice-b-diamonds');
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-justice-b-diamonds')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await openPublicSeat(tester, 'justice-b-diamonds');
      await doubleTapCard(tester, 'justice-b-diamonds');
      final table = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      expect(table.selectedCardIds, contains('justice-b-diamonds'));
      expect(table.selectedCardIds, isNot(contains('justice-a-hearts')));
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'justice',
        'payload': {'ace': 'justice-b-diamonds'},
      });
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'Exile activation alone can choose exact attacking Clubs and targets on the board',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 2,
          'players': [
            {
              'seat': 1,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['clubs'],
            },
            {
              'seat': 2,
              'hand_count': 0,
              'ace_count': 0,
              'history': ['hearts'],
            },
          ],
          'cards': [
            {
              ...playCard('alone-exile', 12, 'clubs', zone: 'attachment'),
              'allocation': 'clubs',
            },
            playCard('alone-club', 7, 'clubs', zone: 'series'),
            playCard(
              'alone-target',
              7,
              'hearts',
              controller: 2,
              zone: 'series',
            ),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await openPublicSeat(tester, 'alone-exile');
      await doubleTapCard(tester, 'alone-exile');
      final attack = find.byKey(const ValueKey('card-intent-attack'));
      if (attack.evaluate().isNotEmpty) {
        await tester.ensureVisible(attack);
        await tester.tap(attack);
        await tester.pumpAndSettle();
      }
      final confirm = find.byKey(const ValueKey('online-submit'));
      expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
      final select = find.text('Select attacking Clubs');
      await tester.ensureVisible(select);
      await tester.tap(select);
      await tester.pumpAndSettle();
      await openPublicSeat(tester, 'alone-club');
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-alone-club')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await openPublicSeat(tester, 'alone-club');
      await doubleTapCard(tester, 'alone-club');
      final targets = find.text('Choose the target card');
      await tester.ensureVisible(targets);
      await tester.tap(targets);
      await tester.pumpAndSettle();
      await openPublicSeat(tester, 'alone-target');
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-alone-target')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await openPublicSeat(tester, 'alone-club');
      await doubleTapCard(tester, 'alone-club');
      final table = tester.widget<CgmsAdaptiveTable>(
        find.byType(CgmsAdaptiveTable),
      );
      expect(table.cardRoles['alone-exile'], TableCardRole.modifier);
      expect(table.cardRoles['alone-club'], TableCardRole.source);
      expect(table.cardRoles['alone-target'], TableCardRole.target);
      expect(client.submissions, isEmpty);
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(client.submissions.single, {
        'type': 'attack',
        'payload': {
          'selection': ['alone-club'],
          'targets': ['alone-target'],
          'suit': 'hearts',
          'exile': true,
        },
      });
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'Negotiator resolves authorized attacker Clubs then defender payment without an implicit Pass',
    (tester) async {
      final players = [
        {
          'seat': 1,
          'hand_count': 0,
          'ace_count': 0,
          'history': ['hearts', 'spades'],
        },
        {
          'seat': 2,
          'hand_count': 0,
          'ace_count': 0,
          'history': ['clubs'],
        },
      ];
      final cards = [
        {
          ...playCard('negotiator', 11, 'spades', zone: 'attachment'),
          'allocation': 'spades',
        },
        playCard('payment-3', 3, 'spades', zone: 'series'),
        playCard('payment-5', 5, 'spades', zone: 'series'),
        playCard('defended-heart', 6, 'hearts', zone: 'series'),
        playCard('attack-6', 6, 'clubs', controller: 2, zone: 'series'),
        playCard('adjust-5', 5, 'clubs', controller: 2, zone: 'series'),
      ];
      Map<String, dynamic> board(int seat, int actor, {String? decision}) => {
        'seat': seat,
        'active': 2,
        'turn': 2,
        'players': players,
        'cards': cards,
        'window_id': 'negotiation',
        'required_actor': actor,
        if (decision != null) ...{
          'decision_id': 'negotiation/$decision',
          'decision_kind': 'negotiator',
          'choices': seat == actor
              ? [
                  decision == 'clubs' ? ['adjust-5'] : ['payment-5'],
                ]
              : <List<String>>[],
        },
      };
      Map<String, dynamic> online({String? stage}) => {
        if (stage != null) 'decision_stage': stage,
        'rules_context': {
          'purchase_unit_price': 1,
          'pending': {
            'type': 'attack',
            'actor': 2,
            'target_seat': 1,
            'response_types': stage == null ? ['negotiate'] : <String>[],
          },
        },
      };
      final defender = ViewClient()..publish(board(1, 1), online: online());
      addTearDown(defender.dispose);
      await showOnline(tester, defender, expand: false);
      await openPublicSeat(tester, 'negotiator');
      await doubleTapCard(tester, 'negotiator');
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(defender.submissions, [
        {
          'type': 'negotiate',
          'payload': {'pending_action_id': 'negotiation'},
        },
      ]);
      defender.publish(
        board(1, 2, decision: 'clubs'),
        online: online(stage: 'clubs'),
      );
      await tester.pumpAndSettle();
      expect(find.text('Pass response'), findsNothing);
      expect(defender.submissions, hasLength(1));
      // A second authorized viewer answers the attacker's persisted stage.
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pumpAndSettle();
      final attacker = ViewClient()
        ..publish(
          board(2, 2, decision: 'clubs'),
          online: online(stage: 'clubs'),
        );
      addTearDown(attacker.dispose);
      await showOnline(tester, attacker, expand: false);
      await openPublicSeat(tester, 'adjust-5');
      await doubleTapCard(tester, 'adjust-5');
      final oldConfirm = tester.widget<FilledButton>(confirm).onPressed!;
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(attacker.submissions, [
        {
          'type': 'decision',
          'payload': {
            'pending_action_id': 'negotiation',
            'decision_id': 'negotiation/clubs',
            'selection': ['adjust-5'],
          },
        },
      ]);
      attacker.publish(
        board(2, 1, decision: 'payment'),
        online: online(stage: 'payment'),
      );
      oldConfirm();
      await tester.pumpAndSettle();
      expect(attacker.submissions, hasLength(1));
      expect(find.text('Pass response'), findsNothing);
      await tester.pumpWidget(const SizedBox.shrink());
      await tester.pumpAndSettle();
      defender.publish(
        board(1, 1, decision: 'payment'),
        online: online(stage: 'payment'),
      );
      await showOnline(tester, defender, expand: false);
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(defender.snapshot!.board['decision_id'], 'negotiation/payment');
      expect(defender.submissions, hasLength(1));
      await tester.tap(find.text('Pending decision'));
      await tester.pumpAndSettle();
      expect(find.byKey(const ValueKey('effect-choice-0')), findsOneWidget);
      await openPublicSeat(tester, 'payment-5');
      await doubleTapCard(tester, 'payment-5');
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle();
      expect(defender.submissions.last, {
        'type': 'decision',
        'payload': {
          'pending_action_id': 'negotiation',
          'decision_id': 'negotiation/payment',
          'selection': ['payment-5'],
        },
      });
      expect(defender.submissions, hasLength(2));
      expect(find.text('Pass response'), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('same ability from two combos has distinct accessible choices', (
    tester,
  ) async {
    final board = <String, dynamic>{
      'game_id': 'g',
      'seat': 1,
      'active': 1,
      'round': 4,
      'phase': 'playing',
      'players': [
        {'seat': 1, 'history': <String>[]},
      ],
      'cards': [
        for (final combo in ['fate-a', 'fate-b'])
          for (final suit in ['hearts', 'clubs', 'spades', 'diamonds'])
            {
              ...playCard('$combo-$suit', 13, suit, zone: 'formation'),
              'allocation': combo,
            },
      ],
      'formations': [
        for (final combo in ['fate-a', 'fate-b'])
          {
            'id': combo,
            'controller': 1,
            'spec': {
              'kind': 'fate',
              'cards': [
                for (final suit in ['hearts', 'clubs', 'spades', 'diamonds'])
                  '$combo-$suit',
              ],
            },
          },
      ],
    };
    final intents = suggestCardPlays(
      board: board,
      online: const {},
      cardIds: {'fate-a-hearts', 'fate-b-hearts'},
    );
    expect(intents.map((intent) => intent.type), ['fate', 'fate']);
    final chosen = <String>[];
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: CardPlayChoices(
            intents: intents,
            label: (_) => 'Fate',
            onSelected: (intent) => chosen.add(intent.seed['formation_id']),
          ),
        ),
      ),
    );
    expect(tester.takeException(), isNull);
    expect(find.text('Fate · combo 1'), findsOneWidget);
    expect(find.text('Fate · combo 2'), findsOneWidget);
    expect(find.textContaining('fate-a'), findsNothing);
    expect(find.textContaining('fate-b'), findsNothing);
    final buttons = tester.widgetList<OutlinedButton>(
      find.byType(OutlinedButton),
    );
    expect(buttons.map((button) => button.key).toSet(), hasLength(2));
    await tester.tap(find.text('Fate · combo 1'));
    await tester.tap(find.text('Fate · combo 2'));
    expect(chosen, ['fate-a', 'fate-b']);
  });

  testWidgets(
    'exposed Kidnapper choice prepares and confirms only its outcome',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'window_id': 'kidnapper-window',
          'required_actor': 1,
          'decision_kind': 'kidnapper-outcome',
          'choices': [
            ['exposed-queen'],
          ],
          'cards': [
            playCard(
              'exposed-queen',
              12,
              'hearts',
              controller: 2,
              zone: 'unassigned',
            ),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(
        find.text('Choose an exposed royal for Kidnapper'),
        findsOneWidget,
      );
      expect(find.text('Pass response'), findsNothing);
      final choice = find.byKey(const ValueKey('effect-choice-0'));
      await tester.ensureVisible(choice);
      await tester.tap(choice);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions, isEmpty);
      final confirm = find.byKey(const ValueKey('online-submit'));
      expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
      expect(find.text('Confirm Kidnapper Outcome'), findsOneWidget);
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions.single, {
        'type': 'kidnapper-outcome',
        'payload': {
          'pending_action_id': 'kidnapper-window',
          'selection': ['exposed-queen'],
        },
      });
    },
  );

  testWidgets('Kidnapper confirmation rejects revoked and replaced choices', (
    tester,
  ) async {
    final client = ViewClient();
    addTearDown(client.dispose);
    final board = <String, dynamic>{
      'window_id': 'kidnapper-window',
      'required_actor': 1,
      'decision_kind': 'kidnapper-outcome',
      'choices': [
        ['exposed-queen'],
      ],
      'cards': [
        playCard(
          'exposed-queen',
          12,
          'hearts',
          controller: 2,
          zone: 'unassigned',
        ),
        playCard('other-king', 13, 'clubs', controller: 2, zone: 'unassigned'),
      ],
    };
    client.publish(board);
    await showOnline(tester, client, expand: false);
    final choice = find.byKey(const ValueKey('effect-choice-0'));
    await tester.ensureVisible(choice);
    await tester.tap(choice);
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    final confirm = find.byKey(const ValueKey('online-submit'));
    final queuedConfirm = tester.widget<FilledButton>(confirm).onPressed!;
    // The face stays public: current eligibility, not mere visibility, matters.
    client.publish({
      ...board,
      'choices': [
        ['other-king'],
      ],
    });
    queuedConfirm();
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(client.submissions, isEmpty);
    expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
    client.publish({...board, 'required_actor': 2});
    queuedConfirm();
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(client.submissions, isEmpty);
    expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
    final clear = find.text('Clear selection');
    await tester.ensureVisible(clear);
    await tester.tap(clear);
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(find.byKey(const ValueKey('effect-choice-0')), findsNothing);
    expect(find.text('Choose an exposed royal for Kidnapper'), findsNothing);
  });

  testWidgets(
    'non-combat responses show the pending move and supported choices',
    (tester) async {
      final client = ViewClient()
        ..publish(
          {
            'window_id': 'response',
            'required_actor': 1,
            'cards': [
              playCard('ace', 1, 'hearts', zone: 'concealed-ace'),
              playCard('negotiator', 11, 'spades', zone: 'attachment'),
            ],
          },
          online: {
            'rules_context': {
              'purchase_unit_price': 5,
              'pending': {
                'type': 'open-series',
                'actor': 2,
                'response_types': <String>[],
              },
            },
          },
        );
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.text('Seat 2 · Open Series'), findsOneWidget);
      expect(
        find.text('No card response is available. Pass to continue.'),
        findsOneWidget,
      );
      expect(find.text('Negotiate'), findsNothing);
      expect(find.text('Pass response'), findsOneWidget);
      chooseAction(tester, 'numerical-defense');
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets('queued old confirmation cannot act in a different game', (
    t,
  ) async {
    final client = ViewClient()..publish({});
    addTearDown(client.dispose);
    await showOnline(t, client, expand: false);
    await chooseGameOption(t, 'Declare Ordinary');
    final oldConfirm = t
        .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
        .onPressed!;
    client.publish({'game_id': 'replacement-game'});
    oldConfirm();
    await t.pumpAndSettle();
    expect(client.submissions, isEmpty);
  });

  testWidgets(
    'Great People protection uses the exact public formation and invalidates when it disappears',
    (tester) async {
      final cards = [
        playCard('queen-one', 12, 'hearts'),
        playCard('queen-two', 12, 'hearts'),
        {
          ...playCard('baron-king', 13, 'diamonds', zone: 'formation'),
          'allocation': 'baron',
        },
      ];
      final client = ViewClient()
        ..publish({
          'cards': cards,
          'formations': [
            {
              'id': 'baron',
              'controller': 1,
              'spec': {
                'kind': 'baron',
                'cards': ['baron-king'],
              },
            },
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await tester.tap(find.byKey(const ValueKey('adaptive-select-queen-one')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await doubleTapCard(tester, 'queen-two');
      final intent = find.byKey(
        const ValueKey('card-intent-open-formation-great-people'),
      );
      await tester.ensureVisible(intent);
      await tester.tap(intent);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      final confirm = find.byKey(const ValueKey('online-submit'));
      expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
      final protection = find.text('Choose a protection target');
      await tester.ensureVisible(protection);
      await tester.tap(protection);
      await tester.pumpAndSettle();
      await openPublicSeat(tester, 'baron-king');
      await tester.tap(
        find.byKey(const ValueKey('adaptive-select-baron-king')),
      );
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      if (find.text('Close public cards').evaluate().isNotEmpty) {
        await tester.tap(find.text('Close public cards'));
        await tester.pumpAndSettle();
      }
      await doubleTapCard(tester, 'queen-one');
      expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
      expect(
        find.text('Protection · Baron · K of Diamonds, copy 1'),
        findsOneWidget,
      );
      expect(find.text('Protection · Great People'), findsNothing);
      final queued = tester.widget<FilledButton>(confirm).onPressed!;
      client.publish({'cards': cards.take(2).toList(), 'formations': []});
      await tester.pumpAndSettle();
      expect(
        find.text('Protection · Choose your public target'),
        findsOneWidget,
      );
      expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
      queued();
      await tester.pumpAndSettle();
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    'drag Spades to the deck prepares a purchase without submitting',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 1,
          'cards': [playCard('money', 10, 'spades')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final source = find.byKey(const ValueKey('adaptive-select-money'));
      final gesture = await tester.startGesture(
        tester.getCenter(source),
        kind: PointerDeviceKind.mouse,
      );
      await gesture.moveBy(const Offset(0, -24));
      await tester.pump();
      await gesture.moveTo(
        tester.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
      );
      await tester.pump();
      await gesture.up();
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.text('Purchase'), findsOneWidget);
      expect(find.textContaining('Payment · 10 of Spades'), findsOneWidget);
      expect(client.submissions, isEmpty);
      final quantity = find.widgetWithText(TextField, 'Quantity');
      await tester.ensureVisible(quantity);
      await tester.enterText(quantity, '2');
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions.single, {
        'type': 'purchase',
        'payload': {
          'quantity': 2,
          'payment': ['money'],
        },
      });
    },
  );

  testWidgets(
    'Escape from a footer or connection utility cancels an unsubmitted pointer drop',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 1,
          'cards': [playCard('escape-money', 10, 'spades')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      for (final utility in ['Game options', 'Refresh connection']) {
        final source = find.byKey(
          const ValueKey('adaptive-select-escape-money'),
        );
        await tester.ensureVisible(source);
        final pointer = await tester.startGesture(
          tester.getCenter(source),
          kind: PointerDeviceKind.mouse,
        );
        await pointer.moveBy(const Offset(0, -24));
        await tester.pump();
        await pointer.moveTo(
          tester.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
        );
        await tester.pump();
        await pointer.up();
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        expect(find.text('Purchase'), findsOneWidget);
        expect(find.byKey(const ValueKey('contextual-popup')), findsOneWidget);
        expect(
          tester
              .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
              .selectedCardIds,
          {'escape-money'},
        );
        final icon = find
            .descendant(
              of: find.byTooltip(utility),
              matching: find.byType(Icon),
            )
            .first;
        final focus = Focus.of(tester.element(icon));
        focus.requestFocus();
        await tester.pumpAndSettle();
        expect(focus.hasPrimaryFocus, isTrue, reason: utility);
        await tester.sendKeyEvent(LogicalKeyboardKey.escape);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
        expect(
          find.byKey(const ValueKey('contextual-popup')),
          findsNothing,
          reason: utility,
        );
        expect(
          tester
              .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
              .selectedCardIds,
          isEmpty,
          reason: utility,
        );
        expect(client.submissions, isEmpty, reason: utility);
        expect(client.snapshot!.board['turn'], 1);
        expect(client.snapshot!.board['active'], 1);
        expect(tester.takeException(), isNull);
      }
    },
  );

  testWidgets(
    'Escape at the root focus clears a completed pointer-drop draft without submitting',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'turn': 1,
          'cards': [playCard('root-focus-money', 10, 'spades')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      final source = find.byKey(
        const ValueKey('adaptive-select-root-focus-money'),
      );
      final pointer = await tester.startGesture(
        tester.getCenter(source),
        kind: PointerDeviceKind.mouse,
      );
      await pointer.moveBy(const Offset(0, -24));
      await tester.pump();
      await pointer.moveTo(
        tester.getCenter(find.byKey(const ValueKey('adaptive-draw-pile'))),
      );
      await tester.pump();
      await pointer.up();
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.text('Purchase'), findsOneWidget);
      expect(find.byKey(const ValueKey('contextual-popup')), findsOneWidget);
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        {'root-focus-money'},
      );
      FocusManager.instance.rootScope.requestScopeFocus();
      await tester.pumpAndSettle();
      expect(
        FocusManager.instance.primaryFocus,
        same(FocusManager.instance.rootScope),
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.escape);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(find.byKey(const ValueKey('contextual-popup')), findsNothing);
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        isEmpty,
      );
      expect(client.submissions, isEmpty);
      expect(client.snapshot!.board['turn'], 1);
      expect(client.snapshot!.board['active'], 1);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets(
    'same-window ownership change invalidates payment without revealing history',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [playCard('money', 10, 'spades')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      await doubleTapCard(tester, 'money');
      await tester.tap(find.byKey(const ValueKey('card-intent-purchase')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await tester.enterText(find.widgetWithText(TextField, 'Quantity'), '1');
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNotNull,
      );
      client.publish({
        'cards': [
          playCard('money', 10, 'spades', controller: 2, zone: 'series'),
        ],
      });
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
            .onPressed,
        isNull,
      );
      expect(client.submissions, isEmpty);
    },
  );

  testWidgets('Doppelganger and People require explicit complete choices', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish({
        'cards': [
          playCard('king1', 13, 'clubs'),
          playCard('king2', 13, 'diamonds'),
          playCard('queen', 12, 'hearts'),
        ],
      });
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    for (final id in ['king1', 'king2']) {
      final card = find.byKey(ValueKey('adaptive-select-$id'));
      await tester.ensureVisible(card);
      await tester.pumpAndSettle();
      await tester.tap(card);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
    }
    await doubleTapCard(tester, 'queen');
    final people = find.byKey(
      const ValueKey('card-intent-open-formation-people'),
    );
    await tester.ensureVisible(people);
    await tester.tap(people);
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(
      tester
          .widget<FilledButton>(find.byKey(const ValueKey('online-submit')))
          .onPressed,
      isNull,
    );
    expect(find.text('Replacement rank'), findsOneWidget);
    expect(find.text('Replacement suit'), findsOneWidget);
    expect(client.submissions, isEmpty);
  });
  testWidgets(
    'Barricade never substitutes another Queen after its activation card is removed',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'cards': [
            {
              ...playCard('barricade', 12, 'hearts', zone: 'attachment'),
              'allocation': 'hearts',
            },
            playCard('other-queen', 12, 'hearts'),
            for (var rank = 2; rank <= 6; rank++)
              playCard('cost-$rank', rank, 'diamonds'),
          ],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      for (final id in [
        'other-queen',
        'cost-2',
        'cost-3',
        'cost-4',
        'cost-5',
      ]) {
        await tester.tap(find.byKey(ValueKey('adaptive-select-$id')));
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
      }
      await openPublicSeat(tester, 'barricade');
      await doubleTapCard(tester, 'barricade');
      final intent = find.byKey(
        const ValueKey('card-intent-barricade-sacrifice'),
      );
      if (intent.evaluate().isNotEmpty) {
        await tester.ensureVisible(intent);
        await tester.tap(intent);
        await tester.pumpAndSettle(const Duration(milliseconds: 350));
      }
      final confirm = find.byKey(const ValueKey('online-submit'));
      expect(tester.widget<FilledButton>(confirm).onPressed, isNotNull);
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .cardRoles['barricade'],
        TableCardRole.source,
      );
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .cardRoles['other-queen'],
        TableCardRole.payment,
      );
      await openPublicSeat(tester, 'barricade');
      await tester.tap(find.byKey(const ValueKey('adaptive-select-barricade')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      if (find.text('Close public cards').evaluate().isNotEmpty) {
        await tester.tap(find.text('Close public cards'));
        await tester.pumpAndSettle();
      }
      await tester.tap(find.byKey(const ValueKey('adaptive-select-cost-6')));
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      await doubleTapCard(tester, 'cost-6');
      expect(
        tester
            .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
            .selectedCardIds,
        hasLength(6),
      );
      expect(tester.widget<FilledButton>(confirm).onPressed, isNull);
      expect(client.submissions, isEmpty);
      expect(tester.takeException(), isNull);
    },
  );

  testWidgets('selecting numbers proposes a series and confirms explicitly', (
    tester,
  ) async {
    final client = ViewClient()
      ..publish({
        'cards': [
          playCard('heart2', 2, 'hearts'),
          playCard('heart4', 4, 'hearts'),
        ],
      });
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    expect(find.text('Prepare your action'), findsNothing);
    await doubleTapCard(tester, 'heart2');
    final intent = find.byKey(const ValueKey('card-intent-open-series-hearts'));
    if (intent.evaluate().isNotEmpty) {
      await tester.ensureVisible(intent);
      await tester.tap(intent);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
    }
    expect(find.textContaining('all number cards'), findsWidgets);
    expect(client.submissions, isEmpty);
    final confirm = find.byKey(const ValueKey('online-submit'));
    await tester.ensureVisible(confirm);
    await tester.tap(confirm);
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    expect(client.submissions.single['type'], 'open-series');
    expect(client.submissions.single['payload']['suit'], 'hearts');
  });

  testWidgets(
    'ordered effect choice prepares exact server choice, never passes a response',
    (tester) async {
      final client = ViewClient()
        ..publish({
          'window_id': 'w',
          'decision_id': 'd',
          'required_actor': 1,
          'decision_kind': 'dexter-prevention',
          'choices': [
            [],
            ['diamond3'],
          ],
          'cards': [playCard('diamond3', 3, 'diamonds', zone: 'series')],
        });
      addTearDown(client.dispose);
      await showOnline(tester, client, expand: false);
      expect(find.text('Pass response'), findsNothing);
      final choice = find.byKey(const ValueKey('effect-choice-0'));
      await tester.ensureVisible(choice);
      await tester.tap(choice);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions, isEmpty);
      final confirm = find.byKey(const ValueKey('online-submit'));
      await tester.ensureVisible(confirm);
      await tester.tap(confirm);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
      expect(client.submissions.single, {
        'type': 'decision',
        'payload': {
          'pending_action_id': 'w',
          'decision_id': 'd',
          'selection': <String>[],
        },
      });
    },
  );

  testWidgets('a changed response invalidates a prepared card move', (
    tester,
  ) async {
    final cards = [playCard('ace', 1, 'hearts', zone: 'concealed-ace')];
    final client = ViewClient()..publish({'cards': cards});
    addTearDown(client.dispose);
    await showOnline(tester, client, expand: false);
    await doubleTapCard(tester, 'ace');
    final intent = find.byKey(const ValueKey('card-intent-compensation'));
    if (intent.evaluate().isNotEmpty) {
      await tester.ensureVisible(intent);
      await tester.tap(intent);
      await tester.pumpAndSettle(const Duration(milliseconds: 350));
    }
    client.publish({
      'cards': cards,
      'window_id': 'new-window',
      'required_actor': 2,
    });
    await tester.pumpAndSettle(const Duration(milliseconds: 350));
    final button = tester.widget<FilledButton>(
      find.byKey(const ValueKey('online-submit')),
    );
    expect(button.onPressed, isNull);
    expect(client.submissions, isEmpty);
    expect(
      tester
          .widget<CgmsAdaptiveTable>(find.byType(CgmsAdaptiveTable))
          .selectedCardIds,
      contains('ace'),
    );
  });
}

Future<void> doubleTapCard(WidgetTester tester, String id) async {
  final card = find.byKey(ValueKey('adaptive-select-$id'));
  await tester.ensureVisible(card);
  await tester.tap(card);
  await tester.pump(const Duration(milliseconds: 80));
  await tester.tap(card);
  await tester.pumpAndSettle(const Duration(milliseconds: 350));
}

class CardConnectionClient extends ViewClient {
  void setConnection(bool value) {
    connected = value;
    notifyListeners();
  }
}

Future<void> dragCard(WidgetTester tester, String id, String target) async {
  final pointer = await tester.startGesture(
    tester.getCenter(find.byKey(ValueKey('adaptive-select-$id'))),
    kind: PointerDeviceKind.mouse,
  );
  await pointer.moveBy(const Offset(0, -25));
  await tester.pump();
  await pointer.moveTo(tester.getCenter(find.byKey(ValueKey(target))));
  await tester.pump();
  await pointer.up();
  await tester.pumpAndSettle();
}
