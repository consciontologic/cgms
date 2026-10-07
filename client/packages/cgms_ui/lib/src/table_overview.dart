part of 'feature_views.dart';

/// Layout state only. The embedding consumer still owns every game decision.
class _CompactTableLayout extends StatefulWidget {
  const _CompactTableLayout({required this.view});
  final CgmsTableView view;

  @override
  State<_CompactTableLayout> createState() => _CompactTableLayoutState();
}

class _CompactTableLayoutState extends State<_CompactTableLayout> {
  final _boardScroll = ScrollController();
  final _handAnchor = GlobalKey();
  final _handFocus = FocusNode(debugLabel: 'Compact hand region');
  String? _reviewedPublicId;

  void _reviewPublic(VisibleCard card) {
    setState(() => _reviewedPublicId = card.id);
    widget.view.onSelect(card.id);
  }

  @override
  void dispose() {
    _boardScroll.dispose();
    _handFocus.dispose();
    super.dispose();
  }

  void _jumpToHand() {
    _handFocus.requestFocus();
    Scrollable.ensureVisible(
      _handAnchor.currentContext!,
      duration: MediaQuery.disableAnimationsOf(context)
          ? Duration.zero
          : const Duration(milliseconds: 180),
      alignment: 0.02,
    );
  }

  @override
  Widget build(BuildContext context) {
    final view = widget.view;
    final data = view.data;
    final strings = CgmsLocalizations.of(context);
    final theme = Theme.of(context);
    final active = data.connection == ConnectionStateView.connected;
    final selectedPublic = data.seats
        .where((seat) => seat.isSelf)
        .expand((seat) => seat.exposed)
        .where((card) => card.id == (_reviewedPublicId ?? data.selectedCardId))
        .firstOrNull;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 20,
          runSpacing: 4,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              data.pending == null ? strings.yourMove : strings.atTheTable,
              style: theme.textTheme.headlineSmall,
            ),
            Text(
              strings.roundPhase(data.round, data.totalRounds, data.phaseLabel),
            ),
            Text(
              data.turnLabel,
              style: const TextStyle(color: CgmsColors.cyan),
            ),
          ],
        ),
        if (!active) ...[
          const SizedBox(height: 8),
          CgmsNotice(
            title: switch (data.connection) {
              ConnectionStateView.loading => strings.loadingTable,
              ConnectionStateView.error => strings.tableLoadError,
              _ => strings.disconnectedActionsPaused,
            },
            message: strings.connectionExplanation,
            icon: Icons.cloud_off,
          ),
          OutlinedButton.icon(
            onPressed: view.onReconnect,
            icon: const Icon(Icons.refresh),
            label: Text(
              data.connection == ConnectionStateView.error
                  ? strings.retryConnection
                  : strings.restoreConnection,
            ),
          ),
        ],
        LayoutBuilder(
          builder: (context, box) {
            final wide =
                box.maxWidth >= 1100 &&
                MediaQuery.textScalerOf(context).scale(14) <= 18;
            final board = Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Wrap(
                  spacing: 12,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    Text(
                      strings.publicTable,
                      style: theme.textTheme.titleMedium,
                    ),
                    TextButton(
                      key: const ValueKey('jump-hand'),
                      onPressed: _jumpToHand,
                      child: Text(strings.yourHand),
                    ),
                    TextButton.icon(
                      onPressed: view.onScores,
                      icon: const Icon(Icons.receipt_long, size: 18),
                      label: Text(strings.scoresAndDebt),
                    ),
                  ],
                ),
                LayoutBuilder(
                  builder: (context, box) {
                    final largeText =
                        MediaQuery.textScalerOf(context).scale(14) > 18;
                    final columns = (wide || box.maxWidth >= 650) && !largeText
                        ? 2
                        : 1;
                    final width = (box.maxWidth - 8 * (columns - 1)) / columns;
                    final board = Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        for (final seat in data.seats)
                          SizedBox(
                            key: ValueKey('overview-seat-${seat.id}'),
                            width: width,
                            child: _CompactSeat(
                              view: view,
                              seat: seat,
                              onChoose: _reviewPublic,
                            ),
                          ),
                      ],
                    );
                    // On medium and small screens the public board scrolls separately,
                    // keeping the hand much closer than four full-height seat panels.
                    // Large text uses natural document flow to avoid a cramped viewport.
                    if (largeText ||
                        (wide &&
                            data.seats.every((seat) => seat.combos.isEmpty)) ||
                        data.seats.length <= columns) {
                      return board;
                    }
                    return Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          strings.scrollPublicTable,
                          style: const TextStyle(color: CgmsColors.muted),
                        ),
                        const SizedBox(height: 4),
                        ConstrainedBox(
                          constraints: BoxConstraints(
                            maxHeight:
                                MediaQuery.sizeOf(context).height *
                                (wide ? .68 : .36),
                          ),
                          child: Scrollbar(
                            controller: _boardScroll,
                            thumbVisibility: true,
                            child: SingleChildScrollView(
                              key: const ValueKey('compact-public-scroll'),
                              controller: _boardScroll,
                              padding: const EdgeInsets.only(right: 12),
                              child: LayoutBuilder(
                                builder: (context, inner) => Wrap(
                                  spacing: 8,
                                  runSpacing: 8,
                                  children: [
                                    for (final seat in data.seats)
                                      SizedBox(
                                        key: ValueKey(
                                          'overview-seat-${seat.id}',
                                        ),
                                        width:
                                            (inner.maxWidth -
                                                8 * (columns - 1)) /
                                            columns,
                                        child: _CompactSeat(
                                          view: view,
                                          seat: seat,
                                          onChoose: _reviewPublic,
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ),
                      ],
                    );
                  },
                ),
              ],
            );
            final hand = Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 10),
                Focus(
                  focusNode: _handFocus,
                  skipTraversal: true,
                  child: Semantics(
                    header: true,
                    child: Text(
                      key: _handAnchor,
                      '${strings.yourHand} · ${strings.privateHandEyebrow}',
                      style: theme.textTheme.titleMedium,
                    ),
                  ),
                ),
                if (data.handCombos.isNotEmpty) ...[
                  Text(
                    strings.handPatterns,
                    style: theme.textTheme.labelMedium,
                  ),
                  for (final combo in data.handCombos)
                    TextButton.icon(
                      key: ValueKey('inspect-combo-${combo.id}'),
                      onPressed: view.onInspectCombo == null
                          ? null
                          : () => view.onInspectCombo!(combo),
                      icon: const Icon(Icons.link, size: 18),
                      label: Text('${combo.title} · ${combo.stateLabel}'),
                    ),
                  if (view.onInspectCombo == null)
                    for (final combo in data.handCombos)
                      Text(combo.description),
                ],
                CgmsHandView(
                  key: ValueKey('hand-${data.gameId}'),
                  cards: data.hand,
                  compact: true,
                  showArt: view.showArt,
                  selectedCardId: data.selectedCardId,
                  onSelect:
                      active &&
                          data.pending == null &&
                          view.onSelectHand != null
                      ? (id) {
                          setState(() => _reviewedPublicId = null);
                          view.onSelectHand!(id);
                        }
                      : null,
                  onClearSelection: active && data.pending == null
                      ? view.onClearHandSelection
                      : null,
                  onInspect: view.onInspect,
                ),
                const SizedBox(height: 8),
                if (selectedPublic != null)
                  Semantics(
                    label: strings.cardIdentity(
                      selectedPublic.rank,
                      _publicSuit(strings, selectedPublic.suit),
                      selectedPublic.copy,
                    ),
                    excludeSemantics: true,
                    child: Text(
                      '${selectedPublic.rank} ${_publicSuit(strings, selectedPublic.suit)}',
                    ),
                  ),
              ],
            );
            final action = Container(
              key: const ValueKey('compact-action-area'),
              padding: const EdgeInsets.all(12),
              decoration: const BoxDecoration(
                color: CgmsColors.surface,
                border: Border(
                  left: BorderSide(color: CgmsColors.cyan, width: 3),
                ),
              ),
              child: LayoutBuilder(
                builder: (context, box) {
                  final description = Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        data.pending?.title ?? data.objective,
                        style: theme.textTheme.titleMedium,
                      ),
                      Text(data.pending?.detail ?? data.actionExplanation),
                      if (data.pending case final pending?)
                        Wrap(
                          spacing: 12,
                          children: [
                            for (var i = 0; i < pending.responders.length; i++)
                              Text(
                                '${i < pending.responseIndex
                                    ? '✓'
                                    : i == pending.responseIndex
                                    ? '→'
                                    : '·'} ${pending.responders[i]}',
                                style: TextStyle(
                                  color: i == pending.responseIndex
                                      ? CgmsColors.cyan
                                      : CgmsColors.muted,
                                ),
                              ),
                          ],
                        ),
                    ],
                  );
                  final action = FilledButton.icon(
                    key: const ValueKey('hand-primary-action'),
                    onPressed: active && data.actionEnabled
                        ? view.onPrimary
                        : null,
                    icon: Icon(
                      data.pending == null ? Icons.arrow_forward : Icons.check,
                    ),
                    label: Text(data.actionLabel),
                  );
                  return box.maxWidth >= 900
                      ? Row(
                          children: [
                            Expanded(child: description),
                            const SizedBox(width: 16),
                            ConstrainedBox(
                              constraints: const BoxConstraints(maxWidth: 350),
                              child: action,
                            ),
                          ],
                        )
                      : Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            description,
                            const SizedBox(height: 8),
                            action,
                          ],
                        );
                },
              ),
            );
            return wide
                ? Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(flex: 6, child: board),
                      const SizedBox(width: 20),
                      Expanded(
                        flex: 5,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [action, hand],
                        ),
                      ),
                    ],
                  )
                : Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [board, hand, const SizedBox(height: 8), action],
                  );
          },
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 12,
          runSpacing: 8,
          children: [
            OutlinedButton.icon(
              onPressed: view.onTrade,
              icon: const Icon(Icons.swap_horiz),
              label: Text(strings.tradeCards),
            ),
            OutlinedButton.icon(
              onPressed:
                  active &&
                      data.pending == null &&
                      data.seats.isNotEmpty &&
                      view.endTurnEnabled
                  ? view.onEndTurn
                  : null,
              icon: const Icon(Icons.last_page),
              label: Text(strings.endTurnAndScore),
            ),
          ],
        ),
        if (!view.endTurnEnabled && view.endTurnExplanation != null)
          Text(
            view.endTurnExplanation!,
            style: const TextStyle(color: CgmsColors.muted),
          ),
        if (data.pending != null) Text(strings.finishResponsesFirst),
        const SizedBox(height: 12),
        ExpansionTile(
          tilePadding: EdgeInsets.zero,
          title: Text(strings.tableJournal),
          children: [
            if (data.events.isEmpty) ListTile(title: Text(strings.noEvents)),
            for (final event in data.events.take(8))
              ListTile(dense: true, title: Text(event)),
          ],
        ),
      ],
    );
  }
}

class _CompactSeat extends StatelessWidget {
  const _CompactSeat({
    required this.view,
    required this.seat,
    required this.onChoose,
  });
  final CgmsTableView view;
  final SeatView seat;
  final ValueChanged<VisibleCard> onChoose;

  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    final active = view.data.connection == ConnectionStateView.connected;
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: CgmsColors.surface,
        border: Border(
          top: BorderSide(
            color: seat.isSelf ? CgmsColors.cyan : CgmsColors.border,
            width: 2,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${seat.number.toString().padLeft(2, '0')}  /  ${seat.name}${seat.isSelf ? ' · ${strings.yourPublicPosition}' : ''}',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          Text(
            seat.status,
            style: const TextStyle(color: CgmsColors.muted, fontSize: 13),
          ),
          const SizedBox(height: 4),
          CgmsSuitStacks(
            cards: _unallocatedCards(seat),
            showArt: view.showArt,
            selectedCardId: view.data.selectedCardId,
            onTap: seat.isSelf && active && view.data.pending == null
                ? onChoose
                : null,
            onInspect: view.onInspect,
          ),
          for (final combo in seat.combos)
            CgmsComboGroup(
              combo: combo,
              showArt: view.showArt,
              onInspectCard: view.onInspect,
              onInspect: view.onInspectCombo == null
                  ? null
                  : () => view.onInspectCombo!(combo),
            ),
          const SizedBox(height: 4),
          Row(
            children: [
              ConcealedCard(width: 16, showArt: view.showArt),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  strings.concealedCounts(
                    seat.concealed.handCount,
                    seat.concealed.aceCount,
                  ),
                  style: const TextStyle(color: CgmsColors.muted, fontSize: 13),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

String _publicSuit(CgmsLocalizations strings, CardSuit suit) => switch (suit) {
  CardSuit.hearts => strings.hearts,
  CardSuit.diamonds => strings.diamonds,
  CardSuit.clubs => strings.clubs,
  CardSuit.spades => strings.spades,
};
