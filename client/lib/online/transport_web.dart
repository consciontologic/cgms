// Browser-owned cookie authentication; credentials never enter URLs or storage.
// ignore_for_file: deprecated_member_use, avoid_web_libraries_in_flutter
import 'dart:async';
import 'dart:convert';
import 'dart:html' as html;
import 'online_client.dart';

OnlineTransport createOnlineTransport(Uri baseUri) => BrowserTransport(baseUri);

class BrowserTransport implements OnlineTransport, OnlineOperationStore {
  BrowserTransport(this.baseUri) {
    if (!['https', 'http'].contains(baseUri.scheme) ||
        baseUri.userInfo.isNotEmpty ||
        baseUri.query.isNotEmpty ||
        baseUri.fragment.isNotEmpty)
      throw ArgumentError('Invalid API origin');
  }
  String _key(String account, String scope) =>
      'cgms.pending.v1:${Uri.encodeComponent(baseUri.origin)}:${Uri.encodeComponent(account)}:${Uri.encodeComponent(scope)}';
  @override
  String? read(String account, String scope) =>
      html.window.sessionStorage[_key(account, scope)];
  @override
  void write(String account, String scope, String value) {
    html.window.sessionStorage[_key(account, scope)] = value;
  }

  @override
  void remove(String account, String scope) {
    html.window.sessionStorage.remove(_key(account, scope));
  }

  final Uri baseUri;
  final Set<html.WebSocket> _sockets = {};
  final Set<html.HttpRequest> _requests = {};
  @override
  Future<OnlineResponse> request(
    String method,
    String path, {
    Map<String, dynamic>? body,
    Map<String, String> headers = const {},
  }) async {
    final request = html.HttpRequest();
    final done = Completer<OnlineResponse>();
    request.open(method, baseUri.resolve(path).toString());
    request.withCredentials = true;
    request.timeout = 2000;
    headers.forEach(request.setRequestHeader);
    _requests.add(request);
    void fail() {
      if (!done.isCompleted) done.completeError(StateError('UNAVAILABLE'));
    }

    request.onLoad.listen((_) {
      if (done.isCompleted) return;
      try {
        done.complete(
          OnlineResponse(
            request.status ?? 0,
            jsonDecode(request.responseText ?? '{}'),
          ),
        );
      } catch (_) {
        final status = request.status ?? 0;
        // Reverse proxies often return HTML while the authority is starting.
        // Keep their retryable status; malformed successful JSON is terminal.
        if (status >= 400 && status <= 599) {
          done.complete(
            OnlineResponse(status, {
              'code': status == 429 ? 'OVERLOADED' : 'UNAVAILABLE',
            }),
          );
        } else {
          done.completeError(const FormatException('INVALID_RESPONSE'));
        }
      }
    });
    request.onError.listen((_) => fail());
    request.onTimeout.listen((_) => fail());
    request.onAbort.listen((_) => fail());
    request.send(body == null ? null : jsonEncode(body));
    try {
      return await done.future;
    } finally {
      _requests.remove(request);
    }
  }

  @override
  Stream<Map<String, dynamic>> events(String path) {
    final uri = baseUri
        .resolve(path)
        .replace(scheme: baseUri.scheme == 'https' ? 'wss' : 'ws');
    final socket = html.WebSocket(uri.toString());
    _sockets.add(socket);
    late StreamController<Map<String, dynamic>> controller;
    controller = StreamController(
      onCancel: () {
        socket.close();
        _sockets.remove(socket);
      },
    );
    socket.onOpen.listen((_) {
      if (!controller.isClosed) controller.add({'type': 'transport-open'});
    });
    socket.onMessage.listen((event) {
      if (controller.isClosed) return;
      try {
        final frame = Map<String, dynamic>.from(
          jsonDecode(event.data as String) as Map,
        );
        if (frame['type'] == 'transport-open')
          throw const FormatException('INVALID_EVENT');
        controller.add(frame);
      } catch (_) {
        controller.addError(const FormatException('INVALID_EVENT'));
      }
    });
    socket.onError.listen((_) {
      if (!controller.isClosed)
        controller.addError(StateError('STREAM_UNAVAILABLE'));
    });
    socket.onClose.listen((_) {
      _sockets.remove(socket);
      controller.close();
    });
    return controller.stream;
  }

  @override
  void close() {
    for (final socket in _sockets.toList()) {
      socket.close();
    }
    for (final request in _requests.toList()) {
      request.abort();
    }
  }
}
