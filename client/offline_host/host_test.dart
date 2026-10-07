import 'dart:io';
import 'dart:convert';
import 'host.dart';

void check(bool condition, String message) {
  if (!condition) throw StateError(message);
}

Future<void> main(List<String> args) async {
  final binary = args[0], directory = args[1], rules = args[2];
  await Directory(directory).create(recursive: true);
  var host = await OfflineHost.start(binary, directory, rules);
  var view = await host.request({
    'kind': 'tutorial',
    'slot': 'lesson',
    'population': 3,
    'games': 1,
    'difficulty': 'beginner',
  });
  for (var i = 0; i < 30 && view['tutorial_step'] != 3; i++) {
    final step = view['tutorial_step'];
    final request = <String, dynamic>{
      'slot': 'lesson',
      'version': view['version'],
    };
    if (step == 0 || step == 2) {
      final actions = view['actions'] as List;
      final action = actions.firstWhere(
        (a) => step == 0
            ? (a['command']?['kind'] == 'open-series')
            : (a['operation']?['kind'] == 'end-turn'),
      );
      request.addAll({'kind': 'act', 'action': action});
    } else {
      request['kind'] = 'step';
    }
    view = await host.request(request);
    await host.close();
    host = await OfflineHost.start(binary, directory, rules);
    final restored = await host.request({'kind': 'load', 'slot': 'lesson'});
    check(
      restored['version'] == view['version'] &&
          restored['tutorial_step'] == view['tutorial_step'],
      'restart lost state',
    );
    view = restored;
  }
  check(view['tutorial_step'] == 3, 'tutorial incomplete');
  var rejected = false;
  try {
    await host.request({'kind': 'step', 'slot': 'lesson', 'version': -1});
  } on LocalRequestException {
    rejected = true;
  }
  check(rejected, 'stale update accepted');
  final loaded = await host.request({'kind': 'load', 'slot': 'lesson'});
  check(loaded['version'] == view['version'], 'failed update changed save');
  final queued = {'kind': 'load', 'slot': 'lesson'};
  final frozen = host.request(queued);
  queued['slot'] = 'missing';
  check(
    (await frozen)['version'] == view['version'],
    'queued request was retargeted',
  );
  await host.close();
  final crashed = await Process.start(binary, [
    '--save-dir',
    directory,
    '--rules-hash',
    rules,
  ]);
  crashed.stderr.drain<void>();
  crashed.stdin.writeln(jsonEncode({'kind': 'load', 'slot': 'lesson'}));
  final reply = await crashed.stdout
      .transform(utf8.decoder)
      .transform(const LineSplitter())
      .first;
  check(
    jsonDecode(reply)['view']['version'] == view['version'],
    'pre-crash load',
  );
  crashed.kill(ProcessSignal.sigkill);
  await crashed.exitCode;
  final recovered = await OfflineHost.start(binary, directory, rules);
  final afterCrash = await recovered.request({
    'kind': 'load',
    'slot': 'lesson',
  });
  check(
    afterCrash['version'] == view['version'] &&
        afterCrash['tutorial_step'] == 3,
    'process death lost durable save',
  );
  await recovered.close();
  stdout.writeln(
    'PASS: native host process restarts, durable tutorial, stale rejection',
  );
}
