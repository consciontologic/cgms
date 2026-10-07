import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';

import 'l10n/cgms_localizations.dart';
import 'theme.dart';

enum DrawDestination { hand, concealedAce }

/// A newly confirmed server draw, containing public categories only. There is
/// deliberately no physical identity, face, artwork URL or remaining deck count.
@immutable
class TableDrawFeedback {
  const TableDrawFeedback({
    required this.id,
    required this.seat,
    required this.destination,
    required this.startedAt,
  });

  static const duration = Duration(milliseconds: 1100);
  final String id;
  final int seat;
  final DrawDestination destination;
  final DateTime startedAt;
}

/// Small pile-local feedback. The original card back stays visible and all
/// interaction/focus belongs to [child]. Recreated layouts continue the same
/// elapsed feedback instead of replaying a draw or its announcement.
class CgmsDrawFeedback extends StatefulWidget {
  const CgmsDrawFeedback({
    super.key,
    this.feedback,
    required this.viewerSeat,
    required this.child,
  });

  final TableDrawFeedback? feedback;
  final int viewerSeat;
  final Widget child;

  @override
  State<CgmsDrawFeedback> createState() => _CgmsDrawFeedbackState();
}

class _CgmsDrawFeedbackState extends State<CgmsDrawFeedback>
    with SingleTickerProviderStateMixin {
  late final _animation = AnimationController(
    vsync: this,
    duration: TableDrawFeedback.duration,
  );
  Timer? _quietTimer;
  bool _active = false, _announce = false;
  bool? _reduced;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final reduced =
        MediaQuery.disableAnimationsOf(context) ||
        MediaQuery.accessibleNavigationOf(context);
    if (_reduced != reduced) {
      _reduced = reduced;
      _sync(announce: false);
    }
  }

  @override
  void didUpdateWidget(CgmsDrawFeedback oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.feedback?.id != widget.feedback?.id) _sync(announce: true);
  }

  void _sync({required bool announce}) {
    _quietTimer?.cancel();
    _animation.stop();
    final feedback = widget.feedback;
    final elapsed = feedback == null
        ? TableDrawFeedback.duration
        : DateTime.now().difference(feedback.startedAt);
    final remaining = TableDrawFeedback.duration - elapsed;
    _active = feedback != null && remaining > Duration.zero;
    _announce = _active && announce;
    if (!_active) return;
    if (_reduced != true) {
      _animation.value =
          (elapsed.inMicroseconds / TableDrawFeedback.duration.inMicroseconds)
              .clamp(0.0, 1.0);
      _animation.forward();
    }
    _quietTimer = Timer(remaining, () {
      if (mounted) setState(() => _active = _announce = false);
    });
  }

  @override
  void dispose() {
    _quietTimer?.cancel();
    _animation.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final feedback = widget.feedback;
    final l10n = CgmsLocalizations.of(context);
    final announcement = !_active || feedback == null
        ? ''
        : feedback.seat != widget.viewerSeat
        ? l10n.turnDrawBySeat(feedback.seat)
        : feedback.destination == DrawDestination.concealedAce
        ? l10n.turnDrawToConcealedAce
        : l10n.turnDrawToHand;
    return Semantics(
      liveRegion: _announce,
      label: announcement,
      explicitChildNodes: true,
      child: Stack(
        fit: StackFit.passthrough,
        children: [
          widget.child,
          if (_active)
            Positioned.fill(
              child: IgnorePointer(
                child: ExcludeSemantics(
                  child: AnimatedBuilder(
                    animation: _animation,
                    builder: (context, child) {
                      final quiet = _reduced == true;
                      final alpha = quiet
                          ? .38
                          : .22 + .4 * math.sin(math.pi * _animation.value);
                      return DecoratedBox(
                        key: ValueKey(
                          quiet ? 'draw-highlight-quiet' : 'draw-highlight',
                        ),
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(
                            CgmsShape.control,
                          ),
                          border: Border.all(
                            color: CgmsColors.diamond.withValues(alpha: alpha),
                            width: 2,
                          ),
                          boxShadow: quiet
                              ? const []
                              : [
                                  BoxShadow(
                                    color: CgmsColors.diamond.withValues(
                                      alpha: alpha / 3,
                                    ),
                                    blurRadius: 8,
                                  ),
                                ],
                        ),
                      );
                    },
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}
