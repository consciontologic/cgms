import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'online_client.dart';
import 'command_contract.dart';
import 'projection.dart';
import 'economy_panel.dart';
import 'card_intents.dart';
import 'card_play_choices.dart';

/// Real online host. Owns drafts independently from adaptive presentation.
class OnlineScreen extends StatefulWidget {
  const OnlineScreen({
    super.key,
    required this.client,
    this.matchId,
    this.roomId,
    this.onMatchChanged,
    this.onRoomChanged,
  });
  final OnlineClient client;
  final String? matchId;
  final String? roomId;
  final ValueChanged<String>? onMatchChanged;
  final ValueChanged<String>? onRoomChanged;
  @override
  State<OnlineScreen> createState() => _OnlineScreenState();
}

class _OnlineScreenState extends State<OnlineScreen> {
  final _username = TextEditingController(),
      _password = TextEditingController();
  final _room = TextEditingController(), _invitation = TextEditingController();
  final _gamesInput = TextEditingController(text: '3');
  final _allowanceController = ExpansibleController();
  final _fields = <String, TextEditingController>{};
  final _selected = <String>{};
  final _sourceCards = <String>{};
  String? _barricadeQueen;
  String? _freshOfferId;
  bool _fixedSubstitution = false;
  bool _intentPrepared = false;
  bool _dropContinuation = false;
  String _intentExplanation = '';
  List<String>? _effectChoice;
  String _action = 'open-series', _suit = 'hearts', _selectionFor = 'selection';
  final _sets = <String, Set<String>>{};
  final _flags = <String, bool>{};
  String _formationKind = 'underground',
      _subSuit = 'hearts',
      _protectSuit = '',
      _protectFormation = '';
  int _subRank = 12;
  String? _invitationResult,
      _localError,
      _draftGame,
      _draftWindow,
      _draftDecision;
  bool _clearing = false, _sessionBusy = false, _allowAnonymousEntry = false;
  String? _inspectedId, _inspectedGame;
  int _capacity = 4, _games = 3;
  String _botDifficulty = 'beginner';
  bool _validateRoom = false;
  bool _art = true, _rtl = false;
  String? _popup;
  final _popupFocus = FocusNode(debugLabel: 'Card context');
  FocusNode? _returnFocus;
  Object? _popupGeometry;
  Object? _draftIdentity;
  bool _submitting = false;
  List<CardPlayIntent> _choices = const [];
  Object? _requiredActionsKey;
  int _requiredActionsGeneration = 0;
  double _scale = 1;
  OnlineClient get c => widget.client;
  bool get _canMutate =>
      !_submitting &&
      c.canSubmit &&
      c.snapshot?.online['financial_phase'] != 'finalized';
  bool get _roomReady => !c.busy && !c.hasPendingRoomOperation;
  bool get _entryReady => !c.busy && !_sessionBusy;
  String? get _displayError {
    final code = _localError ?? c.errorCode;
    if (c.sessionLoading) return null;
    // No cookie is the normal entry state. Explicit authentication failures and
    // expired room/match links still need their recovery feedback.
    if (_allowAnonymousEntry &&
        c.account == null &&
        c.snapshot == null &&
        widget.roomId == null &&
        widget.matchId == null &&
        code == 'UNAUTHENTICATED')
      return null;
    return code;
  }

  bool get _ownsRoom => c.account?['id'] == c.room?['owner_id'];
  bool get _roomFull =>
      (c.room?['members'] as List? ?? []).length == c.room?['capacity'];
  bool get _roomStarted => (c.room?['match_id'] as String? ?? '').isNotEmpty;
  bool get _botRoom =>
      c.room?['bot_difficulty'] is String &&
      (c.room!['bot_difficulty'] as String).isNotEmpty;
  CgmsLocalizations get l10n => CgmsLocalizations.of(context);
  @override
  void initState() {
    super.initState();
    HardwareKeyboard.instance.addHandler(_cancelContextOnEscape);
    c.addListener(_changed);
    _showRequiredActions();
    _restore();
  }

  bool _cancelContextOnEscape(KeyEvent event) {
    // A pointer drop can leave browser focus outside Flutter's focus tree.
    // Escape still cancels this route's unsubmitted intent in that state.
    if (mounted &&
        ModalRoute.of(context)?.isCurrent != false &&
        event is KeyDownEvent &&
        event.logicalKey == LogicalKeyboardKey.escape &&
        (_popup != null || _selected.isNotEmpty || _intentPrepared)) {
      _dismissPopup(clear: true);
    }
    // Let the table invalidate any held drag before the pointer is released.
    return false;
  }

  Future<void> _restore() async {
    await _sessionRequest(() async {
      await c.restoreSession();
      await _loadLocation();
    }, allowAnonymousEntry: true);
  }

  Future<void> _sessionRequest(
    Future<void> Function() action, {
    bool allowAnonymousEntry = false,
  }) async {
    if (!_entryReady) return;
    setState(() {
      _sessionBusy = true;
      _allowAnonymousEntry = allowAnonymousEntry;
    });
    try {
      await _run(action);
    } finally {
      if (mounted) setState(() => _sessionBusy = false);
    }
  }

  Future<void> _loadLocation() async {
    if (widget.matchId != null) {
      await c.watchMatch(widget.matchId!);
    } else if (widget.roomId != null) {
      await c.refreshRoom(widget.roomId!);
    }
  }

  @override
  void didUpdateWidget(OnlineScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.matchId != oldWidget.matchId &&
        widget.matchId != null &&
        widget.matchId != c.matchId)
      _run(() => c.watchMatch(widget.matchId!));
    else if (widget.matchId == null &&
        widget.roomId != oldWidget.roomId &&
        widget.roomId != null &&
        widget.roomId != c.room?['id'])
      _run(() => c.refreshRoom(widget.roomId!));
  }

  Object? _requiredActionsFor(AuthoritativeSnapshot? s) {
    if (s == null) return null;
    final requiredActor = s.board['required_actor'];
    final completing =
        s.phase != 'playing' ||
        s.online['departure_pending'] == true ||
        s.admissionWaiting ||
        s.online['financial_phase'] == 'finalized';
    if (!completing &&
        (requiredActor == null ||
            requiredActor == 0 ||
            requiredActor != s.seat)) {
      return null;
    }
    return (
      s.gameId,
      s.phase,
      s.admissionWaiting,
      s.online['financial_phase'],
      s.online['completed_games'],
      s.online['departure_pending'] == true ? s.board['round'] : null,
      s.online['departure_pending'],
      requiredActor,
      s.board['window_id'],
      s.board['decision_id'],
      s.online['decision_stage'],
    );
  }

  void _showRequiredActions() {
    final key = _requiredActionsFor(c.snapshot);
    if (key == _requiredActionsKey) return;
    _requiredActionsKey = key;
    final generation = ++_requiredActionsGeneration;
    if (key == null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted &&
            generation == _requiredActionsGeneration &&
            _popup == 'required' &&
            _requiredActionsFor(c.snapshot) == null) {
          _dismissPopup();
        }
      });
      return;
    }
    // A dismissed decision stays pending; only a new stage opens its context.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted &&
          generation == _requiredActionsGeneration &&
          key == _requiredActionsFor(c.snapshot)) {
        setState(() => _openPopup('required'));
      }
    });
  }

  String? _reportedMatch;
  String? _reportedRoom;
  void _changed() {
    if (!mounted) return;
    setState(() {});
    _showRequiredActions();
    final id = c.matchId;
    if (id != null && id != widget.matchId && id != _reportedMatch) {
      _reportedMatch = id;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && c.matchId == id) widget.onMatchChanged?.call(id);
      });
    }
    final roomId = c.room?['id'] as String?;
    if (id == null &&
        roomId != null &&
        roomId != widget.roomId &&
        roomId != _reportedRoom) {
      _reportedRoom = roomId;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && c.matchId == null && c.room?['id'] == roomId)
          widget.onRoomChanged?.call(roomId);
      });
    }
  }

  Future<void> _run(Future<void> Function() action) async {
    try {
      await action();
      if (mounted) setState(() => _localError = null);
    } on FormatException {
      if (mounted)
        setState(
          () => _localError = c.errorCode == 'INVALID_RESPONSE'
              ? 'INVALID_RESPONSE'
              : 'INVALID_COMMAND',
        );
    } on StateError catch (error) {
      if (mounted)
        setState(
          () => _localError = error.message == 'STATE_CONFLICT'
              ? 'STATE_CONFLICT'
              : c.errorCode ?? 'UNAVAILABLE',
        );
    } catch (_) {
      if (mounted) setState(() => _localError = c.errorCode ?? 'UNAVAILABLE');
    }
  }

  @override
  void dispose() {
    HardwareKeyboard.instance.removeHandler(_cancelContextOnEscape);
    _popupFocus.dispose();
    c.removeListener(_changed);
    _allowanceController.dispose();
    for (final v in [
      _username,
      _password,
      _room,
      _invitation,
      _gamesInput,
      ..._fields.values,
    ]) {
      v.dispose();
    }
    super.dispose();
  }

  TextEditingController field(String name, [String initial = '']) =>
      _fields.putIfAbsent(name, () {
        final controller = TextEditingController(text: initial);
        controller.addListener(() {
          if (!_clearing) {
            _captureDraft();
            if (mounted) setState(() {});
          }
        });
        return controller;
      });
  void _captureDraft() {
    final s = c.snapshot;
    if (s == null || _draftGame != null) return;
    _draftGame = s.gameId;
    _draftWindow = s.board['window_id'] as String?;
    _draftDecision = s.board['decision_id'] as String?;
    _draftIdentity = _snapshotIdentity(s);
  }

  Object _snapshotIdentity(AuthoritativeSnapshot s) => (
    s.gameId,
    s.version,
    s.phase,
    s.board['round'],
    s.board['turn'],
    s.board['active'],
    s.board['window_id'],
    s.board['decision_id'],
    s.board['required_actor'],
    s.online['decision_stage'],
  );

  bool _draftMatches(AuthoritativeSnapshot s) =>
      _draftGame == null ||
      (_draftGame == s.gameId &&
          (const {'coup', 'declare-ordinary'}.contains(_action) &&
                  _intentPrepared ||
              _draftIdentity == _snapshotIdentity(s)));
  List<String> get _selectionFields => [
    for (final field in commandFields[_action]!)
      if (['selection', 'targets', 'payment'].contains(field)) field,
    if (commandFields[_action]!.contains('terms')) ...['give', 'receive'],
    if (commandFields[_action]!.contains('formation')) ...[
      'formation',
      if (_flags['substitute'] == true) 'kings',
    ],
  ];
  void _selectAction(String action) {
    _intentPrepared = true;
    _action = action;
    _selectionFor = _selectionFields.firstOrNull ?? 'selection';
    _selected
      ..clear()
      ..addAll(_selectionFields.expand((field) => _sets[field] ?? <String>{}));
  }

  void _clearDraft() {
    _clearing = true;
    for (final field in _fields.values) {
      field.clear();
    }
    _intentPrepared = false;
    _dropContinuation = false;
    _intentExplanation = '';
    _effectChoice = null;
    _choices = const [];
    _flags.clear();
    _sets.clear();
    _selected.clear();
    _sourceCards.clear();
    _barricadeQueen = null;
    _freshOfferId = null;
    _fixedSubstitution = false;
    _selectionFor = _selectionFields.firstOrNull ?? 'selection';
    _suit = 'hearts';
    _formationKind = 'underground';
    _subSuit = 'hearts';
    _protectSuit = '';
    _protectFormation = '';
    _subRank = 12;
    _draftIdentity = null;
    _draftGame = null;
    _draftWindow = null;
    _draftDecision = null;
    _clearing = false;
  }

  String _currentCardNames(AuthoritativeSnapshot s, Iterable<String> ids) {
    final cards = (s.board['cards'] as List? ?? []).cast<Map>();
    return ids
        .map((id) {
          final card = cards.where((p) => p['card']['id'] == id).firstOrNull;
          return card == null
              ? l10n.onlineCardUnavailable
              : projectCard(card).label;
        })
        .join(', ');
  }

  // A first opening can contain every held number copy of a suit. Keep that
  // exact physical set readable during a drag without changing normal card
  // semantics or the full names in inspection and confirmation.
  String _previewCardNames(AuthoritativeSnapshot s, Iterable<String> ids) {
    final requested = ids.toList();
    final cards = (s.board['cards'] as List? ?? []).cast<Map>();
    final grouped = <String, Map<int, List<int>>>{};
    for (final id in requested) {
      final placed = cards.where((p) => p['card']['id'] == id).firstOrNull;
      if (placed == null) return l10n.onlineCardUnavailable;
      final face = placed['card'] as Map;
      final suit = face['suit'] as String;
      final rank = face['rank'] as int;
      grouped
          .putIfAbsent(suit, () => {})
          .putIfAbsent(rank, () => [])
          .add(face['deck'] as int);
    }
    final groups = <String>[];
    for (final suit in CardSuit.values.map((suit) => suit.name)) {
      final ranks = grouped[suit];
      if (ranks == null) continue;
      final ordered = ranks.keys.toList()..sort();
      groups.add(
        '${label(suit)}: ${ordered.map((rank) {
          final copies = ranks[rank]!..sort();
          final face = switch (rank) {
            1 => 'A',
            11 => 'J',
            12 => 'Q',
            13 => 'K',
            _ => '$rank',
          };
          return '$face[${copies.join(',')}]';
        }).join(' ')}',
      );
    }
    return groups.join('; ');
  }

  bool _draftCardsVisible(AuthoritativeSnapshot s) {
    final visible = (s.board['cards'] as List? ?? [])
        .map((p) => p['card']['id'])
        .toSet();
    final ids = <String>{..._selected, ...?_effectChoice};
    final ace = _fields['ace']?.text ?? '';
    if (ace.isNotEmpty) ids.add(ace);
    if (!ids.every(visible.contains)) return false;
    if (!_sourceCards.every(
      (id) => (s.board['cards'] as List? ?? []).any(
        (p) =>
            p['card']['id'] == id &&
            p['controller'] == s.seat &&
            (p['available_from_round'] as int? ?? 0) <=
                (s.board['round'] as int),
      ),
    ))
      return false;
    if (_effectChoice != null &&
        !_currentEffectChoices(s).any(
          (choice) =>
              choice.length == _effectChoice!.length &&
              choice.every(_effectChoice!.contains),
        ))
      return false;
    if (!_intentPrepared) return true;
    for (final role in _selectionFields) {
      if (!(_sets[role] ?? {}).every((id) => _acceptCardForRole(s, id, role)))
        return false;
    }
    if (ace.isNotEmpty && !_acceptCardForRole(s, ace, 'ace')) return false;
    final formationId = _fields['formation_id']?.text ?? '';
    if (_action != 'open-formation' &&
        (commandFields[_action]!.contains('formation_id') ||
            _action == 'justice' && formationId.isNotEmpty) &&
        !(s.board['formations'] as List? ?? []).any(
          (f) => f['id'] == formationId && f['controller'] == s.seat,
        ))
      return false;
    return true;
  }

  bool _isKidnapperOutcome(AuthoritativeSnapshot s) =>
      s.board['required_actor'] == s.seat &&
      s.board['window_id'] != null &&
      s.board['decision_id'] == null &&
      s.board['decision_kind'] == 'kidnapper-outcome';

  Iterable<List> _currentEffectChoices(AuthoritativeSnapshot s) {
    if (_action == 'kidnapper-outcome') {
      if (!_isKidnapperOutcome(s)) return const [];
    } else if (_action != 'decision' ||
        s.board['required_actor'] != s.seat ||
        s.board['window_id'] == null ||
        s.board['decision_id'] == null) {
      return const [];
    }
    return (s.board['choices'] as List? ?? []).whereType<List>();
  }

  bool _acceptCardForRole(AuthoritativeSnapshot s, String id, String role) {
    final card = (s.board['cards'] as List? ?? [])
        .cast<Map>()
        .where((p) => p['card']['id'] == id)
        .firstOrNull;
    if (card == null) return false;
    if (_action == 'decision')
      return _currentEffectChoices(s).any(
        (choice) => choice.contains(id),
      ); // The offered choice is checked before submission.
    if (_action == 'kidnapper-outcome')
      return role == 'selection' &&
          card['controller'] != s.seat &&
          (card['card']['rank'] as int) >= 11 &&
          ['attachment', 'formation', 'unassigned'].contains(card['zone']) &&
          _currentEffectChoices(
            s,
          ).any((choice) => choice.length == 1 && choice.single == id);
    final owner = card['controller'] == s.seat;
    final zone = card['zone'];
    if (role == 'targets' || role == 'receive') {
      return !owner &&
          ['series', 'attachment', 'formation', 'unassigned'].contains(zone);
    }
    if (!owner ||
        ![
          'hand',
          'concealed-ace',
          'series',
          'attachment',
          'formation',
          'unassigned',
        ].contains(zone))
      return false;
    if (_action != 'expose-unused-ace' &&
        (card['available_from_round'] as int? ?? 0) > (s.board['round'] as int))
      return false;
    final rank = card['card']['rank'] as int;
    final suit = card['card']['suit'];
    if (role == 'ace')
      return _action == 'justice'
          ? rank == 12 &&
                zone == 'formation' &&
                (s.board['formations'] as List? ?? []).any(
                  (f) =>
                      f['id'] == card['allocation'] &&
                      f['controller'] == s.seat &&
                      ((_fields['formation_id']?.text ?? '').isEmpty ||
                          f['id'] == _fields['formation_id']!.text) &&
                      f['spec']['kind'] == 'justice',
                )
          : rank == 1 && zone == 'concealed-ace';
    if (role == 'kings')
      return rank == 13 && (_sets['formation'] ?? {}).contains(id);
    if (role == 'selection' && _action == 'dexter-assassination')
      return rank >= 2 && rank <= 10 && suit == 'diamonds' && zone == 'series';
    if (role == 'selection' && _action == 'fate')
      return rank == 13 && card['allocation'] == _fields['formation_id']?.text;

    if (role == 'payment')
      return rank >= 2 &&
          rank <= 10 &&
          suit == 'spades' &&
          ['hand', 'series'].contains(zone);
    if (role == 'selection' && _action.contains('attack'))
      return rank >= 2 &&
          rank <= 10 &&
          suit == 'clubs' &&
          zone == 'series' &&
          ((s.board['turn'] as int? ?? 0) == 0 ||
              card['used_turn'] != s.board['turn']);
    if (role == 'formation')
      return rank >= 11 && ['hand', 'unassigned', 'formation'].contains(zone);
    return true;
  }

  bool _formationReady(AuthoritativeSnapshot s) {
    if (_action != 'open-formation') return true;
    if (_formationKind.isEmpty) return false;
    if (!formationHonorsFixedBindings(
      s.board,
      _sets['formation'] ?? {},
      substitute: _flags['substitute'] == true
          ? {
              'kings': (_sets['kings'] ?? {}).toList(),
              'slot': {'rank': _subRank, 'suit': _subSuit},
            }
          : null,
    ))
      return false;
    if (_flags['substitute'] == true) {
      final kings = _sets['kings'] ?? <String>{};
      if (kings.length != 2 ||
          !(_sets['formation'] ?? {}).containsAll(kings) ||
          ![11, 12, 13].contains(_subRank) ||
          _subSuit.isEmpty ||
          ['coup', 'ponzi'].contains(_formationKind))
        return false;
      final cards = (s.board['cards'] as List? ?? []).cast<Map>();
      if (!kings.every(
        (id) => cards.any(
          (p) =>
              p['card']['id'] == id &&
              p['card']['rank'] == 13 &&
              p['controller'] == s.seat,
        ),
      ))
        return false;
    }
    if (['people', 'great-people'].contains(_formationKind)) {
      if (_flags['protection'] != true ||
          _protectSuit.isEmpty == _protectFormation.isEmpty)
        return false;
      if (_protectSuit.isNotEmpty &&
          !['clubs', 'spades', 'diamonds'].contains(_protectSuit))
        return false;
      if (_protectSuit.isNotEmpty &&
          !(s.board['cards'] as List? ?? []).any(
            (p) =>
                p['controller'] == s.seat &&
                p['zone'] == 'series' &&
                p['card']['suit'] == _protectSuit &&
                (p['card']['rank'] as int) >= 2 &&
                (p['card']['rank'] as int) <= 10,
          ))
        return false;
      if (_protectFormation.isNotEmpty &&
          !_canProtectFormation(s, _protectFormation))
        return false;
    }
    return true;
  }

  bool _canProtectFormation(AuthoritativeSnapshot s, String id) =>
      _formationKind == 'great-people' &&
      _flags['substitute'] != true &&
      id != _fields['formation_id']?.text &&
      (s.board['formations'] as List? ?? []).any(
        (f) =>
            f['id'] == id &&
            f['controller'] == s.seat &&
            !const {'people', 'great-people'}.contains(f['spec']['kind']),
      );

  bool _combatChoiceReady() {
    final value = int.tryParse(_fields['value']?.text ?? '');
    if (_action == 'numerical-defense')
      return value != null && value >= 2 && value <= 10;
    if (_action.contains('attack') && (_fields['ace']?.text ?? '').isNotEmpty) {
      final clubAce = (c.snapshot?.board['cards'] as List? ?? []).any(
        (p) =>
            p['card']['id'] == _fields['ace']?.text &&
            p['card']['suit'] == 'clubs',
      );
      return value != null &&
          (value == 0 && clubAce || value >= 2 && value <= 10);
    }
    return true;
  }

  void _selectCard(String id) {
    final s = c.snapshot;
    if (s == null || !_canMutate || _effectChoice != null) return;
    final card = (s.board['cards'] as List? ?? [])
        .cast<Map>()
        .where((p) => p['card']['id'] == id)
        .firstOrNull;
    if (card == null) return;
    if (!_draftMatches(s)) {
      setState(() => _openPopup('review'));
      return;
    }
    setState(() {
      _captureDraft();
      if (!_intentPrepared) {
        if (!_selected.add(id)) _selected.remove(id);
        return;
      }
      if (_action.contains('attack') &&
          card['controller'] == s.seat &&
          card['zone'] == 'attachment' &&
          card['allocation'] == 'clubs' &&
          card['card']['suit'] == 'clubs' &&
          [11, 12].contains(card['card']['rank'])) {
        final flag = card['card']['rank'] == 11 ? 'infiltrator' : 'exile';
        _flags[flag] = !(_flags[flag] ?? false);
        if (_flags[flag]!) {
          _selected.add(id);
          _sourceCards.add(id);
        } else {
          _selected.remove(id);
          _sourceCards.remove(id);
        }
        return;
      }
      if (_selectionFor == 'ace') {
        if (!_acceptCardForRole(s, id, 'ace')) return;
        final old = field('ace').text;
        field('ace').text = old == id ? '' : id;
        _selected.remove(old);
        if (old != id) _selected.add(id);
        return;
      }
      if (_selectionFor == 'attachment' || _selectionFor == 'protection') {
        if (card['controller'] != s.seat) return;
        if (_selectionFor == 'attachment' && card['zone'] == 'series')
          _suit = card['card']['suit'] as String;
        if (_selectionFor == 'protection') {
          if (card['zone'] == 'series' && card['card']['suit'] != 'hearts') {
            _protectSuit = card['card']['suit'] as String;
            _protectFormation = '';
          } else if (card['zone'] == 'formation') {
            _protectFormation = card['allocation'] as String;
            _protectSuit = '';
          }
        }
        return;
      }
      final role =
          _action.contains('attack') &&
              card['controller'] == s.seat &&
              card['zone'] == 'series' &&
              card['card']['suit'] == 'clubs'
          ? 'selection'
          : _selectionFor;
      if (!_selectionFields.contains(role) || !_acceptCardForRole(s, id, role))
        return;
      final selection = _sets.putIfAbsent(role, () => {});
      if (!selection.add(id)) selection.remove(id);
      if (role == 'targets' &&
          _action.contains('attack') &&
          selection.isNotEmpty) {
        _suit = card['zone'] == 'attachment'
            ? '${card['allocation']}'
            : '${card['card']['suit']}';
      }
      _selected
        ..clear()
        ..addAll(_sourceCards)
        ..addAll(_sets.values.expand((ids) => ids));
      final ace = _fields['ace']?.text ?? '';
      if (ace.isNotEmpty) _selected.add(ace);
    });
  }

  void _prepareIntent(CardPlayIntent intent, {bool showContext = true}) {
    final s = c.snapshot;
    if (s == null || !commandFields.containsKey(intent.type)) return;
    final sources = Set<String>.of(_selected);
    _clearDraft();
    _selectAction(intent.type);
    if (intent.sourceFormationId != null) {
      field('formation_id').text = intent.sourceFormationId!;
    }
    _sourceCards.addAll(
      sources.where(
        (id) => (s.board['cards'] as List? ?? []).any(
          (p) => p['card']['id'] == id && p['controller'] == s.seat,
        ),
      ),
    );
    _captureDraft();
    _intentExplanation = intent.explanation;
    if (intent.needsSubstitution) {
      _flags['substitute'] = true;
      _subRank = 0;
      _subSuit = '';
    }
    if (intent.needsProtection) _flags['protection'] = true;
    _clearing = true;
    for (final entry in intent.seed.entries) {
      final value = entry.value;
      if ([
        'selection',
        'targets',
        'payment',
        'give',
        'receive',
      ].contains(entry.key)) {
        _sets[entry.key] = (value as List).cast<String>().toSet();
      } else if (entry.key == 'suit') {
        _suit = value as String;
      } else if (entry.key == 'formation') {
        final formation = value as Map;
        _formationKind = formation['kind'] as String? ?? '';
        _sets['formation'] = (formation['cards'] as List? ?? [])
            .cast<String>()
            .toSet();
        final substitute = formation['substitute'] as Map?;
        if (substitute != null) {
          _fixedSubstitution = true;
          _flags['substitute'] = true;
          _sets['kings'] = (substitute['kings'] as List).cast<String>().toSet();
          _subRank = substitute['slot']['rank'] as int;
          _subSuit = substitute['slot']['suit'] as String;
        }
      } else if (entry.key == 'terms') {
        final terms = value as Map;
        _sets['give'] = (terms['give'] as List? ?? []).cast<String>().toSet();
        _sets['receive'] = (terms['receive'] as List? ?? [])
            .cast<String>()
            .toSet();
        _flags['loan'] = terms['loan'] == true;
        if (terms['to'] != null) field('target_seat').text = '${terms['to']}';
      } else if (value is bool) {
        _flags[entry.key] = value;
      } else {
        field(entry.key).text = '$value';
      }
    }
    if (intent.type == 'open-formation' && field('formation_id').text.isEmpty) {
      field('formation_id').text =
          'formation-${DateTime.now().microsecondsSinceEpoch}';
    }
    if (intent.type == 'offer') {
      _freshOfferId = 'offer-${DateTime.now().microsecondsSinceEpoch}';
      field('offer_id').text = _freshOfferId!;
      field('revision').text = '1';
    }
    if (intent.type == 'barricade-sacrifice') {
      _barricadeQueen = _sets['selection']?.firstOrNull;
    }
    if (intent.type == 'attach' && !intent.seed.containsKey('suit')) {
      _suit = '';
    }
    if (intent.type == 'purchase' && field('quantity').text.isEmpty)
      field('quantity').text = '1';
    if (intent.type == 'promise-offer') {
      field('promise_id').text =
          'promise-${DateTime.now().microsecondsSinceEpoch}';
    }
    _clearing = false;
    // Explicit roles own their selection state. Only activation cards which
    // are not payload members stay in the supplemental source set.
    _sourceCards.removeAll(_sets.values.expand((ids) => ids));
    if (intent.type == 'expose-unused-ace') {
      // Its exact Ace field has disclosure-specific validation. A restriction
      // blocks abilities, not the controller's explicit decision to waste it.
      _sourceCards.remove(_fields['ace']?.text);
    }
    final sourceFormation = _fields['formation_id']?.text ?? '';
    if (sourceFormation.isNotEmpty) {
      _sourceCards.removeWhere(
        (id) => (s.board['cards'] as List? ?? []).any(
          (p) =>
              p['card']['id'] == id &&
              p['zone'] == 'formation' &&
              p['allocation'] != sourceFormation,
        ),
      );
    }
    if (['decision', 'kidnapper-outcome'].contains(intent.type) &&
        intent.seed['selection'] is List) {
      _effectChoice = (intent.seed['selection'] as List)
          .cast<String>()
          .toList();
    }
    if (intent.type == 'take-back') {
      final formation = (s.board['formations'] as List? ?? [])
          .cast<Map>()
          .where(
            (f) =>
                f['id'] == _fields['formation_id']?.text &&
                f['controller'] == s.seat,
          )
          .firstOrNull;
      if (formation != null)
        _sourceCards.addAll(
          (formation['spec']['cards'] as List).cast<String>(),
        );
    }
    _selected.addAll(_sourceCards);
    _selected.addAll(
      _selectionFields.expand((name) => _sets[name] ?? <String>{}),
    );
    final ace = _fields['ace']?.text ?? '';
    if (ace.isNotEmpty) _selected.add(ace);
    // Victory is admitted against current holdings in this same game. An
    // unrelated selection is neither a cost nor authority for that claim.
    if (const {'coup', 'declare-ordinary'}.contains(intent.type)) {
      _sourceCards.clear();
      _selected.clear();
    }
    _selectionFor = _selectionFields.firstOrNull ?? 'selection';
    if (intent.needsSubstitution) _selectionFor = 'kings';
    if ((_sets['targets'] ?? {}).isEmpty &&
        _selectionFields.contains('targets')) {
      _selectionFor = 'targets';
    }
    if (showContext) _openPopup('review');
  }

  void _openPopup(String kind) {
    if (_popup == null) _returnFocus = FocusManager.instance.primaryFocus;
    _popup = kind;
    _popupGeometry = null;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && _popup != null) _popupFocus.requestFocus();
    });
  }

  void _dismissPopup({bool clear = false}) {
    setState(() {
      _popup = null;
      _inspectedId = null;
      _inspectedGame = null;
      if (clear) _clearDraft();
    });
    final focus = _returnFocus;
    if (focus != null && focus.context != null && focus.canRequestFocus) {
      focus.requestFocus();
    }
  }

  void _activateCard(String id) {
    final s = c.snapshot;
    if (s == null || !_canMutate) return;
    if (_intentPrepared && _selected.contains(id) && _draftMatches(s)) {
      setState(() => _openPopup('review'));
      return;
    }
    setState(() {
      if (!_draftMatches(s) || _intentPrepared) _clearDraft();
      _captureDraft();
      _selected.add(id);
      _chooseMoves();
    });
  }

  void _chooseMoves([TableDropTarget? target]) {
    final s = c.snapshot;
    if (s == null || !_draftMatches(s)) return;
    final intents = suggestCardPlays(
      board: s.board,
      online: s.online,
      cardIds: _selected,
      target: target,
    );
    if (intents.length == 1 && intents.single.type != 'expose-unused-ace') {
      _prepareIntent(intents.single);
    } else {
      _choices = intents;
      _openPopup('choices');
    }
  }

  bool _suggestedTarget(AuthoritativeSnapshot s, TableDropTarget target) {
    if (!_canMutate || !_draftMatches(s) || _selected.isEmpty) return false;
    if (!_intentPrepared)
      return suggestCardDrops(
        board: s.board,
        online: s.online,
        cardIds: _selected,
        target: target,
      ).isNotEmpty;
    if (_selectionFor == 'attachment')
      return target.seat == s.seat &&
          target.zone == TableDropZone.series &&
          royalAttachmentSuits(
            s.board,
            _sets['selection'] ?? {},
          ).contains(target.suit);
    if (_selectionFor == 'protection')
      return target.seat == s.seat &&
          (target.zone == TableDropZone.series && target.suit != 'hearts' ||
              target.zone == TableDropZone.formation &&
                  _canProtectFormation(s, target.formationId ?? ''));
    if (target.cardId != null)
      return _acceptCardForRole(s, target.cardId!, _selectionFor);
    if (_selectionFor == 'targets')
      return target.seat != s.seat && target.zone == TableDropZone.series;
    return false;
  }

  String _dropPreview(
    AuthoritativeSnapshot s,
    TableCardDrag drag,
    TableDropTarget target,
  ) {
    final intents = suggestCardDrops(
      board: s.board,
      online: s.online,
      cardIds: drag.cardIds.toSet(),
      target: target,
    );
    if (intents.isEmpty) return l10n.onlineNoMoveDestination;
    final destination = [
      if (target.seat != null) l10n.onlineDropSeat(target.seat!),
      if (target.suit != null) label(target.suit!),
      if (target.cardId != null) _previewCardNames(s, [target.cardId!]),
      if (target.zone == TableDropZone.drawPile) label('purchase'),
    ].join(' · ');
    if (intents.length != 1)
      return '${l10n.onlineReleaseChoice(intents.map((i) => label(i.type)).toSet().join(' / '))}'
          '\n${_previewCardNames(s, drag.cardIds)} / $destination\n${l10n.onlineDropCopyLegend}';
    final intent = intents.single;
    final seed = intent.seed;
    final ids = <String>{
      ...drag.cardIds,
      for (final role in ['selection', 'payment', 'give'])
        ...((seed[role] as List?)?.cast<String>() ?? <String>[]),
      ...((seed['formation']?['cards'] as List?)?.cast<String>() ?? <String>[]),
      ...((seed['terms']?['give'] as List?)?.cast<String>() ?? <String>[]),
      if (seed['formation_id'] != null || intent.sourceFormationId != null)
        ..._formationMembers(
          s,
          (seed['formation_id'] ?? intent.sourceFormationId) as String,
        ),
    };
    final isLoan = intent.type == 'offer' && seed['terms']?['loan'] == true;
    final action = isLoan
        ? l10n.onlineLoanIntactFormation
        : '${label(intent.type)}${seed['formation']?['kind'] == null ? '' : ' · ${label(seed['formation']['kind'] as String)}'}';
    final targets = (seed['targets'] as List? ?? []).cast<String>();
    final pending = (s.online['rules_context'] as Map?)?['pending'];
    final substitute = seed['formation']?['substitute'] as Map?;
    final details = [
      if (substitute != null)
        '${l10n.onlineReplacementRank} · ${substitute['slot']['rank']} · ${label(substitute['slot']['suit'] as String)}',
      if (targets.isNotEmpty)
        l10n.onlineFrozenTargets(_previewCardNames(s, targets)),
      if (seed['ace'] is String)
        l10n.onlineCardRoleSummary(
          intent.type == 'justice'
              ? l10n.onlineConsumedQueen
              : l10n.onlineModifier,
          _previewCardNames(s, [seed['ace'] as String]),
        ),
      if (seed['value'] != null) '${label('value')} · ${seed['value']}',
      if (seed['pending_action_id'] != null && pending is Map)
        l10n.onlinePendingSeatAction(
          '${pending['actor']}',
          label('${pending['type']}'),
        ),
    ];
    final costs = switch (intent.type) {
      'open-series' =>
        (s.board['players'] as List? ?? []).any(
              (player) =>
                  player['seat'] == s.seat &&
                  (player['history'] as List? ?? []).contains(seed['suit']),
            )
            ? l10n.onlineDropCostsOpeningExisting
            : l10n.onlineDropCostsOpeningNew,
      'open-formation' => l10n.onlineDropCostsFormation,
      'attach' => l10n.onlineDropCostsAttach,
      'richer-sacrifice' => l10n.onlineDropCostsRicher,
      'ringleader' => l10n.onlineDropCostsRingleader,
      'confinement' => l10n.onlineDropCostsConfinement,
      'compensation' ||
      'compensation-response' => l10n.onlineDropCostsCompensation,
      'main-inflation' || 'inflation' => l10n.onlineDropCostsInflation,
      'advancement-defense' ||
      'numerical-defense' => l10n.onlineDropCostsCombatResponse,
      'offer' when isLoan => l10n.onlineLoanDropTerms,
      'purchase' => l10n.onlineDropCostsPurchase,
      'return-loan' => l10n.onlineDropCostsReturnLoan,
      'fate' => l10n.onlineDropCostsFate,
      'justice' => l10n.onlineDropCostsJustice,
      'kidnapper' => l10n.onlineDropCostsKidnapper,
      'ponzi' => l10n.onlineDropCostsPonzi,
      'dexter-assassination' => l10n.onlineDropCostsDexter,
      'barricade-sacrifice' => l10n.onlineDropCostsBarricade,
      'negotiate' => l10n.onlineDropCostsNegotiator,
      'take-back' => l10n.onlineDropCostsTakeBack,
      'attack' ||
      'infiltrator-attack' ||
      'baron-attack' ||
      'bomb-attack' ||
      'baron-bomb-attack' => [
        l10n.onlineDropCostsAttack,
        if (intent.type.startsWith('baron-')) l10n.onlineDropCostsBaron,
        if (intent.type.contains('bomb')) l10n.onlineDropCostsBomb,
        if (intent.type == 'infiltrator-attack' || seed['infiltrator'] == true)
          l10n.onlineDropCostsInfiltrator,
        if (seed['exile'] == true) l10n.onlineDropCostsExile,
        if (seed.containsKey('ace')) l10n.onlineDropCostsCombatAce,
      ].join(' '),
      _ => intent.explanation,
    };
    return '${_completeDrop(intent) ? l10n.onlineReleaseMove(action) : l10n.onlineReleaseChoice(action)}'
        '\n${_previewCardNames(s, ids)} / $destination\n${[l10n.onlineDropCopyLegend, ...details, costs].join('\n')}';
  }

  bool _canDrop(TableCardDrag drag, TableDropTarget target) {
    final s = c.snapshot;
    return s != null &&
        _canMutate &&
        s.gameId == drag.gameId &&
        drag.intentIdentity == _snapshotIdentity(s) &&
        draggableCardIds(s.board, s.online).containsAll(drag.cardIds) &&
        suggestCardDrops(
          board: s.board,
          online: s.online,
          cardIds: drag.cardIds.toSet(),
          target: target,
        ).isNotEmpty;
  }

  void _onCardDrop(TableCardDrag drag, TableDropTarget target) {
    if (!_canDrop(drag, target)) return;
    final s = c.snapshot!;
    final intents = suggestCardDrops(
      board: s.board,
      online: s.online,
      cardIds: drag.cardIds.toSet(),
      target: target,
    );
    var submit = false;
    setState(() {
      _clearDraft();
      _selected.addAll(drag.cardIds);
      _captureDraft();
      _dropContinuation = true;
      if (intents.length != 1) {
        _choices = intents;
        _openPopup('choices');
        return;
      }
      final intent = intents.single;
      _prepareIntent(intent, showContext: false);
      _dropContinuation = true;
      // Quantity, trade requests, mode/value, settings and protection are real
      // choices. Never let an editor default silently complete a released move.
      submit = _completeDrop(intent) && _readyToSubmit(s);
      if (!submit) _openPopup('review');
    });
    if (submit) _submit(s);
  }

  bool _completeDrop(CardPlayIntent intent) {
    final seed = intent.seed;
    if (intent.needsSubstitution ||
        intent.needsProtection ||
        const {'purchase', 'code', 'expose-unused-ace'}.contains(intent.type))
      return false;
    if (intent.type == 'offer')
      return seed['terms']?['loan'] == true && seed['terms']?['to'] != null;
    if (intent.type == 'attach') return seed['suit'] != null;
    if (intent.type.contains('attack'))
      return _nonempty(seed['selection']) &&
          _nonempty(seed['targets']) &&
          (!seed.containsKey('ace') || seed.containsKey('value'));
    if (intent.type == 'numerical-defense') return seed.containsKey('value');
    if (intent.type == 'fate')
      return _nonempty(seed['selection']) && seed['target_seat'] != null;
    if (intent.type == 'justice') return seed['ace'] != null;
    if (intent.type == 'kidnapper' || intent.type == 'main-inflation')
      return seed['target_seat'] != null;
    if (intent.type == 'ponzi') return seed['value'] != null;
    if (intent.type == 'dexter-assassination')
      return _nonempty(seed['selection']) && _nonempty(seed['targets']);
    if (intent.type == 'barricade-sacrifice')
      return (seed['selection'] as List? ?? []).length == 6;
    return true;
  }

  bool _nonempty(dynamic value) => value is List && value.isNotEmpty;

  void _chooseCardIntent(CardPlayIntent intent) {
    final fromDrop = _dropContinuation;
    final s = c.snapshot;
    if (s == null || !_draftMatches(s) || !_canMutate) return;
    var submit = false;
    setState(() {
      _prepareIntent(intent, showContext: false);
      _dropContinuation = fromDrop;
      submit = fromDrop && _completeDrop(intent) && _readyToSubmit(s);
      if (submit)
        _popup = null;
      else
        _openPopup('review');
    });
    if (submit) _submit(s);
  }

  bool _readyToSubmit(AuthoritativeSnapshot s) =>
      _canMutate &&
      _draftMatches(s) &&
      _draftCardsVisible(s) &&
      _responseAllowed(_action, s) &&
      _formationReady(s) &&
      _combatChoiceReady() &&
      _parametersReady(s) &&
      commandTiming(_action, s.board, s.online) == CommandTiming.available;

  List<String> _seriesNumberIds(
    AuthoritativeSnapshot s,
    TableDropTarget target,
  ) => [
    for (final card in (s.board['cards'] as List? ?? []).cast<Map>())
      if (card['controller'] == target.seat &&
          card['zone'] == 'series' &&
          card['card']['suit'] == target.suit &&
          (card['card']['rank'] as int) >= 2 &&
          (card['card']['rank'] as int) <= 10)
        card['card']['id'] as String,
  ];

  void _activateTarget(TableDropTarget target) {
    final s = c.snapshot;
    if (s == null || !_canMutate) return;
    if (!_intentPrepared && _draftMatches(s) && target.seat == s.seat) {
      final members = target.zone == TableDropZone.series
          ? _seriesNumberIds(s, target)
          : _formationMembers(s, target.formationId ?? '');
      if (_selected.any(members.contains)) {
        setState(() => _chooseMoves());
        return;
      }
    }
    final activateSource = !_intentPrepared && _selected.isEmpty;
    _targetTapped(target);
    if (activateSource && _selected.isNotEmpty) {
      setState(() => _chooseMoves());
    }
  }

  void _targetTapped(TableDropTarget target) {
    final s = c.snapshot;
    if (s == null || !_canMutate) return;
    if (!_draftMatches(s)) {
      setState(() => _openPopup('review'));
      return;
    }
    setState(() {
      if (_intentPrepared) {
        if (_flags['protection'] == true && target.seat == s.seat) {
          if (target.zone == TableDropZone.formation) {
            _protectFormation = target.formationId ?? '';
            _protectSuit = '';
          } else if (target.suit != null && target.suit != 'hearts') {
            _protectSuit = target.suit!;
            _protectFormation = '';
          }
        } else if (_action == 'attach' &&
            target.suit != null &&
            target.seat == s.seat) {
          _suit = target.suit!;
        } else if (target.zone == TableDropZone.series) {
          if (_action.contains('attack') &&
              !_action.contains('bomb') &&
              target.seat != s.seat &&
              !((_action == 'infiltrator-attack' ||
                      _flags['infiltrator'] == true) &&
                  target.suit == 'diamonds')) {
            final targets = _seriesNumberIds(s, target);
            if (targets.isNotEmpty) {
              _sets['targets'] = targets.toSet();
              _suit = target.suit!;
              _selected
                ..clear()
                ..addAll(_sourceCards)
                ..addAll(_sets.values.expand((ids) => ids));
              final ace = _fields['ace']?.text ?? '';
              if (ace.isNotEmpty) _selected.add(ace);
            }
          }
        } else if (target.cardId != null) {
          _selectCard(target.cardId!);
        } else if (target.zone == TableDropZone.seat &&
            target.seat != null &&
            target.seat != s.seat) {
          field(_action == 'ponzi' ? 'value' : 'target_seat').text =
              '${target.seat}';
        } else if (_action == 'attach' && target.suit != null) {
          _suit = target.suit!;
        }
        _openPopup('review');
      } else if (_selected.isNotEmpty) {
        _chooseMoves(target);
      } else if (target.zone == TableDropZone.series && target.seat == s.seat) {
        final allowed = selectableCardIds(s.board, s.online);
        final members = _seriesNumberIds(s, target).where(allowed.contains);
        if (members.isNotEmpty) {
          _captureDraft();
          _selected.addAll(members);
        }
      } else if (target.formationId != null) {
        final f = (s.board['formations'] as List? ?? [])
            .cast<Map>()
            .where(
              (f) => f['id'] == target.formationId && f['controller'] == s.seat,
            )
            .firstOrNull;
        if (f != null) {
          _captureDraft();
          _selected.addAll((f['spec']['cards'] as List).cast<String>());
        }
      }
    });
  }

  bool _responseAllowed(String type, AuthoritativeSnapshot s) {
    const responses = {
      'compensation-response',
      'numerical-defense',
      'advancement-defense',
      'confinement',
      'inflation',
      'negotiate',
    };
    if (!responses.contains(type)) return true;
    final pending = (s.online['rules_context'] as Map?)?['pending'];
    final types = pending is Map ? pending['response_types'] : null;
    return types is! List || types.contains(type);
  }

  Widget _gameOptions(AuthoritativeSnapshot s) => PopupMenuButton<String>(
    tooltip: l10n.onlineGameOptions,
    onSelected: (type) => setState(() {
      if (type == 'records' || type == 'help') {
        _openPopup(type);
        return;
      }
      _clearDraft();
      _prepareIntent(
        CardPlayIntent(
          type: type,
          seed: const {},
          explanation: l10n.onlineReviewBeforeConfirm(
            label(type).toLowerCase(),
          ),
        ),
      );
    }),
    itemBuilder: (_) => [
      if (canDeclareCoup(s))
        PopupMenuItem<String>(
          enabled: false,
          child: _coupCards(s, closeMenu: true),
        ),
      PopupMenuItem(value: 'records', child: Text(l10n.onlineRecords)),
      PopupMenuItem(value: 'help', child: Text(l10n.onlineCardGestures)),
      for (final type
          in s.phase == 'playing'
              ? [
                  'promise-offer',
                  'voluntary-transfer',
                  'declare-ordinary',
                  'nullify',
                ]
              : ['finish-settlement', 'voluntary-transfer', 'nullify'])
        PopupMenuItem(
          value: type,
          enabled:
              _canMutate &&
              commandTiming(type, s.board, s.online) == CommandTiming.available,
          child: Text(label(type)),
        ),
    ],
    icon: const Icon(Icons.more_horiz),
  );

  String label(String name) => switch (name) {
    'accept' => CgmsLocalizations.of(context).onlineAccept,
    'accept-offer' => CgmsLocalizations.of(context).onlineAcceptOffer,
    'ace' => CgmsLocalizations.of(context).onlineAce,
    'advancement-defense' => CgmsLocalizations.of(
      context,
    ).onlineAdvancementDefense,
    'amount' => CgmsLocalizations.of(context).onlineAmount,
    'array' => CgmsLocalizations.of(context).onlineArray,
    'attach' => CgmsLocalizations.of(context).onlineAttach,
    'attack' => CgmsLocalizations.of(context).onlineAttack,
    'award_id' => CgmsLocalizations.of(context).onlineAwardId,
    'baron-attack' => CgmsLocalizations.of(context).onlineBaronAttack,
    'baron-bomb-attack' => CgmsLocalizations.of(context).onlineBaronBombAttack,
    'barricade-sacrifice' => CgmsLocalizations.of(
      context,
    ).onlineBarricadeSacrifice,
    'bomb-attack' => CgmsLocalizations.of(context).onlineBombAttack,
    'boolean' => CgmsLocalizations.of(context).onlineBoolean,
    'code' => CgmsLocalizations.of(context).onlineCode,
    'compensation' => CgmsLocalizations.of(context).onlineCompensation,
    'compensation-response' => CgmsLocalizations.of(
      context,
    ).onlineCompensationResponse,
    'condition' => CgmsLocalizations.of(context).onlineCondition,
    'confinement' => CgmsLocalizations.of(context).onlineConfinement,
    'const' => CgmsLocalizations.of(context).onlineConst,
    'coup' => CgmsLocalizations.of(context).onlineCoup,
    'debt_id' => CgmsLocalizations.of(context).onlineDebtId,
    'decision' => CgmsLocalizations.of(context).onlineDecision,
    'decision-closed' => CgmsLocalizations.of(context).onlineDecisionClosed,
    'decision_id' => CgmsLocalizations.of(context).onlineDecisionId,
    'declare-ordinary' => CgmsLocalizations.of(context).onlineDeclareOrdinary,
    'decline-offer' => CgmsLocalizations.of(context).onlineDeclineOffer,
    'departure-choice' => CgmsLocalizations.of(context).onlineDepartureChoice,
    'dexter-assassination' => CgmsLocalizations.of(
      context,
    ).onlineDexterAssassination,
    'end-turn' => CgmsLocalizations.of(context).onlineEndTurn,
    'exile' => CgmsLocalizations.of(context).onlineExile,
    'expose-unused-ace' => CgmsLocalizations.of(context).onlineExposeUnusedAce,
    'fate' => CgmsLocalizations.of(context).onlineFate,
    'finish-settlement' => CgmsLocalizations.of(context).onlineFinishSettlement,
    'forgive' => CgmsLocalizations.of(context).onlineForgive,
    'formation' => CgmsLocalizations.of(context).onlineFormation,
    'formation_id' => CgmsLocalizations.of(context).onlineFormationId,
    'infiltrator' => CgmsLocalizations.of(context).onlineInfiltrator,
    'infiltrator-attack' => CgmsLocalizations.of(
      context,
    ).onlineInfiltratorAttack,
    'inflation' => CgmsLocalizations.of(context).onlineInflation,
    'integer' => CgmsLocalizations.of(context).onlineInteger,
    'items' => CgmsLocalizations.of(context).onlineItems,
    'justice' => CgmsLocalizations.of(context).onlineJustice,
    'kidnapper' => CgmsLocalizations.of(context).onlineKidnapper,
    'kidnapper-outcome' => CgmsLocalizations.of(context).onlineKidnapperOutcome,
    'main-inflation' => CgmsLocalizations.of(context).onlineMainInflation,
    'mode' => CgmsLocalizations.of(context).onlineMode,
    'negotiate' => CgmsLocalizations.of(context).onlineNegotiate,
    'nullify' => CgmsLocalizations.of(context).onlineNullify,
    'numerical-defense' => CgmsLocalizations.of(context).onlineNumericalDefense,
    'object' => CgmsLocalizations.of(context).onlineObject,
    'offer' => CgmsLocalizations.of(context).onlineOffer,
    'offer_id' => CgmsLocalizations.of(context).onlineOfferId,
    'open-formation' => CgmsLocalizations.of(context).onlineOpenFormation,
    'open-series' => CgmsLocalizations.of(context).onlineOpenSeries,
    'pass' => CgmsLocalizations.of(context).onlinePass,
    'payment' => CgmsLocalizations.of(context).onlinePayment,
    'pending_action_id' => CgmsLocalizations.of(context).onlinePendingActionId,
    'ponzi' => CgmsLocalizations.of(context).onlinePonzi,
    'price' => CgmsLocalizations.of(context).onlinePrice,
    'promise-accept' => CgmsLocalizations.of(context).onlinePromiseAccept,
    'promise-final-answer' => CgmsLocalizations.of(
      context,
    ).onlinePromiseFinalAnswer,
    'promise-final-offer' => CgmsLocalizations.of(
      context,
    ).onlinePromiseFinalOffer,
    'promise-offer' => CgmsLocalizations.of(context).onlinePromiseOffer,
    'promise-pay' => CgmsLocalizations.of(context).onlinePromisePay,
    'promise-refuse' => CgmsLocalizations.of(context).onlinePromiseRefuse,
    'promise_id' => CgmsLocalizations.of(context).onlinePromiseId,
    'properties' => CgmsLocalizations.of(context).onlineProperties,
    'purchase' => CgmsLocalizations.of(context).onlinePurchase,
    'quantity' => CgmsLocalizations.of(context).onlineQuantity,
    'recipient' => CgmsLocalizations.of(context).onlineRecipient,
    'required' => CgmsLocalizations.of(context).onlineRequired,
    'return-loan' => CgmsLocalizations.of(context).onlineReturnLoan,
    'revision' => CgmsLocalizations.of(context).onlineRevision,
    'richer-sacrifice' => CgmsLocalizations.of(context).onlineRicherSacrifice,
    'ringleader' => CgmsLocalizations.of(context).onlineRingleader,
    'selection' => CgmsLocalizations.of(context).onlineSelection,
    'string' => CgmsLocalizations.of(context).onlineString,
    'suit' => CgmsLocalizations.of(context).onlineSuit,
    'take-back' => CgmsLocalizations.of(context).onlineTakeBack,
    'target_seat' => CgmsLocalizations.of(context).onlineTargetSeat,
    'targets' => CgmsLocalizations.of(context).onlineTargets,
    'terms' => CgmsLocalizations.of(context).onlineTerms,
    'threshold' => CgmsLocalizations.of(context).onlineThreshold,
    'type' => CgmsLocalizations.of(context).onlineType,
    'value' => CgmsLocalizations.of(context).onlineValue,
    'voluntary-transfer' => CgmsLocalizations.of(
      context,
    ).onlineVoluntaryTransfer,
    'withdraw-offer' => CgmsLocalizations.of(context).onlineWithdrawOffer,
    _ =>
      name
          .split(RegExp('[-_]'))
          .map((v) => v.isEmpty ? v : '${v[0].toUpperCase()}${v.substring(1)}')
          .join(' '),
  };

  @override
  Widget build(BuildContext context) {
    final s = c.snapshot;
    final table = s == null
        ? null
        : projectTable(
            s,
            botLabel: l10n.onlineBot,
            departureLabel: l10n.onlineRoundDeparture,
            roundWaitingLabel: l10n.onlineRoundWaiting,
            serverProgressLabel: l10n.onlineServerProgress,
            connected: c.connected,
            usedClubStatus: l10n.onlineClubUsedThisTurn,
          );
    // Retain only the inspection identity. Every frame uses the current
    // authorized card and controller style; a hidden card or new game closes it.
    final inspected = table != null && _inspectedGame == s?.gameId
        ? table.hand
              .followedBy(table.seats.expand((seat) => seat.exposed))
              .where((card) => card.id == _inspectedId)
              .firstOrNull
        : null;
    if (inspected == null) {
      _inspectedId = null;
      _inspectedGame = null;
    }
    if (s != null &&
        _protectFormation.isNotEmpty &&
        (_formationKind != 'great-people' ||
            _flags['substitute'] == true ||
            !(s.board['formations'] as List? ?? []).any(
              (f) =>
                  f['id'] == _protectFormation &&
                  f['controller'] == s.seat &&
                  f['id'] != _fields['formation_id']?.text &&
                  !['people', 'great-people'].contains(f['spec']['kind']),
            ))) {
      _protectFormation = '';
    }
    final content = s == null
        ? _setup()
        : CgmsAdaptiveTable(
            data: table!,
            actionsIdentity: (
              s.gameId,
              s.phase,
              table.round,
              s.online['round_closed'] == true,
            ),
            header: s.phase == 'playing'
                ? null
                : Semantics(liveRegion: true, child: Text(table.turnLabel)),
            showArt: _art,
            drawFeedback: c.drawFeedback,
            selectedCardIds: _selected,
            draggableCardIds: _canMutate
                ? draggableCardIds(s.board, s.online)
                : const {},
            onCardDrop: _canMutate ? _onCardDrop : null,
            dragIdentity: _snapshotIdentity(s),
            canDrop: _canDrop,
            isSuggestedTarget: (target) => _suggestedTarget(s, target),
            dropPreview: (drag, target) => _dropPreview(s, drag, target),
            cardRoles: _cardRoles,
            onTargetTap: _targetTapped,
            onActivateCard: _activateCard,
            onActivateTarget: _activateTarget,
            onSelect: _selectCard,
            onInspect: (card) {
              setState(() {
                _inspectedId = card.id;
                _inspectedGame = s.gameId;
              });
              setState(() => _openPopup('inspect'));
            },
            onCancelSelection: () => _dismissPopup(clear: true),
            publicChooserTools: canDeclareCoup(s)
                ? _coupCards(s, closeMenu: true)
                : null,
          );
    return Theme(
      data: Theme.of(context).copyWith(visualDensity: VisualDensity.standard),
      child: Directionality(
        textDirection: _rtl ? TextDirection.rtl : TextDirection.ltr,
        child: MediaQuery(
          data: MediaQuery.of(context).copyWith(
            textScaler: _scale == 1
                ? MediaQuery.textScalerOf(context)
                : TextScaler.linear(_scale),
          ),
          child: CallbackShortcuts(
            bindings: {
              if (s != null)
                const SingleActivator(LogicalKeyboardKey.escape): () =>
                    _dismissPopup(clear: true),
            },
            child: Scaffold(
              appBar: AppBar(
                title: Text(CgmsLocalizations.of(context).onlineCGMSOnline),
                actions: [
                  if (c.account != null)
                    IconButton(
                      tooltip: l10n.economyTitle,
                      onPressed: _openEconomy,
                      icon: const Icon(Icons.account_balance_wallet_outlined),
                    ),
                  IconButton(
                    tooltip: l10n.onlineRefreshConnection,
                    onPressed: !_entryReady
                        ? null
                        : s == null
                        ? _restore
                        : () => _run(c.reconnect),
                    icon: const Icon(Icons.sync),
                  ),
                  PopupMenuButton<String>(
                    tooltip: l10n.onlineDisplaySettings,
                    onSelected: (v) => setState(() {
                      if (v == 'art') _art = !_art;
                      if (v == 'rtl') _rtl = !_rtl;
                      if (v == 'text') _scale = _scale == 1 ? 2 : 1;
                    }),
                    itemBuilder: (_) => [
                      if (s != null && canDeclareCoup(s))
                        PopupMenuItem<String>(
                          enabled: false,
                          child: _coupCards(s, closeMenu: true),
                        ),
                      CheckedPopupMenuItem(
                        value: 'art',
                        checked: _art,
                        child: Text(
                          CgmsLocalizations.of(context).onlineArtwork,
                        ),
                      ),
                      CheckedPopupMenuItem(
                        value: 'rtl',
                        checked: _rtl,
                        child: Text(
                          CgmsLocalizations.of(context).onlineRightToLeft,
                        ),
                      ),
                      CheckedPopupMenuItem(
                        value: 'text',
                        checked: _scale == 2,
                        child: Text(
                          CgmsLocalizations.of(context).online200Text,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              body: SafeArea(
                child: LayoutBuilder(
                  builder: (context, box) => Column(
                    children: [
                      ConstrainedBox(
                        constraints: BoxConstraints(
                          maxHeight: box.maxHeight * .30,
                        ),
                        child: SingleChildScrollView(
                          key: const ValueKey('online-notices'),
                          child: Column(
                            children: [
                              if (c.busy || _sessionBusy)
                                const LinearProgressIndicator(),
                              if (c.sessionLoading)
                                Semantics(
                                  liveRegion: true,
                                  child: Padding(
                                    padding: const EdgeInsets.all(
                                      CgmsSpacing.controlGap,
                                    ),
                                    child: Text(
                                      l10n.onlineSessionConnecting(
                                        c.sessionAttempt,
                                        OnlineClient.recoveryLimit,
                                      ),
                                    ),
                                  ),
                                ),
                              if (_displayError case final code?)
                                Semantics(
                                  liveRegion: true,
                                  child: Padding(
                                    padding: const EdgeInsets.all(
                                      CgmsSpacing.controlGap,
                                    ),
                                    child: CgmsPanel(
                                      accent: CgmsColors.error,
                                      child: Text(
                                        _error(code),
                                        key: const ValueKey('online-error'),
                                      ),
                                    ),
                                  ),
                                ),
                              if (c.unknownOperation != null)
                                Wrap(
                                  children: [
                                    Text(
                                      l10n.onlineOutcomeUnknownKeepThisPageOpenWhileReconciling,
                                    ),
                                    TextButton(
                                      onPressed: c.busy
                                          ? null
                                          : () => _run(c.reconcile),
                                      child: Text(
                                        CgmsLocalizations.of(
                                          context,
                                        ).onlineCheckOperation,
                                      ),
                                    ),
                                    TextButton(
                                      onPressed: c.busy
                                          ? null
                                          : () => _run(c.retryUnknownOperation),
                                      child: Text(
                                        CgmsLocalizations.of(
                                          context,
                                        ).onlineRetryOriginalOperation,
                                      ),
                                    ),
                                  ],
                                ),
                              if (s != null && !c.connected)
                                Semantics(
                                  liveRegion: true,
                                  child: CgmsNotice(
                                    title: l10n.onlineConnectionLost,
                                    message: c.reconnecting
                                        ? l10n.onlineReconnectHelp
                                        : l10n.onlineReconnectStopped,
                                    icon: Icons.wifi_off,
                                    color: CgmsColors.action,
                                  ),
                                ),
                            ],
                          ),
                        ),
                      ),
                      Expanded(
                        // Recovery notices move this viewport without replacing
                        // its cards. Keep one semantic transform boundary so
                        // browser accessibility geometry moves with the board.
                        child: Semantics(
                          identifier: 'online-board-context',
                          container: true,
                          explicitChildNodes: true,
                          child: s == null
                              ? content
                              : LayoutBuilder(
                                  builder: (context, bounds) => Stack(
                                    children: [
                                      Positioned.fill(child: content),
                                      if (_popup != null)
                                        _contextPopup(s, inspected, bounds),
                                    ],
                                  ),
                                ),
                        ),
                      ),
                      if (s != null)
                        ConstrainedBox(
                          constraints: BoxConstraints(
                            minWidth: box.maxWidth,
                            maxHeight: box.maxHeight * .28,
                          ),
                          // Resolve the dock's exact geometry while the display
                          // menu temporarily excludes underlying semantics.
                          child: IntrinsicHeight(
                            child: Semantics(
                              identifier: 'game-footer',
                              container: true,
                              explicitChildNodes: true,
                              child: SingleChildScrollView(
                                child: Padding(
                                  padding: const EdgeInsets.all(8),
                                  child: _utilities(s),
                                ),
                              ),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _openEconomy() async {
    if (c.snapshot != null) {
      setState(() => _openPopup('economy'));
      return;
    }
    await showDialog<void>(
      context: context,
      builder: (dialogContext) => Theme(
        data: Theme.of(context).copyWith(visualDensity: VisualDensity.standard),
        child: Directionality(
          textDirection: _rtl ? TextDirection.rtl : TextDirection.ltr,
          child: MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: _scale == 1
                  ? MediaQuery.textScalerOf(context)
                  : TextScaler.linear(_scale),
            ),
            child: Dialog(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 520),
                child: AnimatedBuilder(
                  animation: c,
                  builder: (context, _) => Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (c.snapshot != null && canDeclareCoup(c.snapshot!))
                        _coupCards(c.snapshot!, closeMenu: true),
                      Flexible(
                        child: SingleChildScrollView(
                          child: EconomyPanel(client: c),
                        ),
                      ),
                      TextButton(
                        onPressed: () => Navigator.of(dialogContext).pop(),
                        child: Text(
                          MaterialLocalizations.of(context).closeButtonLabel,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  String _error(String code) => switch (code) {
    'UNAUTHENTICATED' => l10n.onlineSignInOrContinueAsAGuest,
    'RECOVERY_EXHAUSTED' => l10n.onlineRecoveryExhausted,
    'INVALID_RESPONSE' => l10n.onlineInvalidResponse,
    'STATE_CONFLICT' =>
      l10n.onlineTheTableChangedReviewYourSelectionBeforeSubmittingAgain,
    'OUTCOME_UNKNOWN' =>
      l10n.onlineYourActionMayHaveSucceededCheckTheOriginalOperation,
    'INVALID_COMMAND' =>
      l10n.onlineThisActionIsNotPermittedWithTheseChoicesReview,
    'FORBIDDEN' => l10n.onlineThisSessionCannotAccessThatRoomOrMatch,
    'ROOM_REJECTED' => l10n.onlineTheRoomCouldNotAcceptThisRequestCheckIts,
    'ALLOWANCE_EXHAUSTED' => l10n.economyAdmissionWaiting,
    _ => l10n.onlineConnectionUnavailableRetryWhenTheAuthorityIsReachable,
  };

  Widget _setup() => LayoutBuilder(
    builder: (context, box) {
      final desktop = box.maxWidth >= 1050;
      final inset = box.maxWidth < 600 ? 16.0 : CgmsSpacing.pageInset;
      final tasks = c.account == null
          ? _entryPanel()
          : c.room == null
          ? _roomSetup()
          : _roomLobby();
      return SingleChildScrollView(
        key: const ValueKey('online-lobby-scroll'),
        padding: EdgeInsets.all(inset),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 1120),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (desktop)
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(flex: 4, child: _lobbyIntroduction()),
                      const SizedBox(width: 32),
                      Expanded(flex: 6, child: tasks),
                    ],
                  )
                else ...[
                  _lobbyIntroduction(),
                  const SizedBox(height: CgmsSpacing.sectionGap),
                  tasks,
                ],
                if (c.account != null) ...[
                  const SizedBox(height: CgmsSpacing.sectionGap),
                  CgmsPanel(
                    padding: EdgeInsets.zero,
                    child: ExpansionTile(
                      key: const PageStorageKey('lobby-economy'),
                      controller: _allowanceController,
                      maintainState: true,
                      title: Text(l10n.economyTitle),
                      leading: const Icon(
                        Icons.account_balance_wallet_outlined,
                      ),
                      children: [
                        AnimatedBuilder(
                          animation: _allowanceController,
                          builder: (context, child) => ExcludeFocus(
                            excluding: !_allowanceController.isExpanded,
                            child: child!,
                          ),
                          child: EconomyPanel(client: c),
                        ),
                      ],
                    ),
                  ),
                ],
              ],
            ),
          ),
        ),
      );
    },
  );

  // DropdownButton defaults to title typography; options are form values.
  // Intrinsic menu rows keep wrapped accessibility text above the 48px minimum.
  Widget _dropdown<T>({
    Key? key,
    required T? initialValue,
    required InputDecoration decoration,
    required List<DropdownMenuItem<T>> items,
    required ValueChanged<T?>? onChanged,
  }) => Builder(
    builder: (fieldContext) {
      final media = MediaQuery.of(fieldContext);
      final direction = Directionality.of(fieldContext);
      return DropdownButtonFormField<T>(
        key: key,
        initialValue: initialValue,
        decoration: decoration,
        style: Theme.of(
          fieldContext,
        ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.normal),
        isExpanded: true,
        isDense: false,
        itemHeight: null,
        menuMaxHeight: (media.size.height * 0.6).clamp(48.0, 400.0),
        // Only the chosen value determines the closed field's natural height.
        // DropdownButton otherwise sizes its IndexedStack to the longest option.
        selectedItemBuilder: (context) => [
          for (final item in items)
            if (item.value == initialValue)
              ConstrainedBox(
                constraints: const BoxConstraints(minHeight: 24),
                child: Align(
                  alignment: AlignmentDirectional.centerStart,
                  child: item.child,
                ),
              )
            else
              const SizedBox.shrink(),
        ],
        borderRadius: BorderRadius.circular(CgmsShape.control),
        // Flutter's popup route captures themes, but not the local text-size
        // and RTL settings. Carry both into each naturally sized popup row.
        items: [
          for (final item in items)
            DropdownMenuItem<T>(
              key: item.key,
              value: item.value,
              enabled: item.enabled,
              onTap: item.onTap,
              alignment: item.alignment.resolve(direction),
              child: MediaQuery(
                data: media,
                child: Directionality(
                  textDirection: direction,
                  child: item.child,
                ),
              ),
            ),
        ],
        onChanged: onChanged,
      );
    },
  );

  Widget _lobbyHeading(String title, {bool introduction = false}) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Semantics(
      header: true,
      child: Text(
        title,
        textAlign: TextAlign.center,
        style: introduction
            ? Theme.of(context).textTheme.headlineMedium
            : Theme.of(context).textTheme.headlineSmall?.copyWith(fontSize: 24),
      ),
    ),
  );

  Widget _lobbyCopy(String text) => Center(
    child: ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 520),
      child: Text(
        text,
        textAlign: TextAlign.center,
        style: Theme.of(
          context,
        ).textTheme.bodyMedium?.copyWith(color: CgmsColors.muted),
      ),
    ),
  );

  Widget _lobbyIntroduction() => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      Text(
        l10n.onlineCGMSOnline,
        textAlign: TextAlign.center,
        style: Theme.of(context).textTheme.labelMedium?.copyWith(
          color: CgmsColors.muted,
          letterSpacing: 0.8,
        ),
      ),
      const SizedBox(height: 8),
      _lobbyHeading(
        c.room == null ? l10n.onlineWelcomeTitle : l10n.onlineYourRoom,
        introduction: true,
      ),
      _lobbyCopy(
        c.room == null
            ? l10n.onlineWelcomeBody
            : l10n.onlineRoomConfiguration(
                '${c.room!['games']}',
                '${c.room!['capacity']}',
              ),
      ),
      const SizedBox(height: CgmsSpacing.sectionGap),
      CgmsPanel(
        child: Semantics(
          image: true,
          label: l10n.onlineWelcomeCards,
          child: ExcludeSemantics(
            child: LayoutBuilder(
              builder: (context, box) {
                final width = ((box.maxWidth - 12) / 3).clamp(48.0, 112.0);
                return Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    for (final card in const [
                      VisibleCard(
                        id: 'intro-clubs',
                        rank: '8',
                        suit: CardSuit.clubs,
                        copy: 1,
                      ),
                      VisibleCard(
                        id: 'intro-hearts',
                        rank: 'Q',
                        suit: CardSuit.hearts,
                        copy: 1,
                      ),
                      VisibleCard(
                        id: 'intro-diamonds',
                        rank: '4',
                        suit: CardSuit.diamonds,
                        copy: 1,
                      ),
                    ])
                      CgmsCard(
                        card: card,
                        width: width,
                        thumbnail: true,
                        showArt: _art,
                      ),
                  ],
                );
              },
            ),
          ),
        ),
      ),
    ],
  );

  Widget _entryPanel() => CgmsPanel(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        _lobbyHeading(l10n.onlineEntryTitle),
        _lobbyCopy(l10n.onlineEntryBody),
        const SizedBox(height: CgmsSpacing.sectionGap),
        FilledButton.icon(
          onPressed: !_entryReady
              ? null
              : () => _sessionRequest(() async {
                  await c.guest();
                  await _loadLocation();
                }),
          icon: const Icon(Icons.login),
          label: Text(l10n.onlineContinueAsGuest),
        ),
        const SizedBox(height: CgmsSpacing.sectionGap),
        const Divider(),
        const SizedBox(height: CgmsSpacing.controlGap),
        Text(
          l10n.onlineAccountSignIn,
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: CgmsSpacing.controlGap),
        AutofillGroup(
          child: Column(
            children: [
              TextField(
                controller: _username,
                decoration: InputDecoration(labelText: l10n.onlineUsername),
                autocorrect: false,
                enableSuggestions: false,
                autofillHints: const [AutofillHints.username],
                textInputAction: TextInputAction.next,
              ),
              const SizedBox(height: CgmsSpacing.controlGap),
              TextField(
                controller: _password,
                decoration: InputDecoration(labelText: l10n.onlinePassword),
                obscureText: true,
                autocorrect: false,
                enableSuggestions: false,
                autofillHints: const [AutofillHints.password],
              ),
            ],
          ),
        ),
        const SizedBox(height: CgmsSpacing.controlGap),
        Wrap(
          alignment: WrapAlignment.center,
          runSpacing: CgmsSpacing.controlGap,
          spacing: CgmsSpacing.controlGap,
          children: [
            for (final action in ['login', 'register'])
              TextButton(
                onPressed: !_entryReady
                    ? null
                    : () => _sessionRequest(() async {
                        await c.authenticate(
                          action,
                          _username.text,
                          _password.text,
                        );
                        await _loadLocation();
                      }),
                child: Text(label(action)),
              ),
          ],
        ),
      ],
    ),
  );

  Widget _roomSetup() => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      _lobbyHeading(l10n.onlineSetupTitle),
      _lobbyCopy(l10n.onlineSetupBody),
      const SizedBox(height: CgmsSpacing.sectionGap),
      if (c.hasPendingRoomOperation) ...[
        FilledButton(
          onPressed: c.busy ? null : () => _run(c.retryPendingRoomOperation),
          child: Text(l10n.onlineRecoverRoomOperation),
        ),
        const SizedBox(height: CgmsSpacing.controlGap),
      ],
      CgmsPanel(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _lobbyHeading(l10n.onlineCreateRoom),
            _lobbyCopy(l10n.onlineCreateAThreeOrFourPlayerRoomMatchLength),
            const SizedBox(height: CgmsSpacing.sectionGap),
            _dropdown<int>(
              initialValue: _capacity,
              decoration: InputDecoration(labelText: l10n.onlinePlayers),
              items: [
                for (final n in [3, 4])
                  DropdownMenuItem(value: n, child: Text('$n')),
              ],
              onChanged: _roomReady
                  ? (n) => setState(() => _capacity = n!)
                  : null,
            ),
            const SizedBox(height: CgmsSpacing.controlGap),
            TextFormField(
              key: const PageStorageKey('room-games'),
              controller: _gamesInput,
              decoration: InputDecoration(
                labelText: l10n.onlineGamesPerMatch,
                errorText: _validateRoom && (_games < 1 || _games > 100)
                    ? l10n.onlineRoomGamesRange
                    : null,
              ),
              enabled: _roomReady,
              keyboardType: TextInputType.number,
              onChanged: (v) => setState(() => _games = int.tryParse(v) ?? 0),
            ),
            const SizedBox(height: CgmsSpacing.sectionGap),
            _lobbyCopy(l10n.onlineBotSetupHelp),
            const SizedBox(height: CgmsSpacing.controlGap),
            _dropdown<String>(
              key: const ValueKey('bot-difficulty'),
              initialValue: _botDifficulty,
              decoration: InputDecoration(labelText: l10n.onlineBotDifficulty),
              items: [
                for (final level in ['beginner', 'standard', 'advanced'])
                  DropdownMenuItem(value: level, child: Text(_botLabel(level))),
              ],
              onChanged: _roomReady
                  ? (value) => setState(() => _botDifficulty = value!)
                  : null,
            ),
            const SizedBox(height: CgmsSpacing.sectionGap),
            FilledButton(
              onPressed: _roomReady ? () => _createRoom(withBots: true) : null,
              child: Text(l10n.onlinePlayAgainstBots),
            ),
            const SizedBox(height: CgmsSpacing.controlGap),
            OutlinedButton(
              onPressed: _roomReady ? () => _createRoom(withBots: false) : null,
              child: Text(l10n.onlineCreateRoom),
            ),
          ],
        ),
      ),
      const SizedBox(height: CgmsSpacing.sectionGap),
      CgmsPanel(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _lobbyHeading(l10n.onlineJoinRoom),
            _lobbyCopy(l10n.onlineJoinHelp),
            const SizedBox(height: CgmsSpacing.sectionGap),
            TextField(
              controller: _room,
              decoration: InputDecoration(labelText: l10n.onlineRoomCode),
              autocorrect: false,
              enableSuggestions: false,
              textInputAction: TextInputAction.next,
            ),
            const SizedBox(height: CgmsSpacing.controlGap),
            TextField(
              controller: _invitation,
              decoration: InputDecoration(labelText: l10n.onlineInvitation),
              autocorrect: false,
              enableSuggestions: false,
            ),
            const SizedBox(height: CgmsSpacing.sectionGap),
            FilledButton.tonal(
              onPressed: _roomReady
                  ? () => _run(() => c.joinRoom(_room.text, _invitation.text))
                  : null,
              child: Text(l10n.onlineJoinRoom),
            ),
          ],
        ),
      ),
    ],
  );

  String _botLabel(String difficulty) => switch (difficulty) {
    'beginner' => l10n.onlineBotBeginner,
    'standard' => l10n.onlineBotStandard,
    'advanced' => l10n.onlineBotAdvanced,
    _ => l10n.onlineBot,
  };

  Future<void> _createRoom({required bool withBots}) async {
    if (!_roomReady) return;
    setState(() => _validateRoom = true);
    if (_games < 1 || _games > 100) return;
    await _run(
      () => c.createRoom(
        _capacity,
        _games,
        botDifficulty: withBots ? _botDifficulty : null,
      ),
    );
  }

  Map? _roomMember(int seat) => (c.room?['members'] as List? ?? [])
      .cast<Map>()
      .where((member) => member['seat'] == seat)
      .firstOrNull;

  Widget _roomLobby() => CgmsPanel(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        CgmsSectionTitle(title: l10n.onlinePlayers),
        for (var seat = 1; seat <= (c.room!['capacity'] as int); seat++)
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: Icon(
              _roomMember(seat)?['bot'] == true
                  ? Icons.smart_toy_outlined
                  : _roomMember(seat) != null
                  ? Icons.person_outline
                  : Icons.person_add_alt,
              color: CgmsColors.muted,
            ),
            title: Text(l10n.onlineSeatNumber('$seat')),
            subtitle: Text(
              _roomMember(seat)?['bot'] == true
                  ? l10n.onlineBotSeat(
                      _botLabel(c.room?['bot_difficulty'] as String? ?? ''),
                    )
                  : _roomMember(seat) != null
                  ? l10n.onlineRoomJoined
                  : l10n.onlineRoomEmptySeat,
            ),
          ),
        const Divider(),
        _shareCode(l10n.onlineRoomCode, c.room!['id'] as String),
        const SizedBox(height: CgmsSpacing.controlGap),
        Wrap(
          spacing: CgmsSpacing.controlGap,
          children: [
            TextButton.icon(
              onPressed: c.busy
                  ? null
                  : () => _run(() => c.refreshRoom(c.room!['id'] as String)),
              icon: const Icon(Icons.refresh),
              label: Text(l10n.onlineRefreshSeats),
            ),
            if (_ownsRoom && !_roomStarted && !_botRoom)
              TextButton.icon(
                onPressed: _roomReady
                    ? () => _run(() async {
                        _invitationResult = await c.invite();
                      })
                    : null,
                icon: const Icon(Icons.link),
                label: Text(l10n.onlineCreateInvitation),
              ),
          ],
        ),
        if (_invitationResult != null &&
            _ownsRoom &&
            !_roomStarted &&
            !_botRoom)
          _shareCode(l10n.onlineInvitation, _invitationResult!),
        if (!_roomStarted && (!_ownsRoom || !_roomFull)) ...[
          const SizedBox(height: CgmsSpacing.sectionGap),
          Semantics(
            liveRegion: true,
            child: Text(
              !_ownsRoom
                  ? l10n.onlineRoomOwnerStarts
                  : l10n.onlineRoomWaiting(
                      '${(c.room!['capacity'] as int) - (c.room!['members'] as List).length}',
                    ),
            ),
          ),
        ],
        if (c.hasPendingRoomOperation) ...[
          const SizedBox(height: CgmsSpacing.controlGap),
          OutlinedButton(
            onPressed: c.busy ? null : () => _run(c.retryPendingRoomOperation),
            child: Text(l10n.onlineRecoverRoomOperation),
          ),
        ],
        const SizedBox(height: CgmsSpacing.sectionGap),
        FilledButton(
          onPressed: _roomReady && (_roomStarted || (_ownsRoom && _roomFull))
              ? () => _run(
                  _roomStarted
                      ? () => c.watchMatch(c.room!['match_id'] as String)
                      : c.startMatch,
                )
              : null,
          child: Text(_roomStarted ? l10n.onlineEnterMatch : l10n.startMatch),
        ),
      ],
    ),
  );

  Widget _shareCode(String label, String value) => Row(
    key: ValueKey('share-$label'),
    children: [
      Expanded(child: Text(l10n.onlineCodeValue(label, value))),
      IconButton(
        tooltip: l10n.onlineCopyCode(label),
        onPressed: () async {
          var copied = false;
          try {
            await Clipboard.setData(ClipboardData(text: value));
            copied = true;
          } catch (_) {
            // Clipboard permission is independent of the online connection.
          }
          if (!mounted) return;
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(
                copied
                    ? l10n.onlineCodeCopied(label)
                    : l10n.onlineClipboardUnavailable,
              ),
            ),
          );
        },
        icon: const Icon(Icons.copy),
      ),
    ],
  );

  Future<void> _chooseDeparture(bool leave) async {
    if (!_canMutate || c.snapshot?.online['departure_pending'] != true) return;
    await _run(() => c.submit('departure-choice', {'accept': leave}));
  }

  Map<String, TableCardRole> get _cardRoles => {
    for (final id in _selected)
      id:
          (c.snapshot?.board['cards'] as List? ?? []).any(
            (p) => p['card']['id'] == id && p['controller'] != c.snapshot?.seat,
          )
          ? TableCardRole.target
          : TableCardRole.source,
    for (final id in _sets['targets'] ?? <String>{}) id: TableCardRole.target,
    for (final id in _sets['receive'] ?? <String>{}) id: TableCardRole.target,
    for (final id in _sets['payment'] ?? <String>{}) id: TableCardRole.payment,
    if (_action == 'dexter-assassination')
      for (final id in _sets['selection'] ?? <String>{})
        id: TableCardRole.payment,
    if (_action == 'barricade-sacrifice')
      for (final id in (_sets['selection'] ?? <String>{}).skip(1))
        id: TableCardRole.payment,
    if ((_fields['ace']?.text ?? '').isNotEmpty)
      _fields['ace']!.text: _action == 'justice'
          ? TableCardRole.payment
          : _action.contains('attack') ||
                const {
                  'numerical-defense',
                  'advancement-defense',
                }.contains(_action)
          ? TableCardRole.modifier
          : TableCardRole.source,
    if (_action.contains('attack'))
      for (final p in c.snapshot?.board['cards'] as List? ?? [])
        if (_selected.contains(p['card']['id']) &&
            p['controller'] == c.snapshot?.seat &&
            p['zone'] == 'attachment' &&
            p['allocation'] == 'clubs' &&
            p['card']['suit'] == 'clubs' &&
            (p['card']['rank'] == 11 && _flags['infiltrator'] == true ||
                p['card']['rank'] == 12 && _flags['exile'] == true))
          p['card']['id'] as String: TableCardRole.modifier,
  };

  bool _hasResponse(AuthoritativeSnapshot s) =>
      s.board['required_actor'] == s.seat &&
      s.board['window_id'] != null &&
      s.board['decision_id'] == null &&
      (s.board['decision_kind'] as String? ?? '').isEmpty;

  Widget _utilities(AuthoritativeSnapshot s) => Semantics(
    container: true,
    explicitChildNodes: true,
    child: Wrap(
      key: const ValueKey('online-utilities'),
      alignment: WrapAlignment.center,
      spacing: 8,
      runSpacing: 4,
      crossAxisAlignment: WrapCrossAlignment.center,
      children:
          [
                _gameOptions(s),
                if (_requiredActionsFor(s) != null)
                  TextButton(
                    onPressed: () => setState(() => _openPopup('required')),
                    child: Text(
                      s.phase == 'playing'
                          ? l10n.onlinePendingDecision
                          : l10n.onlineResultsSettlement,
                    ),
                  ),
                if (_hasResponse(s))
                  OutlinedButton(
                    onPressed: _canMutate
                        ? () => _run(
                            () => c.submit('pass', {
                              'pending_action_id': s.board['window_id'],
                            }),
                          )
                        : null,
                    child: Text(l10n.onlinePassResponse),
                  ),
                if (s.phase == 'playing' && s.online['round_closed'] != true)
                  FilledButton(
                    onPressed:
                        _canMutate &&
                            commandTiming('end-turn', s.board, s.online) ==
                                CommandTiming.available
                        ? () => _run(() => c.submit('end-turn', {}))
                        : null,
                    child: Text(l10n.onlineEndTurn),
                  ),
              ]
              .map(
                (child) => Semantics(
                  container: true,
                  explicitChildNodes: true,
                  child: child,
                ),
              )
              .toList(),
    ),
  );

  Widget _contextPopup(
    AuthoritativeSnapshot s,
    VisibleCard? inspected,
    BoxConstraints bounds,
  ) {
    final geometry = (
      bounds.maxWidth,
      bounds.maxHeight,
      _scale,
      _rtl,
      MediaQuery.textScalerOf(context).scale(16),
    );
    if (_popupGeometry != geometry) {
      _popupGeometry = geometry;
      final focused = FocusManager.instance.primaryFocus;
      if (_popupFocus.hasFocus && focused != null && focused != _popupFocus) {
        WidgetsBinding.instance.addPostFrameCallback((_) {
          // Reflow retains the control's focus node, so onFocusChange does not
          // run. Reveal its new geometry without replacing or moving focus.
          if (mounted &&
              _popup != null &&
              _popupFocus.hasFocus &&
              identical(FocusManager.instance.primaryFocus, focused) &&
              focused.context != null) {
            Scrollable.ensureVisible(
              focused.context!,
              alignmentPolicy: ScrollPositionAlignmentPolicy.keepVisibleAtStart,
            );
            Scrollable.ensureVisible(
              focused.context!,
              alignmentPolicy: ScrollPositionAlignmentPolicy.keepVisibleAtEnd,
            );
          }
        });
      }
    }
    final title = switch (_popup) {
      'inspect' => inspected?.label ?? l10n.onlineCardInvisible,
      'choices' => l10n.onlineChooseMove,
      'review' => label(_action),
      'required' =>
        s.phase == 'playing'
            ? l10n.onlinePendingDecision
            : l10n.onlineResultsSettlement,
      'help' => l10n.onlineCardGestures,
      'economy' => l10n.economyTitle,
      _ => l10n.onlineRecords,
    };
    return Positioned(
      left: 8,
      right: 8,
      bottom: 8,
      child: Align(
        alignment: Alignment.bottomCenter,
        child: ConstrainedBox(
          constraints: BoxConstraints(
            maxWidth: 560,
            maxHeight: bounds.maxHeight < 300
                ? (bounds.maxHeight - 16).clamp(48.0, 300.0)
                : bounds.maxHeight * .62,
          ),
          child: Material(
            key: const ValueKey('contextual-popup'),
            elevation: 16,
            color: CgmsColors.raised,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
              side: const BorderSide(color: CgmsColors.cyan),
            ),
            clipBehavior: Clip.antiAlias,
            // Keep this overlay's semantic children in one stable group. When
            // the board reflows, unrelated cards must not reorder its focused
            // controls in the browser accessibility tree.
            child: Semantics(
              container: true,
              explicitChildNodes: true,
              child: FocusTraversalGroup(
                child: Focus(
                  focusNode: _popupFocus,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Padding(
                              padding: const EdgeInsetsDirectional.only(
                                start: 16,
                              ),
                              child: Semantics(
                                header: true,
                                headingLevel: 2,
                                liveRegion: true,
                                child: Text(
                                  title,
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                  style: Theme.of(
                                    context,
                                  ).textTheme.titleMedium,
                                ),
                              ),
                            ),
                          ),
                          IconButton(
                            key: const ValueKey('close-card-context'),
                            tooltip: l10n.onlineCloseCardContext,
                            onPressed: _dismissPopup,
                            icon: const Icon(Icons.close),
                          ),
                        ],
                      ),
                      Flexible(
                        child: SingleChildScrollView(
                          key: ValueKey('context-scroll-$_popup'),
                          padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            spacing: 8,
                            children: [
                              // Same authorized cards, including while choosing
                              // another move. This also scrolls on short screens.
                              if (canDeclareCoup(s) &&
                                  !(_popup == 'review' && _action == 'coup'))
                                _coupCards(s),
                              ...switch (_popup) {
                                'inspect' => [
                                  if (inspected != null)
                                    Center(
                                      child: CgmsCard(
                                        key: const ValueKey(
                                          'online-inspected-card',
                                        ),
                                        card: inspected,
                                        width: (bounds.maxWidth - 64).clamp(
                                          48.0,
                                          220.0,
                                        ),
                                        showArt: _art,
                                      ),
                                    ),
                                ],
                                'choices' => [
                                  Text(_currentCardNames(s, _selected)),
                                  if (!_draftMatches(s))
                                    Text(
                                      l10n.onlineTheGameOrDecisionChangedClearThisDraftBefore,
                                    )
                                  else if (_choices.isEmpty)
                                    Text(l10n.onlineNoMoveDestination)
                                  else
                                    CardPlayChoices(
                                      intents: _choices,
                                      label: label,
                                      enabled: _canMutate && _draftMatches(s),
                                      onSelected: _chooseCardIntent,
                                    ),
                                  _boardButton(),
                                ],
                                'review' => _reviewContent(s),
                                'required' => _requiredContent(s),
                                'economy' => [EconomyPanel(client: c)],
                                'help' => [
                                  Text(l10n.onlineCardGestureHelp),
                                  ..._quotaContext(s),
                                ],
                                _ => [..._records(s)],
                              },
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _coupCards(AuthoritativeSnapshot s, {bool closeMenu = false}) {
    final queens = (s.board['cards'] as List? ?? [])
        .cast<Map>()
        .where(
          (p) =>
              p['controller'] == s.seat &&
              p['card']['rank'] == 12 &&
              ['hearts', 'diamonds'].contains(p['card']['suit']),
        )
        .toList();
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          for (final queen in queens)
            CgmsCard(
              card: projectCard(queen),
              width: 72,
              compact: true,
              boardPreview: true,
              showArt: _art,
              onTap: () => _selectCard(queen['card']['id'] as String),
              onActivate: () {
                if (closeMenu) Navigator.of(context).pop();
                setState(
                  () => _prepareIntent(
                    CardPlayIntent(
                      type: 'coup',
                      seed: {},
                      explanation: l10n.onlineCoupCardConfirmation,
                    ),
                  ),
                );
              },
            ),
        ],
      ),
    );
  }

  Widget _boardButton([String? role]) => TextButton.icon(
    onPressed: () {
      if (role != null) _selectionFor = role;
      _dismissPopup();
    },
    icon: const Icon(Icons.touch_app_outlined),
    label: Text(_boardGuidance(role)),
  );

  String _boardGuidance(String? role) => switch (role) {
    'attachment' => l10n.onlineChooseSupportedSeries,
    'targets' => l10n.onlineChooseTargetCard,
    'selection' when _action == 'dexter-assassination' =>
      l10n.onlineChooseDiamondPayment,
    'selection' when _action.contains('attack') =>
      l10n.onlineChooseAttackingClubs,
    'ace' when _action == 'justice' => l10n.onlineConsumedQueen,
    'ace' => l10n.onlineChooseAce,
    'protection' => l10n.onlineChooseProtection,
    'payment' => l10n.onlineChoosePayment,
    'formation' => l10n.onlineChooseFormationMembers,
    'kings' => l10n.onlineChooseKings,
    'receive' => l10n.onlineChooseRequestedCards,
    _ =>
      role == null
          ? l10n.onlineChooseBoard
          : l10n.onlineChooseRoleBoard(label(role).toLowerCase()),
  };

  List<Widget> _pendingSummary(AuthoritativeSnapshot s) {
    final b = s.board;
    final pending = (s.online['rules_context'] as Map?)?['pending'];
    final combat = (s.online['rules_context'] as Map?)?['combat'];
    return [
      Text(
        l10n.onlineDecisionOwner(
          '${b['required_actor']}',
          '${b['decision_kind'] ?? 'response'}',
          '${s.online['decision_stage'] ?? ''}',
        ),
      ),
      if (pending is Map)
        Text(
          pending['target_seat'] != null
              ? l10n.onlinePendingSeatTarget(
                  '${pending['actor']}',
                  label('${pending['type']}'),
                  '${pending['target_seat']}',
                )
              : l10n.onlinePendingSeatAction(
                  '${pending['actor']}',
                  label('${pending['type']}'),
                ),
        ),
      if (combat is Map) ...[
        Text(
          l10n.onlineCombatSeats('${combat['actor']}', '${combat['defender']}'),
        ),
        Text(
          l10n.onlineCommittedClubs(
            _currentCardNames(
              s,
              (combat['attack_cards'] as List? ?? []).cast<String>(),
            ),
          ),
        ),
        Text(
          l10n.onlineFrozenTargets(
            _currentCardNames(
              s,
              (combat['target_cards'] as List? ?? []).cast<String>(),
            ),
          ),
        ),
      ],
    ];
  }

  List<Widget> _requiredContent(AuthoritativeSnapshot s) {
    if (s.admissionWaiting)
      return [
        Text(l10n.economyAdmissionWaiting),
        FilledButton(onPressed: _openEconomy, child: Text(l10n.economyTitle)),
        ..._records(s),
      ];
    if (s.online['financial_phase'] == 'finalized')
      return [
        Text(
          (s.online['completed_games'] as int? ?? 0) >=
                  (s.online['games_per_match'] as int? ?? 1)
              ? l10n.onlineMatchComplete
              : l10n.onlineNextGamePending,
        ),
        ..._records(s),
      ];
    if (s.online['round_closed'] == true)
      return [
        Text(
          s.online['departure_pending'] == true
              ? l10n.onlineRoundChoiceHelp
              : l10n.onlineDepartureRecorded,
        ),
        if (s.online['departure_pending'] == true) ...[
          FilledButton(
            onPressed: _canMutate ? () => _chooseDeparture(false) : null,
            child: Text(l10n.onlineStayInGame),
          ),
          OutlinedButton(
            onPressed: _canMutate ? () => _chooseDeparture(true) : null,
            child: Text(l10n.onlineLeaveGame),
          ),
        ],
      ];
    if (s.phase != 'playing')
      return [
        ..._records(s),
        FilledButton(
          onPressed: _canMutate
              ? () => _run(() => c.submit('finish-settlement', {}))
              : null,
          child: Text(l10n.onlineFinishSettlement),
        ),
      ];
    final b = s.board;
    final pending = (s.online['rules_context'] as Map?)?['pending'];
    final combat = (s.online['rules_context'] as Map?)?['combat'];
    final effect =
        b['required_actor'] == s.seat &&
        (b['decision_id'] != null || _isKidnapperOutcome(s));
    return [
      ..._pendingSummary(s),
      if (combat is Map) ..._actionCalculations(s),
      if (_hasResponse(s))
        Text(
          pending is Map && (pending['response_types'] as List? ?? []).isEmpty
              ? l10n.onlineNoCardResponse
              : l10n.onlineCardResponseHelp,
        ),
      if (effect) ...[
        Text(
          _isKidnapperOutcome(s)
              ? l10n.onlineKidnapperExposedChoice
              : l10n.onlineResolveKind(label('${b['decision_kind']}')),
        ),
        for (final entry in (b['choices'] as List? ?? []).asMap().entries)
          if (entry.value is List)
            OutlinedButton(
              key: ValueKey('effect-choice-${entry.key}'),
              onPressed: _canMutate
                  ? () => setState(
                      () => _prepareIntent(
                        CardPlayIntent(
                          type: _isKidnapperOutcome(s)
                              ? 'kidnapper-outcome'
                              : 'decision',
                          seed: {'selection': entry.value},
                          explanation: l10n.onlineResolveAuthorizedStage,
                        ),
                      ),
                    )
                  : null,
              child: Text(
                (entry.value as List).isEmpty
                    ? l10n.onlineContinueWithoutCard
                    : _currentCardNames(
                        s,
                        (entry.value as List).cast<String>(),
                      ),
              ),
            ),
        if (b['decision_kind'] == 'justice')
          OutlinedButton(
            onPressed: _canMutate
                ? () => setState(
                    () => _prepareIntent(
                      CardPlayIntent(
                        type: 'decision-closed',
                        seed: {},
                        explanation: l10n.onlineJusticeRecordedOutcome,
                      ),
                    ),
                  )
                : null,
            child: Text(l10n.onlineJusticeConcealedChoice),
          ),
        _boardButton(),
      ],
    ];
  }

  List<Widget> _reviewContent(AuthoritativeSnapshot s) {
    final ready = _draftMatches(s) && _draftCardsVisible(s);
    final timing = commandTiming(_action, s.board, s.online);
    final sources = _sourceCards.where((id) => id != _fields['ace']?.text);
    return [
      if (!ready) Text(l10n.onlineTheGameOrDecisionChangedClearThisDraftBefore),
      if (_intentExplanation.isNotEmpty) Text(_intentExplanation),
      if (commandFields[_action]!.contains('pending_action_id') &&
          s.board['window_id'] != null)
        ..._pendingSummary(s),
      if (sources.isNotEmpty)
        Text(l10n.onlineSourceCards(_currentCardNames(s, sources))),
      for (final entry in _sets.entries)
        if (entry.value.isNotEmpty)
          Text(
            l10n.onlineCardRoleSummary(
              _roleLabel(entry.key),
              _currentCardNames(s, entry.value),
            ),
          ),
      if ((_fields['ace']?.text ?? '').isNotEmpty)
        Text(
          l10n.onlineCardRoleSummary(
            _action == 'justice'
                ? l10n.onlineConsumedQueen
                : _cardRoles[_fields['ace']!.text] == TableCardRole.modifier
                ? l10n.onlineModifier
                : l10n.cardRoleSource,
            _currentCardNames(s, [_fields['ace']!.text]),
          ),
        ),
      if (_action == 'coup') _coupCards(s),
      if ((_fields['formation_id']?.text ?? '').isNotEmpty) ...[
        Text(l10n.onlineFormationSummary(_formationKindFor(s))),
        Text(
          _currentCardNames(
            s,
            _formationMembers(s, _fields['formation_id']!.text),
          ),
        ),
      ],
      if (_effectChoice != null && _effectChoice!.isEmpty)
        Text(l10n.onlineContinueWithoutCard),
      if (ready) ..._contextParameters(s),
      ..._actionCalculations(s),
      if (timing != CommandTiming.available)
        Text(_timingExplanation(timing, s)),
      FilledButton(
        key: const ValueKey('online-submit'),
        onPressed:
            _canMutate &&
                ready &&
                _responseAllowed(_action, s) &&
                _formationReady(s) &&
                _combatChoiceReady() &&
                _parametersReady(s) &&
                timing == CommandTiming.available
            ? () => _submit(s)
            : null,
        child: Text(l10n.onlineConfirmAction(label(_action))),
      ),
      TextButton(
        onPressed: () => _dismissPopup(clear: true),
        child: Text(l10n.onlineClearSelection),
      ),
      ExpansionTile(
        title: Text(l10n.onlineCostsTitle),
        children: _costGuidance(s),
      ),
    ];
  }

  String _roleLabel(String role) => switch (role) {
    'selection' when _action == 'dexter-assassination' =>
      l10n.onlineDiamondPayment,
    'selection' when _action == 'barricade-sacrifice' =>
      l10n.onlineBarricadeOrder,
    'kings' => l10n.onlineDoppelgangerPair,
    'formation' => l10n.onlineComboMembers,
    _ => label(role),
  };

  String _formationKindFor(AuthoritativeSnapshot s) {
    final f = (s.board['formations'] as List? ?? [])
        .cast<Map>()
        .where((f) => f['id'] == _fields['formation_id']?.text)
        .firstOrNull;
    return label(f?['spec']['kind'] as String? ?? _formationKind);
  }

  List<String> _formationMembers(AuthoritativeSnapshot s, String id) {
    final formation = (s.board['formations'] as List? ?? [])
        .cast<Map>()
        .where((formation) => formation['id'] == id)
        .firstOrNull;
    return (formation?['spec']['cards'] as List? ?? []).cast<String>();
  }

  String _protectionTargetSummary(AuthoritativeSnapshot s) {
    if (_protectSuit.isNotEmpty) return label(_protectSuit);
    final formation = (s.board['formations'] as List? ?? [])
        .cast<Map>()
        .where((formation) => formation['id'] == _protectFormation)
        .firstOrNull;
    if (formation == null) return l10n.onlineChoosePublicTarget;
    return '${label(formation['spec']['kind'] as String)} · '
        '${_currentCardNames(s, _formationMembers(s, _protectFormation))}';
  }

  Widget _seatChoices(AuthoritativeSnapshot s, String name) => Wrap(
    spacing: 8,
    runSpacing: 8,
    children: [
      for (final seat in projectTable(s).seats.where((seat) => !seat.isSelf))
        _revealOnFocus(
          Semantics(
            selected: field(name).text == '${seat.number}',
            child: OutlinedButton(
              style: OutlinedButton.styleFrom(
                minimumSize: const Size(48, 48),
                backgroundColor: field(name).text == '${seat.number}'
                    ? CgmsColors.cyan.withValues(alpha: .15)
                    : null,
              ),
              onPressed: _canMutate
                  ? () => setState(() => field(name).text = '${seat.number}')
                  : null,
              child: Text(seat.name),
            ),
          ),
        ),
    ],
  );

  Widget _inlineChoice<T>(
    String title,
    T value,
    Map<T, String> options,
    ValueChanged<T> onChanged,
  ) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(title),
      Wrap(
        spacing: 8,
        runSpacing: 8,
        children: [
          for (final entry in options.entries)
            _revealOnFocus(
              Semantics(
                selected: entry.key == value,
                child: OutlinedButton(
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size(48, 48),
                    backgroundColor: entry.key == value
                        ? CgmsColors.cyan.withValues(alpha: .15)
                        : null,
                  ),
                  onPressed: _canMutate
                      ? () => setState(() => onChanged(entry.key))
                      : null,
                  child: Text(entry.value),
                ),
              ),
            ),
        ],
      ),
    ],
  );

  Widget _revealOnFocus(Widget child) => Builder(
    builder: (targetContext) => Focus(
      canRequestFocus: false,
      skipTraversal: true,
      onFocusChange: (focused) {
        if (focused)
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (targetContext.mounted) {
              Scrollable.ensureVisible(
                targetContext,
                alignmentPolicy:
                    ScrollPositionAlignmentPolicy.keepVisibleAtStart,
              );
              Scrollable.ensureVisible(
                targetContext,
                alignmentPolicy: ScrollPositionAlignmentPolicy.keepVisibleAtEnd,
              );
            }
          });
      },
      child: child,
    ),
  );

  List<Widget> _contextParameters(AuthoritativeSnapshot s) {
    if (_effectChoice != null) return [];
    if (_action == 'open-series')
      return [Text(l10n.onlineSeriesSummary(label(_suit)))];
    if (_action == 'open-formation')
      return [
        if (_formationKind.isEmpty)
          _inlineChoice<String>(
            l10n.onlineDoppelgangerFormation,
            _formationKind,
            {
              for (final kind in [
                'fate',
                'baron',
                'underground',
                'people',
                'justice',
                'code',
                'kidnapper',
              ])
                kind: label(kind),
            },
            (v) {
              _formationKind = v;
              _flags['protection'] = v == 'people';
            },
          ),
        Text(l10n.onlineComboSummary(label(_formationKind))),
        if (_flags['substitute'] == true) _boardButton('formation'),
        if (_flags['substitute'] == true && !_fixedSubstitution) ...[
          _boardButton('kings'),
          _inlineChoice<int>(l10n.onlineReplacementRank, _subRank, {
            11: l10n.onlineJackRank,
            12: l10n.onlineQueenRank,
            13: l10n.onlineKingRank,
          }, (v) => _subRank = v),
          _inlineChoice<String>(l10n.onlineReplacementSuit, _subSuit, {
            for (final suit in CardSuit.values) suit.name: label(suit.name),
          }, (v) => _subSuit = v),
        ],
        if (_fixedSubstitution)
          Text(
            '${l10n.onlineReplacementRank} · $_subRank · ${label(_subSuit)}',
          ),
        if (_flags['protection'] == true) ...[
          Text(l10n.onlineProtectionSummary(_protectionTargetSummary(s))),
          _boardButton('protection'),
        ],
      ];
    if (_action == 'purchase')
      return [_input('quantity', s), _boardButton('payment')];
    if (_action.contains('attack'))
      return [
        if ((_fields['ace']?.text ?? '').isNotEmpty) _input('value', s),
        if (_flags['infiltrator'] == true) Text(l10n.onlineInfiltratorCommit),
        if (_flags['exile'] == true) Text(l10n.onlineExileCommit),
        _boardButton('selection'),
        _boardButton('targets'),
        _boardButton('ace'),
      ];
    if (_action == 'numerical-defense') return [_input('value', s)];
    if (_action == 'barricade-sacrifice') return [_boardButton('selection')];
    if (_action == 'dexter-assassination')
      return [_boardButton('selection'), _boardButton('targets')];
    if (_action == 'attach')
      return [
        if (_suit.isNotEmpty) Text(l10n.onlineAttachTo(label(_suit))),
        if (_suit.isEmpty) ...[
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final suit in royalAttachmentSuits(
                s.board,
                _sets['selection'] ?? {},
              ))
                OutlinedButton(
                  onPressed: _canMutate
                      ? () {
                          setState(() => _suit = suit);
                          if (_dropContinuation && _readyToSubmit(s))
                            _submit(s);
                        }
                      : null,
                  child: Text(l10n.onlineAttachTo(label(suit))),
                ),
            ],
          ),
        ],
        _boardButton('attachment'),
      ];
    if (_action == 'fate')
      return [_seatChoices(s, 'target_seat'), _boardButton('selection')];
    if (_action == 'justice') return [_boardButton('ace')];
    if (_action == 'code') return [_input('value', s), _input('price', s)];
    if (_action == 'kidnapper' || _action == 'main-inflation')
      return [_seatChoices(s, 'target_seat')];
    if (_action == 'ponzi') return [_seatChoices(s, 'value')];
    if (_action == 'offer')
      return [
        Text(
          _flags['loan'] == true
              ? l10n.onlineLoanIntactFormation
              : l10n.onlinePrivateTrade,
        ),
        _seatChoices(s, 'target_seat'),
        if (_flags['loan'] != true) _boardButton('receive'),
        if (_freshOfferId != null &&
            (s.board['proposals'] as List? ?? []).any(
              (p) =>
                  p['from'] == s.seat &&
                  p['status'] == 'offered' &&
                  p['turn'] == s.board['turn'],
            ))
          _inlineChoice<String>(
            l10n.onlineProposalRevision,
            field('offer_id').text,
            {
              _freshOfferId!: l10n.onlineNewOffer,
              for (final p in s.board['proposals'] as List)
                if (p['from'] == s.seat &&
                    p['status'] == 'offered' &&
                    p['turn'] == s.board['turn'])
                  p['id'] as String: l10n.onlineOfferReplacement(
                    '${p['terms']['to']}',
                    '${(p['revision'] as int) + 1}',
                  ),
            },
            (id) {
              field('offer_id').text = id;
              final previous = (s.board['proposals'] as List)
                  .cast<Map>()
                  .where((p) => p['id'] == id)
                  .firstOrNull;
              field('revision').text =
                  '${previous == null ? 1 : (previous['revision'] as int) + 1}';
            },
          ),
        Text(l10n.onlinePrivateOfferTerms),
      ];
    if (_action == 'declare-ordinary')
      return [
        _inlineChoice<String>(
          l10n.onlineVictoryThreshold,
          field('threshold').text,
          {
            '': l10n.onlineNone,
            'diamonds': l10n.onlineNumberDiamondThreshold,
            'royals': l10n.onlineRoyalThreshold,
          },
          (v) => field('threshold').text = v,
        ),
      ];
    const financial = {
      'promise-offer',
      'promise-final-offer',
      'voluntary-transfer',
      'forgive',
    };
    if (financial.contains(_action))
      return [
        for (final name in commandFields[_action]!)
          if (!{'promise_id', 'debt_id'}.contains(name)) _input(name, s),
      ];
    return [];
  }

  bool _parametersReady(AuthoritativeSnapshot s) {
    if (_effectChoice != null) return true;
    if (_action == 'coup') return canDeclareCoup(s);
    if (_action == 'expose-unused-ace') {
      final ace = _fields['ace']?.text ?? '';
      return suggestCardPlays(
        board: s.board,
        online: s.online,
        cardIds: {ace},
      ).any((intent) => intent.type == 'expose-unused-ace');
    }
    if (_action == 'open-series') return (_sets['selection'] ?? {}).isNotEmpty;
    if (_action == 'attach') {
      return royalAttachmentSuits(
        s.board,
        _sets['selection'] ?? {},
      ).contains(_suit);
    }
    if (_action == 'purchase') {
      final quantity = int.tryParse(_fields['quantity']?.text ?? '') ?? 0;
      if (quantity < 1 || (_sets['payment'] ?? {}).isEmpty) return false;
      final price = (s.online['rules_context'] as Map?)?['purchase_unit_price'];
      // Use only the authority's public price and own effective payment. Supply
      // and final legality remain private server checks; never infer a deck count.
      return price is! int ||
          _purchasePayment(s).compareTo(
                ExactPoints.integer(quantity) * ExactPoints.integer(price),
              ) >=
              0;
    }
    if (_action.contains('attack'))
      return (_sets['selection'] ?? {}).isNotEmpty &&
          (_sets['targets'] ?? {}).isNotEmpty;
    if (_action == 'barricade-sacrifice')
      return (_sets['selection'] ?? {}).length == 6 &&
          _barricadeQueen != null &&
          _sets['selection']!.first == _barricadeQueen;
    if (_action == 'dexter-assassination')
      return (_sets['selection'] ?? {}).length == 1 &&
          (_sets['targets'] ?? {}).length == 1;
    if (_action == 'fate')
      return (_sets['selection'] ?? {}).length == 1 &&
          (_fields['target_seat']?.text ?? '').isNotEmpty;
    if (_action == 'justice') return (_fields['ace']?.text ?? '').isNotEmpty;
    if (_action == 'decision' || _action == 'kidnapper-outcome') return false;
    if (_action == 'offer' ||
        _action == 'kidnapper' ||
        _action == 'main-inflation')
      return (_fields['target_seat']?.text ?? '').isNotEmpty;
    if (_action == 'ponzi') return (_fields['value']?.text ?? '').isNotEmpty;
    return true;
  }

  String _timingExplanation(CommandTiming timing, AuthoritativeSnapshot s) =>
      switch (timing) {
        CommandTiming.ownTurn => l10n.onlineTimingOwnTurn,
        CommandTiming.pending => l10n.onlineTimingPending,
        CommandTiming.otherResponder => l10n.onlineTimingResponder(
          '${s.board['required_actor']}',
        ),
        CommandTiming.effectInput => l10n.onlineTimingEffect,
        CommandTiming.noResponse => l10n.onlineTimingNoResponse,
        CommandTiming.noDecision => l10n.onlineTimingNoDecision,
        CommandTiming.automatic =>
          s.online['server_pending'] == true
              ? l10n.onlineServerProgress
              : l10n.onlineExactSettlementIsProgressingOrdinaryActionsWait,
        CommandTiming.available => '',
      };

  // Explanations quote the accepted lifecycle, not a second legality engine.
  // Availability, support and exact costs are still validated by the authority.
  List<Widget> _costGuidance(AuthoritativeSnapshot s) {
    final lines = switch (_action) {
      'attack' ||
      'infiltrator-attack' ||
      'baron-attack' ||
      'bomb-attack' ||
      'baron-bomb-attack' => [
        l10n.onlineCostsAttack,
        l10n.onlineCostsCombatAce,
        if (_action.startsWith('baron-')) l10n.onlineCostsBaron,
        if (_action.contains('bomb')) l10n.onlineCostsBomb,
        if (_action == 'infiltrator-attack' || _flags['infiltrator'] == true)
          l10n.onlineCostsInfiltrator,
        if (_flags['exile'] == true) l10n.onlineCostsExile,
      ],
      'open-series' => [l10n.onlineCostsOpening],
      'open-formation' => [l10n.onlineCostsOpening, l10n.onlineCostsFormation],
      'take-back' => [l10n.onlineCostsTakeBack],
      'attach' => [l10n.onlineCostsAttach],
      'fate' => [l10n.onlineCostsFate],
      'coup' => [l10n.onlineCostsCoup],
      'justice' => [l10n.onlineCostsJustice],
      'code' => [l10n.onlineCostsCode],
      'kidnapper' || 'kidnapper-outcome' => [l10n.onlineCostsKidnapper],
      'ponzi' => [l10n.onlineCostsPonzi],
      'ringleader' => [l10n.onlineCostsRingleader],
      'richer-sacrifice' => [l10n.onlineCostsRicher],
      'barricade-sacrifice' => [l10n.onlineCostsBarricade],
      'negotiate' => [l10n.onlineCostsNegotiator],
      'dexter-assassination' => [l10n.onlineCostsDexter],
      'confinement' => [l10n.onlineCostsConfinement],
      'numerical-defense' ||
      'advancement-defense' => [l10n.onlineCostsCombatResponse],
      'inflation' || 'main-inflation' => [l10n.onlineCostsInflation],
      'compensation' ||
      'compensation-response' => [l10n.onlineCostsCompensation],
      'purchase' => [l10n.onlineCostsPurchase],
      'offer' || 'accept-offer' => [l10n.onlineCostsOffer],
      'return-loan' => [l10n.onlineCostsReturnLoan],
      'decision' || 'decision-closed' => [
        l10n.onlineCostsDecision,
        if ((s.board['decision_kind'] as String? ?? '').contains('dexter'))
          l10n.onlineCostsDexterPrevention,
        if ((s.board['decision_kind'] as String? ?? '').contains('negotiator'))
          l10n.onlineCostsNegotiator,
        if (s.board['decision_kind'] == 'justice') l10n.onlineCostsJustice,
      ],
      'pass' => [l10n.onlineCostsPass],
      _ => <String>[],
    };
    if (lines.isEmpty) return [];
    return [
      Card(
        key: const ValueKey('online-cost-guidance'),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l10n.onlineCostsTitle,
                style: Theme.of(context).textTheme.titleSmall,
              ),
              for (final line in [...lines, l10n.onlineCostsGeneral])
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Text(line),
                ),
            ],
          ),
        ),
      ),
    ];
  }

  Widget _input(String name, AuthoritativeSnapshot s) {
    if (name == 'value' &&
        (_action.contains('attack') || _action == 'numerical-defense')) {
      final value = field(name);
      final aceId = field('ace').text;
      final clubAce = (s.board['cards'] as List? ?? []).any(
        (p) =>
            p['card']['id'] == aceId &&
            p['card']['rank'] == 1 &&
            p['card']['suit'] == 'clubs',
      );
      final options = <String>{
        '',
        if (_action != 'numerical-defense' && clubAce) '0',
        for (var n = 2; n <= 10; n++) '$n',
      };
      return _inlineChoice<String>(l10n.onlineAceCombatValue, value.text, {
        for (final option in options.where((v) => v.isNotEmpty))
          option: option == '0' ? l10n.onlineAdvancementValue : option,
      }, (v) => value.text = v);
    }
    if (name == 'condition')
      return Text(l10n.onlineConditionTheNamedAwardOccursForYouInThis);
    if (name == 'amount')
      return Row(
        children: [
          Expanded(
            child: _number('numerator', l10n.onlineAmountNumerator, '1'),
          ),
          Text(CgmsLocalizations.of(context).online),
          Expanded(child: _number('denominator', l10n.onlineDenominator, '1')),
        ],
      );
    if ([
      'value',
      'quantity',
      'price',
      'revision',
      'target_seat',
      'recipient',
    ].contains(name))
      return _number(
        name,
        _action == 'ponzi' && name == 'value'
            ? l10n.onlineTargetSeat
            : _action == 'code' && name == 'value'
            ? l10n.onlineDiamondBonusRange
            : _action == 'code' && name == 'price'
            ? l10n.onlinePurchasePriceRange
            : label(name),
        name == 'price' && _action == 'code' ? '5' : '1',
      );
    return TextField(
      controller: field(name),
      decoration: InputDecoration(labelText: label(name)),
    );
  }

  Widget _number(String key, String name, String initial) => TextField(
    controller: field(key, initial),
    decoration: InputDecoration(labelText: name),
    keyboardType: TextInputType.number,
  );

  Future<void> _submit(AuthoritativeSnapshot rendered) async {
    if (_submitting) return;
    await _run(() async {
      final s = c.snapshot;
      if (s == null ||
          !_canMutate ||
          !_intentPrepared ||
          s.gameId != rendered.gameId ||
          s.phase != rendered.phase ||
          (!const {'coup', 'declare-ordinary'}.contains(_action) &&
              (s.version != rendered.version ||
                  s.board['window_id'] != rendered.board['window_id'] ||
                  s.board['decision_id'] != rendered.board['decision_id'])) ||
          !_draftMatches(s) ||
          !_formationReady(s) ||
          !_combatChoiceReady() ||
          !_parametersReady(s) ||
          !_responseAllowed(_action, s) ||
          commandTiming(_action, s.board, s.online) !=
              CommandTiming.available) {
        throw StateError('STATE_CONFLICT');
      }
      _captureDraft();
      if (!_draftCardsVisible(s)) throw StateError('STATE_CONFLICT');
      if (_effectChoice != null) {
        final choices = _currentEffectChoices(s);
        if (!choices.any(
          (v) =>
              v.length == _effectChoice!.length &&
              v.every(_effectChoice!.contains),
        ))
          throw StateError('STATE_CONFLICT');
        _submitting = true;
        try {
          await c.submit(_action, {
            'pending_action_id': _draftWindow,
            if (_action == 'decision') 'decision_id': _draftDecision,
            'selection': _effectChoice,
          });
        } finally {
          _submitting = false;
        }
        if (c.unknownOperation == null && c.errorCode == null)
          setState(() {
            _clearDraft();
            _popup = null;
          });
        return;
      }
      final payload = <String, dynamic>{};
      for (final name in commandFields[_action]!) {
        switch (name) {
          case 'pending_action_id':
            payload[name] = _draftWindow;
          case 'decision_id':
            payload[name] = _draftDecision;
          case 'selection':
          case 'targets':
          case 'payment':
            payload[name] = (_sets[name] ?? {}).toList();
          case 'suit':
            payload[name] = _suit;
          case 'accept':
          case 'exile':
          case 'infiltrator':
            if (_flags.containsKey(name)) payload[name] = _flags[name];
          case 'value':
          case 'quantity':
          case 'price':
          case 'revision':
          case 'target_seat':
          case 'recipient':
            if (field(name).text.isNotEmpty)
              payload[name] = int.parse(field(name).text);
          case 'amount':
            payload[name] = {
              'numerator': field('numerator', '1').text,
              'denominator': field('denominator', '1').text,
            };
          case 'condition':
            payload[name] = {
              'kind': 'award-occurrence',
              'game_id': s.gameId,
              'award_id': field('award_id').text,
              'payer': s.seat,
            };
          case 'terms':
            payload[name] = {
              'to': int.parse(field('target_seat', '2').text),
              'give': (_sets['give'] ?? {}).toList(),
              'receive': (_sets['receive'] ?? {}).toList(),
              'loan': _flags['loan'] ?? false,
            };
          case 'formation':
            payload[name] = {
              'kind': _formationKind,
              'cards': (_sets['formation'] ?? {}).toList(),
              if (_flags['substitute'] == true)
                'substitute': {
                  'kings': (_sets['kings'] ?? {}).toList(),
                  'slot': {'rank': _subRank, 'suit': _subSuit},
                },
              if (_flags['protection'] == true)
                'protection': {
                  if (_protectSuit.isNotEmpty) 'series': _protectSuit,
                  if (_protectFormation.isNotEmpty)
                    'formation': _protectFormation,
                },
            };
          default:
            if (field(name).text.isNotEmpty) payload[name] = field(name).text;
        }
      }
      _submitting = true;
      try {
        await c.submit(_action, payload);
      } finally {
        _submitting = false;
      }
      if (c.unknownOperation == null && c.errorCode == null)
        setState(() {
          _clearDraft();
          _popup = null;
        });
    });
  }

  ExactPoints _purchasePayment(AuthoritativeSnapshot s) {
    final rules = s.online['rules_context'] as Map?;
    return (s.board['cards'] as List? ?? [])
        .cast<Map>()
        .where(
          (p) => (_sets['payment'] ?? {}).contains((p['card'] as Map)['id']),
        )
        .fold<ExactPoints>(
          ExactPoints.integer(0),
          (sum, p) =>
              sum +
              ExactPoints.fromStrings(
                '${(p['card'] as Map)['rank']}',
                rules?['inflation'] == true ? '2' : '1',
              ),
        );
  }

  List<Widget> _actionCalculations(AuthoritativeSnapshot s) {
    final rules = s.online['rules_context'] as Map?;
    if (rules == null) return [];
    final combat = rules['combat'] as Map?;
    final payment = _purchasePayment(s);
    final quantity = int.tryParse(field('quantity', '1').text) ?? 0;
    final cost =
        ExactPoints.integer(quantity) *
        ExactPoints.integer(rules['purchase_unit_price'] as int);
    return [
      if (_action == 'purchase')
        Text(
          l10n.onlinePurchaseCalculation(
            (rules['purchase_unit_price']).toString(),
            (quantity).toString(),
            (payment.format()).toString(),
            (cost.format()).toString(),
            ((payment - cost).format()).toString(),
          ),
        ),
      if (combat != null)
        Text(
          l10n.onlineCombatCalculation(
            (exactAmount(combat['physical_attack']).format()).toString(),
            (combat['attack_modifier']).toString(),
            (exactAmount(combat['attack']).format()).toString(),
            (exactAmount(combat['physical_defense']).format()).toString(),
            (combat['defense_modifier']).toString(),
            (exactAmount(combat['defense']).format()).toString(),
            (label(combat['comparison'] as String)).toString(),
          ),
        ),
    ];
  }

  List<Widget> _quotaContext(AuthoritativeSnapshot s) {
    final rules = s.online['rules_context'] as Map?;
    if (rules == null) return [];
    return [
      Text(
        l10n.onlineOpeningAceStatus(
          rules['opening_used'] == true
              ? l10n.onlineUsedThisTurn
              : l10n.onlineAvailable,
          rules['ace_used_this_round'] == true
              ? l10n.onlineUsedThisRound
              : l10n.onlineAvailable,
        ),
      ),
      Text(
        l10n.onlinePrivilegesStatus(
          rules['compensation_used'] == true
              ? l10n.onlineUsed
              : l10n.onlineAvailable,
          rules['underground'] == true ? l10n.onlineHeld : l10n.onlineAbsent,
        ),
      ),
      for (final entry in (rules['quota_used'] as Map? ?? {}).entries)
        Text(
          CgmsLocalizations.of(context).onlineQuotaStatus(
            label(entry.key as String),
            entry.value == true
                ? CgmsLocalizations.of(context).onlineQuotaUsed
                : CgmsLocalizations.of(context).onlineQuotaAvailable,
          ),
        ),
    ];
  }

  String _termCards(AuthoritativeSnapshot s, dynamic ids) {
    if (ids == null || (ids as List).isEmpty)
      return l10n.onlineNone.toLowerCase();
    final cards = (s.board['cards'] as List? ?? []).cast<Map>();
    return ids
        .map((id) {
          final card = cards
              .where((p) => (p['card'] as Map)['id'] == id)
              .firstOrNull;
          if (card != null) return projectCard(card).label;
          final face = (s.projection['known_faces'] as List? ?? [])
              .cast<Map>()
              .where((v) => v['handle'] == id)
              .firstOrNull;
          if (face == null) return l10n.onlineUnrevealedCard;
          final rank = switch (face['rank']) {
            1 => 'A',
            11 => 'J',
            12 => 'Q',
            13 => 'K',
            _ => '${face['rank']}',
          };
          return l10n.onlineKnownOfferCard(
            (rank).toString(),
            (label(face['suit'] as String)).toString(),
            (ids.indexOf(id) + 1).toString(),
          );
        })
        .join(', ');
  }

  Widget _recordCommand(
    String title,
    String action,
    Map<String, dynamic> payload,
  ) => TextButton(
    onPressed:
        _canMutate &&
            c.snapshot != null &&
            commandTiming(action, c.snapshot!.board, c.snapshot!.online) ==
                CommandTiming.available
        ? () => _run(() => c.submit(action, payload))
        : null,
    child: Text(title),
  );

  Widget _editRecord(String title, String action, String key, String id) =>
      TextButton(
        onPressed: _canMutate
            ? () => setState(() {
                _clearDraft();
                _selectAction(action);
                field(key).text = id;
                _captureDraft();
                _openPopup('review');
              })
            : null,
        child: Text(title),
      );

  List<Widget> _records(AuthoritativeSnapshot s) => [
    if (s.online['ending'] != null)
      Text(
        s.online['declarer'] != null
            ? l10n.onlineEndingDeclarer(
                '${s.online['ending']}',
                '${s.online['declarer']}',
              )
            : l10n.onlineEnding('${s.online['ending']}'),
      ),

    Text(
      l10n.onlineOwnMatchScore(
        exactAmount(s.online['own_match_score']).format(),
      ),
    ),
    for (final own in projectTable(s).seats.where((seat) => seat.isSelf))
      Text(
        l10n.adaptiveFinances(
          own.finances.gameScore.format(),
          own.finances.available.format(),
          own.finances.debt.format(),
        ),
      ),
    for (final p in s.projection['promises'] as List? ?? [])
      Card(
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            children: [
              Text(
                l10n.onlinePromiseParties(
                  ((p['payer'] as int) + 1).toString(),
                  ((p['recipient'] as int) + 1).toString(),
                  (p['status']).toString(),
                ),
              ),
              Text(
                l10n.onlinePromiseTerms(
                  (p['mode']).toString(),
                  (exactAmount(p['amount']).format()).toString(),
                  (p['award_id']).toString(),
                ),
              ),
              Text(
                l10n.onlinePromiseSettlement(
                  (exactAmount(p['promised']).format()).toString(),
                  (exactAmount(p['paid']).format()).toString(),
                  (exactAmount(p['final']).format()).toString(),
                ),
              ),
              if (p['recipient'] == s.seat - 1 &&
                  p['status'] == 'offered' &&
                  s.phase == 'playing')
                _recordCommand(l10n.onlineAcceptPromise, 'promise-accept', {
                  'promise_id': p['id'],
                }),
              if (p['recipient'] == s.seat - 1 &&
                  p['status'] == 'final-offered')
                Wrap(
                  children: [
                    _recordCommand(
                      l10n.onlineAcceptFinalOffer,
                      'promise-final-answer',
                      {'promise_id': p['id'], 'accept': true},
                    ),
                    _recordCommand(
                      l10n.onlineRejectFinalOffer,
                      'promise-final-answer',
                      {'promise_id': p['id'], 'accept': false},
                    ),
                  ],
                ),
              if (p['payer'] == s.seat - 1 &&
                  p['triggered'] == true &&
                  ['accepted', 'final-accepted'].contains(p['status']))
                Wrap(
                  children: [
                    _recordCommand(l10n.onlinePayPromise, 'promise-pay', {
                      'promise_id': p['id'],
                    }),
                    _recordCommand(l10n.onlineRefusePromise, 'promise-refuse', {
                      'promise_id': p['id'],
                    }),
                    if (p['status'] == 'accepted')
                      _editRecord(
                        l10n.onlineProposeFinalAmount,
                        'promise-final-offer',
                        'promise_id',
                        p['id'] as String,
                      ),
                  ],
                ),
            ],
          ),
        ),
      ),
    for (final d in s.projection['debts'] as List? ?? [])
      Card(
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            children: [
              Text(
                l10n.onlineDebtParties(
                  (d['debtor'] == s.seat - 1
                          ? l10n.onlineYouOwe
                          : l10n.onlineOwedToYou)
                      .toString(),
                  (exactAmount(d['remaining']).format()).toString(),
                  ((d['debtor'] as int) + 1).toString(),
                  ((d['creditor'] as int) + 1).toString(),
                ),
              ),
              Text(l10n.onlineOriginGame((d['game_id']).toString())),
              if (d['creditor'] == s.seat - 1 || d['creditor'] == -1)
                _editRecord(
                  l10n.onlineForgiveDebt,
                  'forgive',
                  'debt_id',
                  d['id'] as String,
                ),
            ],
          ),
        ),
      ),
    for (final p in s.board['proposals'] as List? ?? [])
      Card(
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            children: [
              Text(
                l10n.onlinePrivateOfferStatus(
                  (p['revision']).toString(),
                  (p['status']).toString(),
                ),
              ),
              Text(
                l10n.onlineOfferParties(
                  (p['from']).toString(),
                  (p['terms']['to']).toString(),
                  ((p['terms']['loan'] as bool? ?? false) ? 'loan' : 'trade')
                      .toString(),
                ),
              ),
              Text(
                '${p['from'] == s.seat ? l10n.onlineYouGive : l10n.onlineYouReceive}: ${_termCards(s, p['terms']['give'])}',
              ),
              Text(
                '${p['from'] == s.seat ? l10n.onlineYouReceive : l10n.onlineYouGive}: ${_termCards(s, p['terms']['receive'])}',
              ),
              if (p['status'] == 'offered')
                Wrap(
                  children: [
                    for (final action
                        in p['from'] == s.seat
                            ? ['withdraw-offer']
                            : ['accept-offer', 'decline-offer'])
                      _recordCommand(label(action), action, {
                        'offer_id': p['id'],
                        'revision': p['revision'],
                      }),
                  ],
                ),
            ],
          ),
        ),
      ),
    for (final effect in s.board['effects'] as List? ?? [])
      Text(
        l10n.onlineLastingEffect(
          (label(effect['kind'] as String)).toString(),
          (effect['source']).toString(),
          (effect['target']).toString(),
          (effect['custodian']).toString(),
          (effect['expiry']).toString(),
        ),
      ),
    for (final draw in s.online['draw_queue'] as List? ?? [])
      Text(
        l10n.onlineCompensationDraw(
          (draw['seat']).toString(),
          (draw['remaining']).toString(),
        ),
      ),
    for (final correction in s.online['corrections'] as List? ?? [])
      Text(
        l10n.onlineMatchCorrection(
          (exactAmount(correction['amount']).format()).toString(),
          (correction['game_id']).toString(),
        ),
      ),
    for (final result in s.online['results'] as List? ?? [])
      Text(
        l10n.onlineCompletedGame(
          (result['scores'] as List)
              .asMap()
              .entries
              .map(
                (e) => l10n.onlineSeatScore(
                  '${e.key + 1}',
                  exactAmount(e.value).format(),
                ),
              )
              .join(' · '),
        ),
      ),
    if (s.online['ranks'] is List && (s.online['ranks'] as List).isNotEmpty)
      Text(l10n.onlineRanksBySeat((s.online['ranks']).toString())),
  ];
}
