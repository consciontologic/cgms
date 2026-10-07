import 'dart:async';
import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:cgms_demo/online/online_client.dart';
import 'online_client_test.dart' show FakeTransport, MemoryOperations, view;

const balance = {
  'enabled': true,
  'free_starts_remaining': 3,
  'resets_at': '2026-10-02T00:00:00Z',
  'reward': 10,
  'pass_day_price': 30,
  'benefits': {
    'dirt': 180,
    'unlimited': false,
    'no_ads': false,
    'pass_until': '1970-01-01T00:00:00Z',
    'premium_until': '1970-01-01T00:00:00Z',
    'legacy_ad_free_until': '1970-01-01T00:00:00Z',
    'cosmetics': <String>[],
  },
};
Map<String, dynamic> receipt(Map<String, dynamic> body) => {
  ...body,
  'cost': (body['days'] as int) * 30,
  'expires_at': '2026-10-07T00:00:00Z',
};
Future<OnlineClient> signedIn(
  FakeTransport t,
  MemoryOperations store, {
  String actor = 'a',
}) async {
  t.handle = (_, _, _) async => OnlineResponse(200, {
    'account': {'id': actor},
    'csrf_token': 'csrf',
  });
  final c = OnlineClient(t, operationStore: store);
  await c.restoreSession();
  t.handle = (_, _, _) async => const OnlineResponse(200, balance);
  await c.refreshEconomy();
  return c;
}

void main() {
  test(
    'malformed account state keeps the last confirmed view until recovery',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      final confirmed = c.economy;
      Map<String, dynamic> benefit(Map<String, dynamic> changes) => {
        ...balance,
        'benefits': {...balance['benefits'] as Map, ...changes},
      };
      final invalid = <Map<String, dynamic>>[
        {...balance, 'benefits': []},
        benefit({'unlimited': null}),
        benefit({'unlimited': 'true'}),
        benefit({'no_ads': null}),
        benefit({'no_ads': 1}),
        benefit({'dirt': -1}),
        benefit({'dirt': 1000000000000001}),
        benefit({'dirt': '180'}),
        benefit({'pass_until': 'not-a-date'}),
        benefit({'premium_until': 42}),
        benefit({'legacy_ad_free_until': null}),
        benefit({
          'cosmetics': [42],
        }),
        benefit({'ad_free_until': null}),
        benefit({'ad_free_until': 'not-a-date'}),
        benefit({'ad_free_until': '2030-02-30T00:00:00Z'}),
        benefit({'ad_free_until': '2030-01-01T00:00:00'}),
        {...balance, 'free_starts_remaining': -1},
        {...balance, 'free_starts_remaining': 4},
        {...balance, 'reward': 0},
        {...balance, 'reward': 1000000001},
        {...balance, 'pass_day_price': 0},
        {...balance, 'pass_day_price': -30},
        {...balance, 'pass_day_price': 1000000001},
        {...balance, 'resets_at': 'tomorrow'},
        {...balance, 'resets_at': '2030-01-01T00:00:00'},
      ];
      t.calls.clear();
      for (final response in invalid) {
        t.handle = (_, _, _) async => OnlineResponse(200, response);
        await c.refreshEconomy();
        expect(c.economy, same(confirmed), reason: '$response');
        expect(c.economyError, 'ECONOMY_UNAVAILABLE', reason: '$response');
        expect(c.economyLoading, isFalse);
        await expectLater(c.buyPass(1), throwsStateError);
      }
      t.handle = (_, _, _) async => OnlineResponse(
        200,
        benefit({'no_ads': true, 'ad_free_until': '2030-01-01T00:00:00Z'}),
      );
      await c.refreshEconomy();
      expect(c.economy!['benefits']['no_ads'], isTrue);
      expect(c.economyError, isNull);
      expect(t.calls.every((call) => call['method'] == 'GET'), isTrue);
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'confirmed spend remains known when follow-up account is malformed',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      t.calls.clear();
      t.handle = (method, _, body) async => OnlineResponse(
        200,
        method == 'POST' ? receipt(body!) : {...balance, 'pass_day_price': 0},
      );
      await c.buyPass(1);
      expect(c.passReceipt!['cost'], 30);
      expect(c.pendingPass, isNull);
      expect(store.values, isEmpty);
      expect(c.economyError, 'ECONOMY_UNAVAILABLE');
      await expectLater(c.buyPass(1), throwsStateError);
      await c.recoverPass();
      expect(t.calls.where((call) => call['method'] == 'POST'), hasLength(1));
      t.handle = (_, _, _) async => const OnlineResponse(200, balance);
      await c.refreshEconomy();
      expect(c.economyError, isNull);
      expect(c.passReceipt!['cost'], 30);
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'unresolved pass recovery remains available after malformed account refresh',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await c.buyPass(1);
      final original = c.pendingPass!;
      final malformed = Map<String, dynamic>.from(balance)..['reward'] = -10;
      t.handle = (_, _, _) async => OnlineResponse(200, malformed);
      await c.refreshEconomy();
      expect(c.pendingPass, original);
      expect(c.economyError, 'ECONOMY_UNAVAILABLE');
      t.calls.clear();
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path.contains('/passes/') ? receipt(original) : balance,
      );
      await c.recoverPass();
      expect(c.passReceipt!['operation_id'], original['operation_id']);
      expect(c.pendingPass, isNull);
      expect(c.economyError, isNull);
      expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'account refresh blocks new spending until its current response arrives',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      final response = Completer<OnlineResponse>();
      t.calls.clear();
      t.handle = (_, _, _) => response.future;
      final refresh = c.refreshEconomy();
      expect(c.economyLoading, isTrue);
      await expectLater(c.buyPass(1), throwsStateError);
      expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
      response.complete(const OnlineResponse(200, balance));
      await refresh;
      expect(c.economyLoading, isFalse);
      expect(c.economyError, isNull);
      t.handle = (method, _, body) async =>
          OnlineResponse(200, method == 'POST' ? receipt(body!) : balance);
      await c.buyPass(1);
      expect(t.calls.where((call) => call['method'] == 'POST'), hasLength(1));
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'partial first account response is unavailable rather than ads-active',
    () async {
      final t = FakeTransport();
      t.handle = (_, _, _) async => const OnlineResponse(200, {
        'account': {'id': 'a'},
        'csrf_token': 'csrf',
      });
      final c = OnlineClient(t);
      await c.restoreSession();
      final partialBenefits = {...balance['benefits'] as Map}..remove('no_ads');
      t.handle = (_, _, _) async =>
          OnlineResponse(200, {...balance, 'benefits': partialBenefits});
      await c.refreshEconomy();
      expect(c.economy, isNull);
      expect(c.economyError, 'ECONOMY_UNAVAILABLE');
      t.calls.clear();
      await expectLater(c.buyPass(1), throwsStateError);
      expect(t.calls, isEmpty);
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'account accepts documented numeric boundaries and older optional expiry',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      for (final optionalExpiry in [
        <String, dynamic>{},
        {'ad_free_until': '1970-01-01T00:00:00Z'},
      ]) {
        t.handle = (_, _, _) async => OnlineResponse(200, {
          ...balance,
          'reward': 1000000000,
          'pass_day_price': 1000000000,
          'free_starts_remaining': 0,
          'benefits': {
            ...balance['benefits'] as Map,
            'dirt': 1000000000000000,
            ...optionalExpiry,
          },
        });
        await c.refreshEconomy();
        expect(c.economyError, isNull);
        expect(c.economy!['benefits']['dirt'], 1000000000000000);
      }
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'streamed finalized and void games refresh next-game admission',
    () async {
      for (final phase in ['finalized', 'void']) {
        final t = FakeTransport();
        final c = OnlineClient(t);
        await c.watchMatch('m');
        final ended = view(2);
        ended['projection']['board']['phase'] = 'settlement';
        ended['projection']['online'] = {
          'financial_phase': phase,
          'completed_games': 1,
          'games_per_match': 3,
        };
        t.handle = (_, _, _) async =>
            OnlineResponse(200, {...ended, 'admission_waiting': true});
        t.frames.add({'cursor': 2, 'data': ended});
        await Future<void>.delayed(Duration.zero);
        expect(c.snapshot!.admissionWaiting, isTrue, reason: phase);
        expect(c.snapshot!.cursor, 2);
        expect(t.calls.length, 2, reason: 'one advisory refresh per boundary');
        t.frames.add({'cursor': 2, 'data': ended});
        await Future<void>.delayed(Duration.zero);
        expect(t.calls.length, 2, reason: 'duplicates do not poll');
        c.dispose();
        await t.frames.close();
      }
    },
  );

  test(
    'snapshot refresh updates allowance wait without replacing game or cursor',
    () async {
      final t = FakeTransport();
      final c = OnlineClient(t);
      t.handle = (_, _, _) async =>
          OnlineResponse(200, {...view(1), 'admission_waiting': true});
      await c.watchMatch('m');
      expect(c.snapshot!.admissionWaiting, isTrue);
      t.handle = (_, _, _) async => OnlineResponse(200, view(1));
      await c.refresh();
      expect(c.snapshot!.admissionWaiting, isFalse);
      expect(c.snapshot!.version, 1);
      expect(c.snapshot!.gameId, 'game-1');
      c.dispose();
      await t.frames.close();
    },
  );

  test(
    'pass spends only after explicit request; validates days and sends CSRF',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      expect(c.economy!['benefits']['dirt'], 180);
      t.calls.clear();
      for (final days in [0, 2, 3, 4, 5, 6, 8, 30, 365]) {
        await expectLater(c.buyPass(days), throwsArgumentError);
      }
      expect(t.calls, isEmpty);
      t.handle = (method, _, body) async =>
          OnlineResponse(200, method == 'POST' ? receipt(body!) : balance);
      await c.buyPass(7);
      final post = t.calls.singleWhere((r) => r['method'] == 'POST');
      expect(post['body']['days'], 7);
      expect(post['headers']['X-CSRF-Token'], 'csrf');
      expect(c.passReceipt!['cost'], 210);
      expect(c.pendingPass, isNull);
      expect(store.values, isEmpty);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'lost pass reply survives reload and receipt lookup never spends twice',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      t.handle = (_, _, _) async => throw TimeoutException('lost');
      await c.buyPass(7);
      final original = c.pendingPass!;
      expect(store.values.values.single, jsonEncode(original));
      await expectLater(c.buyPass(1), throwsStateError);
      await expectLater(c.logout(), throwsStateError);
      c.dispose();
      final t2 = FakeTransport();
      final restored = await signedIn(t2, store);
      expect(restored.pendingPass, original);
      t2.calls.clear();
      t2.handle = (_, path, _) async => OnlineResponse(
        200,
        path.contains('/passes/') ? receipt(original) : balance,
      );
      await restored.recoverPass();
      expect(t2.calls.where((r) => r['method'] == 'POST'), isEmpty);
      expect(restored.pendingPass, isNull);
      expect(store.values, isEmpty);
      restored.dispose();
      await t.frames.close();
      await t2.frames.close();
    },
  );
  test(
    'absent receipt retries exact pass; authentication failure retains uncertainty',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      t.handle = (_, _, _) async =>
          const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
      await c.buyPass(1);
      final original = c.pendingPass;
      t.calls.clear();
      t.handle = (method, _, _) async => method == 'GET'
          ? const OnlineResponse(404, {'code': 'OPERATION_NOT_FOUND'})
          : const OnlineResponse(401, {'code': 'UNAUTHENTICATED'});
      await c.recoverPass();
      expect(t.calls.first['method'], 'GET');
      expect(
        t.calls.singleWhere((r) => r['method'] == 'POST')['body'],
        original,
      );
      expect(c.pendingPass, original);
      expect(c.economyError, 'OUTCOME_UNKNOWN');
      c.dispose();
      await t.frames.close();
    },
  );
  test('foreign or malformed receipt cannot clear unresolved spend', () async {
    final t = FakeTransport(), store = MemoryOperations();
    final c = await signedIn(t, store);
    t.handle = (_, _, body) async => OnlineResponse(200, {
      ...receipt(body!),
      'operation_id': 'another-operation',
    });
    await c.buyPass(1);
    expect(c.pendingPass, isNotNull);
    expect(c.economyError, 'OUTCOME_UNKNOWN');
    c.dispose();
    await t.frames.close();
  });
  test(
    'definitive insufficient balance rejection leaves no pending purchase',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      t.handle = (_, _, _) async =>
          const OnlineResponse(409, {'code': 'INSUFFICIENT_DIRT'});
      await c.buyPass(7);
      expect(c.pendingPass, isNull);
      expect(c.economyError, 'INSUFFICIENT_DIRT');
      expect(store.values, isEmpty);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'account switch isolates pending pass and ignores stale balance response',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final c = await signedIn(t, store);
      final delayed = Completer<OnlineResponse>();
      t.handle = (_, path, _) async => path == '/v1/economy'
          ? delayed.future
          : const OnlineResponse(200, {
              'account': {'id': 'b'},
              'csrf_token': 'test-csrf',
            });
      final loading = c.refreshEconomy();
      await c.restoreSession();
      delayed.complete(const OnlineResponse(200, balance));
      await loading;
      expect(c.account!['id'], 'b');
      expect(c.economy, isNull);
      expect(c.pendingPass, isNull);
      c.dispose();
      await t.frames.close();
    },
  );
  test(
    'legacy durations restore their original receipt without a new spend',
    () async {
      for (final days in [2, 3, 4, 5, 6]) {
        final t = FakeTransport(), store = MemoryOperations();
        final original = {'operation_id': 'historic', 'days': days};
        store.values['a/economy-pass'] = jsonEncode(original);
        final c = await signedIn(t, store);
        t.calls.clear();
        t.handle = (_, path, _) async => OnlineResponse(
          200,
          path.contains('/passes/') ? receipt(original) : balance,
        );
        await c.recoverPass();
        expect(c.passReceipt!['days'], days);
        expect(c.pendingPass, isNull);
        expect(t.calls.where((r) => r['method'] == 'POST'), isEmpty);
        c.dispose();
        await t.frames.close();
      }
    },
  );
  test(
    'retired pending product clears only after locked server rejection',
    () async {
      final t = FakeTransport(), store = MemoryOperations();
      final original = {'operation_id': 'historic', 'days': 6};
      store.values['a/economy-pass'] = jsonEncode(original);
      final c = await signedIn(t, store);
      t.calls.clear();
      t.handle = (method, _, _) async => method == 'GET'
          ? const OnlineResponse(404, {'code': 'OPERATION_NOT_FOUND'})
          : const OnlineResponse(409, {'code': 'PRODUCT_UNAVAILABLE'});
      await c.recoverPass();
      expect(
        t.calls.singleWhere((r) => r['method'] == 'POST')['body'],
        original,
      );
      expect(c.pendingPass, isNull);
      expect(c.economyError, 'PRODUCT_UNAVAILABLE');
      expect(store.values, isEmpty);
      c.dispose();
      await t.frames.close();
    },
  );
  test('corrupt stored pass fails closed without a request', () async {
    final t = FakeTransport(), store = MemoryOperations();
    store.values['a/economy-pass'] = '{"operation_id":"old","days":8}';
    final c = await signedIn(t, store);
    t.calls.clear();
    await expectLater(c.buyPass(1), throwsStateError);
    expect(t.calls, isEmpty);
    expect(store.values, isNotEmpty);
    c.dispose();
    await t.frames.close();
  });
}
