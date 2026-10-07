import 'dart:async';

import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/semantics.dart';

import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

/// The complete native painting with live identity above it.
class CgmsCard extends StatefulWidget {
  const CgmsCard({
    super.key,
    required this.card,
    this.width = 100,
    this.compact = false,
    this.stackHeader = false,
    this.thumbnail = false,
    this.boardPreview = false,
    this.handStackPreview = false,
    this.semanticsSortKey,
    this.selected = false,
    this.showArt = true,
    this.onTap,
    this.onInspect,
    this.onActivate,
    this.selectionRole,
  });

  final VisibleCard card;
  final double width;

  /// Smaller live identity and spacing; the painting keeps its native ratio.
  final bool compact;

  /// Retained for collection callers; every card now exposes its complete face.
  final bool stackHeader;

  /// The smallest live identity for board previews, without covering artwork.
  final bool thumbnail;

  /// Compact reference-board face. Inspection remains available by double tap,
  /// keyboard I and the semantic magnify action; full inspectors use the normal
  /// face with its independent magnification button.
  final bool boardPreview;

  /// Uniform 4:5 hand-stack envelope containing the complete native painting.
  /// Identity, copy, selection and focus remain visible on the leading strip.
  /// Other card/inspector modes retain their native painting dimensions.
  final bool handStackPreview;

  /// Explicit collection order keeps overlapping, clipped cards stable in the
  /// accessibility tree while their scroll positions change.
  final SemanticsSortKey? semanticsSortKey;
  final bool selected;
  final bool showArt;
  final VoidCallback? onTap;
  final VoidCallback? onInspect;

  /// Double tap / Enter activates a contextual card flow; Space still selects.
  /// Omitted by compatibility surfaces whose double tap continues to inspect.
  final VoidCallback? onActivate;
  final TableCardRole? selectionRole;

  @override
  State<CgmsCard> createState() => _CgmsCardState();
}

class _CgmsCardState extends State<CgmsCard> {
  final _focus = FocusNode();
  bool _focused = false;
  Timer? _semanticTapTimer;

  @override
  void initState() {
    super.initState();
    _focus.onKeyEvent = (node, event) {
      if (widget.onInspect != null &&
          event is KeyDownEvent &&
          event.logicalKey == LogicalKeyboardKey.keyI) {
        _inspect();
        return KeyEventResult.handled;
      }
      if (widget.onActivate != null &&
          event is KeyDownEvent &&
          event.logicalKey == LogicalKeyboardKey.enter) {
        _activate();
        return KeyEventResult.handled;
      }
      return KeyEventResult.ignored;
    };
  }

  @override
  void didUpdateWidget(CgmsCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.card.id != widget.card.id) _cancelSemanticTap();
  }

  void _cancelSemanticTap() {
    _semanticTapTimer?.cancel();
    _semanticTapTimer = null;
  }

  void _inspect() {
    _cancelSemanticTap();
    widget.onInspect?.call();
  }

  void _activate() {
    _cancelSemanticTap();
    widget.onActivate?.call();
  }

  VoidCallback? get _doubleTap => widget.onActivate != null
      ? _activate
      : widget.onInspect != null
      ? _inspect
      : null;

  Color get _selectionColor => switch (widget.selectionRole) {
    TableCardRole.target => CgmsColors.action,
    TableCardRole.payment => CgmsColors.diamond,
    TableCardRole.modifier => CgmsColors.spade,
    _ => CgmsColors.cyan,
  };

  IconData get _selectionIcon => switch (widget.selectionRole) {
    TableCardRole.target => Icons.gps_fixed,
    TableCardRole.payment => Icons.payments_outlined,
    TableCardRole.modifier => Icons.add_circle_outline,
    _ => Icons.check,
  };

  String? _semanticValue(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    final role = switch (widget.selectionRole) {
      TableCardRole.source => strings.cardRoleSource,
      TableCardRole.target => strings.cardRoleTarget,
      TableCardRole.payment => strings.cardRolePayment,
      TableCardRole.modifier => strings.cardRoleModifier,
      null => '',
    };
    final value = [
      widget.card.status,
      role,
    ].where((s) => s.isNotEmpty).join(' · ');
    return value.isEmpty ? null : value;
  }

  Map<CustomSemanticsAction, VoidCallback>? _semanticActions(
    BuildContext context,
  ) {
    final strings = CgmsLocalizations.of(context);
    final actions = <CustomSemanticsAction, VoidCallback>{
      if (widget.onInspect != null)
        CustomSemanticsAction(label: strings.magnifyCard): _inspect,
      if (widget.onActivate != null)
        CustomSemanticsAction(label: strings.activateCard): _activate,
    };
    return actions.isEmpty ? null : actions;
  }

  void _semanticTap() {
    if (_doubleTap == null) {
      _cancelSemanticTap();
      widget.onTap?.call();
      return;
    }
    if (_semanticTapTimer != null) {
      _doubleTap!();
      return;
    }
    // Flutter web replaces raw pointer events on tappable accessibility nodes
    // with semantic taps. Preserve the same single/double-tap split here.
    _semanticTapTimer = Timer(kDoubleTapTimeout, () {
      _semanticTapTimer = null;
      if (mounted) widget.onTap?.call();
    });
  }

  @override
  void dispose() {
    _cancelSemanticTap();
    _focus.dispose();
    super.dispose();
  }

  static const _rankStyle = TextStyle(
    fontFamily: 'Barlow Condensed',
    package: 'cgms_ui',
    fontSize: 20,
    fontWeight: FontWeight.w700,
    height: 1,
    color: CgmsColors.text,
  );

  static double _faceWidth(
    BuildContext context,
    VisibleCard card,
    double requestedWidth, {
    double rankSize = 20,
    double suitSize = 12,
  }) {
    final painter = TextPainter(
      text: TextSpan(
        text: card.rank,
        style: DefaultTextStyle.of(
          context,
        ).style.merge(_rankStyle.copyWith(fontSize: rankSize)),
      ),
      textDirection: Directionality.of(context),
      textScaler: MediaQuery.textScalerOf(context),
      maxLines: 1,
    )..layout();
    // Rank + suit + reserved selection mark + gaps and header padding.
    final identityWidth = painter.width + suitSize + 24;
    painter.dispose();
    return [
      requestedWidth,
      48.0,
      identityWidth,
    ].reduce((first, second) => first > second ? first : second);
  }

  @override
  Widget build(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    final card = widget.card;
    final identity = strings.cardIdentity(
      card.rank,
      _suitName(strings, card.suit),
      card.copy,
    );
    if (widget.boardPreview || widget.handStackPreview) {
      return _boardFace(context, identity);
    }
    final rankSize = widget.thumbnail
        ? 17.0
        : widget.compact
        ? 20.0
        : 30.0;
    final suitSize = widget.thumbnail || widget.compact ? 12.0 : 16.0;
    final inset = widget.thumbnail
        ? 1.0
        : widget.compact
        ? 4.0
        : 6.0;
    final width = _faceWidth(
      context,
      card,
      widget.onInspect != null && widget.width < 100 ? 100 : widget.width,
      rankSize: rankSize,
      suitSize: suitSize,
    );
    final face = DecoratedBox(
      decoration: BoxDecoration(
        border: Border.all(
          color: _focused
              ? CgmsColors.lime
              : widget.selected
              ? _selectionColor
              : CgmsColors.border,
          width: _focused || widget.selected ? 3 : 1,
          strokeAlign: BorderSide.strokeAlignOutside,
        ),
      ),
      child: Material(
        color: CgmsColors.surface,
        child: Stack(
          children: [
            Semantics(
              key: ValueKey('select-${card.id}'),
              sortKey: widget.semanticsSortKey,
              container: true,
              label: identity,
              value: _semanticValue(context),
              selected: widget.selected,
              button:
                  widget.onTap != null ||
                  widget.onInspect != null ||
                  widget.onActivate != null,
              onTap:
                  widget.onTap != null ||
                      widget.onInspect != null ||
                      widget.onActivate != null
                  ? _semanticTap
                  : null,
              customSemanticsActions: _semanticActions(context),
              hint: widget.onActivate != null
                  ? strings.cardGestureHint
                  : widget.onInspect == null
                  ? null
                  : strings.actionBoardInspectHint,
              child: InkWell(
                focusNode: _focus,
                onFocusChange: (value) => setState(() => _focused = value),
                onTap: widget.onTap,
                onDoubleTap: _doubleTap,
                // Holding a selectable card must not become a delayed tap.
                onLongPress: widget.onTap == null ? null : () {},
                excludeFromSemantics: true,
                child: ExcludeSemantics(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Padding(
                        key: ValueKey('card-header-${card.id}'),
                        padding: const EdgeInsets.symmetric(
                          horizontal: 4,
                          vertical: 3,
                        ),
                        child: Row(
                          children: [
                            Text(
                              card.rank,
                              maxLines: 1,
                              style: _rankStyle.copyWith(fontSize: rankSize),
                            ),
                            const SizedBox(width: 2),
                            CustomPaint(
                              size: Size(suitSize, suitSize + 2),
                              painter: _SuitPainter(
                                card.suit,
                                _suitColor(card.suit),
                              ),
                            ),
                            const Spacer(),
                            const SizedBox(width: 2),
                            SizedBox(
                              width: 12,
                              height: 14,
                              child: widget.selected
                                  ? Icon(
                                      _selectionIcon,
                                      size: 12,
                                      color: _selectionColor,
                                    )
                                  : null,
                            ),
                          ],
                        ),
                      ),
                      AspectRatio(
                        key: ValueKey('card-art-${card.id}'),
                        aspectRatio: card.artworkAspectRatio,
                        child: widget.showArt
                            ? Image.asset(
                                card.artworkAsset,
                                package: 'cgms_ui',
                                fit: BoxFit.contain,
                                excludeFromSemantics: true,
                                errorBuilder: (_, _, _) =>
                                    _CardArtFallback(suit: card.suit),
                              )
                            : _CardArtFallback(suit: card.suit),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            if (widget.onInspect != null)
              Positioned(
                right: 0,
                bottom: 0,
                child: Tooltip(
                  message: strings.magnifyCard,
                  excludeFromSemantics: true,
                  child: IconButton(
                    key: ValueKey('magnify-${card.id}'),
                    onPressed: _inspect,
                    constraints: const BoxConstraints(
                      minWidth: 48,
                      minHeight: 48,
                    ),
                    icon: Icon(
                      Icons.zoom_in_outlined,
                      semanticLabel: strings.inspectCard(identity),
                      size: 20,
                      color: CgmsColors.text,
                      shadows: const [
                        Shadow(color: CgmsColors.canvas, blurRadius: 3),
                      ],
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
    return SizedBox(
      width: width + inset * 2,
      child: Padding(
        padding: EdgeInsets.all(inset),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            face,
            if (card.displayStatus.isNotEmpty)
              ExcludeSemantics(
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 4,
                    vertical: 3,
                  ),
                  child: Text(
                    card.displayStatus,
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      color: CgmsColors.muted,
                      fontSize: 12,
                      height: 1.2,
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _boardFace(BuildContext context, String identity) {
    final card = widget.card;
    final interactive =
        widget.onTap != null ||
        widget.onInspect != null ||
        widget.onActivate != null;
    final width = widget.width.clamp(48.0, double.infinity);
    final identityScale = MediaQuery.textScalerOf(context).scale(14) / 14;
    return SizedBox(
      width: width,
      child: Semantics(
        key: ValueKey('select-${card.id}'),
        sortKey: widget.semanticsSortKey,
        container: true,
        label: identity,
        value: _semanticValue(context),
        selected: widget.selected,
        button: interactive,
        onTap: interactive ? _semanticTap : null,
        customSemanticsActions: _semanticActions(context),
        hint: widget.onActivate == null
            ? null
            : CgmsLocalizations.of(context).cardGestureHint,
        child: Material(
          color: CgmsColors.surface,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(4),
            side: BorderSide(
              color: _focused
                  ? CgmsColors.lime
                  : widget.selected
                  ? _selectionColor
                  : CgmsColors.border,
              width: _focused || widget.selected ? 2 : 1,
            ),
          ),
          clipBehavior: Clip.antiAlias,
          child: InkWell(
            focusNode: _focus,
            onFocusChange: (value) => setState(() => _focused = value),
            onTap: widget.onTap,
            onDoubleTap: _doubleTap,
            onLongPress: widget.onTap == null ? null : () {},
            excludeFromSemantics: true,
            child: ExcludeSemantics(
              child: AspectRatio(
                key: ValueKey('card-frame-${card.id}'),
                aspectRatio: widget.handStackPreview
                    ? 4 / 5
                    : card.artworkAspectRatio,
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    Center(
                      child: AspectRatio(
                        key: ValueKey('card-art-${card.id}'),
                        aspectRatio: card.artworkAspectRatio,
                        child: widget.showArt
                            ? Image.asset(
                                card.artworkAsset,
                                package: 'cgms_ui',
                                fit: BoxFit.contain,
                                errorBuilder: (_, _, _) =>
                                    _CardArtFallback(suit: card.suit),
                              )
                            : _CardArtFallback(suit: card.suit),
                      ),
                    ),
                    Align(
                      alignment: AlignmentDirectional.topStart,
                      child: Container(
                        key: ValueKey('card-header-${card.id}'),
                        padding: const EdgeInsets.symmetric(horizontal: 2),
                        color: CgmsColors.canvas.withValues(alpha: .85),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              card.rank,
                              style: _rankStyle.copyWith(fontSize: 16),
                            ),
                            const SizedBox(width: 1),
                            CustomPaint(
                              size: Size(
                                11 * identityScale,
                                13 * identityScale,
                              ),
                              painter: _SuitPainter(
                                card.suit,
                                _suitColor(card.suit),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    Align(
                      alignment: widget.handStackPreview
                          ? AlignmentDirectional.bottomStart
                          : AlignmentDirectional.bottomEnd,
                      child: Container(
                        key: ValueKey('card-copy-${card.id}'),
                        padding: const EdgeInsets.symmetric(horizontal: 2),
                        color: CgmsColors.canvas.withValues(alpha: .85),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(
                              '${card.copy}',
                              style: const TextStyle(fontSize: 9, height: 1),
                            ),
                            if (widget.handStackPreview && widget.selected)
                              Icon(
                                _selectionIcon,
                                key: ValueKey('card-selected-${card.id}'),
                                size: 16,
                                color: _selectionColor,
                              ),
                          ],
                        ),
                      ),
                    ),
                    if (widget.selected && !widget.handStackPreview)
                      Align(
                        alignment: Alignment.topRight,
                        child: Icon(
                          _selectionIcon,
                          size: 16,
                          color: _selectionColor,
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// Full visible cards in side-by-side suit columns, preserving supplied order.
class CgmsSuitStacks extends StatelessWidget {
  const CgmsSuitStacks({
    super.key,
    required this.cards,
    this.onTap,
    this.onInspect,
    this.selectedCardId,
    this.cardWidth = 48,
    this.showArt = true,
    this.cardKeyPrefix = 'public-card',
    this.expandSelectedCard = true,
  });

  final List<VisibleCard> cards;
  final ValueChanged<VisibleCard>? onTap;
  final ValueChanged<VisibleCard>? onInspect;
  final String? selectedCardId;
  final double cardWidth;
  final bool showArt;

  /// Namespaces physical card keys when several collections share a view.
  final String cardKeyPrefix;

  /// Retained for compatibility; all paintings remain fully exposed.
  final bool expandSelectedCard;

  @override
  Widget build(BuildContext context) => Wrap(
    spacing: 2,
    runSpacing: 8,
    children: [
      for (final suit in CardSuit.values)
        if (cards.any((card) => card.suit == suit))
          _suitStack(
            context,
            cards.where((card) => card.suit == suit).toList(),
          ),
    ],
  );

  Widget _suitStack(BuildContext context, List<VisibleCard> group) {
    final width = group
        .map((card) => _CgmsCardState._faceWidth(context, card, cardWidth))
        .reduce((a, b) => a > b ? a : b);
    return Column(
      key: ValueKey('suit-stack-${group.first.suit.name}'),
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        for (final card in group)
          CgmsCard(
            key: ValueKey('$cardKeyPrefix-${card.id}'),
            card: card,
            width: width,
            compact: true,
            stackHeader: true,
            showArt: showArt,
            selected: card.id == selectedCardId,
            onTap: onTap == null ? null : () => onTap!(card),
            onInspect: onInspect == null ? null : () => onInspect!(card),
          ),
      ],
    );
  }
}

/// A concealed object with one shared asset and no identity-bearing inputs.
class ConcealedCard extends StatelessWidget {
  const ConcealedCard({super.key, this.width = 60, this.showArt = true});

  final double width;
  final bool showArt;

  @override
  Widget build(BuildContext context) => Semantics(
    label: CgmsLocalizations.of(context).concealedCard,
    image: true,
    child: ExcludeSemantics(
      child: Container(
        width: width,
        height: width * 1.4,
        clipBehavior: Clip.antiAlias,
        decoration: BoxDecoration(
          color: CgmsColors.surface,
          border: Border.all(color: CgmsColors.border),
          borderRadius: BorderRadius.circular(5),
        ),
        child: showArt
            ? Image.asset(
                'assets/back.png',
                package: 'cgms_ui',
                fit: BoxFit.cover,
                excludeFromSemantics: true,
                errorBuilder: (_, _, _) => const _BackFallback(),
              )
            : const _BackFallback(),
      ),
    ),
  );
}

/// Opaque reading surface shared by setup, decisions and table territories.
/// The optional accent is directional; it carries no information without text.
class CgmsPanel extends StatelessWidget {
  const CgmsPanel({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(CgmsSpacing.panelInset),
    this.accent,
  });
  final Widget child;
  final EdgeInsetsGeometry padding;
  final Color? accent;

  @override
  Widget build(BuildContext context) => Material(
    color: CgmsColors.surface,
    shape: RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(CgmsShape.panel),
      side: const BorderSide(color: CgmsColors.divider),
    ),
    child: DecoratedBox(
      decoration: BoxDecoration(
        border: accent == null
            ? null
            : BorderDirectional(start: BorderSide(color: accent!, width: 3)),
      ),
      child: Padding(padding: padding, child: child),
    ),
  );
}

class CgmsNotice extends StatelessWidget {
  const CgmsNotice({
    super.key,
    required this.title,
    required this.message,
    this.icon = Icons.info_outline,
    this.color,
  });

  final String title;
  final String message;
  final IconData icon;
  final Color? color;

  @override
  Widget build(BuildContext context) => CgmsPanel(
    accent: color ?? CgmsColors.cyan,
    child: Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        ExcludeSemantics(
          child: Icon(icon, color: color ?? CgmsColors.cyan, size: 24),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: Theme.of(context).textTheme.titleMedium),
              const SizedBox(height: 4),
              Text(
                message,
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(color: CgmsColors.muted),
              ),
            ],
          ),
        ),
      ],
    ),
  );
}

class CgmsSectionTitle extends StatelessWidget {
  const CgmsSectionTitle({
    super.key,
    required this.title,
    this.eyebrow,
    this.trailing,
  });

  final String title;
  final String? eyebrow;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: LayoutBuilder(
      builder: (context, constraints) {
        final titleWidget = Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (eyebrow != null) ...[
              Text(
                eyebrow!,
                style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: CgmsColors.muted,
                  letterSpacing: 0.8,
                ),
              ),
              const SizedBox(height: 4),
            ],
            Semantics(
              header: true,
              child: Text(
                title,
                style: Theme.of(context).textTheme.headlineMedium,
              ),
            ),
          ],
        );
        if (trailing == null) return titleWidget;
        if (constraints.maxWidth < 480 ||
            MediaQuery.textScalerOf(context).scale(16) > 24) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [titleWidget, const SizedBox(height: 8), trailing!],
          );
        }
        return Row(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Expanded(child: titleWidget),
            const SizedBox(width: 16),
            trailing!,
          ],
        );
      },
    ),
  );
}

class _CardArtFallback extends StatelessWidget {
  const _CardArtFallback({required this.suit});
  final CardSuit suit;

  @override
  Widget build(BuildContext context) => ColoredBox(
    color: CgmsColors.raised,
    child: Center(
      child: CustomPaint(
        size: const Size(36, 42),
        painter: _SuitPainter(suit, _suitColor(suit)),
      ),
    ),
  );
}

class _BackFallback extends StatelessWidget {
  const _BackFallback();

  @override
  Widget build(BuildContext context) => const Center(
    child: Icon(Icons.diamond_outlined, color: CgmsColors.border),
  );
}

String _suitName(CgmsLocalizations strings, CardSuit suit) => switch (suit) {
  CardSuit.hearts => strings.hearts,
  CardSuit.diamonds => strings.diamonds,
  CardSuit.clubs => strings.clubs,
  CardSuit.spades => strings.spades,
};

Color _suitColor(CardSuit suit) => switch (suit) {
  CardSuit.hearts => CgmsColors.heart,
  CardSuit.diamonds => CgmsColors.diamond,
  CardSuit.clubs => CgmsColors.club,
  CardSuit.spades => CgmsColors.spade,
};

/// Original vector emblems avoid emoji variation and absent font glyphs.
class _SuitPainter extends CustomPainter {
  const _SuitPainter(this.suit, this.color);
  final CardSuit suit;
  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.save();
    canvas.scale(size.width / 100, size.height / 100);
    final paint = Paint()..color = color;
    final path = Path();
    switch (suit) {
      case CardSuit.hearts:
        path.moveTo(50, 94);
        path.cubicTo(-18, 49, 0, -11, 50, 24);
        path.cubicTo(100, -11, 118, 49, 50, 94);
        path.close();
      case CardSuit.diamonds:
        path.moveTo(50, 2);
        path.lineTo(96, 50);
        path.lineTo(50, 98);
        path.lineTo(4, 50);
        path.close();
      case CardSuit.clubs:
        canvas.drawCircle(const Offset(50, 26), 25, paint);
        canvas.drawCircle(const Offset(26, 56), 25, paint);
        canvas.drawCircle(const Offset(74, 56), 25, paint);
        path.moveTo(47, 48);
        path.cubicTo(43, 70, 40, 87, 31, 97);
        path.lineTo(69, 97);
        path.cubicTo(60, 87, 57, 70, 53, 48);
        path.close();
      case CardSuit.spades:
        path.moveTo(50, 2);
        path.cubicTo(-18, 47, 0, 98, 43, 73);
        path.cubicTo(42, 83, 36, 92, 30, 98);
        path.lineTo(70, 98);
        path.cubicTo(64, 92, 58, 83, 57, 73);
        path.cubicTo(100, 98, 118, 47, 50, 2);
        path.close();
    }
    canvas.drawPath(path, paint);
    canvas.restore();
  }

  @override
  bool shouldRepaint(_SuitPainter oldDelegate) =>
      oldDelegate.suit != suit || oldDelegate.color != color;
}
