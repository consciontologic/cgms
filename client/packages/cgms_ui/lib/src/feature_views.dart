import 'package:flutter/material.dart';

import 'components.dart';
import 'combo_views.dart';
import 'card_collections.dart';
import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

part 'table_overview.dart';

/// A local form only; match creation and navigation belong to the consumer.
class CgmsSetupView extends StatefulWidget {
  const CgmsSetupView({
    super.key,
    required this.onStart,
    this.initialName = '',
    this.startLabel,
    this.noticeTitle,
    this.noticeMessage,
    this.footer = '',
  });
  final String initialName, footer;
  final String? startLabel, noticeTitle, noticeMessage;
  final void Function(String name, int games) onStart;
  @override
  State<CgmsSetupView> createState() => _CgmsSetupViewState();
}

class _CgmsSetupViewState extends State<CgmsSetupView> {
  final _form = GlobalKey<FormState>();
  late final _name = TextEditingController(text: widget.initialName);
  int _games = 3;
  @override
  void dispose() {
    _name.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final strings = CgmsLocalizations.of(context);
    return LayoutBuilder(
      builder: (context, constraints) {
        final wide = constraints.maxWidth >= 800;
        final intro = Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _Eyebrow(strings.setupEyebrow),
            const SizedBox(height: 20),
            Text(strings.takeYourSeat, style: theme.textTheme.displayLarge),
            const SizedBox(height: 20),
            Text(
              strings.setupTagline,
              style: theme.textTheme.headlineSmall?.copyWith(
                color: CgmsColors.muted,
              ),
            ),
            const SizedBox(height: 32),
            Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                CgmsCard(
                  card: const VisibleCard(
                    id: 'preview-heart',
                    rank: '4',
                    suit: CardSuit.hearts,
                    copy: 1,
                  ),
                  width: wide ? 128 : 88,
                ),
                CgmsCard(
                  card: const VisibleCard(
                    id: 'preview-queen',
                    rank: 'Q',
                    suit: CardSuit.spades,
                    copy: 1,
                  ),
                  width: wide ? 128 : 88,
                ),
                CgmsCard(
                  card: const VisibleCard(
                    id: 'preview-club',
                    rank: '7',
                    suit: CardSuit.clubs,
                    copy: 1,
                  ),
                  width: wide ? 128 : 88,
                ),
              ],
            ),
            const SizedBox(height: 20),
            Text(
              strings.editionName,
              style: const TextStyle(color: CgmsColors.muted, letterSpacing: 2),
            ),
          ],
        );
        final form = _Surface(
          child: Form(
            key: _form,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  strings.setupFormTitle,
                  style: theme.textTheme.headlineMedium,
                ),
                const SizedBox(height: 12),
                Text(strings.setupDeckDescription),
                const SizedBox(height: 24),
                TextFormField(
                  controller: _name,
                  maxLength: 32,
                  decoration: InputDecoration(
                    labelText: strings.yourName,
                    hintText: strings.nameHint,
                  ),
                  validator: (value) => (value?.trim().isEmpty ?? true)
                      ? strings.nameRequired
                      : null,
                  textInputAction: TextInputAction.done,
                  onFieldSubmitted: (_) => _start(),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<int>(
                  initialValue: _games,
                  isExpanded: true,
                  decoration: InputDecoration(
                    labelText: strings.matchGameCount,
                  ),
                  items: [
                    DropdownMenuItem(value: 1, child: Text(strings.oneGame)),
                    DropdownMenuItem(value: 3, child: Text(strings.threeGames)),
                  ],
                  onChanged: (value) => setState(() => _games = value!),
                ),
                const SizedBox(height: 24),
                CgmsNotice(
                  title: widget.noticeTitle ?? strings.nextGame,
                  message: widget.noticeMessage ?? strings.setupNotice,
                ),
                const SizedBox(height: 24),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton.icon(
                    onPressed: _start,
                    icon: const Icon(Icons.arrow_forward),
                    label: Text(widget.startLabel ?? strings.startMatch),
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  widget.footer,
                  style: const TextStyle(color: CgmsColors.muted),
                ),
              ],
            ),
          ),
        );
        return wide
            ? Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(flex: 6, child: intro),
                  const SizedBox(width: 56),
                  Expanded(flex: 5, child: form),
                ],
              )
            : Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [intro, const SizedBox(height: 32), form],
              );
      },
    );
  }

  void _start() {
    if (_form.currentState!.validate())
      widget.onStart(_name.text.trim(), _games);
  }
}

/// A seat-authorized projection. This view does not calculate game legality.
class CgmsTableView extends StatefulWidget {
  const CgmsTableView({
    super.key,
    required this.data,
    this.showArt = true,
    required this.onSelect,
    required this.onInspect,
    required this.onPrimary,
    required this.onTrade,
    required this.onScores,
    required this.onEndTurn,
    required this.onReconnect,
    this.onInspectCombo,
    this.onSelectHand,
    this.onClearHandSelection,
    this.endTurnEnabled = true,
    this.endTurnExplanation,
    this.compact = false,
  });
  final TableViewData data;
  final bool showArt;
  final ValueChanged<String> onSelect;
  final ValueChanged<VisibleCard> onInspect;
  final ValueChanged<ComboView>? onInspectCombo;
  final VoidCallback onPrimary, onTrade, onScores, onEndTurn, onReconnect;
  final ValueChanged<String>? onSelectHand;
  final VoidCallback? onClearHandSelection;
  final bool endTurnEnabled;
  final String? endTurnExplanation;

  /// A shared public board and smaller hand for comparing positions at a glance.
  /// Large text and narrow consumers retain natural reflow and scrolling.
  final bool compact;

  @override
  State<CgmsTableView> createState() => _CgmsTableViewState();
}

class _CgmsTableViewState extends State<CgmsTableView> {
  final _handAnchor = GlobalKey();
  final _exposedAnchor = GlobalKey();
  final _publicAnchor = GlobalKey();
  final _handFocus = FocusNode(debugLabel: 'Hand region');
  final _exposedFocus = FocusNode(debugLabel: 'Exposed cards region');
  final _publicFocus = FocusNode(debugLabel: 'Public table region');

  @override
  void dispose() {
    _handFocus.dispose();
    _exposedFocus.dispose();
    _publicFocus.dispose();
    super.dispose();
  }

  void _jumpTo(GlobalKey key, FocusNode focus) {
    final destination = key.currentContext;
    if (destination == null) return;
    focus.requestFocus();
    Scrollable.ensureVisible(
      destination,
      duration: MediaQuery.disableAnimationsOf(context)
          ? Duration.zero
          : const Duration(milliseconds: 180),
      alignment: 0.02,
    );
  }

  @override
  Widget build(BuildContext context) {
    if (widget.compact) return _CompactTableLayout(view: widget);
    final data = widget.data;
    final showArt = widget.showArt;
    final onSelect = widget.onSelect;
    final onInspect = widget.onInspect;
    final onPrimary = widget.onPrimary;
    final onTrade = widget.onTrade;
    final onScores = widget.onScores;
    final onEndTurn = widget.onEndTurn;
    final onReconnect = widget.onReconnect;
    final theme = Theme.of(context);
    final strings = CgmsLocalizations.of(context);
    final active = data.connection == ConnectionStateView.connected;
    final self = data.seats.where((s) => s.isSelf).firstOrNull;
    final dense =
        data.hand.length >= 10 ||
        data.seats.any((seat) => seat.exposed.length >= 8);
    final handAction = widget.onSelectHand == null
        ? null
        : _Surface(
            accent: CgmsColors.cyan,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  data.pending?.title ?? data.objective,
                  style: theme.textTheme.titleLarge,
                ),
                const SizedBox(height: 8),
                Text(data.pending?.detail ?? data.actionExplanation),
                const SizedBox(height: 12),
                FilledButton.icon(
                  key: const ValueKey('hand-primary-action'),
                  onPressed: active && data.actionEnabled ? onPrimary : null,
                  icon: Icon(
                    data.pending == null ? Icons.arrow_forward : Icons.check,
                  ),
                  label: Text(data.actionLabel),
                ),
              ],
            ),
          );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          alignment: WrapAlignment.spaceBetween,
          crossAxisAlignment: WrapCrossAlignment.center,
          spacing: 24,
          runSpacing: 12,
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _Eyebrow(
                  strings.roundPhase(
                    data.round,
                    data.totalRounds,
                    data.phaseLabel.toUpperCase(),
                  ),
                ),
                Text(
                  data.pending == null ? strings.yourMove : strings.atTheTable,
                  style: theme.textTheme.displayMedium,
                ),
              ],
            ),
            Text(
              data.turnLabel,
              style: theme.textTheme.titleLarge?.copyWith(
                color: CgmsColors.cyan,
              ),
            ),
          ],
        ),
        const SizedBox(height: 20),
        if (!active) ...[
          CgmsNotice(
            title: switch (data.connection) {
              ConnectionStateView.loading => strings.loadingTable,
              ConnectionStateView.error => strings.tableLoadError,
              _ => strings.disconnectedActionsPaused,
            },
            message: strings.connectionExplanation,
            icon: Icons.cloud_off,
          ),
          const SizedBox(height: 8),
          OutlinedButton.icon(
            onPressed: onReconnect,
            icon: const Icon(Icons.refresh),
            label: Text(
              data.connection == ConnectionStateView.error
                  ? strings.retryConnection
                  : strings.restoreConnection,
            ),
          ),
          const SizedBox(height: 20),
        ],
        _Surface(
          accent: CgmsColors.cyan,
          child: LayoutBuilder(
            builder: (context, box) {
              final description = Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    data.pending?.title ?? data.objective,
                    style: theme.textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 8),
                  Text(data.pending?.detail ?? data.actionExplanation),
                  if (data.pending case final pending?) ...[
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        for (var i = 0; i < pending.responders.length; i++)
                          Chip(
                            avatar: Icon(
                              i < pending.responseIndex
                                  ? Icons.check
                                  : i == pending.responseIndex
                                  ? Icons.more_horiz
                                  : Icons.hourglass_empty,
                              size: 18,
                            ),
                            label: Text(pending.responders[i]),
                          ),
                      ],
                    ),
                  ],
                ],
              );
              final button = FilledButton.icon(
                onPressed: active && data.actionEnabled ? onPrimary : null,
                icon: Icon(
                  data.pending == null ? Icons.arrow_forward : Icons.check,
                ),
                label: Text(data.actionLabel),
              );
              return box.maxWidth > 750
                  ? Row(
                      children: [
                        Expanded(child: description),
                        const SizedBox(width: 24),
                        ConstrainedBox(
                          constraints: const BoxConstraints(maxWidth: 280),
                          child: button,
                        ),
                      ],
                    )
                  : Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        description,
                        const SizedBox(height: 16),
                        button,
                      ],
                    );
            },
          ),
        ),
        const SizedBox(height: 20),
        if (dense) ...[
          Text(strings.jumpToCards, style: theme.textTheme.labelMedium),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              OutlinedButton(
                key: const ValueKey('jump-hand'),
                onPressed: () => _jumpTo(_handAnchor, _handFocus),
                child: Text(strings.yourHand),
              ),
              if (self != null)
                OutlinedButton(
                  key: const ValueKey('jump-exposed'),
                  onPressed: () => _jumpTo(_exposedAnchor, _exposedFocus),
                  child: Text(strings.yourExposedCards),
                ),
              OutlinedButton(
                key: const ValueKey('jump-public'),
                onPressed: () => _jumpTo(_publicAnchor, _publicFocus),
                child: Text(strings.publicTable),
              ),
            ],
          ),
          const SizedBox(height: 20),
        ],
        Focus(
          focusNode: _publicFocus,
          skipTraversal: true,
          child: CgmsSectionTitle(
            key: _publicAnchor,
            title: strings.publicTable,
            eyebrow: strings.readTheirPosition,
            trailing: TextButton.icon(
              onPressed: onScores,
              icon: const Icon(Icons.receipt_long),
              label: Text(strings.scoresAndDebt),
            ),
          ),
        ),
        const SizedBox(height: 12),
        LayoutBuilder(
          builder: (context, box) {
            final opponents = data.seats.where((s) => !s.isSelf).toList();
            final columns = box.maxWidth >= 1100
                ? 3
                : box.maxWidth >= 720
                ? 2
                : 1;
            final width = (box.maxWidth - 12 * (columns - 1)) / columns;
            return Wrap(
              spacing: 12,
              runSpacing: 12,
              children: [
                for (final seat in opponents)
                  SizedBox(
                    width: width,
                    child: _SeatPanel(
                      seat: seat,
                      showArt: showArt,
                      onInspect: onInspect,
                      onInspectCombo: widget.onInspectCombo,
                    ),
                  ),
              ],
            );
          },
        ),
        const SizedBox(height: 24),
        if (self != null) ...[
          Focus(
            focusNode: _exposedFocus,
            skipTraversal: true,
            child: CgmsSectionTitle(
              key: _exposedAnchor,
              title: strings.exposedCardsOf(self.name),
              eyebrow: strings.yourPublicPosition,
            ),
          ),
          const SizedBox(height: 6),
          Text(self.status, style: const TextStyle(color: CgmsColors.muted)),
          const SizedBox(height: 12),
          CgmsExposedCardsView(
            cards: _unallocatedCards(self),
            showArt: showArt,
            selectedCardId: data.selectedCardId,
            onSelect: active && data.pending == null ? onSelect : null,
            onInspect: onInspect,
            cardWidth: self.exposed.length >= 8 ? 96 : 102,
          ),
          for (final combo in self.combos)
            CgmsComboGroup(
              combo: combo,
              showArt: showArt,
              onInspectCard: onInspect,
              onInspect: widget.onInspectCombo == null
                  ? null
                  : () => widget.onInspectCombo!(combo),
            ),
          const SizedBox(height: 24),
        ],
        Focus(
          focusNode: _handFocus,
          skipTraversal: true,
          child: CgmsSectionTitle(
            key: _handAnchor,
            title: strings.yourHand,
            eyebrow: strings.privateHandEyebrow,
          ),
        ),
        if (data.handCombos.isNotEmpty) ...[
          Text(strings.handPatterns, style: theme.textTheme.labelMedium),
          for (final combo in data.handCombos)
            TextButton.icon(
              key: ValueKey('inspect-combo-${combo.id}'),
              onPressed: widget.onInspectCombo == null
                  ? null
                  : () => widget.onInspectCombo!(combo),
              icon: const Icon(Icons.link, size: 18),
              label: Text('${combo.title} · ${combo.stateLabel}'),
            ),
          if (widget.onInspectCombo == null)
            for (final combo in data.handCombos) Text(combo.description),
        ],
        CgmsHandView(
          key: ValueKey('hand-${data.gameId}'),
          cards: data.hand,
          showArt: showArt,
          selectedCardId: data.selectedCardId,
          onSelect: active && data.pending == null ? widget.onSelectHand : null,
          onClearSelection: active && data.pending == null
              ? widget.onClearHandSelection
              : null,
          onInspect: onInspect,
        ),
        if (handAction != null) ...[const SizedBox(height: 16), handAction],
        const SizedBox(height: 24),
        Wrap(
          spacing: 12,
          runSpacing: 12,
          children: [
            OutlinedButton.icon(
              onPressed: onTrade,
              icon: const Icon(Icons.swap_horiz),
              label: Text(strings.tradeCards),
            ),
            OutlinedButton.icon(
              onPressed:
                  active &&
                      data.pending == null &&
                      data.seats.isNotEmpty &&
                      widget.endTurnEnabled
                  ? onEndTurn
                  : null,
              icon: const Icon(Icons.last_page),
              label: Text(strings.endTurnAndScore),
            ),
          ],
        ),
        if (!widget.endTurnEnabled && widget.endTurnExplanation != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              widget.endTurnExplanation!,
              style: const TextStyle(color: CgmsColors.muted),
            ),
          ),
        if (data.pending != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              strings.finishResponsesFirst,
              style: const TextStyle(color: CgmsColors.muted),
            ),
          ),
        const SizedBox(height: 28),
        CgmsSectionTitle(
          title: strings.tableJournal,
          eyebrow: strings.eventEyebrow,
        ),
        const SizedBox(height: 8),
        if (data.events.isEmpty)
          Text(strings.noEvents)
        else
          for (final event in data.events.take(8))
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 6),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(
                    Icons.subdirectory_arrow_right,
                    color: CgmsColors.muted,
                    size: 18,
                  ),
                  const SizedBox(width: 10),
                  Expanded(child: Text(event)),
                ],
              ),
            ),
      ],
    );
  }
}

class _SeatPanel extends StatelessWidget {
  const _SeatPanel({
    required this.seat,
    required this.showArt,
    required this.onInspect,
    this.onInspectCombo,
  });
  final SeatView seat;
  final bool showArt;
  final ValueChanged<VisibleCard> onInspect;
  final ValueChanged<ComboView>? onInspectCombo;
  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    return _Surface(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${seat.number.toString().padLeft(2, '0')}  /  ${seat.name}',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          Text(seat.status, style: const TextStyle(color: CgmsColors.muted)),
          const SizedBox(height: 16),
          CgmsExposedCardsView(
            cards: _unallocatedCards(seat),
            showArt: showArt,
            onInspect: onInspect,
            cardWidth: seat.exposed.length >= 8 ? 84 : 62,
          ),
          for (final combo in seat.combos)
            CgmsComboGroup(
              combo: combo,
              showArt: showArt,
              onInspectCard: onInspect,
              onInspect: onInspectCombo == null
                  ? null
                  : () => onInspectCombo!(combo),
            ),
          const SizedBox(height: 16),
          Row(
            children: [
              ConcealedCard(width: 28, showArt: showArt),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  strings.concealedCounts(
                    seat.concealed.handCount,
                    seat.concealed.aceCount,
                  ),
                  style: const TextStyle(color: CgmsColors.muted),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

/// Controlled transfer view. The adapter determines all available actions.
class CgmsTradeView extends StatelessWidget {
  const CgmsTradeView({
    super.key,
    required this.give,
    required this.receive,
    required this.status,
    required this.description,
    this.showArt = true,
    this.onOffer,
    this.onAccept,
    this.onDecline,
    this.onWithdraw,
    this.acceptLabel,
    this.declineLabel,
  });
  final VisibleCard give, receive;
  final String status, description;
  final String? acceptLabel, declineLabel;
  final bool showArt;
  final VoidCallback? onOffer, onAccept, onDecline, onWithdraw;
  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _Eyebrow(strings.tradeEyebrow),
        Text(
          strings.makeADeal,
          style: Theme.of(context).textTheme.displayMedium,
        ),
        const SizedBox(height: 12),
        Text(description),
        const SizedBox(height: 28),
        _Surface(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(status, style: Theme.of(context).textTheme.headlineSmall),
              const SizedBox(height: 24),
              Wrap(
                spacing: 16,
                runSpacing: 24,
                crossAxisAlignment: WrapCrossAlignment.center,
                children: [
                  Column(
                    children: [
                      Text(strings.youGive),
                      const SizedBox(height: 12),
                      CgmsCard(
                        card: give,
                        width: MediaQuery.sizeOf(context).width < 600
                            ? 96
                            : 128,
                        showArt: showArt,
                      ),
                    ],
                  ),
                  const Icon(
                    Icons.swap_horiz,
                    color: CgmsColors.cyan,
                    size: 36,
                  ),
                  Column(
                    children: [
                      Text(strings.youReceive),
                      const SizedBox(height: 12),
                      CgmsCard(
                        card: receive,
                        width: MediaQuery.sizeOf(context).width < 600
                            ? 96
                            : 128,
                        showArt: showArt,
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 24),
              Text(strings.tradeExplanation),
              const SizedBox(height: 24),
              Wrap(
                spacing: 12,
                runSpacing: 12,
                children: [
                  if (onOffer != null)
                    FilledButton.icon(
                      onPressed: onOffer,
                      icon: const Icon(Icons.send_outlined),
                      label: Text(strings.proposeTrade),
                    ),
                  if (onAccept != null)
                    FilledButton(
                      onPressed: onAccept,
                      child: Text(acceptLabel ?? strings.acceptOffer),
                    ),
                  if (onDecline != null)
                    OutlinedButton(
                      onPressed: onDecline,
                      child: Text(declineLabel ?? strings.declineOffer),
                    ),
                  if (onWithdraw != null)
                    TextButton(
                      onPressed: onWithdraw,
                      child: Text(strings.withdrawOffer),
                    ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }
}

/// Scores are already computed by the adapter. No balances are derived here.
class CgmsScoresView extends StatelessWidget {
  const CgmsScoresView({
    super.key,
    required this.seats,
    required this.title,
    required this.description,
    required this.finalized,
    this.onFinish,
  });
  final List<SeatView> seats;
  final String title, description;
  final bool finalized;
  final VoidCallback? onFinish;
  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _Eyebrow(
          finalized ? strings.resultRecorded : strings.pointsHaveHistory,
        ),
        Text(title, style: Theme.of(context).textTheme.displayMedium),
        const SizedBox(height: 16),
        Text(description),
        const SizedBox(height: 28),
        for (final seat in seats)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: _Surface(
              accent: seat.isSelf ? CgmsColors.cyan : null,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    seat.name,
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 28,
                    runSpacing: 16,
                    children: [
                      _Amount(
                        strings.matchScore,
                        seat.finances.matchScore.format(),
                      ),
                      _Amount(
                        strings.thisGame,
                        seat.finances.gameScore.format(),
                      ),
                      _Amount(
                        strings.availablePoints,
                        seat.finances.available.format(),
                      ),
                      _Amount(strings.debt, seat.finances.debt.format()),
                    ],
                  ),
                ],
              ),
            ),
          ),
        const SizedBox(height: 12),
        CgmsNotice(
          title: strings.scoreNoticeTitle,
          message: strings.scoreNoticeMessage,
        ),
        if (onFinish != null) ...[
          const SizedBox(height: 20),
          FilledButton.icon(
            onPressed: onFinish,
            icon: const Icon(Icons.check),
            label: Text(strings.finishSettlement),
          ),
        ],
      ],
    );
  }
}

class _Amount extends StatelessWidget {
  const _Amount(this.label, this.value);
  final String label, value;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(
        label,
        style: const TextStyle(
          color: CgmsColors.muted,
          fontSize: 14,
          letterSpacing: 1,
        ),
      ),
      Text(value, style: Theme.of(context).textTheme.headlineMedium),
    ],
  );
}

class CgmsCardInspector extends StatelessWidget {
  const CgmsCardInspector({super.key, required this.card, this.showArt = true});
  final VisibleCard card;
  final bool showArt;
  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    final suit = switch (card.suit) {
      CardSuit.hearts => strings.hearts,
      CardSuit.diamonds => strings.diamonds,
      CardSuit.clubs => strings.clubs,
      CardSuit.spades => strings.spades,
    };
    return AlertDialog(
      title: Semantics(
        label: strings.cardIdentity(card.rank, suit, card.copy),
        excludeSemantics: true,
        child: Text('${card.rank} $suit'),
      ),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CgmsCard(card: card, width: 180, showArt: showArt),
            const SizedBox(height: 16),
            Text(strings.artworkCosmetic),
          ],
        ),
      ),
      actions: [
        TextButton(
          autofocus: true,
          onPressed: () => Navigator.of(context).pop(),
          child: Text(strings.closeInspection),
        ),
      ],
    );
  }
}

class _Eyebrow extends StatelessWidget {
  const _Eyebrow(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Text(
    text,
    style: const TextStyle(
      color: CgmsColors.muted,
      fontSize: 14,
      letterSpacing: 2,
      height: 1.6,
    ),
  );
}

class _Surface extends StatelessWidget {
  const _Surface({required this.child, this.accent});
  final Widget child;
  final Color? accent;
  @override
  Widget build(BuildContext context) => SizedBox(
    width: double.infinity,
    child: CgmsPanel(accent: accent, child: child),
  );
}

List<VisibleCard> _unallocatedCards(SeatView seat) {
  final allocated = seat.combos
      .expand((combo) => combo.cards)
      .map((card) => card.id)
      .toSet();
  return seat.exposed.where((card) => !allocated.contains(card.id)).toList();
}
