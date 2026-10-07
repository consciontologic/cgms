import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';

import 'approved_demo_controller.dart';

/// Host-owned routes and actions; the reusable board receives projections only.
class ApprovedFlowView extends StatelessWidget {
  const ApprovedFlowView({
    super.key,
    required this.controller,
    required this.route,
    required this.tableRoute,
    required this.layoutMenu,
    required this.onNavigate,
    required this.onSettings,
    required this.showArt,
  });
  final ApprovedDemoController controller;
  final String route, tableRoute;
  final Widget layoutMenu;
  final ValueChanged<String> onNavigate;
  final VoidCallback onSettings;
  final bool showArt;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: route == '/setup' || route == '/scenarios'
        ? AppBar(
            toolbarHeight: MediaQuery.textScalerOf(context).scale(24) + 24,
            title: const Text(
              'CGMS',
              style: TextStyle(fontSize: 20, letterSpacing: 2),
            ),
            actions: [
              layoutMenu,
              IconButton(
                tooltip: 'Flow examples',
                onPressed: () => onNavigate('/scenarios'),
                icon: const Icon(Icons.grid_view),
              ),
              PopupMenuButton<String>(
                tooltip: 'Game menu',
                onSelected: (value) {
                  switch (value) {
                    case 'online':
                      onNavigate('/online');
                    case 'room':
                      controller.loadExample(1);
                      onNavigate('/setup');
                    case 'rules':
                      _rules(context);
                    case 'settings':
                      onSettings();
                  }
                },
                itemBuilder: (_) => const [
                  PopupMenuItem(value: 'online', child: Text('Online game')),
                  PopupMenuItem(value: 'room', child: Text('Room setup')),
                  PopupMenuItem(value: 'rules', child: Text('Game rules')),
                  PopupMenuItem(
                    value: 'settings',
                    child: Text('Display settings'),
                  ),
                ],
              ),
            ],
          )
        : null,
    body: SafeArea(
      child: route == '/scenarios'
          ? _examples(context)
          : route == '/setup'
          ? _room(context)
          : _frame(context, _board(context)),
    ),
  );

  Widget _frame(BuildContext context, Widget child) => LayoutBuilder(
    builder: (context, box) {
      // Keep one element path for Auto and all labelled review overrides.
      // Only overrides may create a horizontally scrollable preview canvas.
      final width = switch (tableRoute) {
        '/phone' => box.maxWidth.clamp(0.0, 440.0),
        '/tablet' => box.maxWidth.clamp(768.0, 1024.0),
        '/table' => box.maxWidth < 1050 ? 1050.0 : box.maxWidth,
        _ => box.maxWidth,
      };
      return Center(
        child: SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: SizedBox(width: width, height: box.maxHeight, child: child),
        ),
      );
    },
  );

  Widget _board(BuildContext context) {
    final data = controller.table;
    return CgmsAdaptiveTable(
      key: const ValueKey('approved-board'),
      data: data,
      showArt: showArt,
      selectedCardIds: controller.selectedIds,
      showDecisions: route == '/scores',
      publicGroupLabels: {
        for (final seat in data.seats)
          for (final suit in CardSuit.values)
            '${seat.id}:${suit.name}':
                '${suit.name[0].toUpperCase()}${suit.name.substring(1)} ${seat.exposed.where((c) => c.suit == suit && int.tryParse(c.rank) != null).fold(0, (sum, c) => sum + int.parse(c.rank))}',
      },
      onInspectCombo: (combo) => _inspectCombo(context, combo),
      onSelect: (id) {
        if (data.hand.any((card) => card.id == id)) {
          controller.selectHand(id);
        } else {
          controller.selectPublic(id);
        }
      },
      onInspect: (card) => _inspect(context, card),
      onCoup: controller.canCoup ? controller.declareCoup : null,
      header: Wrap(
        spacing: MediaQuery.sizeOf(context).width < 1050 ? 6 : 12,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          Text(
            'Round ${data.round} / ${data.totalRounds}',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          Text(
            controller.isYourTurn
                ? 'Your turn'
                : '${controller.activePlayer}’s turn',
            style: const TextStyle(
              color: CgmsColors.cyan,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
      headerTools: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          layoutMenu,
          PopupMenuButton<String>(
            tooltip: 'Game menu',
            onSelected: (value) {
              switch (value) {
                case 'examples':
                  onNavigate('/scenarios');
                case 'online':
                  onNavigate('/online');
                case 'room':
                  controller.loadExample(1);
                  onNavigate('/setup');
                case 'rules':
                  _rules(context);
                case 'settings':
                  onSettings();
              }
            },
            itemBuilder: (_) => const [
              PopupMenuItem(value: 'examples', child: Text('Flow examples')),
              PopupMenuItem(value: 'online', child: Text('Online game')),
              PopupMenuItem(value: 'room', child: Text('Room setup')),
              PopupMenuItem(value: 'rules', child: Text('Game rules')),
              PopupMenuItem(value: 'settings', child: Text('Display settings')),
            ],
          ),
        ],
      ),
      actions: route == '/scores'
          ? _scores(context, withReturn: true)
          : controller.screen == 3 &&
                data.pending == null &&
                controller.connected
          ? _actionChooser()
          : _actions(context),
      footer: CgmsGameActions(
        secondary: OutlinedButton(
          onPressed: controller.canTrade ? controller.beginTrade : null,
          child: const Text('Trade cards'),
        ),
        primary: FilledButton(
          onPressed: controller.canEndTurn ? controller.endTurn : null,
          child: const Text('End turn'),
        ),
        utilities: [
          IconButton(
            tooltip: 'History',
            onPressed: () => _history(context),
            icon: const Icon(Icons.history),
          ),
          IconButton(
            tooltip: 'Scores',
            onPressed: () => onNavigate('/scores'),
            icon: const Icon(Icons.receipt_long),
          ),
        ],
      ),
    );
  }

  Widget _actionChooser() => Align(
    alignment: AlignmentDirectional.centerStart,
    child: PopupMenuButton<String>(
      tooltip: 'Board actions',
      onSelected: (value) {
        switch (value) {
          case 'open':
            controller.beginOpening();
          case 'attack':
            controller.beginAttack();
          case 'trade':
            controller.beginTrade();
          case 'buy':
            controller.beginPurchase();
          case 'end':
            controller.endTurn();
        }
      },
      itemBuilder: (_) => [
        PopupMenuItem(
          value: 'open',
          enabled: controller.canOpen,
          child: const Text('Open Hearts'),
        ),
        PopupMenuItem(
          value: 'attack',
          enabled: controller.canAttack,
          child: const Text('Attack Bob’s Hearts'),
        ),
        PopupMenuItem(
          value: 'trade',
          enabled: controller.canTrade,
          child: const Text('Trade with Bob'),
        ),
        PopupMenuItem(
          value: 'buy',
          enabled: controller.canBuy,
          child: const Text('Buy cards'),
        ),
        PopupMenuItem(
          value: 'end',
          enabled: controller.canEndTurn,
          child: const Text('End turn'),
        ),
      ],
      child: const SizedBox(
        height: 48,
        child: Center(widthFactor: 1, child: Text('Actions ▾')),
      ),
    ),
  );

  Widget _actions(BuildContext context) {
    final c = controller;
    if (c.boardClosed) return _scores(context);
    if (c.screen == 6 && c.table.pending == null && true) {
      return Row(
        children: [
          Expanded(
            child: FilledButton(
              key: const ValueKey('approved-primary'),
              onPressed: c.primaryEnabled
                  ? () => c.primary(gameId: c.table.gameId)
                  : null,
              child: const Text('Hearts opened · choose attack'),
            ),
          ),
          IconButton(
            tooltip: 'Hearts opened details',
            onPressed: () => showDialog<void>(
              context: context,
              builder: (dialog) => AlertDialog(
                title: Text(c.title),
                content: Text(c.detail),
                actions: [
                  TextButton(
                    onPressed: () {
                      Navigator.pop(dialog);
                      c.backToBoard();
                    },
                    child: const Text('Back to board'),
                  ),
                  TextButton(
                    onPressed: () => Navigator.pop(dialog),
                    child: const Text('Close'),
                  ),
                ],
              ),
            ),
            icon: const Icon(Icons.info_outline),
          ),
        ],
      );
    }
    final compact =
        c.screen == 3 && c.table.pending == null && !c.loanOffered && true;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (!compact) ...[
            if (c.table.pending == null)
              Text(
                c.title,
                style: Theme.of(
                  context,
                ).textTheme.titleLarge?.copyWith(color: CgmsColors.cyan),
              ),
            const SizedBox(height: 4),
            if (c.screen != 20)
              Text(c.detail, style: const TextStyle(fontSize: 12)),
            const SizedBox(height: 6),
          ],
          if (c.screen == 13 && c.table.pending == null)
            Wrap(
              spacing: 8,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                IconButton(
                  tooltip: 'Buy fewer cards',
                  onPressed: c.purchaseQuantity > 1
                      ? () => c.setPurchaseQuantity(c.purchaseQuantity - 1)
                      : null,
                  icon: const Icon(Icons.remove),
                ),
                Text(
                  '${c.purchaseQuantity}',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                IconButton(
                  tooltip: 'Buy more cards',
                  onPressed: c.purchaseQuantity < 2
                      ? () => c.setPurchaseQuantity(c.purchaseQuantity + 1)
                      : null,
                  icon: const Icon(Icons.add),
                ),
                const Text(
                  'Select payment Spades in your hand.',
                  style: TextStyle(fontSize: 12),
                ),
              ],
            ),
          if (c.screen == 10 || c.screen == 11)
            Wrap(
              spacing: 12,
              children: [
                const Text(
                  'Give Q♣ → request 9♠',
                  style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                ),
                if (c.offerPending)
                  Text(
                    'Revision ${c.offerRevision} · private',
                    style: const TextStyle(fontSize: 12),
                  ),
              ],
            ),
          if (!c.connected)
            FilledButton.icon(
              key: const ValueKey('approved-reconnect'),
              onPressed: c.reconnect,
              icon: const Icon(Icons.sync),
              label: const Text('Reconnect'),
            )
          else
            FilledButton(
              key: const ValueKey('approved-primary'),
              onPressed: c.primaryEnabled
                  ? () => c.primary(gameId: c.table.gameId)
                  : null,
              child: Text(c.primaryLabel),
            ),
          if (c.canConfine)
            FilledButton.tonal(
              key: const ValueKey('approved-confinement'),
              onPressed: c.useConfinement,
              child: const Text('Use Confinement · A♦'),
            ),
          if (c.screen != 17 && c.canCoup)
            FilledButton(
              onPressed: c.declareCoup,
              child: const Text('Declare Coup now'),
            ),
          if (c.screen == 17 && c.table.pending == null)
            TextButton(
              onPressed: () => c.beginPonzi(coupReady: true),
              child: const Text('Try Coup during a response'),
            ),
          if (c.offerPending)
            Wrap(
              spacing: 8,
              children: [
                TextButton(
                  onPressed: c.withdrawOffer,
                  child: const Text('Withdraw offer'),
                ),
                TextButton(
                  onPressed: c.declineOffer,
                  child: const Text('Simulate Bob declining'),
                ),
              ],
            ),
          if (c.screen == 4)
            TextButton(
              onPressed: () =>
                  _inspectCombo(context, c.table.seats[2].combos.single),
              child: const Text('Inspect five kings'),
            ),
          if (c.screen == 15)
            TextButton(
              onPressed: c.beginLoan,
              child: const Text('Replay loan flow'),
            ),
          if (c.screen == 16)
            TextButton(
              onPressed: c.beginPonzi,
              child: const Text('Replay Confinement response'),
            ),
          if (c.screen == 20)
            Text(c.detail, style: const TextStyle(fontSize: 12)),
          if (c.table.pending == null &&
              !c.boardClosed &&
              c.screen != 3 &&
              c.screen != 20 &&
              c.screen != 17)
            TextButton(
              onPressed: c.backToBoard,
              child: const Text('Back to board'),
            ),
        ],
      ),
    );
  }

  Widget _scores(BuildContext context, {bool withReturn = false}) {
    final c = controller;
    final seats = [...c.table.seats];
    if (c.finalized)
      seats.sort(
        (a, b) => b.finances.matchScore.compareTo(a.finances.matchScore),
      );
    Widget cell(String text, {int flex = 1, bool bold = false}) => Expanded(
      flex: flex,
      child: Padding(
        padding: const EdgeInsets.all(3),
        child: Text(
          text,
          style: TextStyle(
            fontSize: 12,
            fontWeight: bold ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          c.boardClosed ? c.title : 'Scores and obligations',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        if (c.boardClosed) Text(c.detail, style: const TextStyle(fontSize: 12)),
        Row(
          children: [
            cell('Player', flex: 2, bold: true),
            cell(c.finalized ? 'Match' : 'Game', bold: true),
            cell('Cash', bold: true),
            cell('Debt', bold: true),
          ],
        ),
        for (final seat in seats)
          Container(
            margin: const EdgeInsets.only(bottom: 4),
            decoration: BoxDecoration(
              color: CgmsColors.surface,
              borderRadius: BorderRadius.circular(6),
            ),
            child: Row(
              children: [
                cell(seat.name, flex: 2, bold: true),
                cell(
                  (c.finalized
                          ? seat.finances.matchScore
                          : seat.finances.gameScore)
                      .format(),
                  bold: true,
                ),
                cell(seat.finances.available.format()),
                cell(seat.finances.debt.format()),
              ],
            ),
          ),
        if (c.finalized)
          const Text(
            'Completed game scores remain recorded. New-game cash starts at zero.',
            style: TextStyle(fontSize: 12),
          ),
        if (c.boardClosed)
          FilledButton(
            key: const ValueKey('approved-primary'),
            onPressed: c.primaryEnabled
                ? () => c.primary(gameId: c.table.gameId)
                : null,
            child: Text(c.primaryLabel),
          ),
        if (withReturn)
          TextButton(
            onPressed: () => onNavigate(tableRoute),
            child: const Text('Back to board'),
          ),
      ],
    );
  }

  Widget _examples(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.all(16),
    child: Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 900),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              '20 game moments',
              style: Theme.of(context).textTheme.displaySmall,
            ),
            const Text(
              'Choose a prepared example. Each choice starts a fresh local branch; actions continue within that branch.',
            ),
            const SizedBox(height: 16),
            for (var i = 0; i < ApprovedDemoController.examples.length; i++)
              Padding(
                padding: const EdgeInsets.only(bottom: 6),
                child: OutlinedButton(
                  key: ValueKey('approved-example-${i + 1}'),
                  onPressed: () {
                    controller.loadExample(i + 1);
                    onNavigate(i == 0 ? '/setup' : tableRoute);
                  },
                  style: OutlinedButton.styleFrom(
                    alignment: Alignment.centerLeft,
                    padding: const EdgeInsets.all(16),
                  ),
                  child: Text(
                    '${(i + 1).toString().padLeft(2, '0')}  ${ApprovedDemoController.examples[i]}',
                  ),
                ),
              ),
            TextButton(
              onPressed: () => onNavigate(tableRoute),
              child: const Text('Return to current board'),
            ),
          ],
        ),
      ),
    ),
  );

  Widget _room(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.all(20),
    child: Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 650),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text('Your room', style: Theme.of(context).textTheme.displayMedium),
            const Text('4 players · local demonstration'),
            const SizedBox(height: 20),
            Wrap(
              spacing: 8,
              alignment: WrapAlignment.center,
              children: [
                for (final card in const [
                  VisibleCard(
                    id: 'room-club',
                    rank: '8',
                    suit: CardSuit.clubs,
                    copy: 1,
                  ),
                  VisibleCard(
                    id: 'room-queen',
                    rank: 'Q',
                    suit: CardSuit.hearts,
                    copy: 1,
                  ),
                  VisibleCard(
                    id: 'room-diamond',
                    rank: '4',
                    suit: CardSuit.diamonds,
                    copy: 1,
                  ),
                ])
                  CgmsCard(card: card, width: 80, showArt: showArt),
              ],
            ),
            const SizedBox(height: 20),
            TextFormField(
              initialValue: controller.playerName,
              decoration: const InputDecoration(labelText: 'Your name'),
              onChanged: controller.setName,
            ),
            for (final name in [controller.playerName, 'Bob', 'Carol', 'Deniz'])
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: Text(name),
                trailing: const Text(
                  'Ready',
                  style: TextStyle(color: CgmsColors.cyan),
                ),
              ),
            Text(
              'Games in match',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            Wrap(
              spacing: 12,
              children: [
                for (final count in [3, 5, 7])
                  ChoiceChip(
                    label: Text('$count games'),
                    selected: controller.games == count,
                    onSelected: (_) => controller.setGames(count),
                  ),
              ],
            ),
            const SizedBox(height: 16),
            OutlinedButton(
              onPressed: () => _rules(context),
              child: const Text('Game rules'),
            ),
            FilledButton(
              key: const ValueKey('approved-start-match'),
              onPressed: () {
                controller.startMatch();
                onNavigate(tableRoute);
              },
              child: const Text('Start match'),
            ),
            TextButton(
              onPressed: () => onNavigate(tableRoute),
              child: const Text('Back to board'),
            ),
          ],
        ),
      ),
    ),
  );

  void _inspect(BuildContext context, VisibleCard card) => showDialog<void>(
    context: context,
    builder: (_) => CgmsCardInspector(card: card, showArt: showArt),
  );
  void _inspectCombo(BuildContext context, ComboView combo) => showDialog<void>(
    context: context,
    builder: (_) => CgmsComboInspector(combo: combo, showArt: showArt),
  );
  void _history(BuildContext context) => showDialog<void>(
    context: context,
    builder: (dialog) => AlertDialog(
      title: const Text('Game history'),
      content: SizedBox(
        width: 500,
        child: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              for (final event in controller.table.events)
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: Text(event),
                ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(dialog),
          child: const Text('Close'),
        ),
      ],
    ),
  );
  void _rules(BuildContext context) => showDialog<void>(
    context: context,
    builder: (dialog) => AlertDialog(
      title: const Text('At this table'),
      content: const SingleChildScrollView(
        child: Text(
          'Open all number cards of a new suit together. Each turn allows one new series or combo opening, unless Underground grants its combo exception.\n\nAttacks require two currently open series. Club attacks match the target exactly, starting with Hearts; physical attacking Clubs can be used once per turn.\n\nOffers reserve no cards. Acceptance starts responses, and transfers happen only on success.\n\nCoup needs both Queens of Diamonds and both Queens of Hearts, a history of opening Clubs, and no alliance. It ends the board immediately.\n\nBoard closure precedes settlement and financial finalization. Reconnecting never passes or repeats an action.',
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(dialog),
          child: const Text('Close'),
        ),
      ],
    ),
  );
}
