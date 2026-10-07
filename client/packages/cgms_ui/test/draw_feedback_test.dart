import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

TableDrawFeedback notice({
  String id = 'game/turn-2/4',
  int seat = 1,
  DrawDestination destination = DrawDestination.hand,
  DateTime? startedAt,
}) => TableDrawFeedback(
  id: id,
  seat: seat,
  destination: destination,
  startedAt: startedAt ?? DateTime.now(),
);

Widget host({
  TableDrawFeedback? feedback,
  bool reduced = false,
  VoidCallback? onTap,
  FocusNode? focus,
  Key? feedbackKey,
}) => MaterialApp(
  theme: CgmsTheme.dark(),
  localizationsDelegates: CgmsLocalizations.localizationsDelegates,
  supportedLocales: CgmsLocalizations.supportedLocales,
  home: MediaQuery(
    data: MediaQueryData(disableAnimations: reduced),
    child: Center(
      child: CgmsDrawFeedback(
        key: feedbackKey,
        feedback: feedback,
        viewerSeat: 1,
        child: SizedBox(
          width: 48,
          height: 60,
          child: TextButton(
            focusNode: focus,
            onPressed: onTap ?? () {},
            child: const Text('Pile'),
          ),
        ),
      ),
    ),
  ),
);

void main() {
  testWidgets(
    'draw feedback retains pile target, focus and quiet privacy-safe announcement',
    (tester) async {
      final semantics = tester.ensureSemantics();
      final focus = FocusNode();
      addTearDown(focus.dispose);
      var taps = 0;
      await tester.pumpWidget(host(focus: focus, onTap: () => taps++));
      focus.requestFocus();
      await tester.pump();
      final before = tester.getRect(find.byType(TextButton));
      final draw = notice(destination: DrawDestination.concealedAce);
      await tester.pumpWidget(
        host(feedback: draw, reduced: true, focus: focus, onTap: () => taps++),
      );
      expect(tester.getRect(find.byType(TextButton)), before);
      expect(focus.hasFocus, isTrue);
      expect(
        find.byKey(const ValueKey('draw-highlight-quiet')),
        findsOneWidget,
      );
      expect(find.byKey(const ValueKey('draw-highlight')), findsNothing);
      expect(
        find.bySemanticsLabel(
          'Turn draw: one card added to your concealed Ace area.',
        ),
        findsOneWidget,
      );
      expect(find.byType(Image), findsNothing);
      await tester.tap(find.byType(TextButton));
      expect(taps, 1);
      await tester.pump(const Duration(milliseconds: 300));
      final decoration =
          tester
                  .widget<DecoratedBox>(
                    find.byKey(const ValueKey('draw-highlight-quiet')),
                  )
                  .decoration
              as BoxDecoration;
      expect(decoration.boxShadow, isEmpty);
      await tester.pump(const Duration(seconds: 2));
      expect(find.byKey(const ValueKey('draw-highlight-quiet')), findsNothing);
      expect(focus.hasFocus, isTrue);
      semantics.dispose();
    },
  );

  testWidgets(
    'same notice rebuild never restarts or restores expired feedback',
    (tester) async {
      await tester.pumpWidget(host());
      final draw = notice();
      await tester.pumpWidget(host(feedback: draw));
      expect(find.byKey(const ValueKey('draw-highlight')), findsOneWidget);
      await tester.pump(const Duration(milliseconds: 700));
      await tester.pumpWidget(host(feedback: draw));
      await tester.pump(const Duration(milliseconds: 500));
      expect(find.byKey(const ValueKey('draw-highlight')), findsNothing);
      await tester.pumpWidget(host(feedback: draw));
      expect(find.byKey(const ValueKey('draw-highlight')), findsNothing);
      final expired = notice(
        startedAt: DateTime.now().subtract(const Duration(seconds: 3)),
      );
      await tester.pumpWidget(
        host(feedback: expired, feedbackKey: const ValueKey('new-layout')),
      );
      expect(find.byKey(const ValueKey('draw-highlight')), findsNothing);
    },
  );

  testWidgets(
    'opponent feedback identifies only seat while ordinary draws identify the private hand',
    (tester) async {
      final semantics = tester.ensureSemantics();
      await tester.pumpWidget(host());
      await tester.pumpWidget(
        host(
          feedback: notice(seat: 2, destination: DrawDestination.concealedAce),
        ),
      );
      expect(find.bySemanticsLabel('Seat 2 drew one card.'), findsOneWidget);
      expect(
        find.bySemanticsLabel(RegExp('Ace|diamonds|clubs|copy')),
        findsNothing,
      );
      await tester.pumpWidget(host(feedback: notice(id: 'next-draw')));
      expect(
        find.bySemanticsLabel('Turn draw: one card added to your hand.'),
        findsOneWidget,
      );
      await tester.pump(const Duration(seconds: 2));
      semantics.dispose();
    },
  );
}
