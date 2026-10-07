import 'package:flutter/foundation.dart';

enum CardSuit { hearts, diamonds, clubs, spades }

/// Presentation destinations only; the host resolves and submits authorized moves.
enum TableDropZone { seat, hand, drawPile, card, formation, series }

/// Presentation role of an exact physical card in a host-owned draft.
enum TableCardRole { source, target, payment, modifier }

@immutable
class TableDropTarget {
  const TableDropTarget({
    required this.zone,
    this.seat,
    this.cardId,
    this.formationId,
    this.suit,
  });

  final TableDropZone zone;
  final int? seat;
  final String? cardId;
  final String? formationId;
  final String? suit;
}

/// A snapshot of authorized physical identities, never hidden card contents.
@immutable
class TableCardDrag {
  TableCardDrag({
    required this.gameId,
    required List<String> cardIds,
    this.intentIdentity,
  }) : cardIds = List.unmodifiable(cardIds);

  final String gameId;
  final List<String> cardIds;

  /// Host-supplied game/version/window identity captured when dragging begins.
  final Object? intentIdentity;
}

/// Public presentation assignment supplied by the authority for a whole match.
/// It never changes the physical identity or rules of a card.
enum CardStyle {
  firstLight('first-light'),
  rainGlaze('rain-glaze'),
  emberGlaze('ember-glaze'),
  drypoint('drypoint');

  const CardStyle(this.slug);

  final String slug;

  /// Unknown or pre-assignment projections keep the accepted default artwork.
  /// No server-provided string is ever interpolated into an asset URL.
  static CardStyle fromWire(Object? value) => values.firstWhere(
    (style) => style.slug == value,
    orElse: () => firstLight,
  );
}

/// A face-up physical card the local viewer is authorized to see.
///
/// Adapters must supply the stable physical [id], even when rank/suit match.
@immutable
class VisibleCard {
  const VisibleCard({
    required this.id,
    required this.rank,
    required this.suit,
    required this.copy,
    this.status = '',
    this.seriesSuit,
    this.style = CardStyle.firstLight,
  }) : assert(copy > 0);

  final String id;
  final String rank;
  final CardSuit suit;
  final int copy;
  final String status;

  /// Authorized supporting series; never changes this card's printed suit.
  final CardSuit? seriesSuit;

  /// The current controller's public match assignment, supplied by the adapter.
  final CardStyle style;

  String get artworkAsset {
    // Preserve the original First Light paths for existing package consumers.
    final directory = style == CardStyle.firstLight ? '' : '${style.slug}/';
    return 'assets/art/$directory${suit.name}-$rank.png';
  }

  /// Native painting proportions from docs/design/cgmsart-decks/manifest.json.
  /// All four cosmetic styles retain these 52 source dimensions.
  double get artworkAspectRatio {
    const ranks = [
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
    ];
    final index = ranks.indexOf(rank);
    if (index < 0) return 5 / 7;
    final (width, height) = _artworkSizes[suit]![index];
    return width / height;
  }

  /// Zones already identified by their surrounding collection add no card state.
  /// Keep the raw value available to accessible identity and host adapters.
  String get displayStatus => switch (status.trim().toLowerCase()) {
    'hand' || 'series' || 'concealed-ace' => '',
    _ => status,
  };

  static const _artworkSizes = <CardSuit, List<(int, int)>>{
    CardSuit.hearts: [
      (284, 349),
      (283, 349),
      (284, 349),
      (284, 349),
      (283, 349),
      (284, 338),
      (283, 338),
      (284, 338),
      (284, 338),
      (283, 338),
      (284, 381),
      (283, 381),
      (284, 381),
    ],
    CardSuit.diamonds: [
      (282, 354),
      (286, 354),
      (283, 354),
      (285, 354),
      (282, 354),
      (282, 357),
      (286, 357),
      (283, 357),
      (285, 357),
      (282, 357),
      (282, 357),
      (286, 357),
      (283, 357),
    ],
    CardSuit.clubs: [
      (281, 355),
      (287, 355),
      (284, 355),
      (285, 355),
      (281, 355),
      (281, 359),
      (287, 359),
      (284, 359),
      (285, 359),
      (281, 359),
      (281, 354),
      (287, 354),
      (284, 354),
    ],
    CardSuit.spades: [
      (283, 354),
      (285, 354),
      (283, 354),
      (284, 354),
      (283, 354),
      (283, 354),
      (285, 354),
      (283, 354),
      (284, 354),
      (283, 354),
      (283, 360),
      (285, 360),
      (283, 360),
    ],
  };

  String get label =>
      '$rank of ${suit.name[0].toUpperCase()}${suit.name.substring(1)}, copy $copy';
}

/// Public counts only. No card identities can enter concealed rendering.
@immutable
class ConcealedCards {
  const ConcealedCards({required this.handCount, required this.aceCount})
    : assert(handCount >= 0),
      assert(aceCount >= 0);

  final int handCount;
  final int aceCount;
}

/// Reduced rational points, parsed directly from decimal integer strings.
///
/// This presentation value never passes through `double`, including on web.
/// It does not calculate awards, repayments, or any other game rule.
@immutable
class ExactPoints implements Comparable<ExactPoints> {
  factory ExactPoints.fromStrings(
    String numerator, [
    String denominator = '1',
  ]) => ExactPoints._normalize(
    BigInt.parse(numerator),
    BigInt.parse(denominator),
  );

  factory ExactPoints.integer(int value) =>
      ExactPoints._normalize(BigInt.from(value), BigInt.one);

  factory ExactPoints._normalize(BigInt numerator, BigInt denominator) {
    if (denominator == BigInt.zero) {
      throw ArgumentError.value(denominator, 'denominator', 'Must not be zero');
    }
    final divisor = numerator.gcd(denominator);
    final sign = denominator.isNegative ? -BigInt.one : BigInt.one;
    return ExactPoints._(
      numerator ~/ divisor * sign,
      denominator ~/ divisor * sign,
    );
  }

  const ExactPoints._(this.numerator, this.denominator);

  final BigInt numerator;
  final BigInt denominator;

  bool get isNegative => numerator.isNegative;

  String format() =>
      denominator == BigInt.one ? '$numerator' : '$numerator/$denominator';

  ExactPoints operator +(ExactPoints other) => ExactPoints._normalize(
    numerator * other.denominator + other.numerator * denominator,
    denominator * other.denominator,
  );

  ExactPoints operator -(ExactPoints other) => ExactPoints._normalize(
    numerator * other.denominator - other.numerator * denominator,
    denominator * other.denominator,
  );

  ExactPoints operator -() => ExactPoints._(-numerator, denominator);

  ExactPoints operator *(ExactPoints other) => ExactPoints._normalize(
    numerator * other.numerator,
    denominator * other.denominator,
  );

  ExactPoints operator /(ExactPoints other) => ExactPoints._normalize(
    numerator * other.denominator,
    denominator * other.numerator,
  );

  @override
  int compareTo(ExactPoints other) =>
      (numerator * other.denominator).compareTo(other.numerator * denominator);

  @override
  bool operator ==(Object other) =>
      other is ExactPoints &&
      numerator == other.numerator &&
      denominator == other.denominator;

  @override
  int get hashCode => Object.hash(numerator, denominator);

  @override
  String toString() => format();
}

@immutable
class FinancialView {
  const FinancialView({
    required this.matchScore,
    required this.gameScore,
    required this.available,
    required this.debt,
  });

  final ExactPoints matchScore;
  final ExactPoints gameScore;
  final ExactPoints available;
  final ExactPoints debt;
}

@immutable
class ComboView {
  const ComboView({
    required this.id,
    required this.title,
    required this.stateLabel,
    required this.description,
    required this.cards,
  });

  final String id;
  final String title;
  final String stateLabel;
  final String description;

  /// Authorized constituent cards, referencing the same physical IDs as the
  /// owner's exposed cards or the viewer's hand. State is supplied by adapters.
  final List<VisibleCard> cards;
}

@immutable
class SeatView {
  const SeatView({
    required this.id,
    required this.name,
    required this.number,
    required this.exposed,
    required this.concealed,
    required this.finances,
    required this.status,
    this.isSelf = false,
    this.financesKnown = true,
    this.combos = const [],
  });

  final String id;
  final String name;
  final int number;
  final List<VisibleCard> exposed;
  final ConcealedCards concealed;
  final FinancialView finances;
  final String status;
  final bool isSelf;
  final bool financesKnown;
  final List<ComboView> combos;
}

enum ConnectionStateView { connected, loading, disconnected, error }

@immutable
class PendingView {
  const PendingView({
    required this.title,
    required this.detail,
    required this.responders,
    required this.responseIndex,
  });

  final String title;
  final String detail;
  final List<String> responders;
  final int responseIndex;
}

/// A viewer-filtered projection. Legal actions and outcomes come from adapters.
@immutable
class TableViewData {
  const TableViewData({
    required this.gameId,
    required this.round,
    required this.totalRounds,
    required this.turnLabel,
    required this.seats,
    required this.hand,
    required this.events,
    required this.objective,
    required this.actionLabel,
    required this.actionExplanation,
    required this.phaseLabel,
    this.selectedCardId,
    this.actionEnabled = false,
    this.connection = ConnectionStateView.connected,
    this.pending,
    this.handCombos = const [],
  });

  final String gameId;
  final int round;
  final int totalRounds;
  final String turnLabel;
  final List<SeatView> seats;
  final List<VisibleCard> hand;
  final List<String> events;
  final String objective;
  final String actionLabel;
  final String actionExplanation;
  final String phaseLabel;
  final String? selectedCardId;
  final bool actionEnabled;
  final ConnectionStateView connection;
  final PendingView? pending;
  final List<ComboView> handCombos;
}
