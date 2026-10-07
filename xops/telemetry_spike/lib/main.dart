import 'package:flutter/material.dart';
import 'package:opentelemetry/sdk.dart';
import 'package:opentelemetry/api.dart' as api;

void main() => runApp(const MaterialApp(home: Spike()));

class Spike extends StatefulWidget {
  const Spike({super.key});
  @override
  State<Spike> createState() => _SpikeState();
}

class _SpikeState extends State<Spike> {
  String result =
      'Ready: no network exporter, no gameplay data, 30 x 1000 spans per mode.';
  Future<void> measure() async {
    setState(() => result = 'Running bounded synthetic measurement…');
    final provider = TracerProviderBase();
    final reports = <String>[];
    for (final entry in {
      'noop': api.globalTracerProvider.getTracer('cgms.noop'),
      'recording': provider.getTracer('cgms.synthetic.spike'),
    }.entries) {
      final tracer = entry.value;
      final timings = <int>[];
      for (var warmup = 0; warmup < 1000; warmup++) {
        tracer.startSpan('synthetic').end();
      }
      for (var block = 0; block < 30; block++) {
        await Future<void>.delayed(const Duration(milliseconds: 1));
        final clock = Stopwatch()..start();
        for (var item = 0; item < 1000; item++) {
          tracer.startSpan('synthetic').end();
        }
        clock.stop();
        timings.add(clock.elapsedMicroseconds);
      }
      timings.sort();
      reports.add(
        '${entry.key} microseconds/span p50=${timings[14] / 1000} p95=${timings[28] / 1000} p99=${timings[29] / 1000}',
      );
    }
    provider.shutdown();
    if (mounted)
      setState(
        () => result =
            'PASS Flutter web span creation/end; 30000 spans per mode; '
            '${reports.join('; ')}. No export cost, frame-time or native device claim.',
      );
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    body: Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(result),
          ElevatedButton(
            onPressed: measure,
            child: const Text('Run bounded measurement'),
          ),
        ],
      ),
    ),
  );
}
