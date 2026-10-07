import 'package:flutter/material.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'card_intents.dart';

/// The selection-to-draft boundary shared by taps, keyboard and table drops.
/// Choosing an intent never sends a command; the host owns review and submission.
class CardPlayChoices extends StatelessWidget {
  const CardPlayChoices({
    super.key,
    required this.intents,
    required this.onSelected,
    required this.label,
    this.enabled = true,
  });
  final List<CardPlayIntent> intents;
  final ValueChanged<CardPlayIntent> onSelected;
  final String Function(String) label;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    final counts = <String, int>{};
    for (final intent in intents) {
      counts.update(_baseKey(intent), (value) => value + 1, ifAbsent: () => 1);
    }
    final ordinals = <String, int>{};
    final choices = <Widget>[];
    for (final intent in intents) {
      final baseKey = _baseKey(intent);
      final repeated = counts[baseKey]! > 1;
      final ordinal = ordinals.update(
        baseKey,
        (value) => value + 1,
        ifAbsent: () => 1,
      );
      final formationId =
          intent.sourceFormationId ?? intent.seed['formation_id'] as String?;
      final identity = formationId == null
          ? '-choice-$ordinal'
          : '-formation-$formationId';
      choices.add(
        OutlinedButton(
          key: ValueKey('$baseKey${repeated ? identity : ''}'),
          onPressed: enabled ? () => onSelected(intent) : null,
          child: Text(
            '${intent.type == 'offer' ? (intent.seed['terms']?['loan'] == true ? CgmsLocalizations.of(context).onlineLoanIntactFormation : CgmsLocalizations.of(context).tradeCards) : label(intent.type)}${_description(intent)}'
            '${repeated ? ' · combo $ordinal' : ''}',
          ),
        ),
      );
    }
    return Wrap(
      alignment: WrapAlignment.center,
      spacing: 8,
      runSpacing: 8,
      children: choices,
    );
  }

  String _baseKey(CardPlayIntent intent) =>
      'card-intent-${intent.type}${_suffix(intent)}';

  String _suffix(CardPlayIntent intent) {
    if (intent.type == 'offer')
      return intent.seed['terms']?['loan'] == true ? '-loan' : '-trade';
    final kind = (intent.seed['formation'] as Map?)?['kind'];
    return kind != null
        ? '-$kind'
        : intent.seed['suit'] != null && intent.type == 'open-series'
        ? '-${intent.seed['suit']}'
        : '';
  }

  String _description(CardPlayIntent intent) {
    final kind = (intent.seed['formation'] as Map?)?['kind'];
    if (kind != null) return ' · ${label(kind as String)}';
    if (intent.type == 'open-series')
      return ' · ${label(intent.seed['suit'] as String)}';
    return '';
  }
}
