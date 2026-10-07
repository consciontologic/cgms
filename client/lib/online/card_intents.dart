import 'package:cgms_ui/cgms_ui.dart';

/// A prepared draft, never an authorization or a submitted command.
class CardPlayIntent {
  const CardPlayIntent({
    required this.type,
    required this.seed,
    required this.explanation,
    this.needsSubstitution = false,
    this.needsProtection = false,
    this.sourceFormationId,
  });
  final String type;
  final Map<String, dynamic> seed;
  final String explanation;
  final bool needsSubstitution;
  final bool needsProtection;

  /// Authorized public source identity used only by presentation, never payload.
  final String? sourceFormationId;
}

List<CardPlayIntent> suggestCardPlays({
  required Map<String, dynamic> board,
  required Map<String, dynamic> online,
  required Set<String> cardIds,
  TableDropTarget? target,
}) {
  final view = _View(board, online);
  if (cardIds.isEmpty) return const [];
  if (cardIds.any((id) => !view.cards.containsKey(id))) return const [];
  final cards = [for (final id in cardIds) view.cards[id]!];
  // Coup has its own admission boundary, including confinement and pending
  // effects. These are only local presentation hints; the server checks the
  // current game, alliance and physical Queens without blanket version rejection.
  final immediate = <CardPlayIntent>[
    if (target == null &&
        view.coupQueens.isNotEmpty &&
        cards.every((card) => view.coupQueens.contains(card.id)))
      const CardPlayIntent(
        type: 'coup',
        seed: {},
        explanation:
            'Declare Coup with both physical Hearts Queens and both '
            'Diamond Queens. The server checks current eligibility immediately.',
      ),
    // Deliberate disclosure is not an Ace ability or a response. In particular,
    // confinement, availability markers and an already used Ace quota do not
    // prevent wasting a different unused Ace. The authority alone knows whether
    // a concealed Ace is already committed to a pending action.
    if (target == null &&
        cards.length == 1 &&
        cards.single.owner == view.seat &&
        cards.single.ace &&
        view.disclosureTiming)
      CardPlayIntent(
        type: 'expose-unused-ace',
        seed: _freeze({'ace': cards.single.id}),
        explanation:
            'Deliberately reveal and discard this unused Ace without using its '
            'power. Confirm only if you intend to waste this card; inspection is private. '
            'The server rejects an Ace already committed to a pending action.',
      ),
  ];
  if (!view.live) return List.unmodifiable(immediate);
  final foreign = cards.where((card) => card.owner != view.seat).toList();
  final currentWindow = board['window_id'];
  if (target == null &&
      foreign.isNotEmpty &&
      (currentWindow is! String || currentWindow.isEmpty)) {
    final ownIds = cards
        .where((card) => card.owner == view.seat)
        .map((card) => card.id)
        .toSet();
    if (ownIds.isEmpty || foreign.any((card) => !view.visible(card)))
      return const [];
    final first = foreign.first;
    final seriesTargets = foreign.every(
      (card) =>
          card.number &&
          card.zone == 'series' &&
          card.owner == first.owner &&
          card.suit == first.suit,
    );
    if (!seriesTargets && !(foreign.length == 1 && first.royal))
      return const [];
    // Neutral taps may include targets. Reuse the same source checks as a drop
    // and retain physical targets; Baron still commits the whole public series.
    final targeted = suggestCardPlays(
      board: board,
      online: online,
      cardIds: ownIds,
      target: TableDropTarget(
        zone: TableDropZone.card,
        seat: first.owner,
        cardId: first.id,
      ),
    );
    return List.unmodifiable([
      for (final intent in targeted)
        if (const {
          'attack',
          'infiltrator-attack',
          'baron-attack',
          'bomb-attack',
          'baron-bomb-attack',
          'dexter-assassination',
        }.contains(intent.type))
          CardPlayIntent(
            type: intent.type,
            seed: _freeze({
              ...intent.seed,
              if (seriesTargets && !intent.type.startsWith('baron-'))
                'targets': foreign.map((card) => card.id).toList(),
            }),
            explanation: intent.explanation,
            sourceFormationId: intent.sourceFormationId,
          ),
    ]);
  }
  final allowed = selectableCardIds(board, online);
  if (!allowed.containsAll(cardIds)) return List.unmodifiable(immediate);
  final destination = target?.cardId == null
      ? null
      : view.cards[target!.cardId];
  if (target?.cardId != null &&
      (destination == null || !view.visible(destination)))
    return const [];
  if (target?.zone == TableDropZone.series &&
      (!_suits.contains(target?.suit) ||
          destination != null &&
              (destination.zone != 'series' ||
                  destination.suit != target?.suit)))
    return const [];
  final targetFormations = view.formations.where(
    (f) => f['id'] == target?.formationId,
  );
  final targetFormation = targetFormations.isEmpty
      ? null
      : targetFormations.first;
  final targetSeat =
      destination?.owner ??
      targetFormation?['controller'] as int? ??
      target?.seat;
  if (targetSeat != null && !view.players.any((p) => p['seat'] == targetSeat)) {
    return const [];
  }
  if (targetSeat != null &&
      target?.seat != null &&
      targetSeat != target!.seat) {
    return const [];
  }
  if (target?.formationId != null &&
      !view.formations.any(
        (f) =>
            f['id'] == target!.formationId &&
            (targetSeat == null || f['controller'] == targetSeat),
      ))
    return const [];
  final seriesDestination = target?.zone == TableDropZone.series;
  final targetSeriesCards = [
    if (seriesDestination)
      for (final card in view.cards.values)
        if (card.owner == targetSeat &&
            card.zone == 'series' &&
            card.number &&
            card.suit == target?.suit)
          card,
  ];
  final result = [...immediate];
  void add(
    String type,
    Map<String, dynamic> seed,
    String explanation, {
    bool needsSubstitution = false,
    bool needsProtection = false,
    String? sourceFormationId,
  }) {
    result.add(
      CardPlayIntent(
        type: type,
        seed: _freeze(seed),
        explanation: explanation,
        needsSubstitution: needsSubstitution,
        needsProtection: needsProtection,
        sourceFormationId: sourceFormationId,
      ),
    );
  }

  final window = board['window_id'];
  if (window is String && window.isNotEmpty) {
    if (board['required_actor'] != view.seat) return List.unmodifiable(result);
    // A response has a frozen actor and affected target. Unrelated drop zones
    // must not be interpreted as a different use of a concealed Ace or payment.
    if (target?.zone == TableDropZone.drawPile ||
        target?.zone == TableDropZone.hand)
      return const [];
    final decision = board['decision_id'];
    final kidnapperOutcome =
        board['decision_kind'] == 'kidnapper-outcome' &&
        (decision == null || decision == '');
    if ((decision is String && decision.isNotEmpty) || kidnapperOutcome) {
      if (target != null &&
          !(target.zone == TableDropZone.seat && targetSeat == view.seat)) {
        return const [];
      }
      for (final choice in _choices(board)) {
        if (choice.length == cardIds.length && cardIds.containsAll(choice)) {
          add(kidnapperOutcome ? 'kidnapper-outcome' : 'decision', {
            'pending_action_id': window,
            if (!kidnapperOutcome) 'decision_id': decision,
            'selection': choice,
          }, 'Use this offered selection to resolve the current effect.');
          break;
        }
      }
      return result;
    }
    if ((board['decision_kind'] as String? ?? '').isNotEmpty) return result;
    final pending = _map(view.context['pending']);
    final responseTypes = _strings(pending['response_types']).toSet();
    final combat = _map(view.context['combat']);
    final incomingCombat = combat.isNotEmpty && combat['defender'] == view.seat;
    final numericalCombat =
        incomingCombat && combat['comparison'] != 'at_least';
    final targetCards = _strings(combat['target_cards']);
    final pendingActor = pending['actor'] ?? combat['actor'] ?? board['active'];
    final actorDestination =
        target == null ||
        targetSeat == pendingActor ||
        target.zone == TableDropZone.seat && targetSeat == view.seat;
    final affectedCards = {
      ...targetCards,
      ..._strings(pending['target_cards']),
    };
    final affectedDestination =
        target == null ||
        targetSeat == view.seat &&
            (target.zone == TableDropZone.seat ||
                target.zone == TableDropZone.card &&
                    affectedCards.contains(target.cardId) ||
                target.zone == TableDropZone.series &&
                    affectedCards.any(
                      (id) =>
                          view.cards[id]?.suit == target.suit &&
                          view.cards[id]?.owner == view.seat,
                    ));
    final heartsTarget =
        targetCards.isNotEmpty &&
        view.cards[targetCards.first]?.suit == 'hearts';
    bool permits(String type, bool legacy) =>
        pending.isNotEmpty ? responseTypes.contains(type) : legacy;
    if (cards.length == 1 && cards.single.ace && !view.aceUsed) {
      final ace = cards.single;
      final seed = {'pending_action_id': window, 'ace': ace.id};
      if (ace.suit == 'diamonds' &&
          actorDestination &&
          permits('confinement', true))
        add(
          'confinement',
          seed,
          'Use Confinement against the player taking this action.',
        );
      if (ace.suit == 'spades' &&
          actorDestination &&
          !view.inflated(pendingActor as int?) &&
          permits('inflation', incomingCombat))
        add(
          'inflation',
          seed,
          'Use Inflation against the attacker when this response permits it.',
        );
      if (ace.suit == 'hearts' &&
          affectedDestination &&
          !view.quota('compensation') &&
          permits('compensation-response', incomingCombat)) {
        add(
          'compensation-response',
          seed,
          'Activate Compensation before eligible losses.',
        );
      }
      if (ace.suit == 'clubs' &&
          affectedDestination &&
          permits('advancement-defense', numericalCombat && heartsTarget))
        add(
          'advancement-defense',
          seed,
          'Add 10 only to an incoming attack on your Hearts.',
        );
      if (affectedDestination && permits('numerical-defense', numericalCombat))
        add(
          'numerical-defense',
          seed,
          'Choose a value from 2 to 10 for the incoming number-series attack.',
        );
    }
    if (cards.length == 1 &&
        (actorDestination || affectedDestination) &&
        cards.single.rank == 11 &&
        cards.single.suit == 'spades' &&
        cards.single.zone == 'attachment' &&
        cards.single.allocation == 'spades' &&
        !view.quota('negotiator') &&
        permits('negotiate', incomingCombat)) {
      add(
        'negotiate',
        {'pending_action_id': window},
        'Use Negotiator, then choose from the available Club and payment options.',
      );
    }
    return result;
  }

  final ids = cards.map((c) => c.id).toList();
  final foreignTarget = targetSeat != null && targetSeat != view.seat;
  final ownTurn = board['active'] == view.seat;
  final onHand = target?.zone == TableDropZone.hand;
  final onDeck = target?.zone == TableDropZone.drawPile;
  final untargeted = target == null;
  final ownArea = !foreignTarget && !onHand && !onDeck;
  final selectedFormationIds = cards
      .where((c) => c.zone == 'formation')
      .map((c) => c.allocation)
      .toSet();
  final ownFormations = view.formations.where(
    (f) => f['controller'] == view.seat,
  );
  final selectedFormations = ownFormations.where(
    (f) => selectedFormationIds.contains(f['id']),
  );
  final reconfigureId =
      target?.formationId ??
      (selectedFormationIds.length == 1 && selectedFormations.length == 1
          ? selectedFormationIds.single
          : null);

  // A card or whole-formation gesture prepares a private proposal; it does not
  // reserve cards or create a transfer/alliance. The host asks for recipient and
  // terms, creates an offer id, then commits the exact revision explicitly.
  final selectedIntactFormation = selectedFormations.any((f) {
    final members = _strings(_map(f['spec'])['cards']);
    return members.length >= 2 &&
        members.length == cardIds.length &&
        cardIds.containsAll(members);
  });
  // Nested card/series/formation hit targets are still inside the recipient's
  // public area. An intact formation can never turn into an attack there.
  final recipientTarget =
      foreignTarget &&
      (target?.zone == TableDropZone.seat || selectedIntactFormation);
  final intactRecipientDrop = recipientTarget && selectedIntactFormation;
  if (recipientTarget && !view.eligibleParty(targetSeat)) return const [];
  if (untargeted || recipientTarget) {
    if (ownTurn &&
        !intactRecipientDrop &&
        cards.every(
          (card) =>
              const {'hand', 'concealed-ace'}.contains(card.zone) ||
              view.context['underground'] == true && !card.ace,
        ) &&
        (cards.every((card) => !card.ace) || !view.aceUsed)) {
      add(
        'offer',
        {
          'terms': {
            if (recipientTarget) 'to': targetSeat,
            'give': ids,
            'receive': <String>[],
            'loan': false,
          },
        },
        'Offer exactly these cards in a private trade. Choose the recipient '
            'and requested cards; only their explicit acceptance starts a transfer.',
      );
    }
    for (final formation in selectedFormations) {
      final members = _strings(_map(formation['spec'])['cards']);
      if (members.isEmpty ||
          recipientTarget && !view.allianceCompatible(targetSeat) ||
          cards.any((card) => !members.contains(card.id)) ||
          members.any(
            (id) =>
                view.cards[id] == null ||
                view.cards[id]!.owner != view.seat ||
                view.cards[id]!.zone != 'formation' ||
                view.cards[id]!.allocation != formation['id'] ||
                !view.available(view.cards[id]!),
          ))
        continue;
      if (!ownTurn &&
          (recipientTarget
              ? targetSeat != board['active']
              : board['active'] == view.seat))
        continue;
      add(
        'offer',
        {
          'terms': {
            if (recipientTarget) 'to': targetSeat,
            if (!ownTurn && untargeted) 'to': board['active'],
            'give': members,
            'receive': <String>[],
            'loan': true,
          },
        },
        'Offer this entire exposed formation as a loan. Review every member '
            'and the permanent two-player alliance before the recipient accepts.',
      );
    }
  }

  // Returning borrowed cards is an intent, never a transfer from a drop alone.
  if (!onDeck && !onHand) {
    for (final loan in _maps(board['loans'])) {
      if (loan['to'] != view.seat || loan['from'] is! int) continue;
      final borrowed = _strings(loan['cards']).toSet();
      if (borrowed.containsAll(ids) &&
          (targetSeat == null || targetSeat == loan['from'])) {
        add(
          'return-loan',
          {'selection': ids, 'target_seat': loan['from']},
          'Return these borrowed cards to their lender after responses resolve.',
        );
      }
    }
  }
  // Moving an intact exposed formation identifies custody/recipient, never
  // activation. The proposal remains private and nonblocking until accepted.
  if (intactRecipientDrop) {
    return List.unmodifiable(
      result.where(
        (intent) =>
            intent.type == 'return-loan' ||
            intent.type == 'offer' && intent.seed['terms']['loan'] == true,
      ),
    );
  }
  for (final f in selectedFormations) {
    final kind = _map(f['spec'])['kind'];
    if (kind == 'ponzi' && !onDeck && !onHand && !view.quota('ponzi')) {
      add(
        'ponzi',
        {if (foreignTarget) 'value': targetSeat},
        'Choose an opponent. If you have an ally, split the stolen amount equally with them.',
      );
    }
  }
  if (!ownTurn) return result;

  if ((untargeted || ownArea) &&
      cards.every((c) => c.number && c.zone == 'hand') &&
      cards.map((c) => c.suit).toSet().length == 1 &&
      target?.zone != TableDropZone.formation &&
      (target?.suit == null || target?.suit == cards.first.suit) &&
      (destination == null ||
          destination.zone == 'series' &&
              destination.suit == cards.first.suit)) {
    final suit = cards.first.suit;
    final ownPlayer = view.players.where((p) => p['seat'] == view.seat).first;
    final historical = _strings(ownPlayer['history']).contains(suit);
    final restrictedSameSuit = view.cards.values.any(
      (c) =>
          c.owner == view.seat &&
          c.zone == 'hand' &&
          c.number &&
          c.suit == suit &&
          !view.available(c),
    );
    if ((historical || !restrictedSameSuit) &&
        (historical || view.context['opening_used'] != true)) {
      add(
        'open-series',
        {
          'suit': suit,
          'selection': historical
              ? ids
              : [
                  for (final card in view.cards.values)
                    if (card.owner == view.seat &&
                        card.zone == 'hand' &&
                        card.number &&
                        card.suit == suit)
                      card.id,
                ],
        },
        historical
            ? 'Add these selected number cards to the historically opened suit. Restoring an empty series does not spend the opening allowance.'
            : 'First opening: open all number cards of this suit held in your hand, not only this selection. This spends the opening allowance.',
      );
    }
  }
  if ((untargeted || onDeck) &&
      cards.every(
        (c) =>
            c.number &&
            c.suit == 'spades' &&
            (c.zone == 'hand' || c.zone == 'series'),
      )) {
    add(
      'purchase',
      {'payment': ids},
      'Choose exactly how many cards to buy with these Spades, then review the price and payment.',
    );
  }
  if ((untargeted || ownArea) &&
      cards.length == 1 &&
      cards.single.royal &&
      const {'hand', 'unassigned'}.contains(cards.single.zone)) {
    final suit = destination?.zone == 'series' && destination!.number
        ? destination.suit
        : target?.zone == TableDropZone.series
        ? target?.suit
        : null;
    final supported = royalAttachmentSuits(board, ids.toSet());
    if (supported.isNotEmpty && (suit == null || supported.contains(suit))) {
      add('attach', {
        'selection': ids,
        if (suit != null) 'suit': suit,
      }, 'Choose the supported series for this royal and its non-combo role.');
    }
  }

  final selectedBindings = view.ownBindings
      .where((binding) => _strings(binding['kings']).any(cardIds.contains))
      .toList();
  final knownSubstitution =
      selectedBindings.length == 1 &&
          _strings(selectedBindings.single['kings']).length == 2 &&
          cardIds.containsAll(_strings(selectedBindings.single['kings']))
      ? selectedBindings.single
      : null;
  if ((untargeted || ownArea) &&
      (selectedBindings.isEmpty || knownSubstitution != null) &&
      cards.every(
        (c) =>
            c.royal &&
            (const {'hand', 'unassigned'}.contains(c.zone) ||
                c.zone == 'formation' && c.allocation == reconfigureId),
      ) &&
      (view.context['opening_used'] != true ||
          view.context['underground'] == true)) {
    final natural = _formationKinds(
      cards.map((c) => (c.rank, c.suit)).toList(),
    );
    final substituted = <String>{};
    for (var i = 0; i < cards.length; i++) {
      if (cards[i].rank != 13) continue;
      for (var j = i + 1; j < cards.length; j++) {
        if (cards[j].rank != 13) continue;
        final other = [
          for (var k = 0; k < cards.length; k++)
            if (k != i && k != j) (cards[k].rank, cards[k].suit),
        ];
        for (final rank in [11, 12, 13]) {
          for (final suit in _suits) {
            final kinds = _formationKinds([...other, (rank, suit)]);
            for (final kind in kinds) {
              if (kind == 'great-people' && rank == 12 && suit == 'hearts') {
                substituted.add('people');
              } else if (!const {
                'great-people',
                'ponzi',
                'coup',
              }.contains(kind)) {
                substituted.add(kind);
              }
            }
          }
        }
      }
    }
    for (final kind in {...natural, ...substituted}) {
      if (kind == 'coup') continue; // The host owns the immediate victory path.
      final usesSubstitution =
          knownSubstitution != null || !natural.contains(kind);
      // Owner-only fixed bindings survive return to hand and separation.
      // Preserve the exact pair/slot; never guess a new printed role.
      if (knownSubstitution != null) {
        if (const {'coup', 'ponzi', 'great-people'}.contains(kind)) continue;
        final pair = _strings(knownSubstitution['kings']);
        final slot = _map(knownSubstitution['slot']);
        final kinds = _formationKinds([
          for (final card in cards)
            if (!pair.contains(card.id)) (card.rank, card.suit),
          (slot['rank'] as int? ?? 0, slot['suit'] as String? ?? ''),
        ]);
        if (!(kind == 'people' && kinds.contains('great-people')) &&
            !kinds.contains(kind))
          continue;
      }
      add(
        'open-formation',
        {
          if (reconfigureId != null) 'formation_id': reconfigureId,
          'formation': {
            'kind': kind,
            'cards': ids,
            if (knownSubstitution != null) 'substitute': knownSubstitution,
          },
        },
        '${usesSubstitution ? 'Choose the Doppelganger king pair and fixed replacement slot. ' : ''}'
            '${const {'people', 'great-people'}.contains(kind) ? 'Choose one permitted protection target. ' : ''}'
            'Review this combo’s support, fixed roles and opening allowance before confirming.',
        needsSubstitution: usesSubstitution && knownSubstitution == null,
        needsProtection: const {'people', 'great-people'}.contains(kind),
      );
    }
    if (cards.length == 2 && cards.every((c) => c.rank == 13)) {
      add(
        'open-formation',
        {
          if (reconfigureId != null) 'formation_id': reconfigureId,
          'formation': {
            'cards': ids,
            if (knownSubstitution != null) 'substitute': knownSubstitution,
          },
        },
        'Choose the intended combo, its other members and the fixed Doppelganger role; two kings alone do not open a combo.',
        needsSubstitution: knownSubstitution == null,
      );
    }
  }

  for (final f in selectedFormations) {
    final id = f['id'];
    final kind = _map(f['spec'])['kind'];
    final members = _strings(_map(f['spec'])['cards']);
    if (untargeted &&
        (view.context['opening_used'] != true ||
            view.context['underground'] == true) &&
        members.length >= 2 &&
        cards.every(
          (card) => card.zone == 'formation' && card.allocation == id,
        ) &&
        cardIds.length < members.length &&
        !result.any(
          (intent) =>
              intent.type == 'open-formation' &&
              intent.seed['formation_id'] == id,
        )) {
      add(
        'open-formation',
        {'formation_id': id, 'formation': _map(f['spec'])},
        'Reconfigure this entire formation. Select its exact members and '
            'protection target; any existing Doppelganger role remains fixed.',
        needsProtection: const {'people', 'great-people'}.contains(kind),
      );
    }
    if (cards.every((c) => c.zone == 'formation' && c.allocation == id) &&
        (untargeted || onHand)) {
      add(
        'take-back',
        {'formation_id': id},
        'Take back this entire formation; any fixed Doppelganger binding remains.',
      );
    }
    if (onHand || onDeck) continue;
    if (kind == 'fate' && !view.quota('fate')) {
      add(
        'fate',
        {
          'formation_id': id,
          if (foreignTarget) 'target_seat': targetSeat,
          if (cards.length == 1 && cards.single.rank == 13) 'selection': ids,
        },
        'Choose the opponent and the Fate king returned on successful resolution.',
      );
    }
    if (kind == 'justice' && !view.quota('justice')) {
      add(
        'justice',
        {
          if (cards.length == 1 && cards.single.rank == 12)
            'ace': cards.single.id,
        },
        'Choose the Justice queen to return, then resolve each offered public or concealed-card choice.',
        sourceFormationId: id,
      );
    }
    if (kind == 'code' && !view.quota('code')) {
      add('code', {
        'formation_id': id,
      }, 'Choose the Diamond bonus and purchase price for Code.');
    }
    if (kind == 'kidnapper' && !view.quota('kidnapper')) {
      add(
        'kidnapper',
        {'formation_id': id, if (foreignTarget) 'target_seat': targetSeat},
        'Choose an opponent. A concealed card is chosen at random; you cannot select its hidden identity.',
      );
    }
  }

  if (!onHand && !onDeck) {
    final queens = cards
        .where(
          (c) =>
              c.rank == 12 &&
              c.suit == 'hearts' &&
              c.zone == 'attachment' &&
              c.allocation == 'hearts',
        )
        .toList();
    if (queens.length == 1 && !view.quota('barricade')) {
      final queen = queens.single;
      add(
        'barricade-sacrifice',
        {
          'selection': [queen.id, ...ids.where((id) => id != queen.id)],
        },
        'Choose five other controlled cards after the Barricade queen; replacement draws are restricted until next round.',
      );
    }
    if (cards.length == 1 &&
        cards.single.zone == 'attachment' &&
        cards.single.rank == 13) {
      final type = cards.single.suit == 'diamonds'
          ? 'ringleader'
          : 'richer-sacrifice';
      if (!view.quota(type == 'ringleader' ? 'ringleader' : 'richer')) {
        add(
          type,
          {'selection': ids},
          'Sacrifice this attached king if the action resolves after responses.',
        );
      }
    }
    final dexters = cards
        .where(
          (c) =>
              c.rank == 11 &&
              c.suit == 'diamonds' &&
              c.zone == 'attachment' &&
              c.allocation == 'diamonds',
        )
        .toList();
    final payments = cards
        .where(
          (card) =>
              card.number && card.suit == 'diamonds' && card.zone == 'series',
        )
        .toList();
    final dexterSelection =
        dexters.length <= 1 &&
        payments.length <= 1 &&
        dexters.length + payments.length == cards.length;
    final hasDexter = view.cards.values.any(
      (c) =>
          c.owner == view.seat &&
          c.rank == 11 &&
          c.suit == 'diamonds' &&
          c.zone == 'attachment' &&
          c.allocation == 'diamonds' &&
          view.available(c),
    );
    if (dexterSelection &&
        (dexters.isNotEmpty || payments.isNotEmpty && hasDexter) &&
        !view.quota('dexter') &&
        (destination == null || foreignTarget && destination.royal)) {
      add(
        'dexter-assassination',
        {
          if (payments.isNotEmpty) 'selection': [payments.single.id],
          if (destination != null) 'targets': [destination.id],
        },
        'Choose one available exposed number Diamond as payment and an exposed enemy royal. The Diamond counts toward support before it is paid.',
      );
    }
  }

  final aces = cards.where((c) => c.ace).toList();
  final exiles = cards
      .where(
        (c) =>
            c.rank == 12 &&
            c.suit == 'clubs' &&
            c.zone == 'attachment' &&
            c.allocation == 'clubs',
      )
      .toList();
  final infiltrators = cards
      .where(
        (c) =>
            c.rank == 11 &&
            c.suit == 'clubs' &&
            c.zone == 'attachment' &&
            c.allocation == 'clubs',
      )
      .toList();
  final baronIds = {
    for (final formation in ownFormations)
      if (_map(formation['spec'])['kind'] == 'baron') formation['id'],
  };
  final baronCards = cards
      .where((c) => c.zone == 'formation' && baronIds.contains(c.allocation))
      .toList();
  final clubs = cards
      .where(
        (c) =>
            c.number &&
            c.suit == 'clubs' &&
            c.zone == 'series' &&
            c.used != view.turn,
      )
      .toList();
  if (!onHand &&
      !onDeck &&
      (untargeted || foreignTarget) &&
      aces.length <= 1 &&
      (aces.isEmpty || !view.aceUsed) &&
      exiles.length <= 1 &&
      infiltrators.length <= 1 &&
      (exiles.isEmpty || !view.quota('exile')) &&
      (infiltrators.isEmpty || !view.quota('infiltrator')) &&
      baronCards.map((card) => card.allocation).toSet().length <= 1 &&
      (baronCards.isEmpty || !view.quota('baron')) &&
      clubs.length +
              aces.length +
              exiles.length +
              infiltrators.length +
              baronCards.length ==
          cards.length) {
    final targetNumber = seriesDestination
        ? targetSeriesCards.isNotEmpty
        : destination == null &&
                  (untargeted || target.zone == TableDropZone.seat) ||
              destination != null &&
                  destination.number &&
                  destination.zone == 'series';
    final targetSuit = seriesDestination ? target?.suit : destination?.suit;
    final bomb =
        destination?.rank == 11 &&
        destination?.suit == 'hearts' &&
        destination?.zone == 'attachment';
    if ((targetNumber || bomb) &&
        (infiltrators.isEmpty || targetSuit != 'diamonds')) {
      final attackSeed = <String, dynamic>{
        if (clubs.isNotEmpty) 'selection': clubs.map((c) => c.id).toList(),
        if (seriesDestination)
          'targets': targetSeriesCards.map((card) => card.id).toList()
        else if (destination != null)
          'targets': [destination.id],
        if (seriesDestination || destination != null)
          'suit': bomb ? destination!.allocation : targetSuit,
        if (aces.isNotEmpty) 'ace': aces.single.id,
        if (exiles.isNotEmpty) 'exile': true,
        if (infiltrators.isNotEmpty) 'infiltrator': true,
      };
      if (baronCards.isEmpty && (infiltrators.isEmpty || bomb)) {
        add(
          bomb ? 'bomb-attack' : 'attack',
          attackSeed,
          '${aces.isNotEmpty ? 'Choose numerical value 2–10, or named Advancement for the Clubs ace. ' : ''}'
          'Choose attacking Clubs and fixed target cards, then review the exact combat total and attack order.',
        );
      }
      if (baronIds.isNotEmpty && !view.quota('baron')) {
        add(
          bomb ? 'baron-bomb-attack' : 'baron-attack',
          {
            ...attackSeed,
            if (!bomb && (seriesDestination || destination != null))
              'targets': view.cards.values
                  .where(
                    (c) =>
                        c.owner == targetSeat &&
                        c.zone == 'series' &&
                        c.number &&
                        c.suit == targetSuit,
                  )
                  .map((c) => c.id)
                  .toList(),
          },
          'Baron targets the entire eligible series, or the Bomb; choose modifiers before committing.',
        );
      }
      if (baronCards.isEmpty &&
          !bomb &&
          targetSuit != 'diamonds' &&
          !view.quota('infiltrator') &&
          view.cards.values.any(
            (c) =>
                c.owner == view.seat &&
                c.rank == 11 &&
                c.suit == 'clubs' &&
                c.zone == 'attachment' &&
                c.allocation == 'clubs' &&
                view.available(c),
          )) {
        add(
          'infiltrator-attack',
          attackSeed,
          'Bypass to a non-Diamond series; the Infiltrator jack is returned when a legal declaration is accepted.',
        );
      }
    }
  }
  if (cards.length == 1 &&
      cards.single.ace &&
      !view.aceUsed &&
      !onHand &&
      !onDeck) {
    final ace = cards.single;
    if (ace.suit == 'hearts' && !foreignTarget && !view.quota('compensation')) {
      add('compensation', {
        'ace': ace.id,
      }, 'Activate Compensation for eligible later losses.');
    }
    if (ace.suit == 'spades' &&
        (untargeted || foreignTarget) &&
        !view.quota('main-inflation') &&
        !view.inflated(targetSeat)) {
      add('main-inflation', {
        'ace': ace.id,
        if (foreignTarget) 'target_seat': targetSeat,
      }, 'Choose an eligible opponent without an existing Inflation effect.');
    }
  }
  return List.unmodifiable(result);
}

/// Validate only fixed owner-authorized Doppelganger roles; missing partner
/// identities stay unknown and cannot be substituted with a different King.
bool formationHonorsFixedBindings(
  Map<String, dynamic> board,
  Set<String> cards, {
  Map<String, dynamic>? substitute,
}) {
  final pair = _strings(substitute?['kings']).toSet();
  final slot = _map(substitute?['slot']);
  for (final binding in _View(board, const {}).ownBindings) {
    final kings = _strings(binding['kings']).toSet();
    if (!kings.any(cards.contains)) continue;
    final fixedSlot = _map(binding['slot']);
    if (kings.length != 2 ||
        !cards.containsAll(kings) ||
        pair.length != 2 ||
        !pair.containsAll(kings) ||
        fixedSlot['rank'] != slot['rank'] ||
        fixedSlot['suit'] != slot['suit'])
      return false;
  }
  return true;
}

/// Named non-combo attachment roles from GAME_RULES 3.3, on current series.
Set<String> royalAttachmentSuits(Map<String, dynamic> board, Set<String> ids) {
  final view = _View(board, const {});
  if (ids.length != 1) return const {};
  final card = view.cards[ids.single];
  if (card == null ||
      card.owner != view.seat ||
      !card.royal ||
      view.ownBindings.any(
        (binding) => _strings(binding['kings']).contains(card.id),
      ) ||
      !const {'hand', 'unassigned'}.contains(card.zone) ||
      !view.available(card))
    return const {};
  return view
      .series(view.seat)
      .where(
        (suit) =>
            card.rank == 13 ||
            card.rank == 12 &&
                const {'clubs', 'hearts'}.contains(card.suit) &&
                card.suit == suit ||
            card.rank == 11 && (card.suit == 'hearts' || card.suit == suit),
      )
      .toSet();
}

/// Drop hints reject contradictions in authorized public state before release.
/// Hidden supply, committed random choices and final legality remain server-owned.
List<CardPlayIntent> suggestCardDrops({
  required Map<String, dynamic> board,
  required Map<String, dynamic> online,
  required Set<String> cardIds,
  required TableDropTarget target,
}) {
  final view = _View(board, online);
  final intents = suggestCardPlays(
    board: board,
    online: online,
    cardIds: cardIds,
    target: target,
  );
  return List.unmodifiable(
    intents.where((intent) {
      final seed = intent.seed;
      final type = intent.type;
      if (target.seat != null &&
          target.seat != view.seat &&
          const {
            'ringleader',
            'richer-sacrifice',
            'barricade-sacrifice',
            'justice',
            'code',
          }.contains(type))
        return false;
      if (type == 'offer' && seed['terms']?['loan'] == true) {
        final give = _strings(seed['terms']['give']);
        if (give.length != cardIds.length || !cardIds.containsAll(give))
          return false;
      }
      final selected = cardIds
          .map((id) => view.cards[id])
          .whereType<_Card>()
          .toList();
      if (type == 'attach') {
        final suit = seed['suit'] as String?;
        if (suit != null && !view.series(view.seat).contains(suit))
          return false;
      }
      final singleAce = selected.length == 1 && selected.single.ace;
      // An Ace placed on a player's area identifies its legal named context;
      // trading the Ace remains an explicit card action with private terms.
      if (singleAce && type == 'offer') return false;
      if (singleAce &&
          selected.single.suit == 'spades' &&
          target.zone == TableDropZone.seat &&
          type.contains('attack'))
        return false;
      final aceUse = selected.any((c) => c.ace) && type != 'offer';
      if (aceUse && view.series(view.seat).length < 2) return false;
      if (type == 'open-formation') {
        final spec = _map(seed['formation']);
        final kind = spec['kind'];
        if (kind != null && kind != 'underground') {
          if (view.series(view.seat).length < 2 &&
              view.context['underground'] != true)
            return false;
          final defensive = const {'people', 'great-people'}.contains(kind);
          if (!view.series(view.seat).contains(defensive ? 'hearts' : 'clubs'))
            return false;
          if ((intent.needsSubstitution || spec['substitute'] != null) &&
              !view.series(view.seat).contains('hearts'))
            return false;
        }
      }
      if (const {
        'ringleader',
        'richer-sacrifice',
        'barricade-sacrifice',
      }.contains(type)) {
        final source = selected
            .where((c) => c.zone == 'attachment')
            .firstOrNull;
        if (source == null ||
            !view.series(view.seat).contains(source.allocation))
          return false;
        if (type == 'richer-sacrifice' &&
            view.cards.values
                    .where(
                      (c) =>
                          c.owner == view.seat &&
                          c.zone == 'attachment' &&
                          c.allocation == source.allocation &&
                          c.rank == 13,
                    )
                    .length !=
                1)
          return false;
      }
      final attack =
          type.contains('attack') ||
          const {
            'dexter-assassination',
            'fate',
            'justice',
            'kidnapper',
            'ponzi',
            'main-inflation',
          }.contains(type);
      if (attack && board['window_id'] == null) {
        if (view.series(view.seat).length < 2) return false;
        final enemy =
            seed['target_seat'] as int? ??
            (type == 'ponzi' ? seed['value'] as int? : null) ??
            target.seat;
        if (enemy != null && enemy != view.seat && !view.attackable(enemy))
          return false;
        if (type.contains('attack') &&
            !view.series(view.seat).contains('clubs'))
          return false;
      }
      if (const {
            'fate',
            'justice',
            'kidnapper',
            'ponzi',
            'code',
          }.contains(type) &&
          !view.series(view.seat).contains('clubs'))
        return false;
      if (type.contains('attack')) {
        final attackers = _strings(
          seed['selection'],
        ).map((id) => view.cards[id]).whereType<_Card>().toList();
        final targets = _strings(
          seed['targets'],
        ).map((id) => view.cards[id]).whereType<_Card>().toList();
        if (view.protectedSeries(view.seat, 'clubs')) return false;
        if (targets.isNotEmpty) {
          final enemy = targets.first.owner;
          final suit = seed['suit'] as String? ?? targets.first.suit;
          final bomb = type.contains('bomb');
          if (!bomb && view.protectedSeries(enemy, suit)) return false;
          if (!bomb &&
              view.cards.values.any(
                (c) =>
                    c.owner == enemy &&
                    c.zone == 'attachment' &&
                    c.rank == 11 &&
                    c.suit == 'hearts' &&
                    c.allocation == suit &&
                    view.series(enemy).contains(suit),
              ))
            return false;
          final bypass =
              type == 'infiltrator-attack' || seed['infiltrator'] == true;
          final ordered = ['hearts', 'clubs', 'spades', 'diamonds'];
          if (!bypass &&
              view
                  .series(enemy)
                  .any(
                    (current) =>
                        ordered.indexOf(current) < ordered.indexOf(suit),
                  ))
            return false;
          // With a missing Ace mode the user still has a real value choice.
          if (attackers.isNotEmpty &&
              (!seed.containsKey('ace') || seed['value'] != null)) {
            final value = seed['value'] as int? ?? 0;
            final modifier = seed.containsKey('ace')
                ? (value == 0 ? 10 : value)
                : 0;
            final attack = attackers.fold<int>(
              modifier * 2,
              (sum, c) => sum + c.rank * 2,
            );
            final defense = targets.fold<int>(
              0,
              (sum, c) =>
                  sum +
                  c.rank * (c.suit == 'spades' && view.inflated(enemy) ? 1 : 2),
            );
            if (bomb
                ? attack < 10
                : type.startsWith('baron-')
                ? attack <= defense
                : attack != defense)
              return false;
          }
        }
      }
      // These are physical costs. A passive attached role is not an activation.
      if (type == 'dexter-assassination') {
        final diamonds = view.cards.values.where(
          (c) =>
              c.owner == view.seat &&
              c.zone == 'series' &&
              c.number &&
              c.suit == 'diamonds',
        );
        if (diamonds.fold<int>(0, (sum, c) => sum + c.rank) < 20) return false;
      }
      return true;
    }),
  );
}

/// Some effect stages deliberately offer an empty selection (for example,
/// declining Dexter prevention). It answers only that exact decision; it is
/// neither a response pass nor permission to dismiss/withdraw an accepted action.
CardPlayIntent? emptyDecisionIntent(
  Map<String, dynamic> board,
  Map<String, dynamic> online,
) {
  final view = _View(board, online);
  final window = board['window_id'];
  final decision = board['decision_id'];
  if (!view.live ||
      board['required_actor'] != view.seat ||
      window is! String ||
      window.isEmpty ||
      decision is! String ||
      decision.isEmpty ||
      !_choices(board).any((choice) => choice.isEmpty))
    return null;
  return CardPlayIntent(
    type: 'decision',
    seed: _freeze({
      'pending_action_id': window,
      'decision_id': decision,
      'selection': <String>[],
    }),
    explanation:
        'Confirm the offered empty selection for this effect stage. '
        'This does not pass a response or withdraw the accepted action.',
  );
}

Set<String> draggableCardIds(
  Map<String, dynamic> board,
  Map<String, dynamic> online,
) {
  final view = _View(board, online);
  return Set.unmodifiable({
    for (final id in selectableCardIds(board, online))
      if (view.cards[id]?.owner == view.seat) id,
  });
}

/// Public cards in an authoritative decision may be selected as targets, but
/// ownership never changes merely to make them draggable player sources.
Set<String> selectableCardIds(
  Map<String, dynamic> board,
  Map<String, dynamic> online,
) {
  final view = _View(board, online);
  if (!view.live) return const {};
  if ((board['decision_id'] is String ||
          board['decision_kind'] == 'kidnapper-outcome') &&
      board['required_actor'] == view.seat) {
    return Set.unmodifiable({
      for (final choice in _choices(board))
        for (final id in choice)
          if (view.cards[id] case final card?)
            if (view.visible(card) &&
                (card.owner != view.seat || view.available(card)))
              id,
    });
  }
  return Set.unmodifiable({
    for (final card in view.cards.values)
      if (card.owner == view.seat &&
          view.visible(card) &&
          view.available(card) &&
          _playable.contains(card.zone))
        card.id,
  });
}

const _suits = ['hearts', 'diamonds', 'clubs', 'spades'];
const _playable = {
  'hand',
  'concealed-ace',
  'series',
  'attachment',
  'formation',
  'unassigned',
};
Map<String, dynamic> _map(dynamic value) =>
    value is Map ? Map<String, dynamic>.from(value) : {};
List<Map<String, dynamic>> _maps(dynamic value) => value is List
    ? value.whereType<Map>().map((m) => Map<String, dynamic>.from(m)).toList()
    : [];
List<String> _strings(dynamic value) =>
    value is List ? value.whereType<String>().toList() : [];
List<List<String>> _choices(Map<String, dynamic> board) =>
    board['choices'] is List
    ? (board['choices'] as List).whereType<List>().map(_strings).toList()
    : [];
dynamic _freeze(dynamic value) {
  if (value is Map)
    return Map<String, dynamic>.unmodifiable(
      value.map((key, value) => MapEntry(key as String, _freeze(value))),
    );
  if (value is List) return List.unmodifiable(value.map(_freeze));
  return value;
}

class _Card {
  _Card(this.raw) : face = _map(raw['card']);
  final Map<String, dynamic> raw, face;
  String get id => face['id'] as String? ?? '';
  int get rank => face['rank'] as int? ?? 0;
  String get suit => face['suit'] as String? ?? '';
  int get owner => raw['controller'] as int? ?? 0;
  String get zone => raw['zone'] as String? ?? '';
  String get allocation => raw['allocation'] as String? ?? '';
  int get used => raw['used_turn'] as int? ?? 0;
  bool get number => rank >= 2 && rank <= 10;
  bool get royal => rank >= 11 && rank <= 13;
  bool get ace => rank == 1 && zone == 'concealed-ace';
}

class _View {
  _View(this.board, this.online);
  final Map<String, dynamic> board, online;
  late final players = _maps(board['players']);
  late final formations = _maps(board['formations']);
  late final ownBindings = board.containsKey('own_bindings')
      ? _maps(board['own_bindings'])
      : [
          // Older snapshots already expose these exact active roles. A dormant
          // hidden binding is never inferred from artwork or card identity.
          for (final formation in formations)
            if (formation['controller'] == seat &&
                _map(formation['spec'])['substitute'] is Map)
              _map(_map(formation['spec'])['substitute']),
        ];
  late final context = _map(online['rules_context']);
  late final cards = {
    for (final raw in _maps(board['cards']))
      if (_Card(raw).id.isNotEmpty) _Card(raw).id: _Card(raw),
  };
  int get seat => board['seat'] as int? ?? 0;
  int get turn => board['turn'] as int? ?? 0;
  Set<String> get coupQueens {
    if (seat <= 0 || board['phase'] != 'playing') return const {};
    final ally = _map(online['alliances'])['$seat'];
    if (ally is int && ally != 0) return const {};
    final own = players.where((player) => player['seat'] == seat);
    if (own.isEmpty ||
        own.single['departed'] == true ||
        !_strings(own.single['history']).contains('clubs'))
      return const {};
    final queens = cards.values
        .where(
          (card) =>
              card.owner == seat &&
              card.rank == 12 &&
              const {'hearts', 'diamonds'}.contains(card.suit) &&
              visible(card) &&
              available(card),
        )
        .toList();
    if (queens.where((card) => card.suit == 'hearts').length != 2 ||
        queens.where((card) => card.suit == 'diamonds').length != 2)
      return const {};
    return {for (final card in queens) card.id};
  }

  bool get live =>
      seat > 0 &&
      board['phase'] == 'playing' &&
      online['server_pending'] != true &&
      online['automatic_pending'] != true &&
      online['departure_pending'] != true &&
      online['round_closed'] != true &&
      players.any(
        (p) =>
            p['seat'] == seat && p['confined'] != true && p['departed'] != true,
      );
  bool get disclosureTiming =>
      seat > 0 &&
      board['phase'] == 'playing' &&
      online['server_pending'] != true &&
      online['automatic_pending'] != true &&
      online['departure_pending'] != true &&
      online['round_closed'] != true &&
      players.any((p) => p['seat'] == seat && p['departed'] != true);
  Set<String> series(int owner) => cards.values
      .where((c) => c.owner == owner && c.zone == 'series' && c.number)
      .map((c) => c.suit)
      .toSet();
  bool attackable(int other) =>
      eligibleParty(other) &&
      players.any(
        (p) => p['seat'] == other && _strings(p['history']).toSet().length >= 2,
      );
  bool protectedSeries(int owner, String suit) => formations.any((f) {
    final spec = _map(f['spec']);
    return f['controller'] == owner &&
        const {'people', 'great-people'}.contains(spec['kind']) &&
        _map(spec['protection'])['series'] == suit &&
        series(owner).contains('hearts') &&
        (series(owner).length >= 2 ||
            owner == seat && context['underground'] == true);
  });
  bool eligibleParty(int other) => players.any(
    (p) => p['seat'] == other && p['departed'] != true && p['confined'] != true,
  );
  bool allianceCompatible(int other) {
    final alliances = _map(online['alliances']);
    int ally(int seat) =>
        alliances['$seat'] as int? ??
        players.where((p) => p['seat'] == seat).firstOrNull?['ally'] as int? ??
        0;
    return (ally(seat) == 0 || ally(seat) == other) &&
        (ally(other) == 0 || ally(other) == seat);
  }

  bool inflated(int? other) =>
      other != null &&
      _maps(board['effects']).any(
        (effect) => effect['kind'] == 'inflation' && effect['target'] == other,
      );
  bool get aceUsed => context['ace_used_this_round'] == true;
  bool quota(String name) =>
      _map(context['quota_used'])[name] == true ||
      name == 'compensation' && context['compensation_used'] == true;
  bool available(_Card card) =>
      (card.raw['available_from_round'] as int? ?? 0) <=
      (board['round'] as int? ?? 0);
  bool visible(_Card card) =>
      card.rank >= 1 &&
      card.rank <= 13 &&
      _suits.contains(card.suit) &&
      _playable.contains(card.zone) &&
      (card.owner == seat ||
          !const {'hand', 'concealed-ace'}.contains(card.zone));
}

Set<String> _formationKinds(List<(int, String)> slots) {
  final result = <String>{};
  bool all(int rank) => slots.isNotEmpty && slots.every((s) => s.$1 == rank);
  int count(int rank, String suit) =>
      slots.where((s) => s == (rank, suit)).length;
  if (all(13)) {
    if (slots.length == 4 && slots.toSet().length == 4) result.add('fate');
    if (slots.length == 3 &&
        count(13, 'diamonds') > 0 &&
        _suits.any((s) => count(13, s) >= (s == 'diamonds' ? 3 : 2)))
      result.add('baron');
    if (slots.length >= 5) result.add('underground');
  }
  if (all(12)) {
    if (slots.length == 2 && count(12, 'hearts') == 2)
      result.add('great-people');
    if (slots.length == 4 && slots.toSet().length == 4) result.add('justice');
    if (slots.length >= 5) result.add('code');
    if (slots.length == 4 &&
        count(12, 'hearts') == 2 &&
        count(12, 'diamonds') == 2)
      result.add('coup');
  }
  if (all(11)) {
    if (slots.length == 2 && slots.toSet().length == 1) result.add('kidnapper');
    if (slots.length == 4 &&
        count(11, 'clubs') == 2 &&
        count(11, 'spades') == 2)
      result.add('ponzi');
  }
  return result;
}
