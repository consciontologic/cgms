import 'package:flutter/material.dart';

import 'components.dart';
import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

/// Browses only the supplied, viewer-authorized hand. Filters are local UI state;
/// selection, action eligibility and all outcomes belong to the consumer.
class CgmsHandView extends StatefulWidget {
  const CgmsHandView({
    super.key,
    required this.cards,
    required this.onInspect,
    this.selectedCardId,
    this.onSelect,
    this.onClearSelection,
    this.showArt = true,
    this.compact = false,
  });

  final List<VisibleCard> cards;
  final ValueChanged<VisibleCard> onInspect;
  final String? selectedCardId;
  final ValueChanged<String>? onSelect;
  final VoidCallback? onClearSelection;
  final bool showArt;

  /// Keeps more cards in view while preserving their identity and hit targets.
  final bool compact;

  @override
  State<CgmsHandView> createState() => _CgmsHandViewState();
}

class _CgmsHandViewState extends State<CgmsHandView> {
  CardSuit? _filter;

  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    final selected = widget.cards
        .where((card) => card.id == widget.selectedCardId)
        .firstOrNull;
    final filtered = widget.cards
        .where((card) => _filter == null || card.suit == _filter)
        .toList();
    final showFilters =
        widget.cards.length >= 8 || widget.onSelect != null || _filter != null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (selected != null && !widget.compact) ...[
          _selectionSummary(context, strings, selected),
          const SizedBox(height: 12),
        ],
        if (showFilters) ...[
          LayoutBuilder(
            builder: (context, constraints) => Wrap(
              spacing: widget.compact ? 4 : 8,
              runSpacing: widget.compact ? 4 : 8,
              children: [
                for (final suit in <CardSuit?>[null, ...CardSuit.values])
                  ChoiceChip(
                    key: ValueKey('hand-filter-${suit?.name ?? 'all'}'),
                    selected: _filter == suit,
                    onSelected: (_) => setState(() => _filter = suit),
                    label: ConstrainedBox(
                      constraints: BoxConstraints(
                        maxWidth: constraints.maxWidth > 110
                            ? constraints.maxWidth - 110
                            : constraints.maxWidth,
                      ),
                      child: Text(
                        strings.suitFilter(
                          suit == null
                              ? strings.allCards
                              : _suitName(strings, suit),
                          widget.cards
                              .where(
                                (card) => suit == null || card.suit == suit,
                              )
                              .length,
                        ),
                      ),
                    ),
                    labelPadding: EdgeInsets.symmetric(
                      horizontal: widget.compact ? 6 : 8,
                      vertical: widget.compact ? 2 : 8,
                    ),
                    labelStyle: TextStyle(
                      color: _filter == suit
                          ? CgmsColors.canvas
                          : CgmsColors.text,
                      fontSize: 14,
                      fontWeight: FontWeight.w700,
                    ),
                    selectedColor: CgmsColors.cyan,
                    backgroundColor: CgmsColors.surface,
                    side: WidgetStateBorderSide.resolveWith(
                      (states) => BorderSide(
                        color: states.contains(WidgetState.focused)
                            ? CgmsColors.lime
                            : CgmsColors.border,
                        width: states.contains(WidgetState.focused) ? 3 : 1,
                      ),
                    ),
                    showCheckmark: false,
                    materialTapTargetSize: MaterialTapTargetSize.padded,
                  ),
              ],
            ),
          ),
          SizedBox(height: widget.compact ? 4 : 12),
        ],
        Semantics(
          liveRegion: true,
          child: Text(
            strings.showingCards(filtered.length, widget.cards.length),
            style: TextStyle(
              color: CgmsColors.muted,
              fontSize: widget.compact ? 12 : null,
            ),
          ),
        ),
        if (widget.onSelect != null && !widget.compact) ...[
          const SizedBox(height: 4),
          Text(
            strings.handSelectionHint,
            style: TextStyle(
              color: CgmsColors.muted,
              fontSize: widget.compact ? 12 : null,
            ),
          ),
        ],
        SizedBox(height: widget.compact ? 4 : 12),
        if (widget.cards.isEmpty)
          CgmsNotice(
            title: strings.emptyHandTitle,
            message: strings.emptyHandMessage,
          )
        else if (filtered.isEmpty)
          CgmsNotice(
            title: strings.noMatchingCards,
            message: strings.changeHandFilter,
          )
        else if (widget.compact)
          CgmsSuitStacks(
            cards: filtered,
            cardKeyPrefix: 'hand-card',
            expandSelectedCard: false,
            selectedCardId: widget.selectedCardId,
            showArt: widget.showArt,
            onTap: widget.onSelect == null
                ? null
                : (card) => widget.onSelect!(card.id),
            onInspect: widget.onInspect,
          )
        else
          LayoutBuilder(
            builder: (context, constraints) => Wrap(
              spacing: 8,
              runSpacing: 12,
              children: [
                for (final card in filtered)
                  CgmsCard(
                    key: ValueKey('hand-card-${card.id}'),
                    card: card,
                    width: constraints.maxWidth < 400 ? 96 : 100,
                    selected: widget.selectedCardId == card.id,
                    showArt: widget.showArt,
                    onTap: widget.onSelect == null
                        ? null
                        : () => widget.onSelect!(card.id),
                    onInspect: () => widget.onInspect(card),
                  ),
              ],
            ),
          ),
        if (selected != null && widget.compact) ...[
          const SizedBox(height: 4),
          _selectionSummary(context, strings, selected),
        ],
      ],
    );
  }

  Widget _selectionSummary(
    BuildContext context,
    CgmsLocalizations strings,
    VisibleCard selected,
  ) {
    final identity = Semantics(
      liveRegion: true,
      label: strings.selectedCard(_identity(strings, selected)),
      excludeSemantics: true,
      child: Text(
        strings.selectedCard(
          '${selected.rank} ${_suitName(strings, selected.suit)}',
        ),
        style: Theme.of(context).textTheme.titleMedium,
      ),
    );
    final controls = [
      if (widget.onClearSelection != null)
        TextButton(
          key: const ValueKey('clear-hand-selection'),
          onPressed: widget.onClearSelection,
          child: Text(strings.clearSelection),
        ),
    ];
    final outsideFilter = _filter != null && selected.suit != _filter;
    final outsideMessage = Text(
      strings.selectionOutsideFilter,
      style: const TextStyle(color: CgmsColors.muted),
    );
    return Container(
      key: const ValueKey('hand-selection-summary'),
      width: double.infinity,
      padding: EdgeInsets.all(widget.compact ? 8 : 16),
      decoration: BoxDecoration(
        color: CgmsColors.surface,
        border: Border.all(color: CgmsColors.cyan),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: widget.compact
            ? [
                Wrap(
                  spacing: 8,
                  runSpacing: 4,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [identity, ...controls],
                ),
                if (outsideFilter) ...[
                  const SizedBox(height: 4),
                  outsideMessage,
                  _filteredSelection(selected),
                ],
              ]
            : [
                identity,
                if (outsideFilter) ...[
                  const SizedBox(height: 4),
                  outsideMessage,
                  _filteredSelection(selected),
                ],
                const SizedBox(height: 8),
                Wrap(spacing: 8, runSpacing: 8, children: controls),
              ],
      ),
    );
  }

  Widget _filteredSelection(VisibleCard selected) => CgmsCard(
    key: const ValueKey('inspect-hand-selection'),
    card: selected,
    width: 100,
    selected: true,
    showArt: widget.showArt,
    onInspect: () => widget.onInspect(selected),
  );
}

/// Public cards are grouped by suit when dense. These groups express identity
/// and counts only; they do not infer open series, combos, powers or legality.
class CgmsExposedCardsView extends StatelessWidget {
  const CgmsExposedCardsView({
    super.key,
    required this.cards,
    required this.onInspect,
    this.onSelect,
    this.selectedCardId,
    this.showArt = true,
    this.cardWidth = 84,
  });

  final List<VisibleCard> cards;
  final ValueChanged<VisibleCard> onInspect;
  final ValueChanged<String>? onSelect;
  final String? selectedCardId;
  final bool showArt;
  final double cardWidth;

  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    Widget row(List<VisibleCard> values) => Wrap(
      spacing: 8,
      runSpacing: 8,
      children: [
        for (final card in values)
          CgmsCard(
            card: card,
            width: cardWidth,
            showArt: showArt,
            selected: card.id == selectedCardId,
            onTap: onSelect == null ? null : () => onSelect!(card.id),
            onInspect: () => onInspect(card),
          ),
      ],
    );
    if (cards.length < 8) return row(cards);
    return LayoutBuilder(
      builder: (context, constraints) {
        const groupGap = 24.0;
        // Three full card footprints, including reserved focus space, must
        // fit in either column. Text and cards retain their requested sizes.
        final faceWidth = cardWidth < 100 ? 100.0 : cardWidth;
        final minimumGroupWidth = 3 * (faceWidth + 12) + 2 * 8;
        final twoColumns =
            constraints.maxWidth.isFinite &&
            constraints.maxWidth >= minimumGroupWidth * 2 + groupGap;
        final groupWidth = twoColumns
            ? (constraints.maxWidth - groupGap) / 2
            : constraints.maxWidth;
        return Wrap(
          spacing: groupGap,
          runSpacing: 16,
          children: [
            for (final suit in CardSuit.values)
              if (cards.any((card) => card.suit == suit))
                SizedBox(
                  width: groupWidth.isFinite ? groupWidth : null,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Semantics(
                        header: true,
                        child: Text(
                          strings.suitGroup(
                            _suitName(strings, suit),
                            cards.where((card) => card.suit == suit).length,
                          ),
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                      ),
                      const SizedBox(height: 4),
                      row(cards.where((card) => card.suit == suit).toList()),
                    ],
                  ),
                ),
          ],
        );
      },
    );
  }
}

String _suitName(CgmsLocalizations strings, CardSuit suit) => switch (suit) {
  CardSuit.hearts => strings.hearts,
  CardSuit.diamonds => strings.diamonds,
  CardSuit.clubs => strings.clubs,
  CardSuit.spades => strings.spades,
};

String _identity(CgmsLocalizations strings, VisibleCard card) =>
    strings.cardIdentity(card.rank, _suitName(strings, card.suit), card.copy);
