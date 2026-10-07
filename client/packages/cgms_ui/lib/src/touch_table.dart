import 'package:flutter/material.dart';

import 'card_collections.dart';
import 'combo_views.dart';
import 'components.dart';
import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

/// A bounded touch table for Scaffold.body. The consumer supplies authorized
/// cards, action eligibility and callbacks; this view owns only browsing state.
class CgmsTouchTableView extends StatefulWidget {
  const CgmsTouchTableView({
    super.key,
    required this.data,
    this.showArt = true,
    required this.onSelect,
    required this.onInspect,
    this.onInspectCombo,
    this.onSelectHand,
    this.onClearHandSelection,
    required this.onPrimary,
    required this.onTrade,
    required this.onScores,
    required this.onEndTurn,
    required this.onReconnect,
    this.endTurnEnabled = true,
    this.endTurnDisabledReason,
  });

  final TableViewData data;
  final bool showArt;
  final ValueChanged<String> onSelect;
  final ValueChanged<VisibleCard> onInspect;
  final ValueChanged<ComboView>? onInspectCombo;
  final ValueChanged<String>? onSelectHand;
  final VoidCallback? onClearHandSelection;
  final VoidCallback onPrimary, onTrade, onScores, onEndTurn, onReconnect;
  final bool endTurnEnabled;
  final String? endTurnDisabledReason;

  @override
  State<CgmsTouchTableView> createState() => _CgmsTouchTableViewState();
}

class _CgmsTouchTableViewState extends State<CgmsTouchTableView> {
  final _publicScroll = ScrollController();
  final _handScroll = ScrollController();
  final _activityScroll = ScrollController();
  final _naturalScroll = ScrollController();
  final _activityAnchor = GlobalKey();
  int _tab = 0;
  String? _seatId;
  String? _reviewedPublicId;

  TableViewData get data => widget.data;
  bool get _active => data.connection == ConnectionStateView.connected;
  SeatView? get _seat =>
      data.seats.where((seat) => seat.id == _seatId).firstOrNull ??
      data.seats.where((seat) => seat.isSelf).firstOrNull ??
      data.seats.firstOrNull;

  @override
  void didUpdateWidget(CgmsTouchTableView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.data.gameId != data.gameId) {
      _tab = 0;
      _seatId = null;
      _reviewedPublicId = null;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        for (final scroll in [
          _publicScroll,
          _handScroll,
          _activityScroll,
          _naturalScroll,
        ]) {
          if (scroll.hasClients) scroll.jumpTo(0);
        }
      });
    }
  }

  @override
  void dispose() {
    _publicScroll.dispose();
    _handScroll.dispose();
    _activityScroll.dispose();
    _naturalScroll.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    return LayoutBuilder(
      key: const ValueKey('touch-table-view'),
      builder: (context, constraints) {
        final natural =
            !constraints.hasBoundedHeight ||
            constraints.maxHeight < 560 ||
            MediaQuery.textScalerOf(context).scale(14) > 21;
        final tablet = constraints.maxWidth >= 600;
        final panels = [
          _publicPosition(strings),
          _hand(strings),
          _activity(strings),
        ];
        if (natural) {
          return SingleChildScrollView(
            key: const ValueKey('touch-natural-scroll'),
            controller: _naturalScroll,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                _header(strings),
                _seatRail(strings),
                _navigation(strings),
                for (var index = 0; index < panels.length; index++)
                  Offstage(
                    offstage: _tab != index,
                    child: ExcludeFocus(
                      excluding: _tab != index,
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: panels[index],
                      ),
                    ),
                  ),
                _dock(strings),
              ],
            ),
          );
        }
        final public = _scroll('touch-public-scroll', _publicScroll, panels[0]);
        final hand = _scroll('touch-hand-scroll', _handScroll, panels[1]);
        final activity = _scroll(
          'touch-activity-scroll',
          _activityScroll,
          panels[2],
        );
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _header(strings),
            _seatRail(strings),
            Expanded(
              child: tablet
                  ? Row(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Expanded(
                          child: IndexedStack(
                            index: _tab == 2 ? 1 : 0,
                            children: [public, activity],
                          ),
                        ),
                        const VerticalDivider(width: 1),
                        Expanded(child: hand),
                      ],
                    )
                  : IndexedStack(
                      index: _tab,
                      children: [public, hand, activity],
                    ),
            ),
            _dock(strings),
            _navigation(strings, tablet: tablet),
          ],
        );
      },
    );
  }

  Widget _scroll(String key, ScrollController controller, Widget child) =>
      Scrollbar(
        controller: controller,
        child: SingleChildScrollView(
          key: ValueKey(key),
          controller: controller,
          padding: const EdgeInsets.all(16),
          child: child,
        ),
      );

  Widget _header(CgmsLocalizations strings) => Padding(
    padding: const EdgeInsets.fromLTRB(16, 8, 8, 4),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    data.turnLabel,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  Text(
                    strings.roundPhase(
                      data.round,
                      data.totalRounds,
                      data.phaseLabel,
                    ),
                    style: const TextStyle(
                      color: CgmsColors.muted,
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            IconButton(
              key: const ValueKey('touch-scores'),
              tooltip: strings.scoresAndDebt,
              onPressed: widget.onScores,
              icon: const Icon(Icons.receipt_long),
            ),
          ],
        ),
        if (!_active)
          Wrap(
            spacing: 8,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [
              Text(switch (data.connection) {
                ConnectionStateView.loading => strings.loadingTable,
                ConnectionStateView.error => strings.tableLoadError,
                _ => strings.disconnectedActionsPaused,
              }, style: const TextStyle(color: CgmsColors.muted, fontSize: 13)),
              TextButton(
                onPressed: widget.onReconnect,
                child: Text(
                  data.connection == ConnectionStateView.error
                      ? strings.retryConnection
                      : strings.restoreConnection,
                ),
              ),
            ],
          ),
      ],
    ),
  );

  Widget _seatRail(CgmsLocalizations strings) => SingleChildScrollView(
    scrollDirection: Axis.horizontal,
    padding: const EdgeInsets.fromLTRB(12, 4, 12, 8),
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (final seat in data.seats)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: SizedBox(
              width: 176,
              child: Semantics(
                selected: _seat?.id == seat.id,
                child: OutlinedButton(
                  key: ValueKey('touch-seat-${seat.id}'),
                  onPressed: () => setState(() {
                    _seatId = seat.id;
                    _tab = 0;
                    if (_publicScroll.hasClients) _publicScroll.jumpTo(0);
                  }),
                  style:
                      OutlinedButton.styleFrom(
                        backgroundColor: _seat?.id == seat.id
                            ? CgmsColors.raised
                            : CgmsColors.canvas,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 10,
                        ),
                      ).copyWith(
                        side: WidgetStateProperty.resolveWith(
                          (states) => BorderSide(
                            color: states.contains(WidgetState.focused)
                                ? CgmsColors.lime
                                : _seat?.id == seat.id
                                ? CgmsColors.cyan
                                : CgmsColors.border,
                            width: states.contains(WidgetState.focused) ? 3 : 1,
                          ),
                        ),
                      ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${seat.number.toString().padLeft(2, '0')}  ${seat.name}',
                        style: const TextStyle(
                          fontWeight: FontWeight.w700,
                          fontSize: 14,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        strings.touchSeatSummary(
                          seat.exposed.length,
                          seat.combos.length,
                        ),
                        style: const TextStyle(
                          fontSize: 12,
                          color: CgmsColors.muted,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
      ],
    ),
  );

  Widget _navigation(CgmsLocalizations strings, {bool tablet = false}) =>
      NavigationBar(
        selectedIndex: tablet ? (_tab == 2 ? 1 : 0) : _tab,
        onDestinationSelected: (index) =>
            setState(() => _tab = tablet ? (index == 1 ? 2 : 0) : index),
        animationDuration: MediaQuery.disableAnimationsOf(context)
            ? Duration.zero
            : const Duration(milliseconds: 160),
        height: 72 + (MediaQuery.textScalerOf(context).scale(12) - 12),
        backgroundColor: CgmsColors.surface,
        indicatorColor: CgmsColors.raised,
        destinations: [
          NavigationDestination(
            key: const ValueKey('touch-tab-table'),
            icon: const Icon(Icons.view_module_outlined),
            label: strings.touchTable,
          ),
          if (!tablet)
            NavigationDestination(
              key: const ValueKey('touch-tab-hand'),
              icon: const Icon(Icons.style_outlined),
              label: strings.touchHand,
            ),
          NavigationDestination(
            key: const ValueKey('touch-tab-activity'),
            icon: const Icon(Icons.history),
            label: strings.touchActivity,
          ),
        ],
      );

  Widget _publicPosition(CgmsLocalizations strings) {
    final seat = _seat;
    if (seat == null)
      return CgmsNotice(
        title: strings.touchNoPlayers,
        message: strings.touchNoPlayersDetail,
      );
    final allocated = {
      for (final combo in seat.combos)
        for (final card in combo.cards) card.id,
    };
    final ordinary = seat.exposed
        .where((card) => !allocated.contains(card.id))
        .toList();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          strings.exposedCardsOf(seat.name),
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 4),
        Text(seat.status, style: const TextStyle(color: CgmsColors.muted)),
        const SizedBox(height: 12),
        if (ordinary.isNotEmpty)
          CgmsSuitStacks(
            cards: ordinary,
            showArt: widget.showArt,
            selectedCardId: data.selectedCardId,
            onTap: seat.isSelf && _active && data.pending == null
                ? (card) {
                    setState(() => _reviewedPublicId = card.id);
                    widget.onSelect(card.id);
                  }
                : null,
            onInspect: widget.onInspect,
          ),
        for (final combo in seat.combos)
          Padding(
            padding: const EdgeInsets.only(top: 12),
            child: CgmsComboGroup(
              combo: combo,
              showArt: widget.showArt,
              onInspectCard: widget.onInspect,
              onInspect: widget.onInspectCombo == null
                  ? null
                  : () => widget.onInspectCombo!(combo),
            ),
          ),
        if (seat.exposed.isEmpty && seat.combos.isEmpty)
          CgmsNotice(
            title: strings.touchNoPublicCards,
            message: strings.touchNoPublicCardsDetail,
          ),
        const SizedBox(height: 12),
        Row(
          children: [
            ConcealedCard(width: 20, showArt: widget.showArt),
            const SizedBox(width: 8),
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
    );
  }

  Widget _hand(CgmsLocalizations strings) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(strings.yourHand, style: Theme.of(context).textTheme.titleLarge),
      Text(
        strings.privateHandEyebrow,
        style: const TextStyle(fontSize: 12, color: CgmsColors.muted),
      ),
      if (data.handCombos.isNotEmpty) ...[
        const SizedBox(height: 8),
        Text(
          strings.handPatterns,
          style: Theme.of(context).textTheme.labelMedium,
        ),
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
      const SizedBox(height: 8),
      CgmsHandView(
        key: ValueKey('touch-hand-${data.gameId}'),
        cards: data.hand,
        compact: true,
        selectedCardId: data.selectedCardId,
        showArt: widget.showArt,
        onInspect: widget.onInspect,
        onSelect: _active && data.pending == null && widget.onSelectHand != null
            ? (id) {
                setState(() => _reviewedPublicId = null);
                widget.onSelectHand!(id);
              }
            : null,
        onClearSelection: _active && data.pending == null
            ? widget.onClearHandSelection
            : null,
      ),
    ],
  );

  Widget _activity(CgmsLocalizations strings) => Column(
    key: _activityAnchor,
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(
        data.pending?.title ?? data.objective,
        style: Theme.of(context).textTheme.titleLarge,
      ),
      const SizedBox(height: 8),
      Text(data.pending?.detail ?? data.actionExplanation),
      if (data.pending case final pending?) ...[
        const SizedBox(height: 12),
        for (var i = 0; i < pending.responders.length; i++)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 4),
            child: Row(
              children: [
                Icon(
                  i < pending.responseIndex
                      ? Icons.check
                      : i == pending.responseIndex
                      ? Icons.more_horiz
                      : Icons.hourglass_empty,
                  size: 18,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '${pending.responders[i]} · ${i < pending.responseIndex
                        ? strings.touchResponsePassed
                        : i == pending.responseIndex
                        ? strings.touchResponseAwaiting
                        : strings.touchResponseWaiting}',
                  ),
                ),
              ],
            ),
          ),
      ],
      if (!_active) ...[
        const SizedBox(height: 12),
        Text(strings.connectionExplanation),
      ],
      const SizedBox(height: 16),
      Wrap(
        spacing: 8,
        runSpacing: 8,
        children: [
          OutlinedButton.icon(
            onPressed: _active ? widget.onTrade : null,
            icon: const Icon(Icons.swap_horiz),
            label: Text(strings.tradeCards),
          ),
          OutlinedButton(
            onPressed: _active && data.pending == null && widget.endTurnEnabled
                ? widget.onEndTurn
                : null,
            child: Text(strings.endTurnAndScore),
          ),
        ],
      ),
      if (data.pending != null || !widget.endTurnEnabled)
        Padding(
          padding: const EdgeInsets.only(top: 8),
          child: Text(
            data.pending != null
                ? strings.finishResponsesFirst
                : widget.endTurnDisabledReason ?? data.actionExplanation,
            style: const TextStyle(color: CgmsColors.muted),
          ),
        ),
      const SizedBox(height: 20),
      Text(
        strings.tableJournal,
        style: Theme.of(context).textTheme.titleMedium,
      ),
      const SizedBox(height: 8),
      if (data.events.isEmpty) Text(strings.noEvents),
      for (final event in data.events)
        Padding(padding: const EdgeInsets.only(bottom: 12), child: Text(event)),
    ],
  );

  Widget _dock(CgmsLocalizations strings) {
    final selectedHand = data.hand
        .where((card) => card.id == data.selectedCardId)
        .firstOrNull;
    final public = data.seats
        .expand((seat) => seat.exposed)
        .where((card) => card.id == (_reviewedPublicId ?? data.selectedCardId))
        .firstOrNull;
    final selected = _reviewedPublicId != null
        ? public ?? selectedHand
        : selectedHand ?? public;
    return Material(
      color: CgmsColors.surface,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (selected != null)
              Row(
                children: [
                  Expanded(
                    child: Semantics(
                      label: _identity(strings, selected),
                      excludeSemantics: true,
                      child: Text(
                        '${selected.rank} ${_suitName(strings, selected.suit)}',
                        style: const TextStyle(
                          color: CgmsColors.cyan,
                          fontWeight: FontWeight.w700,
                          fontSize: 14,
                        ),
                      ),
                    ),
                  ),
                  if (selectedHand != null &&
                      selected.id == selectedHand.id &&
                      widget.onClearHandSelection != null)
                    IconButton(
                      tooltip: strings.clearSelection,
                      onPressed: _active && data.pending == null
                          ? widget.onClearHandSelection
                          : null,
                      icon: const Icon(Icons.close),
                    ),
                ],
              ),
            Text(
              data.pending?.title ?? data.objective,
              style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700),
            ),
            if (!data.actionEnabled && _active)
              Align(
                alignment: Alignment.centerLeft,
                child: TextButton(
                  key: const ValueKey('touch-move-details'),
                  onPressed: _showMoveDetails,
                  child: Text(strings.touchMoveDetails),
                ),
              ),
            const SizedBox(height: 6),
            FilledButton.icon(
              key: const ValueKey('touch-primary-action'),
              onPressed: _active && data.actionEnabled
                  ? widget.onPrimary
                  : null,
              icon: Icon(
                data.pending == null ? Icons.arrow_forward : Icons.check,
              ),
              label: Text(data.actionLabel),
            ),
          ],
        ),
      ),
    );
  }

  void _showMoveDetails() {
    setState(() => _tab = 2);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      if (_activityScroll.hasClients) _activityScroll.jumpTo(0);
      final anchor = _activityAnchor.currentContext;
      if (anchor != null)
        Scrollable.ensureVisible(anchor, duration: Duration.zero);
    });
  }

  String _identity(CgmsLocalizations strings, VisibleCard card) =>
      strings.cardIdentity(card.rank, _suitName(strings, card.suit), card.copy);

  String _suitName(CgmsLocalizations strings, CardSuit suit) => switch (suit) {
    CardSuit.hearts => strings.hearts,
    CardSuit.diamonds => strings.diamonds,
    CardSuit.clubs => strings.clubs,
    CardSuit.spades => strings.spades,
  };
}
