import 'package:flutter/material.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'online_client.dart';

/// Private account economy. This view never grants rewards or starts games.
class EconomyPanel extends StatefulWidget {
  const EconomyPanel({super.key, required this.client});
  final OnlineClient client;
  @override
  State<EconomyPanel> createState() => _EconomyPanelState();
}

class _EconomyPanelState extends State<EconomyPanel> {
  int _days = 1;
  bool _actionFailed = false;
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) widget.client.refreshEconomy();
    });
  }

  String _utc(dynamic value) =>
      DateTime.tryParse(value?.toString() ?? '')
          ?.toUtc()
          .toIso8601String()
          .replaceFirst('T', ' ')
          .replaceFirst('.000Z', '')
          .replaceFirst('Z', '') ??
      '—';

  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: widget.client,
    builder: (context, _) {
      final c = widget.client, l = CgmsLocalizations.of(context);
      final e = c.economy;
      final enabled = e?['enabled'] == true;
      final b = enabled ? e!['benefits'] as Map : null;
      final cost = enabled ? (e!['pass_day_price'] as int) * _days : 0;
      final allowed =
          enabled &&
          !c.economyLoading &&
          c.economyError != 'ECONOMY_UNAVAILABLE' &&
          !c.busy &&
          c.pendingPass == null &&
          c.unknownOperation == null &&
          !c.hasPendingRoomOperation &&
          (b?['dirt'] as int? ?? 0) >= cost;
      final error = switch (c.economyError) {
        'INSUFFICIENT_DIRT' => l.economyInsufficient,
        'PRODUCT_UNAVAILABLE' => l.economyProductUnavailable,
        'OUTCOME_UNKNOWN' => l.economyUnknown,
        'RECOVERY_UNAVAILABLE' => l.economyRecoveryUnavailable,
        'ECONOMY_DISABLED' => l.economyDisabled,
        null => null,
        _ => l.economyUnavailable,
      };
      return Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(l.economyTitle, style: Theme.of(context).textTheme.titleLarge),
            if (c.economyLoading) const LinearProgressIndicator(),
            if (_actionFailed) Text(l.economyUnavailable),
            if (error != null) Semantics(liveRegion: true, child: Text(error)),
            if (e?['enabled'] == false) Text(l.economyDisabled),
            if (e?['enabled'] == true) ...[
              Text(
                l.economyDirt(b!['dirt'] as int),
                style: Theme.of(context).textTheme.titleMedium,
              ),
              Text(l.economyStarts(e!['free_starts_remaining'] as int)),
              Text(l.economyReset(_utc(e['resets_at']))),
              Text(l.economyReward(e['reward'] as int)),
              if (b['unlimited'] == true) Text(l.economyUnlimited),
              Text(b['no_ads'] == true ? l.economyNoAds : l.economyAdsActive),
              if (b['no_ads'] == true &&
                  b['ad_free_until'] != null &&
                  DateTime.tryParse(b['ad_free_until'].toString())?.isAfter(
                        DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
                      ) ==
                      true)
                Text(l.economyAdFreeUntil(_utc(b['ad_free_until']))),
              if (b['pass_until'] != null &&
                  DateTime.tryParse(b['pass_until'].toString())?.isAfter(
                        DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
                      ) ==
                      true)
                Text(l.economyPassUntil(_utc(b['pass_until']))),
              if (b['premium_until'] != null &&
                  DateTime.tryParse(b['premium_until'].toString())?.isAfter(
                        DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
                      ) ==
                      true)
                Text(l.economyPremiumUntil(_utc(b['premium_until']))),
              const SizedBox(height: 12),
              Text(l.economyPassHelp),
              const SizedBox(height: 8),
              DropdownButtonFormField<int>(
                initialValue: _days,
                isExpanded: true,
                decoration: InputDecoration(labelText: l.economyDays),
                items: [
                  for (final day in [1, 7])
                    DropdownMenuItem(
                      value: day,
                      child: Text(day == 1 ? l.economyDaily : l.economyWeekly),
                    ),
                ],
                onChanged: c.busy || c.pendingPass != null
                    ? null
                    : (value) => setState(() => _days = value!),
              ),
              const SizedBox(height: 8),
              FilledButton(
                key: const ValueKey('buy-pass'),
                onPressed: allowed
                    ? () async {
                        try {
                          await c.buyPass(_days);
                        } catch (_) {
                          if (mounted) setState(() => _actionFailed = true);
                        }
                      }
                    : null,
                child: Text(l.economySpend(cost)),
              ),
            ],
            Text(l.economyPaidUnavailable),
            const SizedBox(height: CgmsSpacing.controlGap),
            Text(l.economyAdFreeTerms),
            const SizedBox(height: CgmsSpacing.controlGap),
            Text(l.economyPaidDailyTerms),
            const SizedBox(height: CgmsSpacing.controlGap),
            Text(l.economyPaidLongerTerms),
            if (c.pendingPass != null) ...[
              if (c.economyError != 'OUTCOME_UNKNOWN') Text(l.economyUnknown),
              OutlinedButton(
                key: const ValueKey('recover-pass'),
                onPressed: c.busy ? null : c.recoverPass,
                child: Text(l.economyRecover),
              ),
            ],
            if (c.passReceipt != null)
              Semantics(
                liveRegion: true,
                child: Text(
                  l.economyConfirmed(
                    c.passReceipt!['cost'] as int,
                    _utc(c.passReceipt!['expires_at']),
                  ),
                ),
              ),
            TextButton(
              onPressed: c.economyLoading ? null : c.refreshEconomy,
              child: Text(l.economyRefresh),
            ),
          ],
        ),
      );
    },
  );
}
