import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/foundation.dart';

enum DemoStage {
  ready,
  response,
  heartsResolved,
  clubsResolved,
  tradeResponse,
  tradeResolved,
  boardClosed,
  finalized,
  empty,
  densityReady,
  densityResolved,
}

enum _Action { hearts, clubs, trade, seriesAddition }

/// A finite, local teaching script, not a CGMS rules engine.
///
/// The package receives only this adapter's authorized projection. Opponent
/// private cards, script progression and simulated commands stay in the host.
class DemoController extends ChangeNotifier {
  DemoController() {
    _restore();
  }

  static const scenarioIds = [
    'ready',
    'pending',
    'trade',
    'results',
    'disconnected',
    'error',
    'loading',
    'empty',
    'hand-10',
    'hand-15',
    'combos',
  ];

  String _name = 'You';
  int _games = 3;
  int _instance = 0;
  late String _gameId;
  late List<_DemoSeat> _seats;
  late List<String> _events;
  DemoStage _stage = DemoStage.ready;
  ConnectionStateView _connection = ConnectionStateView.connected;
  _Action? _pendingAction;
  List<String> _responders = const [];
  int _responseIndex = 0;
  String? _selectedCardId;
  bool _heartsCleared = false;
  bool _clubsCleared = false;
  String _tradeStatus = 'No offer';
  int _offerRevision = 0;
  String? _offerId;
  bool _tradeResolved = false;
  bool _densityReview = false;
  bool _densityAdditionComplete = false;
  List<ComboView> _handCombos = const [];

  DemoStage get stage => _stage;
  int get games => _games;
  String get playerName => _name;
  String get tradeStatus => _tradeStatus;
  String? get offerId => _offerId;
  int get offerRevision => _offerRevision;
  bool get tradeOffered => _tradeStatus == 'Offered';
  bool get tradeResolved => _tradeResolved;
  bool get boardClosed =>
      _stage == DemoStage.boardClosed || _stage == DemoStage.finalized;
  bool get finalized => _stage == DemoStage.finalized;
  bool get densityReview => _densityReview;
  bool get canTrade =>
      !densityReview && _idle && !tradeOffered && !_tradeResolved;
  bool get canEndTurn => !densityReview && _idle;
  bool get _connected => _connection == ConnectionStateView.connected;
  bool get _idle =>
      _connected &&
      !boardClosed &&
      _pendingAction == null &&
      _stage != DemoStage.empty;

  /// An example admission check. Production adapters must authenticate and
  /// deduplicate real commands; this local host has no network transport.
  bool acceptsGame(String gameId) => gameId == _gameId;

  TableViewData get table {
    final empty = _stage == DemoStage.empty;
    final action = _actionPresentation;
    return TableViewData(
      gameId: _gameId,
      round: densityReview ? 8 : 13,
      totalRounds: 13,
      turnLabel: boardClosed
          ? 'Game 1 of $_games · ${finalized ? 'financially finalized' : 'settlement'}'
          : 'Game 1 of $_games · ${_name == 'You' ? 'Your turn' : '$_name’s turn'} · ${densityReview ? 'density review · round 8' : 'last turn of round 13'}',
      seats: empty
          ? const []
          : _seats.map(_projectSeat).toList(growable: false),
      hand: empty
          ? const []
          : List<VisibleCard>.unmodifiable(_seats.first.hand),
      handCombos: empty ? const [] : _handCombos,
      events: List<String>.unmodifiable(_events.reversed),
      objective: _objective,
      actionLabel: action.$1,
      actionExplanation: action.$2,
      actionEnabled: _connected && action.$3,
      phaseLabel: _phaseLabel,
      selectedCardId: _selectedCardId,
      connection: _connection,
      pending: _pendingAction == null
          ? null
          : PendingView(
              title: _pendingAction == _Action.seriesAddition
                  ? 'Series addition · responses'
                  : _pendingAction == _Action.trade
                  ? 'Accepted trade · responses'
                  : 'Declared attack · responses',
              detail: _pendingAction == _Action.seriesAddition
                  ? 'Add your later-acquired 7 of Clubs to your already-open Clubs series on your own turn. This does not consume the new-series opening allowance. The card moves after the ordinary response window resolves.'
                  : _pendingAction == _Action.trade
                  ? 'Ivo accepted revision $_offerRevision. Cards move only after all responses.'
                  : _pendingAction == _Action.hearts
                  ? '7 of Clubs targets Mara’s 4 of Hearts + 3 of Hearts. Attack 7 = defense 7. The physical Club is now used.'
                  : '5 of Clubs targets Mara’s 5 of Clubs. Attack 5 = defense 5. Both Clubs return only on success.',
              responders: List<String>.unmodifiable(_responders),
              responseIndex: _responseIndex,
            ),
    );
  }

  SeatView _projectSeat(_DemoSeat seat) => SeatView(
    id: seat.id,
    name: seat.name,
    number: seat.number,
    exposed: List<VisibleCard>.unmodifiable(seat.exposed),
    combos: List<ComboView>.unmodifiable(seat.combos),
    concealed: ConcealedCards(
      handCount: seat.hand.length,
      aceCount: seat.aceCount,
    ),
    finances: FinancialView(
      matchScore: ExactPoints.integer(seat.score),
      gameScore: ExactPoints.integer(seat.score),
      available: ExactPoints.integer(seat.available),
      debt: ExactPoints.integer(seat.debt),
    ),
    status: densityReview
        ? 'Four open series · initial protection ended'
        : seat.number == 1
        ? boardClosed
              ? 'Ordinary ending · no threshold bonus'
              : 'Two open series · may attack'
        : seat.number == 2
        ? 'Initial protection stays ended'
        : 'Initial protection · cannot be attacked',
    isSelf: seat.number == 1,
  );

  (String, String, bool) get _actionPresentation {
    if (_stage == DemoStage.empty) {
      return (
        'No action available',
        'Choose a populated review scenario.',
        false,
      );
    }
    if (_pendingAction != null) {
      final responder = _responders[_responseIndex];
      return (
        responder == 'You'
            ? 'Simulate your pass'
            : 'Simulate $responder’s pass',
        'An explicit local pass advances one response. Disconnecting never passes.',
        true,
      );
    }
    if (densityReview) {
      if (_densityAdditionComplete) {
        return (
          'Card addition complete',
          'The same physical 7 of Clubs is now public. Remaining cards are available for inspection; trade and end scoring are outside this density review.',
          false,
        );
      }
      final legalSelection = _selectedCardId == 'deck-1-clubs-7';
      return (
        legalSelection
            ? 'Add 7 of Clubs to your series'
            : _selectedCardId == null
            ? 'Select the 7 of Clubs in your hand'
            : 'Selected for inspection',
        legalSelection
            ? 'This later-acquired card may join your existing Clubs series on your turn. No new series opens and no points are awarded.'
            : 'Select any hand card to review it. This scenario implements only adding the 7 of Clubs to your already-open series; trade and end scoring are unavailable.',
        legalSelection,
      );
    }
    if (finalized) {
      return (
        'Game finalized',
        _games == 1
            ? 'The one-game demo match is complete. Reset to play again.'
            : 'Game 1 is complete. Later games are outside this scripted demo.',
        false,
      );
    }
    if (boardClosed) {
      return (
        'Finish settlement',
        'Confirm your finish; the three local opponents explicitly finish too. No promises remain.',
        true,
      );
    }
    if (!_heartsCleared || !_clubsCleared) {
      final rank = _heartsCleared ? '5' : '7';
      return (
        _selectedCardId == null
            ? 'Select the $rank of Clubs on your table'
            : 'Declare $rank of Clubs attack',
        _heartsCleared
            ? 'Hearts are cleared. Match Mara’s 5 of Clubs with your unused physical 5 of Clubs.'
            : 'Select your exposed 7 of Clubs to match Mara’s 4 of Hearts + 3 of Hearts exactly.',
        _selectedCardId != null,
      );
    }
    if (tradeOffered) {
      return (
        'Awaiting Ivo’s decision',
        'This offer reserves no cards. You may withdraw it or end your turn to expire it.',
        false,
      );
    }
    if (!_tradeResolved && _tradeStatus == 'No offer') {
      return (
        'Offer a trade',
        'Offer your concealed 5 of Hearts for Ivo’s 6 of Hearts.',
        true,
      );
    }
    return (
      'End final turn',
      'Close round 13 and award retained number Diamonds and Spade royals.',
      true,
    );
  }

  String get _phaseLabel {
    if (finalized) return 'Final results';
    if (boardClosed) return 'Board closed · settlement';
    if (_pendingAction != null) return 'Waiting for a response';
    if (tradeOffered) return 'Trade proposal · nonblocking';
    if (_stage == DemoStage.empty) return 'Empty state';
    if (densityReview) return 'Dense table · hand review';
    return 'Your action';
  }

  String get _objective {
    if (finalized)
      return 'Recorded scores stay; spendable balances close at finalization.';
    if (boardClosed)
      return 'Review scores, debt repayment and the ordinary ending before finishing settlement.';
    if (_stage == DemoStage.empty)
      return 'No match state has been supplied by this adapter.';
    if (_pendingAction != null)
      return 'Resolve the response window in fixed seat order.';
    if (densityReview) {
      return _densityAdditionComplete
          ? 'Your hand is one card smaller and the same physical card has joined your public Clubs series.'
          : 'Inspect the larger hand, then add its 7 of Clubs to your existing series.';
    }
    if (tradeOffered)
      return 'Ivo can accept or decline your exact offer; acceptance begins a transfer action.';
    if (!_heartsCleared)
      return 'Clear Mara’s Hearts before attacking her Clubs.';
    if (!_clubsCleared)
      return 'Use your remaining 5 of Clubs to earn 3 points and repay part of your debt.';
    if (!_tradeResolved)
      return 'Try a private card trade, then end the last turn.';
    return 'Your 6 of Hearts is now in hand. End the turn to see the final accounts.';
  }

  void configure({required String name, required int games}) {
    if (games != 1 && games != 3) {
      throw ArgumentError.value(
        games,
        'games',
        'Demo supports one or three games.',
      );
    }
    _name = name.trim().isEmpty ? 'You' : name.trim();
    _games = games;
    _restore();
    notifyListeners();
  }

  void selectCard(String id) {
    if (!_idle) return;
    if (densityReview) {
      if (!_seats.first.hand.any((card) => card.id == id)) return;
      _selectedCardId = _selectedCardId == id ? null : id;
      notifyListeners();
      return;
    }
    final legalId = !_heartsCleared
        ? 'deck-1-clubs-7'
        : !_clubsCleared
        ? 'deck-1-clubs-5'
        : null;
    if (id != legalId) return;
    _selectedCardId = _selectedCardId == id ? null : id;
    notifyListeners();
  }

  void clearSelection() {
    if (!_idle || _selectedCardId == null) return;
    _selectedCardId = null;
    notifyListeners();
  }

  void primaryAction() {
    if (!_connected) return;
    if (_pendingAction != null) {
      _pass();
      return;
    }
    if (boardClosed) {
      finishSettlement();
      return;
    }
    if (!_idle) return;
    if (densityReview) {
      _declareSeriesAddition();
      return;
    }
    if (!_heartsCleared || !_clubsCleared) {
      if (_selectedCardId == null) return;
      _declareAttack();
    } else if (tradeOffered) {
      return;
    } else if (!_tradeResolved && _tradeStatus == 'No offer') {
      offerTrade();
    } else {
      endTurn();
    }
  }

  void _declareSeriesAddition() {
    if (_densityAdditionComplete || _selectedCardId != 'deck-1-clubs-7') {
      return;
    }
    if (!_seats.first.hand.any((card) => card.id == _selectedCardId)) return;
    _pendingAction = _Action.seriesAddition;
    _responders = ['Mara', 'Ivo', 'Nia'];
    _responseIndex = 0;
    _stage = DemoStage.response;
    _events.add(
      '$_name declared adding the later-acquired 7 of Clubs to the existing Clubs series. No new series opening or score award.',
    );
    notifyListeners();
  }

  void _declareAttack() {
    final action = _heartsCleared ? _Action.clubs : _Action.hearts;
    final id = action == _Action.hearts ? 'deck-1-clubs-7' : 'deck-1-clubs-5';
    if (_selectedCardId != id) return;
    final index = _seats.first.exposed.indexWhere((card) => card.id == id);
    if (index < 0 || _seats.first.exposed[index].status.isNotEmpty) return;
    final card = _seats.first.exposed[index];
    _seats.first.exposed[index] = VisibleCard(
      id: card.id,
      rank: card.rank,
      suit: card.suit,
      copy: card.copy,
      status: 'Used this turn',
    );
    _pendingAction = action;
    _responders = ['Mara', 'Ivo', 'Nia'];
    _responseIndex = 0;
    _stage = DemoStage.response;
    _events.add(
      action == _Action.hearts
          ? '$_name declared 7 of Clubs against Mara’s 4 of Hearts + 3 of Hearts: 7 = 7.'
          : '$_name declared 5 of Clubs against Mara’s 5 of Clubs: 5 = 5.',
    );
    notifyListeners();
  }

  void _pass() {
    _events.add(
      '${_responders[_responseIndex]} explicitly passed (local simulation).',
    );
    if (_responseIndex + 1 < _responders.length) {
      _responseIndex++;
    } else {
      _resolvePending();
    }
    notifyListeners();
  }

  void _resolvePending() {
    switch (_pendingAction!) {
      case _Action.hearts:
        _seats[1].exposed.removeWhere((card) => card.suit == CardSuit.hearts);
        _heartsCleared = true;
        _stage = DemoStage.heartsResolved;
        _events.add(
          'Resolved: Mara’s 4 of Hearts and 3 of Hearts returned to the deck. Your 7 of Clubs stays exposed and used; no combat points.',
        );
      case _Action.clubs:
        _seats.first.exposed.removeWhere((card) => card.id == 'deck-1-clubs-5');
        _seats[1].exposed.removeWhere((card) => card.id == 'deck-2-clubs-5');
        _clubsCleared = true;
        _receive(_seats.first, 3);
        _stage = DemoStage.clubsResolved;
        _events.add(
          'Resolved: both physical 5 of Clubs cards returned. +3 gross points repaid system debt 5 → 2; recorded score −5 → −2.',
        );
      case _Action.trade:
        final ours = _seats.first.hand.singleWhere(
          (card) => card.id == 'deck-1-hearts-5',
        );
        final theirs = _seats[2].hand.singleWhere(
          (card) => card.id == 'deck-1-hearts-6',
        );
        _seats.first.hand.remove(ours);
        _seats[2].hand.remove(theirs);
        _seats.first.hand.add(theirs);
        _seats[2].hand.add(ours);
        _tradeStatus = 'Completed';
        _tradeResolved = true;
        _stage = DemoStage.tradeResolved;
        _events.add(
          'Private receipt: your 5 of Hearts traded for Ivo’s 6 of Hearts. Both remain hand cards; no alliance formed.',
        );
      case _Action.seriesAddition:
        final card = _seats.first.hand.singleWhere(
          (card) => card.id == 'deck-1-clubs-7',
        );
        _seats.first.hand.remove(card);
        _seats.first.exposed.add(card);
        _densityAdditionComplete = true;
        _stage = DemoStage.densityResolved;
        _events.add(
          'Resolved: 7 of Clubs, copy 1, moved from your hand to your exposed Clubs series. No card was copied and no points were awarded.',
        );
    }
    _pendingAction = null;
    _selectedCardId = null;
    _responders = const [];
    _responseIndex = 0;
  }

  void offerTrade() {
    if (!canTrade) return;
    _offerRevision++;
    _offerId = '$_gameId-turn-13-seat-1-offer-$_offerRevision';
    _tradeStatus = 'Offered';
    _events.add(
      'Private offer to Ivo: your 5 of Hearts for his 6 of Hearts. No cards reserved; no alliance.',
    );
    notifyListeners();
  }

  void acceptTrade() {
    if (!_idle || !tradeOffered) return;
    _tradeStatus = 'Accepted · responses pending';
    _pendingAction = _Action.trade;
    _responders = ['Nia', _name, 'Mara'];
    _responseIndex = 0;
    _stage = DemoStage.tradeResponse;
    _selectedCardId = null;
    _events.add(
      'Ivo accepted offer revision $_offerRevision (local simulation). Transfer awaits responses.',
    );
    notifyListeners();
  }

  void declineTrade() {
    if (!_idle || !tradeOffered) return;
    _tradeStatus = 'Declined';
    _events.add(
      'Ivo explicitly declined the proposal (local simulation). No cards moved.',
    );
    notifyListeners();
  }

  void withdrawTrade() {
    if (!_idle || !tradeOffered) return;
    _tradeStatus = 'Withdrawn';
    _events.add('$_name withdrew the unanswered proposal. No cards moved.');
    notifyListeners();
  }

  void endTurn() {
    if (!canEndTurn) return;
    if (tradeOffered) _tradeStatus = 'Expired · originating turn ended';
    _selectedCardId = null;
    _stage = DemoStage.boardClosed;
    for (final seat in _seats) {
      // This small fixture uses only integer awards and one system obligation.
      // No player debts, Code replacement or rational settlement engine is here.
      final holdings = [...seat.exposed, ...seat.hand];
      var award = 0;
      for (final card in holdings) {
        final number = int.tryParse(card.rank);
        if (card.suit == CardSuit.diamonds && number != null) {
          award += number + 5;
        }
        if (card.suit == CardSuit.spades &&
            ['J', 'Q', 'K'].contains(card.rank)) {
          award += 10;
        }
      }
      _receive(seat, award);
    }
    _events.add(
      'Round 13 closed the board. No functioning round-income or Code formation remains.',
    );
    _events.add(
      'Ordinary awards: $_name 25 (3 of Diamonds = 8, 2 of Diamonds = 7, Queen of Spades = 10); Mara 13; Ivo 11; Nia 12. No +50 declaration bonus.',
    );
    _events.add(
      'Gross awards repay the oldest system debt first. Repayment is not charged to recorded score again.',
    );
    notifyListeners();
  }

  void finishSettlement() {
    if (!_connected || _stage != DemoStage.boardClosed) return;
    for (final seat in _seats) {
      _events.add(
        '${seat.name} explicitly finished settlement${seat.number == 1 ? '' : ' (local simulation)'}.',
      );
      seat.available = 0;
    }
    _stage = DemoStage.finalized;
    _events.add(
      'Game 1 financially finalized. Available balances expired; recorded scores are unchanged.',
    );
    notifyListeners();
  }

  void _receive(_DemoSeat seat, int amount) {
    seat.score += amount;
    final repaid = amount < seat.debt ? amount : seat.debt;
    seat.debt -= repaid;
    seat.available += amount - repaid;
  }

  void setConnection(ConnectionStateView connection) {
    _connection = connection;
    notifyListeners();
  }

  void reset() {
    _restore();
    notifyListeners();
  }

  void loadScenario(String id) {
    if (!scenarioIds.contains(id)) {
      throw ArgumentError.value(id, 'id', 'Unknown review scenario.');
    }
    _restore();
    switch (id) {
      case 'pending':
      case 'disconnected':
        _selectedCardId = 'deck-1-clubs-7';
        _declareAttack();
        if (id == 'disconnected')
          _connection = ConnectionStateView.disconnected;
      case 'trade':
        offerTrade();
      case 'results':
        _selectedCardId = 'deck-1-clubs-7';
        _declareAttack();
        _completeScenarioResponses();
        _selectedCardId = 'deck-1-clubs-5';
        _declareAttack();
        _completeScenarioResponses();
        endTurn();
      case 'error':
        _connection = ConnectionStateView.error;
      case 'loading':
        _connection = ConnectionStateView.loading;
      case 'empty':
        _stage = DemoStage.empty;
        _events = ['Empty adapter state for presentation review.'];
      case 'hand-10':
        _restoreDensity(10);
      case 'hand-15':
        _restoreDensity(15);
      case 'combos':
        _restoreCombos();
      case 'ready':
        break;
    }
    notifyListeners();
  }

  void _completeScenarioResponses() {
    while (_pendingAction != null) {
      _pass();
    }
  }

  void _restore() {
    _gameId = 'cgms-demo-game-${++_instance}';
    _stage = DemoStage.ready;
    _connection = ConnectionStateView.connected;
    _pendingAction = null;
    _responders = const [];
    _responseIndex = 0;
    _selectedCardId = null;
    _heartsCleared = false;
    _clubsCleared = false;
    _tradeStatus = 'No offer';
    _offerRevision = 0;
    _offerId = null;
    _tradeResolved = false;
    _densityReview = false;
    _densityAdditionComplete = false;
    _handCombos = const [];
    _seats = [
      _DemoSeat(
        id: 'self',
        name: _name,
        number: 1,
        exposed: [
          _card('3', CardSuit.diamonds),
          _card('2', CardSuit.diamonds),
          _card('7', CardSuit.clubs),
          _card('5', CardSuit.clubs),
        ],
        hand: [
          _card('Q', CardSuit.spades),
          _card('5', CardSuit.hearts),
          _card('J', CardSuit.hearts),
          _card('6', CardSuit.clubs),
        ],
        score: -5,
        debt: 5,
      ),
      _DemoSeat(
        id: 'mara',
        name: 'Mara',
        number: 2,
        exposed: [
          _card('8', CardSuit.diamonds),
          _card('4', CardSuit.hearts),
          _card('3', CardSuit.hearts),
          _card('5', CardSuit.clubs, copy: 2),
        ],
        hand: [
          _card('2', CardSuit.hearts),
          _card('8', CardSuit.hearts),
          _card('2', CardSuit.clubs),
          _card('4', CardSuit.clubs),
          _card('2', CardSuit.spades),
        ],
        aceCount: 1,
      ),
      _DemoSeat(
        id: 'ivo',
        name: 'Ivo',
        number: 3,
        exposed: [_card('6', CardSuit.diamonds)],
        hand: [
          _card('6', CardSuit.hearts),
          _card('3', CardSuit.spades),
          _card('6', CardSuit.spades),
          _card('4', CardSuit.clubs, copy: 2),
        ],
      ),
      _DemoSeat(
        id: 'nia',
        name: 'Nia',
        number: 4,
        exposed: [_card('7', CardSuit.diamonds)],
        hand: [
          _card('8', CardSuit.hearts, copy: 2),
          _card('4', CardSuit.spades),
          _card('10', CardSuit.spades),
          _card('9', CardSuit.clubs),
        ],
        aceCount: 1,
      ),
    ];
    _events = [
      'Prepared late-game fixture, not a fresh deal: round 13. Initiative Mara (8), Nia (7), Ivo (6), $_name (5); their turns are complete.',
      'Earlier this game, Code charged $_name 5 with no available points: score −5, system debt 5. Its formation broke; retained Diamond bonus and purchase price are both 5.',
      'Mara previously opened a second series; losing it never restores initial protection. Ivo and Nia remain initially protected.',
    ];
  }

  void _restoreDensity(int handSize) {
    _densityReview = true;
    _densityAdditionComplete = false;
    _stage = DemoStage.densityReady;
    _tradeStatus = 'Unavailable in density review';
    final ownHand = [
      _card('7', CardSuit.clubs),
      _card('Q', CardSuit.spades),
      _card('J', CardSuit.hearts),
      _card('K', CardSuit.clubs),
      ..._denseSeries(CardSuit.hearts, [5, 6, 7, 8, 9, 10]),
      ..._denseSeries(CardSuit.clubs, [5, 6, 8, 9, 10]),
    ];
    _seats = [
      _DemoSeat(
        id: 'self',
        name: _name,
        number: 1,
        exposed: [
          for (final suit in CardSuit.values) ..._denseSeries(suit, [2, 3, 4]),
        ],
        hand: ownHand.take(handSize).toList(),
      ),
      _DemoSeat(
        id: 'mara',
        name: 'Mara',
        number: 2,
        exposed: [
          ..._denseSeries(CardSuit.diamonds, [5, 6, 7]),
          for (final suit in [CardSuit.hearts, CardSuit.clubs, CardSuit.spades])
            ..._denseSeries(suit, [2, 3, 4], copy: 2),
        ],
        hand: _denseSeries(CardSuit.spades, [5, 6, 7], copy: 2),
      ),
      _DemoSeat(
        id: 'ivo',
        name: 'Ivo',
        number: 3,
        exposed: [
          ..._denseSeries(CardSuit.diamonds, [8, 9, 10]),
          ..._denseSeries(CardSuit.hearts, [5, 6, 7], copy: 2),
          ..._denseSeries(CardSuit.clubs, [5, 6, 7], copy: 2),
          ..._denseSeries(CardSuit.spades, [5, 6, 7]),
        ],
        hand: _denseSeries(CardSuit.spades, [8, 9, 10], copy: 2),
      ),
      _DemoSeat(
        id: 'nia',
        name: 'Nia',
        number: 4,
        exposed: [
          ..._denseSeries(CardSuit.diamonds, [2, 3, 4], copy: 2),
          ..._denseSeries(CardSuit.hearts, [8, 9, 10], copy: 2),
          ..._denseSeries(CardSuit.clubs, [8, 9, 10], copy: 2),
          ..._denseSeries(CardSuit.spades, [8, 9, 10]),
        ],
        hand: [
          _card('J', CardSuit.spades),
          _card('K', CardSuit.spades),
          _card('Q', CardSuit.hearts),
        ],
      ),
    ];
    _validateFixture();
    _events = [
      'Prepared round 8 review: each player has 12 public number cards across four open series. Everyone has permanently left initial attack protection.',
      'Every number card in your $handSize-card hand was acquired after its corresponding series first opened. Its 7 of Clubs may now be added to the existing Clubs series on your turn.',
      'Recorded scores, available points and debts are all zero in this fixture. No card awards are due yet; trade, combat and end scoring are outside this density review.',
    ];
  }

  void _restoreCombos() {
    _restoreDensity(15);
    final hand = _seats.first.hand;
    hand[hand.indexWhere((card) => card.id == 'deck-1-hearts-10')] = _card(
      'J',
      CardSuit.hearts,
      copy: 2,
    );
    // The natural queen pair is public in this prepared snapshot. Replace
    // Nia's former hidden queen with an unused card; never duplicate custody.
    final niaHand = _seats[3].hand;
    niaHand[niaHand.indexWhere((card) => card.id == 'deck-1-hearts-Q')] = _card(
      'Q',
      CardSuit.diamonds,
    );
    final greatPeopleCards = [
      _card('Q', CardSuit.hearts),
      _card('Q', CardSuit.hearts, copy: 2),
    ];
    _seats[1].exposed.addAll(greatPeopleCards);
    _seats[1].combos.add(
      ComboView(
        id: 'mara-great-people',
        title: 'Great People',
        stateLabel: 'Active · protecting Diamonds',
        description:
            'Mara’s natural twin Queens of Hearts protect her Diamonds series from club attacks, including Baron. Her exposed number Hearts and four current series support this passive formation. Inspection only; combo reconfiguration is not simulated.',
        cards: List<VisibleCard>.unmodifiable(greatPeopleCards),
      ),
    );
    final fateCards = [
      for (final suit in CardSuit.values) _card('K', suit, copy: 2),
    ];
    _seats[2].exposed.addAll(fateCards);
    _seats[2].combos.add(
      ComboView(
        id: 'ivo-fate',
        title: 'Fate',
        stateLabel: 'Opened · waiting for Ivo’s turn',
        description:
            'Four different Kings form Fate. Ivo has exposed number Clubs and four current series, but may activate it only on his own turn against an eligible opponent. It resets that opponent’s personal cards while preserving scores, debts and active persistent Aces; one Fate King returns before the reset shuffle. Inspection only; Fate activation is not simulated.',
        cards: List<VisibleCard>.unmodifiable(fateCards),
      ),
    );
    _handCombos = List<ComboView>.unmodifiable([
      ComboView(
        id: 'self-potential-kidnapper',
        title: 'Kidnapper',
        stateLabel: 'Potential · still in hand',
        description:
            'Your twin Jacks of Hearts match the Kidnapper recipe but are not an opened combo. Opening normally uses your one opening action on your turn and requires two current series and exposed number Clubs. Those suits are present here. Once per own turn, it steals a random eligible concealed Ace or royal from an eligible opponent; open-card choice is allowed only when they have no eligible closed cards. Inspection only; opening and activating Kidnapper are not simulated.',
        cards: List<VisibleCard>.unmodifiable([
          hand.singleWhere((card) => card.id == 'deck-1-hearts-J'),
          hand.singleWhere((card) => card.id == 'deck-2-hearts-J'),
        ]),
      ),
    ]);
    _validateFixture();
    _events.add(
      'Public combo review: Mara’s Great People protects her Diamonds; Ivo’s Fate waits for his turn. Both formations were opened on earlier turns. Combo abilities are inspect-only in this snapshot; displaying them awards no points and creates no debt.',
    );
  }

  void _validateFixture() {
    // Include private cards when checking custody without projecting them to
    // other viewers. Combo metadata references these holdings; it adds none.
    final controlledCards = _seats
        .expand((seat) => [...seat.exposed, ...seat.hand])
        .toList();
    assert(controlledCards.length <= 104);
    assert(
      controlledCards.map((card) => card.id).toSet().length ==
          controlledCards.length,
    );
    assert(controlledCards.every((card) => card.copy == 1 || card.copy == 2));
  }

  static List<VisibleCard> _denseSeries(
    CardSuit suit,
    List<int> ranks, {
    int copy = 1,
  }) => ranks.map((rank) => _card('$rank', suit, copy: copy)).toList();

  static VisibleCard _card(String rank, CardSuit suit, {int copy = 1}) =>
      VisibleCard(
        id: 'deck-$copy-${suit.name}-$rank',
        rank: rank,
        suit: suit,
        copy: copy,
      );
}

class _DemoSeat {
  _DemoSeat({
    required this.id,
    required this.name,
    required this.number,
    required this.exposed,
    required this.hand,
    this.aceCount = 0,
    this.score = 0,
    this.debt = 0,
  });

  final String id;
  final String name;
  final int number;
  final List<VisibleCard> exposed;
  final List<VisibleCard> hand;
  final List<ComboView> combos = [];
  final int aceCount;
  int score;
  int available = 0;
  int debt;
}
