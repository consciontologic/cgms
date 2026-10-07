import 'dart:ui' show Tristate;

import 'package:cgms_ui/src/components.dart';
import 'package:cgms_ui/src/l10n/cgms_localizations.dart';
import 'package:cgms_ui/src/models.dart';
import 'package:cgms_ui/src/theme.dart';
import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  setUpAll(() async {
    // Density assertions need the package typefaces, not the square Ahem test
    // fallback. This also exercises the real identity metrics at 200% scale.
    for (final font in {
      'Barlow Condensed': 'BarlowCondensed-Bold.ttf',
      'Atkinson Hyperlegible Next': 'AtkinsonHyperlegibleNext-Regular.ttf',
    }.entries) {
      await (FontLoader('packages/cgms_ui/${font.key}')..addFont(
            rootBundle.load('packages/cgms_ui/assets/fonts/${font.value}'),
          ))
          .load();
    }
  });

  test('reading colors meet contrast on every dark reading surface', () {
    for (final background in [
      CgmsColors.canvas,
      CgmsColors.surface,
      CgmsColors.raised,
    ]) {
      for (final foreground in [
        CgmsColors.text,
        CgmsColors.muted,
        CgmsColors.heart,
        CgmsColors.diamond,
        CgmsColors.club,
        CgmsColors.spade,
        CgmsColors.action,
        CgmsColors.cyan,
        CgmsColors.error,
      ]) {
        expect(_contrast(foreground, background), greaterThanOrEqualTo(4.5));
      }
      expect(_contrast(CgmsColors.lime, background), greaterThanOrEqualTo(3));
      expect(_contrast(CgmsColors.border, background), greaterThanOrEqualTo(3));
    }
    expect(
      _contrast(CgmsColors.canvas, CgmsColors.action),
      greaterThanOrEqualTo(4.5),
    );
  });

  group('ExactPoints', () {
    test('normalizes rational values without floating point rounding', () {
      expect(ExactPoints.fromStrings('10', '4').format(), '5/2');
      expect(ExactPoints.fromStrings('-10', '-4').format(), '5/2');
      expect(ExactPoints.fromStrings('0', '-5').format(), '0');
      expect(ExactPoints.fromStrings('10', '-4').format(), '-5/2');
      expect(ExactPoints.fromStrings('20', '2').format(), '10');
      expect(() => ExactPoints.fromStrings('1', '0'), throwsArgumentError);
    });

    test('retains values larger than web safe integers and exact division', () {
      final large = ExactPoints.fromStrings('9007199254740993000000001');
      final half = ExactPoints.fromStrings('1', '2');
      expect((large + half).format(), '18014398509481986000000003/2');
      expect((half + half).format(), '1');
      expect((ExactPoints.integer(7) / ExactPoints.integer(3)).format(), '7/3');
      expect((ExactPoints.integer(5) - half).format(), '9/2');
      expect((half * ExactPoints.integer(3)).format(), '3/2');
      expect(half, ExactPoints.fromStrings('2', '4'));
      expect(half.compareTo(ExactPoints.integer(1)), lessThan(0));
    });
  });

  test('matching ranks retain physical identity and copy labels', () {
    const first = VisibleCard(
      id: 'deck1-clubs-8',
      rank: '8',
      suit: CardSuit.clubs,
      copy: 1,
    );
    const second = VisibleCard(
      id: 'deck2-clubs-8',
      rank: '8',
      suit: CardSuit.clubs,
      copy: 2,
    );
    expect(first.id, isNot(second.id));
    expect(first.label, contains('copy 1'));
    expect(second.label, contains('copy 2'));
  });

  test(
    'display status omits collection names while preserving meaningful state',
    () {
      for (final status in ['hand', 'series', 'concealed-ace', ' HAND ']) {
        final card = VisibleCard(
          id: 'status',
          rank: 'A',
          suit: CardSuit.hearts,
          copy: 1,
          status: status,
        );
        expect(card.displayStatus, isEmpty);
        expect(card.status, status);
      }
      for (final status in [
        'Used this turn',
        'Available from round 4',
        'Locked',
      ]) {
        final card = VisibleCard(
          id: 'status',
          rank: 'A',
          suit: CardSuit.hearts,
          copy: 1,
          status: status,
        );
        expect(card.displayStatus, status);
      }
    },
  );

  Widget shell(Widget child) => MaterialApp(
    theme: CgmsTheme.dark(),
    localizationsDelegates: CgmsLocalizations.localizationsDelegates,
    supportedLocales: CgmsLocalizations.supportedLocales,
    home: Scaffold(body: Center(child: child)),
  );

  testWidgets('dialog titles render with readable dark-surface contrast', (
    tester,
  ) async {
    await tester.pumpWidget(
      shell(const AlertDialog(title: Text('Card details'))),
    );
    final paragraph = tester.renderObject<RenderParagraph>(
      find.descendant(
        of: find.text('Card details'),
        matching: find.byType(RichText),
      ),
    );
    final color = paragraph.text.style?.color;
    expect(color, isNotNull, reason: 'Resolve a visible dialog foreground.');
    expect(_contrast(color!, CgmsColors.surface), greaterThanOrEqualTo(4.5));
  });

  test(
    'all four native illustration sets are bundled for offline rendering',
    () async {
      for (final style in CardStyle.values) {
        final prefix = style == CardStyle.firstLight ? '' : '${style.slug}/';
        for (final suit in CardSuit.values) {
          for (final rank in [
            'A',
            '2',
            '3',
            '4',
            '5',
            '6',
            '7',
            '8',
            '9',
            '10',
            'J',
            'Q',
            'K',
          ]) {
            final bytes = await rootBundle.load(
              'packages/cgms_ui/assets/art/$prefix${suit.name}-$rank.png',
            );
            expect(bytes.lengthInBytes, greaterThan(1000));
            final card = VisibleCard(
              id: 'asset',
              rank: rank,
              suit: suit,
              copy: 1,
              style: style,
            );
            // PNG IHDR dimensions are big-endian at bytes 16 and 20.
            expect(
              card.artworkAspectRatio,
              bytes.getUint32(16) / bytes.getUint32(20),
              reason: card.artworkAsset,
            );
          }
        }
      }
    },
  );

  testWidgets(
    'hand envelopes are uniform while every artwork style remains contained',
    (tester) async {
      for (final style in CardStyle.values) {
        for (final suit in CardSuit.values) {
          for (final rank in ['2', '10', 'J', 'K']) {
            final card = VisibleCard(
              id: 'uniform',
              rank: rank,
              suit: suit,
              copy: 2,
              style: style,
            );
            await tester.pumpWidget(
              shell(
                CgmsCard(
                  card: card,
                  width: 72,
                  handStackPreview: true,
                  onTap: () {},
                  onInspect: () {},
                ),
              ),
            );
            expect(tester.getSize(find.byType(CgmsCard)), const Size(72, 90));
            final art = tester.getRect(
              find.byKey(const ValueKey('card-art-uniform')),
            );
            final frame = tester.getRect(
              find.byKey(const ValueKey('card-frame-uniform')),
            );
            expect(
              art.width / art.height,
              closeTo(card.artworkAspectRatio, .001),
            );
            expect(frame.contains(art.center), isTrue);
            expect(art.width, lessThanOrEqualTo(frame.width));
            expect(art.height, lessThanOrEqualTo(frame.height));
            final image = tester.widget<Image>(find.byType(Image));
            expect(image.fit, BoxFit.contain);
            expect((image.image as AssetImage).assetName, card.artworkAsset);
            expect(tester.takeException(), isNull);
          }
        }
      }
    },
  );

  testWidgets(
    'every card mode shows uncropped art with identity outside the painting',
    (tester) async {
      for (final mode in ['full', 'compact', 'stack', 'thumbnail']) {
        await tester.pumpWidget(
          shell(
            CgmsCard(
              card: const VisibleCard(
                id: 'full-art',
                rank: '10',
                suit: CardSuit.clubs,
                copy: 2,
                status: 'hand',
              ),
              width: 100,
              compact: mode == 'compact' || mode == 'stack',
              stackHeader: mode == 'stack',
              thumbnail: mode == 'thumbnail',
              selected: true,
              onInspect: () {},
            ),
          ),
        );
        await tester.pumpAndSettle();
        final art = find.byType(Image);
        final artRect = tester.getRect(art);
        expect(artRect.width / artRect.height, closeTo(281 / 359, .001));
        expect(tester.widget<Image>(art).fit, BoxFit.contain);
        expect(
          tester.getRect(find.text('10')).bottom,
          lessThanOrEqualTo(artRect.top),
        );
        expect(
          tester.getRect(find.byIcon(Icons.check)).bottom,
          lessThanOrEqualTo(artRect.top),
        );
        expect(find.textContaining('Copy'), findsNothing);
        expect(find.text('2'), findsNothing);
        expect(find.text('hand'), findsNothing);
        final magnify = tester.getRect(
          find.byKey(const ValueKey('magnify-full-art')),
        );
        expect(magnify.width, greaterThanOrEqualTo(48));
        expect(magnify.height, greaterThanOrEqualTo(48));
        expect(magnify.right, closeTo(artRect.right, .001));
        expect(magnify.bottom, closeTo(artRect.bottom, .001));
        expect(tester.takeException(), isNull);
      }
    },
  );

  testWidgets('double tap magnifies independently of single tap selection', (
    tester,
  ) async {
    var selected = 0;
    var inspected = 0;
    await tester.pumpWidget(
      shell(
        CgmsCard(
          card: const VisibleCard(
            id: 'double-tap',
            rank: '8',
            suit: CardSuit.clubs,
            copy: 1,
          ),
          compact: true,
          showArt: false,
          onTap: () => selected++,
          onInspect: () => inspected++,
        ),
      ),
    );
    final target = find.byKey(const ValueKey('select-double-tap'));
    await tester.tap(target);
    await tester.pump(const Duration(milliseconds: 50));
    await tester.tap(target);
    await tester.pumpAndSettle();
    expect(inspected, 1);
    expect(selected, 0);
    await tester.tap(target);
    await tester.pump(const Duration(milliseconds: 350));
    expect(selected, 1);
    expect(inspected, 1);
    await tester.longPress(target);
    await tester.pumpAndSettle();
    expect(inspected, 1, reason: 'Holding a card must not magnify it.');
  });

  testWidgets(
    'semantic taps retain single selection and double magnification',
    (tester) async {
      final handle = tester.ensureSemantics();
      var selected = 0;
      var inspected = 0;
      await tester.pumpWidget(
        shell(
          CgmsCard(
            card: const VisibleCard(
              id: 'semantic-taps',
              rank: '8',
              suit: CardSuit.clubs,
              copy: 1,
            ),
            showArt: false,
            onTap: () => selected++,
            onInspect: () => inspected++,
          ),
        ),
      );
      void tapSemantics() {
        final node = tester.getSemantics(
          find.byKey(const ValueKey('select-semantic-taps')),
        );
        node.owner!.performAction(node.id, SemanticsAction.tap);
      }

      tapSemantics();
      await tester.pump(const Duration(milliseconds: 80));
      expect(
        selected,
        0,
        reason:
            'The first web semantic click must wait for a possible double tap.',
      );
      tapSemantics();
      await tester.pump(const Duration(milliseconds: 350));
      expect(inspected, 1);
      expect(selected, 0);
      tapSemantics();
      await tester.pump(const Duration(milliseconds: 350));
      expect(selected, 1);
      expect(inspected, 1);
      handle.dispose();
    },
  );

  testWidgets(
    'inspection and card lifecycle cancel pending semantic selection',
    (tester) async {
      final handle = tester.ensureSemantics();
      final selected = <String>[];
      var inspected = 0;
      Widget card(String id) => shell(
        CgmsCard(
          card: VisibleCard(id: id, rank: '8', suit: CardSuit.clubs, copy: 1),
          showArt: false,
          onTap: () => selected.add(id),
          onInspect: () => inspected++,
        ),
      );
      await tester.pumpWidget(card('pending-selection'));
      final target = find.byKey(const ValueKey('select-pending-selection'));
      void queueSelection() {
        final node = tester.getSemantics(target);
        node.owner!.performAction(node.id, SemanticsAction.tap);
      }

      queueSelection();
      await tester.pump(const Duration(milliseconds: 80));
      await tester.tap(find.byKey(const ValueKey('magnify-pending-selection')));
      await tester.pump(const Duration(milliseconds: 350));
      expect(inspected, 1);
      expect(selected, isEmpty);
      queueSelection();
      final ink = tester.widget<InkWell>(
        find.descendant(of: target, matching: find.byType(InkWell)),
      );
      ink.focusNode!.requestFocus();
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
      await tester.pump(const Duration(milliseconds: 350));
      expect(inspected, 2);
      expect(selected, isEmpty);
      queueSelection();
      final node = tester.getSemantics(target);
      node.owner!.performAction(
        node.id,
        SemanticsAction.customAction,
        node.getSemanticsData().customSemanticsActionIds!.single,
      );
      await tester.pump(const Duration(milliseconds: 350));
      expect(inspected, 3);
      expect(selected, isEmpty);
      queueSelection();
      await tester.pump(const Duration(milliseconds: 80));
      await tester.pumpWidget(card('replacement'));
      await tester.pump(const Duration(milliseconds: 350));
      expect(
        selected,
        isEmpty,
        reason: 'A pending tap must not select the replacement physical card.',
      );
      final replacement = tester.getSemantics(
        find.byKey(const ValueKey('select-replacement')),
      );
      replacement.owner!.performAction(replacement.id, SemanticsAction.tap);
      await tester.pumpWidget(shell(const SizedBox()));
      await tester.pump(const Duration(milliseconds: 350));
      expect(selected, isEmpty);
      expect(tester.takeException(), isNull);
      handle.dispose();
    },
  );

  testWidgets('semantic double tap magnifies an inspect-only card', (
    tester,
  ) async {
    final handle = tester.ensureSemantics();
    var inspected = 0;
    await tester.pumpWidget(
      shell(
        CgmsCard(
          card: const VisibleCard(
            id: 'semantic-inspect-only',
            rank: '8',
            suit: CardSuit.clubs,
            copy: 2,
          ),
          showArt: false,
          onInspect: () => inspected++,
        ),
      ),
    );
    final node = tester.getSemantics(
      find.byKey(const ValueKey('select-semantic-inspect-only')),
    );
    node.owner!.performAction(node.id, SemanticsAction.tap);
    await tester.pump(const Duration(milliseconds: 350));
    expect(inspected, 0);
    node.owner!.performAction(node.id, SemanticsAction.tap);
    await tester.pump(const Duration(milliseconds: 80));
    node.owner!.performAction(node.id, SemanticsAction.tap);
    await tester.pump(const Duration(milliseconds: 350));
    expect(inspected, 1);
    handle.dispose();
  });

  testWidgets(
    'inspect-only cards preserve focus and semantic magnification without tap selection',
    (tester) async {
      final handle = tester.ensureSemantics();
      var inspected = 0;
      await tester.pumpWidget(
        shell(
          CgmsCard(
            card: const VisibleCard(
              id: 'inspect-only',
              rank: '8',
              suit: CardSuit.clubs,
              copy: 2,
            ),
            compact: true,
            showArt: false,
            onInspect: () => inspected++,
          ),
        ),
      );
      final target = find.byKey(const ValueKey('select-inspect-only'));
      final node = tester.getSemantics(target);
      final data = node.getSemanticsData();
      expect(data.label, '8 of Clubs, copy 2');
      expect(data.hasAction(SemanticsAction.tap), isTrue);
      expect(data.customSemanticsActionIds, hasLength(1));
      node.owner!.performAction(
        node.id,
        SemanticsAction.customAction,
        data.customSemanticsActionIds!.single,
      );
      expect(inspected, 1);
      await tester.tap(target);
      await tester.pump(const Duration(milliseconds: 350));
      expect(inspected, 1);
      await tester.tap(target);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(target);
      await tester.pumpAndSettle();
      expect(inspected, 2);
      final ink = tester.widget<InkWell>(
        find.descendant(of: target, matching: find.byType(InkWell)),
      );
      ink.focusNode!.requestFocus();
      await tester.pump();
      expect(ink.focusNode!.hasFocus, isTrue);
      await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
      expect(inspected, 3);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(inspected, 3);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      expect(inspected, 3);
      final magnify = find.byKey(const ValueKey('magnify-inspect-only'));
      final magnifyData = tester.getSemantics(magnify).getSemanticsData();
      expect(magnifyData.label, 'Inspect 8 of Clubs, copy 2');
      expect(magnifyData.flagsCollection.isFocused, isNot(Tristate.none));
      await tester.tap(magnify);
      expect(inspected, 4);
      expect(tester.takeException(), isNull);
      handle.dispose();
    },
  );

  for (final mode in ['full', 'compact', 'stack', 'thumbnail']) {
    testWidgets(
      'controller style changes art but preserves identity and focus in $mode cards',
      (tester) async {
        final semantics = tester.ensureSemantics();
        var selected = 0;
        var inspected = 0;
        FocusNode? originalFocus;
        for (final style in CardStyle.values) {
          await tester.pumpWidget(
            shell(
              CgmsCard(
                card: VisibleCard(
                  id: 'captured-diamonds-8-copy-2',
                  rank: '8',
                  suit: CardSuit.diamonds,
                  copy: 2,
                  status: 'Available from round 4',
                  style: style,
                ),
                compact: mode == 'compact' || mode == 'stack',
                stackHeader: mode == 'stack',
                thumbnail: mode == 'thumbnail',
                selected: true,
                onTap: () => selected++,
                onInspect: () => inspected++,
              ),
            ),
          );
          await tester.pumpAndSettle();
          final prefix = style == CardStyle.firstLight ? '' : '${style.slug}/';
          expect(
            (tester.widget<Image>(find.byType(Image)).image as AssetImage)
                .keyName,
            'packages/cgms_ui/assets/art/${prefix}diamonds-8.png',
          );
          final target = find.byKey(
            const ValueKey('select-captured-diamonds-8-copy-2'),
          );
          final data = tester.getSemantics(target).getSemanticsData();
          expect(data.label, '8 of Diamonds, copy 2');
          expect(data.value, 'Available from round 4');
          expect(data.flagsCollection.isSelected, Tristate.isTrue);
          final ink = tester.widget<InkWell>(
            find.descendant(of: target, matching: find.byType(InkWell)),
          );
          if (originalFocus == null) {
            originalFocus = ink.focusNode!;
            originalFocus.requestFocus();
            await tester.pump();
          } else {
            expect(ink.focusNode, same(originalFocus));
            expect(originalFocus.hasFocus, isTrue);
          }
          await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
          expect(
            selected,
            0,
            reason:
                'Inspection and a style update must never select or submit.',
          );
          expect(inspected, style.index + 1);
          expect(tester.takeException(), isNull);
        }
        await tester.sendKeyEvent(LogicalKeyboardKey.enter);
        expect(selected, 1);
        semantics.dispose();
      },
    );
  }

  for (final compact in [false, true]) {
    testWidgets('inspection shortcuts do not select a compact=$compact card', (
      tester,
    ) async {
      var selected = 0;
      var inspected = 0;
      await tester.pumpWidget(
        shell(
          CgmsCard(
            card: const VisibleCard(
              id: 'inspection-shortcuts',
              rank: '8',
              suit: CardSuit.clubs,
              copy: 2,
            ),
            compact: compact,
            showArt: false,
            onTap: () => selected++,
            onInspect: () => inspected++,
          ),
        ),
      );
      final target = find.byKey(const ValueKey('select-inspection-shortcuts'));
      tester
          .widget<InkWell>(
            find.descendant(of: target, matching: find.byType(InkWell)),
          )
          .focusNode!
          .requestFocus();
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.keyI);
      expect(inspected, 1);
      expect(selected, 0);
      await tester.longPress(target);
      expect(inspected, 1);
      expect(selected, 0);
      await tester.tap(target);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(target);
      await tester.pumpAndSettle();
      expect(inspected, 2);
      expect(selected, 0);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(selected, 1);
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets(
    'concealed card requests only shared back and generic semantics',
    (tester) async {
      final handle = tester.ensureSemantics();
      await tester.pumpWidget(shell(const ConcealedCard()));
      final image = tester.widget<Image>(find.byType(Image));
      expect(
        (image.image as AssetImage).keyName,
        'packages/cgms_ui/assets/back.png',
      );
      expect(find.bySemanticsLabel('Concealed card'), findsOneWidget);
      expect(find.textContaining('copy'), findsNothing);
      expect(find.byType(CgmsCard), findsNothing);
      handle.dispose();
    },
  );

  testWidgets(
    'selection, pointer, keyboard and inspection remain independent',
    (tester) async {
      final handle = tester.ensureSemantics();
      var selected = 0;
      var inspected = 0;
      const card = VisibleCard(
        id: 'physical-clubs-8-copy-1',
        rank: '8',
        suit: CardSuit.clubs,
        copy: 1,
      );
      await tester.pumpWidget(
        shell(
          CgmsCard(
            card: card,
            selected: true,
            showArt: false,
            onTap: () => selected++,
            onInspect: () => inspected++,
          ),
        ),
      );
      expect(find.byIcon(Icons.check), findsOneWidget);
      final selection = find.byKey(
        const ValueKey('select-physical-clubs-8-copy-1'),
      );
      expect(
        tester.getSemantics(selection).flagsCollection.isSelected,
        Tristate.isTrue,
      );
      await tester.tap(selection);
      await tester.pump(const Duration(milliseconds: 350));
      expect(selected, 1);
      await tester.tap(find.byTooltip('Magnify card'));
      expect(inspected, 1);
      expect(selected, 1);
      tester
          .widget<InkWell>(
            find.descendant(of: selection, matching: find.byType(InkWell)),
          )
          .focusNode!
          .requestFocus();
      await tester.pump();
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(selected, 2);
      handle.dispose();
    },
  );

  testWidgets('missing art preserves visible labels without an exception', (
    tester,
  ) async {
    const card = VisibleCard(
      id: 'missing-asset-copy-1',
      rank: '8',
      suit: CardSuit.clubs,
      copy: 1,
      style: CardStyle.drypoint,
    );
    await tester.pumpWidget(
      shell(
        DefaultAssetBundle(
          bundle: _MissingArtBundle(),
          child: const CgmsCard(card: card),
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.text('8'), findsOneWidget);
    expect(find.text('Copy 1'), findsNothing);
    final identity = tester.widget<Semantics>(
      find.byKey(const ValueKey('select-missing-asset-copy-1')),
    );
    expect(identity.properties.label, '8 of Clubs, copy 1');
    expect(
      tester
          .getSize(find.byKey(const ValueKey('card-art-missing-asset-copy-1')))
          .aspectRatio,
      closeTo(card.artworkAspectRatio, .001),
    );
    expect(tester.takeException(), isNull);
  });

  testWidgets(
    'large text does not overflow card labels and inspection target',
    (tester) async {
      await tester.pumpWidget(
        shell(
          MediaQuery(
            data: const MediaQueryData(textScaler: TextScaler.linear(2)),
            child: CgmsCard(
              width: 100,
              card: const VisibleCard(
                id: 'diamonds-10-copy-2',
                rank: '10',
                suit: CardSuit.diamonds,
                copy: 2,
              ),
              selected: true,
              showArt: false,
              onInspect: () {},
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    },
  );
  testWidgets(
    'compact cards preserve full identity without a wrapping suit caption',
    (tester) async {
      final handle = tester.ensureSemantics();
      await tester.pumpWidget(
        shell(
          const CgmsCard(
            width: 62,
            showArt: false,
            card: VisibleCard(
              id: 'compact-diamonds',
              rank: '10',
              suit: CardSuit.diamonds,
              copy: 2,
            ),
          ),
        ),
      );
      expect(find.text('Diamonds'), findsNothing);
      expect(find.text('Copy 2'), findsNothing);
      expect(find.bySemanticsLabel('10 of Diamonds, copy 2'), findsOneWidget);
      expect(tester.takeException(), isNull);
      handle.dispose();
    },
  );

  testWidgets('compact portrait keeps art and reduces table card footprint', (
    tester,
  ) async {
    const card = VisibleCard(
      id: 'compact-geometry',
      rank: '10',
      suit: CardSuit.clubs,
      copy: 1,
    );
    await tester.pumpWidget(shell(const CgmsCard(card: card, width: 56)));
    final regularSize = tester.getSize(find.byType(CgmsCard));
    await tester.pumpWidget(
      shell(const CgmsCard(card: card, width: 56, compact: true)),
    );
    final compactSize = tester.getSize(find.byType(CgmsCard));
    expect(compactSize.width, 64);
    expect(compactSize.height, lessThanOrEqualTo(110));
    expect(
      compactSize.height,
      closeTo(56 / card.artworkAspectRatio + 26 + 8, .001),
    );
    expect(compactSize.height, lessThan(regularSize.height));
    expect(find.byType(Image), findsOneWidget);
    expect(find.byType(IconButton), findsNothing);
    final art = tester.widget<Image>(find.byType(Image));
    expect(
      (art.image as AssetImage).keyName,
      'packages/cgms_ui/assets/art/clubs-10.png',
    );
    expect(tester.takeException(), isNull);
  });

  for (final width in [56.0, 100.0, 144.0]) {
    for (final stackHeader in [false, true]) {
      testWidgets(
        'compact painting keeps its native ratio at $width with stackHeader=$stackHeader',
        (tester) async {
          for (final scale in [1.0, 2.0]) {
            for (final inspectable in [false, true]) {
              await tester.pumpWidget(
                shell(
                  Directionality(
                    textDirection: TextDirection.rtl,
                    child: MediaQuery(
                      data: MediaQueryData(
                        textScaler: TextScaler.linear(scale),
                      ),
                      child: CgmsCard(
                        width: width,
                        compact: true,
                        stackHeader: stackHeader,
                        showArt: false,
                        selected: true,
                        card: const VisibleCard(
                          id: 'phone-proportions',
                          rank: '10',
                          suit: CardSuit.diamonds,
                          copy: 2,
                          status: 'Available from round 4',
                        ),
                        onInspect: inspectable ? () {} : null,
                      ),
                    ),
                  ),
                ),
              );
              final face = tester.getRect(
                find
                    .descendant(
                      of: find.byType(CgmsCard),
                      matching: find.byType(Material),
                    )
                    .first,
              );
              final art = tester.getRect(
                find.byKey(const ValueKey('card-art-phone-proportions')),
              );
              final header = tester.getRect(
                find.byKey(const ValueKey('card-header-phone-proportions')),
              );
              expect(art.width / art.height, closeTo(282 / 357, .001));
              final rank = tester.getRect(find.text('10'));
              expect(header.contains(rank.topLeft), isTrue);
              expect(header.contains(rank.bottomRight), isTrue);
              expect(header.bottom, closeTo(art.top, .001));
              expect(find.text('Copy 2'), findsNothing);
              expect(
                tester.getRect(find.byIcon(Icons.check)).bottom,
                lessThanOrEqualTo(art.top),
              );
              expect(
                tester.getRect(find.text('Available from round 4')).top,
                greaterThanOrEqualTo(face.bottom),
              );
              if (inspectable) {
                final magnify = tester.getRect(
                  find.byKey(const ValueKey('magnify-phone-proportions')),
                );
                expect(magnify.width, greaterThanOrEqualTo(48));
                expect(magnify.height, greaterThanOrEqualTo(48));
                expect(magnify.right, closeTo(art.right, .001));
                expect(magnify.bottom, closeTo(art.bottom, .001));
              }
              expect(tester.takeException(), isNull);
            }
          }
        },
      );
    }
  }

  testWidgets('stacked card status remains visible between physical cards', (
    tester,
  ) async {
    await tester.pumpWidget(
      shell(
        MediaQuery(
          data: const MediaQueryData(textScaler: TextScaler.linear(2)),
          child: CgmsSuitStacks(
            cards: const [
              VisibleCard(
                id: 'used-stack-card',
                rank: '10',
                suit: CardSuit.clubs,
                copy: 1,
                status: 'Used this turn',
              ),
              VisibleCard(
                id: 'next-stack-card',
                rank: '10',
                suit: CardSuit.clubs,
                copy: 2,
              ),
            ],
            showArt: false,
            onTap: (_) {},
          ),
        ),
      ),
    );
    final usedFace = tester.getRect(
      find.byKey(const ValueKey('select-used-stack-card')),
    );
    final status = tester.getRect(find.text('Used this turn'));
    final nextFace = tester.getRect(
      find.byKey(const ValueKey('select-next-stack-card')),
    );
    final usedArt = tester.getRect(
      find.byKey(const ValueKey('card-art-used-stack-card')),
    );
    expect(usedArt.width / usedArt.height, closeTo(281 / 359, .001));
    expect(status.top, greaterThanOrEqualTo(usedFace.bottom));
    expect(status.bottom, lessThanOrEqualTo(nextFace.top));
    expect(tester.takeException(), isNull);
  });

  for (final scale in [1.0, 2.0]) {
    testWidgets(
      'compact rank 10, semantic copy and status survive $scale text scale',
      (tester) async {
        final semantics = tester.ensureSemantics();
        await tester.pumpWidget(
          shell(
            MediaQuery(
              data: MediaQueryData(textScaler: TextScaler.linear(scale)),
              child: const CgmsCard(
                width: 56,
                compact: true,
                showArt: false,
                selected: true,
                card: VisibleCard(
                  id: 'compact-copy-2',
                  rank: '10',
                  suit: CardSuit.diamonds,
                  copy: 2,
                  status: 'Used this turn',
                ),
              ),
            ),
          ),
        );
        final cardRect = tester.getRect(find.byType(CgmsCard));
        final rankRect = tester.getRect(find.text('10'));
        expect(find.text('Copy 2'), findsNothing);
        expect(tester.widget<Text>(find.text('10')).maxLines, 1);
        expect(rankRect.height, lessThanOrEqualTo(22 * scale));
        expect(cardRect.contains(rankRect.topLeft), isTrue);
        expect(cardRect.contains(rankRect.bottomRight), isTrue);
        final artRect = tester.getRect(
          find.byKey(const ValueKey('card-art-compact-copy-2')),
        );
        expect(rankRect.bottom, lessThanOrEqualTo(artRect.top));
        expect(find.byType(Image), findsNothing);
        expect(find.text('Used this turn'), findsOneWidget);
        final selection = find.byKey(const ValueKey('select-compact-copy-2'));
        final data = tester.getSemantics(selection).getSemanticsData();
        expect(data.label, '10 of Diamonds, copy 2');
        expect(data.value, 'Used this turn');
        expect(data.flagsCollection.isSelected, Tristate.isTrue);
        expect(tester.takeException(), isNull);
        semantics.dispose();
      },
    );
  }

  testWidgets(
    'compact card has independent pointer, focus and inspect actions',
    (tester) async {
      var selected = 0;
      var inspected = 0;
      await tester.pumpWidget(
        shell(
          CgmsCard(
            width: 56,
            compact: true,
            showArt: false,
            selected: true,
            card: const VisibleCard(
              id: 'compact-actions',
              rank: '10',
              suit: CardSuit.diamonds,
              copy: 2,
            ),
            onTap: () => selected++,
            onInspect: () => inspected++,
          ),
        ),
      );
      final selection = find.byKey(const ValueKey('select-compact-actions'));
      final targetSize = tester.getSize(selection);
      expect(targetSize.width, greaterThanOrEqualTo(48));
      expect(targetSize.height, greaterThanOrEqualTo(48));
      expect(
        tester
            .getRect(find.byKey(const ValueKey('magnify-compact-actions')))
            .contains(tester.getCenter(selection)),
        isFalse,
      );
      await tester.tap(selection);
      await tester.pump(const Duration(milliseconds: 350));
      expect(selected, 1);
      final ink = tester.widget<InkWell>(
        find.descendant(of: selection, matching: find.byType(InkWell)),
      );
      ink.focusNode!.requestFocus();
      await tester.pump();
      final focusRing = tester.widget<DecoratedBox>(
        find
            .descendant(
              of: find.byType(CgmsCard),
              matching: find.byType(DecoratedBox),
            )
            .first,
      );
      expect(
        ((focusRing.decoration as BoxDecoration).border as Border).top.color,
        CgmsColors.lime,
      );
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      expect(selected, 2);
      await tester.sendKeyEvent(LogicalKeyboardKey.space);
      expect(selected, 3);
      await tester.tap(selection);
      await tester.pump(const Duration(milliseconds: 50));
      await tester.tap(selection);
      await tester.pumpAndSettle();
      expect(inspected, 1);
      expect(selected, 3);
      await tester.tap(find.byTooltip('Magnify card'));
      expect(inspected, 2);
      expect(selected, 3);
      expect(tester.takeException(), isNull);
    },
  );
}

class _MissingArtBundle extends CachingAssetBundle {
  @override
  Future<ByteData> load(String key) {
    if (key.contains('/art/')) {
      throw FlutterError('Simulated missing artwork');
    }
    return rootBundle.load(key);
  }
}

double _contrast(Color first, Color second) {
  final firstLuminance = first.computeLuminance();
  final secondLuminance = second.computeLuminance();
  final lighter = firstLuminance > secondLuminance
      ? firstLuminance
      : secondLuminance;
  final darker = firstLuminance > secondLuminance
      ? secondLuminance
      : firstLuminance;
  return (lighter + 0.05) / (darker + 0.05);
}
