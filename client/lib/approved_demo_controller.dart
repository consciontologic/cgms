import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/foundation.dart';

enum _Move { opening, attack, trade, purchase, loan, ponzi }

/// Deterministic, viewer-filtered journeys for the approved mockups.
/// This adapter deliberately implements these fixtures, not the full game engine.
class ApprovedDemoController extends ChangeNotifier {
  ApprovedDemoController() {
    _restore();
  }

  static const examples = <String>[
    'Room and match setup',
    'Initial deal and turn order',
    'Compact main board',
    'Inspect Underground',
    'Select Hearts to open',
    'Hearts opened',
    'Select an exact attack',
    'Wait for attack responses',
    'Attack resolved',
    'Draft a private trade',
    'Offer pending during play',
    'Trade completed',
    'Choose a card purchase',
    'Purchase resolved',
    'Combo loan and alliance',
    'Confinement stops Ponzi',
    'Immediate Coup declaration',
    'Coup and settlement',
    'Finalized match standings',
    'Reconnect an uncertain action',
  ];

  late List<_Seat> _seats;
  late List<VisibleCard> _deck;
  final List<VisibleCard> _discard = [];
  final List<String> _events = [];
  final Set<String> _selected = {};
  int _instance = 0;
  late String _gameId;
  int screen = 3;
  int gameNumber = 2;
  int games = 3;
  int round = 7;
  String playerName = 'Alice';
  int _activeSeat = 1;
  bool _continuingGame = false;
  String get activePlayer => _seats[_activeSeat - 1].name;
  bool get isYourTurn => _activeSeat == 1;
  bool openingUsed = false;
  bool _clubsUsed = false;
  bool allied = false;
  bool ponziUsed = false;
  bool _tradeDone = false;
  bool offerPending = false;
  bool _purchaseDone = false;
  bool _loanOffered = false;
  int offerRevision = 0;
  int purchaseQuantity = 1;
  int acceptedOperations = 0;
  String? operationId;
  _Move? _pending;
  List<String> _responders = [];
  int _responseIndex = 0;
  ConnectionStateView _connection = ConnectionStateView.connected;
  TableViewData? _lastConfirmed;
  bool get boardClosed => screen == 18 || screen == 19;
  bool get finalized => screen == 19;
  bool get connected => _connection == ConnectionStateView.connected;
  bool get idle => connected && _pending == null && !boardClosed;
  bool get canOpen =>
      idle &&
      _activeSeat == 1 &&
      !openingUsed &&
      _heldHearts.isNotEmpty &&
      !_seats.first.exposed.any((c) => c.suit == CardSuit.hearts);
  bool get canAttack =>
      idle &&
      _activeSeat == 1 &&
      !_clubsUsed &&
      _seats.first.exposed.any((c) => c.id == 'deck-1-clubs-8') &&
      _seats[1].exposed.any((c) => c.id == 'deck-1-hearts-8');
  bool get canTrade =>
      idle &&
      _activeSeat == 1 &&
      !_tradeDone &&
      !offerPending &&
      _seats.first.hand.any((c) => c.id == 'deck-1-clubs-Q');
  bool get canBuy =>
      idle &&
      _activeSeat == 1 &&
      !_purchaseDone &&
      _seats.first.hand.any(_isCurrency);
  bool get canCoup =>
      connected &&
      !boardClosed &&
      !allied &&
      _seats.first.exposed.any((c) => c.suit == CardSuit.clubs) &&
      [
        'deck-1-diamonds-Q',
        'deck-2-diamonds-Q',
        'deck-1-hearts-Q',
        'deck-2-hearts-Q',
      ].every((id) => _seats.first.hand.any((c) => c.id == id));
  bool get canConfine =>
      connected &&
      _pending == _Move.ponzi &&
      _responseIndex == 0 &&
      _seats.first.aces.any((c) => c.suit == CardSuit.diamonds);
  bool get loanOffered => _loanOffered;
  bool get canEndTurn => idle && _activeSeat == 1 && screen >= 3 && screen < 17;
  Set<String> get selectedIds => Set.unmodifiable(_selected);
  List<VisibleCard> get _heldHearts => _seats.first.hand
      .where((c) => c.suit == CardSuit.hearts && int.tryParse(c.rank) != null)
      .toList();
  int get paymentValue => _seats.first.hand
      .where((c) => _selected.contains(c.id) && _isCurrency(c))
      .fold(0, (n, c) => n + int.parse(c.rank));
  int get purchasePrice => purchaseQuantity * 5;
  bool get purchaseLegal =>
      idle &&
      screen == 13 &&
      paymentValue >= purchasePrice &&
      _deck.length + _selected.length >= purchaseQuantity;
  bool acceptsGame(String id) => id == _gameId;
  bool _admitted(String? id) => id == null || acceptsGame(id);
  static bool _isCurrency(VisibleCard c) =>
      c.suit == CardSuit.spades && int.tryParse(c.rank) != null;

  String get title {
    if (!connected) return 'Connection lost';
    if (_pending != null)
      return _pending == _Move.ponzi
          ? 'Ponzi targets you'
          : 'Waiting for responses';
    if (_loanOffered) return 'Borrow Underground';
    return switch (screen) {
      1 => 'Your room',
      2 => 'Initial diamonds opened',
      3 => 'Your table',
      4 => 'Underground · Carol',
      5 => 'Open your Hearts',
      6 => 'Hearts opened',
      7 => 'An exact attack',
      9 => 'Attack resolved',
      10 => 'Trade with Bob · Private',
      11 => 'Offer pending',
      12 => 'Trade completed',
      13 => 'Buy cards',
      14 =>
        '$purchaseQuantity ${purchaseQuantity == 1 ? 'card' : 'cards'} drawn',
      15 => 'Alliance formed',
      16 => 'Ponzi canceled',
      17 => 'Coup ready',
      18 => '$playerName wins by Coup',
      19 => 'Match standings',
      _ => 'Your table',
    };
  }

  String get detail {
    if (!connected)
      return 'Attack sent · outcome unknown. Reconnect to check the same operation; no move is replayed.';
    if (_pending != null)
      return _pending == _Move.ponzi
          ? 'Deniz declared Ponzi outside their turn. Bob remains active. Your response can use Confinement.'
          : 'Waiting for ${_responders[_responseIndex]}. Each pass below is an explicit local simulation. Cards move only when the action resolves.';
    if (_loanOffered)
      return 'Carol offers her exposed five-king Underground on her turn. Acceptance starts responses; only success creates your alliance.';
    return switch (screen) {
      1 => 'Four local seats · choose the match length before starting.',
      2 =>
        _continuingGame
            ? '$playerName → Bob → Carol → Deniz. New-game cash starts at zero; match scores and debts are preserved.'
            : '$playerName → Bob → Carol → Deniz. All players begin protected; the round 7 example is a separate prepared position.',
      3 =>
        'Every exposed card and your complete private hand share this board.',
      4 =>
        'Five kings: +10 each round. Carol keeps combo privileges without Hearts, but cannot attack with only one current series.',
      5 =>
        'All seven number Hearts in your hand must open together. This uses your one opening action this turn.',
      6 =>
        'Seven physical cards moved to your Hearts series. Your opening allowance is used; your 11 other hand cards remain private.',
      7 =>
        'Your 8 of Clubs targets Bob’s 8 of Hearts: attack 8 = defense 8. Bob’s 6 of Hearts remains.',
      9 =>
        'Bob’s 8 of Hearts returned to the deck. Your Club stays used; no points are awarded for Hearts.',
      10 =>
        'Give Q of Clubs; request 9 of Spades. Bob’s hand stays hidden. A requested card is not a reveal of his inventory.',
      11 =>
        'No cards are reserved. Withdraw the offer, continue another action, or end your turn to expire it.',
      12 =>
        'Q of Clubs exchanged for Bob’s physical 9 of Spades. Two separate nines now remain in your hand. No alliance formed.',
      13 =>
        'Price 5 per card. Payment $paymentValue · required $purchasePrice · excess ${paymentValue - purchasePrice}; no change.',
      14 =>
        'Payment returned to the deck; the chosen quantity was drawn atomically. New cards remain in your hand.',
      15 =>
        '$playerName ↔ Carol · this game. Underground is borrowed from Carol; +10 at round end while five kings remain. Coup unavailable while allied.',
      16 =>
        'Deniz is confined until the start of their next scheduled turn. Bob’s turn continues. A of Diamonds discarded; Ponzi’s use remains spent.',
      17 =>
        'Both diamond Queens and both heart Queens · Clubs opened · unallied. Declaration ends this game immediately, with no response window.',
      18 =>
        'Board closed · settlement. The three simulated opponents have explicitly finished; your finish remains. No ordinary card awards apply.',
      19 =>
        'Game $gameNumber finalized. Available points expire; scores and unpaid debts remain. ${gameNumber < games ? 'Game ${gameNumber + 1} is next.' : 'The match is complete.'}',
      _ => 'Choose a legal action or inspect the table.',
    };
  }

  String get primaryLabel {
    if (_pending != null)
      return 'Simulate ${_responders[_responseIndex]} passing';
    if (_loanOffered) return 'Accept combo loan';
    return switch (screen) {
      1 => 'Start match',
      2 => _continuingGame ? 'Continue to this board' : 'Explore round 7 board',
      3 =>
        canOpen
            ? 'Open Hearts'
            : isYourTurn
            ? 'Choose an action below'
            : '$activePlayer’s turn',
      4 => 'Back to board',
      5 => 'Open all 7 Hearts',
      6 => 'Choose an attack',
      7 => 'Declare attack',
      9 => 'Draft a trade',
      10 => 'Send offer',
      11 => 'Simulate Bob accepting',
      12 => 'Choose a purchase',
      13 => 'Buy $purchaseQuantity ${purchaseQuantity == 1 ? 'card' : 'cards'}',
      14 || 15 || 16 => 'Back to actions',
      17 => 'Declare Coup',
      18 => 'Finish settlement',
      19 => gameNumber < games ? 'Start next game' : 'Review completed match',
      _ => 'Continue',
    };
  }

  bool get primaryEnabled =>
      connected &&
      (screen != 13 || _pending != null || purchaseLegal) &&
      (screen != 17 || canCoup) &&
      !(screen == 19 && gameNumber >= games) &&
      !(screen == 3 && !canOpen);

  TableViewData get table {
    final snapshot = _lastConfirmed;
    if (snapshot != null)
      return TableViewData(
        gameId: snapshot.gameId,
        round: snapshot.round,
        totalRounds: 13,
        turnLabel: 'Game $gameNumber · Last confirmed board',
        seats: snapshot.seats,
        hand: snapshot.hand,
        events: snapshot.events,
        objective: detail,
        actionLabel: 'Awaiting confirmation',
        actionExplanation: detail,
        phaseLabel: 'Action outcome unknown',
        connection: _connection,
      );
    return TableViewData(
      gameId: _gameId,
      round: round,
      totalRounds: 13,
      turnLabel:
          'Game $gameNumber of $games · ${boardClosed ? (finalized ? 'Finalized' : 'Board closed') : '$activePlayer’s turn'}',
      seats: List.unmodifiable(
        _seats.map(
          (s) => SeatView(
            id: 'seat-${s.number}',
            name: s.name,
            number: s.number,
            exposed: List.unmodifiable(s.exposed),
            combos: List.unmodifiable(s.combos),
            concealed: ConcealedCards(
              handCount: s.hand.length,
              aceCount: s.aces.length,
            ),
            finances: FinancialView(
              matchScore: ExactPoints.integer(s.prior + s.score),
              gameScore: ExactPoints.integer(s.score),
              available: ExactPoints.integer(s.available),
              debt: ExactPoints.integer(s.debt),
            ),
            status: s.status,
            isSelf: s.number == 1,
          ),
        ),
      ),
      hand: List.unmodifiable(_seats.first.hand),
      events: List.unmodifiable(_events.reversed),
      objective: detail,
      actionLabel: primaryLabel,
      actionExplanation: detail,
      phaseLabel: title,
      actionEnabled: primaryEnabled,
      connection: _connection,
      selectedCardId: _selected.length == 1 ? _selected.first : null,
      pending: _pending == null
          ? null
          : PendingView(
              title: title,
              detail: detail,
              responders: List.unmodifiable(_responders),
              responseIndex: _responseIndex,
            ),
    );
  }

  void primary({String? gameId}) {
    if (!_admitted(gameId) || !connected) return;
    if (_pending != null) {
      _pass();
      return;
    }
    if (_loanOffered) {
      _loanOffered = false;
      _start(_Move.loan, ['Bob', 'Carol', 'Deniz']);
      return;
    }
    switch (screen) {
      case 1:
        _restore(initial: true, firstGame: true);
        screen = 2;
      case 2:
        if (_continuingGame) {
          _drawFor(_seats.first);
          screen = 3;
        } else {
          _restore();
        }
      case 3:
        beginOpening();
        return;
      case 4:
        screen = 3;
      case 5:
        if (canOpen) _start(_Move.opening, ['Bob', 'Carol', 'Deniz']);
        return;
      case 6:
        beginAttack();
        return;
      case 7:
        if (canAttack) _startAttack();
        return;
      case 9:
        beginTrade();
        return;
      case 10:
        if (canTrade) {
          offerPending = true;
          offerRevision++;
          screen = 11;
          _selected.clear();
          _events.add(
            'Private offer revision $offerRevision sent to Bob; no cards reserved.',
          );
        }
      case 11:
        if (offerPending) {
          offerPending = false;
          _start(_Move.trade, ['Carol', 'Deniz', playerName]);
        }
        return;
      case 12:
        beginPurchase();
        return;
      case 13:
        if (purchaseLegal) _start(_Move.purchase, ['Bob', 'Carol', 'Deniz']);
        return;
      case 14 || 15 || 16:
        screen = 3;
        _selected.clear();
      case 17:
        if (canCoup) _declareCoup();
      case 18:
        _finish();
      case 19:
        if (gameNumber < games) _nextGame();
    }
    notifyListeners();
  }

  void beginOpening() {
    if (!canOpen) return;
    screen = 5;
    _selected
      ..clear()
      ..addAll(_heldHearts.map((c) => c.id));
    notifyListeners();
  }

  void beginAttack() {
    if (!canAttack) return;
    screen = 7;
    _selected
      ..clear()
      ..addAll(['deck-1-clubs-8', 'deck-1-hearts-8']);
    notifyListeners();
  }

  void beginTrade() {
    if (!canTrade) return;
    screen = 10;
    _selected
      ..clear()
      ..add('deck-1-clubs-Q');
    notifyListeners();
  }

  void beginPurchase() {
    if (!canBuy) return;
    screen = 13;
    purchaseQuantity = 1;
    _selected
      ..clear()
      ..add(
        _seats.first.hand
            .where(_isCurrency)
            .firstWhere(
              (c) => c.rank == '10',
              orElse: () => _seats.first.hand.firstWhere(_isCurrency),
            )
            .id,
      );
    notifyListeners();
  }

  void setPurchaseQuantity(int quantity) {
    if (!idle || screen != 13 || quantity < 1 || quantity > 2) return;
    purchaseQuantity = quantity;
    notifyListeners();
  }

  bool selectHand(String id) {
    if (!idle || !_seats.first.hand.any((c) => c.id == id)) return false;
    if (screen == 13 &&
        _isCurrency(_seats.first.hand.firstWhere((c) => c.id == id))) {
      if (!_selected.remove(id)) _selected.add(id);
      notifyListeners();
      return true;
    } else if (screen == 3 && _heldHearts.any((c) => c.id == id) && canOpen) {
      beginOpening();
      return true;
    }
    return false;
  }

  bool selectPublic(String id) {
    if (canAttack && (id == 'deck-1-clubs-8' || id == 'deck-1-hearts-8')) {
      beginAttack();
      return true;
    }
    return false;
  }

  void clearSelection() {
    if (!idle) return;
    _selected.clear();
    screen = 3;
    notifyListeners();
  }

  void backToBoard() {
    if (!idle) return;
    screen = 3;
    _selected.clear();
    notifyListeners();
  }

  void withdrawOffer() {
    if (!idle || !offerPending) return;
    offerPending = false;
    screen = 10;
    _events.add('Unanswered offer withdrawn; no cards moved.');
    notifyListeners();
  }

  void declineOffer() {
    if (!idle || !offerPending) return;
    offerPending = false;
    screen = 10;
    _events.add('Bob explicitly declined; no cards moved.');
    notifyListeners();
  }

  void endTurn() {
    if (!canEndTurn) return;
    if (offerPending) {
      offerPending = false;
      _events.add('Offer expired at the originating turn boundary.');
    }
    _activeSeat = 2;
    _drawFor(_seats[1]);
    screen = 3;
    _selected.clear();
    _events.add(
      'Your turn ended. Bob is now active; no timer or automatic pass is used.',
    );
    notifyListeners();
  }

  void _drawFor(_Seat seat) {
    if (_deck.isEmpty && _discard.isNotEmpty) {
      _deck.addAll(_discard);
      _discard.clear();
    }
    if (_deck.isEmpty) return;
    final card = _deck.removeAt(0);
    (card.rank == 'A' ? seat.aces : seat.hand).add(card);
    _events.add('${seat.name} drew one card at the start of their turn.');
  }

  void _start(_Move move, List<String> responders) {
    _pending = move;
    _responders = responders;
    _responseIndex = 0;
    operationId = '$_gameId-operation-${++acceptedOperations}';
    _events.add(
      '${move.name} accepted; operation $acceptedOperations awaits explicit responses.',
    );
    notifyListeners();
  }

  void _startAttack() {
    _clubsUsed = true;
    _markClubUsed();
    screen = 8;
    _start(_Move.attack, ['Bob', 'Carol', 'Deniz']);
  }

  void _markClubUsed() {
    final i = _seats.first.exposed.indexWhere((c) => c.id == 'deck-1-clubs-8');
    if (i >= 0) {
      final c = _seats.first.exposed[i];
      _seats.first.exposed[i] = VisibleCard(
        id: c.id,
        rank: c.rank,
        suit: c.suit,
        copy: c.copy,
        status: 'Used this turn',
      );
    }
  }

  void _pass() {
    _events.add(
      '${_responders[_responseIndex]} explicitly passed (local simulation).',
    );
    if (++_responseIndex < _responders.length) {
      notifyListeners();
      return;
    }
    final move = _pending!;
    _pending = null;
    _responders = [];
    _responseIndex = 0;
    switch (move) {
      case _Move.opening:
        final cards = _heldHearts;
        _seats.first.hand.removeWhere((c) => cards.any((h) => h.id == c.id));
        _seats.first.exposed.addAll(cards.reversed);
        _seats.first.status = 'Hearts opened · unprotected';
        openingUsed = true;
        screen = 6;
        _events.add(
          'All seven number Hearts opened. Hand 18 → 11; one opening action used.',
        );
      case _Move.attack:
        final removed = _seats[1].exposed.firstWhere(
          (c) => c.id == 'deck-1-hearts-8',
        );
        _seats[1].exposed.remove(removed);
        _deck.add(removed);
        screen = 9;
        _events.add(
          '8 of Hearts returned. Attacking 8 of Clubs remains used. No points.',
        );
      case _Move.trade:
        final give = _seats.first.hand.firstWhere(
          (c) => c.id == 'deck-1-clubs-Q',
        );
        final receive = _seats[1].hand.firstWhere(
          (c) => c.id == 'deck-2-spades-9',
        );
        _seats.first.hand
          ..remove(give)
          ..add(receive);
        _seats[1].hand
          ..remove(receive)
          ..add(give);
        _tradeDone = true;
        screen = 12;
        _events.add(
          'Private trade completed atomically: Q of Clubs for 9 of Spades, copy 2. No alliance.',
        );
      case _Move.purchase:
        final payment = _seats.first.hand
            .where((c) => _selected.contains(c.id) && _isCurrency(c))
            .toList();
        _seats.first.hand.removeWhere((c) => payment.any((p) => p.id == c.id));
        _deck.addAll(payment);
        // Reserved deterministic outcomes stand in for an authority's shuffle.
        final draws = _deck.take(purchaseQuantity).toList();
        _deck.removeRange(0, purchaseQuantity);
        _seats.first.hand.addAll(draws);
        _purchaseDone = true;
        screen = 14;
        _events.add(
          'Purchased $purchaseQuantity cards with ${payment.map((c) => c.label).join(', ')}; excess has no change. Deterministic local draw.',
        );
      case _Move.loan:
        final combo = _seats[2].combos.single;
        _seats[2].exposed.removeWhere(
          (c) => combo.cards.any((k) => k.id == c.id),
        );
        _seats[2].combos.clear();
        _seats.first.exposed.addAll(combo.cards);
        _seats.first.combos.add(_underground(combo.cards, borrowed: true));
        _seats[2].status =
            'Combo privileges retained · allied with $playerName';
        _seats.first.status = 'Allied with Carol · Underground borrowed';
        allied = true;
        screen = 15;
        _events.add(
          'Loan succeeded: physical custody and exclusive alliance changed together. Carol keeps earned privileges.',
        );
      case _Move.ponzi:
        final combo = _seats[3].combos.single;
        _seats[3].exposed.removeWhere(
          (c) => combo.cards.any((j) => j.id == c.id),
        );
        _deck.addAll(combo.cards);
        _seats[3].combos.clear();
        screen = 3;
        _events.add(
          'Ponzi resolved against zero available points. No debt created; four jacks returned.',
        );
    }
    _selected.clear();
    assert(fixtureIsValid);
    notifyListeners();
  }

  void beginLoan() {
    _restore();
    _activeSeat = 3;
    round = 8;
    _loanOffered = true;
    _events.add('Independent loan branch, on Carol’s turn.');
    notifyListeners();
  }

  void beginPonzi({bool coupReady = false}) {
    _restore(coup: coupReady);
    _activeSeat = 2;
    round = 8;
    ponziUsed = true;
    _start(_Move.ponzi, [playerName, 'Bob', 'Carol']);
  }

  void useConfinement() {
    if (!canConfine) return;
    final ace = _seats.first.aces.firstWhere(
      (c) => c.suit == CardSuit.diamonds,
    );
    _seats.first.aces.remove(ace);
    _discard.add(ace);
    _pending = null;
    _responders = [];
    _responseIndex = 0;
    _seats[3].status = 'Confined · until next scheduled turn';
    screen = 16;
    final old = _seats[3].combos.single;
    _seats[3].combos[0] = ComboView(
      id: old.id,
      title: old.title,
      stateLabel: 'Used · canceled',
      description: old.description,
      cards: old.cards,
    );
    _events.add(
      'Confinement canceled Deniz’s out-of-turn Ponzi. Bob continues. Jacks remain; allowance spent.',
    );
    assert(fixtureIsValid);
    notifyListeners();
  }

  void declareCoup({String? gameId}) {
    if (!_admitted(gameId) || !canCoup) return;
    _declareCoup();
    notifyListeners();
  }

  void _declareCoup() {
    _pending = null;
    _responders = [];
    _responseIndex = 0;
    for (final s in _seats) {
      s.available = 0;
      s.score = s.number == 1 ? 50 : -50;
      if (s.number == 1)
        s.available = 50;
      else
        s.debt += 50;
    }
    offerPending = false;
    _selected.clear();
    screen = 18;
    acceptedOperations++;
    operationId = '$_gameId-operation-$acceptedOperations';
    _events.add(
      'Coup ended the board immediately. No response window or ordinary card awards.',
    );
    for (final s in _seats.skip(1)) {
      _events.add(
        '${s.name} explicitly finished settlement (local simulation).',
      );
    }
  }

  void _finish() {
    if (screen != 18) return;
    for (final s in _seats) {
      s.available = 0;
    }
    screen = 19;
    _events.add(
      'Your settlement finish recorded. Game $gameNumber financially finalized; only available cash expires.',
    );
  }

  void _nextGame() {
    final previous = _seats.map((s) => (s.prior + s.score, s.debt)).toList();
    final next = gameNumber + 1;
    _restore(initial: true);
    gameNumber = next;
    _continuingGame = true;
    screen = 2;
    for (var i = 0; i < 4; i++) {
      _seats[i].prior = previous[i].$1;
      _seats[i].debt = previous[i].$2;
    }
    _events.add(
      'New immutable game instance. Scores and debts retained; available points start at zero.',
    );
  }

  void setGames(int value) {
    if (screen != 1 || ![3, 5, 7].contains(value)) return;
    games = value;
    notifyListeners();
  }

  void setName(String value) {
    if (screen != 1) return;
    playerName = value.trim().isEmpty ? 'Alice' : value.trim();
    _seats.first.name = playerName;
    _activeSeat = 1;
    notifyListeners();
  }

  void startMatch() {
    _restore(initial: true, firstGame: true);
    screen = 2;
    notifyListeners();
  }

  void reconnect() {
    if (connected) return;
    _connection = ConnectionStateView.connected;
    _lastConfirmed = null;
    screen = _pending == _Move.attack ? 8 : screen;
    _events.add(
      'Reconciled the same operation. No replay and no automatic response.',
    );
    notifyListeners();
  }

  void loadExample(int id) {
    if (id < 1 || id > 20) throw ArgumentError.value(id, 'id');
    _restore(initial: id == 2);
    if (id == 1) {
      screen = 1;
      notifyListeners();
      return;
    }
    if (id == 2) {
      screen = 2;
      notifyListeners();
      return;
    }
    if (id >= 5 && id <= 14 || id == 20) {
      beginOpening();
      if (id >= 6) {
        primary();
        _complete();
      }
      if (id >= 7) {
        beginAttack();
      }
      if (id >= 8 && id <= 14) {
        primary();
      }
      if (id >= 9 && id <= 14) {
        _complete();
      }
      if (id >= 10 && id <= 14) {
        beginTrade();
      }
      if (id >= 11 && id <= 14) {
        primary();
      }
      if (id >= 12 && id <= 14) {
        primary();
        _complete();
      }
      if (id >= 13 && id <= 14) {
        beginPurchase();
      }
      if (id == 14) {
        primary();
        _complete();
      }
    }
    if (id == 15) {
      beginLoan();
      primary();
      _complete();
    }
    if (id == 16) {
      beginPonzi();
      useConfinement();
    }
    if (id >= 17 && id <= 19) {
      _restore(coup: true);
      _activeSeat = 2;
      screen = 17;
      if (id >= 18) primary();
      if (id == 19) primary();
    }
    if (id == 20) {
      _lastConfirmed = table;
      _startAttack();
      _connection = ConnectionStateView.disconnected;
    }
    screen = id;
    notifyListeners();
  }

  void _complete() {
    while (_pending != null) {
      _pass();
    }
  }

  bool get fixtureIsValid {
    final cards = [
      ..._seats.expand((s) => [...s.hand, ...s.exposed, ...s.aces]),
      ..._deck,
      ..._discard,
    ];
    return cards.length == 104 &&
        cards.map((c) => c.id).toSet().length == 104 &&
        cards.every((c) => c.copy == 1 || c.copy == 2) &&
        _seats.every(
          (s) => s.combos.every(
            (combo) =>
                combo.cards.every((c) => s.exposed.any((h) => h.id == c.id)),
          ),
        );
  }

  void _restore({
    bool initial = false,
    bool coup = false,
    bool firstGame = false,
  }) {
    _gameId = 'cgmsart-demo-${++_instance}';
    screen = initial ? 2 : 3;
    gameNumber = firstGame ? 1 : 2;
    round = initial ? 1 : 7;
    _activeSeat = 1;
    _continuingGame = false;
    openingUsed = false;
    _clubsUsed = false;
    allied = false;
    ponziUsed = false;
    _tradeDone = false;
    offerPending = false;
    _purchaseDone = false;
    _loanOffered = false;
    offerRevision = 0;
    purchaseQuantity = 1;
    acceptedOperations = 0;
    operationId = null;
    _pending = null;
    _responders = [];
    _responseIndex = 0;
    _connection = ConnectionStateView.connected;
    _lastConfirmed = null;
    _selected.clear();
    _discard.clear();
    _events.clear();
    final own = [
      _c(
        coup ? 'Q' : 'Q',
        coup ? CardSuit.diamonds : CardSuit.clubs,
        copy: coup ? 2 : 1,
      ),
      _c('Q', CardSuit.diamonds),
      _c('Q', CardSuit.hearts),
      _c('J', CardSuit.hearts),
      _c('J', CardSuit.diamonds),
      coup ? _c('Q', CardSuit.hearts, copy: 2) : _c('K', CardSuit.spades),
      for (final r in [2, 3, 4, 5, 7, 9, 10]) _c('$r', CardSuit.hearts),
      for (final r in [2, 4, 6, 9, 10]) _c('$r', CardSuit.spades),
    ];
    final kings = [
      _c('K', CardSuit.hearts),
      _c('K', CardSuit.hearts, copy: 2),
      _c('K', CardSuit.clubs),
      _c('K', CardSuit.diamonds),
      _c('K', CardSuit.spades),
    ];
    // The hand's spade King uses the second physical copy.
    if (!coup) own[5] = _c('K', CardSuit.spades, copy: 2);
    final jacks = [
      _c('J', CardSuit.clubs),
      _c('J', CardSuit.clubs, copy: 2),
      _c('J', CardSuit.spades),
      _c('J', CardSuit.spades, copy: 2),
    ];
    _seats = [
      _Seat(
        playerName,
        1,
        [
          _c('10', CardSuit.diamonds),
          _c('8', CardSuit.diamonds),
          if (!initial) _c('8', CardSuit.clubs),
        ],
        own,
        initial ? [] : [_c('A', CardSuit.diamonds)],
        16,
      ),
      _Seat(
        'Bob',
        2,
        [
          _c('9', CardSuit.diamonds),
          _c('7', CardSuit.diamonds),
          if (!initial) ...[_c('8', CardSuit.hearts), _c('6', CardSuit.hearts)],
        ],
        [],
        initial ? [] : [_c('A', CardSuit.hearts)],
        14,
      ),
      _Seat(
        'Carol',
        3,
        [
          _c('8', CardSuit.diamonds, copy: 2),
          _c('6', CardSuit.diamonds),
          if (!initial) ...kings,
        ],
        [],
        initial ? [] : [_c('A', CardSuit.spades)],
        22,
      ),
      _Seat(
        'Deniz',
        4,
        [
          _c('7', CardSuit.diamonds, copy: 2),
          _c('5', CardSuit.diamonds),
          if (!initial) ...[_c('6', CardSuit.clubs), ...jacks],
        ],
        [],
        initial
            ? []
            : [_c('A', CardSuit.clubs), _c('A', CardSuit.clubs, copy: 2)],
        5,
      ),
    ];
    if (firstGame)
      for (final s in _seats) {
        s.prior = 0;
      }
    if (!initial) {
      _seats[2].combos.add(_underground(kings));
      _seats[3].combos.add(
        ComboView(
          id: 'deniz-ponzi',
          title: 'Ponzi',
          stateLabel: '4 jacks',
          description:
              'Both club Jacks and both spade Jacks. Steals eligible available current-game points; response opportunities still apply.',
          cards: List.unmodifiable(jacks),
        ),
      );
    }
    final occupied = _seats
        .expand((s) => [...s.hand, ...s.exposed, ...s.aces])
        .map((c) => c.id)
        .toSet();
    final free = [
      for (var copy = 1; copy <= 2; copy++)
        for (final suit in CardSuit.values)
          for (final rank in [
            'A',
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
            'K',
          ])
            if (!occupied.contains('deck-$copy-${suit.name}-$rank'))
              _c(rank, suit, copy: copy),
    ];
    final reserved = {'deck-1-spades-8', 'deck-1-spades-3'};
    final bobNine = free.firstWhere((c) => c.id == 'deck-2-spades-9');
    free.remove(bobNine);
    _seats[1].hand.add(bobNine);
    for (var i = 1; i < 4; i++) {
      final count = initial ? 18 : (i == 1 ? 16 : 17);
      while (_seats[i].hand.length < count) {
        final c = free.firstWhere(
          (c) =>
              !reserved.contains(c.id) &&
              c.rank != 'A' &&
              (!initial ||
                  c.suit != CardSuit.diamonds ||
                  int.tryParse(c.rank) == null),
        );
        free.remove(c);
        _seats[i].hand.add(c);
      }
    }
    _deck = [
      for (final id in reserved) free.firstWhere((c) => c.id == id),
      ...free.where((c) => !reserved.contains(c.id)),
    ];
    for (final s in _seats) {
      s.status = initial
          ? 'Initial protection'
          : s.number == 3
          ? 'Underground · 1 series'
          : 'Initial protection ended';
    }
    _events.add(
      initial
          ? 'Prepared initial deal: public Diamonds determine initiative.'
          : 'Prepared game 2, round 7 position · local demonstration.',
    );
    assert(fixtureIsValid);
  }

  ComboView _underground(
    List<VisibleCard> cards, {
    bool borrowed = false,
  }) => ComboView(
    id: 'underground',
    title: 'Underground',
    stateLabel: borrowed ? 'Borrowed · 5 kings' : '5 kings',
    description:
        'Five allocated kings earn +10 at round end. Hearts are not required. Earned combo/open-card-trade privileges last this game; attacks still need two current number series.${borrowed ? ' Borrowed from Carol; exclusive alliance.' : ''}',
    cards: List.unmodifiable(cards),
  );
  static VisibleCard _c(String rank, CardSuit suit, {int copy = 1}) =>
      VisibleCard(
        id: 'deck-$copy-${suit.name}-$rank',
        rank: rank,
        suit: suit,
        copy: copy,
      );
}

class _Seat {
  _Seat(this.name, this.number, this.exposed, this.hand, this.aces, this.prior);
  String name;
  final int number;
  final List<VisibleCard> exposed, hand, aces;
  final List<ComboView> combos = [];
  int prior, score = 0, available = 0, debt = 0;
  String status = '';
}
