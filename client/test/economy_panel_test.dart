import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:cgms_ui/cgms_ui.dart';
import 'package:cgms_demo/online/economy_panel.dart';
import 'package:cgms_demo/online/online_screen.dart';
import 'online/economy_client_test.dart' show signedIn, balance, receipt;
import 'online/online_client_test.dart' show FakeTransport, MemoryOperations;
import 'package:cgms_demo/online/online_client.dart';

const _adFreeTerms =
    'Ad-free: 30 days without ads. The 3 free online starts per day still apply '
    'unless you also have unlimited access.';
const _paidDailyTerms =
    'Paid Daily: 24 hours of unlimited online starts and no ads from confirmed '
    'purchase. Does not renew automatically.';
const _paidLongerTerms =
    'Paid Weekly, Monthly and Yearly: unlimited online starts and no ads.';

void main() {
  testWidgets('paid terms explain approved offers without granting benefits', (
    tester,
  ) async {
    final t = FakeTransport();
    final c = await signedIn(t, MemoryOperations());
    for (final flags in [
      (unlimited: false, noAds: false),
      (unlimited: false, noAds: true),
      (unlimited: true, noAds: false),
      (unlimited: true, noAds: true),
    ]) {
      t.handle = (_, _, _) async => OnlineResponse(200, {
        ...balance,
        'benefits': {
          ...balance['benefits'] as Map,
          'unlimited': flags.unlimited,
          'no_ads': flags.noAds,
        },
      });
      await c.refreshEconomy();
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: SingleChildScrollView(child: EconomyPanel(client: c)),
          ),
        ),
      );
      await tester.pumpAndSettle();
      for (final terms in [_adFreeTerms, _paidDailyTerms, _paidLongerTerms]) {
        expect(find.text(terms), findsOneWidget);
      }
      expect(
        find.text(
          'Earned Daily and Weekly access allow unlimited online starts and '
          'keep ads unless separate ad-free access is active. Time extends '
          'from now or your current pass expiry, whichever is later.',
        ),
        findsOneWidget,
      );
      expect(
        find.text('Unlimited online starts'),
        flags.unlimited ? findsOneWidget : findsNothing,
      );
      expect(
        find.text('Ad-free access active'),
        flags.noAds ? findsOneWidget : findsNothing,
      );
      expect(
        find.text('Ads remain active'),
        flags.noAds ? findsNothing : findsOneWidget,
      );
      expect(
        find.text(
          'Paid Ad-free, Daily, Weekly, Monthly and Yearly products are not '
          'available in this build.',
        ),
        findsOneWidget,
      );
      expect(find.byType(FilledButton), findsOneWidget);
      expect(find.byKey(const ValueKey('buy-pass')), findsOneWidget);
      expect(tester.takeException(), isNull);
    }
    expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
    await tester.pumpWidget(const SizedBox());
    c.dispose();
    await t.frames.close();
  });

  for (final size in [
    const Size(390, 844),
    const Size(768, 1024),
    const Size(1440, 900),
  ]) {
    for (final largeText in [false, true]) {
      for (final rtl in [false, true]) {
        testWidgets(
          'paid terms and controls stay reachable at $size large=$largeText rtl=$rtl',
          (tester) async {
            tester.view.physicalSize = size;
            tester.view.devicePixelRatio = 1;
            addTearDown(tester.view.resetPhysicalSize);
            addTearDown(tester.view.resetDevicePixelRatio);
            final t = FakeTransport();
            final c = await signedIn(t, MemoryOperations());
            t.handle = (_, path, _) async => OnlineResponse(
              200,
              path == '/v1/session'
                  ? {
                      'account': {'id': 'a'},
                      'csrf_token': 'csrf',
                    }
                  : balance,
            );
            await tester.pumpWidget(
              MaterialApp(
                theme: CgmsTheme.dark(),
                localizationsDelegates:
                    CgmsLocalizations.localizationsDelegates,
                supportedLocales: CgmsLocalizations.supportedLocales,
                home: OnlineScreen(client: c),
              ),
            );
            await tester.pumpAndSettle();
            for (final setting in [
              if (rtl) 'Right to left',
              if (largeText) '200% text',
            ]) {
              await tester.tap(find.byTooltip('Display settings'));
              await tester.pumpAndSettle();
              await tester.tap(
                find.widgetWithText(CheckedPopupMenuItem<String>, setting),
              );
              await tester.pumpAndSettle();
            }
            await tester.tap(find.byTooltip('Allowance and dirt'));
            await tester.pumpAndSettle();
            final dialog = find.byType(Dialog);
            final panel = find.descendant(
              of: dialog,
              matching: find.byType(EconomyPanel),
            );
            final viewport = find.descendant(
              of: dialog,
              matching: find.byType(SingleChildScrollView),
            );
            expect(
              Directionality.of(tester.element(panel)),
              rtl ? TextDirection.rtl : TextDirection.ltr,
            );
            expect(
              MediaQuery.textScalerOf(tester.element(panel)).scale(16),
              largeText ? 32 : 16,
            );
            for (final terms in [
              _adFreeTerms,
              _paidDailyTerms,
              _paidLongerTerms,
            ]) {
              final text = find.descendant(
                of: panel,
                matching: find.text(terms),
              );
              expect(text, findsOneWidget);
              await tester.ensureVisible(text);
              await tester.pumpAndSettle();
              final bounds = tester.getRect(text);
              final visible = tester.getRect(viewport);
              expect(bounds.left, greaterThanOrEqualTo(visible.left));
              expect(bounds.right, lessThanOrEqualTo(visible.right));
              expect(bounds.top, greaterThanOrEqualTo(visible.top - .01));
              expect(bounds.bottom, lessThanOrEqualTo(visible.bottom + .01));
              expect(tester.takeException(), isNull);
            }
            for (final control in [
              find.byKey(const ValueKey('buy-pass')),
              find.widgetWithText(TextButton, 'Refresh allowance'),
              find.widgetWithText(TextButton, 'Close'),
            ]) {
              final button = find.descendant(of: dialog, matching: control);
              await tester.ensureVisible(button);
              await tester.pumpAndSettle();
              expect(button.hitTestable(), findsOneWidget);
              expect(tester.getSize(button).height, greaterThanOrEqualTo(48));
              expect(tester.getSize(button).width, greaterThanOrEqualTo(48));
              expect(tester.takeException(), isNull);
            }
            expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
            await tester.tap(find.widgetWithText(TextButton, 'Close'));
            await tester.pumpAndSettle();
            expect(dialog, findsNothing);
            await tester.pumpWidget(const SizedBox());
            c.dispose();
            await t.frames.close();
          },
        );
      }
    }
  }

  testWidgets(
    'ad-free expiry follows authoritative flags independently of access',
    (tester) async {
      final t = FakeTransport();
      final c = await signedIn(t, MemoryOperations());
      for (final state in [
        (unlimited: false, noAds: true, expiry: '2030-01-02T03:04:05Z'),
        (unlimited: true, noAds: false, expiry: '1970-01-01T00:00:00Z'),
        (unlimited: true, noAds: true, expiry: '2030-01-02T03:04:05Z'),
        (unlimited: false, noAds: false, expiry: '2000-01-02T03:04:05Z'),
      ]) {
        t.handle = (_, _, _) async => OnlineResponse(200, {
          ...balance,
          'benefits': {
            ...balance['benefits'] as Map,
            'unlimited': state.unlimited,
            'no_ads': state.noAds,
            'ad_free_until': state.expiry,
          },
        });
        await c.refreshEconomy();
        await tester.pumpWidget(
          MaterialApp(
            theme: CgmsTheme.dark(),
            localizationsDelegates: CgmsLocalizations.localizationsDelegates,
            supportedLocales: CgmsLocalizations.supportedLocales,
            home: Scaffold(
              body: SingleChildScrollView(child: EconomyPanel(client: c)),
            ),
          ),
        );
        await tester.pumpAndSettle();
        expect(
          find.text('Unlimited online starts'),
          state.unlimited ? findsOneWidget : findsNothing,
        );
        expect(
          find.text('Ad-free access active'),
          state.noAds ? findsOneWidget : findsNothing,
        );
        expect(
          find.text('Ads remain active'),
          state.noAds ? findsNothing : findsOneWidget,
        );
        expect(
          find.text('Ad-free expiry: 2030-01-02 03:04:05 UTC'),
          state.noAds ? findsOneWidget : findsNothing,
        );
        expect(
          find.textContaining('Ad-free expiry:'),
          state.noAds ? findsOneWidget : findsNothing,
        );
        expect(
          find.textContaining(
            'Paid Ad-free, Daily, Weekly, Monthly and Yearly products are not available',
          ),
          findsOneWidget,
        );
        expect(tester.takeException(), isNull);
      }
      expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
      await tester.pumpWidget(const SizedBox());
      c.dispose();
      await t.frames.close();
    },
  );

  testWidgets(
    'older ad-free account omits unknown expiry without inventing a date',
    (tester) async {
      final t = FakeTransport();
      final c = await signedIn(t, MemoryOperations());
      for (final expiry in [
        <String, dynamic>{},
        {'ad_free_until': '1970-01-01T00:00:00Z'},
      ]) {
        t.handle = (_, _, _) async => OnlineResponse(200, {
          ...balance,
          'benefits': {
            ...balance['benefits'] as Map,
            'no_ads': true,
            ...expiry,
          },
        });
        await c.refreshEconomy();
        await tester.pumpWidget(
          MaterialApp(
            theme: CgmsTheme.dark(),
            localizationsDelegates: CgmsLocalizations.localizationsDelegates,
            supportedLocales: CgmsLocalizations.supportedLocales,
            home: Scaffold(
              body: SingleChildScrollView(child: EconomyPanel(client: c)),
            ),
          ),
        );
        await tester.pumpAndSettle();
        expect(find.text('Ad-free access active'), findsOneWidget);
        expect(find.textContaining('Ad-free expiry:'), findsNothing);
        expect(find.textContaining('1970'), findsNothing);
      }
      await tester.pumpWidget(const SizedBox());
      c.dispose();
      await t.frames.close();
    },
  );

  testWidgets('disabled account ignores unused malformed benefits and prices', (
    tester,
  ) async {
    final t = FakeTransport();
    final c = await signedIn(t, MemoryOperations());
    t.handle = (_, _, _) async => const OnlineResponse(200, {
      'enabled': false,
      'benefits': [],
      'pass_day_price': 'unavailable',
    });
    await c.refreshEconomy();
    await tester.pumpWidget(
      MaterialApp(
        theme: CgmsTheme.dark(),
        localizationsDelegates: CgmsLocalizations.localizationsDelegates,
        supportedLocales: CgmsLocalizations.supportedLocales,
        home: Scaffold(body: EconomyPanel(client: c)),
      ),
    );
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(
      find.text('Online allowances are not enabled on this server.'),
      findsOneWidget,
    );
    expect(find.byKey(const ValueKey('buy-pass')), findsNothing);
    expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
    await tester.pumpWidget(const SizedBox());
    c.dispose();
    await t.frames.close();
  });

  testWidgets(
    'malformed refresh keeps confirmed text but disables new spending until recovery',
    (tester) async {
      final t = FakeTransport();
      final c = await signedIn(t, MemoryOperations());
      final malformed = Map<String, dynamic>.from(balance)
        ..['pass_day_price'] = 0;
      t.handle = (_, _, _) async => OnlineResponse(200, malformed);
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: SingleChildScrollView(child: EconomyPanel(client: c)),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('180 dirt'), findsOneWidget);
      expect(find.text('Spend 30 dirt'), findsOneWidget);
      expect(
        find.text('Could not refresh your allowance. Try again.'),
        findsOneWidget,
      );
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('buy-pass')))
            .onPressed,
        isNull,
      );
      expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
      t.handle = (_, _, _) async => const OnlineResponse(200, balance);
      await tester.ensureVisible(find.text('Refresh allowance'));
      await tester.tap(find.text('Refresh allowance'));
      await tester.pumpAndSettle();
      expect(
        find.text('Could not refresh your allowance. Try again.'),
        findsNothing,
      );
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('buy-pass')))
            .onPressed,
        isNotNull,
      );
      expect(t.calls.where((call) => call['method'] == 'POST'), isEmpty);
      await tester.pumpWidget(const SizedBox());
      c.dispose();
      await t.frames.close();
    },
  );

  testWidgets(
    'account dialog keeps 48 pixel controls under compact desktop density',
    (tester) async {
      final t = FakeTransport();
      final c = await signedIn(t, MemoryOperations());
      t.handle = (_, path, _) async => OnlineResponse(
        200,
        path == '/v1/session'
            ? {
                'account': {'id': 'a'},
                'csrf_token': 'csrf',
              }
            : balance,
      );
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark().copyWith(
            visualDensity: VisualDensity.compact,
          ),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: OnlineScreen(client: c),
        ),
      );
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('Allowance and dirt'));
      await tester.pumpAndSettle();
      final dialog = find.byType(Dialog);
      final buy = find.descendant(
        of: dialog,
        matching: find.byKey(const ValueKey('buy-pass')),
      );
      expect(tester.getSize(buy).height, greaterThanOrEqualTo(48));
      for (final button
          in find
              .descendant(of: dialog, matching: find.byType(TextButton))
              .evaluate()) {
        expect(
          tester.getSize(find.byWidget(button.widget)).height,
          greaterThanOrEqualTo(48),
        );
      }
      await tester.pumpWidget(const SizedBox());
      c.dispose();
      await t.frames.close();
    },
  );
  testWidgets(
    'allowance panel explains UTC, offline limits and explicit pass cost',
    (tester) async {
      final t = FakeTransport();
      final c = await signedIn(t, MemoryOperations());
      t.handle = (_, _, _) async => OnlineResponse(200, {
        ...balance,
        'benefits': {
          ...balance['benefits'] as Map,
          'pass_until': '1970-01-01T00:00:00Z',
          'premium_until': '1970-01-01T00:00:00Z',
        },
      });
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: SingleChildScrollView(child: EconomyPanel(client: c)),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('180 dirt'), findsOneWidget);
      expect(find.textContaining('Earned Daily and Weekly'), findsOneWidget);
      expect(
        find.textContaining('Ad-free, Daily, Weekly, Monthly and Yearly'),
        findsOneWidget,
      );
      expect(
        find.textContaining('Unlimited online starts · no ads'),
        findsNothing,
      );
      final dropdown = tester.widget<DropdownButton<int>>(
        find.byType(DropdownButton<int>),
      );
      expect(dropdown.items!.map((item) => item.value).toList(), [1, 7]);
      expect(dropdown.value, 1);
      expect(find.textContaining('1970'), findsNothing);
      expect(find.text('3 free online starts remaining today'), findsOneWidget);
      expect(find.textContaining('Offline play earns no dirt'), findsOneWidget);
      expect(t.calls.where((r) => r['path'] == '/v1/economy/passes'), isEmpty);
      t.handle = (method, _, body) async =>
          OnlineResponse(200, method == 'POST' ? receipt(body!) : balance);
      await tester.tap(find.byKey(const ValueKey('buy-pass')));
      await tester.pumpAndSettle();
      expect(c.passReceipt!['cost'], 30);
      expect(find.textContaining('Pass confirmed'), findsOneWidget);
      await tester.pumpWidget(const SizedBox());
      c.dispose();
      await t.frames.close();
    },
  );
  testWidgets('ad removal is displayed independently from unlimited starts', (
    tester,
  ) async {
    final t = FakeTransport();
    final c = await signedIn(t, MemoryOperations());
    for (final flags in [(true, false), (false, true), (true, true)]) {
      t.handle = (_, _, _) async => OnlineResponse(200, {
        ...balance,
        'benefits': {
          ...balance['benefits'] as Map,
          'unlimited': flags.$1,
          'no_ads': flags.$2,
        },
      });
      await c.refreshEconomy();
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: Scaffold(
            body: SingleChildScrollView(child: EconomyPanel(client: c)),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(
        find.text('Unlimited online starts'),
        flags.$1 ? findsOneWidget : findsNothing,
      );
      expect(
        find.text('Ad-free access active'),
        flags.$2 ? findsOneWidget : findsNothing,
      );
      expect(
        find.text('Ads remain active'),
        flags.$2 ? findsNothing : findsOneWidget,
      );
    }
    await tester.pumpWidget(const SizedBox());
    c.dispose();
    await t.frames.close();
  });
  testWidgets(
    'uncertain pass disables spending and offers receipt recovery at large text',
    (tester) async {
      tester.view.physicalSize = const Size(390, 844);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);
      final t = FakeTransport();
      final c = await signedIn(t, MemoryOperations());
      t.handle = (_, _, _) async =>
          const OnlineResponse(503, {'code': 'OUTCOME_UNKNOWN'});
      await c.buyPass(1);
      t.handle = (_, _, _) async => OnlineResponse(200, {
        ...balance,
        'benefits': {
          ...balance['benefits'] as Map,
          'pass_until': '1970-01-01T00:00:00Z',
          'premium_until': '1970-01-01T00:00:00Z',
        },
      });
      await tester.pumpWidget(
        MaterialApp(
          theme: CgmsTheme.dark(),
          localizationsDelegates: CgmsLocalizations.localizationsDelegates,
          supportedLocales: CgmsLocalizations.supportedLocales,
          home: MediaQuery(
            data: const MediaQueryData(textScaler: TextScaler.linear(2)),
            child: Scaffold(
              body: SingleChildScrollView(child: EconomyPanel(client: c)),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      expect(find.byKey(const ValueKey('recover-pass')), findsOneWidget);
      expect(
        tester
            .widget<FilledButton>(find.byKey(const ValueKey('buy-pass')))
            .onPressed,
        isNull,
      );
      await tester.pumpWidget(const SizedBox());
      c.dispose();
      await t.frames.close();
    },
  );
}
