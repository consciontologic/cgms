import 'package:cgms_telemetry_spike/main.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('bounded measurement returns both modes without an exporter', (
    tester,
  ) async {
    await tester.pumpWidget(const MaterialApp(home: Spike()));
    await tester.tap(find.text('Run bounded measurement'));
    await tester.pumpAndSettle();
    expect(
      find.textContaining('PASS Flutter web span creation/end'),
      findsOneWidget,
    );
    expect(find.textContaining('noop microseconds/span'), findsOneWidget);
    expect(find.textContaining('recording microseconds/span'), findsOneWidget);
  });
}
