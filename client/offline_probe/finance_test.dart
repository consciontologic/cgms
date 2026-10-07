import 'dart:convert';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';
import 'finance.dart';
import 'probe.dart';

void check(bool condition, String message) {
  if (!condition) throw StateError(message);
}

String normalized(dynamic value) {
  if (value is Map) {
    final keys = value.keys.cast<String>().toList()..sort();
    return '{${keys.map((k) => '${jsonEncode(k)}:${normalized(value[k])}').join(',')}}';
  }
  if (value is List) return '[${value.map(normalized).join(',')}]';
  return jsonEncode(value);
}

Future<void> main(List<String> args) async {
  check(
    Exact.parse({
          'numerator': '9007199254740993',
          'denominator': '2',
        }).add(Exact.one).toJson()['numerator'] ==
        '9007199254740995',
    'large exact arithmetic',
  );
  check(
    Exact.parse({'numerator': '1', 'denominator': '3'})
            .add(Exact.parse({'numerator': '2', 'denominator': '3'}))
            .compareTo(Exact.one) ==
        0,
    'thirds exact',
  );
  for (final bad in [
    {'numerator': '2', 'denominator': '4'},
    {'numerator': '-0', 'denominator': '1'},
    {'numerator': '1', 'denominator': '0'},
    {'numerator': 1, 'denominator': '2'},
  ]) {
    var rejected = false;
    try {
      Exact.parse(bad);
    } on FormatException {
      rejected = true;
    }
    check(rejected, 'noncanonical rational accepted');
  }
  if (args.length != 2)
    throw ArgumentError('library and fixture directory required');
  final probe = NativeProbe(args[0]);
  final reports = <Object>[];
  for (final population in [
    '3',
    '4',
    'cycle',
    'large',
    'same-game',
    'nullification',
  ]) {
    final raw = File(
      '${args[1]}/finance-$population.request.json',
    ).readAsBytesSync();
    final request = jsonDecode(utf8.decode(raw)) as Map<String, dynamic>;
    final expected = jsonDecode(
      File('${args[1]}/finance-$population.expected.json').readAsStringSync(),
    );
    final untouched = normalized(request);
    final native = jsonDecode(
      utf8.decode(probe.execute(Uint8List.fromList(raw))),
    );
    check(
      normalized(native) == normalized(expected),
      'native financial oracle differs',
    );
    final actual = runFinance(request);
    check(
      normalized(actual) == normalized(expected),
      'independent Dart finance diverged for $population',
    );
    check(normalized(request) == untouched, 'Dart candidate mutated input');
    final frames = actual['frames'] as List;
    final commands = request['commands'] as List;
    for (var i = 0; i < frames.length; i++) {
      final resume = jsonDecode(jsonEncode(request)) as Map<String, dynamic>;
      resume['ledger'] = jsonDecode(jsonEncode(frames[i]['ledger']));
      resume['commands'] = commands.sublist(i);
      final replay = runFinance(resume);
      check(
        normalized((replay['frames'] as List).last) == normalized(frames.last),
        'restart $i diverged',
      );
    }
    if (population == 'cycle') {
      // Probe admission refusal is deliberately different from Go's cycle batch;
      // it must never be mistaken for forgiveness or partial commitment.
      final bounded = copyMap(request);
      bounded['commands'].last['receipts'][0]['amount'] = {
        'numerator': '1',
        'denominator': '4096',
      };
      final beforeBudget = normalized(bounded);
      final refusal = runFinance(bounded);
      check(
        refusal['error'] == 'command_rejected' &&
            refusal['failed_step'] == 2 &&
            !refusal.containsKey('frames'),
        'budget returned partial financial result',
      );
      check(normalized(bounded) == beforeBudget, 'budget mutated input debt');
    }
    final malformed = jsonDecode(jsonEncode(request)) as Map<String, dynamic>;
    malformed['commands'][0]['game_id'] = 'wrong';
    check(
      runFinance(malformed)['error'] == 'command_rejected',
      'wrong game accepted',
    );
    final incomplete = jsonDecode(jsonEncode(request)) as Map<String, dynamic>;
    incomplete['commands'] = [
      {...(commands[0] as Map), 'kind': 'finalize'},
    ];
    check(
      runFinance(incomplete)['error'] == 'command_rejected',
      'early finalize accepted',
    );
    for (final mutate in <void Function(Map<String, dynamic>)>[
      (r) => r['schema'] = 'unsupported',
      (r) => r['unknown'] = true,
      (r) => r['ledger']['cash'][0] = {'numerator': '-1', 'denominator': '1'},
      (r) => r['commands'] = [
        {
          ...(commands.first as Map),
          'kind': 'receipt',
          'receipts': [
            {
              'seat': 0,
              'amount': {'numerator': '-1', 'denominator': '1'},
            },
          ],
        },
      ],
      (r) => r['commands'] = [
        {...(commands.first as Map), 'kind': 'unsupported'},
      ],
    ]) {
      final bad = copyMap(request);
      mutate(bad);
      final rejected = runFinance(bad);
      check(
        rejected.containsKey('error') && !rejected.containsKey('frames'),
        'invalid request returned a partial trace',
      );
      final nativeRejected = jsonDecode(
        utf8.decode(
          probe.execute(Uint8List.fromList(utf8.encode(jsonEncode(bad)))),
        ),
      );
      if (bad['schema'] != financeSchema) {
        // The shared ABI routes unknown versions to its original board envelope.
        check(
          nativeRejected['schema'] == 'cgms-offline-probe-v1' &&
              nativeRejected['error'] == 'invalid_request' &&
              nativeRejected['failed_step'] == -1 &&
              !nativeRejected.containsKey('frames'),
          'unknown ABI schema accepted',
        );
      } else {
        check(
          normalized(rejected) == normalized(nativeRejected),
          'negative fixture differs',
        );
      }
    }
    final isolates = await Future.wait(
      List.generate(2, (_) => Isolate.run(() => runFinance(copyMap(request)))),
    );
    check(
      isolates.every((result) => normalized(result) == normalized(expected)),
      'independent candidate isolate parity',
    );
    final durations = <int>[];
    final nativeDurations = <int>[];
    for (var i = 0; i < 5; i++) {
      runFinance(request);
      probe.execute(Uint8List.fromList(raw));
    }
    for (var i = 0; i < 30; i++) {
      final timer = Stopwatch()..start();
      final result = jsonEncode(
        runFinance(jsonDecode(utf8.decode(raw)) as Map<String, dynamic>),
      );
      timer.stop();
      check(
        normalized(jsonDecode(result)) == normalized(expected),
        'timed parity',
      );
      final nativeTimer = Stopwatch()..start();
      final nativeResult = probe.execute(Uint8List.fromList(raw));
      nativeTimer.stop();
      check(
        normalized(jsonDecode(utf8.decode(nativeResult))) ==
            normalized(expected),
        'timed native parity',
      );
      nativeDurations.add(nativeTimer.elapsedMicroseconds);
      durations.add(timer.elapsedMicroseconds);
    }
    durations.sort();
    nativeDurations.sort();
    reports.add({
      'fixture': population,
      'frames': frames.length,
      'samples': 30,
      'dart_p50_us': durations[14],
      'dart_p95_us': durations[28],
      'native_p50_us': nativeDurations[14],
      'native_p95_us': nativeDurations[28],
    });
  }
  stdout.writeln(
    jsonEncode({
      'scope':
          'independent financial candidate; no board gameplay or mobile acceptance',
      'parity': 'passed',
      'measurements': reports,
    }),
  );
}
