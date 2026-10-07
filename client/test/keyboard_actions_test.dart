import 'package:cgms_demo/keyboard_actions.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  for (final forward in [true, false]) {
    for (final edge in [
      TraversalEdgeBehavior.closedLoop,
      TraversalEdgeBehavior.leaveFlutterView,
    ]) {
      testWidgets(
        'browser traversal $forward respects $edge with one control',
        (tester) async {
          final control = FocusNode();
          final scope = FocusScopeNode(traversalEdgeBehavior: edge);
          addTearDown(control.dispose);
          addTearDown(scope.dispose);
          await tester.pumpWidget(
            MaterialApp(
              home: FocusScope(
                node: scope,
                child: TextButton(
                  focusNode: control,
                  autofocus: true,
                  onPressed: () {},
                  child: const Text('Close inspection'),
                ),
              ),
            ),
          );
          await tester.pumpAndSettle();
          expect(control.hasPrimaryFocus, isTrue);
          final handled = forward
              ? BrowserNextFocusAction().invoke(const NextFocusIntent())
              : BrowserPreviousFocusAction().invoke(
                  const PreviousFocusIntent(),
                );
          await tester.pumpAndSettle();
          expect(handled, edge == TraversalEdgeBehavior.closedLoop);
          if (edge == TraversalEdgeBehavior.closedLoop) {
            expect(control.hasPrimaryFocus, isTrue);
          }
        },
      );
    }
  }
}
