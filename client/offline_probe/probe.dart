import 'dart:ffi';
import 'dart:typed_data';

/// P01 research adapter. Full-state buffers are local synthetic evidence only.
/// Call from a worker isolate; Go work must not block Flutter's UI isolate.
/// This adapter deliberately has no web implementation or network dependency.
final class NativeProbe {
  static const maxBytes = 4 << 20;
  final Pointer<Void> Function(int) _allocate;
  final void Function(Pointer<Void>) _free;
  final int Function(Pointer<Void>, int, Pointer<Void>, int) _call;

  factory NativeProbe(String libraryPath) {
    final library = DynamicLibrary.open(libraryPath);
    return NativeProbe._(
      library.lookupFunction<
        Pointer<Void> Function(Int32),
        Pointer<Void> Function(int)
      >('CGMSProbeAlloc'),
      library.lookupFunction<
        Void Function(Pointer<Void>),
        void Function(Pointer<Void>)
      >('CGMSProbeFree'),
      library.lookupFunction<
        Int32 Function(Pointer<Void>, Int32, Pointer<Void>, Int32),
        int Function(Pointer<Void>, int, Pointer<Void>, int)
      >('CGMSProbe'),
    );
  }
  NativeProbe._(this._allocate, this._free, this._call);

  Uint8List execute(Uint8List request) {
    if (request.isEmpty || request.length > maxBytes) {
      throw ArgumentError('Probe request must contain 1..$maxBytes bytes');
    }
    final input = _allocate(request.length);
    if (input == nullptr) throw StateError('Input allocation failed');
    try {
      final output = _allocate(maxBytes);
      if (output == nullptr) throw StateError('Output allocation failed');
      try {
        input.cast<Uint8>().asTypedList(request.length).setAll(0, request);
        final length = _call(input, request.length, output, maxBytes);
        if (length < 1 || length > maxBytes) {
          throw StateError('Native probe failed: $length');
        }
        // Detach the Dart result before releasing native memory.
        return Uint8List.fromList(output.cast<Uint8>().asTypedList(length));
      } finally {
        _free(output);
      }
    } finally {
      _free(input);
    }
  }
}
