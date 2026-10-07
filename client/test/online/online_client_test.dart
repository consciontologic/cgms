import 'dart:async';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:cgms_demo/online/online_client.dart';

Map<String, dynamic> view(int version, {String game = 'game-1'}) => {
  'version': version,
  'cursor': version,
  'projection': {
    'board': {
      'game_id': game,
      'seat': 1,
      'phase': 'playing',
      'cards': [],
      'players': [],
    },
    'online': {},
    'cash': {'numerator': '0', 'denominator': '1'},
    'score': {'numerator': '0', 'denominator': '1'},
  },
};

class FakeTransport implements OnlineTransport {
  final calls = <Map<String, dynamic>>[];
  final streams = <String>[];
  final frames = StreamController<Map<String, dynamic>>.broadcast();
  Future<OnlineResponse> Function(String, String, Map<String, dynamic>?)?
  handle;
  @override
  Future<OnlineResponse> request(
    String method,
    String path, {
    Map<String, dynamic>? body,
    Map<String, String> headers = const {},
  }) async {
    calls.add({
      'method': method,
      'path': path,
      'body': body,
      'headers': headers,
    });
    return handle == null
        ? OnlineResponse(200, view(1))
        : handle!(method, path, body);
  }

  @override
  Stream<Map<String, dynamic>> events(String path) {
    streams.add(path);
    return frames.stream;
  }

  @override
  void close() {}
}

Map<String, dynamic> drawView(
  int version, {
  bool started = false,
  int turn = 1,
  int active = 1,
  int hand = 10,
  int aces = 1,
  String game = 'game-1',
}) {
  final result = view(version, game: game);
  result['projection']['online']['turn_started'] = started;
  result['projection']['board'].addAll({
    'turn': turn,
    'active': active,
    'players': [
      for (var seat = 1; seat <= 3; seat++)
        {'seat': seat, 'hand_count': hand, 'ace_count': aces},
    ],
  });
  return result;
}

void main() {
  testWidgets(
    'scheduled live draw emits once and refresh/reconnect never replay',
    (tester) async {
      var current = drawView(1);
      final transport = FakeTransport()
        ..handle = (_, _, _) async => OnlineResponse(200, current);
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await client.watchMatch('match');
      expect(client.drawFeedback, isNull);
      current = drawView(2, started: true, hand: 11);
      transport.frames.add({'cursor': 2, 'data': current});
      await tester.pump();
      final draw = client.drawFeedback;
      expect(draw, isNotNull);
      expect(draw!.seat, 1);
      expect(draw.destination.name, 'hand');
      transport.frames.add({'cursor': 2, 'data': current});
      await tester.pump();
      await client.refresh();
      expect(identical(client.drawFeedback, draw), isTrue);
      await tester.pump(const Duration(seconds: 2));
      expect(client.drawFeedback, isNull);
      await client.refresh();
      await tester.runAsync(client.reconnect);
      expect(client.drawFeedback, isNull);
      current = drawView(3, turn: 2, active: 2);
      transport.frames.add({'cursor': 3, 'data': current});
      await tester.pump();
      current = drawView(4, turn: 2, active: 2, started: true, aces: 2);
      transport.frames.add({'cursor': 4, 'data': current});
      await tester.pump();
      expect(client.drawFeedback!.seat, 2);
      expect(client.drawFeedback!.destination.name, 'concealedAce');
      expect(transport.calls.every((c) => c['method'] == 'GET'), isTrue);
    },
  );

  testWidgets('routine refresh cannot consume an undelivered live draw', (
    tester,
  ) async {
    var current = drawView(1);
    final transport = FakeTransport()
      ..handle = (_, _, _) async => OnlineResponse(200, current);
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('match');
    current = drawView(2, started: true, hand: 11);
    await client.refresh();
    expect(client.drawFeedback, isNull);
    transport.frames.add({'cursor': 2, 'data': current});
    await tester.pump();
    final draw = client.drawFeedback;
    expect(draw, isNotNull);
    expect(client.snapshot!.cursor, 2);
    transport.frames.add({'cursor': 2, 'data': current});
    await tester.pump();
    expect(identical(draw, client.drawFeedback), isTrue);
    await tester.pump(const Duration(seconds: 2));
    expect(client.drawFeedback, isNull);
  });

  testWidgets(
    'one live frame batch presents each scheduled draw exactly once',
    (tester) async {
      final semantics = tester.ensureSemantics();
      var current = drawView(1);
      final transport = FakeTransport()
        ..handle = (_, _, _) async => OnlineResponse(200, current);
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await client.watchMatch('match');
      await tester.pumpWidget(
        MaterialApp(
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Center(
            child: AnimatedBuilder(
              animation: client,
              builder: (_, _) => CgmsDrawFeedback(
                feedback: client.drawFeedback,
                viewerSeat: 1,
                child: const SizedBox(width: 48, height: 60),
              ),
            ),
          ),
        ),
      );
      for (var seat = 1; seat <= 3; seat++) {
        if (seat > 1) {
          transport.frames.add({
            'cursor': seat * 2 - 1,
            'data': drawView(seat * 2 - 1, turn: seat, active: seat),
          });
        }
        current = drawView(
          seat * 2,
          turn: seat,
          active: seat,
          started: true,
          hand: 11,
        );
        transport.frames.add({'cursor': seat * 2, 'data': current});
        transport.frames.add({'cursor': seat * 2, 'data': current});
      }
      await tester.pump();
      expect(client.snapshot!.cursor, 6);
      expect(client.drawFeedback!.seat, 1);
      expect(
        find.bySemanticsLabel('Turn draw: one card added to your hand.'),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('draw-highlight')), findsOneWidget);
      final first = client.drawFeedback;
      await client.refresh();
      expect(identical(client.drawFeedback, first), isTrue);
      for (var seat = 2; seat <= 3; seat++) {
        await tester.pump(const Duration(milliseconds: 1100));
        expect(client.drawFeedback!.seat, seat);
        expect(
          find.bySemanticsLabel('Seat $seat drew one card.'),
          findsOneWidget,
        );
        expect(find.byKey(const ValueKey('draw-highlight')), findsOneWidget);
      }
      await tester.pump(const Duration(milliseconds: 1100));
      expect(client.drawFeedback, isNull);
      expect(find.byKey(const ValueKey('draw-highlight')), findsNothing);
      await client.refresh();
      expect(client.drawFeedback, isNull);
      expect(transport.calls.every((c) => c['method'] == 'GET'), isTrue);
      semantics.dispose();
    },
  );

  testWidgets(
    'a same-game refresh ahead of unseen live draws does not replay or lose them',
    (tester) async {
      var current = drawView(1);
      final transport = FakeTransport()
        ..handle = (_, _, _) async => OnlineResponse(200, current);
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await client.watchMatch('match');
      current = drawView(4, turn: 2, active: 2, started: true, aces: 2);
      await client.refresh();
      expect(client.drawFeedback, isNull);
      for (final frame in [
        drawView(2, started: true, hand: 11),
        drawView(3, turn: 2, active: 2),
        current,
      ]) {
        transport.frames.add({'cursor': frame['cursor'], 'data': frame});
      }
      await tester.pump();
      expect(client.drawFeedback!.seat, 1);
      await tester.pump(const Duration(milliseconds: 1100));
      expect(client.drawFeedback!.seat, 2);
      expect(client.drawFeedback!.destination.name, 'concealedAce');
      await tester.pump(const Duration(milliseconds: 1100));
      expect(client.drawFeedback, isNull);
      expect(client.snapshot!.cursor, 4);
    },
  );

  for (final boundary in ['reconnect', 'new-game']) {
    testWidgets('$boundary clears both displayed and queued draw feedback', (
      tester,
    ) async {
      var current = drawView(1);
      final transport = FakeTransport()
        ..handle = (_, _, _) async => OnlineResponse(200, current);
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await client.watchMatch('match');
      for (final frame in [
        drawView(2, started: true, hand: 11),
        drawView(3, turn: 2, active: 2),
        drawView(4, turn: 2, active: 2, started: true, hand: 11),
      ]) {
        current = frame;
        transport.frames.add({'cursor': frame['cursor'], 'data': frame});
      }
      await tester.pump();
      expect(client.drawFeedback!.seat, 1);
      if (boundary == 'reconnect') {
        await tester.runAsync(client.reconnect);
      } else {
        current = drawView(5, game: 'new-game');
        await client.refresh();
      }
      expect(client.drawFeedback, isNull);
      await tester.pump(const Duration(seconds: 5));
      expect(client.drawFeedback, isNull);
    });
  }

  for (final kind in ['trade', 'exhausted', 'gap', 'old-game', 'legacy']) {
    testWidgets('$kind cannot be mistaken for a newly confirmed turn draw', (
      tester,
    ) async {
      var current = drawView(1, started: kind == 'trade');
      if (kind == 'legacy') current['projection']['online'].clear();
      final transport = FakeTransport()
        ..handle = (_, _, _) async => OnlineResponse(200, current);
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await client.watchMatch('match');
      final draw = drawView(
        kind == 'gap' ? 3 : 2,
        started: true,
        hand: kind == 'exhausted' ? 10 : 11,
      );
      if (kind == 'old-game') {
        current = drawView(
          4,
          turn: 2,
          game: kind == 'old-game' ? 'game-2' : 'game-1',
        );
        await client.refresh();
      } else {
        current = draw;
      }
      transport.frames.add({'cursor': draw['cursor'], 'data': draw});
      await tester.pump();
      expect(client.drawFeedback, isNull);
      expect(transport.calls.every((c) => c['method'] == 'GET'), isTrue);
    });
  }

  test(
    'bot room difficulty is explicit while human requests stay unchanged',
    () async {
      final transport = FakeTransport()
        ..handle = (_, _, _) async => const OnlineResponse(200, {'id': 'room'});
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      for (final difficulty in ['beginner', 'standard', 'advanced']) {
        await client.createRoom(4, 3, botDifficulty: difficulty);
        expect(
          transport.calls.last['body'],
          containsPair('bot_difficulty', difficulty),
        );
        expect(transport.calls.last['body'], containsPair('capacity', 4));
        expect(
          transport.calls.last['body'],
          containsPair('games_per_match', 3),
        );
      }
      await client.createRoom(3, 3);
      expect((transport.calls.last['body'] as Map).keys.toSet(), {
        'command_id',
        'capacity',
        'games_per_match',
      });
    },
  );

  test(
    'bot room validation and pending retry preserve the original difficulty',
    () async {
      final transport = FakeTransport();
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      Future<void> bots(String difficulty) =>
          client.createRoom(3, 1, botDifficulty: difficulty);
      await expectLater(bots('unrecognized'), throwsArgumentError);
      expect(transport.calls, isEmpty);
      transport.handle = (_, _, _) async => throw TimeoutException('lost');
      await expectLater(bots('beginner'), throwsA(isA<TimeoutException>()));
      final original = transport.calls.last['body'];
      await expectLater(bots('advanced'), throwsStateError);
      await expectLater(client.createRoom(3, 1), throwsStateError);
      expect(transport.calls, hasLength(1));
      transport.handle = (_, _, _) async => const OnlineResponse(200, {
        'id': 'bot-room',
        'bot_difficulty': 'beginner',
      });
      await client.retryPendingRoomOperation();
      expect(transport.calls.last['body'], original);
      expect(client.room?['bot_difficulty'], 'beginner');
      expect(client.hasPendingRoomOperation, isFalse);
    },
  );

  testWidgets('session bootstrap recovers delayed authority without mutation', (
    tester,
  ) async {
    final transport = FakeTransport();
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    var attempts = 0;
    transport.handle = (method, path, body) async {
      attempts++;
      if (attempts < 3)
        return const OnlineResponse(503, {'code': 'UNAVAILABLE'});
      return const OnlineResponse(200, {
        'account': {'id': 'existing'},
        'csrf_token': 'private',
      });
    };
    Object? failure;
    final loading = client.restoreSession().catchError((Object error) {
      failure = error;
    });
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 250));
    await tester.pump(const Duration(milliseconds: 500));
    await loading;
    expect(failure, isNull);
    expect(client.account?['id'], 'existing');
    expect(attempts, 3);
    expect(transport.calls.every((call) => call['method'] == 'GET'), isTrue);
  });

  test(
    'reloading an uncertain bot room preserves its difficulty and command identity',
    () async {
      final store = MemoryOperations();
      final transport = FakeTransport()
        ..handle = (_, _, _) async => const OnlineResponse(200, {
          'account': {'id': 'owner'},
          'csrf_token': 'test-csrf',
        });
      final first = OnlineClient(transport, operationStore: store);
      await first.restoreSession();
      transport.handle = (_, _, _) async => throw TimeoutException('lost');
      await expectLater(
        first.createRoom(4, 3, botDifficulty: 'advanced'),
        throwsA(isA<TimeoutException>()),
      );
      final original = transport.calls.last['body'];
      first.dispose();
      final nextTransport = FakeTransport()
        ..handle = (_, path, _) async => path == '/v1/session'
            ? const OnlineResponse(200, {
                'account': {'id': 'owner'},
                'csrf_token': 'new-test-csrf',
              })
            : const OnlineResponse(200, {
                'id': 'bot-room',
                'bot_difficulty': 'advanced',
              });
      final reloaded = OnlineClient(nextTransport, operationStore: store);
      addTearDown(reloaded.dispose);
      await reloaded.restoreSession();
      expect(reloaded.hasPendingRoomOperation, isTrue);
      expect(nextTransport.calls, hasLength(1));
      await expectLater(
        reloaded.createRoom(4, 3, botDifficulty: 'beginner'),
        throwsStateError,
      );
      await reloaded.retryPendingRoomOperation();
      expect(nextTransport.calls.last['body'], original);
      expect(reloaded.room?['bot_difficulty'], 'advanced');
      expect(store.values, isEmpty);
    },
  );

  testWidgets('session bootstrap stops after eight unavailable reads', (
    tester,
  ) async {
    final transport = FakeTransport()
      ..handle = (_, _, _) async =>
          const OnlineResponse(503, {'code': 'UNAVAILABLE'});
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    Object? failure;
    var finished = false;
    final loading = client
        .restoreSession()
        .catchError((Object error) {
          failure = error;
        })
        .whenComplete(() {
          finished = true;
        });
    await tester.pump();
    for (var i = 0; i < 8; i++) {
      await tester.pump(const Duration(seconds: 4));
    }
    await loading;
    expect(finished, isTrue);
    expect(failure, isA<StateError>());
    expect(transport.calls, hasLength(8));
    await tester.pump(const Duration(minutes: 1));
    expect(transport.calls, hasLength(8));
  });

  testWidgets(
    'anonymous bootstrap is terminal and concurrent restore is shared',
    (tester) async {
      final transport = FakeTransport()
        ..handle = (_, _, _) async =>
            const OnlineResponse(401, {'code': 'UNAUTHENTICATED'});
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      final first = client.restoreSession();
      final second = client.restoreSession();
      expect(identical(first, second), isTrue);
      final result = first.catchError((Object _) {});
      await tester.pump();
      await result;
      await tester.pump(const Duration(minutes: 1));
      expect(transport.calls, hasLength(1));
      expect(client.errorCode, 'UNAUTHENTICATED');
      expect(client.sessionLoading, isFalse);
    },
  );

  testWidgets(
    'disposing cancels pending bootstrap recovery and further reads',
    (tester) async {
      final transport = FakeTransport()
        ..handle = (_, _, _) async => throw TimeoutException('offline');
      final client = OnlineClient(transport);
      final loading = client.restoreSession().catchError((Object _) {});
      await tester.pump();
      client.dispose();
      await tester.pump();
      await loading;
      await tester.pump(const Duration(minutes: 1));
      expect(transport.calls, hasLength(1));
    },
  );

  testWidgets('stream outage automatically reconnects with reads only', (
    tester,
  ) async {
    final transport = FakeTransport();
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('m');
    var recoveryReads = 0;
    transport.handle = (_, _, _) async {
      recoveryReads++;
      return recoveryReads < 3
          ? const OnlineResponse(503, {'code': 'UNAVAILABLE'})
          : OnlineResponse(200, view(4));
    };
    transport.frames.addError(StateError('connection lost'));
    await tester.pump();
    expect(client.connected, isFalse);
    for (var i = 0; i < 3; i++) {
      await tester.pump(const Duration(seconds: 4));
      // Stream cancellation completes outside the widget clock.
      await tester.runAsync(() => Future<void>.delayed(Duration.zero));
      await tester.pump();
    }
    expect(recoveryReads, 3, reason: 'scheduled read recovery');
    expect(client.connected, isTrue);
    expect(client.snapshot?.version, 4);
    expect(transport.streams, hasLength(2));
    expect(transport.calls.every((call) => call['method'] == 'GET'), isTrue);
  });

  testWidgets(
    'bad frame followed by offline refresh still schedules recovery',
    (tester) async {
      final transport = FakeTransport();
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      await client.watchMatch('m');
      transport.handle = (_, _, _) async => throw TimeoutException('offline');
      transport.frames.add({'cursor': 'bad cursor'});
      await tester.pump();
      expect(client.errorCode, 'UNAVAILABLE');
      transport.handle = (_, _, _) async => OnlineResponse(200, view(4));
      await tester.pump(const Duration(milliseconds: 250));
      await tester.runAsync(() => Future<void>.delayed(Duration.zero));
      await tester.pump();
      expect(client.connected, isTrue);
      expect(client.snapshot?.version, 4);
      expect(transport.streams, hasLength(2));
    },
  );

  testWidgets('healthy quiet sockets reset consecutive outage budget', (
    tester,
  ) async {
    final transport = FakeTransport();
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('m');
    for (var outage = 0; outage < 9; outage++) {
      transport.frames.add({'type': 'transport-open'});
      await tester.pump();
      await tester.pump(const Duration(seconds: 5));
      transport.frames.addError(StateError('offline'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 250));
      await tester.runAsync(() => Future<void>.delayed(Duration.zero));
      await tester.pump();
      expect(client.connected, isTrue, reason: 'outage $outage');
    }
    expect(transport.streams, hasLength(10));
  });

  testWidgets('briefly open broken sockets exhaust a bounded recovery budget', (
    tester,
  ) async {
    final transport = FakeTransport();
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('m');
    for (var outage = 0; outage < 9; outage++) {
      transport.frames.add({'type': 'transport-open'});
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 4900));
      transport.frames.addError(StateError('offline'));
      await tester.pump();
      await tester.pump(const Duration(seconds: 4));
      await tester.runAsync(() => Future<void>.delayed(Duration.zero));
      await tester.pump();
    }
    expect(client.errorCode, 'RECOVERY_EXHAUSTED');
    expect(client.connected, isFalse);
    expect(transport.streams, hasLength(9));
    await tester.pump(const Duration(minutes: 1));
    expect(transport.streams, hasLength(9));
  });

  testWidgets('explicit disconnect fences an in-flight automatic reconnect', (
    tester,
  ) async {
    final transport = FakeTransport();
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('m');
    final pending = Completer<OnlineResponse>();
    transport.handle = (_, _, _) => pending.future;
    transport.frames.addError(StateError('offline'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 250));
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    await tester.pump();
    expect(transport.calls, hasLength(2));
    client.disconnect();
    pending.complete(OnlineResponse(200, view(4)));
    await tester.pump();
    await tester.pump(const Duration(minutes: 1));
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    await tester.pump();
    expect(client.connected, isFalse);
    expect(client.snapshot?.version, 1);
    expect(transport.calls, hasLength(2));
    expect(transport.streams, hasLength(1));
  });

  testWidgets('automatic reconnect retains an unknown command without replay', (
    tester,
  ) async {
    final transport = FakeTransport();
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('m');
    transport.handle = (_, _, _) async =>
        const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
    await client.submit('end-turn', {});
    final pending = client.unknownOperation;
    expect(pending, isNotNull);
    final postCount = transport.calls
        .where((call) => call['method'] == 'POST')
        .length;
    transport.handle = (_, _, _) async => OnlineResponse(200, view(4));
    transport.frames.addError(StateError('offline'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 250));
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    await tester.pump();
    expect(client.connected, isTrue);
    expect(client.unknownOperation, same(pending));
    expect(client.canSubmit, isFalse);
    expect(
      transport.calls.where((call) => call['method'] == 'POST'),
      hasLength(postCount),
    );
  });

  testWidgets(
    'malformed successful session response is terminal without retry',
    (tester) async {
      for (final invalid in [null, <dynamic>[], 'not an object']) {
        final transport = FakeTransport()
          ..handle = (_, _, _) async => OnlineResponse(200, invalid);
        final client = OnlineClient(transport);
        Object? error;
        final loading = client.restoreSession().catchError((Object value) {
          error = value;
        });
        await tester.pump();
        for (var i = 0; i < 8; i++) {
          await tester.pump(const Duration(seconds: 4));
        }
        await loading;
        expect(error, isNotNull);
        expect(transport.calls, hasLength(1));
        expect(client.errorCode, 'INVALID_RESPONSE');
        client.dispose();
      }
    },
  );

  testWidgets('malformed JSON during snapshot load is terminal without retry', (
    tester,
  ) async {
    final transport = FakeTransport()
      ..handle = (_, _, _) async =>
          throw const FormatException('INVALID_RESPONSE');
    final client = OnlineClient(transport);
    addTearDown(client.dispose);
    await client.watchMatch('m');
    await tester.pump(const Duration(minutes: 1));
    await tester.runAsync(() => Future<void>.delayed(Duration.zero));
    await tester.pump();
    expect(transport.calls, hasLength(1));
    expect(client.errorCode, 'INVALID_RESPONSE');
    expect(client.connected, isFalse);
  });

  testWidgets(
    'malformed session envelope cannot change actor or pending command',
    (tester) async {
      final transport = FakeTransport();
      final client = OnlineClient(transport);
      addTearDown(client.dispose);
      transport.handle = (_, _, _) async => const OnlineResponse(200, {
        'account': {'id': 'original'},
        'csrf_token': 'original-token',
      });
      await client.restoreSession();
      transport.handle = (_, _, _) async => OnlineResponse(200, view(1));
      await client.watchMatch('m');
      transport.handle = (_, _, _) async =>
          const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
      await client.submit('end-turn', {});
      final account = client.account;
      final pending = client.unknownOperation;
      final snapshot = client.snapshot;
      for (final invalid in [
        <String, dynamic>{},
        {
          'account': {'id': 'other'},
          'csrf_token': 123,
        },
        {
          'account': {'id': ''},
          'csrf_token': 'token',
        },
        {
          'account': {'id': 'other'},
        },
      ]) {
        transport.handle = (_, _, _) async => OnlineResponse(200, invalid);
        final before = transport.calls.length;
        await expectLater(
          client.restoreSession(),
          throwsA(isA<FormatException>()),
        );
        expect(transport.calls.length, before + 1);
        expect(client.errorCode, 'INVALID_RESPONSE');
        expect(client.account, same(account));
        expect(client.unknownOperation, same(pending));
        expect(client.snapshot, same(snapshot));
      }
    },
  );

  testWidgets(
    'authentication statuses are terminal without a valid error body',
    (tester) async {
      for (final status in [401, 403]) {
        for (final body in [
          null,
          <String, dynamic>{},
          {'code': 123},
          {'code': 'UNAVAILABLE'},
        ]) {
          final transport = FakeTransport()
            ..handle = (_, _, _) async => OnlineResponse(status, body);
          final client = OnlineClient(transport);
          final loading = client.restoreSession().catchError((Object _) {});
          await tester.pump();
          for (var i = 0; i < 8; i++) {
            await tester.pump(const Duration(seconds: 4));
          }
          await loading;
          expect(transport.calls, hasLength(1));
          expect(
            client.errorCode,
            status == 401 ? 'UNAUTHENTICATED' : 'FORBIDDEN',
          );
          await client.watchMatch('m');
          await tester.pump(const Duration(minutes: 1));
          await tester.runAsync(() => Future<void>.delayed(Duration.zero));
          await tester.pump();
          expect(transport.calls, hasLength(2));
          expect(
            client.errorCode,
            status == 401 ? 'UNAUTHENTICATED' : 'FORBIDDEN',
          );
          client.dispose();
        }
      }
    },
  );

  test(
    'lost response preserves identity and origin while blocking new intent',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('match-1');
      t.handle = (method, path, body) async =>
          const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
      await c.submit('end-turn', {});
      expect(c.unknownOperation, isNotNull);
      final original = c.unknownOperation!;
      expect(original['game_id'], 'game-1');
      expect(c.snapshot!.version, 1);
      expect(c.canSubmit, isFalse);
      final mutations = t.calls.where((v) => v['method'] == 'POST').toList();
      expect(mutations.length, lessThanOrEqualTo(2));
      for (final mutation in mutations) {
        expect(mutation['body'], original);
      }
      await expectLater(c.submit('coup', {}), throwsStateError);
      t.handle = (method, path, body) async => OnlineResponse(
        200,
        path.contains('/commands/')
            ? {
                'version': 2,
                'game_id': 'game-1',
                'status': 'committed',
                'projection': view(2)['projection'],
              }
            : view(3, game: 'game-2'),
      );
      await c.reconcile();
      expect(c.unknownOperation, isNull);
      expect(c.snapshot!.gameId, 'game-2');
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'snapshots are immutable and repeated events ignored; gaps resnapshot',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      expect(
        () => c.snapshot!.board['game_id'] = 'bad',
        throwsUnsupportedError,
      );
      t.frames.add({'cursor': 1, 'data': view(1)});
      await Future<void>.delayed(Duration.zero);
      expect(t.calls.length, 1);
      t.handle = (_, _, _) async => OnlineResponse(200, view(4));
      t.frames.add({'cursor': 4, 'data': view(4)});
      await Future<void>.delayed(Duration.zero);
      expect(c.snapshot!.cursor, 4);
      c.dispose();
      await t.frames.close();
    },
  );
  test('permanent rejection never retries or changes projection', () async {
    final t = FakeTransport();
    final c = OnlineClient(t);
    await c.watchMatch('m');
    t.handle = (_, _, _) async =>
        const OnlineResponse(400, {'code': 'INVALID_COMMAND'});
    await c.submit('purchase', {'quantity': 2, 'payment': <String>[]});
    expect(c.errorCode, 'INVALID_COMMAND');
    expect(c.unknownOperation, isNull);
    expect(t.calls.where((v) => v['method'] == 'POST').length, 1);
    expect(c.snapshot!.version, 1);
    c.dispose();
    await t.frames.close();
  });
  test(
    'disconnect gates actions; refresh never resolves an unknown mutation',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      c.disconnect();
      expect(c.canSubmit, isFalse);
      await c.refresh();
      expect(c.canSubmit, isTrue);
      t.handle = (_, _, _) async =>
          const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
      await c.submit('coup', {});
      t.handle = (_, _, _) async =>
          OnlineResponse(200, view(8, game: 'new-game'));
      await c.refresh();
      expect(c.canSubmit, isFalse);
      expect(c.unknownOperation!['game_id'], 'game-1');
      await expectLater(c.watchMatch('other-match'), throwsStateError);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'mutating caller payload cannot alter request or retry identity',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      final gate = Completer<OnlineResponse>();
      t.handle = (_, _, _) => gate.future;
      final cards = ['first'];
      final sent = c.submit('open-series', {'selection': cards});
      cards.add('second');
      expect(c.unknownOperation!['payload']['selection'], ['first']);
      gate.complete(const OnlineResponse(400, {'code': 'INVALID_COMMAND'}));
      await sent;
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'lost room creation is retried with the same identity and terms',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await expectLater(c.createRoom(3, 3), throwsA(isA<TimeoutException>()));
      final first = t.calls.last['body'];
      await expectLater(c.createRoom(4, 3), throwsStateError);
      t.handle = (_, _, _) async => const OnlineResponse(200, {'id': 'room'});
      await c.createRoom(3, 3);
      expect(t.calls.last['body'], first);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'reload restores original pending body without replay or cross-account access',
    () async {
      final store = MemoryOperations();
      final t = FakeTransport();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'actor-a'},
                'csrf_token': 'private-csrf',
              }
            : view(1),
      );
      final first = OnlineClient(t, operationStore: store);
      await first.restoreSession();
      await first.watchMatch('m');
      t.handle = (_, _, _) async =>
          const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
      await first.submit('end-turn', {});
      final original = first.unknownOperation;
      first.dispose();
      expect(store.values.values.single, isNot(contains('private-csrf')));
      final t2 = FakeTransport();
      t2.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'actor-a'},
                'csrf_token': 'test-csrf',
              }
            : view(9, game: 'later-game'),
      );
      final second = OnlineClient(t2, operationStore: store);
      await second.restoreSession();
      await second.watchMatch('m');
      expect(second.unknownOperation, original);
      expect(second.canSubmit, isFalse);
      expect(t2.calls.where((v) => v['method'] == 'POST'), isEmpty);
      expect(
        () => second.unknownOperation!['payload']['new'] = true,
        throwsUnsupportedError,
      );
      final t3 = FakeTransport();
      t3.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'actor-b'},
                'csrf_token': 'test-csrf',
              }
            : view(9),
      );
      final other = OnlineClient(t3, operationStore: store);
      await other.restoreSession();
      await other.watchMatch('m');
      expect(other.unknownOperation, isNull);
      second.dispose();
      other.dispose();
      await t.frames.close();
      await t2.frames.close();
      await t3.frames.close();
    },
  );
  test(
    'reload retains room-create command ID and rejects changed terms',
    () async {
      final store = MemoryOperations();
      final t = FakeTransport();
      t.handle = (_, _, _) async => const OnlineResponse(200, {
        'account': {'id': 'a'},
        'csrf_token': 'test-csrf',
      });
      final first = OnlineClient(t, operationStore: store);
      await first.restoreSession();
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await expectLater(
        first.createRoom(3, 3),
        throwsA(isA<TimeoutException>()),
      );
      final original = t.calls.last['body'];
      first.dispose();
      final t2 = FakeTransport();
      t2.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : {'id': 'room'},
      );
      final second = OnlineClient(t2, operationStore: store);
      await second.restoreSession();
      expect(second.hasPendingRoomOperation, isTrue);
      await expectLater(second.createRoom(4, 3), throwsStateError);
      await second.createRoom(3, 3);
      expect(t2.calls.last['body'], original);
      expect(store.values, isEmpty);
      second.dispose();
      await t.frames.close();
      await t2.frames.close();
    },
  );
  test(
    'overlapping loads ignore old snapshots and do not duplicate newer stream',
    () async {
      final t = FakeTransport();
      final slow = Completer<OnlineResponse>();
      t.handle = (_, path, _) async => path.endsWith('/a')
          ? slow.future
          : OnlineResponse(200, view(7, game: 'b-game'));
      final client = OnlineClient(t);
      final first = client.watchMatch('a');
      await Future<void>.delayed(Duration.zero);
      await client.watchMatch('b');
      slow.complete(OnlineResponse(200, view(90, game: 'a-game')));
      await first;
      expect(client.snapshot!.gameId, 'b-game');
      expect(t.streams, hasLength(1));
      client.dispose();
      await t.frames.close();
    },
  );
  test(
    'payload credentials and card-face objects never reach recovery storage',
    () async {
      final store = MemoryOperations();
      final t = FakeTransport();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : view(1),
      );
      final client = OnlineClient(t, operationStore: store);
      await client.restoreSession();
      await client.watchMatch('m');
      await expectLater(
        client.submit('offer', {
          'terms': {'to': 2, 'password': 'secret'},
        }),
        throwsArgumentError,
      );
      await expectLater(
        client.submit('open-series', {
          'selection': [
            {'rank': 13, 'suit': 'spades'},
          ],
        }),
        throwsArgumentError,
      );
      expect(store.values, isEmpty);
      expect(t.calls.where((v) => v['method'] == 'POST'), isEmpty);
      client.dispose();
      await t.frames.close();
    },
  );
  test(
    'corrupt recovery data blocks fresh commands instead of forgetting identity',
    () async {
      final store = MemoryOperations()..values['a/match:m'] = '{bad json';
      final t = FakeTransport();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : view(1),
      );
      final client = OnlineClient(t, operationStore: store);
      await client.restoreSession();
      await client.watchMatch('m');
      expect(client.canSubmit, isFalse);
      expect(client.errorCode, 'RECOVERY_UNAVAILABLE');
      client.dispose();
      await t.frames.close();
    },
  );
  test(
    'reload retries match start with original room and command ID',
    () async {
      final store = MemoryOperations();
      final t = FakeTransport();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : {'id': 'r'},
      );
      final first = OnlineClient(t, operationStore: store);
      await first.restoreSession();
      await first.refreshRoom('r');
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await expectLater(first.startMatch(), throwsA(isA<TimeoutException>()));
      final original = t.calls.last['body'];
      first.dispose();
      final t2 = FakeTransport();
      t2.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : path.endsWith('/matches')
            ? {'match_id': 'm'}
            : view(1),
      );
      final second = OnlineClient(t2, operationStore: store);
      await second.restoreSession();
      await second.retryPendingRoomOperation();
      final post = t2.calls.singleWhere((call) => call['method'] == 'POST');
      expect(post['path'], '/v1/rooms/r/matches');
      expect(post['body'], original);
      expect(second.matchId, 'm');
      expect(store.values, isEmpty);
      second.dispose();
      await t.frames.close();
      await t2.frames.close();
    },
  );
  test(
    'authorization loss after an ambiguous submit cannot erase original intent',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      var posts = 0;
      t.handle = (method, _, _) async {
        if (method == 'POST' && posts++ == 0) throw TimeoutException('lost');
        return const OnlineResponse(401, {'code': 'UNAUTHENTICATED'});
      };
      await c.submit('end-turn', {});
      expect(c.unknownOperation, isNotNull);
      expect(c.canSubmit, isFalse);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'room retry authorization failure retains its unresolved durable identity',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await expectLater(c.createRoom(3, 3), throwsA(isA<TimeoutException>()));
      t.handle = (_, _, _) async =>
          const OnlineResponse(401, {'code': 'UNAUTHENTICATED'});
      await expectLater(c.createRoom(3, 3), throwsStateError);
      expect(c.hasPendingRoomOperation, isTrue);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'explicit restored retry reconciles first then sends unchanged old-game intent',
    () async {
      final store = MemoryOperations();
      final t = FakeTransport();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : view(1),
      );
      final first = OnlineClient(t, operationStore: store);
      await first.restoreSession();
      await first.watchMatch('m');
      t.handle = (_, _, _) async =>
          throw TimeoutException('never reached authority');
      await first.submit('end-turn', {});
      final body = first.unknownOperation!;
      first.dispose();
      final t2 = FakeTransport();
      t2.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'test-csrf',
              }
            : view(9, game: 'new-game'),
      );
      final restored = OnlineClient(t2, operationStore: store);
      await restored.restoreSession();
      await restored.watchMatch('m');
      t2.calls.clear();
      t2.handle = (method, path, _) async {
        if (path.contains('/commands/') && method == 'GET')
          return const OnlineResponse(404, {'code': 'NOT_FOUND'});
        if (method == 'POST')
          return const OnlineResponse(200, {
            'version': 2,
            'game_id': 'game-1',
            'status': 'committed',
          });
        return OnlineResponse(200, view(9, game: 'new-game'));
      };
      await restored.retryUnknownOperation();
      expect(t2.calls.first['method'], 'GET');
      expect(t2.calls.first['path'], contains('/commands/'));
      final submitted = t2.calls.singleWhere(
        (call) => call['method'] == 'POST',
      )['body'];
      expect(submitted, body);
      expect(submitted['game_id'], 'game-1');
      expect(submitted['expected_version'], 1);
      expect(restored.unknownOperation, isNull);
      expect(store.values, isEmpty);
      expect(restored.snapshot!.gameId, 'new-game');
      restored.dispose();
      await t.frames.close();
      await t2.frames.close();
    },
  );
  test(
    'explicit retry never resubmits a receipt found by first lookup',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await c.submit('end-turn', {});
      t.calls.clear();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path.contains('/commands/')
            ? {'version': 2, 'game_id': 'game-1', 'status': 'committed'}
            : view(2),
      );
      await c.retryUnknownOperation();
      expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
      expect(c.unknownOperation, isNull);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'explicit stale-game rejection clears only after identical submission',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await c.submit('end-turn', {});
      final original = c.unknownOperation;
      t.calls.clear();
      t.handle = (method, _, _) async => method == 'POST'
          ? const OnlineResponse(409, {'code': 'STATE_CONFLICT'})
          : const OnlineResponse(404, {'code': 'NOT_FOUND'});
      await c.retryUnknownOperation();
      expect(t.calls.last['body'], original);
      expect(c.unknownOperation, isNull);
      expect(c.errorCode, 'STATE_CONFLICT');
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'explicit ambiguous retry remains bounded and preserves original operation',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      await c.watchMatch('m');
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await c.submit('end-turn', {});
      final original = c.unknownOperation;
      t.calls.clear();
      await c.retryUnknownOperation();
      expect(t.calls.where((call) => call['method'] == 'POST'), hasLength(1));
      expect(t.calls.where((call) => call['method'] == 'GET'), hasLength(2));
      expect(c.unknownOperation, original);
      expect(c.errorCode, 'OUTCOME_UNKNOWN');
      c.dispose();
      await t.frames.close();
    },
  );
  test('browser session CSRF is sent on room mutation', () async {
    final t = FakeTransport();
    final c = OnlineClient(t);
    t.handle = (_, path, _) async => OnlineResponse(
      200,
      path == '/v1/guests'
          ? {
              'account': {'id': 'a'},
              'csrf_token': 'csrf',
            }
          : {'id': 'room', 'games': 3},
    );
    await c.guest();
    await c.createRoom(3, 3);
    expect(t.calls.last['headers'], containsPair('X-CSRF-Token', 'csrf'));
    expect(t.calls.last['body'], containsPair('games_per_match', 3));
    c.dispose();
    await t.frames.close();
  });
}

class MemoryOperations implements OnlineOperationStore {
  final values = <String, String>{};
  @override
  String? read(String account, String scope) => values['$account/$scope'];
  @override
  void write(String account, String scope, String value) {
    values['$account/$scope'] = value;
  }

  @override
  void remove(String account, String scope) {
    values.remove('$account/$scope');
  }
}
