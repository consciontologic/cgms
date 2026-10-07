import 'dart:convert';
import 'dart:ffi';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';

import 'probe.dart';

void check(bool condition, String message) {
  if (!condition) throw StateError(message);
}

Future<void> main(List<String> args) async {
  if (args.length != 2)
    throw ArgumentError('library path and fixture directory required');
  final probe = NativeProbe(args[0]);
  final library = DynamicLibrary.open(args[0]);
  final alloc = library
      .lookupFunction<
        Pointer<Void> Function(Int32),
        Pointer<Void> Function(int)
      >('CGMSProbeAlloc');
  final free = library
      .lookupFunction<
        Void Function(Pointer<Void>),
        void Function(Pointer<Void>)
      >('CGMSProbeFree');
  final call = library
      .lookupFunction<
        Int32 Function(Pointer<Void>, Int32, Pointer<Void>, Int32),
        int Function(Pointer<Void>, int, Pointer<Void>, int)
      >('CGMSProbe');
  check(
    alloc(0) == nullptr && alloc(NativeProbe.maxBytes + 1) == nullptr,
    'allocation limits',
  );
  check(call(nullptr, 1, nullptr, 1) == -1, 'null buffers');
  final tiny = alloc(2);
  try {
    tiny.cast<Uint8>().asTypedList(2).setAll(0, utf8.encode('{}'));
    check(call(tiny, 2, tiny, 0) == -1, 'zero output size');
    final out = alloc(1);
    try {
      check(call(tiny, 2, out, 1) == -2, 'short output buffer');
      check(call(tiny, -1, out, 1) == -1, 'negative input size');
      check(
        call(tiny, NativeProbe.maxBytes + 1, out, 1) == -1,
        'oversize input',
      );
    } finally {
      free(out);
    }
  } finally {
    free(tiny);
  }
  free(nullptr);
  for (final invalid in [
    '{}',
    '{"schema":"bad"}',
    '{"schema":"x","schema":"y"}',
    '{"unknown":1}',
  ]) {
    final response =
        jsonDecode(
              utf8.decode(
                probe.execute(Uint8List.fromList(utf8.encode(invalid))),
              ),
            )
            as Map;
    check(
      response['error'] == 'invalid_request' && !response.containsKey('frames'),
      'malformed request',
    );
  }
  var rejected = false;
  try {
    probe.execute(Uint8List(0));
  } on ArgumentError {
    rejected = true;
  }
  check(rejected, 'empty input was accepted');
  final results = <Map<String, Object>>[];
  for (final population in [3, 4]) {
    final input = File('${args[1]}/$population.request.json').readAsBytesSync();
    final expected = File(
      '${args[1]}/$population.expected.json',
    ).readAsStringSync();
    for (var i = 0; i < 5; i++) {
      check(utf8.decode(probe.execute(input)) == expected, 'warmup parity');
    }
    final rssBefore = ProcessInfo.currentRss;
    final timings = <int>[];
    for (var i = 0; i < 30; i++) {
      final timer = Stopwatch()..start();
      final actual = probe.execute(input);
      timer.stop();
      check(
        utf8.decode(actual) == expected,
        'state/event/observation/decision parity',
      );
      timings.add(timer.elapsedMicroseconds);
    }
    final rssAfter = ProcessInfo.currentRss;
    // The same bytes are verified after Dart JSON decode/encode and in fresh
    // isolates with independent allocations. No state or pointers cross isolates.
    final roundtrip = Uint8List.fromList(
      utf8.encode(jsonEncode(jsonDecode(utf8.decode(input)))),
    );
    check(
      utf8.decode(probe.execute(roundtrip)) == expected,
      'Dart JSON roundtrip',
    );
    final workers = await Future.wait(
      List.generate(
        2,
        (_) => Isolate.run(() {
          final worker = NativeProbe(args[0]);
          return utf8.decode(worker.execute(input));
        }),
      ),
    );
    check(workers.every((x) => x == expected), 'concurrent isolate parity');
    timings.sort();
    results.add({
      'population': population,
      'samples': timings.length,
      'p50_us': timings[(timings.length * .50).ceil() - 1],
      'p95_us': timings[(timings.length * .95).ceil() - 1],
      'max_us': timings.last,
      'rss_before_bytes': rssBefore,
      'rss_after_bytes': rssAfter,
      'response_bytes': utf8.encode(expected).length,
    });
  }
  stdout.writeln(
    jsonEncode({
      'schema': 'cgms-offline-probe-measurement-v1',
      'platform': Platform.operatingSystem,
      'dart': Platform.version,
      'scope':
          'synthetic board traces including narrow heuristic; not per-action mobile latency',
      'parity': 'passed',
      'measurements': results,
    }),
  );
}
