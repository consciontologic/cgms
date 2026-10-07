import 'package:flutter/widgets.dart';

import 'keyboard_stub.dart'
    if (dart.library.js_interop) 'keyboard_web.dart'
    as keyboard;

class BrowserNextFocusAction extends NextFocusAction {
  @override
  bool invoke(NextFocusIntent intent) {
    final previous = primaryFocus;
    return _completeTraversal(super.invoke(intent), previous);
  }
}

class BrowserPreviousFocusAction extends PreviousFocusAction {
  @override
  bool invoke(PreviousFocusIntent intent) {
    final previous = primaryFocus;
    return _completeTraversal(super.invoke(intent), previous);
  }
}

bool _completeTraversal(bool moved, FocusNode? previous) {
  // Flutter reports false when a closed loop selects the already focused
  // control. A singleton dialog still consumes Tab: native traversal must
  // not escape it. Root scopes using leaveFlutterView retain their false.
  final handled =
      moved ||
      (previous != null &&
          identical(primaryFocus, previous) &&
          previous.nearestScope?.traversalEdgeBehavior ==
              TraversalEdgeBehavior.closedLoop);
  keyboard.completeTabTraversal(handled);
  return handled;
}
