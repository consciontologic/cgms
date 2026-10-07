import 'package:cgms_demo/online/card_intents.dart';
import 'package:cgms_demo/online/card_play_choices.dart';
import 'package:flutter/material.dart';
import 'package:cgms_demo/online/command_contract.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter_test/flutter_test.dart';

Map<String, dynamic> card(
  String id,
  String suit,
  int rank, {
  int owner = 1,
  String zone = 'hand',
  String allocation = '',
  int available = 0,
  int used = 0,
}) => {
  'card': {'id': id, 'suit': suit, 'rank': rank, 'deck': 1},
  'controller': owner,
  'zone': zone,
  'allocation': allocation,
  'available_from_round': available,
  'used_turn': used,
};

Map<String, dynamic> board(List<Map<String, dynamic>> cards) => {
  'game_id': 'g',
  'seat': 1,
  'active': 1,
  'round': 3,
  'turn': 7,
  'phase': 'playing',
  'players': [
    for (var seat = 1; seat <= 4; seat++)
      {
        'seat': seat,
        'history': ['diamonds', 'clubs'],
      },
  ],
  'cards': cards,
  'formations': <Map<String, dynamic>>[],
};

Map<String, dynamic> formation(String id, String kind, List<String> cards) => {
  'id': id,
  'controller': 1,
  'spec': {'kind': kind, 'cards': cards},
};

List<CardPlayIntent> plays(
  Map<String, dynamic> b,
  Set<String> ids, {
  TableDropTarget? target,
  Map<String, dynamic> online = const {},
}) => suggestCardPlays(board: b, online: online, cardIds: ids, target: target);

void main() {
  test('self sacrifice cannot be inferred from an unrelated foreign drop', () {
    final b = board([
      card('king', 'diamonds', 13, zone: 'attachment', allocation: 'hearts'),
      card('h', 'hearts', 5, zone: 'series'),
    ]);
    expect(
      suggestCardDrops(
        board: b,
        online: {},
        cardIds: {'king'},
        target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
      ),
      isEmpty,
    );
    expect(
      suggestCardDrops(
        board: b,
        online: {},
        cardIds: {'king'},
        target: const TableDropTarget(zone: TableDropZone.seat, seat: 1),
      ).single.type,
      'ringleader',
    );
  });
  test('royal attachment highlights only its legal supported role', () {
    final b = board([
      card('q', 'clubs', 12),
      card('h', 'hearts', 5, zone: 'series'),
      card('c', 'clubs', 5, zone: 'series'),
    ]);
    expect(royalAttachmentSuits(b, {'q'}), {'clubs'});
    expect(
      suggestCardDrops(
        board: b,
        online: {},
        cardIds: {'q'},
        target: const TableDropTarget(
          zone: TableDropZone.series,
          seat: 1,
          suit: 'hearts',
        ),
      ),
      isEmpty,
    );
  });
  test('unequal physical attack drop submits no inferred move', () {
    final b = board([
      card('c', 'clubs', 4, zone: 'series'),
      card('d', 'diamonds', 3, zone: 'series'),
      card('enemy', 'hearts', 5, zone: 'series', owner: 2),
    ]);
    expect(
      suggestCardDrops(
        board: b,
        online: {},
        cardIds: {'c'},
        target: const TableDropTarget(
          zone: TableDropZone.card,
          seat: 2,
          cardId: 'enemy',
        ),
      ),
      isEmpty,
    );
  });

  test(
    'drops require current Ace support and reject duplicate active Inflation',
    () {
      final b = board([card('ace', 'spades', 1, zone: 'concealed-ace')]);
      List<CardPlayIntent> dropped() => suggestCardDrops(
        board: b,
        online: {},
        cardIds: {'ace'},
        target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
      );
      expect(dropped(), isEmpty);
      (b['cards'] as List).addAll(<Map<String, dynamic>>[
        card('c', 'clubs', 5, zone: 'series'),
        card('h', 'hearts', 5, zone: 'series'),
      ]);
      expect(dropped().any((i) => i.type == 'main-inflation'), isTrue);
      b['effects'] = [
        {'kind': 'inflation', 'target': 2},
      ];
      expect(dropped().any((i) => i.type == 'main-inflation'), isFalse);
    },
  );
  test(
    'a restricted unselected copy does not prevent exact historical addition',
    () {
      final b = board([
        card('d1', 'diamonds', 4),
        card('d2', 'diamonds', 4, available: 5),
      ]);
      final result = plays(b, {
        'd1',
      }, target: const TableDropTarget(zone: TableDropZone.seat, seat: 1));
      expect(result.single.seed['selection'], ['d1']);
    },
  );

  test(
    'intact exposed recipient drop only proposes that loan, never attack',
    () {
      final b = board([
        card('jc1', 'clubs', 11, zone: 'formation', allocation: 'ponzi'),
        card('jc2', 'clubs', 11, zone: 'formation', allocation: 'ponzi'),
        card('js1', 'spades', 11, zone: 'formation', allocation: 'ponzi'),
        card('js2', 'spades', 11, zone: 'formation', allocation: 'ponzi'),
      ]);
      b['formations'] = [
        formation('ponzi', 'ponzi', ['jc1', 'jc2', 'js1', 'js2']),
      ];
      const recipient = TableDropTarget(zone: TableDropZone.seat, seat: 2);
      final result = plays(b, {'jc1', 'jc2', 'js1', 'js2'}, target: recipient);
      expect(result.map((i) => i.type), ['offer']);
      (b['cards'] as List).add(
        card('target-heart', 'hearts', 5, owner: 2, zone: 'series'),
      );
      for (final nested in [
        const TableDropTarget(
          zone: TableDropZone.series,
          seat: 2,
          suit: 'hearts',
        ),
        const TableDropTarget(
          zone: TableDropZone.card,
          seat: 2,
          cardId: 'target-heart',
        ),
      ]) {
        final nestedResult = suggestCardDrops(
          board: b,
          online: {},
          cardIds: {'jc1', 'jc2', 'js1', 'js2'},
          target: nested,
        );
        expect(nestedResult.map((i) => i.type), ['offer']);
        expect(nestedResult.single.seed['terms']['loan'], isTrue);
      }
      expect(result.single.seed['terms'], {
        'to': 2,
        'give': ['jc1', 'jc2', 'js1', 'js2'],
        'receive': <String>[],
        'loan': true,
      });
      expect(
        plays(
          b,
          {'jc1', 'jc2', 'js1', 'js2'},
          target: recipient,
          online: {
            'alliances': {'1': 3, '2': 0},
          },
        ),
        isEmpty,
      );
      (b['players'] as List)[1]['confined'] = true;
      expect(
        plays(b, {'jc1', 'jc2', 'js1', 'js2'}, target: recipient),
        isEmpty,
      );
    },
  );

  test('historical addition preserves selected physical duplicate only', () {
    final b = board([card('d1', 'diamonds', 4), card('d2', 'diamonds', 4)]);
    final result = plays(b, {
      'd2',
    }, target: const TableDropTarget(zone: TableDropZone.seat, seat: 1));
    expect(result.single.seed['selection'], ['d2']);
  });

  test(
    'unused Ace disposal follows disclosure timing rather than ability timing',
    () {
      final original = board([
        card('ace', 'hearts', 1, zone: 'concealed-ace', available: 4),
      ]);
      for (final b in [
        {...original, 'active': 2},
        {
          ...original,
          'players': [
            ...List<Map<String, dynamic>>.from(
              original['players'] as List,
            ).map((p) => {...p, if (p['seat'] == 1) 'confined': true}),
          ],
        },
        {
          ...original,
          'window_id': 'pending',
          'decision_id': 'decision',
          'decision_kind': 'justice',
          'required_actor': 2,
        },
      ]) {
        final result = plays(
          b,
          {'ace'},
          online: {
            'rules_context': {'ace_used_this_round': true},
          },
        );
        expect(result.map((i) => i.type), ['expose-unused-ace']);
        expect(result.single.seed, {'ace': 'ace'});
        expect(draggableCardIds(b, {}), isEmpty);
      }
      for (final blocked in [
        'server_pending',
        'automatic_pending',
        'round_closed',
        'departure_pending',
      ]) {
        expect(
          plays(original, {'ace'}, online: {blocked: true}),
          isEmpty,
          reason: blocked,
        );
      }
      expect(plays({...original, 'phase': 'settlement'}, {'ace'}), isEmpty);
      expect(
        plays(
          {
            ...original,
            'players': [
              {'seat': 1, 'departed': true},
            ],
          },
          {'ace'},
        ),
        isEmpty,
      );
    },
  );

  test(
    'Kidnapper outcome selects only the exact current public offered card',
    () {
      final b =
          board([
            card('queen', 'hearts', 12, owner: 2, zone: 'unassigned'),
            card('king', 'clubs', 13, owner: 2, zone: 'unassigned'),
            card('hidden', 'spades', 12, owner: 2),
          ])..addAll({
            'window_id': 'kidnap',
            'required_actor': 1,
            'decision_kind': 'kidnapper-outcome',
            'choices': [
              ['queen'],
            ],
          });
      expect(selectableCardIds(b, {}), {'queen'});
      expect(draggableCardIds(b, {}), isEmpty);
      final intent = plays(b, {'queen'}).single;
      expect(intent.type, 'kidnapper-outcome');
      expect(intent.seed, {
        'pending_action_id': 'kidnap',
        'selection': ['queen'],
      });
      expect(plays(b, {'king'}), isEmpty);
      expect(plays(b, {'queen', 'king'}), isEmpty);
      expect(plays({...b, 'required_actor': 2}, {'queen'}), isEmpty);
      expect(
        plays(
          {
            ...b,
            'choices': [
              ['hidden'],
            ],
          },
          {'hidden'},
        ),
        isEmpty,
      );
      expect(plays({...b, 'choices': <List<String>>[]}, {'queen'}), isEmpty);
    },
  );

  test('number drop prepares a suit action and explains all held numbers', () {
    final b = board([card('h3', 'hearts', 3), card('h8', 'hearts', 8)]);
    final intent = plays(b, {'h3'}).singleWhere((i) => i.type == 'open-series');
    expect(intent.seed, {
      'suit': 'hearts',
      'selection': ['h3', 'h8'],
    });
    expect(intent.explanation, contains('all'));
    expect(intent.explanation, contains('hand'));
  });

  test('purchase retains chosen payment but never invents quantity', () {
    final b = board([card('s9', 'spades', 9)]);
    final result = plays(b, {
      's9',
    }, target: const TableDropTarget(zone: TableDropZone.drawPile));
    expect(result.map((i) => i.type), ['purchase']);
    expect(result.single.seed, {
      'payment': ['s9'],
    });
  });

  test(
    'foreign hidden stale restricted and active-effect cards cannot source plays',
    () {
      final b = board([
        card('mine', 'spades', 5),
        card('enemy', 'spades', 5, owner: 2, zone: 'series'),
        card('hidden', 'clubs', 8, owner: 2),
        card('later', 'hearts', 9, available: 4),
        card('effect', 'spades', 1, zone: 'active-ace'),
        card('draw', 'clubs', 4, owner: 0, zone: 'draw'),
      ]);
      expect(draggableCardIds(b, {}), {'mine'});
      for (final id in ['enemy', 'hidden', 'later', 'effect', 'draw', 'gone']) {
        expect(plays(b, {'mine', id}), isEmpty, reason: id);
      }
    },
  );

  test(
    'one restricted number blocks the whole suit opening, not valid payment',
    () {
      final b = board([
        card('s5', 'spades', 5),
        card('s9', 'spades', 9, available: 4),
      ]);
      expect(plays(b, {'s5'}).any((i) => i.type == 'open-series'), isFalse);
      expect(plays(b, {'s5'}).any((i) => i.type == 'purchase'), isTrue);
    },
  );

  test(
    'closed automatic confined and non-own-turn contexts do not open cards',
    () {
      final b = board([card('h3', 'hearts', 3)]);
      for (final state in [
        {...b, 'phase': 'settlement'},
        {...b, 'active': 2},
        {...b, 'window_id': 'w', 'required_actor': 2},
        {
          ...b,
          'players': [
            {'seat': 1, 'confined': true},
          ],
        },
      ]) {
        expect(plays(state, {'h3'}), isEmpty);
      }
      expect(plays(b, {'h3'}, online: {'server_pending': true}), isEmpty);
      expect(plays(b, {'h3'}, online: {'automatic_pending': true}), isEmpty);
    },
  );

  test('royal attachment uses destination series rather than royal suit', () {
    final b = board([
      card('ks', 'spades', 13),
      card('h5', 'hearts', 5, zone: 'series'),
    ]);
    final result = plays(
      b,
      {'ks'},
      target: const TableDropTarget(
        zone: TableDropZone.card,
        seat: 1,
        cardId: 'h5',
      ),
    );
    expect(result.singleWhere((i) => i.type == 'attach').seed, {
      'selection': ['ks'],
      'suit': 'hearts',
    });
  });

  test('natural combo and Doppelganger role remain distinct drafts', () {
    final b = board([
      card('qh', 'hearts', 12),
      card('k1', 'clubs', 13),
      card('k2', 'spades', 13),
    ]);
    final people = plays(b, {
      'qh',
      'k1',
      'k2',
    }).singleWhere((i) => i.seed['formation']?['kind'] == 'people');
    expect(people.seed['formation']['cards'], ['qh', 'k1', 'k2']);
    expect(people.seed['formation'].containsKey('substitute'), isFalse);
    expect(people.explanation, contains('Doppelganger'));
    expect(people.needsSubstitution, isTrue);
    expect(people.needsProtection, isTrue);
    final pair = plays(b, {'k1', 'k2'});
    expect(pair.any((i) => i.type == 'open-formation'), isTrue);
    expect(
      pair.every((i) => i.seed['formation']?['kind'] != 'doppelganger'),
      isTrue,
    );
    expect(
      pair.singleWhere((i) => i.type == 'open-formation').needsSubstitution,
      isTrue,
    );
    expect(pair.singleWhere((i) => i.type == 'offer').seed['terms']['give'], [
      'k1',
      'k2',
    ]);
  });

  test('great People requires an explicit protection target', () {
    final b = board([card('q1', 'hearts', 12), card('q2', 'hearts', 12)]);
    final intent = plays(b, {
      'q1',
      'q2',
    }).singleWhere((i) => i.type == 'open-formation');
    expect(intent.seed['formation'], {
      'kind': 'great-people',
      'cards': ['q1', 'q2'],
    });
    expect(intent.explanation, contains('protection'));
    expect(intent.needsProtection, isTrue);
    expect(intent.needsSubstitution, isFalse);
  });

  test('extending an own combo retains its current formation identity', () {
    final b = board([
      for (final suit in ['hearts', 'diamonds', 'clubs', 'spades'])
        card('k$suit', suit, 13, zone: 'formation', allocation: 'current'),
      card('extra', 'clubs', 13),
    ]);
    b['formations'] = [
      formation('current', 'fate', [
        'khearts',
        'kdiamonds',
        'kclubs',
        'kspades',
      ]),
    ];
    final ids = {'khearts', 'kdiamonds', 'kclubs', 'kspades', 'extra'};
    final intent = plays(
      b,
      ids,
    ).singleWhere((i) => i.seed['formation']?['kind'] == 'underground');
    expect(intent.seed['formation_id'], 'current');
    expect(intent.seed['formation']['cards'], ids.toList());
    expect(draggableCardIds(b, {}), ids);
  });

  test('enemy formation is not an ordinary attack target', () {
    final b = board([card('c8', 'clubs', 8, zone: 'series')]);
    b['formations'] = [
      {...formation('enemy', 'fate', []), 'controller': 2},
    ];
    expect(
      plays(
        b,
        {'c8'},
        target: const TableDropTarget(
          zone: TableDropZone.formation,
          seat: 2,
          formationId: 'enemy',
        ),
      ),
      isEmpty,
    );
  });

  test(
    'taking a formation back uses formation identity not a partial subset',
    () {
      final b = board([
        card('j1', 'clubs', 11, zone: 'formation', allocation: 'kid'),
        card('j2', 'clubs', 11, zone: 'formation', allocation: 'kid'),
      ]);
      b['formations'] = [
        formation('kid', 'kidnapper', ['j1', 'j2']),
      ];
      final intent = plays(
        b,
        {'j1'},
        target: const TableDropTarget(zone: TableDropZone.hand),
      ).singleWhere((i) => i.type == 'take-back');
      expect(intent.seed, {'formation_id': 'kid'});
    },
  );

  test(
    'whole-series destination selects numbers explicitly without aliasing its attachments',
    () {
      final b = board([
        card('c8', 'clubs', 8, zone: 'series'),
        card(
          'royal-first',
          'spades',
          13,
          zone: 'attachment',
          allocation: 'hearts',
          owner: 2,
        ),
        card('h3', 'hearts', 3, zone: 'series', owner: 2),
        card('h5', 'hearts', 5, zone: 'series', owner: 2),
      ]);
      final result = plays(
        b,
        {'c8'},
        target: const TableDropTarget(
          zone: TableDropZone.series,
          seat: 2,
          suit: 'hearts',
        ),
      );
      expect(result.singleWhere((intent) => intent.type == 'attack').seed, {
        'selection': ['c8'],
        'targets': ['h3', 'h5'],
        'suit': 'hearts',
      });
      expect(
        plays(
          b,
          {'c8'},
          target: const TableDropTarget(
            zone: TableDropZone.card,
            seat: 2,
            cardId: 'h3',
          ),
        ).singleWhere((intent) => intent.type == 'attack').seed['targets'],
        ['h3'],
      );
      expect(
        plays(
          b,
          {'c8'},
          target: const TableDropTarget(
            zone: TableDropZone.series,
            seat: 2,
            suit: 'spades',
          ),
        ),
        isEmpty,
      );
    },
  );

  test(
    'ordinary target stays selected while Baron uses the entire target series',
    () {
      final b = board([
        card('c8', 'clubs', 8, zone: 'series'),
        card('h3', 'hearts', 3, zone: 'series', owner: 2),
        card('h5', 'hearts', 5, zone: 'series', owner: 2),
      ]);
      b['formations'] = [formation('baron', 'baron', [])];
      final result = plays(
        b,
        {'c8'},
        target: const TableDropTarget(
          zone: TableDropZone.card,
          seat: 2,
          cardId: 'h3',
        ),
      );
      expect(result.singleWhere((i) => i.type == 'attack').seed, {
        'selection': ['c8'],
        'targets': ['h3'],
        'suit': 'hearts',
      });
      expect(
        result.singleWhere((i) => i.type == 'baron-attack').seed['targets'],
        ['h3', 'h5'],
      );
      expect(
        plays(
          b,
          {'c8'},
          target: const TableDropTarget(
            zone: TableDropZone.card,
            seat: 2,
            cardId: 'missing',
          ),
        ),
        isEmpty,
      );
    },
  );

  test('Bomb and used Clubs cannot be treated as ordinary attack cards', () {
    final b = board([
      card('c8', 'clubs', 8, zone: 'series'),
      card('used', 'clubs', 6, zone: 'series', used: 7),
      card(
        'bomb',
        'hearts',
        11,
        owner: 2,
        zone: 'attachment',
        allocation: 'spades',
      ),
    ]);
    final target = const TableDropTarget(
      zone: TableDropZone.card,
      seat: 2,
      cardId: 'bomb',
    );
    final result = plays(b, {'c8'}, target: target);
    expect(result.map((i) => i.type), ['bomb-attack']);
    expect(result.single.seed['suit'], 'spades');
    expect(plays(b, {'used'}, target: target), isEmpty);
  });

  test(
    'multiple Justice formations retain distinct physical source identities',
    () {
      final b =
          board([
              card(
                'queen-a',
                'hearts',
                12,
                zone: 'formation',
                allocation: 'justice-a',
              ),
              card(
                'queen-b',
                'hearts',
                12,
                zone: 'formation',
                allocation: 'justice-b',
              ),
            ])
            ..['formations'] = [
              formation('justice-a', 'justice', ['queen-a']),
              formation('justice-b', 'justice', ['queen-b']),
            ];
      final intents = plays(b, {
        'queen-a',
        'queen-b',
      }).where((intent) => intent.type == 'justice').toList();
      expect(intents, hasLength(2));
      expect(intents.map((intent) => intent.sourceFormationId), [
        'justice-a',
        'justice-b',
      ]);
      expect(
        intents.every((intent) => !intent.seed.containsKey('formation_id')),
        isTrue,
      );
      expect(
        intents.every((intent) => !intent.seed.containsKey('ace')),
        isTrue,
        reason: 'Choosing a formation never silently spends one of its Queens',
      );
    },
  );

  test(
    'Justice selected queen uses ace payload and Ponzi value is target seat',
    () {
      final b = board([
        card('qh', 'hearts', 12, zone: 'formation', allocation: 'justice'),
        card('jc', 'clubs', 11, zone: 'formation', allocation: 'ponzi'),
      ]);
      b['formations'] = [
        formation('justice', 'justice', ['qh']),
        formation('ponzi', 'ponzi', ['jc']),
      ];
      expect(plays(b, {'qh'}).singleWhere((i) => i.type == 'justice').seed, {
        'ace': 'qh',
      });
      expect(
        plays(
          b,
          {'jc'},
          target: const TableDropTarget(zone: TableDropZone.seat, seat: 3),
        ).singleWhere((i) => i.type == 'ponzi').seed,
        {'value': 3},
      );
    },
  );

  test('Barricade cost queen is first independent of selection order', () {
    final b = board([
      card('qh', 'hearts', 12, zone: 'attachment', allocation: 'hearts'),
      for (var i = 2; i <= 6; i++) card('n$i', 'clubs', i),
    ]);
    final result = plays(b, {'n2', 'n3', 'qh', 'n4', 'n5', 'n6'});
    expect(result.singleWhere((i) => i.type == 'barricade-sacrifice').seed, {
      'selection': ['qh', 'n2', 'n3', 'n4', 'n5', 'n6'],
    });
  });

  test(
    'ace intent is ambiguous and never exposes or spends an ace automatically',
    () {
      final b = board([card('as', 'spades', 1, zone: 'concealed-ace')]);
      final result = plays(b, {
        'as',
      }, target: const TableDropTarget(zone: TableDropZone.seat, seat: 2));
      expect(result.any((i) => i.type == 'main-inflation'), isTrue);
      expect(result.any((i) => i.type == 'attack'), isTrue);
      expect(result.any((i) => i.type == 'expose-unused-ace'), isFalse);
      expect(result.singleWhere((i) => i.type == 'main-inflation').seed, {
        'ace': 'as',
        'target_seat': 2,
      });
    },
  );

  test(
    'responses bind the current window and preserve named/numerical choices',
    () {
      final b = board([card('ad', 'diamonds', 1, zone: 'concealed-ace')]);
      b.addAll({'active': 2, 'window_id': 'pending', 'required_actor': 1});
      final result = plays(
        b,
        {'ad'},
        online: {
          'rules_context': {
            'combat': {'defender': 1, 'comparison': 'equal'},
          },
        },
      );
      expect(result.map((i) => i.type), [
        'expose-unused-ace',
        'confinement',
        'numerical-defense',
      ]);
      expect(
        result
            .where((i) => i.type != 'expose-unused-ace')
            .every((i) => i.seed['pending_action_id'] == 'pending'),
        isTrue,
      );
      expect(result.every((i) => !i.seed.containsKey('value')), isTrue);
      expect(result.first.seed, {'ace': 'ad'});
      expect(plays({...b, 'required_actor': 3}, {'ad'}).map((i) => i.type), [
        'expose-unused-ace',
      ]);
      expect(
        plays(
          b,
          {'ad'},
          online: {
            'rules_context': {'ace_used_this_round': true},
          },
        ).map((i) => i.type),
        ['expose-unused-ace'],
      );
    },
  );

  test(
    'pending response drops use the actual actor or affected target only',
    () {
      final b = board([
        card('ad', 'diamonds', 1, zone: 'concealed-ace'),
        card('h8', 'hearts', 8, zone: 'series', allocation: 'hearts'),
        card('d8', 'diamonds', 8, zone: 'series', allocation: 'diamonds'),
        card('c8', 'clubs', 8, owner: 2, zone: 'series', allocation: 'clubs'),
      ])..addAll({'active': 2, 'window_id': 'w', 'required_actor': 1});
      final online = {
        'rules_context': {
          'pending': {
            'actor': 2,
            'target_seat': 1,
            'response_types': ['confinement', 'numerical-defense'],
          },
          'combat': {
            'actor': 2,
            'defender': 1,
            'comparison': 'equal',
            'target_cards': ['h8'],
          },
        },
      };
      expect(
        plays(
          {...b, 'active': 4},
          {'ad'},
          online: {
            'rules_context': {
              'combat': {'actor': 2, 'defender': 1, 'comparison': 'equal'},
            },
          },
          target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
        ).map((i) => i.type),
        ['confinement'],
      );
      for (final target in const [
        TableDropTarget(zone: TableDropZone.drawPile),
        TableDropTarget(zone: TableDropZone.hand, seat: 1),
        TableDropTarget(zone: TableDropZone.seat, seat: 3),
        TableDropTarget(zone: TableDropZone.card, seat: 1, cardId: 'd8'),
      ]) {
        expect(plays(b, {'ad'}, online: online, target: target), isEmpty);
      }
      expect(plays(b, {'ad'}, online: online).map((i) => i.type), [
        'expose-unused-ace',
        'confinement',
        'numerical-defense',
      ]);
      expect(
        plays(
          b,
          {'ad'},
          online: online,
          target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
        ).map((i) => i.type),
        ['confinement'],
      );
      expect(
        plays(
          b,
          {'ad'},
          online: online,
          target: const TableDropTarget(
            zone: TableDropZone.card,
            seat: 1,
            cardId: 'h8',
          ),
        ).map((i) => i.type),
        ['numerical-defense'],
      );
      expect(
        plays(
          b,
          {'ad'},
          online: online,
          target: const TableDropTarget(
            zone: TableDropZone.series,
            seat: 1,
            suit: 'hearts',
            cardId: 'h8',
          ),
        ).map((i) => i.type),
        ['numerical-defense'],
      );
    },
  );

  test(
    'ordered payment drops never reinterpret draw pile or foreign seats',
    () {
      final b = board([card('d4', 'diamonds', 4, zone: 'series')])
        ..addAll({
          'window_id': 'w',
          'required_actor': 1,
          'decision_id': 'd',
          'decision_kind': 'dexter',
          'choices': [
            ['d4'],
          ],
        });
      for (final target in const [
        TableDropTarget(zone: TableDropZone.drawPile),
        TableDropTarget(zone: TableDropZone.hand, seat: 1),
        TableDropTarget(zone: TableDropZone.seat, seat: 3),
      ]) {
        expect(plays(b, {'d4'}, target: target), isEmpty);
      }
      expect(plays(b, {'d4'}).single.type, 'decision');
      expect(
        plays(
          b,
          {'d4'},
          target: const TableDropTarget(zone: TableDropZone.seat, seat: 1),
        ).single.type,
        'decision',
      );
    },
  );

  test(
    'inapplicable Ace drops never become disposal while an explicit choice remains',
    () {
      final b = board([card('as', 'spades', 1, zone: 'concealed-ace')]);
      b['active'] = 2;
      expect(plays(b, {'as'}).map((i) => i.type), ['expose-unused-ace']);
      expect(
        plays(b, {
          'as',
        }, target: const TableDropTarget(zone: TableDropZone.drawPile)),
        isEmpty,
      );
    },
  );

  test('Bomb combat offers no numerical or Advancement defence', () {
    final b = board([card('ac', 'clubs', 1, zone: 'concealed-ace')]);
    b.addAll({'active': 2, 'window_id': 'w', 'required_actor': 1});
    final result = plays(
      b,
      {'ac'},
      online: {
        'rules_context': {
          'combat': {'defender': 1, 'comparison': 'at_least'},
        },
      },
    );
    expect(result.map((i) => i.type), ['expose-unused-ace']);
  });

  test('explicit noncombat responses suppress attack-only Ace choices', () {
    final b = board([
      card('ah', 'hearts', 1, zone: 'concealed-ace'),
      card('as', 'spades', 1, zone: 'concealed-ace'),
      card('ad', 'diamonds', 1, zone: 'concealed-ace'),
      card('ac', 'clubs', 1, zone: 'concealed-ace'),
    ])..addAll({'window_id': 'w', 'required_actor': 1, 'active': 2});
    final online = {
      'rules_context': {
        'pending': {
          'type': 'open-series',
          'actor': 2,
          'response_types': ['confinement'],
        },
      },
    };
    for (final id in ['ah', 'as', 'ac']) {
      expect(plays(b, {id}, online: online).map((i) => i.type), [
        'expose-unused-ace',
      ]);
    }
    expect(plays(b, {'ad'}, online: online).map((i) => i.type), [
      'expose-unused-ace',
      'confinement',
    ]);
    expect(
      plays(
        b,
        {'ad'},
        online: {
          'rules_context': {
            'pending': {
              'type': 'attack',
              'actor': 2,
              'response_types': <String>[],
            },
            'combat': {'comparison': 'equal', 'defender': 1},
          },
        },
      ).map((i) => i.type),
      ['expose-unused-ace'],
    );
    // Missing old-server metadata must never manufacture incoming combat.
    for (final id in ['ah', 'as', 'ac']) {
      expect(plays(b, {id}).map((i) => i.type), ['expose-unused-ace']);
    }
    expect(plays(b, {'ad'}).map((i) => i.type), [
      'expose-unused-ace',
      'confinement',
    ]);
  });

  test(
    'response categories retain named attacks and reject irrelevant roles',
    () {
      final b = board([
        card('ah', 'hearts', 1, zone: 'concealed-ace'),
        card('as', 'spades', 1, zone: 'concealed-ace'),
        card('ac', 'clubs', 1, zone: 'concealed-ace'),
        card('js', 'spades', 11, zone: 'attachment', allocation: 'spades'),
      ])..addAll({'window_id': 'w', 'required_actor': 1, 'active': 2});
      final online = {
        'rules_context': {
          'pending': {
            'type': 'kidnapper',
            'actor': 2,
            'target_seat': 1,
            'response_types': [
              'confinement',
              'compensation-response',
              'inflation',
            ],
          },
        },
      };
      expect(plays(b, {'ah'}, online: online).map((i) => i.type), [
        'expose-unused-ace',
        'compensation-response',
      ]);
      expect(plays(b, {'as'}, online: online).map((i) => i.type), [
        'expose-unused-ace',
        'inflation',
      ]);
      expect(plays(b, {'ac'}, online: online).map((i) => i.type), [
        'expose-unused-ace',
      ]);
      expect(plays(b, {'js'}, online: online), isEmpty);
      final negotiator = {
        'rules_context': {
          'pending': {
            'type': 'attack',
            'actor': 2,
            'response_types': ['negotiate'],
          },
        },
      };
      expect(plays(b, {'js'}, online: negotiator).single.seed, {
        'pending_action_id': 'w',
      });
      expect(
        plays({...b, 'required_actor': 3}, {'js'}, online: negotiator),
        isEmpty,
      );
    },
  );

  test(
    'Advancement follows actual Hearts target and exact response category',
    () {
      final b = board([
        card('ac', 'clubs', 1, zone: 'concealed-ace'),
        card('h8', 'hearts', 8, zone: 'series'),
        card('d8', 'diamonds', 8, zone: 'series'),
      ])..addAll({'window_id': 'w', 'required_actor': 1, 'active': 2});
      Map<String, dynamic> legacy(String id) => {
        'rules_context': {
          'combat': {
            'defender': 1,
            'comparison': 'equal',
            'target_cards': [id],
          },
        },
      };
      expect(plays(b, {'ac'}, online: legacy('h8')).map((i) => i.type), [
        'expose-unused-ace',
        'advancement-defense',
        'numerical-defense',
      ]);
      expect(plays(b, {'ac'}, online: legacy('d8')).map((i) => i.type), [
        'expose-unused-ace',
        'numerical-defense',
      ]);
      expect(
        plays(
          b,
          {'ac'},
          online: {
            'rules_context': {
              'pending': {
                'response_types': ['numerical-defense'],
              },
              'combat': {
                'defender': 1,
                'comparison': 'equal',
                'target_cards': ['h8'],
              },
            },
          },
        ).map((i) => i.type),
        ['expose-unused-ace', 'numerical-defense'],
      );
    },
  );

  test(
    'effect choices are exact authorized selections, not arbitrary response cards',
    () {
      final b = board([
        card('d4', 'diamonds', 4, zone: 'series'),
        card('d6', 'diamonds', 6, zone: 'series'),
      ]);
      b.addAll({
        'window_id': 'w',
        'decision_id': 'd',
        'decision_kind': 'dexter',
        'required_actor': 1,
        'choices': [
          [],
          ['d4'],
        ],
      });
      expect(plays(b, {'d4'}).single.seed, {
        'pending_action_id': 'w',
        'decision_id': 'd',
        'selection': ['d4'],
      });
      expect(plays(b, {'d6'}), isEmpty);
      expect(plays(b, {'d4', 'd6'}), isEmpty);
    },
  );

  test(
    'Justice offered public choices can be selected without authorizing hidden faces',
    () {
      final b = board([
        card('public', 'hearts', 8, zone: 'series', owner: 2),
        card('private', 'spades', 12, owner: 2),
      ]);
      b.addAll({
        'window_id': 'w',
        'decision_id': 'd',
        'decision_kind': 'justice',
        'required_actor': 1,
        'choices': [
          ['public'],
          ['private'],
        ],
      });
      expect(plays(b, {'public'}).single.type, 'decision');
      expect(plays(b, {'private'}), isEmpty);
      expect(selectableCardIds(b, {}), {'public'});
      expect(draggableCardIds(b, {}), isEmpty);
    },
  );

  test('loan return names only the lender for selected borrowed holdings', () {
    final b = board([card('borrowed', 'clubs', 11, zone: 'unassigned')]);
    b['loans'] = [
      {
        'from': 2,
        'to': 1,
        'cards': ['borrowed'],
      },
    ];
    final intent = plays(
      b,
      {'borrowed'},
      target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
    ).singleWhere((i) => i.type == 'return-loan');
    expect(intent.seed, {
      'selection': ['borrowed'],
      'target_seat': 2,
    });
  });

  test(
    'tap selection separates own Clubs from exact visible enemy targets',
    () {
      final b = board([
        card('c8', 'clubs', 8, zone: 'series'),
        card('h3', 'hearts', 3, owner: 2, zone: 'series'),
        card('h5', 'hearts', 5, owner: 2, zone: 'series'),
        card('h7', 'hearts', 7, owner: 2, zone: 'series'),
        card('other', 'hearts', 8, owner: 3, zone: 'series'),
      ]);
      b['formations'] = [formation('baron', 'baron', [])];
      final result = plays(b, {'h3', 'c8', 'h5'});
      expect(result.singleWhere((i) => i.type == 'attack').seed, {
        'selection': ['c8'],
        'targets': ['h3', 'h5'],
        'suit': 'hearts',
      });
      expect(
        result.singleWhere((i) => i.type == 'baron-attack').seed['targets'],
        ['h3', 'h5', 'h7'],
      );
      expect(plays(b, {'c8', 'h3', 'other'}), isEmpty);
      expect(plays(b, {'h3', 'h5'}), isEmpty);
    },
  );

  test('tap selection separates Dexter Diamond payment and exposed royal', () {
    final b = board([
      card('d7', 'diamonds', 7, zone: 'series'),
      card(
        'dexter',
        'diamonds',
        11,
        zone: 'attachment',
        allocation: 'diamonds',
      ),
      card(
        'king',
        'hearts',
        13,
        owner: 2,
        zone: 'attachment',
        allocation: 'clubs',
      ),
      card('hidden', 'hearts', 13, owner: 2),
    ]);
    expect(plays(b, {'d7', 'king'}).single.seed, {
      'selection': ['d7'],
      'targets': ['king'],
    });
    expect(plays(b, {'d7', 'hidden'}), isEmpty);
  });

  test(
    'formation targets derive current ownership even without a seat hint',
    () {
      final b = board([card('k1', 'clubs', 13), card('k2', 'spades', 13)]);
      b['formations'] = [
        {...formation('foreign', 'fate', []), 'controller': 2},
      ];
      expect(
        plays(
          b,
          {'k1', 'k2'},
          target: const TableDropTarget(
            zone: TableDropZone.formation,
            formationId: 'foreign',
          ),
        ),
        isEmpty,
      );
    },
  );

  test(
    'incomplete two-king reconfiguration keeps the existing formation id',
    () {
      final b = board([
        card('k1', 'clubs', 13, zone: 'formation', allocation: 'fate'),
        card('k2', 'spades', 13, zone: 'formation', allocation: 'fate'),
      ]);
      b['formations'] = [
        formation('fate', 'fate', ['k1', 'k2']),
      ];
      final draft = plays(b, {
        'k1',
        'k2',
      }).singleWhere((intent) => intent.type == 'open-formation');
      expect(draft.seed['formation_id'], 'fate');
      expect(draft.needsSubstitution, isTrue);
    },
  );

  test('returned seeds are detached from later board and input mutations', () {
    final b = board([card('s5', 'spades', 5)]);
    final ids = {'s5'};
    final intent = plays(b, ids).singleWhere((i) => i.type == 'purchase');
    ids.clear();
    (b['cards'] as List).clear();
    expect(intent.seed, {
      'payment': ['s5'],
    });
    expect(
      () => (intent.seed['payment'] as List).add('bad'),
      throwsUnsupportedError,
    );
  });

  test('Coup is card driven through confinement and pending decisions', () {
    final b =
        board([
          card('qh1', 'hearts', 12),
          card('qh2', 'hearts', 12),
          card('qd1', 'diamonds', 12),
          card('qd2', 'diamonds', 12),
        ])..addAll({
          'active': 2,
          'window_id': 'w',
          'decision_id': 'd',
          'required_actor': 3,
          'decision_kind': 'justice',
        });
    (b['players'] as List).first['confined'] = true;
    final intents = plays(b, {'qh1'}, online: {'automatic_pending': true});
    expect(intents.map((i) => i.type), ['coup']);
    expect(intents.single.seed, isEmpty);
    expect(plays({...b, 'phase': 'settlement'}, {'qh1'}), isEmpty);
  });

  test('Coup never substitutes restricted or foreign Queens', () {
    final b = board([
      card('qh1', 'hearts', 12),
      card('qh2', 'hearts', 12),
      card('qd1', 'diamonds', 12),
      card('qd2', 'diamonds', 12, available: 4),
    ]);
    expect(plays(b, {'qh1'}).any((i) => i.type == 'coup'), isFalse);
    (b['cards'] as List).last['available_from_round'] = 0;
    (b['cards'] as List).last['controller'] = 2;
    expect(plays(b, {'qh1'}).any((i) => i.type == 'coup'), isFalse);
  });

  test('foreign effect choices are targets and never draggable sources', () {
    final b =
        board([card('target', 'hearts', 12, owner: 2, zone: 'unassigned')])
          ..addAll({
            'window_id': 'w',
            'decision_id': 'd',
            'decision_kind': 'justice',
            'required_actor': 1,
            'choices': [
              ['target'],
            ],
          });
    expect(draggableCardIds(b, {}), isEmpty);
    expect(plays(b, {'target'}).single.type, 'decision');
  });

  test('a wrong own series drop does not silently open a different suit', () {
    final b = board([
      card('h4', 'hearts', 4),
      card('s5', 'spades', 5, zone: 'series'),
    ]);
    expect(
      plays(
        b,
        {'h4'},
        target: const TableDropTarget(
          zone: TableDropZone.card,
          seat: 1,
          cardId: 's5',
        ),
      ).any((i) => i.type == 'open-series'),
      isFalse,
    );
  });

  test('Dexter plus exact payment plus target retains distinct roles', () {
    final b = board([
      card(
        'dexter',
        'diamonds',
        11,
        zone: 'attachment',
        allocation: 'diamonds',
      ),
      card('d9', 'diamonds', 9, zone: 'series'),
      card(
        'queen',
        'clubs',
        12,
        owner: 2,
        zone: 'attachment',
        allocation: 'clubs',
      ),
    ]);
    final intent = plays(b, {
      'dexter',
      'd9',
      'queen',
    }).singleWhere((i) => i.type == 'dexter-assassination');
    expect(intent.seed, {
      'selection': ['d9'],
      'targets': ['queen'],
    });
  });

  test('selected Infiltrator and Exile are exact attack modifiers', () {
    final b = board([
      card('c8', 'clubs', 8, zone: 'series'),
      card('infiltrator', 'clubs', 11, zone: 'attachment', allocation: 'clubs'),
      card('exile', 'clubs', 12, zone: 'attachment', allocation: 'clubs'),
      card('target', 'spades', 8, zone: 'series', owner: 2),
    ]);
    final intent = plays(b, {
      'c8',
      'infiltrator',
      'exile',
      'target',
    }).singleWhere((i) => i.type == 'infiltrator-attack');
    expect(intent.seed, {
      'selection': ['c8'],
      'targets': ['target'],
      'suit': 'spades',
      'exile': true,
      'infiltrator': true,
    });
    expect(
      plays(b, {
        'c8',
        'exile',
        'target',
      }).singleWhere((i) => i.type == 'attack').seed['exile'],
      isTrue,
    );
  });

  test(
    'an ability card alone enters attack targeting without choosing Clubs',
    () {
      final b = board([
        card('exile', 'clubs', 12, zone: 'attachment', allocation: 'clubs'),
        card('c8', 'clubs', 8, zone: 'series'),
      ]);
      final attack = plays(b, {'exile'}).singleWhere((i) => i.type == 'attack');
      expect(attack.seed, {'exile': true});
    },
  );

  test(
    'Baron card initiates its formation attack without selecting payment',
    () {
      final b = board([
        card('king', 'diamonds', 13, zone: 'formation', allocation: 'baron'),
      ]);
      b['formations'] = [
        formation('baron', 'baron', ['king']),
      ];
      final intent = plays(b, {
        'king',
      }).singleWhere((i) => i.type == 'baron-attack');
      expect(intent.seed, isEmpty);
    },
  );

  test('hand card to opponent drafts private revisioned trade terms only', () {
    final b = board([card('qh', 'hearts', 12)]);
    final intent = plays(
      b,
      {'qh'},
      target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
    ).singleWhere((i) => i.type == 'offer');
    expect(intent.seed, {
      'terms': {
        'to': 2,
        'give': ['qh'],
        'receive': <String>[],
        'loan': false,
      },
    });
    expect(intent.seed.containsKey('offer_id'), isFalse);
    expect(intent.seed.containsKey('revision'), isFalse);
  });

  test('intact combo loan on recipient turn retains every physical member', () {
    final b = board([
      card('j1', 'clubs', 11, zone: 'formation', allocation: 'kid'),
      card('j2', 'clubs', 11, zone: 'formation', allocation: 'kid'),
    ])..['active'] = 2;
    b['formations'] = [
      formation('kid', 'kidnapper', ['j1', 'j2']),
    ];
    final intent = plays(
      b,
      {'j1'},
      target: const TableDropTarget(zone: TableDropZone.seat, seat: 2),
    ).singleWhere((i) => i.type == 'offer');
    expect(intent.seed['terms'], {
      'to': 2,
      'give': ['j1', 'j2'],
      'receive': <String>[],
      'loan': true,
    });
    expect(
      plays(
        b,
        {'j1'},
        target: const TableDropTarget(zone: TableDropZone.seat, seat: 3),
      ).any((i) => i.type == 'offer'),
      isFalse,
    );
  });

  test(
    'Ringleader preserves explicitly selected Ace instead of changing payment',
    () {
      final b = board([
        card('king', 'diamonds', 13, zone: 'attachment', allocation: 'hearts'),
        card('ace', 'clubs', 1, zone: 'concealed-ace'),
      ]);
      // No rule permits spending a second card for Ringleader: the mixed
      // selection must not silently sacrifice only the king.
      expect(
        plays(b, {'king', 'ace'}).any((i) => i.type == 'ringleader'),
        isFalse,
      );
    },
  );

  test(
    'returned bound kings keep their pair and cannot attach or use printed roles',
    () {
      final b = board([
        card('k1', 'clubs', 13),
        card('k2', 'spades', 13),
        card('k3', 'hearts', 13),
        card('k4', 'diamonds', 13),
        card('qh', 'hearts', 12),
        card('h', 'hearts', 5, zone: 'series'),
        card('c', 'clubs', 5, zone: 'series'),
      ]);
      b['own_bindings'] = [
        {
          'kings': ['k1', 'k2'],
          'slot': {'rank': 12, 'suit': 'hearts'},
        },
      ];
      expect(royalAttachmentSuits(b, {'k1'}), isEmpty);
      expect(
        plays(b, {
          'k1',
          'k2',
          'k3',
          'k4',
        }).where((i) => i.type == 'open-formation'),
        isEmpty,
      );
      expect(
        plays(b, {'k1', 'k3'}).where((i) => i.type == 'open-formation'),
        isEmpty,
      );
      final reopened = plays(b, {
        'qh',
        'k1',
        'k2',
      }).where((i) => i.type == 'open-formation').single;
      expect(reopened.seed['formation']['kind'], 'people');
      expect(reopened.seed['formation']['substitute'], b['own_bindings'][0]);
      expect(reopened.needsSubstitution, isFalse);
      expect(
        formationHonorsFixedBindings(b, {
          'qh',
          'k1',
          'k2',
        }, substitute: Map<String, dynamic>.from(b['own_bindings'][0])),
        isTrue,
      );
      expect(formationHonorsFixedBindings(b, {'qh', 'k1', 'k2'}), isFalse);
      expect(
        formationHonorsFixedBindings(
          b,
          {'qh', 'k1', 'k2'},
          substitute: {
            'kings': ['k1', 'k2'],
            'slot': {'rank': 11, 'suit': 'clubs'},
          },
        ),
        isFalse,
      );
      expect(
        formationHonorsFixedBindings(
          b,
          {'qh', 'k1', 'k3'},
          substitute: {
            'kings': ['k1', 'k3'],
            'slot': {'rank': 12, 'suit': 'hearts'},
          },
        ),
        isFalse,
      );
      expect(reopened.needsProtection, isTrue);
    },
  );

  test(
    'separated bound king stays unavailable without exposing a partner identity',
    () {
      final b = board([
        card('k1', 'clubs', 13),
        card('fresh', 'spades', 13),
        card('h', 'hearts', 5, zone: 'series'),
      ]);
      b['own_bindings'] = [
        {
          'kings': ['k1'],
          'slot': {'rank': 11, 'suit': 'clubs'},
        },
      ];
      expect(royalAttachmentSuits(b, {'k1'}), isEmpty);
      expect(royalAttachmentSuits(b, {'fresh'}), {'hearts'});
      expect(
        plays(b, {'k1', 'fresh'}).where((i) => i.type == 'open-formation'),
        isEmpty,
      );
    },
  );

  test(
    'known Doppelganger binding survives same-formation reconfiguration',
    () {
      final b = board([
        card('qh', 'hearts', 12, zone: 'formation', allocation: 'people'),
        card('k1', 'clubs', 13, zone: 'formation', allocation: 'people'),
        card('k2', 'spades', 13, zone: 'formation', allocation: 'people'),
      ]);
      b['formations'] = [
        {
          ...formation('people', 'people', ['qh', 'k1', 'k2']),
          'spec': {
            'kind': 'people',
            'cards': ['qh', 'k1', 'k2'],
            'substitute': {
              'kings': ['k1', 'k2'],
              'slot': {'rank': 12, 'suit': 'hearts'},
            },
            'protection': {'series': 'clubs'},
          },
        },
      ];
      final intent = plays(b, {'qh', 'k1', 'k2'}).singleWhere(
        (i) =>
            i.type == 'open-formation' &&
            i.seed['formation']['kind'] == 'people',
      );
      expect(intent.seed['formation']['substitute'], {
        'kings': ['k1', 'k2'],
        'slot': {'rank': 12, 'suit': 'hearts'},
      });
      expect(intent.needsSubstitution, isFalse);
      expect(intent.needsProtection, isTrue);
    },
  );

  test('Coup timing does not inherit ordinary automatic-processing lock', () {
    expect(
      commandTiming(
        'coup',
        {'phase': 'playing'},
        {'automatic_pending': true, 'server_pending': true},
      ),
      CommandTiming.available,
    );
    expect(
      commandTiming(
        'attack',
        {'phase': 'playing'},
        {'automatic_pending': true},
      ),
      CommandTiming.automatic,
    );
  });

  test('empty required decision is explicit and never a response pass', () {
    final b = board([])
      ..addAll({
        'window_id': 'w',
        'decision_id': 'd',
        'decision_kind': 'dexter',
        'required_actor': 1,
        'choices': [<String>[]],
      });
    final intent = emptyDecisionIntent(b, {});
    expect(intent?.type, 'decision');
    expect(intent?.seed, {
      'pending_action_id': 'w',
      'decision_id': 'd',
      'selection': <String>[],
    });
    expect(emptyDecisionIntent({...b, 'required_actor': 2}, {}), isNull);
    expect(
      emptyDecisionIntent({
        ...b,
        'choices': [
          ['unknown'],
        ],
      }, {}),
      isNull,
    );
  });

  test('Coup is omitted for a known permanent alliance', () {
    final b = board([
      card('qh1', 'hearts', 12),
      card('qh2', 'hearts', 12),
      card('qd1', 'diamonds', 12),
      card('qd2', 'diamonds', 12),
    ]);
    expect(
      plays(
        b,
        {'qh1'},
        online: {
          'alliances': {'1': 2, '2': 1},
        },
      ).any((intent) => intent.type == 'coup'),
      isFalse,
    );
  });

  test(
    'formation card offers whole exact reconfiguration without extra opening',
    () {
      final b = board([
        card('qh1', 'hearts', 12, zone: 'formation', allocation: 'protect'),
        card('qh2', 'hearts', 12, zone: 'formation', allocation: 'protect'),
      ]);
      b['formations'] = [
        {
          ...formation('protect', 'great-people', ['qh1', 'qh2']),
          'spec': {
            'kind': 'great-people',
            'cards': ['qh1', 'qh2'],
            'protection': {'series': 'clubs'},
          },
        },
      ];
      final intent = plays(b, {
        'qh1',
      }).singleWhere((i) => i.type == 'open-formation');
      expect(intent.seed['formation_id'], 'protect');
      expect(intent.seed['formation']['cards'], ['qh1', 'qh2']);
      expect(intent.needsProtection, isTrue);
      expect(
        plays(
          b,
          {'qh1'},
          online: {
            'rules_context': {'opening_used': true},
          },
        ).any((i) => i.type == 'open-formation'),
        isFalse,
      );
    },
  );

  test(
    'series identity uses suit and cannot redirect a contradictory card hint',
    () {
      final b = board([
        card('h4', 'hearts', 4),
        card('ks', 'spades', 13),
        card('d4', 'diamonds', 4, zone: 'series'),
      ]);
      expect(
        plays(
          b,
          {'h4'},
          target: const TableDropTarget(
            zone: TableDropZone.series,
            seat: 1,
            suit: 'spades',
          ),
        ).any((i) => i.type == 'open-series'),
        isFalse,
      );
      expect(
        plays(
          b,
          {'ks'},
          target: const TableDropTarget(
            zone: TableDropZone.series,
            seat: 1,
            suit: 'diamonds',
          ),
        ).singleWhere((i) => i.type == 'attach').seed['suit'],
        'diamonds',
      );
      expect(
        plays(
          b,
          {'ks'},
          target: const TableDropTarget(
            zone: TableDropZone.series,
            seat: 1,
            suit: 'clubs',
            cardId: 'd4',
          ),
        ),
        isEmpty,
      );
    },
  );

  test(
    'explicit unused Ace exposure is offered only as a named contextual choice',
    () {
      final b = board([card('ace', 'hearts', 1, zone: 'concealed-ace')]);
      final intent = plays(b, {
        'ace',
      }).singleWhere((i) => i.type == 'expose-unused-ace');
      expect(intent.seed, {'ace': 'ace'});
      expect(intent.explanation, contains('discard'));
      expect(
        plays(b, {
          'ace',
        }, target: const TableDropTarget(zone: TableDropZone.drawPile)),
        isEmpty,
      );
      expect(
        plays(b, {
          'ace',
        }, target: const TableDropTarget(zone: TableDropZone.hand)),
        isEmpty,
      );
    },
  );

  testWidgets(
    'trade and whole-formation loan have distinct contextual labels',
    (tester) async {
      final chosen = <bool>[];
      await tester.pumpWidget(
        MaterialApp(
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: CardPlayChoices(
              intents: [
                for (final loan in [false, true])
                  CardPlayIntent(
                    type: 'offer',
                    seed: {
                      'terms': {
                        'give': ['member'],
                        'loan': loan,
                      },
                    },
                    explanation: '',
                  ),
              ],
              label: (_) => 'Offer',
              onSelected: (intent) => chosen.add(intent.seed['terms']['loan']),
            ),
          ),
        ),
      );
      await tester.tap(find.text('Trade cards'));
      await tester.tap(find.text('Loan this intact formation'));
      expect(chosen, [false, true]);
      expect(find.textContaining('combo 1'), findsNothing);
    },
  );
}
