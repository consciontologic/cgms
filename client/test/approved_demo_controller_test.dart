import 'package:cgms_demo/approved_demo_controller.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'main board CTA describes the actual turn or remaining action choice',
    () {
      final controller = ApprovedDemoController();
      controller.loadExample(6);
      controller.backToBoard();
      expect(controller.primaryLabel, 'Choose an action below');
      expect(controller.primaryEnabled, isFalse);
      controller.endTurn();
      expect(controller.primaryLabel, 'Bob’s turn');
      expect(controller.primaryEnabled, isFalse);
    },
  );
  late ApprovedDemoController demo;
  setUp(() => demo = ApprovedDemoController());
  tearDown(() => demo.dispose());

  void passAll() {
    while (demo.table.pending != null) {
      demo.primary();
    }
  }

  void openHearts() {
    demo.beginOpening();
    demo.primary();
    passAll();
  }

  void attack() {
    demo.beginAttack();
    demo.primary();
    passAll();
  }

  test(
    'approved board projects all 18 own cards and public formations once',
    () {
      final table = demo.table;
      expect(table.hand, hasLength(18));
      expect(table.seats.map((s) => s.name), [
        'Alice',
        'Bob',
        'Carol',
        'Deniz',
      ]);
      expect(table.seats.map((s) => s.concealed.handCount), [18, 16, 17, 17]);
      expect(table.seats.map((s) => s.exposed.length), [3, 4, 7, 7]);
      expect(table.seats[2].combos.single.title, 'Underground');
      expect(table.seats[3].combos.single.title, 'Ponzi');
      final ids = [
        ...table.hand,
        ...table.seats.expand((s) => s.exposed),
      ].map((c) => c.id).toList();
      expect(ids.toSet(), hasLength(ids.length));
      expect(demo.fixtureIsValid, isTrue);
      expect(ApprovedDemoController.examples, hasLength(20));
    },
  );

  test(
    'opening commits all seven held Hearts only after explicit responses',
    () {
      final before = demo.table.hand.map((c) => c.id).toSet();
      demo.beginOpening();
      expect(demo.selectedIds, hasLength(7));
      demo.primary();
      expect(demo.table.hand, hasLength(18));
      expect(demo.table.pending!.responders, ['Bob', 'Carol', 'Deniz']);
      passAll();
      expect(demo.table.hand, hasLength(11));
      expect(
        demo.table.seats.first.exposed.where((c) => c.suit == CardSuit.hearts),
        hasLength(7),
      );
      expect({
        ...demo.table.hand.map((c) => c.id),
        ...demo.table.seats.first.exposed
            .where((c) => c.suit == CardSuit.hearts)
            .map((c) => c.id),
      }, before);
      expect(demo.openingUsed, isTrue);
      expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(0));
      demo.beginOpening();
      demo.primary();
      expect(demo.table.hand, hasLength(11));
    },
  );

  test(
    '8 Clubs eliminates Bob 8 Hearts exactly and remains used with no score',
    () {
      openHearts();
      demo.beginAttack();
      demo.primary();
      expect(
        demo.table.seats.first.exposed
            .singleWhere((c) => c.suit == CardSuit.clubs)
            .status,
        contains('Used'),
      );
      expect(
        demo.table.seats[1].exposed.where((c) => c.suit == CardSuit.hearts),
        hasLength(2),
      );
      passAll();
      expect(
        demo.table.seats[1].exposed
            .where((c) => c.suit == CardSuit.hearts)
            .single
            .rank,
        '6',
      );
      expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(0));
      final accepted = demo.acceptedOperations;
      demo.beginAttack();
      expect(demo.acceptedOperations, accepted);
      expect(demo.canAttack, isFalse);
    },
  );

  test(
    'private offer reserves nothing; acceptance and ordered passes exchange physical cards',
    () {
      openHearts();
      attack();
      demo.beginTrade();
      demo.primary();
      expect(demo.offerPending, isTrue);
      expect(demo.table.pending, isNull);
      expect(demo.table.hand.any((c) => c.id == 'deck-1-clubs-Q'), isTrue);
      demo.primary();
      expect(demo.table.pending!.responders, ['Carol', 'Deniz', 'Alice']);
      expect(demo.table.hand.any((c) => c.id == 'deck-1-clubs-Q'), isTrue);
      passAll();
      expect(demo.table.hand, hasLength(11));
      expect(
        demo.table.hand
            .where((c) => c.rank == '9' && c.suit == CardSuit.spades)
            .map((c) => c.copy)
            .toSet(),
        {1, 2},
      );
      expect(demo.allied, isFalse);
      expect(demo.fixtureIsValid, isTrue);
    },
  );

  test(
    'purchase keeps selected quantity and no-change payment until atomic success',
    () {
      demo.loadExample(13);
      expect(demo.purchaseQuantity, 1);
      demo.primary();
      expect(demo.table.hand.any((c) => c.id == 'deck-1-spades-10'), isTrue);
      passAll();
      expect(demo.table.hand, hasLength(11));
      expect(demo.table.hand.any((c) => c.id == 'deck-1-spades-10'), isFalse);
      expect(demo.table.hand.any((c) => c.id == 'deck-1-spades-8'), isTrue);
      expect(demo.table.seats.first.finances.available, ExactPoints.integer(0));
      expect(demo.fixtureIsValid, isTrue);
      demo.loadExample(13);
      demo.setPurchaseQuantity(2);
      demo.primary();
      passAll();
      expect(demo.table.hand, hasLength(12));
      expect(demo.fixtureIsValid, isTrue);
    },
  );

  test(
    'a loan forms alliance only after successful transfer; hand is untouched',
    () {
      demo.beginLoan();
      expect(demo.allied, isFalse);
      demo.primary();
      expect(demo.allied, isFalse);
      passAll();
      expect(demo.allied, isTrue);
      expect(demo.table.hand, hasLength(18));
      expect(demo.table.seats.first.combos.single.title, 'Underground');
      expect(demo.table.seats[2].combos, isEmpty);
      expect(demo.table.seats[2].status, contains('privileges retained'));
      expect(demo.canCoup, isFalse);
      expect(demo.fixtureIsValid, isTrue);
    },
  );

  test(
    'Confinement cancels out-of-turn Ponzi but preserves Bob turn and jacks',
    () {
      demo.beginPonzi();
      final jacks = demo.table.seats[3].combos.single.cards
          .map((c) => c.id)
          .toList();
      demo.useConfinement();
      expect(demo.table.pending, isNull);
      expect(demo.activePlayer, 'Bob');
      expect(demo.table.seats.first.concealed.aceCount, 0);
      expect(demo.table.seats[3].combos.single.cards.map((c) => c.id), jacks);
      expect(demo.table.seats[3].status, contains('Confined'));
      expect(demo.ponziUsed, isTrue);
      expect(demo.table.seats.first.finances.gameScore, ExactPoints.integer(0));
    },
  );

  test(
    'immediate Coup replaces current game, then finalization expires only cash',
    () {
      demo.loadExample(17);
      expect(demo.canCoup, isTrue);
      demo.primary();
      expect(demo.boardClosed, isTrue);
      expect(demo.table.pending, isNull);
      expect(demo.table.seats.map((s) => s.finances.gameScore.format()), [
        '50',
        '-50',
        '-50',
        '-50',
      ]);
      expect(demo.table.seats.map((s) => s.finances.debt.format()), [
        '0',
        '50',
        '50',
        '50',
      ]);
      expect(
        demo.table.seats.first.finances.available,
        ExactPoints.integer(50),
      );
      demo.primary();
      expect(demo.finalized, isTrue);
      expect(
        demo.table.seats.every(
          (s) => s.finances.available == ExactPoints.integer(0),
        ),
        isTrue,
      );
      expect(demo.table.seats.map((s) => s.finances.matchScore.format()), [
        '66',
        '-36',
        '-28',
        '-45',
      ]);
      final oldGame = demo.table.gameId;
      demo.primary();
      expect(demo.gameNumber, 3);
      expect(demo.table.gameId, isNot(oldGame));
      expect(demo.table.seats.map((s) => s.finances.debt.format()), [
        '0',
        '50',
        '50',
        '50',
      ]);
      expect(
        demo.table.seats.every(
          (s) => s.finances.available == ExactPoints.integer(0),
        ),
        isTrue,
      );
      final state = demo.screen;
      demo.primary(gameId: oldGame);
      expect(demo.screen, state);
    },
  );

  test(
    'unknown attack reconnect reconciles same operation without another commit or pass',
    () {
      demo.loadExample(20);
      expect(demo.table.connection, ConnectionStateView.disconnected);
      expect(demo.table.hand, hasLength(11));
      final operation = demo.operationId;
      final count = demo.acceptedOperations;
      demo.primary();
      expect(demo.table.connection, ConnectionStateView.disconnected);
      demo.reconnect();
      expect(demo.table.pending!.responseIndex, 0);
      expect(demo.operationId, operation);
      expect(demo.acceptedOperations, count);
      demo.reconnect();
      expect(demo.acceptedOperations, count);
      passAll();
      expect(
        demo.table.seats[1].exposed.where((c) => c.suit == CardSuit.hearts),
        hasLength(1),
      );
    },
  );

  test(
    'every numbered example has valid distinct custody and no public private faces',
    () {
      for (var id = 1; id <= 20; id++) {
        demo.loadExample(id);
        expect(demo.fixtureIsValid, isTrue, reason: 'example $id');
        expect(
          demo.table.seats.every((s) => s.concealed.handCount >= 0),
          isTrue,
        );
        for (final seat in demo.table.seats) {
          for (final combo in seat.combos) {
            expect(
              combo.cards.every(
                (c) => seat.exposed.any((held) => held.id == c.id),
              ),
              isTrue,
            );
          }
        }
      }
    },
  );

  test('next-game board keeps the new game and debts after initial review', () {
    demo.loadExample(19);
    demo.primary();
    final newId = demo.table.gameId;
    demo.primary();
    expect(demo.gameNumber, 3);
    expect(demo.table.gameId, newId);
    expect(demo.table.seats.map((s) => s.finances.debt.format()), [
      '0',
      '50',
      '50',
      '50',
    ]);
  });

  test('a duplicate display name cannot grant another seat turn actions', () {
    demo.loadExample(1);
    demo.setName('Bob');
    demo.loadExample(3);
    demo.endTurn();
    expect(demo.canOpen, isFalse);
    expect(demo.canTrade, isFalse);
    expect(demo.canEndTurn, isFalse);
  });

  test(
    'ending a turn expires its offer and draws one private card for Bob',
    () {
      demo.beginTrade();
      demo.primary();
      final before = demo.table.seats[1].concealed.handCount;
      demo.endTurn();
      expect(demo.offerPending, isFalse);
      expect(demo.table.seats[1].concealed.handCount, before + 1);
      expect(demo.table.hand.any((c) => c.id == 'deck-1-clubs-Q'), isTrue);
      expect(demo.fixtureIsValid, isTrue);
    },
  );

  test('Coup eligibility remains immediate through a pending response', () {
    demo.beginPonzi(coupReady: true);
    final jacks = demo.table.seats[3].combos.single.cards
        .map((c) => c.id)
        .toList();
    expect(demo.table.pending, isNotNull);
    expect(demo.canCoup, isTrue);
    demo.declareCoup();
    expect(demo.boardClosed, isTrue);
    expect(demo.table.pending, isNull);
    expect(demo.table.seats[3].combos.single.cards.map((c) => c.id), jacks);
  });
}
