import 'package:flutter/material.dart';

import 'components.dart';
import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

/// A formation supplied by the adapter, never inferred from card artwork.
class CgmsComboGroup extends StatelessWidget {
  const CgmsComboGroup({
    super.key,
    required this.combo,
    required this.onInspectCard,
    this.onInspect,
    this.showArt = true,
  });

  final ComboView combo;
  final ValueChanged<VisibleCard> onInspectCard;
  final VoidCallback? onInspect;
  final bool showArt;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.only(left: 8, top: 6, bottom: 6),
    decoration: const BoxDecoration(
      border: Border(left: BorderSide(color: CgmsColors.cyan, width: 2)),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(combo.title, style: Theme.of(context).textTheme.titleMedium),
        Text(
          combo.stateLabel,
          style: const TextStyle(color: CgmsColors.cyan, fontSize: 13),
        ),
        Wrap(
          spacing: 2,
          children: [
            for (final card in combo.cards)
              CgmsCard(
                key: ValueKey('combo-card-${card.id}'),
                card: card,
                compact: true,
                width: 48,
                showArt: showArt,
                onInspect: () => onInspectCard(card),
              ),
          ],
        ),
        if (onInspect != null)
          TextButton.icon(
            key: ValueKey('inspect-combo-${combo.id}'),
            onPressed: onInspect,
            icon: const Icon(Icons.info_outline, size: 18),
            label: Text(
              CgmsLocalizations.of(context).inspectCombo(combo.title),
            ),
          )
        else
          Text(combo.description),
      ],
    ),
  );
}

class CgmsComboInspector extends StatelessWidget {
  const CgmsComboInspector({
    super.key,
    required this.combo,
    this.showArt = true,
  });

  final ComboView combo;
  final bool showArt;

  @override
  Widget build(BuildContext context) => AlertDialog(
    title: Text(combo.title),
    scrollable: true,
    content: SizedBox(
      width: 520,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            combo.stateLabel,
            style: const TextStyle(color: CgmsColors.cyan),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final card in combo.cards)
                CgmsCard(
                  card: card,
                  width: MediaQuery.textScalerOf(context).scale(88),
                  showArt: showArt,
                ),
            ],
          ),
          const SizedBox(height: 12),
          Text(combo.description),
        ],
      ),
    ),
    actions: [
      TextButton(
        onPressed: () => Navigator.of(context).pop(),
        child: Text(CgmsLocalizations.of(context).closeInspection),
      ),
    ],
  );
}
