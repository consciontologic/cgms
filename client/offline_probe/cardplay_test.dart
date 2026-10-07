import 'dart:convert';
import 'dart:io';
import 'dart:isolate';
import 'dart:typed_data';
import 'cardplay.dart';
import 'probe.dart';

void check(bool value, String message) {
  if (!value) throw StateError(message);
}

String firstDifference(dynamic a, dynamic b, [String path = '']) {
  if (a is Map && b is Map) {
    for (final key in {...a.keys, ...b.keys}) {
      if (!a.containsKey(key) || !b.containsKey(key))
        return '$path.$key missing';
      if (canonical(a[key]) != canonical(b[key]))
        return firstDifference(a[key], b[key], '$path.$key');
    }
  }
  if (a is List && b is List) {
    if (a.length != b.length) return '$path length ${a.length}/${b.length}';
    for (var i = 0; i < a.length; i++) {
      if (canonical(a[i]) != canonical(b[i]))
        return firstDifference(a[i], b[i], '$path[$i]');
    }
  }
  return '$path differs';
}

Future<void> main(List<String> args) async {
  check(
    sha256('abc') ==
        'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad',
    'SHA256 vector',
  );
  check(
    sha256('') ==
        'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    'empty SHA256 vector',
  );
  final mode = args.length > 2 ? args[2] : 'verify';
  final native = mode == 'dart' ? null : NativeProbe(args[1]);
  Map nativeRun(Map request) =>
      jsonDecode(
            utf8.decode(
              native!.execute(
                Uint8List.fromList(utf8.encode(jsonEncode(request))),
              ),
            ),
          )
          as Map;
  final reports = <Map>[];
  for (final n in [3, 4]) {
    final raw = File('${args[0]}/cardplay-$n.request.json').readAsStringSync();
    final request = jsonDecode(raw) as Map<String, dynamic>;
    final expected = jsonDecode(
      File('${args[0]}/cardplay-$n.expected.json').readAsStringSync(),
    );
    final original = canonical(request);
    if (mode == 'verify') {
      final references = File(
        '${args[0]}/cardplay-$n.frames.jsonl',
      ).readAsLinesSync();
      String? mismatch;
      final actual = runCardplay(
        request,
        onFrame: (index, frame) {
          if (mismatch == null && canonical(frame) != references[index]) {
            final wanted = jsonDecode(references[index]);
            mismatch = 'frame $index ${firstDifference(frame, wanted)}';
            File(
              '${args[0]}/cardplay-$n.first-divergence.json',
            ).writeAsStringSync(
              jsonEncode({'index': index, 'actual': frame, 'expected': wanted}),
            );
          }
        },
      );
      check(mismatch == null, mismatch ?? '');
      check(
        canonical(actual) == canonical(expected),
        'independent complete $n-player trace differs: ${actual['error']}',
      );
      check(
        canonical(nativeRun(request)) == canonical(expected),
        'native complete trace differs',
      );
      check(canonical(request) == original, 'input alias');
      for (var i = 0; i < references.length - 1; i++) {
        final checkpoint = jsonDecode(references[i]);
        final resumed = {
          'schema': cardplaySchema,
          'match': checkpoint['match'],
          'steps': [request['steps'][i]],
        };
        final dartResult = runCardplay(resumed);
        final nativeResult = nativeRun(resumed);
        check(
          dartResult['digests']?[1] == expected['digests'][i + 1],
          'Dart checkpoint $i: ${dartResult['error']}',
        );
        check(
          canonical(nativeResult) == canonical(dartResult),
          'native restart $i differs',
        );
      }
      for (final kind in [
        'unsupported',
        'wrong-game',
        'confined',
        'partition',
        'chance',
      ]) {
        final bad = copy(request);
        bad['steps'] = [bad['steps'].first];
        if (kind == 'unsupported') bad['steps'][0]['kind'] = 'purchase';
        if (kind == 'wrong-game') bad['steps'][0]['game_id'] = 'another';
        if (kind == 'confined')
          bad['match']['game']['board']['players'][0]['confined'] = true;
        if (kind == 'partition')
          bad['match']['game']['board']['cards'].removeLast();
        if (kind == 'chance')
          bad['steps'][0]['order'][0] = bad['steps'][0]['order'][1];
        final rejected = runCardplay(bad);
        check(
          rejected.containsKey('error') && !rejected.containsKey('final'),
          'negative accepted $kind',
        );
        check(
          canonical(rejected) == canonical(nativeRun(bad)),
          'negative $kind disagrees',
        );
      }
      final isolated = await Isolate.run(() => runCardplay(copy(request)));
      check(
        canonical(isolated) == canonical(expected),
        'independent isolate mismatch',
      );
      reports.add({
        'population': n,
        'frames': references.length,
        'games': 3,
        'restarts': references.length - 1,
        'parity': 'passed',
      });
    } else {
      // Fresh processes separate Dart-only RSS from the Go runtime loaded by FFI.
      Map run() => mode == 'dart'
          ? runCardplay(jsonDecode(raw) as Map<String, dynamic>)
          : nativeRun(jsonDecode(raw));
      run();
      final before = ProcessInfo.currentRss;
      final durations = <int>[];
      for (var sample = 0; sample < 3; sample++) {
        final timer = Stopwatch()..start();
        final result = jsonEncode(run());
        timer.stop();
        check(
          canonical(jsonDecode(result)) == canonical(expected),
          'timed parity',
        );
        durations.add(timer.elapsedMicroseconds);
      }
      durations.sort();
      final rssAfter = ProcessInfo.currentRss, peak = ProcessInfo.maxRss;
      // Pick actual authorized snapshots covering each policy branch, never state.
      final choices = <String, Map>{};
      final lines = File(
        '${args[0]}/cardplay-$n.frames.jsonl',
      ).openRead().transform(utf8.decoder).transform(const LineSplitter());
      await for (final line in lines) {
        final f = jsonDecode(line);
        for (var seat = 0; seat < n; seat++) {
          final d = f['decisions'][seat];
          if (d != null)
            choices.putIfAbsent(
              d['kind'],
              () => f['observations'][seat] as Map,
            );
        }
        if (choices.length == 3) break;
      }
      final decisions = <Map>[];
      for (final entry in choices.entries) {
        final input = jsonEncode({
          'schema': 'cgms-offline-cardplay-choice-v1',
          'observation': entry.value,
        });
        dynamic decision() {
          if (mode == 'native')
            return jsonDecode(
              utf8.decode(
                native!.execute(Uint8List.fromList(utf8.encode(input))),
              ),
            );
          final o = jsonDecode(input)['observation'];
          return choose(
            o['board'],
            o['opening_used'],
            o['turn_started'],
            o['round_closed'],
          );
        }

        final times = <int>[];
        for (var i = 0; i < 110; i++) {
          final watch = Stopwatch()..start();
          final result = jsonEncode(decision());
          watch.stop();
          check(jsonDecode(result)['kind'] == entry.key, 'decision branch');
          if (i >= 10) times.add(watch.elapsedMicroseconds);
        }
        times.sort();
        decisions.add({
          'kind': entry.key,
          'samples': 100,
          'p50_us': times[49],
          'p95_us': times[94],
        });
      }
      reports.add({
        'population': n,
        'whole_match_samples': 3,
        'whole_match_p50_us': durations[1],
        'whole_match_max_us': durations.last,
        'rss_before_bytes': before,
        'rss_after_bytes': rssAfter,
        'process_peak_rss_bytes': peak,
        'decision_samples': decisions,
      });
    }
  }
  stdout.writeln(
    jsonEncode({'mode': mode, 'parity': 'passed', 'reports': reports}),
  );
}
