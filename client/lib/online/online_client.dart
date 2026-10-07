import 'dart:async';
import 'dart:collection';
import 'dart:convert';
import 'dart:math';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/foundation.dart';
import 'command_contract.dart';

const _botDifficulties = {'beginner', 'standard', 'advanced'};

class OnlineResponse {
  const OnlineResponse(this.status, this.data);
  final int status;
  final dynamic data;
}

abstract interface class OnlineTransport {
  Future<OnlineResponse> request(
    String method,
    String path, {
    Map<String, dynamic>? body,
    Map<String, String> headers = const {},
  });
  Stream<Map<String, dynamic>> events(String path);
  void close();
}

/// Stores only unresolved operation envelopes, scoped to an authenticated actor.
/// Implementations must never persist sessions, headers, snapshots or credentials.
abstract interface class OnlineOperationStore {
  String? read(String account, String scope);
  void write(String account, String scope, String value);
  void remove(String account, String scope);
}

// Freeze nested payloads too: later UI edits must never alter a queued intent.
dynamic _freeze(dynamic value) {
  if (value is Map)
    return Map<String, dynamic>.unmodifiable(
      value.map((key, value) => MapEntry(key as String, _freeze(value))),
    );
  if (value is List) return List<dynamic>.unmodifiable(value.map(_freeze));
  return value;
}

bool _accountDate(dynamic value) {
  if (value is! String) return false;
  // Account timestamps are RFC3339 instants. DateTime.tryParse alone accepts
  // local dates and silently normalizes impossible calendar dates.
  final parts = RegExp(
    r'^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$',
  ).firstMatch(value);
  if (parts == null || DateTime.tryParse(value)?.isUtc != true) return false;
  final values = [for (var i = 1; i <= 6; i++) int.parse(parts.group(i)!)];
  final normalized = DateTime.utc(
    values[0],
    values[1],
    values[2],
    values[3],
    values[4],
    values[5],
  );
  return normalized.year == values[0] &&
      normalized.month == values[1] &&
      normalized.day == values[2] &&
      normalized.hour == values[3] &&
      normalized.minute == values[4] &&
      normalized.second == values[5];
}

bool _validEconomyAccount(Map<String, dynamic> data) {
  if (data['enabled'] == false) return true;
  final benefits = data['benefits'];
  bool inRange(dynamic value, int minimum, int maximum) =>
      value is int && value >= minimum && value <= maximum;
  // Bounds mirror the authority's economy Policy/MaxBalance and account schema.
  return data['enabled'] == true &&
      benefits is Map &&
      inRange(benefits['dirt'], 0, 1000000000000000) &&
      benefits['unlimited'] is bool &&
      benefits['no_ads'] is bool &&
      _accountDate(benefits['pass_until']) &&
      _accountDate(benefits['premium_until']) &&
      _accountDate(benefits['legacy_ad_free_until']) &&
      (!benefits.containsKey('ad_free_until') ||
          _accountDate(benefits['ad_free_until'])) &&
      benefits['cosmetics'] is List &&
      (benefits['cosmetics'] as List).every((value) => value is String) &&
      inRange(data['free_starts_remaining'], 0, 3) &&
      _accountDate(data['resets_at']) &&
      inRange(data['reward'], 1, 1000000000) &&
      inRange(data['pass_day_price'], 1, 1000000000);
}

class AuthoritativeSnapshot {
  AuthoritativeSnapshot.fromJson(Map<String, dynamic> data)
    : admissionWaiting = data['admission_waiting'] as bool? ?? false,
      version = data['version'] as int,
      cursor = (data['cursor'] ?? data['version']) as int,
      projection = _freeze(data['projection']) as Map<String, dynamic> {
    if (version < 0 || cursor != version || gameId.isEmpty)
      throw const FormatException('INVALID_SNAPSHOT');
  }
  final int version, cursor;
  final bool admissionWaiting;
  final Map<String, dynamic> projection;
  Map<String, dynamic> get board => projection['board'] as Map<String, dynamic>;
  Map<String, dynamic> get online =>
      projection['online'] as Map<String, dynamic>;
  String get gameId => board['game_id'] as String;
  int get seat => board['seat'] as int;
  String get phase => board['phase'] as String;
}

/// A single operation owner. Transport timeouts never make gameplay decisions.
class OnlineClient extends ChangeNotifier {
  OnlineClient(this.transport, {OnlineOperationStore? operationStore})
    : _operationStore =
          operationStore ??
          (transport is OnlineOperationStore
              ? transport as OnlineOperationStore
              : null);
  final OnlineOperationStore? _operationStore;
  bool _storageFault = false;
  String? get _actor => account?['id'] as String?;
  void _store(String scope, Map<String, dynamic>? value) {
    if (_operationStore == null) return;
    if (_actor == null) throw StateError('UNAUTHENTICATED');
    try {
      if (value == null) {
        _operationStore.remove(_actor!, scope);
      } else {
        _operationStore.write(_actor!, scope, jsonEncode(value));
      }
    } catch (_) {
      _storageFault = true;
      errorCode = 'RECOVERY_UNAVAILABLE';
      rethrow;
    }
  }

  Map<String, dynamic>? _load(String scope) {
    if (_operationStore == null || _actor == null) return null;
    try {
      final raw = _operationStore.read(_actor!, scope);
      return raw == null
          ? null
          : _freeze(jsonDecode(raw)) as Map<String, dynamic>;
    } catch (_) {
      _storageFault = true;
      errorCode = 'RECOVERY_UNAVAILABLE';
      return null;
    }
  }

  void _clearUnknown() {
    _store('match:$matchId', null);
    _unknown = null;
  }

  void _clearRoomOperation() {
    _store('room', null);
    _roomOperation = null;
    _roomOperationPath = null;
  }

  void _restoreRoomOperation() {
    final saved = _load('room');
    if (saved == null) return;
    try {
      final path = saved['path'] as String;
      final body = saved['body'] as Map<String, dynamic>;
      if (!(path == '/v1/rooms' ||
              RegExp(r'^/v1/rooms/[^/]+/matches$').hasMatch(path)) ||
          body['command_id'] is! String ||
          body.keys.any(
            (key) =>
                !(path == '/v1/rooms'
                        ? [
                            'command_id',
                            'capacity',
                            'games_per_match',
                            'bot_difficulty',
                          ]
                        : ['command_id'])
                    .contains(key),
          ) ||
          (path == '/v1/rooms' &&
              (body['capacity'] is! int ||
                  body['games_per_match'] is! int ||
                  (body.containsKey('bot_difficulty') &&
                      !_botDifficulties.contains(body['bot_difficulty'])))))
        throw const FormatException();
      _roomOperationPath = path;
      _roomOperation = body;
    } catch (_) {
      _storageFault = true;
      errorCode = 'RECOVERY_UNAVAILABLE';
    }
  }

  void _restoreMatchOperation() {
    final saved = _load('match:$matchId');
    if (saved == null) return;
    if (saved.keys.any(
          (key) => ![
            'command_id',
            'game_id',
            'expected_version',
            'type',
            'payload',
          ].contains(key),
        ) ||
        saved['command_id'] is! String ||
        saved['game_id'] is! String ||
        saved['expected_version'] is! int ||
        saved['payload'] is! Map ||
        !commandFields.containsKey(saved['type'])) {
      _storageFault = true;
      errorCode = 'RECOVERY_UNAVAILABLE';
      return;
    }
    try {
      validateCommandPayload(
        saved['type'] as String,
        Map<String, dynamic>.from(saved['payload'] as Map),
      );
    } catch (_) {
      _storageFault = true;
      errorCode = 'RECOVERY_UNAVAILABLE';
      return;
    }
    _unknown = saved;
  }

  Map<String, dynamic>? economy, passReceipt;
  Map<String, dynamic>? _pendingPass;
  Map<String, dynamic>? get pendingPass => _pendingPass;
  String? economyError;
  bool economyLoading = false;
  int _economyGeneration = 0;

  void _restorePass() {
    final saved = _load('economy-pass');
    if (saved == null) return;
    if (saved.length != 2 ||
        saved['operation_id'] is! String ||
        !RegExp(
          r'^[a-zA-Z0-9_-]{1,128}$',
        ).hasMatch(saved['operation_id'] as String) ||
        saved['days'] is! int ||
        (saved['days'] as int) < 1 ||
        (saved['days'] as int) > 7) {
      _storageFault = true;
      economyError = 'RECOVERY_UNAVAILABLE';
      return;
    }
    _pendingPass = saved;
  }

  /// Balances are private account state, never a reconstruction from game scores.
  Future<void> refreshEconomy() async {
    if (_actor == null || _disposed) return;
    final actor = _actor;
    final generation = ++_economyGeneration;
    economyLoading = true;
    _notify();
    try {
      final result = await _request('GET', '/v1/economy');
      if (_disposed || actor != _actor || generation != _economyGeneration)
        return;
      if (result.status != 200 ||
          result.data is! Map ||
          result.data['enabled'] is! bool)
        throw const FormatException('ECONOMY_UNAVAILABLE');
      final data = Map<String, dynamic>.from(result.data as Map);
      if (!_validEconomyAccount(data))
        throw const FormatException('ECONOMY_UNAVAILABLE');
      economy = _freeze(data);
      if (_pendingPass == null && !_storageFault) economyError = null;
    } catch (_) {
      if (!_disposed && actor == _actor && generation == _economyGeneration)
        economyError = 'ECONOMY_UNAVAILABLE';
    } finally {
      if (!_disposed && actor == _actor && generation == _economyGeneration) {
        economyLoading = false;
        _notify();
      }
    }
  }

  bool _passResult(OnlineResponse response, {required bool retrying}) {
    final body = _pendingPass!;
    final data = response.data;
    if (response.status == 200 &&
        data is Map &&
        data['operation_id'] == body['operation_id'] &&
        data['days'] == body['days'] &&
        data['cost'] is int &&
        (data['cost'] as int) > 0 &&
        data['expires_at'] is String &&
        DateTime.tryParse(data['expires_at']) != null) {
      _store('economy-pass', null);
      _pendingPass = null;
      passReceipt = _freeze(data);
      economyError = null;
      return true;
    }
    final code = data is Map ? data['code'] : null;
    // Only definitive business rejection after locked receipt lookup proves no
    // charge. Authentication/disablement/absence cannot resolve a lost reply.
    final definitive =
        response.status == 409 &&
            (code == 'INSUFFICIENT_DIRT' || code == 'PRODUCT_UNAVAILABLE') ||
        !retrying && response.status == 409 && code == 'ECONOMY_DISABLED' ||
        !retrying && response.status == 400 && code == 'INVALID_REQUEST';
    if (definitive) {
      _store('economy-pass', null);
      _pendingPass = null;
      economyError = code as String;
    } else {
      economyError = 'OUTCOME_UNKNOWN';
    }
    return false;
  }

  Future<void> buyPass(int days) async {
    if (days != 1 && days != 7) throw ArgumentError.value(days);
    if (_disposed ||
        _storageFault ||
        busy ||
        _pendingPass != null ||
        _unknown != null ||
        _roomOperation != null ||
        _actor == null)
      throw StateError('OUTCOME_UNKNOWN');
    if (economy?['enabled'] != true) throw StateError('ECONOMY_DISABLED');
    if (economyLoading || economyError == 'ECONOMY_UNAVAILABLE')
      throw StateError('ECONOMY_UNAVAILABLE');
    final body =
        _freeze({'operation_id': _id(), 'days': days}) as Map<String, dynamic>;
    _store('economy-pass', body);
    _pendingPass = body;
    passReceipt = null;
    await _sendPass(recover: false);
  }

  Future<void> recoverPass() async {
    if (_pendingPass == null || busy || _storageFault || _disposed) return;
    await _sendPass(recover: true);
  }

  Future<void> _sendPass({required bool recover}) async {
    final actor = _actor;
    final body = _pendingPass!;
    busy = true;
    economyError = null;
    _notify();
    try {
      var send = !recover;
      if (recover) {
        final found = await _request(
          'GET',
          '/v1/economy/passes/${_segment(body['operation_id'] as String)}',
        );
        if (_disposed || actor != _actor) return;
        if (found.status == 200) {
          _passResult(found, retrying: true);
        } else {
          send =
              found.status == 404 &&
              found.data is Map &&
              found.data['code'] == 'OPERATION_NOT_FOUND';
          economyError = 'OUTCOME_UNKNOWN';
        }
      }
      if (send && _pendingPass != null) {
        final response = await _request('POST', '/v1/economy/passes', body);
        if (_disposed || actor != _actor) return;
        _passResult(response, retrying: recover);
      }
      if (_pendingPass == null && economyError == null) {
        await refreshEconomy();
        if (matchId != null) await refresh();
      }
    } catch (_) {
      if (!_disposed && actor == _actor) economyError = 'OUTCOME_UNKNOWN';
    } finally {
      if (!_disposed && actor == _actor) {
        busy = false;
        _notify();
      }
    }
  }

  final OnlineTransport transport;
  AuthoritativeSnapshot? snapshot;
  Map<String, dynamic>? account, room;
  Map<String, dynamic>? _unknown;
  Map<String, dynamic>? _roomOperation;
  String? _roomOperationPath;
  bool get hasPendingRoomOperation => _roomOperation != null;
  Map<String, dynamic>? get unknownOperation => _unknown;
  String? matchId, errorCode;
  String _csrf = '';
  bool busy = false, connected = false, _disposed = false;
  int _generation = 0;
  static const recoveryLimit = 8;
  int sessionAttempt = 0;
  bool sessionLoading = false;
  int _reconnectAttempt = 0;
  int? _reconnectingGeneration;
  bool get _reconnecting => _reconnectingGeneration == _generation;
  Timer? _reconnectTimer, _streamHealthyTimer;
  bool get reconnecting => _reconnecting || _reconnectTimer != null;
  Future<void>? _restoringSession;
  final _readRecoveryWaits = <Timer, Completer<void>>{};
  static Duration _recoveryDelay(int attempt) =>
      Duration(milliseconds: min(4000, 250 * (1 << (attempt - 1))));
  Future<void> _waitToRecover(Duration delay) {
    final completion = Completer<void>();
    late final Timer timer;
    timer = Timer(delay, () {
      _readRecoveryWaits.remove(timer);
      completion.complete();
    });
    _readRecoveryWaits[timer] = completion;
    return completion.future;
  }

  bool get _transientReadError =>
      errorCode == null ||
      const {
        'UNAVAILABLE',
        'OVERLOADED',
        'STREAM_UNAVAILABLE',
      }.contains(errorCode);

  StreamSubscription<Map<String, dynamic>>? _stream;
  AuthoritativeSnapshot? _drawStreamSnapshot;
  Timer? _drawFeedbackTimer;
  TableDrawFeedback? _drawFeedback;
  final _drawFeedbackQueue =
      Queue<({String id, int seat, DrawDestination destination})>();

  /// Ephemeral presentation only. No draw identity/face or command is stored.
  TableDrawFeedback? get drawFeedback => _drawFeedback;

  void _clearDrawFeedback() {
    _drawFeedbackTimer?.cancel();
    _drawFeedbackTimer = null;
    _drawFeedback = null;
    _drawFeedbackQueue.clear();
  }

  void _advanceDrawFeedback() {
    _drawFeedbackTimer?.cancel();
    _drawFeedbackTimer = null;
    _drawFeedback = null;
    if (_drawFeedbackQueue.isNotEmpty) {
      final next = _drawFeedbackQueue.removeFirst();
      _drawFeedback = TableDrawFeedback(
        id: next.id,
        seat: next.seat,
        destination: next.destination,
        startedAt: DateTime.now(),
      );
      _drawFeedbackTimer = Timer(
        TableDrawFeedback.duration,
        _advanceDrawFeedback,
      );
    }
    _notify();
  }

  void _observeLiveDraw(AuthoritativeSnapshot next) {
    final previous = _drawStreamSnapshot;
    if (previous != null && next.cursor <= previous.cursor) return;
    _drawStreamSnapshot = next;
    final current = snapshot;
    // Refresh can race the live stream. Its newer state stays authoritative;
    // preserve a separate contiguous stream baseline so an undelivered draw in
    // the same game is not swallowed by a successful command's refresh.
    if (previous == null ||
        current == null ||
        next.cursor != previous.cursor + 1 ||
        previous.gameId != next.gameId ||
        current.gameId != next.gameId ||
        next.phase != 'playing' ||
        current.phase != 'playing' ||
        previous.board['turn'] != next.board['turn'] ||
        previous.board['active'] != next.board['active'] ||
        previous.online['turn_started'] != false ||
        next.online['turn_started'] != true)
      return;
    final actor = next.board['active'];
    if (actor is! int || actor < 1) return;
    Map? player(AuthoritativeSnapshot view) =>
        (view.board['players'] as List? ?? [])
            .whereType<Map>()
            .where((p) => p['seat'] == actor)
            .firstOrNull;
    final before = player(previous), after = player(next);
    if (before == null || after == null) return;
    final oldHand = before['hand_count'], newHand = after['hand_count'];
    final oldAces = before['ace_count'], newAces = after['ace_count'];
    if (oldHand is! int ||
        newHand is! int ||
        oldAces is! int ||
        newAces is! int)
      return;
    final handChange = newHand - oldHand, aceChange = newAces - oldAces;
    if (!((handChange == 1 && aceChange == 0) ||
        (handChange == 0 && aceChange == 1)))
      return;
    // Several bot turns may arrive in one delivery batch. Queue the already
    // confirmed public cues so Flutter frame coalescing cannot lose a draw.
    // A later live turn is not a replay; reconnect/game closure clears the queue.
    _drawFeedbackQueue.add((
      id: '${next.gameId}/turn-${next.board['turn']}/${next.cursor}',
      seat: actor,
      destination: aceChange == 1
          ? DrawDestination.concealedAce
          : DrawDestination.hand,
    ));
    if (_drawFeedback == null) _advanceDrawFeedback();
  }

  bool get canSubmit =>
      snapshot != null &&
      !_storageFault &&
      connected &&
      !busy &&
      _unknown == null &&
      _roomOperation == null;
  void _notify() {
    if (!_disposed) notifyListeners();
  }

  String _id() => List.generate(
    24,
    (_) => Random.secure().nextInt(256).toRadixString(16).padLeft(2, '0'),
  ).join();
  String _segment(String value) => Uri.encodeComponent(value);
  String get _matchPath => '/v1/matches/${_segment(matchId!)}';
  Future<OnlineResponse> _request(
    String method,
    String path, [
    Map<String, dynamic>? body,
    Duration timeout = const Duration(seconds: 2),
  ]) => transport
      .request(
        method,
        path,
        body: body,
        headers: {
          'Content-Type': 'application/json',
          if (_csrf.isNotEmpty && method != 'GET') 'X-CSRF-Token': _csrf,
        },
      )
      .timeout(timeout);
  String _responseError(OnlineResponse response) {
    if (response.status >= 500) return 'UNAVAILABLE';
    if (response.status == 429) return 'OVERLOADED';
    if (response.status == 401) return 'UNAUTHENTICATED';
    if (response.status == 403) return 'FORBIDDEN';
    final code = response.data is Map ? response.data['code'] : null;
    return code is String &&
            !const {
              'UNAVAILABLE',
              'OVERLOADED',
              'STREAM_UNAVAILABLE',
            }.contains(code)
        ? code
        : 'REQUEST_REJECTED';
  }

  Future<Map<String, dynamic>> _object(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    late final OnlineResponse result;
    try {
      result = await _request(method, path, body);
    } on FormatException {
      errorCode = 'INVALID_RESPONSE';
      rethrow;
    }
    if (result.status != 200) {
      errorCode = _responseError(result);
      _notify();
      throw StateError(errorCode ?? 'UNAVAILABLE');
    }
    try {
      return Map<String, dynamic>.from(result.data as Map);
    } catch (_) {
      errorCode = 'INVALID_RESPONSE';
      throw const FormatException('INVALID_RESPONSE');
    }
  }

  // Only reads are recovered automatically. A mutation with an uncertain
  // outcome retains its original operation and requires receipt reconciliation.
  Future<Map<String, dynamic>> _readObject(
    String path, {
    bool session = false,
  }) async {
    for (var attempt = 1; attempt <= recoveryLimit; attempt++) {
      if (_disposed) throw StateError('DISPOSED');
      errorCode = null;
      if (session) {
        sessionAttempt = attempt;
        _notify();
      }
      try {
        return await _object('GET', path);
      } on FormatException {
        errorCode = 'INVALID_RESPONSE';
        rethrow;
      } catch (_) {
        if (_disposed || !_transientReadError) rethrow;
        if (attempt == recoveryLimit) {
          errorCode = 'RECOVERY_EXHAUSTED';
          _notify();
          rethrow;
        }
        await _waitToRecover(_recoveryDelay(attempt));
      }
    }
    throw StateError('UNAVAILABLE');
  }

  Future<void> _session(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    if (busy ||
        ((_unknown != null || _roomOperation != null || _pendingPass != null) &&
            path != '/v1/session' &&
            path != '/v1/login'))
      throw StateError('OUTCOME_UNKNOWN');
    final result = method == 'GET'
        ? await _readObject(path, session: true)
        : await _object(method, path, body);
    if (_disposed) return;
    late final Map<String, dynamic> nextAccount;
    late final String nextCsrf;
    try {
      final rawAccount = result['account'];
      final rawCsrf = result['csrf_token'];
      if (rawAccount is! Map ||
          rawAccount['id'] is! String ||
          (rawAccount['id'] as String).isEmpty ||
          rawCsrf is! String ||
          rawCsrf.isEmpty) {
        throw const FormatException('INVALID_RESPONSE');
      }
      nextAccount = _freeze(rawAccount) as Map<String, dynamic>;
      nextCsrf = rawCsrf;
    } catch (_) {
      errorCode = 'INVALID_RESPONSE';
      _notify();
      throw const FormatException('INVALID_RESPONSE');
    }
    if (account?['id'] != nextAccount['id']) {
      // Prior actor's unresolved records remain in their isolated store.
      _unknown = null;
      _roomOperation = null;
      _roomOperationPath = null;
      _pendingPass = null;
      economy = null;
      passReceipt = null;
      economyError = null;
      economyLoading = false;
      _economyGeneration++;
      _storageFault = false;
      disconnect();
      snapshot = null;
      room = null;
      matchId = null;
    }
    account = nextAccount;
    _csrf = nextCsrf;
    errorCode = null;
    _restoreRoomOperation();
    _restorePass();
    _notify();
  }

  Future<void> guest() => _session('POST', '/v1/guests', {});
  Future<void> restoreSession() => _restoringSession ??= _restoreSession()
      .whenComplete(() => _restoringSession = null);
  Future<void> _restoreSession() async {
    sessionLoading = true;
    _notify();
    try {
      await _session('GET', '/v1/session');
    } finally {
      sessionLoading = false;
      _notify();
    }
  }

  Future<void> authenticate(String action, String username, String password) {
    if (!['login', 'register', 'upgrade'].contains(action))
      throw ArgumentError.value(action);
    return _session('POST', '/v1/$action', {
      'username': username,
      'password': password,
    });
  }

  Future<void> logout({bool deleteAccount = false}) async {
    if (_unknown != null ||
        _roomOperation != null ||
        _pendingPass != null ||
        busy)
      throw StateError('OUTCOME_UNKNOWN');
    await _object(
      deleteAccount ? 'DELETE' : 'POST',
      deleteAccount ? '/v1/account' : '/v1/logout',
      {},
    );
    disconnect();
    account = null;
    economy = null;
    passReceipt = null;
    economyError = null;
    economyLoading = false;
    _economyGeneration++;
    room = null;
    snapshot = null;
    _csrf = '';
    _notify();
  }

  Future<Map<String, dynamic>> _roomMutation(
    String path,
    Map<String, dynamic> terms,
  ) async {
    if (_storageFault || busy || _unknown != null || _pendingPass != null)
      throw StateError('OUTCOME_UNKNOWN');
    final retrying = _roomOperation != null;
    if (_roomOperation != null) {
      final prior = Map<String, dynamic>.from(_roomOperation!)
        ..remove('command_id');
      if (path != _roomOperationPath || jsonEncode(prior) != jsonEncode(terms))
        throw StateError('OUTCOME_UNKNOWN');
    } else {
      _roomOperation = _freeze({'command_id': _id(), ...terms});
      _roomOperationPath = path;
    }
    _store('room', {'path': path, 'body': _roomOperation});
    busy = true;
    _notify();
    try {
      final response = await _request('POST', path, _roomOperation);
      if (response.status == 200) {
        final result = Map<String, dynamic>.from(response.data as Map);
        _clearRoomOperation();
        errorCode = null;
        return result;
      }
      errorCode = response.data is Map
          ? response.data['code'] as String?
          : 'OUTCOME_UNKNOWN';
      if (!retrying &&
          response.status >= 400 &&
          response.status < 500 &&
          response.status != 429) {
        _clearRoomOperation();
      }
      throw StateError(errorCode ?? 'OUTCOME_UNKNOWN');
    } catch (_) {
      errorCode ??= 'OUTCOME_UNKNOWN';
      rethrow;
    } finally {
      busy = false;
      _notify();
    }
  }

  Future<void> createRoom(
    int capacity,
    int games, {
    String? botDifficulty,
  }) async {
    if (botDifficulty != null && !_botDifficulties.contains(botDifficulty)) {
      throw ArgumentError.value(botDifficulty, 'botDifficulty');
    }
    room = _freeze(
      await _roomMutation('/v1/rooms', {
        'capacity': capacity,
        'games_per_match': games,
        if (botDifficulty != null) 'bot_difficulty': botDifficulty,
      }),
    );
    _notify();
  }

  Future<void> refreshRoom(String id) async {
    room = _freeze(await _readObject('/v1/rooms/${_segment(id)}'));
    _notify();
  }

  Future<void> joinRoom(String id, String invitation) async {
    room = _freeze(
      await _object('POST', '/v1/rooms/${_segment(id)}/join', {
        'invitation': invitation,
      }),
    );
    _notify();
  }

  Future<String> invite() async =>
      (await _object(
            'POST',
            '/v1/rooms/${_segment(room!['id'] as String)}/invitations',
            {},
          ))['invitation']
          as String;
  Future<void> startMatch() async {
    final result = await _roomMutation(
      '/v1/rooms/${_segment(room!['id'] as String)}/matches',
      {},
    );
    await watchMatch(result['match_id'] as String);
  }

  Future<void> retryPendingRoomOperation() async {
    if (_roomOperation == null || _roomOperationPath == null) return;
    final path = _roomOperationPath!;
    final terms = Map<String, dynamic>.from(_roomOperation!)
      ..remove('command_id');
    final result = await _roomMutation(path, terms);
    if (path == '/v1/rooms') {
      room = _freeze(result);
      _notify();
    } else {
      await watchMatch(result['match_id'] as String);
    }
  }

  Future<void> watchMatch(String id) async {
    if (_unknown != null && id != matchId) throw StateError('OUTCOME_UNKNOWN');
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempt = 0;
    _clearDrawFeedback();
    _drawStreamSnapshot = null;
    await _stream?.cancel();
    _generation++;
    matchId = id;
    _restoreMatchOperation();
    snapshot = null;
    final generation = _generation;
    await refresh();
    if (!_disposed && generation == _generation) {
      if (connected)
        _listen();
      else
        _scheduleReconnect();
    }
  }

  Future<void> refresh() async {
    if (matchId == null) return;
    final generation = _generation;
    var receivedResponse = false;
    try {
      final response = await _request('GET', _matchPath);
      receivedResponse = true;
      if (_disposed || generation != _generation) return;
      if (response.status != 200) {
        errorCode = _responseError(response);
        throw StateError(errorCode ?? 'UNAVAILABLE');
      }
      late final AuthoritativeSnapshot value;
      try {
        value = AuthoritativeSnapshot.fromJson(
          Map<String, dynamic>.from(response.data as Map),
        );
      } catch (_) {
        throw const FormatException('INVALID_RESPONSE');
      }
      if (generation == _generation &&
          !_disposed &&
          (snapshot == null || value.version >= snapshot!.version)) {
        if (snapshot?.gameId != value.gameId || value.phase != 'playing')
          _clearDrawFeedback();
        snapshot = value;
        connected = true;
        if (!_storageFault) errorCode = null;
        _notify();
      }
    } on FormatException {
      if (!_disposed && generation == _generation) {
        errorCode = 'INVALID_RESPONSE';
        connected = false;
        _notify();
      }
    } catch (_) {
      if (!_disposed && generation == _generation) {
        if (!receivedResponse) errorCode = 'UNAVAILABLE';
        errorCode ??= 'UNAVAILABLE';
        connected = false;
        _notify();
        _scheduleReconnect();
      }
    }
  }

  void _listen() {
    final generation = _generation;
    // A new connection establishes a baseline, never replays historical draws.
    _drawStreamSnapshot = snapshot;
    _stream = transport
        .events('$_matchPath/stream?after=${snapshot?.cursor ?? 0}')
        .listen(
          (frame) async {
            if (_disposed || generation != _generation) return;
            if (frame['type'] == 'transport-open') {
              // Reset consecutive failures only after an actually opened socket
              // remains healthy for five seconds. HTTP success alone cannot
              // reset a permanently broken WebSocket's bounded retry budget.
              _streamHealthyTimer?.cancel();
              _streamHealthyTimer = Timer(const Duration(seconds: 5), () {
                if (!_disposed && generation == _generation && connected) {
                  _reconnectAttempt = 0;
                }
              });
              return;
            }
            connected = true;
            try {
              if (frame['type'] == 'error') {
                errorCode = frame['code'] as String?;
                await refresh();
                return;
              }
              final cursor = frame['cursor'] as int;
              if (cursor <= (snapshot?.cursor ?? -1) &&
                  cursor <= (_drawStreamSnapshot?.cursor ?? -1))
                return;
              final next = AuthoritativeSnapshot.fromJson(
                Map<String, dynamic>.from(frame['data'] as Map),
              );
              if (next.cursor != cursor)
                throw const FormatException('INVALID_CURSOR');
              _observeLiveDraw(next);
              if (cursor <= (snapshot?.cursor ?? -1)) return;
              if (snapshot == null || cursor != snapshot!.cursor + 1) {
                await refresh();
                return;
              }
              _reconnectAttempt = 0;
              if (snapshot?.gameId != next.gameId || next.phase != 'playing')
                _clearDrawFeedback();
              snapshot = next;
              _notify();
              // Admission depends on the current roster's access, so it is a
              // snapshot advisory rather than immutable game-event data.
              // Fetch it at the boundary even when the durable cursor is unchanged.
              final completed = next.online['completed_games'] as int?;
              final limit = next.online['games_per_match'] as int?;
              final financialPhase = next.online['financial_phase'];
              if ((financialPhase == 'finalized' || financialPhase == 'void') &&
                  completed != null &&
                  limit != null &&
                  completed < limit) {
                await refresh();
              }
            } catch (_) {
              errorCode = 'RESNAPSHOT_REQUIRED';
              await refresh();
            }
          },
          onError: (Object _) {
            if (_disposed || generation != _generation) return;
            connected = false;
            errorCode = 'STREAM_UNAVAILABLE';
            _streamHealthyTimer?.cancel();
            _notify();
            _scheduleReconnect();
          },
          onDone: () {
            if (_disposed || generation != _generation) return;
            connected = false;
            errorCode ??= 'STREAM_UNAVAILABLE';
            _streamHealthyTimer?.cancel();
            _notify();
            _scheduleReconnect();
          },
        );
  }

  void _scheduleReconnect() {
    if (_disposed ||
        matchId == null ||
        _reconnecting ||
        _reconnectTimer != null ||
        !_transientReadError)
      return;
    if (_reconnectAttempt >= recoveryLimit) {
      errorCode = 'RECOVERY_EXHAUSTED';
      _notify();
      return;
    }
    final generation = _generation;
    _reconnectTimer = Timer(_recoveryDelay(++_reconnectAttempt), () async {
      _reconnectTimer = null;
      if (_disposed || generation != _generation) return;
      final owner = _generation + 1;
      _reconnectingGeneration = owner;
      try {
        await _reconnect();
      } finally {
        if (_reconnectingGeneration == owner) _reconnectingGeneration = null;
        if (_generation == owner && !connected) _scheduleReconnect();
      }
    });
  }

  Future<void> reconnect() async {
    if (_reconnecting || _disposed) return;
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempt = 0;
    final owner = _generation + 1;
    _reconnectingGeneration = owner;
    try {
      await _reconnect();
    } finally {
      if (_reconnectingGeneration == owner) _reconnectingGeneration = null;
      if (_generation == owner && !connected) _scheduleReconnect();
    }
  }

  Future<void> _reconnect() async {
    _streamHealthyTimer?.cancel();
    _clearDrawFeedback();
    _drawStreamSnapshot = null;
    _generation++;
    final generation = _generation;
    await _stream?.cancel();
    if (_disposed || generation != _generation) return;
    errorCode = null;
    await refresh();
    if (!_disposed && connected && matchId != null && generation == _generation)
      _listen();
  }

  Future<void> submit(String type, Map<String, dynamic> payload) async {
    if (!canSubmit) throw StateError('OUTCOME_UNKNOWN');
    validateCommandPayload(type, payload);
    // JSON validation rejects unserializable UI objects before admission.
    final body =
        _freeze(
              jsonDecode(
                jsonEncode({
                  'command_id': _id(),
                  'game_id': snapshot!.gameId,
                  'expected_version': snapshot!.version,
                  'type': type,
                  'payload': payload,
                }),
              ),
            )
            as Map<String, dynamic>;
    _store('match:$matchId', body);
    _unknown = body;
    busy = true;
    errorCode = null;
    _notify();
    var uncertain = false;
    final deadline = DateTime.now().add(const Duration(seconds: 5));
    try {
      for (var attempt = 0; attempt < 2 && _unknown != null; attempt++) {
        final remaining = deadline.difference(DateTime.now());
        if (remaining <= Duration.zero) break;
        try {
          final result = await _request(
            'POST',
            '$_matchPath/commands',
            body,
            remaining < const Duration(seconds: 2)
                ? remaining
                : const Duration(seconds: 2),
          );
          if (_acceptResult(result, definitiveRejection: !uncertain)) break;
          if (result.status >= 500) uncertain = true;
        } catch (_) {
          uncertain = true;
          errorCode = 'OUTCOME_UNKNOWN';
        }
        if (_unknown == null) break;
        final budget = deadline.difference(DateTime.now());
        if (budget <= Duration.zero) break;
        await _lookup(
          budget < const Duration(seconds: 2)
              ? budget
              : const Duration(seconds: 2),
        );
      }
      if (_unknown != null)
        errorCode = 'OUTCOME_UNKNOWN';
      else if (errorCode == null)
        await refresh();
    } finally {
      busy = false;
      _notify();
    }
  }

  bool _acceptResult(OnlineResponse result, {bool definitiveRejection = true}) {
    if (result.status == 200) {
      // Reconciliation may be historical. Never replace newer projected state.
      if (result.data is! Map ||
          result.data['version'] is! int ||
          result.data['game_id'] is! String ||
          result.data['status'] is! String) {
        errorCode = 'OUTCOME_UNKNOWN';
        return false;
      }
      _clearUnknown();
      errorCode = null;
      return true;
    }
    errorCode = result.data is Map
        ? result.data['code'] as String?
        : 'OUTCOME_UNKNOWN';
    if (definitiveRejection &&
        result.status >= 400 &&
        result.status < 500 &&
        result.status != 429) {
      _clearUnknown();
      return true;
    }
    return false;
  }

  Future<void> _lookup(Duration budget) async {
    if (_unknown == null) return;
    try {
      final result = await _request(
        'GET',
        '$_matchPath/commands/${_segment(_unknown!['command_id'] as String)}',
        null,
        budget,
      );
      // Absence or failed authorization is not proof the original could not commit.
      if (result.status == 200) _acceptResult(result);
    } catch (_) {
      errorCode = 'OUTCOME_UNKNOWN';
    }
  }

  /// Explicit user recovery after the automatic budget is exhausted. Lookup
  /// precedes one identical submission; no snapshot can rewrite this envelope.
  Future<void> retryUnknownOperation() async {
    if (_unknown == null || busy || _storageFault || _disposed) return;
    final body = _unknown!;
    final path = _matchPath;
    busy = true;
    errorCode = null;
    _notify();
    final deadline = DateTime.now().add(const Duration(seconds: 5));
    Duration budget() {
      final remaining = deadline.difference(DateTime.now());
      return remaining < const Duration(seconds: 2)
          ? remaining
          : const Duration(seconds: 2);
    }

    try {
      await _lookup(budget());
      if (_unknown != null && !_disposed && budget() > Duration.zero) {
        try {
          final result = await _request(
            'POST',
            '$path/commands',
            body,
            budget(),
          );
          final code = result.data is Map ? result.data['code'] : null;
          // These outcomes follow the server's locked duplicate lookup and
          // command validation. Authentication failures cannot establish absence.
          final definitive =
              (result.status == 409 && code == 'STATE_CONFLICT') ||
              (result.status == 400 &&
                  (code == 'INVALID_COMMAND' || code == 'INVALID_REQUEST'));
          _acceptResult(result, definitiveRejection: definitive);
        } catch (_) {
          errorCode = 'OUTCOME_UNKNOWN';
        }
        if (_unknown != null && !_disposed && budget() > Duration.zero)
          await _lookup(budget());
      }
      if (_unknown != null) {
        errorCode = 'OUTCOME_UNKNOWN';
      } else if (errorCode == null && !_disposed) {
        await refresh();
      }
    } finally {
      busy = false;
      _notify();
    }
  }

  Future<void> reconcile() async {
    if (_unknown == null || busy) return;
    busy = true;
    _notify();
    try {
      await _lookup(const Duration(seconds: 2));
      if (_unknown == null) await refresh();
    } finally {
      busy = false;
      _notify();
    }
  }

  void disconnect() {
    _clearDrawFeedback();
    _drawStreamSnapshot = null;
    _streamHealthyTimer?.cancel();
    _generation++;
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempt = 0;
    _stream?.cancel();
    _stream = null;
    connected = false;
  }

  @override
  void dispose() {
    _disposed = true;
    for (final entry in _readRecoveryWaits.entries) {
      entry.key.cancel();
      entry.value.complete();
    }
    _readRecoveryWaits.clear();
    disconnect();
    transport.close();
    super.dispose();
  }
}
