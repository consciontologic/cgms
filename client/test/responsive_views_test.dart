import 'package:cgms_demo/demo_controller.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  for (final width in [390.0, 768.0, 1440.0]) {
    testWidgets('all feature views at $width with large text and long name', (
      tester,
    ) async {
      tester.view.physicalSize = Size(width, 1024);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final controller = DemoController()
        ..configure(name: 'A very long player display name', games: 1);
      addTearDown(controller.dispose);
      Widget shell(Widget child) => MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: MediaQuery(
          data: MediaQueryData(
            size: Size(width, 1024),
            textScaler: const TextScaler.linear(2),
            disableAnimations: true,
          ),
          child: Scaffold(
            body: SingleChildScrollView(
              padding: const EdgeInsets.all(20),
              child: child,
            ),
          ),
        ),
      );
      for (final compact in [false, true]) {
        for (final scenario in [
          'ready',
          'pending',
          'disconnected',
          'empty',
          'hand-10',
          'hand-15',
          'combos',
        ]) {
          controller.loadScenario(scenario);
          await tester.pumpWidget(
            shell(
              CgmsTableView(
                key: ValueKey(controller.table.gameId),
                data: controller.table,
                compact: compact,
                onSelect: controller.selectCard,
                onSelectHand: controller.densityReview
                    ? controller.selectCard
                    : null,
                onClearHandSelection: controller.densityReview
                    ? controller.clearSelection
                    : null,
                onInspect: (_) {},
                onPrimary: controller.primaryAction,
                onTrade: () {},
                onScores: () {},
                onEndTurn: controller.endTurn,
                endTurnEnabled: controller.canEndTurn,
                endTurnExplanation: controller.densityReview
                    ? 'Choose the guided ending scenario to review scoring.'
                    : null,
                onReconnect: () {},
              ),
            ),
          );
          await tester.pumpAndSettle();
          expect(
            tester.takeException(),
            isNull,
            reason: '$scenario $width compact=$compact',
          );
        }
      }
      await tester.pumpWidget(
        shell(
          CgmsTradeView(
            give: const VisibleCard(
              id: 'one',
              rank: '10',
              suit: CardSuit.hearts,
              copy: 1,
            ),
            receive: const VisibleCard(
              id: 'two',
              rank: '10',
              suit: CardSuit.clubs,
              copy: 2,
            ),
            status: 'An unusually long translated transfer status',
            description: 'Offer terms provided by another consumer.',
            onOffer: () {},
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      controller.loadScenario('results');
      await tester.pumpWidget(
        shell(
          CgmsScoresView(
            seats: controller.table.seats,
            title: 'A long results heading',
            description: 'Exact scores',
            finalized: false,
            onFinish: controller.finishSettlement,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets(
    'card responds to keyboard activation and exposes distinct identity',
    (tester) async {
      var count = 0;
      final semantics = tester.ensureSemantics();
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: CgmsCard(
              card: const VisibleCard(
                id: 'clubs-7-copy2',
                rank: '7',
                suit: CardSuit.clubs,
                copy: 2,
              ),
              showArt: false,
              onTap: () => count++,
            ),
          ),
        ),
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.tab);
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      await tester.pump();
      expect(count, 1);
      expect(
        find.bySemanticsLabel(RegExp('7 of Clubs, copy 2')),
        findsOneWidget,
      );
      semantics.dispose();
    },
  );
}
