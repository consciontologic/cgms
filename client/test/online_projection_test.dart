import 'package:flutter_test/flutter_test.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:cgms_demo/online/online_client.dart';
import 'package:cgms_demo/online/projection.dart';

void main() {
  test('public bot seats stay labelled after restoring a match projection', () {
    AuthoritativeSnapshot snapshot(Map<String, dynamic> online) =>
        AuthoritativeSnapshot.fromJson({
          'version': 1,
          'cursor': 1,
          'projection': {
            'board': {
              'game_id': 'bot-match',
              'seat': 1,
              'phase': 'playing',
              'round': 1,
              'active': 2,
              'players': [
                for (final seat in [1, 2, 3])
                  {'seat': seat, 'hand_count': 5, 'ace_count': 1},
              ],
              'cards': [],
            },
            'online': online,
          },
        });
    final bots = projectTable(
      snapshot({
        'bot_seats': [2, 3],
        'bot_difficulty': 'beginner',
      }),
    );
    expect(bots.seats.map((seat) => seat.name), ['Seat 1', 'Bot 2', 'Bot 3']);
    expect(bots.turnLabel, 'Bot 2’s turn');
    expect(bots.hand, isEmpty);
    expect(
      bots.seats
          .skip(1)
          .every((seat) => seat.exposed.isEmpty && !seat.financesKnown),
      isTrue,
    );
    final people = projectTable(snapshot({}));
    expect(people.seats.map((seat) => seat.name), [
      'Seat 1',
      'Seat 2',
      'Seat 3',
    ]);
    expect(people.turnLabel, 'Seat 2’s turn');
  });

  test(
    'viewer projection keeps authorized private cards and exact balances',
    () {
      final snapshot = AuthoritativeSnapshot.fromJson({
        'version': 2,
        'cursor': 2,
        'projection': {
          'board': {
            'game_id': 'g1',
            'seat': 1,
            'phase': 'playing',
            'round': 3,
            'active': 2,
            'players': [
              {
                'seat': 1,
                'hand_count': 1,
                'ace_count': 0,
                'confined': false,
                'departed': false,
              },
              {
                'seat': 2,
                'hand_count': 8,
                'ace_count': 2,
                'confined': true,
                'departed': false,
              },
            ],
            'cards': [
              {
                'card': {
                  'id': 'opaque-own',
                  'rank': 12,
                  'suit': 'hearts',
                  'deck': 2,
                },
                'zone': 'hand',
                'controller': 1,
                'available_from_round': 4,
              },
              {
                'card': {
                  'id': 'opaque-public',
                  'rank': 8,
                  'suit': 'diamonds',
                  'deck': 1,
                },
                'zone': 'series',
                'controller': 2,
                'available_from_round': 0,
              },
            ],
          },
          'online': <String, dynamic>{
            'games_per_match': 3,
            'own_match_score': {
              'numerator': '9007199254740993',
              'denominator': '3',
            },
          },
          'score': {'numerator': '11', 'denominator': '2'},
          'cash': {'numerator': '1', 'denominator': '3'},
          'debts': [
            {
              'debtor': 0,
              'creditor': 1,
              'remaining': {'numerator': '1', 'denominator': '3'},
            },
            {
              'debtor': 2,
              'creditor': 0,
              'remaining': {'numerator': '10', 'denominator': '1'},
            },
          ],
        },
      });
      final table = projectTable(snapshot);
      expect(table.seats.first.finances.debt.format(), '1/3');
      expect(table.hand.single.id, 'opaque-own');
      expect(table.hand.single.copy, 2);
      expect(table.hand.single.status, contains('4'));
      expect(table.seats[1].exposed.single.id, 'opaque-public');
      expect(table.seats[1].concealed.handCount, 8);
      expect(table.seats.first.finances.available.format(), '1/3');
      expect(
        table.seats.first.finances.matchScore.format(),
        '3002399751580331',
      );
    },
  );
  test(
    'Coup availability respects own history, alliance and physical restriction',
    () {
      Map<String, dynamic> value = {
        'version': 1,
        'cursor': 1,
        'projection': {
          'board': {
            'game_id': 'g',
            'seat': 1,
            'phase': 'playing',
            'round': 4,
            'players': [
              {
                'seat': 1,
                'history': ['clubs'],
                'departed': false,
                'confined': true,
              },
            ],
            'cards': [
              for (final suit in ['hearts', 'diamonds'])
                for (var copy = 1; copy <= 2; copy++)
                  {
                    'controller': 1,
                    'available_from_round': 4,
                    'card': {'rank': 12, 'suit': suit},
                  },
            ],
          },
          'online': <String, dynamic>{
            'alliances': {'1': 0},
          },
        },
      };
      expect(canDeclareCoup(AuthoritativeSnapshot.fromJson(value)), isTrue);
      ((value['projection'] as Map)['online'] as Map)['server_pending'] = true;
      expect(
        canDeclareCoup(AuthoritativeSnapshot.fromJson(value)),
        isTrue,
        reason:
            'Server readiness must not remove the immediate Coup exception.',
      );
      ((value['projection'] as Map)['online'] as Map)['automatic_pending'] =
          true;
      expect(
        canDeclareCoup(AuthoritativeSnapshot.fromJson(value)),
        isTrue,
        reason:
            'Automatic work is admitted at its authoritative atomic boundary; client hints cannot hide Coup.',
      );
      ((value['projection'] as Map)['online'] as Map)['automatic_pending'] =
          false;
      ((value['projection'] as Map)['online'] as Map).remove('alliances');
      expect(
        canDeclareCoup(AuthoritativeSnapshot.fromJson(value)),
        isTrue,
        reason:
            'Missing optional presentation metadata is not evidence of an alliance.',
      );
      ((value['projection'] as Map)['online'] as Map)['alliances'] = {'1': 2};
      expect(canDeclareCoup(AuthoritativeSnapshot.fromJson(value)), isFalse);
      ((value['projection'] as Map)['online'] as Map)['alliances'] = {'1': 0};
      (((value['projection'] as Map)['board'] as Map)['cards'] as List)
              .first['available_from_round'] =
          5;
      expect(canDeclareCoup(AuthoritativeSnapshot.fromJson(value)), isFalse);
    },
  );

  test(
    'public controller styles follow captured physical cards for every viewer',
    () {
      final players = [
        for (var seat = 1; seat <= 4; seat++)
          {
            'seat': seat,
            'hand_count': 8,
            'ace_count': 2,
            'card_style': CardStyle.values[seat - 1].slug,
          },
      ];
      final public = {
        'card': {
          'id': 'captured-card',
          'rank': 8,
          'suit': 'diamonds',
          'deck': 2,
        },
        'controller': 2,
        'zone': 'series',
        'available_from_round': 0,
      };
      final ownHand = {
        'card': {'id': 'own-card', 'rank': 12, 'suit': 'hearts', 'deck': 1},
        'controller': 1,
        'zone': 'hand',
        'available_from_round': 0,
      };
      TableViewData project(int viewer, {bool includePrivate = false}) =>
          projectTable(
            AuthoritativeSnapshot.fromJson({
              'version': 3,
              'cursor': 3,
              'projection': {
                'board': {
                  'game_id': 'style-game',
                  'seat': viewer,
                  'phase': 'playing',
                  'round': 3,
                  'active': 1,
                  'players': players,
                  'cards': [public, if (includePrivate) ownHand],
                  'formations': [
                    {
                      'id': 'captured-series',
                      'controller': public['controller'],
                      'spec': {
                        'kind': 'number-series',
                        'cards': ['captured-card'],
                      },
                    },
                  ],
                },
                'online': <String, dynamic>{},
              },
            }),
          );
      for (var controller = 1; controller <= 4; controller++) {
        public['controller'] = controller;
        for (var viewer = 1; viewer <= 4; viewer++) {
          final table = project(viewer, includePrivate: viewer == 1);
          final card = table.seats[controller - 1].exposed.single;
          expect(card.style, CardStyle.values[controller - 1]);
          expect(card.id, 'captured-card');
          expect(card.copy, 2);
          expect(card.label, '8 of Diamonds, copy 2');
          expect(
            table.seats[controller - 1].combos.single.cards.single.style,
            card.style,
          );
          expect(
            table.hand.map((c) => c.id),
            viewer == 1 ? ['own-card'] : isEmpty,
          );
          if (viewer == 1)
            expect(table.hand.single.style, CardStyle.firstLight);
          expect(table.seats[1].concealed.handCount, 8);
          expect(table.seats[1].concealed.aceCount, 2);
        }
      }
      players[3]['card_style'] = '../../concealed/identity';
      expect(project(1).seats[3].exposed.single.style, CardStyle.firstLight);
      players[3].remove('card_style');
      expect(project(1).seats[3].exposed.single.style, CardStyle.firstLight);
      players[0]['card_style'] = 'ember-glaze';
      expect(
        project(1, includePrivate: true).hand.single.style,
        CardStyle.emberGlaze,
      );
      expect(
        projectCard({...public, 'controller': 0}).style,
        CardStyle.firstLight,
      );
    },
  );

  test(
    'attached royal display follows public allocation and preserves its face',
    () {
      final royal = {
        'card': {'id': 'king', 'rank': 13, 'suit': 'spades', 'deck': 1},
        'controller': 1,
        'zone': 'attachment',
        'allocation': 'diamonds',
      };
      final projected = projectCard(royal);
      expect(projected.suit, CardSuit.spades);
      expect(projected.seriesSuit, CardSuit.diamonds);
      expect(projectCard({...royal, 'zone': 'hand'}).seriesSuit, isNull);
      expect(
        projectCard({...royal, 'allocation': 'invalid'}).seriesSuit,
        isNull,
      );
      final number = projectCard({
        'card': {'id': 'h8', 'rank': 8, 'suit': 'hearts', 'deck': 1},
        'controller': 1,
        'zone': 'series',
      });
      expect(number.seriesSuit, CardSuit.hearts);
    },
  );
}
