import 'package:cgms_ui/cgms_ui.dart';
import 'online_client.dart';

ExactPoints exactAmount(dynamic value) {
  if (value is! Map) return ExactPoints.integer(0);
  return ExactPoints.fromStrings(
    value['numerator'] as String,
    value['denominator'] as String,
  );
}

VisibleCard projectCard(
  Map<dynamic, dynamic> placed, {
  String extraStatus = '',
  CardStyle style = CardStyle.firstLight,
}) {
  final card = placed['card'] as Map;
  final rank = card['rank'] as int;
  return VisibleCard(
    id: card['id'] as String,
    rank: switch (rank) {
      1 => 'A',
      11 => 'J',
      12 => 'Q',
      13 => 'K',
      _ => '$rank',
    },
    suit: CardSuit.values.byName(card['suit'] as String),
    seriesSuit: placed['zone'] == 'series'
        ? CardSuit.values.byName(card['suit'] as String)
        : placed['zone'] == 'attachment'
        ? CardSuit.values
              .where((suit) => suit.name == placed['allocation'])
              .firstOrNull
        : null,
    copy: card['deck'] as int,
    style: style,
    status: [
      (placed['available_from_round'] as int? ?? 0) > 0
          ? 'Available from round ${placed['available_from_round']}'
          : placed['zone'] as String,
      if (extraStatus.isNotEmpty) extraStatus,
    ].join(' · '),
  );
}

/// Only already-authorized cards enter presentation. Unknown opponent finances
/// remain undisplayed, not represented as factual zero balances.
TableViewData projectTable(
  AuthoritativeSnapshot snapshot, {
  bool connected = true,
  String usedClubStatus = '',
  String botLabel = 'Bot',
  String departureLabel = 'Round complete · Choose whether to stay',
  String roundWaitingLabel = 'Round complete · Waiting for players',
  String serverProgressLabel = 'Preparing the next move…',
}) {
  final board = snapshot.board;
  final placed = (board['cards'] as List? ?? []).cast<Map>();
  final players = (board['players'] as List? ?? []).cast<Map>();
  final botSeats = (snapshot.online['bot_seats'] as List? ?? [])
      .whereType<int>()
      .toSet();
  String seatName(int seat) =>
      '${botSeats.contains(seat) ? botLabel : 'Seat'} $seat';
  // Public match assignments follow the current controller, never the card's
  // physical copy, previous owner or the viewing seat. Missing legacy mappings
  // and unowned cards retain First Light; concealed cards never enter cardView.
  final styles = {
    for (final player in players)
      player['seat'] as int: CardStyle.fromWire(player['card_style']),
  };
  final own = snapshot.seat;
  final required = board['required_actor'] as int? ?? 0;
  final zero = ExactPoints.integer(0);
  VisibleCard cardView(Map card) => projectCard(
    card,
    style: styles[card['controller']] ?? CardStyle.firstLight,
    extraStatus:
        card['controller'] == own &&
            card['zone'] == 'series' &&
            card['card']['suit'] == 'clubs' &&
            (card['card']['rank'] as int) >= 2 &&
            (card['card']['rank'] as int) <= 10 &&
            (board['turn'] as int? ?? 0) > 0 &&
            card['used_turn'] == board['turn']
        ? usedClubStatus
        : '',
  );
  return TableViewData(
    gameId: snapshot.gameId,
    round: board['round'] as int,
    totalRounds: 13,
    turnLabel: snapshot.online['server_pending'] == true
        ? serverProgressLabel
        : snapshot.phase != 'playing'
        ? 'Board closed · ${snapshot.phase}'
        : snapshot.online['round_closed'] == true
        ? snapshot.online['departure_pending'] == true
              ? departureLabel
              : roundWaitingLabel
        : required > 0
        ? botSeats.contains(required)
              ? 'Waiting for ${seatName(required)}'
              : 'Waiting for seat $required'
        : '${seatName(board['active'] as int)}’s turn',
    phaseLabel: snapshot.phase,
    objective: '',
    actionLabel: '',
    actionExplanation: '',
    connection: connected
        ? ConnectionStateView.connected
        : ConnectionStateView.disconnected,
    seats: [
      for (final raw in players)
        SeatView(
          id: '${raw['seat']}',
          name: seatName(raw['seat'] as int),
          number: raw['seat'] as int,
          isSelf: raw['seat'] == own,
          financesKnown: raw['seat'] == own,
          exposed: [
            for (final card in placed)
              if (card['controller'] == raw['seat'] &&
                  !['hand', 'concealed-ace', 'draw'].contains(card['zone']))
                cardView(card),
          ],
          concealed: ConcealedCards(
            handCount: raw['hand_count'] as int,
            aceCount: raw['ace_count'] as int,
          ),
          finances: FinancialView(
            matchScore: raw['seat'] == own
                ? exactAmount(snapshot.online['own_match_score'])
                : zero,
            gameScore: raw['seat'] == own
                ? exactAmount(snapshot.projection['score'])
                : zero,
            available: raw['seat'] == own
                ? exactAmount(snapshot.projection['cash'])
                : zero,
            debt: raw['seat'] == own
                ? (snapshot.projection['debts'] as List? ?? [])
                      .where((d) => d['debtor'] == own - 1)
                      .fold<ExactPoints>(
                        zero,
                        (sum, d) =>
                            sum + exactAmount(d['remaining'] ?? d['amount']),
                      )
                : zero,
          ),
          combos: [
            for (final formation
                in (board['formations'] as List? ?? []).cast<Map>())
              if (formation['controller'] == raw['seat'])
                ComboView(
                  id: formation['id'] as String,
                  title: (formation['spec'] as Map)['kind'] as String,
                  stateLabel: 'Exposed formation',
                  description: '',
                  cards: [
                    for (final id
                        in (formation['spec'] as Map)['cards'] as List)
                      for (final card in placed)
                        if ((card['card'] as Map)['id'] == id) cardView(card),
                  ],
                ),
          ],
          status: [
            if (raw['confined'] == true) 'Confined',
            if (raw['departed'] == true) 'Departed',
            if (raw['seat'] == own) 'You',
          ].join(' · '),
        ),
    ],
    hand: [
      for (final card in placed)
        if (card['controller'] == own &&
            ['hand', 'concealed-ace'].contains(card['zone']))
          cardView(card),
    ],
    events: const [],
  );
}

/// Presentation admission hint from the viewer's own authorized holdings.
/// The authority performs the final current-state check, including stale versions.
bool canDeclareCoup(AuthoritativeSnapshot snapshot) {
  if (snapshot.phase != 'playing') return false;
  final seat = snapshot.seat, board = snapshot.board;
  final players = (board['players'] as List? ?? []).cast<Map>();
  final self = players.where((p) => p['seat'] == seat).firstOrNull;
  if (self == null ||
      self['departed'] == true ||
      !(self['history'] as List? ?? []).contains('clubs'))
    return false;
  final ally = (snapshot.online['alliances'] as Map?)?['$seat'];
  if (ally is int && ally != 0) return false;
  final cards = (board['cards'] as List? ?? []).cast<Map>().where(
    (p) =>
        p['controller'] == seat &&
        (p['available_from_round'] as int? ?? 0) <= (board['round'] as int) &&
        (p['card'] as Map)['rank'] == 12,
  );
  return ['hearts', 'diamonds'].every(
    (suit) =>
        cards.where((p) => (p['card'] as Map)['suit'] == suit).length == 2,
  );
}
