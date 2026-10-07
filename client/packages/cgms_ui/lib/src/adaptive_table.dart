import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter/services.dart';

import 'components.dart';
import 'draw_feedback.dart';
import 'l10n/cgms_localizations.dart';
import 'models.dart';
import 'theme.dart';

/// A full-width public overview and adjacent private hand at every width.
/// Card-driven hosts omit [actions]; legacy hosts retain their action workspace.
class CgmsAdaptiveTable extends StatefulWidget {
  const CgmsAdaptiveTable({
    super.key,
    required this.data,
    this.actions,
    this.actionsIdentity,
    required this.onSelect,
    required this.onInspect,
    this.onCoup,
    this.onInspectCombo,
    this.onCloseActions,
    this.publicGroupLabels = const {},
    this.showDecisions = false,
    this.selectedCardIds = const {},
    this.draggableCardIds = const {},
    this.onCardDrop,
    this.onActivateCard,
    this.onActivateTarget,
    this.onTargetTap,
    this.onCancelSelection,
    this.canDrop,
    this.isSuggestedTarget,
    this.drawFeedback,
    this.dropPreview,
    this.dragIdentity,
    this.cardRoles = const {},
    this.showArt = true,
    this.header,
    this.headerTools,
    this.publicChooserTools,
    this.footer,
  });
  final TableViewData data;
  final Widget? actions;

  /// Stable identity of the current decision workspace, independent of updates
  /// and viewport size. Different decisions retain separate scroll positions.
  final Object? actionsIdentity;
  final ValueChanged<String> onSelect;
  final ValueChanged<VisibleCard> onInspect;
  final VoidCallback? onCoup;
  final ValueChanged<ComboView>? onInspectCombo;
  final VoidCallback? onCloseActions;
  final Map<String, String> publicGroupLabels;
  final bool showDecisions;
  final Set<String> selectedCardIds;

  /// Host-authorized candidates. Current own-card visibility is checked again
  /// when a drag starts and when it lands; the host decides whether the complete
  /// move can commit directly or requires a missing parameter.
  final Set<String> draggableCardIds;
  final void Function(TableCardDrag, TableDropTarget)? onCardDrop;

  /// Opens only the contextual flow for this authorized physical card.
  final ValueChanged<String>? onActivateCard;

  /// Activates an explicit series or formation, never a representative card.
  final ValueChanged<TableDropTarget>? onActivateTarget;
  final ValueChanged<TableDropTarget>? onTargetTap;
  final VoidCallback? onCancelSelection;
  final bool Function(TableCardDrag, TableDropTarget)? canDrop;

  /// Exact authorized destinations for the host's current incomplete choice.
  /// This only highlights the board; it cannot submit or choose a destination.
  final bool Function(TableDropTarget)? isSuggestedTarget;

  /// Presentation of a newly confirmed draw; this never requests a card draw.
  final TableDrawFeedback? drawFeedback;

  /// Authorized presentation hint for the exact bundle and hovered target.
  /// Hosts must not include concealed candidate faces or binding metadata.
  final String? Function(TableCardDrag, TableDropTarget)? dropPreview;
  final Object? dragIdentity;
  final Map<String, TableCardRole> cardRoles;
  final bool showArt;
  final Widget? header, headerTools, footer;

  /// Host-provided immediate card entry, e.g. authorized Coup queens, while
  /// the public-card chooser owns modal focus.
  final Widget? publicChooserTools;
  @override
  State<CgmsAdaptiveTable> createState() => _CgmsAdaptiveTableState();
}

class _CgmsAdaptiveTableState extends State<CgmsAdaptiveTable> {
  static const _cardWidth = 72.0;
  static const _publicSeatHeight = 180.0;
  final _actionToggleFocus = FocusNode(debugLabel: 'Choose an action');
  bool _actionsOpen = false;
  int _actionsGeneration = 0;
  final _actionsKey = GlobalKey();
  final _projection = ValueNotifier<TableViewData?>(null);
  final _chooserDragging = ValueNotifier(false);
  DialogRoute<void>? _pileRoute;
  int _projectionGeneration = 0;
  ({Size size, double textScale, TextDirection direction})? _layout;
  TableCardDrag? _activeDrag;
  Offset? _dragOrigin;
  bool _dragMoved = false;
  final _dragPreview = ValueNotifier<String?>(null);
  final _dragPosition = ValueNotifier(Offset.zero);
  TableDropTarget? _previewTarget;
  final _handScrollController = ScrollController(keepScrollOffset: false);
  final _handContentKey = GlobalKey();
  bool? _handCanScroll;
  bool _handScrollResetScheduled = false;
  int _handScrollGeneration = 0;

  @override
  void initState() {
    super.initState();
    _projection.value = widget.data;
    FocusManager.instance.addListener(_revealFocusedControl);
    HardwareKeyboard.instance.addHandler(_cancelDragOnEscape);
    if (widget.showDecisions) _revealActions(requiredByHost: true);
  }

  @override
  void didUpdateWidget(CgmsAdaptiveTable oldWidget) {
    super.didUpdateWidget(oldWidget);
    final generation = ++_projectionGeneration;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted && generation == _projectionGeneration) {
        _projection.value = widget.data;
        if (_activeDrag != null &&
            (!_validDrag(_activeDrag!) ||
                (_previewTarget != null &&
                    (!_validTarget(_previewTarget!) ||
                        !(widget.canDrop?.call(_activeDrag!, _previewTarget!) ??
                            true))))) {
          _dragPreview.value = null;
          _previewTarget = null;
        }
      }
    });
    if (!widget.showDecisions && oldWidget.showDecisions) _actionsGeneration++;
    if (widget.showDecisions && !oldWidget.showDecisions) {
      _revealActions(requiredByHost: true);
    }
  }

  @override
  void dispose() {
    HardwareKeyboard.instance.removeHandler(_cancelDragOnEscape);
    FocusManager.instance.removeListener(_revealFocusedControl);
    final route = _pileRoute;
    _pileRoute = null;
    if (route != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (route.isActive) route.navigator?.removeRoute(route);
      });
    }
    _actionToggleFocus.dispose();
    _projection.dispose();
    _chooserDragging.dispose();
    _dragPreview.dispose();
    _dragPosition.dispose();
    _handScrollController.dispose();
    super.dispose();
  }

  bool _handMetricsChanged(ScrollMetricsNotification notification) {
    if (notification.depth != 0 || notification.metrics.axis != Axis.vertical) {
      return false;
    }
    final canScroll =
        notification.metrics.maxScrollExtent >
        notification.metrics.minScrollExtent;
    final stoppedScrolling = _handCanScroll == true && !canScroll;
    _handCanScroll = canScroll;
    if (stoppedScrolling && !_handScrollResetScheduled) {
      _handScrollResetScheduled = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _handScrollResetScheduled = false;
        if (!mounted || _handCanScroll != false) return;
        // Flutter web retains native scroll compensation when the directional
        // scroll actions disappear. Renew only that viewport's semantics; the
        // GlobalKey moves its physical card content intact in the same frame.
        setState(() => _handScrollGeneration++);
      });
      WidgetsBinding.instance.ensureVisualUpdate();
    }
    return false;
  }

  bool _cancelDragOnEscape(KeyEvent event) {
    // A host-owned popup or utility can own focus during the pointer gesture.
    // Invalidate this drag even when that ancestor consumes the Escape key.
    if (_activeDrag != null &&
        event is KeyDownEvent &&
        event.logicalKey == LogicalKeyboardKey.escape) {
      _clearDrag();
    }
    return false;
  }

  void _revealFocusedControl() {
    final focused = FocusManager.instance.primaryFocus;
    final target = focused?.context;
    if (target == null) return;
    // Native browser/accessibility focus calls FocusNode.requestFocus directly,
    // bypassing the traversal policy's scroll-to-visible callback. Reveal only
    // the current board descendant after layout; never reclaim focus or scroll
    // behind a modal opened in the meantime.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted ||
          !target.mounted ||
          !identical(FocusManager.instance.primaryFocus, focused) ||
          target.findAncestorStateOfType<_CgmsAdaptiveTableState>() != this ||
          !(ModalRoute.of(target)?.isCurrent ?? false))
        return;
      Scrollable.ensureVisible(
        target,
        alignmentPolicy: ScrollPositionAlignmentPolicy.keepVisibleAtStart,
      );
      Scrollable.ensureVisible(
        target,
        alignmentPolicy: ScrollPositionAlignmentPolicy.keepVisibleAtEnd,
      );
    });
    WidgetsBinding.instance.ensureVisualUpdate();
  }

  void _revealActions({bool requiredByHost = false}) {
    if (widget.actions == null) return;
    final generation = ++_actionsGeneration;
    final game = widget.data.gameId;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted &&
          generation == _actionsGeneration &&
          game == widget.data.gameId &&
          (!requiredByHost || widget.showDecisions) &&
          // Authority decisions may arrive while a chooser is open. Preparing
          // the panel underneath preserves the modal's focus and makes the
          // decision available immediately when that modal is dismissed.
          (requiredByHost || (ModalRoute.of(context)?.isCurrent ?? true))) {
        if (!_actionsOpen) setState(() => _actionsOpen = true);
      }
    });
  }

  void _closeActions() {
    _actionsGeneration++;
    final current = FocusManager.instance.primaryFocus?.context;
    // A focused form control must leave the body before it becomes offstage.
    if (current != null) {
      var inside = false;
      current.visitAncestorElements((element) {
        if (element == _actionsKey.currentContext) inside = true;
        return !inside;
      });
      if (inside) {
        _actionToggleFocus.requestFocus();
        FocusManager.instance.applyFocusChangesIfNeeded();
      }
    }
    setState(() => _actionsOpen = false);
    widget.onCloseActions?.call();
  }

  bool _selected(VisibleCard card) =>
      widget.selectedCardIds.contains(card.id) ||
      widget.data.selectedCardId == card.id;
  bool get _canSelect =>
      widget.data.connection == ConnectionStateView.connected &&
      widget.data.pending == null;

  Set<String> get _ownVisibleIds => {
    for (final card in widget.data.hand) card.id,
    for (final seat in widget.data.seats.where((seat) => seat.isSelf)) ...{
      for (final card in seat.exposed) card.id,
      for (final combo in seat.combos)
        for (final card in combo.cards) card.id,
    },
  };

  Set<String> get _allowedDragIds =>
      _ownVisibleIds.intersection(widget.draggableCardIds);

  bool _validDrag(TableCardDrag drag) {
    if (!_canSelect ||
        widget.onCardDrop == null ||
        drag.gameId != widget.data.gameId ||
        drag.intentIdentity != widget.dragIdentity ||
        drag.cardIds.isEmpty)
      return false;
    final allowed = _allowedDragIds;
    return drag.cardIds.every(allowed.contains);
  }

  bool _validTarget(TableDropTarget target) {
    if (target.zone == TableDropZone.hand ||
        target.zone == TableDropZone.drawPile)
      return true;
    final seat = widget.data.seats
        .where((seat) => seat.number == target.seat)
        .firstOrNull;
    if (seat == null) return false;
    return switch (target.zone) {
      TableDropZone.seat => true,
      TableDropZone.series => _groups(
        seat,
      ).any((group) => group.combo == null && group.id == target.suit),
      TableDropZone.formation => seat.combos.any(
        (combo) => combo.id == target.formationId,
      ),
      TableDropZone.card => _groups(seat).any(
        (group) =>
            (target.formationId == null ||
                group.combo?.id == target.formationId) &&
            group.cards.any((card) => card.id == target.cardId),
      ),
      _ => false,
    };
  }

  void _trackDrag(Offset position) {
    final origin = _dragOrigin;
    if (origin != null && (position - origin).distance > kTouchSlop) {
      _dragMoved = true;
    }
  }

  void _clearDrag({bool closeChooser = false}) {
    final hadDrag = _activeDrag != null;
    _activeDrag = null;
    _dragOrigin = null;
    _dragMoved = false;
    _dragPreview.value = null;
    _previewTarget = null;
    if (mounted && hadDrag) setState(() {});
    if (mounted && _chooserDragging.value) {
      _chooserDragging.value = false;
      final route = _pileRoute;
      if (closeChooser && route != null && route.isCurrent) {
        route.navigator?.pop();
      }
    }
  }

  Widget _dropTarget(TableDropTarget target, Widget child) =>
      DragTarget<TableCardDrag>(
        onWillAcceptWithDetails: (details) =>
            _validDrag(details.data) &&
            _validTarget(target) &&
            // A rejected physical card/series/formation must not fall through
            // to its enclosing seat and become a different move.
            (const {
                  TableDropZone.card,
                  TableDropZone.series,
                  TableDropZone.formation,
                }.contains(target.zone) ||
                (widget.canDrop?.call(details.data, target) ?? true)),
        onMove: (details) {
          _dragPosition.value = details.offset;
          _trackDrag(details.offset);
          if (_validDrag(details.data) && _validTarget(target)) {
            _previewTarget = target;
            _dragPreview.value =
                (widget.canDrop?.call(details.data, target) ?? true)
                ? widget.dropPreview?.call(details.data, target)
                : null;
          }
        },
        onLeave: (_) {
          if (identical(_previewTarget, target)) {
            _previewTarget = null;
            _dragPreview.value = null;
          }
        },
        onAcceptWithDetails: (details) {
          // A chooser may have unmounted its source as the drag began, so its
          // onDragUpdate is no longer delivered. The final pointer is still
          // authoritative for distinguishing a drag from a stationary hold.
          _trackDrag(details.offset);
          // Re-resolve both sides against the latest authority projection. A
          // changed game, revoked card, or removed formation cancels the draft.
          if (!identical(_activeDrag, details.data) ||
              !_dragMoved ||
              !_validDrag(details.data) ||
              !_validTarget(target) ||
              !(widget.canDrop?.call(details.data, target) ?? true))
            return;
          _clearDrag(closeChooser: true);
          widget.onCardDrop!(details.data, target);
        },
        builder: (context, candidates, rejected) => DecoratedBox(
          position: DecorationPosition.foreground,
          decoration: BoxDecoration(
            border:
                candidates.any(
                      (drag) =>
                          drag != null &&
                          _validDrag(drag) &&
                          (widget.canDrop?.call(drag, target) ?? true),
                    ) ||
                    (_activeDrag != null &&
                        widget.canDrop != null &&
                        _validDrag(_activeDrag!) &&
                        _validTarget(target) &&
                        widget.canDrop!(_activeDrag!, target)) ||
                    (_activeDrag == null &&
                        widget.data.connection ==
                            ConnectionStateView.connected &&
                        _validTarget(target) &&
                        (widget.isSuggestedTarget?.call(target) ?? false))
                ? Border.all(color: CgmsColors.cyan, width: 2)
                : null,
          ),
          child: child,
        ),
      );

  Widget _cardDrag(VisibleCard card, Widget child, {VoidCallback? onStart}) {
    final allowed = _allowedDragIds;
    final selected = {...widget.selectedCardIds, ?widget.data.selectedCardId};
    final payload = TableCardDrag(
      gameId: widget.data.gameId,
      intentIdentity: widget.dragIdentity,
      cardIds: selected.contains(card.id)
          ? (widget.actions == null
                ? selected.toList()
                : selected.intersection(allowed).toList())
          : [card.id],
    );
    return _bundleDrag(payload, child, onStart: onStart);
  }

  Widget _bundleDrag(
    TableCardDrag payload,
    Widget child, {
    VoidCallback? onStart,
  }) {
    final enabled = _validDrag(payload);
    Widget draggable(Widget child, {required bool touch}) => _CardDraggable(
      touch: touch,
      data: payload,
      maxSimultaneousDrags: enabled ? 1 : 0,
      rootOverlay: true,
      ignoringFeedbackSemantics: false,
      dragAnchorStrategy: (draggable, context, position) {
        _activeDrag = payload;
        _dragOrigin = position;
        _dragPosition.value = position;
        _dragMoved = false;
        return Offset.zero;
      },
      onDragStarted: () {
        if (mounted) setState(() {});
        onStart?.call();
      },
      onDragUpdate: (details) => _trackDrag(details.globalPosition),
      onDragCompleted: _clearDrag,
      onDraggableCanceled: (_, __) => _clearDrag(),
      feedback: Material(
        color: Colors.transparent,
        child: IgnorePointer(
          child: _CardDragFeedback(
            position: _dragPosition,
            preview: _dragPreview,
            count: payload.cardIds.length,
            showArt: widget.showArt,
            textScaler: MediaQuery.textScalerOf(context),
            direction: Directionality.of(context),
          ),
        ),
      ),
      child: child,
    );
    // Keep both wrappers mounted as authorization changes so the card's
    // selection, inspection and keyboard focus retain their existing state.
    return draggable(draggable(child, touch: true), touch: false);
  }

  Widget _publicCard(
    SeatView seat,
    VisibleCard card,
    ComboView? combo,
    Widget child, {
    VoidCallback? onStart,
  }) => _dropTarget(
    TableDropTarget(
      zone: TableDropZone.card,
      seat: seat.number,
      cardId: card.id,
      formationId: combo?.id,
    ),
    // Opponents can be destinations but can never supply draggable identities.
    seat.isSelf ? _cardDrag(card, child, onStart: onStart) : child,
  );

  Widget _formationTarget(SeatView seat, ComboView? combo, Widget child) =>
      combo == null
      ? child
      : _dropTarget(
          TableDropTarget(
            zone: TableDropZone.formation,
            seat: seat.number,
            formationId: combo.id,
          ),
          child,
        );

  @override
  Widget build(BuildContext context) => LayoutBuilder(
    builder: (context, box) {
      final layout = (
        size: box.biggest,
        textScale: MediaQuery.textScalerOf(context).scale(14),
        direction: Directionality.of(context),
      );
      if (_layout != layout) {
        _layout = layout;
        _revealFocusedControl();
      }
      final kind = box.maxWidth < 600
          ? 'phone'
          : box.maxWidth < 1050
          ? 'tablet'
          : 'desktop';
      // Height belongs to the actual host slot (after its toolbar/notices).
      // Width alone must never grow a public table beyond the available screen.
      final unit = math
          .min(box.maxWidth / 390, box.maxHeight / 700)
          .clamp(1.0, 2.0);
      // Bars follow viewport density without multiplying the user text scale
      // by the separate artwork scale used for the table and private hand.
      final barMedia = MediaQuery.of(context);
      final bars = _GameBarMetrics(box.maxWidth, box.maxHeight);
      final footerLimit = math.max(48.0, box.maxHeight * .30);
      // Public and private cards share one size scale. Reserve enough height
      // for two rows of public cards and metadata, bounded below by the hand.
      final preferredHeight = math.min(
        box.maxWidth - 16,
        2 * (_publicSeatHeight * unit + 6),
      );
      final table = Focus(
        canRequestFocus: false,
        onKeyEvent: (_, event) {
          if (event is KeyDownEvent &&
              event.logicalKey == LogicalKeyboardKey.escape) {
            final handled =
                _activeDrag != null || widget.onCancelSelection != null;
            _clearDrag();
            widget.onCancelSelection?.call();
            return handled ? KeyEventResult.handled : KeyEventResult.ignored;
          }
          return KeyEventResult.ignored;
        },
        child: KeyedSubtree(
          key: const ValueKey('reference-layout'),
          child: MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(
                MediaQuery.textScalerOf(context).scale(14) / 14 * unit,
              ),
            ),
            child: Builder(
              builder: (context) => FocusTraversalGroup(
                policy: WidgetOrderTraversalPolicy(),
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  child: Semantics(
                    identifier: 'board-viewport',
                    container: true,
                    explicitChildNodes: true,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        SizedBox(key: ValueKey('$kind-table'), height: 0),
                        MediaQuery(
                          data: barMedia,
                          child: _barTheme(
                            context,
                            bars,
                            header: true,
                            child: _header(
                              bars,
                              box.maxHeight - footerLimit - 232,
                            ),
                          ),
                        ),
                        Expanded(
                          child: _workspace(
                            context,
                            preferredHeight,
                            actionTextScale: barMedia.textScaler.scale(14) / 14,
                          ),
                        ),
                        if (widget.footer != null)
                          MediaQuery(
                            data: barMedia,
                            child: _barTheme(
                              context,
                              bars,
                              child: Container(
                                key: const ValueKey('board-footer'),
                                constraints: BoxConstraints(
                                  maxHeight: footerLimit,
                                ),
                                decoration: const BoxDecoration(
                                  color: CgmsColors.surface,
                                  border: Border(
                                    top: BorderSide(color: CgmsColors.border),
                                  ),
                                ),
                                // Tight semantic bounds also update while a
                                // settings popup blocks the board's semantics.
                                child: IntrinsicHeight(
                                  child: Semantics(
                                    identifier: 'game-footer',
                                    container: true,
                                    explicitChildNodes: true,
                                    child: SingleChildScrollView(
                                      child: Padding(
                                        padding: EdgeInsets.symmetric(
                                          horizontal: bars.inset,
                                          vertical: bars.verticalInset,
                                        ),
                                        child: widget.footer!,
                                      ),
                                    ),
                                  ),
                                ),
                              ),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      );
      if (widget.actions != null) return table;
      return SingleChildScrollView(
        key: const PageStorageKey('card-board-scroll'),
        child: SizedBox(height: math.max(144, box.maxHeight), child: table),
      );
    },
  );

  Widget _workspace(
    BuildContext context,
    double preferredHeight, {
    required double actionTextScale,
  }) => LayoutBuilder(
    builder: (context, box) {
      // The naturally sized header and footer have already taken their space.
      // Shrink informational previews before consuming the first hand target.
      // Recovery notices can leave less than the combined 48px target minima.
      // Keep the same mounted workspace, scrolling this last-resort short slot.
      final actionTarget = math.max(48.0, 48 + 64 * (actionTextScale - 1));
      final minimumActions = widget.actions != null && _actionsOpen
          ? actionTarget + 4
          : 0.0;
      final scale = MediaQuery.textScalerOf(context).scale(14) / 14;
      // Direct card controls never shrink below their readable 72px faces.
      // Short/enlarged screens scroll this complete shared vertical composition.
      final directPublicHeight =
          2 * (48 + 48 * scale + 90 * scale + 8) +
          _drawPileSpaceHeight(compact: box.maxWidth + 16 < 600);
      final height = math.max(
        widget.actions == null
            ? directPublicHeight + 160 * scale
            : math.max(200.0, 148 + minimumActions),
        box.maxHeight,
      );
      // Reserve a complete first hand card, its title and suit heading before
      // assigning the remaining space. Very short slots retain a 48px target.
      final handReserve = math.min(
        (_cardWidth * 5 / 4 + 15 * 1.1 + 14 * 1.1) * scale + 10,
        math.max(48.0, height - 96 - 4 - minimumActions),
      );
      // Keep public faces at 64px with readable 12px metadata when the viewport
      // can support them. A percentage cap made roomy views collapse to 24px.
      final readablePublicHeight =
          2 * (_publicSeatHeight * 64 / _cardWidth + 6);
      final publicHeight = widget.actions == null
          ? directPublicHeight
          : math.max(
              96.0,
              math.min(
                _actionsOpen
                    ? math.min(
                        preferredHeight,
                        math.max(
                          readablePublicHeight,
                          height - handReserve - 8 - 320 * actionTextScale,
                        ),
                      )
                    : preferredHeight,
                height - 4 - handReserve - minimumActions,
              ),
            );
      final actionHeight = math.max(
        actionTarget,
        height - publicHeight - 8 - handReserve,
      );
      return SingleChildScrollView(
        key: const PageStorageKey('adaptive-workspace'),
        child: SizedBox(
          height: height,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _public(context, box.maxWidth, publicHeight),
              const SizedBox(height: 4),
              Expanded(
                child: Semantics(
                  identifier: 'private-hand-viewport',
                  container: true,
                  explicitChildNodes: true,
                  child: SizedBox(
                    key: const ValueKey('adaptive-hand-viewport'),
                    child: _dropTarget(
                      const TableDropTarget(zone: TableDropZone.hand),
                      widget.onTargetTap == null
                          ? _hand(context)
                          : InkWell(
                              onTap: () => _tapTarget(
                                const TableDropTarget(zone: TableDropZone.hand),
                              ),
                              child: _hand(context),
                            ),
                    ),
                  ),
                ),
              ),
              // Decisions never replace the public table. Short forms return their
              // unused height to the adjacent hand; longer forms scroll in place.
              if (widget.actions != null)
                Offstage(
                  offstage: !_actionsOpen,
                  child: ExcludeFocus(
                    excluding: !_actionsOpen,
                    child: Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: ConstrainedBox(
                        constraints: BoxConstraints(maxHeight: actionHeight),
                        child: Material(
                          color: CgmsColors.surface,
                          child: SingleChildScrollView(
                            // A different decision gets a fresh semantic scroll role.
                            // Keep the global child key for form state.
                            key: widget.actionsIdentity == null
                                ? const PageStorageKey('adaptive-decisions')
                                : PageStorageKey((
                                    'adaptive-decisions',
                                    widget.actionsIdentity,
                                  )),
                            child: KeyedSubtree(
                              key: _actionsKey,
                              child: widget.actions!,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      );
    },
  );

  Widget _header(_GameBarMetrics bars, double limit) => Builder(
    builder: (context) {
      return Container(
        key: const ValueKey('game-header'),
        constraints: BoxConstraints(
          minHeight: 48,
          maxHeight: math.max(80, limit),
        ),
        decoration: const BoxDecoration(
          color: CgmsColors.surface,
          border: Border(bottom: BorderSide(color: CgmsColors.divider)),
        ),
        // A tight boundary invalidates cached geometry when text is resized
        // behind a popup. It preserves the same semantic IDs and focus nodes.
        child: IntrinsicHeight(
          child: Semantics(
            identifier: 'game-header',
            container: true,
            explicitChildNodes: true,
            child: SingleChildScrollView(
              child: Padding(
                padding: EdgeInsets.symmetric(
                  horizontal: bars.inset,
                  vertical: bars.verticalInset,
                ),
                child: Flex(
                  direction: Axis.horizontal,
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Flexible(
                      fit: FlexFit.tight,
                      child: Semantics(
                        identifier: 'header-status',
                        container: true,
                        child: ConstrainedBox(
                          constraints: BoxConstraints(
                            maxHeight:
                                math.max(80, limit) - 2 * bars.verticalInset,
                          ),
                          child: SingleChildScrollView(
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                widget.header ?? _status(context),
                                if (widget.data.pending case final pending?)
                                  Semantics(
                                    liveRegion: true,
                                    child: Text(
                                      '${pending.title} · ${pending.detail}',
                                    ),
                                  ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),
                    SizedBox(width: bars.headerGap),
                    Align(
                      alignment: AlignmentDirectional.centerEnd,
                      widthFactor: 1,
                      child: SingleChildScrollView(
                        scrollDirection: Axis.horizontal,
                        child: Semantics(
                          identifier: 'header-tools',
                          container: true,
                          explicitChildNodes: true,
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              if (widget.onCoup != null)
                                IconButton(
                                  key: const ValueKey('immediate-coup'),
                                  tooltip: CgmsLocalizations.of(
                                    context,
                                  ).adaptiveCoup,
                                  onPressed: widget.onCoup,
                                  icon: const Icon(Icons.bolt),
                                ),
                              if (widget.actions != null)
                                IconButton(
                                  focusNode: _actionToggleFocus,
                                  tooltip: _actionsOpen
                                      ? 'Close actions'
                                      : 'Choose an action',
                                  onPressed: _actionsOpen
                                      ? _closeActions
                                      : _revealActions,
                                  icon: Icon(
                                    _actionsOpen ? Icons.close : Icons.tune,
                                  ),
                                ),
                              if (widget.headerTools != null)
                                widget.headerTools!,
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      );
    },
  );

  Widget _drawPile(BuildContext context, {required bool compact}) {
    final width = compact ? 28.0 : 36.0;
    final label = CgmsLocalizations.of(context).drawPile;
    // A pile location, not a stock estimate: the viewer projection deliberately
    // contains no draw-card identities or remaining-supply field.
    return _dropTarget(
      const TableDropTarget(zone: TableDropZone.drawPile),
      CgmsDrawFeedback(
        feedback: widget.drawFeedback,
        viewerSeat:
            widget.data.seats
                .where((seat) => seat.isSelf)
                .firstOrNull
                ?.number ??
            0,
        child: InkWell(
          onTap: widget.onTargetTap == null
              ? null
              : () => _tapTarget(
                  const TableDropTarget(zone: TableDropZone.drawPile),
                ),
          child: ConstrainedBox(
            constraints: const BoxConstraints(minWidth: 48, minHeight: 48),
            child: Tooltip(
              message: label,
              excludeFromSemantics: true,
              child: Semantics(
                key: const ValueKey('adaptive-draw-pile'),
                label: label,
                image: true,
                child: ExcludeSemantics(
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      SizedBox(
                        width: width + 6,
                        height: width * 1.4 + 6,
                        child: Stack(
                          children: [
                            for (final offset in [6.0, 3.0, 0.0])
                              PositionedDirectional(
                                start: offset,
                                top: offset,
                                child: ConcealedCard(
                                  width: width,
                                  showArt: widget.showArt,
                                ),
                              ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  // The band includes the complete painted stack and its 48px interaction
  // minimum. It is separate from seat hit testing at every responsive size.
  double _drawPileSpaceHeight({required bool compact}) =>
      math.max(48.0, (compact ? 28.0 : 36.0) * 1.4 + 6) + 8;

  Widget _barTheme(
    BuildContext context,
    _GameBarMetrics bars, {
    required Widget child,
    bool header = false,
  }) {
    final theme = Theme.of(context);
    final fontSize = header ? bars.headerFontSize : bars.fontSize;
    final iconSize = header ? bars.headerIconSize : bars.icon;
    ButtonStyle fit(ButtonStyle? style) =>
        (style ?? const ButtonStyle()).copyWith(
          minimumSize: WidgetStatePropertyAll(Size(48, bars.target)),
          padding: WidgetStatePropertyAll(
            EdgeInsets.symmetric(horizontal: bars.inset + 8, vertical: 6),
          ),
          textStyle: WidgetStatePropertyAll(
            theme.textTheme.labelLarge!.copyWith(fontSize: fontSize),
          ),
        );
    return Theme(
      data: theme.copyWith(
        // Desktop compact density otherwise subtracts eight pixels from the
        // painted buttons while the icon controls keep their intended size.
        visualDensity: VisualDensity.standard,
        filledButtonTheme: FilledButtonThemeData(
          style: fit(theme.filledButtonTheme.style),
        ),
        outlinedButtonTheme: OutlinedButtonThemeData(
          style: fit(theme.outlinedButtonTheme.style).copyWith(
            foregroundColor: WidgetStateProperty.resolveWith(
              (states) => states.contains(WidgetState.disabled)
                  ? CgmsColors.muted
                  : CgmsColors.cyan,
            ),
          ),
        ),
        textButtonTheme: TextButtonThemeData(
          style: fit(theme.textButtonTheme.style),
        ),
        iconButtonTheme: IconButtonThemeData(
          style: (theme.iconButtonTheme.style ?? const ButtonStyle()).copyWith(
            minimumSize: WidgetStatePropertyAll(Size(bars.target, bars.target)),
            iconSize: WidgetStatePropertyAll(iconSize),
            foregroundColor: const WidgetStatePropertyAll(CgmsColors.muted),
          ),
        ),
        iconTheme: theme.iconTheme.copyWith(size: iconSize),
      ),
      child: _GameBarScope(
        width: bars.innerWidth,
        child: DefaultTextStyle.merge(
          style: TextStyle(fontSize: fontSize),
          child: child,
        ),
      ),
    );
  }

  Widget _status(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    return Padding(
      padding: EdgeInsets.zero,
      child: Wrap(
        spacing: 8,
        runSpacing: 4,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [
          Text(
            strings.actionBoardRound(
              widget.data.round,
              widget.data.totalRounds,
            ),
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          Semantics(
            liveRegion: true,
            child: Text(
              widget.data.turnLabel,
              style: const TextStyle(color: CgmsColors.cyan),
            ),
          ),
        ],
      ),
    );
  }

  Widget _public(BuildContext context, double width, double height) => SizedBox(
    key: const ValueKey('adaptive-public-square'),
    width: width,
    height: height,
    child: Semantics(
      identifier: 'public-square',
      container: true,
      header: true,
      label: CgmsLocalizations.of(context).actionBoardPublic,
      explicitChildNodes: true,
      child: widget.actions == null
          ? Column(
              children: [
                Expanded(child: _publicRow(context, 0)),
                Semantics(
                  identifier: 'draw-pile-space',
                  container: true,
                  explicitChildNodes: true,
                  child: SizedBox(
                    key: const ValueKey('adaptive-draw-pile-space'),
                    height: _drawPileSpaceHeight(compact: width + 16 < 600),
                    child: Center(
                      child: _drawPile(context, compact: width + 16 < 600),
                    ),
                  ),
                ),
                Expanded(child: _publicRow(context, 1)),
              ],
            )
          : Stack(
              fit: StackFit.expand,
              children: [
                Column(
                  children: [
                    for (var row = 0; row < 2; row++)
                      Expanded(child: _publicRow(context, row)),
                  ],
                ),
                Center(
                  child: ColoredBox(
                    color: CgmsColors.canvas,
                    child: Padding(
                      padding: const EdgeInsets.all(4),
                      child: _drawPile(context, compact: width + 16 < 600),
                    ),
                  ),
                ),
              ],
            ),
    ),
  );

  Widget _publicRow(BuildContext context, int row) => Row(
    children: [
      for (var column = 0; column < 2; column++)
        Expanded(
          child: row * 2 + column < widget.data.seats.length
              ? _seat(
                  context,
                  widget.data.seats[row * 2 + column],
                  column: column,
                )
              : Container(
                  key: ValueKey('adaptive-seat-empty-${row * 2 + column}'),
                  decoration: BoxDecoration(
                    border: Border.all(color: CgmsColors.border),
                  ),
                ),
        ),
    ],
  );

  List<({String id, String label, List<VisibleCard> cards, ComboView? combo})>
  _groups(SeatView seat) {
    final strings = CgmsLocalizations.of(context);
    final allocated = seat.combos
        .expand((c) => c.cards)
        .map((c) => c.id)
        .toSet();
    final free = seat.exposed.where((c) => !allocated.contains(c.id));
    return [
      for (final suit in [
        CardSuit.diamonds,
        CardSuit.clubs,
        CardSuit.hearts,
        CardSuit.spades,
      ])
        if (free.any((c) => (c.seriesSuit ?? c.suit) == suit))
          (
            id: suit.name,
            label:
                widget.publicGroupLabels['${seat.id}:${suit.name}'] ??
                switch (suit) {
                  CardSuit.hearts => strings.hearts,
                  CardSuit.diamonds => strings.diamonds,
                  CardSuit.clubs => strings.clubs,
                  CardSuit.spades => strings.spades,
                },
            cards: free.where((c) => (c.seriesSuit ?? c.suit) == suit).toList(),
            combo: null,
          ),
      for (final combo in seat.combos)
        (id: combo.id, label: combo.title, cards: combo.cards, combo: combo),
    ];
  }

  Widget _seat(BuildContext context, SeatView seat, {required int column}) {
    if (widget.actions == null) return _cardDrivenSeat(context, seat);
    final strings = CgmsLocalizations.of(context);
    final groups = _groups(seat);
    return _dropTarget(
      TableDropTarget(zone: TableDropZone.seat, seat: seat.number),
      Semantics(
        button: true,
        label: 'View ${seat.name} public cards',
        value:
            '${strings.actionBoardCounts(seat.concealed.handCount, seat.concealed.aceCount)} ${seat.status}',
        onTap: () => _openSeat(seat.id),
        child: InkWell(
          key: ValueKey('adaptive-seat-${seat.id}'),
          excludeFromSemantics: true,
          onTap: () => _openSeat(seat.id),
          child: DecoratedBox(
            decoration: BoxDecoration(
              border: Border.all(color: CgmsColors.border),
            ),
            child: Padding(
              // Leave the middle of the table for the shared face-down deck.
              padding: EdgeInsetsDirectional.fromSTEB(
                column == 1 ? 28 : 3,
                3,
                column == 0 ? 28 : 3,
                3,
              ),
              child: ExcludeSemantics(
                child: ExcludeFocus(
                  child: IgnorePointer(
                    ignoring: widget.onCardDrop == null,
                    child: LayoutBuilder(
                      builder: (context, box) {
                        // Match the hand scale in roomy cells. Short views can
                        // shrink decorative previews; the entire cell remains
                        // an accessible control for full authorized details.
                        final scale = math.min(
                          MediaQuery.textScalerOf(context).scale(14) / 14,
                          box.maxHeight / _publicSeatHeight,
                        );
                        return FittedBox(
                          alignment: AlignmentDirectional.topStart,
                          fit: BoxFit.contain,
                          child: MediaQuery.withNoTextScaling(
                            child: SizedBox(
                              width: box.maxWidth / scale,
                              height: _publicSeatHeight,
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.stretch,
                                children: [
                                  Text(
                                    '${seat.number} ${seat.name}${seat.isSelf ? ' · ${strings.actionBoardYou}' : ''}',
                                    maxLines: 1,
                                    style: const TextStyle(
                                      fontSize: 14,
                                      height: 1.2,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  Text(
                                    strings.actionBoardCounts(
                                      seat.concealed.handCount,
                                      seat.concealed.aceCount,
                                    ),
                                    maxLines: 1,
                                    style: const TextStyle(
                                      fontSize: 11,
                                      height: 1.2,
                                    ),
                                  ),
                                  Text(
                                    seat.status,
                                    maxLines: 1,
                                    style: const TextStyle(
                                      fontSize: 10,
                                      height: 1.2,
                                      color: CgmsColors.muted,
                                    ),
                                  ),
                                  Expanded(
                                    child: groups.isEmpty
                                        ? Center(
                                            child: Text(
                                              strings.touchNoPublicCards,
                                              style: const TextStyle(
                                                fontSize: 11,
                                              ),
                                            ),
                                          )
                                        : Row(
                                            children: [
                                              for (final group in groups)
                                                Expanded(
                                                  child: _formationTarget(
                                                    seat,
                                                    group.combo,
                                                    Column(
                                                      children: [
                                                        Expanded(
                                                          child: FittedBox(
                                                            fit: BoxFit
                                                                .scaleDown,
                                                            child: SizedBox(
                                                              width: _cardWidth,
                                                              height: group
                                                                  .cards
                                                                  .asMap()
                                                                  .entries
                                                                  .fold<double>(
                                                                    1,
                                                                    (
                                                                      height,
                                                                      entry,
                                                                    ) => math.max(
                                                                      height,
                                                                      entry.key *
                                                                              _pileStep(
                                                                                group.cards.length,
                                                                              ) +
                                                                          _cardWidth /
                                                                              entry.value.artworkAspectRatio,
                                                                    ),
                                                                  ),
                                                              child: Stack(
                                                                children: [
                                                                  for (
                                                                    var i = 0;
                                                                    i <
                                                                        group
                                                                            .cards
                                                                            .length;
                                                                    i++
                                                                  )
                                                                    Positioned(
                                                                      top:
                                                                          i *
                                                                          _pileStep(
                                                                            group.cards.length,
                                                                          ),
                                                                      child: _publicCard(
                                                                        seat,
                                                                        group
                                                                            .cards[i],
                                                                        group
                                                                            .combo,
                                                                        CgmsCard(
                                                                          key: ValueKey(
                                                                            'public-preview-${group.cards[i].id}',
                                                                          ),
                                                                          card:
                                                                              group.cards[i],
                                                                          width:
                                                                              _cardWidth,
                                                                          boardPreview:
                                                                              true,
                                                                          showArt:
                                                                              widget.showArt,
                                                                          selected: _selected(
                                                                            group.cards[i],
                                                                          ),
                                                                        ),
                                                                      ),
                                                                    ),
                                                                ],
                                                              ),
                                                            ),
                                                          ),
                                                        ),
                                                        Text(
                                                          group.label,
                                                          maxLines: 1,
                                                          style:
                                                              const TextStyle(
                                                                fontSize: 10,
                                                                height: 1.2,
                                                              ),
                                                        ),
                                                        if (group.combo != null)
                                                          Text(
                                                            group
                                                                .combo!
                                                                .stateLabel,
                                                            maxLines: 1,
                                                            style:
                                                                const TextStyle(
                                                                  fontSize: 9,
                                                                  height: 1.2,
                                                                ),
                                                          ),
                                                      ],
                                                    ),
                                                  ),
                                                ),
                                            ],
                                          ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  void _tapTarget(TableDropTarget target) {
    if (widget.data.connection == ConnectionStateView.connected &&
        _validTarget(target)) {
      widget.onTargetTap?.call(target);
    }
  }

  void _activateGroup(TableDropTarget target) {
    if (widget.data.connection == ConnectionStateView.connected &&
        _validTarget(target)) {
      widget.onActivateTarget?.call(target);
    }
  }

  Widget _groupHeader(
    TableDropTarget target, {
    required Key key,
    required String label,
    ComboView? dragFormation,
  }) {
    final header = Semantics(
      identifier: target.formationId == null
          ? ''
          : 'public-formation-${target.formationId}',
      button: true,
      customSemanticsActions: widget.onActivateTarget == null
          ? null
          : {
              CustomSemanticsAction(
                label: CgmsLocalizations.of(context).activateCard,
              ): () =>
                  _activateGroup(target),
            },
      child: Focus(
        canRequestFocus: false,
        onKeyEvent: (_, event) {
          if (widget.onActivateTarget != null &&
              event is KeyDownEvent &&
              event.logicalKey == LogicalKeyboardKey.enter) {
            _activateGroup(target);
            return KeyEventResult.handled;
          }
          return KeyEventResult.ignored;
        },
        child: InkWell(
          key: key,
          onTap: () => _tapTarget(target),
          onDoubleTap: widget.onActivateTarget == null
              ? null
              : () => _activateGroup(target),
          child: ConstrainedBox(
            constraints: const BoxConstraints(minHeight: 48, minWidth: 48),
            child: Center(
              child: Text(label, style: const TextStyle(fontSize: 12)),
            ),
          ),
        ),
      ),
    );
    // A formation handle always carries its exact current physical members.
    // Unrelated selected cards never expand or replace this explicit source.
    return dragFormation == null
        ? header
        : _bundleDrag(
            TableCardDrag(
              gameId: widget.data.gameId,
              intentIdentity: widget.dragIdentity,
              cardIds: dragFormation.cards.map((card) => card.id).toList(),
            ),
            header,
          );
  }

  Widget _cardDrivenSeat(BuildContext context, SeatView seat) {
    final strings = CgmsLocalizations.of(context);
    final cardWidth =
        _cardWidth * MediaQuery.textScalerOf(context).scale(14) / 14;
    final cardStep = math.max(48.0, cardWidth * .66);
    final target = TableDropTarget(zone: TableDropZone.seat, seat: seat.number);
    final groups = _groups(seat);
    return _dropTarget(
      target,
      Semantics(
        identifier: 'public-seat-${seat.number}',
        container: true,
        explicitChildNodes: true,
        child: Container(
          key: ValueKey('adaptive-seat-${seat.id}'),
          decoration: BoxDecoration(
            border: Border.all(color: CgmsColors.border),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                children: [
                  Expanded(
                    child: InkWell(
                      key: ValueKey('seat-target-${seat.number}'),
                      onTap: () => _tapTarget(target),
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(minHeight: 48),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 4),
                          child: Align(
                            alignment: AlignmentDirectional.centerStart,
                            child: Text(
                              '${seat.number} ${seat.name}${seat.isSelf ? ' · ${strings.actionBoardYou}' : ''}',
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                  IconButton(
                    onPressed: () => _openSeat(seat.id),
                    tooltip: strings.inspectPublicCards,
                    icon: const Icon(Icons.zoom_in, size: 18),
                  ),
                ],
              ),
              Expanded(
                child: SingleChildScrollView(
                  key: PageStorageKey('public-cards-${seat.id}'),
                  padding: const EdgeInsetsDirectional.fromSTEB(4, 0, 4, 4),
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      for (final group in groups)
                        Padding(
                          padding: const EdgeInsetsDirectional.only(end: 8),
                          child: _dropTarget(
                            TableDropTarget(
                              zone: group.combo == null
                                  ? TableDropZone.series
                                  : TableDropZone.formation,
                              seat: seat.number,
                              suit: group.combo == null ? group.id : null,
                              formationId: group.combo?.id,
                            ),
                            Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                _groupHeader(
                                  TableDropTarget(
                                    zone: group.combo == null
                                        ? TableDropZone.series
                                        : TableDropZone.formation,
                                    seat: seat.number,
                                    formationId: group.combo?.id,
                                    suit: group.combo == null ? group.id : null,
                                  ),
                                  key: ValueKey(
                                    'formation-target-${seat.number}-${group.id}',
                                  ),
                                  label: group.label,
                                  dragFormation: seat.isSelf
                                      ? group.combo
                                      : null,
                                ),
                                SizedBox(
                                  width:
                                      cardWidth +
                                      (group.cards.length - 1) * cardStep,
                                  height: cardWidth * 1.25,
                                  child: Stack(
                                    children: [
                                      for (
                                        var index = 0;
                                        index < group.cards.length;
                                        index++
                                      )
                                        PositionedDirectional(
                                          start: index * cardStep,
                                          top: 0,
                                          child: _publicCard(
                                            seat,
                                            group.cards[index],
                                            group.combo,
                                            CgmsCard(
                                              key: ValueKey(
                                                'public-preview-${group.cards[index].id}',
                                              ),
                                              card: group.cards[index],
                                              width: cardWidth,
                                              boardPreview: true,
                                              handStackPreview: true,
                                              showArt: widget.showArt,
                                              selected: _selected(
                                                group.cards[index],
                                              ),
                                              selectionRole:
                                                  widget.cardRoles[group
                                                      .cards[index]
                                                      .id],
                                              onTap: _canSelect
                                                  ? () => widget.onSelect(
                                                      group.cards[index].id,
                                                    )
                                                  : null,
                                              onActivate:
                                                  widget.onActivateCard == null
                                                  ? null
                                                  : () =>
                                                        widget.onActivateCard!(
                                                          group.cards[index].id,
                                                        ),
                                              onInspect: () =>
                                                  _inspect(group.cards[index]),
                                            ),
                                          ),
                                        ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // Dense holdings must not make previews progressively smaller. Full physical
  // identities remain available through the authorized quadrant chooser.
  double _pileStep(int count) => math.min(4.0, 16 / math.max(1, count - 1));

  String stringsForSeat(BuildContext context, SeatView seat) =>
      '${CgmsLocalizations.of(context).actionBoardCounts(seat.concealed.handCount, seat.concealed.aceCount)} · ${seat.status}';

  void _openSeat(String seatId) {
    if (_pileRoute?.isActive ?? false) return;
    final gameId = widget.data.gameId;
    final route = _PublicCardsRoute(
      context: context,
      dragging: _chooserDragging,
      builder: (dialog) => ValueListenableBuilder<TableViewData?>(
        valueListenable: _projection,
        builder: (_, data, __) {
          if (!mounted) return const SizedBox.shrink();
          final seat = data?.gameId == gameId
              ? data?.seats.where((s) => s.id == seatId).firstOrNull
              : null;
          final groups = seat == null
              ? <
                  ({
                    String id,
                    String label,
                    List<VisibleCard> cards,
                    ComboView? combo,
                  })
                >[]
              : _groups(seat);
          void choose(String id, {required bool inspect}) {
            if (!mounted || !(ModalRoute.of(dialog)?.isCurrent ?? false))
              return;
            final currentSeat = widget.data.gameId == gameId
                ? widget.data.seats.where((s) => s.id == seatId).firstOrNull
                : null;
            final card = currentSeat == null
                ? null
                : _groups(
                    currentSeat,
                  ).expand((g) => g.cards).where((c) => c.id == id).firstOrNull;
            if (card == null || (!inspect && !_canSelect)) return;
            Navigator.of(dialog).pop();
            if (inspect) {
              _inspect(card);
            } else {
              widget.onSelect(card.id);
            }
          }

          return Directionality(
            textDirection: Directionality.of(context),
            child: MediaQuery(
              data: MediaQuery.of(context),
              child: AlertDialog(
                title: Text(
                  seat == null
                      ? 'Public cards updated'
                      : '${seat.name} public cards',
                ),
                content: SizedBox(
                  width: 520,
                  child: SingleChildScrollView(
                    child: seat == null || groups.isEmpty
                        ? const Text('These cards are no longer exposed.')
                        : Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(stringsForSeat(context, seat)),
                              for (final group in groups) ...[
                                Text(
                                  group.label,
                                  key: ValueKey(
                                    'public-pile-${seat.id}-${group.id}',
                                  ),
                                ),
                                if (group.combo != null)
                                  Text(group.combo!.stateLabel),
                                Wrap(
                                  spacing: 12,
                                  runSpacing: 12,
                                  children: [
                                    for (final card in group.cards)
                                      _publicCard(
                                        seat,
                                        card,
                                        group.combo,
                                        CgmsCard(
                                          key: ValueKey(
                                            'adaptive-select-${card.id}',
                                          ),
                                          card: card,
                                          width: 120,
                                          compact: true,
                                          showArt: widget.showArt,
                                          selected: _selected(card),
                                          onTap: _canSelect
                                              ? () => choose(
                                                  card.id,
                                                  inspect: false,
                                                )
                                              : null,
                                          onInspect: () =>
                                              choose(card.id, inspect: true),
                                          selectionRole:
                                              widget.cardRoles[card.id],
                                          onActivate:
                                              widget.onActivateCard == null
                                              ? null
                                              : () {
                                                  Navigator.of(dialog).pop();
                                                  widget.onActivateCard!(
                                                    card.id,
                                                  );
                                                },
                                        ),
                                        onStart: () =>
                                            _chooserDragging.value = true,
                                      ),
                                  ],
                                ),
                                if (group.combo != null)
                                  Text(group.combo!.description),
                                if (group.combo != null &&
                                    widget.onInspectCombo != null)
                                  TextButton(
                                    onPressed: () {
                                      if (!mounted ||
                                          !(ModalRoute.of(dialog)?.isCurrent ??
                                              false))
                                        return;
                                      final current =
                                          widget.data.gameId == gameId
                                          ? widget.data.seats
                                                .where((s) => s.id == seatId)
                                                .expand((s) => s.combos)
                                                .where((c) => c.id == group.id)
                                                .firstOrNull
                                          : null;
                                      if (current == null) return;
                                      Navigator.of(dialog).pop();
                                      widget.onInspectCombo!(current);
                                    },
                                    child: Text(
                                      CgmsLocalizations.of(
                                        context,
                                      ).inspectCombo(group.combo!.title),
                                    ),
                                  ),
                              ],
                            ],
                          ),
                  ),
                ),
                actions: [
                  if (widget.publicChooserTools != null)
                    widget.publicChooserTools!,
                  if (widget.onCoup != null)
                    FilledButton(
                      onPressed: () {
                        if (!mounted ||
                            widget.onCoup == null ||
                            !(ModalRoute.of(dialog)?.isCurrent ?? false))
                          return;
                        Navigator.of(dialog).pop();
                        widget.onCoup!();
                      },
                      child: Text(CgmsLocalizations.of(context).adaptiveCoup),
                    ),
                  TextButton(
                    autofocus: true,
                    onPressed: () => Navigator.of(dialog).pop(),
                    child: const Text('Close public cards'),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
    _pileRoute = route;
    Navigator.of(context).push(route).whenComplete(() {
      if (identical(_pileRoute, route)) _pileRoute = null;
    });
  }

  void _inspect(VisibleCard card) {
    widget.onInspect(card);
    if (widget.actions != null) _revealActions();
  }

  Widget _hand(BuildContext context) {
    final strings = CgmsLocalizations.of(context);
    final self = widget.data.seats.where((seat) => seat.isSelf).firstOrNull;
    return LayoutBuilder(
      builder: (context, constraints) {
        final scale = MediaQuery.textScalerOf(context).scale(14) / 14;
        // In a short, enlarged view, keep an actual card target instead of
        // spending the complete hand viewport on repeated hand/suit headings.
        final showHeadings = constraints.maxHeight >= 56 + 32 * scale;
        final centered = constraints.maxWidth + 16 >= 600;
        final selected = widget.data.hand.where(_selected).firstOrNull;
        final heading = Semantics(
          header: true,
          headingLevel: 2,
          child: Text(
            '${strings.yourHand} · ${widget.data.hand.length}',
            textAlign: centered ? TextAlign.center : TextAlign.start,
            maxLines: 1,
            style: const TextStyle(
              fontSize: 15,
              height: 1.1,
              fontWeight: FontWeight.bold,
            ),
          ),
        );
        return Column(
          key: const PageStorageKey('adaptive-hand'),
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Offstage(
              offstage: !showHeadings,
              child: ConstrainedBox(
                constraints: BoxConstraints(
                  maxHeight: math.max(0, constraints.maxHeight - 4),
                ),
                child: widget.actions != null
                    ? heading
                    : Row(
                        children: [
                          if (centered) const SizedBox(width: 48),
                          Expanded(child: heading),
                          // Inspection belongs to the existing heading line.
                          // Selecting a card must not insert a new row and move
                          // the same physical card away from the next tap.
                          SizedBox(
                            width: 48,
                            height: 48,
                            child: selected == null
                                ? null
                                : IconButton(
                                    tooltip: strings.inspectSelectedCards,
                                    onPressed: () => _inspect(selected),
                                    icon: const Icon(Icons.zoom_in, size: 18),
                                  ),
                          ),
                        ],
                      ),
              ),
            ),
            SizedBox(height: showHeadings ? 4 : 0),
            Expanded(
              child: NotificationListener<ScrollMetricsNotification>(
                key: const PageStorageKey('private-hand-scroll'),
                onNotification: _handMetricsChanged,
                child: SingleChildScrollView(
                  key: ValueKey((
                    'private-hand-viewport',
                    _handScrollGeneration,
                  )),
                  controller: _handScrollController,
                  primary: false,
                  child: LayoutBuilder(
                    key: _handContentKey,
                    builder: (context, box) {
                      final textScale =
                          MediaQuery.textScalerOf(context).scale(14) / 14;
                      final width = _cardWidth * textScale;
                      final step = math.max(48.0, 36 * textScale);
                      final suitCounts = [
                        for (final suit in CardSuit.values)
                          widget.data.hand
                              .where((card) => card.suit == suit)
                              .length,
                      ];
                      final columns = math.min(
                        box.maxWidth >= 2 * (width + step) + 8 ? 2 : 1,
                        math.max(
                          1,
                          suitCounts.where((count) => count > 0).length,
                        ),
                      );
                      final availableGroupWidth =
                          (box.maxWidth - 8 * (columns - 1)) / columns;
                      final longestSuit = suitCounts.fold<int>(0, math.max);
                      final groupWidth = centered
                          ? math.min(
                              availableGroupWidth,
                              math.max(
                                width + step,
                                width + (longestSuit - 1) * step,
                              ),
                            )
                          : availableGroupWidth;
                      final textAlign = centered
                          ? TextAlign.center
                          : TextAlign.start;
                      return Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: [
                          Align(
                            alignment: Alignment.center,
                            child: SizedBox(
                              width: groupWidth * columns + 8 * (columns - 1),
                              child: Wrap(
                                spacing: 8,
                                runSpacing: 8,
                                children: [
                                  for (final suit in const [
                                    CardSuit.hearts,
                                    CardSuit.spades,
                                    CardSuit.diamonds,
                                    CardSuit.clubs,
                                  ])
                                    if (widget.data.hand.any(
                                      (card) => card.suit == suit,
                                    ))
                                      _handSuit(
                                        context,
                                        suit,
                                        groupWidth,
                                        width,
                                        step,
                                        showHeading: showHeadings,
                                      ),
                                ],
                              ),
                            ),
                          ),
                          Text(
                            strings.actionBoardPrivate,
                            textAlign: textAlign,
                          ),
                          if (self != null)
                            Text(
                              strings.actionBoardAces(self.concealed.aceCount),
                              textAlign: textAlign,
                            ),
                          if (self != null && self.financesKnown)
                            Text(
                              strings.adaptiveFinances(
                                self.finances.gameScore.format(),
                                self.finances.available.format(),
                                self.finances.debt.format(),
                              ),
                              textAlign: textAlign,
                            ),
                        ],
                      );
                    },
                  ),
                ),
              ),
            ),
          ],
        );
      },
    );
  }

  Widget _handSuit(
    BuildContext context,
    CardSuit suit,
    double groupWidth,
    double cardWidth,
    double step, {
    required bool showHeading,
  }) {
    final strings = CgmsLocalizations.of(context);
    final cards = widget.data.hand.where((card) => card.suit == suit).toList();
    final label = switch (suit) {
      CardSuit.hearts => strings.hearts,
      CardSuit.spades => strings.spades,
      CardSuit.diamonds => strings.diamonds,
      CardSuit.clubs => strings.clubs,
    };
    return SizedBox(
      key: ValueKey('hand-suit-${suit.name}'),
      width: groupWidth,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Offstage(
            offstage: !showHeading,
            child: Semantics(
              header: true,
              child: Text(
                strings.suitGroup(label, cards.length),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(fontSize: 14, height: 1.1),
              ),
            ),
          ),
          SizedBox(height: showHeading ? 4 : 0),
          FocusTraversalGroup(
            policy: WidgetOrderTraversalPolicy(),
            child: SingleChildScrollView(
              key: PageStorageKey('hand-suit-scroll-${suit.name}'),
              scrollDirection: Axis.horizontal,
              child: SizedBox(
                width: cardWidth + (cards.length - 1) * step,
                height: cardWidth * 5 / 4,
                child: Stack(
                  children: [
                    for (var i = 0; i < cards.length; i++)
                      PositionedDirectional(
                        key: ValueKey('hand-position-${cards[i].id}'),
                        start: i * step,
                        top: 0,
                        child: _cardDrag(
                          cards[i],
                          CgmsCard(
                            key: ValueKey('adaptive-select-${cards[i].id}'),
                            card: cards[i],
                            width: cardWidth,
                            boardPreview: true,
                            handStackPreview: true,
                            // Clipped rectangles must not reorder semantic DOM
                            // nodes: moving a focused node drops browser focus.
                            semanticsSortKey: OrdinalSortKey(i.toDouble()),
                            showArt: widget.showArt,
                            selected: _selected(cards[i]),
                            onTap: _canSelect
                                ? () => widget.onSelect(cards[i].id)
                                : null,
                            onInspect: () => _inspect(cards[i]),
                            selectionRole: widget.cardRoles[cards[i].id],
                            onActivate: widget.onActivateCard == null
                                ? null
                                : () => widget.onActivateCard!(cards[i].id),
                          ),
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// A transient, input-transparent preview positioned inside the root overlay.
/// The draggable still hit-tests at the actual pointer; moving its visual
/// feedback must never shift a destination or steal focus from the board.
class _CardDragFeedback extends StatelessWidget {
  const _CardDragFeedback({
    required this.position,
    required this.preview,
    required this.count,
    required this.showArt,
    required this.textScaler,
    required this.direction,
  });

  final ValueListenable<Offset> position;
  final ValueListenable<String?> preview;
  final int count;
  final bool showArt;
  final TextScaler textScaler;
  final TextDirection direction;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: Listenable.merge([position, preview]),
    builder: (context, _) {
      final overlay =
          Overlay.of(context, rootOverlay: true).context.findRenderObject()!
              as RenderBox;
      final pointer = overlay.globalToLocal(position.value);
      final media = MediaQuery.of(context);
      final bounds = Rect.fromLTRB(
        media.padding.left + 8,
        media.padding.top + 8,
        overlay.size.width - media.padding.right - 8,
        overlay.size.height -
            math.max(media.padding.bottom, media.viewInsets.bottom) -
            8,
      );
      final style = Theme.of(
        context,
      ).textTheme.bodyMedium!.copyWith(height: 1.25);
      final text = preview.value;
      var width = math.min(text == null ? 48.0 : 600.0, bounds.width);
      if (text != null) {
        final measure = TextPainter(
          text: TextSpan(text: text, style: style),
          textDirection: direction,
          textScaler: textScaler,
          locale: Localizations.maybeLocaleOf(context),
        )..layout(maxWidth: math.max(1, width - 16));
        // A short landscape window needs width, not smaller text or an
        // unreadable clipped tooltip. Exact host-provided details stay intact.
        if (measure.height + 16 > bounds.height) width = bounds.width;
        measure.dispose();
      }
      return Directionality(
        textDirection: direction,
        child: MediaQuery(
          data: media.copyWith(textScaler: textScaler),
          child: Transform.translate(
            // Draggable already places feedback at the pointer. Cancel that
            // translation so layout can clamp against the whole safe viewport.
            offset: -pointer,
            child: SizedBox.fromSize(
              size: overlay.size,
              child: CustomSingleChildLayout(
                delegate: _CardDragFeedbackLayout(
                  pointer: pointer,
                  bounds: bounds,
                  width: width,
                  direction: direction,
                ),
                child: text == null
                    ? Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          ConcealedCard(width: 48, showArt: showArt),
                          if (count > 1) Text('$count'),
                        ],
                      )
                    : Semantics(
                        liveRegion: true,
                        child: Container(
                          key: const ValueKey('adaptive-drag-preview'),
                          padding: const EdgeInsets.all(8),
                          color: CgmsColors.raised,
                          child: Text(text, style: style),
                        ),
                      ),
              ),
            ),
          ),
        ),
      );
    },
  );
}

class _CardDragFeedbackLayout extends SingleChildLayoutDelegate {
  const _CardDragFeedbackLayout({
    required this.pointer,
    required this.bounds,
    required this.width,
    required this.direction,
  });

  final Offset pointer;
  final Rect bounds;
  final double width;
  final TextDirection direction;

  @override
  BoxConstraints getConstraintsForChild(BoxConstraints constraints) =>
      BoxConstraints.tightFor(width: width);

  @override
  Offset getPositionForChild(Size size, Size childSize) {
    final left = direction == TextDirection.ltr
        ? pointer.dx + 16
        : pointer.dx - childSize.width - 16;
    final top = pointer.dy + childSize.height + 20 <= bounds.bottom
        ? pointer.dy + 20
        : pointer.dy - childSize.height - 20;
    return Offset(
      left.clamp(
        bounds.left,
        math.max(bounds.left, bounds.right - childSize.width),
      ),
      top.clamp(
        bounds.top,
        math.max(bounds.top, bounds.bottom - childSize.height),
      ),
    );
  }

  @override
  bool shouldRelayout(_CardDragFeedbackLayout oldDelegate) =>
      pointer != oldDelegate.pointer ||
      bounds != oldDelegate.bounds ||
      width != oldDelegate.width ||
      direction != oldDelegate.direction;
}

/// Mouse movement starts immediately; touch must hold before moving so a
/// normal swipe still scrolls the hand. The card keeps its existing tap,
/// double-tap and keyboard inspection gestures.
class _CardDraggable extends Draggable<TableCardDrag> {
  const _CardDraggable({
    required this.touch,
    required super.data,
    required super.child,
    required super.feedback,
    required super.maxSimultaneousDrags,
    required super.rootOverlay,
    required super.ignoringFeedbackSemantics,
    required super.dragAnchorStrategy,
    required super.onDragStarted,
    required super.onDragUpdate,
    required super.onDragCompleted,
    required super.onDraggableCanceled,
  });

  final bool touch;

  @override
  MultiDragGestureRecognizer createRecognizer(
    GestureMultiDragStartCallback onStart,
  ) =>
      (touch
            ? DelayedMultiDragGestureRecognizer(
                delay: const Duration(milliseconds: 350),
                supportedDevices: const {
                  PointerDeviceKind.touch,
                  PointerDeviceKind.stylus,
                  PointerDeviceKind.invertedStylus,
                },
              )
            : ImmediateMultiDragGestureRecognizer(
                supportedDevices: const {PointerDeviceKind.mouse},
                allowedButtonsFilter: (buttons) => buttons == kPrimaryButton,
              ))
        ..onStart = onStart;
}

/// Navigating away cancels active pointers. Keep the chooser and its gesture
/// mounted while making the board behind its modal barrier a drop destination.
class _PublicCardsRoute extends DialogRoute<void> {
  _PublicCardsRoute({
    required super.context,
    required super.builder,
    required this.dragging,
  });

  final ValueListenable<bool> dragging;

  @override
  Widget buildModalBarrier() => ValueListenableBuilder<bool>(
    valueListenable: dragging,
    child: super.buildModalBarrier(),
    builder: (_, active, child) => IgnorePointer(
      ignoring: active,
      child: Opacity(opacity: active ? 0 : 1, child: child),
    ),
  );

  @override
  Widget buildPage(
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
  ) => ValueListenableBuilder<bool>(
    valueListenable: dragging,
    child: super.buildPage(context, animation, secondaryAnimation),
    builder: (_, active, child) => IgnorePointer(
      ignoring: active,
      child: Opacity(opacity: active ? 0 : 1, child: child),
    ),
  );
}

/// Host-owned controls arranged in the shared responsive game dock.
class CgmsGameActions extends StatelessWidget {
  const CgmsGameActions({
    super.key,
    required this.primary,
    required this.secondary,
    this.utilities = const [],
  });
  final Widget primary, secondary;
  final List<Widget> utilities;

  @override
  Widget build(BuildContext context) {
    final slot = context.dependOnInheritedWidgetOfExactType<_GameBarScope>();
    return slot == null
        ? LayoutBuilder(
            builder: (context, box) => _build(context, box.maxWidth),
          )
        : _build(context, slot.width);
  }

  Widget _build(BuildContext context, double width) {
    final scale = MediaQuery.textScalerOf(context).scale(16) / 16;
    final stacked = width < 560 && utilities.isNotEmpty;
    final gap = width < 600 ? 6.0 : 12.0;
    final utilityWidth =
        utilities.length *
        (Theme.of(
              context,
            ).iconButtonTheme.style?.minimumSize?.resolve({})?.width ??
            48);
    final actionsWidth = stacked || width < 560
        ? width
        : math.min(
            width - utilityWidth - (utilities.isEmpty ? 0 : gap),
            400 * scale,
          );
    // Keep the same element path while changing direction: active keyboard
    // focus and host-owned callbacks must survive a live resize.
    return Flex(
      direction: stacked ? Axis.vertical : Axis.horizontal,
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: stacked
          ? CrossAxisAlignment.stretch
          : CrossAxisAlignment.center,
      children: [
        Flexible(
          flex: stacked ? 0 : 1,
          fit: stacked ? FlexFit.loose : FlexFit.tight,
          child: Semantics(
            identifier: 'footer-utilities',
            container: true,
            explicitChildNodes: true,
            child: Row(mainAxisSize: MainAxisSize.min, children: utilities),
          ),
        ),
        SizedBox(
          width: stacked || utilities.isEmpty ? 0 : gap,
          height: stacked ? 4 : 0,
        ),
        SizedBox(
          width: actionsWidth,
          child: IntrinsicHeight(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Expanded(
                  child: Semantics(
                    identifier: 'footer-secondary',
                    container: true,
                    child: secondary,
                  ),
                ),
                SizedBox(width: gap),
                Expanded(
                  child: Semantics(
                    identifier: 'footer-primary',
                    container: true,
                    child: primary,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}

class _GameBarScope extends InheritedWidget {
  const _GameBarScope({required this.width, required super.child});
  final double width;
  @override
  bool updateShouldNotify(_GameBarScope oldWidget) => width != oldWidget.width;
}

class _GameBarMetrics {
  _GameBarMetrics(this.width, double height)
    : growth = ((width - 390) / 660).clamp(0.0, 1.0),
      short = height < 500;
  final double growth, width;
  final bool short;
  double get innerWidth => width - 16 - 2 * inset;
  double get target => short ? 48 : 48 + 8 * growth;
  double get inset => short ? 4 : 4 + 8 * growth;
  double get verticalInset => short ? 0 : 4 + 4 * growth;
  double get gap => short ? 4 : 6 + 6 * growth;
  double get icon => 22 + 4 * growth;
  double get fontSize => short ? 15 : 16 + 2 * growth;
  double get headerGrowth => ((width - 319) / 731).clamp(0.0, 1.0);
  double get headerFontSize =>
      width >= 1050 ? fontSize : math.min(fontSize, 12 + 6 * headerGrowth);
  double get headerIconSize => width >= 1050 ? icon : 18 + 8 * headerGrowth;
  double get headerGap => width >= 1050 || short ? gap : 4 + 8 * headerGrowth;
}
