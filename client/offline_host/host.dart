// VM/desktop host harness only. dart:io intentionally prevents a web fallback.
import 'dart:async';
import 'dart:convert';
import 'dart:io';

class LocalRequestException implements Exception {
  const LocalRequestException(this.code);
  final String code;
  @override
  String toString() => 'Local request failed: $code';
}

class OfflineHost {
  OfflineHost._(this._process)
    : _lines = StreamIterator(
        _process.stdout.transform(utf8.decoder).transform(const LineSplitter()),
      ) {
    _stderr = _process.stderr.drain<void>();
  }
  final Process _process;
  final StreamIterator<String> _lines;
  late final Future<void> _stderr;
  Future<void> _tail = Future.value();
  bool _closed = false;
  static Future<OfflineHost> start(
    String executable,
    String directory,
    String rulesHash,
  ) async {
    return OfflineHost._(
      await Process.start(executable, [
        '--save-dir',
        directory,
        '--rules-hash',
        rulesHash,
      ], runInShell: false),
    );
  }

  Future<Map<String, dynamic>> request(Map<String, dynamic> request) {
    late final String raw;
    try {
      raw = jsonEncode(request);
    } catch (_) {
      return Future.error(const LocalRequestException('invalid_request'));
    }
    if (utf8.encode(raw).length > 64 * 1024)
      return Future.error(const LocalRequestException('request_too_large'));
    final completed = Completer<Map<String, dynamic>>();
    _tail = _tail.then((_) async {
      try {
        if (_closed) throw const LocalRequestException('closed');
        _process.stdin.writeln(raw);
        await _process.stdin.flush();
        if (!await _lines.moveNext().timeout(const Duration(seconds: 30)))
          throw const LocalRequestException('host_stopped');
        final line = _lines.current;
        if (line.length > 4 * 1024 * 1024)
          throw const LocalRequestException('invalid_response');
        final response = jsonDecode(line);
        if (response is! Map)
          throw const LocalRequestException('invalid_response');
        if (response['error'] is String)
          throw LocalRequestException(response['error']);
        if (response['schema'] != 'cgms-offline-view-v1' ||
            response['view'] is! Map)
          throw const LocalRequestException('invalid_response');
        completed.complete(Map<String, dynamic>.from(response['view']));
      } on LocalRequestException catch (e) {
        completed.completeError(e);
      } catch (_) {
        _closed = true;
        _process.kill();
        completed.completeError(
          const LocalRequestException('outcome_unknown_reopen_and_load'),
        );
      }
    });
    return completed.future;
  }

  Future<void> close() async {
    await _tail;
    _closed = true;
    await _process.stdin.close();
    await _process.exitCode.timeout(
      const Duration(seconds: 2),
      onTimeout: () {
        _process.kill();
        return -1;
      },
    );
    await _lines.cancel();
    await _stderr;
  }
}
