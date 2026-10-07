import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter_test/flutter_test.dart';

import '../lib/demo_controller.dart';

void main() {
  late DemoController demo;

  setUp(() => demo = DemoController());
  tearDown(() => demo.dispose());

  void passAll() {
    while (demo.table.pending != null) {
      demo.primaryAction();
    }
  }

  void clearHearts() {
    demo.selectCard('deck-1-clubs-7');
    demo.primaryAction();
    passAll();
  }

  void clearClubs() {
    clearHearts();
    demo.selectCard('deck-1-clubs-5');
    demo.primaryAction();
    passAll();
  }

  test(
    'prepared snapshot keeps physical copies and opponent secrets distinct',
    () {
      final ids = [
        ...demo.table.hand.map((card) => card.id),
        ...demo.table.seats.expand(
          (seat) => seat.exposed.map((card) => card.id),
        ),
      ];
      expect(ids.toSet().length, ids.length);
      expect(
        demo.table.seats
            .expand((seat) => seat.exposed)
            .where((card) => card.rank == '5' && card.suit == CardSuit.clubs)
            .map((card) => card.copy)
            .toSet(),
        {1, 2},
      );
      expect(
        demo.table.hand.map((card) => card.id),
        contains('deck-1-hearts-5'),
      );
      expect(
        demo.table.hand.map((card) => card.id),
        isNot(contains('deck-1-hearts-6')),
      );
      expect(demo.table.seats[2].concealed.handCount, greaterThan(0));
    },
  );

  test('only legal clubs select; declaration freezes response order', () {
    demo.selectCard('deck-1-spades-Q');
    expect(demo.table.selectedCardId, isNull);
    demo.selectCard('deck-1-clubs-5');
    expect(demo.table.selectedCardId, isNull);
    demo.selectCard('deck-1-clubs-7');
    expect(demo.table.actionEnabled, isTrue);
    demo.primaryAction();
    expect(demo.table.pending!.responders, ['Mara', 'Ivo', 'Nia']);
    expect(demo.table.pending!.responseIndex, 0);
    demo.selectCard('deck-1-clubs-5');
    expect(demo.table.selectedCardId, 'deck-1-clubs-7');
  });

  test('7 clubs eliminates 4+3 hearts and remains committed without score', () {
    clearHearts();
    expect(
      demo.table.seats[1].exposed.any((card) => card.suit == CardSuit.hearts),
      isFalse,
    );
    expect(
      demo.table.seats.first.exposed.any((card) => card.id == 'deck-1-clubs-7'),
      isTrue,
    );
    expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(-5));
    demo.selectCard('deck-1-clubs-7');
    expect(demo.table.selectedCardId, isNull);
  });

  test(
    'unused physical club exchanges once and gross receipt repays debt once',
    () {
      clearClubs();
      expect(
        demo.table.seats.first.exposed.any(
          (card) => card.id == 'deck-1-clubs-5',
        ),
        isFalse,
      );
      expect(
        demo.table.seats[1].exposed.any((card) => card.id == 'deck-2-clubs-5'),
        isFalse,
      );
      expect(demo.table.seats[1].status, contains('protection stays ended'));
      final money = demo.table.seats.first.finances;
      expect(money.gameScore, ExactPoints.integer(-2));
      expect(money.available, ExactPoints.integer(0));
      expect(money.debt, ExactPoints.integer(2));
      demo.selectCard('deck-1-clubs-5');
      expect(demo.table.selectedCardId, isNull);
    },
  );

  test(
    'trade is nonblocking then transfers atomically after acceptor response order',
    () {
      clearClubs();
      demo.offerTrade();
      expect(demo.tradeOffered, isTrue);
      expect(demo.table.pending, isNull);
      expect(
        demo.table.hand.any((card) => card.id == 'deck-1-hearts-5'),
        isTrue,
      );
      demo.acceptTrade();
      expect(demo.table.pending!.responders, ['Nia', 'You', 'Mara']);
      expect(
        demo.table.hand.any((card) => card.id == 'deck-1-hearts-5'),
        isTrue,
      );
      passAll();
      expect(demo.tradeResolved, isTrue);
      expect(
        demo.table.hand.any((card) => card.id == 'deck-1-hearts-5'),
        isFalse,
      );
      expect(
        demo.table.hand.any((card) => card.id == 'deck-1-hearts-6'),
        isTrue,
      );
      expect(
        demo.table.seats.first.finances.gameScore,
        ExactPoints.integer(-2),
      );
      expect(
        demo.table.seats[2].exposed.any((card) => card.suit == CardSuit.hearts),
        isFalse,
      );
      final once = demo.table.hand.map((card) => card.id).toList();
      demo.acceptTrade();
      expect(demo.table.hand.map((card) => card.id).toList(), once);
    },
  );

  test('an unanswered offer expires on end turn without transferring', () {
    demo.offerTrade();
    demo.endTurn();
    expect(demo.boardClosed, isTrue);
    expect(demo.tradeOffered, isFalse);
    expect(demo.tradeStatus, contains('Expired'));
    expect(demo.table.hand.any((card) => card.id == 'deck-1-hearts-5'), isTrue);
  });

  test('an unanswered proposal does not block another legal attack', () {
    demo.offerTrade();
    demo.selectCard('deck-1-clubs-7');
    expect(demo.table.actionEnabled, isTrue);
    demo.primaryAction();
    expect(demo.table.pending, isNotNull);
    expect(demo.tradeOffered, isTrue);
    demo.acceptTrade();
    expect(demo.table.pending!.title, contains('attack'));
    passAll();
    expect(demo.tradeOffered, isTrue);
  });

  test(
    'disconnect and reconnect preserve required response without inferred pass',
    () {
      demo.selectCard('deck-1-clubs-7');
      demo.primaryAction();
      final before = demo.table.pending!.responseIndex;
      demo.setConnection(ConnectionStateView.disconnected);
      demo.primaryAction();
      demo.endTurn();
      expect(demo.table.pending!.responseIndex, before);
      expect(demo.boardClosed, isFalse);
      demo.setConnection(ConnectionStateView.connected);
      expect(demo.table.pending!.responseIndex, before);
      demo.primaryAction();
      expect(demo.table.pending!.responseIndex, before + 1);
    },
  );

  test('ordinary awards, debt and cash expiry are separate and idempotent', () {
    clearClubs();
    demo.endTurn();
    expect(demo.boardClosed, isTrue);
    var money = demo.table.seats.first.finances;
    expect(money.gameScore, ExactPoints.integer(23));
    expect(money.matchScore, ExactPoints.integer(23));
    expect(money.available, ExactPoints.integer(23));
    expect(money.debt, ExactPoints.integer(0));
    expect(demo.table.seats[1].finances.gameScore, ExactPoints.integer(13));
    expect(demo.table.seats[2].finances.gameScore, ExactPoints.integer(11));
    expect(demo.table.seats[3].finances.gameScore, ExactPoints.integer(12));
    demo.endTurn();
    expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(23));
    demo.finishSettlement();
    expect(demo.finalized, isTrue);
    money = demo.table.seats.first.finances;
    expect(money.gameScore, ExactPoints.integer(23));
    expect(money.available, ExactPoints.integer(0));
    demo.finishSettlement();
    expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(23));
  });

  test('configuration survives reset and stale game commands are rejected', () {
    demo.configure(name: 'Ada', games: 3);
    expect(demo.table.seats.first.name, 'Ada');
    expect(demo.games, 3);
    final oldId = demo.table.gameId;
    demo.reset();
    expect(demo.table.seats.first.name, 'Ada');
    expect(demo.table.gameId, isNot(oldId));
    expect(demo.acceptsGame(oldId), isFalse);
    expect(demo.acceptsGame(demo.table.gameId), isTrue);
  });

  test('host permissions exclude empty, pending and disconnected states', () {
    expect(demo.canTrade, isTrue);
    expect(demo.canEndTurn, isTrue);
    demo.loadScenario('empty');
    expect(demo.canTrade, isFalse);
    expect(demo.canEndTurn, isFalse);
    demo.loadScenario('pending');
    expect(demo.canTrade, isFalse);
    expect(demo.canEndTurn, isFalse);
    demo.loadScenario('disconnected');
    expect(demo.canTrade, isFalse);
    expect(demo.canEndTurn, isFalse);
  });

  test(
    'instructional card text uses readable suit names and natural pronouns',
    () {
      expect(demo.table.turnLabel, contains('Your turn'));
      final unsupportedGlyphs = RegExp('[♣♥♦♠]');
      for (final id in DemoController.scenarioIds) {
        demo.loadScenario(id);
        final state = demo.table;
        final visibleText = [
          state.actionLabel,
          state.actionExplanation,
          state.objective,
          ...state.events,
          if (state.pending != null) state.pending!.detail,
        ].join(' ');
        expect(unsupportedGlyphs.hasMatch(visibleText), isFalse, reason: id);
      }
      demo.reset();
      demo.offerTrade();
      demo.acceptTrade();
      demo.primaryAction();
      expect(demo.table.actionLabel, 'Simulate your pass');
    },
  );

  test(
    'review scenarios are deterministic and error/loading/empty are inert',
    () {
      for (final id in ['loading', 'error', 'empty']) {
        demo.loadScenario(id);
        expect(demo.table.actionEnabled, isFalse);
        demo.primaryAction();
        expect(demo.boardClosed, isFalse);
      }
      demo.loadScenario('pending');
      expect(demo.table.pending!.responseIndex, 0);
      demo.loadScenario('ready');
      expect(demo.table.pending, isNull);
      expect(demo.table.hand.length, 4);
    },
  );

  for (final handSize in [10, 15]) {
    test(
      'hand-$handSize has a dense legal public fixture and neutral accounts',
      () {
        demo.loadScenario('hand-$handSize');
        final state = demo.table;
        expect(demo.densityReview, isTrue);
        expect(state.round, 8);
        expect(state.hand.length, handSize);
        expect(state.seats.first.concealed.handCount, handSize);
        expect(state.seats.map((seat) => seat.exposed.length), [
          12,
          12,
          12,
          12,
        ]);
        for (final seat in state.seats) {
          expect(seat.status, contains('initial protection ended'));
          expect(seat.concealed.aceCount, 0);
          expect(seat.finances.gameScore, ExactPoints.integer(0));
          expect(seat.finances.available, ExactPoints.integer(0));
          expect(seat.finances.debt, ExactPoints.integer(0));
          for (final suit in CardSuit.values) {
            expect(seat.exposed.where((card) => card.suit == suit).length, 3);
          }
        }
        final visibleIds = [
          ...state.hand.map((card) => card.id),
          ...state.seats.expand((seat) => seat.exposed.map((card) => card.id)),
        ];
        expect(visibleIds.toSet().length, visibleIds.length);
        final controlledCount = state.seats.fold<int>(
          0,
          (count, seat) =>
              count +
              seat.exposed.length +
              seat.concealed.handCount +
              seat.concealed.aceCount,
        );
        expect(controlledCount, 48 + handSize + 9);
        expect(controlledCount, lessThanOrEqualTo(104));
        expect(demo.canTrade, isFalse);
        expect(demo.canEndTurn, isFalse);
        demo.offerTrade();
        demo.endTurn();
        expect(demo.tradeOffered, isFalse);
        expect(demo.boardClosed, isFalse);
      },
    );

    test(
      'hand-$handSize adds one physical card only after ordered explicit passes',
      () {
        demo.configure(name: 'Ada', games: 3);
        demo.loadScenario('hand-$handSize');
        final beforeIds = demo.table.hand.map((card) => card.id).toSet();
        demo.selectCard('deck-1-clubs-7');
        expect(demo.table.actionEnabled, isTrue);
        expect(demo.table.actionLabel, 'Add 7 of Clubs to your series');
        demo.primaryAction();
        expect(demo.table.pending!.responders, ['Mara', 'Ivo', 'Nia']);
        expect(demo.table.pending!.detail, contains('opening allowance'));
        expect(demo.table.events, contains(contains('Ada declared')));
        demo.clearSelection();
        demo.selectCard('deck-1-spades-Q');
        expect(demo.table.selectedCardId, 'deck-1-clubs-7');
        for (var response = 0; response < 2; response++) {
          demo.primaryAction();
          expect(demo.table.hand.length, handSize);
          expect(demo.table.seats.first.exposed.length, 12);
        }
        demo.setConnection(ConnectionStateView.disconnected);
        demo.primaryAction();
        expect(demo.table.pending!.responseIndex, 2);
        demo.setConnection(ConnectionStateView.connected);
        demo.primaryAction();
        expect(demo.table.pending, isNull);
        expect(demo.table.hand.length, handSize - 1);
        expect(demo.table.seats.first.concealed.handCount, handSize - 1);
        expect(demo.table.seats.first.exposed.length, 13);
        expect(
          demo.table.hand.map((card) => card.id).toSet(),
          beforeIds.difference({'deck-1-clubs-7'}),
        );
        expect(
          demo.table.seats.first.exposed
              .where((card) => card.id == 'deck-1-clubs-7')
              .length,
          1,
        );
        expect(
          demo.table.seats.first.finances.gameScore,
          ExactPoints.integer(0),
        );
        expect(demo.table.actionEnabled, isFalse);
        demo.primaryAction();
        expect(demo.table.seats.first.exposed.length, 13);
      },
    );
  }

  test(
    'combo review shows legal public formations and a private hand pattern',
    () {
      demo.loadScenario('combos');
      final state = demo.table;
      expect(demo.densityReview, isTrue);
      expect(state.hand.length, 15);
      expect(state.seats.map((seat) => seat.exposed.length), [12, 14, 16, 12]);
      final greatPeople = state.seats[1].combos.single;
      expect(greatPeople.title, 'Great People');
      expect(greatPeople.stateLabel, 'Active · protecting Diamonds');
      expect(greatPeople.cards.map((card) => card.id), [
        'deck-1-hearts-Q',
        'deck-2-hearts-Q',
      ]);
      final fate = state.seats[2].combos.single;
      expect(fate.title, 'Fate');
      expect(fate.stateLabel, 'Opened · waiting for Ivo’s turn');
      expect(fate.cards.map((card) => card.rank).toSet(), {'K'});
      expect(
        fate.cards.map((card) => card.suit).toSet(),
        CardSuit.values.toSet(),
      );
      expect(fate.cards.every((card) => card.copy == 2), isTrue);
      for (final seat in state.seats) {
        expect(seat.finances.gameScore, ExactPoints.integer(0));
        expect(seat.finances.available, ExactPoints.integer(0));
        expect(seat.finances.debt, ExactPoints.integer(0));
        for (final combo in seat.combos) {
          expect(combo.description, contains('Inspection only'));
          for (final card in combo.cards) {
            expect(
              seat.exposed.where((held) => held.id == card.id),
              hasLength(1),
            );
          }
        }
        for (final suit in CardSuit.values) {
          expect(
            seat.exposed.where(
              (card) => card.suit == suit && int.tryParse(card.rank) != null,
            ),
            hasLength(3),
          );
        }
      }
      final potential = state.handCombos.single;
      expect(potential.title, 'Kidnapper');
      expect(potential.stateLabel, 'Potential · still in hand');
      expect(potential.cards.map((card) => card.id), [
        'deck-1-hearts-J',
        'deck-2-hearts-J',
      ]);
      expect(
        potential.cards.every((card) => state.hand.contains(card)),
        isTrue,
      );
      expect(state.seats.first.combos, isEmpty);
      expect(state.seats[3].combos, isEmpty);
      final visibleIds = [
        ...state.hand.map((card) => card.id),
        ...state.seats.expand((seat) => seat.exposed.map((card) => card.id)),
      ];
      expect(visibleIds.toSet().length, visibleIds.length);
      expect(visibleIds, isNot(contains('deck-1-diamonds-Q')));
      expect(visibleIds, isNot(contains('deck-1-spades-K')));
      expect(state.events.join(' '), isNot(contains('Queen of Diamonds')));
      expect(
        state.seats.fold<int>(
          0,
          (count, seat) =>
              count +
              seat.exposed.length +
              seat.concealed.handCount +
              seat.concealed.aceCount,
        ),
        78,
      );
      expect(demo.canTrade, isFalse);
      expect(demo.canEndTurn, isFalse);
    },
  );

  test(
    'combo review preserves addition flow and resets all combo projections',
    () {
      demo.loadScenario('combos');
      final originalHand = demo.table.hand.map((card) => card.id).toList();
      final originalGame = demo.table.gameId;
      demo.selectCard('deck-2-hearts-J');
      expect(demo.table.actionEnabled, isFalse);
      demo.primaryAction();
      expect(demo.table.pending, isNull);
      demo.selectCard('deck-1-clubs-7');
      demo.primaryAction();
      passAll();
      expect(demo.table.hand.length, 14);
      expect(demo.table.handCombos.single.cards, hasLength(2));
      expect(demo.table.seats.first.exposed, hasLength(13));
      expect(demo.table.seats[1].combos.single.title, 'Great People');
      expect(demo.table.seats[2].combos.single.title, 'Fate');
      expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(0));
      demo.loadScenario('combos');
      expect(demo.table.gameId, isNot(originalGame));
      expect(demo.table.hand.map((card) => card.id), originalHand);
      demo.reset();
      expect(demo.table.handCombos, isEmpty);
      expect(demo.table.seats.expand((seat) => seat.combos), isEmpty);
      demo.loadScenario('hand-15');
      expect(demo.table.handCombos, isEmpty);
      expect(demo.table.seats.expand((seat) => seat.combos), isEmpty);
    },
  );

  test(
    'density selection and clearing review any own hand card without inventing actions',
    () {
      demo.loadScenario('hand-15');
      final ids = demo.table.hand.map((card) => card.id).toList();
      for (final id in ids) {
        demo.selectCard(id);
        expect(demo.table.selectedCardId, id);
        if (id != 'deck-1-clubs-7') {
          expect(demo.table.actionEnabled, isFalse);
          expect(demo.table.actionExplanation, contains('7 of Clubs'));
          demo.primaryAction();
          expect(demo.table.pending, isNull);
        }
        demo.clearSelection();
        expect(demo.table.selectedCardId, isNull);
      }
      expect(demo.table.hand.map((card) => card.id).toList(), ids);
      demo.reset();
      expect(demo.densityReview, isFalse);
      expect(demo.table.round, 13);
      expect(demo.table.hand.length, 4);
      expect(demo.canEndTurn, isTrue);
    },
  );
}
