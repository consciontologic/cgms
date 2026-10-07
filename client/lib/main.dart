import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

import 'keyboard_actions.dart';

import 'demo_controller.dart';
import 'approved_demo_controller.dart';
import 'approved_flow.dart';
import 'online/online_client.dart';
import 'online/browser_transport.dart';
import 'online/online_screen.dart';

const _tableModes = {
  '/play': 'Auto · screen size',
  '/phone': 'Phone',
  '/tablet': 'Tablet',
  '/table': 'Desktop',
};

bool _isTableRoute(String route) =>
    _tableModes.containsKey(route) || route == '/touch';

String _tableKind(String route, double width) {
  if (route == '/table') return 'desktop';
  if (route == '/phone') return 'phone';
  if (route == '/tablet') return 'tablet';
  if (route != '/touch' && width >= 1050) return 'desktop';
  return width >= 600 ? 'tablet' : 'phone';
}

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  // Keep the accessible DOM available to keyboard and assistive technology.
  SemanticsBinding.instance.ensureSemantics();
  runApp(const DemoApp());
}

class DemoApp extends StatefulWidget {
  const DemoApp({
    super.key,
    this.initialTextScale = 1,
    this.legacy = false,
    this.onlineClient,
  });
  final double initialTextScale;
  final bool legacy;

  /// Optional caller-owned client for embedding and route integration tests.
  final OnlineClient? onlineClient;
  @override
  State<DemoApp> createState() => _DemoAppState();
}

class _DemoAppState extends State<DemoApp> {
  final _controller = DemoController();
  final _approved = ApprovedDemoController();
  late final _online =
      widget.onlineClient ??
      OnlineClient(createOnlineTransport(Uri.parse(Uri.base.origin)));
  late bool _approvedMode;
  final _previewScroll = ScrollController();
  late final _DemoRouter _router;
  late double _textScale;
  bool _showArt = true;
  late String _localRoute;
  bool _onlineOpened = false;
  String? _onlineMatch;
  String? _onlineRoom;
  bool _compactTable = true;
  bool _reduceMotion = false;
  bool get _renderLocalArt =>
      _showArt && Uri.parse(_router.currentConfiguration).path != '/online';
  @override
  void initState() {
    super.initState();
    _textScale = widget.initialTextScale;
    _approvedMode = !widget.legacy;
    _localRoute = widget.legacy ? '/setup' : '/play';
    _router = _DemoRouter(
      builder: _shell,
      onNavigate: (route) {
        if (_approvedMode && route == '/setup' && _approved.screen != 1) {
          _approved.loadExample(1);
        }
      },
    );
    if (_approvedMode) _router.go('/play');
    _controller.addListener(_router.refresh);
    _approved.addListener(_router.refresh);
  }

  @override
  void dispose() {
    _controller.removeListener(_router.refresh);
    _controller.dispose();
    _approved.removeListener(_router.refresh);
    _approved.dispose();
    _previewScroll.dispose();
    if (_onlineOpened && widget.onlineClient == null) _online.dispose();
    _router.dispose();
    super.dispose();
  }

  void _go(BuildContext context, String route) {
    Router.navigate(context, () => _router.go(route));
  }

  @override
  Widget build(BuildContext context) => MaterialApp.router(
    title: 'CGMS · cgmsart',
    debugShowCheckedModeBanner: false,
    theme: CgmsTheme.dark(),
    localizationsDelegates: CgmsLocalizations.localizationsDelegates,
    supportedLocales: CgmsLocalizations.supportedLocales,
    actions: kIsWeb
        ? {
            ...WidgetsApp.defaultActions,
            NextFocusIntent: BrowserNextFocusAction(),
            PreviousFocusIntent: BrowserPreviousFocusAction(),
          }
        : null,
    routerDelegate: _router,
    routeInformationParser: _DemoRouteParser(
      defaultRoute: widget.legacy ? '/setup' : '/play',
    ),
    scrollBehavior: const MaterialScrollBehavior().copyWith(
      dragDevices: {
        PointerDeviceKind.touch,
        PointerDeviceKind.mouse,
        PointerDeviceKind.trackpad,
        PointerDeviceKind.stylus,
      },
    ),
    builder: (context, child) {
      final media = MediaQuery.of(context);
      return MediaQuery(
        data: media.copyWith(
          textScaler: _textScale == 1
              ? media.textScaler
              : TextScaler.linear(_textScale),
          disableAnimations: _reduceMotion || media.disableAnimations,
        ),
        // OverlayPortal tooltips otherwise turn the route's merged semantics
        // into a new ancestor on hover. Keep that boundary present from the
        // first frame so native accessibility scroll offsets remain attached.
        child: Semantics(
          container: true,
          explicitChildNodes: true,
          child: child!,
        ),
      );
    },
  );

  Widget _shell(BuildContext context, String route) {
    final uri = Uri.parse(route);
    final online = uri.path == '/online';
    if (online) {
      _onlineOpened = true;
      _onlineMatch = uri.queryParameters['match'];
      _onlineRoom = uri.queryParameters['room'];
    } else {
      _localRoute = route;
    }
    return IndexedStack(
      index: online ? 1 : 0,
      children: [
        // Keep the local subtree and its route/drafts mounted. Its artwork is
        // disabled while hidden; the user's artwork preference is unchanged.
        _localShell(context, _localRoute),
        if (_onlineOpened)
          OnlineScreen(
            client: _online,
            matchId: _onlineMatch,
            roomId: _onlineRoom,
            onMatchChanged: (id) => _go(
              context,
              Uri(path: '/online', queryParameters: {'match': id}).toString(),
            ),
            onRoomChanged: (id) => _go(
              context,
              Uri(path: '/online', queryParameters: {'room': id}).toString(),
            ),
          ),
      ],
    );
  }

  Widget _localShell(BuildContext context, String route) {
    if (_approvedMode && route != '/gallery') {
      return ApprovedFlowView(
        controller: _approved,
        route: route,
        tableRoute: _router.tableRoute,
        layoutMenu: _layoutMenu(context),
        showArt: _renderLocalArt,
        onNavigate: (path) => _go(context, path),
        onSettings: () => _settings(context),
      );
    }
    if (!_isTableRoute(route)) return _standardShell(context, route);
    final kind = _tableKind(route, MediaQuery.sizeOf(context).width);
    return KeyedSubtree(
      key: ValueKey('demo-layout-$kind'),
      child: kind == 'desktop'
          ? _standardShell(context, '/table')
          : _touchShell(context, route),
    );
  }

  Widget _layoutMenu(BuildContext context) {
    final route = _router.tableRoute;
    final width = MediaQuery.sizeOf(context).width;
    final kind = _tableKind(route, width);
    final compact = !widget.legacy && width < 1050;
    final label = switch (route) {
      '/play' => 'Auto',
      '/touch' => 'Touch',
      _ => _tableModes[route]!,
    };
    return PopupMenuButton<String>(
      tooltip: 'Table layout',
      onSelected: (route) => _go(context, route),
      itemBuilder: (_) => [
        for (final mode in _tableModes.entries)
          CheckedPopupMenuItem(
            value: mode.key,
            checked: route == mode.key,
            child: Text(mode.value),
          ),
      ],
      child: ConstrainedBox(
        constraints: const BoxConstraints(minWidth: 48, minHeight: 48),
        child: Padding(
          padding: EdgeInsets.symmetric(
            horizontal: compact ? 2 : 12,
            vertical: 12,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (!compact) ...[
                Icon(switch (kind) {
                  'phone' => Icons.phone_android,
                  'tablet' => Icons.tablet_mac,
                  _ => Icons.desktop_windows,
                }, size: 18),
                const SizedBox(width: 8),
              ],
              Text(label),
              Icon(Icons.arrow_drop_down, size: compact ? 16 : 20),
            ],
          ),
        ),
      ),
    );
  }

  Widget _touchFrame(String route, Widget child) => LayoutBuilder(
    builder: (context, constraints) {
      final width = switch (route) {
        '/phone' => constraints.maxWidth.clamp(0.0, 440.0),
        '/tablet' => constraints.maxWidth.clamp(768.0, 1024.0),
        _ => constraints.maxWidth.clamp(0.0, 1024.0),
      };
      if (width <= constraints.maxWidth) {
        return Center(
          child: SizedBox(
            width: width,
            height: constraints.maxHeight,
            child: child,
          ),
        );
      }
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Padding(
            padding: EdgeInsets.fromLTRB(16, 8, 16, 4),
            child: Text(
              'Tablet preview · scroll sideways, or choose Auto to fit this screen.',
              style: TextStyle(color: CgmsColors.muted, fontSize: 12),
            ),
          ),
          Expanded(
            child: LayoutBuilder(
              builder: (context, frame) => Scrollbar(
                controller: _previewScroll,
                thumbVisibility: true,
                child: SingleChildScrollView(
                  key: const ValueKey('tablet-preview-scroll'),
                  controller: _previewScroll,
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.only(bottom: 12),
                  child: SizedBox(
                    width: width,
                    height: (frame.maxHeight - 12).clamp(0.0, double.infinity),
                    child: child,
                  ),
                ),
              ),
            ),
          ),
        ],
      );
    },
  );

  Widget _standardShell(BuildContext context, String route) => Scaffold(
    body: SafeArea(
      child: Column(
        children: [
          Container(
            decoration: const BoxDecoration(
              color: CgmsColors.canvas,
              border: Border(bottom: BorderSide(color: CgmsColors.raised)),
            ),
            child: Center(
              child: ConstrainedBox(
                constraints: BoxConstraints(
                  maxWidth: route == '/table' ? 1440 : 1248,
                ),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 12,
                  ),
                  child: Wrap(
                    alignment: WrapAlignment.spaceBetween,
                    crossAxisAlignment: WrapCrossAlignment.center,
                    spacing: 24,
                    runSpacing: 8,
                    children: [
                      TextButton(
                        onPressed: () => _go(context, '/setup'),
                        child: const Text(
                          'CGMS / Cgmsart',
                          style: TextStyle(
                            fontFamily: 'packages/cgms_ui/Barlow Condensed',
                            fontSize: 23,
                            letterSpacing: 3,
                            color: CgmsColors.text,
                          ),
                        ),
                      ),
                      Wrap(
                        spacing: 4,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        children: [
                          for (final entry in {
                            _router.tableRoute: 'Table',
                            '/trade': 'Trade',
                            '/scores': 'Scores',
                            '/scenarios': 'Scenarios',
                            '/gallery': 'Gallery',
                          }.entries)
                            TextButton(
                              style: TextButton.styleFrom(
                                foregroundColor:
                                    route == entry.key ||
                                        (_isTableRoute(route) &&
                                            entry.key == _router.tableRoute)
                                    ? CgmsColors.cyan
                                    : CgmsColors.muted,
                                backgroundColor:
                                    route == entry.key ||
                                        (_isTableRoute(route) &&
                                            entry.key == _router.tableRoute)
                                    ? CgmsColors.surface
                                    : null,
                              ),
                              onPressed: () => _go(context, entry.key),
                              child: Text(entry.value),
                            ),
                          _layoutMenu(context),
                          IconButton(
                            tooltip: 'Display settings',
                            onPressed: () => _settings(context),
                            icon: const Icon(Icons.tune),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          Expanded(
            child: SingleChildScrollView(
              key: ValueKey(route),
              padding: EdgeInsets.fromLTRB(
                20,
                route == '/table' && _compactTable ? 12 : 32,
                20,
                48,
              ),
              child: Center(
                child: ConstrainedBox(
                  constraints: BoxConstraints(
                    maxWidth: route == '/table' && _compactTable ? 1400 : 1200,
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _content(context, route),
                      const SizedBox(height: 40),
                      const Divider(),
                      const SizedBox(height: 16),
                      Wrap(
                        spacing: 20,
                        runSpacing: 8,
                        alignment: WrapAlignment.spaceBetween,
                        children: [
                          const Text(
                            'LOCAL DEMO  /  THREE SIMULATED SEATS',
                            style: TextStyle(
                              color: CgmsColors.muted,
                              fontSize: 12,
                              letterSpacing: 1,
                            ),
                          ),
                          TextButton.icon(
                            onPressed: () {
                              _controller.reset();
                              _go(context, '/setup');
                            },
                            icon: const Icon(Icons.restart_alt, size: 18),
                            label: const Text('Reset demo'),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    ),
  );

  Widget _touchShell(BuildContext context, String route) => Scaffold(
    appBar: AppBar(
      toolbarHeight: MediaQuery.textScalerOf(context).scale(56).clamp(56, 84),
      title: const Text(
        'Cgmsart',
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          fontFamily: 'packages/cgms_ui/Barlow Condensed',
          fontSize: 22,
          letterSpacing: 2,
        ),
      ),
      actions: [
        _layoutMenu(context),
        PopupMenuButton<String>(
          tooltip: 'Demo menu',
          onSelected: (destination) {
            if (destination == 'settings') {
              _settings(context);
            } else if (destination == 'reset') {
              _controller.reset();
              _go(context, '/setup');
            } else {
              _go(context, destination);
            }
          },
          itemBuilder: (_) => [
            for (final entry in const {
              '/table': 'Desktop table',
              '/scenarios': 'Scenarios',
              '/setup': 'Match setup',
              '/trade': 'Trade',
              '/scores': 'Scores',
              '/gallery': 'Gallery',
              'settings': 'Display settings',
              'reset': 'Reset demo',
            }.entries)
              PopupMenuItem(value: entry.key, child: Text(entry.value)),
          ],
        ),
      ],
    ),
    body: SafeArea(
      child: _touchFrame(
        route,
        _controller.boardClosed
            ? SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: _scores(context),
              )
            : CgmsTouchTableView(
                key: ValueKey(_controller.table.gameId),
                data: _controller.table,
                showArt: _renderLocalArt,
                onSelect: (id) => _selectPublic(context, id),
                onSelectHand: _controller.densityReview
                    ? _controller.selectCard
                    : null,
                onClearHandSelection: _controller.densityReview
                    ? _controller.clearSelection
                    : null,
                onInspect: (card) => _inspect(context, card),
                onInspectCombo: (combo) => showDialog<void>(
                  context: context,
                  builder: (_) => CgmsComboInspector(
                    combo: combo,
                    showArt: _renderLocalArt,
                  ),
                ),
                onPrimary: () => _performPrimary(context),
                onTrade: () => _go(context, '/trade'),
                onScores: () => _go(context, '/scores'),
                onEndTurn: () => _confirmEnd(context),
                endTurnEnabled: _controller.canEndTurn,
                endTurnDisabledReason: _controller.densityReview
                    ? 'This review scripts adding one card to an open series. Choose The last move for scoring, trades and the full guided ending.'
                    : null,
                onReconnect: () =>
                    _controller.setConnection(ConnectionStateView.connected),
              ),
      ),
    ),
  );

  void _selectPublic(BuildContext context, String id) {
    if (_controller.densityReview) return;
    final before = _controller.table.selectedCardId;
    _controller.selectCard(id);
    if (_controller.table.selectedCardId == before) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'This card cannot perform the next attack. Select the unused Club described above, or inspect a card for details.',
          ),
        ),
      );
    }
  }

  void _performPrimary(BuildContext context) {
    _controller.primaryAction();
    if (_controller.tradeOffered) _go(context, '/trade');
    if (_controller.boardClosed) _go(context, '/scores');
  }

  Widget _content(BuildContext context, String route) {
    final table = _controller.table;
    switch (route) {
      case '/setup':
        return CgmsSetupView(
          initialName: 'You',
          startLabel: 'Start guided game',
          noticeTitle: 'A guided final round',
          noticeMessage:
              'Start in a prepared round 13 position. Try two attacks, exchange a card, then see how scores and debt settle. The other three seats are simulated.',
          footer: 'Local demonstration · resets on refresh',
          onStart: (name, games) {
            _controller.configure(name: name, games: games);
            _go(context, _router.tableRoute);
          },
        );
      case '/table':
        if (_controller.boardClosed) return _scores(context);
        return CgmsTableView(
          key: ValueKey(table.gameId),
          data: table,
          compact: _compactTable,
          showArt: _renderLocalArt,
          onSelect: (id) => _selectPublic(context, id),
          onSelectHand: _controller.densityReview
              ? _controller.selectCard
              : null,
          onClearHandSelection: _controller.densityReview
              ? _controller.clearSelection
              : null,
          onInspect: (card) => _inspect(context, card),
          onInspectCombo: (combo) => showDialog<void>(
            context: context,
            builder: (_) =>
                CgmsComboInspector(combo: combo, showArt: _renderLocalArt),
          ),
          onPrimary: () => _performPrimary(context),
          onTrade: () => _go(context, '/trade'),
          onScores: () => _go(context, '/scores'),
          onEndTurn: () => _confirmEnd(context),
          endTurnEnabled: _controller.canEndTurn,
          endTurnExplanation: _controller.densityReview
              ? 'This review scripts adding one card to an open series. Choose The last move for scoring, trades and the full guided ending.'
              : null,
          onReconnect: () =>
              _controller.setConnection(ConnectionStateView.connected),
        );
      case '/trade':
        if (_controller.densityReview) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const CgmsNotice(
                title: 'Trade unavailable',
                message:
                    'This large-hand review demonstrates browsing and adding a card to an open series. Choose The last move or An offer on the table to try the separate trade scenario.',
              ),
              const SizedBox(height: 20),
              OutlinedButton(
                onPressed: () => _go(context, _router.tableRoute),
                child: const Text('Back to the table'),
              ),
            ],
          );
        }
        final enabled =
            table.connection == ConnectionStateView.connected &&
            table.pending == null &&
            table.seats.isNotEmpty &&
            !_controller.boardClosed;
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            CgmsTradeView(
              give: const VisibleCard(
                id: 'deck-1-hearts-5',
                rank: '5',
                suit: CardSuit.hearts,
                copy: 1,
              ),
              receive: const VisibleCard(
                id: 'deck-1-hearts-6',
                rank: '6',
                suit: CardSuit.hearts,
                copy: 1,
              ),
              status: _controller.tradeStatus,
              acceptLabel: 'Simulate Ivo accepting',
              declineLabel: 'Simulate Ivo declining',
              description:
                  'Exchange your 5 of Hearts for Ivo’s 6 of Hearts. This voluntary exchange changes neither scores nor alliances.',
              showArt: _renderLocalArt,
              onOffer:
                  enabled &&
                      !_controller.tradeOffered &&
                      !_controller.tradeResolved
                  ? _controller.offerTrade
                  : null,
              onAccept: enabled && _controller.tradeOffered
                  ? () {
                      _controller.acceptTrade();
                      _go(context, _router.tableRoute);
                    }
                  : null,
              onDecline: enabled && _controller.tradeOffered
                  ? _controller.declineTrade
                  : null,
              onWithdraw: enabled && _controller.tradeOffered
                  ? _controller.withdrawTrade
                  : null,
            ),
            if (!enabled)
              Padding(
                padding: const EdgeInsets.only(top: 20),
                child: CgmsNotice(
                  title: 'Trade unavailable',
                  message: _controller.boardClosed
                      ? 'The board is closed. Trading is finished for this game.'
                      : 'Load a populated scenario, restore the connection and finish any response window before transferring cards.',
                ),
              ),
            const SizedBox(height: 20),
            OutlinedButton.icon(
              onPressed: () => _go(context, _router.tableRoute),
              icon: const Icon(Icons.arrow_back),
              label: const Text('Back to the table'),
            ),
          ],
        );
      case '/scores':
        return _scores(context);
      case '/scenarios':
        return _scenarios(context);
      case '/gallery':
        return _gallery(context);
      default:
        return const CgmsNotice(
          title: 'This page is not at the table',
          message:
              'Use Table, Scenarios or the CGMS wordmark to return. This address does not identify a game.',
        );
    }
  }

  void _inspect(BuildContext context, VisibleCard card) => showDialog<void>(
    context: context,
    builder: (_) => CgmsCardInspector(card: card, showArt: _renderLocalArt),
  );

  Widget _scores(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      CgmsScoresView(
        seats: _controller.table.seats,
        title: _controller.finalized
            ? (_controller.games == 1
                  ? 'THE FINAL RECKONING.'
                  : 'GAME 1 RECORDED.')
            : _controller.boardClosed
            ? 'SETTLE THE TABLE.'
            : 'SCORES & OBLIGATIONS.',
        description: _controller.densityReview
            ? 'These prepared round 8 snapshots use zero financial balances to isolate card browsing. Adding the demonstrated Club does not award points. Round-end scoring is outside this review.'
            : _controller.boardClosed
            ? 'Round 13 ended normally. There is no victory declarer or threshold bonus. Each number Diamond scores its printed value + 5; the Queen of Spades adds 10. Match ranking follows recorded scores.'
            : 'Your earlier Code penalty recorded −5 and left 5 owed to the system. Breaking that formation did not erase the debt. A successful Club elimination earns 3 before automatic debt repayment.',
        finalized: _controller.finalized,
        onFinish: _controller.boardClosed && !_controller.finalized
            ? _controller.finishSettlement
            : null,
      ),
      if (_controller.finalized && _controller.games > 1)
        const Padding(
          padding: EdgeInsets.only(top: 16),
          child: CgmsNotice(
            title: 'First game complete',
            message:
                'This match was configured for three games. The demo scripts game 1 only; later games require the full game engine.',
          ),
        ),
      if (_controller.boardClosed)
        Padding(
          padding: const EdgeInsets.only(top: 20),
          child: Text(
            _ranking(),
            style: Theme.of(context).textTheme.titleLarge,
          ),
        ),
      if (_controller.boardClosed && !_controller.finalized)
        const Padding(
          padding: EdgeInsets.only(top: 12),
          child: Text(
            'Finish confirms your settlement and explicitly simulates the other three seats finishing. No promises are outstanding in this fixture.',
          ),
        ),
      const SizedBox(height: 20),
      OutlinedButton(
        onPressed: () => _go(
          context,
          _controller.boardClosed ? '/scenarios' : _router.tableRoute,
        ),
        child: Text(
          _controller.boardClosed
              ? 'Explore another scenario'
              : 'Back to the table',
        ),
      ),
    ],
  );

  String _ranking() {
    final sorted = [..._controller.table.seats]
      ..sort((a, b) => b.finances.matchScore.compareTo(a.finances.matchScore));
    var rank = 1;
    final lines = <String>['Recorded match ranking'];
    for (var i = 0; i < sorted.length; i++) {
      if (i > 0 &&
          sorted[i].finances.matchScore != sorted[i - 1].finances.matchScore)
        rank = i + 1;
      lines.add(
        '$rank. ${sorted[i].name} · ${sorted[i].finances.matchScore.format()} points',
      );
    }
    return lines.join('\n');
  }

  Future<void> _confirmEnd(BuildContext context) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('End the final turn?'),
        content: const Text(
          'You are last in round 13. Ending this turn closes the board and scores everyone’s remaining cards. Any unaccepted offer expires. Settlement comes next.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Keep playing'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('End turn'),
          ),
        ],
      ),
    );
    if (confirmed == true && context.mounted) {
      _controller.endTurn();
      if (_controller.boardClosed) _go(context, '/scores');
    }
  }

  Widget _scenarios(BuildContext context) {
    const choices = [
      (
        'combos',
        '11',
        'Combos at the table',
        'Great People protects Mara’s Diamonds, Ivo’s Fate waits for his turn, and a possible Kidnapper pair stays private in your stacked hand.',
      ),
      (
        'ready',
        '01',
        'The last move',
        'Begin with a legal exact-match attack and follow the whole guided game.',
      ),
      (
        'pending',
        '02',
        'An answer is owed',
        'A declared attack waits for the next seat. Responses have no countdown.',
      ),
      (
        'disconnected',
        '03',
        'A quiet connection',
        'A pending response survives disconnection. Restore it without an automatic pass.',
      ),
      (
        'loading',
        '04',
        'Taking a seat',
        'Loading state with readable last-known information and paused actions.',
      ),
      (
        'error',
        '05',
        'A recoverable error',
        'An explicit retry restores the deterministic local table.',
      ),
      (
        'trade',
        '06',
        'An offer on the table',
        'Review a private proposal, decline or withdraw it, or simulate acceptance.',
      ),
      (
        'results',
        '07',
        'The final accounts',
        'Review the ordinary game ending before explicitly finalizing settlement.',
      ),
      (
        'empty',
        '08',
        'Nothing up your sleeve',
        'An empty adapter projection with no cards or available actions.',
      ),
      (
        'hand-10',
        '09',
        'Ten cards in hand',
        'Twelve exposed cards per player. Browse ten private cards by suit, inspect physical copies, then add the 7 of Clubs to your series.',
      ),
      (
        'hand-15',
        '10',
        'Fifteen cards in hand',
        'The fullest review table: twelve exposed cards per player and fifteen in your hand. Filter, select and follow one card from hand to table.',
      ),
    ];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        FilledButton(
          onPressed: () {
            _approvedMode = true;
            _go(context, '/scenarios');
          },
          child: const Text('Approved 20 mobile flows'),
        ),
        Text(
          'CHOOSE A MOMENT.',
          style: Theme.of(context).textTheme.displayMedium,
        ),
        const SizedBox(height: 12),
        const Text(
          'Each scenario resets local progress. These are controlled review states, not live multiplayer rooms.',
        ),
        const SizedBox(height: 28),
        for (final (id, number, title, description) in choices)
          Padding(
            padding: const EdgeInsets.only(bottom: 12),
            child: OutlinedButton(
              onPressed: () {
                _controller.loadScenario(id);
                _go(context, _router.tableRoute);
              },
              style: OutlinedButton.styleFrom(
                padding: const EdgeInsets.all(24),
                alignment: Alignment.centerLeft,
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    number,
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      color: CgmsColors.cyan,
                    ),
                  ),
                  const SizedBox(width: 24),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          title,
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        const SizedBox(height: 6),
                        Text(
                          description,
                          style: const TextStyle(color: CgmsColors.muted),
                        ),
                      ],
                    ),
                  ),
                  const Icon(Icons.arrow_forward),
                ],
              ),
            ),
          ),
      ],
    );
  }

  Widget _gallery(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(
        'THE VISUAL LANGUAGE.',
        style: Theme.of(context).textTheme.displayMedium,
      ),
      const SizedBox(height: 12),
      const Text(
        'Live components, distinct physical copies and readable states. Use Display settings to hide artwork, enlarge text or reduce motion.',
      ),
      const SizedBox(height: 28),
      const CgmsSectionTitle(
        title: 'Cards with a point of view',
        eyebrow: 'ILLUSTRATION + LIVE LABELS',
      ),
      const SizedBox(height: 16),
      Wrap(
        spacing: 20,
        runSpacing: 20,
        children: [
          for (final card in const [
            VisibleCard(
              id: 'gallery-c7-1',
              rank: '7',
              suit: CardSuit.clubs,
              copy: 1,
              status: 'Selected',
            ),
            VisibleCard(
              id: 'gallery-c7-2',
              rank: '7',
              suit: CardSuit.clubs,
              copy: 2,
              status: 'Used this turn',
            ),
            VisibleCard(
              id: 'gallery-qh',
              rank: 'Q',
              suit: CardSuit.hearts,
              copy: 1,
              status: 'In your hand',
            ),
            VisibleCard(
              id: 'gallery-d8',
              rank: '8',
              suit: CardSuit.diamonds,
              copy: 1,
              status: 'Public',
            ),
          ])
            CgmsCard(
              card: card,
              width: 112,
              selected: card.id == 'gallery-c7-1',
              showArt: _renderLocalArt,
              onInspect: () => showDialog<void>(
                context: context,
                builder: (_) =>
                    CgmsCardInspector(card: card, showArt: _renderLocalArt),
              ),
            ),
          ConcealedCard(width: 112, showArt: _renderLocalArt),
        ],
      ),
      const SizedBox(height: 28),
      const CgmsSectionTitle(title: 'Actions & feedback'),
      const SizedBox(height: 16),
      Wrap(
        spacing: 12,
        runSpacing: 12,
        children: [
          FilledButton(
            onPressed: () => ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text(
                  'Action received. This gallery does not change the game.',
                ),
              ),
            ),
            child: const Text('Try an action'),
          ),
          const FilledButton(
            onPressed: null,
            child: Text('Attack unavailable'),
          ),
          OutlinedButton(
            onPressed: () => _go(context, '/scenarios'),
            child: const Text('Open scenarios'),
          ),
        ],
      ),
      const SizedBox(height: 10),
      const Text(
        'Attack unavailable: select an unused exposed Club with an exact legal target.',
      ),
      const SizedBox(height: 24),
      for (final (title, message, icon) in const [
        (
          'Pending',
          'Waiting for Mara’s response. There is no automatic pass.',
          Icons.hourglass_empty,
        ),
        (
          'Loading',
          'Reading the last confirmed table. Actions stay paused.',
          Icons.sync,
        ),
        ('Empty', 'There are no offers at the moment.', Icons.inbox_outlined),
        (
          'Error',
          'The action was not submitted. Review the table before retrying.',
          Icons.error_outline,
        ),
        (
          'Disconnected',
          'Your required response remains pending until reconnection.',
          Icons.cloud_off,
        ),
        (
          'Resolved',
          'The Hearts returned to the deck. Your Club remains exposed and used.',
          Icons.check_circle_outline,
        ),
      ])
        Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: CgmsNotice(title: title, message: message, icon: icon),
        ),
    ],
  );

  Future<void> _settings(BuildContext context) => showDialog<void>(
    context: context,
    builder: (dialogContext) => StatefulBuilder(
      builder: (context, update) => AlertDialog(
        title: const Text('Display settings'),
        content: SingleChildScrollView(
          child: SizedBox(
            width: 400,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (!_approvedMode &&
                    _tableKind(
                          _router.tableRoute,
                          MediaQuery.sizeOf(context).width,
                        ) ==
                        'desktop')
                  SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    title: const Text('Compact table'),
                    subtitle: const Text(
                      'Smaller cards keep public positions and your hand together.',
                    ),
                    value: _compactTable,
                    onChanged: (value) {
                      setState(() => _compactTable = value);
                      _router.refresh();
                      update(() {});
                    },
                  ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Card artwork'),
                  subtitle: const Text('Labels remain when artwork is hidden.'),
                  value: _showArt,
                  onChanged: (value) {
                    setState(() => _showArt = value);
                    _router.refresh();
                    update(() {});
                  },
                ),
                SwitchListTile(
                  contentPadding: EdgeInsets.zero,
                  title: const Text('Reduced motion'),
                  subtitle: const Text(
                    'System preference is always respected.',
                  ),
                  value: _reduceMotion,
                  onChanged: (value) {
                    setState(() => _reduceMotion = value);
                    update(() {});
                  },
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<double>(
                  initialValue: _textScale,
                  decoration: const InputDecoration(labelText: 'Text size'),
                  items: const [
                    DropdownMenuItem(value: 1, child: Text('System / 100%')),
                    DropdownMenuItem(value: 1.5, child: Text('150%')),
                    DropdownMenuItem(value: 2, child: Text('200%')),
                  ],
                  onChanged: (value) {
                    setState(() => _textScale = value!);
                    update(() {});
                  },
                ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Done'),
          ),
        ],
      ),
    ),
  );
}

class _DemoRouteParser extends RouteInformationParser<String> {
  _DemoRouteParser({required this.defaultRoute});
  final String defaultRoute;
  @override
  Future<String> parseRouteInformation(RouteInformation information) =>
      SynchronousFuture(
        information.uri.path.isEmpty || information.uri.path == '/'
            ? defaultRoute
            : information.uri.toString(),
      );
  @override
  RouteInformation restoreRouteInformation(String configuration) =>
      RouteInformation(uri: Uri.parse(configuration));
}

class _DemoRouter extends RouterDelegate<String>
    with ChangeNotifier, PopNavigatorRouterDelegateMixin<String> {
  _DemoRouter({required this.builder, this.onNavigate});
  final Widget Function(BuildContext, String) builder;
  final ValueChanged<String>? onNavigate;
  String _route = '/setup';
  String tableRoute = '/play';
  @override
  final GlobalKey<NavigatorState> navigatorKey = GlobalKey<NavigatorState>();
  @override
  String get currentConfiguration => _route;
  void go(String route) {
    _route = route;
    if (_isTableRoute(route)) tableRoute = route;
    onNavigate?.call(route);
    notifyListeners();
  }

  void refresh() => notifyListeners();
  @override
  Future<void> setNewRoutePath(String configuration) {
    _route = configuration;
    if (_isTableRoute(configuration)) {
      tableRoute = configuration;
    }
    onNavigate?.call(configuration);
    return SynchronousFuture<void>(null);
  }

  @override
  Widget build(BuildContext context) => Navigator(
    key: navigatorKey,
    pages: [
      MaterialPage<void>(
        key: const ValueKey('demo-shell'),
        child: Builder(builder: (context) => builder(context, _route)),
      ),
    ],
    onDidRemovePage: (_) {},
  );
}
