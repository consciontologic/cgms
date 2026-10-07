import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'components.dart';
import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

/// A shared public board and complete private hand, without player/hand tabs.
///
/// Only presentation is owned here. Physical identities, permitted selection,
/// financial values, effects and action content remain supplied by the host.
/// Every supplied physical card has its own complete artwork and independent
/// selection and magnification gestures.
class CgmsActionBoard extends StatelessWidget {
  const CgmsActionBoard({
    super.key,
    required this.data,
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
    this.showArt = true,
    this.endTurnEnabled = true,
    this.endTurnExplanation,
    this.selectedCardIds = const {},
    this.actionPanel,
    this.header,
    this.footer,
    this.publicGroupLabels = const {},
  });

  final TableViewData data;
  final ValueChanged<String> onSelect;
  final ValueChanged<VisibleCard> onInspect;
  final ValueChanged<ComboView>? onInspectCombo;
  final ValueChanged<String>? onSelectHand;
  final VoidCallback? onClearHandSelection;
  final VoidCallback onPrimary, onTrade, onScores, onEndTurn, onReconnect;
  final bool showArt, endTurnEnabled;
  final String? endTurnExplanation;
  final Set<String> selectedCardIds;
  final Widget? actionPanel, header, footer;

  /// Optional adapter-calculated labels, keyed by `seatId:suitName`.
  /// The default describes the suit only, never inferring effective strength.
  final Map<String, String> publicGroupLabels;

  bool get _active => data.connection == ConnectionStateView.connected;
  bool get _selectable => _active && data.pending == null;
  bool _selected(VisibleCard card) =>
      selectedCardIds.contains(card.id) || data.selectedCardId == card.id;

  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    return LayoutBuilder(
      builder: (context, constraints) {
        final scale = MediaQuery.textScalerOf(context).scale(14) / 14;
        final wide = constraints.maxWidth >= 1050 && scale < 1.5;
        final padding = constraints.maxWidth < 600 ? 8.0 : 18.0;
        return SingleChildScrollView(
          key: const ValueKey('action-board-scroll'),
          padding: EdgeInsets.all(padding),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              header ??
                  Row(
                    children: [
                      Expanded(
                        child: Wrap(
                          spacing: 12,
                          children: [
                            Text(
                              strings.actionBoardRound(
                                data.round,
                                data.totalRounds,
                              ),
                              style: Theme.of(context).textTheme.titleMedium,
                            ),
                            Text(
                              data.turnLabel,
                              style: const TextStyle(
                                color: CgmsColors.cyan,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                      IconButton(
                        onPressed: onScores,
                        tooltip: strings.scoresAndDebt,
                        icon: const Icon(Icons.receipt_long),
                      ),
                    ],
                  ),
              if (!_active) ...[
                CgmsNotice(
                  title: switch (data.connection) {
                    ConnectionStateView.loading => strings.loadingTable,
                    ConnectionStateView.error => strings.tableLoadError,
                    _ => strings.disconnectedActionsPaused,
                  },
                  message: strings.connectionExplanation,
                  icon: Icons.cloud_off,
                ),
                TextButton.icon(
                  onPressed: onReconnect,
                  icon: const Icon(Icons.refresh),
                  label: Text(strings.restoreConnection),
                ),
              ],
              const SizedBox(height: 2),
              if (wide)
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(flex: 3, child: _table(context, cardWidth: 116)),
                    const SizedBox(width: 24),
                    Expanded(
                      flex: 2,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          _actions(context),
                          const SizedBox(height: 12),
                          _hand(context, large: false),
                          const SizedBox(height: 12),
                          _footer(context),
                        ],
                      ),
                    ),
                  ],
                )
              else ...[
                _table(
                  context,
                  cardWidth: constraints.maxWidth >= 600 ? 82 : 52,
                ),
                const SizedBox(height: 4),
                _actions(context),
                const SizedBox(height: 4),
                _hand(context, large: constraints.maxWidth >= 600),
                const SizedBox(height: 8),
                _footer(context),
              ],
            ],
          ),
        );
      },
    );
  }

  Widget _table(BuildContext context, {required double cardWidth}) {
    final strings = CgmsLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          strings.actionBoardPublic,
          style: Theme.of(context).textTheme.titleSmall,
        ),
        const SizedBox(height: 4),
        if (data.seats.isEmpty)
          CgmsNotice(
            title: strings.touchNoPlayers,
            message: strings.touchNoPlayersDetail,
          )
        else
          LayoutBuilder(
            builder: (context, box) {
              final scale = MediaQuery.textScalerOf(context).scale(14) / 14;
              final columns = box.maxWidth >= 340 && scale < 1.6 ? 2 : 1;
              final width = (box.maxWidth - 12 * (columns - 1)) / columns;
              // Rows share their top edge while remaining content-sized.
              return Column(
                children: [
                  for (
                    var index = 0;
                    index < data.seats.length;
                    index += columns
                  )
                    Padding(
                      padding: const EdgeInsets.only(bottom: 4),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          for (
                            var offset = 0;
                            offset < columns &&
                                index + offset < data.seats.length;
                            offset++
                          ) ...[
                            if (offset != 0) const SizedBox(width: 12),
                            SizedBox(
                              width: width,
                              child: _seat(
                                context,
                                data.seats[index + offset],
                                cardWidth: cardWidth,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                ],
              );
            },
          ),
      ],
    );
  }

  Widget _seat(
    BuildContext context,
    SeatView seat, {
    required double cardWidth,
  }) {
    final strings = CgmsLocalizations.of(context);
    final allocated = seat.combos
        .expand((c) => c.cards)
        .map((c) => c.id)
        .toSet();
    final free = seat.exposed.where((c) => !allocated.contains(c.id));
    final groups =
        <
          ({String id, String label, List<VisibleCard> cards, ComboView? combo})
        >[
          for (final suit in [
            CardSuit.diamonds,
            CardSuit.clubs,
            CardSuit.hearts,
            CardSuit.spades,
          ])
            if (free.any((c) => c.suit == suit))
              (
                id: suit.name,
                label:
                    publicGroupLabels['${seat.id}:${suit.name}'] ??
                    _suit(strings, suit),
                cards: free.where((c) => c.suit == suit).toList(),
                combo: null,
              ),
          for (final combo in seat.combos)
            (
              id: 'combo-${combo.id}',
              label: '${combo.title}\n${combo.stateLabel}',
              cards: combo.cards,
              combo: combo,
            ),
        ];
    return Container(
      key: ValueKey('action-seat-${seat.id}'),
      padding: const EdgeInsets.only(top: 1),
      decoration: const BoxDecoration(
        border: Border(top: BorderSide(color: CgmsColors.border)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${seat.number}  ${seat.name}${seat.isSelf ? ' · ${strings.actionBoardYou}' : ''}',
            style: const TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 14,
              height: 1.2,
            ),
          ),
          Text(
            strings.actionBoardCounts(
              seat.concealed.handCount,
              seat.concealed.aceCount,
            ),
            style: const TextStyle(
              color: CgmsColors.muted,
              fontSize: 12,
              height: 1.15,
            ),
          ),
          const SizedBox(height: 4),
          if (groups.isEmpty) Text(strings.touchNoPublicCards),
          LayoutBuilder(
            builder: (context, box) {
              final scale = MediaQuery.textScalerOf(context).scale(14) / 14;
              final desired =
                  (box.maxWidth - 6 * math.max(0, groups.length - 1)) /
                  math.max(1, groups.length);
              final itemWidth = math.min(
                box.maxWidth,
                math.max(102.0, desired),
              );
              final width = math.min(
                math.max(102.0, cardWidth * scale),
                itemWidth,
              );
              return Wrap(
                spacing: 6,
                runSpacing: 8,
                children: [
                  for (final group in groups)
                    SizedBox(
                      width: itemWidth,
                      child: _pile(
                        context,
                        seat,
                        id: group.id,
                        label: group.label,
                        cards: group.cards,
                        combo: group.combo,
                        width: math.min(width, box.maxWidth),
                      ),
                    ),
                ],
              );
            },
          ),
          if (seat.status.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: Text(
                seat.status,
                style: const TextStyle(fontSize: 11, color: CgmsColors.muted),
              ),
            ),
        ],
      ),
    );
  }

  Widget _pile(
    BuildContext context,
    SeatView seat, {
    required String id,
    required String label,
    required List<VisibleCard> cards,
    required double width,
    ComboView? combo,
  }) {
    return Column(
      key: ValueKey('board-pile-${seat.id}-$id'),
      mainAxisSize: MainAxisSize.min,
      children: [
        for (final card in cards)
          Padding(
            padding: const EdgeInsets.only(bottom: 4),
            child: CgmsCard(
              key: ValueKey('board-public-${card.id}'),
              card: card,
              width: width - 2,
              compact: true,
              thumbnail: true,
              showArt: showArt,
              selected: _selected(card),
              onTap: _selectable ? () => onSelect(card.id) : null,
              onInspect: () => onInspect(card),
            ),
          ),
        Text(
          label,
          textAlign: TextAlign.center,
          style: const TextStyle(
            fontSize: 11,
            color: CgmsColors.muted,
            height: 1.15,
          ),
        ),
        if (combo != null && onInspectCombo != null)
          TextButton(
            onPressed: () => onInspectCombo!(combo),
            child: Text(
              CgmsLocalizations.of(context).inspectCombo(combo.title),
            ),
          )
        else if (combo != null)
          Text(combo.description),
      ],
    );
  }

  Widget _hand(BuildContext context, {required bool large}) {
    final strings = CgmsLocalizations.of(context);
    final self = data.seats.where((seat) => seat.isSelf).firstOrNull;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Divider(height: 8),
        Wrap(
          spacing: 12,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(
              '${strings.yourHand} · ${data.hand.length}',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            Text(
              strings.actionBoardPrivate,
              style: const TextStyle(fontSize: 12, color: CgmsColors.muted),
            ),
            if (self != null)
              Text(
                strings.actionBoardAces(self.concealed.aceCount),
                style: const TextStyle(fontSize: 12, color: CgmsColors.muted),
              ),
          ],
        ),
        const SizedBox(height: 4),
        if (data.hand.isEmpty)
          CgmsNotice(
            title: strings.emptyHandTitle,
            message: strings.emptyHandMessage,
          )
        else
          LayoutBuilder(
            builder: (context, box) {
              final scale = MediaQuery.textScalerOf(context).scale(14) / 14;
              final minimum = math.max(102.0, 50.0 * scale);
              final columns = math.max(
                1,
                math.min(
                  large ? 9 : 6,
                  ((box.maxWidth + 4) / (minimum + 4)).floor(),
                ),
              );
              final width = (box.maxWidth - (columns - 1) * 4) / columns;
              return Wrap(
                spacing: 4,
                runSpacing: 4,
                children: [
                  for (final card in data.hand)
                    CgmsCard(
                      key: ValueKey('board-hand-${card.id}'),
                      card: card,
                      width: width - 2,
                      thumbnail: true,
                      compact: true,
                      showArt: showArt,
                      selected: _selected(card),
                      onInspect: () => onInspect(card),
                      onTap: _selectable && onSelectHand != null
                          ? () => onSelectHand!(card.id)
                          : null,
                    ),
                ],
              );
            },
          ),
        if (data.handCombos.isNotEmpty)
          Wrap(
            spacing: 8,
            children: [
              for (final combo in data.handCombos)
                TextButton(
                  onPressed: onInspectCombo == null
                      ? null
                      : () => onInspectCombo!(combo),
                  child: Text('${combo.title} · ${combo.stateLabel}'),
                ),
            ],
          ),
        if (onClearHandSelection != null && (_selectedInHand.isNotEmpty))
          TextButton(
            onPressed: _selectable ? onClearHandSelection : null,
            child: Text(strings.clearSelection),
          ),
      ],
    );
  }

  Iterable<VisibleCard> get _selectedInHand => data.hand.where(_selected);

  Widget _actions(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    if (data.pending case final pending?) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(pending.title, style: Theme.of(context).textTheme.titleMedium),
          Text(pending.detail),
          Wrap(
            spacing: 12,
            children: [
              for (var i = 0; i < pending.responders.length; i++)
                Text(
                  '${pending.responders[i]} · ${i < pending.responseIndex
                      ? strings.touchResponsePassed
                      : i == pending.responseIndex
                      ? strings.touchResponseAwaiting
                      : strings.touchResponseWaiting}',
                  style: TextStyle(
                    color: i == pending.responseIndex
                        ? CgmsColors.cyan
                        : CgmsColors.muted,
                  ),
                ),
            ],
          ),
          if (actionPanel != null)
            actionPanel!
          else
            FilledButton(
              onPressed: _active && data.actionEnabled ? onPrimary : null,
              child: Text(data.actionLabel),
            ),
        ],
      );
    }
    if (actionPanel != null) return actionPanel!;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(data.objective, style: Theme.of(context).textTheme.titleMedium),
        Text(data.actionExplanation),
        FilledButton(
          onPressed: _active && data.actionEnabled ? onPrimary : null,
          child: Text(data.actionLabel),
        ),
      ],
    );
  }

  Widget _footer(BuildContext context) {
    if (footer != null) return footer!;
    final strings = CgmsLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 12,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            TextButton(onPressed: onTrade, child: Text(strings.tradeCards)),
            FilledButton(
              onPressed: _selectable && endTurnEnabled && data.seats.isNotEmpty
                  ? onEndTurn
                  : null,
              child: Text(strings.actionBoardEndTurn),
            ),
          ],
        ),
        if (!endTurnEnabled && endTurnExplanation != null)
          Text(endTurnExplanation!),
      ],
    );
  }
}

String _suit(CgmsLocalizations strings, CardSuit suit) => switch (suit) {
  CardSuit.hearts => strings.hearts,
  CardSuit.diamonds => strings.diamonds,
  CardSuit.clubs => strings.clubs,
  CardSuit.spades => strings.spades,
};
